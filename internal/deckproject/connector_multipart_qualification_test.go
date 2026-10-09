package deckproject

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func multipartQualificationFixture(t *testing.T) (*Project, *TextBaseline) {
	t.Helper()
	p, _ := geometryFixture(t)
	_, source, e := diagramSlide(p, "architecture-slide")
	if e != nil {
		t.Fatal(e)
	}
	block := func(id, label string, rect wmdesign.Rect) Node {
		return Node{ID: id, Kind: "component", Placement: &Placement{Zone: "body", Rect: &rect}, Definition: &Reference{Scope: "shared", ID: "wmds/component/editable-block"}, Arguments: map[string]any{"text": label, "style": "body", "surface": "subtle"}}
	}
	source.Nodes = []Node{block("start", "Input", wmdesign.Rect{X: 20, Y: 70, W: 100, H: 50}), block("finish", "Output", wmdesign.Rect{X: 420, Y: 70, W: 100, H: 50}), block("blocker", "Service", wmdesign.Rect{X: 180, Y: 70, W: 160, H: 50})}
	for _, fixture := range []struct {
		id     string
		points [][2]float64
	}{{"detour-u", [][2]float64{{140, 95}, {140, 35}, {400, 35}, {400, 95}}}, {"many-bends", [][2]float64{{140, 95}, {140, 30}, {160, 30}, {160, 10}, {360, 10}, {360, 30}, {400, 30}, {400, 95}}}} {
		rect := wmdesign.Rect{W: 800, H: 280}
		labelY := 30.0
		if fixture.id == "many-bends" {
			labelY = 70
		}
		source.Nodes = append(source.Nodes, Node{ID: fixture.id, Kind: "component", Placement: &Placement{Zone: "body", Rect: &rect}, Definition: &Reference{Scope: "shared", ID: "wmds/component/attached-connector"}, Arguments: map[string]any{"from": map[string]any{"node": "start", "site": "right"}, "to": map[string]any{"node": "finish", "site": "left"}, "route": "polyline", "waypoints": fixture.points, "head": "end", "style": "solid", "label": fixture.id, "label_position": []float64{550, labelY}, "label_width": 160.0}})
	}
	if _, e = CompositionCandidate(p, "architecture-slide", "multipart-fixture", "Qualification", "Native route qualification with service blocker, U detour and eight interior waypoints", source, bundle(t), wmdesign.CandidateEngine, true); e != nil {
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
	if len(inspect.Connections) != 2 {
		t.Fatalf("logical connections lost %+v", inspect.Connections)
	}
	for _, connection := range inspect.Connections {
		if connection.Representation != "attached-native-segments" || connection.Route != "polyline" || len(connection.Points) < 6 || len(connection.NativeParts) != 2*len(connection.Points)-3 {
			t.Fatalf("multipart topology lost %+v", connection)
		}
		for i := 1; i < len(connection.Points); i++ {
			if !routeSegmentClear(connection.Points[i-1], connection.Points[i], []wmdesign.Rect{{X: 180 + inspect.Frame.Body.X, Y: 70 + inspect.Frame.Body.Y, W: 160, H: 50}}) {
				t.Fatal("source detour crossed blocker")
			}
		}
	}
	out, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	pkg, e := openLineagePackage(b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	raw, e := pkg.read("ppt/slides/slide1.xml")
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(raw), "<a:custGeom>") {
		t.Fatal("repair-prone custom connectors remain")
	}
	_ = out
	return p, b
}

func TestMultipartNativeConnectorSourceInspectRebuildAndFixture(t *testing.T) {
	p, b := multipartQualificationFixture(t)
	var e error
	root := os.Getenv("PPTXGENGO_COMPOSITION_FIXTURES")
	if root != "" {
		target := filepath.Join(root, "native-multipart-routing")
		if e = os.MkdirAll(target, 0700); e != nil {
			t.Fatal(e)
		}
		for name, data := range map[string][]byte{"deck.yaml": p.Raw, "baseline.pptx": b.files["deck.pptx"], "fixture-build.json": canonical(map[string]string{"bundle": bundle(t), "engine": wmdesign.CandidateEngine})} {
			if e = os.WriteFile(filepath.Join(target, name), data, 0600); e != nil {
				t.Fatal(e)
			}
		}
	}
}

// Movement persists as native geometry; it does not rewrite authored routing facts.
func TestMultipartNativeGuideMovementAdoptsAndRebuilds(t *testing.T) {
	p, b := multipartQualificationFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	name := "detour-u.waypoint-002"
	guide := objects[name].geometry
	guide.Y -= 5
	changes := map[string]NativeGeometry{name: guide}
	// Move the adjacent segment endpoints exactly with their attached guide.
	for _, step := range []struct {
		name string
		end  int
	}{{"detour-u.segment-002", 1}, {"detour-u.segment-003", 0}} {
		old := objects[step.name].geometry
		pts := nativeConnectorPoints(old)
		for i := range pts {
			var e error
			pts[i], e = nativeObjectPoint(objects, step.name, pts[i][0], pts[i][1])
			if e != nil {
				t.Fatal(e)
			}
		}
		pts[step.end][1] -= 5
		old.X, old.Y = math.Min(pts[0][0], pts[1][0]), math.Min(pts[0][1], pts[1][1])
		old.W, old.H = math.Abs(pts[1][0]-pts[0][0]), math.Abs(pts[1][1]-pts[0][1])
		old.FlipH, old.FlipV = pts[0][0] > pts[1][0], pts[0][1] > pts[1][1]
		changes[step.name] = old
	}
	edited := geometryEdited(t, p, b, changes, nil)
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(packet.Report.ManualReview) != 0 {
		t.Fatalf("multipart movement left manual issues: %+v", packet.Report.ManualReview)
	}
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Document.Slides[0].NativeGeometry) != 3 {
		t.Fatal("expected three persistent native overrides")
	}
	source, e := InspectDiagram(p, "architecture-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	var route DiagramConnection
	for _, c := range source.Connections {
		if c.Node == "detour-u" {
			route = c
		}
	}
	if len(route.Points) != 6 || math.Abs(route.Points[2][1]-guide.Y) > .001 {
		t.Fatal("inspection ignored native movement", route)
	}
	if _, e = PlanDiagramRoute(p, "architecture-slide", DiagramRoutePatch{Schema: DiagramRoutePatchSchema, ExpectedSourceSHA256: p.SourceHash(), ID: "detour-u", Actor: "review", Reason: "Reroute"}, bundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("reroute silently discarded native edits")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	fresh, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	rebuilt, _ := geometryFromPackage(t, fresh.files["deck.pptx"])
	native, _ := geometryFromPackage(t, edited)
	for name := range changes {
		if !bytes.Equal(canonical(native[name].geometry), canonical(rebuilt[name].geometry)) {
			t.Fatal("native movement did not persist", name)
		}
	}
}

func TestMultipartNativeContainmentRequiresCompleteClosedNamespace(t *testing.T) {
	valid := map[string]*geometryObject{}
	for i := 1; i <= 3; i++ {
		valid[fmt.Sprintf("edge.segment-%03d", i)] = &geometryObject{geometry: NativeGeometry{Kind: "cxnSp"}}
		if i < 3 {
			valid[fmt.Sprintf("edge.waypoint-%03d", i)] = &geometryObject{geometry: NativeGeometry{Kind: "sp"}}
		}
	}
	if parts, e := multipartNativeRouteNames(valid, "edge"); e != nil || len(parts) != 5 {
		t.Fatal(parts, e)
	}
	for _, scenario := range []string{"missing-guide", "extra-guide", "bad-kind", "index-gap"} {
		t.Run(scenario, func(t *testing.T) {
			objects := map[string]*geometryObject{}
			for k, v := range valid {
				objects[k] = v
			}
			switch scenario {
			case "missing-guide":
				delete(objects, "edge.waypoint-001")
			case "extra-guide":
				objects["edge.waypoint-004"] = &geometryObject{geometry: NativeGeometry{Kind: "sp"}}
			case "bad-kind":
				objects["edge.segment-002"] = &geometryObject{geometry: NativeGeometry{Kind: "sp"}}
			case "index-gap":
				delete(objects, "edge.segment-002")
				objects["edge.segment-004"] = &geometryObject{geometry: NativeGeometry{Kind: "cxnSp"}}
			}
			if _, e := multipartNativeRouteNames(objects, "edge"); e == nil {
				t.Fatal("incomplete/unknown native parts escaped containment")
			}
		})
	}
}
