package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPathsIncludesLatestBundleOnly(t *testing.T) {
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
	for _, revision := range []string{"v12"} {
		want := filepath.Join(paths["root"], "library", "wm-design-system", revision)
		if got := paths["design_system_"+revision]; got != want {
			t.Errorf("%s bundle path = %q; want %q", revision, got, want)
		}
	}
	for _, revision := range []string{"v1", "v2", "v3", "v4", "v5", "v6"} {
		if _, ok := paths["design_system_"+revision]; ok {
			t.Errorf("removed bundle advertised: %s", revision)
		}
	}
	if paths["design_index"] != filepath.Join(paths["design_system_v12"], "library.sqlite") {
		t.Fatalf("index outside current bundle: %v", paths)
	}
}
