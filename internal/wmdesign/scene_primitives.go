package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

const PrimitiveSceneContract = "pptxgengo.wmds-source-primitives.v1"

type primitiveSource struct {
	Type      string            `json:"type"`
	ID        string            `json:"id,omitempty"`
	X         float64           `json:"x"`
	Y         float64           `json:"y"`
	W         float64           `json:"w"`
	H         float64           `json:"h,omitempty"`
	Style     string            `json:"style,omitempty"`
	Ink       string            `json:"ink,omitempty"`
	On        string            `json:"on,omitempty"`
	Text      string            `json:"text,omitempty"`
	Align     string            `json:"align,omitempty"`
	VAlign    string            `json:"valign,omitempty"`
	Emphasis  string            `json:"emphasis,omitempty"`
	MarkInk   string            `json:"markInk,omitempty"`
	Variant   json.RawMessage   `json:"variant,omitempty"`
	Label     string            `json:"label,omitempty"`
	Title     string            `json:"title,omitempty"`
	Body      string            `json:"body,omitempty"`
	N         string            `json:"n,omitempty"`
	NumInk    string            `json:"numInk,omitempty"`
	NumStyle  string            `json:"numStyle,omitempty"`
	Size      string            `json:"size,omitempty"`
	Rule      string            `json:"rule,omitempty"`
	Weight    float64           `json:"weight,omitempty"`
	By        string            `json:"by,omitempty"`
	Items     []json.RawMessage `json:"items,omitempty"`
	RowHeight float64           `json:"rowHeight,omitempty"`
	KeyW      float64           `json:"keyW,omitempty"`
	KeyInk    string            `json:"keyInk,omitempty"`
	K         string            `json:"k,omitempty"`
}

var primitiveFields = map[string]string{
	"text": "style ink on text align valign emphasis markInk variant h weight", "textblock": "label title body h", "bullets": "items size on k label", "ol": "items size on", "list": "items variant rowHeight", "schedule": "items keyInk keyW rowHeight", "grouplabel": "text", "numhead": "n text numInk ink", "colhead": "rule weight label title", "strongnum": "items numInk numStyle size", "pullquote": "text by markInk",
}

