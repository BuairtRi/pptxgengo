package wmdesign

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"math"
)

type peopleLegendItem struct {
	Key       string   `json:"key"`
	Text      string   `json:"text"`
	Label     string   `json:"label"`
	Series    int      `json:"series"`
	Deemph    bool     `json:"deemph"`
	Swatch    string   `json:"swatch"`
	Dashed    bool     `json:"dashed"`
	Color     string   `json:"color"`
	Line      bool     `json:"line"`
	Marker    string   `json:"marker"`
	Hatch     bool     `json:"hatch"`
	Ink       string   `json:"ink,omitempty"`
	Heat      *float64 `json:"heat,omitempty"`
	HeatMax   *float64 `json:"heatMax,omitempty"`
	HeatMin   *float64 `json:"heatMin,omitempty"`
	HeatScale string   `json:"heatScale,omitempty"`
	Status    string   `json:"status,omitempty"`
}
type peopleOrg struct {
	Key      string      `json:"key"`
	Org      string      `json:"org"`
	Title    string      `json:"title"`
	Name     string      `json:"name"`
	Dotted   bool        `json:"dotted"`
	Children []peopleOrg `json:"children"`
}
type peopleTier struct {
	Key       string            `json:"key"`
	Name      string            `json:"name"`
	Cadence   string            `json:"cadence"`
	Members   [][]string        `json:"members"`
	Decisions []json.RawMessage `json:"decisions"`
}
type peopleSpec struct {
	Type      string             `json:"type"`
	X         float64            `json:"x"`
	Y         float64            `json:"y"`
	W         float64            `json:"w"`
	H         float64            `json:"h"`
	Size      float64            `json:"size"`
	Layout    string             `json:"layout"`
	Title     string             `json:"title"`
	Band      string             `json:"band"`
	Roles     []string           `json:"roles"`
	Surface   string             `json:"surface"`
	Edge      string             `json:"edge"`
	Meta      string             `json:"meta"`
	Photo     string             `json:"photo"`
	Focus     string             `json:"focus"`
	Grayscale bool               `json:"grayscale"`
	Initials  string             `json:"initials"`
	Name      string             `json:"name"`
	Role      string             `json:"role"`
	Org       string             `json:"org"`
	Items     []peopleLegendItem `json:"items"`
	Root      peopleOrg          `json:"root"`
	Tiers     []peopleTier       `json:"tiers"`
}

