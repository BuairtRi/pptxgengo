package compose

import (
	"strings"
	"testing"
)

func TestPhraseSelectors(t *testing.T) {
	a, b, err := ResolvePhrase("Act now, then act now", PhraseRequest{ID: "x", Phrase: "act now"})
	if err != nil || a != 14 || b != 21 {
		t.Fatalf("exact selection: %d %d %v", a, b, err)
	}
	_, _, err = ResolvePhrase("now and now", PhraseRequest{ID: "x", Phrase: "now"})
	if err == nil {
		t.Fatal("ambiguous accepted")
	}
	a, b, err = ResolvePhrase("now and now", PhraseRequest{ID: "x", Phrase: "now", Occurrence: 2})
	if err != nil || a != 8 || b != 11 {
		t.Fatalf("occurrence: %d %d %v", a, b, err)
	}
	lo, hi := 2, 6
	a, b, err = ResolvePhrase("é now!", PhraseRequest{ID: "x", Phrase: "now!", StartRune: &lo, EndRune: &hi})
	if err != nil || a != 2 || b != 6 {
		t.Fatalf("unicode range: %d %d %v", a, b, err)
	}
	for _, q := range []PhraseRequest{{ID: "x", Phrase: "absent"}, {ID: "x", Phrase: "now", Occurrence: -1}, {ID: "x", Phrase: "now", Occurrence: 3}, {ID: "x", Phrase: "now", StartRune: &lo}} {
		if _, _, err = ResolvePhrase("now and now", q); err == nil {
			t.Fatalf("bad selector accepted %+v", q)
		}
	}
}
func accentTestFixture() (SlideSpec, *PlannedSlide, Measurements) {
	a := AccentSpec{ID: "u", Target: "copy", Mode: "underline", Phrase: "key phrase", AssetPath: "test.png", AssetSHA256: strings.Repeat("a", 64), AlphaBounds: Rect{Width: 1, Height: 1}, StrokeHeightPt: 2}
	c := CanvasSpec{ID: "copy", Kind: "text", Bounds: Rect{X: 100, Y: 80, Width: 300, Height: 100}, Text: "A key phrase", FontFace: "Arial", FontSizePt: 20, Foreground: "#070154", Align: "left", Valign: "top"}
	s := SlideSpec{ID: "s", WidthPt: 960, HeightPt: 540, Canvas: []CanvasSpec{c}, Accents: []AccentSpec{a}}
	p := &PlannedSlide{Canvas: []PlannedCanvas{{CanvasSpec: c, MeasurementID: "m"}}}
	m := Measurements{ByRequestID: map[string]Measurement{"m": {RenderedWidthPt: 180, RenderedHeightPt: 24, PhraseBounds: map[string][]Rect{"u": {{X: 30, Y: 2, Width: 120, Height: 24}}}}}}
	return s, p, m
}
func TestPhraseAccentFollowsTranslationAndPerLineWrapping(t *testing.T) {
	s, p, m := accentTestFixture()
	if err := planAccents(s, p, m); err != nil {
		t.Fatal(err)
	}
	if p.Accents[0].Bounds.X != 130 || p.Accents[0].Bounds.Y != 106 {
		t.Fatalf("bad anchor %+v", p.Accents[0])
	}
	p.Canvas[0].Bounds.X += 40
	p.Canvas[0].Bounds.Y += 20
	p.Accents = nil
	if err := planAccents(s, p, m); err != nil {
		t.Fatal(err)
	}
	if p.Accents[0].Bounds.X != 170 || p.Accents[0].Bounds.Y != 126 {
		t.Fatal("translation not inherited")
	}
	z := m.ByRequestID["m"]
	z.PhraseBounds["u"] = []Rect{{X: 30, Y: 2, Width: 90, Height: 24}, {X: 0, Y: 26, Width: 30, Height: 24}}
	m.ByRequestID["m"] = z
	p.Accents = nil
	if err := planAccents(s, p, m); err == nil || !strings.Contains(err.Error(), "manual_required") {
		t.Fatalf("ambiguous multiline accepted: %v", err)
	}
	s.Accents[0].Multiline = "per_line"
	if err := planAccents(s, p, m); err != nil {
		t.Fatal(err)
	}
	if len(p.Accents) != 2 || p.Accents[1].ID != "u/line-2" {
		t.Fatalf("missing fragment placements: %+v", p.Accents)
	}
}
func TestPhraseManualStaging(t *testing.T) {
	s, p, m := accentTestFixture()
	s.Accents[0].Phrase = "missing"
	s.Accents[0].Staging = &AccentStaging{AssetBounds: Rect{X: 450, Y: 300, Width: 100, Height: 20}, NoteBounds: Rect{X: 450, Y: 335, Width: 350, Height: 70}, Note: "Place this underline beneath the intended key phrase after choosing its occurrence."}
	q := accentNoteProbe(s, s.Accents[0])
	m.ByRequestID[q.ID] = Measurement{RenderedWidthPt: 300, RenderedHeightPt: 40}
	if err := planAccents(s, p, m); err != nil {
		t.Fatal(err)
	}
	if len(p.ManualRequired) != 1 || p.Accents[0].Status != "manual_required" || len(p.Canvas) != 2 {
		t.Fatalf("staging must carry visible note and unfinished state: %+v", p)
	}
}
func TestRotatedHighlightFrame(t *testing.T) {
	a := AccentSpec{Mode: "highlight", RotationDeg: 1, AlphaBounds: Rect{X: .1, Y: .2, Width: .8, Height: .6}}
	v := Rect{X: 100, Y: 50, Width: 200, Height: 30}
	b, err := accentImageFrame(a, v)
	if err != nil || !validRect(b) {
		t.Fatalf("rotated highlight: %+v %v", b, err)
	}
	a.RotationDeg = 15
	v.Height = 1
	if _, err = accentImageFrame(a, v); err == nil {
		t.Fatal("impossible rotated envelope accepted")
	}
}
