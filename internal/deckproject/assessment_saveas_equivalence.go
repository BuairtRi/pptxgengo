package deckproject

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// Assessment-only analysis equivalence for observed PowerPoint table saves.
// It preserves raw evidence and all meaningful fonts, fills, dimensions and
// effective boundary strokes. Unknown extensions/properties remain compared.
func assessmentNativeShell(n *xmlNode) (*xmlNode, error) {
	root := normalizePowerPointSerialization(n)
	var visit func(*xmlNode)
	visit = func(n *xmlNode) {
		children := []*xmlNode{}
		for _, c := range n.Children {
			if c.Name.Local == "extLst" && (n.Name.Local == "gridCol" || n.Name.Local == "tr" || n.Name.Local == "nvPr") && assessmentOfficeAxisIDs(c, n.Name.Local) {
				continue
			}
			visit(c)
			children = append(children, c)
		}
		n.Children = children
		if n.Name.Local != "t" && strings.TrimSpace(n.Text) == "" {
			n.Text = ""
		}
		if n.Name.Space == drawingML && n.Name.Local == "t" {
			attrs := []xml.Attr{}
			for _, a := range n.Attrs {
				if a.Name.Space == "http://www.w3.org/XML/1998/namespace" && a.Name.Local == "space" && a.Value == "preserve" {
					continue
				}
				attrs = append(attrs, a)
			}
			n.Attrs = attrs
		}
		if n.Name.Space == drawingML && n.Name.Local == "p" {
			assessmentNormalizePlainParagraph(n)
		}
		if n.Name.Space == drawingML && n.Name.Local == "tc" {
			assessmentNormalizeCell(n)
		}
	}
	visit(root)
	var tables []*xmlNode
	var find func(*xmlNode)
	find = func(n *xmlNode) {
		if n.Name.Space == drawingML && n.Name.Local == "tbl" {
			tables = append(tables, n)
		}
		for _, c := range n.Children {
			find(c)
		}
	}
	find(root)
	for _, table := range tables {
		rows := [][]*xmlNode{}
		for _, tr := range table.Children {
			if tr.Name.Space == drawingML && tr.Name.Local == "tr" {
				cells := []*xmlNode{}
				for _, tc := range tr.Children {
					if tc.Name.Space == drawingML && tc.Name.Local == "tc" {
						cells = append(cells, tc)
					}
				}
				rows = append(rows, cells)
			}
		}
		for r, cells := range rows {
			for c, tc := range cells {
				if c+1 < len(cells) {
					if e := assessmentNormalizeSharedEdge(tc, "lnR", cells[c+1], "lnL"); e != nil {
						return nil, e
					}
				}
				if r+1 < len(rows) && c < len(rows[r+1]) {
					if e := assessmentNormalizeSharedEdge(tc, "lnB", rows[r+1][c], "lnT"); e != nil {
						return nil, e
					}
				}
			}
		}
	}
	return assessmentTableShell(root), nil
}
func assessmentOfficeAxisIDs(n *xmlNode, parent string) bool {
	expectedSpace := drawingML
	if parent == "nvPr" {
		expectedSpace = lineagePML
	}
	if n.Name.Space != expectedSpace || n.Name.Local != "extLst" {
		return false
	}
	if !noNativeAttributes(n) || strings.TrimSpace(n.Text) != "" || len(n.Children) != 1 {
		return false
	}
	ext := n.Children[0]
	uri, ok := oneNativeAttribute(ext, "uri")
	kind, expected := "colId", "{9D8B030D-6E8A-4147-A177-3AD203B41FA5}"
	if parent == "tr" {
		kind, expected = "rowId", "{0D108BD9-81ED-4DB2-BD59-A6C34878D82A}"
	}
	namespace, extensionSpace := officeCreationIDNamespace, drawingML
	if parent == "nvPr" {
		kind, expected, namespace, extensionSpace = "modId", "{D42A27DB-BD31-4B8C-83A1-F6EECF244321}", "http://schemas.microsoft.com/office/powerpoint/2010/main", lineagePML
	}
	if ext.Name.Space != extensionSpace || ext.Name.Local != "ext" || !ok || uri != expected || strings.TrimSpace(ext.Text) != "" || len(ext.Children) != 1 {
		return false
	}
	id := ext.Children[0]
	value, ok := oneNativeAttribute(id, "val")
	v, e := strconv.ParseUint(value, 10, 32)
	return id.Name.Space == namespace && id.Name.Local == kind && ok && e == nil && strconv.FormatUint(v, 10) == value && len(id.Children) == 0 && strings.TrimSpace(id.Text) == ""
}
func assessmentNormalizeCell(tc *xmlNode) {
	body := assessmentChild(assessmentChild(tc, drawingML, "txBody"), drawingML, "bodyPr")
	props := assessmentChild(tc, drawingML, "tcPr")
	if body != nil && props != nil && len(body.Children) == 1 && body.Children[0].Name == (xml.Name{Space: drawingML, Local: "noAutofit"}) && noNativeAttributes(body.Children[0]) && len(body.Children[0].Children) == 0 && strings.TrimSpace(body.Children[0].Text) == "" {
		expected := map[string]string{"wrap": "square", "anchor": "ctr", "lIns": "0", "rIns": "0", "tIns": "0", "bIns": "0"}
		valid := true
		count := 0
		for _, a := range body.Attrs {
			if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
				continue
			}
			count++
			if a.Name.Space != "" || expected[a.Name.Local] != a.Value {
				valid = false
			}
		}
		if valid && count == len(expected) && lineageAttr(props, "", "anchor") == "ctr" {
			body.Attrs = nil
			body.Children = nil
		}
	}
	if props == nil {
		return
	}
	for _, line := range props.Children {
		if line.Name.Space == drawingML && strings.HasPrefix(line.Name.Local, "ln") {
			assessmentNormalizeLine(line)
		}
	}
}
func assessmentNormalizeLine(n *xmlNode) {
	filtered := []*xmlNode{}
	for _, c := range n.Children {
		defaultChild := false
		if c.Name.Space == drawingML && len(c.Children) == 0 && strings.TrimSpace(c.Text) == "" {
			switch c.Name.Local {
			case "prstDash":
				v, ok := oneNativeAttribute(c, "val")
				defaultChild = ok && v == "solid"
			case "round":
				defaultChild = noNativeAttributes(c)
			case "headEnd", "tailEnd":
				defaultChild = len(c.Attrs) == 3 && lineageAttr(c, "", "type") == "none" && lineageAttr(c, "", "w") == "med" && lineageAttr(c, "", "len") == "med"
			}
		}
		if !defaultChild {
			filtered = append(filtered, c)
		}
	}
	n.Children = filtered
}
func assessmentNormalizeSharedEdge(a *xmlNode, sideA string, b *xmlNode, sideB string) error {
	pa, pb := assessmentChild(a, drawingML, "tcPr"), assessmentChild(b, drawingML, "tcPr")
	la, lb := assessmentChild(pa, drawingML, sideA), assessmentChild(pb, drawingML, sideB)
	if la == nil || lb == nil {
		return nil
	}
	score := func(n *xmlNode) (int64, bool) {
		if strings.TrimSpace(n.Text) != "" || len(n.Children) != 1 {
			return 0, false
		}
		for _, a := range n.Attrs {
			if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
				continue
			}
			if a.Name.Space != "" {
				return 0, false
			}
			switch a.Name.Local {
			case "w":
			case "cap":
				if a.Value != "flat" {
					return 0, false
				}
			case "cmpd":
				if a.Value != "sng" {
					return 0, false
				}
			case "algn":
				if a.Value != "ctr" {
					return 0, false
				}
			default:
				return 0, false
			}
		}
		fill := n.Children[0]
		if fill.Name.Space != drawingML || !noNativeAttributes(fill) || strings.TrimSpace(fill.Text) != "" {
			return 0, false
		}
		if fill.Name.Local == "noFill" && len(fill.Children) == 0 {
			return 0, true
		}
		if fill.Name.Local != "solidFill" || len(fill.Children) != 1 {
			return 0, false
		}
		color := fill.Children[0]
		value, ok := oneNativeAttribute(color, "val")
		if color.Name.Space != drawingML || color.Name.Local != "srgbClr" || !ok || len(value) != 6 || len(color.Children) != 0 || strings.TrimSpace(color.Text) != "" {
			return 0, false
		}
		if _, e := strconv.ParseUint(value, 16, 24); e != nil {
			return 0, false
		}
		w, e := strconv.ParseInt(lineageAttr(n, "", "w"), 10, 64)
		return w, e == nil && w > 0 && w <= 2147483647
	}
	wa, oka := score(la)
	wb, okb := score(lb)
	if !oka || !okb {
		return nil
	}
	winner := la
	if wb > wa {
		winner = lb
	}
	if wa == wb {
		ca, cb := *la, *lb
		ca.Name.Local = "line"
		cb.Name.Local = "line"
		if nativeViewMismatch(reconcileStructureView(&ca), reconcileStructureView(&cb), "/shared-edge") != "" {
			return fmt.Errorf("assessment native shared border has ambiguous equal-width styles")
		}
	}
	replace := func(props *xmlNode, side string) {
		for i, c := range props.Children {
			if c.Name.Space == drawingML && c.Name.Local == side {
				copyLine := *winner
				copyLine.Name.Local = side
				props.Children[i] = &copyLine
			}
		}
	}
	replace(pa, sideA)
	replace(pb, sideB)
	return nil
}

