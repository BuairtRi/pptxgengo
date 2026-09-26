package compose

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
)

// Connections target root roles or complete pods. Nested role routing is not
// supported: the enclosing pod must own routing through its internal content.
type ConnectionSpec struct {
	ID           string  `json:"id"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	Relationship string  `json:"relationship"`
	FromAnchor   string  `json:"from_anchor"`
	ToAnchor     string  `json:"to_anchor"`
	Color        string  `json:"color"`
	WidthPt      float64 `json:"width_pt"`
	ClearancePt  float64 `json:"clearance_pt"`
}
type PlannedConnection struct {
	ID           string  `json:"id"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	Relationship string  `json:"relationship"`
	Color        string  `json:"color"`
	WidthPt      float64 `json:"width_pt"`
	ClearancePt  float64 `json:"clearance_pt"`
	Points       []Point `json:"points"`
	Strategy     string  `json:"strategy"`
}
type obstacle struct {
	id   string
	rect Rect
}

func expand(r Rect, d float64) Rect { return Rect{r.X - d, r.Y - d, r.Width + 2*d, r.Height + 2*d} }
func samePoint(a, b Point) bool     { return math.Abs(a.X-b.X) < 1e-7 && math.Abs(a.Y-b.Y) < 1e-7 }
func pointIn(p Point, r Rect) bool {
	return p.X > r.X+1e-7 && p.X < r.X+r.Width-1e-7 && p.Y > r.Y+1e-7 && p.Y < r.Y+r.Height-1e-7
}
func segmentHits(a, b Point, r Rect) bool {
	if math.Abs(a.X-b.X) < 1e-7 {
		return a.X > r.X+1e-7 && a.X < r.X+r.Width-1e-7 && math.Max(a.Y, b.Y) > r.Y+1e-7 && math.Min(a.Y, b.Y) < r.Y+r.Height-1e-7
	}
	if math.Abs(a.Y-b.Y) < 1e-7 {
		return a.Y > r.Y+1e-7 && a.Y < r.Y+r.Height-1e-7 && math.Max(a.X, b.X) > r.X+1e-7 && math.Min(a.X, b.X) < r.X+r.Width-1e-7
	}
	return true
}
func anchor(r Rect, side string) (Point, Point, error) {
	switch side {
	case "top":
		return rectAnchors(r).Top, Point{0, -1}, nil
	case "bottom":
		return rectAnchors(r).Bottom, Point{0, 1}, nil
	case "left":
		return rectAnchors(r).Left, Point{-1, 0}, nil
	case "right":
		return rectAnchors(r).Right, Point{1, 0}, nil
	}
	return Point{}, Point{}, fmt.Errorf("anchor must be top, bottom, left or right: %q", side)
}
func cross(a, b, c, d Point) bool {
	// Closed intersection of axis-aligned segments, including touching/overlap.
	return math.Max(math.Min(a.X, b.X), math.Min(c.X, d.X)) <= math.Min(math.Max(a.X, b.X), math.Max(c.X, d.X))+1e-7 && math.Max(math.Min(a.Y, b.Y), math.Min(c.Y, d.Y)) <= math.Min(math.Max(a.Y, b.Y), math.Max(c.Y, d.Y))+1e-7
}
func simplePoints(p []Point) []Point {
	var out []Point
	for _, q := range p {
		if len(out) > 0 && samePoint(q, out[len(out)-1]) {
			continue
		}
		out = append(out, q)
		for len(out) >= 3 {
			n := len(out)
			a, b, c := out[n-3], out[n-2], out[n-1]
			if (math.Abs(a.X-b.X) < 1e-7 && math.Abs(b.X-c.X) < 1e-7) || (math.Abs(a.Y-b.Y) < 1e-7 && math.Abs(b.Y-c.Y) < 1e-7) {
				out[n-2] = c
				out = out[:n-1]
			} else {
				break
			}
		}
	}
	return out
}
func planConnections(s SlideSpec, p *PlannedSlide) error {
	if len(s.Connections) == 0 {
		return nil
	}
	roots := map[string]Rect{}
	ids := map[string]bool{}
	for _, q := range p.Pods {
		roots[q.ID] = q.Bounds
		ids[q.ID] = true
		for _, r := range q.Roles {
			ids[r.ID] = true
		}
	}
	for _, r := range p.Roles {
		roots[r.ID] = r.Bounds
		ids[r.ID] = true
	}
	var obs []obstacle
	// Stable obstacle order, independent of map iteration.
	for _, q := range p.Pods {
		obs = append(obs, obstacle{requestID("route-root", q.ID), q.Bounds})
	}
	for _, q := range p.Roles {
		obs = append(obs, obstacle{requestID("route-root", q.ID), q.Bounds})
	}
	if p.TitleMeasurementID != "" {
		obs = append(obs, obstacle{requestID("route-title"), p.TitleBounds})
	}
	for _, ph := range p.Phases {
		ids[ph.ID] = true
		obs = append(obs, obstacle{requestID("route-phase", ph.ID), ph.TitleRect})
	}
	if p.Legend != nil {
		obs = append(obs, obstacle{requestID("route-legend"), p.Legend.Bounds})
	}
	if err := validateConnectionGraph(s, roots, ids); err != nil {
		return err
	}
	for _, c := range s.Connections {
		a, da, e := anchor(roots[c.From], c.FromAnchor)
		if e != nil {
			return e
		}
		b, db, e := anchor(roots[c.To], c.ToAnchor)
		if e != nil {
			return e
		}
		clearance := c.ClearancePt + c.WidthPt/2
		start, end := Point{a.X + da.X*clearance, a.Y + da.Y*clearance}, Point{b.X + db.X*clearance, b.Y + db.Y*clearance}
		canvas := Rect{c.WidthPt / 2, c.WidthPt / 2, s.WidthPt - c.WidthPt, s.HeightPt - c.WidthPt}
		inCanvas := func(q Point) bool {
			return q.X >= canvas.X && q.X <= canvas.X+canvas.Width && q.Y >= canvas.Y && q.Y <= canvas.Y+canvas.Height
		}
		if !inCanvas(start) || !inCanvas(end) {
			return fmt.Errorf("connection %s anchor clearance exceeds slide", c.ID)
		}
		var expanded []obstacle
		for _, o := range obs {
			expanded = append(expanded, obstacle{o.id, expand(o.rect, clearance)})
		}
		priorOK := func(x, y Point) bool {
			for _, prev := range p.Connections {
				if prev.From == c.From && prev.Relationship == c.Relationship && prev.Color == mustColor(c.Color) && prev.WidthPt == c.WidthPt && samePoint(prev.Points[0], a) {
					continue
				} // A deliberate common-source reporting bus.
				for i := 1; i < len(prev.Points); i++ {
					u, v := prev.Points[i-1], prev.Points[i]
					if !cross(x, y, u, v) {
						dx := math.Max(0, math.Max(math.Min(x.X, y.X), math.Min(u.X, v.X))-math.Min(math.Max(x.X, y.X), math.Max(u.X, v.X)))
						dy := math.Max(0, math.Max(math.Min(x.Y, y.Y), math.Min(u.Y, v.Y))-math.Min(math.Max(x.Y, y.Y), math.Max(u.Y, v.Y)))
						if math.Hypot(dx, dy) < (c.WidthPt+prev.WidthPt)/2+0.5 {
							return false
						}
						continue
					}
					// Different relationships can touch only at an actual shared endpoint.
					allowed := false
					for _, q := range []Point{a, b} {
						if (samePoint(x, q) || samePoint(y, q)) && (samePoint(u, q) || samePoint(v, q)) && (samePoint(q, prev.Points[0]) || samePoint(q, prev.Points[len(prev.Points)-1])) && intersectionOnlyAt(x, y, u, v, q) {
							allowed = true
						}
					}
					if !allowed {
						return false
					}
				}
			}
			return true
		}
		clear := func(x, y Point, skip string) bool {
			if !inCanvas(x) || !inCanvas(y) {
				return false
			}
			for _, o := range expanded {
				if o.id != skip && segmentHits(x, y, o.rect) {
					return false
				}
			}
			return priorOK(x, y)
		}
		if !clear(a, start, requestID("route-root", c.From)) || !clear(end, b, requestID("route-root", c.To)) {
			return fmt.Errorf("connection %s endpoint clearance is blocked", c.ID)
		}
		valid := func(path []Point) bool {
			for i := 1; i < len(path); i++ {
				if !clear(path[i-1], path[i], "") {
					return false
				}
			}
			full := simplePoints(append(append([]Point{a}, path...), b))
			for _, prev := range p.Connections {
				if prev.From == c.From && samePoint(prev.Points[0], a) && !sharedPrefixOnly(full, prev.Points) {
					return false
				}
			}
			return true
		}
		// Prefer an open, centered lane; use a visibility grid only when obstructed.
		midY, midX := (start.Y+end.Y)/2, (start.X+end.X)/2
		candidates := [][]Point{}
		if math.Abs(start.X-end.X) < 1e-7 || math.Abs(start.Y-end.Y) < 1e-7 {
			candidates = append(candidates, []Point{start, end})
		}
		if da.Y != 0 && db.Y != 0 {
			candidates = append(candidates, []Point{start, {start.X, midY}, {end.X, midY}, end})
		}
		if da.X != 0 && db.X != 0 {
			candidates = append(candidates, []Point{start, {midX, start.Y}, {midX, end.Y}, end})
		}
		candidates = append(candidates, []Point{start, {start.X, end.Y}, end}, []Point{start, {end.X, start.Y}, end})
		strategy := "clear_candidate"
		var route []Point
		for _, candidate := range candidates {
			if valid(candidate) {
				route = candidate
				break
			}
		}
		if route == nil {
			strategy = "visibility_grid"
			route = gridRoute(start, end, canvas, expanded, func(x, y Point) bool { return clear(x, y, "") })
		}
		if route == nil || !valid(route) {
			return fmt.Errorf("connection %s has no clear orthogonal route; move components or reserve a routing lane", c.ID)
		}
		points := append([]Point{a}, route...)
		points = append(points, b)
		points = simplePoints(points)
		for _, ph := range p.Phases {
			touches := false
			for i := 1; i < len(points); i++ {
				if segmentHits(points[i-1], points[i], ph.Bounds) {
					touches = true
				}
			}
			if touches {
				ratio, _ := ContrastRatio(c.Color, ph.Surface)
				if ratio < 3 {
					return fmt.Errorf("connection %s stroke contrast on phase %s must be >=3:1", c.ID, ph.ID)
				}
			}
		}

		p.Connections = append(p.Connections, PlannedConnection{ID: c.ID, From: c.From, To: c.To, Relationship: c.Relationship, Color: mustColor(c.Color), WidthPt: c.WidthPt, ClearancePt: c.ClearancePt, Points: points, Strategy: strategy})
	}
	return nil
}
func mustColor(s string) string { v, _ := resolveColor(s); return v }

