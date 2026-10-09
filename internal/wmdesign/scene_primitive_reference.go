package wmdesign

import (
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/assetregistry"
)

// primitiveAssetRegistry pins canonical local bytes. Runtime roots may relocate
// these files but cannot alter any registered payload.
var primitiveAssetRegistry = assetregistry.Catalog()

// PrimitiveSceneReference is synthetic diagnostic copy for source primitives.
// It is generation input, not native qualification or a test harness.
func PrimitiveSceneReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	node := func(id string, data any) Node {
		raw, _ := json.Marshal(data)
		return Node{ID: id, Kind: "scene", Rect: Rect{X: 57, Y: 144, W: 846, H: 288}, Scene: &SceneSpec{Node: raw, Path: "/reference/" + id}}
	}
	slide := func(id, title string, nodes ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: "primitive-" + id, Eyebrow: "Synthetic WMDS source reference", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact", Surface: "light"}, Nodes: nodes})
	}
	source := func(kind string, x, y, w float64, fields map[string]any) map[string]any {
		n := map[string]any{"type": kind, "x": x, "y": y, "w": w}
		for k, v := range fields {
			n[k] = v
		}
		return n
	}
	for _, mark := range []string{"highlight", "underscore", "circle", "spark"} {
		slide("inline-"+mark, "Measured inline "+mark, node("inline."+mark, source("text", 57, 162, 558, map[string]any{"style": "heading", "ink": "display", "emphasis": mark, "text": "One [[shared calendar]] keeps decisions visible."})))
	}
	slide("headings", "Source heading anatomy", node("heading.number", source("numhead", 57, 144, 342, map[string]any{"n": "01", "text": "Define the scope", "numInk": "emphasis"})), node("heading.column", source("colhead", 489, 144, 342, map[string]any{"label": "Week 4", "title": "Operating baseline", "rule": "emphasis"})), node("heading.group", source("grouplabel", 57, 288, 774, map[string]any{"text": "Reference group"})))
	slide("textblocks", "Label-only and multiline content", node("block.label", source("textblock", 57, 144, 342, map[string]any{"label": "Outcome", "body": "An agreed scope with an accountable owner."})), node("block.full", source("textblock", 489, 144, 342, map[string]any{"label": "Approach", "title": "Review work early", "body": "The team reviews open decisions before period end."})))
	slide("bullets", "Keyed rich bullet leads", node("bullets.rich", source("bullets", 57, 144, 558, map[string]any{"items": []any{map[string]any{"lead": "Visible status.", "text": "Decisions have an owner."}, map[string]any{"text": "Review before period end.", "sub": []any{"Open exceptions", "Recorded decisions"}}}})))
	slide("numbers", "Strong numbers and numbered lists", node("numbers.strong", source("strongnum", 57, 144, 414, map[string]any{"numStyle": "stat-sm", "numInk": "emphasis", "items": []any{map[string]any{"title": "Set direction", "text": "Define the outcome."}, map[string]any{"title": "Name owners", "text": "Assign each decision."}}})), node("numbers.list", source("ol", 561, 144, 270, map[string]any{"items": []any{"Review scope", "Name owners", "Agree the next date"}})))
	slide("schedule", "Fixed schedule row capacity", node("schedule", source("schedule", 57, 144, 774, map[string]any{"keyW": 144, "rowHeight": 54, "keyInk": "emphasis", "items": []any{map[string]any{"k": "9:00–9:30", "t": "Welcome and goals"}, map[string]any{"k": "9:30–10:30", "t": "Review open decisions"}, map[string]any{"k": "10:30–11:00", "t": "Agree owners"}}})))
	slide("index", "Index row alignment", node("index", source("list", 57, 144, 774, map[string]any{"variant": "index", "rowHeight": 72, "items": []any{map[string]any{"n": "01", "title": "Current operating model", "page": "3"}, map[string]any{"n": "02", "title": "Phased roadmap", "page": "7"}, map[string]any{"n": "03", "title": "Team and next steps", "page": "12"}}})))
	slide("quote", "Editable quote anatomy", node("quote", source("pullquote", 57, 144, 774, map[string]any{"text": "Every decision now has an owner and a date.", "by": "Fictional operations lead"})))
	slide("photos", "Pinned cover crop and grayscale", node("photo.cover", source("imageframe", 57, 144, 342, map[string]any{"h": 234, "photo": "photo-warehouse", "focus": "62% 35%", "alt": "Canonical warehouse photo"})), node("photo.gray", source("imageframe", 489, 144, 342, map[string]any{"h": 234, "photo": "photo-team-meeting", "grayscale": true})))
	slide("squares", "Photo and bottom aligned stat squares", node("square.photo", map[string]any{"type": "square", "x": 57, "y": 144, "size": 234, "photo": "photo-technician", "focus": "50% 30%"}), node("square.stat", map[string]any{"type": "square", "x": 489, "y": 144, "size": 234, "surface": "inverse", "stat": "12 → 5", "label": "days to close"}))
	slide("thumbnail", "Native deliverable thumbnails", node("thumb.diagram", source("thumbnail", 57, 162, 162, map[string]any{"h": 90, "kind": "diagram", "stack": true, "caption": "Reference architecture"})), node("thumb.table", source("thumbnail", 273, 162, 162, map[string]any{"h": 90, "kind": "table", "caption": "Reference operating model"})), node("thumb.chart", source("thumbnail", 489, 162, 162, map[string]any{"h": 90, "kind": "chart", "caption": "Reference metrics"})), node("thumb.text", source("thumbnail", 705, 162, 126, map[string]any{"h": 90, "kind": "text", "caption": "Reference decisions"})))
	slide("artwork", "Canonical logo and editable text separation", node("art.logo", source("logo", 57, 162, 198, map[string]any{"variant": "pos"})), node("art.mark", source("mark", 345, 180, 126, map[string]any{"mark": "arrow-dashed"})), node("art.copy", source("text", 489, 162, 342, map[string]any{"style": "heading", "text": "Clear decisions move work forward."})))
	return d
}