func primitiveDecode(raw json.RawMessage, target any, allowed string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	allow := map[string]bool{}
	for _, f := range strings.Fields("type id x y w " + allowed) {
		allow[f] = true
	}
	for f := range fields {
		if !allow[f] {
			return fmt.Errorf("scene.unsupported_source_field: %s", f)
		}
	}
	return sceneDecode(raw, target)
}
func primitiveArrayKeys(ctx SceneContext, suffix string, count int) ([]string, error) {
	path := ctx.Path + suffix
	if keys, ok := ctx.Keys[path]; ok {
		if len(keys) != count {
			return nil, fmt.Errorf("scene.key_count: %s", path)
		}
		seen := map[string]bool{}
		for _, k := range keys {
			if !validPartKey(k) || seen[k] {
				return nil, fmt.Errorf("scene.invalid_or_duplicate_key: %s", path)
			}
			seen[k] = true
		}
		return keys, nil
	}
	if ctx.Keys != nil {
		return nil, fmt.Errorf("scene.missing_array_keys: %s", path)
	}
	// Source references retain pinned source ordinals. Bound arrays must provide
	// a key overlay at their exact source-relative array path.
	keys := make([]string, count)
	for i := range keys {
		keys[i] = fmt.Sprintf("source-%03d", i+1)
	}
	return keys, nil
}
func primitiveFinish(p *scenePlan, kind string) {
	rec := ComponentRecord{ID: p.ID, Definition: kind, Contract: PrimitiveSceneContract, Rect: p.Bounds, Density: "source", Limits: map[string]int{}}
	for _, item := range p.Items {
		if item.Text != nil {
			rec.Parts = append(rec.Parts, item.Text.ID)
			rec.RequiredHeight = math.Max(rec.RequiredHeight, item.Text.Rect.Y+item.Text.Rect.H-p.Bounds.Y)
		}
		if item.Shape != nil {
			rec.Parts = append(rec.Parts, item.Shape.Record.ID)
			rec.Shapes = append(rec.Shapes, item.Shape.Record)
		}
		if item.Image != nil {
			rec.Parts = append(rec.Parts, item.Image.ObjectName)
		}
	}
	if p.Bounds.H == 0 {
		p.Bounds.H = rec.RequiredHeight
		rec.Rect = p.Bounds
	}
	p.Groups = append(p.Groups, rec)
}
func (r *renderer) primitiveShape(p *scenePlan, id string, b Rect, surface, ink string, outline bool) error {
	c, e := r.sceneColor(surface, ink)
	if e != nil {
		return e
	}
	if b.W <= 0 || b.H <= 0 {
		return fmt.Errorf("scene.invalid_shape_geometry: %s", id)
	}
	props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Color: c}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}
	if outline {
		props.Fill = &pptx.ShapeFillProps{Type: "none"}
		props.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: c}, Width: .75}
	}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeRect, Props: props, Record: ShapeRecord{ID: id, Rect: b, Color: c}}})
	return nil
}
func (r *renderer) planPrimitiveScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var head struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &head); e != nil {
		return nil, false, e
	}
	allowed, ok := primitiveFields[head.Type]
	if !ok {
		return nil, false, nil
	}
	var n primitiveSource
	if e := primitiveDecode(raw, &n, allowed); e != nil {
		return nil, true, e
	}
	if n.W <= 0 || n.H < 0 || math.IsNaN(n.X+n.Y+n.W+n.H) || math.IsInf(n.X+n.Y+n.W+n.H, 0) {
		return nil, true, fmt.Errorf("scene.invalid_primitive_geometry: %s", id)
	}
	surface := ctx.Surface
	if n.On != "" {
		surface = n.On
	}
	if surface == "" {
		surface = "light"
	}
	p := &scenePlan{ID: id, Bounds: Rect{X: n.X, Y: n.Y, W: n.W, H: n.H}}
	y := n.Y
	add := func(part, text, token, ink string, x, w, h float64, weight int) error {
		if text == "" {
			return fmt.Errorf("scene.missing_content: %s.%s", id, part)
		}
		st, e := r.sceneStyle(token)
		if e != nil {
			return e
		}
		if weight != 0 {
			st.Weight = weight
		}
		e = r.primitiveRichText(p, id+"."+part, text, st, Rect{X: x, Y: y, W: w, H: h}, surface, ink, "left", "", "", ctx)
		if e == nil {
			t := p.Items[len(p.Items)-1].Text
			y = t.Rect.Y + st.Leading*float64(len(t.Layout.Lines))
		}
		return e
	}
	var err error
	switch n.Type {
	case "text":
		if n.Style == "" || n.Text == "" {
			err = fmt.Errorf("scene.missing_text_style_or_content: %s", id)
			break
		}
		if n.VAlign != "" && n.H <= 0 {
			err = fmt.Errorf("scene.valign_requires_height: %s", id)
			break
		}
		st, e := r.sceneStyle(n.Style)
		if e != nil {
			err = e
			break
		}
		if n.Weight != 0 {
			switch n.Weight {
			case 400, 500, 600, 700:
				st.Weight = int(n.Weight)
			default:
				err = fmt.Errorf("scene.unsupported_text_weight: %s", id)
				break
			}
			if err != nil {
				break
			}
		}
		if n.VAlign != "" && n.VAlign != "top" && n.VAlign != "middle" && n.VAlign != "bottom" {
			err = fmt.Errorf("scene.unsupported_valign: %s", n.VAlign)
			break
		}
		variant := 1
		if len(n.Variant) > 0 {
			var v int
			if e := json.Unmarshal(n.Variant, &v); e == nil {
				if v < 1 || v > 4 {
					err = fmt.Errorf("scene.invalid_highlight_variant: %s", id)
					break
				}
				variant = v
			} else {
				var mode string
				if e = json.Unmarshal(n.Variant, &mode); e != nil || mode != "auto" {
					err = fmt.Errorf("scene.invalid_highlight_variant: %s", id)
					break
				}
				variant = 0
			}
		}
		err = r.primitiveRichTextVariant(p, id+".text", n.Text, st, p.Bounds, surface, n.Ink, n.Align, n.Emphasis, n.MarkInk, ctx, variant)
		if err == nil && n.H > 0 && n.VAlign != "" && n.VAlign != "top" {
			t := p.Items[len(p.Items)-1].Text
			need := math.Max(t.Layout.AllocationHeight, t.Layout.OccupiedTop+t.Layout.EstimatedOccupiedHeight)
			shift := n.H - need
			if n.VAlign == "middle" {
				shift /= 2
			}
			for i := range p.Items {
				if p.Items[i].Text != nil {
					p.Items[i].Text.Rect.Y += shift
					p.Items[i].Text.Rect.H = need
				}
				if p.Items[i].Image != nil {
					p.Items[i].Image.Y.Val += shift / 72
				}
			}
		}
	case "textblock":
		if n.Body == "" {
			err = fmt.Errorf("scene.missing_textblock_body: %s", id)
			break
		}
		for _, part := range []struct{ k, text, st, ink string }{{"label", n.Label, "label", "emphasis"}, {"title", n.Title, "subhead", "display"}, {"body", n.Body, "body", "secondary"}} {
			if part.text == "" {
				continue
			}
			if err = add(part.k, part.text, part.st, part.ink, n.X, n.W, 0, 0); err != nil {
				break
			}
			y += 6
		}
		if len(p.Items) == 0 {
			err = fmt.Errorf("scene.empty_textblock: %s", id)
		}
		if err == nil && n.H > 0 {
			for _, item := range p.Items {
				if item.Text != nil && item.Text.Rect.Y+item.Text.Rect.H > n.Y+n.H+.02 {
					err = fmt.Errorf("scene.textblock_overflow: %s bottom %.3fpt, capacity %.3fpt", item.Text.ID, item.Text.Rect.Y+item.Text.Rect.H, n.Y+n.H)
					break
				}
			}
		}
	case "colhead":
		weight := n.Weight
		if weight == 0 {
			weight = 3
		}
		err = r.primitiveShape(p, id+".rule", Rect{X: n.X, Y: y, W: n.W, H: weight}, surface, n.Rule, false)
		y += weight + 9
		if err == nil && n.Label != "" {
			err = add("label", n.Label, "label", "emphasis", n.X, n.W, 0, 0)
		}
		if err == nil {
			err = add("title", n.Title, "subhead", "display", n.X, n.W, 0, 0)
		}
	case "numhead":
		st, e := r.sceneStyle("number")
		if e != nil {
			err = e
			break
		}
		l, e := r.typeEngine.Measure(n.N, st, n.W)
		if e != nil {
			err = e
			break
		}
		nw := 0.
		for _, line := range l.Lines {
			nw = math.Max(nw, line.Advance)
		}
		err = add("number", n.N, "number", n.NumInk, n.X, nw+.1, 0, 0)
		y = n.Y
		if err == nil {
			err = add("text", n.Text, "heading", n.Ink, n.X+nw+9, n.W-nw-9, 0, 0)
		}
		if err == nil {
			base := 0.
			for _, item := range p.Items {
				if item.Text != nil {
					base = math.Max(base, item.Text.Layout.Lines[0].Baseline)
				}
			}
			for i := range p.Items {
				if t := p.Items[i].Text; t != nil {
					t.Rect.Y += base - t.Layout.Lines[0].Baseline
				}
			}
		}
	case "grouplabel":
		st, e := r.sceneStyle("label")
		if e != nil {
			err = e
			break
		}
		l, e := r.typeEngine.Measure(n.Text, st, n.W)
		if e != nil {
			err = e
			break
		}
		if len(l.Lines) != 1 {
			err = fmt.Errorf("scene.group_label_wrap: %s", id)
			break
		}
		// Native PowerPoint may retain the final tracked glyph spacing where
		// the Go advance does not. Keep the measured label and following rule
		// at their authored positions, with a small box reserve inside the gap.
		labelW := l.Lines[0].Advance + 2
		if labelW > n.W {
			err = fmt.Errorf("scene.group_label_native_reserve: %s needs %.3fpt, capacity %.3fpt", id, labelW, n.W)
			break
		}
		err = add("label", n.Text, "label", "secondary", n.X, labelW, 18, 0)
		if err == nil {
			err = r.primitiveShape(p, id+".rule", Rect{X: n.X + l.Lines[0].Advance + 12, Y: n.Y + 8.625, W: n.W - l.Lines[0].Advance - 12, H: .75}, surface, "line", false)
			p.Warnings = append(p.Warnings, "wmds.grouplabel-native-width.v1: "+id+" reserves2pt within the label-to-rule gap; source font and rule endpoint preserved.")
		}
	case "pullquote":
		st, e := r.sceneStyle("heading")
		if e != nil {
			err = e
			break
		}
		st, e = primitiveStyleSize(st, 60)
		if e != nil {
			err = e
			break
		}
		st.Leading = 30
		ink := n.MarkInk
		if ink == "" {
			ink = "emphasis"
			if surface == "inverse" || surface == "deep" {
				ink = "display"
			}
		}
		err = r.primitiveRichText(p, id+".quote-mark", "“", st, Rect{X: n.X, Y: y, W: n.W}, surface, ink, "left", "", "", ctx)
		y += 39
		if err == nil {
			err = add("quote", n.Text, "heading", "display", n.X, n.W, 0, 0)
			y += 9
		}
		if err == nil {
			err = add("by", n.By, "label", "secondary", n.X, n.W, 0, 0)
		}
	case "bullets", "ol":
		if n.Size != "" && n.Size != "small" && n.Size != "body" {
			err = fmt.Errorf("scene.unsupported_bullet_size: %s", n.Size)
			break
		}
		if n.K != "" || n.Label != "" {
			p.Warnings = append(p.Warnings, "source.browser_ignored_bullets_metadata: k/label consumed as metadata")
		}
		err = r.primitiveBullets(p, id, n.Items, Rect{X: n.X, Y: n.Y, W: n.W}, surface, n.Size, ctx, "/items", n.Type == "ol", 0)
	case "strongnum":
		keys, e := primitiveArrayKeys(ctx, "/items", len(n.Items))
		if e != nil {
			err = e
			break
		}
		if len(n.Items) == 0 {
			err = fmt.Errorf("scene.empty_strongnum: %s", id)
			break
		}
		if n.NumStyle != "" && n.NumStyle != "number" && n.NumStyle != "stat-sm" {
			err = fmt.Errorf("scene.unsupported_num_style: %s", n.NumStyle)
			break
		}
		token := "body"
		if n.Size == "small" {
			token = "small"
		} else if n.Size != "" && n.Size != "body" {
			err = fmt.Errorf("scene.unsupported_size: %s", n.Size)
			break
		}
		col := 30.
		nst, e := r.sceneStyle("number")
		if n.NumStyle == "stat-sm" {
			col = 54
			nst, e = r.sceneStyle("stat-sm")
		} else {
			nst.Size = 14
			nst.Leading = 21
			if token == "small" {
				nst.Size = 12
				nst.Leading = 18
			}
		}
		if e != nil {
			err = e
			break
		}
		nst, e = primitiveStyleSize(nst, nst.Size)
		if e != nil {
			err = e
			break
		}
		for i, item := range n.Items {
			var it struct {
				Title string `json:"title"`
				Text  string `json:"text"`
			}
			if err = sceneDecode(item, &it); err != nil {
				break
			}
			start := y
			key := keys[i]
			if err = r.primitiveRichText(p, id+"."+key+".number", fmt.Sprintf("%02d", i+1), nst, Rect{X: n.X, Y: y, W: col}, surface, n.NumInk, "left", "", "", ctx); err != nil {
				break
			}
			if err = add(key+".title", it.Title, token, "primary", n.X+col+12, n.W-col-12, 0, 600); err != nil {
				break
			}
			if err = add(key+".text", it.Text, token, "secondary", n.X+col+12, n.W-col-12, 0, 0); err != nil {
				break
			}
			y = math.Max(y, start+nst.Leading) + 12
		}
	case "schedule", "list":
		if n.RowHeight <= 0 || len(n.Items) == 0 {
			err = fmt.Errorf("scene.invalid_row_capacity: %s", id)
			break
		}
		keys, e := primitiveArrayKeys(ctx, "/items", len(n.Items))
		if e != nil {
			err = e
			break
		}
		if n.Type == "list" {
			var v string
			if e = json.Unmarshal(n.Variant, &v); e != nil || v != "index" {
				err = fmt.Errorf("scene.unsupported_list_variant: %s", id)
				break
			}
		} else if n.KeyW <= 0 || n.KeyW+18 >= n.W {
			err = fmt.Errorf("scene.invalid_schedule_columns: %s", id)
			break
		}
		for i, item := range n.Items {
			y = n.Y + float64(i)*n.RowHeight
			key := keys[i]
			if n.Type == "schedule" {
				var it struct {
					K string `json:"k"`
					T string `json:"t"`
				}
				if err = sceneDecode(item, &it); err != nil {
					break
				}
				if it.K == "" || it.T == "" {
					err = fmt.Errorf("scene.missing_schedule_row_content: %s.%s", id, key)
					break
				}
				st, e := r.sceneStyle("body")
				if e != nil {
					err = e
					break
				}
				st.Weight = 700
				err = r.primitiveRichText(p, id+"."+key+".key", it.K, st, Rect{X: n.X, Y: y, W: n.KeyW, H: n.RowHeight}, surface, n.KeyInk, "right", "", "", ctx)
				if err == nil {
					err = add(key+".text", it.T, "body", "display", n.X+n.KeyW+18, n.W-n.KeyW-18, n.RowHeight, 600)
				}
			} else {
				var it struct {
					N     string `json:"n"`
					Title string `json:"title"`
					Page  string `json:"page"`
				}
				if err = sceneDecode(item, &it); err != nil {
					break
				}
				if it.N == "" || it.Title == "" || it.Page == "" {
					err = fmt.Errorf("scene.missing_index_row_content: %s.%s", id, key)
					break
				}
				st, e := r.sceneStyle("label")
				if e != nil {
					err = e
					break
				}
				st, e = primitiveStyleSize(st, 11)
				if e != nil {
					err = e
					break
				}
				st.Weight = 600
				err = r.primitiveRichText(p, id+"."+key+".number", it.N, st, Rect{X: n.X, Y: y, W: 72, H: n.RowHeight}, surface, "emphasis", "left", "", "", ctx)
				if err == nil {
					err = add(key+".title", it.Title, "subhead", "display", n.X+72, n.W-126, n.RowHeight, 0)
				}
				y = n.Y + float64(i)*n.RowHeight
				if err == nil {
					st, e = r.sceneStyle("label")
					if e != nil {
						err = e
						break
					}
					err = r.primitiveRichText(p, id+"."+key+".page", "p. "+it.Page, st, Rect{X: n.X + n.W - 54, Y: y, W: 54, H: n.RowHeight}, surface, "secondary", "right", "", "", ctx)
				}
				if err == nil {
					err = r.primitiveShape(p, id+"."+key+".rule", Rect{X: n.X, Y: y + n.RowHeight - .75, W: n.W, H: .75}, surface, "line", false)
				}
			}
			if err != nil {
				break
			}
		}
		if err == nil {
			for i := 0; i < len(n.Items); i++ {
				prefix := id + "." + keys[i] + "."
				rowY := n.Y + float64(i)*n.RowHeight
				baseline := 0.
				if n.Type == "schedule" {
					for _, item := range p.Items {
						if item.Text != nil && strings.HasPrefix(item.Text.ID, prefix) {
							baseline = math.Max(baseline, item.Text.Layout.Lines[0].Baseline)
						}
					}
				}
				for j := range p.Items {
					t := p.Items[j].Text
					if t == nil || !strings.HasPrefix(t.ID, prefix) {
						continue
					}
					need := math.Max(t.Layout.AllocationHeight, t.Layout.OccupiedTop+t.Layout.EstimatedOccupiedHeight)
					shift := (n.RowHeight - need) / 2
					if n.Type == "schedule" {
						shift = baseline - t.Layout.Lines[0].Baseline
					}
					if shift+need > n.RowHeight+.02 {
						err = fmt.Errorf("scene.row_text_overflow: %s", t.ID)
						break
					}
					t.Rect.Y = rowY + shift
					t.Rect.H = need
				}
			}
		}
		p.Bounds.H = float64(len(n.Items)) * n.RowHeight
	}
	if err != nil {
		return nil, true, err
	}
	primitiveFinish(p, n.Type)
	return p, true, nil
}

