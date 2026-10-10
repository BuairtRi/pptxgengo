package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

func sceneTableScoreRange(value, maximum float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.IsNaN(maximum) || math.IsInf(maximum, 0) || maximum != math.Trunc(maximum) || maximum < 1 || maximum > 12 || value != math.Trunc(value) || value < 0 || value > maximum {
		return fmt.Errorf("scene.table_score_range: integral value0_to_max and max1_to12 required")
	}
	return nil
}

type sceneTableScore struct {
	Value *float64        `json:"value"`
	Max   *float64        `json:"max,omitempty"`
	Ink   string          `json:"ink,omitempty"`
	Text  json.RawMessage `json:"text,omitempty"`
}

func sceneTableScoreValue(raw json.RawMessage, column sceneTableColumn, kind string) (sceneTableScore, error) {
	var n sceneTableScore
	var scalar float64
	if len(raw) > 0 && raw[0] != '{' {
		if e := json.Unmarshal(raw, &scalar); e != nil {
			return n, fmt.Errorf("scene.table_score_number")
		}
		n.Value = &scalar
	} else if e := sceneDecode(raw, &n); e != nil {
		return n, e
	}
	if n.Value == nil {
		return n, fmt.Errorf("scene.table_score_value_required")
	}
	maximum := 4.0
	if kind == "dots" {
		maximum = 5
	}
	if column.Max != nil {
		maximum = *column.Max
	}
	if n.Max != nil {
		maximum = *n.Max
	}
	if kind == "harvey" && (n.Max != nil || column.Max != nil || len(n.Text) > 0) {
		return n, fmt.Errorf("scene.table_harvey_fields: only value and ink")
	}
	if kind != "dots" && len(n.Text) > 0 {
		return n, fmt.Errorf("scene.table_score_text_only_dots")
	}
	if e := sceneTableScoreRange(*n.Value, maximum); e != nil {
		return n, e
	}
	n.Max = &maximum
	if n.Ink == "" {
		n.Ink = column.Ink
	}
	return n, nil
}

func (r *renderer) sceneTableScoreCell(p *scenePlan, id string, c sceneTableColumn, raw json.RawMessage, st Style, b Rect, surface string, ctx SceneContext, path string) (pptx.TableCell, TextRecord, error) {
	n, e := sceneTableScoreValue(raw, c, c.Type)
	if e != nil {
		return pptx.TableCell{}, TextRecord{}, e
	}
	count := int(*n.Max)
	spacing := 11.0
	if c.Type == "dots" {
		spacing = 12
	}
	width := float64(count)*8 + float64(count-1)*(spacing-8)
	if c.Type == "harvey" {
		width = 16
	}
	available := b.W - 12
	if c.Type == "harvey" {
		available = b.W
	}
	minHeight := 8.0
	if c.Type == "harvey" {
		minHeight = 16
	}
	if width > available+.02 || b.H < minHeight {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_score_geometry")
	}
	// Retain the existing scalar rating/Harvey geometry and colors exactly.
	if c.Type != "dots" && n.Ink == "" && c.Max == nil && (len(raw) == 0 || raw[0] != '{' || n.Max != nil && *n.Max == 4) {
		if e := r.sceneTableNumericMark(p, id, c.Type, *n.Value, b, surface); e != nil {
			return pptx.TableCell{}, TextRecord{}, e
		}
		return r.sceneNativeCell(id, "", st, b, surface, "primary", sceneTableAlign(c.Type), 12, 12, ctx)
	}
	color := ""
	if n.Ink != "" {
		color, e = r.sceneColor(ctx.Surface, n.Ink)
	} else if c.Type == "dots" {
		color, e = r.sceneColor(ctx.Surface, "strong")
	} else {
		color, e = r.sceneColor("light", "strong")
	}
	if e != nil {
		return pptx.TableCell{}, TextRecord{}, e
	}
	empty, e := r.sceneColor(surface, "bg")
	if e != nil {
		return pptx.TableCell{}, TextRecord{}, e
	}
	cell, tr, e := r.sceneNativeCell(id, "", st, b, surface, "primary", sceneTableAlign(c.Type), 12, 12, ctx)
	if e != nil {
		return cell, tr, e
	}
	markY := b.Y + (b.H-8)/2
	if c.Type == "dots" && len(n.Text) > 0 && string(n.Text) != "null" && string(n.Text) != `""` {
		tb := Rect{b.X, b.Y + 14, b.W, b.H - 14}
		if tb.H <= 0 {
			return cell, tr, fmt.Errorf("scene.table_dots_text_overflow")
		}
		var text string
		if e = json.Unmarshal(n.Text, &text); e == nil {
			small, _ := r.sceneStyle("small")
			cell, tr, e = r.sceneNativeCell(id, text, small, tb, surface, "primary", "left", 12, 12, ctx)
		} else {
			cell, tr, e = r.sceneTableBulletCell(id, n.Text, tb, surface, ctx, path+"/text")
		}
		if e != nil {
			return cell, tr, e
		}
		height := math.Max(tr.Layout.AllocationHeight, tr.Layout.OccupiedTop+tr.Layout.EstimatedOccupiedHeight)
		if height+14 > b.H+.02 {
			return cell, tr, fmt.Errorf("scene.table_dots_text_overflow")
		}
		markY = b.Y + (b.H-height-14)/2
		tr.Rect.Y = markY + 14
		tr.Rect.H = height
		// Native paragraph XML uses the record's text allocation; shrink tc margins
		// to the same centered block rather than placing the copy through dot glyphs.
		cell.Options.Margin = pptx.Margin{(tr.Rect.Y - b.Y) / 72, 12. / 72, (b.Y + b.H - tr.Rect.Y - height) / 72, 12. / 72}
	}
	switch c.Type {
	case "rating", "dots":
		for i := 0; i < count; i++ {
			fill := empty
			kind := pptx.ShapeTypeRect
			var outline *pptx.ShapeLineProps
			if c.Type == "dots" {
				kind = pptx.ShapeTypeEllipse
				fill = "CED7E6"
				if ctx.Surface == "inverse" || ctx.Surface == "deep" {
					fill = "50658E"
				}
			} else {
				outline = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: .75}
			}
			if float64(i) < *n.Value {
				fill = color
			}
			r.sceneDataShape(p, fmt.Sprintf("%s.%s-%d", id, c.Type, i+1), Rect{b.X + 12 + float64(i)*spacing, markY, 8, 8}, kind, fill, outline)
		}
	case "harvey":
		box := Rect{b.X + (b.W-16)/2, b.Y + (b.H-16)/2, 16, 16}
		if b.H < 16 {
			return cell, tr, fmt.Errorf("scene.table_harvey_geometry")
		}
		r.sceneDataShape(p, id+".harvey-outline", box, pptx.ShapeTypeEllipse, empty, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: .96})
		if *n.Value > 0 {
			kind := pptx.ShapeTypePie
			if *n.Value == 4 {
				kind = pptx.ShapeTypeEllipse
			}
			r.sceneDataShape(p, id+".harvey-fill", box, kind, color, nil)
			if *n.Value < 4 {
				p.Items[len(p.Items)-1].Shape.Props.AngleRange = &[2]float64{270, 270 + *n.Value*90}
			}
		}
	}
	return cell, tr, nil
}

