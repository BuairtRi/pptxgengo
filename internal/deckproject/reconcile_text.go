package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const TextReconciliationSchema = "pptxgengo.text-reconciliation.v1"

type TextReconciliationField struct {
	ID            string `json:"id"`
	FieldIdentity string `json:"field_identity"`
	ShapeToken    string `json:"shape_token"`
	SlideID       string `json:"slide_id"`
	LogicalID     string `json:"logical_id"`
	SourceSlot    string `json:"source_slot"`
	SourcePointer string `json:"current_source_pointer,omitempty"`
	Baseline      string `json:"baseline"`
	CurrentYAML   string `json:"current_yaml"`
	EditedNative  string `json:"edited_native"`
	Status        string `json:"status"`
	Reason        string `json:"reason,omitempty"`
}
type TextReconciliationIssue struct {
	Kind       string `json:"kind"`
	ShapeToken string `json:"shape_token,omitempty"`
	SlideID    string `json:"slide_id,omitempty"`
	NativePart string `json:"native_part,omitempty"`
	Detail     string `json:"detail"`
}
type TextReconciliationReport struct {
	Schema                string                    `json:"schema"`
	ProjectID             string                    `json:"project_id"`
	BaselineBuildID       string                    `json:"baseline_build_id"`
	BaselineReceiptSHA256 string                    `json:"baseline_receipt_sha256"`
	BaselinePPTXSHA256    string                    `json:"baseline_pptx_sha256"`
	CurrentSourceSHA256   string                    `json:"current_source_sha256"`
	CurrentSemanticSHA256 string                    `json:"current_semantic_sha256"`
	EditedPPTXSHA256      string                    `json:"edited_pptx_sha256"`
	LockSHA256            string                    `json:"lock_sha256"`
	Fields                []TextReconciliationField `json:"fields"`
	ManualReview          []TextReconciliationIssue `json:"manual_review"`
	Counts                map[string]int            `json:"counts"`
	NativeQualification   string                    `json:"native_qualification"`
	AdoptionScope         string                    `json:"adoption_scope"`
}

