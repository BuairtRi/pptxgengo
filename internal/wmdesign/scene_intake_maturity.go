package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

const IntakeMaturityContract = "pptxgengo.wmds-source-intake-maturity.v1"

// Freeze the working source for this adapter independently of future WM edits.
const IntakeMaturitySourceSHA256 = "63eca0027e5797a19be2b1236986772960889f2bb4ad1bb18a93a08416145c5b"
const IntakeMaturityRendererSHA256 = "a7f1f046d38723cb86d6e0dbacd97886c183784df1e4949192322e74d6a33cf6"

type maturityStage struct {
	Label  string          `json:"label"`
	Text   string          `json:"text,omitempty"`
	Number json.RawMessage `json:"n,omitempty"`
}
type maturityBranch struct {
	From   *int            `json:"from"`
	Number json.RawMessage `json:"n,omitempty"`
	Label  string          `json:"label"`
	Text   string          `json:"text,omitempty"`
}
type maturityHere struct {
	Stage *int   `json:"stage,omitempty"`
	Label string `json:"label,omitempty"`
}

type maturitySource struct {
	Type            string          `json:"type"`
	ID              string          `json:"id,omitempty"`
	X               float64         `json:"x"`
	Y               float64         `json:"y"`
	W               float64         `json:"w"`
	H               float64         `json:"h"`
	Stages          []maturityStage `json:"stages"`
	At              []float64       `json:"at,omitempty"`
	Headroom        *float64        `json:"headroom,omitempty"`
	Shape           *float64        `json:"shape,omitempty"`
	Active          *int            `json:"active,omitempty"`
	Axis            *bool           `json:"axis,omitempty"`
	AxisLabel       string          `json:"axisLabel,omitempty"`
	Inflection      *int            `json:"inflection,omitempty"`
	InflectionLabel string          `json:"inflectionLabel,omitempty"`
	Branch          *maturityBranch `json:"branch,omitempty"`
	LabelW          *float64        `json:"labelW,omitempty"`
	Here            *maturityHere   `json:"here,omitempty"`
	CanvasH         float64         `json:"_h,omitempty"`
}

func maturityNumber(raw json.RawMessage, fallback string) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return fallback, nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > 24 {
			return "", fmt.Errorf("scene.maturity_invalid_number")
		}
		return text, nil
	}
	var number float64
	if err := json.Unmarshal(raw, &number); err != nil || !intakeFinite(number) {
		return "", fmt.Errorf("scene.maturity_invalid_number")
	}
	return strconv.FormatFloat(number, 'f', -1, 64), nil
}

func maturityPoint(n maturitySource, plotH, shape, u float64) curvePoint {
	// expm1 retains the authored exponential shape for very shallow bends.
	f := math.Expm1(shape*u) / math.Expm1(shape)
	headroom := 0.
	if n.Headroom != nil {
		headroom = *n.Headroom
	}
	return curvePoint{n.X + u*n.W, n.Y + headroom + plotH - f*plotH*.9 - plotH*.05}
}

