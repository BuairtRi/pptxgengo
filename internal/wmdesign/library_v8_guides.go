package wmdesign

import (
	"encoding/json"
	"fmt"
)

// Guide amendments repair two frozen allocation defects. The source examples,
// typography and table values are retained; these are fixed template geometry.
func applyV8WorkshopGuideRefinements(key string, slide *SlideSpec) error {
	type amendment struct {
		ordinal       int
		kind, field   string
		before, after float64
	}
	var changes []amendment
	switch key {
	case "workshop-guide/question-bank":
		// A 24pt group rail plus its 6pt gutter leaves 816pt for columns.
		// Reserve one line for the longest question and rationale at 12/18pt.
		changes = []amendment{{1, "table", "question-width", 342, 354}, {1, "table", "why-width", 240, 246}, {1, "table", "listen-width", 228, 216}}
	case "workshop-guide/question-ladder":
		changes = []amendment{{5, "grouplabel", "y", 162, 180}, {6, "rule", "y", 180, 198}, {7, "bullets", "y", 198, 216}}
	default:
		return nil
	}
	if slide == nil {
		return fmt.Errorf("library.v8_guide_amendment_topology: %s", key)
	}
	nodes := append([]Node(nil), slide.Nodes...)
	for _, change := range changes {
		if change.ordinal > len(nodes) {
			return fmt.Errorf("library.v8_guide_amendment_topology: %s", key)
		}
		n := &nodes[change.ordinal-1]
		if n.ID != fmt.Sprintf("node%02d", change.ordinal) || n.Scene == nil {
			return fmt.Errorf("library.v8_guide_amendment_node: %s", key)
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		target, field := obj, change.field
		if field == "question-width" || field == "why-width" || field == "listen-width" {
			column := map[string]int{"question-width": 0, "why-width": 1, "listen-width": 2}[field]
			columnKey := []string{"q", "w", "l"}[column]
			cols, ok := obj["cols"].([]any)
			if !ok || len(cols) != 3 {
				return fmt.Errorf("library.v8_guide_amendment_columns: %s", key)
			}
			target, ok = cols[column].(map[string]any)
			if !ok || target["k"] != columnKey || intakeRepairNumber(obj, "w") != 846 || intakeRepairNumber(obj, "groupW") != 24 {
				return fmt.Errorf("library.v8_guide_amendment_column: %s", key)
			}
			field = "w"
		}
		value := intakeRepairNumber(target, field)
		if obj["type"] != change.kind || (value != change.before && value != change.after) {
			return fmt.Errorf("library.v8_guide_amendment_geometry: %s/%d/%s", key, change.ordinal, change.field)
		}
		target[field] = change.after
		scene := *n.Scene
		scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		marker := fmt.Sprintf("wmds.v8.guide-allocation.%d.%s", change.ordinal, change.field)
		if !intakeRepairHasResolution(scene.Resolutions, marker) {
			scene.Resolutions = append(append([]string(nil), scene.Resolutions...), marker)
		}
		n.Scene = &scene
	}
	slide.Nodes = nodes
	return nil
}
