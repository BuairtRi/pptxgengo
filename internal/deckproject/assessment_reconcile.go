package deckproject

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"sort"
	"strconv"
)

const AssessmentSemanticDecisionsSchema = "pptxgengo.assessment-semantic-decisions.v1"

type AssessmentSemanticProposal struct {
	ID         string `json:"id"`
	Row        string `json:"row"`
	Column     string `json:"column"`
	Before     *int   `json:"before"`
	Proposed   *int   `json:"proposed"`
	Missing    bool   `json:"missing"`
	NativeText string `json:"native_text"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}
type AssessmentSemanticReport struct {
	Schema                 string                       `json:"schema"`
	ProjectID              string                       `json:"project_id"`
	SlideID                string                       `json:"slide_id"`
	NodeID                 string                       `json:"node_id"`
	SourceSHA256           string                       `json:"source_sha256"`
	GeometryReportSHA256   string                       `json:"geometry_report_sha256"`
	EditedPPTXSHA256       string                       `json:"edited_pptx_sha256"`
	TableShapeToken        string                       `json:"table_shape_token"`
	Policy                 string                       `json:"policy"`
	Proposals              []AssessmentSemanticProposal `json:"proposals"`
	ManualReview           []TextReconciliationIssue    `json:"manual_review"`
	UnresolvedGeometryIDs  []string                     `json:"unresolved_geometry_ids"`
	UnresolvedTextIDs      []string                     `json:"unresolved_text_ids"`
	UnresolvedStructureIDs []string                     `json:"unresolved_structure_ids"`
}

func AssessmentSemanticReportHash(r AssessmentSemanticReport) string { return digest(canonical(r)) }
func assessmentTableShell(n *xmlNode) *xmlNode {
	if n == nil {
		return nil
	}
	clone := *n
	clone.Children = nil
	if n.Name.Space == drawingML && n.Name.Local == "t" {
		clone.Text = ""
	}
	for _, c := range n.Children {
		if c.Name.Local == "xfrm" && (c.Name.Space == drawingML || c.Name.Space == lineagePML) {
			continue
		}
		clone.Children = append(clone.Children, assessmentTableShell(c))
	}
	return &clone
}
func semanticNativeCells(shape *xmlNode) (map[[2]int]string, error) {
	out := map[[2]int]string{}
	for _, paragraph := range nativeParagraphs(shape) {
		if paragraph.Address.TableRow == nil || paragraph.Address.TableColumn == nil || len(paragraph.ReviewItems) > 0 || paragraph.Address.Paragraph != 0 {
			return nil, fmt.Errorf("assessment requires one plain paragraph per native cell")
		}
		key := [2]int{*paragraph.Address.TableRow, *paragraph.Address.TableColumn}
		if _, exists := out[key]; exists {
			return nil, fmt.Errorf("ambiguous native cell address")
		}
		for _, run := range paragraph.Runs {
			if run.Kind != "r" {
				return nil, fmt.Errorf("native score cells cannot contain fields/breaks")
			}
		}
		out[key] = paragraph.Text
	}
	return out, nil
}
func ProposeAssessmentSemantics(p *Project, packet *TextReviewPacket, slideID, nodeID, bundle, engine string) (AssessmentSemanticReport, error) {
	out := AssessmentSemanticReport{Schema: "pptxgengo.assessment-semantic-report.v1", ProjectID: p.Document.ID, SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash(), Policy: "exact authenticated native table; unchanged keyed axes/grid/domain/source; visible canonical ordinal score or blank; explicit score acceptance; regenerate fills/legend", Proposals: []AssessmentSemanticProposal{}, UnresolvedGeometryIDs: []string{}, UnresolvedTextIDs: []string{}, UnresolvedStructureIDs: []string{}}
	packet, b, e := verifiedGanttPacket(p, packet, bundle, engine)
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
	n, e := assessmentNode(&t, nodeID)
	if e != nil {
		return out, e
	}
	model, bound, e := assessmentSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound || !model.ShowScores {
		return out, fmt.Errorf("assessment semantic review requires materialized source with visible scores and a fresh baseline")
	}
	if len(p.Document.Slides[idx].NativeGeometry) > 0 || len(p.Document.Slides[idx].NativeOrder) > 0 {
		return out, fmt.Errorf("assessment semantic review requires clean native layout baseline")
	}
	original, e := InspectNativeLineage(b.files["deck.pptx"], b.Objects)
	if e != nil {
		return out, e
	}
	edited, e := InspectNativeLineage(packet.Edited, b.Objects)
	if e != nil {
		return out, e
	}
	nativeName := ganttNativeNodeID(t.Nodes, nodeID, "") + ".matrix.native"
	var before, after *NativeLineageObject
	for i, o := range original.Objects {
		if o.NativeName == nativeName && b.Objects.Lineage.Slides[o.SlideToken] == slideID {
			if before != nil {
				return out, fmt.Errorf("ambiguous baseline assessment table")
			}
			before = &original.Objects[i]
		}
	}
	if before == nil || before.Kind != "graphicFrame" {
		return out, fmt.Errorf("authenticated assessment table missing")
	}
	out.TableShapeToken = before.ShapeToken
	for i, o := range edited.Objects {
		if o.ShapeToken == before.ShapeToken {
			if after != nil {
				return out, fmt.Errorf("copied native assessment table is ambiguous")
			}
			after = &edited.Objects[i]
		}
	}
	if after == nil || after.Kind != before.Kind || after.ParentToken != before.ParentToken || after.NativeName != before.NativeName {
		return out, fmt.Errorf("native table removed, renamed or reparented")
	}
	for _, issue := range edited.Issues {
		if issue.ShapeToken == before.ShapeToken || issue.SlideToken == before.SlideToken && issue.ShapeToken == "" {
			return out, fmt.Errorf("assessment native table identity issue: %s", issue.Kind)
		}
	}
	a, e := assessmentNativeShell(before.shape)
	if e != nil {
		return out, e
	}
	bShell, e := assessmentNativeShell(after.shape)
	if e != nil {
		return out, e
	}
	if mismatch := nativeViewMismatch(reconcileStructureView(a), reconcileStructureView(bShell), "/graphicFrame"); mismatch != "" {
		return out, fmt.Errorf("native assessment table grid/style/structure changed; use explicit source interpretation (%s)", mismatch)
	}
	baselineCells, e := semanticNativeCells(before.shape)
	if e != nil {
		return out, e
	}
	nativeCells, e := semanticNativeCells(after.shape)
	if e != nil {
		return out, e
	}
	count := (len(model.Rows) + 1) * (len(model.Columns) + 1)
	if len(baselineCells) != count || len(nativeCells) != count {
		return out, fmt.Errorf("native table axes/count changed")
	}
	for c := 0; c <= len(model.Columns); c++ {
		if nativeCells[[2]int{0, c}] != baselineCells[[2]int{0, c}] {
			return out, fmt.Errorf("native assessment header changed; review keyed source axes explicitly")
		}
	}
	for r, row := range model.Rows {
		if nativeCells[[2]int{r + 1, 0}] != baselineCells[[2]int{r + 1, 0}] {
			return out, fmt.Errorf("native row identity/label changed; review source axes explicitly")
		}
		for c, col := range model.Columns {
			key := [2]int{r + 1, c + 1}
			source := row.Scores[col.Key]
			sourceText := ""
			if source != nil {
				sourceText = strconv.Itoa(*source)
			}
			if baselineCells[key] != sourceText {
				return out, fmt.Errorf("baseline score does not match source facts")
			}
			native := nativeCells[key]
			if native == sourceText {
				continue
			}
			q := AssessmentSemanticProposal{Row: row.Key, Column: col.Key, Before: source, NativeText: native, Status: "manual_review", Reason: "Native score must be a canonical ordinal integer within the authored domain, or a blank explicitly meaning unassessed."}
			if native == "" {
				q.Missing = true
				q.Status = "proposed"
				q.Reason = "Confirm that blank means unassessed; zero remains a numeric score."
			} else if value, e := strconv.Atoi(native); e == nil && strconv.Itoa(value) == native && value >= 0 && value <= model.Domain.Max {
				v := value
				q.Proposed = &v
				q.Status = "proposed"
				q.Reason = "Confirm the visible ordinal score. Source rebuild regenerates its heat fill and legend."
			}
			q.ID = digest(canonical(struct{ Source, Packet, Row, Col, Value string }{p.SourceHash(), packet.ReportSHA256, row.Key, col.Key, native}))
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
func DecodeAssessmentSemanticDecisions(raw []byte, file string) (GanttSemanticDecisions, error) {
	return decodeSemanticDecisions(raw, file, AssessmentSemanticDecisionsSchema, "set_score")
}
func AdoptAssessmentSemantics(p *Project, packet *TextReviewPacket, slideID, nodeID string, raw []byte, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	report, e := ProposeAssessmentSemantics(p, packet, slideID, nodeID, bundle, engine)
	if e != nil {
		return empty, e
	}
	decisions, e := DecodeAssessmentSemanticDecisions(raw, "assessment-semantic-decisions")
	if e != nil {
		return empty, e
	}
	if decisions.ReportSHA256 != AssessmentSemanticReportHash(report) {
		return empty, fmt.Errorf("assessment semantic report hash mismatch")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := assessmentNode(&clone, nodeID)
	if e != nil {
		return empty, e
	}
	model, _, e := assessmentSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	proposals := map[string]AssessmentSemanticProposal{}
	for _, q := range report.Proposals {
		proposals[q.ID] = q
	}
	for _, d := range decisions.Decisions {
		q, ok := proposals[d.ProposalID]
		if !ok {
			return empty, fmt.Errorf("unknown assessment semantic proposal")
		}
		if d.Action == "keep_source" {
			continue
		}
		if q.Status != "proposed" {
			return empty, fmt.Errorf("invalid native score requires explicit source patch")
		}
		for i := range model.Rows {
			if model.Rows[i].Key == q.Row {
				model.Rows[i].Scores[q.Column] = q.Proposed
			}
		}
	}
	encoded, e := json.Marshal(model)
	if e != nil {
		return empty, e
	}
	var args map[string]any
	if e = json.Unmarshal(encoded, &args); e != nil {
		return empty, e
	}
	for _, k := range []string{"type", "x", "y", "w", "h"} {
		delete(args, k)
	}
	if geometry, exists := n.Arguments[wmdesign.SceneSourceGeometryArgument]; exists {
		args[wmdesign.SceneSourceGeometryArgument] = geometry
	}
	n.Arguments = args
	return semanticEvidenceCandidate(p, packet, slideID, "assessment-semantic-reconcile", decisions.Actor, decisions.Reason, clone, report, raw, bundle, engine, apply)
}