func (r *renderer) primitiveBullets(p *scenePlan, id string, items []json.RawMessage, b Rect, surface, size string, ctx SceneContext, path string, ordered bool, depth int) error {
	if len(items) == 0 {
		return fmt.Errorf("scene.empty_bullets: %s", id)
	}
	if depth > 8 {
		return fmt.Errorf("scene.bullet_nesting_limit: %s", id)
	}
	keys, e := primitiveArrayKeys(ctx, path, len(items))
	if e != nil {
		return e
	}
	token := "body"
	gap, indent, dot, top := 6., 15., 4., 8.5
	if size == "small" {
		token = "small"
		gap, indent, dot, top = 3, 12, 3, 7
	}
	st, e := r.sceneStyle(token)
	if e != nil {
		return e
	}
	y := b.Y
	for i, raw := range items {
		var text, lead string
		var sub []json.RawMessage
		if len(raw) > 0 && raw[0] == '"' {
			if e = json.Unmarshal(raw, &text); e != nil {
				return e
			}
		} else {
			var it struct {
				Lead string            `json:"lead,omitempty"`
				Text string            `json:"text"`
				Sub  []json.RawMessage `json:"sub,omitempty"`
			}
			if e = sceneDecode(raw, &it); e != nil {
				return e
			}
			text, lead, sub = it.Text, it.Lead, it.Sub
		}
		if text == "" {
			return fmt.Errorf("scene.missing_bullet_content: %s", id)
		}
		part := id + "." + keys[i]
		if ordered {
			ns, e := r.sceneStyle(token)
			if e != nil {
				return e
			}
			if e = r.primitiveRichText(p, part+".number", strconv.Itoa(i+1)+".", ns, Rect{X: b.X, Y: y, W: indent}, surface, "primary", "left", "", "", ctx); e != nil {
				return e
			}
		} else if e = r.primitiveShape(p, part+".bullet", Rect{X: b.X, Y: y + top, W: dot, H: dot}, surface, "primary", depth > 0); e != nil {
			return e
		}
		var rec TextRecord
		if lead == "" {
			if e = r.primitiveRichText(p, part+".text", text, st, Rect{X: b.X + indent, Y: y, W: b.W - indent}, surface, "primary", "left", "", "", ctx); e != nil {
				return e
			}
			rec = *p.Items[len(p.Items)-1].Text
		} else {
			runs := []primitiveRun{{Text: lead, Weight: 600, Ink: "primary"}, {Text: " " + text, Ink: "secondary"}}
			rec, e = r.primitiveMeasureRuns(part+".text", runs, st, Rect{X: b.X + indent, Y: y, W: b.W - indent}, surface, "primary", "left")
			if e != nil {
				return e
			}
			p.Items = append(p.Items, sceneItem{Text: &rec})
		}
		y += rec.Layout.AllocationHeight
		if len(sub) > 0 {
			before := len(p.Items)
			if e = r.primitiveBullets(p, part+".sub", sub, Rect{X: b.X + indent, Y: y + 6, W: b.W - indent}, surface, size, ctx, path+"/"+strconv.Itoa(i)+"/sub", false, depth+1); e != nil {
				return e
			}
			for _, it := range p.Items[before:] {
				if it.Text != nil {
					y = math.Max(y, it.Text.Rect.Y+it.Text.Layout.AllocationHeight)
				}
			}
		}
		y += gap
	}
	return nil
}

