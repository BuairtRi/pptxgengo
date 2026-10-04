package wmdesign

import (
	"encoding/json"
	"fmt"
)

// These v3 amendments preserve the frozen maps, geographic projections, names,
// and label anchors. They reserve the native label boxes and SVG border bleed
// inside the tall-right zone, and leave the US legend below the actual map.
func applyGeographyLibraryRefinements(key, revision string, doc *SlideSpec) error {
	if revision != LibraryRevisionV3 {
		return nil
	}
	var target string
	var x, w float64
	switch key {
	case "about/locations":
		target, x, w = "node05", 399, 414
	case "about/locations-international":
		target, x, w = "node12", 363, 504
	default:
		return nil
	}
	for i := range doc.Nodes {
		node := &doc.Nodes[i]
		if node.ID != target {
			continue
		}
		if node.Scene == nil {
			return fmt.Errorf("library.refinement_target_not_scene: %s", target)
		}
		raw, err := libraryObject(node.Scene.Node)
		if err != nil {
			return err
		}
		if raw["type"] != "dotmap" {
			return fmt.Errorf("library.refinement_type_mismatch: %s", target)
		}
		raw["x"], raw["w"] = x, w
		node.Scene.Node, err = json.Marshal(raw)
		if err != nil {
			return err
		}
		node.Scene.Resolutions = append(node.Scene.Resolutions, "wmds.locations-map-label-and-legend-capacity.v3")
		return nil
	}
	return fmt.Errorf("library.refinement_target_missing: %s", target)
}
