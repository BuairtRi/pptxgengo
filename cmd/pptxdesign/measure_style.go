package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type measuredStyle struct {
	Status    string              `json:"status"`
	NativeFit string              `json:"native_fit"`
	Density   string              `json:"density,omitempty"`
	Scope     string              `json:"scope"`
	Layout    wmdesign.TextLayout `json:"layout"`
}

func runMeasureStyle(args []string) error {
	f := flag.NewFlagSet("measure-style", flag.ContinueOnError)
	bundle := f.String("bundle", currentDesignBundle, "pinned foundation bundle path or v11")
	engine := f.String("engine", wmdesign.CandidateEngine, "typography engine")
	sourcePath := f.String("source", "", "optional WMDS source override; must match pinned snapshot")
	text := f.String("text", "", "text to shape")
	styleID := f.String("style", "", "source typography role to measure")
	width := f.Float64("width", 0, "available line width in points")
	density := f.String("density", "", "optional density: comfortable, compact, or dense")
	scope := f.String("scope", "body", "style scope: body, header, or cell")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *text == "" {
		return fmt.Errorf("--text is required")
	}
	if *styleID == "" {
		return fmt.Errorf("--style is required")
	}
	if *width <= 0 {
		return fmt.Errorf("--width must be greater than zero")
	}
	if validLockedBundle(*bundle) {
		*bundle = designBundlePath(*bundle)
	}
	source, err := wmdesign.Load(*bundle, *sourcePath)
	if err != nil {
		return err
	}
	style := wmdesign.Style{}
	resolvedDensity := ""
	if *density == "" && *scope == "body" {
		// Keep the no-option path identical to the original source-style lookup.
		style, err = source.Style(*styleID)
	} else {
		level := *density
		if level == "" {
			level = "comfortable"
		}
		style, err = source.StyleForDensity(*styleID, level, *scope)
		resolvedDensity = level
	}
	if err != nil {
		return err
	}
	fontRoot := filepath.Join(*bundle, "fonts")
	typography, err := wmdesign.NewSourceTypographyEngine(source, fontRoot, *engine)
	if err != nil {
		return err
	}
	layout, err := typography.Measure(*text, style, *width)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(measuredStyle{
		Status:    "measurement_only",
		NativeFit: "not_evaluated",
		Density:   resolvedDensity,
		Scope:     *scope,
		Layout:    layout,
	})
}
