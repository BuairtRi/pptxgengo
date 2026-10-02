package main

import (
	"fmt"
	"strings"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/internal/textlayout"
)

type fontDirectories []string

func (d *fontDirectories) String() string { return strings.Join(*d, ", ") }
func (d *fontDirectories) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("--font-dir cannot be empty")
	}
	*d = append(*d, value)
	return nil
}

// This path measures the original text contracts entirely in process. Go
// results never enter the native evidence/cache path or claim verification.
func runGoLayout(command, specPath, out string, roots []string) error {
	if specPath == "" {
		return fmt.Errorf("--spec required with --engine go")
	}
	var spec compose.Spec
	raw, err := readJSON(specPath, &spec)
	if err != nil {
		return err
	}
	requests, err := compose.ProbeRequests(spec)
	if err != nil {
		return err
	}
	engine, err := textlayout.New(roots)
	if err != nil {
		return err
	}
	report, err := engine.Measure(requests)
	if err != nil {
		return err
	}
	report.SpecSHA256 = hash(raw)
	switch command {
	case "measure":
		err = writeNew(out, jsonBytes(report))
	case "build":
		plan, planErr := compose.Plan(spec, report.Measurements)
		if planErr != nil {
			return fmt.Errorf("Go prototype planner: %w", planErr)
		}
		m := manifest{Schema: "pptxgengo.compose-bundle.v1", SpecSHA: hash(raw), Plan: &plan, Slides: finalSlides(spec, plan), GoLayout: &report}
		err = bundle(out, m)
	case "fit-report":
		_, planErr := compose.Plan(spec, report.Measurements)
		rows := compose.FixedTextFitReport(spec, report.Measurements)
		layoutFailures := compose.LayoutFitReport(spec, report.Measurements)
		failures := 0
		for _, row := range rows {
			if !row.Fits {
				failures++
			}
		}
		message := ""
		if planErr != nil {
			message = planErr.Error()
		}
		boundaryWarnings := 0
		heightWarnings := 0
		for _, layout := range report.Requests {
			boundaryWarnings += len(layout.BoundaryWarnings)
			heightWarnings += len(layout.HeightWarnings)
		}
		err = writeNew(out, jsonBytes(map[string]any{"schema": "pptxgengo.go-fit.v1", "engine": textlayout.Engine, "spec_sha256": hash(raw), "powerpoint_verified": false, "fixed_zone_count": len(rows), "overflow_count": failures, "layout_failure_count": len(layoutFailures), "layout_failures": layoutFailures, "zones": rows, "planner_passed": planErr == nil, "planner_error": message, "wrap_boundary_warning_count": boundaryWarnings, "height_review_warning_count": heightWarnings, "go_layout": report, "scope": "Go prototype measurements and planner fit only. Wrap and height warnings identify requests needing native review; they do not alter predicted fit. PowerPoint can wrap or space text differently; use native verify for its final rendering."}))
	default:
		return fmt.Errorf("unsupported Go layout command %q", command)
	}
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
