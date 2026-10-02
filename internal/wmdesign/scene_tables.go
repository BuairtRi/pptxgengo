package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

type sceneTableColumn struct {
	Key   string  `json:"k"`
	Label string  `json:"label"`
	Width float64 `json:"w"`
	Type  string  `json:"type,omitempty"`
}
type sceneTableSource struct {
	Type      string                       `json:"type"`
	X         float64                      `json:"x"`
	Y         float64                      `json:"y"`
	W         float64                      `json:"w"`
	Header    string                       `json:"header,omitempty"`
	Dense     bool                         `json:"dense,omitempty"`
	RowH      float64                      `json:"rowH,omitempty"`
	RowHeader bool                         `json:"rowHeader,omitempty"`
	Preset    string                       `json:"preset,omitempty"`
	Highlight string                       `json:"highlight,omitempty"`
	Continued string                       `json:"continued,omitempty"`
	DeltaUnit string                       `json:"deltaUnit,omitempty"`
	Columns   []sceneTableColumn           `json:"cols"`
	Rows      []map[string]json.RawMessage `json:"rows"`
	Groups    []struct {
		Label string `json:"label"`
		From  int    `json:"from"`
		To    int    `json:"to"`
	} `json:"groups,omitempty"`
	RunRate *struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"runRate,omitempty"`
}

func (r *renderer) planTableScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "table" {
		return nil, false, nil
	}
	var n sceneTableSource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	p, err := r.sceneTable(id, n, ctx)
	return p, true, err
}

func (r *renderer) sceneTable(id string, n sceneTableSource, ctx SceneContext) (*scenePlan, error) {
	if n.W <= 0 || len(n.Columns) < 1 || len(n.Columns) > 12 || len(n.Rows) > 60 {
		return nil, fmt.Errorf("scene.table_invalid_dimensions_or_count: %s", id)
	}
	if n.Header != "" && n.Header != "dark" && n.Header != "light" {
		return nil, fmt.Errorf("scene.table_invalid_header: %s", n.Header)
	}
	switch n.Preset {
	case "", "fees", "capacity", "raci", "ratecard", "requirements", "compare", "comparison", "scorecard", "status", "maturity", "allocation":
	default:
		return nil, fmt.Errorf("scene.table_unknown_preset: %s", n.Preset)
	}
	if n.DeltaUnit != "" && n.DeltaUnit != "$M" {
		return nil, fmt.Errorf("scene.table_unknown_delta_unit: %s", n.DeltaUnit)
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	bodyName := "body"
	headH, rowH := 36., 36.
	if n.Dense {
		bodyName = "small"
		headH, rowH = 28, 28
	}
	if n.RowH != 0 {
		rowH = n.RowH
	}
	if rowH <= 0 {
		return nil, fmt.Errorf("scene.table_invalid_row_height")
	}
	cols := map[string]int{}
	xs := make([]float64, len(n.Columns))
	colW := make([]float64, len(n.Columns))
	acc := n.X
	hi := -1
	supported := map[string]bool{"": true, "num": true, "delta": true, "status": true, "rating": true, "allocation": true, "harvey": true, "maturity": true, "gauge": true, "tag": true, "raci": true, "bullets": true, "icon": true, "check": true}
	for i, c := range n.Columns {
		if !validPartKey(c.Key) || c.Width <= 24 || cols[c.Key] != 0 || !supported[c.Type] {
			return nil, fmt.Errorf("scene.table_invalid_column: %s", c.Key)
		}
		cols[c.Key] = i + 1
		xs[i] = acc
		colW[i] = c.Width / 72
		acc += c.Width
		if c.Key == n.Highlight {
			hi = i
		}
	}
	if math.Abs(acc-n.X-n.W) > .02 {
		return nil, fmt.Errorf("scene.table_width_sum: %s", id)
	}
	if n.Highlight != "" && hi < 0 {
		return nil, fmt.Errorf("scene.table_unknown_highlight")
	}
	base, _ := r.sceneStyle(bodyName)
	font, err := r.typeEngine.Resolve(base)
	if err != nil {
		return nil, err
	}
	bg, err := r.sceneColor(surface, "bg")
	if err != nil {
		return nil, err
	}
	line, err := r.sceneColor(surface, "line")
	if err != nil {
		return nil, err
	}
	strong, err := r.sceneColor(surface, "strong")
	if err != nil {
		return nil, err
	}
	p := &scenePlan{ID: id}
	tab := &sceneTable{ID: id + ".native", Options: pptx.TableProps{TextBaseProps: pptx.TextBaseProps{FontFace: font.Typeface, FontSize: base.Size, Color: strong, Valign: "middle"}, AutoPage: ptrSceneBool(false), ColW: colW, Margin: pptx.Margin{0}, Border: []pptx.BorderProps{{Type: "none", Color: line}}, Fill: &pptx.ShapeFillProps{Color: bg}}}
	p.Items = append(p.Items, sceneItem{Table: tab})
	y := n.Y
	// A spanning heading is a real merged native table row, including blanks.
	if len(n.Groups) > 0 {
		groups := map[int]struct {
			to    int
			label string
		}{}
		used := map[int]bool{}
		for _, g := range n.Groups {
			if g.From < 0 || g.To < g.From || g.To >= len(n.Columns) || g.Label == "" {
				return nil, fmt.Errorf("scene.table_invalid_group")
			}
			for i := g.From; i <= g.To; i++ {
				if used[i] {
					return nil, fmt.Errorf("scene.table_overlapping_groups")
				}
				used[i] = true
			}
			groups[g.From] = struct {
				to    int
				label string
			}{g.To, g.Label}
		}
		var row pptx.TableRow
		for c := 0; c < len(n.Columns); {
			w := n.Columns[c].Width
			text := ""
			span := 1
			if g, ok := groups[c]; ok {
				text = g.label
				span = g.to - c + 1
				for j := c + 1; j <= g.to; j++ {
					w += n.Columns[j].Width
				}
			}
			st, _ := r.sceneDataToken("label", 600)
			cell, tr, err := r.sceneNativeCell(id+fmt.Sprintf(".group.%s", n.Columns[c].Key), text, st, Rect{xs[c], y, w, 24}, surface, "primary", "center", 12, 12, ctx)
			if err != nil {
				return nil, err
			}
			cell.Options.Colspan = span
			cell.Options.Border = []pptx.BorderProps{{Type: "none"}, {Type: "none"}, {Type: "solid", Color: strong, Pt: 1}, {Type: "none"}}
			tab.CellRecords = append(tab.CellRecords, sceneTableCell{len(tab.Rows), len(row), tr})
			row = append(row, cell)
			c += span
		}
		tab.Rows = append(tab.Rows, row)
		tab.Options.RowH = append(tab.Options.RowH, 24./72)
		y += 24
	}
	var headers pptx.TableRow
	for i, c := range n.Columns {
		st, _ := r.sceneDataToken("label", 600)
		on, role := surface, "primary"
		if n.Header == "dark" {
			on = "inverse"
		} else if i == hi {
			role = "emphasis"
		}
		cell, tr, err := r.sceneNativeCell(id+".header."+c.Key, c.Label, st, Rect{xs[i], y, c.Width, headH}, on, role, sceneTableAlign(c.Type), 12, 12, ctx)
		if err != nil {
			return nil, err
		}
		if n.Header != "dark" {
			cell.Options.Border = []pptx.BorderProps{{Type: "none"}, {Type: "none"}, {Type: "solid", Color: strong, Pt: 1.5}, {Type: "none"}}
		}
		if i == hi && n.Header != "dark" {
			color, _ := r.sceneColor("subtle", "bg")
			cell.Options.Fill = &pptx.ShapeFillProps{Color: color}
		}
		headers = append(headers, cell)
		tab.CellRecords = append(tab.CellRecords, sceneTableCell{len(tab.Rows), i, tr})
	}
	tab.Rows = append(tab.Rows, headers)
	tab.Options.RowH = append(tab.Options.RowH, headH/72)
	y += headH
	for ri, values := range n.Rows {
		key, err := sceneDataKey(ctx, "rows", ri)
		if err != nil {
			return nil, err
		}
		for k := range values {
			if cols[k] == 0 && k != "group" && k != "total" {
				return nil, fmt.Errorf("scene.table_unknown_row_field: %s", k)
			}
		}
		var group string
		if raw, ok := values["group"]; ok {
			if err := json.Unmarshal(raw, &group); err != nil || group == "" {
				return nil, fmt.Errorf("scene.table_invalid_group_row")
			}
		}
		var total bool
		if raw, ok := values["total"]; ok {
			if err := json.Unmarshal(raw, &total); err != nil {
				return nil, err
			}
		}
		if group != "" {
			if total || len(values) != 1 {
				return nil, fmt.Errorf("scene.table_group_row_content_union")
			}
			st, _ := r.sceneDataToken("label", 600)
			on := "subtle"
			if hi >= 0 {
				on = "light"
			}
			cell, tr, err := r.sceneNativeCell(id+".row."+key+".group", group, st, Rect{n.X, y, n.W, rowH}, on, "emphasis", "left", 12, 12, ctx)
			if err != nil {
				return nil, err
			}
			cell.Options.Colspan = len(n.Columns)
			cell.Options.Border = []pptx.BorderProps{{Type: "none"}, {Type: "none"}, {Type: "solid", Color: line, Pt: .75}, {Type: "none"}}
			tab.CellRecords = append(tab.CellRecords, sceneTableCell{len(tab.Rows), 0, tr})
			tab.Rows = append(tab.Rows, pptx.TableRow{cell})
		} else {
			var row pptx.TableRow
			for ci, c := range n.Columns {
				iid := id + ".row." + key + "." + c.Key
				on := surface
				if ci == hi || total {
					on = "subtle"
				}
				st := base
				if n.RowHeader && ci == 0 || total {
					st.Weight = 600
				}
				cb := Rect{xs[ci], y, c.Width, rowH}
				v := values[c.Key]
				cell, tr, err := r.sceneTableValue(p, iid, c, v, st, cb, on, n.DeltaUnit, ctx, fmt.Sprintf("rows/%d/%s", ri, c.Key))
				if err != nil {
					return nil, err
				}
				cell.Options.Border = []pptx.BorderProps{{Type: "none"}, {Type: "none"}, {Type: "solid", Color: line, Pt: .75}, {Type: "none"}}
				if total {
					cell.Options.Border = []pptx.BorderProps{{Type: "solid", Color: strong, Pt: 1.5}, {Type: "none"}, {Type: "none"}, {Type: "none"}}
				}
				row = append(row, cell)
				tab.CellRecords = append(tab.CellRecords, sceneTableCell{len(tab.Rows), ci, tr})
			}
			tab.Rows = append(tab.Rows, row)
		}
		tab.Options.RowH = append(tab.Options.RowH, rowH/72)
		y += rowH
	}
	if n.RunRate != nil {
		st := base
		st.Weight = 600
		rw := n.Columns[len(n.Columns)-1].Width
		label, tr, err := r.sceneNativeCell(id+".run-rate.label", n.RunRate.Label, st, Rect{n.X, y, n.W - rw, rowH}, "inverse", "primary", "left", 12, 0, ctx)
		if err != nil {
			return nil, err
		}
		label.Options.Colspan = len(n.Columns) - 1
		st.Family = "IBM Plex Mono"
		value, vr, err := r.sceneNativeCell(id+".run-rate.value", n.RunRate.Value, st, Rect{n.X + n.W - rw, y, rw, rowH}, "inverse", "display", "right", 0, 12, ctx)
		if err != nil {
			return nil, err
		}
		tab.CellRecords = append(tab.CellRecords, sceneTableCell{len(tab.Rows), 0, tr}, sceneTableCell{len(tab.Rows), 1, vr})
		tab.Rows = append(tab.Rows, pptx.TableRow{label, value})
		tab.Options.RowH = append(tab.Options.RowH, rowH/72)
		y += rowH
	}
	tab.Rect = Rect{n.X, n.Y, n.W, y - n.Y}
	sceneDataBounds(p, tab.Rect)
	if hi >= 0 {
		if err := r.sceneDataInkShape(p, id+".highlight-edge", Rect{xs[hi], n.Y, n.Columns[hi].Width, 3}, surface, "emphasis"); err != nil {
			return nil, err
		}
		p.Items[len(p.Items)-1].Shape.Record.Rect.Y = n.Y - 3
		if len(n.Groups) > 0 {
			p.Items[len(p.Items)-1].Shape.Record.Rect.Y = n.Y + 21
		}
		p.Items[len(p.Items)-1].Shape.Props.PositionProps = pos(p.Items[len(p.Items)-1].Shape.Record.Rect)
	}
	if n.Continued != "" {
		st, _ := r.sceneStyle("label")
		if _, err := r.sceneDataText(p, id+".continued", n.Continued, st, Rect{n.X, n.Y - 18, n.W, 0}, surface, "secondary", "left"); err != nil {
			return nil, err
		}
	}
	if n.Preset == "raci" {
		if err := r.sceneRACILegend(p, id+".legend", Rect{n.X, y + 12, n.W, 18}, surface); err != nil {
			return nil, err
		}
	}
	sceneDataGroup(p, id, "table.native.source", 0, p.Bounds)
	p.Warnings = append(p.Warnings, "Native table typography and anchored cell adornments require native review; adornments retain their authored coordinates after cell text edits.")
	return p, nil
}

func sceneTableAlign(kind string) string {
	switch kind {
	case "num", "delta":
		return "right"
	case "check", "harvey", "raci", "gauge":
		return "center"
	}
	return "left"
}

// Cell records include empty cells. They supply exact leading, fonts and wrap
// decisions to the native table XML pass instead of the writer's heuristics.
func (r *renderer) sceneNativeCell(id, text string, st Style, b Rect, surface, role, align string, left, right float64, ctx SceneContext) (pptx.TableCell, TextRecord, error) {
	var cell pptx.TableCell
	var tr TextRecord
	if b.W-left-right <= 0 {
		return cell, tr, fmt.Errorf("scene.table_cell_negative_width: %s", id)
	}
	if strings.Contains(text, "[[") || strings.Contains(text, "]]") {
		return cell, tr, fmt.Errorf("scene.table_cell_marks_unsupported: %s", id)
	}
	pp := &scenePlan{}
	if strings.Contains(text, "[^") {
		if err := r.primitiveRichText(pp, id, text, st, Rect{b.X + left, b.Y, b.W - left - right, b.H}, surface, role, align, "", "", ctx); err != nil {
			return cell, tr, err
		}
		tr = *pp.Items[len(pp.Items)-1].Text
	} else {
		l, err := r.typeEngine.Measure(text, st, b.W-left-right)
		if err != nil {
			return cell, tr, err
		}
		need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
		if text != "" && need > b.H+.02 {
			return cell, tr, fmt.Errorf("scene.table_cell_overflow: %s needs%.3fpt capacity%.3fpt", id, need, b.H)
		}
		color, err := r.sceneColor(surface, role)
		if err != nil {
			return cell, tr, err
		}
		tr = TextRecord{ID: id, Rect: Rect{b.X + left, b.Y, b.W - left - right, b.H}, Color: color, Align: align, Layout: l}
	}
	bg, err := r.sceneColor(surface, "bg")
	if err != nil {
		return cell, tr, err
	}
	cell = pptx.TableCell{Text: tr.Layout.Displayed, Options: &pptx.TableCellProps{TextBaseProps: pptx.TextBaseProps{FontFace: tr.Layout.Font.Typeface, FontSize: st.Size, Bold: &tr.Layout.Font.Bold, Italic: &tr.Layout.Font.NativeItalic, Color: tr.Color, Align: align, Valign: "middle"}, Margin: pptx.Margin{0, right / 72, 0, left / 72}, Fill: &pptx.ShapeFillProps{Color: bg}, Border: []pptx.BorderProps{{Type: "none"}}}}
	return cell, tr, nil
}

func (r *renderer) sceneTableValue(p *scenePlan, id string, c sceneTableColumn, raw json.RawMessage, st Style, b Rect, surface, deltaUnit string, ctx SceneContext, path string) (pptx.TableCell, TextRecord, error) {
	var text string
	left, right := 12., 12.
	align := sceneTableAlign(c.Type)
	blank := len(raw) == 0 || string(raw) == "null" || string(raw) == `""`
	if blank {
		return r.sceneNativeCell(id, "", st, b, surface, "primary", align, left, right, ctx)
	}
	str := func() error { return json.Unmarshal(raw, &text) }
	number := func() (float64, error) {
		var v float64
		err := json.Unmarshal(raw, &v)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, fmt.Errorf("scene.table_requires_finite_number: %s", id)
		}
		return v, nil
	}
	switch c.Type {
	case "":
		if err := str(); err != nil {
			return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_text_requires_string: %s", id)
		}
	case "num":
		st.Family = "IBM Plex Mono"
		st.Weight = 600
		if err := str(); err != nil {
			var format NumberFormatSpec
			if err := sceneDecode(raw, &format); err != nil {
				return pptx.TableCell{}, TextRecord{}, err
			}
			var err error
			text, err = FormatNumber(format)
			if err != nil {
				return pptx.TableCell{}, TextRecord{}, err
			}
		}
	case "delta":
		st.Family = "IBM Plex Mono"
		st.Weight = 400
		v, err := number()
		if err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		text = "—"
		if v != 0 {
			mag := strconv.FormatFloat(math.Abs(v), 'f', 1, 64)
			sign := "+"
			if v < 0 {
				sign = "−"
			}
			text = sign + mag
			if deltaUnit == "$M" {
				text = sign + "$" + mag + "M"
			}
			layout, err := r.typeEngine.Measure(text, st, b.W-24)
			if err != nil {
				return pptx.TableCell{}, TextRecord{}, err
			}
			if len(layout.Lines) != 1 {
				return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_delta_wrap")
			}
			mark := Rect{b.X + b.W - 12 - layout.Lines[0].Advance - 13, b.Y + (b.H-7)/2, 7, 7}
			col, _ := r.sceneColor(surface, "emphasis")
			r.sceneDataShape(p, id+".direction", mark, pptx.ShapeTypeTriangle, col, nil)
			if v < 0 {
				p.Items[len(p.Items)-1].Shape.Props.Rotate = 180
				p.Items[len(p.Items)-1].Shape.Record.Rotation = 180
			}
			left += 13
		}
	case "status":
		if err := str(); err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		label, ok := map[string]string{"on": "On track", "risk": "At risk", "off": "Off track"}[text]
		if !ok {
			return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_status_enum: %s", text)
		}
		if err := r.sceneTableStatus(p, id, Rect{b.X + 12, b.Y + (b.H-8)/2, 8, 8}, surface, text); err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		text = label
		left += 14
		st, _ = r.sceneStyle("small")
	case "rating", "allocation", "harvey", "gauge":
		v, err := number()
		if err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		if err = r.sceneTableNumericMark(p, id, c.Type, v, b, surface); err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		text = ""
		if c.Type == "allocation" {
			st, _ = r.sceneDataToken("label", 600)
			text = strconv.Itoa(int(math.Floor(v*100+.5))) + "%"
			left += 66
		}
	case "maturity":
		var pair []int
		if err := json.Unmarshal(raw, &pair); err != nil || len(pair) != 2 || pair[0] < 0 || pair[0] > 5 || pair[1] < 1 || pair[1] > 5 {
			return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_invalid_maturity")
		}
		for i := 1; i <= 5; i++ {
			fill := "bg"
			if i <= pair[0] {
				fill = "strong"
			}
			col, _ := r.sceneColor(surface, fill)
			line, _ := r.sceneColor(surface, "strong")
			width := .75
			if i == pair[1] {
				line = "F900D3"
				width = 1.5
			}
			r.sceneDataShape(p, fmt.Sprintf("%s.stage-%d", id, i), Rect{b.X + 12 + float64(i-1)*17, b.Y + (b.H-10)/2, 14, 10}, pptx.ShapeTypeRect, col, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: line}, Width: width})
		}
		st, _ = r.sceneStyle("label")
		text = fmt.Sprintf("%d → %d", pair[0], pair[1])
		left += 88
	case "tag":
		if err := str(); err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		st, _ = r.sceneStyle("label")
		l, err := r.typeEngine.Measure(text, st, b.W-32)
		if err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		if len(l.Lines) != 1 {
			return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_tag_wrap")
		}
		strong, _ := r.sceneColor(surface, "strong")
		r.sceneDataShape(p, id+".outline", Rect{b.X + 12, b.Y + (b.H-st.Leading)/2, l.Lines[0].Advance + 10, st.Leading}, pptx.ShapeTypeRect, strong, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: strong}, Width: .75})
		p.Items[len(p.Items)-1].Shape.Props.Fill = &pptx.ShapeFillProps{Type: "none"}
		p.Warnings = append(p.Warnings, "user-review.v1/native-tag-transparent: "+id+" outline does not cover native cell text; fixed2pt horizontal reserve.")
		left += 4
	case "raci":
		if err := str(); err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		parts := strings.Split(text, "/")
		if len(parts) > 4 {
			return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_raci_count")
		}
		width := 18.*float64(len(parts)) + 6*float64(len(parts)-1)
		if width > b.W-24 {
			return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_raci_width")
		}
		for i, t := range parts {
			if err := r.sceneRACISquare(p, fmt.Sprintf("%s.role-%d", id, i+1), t, Rect{b.X + (b.W-width)/2 + 24*float64(i), b.Y + (b.H-18)/2, 18, 18}); err != nil {
				return pptx.TableCell{}, TextRecord{}, err
			}
		}
		text = ""
	case "bullets":
		return r.sceneTableBulletCell(id, raw, b, surface, ctx, path)
	case "icon":
		var name string
		if err := json.Unmarshal(raw, &name); err != nil {
			var value struct {
				Icon string `json:"icon"`
				Text string `json:"text,omitempty"`
			}
			if err = sceneDecode(raw, &value); err != nil {
				return pptx.TableCell{}, TextRecord{}, err
			}
			name, text = value.Icon, value.Text
		}
		size := math.Min(36, b.H-12)
		if err := r.sceneIcon(p, id+".icon", name, Rect{b.X + 12, b.Y + (b.H-size)/2, size, size}, surface, "strong"); err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		left += size + 9
	case "check":
		var on bool
		if err := json.Unmarshal(raw, &on); err != nil {
			return pptx.TableCell{}, TextRecord{}, err
		}
		if on {
			color, _ := r.sceneColor(surface, "strong")
			r.sceneDataShape(p, id+".checked", Rect{b.X + (b.W-12)/2, b.Y + (b.H-12)/2, 12, 12}, pptx.ShapeTypeRect, color, nil)
			r.sceneDataCheck(p, id+".tick", Rect{b.X + (b.W-12)/2, b.Y + (b.H-12)/2, 12, 12}, "FFFFFF")
		} else {
			if err := r.sceneDataInkShape(p, id+".dash", Rect{b.X + (b.W-9)/2, b.Y + (b.H-1)/2, 9, 1}, surface, "secondary"); err != nil {
				return pptx.TableCell{}, TextRecord{}, err
			}
		}
		text = ""
	default:
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_unknown_cell_type: %s", c.Type)
	}
	return r.sceneNativeCell(id, text, st, b, surface, "primary", align, left, right, ctx)
}

func (r *renderer) sceneDataCheck(p *scenePlan, id string, b Rect, color string) {
	sh := sceneShape{Type: pptx.ShapeTypeCustGeom, Props: pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Points: []pptx.ShapePoint{{X: pptx.Inches(3. / 72), Y: pptx.Inches(6.2 / 72), MoveTo: ptrSceneBool(true)}, {X: pptx.Inches(5. / 72), Y: pptx.Inches(8.2 / 72)}, {X: pptx.Inches(9. / 72), Y: pptx.Inches(3.8 / 72)}}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: 1.4}}, Record: ShapeRecord{ID: id, Rect: b, Color: color, Geometry: "custGeom"}}
	p.Items = append(p.Items, sceneItem{Shape: &sh})
	sceneDataBounds(p, b)
}
func (r *renderer) sceneTableStatus(p *scenePlan, id string, b Rect, surface, status string) error {
	strong, _ := r.sceneColor("light", "strong")
	color, err := r.sceneColor(surface, "kpi."+status)
	if err != nil {
		return err
	}
	r.sceneDataShape(p, id+".status-border", b, pptx.ShapeTypeRect, strong, nil)
	r.sceneDataShape(p, id+".status-fill", Rect{b.X + 1, b.Y + 1, b.W - 2, b.H - 2}, pptx.ShapeTypeRect, color, nil)
	return nil
}
func (r *renderer) sceneTableNumericMark(p *scenePlan, id, kind string, v float64, b Rect, surface string) error {
	strong, _ := r.sceneColor("light", "strong")
	bg, _ := r.sceneColor(surface, "bg")
	outline := &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: strong}, Width: .75}
	switch kind {
	case "rating":
		if v != math.Trunc(v) || v < 0 || v > 4 {
			return fmt.Errorf("scene.table_rating_range")
		}
		for i := 0; i < 4; i++ {
			fill := bg
			if float64(i) < v {
				fill = strong
			}
			r.sceneDataShape(p, fmt.Sprintf("%s.rating-%d", id, i+1), Rect{b.X + 12 + float64(i)*11, b.Y + (b.H-8)/2, 8, 8}, pptx.ShapeTypeRect, fill, outline)
		}
	case "allocation":
		if v < 0 || v > 1 {
			return fmt.Errorf("scene.table_allocation_range")
		}
		bar := Rect{b.X + 12, b.Y + (b.H-8)/2, 60, 8}
		r.sceneDataShape(p, id+".bar", bar, pptx.ShapeTypeRect, bg, nil)
		if v > 0 {
			r.sceneDataShape(p, id+".allocation", Rect{bar.X, bar.Y, bar.W * v, bar.H}, pptx.ShapeTypeRect, strong, nil)
		}
		r.sceneDataShape(p, id+".bar-outline", bar, pptx.ShapeTypeRect, strong, outline)
		p.Items[len(p.Items)-1].Shape.Props.Fill = &pptx.ShapeFillProps{Type: "none"}
		p.Warnings = append(p.Warnings, "user-review.v1/allocation-full-height: "+id+" fill shares wrapper top/bottom; outline paints above fill.")
	case "harvey":
		if v != math.Trunc(v) || v < 0 || v > 4 {
			return fmt.Errorf("scene.table_harvey_range")
		}
		box := Rect{b.X + (b.W-16)/2, b.Y + (b.H-16)/2, 16, 16}
		r.sceneDataShape(p, id+".harvey-outline", box, pptx.ShapeTypeEllipse, bg, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: strong}, Width: .96})
		if v > 0 {
			typ := pptx.ShapeTypePie
			if v == 4 {
				typ = pptx.ShapeTypeEllipse
			}
			r.sceneDataShape(p, id+".harvey-fill", box, typ, strong, nil)
			if v < 4 {
				p.Items[len(p.Items)-1].Shape.Props.AngleRange = &[2]float64{270, 270 + v*90}
			}
		}
	case "gauge":
		if v < 0 || v > 1 {
			return fmt.Errorf("scene.table_gauge_range")
		}
		box := Rect{b.X + (b.W-40)/2, b.Y + (b.H-22)/2, 40, 40}
		for i, role := range []string{"kpi.off", "kpi.risk", "kpi.risk", "kpi.on", "kpi.on"} {
			color, err := r.sceneColor(surface, role)
			if err != nil {
				return err
			}
			r.sceneDataShape(p, fmt.Sprintf("%s.segment-%d", id, i+1), box, pptx.ShapeTypeBlockArc, color, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: strong}, Width: .5})
			sh := p.Items[len(p.Items)-1].Shape
			sh.Props.AngleRange = &[2]float64{180 + float64(i)*36 + 1.1459, 180 + float64(i+1)*36 - 1.1459}
			sh.Props.ArcThicknessRatio = .45
		}
		a := math.Pi + v*math.Pi
		needle := sceneShape{Type: pptx.ShapeTypeLine, Props: pptx.ShapeProps{PositionProps: pos(Rect{box.X, box.Y, 40, 40}), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id + ".needle"}, Points: []pptx.ShapePoint{{X: pptx.Inches(20. / 72), Y: pptx.Inches(20. / 72), MoveTo: ptrSceneBool(true)}, {X: pptx.Inches((20 + 18*math.Cos(a)) / 72), Y: pptx.Inches((20 + 18*math.Sin(a)) / 72)}}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: strong}, Width: 1.6}}, Record: ShapeRecord{ID: id + ".needle", Rect: box, Color: strong, Geometry: "polyline"}}
		p.Items = append(p.Items, sceneItem{Shape: &needle})
		r.sceneDataShape(p, id+".hub", Rect{box.X + 17.6, box.Y + 17.6, 4.8, 4.8}, pptx.ShapeTypeEllipse, strong, nil)
	}
	return nil
}
func (r *renderer) sceneRACISquare(p *scenePlan, id, role string, b Rect) error {
	surface, ink := map[string]string{"R": "inverse", "A": "callout", "C": "light", "I": "subtle"}[role], "primary"
	if surface == "" {
		return fmt.Errorf("scene.table_raci_enum: %s", role)
	}
	if err := r.sceneRect(p, id+".square", b, surface); err != nil {
		return err
	}
	if role == "C" {
		bg, _ := r.sceneColor(surface, "bg")
		strong, _ := r.sceneColor(surface, "strong")
		r.sceneDataShape(p, id+".outline", b, pptx.ShapeTypeRect, bg, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: strong}, Width: .75})
	}
	st, _ := r.sceneDataToken("label", 600)
	_, err := r.sceneDataText(p, id+".text", role, st, Rect{b.X + 1, b.Y + 3, b.W - 2, 12}, surface, ink, "center")
	return err
}
func (r *renderer) sceneRACILegend(p *scenePlan, id string, b Rect, surface string) error {
	x := b.X
	st, _ := r.sceneStyle("small")
	for _, item := range []struct{ role, label string }{{"R", "Responsible"}, {"A", "Accountable"}, {"C", "Consulted"}, {"I", "Informed"}} {
		if err := r.sceneRACISquare(p, id+"."+item.role, item.role, Rect{x, b.Y, 18, 18}); err != nil {
			return err
		}
		l, err := r.typeEngine.Measure(item.label, st, b.W)
		if err != nil {
			return err
		}
		w := l.Lines[0].Advance + 1
		if x+24+w > b.X+b.W {
			return fmt.Errorf("scene.table_raci_legend_overflow")
		}
		if _, err := r.sceneDataText(p, id+"."+item.role+".label", item.label, st, Rect{x + 24, b.Y, w, 18}, surface, "primary", "left"); err != nil {
			return err
		}
		x += 24 + w + 18
	}
	return nil
}

// A bullet is a real native paragraph. Its square is 3pt and its measured text
// starts 12pt after the native cell's 12pt left margin.
func (r *renderer) sceneTableBulletCell(id string, raw json.RawMessage, b Rect, surface string, ctx SceneContext, path string) (pptx.TableCell, TextRecord, error) {
	var items []json.RawMessage
	if err := sceneDecode(raw, &items); err != nil || len(items) == 0 {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_empty_or_invalid_bullets")
	}
	st, err := r.sceneStyle("small")
	if err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	cell, tr, err := r.sceneNativeCell(id, "", st, b, surface, "primary", "left", 12, 12, ctx)
	if err != nil {
		return cell, tr, err
	}
	rich := RichTextLayout{Contract: RichTextContract}
	layout := tr.Layout
	layout.Lines = nil
	var originals, displays []string
	offset, firstBaseline, lastOccupied := 0., 0., 0.
	for i, item := range items {
		key, err := sceneDataKey(ctx, path, i)
		if err != nil {
			return cell, tr, err
		}
		var text string
		if err := json.Unmarshal(item, &text); err != nil || strings.TrimSpace(text) == "" {
			return cell, tr, fmt.Errorf("scene.table_bullet_requires_nonempty_string: %s", id)
		}
		if strings.Contains(text, "[[") || strings.Contains(text, "]]") {
			return cell, tr, fmt.Errorf("scene.table_cell_marks_unsupported: %s", id)
		}
		tmp := &scenePlan{}
		if err := r.primitiveRichText(tmp, id+"."+key, text, st, Rect{b.X + 24, b.Y, b.W - 36, 0}, surface, "primary", "left", "", "", ctx); err != nil {
			return cell, tr, err
		}
		var part TextRecord
		for _, it := range tmp.Items {
			if it.Text != nil {
				part = *it.Text
			}
		}
		if part.ID == "" {
			return cell, tr, fmt.Errorf("scene.table_bullet_text_record_missing")
		}
		var paragraph RichParagraphLayout
		if part.Rich == nil {
			paragraph = RichParagraphLayout{Key: key, Displayed: part.Layout.Displayed, Runs: []RichRunLayout{{Key: "text", Original: part.Layout.Original, Displayed: part.Layout.Displayed, Style: part.Layout.Style, Font: part.Layout.Font, Color: part.Color, Start: 0, End: len([]rune(part.Layout.Displayed))}}}
		} else {
			if len(part.Rich.Paragraphs) != 1 {
				return cell, tr, fmt.Errorf("scene.table_bullet_multiple_paragraphs")
			}
			paragraph = part.Rich.Paragraphs[0]
			paragraph.Key = key
		}
		paragraph.Bullet = true
		paragraph.FirstLine = len(layout.Lines)
		paragraph.LineCount = len(part.Layout.Lines)
		if i < len(items)-1 {
			paragraph.ParagraphGapAfter = 3
		}
		for _, line := range part.Layout.Lines {
			line.Baseline += offset
			layout.Lines = append(layout.Lines, line)
		}
		if i == 0 {
			firstBaseline = part.Layout.OccupiedTop
		}
		lastOccupied = offset + part.Layout.OccupiedTop + part.Layout.EstimatedOccupiedHeight
		offset += part.Layout.AllocationHeight + paragraph.ParagraphGapAfter
		rich.Paragraphs = append(rich.Paragraphs, paragraph)
		originals = append(originals, part.Layout.Original)
		displays = append(displays, part.Layout.Displayed)
	}
	layout.Original = strings.Join(originals, "\n")
	layout.Displayed = strings.Join(displays, "\n")
	layout.OccupiedTop = firstBaseline
	layout.AllocationHeight = offset
	layout.EstimatedOccupiedHeight = lastOccupied - firstBaseline
	layout.VerticalPolicy = "Source small bullet paragraphs, exact 18pt leading and 3pt gaps; native paragraph parity unqualified"
	if math.Max(layout.AllocationHeight, lastOccupied) > b.H+.02 {
		return cell, tr, fmt.Errorf("scene.table_bullet_overflow: %s needs%.3fpt capacity%.3fpt", id, math.Max(layout.AllocationHeight, lastOccupied), b.H)
	}
	tr.Layout = layout
	tr.Rich = &rich
	tr.Rect.X += 12
	tr.Rect.W -= 12
	cell.Text = layout.Displayed
	return cell, tr, nil
}
