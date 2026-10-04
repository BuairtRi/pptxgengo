package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

const IntakeCycleContract = "pptxgengo.wmds-source-intake-cycle.v1"
const IntakeCycleSourceSHA256 = "0c159e90675f86812ec73a897d0a7ace628e025e19bfb2f2a675814905288fe7"
const IntakeCycleRendererSHA256 = "3284f866ee67e37d8960a252fbc579ac10f332db1212f7806c98d70d45eacc64"

type cycleItem struct {
	Label   string          `json:"label"`
	Text    string          `json:"text,omitempty"`
	N       json.RawMessage `json:"n,omitempty"`
	Surface string          `json:"surface,omitempty"`
}
type cycleLoop struct {
	From  *int     `json:"from"`
	To    *int     `json:"to"`
	Label string   `json:"label,omitempty"`
	Side  *float64 `json:"side,omitempty"`
	Bend  *float64 `json:"bend,omitempty"`
	Ink   string   `json:"ink,omitempty"`
}
type cycleCenter struct {
	Label string   `json:"label,omitempty"`
	Title string   `json:"title,omitempty"`
	Text  string   `json:"text,omitempty"`
	W     *float64 `json:"w,omitempty"`
}
type cycleSource struct {
	Type     string       `json:"type"`
	ID       string       `json:"id,omitempty"`
	On       string       `json:"on,omitempty"`
	X        float64      `json:"x"`
	Y        float64      `json:"y"`
	W        float64      `json:"w"`
	H        float64      `json:"h"`
	CanvasH  float64      `json:"_h,omitempty"`
	Items    []cycleItem  `json:"items"`
	Loops    []cycleLoop  `json:"loops,omitempty"`
	Center   *cycleCenter `json:"center,omitempty"`
	Active   *int         `json:"active,omitempty"`
	NodeW    *float64     `json:"nodeW,omitempty"`
	NodeH    *float64     `json:"nodeH,omitempty"`
	Start    *float64     `json:"start,omitempty"`
	Closed   *bool        `json:"closed,omitempty"`
	Ring     *bool        `json:"ring,omitempty"`
	Numbered bool         `json:"numbered,omitempty"`
	Align    string       `json:"align,omitempty"`
	Surface  string       `json:"surface,omitempty"`
}

func cyclePoint(cx, cy, rx, ry, degrees float64) curvePoint {
	a := degrees * math.Pi / 180
	return curvePoint{cx + rx*math.Cos(a), cy + ry*math.Sin(a)}
}
func cycleClearDegrees(cx, cy, rx, ry, w, h, degrees float64, count int) float64 {
	c := cyclePoint(cx, cy, rx, ry, degrees)
	for d := 1.; d < 180/float64(count); d++ {
		q := cyclePoint(cx, cy, rx, ry, degrees+d)
		if math.Abs(q.x-c.x) > w/2+8 || math.Abs(q.y-c.y) > h/2+8 {
			return d
		}
	}
	return 180/float64(count) - 2
}
func cycleQuadraticPoint(a, c, b curvePoint, t float64) curvePoint {
	v := 1 - t
	return curvePoint{v*v*a.x + 2*v*t*c.x + t*t*b.x, v*v*a.y + 2*v*t*c.y + t*t*b.y}
}
func cycleQuadraticBounds(a, c, b curvePoint) Rect {
	lo, hi := curvePoint{math.Min(a.x, b.x), math.Min(a.y, b.y)}, curvePoint{math.Max(a.x, b.x), math.Max(a.y, b.y)}
	for _, v := range [][3]float64{{a.x, c.x, b.x}, {a.y, c.y, b.y}} {
		den := v[0] - 2*v[1] + v[2]
		if den == 0 {
			continue
		}
		t := (v[0] - v[1]) / den
		if t > 0 && t < 1 {
			q := cycleQuadraticPoint(a, c, b, t)
			lo.x = math.Min(lo.x, q.x)
			lo.y = math.Min(lo.y, q.y)
			hi.x = math.Max(hi.x, q.x)
			hi.y = math.Max(hi.y, q.y)
		}
	}
	return Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
}
func cycleOffset(p, c curvePoint, w, h float64) curvePoint {
	dx, dy := c.x-p.x, c.y-p.y
	length := math.Hypot(dx, dy)
	if length == 0 {
		length = 1
	}
	dx /= length
	dy /= length
	ex, ey := dx, dy
	if ex == 0 {
		ex = 1e-6
	}
	if ey == 0 {
		ey = 1e-6
	}
	k := math.Min((w/2+6)/math.Abs(ex), (h/2+6)/math.Abs(ey))
	return curvePoint{p.x + dx*k, p.y + dy*k}
}

