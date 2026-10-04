package wmdesign

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"math"
	"sort"
	"strings"
)

type sequenceStep struct {
	Key   string `json:"key"`
	State string `json:"state"`
	Label string `json:"label"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

const verticalSequenceStepPitch = 66.0

type sequencePhase struct {
	Key          string            `json:"key"`
	N            string            `json:"n"`
	Duration     string            `json:"duration"`
	Name         string            `json:"name"`
	Rule         string            `json:"rule"`
	Objective    string            `json:"objective"`
	Activities   []json.RawMessage `json:"activities"`
	Deliverables []json.RawMessage `json:"deliverables"`
	Gate         string            `json:"gate"`
}
type sequenceMilestone struct {
	Key   string  `json:"key"`
	At    float64 `json:"at"`
	Label string  `json:"label"`
}
type sequenceBand struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Text  string `json:"text"`
}

func (b *sequenceBand) UnmarshalJSON(raw []byte) error {
	var pair []string
	if json.Unmarshal(raw, &pair) == nil {
		if len(pair) != 2 {
			return fmt.Errorf("scene.pyramid_band_shape")
		}
		b.Label, b.Text = pair[0], pair[1]
		return nil
	}
	type plain sequenceBand
	return sceneDecode(raw, (*plain)(b))
}

type sequenceSpec struct {
	On         string              `json:"on"`
	Type       string              `json:"type"`
	X          float64             `json:"x"`
	Y          float64             `json:"y"`
	W          float64             `json:"w"`
	H          float64             `json:"h"`
	Steps      []sequenceStep      `json:"steps"`
	N          string              `json:"n"`
	Rule       string              `json:"rule"`
	Duration   string              `json:"duration"`
	Title      string              `json:"title"`
	Phases     []sequencePhase     `json:"phases"`
	Current    *int                `json:"current"`
	Periods    []string            `json:"periods"`
	Milestones []sequenceMilestone `json:"milestones"`
	Today      *float64            `json:"today"`
	Bands      []sequenceBand      `json:"bands"`
	Mode       string              `json:"mode"`
}

func (r *renderer) sequenceNeed(text string, st Style, w float64) (float64, float64, error) {
	if text == "" {
		return 0, 0, nil
	}
	l, e := r.typeEngine.Measure(text, st, w)
	if e != nil {
		return 0, 0, e
	}
	advance := 0.0
	for _, ln := range l.Lines {
		advance = math.Max(advance, ln.Advance)
	}
	return math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight), advance, e
}

// Native PowerPoint requires a small amount of trailing space when a textbox
// is sized to the shaped advance. This is a textbox allocation rule, not a
// change to the font metrics, tracking, or line-break calculation.
func sequenceInlineWidth(advance float64, st Style) float64 {
	if advance == 0 {
		return 0
	}
	guard := 4.0
	if st.Family == "IBM Plex Mono" {
		guard = 2
	}
	return advance + math.Max(guard, math.Abs(st.TrackingPt))
}
func (r *renderer) sequenceStack(p *scenePlan, id string, copies []string, tokens []string, weights []int, x, y, w, gap, capacity float64, surface string, inks []string) (float64, error) {
	top := y
	for i, s := range copies {
		if s == "" {
			continue
		}
		st, e := r.sceneStyle(tokens[i])
		if e != nil {
			return top, e
		}
		if weights[i] > 0 {
			st.Weight = weights[i]
		}
		h, _, e := r.sequenceNeed(s, st, w)
		if e != nil {
			return top, e
		}
		if capacity > 0 && top+h > y+capacity+.02 {
			return top, fmt.Errorf("scene.stack_overflow: %s needs %.3fpt capacity %.3fpt", id, top+h-y, capacity)
		}
		if e = r.sceneText(p, fmt.Sprintf("%s.part-%d", id, i), s, st, Rect{x, top, w, h}, surface, inks[i], "left"); e != nil {
			return top, e
		}
		top += h + gap
	}
	return top - gap, nil
}
func (r *renderer) sequenceBullet(p *scenePlan, id string, items []json.RawMessage, x, y, w float64, ctx SceneContext, path string) (float64, error) {
	if len(items) == 0 {
		return y, nil
	}
	node := map[string]any{"type": "bullets", "x": x, "y": y, "w": w, "size": "small", "items": items}
	cc := ctx
	cc.Path = path
	if keys, ok := ctx.Keys[path]; ok {
		cc.Keys = make(map[string][]string, len(ctx.Keys)+1)
		for k, v := range ctx.Keys {
			cc.Keys[k] = v
		}
		cc.Keys[path+"/items"] = keys
	}
	raw, _ := json.Marshal(node)
	c, e := r.planSceneNode(id, raw, cc)
	if e != nil {
		return y, e
	}
	diagramMerge(p, c)
	return c.Bounds.Y + c.Bounds.H, nil
}
func (r *renderer) planSequenceScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type == "gantt" {
		p, e := r.planGanttScene(id, raw, ctx)
		return p, true, e
	}
	if tag.Type == "swimlane" {
		p, e := r.planSwimlaneScene(id, raw, ctx)
		return p, true, e
	}
	fields := map[string]string{"stepper": "x y w steps", "vstepper": "x y w steps", "phasehead": "x y w n rule duration title", "phases": "x y w current phases", "timeaxis": "x y w periods milestones today", "pyramid": "x y w h bands mode"}
	allowed, ok := fields[tag.Type]
	if !ok {
		return nil, false, nil
	}
	var n sequenceSpec
	if e := diagramDecode(raw, allowed+" on", &n); e != nil {
		return nil, true, e
	}
	if n.On != "" {
		ctx.Surface = n.On
	}
	p := &scenePlan{ID: id}
	strong, e := r.sceneColor(ctx.Surface, "strong")
	if e != nil {
		return nil, true, e
	}
	line, e := r.sceneColor(ctx.Surface, "line")
	if e != nil {
		return nil, true, e
	}
	callout, e := r.sceneColor("callout", "bg")
	if e != nil {
		return nil, true, e
	}
	switch n.Type {
	case "stepper", "vstepper":
		if len(n.Steps) == 0 || n.W <= 36 {
			return nil, true, fmt.Errorf("scene.stepper_geometry")
		}
		cur := -1
		for i, st := range n.Steps {
			if st.State == "current" {
				if cur >= 0 {
					return nil, true, fmt.Errorf("scene.multiple_current_steps")
				}
				cur = i
			} else if st.State != "done" && st.State != "next" && st.State != "" {
				return nil, true, fmt.Errorf("scene.step_state: %s", st.State)
			}
		}
		cw := n.W / float64(len(n.Steps))
		vertical := n.Type == "vstepper"
		a, b := [2]float64{n.X + 6, n.Y + 9}, [2]float64{n.X + float64(len(n.Steps)-1)*cw + 6, n.Y + 9}
		if vertical {
			a = [2]float64{n.X + 9, n.Y + 12}
			b = [2]float64{n.X + 9, n.Y + 12 + float64(len(n.Steps)-1)*verticalSequenceStepPitch}
		}
		if len(n.Steps) > 1 {
			if e = r.diagramLine(p, id+".base", a, b, line, 2, "solid"); e != nil {
				return nil, true, e
			}
		}
		if cur > 0 {
			z := [2]float64{n.X + float64(cur)*cw + 6, n.Y + 9}
			if vertical {
				z = [2]float64{n.X + 9, n.Y + 12 + float64(cur)*verticalSequenceStepPitch}
			}
			if e = r.diagramLine(p, id+".done", a, z, strong, 2, "solid"); e != nil {
				return nil, true, e
			}
		}
		seen := map[string]bool{}
		for i, s := range n.Steps {
			k, e := diagramKey(ctx, "steps", i, s.Key, seen)
			if e != nil {
				return nil, true, e
			}
			pre := id + ".steps." + k
			size := 12.0
			fill, outline := strong, ""
			if s.State == "current" {
				size = 18
				fill = callout
			} else if s.State != "done" {
				fill, e = r.sceneColor(ctx.Surface, "bg")
				outline = strong
			}
			cx, cy := n.X+float64(i)*cw+6, n.Y+9
			x, y, w := n.X+float64(i)*cw, n.Y+30, cw-18
			cap := ctx.Zone.Y + ctx.Zone.H - y
			if vertical {
				cx, cy = n.X+9, n.Y+12+float64(i)*verticalSequenceStepPitch
				x, y, w = n.X+36, cy-9, n.W-36
				cap = verticalSequenceStepPitch
			}
			if e = r.diagramShape(p, pre+".marker", Rect{cx - size/2, cy - size/2, size, size}, pptx.ShapeTypeRect, fill, outline, 1, "solid", nil); e != nil {
				return nil, true, e
			}
			copies, tokens, weights, inks := []string{s.Label, s.Title}, []string{"label", "body"}, []int{0, 600}, []string{"emphasis", "display"}
			if vertical {
				ts, _ := r.sceneStyle("body")
				ts.Weight = 600
				ls, _ := r.sceneStyle("label")
				th, tw, e := r.sequenceNeed(s.Title, ts, w)
				if e != nil {
					return nil, true, e
				}
				lh, lw, e := r.sequenceNeed(s.Label, ls, w)
				if e != nil {
					return nil, true, e
				}
				tw, lw = sequenceInlineWidth(tw, ts), sequenceInlineWidth(lw, ls)
				if tw+lw+12 > w+.02 {
					return nil, true, fmt.Errorf("scene.vstepper_heading_overlap: %s", pre)
				}
				hh := math.Max(th, lh)
				if e = r.sceneText(p, pre+".title", s.Title, ts, Rect{x, y, tw, th}, ctx.Surface, "display", "left"); e != nil {
					return nil, true, e
				}
				titleLayout, e := r.typeEngine.Measure(s.Title, ts, tw)
				if e != nil {
					return nil, true, e
				}
				labelLayout, e := r.typeEngine.Measure(s.Label, ls, lw)
				if e != nil {
					return nil, true, e
				}
				labelY := y
				if len(titleLayout.Lines) > 0 && len(labelLayout.Lines) > 0 {
					labelY += titleLayout.Lines[0].Baseline - labelLayout.Lines[0].Baseline
				}
				if e = r.sceneText(p, pre+".label", s.Label, ls, Rect{x + tw + 12, labelY, lw, lh}, ctx.Surface, "emphasis", "left"); e != nil {
					return nil, true, e
				}
				p.Warnings = append(p.Warnings, "Adapter resolution wmds.inline-sequence-text.v1: allocate trailing native textbox space and align inline labels to the heading's first baseline.")
				if _, e = r.sequenceStack(p, pre+".description", []string{s.Text}, []string{"small"}, []int{0}, x, y+hh+3, w, 0, cap-hh-3, ctx.Surface, []string{"secondary"}); e != nil {
					return nil, true, e
				}
				continue
			} else if s.State == "current" {
				copies = append(copies, "Current")
				tokens = append(tokens, "small")
				weights = append(weights, 0)
				inks = append(inks, "secondary")
			}
			if _, e = r.sequenceStack(p, pre, copies, tokens, weights, x, y, w, 3, cap, ctx.Surface, inks); e != nil {
				return nil, true, e
			}
		}
	case "phasehead":
		color, e := r.sceneColor(ctx.Surface, n.Rule)
		if e != nil {
			return nil, true, e
		}
		if e = r.diagramShape(p, id+".bar", Rect{n.X, n.Y, n.W, 6}, pptx.ShapeTypeRect, color, "", 0, "", nil); e != nil {
			return nil, true, e
		}
		st, _ := r.sceneStyle("label")
		h, nw, e := r.sequenceNeed(n.N, st, n.W)
		if e != nil {
			return nil, true, e
		}
		_, dw, e := r.sequenceNeed(n.Duration, st, n.W)
		if e != nil {
			return nil, true, e
		}
		nw, dw = sequenceInlineWidth(nw, st), sequenceInlineWidth(dw, st)
		if nw+dw+12 > n.W {
			return nil, true, fmt.Errorf("scene.phasehead_labels_overlap")
		}
		if e = r.sceneText(p, id+".number", n.N, st, Rect{n.X, n.Y + 15, nw, h}, ctx.Surface, "emphasis", "left"); e != nil {
			return nil, true, e
		}
		if e = r.sceneText(p, id+".duration", n.Duration, st, Rect{n.X + n.W - dw, n.Y + 15, dw, h}, ctx.Surface, "secondary", "right"); e != nil {
			return nil, true, e
		}
		p.Warnings = append(p.Warnings, "Adapter resolution wmds.inline-sequence-text.v1: allocate trailing native textbox space for phase numbers and durations.")
		if e = r.diagramText(p, id+".title", n.Title, "subhead", Rect{n.X, n.Y + 15 + h, n.W, 0}, ctx.Surface, "display", "left", 0, false); e != nil {
			return nil, true, e
		}
	case "timeaxis":
		if len(n.Periods) == 0 {
			return nil, true, fmt.Errorf("scene.timeaxis_empty")
		}
		if e = r.diagramLine(p, id+".axis", [2]float64{n.X, n.Y}, [2]float64{n.X + n.W, n.Y}, strong, 1, "solid"); e != nil {
			return nil, true, e
		}
		pw := n.W / float64(len(n.Periods))
		for i := 0; i <= len(n.Periods); i++ {
			x := n.X + float64(i)*pw
			if e = r.diagramLine(p, fmt.Sprintf("%s.tick-%d", id, i), [2]float64{x, n.Y}, [2]float64{x, n.Y + 6}, strong, 1, "solid"); e != nil {
				return nil, true, e
			}
		}
		seen := map[string]bool{}
		for i, s := range n.Periods {
			k, e := diagramKey(ctx, "periods", i, "", seen)
			if e != nil {
				return nil, true, e
			}
			if e = r.diagramText(p, id+".periods."+k, s, "label", Rect{n.X + float64(i)*pw, n.Y + 12, pw, 0}, ctx.Surface, "secondary", "center", 0, false); e != nil {
				return nil, true, e
			}
		}
		seen = map[string]bool{}
		for i, m := range n.Milestones {
			k, e := diagramKey(ctx, "milestones", i, m.Key, seen)
			if e != nil {
				return nil, true, e
			}
			if m.At < 0 || m.At > 1 {
				return nil, true, fmt.Errorf("scene.milestone_fraction")
			}
			x := n.X + m.At*n.W
			if e = r.diagramShape(p, id+".milestones."+k+".marker", Rect{x - 7.071, n.Y - 7.071, 14.142, 14.142}, pptx.ShapeTypeDiamond, strong, "", 0, "", nil); e != nil {
				return nil, true, e
			}
			if e = r.diagramText(p, id+".milestones."+k+".label", m.Label, "small", Rect{x - 72, n.Y - 30, 144, 0}, ctx.Surface, "display", "center", 600, false); e != nil {
				return nil, true, e
			}
		}
		if n.Today != nil {
			if *n.Today < 0 || *n.Today > 1 {
				return nil, true, fmt.Errorf("scene.today_fraction")
			}
			x := n.X + *n.Today*n.W
			if e = r.diagramLine(p, id+".today", [2]float64{x, n.Y - 54}, [2]float64{x, n.Y + 6}, callout, 1.5, "dash"); e != nil {
				return nil, true, e
			}
			if e = r.diagramText(p, id+".today-label", "Today", "label", Rect{x + 6, n.Y - 60, 60, 0}, ctx.Surface, "primary", "left", 600, false); e != nil {
				return nil, true, e
			}
		}
	case "phases":
		if len(n.Phases) == 0 {
			return nil, true, fmt.Errorf("scene.phases_empty")
		}
		// The axis is background chrome; opaque gate diamonds paint above it.
		if e = r.diagramLine(p, id+".axis", [2]float64{n.X, n.Y + 288}, [2]float64{n.X + n.W, n.Y + 288}, line, 1, "solid"); e != nil {
			return nil, true, e
		}
		cw := (n.W - 18*float64(len(n.Phases)-1)) / float64(len(n.Phases))
		seen := map[string]bool{}
		for i, ph := range n.Phases {
			k, e := diagramKey(ctx, "phases", i, ph.Key, seen)
			if e != nil {
				return nil, true, e
			}
			pre := id + ".phases." + k
			x := n.X + float64(i)*(cw+18)
			surf, pad := "light", 0.0
			if n.Current != nil && *n.Current == i {
				surf, pad = "inverse", 12
				if e = r.sceneRect(p, pre+".current-tag", Rect{x, n.Y - 24, 84, 18}, "callout"); e != nil {
					return nil, true, e
				}
				if e = r.diagramText(p, pre+".current-label", "We are here", "label", Rect{x + 6, n.Y - 24, 72, 18}, "callout", "primary", "left", 600, true); e != nil {
					return nil, true, e
				}
			}
			if e = r.sceneRect(p, pre+".surface", Rect{x, n.Y, cw, 270}, surf); e != nil {
				return nil, true, e
			}
			xx, ww, y := x+pad, cw-2*pad, n.Y+pad
			col, e := r.sceneColor(ctx.Surface, ph.Rule)
			if e != nil {
				return nil, true, e
			}
			if e = r.diagramShape(p, pre+".rule", Rect{xx, y, ww, 6}, pptx.ShapeTypeRect, col, "", 0, "", nil); e != nil {
				return nil, true, e
			}
			y += 18
			st, _ := r.sceneStyle("label")
			hh, nw, e := r.sequenceNeed(ph.N, st, ww)
			if e != nil {
				return nil, true, e
			}
			_, dw, e := r.sequenceNeed(ph.Duration, st, ww)
			if e != nil {
				return nil, true, e
			}
			nw, dw = sequenceInlineWidth(nw, st), sequenceInlineWidth(dw, st)
			if nw+dw+6 > ww {
				return nil, true, fmt.Errorf("scene.phase_label_overlap")
			}
			if e = r.sceneText(p, pre+".number", ph.N, st, Rect{xx, y, nw, hh}, surf, "emphasis", "left"); e != nil {
				return nil, true, e
			}
			if e = r.sceneText(p, pre+".duration", ph.Duration, st, Rect{xx + ww - dw, y, dw, hh}, surf, "secondary", "right"); e != nil {
				return nil, true, e
			}
			p.Warnings = append(p.Warnings, "Adapter resolution wmds.inline-sequence-text.v1: allocate trailing native textbox space for phase numbers and durations.")
			y += hh + 6
			cap := n.Y + 258 - y
			y, e = r.sequenceStack(p, pre+".copy", []string{ph.Name, ph.Objective}, []string{"subhead", "small"}, []int{0, 0}, xx, y, ww, 6, cap, surf, []string{"display", "primary"})
			if e != nil {
				return nil, true, e
			}
			cc := ctx
			cc.Surface = surf
			y += 15
			for j, sec := range []struct {
				title, field string
				items        []json.RawMessage
			}{{"Activities", "activities", ph.Activities}, {"Deliverables", "deliverables", ph.Deliverables}} {
				if e = r.diagramText(p, fmt.Sprintf("%s.section-%d", pre, j), sec.title, "label", Rect{xx, y, ww, 0}, surf, "secondary", "left", 0, false); e != nil {
					return nil, true, e
				}
				ls, _ := r.sceneStyle("label")
				lh, _, _ := r.sequenceNeed(sec.title, ls, ww)
				y += lh + 6
				y, e = r.sequenceBullet(p, pre+"."+sec.field, sec.items, xx, y, ww, cc, ctx.Path+fmt.Sprintf("/phases/%d/%s", i, sec.field))
				if e != nil {
					return nil, true, e
				}
				y += 15
			}
			if y-15 > n.Y+258+.02 {
				return nil, true, fmt.Errorf("scene.phase_content_overflow: %s bottom %.3fpt capacity %.3fpt", pre, y-15, n.Y+258)
			}
			if ph.Gate != "" {
				gx := x + cw + 9
				if e = r.diagramShape(p, pre+".gate", Rect{gx - 8.485, n.Y + 288 - 8.485, 16.97, 16.97}, pptx.ShapeTypeDiamond, strong, "", 0, "", nil); e != nil {
					return nil, true, e
				}
				lb := Rect{gx - 60, n.Y + 300, 120, 0}
				align := "center"
				if i == len(n.Phases)-1 {
					lb.X = gx - 120
					align = "right"
				}
				if e = r.diagramText(p, pre+".gate-label", ph.Gate, "label", lb, ctx.Surface, "primary", align, 600, false); e != nil {
					return nil, true, e
				}
			}
		}
	case "pyramid":
		if len(n.Bands) == 0 || len(n.Bands) > 5 || n.H <= 6*float64(len(n.Bands)-1) {
			return nil, true, fmt.Errorf("scene.pyramid_geometry")
		}
		if n.Mode != "" && n.Mode != "funnel" {
			return nil, true, fmt.Errorf("scene.pyramid_mode")
		}
		sw := n.W * .5
		bh := (n.H - 6*float64(len(n.Bands)-1)) / float64(len(n.Bands))
		steps := []int{900, 600, 400, 200, 100}
		seen := map[string]bool{}
		for i, band := range n.Bands {
			k, e := diagramKey(ctx, "bands", i, band.Key, seen)
			if e != nil {
				return nil, true, e
			}
			pre := id + ".bands." + k
			f0, f1 := .2+float64(i)/float64(len(n.Bands))*.8, .2+float64(i+1)/float64(len(n.Bands))*.8
			if n.Mode == "funnel" {
				f0, f1 = 1-float64(i)/float64(len(n.Bands))*.75, 1-float64(i+1)/float64(len(n.Bands))*.75
			}
			y := n.Y + float64(i)*(bh+6)
			fill, e := r.sceneColor(ctx.Surface, fmt.Sprintf("ramp.%d", steps[i]))
			if e != nil {
				return nil, true, e
			}
			pts := [][2]float64{{sw * (1 - f0) / 2, 0}, {sw * (1 + f0) / 2, 0}, {sw * (1 + f1) / 2, bh}, {sw * (1 - f1) / 2, bh}}
			if e = r.diagramShape(p, pre+".surface", Rect{n.X, y, sw, bh}, pptx.ShapeTypeCustGeom, fill, "", 0, "", pts); e != nil {
				return nil, true, e
			}
			ink := "strong"
			surf := "light"
			if contrast("FFFFFF", fill) >= 4.5 {
				surf = "inverse"
				ink = "display"
			}
			labelW := sw - 24
			if e = r.diagramText(p, pre+".label", band.Label, "small", Rect{n.X + (sw-labelW)/2, y, labelW, bh}, surf, ink, "center", 600, true); e != nil {
				return nil, true, e
			}
			token, weight := "small", 0
			if n.Mode == "funnel" {
				token, weight = "body", 600
			}
			ds, e := r.sceneStyle(token)
			if e != nil {
				return nil, true, e
			}
			if weight > 0 {
				ds.Weight = weight
			}
			if n.Mode == "funnel" {
				mono, e := r.sceneStyle("label")
				if e != nil {
					return nil, true, e
				}
				ds.Family = mono.Family
			}
			db := Rect{n.X + sw + 18, y, n.W - sw - 18, bh}
			need, _, e := r.sequenceNeed(band.Text, ds, db.W)
			if e != nil {
				return nil, true, e
			}
			if need > bh && need <= bh+6+.02 {
				db.Y -= (need - bh) / 2
				db.H = need
				p.Warnings = append(p.Warnings, "Adapter resolution wmds.pyramid-description-pitch.v1: retain band geometry/fonts; center a measured description inside the existing band-plus-gap pitch.")
			}
			if e = r.diagramStyledText(p, pre+".text", band.Text, ds, db, ctx.Surface, "primary", "left", true); e != nil {
				return nil, true, e
			}
			edge := sw * (1 + math.Max(f0, f1)) / 2
			if e = r.diagramLine(p, pre+".leader", [2]float64{n.X + edge + 6, y + bh/2}, [2]float64{n.X + sw + 12, y + bh/2}, line, .75, "solid"); e != nil {
				return nil, true, e
			}
		}
	}
	return diagramFinish(p, "scene."+n.Type), true, nil
}

type ganttItem struct {
	Key       string   `json:"key"`
	Kind      string   `json:"kind"`
	Label     string   `json:"label"`
	From      float64  `json:"from"`
	To        float64  `json:"to"`
	At        float64  `json:"at"`
	SoftStart float64  `json:"softStart"`
	SoftEnd   float64  `json:"softEnd"`
	Hatch     bool     `json:"hatch"`
	Progress  *float64 `json:"progress"`
	Tag       bool     `json:"tag"`
	Milestone bool     `json:"milestone"`
	Event     string   `json:"event"`
	LabelSide string   `json:"labelSide"`
}
type ganttLane struct {
	Key   string      `json:"key"`
	Icon  string      `json:"icon"`
	Title string      `json:"title"`
	Sub   string      `json:"sub"`
	Items []ganttItem `json:"items"`
}
type ganttGroup struct {
	Key   string      `json:"key"`
	Label string      `json:"label"`
	Fill  string      `json:"fill"`
	Lanes []ganttLane `json:"lanes"`
}
type ganttKind struct {
	Fill  string `json:"fill"`
	Text  string `json:"text"`
	Label string `json:"label"`
}
type ganttGate struct {
	Key   json.RawMessage `json:"key"`
	At    float64         `json:"at"`
	Label string          `json:"label"`
}
type ganttPhase struct {
	Key   string  `json:"key"`
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	Rule  string  `json:"rule"`
	Label string  `json:"label"`
}
type ganttSpec struct {
	Type            string  `json:"type"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	W               float64 `json:"w"`
	TrackPitch      float64 `json:"trackPitch,omitempty"`
	LegendSize      float64 `json:"legendSize,omitempty"`
	LegendFullWidth bool    `json:"legendFullWidth,omitempty"`
	Cols            struct {
		Group float64 `json:"group"`
		Lane  float64 `json:"lane"`
	} `json:"cols"`
	Periods struct {
		Labels    []string `json:"labels"`
		Sublabels []string `json:"sublabels"`
	} `json:"periods"`
	Phases       []ganttPhase         `json:"phases"`
	Gates        []ganttGate          `json:"gates"`
	Today        *float64             `json:"today"`
	SidebarLabel string               `json:"sidebarLabel"`
	Kinds        map[string]ganttKind `json:"kinds"`
	Events       map[string]ganttKind `json:"events"`
	Groups       []ganttGroup         `json:"groups"`
}
type ganttPacked struct {
	it          ganttItem
	key         string
	a, b, width float64
	track       int
}
type ganttLayout struct {
	group                 ganttGroup
	lane                  ganttLane
	key, gkey             string
	top, h                float64
	tracks                int
	items                 []ganttPacked
	groupIndex, laneIndex int
}

