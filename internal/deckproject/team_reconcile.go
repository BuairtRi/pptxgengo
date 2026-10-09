package deckproject

import (
	"fmt"
	"math"
	"sort"
)

const TeamSemanticDecisionsSchema = "pptxgengo.team-semantic-decisions.v1"

type TeamSemanticProposal struct {
	ID               string   `json:"id"`
	Role             string   `json:"role"`
	Label            string   `json:"label"`
	FromPod          string   `json:"from_pod"`
	ToPod            string   `json:"to_pod,omitempty"`
	SourceObjects    []string `json:"source_objects"`
	GeometryFieldIDs []string `json:"geometry_field_ids"`
	Status           string   `json:"status"`
	Reason           string   `json:"reason"`
}
type TeamSemanticReport struct {
	Schema                 string                    `json:"schema"`
	ProjectID              string                    `json:"project_id"`
	SlideID                string                    `json:"slide_id"`
	SourceSHA256           string                    `json:"source_sha256"`
	GeometryReportSHA256   string                    `json:"geometry_report_sha256"`
	EditedPPTXSHA256       string                    `json:"edited_pptx_sha256"`
	Policy                 string                    `json:"policy"`
	Proposals              []TeamSemanticProposal    `json:"proposals"`
	ManualReview           []TextReconciliationIssue `json:"manual_review"`
	UnresolvedGeometryIDs  []string                  `json:"unresolved_geometry_ids"`
	UnresolvedTextIDs      []string                  `json:"unresolved_text_ids"`
	UnresolvedStructureIDs []string                  `json:"unresolved_structure_ids"`
}

func TeamSemanticReportHash(r TeamSemanticReport) string { return digest(canonical(r)) }
func sameTranslatedPair(a, b GeometryReconciliationField) bool {
	if a.Status != "native_only" || b.Status != "native_only" || !plainTimelineGeometry(a.Baseline) || !plainTimelineGeometry(a.EditedNative) || !plainTimelineGeometry(b.Baseline) || !plainTimelineGeometry(b.EditedNative) {
		return false
	}
	for _, f := range []GeometryReconciliationField{a, b} {
		if f.Baseline.Kind != f.EditedNative.Kind || f.Baseline.Parent != f.EditedNative.Parent || math.Abs(f.Baseline.W-f.EditedNative.W) > .001 || math.Abs(f.Baseline.H-f.EditedNative.H) > .001 {
			return false
		}
	}
	return math.Abs((a.EditedNative.X-a.Baseline.X)-(b.EditedNative.X-b.Baseline.X)) < .001 && math.Abs((a.EditedNative.Y-a.Baseline.Y)-(b.EditedNative.Y-b.Baseline.Y)) < .001
}

