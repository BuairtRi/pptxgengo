package deckproject

import (
	"encoding/xml"
	"math"
	"reflect"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestRoutingSegmentGeometry(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		a, b, c, d [2]float64
	}{
		{"cross", "crossing", [2]float64{0, 5}, [2]float64{10, 5}, [2]float64{5, 0}, [2]float64{5, 10}},
		{"diagonal", "crossing", [2]float64{0, 0}, [2]float64{10, 10}, [2]float64{0, 10}, [2]float64{10, 0}},
		{"touch", "crossing", [2]float64{0, 5}, [2]float64{5, 5}, [2]float64{5, 0}, [2]float64{5, 10}},
		{"reversed overlap", "overlap", [2]float64{10, 5}, [2]float64{0, 5}, [2]float64{5, 5}, [2]float64{15, 5}},
		{"parallel", "", [2]float64{0, 5}, [2]float64{10, 5}, [2]float64{0, 6}, [2]float64{10, 6}},
		{"outside", "", [2]float64{0, 5}, [2]float64{4, 5}, [2]float64{5, 0}, [2]float64{5, 10}},
		{"zero", "", [2]float64{0, 5}, [2]float64{0, 5}, [2]float64{0, 0}, [2]float64{0, 10}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kind, _ := routeSegmentIntersection(tc.a, tc.b, tc.c, tc.d)
			if kind != tc.kind {
				t.Fatalf("want %q, got %q", tc.kind, kind)
			}
		})
	}
	r := wmdesign.Rect{X: 10, Y: 10, W: 10, H: 10}
	for _, tc := range []struct {
		a, b [2]float64
		d    float64
	}{
		{[2]float64{0, 15}, [2]float64{30, 15}, 0},
		{[2]float64{12, 12}, [2]float64{18, 18}, 0},
		{[2]float64{0, 6}, [2]float64{30, 6}, 4},
		{[2]float64{0, 0}, [2]float64{7, 6}, 5},
		{[2]float64{0, 10}, [2]float64{10, 10}, 0},
	} {
		if got := routeSegmentRectDistance(tc.a, tc.b, r); math.Abs(got-tc.d) > 1e-9 {
			t.Fatalf("distance %v to %v = %v, want %v", tc.a, tc.b, got, tc.d)
		}
	}
}

func TestRoutingObstaclesClearanceAndContainmentExclusions(t *testing.T) {
	c := DiagramConnection{Node: "edge", Points: [][2]float64{{10, 20}, {100, 20}}, From: DiagramPort{Node: "a", NativeObject: "a.surface", Site: "right", X: 10, Y: 20}, To: DiagramPort{Node: "b", NativeObject: "b.surface", Site: "left", X: 100, Y: 20}}
	obstacles := []DiagramRoutingObstacle{
		{"a.surface", "block", wmdesign.Rect{X: 0, Y: 10, W: 10, H: 20}, ""},
		{"b.surface", "block", wmdesign.Rect{X: 100, Y: 10, W: 10, H: 20}, ""},
		{"container", "block", wmdesign.Rect{X: 0, Y: 0, W: 120, H: 60}, ""},
		{"middle", "block", wmdesign.Rect{X: 40, Y: 10, W: 10, H: 20}, ""},
		{"label", "text", wmdesign.Rect{X: 60, Y: 24, W: 20, H: 10}, ""},
		{"far", "text", wmdesign.Rect{X: 60, Y: 30, W: 20, H: 10}, ""},
	}
	rules := map[string]wmdesign.DiagramContainment{"a": {Container: "container"}, "b": {Container: "container"}}
	out := inspectDiagramRoutes([]DiagramConnection{c}, obstacles, rules, nil)
	if len(out.Diagnostics) != 2 || out.Diagnostics[0].Kind != "obstacle_intersection" || out.Diagnostics[0].Other != "middle" || out.Diagnostics[1].Kind != "obstacle_clearance" || *out.Diagnostics[1].ClearancePT != 4 {
		t.Fatalf("wrong diagnostics: %+v", out.Diagnostics)
	}
	// A connector is excluded from its explicit containing block even if
	// endpoint memberships are absent. Header text is still diagnosed.
	out = inspectDiagramRoutes([]DiagramConnection{c}, obstacles, map[string]wmdesign.DiagramContainment{"edge": {Container: "container"}}, nil)
	if len(out.Diagnostics) != 2 {
		t.Fatal("connector membership ignored", out.Diagnostics)
	}
}

func TestRoutingCrossingsDeduplicateJoinsAndReportOverlaps(t *testing.T) {
	a := DiagramConnection{Node: "a", Points: [][2]float64{{0, 10}, {10, 10}, {20, 10}}, From: DiagramPort{NativeObject: "shared", Site: "right", X: 0, Y: 10}}
	b := DiagramConnection{Node: "b", Points: [][2]float64{{10, 0}, {10, 10}, {10, 20}}}
	out := inspectDiagramRoutes([]DiagramConnection{a, b}, nil, nil, nil)
	if len(out.Diagnostics) != 1 || out.Diagnostics[0].Kind != "route_crossing" {
		t.Fatal("bend intersection duplicated", out.Diagnostics)
	}
	b.Points = [][2]float64{{0, 10}, {0, 20}}
	b.From = a.From
	out = inspectDiagramRoutes([]DiagramConnection{a, b}, nil, nil, nil)
	if len(out.Diagnostics) != 0 {
		t.Fatal("shared endpoint flagged", out.Diagnostics)
	}
	b.Points = [][2]float64{{0, 10}, {8, 10}, {8, 20}}
	out = inspectDiagramRoutes([]DiagramConnection{a, b}, nil, nil, nil)
	if len(out.Diagnostics) != 2 || out.Diagnostics[0].Kind != "route_overlap" {
		t.Fatal("shared endpoint hid overlapping route", out.Diagnostics)
	}
}

