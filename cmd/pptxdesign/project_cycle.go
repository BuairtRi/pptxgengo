package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectCycle(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch") {
		return fmt.Errorf("usage: project cycle inspect --project PATH --slide ID --node ID; project cycle patch --project PATH --slide ID --patch FILE [--apply]")
	}
	f := flag.NewFlagSet("project cycle "+args[0], flag.ContinueOnError)
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID with detached local semantic cycle")
	node := f.String("node", "", "cycle node ID for inspect; patch carries node_id")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to lock")
	engine := f.String("engine", "", "engine; defaults to lock")
	patch := f.String("patch", "", "strict YAML/JSON cycle patch")
	apply := f.Bool("apply", false, "apply measured preview through guarded source transaction")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("cycle requires --slide and no positional arguments")
	}
	var bad error
	f.Visit(func(v *flag.Flag) {
		if args[0] == "inspect" && (v.Name == "patch" || v.Name == "apply") || args[0] == "patch" && v.Name == "node" {
			bad = fmt.Errorf("--%s is not accepted by cycle %s", v.Name, args[0])
		}
	})
	if bad != nil {
		return bad
	}
	if args[0] == "inspect" && *node == "" {
		return fmt.Errorf("cycle inspect requires --node")
	}
	if args[0] == "patch" && *patch == "" {
		return fmt.Errorf("cycle patch requires --patch")
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
		result, e = deckproject.InspectCycle(p, *slide, *node, b, en)
	} else {
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeCyclePatch(raw, *patch)
		if err != nil {
			return err
		}
		result, e = deckproject.PatchCycle(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