// PowerPoint recomputes a group's outer/child envelope when a child is moved
// outside it. Equal child-to-parent affine transforms retain child geometry;
// the changed envelope still remains unresolved native evidence.
func semanticEquivalentParentFrame(f GeometryReconciliationField) bool {
	if f.Baseline == nil || f.EditedNative == nil || f.Baseline.Kind != "grpSp" || f.EditedNative.Kind != "grpSp" || f.Baseline.Parent != f.EditedNative.Parent || f.Baseline.Child == nil || f.EditedNative.Child == nil {
		return false
	}
	for _, g := range []*NativeGeometry{f.Baseline, f.EditedNative} {
		if g.Rotation != 0 || g.FlipH || g.FlipV || g.W <= 0 || g.H <= 0 || g.Child.W <= 0 || g.Child.H <= 0 {
			return false
		}
	}
	if f.Status == "no_op" {
		return true
	}
	if f.Status != "native_only" {
		return false
	}
	affine := func(g *NativeGeometry) [4]float64 {
		sx, sy := g.W/g.Child.W, g.H/g.Child.H
		return [4]float64{sx, sy, g.X - sx*g.Child.X, g.Y - sy*g.Child.Y}
	}
	before, after := affine(f.Baseline), affine(f.EditedNative)
	for i := range before {
		if math.IsNaN(before[i]) || math.IsInf(before[i], 0) || math.IsNaN(after[i]) || math.IsInf(after[i], 0) || math.Abs(before[i]-after[i]) > 0.000001 {
			return false
		}
	}
	return true
}
func semanticUnchangedParents(fields map[string]GeometryReconciliationField, parent string) bool {
	for parent != "" {
		f, ok := fields[parent]
		if !ok || !semanticEquivalentParentFrame(f) {
			return false
		}
		parent = f.Baseline.Parent
	}
	return true
}
func ProposeTeamSemantics(p *Project, packet *TextReviewPacket, slideID, bundle, engine string) (TeamSemanticReport, error) {
	out := TeamSemanticReport{Schema: "pptxgengo.team-semantic-report.v1", ProjectID: p.Document.ID, SlideID: slideID, SourceSHA256: p.SourceHash(), Policy: "exact tagged role surface+text translation; unchanged source/pod/frame; one containing target allocation; explicit membership confirmation; no reporting inference", Proposals: []TeamSemanticProposal{}, UnresolvedGeometryIDs: []string{}, UnresolvedTextIDs: []string{}, UnresolvedStructureIDs: []string{}}
	packet, _, e := verifiedGanttPacket(p, packet, bundle, engine)
	if e != nil {
		return out, e
	}
	out.GeometryReportSHA256 = packet.ReportSHA256
	out.EditedPPTXSHA256 = packet.Report.EditedPPTXSHA256
	out.ManualReview = packet.Report.ManualReview
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	slide := p.Document.Slides[idx]
	if len(slide.NativeGeometry) > 0 || len(slide.NativeOrder) > 0 {
		return out, fmt.Errorf("team semantic review requires clean native layout baseline")
	}
	inspect, e := InspectTeam(p, slideID, bundle, engine)
	if e != nil {
		return out, e
	}
	fields := map[string]GeometryReconciliationField{}
	objects := map[string]*geometryObject{}
	blocked := map[string]bool{}
	for _, f := range packet.Report.Geometry {
		if f.SlideID == slideID && f.Property == "transform" {
			fields[f.Name] = f
			if f.Baseline != nil {
				objects[f.Name] = &geometryObject{geometry: *f.Baseline}
			}
		}
	}
	for _, issue := range packet.Report.ManualReview {
		if issue.ShapeToken != "" {
			blocked[issue.ShapeToken] = true
		}
	}
	pods := []TeamComponent{}
	nativeNames := map[string]string{}
	for _, c := range inspect.Components {
		if c.Type != "pod" {
			continue
		}
		if c.BoundArguments {
			return out, fmt.Errorf("team semantic review requires explicitly materialized pods and a fresh baseline")
		}
		if c.Placement == nil || c.Placement.Zone != "body" || c.Placement.Rect == nil {
			return out, fmt.Errorf("pod semantic review requires explicit body allocations")
		}
		native := ganttNativeNodeID(t.Nodes, c.ID, "")
		if native != c.ID {
			return out, fmt.Errorf("nested pod semantic review requires explicit source membership patch")
		}
		pods = append(pods, c)
		nativeNames[c.ID] = native
	}
	if len(pods) < 2 {
		return out, fmt.Errorf("membership review requires two or more typed pods")
	}

	for _, c := range pods {
		roles, _ := teamSlice(c.ResolvedArguments["roles"])
		for i, key := range c.Keys["roles"] {
			prefix := nativeNames[c.ID] + ".roles." + key
			surface, ok := fields[prefix+".surface"]
			text, ok2 := fields[prefix+".text"]
			if !ok || !ok2 || surface.Status == "no_op" && text.Status == "no_op" {
				continue
			}
			q := TeamSemanticProposal{Role: key, Label: roles[i].(string), FromPod: c.ID, SourceObjects: []string{prefix + ".surface", prefix + ".text"}, GeometryFieldIDs: []string{surface.ID, text.ID}, Status: "manual_review", Reason: "Move the original role surface and text together without resizing, source changes, grouping changes or formatting edits."}
			if sameTranslatedPair(surface, text) && !blocked[surface.ShapeToken] && !blocked[text.ShapeToken] && semanticUnchangedParents(fields, surface.Baseline.Parent) && semanticUnchangedParents(fields, text.Baseline.Parent) {
				base := objects[prefix+".surface"]
				original := base.geometry
				base.geometry = *surface.EditedNative
				matrix, err := nativeOuterMatrix(objects, prefix+".surface")
				base.geometry = original
				if err != nil {
					return out, err
				}
				bounds := matrix.bounds(nativeGeometryAllocation(*surface.EditedNative))
				targets := []string{}
				for _, target := range pods {
					if target.ID == c.ID {
						continue
					}
					targetNative := nativeNames[target.ID]
					outline, exists := fields[targetNative+".outline"]
					band, hasBand := fields[targetNative+".band"]
					title, hasTitle := fields[targetNative+".title"]
					if !exists || !hasBand || !hasTitle || outline.Status != "no_op" || band.Status != "no_op" || title.Status != "no_op" || !semanticUnchangedParents(fields, outline.Baseline.Parent) {
						continue
					}
					r := *target.Placement.Rect
					r.X += inspect.Diagram.Frame.Body.X
					r.Y += inspect.Diagram.Frame.Body.Y
					r.Y += 42
					r.H -= 42
					if bounds.X >= r.X-.001 && bounds.Y >= r.Y-.001 && bounds.X+bounds.W <= r.X+r.W+.001 && bounds.Y+bounds.H <= r.Y+r.H+.001 {
						targets = append(targets, target.ID)
					}
				}
				if len(targets) == 1 {
					q.ToPod = targets[0]
					q.Status = "proposed"
					q.Reason = "The original role surface and text moved together fully inside one unchanged target pod allocation. Confirm reassignment explicitly; proximity alone is not membership."
				} else {
					q.Reason = "The role is not fully inside exactly one unchanged target pod allocation; retain geometry and choose an explicit source interpretation."
				}
			}
			q.ID = digest(canonical(struct{ Source, Packet, Role, From, To string }{p.SourceHash(), packet.ReportSHA256, key, c.ID, q.ToPod}))
			out.Proposals = append(out.Proposals, q)
		}
	}
	for _, f := range packet.Report.Geometry {
		if f.SlideID == slideID && f.Status != "no_op" {
			out.UnresolvedGeometryIDs = append(out.UnresolvedGeometryIDs, f.ID)
		}
	}
	for _, f := range packet.Report.Fields {
		if f.SlideID == slideID && f.Status != "no_op" {
			out.UnresolvedTextIDs = append(out.UnresolvedTextIDs, f.ID)
		}
	}
	for _, f := range packet.Report.Structure {
		if f.SlideID == slideID && f.Status != "no_op" {
			out.UnresolvedStructureIDs = append(out.UnresolvedStructureIDs, f.ID)
		}
	}
	sort.Slice(out.Proposals, func(i, j int) bool { return out.Proposals[i].ID < out.Proposals[j].ID })
	return out, nil
}

