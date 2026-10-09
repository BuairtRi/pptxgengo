package deckproject

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
	"strings"
)

const ProcessPatchSchema = "pptxgengo.process-patch.v1"

type ProcessOperation struct {
	Action     string                `json:"action"`
	Entity     string                `json:"entity"`
	Key        string                `json:"key,omitempty"`
	Cascade    bool                  `json:"cascade,omitempty"`
	Order      []string              `json:"order,omitempty"`
	Lane       *wmdesign.ProcessLane `json:"lane,omitempty"`
	Step       *wmdesign.ProcessStep `json:"step,omitempty"`
	Link       *wmdesign.ProcessLink `json:"link,omitempty"`
	Model      *wmdesign.ProcessSpec `json:"model,omitempty"`
	Columns    *int                  `json:"columns,omitempty"`
	LabelWidth *float64              `json:"label_width_pt,omitempty"`
	StepHeight *float64              `json:"step_height_pt,omitempty"`
	Rect       *wmdesign.Rect        `json:"rect,omitempty"`
}
type ProcessPatch struct {
	Schema               string             `json:"schema"`
	ExpectedSourceSHA256 string             `json:"expected_source_sha256"`
	Actor                string             `json:"actor"`
	Reason               string             `json:"reason"`
	NodeID               string             `json:"node_id"`
	Operations           []ProcessOperation `json:"operations"`
}
type ProcessInspection struct {
	Schema          string               `json:"schema"`
	SlideID         string               `json:"slide_id"`
	NodeID          string               `json:"node_id"`
	SourceSHA256    string               `json:"source_sha256"`
	Model           wmdesign.ProcessSpec `json:"model"`
	MutationBlocked string               `json:"mutation_blocked,omitempty"`
	RenderError     string               `json:"render_error,omitempty"`
	Geometry        DiagramInspection    `json:"geometry"`
}

