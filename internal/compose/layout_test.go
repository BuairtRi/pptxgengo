package compose

import (
	"encoding/json"
	"strings"
	"testing"
)

// These are structural unit fixtures with supplied measurements. Native font
// measurement and rendered-slide approval are separate integration gates.
func layoutFixture() Spec {
	return Spec{Schema: SpecSchema, Slides: []SlideSpec{{ID: "fixture", TitleFontFace: "Arial", TitleFontSizePt: 23, TitleForeground: "#070154", WidthPt: 960, HeightPt: 540, Layouts: []ContainerSpec{{ID: "panel", Bounds: Rect{X: 30, Y: 30, Width: 300, Height: 200}, Padding: Insets{Top: 10, Right: 10, Bottom: 10, Left: 10}, Columns: []TrackSpec{{FixedPt: 130}, {FixedPt: 140}}, ColumnGapPt: 10, Rows: []TrackSpec{{MinPt: 30, MaxPt: 80}, {MinPt: 30, MaxPt: 80}}, RowGapPt: 5, Surface: "#E8EEF8", Cells: []CellSpec{{ID: "a", Row: 0, Column: 0, Padding: Insets{Top: 5, Bottom: 5}, Blocks: []BlockSpec{{ID: "body", Text: "First body", FontFace: "Arial", FontSizePt: 10, Foreground: "#070154"}}}, {ID: "b", Row: 0, Column: 1, Padding: Insets{Top: 5, Bottom: 5}, Blocks: []BlockSpec{{ID: "body", Text: "Second body", FontFace: "Arial", FontSizePt: 10, Foreground: "#070154"}}}, {ID: "c", Row: 1, Column: 0, Blocks: []BlockSpec{{ID: "body", Text: "Third body", FontFace: "Arial", FontSizePt: 10, Foreground: "#070154"}}}}}}}}}
}
func fixtureMeasurements(t *testing.T, s Spec) Measurements {
	t.Helper()
	qs, e := ProbeRequests(s)
	if e != nil {
		t.Fatal(e)
	}
	m := Measurements{ByRequestID: map[string]Measurement{}}
	for _, q := range qs {
		m.ByRequestID[q.ID] = Measurement{RenderedWidthPt: 30, RenderedHeightPt: 12}
	}
	return m
}
func TestLayoutProbeDoesNotMutateSource(t *testing.T) {
	s := layoutFixture()
	before, _ := json.Marshal(s)
	m := fixtureMeasurements(t, s)
	if _, e := Plan(s, m); e != nil {
		t.Fatal(e)
	}
	after, _ := json.Marshal(s)
	if string(before) != string(after) {
		t.Fatal("expansion mutated original spec")
	}
}
func TestLayoutTallestCellAndParentMove(t *testing.T) {
	s := layoutFixture()
	m := fixtureMeasurements(t, s)
	id := requestID("fixture", "canvas", "panel/b/body")
	m.ByRequestID[id] = Measurement{RenderedWidthPt: 30, RenderedHeightPt: 45}
	a, e := ExpandLayouts(s, m)
	if e != nil {
		t.Fatal(e)
	}
	positions := map[string]Rect{}
	for _, c := range a.Slides[0].Canvas {
		positions[c.ID] = c.Bounds
	}
	if positions["panel/c/body"].Y != 100 {
		t.Fatalf("second row must follow tallest cell + padding + gap: %+v", positions)
	}
	s.Slides[0].Layouts[0].Bounds.X += 17
	s.Slides[0].Layouts[0].Bounds.Y += 9
	b, e := ExpandLayouts(s, m)
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range b.Slides[0].Canvas {
		old := positions[c.ID]
		if c.Bounds.X != old.X+17 || c.Bounds.Y != old.Y+9 {
			t.Fatalf("child did not follow parent: %s", c.ID)
		}
	}
}
func TestLayoutAllCapacityFailuresAndCycle(t *testing.T) {
	s := layoutFixture()
	m := fixtureMeasurements(t, s)
	for _, id := range []string{"panel/a/body", "panel/b/body"} {
		m.ByRequestID[requestID("fixture", "canvas", id)] = Measurement{RenderedWidthPt: 200, RenderedHeightPt: 100}
	}
	fs := LayoutFitReport(s, m)
	if len(fs) < 3 {
		t.Fatalf("need both width failures and row overflow, got %+v", fs)
	}
	s = layoutFixture()
	s.Slides[0].Layouts[0].ParentID = "panel"
	if _, e := ProbeRequests(s); e == nil || !strings.Contains(e.Error(), "cycle") {
		t.Fatalf("cycle accepted: %v", e)
	}
}
func TestLayoutNestedPaddingAndContrast(t *testing.T) {
	s := layoutFixture()
	p := &s.Slides[0].Layouts[0]
	p.Cells = nil
	p.Surface = "#070154"
	child := ContainerSpec{ID: "child", ParentID: "panel", Bounds: Rect{X: 0, Y: 0, Width: 100, Height: 50}, Columns: []TrackSpec{{FixedPt: 100}}, Rows: []TrackSpec{{FixedPt: 50}}, Cells: []CellSpec{{ID: "cell", Blocks: []BlockSpec{{ID: "text", Text: "White text", FontFace: "Arial", FontSizePt: 10, Foreground: "#FFFFFF"}}}}}
	s.Slides[0].Layouts = append(s.Slides[0].Layouts, child)
	m := fixtureMeasurements(t, s)
	x, e := ExpandLayouts(s, m)
	if e != nil {
		t.Fatal(e)
	}
	var found bool
	for _, c := range x.Slides[0].Canvas {
		if c.ID == "child/cell/text" {
			found = true
			if c.Bounds.X != 40 || c.Bounds.Y != 40 || c.Background != "" || c.ContrastBackground != "#070154" {
				t.Fatalf("wrong inheritance: %+v", c)
			}
		}
	}
	if !found {
		t.Fatal("missing child")
	}
	s.Slides[0].Layouts[1].Bounds.X = 250
	if _, e := ProbeRequests(s); e == nil {
		t.Fatal("child escaped padded parent")
	}
}

