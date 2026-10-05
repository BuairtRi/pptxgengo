package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectSplit(args []string) error {
	f := flag.NewFlagSet("project split", flag.ContinueOnError)
	path := f.String("project", ".", "existing project directory or deck.yaml")
	bundle := f.String("bundle", "", "optional pinned bundle for stock content aliases")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected project split positional arguments")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	options := deckproject.SplitOptions{}
	if *bundle != "" {
		options.Bundle = deckproject.BundlePath(*bundle)
		if validLockedBundle(*bundle) {
			options.Bundle = designBundlePath(*bundle)
		}
		options.StockEditor = deckproject.StockEditableSlide
	}
	receipt, err := deckproject.Split(p, options)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(receipt)
}
