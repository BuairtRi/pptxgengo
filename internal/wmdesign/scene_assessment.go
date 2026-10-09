package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// AssessmentSpec owns the meaning of an ordinal heat matrix and its legend.
// A nil or absent score means not assessed; zero is a measured lowest score.
type AssessmentAxis struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}
type AssessmentDomain struct {
	Max          int      `json:"max"`
	Scale        string   `json:"scale"`
	Labels       []string `json:"labels"`
	MissingLabel string   `json:"missing_label"`
}
type AssessmentRow struct {
	Key    string          `json:"key"`
	Label  string          `json:"label"`
	Scores map[string]*int `json:"scores"`
}
type AssessmentSpec struct {
	Type       string           `json:"type,omitempty"`
	X          float64          `json:"x,omitempty"`
	Y          float64          `json:"y,omitempty"`
	W          float64          `json:"w,omitempty"`
	H          float64          `json:"h,omitempty"`
	RowLabel   string           `json:"row_label"`
	LabelWidth float64          `json:"label_width_pt"`
	RowHeight  float64          `json:"row_height_pt"`
	ShowScores bool             `json:"show_scores"`
	Dense      bool             `json:"dense,omitempty"`
	Domain     AssessmentDomain `json:"domain"`
	Columns    []AssessmentAxis `json:"columns"`
	Rows       []AssessmentRow  `json:"rows"`
}

func ValidateAssessment(s AssessmentSpec) error {
	if s.Domain.Max < 1 || s.Domain.Max > 4 || len(s.Domain.Labels) != s.Domain.Max+1 || (s.Domain.Scale != "seq" && s.Domain.Scale != "risk") || strings.TrimSpace(s.Domain.MissingLabel) == "" {
		return fmt.Errorf("assessment requires integer 0..max domain (max 1..4), seq/risk scale, one label per score and a missing label")
	}
	for _, label := range s.Domain.Labels {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("assessment domain labels cannot be blank")
		}
	}
	if len(s.Columns) < 1 || len(s.Columns) > 11 || len(s.Rows) < 1 || len(s.Rows) > 60 || strings.TrimSpace(s.RowLabel) == "" || s.LabelWidth <= 24 || s.RowHeight < 24 || math.IsNaN(s.LabelWidth+s.RowHeight) || math.IsInf(s.LabelWidth+s.RowHeight, 0) {
		return fmt.Errorf("assessment requires 1..11 columns, 1..60 rows, row label, label width >24pt and row height >=24pt")
	}
	columns, rows := map[string]bool{}, map[string]bool{}
	for _, c := range s.Columns {
		if !validPartKey(c.Key) || (c.Key == "assessment-label" || c.Key == "group" || c.Key == "total" || c.Key == "ink" || c.Key == "scale" || c.Key == "h") || strings.TrimSpace(c.Label) == "" || columns[c.Key] {
			return fmt.Errorf("assessment invalid or duplicate column %s", c.Key)
		}
		columns[c.Key] = true
	}
	for _, row := range s.Rows {
		if !validPartKey(row.Key) || strings.TrimSpace(row.Label) == "" || rows[row.Key] {
			return fmt.Errorf("assessment invalid or duplicate row %s", row.Key)
		}
		rows[row.Key] = true
		for key, score := range row.Scores {
			if !columns[key] {
				return fmt.Errorf("assessment row %s references unknown column %s", row.Key, key)
			}
			if score != nil && (*score < 0 || *score > s.Domain.Max) {
				return fmt.Errorf("assessment score %s/%s outside 0..%d", row.Key, key, s.Domain.Max)
			}
		}
	}
	return nil
}

func (r *renderer) planAssessmentScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type != "assessment" {
		return nil, false, nil
	}
	var s AssessmentSpec
	if e := sceneDecode(raw, &s); e != nil {
		return nil, true, e
	}
	if e := ValidateAssessment(s); e != nil {
		return nil, true, e
	}
	if s.W <= s.LabelWidth || s.H <= 0 || math.IsNaN(s.X+s.Y+s.W+s.H) || math.IsInf(s.X+s.Y+s.W+s.H, 0) {
		return nil, true, fmt.Errorf("assessment invalid allocation")
	}
	width := (s.W - s.LabelWidth) / float64(len(s.Columns))
	max := float64(s.Domain.Max)
	table := sceneTableSource{Type: "table", X: s.X, Y: s.Y, W: s.W, Header: "light", Dense: s.Dense, RowH: s.RowHeight, Columns: []sceneTableColumn{{Key: "assessment-label", Label: s.RowLabel, Width: s.LabelWidth}}}
	table.RowHeader = sceneTableRowHeader(true)
	for _, col := range s.Columns {
		table.Columns = append(table.Columns, sceneTableColumn{Key: col.Key, Label: col.Label, Width: width, Type: "heat", Scale: s.Domain.Scale, Max: &max, ShowValue: s.ShowScores})
	}
	tctx := ctx
	tctx.Keys = map[string][]string{}
	tctx.Keys[strings.TrimRight(ctx.Path, "/")+"/rows"] = []string{}
	for _, row := range s.Rows {
		values := map[string]json.RawMessage{}
		values["assessment-label"], _ = json.Marshal(row.Label)
		for _, col := range s.Columns {
			values[col.Key], _ = json.Marshal(row.Scores[col.Key])
		}
		table.Rows = append(table.Rows, values)
		tctx.Keys[strings.TrimRight(ctx.Path, "/")+"/rows"] = append(tctx.Keys[strings.TrimRight(ctx.Path, "/")+"/rows"], row.Key)
	}
	p, e := r.sceneTable(id+".matrix", table, tctx)
	if e != nil {
		return nil, true, e
	}
	legend := peopleSpec{Type: "legend", X: s.X, Y: p.Bounds.Y + p.Bounds.H + 12, W: s.W, Layout: "horizontal"}
	for value, label := range s.Domain.Labels {
		v := float64(value)
		legend.Items = append(legend.Items, peopleLegendItem{Key: fmt.Sprintf("score-%d", value), Text: fmt.Sprintf("%d: %s", value, label), Heat: &v, HeatMax: &max, HeatScale: s.Domain.Scale})
	}
	legend.Items = append(legend.Items, peopleLegendItem{Key: "missing", Text: "Blank: " + s.Domain.MissingLabel, Color: "bg"})
	lraw, _ := json.Marshal(map[string]any{"type": "legend", "x": legend.X, "y": legend.Y, "w": legend.W, "layout": legend.Layout, "items": legend.Items})
	lctx := ctx
	lctx.Keys = nil
	l, _, e := r.planPeopleScene(id+".legend", lraw, lctx)
	if e != nil {
		return nil, true, e
	}
	p.Items = append(p.Items, l.Items...)
	p.Groups = append(p.Groups, l.Groups...)
	p.Warnings = append(p.Warnings, l.Warnings...)
	p.Bounds = diagramUnion(p.Bounds, l.Bounds)
	p.Definition = "scene.assessment"
	if p.Bounds.X < s.X-.01 || p.Bounds.Y < s.Y-.01 || p.Bounds.X+p.Bounds.W > s.X+s.W+.01 || p.Bounds.Y+p.Bounds.H > s.Y+s.H+.01 {
		return nil, true, fmt.Errorf("assessment content/legend exceeds allocation; widen, raise row height/allocation, or split the assessment")
	}
	return p, true, nil
}