type routeState struct{ node, dir int }
type routeStep struct {
	state  routeState
	cost   float64
	serial int
}
type routeQueue []routeStep

func (q routeQueue) Len() int { return len(q) }
func (q routeQueue) Less(i, j int) bool {
	if math.Abs(q[i].cost-q[j].cost) > 1e-7 {
		return q[i].cost < q[j].cost
	}
	return q[i].serial < q[j].serial
}
func (q routeQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *routeQueue) Push(x any)   { *q = append(*q, x.(routeStep)) }
func (q *routeQueue) Pop() any     { old := *q; x := old[len(old)-1]; *q = old[:len(old)-1]; return x }
func sortedUnique(v []float64, lo, hi float64) []float64 {
	sort.Float64s(v)
	var out []float64
	for _, x := range v {
		if x < lo || x > hi {
			continue
		}
		if len(out) == 0 || math.Abs(x-out[len(out)-1]) > 1e-7 {
			out = append(out, x)
		}
	}
	return out
}
func gridRoute(a, b Point, canvas Rect, obs []obstacle, clear func(Point, Point) bool) []Point {
	xs, ys := []float64{a.X, b.X, canvas.X, canvas.X + canvas.Width}, []float64{a.Y, b.Y, canvas.Y, canvas.Y + canvas.Height}
	for _, o := range obs {
		xs = append(xs, o.rect.X, o.rect.X+o.rect.Width)
		ys = append(ys, o.rect.Y, o.rect.Y+o.rect.Height)
	}
	xs = sortedUnique(xs, canvas.X, canvas.X+canvas.Width)
	ys = sortedUnique(ys, canvas.Y, canvas.Y+canvas.Height)
	if len(xs)*len(ys) > 50000 {
		return nil
	}
	points := make([]Point, len(xs)*len(ys))
	allowed := make([]bool, len(points))
	start, goal := -1, -1
	for j, y := range ys {
		for i, x := range xs {
			k := j*len(xs) + i
			q := Point{x, y}
			points[k] = q
			allowed[k] = true
			for _, o := range obs {
				if pointIn(q, o.rect) {
					allowed[k] = false
					break
				}
			}
			if samePoint(q, a) {
				start = k
			}
			if samePoint(q, b) {
				goal = k
			}
		}
	}
	if start < 0 || goal < 0 || !allowed[start] || !allowed[goal] {
		return nil
	}
	first := routeState{start, 0}
	dist := map[routeState]float64{first: 0}
	prev := map[routeState]routeState{}
	queue := &routeQueue{{first, 0, 0}}
	heap.Init(queue)
	serial := 0
	for queue.Len() > 0 {
		cur := heap.Pop(queue).(routeStep)
		if cur.cost > dist[cur.state]+1e-7 {
			continue
		}
		if cur.state.node == goal {
			var path []Point
			v := cur.state
			for {
				path = append(path, points[v.node])
				if v == first {
					break
				}
				v = prev[v]
			}
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return simplePoints(path)
		}
		n := cur.state.node
		i, j := n%len(xs), n/len(xs)
		neighbors := []int{}
		if i > 0 {
			neighbors = append(neighbors, n-1)
		}
		if i+1 < len(xs) {
			neighbors = append(neighbors, n+1)
		}
		if j > 0 {
			neighbors = append(neighbors, n-len(xs))
		}
		if j+1 < len(ys) {
			neighbors = append(neighbors, n+len(xs))
		}
		for _, v := range neighbors {
			if !allowed[v] || !clear(points[n], points[v]) {
				continue
			}
			dir := 1
			if math.Abs(points[n].X-points[v].X) < 1e-7 {
				dir = 2
			}
			cost := cur.cost + math.Abs(points[n].X-points[v].X) + math.Abs(points[n].Y-points[v].Y)
			if cur.state.dir != 0 && dir != cur.state.dir {
				cost += 24
			}
			next := routeState{v, dir}
			old, ok := dist[next]
			if !ok || cost < old-1e-7 {
				dist[next] = cost
				prev[next] = cur.state
				serial++
				heap.Push(queue, routeStep{next, cost, serial})
			}
		}
	}
	return nil
}

