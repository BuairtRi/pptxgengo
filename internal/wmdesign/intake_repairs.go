package wmdesign

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// IntakeRepairRevision is an incoming, unreleased composition contract. These
// fixed amendments never affect the installed v3 library or shrink user copy.
const IntakeRepairRevision = "wmds-library.v4"

// ApplyIncomingIntakeRepairs lowers reviewed incoming compositions to native
// allocations. Call after binding/compilation and before planning. Each change
// is explicit in the scene's resolution receipt. The original source stays
// immutable; unexpected source topology is rejected rather than guessed.
func ApplyIncomingIntakeRepairs(key, revision string, slide *SlideSpec) error {
	if revision != IntakeRepairRevision {
		return nil
	}
	if slide == nil {
		return fmt.Errorf("intake.repair_nil_slide")
	}
	// Commit only after the whole amendment has validated, so a stale incoming
	// definition cannot leave half of a slide changed.
	copy := *slide
	copy.Nodes = append([]Node(nil), slide.Nodes...)
	for i, original := range slide.Nodes {
		if original.Scene == nil {
			continue
		}
		scene := *original.Scene
		scene.Resolutions = append([]string(nil), scene.Resolutions...)
		scene.Keys = intakeRepairKeys(scene.Keys)
		copy.Nodes[i].Scene = &scene
	}
	if err := applyIncomingIntakeRepair(key, &copy); err != nil {
		return err
	}
	if err := applyIncomingTextFitRepairs(key, &copy); err != nil {
		return err
	}
	if err := applyIncomingCardFitRepairs(key, &copy); err != nil {
		return err
	}
	if err := applyIncomingTableFitRepairs(key, &copy); err != nil {
		return err
	}
	*slide = copy
	return nil
}

