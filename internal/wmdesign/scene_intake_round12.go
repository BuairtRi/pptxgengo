package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

const IntakeRound12Contract = "pptxgengo.wmds-source-intake-round12.v1"
const IntakeRound12RendererSHA256 = "3284f866ee67e37d8960a252fbc579ac10f332db1212f7806c98d70d45eacc64"
const IntakeRound12SourceRevision = "wmds-library.v4"

type round12Band struct {
	Label   string          `json:"label,omitempty"`
	Value   json.RawMessage `json:"value,omitempty"`
	Text    string          `json:"text,omitempty"`
	Surface string          `json:"surface,omitempty"`
	Active  bool            `json:"active,omitempty"`
}
type round12Bands struct {
	Type       string        `json:"type"`
	ID         string        `json:"id,omitempty"`
	X          float64       `json:"x"`
	Y          float64       `json:"y"`
	W          float64       `json:"w"`
	H          float64       `json:"h"`
	ShapeW     *float64      `json:"shapeW,omitempty"`
	Neck       *float64      `json:"neck,omitempty"`
	Gap        *float64      `json:"gap,omitempty"`
	Stages     []round12Band `json:"stages,omitempty"`
	Levels     []round12Band `json:"levels,omitempty"`
	LabelSide  bool          `json:"labelSide,omitempty"`
	Ramp       string        `json:"ramp,omitempty"`
	Dark       *bool         `json:"dark,omitempty"`
	Apex       *bool         `json:"apex,omitempty"`
	ValueStyle string        `json:"valueStyle,omitempty"`
	CanvasH    float64       `json:"_h,omitempty"`
}
type round12Bracket struct {
	Type    string   `json:"type"`
	ID      string   `json:"id,omitempty"`
	X       float64  `json:"x"`
	Y       float64  `json:"y"`
	W       float64  `json:"w,omitempty"`
	H       float64  `json:"h,omitempty"`
	Orient  string   `json:"orient,omitempty"`
	Depth   *float64 `json:"depth,omitempty"`
	Label   string   `json:"label,omitempty"`
	Ink     string   `json:"ink,omitempty"`
	CanvasH float64  `json:"_h,omitempty"`
}

// New pyramid syntax is selected by its authored levels/stages arrays. The
// historical bands/mode adapter remains authoritative for every frozen v3 node.
func round12PyramidInput(fields map[string]json.RawMessage) bool {
	_, levels := fields["levels"]
	_, stages := fields["stages"]
	return levels || stages
}
func round12RequiredNumbers(fields map[string]json.RawMessage, names ...string) error {
	for _, name := range names {
		raw, ok := fields[name]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("scene.round12_required_number: %s", name)
		}
		var n float64
		if json.Unmarshal(raw, &n) != nil || !intakeFinite(n) {
			return fmt.Errorf("scene.round12_required_number: %s", name)
		}
	}
	return nil
}
func (r *renderer) planIntakeRound12Scene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "funnel" && tag.Type != "pyramid" && tag.Type != "bracket" {
		return nil, false, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, true, err
	}
	if tag.Type == "pyramid" && !round12PyramidInput(fields) {
		return nil, false, nil
	}
	if err := round12RequiredNumbers(fields, "x", "y"); err != nil {
		return nil, true, err
	}
	if tag.Type == "bracket" {
		var n round12Bracket
		if err := sceneDecode(raw, &n); err != nil {
			return nil, true, err
		}
		if n.Orient == "" {
			n.Orient = "down"
		}
		dim := "w"
		if n.Orient == "left" || n.Orient == "right" {
			dim = "h"
		}
		if err := round12RequiredNumbers(fields, dim); err != nil {
			return nil, true, err
		}
		p, err := r.round12Bracket(id, n, ctx)
		return p, true, err
	}
	if err := round12RequiredNumbers(fields, "w", "h"); err != nil {
		return nil, true, err
	}
	var n round12Bands
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	p, err := r.round12Bands(id, n, ctx)
	return p, true, err
}

type round12TextPart struct {
	text, token, ink, family string
	weight                   int
}

