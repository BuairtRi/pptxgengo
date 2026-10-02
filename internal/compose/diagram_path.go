package compose

import "fmt"

// ProcessPathSpec expands an ordered row of measured text nodes and explicit
// native rightArrow relations. Bounds are slide-relative, or parent-content
// relative when ParentID is set, exactly like ContainerSpec.
type ProcessPathSpec struct {
	ID             string   `json:"id"`
	ParentID       string   `json:"parent_id,omitempty"`
	Bounds         Rect     `json:"bounds"`
	Labels         []string `json:"labels"`
	GapPt          float64  `json:"gap_pt"`
	NodePadding    Insets   `json:"node_padding"`
	NodeBackground string   `json:"node_background"`
	FontFace       string   `json:"font_face"`
	FontSizePt     float64  `json:"font_size_pt"`
	Foreground     string   `json:"foreground"`
	ArrowColor     string   `json:"arrow_color"`
	ArrowHeightPt  float64  `json:"arrow_height_pt,omitempty"`
	ArrowMarginPt  float64  `json:"arrow_margin_pt,omitempty"`
}

func expandProcessPaths(spec Spec) (Spec, []LayoutZoneFit) {
	out := spec
	out.Slides = append([]SlideSpec(nil), spec.Slides...)
	var failures []LayoutZoneFit
	for i := range out.Slides {
		s := &out.Slides[i]
		if len(s.Paths) == 0 {
			continue
		}
		s.Layouts = append([]ContainerSpec(nil), s.Layouts...)
		s.Connections = append([]ConnectionSpec(nil), s.Connections...)
		for _, path := range s.Paths {
			if !validID(path.ID) || !validRect(path.Bounds) || len(path.Labels) < 2 || !positive(path.GapPt) || !validInsets(path.NodePadding) || !ValidFontFace(path.FontFace) || !positive(path.FontSizePt) || !nonnegative(path.ArrowHeightPt) || !nonnegative(path.ArrowMarginPt) {
				failures = append(failures, fail(s.ID, path.ID, "", "", "invalid process path geometry, node count, or typography"))
				continue
			}
			if _, err := resolveColor(path.NodeBackground); err != nil {
				failures = append(failures, fail(s.ID, path.ID, "", "", "node background: "+err.Error()))
				continue
			}
			if _, err := foreground(path.Foreground, path.NodeBackground); err != nil {
				failures = append(failures, fail(s.ID, path.ID, "", "", "node foreground: "+err.Error()))
				continue
			}
			if _, err := resolveColor(path.ArrowColor); err != nil {
				failures = append(failures, fail(s.ID, path.ID, "", "", "arrow color: "+err.Error()))
				continue
			}
			if path.Bounds.Width-float64(len(path.Labels)-1)*path.GapPt <= 0 {
				failures = append(failures, fail(s.ID, path.ID, "", "", "node tracks have no width after gaps"))
				continue
			}
			container := ContainerSpec{ID: path.ID, ParentID: path.ParentID, Bounds: path.Bounds, ColumnGapPt: path.GapPt, Columns: make([]TrackSpec, len(path.Labels)), Rows: []TrackSpec{{FixedPt: path.Bounds.Height}}, Cells: make([]CellSpec, len(path.Labels))}
			for n, label := range path.Labels {
				if !validID(label) {
					failures = append(failures, fail(s.ID, path.ID, fmt.Sprintf("node-%d", n+1), "", "node label is empty"))
					continue
				}
				container.Columns[n] = TrackSpec{Weight: 1}
				container.Cells[n] = CellSpec{ID: fmt.Sprintf("node-%d", n+1), Row: 0, Column: n, Padding: path.NodePadding, Background: path.NodeBackground, Valign: "middle", Blocks: []BlockSpec{{ID: "label", Text: label, FontFace: path.FontFace, FontSizePt: path.FontSizePt, Foreground: path.Foreground, Align: "center", Valign: "middle"}}}
				if n > 0 {
					conn := ConnectionSpec{ID: fmt.Sprintf("%s/flow-%d", path.ID, n), From: fmt.Sprintf("%s/node-%d", path.ID, n), To: fmt.Sprintf("%s/node-%d", path.ID, n+1), Relationship: "flow", FromAnchor: "right", ToAnchor: "left", Color: path.ArrowColor, WidthPt: 1, ClearancePt: 2, ArrowHeightPt: path.ArrowHeightPt, GapMarginPt: path.ArrowMarginPt}
					s.Connections = append(s.Connections, conn)
				}
			}
			s.Layouts = append(s.Layouts, container)
		}
		s.Paths = nil
	}
	return out, failures
}
