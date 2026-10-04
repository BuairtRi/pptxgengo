package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestV5CatalogAndIndexShorthand(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
	for _, route := range []string{"catalog", "index"} {
		t.Run(route, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "output-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			previousArgs, previousStdout := os.Args, os.Stdout
			os.Stdout = output
			defer func() { os.Args, os.Stdout = previousArgs, previousStdout }()
			if route == "catalog" {
				os.Args = []string{"pptxdesign", "library-catalog", "--bundle", "v5", "--engine", wmdesign.CandidateEngine, "--include-deprecated"}
				err = run()
			} else {
				err = runLibraryIndex("library-index", []string{"--bundle", "v5", "--out", filepath.Join(t.TempDir(), "new.sqlite")})
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			if route == "catalog" {
				var rows []wmdesign.LibraryTemplate
				if err := json.NewDecoder(output).Decode(&rows); err != nil {
					t.Fatal(err)
				}
				if len(rows) != 587 {
					t.Fatalf("v5 catalog returned %d definitions; want 587", len(rows))
				}
				for _, row := range rows {
					if row.SourceRevision != wmdesign.LibraryRevisionV5 {
						t.Fatalf("v5 catalog mixed source revisions: %s: %s", row.Key, row.SourceRevision)
					}
				}
			} else {
				var report wmdesign.LibraryIndexReport
				if err := json.NewDecoder(output).Decode(&report); err != nil {
					t.Fatal(err)
				}
				if report.SourceRevision != wmdesign.LibraryRevisionV5 || report.Counts["template"] != 587 {
					t.Fatalf("v5 index returned revision/count %s/%d", report.SourceRevision, report.Counts["template"])
				}
			}
		})
	}
}
