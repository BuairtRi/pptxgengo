package nativeexport

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"rsc.io/pdf"
)

//go:embed powerpoint.ps1
var windowsExportScript []byte

type windowsExportRequest struct {
	Action, PPTX, PDF, PagesDirectory string
	PNG, RequireFonts                 bool
}

type windowsExportResult struct {
	Pages        int      `json:"pages"`
	Version      string   `json:"powerpoint_version"`
	MissingFonts []string `json:"missing_fonts"`
}

func windowsPowerPoint(ctx context.Context, run runner, work string, request windowsExportRequest) ([]byte, error) {
	script := filepath.Join(work, "powerpoint.ps1")
	config := filepath.Join(work, "powerpoint-request.json")
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(script, windowsExportScript, 0600); err != nil {
		return nil, err
	}
	if err = os.WriteFile(config, data, 0600); err != nil {
		return nil, err
	}
	// Arguments and UTF-8 JSON keep spaces/Unicode/metacharacters out of code.
	// Bypass applies only to this process; machine policy is never changed.
	return run(ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-STA", "-ExecutionPolicy", "Bypass", "-File", script, "-Config", config)
}

func windowsNativeExport(ctx context.Context, run runner, work, taskPath, pdfPath string, png, requireFonts bool) (windowsExportResult, error) {
	var result windowsExportResult
	data, err := windowsPowerPoint(ctx, run, work, windowsExportRequest{Action: "export", PPTX: taskPath, PDF: pdfPath, PagesDirectory: filepath.Join(work, "native-pages"), PNG: png, RequireFonts: requireFonts})
	if err != nil {
		return result, fmt.Errorf("Windows PowerPoint COM export failed; check desktop PowerPoint for setup, Protected View or policy prompts: %w", err)
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("PowerPoint returned invalid Windows export metadata: %w", err)
	}
	if result.Pages < 1 {
		return result, fmt.Errorf("PowerPoint reported no exported Windows pages")
	}
	return result, nil
}

func windowsPDFPageCount(data []byte) (pages int, err error) {
	// The PDF reader can panic on malformed producer output. Never issue a
	// successful native receipt for a malformed or encrypted/unreadable PDF.
	defer func() {
		if r := recover(); r != nil {
			pages, err = 0, fmt.Errorf("invalid native PDF: %v", r)
		}
	}()
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("cannot inspect native PDF: %w", err)
	}
	pages = reader.NumPage()
	if pages < 1 {
		return 0, fmt.Errorf("native PDF has no readable pages")
	}
	return pages, nil
}
