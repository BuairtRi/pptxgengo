package deckproject

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Only complete supported preset metadata is eligible for geometry-only adoption.
// Formula expressions, additional guides/children and unknown presets stay manual.
func readConnectorRoute(n *xmlNode) (*wmdesign.NativeConnectorRoute, error) {
	if n == nil || n.Name.Space != lineagePML || n.Name.Local != "cxnSp" {
		return nil, fmt.Errorf("not a native connector")
	}
	props := lineageChild(n, lineagePML, "spPr")
	shape := lineageChild(props, drawingML, "prstGeom")
	count := 0
	if props != nil {
		for _, child := range props.Children {
			if child.Name.Space == drawingML && (child.Name.Local == "prstGeom" || child.Name.Local == "custGeom") {
				count++
			}
		}
	}
	if count != 1 {
		return nil, fmt.Errorf("ambiguous connector geometry")
	}
	if shape == nil {
		return nil, fmt.Errorf("custom connector geometry is unsupported")
	}
	preset := lineageAttr(shape, "", "prst")
	if preset != "line" && preset != "bentConnector3" {
		return nil, fmt.Errorf("unsupported connector preset")
	}
	allowedAttrs := func(n *xmlNode, allowed map[string]bool) bool {
		if strings.TrimSpace(n.Text) != "" {
			return false
		}
		for _, a := range n.Attrs {
			if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
				continue
			}
			if a.Name.Space != "" || !allowed[a.Name.Local] {
				return false
			}
		}
		return true
	}
	if !allowedAttrs(shape, map[string]bool{"prst": true}) || len(shape.Children) > 1 {
		return nil, fmt.Errorf("unsupported connector geometry metadata")
	}
	value := 50000
	if len(shape.Children) == 1 {
		av := shape.Children[0]
		if av.Name.Space != drawingML || av.Name.Local != "avLst" || !allowedAttrs(av, nil) || len(av.Children) > 1 {
			return nil, fmt.Errorf("unsupported connector adjustment list")
		}
		for _, gd := range av.Children {
			if preset == "line" || gd.Name.Space != drawingML || gd.Name.Local != "gd" || len(gd.Children) != 0 || !allowedAttrs(gd, map[string]bool{"name": true, "fmla": true}) || lineageAttr(gd, "", "name") != "adj1" {
				return nil, fmt.Errorf("unsupported connector guide")
			}
			parts := strings.Fields(lineageAttr(gd, "", "fmla"))
			if len(parts) != 2 || parts[0] != "val" {
				return nil, fmt.Errorf("connector bend must be a literal guide")
			}
			number, e := strconv.ParseInt(parts[1], 10, 32)
			if e != nil || number < -2147483647 {
				return nil, fmt.Errorf("invalid connector bend guide")
			}
			value = int(number)
		}
	}
	if preset == "line" {
		return nil, nil
	}
	return &wmdesign.NativeConnectorRoute{Preset: preset, Adjustment: value}, nil
}
func connectorRouteXML(route *wmdesign.NativeConnectorRoute) string {
	if route == nil {
		return fmt.Sprintf(`<a:prstGeom xmlns:a="%s" prst="line"><a:avLst/></a:prstGeom>`, drawingML)
	}
	return fmt.Sprintf(`<a:prstGeom xmlns:a="%s" prst="bentConnector3"><a:avLst><a:gd name="adj1" fmla="val %d"/></a:avLst></a:prstGeom>`, drawingML, route.Adjustment)
}
func nativeConnectorPoints(g NativeGeometry) [][2]float64 {
	a, z := [2]float64{g.X, g.Y}, [2]float64{g.X + g.W, g.Y + g.H}
	if g.Route == nil {
		return [][2]float64{a, z}
	}
	bend := g.X + g.W*float64(g.Route.Adjustment)/100000
	return [][2]float64{a, {bend, g.Y}, {bend, g.Y + g.H}, z}
}
func nativeGeometryAllocation(g NativeGeometry) wmdesign.Rect {
	r := wmdesign.Rect{X: g.X, Y: g.Y, W: g.W, H: g.H}
	if g.Route != nil {
		x := g.X + g.W*float64(g.Route.Adjustment)/100000
		if x < r.X {
			r.W += r.X - x
			r.X = x
		} else if x > r.X+r.W {
			r.W = x - r.X
		}
	}
	return r
}
