package deckproject

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Only complete supported preset metadata is eligible for geometry-only adoption.
// Formula expressions, additional guides/children and unknown presets stay manual.
func nativePresetGuideCount(preset string) int {
	switch preset {
	case "line":
		return 0
	case "bentConnector2":
		return 0
	case "bentConnector3":
		return 1
	case "bentConnector4":
		return 2
	case "bentConnector5":
		return 3
	}
	return -1
}
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
	if count != 1 || shape == nil {
		return nil, fmt.Errorf("custom/ambiguous native connector geometry remains manual; generated routes use presets or tagged segments")
	}
	preset := lineageAttr(shape, "", "prst")
	guides := nativePresetGuideCount(preset)
	if guides < 0 || !connectorAttrsOnly(shape, "prst") || len(shape.Children) > 1 {
		return nil, fmt.Errorf("unsupported connector preset metadata")
	}
	values := []int{0, 0, 0}
	for i := 0; i < guides; i++ {
		values[i] = 50000
	}
	seen := map[string]bool{}
	if len(shape.Children) == 1 {
		av := shape.Children[0]
		if av.Name.Space != drawingML || av.Name.Local != "avLst" || !connectorAttrsOnly(av) || len(av.Children) > guides {
			return nil, fmt.Errorf("unsupported connector adjustment list")
		}
		for _, gd := range av.Children {
			name := lineageAttr(gd, "", "name")
			index := -1
			for i := 0; i < guides; i++ {
				if name == fmt.Sprintf("adj%d", i+1) {
					index = i
				}
			}
			if index < 0 || seen[name] || gd.Name.Space != drawingML || gd.Name.Local != "gd" || len(gd.Children) != 0 || !connectorAttrsOnly(gd, "name", "fmla") {
				return nil, fmt.Errorf("unsupported/duplicate connector guide")
			}
			seen[name] = true
			parts := strings.Fields(lineageAttr(gd, "", "fmla"))
			if len(parts) != 2 || parts[0] != "val" {
				return nil, fmt.Errorf("connector bend must be literal guide")
			}
			value, e := strconv.ParseInt(parts[1], 10, 32)
			if e != nil || value < -2147483647 {
				return nil, fmt.Errorf("invalid connector guide")
			}
			values[index] = int(value)
		}
	}
	if preset == "line" {
		return nil, nil
	}
	return &wmdesign.NativeConnectorRoute{Preset: preset, Adjustment: values[0], Adjustment2: values[1], Adjustment3: values[2]}, nil
}
func connectorRouteXML(route *wmdesign.NativeConnectorRoute) string {
	preset := "line"
	if route != nil {
		preset = route.Preset
	}
	count := nativePresetGuideCount(preset)
	if count < 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<a:prstGeom xmlns:a="%s" prst="%s"><a:avLst>`, drawingML, preset)
	if route != nil {
		for i, value := range []int{route.Adjustment, route.Adjustment2, route.Adjustment3} {
			if i < count {
				fmt.Fprintf(&b, `<a:gd name="adj%d" fmla="val %d"/>`, i+1, value)
			}
		}
	}
	b.WriteString(`</a:avLst></a:prstGeom>`)
	return b.String()
}
func nativeConnectorPoints(g NativeGeometry) [][2]float64 {
	if g.Route != nil && g.Route.Preset == "polyline" {
		points := make([][2]float64, len(g.Route.Points))
		for i, p := range g.Route.Points {
			points[i] = [2]float64{g.X + g.W*p[0], g.Y + g.H*p[1]}
		}
		return points
	}
	a, z := [2]float64{g.X, g.Y}, [2]float64{g.X + g.W, g.Y + g.H}
	if g.Route == nil {
		return [][2]float64{a, z}
	}
	x1, y2, x3 := g.X+g.W*float64(g.Route.Adjustment)/100000, g.Y+g.H*float64(g.Route.Adjustment2)/100000, g.X+g.W*float64(g.Route.Adjustment3)/100000
	switch g.Route.Preset {
	case "bentConnector2":
		return [][2]float64{a, {z[0], a[1]}, z}
	case "bentConnector3":
		return [][2]float64{a, {x1, a[1]}, {x1, z[1]}, z}
	case "bentConnector4":
		return [][2]float64{a, {x1, a[1]}, {x1, y2}, {z[0], y2}, z}
	case "bentConnector5":
		return [][2]float64{a, {x1, a[1]}, {x1, y2}, {x3, y2}, {x3, z[1]}, z}
	}
	return nil
}
func nativeGeometryAllocation(g NativeGeometry) wmdesign.Rect {
	r := wmdesign.Rect{X: g.X, Y: g.Y, W: g.W, H: g.H}
	for _, p := range nativeConnectorPoints(g) {
		right, bottom := math.Max(r.X+r.W, p[0]), math.Max(r.Y+r.H, p[1])
		r.X, r.Y = math.Min(r.X, p[0]), math.Min(r.Y, p[1])
		r.W, r.H = right-r.X, bottom-r.Y
	}
	return r
}

func connectorAttrsOnly(n *xmlNode, allowed ...string) bool {
	if n == nil || strings.TrimSpace(n.Text) != "" {
		return false
	}
	keys := map[string]bool{}
	for _, k := range allowed {
		keys[k] = true
	}
	seen := map[string]bool{}
	for _, a := range n.Attrs {
		if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
			continue
		}
		if a.Name.Space != "" || !keys[a.Name.Local] || seen[a.Name.Local] {
			return false
		}
		seen[a.Name.Local] = true
	}
	return true
}
