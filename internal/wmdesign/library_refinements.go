package wmdesign

import (
	"encoding/json"
	"fmt"
)

// These named composition amendments preserve the pinned source and caller copy.
// They are not content-driven shrinking or geometry supplied by a binding.
func applyLibraryRefinements(key, revision string, doc *SlideSpec) error {
	if revision == LibraryRevisionV6 {
		if isV6HeatmapDelta(key) {
			return applyV6HeatmapRefinements(key, doc)
		}
		return applyLibraryRefinements(key, LibraryRevisionV5, doc)
	}
	if revision == LibraryRevisionV5 {
		if v5RetainedV4Compositions[key] {
			if err := applyLibraryRefinements(key, LibraryRevisionV4, doc); err != nil {
				return err
			}
		}
		if err := ApplyIncomingV5Repairs(key, revision, doc); err != nil {
			return err
		}
		return applyV5NativeChartRefinement(key, doc)
	}
	if revision == LibraryRevisionV4 {
		if v4RetainedV3Compositions[key] {
			if err := applyLibraryRefinements(key, LibraryRevisionV3, doc); err != nil {
				return err
			}
		}
		return ApplyIncomingIntakeRepairs(key, revision, doc)
	}
	for _, refine := range []func(string, string, *SlideSpec) error{
		ApplyIncomingIntakeRepairs,
		applyIntakeLibraryRefinements,
		applyApproachCommercialLibraryRefinements,
		applyProofEvidenceLibraryRefinements,
		applyTeamSolutionLibraryRefinements,
		applyGeographyLibraryRefinements,
	} {
		if err := refine(key, revision, doc); err != nil {
			return err
		}
	}
	// The refreshed designs author these elements directly; v1 amendments must
	// not append a duplicate photo or overwrite the new hand-drawn arrows.
	if isModernLibrary(revision) && (key == "agenda/schedule" || key == "pillars/four-why-matters") {
		return nil
	}
	if err := applyPrimitiveLibraryRefinement(key, doc); err != nil {
		return err
	}
	change := func(ordinal int, kind, resolution string, values map[string]any) error {
		id := fmt.Sprintf("node%02d", ordinal)
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
			for k, v := range values {
				if _, ok := raw[k]; !ok {
					return fmt.Errorf("library.refinement_field_missing: %s/%s", id, k)
				}
				raw[k] = v
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
	case "divider/inverse":
		if isModernLibrary(revision) {
			const resolution = "wmds.inverse-divider-separated-columns.v2"
			if err := change(1, "square", resolution, map[string]any{"x": 507, "y": 126, "size": 270}); err != nil {
				return err
			}
			// Retain the photo-over-whiteboard layering with three 18pt grid
			// steps. The left column reserves an 18pt gap before the photo.
			doc.LibraryChrome.Whiteboard = []LibraryWhiteboard{{X: 561, Y: 54, W: 342, H: 396, Fade: "none"}}
			for _, item := range []struct {
				ordinal int
				y       float64
			}{{2, 216}, {3, 234}} {
				if err := change(item.ordinal, "text", resolution, map[string]any{"w": 432, "y": item.y}); err != nil {
					return err
				}
			}
		}
	case "stats/circled-headline":
		if isModernLibrary(revision) {
			return change(4, "card", "wmds.circled-headline-explanation-capacity.v2", map[string]any{"h": 108})
		}
	case "architecture/layer-map":
		if err := change(16, "container", "wmds.architecture-layer-padding.v1", map[string]any{"h": 294}); err != nil {
			return err
		}
		return change(17, "matrix", "wmds.architecture-layer-padding.v1", map[string]any{"y": 186})
	case "context/three-zones":
		return change(30, "connector", "wmds.three-zones-arrow-clearance.v1", map[string]any{"points": [][]float64{{603, 279}, {621, 279}}})
	case "phase-detail/rail-cohort":
		if err := change(8, "text", "wmds.rail-cohort-scorecard-spacing.v1", map[string]any{"y": 210, "w": 194}); err != nil {
			return err
		}
		for i := 0; i < 5; i++ {
			for j, item := range []struct {
				kind string
				y    float64
			}{{"text", 264}, {"text", 277}, {"rule", 299}} {
				if err := change(9+3*i+j, item.kind, "wmds.rail-cohort-scorecard-spacing.v1", map[string]any{"y": item.y + 38*float64(i)}); err != nil {
					return err
				}
			}
		}
	case "pillars/four-why-matters":
		const resolution = "wmds.pillar-handdrawn-arrow-clearance.v1"
		for i := 0; i < 4; i++ {
			x := 57 + 216*float64(i)
			if err := change(2+4*i, "block", resolution, map[string]any{"x": x + 30, "w": 168}); err != nil {
				return err
			}
			for _, ordinal := range []int{3 + 4*i, 4 + 4*i} {
				if err := change(ordinal, "text", resolution, map[string]any{"x": x + 42, "w": 144}); err != nil {
					return err
				}
			}
			node := &doc.Nodes[16+i]
			obj, err := libraryObject(node.Scene.Node)
			if err != nil {
				return err
			}
			if node.ID != fmt.Sprintf("node%02d", 17+i) || obj["type"] != "connector" {
				return fmt.Errorf("library.refinement_type_mismatch: pillar arrow")
			}
			node.Scene.Node, err = json.Marshal(map[string]any{"type": "mark", "mark": "arrow-right-angle", "x": x + 3, "y": 366, "w": 24, "h": 36, "ink": "strong"})
			if err != nil {
				return err
			}
			node.Scene.Resolutions = append(node.Scene.Resolutions, resolution)
		}
	}
	return nil
}
