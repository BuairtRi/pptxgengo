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

const CyclePatchSchema = "pptxgengo.cycle-patch.v1"

// Cycle identities belong to source; ring order supplies the primary sequence.
// Additional loops and the current marker address keys rather than array indices.
type CycleStep struct {
	Key     string          `json:"key"`
	Label   string          `json:"label"`
	Text    string          `json:"text,omitempty"`
	Number  json.RawMessage `json:"n,omitempty"`
	Surface string          `json:"surface,omitempty"`
}
type CycleLoop struct {
	Key   string   `json:"key"`
	From  string   `json:"from"`
	To    string   `json:"to"`
	Label string   `json:"label,omitempty"`
	Side  *float64 `json:"side,omitempty"`
	Bend  *float64 `json:"bend,omitempty"`
	Ink   string   `json:"ink,omitempty"`
}
type CycleCenter struct {
	Label string   `json:"label,omitempty"`
	Title string   `json:"title,omitempty"`
	Text  string   `json:"text,omitempty"`
	W     *float64 `json:"w,omitempty"`
}
type CycleLayout struct {
	Rect     *wmdesign.Rect `json:"rect,omitempty"`
	NodeW    *float64       `json:"node_width_pt,omitempty"`
	NodeH    *float64       `json:"node_height_pt,omitempty"`
	Start    *float64       `json:"start_degrees,omitempty"`
	Closed   *bool          `json:"closed,omitempty"`
	Ring     *bool          `json:"ring,omitempty"`
	Numbered bool           `json:"numbered"`
	Align    string         `json:"align,omitempty"`
}
type CycleModel struct {
	Steps  []CycleStep  `json:"steps"`
	Loops  []CycleLoop  `json:"loops,omitempty"`
	Active string       `json:"active,omitempty"`
	Center *CycleCenter `json:"center,omitempty"`
	Layout CycleLayout  `json:"layout"`
}
type CycleOperation struct {
	Action  string       `json:"action"`
	Entity  string       `json:"entity"`
	Key     string       `json:"key,omitempty"`
	Order   []string     `json:"order,omitempty"`
	Cascade bool         `json:"cascade,omitempty"`
	Step    *CycleStep   `json:"step,omitempty"`
	Loop    *CycleLoop   `json:"loop,omitempty"`
	Center  *CycleCenter `json:"center,omitempty"`
	Layout  *CycleLayout `json:"layout,omitempty"`
}
type CyclePatch struct {
	Schema               string           `json:"schema"`
	ExpectedSourceSHA256 string           `json:"expected_source_sha256"`
	Actor                string           `json:"actor"`
	Reason               string           `json:"reason"`
	NodeID               string           `json:"node_id"`
	Operations           []CycleOperation `json:"operations"`
}
type CycleInspection struct {
	Schema          string            `json:"schema"`
	SlideID         string            `json:"slide_id"`
	NodeID          string            `json:"node_id"`
	SourceSHA256    string            `json:"source_sha256"`
	Model           CycleModel        `json:"model"`
	MutationBlocked string            `json:"mutation_blocked,omitempty"`
	RenderError     string            `json:"render_error,omitempty"`
	Geometry        DiagramInspection `json:"geometry"`
}

