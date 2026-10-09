package deckproject

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

const AssessmentPatchSchema = "pptxgengo.assessment-patch.v1"

type AssessmentOperation struct {
	Action     string                     `json:"action"`
	Entity     string                     `json:"entity"`
	Key        string                     `json:"key,omitempty"`
	Column     string                     `json:"column,omitempty"`
	Label      string                     `json:"label,omitempty"`
	Score      *int                       `json:"score,omitempty"`
	Missing    bool                       `json:"missing,omitempty"`
	Cascade    bool                       `json:"cascade,omitempty"`
	Order      []string                   `json:"order,omitempty"`
	Model      *wmdesign.AssessmentSpec   `json:"model,omitempty"`
	LegendNode string                     `json:"legend_node,omitempty"`
	Domain     *wmdesign.AssessmentDomain `json:"domain,omitempty"`
	LabelWidth *float64                   `json:"label_width_pt,omitempty"`
	ShowScores *bool                      `json:"show_scores,omitempty"`
	RowHeight  *float64                   `json:"row_height_pt,omitempty"`
}
type AssessmentPatch struct {
	Schema               string                `json:"schema"`
	ExpectedSourceSHA256 string                `json:"expected_source_sha256"`
	Actor                string                `json:"actor"`
	Reason               string                `json:"reason"`
	NodeID               string                `json:"node_id"`
	Operations           []AssessmentOperation `json:"operations"`
}
type AssessmentInspection struct {
	Schema          string                  `json:"schema"`
	SlideID         string                  `json:"slide_id"`
	NodeID          string                  `json:"node_id"`
	SourceSHA256    string                  `json:"source_sha256"`
	Contract        string                  `json:"contract"`
	Model           wmdesign.AssessmentSpec `json:"model"`
	MutationBlocked string                  `json:"mutation_blocked,omitempty"`
	RenderError     string                  `json:"render_error,omitempty"`
	Geometry        DiagramInspection       `json:"geometry"`
}

