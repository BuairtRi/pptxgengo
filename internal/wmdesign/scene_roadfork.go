package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

const RoadForkContract = "pptxgengo.wmds-roadfork.v1"

// Source nowrap labels reserve native rounding clearance inside their existing gutter.
const roadForkNoWrapClearance = 2.0

func roadForkNumber(raw json.RawMessage, fallback string) (string, error) {
	var n float64
	if len(raw) != 0 && json.Unmarshal(raw, &n) == nil && n == 0 {
		return fallback, nil
	}
	return maturityNumber(raw, fallback)
}

type roadForkMilestone struct {
	Label     string          `json:"label,omitempty"`
	Date      string          `json:"date,omitempty"`
	Text      string          `json:"text,omitempty"`
	Number    json.RawMessage `json:"n,omitempty"`
	Here      bool            `json:"here,omitempty"`
	HereLabel string          `json:"hereLabel,omitempty"`
}
type roadForkSpec struct {
	Type    string              `json:"type"`
	ID      string              `json:"id,omitempty"`
	X       float64             `json:"x"`
	Y       float64             `json:"y"`
	W       float64             `json:"w"`
	H       float64             `json:"h"`
	CanvasH float64             `json:"_h,omitempty"`
	Mode    string              `json:"mode,omitempty"`
	Chosen  *int                `json:"chosen,omitempty"`
	Small   *bool               `json:"small,omitempty"`
	RoadW   *float64            `json:"roadW,omitempty"`
	ForkW   *float64            `json:"forkW,omitempty"`
	Trunk   []roadForkMilestone `json:"trunk,omitempty"`
	Fork    struct {
		Label string   `json:"label,omitempty"`
		Text  string   `json:"text,omitempty"`
		Tag   string   `json:"tag,omitempty"`
		At    *float64 `json:"at,omitempty"`
	} `json:"fork,omitempty"`
	Branches []struct {
		Title      string              `json:"title"`
		Text       string              `json:"text,omitempty"`
		Milestones []roadForkMilestone `json:"milestones"`
	} `json:"branches"`
}

func (r *renderer) planRoadForkScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "roadfork" {
		return nil, false, nil
	}
	if r.source.Revision != LibraryRevisionV10 && r.source.Revision != LibraryRevisionV11 {
		return nil, true, fmt.Errorf("scene.roadfork_requires_v10")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, true, err
	}
	if err := round12RequiredNumbers(fields, "x", "y", "w", "h"); err != nil {
		return nil, true, err
	}
	var n roadForkSpec
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	p, err := r.roadFork(id, n, ctx)
	return p, true, err
}

