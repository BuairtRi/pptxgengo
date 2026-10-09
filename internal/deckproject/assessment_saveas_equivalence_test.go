package deckproject

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

func TestAssessmentOfficeSaveAsQualifiedSpecimen(t *testing.T) {
	baseline, edited := os.Getenv("PPTXGENGO_ASSESSMENT_OFFICE_BASELINE"), os.Getenv("PPTXGENGO_ASSESSMENT_OFFICE_EDITED")
	if baseline == "" || edited == "" {
		t.Skip("named native SaveAs specimen not requested")
	}
	read := func(path string) *xmlNode {
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		objects, _ := geometryFromPackage(t, raw)
		return objects["heat.matrix.native"].node
	}
	a, e := assessmentNativeShell(read(baseline))
	if e != nil {
		t.Fatal(e)
	}
	b, e := assessmentNativeShell(read(edited))
	if e != nil {
		t.Fatal(e)
	}
	if mismatch := nativeViewMismatch(reconcileStructureView(a), reconcileStructureView(b), "/graphicFrame"); mismatch != "" {
		t.Fatal(mismatch)
	}
}

func TestAssessmentSerializationEquivalentDefaultsAndStyleGuards(t *testing.T) {
	_, b, id := assessmentSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	source := objects[id+".matrix.native"].node
	original := digest(canonical(source))
	baseline, e := assessmentNativeShell(source)
	if e != nil {
		t.Fatal(e)
	}
	copy := normalizePowerPointSerialization(source)
	var rows []*xmlNode
	var walk func(*xmlNode)
	walk = func(n *xmlNode) {
		if n.Name.Space == drawingML && n.Name.Local == "tr" {
			rows = append(rows, n)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(copy)
	cells := func(row *xmlNode) []*xmlNode {
		out := []*xmlNode{}
		for _, c := range row.Children {
			if c.Name.Space == drawingML && c.Name.Local == "tc" {
				out = append(out, c)
			}
		}
		return out
	}
	header, body := cells(rows[0])[0], cells(rows[1])[0]
	stroke := *assessmentChild(assessmentChild(header, drawingML, "tcPr"), drawingML, "lnB")
	stroke.Name.Local = "lnT"
	props := assessmentChild(body, drawingML, "tcPr")
	for i, c := range props.Children {
		if c.Name.Local == "lnT" {
			props.Children[i] = &stroke
		}
	}
	normalized, e := assessmentNativeShell(copy)
	if e != nil {
		t.Fatal(e)
	}
	if mismatch := nativeViewMismatch(reconcileStructureView(baseline), reconcileStructureView(normalized), "/frame"); mismatch != "" {
		t.Fatal("shared effective stroke duplication changed view", mismatch)
	}
	if digest(canonical(source)) != original {
		t.Fatal("analysis changed retained raw input")
	}
	for _, mutate := range []func(*xmlNode){
		func(n *xmlNode) {
			for i, a := range n.Attrs {
				if a.Name.Local == "w" {
					n.Attrs = append([]xml.Attr{}, n.Attrs...)
					n.Attrs[i].Value = "99999"
				}
			}
		},
		func(n *xmlNode) {
			fill := assessmentChild(n, drawingML, "solidFill")
			color := assessmentChild(fill, drawingML, "srgbClr")
			if color != nil {
				color.Attrs = append([]xml.Attr{}, color.Attrs...)
				color.Attrs[0].Value = "FF0000"
			}
		},
	} {
		changed := normalizePowerPointSerialization(source)
		rows = nil
		walk(changed)
		edge := assessmentChild(assessmentChild(cells(rows[0])[0], drawingML, "tcPr"), drawingML, "lnB")
		mutate(edge)
		view, e := assessmentNativeShell(changed)
		if e == nil && nativeViewMismatch(reconcileStructureView(baseline), reconcileStructureView(view), "/frame") == "" {
			t.Fatal("meaningful visible stroke change was ignored")
		}
	}
	changed := normalizePowerPointSerialization(source)
	rows = nil
	walk(changed)
	unknownEdge := assessmentChild(assessmentChild(cells(rows[1])[0], drawingML, "tcPr"), drawingML, "lnT")
	unknownEdge.Children = append(unknownEdge.Children, &xmlNode{Name: xml.Name{Space: "urn:unknown", Local: "payload"}})
	view, e := assessmentNativeShell(changed)
	if e == nil && nativeViewMismatch(reconcileStructureView(baseline), reconcileStructureView(view), "/frame") == "" {
		t.Fatal("unknown metadata on weaker shared edge was discarded")
	}
}

func TestAssessmentOfficeAxisIDExtensionsAreClosed(t *testing.T) {
	const extension = `<a:extLst xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:a16="http://schemas.microsoft.com/office/drawing/2014/main"><a:ext uri="{9D8B030D-6E8A-4147-A177-3AD203B41FA5}"><a16:colId val="20001"/></a:ext></a:extLst>`
	parse := func(raw string) *xmlNode {
		n, e := readLineageXML([]byte(raw))
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	if !assessmentOfficeAxisIDs(parse(extension), "gridCol") {
		t.Fatal("observed native column metadata rejected")
	}
	for _, raw := range []string{strings.Replace(extension, "20001", "-1", 1), strings.Replace(extension, "20001", "020001", 1), strings.Replace(extension, "colId", "rowId", 1), strings.Replace(extension, "20001\"", "20001\" unrecognized=\"yes\"", 1), strings.Replace(extension, "9D8B030D", "8D8B030D", 1), strings.Replace(extension, "http://schemas.openxmlformats.org/drawingml/2006/main", "urn:unknown", 1)} {
		if assessmentOfficeAxisIDs(parse(raw), "gridCol") {
			t.Fatal("unknown/malformed column metadata hidden")
		}
	}
}
