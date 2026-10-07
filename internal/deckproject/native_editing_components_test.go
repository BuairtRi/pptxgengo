package deckproject

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func nativeEditingDiagramFixture(t *testing.T) (*Project, *TextBaseline) {
	t.Helper()
	p := example(t)
	zones := map[string]Zone{}
	for _, name := range []string{"title", "input_copy", "output_copy"} {
		role := "node-copy"
		if name == "title" {
			role = "slide-title"
		}
		zones[name] = Zone{Role: role, Required: true, Schema: map[string]any{"type": "string", "maxLength": 70}}
	}
	placement := func(start, count int) *Placement {
		return &Placement{Zone: "body", Span: &Span{Start: start, Count: count, Y: 36, H: 144}}
	}
	node := func(id, copy string, start int) Node {
		return Node{ID: id, Kind: "component", Placement: placement(start, 6), Definition: &Reference{Scope: "shared", ID: "wmds/component/editable-block"}, Arguments: map[string]any{"surface": "subtle", "text": map[string]any{"binding": copy}, "style": "body"}}
	}
	p.Document.LocalTemplates["native-diagram"] = LocalTemplate{Name: "Synthetic native editing diagram", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: zones, Nodes: []Node{
		// Forward endpoints deliberately precede their targets in paint order.
		{ID: "input-output", Kind: "component", Placement: placement(1, 12), Definition: &Reference{Scope: "shared", ID: "wmds/component/attached-connector"}, Arguments: map[string]any{"from": map[string]any{"node": "input", "site": "right"}, "to": map[string]any{"node": "output", "site": "left"}, "head": "end"}},
		node("input", "input_copy", 1), node("output", "output_copy", 7),
	}}
	p.Document.Slides = []Slide{{ID: "native-diagram", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "native-diagram"}, Values: map[string]any{"title": "Synthetic native diagram", "input_copy": "Editable input", "output_copy": "Editable output"}}}
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

func TestNativeEditabilityCombinedBlocksAndAttachedConnector(t *testing.T) {
	p, b := nativeEditingDiagramFixture(t)
	r, e := NativeEditability(b)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counts["native_cxnSp"] != 1 || r.Counts["text_and_paint_in_same_shape"] != 2 || r.Counts["maximum_group_depth"] != 0 {
		t.Fatal("incorrect native editing structure", r.Counts)
	}
	objects := map[string]ObjectRecord{}
	for _, o := range b.Objects.Objects {
		objects[o.NativeName] = o
	}
	for _, name := range []string{"input", "output"} {
		o := objects[name]
		if o.NativeKind != "sp" || o.NativeParentToken != "" || o.ShapeToken == "" || len(o.Fields) != 1 || o.Fields[0].Status != "plain_text_baseline" || o.Fields[0].SourceSlot != name+"_copy" {
			t.Fatalf("combined shape lost its exact source field: %+v", o)
		}
		if _, exists := objects[name+".surface"]; exists {
			t.Fatal("retained separate surface")
		}
		if _, exists := objects[name+".text"]; exists {
			t.Fatal("retained separate text")
		}
	}
	// Serialized outer geometry/insets retain the measured inner text envelope.
	var layout wmdesign.Report
	if e := json.Unmarshal(b.files["layout-report.json"], &layout); e != nil {
		t.Fatal(e)
	}
	for _, text := range layout.Slides[0].Texts {
		if text.ID != "input" && text.ID != "output" {
			continue
		}
		if text.NativeShape == nil {
			t.Fatal("missing combined geometry record")
		}
		var shape *xmlNode
		for _, n := range b.inspection.Objects {
			if n.NativeName == text.ID {
				shape = n.shape
			}
		}
		body := directXML(directXML(shape, lineagePML, "txBody"), drawingML, "bodyPr")
		outer := text.NativeShape.Rect
		for attribute, points := range map[string]float64{"lIns": text.Rect.X - outer.X, "rIns": outer.X + outer.W - text.Rect.X - text.Rect.W, "tIns": text.Rect.Y - outer.Y, "bIns": outer.Y + outer.H - text.Rect.Y - text.Rect.H} {
			value, e := strconv.Atoi(attr(body, attribute))
			if e != nil || math.Abs(float64(value)-math.Round(points*12700)) > 1 {
				t.Fatalf("inner text envelope changed %s/%s: %s != %f", text.ID, attribute, attr(body, attribute), points)
			}
		}
	}
	edge := objects["input-output"]
	if edge.NativeKind != "cxnSp" || edge.NativeParentToken != "" {
		t.Fatal("connector is not native top-level", edge)
	}
	for _, unit := range r.Objects {
		if unit.SelectionName == "input-output" && (unit.Definition != "scene.attached-connector" || unit.Connection == nil || unit.Connection.Begin.ShapeToken != objects["input"].ShapeToken || unit.Connection.End.ShapeToken != objects["output"].ShapeToken || unit.Connection.Begin.Site != 3 || unit.Connection.End.Site != 1) {
			t.Fatal("inventory lost declared native endpoint evidence", unit)
		}
		if unit.SelectionName == "input" && unit.Definition != "scene.editable-block" {
			t.Fatal("combined definition absent", unit)
		}
	}
	var shape *xmlNode
	for _, n := range b.inspection.Objects {
		if n.ShapeToken == edge.ShapeToken {
			shape = n.shape
		}
	}
	props := directXML(directXML(shape, lineagePML, "nvCxnSpPr"), lineagePML, "cNvCxnSpPr")
	start, end := directXML(props, drawingML, "stCxn"), directXML(props, drawingML, "endCxn")
	if start == nil || end == nil || attr(start, "id") != objects["input"].NativeID || attr(end, "id") != objects["output"].NativeID || attr(start, "idx") != "3" || attr(end, "idx") != "1" {
		t.Fatal("connection targets drifted")
	}
	if output := os.Getenv("PPTXGENGO_NATIVE_DIAGRAM_PILOT_OUT"); output != "" {
		writeNativeEditabilityPilotFixture(t, p, r, output)
	}
}