// ReconcileText compares verified source/native baselines with current YAML and
// an edited deck. It proposes exact named string fields, not arbitrary reverse
// compilation. Unsupported payload stays visible even when text proposals exist.
func ReconcileText(p *Project, b *TextBaseline, edited []byte) (TextReconciliationReport, error) {
	out := TextReconciliationReport{Schema: TextReconciliationSchema, ProjectID: p.Document.ID, Fields: []TextReconciliationField{}, ManualReview: []TextReconciliationIssue{}, Counts: map[string]int{}, NativeQualification: "desktop_identity_and_editing_qualification_not_recorded", AdoptionScope: "reviewed_named_plain_text_only; geometry, formatting, topology, notes and shared definitions require separate review"}
	if b == nil {
		return out, fmt.Errorf("reconcile.verified_baseline_required")
	}
	// Public baseline metadata is an observation, not a mutable input override.
	var stored Receipt
	if e := strictInto(json.RawMessage(b.files["receipt.json"]), &stored); e != nil {
		return out, e
	}
	if !bytes.Equal(canonical(stored), canonical(b.Receipt)) || digest(b.files["receipt.json"]) != b.ReceiptSHA256 || !bytes.Equal(canonical(b.Objects), b.files["object-map.json"]) {
		return out, fmt.Errorf("reconcile.baseline_metadata_changed")
	}
	if b.Receipt.ProjectID != p.Document.ID {
		return out, fmt.Errorf("reconcile.project_identity_mismatch")
	}
	_, lock, e := ReadLock(p)
	if e != nil {
		return out, e
	}
	if digest(lock) != b.Receipt.LockSHA256 {
		return out, fmt.Errorf("reconcile.toolchain_changed: select/rebuild a baseline using the current exact lock")
	}
	native, e := InspectNativeLineage(edited, b.Objects)
	if e != nil {
		return out, e
	}
	out.BaselineBuildID = b.Receipt.BuildID
	out.BaselineReceiptSHA256 = b.ReceiptSHA256
	out.BaselinePPTXSHA256 = b.Receipt.Outputs["deck.pptx"]
	out.CurrentSourceSHA256 = p.SourceHash()
	out.CurrentSemanticSHA256 = digest(p.Canonical)
	out.EditedPPTXSHA256 = digest(edited)
	out.LockSHA256 = digest(lock)
	// Charge each record's exact serialized bytes as it is collected. Refuse
	// oversized context rather than truncate baseline/copy or publish a partial
	// report. Reserve room for top-level metadata below the packet's 64 MiB cap.
	reportBytes, reportTooLarge := 0, false
	charge := func(record any) bool {
		reportBytes += len(canonical(record)) + 1
		if reportBytes > 60<<20 {
			reportTooLarge = true
			return false
		}
		return true
	}
	addIssue := func(kind, token, slide, part, detail string) {
		issue := TextReconciliationIssue{kind, token, slide, part, detail}
		if charge(issue) {
			out.ManualReview = append(out.ManualReview, issue)
		}
	}
	blockedShapes, blockedSlides := map[string]bool{}, map[string]bool{}
	for _, i := range native.Issues {
		addIssue(i.Kind, i.ShapeToken, b.Objects.Lineage.Slides[i.SlideToken], i.NativePart, i.Detail)
		if i.ShapeToken != "" {
			blockedShapes[i.ShapeToken] = true
		}
		if strings.HasPrefix(i.Kind, "slide_") && i.SlideToken != "" {
			blockedSlides[i.SlideToken] = true
		}
	}
	baselineObjects := map[string]NativeLineageObject{}
	for _, o := range b.inspection.Objects {
		baselineObjects[o.ShapeToken] = o
	}
	editedObjects := map[string]NativeLineageObject{}
	for _, o := range native.Objects {
		if !blockedShapes[o.ShapeToken] && !blockedSlides[o.SlideToken] {
			editedObjects[o.ShapeToken] = o
		}
	}
	sourceUses := map[string]int{}
	for _, object := range b.Objects.Objects {
		for _, field := range object.Fields {
			if field.Status == "plain_text_baseline" {
				sourceUses[object.SlideID+"\x00"+field.SourceSlot]++
			}
		}
	}
	for _, object := range b.Objects.Objects {
		original := baselineObjects[object.ShapeToken]
		current, matched := editedObjects[object.ShapeToken]
		if !matched || current.shape == nil {
			continue
		}
		paragraphs := nativeParagraphs(current.shape)
		text := nativeParagraphText(paragraphs)
		if reconcileStructureHash(original.shape) != reconcileStructureHash(current.shape) {
			addIssue("native_structure_or_format_changed", object.ShapeToken, object.SlideID, current.NativePart, "nontext shape payload changed; text adoption does not adopt geometry, style or topology")
		}
		if len(object.Fields) == 0 {
			if text != object.NativeText {
				addIssue("unmapped_native_text_changed", object.ShapeToken, object.SlideID, current.NativePart, "native text has no unique maintained source field")
			}
			continue
		}
		for _, field := range object.Fields {
			if field.Status != "plain_text_baseline" && text == object.NativeText {
				continue
			}
			entry := TextReconciliationField{FieldIdentity: field.Identity, ShapeToken: object.ShapeToken, SlideID: object.SlideID, LogicalID: object.LogicalID, SourceSlot: field.SourceSlot, Baseline: object.NativeText, EditedNative: text, Status: "manual_review"}
			pointer, yamlText, resolveErr := ResolveNativeSourceField(p, object, field)
			entry.SourcePointer = pointer
			entry.CurrentYAML = yamlText
			switch {
			case field.Status != "plain_text_baseline":
				entry.Reason = field.Reason
			case len(object.Fields) != 1 || sourceUses[object.SlideID+"\x00"+field.SourceSlot] != 1:
				entry.Reason = "One maintained source field maps to multiple native fields or objects; correspondence requires review."
			case resolveErr != nil:
				entry.Reason = resolveErr.Error()
			case !supportedEditedPlainText(current.Kind, paragraphs):
				entry.Reason = "Native field includes unsupported rich text, bullets, dynamic fields, cells or missing text body."
			case !utf8.ValidString(text) || len(text) > 64<<10:
				entry.Reason = "Edited field exceeds the 64 KiB UTF-8 plain-text proposal bound."
			default:
				switch {
				case yamlText == entry.Baseline && text == entry.Baseline:
					entry.Status = "no_op"
				case text == entry.Baseline:
					entry.Status = "yaml_only"
				case yamlText == text:
					entry.Status = "matching_changes"
				case yamlText == entry.Baseline:
					entry.Status = "native_only"
				default:
					entry.Status = "conflict"
				}
			}
			entry.ID = digest(canonical([]string{TextReconciliationSchema, b.Objects.Lineage.BuildToken, entry.SlideID, entry.SourceSlot, entry.FieldIdentity, entry.ShapeToken, entry.Baseline, entry.CurrentYAML, entry.EditedNative, entry.Status, entry.Reason}))
			if entry.Status == "manual_review" {
				addIssue("source_field_requires_review", object.ShapeToken, object.SlideID, current.NativePart, entry.Reason)
			}
			if !charge(entry) {
				return out, fmt.Errorf("reconcile.report_context_too_large")
			}
			out.Fields = append(out.Fields, entry)
		}
	}
	if e = reconcilePackageChanges(b.files["deck.pptx"], edited, b.inspection, native, addIssue); e != nil {
		return out, e
	}
	if reportTooLarge {
		return out, fmt.Errorf("reconcile.report_context_too_large")
	}
	sort.Slice(out.Fields, func(i, j int) bool {
		a, z := out.Fields[i], out.Fields[j]
		return a.SlideID+"\x00"+a.SourceSlot+"\x00"+a.ID < z.SlideID+"\x00"+z.SourceSlot+"\x00"+z.ID
	})
	sort.Slice(out.ManualReview, func(i, j int) bool {
		a, z := out.ManualReview[i], out.ManualReview[j]
		return a.Kind+"\x00"+a.SlideID+"\x00"+a.ShapeToken+"\x00"+a.NativePart < z.Kind+"\x00"+z.SlideID+"\x00"+z.ShapeToken+"\x00"+z.NativePart
	})
	for _, f := range out.Fields {
		out.Counts[f.Status]++
	}
	out.Counts["manual_review_items"] = len(out.ManualReview)
	return out, nil
}

