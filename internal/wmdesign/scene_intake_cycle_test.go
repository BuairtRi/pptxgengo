package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cyclePlan(t *testing.T, r *renderer, raw string, ctx SceneContext) *scenePlan {
	t.Helper()
	p, ok, e := r.planIntakeCycleScene("cycle", json.RawMessage(raw), ctx)
	if e != nil || !ok {
		t.Fatalf("recognized%t %v", ok, e)
	}
	return p
}

const cycleSimple = `{"type":"cycle","x":100,"y":120,"w":600,"h":300,"nodeW":144,"nodeH":54,"numbered":true,"active":1,"items":[{"label":"Start"},{"label":"Assess"},{"label":"Deliver"},{"label":"Review"}],"center":{"label":"Cycle","title":"Shared review"},"loops":[{"from":3,"to":1,"label":"Feedback","side":-1}]}`

func TestIntakeCycleFrozenSourceNodes(t *testing.T) {
	r := intakeTestRenderer(t)
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-round12", "source")
	data, e := os.ReadFile(filepath.Join(root, "templates", "library", "lifecycle.json"))
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != IntakeCycleSourceSHA256 {
		t.Fatal("frozen lifecycle source changed")
	}
	renderer, e := os.ReadFile(filepath.Join(root, "explorations", "components.src.html"))
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(renderer)) != IntakeCycleRendererSHA256 {
		t.Fatal("frozen cycle renderer changed")
	}
	var j struct {
		Templates []struct {
			Variant string
			Slide   struct{ Body []json.RawMessage }
		}
	}
	if e = json.Unmarshal(data, &j); e != nil {
		t.Fatal(e)
	}
	count, passed, rejected := 0, 0, 0
	for _, entry := range j.Templates {
		for i, raw := range entry.Slide.Body {
			var tag struct{ Type string }
			json.Unmarshal(raw, &tag)
			if tag.Type != "cycle" {
				continue
			}
			count++
			p, ok, e := r.planIntakeCycleScene("cycle", raw, SceneContext{Surface: "light"})
			if !ok {
				t.Fatal("source cycle not recognized")
			}
			if e != nil {
				if !strings.Contains(e.Error(), "scene.cycle_node_text_overflow") {
					t.Fatalf("%s node%d: %v", entry.Variant, i, e)
				}
				rejected++
				t.Logf("honest source fit rejection %s/body/%d: %v", entry.Variant, i, e)
				continue
			}
			if len(p.Groups) != 1 || p.Groups[0].Contract != IntakeCycleContract {
				t.Fatal("native group/contract missing")
			}
			if e = sceneTextEnvelope(p, SceneContext{Zone: Rect{0, 0, 960, 486}}); e != nil {
				t.Fatal(e)
			}
			passed++
			t.Logf("accepted source %s/body/%d bounds%+v", entry.Variant, i, p.Bounds)
		}
	}
	if count != 18 || passed != 14 || rejected != 4 {
		t.Fatalf("source nodes%d passed%d rejected%d", count, passed, rejected)
	}
	t.Logf("cycle source original fit: %d accepted, %d rejected; named repairs remain separate", passed, rejected)
}
func TestIntakeCycleGeometryAndClosedRing(t *testing.T) {
	r := intakeTestRenderer(t)
	p := cyclePlan(t, r, cycleSimple, SceneContext{Surface: "light"})
	arcs, loops := 0, 0
	var active *sceneShape
	var center *TextRecord
	for _, it := range p.Items {
		if it.Shape != nil {
			s := it.Shape
			if strings.HasSuffix(s.Record.ID, ".arc") {
				arcs++
				if len(s.Props.Points) != 25 {
					t.Fatal("arc is not native25point path")
				}
			}
			if strings.HasSuffix(s.Record.ID, ".curve") {
				loops++
				if s.Props.Points[1].Curve.Type != "quadratic" || s.Props.Line.DashType != "dash" || s.Props.Line.Width != 1.5 {
					t.Fatal("quadratic loop contract")
				}
			}
			if s.Record.ID == "cycle.items.source-002.surface" {
				active = s
			}
		}
		if it.Text != nil && it.Text.ID == "cycle.center.label" {
			center = it.Text
		}
	}
	if arcs != 4 || loops != 1 || active == nil || active.Props.Fill.Color != "070154" || center == nil || center.Align != "center" {
		t.Fatal("arc/loop/active/center semantics")
	}
	raw := strings.TrimSuffix(cycleSimple, "}") + `,"closed":false}`
	p = cyclePlan(t, r, raw, SceneContext{Surface: "light"})
	arcs = 0
	for _, it := range p.Items {
		if it.Shape != nil && strings.HasSuffix(it.Shape.Record.ID, ".arc") {
			arcs++
		}
	}
	if arcs != 3 {
		t.Fatal("open cycle retained closing arc")
	}
	raw = strings.TrimSuffix(cycleSimple, "}") + `,"ring":false,"align":"left"}`
	p = cyclePlan(t, r, raw, SceneContext{Surface: "light"})
	for _, it := range p.Items {
		if it.Shape != nil && strings.HasSuffix(it.Shape.Record.ID, ".arc") {
			t.Fatal("ringfalse emittedarc")
		}
		if it.Text != nil && strings.Contains(it.Text.ID, ".items.") && it.Text.Align != "left" {
			t.Fatal("node alignment")
		}
	}
}
func TestIntakeCycleQuadraticTrueBoundsAndOffsets(t *testing.T) {
	a, c, b := curvePoint{0, 0}, curvePoint{80, -100}, curvePoint{100, 20}
	box := cycleQuadraticBounds(a, c, b)
	for i := 0; i <= 1000; i++ {
		q := cycleQuadraticPoint(a, c, b, float64(i)/1000)
		if q.x < box.X-1e-8 || q.x > box.X+box.W+1e-8 || q.y < box.Y-1e-8 || q.y > box.Y+box.H+1e-8 {
			t.Fatal("quadratic bounds exclude curve")
		}
	}
	if box.Y <= c.y || box.Y >= 0 {
		t.Fatal("control polygon incorrectly treated as curve ink")
	}
	if q := cycleOffset(curvePoint{0, 0}, curvePoint{0, 100}, 100, 50); math.Abs(q.x) > 1e-8 || q.y != 31 {
		t.Fatalf("box clearance %+v", q)
	}
}
func TestIntakeCycleStrictGeometryCopyAndKeys(t *testing.T) {
	r := intakeTestRenderer(t)
	var base map[string]any
	json.Unmarshal([]byte(cycleSimple), &base)
	for _, field := range []string{"x", "y", "w", "h"} {
		for _, omit := range []bool{false, true} {
			var n map[string]any
			json.Unmarshal([]byte(cycleSimple), &n)
			if omit {
				delete(n, field)
			} else {
				n[field] = nil
			}
			raw, _ := json.Marshal(n)
			if _, _, e := r.planIntakeCycleScene("cycle", raw, SceneContext{Surface: "light"}); e == nil {
				t.Fatal("missing/null geometry accepted")
			}
		}
	}
	for _, change := range []string{`"active":20`, `"nodeW":0`, `"nodeH":400`, `"start":1e200`, `"unknown":1`, `"align":"bad"`, `"ring":"false"`, `"center":{"w":1e200}`, `"items":[{"label":"A","unknown":1},{"label":"B"}]`, `"loops":[{"from":0,"to":8}]`, `"loops":[{"to":0}]`, `"loops":[{"from":0,"to":1,"side":0}]`, `"loops":[{"from":0,"to":1,"bend":1e200}]`} {
		// Replace matching top-level fields via decoding, so duplicate members aren't used.
		var n map[string]any
		json.Unmarshal([]byte(cycleSimple), &n)
		var v map[string]any
		json.Unmarshal([]byte("{"+change+"}"), &v)
		for k, x := range v {
			n[k] = x
		}
		raw, _ := json.Marshal(n)
		if _, _, e := r.planIntakeCycleScene("cycle", raw, SceneContext{Surface: "light"}); e == nil {
			t.Fatalf("invalid accepted %s", change)
		}
	}
	base["items"] = []any{map[string]any{"label": strings.Repeat("long ", 2000)}, map[string]any{"label": "B"}}
	raw, _ := json.Marshal(base)
	if _, _, e := r.planIntakeCycleScene("cycle", raw, SceneContext{Surface: "light"}); e == nil {
		t.Fatal("unbounded copy accepted")
	}
	for _, keys := range []map[string][]string{{}, {"/body/0/items": {"a", "b", "c", "d"}}, {"/body/0/items": {"a", "a", "c", "d"}, "/body/0/loops": {"feedback"}}} {
		if _, _, e := r.planIntakeCycleScene("cycle", json.RawMessage(cycleSimple), SceneContext{Surface: "light", Path: "/body/0", Keys: keys}); e == nil {
			t.Fatal("invalid key overlay accepted")
		}
	}
	p := cyclePlan(t, r, cycleSimple, SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/items": {"start", "assess", "deliver", "review"}, "/body/0/loops": {"feedback"}}})
	found := false
	for _, it := range p.Items {
		if it.Text != nil && it.Text.ID == "cycle.items.assess.label" {
			found = true
		}
	}
	if !found {
		t.Fatal("stable item identity missing")
	}
}
func TestIntakeCycleNativeEditableXML(t *testing.T) {
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "cycle", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "cycle", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(cycleSimple)}}}}}}
	data, _, e := BuildWithEngine(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "", doc, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		t.Fatal(e)
	}
	var xml string
	for _, f := range z.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rd, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			b, e := io.ReadAll(rd)
			rd.Close()
			if e != nil {
				t.Fatal(e)
			}
			xml = string(b)
		}
	}
	if !strings.Contains(xml, "<a:quadBezTo>") || !strings.Contains(xml, "<p:grpSp>") || strings.Contains(xml, `name="cycle.items.source-001.surface"`) == false || strings.Count(xml, "<a:lnTo>") < 96 || !strings.Contains(xml, `val="dash"`) {
		t.Fatal("cycle isn't native editable paths/nodes/groups")
	}
}

