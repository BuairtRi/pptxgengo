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

const PortfolioPatchSchema = "pptxgengo.portfolio-patch.v1"

type PortfolioOperation struct {
	Action     string                        `json:"action"`
	Entity     string                        `json:"entity"`
	Key        string                        `json:"key,omitempty"`
	Cascade    bool                          `json:"cascade,omitempty"`
	Order      []string                      `json:"order,omitempty"`
	Horizon    *wmdesign.PortfolioHorizon    `json:"horizon,omitempty"`
	Status     *wmdesign.PortfolioStatus     `json:"status,omitempty"`
	Initiative *wmdesign.PortfolioInitiative `json:"initiative,omitempty"`
	Dependency *wmdesign.PortfolioDependency `json:"dependency,omitempty"`
	Model      *wmdesign.PortfolioSpec       `json:"model,omitempty"`
	CardHeight *float64                      `json:"card_height_pt,omitempty"`
	Rect       *wmdesign.Rect                `json:"rect,omitempty"`
}
type PortfolioPatch struct {
	Schema               string               `json:"schema"`
	ExpectedSourceSHA256 string               `json:"expected_source_sha256"`
	Actor                string               `json:"actor"`
	Reason               string               `json:"reason"`
	NodeID               string               `json:"node_id"`
	Operations           []PortfolioOperation `json:"operations"`
}
type PortfolioInspection struct {
	Schema          string                 `json:"schema"`
	SlideID         string                 `json:"slide_id"`
	NodeID          string                 `json:"node_id"`
	SourceSHA256    string                 `json:"source_sha256"`
	Model           wmdesign.PortfolioSpec `json:"model"`
	MutationBlocked string                 `json:"mutation_blocked,omitempty"`
	RenderError     string                 `json:"render_error,omitempty"`
	Geometry        DiagramInspection      `json:"geometry"`
}

