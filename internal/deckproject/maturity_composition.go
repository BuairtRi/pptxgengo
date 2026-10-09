package deckproject

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"math"
	"strings"
)

const MaturityPatchSchema = "pptxgengo.maturity-patch.v1"

type MaturityStage struct {
	Key    string          `json:"key"`
	Label  string          `json:"label"`
	Text   string          `json:"text,omitempty"`
	Number json.RawMessage `json:"n,omitempty"`
	At     float64         `json:"at"`
}
type MaturityMarker struct {
	Stage string `json:"stage"`
	Label string `json:"label,omitempty"`
}
type MaturityBranch struct {
	From   string          `json:"from"`
	Label  string          `json:"label"`
	Text   string          `json:"text,omitempty"`
	Number json.RawMessage `json:"n,omitempty"`
}
type MaturityLayout struct {
	Rect       *wmdesign.Rect `json:"rect,omitempty"`
	Shape      *float64       `json:"shape,omitempty"`
	Headroom   *float64       `json:"headroom_pt,omitempty"`
	LabelWidth *float64       `json:"label_width_pt,omitempty"`
	Axis       *bool          `json:"axis,omitempty"`
	AxisLabel  *string        `json:"axis_label,omitempty"`
}
type MaturityModel struct {
	Stages     []MaturityStage `json:"stages"`
	Current    string          `json:"current,omitempty"`
	Here       *MaturityMarker `json:"here,omitempty"`
	Target     *MaturityMarker `json:"target,omitempty"`
	Inflection *MaturityMarker `json:"inflection,omitempty"`
	Branch     *MaturityBranch `json:"branch,omitempty"`
	Layout     MaturityLayout  `json:"layout"`
}
type MaturityOperation struct {
	Action  string          `json:"action"`
	Entity  string          `json:"entity"`
	Key     string          `json:"key,omitempty"`
	Stage   *MaturityStage  `json:"stage,omitempty"`
	Marker  *MaturityMarker `json:"marker,omitempty"`
	Branch  *MaturityBranch `json:"branch,omitempty"`
	Layout  *MaturityLayout `json:"layout,omitempty"`
	Order   []string        `json:"order,omitempty"`
	Cascade bool            `json:"cascade,omitempty"`
	Spacing string          `json:"spacing,omitempty"`
}
type MaturityPatch struct {
	Schema               string              `json:"schema"`
	ExpectedSourceSHA256 string              `json:"expected_source_sha256"`
	Actor                string              `json:"actor"`
	Reason               string              `json:"reason"`
	NodeID               string              `json:"node_id"`
	Operations           []MaturityOperation `json:"operations"`
}
type MaturityInspection struct {
	Schema          string            `json:"schema"`
	SlideID         string            `json:"slide_id"`
	NodeID          string            `json:"node_id"`
	SourceSHA256    string            `json:"source_sha256"`
	Model           MaturityModel     `json:"model"`
	MutationBlocked string            `json:"mutation_blocked,omitempty"`
	RenderError     string            `json:"render_error,omitempty"`
	Geometry        DiagramInspection `json:"geometry"`
}
type maturitySourceStage struct {
	Label  string          `json:"label"`
	Text   string          `json:"text,omitempty"`
	Number json.RawMessage `json:"n,omitempty"`
}
type maturitySourceMarker struct {
	Stage *int   `json:"stage,omitempty"`
	Label string `json:"label,omitempty"`
}
type maturitySourceBranch struct {
	From   *int            `json:"from"`
	Label  string          `json:"label"`
	Text   string          `json:"text,omitempty"`
	Number json.RawMessage `json:"n,omitempty"`
}
type maturitySourceSpec struct {
	SourceGeometry  json.RawMessage       `json:"_source_geometry,omitempty"`
	Stages          []maturitySourceStage `json:"stages"`
	At              []float64             `json:"at,omitempty"`
	Shape           *float64              `json:"shape,omitempty"`
	Headroom        *float64              `json:"headroom,omitempty"`
	LabelW          *float64              `json:"labelW,omitempty"`
	Axis            *bool                 `json:"axis,omitempty"`
	AxisLabel       string                `json:"axisLabel,omitempty"`
	Active          *int                  `json:"active,omitempty"`
	Here            *maturitySourceMarker `json:"here,omitempty"`
	Target          *maturitySourceMarker `json:"target,omitempty"`
	Inflection      *int                  `json:"inflection,omitempty"`
	InflectionLabel string                `json:"inflectionLabel,omitempty"`
	Branch          *maturitySourceBranch `json:"branch,omitempty"`
	CanvasH         *float64              `json:"_h,omitempty"`
}