func TestNativeEditabilityConnectorChangesRemainManualDuringTextReview(t *testing.T) {
	p, b := nativeEditingDiagramFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"input": "Reviewed input"})
	var part string
	for _, o := range b.Objects.Objects {
		if o.NativeName == "input-output" {
			part = o.NativePart
		}
	}
	edited = lineageEdit(t, edited, part, func(raw []byte) []byte {
		return bytes.Replace(raw, []byte(`idx="3"/><a:endCxn`), []byte(`idx="0"/><a:endCxn`), 1)
	})
	if bytes.Equal(edited, reconcileEditFields(t, b, map[string]string{"input": "Reviewed input"})) {
		t.Fatal("fixture did not change connector")
	}
	packet, e := WriteTextReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"))
	if e != nil {
		t.Fatal(e)
	}
	if packet.Report.Counts["native_only"] != 1 || !hasReconcileIssue(packet.Report, "native_structure_or_format_changed") {
		t.Fatal("connector change hidden", packet.Report)
	}
	receipt, e := AdoptTextReviewPacket(p, packet, reconciliationDecisions(t, packet, "use_native"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(receipt.Changed) != 1 || len(receipt.ManualReview) == 0 {
		t.Fatal("connector silently adopted or omitted", receipt)
	}
}

func TestNativeEditabilityCombinedTextReviewAdoptionAndRebuild(t *testing.T) {
	p, b := nativeEditingDiagramFixture(t)
	edits := map[string]string{"input": "Reviewed input", "output": "Reviewed output", "title": "Reviewed diagram"}
	edited := reconcileEditFields(t, b, edits)
	packet, e := WriteTextReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"))
	if e != nil {
		t.Fatal(e)
	}
	if packet.Report.Counts["native_only"] != 3 || len(packet.Report.ManualReview) != 0 {
		t.Fatal("combined text not precisely proposed", packet.Report)
	}
	decisions := reconciliationDecisions(t, packet, "use_native")
	receipt, e := AdoptTextReviewPacket(p, packet, decisions, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(receipt.Changed) != 3 {
		t.Fatal("reviewed fields not adopted", receipt)
	}
	after, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	again, e := AdoptTextReviewPacket(after, packet, decisions, bundle(t), wmdesign.CandidateEngine)
	if e != nil || !bytes.Equal(canonical(receipt), canonical(again)) {
		t.Fatal("repeat adoption changed result", e)
	}
	if _, e = Build(after, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	rebuilt, e := ReadTextBaseline(after, "", "")
	if e != nil {
		t.Fatal(e)
	}
	found := 0
	for _, o := range rebuilt.Objects.Objects {
		if text, ok := edits[o.NativeName]; ok {
			found++
			if o.NativeText != text || len(o.Fields) != 1 || o.Fields[0].Status != "plain_text_baseline" {
				t.Fatal("rebuilt source field changed", o)
			}
		}
	}
	r, e := NativeEditability(rebuilt)
	if e != nil || found != 3 || r.Counts["native_cxnSp"] != 1 || r.Counts["text_and_paint_in_same_shape"] != 2 {
		t.Fatal("rebuild lost native structure", e, r.Counts)
	}
}
