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

const frozenGaugeNode = `{"type":"gauge","x":687,"y":144,"w":216,"segments":["kpi.off","kpi.off","kpi.risk","kpi.risk","kpi.on","kpi.on"],"value":0.58,"valueText":"2.6","caption":"Overall readiness (of 4)","ends":["Not ready","Ready"]}`

func gaugePlan(t *testing.T, r *renderer, raw string, ctx SceneContext) *scenePlan {
	t.Helper()
	p, handled, err := r.planIntakeGaugeScene("gauge", json.RawMessage(raw), ctx)
	if err != nil || !handled {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	return p
}
func TestIntakeGaugeFrozenGeometryAndSource(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	for _, file := range []struct{ path, hash string }{{"explorations/components.src.html", IntakeGaugeRendererSHA256}, {"templates/library/change.json", IntakeGaugeSourceSHA256}} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.path)))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != file.hash {
			t.Fatal("frozen gauge source drift")
		}
	}
	r := intakeTestRenderer(t)
	p := gaugePlan(t, r, frozenGaugeNode, SceneContext{Surface: "light"})
	count := 0
	for _, it := range p.Items {
		if it.Shape == nil {
			continue
		}
		if it.Shape.Type == pptx.ShapeTypeBlockArc {
			i := count
			count++
			shape := it.Shape
			start, end := 180+float64(i)*30, 180+float64(i+1)*30
			if i > 0 {
				start += .8
			}
			if i < 5 {
				end -= .8
			}
			if shape.Props.AngleRange == nil || *shape.Props.AngleRange != ([2]float64{start, end}) || shape.Props.ArcThicknessRatio != .4 || shape.Props.Line.Width != .75 {
				t.Fatal("ring thickness/gap/stroke differs from source")
			}
			if shape.Record.Rect.Y+shape.Record.Rect.H > 252+.54 {
				t.Fatal("segment bounds include unpainted lower semicircle")
			}
		}
	}
	if count != 6 {
		t.Fatal("expected6editable ring bands")
	}
	needle := enhancementShape(t, p, ".needle")
	hub := enhancementShape(t, p, ".hub")
	if len(needle.Props.Points) != 4 || hub.Record.Rect != (Rect{788, 245, 14, 14}) || hub.Record.Color != "070154" {
		t.Fatal("source needle/hub missing")
	}
	tip := needle.Props.Points[0]
	tipx := needle.Props.X.Val*72 + tip.X.Val*72
	tipy := needle.Props.Y.Val*72 + tip.Y.Val*72
	angle := math.Pi + .58*math.Pi
	if tipx != math.Round((795+99.36*math.Cos(angle))*100)/100 || tipy != math.Round((252+99.36*math.Sin(angle))*100)/100 {
		t.Fatal("needle projection differs")
	}
	for _, it := range p.Items {
		if it.Text != nil && !inside(it.Text.Rect, p.Bounds) || it.Shape != nil && !inside(it.Shape.Record.Rect, p.Bounds) {
			t.Fatal("visible native ink outside reported bounds")
		}
	}
}
func TestIntakeGaugeDarkSurfaceAndText(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, surface := range []string{"inverse", "deep"} {
		p := gaugePlan(t, r, frozenGaugeNode, SceneContext{Surface: surface})
		for _, it := range p.Items {
			if it.Shape != nil && it.Shape.Type == pptx.ShapeTypeBlockArc && it.Shape.Props.Line.Type != "none" {
				t.Fatal("dark gauge should have nooutline")
			}
		}
		if enhancementShape(t, p, ".needle").Record.Color != "FFFFFF" || enhancementShape(t, p, ".hub").Record.Color != "FFFFFF" {
			t.Fatal("dark needle/hub notwhite")
		}
	}
	p := gaugePlan(t, r, strings.Replace(frozenGaugeNode, `"Not ready"`, `"Not ready[^1]"`, 1), SceneContext{Surface: "light", Notes: []string{"Source"}})
	seen := false
	for _, it := range p.Items {
		if it.Text != nil && strings.Contains(it.Text.ID, "end-1") {
			seen = it.Text.Rich != nil && it.Text.Rect.Y == 261
		}
		if it.Text != nil && strings.Contains(it.Text.ID, ".value.part-1") && it.Text.Rect.Y != 279 {
			t.Fatal("value stack origin shifted")
		}
	}
	if !seen {
		t.Fatal("end footnote semantics missing")
	}
}
func TestIntakeGaugeStrictFiniteAndResourceLimits(t *testing.T) {
	r := intakeTestRenderer(t)
	var object map[string]any
	json.Unmarshal([]byte(frozenGaugeNode), &object)
	for _, change := range []map[string]any{{"x": nil}, {"value": nil}, {"value": -1}, {"value": 1.1}, {"w": 0}, {"ends": []string{"Onlyone"}}, {"segments": []string{}}, {"segments": []string{"invalid"}}, {"h": 200}, {"caption": strings.Repeat("X", 8193)}, {"segments": make([]string, 13)}} {
		copy := map[string]any{}
		for k, v := range object {
			copy[k] = v
		}
		for k, v := range change {
			copy[k] = v
		}
		raw, _ := json.Marshal(copy)
		if _, handled, err := r.planIntakeGaugeScene("bad", raw, SceneContext{Surface: "light"}); !handled || err == nil {
			t.Fatalf("accepted invalid %v", change)
		}
	}
	var n intakeGaugeSource
	json.Unmarshal([]byte(frozenGaugeNode), &n)
	n.W = math.Inf(1)
	if _, err := r.intakeGauge("bad", n, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("nonfinite geometry")
	}
}
func TestIntakeGaugeNativeEditableXML(t *testing.T) {
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "gauge", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "gauge", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(frozenGaugeNode)}}}}}}
	data, report, err := BuildWithEngine(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if report.PowerPointVerified || report.VisuallyReviewed {
		t.Fatal("unreviewed gauge overstated qualification")
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
	if bytes.Count(body, []byte(`prst="blockArc"`)) != 6 || !bytes.Contains(body, []byte(`name="adj3" fmla="val 20000"`)) || !bytes.Contains(body, []byte("<a:custGeom>")) || !bytes.Contains(body, []byte("<p:grpSp>")) || bytes.Contains(body, []byte("<p:pic>")) {
		t.Fatal("native ring/needle/group missing or rasterized")
	}
	for _, text := range []string{"2.6", "Overall readiness (of 4)", "NOT READY", "READY"} {
		if !bytes.Contains(body, []byte(text)) {
			t.Fatal("gauge editable copy missing", text)
		}
	}
}