// Native cubics retain the source SVG control points. Centerline dashes are
// editable subpaths, with source point lengths independent of stroke width.
func (r *renderer) roadForkPath(p *scenePlan, id string, start, c1, c2, end, last curvePoint, color string, width, on, off float64, cubic bool) error {
	points := []curvePoint{start, c1, c2, end, last}
	lo, hi := start, start
	for _, q := range points {
		lo.x = math.Min(lo.x, q.x)
		lo.y = math.Min(lo.y, q.y)
		hi.x = math.Max(hi.x, q.x)
		hi.y = math.Max(hi.y, q.y)
	}
	b := Rect{lo.x, lo.y, hi.x - lo.x, hi.y - lo.y}
	coord := func(v float64) pptx.Coord { return pptx.Inches(v / 72) }
	point := func(q curvePoint, move bool) pptx.ShapePoint {
		return pptx.ShapePoint{X: coord(q.x - b.X), Y: coord(q.y - b.Y), MoveTo: ptrSceneBool(move)}
	}
	curve := func(q, a, z curvePoint) pptx.ShapePoint {
		x := point(q, false)
		x.Curve = &pptx.ShapeCurve{Type: "cubic", X1: coord(a.x - b.X), Y1: coord(a.y - b.Y), X2: coord(z.x - b.X), Y2: coord(z.y - b.Y)}
		return x
	}
	props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: width, LineJoin: "round"}}
	if on == 0 {
		props.Points = append(props.Points, point(start, true))
		if cubic {
			props.Points = append(props.Points, curve(end, c1, c2))
		} else {
			props.Points = append(props.Points, point(end, false))
		}
		if last != end {
			props.Points = append(props.Points, point(last, false))
		}
	} else {
		// Arclength is bounded using 512 segments; each dash retains an exact
		// Bezier subcurve rather than approximating its visible shape with a raster.
		sample := func(t float64) curvePoint {
			u := 1 - t
			return curvePoint{u*u*u*start.x + 3*u*u*t*c1.x + 3*u*t*t*c2.x + t*t*t*end.x, u*u*u*start.y + 3*u*u*t*c1.y + 3*u*t*t*c2.y + t*t*t*end.y}
		}
		derivative := func(t float64) curvePoint {
			u := 1 - t
			return curvePoint{3*u*u*(c1.x-start.x) + 6*u*t*(c2.x-c1.x) + 3*t*t*(end.x-c2.x), 3*u*u*(c1.y-start.y) + 6*u*t*(c2.y-c1.y) + 3*t*t*(end.y-c2.y)}
		}
		lengths := make([]float64, 513)
		for i := 1; i < len(lengths); i++ {
			a, z := sample(float64(i-1)/512), sample(float64(i)/512)
			lengths[i] = lengths[i-1] + math.Hypot(z.x-a.x, z.y-a.y)
		}
		curveLen := 0.
		if cubic {
			curveLen = lengths[512]
		}
		lineStart := end
		if !cubic {
			lineStart = start
		}
		lineLen := math.Hypot(last.x-lineStart.x, last.y-lineStart.y)
		total := curveLen + lineLen
		parameter := func(distance float64) float64 {
			i := 1
			for i < 512 && lengths[i] < distance {
				i++
			}
			den := lengths[i] - lengths[i-1]
			if den == 0 {
				return float64(i) / 512
			}
			return (float64(i-1) + (distance-lengths[i-1])/den) / 512
		}
		linePoint := func(d float64) curvePoint {
			t := (d - curveLen) / lineLen
			return curvePoint{lineStart.x + (last.x-lineStart.x)*t, lineStart.y + (last.y-lineStart.y)*t}
		}
		for d := 0.; d < total; d += on + off {
			stop := math.Min(d+on, total)
			if d < curveLen {
				a, z := parameter(d), parameter(math.Min(stop, curveLen))
				qa, qz := sample(a), sample(z)
				da, dz := derivative(a), derivative(z)
				span := (z - a) / 3
				props.Points = append(props.Points, point(qa, true), curve(qz, curvePoint{qa.x + da.x*span, qa.y + da.y*span}, curvePoint{qz.x - dz.x*span, qz.y - dz.y*span}))
				if stop > curveLen && lineLen > 0 {
					props.Points = append(props.Points, point(linePoint(stop), false))
				}
			} else if lineLen > 0 {
				props.Points = append(props.Points, point(linePoint(d), true), point(linePoint(stop), false))
			}
		}
	}
	extent := Rect{b.X - width/2, b.Y - width/2, b.W + width, b.H + width}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeCustGeom, Props: props, Record: ShapeRecord{ID: id, Rect: extent, Color: color, Geometry: "native-roadfork-path"}}})
	p.Bounds = diagramUnion(p.Bounds, extent)
	if on == 0 {
		for i, q := range []curvePoint{start, last} {
			if err := r.diagramShape(p, fmt.Sprintf("%s.cap-%d", id, i+1), Rect{q.x - width/2, q.y - width/2, width, width}, pptx.ShapeTypeEllipse, color, "", 0, "", nil); err != nil {
				return err
			}
		}
	}
	return nil
}

type roadForkTextPart struct {
	text, token, ink string
	weight           int
	leading          float64
}

