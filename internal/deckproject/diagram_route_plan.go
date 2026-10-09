package deckproject

import (
	"bytes"
	"container/heap"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

const DiagramRoutePatchSchema = "pptxgengo.diagram-route-patch.v1"

type DiagramRouteEndpoint struct {
	Node string `json:"node"`
	Site string `json:"site"`
}
type DiagramRoutePatch struct {
	Schema               string                `json:"schema"`
	ExpectedSourceSHA256 string                `json:"expected_source_sha256"`
	Actor                string                `json:"actor"`
	Reason               string                `json:"reason"`
	ID                   string                `json:"id"`
	From                 *DiagramRouteEndpoint `json:"from,omitempty"`
	To                   *DiagramRouteEndpoint `json:"to,omitempty"`
	Clearance            *float64              `json:"clearance_pt,omitempty"`
	Exclusions           map[string]string     `json:"exclusions,omitempty"`
	Reserved             []wmdesign.Rect       `json:"reserved,omitempty"`
	Label                *string               `json:"label,omitempty"`
	LabelPosition        *[2]float64           `json:"label_position,omitempty"`
	LabelWidth           *float64              `json:"label_width,omitempty"`
}
type DiagramRoutePlan struct {
	Schema       string                          `json:"schema"`
	Status       string                          `json:"status"`
	Message      string                          `json:"message"`
	SourceSHA256 string                          `json:"source_sha256"`
	Connection   string                          `json:"connection"`
	Basis        string                          `json:"basis"`
	Bounds       wmdesign.Rect                   `json:"world_route_allocation"`
	Policy       wmdesign.ConnectorRoutingPolicy `json:"policy"`
	Obstacles    []DiagramRoutingObstacle        `json:"obstacles"`
	Excluded     map[string]string               `json:"excluded"`
	Points       [][2]float64                    `json:"world_points,omitempty"`
	Candidate    *CompositionResult              `json:"candidate,omitempty"`
}

func DecodeDiagramRoutePatch(raw []byte, file string) (DiagramRoutePatch, error) {
	var out DiagramRoutePatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("route patch exceeds 1 MiB")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty route patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("route patch requires one document")
	}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = p.shapeType(v, reflect.TypeOf(out), ""); e != nil {
		return out, e
	}
	if e = strictInto(v, &out); e != nil {
		return out, e
	}
	if out.Schema != DiagramRoutePatchSchema || !validTeamDigest(out.ExpectedSourceSHA256) || !stableID.MatchString(out.ID) || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || strings.TrimSpace(out.Reason) == "" || len(out.Reason) > 4096 {
		return out, fmt.Errorf("route patch requires schema, source SHA256, stable ID, actor and reason")
	}
	if (out.From == nil) != (out.To == nil) {
		return out, fmt.Errorf("new route requires both endpoints")
	}
	if len(out.Exclusions) > 128 || len(out.Reserved) > 64 {
		return out, fmt.Errorf("route policy exceeds bounded obstacle budget")
	}
	if out.Clearance != nil && (math.IsNaN(*out.Clearance) || math.IsInf(*out.Clearance, 0) || *out.Clearance < 0 || *out.Clearance > 72) {
		return out, fmt.Errorf("clearance must be finite in 0..72 pt")
	}
	for name, reason := range out.Exclusions {
		if name == "" || len(name) > 512 || strings.TrimSpace(reason) == "" || len(reason) > 4096 {
			return out, fmt.Errorf("exclusions require exact object names and review reasons")
		}
	}
	for _, r := range out.Reserved {
		if r.W <= 0 || r.H <= 0 || math.IsNaN(r.X+r.Y+r.W+r.H) || math.IsInf(r.X+r.Y+r.W+r.H, 0) {
			return out, fmt.Errorf("reserved allocations require positive finite bounds")
		}
	}
	return out, nil
}
func routeRectIntersect(a, b wmdesign.Rect) wmdesign.Rect {
	x, y := math.Max(a.X, b.X), math.Max(a.Y, b.Y)
	return wmdesign.Rect{X: x, Y: y, W: math.Min(a.X+a.W, b.X+b.W) - x, H: math.Min(a.Y+a.H, b.Y+b.H) - y}
}
func routeRectContains(r wmdesign.Rect, p [2]float64) bool {
	return p[0] >= r.X-1e-8 && p[0] <= r.X+r.W+1e-8 && p[1] >= r.Y-1e-8 && p[1] <= r.Y+r.H+1e-8
}
func routeInflate(r wmdesign.Rect, c float64) wmdesign.Rect {
	return wmdesign.Rect{X: r.X - c, Y: r.Y - c, W: r.W + 2*c, H: r.H + 2*c}
}
func routeStub(p DiagramPort, c float64) [2]float64 {
	xy := [2]float64{p.X, p.Y}
	switch p.Site {
	case "left":
		xy[0] -= c
	case "right":
		xy[0] += c
	case "top":
		xy[1] -= c
	case "bottom":
		xy[1] += c
	}
	return xy
}
func routePointInsideObstacle(p [2]float64, rs []wmdesign.Rect) bool {
	for _, r := range rs {
		if p[0] > r.X+1e-7 && p[0] < r.X+r.W-1e-7 && p[1] > r.Y+1e-7 && p[1] < r.Y+r.H-1e-7 {
			return true
		}
	}
	return false
}
func routeSegmentClear(a, b [2]float64, rs []wmdesign.Rect) bool {
	for _, r := range rs {
		if a[0] == b[0] {
			if a[0] > r.X+1e-7 && a[0] < r.X+r.W-1e-7 && math.Max(a[1], b[1]) > r.Y+1e-7 && math.Min(a[1], b[1]) < r.Y+r.H-1e-7 {
				return false
			}
		}
		if a[1] == b[1] {
			if a[1] > r.Y+1e-7 && a[1] < r.Y+r.H-1e-7 && math.Max(a[0], b[0]) > r.X+1e-7 && math.Min(a[0], b[0]) < r.X+r.W-1e-7 {
				return false
			}
		}
	}
	return true
}

