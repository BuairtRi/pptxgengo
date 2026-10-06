package wmdesign

import (
	"fmt"
	"math"
	"strings"
)

// ComponentsContract is additive to the foundation document, and requires v2.
const ComponentsContract = "pptxgengo.wmds-components.v1"

// TextBlockSpec supports the three text.block variants. Label requires title.
type TextBlockSpec struct {
	Label string `json:"label,omitempty"`
	Title string `json:"title,omitempty"`
	Body  string `json:"body"`
}

// BodyBlock is exactly one paragraph or a keyed list of uniform square bullets.
// Keys survive reordering and must be unique across the component body.
type BodyBlock struct {
	Key       string       `json:"key"`
	Paragraph string       `json:"p,omitempty"`
	Bullets   []BulletItem `json:"bullets,omitempty"`
}
type BulletItem struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}
type CardBand struct {
	Surface string `json:"surface"`
}
type CardEdge struct {
	Side   string  `json:"side"`
	Weight float64 `json:"weight_pt"`
	Ink    string  `json:"ink"`
}

// CardSpec deliberately exposes only implemented card anatomy. Unknown JSON
// fields fail at the CLI boundary. Body and metric content have separate contracts.
type CardSpec struct {
	Label        string           `json:"label,omitempty"`
	Title        string           `json:"title,omitempty"`
	Body         []BodyBlock      `json:"body,omitempty"`
	Metric       *MetricSpec      `json:"metric,omitempty"`
	MetricGroup  *MetricGroupSpec `json:"metric_group,omitempty"`
	Padding      float64          `json:"padding_pt,omitempty"`
	TitleStyle   string           `json:"title_style,omitempty"`
	TitleInk     string           `json:"title_ink,omitempty"`
	InlineNumber string           `json:"inline_number,omitempty"`
	NumberInk    string           `json:"number_ink,omitempty"`
	BodySize     string           `json:"body_size,omitempty"`
	Dense        bool             `json:"dense,omitempty"`
	Band         *CardBand        `json:"band,omitempty"`
	Edge         *CardEdge        `json:"edge,omitempty"`
}
type ShapeRecord struct {
	ID       string  `json:"id"`
	Rect     Rect    `json:"rect"`
	Color    string  `json:"color"`
	Geometry string  `json:"geometry,omitempty"`
	Rotation float64 `json:"rotation_deg,omitempty"`
}
type ComponentRecord struct {
	ID             string         `json:"id"`
	Definition     string         `json:"definition"`
	Contract       string         `json:"contract"`
	Rect           Rect           `json:"rect"`
	RequiredHeight float64        `json:"required_height_pt"`
	Padding        float64        `json:"padding_pt,omitempty"`
	Density        string         `json:"density"`
	Parts          []string       `json:"parts"`
	Shapes         []ShapeRecord  `json:"shapes,omitempty"`
	Limits         map[string]int `json:"line_limits"`
}
type componentPlan struct {
	record ComponentRecord
	texts  []TextRecord
}

