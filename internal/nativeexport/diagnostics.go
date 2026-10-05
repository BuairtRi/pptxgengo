package nativeexport

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/pptx"
)

const powerPointApp = "/Applications/Microsoft PowerPoint.app"

//go:embed dialog.swift
var dialogScript []byte

// A read-only optional probe never requests Accessibility or Screen Recording access.
func nativeCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	if name == "/usr/bin/osascript" && monitorAppleScript(args) {
		path := filepath.Join(filepath.Dir(args[0]), "dialog.swift")
		if err := os.WriteFile(path, dialogScript, 0600); err == nil {
			return monitoredCommand(ctx, func(child context.Context) ([]byte, error) {
				return command(child, name, args...)
			}, func(child context.Context, observations chan<- dialogObservation) {
				observeDialogs(child, path, observations)
			})
		}
	}
	return command(ctx, name, args...)
}

func monitorAppleScript(args []string) bool {
	// Only private task scripts have an absolute file path. Inline doctor
	// commands start with -e and must not write probe files into the caller's cwd.
	return len(args) >= 5 && filepath.IsAbs(args[0]) && filepath.Ext(args[0]) == ".applescript" && (len(args) == 5 || args[5] != "close")
}

type dialogObservation struct{ Status, Dialog, Detail string }

func observeDialogs(ctx context.Context, path string, observations chan<- dialogObservation) {
	cmd := exec.CommandContext(ctx, "/usr/bin/swift", path, "--watch")
	cmd.WaitDelay = 2 * time.Second
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var result dialogObservation
		if json.Unmarshal(scanner.Bytes(), &result) == nil {
			select {
			case observations <- result:
			case <-ctx.Done():
			}
		}
	}
	_ = cmd.Wait()
}

// The visibility monitor runs throughout the open/export event, rather than
// checking once before a new file-access prompt could appear. It never requests
// Accessibility/Screen Recording access. Unknown visibility is not a denial.
func monitoredCommand(ctx context.Context, export func(context.Context) ([]byte, error), observe func(context.Context, chan<- dialogObservation)) ([]byte, error) {
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		data []byte
		err  error
	}
	completed := make(chan result, 1)
	observations := make(chan dialogObservation, 1)
	monitorDone := make(chan struct{})
	go func() { data, err := export(child); completed <- result{data, err} }()
	go func() { defer close(monitorDone); observe(child, observations) }()
	defer func() { cancel(); <-monitorDone }()
	for {
		select {
		case result := <-completed:
			return result.data, result.err
		case observation := <-observations:
			if observation.Status == "blocked" {
				cancel()
				<-completed
				return nil, fmt.Errorf("file_access_denied: PowerPoint is showing %s: %s; click Select/Grant for the displayed staging folder, then rerun render-doctor or render", observation.Dialog, observation.Detail)
			}
		}
	}
}

// canonicalPath resolves even an output path whose last components do not exist.
func canonicalPath(name string) (string, error) {
	abs, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	current := abs
	var suffix []string
	for {
		resolved, e := filepath.EvalSymlinks(current)
		if e == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(e) {
			return "", e
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", e
		}
		suffix = append(suffix, filepath.Base(current))
		current = parent
	}
}

func stagingDirectory(configured string) (string, string, error) {
	configured, kind, err := stagingConfiguration(configured)
	if err != nil {
		return "", "", err
	}
	root, err := canonicalPath(configured)
	if err != nil {
		return "", "", err
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		return "", "", fmt.Errorf("staging folder %s is unavailable: %w; choose --staging-dir with PowerPoint file access", root, err)
	}
	return root, kind, nil
}

func stagingConfiguration(configured string) (string, string, error) {
	if configured == "" {
		configured = os.Getenv("PPTXGENGO_NATIVE_STAGING")
	}
	kind := "configured; PowerPoint access unverified until export"
	if configured == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", err
		}
		configured = filepath.Join(home, "Library/Caches/pptxgengo/native")
		kind = "user cache; PowerPoint access unverified until export"
	}
	return configured, kind, nil
}

