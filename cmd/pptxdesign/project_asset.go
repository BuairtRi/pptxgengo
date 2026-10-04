package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectAsset(args []string) error {
	if len(args) == 0 || args[0] != "add" {
		return fmt.Errorf("usage: project asset add --project PATH --id ID --file IMAGE --description TEXT [--focus 0.5,0.5]")
	}
	f := flag.NewFlagSet("project asset add", flag.ContinueOnError)
	project := f.String("project", ".", "project path")
	id := f.String("id", "", "stable asset ID")
	file := f.String("file", "", "approved or client-supplied image original")
	description := f.String("description", "", "human-readable subject/source description")
	focus := f.String("focus", "", "optional normalized focus x,y; advisory composition metadata")
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 || *file == "" {
		return fmt.Errorf("asset add requires --file")
	}
	p, err := deckproject.Load(*project)
	if err != nil {
		return err
	}
	info, err := os.Stat(*file)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > 64<<20 {
		return fmt.Errorf("asset source must be a regular file <=64MiB")
	}
	data, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	o := deckproject.AssetRegistration{ID: *id, Data: data, Description: *description}
	if *focus != "" {
		parts := strings.Split(*focus, ",")
		if len(parts) != 2 {
			return fmt.Errorf("focus requires x,y")
		}
		x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return err
		}
		y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return err
		}
		o.Focus = &deckproject.AssetFocus{X: x, Y: y}
	}
	receipt, err := deckproject.RegisterAsset(p, o)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(receipt)
}
