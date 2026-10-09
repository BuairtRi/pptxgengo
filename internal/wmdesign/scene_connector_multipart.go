package wmdesign

import (
	"fmt"
	"math"

	"github.com/buairtri/pptxgengo/pptx"
)

// appendNativeRoute uses an editable preset when possible. Other paths retain
// every vertex as a closed, ordered namespace of native attached segments and
// invisible waypoint rectangles. Both adjacent segments use the same top site,
// so their junction remains the exact authored waypoint, not a guessed corner.
func appendNativeRoute(p *scenePlan, id string, points [][2]float64, line *pptx.ShapeLineProps, color string, connection pptx.ConnectorConnection) error {
	if len(points) < 2 || len(points) > 128 {
		return fmt.Errorf("native route requires 2..128 vertices")
	}
	for i, q := range points {
		if math.IsNaN(q[0]+q[1]) || math.IsInf(q[0]+q[1], 0) || i > 0 && q == points[i-1] {
			return fmt.Errorf("invalid native route vertex")
		}
	}
	bounds := nativeRouteEnvelope(points)
	base := pptx.ShapeProps{PositionProps: pos(bounds), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Line: line}
	if len(points) == 2 {
		fh, fv := points[0][0] > points[1][0], points[0][1] > points[1][1]
		base.FlipH, base.FlipV = &fh, &fv
		p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeLine, Connection: &connection, Props: base, Record: ShapeRecord{ID: id, Rect: bounds, Color: color, Geometry: "native-semantic-connector"}}})
		return nil
	}
	if bounds.W > 0 && bounds.H > 0 {
		normalized := make([][2]float64, len(points))
		for i, q := range points {
			normalized[i] = [2]float64{(q[0] - bounds.X) / bounds.W, (q[1] - bounds.Y) / bounds.H}
		}
		preset, props, e := pptx.NativePolylineConnectorPreset(base, normalized)
		if e == nil {
			route := &pptx.ConnectorRoute{Preset: string(preset), Adjustment: props.Adjustments["adj1"], Adjustment2: props.Adjustments["adj2"], Adjustment3: props.Adjustments["adj3"]}
			props.Adjustments = nil
			p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeLine, Route: route, Connection: &connection, Props: props, Record: ShapeRecord{ID: id, Rect: bounds, Color: color, Geometry: "native-semantic-connector"}}})
			return nil
		}
	}
	const guideSize = 1.0 / 12700 // one native EMU in points; top-center is the waypoint.
	for i, q := range points[1 : len(points)-1] {
		name := fmt.Sprintf("%s.waypoint-%03d", id, i+1)
		rect := Rect{X: q[0] - guideSize/2, Y: q[1], W: guideSize, H: guideSize}
		p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeRect, Props: pptx.ShapeProps{PositionProps: pos(rect), ObjectNameProps: pptx.ObjectNameProps{ObjectName: name}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}, Record: ShapeRecord{ID: name, Rect: rect, Geometry: "native-route-waypoint-guide"}}})
	}
	for i := 0; i < len(points)-1; i++ {
		a, z := points[i], points[i+1]
		rect := nativeRouteEnvelope([][2]float64{a, z})
		fh, fv := a[0] > z[0], a[1] > z[1]
		name := fmt.Sprintf("%s.segment-%03d", id, i+1)
		endpoints := connection
		if i > 0 {
			endpoints.Begin = pptx.ConnectorEndpoint{ObjectName: fmt.Sprintf("%s.waypoint-%03d", id, i), Site: 0}
		}
		if i < len(points)-2 {
			endpoints.End = pptx.ConnectorEndpoint{ObjectName: fmt.Sprintf("%s.waypoint-%03d", id, i+1), Site: 0}
		}
		segmentLine := *line
		if i > 0 {
			segmentLine.BeginArrowType = ""
		}
		if i < len(points)-2 {
			segmentLine.EndArrowType = ""
		}
		p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeLine, Connection: &endpoints, Props: pptx.ShapeProps{PositionProps: pos(rect), ObjectNameProps: pptx.ObjectNameProps{ObjectName: name}, Line: &segmentLine, FlipH: &fh, FlipV: &fv}, Record: ShapeRecord{ID: name, Rect: rect, Color: color, Geometry: "native-semantic-connector-segment"}}})
	}
	p.Warnings = append(p.Warnings, "One logical connection is represented by ordered attached native segments and transparent waypoint guides; preserve their complete authored namespace when adopting geometry.")
	return nil
}
func nativeRouteEnvelope(points [][2]float64) Rect {
	minX, minY, maxX, maxY := points[0][0], points[0][1], points[0][0], points[0][1]
	for _, q := range points {
		minX = math.Min(minX, q[0])
		minY = math.Min(minY, q[1])
		maxX = math.Max(maxX, q[0])
		maxY = math.Max(maxY, q[1])
	}
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}
