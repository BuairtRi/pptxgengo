package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/pptx"
)

const IntakeVennContract = "pptxgengo.wmds-source-intake-venn.v1"

type intakeVennSet struct {
	Label   string            `json:"label"`
	Text    string            `json:"text,omitempty"`
	Bullets []json.RawMessage `json:"bullets,omitempty"`
	Fill    string            `json:"fill,omitempty"`
	W       *float64          `json:"w,omitempty"`
	DX      float64           `json:"dx,omitempty"`
	DY      float64           `json:"dy,omitempty"`
}
type intakeVennRegion struct {
	In      []int             `json:"in"`
	Label   string            `json:"label,omitempty"`
	Text    string            `json:"text,omitempty"`
	Bullets []json.RawMessage `json:"bullets,omitempty"`
	W       *float64          `json:"w,omitempty"`
	DX      float64           `json:"dx,omitempty"`
	DY      float64           `json:"dy,omitempty"`
}
type intakeVennPoint struct {
	X     *float64        `json:"x"`
	Y     *float64        `json:"y"`
	Label string          `json:"label,omitempty"`
	N     json.RawMessage `json:"n,omitempty"`
	Side  string          `json:"side,omitempty"`
}
type intakeVennSource struct {
	Type       string             `json:"type"`
	ID         string             `json:"id,omitempty"`
	X          float64            `json:"x"`
	Y          float64            `json:"y"`
	W          float64            `json:"w"`
	H          float64            `json:"h"`
	Sets       []intakeVennSet    `json:"sets"`
	Regions    []intakeVennRegion `json:"regions,omitempty"`
	Points     []intakeVennPoint  `json:"points,omitempty"`
	TextW      *float64           `json:"textW,omitempty"`
	Opacity    *float64           `json:"opacity,omitempty"`
	LabelStyle string             `json:"labelStyle,omitempty"`
	Numbered   *bool              `json:"numbered,omitempty"`
	CanvasH    float64            `json:"_h,omitempty"`
}

// Geometry mirrors the source renderer, including the outward pairwise lens
// anchors. Text uses the pinned Go measurer, not a rasterized browser image.
func intakeVennGeometry(n intakeVennSource) (float64, [][2]float64) {
	x, y := n.X+n.W/2, n.Y+n.H/2
	switch len(n.Sets) {
	case 2:
		r := math.Min(n.H/2, n.W/3.2)
		return r, [][2]float64{{x - r*.6, y}, {x + r*.6, y}}
	case 3:
		r := math.Min(n.H/3.1, n.W/3.3)
		d := r * .66
		return r, [][2]float64{{x - d, y - d*.55}, {x + d, y - d*.55}, {x, y + d*.85}}
	default:
		r := math.Min(n.H, n.W) / 3.3
		d := r * .62
		return r, [][2]float64{{x - d, y - d}, {x + d, y - d}, {x - d, y + d}, {x + d, y + d}}
	}
}
func intakeVennAnchor(centers [][2]float64, indices []int, center [2]float64, radius float64, own bool) [2]float64 {
	p := [2]float64{}
	for _, i := range indices {
		p[0] += centers[i][0]
		p[1] += centers[i][1]
	}
	p[0] /= float64(len(indices))
	p[1] /= float64(len(indices))
	dx, dy := p[0]-center[0], p[1]-center[1]
	length := math.Hypot(dx, dy)
	push := 0.
	if own {
		push = radius * .42
	} else if len(indices) < len(centers) && length > 1 {
		if len(centers) == 3 {
			push = radius * .5
		} else if len(centers) == 4 {
			push = radius * .38
		}
	}
	if length > 0 {
		p[0] += dx / length * push
		p[1] += dy / length * push
	}
	return p
}
func intakeVennBound(p *scenePlan, b Rect) { p.Bounds = diagramUnion(p.Bounds, b) }

