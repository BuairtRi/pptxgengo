package wmdesign

import (
	"encoding/json"
	"fmt"
)

// Explicit v3 composition amendments keep authored copy and the frozen source
// intact. They are fixed design choices, never an automatic text-shrink policy.
func applyIntakeLibraryRefinements(key, revision string, slide *SlideSpec) error {
	if revision != LibraryRevisionV3 {
		return nil
	}
	for i := range slide.Nodes {
		n := &slide.Nodes[i]
		if n.Scene == nil {
			continue
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		resolution := ""
		switch key {
		case "about/glance", "case-study/quote-outcome-right":
			if quote, ok := obj["quote"].(map[string]any); ok {
				quote["layout"] = "side"
				quote["textStyle"] = "body"
				resolution = "wmds.v3.compact-quote-side-mark"
			}
		case "about/industries-left":
			if obj["type"] == "card" {
				obj["pad"] = 12
				resolution = "wmds.v3.industry-card-padding-12"
			}
		case "architecture/layers-icons":
			if n.ID == "node07" {
				obj["pad"] = 9
				resolution = "wmds.v3.layer-controls-padding-9"
			}
		case "architecture/app-ecosystem":
			if n.ID == "node06" {
				obj["style"] = "small"
				resolution = "wmds.v3.compact-api-gateway-label"
			}
		case "cards/narrative-2x3-badge":
			if obj["type"] == "card" {
				obj["gap"] = 3
				resolution = "wmds.v3.narrative-card-gap-3"
			}
		case "process/current-future-right":
			if obj["type"] == "chevron" {
				obj["style"] = "small"
				resolution = "wmds.v3.compact-process-chevron-labels"
			}
		case "status/four-panel-right", "status/four-panel-tall":
			if obj["type"] == "card" && obj["metric"] != nil {
				obj["pad"] = 12
				resolution = "wmds.v3.status-metric-padding-12"
			}
			if key == "status/four-panel-right" && obj["type"] == "table" {
				cols, ok := obj["cols"].([]any)
				if ok && len(cols) == 3 {
					// Give the count header room without narrowing the status pill.
					widths := []float64{66, 72, 132}
					for j, raw := range cols {
						if col, ok := raw.(map[string]any); ok {
							col["w"] = widths[j]
						}
					}
					resolution = "wmds.v3.status-table-count-column-72"
				}
			}
		}
		if resolution != "" {
			n.Scene.Node, err = json.Marshal(obj)
			if err != nil {
				return fmt.Errorf("library.intake_refinement: %w", err)
			}
			n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
		}
	}
	return nil
}
