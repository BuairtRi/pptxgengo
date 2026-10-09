package wmdesign

import (
	"encoding/json"
	"math"
	"path/filepath"
	"testing"
)

func TestSceneSourceGeometryPreservesAnchorsAndResize(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		args       map[string]any
		measured   Rect
	}{
		{"beforeafter-heading-ink", "beforeafter", map[string]any{"x": 75., "y": 378., "w": 810., "left": "Before", "right": "After", "rows": []any{[]any{"Manual", "Connected"}}}, Rect{75, 387, 810, 66}},
		{"quadrant-ink-offset", "chart", map[string]any{"x": 411., "y": 110., "w": 300., "h": 240., "chart": "quadrants"}, Rect{418, 117, 286, 226}},
		{"time-axis-external-labels", "timeaxis", map[string]any{"x": 129., "y": 342., "w": 702., "periods": []any{"Week 1", "Week 2"}, "milestones": []any{map[string]any{"at": 0., "label": "Start"}}}, Rect{57, 312, 774, 66}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := map[string]float64{}
			for _, k := range []string{"x", "y", "w", "h"} {
				if v, ok := tc.args[k].(float64); ok {
					original[k] = v
				}
			}
			if e := CaptureSceneSourceGeometry(tc.kind, tc.args, tc.measured); e != nil {
				t.Fatal(e)
			}
			for _, k := range []string{"x", "y", "w", "h"} {
				delete(tc.args, k)
			}
			for _, scale := range []float64{1, 1.5} {
				allocation := Rect{tc.measured.X + 13, tc.measured.Y + 17, tc.measured.W * scale, tc.measured.H * scale}
				raw, e := ComposeSceneNode(tc.kind, tc.args, allocation)
				if e != nil {
					t.Fatal(e)
				}
				var got map[string]any
				if e = json.Unmarshal(raw, &got); e != nil {
					t.Fatal(e)
				}
				if _, leaked := got[SceneSourceGeometryArgument]; leaked {
					t.Fatal("adapter metadata leaked into source planner")
				}
				for k, v := range original {
					expected := v * scale
					if k == "x" {
						expected = allocation.X + (v-tc.measured.X)*scale
					}
					if k == "y" {
						expected = allocation.Y + (v-tc.measured.Y)*scale
					}
					if math.Abs(got[k].(float64)-expected) > 1e-9 {
						t.Fatalf("%s anchor %s changed: %v != %v", tc.kind, k, got[k], expected)
					}
				}
				if tc.kind == "beforeafter" || tc.kind == "timeaxis" {
					if _, exists := got["h"]; exists {
						t.Fatal("height added to strict no-height source")
					}
				}
			}
		})
	}
}
func TestSceneSourceGeometryCardrowCountsPreserveGapAndTotalExtent(t *testing.T) {
	args := map[string]any{"x": 75., "y": 140., "w": 249., "h": 180., "gap": 18., "card": map[string]any{"surface": "subtle"}, "items": []any{map[string]any{"title": "One"}, map[string]any{"title": "Two"}, map[string]any{"title": "Three"}}}
	allocation := Rect{75, 140, 783, 180}
	if e := CaptureSceneSourceGeometry("cardrow", args, allocation); e != nil {
		t.Fatal(e)
	}
	for _, k := range []string{"x", "y", "w", "h"} {
		delete(args, k)
	}
	for _, count := range []int{1, 2, 3, 5, 6} {
		items := []any{}
		for i := 0; i < count; i++ {
			items = append(items, map[string]any{"title": "Illustrative item"})
		}
		args["items"] = items
		raw, e := ComposeSceneNode("cardrow", args, allocation)
		if e != nil {
			t.Fatal(e)
		}
		var got map[string]any
		json.Unmarshal(raw, &got)
		width := got["w"].(float64)
		gap := got["gap"].(float64)
		if gap != 18 || math.Abs(width*float64(count)+gap*float64(count-1)-allocation.W) > 1e-9 {
			t.Fatalf("row %d escaped: item=%g gap=%g", count, width, gap)
		}
		if count == 3 && width != 249 {
			t.Fatal("original item width changed", width)
		}
	}
	args["gap"] = 1000.
	if _, e := ComposeSceneNode("cardrow", args, allocation); e == nil {
		t.Fatal("gap consumed allocation")
	}
}

