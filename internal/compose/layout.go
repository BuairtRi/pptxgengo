package compose

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Insets gives each side an independent native point value.
type Insets struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// TrackSpec describes one shared grid track. FixedPt is exclusive with Weight.
// A zero Weight on a flexible track means weight 1.
type TrackSpec struct {
	FixedPt float64 `json:"fixed_pt,omitempty"`
	MinPt   float64 `json:"min_pt,omitempty"`
	MaxPt   float64 `json:"max_pt,omitempty"`
	Weight  float64 `json:"weight,omitempty"`
}

// ContainerSpec is a bounded, ordered grid panel. Nested container bounds are
// relative to their parent's content top-left and must remain inside that padded
// content area; top-level bounds are slide-relative. A nested Layer is relative
// to and always above its parent container layer.
type ContainerSpec struct {
	ID           string      `json:"id"`
	ParentID     string      `json:"parent_id,omitempty"`
	Bounds       Rect        `json:"bounds"`
	Padding      Insets      `json:"padding"`
	Columns      []TrackSpec `json:"columns"`
	Rows         []TrackSpec `json:"rows"`
	ColumnGapPt  float64     `json:"column_gap_pt,omitempty"`
	ColumnGapsPt []float64   `json:"column_gaps_pt,omitempty"`
	RowGapPt     float64     `json:"row_gap_pt,omitempty"`
	Surface      string      `json:"surface,omitempty"`
	RowRule      *RuleSpec   `json:"row_rule,omitempty"`
	Layer        int         `json:"layer,omitempty"`
	AllowOverlap []string    `json:"allow_overlap,omitempty"`
	Cells        []CellSpec  `json:"cells"`
}

// RuleSpec draws a native rule at each internal dynamic row boundary.
type RuleSpec struct {
	Color    string  `json:"color"`
	WidthPt  float64 `json:"width_pt"`
	OffsetPt float64 `json:"offset_pt,omitempty"`
	Layer    int     `json:"layer,omitempty"`
}

type CellSpec struct {
	ID         string      `json:"id"`
	Row        int         `json:"row"`
	Column     int         `json:"column"`
	RowSpan    int         `json:"row_span,omitempty"`
	ColumnSpan int         `json:"column_span,omitempty"`
	Padding    Insets      `json:"padding"`
	Background string      `json:"background,omitempty"`
	Layer      int         `json:"layer,omitempty"`
	Valign     string      `json:"valign,omitempty"` // top, middle, bottom
	Blocks     []BlockSpec `json:"blocks,omitempty"`
}

// BlockSpec represents one measured text flow item. Marker creates a paired
// marker/text shape with a shared measured height and reserved marker width.
type BlockSpec struct {
	ID            string          `json:"id"`
	Text          string          `json:"text"`
	Paragraphs    []ParagraphSpec `json:"paragraphs,omitempty"`
	Marker        string          `json:"marker,omitempty"`
	MarkerWidthPt float64         `json:"marker_width_pt,omitempty"`
	GapBeforePt   float64         `json:"gap_before_pt,omitempty"`
	MinHeightPt   float64         `json:"min_height_pt,omitempty"`
	MaxHeightPt   float64         `json:"max_height_pt,omitempty"`
	LeftInsetPt   float64         `json:"left_inset_pt,omitempty"`
	RightInsetPt  float64         `json:"right_inset_pt,omitempty"`
	FontFace      string          `json:"font_face"`
	FontSizePt    float64         `json:"font_size_pt"`
	Bold          bool            `json:"bold,omitempty"`
	Foreground    string          `json:"foreground"`
	Background    string          `json:"background,omitempty"`
	Align         string          `json:"align,omitempty"`
	Valign        string          `json:"valign,omitempty"`
	Layer         int             `json:"layer,omitempty"`
}

type LayoutZoneFit struct {
	SlideID           string  `json:"slide_id"`
	ContainerID       string  `json:"container_id"`
	CellID            string  `json:"cell_id,omitempty"`
	BlockID           string  `json:"block_id,omitempty"`
	RequestID         string  `json:"request_id,omitempty"`
	AvailableWidthPt  float64 `json:"available_width_pt,omitempty"`
	AvailableHeightPt float64 `json:"available_height_pt,omitempty"`
	RequiredWidthPt   float64 `json:"required_width_pt,omitempty"`
	RequiredHeightPt  float64 `json:"required_height_pt,omitempty"`
	Fits              bool    `json:"fits"`
	Reason            string  `json:"reason,omitempty"`
}

