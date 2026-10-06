package wmdesign

import (
	"fmt"
	"math"
	"strings"
)

// MetricsContract adds card metric content without changing the body-card contract.
const MetricsContract = "pptxgengo.wmds-metrics.v1"

// MetricValue contains authored display strings; it does not calculate or format data.
type MetricValue struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
type MetricSpec struct {
	Value     string       `json:"value"`
	Label     string       `json:"label"`
	Secondary *MetricValue `json:"secondary,omitempty"`
	Change    string       `json:"change,omitempty"`
	Status    string       `json:"status,omitempty"`
}
type KeyedMetricValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Label string `json:"label"`
}
type MetricGroupSpec struct {
	Primary   MetricValue        `json:"primary"`
	Secondary []KeyedMetricValue `json:"secondary"`
}

// planMetricContent returns the minimum absolute bottom. Fixed or rounded card
// heights move the comparison/status footer to the bottom; they never compress it.
func (r *renderer) planMetricContent(p *componentPlan, id string, card *CardSpec, b Rect, surface string) (float64, error) {
	add := func(part, text, style, role string, x, y, w float64, lines int) (TextRecord, error) {
		return r.planText(p, id+"."+part, text, style, surface, role, x, y, w, lines)
	}
	if card.MetricGroup != nil {
		if p.record.Rect.W < 414-.01 {
			return 0, fmt.Errorf("metric_group.minimum_width: 6 columns required")
		}
		mg := card.MetricGroup
		if len(mg.Secondary) < 1 || len(mg.Secondary) > 2 {
			return 0, fmt.Errorf("metric_group.secondary_count: one or two required")
		}
		keys := map[string]bool{}
		for _, sm := range mg.Secondary {
			if !validPartKey(sm.Key) || keys[sm.Key] {
				return 0, fmt.Errorf("metric_group.invalid_or_duplicate_key: %s", sm.Key)
			}
			keys[sm.Key] = true
		}
		// Frozen renderer: 1.3fr primary, 1fr per secondary, 18pt gutters.
		unit := (b.W - 18*float64(len(mg.Secondary))) / (1.3 + float64(len(mg.Secondary)))
		pw := 1.3 * unit
		y := b.Y + 6
		value, e := add("metric.primary.value", mg.Primary.Value, "stat", "display", b.X, y, pw, 1)
		if e != nil {
			return 0, e
		}
		label, e := add("metric.primary.label", mg.Primary.Label, "body", "secondary", b.X, y+value.Rect.H+6, pw, 3)
		if e != nil {
			return 0, e
		}
		bottom := label.Rect.Y + label.Rect.H
		x := b.X + pw + 18
		for _, sm := range mg.Secondary {
			part := "metric.secondary." + sm.Key
			// .75pt divider and 12pt padding belong to each secondary track.
			sx, sw := x+.75+12, unit-.75-12
			val, e := add(part+".value", sm.Value, "number", "display", sx, y+6, sw, 1)
			if e != nil {
				return 0, e
			}
			lab, e := add(part+".label", sm.Label, "small", "secondary", sx, val.Rect.Y+val.Rect.H+6, sw, 3)
			if e != nil {
				return 0, e
			}
			end := lab.Rect.Y + lab.Rect.H
			if e = r.planShape(p, id+"."+part+".divider", Rect{x, y, .75, end - y}, surface, "line"); e != nil {
				return 0, e
			}
			bottom = math.Max(bottom, end)
			x += unit + 18
		}
		return bottom, nil
	}
	m := card.Metric
	style := "stat"
	if p.record.Rect.W < 270 {
		style = "stat-sm"
	}
	value, e := add("metric.value", m.Value, style, "display", b.X, b.Y, b.W, 1)
	if e != nil {
		return 0, e
	}
	label, e := add("metric.label", m.Label, "body", "secondary", b.X, b.Y+value.Rect.H+6, b.W, 3)
	if e != nil {
		return 0, e
	}
	bottom := label.Rect.Y + label.Rect.H
	if m.Secondary == nil && m.Change == "" && m.Status == "" {
		return bottom, nil
	}
	y := bottom + 12 // source flex filler has a 6pt gap on each side, even at zero height.
	if m.Secondary != nil {
		if strings.TrimSpace(m.Secondary.Label) == "" {
			return 0, fmt.Errorf("component.empty_content: %s.metric.footer.secondary.label", id)
		}
		// Number and label share their first baseline. The value gets its measured
		// one-line width, leaving the rest of the content width to the label.
		st, _ := r.bodyStyle("number")
		vl, e := r.measureText(m.Secondary.Value, st, b.W)
		if e != nil {
			return 0, e
		}
		if len(vl.Lines) != 1 {
			return 0, fmt.Errorf("metric.secondary_value_must_be_one_line")
		}
		nw := vl.Lines[0].Advance + .01
		if b.W-nw-9 <= 0 {
			return 0, fmt.Errorf("metric.secondary_label_has_no_width")
		}
		if e = r.planShape(p, id+".metric.footer.divider", Rect{b.X, y, b.W, .75}, surface, "line"); e != nil {
			return 0, e
		}
		y += .75 + 9
		val, e := add("metric.footer.secondary.value", m.Secondary.Value, "number", "display", b.X, y, nw, 1)
		if e != nil {
			return 0, e
		}
		ls, _ := r.bodyStyle("small")
		ll, e := r.measureText(m.Secondary.Label, ls, b.W-nw-9)
		if e != nil {
			return 0, e
		}
		shift := val.Layout.Lines[0].Baseline - ll.Lines[0].Baseline
		lab, e := add("metric.footer.secondary.label", m.Secondary.Label, "small", "secondary", b.X+nw+9, y+shift, b.W-nw-9, 3)
		if e != nil {
			return 0, e
		}
		y = math.Max(val.Rect.Y+val.Rect.H, lab.Rect.Y+lab.Rect.H) + 6
	}
	if m.Change != "" {
		tr, e := add("metric.footer.change", m.Change, "label", "emphasis", b.X, y, b.W, 1)
		if e != nil {
			return 0, e
		}
		y = tr.Rect.Y + tr.Rect.H + 6
	}
	if m.Status != "" {
		names := map[string]string{"on": "On track", "risk": "At risk", "off": "Off track"}
		name, ok := names[m.Status]
		if !ok {
			return 0, fmt.Errorf("metric.unsupported_status: %s", m.Status)
		}
		color, ok := r.source.Tokens.Colors.Dataviz.KPI[m.Status]
		if !ok {
			return 0, fmt.Errorf("source.missing_kpi_color: %s", m.Status)
		}
		color = strings.TrimPrefix(color, "#")
		// Tokens require a 1pt Grounded outline on data marks. A native outer square
		// gives the mark an accessible boundary without changing its semantic hue.
		grounded := r.ink("light", "display")
		bg := r.ink(surface, "bg")
		if math.Max(contrast(grounded, bg), contrast(color, bg)) < 3 {
			return 0, fmt.Errorf("metric.insufficient_status_mark_contrast")
		}
		mark := Rect{b.X, y + 2, 8, 8}
		// The one-point border is contained in the 8pt mark, not an outside stroke.
		p.record.Shapes = append(p.record.Shapes, ShapeRecord{ID: id + ".metric.footer.status.outline", Rect: mark, Color: grounded}, ShapeRecord{ID: id + ".metric.footer.status.mark", Rect: Rect{mark.X + 1, mark.Y + 1, 6, 6}, Color: color})
		p.record.Parts = append(p.record.Parts, id+".metric.footer.status.outline", id+".metric.footer.status.mark")
		tr, e := add("metric.footer.status.label", name, "label", "primary", b.X+14, y, b.W-14, 1)
		if e != nil {
			return 0, e
		}
		y = tr.Rect.Y + tr.Rect.H + 6
	}
	return y - 6, nil
}

func shiftMetricFooter(p *componentPlan, id string, dy float64) {
	prefix := id + ".metric.footer."
	for i := range p.texts {
		if strings.HasPrefix(p.texts[i].ID, prefix) {
			p.texts[i].Rect.Y += dy
		}
	}
	for i := range p.record.Shapes {
		if strings.HasPrefix(p.record.Shapes[i].ID, prefix) {
			p.record.Shapes[i].Rect.Y += dy
		}
	}
}
