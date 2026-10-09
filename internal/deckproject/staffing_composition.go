package deckproject

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"strings"
)

const StaffingPatchSchema = "pptxgengo.staffing-patch.v1"

type StaffingPoint struct {
	Key    string             `json:"key"`
	At     float64            `json:"at"`
	Label  string             `json:"label,omitempty"`
	Values map[string]float64 `json:"values"`
}
type StaffingSeries struct {
	Key        string             `json:"key"`
	Name       string             `json:"name"`
	Fill       string             `json:"fill,omitempty"`
	Style      string             `json:"style,omitempty"`
	Dashed     bool               `json:"dashed,omitempty"`
	Hatch      bool               `json:"hatch,omitempty"`
	LabelPoint string             `json:"label_point,omitempty"`
	HideLabel  bool               `json:"hide_label,omitempty"`
	Values     map[string]float64 `json:"values,omitempty"`
}
type StaffingPhase struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Text  string  `json:"text,omitempty"`
	At    float64 `json:"at"`
	Point string  `json:"point,omitempty"`
}
type StaffingScale struct {
	Unit     string   `json:"unit"`
	TimeUnit string   `json:"time_unit"`
	Source   string   `json:"source"`
	Maximum  *float64 `json:"maximum,omitempty"`
}
type StaffingLayout struct {
	Rect        *wmdesign.Rect `json:"rect,omitempty"`
	Curve       *string        `json:"curve,omitempty"`
	LineBasis   *string        `json:"line_basis,omitempty"`
	Smooth      *float64       `json:"smooth,omitempty"`
	Tension     *float64       `json:"tension,omitempty"`
	PhaseHeight *float64       `json:"phase_height_pt,omitempty"`
	PhaseLabels *bool          `json:"phase_labels,omitempty"`
}
type StaffingModel struct {
	Points []StaffingPoint  `json:"points"`
	Series []StaffingSeries `json:"series"`
	Phases []StaffingPhase  `json:"phases,omitempty"`
	Scale  StaffingScale    `json:"scale"`
	Layout StaffingLayout   `json:"layout"`
}
type StaffingOperation struct {
	Action  string          `json:"action"`
	Entity  string          `json:"entity"`
	Key     string          `json:"key,omitempty"`
	Point   *StaffingPoint  `json:"point,omitempty"`
	Series  *StaffingSeries `json:"series,omitempty"`
	Phase   *StaffingPhase  `json:"phase,omitempty"`
	Scale   *StaffingScale  `json:"scale,omitempty"`
	Layout  *StaffingLayout `json:"layout,omitempty"`
	Order   []string        `json:"order,omitempty"`
	Cascade bool            `json:"cascade,omitempty"`
	Spacing string          `json:"spacing,omitempty"`
	At      []float64       `json:"at,omitempty"`
}
type StaffingPatch struct {
	Schema               string              `json:"schema"`
	ExpectedSourceSHA256 string              `json:"expected_source_sha256"`
	Actor                string              `json:"actor"`
	Reason               string              `json:"reason"`
	NodeID               string              `json:"node_id"`
	Operations           []StaffingOperation `json:"operations"`
}
type StaffingInspection struct {
	Schema          string            `json:"schema"`
	SlideID         string            `json:"slide_id"`
	NodeID          string            `json:"node_id"`
	SourceSHA256    string            `json:"source_sha256"`
	Model           StaffingModel     `json:"model"`
	MutationBlocked string            `json:"mutation_blocked,omitempty"`
	RenderError     string            `json:"render_error,omitempty"`
	Geometry        DiagramInspection `json:"geometry"`
}
type staffingSourceSeries struct {
	Name    string          `json:"name"`
	Values  []float64       `json:"values"`
	Fill    string          `json:"fill,omitempty"`
	Style   string          `json:"style,omitempty"`
	Dashed  bool            `json:"dashed,omitempty"`
	Hatch   bool            `json:"hatch,omitempty"`
	LabelAt json.RawMessage `json:"labelAt,omitempty"`
}
type staffingSourcePhase struct {
	Label string   `json:"label"`
	Sub   string   `json:"sub,omitempty"`
	At    *float64 `json:"at,omitempty"`
	Point string   `json:"point,omitempty"`
}
type staffingSource struct {
	SourceGeometry json.RawMessage        `json:"_source_geometry,omitempty"`
	Series         []staffingSourceSeries `json:"series"`
	Phases         []staffingSourcePhase  `json:"phases,omitempty"`
	At             []float64              `json:"at,omitempty"`
	PointLabels    []string               `json:"pointLabels,omitempty"`
	Unit           string                 `json:"unit,omitempty"`
	TimeUnit       string                 `json:"timeUnit,omitempty"`
	Source         string                 `json:"source,omitempty"`
	Max            *float64               `json:"max,omitempty"`
	Curve          *string                `json:"curve,omitempty"`
	LineBasis      *string                `json:"lineBasis,omitempty"`
	Smooth         *float64               `json:"smooth,omitempty"`
	Tension        *float64               `json:"tension,omitempty"`
	PhaseH         *float64               `json:"phaseH,omitempty"`
	PhaseLabels    *bool                  `json:"phaseLabels,omitempty"`
	CanvasH        *float64               `json:"_h,omitempty"`
}

