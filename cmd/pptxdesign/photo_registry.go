package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
)

func runPhotoRegistration(args []string) error {
	f := flag.NewFlagSet("photo-register", flag.ContinueOnError)
	root := f.String("branding-root", "", "local branding root containing West Monroe Photos")
	out := f.String("out", "", "new versioned photo snapshot JSON")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *out == "" {
		return fmt.Errorf("photo-register requires --branding-root ROOT --out NEWFILE")
	}
	if *root == "" {
		*root = os.Getenv("WMDS_BRANDING_ROOT")
		if *root == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			*root = filepath.Join(home, "Documents/branding")
		}
	}
	report, err := wmdesign.RegisterPhotoLibrary(*root, *out)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(report)
}