type layoutMode int

const (
	layoutProbe layoutMode = iota
	layoutFinal
)

// ExpandLayoutsForProbes lowers layouts using only declared geometry. Text
// widths and stable IDs match final expansion; placeholder heights are never
// used as measurement evidence.
func ExpandLayoutsForProbes(spec Spec) (Spec, error) {
	return expandLayouts(spec, Measurements{}, layoutProbe)
}

// ExpandLayouts lowers declarative panels into the existing CanvasSpec model
// using native text measurements for row and block heights.
func ExpandLayouts(spec Spec, measured Measurements) (Spec, error) {
	return expandLayouts(spec, measured, layoutFinal)
}

// LayoutFitReport returns every capacity failure it can identify, rather than
// stopping at the first overfull block or track.
func LayoutFitReport(spec Spec, measured Measurements) []LayoutZoneFit {
	_, rows := expandLayoutsDetailed(spec, measured, layoutFinal)
	return rows
}

func expandLayouts(spec Spec, measured Measurements, mode layoutMode) (Spec, error) {
	out, failures := expandLayoutsDetailed(spec, measured, mode)
	if len(failures) == 0 {
		return out, nil
	}
	parts := make([]string, len(failures))
	for i, f := range failures {
		where := f.ContainerID
		if f.CellID != "" {
			where += "/" + f.CellID
		}
		if f.BlockID != "" {
			where += "/" + f.BlockID
		}
		parts[i] = fmt.Sprintf("%s: %s", where, f.Reason)
	}
	return Spec{}, fmt.Errorf("layout expansion failed: %s", strings.Join(parts, "; "))
}

type resolvedContainer struct {
	spec       ContainerSpec
	bounds     Rect
	layer      int
	cols, rows []float64
	colX, rowY []float64
}

func expandLayoutsDetailed(spec Spec, measured Measurements, mode layoutMode) (Spec, []LayoutZoneFit) {
	out := spec
	// SlideSpec contains slices, but lowering only replaces the slide-level
	// Canvas and Layouts headers. Copy the slide array so probe expansion never
	// mutates the caller's reusable source spec through a shared backing array.
	out.Slides = append([]SlideSpec(nil), spec.Slides...)
	var failures []LayoutZoneFit
	for si := range out.Slides {
		s := &out.Slides[si]
		s.Canvas = append([]CanvasSpec(nil), s.Canvas...)
		if len(s.Layouts) == 0 {
			continue
		}
		generated, fs := lowerSlideLayouts(*s, measured, mode)
		failures = append(failures, fs...)
		if len(fs) == 0 {
			s.Canvas = append(s.Canvas, generated...)
			s.Layouts = nil
		}
	}
	if len(failures) == 0 {
		if err := validateSpec(out); err != nil {
			failures = append(failures, LayoutZoneFit{Fits: false, Reason: err.Error()})
		}
	}
	return out, failures
}

