package wmdesign

import (
	"encoding/json"
	"fmt"
)

// applyProofEvidenceLibraryRefinements preserves pinned source and caller copy.
// Amendments allocate explicit grid space; they never shrink text to fit.
func applyProofEvidenceLibraryRefinements(key, revision string, doc *SlideSpec) error {
	if revision != LibraryRevisionV2 {
		return nil
	}
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
				if _, ok := raw[field]; !ok && !(kind == "textblock" && field == "h") {
					return fmt.Errorf("library.refinement_field_missing: %s/%s", id, field)
				}
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
	case "case-study/what-we-did", "case-study/what-we-did-split":
		// Fixed row allocations reserve9pt before separators and the Results
		// label. The renderer rejects caller copy exceeding these capacities.
		for _, row := range []struct {
			id string
			h  float64
		}{{"node02", 57}, {"node03", 57}, {"node05", 57}, {"node06", 57}, {"node08", 75}, {"node09", 75}} {
			if err := change(row.id, "textblock", "wmds.what-we-did-row-clearance.v2", map[string]any{"h": row.h}); err != nil {
				return err
			}
		}
	case "case-study/quote-outcome":
		// The original 126pt card cannot hold the authored quote mark, copy and
		// attribution. Its 162pt replacement ends at the same450pt body boundary.
		return change("node06", "card", "wmds.quote-outcome-quote-capacity.v2", map[string]any{"y": 288, "h": 162, "pad": 9})
	case "chart/column-full-split":
		// Preserve the source's native chart and source-line reservation.
		return change("node01", "chart", "wmds.column-full-split-source-clearance.v2", map[string]any{"h": 414})
	}
	return nil
}
