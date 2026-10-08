package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func removeGeometryObjects(t *testing.T, raw []byte, names ...string) []byte {
	t.Helper()
	return lineageEdit(t, raw, "ppt/slides/slide1.xml", func(data []byte) []byte {
		inv, _, e := geometryInventory(data)
		if e != nil {
			t.Fatal(e)
		}
		patches := []lineagePatch{}
		for _, name := range names {
			o := inv[name]
			if o == nil {
				t.Fatalf("missing %s", name)
			}
			patches = append(patches, lineagePatch{o.span.start, o.span.end, ""})
		}
		data, e = lineageApply(data, patches)
		if e != nil {
			t.Fatal(e)
		}
		return data
	})
}
func structurePacket(t *testing.T, p *Project, b *TextBaseline, edited []byte) *TextReviewPacket {
	t.Helper()
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	return packet
}
func structureDecisions(packet *TextReviewPacket, action string) []byte {
	d := TextReviewDecisions{Schema: TextReviewDecisionsSchema, ReportSHA256: packet.ReportSHA256, Actor: "structure review fixture"}
	for _, f := range packet.Report.Structure {
		if f.Status == "native_only" || f.Status == "conflict" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, action, "Reviewed complete native component deletion"})
		}
	}
	return canonical(d)
}
func TestStructureWholeLocalComponentDeletionRebuild(t *testing.T) {
	p, b := geometryFixture(t)
	edited := removeGeometryObjects(t, b.files["deck.pptx"], "node08")
	before := p.SourceHash()
	packet := structurePacket(t, p, b, edited)
	if len(packet.Report.Structure) != 1 || packet.Report.Structure[0].Status != "native_only" || len(packet.Report.ManualReview) != 0 {
		t.Fatalf("report %+v", packet.Report)
	}
	if p.SourceHash() != before {
		t.Fatal("proposal wrote source")
	}
	receipt, e := AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(receipt.Changed) != 1 || len(receipt.RemainingFieldIDs) != 0 {
		t.Fatalf("receipt %+v", receipt)
	}
	after, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if topDiagramNode(after.Document.LocalTemplates["architecture"], "node08") != nil {
		t.Fatal("node survived")
	}
	if _, ok := after.Document.Slides[0].Values["node08.text"]; ok {
		t.Fatal("unused copy survived")
	}
	if _, e = AdoptTextReviewPacket(after, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	built, e := Build(after, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	rebuilt, e := os.ReadFile(filepath.Join(after.Root, "builds", built.BuildID, "deck.pptx"))
	if e != nil {
		t.Fatal(e)
	}
	a, ao := geometryFromPackage(t, edited)
	z, zo := geometryFromPackage(t, rebuilt)
	if len(a) != len(z) || !bytes.Equal(canonical(ao), canonical(zo)) {
		t.Fatal("rebuilt topology/order differs")
	}
	for name, o := range a {
		if z[name] == nil || !bytes.Equal(canonical(o.geometry), canonical(z[name].geometry)) {
			t.Fatalf("rebuilt geometry differs: %s", name)
		}
	}
}
func TestStructurePartialDeletionAndLostTagsStayManual(t *testing.T) {
	for _, mode := range []string{"partial", "lost-tags", "duplicate"} {
		t.Run(mode, func(t *testing.T) {
			p, b := geometryFixture(t)
			edited := b.files["deck.pptx"]
			switch mode {
			case "partial":
				edited = removeGeometryObjects(t, edited, "node08.text")
			case "lost-tags":
				edited = lineageEdit(t, edited, "ppt/slides/slide1.xml", func(data []byte) []byte {
					inv, _, e := geometryInventory(data)
					if e != nil {
						t.Fatal(e)
					}
					patches := []lineagePatch{}
					var walk func(*lineageSpan)
					walk = func(s *lineageSpan) {
						if s.node.Name.Local == "custDataLst" {
							patches = append(patches, lineagePatch{s.start, s.end, ""})
							return
						}
						for _, c := range s.children {
							walk(c)
						}
					}
					walk(inv["node08"].span)
					data, e = lineageApply(data, patches)
					if e != nil {
						t.Fatal(e)
					}
					return data
				})
			case "duplicate":
				edited = lineageEdit(t, edited, "ppt/slides/slide1.xml", func(data []byte) []byte {
					inv, _, e := geometryInventory(data)
					if e != nil {
						t.Fatal(e)
					}
					o := inv["node08"]
					data, e = lineageApply(data, []lineagePatch{{o.span.end, o.span.end, string(data[o.span.start:o.span.end])}})
					if e != nil {
						t.Fatal(e)
					}
					return data
				})
			}
			packet := structurePacket(t, p, b, edited)
			for _, f := range packet.Report.Structure {
				if f.Status != "manual_review" {
					t.Fatalf("unsafe proposal: %+v", f)
				}
			}
			if len(packet.Report.ManualReview) == 0 {
				t.Fatal("unsupported edit ignored")
			}
		})
	}
}
func TestStructureKeepYAMLConflictAndDrift(t *testing.T) {
	p, b := geometryFixture(t)
	edited := removeGeometryObjects(t, b.files["deck.pptx"], "node08")
	changed := p.Document.LocalTemplates["architecture"]
	changed.Nodes[7].Placement.Rect.Y += 6
	p.Document.LocalTemplates["architecture"] = changed
	raw := canonical(p.Document)
	if e := os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	packet := structurePacket(t, p, b, edited)
	if len(packet.Report.Structure) != 1 || packet.Report.Structure[0].Status != "conflict" {
		t.Fatalf("report %+v", packet.Report.Structure)
	}
	before := p.SourceHash()
	receipt, e := AdoptTextReviewPacket(p, packet, structureDecisions(packet, "keep_yaml"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(receipt.Changed) != 0 {
		t.Fatal("keep_yaml changed source")
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if p.SourceHash() != before {
		t.Fatal("keep_yaml rewrote source")
	}
	// A new proposal's current tree must still be present when adopting.
	packet = structurePacket(t, p, b, edited)
	raw = bytes.Replace(raw, []byte(`"Architecture geometry fixture"`), []byte(`"Unreviewed source change"`), 1)
	if e = os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "source_changed_since_review") {
		t.Fatalf("drift: %v", e)
	}
}
func TestStructureDeletionPreservesReviewedTextAndOtherNativeLayout(t *testing.T) {
	p, b := geometryFixture(t)
	inv, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := inv["node06"].geometry
	g.Y += 12
	edited := reconcileEditFields(t, b, map[string]string{"node06.text": "Reconciled rules"}) // text helper uses baseline
	edited = geometryEdited(t, p, &TextBaseline{files: map[string][]byte{"deck.pptx": edited, "layout-report.json": b.files["layout-report.json"]}}, map[string]NativeGeometry{"node06": g}, nil)
	edited = removeGeometryObjects(t, edited, "node08")
	packet := structurePacket(t, p, b, edited)
	var d TextReviewDecisions
	if e := strictInto(json.RawMessage(structureDecisions(packet, "use_native")), &d); e != nil {
		t.Fatal(e)
	}
	for _, f := range packet.Report.Geometry {
		if f.Status == "native_only" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, "use_native", "Accept move"})
		}
	}
	for _, f := range packet.Report.Fields {
		if f.Status == "native_only" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, "use_native", "Accept text"})
		}
	}
	if _, e := AdoptTextReviewPacket(p, packet, canonical(d), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	after, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if after.Document.Slides[0].Values["node06.text"] != "Reconciled rules" || after.Document.Slides[0].NativeGeometry["node06"].Y != g.Y {
		t.Fatal("lost reviewed edits")
	}
}

