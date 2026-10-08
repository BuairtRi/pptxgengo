package deckproject

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// NativeGeometry is a lossless DrawingML transform in points. Group child
// coordinates remain in their own space; changing a parent does not change its
// children's source coordinates. Kind and parent protect template upgrades
// against accidentally applying an override to a different object.
type NativeGeometry = wmdesign.NativeGeometry
type NativeGeometryObservation struct {
	Name        string         `json:"name"`
	Geometry    NativeGeometry `json:"geometry"`
	WorldBounds wmdesign.Rect  `json:"world_bounds"`
}

func nativeObjectKind(n *xmlNode) bool {
	if n.Name.Space != lineagePML {
		return false
	}
	switch n.Name.Local {
	case "sp", "pic", "graphicFrame", "grpSp", "cxnSp":
		return true
	}
	return false
}
func geometryXfrm(n *xmlNode) *xmlNode {
	space, prop := lineagePML, "spPr"
	switch n.Name.Local {
	case "grpSp":
		prop = "grpSpPr"
	case "graphicFrame":
		return lineageChild(n, lineagePML, "xfrm")
	}
	if p := lineageChild(n, space, prop); p != nil {
		return lineageChild(p, drawingML, "xfrm")
	}
	return nil
}
func geometryName(n *xmlNode) string {
	for _, c := range n.Children {
		if c.Name.Space == lineagePML {
			for _, id := range c.Children {
				if id.Name.Space == lineagePML && id.Name.Local == "cNvPr" {
					return lineageAttr(id, "", "name")
				}
			}
		}
	}
	return ""
}
func geometryNumber(n *xmlNode, key string) (float64, error) {
	if n == nil {
		return 0, fmt.Errorf("geometry missing %s", key)
	}
	v, e := strconv.ParseInt(lineageAttr(n, "", key), 10, 64)
	if e != nil || v < -1e12 || v > 1e12 {
		return 0, fmt.Errorf("invalid geometry %s", key)
	}
	return float64(v) / 12700, nil
}
func readNativeGeometry(n *xmlNode, parent string) (NativeGeometry, error) {
	g := NativeGeometry{Kind: n.Name.Local, Parent: parent}
	x := geometryXfrm(n)
	if x == nil {
		return g, fmt.Errorf("native transform unavailable")
	}
	var e error
	if g.X, e = geometryNumber(lineageChild(x, drawingML, "off"), "x"); e != nil {
		return g, e
	}
	if g.Y, e = geometryNumber(lineageChild(x, drawingML, "off"), "y"); e != nil {
		return g, e
	}
	if g.W, e = geometryNumber(lineageChild(x, drawingML, "ext"), "cx"); e != nil {
		return g, e
	}
	if g.H, e = geometryNumber(lineageChild(x, drawingML, "ext"), "cy"); e != nil {
		return g, e
	}
	if rot := lineageAttr(x, "", "rot"); rot != "" {
		v, err := strconv.ParseInt(rot, 10, 64)
		if err != nil {
			return g, err
		}
		g.Rotation = float64(v) / 60000
	}
	boolean := func(k string) (bool, error) {
		v := lineageAttr(x, "", k)
		switch v {
		case "", "0", "false":
			return false, nil
		case "1", "true":
			return true, nil
		}
		return false, fmt.Errorf("invalid native %s", k)
	}
	if g.FlipH, e = boolean("flipH"); e != nil {
		return g, e
	}
	if g.FlipV, e = boolean("flipV"); e != nil {
		return g, e
	}
	if n.Name.Local == "grpSp" {
		c := &wmdesign.Rect{}
		g.Child = c
		if c.X, e = geometryNumber(lineageChild(x, drawingML, "chOff"), "x"); e != nil {
			return g, e
		}
		if c.Y, e = geometryNumber(lineageChild(x, drawingML, "chOff"), "y"); e != nil {
			return g, e
		}
		if c.W, e = geometryNumber(lineageChild(x, drawingML, "chExt"), "cx"); e != nil {
			return g, e
		}
		if c.H, e = geometryNumber(lineageChild(x, drawingML, "chExt"), "cy"); e != nil {
			return g, e
		}
	}
	if n.Name.Local == "cxnSp" {
		route, err := readConnectorRoute(n)
		if err == nil {
			g.Route = route
		}
	}
	return g, validateNativeGeometry(g)
}
func validateNativeGeometry(g NativeGeometry) error {
	if g.Route != nil && (g.Kind != "cxnSp" || g.Route.Preset != "bentConnector3" || g.Route.Adjustment < -2147483647 || g.Route.Adjustment > 2147483647) {
		return fmt.Errorf("unsupported native connector route")
	}
	values := []float64{g.X, g.Y, g.W, g.H, g.Rotation}
	if g.Child != nil {
		values = append(values, g.Child.X, g.Child.Y, g.Child.W, g.Child.H)
	}
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e7 {
			return fmt.Errorf("invalid native geometry range")
		}
	}
	if g.W < 0 || g.H < 0 || (g.Kind == "grpSp" && (g.W <= 0 || g.H <= 0)) || (g.Kind == "grpSp") != (g.Child != nil) || (g.Child != nil && (g.Child.W <= 0 || g.Child.H <= 0)) {
		return fmt.Errorf("invalid native transform extent")
	}
	if g.Kind == "graphicFrame" && (g.Rotation != 0 || g.FlipH || g.FlipV) {
		return fmt.Errorf("graphic-frame transform supports placement only")
	}
	switch g.Kind {
	case "sp", "pic", "graphicFrame", "grpSp", "cxnSp":
	default:
		return fmt.Errorf("unsupported geometry kind")
	}
	return nil
}
func geometryXML(g NativeGeometry, space string) string {
	emu := func(v float64) int64 { return int64(math.Round(v * 12700)) }
	s := fmt.Sprintf(`<x:xfrm xmlns:x="%s" rot="%d" flipH="%t" flipV="%t"><a:off xmlns:a="%s" x="%d" y="%d"/><a:ext xmlns:a="%s" cx="%d" cy="%d"/>`, space, int64(math.Round(g.Rotation*60000)), g.FlipH, g.FlipV, drawingML, emu(g.X), emu(g.Y), drawingML, emu(g.W), emu(g.H))
	if g.Kind == "graphicFrame" {
		s = fmt.Sprintf(`<x:xfrm xmlns:x="%s"><a:off xmlns:a="%s" x="%d" y="%d"/><a:ext xmlns:a="%s" cx="%d" cy="%d"/>`, space, drawingML, emu(g.X), emu(g.Y), drawingML, emu(g.W), emu(g.H))
	}
	if c := g.Child; c != nil {
		s += fmt.Sprintf(`<a:chOff xmlns:a="%s" x="%d" y="%d"/><a:chExt xmlns:a="%s" cx="%d" cy="%d"/>`, drawingML, emu(c.X), emu(c.Y), drawingML, emu(c.W), emu(c.H))
	}
	return s + `</x:xfrm>`
}