func TestNestedPanelsDoNotHideTextCollision(t *testing.T) {
	s := layoutFixture()
	child := ContainerSpec{ID: "child", ParentID: "panel", Bounds: Rect{X: 0, Y: 0, Width: 100, Height: 50}, Columns: []TrackSpec{{FixedPt: 100}}, Rows: []TrackSpec{{FixedPt: 50}}, Cells: []CellSpec{{ID: "cell", Blocks: []BlockSpec{{ID: "text", Text: "Collision", FontFace: "Arial", FontSizePt: 10, Foreground: "#070154"}}}}}
	s.Slides[0].Layouts = append(s.Slides[0].Layouts, child)
	m := fixtureMeasurements(t, s)
	if _, e := Plan(s, m); e == nil || !strings.Contains(e.Error(), "overlap") {
		t.Fatalf("nested text collision accepted: %v", e)
	}
}

func TestLayoutRejectsChildBehindSurface(t *testing.T) {
	s := layoutFixture()
	s.Slides[0].Layouts[0].Cells[0].Blocks[0].Layer = -20
	if _, e := ProbeRequests(s); e == nil {
		t.Fatal("negative child layer must not cover text with its own parent surface")
	}
}
func TestOpaqueFillOwnsContrast(t *testing.T) {
	s := layoutFixture()
	s.Slides[0].Layouts = nil
	s.Slides[0].Canvas = []CanvasSpec{{ID: "bad", Kind: "text", Bounds: Rect{X: 30, Y: 30, Width: 200, Height: 40}, Text: "Invisible", FontFace: "Arial", FontSizePt: 12, Foreground: "#FFFFFF", Background: "#FFFFFF", ContrastBackground: "#070154", Align: "left", Valign: "top"}}
	if _, e := ProbeRequests(s); e == nil {
		t.Fatal("inherited contrast must not override an actual opaque white fill")
	}
}
