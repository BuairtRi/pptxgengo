package wmdesign

import (
	"encoding/json"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/pptx"
)

func TestNativeEditingBlockPreservesMeasuredCopyAndDensity(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	typography, e := NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	r.typeEngine = typography
	for _, level := range []string{"comfortable", "compact", "dense"} {
		t.Run(level, func(t *testing.T) {
			r.bodyDensity = level
			for _, role := range []string{"body", "small", "label"} {
				raw, e := json.Marshal(map[string]any{"type": "editable-block", "x": 100, "y": 120, "w": 250, "h": 144, "surface": "subtle", "text": "Same measured copy", "style": role, "align": "center"})
				if e != nil {
					t.Fatal(e)
				}
				p, ok, e := r.planNativeEditingScene("unit", raw, SceneContext{Surface: "light"})
				if e != nil || !ok {
					t.Fatal(e)
				}
				old, _, e := r.planDiagramScene("unit", json.RawMessage(strings.Replace(string(raw), "editable-block", "block", 1)), SceneContext{Surface: "light"})
				if e != nil {
					t.Fatal(e)
				}
				if len(p.Items) != 1 || p.Items[0].Text == nil || len(p.Groups) != 0 || len(old.Items) != 2 {
					t.Fatal("not a single semantic unit")
				}
				combined := *p.Items[0].Text
				if combined.NativeShape == nil || combined.NativeShape.Rect != (Rect{100, 120, 250, 144}) {
					t.Fatal("containing rectangle lost")
				}
				var original *TextRecord
				for _, item := range old.Items {
					if item.Text != nil {
						original = item.Text
					}
					if item.Shape != nil && item.Shape.Props.Fill.Color != combined.NativeShape.Fill {
						t.Fatal("fill changed")
					}
				}
				combined.ID, combined.NativeShape = original.ID, nil
				if !reflect.DeepEqual(combined, *original) {
					t.Fatalf("density %s role %s changed measured layout/style", level, role)
				}
			}
			if _, _, e := r.planNativeEditingScene("unit", json.RawMessage(`{"type":"editable-block","w":200,"h":90,"surface":"outline","text":"Copy"}`), SceneContext{Surface: "light"}); e == nil {
				t.Fatal("outlined surface lost its separate border geometry")
			}
		})
	}
}

func TestNativeEditingConnectorExplicitSitesAndReverseGeometry(t *testing.T) {
	r := intakeTestRenderer(t)
	r.editableTargets = map[string]Rect{"input": {100, 120, 200, 90}, "output": {400, 210, 200, 90}}
	for _, pair := range [][2]string{{"right", "left"}, {"bottom", "top"}, {"left", "right"}, {"top", "bottom"}} {
		from, to := nativeEditingEndpoint{"output", pair[0]}, nativeEditingEndpoint{"input", pair[1]}
		raw, _ := json.Marshal(nativeEditingSource{Type: "attached-connector", W: 900, H: 400, From: &from, To: &to, Head: "both", Style: "dashed"})
		p, ok, e := r.planNativeEditingScene("edge", raw, SceneContext{Surface: "light"})
		if e != nil || !ok {
			t.Fatal(e)
		}
		if len(p.Items) != 1 || p.Items[0].Shape == nil || len(p.Groups) != 0 {
			t.Fatal("connector generated decorative fragments")
		}
		shape := p.Items[0].Shape
		a, start, _ := nativeRectangleSite(r.editableTargets["output"], from.Site)
		z, end, _ := nativeRectangleSite(r.editableTargets["input"], to.Site)
		if shape.Type != pptx.ShapeTypeLine || shape.Connection == nil || shape.Connection.Begin.Site != start || shape.Connection.End.Site != end || *shape.Props.FlipH != (a[0] > z[0]) || *shape.Props.FlipV != (a[1] > z[1]) || shape.Props.Line.DashType != "dash" || shape.Props.Line.BeginArrowType != "triangle" || shape.Props.Line.EndArrowType != "triangle" {
			t.Fatal("native endpoint or geometry mismatch", shape)
		}
	}
}

