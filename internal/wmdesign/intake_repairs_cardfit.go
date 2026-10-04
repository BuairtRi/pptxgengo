package wmdesign

import (
	"encoding/json"
	"fmt"
)

// Source-pinned card allocations retain copy and typography. Each named layout
// spends unused outer padding on readable content rather than weakening its
// measured bounds. This helper runs inside the incoming-only atomic wrapper.
func applyIncomingCardFitRepairs(key string, slide *SlideSpec) error {
	type policy struct {
		nodes, cards    int
		kind            string
		w, h, pad, next float64
	}
	var q policy
	switch key {
	case "cards/narrative-2x3-rows":
		q = policy{11, 6, "card", 198, 144, 18, 12}
	case "cards/narrative-2x3-slim":
		q = policy{10, 6, "card", 198, 144, 18, 12}
	case "key-message/icons-tint":
		q = policy{7, 4, "card", 270, 72, 0, 9}
	case "key-message/photo-icons":
		q = policy{7, 4, "card", 270, 60, 0, 9}
	case "key-message/stat", "key-message/stat-nav":
		q = policy{5, 1, "card", 54, 54, 9, 6}
	case "key-message/stat-left":
		q = policy{7, 1, "card", 54, 54, 9, 6}
	case "lifecycle/columns-outcomes":
		q = policy{18, 3, "card", 270, 72, 12, 9}
	case "phase/define", "phase/build":
		q = policy{9, 1, "card", 522, 54, 12, 9}
	case "road/cards":
		q = policy{2, 1, "cardrow", 198, 90, 12, 9}
	case "roadmap/grid-columns":
		q = policy{26, 12, "card", 162, 72, 9, 6}
	case "sequence/evidence-to-model":
		q = policy{19, 3, "card", 198, 72, 9, 6}
	case "sequence/seven-r":
		q = policy{29, 7, "card", 0, 90, 9, 6}
	default:
		return nil
	}
	if slide == nil || len(slide.Nodes) != q.nodes {
		return fmt.Errorf("intake.cardfit_source_topology: %s", key)
	}
	count := 0
	seen := map[string]bool{}
	for i := range slide.Nodes {
		n := &slide.Nodes[i]
		if n.Scene == nil {
			return fmt.Errorf("intake.cardfit_non_scene_topology: %s", key)
		}
		obj, err := libraryObject(n.Scene.Node)
		if err != nil {
			return err
		}
		if obj["type"] != q.kind {
			continue
		}
		w, h := intakeRepairNumber(obj, "w"), intakeRepairNumber(obj, "h")
		if key == "key-message/icons-tint" && h == q.h {
			// Two 252pt cards and an18pt gutter fit the558pt tall zone with
			// an18pt inset at both sides. Source270pt columns ended at921.
			x, y := intakeRepairNumber(obj, "x"), intakeRepairNumber(obj, "y")
			if (w != 270 && w != 252) || (x != 363 && x != 651 && x != 633) || (y != 108 && y != 216) {
				return fmt.Errorf("intake.cardfit_unexpected_icon_grid: %s/%s", key, n.ID)
			}
			if x == 651 {
				obj["x"] = 633
			}
			obj["w"] = 252
			w = q.w // The reviewed final grid replaces the source width.
		}
		// The evidence-model's central synthesis card and phase-detail's
		// activity cards are separate, deliberately taller allocations.
		if key == "sequence/evidence-to-model" && w == 180 && h == 252 || (key == "phase/define" || key == "phase/build") && w == 162 && h == 162 {
			continue
		}
		x, y := intakeRepairNumber(obj, "x"), intakeRepairNumber(obj, "y")
		if key == "key-message/icons-tint" && x == 633 {
			x = 651
		}
		if !intakeCardFitSourcePosition(key, x, y) {
			return fmt.Errorf("intake.cardfit_unexpected_position: %s/%s", key, n.ID)
		}
		if key == "sequence/seven-r" {
			wantWidth := 102.0
			if x == 777 {
				wantWidth = 126
			}
			if w != wantWidth {
				return fmt.Errorf("intake.cardfit_unexpected_seven_r_width: %s", n.ID)
			}
		}
		position := fmt.Sprintf("%g/%g", x, y)
		if seen[position] {
			return fmt.Errorf("intake.cardfit_duplicate_position: %s/%s", key, n.ID)
		}
		seen[position] = true
		if h != q.h || q.w != 0 && w != q.w || q.w == 0 && w != 102 && w != 126 {
			return fmt.Errorf("intake.cardfit_unexpected_allocation: %s/%s", key, n.ID)
		}
		padObject := obj
		if q.kind == "cardrow" {
			var ok bool
			padObject, ok = obj["card"].(map[string]any)
			if !ok {
				return fmt.Errorf("intake.cardfit_cardrow_style: %s", key)
			}
			items, ok := obj["items"].([]any)
			if !ok || len(items) != 4 {
				return fmt.Errorf("intake.cardfit_cardrow_cardinality: %s", key)
			}
		}
		pad := intakeRepairNumber(padObject, "pad")
		if pad != q.pad && pad != q.next {
			return fmt.Errorf("intake.cardfit_unexpected_padding: %s/%s", key, n.ID)
		}
		padObject["pad"] = q.next
		n.Scene.Node, err = json.Marshal(obj)
		if err != nil {
			return err
		}
		resolution := fmt.Sprintf("wmds.v4.cardfit.%s.padding-%g", key, q.next)
		if !intakeRepairHasResolution(n.Scene.Resolutions, resolution) {
			n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
		}
		if key == "key-message/icons-tint" {
			resolution = "wmds.v4.cardfit.key-message/icons-tint.two-columns-252-gap-18"
			if !intakeRepairHasResolution(n.Scene.Resolutions, resolution) {
				n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
			}
		}
		count++
	}
	if count != q.cards {
		return fmt.Errorf("intake.cardfit_card_count: %s got%d want%d", key, count, q.cards)
	}
	return nil
}

func intakeCardFitSourcePosition(key string, x, y float64) bool {
	var xs, ys []float64
	switch key {
	case "cards/narrative-2x3-rows":
		xs, ys = []float64{273, 489, 705}, []float64{144, 324}
	case "cards/narrative-2x3-slim":
		xs, ys = []float64{273, 489, 705}, []float64{126, 288}
	case "key-message/icons-tint":
		xs, ys = []float64{363, 651}, []float64{108, 216}
	case "key-message/photo-icons":
		xs, ys = []float64{345, 633}, []float64{270, 342}
	case "key-message/stat", "key-message/stat-nav":
		xs, ys = []float64{831}, []float64{180}
	case "key-message/stat-left":
		xs, ys = []float64{471}, []float64{288}
	case "lifecycle/columns-outcomes":
		xs, ys = []float64{57, 345, 633}, []float64{396}
	case "phase/define", "phase/build":
		xs, ys = []float64{381}, []float64{414}
	case "road/cards":
		xs, ys = []float64{57}, []float64{360}
	case "roadmap/grid-columns":
		xs, ys = []float64{201, 381, 561, 741}, []float64{180, 270, 360}
	case "sequence/evidence-to-model":
		xs, ys = []float64{57}, []float64{180, 270, 360}
	case "sequence/seven-r":
		xs, ys = []float64{57, 165, 309, 417, 525, 669, 777}, []float64{342}
	}
	for _, xx := range xs {
		if xx == x {
			for _, yy := range ys {
				if yy == y {
					return true
				}
			}
		}
	}
	return false
}
