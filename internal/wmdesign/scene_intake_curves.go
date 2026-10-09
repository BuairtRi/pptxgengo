package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/buairtri/pptxgengo/pptx"
)

// The next source revision is the only default opt-in; current frozen bundles
// retain Catmull unless the authored node explicitly requests monotone.
const IntakeTeamCurveMonotoneRevision = "wmds-library.v4"

type teamCurveSeries struct {
	Name    string          `json:"name"`
	Values  []float64       `json:"values"`
	Fill    string          `json:"fill,omitempty"`
	Style   string          `json:"style,omitempty"`
	Dashed  bool            `json:"dashed,omitempty"`
	Hatch   bool            `json:"hatch,omitempty"`
	LabelAt json.RawMessage `json:"labelAt,omitempty"`
}
type teamCurvePhase struct {
	Label string   `json:"label"`
	Sub   string   `json:"sub,omitempty"`
	At    *float64 `json:"at,omitempty"`
	Point string   `json:"point,omitempty"`
}
type teamCurveSource struct {
	Type        string            `json:"type"`
	ID          string            `json:"id,omitempty"`
	X           float64           `json:"x"`
	Y           float64           `json:"y"`
	W           float64           `json:"w"`
	H           float64           `json:"h"`
	At          []float64         `json:"at,omitempty"`
	Max         *float64          `json:"max,omitempty"`
	Curve       string            `json:"curve,omitempty"`
	LineBasis   string            `json:"lineBasis,omitempty"`
	Smooth      *float64          `json:"smooth,omitempty"`
	Tension     *float64          `json:"tension,omitempty"`
	PhaseH      *float64          `json:"phaseH,omitempty"`
	PhaseLabels *bool             `json:"phaseLabels,omitempty"`
	Series      []teamCurveSeries `json:"series"`
	Phases      []teamCurvePhase  `json:"phases,omitempty"`
	Unit        string            `json:"unit,omitempty"`
	TimeUnit    string            `json:"timeUnit,omitempty"`
	Source      string            `json:"source,omitempty"`
	PointLabels []string          `json:"pointLabels,omitempty"`
	CanvasH     float64           `json:"_h,omitempty"`
}
type curvePoint struct{ x, y float64 }

func teamCurvePath(points []curvePoint, tension float64, move bool) []pptx.ShapePoint {
	out := make([]pptx.ShapePoint, 0, len(points))
	for i, p := range points {
		q := pptx.ShapePoint{X: pptx.Inches(p.x / 72), Y: pptx.Inches(p.y / 72)}
		if i == 0 {
			q.MoveTo = ptrSceneBool(move)
		} else {
			p0, p1, p2, p3 := points[max(0, i-2)], points[i-1], p, points[min(len(points)-1, i+1)]
			t := tension / 3
			q.Curve = &pptx.ShapeCurve{Type: "cubic", X1: pptx.Inches((p1.x + (p2.x-p0.x)*t) / 72), Y1: pptx.Inches((p1.y + (p2.y-p0.y)*t) / 72), X2: pptx.Inches((p2.x - (p3.x-p1.x)*t) / 72), Y2: pptx.Inches((p2.y - (p3.y-p1.y)*t) / 72)}
		}
		out = append(out, q)
	}
	return out
}