func validateConnectionGraph(s SlideSpec, roots map[string]Rect, ids map[string]bool) error {
	// Reject duplicate directed relationships and cycles in the reporting graph.
	pairs := map[string]bool{}
	graph := map[string][]string{}
	for _, c := range s.Connections {
		if !validID(c.ID) || ids[c.ID] {
			return fmt.Errorf("slide %s duplicate/empty connection ID %q", s.ID, c.ID)
		}
		ids[c.ID] = true
		if c.From == c.To {
			return fmt.Errorf("connection %s cannot connect a component to itself", c.ID)
		}
		if _, ok := roots[c.From]; !ok {
			return fmt.Errorf("connection %s unknown root source %q (nested roles are unsupported)", c.ID, c.From)
		}
		if _, ok := roots[c.To]; !ok {
			return fmt.Errorf("connection %s unknown root target %q (nested roles are unsupported)", c.ID, c.To)
		}
		if c.Relationship != "reporting" {
			return fmt.Errorf("connection %s relationship must be reporting", c.ID)
		}
		key := requestID(c.From, c.To, c.Relationship)
		if pairs[key] {
			return fmt.Errorf("duplicate relationship %s", c.ID)
		}
		pairs[key] = true
		if c.Relationship == "reporting" {
			graph[c.From] = append(graph[c.From], c.To)
		}
		if !positive(c.WidthPt) || c.WidthPt > 6 || !positive(c.ClearancePt) || c.ClearancePt < 2 {
			return fmt.Errorf("connection %s requires width (0,6]pt and clearance >=2pt", c.ID)
		}
		if _, e := resolveColor(c.Color); e != nil {
			return fmt.Errorf("connection %s: %w", c.ID, e)
		}
		if _, _, e := anchor(Rect{}, c.FromAnchor); e != nil {
			return e
		}
		if _, _, e := anchor(Rect{}, c.ToAnchor); e != nil {
			return e
		}
		ratio, _ := ContrastRatio(c.Color, white)
		if ratio < 3 {
			return fmt.Errorf("connection %s stroke contrast on white must be >=3:1", c.ID)
		}
	}
	state := map[string]int{}
	var visit func(string) bool
	visit = func(id string) bool {
		if state[id] == 1 {
			return false
		}
		if state[id] == 2 {
			return true
		}
		state[id] = 1
		for _, n := range graph[id] {
			if !visit(n) {
				return false
			}
		}
		state[id] = 2
		return true
	}
	for _, c := range s.Connections {
		if !visit(c.From) {
			return fmt.Errorf("slide %s reporting relationships contain a cycle", s.ID)
		}
	}

	return nil
}

