package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRound12GaussianEdges(t *testing.T) {
	at, values := []float64{0, .5, 1}, []float64{0, 10, 0}
	hi, lo, next := teamCurveGaussianEdges(at, values, make([]float64, 241), .07, false)
	if len(hi) != 241 || hi[120] >= 10 || hi[120] < 8 || hi[0] <= 0 || hi[0] > 1 {
		t.Fatalf("unexpected smoothed peaks/ends: %g %g", hi[120], hi[0])
	}
	for i := range hi {
		if lo[i] != 0 || math.Abs(hi[i]-hi[240-i]) > 1e-12 {
			t.Fatalf("bad base or symmetry at %d", i)
		}
	}
	line, lineLo, unchanged := teamCurveGaussianEdges(at, []float64{1, 1, 1}, next, .07, true)
	for i := range line {
		if unchanged[i] != next[i] || math.Abs(line[i]-lineLo[i]-1) > 1e-12 {
			t.Fatalf("line changed stack or height at %d", i)
		}
	}
	for _, v := range teamCurveGaussianBlur(values, 1e-320) {
		if !intakeFinite(v) {
			t.Fatal("subnormal blur produced nonfinite value")
		}
	}
}

func TestRound12TableScalarNumbersAndHeader(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"table","x":100,"y":150,"w":400,"rowHeader":"name","cols":[{"k":"name","label":"Name","w":200},{"k":"count","label":"Count","w":200,"type":"num"}],"rows":[{"name":"A","count":123},{"name":"B","count":-2.5}]}`)
	p, ok, err := r.planTableScene("table", raw, SceneContext{Surface: "light"})
	if err != nil || !ok {
		t.Fatal(err)
	}
	for _, it := range p.Items {
		if it.Table != nil {
			seen := map[string]bool{}
			for _, cell := range it.Table.CellRecords {
				seen[cell.Text.Layout.Displayed] = true
			}
			if !seen["123"] || !seen["-2.5"] {
				t.Fatalf("numeric copy missing: %+v", seen)
			}
		}
	}
}

func TestRound13LineSamples(t *testing.T) {
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source", "templates", "library"))
	r := intakeTestRenderer(t)
	for _, key := range []string{"value-curve/break-even", "value-curve/break-even-split", "value-curve/scenarios"} {
		obj, err := libraryObject(entries[key].Slide)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(obj["body"].([]any)[0])
		r.source.Revision = LibraryRevisionV3
		if _, _, err = r.planChartScene("chart", raw, SceneContext{Surface: "light"}); err == nil {
			t.Fatal("changed frozen v3 category limit")
		}
		r.source.Revision = IntakeTeamCurveMonotoneRevision
		p, ok, err := r.planChartScene("chart", raw, SceneContext{Surface: "light"})
		if err != nil || !ok {
			t.Fatal(err)
		}
		found := false
		for _, it := range p.Items {
			if it.Chart != nil {
				found = true
				if len(it.Chart.Data[0].Values) != 20 {
					t.Fatal("curve samples lost")
				}
			}
		}
		if !found {
			t.Fatal("native chart missing")
		}
	}
}

func TestRound12SlimChromeAndTint(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	s, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	frames, err := os.ReadFile(filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-round12", "source", "frames", "v0", "frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(frames, &s.Frames); err != nil {
		t.Fatal(err)
	}
	s.Revision = IntakeTeamCurveMonotoneRevision
	slide, err := compileLibrarySlide(json.RawMessage(`{"type":"slide","rail":"none","footer":"slim","title":"Title","eyebrow":"Context","tint":{"x":600,"w":360},"body":[{"type":"text","x":57,"y":126,"w":500,"text":"Editable","style":"body"}]}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	slide.ID = "slim-chrome"
	_, report, err := buildWithLoadedSource(bundle, s, Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}, CandidateEngine, nil)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, text := range report.Slides[0].Texts {
		if text.ID == "wm.page" || text.ID == "wm.legal" {
			if math.Abs(text.Rect.Y+text.Rect.H/2-510) > .01 {
				t.Fatalf("footer not centered: %+v", text.Rect)
			}
			seen[text.ID] = true
		}
	}
	if !seen["wm.page"] || !seen["wm.legal"] {
		t.Fatal("footer copy missing")
	}
}

