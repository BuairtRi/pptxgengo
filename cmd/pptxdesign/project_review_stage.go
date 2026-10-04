package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func hasProjectFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name || strings.HasPrefix(arg, name+"=") {
			return true
		}
	}
	return false
}

func runProjectReviewStage(args []string) error {
	f := flag.NewFlagSet("project review", flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	stage := f.String("stage", "deck", "outline, content or deck")
	out := f.String("out", "", "new reviewer packet directory containing index.html")
	bundle := f.String("bundle", "", "bundle path/revision, defaults to lock")
	engine := f.String("engine", "", "engine, defaults to lock")
	audience := f.Bool("audience", false, "independent audience packet: visible copy/pages and audience context; omit internal material and prior verdicts")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *out == "" {
		return fmt.Errorf("review requires --stage outline|content|deck --out NEW-DIR")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	b, e, err := projectRuntime(p, *bundle, *engine)
	if err != nil {
		return err
	}
	result, err := deckproject.ProjectReview(p, deckproject.ProjectReviewOptions{Stage: *stage, Out: *out, Bundle: b, Engine: e, Audience: *audience})
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
