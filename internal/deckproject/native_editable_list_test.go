package deckproject

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func editableListFixture(t *testing.T) (*Project, *TextBaseline, NativeLineageObject) {
	t.Helper()
	p := example(t)
	p.Document.LocalTemplates["native-list"] = LocalTemplate{
		Name:  "Synthetic native unordered list",
		Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"},
		Grid:  Reference{Scope: "shared", ID: "wmds/grid/12-columns"},
		Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string", "maxLength": 70}}},
		Nodes: []Node{{ID: "editable-list", Kind: "component",
			Placement:  &Placement{Zone: "body", Span: &Span{Start: 1, Count: 6, Y: 36, H: 216}},
			Definition: &Reference{Scope: "shared", ID: "wmds/component/editable-list"},
			Arguments:  map[string]any{"items": []any{"Review the source", "This longer item wraps within the text box while the final item follows it without needing a separate shape move.", "Share the deck"}},
			Keys:       map[string][]string{"items": {"review", "wrap", "share"}},
		}},
	}
	p.Document.Slides = []Slide{{ID: "native-list", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "native-list"}, Values: map[string]any{"title": "Native list editing"}}}
	node, e := editYAMLNode(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := encodeSourceYAML(node)
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
	var found NativeLineageObject
	n := 0
	for _, o := range b.inspection.Objects {
		if o.NativeName == "editable-list" {
			found = o
			n++
		}
		if strings.HasPrefix(o.NativeName, "editable-list.") {
			t.Fatal("unexpected companion shape", o)
		}
	}
	if n != 1 || found.Kind != "sp" || found.ParentToken != "" {
		t.Fatal("list must be one top-level native text box", found, n)
	}
	return p, b, found
}

func TestNativeEditabilityEditableListActualPackage(t *testing.T) {
	p, b, o := editableListFixture(t)
	paragraphs := nativeParagraphs(o.shape)
	if len(paragraphs) != 3 {
		t.Fatal(paragraphs)
	}
	pkg, e := openLineagePackage(b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	part, e := pkg.read(o.NativePart)
	if e != nil {
		t.Fatal(e)
	}
	span := editableListSpan(t, part)
	x := string(part[span.start:span.end])
	if strings.Count(x, `<a:buChar char=""/>`) != 3 || strings.Count(x, "<a:buClr>") != 3 || strings.Contains(x, "<a:br") || strings.Count(x, "<a:noAutofit/>") != 1 || strings.Contains(x, "<a:normAutofit") || strings.Contains(x, "<a:spAutoFit") {
		t.Fatal(x)
	}
	tx := directXML(o.shape, lineagePML, "txBody")
	body := directXML(tx, drawingML, "bodyPr")
	for _, name := range []string{"lIns", "rIns", "tIns", "bIns"} {
		if attr(body, name) != "0" {
			t.Fatal("list must not inherit text box insets", name, attr(body, name))
		}
	}
	for _, para := range paragraphs {
		if !strings.Contains(strings.Join(para.ReviewItems, ","), "bullet_or_numbered_paragraph") {
			t.Fatal("native bullet lost", para)
		}
	}
	report, e := NativeEditability(b)
	if e != nil {
		t.Fatal(e)
	}
	for _, unit := range report.Objects {
		if unit.ShapeToken == o.ShapeToken && (unit.ParagraphCount != 3 || unit.GroupDepth != 0) {
			t.Fatal(unit)
		}
	}
	if output := os.Getenv("PPTXGENGO_EDITABLE_LIST_PILOT_OUT"); output != "" {
		writeNativeEditabilityPilotFixture(t, p, report, output)
	}
}

func TestReconcileEditableListTopologyRequiresManualReview(t *testing.T) {
	p, b, o := editableListFixture(t)
	before := append([]byte(nil), p.Raw...)
	pkg, e := openLineagePackage(b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	part, e := pkg.read(o.NativePart)
	if e != nil {
		t.Fatal(e)
	}
	span := editableListSpan(t, part)
	shape := part[span.start:span.end]
	var first []byte
	for _, c := range span.children {
		if c.node.Name.Local == "txBody" {
			for _, para := range c.children {
				if para.node.Name.Local == "p" {
					first = part[para.start:para.end]
					break
				}
			}
		}
	}
	if len(first) == 0 {
		t.Fatal("missing paragraph")
	}
	for _, action := range []string{"add", "delete", "reorder", "longer", "format"} {
		t.Run(action, func(t *testing.T) {
			changed := append([]byte(nil), shape...)
			switch action {
			case "add":
				changed = bytes.Replace(changed, []byte("</p:txBody>"), append(append([]byte{}, first...), []byte("</p:txBody>")...), 1)
			case "delete":
				changed = bytes.Replace(changed, first, nil, 1)
			case "reorder":
				changed = bytes.Replace(changed, first, nil, 1)
				changed = bytes.Replace(changed, []byte("</p:txBody>"), append(append([]byte{}, first...), []byte("</p:txBody>")...), 1)
			case "longer":
				changed = bytes.Replace(changed, []byte("Review the source"), []byte("Review the source and add a longer explanation that can wrap onto another line in PowerPoint."), 1)
			case "format":
				changed = bytes.Replace(changed, []byte(`<a:buChar char=""/>`), []byte(`<a:buChar char="•"/>`), 1)
			}
			if bytes.Equal(shape, changed) {
				t.Fatal("fixture mutation did not apply")
			}
			rewritten := append([]byte{}, part[:span.start]...)
			rewritten = append(rewritten, changed...)
			rewritten = append(rewritten, part[span.end:]...)
			edited, e := lineageRewrite(pkg, map[string][]byte{o.NativePart: rewritten})
			if e != nil {
				t.Fatal(e)
			}
			actual, e := InspectNativeLineage(edited, b.Objects)
			if e != nil || len(actual.Issues) != 0 || len(actual.Objects) != len(b.inspection.Objects) {
				t.Fatal("shape identities changed", e, actual.Issues)
			}
			r, e := ReconcileText(p, b, edited)
			if e != nil {
				t.Fatal(e)
			}
			if len(r.ManualReview) == 0 || r.Counts["native_only"] != 0 {
				t.Fatal("list edit silently adopted", r)
			}
			if !bytes.Equal(before, p.Raw) {
				t.Fatal("review mutated source")
			}
		})
	}
}

func editableListSpan(t *testing.T, part []byte) *lineageSpan {
	t.Helper()
	root, e := lineageSpans(part)
	if e != nil {
		t.Fatal(e)
	}
	var found *lineageSpan
	var walk func(*lineageSpan)
	walk = func(s *lineageSpan) {
		if s.node.Name.Space == lineagePML && s.node.Name.Local == "sp" {
			for _, nv := range s.children {
				if nv.node.Name.Local == "nvSpPr" {
					for _, id := range nv.children {
						if id.node.Name.Local == "cNvPr" && attr(id.node, "name") == "editable-list" {
							found = s
						}
					}
				}
			}
		}
		for _, c := range s.children {
			walk(c)
		}
	}
	walk(root)
	if found == nil {
		t.Fatal("missing editable list span")
	}
	return found
}