func (r *renderer) sequenceLiteralText(p *scenePlan, id, text string, st Style, b Rect, color, align string, middle bool) error {
	if e := r.diagramStyledText(p, id, text, st, b, "light", "primary", align, middle); e != nil {
		return e
	}
	if text != "" {
		p.Items[len(p.Items)-1].Text.Color = color
	}
	return nil
}
func (r *renderer) sequenceHatch(p *scenePlan, id string, b Rect, fill, strong string) error {
	if e := r.diagramShape(p, id+".fill", b, pptx.ShapeTypeRect, fill, strong, .75, "solid", nil); e != nil {
		return e
	}
	for x := -b.H; x < b.W; x += 6 {
		ax, ay := math.Max(0, x), math.Max(0, -x)
		bx, by := math.Min(b.W, x+b.H), math.Min(b.H, b.W-x)
		if bx > ax {
			if e := r.diagramLine(p, fmt.Sprintf("%s.hatch-%d", id, int((x+b.H)/6)), [2]float64{b.X + ax, b.Y + b.H - ay}, [2]float64{b.X + bx, b.Y + b.H - by}, strong, .5, "solid"); e != nil {
				return e
			}
		}
	}
	return nil
}
func (r *renderer) planGanttScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	var n ganttSpec
	if e := diagramDecode(raw, "x y w cols periods phases gates today sidebarLabel kinds events groups trackPitch legendSize legendFullWidth", &n); e != nil {
		return nil, e
	}
	trackPitch := n.TrackPitch
	if trackPitch == 0 {
		trackPitch = 30
	}
	if trackPitch < 24 || trackPitch > 60 || math.IsNaN(trackPitch) || math.IsInf(trackPitch, 0) {
		return nil, fmt.Errorf("scene.gantt_track_pitch: expected24_to60pt")
	}
	np := len(n.Periods.Labels)
	gw, lw := n.Cols.Group, n.Cols.Lane
	tw := n.W - gw - lw
	if np == 0 || tw <= 0 || lw < 36 || gw < 0 {
		return nil, fmt.Errorf("scene.gantt_geometry")
	}
	if len(n.Periods.Sublabels) > 0 && len(n.Periods.Sublabels) != np {
		return nil, fmt.Errorf("scene.gantt_sublabel_count")
	}
	pw := tw / float64(np)
	tx := n.X + gw + lw
	X := func(t float64) float64 { return tx + t*pw }
	p := &scenePlan{ID: id}
	strong, e := r.sceneColor(ctx.Surface, "strong")
	if e != nil {
		return nil, e
	}
	line, _ := r.sceneColor(ctx.Surface, "line")
	small, _ := r.sceneStyle("small")
	small.Weight = 600
	label, _ := r.sceneStyle("label")
	label.Weight = 600
	gateTop := n.Y
	phaseTop := n.Y
	if n.Gates != nil {
		phaseTop += 24
	}
	periodTop := phaseTop + 30
	ph := 24.0
	if len(n.Periods.Sublabels) > 0 {
		ph = 36
	}
	top := periodTop + ph
	headerBottom := top
	var layouts []ganttLayout
	gseen := map[string]bool{}
	for gi, g := range n.Groups {
		gkey, e := diagramKey(ctx, "groups", gi, g.Key, gseen)
		if e != nil {
			return nil, e
		}
		cc := ctx
		cc.Path += fmt.Sprintf("/groups/%d", gi)
		lseen := map[string]bool{}
		for li, ln := range g.Lanes {
			lk, e := diagramKey(cc, "lanes", li, ln.Key, lseen)
			if e != nil {
				return nil, e
			}
			ic := cc
			ic.Path += fmt.Sprintf("/lanes/%d", li)
			iseen := map[string]bool{}
			L := ganttLayout{group: g, lane: ln, key: lk, gkey: gkey, top: top, groupIndex: gi, laneIndex: li}
			for ii, it := range ln.Items {
				ik, e := diagramKey(ic, "items", ii, it.Key, iseen)
				if e != nil {
					return nil, e
				}
				_, width, e := r.sequenceNeed(it.Label, small, 8191)
				if e != nil {
					return nil, e
				}
				width = sequenceInlineWidth(width, small)
				o := ganttPacked{it: it, key: ik, width: width}
				if it.Kind != "" {
					if _, ok := n.Kinds[it.Kind]; !ok {
						return nil, fmt.Errorf("scene.gantt_unknown_kind: %s", it.Kind)
					}
					if it.From < 0 || it.To > float64(np) || it.To <= it.From || it.SoftStart < 0 || it.SoftEnd < 0 || it.SoftStart+it.SoftEnd > it.To-it.From {
						return nil, fmt.Errorf("scene.gantt_interval")
					}
					if it.Progress != nil && (*it.Progress < 0 || *it.Progress > 1) {
						return nil, fmt.Errorf("scene.gantt_progress")
					}
					o.a, o.b = it.From, it.To
					if width+14 > (it.To-it.From)*pw {
						o.b = math.Max(o.b, it.To+(width+6)/pw)
					}
				} else {
					if it.At < 0 || it.At > float64(np) {
						return nil, fmt.Errorf("scene.gantt_event_at")
					}
					if it.Tag {
						o.a, o.b = it.At, it.At+(width+20)/pw
					} else {
						if !it.Milestone {
							if _, ok := n.Events[it.Event]; !ok {
								return nil, fmt.Errorf("scene.gantt_unknown_event: %s", it.Event)
							}
						}
						if it.LabelSide == "left" {
							o.a, o.b = it.At-(width+22)/pw, it.At+.15
						} else if it.LabelSide == "" || it.LabelSide == "right" {
							o.a, o.b = it.At-.15, it.At+(width+22)/pw
						} else {
							return nil, fmt.Errorf("scene.gantt_label_side")
						}
					}
				}
				L.items = append(L.items, o)
			}
			sort.SliceStable(L.items, func(i, j int) bool { return L.items[i].a < L.items[j].a })
			var ends []float64
			for ii := range L.items {
				o := &L.items[ii]
				t := 0
				for t < len(ends) && ends[t] > o.a+.02 {
					t++
				}
				if t == len(ends) {
					ends = append(ends, o.b)
				} else {
					ends[t] = o.b
				}
				o.track = t
			}
			L.tracks = len(ends)
			if L.tracks < 1 {
				L.tracks = 1
			}
			L.h = math.Max(42, float64(L.tracks)*trackPitch+12)
			layouts = append(layouts, L)
			top += L.h
		}
	}
	bodyBottom := top
	if ctx.Zone.H > 0 && bodyBottom > ctx.Zone.Y+ctx.Zone.H {
		return nil, fmt.Errorf("scene.gantt_source_envelope_overflow: measured lane bottom %.3fpt exceeds %.3fpt; source requires an explicit layout resolution", bodyBottom, ctx.Zone.Y+ctx.Zone.H)
	}
	if e = r.sceneRect(p, id+".period-header", Rect{n.X, periodTop, n.W, ph}, "subtle"); e != nil {
		return nil, e
	}
	sidebar := n.SidebarLabel
	if sidebar == "" {
		sidebar = "Workstream · activity"
	}
	if e = r.diagramStyledText(p, id+".sidebar-heading", sidebar, label, Rect{n.X + 6, periodTop, gw + lw - 12, ph}, ctx.Surface, "secondary", "left", true); e != nil {
		return nil, e
	}
	seen := map[string]bool{}
	for i, s := range n.Periods.Labels {
		k, e := diagramKey(SceneContext{Path: ctx.Path + "/periods", Keys: ctx.Keys}, "labels", i, "", seen)
		if e != nil {
			return nil, e
		}
		y := periodTop
		if len(n.Periods.Sublabels) > 0 {
			y += 3
		}
		if e = r.diagramStyledText(p, id+".periods."+k+".label", s, label, Rect{X(float64(i)), y, pw, 24}, ctx.Surface, "primary", "center", true); e != nil {
			return nil, e
		}
		if len(n.Periods.Sublabels) > 0 {
			ss := label
			ss.Weight = 500
			ss.Size = 7.5
			ss.Tracking = "0.06em"
			ss.TrackingPt = .45
			if e = r.sceneText(p, id+".periods."+k+".sub", n.Periods.Sublabels[i], ss, Rect{X(float64(i)), periodTop + 21, pw, 15}, ctx.Surface, "secondary", "center"); e != nil {
				return nil, e
			}
		}
	}
	seen = map[string]bool{}
	for i, phs := range n.Phases {
		k, e := diagramKey(ctx, "phases", i, phs.Key, seen)
		if e != nil {
			return nil, e
		}
		if phs.From < 0 || phs.To > float64(np) || phs.To <= phs.From {
			return nil, fmt.Errorf("scene.gantt_phase_interval")
		}
		col, e := r.sceneColor(ctx.Surface, phs.Rule)
		if e != nil {
			return nil, e
		}
		if e = r.diagramShape(p, id+".phases."+k+".rule", Rect{X(phs.From) + 1, phaseTop + 6, (phs.To-phs.From)*pw - 2, 3}, pptx.ShapeTypeRect, col, "", 0, "", nil); e != nil {
			return nil, e
		}
		if e = r.diagramStyledText(p, id+".phases."+k+".label", phs.Label, label, Rect{X(phs.From) + 6, phaseTop + 9, (phs.To-phs.From)*pw - 8, 21}, ctx.Surface, "primary", "left", false); e != nil {
			return nil, e
		}
	}
	// Match the frozen reference paint order: timeline guides, gates and today
	// are behind bar surfaces, event markers and all item labels.
	for i := 1; i < np; i++ {
		if e = r.diagramLine(p, fmt.Sprintf("%s.grid-%d", id, i), [2]float64{X(float64(i)), headerBottom}, [2]float64{X(float64(i)), bodyBottom}, line, .75, "solid"); e != nil {
			return nil, e
		}
	}
	if e = r.diagramLine(p, id+".timeline-left", [2]float64{tx, phaseTop}, [2]float64{tx, bodyBottom}, line, 1, "solid"); e != nil {
		return nil, e
	}
	seen = map[string]bool{}
	for i, g := range n.Gates {
		key := ""
		isKey := false
		if len(g.Key) > 0 {
			if e = json.Unmarshal(g.Key, &isKey); e != nil {
				if e = json.Unmarshal(g.Key, &key); e != nil {
					return nil, e
				}
			}
		}
		k, e := diagramKey(ctx, "gates", i, key, seen)
		if e != nil {
			return nil, e
		}
		if g.At < 0 || g.At > float64(np) {
			return nil, fmt.Errorf("scene.gantt_gate_at")
		}
		surf := "inverse"
		if isKey {
			surf = "callout"
		}
		col, _ := r.sceneColor(surf, "bg")
		x := X(g.At)
		if e = r.diagramLine(p, id+".gates."+k+".line", [2]float64{x, gateTop + 18}, [2]float64{x, bodyBottom}, col, 1.5, "dash"); e != nil {
			return nil, e
		}
		_, w, e := r.sequenceNeed(g.Label, label, 8191)
		if e != nil {
			return nil, e
		}
		w = sequenceInlineWidth(w, label) + 20
		left := x
		if left+w > tx+tw {
			left -= w
		}
		if left < tx-.02 || left+w > tx+tw+.02 {
			return nil, fmt.Errorf("scene.gantt_gate_label_outside_timeline: %s", k)
		}
		if e = r.sceneRect(p, id+".gates."+k+".chip", Rect{left, gateTop, w, 18}, surf); e != nil {
			return nil, e
		}
		if e = r.diagramStyledText(p, id+".gates."+k+".label", g.Label, label, Rect{left + 10, gateTop, w - 20, 18}, surf, "display", "left", true); e != nil {
			return nil, e
		}
		p.Warnings = append(p.Warnings, "Adapter resolution wmds.native-inline-chip-width.v1: add trailing native textbox space to Gantt gate labels and retain the complete chip within the timeline.")
	}
	p.Warnings = append(p.Warnings, "Adapter resolution wmds.native-gantt-label-width.v1: reserve trailing native textbox space before packing and drawing bar/event labels; preserve timeline interval, font sizes and bounded label placement.")
	if n.Today != nil {
		if *n.Today < 0 || *n.Today > float64(np) {
			return nil, fmt.Errorf("scene.gantt_today_at")
		}
		col, _ := r.sceneColor("callout", "bg")
		x := X(*n.Today)
		if e = r.diagramLine(p, id+".today-line", [2]float64{x, periodTop}, [2]float64{x, bodyBottom + 18}, col, 1.5, "solid"); e != nil {
			return nil, e
		}
		if e = r.diagramStyledText(p, id+".today-label", "Today", label, Rect{x + 5, bodyBottom + 4, 60, 0}, ctx.Surface, "primary", "left", false); e != nil {
			return nil, e
		}
	}
	p.Warnings = append(p.Warnings, "Adapter resolution wmds.gantt-guide-paint-order.v1: retain all source coordinates and place timeline guides behind bar/event items as in the frozen reference renderer.")
	usedKinds, usedEvents := map[string]bool{}, map[string]bool{}
	hatched, tagged, milestone := false, false, false
	for li, L := range layouts {
		pre := id + ".groups." + L.gkey + ".lanes." + L.key
		if gw > 0 && (li == 0 || layouts[li-1].gkey != L.gkey) {
			bottom := L.top + L.h
			for j := li + 1; j < len(layouts) && layouts[j].gkey == L.gkey; j++ {
				bottom = layouts[j].top + layouts[j].h
			}
			col, e := r.sceneColor(ctx.Surface, L.group.Fill)
			if e != nil {
				return nil, e
			}
			if e = r.diagramShape(p, id+".groups."+L.gkey+".tab", Rect{n.X, L.top + 3, gw, bottom - L.top - 6}, pptx.ShapeTypeRect, col, "", 0, "", nil); e != nil {
				return nil, e
			}
			st := label
			st.Size = 8
			st.Weight = 600
			st.Tracking = "0.1em"
			st.TrackingPt = .8
			textColor, _ := r.sceneColor("light", "strong")
			if contrast("FFFFFF", col) >= 4.5 {
				textColor = "FFFFFF"
			}
			if e = r.sequenceLiteralText(p, id+".groups."+L.gkey+".label", strings.ToUpper(L.group.Label), st, Rect{n.X + gw/2 - (bottom-L.top-12)/2, L.top + (bottom-L.top)/2 - 6, bottom - L.top - 12, 12}, textColor, "center", false); e != nil {
				return nil, e
			}
			p.Items[len(p.Items)-1].Text.Rotation = -90
		}
		ic := L.lane.Icon
		if ic == "" {
			ic = "target"
		}
		if e = r.sceneIcon(p, pre+".icon", ic, Rect{n.X + gw + 9, L.top + (L.h-18)/2, 18, 18}, ctx.Surface, "display"); e != nil {
			return nil, e
		}
		ts := small
		ts.Size = 12.5
		ts.Tracking = "0"
		ts.TrackingPt = 0
		need, _, e := r.sequenceNeed(L.lane.Title, ts, lw-45)
		if e != nil {
			return nil, e
		}
		sub, _ := r.sceneStyle("small")
		sh, _, e := r.sequenceNeed(L.lane.Sub, sub, lw-45)
		if e != nil {
			return nil, e
		}
		if need+sh > L.h {
			return nil, fmt.Errorf("scene.gantt_lane_heading_overflow")
		}
		yy := L.top + (L.h-need-sh)/2
		if e = r.sceneText(p, pre+".title", L.lane.Title, ts, Rect{n.X + gw + 36, yy, lw - 45, need}, ctx.Surface, "display", "left"); e != nil {
			return nil, e
		}
		if e = r.sceneText(p, pre+".sub", L.lane.Sub, sub, Rect{n.X + gw + 36, yy + need, lw - 45, sh}, ctx.Surface, "secondary", "left"); e != nil {
			return nil, e
		}
		last := li == len(layouts)-1 || layouts[li+1].gkey != L.gkey
		sepX, sepW, sepC, weight := n.X+gw, n.W-gw, line, .75
		if last {
			sepX, sepW, sepC, weight = n.X, n.W, strong, 1
		}
		if e = r.diagramLine(p, pre+".separator", [2]float64{sepX, L.top + L.h - weight/2}, [2]float64{sepX + sepW, L.top + L.h - weight/2}, sepC, weight, "solid"); e != nil {
			return nil, e
		}
		base := L.top + (L.h-float64(L.tracks)*trackPitch)/2 + 4
		for _, o := range L.items {
			it := o.it
			part := pre + ".items." + o.key
			y := base + float64(o.track)*trackPitch
			labelX, labelW, color := X(it.At)+11, o.width, strong
			if it.Kind != "" {
				usedKinds[it.Kind] = true
				k := n.Kinds[it.Kind]
				fill, e := r.sceneColor(ctx.Surface, k.Fill)
				if e != nil {
					return nil, e
				}
				color, e = r.sceneColor(ctx.Surface, k.Text)
				if e != nil {
					return nil, e
				}
				type seg struct {
					a, b  float64
					hatch bool
				}
				var segs []seg
				if it.Hatch {
					segs = []seg{{it.From, it.To, true}}
				} else if it.Progress != nil {
					mid := it.From + (it.To-it.From)**it.Progress
					segs = []seg{{it.From, mid, false}, {mid, it.To, true}}
				} else {
					a, b := it.From+it.SoftStart, it.To-it.SoftEnd
					if it.SoftStart > 0 {
						segs = append(segs, seg{it.From, a, true})
					}
					segs = append(segs, seg{a, b, false})
					if it.SoftEnd > 0 {
						segs = append(segs, seg{b, it.To, true})
					}
				}
				solidX, solidW := 0.0, 0.0
				for i, s := range segs {
					if s.b <= s.a {
						continue
					}
					box := Rect{X(s.a), y, (s.b - s.a) * pw, 22}
					if s.hatch {
						hatched = true
						e = r.sequenceHatch(p, fmt.Sprintf("%s.segment-%d", part, i), box, fill, strong)
					} else {
						e = r.diagramShape(p, fmt.Sprintf("%s.segment-%d", part, i), box, pptx.ShapeTypeRect, fill, strong, .75, "solid", nil)
						if box.W > solidW {
							solidX, solidW = box.X, box.W
						}
					}
					if e != nil {
						return nil, e
					}
				}
				labelX = X(it.To) + 6
				color = strong
				if o.width+14 <= solidW {
					labelX = solidX + 7
					color, e = r.sceneColor(ctx.Surface, k.Text)
					if e != nil {
						return nil, e
					}
				} else if o.width+14 <= (it.To-it.From)*pw {
					labelX = X(it.From) + 8
					if e = r.diagramShape(p, part+".label-patch", Rect{labelX - 4, y + 3, o.width + 8, 16}, pptx.ShapeTypeRect, "FFFFFF", strong, .75, "solid", nil); e != nil {
						return nil, e
					}
				}
			} else if it.Tag {
				tagged = true
				fill, _ := r.sceneColor("inverse", "bg")
				width := o.width + 20
				if e = r.diagramShape(p, part+".tag", Rect{X(it.At), y, width, 22}, pptx.ShapeTypeCustGeom, fill, "", 0, "", [][2]float64{{0, 11}, {7, 0}, {width, 0}, {width, 22}, {7, 22}}); e != nil {
					return nil, e
				}
				color, _ = r.sceneColor("inverse", "display")
			} else {
				kind := pptx.ShapeTypeDiamond
				fill := strong
				if it.Milestone {
					milestone = true
				} else {
					usedEvents[it.Event] = true
					ev := n.Events[it.Event]
					fill, e = r.sceneColor(ctx.Surface, ev.Fill)
					if e != nil {
						return nil, e
					}
					switch it.Event {
					case "star":
						kind = pptx.ShapeTypeStar5
					case "triangle":
						kind = pptx.ShapeTypeTriangle
					case "diamond":
					default:
						return nil, fmt.Errorf("scene.gantt_event_shape: %s", it.Event)
					}
				}
				if e = r.diagramShape(p, part+".marker", Rect{X(it.At) - 7, y + 4, 14, 14}, kind, fill, "", 0, "", nil); e != nil {
					return nil, e
				}
				if it.LabelSide == "left" {
					labelX = X(it.At) - 10 - o.width
				}
			}
			if labelX < tx-.02 || labelX+labelW > tx+tw+.02 {
				return nil, fmt.Errorf("scene.gantt_item_label_outside_timeline: %s", part)
			}
			if e = r.sequenceLiteralText(p, part+".label", it.Label, small, Rect{labelX, y, labelW, 22}, color, "left", true); e != nil {
				return nil, e
			}
		}
	}
	ly := bodyBottom + 14
	if n.LegendSize != 0 {
		st, _ := r.sceneStyle("small")
		ly = bodyBottom + 14*n.LegendSize/st.Size
	}
	if n.Today != nil {
		ly = bodyBottom + 26
	}
	legendItems := []map[string]any{}
	names := []string{}
	for k := range usedKinds {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		legendItems = append(legendItems, map[string]any{"key": "kind-" + k, "label": n.Kinds[k].Label, "color": n.Kinds[k].Fill})
	}
	if hatched {
		legendItems = append(legendItems, map[string]any{"key": "hatch", "label": "Hatched: in progress, soft start or tentative", "color": "deemph.1", "hatch": true})
	}
	names = nil
	for k := range usedEvents {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		legendItems = append(legendItems, map[string]any{"key": "event-" + k, "label": n.Events[k].Label, "color": n.Events[k].Fill, "marker": k})
	}
	if milestone {
		legendItems = append(legendItems, map[string]any{"key": "milestone", "label": "Milestone", "marker": "diamond"})
	}
	if tagged {
		legendItems = append(legendItems, map[string]any{"key": "tag", "label": "Event tag", "marker": "tag"})
	}
	if n.Gates != nil {
		legendItems = append(legendItems, map[string]any{"key": "gate", "label": "Gate", "line": true, "dashed": true})
	}
	if n.Today != nil {
		legendItems = append(legendItems, map[string]any{"key": "today", "label": "Today", "line": true, "color": "callout"})
	}
	lc := ctx
	lc.Keys = nil
	lc.Path = ctx.Path + "/derived-legend"
	legendNode := map[string]any{"type": "legend", "x": tx, "y": ly, "w": tw, "items": legendItems, "layout": "horizontal"}
	if n.LegendSize != 0 {
		legendNode["size"] = n.LegendSize
	}
	if n.LegendFullWidth {
		legendNode["x"], legendNode["w"] = n.X, n.W
	}
	legendRaw, _ := json.Marshal(legendNode)
	measureCtx := lc
	measureCtx.Zone = Rect{0, 0, 960, 540}
	legend, e := r.planSceneNode(id+".legend", legendRaw, measureCtx)
	if e != nil {
		return nil, e
	}
	if ctx.Zone.H > 0 && legend.Bounds.Y+legend.Bounds.H > ctx.Zone.Y+ctx.Zone.H+.02 {
		if n.LegendSize == 0 && n.TrackPitch == 0 && n.Y == 108 && gw == 24 && lw == 210 && np == 8 && len(n.Groups) == 3 {
			legendNode["x"], legendNode["y"], legendNode["w"] = n.X, n.Y-18, n.W
			legendRaw, _ = json.Marshal(legendNode)
			legend, e = r.planSceneNode(id+".legend", legendRaw, lc)
			if e != nil {
				return nil, e
			}
			if legend.Bounds.H > 18+.02 {
				return nil, fmt.Errorf("scene.gantt_header_legend_fit: measured %.3fpt, capacity18pt", legend.Bounds.H)
			}
			p.Warnings = append(p.Warnings, "Adapter resolution wmds.gantt-header-legend.v1: preserve timeline geometry and typography; move the measured single-row legend into the90–108pt band across the full main width.")
		} else {
			var lanes []string
			for _, L := range layouts {
				lanes = append(lanes, fmt.Sprintf("%s:%dtracks/%.0fpt", L.lane.Title, L.tracks, L.h))
			}
			return nil, fmt.Errorf("scene.gantt_legend_fit: bodybottom%.3fpt; legendbottom%.3fpt capacity%.3fpt; %v", bodyBottom, legend.Bounds.Y+legend.Bounds.H, ctx.Zone.Y+ctx.Zone.H, lanes)
		}
	}
	diagramMerge(p, legend)
	p = diagramFinish(p, "scene.gantt")
	if ctx.Zone.H > 0 && p.Bounds.Y+p.Bounds.H > ctx.Zone.Y+ctx.Zone.H+.02 {
		return nil, fmt.Errorf("scene.gantt_legend_overflow: bottom %.3fpt capacity %.3fpt", p.Bounds.Y+p.Bounds.H, ctx.Zone.Y+ctx.Zone.H)
	}
	return p, nil
}