func TestSceneSourceGeometryVennTopologyKeepsOriginalAndReflowsChangedCount(t *testing.T) {
	args := map[string]any{"x": 345., "y": 36., "w": 558., "h": 414., "sets": []any{map[string]any{"label": "A"}, map[string]any{"label": "B"}, map[string]any{"label": "C"}}}
	allocation := Rect{354, 36.5, 540, 413.5}
	if err := CaptureSceneSourceGeometry("venn", args, allocation); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"x", "y", "w", "h"} {
		delete(args, field)
	}
	decode := func() map[string]any {
		raw, err := ComposeSceneNode("venn", args, allocation)
		if err != nil {
			t.Fatal(err)
		}
		var result map[string]any
		if err = json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	original := decode()
	if original["x"] != 345. || original["y"] != 36. || original["w"] != 558. || original["h"] != 414. {
		t.Fatal("unchanged topology moved source geometry", original)
	}
	args["sets"] = args["sets"].([]any)[:2]
	reflow := decode()
	if reflow["x"] != allocation.X+.5 || reflow["y"] != allocation.Y+.5 || reflow["w"] != allocation.W-1 || reflow["h"] != allocation.H-1 {
		t.Fatal("changed topology did not reserve its stroke envelope", reflow)
	}
	args["sets"] = []any{map[string]any{"label": "A"}, map[string]any{"label": "B"}, map[string]any{"label": "C"}}
	if restored := decode(); restored["x"] != original["x"] || restored["w"] != original["w"] {
		t.Fatal("restored topology did not retain original geometry")
	}
}
func TestSceneSourceGeometryStrictMetadataAndPhasesHeadroom(t *testing.T) {
	for _, metadata := range []any{
		map[string]any{"schema": "wrong", "x_fraction": 0., "y_fraction": 0.},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "extra": true},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "width_fraction": -1.},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": nil, "y_fraction": 0.},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "width_fraction": nil},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "venn_set_count": 3},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "venn_set_count": nil},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "venn_set_count": 5},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "venn_area_height_fraction": 1},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "venn_set_count": 2, "venn_area_height_fraction": 1},
		map[string]any{"schema": SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "venn_set_count": 3, "venn_area_height_fraction": nil},
		map[string]any{"schema": SceneSourceGeometrySchema},
	} {
		if _, e := ComposeSceneNode("block", map[string]any{SceneSourceGeometryArgument: metadata}, Rect{0, 0, 100, 100}); e == nil {
			t.Fatal("invalid metadata accepted", metadata)
		}
	}
	args := map[string]any{"phases": []any{map[string]any{"title": "Discover"}, map[string]any{"title": "Deliver"}}, "current": 0}
	raw, e := ComposeSceneNode("phases", args, Rect{100, 100, 600, 100})
	if e != nil {
		t.Fatal(e)
	}
	var got map[string]any
	json.Unmarshal(raw, &got)
	if got["y"] != 124. {
		t.Fatal("current marker headroom not reserved", got)
	}
	args["x"], args["y"], args["w"] = 100., 124., 600.
	if e = CaptureSceneSourceGeometry("phases", args, Rect{100, 100, 600, 100}); e != nil {
		t.Fatal(e)
	}
	delete(args, "x")
	delete(args, "y")
	delete(args, "w")
	raw, e = ComposeSceneNode("phases", args, Rect{100, 100, 600, 100})
	if e != nil {
		t.Fatal(e)
	}
	json.Unmarshal(raw, &got)
	if got["y"] != 124. {
		t.Fatal("captured headroom applied twice", got)
	}
}

