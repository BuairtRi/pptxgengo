package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectGantt(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch") {
		return fmt.Errorf("usage: project gantt inspect --project PATH --slide ID --node ID; project gantt patch --project PATH --slide ID --patch FILE [--apply]")
	}
	f := flag.NewFlagSet("project gantt "+args[0], flag.ContinueOnError)
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID; requires detached local semantic Gantt component")
	node := f.String("node", "", "Gantt node ID for inspect; patch carries node_id")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to lock")
	engine := f.String("engine", "", "engine; defaults to lock")
	patch := f.String("patch", "", "strict YAML/JSON Gantt composition patch")
	apply := f.Bool("apply", false, "apply measured preview through guarded source transaction")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("Gantt requires --slide and no positional arguments")
	}
	var bad error
	f.Visit(func(v *flag.Flag) {
		if args[0] == "inspect" && (v.Name == "patch" || v.Name == "apply") || args[0] == "patch" && v.Name == "node" {
			bad = fmt.Errorf("--%s is not accepted by Gantt %s", v.Name, args[0])
		}
	})
	if bad != nil {
		return bad
	}
	if args[0] == "inspect" && *node == "" {
		return fmt.Errorf("Gantt inspect requires --node")
	}
	if args[0] == "patch" && *patch == "" {
		return fmt.Errorf("Gantt patch requires --patch")
	}
	p, e := deckproject.Load(*project)
	if e != nil {
		return e
	}
	b, en, e := projectRuntime(p, *bundle, *engine)
	if e != nil {
		return e
	}
	var result any
	if args[0] == "inspect" {
		result, e = deckproject.InspectGantt(p, *slide, *node, b, en)
	} else {
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeGanttPatch(raw, *patch)
		if err != nil {
			return err
		}
		result, e = deckproject.PatchGantt(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
