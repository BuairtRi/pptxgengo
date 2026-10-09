package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"math"
	"strings"
)

type ProcessLane struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}
type ProcessStep struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Lane    string `json:"lane"`
	Column  int    `json:"column"`
	Kind    string `json:"kind"`
	Surface string `json:"surface,omitempty"`
}
type ProcessLink struct {
	Key           string       `json:"key"`
	From          string       `json:"from"`
	To            string       `json:"to"`
	Outcome       string       `json:"outcome,omitempty"`
	Route         [][2]float64 `json:"route,omitempty"`
	LabelPosition *[2]float64  `json:"label_position,omitempty"`
}

func (l *ProcessLink) UnmarshalJSON(raw []byte) error {
	type plain ProcessLink
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if value, present := fields["label_position"]; present {
		var coordinates []any
		if err := json.Unmarshal(value, &coordinates); err != nil || len(coordinates) != 2 {
			return fmt.Errorf("process label_position requires exactly two numeric coordinates")
		}
		for _, coordinate := range coordinates {
			if _, ok := coordinate.(float64); !ok {
				return fmt.Errorf("process label_position requires exactly two numeric coordinates")
			}
		}
	}
	if value, present := fields["route"]; present {
		var points []json.RawMessage
		if err := json.Unmarshal(value, &points); err != nil {
			return fmt.Errorf("process route requires an array of two-number waypoints")
		}
		for _, point := range points {
			var coordinates []any
			if err := json.Unmarshal(point, &coordinates); err != nil || len(coordinates) != 2 {
				return fmt.Errorf("process route waypoint requires exactly two numeric coordinates")
			}
			for _, coordinate := range coordinates {
				if _, ok := coordinate.(float64); !ok {
					return fmt.Errorf("process route waypoint requires exactly two numeric coordinates")
				}
			}
		}
	}
	var value plain
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&value); err != nil {
		return err
	}
	*l = ProcessLink(value)
	return nil
}

type ProcessSpec struct {
	Type           string        `json:"type,omitempty"`
	X              float64       `json:"x,omitempty"`
	Y              float64       `json:"y,omitempty"`
	W              float64       `json:"w,omitempty"`
	H              float64       `json:"h,omitempty"`
	Columns        int           `json:"columns"`
	LabelWidth     float64       `json:"label_width_pt"`
	HideLaneLabels bool          `json:"hide_lane_labels,omitempty"`
	StepHeight     float64       `json:"step_height_pt"`
	Lanes          []ProcessLane `json:"lanes"`
	Steps          []ProcessStep `json:"steps"`
	Links          []ProcessLink `json:"links"`
	Start          string        `json:"start,omitempty"`
	Current        string        `json:"current,omitempty"`
	End            []string      `json:"end,omitempty"`
}

