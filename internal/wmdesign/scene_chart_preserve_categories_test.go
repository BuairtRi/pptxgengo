package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func preserveCategoryParts(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	parts := map[string][]byte{}
	for _, file := range z.File {
		r, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		part, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		parts[file.Name] = part
	}
	return parts
}

func preserveCategoryBuild(t *testing.T, kind string, option any, include bool) []byte {
	t.Helper()
	node := map[string]any{"type": "chart", "kind": kind, "x": 100, "y": 150, "w": 650, "h": 250,
		"categories": []string{"2026 (to Sep 1)", "café Q2", "MiXeD & <R&D>"},
		"series":     []any{map[string]any{"name": "Original data", "values": []float64{10, 20, 30}}}}
	if include {
		node["preserveCategories"] = option
	}
	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026,
		BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "preserve-category-regression"},
		Slides:        []SlideSpec{{ID: "original-chart", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "chart", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
	pkg, _, err := BuildWithEngine(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	return pkg
}

func assertPreservedCategoryPackage(t *testing.T, pkg []byte, kind string, want []string) {
	t.Helper()
	parts := preserveCategoryParts(t, pkg)
	var chart struct {
		Chart struct {
			Plot struct {
				Bar struct {
					Series []struct {
						Category struct {
							Ref struct {
								Cache struct {
									Levels []struct {
										Points []struct {
											Value string `xml:"v"`
										} `xml:"pt"`
									} `xml:"lvl"`
								} `xml:"multiLvlStrCache"`
							} `xml:"multiLvlStrRef"`
						} `xml:"cat"`
					} `xml:"ser"`
				} `xml:"barChart"`
				Line struct {
					Series []struct {
						Category struct {
							Ref struct {
								Cache struct {
									Levels []struct {
										Points []struct {
											Value string `xml:"v"`
										} `xml:"pt"`
									} `xml:"lvl"`
								} `xml:"multiLvlStrCache"`
							} `xml:"multiLvlStrRef"`
						} `xml:"cat"`
					} `xml:"ser"`
				} `xml:"lineChart"`
			} `xml:"plotArea"`
		} `xml:"chart"`
	}
	if err := xml.Unmarshal(parts["ppt/charts/chart1.xml"], &chart); err != nil {
		t.Fatal(err)
	}
	got := []string{}
	if kind == "line" {
		if len(chart.Chart.Plot.Line.Series) != 1 || len(chart.Chart.Plot.Line.Series[0].Category.Ref.Cache.Levels) != 1 {
			t.Fatal("missing editable line category cache")
		}
		for _, point := range chart.Chart.Plot.Line.Series[0].Category.Ref.Cache.Levels[0].Points {
			got = append(got, point.Value)
		}
	} else {
		if len(chart.Chart.Plot.Bar.Series) != 1 || len(chart.Chart.Plot.Bar.Series[0].Category.Ref.Cache.Levels) != 1 {
			t.Fatal("missing editable bar/column category cache")
		}
		for _, point := range chart.Chart.Plot.Bar.Series[0].Category.Ref.Cache.Levels[0].Points {
			got = append(got, point.Value)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("native displayed category cache changed: got %q want %q", got, want)
	}
	wb := preserveCategoryParts(t, parts["ppt/embeddings/Microsoft_Excel_Worksheet1.xlsx"])
	var shared struct {
		Strings []struct {
			Text string `xml:"t"`
		} `xml:"si"`
	}
	if err := xml.Unmarshal(wb["xl/sharedStrings.xml"], &shared); err != nil {
		t.Fatal(err)
	}
	var sheet struct {
		Rows []struct {
			Cells []struct {
				Reference string `xml:"r,attr"`
				Type      string `xml:"t,attr"`
				Value     int    `xml:"v"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal(wb["xl/worksheets/sheet1.xml"], &sheet); err != nil {
		t.Fatal(err)
	}
	got = nil
	for _, row := range sheet.Rows[1:] {
		for _, cell := range row.Cells {
			if strings.HasPrefix(cell.Reference, "A") {
				if cell.Type != "s" || cell.Value < 0 || cell.Value >= len(shared.Strings) {
					t.Fatal("category workbook cell missing shared string")
				}
				got = append(got, shared.Strings[cell.Value].Text)
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("editable workbook category labels changed: got %q want %q", got, want)
	}
}

func TestSceneChartPreserveCategoriesNative(t *testing.T) {
	original := []string{"2026 (to Sep 1)", "café Q2", "MiXeD & <R&D>"}
	for _, kind := range []string{"column", "bar", "line"} {
		t.Run(kind, func(t *testing.T) {
			assertPreservedCategoryPackage(t, preserveCategoryBuild(t, kind, true, true), kind, original)
			legacy := append([]string(nil), original...)
			if kind != "bar" {
				for i := range legacy {
					legacy[i] = strings.ToUpper(legacy[i])
				}
			}
			absent := preserveCategoryBuild(t, kind, nil, false)
			assertPreservedCategoryPackage(t, absent, kind, legacy)
			explicitFalse := preserveCategoryBuild(t, kind, false, true)
			if !bytes.Equal(absent, explicitFalse) {
				t.Fatal("false/absent option changed historical package bytes")
			}
		})
	}
}

func TestSceneChartPreserveCategoriesStrictBoolean(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, value := range []string{`null`, `"true"`, `1`, `[]`, `{}`} {
		raw := json.RawMessage(`{"type":"chart","kind":"column","x":100,"y":150,"w":650,"h":250,"categories":["Original"],"series":[{"name":"S","values":[1]}],"preserveCategories":` + value + `}`)
		if _, _, err := r.planChartScene("chart", raw, SceneContext{Surface: "light"}); err == nil || !strings.Contains(err.Error(), "preserve_categories_requires_boolean") {
			t.Fatalf("invalid option %s accepted: %v", value, err)
		}
	}
}