// primitiveRun is the source-derived rich run, including an explicit baseline
// shift for footnotes. It is separate from public authored rich-run overrides.
type primitiveRun struct {
	Text     string
	Weight   int
	Ink      string
	Footnote int
	Mark     string
}

func primitiveMarkup(text, emphasis string, notes []string) ([]primitiveRun, error) {
	if emphasis != "" && emphasis != "highlight" && emphasis != "underscore" && emphasis != "circle" && emphasis != "spark" {
		return nil, fmt.Errorf("scene.unsupported_emphasis: %s", emphasis)
	}
	var runs []primitiveRun
	marked := false
	for len(text) > 0 {
		next := len(text)
		kind := ""
		for _, s := range []string{"[[", "]]", "[^"} {
			if i := strings.Index(text, s); i >= 0 && i < next {
				next = i
				kind = s
			}
		}
		if next > 0 {
			mark := ""
			if marked {
				mark = emphasis
			}
			runs = append(runs, primitiveRun{Text: text[:next], Mark: mark})
			text = text[next:]
		}
		if kind == "" {
			break
		}
		switch kind {
		case "[[":
			if marked {
				return nil, fmt.Errorf("scene.nested_inline_mark")
			}
			marked = true
			text = text[2:]
		case "]]":
			if !marked {
				return nil, fmt.Errorf("scene.unmatched_inline_mark")
			}
			marked = false
			text = text[2:]
		case "[^":
			end := strings.IndexByte(text, ']')
			if end < 0 {
				return nil, fmt.Errorf("scene.unclosed_footnote")
			}
			n, e := strconv.Atoi(text[2:end])
			if e != nil || n < 1 || n > len(notes) || strings.TrimSpace(notes[n-1]) == "" {
				return nil, fmt.Errorf("scene.unresolved_footnote: %s", text[:end+1])
			}
			runs = append(runs, primitiveRun{Text: strconv.Itoa(n), Footnote: n, Ink: "emphasis"})
			text = text[end+1:]
		}
	}
	if marked {
		return nil, fmt.Errorf("scene.unclosed_inline_mark")
	}
	return runs, nil
}
func (r *renderer) primitiveMeasureRuns(id string, runs []primitiveRun, st Style, b Rect, surface, ink, align string) (TextRecord, error) {
	var rec TextRecord
	if align == "" {
		align = "left"
	}
	if align != "left" && align != "center" && align != "right" {
		return rec, fmt.Errorf("scene.unsupported_align: %s", align)
	}
	color, e := r.sceneColor(surface, ink)
	if e != nil {
		return rec, e
	}
	base, e := r.typeEngine.Resolve(st)
	if e != nil {
		return rec, e
	}
	q := RichParagraphLayout{Key: "source-paragraph"}
	spacer := func(key string, width float64) error {
		if width <= 0 {
			return nil
		}
		ss := st
		ss.TrackingPt = 0
		layout, e := r.typeEngine.Measure(" ", ss, 8191)
		if e != nil {
			return e
		}
		ss.TrackingPt = math.Round((width-layout.Lines[0].Advance)*100) / 100
		start := len([]rune(q.Displayed))
		q.Runs = append(q.Runs, RichRunLayout{Key: key, Displayed: " ", Style: ss, Font: base, Color: color, Start: start, End: start + 1})
		q.Displayed += " "
		return nil
	}
	original := ""
	for i, run := range runs {
		if run.Text == "" {
			continue
		}
		rs := st
		if run.Weight != 0 {
			rs.Weight = run.Weight
		}
		shift := 0.
		if run.Footnote != 0 {
			rs.Size = st.Size * .6
			rs.Weight = 600
			shift = st.Size * .35
		}
		font, e := r.typeEngine.Resolve(rs)
		if e != nil {
			return rec, e
		}
		c := color
		if run.Ink != "" {
			c, e = r.sceneColor(surface, run.Ink)
			if e != nil {
				return rec, e
			}
		}
		original += run.Text
		before, after := 0., 0.
		switch run.Mark {
		case "highlight":
			before = .05 * st.Size
			after = before
		case "circle":
			before = .15 * st.Size
			after = before
		case "spark":
			after = .75 * st.Size
		}
		if run.Footnote != 0 {
			before += .06 * st.Size
		}
		if e := spacer(fmt.Sprintf("source-run-%03d-margin-before", i+1), before); e != nil {
			return rec, e
		}
		display := run.Text
		if run.Mark == "circle" || run.Mark == "spark" {
			display = strings.ReplaceAll(display, " ", "\u00a0")
		}
		if st.Case == "upper" {
			display = strings.ToUpper(display)
		}
		start := len([]rune(q.Displayed))
		q.Runs = append(q.Runs, RichRunLayout{Key: fmt.Sprintf("source-run-%03d", i+1), Original: run.Text, Displayed: display, Style: rs, Font: font, Color: c, Start: start, End: start + len([]rune(display)), BaselineShift: shift})
		q.Displayed += display
		if e := spacer(fmt.Sprintf("source-run-%03d-margin-after", i+1), after); e != nil {
			return rec, e
		}
	}
	if strings.TrimSpace(q.Displayed) == "" {
		return rec, fmt.Errorf("scene.empty_text: %s", id)
	}
	// Validate every run through the same Go font engine before combined wrapping.
	for _, rr := range q.Runs {
		if _, e = r.typeEngine.Measure(rr.Displayed, rr.Style, 8191); e != nil {
			return rec, e
		}
	}
	lines, e := r.typeEngine.richParagraphLines(q, b.W)
	if e != nil {
		return rec, fmt.Errorf("%s: %w", id, e)
	}
	baseline, terminal := 0., 0.
	allKnown := true
	for _, rr := range q.Runs {
		if rr.BaselineShift != 0 {
			continue
		}
		a, known := r.typeEngine.anchors[anchorKey(rr.Font.SHA256, rr.Style.Size, rr.Style.Leading)]
		if known {
			baseline = math.Max(baseline, a.Baseline)
			terminal = math.Max(terminal, a.TerminalHeight)
		} else {
			allKnown = false
			baseline = math.Max(baseline, st.Leading*.75)
			terminal = math.Max(terminal, math.Max(st.Leading, st.Size*1.5))
		}
	}
	for i := range lines {
		lines[i].Baseline = baseline + float64(i)*st.Leading
	}
	q.LineCount = len(lines)
	l := TextLayout{Original: original, Displayed: q.Displayed, Style: st, Font: base, Lines: lines, AllocationHeight: float64(len(lines)) * st.Leading, EstimatedOccupiedHeight: float64(len(lines)-1)*st.Leading + terminal, Features: map[string]int{"kern": 0, "liga": 1, "clig": 1}, VerticalPolicy: "source rich text: measured combined Go runs; native mixed-size baseline unqualified"}
	if allKnown {
		l.CalibrationSHA256 = CandidateCalibrationSHA
	}
	need := math.Max(l.AllocationHeight, l.EstimatedOccupiedHeight)
	if b.H == 0 {
		b.H = need
	}
	if need > b.H+.02 {
		return rec, fmt.Errorf("scene.text_vertical_overflow: %s needs %.3fpt, capacity %.3fpt", id, need, b.H)
	}
	return TextRecord{ID: id, Rect: b, Color: color, Align: align, Layout: l, Rich: &RichTextLayout{Contract: RichTextContract, Paragraphs: []RichParagraphLayout{q}}}, nil
}
func (r *renderer) primitiveRichText(p *scenePlan, id, text string, st Style, b Rect, surface, ink, align, emphasis, markInk string, ctx SceneContext) error {
	return r.primitiveRichTextVariant(p, id, text, st, b, surface, ink, align, emphasis, markInk, ctx, 1)
}
func (r *renderer) primitiveRichTextVariant(p *scenePlan, id, text string, st Style, b Rect, surface, ink, align, emphasis, markInk string, ctx SceneContext, variant int) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("scene.empty_text: %s", id)
	}
	if emphasis != "" && emphasis != "highlight" && emphasis != "underscore" && emphasis != "circle" && emphasis != "spark" {
		return fmt.Errorf("scene.unsupported_emphasis: %s", emphasis)
	}
	if !strings.Contains(text, "[[") && !strings.Contains(text, "]]") && !strings.Contains(text, "[^") {
		return r.sceneText(p, id, text, st, b, surface, ink, align)
	}
	runs, e := primitiveMarkup(text, emphasis, ctx.Notes)
	if e != nil {
		return fmt.Errorf("%s: %w", id, e)
	}
	rec, e := r.primitiveMeasureRuns(id, runs, st, b, surface, ink, align)
	if e != nil {
		return e
	}
	if emphasis == "highlight" && (surface == "inverse" || surface == "deep" || surface == "callout") {
		return fmt.Errorf("scene.highlight_disallowed_surface: %s", surface)
	}
	q := rec.Rich.Paragraphs[0]
	// Marks are placed from the Go line plan. Explicit native soft breaks keep
	// the marked phrase and its artwork on those same lines when PowerPoint's
	// font selection would otherwise move a nearby word across the boundary.
	if !strings.Contains(q.Displayed, "\n") && len(rec.Layout.Lines) > 1 {
		end := 0
		for _, line := range rec.Layout.Lines[:len(rec.Layout.Lines)-1] {
			end += len([]rune(line.Text))
			rec.Rich.Paragraphs[0].LineBreaks = append(rec.Rich.Paragraphs[0].LineBreaks, end)
		}
		p.Warnings = append(p.Warnings, "wmds.marked-text-explicit-lines.v1: native soft breaks preserve the measured line plan and accent placement")
	}
	global := 0
	for li, line := range rec.Layout.Lines {
		count := len([]rune(line.Text))
		offset := 0.
		if rec.Align == "center" {
			offset = (b.W - line.Advance) / 2
		} else if rec.Align == "right" {
			offset = b.W - line.Advance
		}
		for ri, run := range runs {
			if run.Mark == "" {
				continue
			}
			rr := RichRunLayout{}
			for _, candidate := range q.Runs {
				if candidate.Key == fmt.Sprintf("source-run-%03d", ri+1) {
					rr = candidate
					break
				}
			}
			start, end := max(global, rr.Start), min(global+count, rr.End)
			if end <= start {
				continue
			}
			if (run.Mark == "circle" || run.Mark == "spark") && (start != rr.Start || end != rr.End) {
				return fmt.Errorf("scene.inline_nonwrapping_mark_wrap: %s run %d", id, ri+1)
			}
			prefix := richPrefix(q, start)
			upto := richPrefix(q, end)
			advance := func(part RichParagraphLayout) (float64, error) {
				part.Displayed = string([]rune(part.Displayed)[global:])
				var local []RichRunLayout
				for _, r := range part.Runs {
					if r.End <= global {
						continue
					}
					r.Start = max(r.Start, global)
					r.Start -= global
					r.End -= global
					local = append(local, r)
				}
				part.Runs = local
				if part.Displayed == "" {
					return 0, nil
				}
				lines, e := r.typeEngine.richParagraphLines(part, 8191)
				if e != nil {
					return 0, e
				}
				return lines[0].Advance, nil
			}
			x, e := advance(prefix)
			if e != nil {
				return e
			}
			endX, e := advance(upto)
			if e != nil {
				return e
			}
			mb := Rect{X: b.X + offset + x, Y: b.Y + float64(li)*st.Leading, W: endX - x, H: st.Leading}
			asset := run.Mark
			switch run.Mark {
			case "highlight":
				hv := variant
				if hv == 0 {
					count := 0
					for j := 0; j <= ri; j++ {
						if runs[j].Mark == "highlight" {
							count++
						}
					}
					hv = (count-1)%4 + 1
				}
				asset = "highlight-" + strconv.Itoa(hv)
				mb.X -= .1 * st.Size
				mb.W += .2 * st.Size
				// CSS background-position:center 58% places an 82%-high
				// background at 58% of the remaining 18%, not 18% of the line.
				mb.Y += (1 - .82) * .58 * st.Leading
				mb.H = .82 * st.Leading
			case "underscore":
				mb.X -= .03 * mb.W
				mb.W *= 1.06
				mb.Y += .86 * st.Leading
				mb.H = 0
			case "circle":
				mb.X -= .2 * st.Size
				mb.W += .4 * st.Size
				mb.X -= .16 * mb.W
				mb.W *= 1.32
				mb.Y -= .28 * st.Leading
				mb.H *= 1.5
			case "spark":
				mb.X += mb.W - .55*st.Size
				mb.Y -= .95 * st.Size
				mb.W = 1.5 * st.Size
				mb.H = 0
			}
			col := markInk
			if col == "" {
				col = "mark"
			}
			image, e := r.primitiveArtworkImage(fmt.Sprintf("%s.mark-%03d-%03d", id, ri+1, li+1), asset, mb, surface, col)
			if e != nil {
				return e
			}
			p.Items = append(p.Items, sceneItem{Image: image})
		}
		global += count
	}
	p.Items = append(p.Items, sceneItem{Text: &rec})
	return nil
}

