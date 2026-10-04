package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

type sceneAnnotationSource struct {
	Type      string  `json:"type"`
	Target    string  `json:"target"`
	Placement string  `json:"placement,omitempty"`
	Arrow     string  `json:"arrow,omitempty"`
	Size      float64 `json:"size,omitempty"`
	Ink       string  `json:"ink,omitempty"`
	Callout   struct {
		W       float64 `json:"w,omitempty"`
		Label   string  `json:"label,omitempty"`
		Text    string  `json:"text"`
		Surface string  `json:"surface,omitempty"`
	} `json:"callout"`
}

type annotationTarget struct {
	Bounds Rect
	Top    float64 // Includes any label drawn above a frame's border.
}

// Targets are measured before paint, so source order does not change resolution.
// Only explicit source IDs participate; ordinals are never guessed as targets.
func (r *renderer) registerSceneTargets(slide SlideSpec, f ResolvedFrame) error {
	r.sceneTargets = map[string]annotationTarget{}
	for _, node := range slide.Nodes {
		if node.Scene == nil {
			continue
		}
		var tag struct{ ID, Type string }
		if err := json.Unmarshal(node.Scene.Node, &tag); err != nil {
			return err
		}
		if tag.ID == "" {
			continue
		}
		if !validPartKey(tag.ID) || tag.Type == "annotation" {
			return fmt.Errorf("scene.invalid_target_id: %s", tag.ID)
		}
		if _, exists := r.sceneTargets[tag.ID]; exists {
			return fmt.Errorf("scene.duplicate_target_id: %s", tag.ID)
		}
		ctx := SceneContext{Surface: f.Request.Surface, Zone: f.Body, Path: node.Scene.Path, Keys: node.Scene.Keys, Notes: node.Scene.Notes}
		plan, err := r.planSceneNode(node.ID, node.Scene.Node, ctx)
		if err != nil {
			return fmt.Errorf("scene.target %s: %w", tag.ID, err)
		}
		bounds := plan.Bounds
		// A named frame/container denotes its drawn rectangle, excluding its
		// label, which may sit above the border.
		if tag.Type == "frame" || tag.Type == "container" {
			for _, item := range plan.Items {
				if item.Shape != nil && item.Shape.Record.ID == node.ID+".surface" {
					bounds = item.Shape.Record.Rect
					break
				}
			}
		}
		r.sceneTargets[tag.ID] = annotationTarget{Bounds: bounds, Top: math.Min(bounds.Y, plan.Bounds.Y)}
	}
	return nil
}