func validateConnectionSpecs(s SlideSpec) error {
	roots := map[string]Rect{}
	ids := map[string]bool{}
	for _, p := range s.Pods {
		roots[p.ID] = Rect{}
		ids[p.ID] = true
		for _, r := range p.Roles {
			ids[r.ID] = true
		}
	}
	for _, r := range s.Roles {
		roots[r.ID] = r.Bounds
		ids[r.ID] = true
	}
	for _, p := range s.Phases {
		ids[p.ID] = true
	}
	return validateConnectionGraph(s, roots, ids)
}

// Shared reporting paths may share only one continuous prefix from their source.
// Branches cannot rejoin/cross later, which would create an ambiguous diagram.
func sharedPrefixOnly(first, second []Point) bool {
	if len(first) == 0 || len(second) == 0 || !samePoint(first[0], second[0]) {
		return false
	}
	a, b := append([]Point(nil), first...), append([]Point(nil), second...)
	direction := func(x, y Point) Point {
		dx, dy := y.X-x.X, y.Y-x.Y
		l := math.Abs(dx) + math.Abs(dy)
		if l == 0 {
			return Point{}
		}
		return Point{dx / l, dy / l}
	}
	for len(a) > 1 && len(b) > 1 {
		da, db := direction(a[0], a[1]), direction(b[0], b[1])
		if !samePoint(da, db) {
			break
		}
		la, lb := math.Abs(a[1].X-a[0].X)+math.Abs(a[1].Y-a[0].Y), math.Abs(b[1].X-b[0].X)+math.Abs(b[1].Y-b[0].Y)
		l := math.Min(la, lb)
		at := Point{a[0].X + da.X*l, a[0].Y + da.Y*l}
		if math.Abs(l-la) < 1e-7 {
			a = a[1:]
		} else {
			a[0] = at
		}
		if math.Abs(l-lb) < 1e-7 {
			b = b[1:]
		} else {
			b[0] = at
		}
	}
	branch := a[0]
	for i := 1; i < len(a); i++ {
		for j := 1; j < len(b); j++ {
			x, y, u, v := a[i-1], a[i], b[j-1], b[j]
			if !cross(x, y, u, v) {
				continue
			}
			lo := Point{math.Max(math.Min(x.X, y.X), math.Min(u.X, v.X)), math.Max(math.Min(x.Y, y.Y), math.Min(u.Y, v.Y))}
			hi := Point{math.Min(math.Max(x.X, y.X), math.Max(u.X, v.X)), math.Min(math.Max(x.Y, y.Y), math.Max(u.Y, v.Y))}
			if !samePoint(lo, hi) || !samePoint(lo, branch) {
				return false
			}
		}
	}
	return true
}

func intersectionOnlyAt(x, y, u, v, p Point) bool {
	if !cross(x, y, u, v) {
		return false
	}
	lo := Point{math.Max(math.Min(x.X, y.X), math.Min(u.X, v.X)), math.Max(math.Min(x.Y, y.Y), math.Min(u.Y, v.Y))}
	hi := Point{math.Min(math.Max(x.X, y.X), math.Max(u.X, v.X)), math.Min(math.Max(x.Y, y.Y), math.Max(u.Y, v.Y))}
	return samePoint(lo, hi) && samePoint(lo, p)
}