// primitiveStyleSize resolves authored relative letter spacing after a source
// font-size override. Explicit point tracking stays absolute. Footnote inherited
// spacing and the measured margin spacers are intentionally set independently.
func primitiveStyleSize(st Style, size float64) (Style, error) {
	if size <= 0 || math.IsNaN(size) || math.IsInf(size, 0) {
		return st, fmt.Errorf("scene.invalid_source_font_size")
	}
	st.Size = size
	if strings.HasSuffix(st.Tracking, "em") {
		v, e := strconv.ParseFloat(strings.TrimSuffix(st.Tracking, "em"), 64)
		if e != nil {
			return st, fmt.Errorf("scene.invalid_source_tracking: %s", st.Tracking)
		}
		st.TrackingPt = math.Round(v*size*100) / 100
	}
	return st, nil
}

// PrimitiveAssetReference identifies a canonical local dependency actually
// named by a generated picture's provenance, before any declared transforms.
type PrimitiveAssetReference struct {
	Key    string `json:"key"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Crop   [4]int `json:"crop,omitempty"`
}

// UsedPrimitiveAssets reads canonical provenance from generated picture XML.
// An overridden human alt description may omit the registry key; in that case
// all keys for that exact pinned hash are returned. Packaging should deduplicate
// Path+SHA256 (e.g. the full headshot and its face crop share source bytes).
func UsedPrimitiveAssets(raw []byte) ([]PrimitiveAssetReference, error) {
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		return nil, e
	}
	byHash := map[string][]string{}
	for key, a := range primitiveAssetRegistry {
		byHash[a.SHA256] = append(byHash[a.SHA256], key)
	}
	used := map[string]bool{}
	for _, f := range z.File {
		if !strings.HasPrefix(f.Name, "ppt/") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		rc, e := f.Open()
		if e != nil {
			return nil, e
		}
		data, e := io.ReadAll(rc)
		ce := rc.Close()
		if e != nil {
			return nil, e
		}
		if ce != nil {
			return nil, ce
		}
		d := xml.NewDecoder(bytes.NewReader(data))
		for {
			tok, e := d.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				return nil, e
			}
			t, ok := tok.(xml.StartElement)
			if !ok || t.Name.Local != "cNvPr" {
				continue
			}
			descr := ""
			for _, attr := range t.Attr {
				if attr.Name.Local == "descr" {
					descr = attr.Value
				}
			}
			hash := ""
			for _, field := range strings.Split(descr, ";") {
				field = strings.TrimSpace(field)
				if strings.HasPrefix(field, "canonical SHA256=") {
					if hash != "" {
						return nil, fmt.Errorf("scene.asset_ambiguous_picture_provenance")
					}
					hash = strings.TrimPrefix(field, "canonical SHA256=")
				}
			}
			if hash == "" {
				continue
			}
			keys := byHash[hash]
			if len(keys) == 0 {
				return nil, fmt.Errorf("scene.asset_unregistered_picture_hash: %s", hash)
			}
			key := strings.TrimSpace(strings.SplitN(descr, ";", 2)[0])
			if a, ok := primitiveAssetRegistry[key]; ok && a.SHA256 == hash {
				used[key] = true
			} else {
				for _, key := range keys {
					used[key] = true
				}
			}
		}
	}
	keys := make([]string, 0, len(used))
	for key := range used {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]PrimitiveAssetReference, 0, len(keys))
	for _, key := range keys {
		a := primitiveAssetRegistry[key]
		out = append(out, PrimitiveAssetReference{Key: key, Path: a.Path, SHA256: a.SHA256, Crop: a.Crop})
	}
	return out, nil
}
func UsedPrimitiveAssetKeys(raw []byte) ([]string, error) {
	assets, e := UsedPrimitiveAssets(raw)
	if e != nil {
		return nil, e
	}
	keys := make([]string, len(assets))
	for i, a := range assets {
		keys[i] = a.Key
	}
	return keys, nil
}