func (r *renderer) roadForkStack(p *scenePlan, id string, parts []roadForkTextPart, b Rect, surface, align string, above bool, gap float64) (float64, error) {
	child := &scenePlan{}
	y := 0.
	for i, part := range parts {
		if part.text == "" {
			continue
		}
		if y > 0 {
			y += gap
		}
		st, err := r.sceneStyle(part.token)
		if err != nil {
			return 0, err
		}
		if part.weight > 0 {
			st.Weight = part.weight
		}
		if part.leading > 0 {
			st.Leading = st.Size * part.leading
		}
		l, err := r.measureText(part.text, st, b.W)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", id, err)
		}
		fg, err := r.sceneColor(surface, part.ink)
		if err != nil {
			return 0, err
		}
		h := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
		tr := TextRecord{ID: fmt.Sprintf("%s.part-%d", id, i+1), Rect: Rect{b.X, y, b.W, h}, Color: fg, Align: align, Layout: l}
		child.Items = append(child.Items, sceneItem{Text: &tr})
		y += l.AllocationHeight
	}
	shift := b.Y
	if above {
		shift -= y
	}
	for _, it := range child.Items {
		it.Text.Rect.Y += shift
		p.Items = append(p.Items, it)
		p.Bounds = diagramUnion(p.Bounds, it.Text.Rect)
	}
	return y, nil
}

