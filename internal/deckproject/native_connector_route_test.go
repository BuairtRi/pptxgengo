package deckproject

import (
	"math"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func routedFixture(t *testing.T, mode string) (*Project, *TextBaseline) {
	t.Helper()
	p, _ := geometryFixture(t)
	dy := 12.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Separate ports vertically", Operations: []DiagramOperation{{Action: "move", ID: "node07", DY: &dy}}}
	if _, e := PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	bend := .35
	if _, e = ConnectRoutedDiagram(p, "architecture-slide", "service-edge", "node06", "right", "node07", "left", "end", "solid", mode, &bend, "test", "Show service relationship", bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	return p, b
}
func editNativeBend(t *testing.T, raw []byte, value int) []byte {
	t.Helper()
	return lineageEdit(t, raw, "ppt/slides/slide1.xml", func(part []byte) []byte {
		objects, _, e := geometryInventory(part)
		if e != nil {
			t.Fatal(e)
		}
		o := objects["service-edge"]
		var span *lineageSpan
		for _, p := range o.span.children {
			if p.node.Name.Local == "spPr" {
				for _, c := range p.children {
					if c.node.Name.Local == "prstGeom" {
						span = c
					}
				}
			}
		}
		if span == nil {
			t.Fatal("missing route")
		}
		result, e := lineageApply(part, []lineagePatch{{span.start, span.end, connectorRouteXML(&wmdesign.NativeConnectorRoute{Preset: "bentConnector3", Adjustment: value})}})
		if e != nil {
			t.Fatal(e)
		}
		return result
	})
}
func TestGeometryRoutedConnectorBendRoundTrip(t *testing.T) {
	for _, mode := range []string{"horizontal", "vertical"} {
		t.Run(mode, func(t *testing.T) {
			p, b := routedFixture(t, mode)
			inspect, e := InspectDiagram(p, "architecture-slide", bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if len(inspect.Connections) != 1 || len(inspect.Connections[0].Points) != 4 {
				t.Fatal("missing route points", inspect.Connections)
			}
			connection := inspect.Connections[0]
			a, z := connection.Points[0], connection.Points[3]
			if math.Hypot(a[0]-connection.From.X, a[1]-connection.From.Y) > .02 || math.Hypot(z[0]-connection.To.X, z[1]-connection.To.Y) > .02 {
				t.Fatal("endpoint mismatch", connection)
			}
			packet := structurePacket(t, p, b, editNativeBend(t, b.files["deck.pptx"], 65000))
			if packet.Report.Counts["geometry_native_only"] != 1 || len(packet.Report.ManualReview) != 0 {
				t.Fatal("bend not recognized", packet.Report.Counts, packet.Report.ManualReview)
			}
			decisions := geometryDecisions(packet)
			if _, e = AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine); e != nil {
				t.Fatal(e)
			}
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			route := p.Document.Slides[0].NativeGeometry["service-edge"].Route
			if route == nil || route.Adjustment != 65000 {
				t.Fatal("bend lost in YAML", route)
			}
			if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
				t.Fatal(e)
			}
			rebuilt, e := ReadTextBaseline(p, "", "")
			if e != nil {
				t.Fatal(e)
			}
			packageView, e := openLineagePackage(rebuilt.files["deck.pptx"])
			if e != nil {
				t.Fatal(e)
			}
			slideRaw, e := packageView.read("ppt/slides/slide1.xml")
			if e != nil {
				t.Fatal(e)
			}
			inv, _, e := geometryInventory(slideRaw)
			if e != nil || inv["service-edge"].geometry.Route.Adjustment != 65000 {
				t.Fatal("bend lost on rebuild", e)
			}
			hash := p.SourceHash()
			if _, e = AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine); e != nil {
				t.Fatal(e)
			}
			p, e = Load(p.SourcePath)
			if e != nil || p.SourceHash() != hash {
				t.Fatal("repeat bend adoption changed source", e)
			}
		})
	}
}
func TestGeometryRoutedConnectorBoundsBeforeWrites(t *testing.T) {
	p, b := routedFixture(t, "horizontal")
	for _, value := range []int{100000000, -100000000} {
		before := p.SourceHash()
		packet := structurePacket(t, p, b, editNativeBend(t, b.files["deck.pptx"], value))
		if _, e := AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "frame zone") {
			t.Fatal("off-frame bend accepted", e)
		}
		after, e := Load(p.SourcePath)
		if e != nil || before != after.SourceHash() {
			t.Fatal("failed route wrote source", e)
		}
	}
	if _, e := ContainDiagram(p, "architecture-slide", []string{"service-edge"}, "node05", wmdesign.DiagramPadding{Top: 28, Right: 12, Bottom: 4, Left: 12}, "test", "Constrain service arrow", bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e = ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	hash := p.SourceHash()
	packet := structurePacket(t, p, b, editNativeBend(t, b.files["deck.pptx"], 2000000))
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "containment") {
		t.Fatal("bend escaped inner box", e)
	}
	after, e := Load(p.SourcePath)
	if e != nil || hash != after.SourceHash() {
		t.Fatal("container escape wrote source", e)
	}
}
func TestGeometryConnectorRouteStrictMetadata(t *testing.T) {
	p, b := routedFixture(t, "horizontal")
	for _, replacement := range []string{
		`<a:prstGeom xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" prst="bentConnector3"><a:avLst><a:gd name="adj1" fmla="*/ w 2 1"/></a:avLst></a:prstGeom>`,
		`<a:prstGeom xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" prst="curvedConnector3"><a:avLst/></a:prstGeom>`,
	} {
		edited := lineageEdit(t, b.files["deck.pptx"], "ppt/slides/slide1.xml", func(part []byte) []byte {
			old := connectorRouteXML(&wmdesign.NativeConnectorRoute{Preset: "bentConnector3", Adjustment: 35000})
			changed := editNativeBend(t, b.files["deck.pptx"], 35000)
			packageView, e := openLineagePackage(changed)
			if e != nil {
				t.Fatal(e)
			}
			normalized, e := packageView.read("ppt/slides/slide1.xml")
			if e != nil {
				t.Fatal(e)
			}
			return []byte(strings.Replace(string(normalized), old, replacement, 1))
		})
		packet := structurePacket(t, p, b, edited)
		if len(packet.Report.ManualReview) == 0 {
			t.Fatal("unsupported route silently accepted")
		}
		for _, field := range packet.Report.Geometry {
			if field.Name == "service-edge" && field.Status != "manual_review" {
				t.Fatal("unsupported route editable", field)
			}
		}
	}
}

