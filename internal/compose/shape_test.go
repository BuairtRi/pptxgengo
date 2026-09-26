package compose

import "testing"

func shapeSpec(c CanvasSpec) Spec {
	return Spec{Schema: SpecSchema, Slides: []SlideSpec{{ID: "s", WidthPt: 960, HeightPt: 540, TitleFontFace: "Arial", TitleFontSizePt: 24, Canvas: []CanvasSpec{c}}}}
}

func TestBoundedRoadmapShapeValidation(t *testing.T) {
	base := CanvasSpec{ID: "tail", Kind: "shape", Bounds: Rect{X: 20, Y: 40, Width: 120, Height: 39.6}, Preset: "homePlate", Adjustments: map[string]int{"adj": 39542}, Pattern: &PatternSpec{Preset: "wdUpDiag", Foreground: "#CED7E6", Background: "#FFFFFF"}}
	if _, err := ProbeRequests(shapeSpec(base)); err != nil {
		t.Fatal(err)
	}
	tests := []CanvasSpec{base, base, base, base, base, base, base}
	tests[0].Preset = "freeform"
	tests[1].Adjustments = map[string]int{"adj2": 2}
	tests[2].Adjustments = map[string]int{"adj": -1}
	tests[3].Pattern = &PatternSpec{Preset: "zigzag", Foreground: "#CED7E6", Background: "#FFFFFF"}
	tests[4].Background = "#CED7E6"
	tests[5].Text = "silently lost"
	tests[6].Pattern = &PatternSpec{Preset: "wdUpDiag", Foreground: "surface.light", Background: "#FFFFFF"}
	for i, c := range tests {
		if _, err := ProbeRequests(shapeSpec(c)); err == nil {
			t.Fatalf("invalid shape %d accepted: %+v", i, c)
		}
	}
}

func TestRoadmapShapePlanResolvesPatternWithoutAliasing(t *testing.T) {
	c := CanvasSpec{ID: "tail", Kind: "shape", Bounds: Rect{X: 20, Y: 40, Width: 120, Height: 39.6}, Preset: "homePlate", Adjustments: map[string]int{"adj": 39542}, Pattern: &PatternSpec{Preset: "wdUpDiag", Foreground: "#CED7E6", Background: "#FFFFFF"}}
	s := shapeSpec(c)
	p, err := Plan(s, Measurements{ByRequestID: map[string]Measurement{}})
	if err != nil {
		t.Fatal(err)
	}
	got := p.Slides[0].Canvas[0]
	if got.Pattern == nil || got.Pattern.Foreground != "#CED7E6" || got.Adjustments["adj"] != 39542 || got.MeasurementID != "" {
		t.Fatalf("shape plan lost contract: %+v", got)
	}
	got.Pattern.Foreground = "#000000"
	got.Adjustments["adj"] = 1
	if s.Slides[0].Canvas[0].Pattern.Foreground != "#CED7E6" {
		t.Fatal("planned pattern aliases source")
	}
	if s.Slides[0].Canvas[0].Adjustments["adj"] != 39542 {
		t.Fatal("planned adjustments alias source")
	}
}
