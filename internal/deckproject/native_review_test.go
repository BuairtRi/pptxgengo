package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func nativeAttachmentFixture(t *testing.T) (*Project, string) {
	t.Helper()
	p := example(t)
	pin(t, p)
	r, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	deck, err := os.ReadFile(filepath.Join(p.Root, "builds", r.BuildID, "deck.pptx"))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	pdf := []byte("%PDF-1.7\nfixture; native rendering is tested separately\n")
	if err := os.WriteFile(filepath.Join(root, "deck.pdf"), pdf, 0644); err != nil {
		t.Fatal(err)
	}
	receipt := nativeexport.Receipt{Renderer: "Microsoft PowerPoint (local native PDF); macOS PDFKit PNG", Source: nativeexport.Artifact{SHA256: digest(deck)}, Slides: 2, Pages: 2, PDF: &nativeexport.Artifact{Path: "deck.pdf", SHA256: digest(pdf)}, PageMappings: []nativeexport.PageMapping{{Page: 1, SourceSlide: 1}, {Page: 2, SourceSlide: 2}}}
	if err := os.WriteFile(filepath.Join(root, "render-manifest.json"), canonical(receipt), 0644); err != nil {
		t.Fatal(err)
	}
	return p, root
}

func TestNativeAttachmentCoverageAndStaleness(t *testing.T) {
	p, root := nativeAttachmentFixture(t)
	decisions := map[string]VisualDecision{"maintain-the-source": {Status: "accepted", Reviewer: "test-reviewer"}}
	a, err := AttachNativeRender(p, root, decisions)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.SlideIDs) != 2 {
		t.Fatal(a)
	}
	if _, err := AttachNativeRender(p, root, decisions); err != nil {
		t.Fatal("idempotence", err)
	}
	coverage, err := NativeReviewCoverage(p)
	if err != nil {
		t.Fatal(err)
	}
	if !coverage[0].Rendered || coverage[0].Status != "accepted" || coverage[1].Status != "not_reviewed" {
		t.Fatal(coverage)
	}
	if _, err := OperateSlide(p, SlideOperation{Action: "hide", ID: "local-composition"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	coverage, err = NativeReviewCoverage(p)
	if err != nil {
		t.Fatal(err)
	}
	if coverage[0].Status != "stale" || coverage[0].Rendered {
		t.Fatal("changed source accepted", coverage)
	}
}

func TestNativeAttachmentRejectsDriftUnknownDecisionAndPreview(t *testing.T) {
	p, root := nativeAttachmentFixture(t)
	if _, err := AttachNativeRender(p, root, map[string]VisualDecision{"missing": {Status: "accepted", Reviewer: "test"}}); err == nil {
		t.Fatal("unknown slide accepted")
	}
	if _, err := AttachNativeRender(p, root, map[string]VisualDecision{"maintain-the-source": {Status: "accepted"}}); err == nil {
		t.Fatal("missing reviewer accepted")
	}
	manifestPath := filepath.Join(root, "render-manifest.json")
	raw, _ := os.ReadFile(manifestPath)
	if err := os.WriteFile(manifestPath, []byte(strings.Replace(string(raw), "Microsoft PowerPoint (local native PDF)", "Go preview", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := AttachNativeRender(p, root, nil); err == nil {
		t.Fatal("preview accepted")
	}
	if err := os.WriteFile(manifestPath, raw, 0644); err != nil {
		t.Fatal(err)
	}
	a, err := AttachNativeRender(p, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(p.Root, "reviews/native", a.ID, "deck.pdf")
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("drift"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := NativeReviewCoverage(p); err == nil {
		t.Fatal("changed native artifact accepted")
	}
	if _, err := AttachNativeRender(p, root, nil); err == nil {
		t.Fatal("idempotent attachment accepted drifted stored artifact")
	}
}

func TestNativeAttachmentMetadataAndGeneratedDeckDrift(t *testing.T) {
	p, root := nativeAttachmentFixture(t)
	a, err := AttachNativeRender(p, root, map[string]VisualDecision{"maintain-the-source": {Status: "accepted", Reviewer: "reviewer"}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(p.Root, "reviews/native", a.ID, "attachment.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*NativeAttachment){func(a *NativeAttachment) {
		a.Decisions["maintain-the-source"] = VisualDecision{Status: "issues_found", Reviewer: "other"}
	}, func(a *NativeAttachment) { a.SlideIDs[0], a.SlideIDs[1] = a.SlideIDs[1], a.SlideIDs[0] }, func(a *NativeAttachment) { delete(a.Files, "deck.pdf") }} {
		var edited NativeAttachment
		if err = json.Unmarshal(original, &edited); err != nil {
			t.Fatal(err)
		}
		change(&edited)
		if err = os.WriteFile(path, canonical(edited), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err = NativeReviewCoverage(p); err == nil {
			t.Fatal("mutated native attachment metadata accepted")
		}
	}
	if err = os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}
	deck := filepath.Join(p.Root, "builds", a.BuildID, "deck.pptx")
	if err = os.Chmod(deck, 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(deck, []byte("manual edit"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = NativeReviewCoverage(p); err == nil {
		t.Fatal("manually edited generated deck retained accepted coverage")
	}
}

func TestDeckReviewEvidenceIsSelfContained(t *testing.T) {
	p, root := nativeAttachmentFixture(t)
	raw, err := os.ReadFile(filepath.Join(root, "render-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt nativeexport.Receipt
	if err = json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	if err = png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 32, 18))); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		name := fmt.Sprintf("page-%d.png", i+1)
		if err = os.WriteFile(filepath.Join(root, name), buffer.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		receipt.PNGs = append(receipt.PNGs, nativeexport.Artifact{Path: name, SHA256: digest(buffer.Bytes())})
		receipt.PageMappings[i].PNG = name
	}
	if err = os.WriteFile(filepath.Join(root, "render-manifest.json"), canonical(receipt), 0644); err != nil {
		t.Fatal(err)
	}
	a, err := AttachNativeRender(p, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "packet")
	m, err := ProjectReview(p, ProjectReviewOptions{Stage: "deck", Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range m.Slides {
		for _, relative := range []string{slide.Image, slide.Review.Image, slide.NativeEvidence, slide.NativePDF} {
			if relative == "" {
				t.Fatal("missing packet evidence path", slide)
			}
			if _, err = os.Stat(filepath.Join(out, relative)); err != nil {
				t.Fatal("packet evidence refers outside its own tree", relative, err)
			}
		}
	}
	for _, relative := range []string{"attachment.json", "render-manifest.json", "deck.pdf"} {
		if _, err = os.Stat(filepath.Join(out, "native", a.ID, relative)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReviewCopiesCompositionRationale(t *testing.T) {
	p, _ := nativeAttachmentFixture(t)
	log := []byte("schema: pptxgengo.composition-log.v1\nslides:\n maintain-the-source:\n  purpose: Delivery controls\n  relationship: parallel\n  chosen_template: cards/3\n  rationale: Three parallel controls\n  candidates: [cards/3, cards/4]\n  unresolved: [Validate ownership]\n local-composition:\n  purpose: Explain customization\n  chosen_template: editorial-photo\n  rationale: Paired editorial content and photograph\n")
	if err := os.WriteFile(filepath.Join(p.Root, "composition-log.yaml"), log, 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "packet")
	m, err := ProjectReview(p, ProjectReviewOptions{Stage: "content", Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	if m.Slides[0].Composition == nil || m.Slides[0].Composition.Relationship != "parallel" {
		t.Fatal(m.Slides[0])
	}
	copied, err := os.ReadFile(filepath.Join(out, "context/composition-log.yaml"))
	if err != nil || !bytes.Equal(copied, log) {
		t.Fatal("composition evidence missing", err)
	}
	html, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Composition rationale", "Delivery controls", "Three parallel controls", "cards/4", "Validate ownership"} {
		if !bytes.Contains(html, []byte(text)) {
			t.Fatal("composition field absent from HTML", text)
		}
	}
}

func TestStagedReviewPackets(t *testing.T) {
	p, root := nativeAttachmentFixture(t)
	if _, err := AttachNativeRender(p, root, nil); err != nil {
		t.Fatal(err)
	}
	for _, stage := range []string{"outline", "content", "deck"} {
		out := filepath.Join(t.TempDir(), "packet")
		m, err := ProjectReview(p, ProjectReviewOptions{Stage: stage, Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
		if err != nil {
			t.Fatal(stage, err)
		}
		if len(m.Slides) != 2 {
			t.Fatal(m)
		}
		if _, err := os.Stat(filepath.Join(out, "index.html")); err != nil {
			t.Fatal(err)
		}
		if stage == "outline" && (m.Slides[0].Copy != nil || m.Slides[0].Notes != "") {
			t.Fatal("outline includes copy")
		}
		if _, err := ProjectReview(p, ProjectReviewOptions{Stage: stage, Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err == nil {
			t.Fatal("review overwrote output")
		}
	}
}
