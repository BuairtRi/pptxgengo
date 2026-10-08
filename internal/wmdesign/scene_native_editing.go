package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

// Native editing components are explicit authoring choices. Shared source
// blocks and grouped path connectors retain their current serialization.
type nativeEditingEndpoint struct {
	Node string `json:"node"`
	Site string `json:"site"`
}
type nativeEditingSource struct {
	Type    string                 `json:"type"`
	X       float64                `json:"x"`
	Y       float64                `json:"y"`
	W       float64                `json:"w"`
	H       float64                `json:"h"`
	Surface string                 `json:"surface,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Style   string                 `json:"style,omitempty"`
	Align   string                 `json:"align,omitempty"`
	Ink     string                 `json:"ink,omitempty"`
	Head    string                 `json:"head,omitempty"`
	From    *nativeEditingEndpoint `json:"from,omitempty"`
	To      *nativeEditingEndpoint `json:"to,omitempty"`
}

// Register authored top-level editable-block IDs before planning/paint. This
// permits forward references without matching text, geometry or paint order.
func (r *renderer) registerEditableTargets(slide SlideSpec) error {
	r.editableTargets = map[string]Rect{}
	r.editableTargetNames = map[string]string{}
	needsStockTargets := false
	for _, node := range slide.Nodes {
		if node.Scene != nil {
			var tag struct{ Type string }
			if e := json.Unmarshal(node.Scene.Node, &tag); e != nil {
				return e
			}
			needsStockTargets = needsStockTargets || tag.Type == "attached-connector"
		}
	}
	for _, node := range slide.Nodes {
		if node.Scene == nil {
			continue
		}
		var tag struct {
			Type       string
			X, Y, W, H float64
		}
		if e := json.Unmarshal(node.Scene.Node, &tag); e != nil {
			return e
		}
		if tag.Type != "editable-block" && (tag.Type != "block" || !needsStockTargets) {
			continue
		}
		if !validPartKey(node.ID) {
			return fmt.Errorf("scene.invalid_editable_target_id: %s", node.ID)
		}
		if _, exists := r.editableTargets[node.ID]; exists {
			return fmt.Errorf("scene.duplicate_editable_target: %s", node.ID)
		}
		b := Rect{tag.X, tag.Y, tag.W, tag.H}
		if b.W <= 24 || b.H <= 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
			return fmt.Errorf("scene.invalid_editable_target_bounds: %s", node.ID)
		}
		r.editableTargets[node.ID] = b
		r.editableTargetNames[node.ID] = node.ID
		if tag.Type == "block" {
			r.editableTargetNames[node.ID] = node.ID + ".surface"
		}
	}
	return nil
}

func nativeRectangleSite(b Rect, site string) ([2]float64, int, error) {
	switch site {
	case "top":
		return [2]float64{b.X + b.W/2, b.Y}, 0, nil
	case "left":
		return [2]float64{b.X, b.Y + b.H/2}, 1, nil
	case "bottom":
		return [2]float64{b.X + b.W/2, b.Y + b.H}, 2, nil
	case "right":
		return [2]float64{b.X + b.W, b.Y + b.H/2}, 3, nil
	default:
		return [2]float64{}, 0, fmt.Errorf("scene.native_connection_site_requires_top_left_bottom_or_right")
	}
}

func (r *renderer) planNativeEditingScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type == "editable-list" {
		p, e := r.planEditableList(id, raw, ctx)
		return p, true, e
	}
	if tag.Type == "editable-card" {
		p, err := r.planEditableCard(id, raw, ctx)
		return p, true, err
	}
	if tag.Type != "editable-block" && tag.Type != "attached-connector" {
		return nil, false, nil
	}
	var n nativeEditingSource
	if e := sceneDecode(raw, &n); e != nil {
		return nil, true, e
	}
	b := Rect{n.X, n.Y, n.W, n.H}
	if b.W < 0 || b.H < 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
		return nil, true, fmt.Errorf("scene.invalid_native_editing_allocation")
	}
	if n.Type == "editable-block" {
		if n.Text == "" || n.From != nil || n.To != nil || n.Head != "" || n.Ink != "" {
			return nil, true, fmt.Errorf("scene.editable_block_requires_plain_text_and_no_connection_options")
		}
		// Reuse the existing measured block typography/padding, then fold its one
		// surface and one plain text record into the same native rectangle.
		adapted, _ := json.Marshal(map[string]any{"type": "block", "x": n.X, "y": n.Y, "w": n.W, "h": n.H, "surface": n.Surface, "text": n.Text, "style": n.Style, "align": n.Align})
		p, _, e := r.planDiagramScene(id, adapted, ctx)
		if e != nil {
			return nil, true, e
		}
		var text *TextRecord
		fill := ""
		for _, item := range p.Items {
			if item.Shape != nil {
				if item.Shape.Record.ID != id+".surface" || item.Shape.Type != pptx.ShapeTypeRect || item.Shape.Props.Fill == nil || item.Shape.Props.Line == nil || item.Shape.Props.Line.Type != "none" {
					return nil, true, fmt.Errorf("scene.editable_block_surface_not_simple")
				}
				fill = item.Shape.Props.Fill.Color
			}
			if item.Text != nil {
				if text != nil || item.Text.Rich != nil {
					return nil, true, fmt.Errorf("scene.editable_block_rich_text_unsupported")
				}
				copied := *item.Text
				text = &copied
			}
		}
		if len(p.Items) != 2 || text == nil || fill == "" || !inside(text.Rect, b) {
			return nil, true, fmt.Errorf("scene.editable_block_requires_one_plain_field")
		}
		text.ID = id
		text.NativeShape = &NativeTextShape{Rect: b, Fill: fill}
		p.Items, p.Groups, p.Bounds = []sceneItem{{Text: text}}, nil, b
		p.Definition = "scene.editable-block"
		p.Warnings = append(p.Warnings, "Editable block pilot combines fill and plain text in one native rectangle; desktop move/resize and visual qualification remain pending.")
		return p, true, nil
	}
	if n.From == nil || n.To == nil || !validPartKey(n.From.Node) || !validPartKey(n.To.Node) || n.From.Node == n.To.Node || n.Text != "" || n.Surface != "" || n.Align != "" {
		return nil, true, fmt.Errorf("scene.attached_connector_requires_distinct_editable_block_endpoints")
	}
	from, ok := r.editableTargets[n.From.Node]
	if !ok {
		return nil, true, fmt.Errorf("scene.native_connection_unknown_editable_block: %s", n.From.Node)
	}
	to, ok := r.editableTargets[n.To.Node]
	if !ok {
		return nil, true, fmt.Errorf("scene.native_connection_unknown_editable_block: %s", n.To.Node)
	}
	a, start, e := nativeRectangleSite(from, n.From.Site)
	if e != nil {
		return nil, true, e
	}
	z, end, e := nativeRectangleSite(to, n.To.Site)
	if e != nil {
		return nil, true, e
	}
	if a == z {
		return nil, true, fmt.Errorf("scene.native_connection_coincident_sites")
	}
	bounds := Rect{math.Min(a[0], z[0]), math.Min(a[1], z[1]), math.Abs(a[0] - z[0]), math.Abs(a[1] - z[1])}
	ink := n.Ink
	if ink == "" {
		ink = "strong"
	}
	color, e := r.sceneColor(ctx.Surface, ink)
	if e != nil {
		return nil, true, e
	}
	dash, weight := "solid", 1.0
	switch n.Style {
	case "", "solid":
	case "dashed":
		dash = "dash"
	case "dotted":
		dash, weight = "sysDot", 1.5
	default:
		return nil, true, fmt.Errorf("scene.unknown_native_connector_style")
	}
	head := n.Head
	if head == "" {
		head = "end"
	}
	line := &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: weight, DashType: dash}
	switch head {
	case "none":
	case "end":
		line.EndArrowType = "triangle"
	case "start":
		line.BeginArrowType = "triangle"
	case "both":
		line.BeginArrowType, line.EndArrowType = "triangle", "triangle"
	default:
		return nil, true, fmt.Errorf("scene.unknown_native_connector_head")
	}
	flipH, flipV := a[0] > z[0], a[1] > z[1]
	fromName, toName := r.editableTargetNames[n.From.Node], r.editableTargetNames[n.To.Node]
	if fromName == "" {
		fromName = n.From.Node
	}
	if toName == "" {
		toName = n.To.Node
	}
	connection := &pptx.ConnectorConnection{Begin: pptx.ConnectorEndpoint{ObjectName: fromName, Site: start}, End: pptx.ConnectorEndpoint{ObjectName: toName, Site: end}}
	shape := &sceneShape{Type: pptx.ShapeTypeLine, Props: pptx.ShapeProps{PositionProps: pos(bounds), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Line: line, FlipH: &flipH, FlipV: &flipV}, Record: ShapeRecord{ID: id, Rect: bounds, Color: color, Geometry: "native-straight-connector"}, Connection: connection}
	p := &scenePlan{Definition: "scene.attached-connector", ID: id, Bounds: bounds, Items: []sceneItem{{Shape: shape}}, Warnings: []string{"Native attached connector pilot uses explicit same-slide rectangular sites; routing, node movement and Save As require desktop qualification."}}
	if strings.Contains(ctx.Path, "/body/") {
		p.Warnings = append(p.Warnings, "Shared-library connector rollout is not qualified by this authoring pilot.")
	}
	return p, true, nil
}
