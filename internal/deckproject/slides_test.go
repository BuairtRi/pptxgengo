package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestSlideOperationsPreserveExternalSources(t *testing.T) {
	p := example(t)
	if _, err := Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); err != nil {
		t.Fatal(err)
	}
	p, err := Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	untouched := append([]byte(nil), p.SourceFiles[p.SlideFiles["maintain-the-source"]]...)
	r, err := OperateSlide(p, SlideOperation{Action: "hide", ID: "local-composition"})
	if err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if !p.Document.Slides[1].Hidden || !bytes.Equal(untouched, p.SourceFiles[p.SlideFiles["maintain-the-source"]]) {
		t.Fatal("hide changed other copy")
	}
	if _, err := os.Stat(filepath.Join(p.Root, r.Decision)); err != nil {
		t.Fatal(err)
	}
	if _, err := OperateSlide(p, SlideOperation{Action: "show", ID: "local-composition"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if _, err := OperateSlide(p, SlideOperation{Action: "move", ID: "local-composition", Before: "maintain-the-source"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if p.Document.Slides[0].ID != "local-composition" {
		t.Fatal("move order")
	}
	if _, err := OperateSlide(p, SlideOperation{Action: "remove", ID: "local-composition"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if len(p.Document.Slides) != 1 || p.Document.Slides[0].ID != "maintain-the-source" {
		t.Fatal("remove order")
	}
	if _, err := OperateSlide(p, SlideOperation{Action: "remove", ID: "maintain-the-source"}); err == nil {
		t.Fatal("removed final slide")
	}
}

func TestSlideAddValidationAndSectionAnchors(t *testing.T) {
	p := example(t)
	if _, err := AddSection(p, SectionAddOptions{ID: "opening", Title: "Opening", BeforeSlideID: "maintain-the-source"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if _, err := AddSection(p, SectionAddOptions{ID: "second", Title: "Second", BeforeSlideID: "local-composition"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	before := append([]byte(nil), p.Raw...)
	if _, err := OperateSlide(p, SlideOperation{Action: "move", ID: "maintain-the-source", After: "local-composition"}); err == nil || !strings.Contains(err.Error(), "--reanchor") {
		t.Fatal("anchor move must be explicit", err)
	}
	if data, _ := os.ReadFile(p.SourcePath); !bytes.Equal(data, before) {
		t.Fatal("failed mutation changed source")
	}
	if _, err := OperateSlide(p, SlideOperation{Action: "remove", ID: "maintain-the-source"}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if len(p.Document.Sections) != 1 || p.Document.Sections[0].ID != "second" {
		t.Fatal("empty section not removed")
	}
	source := []byte("id: extra\ncontent_kind: supplied_content\ntemplate: {scope: shared, id: cards/3}\nvalues: {}\n")
	if _, err := OperateSlide(p, SlideOperation{Action: "add", ID: "extra", Source: source, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err == nil {
		t.Fatal("invalid content accepted")
	}
	stock, err := StockScaffoldSlide(bundle(t), "cards/3", "extra", 2026)
	if err != nil {
		t.Fatal(err)
	}
	source, err = MarshalSlideSource(stock)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OperateSlide(p, SlideOperation{Action: "add", ID: "extra", Source: source, Before: "local-composition", Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if p.Document.Slides[0].ID != "extra" || p.Document.Sections[0].BeforeSlideID != "extra" {
		t.Fatal("first section failed to include added prefix")
	}
}

func TestSourceMutationsShareGuard(t *testing.T) {
	p := example(t)
	if err := os.WriteFile(filepath.Join(p.Root, ".deck-source-mutation.lock"), []byte("another writer"), 0600); err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), p.Raw...)
	operations := []func() error{
		func() error {
			_, err := OperateSlide(p, SlideOperation{Action: "hide", ID: "maintain-the-source"})
			return err
		},
		func() error {
			_, err := AddSection(p, SectionAddOptions{ID: "opening", Title: "Opening", BeforeSlideID: "maintain-the-source"})
			return err
		},
		func() error {
			_, err := EditSlides(p, map[string]SlideEdit{"maintain-the-source": {Values: p.Document.Slides[0].Values}}, bundle(t), wmdesign.CandidateEngine)
			return err
		},
	}
	for _, run := range operations {
		if err := run(); err == nil || !strings.Contains(err.Error(), "active") {
			t.Fatal("guard not honored", err)
		}
	}
	raw, err := os.ReadFile(p.SourcePath)
	if err != nil || !bytes.Equal(raw, before) {
		t.Fatal("guarded mutation changed source", err)
	}
}