func TestRound12TintLayoutIdentityAndDirectValidation(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	s, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	base := SlideSpec{ID: "plain", Title: "Title", Eyebrow: "Context", Frame: FrameRequest{Rail: "none", Footer: "compact", Surface: "light", TitleLines: 1}}
	a, b := base, base
	a.ID, b.ID = "tint-a", "tint-b"
	a.LibraryChrome = &LibraryChrome{Tint: []LibraryTint{{X: 600, W: 360, Surface: "subtle"}}}
	b.LibraryChrome = &LibraryChrome{Tint: []LibraryTint{{X: 500, W: 460, Surface: "light"}}}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{base, a, b}}
	raw, _, err := buildWithLoadedSource(bundle, s, doc, CandidateEngine, nil)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	layouts := 0
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "ppt/slideLayouts/slideLayout") && strings.HasSuffix(f.Name, ".xml") {
			layouts++
		}
	}
	if layouts != 4 {
		t.Fatalf("distinct tints shared frame layout: %d", layouts)
	}
	for _, tints := range [][]LibraryTint{{{X: -1, W: 360}}, {{X: 600, W: 361}}, {{X: 0, W: 0}}, {{X: 600, W: 360, Surface: "unknown"}}, make([]LibraryTint, 9)} {
		a.LibraryChrome = &LibraryChrome{Tint: tints}
		doc.Slides = []SlideSpec{a}
		if _, _, err := buildWithLoadedSource(bundle, s, doc, CandidateEngine, nil); err == nil {
			t.Fatalf("invalid direct tint accepted: %+v", tints)
		}
	}
}

func TestRound12CurveDefaultsAndOverrides(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"teamcurve","x":100,"y":80,"w":500,"h":200,"phaseH":0,"series":[{"values":[0,10,0],"labelAt":false}]}`
	p := intakeCurvePlan(t, r, raw, SceneContext{Surface: "light"})
	if p.Items[0].Shape.Record.Geometry != "cubic-stacked-band" {
		t.Fatal("changed v3 default")
	}
	r.source.Revision = IntakeTeamCurveMonotoneRevision
	p = intakeCurvePlan(t, r, raw, SceneContext{Surface: "light"})
	if p.Items[0].Shape.Record.Geometry != "gaussian-stacked-band" || len(p.Items[0].Shape.Props.Points) != 483 {
		t.Fatal("missing editable gaussian band")
	}
	p = intakeCurvePlan(t, r, strings.Replace(raw, `"phaseH":0`, `"smooth":0,"phaseH":0`, 1), SceneContext{Surface: "light"})
	if p.Items[0].Shape.Props.Points[1].Curve == nil {
		t.Fatal("smooth zero did not restore cubic")
	}
	for _, smooth := range []string{"-1", "1.1", "null"} {
		x := strings.Replace(raw, `"phaseH":0`, `"smooth":`+smooth+`,"phaseH":0`, 1)
		_, _, err := r.planIntakeCurveScene("curve", json.RawMessage(x), SceneContext{Surface: "light"})
		if smooth != "null" && err == nil {
			t.Fatal("invalid smooth accepted")
		}
	}
}

func TestRound12VennExclusiveAnchorAndType(t *testing.T) {
	n := intakeVennSource{X: 0, Y: 0, W: 540, H: 396, Sets: make([]intakeVennSet, 3)}
	radius, centers := intakeVennGeometry(n)
	a := intakeVennRound12Anchor(centers, []int{0, 1}, [2]float64{270, 198}, radius)
	for i, c := range centers {
		if (math.Hypot(a[0]-c[0], a[1]-c[1]) <= radius) != (i < 2) {
			t.Fatalf("anchor outside exclusive lens: %+v", a)
		}
	}
	r := intakeTestRenderer(t)
	r.source.Revision = IntakeTeamCurveMonotoneRevision
	raw := json.RawMessage(`{"type":"venn","x":100,"y":100,"w":540,"h":396,"sets":[{"label":"First","dx":2,"dy":3},{"label":"Second"},{"label":"Third"}],"regions":[{"in":[0,1],"label":"Overlap"},{"in":[0,1,2],"label":"Core"}]}`)
	p, ok, err := r.planIntakeVennScene("venn", raw, SceneContext{Surface: "light"})
	if !ok || err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, it := range p.Items {
		if it.Text != nil && strings.HasSuffix(it.Text.ID, ".label") {
			want := 13.
			if strings.Contains(it.Text.ID, "regions.") && it.Text.Layout.Displayed == "OVERLAP" {
				want = 11
			}
			if it.Text.Layout.Style.Size != want {
				t.Fatalf("unexpected font size: %+v", it.Text.Layout.Style)
			}
			seen++
		}
	}
	if seen != 5 {
		t.Fatalf("labels=%d", seen)
	}
}

func TestRound12StatusAndInverseCallout(t *testing.T) {
	r := intakeTestRenderer(t)
	for key, spec := range sceneStatuses {
		p := &scenePlan{}
		if err := r.sceneTableStatus(p, "status", Rect{0, 0, 8, 8}, "light", key); err != nil {
			t.Fatal(err)
		}
		if len(p.Items) != 1 || spec.Label == "" {
			t.Fatal("bad native status")
		}
	}
	p, err := r.sceneDirectMetric("callout", json.RawMessage(`{"type":"callout","x":100,"y":100,"w":300,"label":"Outcome","value":"5","text":"days","surface":"inverse"}`), SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Items[0].Shape.Record.Color != "070154" {
		t.Fatal("inverse callout surface ignored")
	}
}
