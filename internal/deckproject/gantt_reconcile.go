package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

func GanttSemanticReportHash(report GanttSemanticReport) string { return digest(canonical(report)) }

const GanttSemanticDecisionsSchema = "pptxgengo.gantt-semantic-decisions.v1"

// A proposal is evidence about possible intent, never an instruction inferred
// from placement. Coordinates are periods, not calendar dates.
type GanttSemanticProposal struct {
	ID              string   `json:"id"`
	GeometryFieldID string   `json:"geometry_field_id"`
	SourceObject    string   `json:"source_object"`
	Entity          string   `json:"entity"`
	Group           string   `json:"group,omitempty"`
	Lane            string   `json:"lane,omitempty"`
	Key             string   `json:"key"`
	BeforeFrom      *float64 `json:"before_from,omitempty"`
	BeforeTo        *float64 `json:"before_to,omitempty"`
	BeforeAt        *float64 `json:"before_at,omitempty"`
	ProposedFrom    *float64 `json:"proposed_from,omitempty"`
	ProposedTo      *float64 `json:"proposed_to,omitempty"`
	ProposedAt      *float64 `json:"proposed_at,omitempty"`
	Status          string   `json:"status"`
	Reason          string   `json:"reason"`
}
type GanttSemanticReport struct {
	UnresolvedTextIDs      []string                  `json:"unresolved_text_ids"`
	UnresolvedStructureIDs []string                  `json:"unresolved_structure_ids"`
	AxisParent             string                    `json:"axis_parent"`
	Schema                 string                    `json:"schema"`
	ProjectID              string                    `json:"project_id"`
	SlideID                string                    `json:"slide_id"`
	NativeNodeID           string                    `json:"native_node_id"`
	NodeID                 string                    `json:"node_id"`
	SourceSHA256           string                    `json:"source_sha256"`
	GeometryReportSHA256   string                    `json:"geometry_report_sha256"`
	BaselineReceiptSHA256  string                    `json:"baseline_receipt_sha256"`
	EditedPPTXSHA256       string                    `json:"edited_pptx_sha256"`
	Timebase               string                    `json:"timebase"`
	Policy                 string                    `json:"policy"`
	AxisLeftPT             float64                   `json:"axis_left_pt"`
	PeriodWidthPT          float64                   `json:"period_width_pt"`
	Proposals              []GanttSemanticProposal   `json:"proposals"`
	ManualReview           []TextReconciliationIssue `json:"manual_review"`
	UnresolvedGeometryIDs  []string                  `json:"unresolved_geometry_ids"`
}
type GanttSemanticDecision struct {
	ProposalID string `json:"proposal_id"`
	Action     string `json:"action"`
	Reason     string `json:"reason"`
}
type GanttSemanticDecisions struct {
	Schema       string                  `json:"schema"`
	ReportSHA256 string                  `json:"report_sha256"`
	Actor        string                  `json:"actor"`
	Reason       string                  `json:"reason"`
	Decisions    []GanttSemanticDecision `json:"decisions"`
}