func ValidateProcess(s ProcessSpec) error {
	if len(s.Lanes) < 1 || len(s.Lanes) > 12 || len(s.Steps) < 1 || len(s.Steps) > 100 || len(s.Links) > 200 || s.Columns < 1 || s.Columns > 40 || (s.LabelWidth < 24 && !s.HideLaneLabels) || s.LabelWidth < 0 || s.StepHeight < 30 || !intakeFinite(s.LabelWidth, s.StepHeight) {
		return fmt.Errorf("process invalid lane/count/layout limits")
	}
	lanes, steps, positions, links := map[string]bool{}, map[string]ProcessStep{}, map[string]bool{}, map[string]bool{}
	for _, l := range s.Lanes {
		if !validPartKey(l.Key) || lanes[l.Key] || strings.TrimSpace(l.Label) == "" {
			return fmt.Errorf("process invalid/duplicate lane %s", l.Key)
		}
		lanes[l.Key] = true
	}
	for _, v := range s.Steps {
		if !validPartKey(v.Key) || steps[v.Key].Key != "" || !lanes[v.Lane] || strings.TrimSpace(v.Label) == "" || v.Column < 0 || v.Column >= s.Columns {
			return fmt.Errorf("process invalid step %s", v.Key)
		}
		if v.Kind != "process" && v.Kind != "decision" && v.Kind != "join" && v.Kind != "start" && v.Kind != "end" {
			return fmt.Errorf("process invalid kind %s", v.Kind)
		}
		pos := fmt.Sprintf("%s/%d", v.Lane, v.Column)
		if positions[pos] {
			return fmt.Errorf("process overlapping lane/column %s", pos)
		}
		positions[pos] = true
		steps[v.Key] = v
	}
	outcomes := map[string]bool{}
	decisionCounts := map[string]int{}
	incoming := map[string]int{}
	for _, e := range s.Links {
		if !validPartKey(e.Key) || links[e.Key] || steps[e.From].Key == "" || steps[e.To].Key == "" || e.From == e.To {
			return fmt.Errorf("process invalid link %s", e.Key)
		}
		links[e.Key] = true
		if e.LabelPosition != nil && (strings.TrimSpace(e.Outcome) == "" || !intakeFinite(e.LabelPosition[0], e.LabelPosition[1])) {
			return fmt.Errorf("process link %s label_position requires a finite named outcome", e.Key)
		}
		incoming[e.To]++
		if steps[e.From].Kind == "decision" {
			if strings.TrimSpace(e.Outcome) == "" {
				return fmt.Errorf("decision link %s requires named outcome", e.Key)
			}
			k := e.From + "/" + e.Outcome
			if outcomes[k] {
				return fmt.Errorf("duplicate decision outcome %s", k)
			}
			outcomes[k] = true
			decisionCounts[e.From]++
		}
		if len(e.Route) > 64 {
			return fmt.Errorf("process route exceeds 64 waypoints")
		}
		for _, p := range e.Route {
			if !intakeFinite(p[0], p[1]) {
				return fmt.Errorf("nonfinite process route")
			}
		}
	}
	for _, step := range s.Steps {
		if step.Kind == "decision" && decisionCounts[step.Key] < 2 {
			return fmt.Errorf("decision %s requires at least two named outcomes", step.Key)
		}
		if step.Kind == "join" && incoming[step.Key] < 2 {
			return fmt.Errorf("join %s requires at least two incoming paths", step.Key)
		}
	}
	for _, k := range append(append([]string{}, s.Start, s.Current), s.End...) {
		if k != "" && steps[k].Key == "" {
			return fmt.Errorf("unknown process marker %s", k)
		}
	}
	ends := map[string]bool{}
	for _, k := range s.End {
		if k == "" || ends[k] {
			return fmt.Errorf("invalid duplicate end marker")
		}
		ends[k] = true
	}
	return nil
}
func (r *renderer) planProcessScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type != "process" {
		return nil, false, nil
	}
	var s ProcessSpec
	if e := sceneDecode(raw, &s); e != nil {
		return nil, true, e
	}
	if e := ValidateProcess(s); e != nil {
		return nil, true, e
	}
	if !intakeFinite(s.X, s.Y, s.W, s.H) || s.W <= s.LabelWidth+36 || s.H <= 0 {
		return nil, true, fmt.Errorf("process invalid allocation")
	}
	lh := s.H / float64(len(s.Lanes))
	cw := (s.W - s.LabelWidth) / float64(s.Columns)
	if lh < s.StepHeight+30 || cw < 48 {
		return nil, true, fmt.Errorf("process allocation too small for counts")
	}
	p := &scenePlan{ID: id}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	strong, e := r.sceneColor(surface, "strong")
	if e != nil {
		return nil, true, e
	}
	line, _ := r.sceneColor(surface, "line")
	boxes := map[string]Rect{}
	lanes := map[string]int{}
	for i, l := range s.Lanes {
		lanes[l.Key] = i
		y := s.Y + float64(i)*lh
		if s.HideLaneLabels {
			continue
		}
		sf := "light"
		if i%2 == 1 {
			sf = "subtle"
		}
		if e = r.sceneRect(p, id+".lanes."+l.Key+".surface", Rect{s.X + s.LabelWidth, y, s.W - s.LabelWidth, lh}, sf); e != nil {
			return nil, true, e
		}
		if e = r.diagramText(p, id+".lanes."+l.Key+".label", l.Label, "small", Rect{s.X + 6, y, s.LabelWidth - 12, lh}, surface, "display", "left", 600, true); e != nil {
			return nil, true, e
		}
		if e = r.diagramLine(p, id+".lanes."+l.Key+".rule", [2]float64{s.X + s.LabelWidth, y}, [2]float64{s.X + s.W, y}, line, .75, "solid"); e != nil {
			return nil, true, e
		}
	}
	for _, v := range s.Steps {
		w := math.Min(cw-18, 150.)
		b := Rect{s.X + s.LabelWidth + float64(v.Column)*cw + (cw-w)/2, s.Y + float64(lanes[v.Lane])*lh + (lh-s.StepHeight)/2, w, s.StepHeight}
		boxes[v.Key] = b
	}
	// Endpoints are recomputed from current semantic layout. Authored waypoints are
	// component-local, retained explicitly; callers review routing after topology edits.
	for _, lk := range s.Links {
		a, z := boxes[lk.From], boxes[lk.To]
		ax, zy := a.X+a.W, a.Y+a.H/2
		zx, zz := z.X, z.Y+z.H/2
		if z.X < a.X {
			ax = a.X
			zx = z.X + z.W
		}
		pts := [][2]float64{{ax, zy}}
		for _, q := range lk.Route {
			pts = append(pts, [2]float64{s.X + q[0], s.Y + q[1]})
		}
		if len(lk.Route) == 0 && math.Abs(zy-zz) > .01 {
			mid := (ax + zx) / 2
			pts = append(pts, [2]float64{mid, zy}, [2]float64{mid, zz})
		}
		pts = append(pts, [2]float64{zx, zz})
		for i := 1; i < len(pts); i++ {
			for key, box := range boxes {
				if key != lk.From && key != lk.To {
					box = Rect{box.X - 2, box.Y - 2, box.W + 4, box.H + 4}
				}
				if processSegmentEntersBox(pts[i-1], pts[i], box) {
					return nil, true, fmt.Errorf("process link %s segment%d crosses step %s or its clearance; author an explicit route through open space or rearrange steps", lk.Key, i, key)
				}
			}
		}
		fromSite, toSite := 3, 1
		if z.X < a.X {
			fromSite, toSite = 1, 3
		}
		linkID := id + ".links." + lk.Key
		firstItem := len(p.Items)
		if e = r.planSemanticConnector(p, linkID, pts, lk.Outcome, pptx.ConnectorConnection{Begin: pptx.ConnectorEndpoint{ObjectName: id + ".steps." + lk.From + ".surface", Site: fromSite}, End: pptx.ConnectorEndpoint{ObjectName: id + ".steps." + lk.To + ".surface", Site: toSite}}, ctx); e != nil {
			return nil, true, e
		}
		// Validate the actual measured patch, including padding, before steps
		// paint over it. Outcome semantics must never be hidden by node ink.
		for _, item := range p.Items[firstItem:] {
			if item.Shape == nil || item.Shape.Record.ID != linkID+".label-patch" {
				continue
			}
			patch := item.Shape.Record.Rect
			if lk.LabelPosition != nil {
				dx := s.X + lk.LabelPosition[0] - (patch.X + patch.W/2)
				dy := s.Y + lk.LabelPosition[1] - (patch.Y + patch.H/2)
				patch.X += dx
				patch.Y += dy
				item.Shape.Record.Rect = patch
				item.Shape.Props.PositionProps = pos(patch)
				for _, text := range p.Items[firstItem:] {
					if text.Text != nil && text.Text.ID == linkID+".label" {
						text.Text.Rect.X += dx
						text.Text.Rect.Y += dy
					}
				}
			}
			for key, box := range boxes {
				if patch.X < box.X+box.W+2 && patch.X+patch.W > box.X-2 && patch.Y < box.Y+box.H+2 && patch.Y+patch.H > box.Y-2 {
					return nil, true, fmt.Errorf("process link %s measured outcome label overlaps step %s or its 2pt clearance; author a clear route and label_position, increase column spacing, or shorten the outcome", lk.Key, key)
				}
			}
		}
	}
	for _, v := range s.Steps {
		b := boxes[v.Key]
		from := len(p.Items)
		sf := v.Surface
		if sf == "" {
			sf = "light"
		}
		if v.Kind == "start" || v.Kind == "end" {
			sf = "inverse"
		}
		kind := pptx.ShapeTypeRect
		if v.Kind == "decision" {
			kind = pptx.ShapeTypeDiamond
		}
		fill, e := r.sceneColor(sf, "bg")
		if e != nil {
			return nil, true, e
		}
		pre := id + ".steps." + v.Key
		if e = r.diagramShape(p, pre+".surface", b, kind, fill, strong, 1, "solid", nil); e != nil {
			return nil, true, e
		}
		tw := b.W - 16
		if v.Kind == "decision" {
			tw = b.W * .58
		}
		if e = r.diagramText(p, pre+".label", v.Label, "small", Rect{b.X + (b.W-tw)/2, b.Y, tw, b.H}, sf, "display", "center", 600, true); e != nil {
			return nil, true, e
		}
		marker := []string{}
		if s.Start == v.Key {
			marker = append(marker, "Start")
		}
		if s.Current == v.Key {
			marker = append(marker, "Current")
		}
		for _, k := range s.End {
			if k == v.Key {
				marker = append(marker, "End")
			}
		}
		if len(marker) > 0 {
			if e = r.diagramText(p, pre+".marker", strings.Join(marker, " · "), "label", Rect{b.X, b.Y + b.H + 3, b.W, 14}, surface, "secondary", "center", 600, true); e != nil {
				return nil, true, e
			}
		}
		groupBounds := b
		if len(marker) > 0 {
			groupBounds.H += 17
		}
		sceneDataGroup(p, pre, "process.step", from, groupBounds)
	}
	p = diagramFinish(p, "scene.process")
	if p.Bounds.X < s.X-.01 || p.Bounds.Y < s.Y-.01 || p.Bounds.X+p.Bounds.W > s.X+s.W+.01 || p.Bounds.Y+p.Bounds.H > s.Y+s.H+.01 {
		return nil, true, fmt.Errorf("process content exceeds allocation")
	}
	return p, true, nil
}

// Open-interior slab intersection permits exact attached edge endpoints while
// refusing hidden paths through either endpoint body or unrelated step boxes.
func processSegmentEntersBox(a, b [2]float64, box Rect) bool {
	lo, hi := 0., 1.
	for axis, bounds := range [][2]float64{{box.X + 1e-6, box.X + box.W - 1e-6}, {box.Y + 1e-6, box.Y + box.H - 1e-6}} {
		d := b[axis] - a[axis]
		if math.Abs(d) < 1e-12 {
			if a[axis] < bounds[0] || a[axis] > bounds[1] {
				return false
			}
			continue
		}
		u, v := (bounds[0]-a[axis])/d, (bounds[1]-a[axis])/d
		if u > v {
			u, v = v, u
		}
		lo = math.Max(lo, u)
		hi = math.Min(hi, v)
		if lo > hi {
			return false
		}
	}
	return hi >= lo && hi >= 0 && lo <= 1
}
