package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectEdit(args []string) error {
	f := flag.NewFlagSet("project edit", flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	patch := f.String("patch", "", "YAML or JSON mapping keyed by stable slide ID; fields: template, values, content, bindings, brief (project-relative page brief file path)")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to project lock")
	engine := f.String("engine", "", "engine; defaults to project lock")
	checkFit := f.Bool("check-fit", false, "measure changed slides with the Go renderer before committing; native review remains required")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *patch == "" {
		return fmt.Errorf("project edit requires --patch FILE with stable-ID keyed slide edits")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	lock, _, err := deckproject.ReadLock(p)
	if err != nil {
		return err
	}
	if *bundle == "" {
		*bundle = strings.TrimPrefix(lock.BundleRevision, "wmds-library.")
		if !validLockedBundle(*bundle) {
			return fmt.Errorf("unsupported locked bundle revision %q", lock.BundleRevision)
		}
		pinned := filepath.Join(p.Root, "runtime", "library", "wm-design-system", "pinned")
		if stat, e := os.Stat(pinned); e == nil && stat.IsDir() {
			*bundle = pinned
		}
	}
	if *engine == "" {
		*engine = lock.Engine
	}
	raw, err := os.ReadFile(*patch)
	if err != nil {
		return err
	}
	edits, err := deckproject.DecodeSlideEdits(raw, *patch)
	if err != nil {
		return err
	}
	b := deckproject.BundlePath(*bundle)
	if validLockedBundle(*bundle) {
		b = filepath.Join(designReleaseRoot(), "library", "wm-design-system", *bundle)
	}
	result, err := deckproject.EditSlidesWithOptions(p, edits, b, *engine, deckproject.EditOptions{CheckFit: *checkFit})
	if err != nil {
		if strings.Contains(err.Error(), "binding.unsupported_field") {
			return fmt.Errorf("%w; replace complete values/content/bindings for the new template (remove stale values.keys, or supply values: {} with complete content/bindings), or use project swap to review a mapping proposal", err)
		}
		return err
	}
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	return out.Encode(result)
}
