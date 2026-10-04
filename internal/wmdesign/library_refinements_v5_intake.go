package wmdesign

import (
	"encoding/json"
	"fmt"
)

// Frozen v5 delta only; unchanged compositions inherit the qualified v4 path.
var v5DeltaNodeCounts = map[string]int{
	"interviews-access/board":                    2,
	"interviews-access/escalation-split":         8,
	"interviews-access/panel-right":              7,
	"interviews-coverage/dept-heat":              2,
	"interviews-coverage/log-page-1":             2,
	"interviews-coverage/log-page-2":             1,
	"interviews-coverage/overview":               6,
	"interviews-detail/dense":                    23,
	"interviews-detail/findings-first":           12,
	"interviews-detail/findings-table-split":     10,
	"interviews-detail/group-session":            10,
	"interviews-detail/headshot-profile":         11,
	"interviews-detail/portrait-column":          13,
	"interviews-detail/profile-split":            12,
	"interviews-detail/topics-table":             2,
	"interviews-detail/two-stakeholders":         12,
	"interviews-planned/table-nav":               1,
	"interviews-planned/timeline-split":          3,
	"interviews-planned/waves":                   3,
	"interviews-readout/compact":                 1,
	"interviews-readout/full":                    1,
	"interviews-readout/full-split":              3,
	"interviews-readout/name-title":              2,
	"interviews-readout/nav":                     5,
	"interviews-requests/panel-left":             10,
	"interviews-requests/tracker":                2,
	"interviews-summary/cards":                   6,
	"interviews-summary/group-table":             1,
	"interviews-summary/list-split":              5,
	"offers-case/proof-panel":                    30,
	"offers-case/split-left-narrow":              29,
	"offers-deliverable/split-tall-right-narrow": 16,
	"offers-deliverable/thumbnails":              18,
	"offers-gantt/eight-week":                    21,
	"offers-gantt/six-week-split":                22,
	"offers-onepager/approach-table-split":       8,
	"offers-onepager/classic":                    30,
	"offers-onepager/left-panel":                 14,
	"offers-onepager/nav":                        30,
	"offers-onepager/panel-right":                16,
	"offers-onepager/three-band":                 26,
	"offers-sku/catalog-split":                   8,
	"offers-sku/compare-table":                   18,
	"offers-sku/three-tiers":                     3,
	"offers-team/pod-inverse":                    35,
	"offers-team/run-rate":                       35,
	"understanding-dense/panel-right":            17,
	"understanding-drivers/central-ask":          5,
	"understanding-drivers/detail":               21,
	"understanding-drivers/split-nav":            5,
	"understanding-heard/quotes":                 12,
	"understanding-objectives/measures-split":    3,
	"understanding-pair/context":                 11,
	"understanding-pair/scope":                   6,
	"understanding-scope/in-out-nav":             7,
	"understanding-situation/bands":              7,
	"understanding-situation/evidence":           16,
	"understanding-situation/panel-right":        16,
	"understanding-situation/split":              17,
	"understanding-statement/evidence-grid":      2,
	"understanding-statement/inverse":            10,
	"understanding-statement/panel-left":         14,
	"understanding-statement/photo-split":        11,
	"understanding-statement/single":             10,
	"understanding-traceability/table":           1,
	"case-study/exhibit-split":                   6,
	"value-bridge/levers-split":                  38,
	"value-curve/break-even-split":               18,
}

