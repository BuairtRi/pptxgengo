package deckproject

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// These are advisory centerline/allocation measurements, not exact painted
// collisions. No route, source property, or build acceptance is changed.
type DiagramRoutingInspection struct {
	Basis              string                   `json:"basis"`
	MinimumClearancePT float64                  `json:"minimum_clearance_pt"`
	Obstacles          []DiagramRoutingObstacle `json:"obstacles"`
	Diagnostics        []DiagramRouteDiagnostic `json:"diagnostics"`
}
type DiagramRoutingObstacle struct {
	Name       string        `json:"native_object"`
	Kind       string        `json:"kind"`
	Bounds     wmdesign.Rect `json:"world_allocation_bounds"`
	SourceNode string        `json:"source_node,omitempty"`
}
type DiagramRouteDiagnostic struct {
	Kind         string      `json:"kind"`
	Connection   string      `json:"connection"`
	Other        string      `json:"other"`
	Segment      int         `json:"segment"`
	OtherSegment *int        `json:"other_segment,omitempty"`
	ClearancePT  *float64    `json:"clearance_pt,omitempty"`
	Intersection *[2]float64 `json:"intersection_pt,omitempty"`
	Message      string      `json:"message"`
}

const routingTolerance = .02 // Native coordinates are rounded to EMU.

func addDiagramRoutingDiagnostics(out *DiagramInspection, objects map[string]*geometryObject, rules map[string]wmdesign.DiagramContainment) {
	if len(out.Connections) == 0 {
		return
	}
	world := map[string]wmdesign.Rect{}
	for _, o := range out.FinalNative {
		world[o.Name] = o.WorldBounds
	}
	blocks := map[string]string{} // Explicit supported port objects and owners.
	for _, port := range out.Ports {
		blocks[port.NativeObject] = port.Node
	}
	obstacles := []DiagramRoutingObstacle{}
	for name := range blocks {
		obstacles = append(obstacles, DiagramRoutingObstacle{name, "block", world[name], blocks[name]})
	}
	// Keep separate labels even inside blocks: a declared enclosing block can
	// be intentional geometry while its heading remains a real obstruction.
	// An editable-block's text shares the very same native allocation, so it
	// is represented once as a block rather than inventing a smaller ink box.
	for name, object := range objects {
		if object.node == nil || !routingHasText(lineageChild(object.node, lineagePML, "txBody")) {
			continue
		}
		if _, sameAllocation := blocks[name]; !sameAllocation {
			obstacles = append(obstacles, DiagramRoutingObstacle{name, "text", world[name], ""})
		}
	}
	sort.Slice(obstacles, func(i, j int) bool { return obstacles[i].Name < obstacles[j].Name })
	out.Routing = inspectDiagramRoutes(out.Connections, obstacles, rules, objects)
	for _, d := range out.Routing.Diagnostics {
		out.Warnings = append(out.Warnings, d.Message)
	}
}

func routingHasText(n *xmlNode) bool {
	if n == nil {
		return false
	}
	if n.Name.Space == drawingML && n.Name.Local == "t" && strings.TrimSpace(n.Text) != "" {
		return true
	}
	for _, c := range n.Children {
		if routingHasText(c) {
			return true
		}
	}
	return false
}
func routingNativeWithin(objects map[string]*geometryObject, name, owner string) bool {
	for name != "" {
		if name == owner {
			return true
		}
		o := objects[name]
		if o == nil {
			return false
		}
		name = o.geometry.Parent
	}
	return false
}
func routingLogicalWithin(rules map[string]wmdesign.DiagramContainment, member, container string) bool {
	for member != "" {
		if member == container {
			return true
		}
		r, ok := rules[member]
		if !ok {
			return false
		}
		member = r.Container
	}
	return false
}
func routingExcludedObstacle(c DiagramConnection, o DiagramRoutingObstacle, rules map[string]wmdesign.DiagramContainment, objects map[string]*geometryObject) bool {
	for _, p := range []DiagramPort{c.From, c.To} {
		if o.Name == p.NativeObject || routingNativeWithin(objects, o.Name, p.Node) {
			return true
		}
		// A containing block is intentional enclosing geometry. Its separate
		// heading text remains an obstacle; containment never excuses labels.
		if o.Kind == "block" {
			for _, owner := range []string{o.Name, o.SourceNode} {
				if owner != "" && (routingLogicalWithin(rules, p.NativeObject, owner) || routingLogicalWithin(rules, p.Node, owner) || routingLogicalWithin(rules, c.Node, owner)) {
					return true
				}
			}
		}
	}
	return false
}

