package pptx

import (
	"encoding/xml"
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"
)

const compatibilityChartNS = "http://schemas.openxmlformats.org/drawingml/2006/chart"

// Parse structure independently of the byte fixtures. These checks protect the
// native compatibility fixes even if somebody regenerates the original JS data.
type compatibilityXMLNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr             `xml:",any,attr"`
	Children []compatibilityXMLNode `xml:",any"`
	Text     string                 `xml:",chardata"`
}

func (n compatibilityXMLNode) attr(name string) string {
	for _, a := range n.Attrs {
		if a.Name.Space == "" && a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func (n compatibilityXMLNode) children(ns, local string) []compatibilityXMLNode {
	var result []compatibilityXMLNode
	for _, c := range n.Children {
		if c.XMLName.Space == ns && c.XMLName.Local == local {
			result = append(result, c)
		}
	}
	return result
}

func compatibilityParse(t *testing.T, value string) compatibilityXMLNode {
	t.Helper()
	var node compatibilityXMLNode
	if err := xml.Unmarshal([]byte(value), &node); err != nil {
		t.Fatal(err)
	}
	return node
}

func compatibilityChild(t *testing.T, n compatibilityXMLNode, ns, local string) compatibilityXMLNode {
	t.Helper()
	children := n.children(ns, local)
	if len(children) != 1 {
		t.Fatalf("%s: expected one %s, got %d", n.XMLName.Local, local, len(children))
	}
	return children[0]
}

func compatibilityAssertPackageTargets(t *testing.T, parts map[string]string) {
	t.Helper()
	const relNS = "http://schemas.openxmlformats.org/package/2006/relationships"
	const typesNS = "http://schemas.openxmlformats.org/package/2006/content-types"
	seenOverrides := map[string]bool{}
	for _, override := range compatibilityParse(t, parts["[Content_Types].xml"]).children(typesNS, "Override") {
		name := strings.TrimPrefix(override.attr("PartName"), "/")
		if seenOverrides[name] {
			t.Errorf("duplicate content-type override: %s", name)
		}
		seenOverrides[name] = true
		if _, ok := parts[name]; !ok {
			t.Errorf("content-type override targets missing part: %s", name)
		}
	}
	for name, value := range parts {
		if !strings.HasSuffix(name, ".rels") {
			continue
		}
		sourceDir := path.Dir(path.Dir(name))
		seenIDs := map[string]bool{}
		for _, rel := range compatibilityParse(t, value).children(relNS, "Relationship") {
			id := rel.attr("Id")
			if id == "" || seenIDs[id] {
				t.Errorf("%s: missing/duplicate relationship ID %q", name, id)
			}
			seenIDs[id] = true
			if rel.attr("TargetMode") == "External" {
				continue
			}
			target := rel.attr("Target")
			resolved := path.Clean(path.Join(sourceDir, target))
			if strings.HasPrefix(target, "/") {
				resolved = strings.TrimPrefix(path.Clean(target), "/")
			}
			if _, ok := parts[resolved]; !ok {
				t.Errorf("%s: relationship %s targets missing part %s", name, id, resolved)
			}
		}
	}
}

func TestNativeCompatibilityNotesGraph(t *testing.T) {
	const presNS = "http://schemas.openxmlformats.org/presentationml/2006/main"
	const drawingNS = "http://schemas.openxmlformats.org/drawingml/2006/main"
	const relNS = "http://schemas.openxmlformats.org/package/2006/relationships"
	for _, notes := range [][]string{{""}, {"Speaker & reviewer <notes> — preserved"}, {"First", "", "Third"}} {
		t.Run(fmt.Sprintf("%d-slides-%v", len(notes), notes), func(t *testing.T) {
			p := New()
			for _, note := range notes {
				s := p.AddSlide()
				if note != "" {
					s.AddNotes(note)
				}
			}
			pkg, err := p.Write()
			if err != nil {
				t.Fatal(err)
			}
			parts := unzipParts(t, pkg)
			compatibilityAssertPackageTargets(t, parts)
			presentation := compatibilityParse(t, parts["ppt/presentation.xml"])
			if len(presentation.children(presNS, "notesMasterIdLst")) != 0 {
				t.Fatal("optional notes list reintroduced; native Mac repair workaround lost")
			}
			if len(presentation.children(presNS, "notesSz")) != 1 {
				t.Fatal("required notes page dimensions lost")
			}
			masterLinks := 0
			for _, rel := range compatibilityParse(t, parts["ppt/_rels/presentation.xml.rels"]).children(relNS, "Relationship") {
				if strings.HasSuffix(rel.attr("Type"), "/notesMaster") {
					masterLinks++
					if rel.attr("Target") != "notesMasters/notesMaster1.xml" {
						t.Fatal("notes master link changed")
					}
				}
			}
			if masterLinks != 1 {
				t.Fatalf("expected one retained notes master relationship, got %d", masterLinks)
			}
			for i, note := range notes {
				number := i + 1
				var texts []string
				var walk func(compatibilityXMLNode)
				walk = func(n compatibilityXMLNode) {
					if n.XMLName.Space == drawingNS && n.XMLName.Local == "t" {
						texts = append(texts, n.Text)
					}
					for _, child := range n.Children {
						walk(child)
					}
				}
				walk(compatibilityParse(t, parts[fmt.Sprintf("ppt/notesSlides/notesSlide%d.xml", number)]))
				if !reflect.DeepEqual(texts, []string{note, fmt.Sprint(number)}) {
					t.Fatalf("slide %d: notes text/number changed: %q", number, texts)
				}
				for _, relation := range []struct{ file, suffix, target string }{
					{fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", number), "/notesSlide", fmt.Sprintf("../notesSlides/notesSlide%d.xml", number)},
					{fmt.Sprintf("ppt/notesSlides/_rels/notesSlide%d.xml.rels", number), "/notesMaster", "../notesMasters/notesMaster1.xml"},
					{fmt.Sprintf("ppt/notesSlides/_rels/notesSlide%d.xml.rels", number), "/slide", fmt.Sprintf("../slides/slide%d.xml", number)},
				} {
					count := 0
					for _, rel := range compatibilityParse(t, parts[relation.file]).children(relNS, "Relationship") {
						if strings.HasSuffix(rel.attr("Type"), relation.suffix) && rel.attr("Target") == relation.target {
							count++
						}
					}
					if count != 1 {
						t.Fatalf("%s: required %s relationship count=%d", relation.file, relation.suffix, count)
					}
				}
			}
		})
	}
}

func TestNativeCompatibilityChartStructureAndWorkbooks(t *testing.T) {
	const sheetNS = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	for _, tc := range []struct {
		name   string
		build  func(*testing.T) *Presentation
		charts map[string]string
		ranges map[string]string
	}{
		{"bar", buildCase05, map[string]string{"chart1.xml": "barChart"}, map[string]string{"Microsoft_Excel_Worksheet1.xlsx": "A1:B5"}},
		{"line-and-pie", buildCase06, map[string]string{"chart2.xml": "lineChart", "chart3.xml": "pieChart"}, map[string]string{"Microsoft_Excel_Worksheet2.xlsx": "A1:C5", "Microsoft_Excel_Worksheet3.xlsx": "A1:B4"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pkg, err := tc.build(t).Write()
			if err != nil {
				t.Fatal(err)
			}
			parts := unzipParts(t, pkg)
			compatibilityAssertPackageTargets(t, parts)
			for file, kind := range tc.charts {
				root := compatibilityParse(t, parts["ppt/charts/"+file])
				plot := compatibilityChild(t, compatibilityChild(t, root, compatibilityChartNS, "chart"), compatibilityChartNS, "plotArea")
				chart := compatibilityChild(t, plot, compatibilityChartNS, kind)
				if kind == "pieChart" {
					continue
				}
				if kind == "lineChart" {
					group := compatibilityChild(t, chart, compatibilityChartNS, "grouping")
					if group.attr("val") != "standard" || chart.Children[0].XMLName.Local != "grouping" {
						t.Fatal("line grouping must be first and standard")
					}
				}
				axes := chart.children(compatibilityChartNS, "axId")
				if len(axes) != 2 {
					t.Fatalf("%s has %d axes; 2D charts require two", file, len(axes))
				}
				declared := map[string]bool{}
				for _, axisType := range []string{"catAx", "valAx"} {
					axis := compatibilityChild(t, plot, compatibilityChartNS, axisType)
					declared[compatibilityChild(t, axis, compatibilityChartNS, "axId").attr("val")] = true
				}
				for _, axis := range axes {
					if !declared[axis.attr("val")] {
						t.Fatal("undeclared axis reference")
					}
				}
				for _, series := range chart.children(compatibilityChartNS, "ser") {
					positions := map[string]int{}
					for i, child := range series.Children {
						positions[child.XMLName.Local] = i
					}
					for _, field := range []string{"dLbls", "cat", "val"} {
						compatibilityChild(t, series, compatibilityChartNS, field)
					}
					if !(positions["dLbls"] < positions["cat"] && positions["cat"] < positions["val"]) {
						t.Fatal("series data labels/categories/values out of schema order")
					}
					if kind == "lineChart" {
						compatibilityChild(t, series, compatibilityChartNS, "marker")
						if positions["marker"] >= positions["dLbls"] || len(series.children(compatibilityChartNS, "invertIfNegative")) != 0 {
							t.Fatal("line series contains bar-only inversion or misplaced marker")
						}
					} else {
						if positions["dPt"] >= positions["dLbls"] {
							t.Fatal("bar point styles must precede data labels")
						}
					}
				}
			}
			for file, expected := range tc.ranges {
				workbook := unzipParts(t, []byte(parts["ppt/embeddings/"+file]))
				compatibilityAssertPackageTargets(t, workbook)
				table := compatibilityParse(t, workbook["xl/tables/table1.xml"])
				if table.XMLName.Space != sheetNS || table.attr("ref") != expected {
					t.Fatalf("%s: table ref=%q, want %s", file, table.attr("ref"), expected)
				}
				rows := compatibilityChild(t, compatibilityParse(t, workbook["xl/worksheets/sheet1.xml"]), sheetNS, "sheetData").children(sheetNS, "row")
				lastRow := rows[len(rows)-1]
				cells := lastRow.children(sheetNS, "c")
				if cells[len(cells)-1].attr("r") != strings.Split(expected, ":")[1] {
					t.Fatal("table range does not cover the last data cell")
				}
			}
		})
	}
}
