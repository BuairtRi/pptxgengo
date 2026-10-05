package wmdesign

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLibraryIndexSharedSourceCandidates(t *testing.T) {
	for _, candidate := range []struct {
		revision, directory string
		templates           int
	}{
		{"v6", "intake-20261004-602-frozen", 602},
		{"v7", "intake-20261005-616-frozen", 616},
	} {
		t.Run(candidate.revision, func(t *testing.T) {
			bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", candidate.revision, candidate.directory, "bundle")
			path := filepath.Join(t.TempDir(), "library.sqlite")
			report, err := BuildLibraryIndex(path, LibraryIndexOptions{Bundle: bundle})
			if err != nil {
				t.Fatal(err)
			}
			index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer index.Close()
			if index.Report.Counts["template"] != report.Counts["template"] {
				t.Fatal("source sharing changed template count")
			}
			if report.Counts["template"] != candidate.templates || report.Counts["asset"] != len(PrimitiveAssetCatalog()) {
				t.Fatalf("candidate inventory counts changed: %+v", report.Counts)
			}
			if candidate.revision == "v7" {
				result, err := index.FindSummary(LibraryIndexFindOptions{Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: "workshop agenda facilitation session", Limit: 10}})
				if err != nil || len(result.Matches) == 0 {
					t.Fatalf("workshop discovery failed: %+v %v", result, err)
				}
				workshop := false
				for _, hit := range result.Matches {
					workshop = workshop || strings.Contains(hit.Key, "workshop")
				}
				if !workshop {
					t.Fatalf("workshop query returned no workshop template: %+v", result.Matches)
				}
			}
		})
	}
}

func TestLibraryIndexSharedSourceTargets(t *testing.T) {
	const relative = "components/v0/components.json"
	original := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	data, err := os.ReadFile(filepath.Join(original, "source", filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	sha := indexDigest(data)
	pin := LibraryIndexPin{Scope: "source", Path: relative, SHA256: sha}
	for _, scenario := range []string{"relocated-chain", "external", "unknown-manifest", "wrong-inventory", "wrong-relative", "bad-bytes", "unknown-pin", "wrong-pin-sha"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "candidate", "source")
			targetBundle := filepath.Join(root, "older", "bundle")
			target := filepath.Join(targetBundle, "source", filepath.FromSlash(relative))
			write := func(path string, data []byte) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"bundle.json", "inventory.json"} {
				bytes, err := os.ReadFile(filepath.Join(original, name))
				if err != nil {
					t.Fatal(err)
				}
				write(filepath.Join(targetBundle, name), bytes)
			}
			write(target, data)
			switch scenario {
			case "external":
				if err := os.Remove(filepath.Join(targetBundle, "bundle.json")); err != nil {
					t.Fatal(err)
				}
			case "unknown-manifest":
				write(filepath.Join(targetBundle, "bundle.json"), []byte(`{}`))
			case "wrong-inventory":
				write(filepath.Join(targetBundle, "inventory.json"), []byte(`{}`))
			case "wrong-relative":
				target = filepath.Join(targetBundle, "source", "components", "v0", "other.json")
				write(target, data)
			case "bad-bytes":
				write(target, []byte(`{"changed":true}`))
			}
			link := filepath.Join(source, filepath.FromSlash(relative))
			if err := os.MkdirAll(filepath.Dir(link), 0700); err != nil {
				t.Fatal(err)
			}
			if scenario == "relocated-chain" {
				middle := filepath.Join(root, "middle", "source", filepath.FromSlash(relative))
				if err := os.MkdirAll(filepath.Dir(middle), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, middle); err != nil {
					t.Fatal(err)
				}
				target = middle
			}
			if err := os.Symlink(target, link); err != nil {
				t.Fatal(err)
			}
			resolver := newIndexSourcePinResolver(&Source{Root: source, Files: []SourceFile{{Path: relative, SHA256: sha}}})
			input := pin
			if scenario == "unknown-pin" {
				input.Path = "other.json"
			}
			if scenario == "wrong-pin-sha" {
				input.SHA256 = "unknown"
			}
			path, err := resolver.resolve(input)
			if scenario == "relocated-chain" {
				if err != nil {
					t.Fatal(err)
				}
				expected := filepath.Join(targetBundle, "source", filepath.FromSlash(relative))
				expected, err = filepath.EvalSymlinks(expected)
				if err != nil {
					t.Fatal(err)
				}
				if path != expected {
					t.Fatalf("resolved %s, want %s", path, expected)
				}
			} else if err == nil {
				t.Fatalf("accepted %s source target", scenario)
			}
			// The source exception must never broaden the generic guard.
			if _, err := indexRelative(source, relative); err == nil || !strings.Contains(err.Error(), "index.symlink_resource") {
				t.Fatal("generic symlink rejection changed", err)
			}
		})
	}
}
