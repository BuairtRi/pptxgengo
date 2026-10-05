package wmdesign

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"math"
	"strings"
)

// Source-scene planners preserve frozen local geometry. Context exceptions are
// explicit here; they do not relax the foundation's general outer-box contract.
func diagramDecode(raw json.RawMessage, fields string, out any) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return err
	}
	allowed := map[string]bool{}
	for _, s := range strings.Fields("type " + fields) {
		allowed[s] = true
	}
	for k := range obj {
		if !allowed[k] {
			return fmt.Errorf("scene.unsupported_field: %s", k)
		}
	}
	return sceneDecode(raw, out)
}
func diagramKey(ctx SceneContext, field string, index int, explicit string, seen map[string]bool) (string, error) {
	key := explicit
	path := strings.TrimRight(ctx.Path, "/") + "/" + field
	if supplied, ok := ctx.Keys[path]; ok {
		if index >= len(supplied) {
			return "", fmt.Errorf("scene.missing_array_key: %s[%d]", path, index)
		}
		if key != "" && key != supplied[index] {
			return "", fmt.Errorf("scene.conflicting_array_key: %s", path)
		}
		key = supplied[index]
	} else if ctx.Keys != nil {
		return "", fmt.Errorf("scene.missing_array_keys: %s", path)
	}
	if key == "" {
		key = fmt.Sprintf("slot-%03d", index+1)
	}
	if !validPartKey(key) || seen[key] {
		return "", fmt.Errorf("scene.invalid_or_duplicate_key: %s", key)
	}
	seen[key] = true
	return key, nil
}
func diagramRect(x, y, w, h float64) Rect { return Rect{x, y, w, h} }
func diagramUnion(a, b Rect) Rect {
	if a.W == 0 && a.H == 0 {
		return b
	}
	x, y := math.Min(a.X, b.X), math.Min(a.Y, b.Y)
	return Rect{x, y, math.Max(a.X+a.W, b.X+b.W) - x, math.Max(a.Y+a.H, b.Y+b.H) - y}
}
func diagramRotatedRect(b Rect, deg float64) Rect {
	if deg == 0 {
		return b
	}
	a := deg * math.Pi / 180
	w := math.Abs(b.W*math.Cos(a)) + math.Abs(b.H*math.Sin(a))
	h := math.Abs(b.W*math.Sin(a)) + math.Abs(b.H*math.Cos(a))
	return Rect{b.X + (b.W-w)/2, b.Y + (b.H-h)/2, w, h}
}
func diagramMerge(p, c *scenePlan) {
	p.Items = append(p.Items, c.Items...)
	p.Groups = append(p.Groups, c.Groups...)
	p.Warnings = append(p.Warnings, c.Warnings...)
	p.Bounds = diagramUnion(p.Bounds, c.Bounds)
}
func diagramFinish(p *scenePlan, definition string) *scenePlan {
	var parts []string
	var shapes []ShapeRecord
	for _, it := range p.Items {
		if it.Shape != nil {
			parts = append(parts, it.Shape.Record.ID)
			shapes = append(shapes, it.Shape.Record)
			p.Bounds = diagramUnion(p.Bounds, it.Shape.Record.Rect)
		}
		if it.Text != nil {
			parts = append(parts, it.Text.ID)
			p.Bounds = diagramUnion(p.Bounds, diagramRotatedRect(it.Text.Rect, it.Text.Rotation))
		}
		if it.Image != nil {
			parts = append(parts, it.Image.ObjectName)
		}
		if it.Table != nil {
			parts = append(parts, it.Table.ID)
			p.Bounds = diagramUnion(p.Bounds, it.Table.Rect)
		}
		if it.Chart != nil {
			parts = append(parts, it.Chart.ID)
			p.Bounds = diagramUnion(p.Bounds, it.Chart.Rect)
		}
	}
	// Bottom-up existing child groups replace their owned flat native parts in
	// the root sequence. Parent serialization validates actual contiguity.
	for _, g := range p.Groups {
		owned := map[string]bool{}
		for _, id := range g.Parts {
			owned[id] = true
		}
		var next []string
		added := false
		for _, id := range parts {
			if owned[id] {
				if !added {
					next = append(next, g.ID)
					added = true
				}
			} else {
				next = append(next, id)
			}
		}
		if added {
			parts = next
		}
	}
	if len(parts) > 0 {
		p.Groups = append(p.Groups, ComponentRecord{ID: p.ID, Definition: definition, Contract: "pptxgengo.wmds-source-scenes.v1", Rect: p.Bounds, RequiredHeight: p.Bounds.H, Density: "source_context", Parts: parts, Shapes: shapes})
	}
	return p
}
func (r *renderer) diagramShape(p *scenePlan, id string, b Rect, kind pptx.ShapeType, fill, line string, weight float64, dash string, points [][2]float64) error {
	if b.W < 0 || b.H < 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
		return fmt.Errorf("scene.invalid_shape_geometry: %s", id)
	}
	props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}
	if fill != "" {
		props.Fill = &pptx.ShapeFillProps{Color: fill}
	}
	if line != "" && weight > 0 {
		props.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: line}, Width: weight, DashType: dash}
	}
	for i, xy := range points {
		mv := i == 0
		props.Points = append(props.Points, pptx.ShapePoint{X: pptx.Inches(xy[0] / 72), Y: pptx.Inches(xy[1] / 72), MoveTo: &mv})
	}
	if len(points) > 0 && fill != "" {
		props.Points = append(props.Points, pptx.ShapePoint{Close: true})
	}
	rec := ShapeRecord{ID: id, Rect: b, Color: fill, Geometry: string(kind)}
	if fill == "" {
		rec.Color = line
	}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: kind, Props: props, Record: rec}})
	p.Bounds = diagramUnion(p.Bounds, b)
	return nil
}
func (r *renderer) diagramColor(surface, ink string) (string, error) {
	return r.sceneColor(surface, ink)
}
func (r *renderer) diagramText(p *scenePlan, id, text, token string, b Rect, surface, ink, align string, weight int, middle bool) error {
	st, e := r.sceneStyle(token)
	if e != nil {
		return e
	}
	if weight != 0 {
		st.Weight = weight
	}
	return r.diagramStyledText(p, id, text, st, b, surface, ink, align, middle)
}
func (r *renderer) diagramStyledText(p *scenePlan, id, text string, st Style, b Rect, surface, ink, align string, middle bool) error {
	if text == "" {
		return nil
	}
	if middle && b.H > 0 {
		l, e := r.typeEngine.Measure(text, st, b.W)
		if e != nil {
			return e
		}
		need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
		if need > b.H+.02 {
			return fmt.Errorf("scene.text_overflow: %s needs %.3fpt, capacity %.3fpt", id, need, b.H)
		}
		b.Y += (b.H - need) / 2
		b.H = need
	}
	return r.sceneText(p, id, text, st, b, surface, ink, align)
}
func (r *renderer) diagramPolygon(p *scenePlan, id string, b Rect, surface, ink string, pts [][2]float64) error {
	fill, e := r.sceneColor(surface, "bg")
	if e != nil {
		return e
	}
	line := ""
	if ink != "" {
		line, e = r.sceneColor(surface, ink)
		if e != nil {
			return e
		}
	}
	return r.diagramShape(p, id, b, pptx.ShapeTypeCustGeom, fill, line, .75, "solid", pts)
}
func (r *renderer) diagramLine(p *scenePlan, id string, a, b [2]float64, color string, weight float64, dash string) error {
	box := Rect{math.Min(a[0], b[0]), math.Min(a[1], b[1]), math.Abs(a[0] - b[0]), math.Abs(a[1] - b[1])}
	pts := [][2]float64{{a[0] - box.X, a[1] - box.Y}, {b[0] - box.X, b[1] - box.Y}}
	return r.diagramShape(p, id, box, pptx.ShapeTypeCustGeom, "", color, weight, dash, pts)
}
func (r *renderer) diagramChild(p *scenePlan, id string, node any, ctx SceneContext) error {
	raw, e := json.Marshal(node)
	if e != nil {
		return e
	}
	child, e := r.planSceneNode(id, raw, ctx)
	if e != nil {
		return e
	}
	diagramMerge(p, child)
	return nil
}