func assessmentChild(n *xmlNode, space, local string) *xmlNode {
	if n == nil {
		return nil
	}
	return lineageChild(n, space, local)
}
func assessmentNormalizePlainParagraph(p *xmlNode) {
	defaults := assessmentChild(assessmentChild(p, drawingML, "pPr"), drawingML, "defRPr")
	if defaults == nil {
		return
	}
	same := func(properties *xmlNode) bool {
		if properties == nil {
			return false
		}
		copyProps := *properties
		copyProps.Name = defaults.Name
		return nativeViewMismatch(reconcileStructureView(defaults), reconcileStructureView(&copyProps), "/paragraph-properties") == ""
	}
	runs, ends := 0, 0
	for _, c := range p.Children {
		if c.Name.Space != drawingML {
			return
		}
		switch c.Name.Local {
		case "pPr":
		case "r":
			runs++
			if !noNativeAttributes(c) || len(c.Children) != 2 || c.Children[0].Name != (xml.Name{Space: drawingML, Local: "rPr"}) || c.Children[1].Name != (xml.Name{Space: drawingML, Local: "t"}) || len(c.Children[1].Children) != 0 || !noNativeAttributes(c.Children[1]) || !same(c.Children[0]) {
				return
			}
		case "endParaRPr":
			ends++
			if !same(c) {
				return
			}
		default:
			return
		}
	}
	if runs > 1 || ends > 1 {
		return
	}
	children := []*xmlNode{}
	for _, c := range p.Children {
		if c.Name.Local == "r" || c.Name.Local == "endParaRPr" {
			continue
		}
		children = append(children, c)
	}
	p.Children = children
}
