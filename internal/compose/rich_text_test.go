package compose

import (
	"math"
	"testing"
)

func richParagraphFixture() []ParagraphSpec {
	return []ParagraphSpec{
		{ID: "lead", Align: "left", SpaceAfterPt: 6, Runs: []RunSpec{{ID: "bold", Text: "Bold ", FontFace: "Arial", FontSizePt: 12, Bold: true, Foreground: "#070154"}, {ID: "italic", Text: "and italic", FontFace: "Arial", FontSizePt: 12, Italic: true, Foreground: "#0047FF"}}},
		{ID: "detail", Align: "left", SpaceBeforePt: 2, Runs: []RunSpec{{ID: "underlined", Text: "Underlined detail", FontFace: "Arial", FontSizePt: 10, Underline: true, Foreground: "#070154"}}},
	}
}

func TestRichCanvasProbePreservesContract(t *testing.T) {
	s := Spec{Schema: SpecSchema, Slides: []SlideSpec{{ID: "s", WidthPt: 960, HeightPt: 540, TitleFontFace: "Arial", TitleFontSizePt: 24, Canvas: []CanvasSpec{{ID: "rich", Kind: "text", Bounds: Rect{X: 20, Y: 20, Width: 300, Height: 100}, Paragraphs: richParagraphFixture(), Align: "left", Valign: "top"}}}}}
	q, err := ProbeRequests(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(q) != 1 || len(q[0].Paragraphs) != 2 || q[0].Text != "Bold and italic\nUnderlined detail" || !q[0].Paragraphs[0].Runs[1].Italic {
		t.Fatalf("rich probe contract lost: %+v", q)
	}
	q[0].Paragraphs[0].Runs[0].Text = "changed"
	if s.Slides[0].Canvas[0].Paragraphs[0].Runs[0].Text != "Bold " {
		t.Fatal("probe result aliases source rich runs")
	}
}

func TestRichTextRejectsLegacyMixAndBadRun(t *testing.T) {
	p := richParagraphFixture()
	c := CanvasSpec{ID: "rich", Kind: "text", Bounds: Rect{X: 20, Y: 20, Width: 300, Height: 100}, Text: "legacy", Paragraphs: p, Align: "left", Valign: "top"}
	s := Spec{Schema: SpecSchema, Slides: []SlideSpec{{ID: "s", WidthPt: 960, HeightPt: 540, TitleFontFace: "Arial", TitleFontSizePt: 24, Canvas: []CanvasSpec{c}}}}
	if _, err := ProbeRequests(s); err == nil {
		t.Fatal("legacy/rich mixture accepted")
	}
	s.Slides[0].Canvas[0].Text = ""
	s.Slides[0].Canvas[0].Paragraphs[0].Runs[0].FontFace = "Helvetica"
	if _, err := ProbeRequests(s); err == nil {
		t.Fatal("unsupported rich font accepted")
	}
	s.Slides[0].Canvas[0].Paragraphs = richParagraphFixture()
	s.Slides[0].Canvas[0].Paragraphs[0].Runs[0].Text = "embedded\nparagraph"
	if _, err := ProbeRequests(s); err == nil {
		t.Fatal("embedded rich paragraph break accepted")
	}
	s.Slides[0].Canvas[0].Paragraphs = richParagraphFixture()
	s.Slides[0].Canvas[0].Paragraphs[0].Runs[0].Text = "emoji 😀"
	if _, err := ProbeRequests(s); err == nil {
		t.Fatal("non-BMP rich text accepted without verifiable native indexing")
	}
}

func TestRichTextLineSpacingBounds(t *testing.T) {
	for _, spacing := range []float64{0, .5, .9, 1, 4} {
		p := richParagraphFixture()
		p[0].LineSpacingMultiple = spacing
		if err := validateRichTextStructure(p); err != nil {
			t.Fatalf("valid spacing %v rejected: %v", spacing, err)
		}
	}
	for _, spacing := range []float64{.49, 4.01, math.NaN()} {
		p := richParagraphFixture()
		p[0].LineSpacingMultiple = spacing
		if err := validateRichTextStructure(p); err == nil {
			t.Fatalf("invalid spacing %v accepted", spacing)
		}
	}
}

func TestRichLayoutBlockLowersWithStableMeasurementID(t *testing.T) {
	s := layoutFixture()
	b := &s.Slides[0].Layouts[0].Cells[0].Blocks[0]
	b.Text, b.FontFace, b.FontSizePt, b.Foreground = "", "", 0, ""
	b.Paragraphs = richParagraphFixture()
	m := fixtureMeasurements(t, s)
	p, err := Plan(s, m)
	if err != nil {
		t.Fatal(err)
	}
	want := requestID("fixture", "canvas", "panel/a/body")
	for _, c := range p.Slides[0].Canvas {
		if c.ID == "panel/a/body" {
			if c.MeasurementID != want || len(c.Paragraphs) != 2 || c.Text != "Bold and italic\nUnderlined detail" {
				t.Fatalf("rich block lowering lost contract: %+v", c)
			}
			return
		}
	}
	t.Fatal("rich layout block was not lowered")
}