func (r *renderer) planAnnotationScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct{ Type string }
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "annotation" {
		return nil, false, nil
	}
	var n sceneAnnotationSource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	registered, exists := r.sceneTargets[n.Target]
	if !exists {
		return nil, true, fmt.Errorf("scene.annotation_unknown_target: %s", n.Target)
	}
	target := registered.Bounds
	if n.Arrow == "" {
		n.Arrow = "arrow-right-angle"
	}
	if n.Arrow != "arrow-double" && n.Arrow != "arrow-right-angle" {
		return nil, true, fmt.Errorf("scene.annotation_arrow_enum: %s", n.Arrow)
	}
	if n.Placement == "" {
		n.Placement = "below-left"
	}
	orientations := map[string][2]float64{"below-left": {-90, 1}, "below-right": {-90, -1}, "above-right": {90, 1}, "above-left": {90, -1}, "left": {0, 1}, "right": {0, -1}}
	o, valid := orientations[n.Placement]
	if !valid {
		return nil, true, fmt.Errorf("scene.annotation_placement_enum")
	}
	if n.Arrow == "arrow-double" {
		if n.Placement != "left" && n.Placement != "right" {
			return nil, true, fmt.Errorf("scene.annotation_double_requires_left_or_right")
		}
		o[0] = 0
	}
	if n.Size == 0 {
		n.Size = 66
	}
	if n.Callout.W == 0 {
		n.Callout.W = 198
	}
	if n.Size <= 0 || n.Callout.W <= 24 || math.IsNaN(n.Size+n.Callout.W) || math.IsInf(n.Size+n.Callout.W, 0) || strings.TrimSpace(n.Callout.Text) == "" {
		return nil, true, fmt.Errorf("scene.annotation_invalid_size_or_copy")
	}
	data, asset, err := r.primitiveAssetBytes(n.Arrow)
	if err != nil {
		return nil, true, err
	}
	paths, err := primitiveSVGPaths(data, true)
	if err != nil {
		return nil, true, err
	}
	ext, err := annotationPathExtrema(paths)
	if err != nil {
		return nil, true, err
	}
	tail, tip := ext[2], ext[1]
	if n.Arrow == "arrow-double" {
		tail = ext[0]
	}
	rad := o[0] * math.Pi / 180
	co, si, flip := math.Cos(rad), math.Sin(rad), o[1]
	transform := func(p [2]float64) [2]float64 { return [2]float64{flip * (p[0]*co - p[1]*si), p[0]*si + p[1]*co} }
	tail, tip = transform(tail), transform(tip)
	span := math.Max(math.Abs(tip[0]-tail[0]), math.Abs(tip[1]-tail[1]))
	if span <= 0 {
		return nil, true, fmt.Errorf("scene.annotation_arrow_no_span")
	}
	scale := n.Size / span
	anchor := [2]float64{target.X + target.W + 4, target.Y + target.H/2}
	switch n.Placement {
	case "left":
		anchor[0] = target.X - 4
	case "below-left", "below-right", "above-left", "above-right":
		fraction := .62
		if strings.HasSuffix(n.Placement, "left") {
			fraction = .38
		}
		anchor[0] = target.X + fraction*target.W
		anchor[1] = target.Y + target.H + 4
		if strings.HasPrefix(n.Placement, "above") {
			anchor[1] = registered.Top - 4
		}
	}
	ox, oy := anchor[0]-tip[0]*scale, anchor[1]-tip[1]*scale
	tail = [2]float64{ox + tail[0]*scale, oy + tail[1]*scale}
	matrix := [6]float64{scale * flip * co, scale * si, -scale * flip * si, scale * co, ox, oy}
	ink := n.Ink
	if ink == "" {
		ink = "mark"
	}
	color, err := r.sceneColor(ctx.Surface, ink)
	if err != nil {
		return nil, true, err
	}
	var svg strings.Builder
	transformed := make([]primitiveSVGPath, len(paths))
	for i, path := range paths {
		transformed[i] = path
		transformed[i].Matrix = primitiveMatrixMul(matrix, path.Matrix)
	}
	vb := primitivePathBounds(transformed)
	vb[0] -= .4
	vb[1] -= .4
	vb[2] += .8
	vb[3] += .8
	fmt.Fprintf(&svg, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%g %g %g %g">`, vb[0], vb[1], vb[2], vb[3])
	for _, path := range transformed {
		m := path.Matrix
		fmt.Fprintf(&svg, `<path fill="#%s" transform="matrix(%g %g %g %g %g %g)" d="%s"/>`, color, m[0], m[1], m[2], m[3], m[4], m[5], path.D)
	}
	svg.WriteString("</svg>")
	bytes := []byte(svg.String())
	fallback, err := primitiveRasterSVG(bytes, int(math.Ceil(vb[2]*3)), int(math.Ceil(vb[3]*3)), vb)
	if err != nil {
		return nil, true, err
	}
	arrowBox := Rect{vb[0], vb[1], vb[2], vb[3]}
	image := &pptx.ImageProps{PositionProps: pos(arrowBox), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id + ".arrow"}, DataOrPathProps: pptx.DataOrPathProps{Data: primitiveDataURI("image/svg+xml", bytes)}, SVGFallbackData: primitiveDataURI("image/png", fallback), AltText: n.Arrow + "; target=" + n.Target + "; canonical SHA256=" + asset.SHA256}
	p := &scenePlan{ID: id, Items: []sceneItem{{Image: image}}, Bounds: arrowBox}
	surface := n.Callout.Surface
	if surface == "" {
		surface = "subtle"
	}
	content := &scenePlan{ID: id + ".callout"}
	y := 12.
	labelRole, textRole := "emphasis", "display"
	if surface == "callout" {
		labelRole, textRole = "primary", "primary"
	}
	if n.Callout.Label != "" {
		st, _ := r.sceneStyle("eyebrow")
		tr, e := r.sceneDataText(content, id+".label", n.Callout.Label, st, Rect{12, y, n.Callout.W - 24, 0}, surface, labelRole, "left", ctx)
		if e != nil {
			return nil, true, e
		}
		y += tr.Rect.H + 3
	}
	st, _ := r.sceneStyle("small")
	tr, err := r.sceneDataText(content, id+".text", n.Callout.Text, st, Rect{12, y, n.Callout.W - 24, 0}, surface, textRole, "left", ctx)
	if err != nil {
		return nil, true, err
	}
	height := y + tr.Rect.H + 12
	x, top := tail[0]-n.Callout.W/2, tail[1]-6-height
	if n.Arrow == "arrow-double" || strings.HasPrefix(n.Placement, "below") || strings.HasPrefix(n.Placement, "above") {
		x = tail[0] + 6
		if strings.HasSuffix(n.Placement, "left") {
			x = tail[0] - 6 - n.Callout.W
		}
		top = tail[1] - height/2
	}
	box := Rect{x, top, n.Callout.W, height}
	if !inside(box, ctx.Zone) || !inside(arrowBox, Rect{0, 0, 960, 540}) {
		return nil, true, fmt.Errorf("scene.annotation_outside_zone: %s callout %+v", id, box)
	}
	if surface == "outline" {
		if err = r.primitiveShape(p, id+".callout.surface", box, "light", "line", true); err != nil {
			return nil, true, err
		}
	} else if err = r.sceneRect(p, id+".callout.surface", box, surface); err != nil {
		return nil, true, err
	}
	// Draw again at the resolved position; rich text may contain separate mark
	// images and groups, so translating only its text would be incomplete.
	if n.Callout.Label != "" {
		ls, _ := r.sceneStyle("eyebrow")
		if _, err = r.sceneDataText(p, id+".label", n.Callout.Label, ls, Rect{x + 12, top + 12, n.Callout.W - 24, 0}, surface, labelRole, "left", ctx); err != nil {
			return nil, true, err
		}
	}
	if _, err = r.sceneDataText(p, id+".text", n.Callout.Text, st, Rect{x + 12, top + y, n.Callout.W - 24, 0}, surface, textRole, "left", ctx); err != nil {
		return nil, true, err
	}
	p.Bounds = diagramUnion(arrowBox, box)
	primitiveFinish(p, "annotation")
	p.Warnings = append(p.Warnings, "Named annotation target "+n.Target+"; canonical arrow tip anchored 4pt outside target border (above placements clear any top label); callout measured with 12pt padding and 6pt tail gap.")
	return p, true, nil
}