func TestRoutingCatalogPreviewIsAdvisoryAndSourceSafe(t *testing.T) {
	p, _ := geometryFixture(t)
	hash := p.SourceHash()
	out, e := ConnectDiagram(p, "architecture-slide", "across-service", "node06", "right", "node08", "left", "end", "solid", "test", "Inspect an intentional crossing", bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if out.Applied || out.Inspection.Routing == nil || hash != p.SourceHash() {
		t.Fatal("preview changed source or omitted routing")
	}
	found := false
	for _, d := range out.Inspection.Routing.Diagnostics {
		found = found || d.Kind == "obstacle_intersection" && d.Other == "node07.surface"
	}
	if !found {
		t.Fatal("route through middle block not reported", out.Inspection.Routing)
	}
	// A warning never silently reroutes or rejects an authored relationship.
	applied, e := ConnectDiagram(p, "architecture-slide", "across-service", "node06", "right", "node08", "left", "end", "solid", "test", "Accept intentional crossing", bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(out.Inspection.Connections, applied.Inspection.Connections) || !reflect.DeepEqual(out.Inspection.Routing, applied.Inspection.Routing) {
		t.Fatal("apply changed route or diagnostics")
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	dy := 40.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Move the middle obstacle", Operations: []DiagramOperation{{Action: "move", ID: "node07", DY: &dy}}}
	moved, e := PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	for _, d := range moved.Inspection.Routing.Diagnostics {
		if d.Other == "node07.surface" {
			t.Fatal("diagnostic did not follow proposed source position", d)
		}
	}
	// The same allocation change, adopted from native edits, is measured in
	// final coordinates instead of the baseline/source text coordinates.
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["node07"].geometry
	g.Y += 40
	packet := structurePacket(t, p, b, geometryEdited(t, p, b, map[string]NativeGeometry{"node07": g}, nil))
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectDiagram(p, "architecture-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	for _, d := range inspect.Routing.Diagnostics {
		if d.Other == "node07.surface" {
			t.Fatal("diagnostic did not follow adopted native position", d)
		}
	}
}

func TestRoutingNativeLabelWorldAllocation(t *testing.T) {
	text := &xmlNode{Name: xml.Name{Space: lineagePML, Local: "sp"}, Children: []*xmlNode{{Name: xml.Name{Space: lineagePML, Local: "txBody"}, Children: []*xmlNode{{Name: xml.Name{Space: drawingML, Local: "p"}, Children: []*xmlNode{{Name: xml.Name{Space: drawingML, Local: "r"}, Children: []*xmlNode{{Name: xml.Name{Space: drawingML, Local: "t"}, Text: "Container heading"}}}}}}}}}
	objects := map[string]*geometryObject{
		"group": {geometry: NativeGeometry{Kind: "grpSp", X: 100, Y: 50, W: 200, H: 100, Rotation: 90, Child: &wmdesign.Rect{W: 100, H: 100}}},
		"label": {node: text, geometry: NativeGeometry{Kind: "sp", Parent: "group", X: 40, Y: 10, W: 20, H: 20}},
	}
	world, e := geometryWorld(objects)
	if e != nil {
		t.Fatal(e)
	}
	r := world["label"]
	out := DiagramInspection{Connections: []DiagramConnection{{Node: "edge", Points: [][2]float64{{r.X - 10, r.Y + r.H/2}, {r.X + r.W + 10, r.Y + r.H/2}}}}}
	for name, object := range objects {
		out.FinalNative = append(out.FinalNative, NativeGeometryObservation{name, object.geometry, world[name]})
	}
	addDiagramRoutingDiagnostics(&out, objects, nil)
	if len(out.Routing.Obstacles) != 1 || out.Routing.Obstacles[0].Bounds != r || len(out.Routing.Diagnostics) != 1 || out.Routing.Diagnostics[0].Other != "label" || out.Routing.Diagnostics[0].Kind != "obstacle_intersection" {
		t.Fatal("label world transform ignored", out.Routing)
	}
	// The surface of an explicitly declared containing source block must be
	// excluded, but its separate native heading must remain an obstacle.
	objects["surface"] = &geometryObject{geometry: NativeGeometry{Kind: "sp", Parent: "group", W: 100, H: 100}}
	world, e = geometryWorld(objects)
	if e != nil {
		t.Fatal(e)
	}
	out.Ports = []DiagramPort{{Node: "group", NativeObject: "surface"}}
	out.FinalNative = nil
	out.Warnings = nil
	for name, object := range objects {
		out.FinalNative = append(out.FinalNative, NativeGeometryObservation{name, object.geometry, world[name]})
	}
	addDiagramRoutingDiagnostics(&out, objects, map[string]wmdesign.DiagramContainment{"edge": {Container: "group"}})
	if len(out.Routing.Diagnostics) != 1 || out.Routing.Diagnostics[0].Other != "label" {
		t.Fatal("containing block hid heading or was counted as obstacle", out.Routing.Diagnostics)
	}
}
