package deckproject

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// This opt-in export uses the same receipt-backed source specimens as the
// semantic tests. Desktop qualification builds a fresh receipt after relocation.
func TestExportNativeFamilyQualificationFixtures(t *testing.T) {
	root := os.Getenv("PPTXGENGO_NATIVE_FAMILY_FIXTURES")
	if root == "" {
		t.Skip("desktop fixture export requires explicit destination")
	}
	cases := []struct {
		name    string
		fixture func(*testing.T) (*Project, *TextBaseline, map[string]string)
	}{
		{"team-membership", func(t *testing.T) (*Project, *TextBaseline, map[string]string) {
			p, b := teamSemanticFixture(t)
			return p, b, map[string]string{"slide": "team-slide", "source_role": "delivery.roles.architect", "source_pod": "delivery", "target_pod": "operations", "instruction": "Move the complete Architect role (surface and text) from Delivery into Operations; membership remains a reviewed decision."}
		}},
		{"assessment-score", func(t *testing.T) (*Project, *TextBaseline, map[string]string) {
			p, b, node := assessmentSemanticFixture(t)
			return p, b, map[string]string{"slide": "assessment-slide", "node": node, "instruction": "Edit one ordinal numeric score in the native table; retain table headings, row/column identity and score formatting."}
		}},
		{"gantt-task-and-gate", func(t *testing.T) (*Project, *TextBaseline, map[string]string) {
			p, b, node, bar := ganttSemanticFixture(t)
			return p, b, map[string]string{"slide": "plan-slide", "node": node, "task_native_object": bar, "gate_native_object": node + ".gates.approval.line", "instruction": "Move/resize Deliver task bar and move Approve gate; retain axis labels and inspect period snapping before adopting dates."}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, b, metadata := c.fixture(t)
			target := filepath.Join(root, c.name)
			if e := os.Mkdir(target, 0700); e != nil {
				t.Fatal(e)
			}
			project := filepath.Join(target, "project")
			if e := filepath.WalkDir(p.Root, func(path string, entry fs.DirEntry, e error) error {
				if e != nil {
					return e
				}
				rel, e := filepath.Rel(p.Root, path)
				if e != nil {
					return e
				}
				dest := filepath.Join(project, rel)
				if entry.IsDir() {
					return os.MkdirAll(dest, 0700)
				}
				info, e := entry.Info()
				if e != nil {
					return e
				}
				if !info.Mode().IsRegular() {
					return &fs.PathError{Op: "export", Path: path, Err: fs.ErrInvalid}
				}
				raw, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				mode := fs.FileMode(0600)
				if info.Mode().Perm()&0200 == 0 {
					mode = 0400
				}
				return os.WriteFile(dest, raw, mode)
			}); e != nil {
				t.Fatal(e)
			}
			baseline := filepath.Join(target, "baseline")
			if e := os.Mkdir(baseline, 0700); e != nil {
				t.Fatal(e)
			}
			for name, data := range b.files {
				if e := os.WriteFile(filepath.Join(baseline, name), data, 0600); e != nil {
					t.Fatal(e)
				}
			}
			source, e := filepath.Rel(p.Root, p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			metadata["project_source"] = filepath.Join("project", source)
			metadata["bundle"] = bundle(t)
			metadata["baseline_policy"] = "Relocated source requires fresh migrate/pin/build; retained baseline is exact test-fixture evidence, not authorization for a new source path."
			if e = os.WriteFile(filepath.Join(target, "qualification.json"), canonical(metadata), 0600); e != nil {
				t.Fatal(e)
			}
		})
	}
}
