package wmdesign

import (
	"encoding/json"
	"fmt"
)

// Multiday amendments fix source table allocations without changing the source
// bytes, copy, typography or total table height. Native acceptance is separate.
func applyV9MultidayWorkshopRefinements(key string, slide *SlideSpec) error {
	type amendment struct {
		ordinal       int
		field         string
		before, after float64
	}
	var changes []amendment
	switch key {
	case "workshop-multiday/three-day-grid":
		// Main text and its subline need 41pt. Rebalance two fixed rows by 6pt.
		changes = []amendment{{4, "midday-height", 36, 42}, {4, "output-height", 54, 48}}
	case "workshop-multiday/two-day-onsite-remote":
		// 558pt minus the 24pt group rail and its 6pt gutter leaves 528pt.
		changes = []amendment{{1, "session-width", 174, 204}}
	default:
		return nil
	}
	if slide == nil {
		return fmt.Errorf("library.v9_multiday_amendment_topology: %s", key)
	}
	nodes := append([]Node(nil), slide.Nodes...)
	for _, change := range changes {
		if change.ordinal > len(nodes) {
			return fmt.Errorf("library.v9_multiday_amendment_topology: %s", key)
		}
		n := &nodes[change.ordinal-1]
		if n.ID != fmt.Sprintf("node%02d", change.ordinal) || n.Scene == nil {
			return fmt.Errorf("library.v9_multiday_amendment_node: %s", key)
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		var target map[string]any
		field := "h"
		if change.field == "session-width" {
			cols, ok := obj["cols"].([]any)
			if !ok || len(cols) != 4 || intakeRepairNumber(obj, "w") != 558 || intakeRepairNumber(obj, "groupW") != 24 {
				return fmt.Errorf("library.v9_multiday_amendment_columns: %s", key)
			}
			target, ok = cols[1].(map[string]any)
			if !ok || target["k"] != "s" {
				return fmt.Errorf("library.v9_multiday_amendment_column: %s", key)
			}
			field = "w"
		} else {
			rows, ok := obj["rows"].([]any)
			if !ok || len(rows) != 5 || intakeRepairNumber(obj, "w") != 846 || intakeRepairNumber(obj, "rowH") != 60 {
				return fmt.Errorf("library.v9_multiday_amendment_rows: %s", key)
			}
			row := 1
			if change.field == "output-height" {
				row = 4
			}
			target, ok = rows[row].(map[string]any)
			if !ok {
				return fmt.Errorf("library.v9_multiday_amendment_row: %s", key)
			}
		}
		value := intakeRepairNumber(target, field)
		if obj["type"] != "table" || (value != change.before && value != change.after) {
			return fmt.Errorf("library.v9_multiday_amendment_geometry: %s/%d/%s", key, change.ordinal, change.field)
		}
		target[field] = change.after
		scene := *n.Scene
		scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		marker := fmt.Sprintf("wmds.v9.multiday-allocation.%d.%s", change.ordinal, change.field)
		if !intakeRepairHasResolution(scene.Resolutions, marker) {
			scene.Resolutions = append(append([]string(nil), scene.Resolutions...), marker)
		}
		n.Scene = &scene
	}
	slide.Nodes = nodes
	return nil
}