func (r *renderer) cycleLine(p *scenePlan, id string, points []curvePoint, width float64, ink string, round bool) error {
	if len(points) < 2 {
		return fmt.Errorf("scene.cycle_invalid_path")
	}
	lo, hi := points[0], points[0]
	for _, q := range points {
		if !intakeFinite(q.x, q.y) {
			return fmt.Errorf("scene.cycle_nonfinite_path")
		}
		lo.x = math.Min(lo.x, q.x)
		lo.y = math.Min(lo.y, q.y)
		hi.x = math.Max(hi.x, q.x)
		hi.y = math.Max(hi.y, q.y)
	}
	b := Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
	props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: ink}, Width: width}}
	for i, q := range points {
		props.Points = append(props.Points, pptx.ShapePoint{X: pptx.Inches((q.x - b.X) / 72), Y: pptx.Inches((q.y - b.Y) / 72), MoveTo: ptrSceneBool(i == 0)})
	}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeCustGeom, Props: props, Record: ShapeRecord{ID: id, Rect: Rect{b.X - width/2, b.Y - width/2, b.W + width, b.H + width}, Color: ink, Geometry: "native-cycle-polyline"}}})
	if round {
		for i, q := range []curvePoint{points[0], points[len(points)-1]} {
			if e := r.diagramShape(p, fmt.Sprintf("%s.cap-%d", id, i+1), Rect{q.x - width/2, q.y - width/2, width, width}, pptx.ShapeTypeEllipse, ink, "", 0, "", nil); e != nil {
				return e
			}
		}
	}
	return nil
}
func (r *renderer) cycleArrow(p *scenePlan, id string, end, previous curvePoint, ink string) error {
	a := math.Atan2(end.y-previous.y, end.x-previous.x)
	for i, o := range []float64{-.5, .5} {
		q := curvePoint{end.x - 9*math.Cos(a+o), end.y - 9*math.Sin(a+o)}
		if e := r.cycleLine(p, fmt.Sprintf("%s.arrow-%d", id, i+1), []curvePoint{end, q}, 2, ink, true); e != nil {
			return e
		}
	}
	return nil
}

type cycleTextPart struct {
	id, text, style, ink string
	weight               int
}

// Source styled() preserves literal brackets and parses only footnote markers.
// Measure the same record that will be emitted before positioning a stack.
func (r *renderer) cycleTextRecord(id, text string, style Style, width float64, surface, ink, align string, ctx SceneContext) (TextRecord, error) {
	if strings.Contains(text, "[^") {
		child := &scenePlan{}
		if err := r.primitiveRichText(child, id, text, style, Rect{0, 0, width, 0}, surface, ink, align, "", "", ctx); err != nil {
			return TextRecord{}, err
		}
		if len(child.Items) != 1 || child.Items[0].Text == nil {
			return TextRecord{}, fmt.Errorf("scene.cycle_inline_copy")
		}
		return *child.Items[0].Text, nil
	}
	l, e := r.typeEngine.Measure(text, style, width)
	if e != nil {
		return TextRecord{}, e
	}
	color, e := r.sceneColor(surface, ink)
	if e != nil {
		return TextRecord{}, e
	}
	h := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
	return TextRecord{ID: id, Rect: Rect{0, 0, width, h}, Color: color, Align: align, Layout: l}, nil
}
func (r *renderer) cycleStack(p *scenePlan, parts []cycleTextPart, b Rect, gap float64, align, surface string, translateCenter bool, ctx SceneContext) error {
	var records []TextRecord
	total := 0.
	for _, part := range parts {
		if part.text == "" {
			continue
		}
		st, e := r.sceneStyle(part.style)
		if e != nil {
			return e
		}
		if part.weight > 0 {
			st.Weight = part.weight
		}
		record, e := r.cycleTextRecord(part.id, part.text, st, b.W, surface, part.ink, align, ctx)
		if e != nil {
			return e
		}
		h := math.Max(record.Layout.AllocationHeight, record.Layout.OccupiedTop+record.Layout.EstimatedOccupiedHeight)
		record.Rect.H = h
		if len(records) > 0 {
			total += gap
		}
		total += h
		records = append(records, record)
	}
	y := b.Y
	if translateCenter {
		y -= total * .25
	} else {
		if total > b.H+.02 {
			return fmt.Errorf("scene.cycle_node_text_overflow: %s needs%.3fpt capacity%.3fpt", records[0].ID, total, b.H)
		}
		y += (b.H - total) / 2
	}
	for _, record := range records {
		record.Rect.X, record.Rect.Y = b.X, y
		p.Items = append(p.Items, sceneItem{Text: &record})
		y += record.Rect.H + gap
	}
	return nil
}