func copyGeometryObject(t *testing.T, raw []byte, source, name, text string, dy float64, keepTags ...bool) []byte {
	t.Helper()
	return lineageEdit(t, raw, "ppt/slides/slide1.xml", func(data []byte) []byte {
		inv, _, e := geometryInventory(data)
		if e != nil {
			t.Fatal(e)
		}
		o := inv[source]
		if o == nil {
			t.Fatal("source missing")
		}
		subtree := append([]byte{}, data[o.span.start:o.span.end]...)
		span, e := lineageSpans(subtree)
		if e != nil {
			t.Fatal(e)
		}
		patches := []lineagePatch{}
		ordinal := 10000
		var walk func(*lineageSpan)
		walk = func(s *lineageSpan) {
			if s.node.Name.Local == "custDataLst" && (len(keepTags) == 0 || !keepTags[0]) {
				patches = append(patches, lineagePatch{s.start, s.end, ""})
				return
			}
			if s.node.Name.Local == "cNvPr" {
				ordinal++
				replacement := fmt.Sprintf(`<p:cNvPr id="%d" name="%s"/>`, ordinal, strings.Replace(geometryAttrName(s.node), source, name, 1))
				patches = append(patches, lineagePatch{s.start, s.end, replacement})
				return
			}
			if s.node.Name.Local == "t" && text != "" {
				patches = append(patches, lineagePatch{s.startEnd, s.endStart, text})
			}
			for _, c := range s.children {
				walk(c)
			}
		}
		walk(span)
		g := o.geometry
		g.Y += dy
		patches = append(patches, lineagePatch{o.xf.start - o.span.start, o.xf.end - o.span.start, geometryXML(g, o.xf.node.Name.Space)})
		subtree, e = lineageApply(subtree, patches)
		if e != nil {
			t.Fatal(e)
		}
		data, e = lineageApply(data, []lineagePatch{{o.span.end, o.span.end, string(subtree)}})
		if e != nil {
			t.Fatal(e)
		}
		return data
	})
}
func geometryAttrName(n *xmlNode) string { return lineageAttr(n, "", "name") }
func mappedStructurePacket(t *testing.T, p *Project, b *TextBaseline, edited []byte) *TextReviewPacket {
	t.Helper()
	m := NativeStructureMap{StructureMapSchema, []NativeCopyMapping{{"architecture-slide", "native-monitoring", "node07", "monitoring"}}}
	packet, e := WriteMappedGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine, canonical(m))
	if e != nil {
		t.Fatal(e)
	}
	return packet
}
func TestStructureMappedNativeCopyPersistsIndependentTextAndGeometry(t *testing.T) {
	p, b := geometryFixture(t)
	edited := copyGeometryObject(t, b.files["deck.pptx"], "node07", "native-monitoring", "Monitoring", 60)
	packet := mappedStructurePacket(t, p, b, edited)
	if len(packet.Report.Structure) != 1 || len(packet.Report.ManualReview) != 0 {
		t.Fatalf("report %+v", packet.Report)
	}
	before := p.SourceHash()
	if _, e := AdoptTextReviewPacket(p, packet, structureDecisions(packet, "keep_yaml"), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if p.SourceHash() != before {
		t.Fatal("keep copy rewrote source")
	}
	packet = mappedStructurePacket(t, p, b, edited)
	receipt, e := AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(receipt.Changed) != 1 || len(receipt.ManualReview) > 0 {
		t.Fatalf("receipt %+v", receipt)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if p.Document.Slides[0].Values["monitoring.text"] != "Monitoring" || p.Document.Slides[0].Values["node07.text"] == "Monitoring" {
		t.Fatal("copy text ownership was not independent")
	}
	if _, e = AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	built, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	rebuilt, e := os.ReadFile(filepath.Join(p.Root, "builds", built.BuildID, "deck.pptx"))
	if e != nil {
		t.Fatal(e)
	}
	a, ao := geometryFromPackage(t, edited)
	z, zo := geometryFromPackage(t, rebuilt)
	normalized := map[string]NativeGeometry{}
	for name, o := range a {
		g := o.geometry
		g.SourceGeometrySHA256 = ""
		g.Parent = strings.Replace(g.Parent, "native-monitoring", "monitoring", 1)
		normalized[strings.Replace(name, "native-monitoring", "monitoring", 1)] = g
	}
	for name, o := range z {
		g := o.geometry
		g.SourceGeometrySHA256 = ""
		if !bytes.Equal(canonical(normalized[name]), canonical(g)) {
			t.Fatalf("copy geometry mismatch %s", name)
		}
	}
	normalizedOrder := map[string][]string{}
	for owner, names := range ao {
		for i := range names {
			names[i] = strings.Replace(names[i], "native-monitoring", "monitoring", 1)
		}
		normalizedOrder[strings.Replace(owner, "native-monitoring", "monitoring", 1)] = names
	}
	ao = normalizedOrder

	if !bytes.Equal(canonical(ao), canonical(zo)) {
		t.Fatal("native copy paint order changed")
	}
}
func TestStructureMappedCopyAndDeletionRequireCoupledDecisions(t *testing.T) {
	p, b := geometryFixture(t)
	edited := copyGeometryObject(t, b.files["deck.pptx"], "node07", "native-monitoring", "Monitoring", 60)
	edited = removeGeometryObjects(t, edited, "node08")
	packet := mappedStructurePacket(t, p, b, edited)
	if len(packet.Report.Structure) != 2 || len(packet.Report.ManualReview) > 0 {
		t.Fatalf("report %+v", packet.Report)
	}
	d := TextReviewDecisions{Schema: TextReviewDecisionsSchema, ReportSHA256: packet.ReportSHA256, Actor: "reviewer"}
	for _, f := range packet.Report.Structure {
		if f.Action == "copy_node" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, "use_native", "Accept copy"})
		}
	}
	before := p.SourceHash()
	if _, e := AdoptTextReviewPacket(p, packet, canonical(d), bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("incomplete structural order accepted")
	}
	after, e := Load(p.SourcePath)
	if e != nil || after.SourceHash() != before {
		t.Fatal("failed coupled decision wrote source", e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
}
func TestStructureCopiesRejectGuessedOwnershipAndStyleChanges(t *testing.T) {
	p, b := geometryFixture(t)
	edited := copyGeometryObject(t, b.files["deck.pptx"], "node07", "native-monitoring", "Monitoring", 60)
	for _, mode := range []string{"wrong-source", "duplicate-map", "style"} {
		t.Run(mode, func(t *testing.T) {
			m := NativeStructureMap{StructureMapSchema, []NativeCopyMapping{{"architecture-slide", "native-monitoring", "node07", "monitoring"}}}
			raw := edited
			switch mode {
			case "wrong-source":
				m.Copies[0].SourceNode = "node01"
			case "duplicate-map":
				m.Copies = append(m.Copies, m.Copies[0])
			case "style":
				raw = lineageEdit(t, raw, "ppt/slides/slide1.xml", func(data []byte) []byte {
					inv, _, e := geometryInventory(data)
					if e != nil {
						t.Fatal(e)
					}
					o := inv["native-monitoring.text"]
					replacement := strings.Replace(string(data[o.span.start:o.span.end]), `sz="`, `sz="1`, 1)
					data, e = lineageApply(data, []lineagePatch{{o.span.start, o.span.end, replacement}})
					if e != nil {
						t.Fatal(e)
					}
					return data
				})
			}
			if _, e := WriteMappedGeometryReviewPacket(p, b, raw, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine, canonical(m)); e == nil {
				t.Fatal("unsupported copy mapped")
			}
		})
	}
}

func TestStructureDeletionRequiresAllMissingAttachedEdges(t *testing.T) {
	p, _ := geometryFixture(t)
	if _, e := ConnectDiagram(p, "architecture-slide", "service-edge", "node06", "right", "node07", "left", "end", "solid", "test", "Dependency fixture", bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	partial := structurePacket(t, p, b, removeGeometryObjects(t, b.files["deck.pptx"], "node06"))
	if len(partial.Report.Structure) != 1 || partial.Report.Structure[0].Status != "manual_review" {
		t.Fatalf("dangling edge proposal %+v", partial.Report.Structure)
	}
	packet := structurePacket(t, p, b, removeGeometryObjects(t, b.files["deck.pptx"], "node06", "service-edge"))
	if len(packet.Report.Structure) != 2 {
		t.Fatalf("report %+v", packet.Report.Structure)
	}
	d := TextReviewDecisions{Schema: TextReviewDecisionsSchema, ReportSHA256: packet.ReportSHA256, Actor: "reviewer"}
	for _, f := range packet.Report.Structure {
		if f.NodeID == "node06" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, "use_native", "Reviewed removal"})
		}
	}
	if _, e = AdoptTextReviewPacket(p, packet, canonical(d), bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("incident edge silently removed")
	}
	if _, e = AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
}
func TestStructureMappedCopyOutsideFrameCannotWriteSource(t *testing.T) {
	p, b := geometryFixture(t)
	edited := copyGeometryObject(t, b.files["deck.pptx"], "node07", "native-monitoring", "Monitoring", 1000)
	packet := mappedStructurePacket(t, p, b, edited)
	before := p.SourceHash()
	if _, e := AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("off-frame copy adopted")
	}
	after, e := Load(p.SourcePath)
	if e != nil || before != after.SourceHash() {
		t.Fatal("failed frame check wrote source", e)
	}
}

