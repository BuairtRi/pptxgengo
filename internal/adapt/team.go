package adapt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buairtri/pptxgengo/internal/compose"
)

type teamInput struct {
	Pods             []teamPod           `json:"pods"`
	ReportingLines   []teamReportingLine `json:"reporting_lines,omitempty"`
	Responsibilities *teamMatrix         `json:"responsibilities,omitempty"`
	MatrixPosition   string              `json:"matrix_position,omitempty"`
}
type teamPod struct {
	ID    string     `json:"id"`
	Title string     `json:"title"`
	Roles []teamRole `json:"roles"`
}
type teamRole struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Staffing string `json:"staffing,omitempty"`
}
type teamReportingLine struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Relationship string `json:"relationship"`
}
type teamMatrix struct {
	Columns []string        `json:"columns"`
	Rows    []teamMatrixRow `json:"rows"`
}
type teamMatrixRow struct {
	Label  string   `json:"label"`
	Values []string `json:"values"`
}

// buildTeam lowers a variable pod/role roster, routed reporting relationships,
// and an optional fixed-column responsibility matrix into measured compose parts.
func buildTeam(c Context, raw json.RawMessage) (Content, error) {
	var in teamInput
	if err := Decode(raw, &in); err != nil {
		return Content{}, fmt.Errorf("team content: %w", err)
	}
	if len(in.Pods) < 1 || len(in.Pods) > 6 {
		return Content{}, fmt.Errorf("team pods must contain 1–6 entries")
	}
	ids := map[string]bool{}
	roles := map[string]bool{}
	staffUsed := false
	for _, p := range in.Pods {
		if !adaptID(p.ID) || ids[p.ID] || strings.TrimSpace(p.Title) == "" || len(p.Roles) < 1 || len(p.Roles) > 8 {
			return Content{}, fmt.Errorf("pod %q needs a unique id, title and 1–8 roles", p.ID)
		}
		ids[p.ID] = true
		for _, r := range p.Roles {
			if !adaptID(r.ID) || roles[r.ID] || strings.TrimSpace(r.Label) == "" {
				return Content{}, fmt.Errorf("role %q has duplicate/invalid id or empty label", r.ID)
			}
			roles[r.ID] = true
			if r.Staffing != "" && !validStaffing(r.Staffing) {
				return Content{}, fmt.Errorf("role %q staffing must be one of staffing.wm_full_time, staffing.wm_part_time, staffing.client_part_time", r.ID)
			}
			staffUsed = staffUsed || r.Staffing != ""
		}
	}
	for _, r := range in.ReportingLines {
		if !ids[r.From] || !ids[r.To] || r.From == r.To || r.Relationship != "reporting" && r.Relationship != "dependency" && r.Relationship != "advisory" && r.Relationship != "annotation" {
			return Content{}, fmt.Errorf("reporting line %q → %q must reference distinct known pod IDs and use a supported relationship type", r.From, r.To)
		}
	}
	if len(in.ReportingLines) > 12 {
		return Content{}, fmt.Errorf("team supports at most 12 reporting lines")
	}
	if in.Responsibilities != nil {
		m := in.Responsibilities
		if in.MatrixPosition == "" {
			in.MatrixPosition = "right"
		}
		maxCols := 4
		if in.MatrixPosition == "bottom" {
			maxCols = 6
		}
		if len(m.Columns) < 2 || len(m.Columns) > maxCols || len(m.Rows) < 1 || len(m.Rows) > 6 {
			return Content{}, fmt.Errorf("responsibilities at %s need 2–%d columns and 1–6 rows", in.MatrixPosition, maxCols)
		}
		for _, r := range m.Rows {
			if strings.TrimSpace(r.Label) == "" || len(r.Values) != len(m.Columns) {
				return Content{}, fmt.Errorf("each responsibility row needs a label and one value per column")
			}
		}
	} else if in.MatrixPosition != "" {
		return Content{}, fmt.Errorf("matrix_position requires responsibilities")
	}
	if in.Responsibilities != nil && in.MatrixPosition != "right" && in.MatrixPosition != "bottom" {
		return Content{}, fmt.Errorf("matrix_position must be right or bottom")
	}
	b := c.Bounds
	st := c.Style
	titleColor := styleColor(st.Navy, "#17365D")
	accent := styleColor(st.Accent, "#1877A8")
	surface := styleColor(st.Surface, "#FFFFFF")
	muted := styleColor(st.MutedSurface, "#EDF3F8")
	fs := styleSize(st.BodyFontPt, 11)
	podTitleFS := styleSize(st.LabelFontPt, 10.5)
	matrixH := 0.0
	legendH := 0.0
	if staffUsed {
		legendH = 22
	}
	podsH := b.Height - legendH
	podsW, matrixX, matrixW := b.Width, b.X, b.Width
	if in.Responsibilities != nil {
		if in.MatrixPosition == "right" {
			podsW = b.Width * .51
			matrixX = b.X + podsW + 8
			matrixW = b.Width - podsW - 8
			matrixH = b.Height - legendH
		} else {
			matrixH = maxFloat(b.Height*.34, float64(len(in.Responsibilities.Rows)+1)*17+8)
			podsH = b.Height - matrixH - 8 - legendH
		}
		if podsH < 90 {
			return Content{}, fmt.Errorf("responsibility matrix leaves less than 90pt for team pods")
		}
	}
	columns := 2
	if len(in.Pods) == 1 {
		columns = 1
	}
	if len(in.Pods) >= 5 {
		columns = 3
	}
	rows := (len(in.Pods) + columns - 1) / columns
	gap := 10.0
	podW := (podsW - float64(columns-1)*gap) / float64(columns)
	podH := (podsH - float64(rows-1)*gap) / float64(rows)
	out := Content{Controls: map[string]any{"pod_count": len(in.Pods), "role_count": countRoles(in.Pods), "reporting_line_count": len(in.ReportingLines), "responsibility_rows": matrixRows(in.Responsibilities), "responsibility_columns": matrixColumns(in.Responsibilities), "matrix_position": in.MatrixPosition}, Limitations: []string{"Pod and role counts are variable within declared bounds; independent subfields per role and free-form matrix topology are not represented."}}
	positions := map[string][2]int{}
	for i, p := range in.Pods {
		positions[p.ID] = [2]int{i / columns, i % columns}
		x := b.X + float64(i%columns)*(podW+gap)
		y := b.Y + float64(i/columns)*(podH+gap)
		rolesOut := make([]compose.RoleSpec, len(p.Roles))
		for j, r := range p.Roles {
			bg := r.Staffing
			if bg == "" {
				bg = muted
			}
			rolesOut[j] = compose.RoleSpec{ID: r.ID, Label: r.Label, Background: bg, Foreground: "auto"}
		}
		out.Pods = append(out.Pods, compose.PodSpec{ID: p.ID, Title: p.Title, Bounds: compose.PodBounds{X: x, Y: y, Width: podW, MaxHeightPt: podH}, Layout: compose.LayoutSpec{Columns: 1, GapPt: 4, PaddingPt: 7, TitleGapPt: 5, MinTileWidthPt: podW - 16}, Style: compose.PodStyle{FontFace: st.FontFace, FontSizePt: fs, Bold: false, TitleFontFace: st.FontFace, TitleFontSizePt: podTitleFS, TitleBold: true, TileMinHeightPt: 22, HorizontalInsetPt: 5, VerticalInsetPt: 3, ParagraphGapPt: 0, Surface: surface, Foreground: st.Ink(surface), TitleForeground: st.Ink(surface)}, Roles: rolesOut})
	}
	for i, r := range in.ReportingLines {
		from, to := positions[r.From], positions[r.To]
		fromAnchor, toAnchor, direction := reportingAnchors(from, to)
		out.Connections = append(out.Connections, compose.ConnectionSpec{ID: fmt.Sprintf("report-%02d", i+1), From: r.From, To: r.To, Relationship: r.Relationship, FromAnchor: fromAnchor, ToAnchor: toAnchor, Color: accent, WidthPt: 1.2, ClearancePt: 2, PreferredDirection: direction})
	}
	if in.Responsibilities != nil {
		m := in.Responsibilities
		ncol := len(m.Columns) + 1
		nrow := len(m.Rows) + 1
		cols := make([]compose.TrackSpec, ncol)
		cols[0] = compose.TrackSpec{Weight: 1.35, MinPt: 110}
		for i := 1; i < ncol; i++ {
			minW := 65.0
			if in.MatrixPosition == "bottom" {
				minW = 75
			}
			cols[i] = compose.TrackSpec{Weight: 1, MinPt: minW}
		}
		rs := make([]compose.TrackSpec, nrow)
		for i := range rs {
			rs[i] = compose.TrackSpec{Weight: 1, MinPt: 17}
		}
		cells := make([]compose.CellSpec, 0, nrow*ncol)
		for col, label := range append([]string{"RESPONSIBILITY"}, m.Columns...) {
			cells = append(cells, compose.CellSpec{ID: fmt.Sprintf("head-%d", col), Row: 0, Column: col, Padding: compose.Insets{Top: 2, Bottom: 2, Left: 4, Right: 4}, Background: titleColor, Valign: "middle", Blocks: []compose.BlockSpec{{ID: "label", Text: label, FontFace: st.FontFace, FontSizePt: styleSize(st.LabelFontPt, 9), Bold: true, Foreground: "#FFFFFF", Align: "left", Valign: "middle"}}})
		}
		for ri, r := range m.Rows {
			vals := append([]string{r.Label}, r.Values...)
			for ci, v := range vals {
				bg := surface
				if (ri+ci)%2 == 1 {
					bg = muted
				}
				cells = append(cells, compose.CellSpec{ID: fmt.Sprintf("row-%d-col-%d", ri, ci), Row: ri + 1, Column: ci, Padding: compose.Insets{Top: 2, Bottom: 2, Left: 4, Right: 4}, Background: bg, Valign: "middle", Blocks: []compose.BlockSpec{{ID: "value", Text: v, FontFace: st.FontFace, FontSizePt: styleSize(st.LabelFontPt, 9), Bold: ci == 0, Foreground: st.Ink(bg), Align: "left", Valign: "middle"}}})
			}
		}
		matrixY := b.Y + podsH + 8
		if in.MatrixPosition == "right" {
			matrixY = b.Y
		}
		out.Layouts = append(out.Layouts, compose.ContainerSpec{ID: "responsibility-matrix", Bounds: compose.Rect{X: matrixX, Y: matrixY, Width: matrixW, Height: matrixH}, Padding: compose.Insets{Top: 1, Bottom: 1, Left: 1, Right: 1}, Columns: cols, Rows: rs, ColumnGapPt: 1, RowGapPt: 1, Layer: 0, Cells: cells})
	}
	if staffUsed {
		out.Legend = &compose.LegendSpec{Bounds: compose.Rect{X: b.X, Y: b.Y + b.Height - 19, Width: b.Width, Height: 17}, FontFace: st.FontFace, FontSizePt: 8.5, Bold: true, Foreground: st.Ink(st.White), SwatchSizePt: 8, GapPt: 4, ItemGapPt: 14}
	}
	return out, nil
}
func countRoles(ps []teamPod) int {
	n := 0
	for _, p := range ps {
		n += len(p.Roles)
	}
	return n
}
func matrixRows(m *teamMatrix) int {
	if m == nil {
		return 0
	}
	return len(m.Rows)
}
func matrixColumns(m *teamMatrix) int {
	if m == nil {
		return 0
	}
	return len(m.Columns)
}
func validStaffing(s string) bool {
	return s == "staffing.wm_full_time" || s == "staffing.wm_part_time" || s == "staffing.client_part_time"
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func reportingAnchors(from, to [2]int) (string, string, string) {
	if from[0] == to[0] {
		if from[1] < to[1] {
			return "right", "left", "horizontal"
		}
		return "left", "right", "horizontal"
	}
	if from[0] < to[0] {
		return "bottom", "top", "vertical"
	}
	return "top", "bottom", "vertical"
}
