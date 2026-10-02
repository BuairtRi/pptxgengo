package wmdesign

import "encoding/json"

// DiagramSceneReference is a focused synthetic reference; it is separate from
// frozen library examples and does not assert native typography qualification.
func DiagramSceneReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	node := func(id string, v any) Node {
		raw, _ := json.Marshal(v)
		return Node{ID: id, Kind: "scene", Scene: &SceneSpec{Node: raw, Path: "/reference/" + id}}
	}
	slide := func(id, title string, nodes ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: "source-diagram-" + id, Eyebrow: "WMDS native source scenes", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact"}, Nodes: nodes})
	}
	slide("shapes", "Source geometry stays editable",
		node("boundary", map[string]any{"type": "container", "x": 57, "y": 144, "w": 846, "h": 252, "style": "boundary", "label": "Delivery boundary"}),
		node("input", map[string]any{"type": "node", "x": 93, "y": 234, "w": 198, "h": 90, "surface": "subtle", "icon": "database", "text": "Input", "sub": "Defined source"}),
		node("decision", map[string]any{"type": "chevron", "x": 363, "y": 234, "w": 234, "h": 90, "first": true, "surface": "inverse", "text": "Review", "sub": "Confirm the owner", "style": "small", "number": "02"}),
		node("output", map[string]any{"type": "block", "x": 687, "y": 234, "w": 180, "h": 90, "surface": "callout", "text": "Approved", "style": "body"}),
		node("input-review", map[string]any{"type": "connector", "points": [][2]float64{{291, 279}, {363, 279}}, "head": "end"}),
		node("review-output", map[string]any{"type": "connector", "points": [][2]float64{{597, 279}, {687, 279}}, "head": "end", "style": "dashed"}))
	steps := []map[string]any{{"key": "discover", "state": "done", "label": "01", "title": "Discover"}, {"key": "design", "state": "current", "label": "02", "title": "Design"}, {"key": "deliver", "state": "next", "label": "03", "title": "Deliver"}}
	slide("sequence", "Persistent keys retain semantic identity",
		node("steps", map[string]any{"type": "stepper", "x": 57, "y": 144, "w": 846, "steps": steps}),
		node("axis", map[string]any{"type": "timeaxis", "x": 129, "y": 342, "w": 702, "periods": []string{"Week 1", "Week 2", "Week 3"}, "milestones": []map[string]any{{"key": "approval", "at": .5, "label": "Approval"}}, "today": .25}))
	slide("people", "People and reporting remain native objects",
		node("tree", map[string]any{"type": "orgchart", "x": 57, "y": 144, "w": 846, "root": map[string]any{"org": "client", "title": "Sponsor", "name": "Morgan Lee", "children": []map[string]any{{"key": "delivery", "org": "wm", "title": "Delivery lead", "name": "Alex Chen"}, {"key": "operations", "org": "client", "title": "Operations lead", "name": "Sam Rivera", "dotted": true}}}}),
		node("legend", map[string]any{"type": "legend", "x": 57, "y": 396, "w": 846, "layout": "horizontal", "items": []map[string]any{{"series": 8, "text": "West Monroe"}, {"series": 4, "text": "Client"}, {"swatch": "line", "dashed": true, "text": "Dotted reporting"}}}))
	slide("gantt", "Measured labels determine timeline packing",
		node("timeline", map[string]any{"type": "gantt", "x": 57, "y": 144, "w": 846, "cols": map[string]any{"group": 24, "lane": 192}, "periods": map[string]any{"labels": []string{"W1", "W2", "W3", "W4"}}, "phases": []map[string]any{{"key": "delivery", "label": "Delivery", "from": 0, "to": 4, "rule": "strong"}}, "kinds": map[string]any{"core": map[string]any{"fill": "#070154", "text": "#FFFFFF", "label": "Core activity"}}, "events": map[string]any{"star": map[string]any{"fill": "#F900D3", "label": "Workshop"}}, "groups": []map[string]any{{"key": "delivery", "label": "Delivery", "fill": "#070154", "lanes": []map[string]any{{"key": "build", "title": "Build", "icon": "layers", "items": []map[string]any{{"key": "design", "kind": "core", "from": 0, "to": 2, "label": "Design", "softEnd": .25}, {"key": "review", "event": "star", "at": 2.3, "label": "Review"}}}, {"key": "release", "title": "Release", "icon": "target", "items": []map[string]any{{"key": "deploy", "kind": "core", "from": 1, "to": 4, "label": "Deploy"}}}}}}}))
	return d
}
