package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectMigrate(args []string) error {
	f := flag.NewFlagSet("project migrate", flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	bundle := f.String("bundle", currentDesignBundle, "target bundle revision or path")
	engine := f.String("engine", "", "target engine (defaults to existing lock)")
	dryRun := f.Bool("dry-run", false, "validate target without changing the project")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *bundle == "" {
		return fmt.Errorf("project migrate requires a nonempty target bundle and no positional arguments")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	lock, _, err := deckproject.ReadLock(p)
	if err != nil {
		return fmt.Errorf("migration requires an existing lock; use project init for a new project: %w", err)
	}
	if *engine == "" {
		*engine = lock.Engine
	}
	b := deckproject.BundlePath(*bundle)
	if validLockedBundle(*bundle) {
		b = designBundlePath(*bundle)
	}
	r, err := deckproject.Migrate(p, b, *engine, *dryRun)
	if err != nil {
		return err
	}
	for _, adjustment := range r.DensityAdjustments {
		fmt.Fprintf(os.Stderr, "warning: Slide %s: body density changed %s → %s to fit supplied content.\n", adjustment.SlideID, adjustment.Requested, adjustment.Resolved)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
