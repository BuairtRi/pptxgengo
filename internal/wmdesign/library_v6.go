package wmdesign

import (
	"encoding/json"
	"fmt"
)

// These compositions differ from the accepted V5 source. Unchanged templates
// retain their exact V5 amendments; new heat maps use their authored topology.
func isV6HeatmapDelta(key string) bool {
	switch key {
	case "capability-heat/dense", "capability-heat/grouped", "capability-heat/grouped-split",
		"heat-notes/full", "heat-notes/full-continued", "heat-notes/grouped",
		"heat-notes/now-next-later", "heat-notes/summary", "heat-notes/split", "heat-notes/split-last",
		"heat-ref/overview", "heat-ref/overview-grouped", "heat-ref/detail-corner",
		"heat-ref/detail-continued", "heat-ref/locator-split", "heat-ref/cards-split":
		return true
	}
	return false
}

// Named candidate amendments correct source allocation defects without changing
// copy, ratings or font sizes. Each target accepts only its original or amended
// geometry, and records the change for native review.
func applyV6HeatmapRefinements(key string, slide *SlideSpec) error {
	type amendment struct {
		ordinal       int
		kind, field   string
		before, after float64
		marker        string
	}
	var changes []amendment
	count := 0
	switch key {
	case "capability-heat/grouped-split":
		count = 4
		changes = []amendment{
			{3, "bullets", "y", 162, 180, "grouped-split-bullets-clear-three-line-title"},
			{4, "callout", "y", 306, 288, "grouped-split-callout-clear-source-footer"},
		}
	case "heat-notes/now-next-later", "heat-notes/summary":
		count = 3
		changes = []amendment{{3, "text", "y", 434, 432, "priority-key-clear-source-footer"}}
	case "heat-notes/split":
		count = 3
		changes = []amendment{{2, "bullets", "y", 162, 180, "notes-split-bullets-clear-three-line-title"}}
	case "heat-notes/split-last":
		count = 3
		changes = []amendment{
			{1, "table", "y", 36, 54, "notes-continuation-label-inside-tall-column"},
			{2, "bullets", "y", 162, 180, "notes-split-bullets-clear-three-line-title"},
		}
	default:
		return nil
	}
	if slide == nil || len(slide.Nodes) != count {
		return fmt.Errorf("library.v6_amendment_topology: %s", key)
	}
	nodes := append([]Node(nil), slide.Nodes...)
	for _, change := range changes {
		n := &nodes[change.ordinal-1]
		if n.ID != fmt.Sprintf("node%02d", change.ordinal) || n.Scene == nil {
			return fmt.Errorf("library.v6_amendment_node: %s/%d", key, change.ordinal)
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		value := intakeRepairNumber(obj, change.field)
		if obj["type"] != change.kind || (value != change.before && value != change.after) {
			return fmt.Errorf("library.v6_amendment_geometry: %s/%d/%s", key, change.ordinal, change.field)
		}
		scene := *n.Scene
		obj[change.field] = change.after
		scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		marker := "wmds.v6." + change.marker
		if !intakeRepairHasResolution(scene.Resolutions, marker) {
			scene.Resolutions = append(append([]string(nil), scene.Resolutions...), marker)
		}
		n.Scene = &scene
	}
	slide.Nodes = nodes
	return nil
}
