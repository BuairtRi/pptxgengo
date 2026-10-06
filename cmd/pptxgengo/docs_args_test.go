package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocsArgsUsesPackagedSiteByDefault(t *testing.T) {
	root := t.TempDir()
	input := []string{"-addr", "localhost:0"}
	got := docsArgs(root, input)
	want := []string{"-addr", "localhost:0", "--dir", filepath.Join(root, "wmds-docs", "site")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("docs defaults: got %v; want %v", got, want)
	}
	if !reflect.DeepEqual(input, []string{"-addr", "localhost:0"}) {
		t.Fatalf("docsArgs mutated input: %v", input)
	}
}

func TestDocsArgsRespectsExplicitDirectoryForms(t *testing.T) {
	root := t.TempDir()
	for _, input := range [][]string{
		{"--dir", "/custom/site"},
		{"--dir=/custom/site"},
		{"-dir", "/custom/site"},
		{"-dir=/custom/site"},
	} {
		if got := docsArgs(root, input); !reflect.DeepEqual(got, input) {
			t.Errorf("explicit directory %v changed to %v", input, got)
		}
	}
}
