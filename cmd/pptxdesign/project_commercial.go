package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/deckproject"
	"os"
)

func runProjectCommercial(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch") {
		return fmt.Errorf("usage: project commercial inspect --project PATH --slide ID --node ID; project commercial patch --project PATH --slide ID --patch FILE [--apply]")
	}
	f := flag.NewFlagSet("project commercial "+args[0], flag.ContinueOnError)
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID")
	node := f.String("node", "", "commercial node ID for inspect; patch carries node_id")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to lock")
	engine := f.String("engine", "", "engine; defaults to lock")
	patch := f.String("patch", "", "strict YAML/JSON commercial patch")
	apply := f.Bool("apply", false, "apply measured candidate through guarded source transaction")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("commercial requires --slide and no positional arguments")
	}
	var bad error
	f.Visit(func(v *flag.Flag) {
		if args[0] == "inspect" && (v.Name == "patch" || v.Name == "apply") || args[0] == "patch" && v.Name == "node" {
			bad = fmt.Errorf("--%s is not accepted by commercial %s", v.Name, args[0])
		}
	})
	if bad != nil {
		return bad
	}
	if args[0] == "inspect" && *node == "" {
		return fmt.Errorf("commercial inspect requires --node")
	}
	if args[0] == "patch" && *patch == "" {
		return fmt.Errorf("commercial patch requires --patch")
	}
	p, e := deckproject.Load(*project)
	if e != nil {
		return e
	}
	b, en, e := projectRuntime(p, *bundle, *engine)
	if e != nil {
		return e
	}
	var out any
	if args[0] == "inspect" {
		out, e = deckproject.InspectCommercial(p, *slide, *node, b, en)
	} else {
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeCommercialPatch(raw, *patch)
		if err != nil {
			return err
		}
		out, e = deckproject.PatchCommercial(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
