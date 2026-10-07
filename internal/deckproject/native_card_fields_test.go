package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func editableCardFixture(t *testing.T, duplicate bool) (*Project, *TextBaseline) {
	return editableCardFixtureOptions(t, duplicate, false)
}

func editableCardFixtureOptions(t *testing.T, duplicate, repeated bool) (*Project, *TextBaseline) {
	t.Helper()
	p := example(t)
	zones := map[string]Zone{}
	for _, name := range []string{"title", "card_title", "card_body"} {
		role := "node-copy"
		if name == "title" {
			role = "slide-title"
		}
		zones[name] = Zone{Role: role, Required: true, Schema: map[string]any{"type": "string", "maxLength": 200}}
	}
	bodyBinding := "card_body"
	if duplicate {
		bodyBinding = "card_title"
		delete(zones, "card_body")
	}
	p.Document.LocalTemplates["editable-card"] = LocalTemplate{Name: "Synthetic editable card", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: zones, Nodes: []Node{{ID: "unit", Kind: "component", Placement: &Placement{Zone: "body", Span: &Span{Start: 1, Count: 6, Y: 36, H: 200}}, Definition: &Reference{Scope: "shared", ID: "wmds/component/editable-card"}, Arguments: map[string]any{"surface": "subtle", "title": map[string]any{"binding": "card_title"}, "body": map[string]any{"binding": bodyBinding}}}}}
	if repeated {
		local := p.Document.LocalTemplates["editable-card"]
		second := local.Nodes[0]
		second.ID = "unit-copy"
		second.Placement = &Placement{Zone: "body", Span: &Span{Start: 7, Count: 6, Y: 36, H: 200}}
		local.Nodes = append(local.Nodes, second)
		p.Document.LocalTemplates["editable-card"] = local
	}
	p.Document.Slides = []Slide{{ID: "native-card", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "editable-card"}, Values: map[string]any{"title": "Synthetic card editing", "card_title": "Editable title", "card_body": "Editable body copy"}}}
	if duplicate {
		delete(p.Document.Slides[0].Values, "card_body")
	}
	tree, e := editYAMLNode(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := encodeSourceYAML(tree)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	return p, b
}

func editableCardReportField(t *testing.T, r TextReconciliationReport, slot string) TextReconciliationField {
	t.Helper()
	for _, f := range r.Fields {
		if f.SourceSlot == slot {
			return f
		}
	}
	t.Fatal("field missing", slot, r)
	return TextReconciliationField{}
}

func editCardCopy(t *testing.T, b *TextBaseline, title, body string) []byte {
	t.Helper()
	return lineageEdit(t, b.files["deck.pptx"], "ppt/slides/slide1.xml", func(raw []byte) []byte {
		raw = bytes.Replace(raw, []byte("Editable title"), []byte(title), 1)
		return bytes.Replace(raw, []byte("Editable body copy"), []byte(body), 1)
	})
}

func TestNativeEditabilityEditableCardExactParagraphFieldsAndReconciliation(t *testing.T) {
	p, b := editableCardFixture(t, false)
	var unit ObjectRecord
	for _, o := range b.Objects.Objects {
		if o.NativeName == "unit" {
			unit = o
		}
	}
	if unit.NativeKind != "sp" || unit.NativeParentToken != "" || len(unit.Fields) != 2 || unit.TextMappingContract != wmdesign.EditableCardContract {
		t.Fatal("not a single two-field unit", unit)
	}
	for i, f := range unit.Fields {
		if len(f.Addresses) != 1 || f.Addresses[0].Paragraph != i || f.Status != "plain_text_baseline" {
			t.Fatal(f)
		}
	}
	inv, e := NativeEditability(b)
	if e != nil {
		t.Fatal(e)
	}
	if inv.Counts["text_and_paint_in_same_shape"] != 1 || inv.Counts["maximum_group_depth"] != 0 {
		t.Fatal(inv.Counts)
	}
	edited := editCardCopy(t, b, "Native title", "Native body")
	r, e := ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counts["native_only"] != 2 || len(r.ManualReview) != 0 {
		t.Fatal(r)
	}
	if f := editableCardReportField(t, r, "card_title"); f.Baseline != "Editable title" || f.EditedNative != "Native title" {
		t.Fatal(f)
	}
	if f := editableCardReportField(t, r, "card_body"); f.Baseline != "Editable body copy" || f.EditedNative != "Native body" {
		t.Fatal(f)
	}
	p = rewrite(t, p, "card_title: Editable title", "card_title: YAML title")
	r, e = ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if editableCardReportField(t, r, "card_title").Status != "conflict" || editableCardReportField(t, r, "card_body").Status != "native_only" {
		t.Fatal(r)
	}
}