type routeQueueItem struct {
	state int
	cost  float64
}
type routeQueue []routeQueueItem

func (q routeQueue) Len() int { return len(q) }
func (q routeQueue) Less(i, j int) bool {
	if q[i].cost == q[j].cost {
		return q[i].state < q[j].state
	}
	return q[i].cost < q[j].cost
}
func (q routeQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *routeQueue) Push(x any)   { *q = append(*q, x.(routeQueueItem)) }
func (q *routeQueue) Pop() any     { x := (*q)[len(*q)-1]; *q = (*q)[:len(*q)-1]; return x }

// Deterministic orthogonal visibility grid with distance + bend cost. The grid
// contains obstacle boundaries, so detours are not limited to endpoint envelopes.
func orthogonalRoute(start, end [2]float64, bounds wmdesign.Rect, obstacles []wmdesign.Rect) ([][2]float64, error) {
	if bounds.W <= 0 || bounds.H <= 0 || !routeRectContains(bounds, start) || !routeRectContains(bounds, end) {
		return nil, fmt.Errorf("endpoint clearance stubs leave route allocation; choose another site/allocation")
	}
	if len(obstacles) > 256 {
		return nil, fmt.Errorf("route obstacle budget exceeds 256 allocations; explicitly review exclusions")
	}
	if routePointInsideObstacle(start, obstacles) || routePointInsideObstacle(end, obstacles) {
		return nil, fmt.Errorf("endpoint clearance corridor blocked; choose another site or move the blocker")
	}
	xs := []float64{bounds.X, bounds.X + bounds.W, start[0], end[0]}
	ys := []float64{bounds.Y, bounds.Y + bounds.H, start[1], end[1]}
	for _, r := range obstacles {
		for _, x := range []float64{r.X, r.X + r.W} {
			if x >= bounds.X && x <= bounds.X+bounds.W {
				xs = append(xs, x)
			}
		}
		for _, y := range []float64{r.Y, r.Y + r.H} {
			if y >= bounds.Y && y <= bounds.Y+bounds.H {
				ys = append(ys, y)
			}
		}
	}
	unique := func(a []float64) []float64 {
		sort.Float64s(a)
		out := a[:0]
		for _, v := range a {
			if len(out) == 0 || math.Abs(v-out[len(out)-1]) > 1e-7 {
				out = append(out, v)
			}
		}
		return out
	}
	xs, ys = unique(xs), unique(ys)
	w, h := len(xs), len(ys)
	if w*h > 160000 {
		return nil, fmt.Errorf("route visibility grid exceeds bounded search budget")
	}
	point := func(i int) [2]float64 { return [2]float64{xs[i%w], ys[i/w]} }
	index := func(p [2]float64) int { return sort.SearchFloat64s(ys, p[1])*w + sort.SearchFloat64s(xs, p[0]) }
	first, last := index(start), index(end)
	n := w * h * 2
	dist := make([]float64, n)
	prev := make([]int, n)
	for i := range dist {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}
	q := &routeQueue{}
	heap.Init(q)
	for d := 0; d < 2; d++ {
		s := first*2 + d
		dist[s] = 0
		heap.Push(q, routeQueueItem{s, 0})
	}
	target := -1
	for q.Len() > 0 {
		item := heap.Pop(q).(routeQueueItem)
		if item.cost > dist[item.state]+1e-8 {
			continue
		}
		cell, dir := item.state/2, item.state%2
		if cell == last {
			target = item.state
			break
		}
		x, y := cell%w, cell/w
		for _, delta := range [][3]int{{-1, 0, 0}, {1, 0, 0}, {0, -1, 1}, {0, 1, 1}} {
			nx, ny := x+delta[0], y+delta[1]
			if nx < 0 || nx >= w || ny < 0 || ny >= h {
				continue
			}
			next := ny*w + nx
			a, b := point(cell), point(next)
			if !routeSegmentClear(a, b, obstacles) {
				continue
			}
			state := next*2 + delta[2]
			cost := item.cost + routePointDistance(a, b)
			if dir != delta[2] {
				cost += 12
			}
			if cost < dist[state]-1e-8 {
				dist[state] = cost
				prev[state] = item.state
				heap.Push(q, routeQueueItem{state, cost})
			}
		}
	}
	if target < 0 {
		return nil, fmt.Errorf("no clear route inside allocation; move blockers, choose other sites, widen allocation or explicitly review exclusions")
	}
	path := [][2]float64{}
	for s := target; s >= 0; s = prev[s] {
		path = append(path, point(s/2))
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return simplifyOrthogonalRoute(path), nil
}
func simplifyOrthogonalRoute(path [][2]float64) [][2]float64 {
	out := [][2]float64{}
	for _, p := range path {
		if len(out) > 0 && routePointDistance(out[len(out)-1], p) < 1e-7 {
			continue
		}
		for len(out) >= 2 {
			a, b := out[len(out)-2], out[len(out)-1]
			if !(math.Abs(a[0]-b[0]) < 1e-7 && math.Abs(b[0]-p[0]) < 1e-7 || math.Abs(a[1]-b[1]) < 1e-7 && math.Abs(b[1]-p[1]) < 1e-7) {
				break
			}
			out = out[:len(out)-1]
		}
		out = append(out, p)
	}
	return out
}

func PlanDiagramRoute(p *Project, slideID string, patch DiagramRoutePatch, bundle, engine string, apply bool) (DiagramRoutePlan, error) {
	out := DiagramRoutePlan{Schema: "pptxgengo.diagram-route-plan.v1", Status: "no_route", Connection: patch.ID, SourceSHA256: p.SourceHash(), Basis: "native_world_allocation_centerlines_not_ink", Excluded: map[string]string{}}
	checked, e := DecodeDiagramRoutePatch(canonical(patch), "route-patch")
	if e != nil {
		return out, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return out, fmt.Errorf("stale route source SHA256")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return out, e
	}
	t = clone
	if len(p.Document.Slides[idx].NativeGeometry) > 0 || len(p.Document.Slides[idx].NativeOrder) > 0 {
		return out, fmt.Errorf("explicitly resolve/reset native layout before source rerouting")
	}
	inspect, e := InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		return out, e
	}
	list, i := findDiagramNode(&t.Nodes, patch.ID)
	var node Node
	if list == nil {
		if patch.From == nil {
			return out, fmt.Errorf("new route requires explicit from/to endpoints")
		}
		r := wmdesign.Rect{W: inspect.Frame.Body.W, H: inspect.Frame.Body.H}
		node = Node{ID: patch.ID, Kind: "component", Placement: &Placement{Zone: "body", Rect: &r}, Definition: &Reference{Scope: "shared", ID: "wmds/component/attached-connector"}, Arguments: map[string]any{"from": map[string]any{"node": patch.From.Node, "site": patch.From.Site}, "to": map[string]any{"node": patch.To.Node, "site": patch.To.Site}, "head": "end", "style": "solid"}}
	} else {
		node = (*list)[i]
		if patch.From != nil {
			return out, fmt.Errorf("reroute preserves endpoint relationships; patch them explicitly first")
		}
		if node.Definition == nil || node.Definition.Scope != "shared" || node.Definition.ID != "wmds/component/attached-connector" || teamBound(node.Arguments) {
			return out, fmt.Errorf("route requires a literal attached connector")
		}
	}
	if node.Placement == nil || node.Placement.Zone != "body" || node.Placement.Rect == nil {
		return out, fmt.Errorf("routing requires an explicit body rectangle allocation")
	}
	var endpoints struct{ From, To DiagramRouteEndpoint }
	if e = strictInto(map[string]any{"from": node.Arguments["from"], "to": node.Arguments["to"]}, &endpoints); e != nil {
		return out, e
	}
	ports := map[string]DiagramPort{}
	for _, port := range inspect.Ports {
		ports[port.Node+"/"+port.Site] = port
	}
	from, ok := ports[endpoints.From.Node+"/"+endpoints.From.Site]
	to, ok2 := ports[endpoints.To.Node+"/"+endpoints.To.Site]
	if !ok || !ok2 || from.Node == to.Node {
		return out, fmt.Errorf("routing requires distinct known node/site endpoints")
	}
	origin := [2]float64{inspect.Frame.Body.X + node.Placement.Rect.X, inspect.Frame.Body.Y + node.Placement.Rect.Y}
	out.Bounds = wmdesign.Rect{X: origin[0], Y: origin[1], W: node.Placement.Rect.W, H: node.Placement.Rect.H}
	out.Bounds = routeRectIntersect(out.Bounds, inspect.Frame.Body)
	c := 6.
	if patch.Clearance != nil {
		c = *patch.Clearance
	}
	out.Policy = wmdesign.ConnectorRoutingPolicy{Clearance: c, Exclusions: patch.Exclusions, Reserved: patch.Reserved}
	objects := map[string]*geometryObject{}
	world := map[string]NativeGeometryObservation{}
	for _, o := range inspect.FinalNative {
		objects[o.Name] = &geometryObject{geometry: o.Geometry}
		world[o.Name] = o
	}
	// Respect explicit enclosing allocation, including its padded axes. Rotated
	// containers use the existing exact fit guard; planning requires unrotated axes.
	rules := p.Document.Slides[idx].DiagramContainment
	for member := patch.ID; member != ""; {
		rule, exists := rules[member]
		if !exists {
			break
		}
		o := world[rule.Container]
		if o.Geometry.Rotation != 0 || o.Geometry.FlipH || o.Geometry.FlipV {
			return out, fmt.Errorf("reroute rotated container with explicit source waypoints and measured fit")
		}
		m, err := nativeOuterMatrix(objects, rule.Container)
		if err != nil {
			return out, err
		}
		if math.Abs(m[1])+math.Abs(m[2]) > 1e-8 || m[0] <= 0 || m[3] <= 0 {
			return out, fmt.Errorf("rotated or reflected ancestor needs explicit route waypoints")
		}
		g := o.Geometry
		b := m.bounds(wmdesign.Rect{X: g.X + rule.Padding.Left, Y: g.Y + rule.Padding.Top, W: g.W - rule.Padding.Left - rule.Padding.Right, H: g.H - rule.Padding.Top - rule.Padding.Bottom})
		out.Bounds = routeRectIntersect(out.Bounds, b)
		member = rule.Container
	}
	conn := DiagramConnection{Node: patch.ID, From: from, To: to}
	blocks := map[string]bool{}
	for _, port := range inspect.Ports {
		blocks[port.NativeObject] = true
	}
	known := map[string]bool{}
	obstacles := []wmdesign.Rect{}
	for _, o := range inspect.FinalNative {
		if o.Geometry.Kind == "grpSp" || o.Geometry.Kind == "cxnSp" || o.WorldBounds.W <= 0 || o.WorldBounds.H <= 0 {
			continue
		}
		// Only authored scene leaves in the body; frame/chrome is the allocation.
		owner := ""
		for _, n := range inspect.Nodes {
			if o.Name == n.ID || strings.HasPrefix(o.Name, n.ID+".") {
				owner = n.ID
				break
			}
		}
		if owner == "" {
			continue
		}
		known[o.Name] = true
		kind := "shape"
		if blocks[o.Name] {
			kind = "block"
		}
		if o.Geometry.Kind == "pic" {
			kind = "image"
		}
		ob := DiagramRoutingObstacle{Name: o.Name, Kind: kind, Bounds: o.WorldBounds, SourceNode: owner}
		if routingExcludedObstacle(conn, ob, rules, objects) || owner == patch.ID {
			out.Excluded[o.Name] = "endpoint-owned allocation or the connector's own label"
			continue
		}
		if reason, excluded := patch.Exclusions[o.Name]; excluded {
			out.Excluded[o.Name] = reason
			continue
		}
		out.Obstacles = append(out.Obstacles, ob)
		obstacles = append(obstacles, routeInflate(ob.Bounds, c+.04))
	}
	for name := range patch.Exclusions {
		if !known[name] {
			return out, fmt.Errorf("excluded object not found: %s", name)
		}
	}
	// Endpoint surfaces constrain the search beyond their outward stubs, preventing
	// a supposedly clear route from travelling back through its own endpoint.
	for _, port := range []DiagramPort{from, to} {
		if o, exists := world[port.NativeObject]; exists {
			obstacles = append(obstacles, routeInflate(o.WorldBounds, c+.04))
		}
	}
	for j, r := range patch.Reserved {
		out.Obstacles = append(out.Obstacles, DiagramRoutingObstacle{Name: fmt.Sprintf("reserved-%03d", j+1), Kind: "reserved", Bounds: r})
		obstacles = append(obstacles, routeInflate(r, c+.04))
	}
	a, z := [2]float64{from.X, from.Y}, [2]float64{to.X, to.Y}
	first, last := routeStub(from, c+.04), routeStub(to, c+.04)
	// Non-endpoint allocations must also leave the outward attachment stubs clear.
	for _, ob := range out.Obstacles {
		inflated := routeInflate(ob.Bounds, c+.04)
		if !routeSegmentClear(a, first, []wmdesign.Rect{inflated}) || !routeSegmentClear(last, z, []wmdesign.Rect{inflated}) {
			out.Message = "endpoint attachment corridor blocked; choose another site or move blocker"
			return out, nil
		}
	}
	middle, e := orthogonalRoute(first, last, out.Bounds, obstacles)
	if e != nil {
		out.Message = e.Error()
		return out, nil
	}
	out.Points = simplifyOrthogonalRoute(append(append([][2]float64{a}, middle...), z))
	delete(node.Arguments, "bend")
	node.Arguments["routing_policy"] = out.Policy
	if len(out.Points) == 2 {
		node.Arguments["route"] = "straight"
		delete(node.Arguments, "waypoints")
	} else {
		node.Arguments["route"] = "polyline"
		waypoints := [][2]float64{}
		for _, xy := range out.Points[1 : len(out.Points)-1] {
			waypoints = append(waypoints, [2]float64{xy[0] - origin[0], xy[1] - origin[1]})
		}
		node.Arguments["waypoints"] = waypoints
	}
	if patch.Label != nil {
		node.Arguments["label"] = *patch.Label
	}
	if patch.LabelPosition != nil {
		node.Arguments["label_position"] = *patch.LabelPosition
	}
	if patch.LabelWidth != nil {
		node.Arguments["label_width"] = *patch.LabelWidth
	}
	if list == nil {
		t.Nodes = append(t.Nodes, node)
	} else {
		(*list)[i] = node
	}
	candidate, e := compositionCandidate(p, slideID, "diagram-route", patch.Actor, patch.Reason, t, bundle, engine, apply, out, nil)
	if e != nil {
		return out, e
	}
	out.Candidate = &candidate
	out.Status = "proposed"
	out.Message = "Explicit orthogonal route; endpoint meaning preserved. Review rendered arrows and labels."
	if apply {
		out.Status = "applied"
	}
	return out, nil
}
