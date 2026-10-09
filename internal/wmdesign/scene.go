package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

// SceneSpec is the source-owned, strictly decoded component input. Content
// bindings project named slots into the pinned source node; this does not expose
// source pointers or geometry mutation through the template authoring API.
type SceneSpec struct {
	Allocation  *Rect               `json:"allocation,omitempty"`
	Node        json.RawMessage     `json:"node"`
	Path        string              `json:"source_pointer"`
	Keys        map[string][]string `json:"keys,omitempty"`
	Notes       []string            `json:"notes,omitempty"`
	Resolutions []string            `json:"resolutions,omitempty"`
}

type SceneContext struct {
	Surface string
	Zone    Rect
	Path    string
	Keys    map[string][]string
	Notes   []string
}

type scenePlan struct {
	Definition string
	ID         string
	Bounds     Rect
	Items      []sceneItem
	Groups     []ComponentRecord
	Warnings   []string
	// Only an auto-height callout owns this text-derived container envelope.
	// It permits a density retry without changing any authored outer rectangle.
	TextFlowBounds bool
}

type sceneItem struct {
	Shape *sceneShape
	Text  *TextRecord
	Image *pptx.ImageProps
	Table *sceneTable
	Chart *sceneChart
}

type sceneShape struct {
	Route      *pptx.ConnectorRoute
	Connection *pptx.ConnectorConnection
	Type       pptx.ShapeType
	Props      pptx.ShapeProps
	Record     ShapeRecord
}

type sceneTableCell struct {
	Row    int
	Column int
	Text   TextRecord
}

type sceneTable struct {
	ID          string
	Rect        Rect
	Rows        []pptx.TableRow
	Options     pptx.TableProps
	Texts       []TextRecord
	CellRecords []sceneTableCell
}

type sceneChart struct {
	ID      string
	Rect    Rect
	Type    pptx.ChartType
	Data    []pptx.ChartData
	Options pptx.ChartOptions
}

// Append retains paint order and bottom-up group ownership. Parent groups may
// refer to child group IDs instead of flattening their constituent objects.
func (p *scenePlan) Append(child *scenePlan) {
	if child == nil {
		return
	}
	p.Items = append(p.Items, child.Items...)
	p.Groups = append(p.Groups, child.Groups...)
	p.Warnings = append(p.Warnings, child.Warnings...)
}

func sceneDecode(raw json.RawMessage, v any) error {
	if err := bindingStrictDecode(raw, v); err != nil {
		return fmt.Errorf("scene.unsupported_source_field: %w", err)
	}
	return nil
}

func (r *renderer) sceneStyle(name string) (Style, error) { return r.bodyStyle(name) }