func verifiedGanttPacket(p *Project, packet *TextReviewPacket, bundle, engine string) (*TextReviewPacket, *TextBaseline, error) {
	if packet == nil {
		return nil, nil, fmt.Errorf("gantt.reconcile_packet_required")
	}
	fresh, e := ReadTextReviewPacket(packet.Root)
	if e != nil {
		return nil, nil, e
	}
	if fresh.ReportSHA256 != packet.ReportSHA256 {
		return nil, nil, fmt.Errorf("gantt.packet_changed_since_read")
	}
	if fresh.Report.GeometryScope == "" {
		return nil, nil, fmt.Errorf("gantt.requires_geometry_review_packet")
	}
	if p.SourceHash() != fresh.Report.CurrentSourceSHA256 || digest(p.Canonical) != fresh.Report.CurrentSemanticSHA256 {
		return nil, nil, fmt.Errorf("gantt.source_changed_since_review")
	}
	b, e := ReadTextBaseline(p, fresh.Report.BaselineBuildID, fresh.Report.BaselineReceiptSHA256)
	if e != nil {
		return nil, nil, e
	}
	if e = verifyTextPacketBaseline(fresh, b); e != nil {
		return nil, nil, e
	}
	// This first slice intentionally refuses concurrent source edits. They change
	// the temporal axis/model; a fresh build provides an unambiguous new baseline.
	if digest(p.Canonical) != b.Receipt.SemanticSHA256 {
		return nil, nil, fmt.Errorf("gantt.source_changed_since_baseline: rebuild before semantic review")
	}
	_, lock, e := ReadLock(p)
	if e != nil {
		return nil, nil, e
	}
	if digest(lock) != fresh.Report.LockSHA256 {
		return nil, nil, fmt.Errorf("gantt.toolchain_changed")
	}
	mappings, e := packetStructureMappings(fresh)
	if e != nil {
		return nil, nil, e
	}
	analysis, e := prepareMappedNativeCopies(fresh.Edited, b, mappings)
	if e != nil {
		return nil, nil, e
	}
	report, e := ReconcileText(p, b, analysis)
	if e != nil {
		return nil, nil, e
	}
	if e = addGeometryReconciliation(p, b, analysis, &report, bundle, engine, mappings...); e != nil {
		return nil, nil, e
	}
	report.EditedPPTXSHA256 = digest(fresh.Edited)
	if !bytes.Equal(canonical(report), canonical(fresh.Report)) {
		return nil, nil, fmt.Errorf("gantt.report_does_not_match_verified_inputs")
	}
	return fresh, b, nil
}

func plainTimelineGeometry(g *NativeGeometry) bool {
	return g != nil && g.Child == nil && g.Rotation == 0 && !g.FlipH && !g.FlipV && g.Route == nil
}
func sameVerticalGeometry(a, b *NativeGeometry) bool {
	return plainTimelineGeometry(a) && plainTimelineGeometry(b) && a.Kind == b.Kind && a.Parent == b.Parent && math.Abs(a.Y-b.Y) < .0001 && math.Abs(a.H-b.H) < .0001
}
func semanticPeriod(x, left, width float64) float64 {
	// Fixed six decimal places preserve fractional periods; there is deliberately
	// no integer/week/date snapping. DrawingML precision is 1/12700 pt.
	return math.Round((x-left)/width*1e6) / 1e6
}

// Local source groups create native namespaces even when the compiler flattens
// them. Resolve the exact qualified identity rather than matching a leaf suffix.
func ganttNativeNodeID(nodes []Node, nodeID, prefix string) string {
	for _, n := range nodes {
		id := prefix + n.ID
		if n.ID == nodeID {
			return id
		}
		if found := ganttNativeNodeID(n.Nodes, nodeID, id+"."); found != "" {
			return found
		}
	}
	return ""
}