func (r *renderer) round12Stack(p *scenePlan, id string, parts []round12TextPart, b Rect, surface, align string, ctx SceneContext) error {
	var records []TextRecord
	y := 0.
	for i, part := range parts {
		if part.text == "" {
			continue
		}
		st, err := r.sceneStyle(part.token)
		if err != nil {
			return err
		}
		if part.family != "" {
			st.Family = part.family
		}
		if part.weight != 0 {
			st.Weight = part.weight
		}
		if len(records) > 0 {
			y += 2
		}
		partID := fmt.Sprintf("%s.part-%d", id, i+1)
		var record TextRecord
		if strings.Contains(part.text, "[^") {
			child := &scenePlan{}
			if err := r.primitiveRichText(child, partID, part.text, st, Rect{b.X, b.Y + y, b.W, 0}, surface, part.ink, align, "", "", ctx); err != nil {
				return err
			}
			if len(child.Items) != 1 || child.Items[0].Text == nil {
				return fmt.Errorf("scene.round12_inline_copy")
			}
			record = *child.Items[0].Text
			p.Warnings = append(p.Warnings, child.Warnings...)
		} else {
			l, err := r.measureText(part.text, st, b.W)
			if err != nil {
				return err
			}
			color, err := r.sceneColor(surface, part.ink)
			if err != nil {
				return err
			}
			h := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
			record = TextRecord{ID: partID, Rect: Rect{b.X, b.Y + y, b.W, h}, Color: color, Align: align, Layout: l}
		}
		h := math.Max(record.Layout.AllocationHeight, record.Layout.OccupiedTop+record.Layout.EstimatedOccupiedHeight)
		record.Rect.H = h
		records = append(records, record)
		y += h
	}
	if y > b.H+.02 {
		return fmt.Errorf("scene.round12_stack_overflow: %s needs%.3f capacity%.3f", id, y, b.H)
	}
	for _, record := range records {
		record.Rect.Y += (b.H - y) / 2
		p.Items = append(p.Items, sceneItem{Text: &record})
		p.Bounds = diagramUnion(p.Bounds, record.Rect)
	}
	return nil
}