func (r *renderer) planIntakeVennScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, false, err
	}
	if head.Type != "venn" {
		return nil, false, nil
	}
	var n intakeVennSource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	bad := func(reason string) (*scenePlan, bool, error) {
		return nil, true, fmt.Errorf("scene.venn_%s: %s", reason, id)
	}
	// encoding/json treats null scalar numbers as zero. Geometry requires
	// actual numbers; a missing origin must not silently become a valid 0pt.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, true, err
	}
	for _, key := range []string{"x", "y", "w", "h"} {
		v, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return bad("invalid_geometry")
		}
	}
	if v, ok := fields["_h"]; ok && bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
		return bad("invalid_geometry")
	}
	if !intakeFinite(n.X, n.Y, n.W, n.H, n.CanvasH) || n.W <= 0 || n.H <= 0 || n.W > 1920 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 || n.CanvasH < 0 || n.CanvasH > 2160 {
		return bad("invalid_geometry")
	}
	if len(n.Sets) < 2 || len(n.Sets) > 4 || len(n.Regions) > 11 || len(n.Points) > 64 {
		return bad("cardinality")
	}
	if n.LabelStyle != "" && n.LabelStyle != "mono" && n.LabelStyle != "subhead" {
		return bad("label_style")
	}
	round12 := isExpandedLibrary(r.source.Revision)
	radius, centers := intakeVennGeometry(n)
	originalRadius, originalCenters := radius, append([][2]float64(nil), centers...)
	center := [2]float64{n.X + n.W/2, n.Y + n.H/2}
	if round12 && len(n.Sets) > 2 {
		radius, centers, center = intakeVennRoomierGeometry(n, ctx, radius, centers)
	}
	textW := radius * .95
	if n.TextW != nil {
		if !intakeFinite(*n.TextW) || *n.TextW <= 0 || *n.TextW > n.W {
			return bad("text_width")
		}
		textW = *n.TextW
	}
	opacity := .7
	if n.Opacity != nil {
		opacity = *n.Opacity
		if !intakeFinite(opacity) || opacity < 0 || opacity > 1 {
			return bad("opacity")
		}
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	// Validate the surface even when a diagram consists only of explicit fills.
	if _, err := r.sceneColor(surface, "primary"); err != nil {
		return nil, true, err
	}
	setKeys, err := primitiveArrayKeys(ctx, "/sets", len(n.Sets))
	if err != nil {
		return nil, true, err
	}
	keys := func(path string, count int) ([]string, error) {
		if count > 0 {
			return primitiveArrayKeys(ctx, path, count)
		}
		if _, ok := ctx.Keys[ctx.Path+path]; ok {
			return primitiveArrayKeys(ctx, path, 0)
		}
		return nil, nil
	}
	regionKeys, err := keys("/regions", len(n.Regions))
	if err != nil {
		return nil, true, err
	}
	pointKeys, err := keys("/points", len(n.Points))
	if err != nil {
		return nil, true, err
	}
	bulletCount, copyRunes := 0, 0
	validCopy := func(label, text string, bullets []json.RawMessage, width *float64) bool {
		labelRunes, textRunes := utf8.RuneCountInString(label), utf8.RuneCountInString(text)
		copyRunes += labelRunes + textRunes
		return copyRunes <= 65536 && labelRunes <= 2048 && textRunes <= 8192 && len(bullets) <= 32 && (width == nil || intakeFinite(*width) && *width > 0 && *width <= n.W)
	}
	for _, s := range n.Sets {
		if !intakeFinite(s.DX, s.DY) || math.Abs(s.DX) > n.W || math.Abs(s.DY) > n.H {
			return bad("label_nudge")
		}
		if strings.TrimSpace(s.Label) == "" || !validCopy(s.Label, s.Text, s.Bullets, s.W) {
			return bad("set_content_or_width")
		}
		if err := intakeVennBullets(s.Bullets, 0, &bulletCount, &copyRunes); err != nil {
			return nil, true, err
		}
	}
	seen := map[string]bool{}
	for _, rg := range n.Regions {
		if !intakeFinite(rg.DX, rg.DY) || math.Abs(rg.DX) > n.W || math.Abs(rg.DY) > n.H {
			return bad("label_nudge")
		}
		if err := intakeVennBullets(rg.Bullets, 0, &bulletCount, &copyRunes); err != nil {
			return nil, true, err
		}
		if len(rg.In) < 2 || len(rg.In) > len(n.Sets) || !validCopy(rg.Label, rg.Text, rg.Bullets, rg.W) || (rg.Label == "" && rg.Text == "" && len(rg.Bullets) == 0) {
			return bad("region_content_or_indices")
		}
		indices := append([]int(nil), rg.In...)
		sort.Ints(indices)
		for i, j := range indices {
			if j < 0 || j >= len(n.Sets) || i > 0 && indices[i-1] == j {
				return bad("region_indices")
			}
		}
		key := fmt.Sprint(indices)
		if seen[key] {
			return bad("duplicate_region")
		}
		seen[key] = true
	}
	for _, pt := range n.Points {
		copyRunes += utf8.RuneCountInString(pt.Label)
		if copyRunes > 65536 {
			return bad("copy_limit")
		}
		if pt.X == nil || pt.Y == nil || !intakeFinite(*pt.X, *pt.Y) || *pt.X < 0 || *pt.X > 1 || *pt.Y < 0 || *pt.Y > 1 || pt.Side != "" && pt.Side != "left" && pt.Side != "right" || utf8.RuneCountInString(pt.Label) > 2048 {
			return bad("point_position_or_side")
		}
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, n.H}}
	pointPositions := make([][2]float64, len(n.Points))
	var pointObstacles []Rect
	for i, pt := range n.Points {
		x, y := n.X+*pt.X*n.W, n.Y+*pt.Y*n.H
		if round12 && len(n.Sets) > 2 {
			x, y = intakeVennPreservePointRegion(x, y, originalCenters, originalRadius, centers, radius)
			if intakeVennPointMask(x, y, centers, radius) != intakeVennPointMask(n.X+*pt.X*n.W, n.Y+*pt.Y*n.H, originalCenters, originalRadius) {
				return bad("point_region_fit")
			}
			pointObstacles = append(pointObstacles, Rect{x - 10.5, y - 10.5, 21, 21})
			if pt.Label != "" {
				st, e := r.sceneStyle("label")
				if e != nil {
					return nil, true, e
				}
				st.Size, st.Leading, st.Weight, st.TrackingPt, st.Case = 9.5, 9.5, 600, 9.5*.06, "upper"
				if densityRoleCorrections(r.source) {
					st, e = r.intakeVennDensityBadgeStyle(st)
					if e != nil {
						return nil, true, e
					}
				}
				layout, e := r.measureText(pt.Label, st, 8191)
				if e != nil {
					return nil, true, e
				}
				if len(layout.Lines) != 1 {
					return bad("point_badge_multiline")
				}
				w := sequenceInlineWidth(layout.Lines[0].Advance, layout.Style) + 10
				bx := x + 12
				if pt.Side == "left" {
					bx = x - 12 - w
				}
				pointObstacles = append(pointObstacles, Rect{bx, y - 15.5/2, w, 15.5})
			}
		}
		pointPositions[i] = [2]float64{x, y}
	}
	fills := []string{"E8EEF8", "CED7E6", "B9C9F0", "F3D6EE"}
	for i, s := range n.Sets {
		fill := fills[i]
		if s.Fill != "" {
			fill, err = r.sceneColor(surface, s.Fill)
			if err != nil {
				return nil, true, err
			}
		}
		// Browser SVG serializes circle geometry to one decimal place.
		rr := math.Round(radius*10) / 10
		cx, cy := math.Round(centers[i][0]*10)/10, math.Round(centers[i][1]*10)/10
		b := Rect{cx - rr, cy - rr, rr * 2, rr * 2}
		if err = r.diagramShape(p, id+".sets."+setKeys[i]+".circle", b, pptx.ShapeTypeEllipse, fill, "070154", 1, "solid", nil); err != nil {
			return nil, true, err
		}
		p.Items[len(p.Items)-1].Shape.Props.Fill.Transparency = (1 - opacity) * 100
		intakeVennBound(p, Rect{b.X - .5, b.Y - .5, b.W + 1, b.H + 1})
	}
	for i, s := range n.Sets {
		w := textW
		if s.W != nil {
			w = *s.W
		}
		anchor := intakeVennAnchor(centers, []int{i}, center, radius, true)
		anchor[0], anchor[1] = anchor[0]+s.DX, anchor[1]+s.DY
		ownLabel, labelSize := n.LabelStyle != "mono", 0.
		if round12 {
			ownLabel, labelSize = n.LabelStyle == "subhead", 13
		}
		var fit *intakeVennTextFit
		if round12 && len(centers) > 2 {
			fit = &intakeVennTextFit{centers: centers, radius: radius, selected: []int{i}, avoid: pointObstacles}
		}
		if err = r.intakeVennText(p, id+".sets."+setKeys[i], anchor, w, s.Label, s.Text, s.Bullets, ownLabel, labelSize, surface, ctx, "/sets/"+strconv.Itoa(i)+"/bullets", fit); err != nil {
			return nil, true, err
		}
	}
	for i, rg := range n.Regions {
		factor := .75
		if len(n.Sets) == 4 && len(rg.In) == 4 {
			factor = .9
		}
		w := math.Min(textW, radius*factor)
		if rg.W != nil {
			w = *rg.W
		}
		anchor := intakeVennAnchor(centers, rg.In, center, radius, false)
		labelSize := 0.
		if round12 {
			anchor = intakeVennRound12Anchor(centers, rg.In, center, radius)
			labelSize = 11
			if len(rg.In) == len(centers) && len(centers) > 2 {
				labelSize = 13
			}
		}
		anchor[0], anchor[1] = anchor[0]+rg.DX, anchor[1]+rg.DY
		var fit *intakeVennTextFit
		if round12 && len(centers) > 2 {
			fit = &intakeVennTextFit{centers: centers, radius: radius, selected: rg.In, avoid: pointObstacles}
		}
		if err = r.intakeVennText(p, id+".regions."+regionKeys[i], anchor, w, rg.Label, rg.Text, rg.Bullets, false, labelSize, surface, ctx, "/regions/"+strconv.Itoa(i)+"/bullets", fit); err != nil {
			return nil, true, err
		}
	}
	for i, pt := range n.Points {
		part := id + ".points." + pointKeys[i]
		x, y := pointPositions[i][0], pointPositions[i][1]
		if err = r.diagramShape(p, part+".halo", Rect{x - 10.5, y - 10.5, 21, 21}, pptx.ShapeTypeEllipse, "FFFFFF", "", 0, "solid", nil); err != nil {
			return nil, true, err
		}
		if err = r.diagramShape(p, part+".marker", Rect{x - 9, y - 9, 18, 18}, pptx.ShapeTypeEllipse, "070154", "", 0, "solid", nil); err != nil {
			return nil, true, err
		}
		intakeVennBound(p, Rect{x - 10.5, y - 10.5, 21, 21})
		num := ""
		if n.Numbered == nil || *n.Numbered {
			num = strconv.Itoa(i + 1)
		}
		if len(pt.N) > 0 {
			num, err = intakeVennNumber(pt.N)
			if err != nil {
				return nil, true, fmt.Errorf("scene.venn_point_number: %s: %w", part, err)
			}
		}
		st, e := r.sceneStyle("label")
		if e != nil {
			return nil, true, e
		}
		st.Size, st.Leading, st.Weight, st.TrackingPt, st.Case = 9, 9, 600, 0, ""
		if round12 {
			st.Size, st.Leading = 10, 10
		}
		if num != "" {
			if err = r.intakeVennSingleText(p, part+".number", num, st, Rect{x - 9, y - 9, 18, 18}, "inverse", "#FFFFFF", "center", true); err != nil {
				return nil, true, err
			}
		}
		if pt.Label != "" {
			st.Size, st.Leading, st.TrackingPt, st.Case = 8, 8, 8*.06, "upper"
			if round12 {
				st.Size, st.Leading, st.TrackingPt = 9.5, 9.5, 9.5*.06
			}
			if densityRoleCorrections(r.source) {
				st, e = r.intakeVennDensityBadgeStyle(st)
				if e != nil {
					return nil, true, e
				}
			}
			layout, e := r.measureText(pt.Label, st, 8191)
			if e != nil {
				return nil, true, e
			}
			if len(layout.Lines) != 1 {
				return bad("point_badge_multiline")
			}
			w := layout.Lines[0].Advance + 10
			if round12 {
				// Native textboxes need trailing space beyond the shaped
				// advance. The source tag's 5pt padding on each side belongs
				// outside that textbox and cannot provide this reserve.
				w = sequenceInlineWidth(layout.Lines[0].Advance, layout.Style) + 10
			}
			limit := 120.
			if round12 {
				limit = 140
			}
			if w > limit+.02 {
				return bad("point_badge_width")
			}
			bx := x + 12
			if pt.Side == "left" {
				bx = x - 12 - w
			}
			// The 16pt source row centers a 14pt tag (8pt line + 3pt
			// vertical padding on each side). Its .75pt shadow is inset.
			badgeH := 14.
			if round12 {
				badgeH = 15.5
			}
			if err = r.diagramShape(p, part+".badge", Rect{bx + .375, y - badgeH/2 + .375, w - .75, badgeH - .75}, pptx.ShapeTypeRect, "FFFFFF", "070154", .75, "solid", nil); err != nil {
				return nil, true, err
			}
			intakeVennBound(p, Rect{bx, y - badgeH/2, w, badgeH})
			if err = r.intakeVennSingleText(p, part+".label", pt.Label, st, Rect{bx + 5, y - badgeH/2, w - 10, badgeH}, "light", "#070154", "center", true); err != nil {
				return nil, true, err
			}
		}
	}
	// Actual text and outlined shape bounds stay visible to frame/allocation guards;
	// edge-positioned points are never silently clamped into the diagram box.
	primitiveFinish(p, "venn")
	p.Groups[len(p.Groups)-1].Contract = IntakeVennContract
	return p, true, nil
}