type diagramSpec struct {
	Highlight string            `json:"highlight"`
	Type      string            `json:"type"`
	SourceID  string            `json:"id,omitempty"`
	X         float64           `json:"x"`
	Y         float64           `json:"y"`
	W         float64           `json:"w"`
	H         float64           `json:"h"`
	Style     string            `json:"style"`
	Surface   string            `json:"surface"`
	Text      string            `json:"text"`
	Sub       string            `json:"sub"`
	Align     string            `json:"align"`
	First     bool              `json:"first"`
	Number    string            `json:"number"`
	Dir       string            `json:"dir"`
	Label     string            `json:"label"`
	LabelPos  string            `json:"labelPos"`
	Bullets   json.RawMessage   `json:"bullets"`
	Icon      string            `json:"icon"`
	Layout    string            `json:"layout"`
	N         string            `json:"n"`
	Points    [][2]float64      `json:"points"`
	Head      string            `json:"head"`
	Elbow     string            `json:"elbow"`
	Ink       string            `json:"ink"`
	Dashed    bool              `json:"dashed"`
	StartDot  bool              `json:"startDot"`
	Labels    string            `json:"labels"`
	LabelW    float64           `json:"labelW"`
	Cols      []string          `json:"cols"`
	CellH     float64           `json:"cellH"`
	Gap       *float64          `json:"gap"`
	RowGap    *float64          `json:"rowGap"`
	Rows      []json.RawMessage `json:"rows"`
	Left      string            `json:"left"`
	Right     string            `json:"right"`
	Arrow     string            `json:"arrow,omitempty"`
	Heat      *float64          `json:"heat,omitempty"`
	HeatMax   *float64          `json:"heatMax,omitempty"`
	HeatMin   *float64          `json:"heatMin,omitempty"`
	HeatScale string            `json:"heatScale,omitempty"`
}