func assessmentFields(action, entity string) (map[string]bool, error) {
	fields := "action entity"
	switch {
	case action == "initialize" && entity == "source":
		fields += " model legend_node cascade"
	case action == "materialize" && entity == "source":
	case action == "set" && (entity == "row" || entity == "column"):
		fields += " key label"
	case action == "remove" && (entity == "row" || entity == "column"):
		fields += " key cascade"
	case action == "reorder" && (entity == "row" || entity == "column"):
		fields += " order"
	case action == "set" && entity == "score":
		fields += " key column score missing"
	case action == "set" && entity == "domain":
		fields += " domain"
	case action == "set" && entity == "layout":
		fields += " label_width_pt row_height_pt show_scores"
	default:
		return nil, fmt.Errorf("unsupported assessment operation %s %s", action, entity)
	}
	out := map[string]bool{}
	for _, field := range strings.Fields(fields) {
		out[field] = true
	}
	return out, nil
}
func DecodeAssessmentPatch(raw []byte, file string) (AssessmentPatch, error) {
	var out AssessmentPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("assessment patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty assessment patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("assessment patch requires one document")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = p.shapeType(v, reflect.TypeOf(out), ""); e != nil {
		return out, e
	}
	if e = strictInto(v, &out); e != nil {
		return out, e
	}
	object, _ := v.(map[string]any)
	ops, _ := object["operations"].([]any)
	for i, rawop := range ops {
		fields, ok := rawop.(map[string]any)
		if !ok {
			return out, fmt.Errorf("assessment operation requires object")
		}
		allowed, e := assessmentFields(out.Operations[i].Action, out.Operations[i].Entity)
		if e != nil {
			return out, e
		}
		if model, ok := fields["model"].(map[string]any); ok {
			for _, reserved := range []string{"type", "x", "y", "w", "h"} {
				if _, present := model[reserved]; present {
					return out, fmt.Errorf("assessment model cannot author reserved geometry field %s", reserved)
				}
			}
		}
		for key := range fields {
			if !allowed[key] {
				return out, fmt.Errorf("assessment %s %s does not accept authored field %s", out.Operations[i].Action, out.Operations[i].Entity, key)
			}
		}
	}
	_, e = hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != AssessmentPatchSchema || len(out.ExpectedSourceSHA256) != 64 || e != nil || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || strings.TrimSpace(out.Reason) == "" || len(out.Reason) > 4096 || !stableID.MatchString(out.NodeID) || len(out.Operations) == 0 || len(out.Operations) > 500 {
		return out, fmt.Errorf("assessment patch requires schema, source SHA256, actor, reason, node_id and 1..500 operations")
	}
	return out, nil
}
func assessmentNode(t *LocalTemplate, id string) (*Node, error) {
	list, i := findDiagramNode(&t.Nodes, id)
	if list == nil {
		return nil, fmt.Errorf("unknown assessment node %s", id)
	}
	return &(*list)[i], nil
}
func assessmentSource(n *Node, values map[string]any) (wmdesign.AssessmentSpec, bool, error) {
	var out wmdesign.AssessmentSpec
	if n.Definition == nil || n.Definition.ID != "wmds/component/assessment" {
		return out, false, fmt.Errorf("node %s has no assessment domain; initialize an explicitly selected local heat table first", n.ID)
	}
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return out, bound, e
	}
	delete(args, wmdesign.SceneSourceGeometryArgument)
	if e = strictInto(args, &out); e != nil {
		return out, bound, e
	}
	return out, bound, wmdesign.ValidateAssessment(out)
}
func InspectAssessment(p *Project, slideID, nodeID, bundle, engine string) (AssessmentInspection, error) {
	out := AssessmentInspection{Schema: "pptxgengo.assessment-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash(), Contract: "Ordinal integer scores 0..max; nil/absent scores are not assessed and remain distinct from numeric zero. No averaging, weighting or percentage inference. Receipt-backed native visible score edits require explicit project assessment reconcile decisions; geometry alone does not establish a score."}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	n, e := assessmentNode(&t, nodeID)
	if e != nil {
		return out, e
	}
	var bound bool
	out.Model, bound, e = assessmentSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Bound source: edit bindings or explicitly begin with materialize source."
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
		out.Geometry.Validation = "renderer_failed_no_measurement_qualification"
	}
	return out, nil
}
func PatchAssessment(p *Project, slideID string, patch AssessmentPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeAssessmentPatch(canonical(patch), "assessment-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("assessment source hash mismatch; inspect again")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := assessmentNode(&clone, patch.NodeID)
	if e != nil {
		return empty, e
	}
	var s wmdesign.AssessmentSpec
	var bound bool
	if patch.Operations[0].Action == "initialize" {
		op := patch.Operations[0]
		if op.Model == nil {
			return empty, fmt.Errorf("initialize requires a complete assessment model")
		}
		if e = initializeAssessment(&clone, n, op, p.Document.Slides[idx].Values); e != nil {
			return empty, e
		}
		n, e = assessmentNode(&clone, patch.NodeID)
		if e != nil {
			return empty, e
		}
		s = *op.Model
	} else {
		s, bound, e = assessmentSource(n, p.Document.Slides[idx].Values)
		if e != nil {
			return empty, e
		}
	}
	for i, op := range patch.Operations {
		if op.Action == "initialize" || op.Action == "materialize" {
			if i != 0 {
				return empty, fmt.Errorf("assessment %s must be first", op.Action)
			}
			bound = false
			continue
		}
		if bound {
			return empty, fmt.Errorf("assessment refuses bound source without explicit first materialize source")
		}
		if e = applyAssessmentOperation(&s, op); e != nil {
			return empty, e
		}
	}
	if e = wmdesign.ValidateAssessment(s); e != nil {
		return empty, e
	}
	if s.Type != "" || s.X != 0 || s.Y != 0 || s.W != 0 || s.H != 0 {
		return empty, fmt.Errorf("assessment model geometry belongs in placement, not arguments")
	}
	var args map[string]any
	if e = json.Unmarshal(canonical(s), &args); e != nil {
		return empty, e
	}
	if geometry, exists := n.Arguments[wmdesign.SceneSourceGeometryArgument]; exists {
		args[wmdesign.SceneSourceGeometryArgument] = geometry
	}
	n.Arguments = args
	n.Keys = nil
	return CompositionCandidate(p, slideID, "assessment", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}

func initializeAssessment(t *LocalTemplate, n *Node, op AssessmentOperation, values map[string]any) error {
	if !op.Cascade {
		return fmt.Errorf("initialize replaces the source model; cascade: true must acknowledge replacement of existing cells")
	}
	if n.Definition == nil || n.Definition.ID != "wmds/component/table" || len(n.Nodes) > 0 || n.Placement == nil || n.Placement.Rect == nil {
		return fmt.Errorf("initialize requires a leaf local table with explicit rectangular placement")
	}
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return e
	}
	delete(args, wmdesign.SceneSourceGeometryArgument)
	for field := range args {
		switch field {
		case "header", "rowHeader", "dense", "rowH", "cols", "rows", "heatMin", "heatMax", "rowGroups", "groups":
		default:
			return fmt.Errorf("initialize refuses unsupported table treatment %s; retain it as a template-specific exception", field)
		}
	}
	var source struct {
		Rows      []map[string]json.RawMessage `json:"rows"`
		Header    string                       `json:"header"`
		RowHeader bool                         `json:"rowHeader"`
		HeatMax   *float64                     `json:"heatMax"`
		HeatMin   *float64                     `json:"heatMin"`
		Dense     bool                         `json:"dense"`
		RowHeight float64                      `json:"rowH"`
		RowGroups []any                        `json:"rowGroups"`
		Groups    []any                        `json:"groups"`
		Columns   []struct {
			Key        string   `json:"k"`
			Type       string   `json:"type"`
			Width      float64  `json:"w"`
			Scale      string   `json:"scale"`
			Min        *float64 `json:"min"`
			Max        *float64 `json:"max"`
			ShowScores bool     `json:"showValue"`
		} `json:"cols"`
	}
	if e = json.Unmarshal(canonical(args), &source); e != nil {
		return e
	}
	if len(source.Columns) < 2 || !source.RowHeader || source.Columns[0].Type != "" || len(source.RowGroups) > 0 || len(source.Groups) > 0 || source.Header == "dark" {
		return fmt.Errorf("initialize supports ungrouped light-header, styled row-header heat assessments only; grouped/status/quantitative tables need explicit domain support")
	}
	columnKeys := map[string]bool{}
	for _, c := range source.Columns {
		if c.Key == "group" || c.Key == "total" || c.Key == "ink" || c.Key == "scale" || c.Key == "h" {
			return fmt.Errorf("initialize refuses reserved legacy column key %s", c.Key)
		}
		columnKeys[c.Key] = true
	}
	for _, row := range source.Rows {
		for field := range row {
			if !columnKeys[field] {
				return fmt.Errorf("initialize refuses inline row metadata %s (group/total/row-scale/ink/height); use a dedicated grouped assessment contract", field)
			}
		}
	}
	for _, c := range source.Columns[1:] {
		scale := c.Scale
		if scale == "" {
			scale = "seq"
		}
		max := 4.0
		if source.HeatMax != nil {
			max = *source.HeatMax
		}
		if source.HeatMin != nil && *source.HeatMin != 0 {
			return fmt.Errorf("initialize requires zero-based heat domain")
		}
		if c.Max != nil {
			max = *c.Max
		}
		if c.Type != "heat" || c.Min != nil && *c.Min != 0 || scale != op.Model.Domain.Scale || max != float64(op.Model.Domain.Max) || c.ShowScores != op.Model.ShowScores {
			return fmt.Errorf("initialize requires zero-based heat columns with matching score domain, palette and show-scores treatment")
		}
	}
	rowHeight := source.RowHeight
	if rowHeight == 0 {
		rowHeight = 36
		if source.Dense {
			rowHeight = 28
		}
	}
	if op.Model.Dense != source.Dense || op.Model.LabelWidth != source.Columns[0].Width || op.Model.RowHeight != rowHeight {
		return fmt.Errorf("initialize must preserve table density, label width and row height; change layout explicitly afterward")
	}
	if op.LegendNode == "" || op.LegendNode == n.ID {
		return fmt.Errorf("initialize requires an explicitly named companion legend node")
	}
	list, i := findDiagramNode(&t.Nodes, op.LegendNode)
	if list == nil {
		return fmt.Errorf("unknown companion legend")
	}
	legend := (*list)[i]
	if legend.Definition == nil || legend.Definition.ID != "wmds/component/legend" || len(legend.Nodes) > 0 || legend.Placement == nil || legend.Placement.Rect == nil || legend.Placement.Zone != n.Placement.Zone {
		return fmt.Errorf("companion must be a leaf local legend with rectangular placement in the same zone")
	}
	// Keep the existing table origin/width and absorb only its named legend's
	// vertical allocation. Other nodes and the chrome remain unchanged.
	a, b := *n.Placement.Rect, *legend.Placement.Rect
	if b.X < a.X-.01 || b.X+b.W > a.X+a.W+.01 || b.Y < a.Y {
		return fmt.Errorf("companion legend must lie below table within its width")
	}
	if bottom := b.Y + b.H; bottom > a.Y+a.H {
		a.H = bottom - a.Y
	}
	*n.Placement.Rect = a
	n.Definition = &Reference{Scope: "shared", ID: "wmds/component/assessment"}
	*list = append((*list)[:i], (*list)[i+1:]...)
	return nil
}