// teamCurveMonotonePath reproduces the current source renderer's weighted
// harmonic Fritsch–Carlson tangents. Reverse band edges swap each cubic's
// controls so their curve is identical in either drawing direction.
func teamCurveMonotonePath(points []curvePoint, move bool) []pptx.ShapePoint {
	rev := points[0].x > points[len(points)-1].x
	forward := append([]curvePoint(nil), points...)
	if rev {
		for i, j := 0, len(forward)-1; i < j; i, j = i+1, j-1 {
			forward[i], forward[j] = forward[j], forward[i]
		}
	}
	n := len(forward)
	dx := make([]float64, n-1)
	slopes := make([]float64, n-1)
	tangents := make([]float64, n)
	for i := range dx {
		dx[i] = forward[i+1].x - forward[i].x
		slopes[i] = (forward[i+1].y - forward[i].y) / dx[i]
	}
	tangents[0], tangents[n-1] = slopes[0], slopes[n-2]
	for i := 1; i < n-1; i++ {
		if slopes[i-1] == 0 || slopes[i] == 0 || math.Signbit(slopes[i-1]) != math.Signbit(slopes[i]) {
			continue
		}
		tangents[i] = 3 * (dx[i-1] + dx[i]) / ((2*dx[i]+dx[i-1])/slopes[i-1] + (dx[i]+2*dx[i-1])/slopes[i])
	}
	out := make([]pptx.ShapePoint, n)
	for i, p := range points {
		out[i] = pptx.ShapePoint{X: pptx.Inches(p.x / 72), Y: pptx.Inches(p.y / 72)}
		if i == 0 {
			out[i].MoveTo = ptrSceneBool(move)
			continue
		}
		k := i - 1
		if rev {
			k = n - 1 - i
		}
		a, b := forward[k], forward[k+1]
		c1 := curvePoint{a.x + dx[k]/3, a.y + tangents[k]*dx[k]/3}
		c2 := curvePoint{b.x - dx[k]/3, b.y - tangents[k+1]*dx[k]/3}
		if rev {
			c1, c2 = c2, c1
		}
		out[i].Curve = &pptx.ShapeCurve{Type: "cubic", X1: pptx.Inches(c1.x / 72), Y1: pptx.Inches(c1.y / 72), X2: pptx.Inches(c2.x / 72), Y2: pptx.Inches(c2.y / 72)}
	}
	return out
}

func teamCurveNativeBounds(path []pptx.ShapePoint) Rect {
	lo, hi := curvePoint{path[0].X.Val * 72, path[0].Y.Val * 72}, curvePoint{path[0].X.Val * 72, path[0].Y.Val * 72}
	include := func(p curvePoint) {
		lo.x, lo.y = math.Min(lo.x, p.x), math.Min(lo.y, p.y)
		hi.x, hi.y = math.Max(hi.x, p.x), math.Max(hi.y, p.y)
	}
	for i := 1; i < len(path); i++ {
		a, b := path[i-1], path[i]
		p0, p1 := curvePoint{a.X.Val * 72, a.Y.Val * 72}, curvePoint{b.X.Val * 72, b.Y.Val * 72}
		include(p1)
		c := b.Curve
		if c == nil {
			continue
		}
		c1, c2 := curvePoint{c.X1.Val * 72, c.Y1.Val * 72}, curvePoint{c.X2.Val * 72, c.Y2.Val * 72}
		roots := append(teamCurveExtrema(p0.x, c1.x, c2.x, p1.x), teamCurveExtrema(p0.y, c1.y, c2.y, p1.y)...)
		for _, u := range roots {
			v := 1 - u
			include(curvePoint{v*v*v*p0.x + 3*v*v*u*c1.x + 3*v*u*u*c2.x + u*u*u*p1.x, v*v*v*p0.y + 3*v*v*u*c1.y + 3*v*u*u*c2.y + u*u*u*p1.y})
		}
	}
	return Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
}

func teamCurveFinitePath(path []pptx.ShapePoint) bool {
	for _, p := range path {
		if !intakeFinite(p.X.Val, p.Y.Val) {
			return false
		}
		if c := p.Curve; c != nil && !intakeFinite(c.X1.Val, c.Y1.Val, c.X2.Val, c.Y2.Val) {
			return false
		}
	}
	return true
}