func peopleOrgColor(org string) (string, error) {
	switch org {
	case "wm":
		return "series.8", nil
	case "client":
		return "series.4", nil
	case "tbd":
		return "deemph.1", nil
	}
	return "", fmt.Errorf("scene.unknown_org: %s", org)
}
func (r *renderer) planPeopleScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	fields := map[string]string{"legend": "x y w title layout items size", "pod": "x y w title band roles", "role": "x y w h surface edge title meta", "person": "x y w size photo focus grayscale initials name role org", "orgchart": "x y w root", "governance": "x y w tiers"}
	allow, ok := fields[tag.Type]
	if !ok {
		return nil, false, nil
	}
	if tag.Type == "legend" && !isV6OrLaterLibrary(r.source.Revision) {
		var rawLegend struct {
			Items []map[string]json.RawMessage `json:"items"`
		}
		if err := json.Unmarshal(raw, &rawLegend); err != nil {
			return nil, true, err
		}
		for _, item := range rawLegend.Items {
			if _, present := item["heatMin"]; present {
				return nil, true, fmt.Errorf("scene.legend_heat_min_requires_v6")
			}
		}
	}
	var n peopleSpec
	if e := diagramDecode(raw, allow, &n); e != nil {
		return nil, true, e
	}
	p := &scenePlan{ID: id}
	strong, e := r.sceneColor(ctx.Surface, "strong")
	if e != nil {
		return nil, true, e
	}
	line, _ := r.sceneColor(ctx.Surface, "line")
	switch n.Type {
	case "legend":
		if n.W <= 0 || math.IsNaN(n.X+n.Y+n.W) || math.IsInf(n.X+n.Y+n.W, 0) || len(n.Items) > 100 {
			return nil, true, fmt.Errorf("scene.legend_geometry")
		}
		horiz := n.Layout == "horizontal"
		if n.Layout != "" && n.Layout != "vertical" && !horiz {
			return nil, true, fmt.Errorf("scene.legend_layout")
		}
		x, y := n.X, n.Y
		rowH := 0.0
		st, _ := r.sceneStyle("small")
		scale := 1.0
		if n.Size != 0 {
			if n.Size < 6 || n.Size > 24 || math.IsNaN(n.Size) || math.IsInf(n.Size, 0) {
				return nil, true, fmt.Errorf("scene.legend_size: expected6_to24pt")
			}
			scale = n.Size / st.Size
			st.Size, st.Leading, st.TrackingPt = n.Size, st.Leading*scale, st.TrackingPt*scale
		}
		labelGap, itemGap := 9*scale, 18*scale
		if isV5OrLaterLibrary(r.source.Revision) && n.Size == 0 && horiz {
			// Use the existing swatch gap for the native text reserve so a
			// tight one-row source legend retains its packing footprint.
			labelGap -= sequenceInlineWidth(1, st) - 1
		}
		if n.Title != "" {
			if e = r.diagramText(p, id+".title", n.Title, "label", Rect{x, y, n.W, 0}, ctx.Surface, "secondary", "left", 0, false); e != nil {
				return nil, true, e
			}
			ls, _ := r.sceneStyle("label")
			hh, ww, _ := r.sequenceNeed(n.Title, ls, n.W)
			if horiz {
				x += ww + itemGap
				rowH = hh
			} else {
				y += hh + labelGap
			}
		}
		seen := map[string]bool{}
		for i, it := range n.Items {
			k, e := diagramKey(ctx, "items", i, it.Key, seen)
			if e != nil {
				return nil, true, e
			}
			pre := id + ".items." + k
			copy := it.Text
			if copy == "" {
				copy = it.Label
			}
			var status sceneStatusSpec
			if it.Status != "" {
				var ok bool
				status, ok = sceneStatuses[it.Status]
				if !isV5OrLaterLibrary(r.source.Revision) || !ok {
					return nil, true, fmt.Errorf("scene.legend_invalid_status")
				}
				if copy == "" {
					copy = status.Label
				}
			}
			sw := 10.0
			if it.Heat != nil || it.Swatch == "line" || it.Line || it.Hatch {
				sw = 18
			}
			if it.Marker != "" && it.Heat == nil {
				sw = 12
			}
			if it.Swatch == "dot" && it.Heat == nil {
				sw = 8
			}
			if it.Status != "" {
				sw = 8
			}
			if it.Swatch != "" && it.Swatch != "line" && it.Swatch != "dot" && it.Swatch != "marker" {
				return nil, true, fmt.Errorf("scene.legend_swatch")
			}
			if it.Heat == nil && (it.HeatMin != nil || it.HeatMax != nil || it.HeatScale != "") {
				return nil, true, fmt.Errorf("scene.legend_heat_requires_value")
			}
			sw *= scale
			height, width, e := r.sequenceNeed(copy, st, n.W-sw-labelGap)
			if e != nil {
				return nil, true, e
			}
			if n.Size != 0 || isV5OrLaterLibrary(r.source.Revision) {
				width = sequenceInlineWidth(width, st)
			}
			total := sw + labelGap + width
			if total > n.W+.02 {
				return nil, true, fmt.Errorf("scene.legend_item_width")
			}
			if horiz && x+total > n.X+n.W+.02 {
				x = n.X
				y += rowH + labelGap
				rowH = 0
			}
			colRef := it.Ink
			if colRef == "" {
				colRef = it.Color
			}
			if colRef == "" {
				colRef = "strong"
				if it.Series > 0 {
					colRef = fmt.Sprintf("series.%d", it.Series)
				}
				if it.Deemph {
					colRef = "deemph.1"
				}
			}
			col, e := r.sceneColor(ctx.Surface, colRef)
			if e != nil {
				return nil, true, e
			}
			cy := y + height/2
			if it.Status != "" {
				fill, e := r.sceneColor(ctx.Surface, status.Ink)
				if e != nil {
					return nil, true, e
				}
				if e = r.diagramShape(p, pre+".swatch", Rect{x, cy - 4*scale, sw, 8 * scale}, pptx.ShapeTypeRect, fill, strong, .75*scale, "solid", nil); e != nil {
					return nil, true, e
				}
			} else if it.Heat != nil {
				fill, _, e := sceneHeatDomain(*it.Heat, it.HeatMin, it.HeatMax, it.HeatScale)
				if e != nil {
					return nil, true, e
				}
				if e = r.diagramShape(p, pre+".swatch", Rect{x, cy - 5*scale, sw, 10 * scale}, pptx.ShapeTypeRect, fill, "", 0, "", nil); e != nil {
					return nil, true, e
				}
				p.Warnings = append(p.Warnings, sceneHeatWarning(it.HeatMin))
			} else if it.Swatch == "dot" {
				if e = r.diagramShape(p, pre+".swatch", Rect{x, cy - 4*scale, sw, 8 * scale}, pptx.ShapeTypeEllipse, col, "", 0, "", nil); e != nil {
					return nil, true, e
				}
			} else if it.Swatch == "line" || it.Line {
				dash := "solid"
				if it.Dashed {
					dash = "dash"
				}
				if e = r.diagramLine(p, pre+".swatch", [2]float64{x, cy}, [2]float64{x + sw, cy}, col, 3*scale, dash); e != nil {
					return nil, true, e
				}
			} else if it.Hatch {
				if e = r.sequenceHatch(p, pre+".swatch", Rect{x, cy - 5*scale, sw, 10 * scale}, col, strong); e != nil {
					return nil, true, e
				}
			} else {
				kind := pptx.ShapeTypeRect
				if it.Swatch == "marker" || it.Marker == "diamond" {
					kind = pptx.ShapeTypeDiamond
				}
				if it.Marker == "triangle" {
					kind = pptx.ShapeTypeTriangle
				}
				if it.Marker == "star" {
					kind = pptx.ShapeTypeStar5
				}
				if it.Deemph {
					col, _ = r.sceneColor(ctx.Surface, "deemph.2")
				}
				if e = r.diagramShape(p, pre+".swatch", Rect{x, cy - sw/2, sw, sw}, kind, col, strong, .75*scale, "solid", nil); e != nil {
					return nil, true, e
				}
			}
			if e = r.sceneText(p, pre+".text", copy, st, Rect{x + sw + labelGap, y, width, height}, ctx.Surface, "primary", "left"); e != nil {
				return nil, true, e
			}
			if horiz {
				x += total + itemGap
				rowH = math.Max(rowH, height)
			} else {
				y += height + labelGap
			}
		}
	case "pod":
		h := 48 + float64(len(n.Roles))*42
		if e = r.diagramShape(p, id+".outline", Rect{n.X, n.Y, n.W, h}, pptx.ShapeTypeRect, "", line, 1, "solid", nil); e != nil {
			return nil, true, e
		}
		if e = r.sceneRect(p, id+".band", Rect{n.X, n.Y, n.W, 36}, n.Band); e != nil {
			return nil, true, e
		}
		if e = r.diagramText(p, id+".title", n.Title, "body", Rect{n.X + 12, n.Y, n.W - 24, 36}, n.Band, "display", "left", 600, true); e != nil {
			return nil, true, e
		}
		seen := map[string]bool{}
		for i, s := range n.Roles {
			k, e := diagramKey(ctx, "roles", i, "", seen)
			if e != nil {
				return nil, true, e
			}
			b := Rect{n.X + 9, n.Y + 45 + float64(i)*42, n.W - 18, 36}
			if e = r.sceneRect(p, id+".roles."+k+".surface", b, "subtle"); e != nil {
				return nil, true, e
			}
			if e = r.diagramText(p, id+".roles."+k+".text", s, "small", Rect{b.X + 12, b.Y, b.W - 24, b.H}, "subtle", "display", "left", 0, true); e != nil {
				return nil, true, e
			}
		}
	case "role":
		surf := n.Surface
		if surf == "" {
			surf = "subtle"
		}
		if e = r.sceneRect(p, id+".surface", Rect{n.X, n.Y, n.W, n.H}, surf); e != nil {
			return nil, true, e
		}
		edge := n.Edge
		if edge == "deemph" {
			edge = "deemph.1"
		}
		col, e := r.sceneColor(surf, edge)
		if e != nil {
			return nil, true, e
		}
		if e = r.diagramShape(p, id+".edge", Rect{n.X, n.Y, 6, n.H}, pptx.ShapeTypeRect, col, "", 0, "", nil); e != nil {
			return nil, true, e
		}
		body, _ := r.sceneStyle("body")
		body.Weight = 600
		small, _ := r.sceneStyle("small")
		a, _, e := r.sequenceNeed(n.Title, body, n.W-30)
		if e != nil {
			return nil, true, e
		}
		b, _, e := r.sequenceNeed(n.Meta, small, n.W-30)
		if e != nil {
			return nil, true, e
		}
		gap := 3.0
		if n.Meta == "" {
			gap = 0
		}
		if a+b+gap > n.H+.02 {
			return nil, true, fmt.Errorf("scene.role_stack_overflow")
		}
		yy := n.Y + (n.H-a-b-gap)/2
		if e = r.sceneText(p, id+".title", n.Title, body, Rect{n.X + 18, yy, n.W - 30, a}, surf, "display", "left"); e != nil {
			return nil, true, e
		}
		if e = r.sceneText(p, id+".meta", n.Meta, small, Rect{n.X + 18, yy + a + gap, n.W - 30, b}, surf, "secondary", "left"); e != nil {
			return nil, true, e
		}
	case "person":
		if n.Size <= 0 || n.W <= n.Size+12 {
			return nil, true, fmt.Errorf("scene.person_geometry")
		}
		ww := n.W - n.Size - 12
		body, _ := r.sceneStyle("body")
		body.Weight = 600
		small, _ := r.sceneStyle("small")
		label, _ := r.sceneStyle("label")
		nh, _, e := r.sequenceNeed(n.Name, body, ww)
		if e != nil {
			return nil, true, e
		}
		rh, _, e := r.sequenceNeed(n.Role, small, ww)
		if e != nil {
			return nil, true, e
		}
		oh, ow, e := r.sequenceNeed(n.Org, label, ww-8)
		if e != nil {
			return nil, true, e
		}
		th := nh + 3 + rh
		if n.Org != "" {
			th += 6 + oh
		}
		height := math.Max(th, n.Size)
		ib := Rect{n.X, n.Y + (height-n.Size)/2, n.Size, n.Size}
		if n.Photo != "" {
			focus := n.Focus
			if focus == "" {
				focus = "50% 30%"
			}
			im, e := r.primitiveMediaImage(id+".avatar", n.Photo, ib, focus, n.Grayscale)
			if e != nil {
				return nil, true, e
			}
			p.Items = append(p.Items, sceneItem{Image: im})
			p.Bounds = diagramUnion(p.Bounds, ib)
		} else {
			if e = r.diagramShape(p, id+".avatar", ib, pptx.ShapeTypeRect, strong, "", 0, "", nil); e != nil {
				return nil, true, e
			}
			if e = r.diagramText(p, id+".initials", n.Initials, "body", ib, "inverse", "display", "center", 600, true); e != nil {
				return nil, true, e
			}
		}
		xx, yy := n.X+n.Size+12, n.Y+(height-th)/2
		if e = r.sceneText(p, id+".name", n.Name, body, Rect{xx, yy, ww, nh}, ctx.Surface, "display", "left"); e != nil {
			return nil, true, e
		}
		if e = r.sceneText(p, id+".role", n.Role, small, Rect{xx, yy + nh + 3, ww, rh}, ctx.Surface, "secondary", "left"); e != nil {
			return nil, true, e
		}
		if n.Org != "" {
			b := Rect{xx, yy + nh + rh + 9, ow + 8, oh}
			if e = r.diagramShape(p, id+".org-outline", b, pptx.ShapeTypeRect, "", strong, .75, "solid", nil); e != nil {
				return nil, true, e
			}
			if e = r.sceneText(p, id+".org", n.Org, label, Rect{b.X + 4, b.Y, b.W - 8, b.H}, ctx.Surface, "primary", "left"); e != nil {
				return nil, true, e
			}
		}
	case "orgchart":
		if e = r.planPeopleOrg(p, id, n, ctx); e != nil {
			return nil, true, e
		}
	case "governance":
		if e = r.planPeopleGovernance(p, id, n, ctx); e != nil {
			return nil, true, e
		}
	}
	return diagramFinish(p, "scene."+n.Type), true, nil
}
func peopleTreeWidth(n peopleOrg) float64 {
	if len(n.Children) == 0 {
		return 162
	}
	w := 18 * float64(len(n.Children)-1)
	for _, c := range n.Children {
		w += peopleTreeWidth(c)
	}
	return math.Max(162, w)
}
func (r *renderer) planPeopleOrg(p *scenePlan, id string, n peopleSpec, ctx SceneContext) error {
	total := peopleTreeWidth(n.Root)
	if total > n.W+.02 {
		return fmt.Errorf("scene.orgchart_width: requires %.3fpt capacity %.3fpt", total, n.W)
	}
	strong, _ := r.sceneColor(ctx.Surface, "strong")
	var place func(peopleOrg, float64, float64, *[2]float64, string, SceneContext) error
	place = func(nd peopleOrg, x, y float64, parent *[2]float64, pre string, cc SceneContext) error {
		w := peopleTreeWidth(nd)
		cx := x + w/2
		b := Rect{cx - 81, y, 162, 54}
		ref, e := peopleOrgColor(nd.Org)
		if e != nil {
			return e
		}
		edge, e := r.sceneColor(ctx.Surface, ref)
		if e != nil {
			return e
		}
		surf, ink, dash := "subtle", "display", "solid"
		outline := ""
		if nd.Org == "tbd" {
			surf, ink, dash = "light", "secondary", "dash"
			outline = edge
		}
		fill, _ := r.sceneColor(surf, "bg")
		if e = r.diagramShape(p, pre+".surface", b, pptx.ShapeTypeRect, fill, outline, 1, dash, nil); e != nil {
			return e
		}
		if e = r.diagramShape(p, pre+".edge", Rect{b.X, b.Y, 6, 54}, pptx.ShapeTypeRect, edge, "", 0, "", nil); e != nil {
			return e
		}
		st, _ := r.sceneStyle("small")
		st.Weight = 600
		ls, _ := r.sceneStyle("label")
		h, _, e := r.sequenceNeed(nd.Title, st, 132)
		if e != nil {
			return e
		}
		nh, _, e := r.sequenceNeed(nd.Name, ls, 132)
		if e != nil {
			return e
		}
		if h+nh > 54+.02 {
			return fmt.Errorf("scene.org_role_overflow: %s", pre)
		}
		yy := y + (54-h-nh)/2
		if e = r.sceneText(p, pre+".title", nd.Title, st, Rect{b.X + 18, yy, 132, h}, surf, ink, "left"); e != nil {
			return e
		}
		if e = r.sceneText(p, pre+".name", nd.Name, ls, Rect{b.X + 18, yy + h, 132, nh}, surf, "secondary", "left"); e != nil {
			return e
		}
		if parent != nil {
			dash := "solid"
			if nd.Dotted {
				dash = "dash"
			}
			pts := [][2]float64{*parent, {parent[0], y - 27}, {cx, y - 27}, {cx, y}}
			for j := 1; j < len(pts); j++ {
				if e = r.diagramLine(p, fmt.Sprintf("%s.reporting-%d", pre, j), pts[j-1], pts[j], strong, 1, dash); e != nil {
					return e
				}
			}
		}
		seen := map[string]bool{}
		xx := x
		for i, ch := range nd.Children {
			k, e := diagramKey(cc, "children", i, ch.Key, seen)
			if e != nil {
				return e
			}
			next := cc
			next.Path += fmt.Sprintf("/children/%d", i)
			pt := [2]float64{cx, y + 54}
			if e = place(ch, xx, y+108, &pt, pre+".children."+k, next); e != nil {
				return e
			}
			xx += peopleTreeWidth(ch) + 18
		}
		return nil
	}
	cc := ctx
	cc.Path += "/root"
	return place(n.Root, n.X+(n.W-total)/2, n.Y, nil, id+".root", cc)
}
func (r *renderer) planPeopleGovernance(p *scenePlan, id string, n peopleSpec, ctx SceneContext) error {
	var e error

	if len(n.Tiers) == 0 || n.W <= 624 {
		return fmt.Errorf("scene.governance_geometry")
	}
	strong, _ := r.sceneColor(ctx.Surface, "strong")
	seen := map[string]bool{}
	for i, t := range n.Tiers {
		k, e := diagramKey(ctx, "tiers", i, t.Key, seen)
		if e != nil {
			return e
		}
		pre := id + ".tiers." + k
		y := n.Y + float64(i)*90
		surf := "subtle"
		if i == 0 {
			surf = "inverse"
		} else if i == 1 {
			surf = "deep"
		}
		if e = r.sceneRect(p, pre+".surface", Rect{n.X, y, n.W - 48, 78}, surf); e != nil {
			return e
		}
		if _, e = r.sequenceStack(p, pre+".name", []string{t.Name, t.Cadence}, []string{"subhead", "label"}, []int{0, 0}, n.X+18, y+12, 198, 3, 54, surf, []string{"display", "emphasis"}); e != nil {
			return e
		}
		ls, _ := r.sceneStyle("label")
		ls.Weight = 600
		mx, my, rowH := n.X+234, y+12, 0.0
		cc := ctx
		cc.Path += fmt.Sprintf("/tiers/%d", i)
		mseen := map[string]bool{}
		for j, m := range t.Members {
			if len(m) != 2 {
				return fmt.Errorf("scene.governance_member_shape")
			}
			mk, e := diagramKey(cc, "members", j, "", mseen)
			if e != nil {
				return e
			}
			ref, e := peopleOrgColor(m[1])
			if e != nil {
				return e
			}
			edge, e := r.sceneColor(ctx.Surface, ref)
			if e != nil {
				return e
			}
			// Reserve the native trailing space before shaping so a long member
			// can own measured multiline height without exceeding the306pt lane.
			memberW := 306.0
			guard := sequenceInlineWidth(1, ls) - 1
			h, w, e := r.sequenceNeed(m[0], ls, memberW-16-guard)
			if e != nil {
				return e
			}
			w = sequenceInlineWidth(w, ls) + 16
			h += 6
			if w > memberW+.02 {
				return fmt.Errorf("scene.governance_member_width_overflow: %s", mk)
			}
			if mx+w > n.X+540 {
				mx = n.X + 234
				my += rowH + 6
				rowH = 0
			}
			if my+h > y+66+.02 {
				return fmt.Errorf("scene.governance_members_overflow")
			}
			b := Rect{mx, my, w, h}
			if e = r.sceneRect(p, pre+".members."+mk+".surface", b, "light"); e != nil {
				return e
			}
			if e = r.diagramShape(p, pre+".members."+mk+".edge", Rect{mx, my, 4, h}, pptx.ShapeTypeRect, edge, "", 0, "", nil); e != nil {
				return e
			}
			if e = r.sceneText(p, pre+".members."+mk+".text", m[0], ls, Rect{mx + 10, my + 3, w - 16, h - 6}, "light", "primary", "left"); e != nil {
				return e
			}
			mx += w + 6
			rowH = math.Max(rowH, h)
		}
		p.Warnings = append(p.Warnings, "Adapter resolution wmds.native-inline-chip-width.v1: add trailing native textbox space to member labels; preserve source font sizes and bound the complete chip rows within the governance member lane.")
		dw := n.W - 624
		if e = r.diagramText(p, pre+".decides-label", "Decides", "label", Rect{n.X + 558, y + 12, dw, 0}, surf, "secondary", "left", 0, false); e != nil {
			return e
		}
		label, _ := r.sceneStyle("label")
		lh, _, _ := r.sequenceNeed("Decides", label, dw)
		bc := ctx
		bc.Surface = surf
		bottom, e := r.sequenceBullet(p, pre+".decisions", t.Decisions, n.X+558, y+15+lh, dw, bc, cc.Path+"/decisions")
		if e != nil {
			return e
		}
		if bottom > y+66+.02 {
			return fmt.Errorf("scene.governance_decisions_overflow")
		}
		if i > 0 {
			x := n.X + n.W - 18
			pts := [][2]float64{{x, y - 12 + 16}, {x, y - 12 - 2}, {x - 6, y - 12 + 3}, {x, y - 12 - 3}, {x + 6, y - 12 + 3}}
			for j := 1; j < len(pts); j++ {
				if e = r.diagramLine(p, fmt.Sprintf("%s.escalation-%d", pre, j), pts[j-1], pts[j], strong, 1.4, "solid"); e != nil {
					return e
				}
			}
		}
	}
	height := float64(len(n.Tiers))*90 - 20
	ls, _ := r.sceneStyle("label")
	ls.Weight = 600
	need, _, e := r.sequenceNeed("Escalation", ls, height)
	if e != nil {
		return e
	}
	b := Rect{n.X + n.W - 22 - height/2, n.Y + 4 + height/2 - need/2, height, need}
	p.Warnings = append(p.Warnings, "Adapter resolution wmds.governance-rotation-envelope.v1: use measured minor-axis text bounds for uniform native rotation.")
	if e = r.sceneText(p, id+".escalation-label", "Escalation", ls, b, ctx.Surface, "secondary", "center"); e != nil {
		return e
	}
	p.Items[len(p.Items)-1].Text.Rotation = -90
	return nil
}