func intakeRepairKeys(keys map[string][]string) map[string][]string {
	if keys == nil {
		return nil
	}
	out := make(map[string][]string, len(keys))
	for key, values := range keys {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func applyIncomingIntakeRepair(key string, slide *SlideSpec) error {
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
		resolution := ""
		if strings.HasPrefix(key, "venn/") && slide.Frame.Split != "" && kind == "venn" {
			sets, ok := obj["sets"].([]any)
			if !ok || len(sets) < 2 || len(sets) > 4 {
				return fmt.Errorf("intake.repair_venn_sets: %s", key)
			}
			// Retain the circle center and inset half a grid step from each
			// horizontal frame boundary. Smaller height leaves actual outer
			// circles and strokes inside the source/footer reservation.
			if intakeRepairNumber(obj, "w") == 558 && (intakeRepairNumber(obj, "x") == 57 || intakeRepairNumber(obj, "x") == 345) && intakeRepairNumber(obj, "y") == 36 {
				obj["x"] = intakeRepairNumber(obj, "x") + 9
				obj["w"] = 540
				if len(sets) == 3 {
					obj["h"] = 396
				} else {
					obj["h"] = 414
				}
				resolution = "wmds.v4.venn-split-circle-inset-9"
			}
		}
		switch key {
		case "stakeholders/quadrant":
			if kind == "chart" && obj["kind"] == "quadrant" && obj["key"] == "true" {
				// This single source specimen uses a string for a truthy legend
				// flag. Normalize its exact intended value without widening the
				// chart planner's boolean-only contract.
				obj["key"] = true
				resolution = "wmds.v4.stakeholder-quadrant-boolean-key"
			}
		case "pdlc/cycle-overview", "pdlc/cycle-overview-split", "sdlc/cycle", "sdlc/cycle-split":
			if kind == "cycle" {
				// These four incoming diagrams need an additional grid step for
				// their two-line node descriptions. The cycle planner derives
				// its ring/edge geometry from this explicit node allocation.
				if obj["nodeH"] == nil || intakeRepairNumber(obj, "nodeH") == 72 {
					obj["nodeH"] = 90
					resolution = "wmds.v4.lifecycle-cycle-node-allocation-90"
				} else if intakeRepairNumber(obj, "nodeH") != 90 {
					return fmt.Errorf("intake.repair_unexpected_cycle_node_height: %s", key)
				}
			}
		case "maturity/split", "maturity/table-tall", "maturity/insights":
			if kind == "maturity" && ((intakeRepairNumber(obj, "x") == 345 && intakeRepairNumber(obj, "w") == 558) || (intakeRepairNumber(obj, "x") == 57 && (intakeRepairNumber(obj, "w") == 558 || intakeRepairNumber(obj, "w") == 270))) {
				obj["x"], obj["w"] = intakeRepairNumber(obj, "x")+9, intakeRepairNumber(obj, "w")-18
				resolution = "wmds.v4.maturity-split-stroke-inset-9"
			}
		case "capability-heat/callouts":
			if kind == "table" {
				cols, err := intakeRepairColumns(obj, []float64{144, 90, 90, 90, 90}, []float64{180, 81, 81, 81, 81})
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				if cols {
					resolution = "wmds.v4.capability-callouts-label-column-180"
				}
			}
			if kind == "block" && intakeRepairNumber(obj, "w") == 18 {
				switch intakeRepairNumber(obj, "x") {
				case 471:
					obj["x"] = 480
					resolution = "wmds.v4.capability-callouts-badge-cell-anchor"
				case 381:
					obj["x"] = 399
					resolution = "wmds.v4.capability-callouts-badge-cell-anchor"
				}
			}
			if kind == "callout" && obj["label"] == nil {
				// This specimen intentionally has no eyebrow. A bounded metric
				// card retains its value/caption without inventing a label.
				for field := range obj {
					switch field {
					case "type", "x", "y", "w", "value", "text":
					default:
						return fmt.Errorf("intake.repair_unexpected_unlabeled_callout_field: %s", field)
					}
				}
				obj = map[string]any{"type": "card", "x": obj["x"], "y": obj["y"], "w": obj["w"], "h": 126, "surface": "callout", "pad": 18, "metric": map[string]any{"value": obj["value"], "label": obj["text"]}}
				resolution = "wmds.v4.capability-callout-without-eyebrow"
			}
		case "capability-heat/split":
			if kind == "bullets" && intakeRepairNumber(obj, "x") == 57 && intakeRepairNumber(obj, "y") == 162 {
				obj["y"] = 180
				resolution = "wmds.v4.capability-split-short-body-start-180"
			}
		case "capability-heat/nav":
			if kind == "table" && intakeRepairNumber(obj, "y") == 156 && intakeRepairNumber(obj, "rowH") == 36 {
				obj["y"], obj["rowH"] = 144, 39
				resolution = "wmds.v4.capability-nav-subtitle-row-39"
			}
			if kind == "legend" {
				obj["size"] = 8
				resolution = "wmds.v4.capability-nav-legend-8"
			}
		case "capacity-heat/team-weeks":
			if kind == "card" && intakeRepairNumber(obj, "y") == 378 && intakeRepairNumber(obj, "h") == 72 {
				obj["pad"] = 12
				resolution = "wmds.v4.capacity-callout-padding-12"
			}
		case "capacity-heat/sprint":
			if kind == "cardrow" && intakeRepairNumber(obj, "y") == 360 && intakeRepairNumber(obj, "h") == 90 {
				obj["y"], obj["h"] = 354, 96
				resolution = "wmds.v4.capacity-sprint-card-allocation-96"
			}
			if kind == "legend" {
				obj["size"] = 8
				resolution = "wmds.v4.capacity-sprint-legend-8"
			}
		case "heat-tile-map/insight-cards":
			if kind == "card" && intakeRepairNumber(obj, "h") == 90 {
				obj["h"] = 96
				resolution = "wmds.v4.heat-insight-card-allocation-96"
			}
		case "compact-heat/left-panel":
			if kind == "table" {
				changed, err := intakeRepairColumns(obj, []float64{156, 102, 102, 102, 102}, []float64{156, 100.5, 100.5, 100.5, 100.5})
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				if changed {
					resolution = "wmds.v4.compact-heat-column-sum-558"
				}
			}
		case "vendors/rated-matrix":
			if kind == "text" {
				text, _ := obj["text"].(string)
				if strings.ContainsAny(text, "●○") {
					items, err := intakeRepairGlyphLegend(text)
					if err != nil {
						return fmt.Errorf("%s: %w", key, err)
					}
					obj = map[string]any{"type": "scorelegend", "x": obj["x"], "y": obj["y"], "w": 432, "max": 4, "items": items}
					if n.Scene.Keys != nil {
						path := n.Scene.Path + "/items"
						n.Scene.Keys[path] = []string{"rating-1", "rating-2", "rating-3", "rating-4"}
					}
					resolution = "wmds.v4.vendor-score-legend-native-circles"
				}
			}
		}
		if resolution != "" {
			data, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			n.Scene.Node = data
			if !intakeRepairHasResolution(n.Scene.Resolutions, resolution) {
				n.Scene.Resolutions = append(n.Scene.Resolutions, resolution)
			}
		}
	}
	return nil
}

func intakeRepairNumber(obj map[string]any, key string) float64 {
	value, ok := discoveryNumber(obj[key])
	if !ok {
		return -1
	}
	return value
}

func intakeRepairHasResolution(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func intakeRepairColumns(obj map[string]any, before, after []float64) (bool, error) {
	cols, ok := obj["cols"].([]any)
	if !ok || len(cols) != len(before) {
		return false, fmt.Errorf("intake.repair_column_topology")
	}
	original, amended := true, true
	for i, raw := range cols {
		column, ok := raw.(map[string]any)
		if !ok {
			return false, fmt.Errorf("intake.repair_column_topology")
		}
		w := intakeRepairNumber(column, "w")
		original = original && w == before[i]
		amended = amended && w == after[i]
	}
	if amended {
		return false, nil
	}
	if !original {
		return false, fmt.Errorf("intake.repair_unexpected_column_widths")
	}
	for i, raw := range cols {
		raw.(map[string]any)["w"] = after[i]
	}
	return true, nil
}

func intakeRepairGlyphLegend(text string) ([]any, error) {
	var items []any
	rest := strings.TrimSpace(text)
	for score := 1; score <= 4; score++ {
		prefix := strings.Repeat("●", score) + strings.Repeat("○", 4-score)
		if !strings.HasPrefix(rest, prefix) {
			return nil, fmt.Errorf("intake.repair_glyph_legend_sequence")
		}
		rest = strings.TrimSpace(strings.TrimPrefix(rest, prefix))
		end := strings.IndexAny(rest, "●○")
		if end < 0 {
			end = len(rest)
		}
		label := strings.TrimSpace(rest[:end])
		if label == "" {
			return nil, fmt.Errorf("intake.repair_glyph_legend_label_%s", strconv.Itoa(score))
		}
		items = append(items, map[string]any{"value": score, "label": label})
		rest = strings.TrimSpace(rest[end:])
	}
	if rest != "" {
		return nil, fmt.Errorf("intake.repair_glyph_legend_trailing_content")
	}
	return items, nil
}