func validPartKey(k string) bool {
	if k == "" || k == "." || k == ".." {
		return false
	}
	for _, c := range k {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}
func (r *renderer) planText(p *componentPlan, id, text, style, surface, role string, x, y, w float64, maxLines int) (TextRecord, error) {
	var out TextRecord
	if strings.TrimSpace(text) == "" {
		return out, fmt.Errorf("component.empty_content: %s", id)
	}
	if strings.Contains(text, "[[") || strings.Contains(text, "]]") || strings.Contains(text, "[^") {
		return out, fmt.Errorf("text.unsupported_emphasis_or_footnote: %s", id)
	}
	st, e := r.bodyStyle(style)
	if e != nil {
		return out, e
	}
	fg, e := r.source.Ink(surface, role)
	if e != nil {
		return out, e
	}
	bg, e := r.source.Ink(surface, "bg")
	if e != nil {
		return out, e
	}
	min := 4.5
	if st.Size >= 18 || st.Size >= 14 && st.Weight >= 700 {
		min = 3
	}
	if !r.contrastAllows(id, st, fg, bg, min) {
		return out, fmt.Errorf("text.insufficient_contrast: %s", id)
	}
	l, e := r.measureText(text, st, w)
	if e != nil {
		return out, fmt.Errorf("%s: %w", id, e)
	}
	// Native reference advances differ by up to 0.132pt in the calibrated
	// matrix. Reject decisions within 0.25pt of a wrap boundary rather than
	// advertise a stable component line allocation at an unqualified boundary.
	if style != "number" {
		for i, line := range l.Lines {
			if math.Abs(w-line.Advance) <= .25 {
				return out, fmt.Errorf("component.uncertain_wrap_boundary: %s line %d is within 0.25pt of %.3fpt", id, i+1, w)
			}
			if i+1 < len(l.Lines) && strings.Contains(l.Displayed, line.Text+l.Lines[i+1].Text) {
				next := strings.Fields(l.Lines[i+1].Text)
				if len(next) > 0 {
					candidate, e := r.measureText(strings.TrimRight(line.Text, " ")+" "+next[0], st, 8191)
					if e == nil && len(candidate.Lines) == 1 && math.Abs(w-candidate.Lines[0].Advance) <= .25 {
						return out, fmt.Errorf("component.uncertain_wrap_boundary: %s next word is within 0.25pt of %.3fpt", id, w)
					}
				}
			}
		}
	}
	if maxLines > 0 && len(l.Lines) > maxLines {
		return out, fmt.Errorf("component.line_limit: %s has %d lines, limit %d", id, len(l.Lines), maxLines)
	}
	out = TextRecord{ID: id, Rect: Rect{x, y, w, math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)}, Color: fg, Align: "left", Layout: l}
	p.texts = append(p.texts, out)
	p.record.Parts = append(p.record.Parts, id)
	return out, nil
}
func (r *renderer) planShape(p *componentPlan, id string, b Rect, surface, role string) error {
	color, e := r.source.Ink(surface, role)
	if e != nil {
		return e
	}
	p.record.Shapes = append(p.record.Shapes, ShapeRecord{ID: id, Rect: b, Color: color})
	p.record.Parts = append(p.record.Parts, id)
	return nil
}
func (r *renderer) planComponent(n Node, b, zone Rect, surface string) (componentPlan, error) {
	p := componentPlan{record: ComponentRecord{ID: n.ID, Contract: ComponentsContract, Rect: b, Density: "standard", Limits: map[string]int{"label": 1, "title": 2, "number": 1}}}
	if r.typeEngine.engine != CandidateEngine {
		return p, fmt.Errorf("component.requires_v2_engine")
	}
	if n.Text != "" || n.Style != "" || n.Ink != "" || n.Align != "" {
		return p, fmt.Errorf("component.conflicting_primitive_fields: %s", n.ID)
	}
	// Context exceptions for narrow five-up and rails are not qualified here.
	if n.Grid == "five-up" || b.W < 198-.01 {
		return p, fmt.Errorf("component.minimum_width: %s needs at least 198pt", n.ID)
	}
	if n.Kind == "textblock" && b.W > 558+.01 {
		return p, fmt.Errorf("textblock.maximum_width: %s exceeds 8 columns", n.ID)
	}
	y := b.Y
	bottom := y
	add := func(part, text, style, surf, role string, x, w float64, limit int) error {
		tr, e := r.planText(&p, n.ID+"."+part, text, style, surf, role, x, y, w, limit)
		if e != nil {
			return e
		}
		bottom = math.Max(bottom, tr.Rect.Y+tr.Rect.H)
		// Flow follows authored line allocation; occupied union independently checks
		// terminal ink. Internal 6/12pt gaps keep adjacent boxes apart.
		y += tr.Layout.AllocationHeight
		return nil
	}
	if n.Kind == "textblock" {
		if n.TextBlock == nil || n.Card != nil {
			return p, fmt.Errorf("component.invalid_payload: %s", n.ID)
		}
		v := n.TextBlock
		p.record.Definition = "text.block"
		if v.Label != "" && v.Title == "" {
			return p, fmt.Errorf("textblock.label_requires_title")
		}
		if v.Label != "" {
			if e := add("label", v.Label, "label", surface, "emphasis", b.X, b.W, 1); e != nil {
				return p, e
			}
			y += 6
		}
		if v.Title != "" {
			if e := add("title", v.Title, "subhead", surface, "display", b.X, b.W, 2); e != nil {
				return p, e
			}
			y += 6
		}
		if e := add("body", v.Body, "body", surface, "secondary", b.X, b.W, 0); e != nil {
			return p, e
		}
	} else {
		if n.Card == nil || n.TextBlock != nil {
			return p, fmt.Errorf("component.invalid_payload: %s", n.ID)
		}
		v := n.Card
		p.record.Definition = "card.presets"
		if v.Band != nil {
			p.record.Definition = "card.band"
			if v.Title == "" {
				return p, fmt.Errorf("card.band_requires_title")
			}
		}
		if v.InlineNumber != "" && v.Title == "" {
			return p, fmt.Errorf("card.number_requires_title")
		}
		if len(v.Body) > 1 && b.W < 414-.01 {
			return p, fmt.Errorf("card.paragraphs_minimum_width: 6 columns required")
		}
		if v.Title == "" && (v.TitleStyle != "" || v.TitleInk != "") {
			return p, fmt.Errorf("card.title_options_without_title")
		}
		kinds := 0
		if len(v.Body) > 0 {
			kinds++
		}
		if v.Metric != nil {
			kinds++
		}
		if v.MetricGroup != nil {
			kinds++
		}
		if kinds == 0 {
			return p, fmt.Errorf("card.body_required")
		}
		if kinds != 1 {
			return p, fmt.Errorf("card.exactly_one_content_kind_required")
		}
		if v.Metric != nil || v.MetricGroup != nil {
			if v.BodySize != "" || v.Dense {
				return p, fmt.Errorf("metric.body_density_options_unsupported")
			}
			p.record.Contract = MetricsContract
			p.record.Limits["metric_value"] = 1
			p.record.Limits["metric_label"] = 3
			p.record.Limits["secondary_value"] = 1
			p.record.Limits["secondary_label"] = 3
			p.record.Limits["change"] = 1
			p.record.Limits["status"] = 1
		}
		pad := v.Padding
		if pad == 0 {
			pad = 18
			if b.W < 270 {
				pad = 12
			}
		}
		if pad != 12 && pad != 18 && pad != 24 {
			return p, fmt.Errorf("card.unsupported_padding")
		}
		p.record.Padding = pad
		titleStyle := v.TitleStyle
		if titleStyle == "" {
			titleStyle = "subhead"
		}
		if titleStyle != "subhead" && titleStyle != "heading" {
			return p, fmt.Errorf("card.unsupported_title_style")
		}
		ti := v.TitleInk
		if ti == "" {
			ti = "display"
		}
		ni := v.NumberInk
		if ni == "" {
			ni = "emphasis"
		}
		if v.InlineNumber == "" && v.NumberInk != "" {
			return p, fmt.Errorf("card.number_ink_without_number")
		}
		bs := "body"
		if v.BodySize != "" && v.BodySize != "body" && v.BodySize != "small" {
			return p, fmt.Errorf("card.unsupported_body_size")
		}
		if v.BodySize == "small" {
			if !v.Dense {
				return p, fmt.Errorf("card.small_body_requires_dense_context")
			}
			bs = "small"
			p.record.Density = "dense"
		} else if v.Dense {
			return p, fmt.Errorf("card.dense_requires_small_body")
		}
		left, right, top, bot := pad, pad, pad, pad
		if v.Edge != nil {
			ed := v.Edge
			if ed.Weight != 3 && ed.Weight != 6 {
				return p, fmt.Errorf("card.unsupported_edge_weight")
			}
			fg, e := r.source.Ink(surface, ed.Ink)
			if e != nil {
				return p, e
			}
			bg, e := r.source.Ink(surface, "bg")
			if e != nil {
				return p, e
			}
			if contrast(fg, bg) < 3 {
				return p, fmt.Errorf("card.insufficient_edge_contrast")
			}
			switch ed.Side {
			case "left":
				left += ed.Weight
			case "right":
				right += ed.Weight
			case "top":
				top += ed.Weight
			case "bottom":
				bot += ed.Weight
			default:
				return p, fmt.Errorf("card.unsupported_edge_side")
			}
		}
		x, w := b.X+left, b.W-left-right
		y = b.Y + top
		bottom = y
		hs := surface
		if v.Band != nil {
			hs = v.Band.Surface
			if _, e := r.source.Ink(hs, "bg"); e != nil {
				return p, e
			}
		}
		if v.Label != "" {
			if e := add("label", v.Label, "label", hs, "emphasis", x, w, 1); e != nil {
				return p, e
			}
			y += 6
		}
		if v.Title != "" {
			tx, tw := x, w
			startY := y
			if v.InlineNumber != "" {
				st, _ := r.bodyStyle("number")
				l, e := r.measureText(v.InlineNumber, st, w)
				if e != nil {
					return p, e
				}
				if len(l.Lines) != 1 {
					return p, fmt.Errorf("card.number_must_be_one_line")
				}
				nw := l.Lines[0].Advance + .01
				// Baseline alignment between Mono number and either title token.
				ts, _ := r.bodyStyle(titleStyle)
				tl, e := r.measureText(v.Title, ts, w-nw-9)
				if e != nil {
					return p, e
				}
				shift := tl.Lines[0].Baseline - l.Lines[0].Baseline
				y = startY + math.Max(0, shift)
				if e := add("number", v.InlineNumber, "number", hs, ni, x, nw, 1); e != nil {
					return p, e
				}
				y = startY + math.Max(0, -shift)
				tx = x + nw + 9
				tw = w - nw - 9
			}
			if e := add("title", v.Title, titleStyle, hs, ti, tx, tw, 2); e != nil {
				return p, e
			}
			y = math.Max(y, bottom)
			if v.Band == nil {
				y += 12
			}
		}
		var bandBottom float64
		if v.Band != nil {
			bandBottom = math.Max(y, bottom) + pad
			if e := r.planShape(&p, n.ID+".band", Rect{b.X, b.Y, b.W, bandBottom - b.Y}, hs, "bg"); e != nil {
				return p, e
			}
			if hs == "outline" {
				outlineParts(&p, n.ID+".band", Rect{b.X, b.Y, b.W, bandBottom - b.Y}, r.ink(hs, "line"), 1)
			}
			y = bandBottom + pad
		}
		if v.Metric == nil && v.MetricGroup == nil {
			keys := map[string]bool{}
			for _, bl := range v.Body {
				if !validPartKey(bl.Key) || keys[bl.Key] {
					return p, fmt.Errorf("card.invalid_or_duplicate_body_key: %s", bl.Key)
				}
				keys[bl.Key] = true
				if (bl.Paragraph != "") == (len(bl.Bullets) > 0) {
					return p, fmt.Errorf("card.body_block_requires_exactly_one_kind: %s", bl.Key)
				}
				if bl.Paragraph != "" {
					if e := add("body."+bl.Key, bl.Paragraph, bs, surface, "secondary", x, w, 0); e != nil {
						return p, e
					}
					y += 6
				} else {
					// Reference CSS: 4pt/15pt square/inset and 6pt gaps; dense 3pt/12pt and 3pt gaps.
					for _, it := range bl.Bullets {
						if !validPartKey(it.Key) || keys[it.Key] {
							return p, fmt.Errorf("card.invalid_or_duplicate_body_key: %s", it.Key)
						}
						keys[it.Key] = true
						id := "body." + bl.Key + "." + it.Key
						inset, mark, offset, gap := 15.0, 4.0, 8.5, 6.0
						if bs == "small" {
							inset, mark, offset, gap = 12, 3, 7, 3
						}
						if e := r.planShape(&p, n.ID+"."+id+".marker", Rect{x, y + offset, mark, mark}, surface, "secondary"); e != nil {
							return p, e
						}
						if e := add(id, it.Text, bs, surface, "secondary", x+inset, w-inset, 0); e != nil {
							return p, e
						}
						y += gap
					}
					if bs == "small" {
						y += 3
					}
				}
			}
			bottom = math.Max(bottom, y-6) + bot
			p.record.RequiredHeight = bottom - b.Y
		} else {
			// Common measured headers precede the metric body. Header occupied
			// bounds remain a separate minimum, including across a title band.
			y = math.Max(y, bottom)
			required, e := r.planMetricContent(&p, n.ID, v, Rect{x, y, w, 0}, surface)
			if e != nil {
				return p, e
			}
			p.record.RequiredHeight = required + bot - b.Y
		}
		if b.H == 0 {
			b.H = math.Ceil(p.record.RequiredHeight/18) * 18
		}
		if v.Metric != nil && p.record.RequiredHeight <= b.H+.02 {
			shiftMetricFooter(&p, n.ID, b.H-p.record.RequiredHeight)
		}
		// Container precedes band, edges and all text in native z-order.
		bg, e := r.source.Ink(surface, "bg")
		if e != nil {
			return p, e
		}
		p.record.Shapes = append([]ShapeRecord{{ID: n.ID + ".container", Rect: b, Color: bg}}, p.record.Shapes...)
		p.record.Parts = append([]string{n.ID + ".container"}, p.record.Parts...)
		if surface == "outline" {
			outlineParts(&p, n.ID+".outline", b, r.ink(surface, "line"), 1)
		}
		if v.Edge != nil {
			ed := v.Edge
			eb := b
			switch ed.Side {
			case "left":
				eb.W = ed.Weight
			case "right":
				eb.X += b.W - ed.Weight
				eb.W = ed.Weight
			case "top":
				eb.H = ed.Weight
			case "bottom":
				eb.Y += b.H - ed.Weight
				eb.H = ed.Weight
			}
			if e := r.planShape(&p, n.ID+".edge", eb, surface, ed.Ink); e != nil {
				return p, e
			}
			// The part crossing both band and body must contrast against both.
			if v.Band != nil && ed.Side != "bottom" {
				fg, _ := r.source.Ink(surface, ed.Ink)
				bbg, _ := r.source.Ink(hs, "bg")
				if contrast(fg, bbg) < 3 {
					return p, fmt.Errorf("card.insufficient_band_edge_contrast")
				}
			}
		}
	}
	if p.record.RequiredHeight == 0 {
		p.record.RequiredHeight = math.Max(y, bottom) - b.Y
	}
	if b.H == 0 {
		b.H = math.Ceil(p.record.RequiredHeight/18) * 18
	}
	if p.record.RequiredHeight > b.H+.02 {
		return p, fmt.Errorf("component.vertical_overflow: %s needs %.3fpt, capacity %.3fpt", n.ID, p.record.RequiredHeight, b.H)
	}
	if e := r.source.Tokens.Grid.OuterBox(b); e != nil {
		return p, e
	}
	if !inside(b, zone) {
		return p, fmt.Errorf("node.outside_zone: %s", n.ID)
	}
	for _, tr := range p.texts {
		if !inside(tr.Rect, b) {
			return p, fmt.Errorf("component.part_outside_bounds: %s", tr.ID)
		}
	}
	p.record.Rect = b
	return p, r.err
}
func outlineParts(p *componentPlan, id string, b Rect, color string, weight float64) {
	for i, r := range []Rect{{b.X, b.Y, b.W, weight}, {b.X, b.Y + b.H - weight, b.W, weight}, {b.X, b.Y, weight, b.H}, {b.X + b.W - weight, b.Y, weight, b.H}} {
		part := fmt.Sprintf("%s.%d", id, i)
		p.record.Shapes = append(p.record.Shapes, ShapeRecord{ID: part, Rect: r, Color: color})
		p.record.Parts = append(p.record.Parts, part)
	}
}
func (r *renderer) drawComponent(p componentPlan) {
	for _, sh := range p.record.Shapes {
		if !r.drawMetricShape(sh) {
			r.shape(sh.ID, sh.Rect, sh.Color, 0)
		}
	}
	for _, tr := range p.texts {
		r.text(tr.ID, tr.Layout.Original, tr.Layout.Style, tr.Rect, tr.Color, tr.Align, 0)
	}
}
