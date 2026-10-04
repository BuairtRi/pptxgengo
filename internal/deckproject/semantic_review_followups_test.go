package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestStockCommentRefreshPreservesHumanNotesAcrossEdits(t *testing.T) {
	p, _ := splitExample(t, true)
	path := filepath.Join(p.Root, p.SlideFiles["maintain-the-source"])
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	node, err := sourceYAML(raw)
	if err != nil {
		t.Fatal(err)
	}
	leaf := stockYAMLAt(node.Content[0], "/content/cards/0/body")
	if leaf == nil {
		t.Fatal("card content missing")
	}
	leaf.HeadComment = "Human drafting note: retain the interview qualifier."
	leaf.LineComment = "Human line note"
	leaf.FootComment = "Human foot note"
	raw, err = encodeSourceYAML(node)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	brief := "context/brief.yaml"
	if err = os.WriteFile(filepath.Join(p.Root, brief), []byte("An internal brief."), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = EditSlides(p, map[string]SlideEdit{"maintain-the-source": {Brief: &brief}}, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	assertComments := func() {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, text := range []string{"Human drafting note", "Human line note", "Human foot note", "Stock slot:", "Metadata:"} {
			if !bytes.Contains(data, []byte(text)) {
				t.Fatalf("comment lost: %s\n%s", text, data)
			}
		}
		if strings.Count(string(data), "Human drafting note") != 1 {
			t.Fatal("comment duplicated")
		}
	}
	assertComments()
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	authored, err := StockEditableSlide(p.Document.Slides[0], matchDef(t, "cards/3"))
	if err != nil {
		t.Fatal(err)
	}
	if err = replaceContentLeaf(authored["content"].(map[string]any), "/cards/0/body", "Changed supplied copy."); err != nil {
		t.Fatal(err)
	}
	var edit SlideEdit
	if err = strictInto(map[string]any{"content": authored["content"], "bindings": authored["bindings"]}, &edit); err != nil {
		t.Fatal(err)
	}
	if _, err = EditSlides(p, map[string]SlideEdit{"maintain-the-source": edit}, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	assertComments()
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = EditSlides(p, map[string]SlideEdit{"maintain-the-source": {Brief: &brief}}, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	assertComments()
	data, _ := os.ReadFile(path)
	if strings.Count(string(data), "Stock slot: body (string) at /cards/0/body") > 1 {
		t.Fatal("generated metadata duplicated")
	}
}

func TestSwapReceiptOperationIsSwap(t *testing.T) {
	p := example(t)
	proposal, err := ProposeSwap(p, "maintain-the-source", "cards/3", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := ApplySwap(p, proposal, false, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Operation != "swap" || !strings.Contains(receipt.Decision, "/swap-") {
		t.Fatal("wrong swap receipt", receipt)
	}
	data, err := os.ReadFile(filepath.Join(p.Root, receipt.Decision))
	if err != nil || !bytes.Contains(data, []byte(`"operation": "swap"`)) {
		t.Fatal("swap decision disagrees", err, string(data))
	}
}
