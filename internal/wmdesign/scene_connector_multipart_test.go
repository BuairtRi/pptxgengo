package wmdesign

import (
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"math"
	"testing"
)

func TestMultipartNativeConnectorClosedPartsPortsAndEndArrows(t *testing.T) {
	points := [][2]float64{{10, 90}, {40, 90}, {40, 20}, {140, 20}, {140, 90}, {180, 90}}
	connection := pptx.ConnectorConnection{Begin: pptx.ConnectorEndpoint{ObjectName: "input", Site: 3}, End: pptx.ConnectorEndpoint{ObjectName: "output", Site: 1}}
	line := &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: "112233"}, Width: 1, BeginArrowType: "triangle", EndArrowType: "triangle"}
	plan := &scenePlan{}
	if e := appendNativeRoute(plan, "edge", points, line, "112233", connection); e != nil {
		t.Fatal(e)
	}
	if len(plan.Items) != 2*len(points)-3 {
		t.Fatal("incomplete native route namespace")
	}
	for i, item := range plan.Items {
		shape := item.Shape
		if shape == nil || shape.Route != nil || len(shape.Props.Points) != 0 {
			t.Fatal("custom geometry remains")
		}
		if i < len(points)-2 {
			guide := fmt.Sprintf("edge.waypoint-%03d", i+1)
			if shape.Record.ID != guide || shape.Type != pptx.ShapeTypeRect || shape.Props.Fill.Type != "none" || shape.Props.Line.Type != "none" {
				t.Fatal("guide is not transparent/owned", shape)
			}
			x, y := shape.Props.X.Val*72+shape.Props.W.Val*36, shape.Props.Y.Val*72
			if math.Hypot(x-points[i+1][0], y-points[i+1][1]) > 1e-8 {
				t.Fatal("guide port moved from authored point")
			}
			continue
		}
		segment := i - (len(points) - 2)
		if shape.Record.ID != fmt.Sprintf("edge.segment-%03d", segment+1) || shape.Type != pptx.ShapeTypeLine || shape.Connection == nil {
			t.Fatal("segment identity/attachment lost")
		}
		want := connection
		if segment > 0 {
			want.Begin = pptx.ConnectorEndpoint{ObjectName: fmt.Sprintf("edge.waypoint-%03d", segment), Site: 0}
		}
		if segment < len(points)-2 {
			want.End = pptx.ConnectorEndpoint{ObjectName: fmt.Sprintf("edge.waypoint-%03d", segment+1), Site: 0}
		}
		if *shape.Connection != want {
			t.Fatal("incorrect segment topology", shape.Connection, want)
		}
		if (shape.Props.Line.BeginArrowType != "") != (segment == 0) || (shape.Props.Line.EndArrowType != "") != (segment == len(points)-2) {
			t.Fatal("intermediate arrow changes logical relationship")
		}
	}
	if line.BeginArrowType != "triangle" || line.EndArrowType != "triangle" {
		t.Fatal("caller style mutated")
	}
}