type Diagnostic struct {
	Check  string `json:"check"`
	Status string `json:"status"`
	Detail string `json:"detail"`
	Fix    string `json:"fix,omitempty"`
}
type DoctorOptions struct {
	StagingRoot string
	Timeout     time.Duration
	taskID      string
}

// Doctor reports observed permission failures separately from checks it cannot prove.
func Doctor(ctx context.Context, opts DoctorOptions) []Diagnostic {
	if opts.Timeout <= 0 {
		opts.Timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	taskID, err := newTaskID()
	if err != nil {
		return []Diagnostic{{Check: "doctor-worker", Status: "fail", Detail: err.Error()}}
	}
	data, err := workerProcess(ctx, workerRequest{Doctor: &opts, TaskID: taskID})
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	_, cleanupErr := workerProcess(cleanupCtx, workerRequest{Cleanup: &cleanupTask{StagingRoot: opts.StagingRoot, TaskID: taskID}})
	cleanupCancel()
	if err != nil {
		return []Diagnostic{{Check: "doctor-worker", Status: "fail", Detail: err.Error(), Fix: "Inspect PowerPoint and macOS for a pending prompt; try a qualified --staging-dir from the signed-in desktop session."}}
	}
	var checks []Diagnostic
	if err = json.Unmarshal(data, &checks); err != nil {
		return []Diagnostic{{Check: "doctor-worker", Status: "fail", Detail: "Invalid native diagnostic metadata: " + err.Error()}}
	}
	if cleanupErr != nil {
		checks = append(checks, Diagnostic{"doctor-cleanup", "unknown", cleanupErr.Error(), "Inspect only the reported doctor task copy in PowerPoint; close it without saving, then rerun render-doctor."})
	}
	return checks
}
func doctor(ctx context.Context, opts DoctorOptions, run runner, platform string) []Diagnostic {
	if opts.Timeout <= 0 {
		opts.Timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	checks := []Diagnostic{}
	if platform != "darwin" {
		return []Diagnostic{{Check: "platform", Status: "fail", Detail: "native rendering requires macOS", Fix: "Use a logged-in Mac with Microsoft PowerPoint installed."}}
	}
	if _, err := os.Stat(powerPointApp); err != nil {
		checks = append(checks, Diagnostic{"powerpoint", "fail", err.Error(), "Install Microsoft PowerPoint at /Applications/Microsoft PowerPoint.app."})
	} else {
		data, err := run(ctx, "/usr/libexec/PlistBuddy", "-c", "Print :CFBundleShortVersionString", filepath.Join(powerPointApp, "Contents/Info.plist"))
		if err != nil {
			checks = append(checks, Diagnostic{"powerpoint", "unknown", powerPointApp + ": version lookup failed: " + err.Error(), "Open PowerPoint once and finish its setup."})
		} else {
			checks = append(checks, Diagnostic{"powerpoint", "pass", powerPointApp + " version " + strings.TrimSpace(string(data)), ""})
		}
	}
	data, err := run(ctx, "/usr/bin/stat", "-f", "%Su", "/dev/console")
	if err != nil {
		checks = append(checks, Diagnostic{"gui-session", "unknown", err.Error(), "Run from the signed-in macOS desktop session."})
	} else {
		owner := strings.TrimSpace(string(data))
		if owner == "root" || owner == "loginwindow" || owner == "" {
			checks = append(checks, Diagnostic{"gui-session", "fail", "Console owner: " + owner, "Sign into the macOS desktop."})
		} else {
			checks = append(checks, Diagnostic{"gui-session", "pass", "Console owner: " + owner, ""})
		}
	}
	root, kind, err := stagingDirectory(opts.StagingRoot)
	if err != nil {
		checks = append(checks, Diagnostic{"staging-writable", "fail", err.Error(), "Choose a writable --staging-dir."})
	} else {
		f, e := os.CreateTemp(root, ".doctor-")
		if e != nil {
			checks = append(checks, Diagnostic{"staging-writable", "fail", e.Error(), "Choose a writable --staging-dir."})
		} else {
			name := f.Name()
			f.Close()
			os.Remove(name)
			checks = append(checks, Diagnostic{"staging-writable", "pass", root, ""})
		}
	}
	probeCtx, probeCancel := context.WithTimeout(ctx, 5*time.Second)
	data, err = run(probeCtx, "/usr/bin/osascript", "-e", `with timeout of 3 seconds`, "-e", `tell application "/Applications/Microsoft PowerPoint.app" to count presentations`, "-e", `end timeout`)
	probeCancel()
	if err != nil {
		status := "unknown"
		fix := "Open PowerPoint from the signed-in macOS desktop and rerun from that same user session."
		if strings.Contains(err.Error(), "-1743") || strings.Contains(strings.ToLower(err.Error()), "not authorized") {
			status = "fail"
			fix = "Allow the calling terminal or agent host in System Settings > Privacy & Security > Automation > Microsoft PowerPoint; then rerun."
		}
		checks = append(checks, Diagnostic{"powerpoint-dispatch", "fail", "PowerPoint did not answer an operational Apple event: " + err.Error(), fix}, Diagnostic{"automation", status, "Operational delivery failed; permission is unconfirmed unless explicitly denied.", fix})
	} else {
		checks = append(checks, Diagnostic{"powerpoint-dispatch", "pass", "PowerPoint answered; open presentations: " + strings.TrimSpace(string(data)), ""}, Diagnostic{"automation", "pass", "PowerPoint answered an operational Apple event from this caller.", ""})
	}
	if err != nil || root == "" {
		checks = append(checks, Diagnostic{"powerpoint-file-access", "unknown", "No file-access probe was sent because staging or operational Apple-event delivery is unconfirmed. " + kind, "Resolve the failed staging/dispatch check, then rerun render-doctor. If PowerPoint displays Grant File Access, click Select/Grant for the staging folder: " + root})
	} else {
		checks = append(checks, doctorFileAccess(ctx, root, opts.taskID, run))
	}
	data, err = run(ctx, "/usr/bin/swift", "-e", `import AppKit; import PDFKit; print("Swift, AppKit and PDFKit available"); print("GUI_APP_COUNT:\(NSWorkspace.shared.runningApplications.count)")`)
	if err != nil {
		checks = append(checks, Diagnostic{"swift-pdfkit", "fail", err.Error(), "Install or select Apple's command line developer tools (xcode-select --install)."})
	} else {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		checks = append(checks, Diagnostic{"swift-pdfkit", "pass", lines[0], ""})
		for _, line := range lines {
			if strings.HasPrefix(line, "GUI_APP_COUNT:") {
				if line == "GUI_APP_COUNT:0" {
					checks = append(checks, Diagnostic{"gui-caller", "unknown", "NSWorkspace exposes no running GUI applications to this caller; a signed-in console alone does not establish GUI dispatch.", "Rerun from the signed-in desktop user session and check PowerPoint operational dispatch."})
				} else {
					checks = append(checks, Diagnostic{"gui-caller", "pass", "NSWorkspace exposes running GUI applications: " + strings.TrimPrefix(line, "GUI_APP_COUNT:"), ""})
				}
			}
		}
	}
	return checks
}

func doctorFileAccess(ctx context.Context, root, taskID string, run runner) Diagnostic {
	fix := "In PowerPoint, click Select/Grant if Grant File Access appears for " + root + "; select the displayed folder and rerun render-doctor. The CLI never grants permissions automatically."
	if taskID == "" {
		taskID, _ = newTaskID()
	}
	if !validTaskID.MatchString(taskID) {
		return Diagnostic{"powerpoint-file-access", "unknown", "Invalid doctor task identity", fix}
	}
	work := filepath.Join(root, taskID)
	if err := os.Mkdir(work, 0700); err != nil {
		return Diagnostic{"powerpoint-file-access", "unknown", err.Error(), fix}
	}
	path, pdfPath := taskPresentationPaths(root, taskID)
	script := filepath.Join(work, "probe.applescript")
	presentation := pptx.New()
	if err := presentation.AddSlide().AddText([]pptx.TextProps{{Text: "PowerPoint staging file-access diagnostic"}}, nil); err != nil {
		_ = removeTaskFiles(root, taskID)
		return Diagnostic{"powerpoint-file-access", "unknown", err.Error(), fix}
	}
	data, err := presentation.Write()
	if err == nil {
		path, pdfPath, err = acquireTaskFiles(root, taskID, data)
	}
	if err == nil {
		err = os.WriteFile(script, exportScript, 0600)
	}
	if err != nil {
		_ = removeTaskFiles(root, taskID)
		return Diagnostic{"powerpoint-file-access", "unknown", err.Error(), fix}
	}
	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	if err = validateTaskExport(root, taskID); err != nil {
		cancel()
		_ = removeTaskFiles(root, taskID)
		return Diagnostic{"powerpoint-file-access", "unknown", err.Error(), fix}
	}
	_, err = run(probeCtx, "/usr/bin/osascript", taskExportArguments(script, path, pdfPath, taskID, 12)...)
	cancel()
	if err == nil {
		pdf, readErr := os.ReadFile(pdfPath)
		_ = removeTaskFiles(root, taskID)
		if readErr != nil || !strings.HasPrefix(string(pdf), "%PDF-") {
			return Diagnostic{"powerpoint-file-access", "unknown", fmt.Sprintf("PowerPoint command returned, but a native PDF write in %s was not confirmed: %v", root, readErr), fix}
		}
		return Diagnostic{"powerpoint-file-access", "pass", "PowerPoint opened the exact diagnostic copy, wrote its PDF, and closed it in the same stable staging folder used by render: " + root + ". This confirms this probe only; future prompts or a different --staging-dir require a new check. No render receipt was issued.", ""}
	}
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, closeErr := run(closeCtx, "/usr/bin/osascript", taskCloseArguments(script, path, pdfPath, taskID, 3, true)...)
	closeCancel()
	detail := exportFailure(err, path).Error()
	if closeErr == nil {
		_ = removeTaskFiles(root, taskID)
	} else {
		detail += "; exact-task close unconfirmed; retained diagnostic copy: " + path
	}
	status := "unknown"
	if strings.Contains(strings.ToLower(err.Error()), "file_access_denied") {
		status = "fail"
	}
	return Diagnostic{"powerpoint-file-access", status, detail, fix}
}

func exportFailure(err error, taskPath string) error {
	message := strings.ToLower(err.Error())
	stagingRoot := filepath.Dir(taskPath)
	switch {
	case strings.Contains(message, "-10827"):
		return fmt.Errorf("PowerPoint PDF export failed: application_dispatch_failed (-10827): this caller could not send an operational command to PowerPoint; file-access and Automation permission are unconfirmed. Open PowerPoint from the signed-in macOS desktop and rerun render-doctor from that same user session: %w", err)
	case strings.Contains(message, "-1743") || strings.Contains(message, "not authorized") || strings.Contains(message, "automation denied"):
		return fmt.Errorf("PowerPoint PDF export failed: automation_denied: allow the calling terminal or agent host in System Settings > Privacy & Security > Automation > Microsoft PowerPoint: %w", err)
	case strings.Contains(message, "file_access_denied") || strings.Contains(message, "permission denied"):
		return fmt.Errorf("PowerPoint PDF export failed: file_access_denied for %s; select/grant PowerPoint access to the stable staging folder %s, then rerun render-doctor with that same --staging-dir: %w", taskPath, stagingRoot, err)
	case strings.Contains(message, "identity_ambiguous"):
		return fmt.Errorf("PowerPoint PDF export failed: multiple presentations identify the exact task copy; close duplicate task copies and rerun: %w", err)
	case strings.Contains(message, "open_identity_timeout") || strings.Contains(message, "-1712") || strings.Contains(message, "deadline exceeded"):
		return fmt.Errorf("PowerPoint PDF export failed: PowerPoint did not answer or identify the task copy within the bounded wait; a blocking dialog or file-access prompt is possible but unconfirmed. Inspect PowerPoint, grant access to the stable staging folder %s if requested, then rerun: %w", stagingRoot, err)
	default:
		return fmt.Errorf("PowerPoint PDF export failed: %w", err)
	}
}
