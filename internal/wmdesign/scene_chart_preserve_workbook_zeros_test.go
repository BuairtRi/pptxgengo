package wmdesign

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"path/filepath"
	"strings"
	"testing"
)

func preservedZeroPackage(t *testing.T, kind string, option any, include bool) []byte {
	t.Helper()
	node := map[string]any{"type": "chart", "kind": kind, "x": 100, "y": 150, "w": 650, "h": 250, "categories": []string{"2024", "2025", "2026 (to Sep 1)"}, "preserveCategories": true, "series": []any{map[string]any{"name": "Patch", "values": []float64{0, 0, 11}}}}
	if include {
		node["preserveWorkbookZeros"] = option
	}
	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "workbook-zero-regression"}, Slides: []SlideSpec{{ID: "source-chart", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "chart", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
	pkg, _, err := BuildWithEngine(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	return pkg
}

func TestSceneChartPreserveWorkbookZerosNative(t *testing.T) {
	for _, kind := range []string{"column", "bar", "line"} {
		t.Run(kind, func(t *testing.T) {
			legacy := preservedZeroPackage(t, kind, nil, false)
			if !bytes.Equal(legacy, preservedZeroPackage(t, kind, false, true)) {
				t.Fatal("absent/false package bytes differ")
			}
			old := preserveCategoryParts(t, legacy)
			pkg := preservedZeroPackage(t, kind, true, true)
			parts := preserveCategoryParts(t, pkg)
			if !bytes.Equal(old["ppt/charts/chart1.xml"], parts["ppt/charts/chart1.xml"]) {
				t.Fatal("workbook option changed native chart XML")
			}
			wb := preserveCategoryParts(t, parts["ppt/embeddings/Microsoft_Excel_Worksheet1.xlsx"])
			var sheet struct {
				Rows []struct {
					Cells []struct {
						Ref   string `xml:"r,attr"`
						Type  string `xml:"t,attr"`
						Value string `xml:"v"`
					} `xml:"c"`
				} `xml:"sheetData>row"`
			}
			if err := xml.Unmarshal(wb["xl/worksheets/sheet1.xml"], &sheet); err != nil {
				t.Fatal(err)
			}
			cells := map[string]string{}
			for _, r := range sheet.Rows {
				for _, c := range r.Cells {
					if strings.HasPrefix(c.Ref, "B") && c.Type != "s" {
						cells[c.Ref] = c.Value
					}
				}
			}
			if cells["B2"] != "0" || cells["B3"] != "0" || cells["B4"] != "11" {
				t.Fatalf("editable workbook values changed: %v", cells)
			}
			assertPreservedCategoryPackage(t, pkg, kind, []string{"2024", "2025", "2026 (to Sep 1)"})
		})
	}
}

func TestSceneChartPreserveWorkbookZerosStrictBoolean(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, value := range []string{`null`, `"true"`, `1`, `[]`, `{}`} {
		raw := json.RawMessage(`{"type":"chart","kind":"column","x":100,"y":150,"w":650,"h":250,"categories":["Original"],"series":[{"name":"S","values":[0]}],"preserveWorkbookZeros":` + value + `}`)
		if _, _, err := r.planChartScene("chart", raw, SceneContext{Surface: "light"}); err == nil || !strings.Contains(err.Error(), "preserve_workbook_zeros_requires_boolean") {
			t.Fatalf("invalid option %s accepted: %v", value, err)
		}
	}
}
