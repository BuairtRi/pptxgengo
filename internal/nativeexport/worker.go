package nativeexport

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type workerRequest struct {
	Render  *Options       `json:"render,omitempty"`
	Doctor  *DoctorOptions `json:"doctor,omitempty"`
	Cleanup *cleanupTask   `json:"cleanup,omitempty"`
	Failure *renderFailure `json:"failure,omitempty"`
	TaskID  string         `json:"task_id,omitempty"`
}

type cleanupTask struct{ StagingRoot, TaskID string }
type renderFailure struct{ Out, Message, TaskID string }

// RecordRenderError is also used for CLI flag/preflight failures before Render
// can start. The existing private worker bounds filesystem access to three seconds.
func RecordRenderError(ctx context.Context, out string, err error) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, recordErr := workerProcess(ctx, workerRequest{Failure: &renderFailure{Out: out, Message: err.Error()}})
	return recordErr
}

func writeRenderFailure(failure renderFailure) error {
	if failure.Out == "" {
		return fmt.Errorf("--out is required to record render-error.txt")
	}
	out, err := canonicalPath(failure.Out)
	if err != nil {
		return err
	}
	// Preserve any already-published successful render, including its evidence.
	manifest := filepath.Join(out, "render-manifest.json")
	if _, err = os.Stat(manifest); err == nil {
		data, readErr := os.ReadFile(manifest)
		receipt, verifyErr := VerifyReceipt(data)
		if readErr != nil || verifyErr != nil || !validTaskID.MatchString(failure.TaskID) || receipt.TaskID != failure.TaskID {
			return fmt.Errorf("output already contains a receipt from another or unconfirmed task; choose a new --out directory")
		}
		// The parent observed failure even if the worker published just before
		// cancellation. Remove only this exact task's authenticated receipt.
		if err = os.Remove(manifest); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		return err
	}
	for _, name := range []string{"render-error.txt", "render-error.json"} {
		path := filepath.Join(out, name)
		if info, e := os.Lstat(path); e == nil && !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", name)
		} else if e != nil && !os.IsNotExist(e) {
			return e
		}
	}
	if err = os.WriteFile(filepath.Join(out, "render-error.txt"), []byte(failure.Message+"\n"), 0644); err != nil {
		return err
	}
	data, err := json.MarshalIndent(classifyNativeFailure(failure.Message, failure.TaskID), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "render-error.json"), append(data, '\n'), 0644)
}

var validTaskID = regexp.MustCompile(`^\.native-work-[0-9a-f]{32}$`)

func newTaskID() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return ".native-work-" + hex.EncodeToString(data[:]), nil
}