func supportedEditedPlainText(kind string, paragraphs []NativeParagraph) bool {
	if kind != "sp" || len(paragraphs) == 0 {
		return false
	}
	for _, p := range paragraphs {
		if p.Address.TableRow != nil || p.Address.TableColumn != nil || p.Address.Body != 0 || len(p.ReviewItems) != 0 {
			return false
		}
		style := ""
		for _, r := range p.Runs {
			if r.Kind != "r" && r.Kind != "br" {
				return false
			}
			if style == "" {
				style = r.PropertiesSHA256
			} else if style != r.PropertiesSHA256 {
				return false
			}
		}
	}
	return true
}

// Normalize only lineage metadata and physical diagnostics. Text leaves are
// removed for the structural comparison; formatting and geometry stay intact.
func reconcileStructureHash(shape *xmlNode) string {
	var clone func(*xmlNode) *xmlNode
	clone = func(n *xmlNode) *xmlNode {
		c := &xmlNode{Name: n.Name, Text: n.Text, Attrs: nil, Children: nil}
		if n.Name.Space == drawingML && n.Name.Local == "t" {
			c.Text = ""
		}
		for _, a := range n.Attrs {
			if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
				continue
			}
			if n.Name.Space == lineagePML && n.Name.Local == "cNvPr" && a.Name.Space == "" && (a.Name.Local == "id" || a.Name.Local == "name") {
				continue
			}
			c.Attrs = append(c.Attrs, a)
		}
		sort.Slice(c.Attrs, func(i, j int) bool {
			a, b := c.Attrs[i].Name, c.Attrs[j].Name
			return a.Space+"\x00"+a.Local < b.Space+"\x00"+b.Local
		})
		for _, child := range n.Children {
			if child.Name.Space == lineagePML && child.Name.Local == "custDataLst" {
				continue
			}
			c.Children = append(c.Children, clone(child))
		}
		return c
	}
	return digest(canonical(clone(shape)))
}

