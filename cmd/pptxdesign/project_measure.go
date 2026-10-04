package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runProjectMeasure(args []string) error {
	f := flag.NewFlagSet("project measure", flag.ContinueOnError)
	path := f.String("report", "", "existing build layout-report.json; no build or source mutation")
	slides := f.String("slides", "", "comma-separated stable slide IDs or 1-based pages/ranges (for example 2,4-6)")
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
	selected, err := filterMeasurementSlides(report.Slides, *slides)
	if err != nil {
		return err
	}
	report.Slides = selected
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	return out.Encode(wmdesign.MeasureReport(report))
}

func filterMeasurementSlides(slides []wmdesign.SlideReport, selector string) ([]wmdesign.SlideReport, error) {
	if selector == "" {
		return slides, nil
	}
	selected := map[int]bool{}
	for _, raw := range strings.Split(selector, ",") {
		part := strings.TrimSpace(raw)
		if part == "" {
			return nil, fmt.Errorf("--slides contains an empty selector")
		}
		if page, err := strconv.Atoi(part); err == nil {
			if page < 1 || page > len(slides) {
				return nil, fmt.Errorf("unknown page %d in --slides (report has %d pages)", page, len(slides))
			}
			selected[page-1] = true
			continue
		}
		// Stable IDs commonly contain hyphens. Resolve an exact identity before
		// interpreting the selector as a numeric range.
		found := false
		for i, slide := range slides {
			if slide.ID == part {
				selected[i] = true
				found = true
			}
		}
		if found {
			continue
		}
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid --slides range %q; use N-M", part)
			}
			start, e1 := strconv.Atoi(rangeParts[0])
			end, e2 := strconv.Atoi(rangeParts[1])
			if e1 != nil || e2 != nil || start < 1 || end < start || end > len(slides) {
				return nil, fmt.Errorf("invalid --slides range %q for %d pages", part, len(slides))
			}
			for page := start; page <= end; page++ {
				selected[page-1] = true
			}
			continue
		}
		return nil, fmt.Errorf("unknown stable slide ID %q in --slides", part)
	}
	out := make([]wmdesign.SlideReport, 0, len(selected))
	for i, slide := range slides {
		if selected[i] {
			out = append(out, slide)
		}
	}
	return out, nil
}
