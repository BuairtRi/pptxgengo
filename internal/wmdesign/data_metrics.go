package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

const DataMetricsContract = "pptxgengo.wmds-data-metrics.v1"

// NumberFormatSpec mirrors the frozen format object. Value is a JSON numeric
// scalar, or an array of exactly two numeric endpoints when Kind is range.
type NumberFormatSpec struct {
	Value    json.RawMessage `json:"value"`
	Kind     string          `json:"kind"`
	Currency string          `json:"currency,omitempty"`
	Scale    string          `json:"scale,omitempty"`
	Unit     string          `json:"unit,omitempty"`
	Of       string          `json:"of,omitempty"`
}

type DataMetricChange struct {
	Direction string `json:"dir"`
	Text      string `json:"text"`
}

type DataMetricSpec struct {
	Value     string             `json:"value,omitempty"`
	Format    *NumberFormatSpec  `json:"format,omitempty"`
	Label     string             `json:"label"`
	Change    *DataMetricChange  `json:"change,omitempty"`
	Secondary []KeyedMetricValue `json:"secondary,omitempty"`
	Target    string             `json:"target,omitempty"`
	Status    string             `json:"status,omitempty"`
	Source    string             `json:"source,omitempty"`
}

// FormatNumber uses decimal arithmetic, independent of host locale and binary
// floating point rounding. Authored strings bypass this formatter altogether.
func FormatNumber(sp NumberFormatSpec) (string, error) {
	if sp.Kind != "currency" && sp.Kind != "percent" && sp.Kind != "number" && sp.Kind != "delta" && sp.Kind != "range" {
		return "", fmt.Errorf("number.unsupported_kind: %s", sp.Kind)
	}
	cur := sp.Currency
	if cur == "" {
		cur = "USD"
	}
	symbol, ok := map[string]string{"USD": "$", "EUR": "€", "GBP": "£"}[cur]
	if !ok {
		return "", fmt.Errorf("number.unsupported_currency: %s", cur)
	}
	if sp.Scale != "" && sp.Scale != "K" && sp.Scale != "M" && sp.Scale != "B" {
		return "", fmt.Errorf("number.unsupported_scale: %s", sp.Scale)
	}
	if sp.Kind != "range" && sp.Of != "" {
		return "", fmt.Errorf("number.of_requires_range")
	}
	if sp.Kind != "currency" && !(sp.Kind == "range" && sp.Of == "currency") && (sp.Currency != "" || sp.Scale != "") {
		return "", fmt.Errorf("number.currency_options_without_currency")
	}
	if sp.Unit != "" {
		if sp.Kind != "number" && sp.Kind != "delta" {
			return "", fmt.Errorf("number.unit_requires_number_or_delta")
		}
		if len(sp.Unit) > 24 || strings.TrimSpace(sp.Unit) != sp.Unit {
			return "", fmt.Errorf("number.invalid_unit")
		}
		for _, c := range sp.Unit {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == ' ') {
				return "", fmt.Errorf("number.invalid_unit")
			}
		}
	}
	if sp.Kind == "range" {
		if sp.Of != "percent" && sp.Of != "currency" {
			return "", fmt.Errorf("number.unsupported_range_kind: %s", sp.Of)
		}
		var raw []json.RawMessage
		if err := json.Unmarshal(sp.Value, &raw); err != nil || len(raw) != 2 {
			return "", fmt.Errorf("number.range_requires_two_endpoints")
		}
		a, err := decimalValue(raw[0])
		if err != nil {
			return "", err
		}
		b, err := decimalValue(raw[1])
		if err != nil {
			return "", err
		}
		if a.Cmp(b) > 0 {
			return "", fmt.Errorf("number.range_reversed")
		}
		if sp.Of == "percent" {
			return percentDigits(a) + "–" + percentDigits(b) + "%", nil
		}
		if a.Sign() < 0 {
			return "", fmt.Errorf("number.negative_currency_range_unsupported")
		}
		scale := sp.Scale
		if scale == "" {
			scale = currencyScale(b)
		}
		return symbol + currencyDigits(a, scale) + "–" + currencyDigits(b, scale) + scale, nil
	}
	v, err := decimalValue(sp.Value)
	if err != nil {
		return "", err
	}
	unit := ""
	if sp.Unit != "" {
		unit = " " + sp.Unit
	}
	switch sp.Kind {
	case "currency":
		scale := sp.Scale
		if scale == "" {
			scale = currencyScale(v)
		}
		return sign(v) + symbol + currencyDigits(absRat(v), scale) + scale, nil
	case "percent":
		return percentDigits(v) + "%", nil
	case "number":
		return sign(v) + groupDigits(roundDecimal(absRat(v), 2)) + unit, nil
	case "delta":
		s := "+"
		if v.Sign() < 0 {
			s = "−"
		}
		digits := groupDigits(roundDecimal(absRat(v), 3))
		if sp.Unit == "pts" {
			digits = percentDigits(absRat(v))
		}
		return s + digits + unit, nil
	}
	return "", fmt.Errorf("number.unsupported_kind")
}