func portfolioFields(a, e string) (map[string]bool, error) {
	s := "action entity"
	switch a + "/" + e {
	case "materialize/source":
	case "initialize/source":
		s += " model cascade"
	case "set/horizon":
		s += " key horizon"
	case "set/status":
		s += " key status"
	case "set/initiative":
		s += " key initiative"
	case "set/dependency":
		s += " key dependency"
	case "remove/horizon", "remove/initiative":
		s += " key cascade"
	case "remove/status", "remove/dependency":
		s += " key"
	case "reorder/horizon", "reorder/status", "reorder/initiative", "reorder/dependency":
		s += " order"
	case "set/layout":
		s += " card_height_pt rect"
	default:
		return nil, fmt.Errorf("unsupported portfolio %s/%s", a, e)
	}
	out := map[string]bool{}
	for _, k := range strings.Fields(s) {
		out[k] = true
	}
	return out, nil
}
func DecodePortfolioPatch(raw []byte, file string) (PortfolioPatch, error) {
	var out PortfolioPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("portfolio patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty portfolio patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("one portfolio patch document required")
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
	ops, _ := v.(map[string]any)["operations"].([]any)
	for i, op := range out.Operations {
		fields, e := portfolioFields(op.Action, op.Entity)
		if e != nil {
			return out, e
		}
		authored := ops[i].(map[string]any)
		if model, ok := authored["model"].(map[string]any); ok {
			for _, k := range []string{"type", "x", "y", "w", "h"} {
				if _, ok := model[k]; ok {
					return out, fmt.Errorf("portfolio model geometry belongs to placement")
				}
			}
		}
		for k := range authored {
			if !fields[k] {
				return out, fmt.Errorf("portfolio operation does not accept field %s", k)
			}
		}
	}
	_, e = hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != PortfolioPatchSchema || len(out.ExpectedSourceSHA256) != 64 || e != nil || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || len(out.Actor) > 256 || len(out.Reason) > 4096 || !stableID.MatchString(out.NodeID) || len(out.Operations) < 1 || len(out.Operations) > 500 {
		return out, fmt.Errorf("portfolio patch requires schema, source SHA256, actor, reason, node_id and 1..500 operations")
	}
	return out, nil
}
func portfolioSource(n *Node, values map[string]any) (wmdesign.PortfolioSpec, bool, error) {
	var s wmdesign.PortfolioSpec
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	if n.Definition == nil || n.Definition.ID != "wmds/component/portfolio" {
		return s, bound, fmt.Errorf("node is not a semantic portfolio; explicitly initialize local leaf")
	}
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return s, bound, e
	}
	delete(args, wmdesign.SceneSourceGeometryArgument)
	if e = strictInto(args, &s); e != nil {
		return s, bound, e
	}
	return s, bound, wmdesign.ValidatePortfolio(s)
}
func InspectPortfolio(p *Project, slideID, nodeID, bundle, engine string) (PortfolioInspection, error) {
	out := PortfolioInspection{Schema: "pptxgengo.portfolio-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	list, i := findDiagramNode(&t.Nodes, nodeID)
	if list == nil {
		return out, fmt.Errorf("unknown portfolio node")
	}
	var bound bool
	out.Model, bound, e = portfolioSource(&(*list)[i], p.Document.Slides[idx].Values)
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
func removePortfolioInitiative(m *wmdesign.PortfolioSpec, key string, cascade bool) error {
	idx := -1
	for i, v := range m.Initiatives {
		if v.Key == key {
			idx = i
		}
	}
	if idx < 0 {
		return fmt.Errorf("unknown initiative")
	}
	incident := false
	for _, e := range m.Dependencies {
		incident = incident || e.From == key || e.To == key
	}
	if incident && !cascade {
		return fmt.Errorf("incident dependencies require explicit cascade")
	}
	m.Initiatives = append(m.Initiatives[:idx], m.Initiatives[idx+1:]...)
	edges := []wmdesign.PortfolioDependency{}
	for _, e := range m.Dependencies {
		if e.From != key && e.To != key {
			edges = append(edges, e)
		}
	}
	m.Dependencies = edges
	return nil
}
func applyPortfolioOperation(m *wmdesign.PortfolioSpec, n *Node, op PortfolioOperation) error {
	switch op.Action + "/" + op.Entity {
	case "set/horizon":
		if op.Horizon == nil || op.Key != op.Horizon.Key {
			return fmt.Errorf("matching complete horizon required")
		}
		for i, v := range m.Horizons {
			if v.Key == op.Key {
				m.Horizons[i] = *op.Horizon
				return nil
			}
		}
		m.Horizons = append(m.Horizons, *op.Horizon)
	case "set/status":
		if op.Status == nil || op.Key != op.Status.Key {
			return fmt.Errorf("matching complete status required")
		}
		for i, v := range m.Statuses {
			if v.Key == op.Key {
				m.Statuses[i] = *op.Status
				return nil
			}
		}
		m.Statuses = append(m.Statuses, *op.Status)
	case "set/initiative":
		if op.Initiative == nil || op.Key != op.Initiative.Key {
			return fmt.Errorf("matching complete initiative required")
		}
		for i, v := range m.Initiatives {
			if v.Key == op.Key {
				m.Initiatives[i] = *op.Initiative
				return nil
			}
		}
		m.Initiatives = append(m.Initiatives, *op.Initiative)
	case "set/dependency":
		if op.Dependency == nil || op.Key != op.Dependency.Key {
			return fmt.Errorf("matching complete dependency required")
		}
		for i, v := range m.Dependencies {
			if v.Key == op.Key {
				m.Dependencies[i] = *op.Dependency
				return nil
			}
		}
		m.Dependencies = append(m.Dependencies, *op.Dependency)
	case "remove/initiative":
		return removePortfolioInitiative(m, op.Key, op.Cascade)
	case "remove/horizon":
		idx := -1
		for i, v := range m.Horizons {
			if v.Key == op.Key {
				idx = i
			}
		}
		if idx < 0 {
			return fmt.Errorf("unknown horizon")
		}
		keys := []string{}
		for _, v := range m.Initiatives {
			if v.Horizon == op.Key {
				keys = append(keys, v.Key)
			}
		}
		if len(keys) > 0 && !op.Cascade {
			return fmt.Errorf("populated horizon requires cascade or prior initiative transfers")
		}
		for _, k := range keys {
			if e := removePortfolioInitiative(m, k, true); e != nil {
				return e
			}
		}
		m.Horizons = append(m.Horizons[:idx], m.Horizons[idx+1:]...)
	case "remove/status":
		for _, v := range m.Initiatives {
			if v.Status == op.Key {
				return fmt.Errorf("used status requires reassignment first")
			}
		}
		for i, v := range m.Statuses {
			if v.Key == op.Key {
				m.Statuses = append(m.Statuses[:i], m.Statuses[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("unknown status")
	case "remove/dependency":
		for i, v := range m.Dependencies {
			if v.Key == op.Key {
				m.Dependencies = append(m.Dependencies[:i], m.Dependencies[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("unknown dependency")
	case "set/layout":
		if op.CardHeight != nil {
			m.CardHeight = *op.CardHeight
		}
		if op.Rect != nil {
			n.Placement.Rect = op.Rect
		}
	case "reorder/horizon", "reorder/status", "reorder/initiative", "reorder/dependency":
		count := 0
		switch op.Entity {
		case "horizon":
			count = len(m.Horizons)
		case "status":
			count = len(m.Statuses)
		case "initiative":
			count = len(m.Initiatives)
		case "dependency":
			count = len(m.Dependencies)
		}
		if len(op.Order) != count {
			return fmt.Errorf("reorder must include all entity keys")
		}
		seen := map[string]bool{}
		horizons := []wmdesign.PortfolioHorizon{}
		statuses := []wmdesign.PortfolioStatus{}
		items := []wmdesign.PortfolioInitiative{}
		deps := []wmdesign.PortfolioDependency{}
		for _, k := range op.Order {
			if seen[k] {
				return fmt.Errorf("duplicate reorder key")
			}
			seen[k] = true
			found := false
			switch op.Entity {
			case "horizon":
				for _, v := range m.Horizons {
					if v.Key == k {
						horizons = append(horizons, v)
						found = true
					}
				}
			case "status":
				for _, v := range m.Statuses {
					if v.Key == k {
						statuses = append(statuses, v)
						found = true
					}
				}
			case "initiative":
				for _, v := range m.Initiatives {
					if v.Key == k {
						items = append(items, v)
						found = true
					}
				}
			case "dependency":
				for _, v := range m.Dependencies {
					if v.Key == k {
						deps = append(deps, v)
						found = true
					}
				}
			}
			if !found {
				return fmt.Errorf("unknown reorder key")
			}
		}
		switch op.Entity {
		case "horizon":
			m.Horizons = horizons
		case "status":
			m.Statuses = statuses
		case "initiative":
			m.Initiatives = items
		case "dependency":
			m.Dependencies = deps
		}
	default:
		return fmt.Errorf("initialize/materialize must be first source operation")
	}
	return nil
}
func PatchPortfolio(p *Project, slideID string, patch PortfolioPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodePortfolioPatch(canonical(patch), "portfolio-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("portfolio source hash mismatch")
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
		return empty, fmt.Errorf("unknown portfolio node")
	}
	n := &(*list)[i]
	if n.Placement == nil || n.Placement.Rect == nil || len(n.Nodes) > 0 {
		return empty, fmt.Errorf("portfolio needs a local rectangular leaf")
	}
	m, bound, e := portfolioSource(n, p.Document.Slides[idx].Values)
	if patch.Operations[0].Action == "initialize" {
		op := patch.Operations[0]
		if op.Entity != "source" || op.Model == nil || !op.Cascade {
			return empty, fmt.Errorf("initialize requires explicit complete model and cascade acknowledgement")
		}
		if n.Definition != nil && n.Definition.ID == "wmds/component/portfolio" {
			return empty, fmt.Errorf("already initialized portfolio")
		}
		m = *op.Model
		bound = false
		e = nil
		n.Definition = &Reference{Scope: "shared", ID: "wmds/component/portfolio"}
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
			return empty, fmt.Errorf("bound portfolio requires materialize/source first")
		}
		if e = applyPortfolioOperation(&m, n, op); e != nil {
			return empty, e
		}
	}
	if e = wmdesign.ValidatePortfolio(m); e != nil {
		return empty, e
	}
	m.Type = ""
	m.X, m.Y, m.W, m.H = 0, 0, 0, 0
	var args map[string]any
	_ = json.Unmarshal(canonical(m), &args)
	if patch.Operations[0].Action != "initialize" {
		if geometry, exists := n.Arguments[wmdesign.SceneSourceGeometryArgument]; exists {
			args[wmdesign.SceneSourceGeometryArgument] = geometry
		}
	}
	n.Arguments = args
	n.Keys = nil
	return CompositionCandidate(p, slideID, "portfolio", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