func (r *renderer) round12Bands(id string, n round12Bands, ctx SceneContext) (*scenePlan, error) {
	bad := func(reason string) (*scenePlan, error) {
		return nil, fmt.Errorf("scene.round12_bands_%s: %s", reason, id)
	}
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH) || n.W <= 0 || n.H <= 0 || n.W > 1920 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 {
		return bad("geometry")
	}
	if len(n.Stages) > 0 && len(n.Levels) > 0 {
		return bad("stages_levels_union")
	}
	bands, path := n.Stages, "/stages"
	if len(bands) == 0 {
		bands, path = n.Levels, "/levels"
	}
	if len(bands) < 1 || len(bands) > 12 {
		return bad("count")
	}
	sw, gap, neck := n.W, 6., .3
	if n.ShapeW != nil {
		sw = *n.ShapeW
	}
	if n.Gap != nil {
		gap = *n.Gap
	}
	if n.Neck != nil {
		neck = *n.Neck
	}
	if !intakeFinite(sw, gap, neck) || sw <= 0 || sw > n.W || gap < 0 || gap > n.H || neck <= 0 || neck > 1 {
		return bad("shape_options")
	}
	if n.Ramp != "" && n.Ramp != "light" {
		return bad("ramp")
	}
	bandH := (n.H - gap*float64(len(bands)-1)) / float64(len(bands))
	if bandH <= 0 {
		return bad("band_height")
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	if _, err := r.sceneColor(surface, "primary"); err != nil {
		return nil, err
	}
	keys, err := primitiveArrayKeys(ctx, path, len(bands))
	if err != nil {
		return nil, err
	}
	valueStyle := n.ValueStyle
	if valueStyle == "" {
		valueStyle = "subhead"
	}
	if _, err = r.sceneStyle(valueStyle); err != nil {
		return nil, err
	}
	ramp := []string{"E8EEF8", "CED7E6", "B9C9F0", "7C9BFF", "0047FF", "070154"}
	if n.Ramp == "light" {
		ramp = []string{"E8EEF8", "DCE4F2", "CED7E6", "C2CEE6", "B9C9F0", "AFC0EC"}
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, n.H}}
	textPlan := &scenePlan{}
	budget := 0
	for i, band := range bands {
		value := ""
		if len(band.Value) > 0 && string(band.Value) != "null" {
			value, err = maturityNumber(band.Value, "")
			if err != nil {
				return nil, err
			}
		}
		budget += utf8.RuneCountInString(band.Label) + utf8.RuneCountInString(band.Text) + utf8.RuneCountInString(value)
		if utf8.RuneCountInString(band.Label) > 2048 || utf8.RuneCountInString(band.Text) > 8192 || budget > 65536 {
			return bad("copy_limit")
		}
		y := n.Y + float64(i)*(bandH+gap)
		fraction := float64(i) / float64(len(bands))
		wt, wb := sw*fraction, sw*float64(i+1)/float64(len(bands))
		ri := int(math.Floor(float64(len(bands)-1-i)/math.Max(1, float64(len(bands)-1))*5 + .5))
		if n.Type == "funnel" {
			wt = sw * (1 - (1-neck)*fraction)
			wb = sw * (1 - (1-neck)*float64(i+1)/float64(len(bands)))
			ri = int(math.Floor(float64(i)/math.Max(1, float64(len(bands)-1))*5 + .5))
		}
		if n.Dark != nil && !*n.Dark {
			ri = min(ri, 3)
		}
		fill := ramp[ri]
		if band.Surface != "" {
			fill, err = r.sceneColor(band.Surface, "bg")
			if err != nil {
				return nil, err
			}
		} else if band.Active {
			fill = "F900D3"
		}
		ink := "#070154"
		if band.Surface == "inverse" || band.Surface == "" && !band.Active && ri >= 4 {
			ink = "#FFFFFF"
		}
		cx := n.X + sw/2
		// SVG coordinates are serialized at one decimal place in the pinned source.
		round := func(v float64) float64 { return math.Round(v*10) / 10 }
		x0, x1, x2, x3 := round(cx-wt/2), round(cx+wt/2), round(cx+wb/2), round(cx-wb/2)
		y0, y1 := round(y), round(y+bandH)
		box := Rect{math.Min(x0, x3), y0, math.Max(x1, x2) - math.Min(x0, x3), y1 - y0}
		if box.W <= 0 || box.H <= 0 {
			return bad("rounded_band_geometry")
		}
		pts := [][2]float64{{x0 - box.X, 0}, {x1 - box.X, 0}, {x2 - box.X, box.H}, {x3 - box.X, box.H}}
		bid := id + "." + strings.TrimPrefix(path, "/") + "." + keys[i]
		if err = r.diagramShape(p, bid+".band", box, pptx.ShapeTypeCustGeom, fill, "", 0, "solid", pts); err != nil {
			return nil, err
		}
		narrow := wt
		if narrow == 0 {
			narrow = wb * .6
		}
		inW := math.Max(40, math.Min(narrow, wb)-24)
		token := "body"
		if bandH < 40 {
			token = "small"
		}
		parts := []round12TextPart{{value, valueStyle, ink, "IBM Plex Mono", 600}}
		if !(n.LabelSide && sw < n.W) {
			parts = append(parts, round12TextPart{band.Label, token, ink, "", 600})
		}
		if err = r.round12Stack(textPlan, bid+".inside", parts, Rect{cx - inW/2, y, inW, bandH}, surface, "center", ctx); err != nil {
			return nil, err
		}
		if sw < n.W && (band.Text != "" || n.LabelSide && band.Label != "") {
			tx := n.X + sw + 18
			sideW := n.X + n.W - tx
			if sideW <= 0 {
				return bad("side_width")
			}
			if err = r.diagramLine(p, bid+".leader", [2]float64{round(cx + (wt+wb)/4 + 6), round(y + bandH/2)}, [2]float64{tx - 6, round(y + bandH/2)}, "97A4BA", 1, "dot"); err != nil {
				return nil, err
			}
			token = "body"
			if bandH < 54 {
				token = "small"
			}
			parts = nil
			if n.LabelSide {
				parts = append(parts, round12TextPart{band.Label, "body", "display", "", 600})
			}
			parts = append(parts, round12TextPart{band.Text, token, "primary", "", 0})
			if err = r.round12Stack(textPlan, bid+".side", parts, Rect{tx, y, sideW, bandH}, surface, "left", ctx); err != nil {
				return nil, err
			}
		}
	}
	p.Items = append(p.Items, textPlan.Items...)
	if len(textPlan.Items) > 0 {
		p.Bounds = diagramUnion(p.Bounds, textPlan.Bounds)
	}
	p.Warnings = append(p.Warnings, IntakeRound12Contract+" renderer_sha256="+IntakeRound12RendererSHA256+"; editable source adapter, native specimen review pending")
	sceneDataGroup(p, id, "diagram.round12."+n.Type, 0, p.Bounds)
	return p, nil
}

