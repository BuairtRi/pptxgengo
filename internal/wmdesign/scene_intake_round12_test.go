package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/pptx"
)

func round12Plan(t *testing.T, r *renderer, raw string, ctx SceneContext) *scenePlan {
	t.Helper()
	p, handled, err := r.planIntakeRound12Scene("round12", json.RawMessage(raw), ctx)
	if err != nil || !handled {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	return p
}
func TestRound12FunnelNativeGeometryAndCopy(t *testing.T) {
	r := intakeTestRenderer(t)
	p := round12Plan(t, r, `{"type":"funnel","x":100,"y":100,"w":600,"h":240,"shapeW":300,"neck":0.4,"labelSide":true,"stages":[{"value":50,"label":"Eligible","text":"Ready to invite"},{"value":20,"label":"Booked","text":"Ready to visit","active":true}]}`, SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/stages": {"eligible", "booked"}}})
	first := enhancementShape(t, p, ".eligible.band")
	last := enhancementShape(t, p, ".booked.band")
	if first.Type != pptx.ShapeTypeCustGeom || len(first.Props.Points) != 5 || first.Record.Rect != (Rect{100, 100, 300, 117}) || first.Record.Color != "E8EEF8" {
		t.Fatalf("first band=%+v", first)
	}
	if last.Record.Rect.W != 210 || last.Record.Color != "F900D3" {
		t.Fatal("neck or active band differs from source")
	}
	leader := enhancementShape(t, p, ".eligible.leader")
	if leader.Props.Line.DashType != "dot" || leader.Record.Rect.X != 383.5 || leader.Record.Rect.W != 28.5 {
		t.Fatalf("source leader=%+v", leader.Record)
	}
	texts := 0
	seenText := false
	for _, it := range p.Items {
		if it.Text != nil {
			seenText = true
			texts++
			if !inside(it.Text.Rect, p.Bounds) {
				t.Fatal("copy outside reported bounds")
			}
			if strings.Contains(it.Text.ID, ".inside.") && it.Text.Layout.Font.Family != "IBM Plex Mono" {
				t.Fatal("numeric value not mono")
			}
		}
		if it.Shape != nil && seenText {
			t.Fatal("band or leader paints above copy")
		}
		if it.Image != nil || it.Table != nil || it.Chart != nil {
			t.Fatal("diagram rasterized")
		}
	}
	if texts != 6 {
		t.Fatalf("copy parts=%d", texts)
	}
}
func TestRound12PyramidSyntaxPreservesLegacy(t *testing.T) {
	r := intakeTestRenderer(t)
	legacy := json.RawMessage(`{"type":"pyramid","x":100,"y":100,"w":600,"h":240,"bands":[["Top","First"],["Bottom","Second"]]}`)
	if _, handled, err := r.planIntakeRound12Scene("legacy", legacy, SceneContext{}); err != nil || handled {
		t.Fatal("legacy pyramid intercepted")
	}
	p := round12Plan(t, r, `{"type":"pyramid","x":100,"y":100,"w":600,"h":240,"shapeW":300,"labelSide":true,"levels":[{"label":"Vision","text":"Direction"},{"label":"Delivery","text":"Execution"}]}`, SceneContext{Surface: "light"})
	top := enhancementShape(t, p, "source-001.band")
	bottom := enhancementShape(t, p, "source-002.band")
	if top.Record.Color != "070154" || bottom.Record.Color != "E8EEF8" {
		t.Fatal("pyramid ramp direction")
	}
	pts := top.Props.Points
	if pts[0].X.Val != pts[1].X.Val || top.Record.Rect.W != 150 || bottom.Record.Rect.W != 300 {
		t.Fatal("apex or width differs from source")
	}
}
func TestRound12BracketDirectionsAndExtent(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, orient := range []string{"down", "up", "left", "right"} {
		raw := fmt.Sprintf(`{"type":"bracket","x":300,"y":200,"w":180,"h":90,"orient":%q,"depth":9,"label":"Lifecycle"}`, orient)
		p := round12Plan(t, r, raw, SceneContext{Surface: "light"})
		count := 0
		for _, it := range p.Items {
			if it.Shape != nil {
				count++
				if it.Shape.Props.Line.Width != 1.5 || !inside(it.Shape.Record.Rect, p.Bounds) {
					t.Fatal("stroke style/extent")
				}
			}
		}
		if count != 4 {
			t.Fatal("missing bracket stroke")
		}
		tick := enhancementShape(t, p, "stroke-4").Record.Rect
		switch orient {
		case "down":
			if tick != (Rect{390, 191, 0, 9}) {
				t.Fatal(tick)
			}
		case "up":
			if tick != (Rect{390, 200, 0, 9}) {
				t.Fatal(tick)
			}
		case "right":
			if tick != (Rect{300, 245, 9, 0}) {
				t.Fatal(tick)
			}
		case "left":
			if tick != (Rect{291, 245, 9, 0}) {
				t.Fatal(tick)
			}
		}
	}
}
func TestRound12StrictInputsAndResourceLimits(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, raw := range []string{
		`{"type":"funnel","y":100,"w":300,"h":240,"stages":[{"label":"X"}]}`,
		`{"type":"funnel","x":null,"y":100,"w":300,"h":240,"stages":[{"label":"X"}]}`,
		`{"type":"funnel","x":100,"y":100,"w":300,"h":240,"neck":0,"stages":[{"label":"X"}]}`,
		`{"type":"funnel","x":100,"y":100,"w":300,"h":240,"shapeW":320,"stages":[{"label":"X"}]}`,
		`{"type":"funnel","x":100,"y":100,"w":300,"h":240,"stages":[{"label":"X"}],"levels":[{"label":"Y"}]}`,
		`{"type":"funnel","x":100,"y":100,"w":300,"h":10,"stages":[{"label":"Too tall"}]}`,
		`{"type":"funnel","x":100,"y":100,"w":300,"h":240,"ramp":"invented","stages":[{"label":"X"}]}`,
		`{"type":"funnel","x":100,"y":100,"w":300,"h":240,"valueStyle":"invented","stages":[{"label":"X"}]}`,
		`{"type":"bracket","x":100,"y":100,"w":180,"orient":"diagonal"}`,
		`{"type":"bracket","x":100,"y":100,"w":180,"depth":0}`,
		`{"type":"bracket","x":100,"y":100,"w":180,"h":null,"orient":"left"}`,
		`{"type":"bracket","x":100,"y":100,"w":180,"extra":true}`,
	} {
		if _, handled, err := r.planIntakeRound12Scene("invalid", json.RawMessage(raw), SceneContext{Surface: "light"}); !handled || err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	for _, n := range []round12Bands{
		{Type: "funnel", X: 100, Y: 100, W: 300, H: 240, Stages: make([]round12Band, 13)},
		{Type: "funnel", X: 100, Y: 100, W: 300, H: 240, Stages: []round12Band{{Text: strings.Repeat("X", 8193)}}},
		{Type: "funnel", X: 100, Y: 100, W: math.Inf(1), H: 240, Stages: []round12Band{{Label: "X"}}},
	} {
		if _, err := r.round12Bands("invalid", n, SceneContext{Surface: "light"}); err == nil {
			t.Fatal("resource/finite limit ignored")
		}
	}
	if _, handled, err := r.planIntakeRound12Scene("keys", json.RawMessage(`{"type":"funnel","x":100,"y":100,"w":300,"h":240,"stages":[{"label":"X"}]}`), SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{}}); !handled || err == nil {
		t.Fatal("missing authored array keys accepted")
	}
}
func TestRound12FrozenSourceNodes(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-round12", "source")
	renderer, err := os.ReadFile(filepath.Join(root, "explorations", "components.src.html"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(renderer)) != IntakeRound12RendererSHA256 {
		t.Fatal("frozen renderer drift")
	}
	r := intakeTestRenderer(t)
	counts := map[string]int{}
	for _, family := range []string{"diagrams", "lifecycle"} {
		data, err := os.ReadFile(filepath.Join(root, "templates", "library", family+".json"))
		if err != nil {
			t.Fatal(err)
		}
		expectedHash := map[string]string{"diagrams": "b79cf191621a7cfd9bc1b837201cd34ab001703490bd9738a9cd7732c7d9d90e", "lifecycle": "0c159e90675f86812ec73a897d0a7ace628e025e19bfb2f2a675814905288fe7"}[family]
		if fmt.Sprintf("%x", sha256.Sum256(data)) != expectedHash {
			t.Fatal("frozen family drift")
		}
		var doc struct {
			Templates []struct {
				ID, Variant string
				Slide       struct {
					Body []json.RawMessage `json:"body"`
				} `json:"slide"`
			} `json:"templates"`
		}
		if err = json.Unmarshal(data, &doc); err != nil {
			t.Fatal(err)
		}
		for _, template := range doc.Templates {
			for i, raw := range template.Slide.Body {
				var tag struct {
					Type string `json:"type"`
				}
				json.Unmarshal(raw, &tag)
				if tag.Type != "funnel" && tag.Type != "pyramid" && tag.Type != "bracket" {
					continue
				}
				counts[tag.Type]++
				t.Run(fmt.Sprintf("%s/%s-%s-%d", family, template.ID, template.Variant, i), func(t *testing.T) {
					round12Plan(t, r, string(raw), SceneContext{Surface: "light", Path: fmt.Sprintf("/body/%d", i)})
				})
			}
		}
	}
	if counts["funnel"] != 9 || counts["pyramid"] != 5 || counts["bracket"] != 7 {
		t.Fatalf("fixture counts=%v", counts)
	}
}

func TestRound12NativeEditablePackage(t *testing.T) {
	raw := json.RawMessage(`{"type":"funnel","x":117,"y":126,"w":600,"h":240,"shapeW":300,"labelSide":true,"stages":[{"value":"50","label":"Eligible","text":"Ready to invite"},{"value":"20","label":"Booked","text":"Ready to visit"}]}`)
	bracket := json.RawMessage(`{"type":"bracket","x":200,"y":420,"w":360,"label":"Range"}`)
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "round12", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "funnel", Kind: "scene", Scene: &SceneSpec{Node: raw}}, {ID: "range", Kind: "scene", Scene: &SceneSpec{Node: bracket}}}}}}
	data, report, err := BuildWithEngine(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Slides) != 1 || len(report.Slides[0].Scenes) != 2 || report.PowerPointVerified || report.VisuallyReviewed {
		t.Fatal("missing or overstated qualification report")
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var body []byte
	for _, f := range z.File {
		if f.Name == "ppt/slides/slide1.xml" {
			reader, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			body, err = io.ReadAll(reader)
			reader.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	d := xml.NewDecoder(bytes.NewReader(body))
	for {
		_, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, copy := range []string{"Eligible", "Booked", "Ready to visit", "RANGE", "IBM Plex Mono"} {
		if !bytes.Contains(body, []byte(copy)) {
			t.Fatalf("editable copy/font missing: %s", copy)
		}
	}
	if !bytes.Contains(body, []byte("<a:custGeom>")) || !bytes.Contains(body, []byte("<p:grpSp>")) || bytes.Contains(body, []byte("<p:pic>")) {
		t.Fatal("expected native custom paths and groups, no raster diagrams")
	}
}

func TestRound12StackFootnotesAndLiteralBrackets(t *testing.T) {
	r := intakeTestRenderer(t)
	p := round12Plan(t, r, `{"type":"funnel","x":117,"y":126,"w":600,"h":240,"shapeW":300,"labelSide":true,"stages":[{"value":"50","label":"Eligible","text":"Ready[^1]"},{"value":"20","label":"Booked","text":"[[Literal]]"}]}`, SceneContext{Surface: "light", Notes: []string{"Source"}})
	note, literal := false, false
	for _, it := range p.Items {
		if it.Text == nil {
			continue
		}
		if it.Text.Rich != nil && strings.Contains(it.Text.Layout.Displayed, "Ready") {
			for _, paragraph := range it.Text.Rich.Paragraphs {
				for _, run := range paragraph.Runs {
					note = note || run.BaselineShift > 0 && run.Displayed == "1"
				}
			}
		}
		if it.Text.Layout.Original == "[[Literal]]" {
			literal = it.Text.Rich == nil && it.Text.Layout.Displayed == "[[Literal]]"
		}
	}
	if !note || !literal {
		t.Fatal("source footnote or literal copy semantics missing")
	}
}
