package nativeexport

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const powerPointApp = "/Applications/Microsoft PowerPoint.app"

//go:embed dialog.swift
var dialogScript []byte

// A read-only optional probe never requests Accessibility or Screen Recording access.
func nativeCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	if name == "/usr/bin/osascript" && len(args) == 5 {
		path := filepath.Join(filepath.Dir(args[0]), "dialog.swift")
		if err := os.WriteFile(path, dialogScript, 0600); err == nil {
			probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			data, probeErr := command(probeCtx, "/usr/bin/swift", path)
			cancel()
			if probeErr == nil {
				var result struct{ Status, Dialog, Detail string }
				if json.Unmarshal(data, &result) == nil && result.Status == "blocked" {
					return nil, fmt.Errorf("file_access_denied: PowerPoint is showing %s: %s; click Select/Grant for the displayed folder, then rerun", result.Dialog, result.Detail)
				}
			}
		}
	}
	return command(ctx, name, args...)
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
}

// Doctor reports observed permission failures separately from checks it cannot prove.
func Doctor(ctx context.Context, opts DoctorOptions) []Diagnostic {
	if opts.Timeout <= 0 {
		opts.Timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	data, err := workerProcess(ctx, workerRequest{Doctor: &opts})
	if err != nil {
		return []Diagnostic{{Check: "doctor-worker", Status: "fail", Detail: err.Error(), Fix: "Inspect PowerPoint and macOS for a pending prompt; try a qualified --staging-dir from the signed-in desktop session."}}
	}
	var checks []Diagnostic
	if err = json.Unmarshal(data, &checks); err != nil {
		return []Diagnostic{{Check: "doctor-worker", Status: "fail", Detail: "Invalid native diagnostic metadata: " + err.Error()}}
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
			checks = append(checks, Diagnostic{"staging-writable", "pass", root, ""}, Diagnostic{"powerpoint-file-access", "unknown", kind, "Run a small native render to qualify this folder; if PowerPoint asks for access, grant access and rerun."})
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

func exportFailure(err error, taskPath string) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "-10827"):
		return fmt.Errorf("PowerPoint PDF export failed: application_dispatch_failed (-10827): this caller could not send an operational command to PowerPoint; file-access and Automation permission are unconfirmed. Open PowerPoint from the signed-in macOS desktop and rerun render-doctor from that same user session: %w", err)
	case strings.Contains(message, "-1743") || strings.Contains(message, "not authorized") || strings.Contains(message, "automation denied"):
		return fmt.Errorf("PowerPoint PDF export failed: automation_denied: allow the calling terminal or agent host in System Settings > Privacy & Security > Automation > Microsoft PowerPoint: %w", err)
	case strings.Contains(message, "file_access_denied") || strings.Contains(message, "permission denied"):
		return fmt.Errorf("PowerPoint PDF export failed: file_access_denied for %s; grant PowerPoint access to the staging folder, then rerun: %w", taskPath, err)
	case strings.Contains(message, "identity_ambiguous"):
		return fmt.Errorf("PowerPoint PDF export failed: multiple presentations identify the exact task copy; close duplicate task copies and rerun: %w", err)
	case strings.Contains(message, "open_identity_timeout") || strings.Contains(message, "-1712") || strings.Contains(message, "deadline exceeded"):
		return fmt.Errorf("PowerPoint PDF export failed: PowerPoint did not answer or identify the task copy within the bounded wait; a blocking dialog or file-access prompt is possible but unconfirmed. Inspect PowerPoint, grant access to %s if requested, then rerun: %w", taskPath, err)
	default:
		return fmt.Errorf("PowerPoint PDF export failed: %w", err)
	}
}
