package nativeexport

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
)

func windowsDoctor(ctx context.Context, opts DoctorOptions, run runner) []Diagnostic {
	checks := []Diagnostic{{"platform", "pass", "Windows PowerPoint COM backend (experimental; requires an interactive desktop)", ""}}
	root, _, err := stagingDirectory(opts.StagingRoot)
	if err != nil {
		return append(checks, Diagnostic{"staging-writable", "fail", err.Error(), "Choose a writable local NTFS --staging-dir; use the same folder for render."})
	}
	taskID := opts.taskID
	if taskID == "" {
		taskID, err = newTaskID()
	}
	if err != nil || !validTaskID.MatchString(taskID) {
		return append(checks, Diagnostic{"powerpoint-file-access", "fail", "Could not acquire a doctor task identity", "Rerun render-doctor."})
	}
	work := filepath.Join(root, taskID)
	if err = os.Mkdir(work, 0700); err != nil {
		return append(checks, Diagnostic{"staging-writable", "fail", err.Error(), "Choose a writable --staging-dir."})
	}
	data, err := doctorPresentation()
	var taskPath, pdfPath string
	if err == nil {
		taskPath, pdfPath, err = acquireTaskFiles(root, taskID, data)
	}
	if err != nil {
		_ = removeTaskFiles(root, taskID)
		return append(checks, Diagnostic{"staging-writable", "fail", err.Error(), "Use a local NTFS folder; native task ownership requires hard links."})
	}
	checks = append(checks, Diagnostic{"staging-writable", "pass", root, ""})
	result, err := windowsNativeExport(ctx, run, work, taskPath, pdfPath, true, false)
	if err != nil {
		return append(checks, Diagnostic{"powerpoint-file-access", "fail", err.Error(), "Open desktop PowerPoint once, finish activation/setup, and inspect Protected View or policy prompts. PowerShell COM must be allowed. Rerun with the same --staging-dir."})
	}
	defer removeTaskFiles(root, taskID)
	data, err = os.ReadFile(pdfPath)
	if err == nil {
		var pages int
		pages, err = windowsPDFPageCount(data)
		if err == nil && (pages != 1 || result.Pages != 1) {
			err = fmt.Errorf("doctor expected one native PDF page")
		}
	}
	if err == nil {
		data, err = os.ReadFile(filepath.Join(work, "native-pages", "slide-001.png"))
		if err == nil {
			_, format, decodeErr := image.DecodeConfig(bytes.NewReader(data))
			if decodeErr != nil || format != "png" {
				err = fmt.Errorf("doctor native PNG is invalid: %v", decodeErr)
			}
		}
	}
	if err != nil {
		return append(checks, Diagnostic{"powerpoint-file-access", "fail", err.Error(), "Inspect PowerPoint's PDF/PNG export configuration and rerun."})
	}
	checks = append(checks, Diagnostic{"powerpoint-file-access", "pass", "PowerPoint " + result.Version + " opened a task copy and exported one PDF page and PNG in " + root, ""})
	if len(result.MissingFonts) > 0 {
		checks = append(checks, Diagnostic{"brand-fonts", "fail", "Missing: " + strings.Join(result.MissingFonts, ", "), "Install the bundled IBM Plex fonts for this Windows user, restart PowerPoint, and rerun."})
	} else {
		checks = append(checks, Diagnostic{"brand-fonts", "pass", "IBM Plex Sans and IBM Plex Mono are installed", ""})
	}
	return checks
}
