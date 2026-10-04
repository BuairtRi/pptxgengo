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

	"github.com/buairtri/pptxgengo/pptx"
)

// Frozen from wm-design-system commit c7678b0efd48e96a2be290fbdee6ed8914a604a1.
// venn.json SHA256 a0c26d8d4a804b6d8016e8e1b9adf1fce6cedb59d28f18aeb5922ebddcd809fa.
// Independently matched all12 nodes to root's 2026-10-03 17:50 frozen intake.
// Frozen full renderer SHA256 a7f1f046d38723cb86d6e0dbacd97886c183784df1e4949192322e74d6a33cf6.
// Frozen Venn case SHA256 b540d7f47097be4e53fbed3e68113d0d26e90d3d3d8896703b03dd8044bf4d29.
// Runtime source edits cannot silently change these review fixtures.
var frozenIntakeVennNodes = []struct {
	key string
	raw string
}{
	{"venn/two-text", `{"type":"venn","x":345,"y":36,"w":558,"h":432,"sets":[{"label":"Clinical operations","text":"Case mix, staffing and patient flow"},{"label":"Finance","text":"Cost, reimbursement and forecasting"}],"regions":[{"in":[0,1],"label":"Margin per case","text":"One shared view"}]}`},
	{"venn/two-points", `{"type":"venn","x":57,"y":36,"w":558,"h":432,"sets":[{"label":"Cost out"},{"label":"Revenue up"}],"points":[{"x":0.204,"y":0.28,"label":"Claims bot"},{"x":0.231,"y":0.72,"label":"Staffing"},{"x":0.446,"y":0.5,"label":"Referrals"},{"x":0.679,"y":0.28,"label":"Outreach"},{"x":0.688,"y":0.72,"label":"Pricing"}],"labelStyle":"mono"}`},
	{"venn/two-bullets", `{"type":"venn","x":345,"y":36,"w":558,"h":432,"sets":[{"label":"Build","bullets":["Own roadmap","Slower start"]},{"label":"Buy","bullets":["Fast launch","Vendor lock-in"]}],"regions":[{"in":[0,1],"label":"Hybrid","bullets":["Buy core","Build edge"]}],"labelStyle":"mono"}`},
	{"venn/two-nav", `{"type":"venn","x":57,"y":126,"w":558,"h":324,"sets":[{"label":"Internal team","text":"Knows Northfield systems"},{"label":"Partner","text":"Brings scale and tooling"}],"regions":[{"in":[0,1],"label":"Co-delivery","text":"Shared plan"}],"labelStyle":"mono"}`},
	{"venn/three-text", `{"type":"venn","x":345,"y":36,"w":558,"h":414,"sets":[{"label":"Desirable","text":"Patients and staff ask for it"},{"label":"Feasible","text":"Data and systems are ready"},{"label":"Viable","text":"Payback inside 18 months"}],"regions":[{"in":[0,1],"label":"No ROI","w":28},{"in":[0,2],"label":"Not yet","w":24},{"in":[1,2],"label":"No pull","w":30},{"in":[0,1,2],"label":"Sweet spot"}],"labelStyle":"mono"}`},
	{"venn/three-bullets", `{"type":"venn","x":345,"y":36,"w":558,"h":414,"sets":[{"label":"AI","bullets":["Predicts denials","Drafts appeals"]},{"label":"Process","bullets":["One close calendar","Standard intake"]},{"label":"Data","bullets":["Clean claims feed","Shared patient ID"]}],"regions":[{"in":[0,1],"label":"Rules"},{"in":[0,2],"label":"ML"},{"in":[1,2],"label":"Flow"},{"in":[0,1,2],"label":"Automation"}],"labelStyle":"mono"}`},
	{"venn/three-points", `{"type":"venn","x":345,"y":36,"w":558,"h":414,"sets":[{"label":"AI"},{"label":"Process"},{"label":"Data"}],"points":[{"x":0.231,"y":0.415,"label":"Denial"},{"x":0.769,"y":0.415,"label":"Intake","side":"left"},{"x":0.428,"y":0.754,"label":"Master index"},{"x":0.446,"y":0.331,"label":"Appeals"},{"x":0.306,"y":0.577,"label":"Claims"},{"x":0.461,"y":0.483,"label":"Coding"}],"labelStyle":"mono"}`},
	{"venn/three-left", `{"type":"venn","x":57,"y":36,"w":558,"h":414,"sets":[{"label":"People","text":"Roles, skills and capacity"},{"label":"Process","text":"Standard ways of working"},{"label":"Technology","text":"Shared platforms"}],"regions":[{"in":[0,1],"label":"Roles"},{"in":[0,2],"label":"UX"},{"in":[1,2],"label":"Auto"},{"in":[0,1,2],"label":"One model"}]}`},
	{"venn/three-nav", `{"type":"venn","x":345,"y":36,"w":558,"h":414,"sets":[{"label":"Desirable","text":"Patients and staff ask for it"},{"label":"Feasible","text":"Data and systems are ready"},{"label":"Viable","text":"Payback inside 18 months"}],"regions":[{"in":[0,1],"label":"No ROI","w":28},{"in":[0,2],"label":"Not yet","w":24},{"in":[1,2],"label":"No pull","w":30},{"in":[0,1,2],"label":"Sweet spot"}],"labelStyle":"mono"}`},
	{"venn/four-text", `{"type":"venn","x":345,"y":36,"w":558,"h":432,"sets":[{"label":"Access","text":"Shorter waits"},{"label":"Quality","text":"Fewer readmits"},{"label":"Cost","text":"Lower unit cost"},{"label":"Experience","text":"Easier journeys"}],"regions":[{"in":[0,1],"label":"Timely care"},{"in":[0,2],"label":"Affordable"},{"in":[1,3],"label":"Trust"},{"in":[2,3],"label":"Fair price"},{"in":[0,1,2,3],"label":"VBC"}],"labelStyle":"mono"}`},
	{"venn/four-points", `{"type":"venn","x":57,"y":36,"w":558,"h":432,"sets":[{"label":"Data"},{"label":"Process"},{"label":"People"},{"label":"Technology"}],"points":[{"x":0.249,"y":0.326,"label":"MDM hub"},{"x":0.751,"y":0.326,"label":"Calendar","side":"left"},{"x":0.751,"y":0.685,"label":"ERP","side":"left"},{"x":0.464,"y":0.234,"label":"Steward"},{"x":0.464,"y":0.766,"label":"Academy"},{"x":0.285,"y":0.5,"label":"Skills"}],"labelStyle":"mono"}`},
	{"venn/four-callout", `{"type":"venn","x":57,"y":36,"w":558,"h":432,"sets":[{"label":"Strategy","text":"Sets direction"},{"label":"Finance","text":"Funds the plan"},{"label":"Operations","text":"Runs the work"},{"label":"Technology","text":"Builds the tools"}],"regions":[{"in":[0,1],"label":"Targets"},{"in":[0,2],"label":"Priorities"},{"in":[1,3],"label":"Business case"},{"in":[2,3],"label":"Delivery"},{"in":[0,1,2,3],"label":"CORE"}],"labelStyle":"mono"}`},
}

