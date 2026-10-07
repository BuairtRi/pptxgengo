package deckproject

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func reconcileFixture(t *testing.T) (*Project, *TextBaseline) {
	t.Helper()
	p := example(t)
	pin(t, p)
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	return p, b
}
func reconcileEditFields(t *testing.T, b *TextBaseline, values map[string]string) []byte {
	t.Helper()
	pkg, e := openLineagePackage(b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	changes := map[string][]byte{}
	byPart := map[string][]ObjectRecord{}
	requested := map[string]bool{}
	for _, o := range b.Objects.Objects {
		if _, ok := values[o.NativeName]; ok {
			byPart[o.NativePart] = append(byPart[o.NativePart], o)
			requested[o.NativeName] = true
		}
	}
	if len(requested) != len(values) {
		t.Fatal("requested an absent native fixture field")
	}
	for part, objects := range byPart {
		raw, e := pkg.read(part)
		if e != nil {
			t.Fatal(e)
		}
		root, e := lineageSpans(raw)
		if e != nil {
			t.Fatal(e)
		}
		patches := []lineagePatch{}
		found := map[string]bool{}
		var walk func(*lineageSpan)
		walk = func(s *lineageSpan) {
			if s.node.Name == (xml.Name{Space: lineagePML, Local: "sp"}) {
				var name string
				for _, nv := range s.children {
					if nv.node.Name.Local == "nvSpPr" {
						for _, id := range nv.children {
							if id.node.Name.Local == "cNvPr" {
								name = lineageAttr(id.node, "", "name")
							}
						}
					}
				}
				for _, o := range objects {
					if name != o.NativeName {
						continue
					}
					var leaves []*lineageSpan
					var texts func(*lineageSpan)
					texts = func(n *lineageSpan) {
						if n.node.Name == (xml.Name{Space: drawingML, Local: "t"}) {
							leaves = append(leaves, n)
						}
						for _, c := range n.children {
							texts(c)
						}
					}
					texts(s)
					if len(leaves) != 1 {
						t.Fatalf("fixture field %s has %d native leaves", name, len(leaves))
					}
					var escaped bytes.Buffer
					if e := xml.EscapeText(&escaped, []byte(values[name])); e != nil {
						t.Fatal(e)
					}
					patches = append(patches, lineagePatch{leaves[0].startEnd, leaves[0].endStart, escaped.String()})
					found[name] = true
				}
			}
			for _, c := range s.children {
				walk(c)
			}
		}
		walk(root)
		if len(found) != len(objects) {
			t.Fatal("field not found")
		}
		changes[part], e = lineageApply(raw, patches)
		if e != nil {
			t.Fatal(e)
		}
	}
	changed, e := lineageRewrite(pkg, changes)
	if e != nil {
		t.Fatal(e)
	}
	return changed
}
func reportField(t *testing.T, r TextReconciliationReport, slot string) TextReconciliationField {
	t.Helper()
	for _, f := range r.Fields {
		if f.SlideID == "maintain-the-source" && f.SourceSlot == slot {
			return f
		}
	}
	t.Fatalf("missing field %s: %+v", slot, r.Fields)
	return TextReconciliationField{}
}
func hasReconcileIssue(r TextReconciliationReport, kind string) bool {
	for _, i := range r.ManualReview {
		if i.Kind == kind {
			return true
		}
	}
	return false
}

func TestReconcileTextThreeFieldsAndParallelChanges(t *testing.T) {
	p, b := reconcileFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "Native source title", "cards.items.source.body.copy": "Native source copy", "cards.items.build.title": "Native build title"})
	r, e := ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counts["native_only"] != 3 {
		t.Fatalf("expected three exact changes: %+v", r.Counts)
	}
	f := reportField(t, r, "cards.source.title")
	if f.CurrentYAML != f.Baseline || f.EditedNative != "Native source title" || f.Status != "native_only" || f.SourcePointer != "/slides/0/values/cards/0/title" {
		t.Fatal(f)
	}
	// Different current YAML stays explicit alongside both baseline and native copy.
	p = rewrite(t, p, "title: Author", "title: YAML source title")
	r, e = ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	f = reportField(t, r, "cards.source.title")
	if f.Status != "conflict" || f.CurrentYAML != "YAML source title" || f.EditedNative != "Native source title" || f.Baseline != "Author" {
		t.Fatal(f)
	}
	p = rewrite(t, p, "YAML source title", "Native source title")
	r, e = ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if reportField(t, r, "cards.source.title").Status != "matching_changes" {
		t.Fatal("matching changes not recognized")
	}
}
func TestReconcileNoOpAndYAMLOnly(t *testing.T) {
	p, b := reconcileFixture(t)
	r, e := ReconcileText(p, b, b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	if len(r.ManualReview) != 0 || r.Counts["native_only"] != 0 || r.Counts["conflict"] != 0 {
		t.Fatalf("baseline is not a no-op: counts=%+v issues=%+v", r.Counts, r.ManualReview)
	}
	p = rewrite(t, p, "title: Author", "title: YAML only")
	r, e = ReconcileText(p, b, b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	if reportField(t, r, "cards.source.title").Status != "yaml_only" {
		t.Fatal("YAML-only change lost")
	}
}
func TestReconcileUnsupportedGeometryAndPackageChangesRemainVisible(t *testing.T) {
	p, b := reconcileFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "New copy"})
	part := "ppt/slides/slide1.xml"
	edited = lineageEdit(t, edited, part, func(raw []byte) []byte { return bytes.Replace(raw, []byte(`<a:off x="`), []byte(`<a:off x="9`), 1) })
	r, e := ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counts["native_only"] != 1 || !hasReconcileIssue(r, "native_structure_or_format_changed") && !hasReconcileIssue(r, "native_slide_settings_changed") {
		t.Fatalf("geometry hidden by text proposals: counts=%+v issues=%+v", r.Counts, r.ManualReview)
	}
	edited = lineageEdit(t, edited, "docProps/core.xml", func(raw []byte) []byte {
		return bytes.Replace(raw, []byte("</cp:coreProperties>"), []byte("<!-- edited metadata --></cp:coreProperties>"), 1)
	})
	r, e = ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if !hasReconcileIssue(r, "native_package_part_changed") {
		t.Fatal("changed nontext part hidden")
	}
}
func TestReconcileDuplicateShapeProducesNoGuessedProposal(t *testing.T) {
	p, b := reconcileFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "Duplicated text"})
	edited = lineageEdit(t, edited, "ppt/slides/slide1.xml", func(raw []byte) []byte {
		s, e := lineageSpans(raw)
		if e != nil {
			t.Fatal(e)
		}
		var selected *lineageSpan
		var walk func(*lineageSpan)
		walk = func(s *lineageSpan) {
			if s.node.Name.Local == "sp" {
				for _, nv := range s.children {
					if nv.node.Name.Local == "nvSpPr" {
						for _, id := range nv.children {
							if id.node.Name.Local == "cNvPr" && lineageAttr(id.node, "", "name") == "cards.items.source.title" {
								selected = s
							}
						}
					}
				}
			}
			for _, c := range s.children {
				walk(c)
			}
		}
		walk(s)
		if selected == nil {
			t.Fatal("shape missing")
		}
		v, e := lineageApply(raw, []lineagePatch{{selected.end, selected.end, string(raw[selected.start:selected.end])}})
		if e != nil {
			t.Fatal(e)
		}
		return v
	})
	r, e := ReconcileText(p, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if !hasReconcileIssue(r, "shape_duplicated") || r.Counts["native_only"] != 0 {
		t.Fatalf("duplicate guessed: %+v", r)
	}
}
func TestReconcileBaselinePinsAndHistory(t *testing.T) {
	p, b := reconcileFixture(t)
	if _, e := ReadTextBaseline(p, b.Receipt.BuildID, strings.Repeat("0", 64)); e == nil {
		t.Fatal("state pin overridden")
	}
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e := ReadTextBaseline(p, b.Receipt.BuildID, ""); e == nil {
		t.Fatal("historical baseline chosen without pin")
	}
	if _, e := ReadTextBaseline(p, b.Receipt.BuildID, b.ReceiptSHA256); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(p.Root, "builds", b.Receipt.BuildID, "object-map.json")
	if e := os.Chmod(file, 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(file, []byte("tampered"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := ReadTextBaseline(p, b.Receipt.BuildID, b.ReceiptSHA256); e == nil {
		t.Fatal("baseline drift accepted")
	}
}
func TestReconcileReviewPacketPreservesAndRejectsDrift(t *testing.T) {
	p, b := reconcileFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "Review me"})
	source := append([]byte(nil), p.Raw...)
	destination := filepath.Join(t.TempDir(), "review packet with spaces")
	packet, e := WriteTextReviewPacket(p, b, edited, destination)
	if e != nil {
		t.Fatal(e)
	}
	if packet.Report.Counts["native_only"] != 1 || !bytes.Equal(packet.Edited, edited) {
		t.Fatal("packet lost edits")
	}
	current, e := os.ReadFile(p.SourcePath)
	if e != nil || !bytes.Equal(current, source) {
		t.Fatal("proposals mutated source")
	}
	if _, e = WriteTextReviewPacket(p, b, edited, destination); e == nil {
		t.Fatal("existing packet replaced")
	}
	if e = os.WriteFile(filepath.Join(destination, "extra.txt"), []byte("unlisted"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = ReadTextReviewPacket(destination); e == nil {
		t.Fatal("unlisted packet content accepted")
	}
}

func reconciliationDecisions(t *testing.T, packet *TextReviewPacket, action string) []byte {
	t.Helper()
	d := TextReviewDecisions{Schema: TextReviewDecisionsSchema, ReportSHA256: packet.ReportSHA256, Actor: "synthetic-review-fixture", Decisions: []TextReviewDecision{}}
	for _, f := range packet.Report.Fields {
		if f.Status == "native_only" || f.Status == "conflict" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, action, "Fixture explicitly selects this named text result; no native visual approval."})
		}
	}
	return canonical(d)
}
func TestReconcileAdoptThreeFieldsRebuildAndRepeat(t *testing.T) {
	p, b := reconcileFixture(t)
	edits := map[string]string{"cards.items.source.title": "Native author", "cards.items.source.body.copy": "Native editable source.", "cards.items.build.title": "Native build"}
	edited := reconcileEditFields(t, b, edits)
	packet, e := WriteTextReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"))
	if e != nil {
		t.Fatal(e)
	}
	decisions := reconciliationDecisions(t, packet, "use_native")
	receipt, e := AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(receipt.Changed) != 3 || receipt.BeforeSourceSHA256 == receipt.AfterSourceSHA256 {
		t.Fatal("three reviewed edits not adopted")
	}
	after, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if after.Document.Slides[0].Values["cards"].([]any)[0].(map[string]any)["title"] != "Native author" {
		t.Fatal("wrong source changed")
	}
	again, e := AdoptTextReviewPacket(after, packet, decisions, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(canonical(receipt), canonical(again)) {
		t.Fatal("repeat adoption created a new mutation/receipt")
	}
	r, e := ReconcileText(after, b, edited)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counts["native_only"] != 0 || r.Counts["matching_changes"] != 3 {
		t.Fatal("repeat reconciliation proposed duplicate updates", r.Counts)
	}
	built, e := Build(after, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	rebuilt, e := ReadTextBaseline(after, built.BuildID, built.Outputs["receipt.json"])
	if e != nil {
		t.Fatal(e)
	}
	found := 0
	for _, o := range rebuilt.Objects.Objects {
		if text, ok := edits[o.NativeName]; ok {
			found++
			if o.NativeText != text {
				t.Fatalf("rebuild did not reproduce %s", o.NativeName)
			}
		}
	}
	if found != 3 {
		t.Fatal("missing rebuilt fields")
	}
	retained, e := os.ReadFile(filepath.Join(p.Root, filepath.FromSlash(receipt.RetainedEdited)))
	if e != nil || !bytes.Equal(retained, edited) {
		t.Fatal("edited deck lost")
	}
	if _, e = ReadTextBaseline(after, b.Receipt.BuildID, b.ReceiptSHA256); e != nil {
		t.Fatal("baseline/history lost", e)
	}
}
func TestReconcileConflictReviewAndStaleSource(t *testing.T) {
	p, b := reconcileFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "Native chosen copy"})
	p = rewrite(t, p, "title: Author", "title: YAML parallel copy")
	packet, e := WriteTextReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"))
	if e != nil {
		t.Fatal(e)
	}
	if reportField(t, packet.Report, "cards.source.title").Status != "conflict" {
		t.Fatal("conflict missing")
	}
	receipt, e := AdoptTextReviewPacket(p, packet, reconciliationDecisions(t, packet, "use_native"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(receipt.Changed) != 1 || receipt.Changed[0].Before != "YAML parallel copy" {
		t.Fatal("parallel copy lost from receipt")
	}
	preimage, e := os.ReadFile(filepath.Join(p.Root, "decisions", "sources", p.SourceHash(), filepath.Base(p.SourcePath)))
	if e != nil || !bytes.Contains(preimage, []byte("YAML parallel copy")) {
		t.Fatal("source predecessor lost")
	}
	p, b = reconcileFixture(t)
	packet, e = WriteTextReviewPacket(p, b, reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "Review first"}), filepath.Join(t.TempDir(), "review"))
	if e != nil {
		t.Fatal(e)
	}
	p = rewrite(t, p, "title: Build", "title: Parallel unreviewed build")
	before := append([]byte(nil), p.Raw...)
	if _, e = AdoptTextReviewPacket(p, packet, reconciliationDecisions(t, packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("stale review overwrote new source")
	}
	raw, e := os.ReadFile(p.SourcePath)
	if e != nil || !bytes.Equal(raw, before) {
		t.Fatal("stale refusal changed source")
	}
}
func TestReconcileKeepYAMLRetainsNativeAndNoDuplicateWrites(t *testing.T) {
	p, b := reconcileFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "Declined native copy"})
	packet, e := WriteTextReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"))
	if e != nil {
		t.Fatal(e)
	}
	decisions := reconciliationDecisions(t, packet, "keep_yaml")
	r, e := AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Changed) != 0 || r.BeforeSourceSHA256 != r.AfterSourceSHA256 {
		t.Fatal("keep_yaml mutated source")
	}
	r2, e := AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine)
	if e != nil || !bytes.Equal(canonical(r), canonical(r2)) {
		t.Fatal("keep_yaml repeat duplicated decision", e)
	}
}
func TestReconcileDecisionsRejectDuplicateAndWrongIdentity(t *testing.T) {
	p, b := reconcileFixture(t)
	packet, e := WriteTextReviewPacket(p, b, reconcileEditFields(t, b, map[string]string{"cards.items.source.title": "Copy"}), filepath.Join(t.TempDir(), "review"))
	if e != nil {
		t.Fatal(e)
	}
	d, e := DecodeTextReviewDecisions(reconciliationDecisions(t, packet, "use_native"), "test")
	if e != nil {
		t.Fatal(e)
	}
	d.Decisions = append(d.Decisions, d.Decisions[0])
	if _, e = DecodeTextReviewDecisions(canonical(d), "test"); e == nil {
		t.Fatal("duplicate decision accepted")
	}
	d.Decisions = d.Decisions[:1]
	d.ReportSHA256 = strings.Repeat("0", 64)
	if _, e = AdoptTextReviewPacket(p, packet, canonical(d), bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("decision for wrong report accepted")
	}
	if _, e = DecodeTextReviewDecisions([]byte(`{"schema":"one","schema":"two"}`), "duplicate"); e == nil {
		t.Fatal("duplicate JSON key accepted")
	}
}