func staffingFields(a, b string) (string, error) {
	switch a + "/" + b {
	case "materialize/source":
		return "", nil
	case "set/point":
		return "key point", nil
	case "set/series":
		return "key series", nil
	case "set/phase":
		return "key phase", nil
	case "remove/point":
		return "key cascade", nil
	case "remove/series", "remove/phase":
		return "key", nil
	case "reorder/point":
		return "order spacing at", nil
	case "reorder/series", "reorder/phase":
		return "order", nil
	case "set/scale":
		return "scale", nil
	case "set/layout":
		return "layout", nil
	}
	return "", fmt.Errorf("unsupported staffing operation %s/%s", a, b)
}
func DecodeStaffingPatch(raw []byte, file string) (StaffingPatch, error) {
	var out StaffingPatch
	e := decodeCurveDocument(raw, file, &out, StaffingPatchSchema, staffingFields)
	if e == nil && !stableID.MatchString(out.NodeID) {
		e = fmt.Errorf("staffing requires stable node_id")
	}
	return out, e
}
func staffingRead(n *Node, values map[string]any) (StaffingModel, staffingSource, bool, error) {
	var m StaffingModel
	var s staffingSource
	args, b, e := curveSource(n, values)
	if e != nil {
		return m, s, b, e
	}
	if e = strictInto(args, &s); e != nil {
		return m, s, b, e
	}
	if len(s.Series) == 0 {
		return m, s, b, fmt.Errorf("staffing has no series")
	}
	count := len(s.Series[0].Values)
	pk, e := curveKeys(n, "at", count)
	if e != nil {
		return m, s, b, e
	}
	sk, e := curveKeys(n, "series", len(s.Series))
	if e != nil {
		return m, s, b, e
	}
	if len(s.At) > 0 && len(s.At) != count || len(s.PointLabels) > 0 && len(s.PointLabels) != count {
		return m, s, b, fmt.Errorf("staffing point metadata count mismatch")
	}
	for i, k := range pk {
		at := float64(i) / float64(count-1)
		if len(s.At) > 0 {
			at = s.At[i]
		}
		pt := StaffingPoint{Key: k, At: at, Values: map[string]float64{}}
		if len(s.PointLabels) > 0 {
			pt.Label = s.PointLabels[i]
		}
		m.Points = append(m.Points, pt)
	}
	for i, se := range s.Series {
		if len(se.Values) != count {
			return m, s, b, fmt.Errorf("staffing series sample count mismatch")
		}
		out := StaffingSeries{Key: sk[i], Name: se.Name, Fill: se.Fill, Style: se.Style, Dashed: se.Dashed, Hatch: se.Hatch}
		label := 0
		if string(se.LabelAt) == "false" {
			out.HideLabel = true
		} else if len(se.LabelAt) > 0 {
			if e = json.Unmarshal(se.LabelAt, &label); e != nil {
				return m, s, b, e
			}
		}
		if label < 0 || label >= len(pk) {
			return m, s, b, fmt.Errorf("staffing label point outside sample range")
		}
		if !out.HideLabel {
			out.LabelPoint = pk[label]
		}
		m.Series = append(m.Series, out)
		for j, v := range se.Values {
			m.Points[j].Values[sk[i]] = v
		}
	}
	if len(s.Phases) > 0 {
		keys, e := curveKeys(n, "phases", len(s.Phases))
		if e != nil {
			return m, s, b, e
		}
		for i, p := range s.Phases {
			at := float64(i) / float64(len(s.Phases))
			if p.At != nil {
				at = *p.At
			}
			m.Phases = append(m.Phases, StaffingPhase{keys[i], p.Label, p.Sub, at, p.Point})
		}
	}
	m.Scale = StaffingScale{s.Unit, s.TimeUnit, s.Source, s.Max}
	m.Layout = StaffingLayout{Rect: n.Placement.Rect, Curve: s.Curve, LineBasis: s.LineBasis, Smooth: s.Smooth, Tension: s.Tension, PhaseHeight: s.PhaseH, PhaseLabels: s.PhaseLabels}
	if m.Layout.LineBasis == nil {
		basis := "independent"
		for _, se := range s.Series {
			if se.Style == "line" {
				basis = "above_stack"
			}
		}
		m.Layout.LineBasis = &basis
	}
	return m, s, b, validateStaffing(m, false)
}
func staffingPointKeys(m StaffingModel) []string {
	k := []string{}
	for _, v := range m.Points {
		k = append(k, v.Key)
	}
	return k
}
func staffingSeriesKeys(m StaffingModel) []string {
	k := []string{}
	for _, v := range m.Series {
		k = append(k, v.Key)
	}
	return k
}
func staffingPhaseKeys(m StaffingModel) []string {
	k := []string{}
	for _, v := range m.Phases {
		k = append(k, v.Key)
	}
	return k
}
func validateStaffing(m StaffingModel, authored bool) error {
	if len(m.Points) < 2 || len(m.Points) > 512 || len(m.Series) < 1 || len(m.Series) > 12 || len(m.Phases) > 12 {
		return fmt.Errorf("staffing requires 2..512 points, 1..12 series, <=12 phases")
	}
	if authored && (strings.TrimSpace(m.Scale.Unit) == "" || strings.TrimSpace(m.Scale.TimeUnit) == "" || strings.TrimSpace(m.Scale.Source) == "") {
		return fmt.Errorf("staffing mutations require explicit scale unit, time_unit and source; legacy counts cannot establish units")
	}
	if len(m.Scale.Unit) > 128 || len(m.Scale.TimeUnit) > 128 || len(m.Scale.Source) > 1024 {
		return fmt.Errorf("staffing scale label too long")
	}
	sk := staffingSeriesKeys(m)
	pk := staffingPointKeys(m)
	for _, keys := range [][]string{sk, pk, staffingPhaseKeys(m)} {
		seen := map[string]bool{}
		for _, k := range keys {
			if !stableID.MatchString(k) || seen[k] {
				return fmt.Errorf("invalid/duplicate staffing identity %s", k)
			}
			seen[k] = true
		}
	}
	for _, se := range m.Series {
		if strings.TrimSpace(se.Name) == "" || len(se.Name) > 4096 || se.Style != "" && se.Style != "area" && se.Style != "line" || len(se.Values) > 0 {
			return fmt.Errorf("invalid staffing series source/model")
		}
		if se.HideLabel && se.LabelPoint != "" || !se.HideLabel && curveIndex(pk, se.LabelPoint) < 0 {
			return fmt.Errorf("series must name existing label_point or hide_label")
		}
	}
	last := -1.
	for _, pt := range m.Points {
		if !curveFinite(pt.At) || pt.At < 0 || pt.At > 1 || pt.At <= last || len(pt.Label) > 128 || len(pt.Values) != len(sk) {
			return fmt.Errorf("invalid staffing point spacing/values %s", pt.Key)
		}
		last = pt.At
		total := 0.
		for _, se := range m.Series {
			v, ok := pt.Values[se.Key]
			if !ok || !curveFinite(v) || v < 0 || v > 1e6 {
				return fmt.Errorf("staffing requires nonnegative finite authored values for each series")
			}
			if se.Style != "line" {
				total += v
			}
			top := v
			if se.Style == "line" && m.Layout.LineBasis != nil && *m.Layout.LineBasis == "above_stack" {
				top += total
			}
			if m.Scale.Maximum != nil && top > *m.Scale.Maximum {
				return fmt.Errorf("staffing series exceeds explicit maximum")
			}
		}
		if m.Scale.Maximum != nil && total > *m.Scale.Maximum {
			return fmt.Errorf("staffing stack exceeds explicit maximum")
		}
	}
	if m.Scale.Maximum != nil && (!curveFinite(*m.Scale.Maximum) || *m.Scale.Maximum <= 0) {
		return fmt.Errorf("staffing maximum must be positive finite")
	}
	last = -1.
	for _, p := range m.Phases {
		at := p.At
		if p.Point != "" {
			j := curveIndex(pk, p.Point)
			if j < 0 {
				return fmt.Errorf("phase references unknown point")
			}
			at = m.Points[j].At
		}
		if !curveFinite(at) || at < 0 || at >= 1 || at <= last || strings.TrimSpace(p.Label) == "" || len(p.Label) > 4096 || len(p.Text) > 16384 {
			return fmt.Errorf("invalid staffing phase boundary/text")
		}
		last = at
	}
	for _, v := range []*float64{m.Layout.Smooth, m.Layout.Tension} {
		if v != nil && (!curveFinite(*v) || *v < 0 || *v > 1) {
			return fmt.Errorf("smooth/tension must be within 0..1")
		}
	}
	if m.Layout.LineBasis != nil && *m.Layout.LineBasis != "independent" && *m.Layout.LineBasis != "above_stack" {
		return fmt.Errorf("line_basis must be independent/above_stack")
	}
	if m.Layout.Curve != nil && *m.Layout.Curve != "monotone" && *m.Layout.Curve != "catmull" {
		return fmt.Errorf("curve must be monotone/catmull")
	}
	if m.Layout.PhaseHeight != nil && (!curveFinite(*m.Layout.PhaseHeight) || *m.Layout.PhaseHeight < 0) {
		return fmt.Errorf("phase height must be nonnegative")
	}
	return nil
}
func applyStaffing(m *StaffingModel, op StaffingOperation) error {
	pk, sk, fk := staffingPointKeys(*m), staffingSeriesKeys(*m), staffingPhaseKeys(*m)
	pi, si, fi := curveIndex(pk, op.Key), curveIndex(sk, op.Key), curveIndex(fk, op.Key)
	switch op.Action + "/" + op.Entity {
	case "set/point":
		if op.Point == nil || op.Point.Key != op.Key || !stableID.MatchString(op.Key) {
			return fmt.Errorf("point requires matching stable key and complete values")
		}
		if pi < 0 {
			m.Points = append(m.Points, *op.Point)
		} else {
			m.Points[pi] = *op.Point
		}
	case "set/series":
		if op.Series == nil || op.Series.Key != op.Key || !stableID.MatchString(op.Key) {
			return fmt.Errorf("series requires matching stable key")
		}
		se := *op.Series
		if len(se.Values) > 0 {
			if len(se.Values) != len(pk) {
				return fmt.Errorf("series values must enumerate each existing point")
			}
			for i, pt := range m.Points {
				v, ok := se.Values[pt.Key]
				if !ok {
					return fmt.Errorf("series missing point value")
				}
				m.Points[i].Values[op.Key] = v
			}
			se.Values = nil
		} else if si < 0 {
			return fmt.Errorf("new series requires explicit values for every point")
		}
		if si < 0 {
			m.Series = append(m.Series, se)
		} else {
			m.Series[si] = se
		}
	case "remove/series":
		if si < 0 {
			return fmt.Errorf("unknown series")
		}
		m.Series = append(m.Series[:si], m.Series[si+1:]...)
		for i := range m.Points {
			delete(m.Points[i].Values, op.Key)
		}
	case "remove/point":
		if pi < 0 {
			return fmt.Errorf("unknown point")
		}
		incident := false
		for _, se := range m.Series {
			incident = incident || se.LabelPoint == op.Key
		}
		for _, ph := range m.Phases {
			incident = incident || ph.Point == op.Key
		}
		if incident && !op.Cascade {
			return fmt.Errorf("point has label/phase references; cascade or reassign first")
		}
		for i := range m.Series {
			if m.Series[i].LabelPoint == op.Key {
				m.Series[i].LabelPoint = ""
				m.Series[i].HideLabel = true
			}
		}
		kept := []StaffingPhase{}
		for _, ph := range m.Phases {
			if ph.Point != op.Key {
				kept = append(kept, ph)
			}
		}
		m.Phases = kept
		m.Points = append(m.Points[:pi], m.Points[pi+1:]...)
	case "set/phase":
		if op.Phase == nil || op.Phase.Key != op.Key {
			return fmt.Errorf("phase requires matching key")
		}
		if fi < 0 {
			m.Phases = append(m.Phases, *op.Phase)
		} else {
			m.Phases[fi] = *op.Phase
		}
	case "remove/phase":
		if fi < 0 {
			return fmt.Errorf("unknown phase")
		}
		m.Phases = append(m.Phases[:fi], m.Phases[fi+1:]...)
	case "reorder/point":
		ii, e := curveReorder(pk, op.Order)
		if e != nil {
			return e
		}
		if op.Spacing != "even" && op.Spacing != "explicit" {
			return fmt.Errorf("point reorder requires spacing even or explicit")
		}
		if op.Spacing == "even" && len(op.At) > 0 || op.Spacing == "explicit" && len(op.At) != len(pk) {
			return fmt.Errorf("explicit spacing requires one at per point; even refuses authored at")
		}
		old := m.Points
		m.Points = nil
		for i, j := range ii {
			pt := old[j]
			pt.At = float64(i) / float64(len(ii)-1)
			if op.Spacing == "explicit" {
				pt.At = op.At[i]
			}
			m.Points = append(m.Points, pt)
		}
	case "reorder/series":
		ii, e := curveReorder(sk, op.Order)
		if e != nil {
			return e
		}
		old := m.Series
		m.Series = nil
		for _, j := range ii {
			m.Series = append(m.Series, old[j])
		}
	case "reorder/phase":
		ii, e := curveReorder(fk, op.Order)
		if e != nil {
			return e
		}
		old := m.Phases
		m.Phases = nil
		for _, j := range ii {
			m.Phases = append(m.Phases, old[j])
		}
	case "set/scale":
		if op.Scale == nil {
			return fmt.Errorf("scale required")
		}
		m.Scale = *op.Scale
	case "set/layout":
		if op.Layout == nil {
			return fmt.Errorf("layout required")
		}
		l := op.Layout
		if l.Rect != nil {
			m.Layout.Rect = l.Rect
		}
		if l.LineBasis != nil {
			m.Layout.LineBasis = l.LineBasis
		}
		if l.Curve != nil {
			m.Layout.Curve = l.Curve
		}
		if l.Smooth != nil {
			m.Layout.Smooth = l.Smooth
		}
		if l.Tension != nil {
			m.Layout.Tension = l.Tension
		}
		if l.PhaseHeight != nil {
			m.Layout.PhaseHeight = l.PhaseHeight
		}
		if l.PhaseLabels != nil {
			m.Layout.PhaseLabels = l.PhaseLabels
		}
	default:
		return fmt.Errorf("unsupported staffing operation")
	}
	return nil
}
func staffingLower(n *Node, m StaffingModel, s staffingSource) error {
	if e := validateStaffing(m, true); e != nil {
		return e
	}
	s.Series = nil
	s.Phases = nil
	s.At = nil
	s.PointLabels = nil
	pk := staffingPointKeys(m)
	for _, pt := range m.Points {
		s.At = append(s.At, pt.At)
		s.PointLabels = append(s.PointLabels, pt.Label)
	}
	for _, se := range m.Series {
		v := staffingSourceSeries{Name: se.Name, Fill: se.Fill, Style: se.Style, Dashed: se.Dashed, Hatch: se.Hatch}
		for _, pt := range m.Points {
			v.Values = append(v.Values, pt.Values[se.Key])
		}
		v.LabelAt = json.RawMessage("false")
		if !se.HideLabel {
			v.LabelAt = canonical(curveIndex(pk, se.LabelPoint))
		}
		s.Series = append(s.Series, v)
	}
	for _, ph := range m.Phases {
		at := ph.At
		if ph.Point != "" {
			at = m.Points[curveIndex(pk, ph.Point)].At
		}
		s.Phases = append(s.Phases, staffingSourcePhase{ph.Label, ph.Text, &at, ph.Point})
	}
	s.Unit, s.TimeUnit, s.Source, s.Max = m.Scale.Unit, m.Scale.TimeUnit, m.Scale.Source, m.Scale.Maximum
	s.LineBasis = m.Layout.LineBasis
	s.Curve, s.Smooth, s.Tension, s.PhaseH, s.PhaseLabels = m.Layout.Curve, m.Layout.Smooth, m.Layout.Tension, m.Layout.PhaseHeight, m.Layout.PhaseLabels
	n.Arguments = nil
	if e := json.Unmarshal(canonical(s), &n.Arguments); e != nil {
		return e
	}
	curveSetKeys(n, "at", pk)
	curveSetKeys(n, "series", staffingSeriesKeys(m))
	curveSetKeys(n, "phases", staffingPhaseKeys(m))
	n.Placement.Rect = m.Layout.Rect
	return nil
}
func InspectStaffing(p *Project, slideID, nodeID, bundle, engine string) (StaffingInspection, error) {
	out := StaffingInspection{Schema: "pptxgengo.staffing-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	i, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	n, e := curveNode(&t, nodeID, "teamcurve")
	if e != nil {
		return out, e
	}
	var bound bool
	out.Model, _, bound, e = staffingRead(n, p.Document.Slides[i].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Bound arguments require explicit first materialize/source."
	}
	if out.Model.Scale.Unit == "" || out.Model.Scale.TimeUnit == "" || out.Model.Scale.Source == "" {
		out.MutationBlocked += " Explicit set/scale declaration required; legacy units are unspecified."
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func PatchStaffing(p *Project, slideID string, patch StaffingPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeStaffingPatch(canonical(patch), "staffing-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("staffing source hash mismatch; inspect again")
	}
	i, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := curveNode(&clone, patch.NodeID, "teamcurve")
	if e != nil {
		return empty, e
	}
	m, s, bound, e := staffingRead(n, p.Document.Slides[i].Values)
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
			return empty, fmt.Errorf("staffing bound source requires materialize")
		}
		if e = applyStaffing(&m, op); e != nil {
			return empty, e
		}
	}
	if e = staffingLower(n, m, s); e != nil {
		return empty, e
	}
	preview, e := CompositionCandidate(p, slideID, "staffing", patch.Actor, patch.Reason, clone, bundle, engine, false)
	if e != nil {
		return preview, e
	}
	if e = staffingScaleContext(preview.Inspection, patch.NodeID); e != nil {
		return preview, e
	}
	if !apply {
		return preview, nil
	}
	return CompositionCandidate(p, slideID, "staffing", patch.Actor, patch.Reason, clone, bundle, engine, true)
}
