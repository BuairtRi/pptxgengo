package compose

import (
	"strings"
	"testing"
)

func TestResolvedLayoutPortsAndDirectFlow(t *testing.T) {
	s := layoutFixture()
	s.Slides[0].Connections = []ConnectionSpec{{ID: "a-to-b", From: "panel/a", To: "panel/b", Relationship: "flow", FromAnchor: "right", ToAnchor: "left", Color: "#070154", WidthPt: 1, ClearancePt: 2, GapMarginPt: 2, ArrowHeightPt: 8}}
	m := fixtureMeasurements(t, s)
	p, err := Plan(s, m)
	if err != nil {
		t.Fatal(err)
	}
	slide := p.Slides[0]
	ports := map[string]ResolvedLayoutPort{}
	for _, port := range slide.LayoutPorts {
		ports[port.ID] = port
	}
	if ports["panel"].Bounds.X != 30 || ports["panel/a"].Bounds.X != 40 || ports["panel/b"].Bounds.X != 180 {
		t.Fatalf("wrong ports: %+v", ports)
	}
	if len(slide.Connections) != 1 || slide.Connections[0].Strategy != "direct_right_arrow" || slide.Connections[0].EndpointBehavior != "editable_segment" {
		t.Fatalf("wrong relation: %+v", slide.Connections)
	}
	var arrow *PlannedCanvas
	for i := range slide.Canvas {
		if slide.Canvas[i].ID == "a-to-b/arrow" {
			arrow = &slide.Canvas[i]
		}
	}
	if arrow == nil || arrow.Preset != "rightArrow" || arrow.Bounds.X != 172.5 || arrow.Bounds.Width != 5 {
		t.Fatalf("wrong editable arrow: %+v", arrow)
	}
	s.Slides[0].Layouts[0].Bounds.X += 17
	s.Slides[0].Layouts[0].Bounds.Y += 9
	p2, err := Plan(s, m)
	if err != nil {
		t.Fatal(err)
	}
	for _, port := range p2.Slides[0].LayoutPorts {
		old := ports[port.ID]
		if port.Bounds.X != old.Bounds.X+17 || port.Bounds.Y != old.Bounds.Y+9 {
			t.Fatalf("port did not move with panel: %+v", port)
		}
	}
	for _, c := range p2.Slides[0].Canvas {
		if c.ID == "a-to-b/arrow" && (c.Bounds.X != arrow.Bounds.X+17 || c.Bounds.Y != arrow.Bounds.Y+9) {
			t.Fatalf("arrow did not move with panel: %+v", c)
		}
	}
}

func TestFlowGapAndUnknownPortFailClearly(t *testing.T) {
	s := layoutFixture()
	c := ConnectionSpec{ID: "flow", From: "panel/a", To: "panel/b", Relationship: "flow", FromAnchor: "right", ToAnchor: "left", Color: "#070154", WidthPt: 1, ClearancePt: 2, GapMarginPt: 6}
	s.Slides[0].Connections = []ConnectionSpec{c}
	m := fixtureMeasurements(t, s)
	if _, err := Plan(s, m); err == nil || !strings.Contains(err.Error(), "flow gap") {
		t.Fatalf("gap failure missing: %v", err)
	}
	s.Slides[0].Connections[0].To = "panel/missing"
	if _, err := ProbeRequests(s); err == nil || !strings.Contains(err.Error(), "unknown target") {
		t.Fatalf("unknown port accepted: %v", err)
	}
}

func TestDirectFlowRejectsPriorLineInkAndLowContrast(t *testing.T) {
	c := ConnectionSpec{ID: "flow", Relationship: "flow", Color: "#070154", WidthPt: 1, ClearancePt: 2, ArrowHeightPt: 8}
	p := PlannedSlide{WidthPt: 300, HeightPt: 200, Connections: []PlannedConnection{{ID: "prior", Relationship: "dependency", WidthPt: 1, Points: []Point{{X: 120, Y: 50}, {X: 180, Y: 50}}}}}
	if err := planDirectFlow(c, Point{X: 100, Y: 50}, Point{X: 200, Y: 50}, &p); err == nil || !strings.Contains(err.Error(), "overlaps connection") {
		t.Fatalf("prior line ink accepted: %v", err)
	}
	p.Connections = nil
	p.Canvas = []PlannedCanvas{{CanvasSpec: CanvasSpec{ID: "dark-surface", Kind: "surface", Bounds: Rect{X: 100, Y: 40, Width: 100, Height: 20}, Background: "#070154"}}}
	if err := planDirectFlow(c, Point{X: 100, Y: 50}, Point{X: 200, Y: 50}, &p); err == nil || !strings.Contains(err.Error(), "contrast") {
		t.Fatalf("low contrast surface accepted: %v", err)
	}
	p.Canvas = []PlannedCanvas{{CanvasSpec: CanvasSpec{ID: "flow/arrow", Kind: "shape", Preset: "rightArrow", Bounds: Rect{X: 20, Y: 20, Width: 10, Height: 10}, Background: "#070154"}}}
	if err := planDirectFlow(c, Point{X: 100, Y: 50}, Point{X: 200, Y: 50}, &p); err == nil || !strings.Contains(err.Error(), "ID collides") {
		t.Fatalf("generated ID collision accepted: %v", err)
	}
}