func intakeVennPlan(t *testing.T, r *renderer, raw string, ctx SceneContext) *scenePlan {
	t.Helper()
	p, handled, err := r.planIntakeVennScene("venn", json.RawMessage(raw), ctx)
	if err != nil || !handled {
		t.Fatalf("Venn handled=%v err=%v", handled, err)
	}
	return p
}
func TestIntakeVennFrozenSourceNodes(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, fixture := range frozenIntakeVennNodes {
		t.Run(fixture.key, func(t *testing.T) {
			p := intakeVennPlan(t, r, fixture.raw, SceneContext{Surface: "light", Path: "/body/0"})
			var node intakeVennSource
			if err := json.Unmarshal([]byte(fixture.raw), &node); err != nil {
				t.Fatal(err)
			}
			circles := 0
			for _, item := range p.Items {
				if item.Image != nil || item.Chart != nil || item.Table != nil {
					t.Fatal("Venn is not native editable shapes/text")
				}
				if item.Shape != nil {
					if !inside(item.Shape.Record.Rect, p.Bounds) {
						t.Fatalf("shape outside reported bounds %+v", item.Shape.Record)
					}
					if strings.Contains(item.Shape.Record.ID, ".sets.") && strings.HasSuffix(item.Shape.Record.ID, ".circle") {
						circles++
						if item.Shape.Type != pptx.ShapeTypeEllipse || math.Abs(item.Shape.Props.Fill.Transparency-30) > 1e-8 {
							t.Fatal("circle fill opacity changed")
						}
					}
				}
				if item.Text != nil && !inside(item.Text.Rect, p.Bounds) {
					t.Fatalf("text outside reported bounds %+v", item.Text)
				}
			}
			if circles != len(node.Sets) || len(p.Groups) != 1 || p.Groups[0].Contract != IntakeVennContract {
				t.Fatal("Venn identities/group missing")
			}
			if err := sceneTextEnvelope(p, SceneContext{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestIntakeVennGeometryLensAndRegionWidth(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":[{"label":"Data"},{"label":"People"},{"label":"Process"}],"regions":[{"in":[0,1],"label":"Pair","w":80},{"in":[0,1,2],"label":"Core"}],"labelStyle":"mono"}`
	p := intakeVennPlan(t, r, raw, SceneContext{})
	var n intakeVennSource
	json.Unmarshal([]byte(raw), &n)
	radius, centers := intakeVennGeometry(n)
	if math.Abs(radius-432/3.1) > 1e-8 {
		t.Fatal("three circle radius")
	}
	first := p.Items[0].Shape.Record.Rect
	if math.Abs(first.X-(math.Round(centers[0][0]*10)/10-math.Round(radius*10)/10)) > 1e-8 || math.Abs(first.W-278.8) > 1e-8 {
		t.Fatalf("browser one-decimal circles changed: %+v", first)
	}
	for _, item := range p.Items {
		if item.Text == nil {
			continue
		}
		tr := item.Text
		if strings.Contains(tr.ID, ".sets.") && (tr.Layout.Style.Family != "IBM Plex Mono" || tr.Layout.Style.Weight != 600 || tr.Layout.Displayed != strings.ToUpper(tr.Layout.Original)) {
			t.Fatal("mono label caps missing")
		}
		if strings.Contains(tr.ID, ".regions.source-001.label") {
			anchor := intakeVennAnchor(centers, []int{0, 1}, [2]float64{379, 296}, radius, false)
			if tr.Rect.W != 80 || math.Abs(tr.Rect.X+40-anchor[0]) > 1e-8 || anchor[1] >= centers[0][1] {
				t.Fatalf("pair lens anchor not pushed outward: %+v", tr.Rect)
			}
			if math.Abs(tr.Rect.Y-(anchor[1]-9-tr.Rect.H*.4)) > 1e-8 {
				t.Fatal("40percent vertical centering changed")
			}
		}
	}
}
func TestIntakeVennPointsBadgesAndExplicitNumbers(t *testing.T) {
	r := intakeTestRenderer(t)
	p := intakeVennPlan(t, r, `{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":[{"label":"A"},{"label":"B"}],"points":[{"x":0.3,"y":0.3,"n":"X","label":"Left","side":"left"},{"x":0.6,"y":0.6,"n":999,"label":"Right"},{"x":0.5,"y":0.5}],"numbered":false,"opacity":0}`, SceneContext{})
	text := map[string]*TextRecord{}
	shapes := map[string]*sceneShape{}
	for _, item := range p.Items {
		if item.Text != nil {
			text[item.Text.ID] = item.Text
		}
		if item.Shape != nil {
			shapes[item.Shape.Record.ID] = item.Shape
		}
	}
	if text["venn.points.source-001.number"].Layout.Displayed != "X" || text["venn.points.source-002.number"].Layout.Displayed != "999" || text["venn.points.source-003.number"] != nil {
		t.Fatal("explicit or unnumbered marker contract changed")
	}
	for _, key := range []string{"source-001", "source-002"} {
		badge := shapes["venn.points."+key+".badge"]
		marker := shapes["venn.points."+key+".marker"]
		halo := shapes["venn.points."+key+".halo"]
		if halo.Record.Rect.W != 21 || marker.Record.Rect.W != 18 || marker.Props.Line.Type != "none" {
			t.Fatal("outside white halo/native marker geometry wrong")
		}
		if key == "source-001" && badge.Record.Rect.X+badge.Record.Rect.W > marker.Record.Rect.X-3+.01 {
			t.Fatal("left badge placement")
		}
		if key == "source-002" && badge.Record.Rect.X < marker.Record.Rect.X+21-.01 {
			t.Fatal("right badge placement")
		}
		label := text["venn.points."+key+".label"]
		if label.Layout.Style.Size != 8 || label.Layout.Style.Weight != 600 || label.Layout.Displayed != strings.ToUpper(label.Layout.Original) {
			t.Fatal("point badge typography")
		}
	}
	if p.Items[0].Shape.Props.Fill.Transparency != 100 {
		t.Fatal("explicit zero opacity ignored")
	}
}
func TestIntakeVennStrictFieldsAndCardinality(t *testing.T) {
	r := intakeTestRenderer(t)
	base := `{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":[{"label":"A"},{"label":"B"}]}`
	for _, extra := range []string{`"w":0`, `"w":1e200`, `"h":-1`, `"opacity":1.1`, `"opacity":-1`, `"textW":0`, `"textW":1000`, `"labelStyle":"body"`, `"numbered":"no"`, `"unknown":true`, `"W":558`, `"regions":[{"in":[0,0],"label":"Bad"}]`, `"regions":[{"in":[0,2],"label":"Bad"}]`, `"regions":[{"in":[0],"label":"Bad"}]`, `"regions":[{"in":[0,1],"label":"A"},{"in":[1,0],"text":"B"}]`, `"regions":[{"in":[0,1],"fill":"#070154"}]`, `"points":[{"x":-0.1,"y":0.5}]`, `"points":[{"y":0.5}]`, `"points":[{"x":0.5,"y":0.5,"side":"up"}]`, `"points":[{"x":0.5,"y":0.5,"n":null}]`, `"points":[{"x":0.5,"y":0.5,"n":true}]`, `"points":[{"x":0.5,"y":0.5,"label":"This point badge is far too wide to fit in its one line allocation"}]`} {
		raw := strings.TrimSuffix(base, "}") + "," + extra + "}"
		if _, handled, err := r.planIntakeVennScene("bad", json.RawMessage(raw), SceneContext{}); !handled || err == nil {
			t.Fatalf("invalid Venn accepted: %s", extra)
		}
	}
	for _, sets := range []string{`[]`, `[{"label":"A"}]`, `[{"label":"A"},{"label":"B"},{"label":"C"},{"label":"D"},{"label":"E"}]`, `[{"label":"A","bullets":[{"text":"x","other":1}]},{"label":"B"}]`, `[{"label":"A","bullets":[{"lead":"Hidden","text":"x","sub":["suppressed"]}]},{"label":"B"}]`} {
		raw := `{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":` + sets + `}`
		if _, _, err := r.planIntakeVennScene("bad", json.RawMessage(raw), SceneContext{}); err == nil {
			t.Fatalf("invalid sets accepted %s", sets)
		}
	}
	// Replace existing fields here so geometry failures are not masked by
	// the separate duplicate-key rejection exercised above.
	for _, changed := range []string{`{"w":0}`, `{"w":1e200}`, `{"h":-1}`, `{"x":1e200}`, `{"y":1e200}`, `{"x":null}`, `{"y":null}`, `{"w":null}`, `{"h":null}`, `{"_h":-1}`, `{"_h":1e200}`, `{"_h":null}`} {
		var object, replacement map[string]json.RawMessage
		json.Unmarshal([]byte(base), &object)
		json.Unmarshal([]byte(changed), &replacement)
		for key, value := range replacement {
			object[key] = value
		}
		raw, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err = r.planIntakeVennScene("bad", raw, SceneContext{}); err == nil || !strings.Contains(err.Error(), "invalid_geometry") {
			t.Fatalf("geometry guard did not reject %s: %v", changed, err)
		}
	}
}

func TestIntakeVennPointLabelsAreLiteral(t *testing.T) {
	r := intakeTestRenderer(t)
	p := intakeVennPlan(t, r, `{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":[{"label":"A"},{"label":"B"}],"points":[{"x":0.5,"y":0.5,"label":"[[Raw]]"}]}`, SceneContext{})
	for _, item := range p.Items {
		if item.Text != nil && strings.HasSuffix(item.Text.ID, ".points.source-001.label") {
			if item.Text.Layout.Displayed != "[[RAW]]" || item.Text.Rich != nil {
				t.Fatal("literal point badge was interpreted as markup")
			}
			return
		}
	}
	t.Fatal("point label missing")
}

func TestIntakeVennNodeWideResourceLimits(t *testing.T) {
	r := intakeTestRenderer(t)
	items := make([]json.RawMessage, 32)
	for i := range items {
		items[i] = json.RawMessage(`"Bullet"`)
	}
	n := intakeVennSource{Type: "venn", X: 100, Y: 80, W: 558, H: 432, Sets: []intakeVennSet{{Label: "A", Bullets: items}, {Label: "B", Bullets: items}, {Label: "C", Bullets: items}, {Label: "D", Bullets: items}}, Regions: []intakeVennRegion{{In: []int{0, 1}, Bullets: []json.RawMessage{json.RawMessage(`"Extra"`)}}}}
	raw, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = r.planIntakeVennScene("bad", raw, SceneContext{}); err == nil || !strings.Contains(err.Error(), "bullet_limit") {
		t.Fatalf("node-wide bullet count did not reject before shaping: %v", err)
	}
	long, _ := json.Marshal(strings.Repeat("x", 8192))
	for i := range items {
		items[i] = long
	}
	n.Regions = nil
	n.Sets = n.Sets[:2]
	raw, err = json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = r.planIntakeVennScene("bad", raw, SceneContext{}); err == nil || !strings.Contains(err.Error(), "copy_limit") {
		t.Fatalf("node-wide copy limit did not reject before shaping: %v", err)
	}
}

func TestIntakeVennSurfaceFillAndPointLimits(t *testing.T) {
	r := intakeTestRenderer(t)
	p := intakeVennPlan(t, r, `{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":[{"label":"A","fill":"#0047FF"},{"label":"B"}],"opacity":0.25,"labelStyle":"mono"}`, SceneContext{Surface: "inverse"})
	if p.Items[0].Shape.Record.Color != "0047FF" || p.Items[0].Shape.Props.Fill.Transparency != 75 {
		t.Fatal("explicit fill/opacity changed")
	}
	for _, item := range p.Items {
		if item.Text != nil && item.Text.Color != "FFFFFF" {
			t.Fatal("inverse surface display ink not preserved")
		}
	}
	points := make([]intakeVennPoint, 65)
	x, y := .5, .5
	for i := range points {
		points[i] = intakeVennPoint{X: &x, Y: &y}
	}
	n := intakeVennSource{Type: "venn", X: 100, Y: 80, W: 558, H: 432, Sets: []intakeVennSet{{Label: "A"}, {Label: "B"}}, Points: points}
	raw, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = r.planIntakeVennScene("bad", raw, SceneContext{}); err == nil || !strings.Contains(err.Error(), "cardinality") {
		t.Fatalf("point cardinality limit not enforced: %v", err)
	}
}
func TestIntakeVennStableKeysAndActualBleed(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":[{"label":"A","bullets":["Alpha"]},{"label":"B"}],"regions":[{"in":[0,1],"label":"Both"}],"points":[{"x":0,"y":0,"label":"Edge","side":"left"}]}`
	ctx := SceneContext{Path: "/body/0", Keys: map[string][]string{"/body/0/sets": {"a", "b"}, "/body/0/sets/0/bullets": {"alpha"}, "/body/0/regions": {"both"}, "/body/0/points": {"edge"}}}
	p := intakeVennPlan(t, r, raw, ctx)
	if p.Bounds.X >= 89.5 || p.Bounds.Y > 69.5 {
		t.Fatalf("edge marker/badge bleed concealed %+v", p.Bounds)
	}
	if !strings.Contains(p.Items[0].Shape.Record.ID, ".sets.a.circle") {
		t.Fatal("stable keyed identity missing")
	}
	for _, path := range []string{"/body/0/sets", "/body/0/regions", "/body/0/points", "/body/0/sets/0/bullets"} {
		copy := ctx
		copy.Keys = map[string][]string{}
		for k, v := range ctx.Keys {
			if k != path {
				copy.Keys[k] = v
			}
		}
		if _, _, err := r.planIntakeVennScene("bad", json.RawMessage(raw), copy); err == nil {
			t.Fatalf("missing key overlay accepted %s", path)
		}
	}
}
func TestIntakeVennNativePackage(t *testing.T) {
	raw := json.RawMessage(frozenIntakeVennNodes[4].raw)
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "venn", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "venn", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
	data, report, err := BuildWithEngine(filepath.Join("..", "..", "library", "wm-design-system", "v2"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Slides) != 1 {
		t.Fatal("Venn native report missing")
	}
	found := false
	for _, scene := range report.Slides[0].Scenes {
		for _, group := range scene.Groups {
			found = found || group.Contract == IntakeVennContract
		}
	}
	if !found {
		t.Fatal("Venn native group report missing")
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	xml := ""
	for _, f := range z.File {
		if f.Name == "ppt/slides/slide1.xml" {
			reader, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			b, e := io.ReadAll(reader)
			reader.Close()
			if e != nil {
				t.Fatal(e)
			}
			xml = string(b)
		}
	}
	if strings.Count(xml, `prst="ellipse"`) != 3 || !strings.Contains(xml, "<p:grpSp>") || strings.Contains(xml, "<p:pic>") || !strings.Contains(xml, `<a:alpha val="70000"`) {
		t.Fatal("editable translucent Venn shapes/group missing")
	}
	if !strings.Contains(xml, "IBM Plex Mono") || !strings.Contains(xml, "SWEET SPOT") {
		t.Fatal("editable set/region text missing")
	}
}
