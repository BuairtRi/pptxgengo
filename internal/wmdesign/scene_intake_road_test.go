package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func roadPlan(t *testing.T, r *renderer, raw string, ctx SceneContext) *scenePlan {
	t.Helper()
	p, handled, err := r.planIntakeRoadScene("road", json.RawMessage(raw), ctx)
	if !handled || err != nil {
		t.Fatalf("road handled=%v err=%v", handled, err)
	}
	return p
}
func TestRound12RoadSourcePathMirrorPinsAndBounds(t *testing.T) {
	r := intakeTestRenderer(t)
	base := `{"type":"road","x":100,"y":100,"w":600,"h":300,"amp":0.16,"waves":1.5,"labelW":180,"milestones":[{"date":"Now","label":"First","text":"Start here","at":0,"active":true},{"label":"Last","at":1,"side":"above","text":"Finish here"}]}`
	right := roadPlan(t, r, base, SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/milestones": {"start", "finish"}}})
	left := roadPlan(t, r, strings.Replace(base, `"labelW":180`, `"direction":"left","labelW":180`, 1), SceneContext{Surface: "light"})
	a := enhancementShape(t, right, ".road")
	b := enhancementShape(t, left, ".road")
	if len(a.Props.Points) != 121 || len(b.Props.Points) != 121 || a.Props.Line.Width != 22 || a.Props.Line.LineJoin != "round" {
		t.Fatal("source121point stroke differs")
	}
	for i := range a.Props.Points {
		x := a.Props.X.Val*72 + a.Props.Points[i].X.Val*72
		y := a.Props.Y.Val*72 + a.Props.Points[i].Y.Val*72
		xx := b.Props.X.Val*72 + b.Props.Points[i].X.Val*72
		yy := b.Props.Y.Val*72 + b.Props.Points[i].Y.Val*72
		if math.Abs(x+xx-800) > 1e-8 || math.Abs(y-yy) > 1e-8 {
			t.Fatal("road direction failed to mirror path")
		}
	}
	pin := enhancementShape(t, right, ".start.pin")
	if pin.Record.Color != "F900D3" || pin.Record.Rect.W != 26 || pin.Props.W.Val*72 != 24 {
		t.Fatal("active pin or inset border")
	}
	dash := enhancementShape(t, right, ".centerline")
	if dash.Props.Line.Width != 1.5 || dash.Record.Color != "FFFFFF" {
		t.Fatal("white road centerline missing")
	}
	for _, it := range right.Items {
		if it.Shape != nil && !inside(it.Shape.Record.Rect, right.Bounds) || it.Text != nil && !inside(it.Text.Rect, right.Bounds) {
			t.Fatal("actual road ink outside reported bounds")
		}
	}
	if right.Bounds.Y >= 100 {
		t.Fatal("above labels not included in actual bounds")
	}
}
func TestRound12RoadExactDashSegments(t *testing.T) {
	r := intakeTestRenderer(t)
	p := &scenePlan{}
	if err := r.roadCenterline(p, "dash", []curvePoint{{100, 100}, {110, 100}, {120, 100}, {140, 100}}); err != nil {
		t.Fatal(err)
	}
	path := p.Items[0].Shape.Props.Points
	expected := []struct {
		x    float64
		move bool
	}{{0, true}, {7, false}, {14, true}, {20, false}, {21, false}, {28, true}, {35, false}}
	if len(path) != len(expected) {
		t.Fatalf("dash path points=%d", len(path))
	}
	for i, want := range expected {
		if *path[i].MoveTo != want.move || math.Abs(path[i].X.Val*72-want.x) > 1e-8 {
			t.Fatal("dash spacing/continuity failed across source vertices")
		}
	}
}
func TestRound12RoadStrictInputsAndCopy(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, raw := range []string{
		`{"type":"road","y":100,"w":600,"h":300,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":null,"y":100,"w":600,"h":300,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":36,"h":300,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":80,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":300,"amp":2,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":300,"waves":0,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":300,"roadW":0,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":300,"labelW":700,"milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":300,"direction":"diagonal","milestones":[{"label":"X"}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":300,"milestones":[{"label":"X","at":1.1}]}`,
		`{"type":"road","x":100,"y":100,"w":600,"h":300,"milestones":[{"label":"X","side":"left"}]}`,
	} {
		if _, handled, err := r.planIntakeRoadScene("bad", json.RawMessage(raw), SceneContext{Surface: "light"}); !handled || err == nil {
			t.Fatalf("accepted invalid%s", raw)
		}
	}
	n := round12Road{Type: "road", X: 100, Y: 100, W: 600, H: 300, Milestones: make([]round12RoadMilestone, 25)}
	if _, err := r.round12Road("bad", n, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("milestone limit")
	}
	n.Milestones = []round12RoadMilestone{{Text: strings.Repeat("X", 8193)}}
	if _, err := r.round12Road("bad", n, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("copy limit")
	}
	n.Milestones = []round12RoadMilestone{{Label: "X"}}
	v := math.NaN()
	n.Amp = &v
	if _, err := r.round12Road("bad", n, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("nonfinite option")
	}
	if err := r.roadCenterline(&scenePlan{}, "long", []curvePoint{{0, 0}, {100000, 0}}); err == nil {
		t.Fatal("dash geometry resource limit")
	}
}
func TestRound12RoadFrozenSixSourceNodes(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-round12", "source", "templates", "library", "diagrams.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Templates []struct {
			Variant string
			Slide   struct {
				Body []json.RawMessage `json:"body"`
			} `json:"slide"`
		} `json:"templates"`
	}
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	r := intakeTestRenderer(t)
	count := 0
	for _, template := range doc.Templates {
		for _, raw := range template.Slide.Body {
			var tag struct {
				Type string `json:"type"`
			}
			json.Unmarshal(raw, &tag)
			if tag.Type != "road" {
				continue
			}
			count++
			t.Run(template.Variant, func(t *testing.T) {
				p := roadPlan(t, r, string(raw), SceneContext{Surface: "light"})
				if len(enhancementShape(t, p, ".road").Props.Points) != 121 {
					t.Fatal("source path samples")
				}
			})
		}
	}
	if count != 6 {
		t.Fatal(fmt.Sprintf("road fixtures=%d", count))
	}
}

func TestRound12RoadNativeEditablePackage(t *testing.T) {
	raw := json.RawMessage(`{"type":"road","x":117,"y":126,"w":700,"h":280,"labelW":150,"milestones":[{"label":"Start","date":"Now","side":"above","text":"Prepare"},{"label":"Finish","side":"below","text":"Deliver","active":true}]}`)
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "road", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "road", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
	data, report, err := BuildWithEngine(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Slides) != 1 || report.PowerPointVerified || report.VisuallyReviewed {
		t.Fatal("missing/overstated report")
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
	if !bytes.Contains(body, []byte("<a:round/>")) || !bytes.Contains(body, []byte("<a:custGeom>")) || !bytes.Contains(body, []byte("<p:grpSp>")) || bytes.Contains(body, []byte("<p:pic>")) || !bytes.Contains(body, []byte("F900D3")) || !bytes.Contains(body, []byte("Finish")) {
		t.Fatal("native editable road/pins/copy missing")
	}
}
