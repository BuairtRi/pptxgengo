package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

// SceneDataContract is the source-context adapter. Its local bounds do not
// broaden the independent foundation/CardSpec minimums or grid contract.
const SceneDataContract = "pptxgengo.wmds-scene-data.v1"

type sceneMetricValue struct {
	Value  string            `json:"value"`
	Label  string            `json:"label"`
	Format *NumberFormatSpec `json:"format,omitempty"`
}
type sceneMetricSource struct {
	Type      string            `json:"type"`
	X         float64           `json:"x"`
	Y         float64           `json:"y"`
	W         float64           `json:"w"`
	H         float64           `json:"h"`
	Value     string            `json:"value,omitempty"`
	Label     string            `json:"label"`
	Format    *NumberFormatSpec `json:"format,omitempty"`
	Circle    bool              `json:"circle,omitempty"`
	Change    json.RawMessage   `json:"change,omitempty"`
	Secondary json.RawMessage   `json:"secondary,omitempty"`
	Target    string            `json:"target,omitempty"`
	Status    string            `json:"status,omitempty"`
	Source    string            `json:"source,omitempty"`
	On        string            `json:"on,omitempty"`
}
type sceneCardSource struct {
	SourceID      string            `json:"id,omitempty"`
	Type          string            `json:"type"`
	X             float64           `json:"x"`
	Y             float64           `json:"y"`
	W             float64           `json:"w"`
	H             float64           `json:"h"`
	Surface       string            `json:"surface,omitempty"`
	Pad           float64           `json:"pad"`
	PadTop        float64           `json:"padTop"`
	PadRight      float64           `json:"padRight"`
	ContentBottom float64           `json:"contentBottom,omitempty"`
	Gap           float64           `json:"gap"`
	Label         string            `json:"label,omitempty"`
	Title         string            `json:"title,omitempty"`
	TitleStyle    string            `json:"titleStyle,omitempty"`
	TitleInk      string            `json:"titleInk,omitempty"`
	Number        string            `json:"number,omitempty"`
	InlineNumber  string            `json:"inlineNumber,omitempty"`
	BandNumber    string            `json:"bandNumber,omitempty"`
	CornerNumber  string            `json:"cornerNumber,omitempty"`
	NumInk        string            `json:"numInk,omitempty"`
	Body          []json.RawMessage `json:"body,omitempty"`
	BodySize      string            `json:"bodySize,omitempty"`
	Band          *CardBand         `json:"band,omitempty"`
	Edge          *struct {
		Side   string  `json:"side"`
		Weight float64 `json:"weight"`
		Ink    string  `json:"ink"`
	} `json:"edge,omitempty"`
	Icon *struct {
		Name   string  `json:"name"`
		Layout string  `json:"layout"`
		Size   float64 `json:"size"`
	} `json:"icon,omitempty"`
	Media *struct {
		Src       string `json:"src"`
		Ratio     string `json:"ratio"`
		Focus     string `json:"focus,omitempty"`
		Grayscale bool   `json:"grayscale,omitempty"`
	} `json:"media,omitempty"`
	Metric      *sceneMetricSource `json:"metric,omitempty"`
	MetricGroup *struct {
		Primary   sceneMetricValue   `json:"primary"`
		Secondary []sceneMetricValue `json:"secondary"`
	} `json:"metricGroup,omitempty"`
	Quote *struct {
		Text    string `json:"text"`
		By      string `json:"by"`
		MarkInk string `json:"markInk,omitempty"`
	} `json:"quote,omitempty"`
	Person *struct {
		Initials string `json:"initials"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		Bio      string `json:"bio"`
	} `json:"person,omitempty"`
	Bio *struct {
		Initials  string            `json:"initials"`
		Name      string            `json:"name"`
		Title     string            `json:"title"`
		Role      string            `json:"role"`
		Photo     string            `json:"photo"`
		Focus     string            `json:"focus"`
		Grayscale bool              `json:"grayscale"`
		Tags      []string          `json:"tags"`
		Points    []json.RawMessage `json:"points"`
	} `json:"bio,omitempty"`
	Case *struct {
		Client    string             `json:"client"`
		Challenge string             `json:"challenge"`
		Did       []json.RawMessage  `json:"did"`
		Results   []sceneMetricValue `json:"results"`
	} `json:"case,omitempty"`
	Badge *struct {
		Kind    string  `json:"kind"`
		Corner  string  `json:"corner"`
		Surface string  `json:"surface"`
		Value   string  `json:"value"`
		Label   string  `json:"label"`
		Text    string  `json:"text"`
		Name    string  `json:"name"`
		Src     string  `json:"src"`
		Focus   string  `json:"focus"`
		Size    float64 `json:"size"`
	} `json:"badge,omitempty"`
	Fee         *sceneMetricValue `json:"fee,omitempty"`
	State       string            `json:"state,omitempty"`
	Tag         string            `json:"tag,omitempty"`
	Placeholder string            `json:"placeholder,omitempty"`
	Grayscale   bool              `json:"grayscale,omitempty"`
}
type sceneBodyBlock struct {
	P         string            `json:"p,omitempty"`
	Label     string            `json:"label,omitempty"`
	Bullets   []json.RawMessage `json:"bullets,omitempty"`
	Checklist []struct {
		Text string `json:"text"`
		On   bool   `json:"on"`
	} `json:"checklist,omitempty"`
	Columns [][]json.RawMessage `json:"columns,omitempty"`
}

func sceneDataKey(ctx SceneContext, path string, i int) (string, error) {
	keys, provided := ctx.Keys[ctx.Path+"/"+path]
	if provided {
		if i >= len(keys) || !validPartKey(keys[i]) {
			return "", fmt.Errorf("scene.invalid_item_key: %s/%s/%d", ctx.Path, path, i)
		}
		return keys[i], nil
	}
	if ctx.Keys != nil {
		return "", fmt.Errorf("scene.required_item_keys_missing: %s/%s", ctx.Path, path)
	}
	return fmt.Sprintf("item-%03d", i+1), nil
}
func sceneDataValue(v sceneMetricValue) (string, error) {
	if v.Format != nil {
		if v.Value != "" {
			return "", fmt.Errorf("metric.exactly_one_value_or_format_required")
		}
		return FormatNumber(*v.Format)
	}
	if strings.TrimSpace(v.Value) == "" {
		return "", fmt.Errorf("metric.value_required")
	}
	return v.Value, nil
}
func sceneDataBounds(p *scenePlan, b Rect) {
	if p.Bounds.W == 0 {
		p.Bounds = b
		return
	}
	x, y := math.Min(p.Bounds.X, b.X), math.Min(p.Bounds.Y, b.Y)
	right, bottom := math.Max(p.Bounds.X+p.Bounds.W, b.X+b.W), math.Max(p.Bounds.Y+p.Bounds.H, b.Y+b.H)
	p.Bounds = Rect{x, y, right - x, bottom - y}
}
func (r *renderer) sceneDataText(p *scenePlan, id, text string, st Style, b Rect, surface, role, align string, contexts ...SceneContext) (TextRecord, error) {
	var tr TextRecord
	if b.W <= 0 || b.H < 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
		return tr, fmt.Errorf("scene.invalid_text_geometry: %s", id)
	}
	if align != "" && align != "left" && align != "right" && align != "center" {
		return tr, fmt.Errorf("scene.invalid_text_align: %s", id)
	}
	if strings.TrimSpace(text) == "" {
		return tr, fmt.Errorf("scene.empty_text: %s", id)
	}
	if strings.Contains(text, "[[") || strings.Contains(text, "]]") || strings.Contains(text, "[^") {
		ctx := SceneContext{}
		if len(contexts) > 0 {
			ctx = contexts[0]
		}
		start := len(p.Items)
		if e := r.primitiveRichText(p, id, text, st, b, surface, role, align, "underscore", "mark", ctx); e != nil {
			return tr, e
		}
		for _, it := range p.Items[start:] {
			if it.Text != nil && it.Text.ID == id {
				tr = *it.Text
				sceneDataBounds(p, tr.Rect)
			}
			if it.Image != nil && it.Image.X != nil && it.Image.Y != nil && it.Image.W != nil && it.Image.H != nil {
				sceneDataBounds(p, Rect{it.Image.X.Val * 72, it.Image.Y.Val * 72, it.Image.W.Val * 72, it.Image.H.Val * 72})
			}
		}
		if tr.ID == "" {
			return tr, fmt.Errorf("scene.rich_text_record_missing: %s", id)
		}
		return tr, nil
	}
	color, e := r.sceneColor(surface, role)
	if e != nil {
		return tr, e
	}
	bg, e := r.sceneColor(surface, "bg")
	if e != nil {
		return tr, e
	}
	minimum := 4.5
	if st.Size >= 18 || st.Size >= 14 && st.Weight >= 700 {
		minimum = 3
	}
	if contrast(color, bg) < minimum {
		return tr, fmt.Errorf("scene.text_contrast: %s", id)
	}
	layout, e := r.typeEngine.Measure(text, st, b.W)
	if e != nil {
		return tr, e
	}
	need := math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight)
	if b.H == 0 {
		b.H = need
	}
	if need > b.H+.02 {
		return tr, fmt.Errorf("scene.text_overflow: %s needs%.3fpt capacity%.3fpt", id, need, b.H)
	}
	if align == "" {
		align = "left"
	}
	tr = TextRecord{ID: id, Rect: b, Color: color, Align: align, Layout: layout}
	p.Items = append(p.Items, sceneItem{Text: &tr})
	sceneDataBounds(p, b)
	if strings.HasPrefix(layout.VerticalPolicy, "uncalibrated:") {
		p.Warnings = append(p.Warnings, "Unanchored source text variant: "+id+" ("+st.ID+").")
	}
	return tr, nil
}
func (r *renderer) sceneDataToken(name string, weight int) (Style, error) {
	st, e := r.sceneStyle(name)
	if e != nil {
		return st, e
	}
	if weight != 0 {
		st.Weight = weight
	}
	return st, nil
}
func (r *renderer) sceneDataShape(p *scenePlan, id string, b Rect, typ pptx.ShapeType, color string, line *pptx.ShapeLineProps) {
	if line == nil {
		line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none", Color: color}}
	}
	sh := sceneShape{Type: typ, Props: pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Color: color}, Line: line}, Record: ShapeRecord{ID: id, Rect: b, Color: color, Geometry: string(typ)}}
	p.Items = append(p.Items, sceneItem{Shape: &sh})
	sceneDataBounds(p, b)
}
func (r *renderer) sceneDataInkShape(p *scenePlan, id string, b Rect, surface, ink string) error {
	color, e := r.sceneColor(surface, ink)
	if e != nil {
		return e
	}
	r.sceneDataShape(p, id, b, pptx.ShapeTypeRect, color, nil)
	return nil
}
func sceneDataGroup(p *scenePlan, id, definition string, from int, b Rect) {
	rec := ComponentRecord{ID: id, Definition: definition, Contract: SceneDataContract, Rect: b, Density: "source_owned"}
	for _, item := range p.Items[from:] {
		if item.Text != nil {
			rec.Parts = append(rec.Parts, item.Text.ID)
		}
		if item.Shape != nil {
			rec.Parts = append(rec.Parts, item.Shape.Record.ID)
			rec.Shapes = append(rec.Shapes, item.Shape.Record)
		}
		if item.Image != nil {
			rec.Parts = append(rec.Parts, item.Image.ObjectName)
		}
		if item.Table != nil {
			rec.Parts = append(rec.Parts, item.Table.ID)
		}
		if item.Chart != nil {
			rec.Parts = append(rec.Parts, item.Chart.ID)
		}
	}
	parts := map[string]bool{}
	for _, part := range rec.Parts {
		parts[part] = true
	}
	groups := map[string]ComponentRecord{}
	references := map[string]bool{}
	for _, g := range p.Groups {
		groups[g.ID] = g
		for _, part := range g.Parts {
			references[part] = true
		}
	}
	owned := map[string]bool{}
	var leaves func(string, map[string]bool)
	leaves = func(gid string, out map[string]bool) {
		g, ok := groups[gid]
		if !ok {
			out[gid] = true
			return
		}
		for _, part := range g.Parts {
			leaves(part, out)
		}
	}
	var mark func(string)
	mark = func(gid string) {
		g, ok := groups[gid]
		if !ok {
			owned[gid] = true
			return
		}
		for _, part := range g.Parts {
			owned[part] = true
			mark(part)
		}
	}
	for _, g := range p.Groups {
		if references[g.ID] {
			continue
		}
		leaf := map[string]bool{}
		leaves(g.ID, leaf)
		contained := len(leaf) > 0
		for part := range leaf {
			if !parts[part] {
				contained = false
			}
		}
		if contained {
			rec.Parts = append(rec.Parts, g.ID)
			mark(g.ID)
		}
	}
	kept := rec.Parts[:0]
	for _, part := range rec.Parts {
		if !owned[part] {
			kept = append(kept, part)
		}
	}
	rec.Parts = kept

	p.Groups = append(p.Groups, rec)
}
func shiftSceneDataItems(items []sceneItem, dy float64) {
	for _, it := range items {
		if it.Text != nil {
			it.Text.Rect.Y += dy
		}
		if it.Shape != nil {
			it.Shape.Record.Rect.Y += dy
			it.Shape.Props.PositionProps = pos(it.Shape.Record.Rect)
		}
		if it.Image != nil && it.Image.Y != nil {
			v := pptx.Inches(it.Image.Y.Val + dy/72)
			it.Image.Y = &v
		}
	}
}

func (r *renderer) sceneBodyFlow(p *scenePlan, id string, blocks []json.RawMessage, ctx SceneContext, path string, b Rect, surface, size string, gap float64) (float64, error) {
	y := b.Y
	style := "body"
	if size == "small" {
		style = "small"
	}
	st, e := r.sceneDataToken(style, 0)
	if e != nil {
		return y, e
	}
	for i, raw := range blocks {
		var bl sceneBodyBlock
		if e := sceneDecode(raw, &bl); e != nil {
			return y, e
		}
		key, e := sceneDataKey(ctx, path, i)
		if e != nil {
			return y, e
		}
		part := id + "." + key
		kinds := 0
		if bl.P != "" {
			kinds++
		}
		if bl.Label != "" {
			kinds++
		}
		if len(bl.Bullets) > 0 {
			kinds++
		}
		if len(bl.Checklist) > 0 {
			kinds++
		}
		if len(bl.Columns) > 0 {
			kinds++
		}
		if kinds != 1 {
			return y, fmt.Errorf("scene.body_requires_one_kind: %s", part)
		}
		if bl.P != "" {
			tr, e := r.sceneDataText(p, part+".paragraph", bl.P, st, Rect{b.X, y, b.W, 0}, surface, "secondary", "left", ctx)
			if e != nil {
				return y, e
			}
			y = tr.Rect.Y + tr.Rect.H + gap
		}
		if bl.Label != "" {
			ls, _ := r.sceneDataToken("label", 0)
			tr, e := r.sceneDataText(p, part+".label", bl.Label, ls, Rect{b.X, y, b.W, 0}, surface, "emphasis", "left", ctx)
			if e != nil {
				return y, e
			}
			y = tr.Rect.Y + tr.Rect.H + gap
		}
		if len(bl.Bullets) > 0 {
			end, e := r.sceneDataBullets(p, part+".bullets", bl.Bullets, ctx, fmt.Sprintf("%s/%d/bullets", path, i), Rect{b.X, y, b.W, 0}, surface, st)
			if e != nil {
				return y, e
			}
			y = end + gap
		}
		if len(bl.Checklist) > 0 {
			y += 3
			for j, it := range bl.Checklist {
				kk, e := sceneDataKey(ctx, fmt.Sprintf("%s/%d/checklist", path, i), j)
				if e != nil {
					return y, e
				}
				iid := part + ".checklist." + kk
				ink := "secondary"
				if it.On {
					ink = "primary"
				}
				fg, e := r.sceneColor(surface, ink)
				if e != nil {
					return y, e
				}
				r.sceneDataShape(p, iid+".marker", Rect{b.X, y + 4, 12, 12}, pptx.ShapeTypeRect, fg, nil)
				if it.On {
					bg, _ := r.sceneColor(surface, "bg")
					r.sceneDataCheck(p, iid+".check", Rect{b.X, y + 4, 12, 12}, bg)
				} else {
					bg, _ := r.sceneColor(surface, "bg")
					r.sceneDataShape(p, iid+".empty", Rect{b.X + 1, y + 5, 10, 10}, pptx.ShapeTypeRect, bg, nil)
					r.sceneDataShape(p, iid+".dash", Rect{b.X + 3.5, y + 9.4, 5, 1.2}, pptx.ShapeTypeRect, fg, nil)
				}
				tr, e := r.sceneDataText(p, iid+".text", it.Text, st, Rect{b.X + 21, y, b.W - 21, 0}, surface, ink, "left", ctx)
				if e != nil {
					return y, e
				}
				y = math.Max(y+16, tr.Rect.Y+tr.Rect.H) + 6
			}
			y += gap - 6
		}
		if len(bl.Columns) > 0 {
			if len(bl.Columns) > 4 {
				return y, fmt.Errorf("scene.too_many_body_columns")
			}
			w := (b.W - 18*float64(len(bl.Columns)-1)) / float64(len(bl.Columns))
			bottom := y
			for j, col := range bl.Columns {
				end, e := r.sceneBodyFlow(p, part+fmt.Sprintf(".column-%d", j+1), col, ctx, fmt.Sprintf("%s/%d/columns/%d", path, i, j), Rect{b.X + float64(j)*(w+18), y, w, 0}, surface, size, gap)
				if e != nil {
					return y, e
				}
				bottom = math.Max(bottom, end)
			}
			y = bottom + gap
		}
	}
	return y - gap, nil
}
func ptrSceneBool(v bool) *bool { return &v }
func (r *renderer) sceneDataBullets(p *scenePlan, id string, items []json.RawMessage, ctx SceneContext, path string, b Rect, surface string, st Style) (float64, error) {
	y := b.Y
	mark, inset, gap, offset := 4.0, 15.0, 6.0, 8.5
	if st.Size <= 12 {
		mark, inset, gap, offset = 3, 12, 3, 7
	}
	for i, raw := range items {
		key, e := sceneDataKey(ctx, path, i)
		if e != nil {
			return y, e
		}
		iid := id + "." + key
		var text string
		var obj struct {
			Lead string            `json:"lead"`
			Text string            `json:"text"`
			Sub  []json.RawMessage `json:"sub"`
		}
		if e := json.Unmarshal(raw, &text); e != nil {
			if e = sceneDecode(raw, &obj); e != nil {
				return y, e
			}
			text = obj.Text
		}
		if e := r.sceneDataInkShape(p, iid+".marker", Rect{b.X, y + offset, mark, mark}, surface, "primary"); e != nil {
			return y, e
		}
		var tr TextRecord
		if obj.Lead != "" {
			rs := RichTextSpec{Paragraphs: []RichParagraphSpec{{Key: "text", Runs: []RichRunSpec{{Key: "lead", Text: obj.Lead, Weight: 600, Ink: "primary"}, {Key: "body", Text: " " + text, Ink: "secondary"}}}}}
			n := Node{ID: iid + ".text", Style: st.ID, Ink: "primary"}
			var e error
			tr, e = r.planRich(n, Rect{b.X + inset, y, b.W - inset, 0}, Rect{b.X + inset, y, b.W - inset, 10000}, surface, &rs)
			if e != nil {
				return y, e
			}
			p.Items = append(p.Items, sceneItem{Text: &tr})
			sceneDataBounds(p, tr.Rect)
		} else {
			var e error
			tr, e = r.sceneDataText(p, iid+".text", text, st, Rect{b.X + inset, y, b.W - inset, 0}, surface, "primary", "left", ctx)
			if e != nil {
				return y, e
			}
		}
		y = tr.Rect.Y + tr.Rect.H + gap
		if len(obj.Sub) > 0 {
			end, e := r.sceneDataBullets(p, iid+".sub", obj.Sub, ctx, fmt.Sprintf("%s/%d/sub", path, i), Rect{b.X + inset, y, b.W - inset, 0}, surface, st)
			if e != nil {
				return y, e
			}
			y = end + gap
		}
	}
	return y - gap, nil
}

func (r *renderer) planCardScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type == "metric" || tag.Type == "callout" {
		p, e := r.sceneDirectMetric(id, raw, ctx)
		return p, true, e
	}
	if tag.Type == "feesummary" {
		p, e := r.sceneFeeSummary(id, raw, ctx)
		return p, true, e
	}
	if tag.Type == "cardrow" {
		p, e := r.sceneSourceCardRow(id, raw, ctx)
		return p, true, e
	}
	if tag.Type != "card" {
		return nil, false, nil
	}
	var n sceneCardSource
	if e := sceneDecode(raw, &n); e != nil {
		return nil, true, e
	}
	p, e := r.sceneCard(id, n, ctx)
	return p, true, e
}
func (r *renderer) sceneSourceCardRow(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	var n struct {
		Type      string            `json:"type"`
		X         float64           `json:"x"`
		Y         float64           `json:"y"`
		W         float64           `json:"w"`
		H         float64           `json:"h"`
		Gap       float64           `json:"gap"`
		Numbering string            `json:"numbering"`
		Card      json.RawMessage   `json:"card"`
		Items     []json.RawMessage `json:"items"`
	}
	if e := sceneDecode(raw, &n); e != nil {
		return nil, e
	}
	if len(n.Items) == 0 || n.Gap < 0 {
		return nil, fmt.Errorf("scene.invalid_card_row")
	}
	p := &scenePlan{ID: id}
	for i, item := range n.Items {
		key, e := sceneDataKey(ctx, "items", i)
		if e != nil {
			return nil, e
		}
		var base, over map[string]json.RawMessage
		if e = json.Unmarshal(n.Card, &base); e != nil {
			return nil, e
		}
		if base == nil {
			base = map[string]json.RawMessage{}
		}
		if e = json.Unmarshal(item, &over); e != nil {
			return nil, e
		}
		for k, v := range over {
			base[k] = v
		}
		set := func(k string, v any) { base[k], _ = json.Marshal(v) }
		set("type", "card")
		set("x", n.X+float64(i)*(n.W+n.Gap))
		set("y", n.Y)
		set("w", n.W)
		set("h", n.H)
		num := fmt.Sprintf("%02d", i+1)
		switch n.Numbering {
		case "", "none":
		case "inline":
			set("inlineNumber", num)
		case "band":
			set("bandNumber", num)
			set("numInk", "emphasis")
		case "corner":
			set("cornerNumber", num)
			set("padRight", 54)
		default:
			return nil, fmt.Errorf("scene.invalid_numbering: %s", n.Numbering)
		}
		rr, _ := json.Marshal(base)
		cc := ctx
		cc.Path = ctx.Path + fmt.Sprintf("/items/%d", i)
		if ctx.Keys != nil {
			cc.Keys = map[string][]string{}
			for path, keys := range ctx.Keys {
				cc.Keys[path] = keys
				if strings.HasPrefix(path, ctx.Path+"/card/") {
					field := strings.Split(strings.TrimPrefix(path, ctx.Path+"/card/"), "/")[0]
					if _, changed := over[field]; !changed {
						cc.Keys[cc.Path+"/"+strings.TrimPrefix(path, ctx.Path+"/card/")] = keys
					}
				}
			}
		}
		child, _, e := r.planCardScene(id+"."+key, rr, cc)
		if e != nil {
			return nil, e
		}
		p.Items = append(p.Items, child.Items...)
		p.Groups = append(p.Groups, child.Groups...)
		p.Warnings = append(p.Warnings, child.Warnings...)
		sceneDataBounds(p, child.Bounds)
	}
	sceneDataGroup(p, id, "cardrow", 0, p.Bounds)
	return p, nil
}

func (r *renderer) sceneCard(id string, n sceneCardSource, ctx SceneContext) (*scenePlan, error) {
	p := &scenePlan{ID: id}
	b := Rect{n.X, n.Y, n.W, n.H}
	if b.W <= 0 || b.H <= 0 {
		return nil, fmt.Errorf("scene.card_requires_fixed_positive_bounds")
	}
	surface := n.Surface
	if surface == "" {
		surface = "light"
	}
	if n.State != "" && n.State != "featured" && n.State != "deemph" && n.State != "placeholder" {
		return nil, fmt.Errorf("scene.invalid_card_state")
	}
	if n.BodySize != "" && n.BodySize != "small" {
		return nil, fmt.Errorf("scene.unsupported_body_size")
	}
	if n.Icon != nil && (n.Icon.Size <= 0 || n.Icon.Layout != "inline" && n.Icon.Layout != "stack" && n.Icon.Layout != "side") {
		return nil, fmt.Errorf("scene.invalid_card_icon")
	}
	if n.Band != nil && n.Media != nil {
		return nil, fmt.Errorf("scene.unsupported_media_band_union")
	}
	if n.Metric != nil {
		m := n.Metric
		if m.Type != "" || m.X != 0 || m.Y != 0 || m.W != 0 || m.H != 0 || m.Circle || m.On != "" || m.Target != "" || m.Source != "" {
			return nil, fmt.Errorf("scene.unsupported_nested_metric_field")
		}
	}
	if n.Pad == 0 {
		n.Pad = 18
	}
	if n.Gap == 0 {
		n.Gap = 6
	}

	// Versioned, fixed source-context presets resolve authored dense-card
	// spacing conflicts. They do not change fonts, copy or outside bounds.
	checklistFee := false
	if n.Fee != nil {
		for _, block := range n.Body {
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(block, &fields)
			if _, ok := fields["checklist"]; ok {
				checklistFee = true
			}
		}
	}
	if n.Gap == 6 && (n.BodySize == "small" || checklistFee) {
		n.Gap = 3
		p.Warnings = append(p.Warnings, "source-layout.v1/dense-card-gap-3: "+id+" uses fixed3pt stack gap; source outer geometry, fonts and copy retained.")
	}
	if n.H == 90 && n.Pad == 18 && n.TitleStyle == "heading" && n.Icon != nil && n.Icon.Layout == "side" {
		n.Pad = 12
		p.Warnings = append(p.Warnings, "source-layout.v1/compact-side-heading-pad-12: "+id+" uses fixed12pt padding within90pt source card.")
	}
	if n.H == 162 && n.Pad == 18 && n.Quote != nil {
		n.Pad = 9
		p.Warnings = append(p.Warnings, "source-layout.v1/compact-quote-pad-9: "+id+" uses fixed9pt padding within162pt source card;48pt quote font and conservative glyph allocation retained.")
	}
	if n.W == 126 && n.H == 126 && n.Pad == 12 && n.Metric != nil {
		n.Pad = 9
		p.Warnings = append(p.Warnings, "source-layout.v1/compact-metric-pad-9: "+id+" uses fixed9pt padding in126pt square metric card; stat-sm and body labels retained.")
	}

	if n.Pad < 0 || n.Gap < 0 {
		return nil, fmt.Errorf("scene.invalid_card_spacing")
	}
	pt, pr, pb, pl := n.Pad, n.Pad, n.Pad, n.Pad
	if n.PadTop != 0 {
		pt = n.PadTop
	}
	if n.PadRight != 0 {
		pr = n.PadRight
	}
	if n.State == "featured" && n.Edge == nil {
		n.Edge = &struct {
			Side   string  `json:"side"`
			Weight float64 `json:"weight"`
			Ink    string  `json:"ink"`
		}{"top", 3, "emphasis"}
	}
	if e := r.sceneRect(p, id+".container", b, surface); e != nil {
		return nil, e
	}
	if surface == "outline" || n.State == "deemph" || n.State == "placeholder" {
		line := &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: "CED7E6"}, Width: 1}
		if n.State == "deemph" {
			line.Width = .75
			bg, e := r.sceneColor(ctx.Surface, "bg")
			if e != nil {
				return nil, e
			}
			p.Items[0].Shape.Props.Fill.Color = bg
			p.Items[0].Shape.Record.Color = bg
		}
		if n.State == "placeholder" {
			line.Color = "97A4BA"
			line.DashType = "dash"
			p.Items[0].Shape.Props.Fill.Color = "FFFFFF"
			p.Items[0].Shape.Record.Color = "FFFFFF"
		}
		p.Items[0].Shape.Props.Line = line
		p.Items[0].Shape.Props.PositionProps = pos(Rect{b.X + line.Width/2, b.Y + line.Width/2, b.W - line.Width, b.H - line.Width})
	}
	if n.Edge != nil {
		ed := n.Edge
		if ed.Weight <= 0 || ed.Weight > 6 {
			return nil, fmt.Errorf("scene.invalid_edge_weight")
		}
		eb := b
		switch ed.Side {
		case "left":
			eb.W = ed.Weight
			pl += ed.Weight
		case "right":
			eb.X += eb.W - ed.Weight
			eb.W = ed.Weight
			pr += ed.Weight
		case "top":
			eb.H = ed.Weight
			pt += ed.Weight
		case "bottom":
			eb.Y += eb.H - ed.Weight
			eb.H = ed.Weight
			pb += ed.Weight
		default:
			return nil, fmt.Errorf("scene.invalid_edge_side")
		}
		if e := r.sceneDataInkShape(p, id+".edge", eb, surface, ed.Ink); e != nil {
			return nil, e
		}
	}
	// A composed card may contain an independently positioned footer. Reserve
	// its space explicitly while preserving the outer card surface and bounds.
	if n.ContentBottom != 0 {
		if n.ContentBottom <= b.Y+pt || n.ContentBottom > b.Y+b.H-pb {
			return nil, fmt.Errorf("scene.invalid_card_content_bottom: %s", id)
		}
		pb = b.Y + b.H - n.ContentBottom
	}
	y := b.Y
	x, w := b.X+pl, b.W-pl-pr
	if n.Media != nil {
		hh := math.Round(b.W * 9 / 16)
		if n.Media.Ratio == "square" {
			hh = b.W
		} else if n.Media.Ratio != "" && n.Media.Ratio != "16:9" {
			return nil, fmt.Errorf("scene.unsupported_media_ratio")
		}
		im, e := r.primitiveMediaImage(id+".media", n.Media.Src, Rect{b.X, y, b.W, hh}, n.Media.Focus, n.Media.Grayscale || n.Grayscale)
		if e != nil {
			return nil, e
		}
		p.Items = append(p.Items, sceneItem{Image: im})
		y += hh
	}
	bodyTop := y + pt
	titleStyle := n.TitleStyle
	if titleStyle == "" {
		titleStyle = "subhead"
	}
	titleInk := n.TitleInk
	if titleInk == "" {
		titleInk = "display"
	}
	numberInk := n.NumInk
	if numberInk == "" {
		numberInk = "emphasis"
	}
	title := func(surface string, top float64, number string, ix, iw float64) (float64, error) {
		st, e := r.sceneDataToken(titleStyle, 0)
		if e != nil {
			return top, e
		}
		if number != "" {
			ns, _ := r.sceneDataToken("number", 0)
			nl, e := r.typeEngine.Measure(number, ns, iw)
			if e != nil {
				return top, e
			}
			if len(nl.Lines) != 1 {
				return top, fmt.Errorf("scene.number_wrap")
			}
			nw := nl.Lines[0].Advance + 1
			nr, e := r.sceneDataText(p, id+".number", number, ns, Rect{ix, top, nw, 0}, surface, numberInk, "left", ctx)
			if e != nil {
				return top, e
			}
			tl, e := r.typeEngine.Measure(n.Title, st, iw-nw-9)
			if e != nil {
				return top, e
			}
			shift := nr.Layout.Lines[0].Baseline - tl.Lines[0].Baseline
			tr, e := r.sceneDataText(p, id+".title", n.Title, st, Rect{ix + nw + 9, top + shift, iw - nw - 9, 0}, surface, titleInk, "left", ctx)
			if e != nil {
				return top, e
			}
			return math.Max(nr.Rect.Y+nr.Rect.H, tr.Rect.Y+tr.Rect.H), nil
		}
		tr, e := r.sceneDataText(p, id+".title", n.Title, st, Rect{ix, top, iw, 0}, surface, titleInk, "left", ctx)
		return tr.Rect.Y + tr.Rect.H, e
	}
	if n.Band != nil {
		hs := n.Band.Surface
		bi := len(p.Items)
		y += pt
		if n.Label != "" {
			st, _ := r.sceneDataToken("label", 0)
			tr, e := r.sceneDataText(p, id+".label", n.Label, st, Rect{x, y, w, 0}, hs, "emphasis", "left", ctx)
			if e != nil {
				return nil, e
			}
			y = tr.Rect.Y + tr.Rect.H + n.Gap
		}
		if n.Title != "" {
			end, e := title(hs, y, n.BandNumber, x, w)
			if e != nil {
				return nil, e
			}
			y = end
		}
		y += n.Pad
		band := &scenePlan{}
		if e := r.sceneRect(band, id+".band", Rect{b.X, b.Y, b.W, y - b.Y}, hs); e != nil {
			return nil, e
		}
		p.Items = append(p.Items[:bi], append(band.Items, p.Items[bi:]...)...)
		bodyTop = y + n.Pad
	} else {
		y = bodyTop
	}
	if n.Band == nil {
		if n.Icon != nil && n.Icon.Layout == "stack" {
			im, e := r.sceneDataIcon(id+".icon", Rect{x, y, n.Icon.Size, n.Icon.Size}, n.Icon.Name, r.ink(surface, "display"))
			if e != nil {
				return nil, e
			}
			p.Items = append(p.Items, im.Items...)
			p.Groups = append(p.Groups, im.Groups...)
			y += n.Icon.Size + 12
		}
		if n.Icon != nil && n.Icon.Layout == "side" {
			im, e := r.sceneDataIcon(id+".icon", Rect{x, y, n.Icon.Size, n.Icon.Size}, n.Icon.Name, r.ink(surface, "display"))
			if e != nil {
				return nil, e
			}
			p.Append(im)
			x += n.Icon.Size + 12
			w -= n.Icon.Size + 12
		}
		if n.Label != "" {
			st, _ := r.sceneDataToken("label", 0)
			tr, e := r.sceneDataText(p, id+".label", n.Label, st, Rect{x, y, w, 0}, surface, "emphasis", "left", ctx)
			if e != nil {
				return nil, e
			}
			y = tr.Rect.Y + tr.Rect.H + n.Gap
		}
		if n.Number != "" && n.InlineNumber == "" {
			st, _ := r.sceneDataToken("number", 0)
			tr, e := r.sceneDataText(p, id+".number", n.Number, st, Rect{x, y, w, 0}, surface, numberInk, "left", ctx)
			if e != nil {
				return nil, e
			}
			y = tr.Rect.Y + tr.Rect.H + n.Gap
		}
		if n.Title != "" {
			tx, tw := x, w
			if n.Icon != nil && n.Icon.Layout == "inline" {
				im, e := r.sceneDataIcon(id+".icon", Rect{x, y, n.Icon.Size, n.Icon.Size}, n.Icon.Name, r.ink(surface, "display"))
				if e != nil {
					return nil, e
				}
				p.Append(im)
				tx += n.Icon.Size + 12
				tw -= n.Icon.Size + 12
			}
			end, e := title(surface, y, n.InlineNumber, tx, tw)
			if e != nil {
				return nil, e
			}
			if n.Icon != nil && n.Icon.Layout == "inline" {
				end = math.Max(end, y+n.Icon.Size)
			}
			margin := 6.
			if n.BodySize == "small" && n.TitleStyle == "body" && n.H <= 108 && n.Pad <= 12 && len(n.Body) > 0 {
				margin = 3
				p.Warnings = append(p.Warnings, "source-layout.v1/compact-body-title-margin-3: "+id+" uses fixed3pt title margin in dense compact body-title card.")
			}
			y = end + n.Gap + margin
		}
	}
	if n.Title != "" && n.Band == nil && len(n.Body) == 0 && n.Metric == nil && n.MetricGroup == nil && n.Person == nil && n.Bio == nil && n.Quote == nil && n.Case == nil && n.Fee == nil && n.State != "placeholder" {
		y -= 6
		p.Warnings = append(p.Warnings, "source-layout.v1/title-only-visible-envelope: "+id+" excludes trailing invisible6pt title margin from fit; text coordinates retained.")
	}

	if n.CornerNumber != "" {
		st, _ := r.sceneDataToken("number", 0)
		_, e := r.sceneDataText(p, id+".corner-number", n.CornerNumber, st, Rect{b.X + b.W - pr - 36, bodyTop - 3, 36, 0}, surface, numberInk, "right", ctx)
		if e != nil {
			return nil, e
		}
	}
	if len(n.Body) > 0 {
		end, e := r.sceneBodyFlow(p, id+".body", n.Body, ctx, "body", Rect{x, y, w, 0}, surface, n.BodySize, n.Gap)
		if e != nil {
			return nil, e
		}
		y = end + n.Gap
	}
	if n.Metric != nil {
		start := len(p.Items)
		end, e := r.sceneMetricFlow(p, id+".metric", *n.Metric, ctx, Rect{x, y, w, 0}, surface, b.W < 270, true)
		if e != nil {
			return nil, e
		}
		y = end + n.Gap
		if len(n.Body) == 0 {
			footer := -1
			for i := start; i < len(p.Items); i++ {
				it := p.Items[i]
				if it.Text != nil && strings.Contains(it.Text.ID, ".footer.") || it.Shape != nil && strings.Contains(it.Shape.Record.ID, ".footer.") {
					footer = i
					break
				}
			}
			if footer >= 0 && end <= b.Y+b.H-pb {
				shiftSceneDataItems(p.Items[footer:], b.Y+b.H-pb-end)
				y = b.Y + b.H - pb + n.Gap
			}
		}
	}
	if n.MetricGroup != nil {
		mg := n.MetricGroup
		if len(mg.Secondary) < 1 || len(mg.Secondary) > 3 {
			return nil, fmt.Errorf("scene.metric_group_count")
		}
		unit := (w - 18*float64(len(mg.Secondary))) / (1.3 + float64(len(mg.Secondary)))
		pw := unit * 1.3
		y += 6
		end, e := r.sceneMetricPair(p, id+".metric.primary", mg.Primary, Rect{x, y, pw, 0}, surface, "stat", "body", 6, ctx)
		if e != nil {
			return nil, e
		}
		bottom := end
		sx := x + pw + 18
		for i, sm := range mg.Secondary {
			key, e := sceneDataKey(ctx, "metricGroup/secondary", i)
			if e != nil {
				return nil, e
			}
			iid := id + ".metric.secondary." + key
			end, e := r.sceneMetricPair(p, iid, sm, Rect{sx + 12.75, y + 6, unit - 12.75, 0}, surface, "number", "small", 6, ctx)
			if e != nil {
				return nil, e
			}
			if e = r.sceneDataInkShape(p, iid+".divider", Rect{sx, y, .75, end - y}, surface, "line"); e != nil {
				return nil, e
			}
			bottom = math.Max(bottom, end)
			sx += unit + 18
		}
		y = bottom + n.Gap
	}
	if n.Person != nil {
		ps := n.Person
		if e := r.sceneRect(p, id+".initials-bg", Rect{x, y, 54, 54}, "inverse"); e != nil {
			return nil, e
		}
		st, _ := r.sceneDataToken("number", 0)
		_, e := r.sceneDataText(p, id+".initials", ps.Initials, st, Rect{x + 3, y + 15, 48, 24}, "inverse", "display", "center", ctx)
		if e != nil {
			return nil, e
		}
		y += 66
		for _, part := range []struct{ name, text, style, ink string }{{"name", ps.Name, "subhead", "display"}, {"role", ps.Role, "small", "secondary"}, {"bio", ps.Bio, "body", "primary"}} {
			if part.text == "" {
				continue
			}
			st, _ := r.sceneDataToken(part.style, 0)
			tr, e := r.sceneDataText(p, id+".person."+part.name, part.text, st, Rect{x, y, w, 0}, surface, part.ink, "left", ctx)
			if e != nil {
				return nil, e
			}
			y = tr.Rect.Y + tr.Rect.H + n.Gap
		}
	}
	if n.Quote != nil {
		q := n.Quote
		st, _ := r.sceneDataToken("heading", 0)
		st.Size = 48
		if strings.HasSuffix(st.Tracking, "em") {
			v, _ := strconv.ParseFloat(strings.TrimSuffix(st.Tracking, "em"), 64)
			st.TrackingPt = math.Round(v*st.Size*100) / 100
		}
		st.Leading = 36
		st.ID = "source.quote-mark.48.36"
		ink := q.MarkInk
		if ink == "" {
			ink = "emphasis"
			if surface == "inverse" || surface == "deep" {
				ink = "display"
			}
		}
		tr, e := r.sceneDataText(p, id+".quote.mark", "“", st, Rect{x, y, w, 0}, surface, ink, "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + n.Gap
		st, _ = r.sceneDataToken("subhead", 0)
		tr, e = r.sceneDataText(p, id+".quote.text", q.Text, st, Rect{x, y, w, 0}, surface, "display", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + n.Gap
		st, _ = r.sceneDataToken("label", 0)
		by, e := r.typeEngine.Measure(q.By, st, w)
		if e != nil {
			return nil, e
		}
		hh := math.Max(by.AllocationHeight, by.OccupiedTop+by.EstimatedOccupiedHeight)
		y = math.Max(y, b.Y+b.H-pb-hh)
		tr, e = r.sceneDataText(p, id+".quote.by", q.By, st, Rect{x, y, w, 0}, surface, "secondary", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + n.Gap
	}
	if n.Bio != nil {
		bio := n.Bio
		if bio.Photo != "" {
			im, e := r.primitiveMediaImage(id+".bio.photo", bio.Photo, Rect{x, y, 54, 54}, bio.Focus, bio.Grayscale || n.Grayscale)
			if e != nil {
				return nil, e
			}
			p.Items = append(p.Items, sceneItem{Image: im})
		} else {
			fg, _ := r.sceneColor(surface, "strong")
			bg, _ := r.sceneColor(surface, "bg")
			r.sceneDataShape(p, id+".bio.initials-bg", Rect{x, y, 54, 54}, pptx.ShapeTypeRect, fg, nil)
			st, _ := r.sceneDataToken("number", 0)
			layout, e := r.typeEngine.Measure(bio.Initials, st, 48)
			if e != nil {
				return nil, e
			}
			tr := TextRecord{ID: id + ".bio.initials", Rect: Rect{x + 3, y + 15, 48, 24}, Color: bg, Align: "center", Layout: layout}
			p.Items = append(p.Items, sceneItem{Text: &tr})
		}
		top := y
		st, _ := r.sceneDataToken("subhead", 0)
		tr, e := r.sceneDataText(p, id+".bio.name", bio.Name, st, Rect{x + 66, y, w - 66, 0}, surface, "display", "left", ctx)
		if e != nil {
			return nil, e
		}
		st, _ = r.sceneDataToken("small", 0)
		tt, e := r.sceneDataText(p, id+".bio.title", bio.Title, st, Rect{x + 66, tr.Rect.Y + tr.Rect.H + 3, w - 66, 0}, surface, "secondary", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = math.Max(top+54, tt.Rect.Y+tt.Rect.H) + n.Gap + 6
		st, _ = r.sceneDataToken("label", 0)
		tr, e = r.sceneDataText(p, id+".bio.role", bio.Role, st, Rect{x, y, w, 0}, surface, "emphasis", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + n.Gap
		tx := x
		tagY := y
		for i, tag := range bio.Tags {
			kk, e := sceneDataKey(ctx, "bio/tags", i)
			if e != nil {
				return nil, e
			}
			tl, e := r.typeEngine.Measure(tag, st, w-12)
			if e != nil {
				return nil, e
			}
			tw := tl.Lines[0].Advance + 14
			if tx+tw > x+w {
				tx = x
				tagY += st.Leading + 6
			}
			line := &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: r.ink(surface, "strong")}, Width: .75}
			r.sceneDataShape(p, id+".bio.tag."+kk+".outline", Rect{tx, tagY, tw, st.Leading + 2}, pptx.ShapeTypeRect, r.ink(surface, "bg"), line)
			_, e = r.sceneDataText(p, id+".bio.tag."+kk, tag, st, Rect{tx + 6, tagY + 1, tw - 12, 0}, surface, "primary", "left", ctx)
			if e != nil {
				return nil, e
			}
			tx += tw + 9
		}
		if len(bio.Tags) > 0 {
			p.Warnings = append(p.Warnings, "user-review.v1/bio-tag-width-reserve: "+id+" tags include fixed2pt native width allowance; authored font size retained.")
			y = tagY + st.Leading + 2 + n.Gap
		}
		y += 3
		st, _ = r.sceneDataToken("small", 0)
		end, e := r.sceneDataBullets(p, id+".bio.points", bio.Points, ctx, "bio/points", Rect{x, y, w, 0}, surface, st)
		if e != nil {
			return nil, e
		}
		y = end + n.Gap
	}
	if n.Case != nil {
		cs := n.Case
		st, _ := r.sceneDataToken("eyebrow", 0)
		tr, e := r.sceneDataText(p, id+".case.client", cs.Client, st, Rect{x, y, w, 0}, surface, "emphasis", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + n.Gap
		st, _ = r.sceneDataToken("body", 0)
		tr, e = r.sceneDataText(p, id+".case.challenge", cs.Challenge, st, Rect{x, y, w, 0}, surface, "display", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + n.Gap
		st, _ = r.sceneDataToken("small", 0)
		end, e := r.sceneDataBullets(p, id+".case.did", cs.Did, ctx, "case/did", Rect{x, y, w, 0}, surface, st)
		if e != nil {
			return nil, e
		}
		y = end + n.Gap
		if len(cs.Results) == 0 {
			return nil, fmt.Errorf("scene.case_results_required")
		}
		start := len(p.Items)
		end, e = r.sceneStackedResults(p, id+".case.results", cs.Results, ctx, "case/results", Rect{x, y, w, 0}, surface)
		if e != nil {
			return nil, e
		}
		if end < b.Y+b.H-pb {
			shiftSceneDataItems(p.Items[start:], b.Y+b.H-pb-end)
			end = b.Y + b.H - pb
		}
		y = end + n.Gap
	}
	if n.Fee != nil {
		value, e := sceneDataValue(*n.Fee)
		if e != nil {
			return nil, e
		}
		start := len(p.Items)
		if e = r.sceneDataInkShape(p, id+".fee.divider", Rect{x, y, w, .75}, surface, "line"); e != nil {
			return nil, e
		}
		y += 9.75
		st, _ := r.sceneDataToken("number", 0)
		tr, e := r.sceneDataText(p, id+".fee.value", value, st, Rect{x, y, w, 0}, surface, "display", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + 2
		st, _ = r.sceneDataToken("small", 0)
		tr, e = r.sceneDataText(p, id+".fee.label", n.Fee.Label, st, Rect{x, y, w, 0}, surface, "secondary", "left", ctx)
		if e != nil {
			return nil, e
		}
		end := tr.Rect.Y + tr.Rect.H
		if end < b.Y+b.H-pb {
			shiftSceneDataItems(p.Items[start:], b.Y+b.H-pb-end)
			end = b.Y + b.H - pb
		}
		y = end + n.Gap
	}
	if n.State == "placeholder" {
		text := n.Placeholder
		if text == "" {
			text = "To be confirmed"
		}
		st, _ := r.sceneDataToken("label", 0)
		tr, e := r.sceneDataText(p, id+".placeholder", text, st, Rect{x, y, w, 0}, surface, "secondary", "left", ctx)
		if e != nil {
			return nil, e
		}
		y = tr.Rect.Y + tr.Rect.H + n.Gap
	}
	if y-n.Gap > b.Y+b.H-pb+.02 {
		return nil, fmt.Errorf("scene.card_vertical_overflow: %s content bottom%.3f capacity%.3f", id, y-n.Gap, b.Y+b.H-pb)
	}
	if n.State == "featured" && n.Tag != "" {
		st, _ := r.sceneDataToken("label", 600)
		tl, e := r.typeEngine.Measure(n.Tag, st, b.W-12)
		if e != nil {
			return nil, e
		}
		// Keep the six-point margins and reserve two native points beyond the
		// shaped advance so PowerPoint does not wrap the final glyph.
		tw := tl.Lines[0].Advance + 14
		fg, _ := r.sceneColor(surface, "emphasis")
		r.sceneDataShape(p, id+".tag-bg", Rect{b.X + b.W - tw, b.Y, tw, st.Leading + 6}, pptx.ShapeTypeRect, fg, nil)
		on := "light"
		if contrast("FFFFFF", fg) >= 4.5 {
			on = "inverse"
		}
		tr, e := r.sceneDataText(p, id+".tag", n.Tag, st, Rect{b.X + b.W - tw + 6, b.Y + 3, tw - 12, 0}, on, "primary", "left", ctx)
		if e != nil {
			return nil, e
		}
		_ = tr
	}
	if n.Badge != nil {
		bb := n.Badge
		size := bb.Size
		if size == 0 {
			size = 54
		}
		bx := b.X - size/2
		if bb.Corner == "top-right" {
			bx = b.X + b.W - size/2
		} else if bb.Corner != "" && bb.Corner != "top-left" {
			return nil, fmt.Errorf("scene.badge_corner")
		}
		br := Rect{bx, b.Y - size/2, size, size}
		surf := bb.Surface
		if surf == "" {
			surf = "inverse"
		}
		if bb.Kind == "image" {
			im, e := r.primitiveMediaImage(id+".badge.image", bb.Src, br, bb.Focus, n.Grayscale)
			if e != nil {
				return nil, e
			}
			p.Items = append(p.Items, sceneItem{Image: im})
		} else {
			if e := r.sceneRect(p, id+".badge", br, surf); e != nil {
				return nil, e
			}
			switch bb.Kind {
			case "icon":
				im, e := r.sceneDataIcon(id+".badge.icon", Rect{br.X + (size-30)/2, br.Y + (size-30)/2, 30, 30}, bb.Name, r.ink(surf, "display"))
				if e != nil {
					return nil, e
				}
				p.Append(im)
			case "number":
				st, _ := r.sceneDataToken("number", 0)
				_, e := r.sceneDataText(p, id+".badge.number", bb.Text, st, Rect{br.X + 3, br.Y + (size-st.Leading)/2, size - 6, 0}, surf, "display", "center", ctx)
				if e != nil {
					return nil, e
				}
			case "metric":
				st, _ := r.sceneDataToken("number", 0)
				if len([]rune(bb.Value)) > 5 {
					st.ID = "source.badge-number.13"
					st.Size = 13
					if strings.HasSuffix(st.Tracking, "em") {
						v, _ := strconv.ParseFloat(strings.TrimSuffix(st.Tracking, "em"), 64)
						st.TrackingPt = math.Round(v*st.Size*100) / 100
					}
				}
				ls, _ := r.sceneDataToken("label", 0)
				vl, e := r.typeEngine.Measure(bb.Value, st, size-6)
				if e != nil {
					return nil, e
				}
				ll, e := r.typeEngine.Measure(bb.Label, ls, size-6)
				if e != nil {
					return nil, e
				}
				hh := math.Max(vl.AllocationHeight, vl.EstimatedOccupiedHeight) + 2 + math.Max(ll.AllocationHeight, ll.EstimatedOccupiedHeight)
				ty := br.Y + (size-hh)/2
				tr, e := r.sceneDataText(p, id+".badge.value", bb.Value, st, Rect{br.X + 3, ty, size - 6, 0}, surf, "display", "center", ctx)
				if e != nil {
					return nil, e
				}
				_, e = r.sceneDataText(p, id+".badge.label", bb.Label, ls, Rect{br.X + 3, tr.Rect.Y + tr.Rect.H + 2, size - 6, 0}, surf, "secondary", "center", ctx)
				if e != nil {
					return nil, e
				}
			default:
				return nil, fmt.Errorf("scene.badge_kind")
			}
		}
		sceneDataBounds(p, br)
	}
	if n.State == "deemph" || n.State == "placeholder" {
		secondary, _ := r.sceneColor(surface, "secondary")
		primary, _ := r.sceneColor(surface, "primary")
		display, _ := r.sceneColor(surface, "display")
		emphasis, _ := r.sceneColor(surface, "emphasis")
		for _, it := range p.Items {
			if it.Text != nil && (it.Text.Color == primary || it.Text.Color == display || n.State == "deemph" && it.Text.Color == emphasis) {
				it.Text.Color = secondary
			}
		}
	}
	sceneDataBounds(p, b)
	sceneDataGroup(p, id, "card.source", 0, p.Bounds)
	return p, r.err
}

func (r *renderer) sceneMetricPair(p *scenePlan, id string, m sceneMetricValue, b Rect, surface, valueStyle, labelStyle string, gap float64, contexts ...SceneContext) (float64, error) {
	value, e := sceneDataValue(m)
	if e != nil {
		return b.Y, e
	}
	st, e := r.sceneDataToken(valueStyle, 0)
	if e != nil {
		return b.Y, e
	}
	tr, e := r.sceneDataText(p, id+".value", value, st, Rect{b.X, b.Y, b.W, 0}, surface, "display", "left", contexts...)
	if e != nil {
		return b.Y, e
	}
	if len(tr.Layout.Lines) != 1 {
		return b.Y, fmt.Errorf("scene.metric_value_wrap: %s", id)
	}
	ls, e := r.sceneDataToken(labelStyle, 0)
	if e != nil {
		return b.Y, e
	}
	lab, e := r.sceneDataText(p, id+".label", m.Label, ls, Rect{b.X, tr.Rect.Y + tr.Rect.H + gap, b.W, 0}, surface, "secondary", "left", contexts...)
	return lab.Rect.Y + lab.Rect.H, e
}
func (r *renderer) sceneStackedResults(p *scenePlan, id string, list []sceneMetricValue, ctx SceneContext, path string, b Rect, surface string) (float64, error) {
	if len(list) > 4 {
		return b.Y, fmt.Errorf("scene.result_count")
	}
	if e := r.sceneDataInkShape(p, id+".divider", Rect{b.X, b.Y, b.W, .75}, surface, "line"); e != nil {
		return b.Y, e
	}
	y := b.Y + 9.75
	w := (b.W - 18*float64(len(list)-1)) / float64(len(list))
	bottom := y
	for i, sm := range list {
		key, e := sceneDataKey(ctx, path, i)
		if e != nil {
			return y, e
		}
		end, e := r.sceneMetricPair(p, id+"."+key, sm, Rect{b.X + float64(i)*(w+18), y, w, 0}, surface, "number", "small", 0, ctx)
		if e != nil {
			return y, e
		}
		bottom = math.Max(bottom, end)
	}
	return bottom, nil
}
func (r *renderer) sceneMetricFlow(p *scenePlan, id string, m sceneMetricSource, ctx SceneContext, b Rect, surface string, small, card bool) (float64, error) {
	vs := "stat"
	if small {
		vs = "stat-sm"
	}
	end, e := r.sceneMetricPair(p, id, sceneMetricValue{m.Value, m.Label, m.Format}, b, surface, vs, "body", 6, ctx)
	if e != nil {
		return b.Y, e
	}
	y := end
	cardChange := json.RawMessage(nil)
	if card && len(m.Change) > 0 && string(m.Change) != "null" {
		cardChange = m.Change
		m.Change = nil
	}
	if len(m.Change) > 0 && string(m.Change) != "null" {
		var text string
		if e := json.Unmarshal(m.Change, &text); e != nil {
			var ch struct {
				Dir  string `json:"dir"`
				Text string `json:"text"`
			}
			if e = sceneDecode(m.Change, &ch); e != nil {
				return y, e
			}
			if ch.Dir != "up" && ch.Dir != "down" {
				return y, fmt.Errorf("scene.change_direction")
			}
			text = ch.Text
			col, _ := r.sceneColor(surface, "emphasis")
			r.sceneDataShape(p, id+".footer.change.mark", Rect{b.X, y + 7, 8, 6}, pptx.ShapeTypeTriangle, col, nil)
			if ch.Dir == "down" {
				p.Items[len(p.Items)-1].Shape.Props.Rotate = 180
				p.Items[len(p.Items)-1].Shape.Record.Rotation = 180
			}
			st, _ := r.sceneDataToken("label", 0)
			tr, e := r.sceneDataText(p, id+".footer.change", text, st, Rect{b.X + 14, y + 6, b.W - 14, 0}, surface, "emphasis", "left", ctx)
			if e != nil {
				return y, e
			}
			y = tr.Rect.Y + tr.Rect.H
			goto secondary
		}
		st, _ := r.sceneDataToken("label", 0)
		tr, e := r.sceneDataText(p, id+".footer.change", text, st, Rect{b.X, y + 6, b.W, 0}, surface, "emphasis", "left", ctx)
		if e != nil {
			return y, e
		}
		y = tr.Rect.Y + tr.Rect.H
	}
secondary:
	if len(m.Secondary) > 0 && string(m.Secondary) != "null" {
		var list []sceneMetricValue
		if e := sceneDecode(m.Secondary, &list); e != nil {
			var sm sceneMetricValue
			if e = sceneDecode(m.Secondary, &sm); e != nil {
				return y, e
			}
			list = []sceneMetricValue{sm}
		}
		if len(list) > 3 {
			return y, fmt.Errorf("scene.secondary_count")
		}
		if len(list) > 0 {
			if e = r.sceneDataInkShape(p, id+".footer.secondary.divider", Rect{b.X, y + 12, b.W, .75}, surface, "line"); e != nil {
				return y, e
			}
			y += 21.75
			track := (b.W - 18*float64(len(list)-1)) / float64(len(list))
			bottom := y
			for i, sm := range list {
				key, e := "secondary", error(nil)
				if len(m.Secondary) > 0 && m.Secondary[0] == '[' {
					key, e = sceneDataKey(ctx, "secondary", i)
				}
				if e != nil {
					return y, e
				}
				value, e := sceneDataValue(sm)
				if e != nil {
					return y, e
				}
				st, _ := r.sceneDataToken("number", 0)
				vl, e := r.typeEngine.Measure(value, st, track)
				if e != nil {
					return y, e
				}
				if len(vl.Lines) != 1 {
					return y, fmt.Errorf("scene.secondary_wrap")
				}
				nw := vl.Lines[0].Advance + 1
				xx := b.X + float64(i)*(track+18)
				tr, e := r.sceneDataText(p, id+".footer.secondary."+key+".value", value, st, Rect{xx, y, nw, 0}, surface, "display", "left", ctx)
				if e != nil {
					return y, e
				}
				ls, _ := r.sceneDataToken("small", 0)
				ll, e := r.typeEngine.Measure(sm.Label, ls, track-nw-9)
				if e != nil {
					return y, e
				}
				shift := tr.Layout.Lines[0].Baseline - ll.Lines[0].Baseline
				lr, e := r.sceneDataText(p, id+".footer.secondary."+key+".label", sm.Label, ls, Rect{xx + nw + 9, y + shift, track - nw - 9, 0}, surface, "secondary", "left", ctx)
				if e != nil {
					return y, e
				}
				bottom = math.Max(bottom, math.Max(tr.Rect.Y+tr.Rect.H, lr.Rect.Y+lr.Rect.H))
			}
			y = bottom
		}
	}
	if len(cardChange) > 0 {
		var text string
		if e := json.Unmarshal(cardChange, &text); e != nil {
			return y, fmt.Errorf("scene.card_change_requires_string")
		}
		st, _ := r.sceneDataToken("label", 0)
		tr, e := r.sceneDataText(p, id+".footer.change", text, st, Rect{b.X, y + 6, b.W, 0}, surface, "emphasis", "left", ctx)
		if e != nil {
			return y, e
		}
		y = tr.Rect.Y + tr.Rect.H
	}

	if m.Target != "" || m.Status != "" {
		y += 6
		xx := b.X
		if m.Status != "" {
			name, ok := map[string]string{"on": "On track", "risk": "At risk", "off": "Off track"}[m.Status]
			if !ok {
				return y, fmt.Errorf("scene.status_enum")
			}
			color := strings.TrimPrefix(r.source.Tokens.Colors.Dataviz.KPI[m.Status], "#")
			navy, _ := r.sceneColor("light", "display")
			r.sceneDataShape(p, id+".footer.status.outline", Rect{xx, y + 2, 8, 8}, pptx.ShapeTypeRect, navy, nil)
			r.sceneDataShape(p, id+".footer.status.mark", Rect{xx + 1, y + 3, 6, 6}, pptx.ShapeTypeRect, color, nil)
			st, _ := r.sceneDataToken("label", 0)
			ll, e := r.typeEngine.Measure(name, st, b.W-14)
			if e != nil {
				return y, e
			}
			ww := ll.Lines[0].Advance + 1
			tr, e := r.sceneDataText(p, id+".footer.status.text", name, st, Rect{xx + 14, y, ww, 0}, surface, "primary", "left", ctx)
			if e != nil {
				return y, e
			}
			xx += 14 + ww + 9
			y = tr.Rect.Y + tr.Rect.H
			if m.Target != "" {
				tr, e = r.sceneDataText(p, id+".footer.target", m.Target, st, Rect{xx, tr.Rect.Y, b.X + b.W - xx, 0}, surface, "secondary", "left", ctx)
				if e != nil {
					return y, e
				}
				y = math.Max(y, tr.Rect.Y+tr.Rect.H)
			}
		} else {
			st, _ := r.sceneDataToken("label", 0)
			tr, e := r.sceneDataText(p, id+".footer.target", m.Target, st, Rect{xx, y, b.W, 0}, surface, "secondary", "left", ctx)
			if e != nil {
				return y, e
			}
			y = tr.Rect.Y + tr.Rect.H
		}
	}
	if m.Source != "" {
		st, _ := r.sceneDataToken("source", 0)
		tr, e := r.sceneDataText(p, id+".footer.source", "Source: "+m.Source, st, Rect{b.X, y + 9, b.W, 0}, surface, "secondary", "left", ctx)
		if e != nil {
			return y, e
		}
		y = tr.Rect.Y + tr.Rect.H
	}
	return y, nil
}
func (r *renderer) sceneDirectMetric(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	var n sceneMetricSource
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, e
	}
	if tag.Type == "callout" {
		var e error
		var c struct {
			Type      string            `json:"type"`
			X         float64           `json:"x"`
			Y         float64           `json:"y"`
			W         float64           `json:"w"`
			H         float64           `json:"h"`
			Label     string            `json:"label"`
			Value     string            `json:"value"`
			Text      string            `json:"text"`
			Secondary *sceneMetricValue `json:"secondary,omitempty"`
		}
		if e = sceneDecode(raw, &c); e != nil {
			return nil, e
		}
		n = sceneMetricSource{Type: c.Type, X: c.X, Y: c.Y, W: c.W, H: c.H, Label: c.Text, Value: c.Value}
		p := &scenePlan{ID: id}
		b := Rect{c.X, c.Y, c.W, c.H}
		if e = r.sceneRect(p, id+".container", Rect{c.X, c.Y, c.W, 1}, "callout"); e != nil {
			return nil, e
		}
		st, _ := r.sceneDataToken("eyebrow", 0)
		tr, e := r.sceneDataText(p, id+".eyebrow", c.Label, st, Rect{c.X + 18, c.Y + 18, c.W - 36, 0}, "callout", "primary", "left", ctx)
		if e != nil {
			return nil, e
		}
		n.Secondary, _ = json.Marshal(c.Secondary)
		end, e := r.sceneMetricFlow(p, id+".metric", n, ctx, Rect{c.X + 18, tr.Rect.Y + tr.Rect.H + 6, c.W - 36, 0}, "callout", false, false)
		if e != nil {
			return nil, e
		}
		b.H = end + 18 - c.Y
		p.Items[0].Shape.Record.Rect = b
		p.Items[0].Shape.Props.PositionProps = pos(b)
		p.Bounds = b
		sceneDataGroup(p, id, "data.callout", 0, b)
		return p, nil
	}
	if e := sceneDecode(raw, &n); e != nil {
		return nil, e
	}

	surface := ctx.Surface
	if n.On != "" {
		surface = n.On
	}
	p := &scenePlan{ID: id}
	b := Rect{n.X, n.Y, n.W, n.H}
	end, e := r.sceneMetricFlow(p, id, n, ctx, b, surface, false, false)
	if e != nil {
		return nil, e
	}
	if b.H == 0 {
		b.H = end - b.Y
	} else if end > b.Y+b.H+.02 {
		return nil, fmt.Errorf("scene.metric_vertical_overflow: %s", id)
	}
	if n.Circle {
		st, _ := r.sceneDataToken("stat", 0)
		val, e := sceneDataValue(sceneMetricValue{n.Value, n.Label, n.Format})
		if e != nil {
			return nil, e
		}
		l, e := r.typeEngine.Measure(val, st, b.W)
		if e != nil {
			return nil, e
		}
		if len(l.Lines) != 1 {
			return nil, fmt.Errorf("scene.circle_value_wrap")
		}
		advance := l.Lines[0].Advance
		visibleH := l.EstimatedOccupiedHeight
		outerW := math.Max(advance+24, advance/.62)
		outerH := math.Max(visibleH+24, visibleH/.55)
		cx := b.X + advance/2
		cy := b.Y + l.OccupiedTop + visibleH/2
		mark := Rect{cx - outerW/2, cy - outerH/2, outerW, outerH}
		for i, item := range p.Items {
			if item.Text == nil || item.Text.ID != id+".label" {
				continue
			}
			dy := math.Max(0, mark.Y+mark.H+6-item.Text.Rect.Y)
			if n.H > 0 && end+dy > b.Y+b.H+.02 {
				return nil, fmt.Errorf("scene.circled_metric_label_clearance_overflow: %s", id)
			}
			shiftSceneDataItems(p.Items[i:], dy)
			if n.H == 0 {
				b.H += dy
			}
			break
		}
		im, e := r.primitiveArtworkImage(id+".circle", "circle", mark, surface, "mark")
		if e != nil {
			return nil, e
		}
		p.Items = append([]sceneItem{{Image: im}}, p.Items...)
		sceneDataBounds(p, mark)
		p.Warnings = append(p.Warnings, "user-review.v1/circled-metric-clearance: "+id+" ring centered on measured value envelope with12pt minimum clearance and ellipse-safe interior ratios.")
	}
	sceneDataBounds(p, b)
	sceneDataGroup(p, id, "data.metric.source", 0, p.Bounds)
	return p, nil
}

func (r *renderer) sceneFeeSummary(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	var n struct {
		Type       string            `json:"type"`
		X          float64           `json:"x"`
		Y          float64           `json:"y"`
		W          float64           `json:"w"`
		Model      string            `json:"model"`
		Total      string            `json:"total"`
		TotalLabel string            `json:"totalLabel"`
		Lines      [][]string        `json:"lines"`
		Terms      []json.RawMessage `json:"terms"`
	}
	if e := sceneDecode(raw, &n); e != nil {
		return nil, e
	}
	p := &scenePlan{ID: id}
	b := Rect{n.X, n.Y, n.W, 1}
	if e := r.sceneRect(p, id+".container", b, "inverse"); e != nil {
		return nil, e
	}
	x, y, w := n.X+18, n.Y+18, n.W-36
	ls, _ := r.sceneDataToken("label", 600)
	ll, e := r.typeEngine.Measure(n.Model, ls, w-12)
	if e != nil {
		return nil, e
	}
	tw := ll.Lines[0].Advance + 14
	if e = r.sceneRect(p, id+".model-bg", Rect{x, y, tw, ls.Leading + 4}, "light"); e != nil {
		return nil, e
	}
	tr, e := r.sceneDataText(p, id+".model", n.Model, ls, Rect{x + 6, y + 2, tw - 12, 0}, "light", "display", "left", ctx)
	if e != nil {
		return nil, e
	}
	y = tr.Rect.Y + tr.Rect.H + 8
	st, _ := r.sceneDataToken("stat", 0)
	tr, e = r.sceneDataText(p, id+".total", n.Total, st, Rect{x, y, w, 0}, "inverse", "display", "left", ctx)
	if e != nil {
		return nil, e
	}
	y = tr.Rect.Y + tr.Rect.H + 6
	st, _ = r.sceneDataToken("small", 0)
	tr, e = r.sceneDataText(p, id+".total-label", n.TotalLabel, st, Rect{x, y, w, 0}, "inverse", "secondary", "left", ctx)
	if e != nil {
		return nil, e
	}
	y = tr.Rect.Y + tr.Rect.H + 15
	if e = r.sceneDataInkShape(p, id+".divider", Rect{x, y, w, .75}, "inverse", "line"); e != nil {
		return nil, e
	}
	y += 9.75
	for i, line := range n.Lines {
		if len(line) != 2 {
			return nil, fmt.Errorf("scene.fee_line_requires_pair")
		}
		key, e := sceneDataKey(ctx, "lines", i)
		if e != nil {
			return nil, e
		}
		vs := st
		vs.ID = "source.fee-line.mono.12.18"
		vs.Family = "IBM Plex Mono"
		vs.Weight = 600
		vl, e := r.typeEngine.Measure(line[1], vs, w)
		if e != nil {
			return nil, e
		}
		vw := vl.Lines[0].Advance + 1
		lt, e := r.sceneDataText(p, id+".line."+key+".label", line[0], st, Rect{x, y, w - vw - 9, 0}, "inverse", "primary", "left", ctx)
		if e != nil {
			return nil, e
		}
		vt, e := r.sceneDataText(p, id+".line."+key+".value", line[1], vs, Rect{x + w - vw, y, vw, 0}, "inverse", "display", "right", ctx)
		if e != nil {
			return nil, e
		}
		y = math.Max(lt.Rect.Y+lt.Rect.H, vt.Rect.Y+vt.Rect.H) + 6
	}
	y += 9
	tr, e = r.sceneDataText(p, id+".terms-label", "Terms", ls, Rect{x, y, w, 0}, "inverse", "emphasis", "left", ctx)
	if e != nil {
		return nil, e
	}
	y = tr.Rect.Y + tr.Rect.H + 6
	end, e := r.sceneDataBullets(p, id+".terms", n.Terms, ctx, "terms", Rect{x, y, w, 0}, "inverse", st)
	if e != nil {
		return nil, e
	}
	b.H = end + 18 - b.Y
	p.Items[0].Shape.Record.Rect = b
	p.Items[0].Shape.Props.PositionProps = pos(b)
	p.Bounds = b
	sceneDataGroup(p, id, "pricing.fee-summary", 0, b)
	return p, nil
}

func (r *renderer) sceneDataIcon(id string, b Rect, name, color string) (*scenePlan, error) {
	return r.planIconScene(id, name, b.W, b, "light", "#"+color)
}