func decimalValue(raw json.RawMessage) (*big.Rat, error) {
	// json.Number decoding rejects strings, booleans, arrays, null and non-finite values.
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var value any
	if err := d.Decode(&value); err != nil {
		return nil, fmt.Errorf("number.invalid_value")
	}
	n, ok := value.(json.Number)
	if !ok {
		return nil, fmt.Errorf("number.numeric_value_required")
	}
	// A modest precision/exponent envelope bounds cost and avoids claiming the
	// exploratory JavaScript's IEEE-754 precision beyond safe integers.
	if len(n.String()) > 64 {
		return nil, fmt.Errorf("number.value_precision_exceeded")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return nil, fmt.Errorf("number.invalid_value")
	}
	s := n.String()
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		var exp int
		if _, err := fmt.Sscan(s[i+1:], &exp); err != nil || exp < -12 || exp > 15 {
			return nil, fmt.Errorf("number.exponent_outside_envelope")
		}
	}
	v, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, fmt.Errorf("number.invalid_decimal")
	}
	limit := new(big.Rat).SetInt64(9007199254740991)
	if absRat(v).Cmp(limit) > 0 {
		return nil, fmt.Errorf("number.value_outside_envelope")
	}
	return v, nil
}
func absRat(v *big.Rat) *big.Rat { return new(big.Rat).Abs(v) }
func sign(v *big.Rat) string {
	if v.Sign() < 0 {
		return "−"
	}
	return ""
}
func roundDecimal(v *big.Rat, places int) string {
	m := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	n := new(big.Int).Mul(v.Num(), m)
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(n, v.Denom(), rem)
	if new(big.Int).Lsh(rem, 1).Cmp(v.Denom()) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	s := q.String()
	if places == 0 {
		return s
	}
	if len(s) <= places {
		s = strings.Repeat("0", places-len(s)+1) + s
	}
	s = s[:len(s)-places] + "." + s[len(s)-places:]
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}
func groupDigits(s string) string {
	parts := strings.SplitN(s, ".", 2)
	whole := parts[0]
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	if len(parts) > 1 {
		whole += "." + parts[1]
	}
	return whole
}
func percentDigits(v *big.Rat) string {
	x := new(big.Rat).Mul(absRat(v), big.NewRat(100, 1))
	places := 0
	if x.Cmp(big.NewRat(10, 1)) < 0 {
		places = 1
	}
	return sign(v) + roundDecimal(x, places)
}
func currencyScale(v *big.Rat) string {
	a := absRat(v)
	for _, p := range []struct {
		n int64
		s string
	}{{1000000000, "B"}, {1000000, "M"}, {1000, "K"}} {
		if a.Cmp(big.NewRat(p.n, 1)) >= 0 {
			return p.s
		}
	}
	return ""
}
func currencyDigits(v *big.Rat, scale string) string {
	div := map[string]int64{"": 1, "K": 1000, "M": 1000000, "B": 1000000000}[scale]
	x := new(big.Rat).Quo(absRat(v), big.NewRat(div, 1))
	places := 0
	if (scale == "M" || scale == "B") && x.Cmp(big.NewRat(10, 1)) < 0 {
		places = 1
	}
	return roundDecimal(x, places)
}