func lowerSlideLayouts(s SlideSpec, measured Measurements, mode layoutMode) ([]CanvasSpec, []LayoutZoneFit) {
	var failures []LayoutZoneFit
	ids := map[string]bool{}
	for _, c := range s.Canvas {
		ids[c.ID] = true
	}
	byID := map[string]ContainerSpec{}
	for _, c := range s.Layouts {
		if !validID(c.ID) || byID[c.ID].ID != "" || ids[c.ID] {
			failures = append(failures, fail(s.ID, c.ID, "", "", "container ID is empty or duplicated"))
			continue
		}
		byID[c.ID] = c
		ids[c.ID] = true
	}
	state := map[string]int{}
	resolved := map[string]*resolvedContainer{}
	var resolve func(string) *resolvedContainer
	resolve = func(id string) *resolvedContainer {
		if state[id] == 2 {
			return resolved[id]
		}
		if state[id] == 1 {
			failures = append(failures, fail(s.ID, id, "", "", "parent cycle"))
			return nil
		}
		c, ok := byID[id]
		if !ok {
			return nil
		}
		state[id] = 1
		b := c.Bounds
		layer := c.Layer
		if c.ParentID != "" {
			p, exists := byID[c.ParentID]
			if !exists {
				failures = append(failures, fail(s.ID, id, "", "", "unknown parent "+c.ParentID))
				return nil
			}
			_ = p
			pr := resolve(c.ParentID)
			if pr == nil {
				return nil
			}
			content := Rect{X: pr.bounds.X + pr.spec.Padding.Left, Y: pr.bounds.Y + pr.spec.Padding.Top, Width: pr.bounds.Width - pr.spec.Padding.Left - pr.spec.Padding.Right, Height: pr.bounds.Height - pr.spec.Padding.Top - pr.spec.Padding.Bottom}
			b.X += content.X
			b.Y += content.Y
			layer = pr.layer + 1 + c.Layer
			if !inside(b, content) {
				failures = append(failures, fail(s.ID, id, "", "", "nested bounds exceed parent content area"))
				return nil
			}
		}
		if !validLayer(c.Layer) || !validRect(b) || !inside(b, Rect{Width: s.WidthPt, Height: s.HeightPt}) || !validInsets(c.Padding) || !nonnegative(c.ColumnGapPt) || !nonnegative(c.RowGapPt) || len(c.Columns) == 0 || len(c.Rows) == 0 {
			failures = append(failures, fail(s.ID, id, "", "", "invalid container geometry, padding, gaps, or empty tracks"))
			return nil
		}
		if len(c.ColumnGapsPt) > 0 {
			if c.ColumnGapPt != 0 || len(c.ColumnGapsPt) != len(c.Columns)-1 {
				failures = append(failures, fail(s.ID, id, "", "", "column_gaps_pt must have columns-1 values and cannot be combined with column_gap_pt"))
				return nil
			}
			for _, gap := range c.ColumnGapsPt {
				if !nonnegative(gap) {
					failures = append(failures, fail(s.ID, id, "", "", "column_gaps_pt values must be finite and nonnegative"))
					return nil
				}
			}
		}
		if c.RowRule != nil {
			if !validLayer(c.RowRule.Layer) || !positive(c.RowRule.WidthPt) || c.RowRule.WidthPt > 6 || !finite(c.RowRule.OffsetPt) {
				failures = append(failures, fail(s.ID, id, "", "", "row_rule requires finite offset and width in (0,6]"))
				return nil
			}
			if _, err := resolveColor(c.RowRule.Color); err != nil {
				failures = append(failures, fail(s.ID, id, "", "", "row_rule: "+err.Error()))
				return nil
			}
		}
		innerW := b.Width - c.Padding.Left - c.Padding.Right
		innerH := b.Height - c.Padding.Top - c.Padding.Bottom
		cols, err := resolveTracks(c.Columns, innerW-sum(columnGaps(c)), false)
		if err != nil {
			failures = append(failures, fail(s.ID, id, "", "", "columns: "+err.Error()))
			return nil
		}
		rc := &resolvedContainer{spec: c, bounds: b, layer: layer, cols: cols}
		resolved[id] = rc
		state[id] = 2
		_ = innerH
		return rc
	}
	for _, c := range s.Layouts {
		resolve(c.ID)
	}
	if len(failures) > 0 {
		return nil, failures
	}
	isAncestor := func(ancestor, child string) bool {
		for child != "" {
			p := byID[child].ParentID
			if p == ancestor {
				return true
			}
			child = p
		}
		return false
	}
	allows := func(c ContainerSpec, peer string) bool {
		for _, id := range c.AllowOverlap {
			if id == peer {
				return true
			}
		}
		return false
	}
	for i, a := range s.Layouts {
		for _, b := range s.Layouts[i+1:] {
			if overlap(resolved[a.ID].bounds, resolved[b.ID].bounds) && !isAncestor(a.ID, b.ID) && !isAncestor(b.ID, a.ID) && !allows(a, b.ID) && !allows(b, a.ID) {
				failures = append(failures, fail(s.ID, a.ID, "", "", "container overlaps "+b.ID+" without a declared relationship"))
			}
		}
	}
	if len(failures) > 0 {
		return nil, failures
	}

	// Compute rows only after widths are final, since measurement contracts are width-specific.
	for _, c := range s.Layouts {
		rc := resolved[c.ID]
		reqs := make([]float64, len(c.Rows))
		occupied := map[[2]int]string{}
		for i, t := range c.Rows {
			reqs[i] = t.MinPt
			if t.FixedPt > 0 {
				reqs[i] = t.FixedPt
			}
		}
		for _, cell := range c.Cells {
			if cell.RowSpan == 0 {
				cell.RowSpan = 1
			}
			if cell.ColumnSpan == 0 {
				cell.ColumnSpan = 1
			}
			if !validLayer(cell.Layer) || !validID(cell.ID) || !validInsets(cell.Padding) || (cell.Valign != "" && cell.Valign != "top" && cell.Valign != "middle" && cell.Valign != "bottom") || cell.Row < 0 || cell.Column < 0 || cell.RowSpan <= 0 || cell.ColumnSpan <= 0 || cell.Row+cell.RowSpan > len(c.Rows) || cell.Column+cell.ColumnSpan > len(c.Columns) {
				failures = append(failures, fail(s.ID, c.ID, cell.ID, "", "invalid cell ID, padding, or span"))
				continue
			}
			cellOverlap := false
			for row := cell.Row; row < cell.Row+cell.RowSpan; row++ {
				for col := cell.Column; col < cell.Column+cell.ColumnSpan; col++ {
					key := [2]int{row, col}
					if prior := occupied[key]; prior != "" {
						failures = append(failures, fail(s.ID, c.ID, cell.ID, "", "grid area overlaps cell "+prior))
						cellOverlap = true
					} else {
						occupied[key] = cell.ID
					}
				}
			}
			if cellOverlap {
				continue
			}
			cw := spanSizeGaps(rc.cols, columnGaps(c), cell.Column, cell.ColumnSpan) - cell.Padding.Left - cell.Padding.Right
			if cw <= 0 {
				failures = append(failures, fail(s.ID, c.ID, cell.ID, "", "cell has no usable width"))
				continue
			}
			required := cell.Padding.Top + cell.Padding.Bottom
			for _, b := range cell.Blocks {
				bh, fs := blockHeight(s.ID, c, cell, b, cw, measured, mode)
				failures = append(failures, fs...)
				required += b.GapBeforePt + bh
			}
			// Attribute bounded row failures to every affected cell before resolving
			// aggregate tracks, so one bad row cannot hide a second overflowing cell.
			if cell.RowSpan == 1 {
				limit := c.Rows[cell.Row].MaxPt
				if c.Rows[cell.Row].FixedPt > 0 {
					limit = c.Rows[cell.Row].FixedPt
				}
				if limit > 0 && required > limit+1e-6 {
					failures = append(failures, LayoutZoneFit{SlideID: s.ID, ContainerID: c.ID, CellID: cell.ID, AvailableHeightPt: limit, RequiredHeightPt: required, Fits: false, Reason: "cell content exceeds bounded row height"})
				}
			}
			if cell.RowSpan == 1 && required > reqs[cell.Row] {
				reqs[cell.Row] = required
			}
			if cell.RowSpan > 1 {
				current := 0.0
				for j := cell.Row; j < cell.Row+cell.RowSpan; j++ {
					current += reqs[j]
				}
				current += float64(cell.RowSpan-1) * c.RowGapPt
				if required > current {
					reqs[cell.Row+cell.RowSpan-1] += required - current
				}
			}
		}
		rows, err := resolveRows(c.Rows, reqs)
		if err != nil {
			failures = append(failures, fail(s.ID, c.ID, "", "", "rows: "+err.Error()))
			continue
		}
		used := sum(rows) + float64(len(rows)-1)*c.RowGapPt
		avail := rc.bounds.Height - c.Padding.Top - c.Padding.Bottom
		if used > avail+1e-6 {
			failures = append(failures, LayoutZoneFit{SlideID: s.ID, ContainerID: c.ID, AvailableHeightPt: avail, RequiredHeightPt: used, Fits: false, Reason: "row tracks overflow container height"})
			continue
		}
		rc.rows = rows
		rc.colX = positionsGaps(rc.bounds.X+c.Padding.Left, rc.cols, columnGaps(c))
		rc.rowY = positions(rc.bounds.Y+c.Padding.Top, rows, c.RowGapPt)
	}
	if len(failures) > 0 {
		return nil, failures
	}

	var out []CanvasSpec
	type ownership struct{ container, cell, kind string }
	owned := map[string]ownership{}
	containerBackground := func(id string) string {
		for id != "" {
			if bg := byID[id].Surface; bg != "" {
				return bg
			}
			id = byID[id].ParentID
		}
		return white
	}
	for _, c := range s.Layouts {
		rc := resolved[c.ID]
		if c.Surface != "" {
			out = append(out, CanvasSpec{ID: c.ID, Kind: "surface", Bounds: rc.bounds, Background: c.Surface, Layer: rc.layer})
			owned[c.ID] = ownership{container: c.ID, kind: "container_surface"}
		}
		if c.RowRule != nil {
			gridWidth := sum(rc.cols) + sum(columnGaps(c))
			for row := 0; row < len(rc.rows)-1; row++ {
				id := fmt.Sprintf("%s/row-rule/%d", c.ID, row+1)
				y := rc.rowY[row] + rc.rows[row] + c.RowRule.OffsetPt
				out = append(out, CanvasSpec{ID: id, Kind: "line", Bounds: Rect{X: rc.bounds.X + c.Padding.Left, Y: y, Width: gridWidth}, Foreground: c.RowRule.Color, LineWidthPt: c.RowRule.WidthPt, Layer: rc.layer + 1 + c.RowRule.Layer})
				owned[id] = ownership{container: c.ID, kind: "rule"}
			}
		}
		seenCells := map[string]bool{}
		for _, cell := range c.Cells {
			if cell.RowSpan == 0 {
				cell.RowSpan = 1
			}
			if cell.ColumnSpan == 0 {
				cell.ColumnSpan = 1
			}
			cid := c.ID + "/" + cell.ID
			if !validID(cell.ID) || seenCells[cell.ID] || ids[cid] {
				failures = append(failures, fail(s.ID, c.ID, cell.ID, "", "duplicate or empty cell ID"))
				continue
			}
			seenCells[cell.ID] = true
			ids[cid] = true
			cb := Rect{X: rc.colX[cell.Column], Y: rc.rowY[cell.Row], Width: spanSizeGaps(rc.cols, columnGaps(c), cell.Column, cell.ColumnSpan), Height: spanSize(rc.rows, c.RowGapPt, cell.Row, cell.RowSpan)}
			if cell.Background != "" {
				out = append(out, CanvasSpec{ID: cid, Kind: "surface", Bounds: cb, Background: cell.Background, Layer: rc.layer + 1 + cell.Layer})
				owned[cid] = ownership{container: c.ID, cell: cid, kind: "cell_surface"}
			}
			x := cb.X + cell.Padding.Left
			y := cb.Y + cell.Padding.Top
			w := cb.Width - cell.Padding.Left - cell.Padding.Right
			contentH := 0.0
			for _, b := range cell.Blocks {
				bh, _ := blockHeight(s.ID, c, cell, b, w, measured, mode)
				contentH += b.GapBeforePt + bh
			}
			freeH := cb.Height - cell.Padding.Top - cell.Padding.Bottom - contentH
			if cell.Valign == "middle" {
				y += freeH / 2
			} else if cell.Valign == "bottom" {
				y += freeH
			}
			seenBlocks := map[string]bool{}
			contrastBackground := cell.Background
			if contrastBackground == "" {
				contrastBackground = containerBackground(c.ID)
			}
			for _, b := range cell.Blocks {
				if !validID(b.ID) || seenBlocks[b.ID] {
					failures = append(failures, fail(s.ID, c.ID, cell.ID, b.ID, "duplicate or empty block ID"))
					continue
				}
				seenBlocks[b.ID] = true
				y += b.GapBeforePt
				bh, _ := blockHeight(s.ID, c, cell, b, w, measured, mode)
				base := cid + "/" + b.ID
				textID := base
				tx := x + b.LeftInsetPt
				tw := w - b.LeftInsetPt - b.RightInsetPt
				if b.Marker != "" {
					textID = base + "/text"
					tx += b.MarkerWidthPt
					tw -= b.MarkerWidthPt
					mid := base + "/marker"
					out = append(out, textCanvas(mid, b.Marker, Rect{X: x + b.LeftInsetPt, Y: y, Width: b.MarkerWidthPt, Height: bh}, b, contrastBackground, rc.layer+2+cell.Layer+b.Layer))
					owned[mid] = ownership{container: c.ID, cell: cid, kind: "text"}
				}
				out = append(out, textCanvas(textID, b.Text, Rect{X: tx, Y: y, Width: tw, Height: bh}, b, contrastBackground, rc.layer+2+cell.Layer+b.Layer))
				owned[textID] = ownership{container: c.ID, cell: cid, kind: "text"}
				y += bh
			}
		}
	}
	if len(failures) > 0 {
		return nil, failures
	}
	// Permit only ownership overlaps: container surfaces behind descendants,
	// and cell surfaces behind blocks in that cell. Text/text collisions remain
	// errors, including across explicitly overlapping sibling containers.
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			a, b := owned[out[i].ID], owned[out[j].ID]
			ownedPair := (a.kind == "container_surface" && (a.container == b.container || isAncestor(a.container, b.container))) ||
				(b.kind == "container_surface" && (b.container == a.container || isAncestor(b.container, a.container))) ||
				(a.kind == "cell_surface" && a.cell == b.cell) || (b.kind == "cell_surface" && b.cell == a.cell) ||
				(a.kind == "rule" && strings.HasSuffix(b.kind, "surface") && a.container == b.container) ||
				(b.kind == "rule" && strings.HasSuffix(a.kind, "surface") && a.container == b.container)
			explicitSurfacePair := a.kind == "container_surface" && b.kind == "container_surface" && (allows(byID[a.container], b.container) || allows(byID[b.container], a.container))
			if ownedPair || explicitSurfacePair {
				out[i].AllowOverlap = append(out[i].AllowOverlap, out[j].ID)
			}
		}
		for _, peer := range byID[owned[out[i].ID].container].AllowOverlap {
			// Layout-to-layout declarations authorize their panel surfaces only;
			// generated text remains collision-checked. Non-layout IDs are external.
			if _, layoutPeer := byID[peer]; !layoutPeer {
				out[i].AllowOverlap = append(out[i].AllowOverlap, peer)
			}
		}
		out[i].AllowOverlap = without(out[i].AllowOverlap, out[i].ID)
	}
	return out, nil
}

