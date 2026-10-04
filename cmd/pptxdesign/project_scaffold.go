package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	if err := f.Parse(args); err != nil {
		return err
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
	if *out == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	file, err := os.OpenFile(*out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	closeErr := file.Close()
	if err != nil {
		os.Remove(*out)
		return err
	}
	return closeErr
}
