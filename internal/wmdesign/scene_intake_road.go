package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

type round12RoadMilestone struct {
	Label  string          `json:"label,omitempty"`
	Date   string          `json:"date,omitempty"`
	Text   string          `json:"text,omitempty"`
	At     *float64        `json:"at,omitempty"`
	Side   string          `json:"side,omitempty"`
	Active bool            `json:"active,omitempty"`
	Number json.RawMessage `json:"n,omitempty"`
}
type round12Road struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id,omitempty"`
	X          float64                `json:"x"`
	Y          float64                `json:"y"`
	W          float64                `json:"w"`
	H          float64                `json:"h"`
	Direction  string                 `json:"direction,omitempty"`
	RoadW      *float64               `json:"roadW,omitempty"`
	Amp        *float64               `json:"amp,omitempty"`
	Waves      *float64               `json:"waves,omitempty"`
	LabelW     *float64               `json:"labelW,omitempty"`
	Milestones []round12RoadMilestone `json:"milestones"`
	CanvasH    float64                `json:"_h,omitempty"`
}

func roadPoint(n round12Road, amp, waves, u float64) curvePoint {
	x := n.X + 18 + u*(n.W-36)
	if n.Direction == "left" {
		x = n.X + 18 + (1-u)*(n.W-36)
	}
	y := n.Y + 40 + (n.H-80)*(.82-.64*u+amp*math.Sin(u*math.Pi*2*waves))
	return curvePoint{x, y}
}
func (r *renderer) planIntakeRoadScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "road" {
		return nil, false, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, true, err
	}
	if err := round12RequiredNumbers(fields, "x", "y", "w", "h"); err != nil {
		return nil, true, err
	}
	var n round12Road
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	p, err := r.round12Road(id, n, ctx)
	return p, true, err
}