func TestIntakeCycleSourceLiteralAndFootnoteMeasurement(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"cycle","x":100,"y":120,"w":600,"h":300,"nodeW":180,"nodeH":90,"items":[{"label":"Ready[^1]","text":"[[Literal]]"},{"label":"Second"}],"center":{"title":"[[Center]]"},"loops":[{"from":0,"to":1,"label":"[[Loop]]"}]}`
	p := cyclePlan(t, r, raw, SceneContext{Surface: "light", Notes: []string{"Qualified example."}})
	note, literal, center, loop := false, false, false, false
	for _, it := range p.Items {
		if it.Text == nil {
			continue
		}
		tr := it.Text
		if tr.ID == "cycle.items.source-001.label" {
			if tr.Rich != nil && strings.Contains(tr.Layout.Displayed, "Ready") {
				for _, para := range tr.Rich.Paragraphs {
					for _, run := range para.Runs {
						note = note || run.BaselineShift > 0 && run.Displayed == "1"
					}
				}
			}
		}
		if tr.ID == "cycle.items.source-001.text" {
			literal = tr.Rich == nil && tr.Layout.Displayed == "[[Literal]]"
		}
		if tr.ID == "cycle.center.title" {
			center = tr.Rich == nil && tr.Layout.Displayed == "[[Center]]"
		}
		if tr.ID == "cycle.loops.source-001.label" {
			loop = tr.Rich == nil && tr.Layout.Displayed == "[[LOOP]]"
		}
	}
	if !note || !literal || !center || !loop {
		t.Fatalf("source styled semantics note%t literal%t center%t loop%t", note, literal, center, loop)
	}
}
