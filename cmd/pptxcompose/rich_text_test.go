package main

import (
	"bytes"
	"github.com/buairtri/pptxgengo/internal/compose"
	"testing"
)

func TestRichContractAndNativeStyleCheck(t *testing.T) {
	p := []compose.ParagraphSpec{{ID: "p", Align: "left", SpaceAfterPt: 4, Runs: []compose.RunSpec{{ID: "a", Text: "A", FontFace: "Arial", FontSizePt: 12, Bold: true, Foreground: "#070154"}, {ID: "b", Text: "b", FontFace: "Arial", FontSizePt: 11, Italic: true, Underline: true, Foreground: "#0047FF"}}}}
	row := nativeRow{Paragraphs: []nativeParagraph{{Alignment: "paragraph align left", SpaceAfterPt: 4, LineRuleWithin: true, SpaceWithin: 1}}, Characters: []nativeCharacter{{Text: "A", FontName: "Arial", FontSizePt: 12, Bold: true, Underline: "no underline", Color: []int{7, 1, 84}}, {Text: "b", FontName: "Arial", FontSizePt: 11, Italic: true, Underline: "underline single line", Color: []int{0, 71, 255}}}}
	if err := checkNativeRichText("shape", p, row); err != nil {
		t.Fatal(err)
	}
	row.Characters[1].Italic = false
	if err := checkNativeRichText("shape", p, row); err == nil {
		t.Fatal("native rich style mismatch accepted")
	}
	q := compose.ProbeRequest{ID: "q", Text: "Ab", Paragraphs: p, TextWidthPt: 100, Background: "#FFFFFF"}
	key := contractKey(q)
	q.Paragraphs[0].Runs[1].Italic = false
	if contractKey(q) == key {
		t.Fatal("rich run style omitted from cache contract")
	}
	q.Paragraphs[0].LineSpacingMultiple = .9
	if contractKey(q) == key {
		t.Fatal("paragraph line spacing omitted from cache contract")
	}
	row.Paragraphs[0].SpaceWithin = .9
	p[0].LineSpacingMultiple = 1
	if err := checkNativeRichText("shape", p, row); err == nil {
		t.Fatal("native paragraph line-spacing mismatch accepted")
	}
}

func TestRichParagraphLineSpacingXML(t *testing.T) {
	p := compose.ParagraphSpec{ID: "p", Align: "left", LineSpacingMultiple: .9, Runs: []compose.RunSpec{{ID: "r", Text: "Two wrapped lines for spacing", FontFace: "Arial", FontSizePt: 9, Italic: true, Foreground: "#070154"}}}
	slide := renderSlide{ID: "rich-spacing", Width: 960, Height: 540, Elements: []element{{Name: "rich", Kind: "text", Frame: frame{X: 20, Y: 20, Width: 60, Height: 40}, Paragraphs: []compose.ParagraphSpec{p}, Text: "Two wrapped lines for spacing", Align: "left", Valign: "top"}}}
	deck, err := render([]renderSlide{slide})
	if err != nil {
		t.Fatal(err)
	}
	xml := shapeZipPart(t, deck, "ppt/slides/slide1.xml")
	if !bytes.Contains(xml, []byte(`<a:lnSpc><a:spcPct val="90000"/></a:lnSpc>`)) {
		t.Fatal("rich paragraph 0.9 line spacing missing from OOXML")
	}
}
