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

func geometryFixture(t *testing.T) (*Project, *TextBaseline) {
	t.Helper()
	p := example(t)
	draft, e := ScaffoldTemplate(bundle(t), "architecture/nested", wmdesign.CandidateEngine, "Architecture geometry fixture", 2026)
	if e != nil {
		t.Fatal(e)
	}
	p.Document.LocalTemplates = map[string]LocalTemplate{"architecture": draft.Template}
	p.Document.Slides = []Slide{{ID: "architecture-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "architecture"}, Values: draft.SyntheticSourceValues}}
	raw, e := json.Marshal(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
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
func geometryFromPackage(t *testing.T, raw []byte) (map[string]*geometryObject, map[string][]string) {
	t.Helper()
	pkg, e := openLineagePackage(raw)
	if e != nil {
		t.Fatal(e)
	}
	data, e := pkg.read("ppt/slides/slide1.xml")
	if e != nil {
		t.Fatal(e)
	}
	objects, orders, e := geometryInventory(data)
	if e != nil {
		t.Fatal(e)
	}
	for name, o := range objects {
		o.geometry.SourceGeometrySHA256 = nativeGeometryBasis(objects, name)
	}
	return objects, orders
}
func geometryEdited(t *testing.T, p *Project, b *TextBaseline, changes map[string]NativeGeometry, orders map[string][]string) []byte {
	t.Helper()
	c, e := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	var report wmdesign.Report
	if e = strictInto(json.RawMessage(b.files["layout-report.json"]), &report); e != nil {
		t.Fatal(e)
	}
	c.Document.Slides[0].NativeGeometry = changes
	c.Document.Slides[0].NativeOrder = orders
	raw, e := applyNativeGeometry(b.files["deck.pptx"], c.Document, &report)
	if e != nil {
		t.Fatal(e)
	}
	return raw
}
func geometryDecisions(packet *TextReviewPacket) []byte {
	d := TextReviewDecisions{Schema: TextReviewDecisionsSchema, ReportSHA256: packet.ReportSHA256, Actor: "geometry test operator"}
	for _, f := range packet.Report.Geometry {
		if f.Status == "native_only" || f.Status == "conflict" {
			d.Decisions = append(d.Decisions, TextReviewDecision{f.ID, "use_native", "Accept the controlled geometry edit"})
		}
	}
	return canonical(d)
}
func TestGeometryNativeGroupTransformAdoptsAndRebuilds(t *testing.T) {
	p, b := geometryFixture(t)
	objects, orders := geometryFromPackage(t, b.files["deck.pptx"])
	move := objects["node06"].geometry
	move.Y += 12
	resize := objects["node07"].geometry
	resize.H += 12
	rotate := objects["node08"].geometry
	rotate.Rotation = 7
	rotate.FlipH = true
	desired := append([]string{}, orders[""]...)
	for i, n := range desired {
		if n == "node06" {
			desired = append(append(desired[:i:i], desired[i+1:]...), n)
			break
		}
	}
	edited := geometryEdited(t, p, b, map[string]NativeGeometry{"node06": move, "node07": resize, "node08": rotate}, map[string][]string{"": desired})
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if n := packet.Report.Counts["geometry_native_only"]; n != 4 {
		t.Fatalf("expected three transforms and order, got %d: %+v", n, packet.Report.Geometry)
	}
	if len(packet.Report.ManualReview) != 0 {
		t.Fatalf("supported changes left manual issues: %+v", packet.Report.ManualReview)
	}
	decisions := geometryDecisions(packet)
	adopted, e := AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(adopted.Changed) != 4 {
		t.Fatal(adopted)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal("idempotent geometry adoption", e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	newBase, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	rebuilt, newOrder := geometryFromPackage(t, newBase.files["deck.pptx"])
	native, _ := geometryFromPackage(t, edited)
	for name, o := range native {
		if !bytes.Equal(canonical(o.geometry), canonical(rebuilt[name].geometry)) {
			t.Fatalf("native transform did not persist: %s", name)
		}
	}
	if !bytes.Equal(canonical(newOrder[""]), canonical(desired)) {
		t.Fatal("paint order not persisted")
	}
	before, e := geometryWorld(native)
	if e != nil {
		t.Fatal(e)
	}
	after, e := geometryWorld(rebuilt)
	if e != nil {
		t.Fatal(e)
	}
	for name, rect := range before {
		if rect != after[name] {
			t.Fatalf("group world bounds changed: %s", name)
		}
	}
	if !bytes.Equal(b.files["deck.pptx"], mustReadGeometryFile(t, filepath.Join(p.Root, "builds", b.Receipt.BuildID, "deck.pptx"))) {
		t.Fatal("original baseline mutated")
	}
}
func mustReadGeometryFile(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestGeometryThreeWayConflictAndSourceDrift(t *testing.T) {
	p, b := geometryFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	native := objects["node06"].geometry
	native.Y += 12
	edited := geometryEdited(t, p, b, map[string]NativeGeometry{"node06": native}, nil)
	yaml := objects["node06"].geometry
	yaml.Y += 6
	p.Document.Slides[0].NativeGeometry = map[string]NativeGeometry{"node06": yaml}
	p.Document.Slides[0].NativeGeometryTemplate = &p.Document.Slides[0].Template
	raw := canonical(p.Document)
	if e := os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	var e error
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if packet.Report.Counts["geometry_conflict"] != 1 {
		t.Fatal(packet.Report.Geometry)
	}
	original := append([]byte{}, p.Raw...)
	if e = os.WriteFile(p.SourcePath, append(raw, '\n'), 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "source_changed_since_review") {
		t.Fatal("source drift accepted", e)
	}
	if e = os.WriteFile(p.SourcePath, original, 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
}
func TestDiagramPatchPreviewMoveAddRemoveAndBounds(t *testing.T) {
	p, _ := geometryFixture(t)
	before := p.SourceHash()
	dy := 12.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Rearrange architecture", Operations: []DiagramOperation{{Action: "move", ID: "node06", DY: &dy}, {Action: "remove", ID: "node08"}}}
	out, e := PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if out.Applied || out.AfterSHA256 == before {
		t.Fatal(out)
	}
	fresh, e := Load(p.SourcePath)
	if e != nil || fresh.SourceHash() != before {
		t.Fatal("preview mutated source", e)
	}
	if p.SourceHash() != before || p.Document.LocalTemplates["architecture"].Nodes[5].Placement.Rect.Y == out.Inspection.Nodes[5].Placement.Rect.Y {
		t.Fatal("preview mutated caller source")
	}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	node := Node{ID: "new-service", Kind: "component", Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 300, Y: 270, W: 162, H: 36}}, Definition: &Reference{Scope: "shared", ID: "wmds/component/editable-block"}, Arguments: map[string]any{"text": "New service", "surface": "light", "style": "body"}}
	patch.Operations = []DiagramOperation{{Action: "add", ID: node.ID, Node: &node}}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	prior := p.SourceHash()
	bad := 1000.
	patch.Operations = []DiagramOperation{{Action: "move", ID: "node06", DX: &bad}}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("off-frame patch accepted")
	}
	p, e = Load(p.SourcePath)
	if e != nil || p.SourceHash() != prior {
		t.Fatal("failed patch mutated source", e)
	}
}
func TestGeometryConnectorRelativeAllocation(t *testing.T) {
	args := map[string]any{"points": [][2]float64{{255, 234}, {291, 234}}, "head": "none"}
	b, e := editableSceneAllocation("connector", args, wmdesign.Rect{X: 255, Y: 234, W: 36, H: 0})
	if e != nil {
		t.Fatal(e)
	}
	if b.H != 1./12700 {
		t.Fatal(b)
	}
	raw, e := wmdesign.ComposeSceneNode("connector", args, b)
	if e != nil {
		t.Fatal(e)
	}
	var src struct{ Points [][2]float64 }
	if e = json.Unmarshal(raw, &src); e != nil {
		t.Fatal(e)
	}
	if src.Points[0] != [2]float64{255, 234} || src.Points[1] != [2]float64{291, 234} {
		t.Fatal("detach moved connector", src)
	}
	b.Y += 12
	raw, e = wmdesign.ComposeSceneNode("connector", args, b)
	if e != nil {
		t.Fatal(e)
	}
	json.Unmarshal(raw, &src)
	if src.Points[1] != [2]float64{291, 246} {
		t.Fatal("placement did not move connector", src)
	}
}

func TestDiagramConnectionsRecalculateAndRequireIncidentEdgeDecision(t *testing.T) {
	p, _ := geometryFixture(t)
	out, e := ConnectDiagram(p, "architecture-slide", "service-edge", "node06", "right", "node07", "left", "end", "solid", "test", "Model the service dependency", bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Inspection.Connections) != 1 {
		t.Fatal("missing calculated endpoint data")
	}
	before := out.Inspection.Connections[0]
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	dy := 12.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Move dependency source", Operations: []DiagramOperation{{Action: "move", ID: "node06", DY: &dy}}}
	out, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	after := out.Inspection.Connections[0]
	if after.From.Y != before.From.Y+12 || after.To != before.To {
		t.Fatal("connection endpoints were not recalculated", before, after)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	patch.Operations = []DiagramOperation{{Action: "remove", ID: "node06"}}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("incident edge was silently deleted")
	}
	patch.Operations[0].IncidentEdges = "remove"
	out, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Inspection.Connections) != 0 {
		t.Fatal("dangling edge survived")
	}
}
func TestDiagramArrangePreservesSourceOnPreview(t *testing.T) {
	p, _ := geometryFixture(t)
	before := p.SourceHash()
	out, e := ArrangeDiagram(p, "architecture-slide", []string{"node06", "node07", "node08"}, "top", "horizontal", "test", "Arrange three service blocks", bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if out.Applied || p.SourceHash() != before {
		t.Fatal("preview mutated source")
	}
	if len(out.Patch.Operations) != 3 {
		t.Fatal(out)
	}
	if _, e = ArrangeDiagram(p, "architecture-slide", []string{"node08", "node07", "node06"}, "", "horizontal", "test", "Invalid reversed distribution", bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("negative spacing accepted")
	}
}
func TestGeometryPacketTamperAndOutOfFrameRejected(t *testing.T) {
	p, b := geometryFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["node06"].geometry
	g.Y += 12
	edited := geometryEdited(t, p, b, map[string]NativeGeometry{"node06": g}, nil)
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	before := p.SourceHash()
	packet.Report.Geometry[0].Name = "different-owner"
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal("public observations should be reloaded", e)
	}
	// A direct source transaction cannot smuggle off-frame transforms through fit.
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	g.X = 2000.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Reject off-frame geometry", Operations: []DiagramOperation{{Action: "transform", ID: "node06", Geometry: &g}}}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("off-frame transform accepted")
	}
	fresh, e := Load(p.SourcePath)
	if e != nil || fresh.SourceHash() != p.SourceHash() || fresh.SourceHash() == before {
		t.Fatal("failed transform changed source", e)
	}
}
func TestGeometrySourceGroupSpaceChangeIsExplicitReview(t *testing.T) {
	p, b := geometryFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["node06"].geometry
	g.Y += 12
	edited := geometryEdited(t, p, b, map[string]NativeGeometry{"node06": g}, nil)
	dy := 6.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Simultaneous source move", Operations: []DiagramOperation{{Action: "move", ID: "node06", DY: &dy}}}
	if _, e := PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	var e error
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range packet.Report.Geometry {
		if f.Name == "node06" {
			if f.Status != "manual_review" || !strings.Contains(f.Reason, "coordinate space changed") {
				t.Fatal("incompatible group basis was silently accepted", f)
			}
			return
		}
	}
	t.Fatal("missing group review field")
}
func TestGeometryTemplatePinPreventsSilentTemplateSwap(t *testing.T) {
	p, b := geometryFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["node06"].geometry
	p.Document.Slides[0].NativeGeometry = map[string]NativeGeometry{"node06": g}
	pinRef := p.Document.Slides[0].Template
	p.Document.Slides[0].NativeGeometryTemplate = &pinRef
	p.Document.Slides[0].Template = Reference{Scope: "shared", ID: "architecture/nested"}
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(p.SourcePath); e == nil || !strings.Contains(e.Error(), "native geometry") {
		t.Fatal("template changed without geometry review", e)
	}
}

func TestGeometryAttachedBlockMoveCannotLeaveArrowBehind(t *testing.T) {
	p, _ := geometryFixture(t)
	if _, e := ConnectDiagram(p, "architecture-slide", "service-edge", "node06", "right", "node07", "left", "end", "solid", "test", "Dependency", bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	var e error
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
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["node06"].geometry
	g.Y += 12
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Reject a disconnected move", Operations: []DiagramOperation{{Action: "transform", ID: "node06", Geometry: &g}}}
	before := p.SourceHash()
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "disconnects") {
		t.Fatal("disconnected arrow accepted", e)
	}
	p, e = Load(p.SourcePath)
	if e != nil || p.SourceHash() != before {
		t.Fatal("rejected move changed source", e)
	}
}

func TestGeometryAuthoredBasisDriftDoesNotDoubleApplyGroupMove(t *testing.T) {
	p, b := geometryFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["node06"].geometry
	g.Y += 12
	edited := geometryEdited(t, p, b, map[string]NativeGeometry{"node06": g}, nil)
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	t1 := p.Document.LocalTemplates["architecture"]
	t1.Nodes[5].Placement.Rect.Y += 6
	p.Document.LocalTemplates["architecture"] = t1
	if e = os.WriteFile(p.SourcePath, canonical(p.Document), 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e == nil || !strings.Contains(e.Error(), "source basis changed") {
		t.Fatal("old transform applied to a rebased source group", e)
	}
}

func TestDiagramExplicitResetEnablesSourceEditing(t *testing.T) {
	p, b := geometryFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	g := objects["node06"].geometry
	g.Y += 12
	edited := geometryEdited(t, p, b, map[string]NativeGeometry{"node06": g}, nil)
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTextReviewPacket(p, packet, geometryDecisions(packet), bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	dy := 6.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Deliberately return to authored placements", Operations: []DiagramOperation{{Action: "move", ID: "node06", DY: &dy}}}
	if _, e = PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("source move silently overrode reconciled layout")
	}
	patch.Operations = append([]DiagramOperation{{Action: "reset_native_layout", ID: "architecture-slide"}}, patch.Operations...)
	out, e := PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Inspection.NativeGeometry) != 0 || len(out.Inspection.NativeOrder) != 0 {
		t.Fatal("reset did not clear native layout")
	}
	for _, o := range out.Inspection.FinalNative {
		if o.Name == "node06" {
			if o.Geometry.Y != objects["node06"].geometry.Y+6 {
				t.Fatal("move was applied twice", o)
			}
			return
		}
	}
	t.Fatal("node missing")
}

func TestGeometryYAMLOnlyGroupMoveRetainsSourceWithoutNativeDecision(t *testing.T) {
	p, b := geometryFixture(t)
	dy := 6.
	patch := DiagramPatch{Schema: DiagramPatchSchema, Actor: "test", Reason: "Source-only move", Operations: []DiagramOperation{{Action: "move", ID: "node06", DY: &dy}}}
	if _, e := PatchDiagram(p, "architecture-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	var e error
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	packet, e := WriteGeometryReviewPacket(p, b, b.files["deck.pptx"], filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range packet.Report.Geometry {
		if f.Name == "node06" {
			if f.Status != "yaml_only" {
				t.Fatal(f)
			}
			return
		}
	}
	t.Fatal("missing source-only group field")
}
