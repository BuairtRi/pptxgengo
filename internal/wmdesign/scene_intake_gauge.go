package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

const IntakeGaugeContract = "pptxgengo.wmds-source-intake-gauge.v1"
const IntakeGaugeRendererSHA256 = "72057c4f23f077f095e7b6f22705b35bbb538eee55d9d9da6cf752f6cc10d20c"
const IntakeGaugeSourceSHA256 = "50c9b7c2bbd968fda1adb78ff0c4efdd2f823df0d17e6043514d442834c033ac"

type intakeGaugeSource struct {
	Type      string   `json:"type"`
	ID        string   `json:"id,omitempty"`
	X         float64  `json:"x"`
	Y         float64  `json:"y"`
	W         float64  `json:"w"`
	Segments  []string `json:"segments"`
	Value     *float64 `json:"value"`
	ValueText string   `json:"valueText"`
	Caption   string   `json:"caption"`
	Ends      []string `json:"ends"`
	CanvasH   float64  `json:"_h,omitempty"`
}

func gaugeArcBounds(cx, cy, r, inner, a, b float64) Rect {
	lo, hi := curvePoint{math.Inf(1), math.Inf(1)}, curvePoint{math.Inf(-1), math.Inf(-1)}
	add := func(angle, radius float64) {
		q := curvePoint{cx + radius*math.Cos(angle*math.Pi/180), cy + radius*math.Sin(angle*math.Pi/180)}
		lo.x = math.Min(lo.x, q.x)
		lo.y = math.Min(lo.y, q.y)
		hi.x = math.Max(hi.x, q.x)
		hi.y = math.Max(hi.y, q.y)
	}
	for _, radius := range []float64{r, inner} {
		add(a, radius)
		add(b, radius)
		for _, angle := range []float64{180, 270, 360} {
			if angle >= a && angle <= b {
				add(angle, radius)
			}
		}
	}
	return Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
}
func (r *renderer) planIntakeGaugeScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "gauge" {
		return nil, false, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, true, err
	}
	if err := round12RequiredNumbers(fields, "x", "y", "w", "value"); err != nil {
		return nil, true, err
	}
	var n intakeGaugeSource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	p, err := r.intakeGauge(id, n, ctx)
	return p, true, err
}
func (r *renderer) intakeGauge(id string, n intakeGaugeSource, ctx SceneContext) (*scenePlan, error) {
	bad := func(reason string) (*scenePlan, error) { return nil, fmt.Errorf("scene.gauge_%s: %s", reason, id) }
	if !intakeFinite(n.X, n.Y, n.W, n.CanvasH) || n.W < 24 || n.W > 1920 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 || n.Value == nil || !intakeFinite(*n.Value) || *n.Value < 0 || *n.Value > 1 || len(n.Segments) < 1 || len(n.Segments) > 12 || len(n.Ends) != 2 {
		return bad("geometry_value_or_count")
	}
	if utf8.RuneCountInString(n.ValueText) > 2048 || utf8.RuneCountInString(n.Caption) > 8192 || utf8.RuneCountInString(n.Ends[0]) > 2048 || utf8.RuneCountInString(n.Ends[1]) > 2048 {
		return bad("copy_limit")
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	if _, err := r.sceneColor(surface, "bg"); err != nil {
		return nil, err
	}
	// Colors are constrained to the pinned palette rather than the browser's
	// off-palette-warning fallback.
	colors := make([]string, len(n.Segments))
	for i, ref := range n.Segments {
		if ref == "" {
			return bad("segment_palette")
		}
		color, err := r.sceneColor(surface, ref)
		if err != nil {
			return nil, err
		}
		colors[i] = color
	}
	radius := n.W / 2
	cx, cy := n.X+radius, n.Y+radius
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, radius + 7}}
	dark := surface == "inverse" || surface == "deep"
	for i, color := range colors {
		a, b := 180+float64(i)*180/float64(len(colors)), 180+float64(i+1)*180/float64(len(colors))
		if i > 0 {
			a += .8
		}
		if i < len(colors)-1 {
			b -= .8
		}
		box := Rect{n.X, n.Y, n.W, n.W}
		outline := ""
		weight := 0.
		if !dark {
			outline, weight = "070154", .75
		}
		if err := r.diagramShape(p, fmt.Sprintf("%s.segment-%02d", id, i+1), box, pptx.ShapeTypeBlockArc, color, outline, weight, "solid", nil); err != nil {
			return nil, err
		}
		shape := p.Items[len(p.Items)-1].Shape
		shape.Props.AngleRange = &[2]float64{a, b}
		shape.Props.ArcThicknessRatio = .4
		inkBox := gaugeArcBounds(cx, cy, radius, radius*.6, a, b)
		// Closed ring bands have right-angle joins; sqrt(2) covers their miter
		// corners in addition to the half-stroke curved edges.
		guard := weight / 2 * math.Sqrt2
		shape.Record.Rect = Rect{inkBox.X - guard, inkBox.Y - guard, inkBox.W + 2*guard, inkBox.H + 2*guard}
	}
	// Do not retain the preset arc's empty lower semicircle as visible bounds.
	p.Bounds = Rect{n.X, n.Y, n.W, radius + 7}
	for _, item := range p.Items {
		p.Bounds = diagramUnion(p.Bounds, item.Shape.Record.Rect)
	}
	angle := math.Pi + *n.Value*math.Pi
	length := radius * .92
	round := func(v float64) float64 { return math.Round(v*100) / 100 }
	tip := curvePoint{round(cx + length*math.Cos(angle)), round(cy + length*math.Sin(angle))}
	a := curvePoint{round(cx + 6*math.Cos(angle+math.Pi/2)), round(cy + 6*math.Sin(angle+math.Pi/2))}
	b := curvePoint{round(cx - 6*math.Cos(angle+math.Pi/2)), round(cy - 6*math.Sin(angle+math.Pi/2))}
	lo, hi := curvePoint{math.Min(tip.x, math.Min(a.x, b.x)), math.Min(tip.y, math.Min(a.y, b.y))}, curvePoint{math.Max(tip.x, math.Max(a.x, b.x)), math.Max(tip.y, math.Max(a.y, b.y))}
	needle := Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
	ink := "070154"
	if dark {
		ink = "FFFFFF"
	}
	if err := r.diagramShape(p, id+".needle", needle, pptx.ShapeTypeCustGeom, ink, "", 0, "solid", [][2]float64{{tip.x - needle.X, tip.y - needle.Y}, {a.x - needle.X, a.y - needle.Y}, {b.x - needle.X, b.y - needle.Y}}); err != nil {
		return nil, err
	}
	if err := r.diagramShape(p, id+".hub", Rect{cx - 7, cy - 7, 14, 14}, pptx.ShapeTypeEllipse, ink, "", 0, "solid", nil); err != nil {
		return nil, err
	}
	endStyle, err := r.sceneStyle("label")
	if err != nil {
		return nil, err
	}
	widths := [2]float64{}
	var ends [2]*TextRecord
	for i, text := range n.Ends {
		if text == "" {
			continue
		}
		child := &scenePlan{}
		if err := r.round12Stack(child, fmt.Sprintf("%s.end-%d", id, i+1), []round12TextPart{{text, "label", "secondary", "", 0}}, Rect{n.X, 0, n.W, 1080}, surface, "left", ctx); err != nil {
			return nil, err
		}
		record := *child.Items[0].Text
		if len(record.Layout.Lines) != 1 {
			return bad("end_label_overflow")
		}
		widths[i] = sequenceInlineWidth(record.Layout.Lines[0].Advance, endStyle)
		record.Rect.W = widths[i]
		ends[i] = &record
		p.Warnings = append(p.Warnings, child.Warnings...)
	}
	if widths[0]+widths[1] > n.W+.02 {
		return bad("end_label_overlap")
	}
	for i, record := range ends {
		if record == nil {
			continue
		}
		record.Rect.X, record.Rect.Y = n.X, n.Y+radius+9
		if i == 1 {
			record.Rect.X = n.X + n.W - widths[i]
			record.Align = "right"
		}
		p.Items = append(p.Items, sceneItem{Text: record})
		p.Bounds = diagramUnion(p.Bounds, record.Rect)
	}
	// The source value stack starts27pt below the dial baseline, with3pt gap.
	// Measure it first so a wrapping caption expands the reported ink extent.
	stack := &scenePlan{}
	if err := r.round12Stack(stack, id+".value", []round12TextPart{{n.ValueText, "stat-sm", "display", "", 0}, {n.Caption, "small", "secondary", "", 0}}, Rect{n.X, 0, n.W, 1080}, surface, "center", ctx); err != nil {
		return nil, err
	}
	if len(stack.Items) > 0 {
		top := math.Inf(1)
		for _, it := range stack.Items {
			if it.Text != nil {
				top = math.Min(top, it.Text.Rect.Y)
			}
		}
		first := true
		for _, it := range stack.Items {
			if it.Text != nil {
				it.Text.Rect.Y += n.Y + radius + 27 - top
				if !first {
					it.Text.Rect.Y++
				}
				first = false
				p.Bounds = diagramUnion(p.Bounds, it.Text.Rect)
			}
		}
		p.Items = append(p.Items, stack.Items...)
		p.Warnings = append(p.Warnings, stack.Warnings...)
	}
	p.Warnings = append(p.Warnings, IntakeGaugeContract+" renderer_sha256="+IntakeGaugeRendererSHA256+"; native semicircle bands and needle, native specimen review pending")
	sceneDataGroup(p, id, "gauge", 0, p.Bounds)
	return p, nil
}
