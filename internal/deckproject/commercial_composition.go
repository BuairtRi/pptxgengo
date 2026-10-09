package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

const CommercialPatchSchema = "pptxgengo.commercial-patch.v1"

type CommercialOperation struct {
	Action      string                      `json:"action"`
	Entity      string                      `json:"entity"`
	Key         string                      `json:"key,omitempty"`
	Path        string                      `json:"path,omitempty"`
	Cascade     bool                        `json:"cascade,omitempty"`
	Order       []string                    `json:"order,omitempty"`
	Row         *wmdesign.CommercialRow     `json:"row,omitempty"`
	Nodes       []string                    `json:"nodes,omitempty"`
	Model       *wmdesign.CommercialModel   `json:"model,omitempty"`
	Targets     []wmdesign.CommercialTarget `json:"targets,omitempty"`
	Assumptions []string                    `json:"assumptions,omitempty"`
	Precision   *int                        `json:"precision,omitempty"`
	Rounding    string                      `json:"rounding,omitempty"`
	Value       any                         `json:"value,omitempty"`
}
type CommercialPatch struct {
	Schema               string                `json:"schema"`
	ExpectedSourceSHA256 string                `json:"expected_source_sha256"`
	Actor                string                `json:"actor"`
	Reason               string                `json:"reason"`
	NodeID               string                `json:"node_id"`
	Operations           []CommercialOperation `json:"operations"`
}
type CommercialInspection struct {
	Schema           string                      `json:"schema"`
	SlideID          string                      `json:"slide_id"`
	NodeID           string                      `json:"node_id"`
	SourceSHA256     string                      `json:"source_sha256"`
	Model            wmdesign.CommercialModel    `json:"model"`
	Results          []wmdesign.CommercialResult `json:"results"`
	Presentation     map[string]any              `json:"presentation"`
	Group            []string                    `json:"group,omitempty"`
	PresentationKeys map[string][]string         `json:"presentation_keys"`
	Contract         string                      `json:"contract"`
	MutationBlocked  string                      `json:"mutation_blocked,omitempty"`
	RenderError      string                      `json:"render_error,omitempty"`
	Geometry         DiagramInspection           `json:"geometry"`
}

