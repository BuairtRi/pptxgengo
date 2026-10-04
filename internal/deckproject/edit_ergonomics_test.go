package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestConcurrentSourceMutationGuard(t *testing.T) {
	for attempt := 0; attempt < 4; attempt++ {
		p := example(t)
		start := make(chan struct{})
		results := make(chan error, 2)
		var ready sync.WaitGroup
		ready.Add(2)
		for _, id := range []string{"maintain-the-source", "local-composition"} {
			go func(id string) {
				ready.Done()
				<-start
				_, err := OperateSlide(p, SlideOperation{Action: "hide", ID: id})
				results <- err
			}(id)
		}
		ready.Wait()
		close(start)
		successes := 0
		for i := 0; i < 2; i++ {
			if <-results == nil {
				successes++
			}
		}
		if successes != 1 {
			t.Fatalf("expected one atomic mutation, got %d", successes)
		}
		current, err := Load(p.Root)
		if err != nil {
			t.Fatal(err)
		}
		if current.Document.Slides[0].Hidden == current.Document.Slides[1].Hidden {
			t.Fatal("concurrent writes lost or merged a mutation")
		}
	}
}

func TestSlideReAddRetainedSourceAndAddAs(t *testing.T) {
	p, _ := splitExample(t, false)
	path := filepath.Join(p.Root, p.SlideFiles["local-composition"])
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = OperateSlide(p, SlideOperation{Action: "remove", ID: "local-composition"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if _, err = OperateSlide(p, SlideOperation{Action: "add", ID: "local-composition", Source: raw, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	retained, _ := os.ReadFile(path)
	if !bytes.Equal(retained, raw) || len(p.Document.Slides) != 2 {
		t.Fatal("retained file was replaced or not re-added")
	}
	if _, err = OperateSlide(p, SlideOperation{Action: "add", As: "renamed-candidate", Source: raw, Bundle: bundle(t), Engine: wmdesign.CandidateEngine, CheckFit: true}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if p.Document.Slides[2].ID != "renamed-candidate" {
		t.Fatal("--as did not assign new ID")
	}
	if now, _ := os.ReadFile(path); !bytes.Equal(now, raw) {
		t.Fatal("candidate source was modified")
	}
}

func TestSlideIntoSectionReanchorsAndRejectsOutsidePosition(t *testing.T) {
	p := example(t)
	if _, err := AddSection(p, SectionAddOptions{ID: "opening", Title: "Opening", BeforeSlideID: "maintain-the-source"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if _, err := AddSection(p, SectionAddOptions{ID: "detail", Title: "Detail", BeforeSlideID: "local-composition"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	source, err := StockScaffoldSlideSource(bundle(t), "cards/3", "candidate-027", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = OperateSlide(p, SlideOperation{Action: "add", As: "detail-opening", Source: source, IntoSection: "detail", Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if p.Document.Sections[1].BeforeSlideID != "detail-opening" || p.Document.Slides[1].ID != "detail-opening" {
		t.Fatal("slide did not become section first slide")
	}
	before := p.SourceHash()
	if _, err = OperateSlide(p, SlideOperation{Action: "add", As: "outside", Source: source, IntoSection: "detail", Before: "maintain-the-source", Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err == nil {
		t.Fatal("accepted outside section position")
	}
	p, _ = Load(p.Root)
	if p.SourceHash() != before {
		t.Fatal("invalid section move changed source")
	}
}

func TestEditFitCheckRejectsOverflowAtomically(t *testing.T) {
	p := example(t)
	values := map[string]any{}
	if err := strictInto(p.Document.Slides[0].Values, &values); err != nil {
		t.Fatal(err)
	}
	cards := values["cards"].([]any)
	cards[0].(map[string]any)["body"] = strings.Repeat("The body is far too long to fit inside this fixed card. ", 150)
	before := p.SourceHash()
	_, err := EditSlidesWithOptions(p, map[string]SlideEdit{"maintain-the-source": {Values: values}}, bundle(t), wmdesign.CandidateEngine, EditOptions{CheckFit: true})
	if err == nil || !strings.Contains(err.Error(), "fit check failed") {
		t.Fatal("overflow accepted", err)
	}
	current, _ := Load(p.Root)
	if current.SourceHash() != before {
		t.Fatal("fit failure changed source")
	}
	if _, err = StockScaffoldSlide(bundle(t), "cards/3", "invalid id", 2026); err == nil {
		t.Fatal("scaffold accepted invalid ID")
	}
}

func TestDetachPreservesFriendlyContentAndClearsStockHeader(t *testing.T) {
	p, _ := splitExample(t, false)
	source, err := StockScaffoldSlideSource(bundle(t), "cards/narrative-2x3", "statement", 2026)
	if err != nil {
		t.Fatal(err)
	}
	source = append([]byte("# Shared template: layout is unchanged\n"), source...)
	if _, err = OperateSlide(p, SlideOperation{Action: "add", ID: "statement", Source: source, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	pin(t, p)
	before, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Detach(p, "statement", "local-statement", bundle(t), wmdesign.CandidateEngine, "Adjust the statement layout"); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	raw, _ := os.ReadFile(filepath.Join(p.Root, p.SlideFiles["statement"]))
	if !bytes.Contains(raw, []byte("content:")) || bytes.Contains(raw, []byte("value-")) || bytes.Contains(raw, []byte("layout is unchanged")) {
		t.Fatalf("detached content lost readable aliases:\n%s", raw)
	}
	after, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	_, first, err := wmdesign.BuildWithEngineAndAssets(bundle(t), "", before.Document, wmdesign.CandidateEngine, before.Assets)
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := wmdesign.BuildWithEngineAndAssets(bundle(t), "", after.Document, wmdesign.CandidateEngine, after.Assets)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Slides[2].Texts) != len(second.Slides[2].Texts) {
		t.Fatal("detachment changed visible text count")
	}
	for i, text := range first.Slides[2].Texts {
		if text.Layout.Original != second.Slides[2].Texts[i].Layout.Original {
			t.Fatal("detachment changed visible copy")
		}
	}
}