func blockHeight(slide string, c ContainerSpec, cell CellSpec, b BlockSpec, width float64, m Measurements, mode layoutMode) (float64, []LayoutZoneFit) {
	var fs []LayoutZoneFit
	rich := len(b.Paragraphs) > 0
	if !validLayer(b.Layer) || !validID(b.ID) || (!rich && (!validID(b.Text) || b.FontFace != "Arial" || !positive(b.FontSizePt))) || (rich && (b.Text != "" || b.FontFace != "" || b.FontSizePt != 0 || b.Bold || b.Foreground != "")) || !nonnegative(b.GapBeforePt) || !nonnegative(b.MinHeightPt) || !nonnegative(b.MaxHeightPt) || !nonnegative(b.LeftInsetPt) || !nonnegative(b.RightInsetPt) || !nonnegative(b.MarkerWidthPt) || (b.Marker != "" && (!positive(b.MarkerWidthPt) || rich)) || (b.MaxHeightPt > 0 && b.MaxHeightPt < b.MinHeightPt) {
		return 0, []LayoutZoneFit{fail(slide, c.ID, cell.ID, b.ID, "invalid block content, typography, inset, gap, or height bounds")}
	}
	if rich {
		if err := validateRichTextStructure(b.Paragraphs); err != nil {
			return 0, []LayoutZoneFit{fail(slide, c.ID, cell.ID, b.ID, err.Error())}
		}
	}
	if b.Align == "" {
		b.Align = "left"
	}
	if b.Valign == "" {
		b.Valign = "top"
	}
	textW := width - b.LeftInsetPt - b.RightInsetPt
	if b.Marker != "" {
		textW -= b.MarkerWidthPt
	}
	if textW <= 0 {
		return 0, []LayoutZoneFit{fail(slide, c.ID, cell.ID, b.ID, "block has no usable text width")}
	}
	if rich {
		if err := validateRichBulletWidth(b.Paragraphs, textW); err != nil {
			return 0, []LayoutZoneFit{fail(slide, c.ID, cell.ID, b.ID, err.Error())}
		}
	}
	h := b.MinHeightPt
	if mode == layoutProbe {
		if h <= 0 {
			h = 1
		}
		return h, nil
	}
	base := c.ID + "/" + cell.ID + "/" + b.ID
	textElement := base
	if b.Marker != "" {
		textElement += "/text"
	}
	ids := []string{requestID(slide, "canvas", textElement)}
	widths := []float64{textW}
	if b.Marker != "" {
		ids = append(ids, requestID(slide, "canvas", base+"/marker"))
		widths = append(widths, b.MarkerWidthPt)
	}
	for i, id := range ids {
		z, ok := m.ByRequestID[id]
		if !ok || !positive(z.RenderedWidthPt) || !positive(z.RenderedHeightPt) {
			fs = append(fs, LayoutZoneFit{SlideID: slide, ContainerID: c.ID, CellID: cell.ID, BlockID: b.ID, RequestID: id, Fits: false, Reason: "missing or invalid native measurement"})
			continue
		}
		const nativeTolerance = .15
		requiredWidth := z.RenderedWidthPt + math.Max(0, z.OffsetXPt)
		requiredHeight := z.RenderedHeightPt + math.Max(0, z.OffsetYPt)
		if z.OffsetXPt < -nativeTolerance || z.OffsetYPt < -nativeTolerance {
			fs = append(fs, LayoutZoneFit{SlideID: slide, ContainerID: c.ID, CellID: cell.ID, BlockID: b.ID, RequestID: id, RequiredWidthPt: requiredWidth, RequiredHeightPt: requiredHeight, Fits: false, Reason: "native text ink begins outside its zero-inset frame"})
		}
		if requiredWidth > widths[i]+nativeTolerance {
			fs = append(fs, LayoutZoneFit{SlideID: slide, ContainerID: c.ID, CellID: cell.ID, BlockID: b.ID, RequestID: id, AvailableWidthPt: widths[i], RequiredWidthPt: requiredWidth, Fits: false, Reason: "measured text exceeds available width"})
		}
		h = math.Max(h, requiredHeight)
	}
	if b.MaxHeightPt > 0 && h > b.MaxHeightPt+1e-6 {
		fs = append(fs, LayoutZoneFit{SlideID: slide, ContainerID: c.ID, CellID: cell.ID, BlockID: b.ID, AvailableHeightPt: b.MaxHeightPt, RequiredHeightPt: h, Fits: false, Reason: "measured text exceeds maximum block height"})
	}
	return h, fs
}

