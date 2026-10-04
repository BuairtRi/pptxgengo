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
	patch := f.String("patch", "", "YAML or JSON mapping keyed by stable slide ID; fields: template, values, brief (project-relative page brief file path)")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to project lock")
	engine := f.String("engine", "", "engine; defaults to project lock")
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
		if !validPublishedBundle(*bundle) {
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
	result, err := deckproject.EditSlides(p, edits, deckproject.BundlePath(*bundle), *engine)
	if err != nil {
		return err
	}
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	return out.Encode(result)
}