func reconcilePackageChanges(before, after []byte, baseline, edited NativeLineageInspection, add func(string, string, string, string, string)) error {
	a, e := openLineagePackage(before)
	if e != nil {
		return e
	}
	b, e := openLineagePackage(after)
	if e != nil {
		return e
	}
	exclude := map[string]bool{}
	for _, view := range []NativeLineageInspection{baseline, edited} {
		for _, o := range view.Objects {
			exclude[o.NativePart] = true
		}
	}
	// Tag parts and owner relationships are physical lineage addresses. Other
	// payloads (masters, media, charts, notes, settings, new/orphan parts) are
	// explicitly reported so supported text cannot hide unsupported native work.
	for _, p := range []*lineagePackage{a, b} {
		owners := map[string]bool{}
		rootRels, e := p.relationships("")
		if e != nil {
			return e
		}
		for _, r := range rootRels {
			if r.Type == lineageRML+"/officeDocument" {
				owner, e := lineageTarget("", r.Target)
				if e != nil {
					return e
				}
				owners[owner] = true
			}
		}
		for _, view := range []NativeLineageInspection{baseline, edited} {
			for _, o := range view.Objects {
				owners[o.NativePart] = true
			}
		}
		for owner := range owners {
			if p.files[owner] == nil {
				continue
			}
			rp := lineageRelPart(owner)
			if p.files[rp] == nil {
				continue
			}
			rels, e := p.relationships(owner)
			if e != nil {
				return e
			}
			for _, r := range rels {
				if r.Type == lineageTagType && (r.Mode == "" || r.Mode == "Internal") {
					target, e := lineageTarget(owner, r.Target)
					if e != nil {
						return e
					}
					exclude[target] = true
				}
			}
		}
	}
	names := map[string]bool{}
	for n := range a.files {
		names[n] = true
	}
	for n := range b.files {
		names[n] = true
	}
	keys := []string{}
	for n := range names {
		keys = append(keys, n)
	}
	sort.Strings(keys)
	for _, n := range keys {
		if exclude[n] || a.files[n] != nil && a.files[n].FileInfo().IsDir() || b.files[n] != nil && b.files[n].FileInfo().IsDir() {
			continue
		}
		kind := "native_package_part_changed"
		if a.files[n] == nil {
			kind = "native_package_part_added"
		} else if b.files[n] == nil {
			kind = "native_package_part_missing"
		}
		var x, y []byte
		if a.files[n] != nil {
			x, e = a.read(n)
			if e != nil {
				return e
			}
		}
		if b.files[n] != nil {
			y, e = b.read(n)
			if e != nil {
				return e
			}
		}
		if !bytes.Equal(x, y) {
			add(kind, "", "", n, "payload outside named source text changed; preserve and review the edited deck")
		}
	}
	// Changes inside slide XML but outside its named shapes (background,
	// transitions, animation or root group settings) need their own comparison.
	bySlide := func(v NativeLineageInspection) map[string]string {
		m := map[string]string{}
		for _, o := range v.Objects {
			m[o.SlideToken] = o.NativePart
		}
		return m
	}
	originals, modified := bySlide(baseline), bySlide(edited)
	for token, part := range originals {
		newPart := modified[token]
		if newPart == "" {
			continue
		}
		x, e := a.tree(part)
		if e != nil {
			return e
		}
		y, e := b.tree(newPart)
		if e != nil {
			return e
		}
		if slideSettingsHash(x) != slideSettingsHash(y) {
			add("native_slide_settings_changed", "", "", newPart, "slide payload outside named shapes changed; requires manual review")
		}
	}
	return nil
}
func slideSettingsHash(root *xmlNode) string {
	var clone func(*xmlNode) *xmlNode
	clone = func(n *xmlNode) *xmlNode {
		c := *n
		c.Children = nil
		for _, child := range n.Children {
			if child.Name.Space == lineagePML && (child.Name.Local == "sp" || child.Name.Local == "pic" || child.Name.Local == "graphicFrame" || child.Name.Local == "grpSp" || child.Name.Local == "cxnSp" || child.Name.Local == "custDataLst") {
				continue
			}
			c.Children = append(c.Children, clone(child))
		}
		return &c
	}
	return reconcileStructureHash(clone(root))
}
