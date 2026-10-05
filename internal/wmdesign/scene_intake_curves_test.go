package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

func intakeCurvePlan(t *testing.T, r *renderer, raw string, ctx SceneContext) *scenePlan {
	t.Helper()
	p, ok, err := r.planIntakeCurveScene("curve", json.RawMessage(raw), ctx)
	if err != nil || !ok {
		t.Fatalf("curve recognized=%t error=%v", ok, err)
	}
	return p
}
func TestIntakeCurveLineLabelAndOffset(t *testing.T) {
	r := intakeTestRenderer(t)
	p := intakeCurvePlan(t, r, `{"type":"teamcurve","x":100,"y":80,"w":500,"h":200,"at":[0.2,0.8],"series":[{"name":"Capacity","values":[10,20],"style":"line","fill":"#070154"}]}`, SceneContext{Surface: "light"})
	var text *TextRecord
	for _, it := range p.Items {
		if it.Text != nil {
			text = it.Text
		}
	}
	if text == nil || text.Color != "070154" || math.Abs(text.Rect.X-209) > 1e-9 {
		t.Fatalf("bad line label color/offset: %+v", text)
	}
	if len(p.Groups) != 1 || len(p.Groups[0].Parts) != 2 {
		t.Fatalf("native editable curve grouping: %+v", p.Groups)
	}
}
func TestIntakeCurveSeriesLabelsPaintLast(t *testing.T) {
	r := intakeTestRenderer(t)
	p := intakeCurvePlan(t, r, `{"type":"teamcurve","x":100,"y":80,"w":500,"h":240,"series":[{"name":"First","values":[1,2]},{"name":"Second","values":[20,30]}],"phases":[{"label":"Build"},{"label":"Run"}]}`, SceneContext{Surface: "light"})
	if len(p.Items) < 2 || p.Items[len(p.Items)-2].Text == nil || p.Items[len(p.Items)-1].Text == nil || !strings.Contains(p.Items[len(p.Items)-2].Text.ID, ".series.") || !strings.Contains(p.Items[len(p.Items)-1].Text.ID, ".series.") {
		t.Fatal("series labels are not painted above all bands and phase rules")
	}
	p = intakeCurvePlan(t, r, `{"type":"teamcurve","x":100,"y":80,"w":500,"h":200,"series":[{"name":"Last","values":[10,20],"labelAt":1}]}`, SceneContext{Surface: "light"})
	text := p.Items[len(p.Items)-1].Text
	if text == nil || text.Rect.X+text.Rect.W > 600+.02 || len(p.Warnings) == 0 {
		t.Fatalf("direct end label must fit chart: %+v", text)
	}
}
func TestIntakeCurveStrictKeysAndScalars(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"series":[{"values":[10,20]}],"phases":[{"label":"Build"}]}`)
	for _, keys := range []map[string][]string{
		{"/body/0/series": {"one", "extra"}, "/body/0/phases": {"build"}},
		{"/body/0/series": {"one"}, "/body/0/phases": {"build", "extra"}},
		{"/body/0/series": {"one"}},
	} {
		_, _, err := r.planIntakeCurveScene("curve", raw, SceneContext{Surface: "light", Path: "/body/0", Keys: keys})
		if err == nil {
			t.Fatalf("invalid key overlay accepted: %+v", keys)
		}
	}
	p, _, err := r.planIntakeCurveScene("curve", raw, SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/series": {"staff"}, "/body/0/phases": {"build"}}})
	if err != nil || !strings.Contains(p.Items[0].Shape.Record.ID, ".staff") {
		t.Fatalf("valid stable keys failed: %v", err)
	}
	for _, geometry := range []string{
		`"w":1e200,"h":200`, `"w":500,"h":200,"x":1e200`, `"w":500,"h":200,"_h":-1`,
	} {
		invalid := json.RawMessage(`{"type":"teamcurve","y":0,` + geometry + `,"series":[{"values":[10,20]}]}`)
		if _, _, err = r.planIntakeCurveScene("curve", invalid, SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("unbounded geometry accepted: %s", geometry)
		}
	}
	for _, value := range []string{
		`{"values":[10,20],"unexpected":1}`, `{"values":[10,20],"labelAt":true}`, `{"values":[10,20],"labelAt":1.5}`, `{"values":[10,-1]}`, `{"values":[10,20],"style":"unsupported"}`,
	} {
		invalid := json.RawMessage(`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"series":[` + value + `]}`)
		if _, _, err = r.planIntakeCurveScene("curve", invalid, SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("invalid nested scalar accepted: %s", value)
		}
	}
}
func TestIntakeCurveActualExtremaAndAllocation(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"teamcurve","x":100,"y":80,"w":500,"h":240,"max":100,"phaseH":40,"curve":"catmull","series":[{"values":[0,100,100,0]}]}`
	p := intakeCurvePlan(t, r, raw, SceneContext{Surface: "light"})
	shape := p.Items[0].Shape
	// The middle plateau reaches -plotH/8, not its control polygon's -plotH/6.
	if math.Abs(shape.Record.Rect.Y-55) > 1e-8 || math.Abs(p.Bounds.Y-55) > 1e-8 {
		t.Fatalf("true extrema missing: %+v %+v", shape.Record.Rect, p.Bounds)
	}
	if shape.Props.Points[2].Curve == nil || math.Abs(shape.Props.Points[2].Curve.Y1.Val*72+200./6) > 1e-8 {
		t.Fatal("cubic changed to hide overshoot")
	}
	allocation := Rect{100, 80, 500, 240}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "curve", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "curve", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(raw), Allocation: &allocation}}}}}}
	_, _, err := BuildWithEngine(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "", doc, CandidateEngine)
	if err == nil || !strings.Contains(err.Error(), "component_exceeds_allocation") {
		t.Fatalf("overshoot concealed from allocation: %v", err)
	}
	// An explicit max must exceed every cumulative sample; auto max includes
	// independent line values, but a line never changes the stacked band base.
	for _, raw := range []string{
		`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"max":0,"series":[{"values":[1,2]}]}`,
		`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"max":10,"series":[{"values":[8,8]},{"values":[3,3]}]}`,
	} {
		if _, _, err := r.planIntakeCurveScene("curve", json.RawMessage(raw), SceneContext{Surface: "light"}); err == nil {
			t.Fatal("invalid max accepted")
		}
	}
	p = intakeCurvePlan(t, r, `{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"max":10,"series":[{"values":[8,8],"style":"line","labelAt":false},{"values":[3,3]}]}`, SceneContext{Surface: "light"})
	if p.Items[1].Shape.Props.Points[0].Y.Val*72 < 100 {
		t.Fatal("line incorrectly became part of band base")
	}
}
func TestIntakeCurveExtremaContainSampledGeometry(t *testing.T) {
	for _, points := range [][]curvePoint{
		{{0, 0}, {10, 100}, {11, 100}, {12, 0}}, {{0, 10}, {20, 10}, {25, 10}}, {{0, 0}, {1, 4}, {2, -5}, {3, 2}},
	} {
		b := teamCurveBounds(points, .5)
		path := teamCurvePath(points, .5, true)
		for i := 1; i < len(path); i++ {
			a, c := path[i-1], path[i]
			for j := 0; j <= 1000; j++ {
				u := float64(j) / 1000
				v := 1 - u
				x := (v*v*v*a.X.Val + 3*v*v*u*c.Curve.X1.Val + 3*v*u*u*c.Curve.X2.Val + u*u*u*c.X.Val) * 72
				y := (v*v*v*a.Y.Val + 3*v*v*u*c.Curve.Y1.Val + 3*v*u*u*c.Curve.Y2.Val + u*u*u*c.Y.Val) * 72
				if x < b.X-1e-8 || x > b.X+b.W+1e-8 || y < b.Y-1e-8 || y > b.Y+b.H+1e-8 {
					t.Fatalf("cubic outside recorded extent: %g,%g %+v", x, y, b)
				}
			}
		}
	}
}

