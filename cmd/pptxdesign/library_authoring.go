package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runLibraryAuthoring(args []string) error {
	f := flag.NewFlagSet("library-authoring", flag.ContinueOnError)
	bundle := f.String("bundle", "v10", "pinned library bundle")
	key := f.String("template", "", "exact template key; omit for coverage report")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if validLockedBundle(*bundle) {
		*bundle = designBundlePath(*bundle)
	}
	catalog, err := wmdesign.LibraryCatalog(*bundle, "")
	if err != nil {
		return err
	}
	var result any
	if *key == "" {
		result, err = wmdesign.AuthoringCoverage(catalog)
	} else {
		for _, def := range catalog {
			if def.Key == *key {
				var metadata wmdesign.LibraryAuthoring
				metadata, err = wmdesign.LibraryAuthoringMetadata(def)
				if err != nil {
					break
				}
				var typography *wmdesign.Typography
				typography, err = wmdesign.NewTypography(filepath.Join(*bundle, "fonts"))
				if err != nil {
					break
				}
				for i := range metadata.Slots {
					metadata.Slots[i].Capacity, err = wmdesign.EstimateLibrarySlotCapacity(typography, metadata.Slots[i].Capacity)
					if err != nil {
						break
					}
				}
				result = metadata
				break
			}
		}
		if result == nil && err == nil {
			return fmt.Errorf("unknown template %s", *key)
		}
	}
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