func TestGeometryConnectorPresetSwitchAndNativeGuideDefaults(t *testing.T) {
	p, b := routedFixture(t, "horizontal")
	edited := lineageEdit(t, b.files["deck.pptx"], "ppt/slides/slide1.xml", func(raw []byte) []byte {
		from := `<a:prstGeom prst="bentConnector3"><a:avLst><a:gd name="adj1" fmla="val 35000"/></a:avLst></a:prstGeom>`
		if !strings.Contains(string(raw), from) {
			t.Fatal("missing generated bent preset")
		}
		return []byte(strings.Replace(string(raw), from, connectorRouteXML(nil), 1))
	})
	packet := structurePacket(t, p, b, edited)
	if packet.Report.Counts["geometry_native_only"] != 1 || len(packet.Report.ManualReview) != 0 {
		t.Fatal("straight switch not proposed", packet.Report.Counts, packet.Report.ManualReview)
	}
	if _, e := AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if p.Document.Slides[0].NativeGeometry["service-edge"].Route != nil {
		t.Fatal("straight switch retained old bend")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e = ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	packet = structurePacket(t, p, b, editNativeBend(t, b.files["deck.pptx"], 50000))
	if packet.Report.Counts["geometry_native_only"] != 1 || len(packet.Report.ManualReview) != 0 {
		t.Fatal("elbow switch not proposed")
	}
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	for _, guide := range []string{"", "<a:avLst/>"} {
		n, e := readLineageXML([]byte(`<p:cxnSp xmlns:p="` + lineagePML + `" xmlns:a="` + drawingML + `"><p:spPr><a:prstGeom prst="bentConnector3">` + guide + `</a:prstGeom></p:spPr></p:cxnSp>`))
		if e != nil {
			t.Fatal(e)
		}
		route, e := readConnectorRoute(n)
		if e != nil || route == nil || route.Adjustment != 50000 {
			t.Fatal("default guide not normalized", route, e)
		}
	}
}
