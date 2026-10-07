package deckproject

import (
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestCompileBrowsingRevision(t *testing.T) {
	p := reuseProject(t)
	root, m := publishReuse(t, p, 1)
	if _, e := CompileBrowsingRevision(root, bundle(t), wmdesign.CandidateEngine, "browse-copy", m); e == nil {
		t.Fatal("draft compiled")
	}
	files, e := finishedslide.ReadPayload(root, m)
	if e != nil {
		t.Fatal(e)
	}
	// Explicit generic fixture metadata is test input, never an operator approval.
	files["preview.png"] = []byte("generic unit test preview")
	files["review.json"] = []byte(`{"fixture_only":true}`)
	m.Files = append(m.Files, finishedslide.File{Path: "preview.png", Role: "preview"}, finishedslide.File{Path: "review.json", Role: "review"})
	m.Lifecycle = "approved"
	m.Approval = &finishedslide.Approval{By: "Unit test fixture", Date: "2026-01-01", ReuseScope: "Generic unit tests only"}
	approved := filepath.Join(t.TempDir(), "approved-fixture")
	m, e = finishedslide.Create(approved, m, files)
	if e != nil {
		t.Fatal(e)
	}
	c, e := CompileBrowsingRevision(approved, bundle(t), wmdesign.CandidateEngine, "browse-copy", m)
	if e != nil {
		t.Fatal(e)
	}
	if len(c.Document.Slides) != 1 || c.Document.Slides[0].ID != "browse-copy" {
		t.Fatal(c)
	}
	native, _, e := wmdesign.BuildWithEngineAndAssets(bundle(t), "", c.Document, wmdesign.CandidateEngine, c.Assets)
	if e != nil || len(native) == 0 {
		t.Fatal(e)
	}
}
