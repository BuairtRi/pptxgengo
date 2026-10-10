package wmdesign

import (
	"encoding/json"
	"fmt"
)

// These explicit native allocations account for source table slack and chip
// padding. Frozen JSON, authored copy and earlier revisions remain intact.
func applyV12AllocationRefinements(key string, slide *SlideSpec, templateRevision ...int) error {
	if key == "toc/levels-split" {
		if slide == nil || len(slide.Nodes) != 27 || slide.Nodes[26].Scene == nil {
			return fmt.Errorf("library.v12_allocation_topology: %s", key)
		}
		node := slide.Nodes[26]
		scene := *node.Scene
		obj, err := libraryObject(scene.Node)
		if err != nil {
			return err
		}
		if obj["type"] != "rule" || (intakeRepairNumber(obj, "y") != 468 && intakeRepairNumber(obj, "y") != 467.25) {
			return fmt.Errorf("library.v12_allocation_geometry: %s", key)
		}
		obj["y"] = 467.25
		scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		scene.Resolutions = append(append([]string(nil), scene.Resolutions...), "wmds.v12.source-allocation.contents-footer-rule")
		node.Scene = &scene
		slide.Nodes = append([]Node(nil), slide.Nodes...)
		slide.Nodes[26] = node
		return nil
	}
	if key == "plan/gantt" && len(templateRevision) > 0 && templateRevision[0] == 2 {
		node := slide.Nodes[0]
		scene := *node.Scene
		obj, err := libraryObject(scene.Node)
		if err != nil {
			return err
		}
		obj["groupLabelInset"] = 6.
		scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		scene.Resolutions = append(append([]string(nil), scene.Resolutions...), "wmds.v12.source-allocation.single-lane-group-label")
		node.Scene = &scene
		slide.Nodes = append([]Node(nil), slide.Nodes...)
		slide.Nodes[0] = node
		return nil
	}
	if key == "milestone-timeline/gantt-milestone-panel" {
		if slide == nil || len(slide.Nodes) == 0 || slide.Nodes[0].Scene == nil {
			return fmt.Errorf("library.v12_allocation_topology: %s", key)
		}
		node := slide.Nodes[0]
		obj, err := libraryObject(node.Scene.Node)
		if err != nil {
			return err
		}
		if obj["type"] != "gantt" || (intakeRepairNumber(obj, "w") != 558 && intakeRepairNumber(obj, "w") != 551) {
			return fmt.Errorf("library.v12_allocation_geometry: %s", key)
		}
		// Reserve the final milestone's 7pt half-width inside the split frame;
		// periods, dates, lanes and task intervals retain their authored values.
		obj["w"] = 551.
		scene := *node.Scene
		scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		scene.Resolutions = append(append([]string(nil), scene.Resolutions...), "wmds.v12.source-allocation.gantt-endpoint-clearance")
		node.Scene = &scene
		slide.Nodes = append([]Node(nil), slide.Nodes...)
		slide.Nodes[0] = node
		if len(slide.Nodes) < 2 || slide.Nodes[1].Scene == nil {
			return fmt.Errorf("library.v12_allocation_topology: %s", key)
		}
		label := slide.Nodes[1]
		labelObj, err := libraryObject(label.Scene.Node)
		if err != nil {
			return err
		}
		if intakeRepairNumber(labelObj, "y") != 143 && intakeRepairNumber(labelObj, "y") != 144 {
			return fmt.Errorf("library.v12_allocation_geometry: %s", key)
		}
		labelObj["y"] = 144.
		labelScene := *label.Scene
		labelScene.Node, err = json.Marshal(labelObj)
		if err != nil {
			return err
		}
		labelScene.Resolutions = append(append([]string(nil), labelScene.Resolutions...), "wmds.v12.source-allocation.gantt-panel-header-clearance")
		label.Scene = &labelScene
		slide.Nodes[1] = label
		return nil
	}
	ordinal, before, after := 0, 0., 0.
	switch key {
	case "backlog/epic-hierarchy":
		ordinal, before, after = 7, 354, 378
	case "backlog/epic-overview":
		ordinal, before, after = 1, 186, 210
	case "backlog/wsjf-scoring":
		ordinal, before, after = 1, 222, 210
	case "roadmap-strategy/cascade-split":
		ordinal, before, after = 7, 150, 156
	default:
		return nil
	}
	if slide == nil || len(slide.Nodes) < ordinal || slide.Nodes[ordinal-1].Scene == nil {
		return fmt.Errorf("library.v12_allocation_topology: %s", key)
	}
	node := slide.Nodes[ordinal-1]
	obj, err := libraryObject(node.Scene.Node)
	if err != nil {
		return err
	}
	cols, ok := obj["cols"].([]any)
	if obj["type"] != "table" || !ok || len(cols) == 0 {
		return fmt.Errorf("library.v12_allocation_columns: %s", key)
	}
	target, ok := cols[0].(map[string]any)
	if !ok {
		return fmt.Errorf("library.v12_allocation_column: %s", key)
	}
	if intakeRepairNumber(target, "w") != before && intakeRepairNumber(target, "w") != after {
		return fmt.Errorf("library.v12_allocation_geometry: %s", key)
	}
	target["w"] = after
	if key == "backlog/wsjf-scoring" {
		if len(cols) != 8 {
			return fmt.Errorf("library.v12_allocation_columns: %s", key)
		}
		last, ok := cols[7].(map[string]any)
		if !ok || last["k"] != "p" || (intakeRepairNumber(last, "w") != 60 && intakeRepairNumber(last, "w") != 72) {
			return fmt.Errorf("library.v12_allocation_geometry: %s", key)
		}
		last["w"] = 72.
	}
	if key == "roadmap-strategy/cascade-split" {
		rows, ok := obj["rows"].([]any)
		if !ok || len(rows) != 6 {
			return fmt.Errorf("library.v12_allocation_rows: %s", key)
		}
		for _, raw := range rows {
			row, ok := raw.(map[string]any)
			if !ok || (intakeRepairNumber(row, "h") != 63 && intakeRepairNumber(row, "h") != 61) {
				return fmt.Errorf("library.v12_allocation_row_height: %s", key)
			}
			row["h"] = 61.
		}
	}
	scene := *node.Scene
	scene.Node, err = json.Marshal(obj)
	if err != nil {
		return err
	}
	marker := "wmds.v12.source-allocation.table-columns"
	if !intakeRepairHasResolution(scene.Resolutions, marker) {
		scene.Resolutions = append(append([]string(nil), scene.Resolutions...), marker)
	}
	node.Scene = &scene
	if key == "roadmap-strategy/cascade-split" {
		slide.Frame.SourceLines = 2
	}
	slide.Nodes = append([]Node(nil), slide.Nodes...)
	slide.Nodes[ordinal-1] = node
	return nil
}