// teamCurveBounds evaluates actual cubic extrema, not the control polygon.
// This preserves Catmull-Rom overshoot while exposing it to allocation checks.
func teamCurveBounds(points []curvePoint, tension float64) Rect {
	lo, hi := points[0], points[0]
	include := func(p curvePoint) {
		lo.x, lo.y = math.Min(lo.x, p.x), math.Min(lo.y, p.y)
		hi.x, hi.y = math.Max(hi.x, p.x), math.Max(hi.y, p.y)
	}
	for i := 1; i < len(points); i++ {
		p0, p1, p2, p3 := points[max(0, i-2)], points[i-1], points[i], points[min(len(points)-1, i+1)]
		t := tension / 3
		c1 := curvePoint{p1.x + (p2.x-p0.x)*t, p1.y + (p2.y-p0.y)*t}
		c2 := curvePoint{p2.x - (p3.x-p1.x)*t, p2.y - (p3.y-p1.y)*t}
		include(p2)
		roots := append(teamCurveExtrema(p1.x, c1.x, c2.x, p2.x), teamCurveExtrema(p1.y, c1.y, c2.y, p2.y)...)
		for _, u := range roots {
			v := 1 - u
			include(curvePoint{v*v*v*p1.x + 3*v*v*u*c1.x + 3*v*u*u*c2.x + u*u*u*p2.x,
				v*v*v*p1.y + 3*v*v*u*c1.y + 3*v*u*u*c2.y + u*u*u*p2.y})
		}
	}
	return Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
}
func teamCurveExtrema(p0, c1, c2, p1 float64) []float64 {
	a, b, c := -p0+3*c1-3*c2+p1, 2*(p0-2*c1+c2), c1-p0
	var roots []float64
	add := func(t float64) {
		if t > 0 && t < 1 {
			roots = append(roots, t)
		}
	}
	if math.Abs(a) < 1e-12 {
		if math.Abs(b) > 1e-12 {
			add(-c / b)
		}
		return roots
	}
	d := b*b - 4*a*c
	if d < 0 {
		return roots
	}
	// Stable quadratic roots avoid cancelling away shallow extrema.
	q := -.5 * (b + math.Copysign(math.Sqrt(d), b))
	if q == 0 {
		add(-b / (2 * a))
	} else {
		add(q / a)
		add(c / q)
	}
	return roots
}

