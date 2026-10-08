package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const TextReviewDecisionsSchema = "pptxgengo.text-review-decisions.v1"
const TextAdoptionSchema = "pptxgengo.text-adoption-receipt.v1"

type TextReviewDecision struct {
	FieldID string `json:"field_id"`
	Action  string `json:"action"`
	Reason  string `json:"reason"`
}
type TextReviewDecisions struct {
	Schema       string               `json:"schema"`
	ReportSHA256 string               `json:"report_sha256"`
	Actor        string               `json:"actor"`
	Decisions    []TextReviewDecision `json:"decisions"`
}
type AdoptedTextField struct {
	FieldID    string `json:"field_id"`
	SlideID    string `json:"slide_id"`
	SourceSlot string `json:"source_slot"`
	Before     string `json:"before"`
	After      string `json:"after"`
}
type TextAdoptionReceipt struct {
	Schema                string                    `json:"schema"`
	ID                    string                    `json:"id"`
	Actor                 string                    `json:"actor"`
	Created               string                    `json:"created"`
	ReportSHA256          string                    `json:"report_sha256"`
	DecisionsSHA256       string                    `json:"decisions_sha256"`
	BaselineBuildID       string                    `json:"baseline_build_id"`
	BaselineReceiptSHA256 string                    `json:"baseline_receipt_sha256"`
	EditedPPTXSHA256      string                    `json:"edited_pptx_sha256"`
	BeforeSourceSHA256    string                    `json:"before_source_sha256"`
	AfterSourceSHA256     string                    `json:"after_source_sha256"`
	DecisionPath          string                    `json:"decision_path,omitempty"`
	RetainedEdited        string                    `json:"retained_edited,omitempty"`
	RetainedReport        string                    `json:"retained_report,omitempty"`
	RetainedDecisions     string                    `json:"retained_decisions,omitempty"`
	Changed               []AdoptedTextField        `json:"changed"`
	RemainingFieldIDs     []string                  `json:"remaining_field_ids"`
	ManualReview          []TextReconciliationIssue `json:"manual_review"`
	ReviewDecisions       []TextReviewDecision      `json:"review_decisions"`
	Validation            string                    `json:"validation"`
}

func DecodeTextReviewDecisions(raw []byte, filename string) (TextReviewDecisions, error) {
	var out TextReviewDecisions
	if len(raw) > 16<<20 {
		return out, fmt.Errorf("reconcile.decisions_too_large")
	}
	p := &Project{SourcePath: filename, Positions: map[string]Position{}}
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := decoder.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("reconcile.empty_decisions")
	}
	if e := decoder.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("reconcile.decisions_require_one_document")
	}
	value, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = p.shapeType(value, reflect.TypeOf(out), ""); e != nil {
		return out, e
	}
	if e = strictInto(value, &out); e != nil {
		return out, e
	}
	if out.Schema != TextReviewDecisionsSchema || !shaPattern.MatchString(out.ReportSHA256) || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || len(out.Decisions) > 10000 {
		return out, fmt.Errorf("reconcile.invalid_decisions_identity")
	}
	seen := map[string]bool{}
	for _, d := range out.Decisions {
		if !shaPattern.MatchString(d.FieldID) || seen[d.FieldID] || (d.Action != "use_native" && d.Action != "keep_yaml") || strings.TrimSpace(d.Reason) == "" || len(d.Reason) > 4096 {
			return out, fmt.Errorf("reconcile.invalid_or_duplicate_decision")
		}
		seen[d.FieldID] = true
	}
	return out, nil
}

