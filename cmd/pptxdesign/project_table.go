package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectTable(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch") {
		return fmt.Errorf("usage: project table inspect --project PATH --slide ID --node ID; project table patch --project PATH --slide ID --patch FILE [--apply]")
	}
	f := flag.NewFlagSet("project table "+args[0], flag.ContinueOnError)
	f.Usage = func() {
		fmt.Fprintln(f.Output(), "Inspect keyed cells, declared inline sections and sparse leading-label ambiguity. Patch preview preserves section membership; move/row requires an explicit target section. Sparse leading labels need per-row source mapping before reordering.")
		f.PrintDefaults()
	}
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID with detached local semantic table")
	node := f.String("node", "", "table node ID for inspect; patch carries node_id")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to lock")
	engine := f.String("engine", "", "engine; defaults to lock")
	patch := f.String("patch", "", "strict YAML/JSON table patch")
	apply := f.Bool("apply", false, "apply measured preview through guarded source transaction")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("table requires --slide and no positional arguments")
	}
	var bad error
	f.Visit(func(v *flag.Flag) {
		if args[0] == "inspect" && (v.Name == "patch" || v.Name == "apply") || args[0] == "patch" && v.Name == "node" {
			bad = fmt.Errorf("--%s is not accepted by table %s", v.Name, args[0])
		}
	})
	if bad != nil {
		return bad
	}
	if args[0] == "inspect" && *node == "" {
		return fmt.Errorf("table inspect requires --node")
	}
	if args[0] == "patch" && *patch == "" {
		return fmt.Errorf("table patch requires --patch")
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
		result, e = deckproject.InspectTable(p, *slide, *node, b, en)
	} else {
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeTablePatch(raw, *patch)
		if err != nil {
			return err
		}
		result, e = deckproject.PatchTable(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