// Kept separate from the renderer's frozen source contract. The composer lowers
// keyed references into that renderer's positional source without changing it.
type cycleSourceStep struct {
	Label   string          `json:"label"`
	Text    string          `json:"text,omitempty"`
	Number  json.RawMessage `json:"n,omitempty"`
	Surface string          `json:"surface,omitempty"`
}
type cycleSourceLoop struct {
	From  *int     `json:"from"`
	To    *int     `json:"to"`
	Label string   `json:"label,omitempty"`
	Side  *float64 `json:"side,omitempty"`
	Bend  *float64 `json:"bend,omitempty"`
	Ink   string   `json:"ink,omitempty"`
}
type cycleSourceSpec struct {
	Type     string            `json:"type,omitempty"`
	ID       string            `json:"id,omitempty"`
	On       string            `json:"on,omitempty"`
	X        float64           `json:"x,omitempty"`
	Y        float64           `json:"y,omitempty"`
	W        float64           `json:"w,omitempty"`
	H        float64           `json:"h,omitempty"`
	CanvasH  float64           `json:"_h,omitempty"`
	Items    []cycleSourceStep `json:"items"`
	Loops    []cycleSourceLoop `json:"loops,omitempty"`
	Center   *CycleCenter      `json:"center,omitempty"`
	Active   *int              `json:"active,omitempty"`
	NodeW    *float64          `json:"nodeW,omitempty"`
	NodeH    *float64          `json:"nodeH,omitempty"`
	Start    *float64          `json:"start,omitempty"`
	Closed   *bool             `json:"closed,omitempty"`
	Ring     *bool             `json:"ring,omitempty"`
	Numbered bool              `json:"numbered,omitempty"`
	Align    string            `json:"align,omitempty"`
	Surface  string            `json:"surface,omitempty"`
}