func maturityFields(a, b string) (string, error) {
	switch a + "/" + b {
	case "materialize/source", "remove/current", "remove/here", "remove/target", "remove/inflection", "remove/branch":
		return "", nil
	case "set/stage":
		return "key stage", nil
	case "remove/stage":
		return "key cascade", nil
	case "reorder/stage":
		return "order spacing", nil
	case "set/current":
		return "key", nil
	case "set/here", "set/target", "set/inflection":
		return "marker", nil
	case "set/branch":
		return "branch", nil
	case "set/layout":
		return "layout", nil
	case "set/spacing":
		return "spacing", nil
	}
	return "", fmt.Errorf("unsupported maturity operation %s/%s", a, b)
}
func DecodeMaturityPatch(raw []byte, file string) (MaturityPatch, error) {
	var out MaturityPatch
	e := decodeCurveDocument(raw, file, &out, MaturityPatchSchema, maturityFields)
	if e == nil && !stableID.MatchString(out.NodeID) {
		e = fmt.Errorf("maturity requires stable node_id")
	}
	return out, e
}
func maturityRead(n *Node, values map[string]any) (MaturityModel, maturitySourceSpec, bool, error) {
	var m MaturityModel
	var s maturitySourceSpec
	args, b, e := curveSource(n, values)
	if e != nil {
		return m, s, b, e
	}
	if e = strictInto(args, &s); e != nil {
		return m, s, b, e
	}
	keys, e := curveKeys(n, "stages", len(s.Stages))
	if e != nil {
		return m, s, b, e
	}
	if len(s.At) != 0 && len(s.At) != len(keys) {
		return m, s, b, fmt.Errorf("maturity stage spacing count mismatch")
	}
	for i, st := range s.Stages {
		at := .5
		if len(s.At) > 0 {
			at = s.At[i]
		} else if len(keys) > 1 {
			at = .04 + math.Pow(float64(i)/float64(len(keys)-1), .8)*.86
		}
		m.Stages = append(m.Stages, MaturityStage{keys[i], st.Label, st.Text, st.Number, at})
	}
	key := func(i *int) (string, error) {
		if i == nil {
			return "", nil
		}
		if *i < 0 || *i >= len(keys) {
			return "", fmt.Errorf("invalid maturity source index")
		}
		return keys[*i], nil
	}
	if m.Current, e = key(s.Active); e != nil {
		return m, s, b, e
	}
	if s.Inflection != nil {
		k, e := key(s.Inflection)
		if e != nil {
			return m, s, b, e
		}
		m.Inflection = &MaturityMarker{k, s.InflectionLabel}
	}
	for _, pair := range []struct {
		src *maturitySourceMarker
		dst **MaturityMarker
	}{{s.Here, &m.Here}, {s.Target, &m.Target}} {
		if pair.src != nil {
			i := pair.src.Stage
			if i == nil {
				i = s.Active
			}
			if i == nil {
				v := 0
				i = &v
			}
			k, e := key(i)
			if e != nil {
				return m, s, b, e
			}
			*pair.dst = &MaturityMarker{k, pair.src.Label}
		}
	}
	if s.Branch != nil {
		if s.Branch.From == nil {
			return m, s, b, fmt.Errorf("maturity branch requires from")
		}
		k, e := key(s.Branch.From)
		if e != nil {
			return m, s, b, e
		}
		m.Branch = &MaturityBranch{k, s.Branch.Label, s.Branch.Text, s.Branch.Number}
	}
	axisLabel := s.AxisLabel
	m.Layout = MaturityLayout{n.Placement.Rect, s.Shape, s.Headroom, s.LabelW, s.Axis, &axisLabel}
	return m, s, b, validateMaturity(m)
}
func validateMaturity(m MaturityModel) error {
	if len(m.Stages) < 1 || len(m.Stages) > 12 {
		return fmt.Errorf("maturity requires 1..12 stages; measured fit determines usable count")
	}
	keys := map[string]bool{}
	last := -1.
	for _, s := range m.Stages {
		if !stableID.MatchString(s.Key) || keys[s.Key] || strings.TrimSpace(s.Label) == "" || len(s.Label) > 4096 || len(s.Text) > 16384 || !curveFinite(s.At) || s.At < 0 || s.At > 1 || s.At <= last {
			return fmt.Errorf("invalid maturity stage identity, text or increasing position %s", s.Key)
		}
		if _, e := maturityNumberValidate(s.Number); e != nil {
			return e
		}
		keys[s.Key] = true
		last = s.At
	}
	if m.Current != "" && !keys[m.Current] {
		return fmt.Errorf("unknown current stage")
	}
	for _, v := range []*MaturityMarker{m.Here, m.Target, m.Inflection} {
		if v != nil && (!keys[v.Stage] || len(v.Label) > 4096) {
			return fmt.Errorf("invalid maturity marker reference")
		}
	}
	if m.Branch != nil {
		if !keys[m.Branch.From] || strings.TrimSpace(m.Branch.Label) == "" || len(m.Branch.Label) > 4096 || len(m.Branch.Text) > 16384 {
			return fmt.Errorf("invalid maturity branch")
		}
		if _, e := maturityNumberValidate(m.Branch.Number); e != nil {
			return e
		}
	}
	for _, p := range []*float64{m.Layout.Shape, m.Layout.Headroom, m.Layout.LabelWidth} {
		if p != nil && !curveFinite(*p) {
			return fmt.Errorf("nonfinite maturity layout")
		}
	}
	if m.Layout.Shape != nil && (*m.Layout.Shape <= 0 || *m.Layout.Shape > 100) {
		return fmt.Errorf("shape intensity must be >0 and <=100")
	}
	if m.Layout.Headroom != nil && *m.Layout.Headroom < 0 {
		return fmt.Errorf("headroom must be nonnegative")
	}
	if m.Layout.LabelWidth != nil && *m.Layout.LabelWidth <= 0 {
		return fmt.Errorf("label width must be positive")
	}
	return nil
}
func maturityNumberValidate(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if strings.TrimSpace(text) == "" || len([]rune(text)) > 24 {
			return "", fmt.Errorf("invalid marker number")
		}
		return text, nil
	}
	var v float64
	if json.Unmarshal(raw, &v) != nil || !curveFinite(v) {
		return "", fmt.Errorf("marker n must be nonempty string or finite number")
	}
	return "", nil
}
func maturitySpacing(m *MaturityModel, kind string) error {
	if kind != "even" && kind != "curve" {
		return fmt.Errorf("spacing requires even or curve")
	}
	for i := range m.Stages {
		u := .5
		if len(m.Stages) > 1 {
			u = float64(i) / float64(len(m.Stages)-1)
			if kind == "curve" {
				u = math.Pow(u, .8)
			}
			u = .04 + u*.86
		}
		m.Stages[i].At = u
	}
	return nil
}
func applyMaturity(m *MaturityModel, op MaturityOperation) error {
	keys := []string{}
	for _, s := range m.Stages {
		keys = append(keys, s.Key)
	}
	i := curveIndex(keys, op.Key)
	switch op.Action + "/" + op.Entity {
	case "set/stage":
		if op.Stage == nil || op.Stage.Key != op.Key || !stableID.MatchString(op.Key) {
			return fmt.Errorf("set stage requires matching stable key and complete stage")
		}
		if i < 0 {
			m.Stages = append(m.Stages, *op.Stage)
		} else {
			m.Stages[i] = *op.Stage
		}
	case "remove/stage":
		if i < 0 {
			return fmt.Errorf("unknown stage")
		}
		incident := m.Current == op.Key || m.Here != nil && m.Here.Stage == op.Key || m.Target != nil && m.Target.Stage == op.Key || m.Inflection != nil && m.Inflection.Stage == op.Key || m.Branch != nil && m.Branch.From == op.Key
		if incident && !op.Cascade {
			return fmt.Errorf("stage has markers/branch; explicitly cascade or reassign first")
		}
		if m.Current == op.Key {
			m.Current = ""
		}
		if m.Here != nil && m.Here.Stage == op.Key {
			m.Here = nil
		}
		if m.Target != nil && m.Target.Stage == op.Key {
			m.Target = nil
		}
		if m.Inflection != nil && m.Inflection.Stage == op.Key {
			m.Inflection = nil
		}
		if m.Branch != nil && m.Branch.From == op.Key {
			m.Branch = nil
		}
		m.Stages = append(m.Stages[:i], m.Stages[i+1:]...)
	case "reorder/stage":
		indices, e := curveReorder(keys, op.Order)
		if e != nil {
			return e
		}
		old := m.Stages
		m.Stages = nil
		for _, j := range indices {
			m.Stages = append(m.Stages, old[j])
		}
		if op.Spacing == "" {
			return fmt.Errorf("reorder requires explicit spacing even/curve to avoid nonmonotonic source positions")
		}
		return maturitySpacing(m, op.Spacing)
	case "set/spacing":
		return maturitySpacing(m, op.Spacing)
	case "set/current":
		if i < 0 {
			return fmt.Errorf("unknown current stage")
		}
		m.Current = op.Key
	case "remove/current":
		m.Current = ""
	case "set/here", "set/target", "set/inflection":
		if op.Marker == nil {
			return fmt.Errorf("set marker requires marker")
		}
		switch op.Entity {
		case "here":
			m.Here = op.Marker
		case "target":
			m.Target = op.Marker
		case "inflection":
			m.Inflection = op.Marker
		}
	case "remove/here":
		m.Here = nil
	case "remove/target":
		m.Target = nil
	case "remove/inflection":
		m.Inflection = nil
	case "set/branch":
		if op.Branch == nil {
			return fmt.Errorf("set branch requires branch")
		}
		m.Branch = op.Branch
	case "remove/branch":
		m.Branch = nil
	case "set/layout":
		if op.Layout == nil {
			return fmt.Errorf("layout required")
		}
		l := op.Layout
		if l.Rect != nil {
			m.Layout.Rect = l.Rect
		}
		if l.Shape != nil {
			m.Layout.Shape = l.Shape
		}
		if l.Headroom != nil {
			m.Layout.Headroom = l.Headroom
		}
		if l.LabelWidth != nil {
			m.Layout.LabelWidth = l.LabelWidth
		}
		if l.Axis != nil {
			m.Layout.Axis = l.Axis
		}
		if l.AxisLabel != nil {
			m.Layout.AxisLabel = l.AxisLabel
		}
	default:
		return fmt.Errorf("unsupported maturity operation")
	}
	return nil
}
func maturityLower(n *Node, m MaturityModel, s maturitySourceSpec) error {
	if e := validateMaturity(m); e != nil {
		return e
	}
	s.Stages = nil
	s.At = nil
	s.Active = nil
	s.Here = nil
	s.Target = nil
	s.Inflection = nil
	s.InflectionLabel = ""
	s.Branch = nil
	keys := []string{}
	for _, st := range m.Stages {
		s.Stages = append(s.Stages, maturitySourceStage{st.Label, st.Text, st.Number})
		s.At = append(s.At, st.At)
		keys = append(keys, st.Key)
	}
	idx := func(k string) *int { i := curveIndex(keys, k); return &i }
	if m.Current != "" {
		s.Active = idx(m.Current)
	}
	if m.Here != nil {
		s.Here = &maturitySourceMarker{idx(m.Here.Stage), m.Here.Label}
	}
	if m.Target != nil {
		s.Target = &maturitySourceMarker{idx(m.Target.Stage), m.Target.Label}
	}
	if m.Inflection != nil {
		s.Inflection = idx(m.Inflection.Stage)
		s.InflectionLabel = m.Inflection.Label
	}
	if m.Branch != nil {
		s.Branch = &maturitySourceBranch{idx(m.Branch.From), m.Branch.Label, m.Branch.Text, m.Branch.Number}
	}
	s.Shape, s.Headroom, s.LabelW, s.Axis = m.Layout.Shape, m.Layout.Headroom, m.Layout.LabelWidth, m.Layout.Axis
	if m.Layout.AxisLabel != nil {
		s.AxisLabel = *m.Layout.AxisLabel
	}
	n.Arguments = nil
	if e := json.Unmarshal(canonical(s), &n.Arguments); e != nil {
		return e
	}
	curveSetKeys(n, "stages", keys)
	n.Placement.Rect = m.Layout.Rect
	return nil
}
func InspectMaturity(p *Project, slideID, nodeID, bundle, engine string) (MaturityInspection, error) {
	out := MaturityInspection{Schema: "pptxgengo.maturity-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	i, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	n, e := curveNode(&t, nodeID, "maturity")
	if e != nil {
		return out, e
	}
	var bound bool
	out.Model, _, bound, e = maturityRead(n, p.Document.Slides[i].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Bound arguments require explicit first materialize/source operation."
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func PatchMaturity(p *Project, slideID string, patch MaturityPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeMaturityPatch(canonical(patch), "maturity-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("maturity source hash mismatch; inspect again")
	}
	i, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := curveNode(&clone, patch.NodeID, "maturity")
	if e != nil {
		return empty, e
	}
	m, s, bound, e := maturityRead(n, p.Document.Slides[i].Values)
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
			return empty, fmt.Errorf("maturity bound arguments require explicit materialize source")
		}
		if e = applyMaturity(&m, op); e != nil {
			return empty, e
		}
	}
	if e = maturityLower(n, m, s); e != nil {
		return empty, e
	}
	return CompositionCandidate(p, slideID, "maturity", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