func TestReconcileEditableCardDuplicateBindingIsManual(t *testing.T) {
	p, b := editableCardFixture(t, true)
	for _, o := range b.Objects.Objects {
		if o.NativeName == "unit" && (len(o.Fields) != 0 || o.TextMapping != "manual_review") {
			t.Fatal(o)
		}
	}
	edited := lineageEdit(t, b.files["deck.pptx"], "ppt/slides/slide1.xml", func(raw []byte) []byte {
		return bytes.Replace(raw, []byte("Editable title"), []byte("No inferred role"), 1)
	})
	r, e := ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counts["native_only"] != 0 || !hasReconcileIssue(r, "unmapped_native_text_changed") {
		t.Fatal(r)
	}
}

func TestReconcileEditableCardRepeatedObjectBindingIsManual(t *testing.T) {
	p, b := editableCardFixtureOptions(t, false, true)
	edited := editCardCopy(t, b, "Reviewed title", "Reviewed body")
	r, e := ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counts["native_only"] != 0 {
		t.Fatal("one source slot controls multiple native fields", r)
	}
	count := 0
	for _, field := range r.Fields {
		if field.SourceSlot == "card_title" || field.SourceSlot == "card_body" {
			count++
			if field.Status != "manual_review" || !strings.Contains(field.Reason, "multiple native") {
				t.Fatal(field)
			}
		}
	}
	if count != 4 {
		t.Fatal("repeated card fields missing", r)
	}
}

func TestReconcileEditableCardFormatTopologyAndGeometryStayExplicit(t *testing.T) {
	p, b := editableCardFixture(t, false)
	edited := editCardCopy(t, b, "Native title", "Native body")
	for _, mutation := range []string{"format", "paragraph", "geometry"} {
		altered := lineageEdit(t, edited, "ppt/slides/slide1.xml", func(raw []byte) []byte {
			// Locate the owned shape only; leave frame metadata and other text untouched.
			start := bytes.Index(raw, []byte(`name="unit"`))
			if start < 0 {
				t.Fatal("unit missing")
			}
			end := start + bytes.Index(raw[start:], []byte("</p:sp>"))
			segment := append([]byte{}, raw[start:end]...)
			switch mutation {
			case "format":
				segment = bytes.Replace(segment, []byte(`b="0"`), []byte(`b="1"`), 1)
			case "paragraph":
				segment = bytes.Replace(segment, []byte("</p:txBody>"), []byte("<a:p/></p:txBody>"), 1)
			case "geometry":
				segment = bytes.Replace(segment, []byte(`<a:off x="`), []byte(`<a:off x="9`), 1)
			}
			if bytes.Equal(segment, raw[start:end]) {
				t.Fatal("mutation did not change fixture", mutation)
			}
			return append(append(append([]byte{}, raw[:start]...), segment...), raw[end:]...)
		})
		r, e := ReconcileText(p, b, altered)
		if e != nil {
			t.Fatal(e)
		}
		if !hasReconcileIssue(r, "native_structure_or_format_changed") {
			t.Fatal("structure concealed", mutation, r)
		}
		if mutation == "geometry" {
			if r.Counts["native_only"] != 2 {
				t.Fatal("geometry concealed supported copy", r)
			}
		} else {
			if r.Counts["native_only"] != 0 || editableCardReportField(t, r, "card_title").Status != "manual_review" {
				t.Fatal("unsupported role structure adopted", mutation, r)
			}
		}
	}
	// Missing or overlapping addresses must not be resolved by equal copy.
	invalid := unitForTest(b)
	invalid.Fields[0].Addresses = []NativeTextAddress{{Paragraph: 1}}
	if editableCardFieldContractValid(invalid) {
		t.Fatal("wrong role address accepted")
	}
	if _, ok := nativeFieldText(NativeSourceField{Addresses: []NativeTextAddress{{Paragraph: 9}}}, invalid.Paragraphs); ok {
		t.Fatal("missing paragraph inferred")
	}
}

func unitForTest(b *TextBaseline) ObjectRecord {
	for _, o := range b.Objects.Objects {
		if o.NativeName == "unit" {
			return o
		}
	}
	return ObjectRecord{}
}

func TestReconcileEditableCardUnsupportedEditedCopyIsManual(t *testing.T) {
	p, b := editableCardFixture(t, false)
	for _, copy := range []string{"[[Bold]]", "New\nparagraph"} {
		edited := editCardCopy(t, b, copy, "Native body")
		r, e := ReconcileText(p, b, edited)
		if e != nil {
			t.Fatal(e)
		}
		if r.Counts["native_only"] != 0 || !strings.Contains(editableCardReportField(t, r, "card_title").Reason, "contract changed") {
			t.Fatal(r)
		}
	}
}

