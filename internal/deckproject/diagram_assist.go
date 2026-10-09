package deckproject

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type DiagramPort struct {
	Node         string  `json:"node"`
	NativeObject string  `json:"native_object"`
	Site         string  `json:"site"`
	X            float64 `json:"x_pt"`
	Y            float64 `json:"y_pt"`
}
type DiagramConnection struct {
	Route          string       `json:"route"`
	Representation string       `json:"representation,omitempty"`
	NativeParts    []string     `json:"native_parts,omitempty"`
	Points         [][2]float64 `json:"points"`
	Node           string       `json:"node"`
	From           DiagramPort  `json:"from"`
	To             DiagramPort  `json:"to"`
}

func diagramPorts(n string, b wmdesign.Rect, native string) []DiagramPort {
	return []DiagramPort{{n, native, "top", b.X + b.W/2, b.Y}, {n, native, "left", b.X, b.Y + b.H/2}, {n, native, "bottom", b.X + b.W/2, b.Y + b.H}, {n, native, "right", b.X + b.W, b.Y + b.H/2}}
}
func addDiagramPorts(out *DiagramInspection, c wmdesign.Document) error {
	native := map[string]NativeGeometryObservation{}
	objects := map[string]*geometryObject{}
	for _, o := range out.FinalNative {
		native[o.Name] = o
		objects[o.Name] = &geometryObject{geometry: o.Geometry}
	}
	for _, node := range c.Slides[0].Nodes {
		if node.Scene == nil {
			continue
		}
		var src struct {
			Type string `json:"type"`
		}
		if e := json.Unmarshal(node.Scene.Node, &src); e != nil {
			return e
		}
		name := node.ID
		if src.Type == "block" {
			name += ".surface"
		} else if src.Type != "editable-block" {
			continue
		}
		o, ok := native[name]
		if !ok {
			return fmt.Errorf("diagram port native target missing: %s", name)
		}
		g := o.Geometry
		for _, port := range diagramPorts(node.ID, wmdesign.Rect{X: g.X, Y: g.Y, W: g.W, H: g.H}, name) {
			xy, e := nativeObjectPoint(objects, name, port.X, port.Y)
			if e != nil {
				return e
			}
			port.X, port.Y = xy[0], xy[1]
			out.Ports = append(out.Ports, port)
		}
	}
	ports := map[string]DiagramPort{}
	for _, p := range out.Ports {
		ports[p.Node+"\x00"+p.Site] = p
	}
	for _, node := range c.Slides[0].Nodes {
		if node.Scene == nil {
			continue
		}
		var src struct {
			Type      string       `json:"type"`
			Waypoints [][2]float64 `json:"waypoints,omitempty"`
			From, To  struct{ Node, Site string }
		}
		if e := json.Unmarshal(node.Scene.Node, &src); e != nil {
			return e
		}
		if src.Type != "attached-connector" {
			continue
		}
		a, ok := ports[src.From.Node+"\x00"+src.From.Site]
		z, ok2 := ports[src.To.Node+"\x00"+src.To.Site]
		if ok && ok2 {
			connection := DiagramConnection{Node: node.ID, From: a, To: z, Route: "straight"}
			if connector, exists := native[node.ID]; exists {
				if connector.Geometry.Route != nil {
					connection.Route = connector.Geometry.Route.Preset
				}
				for _, point := range nativeConnectorPoints(connector.Geometry) {
					xy, e := nativeObjectPoint(objects, node.ID, point[0], point[1])
					if e != nil {
						return e
					}
					connection.Points = append(connection.Points, xy)
				}
			}
			if _, single := native[node.ID]; !single {
				count := len(src.Waypoints) + 1
				if count < 2 || count > 127 {
					return fmt.Errorf("multipart connection lacks bounded authored waypoint count")
				}
				connection.Route = "polyline"
				connection.Representation = "attached-native-segments"
				expected := map[string]bool{}
				for i := 1; i <= count; i++ {
					name := fmt.Sprintf("%s.segment-%03d", node.ID, i)
					expected[name] = true
					segment, ok := native[name]
					if !ok || segment.Geometry.Kind != "cxnSp" || segment.Geometry.Route != nil {
						return fmt.Errorf("multipart route segment missing or unsupported: %s", name)
					}
					connection.NativeParts = append(connection.NativeParts, name)
					points := nativeConnectorPoints(segment.Geometry)
					for j, point := range points {
						world, e := nativeObjectPoint(objects, name, point[0], point[1])
						if e != nil {
							return e
						}
						if i > 1 && j == 0 {
							prior := connection.Points[len(connection.Points)-1]
							if math.Hypot(prior[0]-world[0], prior[1]-world[1]) > .02 {
								return fmt.Errorf("multipart native segments are disconnected")
							}
							continue
						}
						connection.Points = append(connection.Points, world)
					}
					if i < count {
						guide := fmt.Sprintf("%s.waypoint-%03d", node.ID, i)
						expected[guide] = true
						g, ok := native[guide]
						if !ok || g.Geometry.Kind != "sp" {
							return fmt.Errorf("multipart waypoint guide missing: %s", guide)
						}
						connection.NativeParts = append(connection.NativeParts, guide)
					}
				}
				for name := range native {
					if (strings.HasPrefix(name, node.ID+".segment-") || strings.HasPrefix(name, node.ID+".waypoint-")) && !expected[name] {
						return fmt.Errorf("unexpected multipart route namespace object: %s", name)
					}
				}
			}

			out.Connections = append(out.Connections, connection)
		}
	}
	return nil
}

