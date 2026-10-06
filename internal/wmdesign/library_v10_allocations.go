package wmdesign

import (
	"encoding/json"
	"fmt"
)

// Frozen source corrections reserve footer clearance for the final fork labels
// and account for the row-group rail. Copy, fonts and topology remain authored.
func applyV10AllocationRefinements(key string, slide *SlideSpec) error {
	if key == "roadmap-narrative/waves-criteria" || key == "roadmap-narrative/horizon-table" {
		return applyV10TableAllocation(key, slide)
	}
	field, before, after := "", 0., 0.
	switch key {
	case "road-fork/parallel", "road-fork/decision", "road-fork/foundation-then-waves":
		field, before, after = "h", 324, 306
	case "roadmap-narrative/objectives-tree":
		field, before, after = "delivery-width", 248, 292
	default:
		return nil
	}
	if slide == nil || len(slide.Nodes) == 0 || slide.Nodes[0].ID != "node01" || slide.Nodes[0].Scene == nil {
		return fmt.Errorf("library.v10_allocation_topology: %s", key)
	}
	node := slide.Nodes[0]
	obj, err := libraryObject(node.Scene.Node)
	if err != nil {
		return err
	}
	target := obj
	property := field
	if field == "delivery-width" {
		cols, ok := obj["cols"].([]any)
		if obj["type"] != "table" || !ok || len(cols) != 6 {
			return fmt.Errorf("library.v10_allocation_columns: %s", key)
		}
		target, ok = cols[1].(map[string]any)
		if !ok || target["k"] != "d" {
			return fmt.Errorf("library.v10_allocation_column: %s", key)
		}
		property = "w"
	} else if obj["type"] != "roadfork" {
		return fmt.Errorf("library.v10_allocation_type: %s", key)
	}
	v := intakeRepairNumber(target, property)
	if v != before && v != after {
		return fmt.Errorf("library.v10_allocation_geometry: %s", key)
	}
	target[property] = after
	scene := *node.Scene
	scene.Node, err = json.Marshal(obj)
	if err != nil {
		return err
	}
	marker := "wmds.v10.source-allocation." + field
	if !intakeRepairHasResolution(scene.Resolutions, marker) {
		scene.Resolutions = append(append([]string(nil), scene.Resolutions...), marker)
	}
	node.Scene = &scene
	nodes := append([]Node(nil), slide.Nodes...)
	nodes[0] = node
	slide.Nodes = nodes
	return nil
}

// These stock allocations are explicit corrections identified in native review.
// Every expected dimension is checked before publishing any change to the slide.
func applyV10TableAllocation(key string, slide *SlideSpec) error {
	if slide == nil || len(slide.Nodes) != 1 || slide.Nodes[0].ID != "node01" || slide.Nodes[0].Scene == nil {
		return fmt.Errorf("library.v10_table_allocation_topology: %s", key)
	}
	node := slide.Nodes[0]
	obj, err := libraryObject(node.Scene.Node)
	if err != nil {
		return err
	}
	if obj["type"] != "table" {
		return fmt.Errorf("library.v10_table_allocation_type: %s", key)
	}
	check := func(target map[string]any, field string, before, after float64) bool {
		value := intakeRepairNumber(target, field)
		return value == before || value == after
	}
	next := *slide
	if key == "roadmap-narrative/waves-criteria" {
		rows, ok := obj["rows"].([]any)
		if !ok || len(rows) != 4 || !check(obj, "y", 126, 162) || (slide.Frame.TitleLines != 1 && slide.Frame.TitleLines != 2) {
			return fmt.Errorf("library.v10_table_allocation_geometry: %s", key)
		}
		for _, raw := range rows {
			row, ok := raw.(map[string]any)
			if !ok || !check(row, "h", 72, 65) {
				return fmt.Errorf("library.v10_table_allocation_row: %s", key)
			}
		}
		next.Frame.TitleLines = 2
		obj["y"] = 162
		for _, raw := range rows {
			raw.(map[string]any)["h"] = float64(65)
		}
	} else {
		cols, ok := obj["cols"].([]any)
		if !ok || len(cols) != 6 {
			return fmt.Errorf("library.v10_table_allocation_columns: %s", key)
		}
		horizon, ok1 := cols[0].(map[string]any)
		investment, ok2 := cols[5].(map[string]any)
		if !ok1 || !ok2 || horizon["k"] != "hz" || investment["k"] != "v" || !check(horizon, "w", 108, 90) || !check(investment, "w", 72, 90) {
			return fmt.Errorf("library.v10_table_allocation_geometry: %s", key)
		}
		horizon["w"], investment["w"] = float64(90), float64(90)
	}
	scene := *node.Scene
	scene.Node, err = json.Marshal(obj)
	if err != nil {
		return err
	}
	marker := "wmds.v10.source-allocation." + key
	if !intakeRepairHasResolution(scene.Resolutions, marker) {
		scene.Resolutions = append(append([]string(nil), scene.Resolutions...), marker)
	}
	node.Scene = &scene
	next.Nodes = append([]Node(nil), slide.Nodes...)
	next.Nodes[0] = node
	*slide = next
	return nil
}
