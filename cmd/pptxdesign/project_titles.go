package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type projectTitleRow struct {
	Page     int    `json:"page"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Hidden   bool   `json:"hidden"`
	Template string `json:"template"`
}

func runProjectTitles(args []string) error {
	f := flag.NewFlagSet("project titles", flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	bundle := f.String("bundle", "", "bundle path or revision; defaults to project lock")
	format := f.String("format", "text", "output format: text or json")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected project titles positional arguments")
	}
	if *format != "text" && *format != "json" {
		return fmt.Errorf("--format must be text or json")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	engine := wmdesign.CandidateEngine
	_, _, lockErr := deckproject.ReadLock(p)
	if lockErr == nil {
		*bundle, engine, err = projectRuntime(p, *bundle, "")
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(lockErr) {
		return lockErr
	} else if *bundle == "" {
		*bundle, err = publishedProjectBundle(designReleaseRoot())
		if err != nil {
			return err
		}
	}
	bundlePath := deckproject.BundlePath(*bundle)
	if validLockedBundle(*bundle) {
		bundlePath = filepath.Join(designReleaseRoot(), "library", "wm-design-system", *bundle)
	}
	compiled, err := deckproject.Compile(p, bundlePath, engine)
	if err != nil {
		return err
	}
	rows := make([]projectTitleRow, 0, len(compiled.Document.Slides))
	for i, slide := range compiled.Document.Slides {
		template := ""
		if i < len(p.Document.Slides) {
			template = p.Document.Slides[i].Template.ID
		}
		rows = append(rows, projectTitleRow{Page: i + 1, ID: slide.ID, Title: slide.Title, Hidden: slide.Hidden, Template: template})
	}
	if *format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}
	for _, row := range rows {
		fmt.Printf("%d\t%s\t%s\t%t\t%s\n", row.Page, row.ID, row.Title, row.Hidden, row.Template)
	}
	return nil
}
