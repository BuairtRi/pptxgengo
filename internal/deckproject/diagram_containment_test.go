package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func containmentFixture(t *testing.T) (*Project, *TextBaseline) {
	t.Helper()
	p, _ := geometryFixture(t)
	out, e := ContainDiagram(p, "architecture-slide", []string{"node06", "node07", "node08"}, "node05", wmdesign.DiagramPadding{Top: 28, Left: 12, Right: 12, Bottom: 4}, "test", "Keep services in the Services box", bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Inspection.Containment) != 3 {
		t.Fatal("missing clearances")
	}
	p, e = Load(p.SourcePath)
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
	return p, b
}
func TestContainmentPreviewSourceMovementAndReset(t *testing.T) {
	p, _ := geometryFixture(t)
	before := p.SourceHash()
	out, e := ContainDiagram(p, "architecture-slide", []string{"node06"}, "node05", wmdesign.DiagramPadding{Top: 28, Left: 12, Right: 12, Bottom: 4}, "test", "Preview boundaries", bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if out.Applied || before != p.SourceHash() {
		t.Fatal("preview wrote source")
	}
	p, _ = containmentFixture(t)
	before = p.SourceHash()
	dy := 60.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Outside Services but inside frame", Operations: []DiagramOperation{{Action: "move", ID: "node06", DY: &dy}}}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "containment") {
		t.Fatalf("inner frame escaped: %v", e)
	}
	after, e := Load(p.SourcePath)
	if e != nil || before != after.SourceHash() {
		t.Fatal("invalid patch wrote source", e)
	}
	patch.Operations = []DiagramOperation{{Action: "reset_native_layout", ID: "architecture-slide"}}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	after, e = Load(p.SourcePath)
	if e != nil || len(after.Document.Slides[0].DiagramContainment) != 3 || after.Document.Slides[0].NativeGeometryTemplate == nil {
		t.Fatal("reset lost constraints or pin", e)
	}
}
func TestContainmentNativeReconciliationRejectsEscapeBeforeWrites(t *testing.T) {
	p, b := containmentFixture(t)
	before := p.SourceHash()
	// Native edits bypass the authored validator, as real PowerPoint edits do.
	edited := lineageEdit(t, b.files["deck.pptx"], "ppt/slides/slide1.xml", func(raw []byte) []byte {
		inv, _, e := geometryInventory(raw)
		if e != nil {
			t.Fatal(e)
		}
		o := inv["node06"]
		g := o.geometry
		g.Y += 60
		raw, e = lineageApply(raw, []lineagePatch{{o.xf.start, o.xf.end, geometryXML(g, o.xf.node.Name.Space)}})
		if e != nil {
			t.Fatal(e)
		}
		return raw
	})
	packet := structurePacket(t, p, b, edited)
	if packet.Report.Counts["geometry_native_only"] != 1 {
		t.Fatal("missing movement proposal")
	}
	if _, e := AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "containment") {
		t.Fatalf("native escape adopted: %v", e)
	}
	after, e := Load(p.SourcePath)
	if e != nil || before != after.SourceHash() {
		t.Fatal("invalid native edit wrote source", e)
	}
}
func TestContainmentGroupsAffineCoordinatesAndDescendants(t *testing.T) {
	// Nonuniform parent scaling and rotation require inverse container axes;
	// world axis-aligned bounding boxes alone would incorrectly reject this.
	objects := map[string]*geometryObject{
		"parent": {geometry: NativeGeometry{Kind: "grpSp", X: 100, Y: 50, W: 400, H: 200, Rotation: 25, Child: &wmdesign.Rect{X: 0, Y: 0, W: 200, H: 200}}},
		"box":    {geometry: NativeGeometry{Kind: "grpSp", Parent: "parent", X: 10, Y: 10, W: 160, H: 160, Rotation: 15, Child: &wmdesign.Rect{X: 0, Y: 0, W: 160, H: 160}}},
		"member": {geometry: NativeGeometry{Kind: "grpSp", Parent: "box", X: 20, Y: 20, W: 30, H: 30, Child: &wmdesign.Rect{X: 0, Y: 0, W: 30, H: 30}}},
		"text":   {geometry: NativeGeometry{Kind: "sp", Parent: "member", X: 0, Y: 0, W: 30, H: 30}},
	}
	rules := map[string]wmdesign.DiagramContainment{"member": {Container: "box", Padding: wmdesign.DiagramPadding{Top: 10, Left: 10, Right: 10, Bottom: 10}}}
	checks, _, e := checkDiagramContainment(objects, rules)
	if e != nil {
		t.Fatal(e)
	}
	if len(checks) != 1 || checks[0].Clearance.Left < 19.99 || checks[0].Clearance.Top < 19.99 {
		t.Fatalf("wrong axes %+v", checks)
	}
	objects["text"].geometry.X = 200
	if _, _, e = checkDiagramContainment(objects, rules); e == nil {
		t.Fatal("child escaped despite unchanged group box")
	}
}
func TestContainmentCycleMissingPaddingAndOverlap(t *testing.T) {
	objects := map[string]*geometryObject{"box": {geometry: NativeGeometry{Kind: "sp", W: 200, H: 200}}, "a": {geometry: NativeGeometry{Kind: "sp", X: 20, Y: 20, W: 50, H: 50}}, "b": {geometry: NativeGeometry{Kind: "sp", X: 40, Y: 40, W: 50, H: 50}}}
	rules := map[string]wmdesign.DiagramContainment{"a": {Container: "box"}, "b": {Container: "box"}}
	_, overlaps, e := checkDiagramContainment(objects, rules)
	if e != nil || len(overlaps) != 1 {
		t.Fatal("missing overlap", e, overlaps)
	}
	rules["box"] = wmdesign.DiagramContainment{Container: "a"}
	if _, _, e = checkDiagramContainment(objects, rules); e == nil {
		t.Fatal("cycle accepted")
	}
	delete(rules, "box")
	rules["a"] = wmdesign.DiagramContainment{Container: "missing"}
	if _, _, e = checkDiagramContainment(objects, rules); e == nil {
		t.Fatal("missing container accepted")
	}
	rules["a"] = wmdesign.DiagramContainment{Container: "box", Padding: wmdesign.DiagramPadding{Left: -1}}
	if _, _, e = checkDiagramContainment(objects, rules); e == nil {
		t.Fatal("negative padding accepted")
	}
}
func TestContainmentDeletionAndCopyRetainMembership(t *testing.T) {
	p, b := containmentFixture(t)
	deleted := structurePacket(t, p, b, removeGeometryObjects(t, b.files["deck.pptx"], "node08"))
	if _, e := AdoptTextReviewPacket(p, deleted, structureDecisions(deleted, "use_native"), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := p.Document.Slides[0].DiagramContainment["node08"]; ok {
		t.Fatal("removed member constraint survived")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e = ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	// The source copy remains protected even if the new copy is placed elsewhere.
	edited := copyGeometryObject(t, b.files["deck.pptx"], "node07", "native-monitoring", "Monitoring", 60)
	packet := mappedStructurePacket(t, p, b, edited)
	before := p.SourceHash()
	if _, e = AdoptTextReviewPacket(p, packet, structureDecisions(packet, "use_native"), bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "containment") {
		t.Fatalf("copy escaped inherited constraint: %v", e)
	}
	p, e = Load(p.SourcePath)
	if e != nil || before != p.SourceHash() {
		t.Fatal("copy failure wrote source", e)
	}
	// An uncontain is explicit, previewed, and retained; geometry overrides persist.
	if _, e = ContainDiagram(p, "architecture-slide", []string{"node07"}, "", wmdesign.DiagramPadding{}, "test", "Remove an obsolete membership", bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
}
func TestContainmentSourcePinAndCanonicalReceipt(t *testing.T) {
	p, _ := containmentFixture(t)
	s := p.Document.Slides[0]
	s.Template.ID = "different"
	p.Document.Slides[0] = s
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(p.SourcePath); e == nil {
		t.Fatal("template pin silently changed")
	}
	p, _ = containmentFixture(t)
	// Exercise normal authored serialization: the optional rule survives reload.
	after, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(canonical(p.Document.Slides[0].DiagramContainment), canonical(after.Document.Slides[0].DiagramContainment)) {
		t.Fatal("rule lost on load")
	}
	var d Document
	if e = strictInto(json.RawMessage(after.Canonical), &d); e != nil {
		t.Fatal(e)
	}
	if len(d.Slides[0].DiagramContainment) != 3 {
		t.Fatal("canonical source omitted rules")
	}
	// Receipt includes rules without mutating baseline bytes.
	b, e := ReadTextBaseline(after, "", "")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(after.Root, "builds", b.Receipt.BuildID, "source.canonical.json"))
	if e != nil || !bytes.Contains(raw, []byte("diagram_containment")) {
		t.Fatal("receipt source omitted rules", e)
	}
}

func TestContainmentSuccessfulMappedCopyAndIdempotence(t *testing.T) {
	p, b := containmentFixture(t)
	edited := copyGeometryObject(t, b.files["deck.pptx"], "node07", "native-monitoring", "Monitoring", 0, true)
	edited = lineageEdit(t, edited, "ppt/slides/slide1.xml", func(raw []byte) []byte {
		inv, _, e := geometryInventory(raw)
		if e != nil {
			t.Fatal(e)
		}
		o := inv["native-monitoring"]
		g := o.geometry
		g.X += 168
		raw, e = lineageApply(raw, []lineagePatch{{o.xf.start, o.xf.end, geometryXML(g, o.xf.node.Name.Space)}})
		if e != nil {
			t.Fatal(e)
		}
		return raw
	})
	edited = removeGeometryObjects(t, edited, "node08")
	packet := mappedStructurePacket(t, p, b, edited)
	decisions := structureDecisions(packet, "use_native")
	if _, e := AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	after, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	rules := after.Document.Slides[0].DiagramContainment
	if rules["monitoring"] != rules["node07"] || rules["monitoring"].Container != "node05" {
		t.Fatal("copy lost inherited membership", rules)
	}
	if _, ok := rules["node08"]; ok {
		t.Fatal("deleted member survived")
	}
	inspect, e := InspectDiagram(after, "architecture-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil || len(inspect.Containment) != 3 || len(inspect.Overlaps) != 0 {
		t.Fatal("copy did not fit freed slot", e, inspect.Overlaps)
	}
	hash := after.SourceHash()
	if _, e = AdoptTextReviewPacket(after, packet, decisions, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	after, e = Load(p.SourcePath)
	if e != nil || after.SourceHash() != hash {
		t.Fatal("repeat adoption changed source", e)
	}
}
func TestContainmentSourceRemovalAndMissingContainerGuard(t *testing.T) {
	p, _ := containmentFixture(t)
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Remove obsolete service", Operations: []DiagramOperation{{Action: "remove", ID: "node08", IncidentEdges: "remove"}}}
	if _, e := PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Document.Slides[0].DiagramContainment) != 2 {
		t.Fatal("deleted source membership survived")
	}
	hash := p.SourceHash()
	patch.Operations[0].ID = "node05"
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "containment") {
		t.Fatal("missing container accepted", e)
	}
	after, e := Load(p.SourcePath)
	if e != nil || hash != after.SourceHash() {
		t.Fatal("missing container wrote source", e)
	}
}

func TestContainmentMembershipChangesConflictWithNativeDeletion(t *testing.T) {
	p, b := containmentFixture(t)
	if _, e := ContainDiagram(p, "architecture-slide", []string{"node08"}, "node05", wmdesign.DiagramPadding{Top: 30, Right: 12, Bottom: 4, Left: 12}, "test", "Change membership after baseline", bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	packet := structurePacket(t, p, b, removeGeometryObjects(t, b.files["deck.pptx"], "node08"))
	if len(packet.Report.Structure) != 1 || packet.Report.Structure[0].Status != "conflict" {
		t.Fatal("membership edit lost during structural comparison", packet.Report.Structure)
	}
}