// AdoptTextReviewPacket replays the complete report against its retained source
// snapshot, then applies only reviewed fields to the unchanged current source.
// It retains the exact edited deck, report, decisions and predecessor source.
func AdoptTextReviewPacket(p *Project, packet *TextReviewPacket, decisionsRaw []byte, bundle, engine string) (TextAdoptionReceipt, error) {
	result := TextAdoptionReceipt{Schema: TextAdoptionSchema, Changed: []AdoptedTextField{}, RemainingFieldIDs: []string{}, ManualReview: []TextReconciliationIssue{}, ReviewDecisions: []TextReviewDecision{}, Validation: "named_text_only_native_rebuild_and_review_required"}
	if packet == nil {
		return result, fmt.Errorf("reconcile.review_packet_required")
	}
	freshPacket, e := ReadTextReviewPacket(packet.Root)
	if e != nil {
		return result, e
	}
	if freshPacket.ReportSHA256 != packet.ReportSHA256 || !bytes.Equal(freshPacket.Edited, packet.Edited) {
		return result, fmt.Errorf("reconcile.packet_changed_since_read")
	}
	packet = freshPacket
	decisions, e := DecodeTextReviewDecisions(decisionsRaw, "review-decisions")
	if e != nil {
		return result, e
	}
	if decisions.ReportSHA256 != packet.ReportSHA256 {
		return result, fmt.Errorf("reconcile.decisions_report_hash_mismatch")
	}
	b, e := ReadTextBaseline(p, packet.Report.BaselineBuildID, packet.Report.BaselineReceiptSHA256)
	if e != nil {
		return result, e
	}
	if e = verifyTextPacketBaseline(packet, b); e != nil {
		return result, e
	}
	original, e := projectFromTextPacket(p, packet)
	if e != nil {
		return result, e
	}
	mappings, e := packetStructureMappings(packet)
	if e != nil {
		return result, e
	}
	analysis, e := prepareMappedNativeCopies(packet.Edited, b, mappings)
	if e != nil {
		return result, e
	}
	replayed, e := ReconcileText(original, b, analysis)
	if e != nil {
		return result, e
	}
	if packet.Report.GeometryScope != "" {
		if e = addGeometryReconciliation(original, b, analysis, &replayed, bundle, engine, mappings...); e != nil {
			return result, e
		}
	}
	replayed.EditedPPTXSHA256 = digest(packet.Edited)

	if !bytes.Equal(canonical(replayed), canonical(packet.Report)) {
		return result, fmt.Errorf("reconcile.report_does_not_match_verified_inputs")
	}
	if packet.Report.GeometryScope != "" {
		result.Validation = "reviewed_text_transforms_and_supported_structure_frame_checked_native_review_required"
	}
	result.Actor = decisions.Actor
	result.ReportSHA256 = packet.ReportSHA256
	result.DecisionsSHA256 = digest(decisionsRaw)
	result.BaselineBuildID = b.Receipt.BuildID
	result.BaselineReceiptSHA256 = b.ReceiptSHA256
	result.EditedPPTXSHA256 = packet.Report.EditedPPTXSHA256
	result.BeforeSourceSHA256 = original.SourceHash()
	result.ManualReview = packet.Report.ManualReview
	result.ReviewDecisions = decisions.Decisions
	result.ID = "text-adoption-" + digest(canonical([]string{result.ReportSHA256, result.DecisionsSHA256, result.Actor}))
	changes, changed, e := reviewedTextSourceChanges(original, packet.Report, decisions, bundle, engine)
	if e != nil {
		return result, e
	}
	result.Changed = changed
	candidate, e := loadProject(p.SourcePath, mergeTextOverrides(original.SourceFiles, changes))
	if e != nil {
		return result, e
	}
	result.AfterSourceSHA256 = candidate.SourceHash()
	selected := map[string]bool{}
	for _, d := range decisions.Decisions {
		selected[d.FieldID] = true
	}
	for _, f := range packet.Report.Structure {
		if (f.Status == "native_only" || f.Status == "conflict") && !selected[f.ID] {
			result.RemainingFieldIDs = append(result.RemainingFieldIDs, f.ID)
		}
	}
	for _, f := range packet.Report.Geometry {
		if (f.Status == "native_only" || f.Status == "conflict") && !selected[f.ID] {
			result.RemainingFieldIDs = append(result.RemainingFieldIDs, f.ID)
		}
	}
	for _, f := range packet.Report.Fields {
		if (f.Status == "native_only" || f.Status == "conflict") && !selected[f.ID] {
			result.RemainingFieldIDs = append(result.RemainingFieldIDs, f.ID)
		}
	}
	if len(decisions.Decisions) == 0 {
		if p.SourceHash() != original.SourceHash() {
			return result, fmt.Errorf("reconcile.source_changed_since_review")
		}
		return result, nil
	}
	result.DecisionPath = "decisions/" + result.ID + ".json"
	result.RetainedEdited = "decisions/reconciliations/" + result.ID + "/edited.pptx"
	result.RetainedReport = "decisions/reconciliations/" + result.ID + "/report.json"
	result.RetainedDecisions = "decisions/reconciliations/" + result.ID + "/review-decisions.yaml"
	receiptPath, e := SafePath(p.Root, result.DecisionPath)
	if e != nil {
		return result, e
	}
	if priorRaw, e := readReconcileFile(receiptPath, 64<<20); e == nil {
		var prior TextAdoptionReceipt
		if e = strictInto(json.RawMessage(priorRaw), &prior); e != nil {
			return result, e
		}
		if _, e = time.Parse(time.RFC3339, prior.Created); e != nil {
			return result, fmt.Errorf("reconcile.invalid_existing_receipt_clock")
		}
		result.Created = prior.Created
		if !bytes.Equal(canonical(prior), canonical(result)) {
			return result, fmt.Errorf("reconcile.existing_adoption_receipt_conflict")
		}
		if p.SourceHash() != result.AfterSourceSHA256 {
			return result, fmt.Errorf("reconcile.source_changed_since_adoption")
		}
		for rel, want := range map[string][]byte{result.RetainedEdited: packet.Edited, result.RetainedReport: packet.files["report.json"], result.RetainedDecisions: decisionsRaw} {
			file, e := SafePath(p.Root, rel)
			if e != nil {
				return result, e
			}
			raw, e := readReconcileFile(file, lineageMaxPackage)
			if e != nil || !bytes.Equal(raw, want) {
				return result, fmt.Errorf("reconcile.retained_adoption_input_drift")
			}
		}
		return result, nil
	} else if !os.IsNotExist(e) {
		return result, e
	}
	if p.SourceHash() != original.SourceHash() || !bytes.Equal(p.Canonical, original.Canonical) {
		return result, fmt.Errorf("reconcile.source_changed_since_review: make and review a fresh packet")
	}
	result.Created = time.Now().UTC().Format(time.RFC3339)
	changes[result.DecisionPath] = canonical(result)
	changes[result.RetainedEdited] = packet.Edited
	changes[result.RetainedReport] = packet.files["report.json"]
	changes[result.RetainedDecisions] = decisionsRaw
	lockPath := p.Document.Toolchain.Lockfile
	lockBytes, e := readReconcileFile(filepath.Join(p.Root, filepath.FromSlash(lockPath)), 64<<20)
	if e != nil {
		return result, e
	}
	if digest(lockBytes) != packet.Report.LockSHA256 {
		return result, fmt.Errorf("reconcile.toolchain_changed")
	}
	guarded := map[string][]byte{}
	for name, raw := range b.files {
		guarded["builds/"+b.Receipt.BuildID+"/"+name] = raw
	}
	_, e = commitSourceChangesChecked(p, changes, map[string][]byte{lockPath: lockBytes}, guarded, func(c *Project) error {
		if c.SourceHash() != result.AfterSourceSHA256 {
			return fmt.Errorf("reconcile.candidate_source_identity_changed")
		}
		if _, e := Check(c, bundle, engine); e != nil {
			return e
		}
		ids := []string{}
		seen := map[string]bool{}
		for _, f := range changed {
			if !seen[f.SlideID] {
				ids = append(ids, f.SlideID)
				seen[f.SlideID] = true
			}
		}
		if len(ids) > 0 {
			if e := CheckSlideFit(c, ids, bundle, engine); e != nil {
				return e
			}
		}
		return nil
	})
	return result, e
}
func mergeTextOverrides(before, changed map[string][]byte) map[string][]byte {
	out := map[string][]byte{}
	for k, v := range before {
		out[k] = v
	}
	for k, v := range changed {
		out[k] = v
	}
	return out
}
func projectFromTextPacket(p *Project, packet *TextReviewPacket) (*Project, error) {
	overrides := map[string][]byte{}
	for name, raw := range packet.files {
		if strings.HasPrefix(name, "current/") {
			overrides[strings.TrimPrefix(name, "current/")] = raw
		}
	}
	original, e := loadProject(p.SourcePath, overrides)
	if e != nil {
		return nil, e
	}
	if len(overrides) != len(original.SourceFiles) || original.SourceHash() != packet.Report.CurrentSourceSHA256 || !bytes.Equal(original.Canonical, packet.files["current-source.canonical.json"]) {
		return nil, fmt.Errorf("reconcile.current_snapshot_identity_mismatch")
	}
	for name, raw := range original.SourceFiles {
		if !bytes.Equal(overrides[name], raw) {
			return nil, fmt.Errorf("reconcile.current_snapshot_file_mismatch")
		}
	}
	return original, nil
}
func reviewedTextSourceChanges(p *Project, report TextReconciliationReport, decisions TextReviewDecisions, bundle, engine string) (map[string][]byte, []AdoptedTextField, error) {
	changes := map[string][]byte{}
	changed := []AdoptedTextField{}
	main, e := sourceYAML(p.Raw)
	if e != nil {
		return nil, nil, e
	}
	slides, documents, e := authoredSlides(p, main)
	if e != nil {
		return nil, nil, e
	}
	fields := map[string]TextReconciliationField{}
	for _, f := range report.Fields {
		if _, ok := fields[f.ID]; ok {
			return nil, nil, fmt.Errorf("reconcile.duplicate_report_field_id")
		}
		fields[f.ID] = f
	}
	geometryFields := map[string]GeometryReconciliationField{}
	for _, f := range report.Geometry {
		if _, ok := geometryFields[f.ID]; ok {
			return nil, nil, fmt.Errorf("reconcile.duplicate_geometry_field")
		}
		geometryFields[f.ID] = f
	}
	touched := map[string]bool{}
	selected := map[string]bool{}
	structureIDs := map[string]bool{}
	for _, f := range report.Structure {
		structureIDs[f.ID] = true
	}
	for _, d := range decisions.Decisions {
		if structureIDs[d.FieldID] {
			continue
		}
		if gf, ok := geometryFields[d.FieldID]; ok {
			if selected[d.FieldID] || (gf.Status != "native_only" && gf.Status != "conflict") {
				return nil, nil, fmt.Errorf("reconcile.invalid_geometry_decision")
			}
			selected[d.FieldID] = true
			if d.Action == "keep_yaml" {
				continue
			}
			node := slides[gf.SlideID]
			if node == nil {
				return nil, nil, fmt.Errorf("reconcile.geometry_slide_missing")
			}
			key := "native_geometry"
			var value any = gf.EditedNative
			if gf.EditedNative != nil {
				copy := *gf.EditedNative
				copy.SourceGeometrySHA256 = gf.SourceGeometrySHA256
				value = &copy
			}
			if gf.Property == "paint_order" {
				key = "native_order"
				value = gf.EditedOrder
			} else if gf.Property != "transform" || gf.EditedNative == nil {
				return nil, nil, fmt.Errorf("reconcile.unsupported_geometry_property")
			}
			mapping := mappingNode(node, key)
			if mapping == nil {
				mapping = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				replaceMappingField(node, key, mapping)
			}
			v, err := editYAMLNode(value)
			if err != nil {
				return nil, nil, err
			}
			replaceMappingField(mapping, gf.Name, v)
			for _, sourceSlide := range p.Document.Slides {
				if sourceSlide.ID == gf.SlideID {
					pin, err := editYAMLNode(sourceSlide.Template)
					if err != nil {
						return nil, nil, err
					}
					replaceMappingField(node, "native_geometry_template", pin)
				}
			}
			file := p.SlideFiles[gf.SlideID]
			if file == "" {
				file = filepath.Base(p.SourcePath)
			}
			touched[file] = true
			changed = append(changed, AdoptedTextField{gf.ID, gf.SlideID, key + "/" + gf.Name, string(canonical(gf.CurrentYAML)), string(canonical(value))})
			continue
		}
		f, ok := fields[d.FieldID]
		if !ok || selected[d.FieldID] || (f.Status != "native_only" && f.Status != "conflict") {
			return nil, nil, fmt.Errorf("reconcile.decision_is_not_a_supported_text_proposal")
		}
		selected[d.FieldID] = true
		if d.Action == "keep_yaml" {
			continue
		}
		node, e := authoredTextScalar(p, slides[f.SlideID], f)
		if e != nil {
			return nil, nil, e
		}
		if node.Value != f.CurrentYAML {
			return nil, nil, fmt.Errorf("reconcile.authored_field_value_mismatch")
		}
		node.Value = f.EditedNative
		node.Tag = "!!str"
		if strings.Contains(f.EditedNative, "\n") {
			node.Style = yaml.LiteralStyle
		}
		file := p.SlideFiles[f.SlideID]
		if file == "" {
			file = filepath.Base(p.SourcePath)
		}
		touched[file] = true
		changed = append(changed, AdoptedTextField{f.ID, f.SlideID, f.SourceSlot, f.CurrentYAML, f.EditedNative})
	}
	structureChanged, err := applyReviewedStructure(p, report, decisions, slides, documents, touched, bundle, engine)
	if err != nil {
		return nil, nil, err
	}
	changed = append(changed, structureChanged...)
	for file := range touched {
		raw, e := encodeSourceYAML(documents[file])
		if e != nil {
			return nil, nil, e
		}
		changes[file] = raw
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i].FieldID < changed[j].FieldID })
	return changes, changed, nil
}
func authoredTextScalar(p *Project, slide *yaml.Node, field TextReconciliationField) (*yaml.Node, error) {
	if slide == nil {
		return nil, fmt.Errorf("reconcile.authored_slide_missing")
	}
	parts, e := contentPointerParts(field.SourcePointer)
	if e != nil || len(parts) < 4 || parts[0] != "slides" || parts[2] != "values" {
		return nil, fmt.Errorf("reconcile.invalid_named_source_pointer")
	}
	index, e := strconv.Atoi(parts[1])
	if e != nil || index < 0 || index >= len(p.Document.Slides) || p.Document.Slides[index].ID != field.SlideID {
		return nil, fmt.Errorf("reconcile.source_pointer_slide_mismatch")
	}
	target := parts[2:]
	if node := yamlTextPointer(slide, target); node != nil && node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		return node, nil
	}
	valuePath := ""
	for _, part := range parts[3:] {
		valuePath += "/" + escape(part)
	}
	bindings := mappingNode(slide, "bindings")
	if bindings != nil {
		for i := 0; i+1 < len(bindings.Content); i += 2 {
			alias, binding := bindings.Content[i].Value, bindings.Content[i+1].Value
			if valuePath != binding && !strings.HasPrefix(valuePath, binding+"/") {
				continue
			}
			contentPath := alias
			if !strings.HasPrefix(alias, "/") {
				contentPath = "/" + escape(alias)
			}
			contentPath += strings.TrimPrefix(valuePath, binding)
			contentParts, e := contentPointerParts(contentPath)
			if e != nil {
				return nil, e
			}
			node := yamlTextPointer(slide, append([]string{"content"}, contentParts...))
			if node != nil && node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
				return node, nil
			}
		}
	}
	return nil, fmt.Errorf("reconcile.authored_scalar_origin_unavailable")
}
func yamlTextPointer(node *yaml.Node, parts []string) *yaml.Node {
	current := node
	for _, part := range parts {
		if current == nil {
			return nil
		}
		switch current.Kind {
		case yaml.MappingNode:
			current = mappingNode(current, part)
		case yaml.SequenceNode:
			index, e := strconv.Atoi(part)
			if e != nil || index < 0 || index >= len(current.Content) {
				return nil
			}
			current = current.Content[index]
		default:
			return nil
		}
	}
	return current
}
