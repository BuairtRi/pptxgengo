package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectSlideDraftReview(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: project slide draft-review <set|clear|show> --project PATH --id ID")
	}
	action := args[0]
	if action != "set" && action != "clear" && action != "show" {
		return fmt.Errorf("unknown draft-review operation %q", action)
	}
	f := flag.NewFlagSet("project slide draft-review "+action, flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	id := f.String("id", "", "stable slide ID")
	fields := map[string]*string{}
	for _, key := range []string{"status", "status-text", "status-color", "owner", "due", "updated", "notes", "placement"} {
		fields[key] = f.String(key, "", "set: draft review "+key)
	}
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected draft-review positional arguments")
	}
	values := map[string]string{}
	f.Visit(func(field *flag.Flag) {
		if value, exists := fields[field.Name]; exists {
			values[strings.ReplaceAll(field.Name, "-", "_")] = *value
		}
	})
	if action != "set" && len(values) > 0 {
		return fmt.Errorf("draft-review fields require set")
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	receipt, err := deckproject.OperateDraftReview(p, deckproject.DraftReviewOperation{Action: action, ID: *id, Fields: values})
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(receipt)
}