func TestNativeEditingStrictInputs(t *testing.T) {
	r := intakeTestRenderer(t)
	r.editableTargets = map[string]Rect{"input": {100, 120, 200, 90}, "output": {400, 210, 200, 90}}
	for _, raw := range []string{
		`{"type":"editable-block","w":200,"h":90,"text":"Copy","unknown":true}`,
		`{"type":"editable-block","w":200,"h":90,"text":"Copy","head":"end"}`,
		`{"type":"editable-block","w":200,"h":90,"text":"[[Rich copy]]"}`,
		`{"type":"editable-block","w":10,"h":90,"text":"Copy"}`,
		`{"type":"attached-connector","from":{"node":"absent","site":"right"},"to":{"node":"output","site":"left"}}`,
		`{"type":"attached-connector","from":{"node":"input","site":"diagonal"},"to":{"node":"output","site":"left"}}`,
		`{"type":"attached-connector","from":{"node":"input","site":"right","unknown":1},"to":{"node":"output","site":"left"}}`,
		`{"type":"attached-connector","from":{"node":"input","site":"right"},"to":{"node":"input","site":"left"}}`,
		`{"type":"attached-connector","from":{"node":"input","site":"right"},"to":{"node":"output","site":"left"},"head":"magic"}`,
		`{"type":"attached-connector","from":{"node":"input","site":"right"},"to":{"node":"output","site":"left"},"route":"magic"}`,
		`{"type":"attached-connector","from":{"node":"input","site":"right"},"to":{"node":"output","site":"left"},"route":"horizontal","bend":2}`,
		`{"type":"attached-connector","from":{"node":"input","site":"right"},"to":{"node":"output","site":"left"},"bend":0.5}`,
	} {
		if _, _, e := r.planNativeEditingScene("unit", json.RawMessage(raw), SceneContext{Surface: "light"}); e == nil {
			t.Fatalf("invalid input accepted: %s", raw)
		}
	}
}

func TestNativeEditingElbowEndpointsAllDirections(t *testing.T) {
	r := intakeTestRenderer(t)
	r.editableTargets = map[string]Rect{"input": {100, 120, 200, 90}, "output": {400, 210, 200, 90}}
	for _, mode := range []string{"horizontal", "vertical"} {
		for _, outputY := range []float64{50, 120, 210} {
			r.editableTargets["output"] = Rect{400, outputY, 200, 90}
			for _, reverse := range []bool{false, true} {
				for _, pair := range [][2]string{{"right", "left"}, {"bottom", "top"}, {"left", "right"}, {"top", "bottom"}} {
					a, z := "input", "output"
					if reverse {
						a, z = z, a
					}
					from, to := nativeEditingEndpoint{a, pair[0]}, nativeEditingEndpoint{z, pair[1]}
					raw, _ := json.Marshal(nativeEditingSource{Type: "attached-connector", W: 900, H: 400, From: &from, To: &to, Route: mode})
					plan, _, e := r.planNativeEditingScene("edge", raw, SceneContext{Surface: "light"})
					if e != nil {
						t.Fatal(e)
					}
					shape := plan.Items[0].Shape
					props := shape.Props
					x, y, w, h := props.X.Val*72, props.Y.Val*72, props.W.Val*72, props.H.Val*72
					cx, cy := x+w/2, y+h/2
					sx, sy := 1., 1.
					if *props.FlipH {
						sx = -1
					}
					if *props.FlipV {
						sy = -1
					}
					theta := props.Rotate * math.Pi / 180
					transform := func(px, py float64) [2]float64 {
						dx, dy := (px-cx)*sx, (py-cy)*sy
						return [2]float64{cx + math.Cos(theta)*dx - math.Sin(theta)*dy, cy + math.Sin(theta)*dx + math.Cos(theta)*dy}
					}
					expectedA, _, _ := nativeRectangleSite(r.editableTargets[a], pair[0])
					expectedZ, _, _ := nativeRectangleSite(r.editableTargets[z], pair[1])
					actualA, actualZ := transform(x, y), transform(x+w, y+h)
					if math.Hypot(actualA[0]-expectedA[0], actualA[1]-expectedA[1]) > 1e-8 || math.Hypot(actualZ[0]-expectedZ[0], actualZ[1]-expectedZ[1]) > 1e-8 || shape.Route == nil || shape.Route.Adjustment != 50000 {
						t.Fatal("elbow endpoints changed", mode, reverse, pair)
					}
				}
			}
		}
	}
}
