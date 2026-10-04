package deckproject

import (
	"bytes"
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"strings"
	"testing"
)

func TestSwapCompatibleAndIncomplete(t *testing.T) {
	p := example(t)
	if _, err := Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	untouched := append([]byte(nil), p.SourceFiles[p.SlideFiles["local-composition"]]...)
	proposal, err := ProposeSwap(p, "maintain-the-source", "cards/3", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(proposal.Unmapped) != 0 || len(proposal.Missing) != 0 || len(proposal.Mapped) < 7 {
		t.Fatalf("incomplete compatible swap: %+v", proposal)
	}
	original := canonical(p.Document.Slides[0].Values)
	if _, err = ApplySwap(p, proposal, false, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if !bytes.Equal(original, canonical(p.Document.Slides[0].Values)) || !bytes.Equal(untouched, p.SourceFiles[p.SlideFiles["local-composition"]]) {
		t.Fatal("compatible swap changed copy")
	}
	proposal, err = ProposeSwap(p, "maintain-the-source", "cards/4", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(proposal.Unmapped) != 0 || len(proposal.Missing) != 2 || len(proposal.Mapped) < 11 {
		t.Fatal("cardinality change did not carry first three cards and report fourth title/body", proposal)
	}
	before := p.SourceHash()
	if _, err = ApplySwap(p, proposal, true, bundle(t), wmdesign.CandidateEngine); err == nil {
		t.Fatal("missing required content applied")
	}
	p, _ = Load(p.Root)
	if p.SourceHash() != before {
		t.Fatal("failed swap modified source")
	}
	proposal.SourceSHA256 = "stale"
	if _, err = ApplySwap(p, proposal, false, bundle(t), wmdesign.CandidateEngine); err == nil {
		t.Fatal("stale proposal applied")
	}
	if _, err = ProposeSwap(p, "local-composition", "cards/3", bundle(t)); err == nil {
		t.Fatal("guessed local mapping")
	}
}

func TestSwapSameTemplateParagraphsAndStatementHeadline(t *testing.T) {
	p := example(t)
	for _, key := range []string{"cards/narrative-2x3", "key-message/statement"} {
		authored, err := StockScaffoldSlide(bundle(t), key, "maintain-the-source", 2026)
		if err != nil {
			t.Fatal(err)
		}
		if err = p.expandContentAliases(authored, "/slide"); err != nil {
			t.Fatal(err)
		}
		var slide Slide
		if err = strictInto(authored, &slide); err != nil {
			t.Fatal(err)
		}
		p.Document.Slides[0] = slide
		proposal, err := ProposeSwap(p, slide.ID, key, bundle(t))
		if err != nil {
			t.Fatal(err)
		}
		if len(proposal.Unmapped) != 0 || len(proposal.Missing) != 0 {
			t.Fatalf("same-template fields lost for %s: %+v", key, proposal)
		}
		edit := proposal.Patch[slide.ID]
		materialized := map[string]any{"values": edit.Values, "content": edit.Content, "bindings": edit.Bindings}
		if err = json.Unmarshal(canonical(materialized), &materialized); err != nil {
			t.Fatal(err)
		}
		if err = p.expandContentAliases(materialized, "/slide"); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(canonical(materialized["values"]), canonical(slide.Values)) {
			t.Fatal("same-template canonical values changed", key)
		}
	}
	p = example(t)
	proposal, err := ProposeSwap(p, "maintain-the-source", "key-message/statement", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, field := range proposal.Mapped {
		if field.Source == "/headline" && field.Destination == "/headline" {
			found = true
		}
	}
	if !found || len(proposal.Unmapped) == 0 || len(proposal.Missing) == 0 {
		t.Fatal("statement headline bridge discarded cards or invented supporting copy", proposal)
	}
}

func TestSwapRefusesTamperedMappingAndArrayPointers(t *testing.T) {
	p := example(t)
	proposal, err := ProposeSwap(p, "maintain-the-source", "cards/3", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	before := p.SourceHash()
	edit := proposal.Patch[proposal.SlideID]
	edit.Content["headline"] = "Unreviewed replacement"
	proposal.Patch[proposal.SlideID] = edit
	if _, err = ApplySwap(p, proposal, true, bundle(t), wmdesign.CandidateEngine); err == nil || !strings.Contains(err.Error(), "verified source mapping") {
		t.Fatal("tampered proposal accepted", err)
	}
	next, err := Load(p.Root)
	if err != nil || next.SourceHash() != before {
		t.Fatal("rejected proposal changed source", err)
	}
	content := map[string]any{"items": []any{map[string]any{"text": "Original"}}}
	if err := replaceContentLeaf(content, "/items/0junk/text", "Changed"); err == nil {
		t.Fatal("partial numeric array pointer accepted")
	}
}