func TestIntakeCurveNativePackageGeometry(t *testing.T) {
	raw := json.RawMessage(`{"type":"teamcurve","x":100,"y":80,"w":500,"h":240,"max":100,"curve":"catmull","series":[{"name":"Capacity","values":[20,30,40],"style":"line","fill":"#070154"},{"name":"Team","values":[10,15,20]}]}`)
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "curve", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "curve", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
	data, report, err := BuildWithEngine(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Slides) != 1 {
		t.Fatal("missing curve report")
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var xml string
	for _, f := range z.File {
		if f.Name != "ppt/slides/slide1.xml" {
			continue
		}
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		xml = string(b)
	}
	// Two cubic segments for the line and four for the closed band, in one
	// native component group with editable text instead of a rendered bitmap.
	if strings.Count(xml, "<a:cubicBezTo>") != 6 || !strings.Contains(xml, `name="curve"`) || !strings.Contains(xml, "<p:grpSp>") || !strings.Contains(xml, "<a:close />") || strings.Contains(xml, "<p:pic>") {
		t.Fatal("native grouped cubic serialization missing or rasterized")
	}
	// Native path coordinates are EMU, converted from the frozen point geometry.
	// For the first line: plotH186, points y148.8,130.2,111.6;
	// its first cubic control y145.7 is 1,850,390 EMU.
	if !strings.Contains(xml, `y="1850390"`) {
		t.Fatal("unexpected native cubic coordinate conversion")
	}
}

func TestIntakeCurveAllFrozenV3Nodes(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewTypographyEngine(filepath.Join(bundle, "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	r := &renderer{source: source, typeEngine: engine, bundle: bundle}
	count := 0
	catalog, err := libraryCatalog(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range catalog {
		key, raw := entry.Key, entry.RawSlide
		doc, err := compileLibrarySlide(raw, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err = applyLibraryRefinements(key, source.Revision, &doc); err != nil {
			t.Fatal(err)
		}
		frame, err := source.ResolveFrame(doc.Frame)
		if err != nil {
			t.Fatal(err)
		}
		for _, node := range doc.Nodes {
			var tag struct{ Type string }
			if err = json.Unmarshal(node.Scene.Node, &tag); err != nil {
				t.Fatal(err)
			}
			if tag.Type != "teamcurve" {
				continue
			}
			p, ok, err := r.planIntakeCurveScene(node.ID, node.Scene.Node, SceneContext{Surface: frame.Request.Surface, Path: node.Scene.Path})
			if err != nil || !ok {
				t.Fatalf("%s curve failed: %v", key, err)
			}
			if err = sceneTextEnvelope(p, SceneContext{Zone: frame.Body}); err != nil {
				t.Fatalf("%s: %v", key, err)
			}
			if frame.Request.Split != "" && !inside(p.Bounds, frame.ShortBody) && !inside(p.Bounds, frame.TallBody) {
				t.Fatalf("%s actual cubic bounds outside split: %+v", key, p.Bounds)
			}
			count++
		}
	}
	if count == 0 {
		t.Fatal("missing frozen curve specimens")
	}
	t.Logf("all %d frozen v3 teamcurves planned with true geometry bounds", count)
}

func TestIntakeCurveTinyMaxRejectsNonfiniteGeometry(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"max":1e-320,"series":[{"values":[1e-9,1e-9]}]}`)
	if _, _, err := r.planIntakeCurveScene("curve", raw, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("tiny max admitted values producing nonfinite path geometry")
	}
	raw = json.RawMessage(`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"max":1e-320,"series":[{"values":[1e-321,1e-321],"labelAt":false}]}`)
	p, _, err := r.planIntakeCurveScene("curve", raw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	if !intakeFinite(p.Bounds.X, p.Bounds.Y, p.Bounds.W, p.Bounds.H) {
		t.Fatal("valid tiny values produce invalid bounds")
	}
}

func TestIntakeCurveMonotoneNoOvershootAndReverse(t *testing.T) {
	for _, points := range [][]curvePoint{{{0, 0}, {10, 100}, {11, 100}, {12, 0}}, {{0, 10}, {20, 10}, {25, 10}}, {{0, 0}, {1, 4}, {2, -5}, {3, 2}}, {{0, 10}, {1, 12}, {10, 60}, {20, 61}}} {
		path := teamCurveMonotonePath(points, true)
		rev := append([]curvePoint(nil), points...)
		for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
			rev[i], rev[j] = rev[j], rev[i]
		}
		backward := teamCurveMonotonePath(rev, false)
		for i := 1; i < len(path); i++ {
			a, b := points[i-1], points[i]
			c := path[i].Curve
			reverse := backward[len(path)-i].Curve
			if math.Abs(c.X1.Val-reverse.X2.Val) > 1e-12 || math.Abs(c.Y1.Val-reverse.Y2.Val) > 1e-12 || math.Abs(c.X2.Val-reverse.X1.Val) > 1e-12 || math.Abs(c.Y2.Val-reverse.Y1.Val) > 1e-12 {
				t.Fatal("reverse controls do not describe identical curve")
			}
			for step := 0; step <= 1000; step++ {
				u := float64(step) / 1000
				v := 1 - u
				y := v*v*v*a.y + 3*v*v*u*c.Y1.Val*72 + 3*v*u*u*c.Y2.Val*72 + u*u*u*b.y
				if y < math.Min(a.y, b.y)-1e-8 || y > math.Max(a.y, b.y)+1e-8 {
					t.Fatalf("monotone overshoot %g between %g,%g", y, a.y, b.y)
				}
			}
		}
	}
}
func TestIntakeCurveMonotoneModesAndAreaAlias(t *testing.T) {
	r := intakeTestRenderer(t)
	makePlan := func(mode string) *scenePlan {
		t.Helper()
		raw := `{"type":"teamcurve","x":100,"y":80,"w":500,"h":240,"max":100,"phaseH":40,` + mode + `"series":[{"values":[0,100,100,0],"style":"area","labelAt":false}]}`
		return intakeCurvePlan(t, r, raw, SceneContext{Surface: "light"})
	}
	cat := makePlan(`"curve":"catmull",`)
	mono := makePlan(`"curve":"monotone",`)
	legacy := makePlan("")
	if math.Abs(cat.Bounds.Y-55) > 1e-8 || math.Abs(legacy.Bounds.Y-cat.Bounds.Y) > 1e-8 {
		t.Fatal("published default Catmull geometry changed")
	}
	if mono.Bounds.Y < 80-1e-8 || !inside(mono.Bounds, Rect{100, 80, 500, 240}) {
		t.Fatalf("monotone escaped sample extrema/allocation: %+v", mono.Bounds)
	}
	r.source.Revision = IntakeTeamCurveMonotoneRevision
	modern := makePlan("")
	if math.Abs(modern.Bounds.Y-mono.Bounds.Y) > 1e-8 {
		t.Fatal("registered next revision default is not monotone")
	}
	r.source.Revision = "future-unregistered"
	if math.Abs(makePlan("").Bounds.Y-cat.Bounds.Y) > 1e-8 {
		t.Fatal("unknown revision silently changed curve default")
	}
	if _, _, err := r.planIntakeCurveScene("curve", json.RawMessage(`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"curve":"spline","series":[{"values":[1,2]}]}`), SceneContext{Surface: "light"}); err == nil {
		t.Fatal("unknown curve mode accepted")
	}
}
func TestIntakeCurveMonotoneNativeCubicAndDashed(t *testing.T) {
	r := intakeTestRenderer(t)
	p := intakeCurvePlan(t, r, `{"type":"teamcurve","x":100,"y":80,"w":500,"h":240,"curve":"monotone","series":[{"values":[20,30,40],"style":"line","dashed":true,"labelAt":false}]}`, SceneContext{Surface: "light"})
	s := p.Items[0].Shape
	if s.Props.Line.DashType != "dash" || s.Props.Points[1].Curve.Type != "cubic" || s.Props.Fill.Type != "none" {
		t.Fatal("monotone line lost native cubic/dash semantics")
	}
	p = intakeCurvePlan(t, r, `{"type":"teamcurve","x":100,"y":80,"w":500,"h":240,"curve":"monotone","max":10,"series":[{"values":[8,8],"style":"line","labelAt":false},{"values":[3,3],"labelAt":false}]}`, SceneContext{Surface: "light"})
	if p.Items[1].Shape.Props.Points[0].Y.Val*72 < 100 {
		t.Fatal("independent line became stacked base")
	}
}

func TestIntakeCurveMonotoneRejectsDerivedNonfiniteAndHugeCopy(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, raw := range []string{
		`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"curve":"monotone","at":[0,1e-320,1],"series":[{"values":[1,2,3]}]}`,
		`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"curve":"monotone","series":[{"values":[1e300,1e300]}]}`,
		`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"curve":"monotone","series":[{"name":"` + strings.Repeat("a", 4097) + `","values":[1,2]}]}`,
		`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"curve":"monotone","series":[{"values":[1,2]}],"phases":[{"label":"` + strings.Repeat("a", 4097) + `"}]}`,
	} {
		if _, _, e := r.planIntakeCurveScene("curve", json.RawMessage(raw), SceneContext{Surface: "light"}); e == nil {
			t.Fatal("unsafe derived path/value/copy accepted")
		}
	}
}

func TestIntakeCurveRequiresNumericGeometry(t *testing.T) {
	r := intakeTestRenderer(t)
	base := map[string]any{"type": "teamcurve", "x": 0, "y": 0, "w": 500, "h": 300, "series": []any{map[string]any{"values": []any{1, 2}, "labelAt": false}}}
	raw, _ := json.Marshal(base)
	intakeCurvePlan(t, r, string(raw), SceneContext{Surface: "light"})
	for _, field := range []string{"x", "y", "w", "h"} {
		for _, value := range []any{nil, "missing"} {
			node := map[string]any{}
			for k, v := range base {
				node[k] = v
			}
			if value == "missing" {
				delete(node, field)
			} else {
				node[field] = nil
			}
			raw, _ := json.Marshal(node)
			if _, _, e := r.planIntakeCurveScene("curve", raw, SceneContext{Surface: "light"}); e == nil {
				t.Fatalf("geometry accepted: %s=%v", field, value)
			}
		}
	}
}