// ProposeGanttSemantics replays receipt-backed geometry evidence. Only a single
// solid task bar or gate's vertical guide can establish a bounded proposal.
func ProposeGanttSemantics(p *Project, packet *TextReviewPacket, slideID, nodeID, bundle, engine string) (GanttSemanticReport, error) {
	out := GanttSemanticReport{Schema: "pptxgengo.gantt-semantic-report.v1", ProjectID: p.Document.ID, SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash(), Timebase: "periods", Policy: "unchanged_tagged_axis; horizontal_only; fractional_periods_rounded_6_decimals; no_date_or_membership_inference", Proposals: []GanttSemanticProposal{}, ManualReview: []TextReconciliationIssue{}, UnresolvedGeometryIDs: []string{}, UnresolvedTextIDs: []string{}, UnresolvedStructureIDs: []string{}}
	packet, _, e := verifiedGanttPacket(p, packet, bundle, engine)
	if e != nil {
		return out, e
	}
	out.GeometryReportSHA256 = packet.ReportSHA256
	out.BaselineReceiptSHA256 = packet.Report.BaselineReceiptSHA256
	out.EditedPPTXSHA256 = packet.Report.EditedPPTXSHA256
	out.ManualReview = packet.Report.ManualReview
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	n, e := ganttNode(&t, nodeID)
	if e != nil {
		return out, e
	}
	nativeID := ganttNativeNodeID(t.Nodes, nodeID, "")
	if nativeID == "" {
		return out, fmt.Errorf("gantt.native_node_namespace_missing")
	}
	out.NativeNodeID = nativeID
	s, keys, bound, e := ganttSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		return out, fmt.Errorf("gantt.semantic_reconcile_requires_materialized_source: explicitly materialize and rebuild first")
	}
	if len(p.Document.Slides[idx].NativeGeometry) > 0 || len(p.Document.Slides[idx].NativeOrder) > 0 {
		return out, fmt.Errorf("gantt.semantic_reconcile_requires_clean_native_layout_baseline")
	}
	for _, f := range packet.Report.Fields {
		if f.Status == "native_only" || f.Status == "conflict" {
			out.UnresolvedTextIDs = append(out.UnresolvedTextIDs, f.ID)
		}
	}
	for _, f := range packet.Report.Structure {
		if f.Status == "native_only" || f.Status == "conflict" || f.Status == "manual_review" {
			out.UnresolvedStructureIDs = append(out.UnresolvedStructureIDs, f.ID)
		}
	}
	blockedTokens := map[string]bool{}
	for _, issue := range packet.Report.ManualReview {
		if issue.ShapeToken != "" {
			blockedTokens[issue.ShapeToken] = true
		}
	}
	fields := map[string]GeometryReconciliationField{}
	for _, f := range packet.Report.Geometry {
		if f.SlideID == slideID && f.Property == "transform" {
			fields[f.Name] = f
		}
	}
	// Every period label is a tagged rectangle whose extent establishes one unit.
	// All must remain unchanged; moving the axis is not a schedule edit.
	for i, key := range keys["periods/labels"] {
		f, ok := fields[nativeID+".periods."+key+".label"]
		if !ok || f.Status != "no_op" || !plainTimelineGeometry(f.Baseline) || f.Baseline.W <= 0 {
			return out, fmt.Errorf("gantt.timeline_axis_missing_or_changed: %s present=%v status=%s geometry=%+v; rebuild or restore period headers", nativeID+".periods."+key+".label", ok, f.Status, f.Baseline)
		}
		if i == 0 {
			out.AxisParent = f.Baseline.Parent
			out.AxisLeftPT = f.Baseline.X
			out.PeriodWidthPT = f.Baseline.W
		}
		if f.Baseline.Parent != out.AxisParent || math.Abs(f.Baseline.X-(out.AxisLeftPT+float64(i)*out.PeriodWidthPT)) > .001 || math.Abs(f.Baseline.W-out.PeriodWidthPT) > .001 {
			return out, fmt.Errorf("gantt.timeline_axis_not_uniform")
		}
	}
	// All anchors and candidate bars must inhabit the same unchanged parent
	// coordinate space. Walk and verify every group ancestor; do not infer dates
	// from a group movement/rescale or mix slide-space with child-space units.
	for parent := out.AxisParent; parent != ""; {
		f, ok := fields[parent]
		if !ok || f.Status != "no_op" || f.Baseline == nil || f.Baseline.Kind != "grpSp" || f.Baseline.Child == nil || f.Baseline.Rotation != 0 || f.Baseline.FlipH || f.Baseline.FlipV {
			return out, fmt.Errorf("gantt.timeline_parent_missing_or_changed")
		}
		parent = f.Baseline.Parent
	}
	if len(keys["periods/labels"]) == 0 {
		return out, fmt.Errorf("gantt.timeline_axis_missing")
	}
	supported := map[string]GanttSemanticProposal{}
	for _, g := range s.Groups {
		for _, l := range g.Lanes {
			for _, it := range l.Items {
				if ganttTaskMeaning(it) == nil && it.Kind != "" && !it.Hatch && it.Progress == nil && it.SoftStart == 0 && it.SoftEnd == 0 {
					a, b := it.From, it.To
					name := nativeID + ".groups." + g.Key + ".lanes." + l.Key + ".items." + it.Key + ".segment-0"
					supported[name] = GanttSemanticProposal{Entity: "task", Group: g.Key, Lane: l.Key, Key: it.Key, BeforeFrom: &a, BeforeTo: &b}
				}
			}
		}
	}
	for _, g := range s.Gates {
		at := g.At
		key := ganttKey(g.Key)
		supported[nativeID+".gates."+key+".line"] = GanttSemanticProposal{Entity: "gate", Key: key, BeforeAt: &at}
	}
	for _, f := range packet.Report.Geometry {
		parentField, parentExists := fields[func() string {
			if f.Baseline != nil {
				return f.Baseline.Parent
			}
			return ""
		}()]
		groupedMovement := parentExists && parentField.Status == "native_only" && f.Status == "no_op"
		if f.SlideID != slideID || (f.Status != "native_only" && f.Status != "conflict" && f.Status != "manual_review" && !groupedMovement) {
			continue
		}
		proposal, ok := supported[f.Name]
		if !ok {
			out.UnresolvedGeometryIDs = append(out.UnresolvedGeometryIDs, f.ID)
			continue
		}
		proposal.GeometryFieldID = f.ID
		proposal.SourceObject = f.Name
		proposal.Status = "manual_review"
		proposal.Reason = "A horizontal-only edit of the original single bar/guide is required; rotation, grouping, vertical/lane changes and source conflicts remain explicit review."
		beforeGeom, afterGeom, frameOK := ganttItemAxisGeometry(f, fields, out.AxisParent, proposal.Entity)
		if (f.Status == "native_only" || groupedMovement) && !blockedTokens[f.ShapeToken] && (!parentExists || !blockedTokens[parentField.ShapeToken]) && frameOK && sameVerticalGeometry(beforeGeom, afterGeom) {
			if groupedMovement {
				proposal.GeometryFieldID = parentField.ID
			}
			a := semanticPeriod(afterGeom.X, out.AxisLeftPT, out.PeriodWidthPT)
			if proposal.Entity == "task" {
				b := semanticPeriod(afterGeom.X+afterGeom.W, out.AxisLeftPT, out.PeriodWidthPT)
				if math.Abs(beforeGeom.X-(out.AxisLeftPT+*proposal.BeforeFrom*out.PeriodWidthPT)) < .001 && math.Abs(beforeGeom.W-(*proposal.BeforeTo-*proposal.BeforeFrom)*out.PeriodWidthPT) < .001 && a >= 0 && b <= float64(len(s.Periods.Labels)) && b > a {
					proposal.ProposedFrom = &a
					proposal.ProposedTo = &b
					proposal.Status = "proposed"
					proposal.Reason = "Single solid bar maps to this fractional period interval. Explicit retime acceptance regenerates associated labels/layout; geometry is not adopted as an override."
				}
			} else if math.Abs(beforeGeom.X-(out.AxisLeftPT+*proposal.BeforeAt*out.PeriodWidthPT)) < .001 && afterGeom.W == 0 && a >= 0 && a <= float64(len(s.Periods.Labels)) {
				proposal.ProposedAt = &a
				proposal.Status = "proposed"
				proposal.Reason = "Unchanged vertical gate guide maps to this fractional period. Explicit retime acceptance regenerates its guide, chip and label."
			}
		}
		proposal.ID = digest(canonical(proposal))
		out.Proposals = append(out.Proposals, proposal)
	}
	sort.Slice(out.Proposals, func(i, j int) bool { return out.Proposals[i].ID < out.Proposals[j].ID })
	sort.Strings(out.UnresolvedGeometryIDs)
	sort.Strings(out.UnresolvedTextIDs)
	sort.Strings(out.UnresolvedStructureIDs)
	return out, nil
}