func textCanvas(id, text string, r Rect, b BlockSpec, inheritedBackground string, layer int) CanvasSpec {
	a := b.Align
	if a == "" {
		a = "left"
	}
	v := b.Valign
	if v == "" {
		v = "top"
	}
	contrast := inheritedBackground
	if b.Background != "" {
		contrast = b.Background
	}
	paragraphs := b.Paragraphs
	if len(paragraphs) > 0 {
		paragraphs = resolveRichText(paragraphs, contrast)
		text = ""
	}
	return CanvasSpec{ID: id, Kind: "text", Bounds: r, Text: text, Paragraphs: paragraphs, FontFace: b.FontFace, FontSizePt: b.FontSizePt, Bold: b.Bold, Foreground: b.Foreground, Background: b.Background, ContrastBackground: contrast, Align: a, Valign: v, Layer: layer}
}
func validLayer(layer int) bool { return layer >= 0 && layer <= 1000000 }

func validInsets(p Insets) bool {
	return nonnegative(p.Top) && nonnegative(p.Right) && nonnegative(p.Bottom) && nonnegative(p.Left)
}
func fail(s, c, cell, b, reason string) LayoutZoneFit {
	return LayoutZoneFit{SlideID: s, ContainerID: c, CellID: cell, BlockID: b, Fits: false, Reason: reason}
}
func sum(v []float64) float64 {
	n := 0.0
	for _, x := range v {
		n += x
	}
	return n
}
func spanSize(v []float64, g float64, start, count int) float64 {
	return sum(v[start:start+count]) + float64(count-1)*g
}
func positions(start float64, v []float64, g float64) []float64 {
	o := make([]float64, len(v))
	x := start
	for i, w := range v {
		o[i] = x
		x += w + g
	}
	return o
}
func columnGaps(c ContainerSpec) []float64 {
	if len(c.ColumnGapsPt) > 0 {
		return c.ColumnGapsPt
	}
	gaps := make([]float64, len(c.Columns)-1)
	for i := range gaps {
		gaps[i] = c.ColumnGapPt
	}
	return gaps
}
func spanSizeGaps(v, gaps []float64, start, count int) float64 {
	n := sum(v[start : start+count])
	if count > 1 {
		n += sum(gaps[start : start+count-1])
	}
	return n
}
func positionsGaps(start float64, v, gaps []float64) []float64 {
	out := make([]float64, len(v))
	x := start
	for i, width := range v {
		out[i] = x
		x += width
		if i < len(gaps) {
			x += gaps[i]
		}
	}
	return out
}
func without(v []string, skip string) []string {
	seen := map[string]bool{}
	o := []string{}
	for _, x := range v {
		if x != "" && x != skip && !seen[x] {
			seen[x] = true
			o = append(o, x)
		}
	}
	return o
}