func (r *renderer) sceneColor(surface, ref string) (string, error) {
	if surface == "" {
		surface = "light"
	}
	if ref == "" {
		ref = "primary"
	}
	if color, err := r.source.Ink(surface, ref); err == nil {
		return color, nil
	}
	// Resolve exclusively from pinned tokens, including colors not retained by
	// the original foundation's narrower typed token subset.
	raw, err := os.ReadFile(filepath.Join(r.source.Root, "tokens/v0/tokens.json"))
	if err != nil {
		return "", err
	}
	var tokens struct {
		Colors struct {
			Brand []struct {
				ID  string `json:"id"`
				Hex string `json:"hex"`
			} `json:"brand"`
			Ramp []struct {
				Step int    `json:"step"`
				Hex  string `json:"hex"`
			} `json:"ramp"`
			Dataviz struct {
				Categorical []string          `json:"categorical"`
				Deemphasis  []string          `json:"deemphasis"`
				KPI         map[string]string `json:"kpi"`
			} `json:"dataviz"`
		} `json:"colors"`
	}
	if err = json.Unmarshal(raw, &tokens); err != nil {
		return "", err
	}
	clean := func(color string) (string, error) {
		color = strings.TrimPrefix(color, "#")
		if len(color) != 6 {
			return "", fmt.Errorf("scene.invalid_token_color: %s", color)
		}
		if _, e := strconv.ParseUint(color, 16, 24); e != nil {
			return "", e
		}
		return strings.ToUpper(color), nil
	}
	parts := strings.Split(ref, ".")
	if len(parts) == 2 {
		switch parts[0] {
		case "series", "deemph":
			i, e := strconv.Atoi(parts[1])
			if e != nil {
				return "", fmt.Errorf("scene.invalid_palette_reference: %s", ref)
			}
			colors := tokens.Colors.Dataviz.Categorical
			if parts[0] == "deemph" {
				colors = tokens.Colors.Dataviz.Deemphasis
			}
			if i < 1 || i > len(colors) {
				return "", fmt.Errorf("scene.invalid_palette_reference: %s", ref)
			}
			return clean(colors[i-1])
		case "kpi":
			if color, ok := tokens.Colors.Dataviz.KPI[parts[1]]; ok {
				return clean(color)
			}
		case "brand":
			ref = parts[1]
		case "ramp":
			i, e := strconv.Atoi(parts[1])
			if e == nil {
				for _, item := range tokens.Colors.Ramp {
					if item.Step == i {
						return clean(item.Hex)
					}
				}
			}
		}
	}
	for _, item := range tokens.Colors.Brand {
		if item.ID == ref {
			return clean(item.Hex)
		}
	}
	// Explicit literal colors must be an exact frozen palette member.
	if strings.HasPrefix(ref, "#") {
		color, e := clean(ref)
		if e != nil {
			return "", e
		}
		for _, item := range tokens.Colors.Brand {
			if c, _ := clean(item.Hex); c == color {
				return color, nil
			}
		}
		for _, item := range tokens.Colors.Ramp {
			if c, _ := clean(item.Hex); c == color {
				return color, nil
			}
		}
		for _, list := range [][]string{tokens.Colors.Dataviz.Categorical, tokens.Colors.Dataviz.Deemphasis} {
			for _, item := range list {
				if c, _ := clean(item); c == color {
					return color, nil
				}
			}
		}
		for _, item := range tokens.Colors.Dataviz.KPI {
			if c, _ := clean(item); c == color {
				return color, nil
			}
		}
	}
	return "", fmt.Errorf("scene.unknown_color_reference: %s on %s", ref, surface)
}

func (r *renderer) sceneText(p *scenePlan, id, text string, style Style, box Rect, surface, ink, align string) error {
	if text == "" {
		return nil
	}
	if strings.Contains(text, "[[") || strings.Contains(text, "[^") {
		return r.primitiveRichText(p, id, text, style, box, surface, ink, align, "underscore", "", r.sceneContext)
	}
	if box.W <= 0 || box.H < 0 || math.IsNaN(box.X+box.Y+box.W+box.H) || math.IsInf(box.X+box.Y+box.W+box.H, 0) {
		return fmt.Errorf("scene.invalid_text_geometry: %s", id)
	}
	if align == "" {
		align = "left"
	}
	if align != "left" && align != "right" && align != "center" {
		return fmt.Errorf("scene.unsupported_align: %s", align)
	}
	l, err := r.measureText(text, style, box.W)
	if err != nil {
		return fmt.Errorf("%s: %w", id, err)
	}
	need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
	if box.H == 0 {
		box.H = need
	}
	if need > box.H+.02 {
		return fmt.Errorf("scene.text_vertical_overflow: %s needs %.3fpt, capacity %.3fpt", id, need, box.H)
	}
	color, err := r.sceneColor(surface, ink)
	if err != nil {
		return err
	}
	r.auditTextContrast(id, style, color, surface)
	p.Items = append(p.Items, sceneItem{Text: &TextRecord{ID: id, Rect: box, Color: color, Align: align, Layout: l}})
	return nil
}

