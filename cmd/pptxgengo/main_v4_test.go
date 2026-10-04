package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPathsIncludesExplicitV4Bundle(t *testing.T) {
	output, err := os.CreateTemp(t.TempDir(), "paths-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	previousArgs, previousStdout := os.Args, os.Stdout
	os.Args, os.Stdout = []string{"pptxgengo", "paths"}, output
	defer func() { os.Args, os.Stdout = previousArgs, previousStdout }()
	if err := run(); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	var paths map[string]string
	if err := json.NewDecoder(output).Decode(&paths); err != nil {
		t.Fatal(err)
	}
	for _, revision := range []string{"v2", "v3", "v4", "v5"} {
		want := filepath.Join(paths["root"], "library", "wm-design-system", revision)
		if got := paths["design_system_"+revision]; got != want {
			t.Errorf("%s bundle path = %q; want %q", revision, got, want)
		}
	}
}
