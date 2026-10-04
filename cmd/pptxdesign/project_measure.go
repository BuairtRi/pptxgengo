package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runProjectMeasure(args []string) error {
	f := flag.NewFlagSet("project measure", flag.ContinueOnError)
	path := f.String("report", "", "existing build layout-report.json; no build or source mutation")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *path == "" {
		return fmt.Errorf("project measure requires --report PATH-TO-LAYOUT-REPORT")
	}
	raw, err := os.ReadFile(*path)
	if err != nil {
		return err
	}
	var report wmdesign.Report
	if err = json.Unmarshal(raw, &report); err != nil {
		return err
	}
	if len(report.Slides) == 0 || report.Engine == "" || report.Schema == "" {
		return fmt.Errorf("measurement requires a renderer layout report with slides and engine")
	}
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	return out.Encode(wmdesign.MeasureReport(report))
}