func (r *renderer) planIntakeCycleScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type != "cycle" {
		return nil, false, nil
	}
	var n cycleSource
	if e := sceneDecode(raw, &n); e != nil {
		return nil, true, e
	}
	bad := func(why string) (*scenePlan, bool, error) {
		return nil, true, fmt.Errorf("scene.cycle_%s: %s", why, id)
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw, &fields); e != nil {
		return nil, true, e
	}
	for _, k := range []string{"x", "y", "w", "h"} {
		v, ok := fields[k]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return bad("invalid_geometry")
		}
	}
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH) || n.W <= 0 || n.H <= 0 || n.W > 1920 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 || len(n.Items) < 2 || len(n.Items) > 12 || len(n.Loops) > 32 {
		return bad("geometry_or_cardinality")
	}
	nw, nh, start := 162., 72., -90.
	if n.NodeW != nil {
		nw = *n.NodeW
	}
	if n.NodeH != nil {
		nh = *n.NodeH
	}
	if n.Start != nil {
		start = *n.Start
	}
	if !intakeFinite(nw, nh, start) || nw <= 24 || nh <= 12 || nw >= n.W || nh >= n.H || math.Abs(start) > 360000 {
		return bad("node_geometry_or_start")
	}
	start = math.Mod(start, 360)
	align := n.Align
	if align == "" {
		align = "center"
	}
	if align != "center" && align != "left" && align != "right" {
		return bad("alignment")
	}
	surface := ctx.Surface
	if n.On != "" {
		surface = n.On
	}
	if surface == "" {
		surface = "light"
	}
	if _, e := r.sceneColor(surface, "bg"); e != nil {
		return nil, true, e
	}
	if n.Active != nil && (*n.Active < 0 || *n.Active >= len(n.Items)) {
		return bad("active_index")
	}
	copyCount := 0
	copyOK := func(s string, max int) bool {
		count := utf8.RuneCountInString(s)
		copyCount += count
		return count <= max && copyCount <= 65536
	}
	for _, it := range n.Items {
		if strings.TrimSpace(it.Label) == "" || !copyOK(it.Label, 2048) || !copyOK(it.Text, 8192) {
			return bad("item_copy")
		}
	}
	for _, lp := range n.Loops {
		if lp.From == nil || lp.To == nil || *lp.From < 0 || *lp.To < 0 || *lp.From >= len(n.Items) || *lp.To >= len(n.Items) || !copyOK(lp.Label, 2048) {
			return bad("loop_index_or_copy")
		}
		if lp.Bend != nil && (!intakeFinite(*lp.Bend) || *lp.Bend < 0 || *lp.Bend > 2) {
			return bad("loop_bend")
		}
		if lp.Side != nil && *lp.Side != -1 && *lp.Side != 1 {
			return bad("loop_side")
		}
	}
	cx, cy, rx, ry := n.X+n.W/2, n.Y+n.H/2, (n.W-nw)/2, (n.H-nh)/2
	if n.Center != nil && (!copyOK(n.Center.Label, 2048) || !copyOK(n.Center.Title, 2048) || !copyOK(n.Center.Text, 8192)) {
		return bad("center_copy")
	}
	itemKeys, e := primitiveArrayKeys(ctx, "/items", len(n.Items))
	if e != nil {
		return nil, true, e
	}
	var loopKeys []string
	if len(n.Loops) > 0 {
		loopKeys, e = primitiveArrayKeys(ctx, "/loops", len(n.Loops))
		if e != nil {
			return nil, true, e
		}
	} else if _, supplied := ctx.Keys[ctx.Path+"/loops"]; supplied {
		return bad("empty_loop_keys")
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, n.H}}
	angles := make([]float64, len(n.Items))
	for i := range angles {
		angles[i] = start + float64(i)*360/float64(len(n.Items))
	}
	if n.Ring == nil || *n.Ring {
		for i, a := range angles {
			if n.Closed != nil && !*n.Closed && i == len(angles)-1 {
				continue
			}
			b := start + 360
			if i < len(angles)-1 {
				b = angles[i+1]
			}
			from, to := a+cycleClearDegrees(cx, cy, rx, ry, nw, nh, a, len(angles)), b-cycleClearDegrees(cx, cy, rx, ry, nw, nh, math.Mod(b, 360), len(angles))
			if to <= from {
				return bad("arc_clearance")
			}
			points := make([]curvePoint, 25)
			for j := range points {
				points[j] = cyclePoint(cx, cy, rx, ry, from+(to-from)*float64(j)/24)
			}
			pre := id + ".items." + itemKeys[i] + ".arc"
			if e = r.cycleLine(p, pre, points, 2, "070154", true); e != nil {
				return nil, true, e
			}
			if e = r.cycleArrow(p, pre, points[24], points[23], "070154"); e != nil {
				return nil, true, e
			}
		}
	}
	type loopLabel struct {
		id, label string
		q         curvePoint
	}
	var loopLabels []loopLabel
	for i, lp := range n.Loops {
		a, b := cyclePoint(cx, cy, rx, ry, angles[*lp.From]), cyclePoint(cx, cy, rx, ry, angles[*lp.To])
		mx, my := (a.x+b.x)/2, (a.y+b.y)/2
		bend := .9
		if lp.Bend != nil {
			bend = *lp.Bend
		}
		c := curvePoint{mx + (cx-mx)*bend, my + (cy-my)*bend}
		if math.Hypot(mx-cx, my-cy) < 24 {
			dx, dy := b.x-a.x, b.y-a.y
			length := math.Hypot(dx, dy)
			if length == 0 {
				length = 1
			}
			side := 1.
			if lp.Side != nil {
				side = *lp.Side
			}
			c = curvePoint{mx - dy/length*ry*.9*side, my + dx/length*ry*.9*side}
		}
		a, b = cycleOffset(a, c, nw, nh), cycleOffset(b, c, nw, nh)
		ink := "0047FF"
		if lp.Ink != "" {
			ink, e = r.sceneColor(surface, lp.Ink)
			if e != nil {
				return nil, true, e
			}
		}
		pre := id + ".loops." + loopKeys[i]
		bounds := cycleQuadraticBounds(a, c, b)
		props := pptx.ShapeProps{PositionProps: pos(bounds), ObjectNameProps: pptx.ObjectNameProps{ObjectName: pre + ".curve"}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: ink}, Width: 1.5, DashType: "dash"}, Points: []pptx.ShapePoint{{X: pptx.Inches((a.x - bounds.X) / 72), Y: pptx.Inches((a.y - bounds.Y) / 72), MoveTo: ptrSceneBool(true)}, {X: pptx.Inches((b.x - bounds.X) / 72), Y: pptx.Inches((b.y - bounds.Y) / 72), Curve: &pptx.ShapeCurve{Type: "quadratic", X1: pptx.Inches((c.x - bounds.X) / 72), Y1: pptx.Inches((c.y - bounds.Y) / 72)}}}}
		p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeCustGeom, Props: props, Record: ShapeRecord{ID: pre + ".curve", Rect: Rect{bounds.X - .75, bounds.Y - .75, bounds.W + 1.5, bounds.H + 1.5}, Color: ink, Geometry: "native-cycle-quadratic"}}})
		if e = r.cycleArrow(p, pre, b, cycleQuadraticPoint(a, c, b, .94), ink); e != nil {
			return nil, true, e
		}
		if lp.Label != "" {
			loopLabels = append(loopLabels, loopLabel{pre, lp.Label, cycleQuadraticPoint(a, c, b, .5)})
		}
	}
	for _, label := range loopLabels {
		st, e := r.sceneStyle("label")
		if e != nil {
			return nil, true, e
		}
		st.Weight = 600
		record, e := r.cycleTextRecord(label.id+".label", label.label, st, 132, surface, "emphasis", "center", ctx)
		layout := record.Layout
		if e != nil {
			return nil, true, e
		}
		if len(layout.Lines) != 1 {
			return bad("loop_label_overflow")
		}
		width := sequenceInlineWidth(layout.Lines[0].Advance, st)
		if width > 132+.02 {
			return bad("loop_label_overflow")
		}
		height := math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight)
		if height > 18+.02 {
			return bad("loop_label_overflow")
		}
		b := Rect{label.q.x - (width+8)/2, label.q.y - height/2, width + 8, height}
		if e = r.sceneRect(p, label.id+".label-bg", b, surface); e != nil {
			return nil, true, e
		}
		record.Rect = Rect{b.X + 4, b.Y, width, height}
		p.Items = append(p.Items, sceneItem{Text: &record})
	}
	if n.Center != nil {
		cw := math.Min(rx*1.1, 234)
		if n.Center.W != nil {
			cw = *n.Center.W
		}
		if !intakeFinite(cw) || cw <= 0 || cw > n.W {
			return bad("center_width")
		}
		parts := []cycleTextPart{{id + ".center.label", n.Center.Label, "label", "emphasis", 600}, {id + ".center.title", n.Center.Title, "subhead", "display", 0}, {id + ".center.text", n.Center.Text, "small", "primary", 0}}
		if e = r.cycleStack(p, parts, Rect{cx - cw/2, cy - 30, cw, 0}, 4, "center", surface, true, ctx); e != nil {
			return nil, true, e
		}
	}
	for i, it := range n.Items {
		pre := id + ".items." + itemKeys[i]
		q := cyclePoint(cx, cy, rx, ry, angles[i])
		sf := it.Surface
		if sf == "" {
			sf = n.Surface
			if sf == "" {
				sf = "subtle"
			}
			if n.Active != nil && *n.Active == i {
				sf = "inverse"
			}
		}
		b := Rect{q.x - nw/2, q.y - nh/2, nw, nh}
		if e = r.sceneRect(p, pre+".surface", b, sf); e != nil {
			return nil, true, e
		}
		num := ""
		if len(it.N) > 0 && !bytes.Equal(bytes.TrimSpace(it.N), []byte("null")) {
			num, e = maturityNumber(it.N, "")
			if e != nil {
				return bad("item_number")
			}
		} else if n.Numbered {
			num = fmt.Sprintf("%02d", i+1)
		}
		style := "subhead"
		if nh < 60 {
			style = "body"
		}
		parts := []cycleTextPart{{pre + ".number", num, "label", "emphasis", 600}, {pre + ".label", it.Label, style, "display", 600}, {pre + ".text", it.Text, "small", "primary", 0}}
		if e = r.cycleStack(p, parts, Rect{b.X + 12, b.Y + 6, b.W - 24, b.H - 12}, 2, align, sf, false, ctx); e != nil {
			return nil, true, e
		}
	}
	p.Warnings = append(p.Warnings, IntakeCycleContract+"; native editable cycle; loop dash uses DrawingML dash preset rather than exact source5pt/4pt spacing; native review pending")
	p = diagramFinish(p, "cycle")
	for i := range p.Groups {
		if p.Groups[i].ID == id {
			p.Groups[i].Contract = IntakeCycleContract
		}
	}
	if !intakeFinite(p.Bounds.X, p.Bounds.Y, p.Bounds.W, p.Bounds.H) {
		return bad("nonfinite_geometry")
	}
	return p, true, nil
}
