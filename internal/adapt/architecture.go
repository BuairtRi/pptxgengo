package adapt

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/compose"
	"strings"
)

type architectureInput struct {
	Layers        []architectureLayer    `json:"layers"`
	LeftRailTitle string                 `json:"left_rail_title,omitempty"`
	LeftRail      []architectureRailItem `json:"left_rail,omitempty"`
	Relations     []architectureRelation `json:"relations,omitempty"`
}
type architectureLayer struct {
	ID         string             `json:"id"`
	Heading    string             `json:"heading"`
	Components []architectureItem `json:"components"`
}
type architectureItem struct {
	ID          string  `json:"id,omitempty"`
	Label       string  `json:"label"`
	Detail      string  `json:"detail,omitempty"`
	Fill        string  `json:"fill,omitempty"`
	WidthWeight float64 `json:"width_weight,omitempty"`
}
type architectureRailItem struct {
	Label  string `json:"label"`
	Detail string `json:"detail,omitempty"`
}
type architectureRelation struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Relationship string `json:"relationship"`
}

// buildArchitecture lowers layer rows into measured panels with independently editable component cells.
func buildArchitecture(c Context, raw json.RawMessage) (Content, error) {
	var in architectureInput
	if err := Decode(raw, &in); err != nil {
		return Content{}, fmt.Errorf("architecture content: %w", err)
	}
	if len(in.Layers) < 2 || len(in.Layers) > 8 {
		return Content{}, fmt.Errorf("architecture layers must contain 2–8 entries")
	}
	seen := map[string]bool{}
	componentIDs := map[string]bool{}
	componentFill := map[string]string{}
	for _, l := range in.Layers {
		if !adaptID(l.ID) || seen[l.ID] {
			return Content{}, fmt.Errorf("invalid or duplicate architecture layer id %q", l.ID)
		}
		seen[l.ID] = true
		if strings.TrimSpace(l.Heading) == "" || len(l.Components) < 1 || len(l.Components) > 8 {
			return Content{}, fmt.Errorf("layer %q needs a heading and 1–8 components", l.ID)
		}
		for j, x := range l.Components {
			if !adaptID(x.ID) || componentIDs[x.ID] || strings.TrimSpace(x.Label) == "" || x.WidthWeight < 0 || x.WidthWeight > 4 {
				return Content{}, fmt.Errorf("layer %q component %d needs unique id, label and width_weight in 0..4", l.ID, j)
			}
			componentIDs[x.ID] = true
			fill := c.Style.Surface
			switch x.Fill {
			case "accent":
				fill = c.Style.Accent
			case "active":
				fill = c.Style.Active
			case "muted":
				fill = c.Style.MutedSurface
			}
			componentFill[x.ID] = fill
			if x.Fill != "" && x.Fill != "accent" && x.Fill != "muted" && x.Fill != "surface" && x.Fill != "active" {
				return Content{}, fmt.Errorf("component %q fill must be accent, muted, surface, or active", x.ID)
			}
		}
	}
	if len(in.LeftRail) > 5 {
		return Content{}, fmt.Errorf("architecture left_rail supports at most 5 cross-cutting concerns")
	}
	if len(in.LeftRail) == 0 && strings.TrimSpace(in.LeftRailTitle) != "" {
		return Content{}, fmt.Errorf("left_rail_title requires at least one left_rail item")
	}
	for i, x := range in.LeftRail {
		if strings.TrimSpace(x.Label) == "" {
			return Content{}, fmt.Errorf("left_rail item %d needs a label", i)
		}
	}
	for _, r := range in.Relations {
		if !componentIDs[r.From] || !componentIDs[r.To] || r.From == r.To || !validConnectionType(r.Relationship) {
			return Content{}, fmt.Errorf("relations must connect distinct known component ids and use reporting, dependency, advisory, or annotation")
		}
	}
	b, st := c.Bounds, c.Style
	railW := 0.0
	separatorW := 0.0
	if len(in.LeftRail) > 0 {
		railW = b.Width * .205
		separatorW = 20
	}
	mainX := b.X + railW + separatorW
	mainW := b.Width - railW - separatorW
	counts := make([]int, len(in.Layers))
	for i, l := range in.Layers {
		counts[i] = len(l.Components)
	}
	out := Content{Controls: map[string]any{"layer_count": len(in.Layers), "components_per_layer": counts, "left_rail_count": len(in.LeftRail), "relation_count": len(in.Relations), "independent_component_cells": sumInts(counts), "crosscut_panel_scope": "all_layers", "separator_gap_pt": 20, "grid_header_height_pt": 24, "layer_gap_pt": 16, "native_content_sized_rail": true}, Limitations: []string{"The adapter supports horizontal component rows within each layer and vertical layer order. It does not infer arbitrary graph topology, component subfields, or nested diagrams."}}
	if railW > 0 {
		bgID := "architecture-crosscut-surface"
		overlaps := []string{"architecture-scope-heading", "architecture-crosscut-heading"}
		for i := range in.LeftRail {
			overlaps = append(overlaps, fmt.Sprintf("architecture-rail-list/rail-%d", i+1))
			overlaps = append(overlaps, fmt.Sprintf("architecture-rail-list/rail-%d/label", i+1))
			if in.LeftRail[i].Detail != "" {
				overlaps = append(overlaps, fmt.Sprintf("architecture-rail-list/rail-%d/detail", i+1))
			}
		}
		pale := st.MutedSurface
		out.Canvas = append(out.Canvas, compose.CanvasSpec{ID: bgID, Kind: "surface", Bounds: compose.Rect{X: b.X, Y: b.Y, Width: railW, Height: b.Height}, Background: pale, AllowOverlap: overlaps, Layer: 0})
		railTitle := strings.TrimSpace(in.LeftRailTitle)
		if railTitle == "" {
			railTitle = "CROSS-CUTTING"
		}
		out.Canvas = append(out.Canvas,
			compose.CanvasSpec{ID: "architecture-scope-heading", Kind: "text", Bounds: compose.Rect{X: b.X + 7, Y: b.Y + 5, Width: railW - 14, Height: 16}, Text: "APPLIES TO ALL LAYERS", FontFace: st.FontFace, FontSizePt: styleSize(st.LabelFontPt, 9), Bold: true, Foreground: st.Ink(pale), Align: "left", Valign: "middle", Background: pale, Layer: 2},
			compose.CanvasSpec{ID: "architecture-crosscut-heading", Kind: "text", Bounds: compose.Rect{X: b.X + 7, Y: b.Y + 23, Width: railW - 14, Height: 20}, Text: railTitle, FontFace: st.FontFace, FontSizePt: st.LabelFontPt, Bold: true, Foreground: st.Ink(pale), Align: "left", Valign: "middle", Background: pale, Layer: 2},
			compose.CanvasSpec{ID: "architecture-rail-separator", Kind: "line", Bounds: compose.Rect{X: b.X + railW + separatorW/2, Y: b.Y + 2, Width: 0, Height: b.Height - 4}, Foreground: st.Ink(st.White), LineWidthPt: 1.25, Layer: 1})
		rows := make([]compose.TrackSpec, len(in.LeftRail))
		cells := make([]compose.CellSpec, len(in.LeftRail))
		for i, item := range in.LeftRail {
			rows[i] = compose.TrackSpec{MinPt: 24}
			blocks := []compose.BlockSpec{{ID: "label", Text: item.Label, FontFace: st.FontFace, FontSizePt: st.BodyFontPt, Bold: true, Foreground: st.Ink(pale), Align: "left", Valign: "middle"}}
			if item.Detail != "" {
				detailSize := st.BodyFontPt - 1
				if detailSize < 9 {
					detailSize = 9
				}
				blocks = append(blocks, compose.BlockSpec{ID: "detail", Text: item.Detail, FontFace: st.FontFace, FontSizePt: detailSize, Foreground: st.Ink(pale), Align: "left", Valign: "middle", GapBeforePt: 1})
			}
			cells[i] = compose.CellSpec{ID: fmt.Sprintf("rail-%d", i+1), Row: i, Column: 0, Padding: compose.Insets{Top: 2, Bottom: 2, Left: 3, Right: 3}, Background: pale, Valign: "middle", Blocks: blocks}
		}
		out.Layouts = append(out.Layouts, compose.ContainerSpec{ID: "architecture-rail-list", Bounds: compose.Rect{X: b.X + 6, Y: b.Y + 47, Width: railW - 12, Height: b.Height - 51}, Padding: compose.Insets{Top: 0, Bottom: 0, Left: 0, Right: 0}, Columns: []compose.TrackSpec{{Weight: 1}}, Rows: rows, RowGapPt: 6, Layer: 1, Cells: cells})
	}
	gap := 16.0
	gridY := b.Y + 24
	gridH := b.Height - 24
	headingW := mainW * .22
	headerH := 20.0
	out.Canvas = append(out.Canvas,
		compose.CanvasSpec{ID: "architecture-layer-column-header", Kind: "text", Bounds: compose.Rect{X: mainX + 2, Y: b.Y + 1, Width: headingW - 4, Height: headerH - 2}, Text: "LAYER", FontFace: st.FontFace, FontSizePt: st.LabelFontPt, Bold: true, Foreground: st.Ink(st.White), Align: "center", Valign: "middle", Layer: 3},
		compose.CanvasSpec{ID: "architecture-component-column-header", Kind: "text", Bounds: compose.Rect{X: mainX + headingW + 3, Y: b.Y + 1, Width: mainW - headingW - 4, Height: headerH - 2}, Text: "COMPONENTS", FontFace: st.FontFace, FontSizePt: st.LabelFontPt, Bold: true, Foreground: st.Ink(st.White), Align: "center", Valign: "middle", Layer: 3})
	h := (gridH - gap*float64(len(in.Layers)-1)) / float64(len(in.Layers))
	for i, l := range in.Layers {
		cols := []compose.TrackSpec{{FixedPt: headingW}}
		cells := []compose.CellSpec{{ID: "layer-heading", Row: 0, Column: 0, Padding: compose.Insets{Top: 5, Bottom: 5, Left: 6, Right: 6}, Background: st.Navy, Valign: "middle", Blocks: []compose.BlockSpec{{ID: "heading", Text: l.Heading, FontFace: st.FontFace, FontSizePt: st.LabelFontPt + 1, Bold: true, Foreground: st.Ink(st.Navy), Align: "center", Valign: "middle"}}}}
		for j, x := range l.Components {
			weight := x.WidthWeight
			if weight == 0 {
				weight = 1
			}
			cols = append(cols, compose.TrackSpec{Weight: weight, MinPt: 52})
			fill := st.Surface
			switch x.Fill {
			case "accent":
				fill = st.Accent
			case "active":
				fill = st.Active
			case "muted":
				fill = st.MutedSurface
			}
			blocks := []compose.BlockSpec{{ID: "label", Text: x.Label, FontFace: st.FontFace, FontSizePt: st.BodyFontPt, Bold: true, Foreground: st.Ink(fill), Align: "center", Valign: "middle"}}
			if x.Detail != "" {
				detailSize := st.BodyFontPt - 1
				if detailSize < 9 {
					detailSize = 9
				}
				blocks = append(blocks, compose.BlockSpec{ID: "detail", Text: x.Detail, FontFace: st.FontFace, FontSizePt: detailSize, Foreground: st.Ink(fill), Align: "center", Valign: "middle", GapBeforePt: 2})
			}
			cells = append(cells, compose.CellSpec{ID: x.ID, Row: 0, Column: j + 1, Padding: compose.Insets{Top: 4, Bottom: 4, Left: 4, Right: 4}, Background: fill, Valign: "middle", Blocks: blocks})
		}
		out.Layouts = append(out.Layouts, compose.ContainerSpec{ID: "architecture-" + l.ID, Bounds: compose.Rect{X: mainX, Y: gridY + float64(i)*(h+gap), Width: mainW, Height: h}, Padding: compose.Insets{Top: 1, Bottom: 1, Left: 1, Right: 1}, Columns: cols, Rows: []compose.TrackSpec{{FixedPt: h - 2}}, ColumnGapPt: 3, Surface: st.Border, Layer: 1, Cells: cells})
	}
	for i, r := range in.Relations {
		color, err := connectorColor(st, componentFill[r.From], componentFill[r.To])
		if err != nil {
			return Content{}, fmt.Errorf("relation %s to %s: %w", r.From, r.To, err)
		}
		fromAnchor, toAnchor, direction := architectureAnchors(in.Layers, r.From, r.To)
		out.Connections = append(out.Connections, compose.ConnectionSpec{ID: fmt.Sprintf("architecture-relation-%02d", i+1), From: componentPort(in.Layers, r.From), To: componentPort(in.Layers, r.To), Relationship: r.Relationship, FromAnchor: fromAnchor, ToAnchor: toAnchor, Color: color, WidthPt: 1.25, ClearancePt: 2, PreferredDirection: direction})
	}
	return out, nil
}
func componentPort(ls []architectureLayer, id string) string {
	for _, l := range ls {
		for _, c := range l.Components {
			if c.ID == id {
				return "architecture-" + l.ID + "/" + id
			}
		}
	}
	return ""
}
func sumInts(xs []int) int {
	n := 0
	for _, x := range xs {
		n += x
	}
	return n
}
func validConnectionType(s string) bool {
	return s == "reporting" || s == "dependency" || s == "advisory" || s == "annotation"
}
func connectorColor(st Style, backgrounds ...string) (string, error) {
	candidates := []string{st.Accent, st.Text, st.Navy, st.Secondary, st.Border, st.White, "#000000", "#FFFFFF"}
	backgrounds = append(backgrounds, st.White)
	for _, color := range candidates {
		ok := true
		for _, bg := range backgrounds {
			ratio, err := compose.ContrastRatio(color, bg)
			if err != nil {
				return "", err
			}
			if ratio < 3 {
				ok = false
				break
			}
		}
		if ok {
			return color, nil
		}
	}
	return "", fmt.Errorf("no available connector color has 3:1 contrast against endpoint surfaces %v (candidates %v)", backgrounds, candidates)
}
func architectureAnchors(ls []architectureLayer, fromID, toID string) (string, string, string) {
	var fromLayer, toLayer, fromColumn, toColumn int
	for i, l := range ls {
		for j, c := range l.Components {
			if c.ID == fromID {
				fromLayer, fromColumn = i, j
			}
			if c.ID == toID {
				toLayer, toColumn = i, j
			}
		}
	}
	if fromLayer < toLayer {
		return "bottom", "top", "vertical"
	}
	if fromLayer > toLayer {
		return "top", "bottom", "vertical"
	}
	if fromColumn < toColumn {
		return "right", "left", "horizontal"
	}
	return "left", "right", "horizontal"
}
func detailSuffix(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return ": " + s
}
func styleSize(v, fallback float64) float64 {
	if v <= 0 {
		return fallback
	}
	return v
}
func adaptID(s string) bool {
	if s == "" || len(s) > 48 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
