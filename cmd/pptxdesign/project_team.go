package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectTeam(args []string) error {
	if len(args) > 0 && args[0] == "reconcile" {
		return runProjectTeamReconcile(args[1:])
	}
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch") {
		return fmt.Errorf("usage: project team <inspect|patch> --project PATH --slide ID; patch --patch FILE [--apply]")
	}
	f := flag.NewFlagSet("project team "+args[0], flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID with a project-owned local template")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to project lock")
	engine := f.String("engine", "", "engine; defaults to project lock")
	patch := f.String("patch", "", "strict team-patch.v1 YAML/JSON operations and inspected source hash")
	apply := f.Bool("apply", false, "apply the measured preview through a guarded source transaction")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("team requires --slide ID and no positional arguments")
	}
	if args[0] == "inspect" {
		var invalid bool
		f.Visit(func(v *flag.Flag) {
			if v.Name == "patch" || v.Name == "apply" {
				invalid = true
			}
		})
		if invalid {
			return fmt.Errorf("team inspect does not accept --patch or --apply")
		}
	} else if *patch == "" {
		return fmt.Errorf("team patch requires --patch FILE")
	}
	p, e := deckproject.Load(*path)
	if e != nil {
		return e
	}
	b, en, e := projectRuntime(p, *bundle, *engine)
	if e != nil {
		return e
	}
	var out any
	if args[0] == "inspect" {
		out, e = deckproject.InspectTeam(p, *slide, b, en)
	} else {
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeTeamPatch(raw, *patch)
		if err != nil {
			return err
		}
		out, e = deckproject.PatchTeam(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
