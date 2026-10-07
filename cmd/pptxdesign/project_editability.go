package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectEditability(args []string) error {
	f := flag.NewFlagSet("project editability", flag.ContinueOnError)
	project := f.String("project", ".", "maintained deck directory or deck.yaml")
	build := f.String("build", "", "baseline build; defaults to the state-pinned current build")
	receipt := f.String("receipt-sha256", "", "trusted receipt SHA-256 required for historical builds")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected project editability positional arguments")
	}
	p, e := deckproject.Load(*project)
	if e != nil {
		return e
	}
	b, e := deckproject.ReadTextBaseline(p, *build, *receipt)
	if e != nil {
		return e
	}
	report, e := deckproject.NativeEditability(b)
	if e != nil {
		return e
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
