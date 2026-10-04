package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/nativeexport"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestAudienceCopyStagesIgnoreUnsignedHistoricalNativeEvidence(t *testing.T) {
	p, root := nativeAttachmentFixture(t)
	a, err := AttachNativeRender(p, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(p.Root, "reviews/native", a.ID)
	manifestPath := filepath.Join(old, "render-manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest nativeexport.Receipt
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Provenance = nil
	raw = canonical(manifest)
	if err = os.Chmod(manifestPath, 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(manifestPath, raw, 0644); err != nil {
		t.Fatal(err)
	}
	a.RenderManifestSHA256 = digest(raw)
	a.Files["render-manifest.json"] = digest(raw)
	a.ID = nativeAttachmentID(a)
	current := filepath.Join(p.Root, "reviews/native", a.ID)
	if err = os.Rename(old, current); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(current, "attachment.json")
	if err = os.Chmod(metadata, 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(metadata, canonical(a), 0644); err != nil {
		t.Fatal(err)
	}
	for _, stage := range []string{"outline", "content"} {
		packet, err := ProjectReview(p, ProjectReviewOptions{Stage: stage, Out: filepath.Join(t.TempDir(), "packet"), Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Audience: true})
		if err != nil {
			t.Fatal("copy-only review required old native trust", stage, err)
		}
		if len(packet.Slides) != len(p.Document.Slides) {
			t.Fatal("copy stages dropped source slides")
		}
	}
	_, err = ProjectReview(p, ProjectReviewOptions{Stage: "deck", Out: filepath.Join(t.TempDir(), "packet"), Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Audience: true})
	if err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatal("deck review accepted unsigned historical native evidence", err)
	}
}

func TestAudienceStockChartCopyReportsGapWithoutWorkbookLeak(t *testing.T) {
	p := example(t)
	const key = "chart/column-full"
	authored, err := StockScaffoldSlide(bundle(t), key, "native-chart", 2026)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := wmdesign.LibraryCatalog(bundle(t), "")
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, definition := range catalog {
		if definition.Key != key {
			continue
		}
		for _, slot := range definition.Slots {
			if !strings.HasSuffix(slot.SourcePointer, "/series/0/name") {
				continue
			}
			for alias, pointer := range authored["bindings"].(map[string]any) {
				if pointer != "/slots/"+slot.Name {
					continue
				}
				if err = replaceContentLeaf(authored["content"].(map[string]any), alias, "SECRET-NONDISPLAYED-WORKBOOK-SERIES"); err != nil {
					t.Fatal(err)
				}
				changed = true
			}
		}
	}
	if !changed {
		t.Fatal("stock native chart did not expose expected series-name binding")
	}
	// This source-pinned single-series column chart has ShowLegend=false and
	// ShowSerName=false. Its workbook series name is not displayed on the slide.
	data, err := MarshalSlideSource(authored)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = OperateSlide(p, SlideOperation{Action: "add", ID: "native-chart", Source: data, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	pin(t, p)
	out := filepath.Join(t.TempDir(), "packet")
	packet, err := ProjectReview(p, ProjectReviewOptions{Stage: "content", Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Audience: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, slide := range packet.Slides {
		if slide.ID == "native-chart" {
			found = true
			if len(slide.CopyReviewGaps) == 0 || !strings.Contains(slide.CopyReviewGaps[0], "rendered slide") {
				t.Fatal("chart text review gap was not explicit", slide)
			}
		}
	}
	if !found {
		t.Fatal("stock native chart absent from audience content packet")
	}
	for relative := range packet.Files {
		raw, err := os.ReadFile(filepath.Join(out, relative))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte("SECRET-NONDISPLAYED-WORKBOOK-SERIES")) {
			t.Fatal("nondisplayed workbook data leaked", relative)
		}
	}
}

func TestAudienceReviewExcludesInternalMaterial(t *testing.T) {
	p := example(t)
	p = rewrite(t, p, "    template: {scope: shared, id: cards/3}", "    notes: SECRET-SPEAKER-NOTES\n    brief: context/brief.yaml\n    template: {scope: shared, id: cards/3}")
	p = rewrite(t, p, "  project: context/project.md", "  project: context/project.md\n  audience: context/audience.md\n  outline: context/outline.md\n  decisions: context/decisions.md\n  win_strategy: context/win.md")
	for path, text := range map[string]string{"audience.md": "Audience decision context", "outline.md": "Audience outline", "decisions.md": "SECRET-DECISIONS", "win.md": "SECRET-WIN-LOGIC", "brief.yaml": "SECRET-SLIDE-BRIEF"} {
		if err := os.WriteFile(filepath.Join(p.Root, "context", path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := OperateSlide(p, SlideOperation{Action: "hide", ID: "local-composition"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	pin(t, p)
	if _, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	for _, stage := range []string{"outline", "content"} {
		out := filepath.Join(t.TempDir(), "packet")
		m, err := ProjectReview(p, ProjectReviewOptions{Stage: stage, Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Audience: true})
		if err != nil {
			t.Fatal(stage, err)
		}
		if len(m.Slides) != 1 || !m.Audience {
			t.Fatalf("hidden slide leaked: %+v", m)
		}
		if _, ok := m.Files["context/audience.md"]; !ok {
			t.Fatal("audience context missing")
		}
		if stage == "outline" && len(m.Slides[0].VisibleCopy) != 0 {
			t.Fatal("outline exposes detailed copy")
		}
		if stage != "outline" && len(m.Slides[0].VisibleCopy) == 0 {
			t.Fatal("visible copy missing")
		}
		for path := range m.Files {
			raw, err := os.ReadFile(filepath.Join(out, path))
			if err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{"SECRET-", "Customize a local template", "Composition rationale"} {
				if bytes.Contains(raw, []byte(secret)) {
					t.Fatalf("%s leaked %s", path, secret)
				}
			}
			if strings.HasSuffix(path, ".pptx") {
				t.Fatal("PPTX may expose notes", path)
			}
		}
		html, _ := os.ReadFile(filepath.Join(out, "index.html"))
		if stage != "outline" && !bytes.Contains(html, []byte("Edit the human-readable YAML source.")) {
			t.Fatal("HTML lacks visible body copy")
		}
	}
	out := filepath.Join(t.TempDir(), "unrendered-deck")
	if _, err := ProjectReview(p, ProjectReviewOptions{Stage: "deck", Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Audience: true}); err == nil || !strings.Contains(err.Error(), "requires current native output") {
		t.Fatalf("unrendered audience deck accepted: %v", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("failed audience review published a packet", err)
	}
}

func TestAudienceReviewOmitsPriorNativeVerdicts(t *testing.T) {
	p, root := nativeAttachmentFixture(t)
	if _, err := AttachNativeRender(p, root, map[string]VisualDecision{"maintain-the-source": {Status: "accepted", Reviewer: "SECRET-REVIEWER", Note: "SECRET-VERDICT"}}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "packet")
	m, err := ProjectReview(p, ProjectReviewOptions{Stage: "deck", Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Audience: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range m.Slides {
		if s.Review.Status != "rendered_review_pending" || s.NativeEvidence != "" || s.Review.Note != "" {
			t.Fatalf("verdict exposed: %+v", s)
		}
	}
	for path := range m.Files {
		if strings.HasSuffix(path, "attachment.json") || strings.HasSuffix(path, "render-manifest.json") || strings.HasSuffix(path, ".pptx") {
			t.Fatal("internal native metadata copied", path)
		}
		raw, _ := os.ReadFile(filepath.Join(out, path))
		if bytes.Contains(raw, []byte("SECRET-")) {
			t.Fatal("prior verdict leaked", path)
		}
	}
}