// Connections retain endpoint semantics in YAML. The renderer resolves
// the real named rectangle/site; moving either authored node recalculates them.
func ConnectDiagram(p *Project, slide, id, from, fromSite, to, toSite, head, style, actor, reason, bundle, engine string, apply bool) (DiagramPatchResult, error) {
	return ConnectRoutedDiagram(p, slide, id, from, fromSite, to, toSite, head, style, "straight", nil, actor, reason, bundle, engine, apply)
}
func ConnectRoutedDiagram(p *Project, slide, id, from, fromSite, to, toSite, head, style, route string, bend *float64, actor, reason, bundle, engine string, apply bool) (DiagramPatchResult, error) {
	inspect, e := InspectDiagram(p, slide, bundle, engine)
	if e != nil {
		return DiagramPatchResult{}, e
	}
	ports := map[string]DiagramPort{}
	for _, port := range inspect.Ports {
		ports[port.Node+"\x00"+port.Site] = port
	}
	if from == to {
		return DiagramPatchResult{}, fmt.Errorf("connection requires distinct nodes")
	}
	if _, ok := ports[from+"\x00"+fromSite]; !ok {
		return DiagramPatchResult{}, fmt.Errorf("unknown supported rectangle/site %s/%s", from, fromSite)
	}
	if _, ok := ports[to+"\x00"+toSite]; !ok {
		return DiagramPatchResult{}, fmt.Errorf("unknown supported rectangle/site %s/%s", to, toSite)
	}
	// Allocation is the frame body, not a guessed path envelope. The renderer
	// still checks the final line against this allocation and split reservations.
	r := wmdesign.Rect{X: 0, Y: 0, W: inspect.Frame.Body.W, H: inspect.Frame.Body.H}
	node := Node{ID: id, Kind: "component", Placement: &Placement{Zone: "body", Rect: &r}, Definition: &Reference{Scope: "shared", ID: "wmds/component/attached-connector"}, Arguments: map[string]any{"from": map[string]any{"node": from, "site": fromSite}, "to": map[string]any{"node": to, "site": toSite}, "head": head, "style": style}}
	node.Arguments["route"] = route
	if bend != nil {
		node.Arguments["bend"] = *bend
	}
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: actor, Reason: reason, Operations: []DiagramOperation{{Action: "add", ID: id, Node: &node}}}
	return PatchDiagram(p, slide, patch, bundle, engine, apply)
}
func ArrangeDiagram(p *Project, slide string, ids []string, align, distribute, actor, reason, bundle, engine string, apply bool) (DiagramPatchResult, error) {
	_, t, e := diagramSlide(p, slide)
	if e != nil {
		return DiagramPatchResult{}, e
	}
	if len(ids) < 2 {
		return DiagramPatchResult{}, fmt.Errorf("arrange requires at least two nodes")
	}
	if align == "" && distribute == "" {
		return DiagramPatchResult{}, fmt.Errorf("arrange requires align or distribute")
	}
	seen := map[string]bool{}
	rects := []wmdesign.Rect{}
	zone := ""
	for _, id := range ids {
		if seen[id] {
			return DiagramPatchResult{}, fmt.Errorf("duplicate arrange node")
		}
		seen[id] = true
		nodes, i := findDiagramNode(&t.Nodes, id)
		if nodes == nil {
			return DiagramPatchResult{}, fmt.Errorf("unknown arrange node %s", id)
		}
		n := (*nodes)[i]
		if n.Placement == nil || n.Placement.Rect == nil {
			return DiagramPatchResult{}, fmt.Errorf("arrange requires rect placements")
		}
		if zone == "" {
			zone = n.Placement.Zone
		}
		if n.Placement.Zone != zone {
			return DiagramPatchResult{}, fmt.Errorf("arrange nodes must occupy the same frame zone")
		}
		rects = append(rects, *n.Placement.Rect)
	}
	anchor := rects[0]
	for i := range rects {
		r := &rects[i]
		switch align {
		case "":
		case "left":
			r.X = anchor.X
		case "right":
			r.X = anchor.X + anchor.W - r.W
		case "top":
			r.Y = anchor.Y
		case "bottom":
			r.Y = anchor.Y + anchor.H - r.H
		case "center":
			r.X = anchor.X + (anchor.W-r.W)/2
		case "middle":
			r.Y = anchor.Y + (anchor.H-r.H)/2
		default:
			return DiagramPatchResult{}, fmt.Errorf("unknown alignment")
		}
	}
	// Distribution uses caller-specified node order, with first/last positions
	// fixed. Negative gaps are surfaced rather than silently reversing topology.
	switch distribute {
	case "":
	case "horizontal", "vertical":
		first, last := rects[0], rects[len(rects)-1]
		extent := last.X + last.W - first.X
		sum := 0.
		for _, r := range rects {
			sum += r.W
		}
		if distribute == "vertical" {
			extent = last.Y + last.H - first.Y
			sum = 0
			for _, r := range rects {
				sum += r.H
			}
		}
		gap := (extent - sum) / float64(len(rects)-1)
		if gap < -1e-6 || math.IsNaN(gap) {
			return DiagramPatchResult{}, fmt.Errorf("distribution requires nonoverlapping space in the supplied node order")
		}
		pos := first.X
		if distribute == "vertical" {
			pos = first.Y
		}
		for i := range rects {
			if distribute == "horizontal" {
				rects[i].X = pos
				pos += rects[i].W + gap
			} else {
				rects[i].Y = pos
				pos += rects[i].H + gap
			}
		}
	default:
		return DiagramPatchResult{}, fmt.Errorf("unknown distribution")
	}
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: actor, Reason: reason}
	for i, id := range ids {
		r := rects[i]
		patch.Operations = append(patch.Operations, DiagramOperation{Action: "move", ID: strings.TrimSpace(id), Rect: &r})
	}
	return PatchDiagram(p, slide, patch, bundle, engine, apply)
}