func (r *renderer) planIntakeCurveScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "teamcurve" {
		return nil, false, nil
	}
	var n teamCurveSource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	bad := func(reason string) (*scenePlan, bool, error) {
		return nil, true, fmt.Errorf("scene.teamcurve_%s: %s", reason, id)
	}
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
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
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH) || n.W <= 0 || n.H <= 0 || n.W > 1920 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 || len(n.Series) == 0 || len(n.Series) > 12 || len(n.Phases) > 12 {
		return bad("invalid_geometry_or_series")
	}
	mode := n.Curve
	if mode == "" {
		mode = "catmull"
		if r.source != nil && isExpandedLibrary(r.source.Revision) {
			mode = "monotone"
		}
	}
	if n.LineBasis != "" && n.LineBasis != "above_stack" && n.LineBasis != "independent" {
		return bad("line_basis")
	}
	if mode != "monotone" && mode != "catmull" {
		return bad("curve_mode")
	}
	smooth := 0.
	if isExpandedLibrary(r.source.Revision) {
		smooth = .07
	}
	if n.Smooth != nil {
		smooth = *n.Smooth
	}
	if !finite(smooth) || smooth < 0 || smooth > 1 {
		return bad("smooth_range")
	}
	dense := smooth > 0 && mode != "catmull"
	pathFor := func(points []curvePoint, move bool) []pptx.ShapePoint {
		if mode == "monotone" {
			return teamCurveMonotonePath(points, move)
		}
		return teamCurvePath(points, tensionForCurve(n), move)
	}
	for _, ph := range n.Phases {
		if len(ph.Label) > 4096 || len(ph.Sub) > 16384 {
			return bad("phase_text")
		}
	}
	np := len(n.Series[0].Values)
	if np < 2 || np > 512 {
		return bad("sample_limit")
	}
	at := n.At
	if len(at) == 0 {
		at = make([]float64, np)
		for i := range at {
			at[i] = float64(i) / float64(np-1)
		}
	}
	if len(at) != np {
		return bad("sample_count")
	}
	for i, v := range at {
		if !finite(v) || v < 0 || v > 1 || i > 0 && v <= at[i-1] {
			return bad("sample_positions")
		}
	}
	if len(n.PointLabels) > 0 && len(n.PointLabels) != np {
		return bad("point_label_count")
	}
	if len(n.Unit) > 128 || len(n.TimeUnit) > 128 || len(n.Source) > 1024 {
		return bad("scale_labels")
	}
	for _, label := range n.PointLabels {
		if len(label) > 128 {
			return bad("point_labels")
		}
	}
	// Optional semantic declarations reserve a native bottom caption, preserving the
	// exact legacy visual when no declaration is authored. Facts are not inferred.
	allocation := Rect{n.X, n.Y, n.W, n.H}
	heading := &scenePlan{}
	if n.Unit != "" || n.TimeUnit != "" || n.Source != "" {
		if n.Unit == "" || n.TimeUnit == "" || n.Source == "" {
			return bad("scale_declaration")
		}
		st, e := r.sceneStyle("small")
		if e != nil {
			return nil, true, e
		}
		label := n.Unit + " · " + n.TimeUnit + " · " + n.Source
		if e = r.sceneText(heading, id+".scale", label, st, Rect{n.X, n.Y + n.H - 18, n.W, 18}, ctx.Surface, "primary", "left"); e != nil {
			return nil, true, e
		}
		n.H -= 22
		if n.H <= 0 {
			return bad("scale_capacity")
		}
	}

	phaseH, tension := 54., .5
	if n.PhaseH != nil {
		phaseH = *n.PhaseH
	}
	if n.Tension != nil {
		tension = *n.Tension
	}
	if !finite(phaseH) || phaseH < 0 || phaseH >= n.H || !finite(tension) || tension < 0 || tension > 1 {
		return bad("phase_or_tension")
	}
	plotH := n.H - phaseH
	totals := make([]float64, np)
	for _, se := range n.Series {
		if len(se.Name) > 4096 {
			return bad("series_text")
		}
		if len(se.Values) != np || se.Style != "" && se.Style != "area" && se.Style != "line" {
			return bad("series_shape")
		}
		for i, v := range se.Values {
			if !finite(v) || v < 0 || v > 1e6 {
				return bad("series_value")
			}
			totals[i] += v
		}
	}
	maxTotal := 0.
	for _, v := range totals {
		maxTotal = math.Max(maxTotal, v)
	}
	maxT := maxTotal * 1.15
	if n.Max != nil {
		maxT = *n.Max
	}
	if !finite(maxT) || maxT <= 0 {
		return bad("max")
	}
	// Browser maxima include independent line values in the auto-scale. Bounds
	// validation below uses the actual cumulative bands rather than that sum.
	p := &scenePlan{ID: id, Bounds: allocation}
	p.Items = append(p.Items, heading.Items...)
	p.Warnings = append(p.Warnings, "Adapter resolution wmds.teamcurve-smoothing.v1: "+mode+"; frozen v1/v2/v3 defaults remain catmull, explicit monotone or expanded v4/v5 libraries use shape-preserving cubics.")
	base := make([]float64, np)
	var denseBase []float64
	if dense {
		denseBase = make([]float64, 241)
		p.Warnings = append(p.Warnings, "Adapter resolution wmds.teamcurve-gaussian.v1: 241 cumulative samples with edge-clamped Gaussian smoothing; native editable polyline bands.")
		pathFor = teamCurvePolylinePath
	}
	seriesKeys, err := primitiveArrayKeys(ctx, "/series", len(n.Series))
	if err != nil {
		return nil, true, err
	}
	var phaseKeys []string
	if len(n.Phases) > 0 {
		phaseKeys, err = primitiveArrayKeys(ctx, "/phases", len(n.Phases))
		if err != nil {
			return nil, true, err
		}
	} else if _, supplied := ctx.Keys[ctx.Path+"/phases"]; supplied {
		if _, err = primitiveArrayKeys(ctx, "/phases", 0); err != nil {
			return nil, true, err
		}
	}
	labels := &scenePlan{}
	for si, se := range n.Series {
		key := seriesKeys[si]
		sid := id + ".series." + key
		fill := se.Fill
		if fill == "" {
			fill = []string{"#CED7E6", "#070154", "#0047FF"}[si%3]
		}
		color, err := r.sceneColor(ctx.Surface, fill)
		if err != nil {
			return nil, true, err
		}
		up, dn := make([]curvePoint, np), make([]curvePoint, np)
		for i, v := range se.Values {
			top := base[i] + v
			bottom := base[i]
			if se.Style == "line" && n.LineBasis == "independent" {
				top = v
				bottom = 0
			}
			if top > maxT && (top-maxT)/maxT > 1e-10 {
				return bad("value_exceeds_max")
			}
			up[i] = curvePoint{at[i] * n.W, plotH - top/maxT*plotH}
			dn[np-1-i] = curvePoint{at[i] * n.W, plotH - bottom/maxT*plotH}
			if !intakeFinite(up[i].x, up[i].y, dn[np-1-i].x, dn[np-1-i].y) {
				return bad("nonfinite_derived_geometry")
			}
		}
		labelUp, labelDn := up, dn
		if dense {
			var hi, lo []float64
			if se.Style == "line" && n.LineBasis == "independent" {
				hi, lo, _ = teamCurveGaussianEdges(at, se.Values, make([]float64, len(denseBase)), smooth, true)
			} else {
				hi, lo, denseBase = teamCurveGaussianEdges(at, se.Values, denseBase, smooth, se.Style == "line")
			}
			up, dn = make([]curvePoint, 241), make([]curvePoint, 241)
			for i := range up {
				u := at[0] + (at[np-1]-at[0])*float64(i)/240
				up[i] = curvePoint{u * n.W, plotH - hi[i]/maxT*plotH}
				dn[240-i] = curvePoint{u * n.W, plotH - lo[i]/maxT*plotH}
			}
			labelUp, labelDn = make([]curvePoint, np), make([]curvePoint, np)
			for i, u := range at {
				j := int(math.Round((u - at[0]) / (at[np-1] - at[0]) * 240))
				labelUp[i], labelDn[np-1-i] = up[j], dn[240-j]
			}
		}
		points := pathFor(up, true)
		if !teamCurveFinitePath(points) {
			return bad("nonfinite_derived_geometry")
		}
		props := pptx.ShapeProps{PositionProps: pos(Rect{n.X, n.Y, n.W, plotH}), ObjectNameProps: pptx.ObjectNameProps{ObjectName: sid}, Points: points}
		if se.Style == "line" {
			props.Fill = &pptx.ShapeFillProps{Type: "none"}
			props.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: 2.5}
			if se.Dashed {
				props.Line.DashType = "dash"
			}
		} else {
			bottom := pathFor(dn, false)
			if !teamCurveFinitePath(bottom) {
				return bad("nonfinite_derived_geometry")
			}
			props.Points = append(props.Points, bottom...)
			props.Points = append(props.Points, pptx.ShapePoint{Close: true})
			props.Fill = &pptx.ShapeFillProps{Color: color}
			if se.Hatch {
				props.Fill.Transparency = 10
			}
			props.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}
		}
		geometry := teamCurveNativeBounds(points)
		if se.Style != "line" {
			geometry = diagramUnion(geometry, teamCurveNativeBounds(pathFor(dn, false)))
		}
		geometry.X += n.X
		geometry.Y += n.Y
		// Preserve the selected source curve and record its true ink extent.
		// Frozen Catmull overshoot remains visible to allocation checks.
		if se.Style == "line" {
			geometry = Rect{geometry.X - 1.25, geometry.Y - 1.25, geometry.W + 2.5, geometry.H + 2.5}
		}
		geometryKind := "cubic-stacked-band"
		if dense {
			geometryKind = "gaussian-stacked-band"
		}
		p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeCustGeom, Props: props, Record: ShapeRecord{ID: sid, Rect: geometry, Color: color, Geometry: geometryKind}}})
		labelIndex := 0
		show := true
		if len(se.LabelAt) > 0 {
			if string(se.LabelAt) == "false" {
				show = false
			} else if err = json.Unmarshal(se.LabelAt, &labelIndex); err != nil {
				return bad("label_index")
			}
		}
		if labelIndex < 0 || labelIndex >= np {
			return bad("label_index")
		}
		if show && se.Name != "" {
			x := n.X + at[labelIndex]*n.W - 60
			if labelIndex == 0 {
				x = n.X + at[0]*n.W + 9
			}
			labelW := math.Min(180, n.W)
			boundedX := math.Max(n.X, math.Min(x, n.X+n.W-labelW))
			if boundedX != x {
				p.Warnings = append(p.Warnings, "Teamcurve direct label shifted horizontally to fit the component: "+sid)
			}
			x = boundedX
			y := n.Y + (labelUp[labelIndex].y+labelDn[np-1-labelIndex].y)/2 - 8
			if se.Style == "line" && n.LineBasis == "independent" {
				y = n.Y + labelUp[labelIndex].y - 8
			}
			ink, surface := "primary", "light"
			if color == "070154" || color == "0047FF" || color == "05013F" {
				surface = "inverse"
			}
			if se.Style == "line" {
				ink = "#" + color
				surface = ctx.Surface
			}
			st, err := r.sceneStyle("label")
			if err != nil {
				return nil, true, err
			}
			st.Weight = 600
			if isExpandedLibrary(r.source.Revision) && n.W <= 414 && se.Style != "line" {
				// A fixed180pt box clamped at the component edge can shift a
				// short label into the adjacent band. Center its measured native
				// width on the authored sample instead.
				layout, e := r.measureText(se.Name, st, labelW)
				if e != nil {
					return nil, true, e
				}
				if len(layout.Lines) == 1 {
					labelW = math.Min(labelW, sequenceInlineWidth(layout.Lines[0].Advance, layout.Style))
					x = math.Max(n.X, math.Min(n.X+at[labelIndex]*n.W-labelW/2, n.X+n.W-labelW))
				}
			}
			if err = r.sceneText(labels, sid+".label", se.Name, st, Rect{x, y, labelW, 0}, surface, ink, "left"); err != nil {
				return nil, true, err
			}
		}
		if se.Style != "line" {
			for i, v := range se.Values {
				base[i] += v
			}
		}
	}
	lastAt := -1.
	for i, ph := range n.Phases {
		a := float64(i) / float64(len(n.Phases))
		if ph.At != nil {
			a = *ph.At
		}
		if !finite(a) || a < 0 || a >= 1 || a <= lastAt {
			return bad("phase_positions")
		}
		lastAt = a
		key := phaseKeys[i]
		pid := id + ".phases." + key
		if i > 0 {
			b := Rect{n.X + a*n.W, n.Y, 0, n.H}
			props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: pid + ".rule"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: "97A4BA"}, Width: .75}}
			p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeLine, Props: props, Record: ShapeRecord{ID: pid + ".rule", Rect: b, Color: "97A4BA"}}})
		}
		if n.PhaseLabels != nil && !*n.PhaseLabels {
			continue
		}
		end := 1.
		if i+1 < len(n.Phases) {
			end = float64(i+1) / float64(len(n.Phases))
			if n.Phases[i+1].At != nil {
				end = *n.Phases[i+1].At
			}
		}
		if !finite(end) || end <= a || end > 1 {
			return bad("phase_positions")
		}
		box := Rect{n.X + a*n.W + 9, n.Y + plotH + 9, (end-a)*n.W - 18, phaseH - 9}
		if box.W <= 0 || box.H <= 0 {
			return bad("phase_label_capacity")
		}
		st, err := r.sceneStyle("eyebrow")
		if err != nil {
			return nil, true, err
		}
		before := len(p.Items)
		if err = r.sceneText(p, pid+".label", ph.Label, st, box, ctx.Surface, "emphasis", "left"); err != nil {
			return nil, true, err
		}
		if ph.Sub != "" {
			height := 0.
			if len(p.Items) > before {
				height = p.Items[len(p.Items)-1].Text.Layout.AllocationHeight
			}
			st, err = r.sceneStyle("small")
			if err != nil {
				return nil, true, err
			}
			box.Y += height + 2
			box.H -= height + 2
			if err = r.sceneText(p, pid+".sub", ph.Sub, st, box, ctx.Surface, "secondary", "left"); err != nil {
				return nil, true, err
			}
		}
	}
	// Browser series labels paint last, above every band and phase divider.
	p.Items = append(p.Items, labels.Items...)
	return diagramFinish(p, "teamcurve"), true, nil
}

func tensionForCurve(n teamCurveSource) float64 {
	if n.Tension != nil {
		return *n.Tension
	}
	return .5
}