// RenderWorker is the private CLI subprocess entry point. It deliberately runs
// all filesystem preparation in a killable process rather than a Go goroutine.
func RenderWorker(ctx context.Context, input io.Reader, output io.Writer) error {
	var request workerRequest
	decoder := json.NewDecoder(io.LimitReader(input, 64*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return fmt.Errorf("invalid native worker request: %w", err)
	}
	operations := 0
	if request.Render != nil {
		operations++
	}
	if request.Doctor != nil {
		operations++
	}
	if request.Cleanup != nil {
		operations++
	}
	if request.Failure != nil {
		operations++
	}
	if operations != 1 {
		return fmt.Errorf("native worker requires exactly one operation")
	}
	if request.Failure != nil {
		if err := writeRenderFailure(*request.Failure); err != nil {
			return err
		}
		return json.NewEncoder(output).Encode(map[string]bool{"recorded": true})
	}
	if request.Cleanup != nil {
		if err := cleanupExactTask(ctx, *request.Cleanup); err != nil {
			return err
		}
		return json.NewEncoder(output).Encode(map[string]bool{"cleaned": true})
	}
	if request.Doctor != nil {
		request.Doctor.taskID = request.TaskID
		return json.NewEncoder(output).Encode(doctor(ctx, *request.Doctor, nativeCommand, runtime.GOOS))
	}
	if !validTaskID.MatchString(request.TaskID) {
		return fmt.Errorf("invalid native worker task identity")
	}
	request.Render.taskID = request.TaskID
	receipt, err := render(ctx, *request.Render, nativeCommand, runtime.GOOS)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(receipt)
}

func cleanupExactTask(ctx context.Context, task cleanupTask) error {
	return cleanupExactTaskWithRunner(ctx, task, command)
}

func cleanupExactTaskWithRunner(ctx context.Context, task cleanupTask, run runner) error {
	if !validTaskID.MatchString(task.TaskID) {
		return fmt.Errorf("invalid cleanup task identity")
	}
	root, _, err := stagingConfiguration(task.StagingRoot)
	if err != nil {
		return err
	}
	root, err = canonicalPath(root)
	if err != nil {
		return err
	}
	work := filepath.Join(root, task.TaskID)
	taskPath, pdfPath := taskPresentationPaths(root, task.TaskID)
	info, err := os.Lstat(work)
	if os.IsNotExist(err) {
		for _, path := range []string{taskPath, pdfPath} {
			if _, fileErr := os.Lstat(path); !os.IsNotExist(fileErr) {
				return fmt.Errorf("exact-task staging metadata missing; unconfirmed task file retained at %s", path)
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("cleanup task path is not a task directory")
	}
	owned, ownershipErr := stableTaskOwned(root, task.TaskID)
	if ownershipErr != nil {
		return ownershipErr
	}
	if owned {
		taskPath, err = exactTaskPresentation(root, task.TaskID)
		if err != nil {
			return fmt.Errorf("could not confirm exact-task identity; retained task metadata: %w", err)
		}
	} else {
		// Older versions placed the PPTX inside the private work directory.
		taskPath = filepath.Join(work, task.TaskID+".pptx")
	}
	if info, statErr := os.Lstat(taskPath); statErr == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("cleanup task presentation is not a regular file")
		}
		scriptPath := filepath.Join(work, "cleanup.applescript")
		if err = os.WriteFile(scriptPath, exportScript, 0600); err != nil {
			return err
		}
		if runtime.GOOS == "windows" {
			_, err = windowsPowerPoint(ctx, run, work, windowsExportRequest{Action: "close", PPTX: taskPath})
		} else {
			_, err = run(ctx, "/usr/bin/osascript", taskCloseArguments(scriptPath, taskPath, pdfPath, task.TaskID, 3, owned)...)
		}
		if err != nil {
			return fmt.Errorf("could not confirm exact-task PowerPoint close; task copy retained at %s: %w", taskPath, err)
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	} else if !owned {
		for _, path := range []string{filepath.Join(root, task.TaskID+".pptx"), pdfPath} {
			if _, fileErr := os.Lstat(path); !os.IsNotExist(fileErr) {
				return fmt.Errorf("stable staging file lacks task ownership; retained at %s", path)
			}
		}
	}
	return removeTaskFiles(root, task.TaskID)
}

func workerProcess(ctx context.Context, request workerRequest) ([]byte, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, executable, "render-native-worker")
	configureWorkerProcess(cmd)
	cmd.Stdin = bytes.NewReader(data)
	cmd.WaitDelay = 2 * time.Second
	// Observations use stderr; the successful worker protocol is JSON stdout.
	// Keeping the streams separate prevents phase logs corrupting receipts.
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	output := append(stdout.Bytes(), stderr.Bytes()...)
	observation := helperObservationSummary(stderr.String())
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%snative worker deadline includes filesystem preparation and PowerPoint probes/export; pending prompts and permissions are unconfirmed: %w; helper observations: %s", observation, ctx.Err(), strings.TrimSpace(stderr.String()))
		}
		if strings.Contains(string(output), "render-native-worker") {
			return nil, fmt.Errorf("native worker unavailable in %s; the calling executable must dispatch render-native-worker to nativeexport.RenderWorker: %s", executable, strings.TrimSpace(string(output)))
		}
		return nil, fmt.Errorf("%snative worker failed: %s: %w", observation, strings.TrimSpace(string(output)), err)
	}
	return stdout.Bytes(), nil
}
