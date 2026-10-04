package wmdesign

import (
	"encoding/json"
	"fmt"
)

// applyIncomingTableFitRepairs is called only by the atomic v4 intake gate.
// It changes allocations, never supplied cell content or measured fit rules.
func applyIncomingTableFitRepairs(key string, slide *SlideSpec) error {
	type allocation struct {
		path, kind, receipt         string
		before, after               map[string]float64
		columnsBefore, columnsAfter []float64
		columnKeys, columnTypes     []string
		rows                        int
	}
	var a allocation
	switch key {
	case "maturity/table-left":
		a = allocation{path: "/body/1", kind: "table", receipt: "wmds.v4.maturity-row-header-allocation-108", before: map[string]float64{"x": 393, "y": 126, "w": 510, "rowH": 72}, columnsBefore: []float64{90, 140, 140, 140}, columnsAfter: []float64{108, 134, 134, 134}, columnKeys: []string{"r", "a", "b", "c"}, columnTypes: []string{"", "", "", ""}, rows: 4}
	case "stakeholders/quadrant-split":
		a = allocation{path: "/body/0", kind: "schedule", receipt: "wmds.v4.stakeholder-schedule-row-allocation-21_4", before: map[string]float64{"x": 57, "y": 180, "w": 270, "keyW": 36, "rowHeight": 21}, after: map[string]float64{"rowHeight": 21.4}, rows: 10}
	case "readiness/scorecard":
		a = allocation{path: "/body/0", kind: "table", receipt: "wmds.v4.readiness-column-allocation-594", before: map[string]float64{"x": 57, "y": 126, "w": 594, "rowH": 42}, columnsBefore: []float64{216, 126, 72, 108}, columnsAfter: []float64{270, 126, 90, 108}, columnKeys: []string{"a", "b", "c", "d"}, columnTypes: []string{"", "dots", "num", "status"}, rows: 7}
	case "activities/by-phase-table":
		a = allocation{path: "/body/0", kind: "table", receipt: "wmds.v4.activities-weeks-column-allocation-90", before: map[string]float64{"x": 57, "y": 126, "w": 846, "rowH": 26}, columnsBefore: []float64{108, 192, 330, 144, 72}, columnsAfter: []float64{108, 192, 312, 144, 90}, columnKeys: []string{"ph", "s", "a", "d", "w"}, columnTypes: []string{"", "", "", "", "num"}, rows: 12}
	case "activities/ownership-split":
		a = allocation{path: "/body/2", kind: "table", receipt: "wmds.v4.ownership-column-allocation", before: map[string]float64{"x": 345, "y": 36, "w": 558, "rowH": 33}, columnsBefore: []float64{168, 96, 102, 120, 72}, columnsAfter: []float64{156, 102, 102, 108, 90}, columnKeys: []string{"a", "p", "w", "n", "k"}, columnTypes: []string{"", "tag", "", "", "num"}, rows: 12}
	case "value-tracking/planned-vs-realized":
		a = allocation{path: "/body/5", kind: "table", receipt: "wmds.v4.realized-value-numeric-column-allocation", before: map[string]float64{"x": 489, "y": 234, "w": 414, "rowH": 36}, columnsBefore: []float64{168, 72, 72, 102}, columnsAfter: []float64{174, 78, 78, 84}, columnKeys: []string{"l", "p", "r", "s"}, columnTypes: []string{"", "num", "num", "status"}, rows: 5}
	default:
		return nil
	}
	if slide == nil {
		return fmt.Errorf("intake.tablefit_nil_slide")
	}
	found := false
	for i := range slide.Nodes {
		scene := slide.Nodes[i].Scene
		if scene == nil || scene.Path != a.path {
			continue
		}
		if found {
			return fmt.Errorf("intake.tablefit_duplicate_target: %s", key)
		}
		found = true
		obj, err := libraryObject(scene.Node)
		if err != nil {
			return err
		}
		if obj["type"] != a.kind {
			return fmt.Errorf("intake.tablefit_type: %s", key)
		}
		field := "rows"
		if a.kind == "schedule" {
			field = "items"
		}
		rows, ok := obj[field].([]any)
		if !ok || len(rows) != a.rows {
			return fmt.Errorf("intake.tablefit_row_topology: %s", key)
		}
		rowKeys := a.columnKeys
		if a.kind == "schedule" {
			rowKeys = []string{"k", "t"}
		}
		for _, raw := range rows {
			row, ok := raw.(map[string]any)
			if !ok || len(row) != len(rowKeys) {
				return fmt.Errorf("intake.tablefit_row_fields: %s", key)
			}
			for _, field := range rowKeys {
				if _, ok := row[field]; !ok {
					return fmt.Errorf("intake.tablefit_row_fields: %s", key)
				}
			}
		}
		for _, field := range []string{"x", "y", "w", "keyW", "rowHeight", "rowH"} {
			before, ok := a.before[field]
			if !ok {
				continue
			}
			want := before
			if value, ok := a.after[field]; ok && intakeRepairNumber(obj, field) == value {
				want = value
			}
			if intakeRepairNumber(obj, field) != want {
				return fmt.Errorf("intake.tablefit_geometry: %s.%s", key, field)
			}
		}
		if a.columnKeys != nil {
			cols, ok := obj["cols"].([]any)
			if !ok || len(cols) != len(a.columnKeys) {
				return fmt.Errorf("intake.tablefit_columns: %s", key)
			}
			for j, raw := range cols {
				c, ok := raw.(map[string]any)
				kind, _ := c["type"].(string)
				declared, exists := c["type"]
				if !ok || c["k"] != a.columnKeys[j] || kind != a.columnTypes[j] || exists && declared != a.columnTypes[j] {
					return fmt.Errorf("intake.tablefit_column_keys: %s", key)
				}
			}
			if _, err = intakeRepairColumns(obj, a.columnsBefore, a.columnsAfter); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
		for field, value := range a.after {
			obj[field] = value
		}
		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		scene.Node = data
		if !intakeRepairHasResolution(scene.Resolutions, a.receipt) {
			scene.Resolutions = append(scene.Resolutions, a.receipt)
		}
	}
	if !found {
		return fmt.Errorf("intake.tablefit_missing_target: %s", key)
	}
	return nil
}
