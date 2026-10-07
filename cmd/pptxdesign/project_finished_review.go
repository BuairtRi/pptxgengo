package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

func runFinishedSlideReview(args []string) error {
	f := flag.NewFlagSet("project slide review-reuse", flag.ContinueOnError)
	packagePath := f.String("package", "", "exact closed revision that was reviewed")
	decisions := f.String("decision", "", "explicit operator JSON decision with revision/source/artifact hashes")
	out := f.String("out", "", "new immutable revision directory")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 0 || *packagePath == "" || *decisions == "" || *out == "" {
		return fmt.Errorf("slide review-reuse requires --package DIRECTORY --decision FILE.json --out NEW_DIRECTORY and no positional arguments")
	}
	raw, e := readReconciliationInput(*decisions, 2<<20)
	if e != nil {
		return e
	}
	result, e := finishedslide.ReviewRevision(*packagePath, *out, raw)
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
