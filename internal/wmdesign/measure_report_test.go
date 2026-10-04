package wmdesign

import "testing"

func TestMeasurementFindsTextCollisionsAndPanelSpills(t *testing.T) {
	text := func(id string, x, y float64) TextRecord {
		return TextRecord{ID: id, Rect: Rect{x, y, 100, 24}, Layout: TextLayout{Style: Style{Leading: 12}, Lines: []TextLine{{Text: "First", Advance: 80}, {Text: "Second", Advance: 70}}, EstimatedOccupiedHeight: 23, AllocationHeight: 24}}
	}
	report := Report{Engine: CandidateEngine, Slides: []SlideReport{{ID: "one", Page: 1, Texts: []TextRecord{text("list", 10, 10), text("label", 10, 30), text("elsewhere", 200, 10)}, Shapes: []ShapeRecord{{ID: "card.container", Rect: Rect{0, 0, 120, 30}}}}}}
	result := MeasureReport(report)
	if result.CollisionCount != 1 || result.Slides[0].Collisions[0].First != "list" || result.PanelSpillCount != 1 || result.Slides[0].PanelSpills[0].TextID != "list" {
		t.Fatalf("lost review findings: %+v", result)
	}
	// A rotated label must not pretend its unrotated rectangle is ink geometry.
	report.Slides[0].Texts[0].Rotation = 90
	result = MeasureReport(report)
	if result.CollisionCount != 0 || result.PanelSpillCount != 0 {
		t.Fatalf("rotated geometry was not excluded: %+v", result)
	}
}

func TestMeasurementRespectsAlignmentAndBlankLines(t *testing.T) {
	text := TextRecord{Rect: Rect{10, 10, 100, 48}, Align: "right", VerticalAlign: "middle", Layout: TextLayout{Style: Style{Leading: 12}, Lines: []TextLine{{}, {Text: "Visible", Advance: 30}}, OccupiedTop: 12, EstimatedOccupiedHeight: 11, AllocationHeight: 24}}
	lines := estimatedTextLines(text)
	if len(lines) != 1 || lines[0].X != 80 || lines[0].Y != 34 || lines[0].H != 11 {
		t.Fatalf("wrong occupied estimate: %+v", lines)
	}
}