func TestStructureMappedCopyWithInheritedTagsRetainsExactInput(t *testing.T) {
	p, b := geometryFixture(t)
	edited := copyGeometryObject(t, b.files["deck.pptx"], "node07", "native-monitoring", "Monitoring", 60, true)
	view, e := InspectNativeLineage(edited, b.Objects)
	if e != nil {
		t.Fatal(e)
	}
	if !lineageIssue(view, "shape_duplicated") {
		t.Fatal("fixture did not inherit tags")
	}
	// An original's simultaneous move remains separately reconcilable.
	edited = lineageEdit(t, edited, "ppt/slides/slide1.xml", func(data []byte) []byte {
		inv, _, e := geometryInventory(data)
		if e != nil {
			t.Fatal(e)
		}
		o := inv["node07"]
		g := o.geometry
		g.Y += 12
		data, e = lineageApply(data, []lineagePatch{{o.xf.start, o.xf.end, geometryXML(g, o.xf.node.Name.Space)}})
		if e != nil {
			t.Fatal(e)
		}
		return data
	})
	packet := mappedStructurePacket(t, p, b, edited)
	if len(packet.Report.ManualReview) != 0 || packet.Report.Counts["geometry_native_only"] != 1 || !bytes.Equal(packet.Edited, edited) {
		t.Fatalf("report %+v", packet.Report)
	}
	var d TextReviewDecisions
	if e = strictInto(json.RawMessage(structureDecisions(packet, "use_native")), &d); e != nil {
		t.Fatal(e)
	}
	for _, f := range packet.Report.Geometry {
		if f.Status == "native_only" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, "use_native", "Accept original move"})
		}
	}
	receipt, e := AdoptTextReviewPacket(p, packet, canonical(d), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	retained, e := os.ReadFile(filepath.Join(p.Root, receipt.RetainedEdited))
	if e != nil || !bytes.Equal(retained, edited) {
		t.Fatal("edited original was changed", e)
	}
}
func TestStructureMappingOriginalAsNewCopyRefused(t *testing.T) {
	p, b := geometryFixture(t)
	m := NativeStructureMap{StructureMapSchema, []NativeCopyMapping{{"architecture-slide", "node07", "node07", "monitoring"}}}
	if _, e := WriteMappedGeometryReviewPacket(p, b, b.files["deck.pptx"], filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine, canonical(m)); e == nil {
		t.Fatal("original identity erased by copy mapping")
	}
}