func processFields(a, e string) (map[string]bool, error) {
	s := "action entity"
	switch a + "/" + e {
	case "materialize/source":
	case "initialize/source":
		s += " model cascade"
	case "set/lane":
		s += " key lane"
	case "set/step":
		s += " key step"
	case "set/link":
		s += " key link"
	case "remove/lane", "remove/step":
		s += " key cascade"
	case "remove/link", "set/start", "set/current", "set/end", "remove/end":
		s += " key"
	case "remove/start", "remove/current":
	case "reorder/lane", "reorder/step", "reorder/link":
		s += " order"
	case "set/layout":
		s += " columns label_width_pt step_height_pt rect"
	default:
		return nil, fmt.Errorf("unsupported process %s/%s", a, e)
	}
	out := map[string]bool{}
	for _, k := range strings.Fields(s) {
		out[k] = true
	}
	return out, nil
}
func DecodeProcessPatch(raw []byte, file string) (ProcessPatch, error) {
	var out ProcessPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("process patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty process patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("one process patch document required")
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
	obj := v.(map[string]any)
	ops, _ := obj["operations"].([]any)
	for i, op := range out.Operations {
		allowed, e := processFields(op.Action, op.Entity)
		if e != nil {
			return out, e
		}
		authored := ops[i].(map[string]any)
		if model, ok := authored["model"].(map[string]any); ok {
			for _, field := range []string{"type", "x", "y", "w", "h"} {
				if _, exists := model[field]; exists {
					return out, fmt.Errorf("process model geometry belongs to placement; authored %s is not accepted", field)
				}
			}
		}
		for k := range authored {
			if !allowed[k] {
				return out, fmt.Errorf("process operation does not accept field %s", k)
			}
		}
	}
	_, e = hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != ProcessPatchSchema || len(out.ExpectedSourceSHA256) != 64 || e != nil || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || len(out.Actor) > 256 || len(out.Reason) > 4096 || !stableID.MatchString(out.NodeID) || len(out.Operations) < 1 || len(out.Operations) > 500 {
		return out, fmt.Errorf("process patch requires schema, source SHA256, actor, reason, node_id and 1..500 operations")
	}
	return out, nil
}
func processSource(n *Node, values map[string]any) (wmdesign.ProcessSpec, bool, error) {
	var s wmdesign.ProcessSpec
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	if n.Definition == nil || n.Definition.ID != "wmds/component/process" {
		return s, bound, fmt.Errorf("node is not a semantic process; explicitly initialize a local leaf model")
	}
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return s, bound, e
	}
	delete(args, wmdesign.SceneSourceGeometryArgument)
	if e = strictInto(args, &s); e != nil {
		return s, bound, e
	}
	return s, bound, wmdesign.ValidateProcess(s)
}
func InspectProcess(p *Project, slideID, nodeID, bundle, engine string) (ProcessInspection, error) {
	out := ProcessInspection{Schema: "pptxgengo.process-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	list, i := findDiagramNode(&t.Nodes, nodeID)
	if list == nil {
		return out, fmt.Errorf("unknown process node")
	}
	var bound bool
	out.Model, bound, e = processSource(&(*list)[i], p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Bound source: edit values or explicitly materialize/source first"
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func processRemoveStep(m *wmdesign.ProcessSpec, key string, cascade bool) error {
	idx := -1
	for i, s := range m.Steps {
		if s.Key == key {
			idx = i
		}
	}
	if idx < 0 {
		return fmt.Errorf("unknown step %s", key)
	}
	dependent := m.Start == key || m.Current == key
	for _, k := range m.End {
		dependent = dependent || k == key
	}
	for _, l := range m.Links {
		dependent = dependent || l.From == key || l.To == key
	}
	if dependent && !cascade {
		return fmt.Errorf("step has incident links/markers; explicitly cascade or remove/reassign first")
	}
	m.Steps = append(m.Steps[:idx], m.Steps[idx+1:]...)
	links := []wmdesign.ProcessLink{}
	for _, l := range m.Links {
		if l.From != key && l.To != key {
			links = append(links, l)
		}
	}
	m.Links = links
	if m.Start == key {
		m.Start = ""
	}
	if m.Current == key {
		m.Current = ""
	}
	ends := []string{}
	for _, k := range m.End {
		if k != key {
			ends = append(ends, k)
		}
	}
	m.End = ends
	return nil
}
func applyProcessOperation(m *wmdesign.ProcessSpec, n *Node, op ProcessOperation) error {
	if op.Key != "" && !stableID.MatchString(op.Key) {
		return fmt.Errorf("invalid process key")
	}
	if (op.Action == "set" || op.Action == "remove") && op.Entity != "layout" && op.Entity != "start" && op.Entity != "current" && op.Key == "" {
		return fmt.Errorf("process operation requires key")
	}
	if op.Action == "set" && (op.Entity == "start" || op.Entity == "current") && op.Key == "" {
		return fmt.Errorf("marker requires key")
	}
	switch op.Action + "/" + op.Entity {
	case "set/lane":
		if op.Lane == nil || op.Key != op.Lane.Key {
			return fmt.Errorf("matching complete lane required")
		}
		for i, v := range m.Lanes {
			if v.Key == op.Key {
				m.Lanes[i] = *op.Lane
				return nil
			}
		}
		m.Lanes = append(m.Lanes, *op.Lane)
	case "set/step":
		if op.Step == nil || op.Key != op.Step.Key {
			return fmt.Errorf("matching complete step required")
		}
		for i, v := range m.Steps {
			if v.Key == op.Key {
				m.Steps[i] = *op.Step
				return nil
			}
		}
		m.Steps = append(m.Steps, *op.Step)
	case "set/link":
		if op.Link == nil || op.Key != op.Link.Key {
			return fmt.Errorf("matching complete link required")
		}
		for i, v := range m.Links {
			if v.Key == op.Key {
				m.Links[i] = *op.Link
				return nil
			}
		}
		m.Links = append(m.Links, *op.Link)
	case "remove/step":
		return processRemoveStep(m, op.Key, op.Cascade)
	case "remove/lane":
		idx := -1
		for i, v := range m.Lanes {
			if v.Key == op.Key {
				idx = i
			}
		}
		if idx < 0 {
			return fmt.Errorf("unknown lane")
		}
		keys := []string{}
		for _, s := range m.Steps {
			if s.Lane == op.Key {
				keys = append(keys, s.Key)
			}
		}
		if len(keys) > 0 && !op.Cascade {
			return fmt.Errorf("populated lane removal requires cascade")
		}
		for _, k := range keys {
			if e := processRemoveStep(m, k, true); e != nil {
				return e
			}
		}
		m.Lanes = append(m.Lanes[:idx], m.Lanes[idx+1:]...)
	case "remove/link":
		for i, v := range m.Links {
			if v.Key == op.Key {
				m.Links = append(m.Links[:i], m.Links[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("unknown link")
	case "set/start":
		m.Start = op.Key
	case "set/current":
		m.Current = op.Key
	case "remove/start":
		m.Start = ""
	case "remove/current":
		m.Current = ""
	case "set/end":
		for _, k := range m.End {
			if k == op.Key {
				return fmt.Errorf("end already marked")
			}
		}
		m.End = append(m.End, op.Key)
	case "remove/end":
		for i, k := range m.End {
			if k == op.Key {
				m.End = append(m.End[:i], m.End[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("unknown end marker")
	case "set/layout":
		if op.Columns != nil {
			m.Columns = *op.Columns
		}
		if op.LabelWidth != nil {
			m.LabelWidth = *op.LabelWidth
		}
		if op.StepHeight != nil {
			m.StepHeight = *op.StepHeight
		}
		if op.Rect != nil {
			n.Placement.Rect = op.Rect
		}
	case "reorder/lane", "reorder/step", "reorder/link":
		seen := map[string]bool{}
		for _, k := range op.Order {
			if seen[k] {
				return fmt.Errorf("duplicate reorder key")
			}
			seen[k] = true
		}
		switch op.Entity {
		case "lane":
			if len(op.Order) != len(m.Lanes) {
				return fmt.Errorf("reorder requires every lane")
			}
			a := []wmdesign.ProcessLane{}
			for _, k := range op.Order {
				found := false
				for _, v := range m.Lanes {
					if k == v.Key {
						a = append(a, v)
						found = true
					}
				}
				if !found {
					return fmt.Errorf("unknown reorder lane")
				}
			}
			m.Lanes = a
		case "step":
			if len(op.Order) != len(m.Steps) {
				return fmt.Errorf("reorder requires every step")
			}
			a := []wmdesign.ProcessStep{}
			for _, k := range op.Order {
				found := false
				for _, v := range m.Steps {
					if k == v.Key {
						a = append(a, v)
						found = true
					}
				}
				if !found {
					return fmt.Errorf("unknown reorder step")
				}
			}
			m.Steps = a
		case "link":
			if len(op.Order) != len(m.Links) {
				return fmt.Errorf("reorder requires every link")
			}
			a := []wmdesign.ProcessLink{}
			for _, k := range op.Order {
				found := false
				for _, v := range m.Links {
					if k == v.Key {
						a = append(a, v)
						found = true
					}
				}
				if !found {
					return fmt.Errorf("unknown reorder link")
				}
			}
			m.Links = a
		}
	default:
		return fmt.Errorf("source initialization/materialization must be first")
	}
	return nil
}
func PatchProcess(p *Project, slideID string, patch ProcessPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeProcessPatch(canonical(patch), "process-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("process source hash mismatch")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	list, i := findDiagramNode(&clone.Nodes, patch.NodeID)
	if list == nil {
		return empty, fmt.Errorf("unknown process node")
	}
	n := &(*list)[i]
	if n.Placement == nil || n.Placement.Rect == nil || len(n.Nodes) > 0 {
		return empty, fmt.Errorf("process needs a local rectangular leaf")
	}
	m, bound, e := processSource(n, p.Document.Slides[idx].Values)
	if patch.Operations[0].Action == "initialize" {
		op := patch.Operations[0]
		if op.Entity != "source" || op.Model == nil || !op.Cascade {
			return empty, fmt.Errorf("initialize requires explicit complete model and cascade acknowledgement")
		}
		if n.Definition != nil && n.Definition.ID == "wmds/component/process" {
			return empty, fmt.Errorf("already initialized process")
		}
		m = *op.Model
		bound = false
		e = nil
		n.Definition = &Reference{Scope: "shared", ID: "wmds/component/process"}
	}
	if e != nil {
		return empty, e
	}
	for j, op := range patch.Operations {
		if op.Action == "initialize" && j == 0 {
			continue
		}
		if op.Action == "materialize" && op.Entity == "source" && j == 0 {
			bound = false
			continue
		}
		if bound {
			return empty, fmt.Errorf("bound process requires materialize/source first")
		}
		if e = applyProcessOperation(&m, n, op); e != nil {
			return empty, e
		}
	}
	if e = wmdesign.ValidateProcess(m); e != nil {
		return empty, e
	}
	m.Type = ""
	m.X, m.Y, m.W, m.H = 0, 0, 0, 0
	var args map[string]any
	_ = json.Unmarshal(canonical(m), &args)
	delete(args, "type")
	if patch.Operations[0].Action != "initialize" {
		if geometry, exists := n.Arguments[wmdesign.SceneSourceGeometryArgument]; exists {
			args[wmdesign.SceneSourceGeometryArgument] = geometry
		}
	}
	n.Arguments = args
	n.Keys = nil
	return CompositionCandidate(p, slideID, "process", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
