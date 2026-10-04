package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

// Embedded specimens freeze the 17:50 UTC intake snapshot nodes independently of mutable source files.
const maturityWorkingExamples = `[{"variant":"four-stage","node":{"type":"maturity","x":57,"y":126,"w":846,"h":342,"stages":[{"label":"Ad hoc","text":"Work depends on individuals and varies by team."},{"label":"Defined","text":"Standard methods are written down and taught."},{"label":"Managed","text":"Teams measure results and correct course."},{"label":"Optimized","text":"Data drives continuous, shared improvement."}],"at":[0.06,0.38,0.66,0.88],"axisLabel":"Increasing maturity","labelW":198}},{"variant":"five-active","node":{"type":"maturity","x":57,"y":126,"w":846,"h":306,"stages":[{"label":"Initial","text":"Effort is improvised and results vary."},{"label":"Repeatable","text":"Teams follow the same steps for similar work."},{"label":"Defined","text":"One standard method is used across teams."},{"label":"Managed","text":"Measures guide decisions and fixes."},{"label":"Optimizing","text":"Learning loops improve the method itself."}],"at":[0.05,0.26,0.47,0.66,0.84],"active":1,"axisLabel":"Increasing maturity","labelW":162,"here":{"label":"You are here"}}},{"variant":"split","node":{"type":"maturity","x":345,"y":36,"w":558,"h":414,"stages":[{"label":"Ad hoc","text":"Varies by person."},{"label":"Defined","text":"Written standards."},{"label":"Managed","text":"Measured results."},{"label":"Optimized","text":"Shared learning."}],"at":[0.05,0.34,0.6,0.84],"axisLabel":"Maturity","labelW":108,"active":1}},{"variant":"nav","node":{"type":"maturity","x":57,"y":126,"w":846,"h":324,"stages":[{"label":"Ad hoc","text":"Work depends on individuals and varies by team."},{"label":"Defined","text":"Standard methods are written down and taught."},{"label":"Managed","text":"Teams measure results and correct course."},{"label":"Optimized","text":"Data drives continuous, shared improvement."}],"at":[0.06,0.38,0.66,0.88],"axisLabel":"Increasing maturity","labelW":180,"active":1}},{"variant":"table","node":{"type":"maturity","x":57,"y":126,"w":846,"h":180,"stages":[{"label":"Ad hoc"},{"label":"Defined"},{"label":"Managed"},{"label":"Optimized"}],"at":[0.08,0.32,0.56,0.76],"axis":false,"labelW":126}},{"variant":"ai-simple","node":{"type":"maturity","x":57,"y":126,"w":846,"h":342,"stages":[{"label":"Manual","text":"People do the work end to end; knowledge stays fragmented."},{"label":"Individually AI-assisted","text":"AI helps with bounded tasks; workflows stay unchanged."},{"label":"Repeatable team practice","text":"Teams reuse shared workflows with clear owners and defined review."},{"label":"Integrated AI-native flow","text":"People and agents coordinate across the lifecycle with clear handoffs."},{"label":"Self-improving delivery system","text":"Telemetry feeds agents that refine shared skills, so delivery improves on its own."}],"at":[0.05,0.26,0.48,0.67,0.84],"axisLabel":"Increasing scope","labelW":180}},{"variant":"ai-beyond","node":{"type":"maturity","x":57,"y":126,"w":846,"h":342,"stages":[{"label":"Manual","text":"People do the work end to end; knowledge stays fragmented."},{"label":"Individually AI-assisted","text":"AI helps with bounded tasks; workflows stay unchanged."},{"label":"Repeatable team practice","text":"Teams reuse shared workflows with clear owners and defined review."},{"label":"Integrated AI-native flow","text":"People and agents coordinate across the lifecycle with clear handoffs."},{"label":"Self-improving delivery system","text":"Telemetry feeds agents that refine shared skills, so delivery improves on its own."}],"at":[0.05,0.24,0.44,0.64,0.84],"axisLabel":"Increasing scope","labelW":162,"inflection":3,"inflectionLabel":"Inflection","branch":{"from":3,"n":"4.5","label":"Expanding beyond product delivery","text":"Integrated flow extends to support, go-to-market and marketing."}}},{"variant":"ai-capabilities","node":{"type":"maturity","x":57,"y":126,"w":846,"h":198,"stages":[{"label":"Manual"},{"label":"Individually AI-assisted"},{"label":"Repeatable team practice"},{"label":"Integrated AI-native flow"},{"label":"Self-improving delivery system"}],"at":[0.06,0.28,0.48,0.66,0.78],"axis":false,"labelW":190}}]`
const maturityWorkingExamplesSHA = "e89e64cef94a5ddc9eb4535d9534d950bd64c9a6c425052e7c719ed52dce84bf"