func commercialFields(a, e string) (map[string]bool, error) {
	f := "action entity"
	switch a + "/" + e {
	case "initialize/source":
		f += " model cascade nodes"
	case "materialize/source":
	case "set/row":
		f += " key row"
	case "remove/row":
		f += " key cascade"
	case "reorder/row":
		f += " order"
	case "set/targets":
		f += " targets"
	case "set/assumptions":
		f += " assumptions"
	case "set/rounding":
		f += " precision rounding"
	case "set/presentation":
		f += " path value"
	case "set/collection":
		f += " path key value"
	case "remove/collection":
		f += " path key cascade"
	case "reorder/collection":
		f += " path order"
	default:
		return nil, fmt.Errorf("unsupported commercial operation %s/%s", a, e)
	}
	out := map[string]bool{}
	for _, s := range strings.Fields(f) {
		out[s] = true
	}
	return out, nil
}
func DecodeCommercialPatch(raw []byte, file string) (CommercialPatch, error) {
	var out CommercialPatch
	e := decodeDomainPatch(raw, file, &out, commercialFields)
	if e == nil {
		e = validDomainEnvelope(out.Schema, CommercialPatchSchema, out.ExpectedSourceSHA256, out.Actor, out.Reason, out.NodeID, len(out.Operations))
	}
	return out, e
}
func commercialNode(t *LocalTemplate, id string) (*Node, error) {
	list, i := findDiagramNode(&t.Nodes, id)
	if list == nil {
		return nil, fmt.Errorf("unknown commercial node %s", id)
	}
	n := &(*list)[i]
	if len(n.Nodes) > 0 {
		return nil, fmt.Errorf("select an individual commercial component, not a layout group")
	}
	return n, nil
}
func commercialSource(n *Node, values map[string]any) (wmdesign.CommercialSpec, bool, error) {
	var out wmdesign.CommercialSpec
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	if n.Definition == nil || n.Definition.ID != "wmds/component/commercial" {
		return out, bound, fmt.Errorf("selected node has no explicit calculation model; inspect geometry then initialize/source with complete facts and mappings")
	}
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return out, bound, e
	}
	e = strictInto(args, &out)
	if e != nil {
		return out, bound, e
	}
	_, _, e = wmdesign.MaterializeCommercialPresentation(out)
	return out, bound, e
}
func InspectCommercial(p *Project, slide, node, bundle, engine string) (CommercialInspection, error) {
	out := CommercialInspection{Schema: "pptxgengo.commercial-inspection.v1", SlideID: slide, NodeID: node, SourceSHA256: p.SourceHash(), Contract: "Decimal strings, rational formulas, explicit input provenance and dimensional assumptions. Round only final outputs; numerical targets are explicit, never derived from native geometry."}
	idx, t, e := diagramSlide(p, slide)
	if e != nil {
		return out, e
	}
	n, e := commercialNode(&t, node)
	if e != nil {
		return out, e
	}
	s, bound, e := commercialSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	out.Group = s.Group
	out.Model = s.Model
	out.Presentation, out.Results, e = wmdesign.MaterializeCommercialPresentation(s)
	if e != nil {
		return out, e
	}
	out.PresentationKeys = map[string][]string{}
	for path, keys := range n.Keys {
		out.PresentationKeys[strings.TrimPrefix(strings.TrimPrefix(path, "/"), "presentation/")] = keys
	}
	if bound {
		out.MutationBlocked = "Begin with explicit materialize/source or edit original bindings."
	}
	out.Geometry, e = InspectDiagram(p, slide, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func initializeCommercial(n *Node, values map[string]any, model *wmdesign.CommercialModel, cascade bool, group []string) (wmdesign.CommercialSpec, error) {
	var s wmdesign.CommercialSpec
	if !cascade || model == nil {
		return s, fmt.Errorf("initialize requires complete model and cascade:true acknowledging replacing hardcoded source amounts with explicit mappings")
	}
	if n.Definition != nil && n.Definition.ID == "wmds/component/commercial" {
		return s, fmt.Errorf("commercial model already initialized")
	}
	if n.Placement == nil {
		return s, fmt.Errorf("commercial source requires explicit placement")
	}
	kind := ""
	var source map[string]any
	if n.Kind == "text" {
		resolved, e := resolveArguments(map[string]any{"text": n.Text}, values)
		if e != nil {
			return s, e
		}
		kind = "text"
		source = map[string]any{"text": resolved["text"], "style": n.Style}
		if n.Ink != "" {
			source["ink"] = n.Ink
		}
		if n.Align != "" {
			source["align"] = n.Align
		}
	} else {
		if n.Kind != "component" || n.Definition == nil {
			return s, fmt.Errorf("initialize requires text or source component")
		}
		kind = strings.TrimPrefix(strings.TrimPrefix(n.Definition.ID, "wmds/component/"), "wmds/composite/")
		switch n.Definition.ID {
		case "card":
			kind = "card"
		case "data.metric":
			kind = "metric"
		case "text.block":
			kind = "textblock"
		}
		if kind == n.Definition.ID && kind != "card" {
			return s, fmt.Errorf("unsupported commercial source definition")
		}
		args, e := resolveArguments(n.Arguments, values)
		if e != nil {
			return s, e
		}
		if e = json.Unmarshal(canonical(args), &source); e != nil {
			return s, e
		}
	}
	source["type"] = kind
	s.NodeID = n.ID
	s.Group = append([]string(nil), group...)
	s.Model = *model
	s.Presentation = source
	if _, _, e := wmdesign.MaterializeCommercialPresentation(s); e != nil {
		return s, e
	}
	newKeys := map[string][]string{}
	for path, keys := range n.Keys {
		newKeys["presentation/"+strings.TrimPrefix(path, "/")] = keys
	}
	n.Keys = newKeys
	n.Kind = "component"
	n.Definition = &Reference{Scope: "shared", ID: "wmds/component/commercial"}
	n.Text = nil
	n.Style = ""
	n.Ink = ""
	n.Align = ""
	return s, nil
}
func commercialRowIndex(m wmdesign.CommercialModel, key string) int {
	for i, r := range m.Rows {
		if r.Key == key {
			return i
		}
	}
	return -1
}
func applyCommercial(s *wmdesign.CommercialSpec, keys map[string][]string, op CommercialOperation) error {
	switch op.Entity {
	case "row":
		i := commercialRowIndex(s.Model, op.Key)
		switch op.Action {
		case "set":
			if op.Row == nil || op.Row.Key != op.Key {
				return fmt.Errorf("row operation requires matching keyed row")
			}
			if i < 0 {
				s.Model.Rows = append(s.Model.Rows, *op.Row)
			} else {
				s.Model.Rows[i] = *op.Row
			}
			return nil
		case "remove":
			if i < 0 {
				return fmt.Errorf("unknown commercial row")
			}
			if !op.Cascade {
				return fmt.Errorf("row removal requires cascade:true; dependent formulas/targets must be explicitly updated before removal")
			}
			for _, r := range s.Model.Rows {
				if r.Formula != nil && teamFind(r.Formula.Inputs, op.Key) >= 0 {
					return fmt.Errorf("row %s still referenced by formula %s", op.Key, r.Key)
				}
			}
			for _, target := range s.Model.Targets {
				if target.Row == op.Key {
					return fmt.Errorf("row still mapped by presentation target")
				}
			}
			s.Model.Rows = append(s.Model.Rows[:i], s.Model.Rows[i+1:]...)
			return nil
		case "reorder":
			if len(op.Order) != len(s.Model.Rows) {
				return fmt.Errorf("row order must list every key")
			}
			seen := map[string]bool{}
			out := []wmdesign.CommercialRow{}
			for _, key := range op.Order {
				j := commercialRowIndex(s.Model, key)
				if j < 0 || seen[key] {
					return fmt.Errorf("invalid row order")
				}
				seen[key] = true
				out = append(out, s.Model.Rows[j])
			}
			s.Model.Rows = out
			return nil
		}
	case "targets":
		s.Model.Targets = op.Targets
		return nil
	case "assumptions":
		s.Model.Assumptions = op.Assumptions
		return nil
	case "rounding":
		if op.Precision == nil || op.Rounding == "" {
			return fmt.Errorf("rounding requires precision and mode")
		}
		s.Model.Precision = *op.Precision
		s.Model.Rounding = op.Rounding
		return nil
	case "presentation":
		if !strings.HasPrefix(op.Path, "/") {
			return fmt.Errorf("presentation requires exact pointer")
		}
		if commercialGeometryPointer(op.Path) {
			return fmt.Errorf("edit source geometry through diagram commands, not numerical source patch")
		}
		if op.Value == nil {
			return fmt.Errorf("presentation set requires value; use explicit field-specific missing handling")
		}
		return setPointer(s.Presentation, op.Path, op.Value)
	case "collection":
		return commercialCollection(s, keys, op)
	}
	return fmt.Errorf("unsupported commercial operation")
}
func commercialGeometryPointer(path string) bool {
	for _, part := range strings.Split(path, "/") {
		switch part {
		case "type", "x", "y", "w", "h", "points", "_source_geometry":
			return true
		}
	}
	return false
}
func commercialCollection(s *wmdesign.CommercialSpec, keys map[string][]string, op CommercialOperation) error {
	if !strings.HasPrefix(op.Path, "/") || commercialGeometryPointer(op.Path) {
		return fmt.Errorf("collection requires non-geometric exact array pointer")
	}
	v, e := lookupPointer(s.Presentation, op.Path)
	if e != nil {
		return e
	}
	items, ok := v.([]any)
	if !ok {
		return fmt.Errorf("collection path must address array")
	}
	path := "presentation" + op.Path
	ids, e := cycleKeys(keys, path, len(items))
	if e != nil {
		return e
	}
	i := teamFind(ids, op.Key)
	var out []any
	var newIDs []string
	switch op.Action {
	case "set":
		if !stableID.MatchString(op.Key) || op.Value == nil {
			return fmt.Errorf("collection requires stable key and item value")
		}
		out, newIDs = items, ids
		if i < 0 {
			out = append(out, op.Value)
			newIDs = append(newIDs, op.Key)
		} else {
			out[i] = op.Value
		}
	case "remove":
		if !op.Cascade || i < 0 {
			return fmt.Errorf("collection removal requires existing key and cascade:true")
		}
		out, newIDs = teamRemove(items, append([]string(nil), ids...), i)
	case "reorder":
		out, newIDs, e = teamOrder(items, ids, op.Order)
		if e != nil {
			return e
		}
	}
	// Targets are tied to the original keyed item; reordering keeps that identity.
	for j := range s.Model.Targets {
		target := &s.Model.Targets[j]
		prefix := op.Path + "/"
		if target.NodeID != "" && target.NodeID != s.NodeID {
			continue
		}
		if !strings.HasPrefix(target.Path, prefix) {
			continue
		}
		tail := strings.TrimPrefix(target.Path, prefix)
		parts := strings.SplitN(tail, "/", 2)
		old := teamFindIndex(parts[0], len(ids))
		if old < 0 {
			return fmt.Errorf("mapped target has invalid index")
		}
		next := teamFind(newIDs, ids[old])
		if next < 0 {
			return fmt.Errorf("remove mapped targets explicitly before removing collection key %s", ids[old])
		}
		suffix := ""
		if len(parts) == 2 {
			suffix = "/" + parts[1]
		}
		target.Path = fmt.Sprintf("%s/%d%s", op.Path, next, suffix)
	}
	if e = setPointer(s.Presentation, op.Path, out); e != nil {
		return e
	}
	remapQuantitativeNestedKeys(keys, path, ids, newIDs)
	keys[path] = newIDs
	return nil
}
func teamFindIndex(text string, count int) int {
	n := -1
	if _, e := fmt.Sscanf(text, "%d", &n); e != nil || fmt.Sprint(n) != text || n < 0 || n >= count {
		return -1
	}
	return n
}
func PatchCommercial(p *Project, slide string, patch CommercialPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeCommercialPatch(canonical(patch), "commercial-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if p.SourceHash() != patch.ExpectedSourceSHA256 {
		return empty, fmt.Errorf("commercial source hash mismatch; inspect again")
	}
	idx, t, e := diagramSlide(p, slide)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	selected, e := commercialNode(&clone, patch.NodeID)
	if e != nil {
		return empty, e
	}
	values := p.Document.Slides[idx].Values
	states := map[string]wmdesign.CommercialSpec{}
	bounds := map[string]bool{}
	group := []string{}
	if patch.Operations[0].Action == "initialize" {
		op := patch.Operations[0]
		if op.Model == nil {
			return empty, fmt.Errorf("initialize requires model")
		}
		group = append([]string{patch.NodeID}, op.Nodes...)
		seen := map[string]bool{}
		for _, id := range group {
			if !stableID.MatchString(id) || seen[id] {
				return empty, fmt.Errorf("initialize requires unique stable source nodes")
			}
			seen[id] = true
		}
		model := *op.Model
		model.Targets = append([]wmdesign.CommercialTarget(nil), model.Targets...)
		for i := range model.Targets {
			if model.Targets[i].NodeID == "" {
				model.Targets[i].NodeID = patch.NodeID
			}
			if !seen[model.Targets[i].NodeID] {
				return empty, fmt.Errorf("target must name a registered calculation node")
			}
		}
		for _, id := range group {
			n, e := commercialNode(&clone, id)
			if e != nil {
				return empty, e
			}
			s, e := initializeCommercial(n, values, &model, op.Cascade, group)
			if e != nil {
				return empty, e
			}
			states[id] = s
		}
	} else {
		original, bound, e := commercialSource(selected, values)
		if e != nil {
			return empty, e
		}
		group = original.Group
		if len(group) == 0 {
			group = []string{patch.NodeID}
		}
		for _, id := range group {
			n, e := commercialNode(&clone, id)
			if e != nil {
				return empty, e
			}
			s, b, e := commercialSource(n, values)
			if e != nil {
				return empty, e
			}
			if !bytes.Equal(canonical(s.Model), canonical(original.Model)) || !bytes.Equal(canonical(s.Group), canonical(original.Group)) {
				return empty, fmt.Errorf("calculation group model drift; reconcile authored models before patch")
			}
			states[id] = s
			bounds[id] = b
		}
		bounds[patch.NodeID] = bound
	}
	state := states[patch.NodeID]
	if selected.Keys == nil {
		selected.Keys = map[string][]string{}
	}
	for i, op := range patch.Operations {
		if op.Action == "initialize" || op.Action == "materialize" {
			if i != 0 {
				return empty, fmt.Errorf("source migration/materialization must be first")
			}
			for id := range bounds {
				bounds[id] = false
			}
			continue
		}
		for _, b := range bounds {
			if b {
				return empty, fmt.Errorf("commercial refuses bound calculation group without explicit materialize/source")
			}
		}
		if e = applyCommercial(&state, selected.Keys, op); e != nil {
			return empty, e
		}
	}
	states[patch.NodeID] = state
	for _, id := range group {
		n, e := commercialNode(&clone, id)
		if e != nil {
			return empty, e
		}
		s := states[id]
		s.Model = state.Model
		commercialModelKeys(n, s.Model)
		if _, _, e = wmdesign.MaterializeCommercialPresentation(s); e != nil {
			return empty, e
		}
		n.Arguments = nil
		if e = json.Unmarshal(canonical(s), &n.Arguments); e != nil {
			return empty, e
		}
		if e = componentPopulateKeys(n, n.Arguments); e != nil {
			return empty, e
		}
	}
	return CompositionCandidate(p, slide, "commercial", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}

// The model rows own their stable identity independently from positional display arrays.
func commercialModelKeys(n *Node, model wmdesign.CommercialModel) {
	if n.Keys == nil {
		n.Keys = map[string][]string{}
	}
	ids := make([]string, len(model.Rows))
	for i, row := range model.Rows {
		ids[i] = row.Key
	}
	old := n.Keys["model/rows"]
	if old == nil {
		old = n.Keys["/model/rows"]
	}
	componentRebaseKeys(n, "model/rows", old, ids)
	// Inputs and whole-list settings are declared replacements, not positional adoption.
	for path := range n.Keys {
		clean := strings.TrimPrefix(path, "/")
		if strings.HasPrefix(clean, "model/targets") || strings.HasPrefix(clean, "model/assumptions") || strings.HasSuffix(clean, "/formula/inputs") {
			delete(n.Keys, path)
		}
	}
}
