package main

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/finishedslide"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestBrowsingCLIReusableNativeDeck(t *testing.T) {
	root := t.TempDir()
	doc := deckproject.Document{Schema: deckproject.Schema, ID: "browsing-test", Title: "Generic authored fixture", Year: 2026, Slides: []deckproject.Slide{{ID: "message", ContentKind: "supplied_content", Template: deckproject.Reference{Scope: "shared", ID: "cards/3"}, Values: map[string]any{"eyebrow": "Generic fixture", "title": "Reusable browsing fixture", "cards": []any{map[string]any{"key": "one", "title": "One", "body": "Generic fixture one."}, map[string]any{"key": "two", "title": "Two", "body": "Generic fixture two."}, map[string]any{"key": "three", "title": "Three", "body": "Generic fixture three."}}}}}}
	doc.Toolchain.Lockfile = "toolchain.lock.json"
	raw, _ := json.Marshal(doc)
	if e := os.WriteFile(filepath.Join(root, "deck.yaml"), raw, 0644); e != nil {
		t.Fatal(e)
	}
	log := deckproject.CompositionLog{Schema: "pptxgengo.composition-log.v1", Slides: map[string]deckproject.CompositionEntry{"message": {Purpose: "Generic browsing test", Rationale: "Exercise native immutable copy", ChosenTemplate: "cards/3"}}}
	raw, _ = json.Marshal(log)
	os.WriteFile(filepath.Join(root, "composition-log.yaml"), raw, 0644)
	p, e := deckproject.Load(root)
	if e != nil {
		t.Fatal(e)
	}
	bundle, e := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = deckproject.Pin(p, bundle, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	library := t.TempDir()
	revision := filepath.Join(library, "message", "1")
	m, e := deckproject.PublishFinishedSlide(p, deckproject.FinishedSlidePublishOptions{SlideID: "message", Out: revision, Bundle: bundle, Engine: wmdesign.CandidateEngine, Manifest: finishedslide.Manifest{ID: "curated/slide/fixture-message", Revision: 1, Name: "Generic fixture", Purpose: "Exercise editable browsing content", Owner: "Unit test fixture", Lifecycle: "approved", Approval: &finishedslide.Approval{By: "Unit test fixture", Date: "2026-01-01", ReuseScope: "Generic tests only"}}, Artifacts: map[string][]byte{"preview.png": []byte("fixture not real approval"), "review.json": []byte(`{"fixture_only":true}`)}})
	if e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(t.TempDir(), "browsing with spaces")
	args := []string{"--kind", "reusable", "--engine", wmdesign.CandidateEngine, "--bundle", bundle, "--as-of", "2026-10-07", "--finished-library", library, "--out", out}
	if e = runBrowsingLibrary(args); e != nil {
		t.Fatal(e)
	}
	z, e := zip.OpenReader(filepath.Join(out, "reusable-slides.pptx"))
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	slides := 0
	for _, f := range z.File {
		if filepath.Dir(f.Name) == "ppt/slides" && filepath.Ext(f.Name) == ".xml" {
			slides++
		}
	}
	if slides != 3 {
		t.Fatal("guide, visible metadata, native content required", slides)
	}
	raw, e = os.ReadFile(filepath.Join(out, "browsing-manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var manifest struct {
		Schema   string                     `json:"schema"`
		Kind     string                     `json:"kind"`
		Coverage wmdesign.BrowsingSelection `json:"coverage"`
	}
	if e = json.Unmarshal(raw, &manifest); e != nil {
		t.Fatal(e)
	}
	if manifest.Kind != "reusable" || len(manifest.Coverage.Revisions) != 1 || manifest.Coverage.Revisions[0].Manifest.RevisionSHA256 != m.RevisionSHA256 {
		t.Fatal(manifest)
	}
	if e = runBrowsingLibrary(args); e == nil {
		t.Fatal("existing output overwritten")
	}
}
func TestBrowsingCLIRequiresExplicitScopeAndDate(t *testing.T) {
	for _, args := range [][]string{{}, {"--kind", "templates", "--out", filepath.Join(t.TempDir(), "new")}, {"--kind", "reusable", "--as-of", "2026-10-07", "--out", filepath.Join(t.TempDir(), "new")}} {
		if e := runBrowsingLibrary(args); e == nil {
			t.Fatal("missing required inputs accepted", args)
		}
	}
}