func DecodeGanttSemanticDecisions(raw []byte, file string) (GanttSemanticDecisions, error) {
	var out GanttSemanticDecisions
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("gantt.semantic_decisions_too_large")
	}
	// Reuse the strict one-document YAML decoding and presence/type checks.
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document, extra yaml.Node
	if e := decoder.Decode(&document); e != nil {
		return out, e
	}
	if len(document.Content) != 1 {
		return out, fmt.Errorf("gantt.empty_semantic_decisions")
	}
	if e := decoder.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("gantt.semantic_decisions_require_one_document")
	}
	doc := &document
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
	if out.Schema != GanttSemanticDecisionsSchema || !shaPattern.MatchString(out.ReportSHA256) || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || strings.TrimSpace(out.Reason) == "" || len(out.Reason) > 4096 || len(out.Decisions) == 0 || len(out.Decisions) > 500 {
		return out, fmt.Errorf("gantt.invalid_semantic_decisions")
	}
	seen := map[string]bool{}
	for _, d := range out.Decisions {
		if !shaPattern.MatchString(d.ProposalID) || seen[d.ProposalID] || (d.Action != "retime" && d.Action != "keep_source") || strings.TrimSpace(d.Reason) == "" || len(d.Reason) > 4096 {
			return out, fmt.Errorf("gantt.invalid_or_duplicate_semantic_decision")
		}
		seen[d.ProposalID] = true
	}
	return out, nil
}