type geometryObject struct {
	name     string
	node     *xmlNode
	span, xf *lineageSpan
	geometry NativeGeometry
	children []*geometryObject
}

func geometryInventory(raw []byte) (map[string]*geometryObject, map[string][]string, error) {
	tree, e := readLineageXML(raw)
	if e != nil {
		return nil, nil, e
	}
	spans, e := lineageSpans(raw)
	if e != nil {
		return nil, nil, e
	}
	objects := map[string]*geometryObject{}
	orders := map[string][]string{}
	var walk func(*xmlNode, *lineageSpan, string) error
	walk = func(n *xmlNode, s *lineageSpan, parent string) error {
		if nativeObjectKind(n) {
			name := geometryName(n)
			if name == "" || objects[name] != nil {
				return fmt.Errorf("native geometry requires unique recorded names")
			}
			g, e := readNativeGeometry(n, parent)
			if e != nil {
				return fmt.Errorf("%s: %w", name, e)
			}
			obj := &geometryObject{name: name, node: n, span: s, geometry: g}
			objects[name] = obj
			orders[parent] = append(orders[parent], name)
			var find func(*lineageSpan)
			target := geometryXfrm(n)
			find = func(t *lineageSpan) {
				if t.node.Name == target.Name && t.node.Name.Local == "xfrm" && obj.xf == nil {
					obj.xf = t
					return
				}
				for _, c := range t.children {
					find(c)
				}
			}
			find(s)
			parent = name
		}
		if len(n.Children) != len(s.children) {
			return fmt.Errorf("native geometry XML inventory differs")
		}
		for i, c := range n.Children {
			if e := walk(c, s.children[i], parent); e != nil {
				return e
			}
		}
		return nil
	}
	if e = walk(tree, spans, ""); e != nil {
		return nil, nil, e
	}
	return objects, orders, nil
}

// matrix maps source coordinates through nested group scales, flips and
// rotations. World envelopes come from transformed corners, never sums of
// parent/child offsets.
type geometryMatrix [6]float64

