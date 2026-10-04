package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectAttachRender(args []string) error {
	f := flag.NewFlagSet("project attach-render", flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	render := f.String("render", "", "native render output directory")
	decisionsPath := f.String("decisions", "", "optional JSON map slide ID -> {status,reviewer,note}")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *render == "" {
		return fmt.Errorf("attach-render requires --render DIRECTORY")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	decisions := map[string]deckproject.VisualDecision{}
	if *decisionsPath != "" {
		raw, e := os.ReadFile(*decisionsPath)
		if e != nil {
			return e
		}
		if decisions, e = deckproject.ParseVisualDecisions(raw); e != nil {
			return e
		}
	}
	result, err := deckproject.AttachNativeRender(p, *render, decisions)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