func (r *renderer) sceneTableHeatCell(p *scenePlan, id string, c sceneTableColumn, raw json.RawMessage, st Style, b Rect, surface string, ctx SceneContext) (pptx.TableCell, TextRecord, error) {
	var n struct {
		Value *float64 `json:"value"`
		Text  *string  `json:"text,omitempty"`
		Scale string   `json:"scale,omitempty"`
	}
	var scalar float64
	if len(raw) > 0 && raw[0] != '{' {
		if e := json.Unmarshal(raw, &scalar); e != nil {
			// V12 staffing totals author numeric strings; the browser uses
			// numeric coercion for heat while preserving the displayed string.
			var value string
			if r.source.Revision != LibraryRevisionV12 || json.Unmarshal(raw, &value) != nil {
				return pptx.TableCell{}, TextRecord{}, e
			}
			var err error
			scalar, err = strconv.ParseFloat(value, 64)
			if err != nil || math.IsNaN(scalar) || math.IsInf(scalar, 0) {
				return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_heat_requires_finite_number")
			}
			n.Text = &value
		}
		n.Value = &scalar
	} else if e := sceneDecode(raw, &n); e != nil {
		return pptx.TableCell{}, TextRecord{}, e
	}
	if n.Value == nil {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_heat_value_required")
	}
	scale := n.Scale
	if scale == "" {
		scale = c.Scale
	}
	fill, ink, e := sceneHeatDomain(*n.Value, c.Min, c.Max, scale)
	if e != nil {
		return pptx.TableCell{}, TextRecord{}, e
	}
	if b.H <= 3 || b.W <= 3 {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_heat_geometry")
	}
	text := ""
	st.Weight = 600
	if n.Text != nil {
		text = *n.Text
	} else if c.ShowValue {
		text = strconv.FormatFloat(*n.Value, 'f', -1, 64)
		st.Family = "IBM Plex Mono"
	}
	cell, tr, e := r.sceneNativeCell(id, text, st, b, surface, "primary", "center", 12, 12, ctx)
	if e != nil {
		return cell, tr, e
	}
	tr.Color = ink
	cell.Options.Color = ink
	if tr.Rich != nil {
		for i := range tr.Rich.Paragraphs {
			for j := range tr.Rich.Paragraphs[i].Runs {
				tr.Rich.Paragraphs[i].Runs[j].Color = ink
			}
		}
	}
	// Paint the effective surface gutters and inset before the native table. An
	// explicit transparent tc fill retains editable native text above both shapes.
	background, e := r.sceneColor(surface, "bg")
	if e != nil {
		return cell, tr, e
	}
	r.sceneDataShape(p, id+".heat-gap", b, pptx.ShapeTypeRect, background, nil)
	r.sceneDataShape(p, id+".heat", Rect{b.X + 1.5, b.Y + 1.5, b.W - 3, b.H - 3}, pptx.ShapeTypeRect, fill, nil)
	cell.Options.Fill = &pptx.ShapeFillProps{Color: "FFFFFF", Transparency: 100}
	return cell, tr, nil
}

