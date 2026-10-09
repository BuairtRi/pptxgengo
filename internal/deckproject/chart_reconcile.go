package deckproject

import (
	"bytes"
	"fmt"
	"math"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const QuantitativeSemanticDecisionsSchema = "pptxgengo.quantitative-semantic-decisions.v1"
const chartML = "http://schemas.openxmlformats.org/drawingml/2006/chart"
const sheetML = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"

type QuantitativeSemanticProposal struct {
	ID       string   `json:"id"`
	Entity   string   `json:"entity"`
	Category string   `json:"category,omitempty"`
	Series   string   `json:"series,omitempty"`
	Point    string   `json:"point,omitempty"`
	Axis     string   `json:"axis,omitempty"`
	Before   *float64 `json:"before"`
	Proposed *float64 `json:"proposed"`
	Status   string   `json:"status"`
	Reason   string   `json:"reason"`
}
type QuantitativeSemanticReport struct {
	Schema                 string                         `json:"schema"`
	ProjectID              string                         `json:"project_id"`
	SlideID                string                         `json:"slide_id"`
	NodeID                 string                         `json:"node_id"`
	SourceSHA256           string                         `json:"source_sha256"`
	GeometryReportSHA256   string                         `json:"geometry_report_sha256"`
	EditedPPTXSHA256       string                         `json:"edited_pptx_sha256"`
	ChartPart              string                         `json:"chart_part"`
	WorkbookPart           string                         `json:"workbook_part"`
	Policy                 string                         `json:"policy"`
	Proposals              []QuantitativeSemanticProposal `json:"proposals"`
	ManualReview           []TextReconciliationIssue      `json:"manual_review"`
	UnresolvedGeometryIDs  []string                       `json:"unresolved_geometry_ids"`
	UnresolvedTextIDs      []string                       `json:"unresolved_text_ids"`
	UnresolvedStructureIDs []string                       `json:"unresolved_structure_ids"`
}

func QuantitativeSemanticReportHash(r QuantitativeSemanticReport) string { return digest(canonical(r)) }

type nativeChartFacts struct {
	part, workbook string
	tree, sheet    *xmlNode
	data           [][]*float64
	names          []string
	categories     []string
	refs           [][]string
	numericCells   map[string]bool
	strings        []string
	book           *lineagePackage
}

func chartChild(n *xmlNode, name string) *xmlNode {
	if n == nil {
		return nil
	}
	return directXML(n, chartML, name)
}
func chartNodes(n *xmlNode, name string) []*xmlNode {
	out := []*xmlNode{}
	for _, child := range descendants(n, name) {
		if child.Name.Space == chartML {
			out = append(out, child)
		}
	}
	return out
}

var chartRange = regexp.MustCompile(`^(?:Sheet1|'Sheet1')!\$([A-Z]{1,3})\$([0-9]{1,4})(?::\$([A-Z]{1,3})\$([0-9]{1,4}))?$`)

func chartReferences(formula string, count int) ([]string, error) {
	m := chartRange.FindStringSubmatch(formula)
	if m == nil {
		return nil, fmt.Errorf("native chart uses unsupported range %q", formula)
	}
	first, e := strconv.Atoi(m[2])
	if e != nil {
		return nil, e
	}
	last := first
	if m[4] != "" {
		last, e = strconv.Atoi(m[4])
		if e != nil {
			return nil, e
		}
		if m[1] != m[3] {
			return nil, fmt.Errorf("chart range spans columns")
		}
	}
	if first < 1 || last-first+1 != count || count < 1 || count > 60 {
		return nil, fmt.Errorf("native chart reference count does not match keyed source")
	}
	out := []string{}
	for i := first; i <= last; i++ {
		out = append(out, m[1]+strconv.Itoa(i))
	}
	return out, nil
}
func nativeChartCache(ref *xmlNode, count int, numeric bool) ([]*float64, []string, []string, error) {
	var number []*float64
	var text []string
	cacheName := "strCache"
	if numeric {
		cacheName = "numCache"
	}
	cache := chartChild(ref, cacheName)
	if !numeric && ref != nil && ref.Name.Local == "multiLvlStrRef" {
		cache = chartChild(ref, "multiLvlStrCache")
	}
	formula := chartChild(ref, "f")
	if cache == nil || formula == nil {
		return nil, nil, nil, fmt.Errorf("native chart requires cached reference and formula")
	}
	counts := map[string]int{}
	for _, n := range ref.Children {
		if n.Name.Space == chartML {
			counts[n.Name.Local]++
		}
	}
	if counts["f"] != 1 || counts[cache.Name.Local] != 1 {
		return nil, nil, nil, fmt.Errorf("native chart duplicate reference/cache refused")
	}
	ptCounts := 0
	for _, n := range cache.Children {
		if n.Name.Space == chartML && n.Name.Local == "ptCount" {
			ptCounts++
		}
	}
	if ptCounts != 1 {
		return nil, nil, nil, fmt.Errorf("native chart duplicate/missing point count")
	}
	pts := chartChild(cache, "ptCount")
	if pts == nil || attr(pts, "val") != strconv.Itoa(count) {
		return nil, nil, nil, fmt.Errorf("native chart cache count mismatch")
	}
	refs, e := chartReferences(formula.Text, count)
	if e != nil {
		return nil, nil, nil, e
	}
	number = make([]*float64, count)
	text = make([]string, count)
	seen := map[int]bool{}
	children := cache.Children
	if cache.Name.Local == "multiLvlStrCache" {
		levels := chartNodes(cache, "lvl")
		if len(levels) != 1 {
			return nil, nil, nil, fmt.Errorf("multi-level category axes require explicit mapping")
		}
		children = levels[0].Children
	}
	for _, p := range children {
		if p.Name.Space != chartML || p.Name.Local != "pt" {
			continue
		}
		i, e := strconv.Atoi(attr(p, "idx"))
		if e != nil || i < 0 || i >= count || seen[i] {
			return nil, nil, nil, fmt.Errorf("native chart duplicate/outside cache index")
		}
		seen[i] = true
		v := chartChild(p, "v")
		if v == nil {
			return nil, nil, nil, fmt.Errorf("native cache value missing")
		}
		if numeric {
			f, e := strconv.ParseFloat(v.Text, 64)
			if e != nil || math.IsNaN(f) || math.IsInf(f, 0) || math.Abs(f) > 1e9 {
				return nil, nil, nil, fmt.Errorf("native chart number invalid")
			}
			number[i] = &f
		} else {
			text[i] = v.Text
		}
	}
	if !numeric && len(seen) != count {
		return nil, nil, nil, fmt.Errorf("native chart category/name cache incomplete")
	}
	return number, text, refs, nil
}
func workbookCells(book *lineagePackage) (*xmlNode, map[string]*xmlNode, []string, error) {
	sheet, e := book.tree("xl/worksheets/sheet1.xml")
	if e != nil {
		return nil, nil, nil, e
	}
	cells := map[string]*xmlNode{}
	for _, n := range descendants(sheet, "c") {
		if n.Name.Space != sheetML {
			continue
		}
		ref := attr(n, "r")
		if ref == "" || cells[ref] != nil {
			return nil, nil, nil, fmt.Errorf("workbook cell missing/duplicate identity")
		}
		if len(descendants(n, "f")) > 0 {
			return nil, nil, nil, fmt.Errorf("formula workbooks require explicit source model")
		}
		cells[ref] = n
	}
	stringsOut := []string{}
	if book.files["xl/sharedStrings.xml"] != nil {
		tree, e := book.tree("xl/sharedStrings.xml")
		if e != nil {
			return nil, nil, nil, e
		}
		for _, si := range descendants(tree, "si") {
			v := ""
			for _, n := range descendants(si, "t") {
				v += n.Text
			}
			stringsOut = append(stringsOut, v)
		}
	}
	return sheet, cells, stringsOut, nil
}
func workbookText(cell *xmlNode, shared []string) (string, error) {
	if cell == nil {
		return "", nil
	}
	if attr(cell, "t") == "inlineStr" {
		out := ""
		for _, n := range descendants(cell, "t") {
			out += n.Text
		}
		return out, nil
	}
	vs := descendants(cell, "v")
	if len(vs) != 1 {
		return "", fmt.Errorf("workbook string cell has no unique value")
	}
	v := vs[0].Text
	if attr(cell, "t") == "s" {
		i, e := strconv.Atoi(v)
		if e != nil || i < 0 || i >= len(shared) {
			return "", fmt.Errorf("workbook shared string invalid")
		}
		return shared[i], nil
	}
	return v, nil
}
func workbookNumber(cell *xmlNode) (*float64, error) {
	if cell == nil {
		return nil, nil
	}
	typ := attr(cell, "t")
	if typ != "" && typ != "n" {
		return nil, fmt.Errorf("native numeric workbook cell requires numeric type")
	}
	vs := descendants(cell, "v")
	if len(vs) == 0 {
		return nil, nil
	}
	if len(vs) != 1 {
		return nil, fmt.Errorf("duplicate workbook value")
	}
	if strings.TrimSpace(vs[0].Text) == "" {
		return nil, nil
	}
	v, e := strconv.ParseFloat(vs[0].Text, 64)
	if e != nil || math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e9 {
		return nil, fmt.Errorf("workbook numerical value invalid")
	}
	return &v, nil
}
func sameChartNumber(a, b *float64) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}
func chartPartTarget(part, target string) (string, error) {
	if strings.HasPrefix(target, "/") {
		name := strings.TrimPrefix(target, "/")
		if name == "" || strings.HasPrefix(name, "/") || strings.ContainsAny(name, "\\:\x00") || path.Clean(name) != name || strings.HasPrefix(name, "../") {
			return "", fmt.Errorf("unsafe absolute OPC chart target")
		}
		return name, nil
	}
	return lineageTarget(part, target)
}
func readNativeChartFacts(data []byte, object NativeLineageObject, categories, series int, scatter bool) (nativeChartFacts, error) {
	out := nativeChartFacts{numericCells: map[string]bool{}}
	pkg, e := openLineagePackage(data)
	if e != nil {
		return out, e
	}
	refs := chartNodes(object.shape, "chart")
	if len(refs) != 1 {
		return out, fmt.Errorf("native graphic frame requires one chart")
	}
	id := attr(refs[0], "id")
	rels, e := pkg.relationships(object.NativePart)
	if e != nil {
		return out, e
	}
	rel, ok := rels[id]
	if !ok || (rel.Mode != "" && rel.Mode != "Internal") || !strings.HasSuffix(rel.Type, "/chart") {
		return out, fmt.Errorf("invalid native chart relationship")
	}
	out.part, e = chartPartTarget(object.NativePart, rel.Target)
	if e != nil {
		return out, e
	}
	out.tree, e = pkg.tree(out.part)
	if e != nil {
		return out, e
	}
	external := chartNodes(out.tree, "externalData")
	if len(external) != 1 {
		return out, fmt.Errorf("native chart requires exactly one embedded workbook")
	}
	cr, e := pkg.relationships(out.part)
	if e != nil {
		return out, e
	}
	wr, ok := cr[attr(external[0], "id")]
	if !ok || (wr.Mode != "" && wr.Mode != "Internal") || !strings.HasSuffix(wr.Type, "/package") {
		return out, fmt.Errorf("external or ambiguous workbook refused")
	}
	out.workbook, e = chartPartTarget(out.part, wr.Target)
	if e != nil {
		return out, e
	}
	raw, e := pkg.read(out.workbook)
	if e != nil {
		return out, e
	}
	out.book, e = openLineagePackage(raw)
	if e != nil {
		return out, e
	}
	sheet, cells, shared, e := workbookCells(out.book)
	if e != nil {
		return out, e
	}
	out.sheet, out.strings = sheet, shared
	ss := chartNodes(out.tree, "ser")
	if len(ss) != series {
		return out, fmt.Errorf("native series count changed")
	}
	for _, s := range ss {
		if scatter {
			for _, axis := range []string{"xVal", "yVal"} {
				a := chartChild(s, axis)
				if a == nil {
					return out, fmt.Errorf("scatter axis missing")
				}
				nums, _, refs, e := nativeChartCache(chartChild(a, "numRef"), categories, true)
				if e != nil {
					return out, e
				}
				out.data = append(out.data, nums)
				out.refs = append(out.refs, refs)
				for i, ref := range refs {
					v, e := workbookNumber(cells[ref])
					if e != nil || !sameChartNumber(v, nums[i]) {
						return out, nativeChartNumericMismatch("scatter", ref, nums[i], v, e)
					}
					out.numericCells[ref] = true
				}
			}
			continue
		}
		tx := chartChild(s, "tx")
		if tx == nil {
			return out, fmt.Errorf("native series name missing")
		}
		_, names, nrefs, e := nativeChartCache(chartChild(tx, "strRef"), 1, false)
		if e != nil {
			return out, e
		}
		v, e := workbookText(cells[nrefs[0]], shared)
		if e != nil || v != names[0] {
			return out, fmt.Errorf("native series name cache/workbook disagree")
		}
		out.names = append(out.names, names[0])
		cat := chartChild(s, "cat")
		if cat == nil {
			return out, fmt.Errorf("native categories missing")
		}
		catRef := chartChild(cat, "strRef")
		if catRef == nil {
			catRef = chartChild(cat, "multiLvlStrRef")
		}
		_, labels, crefs, e := nativeChartCache(catRef, categories, false)
		if e != nil {
			return out, e
		}
		for i, ref := range crefs {
			v, e := workbookText(cells[ref], shared)
			if e != nil || v != labels[i] {
				return out, fmt.Errorf("category cache/workbook disagree at %s", ref)
			}
		}
		if out.categories == nil {
			out.categories = labels
		} else if !bytes.Equal(canonical(out.categories), canonical(labels)) {
			return out, fmt.Errorf("native series category axes disagree")
		}
		vals := chartChild(s, "val")
		if vals == nil {
			return out, fmt.Errorf("native series values missing")
		}
		nums, _, vrefs, e := nativeChartCache(chartChild(vals, "numRef"), categories, true)
		if e != nil {
			return out, e
		}
		for i, ref := range vrefs {
			v, e := workbookNumber(cells[ref])
			if e != nil || !sameChartNumber(v, nums[i]) {
				return out, nativeChartNumericMismatch("chart", ref, nums[i], v, e)
			}
			out.numericCells[ref] = true
		}
		out.data = append(out.data, nums)
		out.refs = append(out.refs, vrefs)
	}
	return out, nil
}