func (r *renderer) roadForkPin(p *scenePlan, id string, q curvePoint, number string, m roadForkMilestone, grey bool, surface string) error {
	fill, ring := "FFFFFF", "070154"
	if grey {
		ring = "97A4BA"
	}
	if m.Here {
		fill = "F900D3"
		ring = "070154"
	}
	if err := r.diagramShape(p, id+".pin", Rect{q.x - 11.25, q.y - 11.25, 22.5, 22.5}, pptx.ShapeTypeEllipse, fill, ring, 1.5, "solid", nil); err != nil {
		return err
	}
	p.Items[len(p.Items)-1].Shape.Record.Rect = Rect{q.x - 12, q.y - 12, 24, 24}
	p.Bounds = diagramUnion(p.Bounds, Rect{q.x - 12, q.y - 12, 24, 24})
	st, err := r.sceneStyle("label")
	if err != nil {
		return err
	}
	st.Family, st.Size, st.Leading, st.Weight, st.TrackingPt, st.Case = "IBM Plex Mono", 10, 10, 600, 0, ""
	ink := ring
	if grey && r.hasDensityContrastRules() {
		ink = "070154"
	}
	if err = r.intakeVennSingleText(p, id+".number", number, st, Rect{q.x - 12, q.y - 12, 24, 24}, surface, "#"+ink, "center", true); err != nil {
		return err
	}
	if m.Here {
		label := m.HereLabel
		if label == "" {
			label = "You are here"
		}
		st, err = r.sceneStyle("label")
		if err != nil {
			return err
		}
		st.Weight = 600
		l, err := r.measureText(label, st, 8191)
		if err != nil {
			return err
		}
		if len(l.Lines) != 1 {
			return fmt.Errorf("scene.roadfork_here_label")
		}
		w := l.Lines[0].Advance + 10 + roadForkNoWrapClearance
		if w > 180 {
			return fmt.Errorf("scene.roadfork_here_label_width")
		}
		h := st.Leading + 2
		if err = r.diagramShape(p, id+".here.tag", Rect{q.x - w/2, q.y - 34, w, h}, pptx.ShapeTypeRect, "070154", "", 0, "", nil); err != nil {
			return err
		}
		textH := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
		if err = r.intakeVennSingleText(p, id+".here.label", label, st, Rect{q.x - w/2 + 5, q.y - 33, w - 10 + .01, textH}, surface, "#FFFFFF", "center", false); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) roadFork(id string, n roadForkSpec, ctx SceneContext) (*scenePlan, error) {
	bad := func(reason string) (*scenePlan, error) { return nil, fmt.Errorf("scene.roadfork_%s: %s", reason, id) }
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH) || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 || n.W <= 100 || n.W > 1920 || n.H <= 80 || n.H > 1080 || len(n.Trunk) > 3 || len(n.Branches) < 2 || len(n.Branches) > 3 {
		return bad("geometry_or_count")
	}
	budget := 0
	checkCopy := func(text string, limit int) bool {
		count := utf8.RuneCountInString(text)
		budget += count
		return utf8.ValidString(text) && count <= limit && budget <= 65536
	}
	checkMilestone := func(m roadForkMilestone) bool {
		return checkCopy(m.Label, 2048) && checkCopy(m.Date, 2048) && checkCopy(m.Text, 8192) && checkCopy(m.HereLabel, 256)
	}
	if !checkCopy(n.Fork.Label, 2048) || !checkCopy(n.Fork.Text, 8192) || !checkCopy(n.Fork.Tag, 2048) {
		return bad("copy_budget")
	}
	for _, m := range n.Trunk {
		if !checkMilestone(m) {
			return bad("copy_budget")
		}
	}
	for _, b := range n.Branches {
		if !checkCopy(b.Title, 2048) || !checkCopy(b.Text, 8192) || len(b.Milestones) < 1 || len(b.Milestones) > 6 {
			return bad("branch_content_or_count")
		}
		for _, m := range b.Milestones {
			if !checkMilestone(m) {
				return bad("copy_budget")
			}
		}
	}
	if n.Mode != "" && n.Mode != "parallel" && n.Mode != "decision" {
		return bad("mode")
	}
	decision := n.Mode == "decision"
	if n.Chosen != nil && (!decision || *n.Chosen < 0 || *n.Chosen >= len(n.Branches)) {
		return bad("chosen")
	}
	at := .1
	if len(n.Trunk) > 0 {
		at = .36
	}
	if n.Fork.At != nil {
		at = *n.Fork.At
	}
	rw, fw := 18., 162.
	if n.RoadW != nil {
		rw = *n.RoadW
	}
	if n.ForkW != nil {
		fw = *n.ForkW
	}
	if !intakeFinite(at, rw, fw) || at <= 0 || at >= 1 || rw <= 0 || rw > 120 || fw <= 0 || fw > n.W {
		return bad("options")
	}
	small := len(n.Branches) >= 3
	if n.Small != nil {
		small = *n.Small
	}
	labelToken := "body"
	if small {
		labelToken = "small"
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	if _, err := r.sceneColor(surface, "primary"); err != nil {
		return nil, err
	}
	tk, err := primitiveArrayKeys(ctx, "/trunk", len(n.Trunk))
	if err != nil {
		return nil, err
	}
	bk, err := primitiveArrayKeys(ctx, "/branches", len(n.Branches))
	if err != nil {
		return nil, err
	}
	x0, x1, fx, cy := n.X+14, n.X+n.W-14, n.X+n.W*at, n.Y+n.H/2
	curveW := math.Min(108, (x1-fx)*.22)
	bx := fx + curveW
	laneH := n.H / float64(len(n.Branches))
	laneY := func(i int) float64 { return n.Y + laneH*(float64(i)+.42) }
	if fx <= x0 || bx >= x1 || len(n.Trunk) > 0 && (fx-x0-30)/float64(len(n.Trunk)) <= 8 {
		return bad("trunk_allocation")
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, n.H}}
	if err = r.roadForkPath(p, id+".trunk.road", curvePoint{x0, cy}, curvePoint{x0, cy}, curvePoint{fx, cy}, curvePoint{fx, cy}, curvePoint{fx, cy}, "070154", rw, 0, 0, false); err != nil {
		return nil, err
	}
	if err = r.roadForkPath(p, id+".trunk.centerline", curvePoint{x0, cy}, curvePoint{x0, cy}, curvePoint{fx, cy}, curvePoint{fx, cy}, curvePoint{fx, cy}, "FFFFFF", 1.5, 7, 7, false); err != nil {
		return nil, err
	}
	order := []int{}
	for i := range n.Branches {
		if n.Chosen == nil || i != *n.Chosen {
			order = append(order, i)
		}
	}
	if n.Chosen != nil {
		order = append(order, *n.Chosen)
	}
	for _, i := range order {
		yy := laneY(i)
		color := "070154"
		if decision {
			if n.Chosen == nil {
				color = "50658E"
			} else if *n.Chosen != i {
				color = "CED7E6"
			}
		}
		a, z := curvePoint{fx, cy}, curvePoint{bx, yy}
		c1, c2 := curvePoint{fx + curveW*.55, cy}, curvePoint{fx + curveW*.45, yy}
		last := curvePoint{x1, yy}
		part := id + ".branches." + bk[i]
		if err = r.roadForkPath(p, part+".road", a, c1, c2, z, last, color, rw, 0, 0, true); err != nil {
			return nil, err
		}
		on, off := 7., 7.
		if decision && (n.Chosen == nil || *n.Chosen != i) {
			on, off = 3, 5
		}
		if err = r.roadForkPath(p, part+".centerline", a, c1, c2, z, last, "FFFFFF", 1.5, on, off, true); err != nil {
			return nil, err
		}
	}
	number := 0
	tw := (fx - x0 - 30) / math.Max(1, float64(len(n.Trunk)))
	label := func(part string, q curvePoint, w float64, m roadForkMilestone, above, grey bool) error {
		dateInk, labelInk, textInk := "emphasis", "display", "primary"
		if grey {
			dateInk, labelInk, textInk = "secondary", "secondary", "secondary"
		}
		py := q.y + 18
		if above {
			py = q.y - 18
		}
		_, e := r.roadForkStack(p, part, []roadForkTextPart{{m.Date, "label", dateInk, 600, 0}, {m.Label, labelToken, labelInk, 600, 1.2}, {m.Text, "small", textInk, 0, 1.25}}, Rect{q.x - w/2, py, w, 0}, surface, "center", above, 2)
		return e
	}
	for i, m := range n.Trunk {
		number++
		q := curvePoint{x0 + 18 + tw*(float64(i)+.5), cy}
		num, e := roadForkNumber(m.Number, strconv.Itoa(number))
		if e != nil {
			return nil, e
		}
		part := id + ".trunk." + tk[i]
		if err = r.roadForkPin(p, part, q, num, m, false, surface); err != nil {
			return nil, err
		}
		if err = label(part+".label", q, tw-8, m, i%2 == 0, false); err != nil {
			return nil, err
		}
	}
	if decision {
		// CSS rotates a 26pt square, so its true diamond has 26*sqrt(2) extents.
		// Inset the centered native stroke to preserve that outer extent.
		size := 24.5 * math.Sqrt2
		if err = r.diagramShape(p, id+".fork.diamond", Rect{fx - size/2, cy - size/2, size, size}, pptx.ShapeTypeDiamond, "F900D3", "070154", 1.5, "solid", nil); err != nil {
			return nil, err
		}
		p.Items[len(p.Items)-1].Shape.Props.Line.LineJoin = "miter"
		p.Items[len(p.Items)-1].Shape.Record.Rect = Rect{fx - 13*math.Sqrt2, cy - 13*math.Sqrt2, 26 * math.Sqrt2, 26 * math.Sqrt2}
		p.Bounds = diagramUnion(p.Bounds, p.Items[len(p.Items)-1].Shape.Record.Rect)
		if n.Chosen == nil {
			st, e := r.sceneStyle("label")
			if e != nil {
				return nil, e
			}
			st.Family, st.Size, st.Leading, st.Weight, st.TrackingPt, st.Case = "IBM Plex Mono", 12, 12, 700, 0, ""
			if err = r.intakeVennSingleText(p, id+".fork.question", "?", st, Rect{fx - 13, cy - 13, 26, 26}, surface, "#070154", "center", true); err != nil {
				return nil, err
			}
		} else {
			r.sceneDataCheck(p, id+".fork.chosen", Rect{fx - 6, cy - 6, 12, 12}, "070154")
		}
	} else {
		if err = r.diagramShape(p, id+".fork.split", Rect{fx - 11, cy - 11, 22, 22}, pptx.ShapeTypeRect, "070154", "070154", 3, "solid", nil); err != nil {
			return nil, err
		}
		p.Items[len(p.Items)-1].Shape.Record.Rect = Rect{fx - 12.5, cy - 12.5, 25, 25}
		p.Bounds = diagramUnion(p.Bounds, p.Items[len(p.Items)-1].Shape.Record.Rect)
		if err = r.diagramShape(p, id+".fork.split.inset", Rect{fx - 9.5, cy - 9.5, 19, 19}, pptx.ShapeTypeRect, "070154", "FFFFFF", 3, "solid", nil); err != nil {
			return nil, err
		}
	}
	if n.Fork.Label != "" || n.Fork.Text != "" {
		tag := n.Fork.Tag
		if tag == "" {
			tag = "Run in parallel"
			if decision {
				tag = "Decision"
				if n.Chosen != nil {
					tag = "Decided"
				}
			}
		}
		if _, err = r.roadForkStack(p, id+".fork.label", []roadForkTextPart{{tag, "label", "emphasis", 600, 0}, {n.Fork.Label, "body", "display", 600, 1.2}, {n.Fork.Text, "small", "secondary", 0, 0}}, Rect{fx - fw/2, n.Y, fw, 0}, surface, "center", false, 2); err != nil {
			return nil, err
		}
	}
	for i, b := range n.Branches {
		if strings.TrimSpace(b.Title) == "" || len(b.Milestones) < 1 || len(b.Milestones) > 6 {
			return bad("branch_content_or_count")
		}
		keys, e := primitiveArrayKeys(ctx, "/branches/"+strconv.Itoa(i)+"/milestones", len(b.Milestones))
		if e != nil {
			return nil, e
		}
		yy := laneY(i)
		grey := decision && n.Chosen != nil && *n.Chosen != i
		part := id + ".branches." + bk[i]
		tagInk, titleInk := "emphasis", "display"
		if grey {
			tagInk, titleInk = "secondary", "secondary"
		}
		tag := "Track "
		if decision {
			tag = "Option "
		}
		tag += string(rune('A' + i))
		st, e := r.sceneStyle("label")
		if e != nil {
			return nil, e
		}
		st.Weight = 600
		l, e := r.measureText(tag, st, 8191)
		if e != nil {
			return nil, e
		}
		tagW := l.Lines[0].Advance
		titleW := math.Min(420, x1-bx) - tagW - 9
		if titleW <= 0 {
			return bad("branch_title_allocation")
		}
		child := &scenePlan{}
		titleH, e := r.roadForkStack(child, part+".title", []roadForkTextPart{{b.Title, "subhead", titleInk, 0, 1.15}}, Rect{bx + tagW + 9, 0, titleW, 0}, surface, "left", false, 0)
		if e != nil {
			return nil, e
		}
		titleStyle, e := r.sceneStyle("subhead")
		if e != nil {
			return nil, e
		}
		titleStyle.Leading = titleStyle.Size * 1.15
		// Both spans share the browser row's baseline.
		tagLayout, e := r.measureText(tag, st, tagW+roadForkNoWrapClearance)
		if e != nil {
			return nil, e
		}
		baseline := child.Items[0].Text.Layout.Lines[0].Baseline
		tagY := baseline - tagLayout.Lines[0].Baseline
		color, e := r.sceneColor(surface, tagInk)
		if e != nil {
			return nil, e
		}
		tagHeight := math.Max(tagLayout.AllocationHeight, tagLayout.OccupiedTop+tagLayout.EstimatedOccupiedHeight)
		tr := TextRecord{ID: part + ".tag", Rect: Rect{bx, tagY, tagW + roadForkNoWrapClearance, tagHeight}, Color: color, Align: "left", Layout: tagLayout}
		child.Items = append(child.Items, sceneItem{Text: &tr})
		flowH := math.Max(titleH, tagY+tagLayout.AllocationHeight)
		if b.Text != "" {
			h, e := r.roadForkStack(child, part+".support", []roadForkTextPart{{b.Text, "small", "secondary", 0, 1.25}}, Rect{bx, flowH + 1, math.Min(420, x1-bx), 0}, surface, "left", false, 0)
			if e != nil {
				return nil, e
			}
			flowH += 1 + h
		}
		shift := yy - 16 - flowH
		for _, it := range child.Items {
			it.Text.Rect.Y += shift
			p.Items = append(p.Items, it)
			p.Bounds = diagramUnion(p.Bounds, it.Text.Rect)
		}
		segW := (x1 - bx - 12) / float64(len(b.Milestones))
		if segW <= 8 {
			return bad("milestone_allocation")
		}
		for j, m := range b.Milestones {
			number++
			q := curvePoint{bx + 12 + segW*(float64(j)+.5), yy}
			num, e := roadForkNumber(m.Number, string(rune('A'+i))+strconv.Itoa(j+1))
			if e != nil {
				return nil, e
			}
			mp := part + ".milestones." + keys[j]
			if err = r.roadForkPin(p, mp, q, num, m, grey, surface); err != nil {
				return nil, err
			}
			if err = label(mp+".label", q, segW-8, m, false, grey); err != nil {
				return nil, err
			}
		}
	}
	p.Warnings = append(p.Warnings, RoadForkContract+"; native editable forked cubic roads; source specimen review pending")
	sceneDataGroup(p, id, "diagram.roadfork", 0, p.Bounds)
	return p, nil
}