// Exact 7pt on/7pt off segments along the source polyline. Preset DrawingML
// dashes scale with stroke width, so native editable path subsegments preserve
// the source spacing without a raster or an unsupported custom-dash setting.
func (r *renderer) roadCenterline(p *scenePlan, id string, points []curvePoint) error {
	lo, hi := points[0], points[0]
	for _, q := range points[1:] {
		lo.x = math.Min(lo.x, q.x)
		lo.y = math.Min(lo.y, q.y)
		hi.x = math.Max(hi.x, q.x)
		hi.y = math.Max(hi.y, q.y)
	}
	b := Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
	props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: "FFFFFF"}, Width: 1.5}}
	remaining, on := 7., true
	newDash := true
	appendPoint := func(q curvePoint, move bool) {
		props.Points = append(props.Points, pptx.ShapePoint{X: pptx.Inches((q.x - b.X) / 72), Y: pptx.Inches((q.y - b.Y) / 72), MoveTo: ptrSceneBool(move)})
	}
	for i := 1; i < len(points); i++ {
		a, z := points[i-1], points[i]
		length := math.Hypot(z.x-a.x, z.y-a.y)
		if length == 0 {
			continue
		}
		done := 0.
		for done < length-1e-10 {
			take := math.Min(remaining, length-done)
			start := curvePoint{a.x + (z.x-a.x)*done/length, a.y + (z.y-a.y)*done/length}
			done += take
			end := curvePoint{a.x + (z.x-a.x)*done/length, a.y + (z.y-a.y)*done/length}
			if on {
				if newDash {
					appendPoint(start, true)
					newDash = false
				}
				appendPoint(end, false)
				if len(props.Points) > 8192 {
					return fmt.Errorf("scene.round12_road_dash_limit")
				}
			}
			remaining -= take
			if remaining < 1e-10 {
				remaining = 7
				on = !on
				newDash = true
			}
		}
	}
	extent := Rect{b.X - .75, b.Y - .75, b.W + 1.5, b.H + 1.5}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeCustGeom, Props: props, Record: ShapeRecord{ID: id, Rect: extent, Color: "FFFFFF", Geometry: "native-7pt-dashed-polyline"}}})
	p.Bounds = diagramUnion(p.Bounds, extent)
	return nil
}
func (r *renderer) round12Road(id string, n round12Road, ctx SceneContext) (*scenePlan, error) {
	bad := func(reason string) (*scenePlan, error) {
		return nil, fmt.Errorf("scene.round12_road_%s: %s", reason, id)
	}
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH) || n.W <= 36 || n.H <= 80 || n.W > 1920 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 || len(n.Milestones) < 1 || len(n.Milestones) > 24 {
		return bad("geometry_or_count")
	}
	if n.Direction != "" && n.Direction != "left" && n.Direction != "right" {
		return bad("direction")
	}
	width, amp, waves := 22., .16, 1.5
	if n.RoadW != nil {
		width = *n.RoadW
	}
	if n.Amp != nil {
		amp = *n.Amp
	}
	if n.Waves != nil {
		waves = *n.Waves
	}
	if !intakeFinite(width, amp, waves) || width <= 0 || width > 120 || math.Abs(amp) > 1 || waves <= 0 || waves > 12 {
		return bad("shape_options")
	}
	labelW := math.Min(180, (n.W-36)/float64(len(n.Milestones)))
	if n.LabelW != nil {
		labelW = *n.LabelW
	}
	if !intakeFinite(labelW) || labelW <= 0 || labelW > n.W {
		return bad("label_width")
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	if _, err := r.sceneColor(surface, "primary"); err != nil {
		return nil, err
	}
	keys, err := primitiveArrayKeys(ctx, "/milestones", len(n.Milestones))
	if err != nil {
		return nil, err
	}
	budget := 0
	for _, m := range n.Milestones {
		budget += utf8.RuneCountInString(m.Label) + utf8.RuneCountInString(m.Date) + utf8.RuneCountInString(m.Text)
		if utf8.RuneCountInString(m.Label) > 2048 || utf8.RuneCountInString(m.Date) > 2048 || utf8.RuneCountInString(m.Text) > 8192 || budget > 65536 || m.Side != "" && m.Side != "above" && m.Side != "below" || m.At != nil && (!intakeFinite(*m.At) || *m.At < 0 || *m.At > 1) {
			return bad("milestone_content_or_position")
		}
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, n.H}}
	points := make([]curvePoint, 121)
	for i := range points {
		q := roadPoint(n, amp, waves, float64(i)/120)
		points[i] = curvePoint{math.Round(q.x*10) / 10, math.Round(q.y*10) / 10}
	}
	if err = r.maturityStroke(p, id+".road", points, width, true); err != nil {
		return nil, err
	}
	// SVG's explicit round joins keep the thick road's ink within its measured
	// half-width envelope even at tight bends.
	p.Items[0].Shape.Props.Line.LineJoin = "round"
	for _, it := range p.Items {
		if it.Shape != nil {
			p.Bounds = diagramUnion(p.Bounds, it.Shape.Record.Rect)
		}
	}
	if err = r.roadCenterline(p, id+".centerline", points); err != nil {
		return nil, err
	}
	for i, m := range n.Milestones {
		u := .5
		if len(n.Milestones) > 1 {
			u = .06 + float64(i)/float64(len(n.Milestones)-1)*.88
		}
		if m.At != nil {
			u = *m.At
		}
		q := roadPoint(n, amp, waves, u)
		mid := id + ".milestones." + keys[i]
		fill := "FFFFFF"
		if m.Active {
			fill = "F900D3"
		}
		if err = r.diagramShape(p, mid+".pin", Rect{q.x - 12, q.y - 12, 24, 24}, pptx.ShapeTypeEllipse, fill, "070154", 2, "solid", nil); err != nil {
			return nil, err
		}
		p.Items[len(p.Items)-1].Shape.Record.Rect = Rect{q.x - 13, q.y - 13, 26, 26}
		p.Bounds = diagramUnion(p.Bounds, p.Items[len(p.Items)-1].Shape.Record.Rect)
		number, err := maturityNumber(m.Number, strconv.Itoa(i+1))
		if err != nil {
			return nil, err
		}
		st, err := r.sceneStyle("label")
		if err != nil {
			return nil, err
		}
		st.Size, st.Leading, st.Weight, st.TrackingPt, st.Case = 11, 11, 600, 0, ""
		if err = r.intakeVennSingleText(p, mid+".number", number, st, Rect{q.x - 11, q.y - 11, 22, 22}, surface, "#070154", "center", true); err != nil {
			return nil, err
		}
		above := i%2 == 0
		if m.Side != "" {
			above = m.Side == "above"
		}
		tickY := q.y + 13
		if above {
			tickY = q.y - 13 - 22
		}
		if err = r.diagramShape(p, mid+".tick", Rect{q.x - .5, tickY, 1, 22}, pptx.ShapeTypeRect, "070154", "", 0, "solid", nil); err != nil {
			return nil, err
		}
		x := math.Max(n.X, math.Min(q.x-labelW/2, n.X+n.W-labelW))
		child := &scenePlan{}
		parts := []round12TextPart{{m.Date, "label", "emphasis", "", 600}, {m.Label, "body", "display", "", 600}, {m.Text, "small", "primary", "", 0}}
		if err = r.round12Stack(child, mid+".label", parts, Rect{x, 0, labelW, 1080}, surface, "center", ctx); err != nil {
			return nil, err
		}
		if len(child.Items) > 0 {
			y0, y1 := math.Inf(1), math.Inf(-1)
			for _, it := range child.Items {
				if it.Text != nil {
					y0 = math.Min(y0, it.Text.Rect.Y)
					y1 = math.Max(y1, it.Text.Rect.Y+it.Text.Rect.H)
				}
			}
			targetY := q.y + 13 + 26
			if above {
				targetY = q.y - 13 - 26 - (y1 - y0)
			}
			for _, it := range child.Items {
				if it.Text != nil {
					it.Text.Rect.Y += targetY - y0
					p.Bounds = diagramUnion(p.Bounds, it.Text.Rect)
				}
			}
			p.Items = append(p.Items, child.Items...)
			p.Warnings = append(p.Warnings, child.Warnings...)
		}
	}
	p.Warnings = append(p.Warnings, IntakeRound12Contract+" renderer_sha256="+IntakeRound12RendererSHA256+"; editable winding road, native specimen review pending")
	sceneDataGroup(p, id, "diagram.round12.road", 0, p.Bounds)
	return p, nil
}