func TestSceneSourceGeometryActualV11CatalogRenderFidelity(t *testing.T) {
	root, e := filepath.Abs("../../library/wm-design-system/v11")
	if e != nil {
		t.Fatal(e)
	}
	catalog, e := LibraryCatalog(root, "")
	if e != nil {
		t.Fatal(e)
	}
	for _, key := range []string{"architecture/app-ecosystem", "quadrant/subtle", "capacity-heat/sprint", "runbook/cutover-timeline", "venn/two-text", "venn/three-text", "venn/four-callout"} {
		t.Run(key, func(t *testing.T) {
			family := ""
			for _, entry := range catalog {
				if entry.Key == key {
					family = entry.Family
					break
				}
			}
			if family == "" {
				t.Fatal("missing source", key)
			}
			examples, e := LibraryReference(root, "", family, 2026)
			if e != nil {
				t.Fatal(e)
			}
			found := false
			for _, slide := range examples.Slides {
				if slide.Template == key {
					examples.Slides = []BoundSlide{slide}
					found = true
					break
				}
			}
			if !found {
				t.Fatal("missing specimen", key)
			}
			original, _, e := BindTemplates(root, "", examples)
			if e != nil {
				t.Fatal(e)
			}
			if key == "runbook/cutover-timeline" {
				// The catalog's unrelated underscore is not a distributed original. Isolate
				// its real time-axis source node without substituting or expanding artwork.
				nodes := []Node{}
				for _, node := range original.Slides[0].Nodes {
					if node.Scene == nil {
						continue
					}
					var tag struct {
						Type string `json:"type"`
					}
					if e = json.Unmarshal(node.Scene.Node, &tag); e != nil {
						t.Fatal(e)
					}
					if tag.Type == "timeaxis" {
						nodes = append(nodes, node)
					}
				}
				if len(nodes) == 0 {
					t.Fatal("actual time axis missing")
				}
				original = Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "catalog-timeaxis", Frame: FrameRequest{NoHeader: true}, Nodes: nodes}}}
			}
			_, before, e := BuildWithEngine(root, "", original, CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			payload, _ := json.Marshal(original)
			var adapted Document
			if e = json.Unmarshal(payload, &adapted); e != nil {
				t.Fatal(e)
			}
			bounds := map[string]Rect{}
			for _, scene := range before.Slides[0].Scenes {
				bounds[scene.ID] = scene.Bounds
			}
			count := 0
			for i, node := range adapted.Slides[0].Nodes {
				if node.Kind != "scene" || node.Scene == nil {
					continue
				}
				var args map[string]any
				if e = json.Unmarshal(node.Scene.Node, &args); e != nil {
					t.Fatal(e)
				}
				kind, _ := args["type"].(string)
				if kind == "connector" {
					continue
				}
				b, exists := bounds[node.ID]
				if !exists {
					t.Fatal("missing measured source", node.ID)
				}
				var sourceZone []Rect
				if kind == "venn" {
					zone, err := SceneSourcePlanningZone(args, before.Slides[0].Frame)
					if err != nil {
						t.Fatal(err)
					}
					sourceZone = append(sourceZone, zone)
				}
				if e = CaptureSceneSourceGeometry(kind, args, b, sourceZone...); e != nil {
					t.Fatal(node.ID, e)
				}
				for _, field := range []string{"type", "id", "x", "y", "w", "h"} {
					delete(args, field)
				}
				raw, e := ComposeSceneNode(kind, args, b)
				if e != nil {
					t.Fatal(node.ID, e)
				}
				adapted.Slides[0].Nodes[i].Scene.Node = raw
				adapted.Slides[0].Nodes[i].Scene.Allocation = &b
				count++
			}
			if count == 0 {
				t.Fatal("no actual scene geometry checked")
			}
			_, after, e := BuildWithEngine(root, "", adapted, CandidateEngine)
			if e != nil {
				t.Fatal("adapted actual source", e)
			}
			a, b := before.Slides[0], after.Slides[0]
			if len(a.Shapes) != len(b.Shapes) || len(a.Texts) != len(b.Texts) {
				t.Fatal("native objects changed", len(a.Shapes), len(b.Shapes), len(a.Texts), len(b.Texts))
			}
			near := func(a, b Rect) bool {
				return math.Abs(a.X-b.X)+math.Abs(a.Y-b.Y)+math.Abs(a.W-b.W)+math.Abs(a.H-b.H) < 1e-7
			}
			for i, s := range a.Shapes {
				if s.ID != b.Shapes[i].ID || !near(s.Rect, b.Shapes[i].Rect) {
					t.Fatalf("native shape moved: %s %+v %+v", s.ID, s.Rect, b.Shapes[i].Rect)
				}
			}
			for i, text := range a.Texts {
				if text.ID != b.Texts[i].ID || !near(text.Rect, b.Texts[i].Rect) {
					t.Fatalf("native text moved: %s %+v %+v", text.ID, text.Rect, b.Texts[i].Rect)
				}
			}
		})
	}
}
