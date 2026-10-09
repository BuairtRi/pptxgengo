package wmdesign

import (
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"strings"
)

// Keep the established measured label and patch, but serialize a semantic
// relationship as one attached connector rather than decorative segments.
func (r *renderer) planSemanticConnector(p *scenePlan, id string, points [][2]float64, label string, connection pptx.ConnectorConnection, ctx SceneContext) error {
	pts := [][2]float64{}
	for _, point := range points {
		if len(pts) > 0 && point == pts[len(pts)-1] {
			continue
		}
		if len(pts) > 1 {
			a, b := pts[len(pts)-2], pts[len(pts)-1]
			if (a[0] == b[0] && b[0] == point[0] && (b[1]-a[1])*(point[1]-b[1]) >= 0) || (a[1] == b[1] && b[1] == point[1] && (b[0]-a[0])*(point[0]-b[0]) >= 0) {
				pts = pts[:len(pts)-1]
			}
		}
		pts = append(pts, point)
	}
	if len(pts) < 2 {
		return fmt.Errorf("semantic connector requires distinct endpoints")
	}
	old := &scenePlan{ID: id}
	if e := r.planDiagramConnector(old, id, diagramSpec{Points: pts, Label: label, Head: "end", Ink: "strong"}, ctx); e != nil {
		return e
	}
	color, e := r.sceneColor(ctx.Surface, "strong")
	if e != nil {
		return e
	}
	line := &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: 1, EndArrowType: "triangle"}
	if e := appendNativeRoute(p, id, pts, line, color, connection); e != nil {
		return e
	}
	for _, item := range old.Items {
		if item.Text != nil || item.Shape != nil && strings.HasPrefix(item.Shape.Record.ID, id+".label-") {
			p.Items = append(p.Items, item)
		}
	}
	return nil
}