// ApplyIncomingV5Repairs preserves source copy and pinned typography while
// reserving the longer v5 source notes and correcting named allocations.
func ApplyIncomingV5Repairs(key, revision string, slide *SlideSpec) error {
	if revision != LibraryRevisionV5 {
		return nil
	}
	count, ok := v5DeltaNodeCounts[key]
	if !ok {
		return nil
	}
	if slide == nil || len(slide.Nodes) != count {
		return fmt.Errorf("library.v5_delta_topology: %s", key)
	}
	copy := *slide
	if slide.LibraryChrome != nil {
		chrome := *slide.LibraryChrome
		copy.LibraryChrome = &chrome
	}
	copy.Nodes = append([]Node(nil), slide.Nodes...)
	for i, n := range slide.Nodes {
		if n.Scene == nil || n.Kind != "scene" || n.ID != fmt.Sprintf("node%02d", i+1) {
			return fmt.Errorf("library.v5_delta_scene_topology: %s/%d", key, i)
		}
		scene := *n.Scene
		scene.Resolutions = append([]string(nil), scene.Resolutions...)
		scene.Keys = intakeRepairKeys(scene.Keys)
		copy.Nodes[i].Scene = &scene
	}
	if key == "case-study/exhibit-split" || key == "value-curve/break-even-split" || key == "value-bridge/levers-split" {
		if err := applyLibraryRefinements(key, LibraryRevisionV4, &copy); err != nil {
			return err
		}
	}
	if copy.Source != "" && (copy.Frame.Split == "tall-left" || copy.Frame.Split == "tall-right") {
		if copy.Frame.SourceLines != 1 && copy.Frame.SourceLines != 2 {
			return fmt.Errorf("library.v5_delta_source_lines: %s", key)
		}
		if copy.Frame.SourceLines != 2 && copy.Nodes[0].Scene != nil {
			copy.Nodes[0].Scene.Resolutions = append(copy.Nodes[0].Scene.Resolutions, "wmds.v5.two-line-source-note-reservation")
		}
		copy.Frame.SourceLines = 2
	}
	if key == "interviews-access/escalation-split" {
		n := &copy.Nodes[0]
		if n.ID != "node01" || n.Scene == nil {
			return fmt.Errorf("library.v5_delta_edge_target")
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		if obj["type"] != "card" {
			return fmt.Errorf("library.v5_delta_edge_type")
		}
		if edge, ok := obj["edge"].(string); ok && edge == "emphasis" {
			obj["edge"] = map[string]any{"side": "left", "weight": 3, "ink": "emphasis"}
			n.Scene.Node, err = json.Marshal(obj)
			if err != nil {
				return err
			}
			n.Scene.Resolutions = append(n.Scene.Resolutions, "wmds.v5.escalation-invalid-string-edge-to-left-3pt-emphasis")
		} else if edge, ok := obj["edge"].(map[string]any); !ok || edge["side"] != "left" || intakeRepairNumber(edge, "weight") != 3 || edge["ink"] != "emphasis" {
			return fmt.Errorf("library.v5_delta_edge_contract")
		}
	}
	if key == "understanding-statement/inverse" {
		if copy.Frame.Surface != "inverse" || copy.LibraryChrome == nil || (copy.LibraryChrome.Emphasis != "highlight" && copy.LibraryChrome.Emphasis != "underscore") {
			return fmt.Errorf("library.v5_delta_inverse_emphasis_contract")
		}
		if copy.LibraryChrome.Emphasis == "highlight" {
			copy.LibraryChrome.Emphasis = "underscore"
			copy.Nodes[0].Scene.Resolutions = append(copy.Nodes[0].Scene.Resolutions, "wmds.v5.inverse-disallowed-highlight-to-supported-underscore")
		}
	}
	if key == "interviews-readout/full" {
		n := &copy.Nodes[0]
		if n.ID != "node01" || n.Scene == nil {
			return fmt.Errorf("library.v5_delta_unused_person_target")
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		cols, ok := obj["cols"].([]any)
		want := []string{"n", "r", "d", "dt", "f", "q", "s"}
		if obj["type"] != "table" || !ok || len(cols) != len(want) {
			return fmt.Errorf("library.v5_delta_unused_person_columns")
		}
		for i, raw := range cols {
			col, ok := raw.(map[string]any)
			if !ok || col["k"] != want[i] {
				return fmt.Errorf("library.v5_delta_unused_person_column_identity")
			}
		}
		rows, ok := obj["rows"].([]any)
		if !ok || len(rows) != 9 {
			return fmt.Errorf("library.v5_delta_unused_person_rows")
		}
		changed := false
		for _, raw := range rows {
			row, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("library.v5_delta_unused_person_row")
			}
			if raw, exists := row["p"]; exists {
				person, ok := raw.(map[string]any)
				if !ok || len(person) != 2 {
					return fmt.Errorf("library.v5_delta_unused_person_shape")
				}
				if _, ok := person["text"].(string); !ok {
					return fmt.Errorf("library.v5_delta_unused_person_text")
				}
				if _, ok := person["sub"].(string); !ok {
					return fmt.Errorf("library.v5_delta_unused_person_sub")
				}
				delete(row, "p")
				changed = true
			}
		}
		if changed {
			n.Scene.Node, err = json.Marshal(obj)
			if err != nil {
				return err
			}
			n.Scene.Resolutions = append(n.Scene.Resolutions, "wmds.v5.readout-unused-person-metadata-not-rendered")
		}
	}
	if key == "value-curve/break-even-split" {
		n := &copy.Nodes[1]
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		points, ok := obj["points"].([]any)
		if obj["type"] != "connector" || !ok || len(points) != 2 {
			return fmt.Errorf("library.v5_delta_break_even_connector")
		}
		first, ok1 := points[0].([]any)
		last, ok2 := points[1].([]any)
		if !ok1 || !ok2 || len(first) != 2 || len(last) != 2 {
			return fmt.Errorf("library.v5_delta_break_even_points")
		}
		number := func(v any) float64 { return intakeRepairNumber(map[string]any{"n": v}, "n") }
		// The fixed title/units leave plotTop84. A line plot reserves18pt
		// for categories, so the source plot base moves432→414. Preserve
		// the exact source ordinate within the new plot, not a guessed dot.
		const oldY = 315.7
		const nextY = 84 + (oldY-84)*(414-84)/(432-84)
		if number(first[0]) != 296.5 || number(last[0]) != 296.5 || number(last[1]) != 88 || (number(first[1]) != oldY && number(first[1]) != nextY) {
			return fmt.Errorf("library.v5_delta_break_even_anchor")
		}
		if number(first[1]) != nextY {
			first[1] = nextY
			n.Scene.Node, err = json.Marshal(obj)
			if err != nil {
				return err
			}
			n.Scene.Resolutions = append(n.Scene.Resolutions, "wmds.v5.break-even-annotation-normalized-to-396pt-chart")
		}
	}
	change := func(ordinal int, kind, field string, old, next float64, resolution string) error {
		id := fmt.Sprintf("node%02d", ordinal)
		for i := range copy.Nodes {
			n := &copy.Nodes[i]
			if n.ID != id {
				continue
			}
			if n.Scene == nil {
				return fmt.Errorf("library.v5_delta_scene: %s/%s", key, id)
			}
			obj, err := libraryObject(n.Scene.Node)
			if err != nil {
				return err
			}
			if obj["type"] != kind {
				return fmt.Errorf("library.v5_delta_type: %s/%s", key, id)
			}
			_, present := obj[field]
			value := intakeRepairNumber(obj, field)
			if !present || (value != old && value != next) {
				return fmt.Errorf("library.v5_delta_geometry: %s/%s/%s", key, id, field)
			}
			obj[field] = next
			n.Scene.Node, err = json.Marshal(obj)
			if err != nil {
				return err
			}
			if value != next {
				n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
			}
			return nil
		}
		return fmt.Errorf("library.v5_delta_missing_node: %s/%s", key, id)
	}
	type amendment struct {
		ordinal     int
		kind, field string
		old, next   float64
	}
	var amendments []amendment
	// These three native findings were caused by compressing lower sections
	// upward into preceding text. The source's spacing fits the supported slim
	// footer, whose extra18pt keeps every authored font and content item intact.
	useSlimFooter := func(resolution string) error {
		if copy.Frame.Footer != "compact" && copy.Frame.Footer != "slim" {
			return fmt.Errorf("library.v5_native_spacing_footer: %s", key)
		}
		if copy.Frame.Footer != "slim" {
			copy.Nodes[0].Scene.Resolutions = append(copy.Nodes[0].Scene.Resolutions, resolution)
		}
		copy.Frame.Footer = "slim"
		return nil
	}
	switch key {
	case "case-study/exhibit-split":
		amendments = append(amendments, amendment{1, "chart", "h", 414, 396})
		amendments = append(amendments, amendment{5, "grouplabel", "y", 342, 336}, amendment{6, "metric", "y", 360, 354})
	case "value-curve/break-even-split":
		amendments = append(amendments, amendment{1, "chart", "h", 414, 396}, amendment{16, "rule", "y", 408, 399}, amendment{17, "text", "y", 417, 408}, amendment{18, "text", "y", 414, 405})
	case "value-bridge/levers-split":
		for _, i := range []int{22, 23, 24} {
			amendments = append(amendments, amendment{i, "block", "y", 414, 396})
		}
		amendments = append(amendments, amendment{25, "text", "y", 144, 162}, amendment{26, "bullets", "y", 168, 186})
	case "interviews-access/board":
		if copy.Frame.Footer != "compact" && copy.Frame.Footer != "slim" {
			return fmt.Errorf("library.v5_delta_board_footer")
		}
		if copy.Frame.Footer != "slim" {
			copy.Nodes[0].Scene.Resolutions = append(copy.Nodes[0].Scene.Resolutions, "wmds.v5.access-board-two-line-rows-and-caption-slim-footer")
		}
		copy.Frame.Footer = "slim"
		amendments = append(amendments, amendment{1, "table", "rowH", 54, 56}, amendment{2, "text", "y", 432, 441})
	case "interviews-access/escalation-split":
		if copy.Frame.TitleLines != 1 && copy.Frame.TitleLines != 2 {
			return fmt.Errorf("library.v5_delta_escalation_title_lines")
		}
		copy.Frame.TitleLines = 2
		for _, i := range []int{1, 2} {
			amendments = append(amendments, amendment{i, "card", "y", 126, 162}, amendment{i, "card", "h", 144, 132})
		}
		for _, i := range []int{3, 4} {
			amendments = append(amendments, amendment{i, "card", "y", 288, 306})
		}
	case "interviews-readout/nav":
		if copy.Frame.Footer != "compact" && copy.Frame.Footer != "slim" {
			return fmt.Errorf("library.v5_delta_readout_footer")
		}
		if copy.Frame.Footer != "slim" {
			copy.Nodes[0].Scene.Resolutions = append(copy.Nodes[0].Scene.Resolutions, "wmds.v5.readout-six-two-line-rows-slim-footer")
		}
		copy.Frame.Footer = "slim"
		amendments = append(amendments, amendment{5, "table", "y", 216, 204}, amendment{5, "table", "rowH", 34, 38})
	case "interviews-detail/portrait-column":
		amendments = append(amendments, amendment{12, "textblock", "y", 390, 381}, amendment{13, "bullets", "y", 414, 405})
	case "interviews-detail/findings-first":
		amendments = append(amendments, amendment{12, "table", "y", 318, 312})
	case "interviews-detail/profile-split":
		amendments = append(amendments, amendment{7, "table", "rowH", 52, 51})
	case "interviews-planned/timeline-split":
		amendments = append(amendments, amendment{3, "callout", "y", 324, 318})
	case "offers-onepager/panel-right":
		if copy.Frame.Rail != "right" {
			return fmt.Errorf("library.v5_native_spacing_panel")
		}
		if err := useSlimFooter("wmds.v5.offers-panel-source-section-spacing-slim-footer"); err != nil {
			return err
		}
		amendments = append(amendments, amendment{12, "rule", "y", 345, 345}, amendment{13, "text", "y", 357, 357}, amendment{14, "text", "y", 375, 375}, amendment{15, "rule", "y", 417, 417}, amendment{16, "text", "y", 426, 426})
	case "offers-onepager/approach-table-split":
		amendments = append(amendments, amendment{1, "table", "rowH", 78, 72})
	case "offers-case/split-left-narrow":
		amendments = append(amendments, amendment{11, "rule", "y", 369, 363}, amendment{12, "text", "y", 375, 369}, amendment{13, "bullets", "y", 393, 387})
		amendments = append(amendments, amendment{27, "text", "w", 78, 108}, amendment{28, "text", "x", 603, 633}, amendment{28, "text", "w", 84, 90}, amendment{29, "text", "x", 693, 741}, amendment{29, "text", "w", 198, 150})
	case "offers-gantt/six-week-split":
		if copy.Frame.Split != "tall-right" {
			return fmt.Errorf("library.v5_native_spacing_gantt_split")
		}
		if err := useSlimFooter("wmds.v5.six-week-source-rule-clearance-slim-footer"); err != nil {
			return err
		}
		for _, i := range []int{8, 11} {
			amendments = append(amendments, amendment{i, "rule", "y", 354, 354})
		}
		for _, i := range []int{9, 12} {
			amendments = append(amendments, amendment{i, "text", "y", 360, 360})
		}
		for _, i := range []int{10, 13} {
			amendments = append(amendments, amendment{i, "bullets", "y", 378, 378})
		}
	case "offers-sku/catalog-split":
		if copy.Frame.Split != "tall-left" {
			return fmt.Errorf("library.v5_native_spacing_catalog_split")
		}
		if err := useSlimFooter("wmds.v5.offer-catalog-source-section-clearance-slim-footer"); err != nil {
			return err
		}
		amendments = append(amendments, amendment{1, "table", "rowH", 63, 63}, amendment{6, "text", "y", 354, 354}, amendment{7, "bullets", "y", 372, 372}, amendment{8, "text", "y", 414, 414})
	case "offers-sku/compare-table":
		amendments = append(amendments, amendment{5, "text", "y", 378, 369}, amendment{6, "bullets", "y", 396, 387})
		amendments = append(amendments, amendment{18, "text", "y", 417, 414})
	case "understanding-objectives/measures-split":
		amendments = append(amendments, amendment{2, "callout", "y", 288, 279}, amendment{3, "table", "rowH", 54, 52})
	case "understanding-situation/split":
		amendments = append(amendments, amendment{2, "callout", "y", 288, 279}, amendment{13, "block", "y", 324, 306})
		for _, i := range []int{14, 16, 17} {
			kind := "text"
			if i == 16 {
				kind = "bullets"
			}
			if i == 17 {
				kind = "metric"
			}
			amendments = append(amendments, amendment{i, kind, "y", 342, 324})
		}
		amendments = append(amendments, amendment{15, "text", "y", 366, 348})
	case "understanding-scope/in-out-nav":
		for _, i := range []int{1, 2, 3} {
			amendments = append(amendments, amendment{i, "card", "h", 216, 222})
		}
	case "understanding-situation/panel-right":
		amendments = append(amendments, amendment{14, "rule", "y", 396, 387}, amendment{15, "text", "y", 414, 405}, amendment{16, "text", "y", 432, 423})
	case "interviews-planned/waves":
		if copy.Frame.TitleLines != 1 && copy.Frame.TitleLines != 2 {
			return fmt.Errorf("library.v5_delta_waves_title_lines")
		}
		copy.Frame.TitleLines = 2
		for _, i := range []int{1, 2, 3} {
			amendments = append(amendments, amendment{i, "card", "y", 126, 162}, amendment{i, "card", "h", 324, 288})
		}
	}
	for _, a := range amendments {
		if err := change(a.ordinal, a.kind, a.field, a.old, a.next, "wmds.v5."+key+".native-allocation"); err != nil {
			return err
		}
	}
	*slide = copy
	return nil
}