// The additive here example was observed in the same frozen renderer contract.
const maturityHereExample = `{"type":"maturity","x":57,"y":126,"w":846,"h":306,"stages":[{"label":"Initial","text":"Effort is improvised and results vary."},{"label":"Repeatable","text":"Teams follow the same steps for similar work."},{"label":"Defined","text":"One standard method is used across teams."},{"label":"Managed","text":"Measures guide decisions and fixes."},{"label":"Optimizing","text":"Learning loops improve the method itself."}],"at":[0.05,0.26,0.47,0.66,0.84],"active":1,"axisLabel":"Increasing maturity","labelW":162,"here":{"label":"You are here"}}`
const maturityHereSourceSHA = "63eca0027e5797a19be2b1236986772960889f2bb4ad1bb18a93a08416145c5b"
const maturityHereRendererSHA = "a7f1f046d38723cb86d6e0dbacd97886c183784df1e4949192322e74d6a33cf6"

func maturityTestPlan(t *testing.T, r *renderer, raw string, ctx SceneContext) *scenePlan {
	t.Helper()
	p, ok, e := r.planIntakeMaturityScene("maturity", json.RawMessage(raw), ctx)
	if e != nil || !ok {
		t.Fatalf("recognized=%t err=%v", ok, e)
	}
	return p
}
func TestIntakeMaturityFrozenWorkingExamples(t *testing.T) {
	if fmt.Sprintf("%x", sha256.Sum256([]byte(maturityWorkingExamples))) != maturityWorkingExamplesSHA {
		t.Fatal("frozen example changed")
	}
	var fixtures []struct {
		Variant string          `json:"variant"`
		Node    json.RawMessage `json:"node"`
	}
	if e := json.Unmarshal([]byte(maturityWorkingExamples), &fixtures); e != nil {
		t.Fatal(e)
	}
	r := intakeTestRenderer(t)
	for _, f := range fixtures {
		t.Run(f.Variant, func(t *testing.T) {
			p := maturityTestPlan(t, r, string(f.Node), SceneContext{Surface: "light"})
			if len(p.Groups) != 1 || p.Groups[0].Contract != IntakeMaturityContract {
				t.Fatal("native grouping/contract missing")
			}
			if e := sceneTextEnvelope(p, SceneContext{Zone: Rect{0, 0, 960, 486}}); e != nil {
				t.Fatal(e)
			}
			t.Logf("actualsourceextent %+v", p.Bounds)
		})
	}
	if len(fixtures) != 8 {
		t.Fatalf("fixtures %d", len(fixtures))
	}
}
func TestIntakeMaturityGeometryBranchAndAlignment(t *testing.T) {
	r := intakeTestRenderer(t)
	p := maturityTestPlan(t, r, `{"type":"maturity","x":100,"y":100,"w":600,"h":330,"stages":[{"label":"One"},{"label":"Two"},{"label":"Three"},{"label":"Four"}],"at":[0.1,0.3,0.6,0.8],"active":1,"inflection":2,"branch":{"from":2,"n":"3.5","label":"Beyond","text":"A flatter path."},"axisLabel":"Scope"}`, SceneContext{Surface: "light"})
	var curve, branch, active *sceneShape
	var first, last, num *TextRecord
	for _, it := range p.Items {
		if it.Shape != nil {
			switch it.Shape.Record.ID {
			case "maturity.curve":
				curve = it.Shape
			case "maturity.branch.curve":
				branch = it.Shape
			case "maturity.stages.source-002.circle":
				active = it.Shape
			}
		}
		if it.Text != nil {
			switch it.Text.ID {
			case "maturity.stages.source-001.label":
				first = it.Text
			case "maturity.stages.source-004.label":
				last = it.Text
			case "maturity.branch.number":
				num = it.Text
			}
		}
	}
	if curve == nil || len(curve.Props.Points) != 61 || branch == nil || branch.Props.Points[1].Curve.Type != "quadratic" {
		t.Fatal("native61pointcurve/quad missing")
	}
	if first.Align != "left" || last.Align != "right" || active.Props.Fill.Color != "070154" || num.Layout.Style.Size != 9 {
		t.Fatal("alignment/markerstyle changed")
	}
	q := maturityPoint(maturitySource{X: 100, Y: 100, W: 600, H: 330}, 294, 4.2, .6)
	endY := q.y - math.Max(24, (394-q.y)*.25)
	x, y := q.x+(700-q.x)*.72, q.y+(endY-q.y)*.72-4
	if math.Abs(num.Rect.X+10-x) > 1e-8 || math.Abs(num.Rect.Y+num.Rect.H/2-y) > 1e-8 {
		t.Fatalf("72%%branchmarker anchor %+v", num.Rect)
	}
}
func TestIntakeMaturityStrictFieldsAndKeys(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"maturity","x":100,"y":100,"w":500,"h":300,"stages":[{"label":"One"},{"label":"Two"}]}`
	for _, extra := range []string{`"shape":0`, `"shape":1e200`, `"active":2`, `"active":1.5`, `"inflection":-1`, `"axis":"false"`, `"at":[0.8,0.2]`, `"at":[0.3]`, `"labelW":0`, `"branch":{"label":"No index"}`, `"branch":{"from":9,"label":"Wrong index"}`, `"here":{"stage":9}`, `"unknown":1`} {
		if _, _, e := r.planIntakeMaturityScene("m", json.RawMessage(strings.TrimSuffix(raw, "}")+","+extra+"}"), SceneContext{Surface: "light"}); e == nil {
			t.Fatalf("invalid accepted %s", extra)
		}
	}
	for _, n := range []string{`{"type":"maturity","w":1e200,"h":300,"stages":[{"label":"One"}]}`, `{"type":"maturity","w":500,"h":30,"stages":[{"label":"One"}]}`, `{"type":"maturity","w":500,"h":300,"stages":[{"label":"One","n":true}]}`, `{"type":"maturity","w":500,"h":300,"stages":[{"label":"One","unknown":1}]}`} {
		if _, _, e := r.planIntakeMaturityScene("m", json.RawMessage(n), SceneContext{Surface: "light"}); e == nil {
			t.Fatalf("invalid nested/geometry accepted %s", n)
		}
	}
	for _, keys := range []map[string][]string{{}, {"/body/0/stages": {"one"}}, {"/body/0/stages": {"one", "one"}}} {
		if _, _, e := r.planIntakeMaturityScene("m", json.RawMessage(raw), SceneContext{Surface: "light", Path: "/body/0", Keys: keys}); e == nil {
			t.Fatal("invalid keys accepted")
		}
	}
	p := maturityTestPlan(t, r, raw, SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/stages": {"begin", "advance"}}})
	found := false
	for _, it := range p.Items {
		if it.Text != nil && it.Text.ID == "maturity.stages.advance.label" {
			found = true
		}
	}
	if !found {
		t.Fatal("stable key missing")
	}
}
func TestIntakeMaturitySingleAxisFalseAndHere(t *testing.T) {
	r := intakeTestRenderer(t)
	p := maturityTestPlan(t, r, `{"type":"maturity","x":100,"y":100,"w":500,"h":300,"axis":false,"shape":1e-12,"stages":[{"label":"One","n":0}]}`, SceneContext{Surface: "light"})
	for _, it := range p.Items {
		if it.Shape != nil && strings.Contains(it.Shape.Record.ID, ".axis") {
			t.Fatal("axisfalse emitsaxis")
		}
	}
	if !intakeFinite(p.Bounds.X, p.Bounds.Y, p.Bounds.W, p.Bounds.H) {
		t.Fatal("smallshape invalid")
	}
	p = maturityTestPlan(t, r, maturityHereExample, SceneContext{Surface: "light"})
	found := false
	for _, it := range p.Items {
		if it.Text != nil && it.Text.ID == "maturity.here.label" {
			found = it.Text.Layout.Original == "You are here"
		}
	}
	if !found {
		t.Fatal("additiveheretag missing")
	}
}
func TestIntakeMaturityNativeEditablePackage(t *testing.T) {
	raw := json.RawMessage(`{"type":"maturity","x":100,"y":100,"w":600,"h":330,"stages":[{"label":"Start"},{"label":"Advance"},{"label":"Scale"}],"at":[0.1,0.4,0.7],"branch":{"from":1,"label":"Beyond","n":"2.5"}}`)
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "maturity", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "maturity", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
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
	if !strings.Contains(xml, "<a:quadBezTo>") || !strings.Contains(xml, "<p:grpSp>") || strings.Contains(xml, "<p:pic>") || !strings.Contains(xml, "maturity.stages.source-001.number") || strings.Count(xml, "<a:lnTo>") < 60 {
		t.Fatal("nativegroup/text/geometry missing orrasterized")
	}
}

func TestIntakeMaturityHeadroomAndCopyBounds(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"maturity","x":100,"y":100,"w":500,"h":300,"headroom":80,"stages":[{"label":"One"},{"label":"Two"}],"at":[0.2,0.7],"branch":{"from":0,"label":"Beyond"}}`
	p := maturityTestPlan(t, r, raw, SceneContext{Surface: "light"})
	headroom := 80.
	q := maturityPoint(maturitySource{X: 100, Y: 100, W: 500, H: 300, Headroom: &headroom}, 184, 4.2, .2)
	found := false
	for _, it := range p.Items {
		if it.Text != nil && it.Text.ID == "maturity.stages.source-001.number" {
			found = true
			if math.Abs(it.Text.Rect.Y+it.Text.Rect.H/2-q.y) > 1e-8 {
				t.Fatal("headroom missing from stage")
			}
		}
		if it.Shape != nil && it.Shape.Record.ID == "maturity.branch.curve" {
			wantY := q.y - math.Max(24, (364-q.y)*.25)
			if math.Abs(it.Shape.Record.Rect.Y+1-wantY) > 1e-8 {
				t.Fatal("headroom missing from branch")
			}
		}
	}
	if !found {
		t.Fatal("headroom example missing")
	}
	for _, field := range []string{`"headroom":-1`, `"headroom":264`, `"headroom":1e300`, `"axisLabel":"` + strings.Repeat("a", 4097) + `"`, `"inflectionLabel":"` + strings.Repeat("a", 4097) + `"`, `"here":{"label":"` + strings.Repeat("a", 4097) + `"}`, `"branch":{"from":0,"label":"` + strings.Repeat("a", 4097) + `"}`} {
		base := `{"type":"maturity","x":100,"y":100,"w":500,"h":300,"stages":[{"label":"One"}]}`
		if _, _, e := r.planIntakeMaturityScene("m", json.RawMessage(strings.TrimSuffix(base, "}")+","+field+"}"), SceneContext{Surface: "light"}); e == nil {
			t.Fatalf("invalid field accepted %s", field[:min(len(field), 50)])
		}
	}
}

func TestIntakeMaturityRequiresNumericGeometry(t *testing.T) {
	r := intakeTestRenderer(t)
	base := map[string]any{"type": "maturity", "x": 0, "y": 0, "w": 500, "h": 300, "stages": []any{map[string]any{"label": "One"}}}
	raw, _ := json.Marshal(base)
	maturityTestPlan(t, r, string(raw), SceneContext{Surface: "light"})
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
			if _, _, e := r.planIntakeMaturityScene("maturity", raw, SceneContext{Surface: "light"}); e == nil {
				t.Fatalf("geometry accepted: %s=%v", field, value)
			}
		}
	}
}
