package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectSwap(args []string) error {
	f := flag.NewFlagSet("project swap", flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	id := f.String("slide", "", "stable shared slide ID")
	key := f.String("template", "", "target stock template key")
	apply := f.Bool("apply", false, "apply complete proposal; otherwise only print proposal")
	allow := f.Bool("allow-unmapped", false, "explicitly permit removing reported visible fields, retaining source history")
	bundle := f.String("bundle", "", "bundle path/revision, defaults to lock")
	engine := f.String("engine", "", "engine, defaults to lock")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *id == "" || *key == "" {
		return fmt.Errorf("swap requires --slide ID --template KEY")
	}
	if *allow && !*apply {
		return fmt.Errorf("--allow-unmapped requires --apply")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	b, e, err := projectRuntime(p, *bundle, *engine)
	if err != nil {
		return err
	}
	proposal, err := deckproject.ProposeSwap(p, *id, *key, b)
	if err != nil {
		return err
	}
	var result any = proposal
	if *apply {
		receipt, err := deckproject.ApplySwap(p, proposal, *allow, b, e)
		if err != nil {
			return err
		}
		result = struct {
			Proposal deckproject.SwapProposal `json:"proposal"`
			Receipt  deckproject.EditReceipt  `json:"receipt"`
		}{proposal, receipt}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