func (r *renderer) planDiagramScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	fields := map[string]string{"block": "x y w h surface text style align heat heatMin heatMax heatScale", "frame": "id x y w h style label", "chevron": "x y w h first surface text style number sub", "textarrow": "x y w h dir surface text style", "connector": "points label labelPos style head elbow ink startDot dashed", "container": "id x y w h style label labelPos bullets", "cylinder": "x y w h surface text sub", "node": "x y w h surface icon text sub layout style", "layerrow": "x y w h surface n label text highlight", "matrix": "x y w labels labelW cols cellH gap rowGap style rows", "beforeafter": "x y w left right rows arrow"}
	allowed, ok := fields[tag.Type]
	if !ok {
		return nil, false, nil
	}
	if tag.Type == "block" && r.source.Revision != LibraryRevisionV6 {
		var rawFields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &rawFields); err != nil {
			return nil, true, err
		}
		if _, present := rawFields["heatMin"]; present {
			return nil, true, fmt.Errorf("scene.block_heat_min_requires_v6")
		}
	}
	var n diagramSpec
	if e := diagramDecode(raw, allowed, &n); e != nil {
		return nil, true, e
	}
	p := &scenePlan{ID: id}
	b := Rect{n.X, n.Y, n.W, n.H}
	surface := n.Surface
	if surface == "" {
		surface = ctx.Surface
	}
	if n.Type == "cylinder" && n.Surface == "" {
		surface = "inverse"
	}
	var err error
	text := func(part, copy, token string, box Rect, surf, ink, align string, w int, mid bool) {
		if err == nil {
			err = r.diagramText(p, id+"."+part, copy, token, box, surf, ink, align, w, mid)
		}
	}
	rect := func(part string, box Rect, surf string) {
		if err == nil {
			err = r.sceneRect(p, id+"."+part, box, surf)
		}
	}
	switch n.Type {
	case "block":
		heatInk := ""
		if n.Heat != nil {
			fill, ink, e := sceneHeatDomain(*n.Heat, n.HeatMin, n.HeatMax, n.HeatScale)
			if e != nil {
				return nil, true, e
			}
			heatInk = ink
			err = r.diagramShape(p, id+".surface", b, pptx.ShapeTypeRect, fill, "", 0, "", nil)
			p.Warnings = append(p.Warnings, sceneHeatWarning(n.HeatMin))
		} else {
			if n.HeatMin != nil || n.HeatMax != nil || n.HeatScale != "" {
				return nil, true, fmt.Errorf("scene.block_heat_requires_value")
			}
			rect("surface", b, surface)
		}
		token := n.Style
		if token == "" {
			token = "body"
		}
		a := n.Align
		if a == "" {
			a = "center"
		}
		pad := 12.0
		if isV5OrLaterLibrary(r.source.Revision) {
			st, e := r.sceneStyle(token)
			if e != nil {
				return nil, true, e
			}
			pad, _, e = r.v5WordInsets(n.Text, st, b.W, pad, pad)
			if e != nil {
				return nil, true, e
			}
		}
		incomingCompact := isExpandedLibrary(r.source.Revision) && ((token == "number" && b.W <= 54) || (token == "small" && b.W <= 72))
		// Retain the compact small/label policy established by the preceding
		// intake wave; only number badges and the72pt small block are new here.
		if incomingCompact || (token == "label" || token == "small") && (b.H <= 18 || b.W <= 54) {
			pad = math.Min(6, b.W/6)
			p.Warnings = append(p.Warnings, "Adapter resolution wmds.compact-label-block.v1: preserve native shape/font sizes; compact label padding is6pt (3pt for18pt badges).")
		}
		text("text", n.Text, token, Rect{b.X + pad, b.Y, b.W - 2*pad, b.H}, surface, "display", a, 0, true)
		if heatInk != "" {
			for i := range p.Items {
				if p.Items[i].Text != nil {
					p.Items[i].Text.Color = heatInk
					if p.Items[i].Text.Rich != nil {
						for j := range p.Items[i].Text.Rich.Paragraphs {
							for k := range p.Items[i].Text.Rich.Paragraphs[j].Runs {
								p.Items[i].Text.Rich.Paragraphs[j].Runs[k].Color = heatInk
							}
						}
					}
				}
			}
		}
	case "frame", "container":
		fill, line, dash, weight := "", "", "solid", 1.0
		if n.Type == "frame" {
			switch n.Style {
			case "solid":
				line, err = r.sceneColor(ctx.Surface, "strong")
			case "dashed":
				line, err = r.sceneColor(ctx.Surface, "secondary")
				dash = "dash"
			case "filled":
				fill, err = r.sceneColor("subtle", "bg")
			default:
				return nil, true, fmt.Errorf("scene.unknown_frame_style: %s", n.Style)
			}
		} else {
			switch n.Style {
			case "region":
				fill, err = r.sceneColor("subtle", "bg")
			case "boundary":
				line, err = r.sceneColor("light", "strong")
				dash = "dash"
			case "layer":
				fill, err = r.sceneColor("light", "bg")
				line, _ = r.sceneColor("light", "line")
			case "frame":
				line, err = r.sceneColor("light", "strong")
			case "external":
				fill, err = r.sceneColor("light", "bg")
				line, _ = r.sceneColor("light", "secondary")
				dash = "dash"
			default:
				return nil, true, fmt.Errorf("scene.unknown_container_style: %s", n.Style)
			}
		}
		if err == nil {
			err = r.diagramShape(p, id+".surface", b, pptx.ShapeTypeRect, fill, line, weight, dash, nil)
		}
		if n.Label != "" {
			if n.Style == "frame" || n.Type == "frame" && n.Style != "dashed" {
				st, e := r.sceneStyle("label")
				if e != nil {
					return nil, true, e
				}
				st.Weight = 600
				l, e := r.typeEngine.Measure(n.Label, st, math.Max(1, b.W-12))
				if e != nil {
					return nil, true, e
				}
				w := sequenceInlineWidth(l.Lines[0].Advance, st) + 12
				if len(l.Lines) != 1 || w > b.W+.02 {
					return nil, true, fmt.Errorf("scene.frame_tab_overflow: %s", id)
				}
				rect("tab", Rect{b.X, b.Y, w, 18}, "inverse")
				text("label", n.Label, "label", Rect{b.X + 6, b.Y, w - 12, 18}, "inverse", "display", "left", 600, true)
				p.Warnings = append(p.Warnings, "Adapter resolution wmds.inline-sequence-text.v1: retain a single-line frame tab with trailing native textbox space.")
			} else {
				token := "body"
				if n.Style == "region" {
					token = "subhead"
				}
				if n.Style == "external" {
					token = "small"
				}
				if n.Type == "frame" {
					token = "label"
				}
				a := "left"
				if n.LabelPos == "tr" {
					a = "right"
				} else if n.LabelPos != "" {
					return nil, true, fmt.Errorf("scene.unknown_label_position")
				}
				yb := b.Y + 9
				if n.Type == "frame" {
					yb = b.Y - 9
				}
				if n.Type == "frame" && n.Style == "dashed" {
					st, _ := r.sceneStyle(token)
					st.Weight = 600
					h, w, e := r.sequenceNeed(n.Label, st, b.W-24)
					if e != nil {
						return nil, true, e
					}
					rect("label-patch", Rect{b.X + 9, yb, w + 6, h}, ctx.Surface)
				}
				text("label", n.Label, token, Rect{b.X + 12, yb, b.W - 24, 0}, "light", "display", a, 600, false)
				if n.Style == "external" && err == nil {
					st, _ := r.sceneStyle(token)
					st.Weight = 600
					l, e := r.typeEngine.Measure(n.Label, st, b.W-24)
					if e != nil {
						return nil, true, e
					}
					col, _ := r.sceneColor("light", "display")
					for i, ln := range l.Lines {
						x := b.X + 12
						if a == "right" {
							x += b.W - 24 - ln.Advance
						}
						if err = r.diagramLine(p, fmt.Sprintf("%s.label-underline-%d", id, i), [2]float64{x, yb + ln.Baseline + 1.2}, [2]float64{x + ln.Advance, yb + ln.Baseline + 1.2}, col, .75, "solid"); err != nil {
							return nil, true, err
						}
					}
				}
			}
		}
		if len(n.Bullets) > 0 && string(n.Bullets) != "null" {
			var items any
			if e := json.Unmarshal(n.Bullets, &items); e != nil {
				return nil, true, e
			}
			childCtx := ctx
			childCtx.Path = ctx.Path + "/bullets"
			if keys, ok := ctx.Keys[childCtx.Path]; ok {
				childCtx.Keys = make(map[string][]string, len(ctx.Keys)+1)
				for k, v := range ctx.Keys {
					childCtx.Keys[k] = v
				}
				childCtx.Keys[childCtx.Path+"/items"] = keys
			}
			if err == nil {
				err = r.diagramChild(p, id+".bullets", map[string]any{"type": "bullets", "x": b.X + 12, "y": b.Y + 33, "w": b.W - 24, "size": "small", "items": items}, childCtx)
			}
		}
	case "chevron", "textarrow":
		var pts [][2]float64
		tx := b
		token := n.Style
		if token == "" {
			token = "body"
			if n.Type == "chevron" && (n.Number != "" || n.Sub != "") {
				token = "small"
			}
		}
		if n.Type == "chevron" {
			d := math.Round(b.H / 3)
			if n.First {
				pts = [][2]float64{{0, 0}, {b.W - d, 0}, {b.W, b.H / 2}, {b.W - d, b.H}, {0, b.H}}
			} else {
				pts = [][2]float64{{0, 0}, {b.W - d, 0}, {b.W, b.H / 2}, {b.W - d, b.H}, {0, b.H}, {d, b.H / 2}}
			}
			if n.Number != "" || n.Sub != "" {
				x := b.X + 12
				if !n.First {
					x = b.X + d + 6
				}
				tx = Rect{x, b.Y, b.W - d - (x - b.X) - 6, b.H}
			} else {
				if !n.First {
					tx.X += d
					tx.W -= 2*d - 6
				} else {
					tx.W -= d - 6
				}
			}
		} else if n.Dir == "" || n.Dir == "right" {
			sh, hl := b.H*2/3, b.H/2
			t := (b.H - sh) / 2
			pts = [][2]float64{{0, t}, {b.W - hl, t}, {b.W - hl, 0}, {b.W, b.H / 2}, {b.W - hl, b.H}, {b.W - hl, t + sh}, {0, t + sh}}
			tx = Rect{b.X, b.Y + t, b.W - hl, sh}
		} else if n.Dir == "down" {
			sw, hl := b.W*2/3, b.W/2
			l := (b.W - sw) / 2
			pts = [][2]float64{{l, 0}, {l + sw, 0}, {l + sw, b.H - hl}, {b.W, b.H - hl}, {b.W / 2, b.H}, {0, b.H - hl}, {l, b.H - hl}}
			tx = Rect{b.X + l, b.Y, sw, b.H - hl}
		} else {
			return nil, true, fmt.Errorf("scene.unknown_arrow_direction")
		}
		if err == nil {
			err = r.diagramPolygon(p, id+".surface", b, surface, "", pts)
		}
		if isExpandedLibrary(r.source.Revision) && n.Type == "chevron" && n.Number != "" && n.Sub == "" && b.H >= 54 && b.W <= 126 {
			// Compact numbered options have enough height for two rows.
			// Sharing their width horizontally splits single-word headings.
			nst, e := r.sceneStyle("number")
			if e != nil {
				return nil, true, e
			}
			st, e := r.sceneStyle(token)
			if e != nil {
				return nil, true, e
			}
			st.Weight = 600
			nl, e := r.typeEngine.Measure(n.Number, nst, tx.W)
			if e != nil {
				return nil, true, e
			}
			l, e := r.typeEngine.Measure(n.Text, st, tx.W)
			if e != nil {
				return nil, true, e
			}
			nh := math.Max(nl.AllocationHeight, nl.OccupiedTop+nl.EstimatedOccupiedHeight)
			th := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
			if len(nl.Lines) != 1 || len(l.Lines) != 1 || nh+th > tx.H+.02 {
				return nil, true, fmt.Errorf("scene.chevron_compact_stack_overflow: %s", id)
			}
			top := tx.Y + (tx.H-nh-th)/2
			text("number", n.Number, "number", Rect{tx.X, top, tx.W, nh}, surface, "emphasis", "center", 0, false)
			text("text", n.Text, token, Rect{tx.X, top + nh, tx.W, th}, surface, "display", "center", 600, false)
			p.Warnings = append(p.Warnings, "wmds.v4.compact-chevron-number-above-heading")
			break
		}
		if n.Number != "" || n.Sub != "" {
			x := tx.X
			if n.Number != "" {
				st, _ := r.sceneStyle("number")
				l, e := r.typeEngine.Measure(n.Number, st, tx.W)
				if e != nil {
					return nil, true, e
				}
				nw := l.Lines[0].Advance + .01
				text("number", n.Number, "number", Rect{x, tx.Y, nw, tx.H}, surface, "emphasis", "left", 0, true)
				tx.X += nw + 9
				tx.W -= nw + 9
			}
			st, _ := r.sceneStyle(token)
			st.Weight = 600
			l, e := r.typeEngine.Measure(n.Text, st, tx.W)
			if e != nil {
				return nil, true, e
			}
			need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
			if n.Sub != "" {
				ss, _ := r.sceneStyle("small")
				sl, e := r.typeEngine.Measure(n.Sub, ss, tx.W)
				if e != nil {
					return nil, true, e
				}
				sh := math.Max(sl.AllocationHeight, sl.OccupiedTop+sl.EstimatedOccupiedHeight)
				if need+sh > tx.H+.02 {
					return nil, true, fmt.Errorf("scene.chevron_stack_overflow: %s", id)
				}
				top := tx.Y + (tx.H-need-sh)/2
				text("text", n.Text, token, Rect{tx.X, top, tx.W, need}, surface, "display", "left", 600, false)
				text("sub", n.Sub, "small", Rect{tx.X, top + need, tx.W, sh}, surface, "secondary", "left", 0, false)
			} else {
				text("text", n.Text, token, tx, surface, "display", "left", 600, true)
			}
		} else {
			text("text", n.Text, token, Rect{tx.X + 12, tx.Y, tx.W - 24, tx.H}, surface, "display", "center", 0, true)
		}
	case "connector":
		err = r.planDiagramConnector(p, id, n, ctx)
	case "cylinder":
		fill, e := r.sceneColor(surface, "bg")
		if e != nil {
			return nil, true, e
		}
		line, e := r.sceneColor("light", "strong")
		if e != nil {
			return nil, true, e
		}
		err = r.diagramShape(p, id+".surface", b, pptx.ShapeTypeCan, fill, line, .75, "solid", nil)
		ry := math.Min(b.H*.14, 12)
		box := Rect{b.X, b.Y + 2*ry, b.W, b.H - 3*ry}
		st, _ := r.sceneStyle("small")
		st.Weight = 600
		l, e := r.typeEngine.Measure(n.Text, st, box.W)
		if e != nil {
			return nil, true, e
		}
		need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
		if n.Sub != "" {
			ss, _ := r.sceneStyle("label")
			sl, e := r.typeEngine.Measure(n.Sub, ss, box.W)
			if e != nil {
				return nil, true, e
			}
			sh := math.Max(sl.AllocationHeight, sl.OccupiedTop+sl.EstimatedOccupiedHeight)
			if need+sh > box.H+.02 {
				return nil, true, fmt.Errorf("scene.cylinder_stack_overflow: %s", id)
			}
			top := box.Y + (box.H-need-sh)/2
			text("text", n.Text, "small", Rect{box.X, top, box.W, need}, surface, "display", "center", 600, false)
			text("sub", n.Sub, "label", Rect{box.X, top + need, box.W, sh}, surface, "secondary", "center", 0, false)
		} else {
			text("text", n.Text, "small", box, surface, "display", "center", 600, true)
		}
	case "layerrow":
		if n.Surface == "" {
			surface = "subtle"
		}
		rect("surface", b, surface)
		label := n.Label
		if n.N != "" {
			label = n.N + "  " + label
		}
		text("label", label, "label", Rect{b.X + 12, b.Y, 120, b.H}, surface, "display", "left", 600, true)
		if n.Highlight == "" {
			text("text", n.Text, "small", Rect{b.X + 144, b.Y, b.W - 156, b.H}, surface, "display", "left", 0, true)
		} else {
			st, _ := r.sceneStyle("small")
			st.Weight = 600
			a, _, e := r.sequenceNeed(n.Text, st, b.W-172)
			if e != nil {
				return nil, true, e
			}
			st.Weight = 400
			h, _, e := r.sequenceNeed(n.Highlight, st, b.W-172)
			if e != nil {
				return nil, true, e
			}
			total := a + h + 8
			if total > b.H+.02 {
				return nil, true, fmt.Errorf("scene.layerrow_highlight_overflow")
			}
			bb := Rect{b.X + 144, b.Y + (b.H-total)/2, b.W - 156, total}
			col, e := r.sceneColor("callout", "bg")
			if e != nil {
				return nil, true, e
			}
			if err == nil {
				err = r.diagramShape(p, id+".highlight-surface", bb, pptx.ShapeTypeRect, "FFFFFF", col, 1, "solid", nil)
			}
			if err == nil {
				_, err = r.sequenceStack(p, id+".highlight", []string{n.Text, n.Highlight}, []string{"small", "small"}, []int{600, 0}, bb.X+8, bb.Y+4, bb.W-16, 0, total-8, "light", []string{"display", "secondary"})
			}
		}
	case "node":
		rect("surface", b, surface)
		size, gap := 24.0, 12.0
		tbW := b.W - 60
		align := "left"
		if n.Layout == "top" {
			size, gap, tbW, align = 30, 9, b.W-24, "center"
		} else if n.Layout != "" {
			return nil, true, fmt.Errorf("scene.unknown_node_layout")
		}
		titleToken := n.Style
		if titleToken == "" {
			titleToken = "body"
		}
		titleStyle, e := r.sceneStyle(titleToken)
		if e != nil {
			return nil, true, e
		}
		titleStyle.Weight = 600
		subStyle, e := r.sceneStyle("small")
		if e != nil {
			return nil, true, e
		}
		th, _, e := r.sequenceNeed(n.Text, titleStyle, tbW)
		if e != nil {
			return nil, true, e
		}
		sh, _, e := r.sequenceNeed(n.Sub, subStyle, tbW)
		if e != nil {
			return nil, true, e
		}
		stackH := th + sh
		ib := Rect{b.X + 12, b.Y + (b.H-size)/2, size, size}
		tb := Rect{b.X + 48, b.Y + (b.H-stackH)/2, tbW, stackH}
		if n.Layout == "top" {
			total := size
			if stackH > 0 {
				total += gap + stackH
			}
			if total > b.H+.02 {
				return nil, true, fmt.Errorf("scene.node_stack_overflow: %s", id)
			}
			ib = Rect{b.X + (b.W-size)/2, b.Y + (b.H-total)/2, size, size}
			tb = Rect{b.X + 12, ib.Y + size + gap, tbW, stackH}
		} else if stackH > b.H+.02 {
			return nil, true, fmt.Errorf("scene.node_stack_overflow: %s", id)
		}
		if err == nil {
			err = r.sceneIcon(p, id+".icon", n.Icon, ib, surface, "display")
		}
		if err == nil {
			err = r.sceneText(p, id+".text", n.Text, titleStyle, Rect{tb.X, tb.Y, tb.W, th}, surface, "display", align)
		}
		if err == nil {
			err = r.sceneText(p, id+".sub", n.Sub, subStyle, Rect{tb.X, tb.Y + th, tb.W, sh}, surface, "secondary", align)
		}
	case "matrix":
		err = r.planDiagramMatrix(p, id, n, ctx)
	case "beforeafter":
		err = r.planDiagramBeforeAfter(p, id, n, ctx)
	}
	if err != nil {
		return nil, true, err
	}
	return diagramFinish(p, "scene."+n.Type), true, nil
}