// AdoptGanttSemantics measures a complete model regenerated from reviewed facts.
// Other native edits remain in the retained evidence and unresolved report.
func AdoptGanttSemantics(p *Project, packet *TextReviewPacket, slideID, nodeID string, raw []byte, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	if packet == nil {
		return empty, fmt.Errorf("gantt.reconcile_packet_required")
	}
	fresh, e := ReadTextReviewPacket(packet.Root)
	if e != nil {
		return empty, e
	}
	if fresh.ReportSHA256 != packet.ReportSHA256 {
		return empty, fmt.Errorf("gantt.packet_changed_since_read")
	}
	packet = fresh
	report, e := ProposeGanttSemantics(p, packet, slideID, nodeID, bundle, engine)
	if e != nil {
		return empty, e
	}
	decisions, e := DecodeGanttSemanticDecisions(raw, "gantt-semantic-decisions")
	if e != nil {
		return empty, e
	}
	if decisions.ReportSHA256 != digest(canonical(report)) {
		return empty, fmt.Errorf("gantt.semantic_report_hash_mismatch")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := ganttNode(&clone, nodeID)
	if e != nil {
		return empty, e
	}
	s, keys, _, e := ganttSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	proposals := map[string]GanttSemanticProposal{}
	for _, q := range report.Proposals {
		proposals[q.ID] = q
	}
	for _, d := range decisions.Decisions {
		q, ok := proposals[d.ProposalID]
		if !ok {
			return empty, fmt.Errorf("gantt.unknown_semantic_proposal")
		}
		if d.Action == "keep_source" {
			continue
		}
		if q.Status != "proposed" {
			return empty, fmt.Errorf("gantt.manual_finding_cannot_retime")
		}
		if q.Entity == "task" {
			gi := ganttGroupIndex(&s, q.Group)
			li := ganttLaneIndex(&s.Groups[gi], q.Lane)
			ii := ganttTaskIndex(&s.Groups[gi].Lanes[li], q.Key)
			s.Groups[gi].Lanes[li].Items[ii].From = *q.ProposedFrom
			s.Groups[gi].Lanes[li].Items[ii].To = *q.ProposedTo
		} else {
			for i := range s.Gates {
				if ganttKey(s.Gates[i].Key) == q.Key {
					s.Gates[i].At = *q.ProposedAt
				}
			}
		}
	}
	var args map[string]any
	if e = json.Unmarshal(canonical(s), &args); e != nil {
		return empty, e
	}
	for _, k := range []string{"type", "x", "y", "w"} {
		delete(args, k)
	}
	if geometry, exists := n.Arguments[wmdesign.SceneSourceGeometryArgument]; exists {
		args[wmdesign.SceneSourceGeometryArgument] = geometry
	}
	n.Arguments, n.Keys = args, keys
	reportRaw := canonical(report)
	evidence := struct {
		Report                 GanttSemanticReport    `json:"report"`
		ReportSHA256           string                 `json:"report_sha256"`
		Decisions              GanttSemanticDecisions `json:"decisions"`
		DecisionsSHA256        string                 `json:"decisions_sha256"`
		RetainedPPTX           string                 `json:"retained_pptx"`
		RetainedReport         string                 `json:"retained_report"`
		RetainedDecisions      string                 `json:"retained_decisions"`
		RetainedGeometryReport string                 `json:"retained_geometry_report"`
	}{report, digest(reportRaw), decisions, digest(raw), "assets/objects/sha256/" + report.EditedPPTXSHA256, "assets/objects/sha256/" + digest(reportRaw), "assets/objects/sha256/" + digest(raw), "assets/objects/sha256/" + packet.ReportSHA256}
	b, e := ReadTextBaseline(p, packet.Report.BaselineBuildID, packet.Report.BaselineReceiptSHA256)
	if e != nil {
		return empty, e
	}
	_, lock, e := ReadLock(p)
	if e != nil {
		return empty, e
	}
	if digest(lock) != packet.Report.LockSHA256 {
		return empty, fmt.Errorf("gantt.toolchain_changed")
	}
	guards := map[string][]byte{p.Document.Toolchain.Lockfile: lock}
	for name, data := range b.files {
		guards["builds/"+b.Receipt.BuildID+"/"+name] = data
	}
	return compositionCandidate(p, slideID, "gantt-semantic-reconcile", decisions.Actor, decisions.Reason, clone, bundle, engine, apply, evidence, map[string][]byte{evidence.RetainedPPTX: packet.Edited, evidence.RetainedReport: reportRaw, evidence.RetainedDecisions: raw, evidence.RetainedGeometryReport: packet.files["report.json"]}, guards)
}

// Rebase only the original direct keyed task group into the unchanged axis
// coordinate space. A real group scale, vertical move, rotation or reparenting
// cannot become authored time merely because its bar happens to look horizontal.
func ganttItemAxisGeometry(f GeometryReconciliationField, fields map[string]GeometryReconciliationField, axisParent, entity string) (*NativeGeometry, *NativeGeometry, bool) {
	if f.Baseline == nil || f.EditedNative == nil {
		return nil, nil, false
	}
	if f.Baseline.Parent == axisParent {
		return f.Baseline, f.EditedNative, true
	}
	if entity != "task" || !strings.HasSuffix(f.Name, ".segment-0") || f.Baseline.Parent != strings.TrimSuffix(f.Name, ".segment-0") || f.EditedNative.Parent != f.Baseline.Parent {
		return nil, nil, false
	}
	group, exists := fields[f.Baseline.Parent]
	if !exists || group.Baseline == nil || group.EditedNative == nil || group.Baseline.Parent != axisParent || group.EditedNative.Parent != axisParent || group.Status != "no_op" && group.Status != "native_only" {
		return nil, nil, false
	}
	affine := func(g *NativeGeometry) ([4]float64, bool) {
		if g.Kind != "grpSp" || g.Child == nil || g.Rotation != 0 || g.FlipH || g.FlipV || g.W <= 0 || g.H <= 0 || g.Child.W <= 0 || g.Child.H <= 0 {
			return [4]float64{}, false
		}
		sx, sy := g.W/g.Child.W, g.H/g.Child.H
		return [4]float64{sx, sy, g.X - sx*g.Child.X, g.Y - sy*g.Child.Y}, true
	}
	before, ok := affine(group.Baseline)
	after, ok2 := affine(group.EditedNative)
	if !ok || !ok2 {
		return nil, nil, false
	}
	for _, i := range []int{0, 1, 3} {
		if math.Abs(before[i]-after[i]) > 0.000001 {
			return nil, nil, false
		}
	}
	rebase := func(g *NativeGeometry, a [4]float64) *NativeGeometry {
		copy := *g
		copy.Parent = axisParent
		copy.X = a[0]*g.X + a[2]
		copy.Y = a[1]*g.Y + a[3]
		copy.W = a[0] * g.W
		copy.H = a[1] * g.H
		return &copy
	}
	return rebase(f.Baseline, before), rebase(f.EditedNative, after), true
}
