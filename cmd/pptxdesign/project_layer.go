package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/deckproject"
	"io"
	"os"
)

func runProjectLayer(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch") {
		return fmt.Errorf("usage: project layer inspect --project PATH --slide ID [--selection FILE]; project layer patch --project PATH --slide ID --patch FILE [--apply]")
	}
	f := flag.NewFlagSet("project layer "+args[0], flag.ContinueOnError)
	path := f.String("project", ".", "project root or deck.yaml")
	slide := f.String("slide", "", "stable slide ID")
	selection := f.String("selection", "", "JSON array of explicitly mapped layer selections")
	patch := f.String("patch", "", "strict layer patch YAML/JSON")
	apply := f.Bool("apply", false, "apply measured guarded preview")
	bundle := f.String("bundle", "", "bundle defaults to lock")
	engine := f.String("engine", "", "engine defaults to lock")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("layer requires --slide and no positional arguments")
	}
	var bad error
	f.Visit(func(v *flag.Flag) {
		if args[0] == "inspect" && (v.Name == "patch" || v.Name == "apply") || args[0] == "patch" && v.Name == "selection" {
			bad = fmt.Errorf("--%s not accepted by layer %s", v.Name, args[0])
		}
	})
	if bad != nil {
		return bad
	}
	if args[0] == "patch" && *patch == "" {
		return fmt.Errorf("layer patch requires --patch")
	}
	p, e := deckproject.Load(*path)
	if e != nil {
		return e
	}
	b, en, e := projectRuntime(p, *bundle, *engine)
	if e != nil {
		return e
	}
	var result any
	if args[0] == "inspect" {
		var selections []deckproject.LayerSelection
		if *selection != "" {
			raw, e := readReconciliationInput(*selection, 1<<20)
			if e != nil {
				return e
			}
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.DisallowUnknownFields()
			if e = decoder.Decode(&selections); e != nil {
				return e
			}
			var extra any
			if e = decoder.Decode(&extra); e != io.EOF {
				return fmt.Errorf("selection requires one JSON array")
			}
		}
		result, e = deckproject.InspectLayers(p, *slide, selections, b, en)
	} else {
		if *patch == "" {
			return fmt.Errorf("layer patch requires --patch")
		}
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeLayerPatch(raw, *patch)
		if err != nil {
			return err
		}
		result, e = deckproject.PatchLayers(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