func (r *renderer) planDiagramConnector(p *scenePlan, id string, n diagramSpec, ctx SceneContext) error {
	if len(n.Points) < 2 {
		return fmt.Errorf("scene.connector_requires_points")
	}
	pts := append([][2]float64(nil), n.Points...)
	if n.Elbow != "" {
		if len(pts) != 2 || n.Elbow != "h" && n.Elbow != "v" {
			return fmt.Errorf("scene.invalid_connector_elbow")
		}
		a, b := pts[0], pts[1]
		if n.Elbow == "h" {
			m := (a[0] + b[0]) / 2
			pts = [][2]float64{a, {m, a[1]}, {m, b[1]}, b}
		} else {
			m := (a[1] + b[1]) / 2
			pts = [][2]float64{a, {a[0], m}, {b[0], m}, b}
		}
	}
	ink := n.Ink
	if ink == "" {
		ink = "strong"
	}
	color, e := r.sceneColor(ctx.Surface, ink)
	if e != nil {
		return e
	}
	dash := n.Style
	if dash == "" {
		if n.Dashed {
			dash = "dashed"
		} else {
			dash = "solid"
		}
	}
	weight := 1.0
	switch dash {
	case "solid":
	case "dashed":
		dash = "dash"
	case "dotted":
		dash = "sysDot"
		weight = 1.5
	default:
		return fmt.Errorf("scene.unknown_connector_style")
	}
	for i := 1; i < len(pts); i++ {
		if pts[i] == pts[i-1] {
			continue
		}
		if e := r.diagramLine(p, fmt.Sprintf("%s.segment-%d", id, i), pts[i-1], pts[i], color, weight, dash); e != nil {
			return e
		}
	}
	head := n.Head
	if head == "" || head == "arrow" {
		head = "end"
	}
	if head != "end" && head != "start" && head != "both" && head != "none" {
		return fmt.Errorf("scene.unknown_connector_head")
	}
	arrow := func(a, b [2]float64, prefix string) error {
		ang := math.Atan2(b[1]-a[1], b[0]-a[0])
		for i, o := range []float64{math.Pi - .5, math.Pi + .5} {
			z := [2]float64{b[0] + 7*math.Cos(ang+o), b[1] + 7*math.Sin(ang+o)}
			if e := r.diagramLine(p, fmt.Sprintf("%s.%s.%d", id, prefix, i), b, z, color, 1, "solid"); e != nil {
				return e
			}
		}
		return nil
	}
	if head == "end" || head == "both" {
		if e := arrow(pts[len(pts)-2], pts[len(pts)-1], "head-end"); e != nil {
			return e
		}
	}
	if head == "start" || head == "both" {
		if e := arrow(pts[1], pts[0], "head-start"); e != nil {
			return e
		}
	}
	if n.StartDot {
		if e := r.diagramShape(p, id+".start-dot", Rect{pts[0][0] - 3.5, pts[0][1] - 3.5, 7, 7}, pptx.ShapeTypeEllipse, color, "", 0, "", nil); e != nil {
			return e
		}
	}
	if n.Label != "" {
		best := 0.0
		mid := [2]float64{}
		for i := 1; i < len(pts); i++ {
			l := math.Hypot(pts[i][0]-pts[i-1][0], pts[i][1]-pts[i-1][1])
			if l > best {
				best = l
				mid = [2]float64{(pts[i][0] + pts[i-1][0]) / 2, (pts[i][1] + pts[i-1][1]) / 2}
			}
		}
		st, _ := r.sceneStyle("label")
		l, e := r.typeEngine.Measure(n.Label, st, 8191)
		if e != nil {
			return e
		}
		w := l.Lines[0].Advance + 6
		y := mid[1] - 6
		if n.LabelPos == "above" {
			y -= 9
		} else if n.LabelPos != "" {
			return fmt.Errorf("scene.unknown_connector_label_position")
		}
		b := Rect{mid[0] - w/2, y, w, 0}
		if n.LabelPos != "above" {
			if e := r.sceneRect(p, id+".label-patch", Rect{b.X - 3, b.Y - 3, b.W + 6, 18}, ctx.Surface); e != nil {
				return e
			}
		}
		if e := r.sceneText(p, id+".label", n.Label, st, b, ctx.Surface, "secondary", "center"); e != nil {
			return e
		}
	}
	p.Warnings = append(p.Warnings, "Native connector paths retain source coordinates; they do not claim attached endpoints or automatic obstacle routing after edits.")
	return nil
}