func inspectDiagramRoutes(connections []DiagramConnection, obstacles []DiagramRoutingObstacle, rules map[string]wmdesign.DiagramContainment, objects map[string]*geometryObject) *DiagramRoutingInspection {
	out := &DiagramRoutingInspection{Basis: "final_native_centerlines_and_world_allocation_envelopes_not_ink", MinimumClearancePT: 6, Obstacles: obstacles, Diagnostics: []DiagramRouteDiagnostic{}}
	for _, c := range connections {
		for _, o := range obstacles {
			if routingExcludedObstacle(c, o, rules, objects) {
				continue
			}
			distance, segment := math.Inf(1), -1
			for i := 1; i < len(c.Points); i++ {
				if routePointDistance(c.Points[i-1], c.Points[i]) < 1e-9 {
					continue
				}
				if d := routeSegmentRectDistance(c.Points[i-1], c.Points[i], o.Bounds); d < distance {
					distance, segment = d, i-1
				}
			}
			if distance >= out.MinimumClearancePT-routingTolerance {
				continue
			}
			kind := "obstacle_clearance"
			if distance <= routingTolerance {
				kind = "obstacle_intersection"
			}
			out.Diagnostics = append(out.Diagnostics, DiagramRouteDiagnostic{Kind: kind, Connection: c.Node, Other: o.Name, Segment: segment, ClearancePT: &distance, Message: fmt.Sprintf("Route %s has %.3f pt centerline clearance from %s allocation %s (advisory minimum %.1f pt); review the diagram.", c.Node, distance, o.Kind, o.Name, out.MinimumClearancePT)})
		}
	}
	for a := 0; a < len(connections); a++ {
		for b := a + 1; b < len(connections); b++ {
			x, y := connections[a], connections[b]
			seen := map[string]bool{}
			for i := 1; i < len(x.Points); i++ {
				for j := 1; j < len(y.Points); j++ {
					kind, point := routeSegmentIntersection(x.Points[i-1], x.Points[i], y.Points[j-1], y.Points[j])
					if kind == "" || (kind == "crossing" && routingSharedJoin(x, y, point)) {
						continue
					}
					key := fmt.Sprintf("%s:%.3f:%.3f", kind, point[0], point[1])
					if seen[key] {
						continue
					}
					seen[key] = true
					otherSegment := j - 1
					out.Diagnostics = append(out.Diagnostics, DiagramRouteDiagnostic{Kind: "route_" + kind, Connection: x.Node, Other: y.Node, Segment: i - 1, OtherSegment: &otherSegment, Intersection: &point, Message: fmt.Sprintf("Routes %s and %s have a centerline %s near (%.3f, %.3f) pt; review relationship readability.", x.Node, y.Node, kind, point[0], point[1])})
				}
			}
		}
	}
	return out
}
func routingSharedJoin(a, b DiagramConnection, point [2]float64) bool {
	for _, x := range []DiagramPort{a.From, a.To} {
		for _, y := range []DiagramPort{b.From, b.To} {
			if x.NativeObject != "" && x.Site != "" && x.NativeObject == y.NativeObject && x.Site == y.Site && routePointDistance([2]float64{x.X, x.Y}, point) <= routingTolerance && routePointDistance([2]float64{y.X, y.Y}, point) <= routingTolerance {
				return true
			}
		}
	}
	return false
}

func routePointDistance(a, b [2]float64) float64 { return math.Hypot(a[0]-b[0], a[1]-b[1]) }
func routePointSegmentDistance(p, a, b [2]float64) float64 {
	dx, dy := b[0]-a[0], b[1]-a[1]
	n := dx*dx + dy*dy
	if n == 0 {
		return routePointDistance(p, a)
	}
	t := math.Max(0, math.Min(1, ((p[0]-a[0])*dx+(p[1]-a[1])*dy)/n))
	return routePointDistance(p, [2]float64{a[0] + t*dx, a[1] + t*dy})
}
func routeSegmentIntersection(a, b, c, d [2]float64) (string, [2]float64) {
	zero := [2]float64{}
	u, v := [2]float64{b[0] - a[0], b[1] - a[1]}, [2]float64{d[0] - c[0], d[1] - c[1]}
	uu, vv := u[0]*u[0]+u[1]*u[1], v[0]*v[0]+v[1]*v[1]
	if uu < 1e-18 || vv < 1e-18 {
		return "", zero
	}
	cross := func(x, y [2]float64) float64 { return x[0]*y[1] - x[1]*y[0] }
	w := [2]float64{c[0] - a[0], c[1] - a[1]}
	den := cross(u, v)
	if math.Abs(den) > 1e-10*math.Sqrt(uu*vv) {
		t, s := cross(w, v)/den, cross(w, u)/den
		if t < -1e-9 || t > 1+1e-9 || s < -1e-9 || s > 1+1e-9 {
			return "", zero
		}
		return "crossing", [2]float64{a[0] + t*u[0], a[1] + t*u[1]}
	}
	if math.Abs(cross(w, u))/math.Sqrt(uu) > 1e-7 {
		return "", zero
	}
	t0 := (w[0]*u[0] + w[1]*u[1]) / uu
	t1 := t0 + (v[0]*u[0]+v[1]*u[1])/uu
	lo, hi := math.Max(0, math.Min(t0, t1)), math.Min(1, math.Max(t0, t1))
	if hi < lo-1e-9 {
		return "", zero
	}
	point := [2]float64{a[0] + lo*u[0], a[1] + lo*u[1]}
	if (hi-lo)*math.Sqrt(uu) > routingTolerance {
		return "overlap", point
	}
	return "crossing", point
}
func routeSegmentRectDistance(a, b [2]float64, r wmdesign.Rect) float64 {
	inside := func(p [2]float64) bool { return p[0] >= r.X && p[0] <= r.X+r.W && p[1] >= r.Y && p[1] <= r.Y+r.H }
	if inside(a) || inside(b) {
		return 0
	}
	corners := [][2]float64{{r.X, r.Y}, {r.X + r.W, r.Y}, {r.X + r.W, r.Y + r.H}, {r.X, r.Y + r.H}}
	best := math.Inf(1)
	for i, c := range corners {
		d := corners[(i+1)%4]
		if kind, _ := routeSegmentIntersection(a, b, c, d); kind != "" {
			return 0
		}
		best = math.Min(best, routePointSegmentDistance(c, a, b))
		best = math.Min(best, routePointSegmentDistance(a, c, d))
		best = math.Min(best, routePointSegmentDistance(b, c, d))
	}
	return best
}
