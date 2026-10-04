package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

const IntakeScoreLegendContract = "pptxgengo.wmds-incoming-score-legend.v1"

// Native replacement for the incoming vendor legend's unqualified circle
// glyphs. Scores retain their ordinal meaning; labels remain editable text.
func (r *renderer) planIntakeScoreLegendScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, false, err
	}
	if head.Type != "scorelegend" {
		return nil, false, nil
	}
	var n struct {
		Type  string  `json:"type"`
		ID    string  `json:"id,omitempty"`
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
		W     float64 `json:"w"`
		Max   int     `json:"max"`
		Items []struct {
			Value *int   `json:"value"`
			Label string `json:"label"`
		} `json:"items"`
	}
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	var fields map[string]json.RawMessage
	json.Unmarshal(raw, &fields)
	for _, key := range []string{"x", "y", "w", "max"} {
		v, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, true, fmt.Errorf("scene.scorelegend_required_field: %s", key)
		}
	}
	if !intakeFinite(n.X, n.Y, n.W) || n.X < 0 || n.Y < 0 || n.W <= 0 || n.X+n.W > 960 || n.Y+18 > 540 || n.Max < 1 || n.Max > 12 || len(n.Items) < 1 || len(n.Items) > 8 {
		return nil, true, fmt.Errorf("scene.scorelegend_geometry_or_cardinality")
	}
	keys, err := primitiveArrayKeys(ctx, "/items", len(n.Items))
	if err != nil {
		return nil, true, err
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	color, err := r.sceneColor(surface, "primary")
	if err != nil {
		return nil, true, err
	}
	st, err := r.sceneStyle("small")
	if err != nil {
		return nil, true, err
	}
	width := n.W / float64(len(n.Items))
	markWidth := float64(n.Max*8 + (n.Max-1)*4)
	labelWidth := width - markWidth - 6 - 12
	if labelWidth <= 0 {
		return nil, true, fmt.Errorf("scene.scorelegend_insufficient_width")
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, 18}}
	for i, item := range n.Items {
		if item.Value == nil || *item.Value < 0 || *item.Value > n.Max || utf8.RuneCountInString(item.Label) == 0 || utf8.RuneCountInString(item.Label) > 128 {
			return nil, true, fmt.Errorf("scene.scorelegend_invalid_item")
		}
		layout, err := r.typeEngine.Measure(item.Label, st, labelWidth)
		if err != nil {
			return nil, true, err
		}
		if len(layout.Lines) != 1 || math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight) > 18+.02 {
			return nil, true, fmt.Errorf("scene.scorelegend_label_overflow")
		}
		x := n.X + float64(i)*width
		prefix := id + ".items." + keys[i]
		for j := 0; j < n.Max; j++ {
			b := Rect{x + float64(j)*12, n.Y + 5, 8, 8}
			fill, line, weight := color, "", 0.
			if j >= *item.Value {
				fill, line, weight = "", color, .75
				// Keep the outer stroke inside its declared 8pt allocation.
				b = Rect{b.X + .375, b.Y + .375, b.W - .75, b.H - .75}
			}
			if err = r.diagramShape(p, fmt.Sprintf("%s.marker-%d", prefix, j+1), b, pptx.ShapeTypeEllipse, fill, line, weight, "solid", nil); err != nil {
				return nil, true, err
			}
		}
		// Source legend labels are literal text, not emphasis markup.
		tr := TextRecord{ID: prefix + ".label", Rect: Rect{x + markWidth + 6, n.Y, labelWidth, 18}, Color: color, Align: "left", Layout: layout}
		p.Items = append(p.Items, sceneItem{Text: &tr})
	}
	primitiveFinish(p, "scorelegend")
	p.Groups[len(p.Groups)-1].Contract = IntakeScoreLegendContract
	return p, true, nil
}
