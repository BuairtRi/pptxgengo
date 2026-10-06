package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

type sceneTableRowGroup struct {
	Label string `json:"label"`
	From  int    `json:"from"`
	To    int    `json:"to"`
	Fill  string `json:"fill,omitempty"`
}

func (g *sceneTableRowGroup) UnmarshalJSON(raw []byte) error {
	var decoded struct {
		Label string `json:"label"`
		From  *int   `json:"from"`
		To    *int   `json:"to"`
		Fill  string `json:"fill,omitempty"`
	}
	if err := sceneDecode(raw, &decoded); err != nil {
		return err
	}
	if decoded.From == nil || decoded.To == nil {
		return fmt.Errorf("scene.table_row_group_range_required")
	}
	*g = sceneTableRowGroup{Label: decoded.Label, From: *decoded.From, To: *decoded.To, Fill: decoded.Fill}
	return nil
}

func sceneTableV6Fields(raw json.RawMessage, revision string) error {
	if isV6OrLaterLibrary(revision) {
		return nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return err
	}
	for _, field := range []string{"heatMin", "groupW", "rowGroups"} {
		if _, ok := obj[field]; ok {
			return fmt.Errorf("scene.table_new_options_require_v6: %s", field)
		}
	}
	var cols []map[string]json.RawMessage
	if err := json.Unmarshal(obj["cols"], &cols); err != nil {
		return err
	}
	for _, col := range cols {
		for _, field := range []string{"min", "size"} {
			if _, ok := col[field]; ok {
				return fmt.Errorf("scene.table_new_column_options_require_v6: %s", field)
			}
		}
	}
	return nil
}

func sceneHeatMinimum(minimum *float64) float64 {
	if minimum != nil {
		return *minimum
	}
	return 0
}

// Source columns already sum to the reduced width; never shrink their authored
// widths a second time. The rail is separate from the editable native table.
func sceneTableRowGroupGeometry(n sceneTableSource) (sceneTableSource, float64, error) {
	if n.RowGroups == nil {
		if n.GroupW != nil {
			return n, 0, fmt.Errorf("scene.table_group_width_without_row_groups")
		}
		return n, 0, nil
	}
	gw := 24.0
	if n.GroupW != nil {
		gw = *n.GroupW
	}
	if gw <= 0 || math.IsNaN(gw) || math.IsInf(gw, 0) || n.W-gw-6 <= 0 {
		return n, 0, fmt.Errorf("scene.table_invalid_group_width")
	}
	used := make(map[int]bool)
	for _, g := range n.RowGroups {
		if strings.TrimSpace(g.Label) == "" || g.From < 0 || g.To < g.From || g.To >= len(n.Rows) {
			return n, 0, fmt.Errorf("scene.table_invalid_row_group")
		}
		for i := g.From; i <= g.To; i++ {
			if used[i] {
				return n, 0, fmt.Errorf("scene.table_overlapping_row_groups")
			}
			used[i] = true
		}
	}
	n.X += gw + 6
	n.W -= gw + 6
	return n, gw, nil
}

func sceneTableRowHeights(n sceneTableSource, fallback float64, revision string) ([]float64, error) {
	hColumn := false
	for _, col := range n.Columns {
		if col.Key == "h" {
			hColumn = true
		}
	}
	heights := make([]float64, len(n.Rows))
	for i, row := range n.Rows {
		heights[i] = fallback
		if raw, ok := row["h"]; ok && !hColumn {
			if !isV6OrLaterLibrary(revision) {
				return nil, fmt.Errorf("scene.table_row_height_requires_v6")
			}
			if string(raw) == "null" || json.Unmarshal(raw, &heights[i]) != nil || heights[i] <= 0 || math.IsNaN(heights[i]) || math.IsInf(heights[i], 0) {
				return nil, fmt.Errorf("scene.table_invalid_row_height: row%d", i)
			}
		}
	}
	return heights, nil
}

func (r *renderer) sceneTableRowGroupLabels(p *scenePlan, id string, n sceneTableSource, gw, bodyY float64, heights []float64, surface string, ctx SceneContext) error {
	if n.RowGroups == nil {
		return nil
	}
	keys, err := primitiveArrayKeys(ctx, "/rowGroups", len(n.RowGroups))
	if err != nil {
		return err
	}
	starts := make([]float64, len(heights)+1)
	starts[0] = bodyY
	for i, h := range heights {
		starts[i+1] = starts[i] + h
	}
	for gi, g := range n.RowGroups {
		box := Rect{n.X - gw - 6, starts[g.From] + 1.5, gw, starts[g.To+1] - starts[g.From] - 3}
		if box.H <= 0 {
			return fmt.Errorf("scene.table_row_group_geometry")
		}
		fill := "070154"
		var err error
		if g.Fill != "" {
			fill, err = r.sceneColor(surface, g.Fill)
			if err != nil {
				return err
			}
		}
		ink := "070154"
		if contrast("FFFFFF", fill) >= 4.5 {
			ink = "FFFFFF"
		}
		if contrast(ink, fill) < 4.5 {
			return fmt.Errorf("scene.table_row_group_contrast")
		}
		prefix := id + ".row-group." + keys[gi]
		r.sceneDataShape(p, prefix+".fill", box, pptx.ShapeTypeRect, fill, nil)
		st, err := r.sceneStyle("label")
		if err != nil {
			return err
		}
		st.Family = "IBM Plex Mono"
		st.Size = 8
		st.Weight = 600
		st.Tracking = "0.1em"
		st.TrackingPt = .8
		st.Case = "upper"
		text := g.Label
		layout, err := r.measureText(text, st, box.H)
		if err != nil {
			return err
		}
		height := math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight)
		if len(layout.Lines) != 1 || height > box.W+.02 {
			return fmt.Errorf("scene.table_row_group_label_overflow: %s", g.Label)
		}
		// Rotate a centered horizontal editable text frame around the rail center.
		tb := Rect{box.X + box.W/2 - box.H/2, box.Y + box.H/2 - height/2, box.H, height}
		if err = r.sequenceLiteralText(p, prefix+".label", text, st, tb, ink, "center", false); err != nil {
			return err
		}
		p.Items[len(p.Items)-1].Text.Rotation = -90
	}
	return nil
}
