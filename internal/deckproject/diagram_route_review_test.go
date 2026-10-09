package deckproject

import (
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"testing"
)

func TestDiagramRouteDeclaredContainingBlockRetainsSeparateHeadingObstacle(t *testing.T) {
	p, _ := geometryFixture(t)
	_, local, e := diagramSlide(p, "architecture-slide")
	if e != nil {
		t.Fatal(e)
	}
	block := func(id, text string, r wmdesign.Rect) Node {
		return Node{ID: id, Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/editable-block"}, Placement: &Placement{Zone: "body", Rect: &r}, Arguments: map[string]any{"text": text, "surface": "subtle", "style": "body"}}
	}
	envelope := wmdesign.Rect{X: 10, Y: 20, W: 500, H: 230}
	start := wmdesign.Rect{X: 35, Y: 85, W: 110, H: 50}
	finish := wmdesign.Rect{X: 360, Y: 85, W: 110, H: 50}
	heading := wmdesign.Rect{X: 205, Y: 85, W: 105, H: 50}
	enclosing := block("envelope", "Enclosure", envelope)
	local.Nodes = []Node{enclosing, block("start", "Input", start), block("finish", "Output", finish), {ID: "separate-heading", Kind: "text", Style: "small", Text: "Separate heading", Placement: &Placement{Zone: "body", Rect: &heading}}}
	edgeRect := wmdesign.Rect{X: 0, Y: 0, W: 800, H: 300}
	local.Nodes = append(local.Nodes, Node{ID: "edge", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/attached-connector"}, Placement: &Placement{Zone: "body", Rect: &edgeRect}, Arguments: map[string]any{"from": map[string]any{"node": "start", "site": "right"}, "to": map[string]any{"node": "finish", "site": "left"}, "head": "end", "style": "solid", "route": "straight"}})
	for key, z := range local.Zones {
		if z.Role != "slide-title" && z.Role != "eyebrow" && z.Role != "source" && z.Role != "nav" {
			delete(local.Zones, key)
			delete(p.Document.Slides[0].Values, key)
		}
	}
	p.Document.LocalTemplates[p.Document.Slides[0].Template.ID] = local
	p.Document.Slides[0].DiagramContainment = map[string]wmdesign.DiagramContainment{"start": {Container: "envelope"}, "finish": {Container: "envelope"}, "edge": {Container: "envelope"}}
	ref := p.Document.Slides[0].Template
	p.Document.Slides[0].NativeGeometryTemplate = &ref
	raw, e := json.Marshal(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	patch := DiagramRoutePatch{Schema: DiagramRoutePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "review", Reason: "Honor explicit enclosure while avoiding its separate heading", ID: "edge"}
	plan, e := PlanDiagramRoute(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if plan.Status != "proposed" || len(plan.Points) < 4 {
		t.Fatalf("authorized enclosure blocked or heading not detoured: %+v", plan)
	}
	excluded := false
	headingObstacle := false
	for _, ob := range plan.Obstacles {
		if ob.Name == "separate-heading" {
			headingObstacle = true
			for i := 1; i < len(plan.Points); i++ {
				if routeSegmentRectDistance(plan.Points[i-1], plan.Points[i], ob.Bounds) < 6-routingTolerance {
					t.Fatal("heading crossed", plan.Points)
				}
			}
		}
	}
	for name := range plan.Excluded {
		if name == "envelope" {
			excluded = true
		}
	}
	if !excluded || !headingObstacle {
		t.Fatal("container/heading policy missing", plan.Excluded, plan.Obstacles)
	}
	if p.SourceHash() != patch.ExpectedSourceSHA256 {
		t.Fatal("preview changed authored source")
	}
}