func (r *renderer) sceneRect(p *scenePlan, id string, box Rect, surface string) error {
	color, err := r.sceneColor(surface, "bg")
	if err != nil {
		return err
	}
	if box.W <= 0 || box.H <= 0 {
		return fmt.Errorf("scene.invalid_shape_geometry: %s", id)
	}
	props := pptx.ShapeProps{PositionProps: pos(box), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Color: color}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}
	if isModernLibrary(r.source.Revision) && surface == "outline" {
		// The source surface uses an inset 1pt border. Keep its outer bounds
		// fixed while centering the native stroke half a point inside them.
		if box.W <= 1 || box.H <= 1 {
			return fmt.Errorf("scene.invalid_outline_geometry: %s", id)
		}
		props.PositionProps = pos(Rect{box.X + .5, box.Y + .5, box.W - 1, box.H - 1})
		props.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: "CED7E6"}, Width: 1}
	}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeRect, Props: props, Record: ShapeRecord{ID: id, Rect: box, Color: color}}})
	return nil
}

func (r *renderer) planSceneNode(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	previous := r.sceneContext
	r.sceneContext = ctx
	defer func() { r.sceneContext = previous }()
	if r.typeEngine.engine != CandidateEngine {
		return nil, fmt.Errorf("scene.requires_v2")
	}
	for _, handler := range []func(string, json.RawMessage, SceneContext) (*scenePlan, bool, error){r.planAssessmentScene, r.planNativeEditingScene, r.planRoadForkScene, r.planIntakeGaugeScene, r.planIntakeCycleScene, r.planIntakeRoadScene, r.planIntakeScoreLegendScene, r.planIntakeRound12Scene, r.planIntakeVennScene, r.planIntakeMaturityScene, r.planIntakeArchitectureScene, r.planIntakeGeographyScene, r.planIntakeCurveScene, r.planAnnotationScene, r.planSourceRule, r.planPrimitiveScene, r.planMediaScene, r.planCardScene, r.planTableScene, r.planChartScene, r.planDiagramScene, r.planSequenceScene, r.planPeopleScene} {
		plan, handled, err := handler(id, raw, ctx)
		if handled || err != nil {
			if err == nil && r.contrastProbe == nil {
				err = sceneTextEnvelope(plan, ctx)
				if err == nil {
					plan = r.applyNativeEditingProfile(plan, raw, ctx)
				}
			}
			return plan, err
		}
	}
	return nil, fmt.Errorf("scene.unsupported_node: %s", id)
}

// Decorations can deliberately bleed. Editable content must remain inside the
// canvas and the frame's declared source/footer reservation.
func sceneTextEnvelope(p *scenePlan, ctx SceneContext) error {
	if p == nil {
		return fmt.Errorf("scene.empty_plan")
	}
	canvas := Rect{0, 0, 960, 540}
	for _, item := range p.Items {
		var box Rect
		var id string
		switch {
		case item.Text != nil:
			box = diagramRotatedRect(item.Text.Rect, item.Text.Rotation)
			id = item.Text.ID
		case item.Table != nil:
			box = item.Table.Rect
			id = item.Table.ID
		case item.Chart != nil:
			box = item.Chart.Rect
			id = item.Chart.ID
		default:
			continue
		}
		if !inside(box, canvas) {
			err := fmt.Errorf("scene.content_outside_canvas: %s at %+v", id, box)
			if item.Text != nil {
				return &densityFitError{cause: err}
			}
			return err
		}
		if ctx.Zone.H > 0 && box.Y+box.H > ctx.Zone.Y+ctx.Zone.H+.02 {
			err := fmt.Errorf("scene.content_overlaps_reserved_footer: %s bottom %.3fpt capacity %.3fpt", id, box.Y+box.H, ctx.Zone.Y+ctx.Zone.H)
			if item.Text != nil {
				return &densityFitError{cause: err}
			}
			return err
		}
	}
	return nil
}
