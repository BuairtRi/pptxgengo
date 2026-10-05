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

func TestCurveFeedbackBeforeAfterLabelUsesAuthoredBandSample(t *testing.T) {
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source", "templates", "library"))
	slide := intakeRepairSlide(t, entries["team-curve/before-after"])
	raw := slide.Nodes[5].Scene.Node
	var node teamCurveSource
	if err := json.Unmarshal(raw, &node); err != nil {
		t.Fatal(err)
	}
	for _, revision := range []string{LibraryRevisionV3, LibraryRevisionV4} {
		r := intakeTestRenderer(t)
		r.source.Revision = revision
		p := intakeCurvePlan(t, r, string(raw), SceneContext{Surface: "light"})
		var label *TextRecord
		for _, it := range p.Items {
			if it.Text != nil && strings.EqualFold(it.Text.Layout.Original, "AI Agents") {
				label = it.Text
			}
		}
		if label == nil {
			t.Fatal("missing AI Agents label")
		}
		if label.Color != "FFFFFF" || len(label.Layout.Lines) != 1 {
			t.Fatalf("label color/wrapping changed: %+v", label)
		}
		if revision == LibraryRevisionV3 {
			if label.Rect.W != 180 || label.Rect.X != node.X+node.W-180 {
				t.Fatal("v3 fixed label allocation changed")
			}
			continue
		}
		// The evenly spaced source sample21 of25 defines the label anchor.
		anchor := node.X + (21.0/24.0)*node.W
		if label.Rect.W >= 180 || math.Abs(label.Rect.X+label.Rect.W/2-anchor) > .02 {
			t.Fatalf("AI label shifted away from authored sample: %+v anchor=%g", label.Rect, anchor)
		}
		if label.Rect.X < node.X || label.Rect.X+label.Rect.W > node.X+node.W+.02 {
			t.Fatal("AI label outside component")
		}
	}
}

func TestRoadFeedbackNativeMarkerCentering(t *testing.T) {
	raw := `{"type":"road","x":117,"y":126,"w":700,"h":280,"labelW":150,"milestones":[{"label":"Start","date":"Now","side":"above"},{"label":"Finish","side":"below","active":true}]}`
	for _, revision := range []string{LibraryRevisionV3, LibraryRevisionV4} {
		r := intakeTestRenderer(t)
		r.source.Revision = revision
		p := roadPlan(t, r, raw, SceneContext{Surface: "light"})
		count := 0
		for _, it := range p.Items {
			if it.Text == nil || !strings.HasSuffix(it.Text.ID, ".number") {
				continue
			}
			count++
			if revision == LibraryRevisionV4 {
				if it.Text.VerticalAlign != "middle" || it.Text.Rect.H != 22 {
					t.Fatalf("road number lacks full native middle box: %+v", it.Text)
				}
			} else if it.Text.VerticalAlign != "" || it.Text.Rect.H == 22 {
				t.Fatal("v3 road numeral allocation changed")
			}
		}
		if count != 2 {
			t.Fatalf("road marker count=%d", count)
		}
	}
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "road", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "road", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(raw)}}}}}}
	data, _, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, f := range z.File {
		if f.Name != "ppt/slides/slide1.xml" {
			continue
		}
		rd, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(rd)
		rd.Close()
		if err != nil {
			t.Fatal(err)
		}
		for _, shape := range strings.Split(string(body), "<p:sp>")[1:] {
			shape = strings.Split(shape, "</p:sp>")[0]
			if !strings.Contains(shape, ".number\"") {
				continue
			}
			count++
			if !strings.Contains(shape, `anchor="ctr"`) || !strings.Contains(shape, `cy="279400"`) {
				t.Fatal("native road number missing full22pt middle alignment")
			}
		}
	}
	if count != 2 {
		t.Fatalf("native road marker count=%d", count)
	}
}
