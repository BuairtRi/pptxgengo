package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestCatalogAndIndexV5AndCurrentDefaults(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
	for _, route := range []string{"catalog", "catalog-default", "index", "index-default"} {
		t.Run(route, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "output-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			previousArgs, previousStdout := os.Args, os.Stdout
			os.Stdout = output
			defer func() { os.Args, os.Stdout = previousArgs, previousStdout }()
			catalog := route == "catalog" || route == "catalog-default"
			wantCount, wantRevision := 587, wmdesign.LibraryRevisionV5
			if route == "catalog-default" || route == "index-default" {
				wantCount, wantRevision = 649, wmdesign.LibraryRevisionV11
			}
			if catalog {
				os.Args = []string{"pptxdesign", "library-catalog", "--include-deprecated"}
				if route == "catalog" {
					os.Args = append(os.Args, "--bundle", "v5", "--engine", wmdesign.CandidateEngine)
				}
				err = run()
			} else {
				args := []string{"--out", filepath.Join(t.TempDir(), "new.sqlite")}
				if route == "index" {
					args = append(args, "--bundle", "v5")
				}
				err = runLibraryIndex("library-index", args)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			if catalog {
				var rows []wmdesign.LibraryTemplate
				if err := json.NewDecoder(output).Decode(&rows); err != nil {
					t.Fatal(err)
				}
				if len(rows) != wantCount {
					t.Fatalf("catalog returned %d definitions; want %d", len(rows), wantCount)
				}
				for _, row := range rows {
					if row.SourceRevision != wantRevision {
						t.Fatalf("catalog mixed source revisions: %s: %s", row.Key, row.SourceRevision)
					}
				}
			} else {
				var report wmdesign.LibraryIndexReport
				if err := json.NewDecoder(output).Decode(&report); err != nil {
					t.Fatal(err)
				}
				if report.SourceRevision != wantRevision || report.Counts["template"] != wantCount {
					t.Fatalf("index returned revision/count %s/%d", report.SourceRevision, report.Counts["template"])
				}
			}
		})
	}
}

func TestLatestDiscoveryResourcesAreDefault(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
	for _, command := range []string{"library-find", "library-inspect", "library-preview"} {
		t.Run(command, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "output-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			previous := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = previous }()
			args := []string{"--id", "quote/light"}
			if command == "library-find" {
				args = []string{"--query", "maturity", "--limit", "1"}
			}
			if err := runLibraryIndex(command, args); err != nil {
				t.Fatal(err)
			}
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			var result json.RawMessage
			if err := json.NewDecoder(output).Decode(&result); err != nil {
				t.Fatal(err)
			}
		})
	}
}
