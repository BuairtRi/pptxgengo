package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func emitProjectDensityWarnings(root string, receipt deckproject.Receipt) error {
	rel := filepath.ToSlash(filepath.Join("builds", receipt.BuildID, "layout-report.json"))
	path, err := deckproject.SafePath(root, rel)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var report wmdesign.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		return fmt.Errorf("decode build layout report: %w", err)
	}
	emitDensityWarnings(report)
	return nil
}

// library-match publishes each passing candidate's layout report as a file,
// while its stdout remains the machine-readable match report. Read those
// reports to surface automatic density changes on stderr just like build.
func libraryMatchDensityWarningMessages(match deckproject.LibraryMatchReport) ([]string, error) {
	var warnings []string
	for _, candidate := range match.Candidates {
		if candidate.LayoutReport == "" {
			continue
		}
		raw, err := os.ReadFile(candidate.LayoutReport)
		if err != nil {
			return nil, fmt.Errorf("read library-match layout report %s: %w", candidate.LayoutReport, err)
		}
		var report wmdesign.Report
		if err := json.Unmarshal(raw, &report); err != nil {
			return nil, fmt.Errorf("decode library-match layout report %s: %w", candidate.LayoutReport, err)
		}
		warnings = append(warnings, projectDensityWarningMessages(report)...)
	}
	return warnings, nil
}

func emitLibraryMatchDensityWarnings(match deckproject.LibraryMatchReport) error {
	warnings, err := libraryMatchDensityWarningMessages(match)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
	return nil
}

func emitDensityWarnings(report wmdesign.Report) {
	for _, warning := range projectDensityWarningMessages(report) {
		fmt.Fprintln(os.Stderr, warning)
	}
}

func projectDensityWarningMessages(report wmdesign.Report) []string {
	headers := make(map[string]string, len(report.Slides))
	for _, slide := range report.Slides {
		if slide.Density != nil {
			headers[slide.ID] = slide.Density.Header
		}
	}
	warnings := make([]string, 0, len(report.DensityAdjustments))
	for _, adjustment := range report.DensityAdjustments {
		header := headers[adjustment.SlideID]
		if header == "" {
			header = "comfortable"
		}
		trace := adjustment.Requested
		for _, step := range adjustment.Steps {
			if step.To != "" {
				trace += " → " + step.To
			}
		}
		if trace == adjustment.Requested {
			trace += " → " + adjustment.Resolved
		}
		warnings = append(warnings, fmt.Sprintf("warning: Slide %s: body density changed %s to fit supplied content (header %s unchanged).", adjustment.SlideID, trace, header))
	}
	return warnings
}