// The shared strict decision envelope is kept schema-specific. Membership is an
// explicit domain decision, never the default consequence of geometry adoption.
func DecodeTeamSemanticDecisions(raw []byte, file string) (GanttSemanticDecisions, error) {
	return decodeSemanticDecisions(raw, file, TeamSemanticDecisionsSchema, "reassign")
}
func AdoptTeamSemantics(p *Project, packet *TextReviewPacket, slideID string, raw []byte, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	report, e := ProposeTeamSemantics(p, packet, slideID, bundle, engine)
	if e != nil {
		return empty, e
	}
	decisions, e := DecodeTeamSemanticDecisions(raw, "team-semantic-decisions")
	if e != nil {
		return empty, e
	}
	if decisions.ReportSHA256 != TeamSemanticReportHash(report) {
		return empty, fmt.Errorf("team semantic report hash mismatch")
	}
	_, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	proposals := map[string]TeamSemanticProposal{}
	for _, q := range report.Proposals {
		proposals[q.ID] = q
	}
	for _, d := range decisions.Decisions {
		q, ok := proposals[d.ProposalID]
		if !ok {
			return empty, fmt.Errorf("unknown team semantic proposal")
		}
		if d.Action == "keep_source" {
			continue
		}
		if q.Status != "proposed" {
			return empty, fmt.Errorf("ambiguous native role movement requires an explicit family source patch")
		}
		if e = applyTeamOperation(&clone, TeamOperation{Action: "reassign-role", Component: q.FromPod, ID: q.Role, Target: q.ToPod}); e != nil {
			return empty, e
		}
	}
	return semanticEvidenceCandidate(p, packet, slideID, "team-semantic-reconcile", decisions.Actor, decisions.Reason, clone, report, raw, bundle, engine, apply)
}