func TestLaterLineRoutesAroundFullFlowArrow(t *testing.T) {
	s := layoutFixture()
	s.Slides[0].Connections = []ConnectionSpec{{ID: "flow", From: "panel/a", To: "panel/b", Relationship: "flow", FromAnchor: "right", ToAnchor: "left", Color: "#070154", WidthPt: 1, ClearancePt: 2, ArrowHeightPt: 8}, {ID: "dependency", From: "panel/a", To: "panel/b", Relationship: "dependency", FromAnchor: "right", ToAnchor: "left", Color: "#070154", WidthPt: 1, ClearancePt: 2}}
	_, err := Plan(s, fixtureMeasurements(t, s))
	if err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("later line crossed flow arrow: %v", err)
	}
}

func TestLargeBoldSourcePinkUsesThreeToOneContrast(t *testing.T) {
	if _, err := foregroundAtSize("#F900D3", "#FFFFFF", 16, true); err != nil {
		t.Fatalf("large bold source color rejected: %v", err)
	}
	if _, err := foregroundAtSize("#F900D3", "#FFFFFF", 12, false); err == nil {
		t.Fatal("small source color accepted below 4.5:1")
	}
}

func TestRelationshipContracts(t *testing.T) {
	for _, rel := range []string{"reporting", "dependency", "advisory", "annotation", "flow"} {
		s := layoutFixture()
		s.Slides[0].Connections = []ConnectionSpec{{ID: "relation", From: "panel/a", To: "panel/b", Relationship: rel, FromAnchor: "right", ToAnchor: "left", Color: "#070154", WidthPt: 1, ClearancePt: 2}}
		if _, err := ProbeRequests(s); err != nil {
			t.Fatalf("%s rejected: %v", rel, err)
		}
	}
}

func TestDependencyRoutesBetweenCellPorts(t *testing.T) {
	s := layoutFixture()
	s.Slides[0].Connections = []ConnectionSpec{{ID: "depends", From: "panel/a", To: "panel/b", Relationship: "dependency", FromAnchor: "right", ToAnchor: "left", Color: "#070154", WidthPt: 1, ClearancePt: 2, PreferredDirection: "horizontal", BendPenaltyPt: 18}}
	p, err := Plan(s, fixtureMeasurements(t, s))
	if err != nil {
		t.Fatal(err)
	}
	c := p.Slides[0].Connections[0]
	if c.Strategy != "clear_candidate" || len(c.Points) != 2 || c.Points[0].X != 170 || c.Points[1].X != 180 || c.EndpointBehavior != "editable_segment" {
		t.Fatalf("wrong dependency route: %+v", c)
	}
}

func TestProcessPathExpandsMeasuredNodes(t *testing.T) {
	s := Spec{Schema: SpecSchema, Slides: []SlideSpec{{ID: "path-slide", WidthPt: 960, HeightPt: 540, TitleFontFace: "Arial", TitleFontSizePt: 23, TitleForeground: "#070154", Paths: []ProcessPathSpec{{ID: "path", Bounds: Rect{X: 100, Y: 100, Width: 520, Height: 60}, Labels: []string{"Start", "Review", "Finish"}, GapPt: 20, NodePadding: Insets{Left: 4, Right: 4}, NodeBackground: "#E8EEF8", FontFace: "Arial", FontSizePt: 10, Foreground: "#070154", ArrowColor: "#070154", ArrowHeightPt: 12, ArrowMarginPt: 2}}}}}
	m := fixtureMeasurements(t, s)
	p, err := Plan(s, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Slides[0].LayoutPorts) != 4 || len(p.Slides[0].Connections) != 2 {
		t.Fatalf("wrong path expansion: %+v", p.Slides[0])
	}
	if p.Slides[0].Connections[0].Strategy != "direct_right_arrow" || p.Slides[0].Connections[1].Strategy != "direct_right_arrow" {
		t.Fatalf("path arrows absent: %+v", p.Slides[0].Connections)
	}
	s.Slides[0].Paths[0].Labels = append(s.Slides[0].Paths[0].Labels, "Publish")
	m = fixtureMeasurements(t, s)
	p, err = Plan(s, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Slides[0].LayoutPorts) != 5 || len(p.Slides[0].Connections) != 3 {
		t.Fatalf("four-node path not expanded")
	}
}