func cycleFields(action, entity string) (map[string]bool, error) {
	fields := "action entity"
	switch action + "/" + entity {
	case "materialize/source", "remove/active", "remove/center":
	case "set/step":
		fields += " key step"
	case "set/loop":
		fields += " key loop"
	case "remove/step":
		fields += " key cascade"
	case "remove/loop", "set/active":
		fields += " key"
	case "reorder/step", "reorder/loop":
		fields += " order"
	case "set/center":
		fields += " center"
	case "set/layout":
		fields += " layout"
	default:
		return nil, fmt.Errorf("unsupported cycle action/entity %s/%s", action, entity)
	}
	out := map[string]bool{}
	for _, f := range strings.Fields(fields) {
		out[f] = true
	}
	return out, nil
}
func DecodeCyclePatch(raw []byte, file string) (CyclePatch, error) {
	var out CyclePatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("cycle patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty cycle patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("cycle patch requires one document")
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
	for i, op := range out.Operations {
		allowed, e := cycleFields(op.Action, op.Entity)
		if e != nil {
			return out, e
		}
		authored, _ := ops[i].(map[string]any)
		for k := range authored {
			if !allowed[k] {
				return out, fmt.Errorf("cycle %s/%s does not accept authored field %s", op.Action, op.Entity, k)
			}
		}
	}
	_, e = hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != CyclePatchSchema || len(out.ExpectedSourceSHA256) != 64 || e != nil || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || strings.TrimSpace(out.Reason) == "" || len(out.Reason) > 4096 || !stableID.MatchString(out.NodeID) || len(out.Operations) < 1 || len(out.Operations) > 500 {
		return out, fmt.Errorf("cycle patch requires schema, source SHA256, actor, reason, node_id and 1..500 operations")
	}
	return out, nil
}
func cycleNode(t *LocalTemplate, id string) (*Node, error) {
	list, i := findDiagramNode(&t.Nodes, id)
	if list == nil {
		return nil, fmt.Errorf("unknown cycle node %s", id)
	}
	n := &(*list)[i]
	if n.Definition == nil || n.Definition.ID != "wmds/component/cycle" {
		return nil, fmt.Errorf("node %s is not a semantic cycle component; detach a typed cycle template first", id)
	}
	return n, nil
}
func cycleKeys(old map[string][]string, path string, count int) ([]string, error) {
	keys, ok := old[path]
	slash, other := old["/"+path]
	if ok && other && !bytes.Equal(canonical(keys), canonical(slash)) {
		return nil, fmt.Errorf("conflicting cycle array keys %s", path)
	}
	if !ok && other {
		keys, ok = slash, true
	}
	if !ok {
		if len(old) > 0 {
			return nil, fmt.Errorf("missing cycle array keys %s", path)
		}
		keys = make([]string, count)
		for i := range keys {
			keys[i] = fmt.Sprintf("source-%03d", i+1)
		}
	}
	if len(keys) != count {
		return nil, fmt.Errorf("cycle array key count mismatch %s", path)
	}
	seen := map[string]bool{}
	for _, key := range keys {
		if !stableID.MatchString(key) || seen[key] {
			return nil, fmt.Errorf("invalid or duplicate cycle key %s", key)
		}
		seen[key] = true
	}
	return keys, nil
}
func cycleSource(n *Node, values map[string]any) (CycleModel, cycleSourceSpec, bool, error) {
	var m CycleModel
	var s cycleSourceSpec
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return m, s, bound, e
	}
	if e = strictInto(args, &s); e != nil {
		return m, s, bound, e
	}
	for path, keys := range n.Keys {
		if path != "items" && path != "/items" && path != "loops" && path != "/loops" {
			return m, s, bound, fmt.Errorf("unsupported cycle array key path %s", path)
		}
		if (path == "loops" || path == "/loops") && len(s.Loops) == 0 && len(keys) > 0 {
			return m, s, bound, fmt.Errorf("cycle loop keys supplied without loops")
		}
	}
	itemKeys, e := cycleKeys(n.Keys, "items", len(s.Items))
	if e != nil {
		return m, s, bound, e
	}
	for i, it := range s.Items {
		m.Steps = append(m.Steps, CycleStep{Key: itemKeys[i], Label: it.Label, Text: it.Text, Number: it.Number, Surface: it.Surface})
	}
	if len(s.Loops) > 0 {
		keys, e := cycleKeys(n.Keys, "loops", len(s.Loops))
		if e != nil {
			return m, s, bound, e
		}
		for i, lp := range s.Loops {
			if lp.From == nil || lp.To == nil || *lp.From < 0 || *lp.From >= len(itemKeys) || *lp.To < 0 || *lp.To >= len(itemKeys) {
				return m, s, bound, fmt.Errorf("cycle loop %s has invalid positional endpoints", keys[i])
			}
			m.Loops = append(m.Loops, CycleLoop{Key: keys[i], From: itemKeys[*lp.From], To: itemKeys[*lp.To], Label: lp.Label, Side: lp.Side, Bend: lp.Bend, Ink: lp.Ink})
		}
	}
	if s.Active != nil {
		if *s.Active < 0 || *s.Active >= len(itemKeys) {
			return m, s, bound, fmt.Errorf("cycle active index is invalid")
		}
		m.Active = itemKeys[*s.Active]
	}
	m.Center = s.Center
	m.Layout = CycleLayout{NodeW: s.NodeW, NodeH: s.NodeH, Start: s.Start, Closed: s.Closed, Ring: s.Ring, Numbered: s.Numbered, Align: s.Align}
	if n.Placement != nil {
		m.Layout.Rect = n.Placement.Rect
	}
	return m, s, bound, validateCycle(m)
}
func cycleStepIndex(m CycleModel, key string) int {
	for i, v := range m.Steps {
		if v.Key == key {
			return i
		}
	}
	return -1
}
func cycleLoopIndex(m CycleModel, key string) int {
	for i, v := range m.Loops {
		if v.Key == key {
			return i
		}
	}
	return -1
}
func validateCycle(m CycleModel) error {
	if len(m.Steps) < 2 || len(m.Steps) > 12 || len(m.Loops) > 32 {
		return fmt.Errorf("cycle requires 2..12 steps and at most 32 feedback loops; actual fit remains measured")
	}
	seen := map[string]bool{}
	for _, s := range m.Steps {
		if !stableID.MatchString(s.Key) || seen[s.Key] || strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("invalid/duplicate cycle step key or empty label %s", s.Key)
		}
		seen[s.Key] = true
	}
	if m.Active != "" && !seen[m.Active] {
		return fmt.Errorf("unknown active cycle step %s", m.Active)
	}
	loops := map[string]bool{}
	for _, l := range m.Loops {
		if !stableID.MatchString(l.Key) || loops[l.Key] {
			return fmt.Errorf("invalid/duplicate cycle loop key %s", l.Key)
		}
		loops[l.Key] = true
		if !seen[l.From] || !seen[l.To] {
			return fmt.Errorf("cycle loop %s has unknown step endpoint", l.Key)
		}
		if l.From == l.To {
			return fmt.Errorf("cycle loop %s cannot connect a step to itself", l.Key)
		}
	}
	return nil
}
func applyCycleOperation(m *CycleModel, op CycleOperation) error {
	allowed, e := cycleFields(op.Action, op.Entity)
	if e != nil {
		return e
	}
	var fields map[string]any
	_ = json.Unmarshal(canonical(op), &fields)
	for k := range fields {
		if !allowed[k] {
			return fmt.Errorf("unexpected cycle operation field %s", k)
		}
	}
	switch op.Action + "/" + op.Entity {
	case "set/step":
		if op.Step == nil || op.Key != op.Step.Key || !stableID.MatchString(op.Key) {
			return fmt.Errorf("set step requires matching stable key and complete step")
		}
		i := cycleStepIndex(*m, op.Key)
		if i < 0 {
			m.Steps = append(m.Steps, *op.Step)
		} else {
			m.Steps[i] = *op.Step
		}
	case "set/loop":
		if op.Loop == nil || op.Key != op.Loop.Key || !stableID.MatchString(op.Key) {
			return fmt.Errorf("set loop requires matching stable key and complete loop")
		}
		i := cycleLoopIndex(*m, op.Key)
		if i < 0 {
			m.Loops = append(m.Loops, *op.Loop)
		} else {
			m.Loops[i] = *op.Loop
		}
	case "remove/step":
		i := cycleStepIndex(*m, op.Key)
		if i < 0 {
			return fmt.Errorf("unknown cycle step %s", op.Key)
		}
		dependent := m.Active == op.Key
		for _, l := range m.Loops {
			dependent = dependent || l.From == op.Key || l.To == op.Key
		}
		if dependent && !op.Cascade {
			return fmt.Errorf("step removal affects active marker or feedback loops; remove/reassign them first or explicitly acknowledge cascade")
		}
		if m.Active == op.Key {
			m.Active = ""
		}
		kept := []CycleLoop{}
		for _, l := range m.Loops {
			if l.From != op.Key && l.To != op.Key {
				kept = append(kept, l)
			}
		}
		m.Loops = kept
		m.Steps = append(m.Steps[:i], m.Steps[i+1:]...)
	case "remove/loop":
		i := cycleLoopIndex(*m, op.Key)
		if i < 0 {
			return fmt.Errorf("unknown cycle loop %s", op.Key)
		}
		m.Loops = append(m.Loops[:i], m.Loops[i+1:]...)
	case "set/active":
		if cycleStepIndex(*m, op.Key) < 0 {
			return fmt.Errorf("unknown active cycle step %s", op.Key)
		}
		m.Active = op.Key
	case "remove/active":
		m.Active = ""
	case "set/center":
		if op.Center == nil {
			return fmt.Errorf("set center requires center")
		}
		m.Center = op.Center
	case "remove/center":
		m.Center = nil
	case "set/layout":
		if op.Layout == nil {
			return fmt.Errorf("set layout requires complete layout")
		}
		rect := m.Layout.Rect
		m.Layout = *op.Layout
		if m.Layout.Rect == nil {
			m.Layout.Rect = rect
		}
	case "reorder/step", "reorder/loop":
		count := len(m.Steps)
		if op.Entity == "loop" {
			count = len(m.Loops)
		}
		if len(op.Order) != count {
			return fmt.Errorf("cycle reorder must enumerate every %s key exactly once", op.Entity)
		}
		seen := map[string]bool{}
		steps := []CycleStep{}
		loops := []CycleLoop{}
		for _, k := range op.Order {
			if seen[k] {
				return fmt.Errorf("duplicate cycle reorder key %s", k)
			}
			seen[k] = true
			if op.Entity == "step" {
				i := cycleStepIndex(*m, k)
				if i < 0 {
					return fmt.Errorf("unknown cycle step %s", k)
				}
				steps = append(steps, m.Steps[i])
			} else {
				i := cycleLoopIndex(*m, k)
				if i < 0 {
					return fmt.Errorf("unknown cycle loop %s", k)
				}
				loops = append(loops, m.Loops[i])
			}
		}
		if op.Entity == "step" {
			m.Steps = steps
		} else {
			m.Loops = loops
		}
	default:
		return fmt.Errorf("materialize is only accepted as the first source operation")
	}
	return nil
}
func lowerCycle(n *Node, m CycleModel, s cycleSourceSpec) error {
	if e := validateCycle(m); e != nil {
		return e
	}
	s.Items = nil
	s.Loops = nil
	s.Active = nil
	s.Center = m.Center
	s.NodeW, s.NodeH, s.Start, s.Closed, s.Ring, s.Numbered, s.Align = m.Layout.NodeW, m.Layout.NodeH, m.Layout.Start, m.Layout.Closed, m.Layout.Ring, m.Layout.Numbered, m.Layout.Align
	keys := map[string][]string{}
	for _, it := range m.Steps {
		s.Items = append(s.Items, cycleSourceStep{it.Label, it.Text, it.Number, it.Surface})
		keys["items"] = append(keys["items"], it.Key)
	}
	for _, lp := range m.Loops {
		from, to := cycleStepIndex(m, lp.From), cycleStepIndex(m, lp.To)
		s.Loops = append(s.Loops, cycleSourceLoop{&from, &to, lp.Label, lp.Side, lp.Bend, lp.Ink})
		keys["loops"] = append(keys["loops"], lp.Key)
	}
	if m.Active != "" {
		i := cycleStepIndex(m, m.Active)
		s.Active = &i
	}
	var args map[string]any
	if e := json.Unmarshal(canonical(s), &args); e != nil {
		return e
	}
	for _, reserved := range []string{"type", "id", "x", "y", "w", "h"} {
		delete(args, reserved)
	}
	n.Arguments, n.Keys = args, keys
	if m.Layout.Rect != nil {
		if n.Placement == nil {
			return fmt.Errorf("cycle layout requires an existing placement")
		}
		n.Placement.Rect = m.Layout.Rect
	}
	return nil
}
func InspectCycle(p *Project, slideID, nodeID, bundle, engine string) (CycleInspection, error) {
	out := CycleInspection{Schema: "pptxgengo.cycle-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	n, e := cycleNode(&t, nodeID)
	if e != nil {
		return out, e
	}
	var bound bool
	out.Model, _, bound, e = cycleSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Arguments contain bindings; edit bound source values or explicitly begin with materialize/source to retain resolved values as local constants."
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
		out.Geometry.Validation = "renderer_failed_no_measurement_qualification"
	}
	return out, nil
}
func PatchCycle(p *Project, slideID string, patch CyclePatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeCyclePatch(canonical(patch), "cycle-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("cycle source hash mismatch; inspect again")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := cycleNode(&clone, patch.NodeID)
	if e != nil {
		return empty, e
	}
	m, s, bound, e := cycleSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	for i, op := range patch.Operations {
		if op.Action == "materialize" && op.Entity == "source" {
			if i != 0 {
				return empty, fmt.Errorf("materialize source must be first")
			}
			bound = false
			continue
		}
		if bound {
			return empty, fmt.Errorf("cycle patch refuses bound arguments without explicit first materialize source operation")
		}
		if e = applyCycleOperation(&m, op); e != nil {
			return empty, e
		}
	}
	if e = lowerCycle(n, m, s); e != nil {
		return empty, e
	}
	return CompositionCandidate(p, slideID, "cycle", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