type diagramMatrixRow struct {
	Key     string            `json:"key"`
	N       string            `json:"n"`
	Label   string            `json:"label"`
	Surface string            `json:"surface"`
	Outline bool              `json:"outline"`
	Cells   []json.RawMessage `json:"cells"`
}

func (r *renderer) planDiagramMatrix(p *scenePlan, id string, n diagramSpec, ctx SceneContext) error {
	if len(n.Rows) == 0 {
		return fmt.Errorf("scene.matrix_empty")
	}
	left := n.Labels == "left"
	if n.Labels != "" && n.Labels != "above" && !left {
		return fmt.Errorf("scene.unknown_matrix_labels")
	}
	lw := 0.0
	if left {
		lw = n.LabelW
		if lw == 0 {
			lw = 162
		}
	}
	ch := n.CellH
	if ch == 0 {
		ch = 18
	}
	gap, rowGap := 3.0, 18.0
	if n.Gap != nil {
		gap = *n.Gap
	}
	if n.RowGap != nil {
		rowGap = *n.RowGap
	}
	if ch <= 0 || gap < 0 || rowGap < 0 || n.W <= lw {
		return fmt.Errorf("scene.invalid_matrix_geometry")
	}
	y := n.Y
	if len(n.Cols) > 0 {
		cw := (n.W - lw - gap*float64(len(n.Cols)-1)) / float64(len(n.Cols))
		colKeys := map[string]bool{}
		for i, s := range n.Cols {
			key, e := diagramKey(ctx, "cols", i, "", colKeys)
			if e != nil {
				return e
			}
			if e := r.diagramText(p, id+".cols."+key, s, "label", Rect{n.X + lw + float64(i)*(cw+gap), y, cw, 18}, ctx.Surface, "secondary", "left", 600, false); e != nil {
				return e
			}
		}
		y += 18
	}
	seen := map[string]bool{}
	outlined := 0
	for i, raw := range n.Rows {
		var row diagramMatrixRow
		if e := sceneDecode(raw, &row); e != nil {
			return e
		}
		key, e := diagramKey(ctx, "rows", i, row.Key, seen)
		if e != nil {
			return e
		}
		prefix := id + ".rows." + key
		if len(row.Cells) == 0 {
			return fmt.Errorf("scene.matrix_empty_row")
		}
		top := y
		if !left {
			top += 18
		}
		lb := Rect{n.X, y, n.W, 18}
		if left {
			lb = Rect{n.X, top, lw - 12, ch}
		}
		middle := left
		if !left {
			lb.Y -= 2
		}
		if row.N != "" && !left {
			if e := r.diagramText(p, prefix+".number", row.N, "label", Rect{lb.X, lb.Y, 24, lb.H}, ctx.Surface, "primary", "left", 600, false); e != nil {
				return e
			}
			lb.X += 30
			lb.W -= 30
		}
		label := row.Label
		if left && row.N != "" {
			label = row.N + "  " + label
		}
		if e := r.diagramText(p, prefix+".label", label, "label", lb, ctx.Surface, "primary", "left", 600, middle); e != nil {
			return e
		}
		if !left {
			p.Warnings = append(p.Warnings, "Adapter resolution wmds.matrix-label-clearance.v1: place above-cell labels 2pt above their row origin with a separate 24pt number gutter and 6pt gap; retain cell and outline dimensions.")
		}
		cw := (n.W - lw - gap*float64(len(row.Cells)-1)) / float64(len(row.Cells))
		if cw <= 0 {
			return fmt.Errorf("scene.matrix_cell_width")
		}
		cellKeys := map[string]bool{}
		for j, rawcell := range row.Cells {
			var s string
			var cell struct {
				Key     string `json:"key"`
				Text    string `json:"text"`
				Surface string `json:"surface"`
			}
			if e := json.Unmarshal(rawcell, &s); e == nil {
				cell.Text = s
			} else if e := sceneDecode(rawcell, &cell); e != nil {
				return e
			}
			cc := ctx
			cc.Path = ctx.Path + fmt.Sprintf("/rows/%d", i)
			ck, e := diagramKey(cc, "cells", j, cell.Key, cellKeys)
			if e != nil {
				return e
			}
			surf := cell.Surface
			if surf == "" {
				surf = row.Surface
			}
			if surf == "" {
				surf = "subtle"
			}
			b := Rect{n.X + lw + float64(j)*(cw+gap), top, cw, ch}
			if e := r.sceneRect(p, prefix+".cells."+ck+".surface", b, surf); e != nil {
				return e
			}
			if cell.Text != "" {
				token := n.Style
				if token == "" {
					token = "small"
				}
				textBox := Rect{b.X + 4, b.Y, b.W - 8, b.H}
				if token == "small" && ch < 18 && rowGap >= 18-ch {
					st, e := r.sceneStyle(token)
					if e != nil {
						return e
					}
					need, _, e := r.sequenceNeed(cell.Text, st, textBox.W)
					if e != nil {
						return e
					}
					if need <= ch+rowGap+.02 && need > ch {
						textBox.Y -= (need - ch) / 2
						textBox.H = need
						p.Warnings = append(p.Warnings, "Adapter resolution wmds.compact-matrix-line.v1: retain cell surfaces; center measured text in the existing cell-plus-gap pitch.")
					}
				}
				if e := r.diagramText(p, prefix+".cells."+ck+".text", cell.Text, token, textBox, surf, "display", "center", 0, true); e != nil {
					return e
				}
			}
		}
		if row.Outline {
			outlined++
			if outlined > 1 {
				return fmt.Errorf("scene.matrix_multiple_outlines")
			}
			col, e := r.sceneColor("light", "callout")
			if e != nil {
				return e
			}
			if e := r.diagramShape(p, prefix+".outline", Rect{n.X + lw - 4.5, top - 4.5, n.W - lw + 9, ch + 9}, pptx.ShapeTypeRect, "", col, 1.5, "solid", nil); e != nil {
				return e
			}
		}
		y = top + ch + rowGap
	}
	return nil
}
func (r *renderer) planDiagramBeforeAfter(p *scenePlan, id string, n diagramSpec, ctx SceneContext) error {
	if len(n.Rows) == 0 {
		return fmt.Errorf("scene.beforeafter_empty")
	}
	cw := (n.W - 126) / 2
	if cw <= 24 {
		return fmt.Errorf("scene.beforeafter_width")
	}
	if e := r.diagramText(p, id+".left-heading", n.Left, "label", Rect{n.X, n.Y, cw, 18}, ctx.Surface, "secondary", "left", 600, false); e != nil {
		return e
	}
	if e := r.diagramText(p, id+".right-heading", n.Right, "label", Rect{n.X + cw + 126, n.Y, cw, 18}, ctx.Surface, "emphasis", "left", 600, false); e != nil {
		return e
	}
	seen := map[string]bool{}
	y := n.Y + 21
	for i, raw := range n.Rows {
		var row []string
		if e := json.Unmarshal(raw, &row); e != nil || len(row) < 2 || len(row) > 3 {
			return fmt.Errorf("scene.beforeafter_row_shape")
		}
		key, e := diagramKey(ctx, "rows", i, "", seen)
		if e != nil {
			return e
		}
		pre := id + ".rows." + key
		left, right := Rect{n.X, y, cw, 54}, Rect{n.X + cw + 126, y, cw, 54}
		if e := r.sceneRect(p, pre+".left-surface", left, "subtle"); e != nil {
			return e
		}
		if e := r.sceneRect(p, pre+".right-surface", right, "inverse"); e != nil {
			return e
		}
		if e := r.diagramText(p, pre+".left", row[0], "small", Rect{left.X + 12, left.Y, left.W - 24, left.H}, "subtle", "secondary", "left", 0, true); e != nil {
			return e
		}
		rw := right.W - 24
		if len(row) == 3 && row[2] != "" {
			st, _ := r.sceneStyle("label")
			st.Weight = 600
			l, e := r.typeEngine.Measure(row[2], st, 8191)
			if e != nil {
				return e
			}
			if len(l.Lines) != 1 {
				return fmt.Errorf("scene.beforeafter_change_requires_single_line: %s", pre)
			}
			w := sequenceInlineWidth(l.Lines[0].Advance, st) + 12
			rw -= w + 12
			if rw <= 0 {
				return fmt.Errorf("scene.beforeafter_result_pair_overflow: %s", pre)
			}
			box := Rect{right.X + right.W - 12 - w, y + 15, w, 24}
			if e := r.sceneRect(p, pre+".change-surface", box, "callout"); e != nil {
				return e
			}
			if e := r.diagramText(p, pre+".change", row[2], "label", Rect{box.X + 6, box.Y, box.W - 12, box.H}, "callout", "primary", "center", 600, true); e != nil {
				return e
			}
			p.Warnings = append(p.Warnings, "Adapter resolution wmds.native-inline-chip-width.v1: add trailing native textbox space to the result label and preserve the complete result/tag pair within its source row.")
		}
		if e := r.diagramText(p, pre+".right", row[1], "small", Rect{right.X + 12, right.Y, rw, right.H}, "inverse", "display", "left", 600, true); e != nil {
			return e
		}
		arrow := Rect{n.X + cw + 18, y + 15, 90, 24}
		if n.Arrow != "" {
			// Preserve the supplied artwork's aspect ratio and center it in
			// the row gutter, with the same 18pt clearance on both sides.
			arrow.H = 0
			im, e := r.primitiveArtworkImage(pre+".arrow", n.Arrow, arrow, ctx.Surface, "strong")
			if e != nil {
				return e
			}
			centerY := pptx.Inches((y+27)/72 - im.H.Val/2)
			im.Y = &centerY
			p.Items = append(p.Items, sceneItem{Image: im})
		} else {
			if e := r.diagramPolygon(p, pre+".arrow", arrow, "inverse", "", [][2]float64{{0, 6}, {78, 6}, {78, 0}, {90, 12}, {78, 24}, {78, 18}, {0, 18}}); e != nil {
				return e
			}
		}
		y += 63
	}
	return nil
}
