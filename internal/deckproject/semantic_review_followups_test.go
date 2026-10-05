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
	// An identical edit must keep the serialized slide source byte-stable after
	// its first refresh, including generated capacity comments.
	for i := 0; i < 3; i++ {
		p, err = Load(p.Root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = EditSlides(p, map[string]SlideEdit{"maintain-the-source": {Brief: &brief}}, bundle(t), wmdesign.CandidateEngine); err != nil {
			t.Fatal(err)
		}
		next, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if i > 0 && !bytes.Equal(data, next) {
			t.Fatalf("identical edit %d changed slide comments/source", i+1)
		}
		if strings.Count(string(next), "Metadata: ") != strings.Count(string(next), "Stock slot: ") {
			t.Fatal("capacity metadata blocks are not one per stock slot")
		}
		data = next
	}
}

func TestSwapRefreshesCapacityCommentsForTargetTemplate(t *testing.T) {
	p, _ := splitExample(t, true)
	id := "maintain-the-source"
	relative := p.SlideFiles[id]
	path := filepath.Join(p.Root, relative)
	oldSource, err := StockScaffoldSlideSource(bundle(t), "cards/4", id, p.Document.Year)
	if err != nil {
		t.Fatal(err)
	}
	oldNode, err := sourceYAML(oldSource)
	if err != nil {
		t.Fatal(err)
	}
	content := mappingNode(oldNode.Content[0], "content")
	if content == nil {
		t.Fatal("scaffold body alias missing")
	}
	content.HeadComment = "Human note: preserve this source note.\n" + content.HeadComment
	content.HeadComment += "\n# Stock slot: obsolete content\n# approximate capacity: stale duplicate\n# Metadata: inferred_from_pinned_source; native fit not evaluated."
	cards := mappingNode(content, "cards")
	if cards == nil {
		t.Fatal("scaffold cards missing")
	}
	cards.HeadComment = "Human note: preserve card context.\n# Stock slot: obsolete key attachment\n# approximate capacity: stale key duplicate\n# Metadata: inferred_from_pinned_source; native fit not evaluated."
	lastBody := stockYAMLAt(oldNode.Content[0], "/content/cards/3/body")
	if lastBody == nil {
		t.Fatal("fourth card body missing")
	}
	lastBody.FootComment = "# Stock slot: obsolete footer attachment\n# approximate capacity: stale footer duplicate\n# Metadata: inferred_from_pinned_source; native fit not evaluated."
	oldSource, err = encodeSourceYAML(oldNode)
	if err != nil {
		t.Fatal(err)
	}
	for _, seeded := range []string{
		"Human note: preserve this source note.", "Human note: preserve card context.",
		"obsolete content", "obsolete key attachment", "obsolete footer attachment",
	} {
		if !bytes.Contains(oldSource, []byte(seeded)) {
			t.Fatalf("seeded comment %q was not encoded\n%s", seeded, oldSource)
		}
	}
	if err = os.WriteFile(path, oldSource, 0600); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := ProposeSwap(p, id, "cards/3", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(proposal.Missing) != 0 {
		t.Fatalf("fixture swap unexpectedly needs copy: %v", proposal.Missing)
	}
	if _, err = ApplySwap(p, proposal, true, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	targetSource, err := StockScaffoldSlideSource(bundle(t), "cards/3", id, p.Document.Year)
	if err != nil {
		t.Fatal(err)
	}
	oldCapacity := stockCapacityForAlias(oldSource, "Card 1 body")
	targetCapacity := stockCapacityForAlias(targetSource, "Card 1 body")
	if oldCapacity == "" || targetCapacity == "" || oldCapacity == targetCapacity {
		t.Fatalf("fixture does not distinguish template capacity: old=%q target=%q", oldCapacity, targetCapacity)
	}
	text := string(got)
	if !strings.Contains(text, targetCapacity) || strings.Contains(text, oldCapacity) {
		t.Fatalf("swap did not refresh generated capacity: old=%q target=%q", oldCapacity, targetCapacity)
	}
	if strings.Count(text, "Human note: preserve this source note.") != 1 || strings.Count(text, "Human note: preserve card context.") != 1 || strings.Contains(text, "obsolete content") || strings.Contains(text, "obsolete key attachment") || strings.Contains(text, "obsolete footer attachment") || strings.Contains(text, "stale duplicate") || strings.Contains(text, "stale key duplicate") || strings.Contains(text, "stale footer duplicate") {
		t.Fatalf("swap lost or duplicated the human comment\n%s", text)
	}
	if strings.Contains(text, "Card 4 body") {
		t.Fatal("removed source-only slot comment survived swap")
	}
	// Repeating a same-template swap after metadata refresh must keep the file
	// byte-identical rather than accumulating comments.
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	firstProposal, err := ProposeSwap(p, id, "cards/3", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ApplySwap(p, firstProposal, true, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	secondProposal, err := ProposeSwap(p, id, "cards/3", bundle(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ApplySwap(p, secondProposal, true, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("repeated same-template swap changed source: %v", err)
	}
}

func stockCapacityForAlias(raw []byte, slotDescription string) string {
	lines := strings.Split(string(raw), "\n")
	for i := 0; i+2 < len(lines); i++ {
		if strings.Contains(stockCommentLine(lines[i]), "Stock slot:") && strings.Contains(stockCommentLine(lines[i]), slotDescription) {
			return stockCommentLine(lines[i+1])
		}
	}
	return ""
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
