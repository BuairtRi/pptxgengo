package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	previewPath, reviewPath := filepath.Join(t.TempDir(), "preview.png"), filepath.Join(t.TempDir(), "review.json")
	if e := os.WriteFile(previewPath, []byte("Owned CLI fixture preview; not native acceptance"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(reviewPath, []byte(`{"fixture":"operator decision plumbing only"}`), 0600); e != nil {
		t.Fatal(e)
	}
	published := projectCommandJSON(t, "slide", "publish", "--project", root, "--bundle", bundle, "--id", "authored-message", "--out", out, "--library-id", "curated/slide/cli-message", "--revision", "1", "--name", "Authored command message", "--purpose", "Exercise publication and insertion", "--owner", "Fixture steward", "--preview", previewPath, "--review", reviewPath)
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
	d := finishedslide.ReviewDecision{Schema: finishedslide.ReviewDecisionSchema, RevisionSHA256: manifest.RevisionSHA256, NewRevision: 2, Action: "approve", Actor: "Owned CLI fixture steward", Date: time.Now().UTC().Format(time.DateOnly), Reason: "Fixture expressly reviews supplied artifacts without real content or native approval", ReuseScope: "Owned fixture invocations", ValidUntil: "none"}
	for _, f := range manifest.Files {
		switch f.Role {
		case "source":
			d.SourceSHA256 = f.SHA256
		case "preview":
			d.Preview = &finishedslide.ReviewedArtifact{Path: f.Path, SHA256: f.SHA256}
		case "review":
			d.Report = &finishedslide.ReviewedArtifact{Path: f.Path, SHA256: f.SHA256}
		}
	}
	raw, _ = json.Marshal(d)
	decisionFile := filepath.Join(t.TempDir(), "decision.json")
	if e := os.WriteFile(decisionFile, raw, 0600); e != nil {
		t.Fatal(e)
	}
	approvedOut := filepath.Join(t.TempDir(), "approved revision")
	approved := projectCommandJSON(t, "slide", "review-reuse", "--package", out, "--decision", decisionFile, "--out", approvedOut)
	var result finishedslide.ReviewResult
	if e := json.Unmarshal(approved, &result); e != nil || result.Manifest.Lifecycle != "approved" || result.Manifest.Approval.By != d.Actor || result.Receipt.PreviousRevisionSHA256 != manifest.RevisionSHA256 {
		t.Fatal("review command lost explicit decision", e)
	}
	inserted = projectCommandJSON(t, "slide", "insert", "--project", root, "--bundle", bundle, "--package", approvedOut, "--id", "reviewed-independent-message", "--rationale", "Owned test accepts reviewed revision in this deck")
	if e := json.Unmarshal(inserted, &receipt); e != nil || receipt.Library.Revision != 2 {
		t.Fatal("approved insertion required draft bypass", e)
	}
}

func TestFinishedSlideReviewCommandRequiresExplicitInputs(t *testing.T) {
	for _, args := range [][]string{{}, {"--package", "missing"}, {"--package", "missing", "--decision", "missing", "--out", "new", "extra"}} {
		if e := runFinishedSlideReview(args); e == nil {
			t.Fatal("incomplete decision accepted", args)
		}
	}
}