func nativeChartNumericMismatch(kind, ref string, cache, cell *float64, cause error) error {
	value := func(n *float64) string {
		if n == nil {
			return "missing"
		}
		return strconv.FormatFloat(*n, 'g', -1, 64)
	}
	if cause != nil {
		return fmt.Errorf("%s cache/workbook disagree at %s: invalid workbook cell: %w", kind, ref, cause)
	}
	return fmt.Errorf("%s cache/workbook disagree at %s: cached=%s workbook=%s; preserve failed save and explicitly repair source or diagnose Office refresh before adoption", kind, ref, value(cache), value(cell))
}
func chartNumericShell(n *xmlNode) *xmlNode {
	if n == nil {
		return nil
	}
	clone := *n
	clone.Children = nil
	for _, c := range n.Children {
		if n.Name.Space == chartML && n.Name.Local == "numCache" && c.Name.Space == chartML && c.Name.Local == "pt" {
			continue
		}
		clone.Children = append(clone.Children, chartNumericShell(c))
	}
	if n.Name.Space == chartML && n.Name.Local == "f" {
		clone.Text = strings.ReplaceAll(clone.Text, "'Sheet1'!", "Sheet1!")
	}
	return &clone
}
func workbookNumericShell(n *xmlNode, cells map[string]bool) *xmlNode {
	if n == nil {
		return nil
	}
	clone := *n
	clone.Children = nil
	for _, c := range n.Children {
		if c.Name.Space == sheetML && c.Name.Local == "c" && cells[attr(c, "r")] {
			continue
		}
		clone.Children = append(clone.Children, workbookNumericShell(c, cells))
	}
	return &clone
}
func verifyChartShell(a, b nativeChartFacts) error {
	if a.part != b.part || a.workbook != b.workbook || !bytes.Equal(canonical(a.refs), canonical(b.refs)) {
		return fmt.Errorf("chart/workbook relationship or reference remapping refused")
	}
	if nativeViewMismatch(reconcileStructureView(chartNumericShell(a.tree)), reconcileStructureView(chartNumericShell(b.tree)), "/chart") != "" {
		return fmt.Errorf("chart style, axes, annotations or structure changed; preserve for manual review")
	}
	if nativeViewMismatch(reconcileStructureView(workbookNumericShell(a.sheet, a.numericCells)), reconcileStructureView(workbookNumericShell(b.sheet, b.numericCells)), "/worksheet") != "" {
		return fmt.Errorf("workbook unrelated cells/structure changed")
	}
	// Numeric cell styles must remain unchanged; deleting a cell means missing,
	// not permission to discard arbitrary existing formatting.
	_, ac, _, e := workbookCells(a.book)
	if e != nil {
		return e
	}
	_, bc, _, e := workbookCells(b.book)
	if e != nil {
		return e
	}
	for ref := range a.numericCells {
		if ac[ref] != nil && bc[ref] != nil {
			ax, bx := *ac[ref], *bc[ref]
			ax.Children, bx.Children = nil, nil
			ax.Text, bx.Text = "", ""
			if nativeViewMismatch(reconcileStructureView(&ax), reconcileStructureView(&bx), "/cell") != "" {
				return fmt.Errorf("numeric cell style/type changed")
			}
		}
	}
	for _, name := range []string{"xl/tables/table1.xml", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/worksheets/_rels/sheet1.xml.rels"} {
		if a.book.files[name] == nil && b.book.files[name] == nil {
			continue
		}
		at, e := a.book.tree(name)
		if e != nil {
			return e
		}
		bt, e := b.book.tree(name)
		if e != nil {
			return e
		}
		if nativeViewMismatch(reconcileStructureView(at), reconcileStructureView(bt), "/"+name) != "" {
			return fmt.Errorf("workbook table/range/relationship changed")
		}
	}
	return nil
}
func ProposeQuantitativeSemantics(p *Project, packet *TextReviewPacket, slide, node, bundle, engine string) (QuantitativeSemanticReport, error) {
	out := QuantitativeSemanticReport{Schema: "pptxgengo.quantitative-semantic-report.v1", ProjectID: p.Document.ID, SlideID: slide, NodeID: node, SourceSHA256: p.SourceHash(), Policy: "authenticated native chart and unchanged geometry/keyed axes/series/source units; strict cache and embedded workbook cell agreement; explicit numeric-only adoption; formatting differences remain manual with edited package retained and source formatting unchanged; never geometry inference or visual-equivalence claim", Proposals: []QuantitativeSemanticProposal{}, UnresolvedGeometryIDs: []string{}, UnresolvedTextIDs: []string{}, UnresolvedStructureIDs: []string{}}
	packet, baseline, e := verifiedGanttPacket(p, packet, bundle, engine)
	if e != nil {
		return out, e
	}
	out.GeometryReportSHA256 = packet.ReportSHA256
	out.EditedPPTXSHA256 = packet.Report.EditedPPTXSHA256
	out.ManualReview = append(out.ManualReview, packet.Report.ManualReview...)
	idx, t, e := diagramSlide(p, slide)
	if e != nil {
		return out, e
	}
	n, e := quantitativeNode(&t, node)
	if e != nil {
		return out, e
	}
	source, keys, bound, e := quantitativeSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound || len(p.Document.Slides[idx].NativeGeometry) > 0 || len(p.Document.Slides[idx].NativeOrder) > 0 {
		return out, fmt.Errorf("chart semantic review requires materialized source and clean fresh baseline")
	}
	kind, _ := source["kind"].(string)
	if kind == "quadrant" || source["progress"] != nil {
		return out, fmt.Errorf("qualitative/progress charts require explicit source interpretation, not cached series adoption")
	}
	if kind != "scatter" && source["preserveWorkbookZeros"] != true {
		return out, fmt.Errorf("chart semantic review requires preserveWorkbookZeros:true and fresh baseline to distinguish zero from missing")
	}
	original, e := InspectNativeLineage(baseline.files["deck.pptx"], baseline.Objects)
	if e != nil {
		return out, e
	}
	edited, e := InspectNativeLineage(packet.Edited, baseline.Objects)
	if e != nil {
		return out, e
	}
	name := ganttNativeNodeID(t.Nodes, node, "") + ".native"
	var before, after *NativeLineageObject
	for i, o := range original.Objects {
		if o.NativeName == name && baseline.Objects.Lineage.Slides[o.SlideToken] == slide {
			if before != nil {
				return out, fmt.Errorf("ambiguous native chart")
			}
			before = &original.Objects[i]
		}
	}
	if before == nil || before.Kind != "graphicFrame" {
		return out, fmt.Errorf("authenticated native chart missing")
	}
	for i, o := range edited.Objects {
		if o.ShapeToken == before.ShapeToken {
			if after != nil {
				return out, fmt.Errorf("copied chart identity ambiguous")
			}
			after = &edited.Objects[i]
		}
	}
	if after == nil || after.NativeName != before.NativeName || after.ParentToken != before.ParentToken || after.Kind != before.Kind {
		return out, fmt.Errorf("native chart removed/renamed/reparented")
	}
	for _, issue := range edited.Issues {
		if issue.ShapeToken == before.ShapeToken || issue.SlideToken == before.SlideToken && issue.ShapeToken == "" {
			return out, fmt.Errorf("native chart lineage issue: %s", issue.Kind)
		}
	}
	count, series := len(keys["categories"]), len(keys["series"])
	if kind == "scatter" {
		count, series = len(keys["points"]), 1
	}
	a, e := readNativeChartFacts(baseline.files["deck.pptx"], *before, count, series, kind == "scatter")
	if e != nil {
		return out, e
	}
	b, e := readNativeChartFacts(packet.Edited, *after, count, series, kind == "scatter")
	if e != nil {
		return out, e
	}
	// Read both caches and workbooks before comparing Office serialization. A
	// stale cache is a contradictory fact, even when Save As also renamed parts.
	if e = verifyChartSemanticClosure(a, b); e != nil {
		return out, e
	}
	ga, e := readNativeGeometry(before.shape, before.ParentToken)
	if e != nil {
		return out, e
	}
	gb, e := readNativeGeometry(after.shape, after.ParentToken)
	if e != nil || !bytes.Equal(canonical(ga), canonical(gb)) {
		return out, fmt.Errorf("native chart geometry changed")
	}
	if e = verifyChartShell(a, b); e != nil {
		out.ManualReview = append(out.ManualReview, TextReconciliationIssue{Kind: "chart_formatting_not_adopted", SlideID: slide, ShapeToken: before.ShapeToken, NativePart: b.part, Detail: "Office chart/workbook serialization or formatting differs; numeric-only adoption retains the exact edited package and source formatting, without claiming visual equivalence"})
	}
	out.ChartPart, out.WorkbookPart = b.part, b.workbook
	add := func(entity, category, series, point, axis string, prior, next *float64) {
		if sameChartNumber(prior, next) {
			return
		}
		q := QuantitativeSemanticProposal{Entity: entity, Category: category, Series: series, Point: point, Axis: axis, Before: prior, Proposed: next, Status: "proposed", Reason: "Confirm edited numeric cache and workbook agree; null is missing and zero is observed. Source units and axes stay unchanged."}
		q.ID = digest(canonical(struct {
			Source, Packet, Entity, Category, Series, Point, Axis string
			Value                                                 *float64
		}{p.SourceHash(), packet.ReportSHA256, entity, category, series, point, axis, next}))
		out.Proposals = append(out.Proposals, q)
	}
	if kind == "scatter" {
		points := source["points"].([]any)
		for i, id := range keys["points"] {
			point := points[i].([]any)
			for axis, j := range map[string]int{"x": 0, "y": 1} {
				v, ok := point[j].(float64)
				if !ok || a.data[j][i] == nil || *a.data[j][i] != v {
					return out, fmt.Errorf("baseline scatter does not match keyed source")
				}
				if b.data[j][i] == nil {
					return out, fmt.Errorf("scatter requires observed x and y")
				}
				add("point", "", "", id, axis, &v, b.data[j][i])
			}
		}
	} else {
		cats := source["categories"].([]any)
		ss := source["series"].([]any)
		for i, id := range keys["series"] {
			s := ss[i].(map[string]any)
			if a.names[i] != s["name"] || b.names[i] != a.names[i] {
				return out, fmt.Errorf("native series identity/name changed")
			}
			values := s["values"].([]any)
			for j, key := range keys["categories"] {
				label := cats[j].(string)
				if source["preserveCategories"] != true && (kind == "column" || kind == "line") {
					label = strings.ToUpper(label)
				}
				if a.categories[j] != label || b.categories[j] != a.categories[j] {
					return out, fmt.Errorf("native category identity/label changed")
				}
				var prior *float64
				if values[j] != nil {
					v := values[j].(float64)
					prior = &v
				}
				if !sameChartNumber(prior, a.data[i][j]) {
					return out, fmt.Errorf("baseline chart does not match source numeric facts")
				}
				add("value", key, id, "", "", prior, b.data[i][j])
			}
		}
	}
	for _, f := range packet.Report.Geometry {
		if f.SlideID == slide && f.Status != "no_op" {
			out.UnresolvedGeometryIDs = append(out.UnresolvedGeometryIDs, f.ID)
		}
	}
	for _, f := range packet.Report.Fields {
		if f.SlideID == slide && f.Status != "no_op" {
			out.UnresolvedTextIDs = append(out.UnresolvedTextIDs, f.ID)
		}
	}
	for _, f := range packet.Report.Structure {
		if f.SlideID == slide && f.Status != "no_op" {
			out.UnresolvedStructureIDs = append(out.UnresolvedStructureIDs, f.ID)
		}
	}
	sort.Slice(out.Proposals, func(i, j int) bool { return out.Proposals[i].ID < out.Proposals[j].ID })
	return out, nil
}
func DecodeQuantitativeSemanticDecisions(raw []byte, file string) (GanttSemanticDecisions, error) {
	return decodeSemanticDecisions(raw, file, QuantitativeSemanticDecisionsSchema, "set_value")
}
func AdoptQuantitativeSemantics(p *Project, packet *TextReviewPacket, slide, node string, raw []byte, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	report, e := ProposeQuantitativeSemantics(p, packet, slide, node, bundle, engine)
	if e != nil {
		return empty, e
	}
	decisions, e := DecodeQuantitativeSemanticDecisions(raw, "quantitative-semantic-decisions")
	if e != nil {
		return empty, e
	}
	if decisions.ReportSHA256 != QuantitativeSemanticReportHash(report) {
		return empty, fmt.Errorf("quantitative report hash mismatch")
	}
	idx, t, e := diagramSlide(p, slide)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := quantitativeNode(&clone, node)
	if e != nil {
		return empty, e
	}
	source, keys, _, e := quantitativeSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	qs := map[string]QuantitativeSemanticProposal{}
	for _, q := range report.Proposals {
		qs[q.ID] = q
	}
	for _, d := range decisions.Decisions {
		q, ok := qs[d.ProposalID]
		if !ok {
			return empty, fmt.Errorf("unknown chart proposal")
		}
		if d.Action == "keep_source" {
			continue
		}
		if q.Status != "proposed" {
			return empty, fmt.Errorf("manual chart proposal cannot be adopted")
		}
		if q.Entity == "value" {
			if e = applyQuantitative(source, keys, QuantitativeOperation{Action: "set", Entity: "value", Key: q.Category, Series: q.Series, Value: q.Proposed, Missing: q.Proposed == nil}); e != nil {
				return empty, e
			}
		} else {
			i := teamFind(keys["points"], q.Point)
			j := 0
			if q.Axis == "y" {
				j = 1
			}
			source["points"].([]any)[i].([]any)[j] = *q.Proposed
		}
	}
	n.Arguments, n.Keys = source, keys
	if e = validateQuantitative(source, keys); e != nil {
		return empty, e
	}
	return semanticEvidenceCandidate(p, packet, slide, "quantitative-semantics", decisions.Actor, decisions.Reason, clone, report, raw, bundle, engine, apply)
}
