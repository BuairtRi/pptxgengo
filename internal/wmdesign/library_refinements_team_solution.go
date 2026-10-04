package wmdesign

import (
	"encoding/json"
	"fmt"
)

// applyTeamSolutionLibraryRefinements applies recorded v2 family amendments.
// These preserve the pinned source and caller copy; they are never font shrink.
func applyTeamSolutionLibraryRefinements(key, revision string, doc *SlideSpec) error {
	if !isModernLibrary(revision) || key != "team/org-roles-nav" {
		return nil
	}
	for i := range doc.Nodes {
		n := &doc.Nodes[i]
		if n.ID != "node25" {
			continue
		}
		if n.Scene == nil {
			return fmt.Errorf("library.refinement_target_not_scene: node25")
		}
		raw, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		if raw["type"] != "table" {
			return fmt.Errorf("library.refinement_type_mismatch: node25")
		}
		if _, ok := raw["rowH"]; !ok {
			return fmt.Errorf("library.refinement_field_missing: node25/rowH")
		}
		// The source dense heading is 28pt: y144+28+6*48=460pt.
		// 46pt rows retain measured two-bullet allocations and end at448pt,
		// inside the 450pt nav-frame footer boundary.
		raw["rowH"] = 46
		n.Scene.Node, err = json.Marshal(raw)
		if err != nil {
			return err
		}
		n.Scene.Resolutions = append(n.Scene.Resolutions, "wmds.org-roles-nav-footer-capacity.v2")
		return nil
	}
	return fmt.Errorf("library.refinement_target_missing: node25")
}