// drawMetricShape handles the native triangle preset owned by standalone metrics.
func (r *renderer) drawMetricShape(sh ShapeRecord) bool {
	if sh.Geometry != "triangle" {
		return false
	}
	if r.err != nil {
		return true
	}
	props := &pptx.ShapeProps{PositionProps: pos(sh.Rect), ObjectNameProps: pptx.ObjectNameProps{ObjectName: sh.ID}, Fill: &pptx.ShapeFillProps{Color: sh.Color}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none", Color: sh.Color}}}
	props.Rotate = sh.Rotation
	r.err = r.slide.AddShape(pptx.ShapeTypeTriangle, props)
	return true
}

func (r *renderer) planDataMetric(n Node, b, zone Rect, surface string) (componentPlan, error) {
	p := componentPlan{record: ComponentRecord{ID: n.ID, Definition: "data.metric", Contract: DataMetricsContract, Rect: b, Density: "standard", Limits: map[string]int{"value": 1, "label": 3, "secondary.value": 1, "secondary.label": 3, "change": 1, "target": 1, "source": 2}}}
	if r.typeEngine.engine != CandidateEngine {
		return p, fmt.Errorf("component.requires_v2_engine")
	}
	if n.DataMetric == nil || n.Card != nil || n.TextBlock != nil || n.Text != "" || n.Style != "" || n.Ink != "" || n.Align != "" {
		return p, fmt.Errorf("component.invalid_metric_payload: %s", n.ID)
	}
	if n.Grid == "five-up" || b.W < 198-.01 {
		return p, fmt.Errorf("metric.minimum_width: 198pt required")
	}
	m := n.DataMetric
	if m.Source != "" && strings.TrimSpace(m.Source) == "" {
		return p, fmt.Errorf("component.empty_content: %s.source", n.ID)
	}
	if (m.Value != "") == (m.Format != nil) {
		return p, fmt.Errorf("metric.exactly_one_value_or_format_required")
	}
	value := m.Value
	if m.Format != nil {
		var err error
		value, err = FormatNumber(*m.Format)
		if err != nil {
			return p, err
		}
	}
	if m.Format != nil || m.Target != "" || m.Status != "" || m.Source != "" {
		p.record.Definition = "data.metric+"
	}
	if len(m.Secondary) > 2 {
		return p, fmt.Errorf("metric.secondary_count: maximum two")
	}
	if len(m.Secondary) == 2 && b.W < 414-.01 {
		return p, fmt.Errorf("metric.two_secondaries_require_414pt")
	}
	keys := map[string]bool{}
	for _, sm := range m.Secondary {
		if !validPartKey(sm.Key) || keys[sm.Key] {
			return p, fmt.Errorf("metric.invalid_or_duplicate_secondary_key: %s", sm.Key)
		}
		keys[sm.Key] = true
	}
	y := b.Y
	add := func(part, text, style, role string, x, w float64, limit int) (TextRecord, error) {
		return r.planText(&p, n.ID+"."+part, text, style, surface, role, x, y, w, limit)
	}
	v, err := add("value", value, "stat", "display", b.X, b.W, 1)
	if err != nil {
		return p, err
	}
	y = v.Rect.Y + v.Rect.H + 6
	l, err := add("label", m.Label, "body", "secondary", b.X, b.W, 3)
	if err != nil {
		return p, err
	}
	y = l.Rect.Y + l.Rect.H + 6
	if m.Change != nil {
		if m.Change.Direction != "up" && m.Change.Direction != "down" {
			return p, fmt.Errorf("metric.unsupported_change_direction: %s", m.Change.Direction)
		}
		color, err := r.source.Ink(surface, "emphasis")
		if err != nil {
			return p, err
		}
		bg, err := r.source.Ink(surface, "bg")
		if err != nil {
			return p, err
		}
		if contrast(color, bg) < 3 {
			return p, fmt.Errorf("metric.insufficient_change_mark_contrast")
		}
		mark := Rect{b.X, y + 1, 8, 6}
		id := n.ID + ".change.marker"
		rotation := 0.0
		if m.Change.Direction == "down" {
			rotation = 180
		}
		p.record.Shapes = append(p.record.Shapes, ShapeRecord{ID: id, Rect: mark, Color: color, Geometry: "triangle", Rotation: rotation})
		p.record.Parts = append(p.record.Parts, id)
		tr, err := add("change.text", m.Change.Text, "label", "emphasis", b.X+14, b.W-14, 1)
		if err != nil {
			return p, err
		}
		y = tr.Rect.Y + tr.Rect.H + 6
	}
	if len(m.Secondary) > 0 {
		// 6pt stack gap plus source secondaryRow margin-top 6pt.
		y += 6
		if err := r.planShape(&p, n.ID+".secondary.divider", Rect{b.X, y, b.W, .75}, surface, "line"); err != nil {
			return p, err
		}
		y += .75 + 9
		track := (b.W - 18*float64(len(m.Secondary)-1)) / float64(len(m.Secondary))
		bottom := y
		for i, sm := range m.Secondary {
			x := b.X + float64(i)*(track+18)
			st, _ := r.source.Style("number")
			vl, err := r.typeEngine.Measure(sm.Value, st, track)
			if err != nil {
				return p, err
			}
			if len(vl.Lines) != 1 {
				return p, fmt.Errorf("metric.secondary_value_must_be_one_line")
			}
			w := vl.Lines[0].Advance + .01
			if track-w-9 <= 0 {
				return p, fmt.Errorf("metric.secondary_label_has_no_width")
			}
			part := "secondary." + sm.Key
			tr, err := add(part+".value", sm.Value, "number", "display", x, w, 1)
			if err != nil {
				return p, err
			}
			ls, _ := r.source.Style("small")
			ll, err := r.typeEngine.Measure(sm.Label, ls, track-w-9)
			if err != nil {
				return p, err
			}
			if len(ll.Lines) == 0 {
				return p, fmt.Errorf("component.empty_content: %s", part)
			}
			shift := tr.Layout.Lines[0].Baseline - ll.Lines[0].Baseline
			lab, err := r.planText(&p, n.ID+"."+part+".label", sm.Label, "small", surface, "secondary", x+w+9, y+shift, track-w-9, 3)
			if err != nil {
				return p, err
			}
			bottom = math.Max(bottom, math.Max(tr.Rect.Y+tr.Rect.H, lab.Rect.Y+lab.Rect.H))
		}
		y = bottom + 6
	}
	if m.Status != "" || m.Target != "" {
		x, w := b.X, b.W
		bottom := y
		if m.Status != "" {
			name, ok := map[string]string{"on": "On track", "risk": "At risk", "off": "Off track"}[m.Status]
			if !ok {
				return p, fmt.Errorf("metric.unsupported_status: %s", m.Status)
			}
			color, ok := r.source.Tokens.Colors.Dataviz.KPI[m.Status]
			if !ok {
				return p, fmt.Errorf("source.missing_kpi_color")
			}
			color = strings.TrimPrefix(color, "#")
			grounded := r.ink("light", "display")
			bg := r.ink(surface, "bg")
			if math.Max(contrast(grounded, bg), contrast(color, bg)) < 3 {
				return p, fmt.Errorf("metric.insufficient_status_mark_contrast")
			}
			p.record.Shapes = append(p.record.Shapes, ShapeRecord{ID: n.ID + ".status.outline", Rect: Rect{x, y + 2, 8, 8}, Color: grounded}, ShapeRecord{ID: n.ID + ".status.mark", Rect: Rect{x + 1, y + 3, 6, 6}, Color: color})
			p.record.Parts = append(p.record.Parts, n.ID+".status.outline", n.ID+".status.mark")
			x += 17
			w -= 17
			// Reserve a content-sized status label, then the source's 9pt gap.
			st, _ := r.source.Style("label")
			sl, err := r.typeEngine.Measure(name, st, w)
			if err != nil {
				return p, err
			}
			if len(sl.Lines) != 1 {
				return p, fmt.Errorf("metric.status_label_must_be_one_line")
			}
			sw := sl.Lines[0].Advance + 1 // keep clear of the conservative .25pt boundary guard.
			tr, err := add("status.label", name, "label", "primary", x, sw, 1)
			if err != nil {
				return p, err
			}
			bottom = math.Max(bottom, tr.Rect.Y+tr.Rect.H)
			x += sw + 9
			w = b.X + b.W - x
		}
		if m.Target != "" {
			if w <= 0 {
				return p, fmt.Errorf("metric.target_has_no_width")
			}
			tr, err := add("target", m.Target, "label", "secondary", x, w, 1)
			if err != nil {
				return p, err
			}
			bottom = math.Max(bottom, tr.Rect.Y+tr.Rect.H)
		}
		y = bottom + 6
	}
	if m.Source != "" {
		y += 3
		tr, err := add("source", "Source: "+m.Source, "source", "secondary", b.X, b.W, 2)
		if err != nil {
			return p, err
		}
		y = tr.Rect.Y + tr.Rect.H + 6
	}
	p.record.RequiredHeight = y - 6 - b.Y
	if b.H == 0 {
		b.H = math.Ceil(p.record.RequiredHeight/18) * 18
	}
	if p.record.RequiredHeight > b.H+.02 {
		return p, fmt.Errorf("component.vertical_overflow: %s needs %.3fpt, capacity %.3fpt", n.ID, p.record.RequiredHeight, b.H)
	}
	if err := r.source.Tokens.Grid.OuterBox(b); err != nil {
		return p, err
	}
	if !inside(b, zone) {
		return p, fmt.Errorf("node.outside_zone: %s", n.ID)
	}
	for _, tr := range p.texts {
		if !inside(tr.Rect, b) {
			return p, fmt.Errorf("component.part_outside_bounds: %s", tr.ID)
		}
	}
	for _, sh := range p.record.Shapes {
		if !inside(sh.Rect, b) {
			return p, fmt.Errorf("component.part_outside_bounds: %s", sh.ID)
		}
	}
	p.record.Rect = b
	return p, r.err
}
