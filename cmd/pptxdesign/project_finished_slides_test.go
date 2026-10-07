package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/finishedslide"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestFinishedSlideProjectCommands(t *testing.T) {
	root := t.TempDir()
	doc := deckproject.Document{Schema: deckproject.Schema, ID: "finished-cli-fixture", Title: "Authored CLI fixture", Year: 2026, Slides: []deckproject.Slide{{ID: "authored-message", ContentKind: "supplied_content", Template: deckproject.Reference{Scope: "shared", ID: "cards/3"}, Values: map[string]any{"title": "Keep maintained content independent", "eyebrow": "Command fixture", "cards": []any{map[string]any{"key": "one", "title": "Publish", "body": "Close the authored revision."}, map[string]any{"key": "two", "title": "Insert", "body": "Create independent copies."}, map[string]any{"key": "three", "title": "Review", "body": "Review the new deck context."}}}}}}
	doc.Toolchain.Lockfile = "toolchain.lock.json"
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "deck.yaml"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	log := deckproject.CompositionLog{Schema: "pptxgengo.composition-log.v1", Slides: map[string]deckproject.CompositionEntry{"authored-message": {Purpose: "Command fixture", Rationale: "Exercise maintained authored content", ChosenTemplate: "cards/3"}}}
	raw, err = json.Marshal(log)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "composition-log.yaml"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	p, err := deckproject.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := deckproject.Pin(p, bundle, wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "revision with spaces")
	published := projectCommandJSON(t, "slide", "publish", "--project", root, "--bundle", bundle, "--id", "authored-message", "--out", out, "--library-id", "curated/slide/cli-message", "--revision", "1", "--name", "Authored command message", "--purpose", "Exercise publication and insertion", "--owner", "Fixture steward")
	var manifest finishedslide.Manifest
	if err := json.Unmarshal(published, &manifest); err != nil || manifest.Lifecycle != "draft" || manifest.Approval != nil {
		t.Fatal("publication invented approval", err, manifest)
	}
	inserted := projectCommandJSON(t, "slide", "insert", "--project", root, "--bundle", bundle, "--package", out, "--id", "independent-message", "--rationale", "Fixture content belongs after the original", "--after", "authored-message", "--allow-draft")
	var receipt deckproject.FinishedSlideInsertReceipt
	if err := json.Unmarshal(inserted, &receipt); err != nil || receipt.Library.RevisionSHA256 != manifest.RevisionSHA256 {
		t.Fatal(receipt, err)
	}
	p, err = deckproject.Load(root)
	if err != nil || len(p.Document.Slides) != 2 {
		t.Fatal(p, err)
	}
	if _, err := deckproject.Check(p, bundle, wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
}