func (r *renderer) round12Bracket(id string, n round12Bracket, ctx SceneContext) (*scenePlan, error) {
	if n.Orient != "down" && n.Orient != "up" && n.Orient != "left" && n.Orient != "right" {
		return nil, fmt.Errorf("scene.round12_bracket_orient")
	}
	depth := 9.
	if n.Depth != nil {
		depth = *n.Depth
	}
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH, depth) || n.W < 0 || n.H < 0 || n.W > 1920 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || depth <= 0 || depth > 180 || n.CanvasH < 0 || n.CanvasH > 2160 || utf8.RuneCountInString(n.Label) > 2048 {
		return nil, fmt.Errorf("scene.round12_bracket_geometry_or_copy")
	}
	horizontal := n.Orient == "down" || n.Orient == "up"
	if horizontal && n.W <= 0 || !horizontal && n.H <= 0 {
		return nil, fmt.Errorf("scene.round12_bracket_span")
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	role := n.Ink
	if role == "" {
		role = "emphasis"
	}
	color, err := r.sceneColor(surface, role)
	if err != nil {
		return nil, err
	}
	p := &scenePlan{ID: id}
	var a, b, c, d, e, f [2]float64
	label := Rect{n.X, n.Y - depth - 18, n.W, 16}
	align := "center"
	if horizontal {
		sign := 1.
		if n.Orient == "up" {
			sign = -1
			label.Y = n.Y + depth + 2
		}
		a = [2]float64{n.X, n.Y + sign*depth}
		b = [2]float64{n.X, n.Y}
		c = [2]float64{n.X + n.W, n.Y}
		d = [2]float64{n.X + n.W, n.Y + sign*depth}
		e = [2]float64{n.X + n.W/2, n.Y}
		f = [2]float64{n.X + n.W/2, n.Y - sign*depth}
	} else {
		sign := 1.
		align = "left"
		label = Rect{n.X + depth + 6, n.Y + n.H/2 - 8, 180, 16}
		if n.Orient == "left" {
			sign = -1
			align = "right"
			label.X = n.X - depth - 6 - 180
		}
		a = [2]float64{n.X - sign*depth, n.Y}
		b = [2]float64{n.X, n.Y}
		c = [2]float64{n.X, n.Y + n.H}
		d = [2]float64{n.X - sign*depth, n.Y + n.H}
		e = [2]float64{n.X, n.Y + n.H/2}
		f = [2]float64{n.X + sign*depth, n.Y + n.H/2}
	}
	for i, pair := range [][2][2]float64{{a, b}, {b, c}, {c, d}, {e, f}} {
		if err = r.diagramLine(p, fmt.Sprintf("%s.stroke-%d", id, i+1), pair[0], pair[1], color, 1.5, "solid"); err != nil {
			return nil, err
		}
	}
	// Record actual stroke extent, including the external centre tick.
	p.Bounds = Rect{p.Bounds.X - .75, p.Bounds.Y - .75, p.Bounds.W + 1.5, p.Bounds.H + 1.5}
	if n.Label != "" {
		if err = r.round12Stack(p, id+".label", []round12TextPart{{n.Label, "label", "emphasis", "", 600}}, label, surface, align, ctx); err != nil {
			return nil, err
		}
	}
	p.Warnings = append(p.Warnings, IntakeRound12Contract+" renderer_sha256="+IntakeRound12RendererSHA256+"; editable source adapter, native specimen review pending")
	sceneDataGroup(p, id, "diagram.round12.bracket", 0, p.Bounds)
	return p, nil
}