// maturityStroke keeps the line native and records its actual stroked extent.
// Round caps are small editable circles because ShapeLineProps has no cap field.
func (r *renderer) maturityStroke(p *scenePlan, id string, points []curvePoint, width float64, round bool) error {
	if len(points) < 2 {
		return fmt.Errorf("scene.maturity_invalid_path")
	}
	lo, hi := points[0], points[0]
	for _, q := range points {
		if !intakeFinite(q.x, q.y) {
			return fmt.Errorf("scene.maturity_nonfinite_geometry")
		}
		lo.x, lo.y = math.Min(lo.x, q.x), math.Min(lo.y, q.y)
		hi.x, hi.y = math.Max(hi.x, q.x), math.Max(hi.y, q.y)
	}
	b := Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
	props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: "070154"}, Width: width}}
	for i, q := range points {
		props.Points = append(props.Points, pptx.ShapePoint{X: pptx.Inches((q.x - b.X) / 72), Y: pptx.Inches((q.y - b.Y) / 72), MoveTo: ptrSceneBool(i == 0)})
	}
	extent := Rect{b.X - width/2, b.Y - width/2, b.W + width, b.H + width}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeCustGeom, Props: props, Record: ShapeRecord{ID: id, Rect: extent, Color: "070154", Geometry: "native-polyline"}}})
	if round {
		for i, q := range []curvePoint{points[0], points[len(points)-1]} {
			if err := r.diagramShape(p, fmt.Sprintf("%s.cap-%d", id, i+1), Rect{q.x - width/2, q.y - width/2, width, width}, pptx.ShapeTypeEllipse, "070154", "", 0, "", nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *renderer) maturityMarker(p *scenePlan, id string, q curvePoint, number string, active bool) error {
	fill, ink := "FFFFFF", "#070154"
	if active {
		fill, ink = "070154", "#FFFFFF"
	}
	// Source inset2pt stroke: inset the native centered stroke by1pt.
	b := Rect{q.x - 11, q.y - 11, 22, 22}
	if err := r.diagramShape(p, id+".circle", b, pptx.ShapeTypeEllipse, fill, "070154", 2, "", nil); err != nil {
		return err
	}
	p.Items[len(p.Items)-1].Shape.Record.Rect = Rect{q.x - 12, q.y - 12, 24, 24}
	st, err := r.sceneStyle("label")
	if err != nil {
		return err
	}
	st.Weight, st.Case, st.Tracking, st.TrackingPt = 600, "", "", 0
	st.Size = 11
	if utf8.RuneCountInString(number) > 2 {
		st.Size = 9
	}
	st.Leading = st.Size
	l, err := r.measureText(number, st, 20)
	if err != nil {
		return err
	}
	if len(l.Lines) != 1 {
		return fmt.Errorf("scene.maturity_marker_number_overflow: %s", id)
	}
	if isExpandedLibrary(r.source.Revision) {
		if r.contrastProbe != nil {
			prior := r.contrastProbe.SuppressChecks
			r.contrastProbe.SuppressChecks = true
			defer func() { r.contrastProbe.SuppressChecks = prior }()
		}
		// Center the native line in the complete marker box. Centering a
		// conservative allocation and then using native top alignment puts the
		// numeral's baseline near the circle center and its ink above it.
		if err = r.sceneText(p, id+".number", number, st, Rect{q.x - 10, q.y - 12, 20, 24}, "light", ink, "center"); err != nil {
			return err
		}
		p.Items[len(p.Items)-1].Text.VerticalAlign = "middle"
		if r.contrastProbe != nil {
			r.contrastProbe.SuppressChecks = false
			r.contrastAllows(id+".number", st, strings.TrimPrefix(ink, "#"), fill, textContrastMinimum(st))
		}
		return nil
	}
	return r.diagramStyledText(p, id+".number", number, st, Rect{q.x - 10, q.y - 12, 20, 24}, "light", ink, "center", true)
}

func (r *renderer) maturityStack(p *scenePlan, id, label, text string, x, y, w float64, align string, above bool, ctx SceneContext) error {
	stack := &scenePlan{}
	st, err := r.sceneStyle("label")
	if err != nil {
		return err
	}
	st.Weight = 600
	if err = r.sceneText(stack, id+".label", label, st, Rect{x, 0, w, 0}, ctx.Surface, "emphasis", align); err != nil {
		return err
	}
	height := 0.
	for _, it := range stack.Items {
		if it.Text != nil {
			height = math.Max(height, it.Text.Rect.Y+it.Text.Rect.H)
		}
	}
	if text != "" {
		st, err = r.sceneStyle("small")
		if err != nil {
			return err
		}
		if err = r.sceneText(stack, id+".text", text, st, Rect{x, height + 4, w, 0}, ctx.Surface, "primary", align); err != nil {
			return err
		}
		for _, it := range stack.Items {
			if it.Text != nil {
				height = math.Max(height, it.Text.Rect.Y+it.Text.Rect.H)
			}
		}
	}
	top := y
	if above {
		top -= height
	}
	for _, it := range stack.Items {
		if it.Text != nil {
			it.Text.Rect.Y += top
		}
		if it.Shape != nil {
			v := it.Shape.Props.Y.Val + top/72
			coord := pptx.Inches(v)
			it.Shape.Props.Y = &coord
			it.Shape.Record.Rect.Y += top
		}
	}
	p.Items = append(p.Items, stack.Items...)
	return nil
}

func (r *renderer) planIntakeMaturityScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "maturity" {
		return nil, false, nil
	}
	var n maturitySource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	bad := func(reason string) (*scenePlan, bool, error) {
		return nil, true, fmt.Errorf("scene.maturity_%s: %s", reason, id)
	}
	// Required geometry must contain numbers; decoding a missing/null scalar
	// as zero would make an incomplete source node look like an authored origin.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, true, err
	}
	for _, key := range []string{"x", "y", "w", "h"} {
		value, exists := fields[key]
		if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return bad("invalid_geometry")
		}
	}
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH) || n.W <= 0 || n.H <= 0 || n.W > 1920 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 || len(n.Stages) < 1 || len(n.Stages) > 12 {
		return bad("invalid_geometry_or_stages")
	}
	plotH := n.H
	if n.Axis == nil || *n.Axis {
		plotH -= 36
	}
	headroom := 0.
	if n.Headroom != nil {
		headroom = *n.Headroom
	}
	if !intakeFinite(headroom) || headroom < 0 || headroom >= plotH {
		return bad("headroom")
	}
	plotH -= headroom
	if plotH <= 0 {
		return bad("axis_height")
	}
	shape := 4.2
	if n.Shape != nil {
		shape = *n.Shape
	}
	if !intakeFinite(shape) || shape <= 0 || shape > 100 {
		return bad("shape")
	}
	ns := len(n.Stages)
	for _, s := range n.Stages {
		if strings.TrimSpace(s.Label) == "" || len(s.Label) > 4096 || len(s.Text) > 16384 {
			return bad("stage_text")
		}
	}
	indexOK := func(v *int) bool { return v == nil || *v >= 0 && *v < ns }
	if !indexOK(n.Active) || !indexOK(n.Inflection) || n.Here != nil && !indexOK(n.Here.Stage) {
		return bad("stage_index")
	}
	if len(n.AxisLabel) > 4096 || len(n.InflectionLabel) > 4096 || n.Here != nil && len(n.Here.Label) > 4096 {
		return bad("label_text")
	}
	if n.Branch != nil && (len(n.Branch.Label) > 4096 || len(n.Branch.Text) > 16384) {
		return bad("branch_text")
	}
	if n.Branch != nil && (n.Branch.From == nil || *n.Branch.From < 0 || *n.Branch.From >= ns || strings.TrimSpace(n.Branch.Label) == "") {
		return bad("branch")
	}
	at := n.At
	if len(at) == 0 {
		at = make([]float64, ns)
		for i := range at {
			at[i] = .5
			if ns > 1 {
				at[i] = .04 + math.Pow(float64(i)/float64(ns-1), .8)*.86
			}
		}
	}
	if len(at) != ns {
		return bad("stage_count")
	}
	positions := make([]curvePoint, ns)
	for i, u := range at {
		if !intakeFinite(u) || u < 0 || u > 1 || i > 0 && u <= at[i-1] {
			return bad("stage_positions")
		}
		positions[i] = maturityPoint(n, plotH, shape, u)
	}
	lw := math.Min(180, n.W/float64(ns)-12)
	if n.LabelW != nil {
		lw = *n.LabelW
	}
	if !intakeFinite(lw) || lw <= 0 || lw > 1920 {
		return bad("label_width")
	}
	keys, err := primitiveArrayKeys(ctx, "/stages", ns)
	if err != nil {
		return nil, true, err
	}
	if isExpandedLibrary(r.source.Revision) && ctx.Zone.H > 0 {
		// Reserve the actual frame's body top (the tall split column has its
		// own top). Increase internal headroom only as much as measured stacks
		// require; keep the authored diagram bottom, type, and stage positions.
		extra := 0.
		for i, stage := range n.Stages {
			lift := 0.
			liftCount := 2
			if ns >= 6 {
				liftCount = 3
			}
			if ns > 2 && i >= ns-liftCount {
				lift = 40
			}
			stack := &scenePlan{}
			if err = r.maturityStack(stack, id+".clearance", stage.Label, stage.Text, 0, positions[i].y-34-lift, lw, "left", true, ctx); err != nil {
				return nil, true, err
			}
			for _, item := range stack.Items {
				if item.Text == nil {
					continue
				}
				shortfall := ctx.Zone.Y - item.Text.Rect.Y
				if shortfall > 0 {
					f := math.Expm1(shape*at[i]) / math.Expm1(shape)
					extra = math.Max(extra, shortfall/(.05+.9*f))
				}
			}
		}
		if extra > 0 {
			if extra >= plotH {
				return bad("stage_label_clearance")
			}
			headroom += extra
			plotH -= extra
			n.Headroom = &headroom
			for i, u := range at {
				positions[i] = maturityPoint(n, plotH, shape, u)
			}
		}
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, n.H}}
	pts := make([]curvePoint, 61)
	for i := range pts {
		pts[i] = maturityPoint(n, plotH, shape, float64(i)/60)
	}
	if err = r.maturityStroke(p, id+".curve", pts, 2.5, true); err != nil {
		return nil, true, err
	}
	last, previous := pts[60], pts[58]
	angle := math.Atan2(last.y-previous.y, last.x-previous.x)
	for i, offset := range []float64{-.5, .5} {
		end := curvePoint{last.x - 10*math.Cos(angle+offset), last.y - 10*math.Sin(angle+offset)}
		if err = r.maturityStroke(p, fmt.Sprintf("%s.arrow-%d", id, i+1), []curvePoint{last, end}, 2.5, true); err != nil {
			return nil, true, err
		}
	}
	var branchPoint curvePoint
	if n.Branch != nil {
		start := positions[*n.Branch.From]
		end := curvePoint{n.X + n.W, start.y - math.Max(24, (n.Y+headroom+plotH-start.y)*.25)}
		control := curvePoint{(start.x + end.x) / 2, start.y - 10}
		b := Rect{start.x, end.y, end.x - start.x, start.y - end.y}
		props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id + ".branch.curve"}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: "070154"}, Width: 2}, Points: []pptx.ShapePoint{{X: pptx.Inches((start.x - b.X) / 72), Y: pptx.Inches((start.y - b.Y) / 72), MoveTo: ptrSceneBool(true)}, {X: pptx.Inches((end.x - b.X) / 72), Y: pptx.Inches((end.y - b.Y) / 72), Curve: &pptx.ShapeCurve{Type: "quadratic", X1: pptx.Inches((control.x - b.X) / 72), Y1: pptx.Inches((control.y - b.Y) / 72)}}}}
		p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeCustGeom, Props: props, Record: ShapeRecord{ID: id + ".branch.curve", Rect: Rect{b.X - 1, b.Y - 1, b.W + 2, b.H + 2}, Color: "070154", Geometry: "native-quadratic"}}})
		for i, q := range []curvePoint{start, end} {
			if err = r.diagramShape(p, fmt.Sprintf("%s.branch.cap-%d", id, i+1), Rect{q.x - 1, q.y - 1, 2, 2}, pptx.ShapeTypeEllipse, "070154", "", 0, "", nil); err != nil {
				return nil, true, err
			}
		}
		branchPoint = curvePoint{start.x + (end.x-start.x)*.72, start.y + (end.y-start.y)*.72 - 4}
	}
	if n.Axis == nil || *n.Axis {
		y := n.Y + n.H - 12
		if err = r.maturityStroke(p, id+".axis", []curvePoint{{n.X, y}, {n.X + n.W, y}}, 1, false); err != nil {
			return nil, true, err
		}
		if err = r.maturityStroke(p, id+".axis.arrow", []curvePoint{{n.X + n.W - 7, y - 4}, {n.X + n.W, y}, {n.X + n.W - 7, y + 4}}, 1, false); err != nil {
			return nil, true, err
		}
	}
	if (n.Axis == nil || *n.Axis) && n.AxisLabel != "" {
		st, e := r.sceneStyle("label")
		if e != nil {
			return nil, true, e
		}
		st.Weight = 600
		label := &scenePlan{}
		if err = r.sceneText(label, id+".axis.label", n.AxisLabel, st, Rect{n.X + n.W/2 - 90, n.Y + n.H - 18, 180, 0}, ctx.Surface, "display", "center"); err != nil {
			return nil, true, err
		}
		bg, e := r.sceneColor(ctx.Surface, "bg")
		if e != nil {
			return nil, true, e
		}
		if len(label.Items) > 0 {
			if err = r.diagramShape(p, id+".axis.label-bg", label.Items[0].Text.Rect, pptx.ShapeTypeRect, bg, "", 0, "", nil); err != nil {
				return nil, true, err
			}
		}
		p.Items = append(p.Items, label.Items...)
	}
	for i, s := range n.Stages {
		q := positions[i]
		prefix := id + ".stages." + keys[i]
		num, e := maturityNumber(s.Number, strconv.Itoa(i+1))
		if e != nil {
			return nil, true, e
		}
		if err = r.maturityMarker(p, prefix, q, num, n.Active != nil && *n.Active == i); err != nil {
			return nil, true, err
		}
		lift := 0.
		align := "left"
		x := math.Min(q.x-12, n.X+n.W-lw)
		liftCount := 2
		if isExpandedLibrary(r.source.Revision) && ns >= 6 {
			liftCount = 3
		}
		if ns > 2 && i >= ns-liftCount {
			lift = 40
			align = "right"
			right := math.Min(q.x+40, n.X+n.W)
			if isExpandedLibrary(r.source.Revision) && i+1 < ns {
				// Narrow curves can put the next stage's raised tick through
				// this label. Keep its text box four points before that tick.
				right = math.Min(right, positions[i+1].x-4)
			}
			x = math.Max(n.X, right-lw)
		}
		tipY := q.y - 30 - lift
		if err = r.diagramShape(p, prefix+".tick", Rect{q.x - .5, tipY, 1, 18 + lift}, pptx.ShapeTypeRect, "070154", "", 0, "", nil); err != nil {
			return nil, true, err
		}
		if err = r.maturityStack(p, prefix, s.Label, s.Text, x, tipY-4, lw, align, true, ctx); err != nil {
			return nil, true, err
		}
	}
	if n.Branch != nil {
		b := n.Branch
		num, e := maturityNumber(b.Number, "+")
		if e != nil {
			return nil, true, e
		}
		if err = r.maturityMarker(p, id+".branch", branchPoint, num, false); err != nil {
			return nil, true, err
		}
		if err = r.maturityStack(p, id+".branch", b.Label, b.Text, math.Min(branchPoint.x-12, n.X+n.W-lw), branchPoint.y+18, lw, "left", false, ctx); err != nil {
			return nil, true, err
		}
	}
	if n.Here != nil {
		index := 0
		if n.Active != nil {
			index = *n.Active
		}
		if n.Here.Stage != nil {
			index = *n.Here.Stage
		}
		q := positions[index]
		label := n.Here.Label
		if label == "" {
			label = "You are here"
		}
		if err = r.diagramShape(p, id+".here.tick", Rect{q.x - .75, q.y + 12, 1.5, 12}, pptx.ShapeTypeRect, "F900D3", "", 0, "", nil); err != nil {
			return nil, true, err
		}
		st, e := r.sceneStyle("label")
		if e != nil {
			return nil, true, e
		}
		st.Weight = 600
		layout, e := r.measureText(label, st, 148)
		if e != nil {
			return nil, true, e
		}
		width := 0.
		for _, line := range layout.Lines {
			width = math.Max(width, line.Advance)
		}
		if isExpandedLibrary(r.source.Revision) {
			width = sequenceInlineWidth(width, layout.Style)
		}
		height := math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight)
		b := Rect{q.x - (width+12)/2, q.y + 24, width + 12, height + 4}
		if err = r.diagramShape(p, id+".here.tag", b, pptx.ShapeTypeRect, "F900D3", "", 0, "", nil); err != nil {
			return nil, true, err
		}
		if err = r.sceneText(p, id+".here.label", label, st, Rect{b.X + 6, b.Y + 2, width, height}, "light", "#070154", "center"); err != nil {
			return nil, true, err
		}
	}
	if n.Inflection != nil {
		q := positions[*n.Inflection]
		label := n.InflectionLabel
		if label == "" {
			label = "Inflection"
		}
		st, e := r.sceneStyle("label")
		if e != nil {
			return nil, true, e
		}
		st.Weight = 600
		ink := "callout"
		if r.hasDensityContrastRules() {
			ink = "emphasis"
		}
		if err = r.sceneText(p, id+".inflection", label, st, Rect{q.x - 110, q.y + 40, 110, 0}, ctx.Surface, ink, "right"); err != nil {
			return nil, true, err
		}
	}
	p = diagramFinish(p, "maturity")
	if !intakeFinite(p.Bounds.X, p.Bounds.Y, p.Bounds.W, p.Bounds.H) {
		return bad("nonfinite_geometry")
	}
	for i := range p.Groups {
		if p.Groups[i].ID == id {
			p.Groups[i].Contract = IntakeMaturityContract
		}
	}
	return p, true, nil
}
