package compose

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func loadDiagramFixture(t *testing.T, name string) Spec {
	t.Helper()
	b, err := os.ReadFile("../../library/diagram-components/" + name)
	if err != nil {
		t.Fatal(err)
	}
	var s Spec
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestWave3ReviewFixtureStructuralPlan(t *testing.T) {
	s := loadDiagramFixture(t, "review.json")
	m := fixtureMeasurements(t, s)
	for id := range m.ByRequestID {
		m.ByRequestID[id] = Measurement{RenderedWidthPt: 20, RenderedHeightPt: 12}
	}
	p, err := Plan(s, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Slides) != 8 {
		t.Fatalf("expected eight review slides, got %d", len(p.Slides))
	}
	if len(p.Slides[0].LayoutPorts) != 12 || len(p.Slides[0].Connections) != 8 {
		t.Fatalf("source path nodes/arrows missing: %+v", p.Slides[0])
	}
}

func TestWave3LongLabelNegativeFixture(t *testing.T) {
	s := loadDiagramFixture(t, "overflow.json")
	m := fixtureMeasurements(t, s)
	id := requestID("process-long-label-overflow", "canvas", "stress/node-2/label")
	m.ByRequestID[id] = Measurement{RenderedWidthPt: 500, RenderedHeightPt: 25}
	if _, err := Plan(s, m); err == nil || !strings.Contains(err.Error(), "width") {
		t.Fatalf("negative label did not overflow: %v", err)
	}
}

func TestWave3PinnedArtTranslationIdentity(t *testing.T) {
	s := loadDiagramFixture(t, "art-translation.json")
	if len(s.Slides) != 2 {
		t.Fatal("expected original and translated art pages")
	}
	a, b := s.Slides[0], s.Slides[1]
	for _, id := range []string{"uhg14-board", "uhg14-preview"} {
		var original, moved *CanvasSpec
		for i := range a.Canvas {
			if a.Canvas[i].ID == id {
				original = &a.Canvas[i]
			}
		}
		for i := range b.Canvas {
			if b.Canvas[i].ID == id {
				moved = &b.Canvas[i]
			}
		}
		if original == nil || moved == nil || original.AssetSHA256 != moved.AssetSHA256 || moved.Bounds.X-original.Bounds.X != 8 || moved.Bounds.Y-original.Bounds.Y != 6 {
			t.Fatalf("art identity/translation changed for %s", id)
		}
	}
}