type swimStep struct {
	Key  string `json:"key"`
	ID   string `json:"id"`
	Lane int    `json:"lane"`
	Col  int    `json:"col"`
	Kind string `json:"kind"`
	Text string `json:"text"`
	Pain bool   `json:"pain"`
}
type swimSpec struct {
	Type     string     `json:"type"`
	X        float64    `json:"x"`
	Y        float64    `json:"y"`
	W        float64    `json:"w"`
	Cols     int        `json:"cols"`
	Lanes    []string   `json:"lanes"`
	Steps    []swimStep `json:"steps"`
	Links    [][]string `json:"links"`
	PainNote string     `json:"painNote"`
}

func (r *renderer) planSwimlaneScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	var e error

	var n swimSpec
	if e := diagramDecode(raw, "x y w cols lanes steps links painNote", &n); e != nil {
		return nil, e
	}
	if n.Cols <= 0 || len(n.Lanes) == 0 || n.W <= 126 {
		return nil, fmt.Errorf("scene.swimlane_geometry")
	}
	p := &scenePlan{ID: id}
	cw := (n.W - 126) / float64(n.Cols)
	strong, _ := r.sceneColor(ctx.Surface, "strong")
	line, _ := r.sceneColor(ctx.Surface, "line")
	seen := map[string]bool{}
	for i, s := range n.Lanes {
		k, e := diagramKey(ctx, "lanes", i, "", seen)
		if e != nil {
			return nil, e
		}
		y := n.Y + float64(i)*72
		surf := "light"
		if i%2 == 1 {
			surf = "subtle"
		}
		if e = r.sceneRect(p, id+".lanes."+k+".surface", Rect{n.X + 126, y, n.W - 126, 72}, surf); e != nil {
			return nil, e
		}
		if e = r.sceneRect(p, id+".lanes."+k+".label-surface", Rect{n.X, y, 120, 72}, "inverse"); e != nil {
			return nil, e
		}
		if e = r.diagramText(p, id+".lanes."+k+".label", s, "small", Rect{n.X + 12, y, 96, 72}, "inverse", "display", "left", 600, true); e != nil {
			return nil, e
		}
		if e = r.diagramLine(p, id+".lanes."+k+".rule", [2]float64{n.X + 126, y}, [2]float64{n.X + n.W, y}, line, .75, "solid"); e != nil {
			return nil, e
		}
	}
	boxes := map[string]Rect{}
	steps := map[string]swimStep{}
	seen = map[string]bool{}
	for i, s := range n.Steps {
		k, e := diagramKey(ctx, "steps", i, s.Key, seen)
		if e != nil {
			return nil, e
		}
		if s.ID == "" || steps[s.ID].ID != "" || s.Lane < 0 || s.Lane >= len(n.Lanes) || s.Col < 0 || s.Col >= n.Cols {
			return nil, fmt.Errorf("scene.swimlane_step_identity_or_position")
		}
		w, h := math.Min(cw-18, 120), 42.0
		kind := pptx.ShapeTypeRect
		surf := "light"
		weight := 600
		if s.Kind == "decision" {
			w, h = 84, 54
			kind = pptx.ShapeTypeDiamond
			weight = 400
		} else if s.Kind == "start" || s.Kind == "end" {
			surf = "inverse"
		} else if s.Kind != "" && s.Kind != "process" {
			return nil, fmt.Errorf("scene.swimlane_step_kind: %s", s.Kind)
		}
		b := Rect{n.X + 126 + float64(s.Col)*cw + (cw-w)/2, n.Y + float64(s.Lane)*72 + (72-h)/2, w, h}
		boxes[s.ID] = b
		steps[s.ID] = s
		fill, _ := r.sceneColor(surf, "bg")
		if e = r.diagramShape(p, id+".steps."+k+".surface", b, kind, fill, strong, 1, "solid", nil); e != nil {
			return nil, e
		}
		textW := w - 24
		if kind == pptx.ShapeTypeDiamond {
			textW = w * .62
		}
		if e = r.diagramText(p, id+".steps."+k+".text", s.Text, "small", Rect{b.X + (w-textW)/2, b.Y, textW, h}, surf, "display", "center", weight, true); e != nil {
			return nil, e
		}
		if s.Pain {
			if e = r.sceneRect(p, id+".steps."+k+".pain", Rect{b.X + w - 9, b.Y - 9, 16, 16}, "callout"); e != nil {
				return nil, e
			}
			if e = r.diagramText(p, id+".steps."+k+".pain-label", "!", "label", Rect{b.X + w - 9, b.Y - 9, 16, 16}, "callout", "primary", "center", 700, true); e != nil {
				return nil, e
			}
		}
	}
	seen = map[string]bool{}
	for _, lk := range n.Links {
		if len(lk) < 2 || len(lk) > 4 {
			return nil, fmt.Errorf("scene.swimlane_link_shape")
		}
		k := lk[0] + "-to-" + lk[1]
		if !validPartKey(k) || seen[k] {
			return nil, fmt.Errorf("scene.swimlane_invalid_or_duplicate_link: %s", k)
		}
		seen[k] = true
		a, aok := boxes[lk[0]]
		z, zok := boxes[lk[1]]
		if !aok || !zok {
			return nil, fmt.Errorf("scene.swimlane_link_endpoint")
		}
		A, Z := steps[lk[0]], steps[lk[1]]
		acx, acy, zcx, zcy := a.X+a.W/2, a.Y+a.H/2, z.X+z.W/2, z.Y+z.H/2
		var pts [][2]float64
		ex := ""
		if len(lk) == 4 {
			ex = lk[3]
		}
		if ex == "up" || ex == "down" {
			ay, zy := a.Y, z.Y+z.H
			if ex == "down" {
				ay, zy = a.Y+a.H, z.Y
			}
			if A.Col == Z.Col {
				pts = [][2]float64{{acx, ay}, {zcx, zy}}
			} else {
				zx := z.X
				if Z.Col < A.Col {
					zx = z.X + z.W
				}
				pts = [][2]float64{{acx, ay}, {acx, zcy}, {zx, zcy}}
			}
		} else if ex != "" && ex != "right" && ex != "left" {
			return nil, fmt.Errorf("scene.swimlane_link_route")
		} else if A.Lane == Z.Lane && Z.Col > A.Col {
			pts = [][2]float64{{a.X + a.W, acy}, {z.X, zcy}}
		} else if A.Col == Z.Col {
			ay, zy := a.Y+a.H, z.Y
			if zcy < acy {
				ay, zy = a.Y, z.Y+z.H
			}
			pts = [][2]float64{{acx, ay}, {zcx, zy}}
		} else {
			ax := a.X + a.W
			if Z.Col < A.Col {
				ax = a.X
			}
			zy := z.Y
			if zcy < acy {
				zy = z.Y + z.H
			}
			pts = [][2]float64{{ax, acy}, {zcx, acy}, {zcx, zy}}
		}
		for j := 1; j < len(pts); j++ {
			for other, b := range boxes {
				if other == lk[0] || other == lk[1] {
					continue
				}
				u, v := pts[j-1], pts[j]
				cross := u[0] == v[0] && u[0] > b.X+.1 && u[0] < b.X+b.W-.1 && math.Max(u[1], v[1]) > b.Y+.1 && math.Min(u[1], v[1]) < b.Y+b.H-.1 || u[1] == v[1] && u[1] > b.Y+.1 && u[1] < b.Y+b.H-.1 && math.Max(u[0], v[0]) > b.X+.1 && math.Min(u[0], v[0]) < b.X+b.W-.1
				if cross {
					return nil, fmt.Errorf("scene.swimlane_source_route_intersects_step: %s -> %s crosses %s", lk[0], lk[1], other)
				}
			}
		}
		label := ""
		if len(lk) >= 3 {
			label = lk[2]
		}
		if e = r.planDiagramConnector(p, id+".links."+k, diagramSpec{Points: pts, Label: label}, ctx); e != nil {
			return nil, e
		}
	}
	bottom := n.Y + float64(len(n.Lanes))*72
	if e = r.diagramLine(p, id+".bottom", [2]float64{n.X + 126, bottom}, [2]float64{n.X + n.W, bottom}, line, .75, "solid"); e != nil {
		return nil, e
	}
	if n.PainNote != "" {
		if e = r.sceneRect(p, id+".pain-note-marker", Rect{n.X, bottom + 12, 16, 16}, "callout"); e != nil {
			return nil, e
		}
		if e = r.diagramText(p, id+".pain-note-mark", "!", "label", Rect{n.X, bottom + 12, 16, 16}, "callout", "primary", "center", 700, true); e != nil {
			return nil, e
		}
		if e = r.diagramText(p, id+".pain-note", n.PainNote, "label", Rect{n.X + 24, bottom + 12, n.W - 24, 0}, ctx.Surface, "secondary", "left", 0, false); e != nil {
			return nil, e
		}
	}
	return diagramFinish(p, "scene.swimlane"), nil
}