func (r *renderer) sceneTablePlainCell(id string, raw json.RawMessage, st Style, b Rect, surface string, ctx SceneContext) (pptx.TableCell, TextRecord, error) {
	var n struct {
		Text *string `json:"text"`
		Sub  string  `json:"sub,omitempty"`
	}
	if e := sceneDecode(raw, &n); e != nil {
		return pptx.TableCell{}, TextRecord{}, e
	}
	if n.Text == nil {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_plain_text_required")
	}
	if n.Sub == "" {
		return r.sceneNativeCell(id, *n.Text, st, b, surface, "primary", "left", 12, 12, ctx)
	}
	small, _ := r.sceneStyle("small")
	small.Weight = 400
	specs := []struct {
		text, key, role string
		style           Style
	}{{*n.Text, "text", "primary", st}, {n.Sub, "sub", "secondary", small}}
	cell, tr, e := r.sceneNativeCell(id, "", st, b, surface, "primary", "left", 12, 12, ctx)
	if e != nil {
		return cell, tr, e
	}
	rich := RichTextLayout{Contract: RichTextContract}
	layout := tr.Layout
	layout.Lines = nil
	offset, occupiedTop, lastOccupied := 0., 0., 0.
	var originals, displayed []string
	for i, part := range specs {
		_, pr, e := r.sceneNativeCell(id+"."+part.key, part.text, part.style, b, surface, part.role, "left", 12, 12, ctx)
		if e != nil {
			return cell, tr, e
		}
		paragraph := RichParagraphLayout{Key: part.key, Displayed: pr.Layout.Displayed, Runs: []RichRunLayout{{Key: "text", Original: pr.Layout.Original, Displayed: pr.Layout.Displayed, Style: pr.Layout.Style, Font: pr.Layout.Font, Color: pr.Color, Start: 0, End: len([]rune(pr.Layout.Displayed))}}}
		if pr.Rich != nil {
			if len(pr.Rich.Paragraphs) != 1 {
				return cell, tr, fmt.Errorf("scene.table_plain_rich_paragraph_count")
			}
			paragraph = pr.Rich.Paragraphs[0]
			paragraph.Key = part.key
		}
		paragraph.FirstLine = len(layout.Lines)
		paragraph.LineCount = len(pr.Layout.Lines)
		if i == 0 {
			paragraph.ParagraphGapAfter = 2
			occupiedTop = pr.Layout.OccupiedTop
		}
		for _, line := range pr.Layout.Lines {
			line.Baseline += offset
			layout.Lines = append(layout.Lines, line)
		}
		lastOccupied = offset + pr.Layout.OccupiedTop + pr.Layout.EstimatedOccupiedHeight
		offset += pr.Layout.AllocationHeight + paragraph.ParagraphGapAfter
		rich.Paragraphs = append(rich.Paragraphs, paragraph)
		originals = append(originals, pr.Layout.Original)
		displayed = append(displayed, pr.Layout.Displayed)
	}
	if math.Max(offset, lastOccupied) > b.H+.02 {
		return cell, tr, fmt.Errorf("scene.table_plain_sub_overflow: %s needs%.3fpt capacity%.3fpt", id, math.Max(offset, lastOccupied), b.H)
	}
	layout.Original = strings.Join(originals, "\n")
	layout.Displayed = strings.Join(displayed, "\n")
	layout.OccupiedTop = occupiedTop
	layout.AllocationHeight = offset
	layout.EstimatedOccupiedHeight = lastOccupied - occupiedTop
	layout.VerticalPolicy = "Source plain text/sub: two native paragraphs,2pt gap; native review pending"
	tr.Layout = layout
	tr.Rich = &rich
	cell.Text = layout.Displayed
	return cell, tr, nil
}
