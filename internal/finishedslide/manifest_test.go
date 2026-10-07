package finishedslide

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T) (string, Manifest) {
	t.Helper()
	root := t.TempDir()
	source := []byte("id: example\ncontent_kind: synthetic_example\n")
	m := Manifest{Schema: Schema, ID: "curated/slide/example", Revision: 1, Name: "Example", Purpose: "Synthetic closure test", Owner: "test", Lifecycle: "draft", Source: "source/slide.yaml", Pins: Pins{Bundle: "v11", SourceRevision: "test", TemplateID: "cards/3", TemplateRevision: 1, TemplateSourceSHA256: sha([]byte("source")), TemplateDefinitionSHA256: sha([]byte("definition")), ToolchainLockSHA256: sha([]byte("lock")), Compiler: "deckproject.v1"}, Files: []File{{Path: "source/slide.yaml", Role: "source", Bytes: int64(len(source)), SHA256: sha(source)}}}
	if err := m.Seal(); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "source"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source/slide.yaml"), source, 0600); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "manifest.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	return root, m
}

func TestClosedRevisionDetectsDriftAndMissingDependencies(t *testing.T) {
	root, m := fixture(t)
	got, err := Read(root)
	if err != nil || got.RevisionSHA256 != m.RevisionSHA256 {
		t.Fatal(got, err)
	}
	if got.Lifecycle != "draft" || got.Approval != nil {
		t.Fatal("read granted approval")
	}
	path := filepath.Join(root, "source/slide.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	changed := append([]byte(nil), data...)
	changed[0] = 'X'
	if err = os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Read(root); err == nil || !strings.Contains(err.Error(), "hash_changed") {
		t.Fatal("modified source accepted", err)
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err = Read(root); err == nil {
		t.Fatal("missing source accepted")
	}
}

func TestUnlistedAndLinkedFilesRejected(t *testing.T) {
	for _, kind := range []string{"file", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			root, _ := fixture(t)
			extra := filepath.Join(root, "extra")
			var err error
			switch kind {
			case "file":
				err = os.WriteFile(extra, []byte("extra"), 0600)
			case "directory":
				err = os.Mkdir(extra, 0700)
			case "symlink":
				err = os.Symlink("source/slide.yaml", extra)
			}
			if err != nil {
				if kind == "symlink" {
					t.Skip("symlink permission unavailable", err)
				}
				t.Fatal(err)
			}
			if _, err = Read(root); err == nil {
				t.Fatal("unlisted/link entry accepted")
			}
		})
	}
}

func TestManifestPinsPortablePathsAndApproval(t *testing.T) {
	_, m := fixture(t)
	for _, path := range []string{"../slide.yaml", "a/../../slide.yaml", "C:/slide.yaml", `a\b`, "AUX.txt", "assets/a.", "assets/a ", "manifest.json/file"} {
		changed := m
		changed.Source = path
		changed.Files = []File{{Path: path, Role: "source", Bytes: 1, SHA256: sha([]byte("x"))}}
		if err := changed.Seal(); err == nil {
			t.Fatal("nonportable path accepted", path)
		}
	}
	for _, names := range [][]string{{"a", "A"}, {"assets", "assets/a"}, {"Assets/file", "assets"}} {
		changed := m
		changed.Files = append([]File(nil), m.Files...)
		for _, name := range names {
			changed.Files = append(changed.Files, File{Path: name, Role: "asset", Bytes: 0, SHA256: sha(nil)})
		}
		if err := changed.Seal(); err == nil {
			t.Fatal("case/file-directory collision accepted", names)
		}
	}
	changed := m
	changed.Pins.TemplateSourceSHA256 = "changed"
	if err := changed.Seal(); err == nil {
		t.Fatal("unpinned source accepted")
	}
	changed = m
	changed.Lifecycle = "approved"
	if err := changed.Seal(); err == nil {
		t.Fatal("approval invented")
	}
	changed.Approval = &Approval{By: "curator", Date: "2026-10-06", ReuseScope: "internal example"}
	if err := changed.Seal(); err == nil {
		t.Fatal("approved without preview/review artifacts")
	}
}

func TestRevisionDigestAndFreshness(t *testing.T) {
	_, m := fixture(t)
	original := m.RevisionSHA256
	m.Purpose = "Changed purpose"
	if err := m.Validate(); err == nil {
		t.Fatal("changed manifest accepted")
	}
	if err := m.Seal(); err != nil || m.RevisionSHA256 == original {
		t.Fatal("revision digest unchanged", err)
	}
	if got := m.Freshness(time.Now()); got != "no_expiry_declared" {
		t.Fatal(got)
	}
	m.ReviewedAt = "2026-10-06"
	m.ValidUntil = "2026-10-07"
	if err := m.Seal(); err != nil {
		t.Fatal(err)
	}
	if m.Freshness(time.Date(2026, 10, 7, 23, 59, 0, 0, time.UTC)) != "within_declared_review_window" || m.Freshness(time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)) != "stale" {
		t.Fatal("expiry boundary wrong")
	}
	m.ValidUntil = "2026-10-05"
	if err := m.Seal(); err == nil {
		t.Fatal("review/expiry order invalid")
	}
}

func TestCreateImmutableRevisionAndSourceIndependence(t *testing.T) {
	root, m := fixture(t)
	source, err := os.ReadFile(filepath.Join(root, m.Source))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{m.Source: source}
	out := filepath.Join(t.TempDir(), "revision with spaces")
	created, err := Create(out, m, files)
	if err != nil {
		t.Fatal(err)
	}
	if created.Lifecycle != "draft" || created.Approval != nil {
		t.Fatal("publication granted approval")
	}
	if _, err = Create(out, m, files); err == nil {
		t.Fatal("revision overwritten")
	}
	source[0] = 'X'
	got, err := Read(out)
	if err != nil || got.RevisionSHA256 != created.RevisionSHA256 {
		t.Fatal("caller mutation affected published bytes", err)
	}
	m.Revision++
	next := filepath.Join(t.TempDir(), "revision-2")
	if _, err = Create(next, m, files); err != nil {
		t.Fatal(err)
	}
	if got, err = Read(out); err != nil || got.RevisionSHA256 != created.RevisionSHA256 {
		t.Fatal("later revision changed earlier one", err)
	}
	bad := filepath.Join(t.TempDir(), "missing")
	if _, err = Create(bad, m, nil); err == nil {
		t.Fatal("missing closure accepted")
	}
	if _, err = os.Lstat(bad); !os.IsNotExist(err) {
		t.Fatal("invalid revision wrote output")
	}
}
