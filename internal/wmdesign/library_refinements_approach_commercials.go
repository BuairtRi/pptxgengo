package wmdesign

import (
	"encoding/json"
	"fmt"
)

// applyApproachCommercialLibraryRefinements preserves the pinned source and
// caller copy while documenting fixed, source-specific geometry amendments.
func applyApproachCommercialLibraryRefinements(key, revision string, doc *SlideSpec) error {
	if !isModernLibrary(revision) {
		return nil
	}
	// Review amendments are applied after closed content binding, so every
	// source and caller specimen uses the same geometry without changing copy.
	change := func(id, kind, resolution string, values map[string]any) error {
		for i := range doc.Nodes {
			n := &doc.Nodes[i]
			if n.ID != id {
				continue
			}
			if n.Scene == nil {
				return fmt.Errorf("library.refinement_target_not_scene: %s", id)
			}
			raw, err := libraryObject(n.Scene.Node)
			if err != nil {
				return err
			}
			if raw["type"] != kind {
				return fmt.Errorf("library.refinement_type_mismatch: %s", id)
			}
			for field, value := range values {
				raw[field] = value
			}
			n.Scene.Node, err = json.Marshal(raw)
			if err != nil {
				return err
			}
			n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
			return nil
		}
		return fmt.Errorf("library.refinement_target_missing: %s", id)
	}
	switch key {
	case "phases/summary-bands":
		for _, id := range []string{"node01", "node05"} {
			if err := change(id, "chevron", "wmds.phase-summary-equal-width.v2", map[string]any{"w": 270}); err != nil {
				return err
			}
		}
		return nil
	case "plan/gantt":
		// 24pt tracks still contain 22pt bars. Retain 42pt single-track lanes
		// for heading padding; 60pt double-track lanes free a lower legend row.
		return change("node01", "gantt", "wmds.work-plan-bottom-legend.v2", map[string]any{
			"trackPitch": 24, "legendSize": 8, "legendFullWidth": true,
		})
	case "roadmap/staggered-phases", "roadmap/staggered-phases-nav":
		return change("node01", "gantt", "wmds.roadmap-compact-legend.v2", map[string]any{
			"legendSize": 8, "legendFullWidth": true,
		})
	case "runbook/escalation-split":
		for i, id := range []string{"node07", "node08", "node09"} {
			if err := change(id, "mark", "wmds.escalation-straight-arrow.v2", map[string]any{
				"mark": "arrow-straight", "rotate": -90, "x": 375, "y": 123.75 + 108*float64(i), "w": 24, "h": 4.5,
			}); err != nil {
				return err
			}
		}
		return nil
	}
	if key == "runbook/step-list-split" {
		const resolution = "wmds.runbook-step-list-two-line-title-reservation.v2"
		for row := 0; row < 4; row++ {
			for part := 0; part < 3; part++ {
				id := fmt.Sprintf("node%02d", 3+5*row+part)
				found := false
				for i := range doc.Nodes {
					n := &doc.Nodes[i]
					if n.ID != id {
						continue
					}
					if n.Scene == nil {
						return fmt.Errorf("library.refinement_target_not_scene: %s", id)
					}
					raw, err := libraryObject(n.Scene.Node)
					if err != nil {
						return err
					}
					kind := "text"
					y := float64(108 + 108*row + 18*(part-1))
					if part == 0 {
						kind, y = "numhead", float64(42+108*row)
					}
					if raw["type"] != kind {
						return fmt.Errorf("library.refinement_type_mismatch: %s", id)
					}
					raw["y"] = y
					n.Scene.Node, err = json.Marshal(raw)
					if err != nil {
						return err
					}
					n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
					found = true
				}
				if !found {
					return fmt.Errorf("library.refinement_target_missing: %s", id)
				}
			}
		}
		return nil
	}
	if key != "runbook/overview" {
		return nil
	}
	// The final escalation role wraps to three body lines. Use the remaining
	// grid row before the compact footer; keep type size and all source copy.
	for ordinal := 9; ordinal <= 12; ordinal++ {
		id := fmt.Sprintf("node%02d", ordinal)
		found := false
		for i := range doc.Nodes {
			n := &doc.Nodes[i]
			if n.ID != id {
				continue
			}
			if n.Scene == nil {
				return fmt.Errorf("library.refinement_target_not_scene: %s", id)
			}
			raw, err := libraryObject(n.Scene.Node)
			if err != nil {
				return err
			}
			if raw["type"] != "chevron" {
				return fmt.Errorf("library.refinement_type_mismatch: overview %s", id)
			}
			// Stack the existing time and role rather than consuming most
			// of the narrow chevron with a horizontal time label.
			time, role := raw["number"], raw["text"]
			raw["h"], raw["number"], raw["text"] = 72, "", ""
			x := 57.0 + 216*float64(ordinal-9)
			left, width := 12.0, 174.0
			if ordinal > 9 {
				left, width = 30, 156
			}
			if ordinal == 12 {
				width = 138
			}
			for _, item := range []struct {
				suffix, style string
				y             float64
				copy          any
			}{
				{"time", "number", 399, time},
			} {
				text, e := json.Marshal(map[string]any{"type": "text", "x": x + left, "y": item.y,
					"w": width, "style": item.style, "ink": "emphasis", "on": raw["surface"], "text": item.copy})
				if e != nil {
					return e
				}
				doc.Nodes = append(doc.Nodes, Node{ID: id + "-" + item.suffix, Kind: "scene", Scene: &SceneSpec{
					Node: text, Path: fmt.Sprintf("/body/%d/%s", ordinal-1, "number"),
					Resolutions: []string{"wmds.runbook-overview-escalation-capacity.v2"}}})
			}
			text, err := json.Marshal(map[string]any{"type": "text", "x": x + left, "y": 426,
				"w": width, "h": 42, "style": "body", "weight": 600,
				"ink": "display", "on": raw["surface"], "text": role})
			if err != nil {
				return err
			}
			doc.Nodes = append(doc.Nodes, Node{ID: id + "-role", Kind: "scene", Scene: &SceneSpec{
				Node: text, Path: fmt.Sprintf("/body/%d/text", ordinal-1),
				Resolutions: []string{"wmds.runbook-overview-escalation-capacity.v2"}}})
			doc.Nodes[i].Scene.Node, err = json.Marshal(raw)
			if err != nil {
				return err
			}
			doc.Nodes[i].Scene.Resolutions = append(doc.Nodes[i].Scene.Resolutions, "wmds.runbook-overview-escalation-capacity.v2")
			found = true
		}
		if !found {
			return fmt.Errorf("library.refinement_target_missing: %s", id)
		}
	}
	return nil
}
