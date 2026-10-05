package main

import (
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryFitRejectsAmbiguousDocumentsBeforeOutput(t *testing.T) {
	const payload = `"year":2026,"slides":[{"id":"candidate","template":"cards/3","content_kind":"supplied_content","values":{"eyebrow":"Review","title":"Controls","cards":[{"key":"a","title":"Trace","body":"Trace evidence."},{"key":"b","title":"Review","body":"Review changes."},{"key":"c","title":"Retain","body":"Retain decisions."}]}}]`
	if _, err := wmdesign.DecodeBoundDocument([]byte(`{"schema":"pptxgengo.wmds-template-document.v1",` + payload + `}`)); err != nil {
		t.Fatal("baseline candidate invalid", err)
	}
	for _, raw := range []string{
		`{"schema":"pptxgengo.wmds-template-document.v1","schema":"pptxgengo.wmds-template-document.v1",` + payload + `}`,
		`{"schema":"pptxgengo.wmds-template-document.v1","Schema":"pptxgengo.wmds-template-document.v1",` + payload + `}`,
		`{"schema":"pptxgengo.wmds-template-document.v1",` + payload + `,"SLIDES":[]}`,
	} {
		root := t.TempDir()
		spec, out := filepath.Join(root, "candidate.json"), filepath.Join(root, "result")
		if err := os.WriteFile(spec, []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
		if err := runLibraryIndex("library-fit", []string{"--spec", spec, "--out", out, "--bundle", filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")}); err == nil {
			t.Fatal("ambiguous content document accepted")
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatal("invalid input created candidate output")
		}
	}
}

func TestLibraryIndexRejectsIrrelevantFlags(t *testing.T) {
	for command, flag := range map[string]string{"library-index": "--query", "library-fit": "--index", "library-inspect": "--items", "library-preview": "--roles"} {
		if err := runLibraryIndex(command, []string{flag, "1"}); err == nil {
			t.Fatalf("%s silently accepted %s", command, flag)
		}
	}
}