// Extrema are sampled from the canonical transformed artwork, not its viewBox.
func annotationPathExtrema(paths []primitiveSVGPath) ([4][2]float64, error) {
	ext := [4][2]float64{{math.Inf(1), 0}, {math.Inf(-1), 0}, {0, math.Inf(1)}, {0, math.Inf(-1)}}
	var x, y float64
	save := func(a, b float64) {
		if a < ext[0][0] {
			ext[0] = [2]float64{a, b}
		}
		if a > ext[1][0] {
			ext[1] = [2]float64{a, b}
		}
		if b < ext[2][1] {
			ext[2] = [2]float64{a, b}
		}
		if b > ext[3][1] {
			ext[3] = [2]float64{a, b}
		}
		x, y = a, b
	}
	sink := primitivePathSink{Move: save, Line: save, Close: func() {}}
	sink.Cube = func(a, b, c, d, e, f float64) {
		ox, oy := x, y
		for i := 1; i <= 256; i++ {
			t := float64(i) / 256
			u := 1 - t
			save(u*u*u*ox+3*u*u*t*a+3*u*t*t*c+t*t*t*e, u*u*u*oy+3*u*u*t*b+3*u*t*t*d+t*t*t*f)
		}
	}
	sink.Quad = func(a, b, c, d float64) {
		ox, oy := x, y
		for i := 1; i <= 256; i++ {
			t := float64(i) / 256
			u := 1 - t
			save(u*u*ox+2*u*t*a+t*t*c, u*u*oy+2*u*t*b+t*t*d)
		}
	}
	for _, path := range paths {
		if err := primitiveWalkPath(path, sink); err != nil {
			return ext, err
		}
	}
	if len(paths) == 0 {
		return ext, fmt.Errorf("scene.annotation_empty_arrow_artwork")
	}
	return ext, nil
}
