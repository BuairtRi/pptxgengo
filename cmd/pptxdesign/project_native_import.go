package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectNativeImport(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch") {
		return fmt.Errorf("usage: project native-import <inspect|patch> --in PPTX; patch --project PATH --slide ID --map FILE [--apply]")
	}
	f := flag.NewFlagSet("project native-import "+args[0], flag.ContinueOnError)
	in := f.String("in", "", "native PowerPoint package (32 MiB maximum)")
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "detached/local slide ID")
	mapping := f.String("map", "", "strict native import map with both source hashes and explicit formatting policy")
	bundle := f.String("bundle", "", "bundle; defaults to project lock")
	engine := f.String("engine", "", "engine; defaults to project lock")
	apply := f.Bool("apply", false, "commit measured source proposal and retained source package atomically")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	var invalid error
	f.Visit(func(v *flag.Flag) {
		if args[0] == "inspect" && v.Name != "in" {
			invalid = fmt.Errorf("--%s is not accepted by native-import inspect", v.Name)
		}
	})
	if invalid != nil {
		return invalid
	}
	if f.NArg() != 0 || *in == "" {
		return fmt.Errorf("native-import requires --in PPTX and no positional arguments")
	}
	data, e := readReconciliationInput(*in, deckproject.NativeImportMaxBytes)
	if e != nil {
		return e
	}
	var result any
	if args[0] == "inspect" {
		result, e = deckproject.InspectNativeImport(data)
	} else {
		if *slide == "" || *mapping == "" {
			return fmt.Errorf("native-import patch requires --slide and --map")
		}
		raw, err := readReconciliationInput(*mapping, 1<<20)
		if err != nil {
			return err
		}
		m, err := deckproject.DecodeNativeImportMap(raw, *mapping)
		if err != nil {
			return err
		}
		p, err := deckproject.Load(*project)
		if err != nil {
			return err
		}
		b, en, err := projectRuntime(p, *bundle, *engine)
		if err != nil {
			return err
		}
		result, e = deckproject.ImportNative(p, *slide, data, m, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
