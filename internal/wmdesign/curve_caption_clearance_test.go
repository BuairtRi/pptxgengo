package wmdesign

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestMaturityTargetAvoidsActualCurveForVariableCounts(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, count := range []int{3, 5, 6} {
		for _, shape := range []float64{1, 4.2, 9} {
			t.Run(fmt.Sprintf("count%d-shape%g", count, shape), func(t *testing.T) {
				stages := []map[string]any{}
				at := []float64{}
				for i := 0; i < count; i++ {
					stages = append(stages, map[string]any{"label": fmt.Sprintf("Stage %d", i+1)})
					at = append(at, .04+float64(i)/float64(count-1)*.86)
				}
				input := map[string]any{"type": "maturity", "x": 57, "y": 126, "w": 846, "h": 342, "stages": stages, "at": at, "shape": shape, "target": map[string]any{"stage": count - 1, "label": "Illustrative target"}}
				raw, _ := json.Marshal(input)
				p := maturityTestPlan(t, r, string(raw), SceneContext{Surface: "light", Zone: Rect{57, 126, 846, 342}})
				var label *TextRecord
				var curve *sceneShape
				for _, it := range p.Items {
					if it.Text != nil && strings.HasSuffix(it.Text.ID, ".target.label") {
						label = it.Text
					}
					if it.Shape != nil && it.Shape.Record.ID == "maturity.curve" {
						curve = it.Shape
					}
				}
				if label == nil || curve == nil {
					t.Fatal("target/curve missing")
				}
				points := []curvePoint{}
				for _, q := range curve.Props.Points {
					points = append(points, curvePoint{curve.Props.X.Val*72 + q.X.Val*72, curve.Props.Y.Val*72 + q.Y.Val*72})
				}
				if label.Rect.Y < maturityCurveBottom(points, label.Rect.X-6, label.Rect.X+label.Rect.W+6)+6-.02 || label.Rect.Y+label.Rect.H > 432+.02 {
					t.Fatalf("label intersects actual curve/axis %+v", label.Rect)
				}
			})
		}
	}
}

func TestMaturityTargetRefusesUnboundedLabel(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := fmt.Sprintf(`{"type":"maturity","x":57,"y":126,"w":846,"h":100,"stages":[{"label":"One"},{"label":"Two"}],"target":{"stage":0,"label":%q}}`, strings.Repeat("Extensive label ", 80))
	if _, _, e := r.planIntakeMaturityScene("m", json.RawMessage(raw), SceneContext{Surface: "light"}); e == nil || !strings.Contains(e.Error(), "target_label_clearance") {
		t.Fatalf("unbounded target accepted: %v", e)
	}
}

func TestStaffingScaleCaptionReservesBottomWithoutMovingPlotOrigin(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"teamcurve","x":57,"y":126,"w":846,"h":342,"phaseH":72,"series":[{"name":"Capacity","fill":"#0047FF","labelAt":false,"values":[10,20,30]}],"unit":"relative capacity","timeUnit":"normalized delivery","source":"Illustrative assumption"}`
	p := intakeCurvePlan(t, r, raw, SceneContext{Surface: "light"})
	var caption *TextRecord
	for _, item := range p.Items {
		if item.Text != nil && item.Text.ID == "curve.scale" {
			caption = item.Text
		}
	}
	if caption == nil || caption.Rect.Y != 450 || caption.Rect.Y+caption.Rect.H > 468+.02 {
		t.Fatalf("scale outside reserved bottom %+v", caption)
	}
	// The native area fills end before the reserved bottom caption; no new
	// heading or shifting of the plot down into retained source annotations.
	for _, item := range p.Items {
		if item.Shape != nil && strings.Contains(item.Shape.Record.ID, ".series.") && item.Shape.Type == "custGeom" && item.Shape.Record.Rect.Y+item.Shape.Record.Rect.H > 446+.02 {
			t.Fatal("area overwrites scale", item.Shape.Record)
		}
	}
}