func (a geometryMatrix) mul(b geometryMatrix) geometryMatrix {
	return geometryMatrix{a[0]*b[0] + a[2]*b[1], a[1]*b[0] + a[3]*b[1], a[0]*b[2] + a[2]*b[3], a[1]*b[2] + a[3]*b[3], a[0]*b[4] + a[2]*b[5] + a[4], a[1]*b[4] + a[3]*b[5] + a[5]}
}
func geometryIdentity() geometryMatrix              { return geometryMatrix{1, 0, 0, 1, 0, 0} }
func geometryTranslate(x, y float64) geometryMatrix { return geometryMatrix{1, 0, 0, 1, x, y} }
func geometryOriented(g NativeGeometry) geometryMatrix {
	cx, cy := g.X+g.W/2, g.Y+g.H/2
	t := g.Rotation * math.Pi / 180
	sx, sy := 1., 1.
	if g.FlipH {
		sx = -1
	}
	if g.FlipV {
		sy = -1
	}
	return geometryTranslate(cx, cy).mul(geometryMatrix{math.Cos(t) * sx, math.Sin(t) * sx, -math.Sin(t) * sy, math.Cos(t) * sy, 0, 0}).mul(geometryTranslate(-cx, -cy))
}
func (m geometryMatrix) bounds(r wmdesign.Rect) wmdesign.Rect {
	x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range [][2]float64{{r.X, r.Y}, {r.X + r.W, r.Y}, {r.X, r.Y + r.H}, {r.X + r.W, r.Y + r.H}} {
		x, y := m[0]*p[0]+m[2]*p[1]+m[4], m[1]*p[0]+m[3]*p[1]+m[5]
		x0 = math.Min(x0, x)
		y0 = math.Min(y0, y)
		x1 = math.Max(x1, x)
		y1 = math.Max(y1, y)
	}
	return wmdesign.Rect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}
}
func geometryWorld(objects map[string]*geometryObject) (map[string]wmdesign.Rect, error) {
	matrices := map[string]geometryMatrix{"": geometryIdentity()}
	visiting := map[string]bool{}
	bounds := map[string]wmdesign.Rect{}
	var resolve func(string) (geometryMatrix, error)
	resolve = func(name string) (geometryMatrix, error) {
		if m, ok := matrices[name]; ok {
			return m, nil
		}
		if visiting[name] {
			return geometryMatrix{}, fmt.Errorf("native parent cycle")
		}
		o := objects[name]
		if o == nil {
			return geometryMatrix{}, fmt.Errorf("native parent missing")
		}
		visiting[name] = true
		parent, e := resolve(o.geometry.Parent)
		if e != nil {
			return parent, e
		}
		g := o.geometry
		outer := parent.mul(geometryOriented(g))
		bounds[name] = outer.bounds(nativeGeometryAllocation(g))
		m := parent
		if c := g.Child; c != nil {
			m = outer.mul(geometryTranslate(g.X, g.Y)).mul(geometryMatrix{g.W / c.W, 0, 0, g.H / c.H, 0, 0}).mul(geometryTranslate(-c.X, -c.Y))
		}
		matrices[name] = m
		return m, nil
	}
	for name := range objects {
		if _, e := resolve(name); e != nil {
			return nil, e
		}
	}
	return bounds, nil
}
func applyNativeGeometry(raw []byte, doc wmdesign.Document, report *wmdesign.Report) ([]byte, error) {
	any := false
	for _, s := range doc.Slides {
		any = any || len(s.NativeGeometry) > 0 || len(s.NativeOrder) > 0 || len(s.DiagramContainment) > 0
	}
	if !any {
		return raw, nil
	}
	pkg, e := openLineagePackage(raw)
	if e != nil {
		return nil, e
	}
	changed := map[string][]byte{}
	for i, s := range doc.Slides {
		if len(s.NativeGeometry) == 0 && len(s.NativeOrder) == 0 && len(s.DiagramContainment) == 0 {
			continue
		}
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		data, e := pkg.read(part)
		if e != nil {
			return nil, e
		}
		objects, orders, e := geometryInventory(data)
		if e != nil {
			return nil, e
		}
		before, e := geometryWorld(objects)
		if e != nil {
			return nil, e
		}
		basises := map[string]string{}
		for name := range objects {
			basises[name] = nativeGeometryBasis(objects, name)
		}
		patches := []lineagePatch{}
		keys := []string{}
		for k := range s.NativeGeometry {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			g := s.NativeGeometry[name]
			o := objects[name]
			if e := validateNativeGeometry(g); e != nil {
				return nil, e
			}
			if o == nil || o.geometry.Kind != g.Kind || o.geometry.Parent != g.Parent {
				return nil, fmt.Errorf("native geometry binding changed: %s", name)
			}
			if !shaPattern.MatchString(g.SourceGeometrySHA256) || g.SourceGeometrySHA256 != basises[name] {
				return nil, fmt.Errorf("native geometry source basis changed: %s; reset/review layout or rebuild a baseline", name)
			}
			if o.xf == nil {
				return nil, fmt.Errorf("native transform missing: %s", name)
			}
			patches = append(patches, lineagePatch{o.xf.start, o.xf.end, geometryXML(g, o.xf.node.Name.Space)})
			if g.Kind == "cxnSp" {
				_, routeErr := readConnectorRoute(o.node)
				if routeErr == nil {
					var span *lineageSpan
					for _, props := range o.span.children {
						if props.node.Name.Space == lineagePML && props.node.Name.Local == "spPr" {
							for _, child := range props.children {
								if child.node.Name.Space == drawingML && child.node.Name.Local == "prstGeom" {
									span = child
								}
							}
						}
					}
					if span == nil {
						return nil, fmt.Errorf("connector preset span missing")
					}
					patches = append(patches, lineagePatch{span.start, span.end, connectorRouteXML(g.Route)})
				} else if g.Route != nil {
					return nil, routeErr
				}
			}
			o.geometry = g
		}
		after, e := geometryWorld(objects)
		if e != nil {
			return nil, e
		}
		for name, b := range after {
			if before[name] == b {
				continue
			}
			zone, _, err := chooseZone(before[name], report.Slides[i].Frame)
			if err != nil {
				return nil, fmt.Errorf("geometry %s has no writable frame zone", name)
			}
			var z wmdesign.Rect
			switch zone {
			case "short_body":
				z = report.Slides[i].Frame.ShortBody
			case "tall_body":
				z = report.Slides[i].Frame.TallBody
			case "body":
				z = report.Slides[i].Frame.Body
			case "rail":
				z = report.Slides[i].Frame.Rail
			}
			if b.X < z.X-.02 || b.Y < z.Y-.02 || b.X+b.W > z.X+z.W+.02 || b.Y+b.H > z.Y+z.H+.02 {
				return nil, fmt.Errorf("native geometry %s exceeds %s frame zone", name, zone)
			}
		}
		data, e = lineageApply(data, patches)
		if e != nil {
			return nil, e
		}
		// Apply paint order separately so its spans include rewritten transforms.
		for parent, desired := range s.NativeOrder {
			objects, orders, e = geometryInventory(data)
			if e != nil {
				return nil, e
			}
			old := orders[parent]
			if len(old) != len(desired) {
				return nil, fmt.Errorf("native paint order topology changed: %s", parent)
			}
			seen := map[string]bool{}
			for _, n := range desired {
				if objects[n] == nil || objects[n].geometry.Parent != parent || seen[n] {
					return nil, fmt.Errorf("invalid native paint order")
				}
				seen[n] = true
			}
			if len(old) == 0 {
				continue
			}
			for j := 1; j < len(old); j++ {
				if len(bytes.TrimSpace(data[objects[old[j-1]].span.end:objects[old[j]].span.start])) != 0 {
					return nil, fmt.Errorf("paint order has interleaved non-object XML; explicit review required")
				}
			}
			first, last := objects[old[0]].span.start, objects[old[len(old)-1]].span.end
			var text string
			for _, n := range desired {
				o := objects[n]
				text += string(data[o.span.start:o.span.end])
			}
			data, e = lineageApply(data, []lineagePatch{{first, last, text}})
			if e != nil {
				return nil, e
			}
		}
		objects, _, e = geometryInventory(data)
		if e != nil {
			return nil, e
		}
		if e = validateTransformedConnections(objects); e != nil {
			return nil, e
		}
		checks, _, e := checkDiagramContainment(objects, s.DiagramContainment)
		if e != nil {
			return nil, e
		}
		report.Slides[i].DiagramContainment = checks
		after, e = geometryWorld(objects)
		if e != nil {
			return nil, e
		}
		names := []string{}
		for name := range objects {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			report.Slides[i].NativeGeometry = append(report.Slides[i].NativeGeometry, wmdesign.NativeGeometryObservation{Name: name, WorldBounds: after[name]})
		}
		changed[part] = data
	}
	result, err := lineageRewrite(pkg, changed)
	if err == nil {
		report.PPTXSHA256 = digest(result)
		if report.MeasurementPolicy == nil {
			report.MeasurementPolicy = map[string]string{}
		}
		report.MeasurementPolicy["native_geometry"] = "Measured text is in authored coordinates; final native world bounds include group transforms. Desktop typography/visual qualification remains pending."
	}
	return result, err
}

