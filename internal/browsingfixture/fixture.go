// Package browsingfixture contains generic artifacts exclusively for tests.
// Its fictional input hashes and test-only approvals never qualify real content.
package browsingfixture

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
	"github.com/buairtri/pptxgengo/pptx"
)

func Deck(t testing.TB, kind string) []byte {
	t.Helper()
	p := pptx.New()
	count := 3
	if kind == "templates" {
		count = 5
	}
	for i := 0; i < count; i++ {
		s := p.AddSlide()
		s.PresSlide().Name = fmt.Sprintf("page-%d", i+1)
		if e := s.AddText([]pptx.TextProps{{Text: fmt.Sprintf("Generic native browsing fixture %d", i)}}, nil); e != nil {
			t.Fatal(e)
		}
	}
	raw, e := p.Write()
	if e != nil {
		t.Fatal(e)
	}
	return raw
}
func Hash(raw []byte) string { sum := sha256.Sum256(raw); return fmt.Sprintf("%x", sum) }
func Manifest(t testing.TB, kind, deckHash string) []byte {
	t.Helper()
	pin := strings.Repeat("a", 64)
	source := []map[string]any{{"path": "templates/fixture.json", "sha256": pin}, {"path": "frames/v0/frames.json", "sha256": pin}}
	var coverage any
	if kind == "templates" {
		coverage = map[string]any{"schema": "pptxgengo.browsing-coverage.v1", "frame_mode": "catalog", "source_revision": "fixture-v1", "source_commit": strings.Repeat("a", 40), "source_files": source, "expected_templates": 1, "expected_frame_requests": 1, "frame_candidates": 1, "entries": []map[string]any{{"slide_id": "page-3", "kind": "template", "key": "cards/3", "revision": 1, "lifecycle": "active", "source_sha256": pin}, {"slide_id": "page-5", "kind": "frame", "source_sha256": pin, "key": "wmds/frame/none-compact", "frame": map[string]any{"rail": "none", "footer": "compact"}}}}
	} else {
		m := finishedslide.Manifest{Schema: finishedslide.Schema, ID: "curated/slide/generic-test", Revision: 1, Name: "Generic fixture", Purpose: "Unit tests only", Owner: "Unit test fixture", Lifecycle: "approved", Approval: &finishedslide.Approval{By: "Unit test fixture", Date: "2026-01-01", ReuseScope: "Generic tests only"}, Source: "slide.yaml", Pins: finishedslide.Pins{Bundle: pin, SourceRevision: "fixture-v1", TemplateID: "cards/3", TemplateRevision: 1, TemplateSourceSHA256: pin, TemplateDefinitionSHA256: pin, ToolchainLockSHA256: pin, Compiler: "generic-test-only"}, Files: []finishedslide.File{{Path: "slide.yaml", Role: "source", Bytes: 1, SHA256: pin}, {Path: "preview.png", Role: "preview", Bytes: 1, SHA256: pin}, {Path: "review.json", Role: "review", Bytes: 1, SHA256: pin}}}
		if e := m.Seal(); e != nil {
			t.Fatal(e)
		}
		if e := m.Validate(); e != nil {
			t.Fatal(e)
		}
		coverage = map[string]any{"schema": "pptxgengo.reusable-browsing-selection.v1", "as_of": "2026-10-07", "library_sha256": pin, "revisions": []map[string]any{{"path": "message/1", "manifest": m, "included": true, "reason": "latest_approved"}}}
	}
	pages := []map[string]any{{"id": "page-1", "kind": "guide"}, {"id": "page-2", "kind": "reuse_metadata"}, {"id": "page-3", "kind": "reusable_slide"}}
	if kind == "templates" {
		pages = []map[string]any{{"id": "page-1", "kind": "guide"}, {"id": "page-2", "kind": "family_divider"}, {"id": "page-3", "kind": "template"}, {"id": "page-4", "kind": "frame_divider"}, {"id": "page-5", "kind": "frame"}}
	}
	raw, e := json.Marshal(map[string]any{"schema": "pptxgengo.browsing-library.v1", "kind": kind, "as_of": "2026-10-07", "deck_sha256": deckHash, "bundle_sha256": pin, "source_revision": "fixture-v1", "source_commit": strings.Repeat("a", 40), "slides": len(pages), "pages": pages, "compiler": "unit-test-only", "release_identity": "unit-test-only", "qualification": "native_visual_copy_paste_qualification_pending", "coverage": coverage, "source_files": source, "fonts": []map[string]any{{"file": "fixture.ttf", "sha256": pin, "postscript_name": "Fixture"}}})
	if e != nil {
		t.Fatal(e)
	}
	return raw
}
