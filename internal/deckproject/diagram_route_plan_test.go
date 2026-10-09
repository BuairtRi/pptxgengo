package deckproject

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestOrthogonalRouteObstacleAndQuadrants(t *testing.T) {
	bounds := wmdesign.Rect{X: 0, Y: 0, W: 400, H: 240}
	blocker := wmdesign.Rect{X: 140, Y: 60, W: 100, H: 100}
	cases := [][2][2]float64{{{20, 100}, {380, 100}}, {{380, 100}, {20, 100}}, {{20, 20}, {380, 220}}, {{380, 220}, {20, 20}}, {{20, 220}, {380, 20}}, {{380, 20}, {20, 220}}, {{100, 20}, {100, 220}}, {{20, 20}, {20, 20}}}
	for _, pair := range cases {
		p, e := orthogonalRoute(pair[0], pair[1], bounds, []wmdesign.Rect{blocker})
		if e != nil {
			t.Fatal(pair, e)
		}
		if p[0] != pair[0] || p[len(p)-1] != pair[1] {
			t.Fatal("endpoints lost", p)
		}
		for i, xy := range p {
			if !routeRectContains(bounds, xy) {
				t.Fatal("escaped", p)
			}
			if i > 0 && (!routeSegmentClear(p[i-1], xy, []wmdesign.Rect{blocker}) || p[i-1][0] != xy[0] && p[i-1][1] != xy[1]) {
				t.Fatal("invalid/colliding path", p)
			}
		}
		again, e := orthogonalRoute(pair[0], pair[1], bounds, []wmdesign.Rect{blocker})
		if e != nil || !reflect.DeepEqual(p, again) {
			t.Fatal("non deterministic", p, again)
		}
	}
}
func TestOrthogonalRouteBlockedAndBudget(t *testing.T) {
	b := wmdesign.Rect{W: 200, H: 100}
	if _, e := orthogonalRoute([2]float64{10, 50}, [2]float64{190, 50}, b, []wmdesign.Rect{{X: 90, Y: -1, W: 20, H: 102}}); e == nil {
		t.Fatal("blocking corridor accepted")
	}
	if _, e := orthogonalRoute([2]float64{10, 50}, [2]float64{190, 50}, b, []wmdesign.Rect{{X: 0, Y: 0, W: 20, H: 100}}); e == nil {
		t.Fatal("blocked origin accepted")
	}
	if _, e := orthogonalRoute([2]float64{-1, 50}, [2]float64{190, 50}, b, nil); e == nil {
		t.Fatal("outside allocation accepted")
	}
	if _, e := orthogonalRoute([2]float64{10, 50}, [2]float64{190, 50}, b, make([]wmdesign.Rect, 257)); e == nil {
		t.Fatal("budget not enforced")
	}
}
func TestDiagramRouteStrictPatch(t *testing.T) {
	p := DiagramRoutePatch{Schema: DiagramRoutePatchSchema, ExpectedSourceSHA256: strings.Repeat("a", 64), Actor: "test", Reason: "Avoid a blocker", ID: "edge"}
	if _, e := DecodeDiagramRoutePatch(canonical(p), "patch"); e != nil {
		t.Fatal(e)
	}
	for _, raw := range []string{strings.Replace(string(canonical(p)), `"schema":`, `"unknown":true,"schema":`, 1), strings.Replace(string(canonical(p)), `"reason":"Avoid a blocker"`, `"reason":""`, 1), string(canonical(p)) + "\n---\n{}"} {
		if _, e := DecodeDiagramRoutePatch([]byte(raw), "patch"); e == nil {
			t.Fatal("invalid patch accepted", raw)
		}
	}
	n := math.NaN()
	p.Clearance = &n
	if _, e := DecodeDiagramRoutePatch(canonical(p), "patch"); e == nil {
		t.Fatal("nonfinite clearance accepted")
	}
}
func TestDiagramRoutePreviewApplyAndNativePolyline(t *testing.T) {
	p, _ := geometryFixture(t)
	_, source, e := diagramSlide(p, "architecture-slide")
	if e != nil {
		t.Fatal(e)
	}
	// Preserve the specimen's chrome while explicitly compose three native blocks.
	r := wmdesign.Rect{X: 20, Y: 80, W: 130, H: 60}
	r2 := wmdesign.Rect{X: 330, Y: 80, W: 130, H: 60}
	r3 := wmdesign.Rect{X: 185, Y: 65, W: 100, H: 90}
	block := func(id, text string, r wmdesign.Rect) Node {
		return Node{ID: id, Kind: "component", Placement: &Placement{Zone: "body", Rect: &r}, Definition: &Reference{Scope: "shared", ID: "wmds/component/editable-block"}, Arguments: map[string]any{"text": text, "surface": "subtle", "style": "body"}}
	}
	source.Nodes = []Node{block("start", "Input", r), block("finish", "Output", r2), block("blocker", "Service", r3)}
	if _, e = CompositionCandidate(p, "architecture-slide", "routing-fixture", "test", "Explicit routing fixture", source, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	before := p.SourceHash()
	raw, _ := os.ReadFile(p.SourcePath)
	patch := DiagramRoutePatch{Schema: DiagramRoutePatchSchema, ExpectedSourceSHA256: before, Actor: "test", Reason: "Clear the service block", ID: "edge", From: &DiagramRouteEndpoint{"start", "right"}, To: &DiagramRouteEndpoint{"finish", "left"}}
	out, e := PlanDiagramRoute(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if out.Status != "proposed" || out.Candidate == nil || len(out.Points) < 4 {
		t.Fatal("no detour", out)
	}
	unchanged, _ := os.ReadFile(p.SourcePath)
	if string(unchanged) != string(raw) {
		t.Fatal("preview wrote source")
	}
	for i := 1; i < len(out.Points); i++ {
		if routeSegmentRectDistance(out.Points[i-1], out.Points[i], wmdesign.Rect{X: out.Obstacles[0].Bounds.X, Y: out.Obstacles[0].Bounds.Y, W: out.Obstacles[0].Bounds.W, H: out.Obstacles[0].Bounds.H}) < 6-routingTolerance {
			t.Fatal("insufficient clearance", out.Points)
		}
	}
	applied, e := PlanDiagramRoute(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if applied.Status != "applied" {
		t.Fatal(applied)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectDiagram(p, "architecture-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(inspect.Connections) != 1 || inspect.Connections[0].Route != "polyline" || len(inspect.Connections[0].Points) != len(out.Points) {
		t.Fatal("native route lost", inspect.Connections)
	}
	for i, xy := range inspect.Connections[0].Points {
		if routePointDistance(xy, out.Points[i]) > .02 {
			t.Fatal("native waypoint moved", xy, out.Points[i])
		}
	}
	if _, e = PlanDiagramRoute(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale hash accepted")
	}
}
func TestPresetRouteReadWriteAndRefusal(t *testing.T) {
	route := &wmdesign.NativeConnectorRoute{Preset: "bentConnector5", Adjustment: 30000, Adjustment2: 40000, Adjustment3: 70000}
	data := `<p:cxnSp xmlns:p="` + lineagePML + `"><p:spPr>` + connectorRouteXML(route) + `</p:spPr></p:cxnSp>`
	tree, e := readLineageXML([]byte(data))
	if e != nil {
		t.Fatal(e)
	}
	r, e := readConnectorRoute(tree)
	if e != nil || !reflect.DeepEqual(r, route) {
		t.Fatal("route lost", r, e)
	}
	bad := strings.Replace(data, "val 40000", "*/ h 2 1", 1)
	tree, e = readLineageXML([]byte(bad))
	if e == nil {
		if _, e = readConnectorRoute(tree); e == nil {
			t.Fatal("curve accepted as polyline")
		}
	}
}

func TestDiagramRouteNativePresetBendRoundTrip(t *testing.T) {
	p, b := routedFixture(t, "horizontal")
	inspect, e := InspectDiagram(p, "architecture-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	_ = inspect
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["service-edge"].geometry
	g.Route = &wmdesign.NativeConnectorRoute{Preset: "bentConnector5", Adjustment: 30000, Adjustment2: 40000, Adjustment3: 70000}
	edited := geometryEdited(t, p, b, map[string]NativeGeometry{"service-edge": g}, nil)
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	after, e := InspectDiagram(p, "architecture-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(after.Connections) != 1 || after.Connections[0].Route != "bentConnector5" || len(after.Connections[0].Points) != 6 {
		t.Fatal("manual waypoints lost", after.Connections)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}
func TestPolylineRejectsUnknownGeometryMetadata(t *testing.T) {
	route := &wmdesign.NativeConnectorRoute{Preset: "polyline", Points: [][2]float64{{0, 0}, {.5, 0}, {.5, 1}, {1, 1}}}
	if connectorRouteXML(route) != "" {
		t.Fatal("repair-prone custom route emitted")
	}
	if e := validateNativeGeometry(NativeGeometry{Kind: "cxnSp", W: 100, H: 100, Route: route}); e == nil {
		t.Fatal("legacy custom route adopted")
	}
	base := `<p:cxnSp xmlns:p="` + lineagePML + `" xmlns:a="` + drawingML + `"><p:spPr><a:custGeom><a:avLst/><a:gdLst/><a:ahLst/><a:cxnLst/><a:rect l="l" t="t" r="r" b="b"/><a:pathLst><a:path w="1000" h="1000"><a:moveTo><a:pt x="0" y="0"/></a:moveTo><a:lnTo><a:pt x="500" y="0"/></a:lnTo><a:lnTo><a:pt x="500" y="1000"/></a:lnTo><a:lnTo><a:pt x="1000" y="1000"/></a:lnTo></a:path></a:pathLst></a:custGeom></p:spPr></p:cxnSp>`

	for _, mutation := range [][2]string{{`<a:path w=`, `<a:path stroke="false" w=`}, {`<a:pt x=`, `<a:pt unknown="1" x=`}, {`<a:rect l=`, `<a:rect unknown="1" l=`}, {`<a:custGeom xmlns:a=`, `<a:custGeom unknown="1" xmlns:a=`}, {`<a:pathLst>`, `<a:pathLst unknown="1">`}, {`<a:rect l="l"`, `<a:rect l="formula"`}} {
		tree, e := readLineageXML([]byte(strings.Replace(base, mutation[0], mutation[1], 1)))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = readConnectorRoute(tree); e == nil {
			t.Fatal("unsupported metadata discarded", mutation)
		}
	}
}
