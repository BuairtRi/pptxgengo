package wmdesign

import (
	"encoding/json"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

func (r *renderer) applyNativeProfileComponent(n Node, plan componentPlan, zone Rect, surface string) componentPlan {
	if r.editingProfile != NativeEditingProfile || n.Card == nil {
		return plan
	}
	if nativeProfileComponentPlainCard(n, plan) {
		if candidate, ok := r.nativeProfileComponentCard(n, plan, zone, surface); ok {
			return candidate
		}
	}
	for _, block := range n.Card.Body {
		if len(block.Bullets) == 0 {
			continue
		}
		if candidate, ok := r.nativeProfileComponentBulletBlock(n, plan, block, zone); ok {
			plan = candidate
		}
	}
	return plan
}

func nativeProfileComponentPlainCard(n Node, plan componentPlan) bool {
	if n.Card == nil || len(plan.record.Shapes) != 1 || plan.record.Shapes[0].ID != n.ID+".container" || n.Card.Edge != nil || n.Card.Band != nil || n.Card.InlineNumber != "" || n.Card.NumberInk != "" || n.Card.Metric != nil || n.Card.MetricGroup != nil {
		return false
	}
	if len(n.Card.Body) == 0 {
		return false
	}
	for _, block := range n.Card.Body {
		if block.Paragraph == "" || len(block.Bullets) != 0 {
			return false
		}
	}
	return true
}

func (r *renderer) nativeProfileComponentCard(n Node, plan componentPlan, zone Rect, surface string) (componentPlan, bool) {
	shape := plan.record.Shapes[0]
	items := []sceneItem{{Shape: &sceneShape{Type: pptx.ShapeTypeRect, Props: pptx.ShapeProps{PositionProps: pos(shape.Rect), ObjectNameProps: pptx.ObjectNameProps{ObjectName: shape.ID}, Fill: &pptx.ShapeFillProps{Color: shape.Color}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}, Record: shape}}}
	for i := range plan.texts {
		tr := &plan.texts[i]
		items = append(items, sceneItem{Text: tr})
	}
	scene := &scenePlan{ID: n.ID, Bounds: plan.record.Rect, Items: items}
	body := make([]json.RawMessage, 0, len(n.Card.Body))
	for _, block := range n.Card.Body {
		encoded, err := json.Marshal(map[string]string{"p": block.Paragraph})
		if err != nil {
			return plan, false
		}
		body = append(body, encoded)
	}
	var label, title string
	if n.Card.Label != "" {
		label = n.Card.Label
	}
	if n.Card.Title != "" {
		title = n.Card.Title
	}
	raw, err := json.Marshal(struct {
		Type    string            `json:"type"`
		X       float64           `json:"x"`
		Y       float64           `json:"y"`
		W       float64           `json:"w"`
		H       float64           `json:"h"`
		Surface string            `json:"surface,omitempty"`
		Label   string            `json:"label,omitempty"`
		Title   string            `json:"title,omitempty"`
		Body    []json.RawMessage `json:"body"`
	}{Type: "card", X: plan.record.Rect.X, Y: plan.record.Rect.Y, W: plan.record.Rect.W, H: plan.record.Rect.H, Surface: surface, Label: label, Title: title, Body: body})
	if err != nil {
		return plan, false
	}
	candidate, _ := nativeProfileCardResult(scene, raw)
	if candidate == nil || len(candidate.Items) != 1 || candidate.Items[0].Text == nil {
		return plan, false
	}
	tr := *candidate.Items[0].Text
	tr.ID = n.ID
	if !inside(tr.Rect, plan.record.Rect) || sceneTextEnvelope(&scenePlan{ID: n.ID, Items: []sceneItem{{Text: &tr}}}, SceneContext{Zone: zone}) != nil {
		return plan, false
	}
	plan.texts = []TextRecord{tr}
	plan.record.Parts = []string{n.ID}
	plan.record.Shapes = nil
	plan.record.NativeObject = true
	return plan, true
}

func (r *renderer) nativeProfileComponentBulletBlock(n Node, plan componentPlan, block BodyBlock, zone Rect) (componentPlan, bool) {
	if len(block.Bullets) == 0 || !validPartKey(block.Key) {
		return plan, false
	}
	listID := n.ID + ".body." + block.Key
	shapeByID := make(map[string]ShapeRecord, len(plan.record.Shapes))
	for _, shape := range plan.record.Shapes {
		shapeByID[shape.ID] = shape
	}
	textByID := make(map[string]*TextRecord, len(plan.texts))
	for i := range plan.texts {
		textByID[plan.texts[i].ID] = &plan.texts[i]
	}
	items := make([]string, len(block.Bullets))
	sceneItems := make([]sceneItem, 0, len(block.Bullets)*2)
	removeShapes, removeParts, removeTexts := map[string]bool{}, map[string]bool{}, map[string]bool{}
	firstMarker := ""
	for i, bullet := range block.Bullets {
		if !validPartKey(bullet.Key) || strings.TrimSpace(bullet.Text) == "" {
			return plan, false
		}
		markerID := n.ID + ".body." + block.Key + "." + bullet.Key + ".marker"
		textID := n.ID + ".body." + block.Key + "." + bullet.Key
		marker, markerOK := shapeByID[markerID]
		text, textOK := textByID[textID]
		if !markerOK || !textOK {
			return plan, false
		}
		if i == 0 {
			firstMarker = markerID
		}
		items[i] = bullet.Text
		sceneItems = append(sceneItems,
			sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeRect, Props: pptx.ShapeProps{PositionProps: pos(marker.Rect), ObjectNameProps: pptx.ObjectNameProps{ObjectName: marker.ID}, Fill: &pptx.ShapeFillProps{Color: marker.Color}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}, Record: marker}},
			sceneItem{Text: text},
		)
		removeShapes[markerID], removeParts[markerID], removeParts[textID], removeTexts[textID] = true, true, true, true
	}
	first := sceneItems[1].Text.Rect
	marker := sceneItems[0].Shape.Record.Rect
	x, y, w := marker.X, first.Y, first.X-marker.X+first.W
	size := ""
	if plan.record.Density == "dense" {
		size = "small"
	}
	raw, err := json.Marshal(struct {
		Type  string   `json:"type"`
		X     float64  `json:"x"`
		Y     float64  `json:"y"`
		W     float64  `json:"w"`
		Size  string   `json:"size,omitempty"`
		Items []string `json:"items"`
	}{Type: "bullets", X: x, Y: y, W: w, Size: size, Items: items})
	if err != nil {
		return plan, false
	}
	temp := &scenePlan{ID: listID, Bounds: plan.record.Rect, Items: sceneItems}
	candidate, _ := r.nativeProfileListResult(temp, raw)
	if candidate == nil || len(candidate.Items) != 1 || candidate.Items[0].Text == nil {
		return plan, false
	}
	tr := *candidate.Items[0].Text
	if !inside(tr.Rect, plan.record.Rect) || sceneTextEnvelope(candidate, SceneContext{Zone: zone}) != nil {
		return plan, false
	}
	textIndex := -1
	texts := make([]TextRecord, 0, len(plan.texts)-len(block.Bullets)+1)
	for _, record := range plan.texts {
		if removeTexts[record.ID] {
			if textIndex < 0 {
				textIndex = len(texts)
				texts = append(texts, tr)
			}
			continue
		}
		texts = append(texts, record)
	}
	if textIndex < 0 {
		return plan, false
	}
	parts := make([]string, 0, len(plan.record.Parts)-2*len(block.Bullets)+1)
	for _, part := range plan.record.Parts {
		if part == firstMarker {
			parts = append(parts, listID)
		}
		if !removeParts[part] {
			parts = append(parts, part)
		}
	}
	shapes := make([]ShapeRecord, 0, len(plan.record.Shapes)-len(block.Bullets))
	for _, shape := range plan.record.Shapes {
		if !removeShapes[shape.ID] {
			shapes = append(shapes, shape)
		}
	}
	plan.texts, plan.record.Parts, plan.record.Shapes = texts, parts, shapes
	return plan, true
}
