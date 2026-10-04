package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runProjectScaffold(args []string) error {
	f := flag.NewFlagSet("project scaffold", flag.ContinueOnError)
	bundle := f.String("bundle", "", "shared bundle path or v1/v2/v3/v4/v5")
	key := f.String("template", "", "actual shared template key")
	engine := f.String("engine", wmdesign.CandidateEngine, "build engine for source scene measurement")
	reason := f.String("reason", "", "reason for adapting the shared topology")
	year := f.Int("year", 2026, "source specimen year")
	out := f.String("out", "", "optional new JSON file; default stdout")
	omit := f.String("omit-nodes", "", "comma-separated source node IDs deliberately excluded from the local derivative; requires adaptation reason")
	stock := f.Bool("stock", false, "emit editable slide YAML using the unchanged shared template")
	id := f.String("id", "", "slide ID for --stock")
	if err := f.Parse(args); err != nil {
		return err
	}
	if *stock {
		if f.NArg() != 0 || *key == "" || *id == "" || *reason != "" || *omit != "" {
			return fmt.Errorf("project scaffold --stock requires --template --id; adaptation flags are not applicable")
		}
		if *bundle == "" {
			*bundle = "v5"
		}
		b := deckproject.BundlePath(*bundle)
		if validLockedBundle(*bundle) {
			b = filepath.Join(designReleaseRoot(), "library", "wm-design-system", *bundle)
		}
		result, err := deckproject.StockScaffoldSlide(b, *key, *id, *year)
		if err != nil {
			return err
		}
		data, err := deckproject.MarshalSlideSource(result)
		if err != nil {
			return err
		}
		data = append([]byte("# Shared template: layout is unchanged. Replace example copy under content.\n# Set content_kind to supplied_content after replacing the examples.\n"), data...)
		return writeScaffoldOutput(*out, data)
	}
	if f.NArg() != 0 || *bundle == "" || *key == "" || *reason == "" {
		return fmt.Errorf("project scaffold requires --bundle --template --reason; it emits an uninstalled local template")
	}
	var omitNodes []string
	if *omit != "" {
		for _, id := range strings.Split(*omit, ",") {
			omitNodes = append(omitNodes, strings.TrimSpace(id))
		}
	}
	result, err := deckproject.ScaffoldTemplate(deckproject.BundlePath(*bundle), *key, *engine, *reason, *year, omitNodes...)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeScaffoldOutput(*out, data)
}

func writeScaffoldOutput(path string, data []byte) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	closeErr := file.Close()
	if err != nil {
		os.Remove(path)
		return err
	}
	return closeErr
}
