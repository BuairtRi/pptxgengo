package wmdesign

import (
	"encoding/json"
	"fmt"
)

// Future-revision allocations for the frozen October 3 compositions. The caller
// applies these to an atomic clone after binding, leaving visible copy intact.
func applyIncomingTextFitRepairs(key string, slide *SlideSpec) error {
	nodes, targets := 0, 0
	switch key {
	case "sequence/evidence-to-model":
		nodes, targets = 19, 1
	case "adoption/dashboard-split":
		nodes, targets = 5, 2
	case "funnel/compare":
		nodes, targets = 6, 2
	case "timeline/swimlanes":
		nodes, targets = 21, 1
	case "timeline/vertical":
		nodes, targets = 30, 25
	case "activities/subtrack-timeline":
		nodes, targets = 44, 1
	default:
		return nil
	}
	if len(slide.Nodes) != nodes {
		return fmt.Errorf("intake.repair_textfit_source_topology: %s", key)
	}
	changed := 0
	seenTargets := map[string]bool{}
	if key == "timeline/vertical" {
		if slide.Frame.TitleLines != 1 && slide.Frame.TitleLines != 2 {
			return fmt.Errorf("intake.repair_timeline_title_lines")
		}
		slide.Frame.TitleLines = 2
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
		kind, _ := obj["type"].(string)
		x, y := intakeRepairNumber(obj, "x"), intakeRepairNumber(obj, "y")
		resolution := ""
		switch key {
		case "sequence/evidence-to-model":
			if kind == "text" && x == 507 && obj["ink"] == "display" {
				if intakeRepairNumber(obj, "w") != 180 || obj["style"] != "small" || (y != 324 && y != 342) {
					return fmt.Errorf("intake.repair_evidence_review_allocation")
				}
				obj["y"] = 342
				resolution = "wmds.v4.evidence-review-followup-start-342"
			}
		case "activities/subtrack-timeline":
			if kind == "block" && x == 831 && y == 432 {
				w := intakeRepairNumber(obj, "w")
				if (w != 66 && w != 72) || intakeRepairNumber(obj, "h") != 30 {
					return fmt.Errorf("intake.repair_hypercare_allocation")
				}
				obj["w"] = 72
				resolution = "wmds.v4.hypercare-full-grid-width-72"
			}
		case "adoption/dashboard-split":
			if kind == "chart" {
				if x != 345 || y != 36 || intakeRepairNumber(obj, "w") != 558 {
					return fmt.Errorf("intake.repair_adoption_chart_topology")
				}
				h := intakeRepairNumber(obj, "h")
				if h != 198 && h != 180 {
					return fmt.Errorf("intake.repair_adoption_chart_height")
				}
				obj["h"] = 180
				resolution = "wmds.v4.adoption-chart-height-180"
			}
			if kind == "table" {
				if x != 345 || intakeRepairNumber(obj, "w") != 558 || (y != 252 && y != 234) {
					return fmt.Errorf("intake.repair_adoption_table_topology")
				}
				obj["y"] = 234
				resolution = "wmds.v4.adoption-table-start-234"
			}
		case "funnel/compare":
			if kind == "text" && (obj["style"] == "small" || obj["style"] == "body") {
				if (x != 57 && x != 489) || intakeRepairNumber(obj, "w") != 414 || (y != 432 && y != 426) {
					return fmt.Errorf("intake.repair_funnel_caption_topology")
				}
				obj["y"] = 426
				resolution = "wmds.v4.funnel-caption-start-426"
			}
		case "timeline/swimlanes":
			if kind == "text" && x == 87 {
				if intakeRepairNumber(obj, "w") != 400 || (y != 433 && y != 432) {
					return fmt.Errorf("intake.repair_timeline_legend_topology")
				}
				obj["y"] = 432
				resolution = "wmds.v4.timeline-legend-align-432"
			}
		case "timeline/vertical":
			// Keep six authored milestones under the two-line title. A 48pt row
			// contains the 27pt subhead and 18pt detail plus3pt inter-row space.
			const marker = "wmds.v4.timeline-six-rows-48"
			already := intakeRepairHasResolution(n.Scene.Resolutions, marker)
			if kind == "connector" {
				if seenTargets["axis"] {
					return fmt.Errorf("intake.repair_timeline_duplicate_axis")
				}
				seenTargets["axis"] = true
				pts, ok := obj["points"].([]any)
				if !ok || len(pts) != 2 {
					return fmt.Errorf("intake.repair_timeline_connector_topology")
				}
				a, okA := pts[0].([]any)
				b, okB := pts[1].([]any)
				if !okA || !okB || len(a) != 2 || len(b) != 2 {
					return fmt.Errorf("intake.repair_timeline_connector_points")
				}
				av, okA := discoveryNumber(a[1])
				bv, okB := discoveryNumber(b[1])
				ax, okX := discoveryNumber(a[0])
				bx, okY := discoveryNumber(b[0])
				firstY, lastY := 144., 414.
				if already {
					firstY, lastY = 180, 420
				}
				if !okA || !okB || !okX || !okY || ax != 183 || bx != 183 || av != firstY || bv != lastY {
					return fmt.Errorf("intake.repair_timeline_connector_geometry")
				}
				obj["points"] = [][]float64{{183, 180}, {183, 420}}
				resolution = marker
			} else if x == 57 || x == 165 || x == 219 {
				origin, offset := 126., 0.
				height := 0.
				width := 396.
				switch {
				case x == 57 && kind == "text":
					origin, offset = 135, 9
					height = 18
					width = 96
				case x == 165 && kind == "block":
					origin = 126
					width = 36
					if intakeRepairNumber(obj, "h") != 36 {
						return fmt.Errorf("intake.repair_timeline_marker_height")
					}
				case x == 219 && kind == "text" && obj["style"] == "subhead":
					origin = 123
					height = 27
				case x == 219 && kind == "text" && obj["style"] == "small":
					origin, offset = 147, 27
					height = 18
				default:
					return fmt.Errorf("intake.repair_timeline_row_topology")
				}
				if intakeRepairNumber(obj, "w") != width {
					return fmt.Errorf("intake.repair_timeline_row_width")
				}
				step := 54.
				if already {
					origin, step = 162+offset, 48
				}
				row := (y - origin) / step
				if row < 0 || row > 5 || row != float64(int(row)) {
					return fmt.Errorf("intake.repair_timeline_row_position")
				}
				target := fmt.Sprintf("%g:%v:%g", x, obj["style"], row)
				if seenTargets[target] {
					return fmt.Errorf("intake.repair_timeline_duplicate_row")
				}
				seenTargets[target] = true
				obj["y"] = 162 + row*48 + offset
				if height > 0 {
					if old, exists := obj["h"]; !already && exists && old != nil || already && intakeRepairNumber(obj, "h") != height {
						return fmt.Errorf("intake.repair_timeline_row_height")
					}
					obj["h"] = height
				}
				resolution = marker
			}
		}
		if resolution != "" {
			changed++
			raw, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			n.Scene.Node = raw
			if !intakeRepairHasResolution(n.Scene.Resolutions, resolution) {
				n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
			}
		}
	}
	if changed != targets {
		return fmt.Errorf("intake.repair_textfit_target_count: %s got%d want%d", key, changed, targets)
	}
	return nil
}
