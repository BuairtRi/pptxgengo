package wmdesign

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

func finishedIndexRevision(t *testing.T, root string, revision int) finishedslide.Manifest {
	t.Helper()
	_, report := indexFixture(t)
	catalog, err := LibraryCatalog(report.Options.Bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	var def LibraryTemplate
	for _, d := range catalog {
		if d.Key == "cards/3" {
			def = d
			break
		}
	}
	tree, err := libraryObject(def.RawSlide)
	if err != nil {
		t.Fatal(err)
	}
	source := indexJSON(map[string]any{"id": "fixture-message", "content_kind": "supplied_content", "template": map[string]any{"scope": "shared", "id": def.Key}, "values": map[string]any{"title": "Authored test message", "eyebrow": "Closed revision fixture", "cards": []any{map[string]any{"key": "one", "title": "One", "body": "Content for fixture one."}, map[string]any{"key": "two", "title": "Two", "body": "Content for fixture two."}, map[string]any{"key": "three", "title": "Three", "body": "Content for fixture three."}}}})
	preview := []byte("Fixture-only preview bytes, not native acceptance.")
	m := finishedslide.Manifest{Schema: finishedslide.Schema, ID: "curated/slide/authored-message", Revision: revision, Name: "Authored reusable message", Purpose: "Summarize a reviewed delivery plan", Owner: "Test fixture", Lifecycle: "draft", Keywords: []string{"release-capsule"}, Source: "slide.yaml", ReviewedAt: "2026-01-01", ValidUntil: "2026-01-02", Pins: finishedslide.Pins{Bundle: strings.Repeat("a", 64), SourceRevision: def.SourceRevision, TemplateID: def.Key, TemplateRevision: def.Revision, TemplateSourceSHA256: def.SourceSHA256, TemplateDefinitionSHA256: indexDigest(indexJSON(tree)), ToolchainLockSHA256: strings.Repeat("b", 64), Compiler: "test-only"}, Files: []finishedslide.File{{Path: "slide.yaml", Role: "source"}, {Path: "preview.png", Role: "preview"}}}
	out := filepath.Join(root, "authored-message", string(rune('0'+revision)))
	m, err = finishedslide.Create(out, m, map[string][]byte{"slide.yaml": source, "preview.png": preview})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestFinishedSlideUnifiedDiscoveryAndLibraryDrift(t *testing.T) {
	root := t.TempDir()
	finishedIndexRevision(t, root, 1)
	newest := finishedIndexRevision(t, root, 2)
	_, report := indexFixture(t)
	options := report.Options
	options.SlideLibrary = root
	path := filepath.Join(t.TempDir(), "with-slides.sqlite")
	built, err := BuildLibraryIndex(path, options)
	if err != nil {
		t.Fatal(err)
	}
	if built.Counts[finishedslide.Kind] != 1 || built.FinishedSlideTreeSHA256 == "" {
		t.Fatal(built)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := index.Find(LibraryIndexFindOptions{Retrieval: "keyword", Kinds: []string{finishedslide.Kind}, Shape: LibrarySearchOptions{Query: "release capsule", Limit: 10}})
	if err != nil || len(result.Matches) != 1 {
		t.Fatal(result, err)
	}
	hit := result.Matches[0]
	if hit.Entity.ID != newest.ID || hit.Entity.Revision != 2 || hit.Entity.Namespace != "curated" || !strings.Contains(hit.Entity.ReuseStatus, "stale") {
		t.Fatal(hit)
	}
	preview, err := index.Preview(newest.ID)
	canonicalRoot, canonicalErr := filepath.EvalSymlinks(root)
	if canonicalErr != nil {
		t.Fatal(canonicalErr)
	}
	if err != nil || len(preview.Paths) != 1 || !strings.HasPrefix(preview.Paths[0], canonicalRoot) {
		t.Fatal(preview, err)
	}
	card, err := index.SelectionCard(newest.ID)
	if err != nil || card.Kind != finishedslide.Kind || card.Revision != 2 || !strings.Contains(card.ReuseStatus, "destination_review_required") {
		t.Fatal(card, err)
	}
	summary, err := index.FindSummary(LibraryIndexFindOptions{Kinds: []string{finishedslide.Kind}, Shape: LibrarySearchOptions{Limit: 10}})
	if err != nil || len(summary.Matches) != 1 || summary.Matches[0].Kind != finishedslide.Kind {
		t.Fatal(summary, err)
	}
	entity, err := index.Inspect(newest.ID)
	if err != nil || len(entity.Dependencies) != 1 || entity.Template != nil {
		t.Fatal(entity, err)
	}
	index.Close()
	finishedIndexRevision(t, root, 3)
	if stale, err := OpenLibraryIndex(path, LibraryIndexOptions{}); err == nil {
		stale.Close()
		t.Fatal("new revision silently bypassed reindex")
	}
}

func TestFinishedSlidePreviewTamperingAndRelocation(t *testing.T) {
	root := t.TempDir()
	m := finishedIndexRevision(t, root, 1)
	_, report := indexFixture(t)
	options := report.Options
	options.SlideLibrary = root
	path := filepath.Join(t.TempDir(), "slides.sqlite")
	if _, err := BuildLibraryIndex(path, options); err != nil {
		t.Fatal(err)
	}
	relocated := filepath.Join(t.TempDir(), "relocated")
	if err := os.Rename(root, relocated); err != nil {
		t.Fatal(err)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{SlideLibrary: relocated})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := index.Preview(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(preview.Paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(preview.Paths[0], append(raw, '!'), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := index.Preview(m.ID); err == nil {
		t.Fatal("preview drift accepted")
	}
	index.Close()
	if index, err := OpenLibraryIndex(path, LibraryIndexOptions{SlideLibrary: relocated}); err == nil {
		index.Close()
		t.Fatal("tampered revision opened")
	}
}

func TestFinishedSlideLibraryRejectsLinksLooseFilesAndDuplicateRevision(t *testing.T) {
	root := t.TempDir()
	m := finishedIndexRevision(t, root, 1)
	if err := os.WriteFile(filepath.Join(root, "loose.txt"), []byte("unclosed"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := readFinishedLibrary(root); err == nil {
		t.Fatal("loose file accepted")
	}
	if err := os.Remove(filepath.Join(root, "loose.txt")); err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "authored-message", "1")
	copy := filepath.Join(root, "duplicate")
	if err := os.Mkdir(copy, 0700); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(original)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		raw, err := os.ReadFile(filepath.Join(original, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(copy, entry.Name()), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	_, report := indexFixture(t)
	source, err := Load(report.Options.Bundle, report.Options.Source)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := collectFinishedSlides(source, LibraryIndexOptions{SlideLibrary: root}); err == nil || !strings.Contains(err.Error(), "duplicate_revision") {
		t.Fatal("duplicate accepted", err)
	}
	var decoded finishedslide.Manifest
	raw, _ := os.ReadFile(filepath.Join(copy, "manifest.json"))
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.RevisionSHA256 != m.RevisionSHA256 || !bytes.Equal(indexJSON(decoded), indexJSON(m)) {
		t.Fatal("fixture copy changed")
	}
}