func resolveTracks(ts []TrackSpec, available float64, rows bool) ([]float64, error) {
	if available <= 0 {
		return nil, fmt.Errorf("no available space")
	}
	out := make([]float64, len(ts))
	flex := []int{}
	for i, t := range ts {
		if !nonnegative(t.FixedPt) || !nonnegative(t.MinPt) || !nonnegative(t.MaxPt) || !nonnegative(t.Weight) || (t.MaxPt > 0 && t.MaxPt < t.MinPt) || (t.FixedPt > 0 && (t.Weight > 0 || t.MinPt > t.FixedPt || (t.MaxPt > 0 && t.MaxPt < t.FixedPt))) {
			return nil, fmt.Errorf("invalid track %d", i)
		}
		if t.FixedPt > 0 {
			out[i] = t.FixedPt
		} else {
			out[i] = t.MinPt
			flex = append(flex, i)
		}
	}
	remain := available - sum(out)
	if remain < -1e-6 {
		return nil, fmt.Errorf("minimum tracks need %.2fpt, only %.2fpt available", sum(out), available)
	}
	for remain > 1e-6 && len(flex) > 0 {
		weight := 0.0
		for _, i := range flex {
			w := ts[i].Weight
			if w == 0 {
				w = 1
			}
			weight += w
		}
		used := 0.0
		next := []int{}
		for _, i := range flex {
			w := ts[i].Weight
			if w == 0 {
				w = 1
			}
			add := remain * w / weight
			if ts[i].MaxPt > 0 && out[i]+add > ts[i].MaxPt {
				add = ts[i].MaxPt - out[i]
			} else {
				next = append(next, i)
			}
			out[i] += add
			used += add
		}
		if used <= 1e-9 {
			break
		}
		remain -= used
		flex = next
	}
	_ = rows
	return out, nil
}
func resolveRows(ts []TrackSpec, required []float64) ([]float64, error) {
	out := make([]float64, len(ts))
	for i, t := range ts {
		if !nonnegative(t.FixedPt) || !nonnegative(t.MinPt) || !nonnegative(t.MaxPt) || !nonnegative(t.Weight) || (t.MaxPt > 0 && t.MaxPt < t.MinPt) || (t.FixedPt > 0 && (t.Weight > 0 || t.MinPt > t.FixedPt || (t.MaxPt > 0 && t.MaxPt < t.FixedPt))) {
			return nil, fmt.Errorf("invalid track %d", i)
		}
		out[i] = math.Max(t.MinPt, required[i])
		if t.FixedPt > 0 {
			out[i] = t.FixedPt
			if required[i] > t.FixedPt+1e-6 {
				return nil, fmt.Errorf("track %d fixed %.2fpt but content needs %.2fpt", i, t.FixedPt, required[i])
			}
		}
		if t.MaxPt > 0 && out[i] > t.MaxPt+1e-6 {
			return nil, fmt.Errorf("track %d maximum %.2fpt but content needs %.2fpt", i, t.MaxPt, out[i])
		}
	}
	return out, nil
}

// DeterministicLayoutOrder sorts a copy by explicit layer while retaining the
// original order for equal layers.
func DeterministicLayoutOrder(in []PlannedCanvas) []PlannedCanvas {
	out := append([]PlannedCanvas(nil), in...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Layer < out[j].Layer })
	return out
}