func nativeCoordinateMatrix(objects map[string]*geometryObject, parent string) (geometryMatrix, error) {
	if parent == "" {
		return geometryIdentity(), nil
	}
	o := objects[parent]
	if o == nil || o.geometry.Child == nil {
		return geometryMatrix{}, fmt.Errorf("invalid native group parent")
	}
	base, e := nativeCoordinateMatrix(objects, o.geometry.Parent)
	if e != nil {
		return base, e
	}
	g := o.geometry
	c := g.Child
	return base.mul(geometryOriented(g)).mul(geometryTranslate(g.X, g.Y)).mul(geometryMatrix{g.W / c.W, 0, 0, g.H / c.H, 0, 0}).mul(geometryTranslate(-c.X, -c.Y)), nil
}
func nativeObjectPoint(objects map[string]*geometryObject, name string, x, y float64) ([2]float64, error) {
	o := objects[name]
	if o == nil {
		return [2]float64{}, fmt.Errorf("unknown native object")
	}
	m, e := nativeCoordinateMatrix(objects, o.geometry.Parent)
	if e != nil {
		return [2]float64{}, e
	}
	m = m.mul(geometryOriented(o.geometry))
	return [2]float64{m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]}, nil
}
func validateTransformedConnections(objects map[string]*geometryObject) error {
	ids := map[string]string{}
	for name, o := range objects {
		id, e := nativeObjectIdentity(o.node)
		if e != nil {
			return e
		}
		key := lineageAttr(id, "", "id")
		if ids[key] != "" {
			return fmt.Errorf("native connection target ID ambiguous")
		}
		ids[key] = name
	}
	for name, o := range objects {
		if o.geometry.Kind != "cxnSp" {
			continue
		}
		nv := lineageChild(o.node, lineagePML, "nvCxnSpPr")
		if nv == nil {
			return fmt.Errorf("native connector metadata missing")
		}
		connection := lineageChild(nv, lineagePML, "cNvCxnSpPr")
		if connection == nil {
			continue
		}
		a, z := lineageChild(connection, drawingML, "stCxn"), lineageChild(connection, drawingML, "endCxn")
		if a == nil && z == nil {
			continue
		}
		if a == nil || z == nil {
			return fmt.Errorf("native connector must retain both endpoint attachments")
		}
		props := lineageChild(o.node, lineagePML, "spPr")
		shape := lineageChild(props, drawingML, "prstGeom")
		if _, routeErr := readConnectorRoute(o.node); shape == nil || routeErr != nil {
			return fmt.Errorf("native connector route is unsupported for transformed geometry")
		}
		for i, end := range []*xmlNode{a, z} {
			targetName := ids[lineageAttr(end, "", "id")]
			target := objects[targetName]
			if target == nil {
				return fmt.Errorf("native connector %s target missing", name)
			}
			site, e := strconv.Atoi(lineageAttr(end, "", "idx"))
			if e != nil || site < 0 || site > 3 {
				return fmt.Errorf("unsupported native connection site")
			}
			g := target.geometry
			var x, y float64
			switch site {
			case 0:
				x, y = g.X+g.W/2, g.Y
			case 1:
				x, y = g.X, g.Y+g.H/2
			case 2:
				x, y = g.X+g.W/2, g.Y+g.H
			case 3:
				x, y = g.X+g.W, g.Y+g.H/2
			}
			expected, e := nativeObjectPoint(objects, targetName, x, y)
			if e != nil {
				return e
			}
			g = o.geometry
			x, y = g.X, g.Y
			if i == 1 {
				x += g.W
				y += g.H
			}
			actual, e := nativeObjectPoint(objects, name, x, y)
			if e != nil {
				return e
			}
			if math.Abs(actual[0]-expected[0]) > .02 || math.Abs(actual[1]-expected[1]) > .02 {
				return fmt.Errorf("native geometry disconnects %s from %s; review and adopt the incident connector geometry too", name, targetName)
			}
		}
	}
	return nil
}

func nativeGeometryBasis(objects map[string]*geometryObject, name string) string {
	records := []NativeGeometryObservation{}
	for key, o := range objects {
		owned := key == name
		for parent := o.geometry.Parent; !owned && parent != ""; {
			if parent == name {
				owned = true
				break
			}
			p := objects[parent]
			if p == nil {
				break
			}
			parent = p.geometry.Parent
		}
		if !owned {
			continue
		}
		g := o.geometry
		g.SourceGeometrySHA256 = ""
		records = append(records, NativeGeometryObservation{Name: key, Geometry: g})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	return digest(canonical(records))
}
