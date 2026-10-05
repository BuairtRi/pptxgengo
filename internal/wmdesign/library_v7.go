package wmdesign

import (
	"encoding/json"
	"fmt"
	"strings"
)

// V7 candidate amendments repair allocation defects while preserving the frozen
// source bytes, copy, typography and semantic values. Native acceptance is separate.
func applyV7WorkshopRefinements(key string, slide *SlideSpec) error {
	type amendment struct {
		ordinal       int
		kind, field   string
		before, after float64
	}
	var changes []amendment
	switch key {
	case "workshop-overview/classic":
		changes = []amendment{{7, "schedule", "rowHeight", 21, 22}}
	case "workshop-overview/dense-leave-with":
		changes = []amendment{{4, "grouplabel", "y", 234, 216}, {5, "rule", "y", 252, 234}, {6, "schedule", "y", 270, 252}, {6, "schedule", "rowHeight", 18, 22}}
	case "workshop-readout/full":
		changes = []amendment{{13, "table", "cols.2.w", 60, 78}, {13, "table", "cols.3.w", 132, 114}}
	case "workshop-readout/split-learnings":
		changes = []amendment{{10, "grouplabel", "y", 198, 180}, {11, "rule", "y", 216, 198}, {12, "bullets", "y", 234, 216}, {13, "grouplabel", "y", 198, 180}, {14, "rule", "y", 216, 198}, {15, "bullets", "y", 234, 216}, {16, "grouplabel", "y", 324, 288}, {17, "rule", "y", 342, 306}, {18, "table", "y", 360, 324}, {18, "table", "cols.2.w", 60, 78}, {18, "table", "cols.3.w", 120, 102}}
	case "workshop-series/phase-groups":
		changes = []amendment{{1, "table", "cols.5.w", 90, 108}}
	case "workshop-series/table":
		changes = []amendment{{1, "table", "cols.4.w", 174, 156}, {1, "table", "cols.5.w", 90, 108}}
	default:
		return nil
	}
	if slide == nil {
		return fmt.Errorf("library.v7_amendment_topology: %s", key)
	}
	nodes := append([]Node(nil), slide.Nodes...)
	for _, change := range changes {
		if change.ordinal > len(nodes) {
			return fmt.Errorf("library.v7_amendment_topology: %s", key)
		}
		n := &nodes[change.ordinal-1]
		if n.ID != fmt.Sprintf("node%02d", change.ordinal) || n.Scene == nil {
			return fmt.Errorf("library.v7_amendment_node: %s", key)
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		target := obj
		field := change.field
		if strings.HasPrefix(field, "cols.") {
			cols, ok := obj["cols"].([]any)
			index := int(field[5] - '0')
			expectedCount := 4
			expectedKey := map[int]string{2: "d", 3: "s", 4: "o", 5: "s"}[index]
			if index >= 4 {
				expectedCount = 6
			}
			if !ok || len(cols) != expectedCount {
				return fmt.Errorf("library.v7_amendment_columns: %s", key)
			}
			target, ok = cols[index].(map[string]any)
			if !ok || target["k"] != expectedKey {
				return fmt.Errorf("library.v7_amendment_column: %s", key)
			}
			field = "w"
		}
		value := intakeRepairNumber(target, field)
		if obj["type"] != change.kind || (value != change.before && value != change.after) {
			return fmt.Errorf("library.v7_amendment_geometry: %s/%d/%s", key, change.ordinal, change.field)
		}
		target[field] = change.after
		scene := *n.Scene
		scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		marker := fmt.Sprintf("wmds.v7.workshop-allocation.%d.%s", change.ordinal, change.field)
		if !intakeRepairHasResolution(scene.Resolutions, marker) {
			scene.Resolutions = append(append([]string(nil), scene.Resolutions...), marker)
		}
		n.Scene = &scene
	}
	if key == "workshop-readout/split-learnings" {
		if slide.Frame.SourceLines != 1 && slide.Frame.SourceLines != 2 {
			return fmt.Errorf("library.v7_amendment_source_lines")
		}
		slide.Frame.SourceLines = 2
	}
	slide.Nodes = nodes
	return nil
}