func TestReconcileEditableCardEachRoleAdoptRebuildAndReplay(t *testing.T) {
	for _, role := range []string{"title", "body"} {
		t.Run(role, func(t *testing.T) {
			p, b := editableCardFixture(t, false)
			title, body := "Editable title", "Editable body copy"
			if role == "title" {
				title = "Reviewed title"
			} else {
				body = "Reviewed body"
			}
			edited := editCardCopy(t, b, title, body)
			packet, e := WriteTextReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"))
			if e != nil {
				t.Fatal(e)
			}
			if packet.Report.Counts["native_only"] != 1 {
				t.Fatal(packet.Report.Counts)
			}
			decisions := reconciliationDecisions(t, packet, "use_native")
			receipt, e := AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if len(receipt.Changed) != 1 || receipt.Changed[0].SourceSlot != "card_"+role {
				t.Fatal(receipt)
			}
			after, e := Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			if after.Document.Slides[0].Values["card_title"] != title || after.Document.Slides[0].Values["card_body"] != body {
				t.Fatal(after.Document.Slides[0].Values)
			}
			again, e := AdoptTextReviewPacket(after, packet, decisions, bundle(t), wmdesign.CandidateEngine)
			if e != nil || !bytes.Equal(canonical(receipt), canonical(again)) {
				t.Fatal("non-idempotent adoption", e)
			}
			built, e := Build(after, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
			if e != nil {
				t.Fatal(e)
			}
			rebuilt, e := ReadTextBaseline(after, built.BuildID, "")
			if e != nil {
				t.Fatal(e)
			}
			unit := unitForTest(rebuilt)
			if unit.NativeText != title+"\n"+body || len(unit.Fields) != 2 || unit.NativeKind != "sp" {
				t.Fatal(unit)
			}
			if _, e = ReadTextBaseline(after, b.Receipt.BuildID, b.ReceiptSHA256); e != nil {
				t.Fatal("historical baseline lost", e)
			}
			r, e := ReconcileText(after, rebuilt, rebuilt.files["deck.pptx"])
			if e != nil || r.Counts["native_only"] != 0 || r.Counts["conflict"] != 0 {
				t.Fatal(r, e)
			}
		})
	}
}

func TestNativeEditabilityEditableCardActualParagraphsAndLegacy(t *testing.T) {
	_, b := editableCardFixture(t, false)
	unit := unitForTest(b)
	var shape *xmlNode
	for _, object := range b.inspection.Objects {
		if object.ShapeToken == unit.ShapeToken {
			shape = object.shape
		}
	}
	actual := nativeParagraphs(shape)
	if !editableCardBaselineParagraphsValid(unit, actual) {
		t.Fatal("actual card rejected")
	}
	for _, kind := range []string{"address", "properties", "run", "text"} {
		t.Run(kind, func(t *testing.T) {
			var changed ObjectRecord
			if e := strictInto(unit, &changed); e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "address":
				changed.Paragraphs[0].Address.Paragraph = 1
			case "properties":
				changed.Paragraphs[0].PropertiesSHA256 = strings.Repeat("0", 64)
			case "run":
				changed.Paragraphs[0].Runs[0].PropertiesSHA256 = strings.Repeat("0", 64)
			case "text":
				changed.Paragraphs[0].Text = "Invented baseline"
			}
			if editableCardBaselineParagraphsValid(changed, actual) {
				t.Fatal("metadata did not match actual native paragraphs", kind)
			}
		})
	}
	legacy := ObjectRecord{LogicalID: "legacy"}
	field := NativeSourceField{Identity: "legacy/text", Addresses: []NativeTextAddress{{Body: 0, Paragraph: 0}}}
	paragraphs := []NativeParagraph{{Address: NativeTextAddress{Body: 0, Paragraph: 0}, Text: "Original"}, {Address: NativeTextAddress{Body: 0, Paragraph: 1}, Text: "Added paragraph"}}
	text, ok := nativeObjectFieldText(legacy, field, paragraphs)
	if !ok || text != "Original\nAdded paragraph" {
		t.Fatal("legacy whole-object behavior weakened", text, ok)
	}
	tx := directXML(shape, lineagePML, "txBody")
	props := directXML(tx, drawingML, "bodyPr")
	for _, key := range []string{"lIns", "rIns", "tIns", "bIns"} {
		if attr(props, key) != "152400" {
			t.Fatal("native card inset differs from twelve points", key, attr(props, key))
		}
	}
	if actual[0].Runs[0].PropertiesSHA256 == actual[1].Runs[0].PropertiesSHA256 {
		t.Fatal("native title/body role formatting collapsed")
	}
}
