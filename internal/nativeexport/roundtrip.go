package nativeexport

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/internal/powershellenv"
)

// RoundTripEdit describes one whole plain-text fixture shape, identified by its
// existing lineage tags. This is desktop qualification tooling, not an editor
// for customer decks or a source adoption API.
type RoundTripEdit struct {
	ShapeToken string `json:"shape_token"`
	SlideToken string `json:"slide_token"`
	Before     string `json:"before"`
	After      string `json:"after"`
}

type RoundTripPlan struct {
	DeckToken      string          `json:"deck_token"`
	BuildToken     string          `json:"build_token"`
	MoveSlideToken string          `json:"move_slide_token"`
	Edits          []RoundTripEdit `json:"edits"`
}

type RoundTripExecution struct {
	Schema            string `json:"schema"`
	Platform          string `json:"platform"`
	PowerPointVersion string `json:"powerpoint_version"`
	OSVersion         string `json:"os_version"`
	InputSHA256       string `json:"input_sha256"`
	SavedAsSHA256     string `json:"saved_as_sha256"`
	EditedSHA256      string `json:"edited_sha256"`
	Started           string `json:"started"`
	Finished          string `json:"finished"`
	SavedAs           string `json:"saved_as"`
	Edited            string `json:"edited"`
	Closed            bool   `json:"closed"`
}

type roundTripRequest struct {
	Action  string
	Input   string
	SavedAs string
	Edited  string
	RoundTripPlan
}

//go:embed roundtrip.ps1
var roundTripScript []byte

// WindowsRoundTrip creates a NEW private working directory, opens only a copy
// of the supplied fixture, saves it under a new name, makes three text edits,
// reorders one slide and saves again. It never quits or kills PowerPoint. All
// outputs, including failures, remain available for inspection. The caller must
// independently verify tags, actual saved text/order and immutable baselines.
func WindowsRoundTrip(ctx context.Context, destination string, fixture []byte, plan RoundTripPlan) (RoundTripExecution, error) {
	if runtime.GOOS != "windows" {
		return RoundTripExecution{}, fmt.Errorf("round-trip COM execution requires an interactive Windows desktop")
	}
	return windowsRoundTrip(ctx, destination, fixture, plan, roundTripCommand)
}

func roundTripCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if strings.EqualFold(filepath.Base(name), "powershell.exe") {
		cmd.Env = powershellenv.ForWindowsPowerShell(os.Environ())
	}
	configureWorkerProcess(cmd)
	cmd.WaitDelay = 2 * time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("round-trip helper: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func roundTripToken(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

func windowsRoundTrip(ctx context.Context, destination string, fixture []byte, plan RoundTripPlan, run runner) (RoundTripExecution, error) {
	var result RoundTripExecution
	if len(fixture) == 0 || len(fixture) > 512<<20 || len(plan.Edits) != 3 || !roundTripToken(plan.DeckToken) || !roundTripToken(plan.BuildToken) || !roundTripToken(plan.MoveSlideToken) {
		return result, fmt.Errorf("invalid round-trip fixture or plan")
	}
	seen := map[string]bool{}
	for _, e := range plan.Edits {
		if !roundTripToken(e.ShapeToken) || !roundTripToken(e.SlideToken) || seen[e.ShapeToken] || !utf8.ValidString(e.Before+e.After) || e.Before == e.After || len(e.Before) > 64<<10 || len(e.After) == 0 || len(e.After) > 64<<10 || strings.ContainsAny(e.Before+e.After, "\x00\r\n\v") {
			return result, fmt.Errorf("round-trip requires three distinct single-paragraph plain-text shapes")
		}
		seen[e.ShapeToken] = true
	}
	dir, err := canonicalPath(destination)
	if err != nil {
		return result, err
	}
	if err = os.Mkdir(dir, 0700); err != nil {
		return result, err
	}
	req := roundTripRequest{Action: "edit", Input: filepath.Join(dir, "input.pptx"), SavedAs: filepath.Join(dir, "saved-as.pptx"), Edited: filepath.Join(dir, "edited.pptx"), RoundTripPlan: plan}
	if err = os.WriteFile(req.Input, fixture, 0600); err != nil {
		return result, err
	}
	if err = os.WriteFile(filepath.Join(dir, "roundtrip.ps1"), roundTripScript, 0600); err != nil {
		return result, err
	}
	call := func(callCtx context.Context, request roundTripRequest, name string) ([]byte, error) {
		raw, err := json.Marshal(request)
		if err != nil {
			return nil, err
		}
		config := filepath.Join(dir, name)
		if err = os.WriteFile(config, raw, 0600); err != nil {
			return nil, err
		}
		return run(callCtx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-STA", "-ExecutionPolicy", "Bypass", "-File", filepath.Join(dir, "roundtrip.ps1"), "-Config", config)
	}
	started := time.Now().UTC().Format(time.RFC3339Nano)
	deadline, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	raw, err := call(deadline, req, "request.json")
	if err == nil {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&result)
		if err == nil {
			var extra any
			if decoder.Decode(&extra) != io.EOF {
				err = fmt.Errorf("trailing round-trip execution metadata")
			}
		}
		if err == nil && (result.Schema != "pptxgengo.windows-roundtrip-execution.v1" || result.PowerPointVersion == "" || result.OSVersion == "" || !result.Closed || result.SavedAs != req.SavedAs || result.Edited != req.Edited) {
			err = fmt.Errorf("incomplete round-trip execution metadata")
		}
		if err == nil {
			result.InputSHA256, err = roundTripFileHash(req.Input)
		}
		if err == nil && result.InputSHA256 != fmt.Sprintf("%x", sha256.Sum256(fixture)) {
			err = fmt.Errorf("owned fixture input changed")
		}
		if err == nil {
			result.SavedAsSHA256, err = roundTripFileHash(req.SavedAs)
		}
		if err == nil {
			result.EditedSHA256, err = roundTripFileHash(req.Edited)
		}
	}
	if err != nil {
		// A COM call can outlive the helper deadline. Retry close only for this
		// new fixture's three exact paths, with a separate bounded context.
		req.Action = "close"
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_, closeErr := call(cleanup, req, "cleanup-request.json")
		if closeErr != nil {
			err = fmt.Errorf("%w; fixture close unconfirmed: %v; retain %s for inspection", err, closeErr, dir)
		}
		_ = os.WriteFile(filepath.Join(dir, "failure.txt"), []byte(err.Error()+"\n"), 0600)
		return result, err
	}
	result.Platform = "windows"
	result.Started = started
	result.Finished = time.Now().UTC().Format(time.RFC3339Nano)
	data, _ := json.MarshalIndent(result, "", "  ")
	if err = os.WriteFile(filepath.Join(dir, "execution.json"), append(data, '\n'), 0600); err != nil {
		return result, err
	}
	return result, nil
}

func roundTripFileHash(path string) (string, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !before.Mode().IsRegular() || before.Size() == 0 || before.Size() > 512<<20 {
		return "", fmt.Errorf("invalid owned round-trip output: %s", filepath.Base(path))
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !os.SameFile(before, opened) {
		return "", fmt.Errorf("owned round-trip output identity changed")
	}
	h := sha256.New()
	size, err := io.Copy(h, io.LimitReader(f, (512<<20)+1))
	if err != nil {
		return "", err
	}
	after, err := f.Stat()
	if err != nil {
		return "", err
	}
	current, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if size != before.Size() || size > 512<<20 || !os.SameFile(before, current) || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return "", fmt.Errorf("owned round-trip output changed during verification")
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
