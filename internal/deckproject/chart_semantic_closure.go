package deckproject

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// Numeric adoption has a separate closure from visual adoption. Office may
// rename an embedded part or rewrite workbook views and chart styles. Those
// changes remain manual findings and retained evidence, never source styling.
func verifyChartSemanticClosure(a, b nativeChartFacts) error {
	if !bytes.Equal(canonical(a.refs), canonical(b.refs)) || !bytes.Equal(canonical(a.names), canonical(b.names)) || !bytes.Equal(canonical(a.categories), canonical(b.categories)) {
		return fmt.Errorf("native chart keyed axes/series/reference remapping refused")
	}
	if !bytes.Equal(canonical(chartBusinessIdentity(a.tree)), canonical(chartBusinessIdentity(b.tree))) {
		return fmt.Errorf("native chart type, axes or series order changed")
	}
	_, ac, as, e := workbookCells(a.book)
	if e != nil {
		return e
	}
	_, bc, bs, e := workbookCells(b.book)
	if e != nil {
		return e
	}
	for ref, cell := range ac {
		if a.numericCells[ref] {
			continue
		}
		other := bc[ref]
		if other == nil {
			return fmt.Errorf("workbook unrelated cell removed at %s", ref)
		}
		av, ae := workbookText(cell, as)
		bv, be := workbookText(other, bs)
		if ae != nil || be != nil || av != bv {
			return fmt.Errorf("workbook unrelated cell changed at %s", ref)
		}
	}
	for ref := range bc {
		if ac[ref] == nil && !a.numericCells[ref] {
			return fmt.Errorf("extra workbook cell at %s", ref)
		}
	}
	for _, f := range []nativeChartFacts{a, b} {
		if e := verifySingleChartWorksheet(f.book); e != nil {
			return e
		}
	}
	// Tables are business structure when actually referenced by the worksheet.
	// Old generated workbooks also contain an orphan table; Excel removes that
	// unreferenced part on Save As. It is not a plotted data range.
	at, e := referencedChartTables(a)
	if e != nil {
		return e
	}
	bt, e := referencedChartTables(b)
	if e != nil {
		return e
	}
	if !bytes.Equal(canonical(at), canonical(bt)) {
		return fmt.Errorf("workbook table/range changed")
	}
	return nil
}

func chartBusinessIdentity(tree *xmlNode) []any {
	out := []any{}
	for _, area := range chartNodes(tree, "plotArea") {
		for _, c := range area.Children {
			if c.Name.Space != chartML {
				continue
			}
			if strings.HasSuffix(c.Name.Local, "Chart") {
				out = append(out, c.Name.Local)
				for _, name := range []string{"barDir", "grouping", "scatterStyle", "axId"} {
					for _, n := range c.Children {
						if n.Name.Space == chartML && n.Name.Local == name {
							out = append(out, []string{name, attr(n, "val")})
						}
					}
				}
				for _, s := range chartNodes(c, "ser") {
					for _, name := range []string{"idx", "order"} {
						n := chartChild(s, name)
						if n == nil {
							out = append(out, []string{name, "missing"})
						} else {
							out = append(out, []string{name, attr(n, "val")})
						}
					}
				}
			} else if c.Name.Local == "catAx" || c.Name.Local == "valAx" || c.Name.Local == "dateAx" || c.Name.Local == "serAx" {
				out = append(out, []string{c.Name.Local, chartBusinessValue(chartChild(c, "axId")), chartBusinessValue(chartChild(c, "crossAx"))})
			}
		}
	}
	return out
}

func chartBusinessValue(n *xmlNode) string {
	if n == nil {
		return "missing"
	}
	return attr(n, "val")
}

func verifySingleChartWorksheet(book *lineagePackage) error {
	for name := range book.files {
		if strings.HasPrefix(name, "xl/externalLinks/") {
			return fmt.Errorf("external workbook links refused")
		}
		if strings.HasPrefix(name, "xl/worksheets/") && strings.HasSuffix(name, ".xml") && name != "xl/worksheets/sheet1.xml" {
			return fmt.Errorf("additional workbook worksheet refused")
		}
	}
	tree, e := book.tree("xl/workbook.xml")
	if e != nil {
		return e
	}
	sheets := descendants(tree, "sheet")
	if len(sheets) != 1 || sheets[0].Name.Space != sheetML || attr(sheets[0], "name") != "Sheet1" {
		return fmt.Errorf("chart worksheet identity changed")
	}
	rels, e := book.relationships("xl/workbook.xml")
	if e != nil {
		return e
	}
	r, ok := rels[attr(sheets[0], "id")]
	if !ok || !strings.HasSuffix(r.Type, "/worksheet") || r.Mode != "" && r.Mode != "Internal" {
		return fmt.Errorf("chart worksheet binding changed")
	}
	target, e := chartPartTarget("xl/workbook.xml", r.Target)
	if e != nil || target != "xl/worksheets/sheet1.xml" {
		return fmt.Errorf("chart worksheet binding changed")
	}
	for _, r := range rels {
		if r.Mode != "" && r.Mode != "Internal" {
			return fmt.Errorf("external workbook relationship refused")
		}
	}
	return nil
}

func referencedChartTables(f nativeChartFacts) ([]any, error) {
	out := []any{}
	parts := descendants(f.sheet, "tablePart")
	if len(parts) == 0 {
		return out, nil
	}
	rels, e := f.book.relationships("xl/worksheets/sheet1.xml")
	if e != nil {
		return nil, e
	}
	for _, n := range parts {
		if n.Name.Space != sheetML {
			return nil, fmt.Errorf("unknown worksheet table binding")
		}
		r, ok := rels[attr(n, "id")]
		if !ok || !strings.HasSuffix(r.Type, "/table") || r.Mode != "" && r.Mode != "Internal" {
			return nil, fmt.Errorf("external/unknown workbook table refused")
		}
		target, e := chartPartTarget("xl/worksheets/sheet1.xml", r.Target)
		if e != nil {
			return nil, e
		}
		t, e := f.book.tree(target)
		if e != nil {
			return nil, e
		}
		cols := []string{}
		for _, c := range descendants(t, "tableColumn") {
			cols = append(cols, attr(c, "name"))
		}
		out = append(out, []any{attr(t, "ref"), attr(t, "name"), attr(t, "displayName"), cols})
	}
	sort.Slice(out, func(i, j int) bool { return string(canonical(out[i])) < string(canonical(out[j])) })
	return out, nil
}