func (r *renderer) intakeVennDensityBadgeStyle(st Style) (Style, error) {
	label, err := r.sceneStyle("label")
	if err != nil {
		return st, err
	}
	st.Size, st.Leading = label.Size*9.5/9, label.Size*9.5/9
	st.Tracking, st.TrackingPt = ".06em", math.Round(st.Size*.06*100)/100
	st.Weight, st.Case = 600, "upper"
	return st, nil
}

func intakeVennNumber(raw json.RawMessage) (string, error) {
	var text string
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &text); err != nil {
			return "", err
		}
		if strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > 16 {
			return "", fmt.Errorf("empty or long marker number")
		}
		return text, nil
	}
	if bytes.Equal(raw, []byte("null")) {
		return "", fmt.Errorf("null marker number")
	}
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	if !intakeFinite(value) || math.Abs(value) > 1e6 {
		return "", fmt.Errorf("invalid marker number")
	}
	return strconv.FormatFloat(value, 'f', -1, 64), nil
}

func intakeVennBullets(items []json.RawMessage, depth int, total, copyRunes *int) error {
	if depth > 8 || len(items) > 32 {
		return fmt.Errorf("scene.venn_bullet_limit")
	}
	for _, raw := range items {
		*total++
		if *total > 128 {
			return fmt.Errorf("scene.venn_bullet_limit")
		}
		var text string
		if len(raw) > 0 && raw[0] == '"' {
			if err := json.Unmarshal(raw, &text); err != nil {
				return err
			}
		} else {
			var item struct {
				Lead string            `json:"lead,omitempty"`
				Text string            `json:"text"`
				Sub  []json.RawMessage `json:"sub,omitempty"`
			}
			if err := sceneDecode(raw, &item); err != nil {
				return err
			}
			text = item.Text
			if utf8.RuneCountInString(item.Lead) > 2048 {
				return fmt.Errorf("scene.venn_bullet_content")
			}
			*copyRunes += utf8.RuneCountInString(item.Lead)
			if item.Lead != "" && len(item.Sub) > 0 {
				// The browser's lead branch suppresses sub-items. Reject this
				// ambiguous input instead of adding invisible source content.
				return fmt.Errorf("scene.venn_bullet_lead_sub_union")
			}
			if err := intakeVennBullets(item.Sub, depth+1, total, copyRunes); err != nil {
				return err
			}
		}
		if strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > 8192 {
			return fmt.Errorf("scene.venn_bullet_content")
		}
		*copyRunes += utf8.RuneCountInString(text)
		if *copyRunes > 65536 {
			return fmt.Errorf("scene.venn_copy_limit")
		}
	}
	return nil
}
func (r *renderer) intakeVennSingleText(p *scenePlan, id, text string, st Style, b Rect, surface, ink, align string, middle bool) error {
	l, err := r.measureText(text, st, b.W)
	if err != nil {
		return err
	}
	if len(l.Lines) != 1 {
		return fmt.Errorf("scene.venn_unbreakable_text_overflow: %s", id)
	}
	need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
	if need > b.H+.02 {
		return fmt.Errorf("scene.venn_point_text_overflow: %s", id)
	}
	if middle {
		if !isExpandedLibrary(r.source.Revision) {
			b.Y += (b.H - need) / 2
			b.H = need
		}
	}
	color, err := r.sceneColor(surface, ink)
	if err != nil {
		return err
	}
	// Marker numbers and badge spans use browser textContent. Keep brackets
	// literal here rather than routing them through the rich-markup parser.
	tr := &TextRecord{ID: id, Rect: b, Color: color, Align: align, Layout: l}
	if middle && isExpandedLibrary(r.source.Revision) {
		tr.VerticalAlign = "middle"
	}
	p.Items = append(p.Items, sceneItem{Text: tr})
	intakeVennBound(p, b)
	return nil
}
func (r *renderer) intakeVennText(p *scenePlan, id string, anchor [2]float64, w float64, label, text string, bullets []json.RawMessage, ownLabel bool, labelSize float64, surface string, ctx SceneContext, path string, fit *intakeVennTextFit) error {
	child := &scenePlan{ID: id}
	originalX := anchor[0]
	y := 0.
	add := func(part, copy, token, ink string, weight int) error {
		if copy == "" {
			return nil
		}
		if y > 0 {
			y += 3
		}
		st, err := r.sceneStyle(token)
		if err != nil {
			return err
		}
		if weight > 0 {
			st.Weight = weight
		}
		if part == ".label" && !ownLabel && labelSize > 0 {
			st.Size, st.Leading, st.Weight, st.TrackingPt, st.Case = labelSize, labelSize*1.25, 600, labelSize*.06, "upper"
		}
		if err := r.diagramStyledText(child, id+part, copy, st, Rect{anchor[0] - w/2, y, w, 0}, surface, ink, "center", false); err != nil {
			return err
		}
		y = child.Items[len(child.Items)-1].Text.Rect.Y + child.Items[len(child.Items)-1].Text.Rect.H
		return nil
	}
	token := "label"
	if ownLabel {
		token = "subhead"
	}
	if err := add(".label", label, token, "display", 600); err != nil {
		return err
	}
	if err := add(".text", text, "small", "primary", 0); err != nil {
		return err
	}
	if len(bullets) > 0 {
		if y > 0 {
			y += 3
		}
		if err := r.primitiveBullets(child, id+".bullets", bullets, Rect{anchor[0] - w/2, y, w, 0}, surface, "small", ctx, path, false, 0); err != nil {
			return err
		}
		for _, it := range child.Items {
			if it.Text != nil {
				y = math.Max(y, it.Text.Rect.Y+it.Text.Rect.H)
			}
			if it.Shape != nil {
				y = math.Max(y, it.Shape.Record.Rect.Y+it.Shape.Record.Rect.H)
			}
		}
	}
	// V4 positions the measured complete stack in its actual Venn region.
	// Width is reduced only when the lens cannot contain the measured stack;
	// the same pinned font then provides natural line breaks.
	if fit != nil {
		position, ok := intakeVennFitTextRect(anchor, w, y, *fit)
		if !ok {
			if w < 24 {
				return fmt.Errorf("scene.venn_region_text_overflow: %s", id)
			}
			return r.intakeVennText(p, id, anchor, w*.85, label, text, bullets, ownLabel, labelSize, surface, ctx, path, fit)
		}
		anchor = position
	}
	// Preserve the older source stack transform for historical revisions.
	dy := anchor[1] - 9 - y*.4
	if isExpandedLibrary(r.source.Revision) {
		dy = anchor[1] - y/2
	}
	for i := range child.Items {
		it := &child.Items[i]
		if it.Text != nil {
			it.Text.Rect.Y += dy
			it.Text.Rect.X += anchor[0] - originalX
			if isExpandedLibrary(r.source.Revision) && strings.HasSuffix(it.Text.ID, ".label") {
				it.Text.VerticalAlign = "middle"
			}
			intakeVennBound(p, it.Text.Rect)
		}
		if it.Shape != nil {
			it.Shape.Record.Rect.Y += dy
			it.Shape.Record.Rect.X += anchor[0] - originalX
			it.Shape.Props.PositionProps = pos(it.Shape.Record.Rect)
			intakeVennBound(p, it.Shape.Record.Rect)
		}
	}
	p.Items = append(p.Items, child.Items...)
	p.Groups = append(p.Groups, child.Groups...)
	p.Warnings = append(p.Warnings, child.Warnings...)
	return nil
}
