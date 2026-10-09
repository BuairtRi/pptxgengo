package pptx

import (
	"fmt"
	"strings"
	"testing"
)

func TestChartAutoUpdateWorkbookExplicitBinding(t *testing.T) {
	for _, automatic := range []bool{false, true} {
		t.Run(fmt.Sprint(automatic), func(t *testing.T) {
			p := New()
			data := []ChartData{{Name: "Actual", Labels: [][]string{{"A", "B"}}, Values: []float64{0, 14}}, {Name: "Plan", Labels: [][]string{{"A", "B"}}, Values: []float64{1, 15}}}
			if e := p.AddSlide().AddChart(ChartTypeBar, data, &ChartOptions{AutoUpdateWorkbook: automatic, PreserveWorkbookZeros: true}); e != nil {
				t.Fatal(e)
			}
			raw, e := p.Write()
			if e != nil {
				t.Fatal(e)
			}
			parts := unzipParts(t, raw)
			chart := parts["ppt/charts/chart1.xml"]
			flag := "0"
			if automatic {
				flag = "1"
			}
			for _, want := range []string{`<c:externalData r:id="rId1"><c:autoUpdate val="` + flag + `"/></c:externalData>`, `Sheet1!$B$1`, `Sheet1!$C$1`, `Sheet1!$B$2:$B$3`, `Sheet1!$C$2:$C$3`, `<c:numRef>`, `<c:numCache>`} {
				if !strings.Contains(chart, want) {
					t.Fatalf("binding missing %q", want)
				}
			}
			if strings.Contains(chart, "numLit") || strings.Contains(parts["ppt/charts/_rels/chart1.xml.rels"], `TargetMode="External"`) {
				t.Fatal("chart data no longer embedded/formula referenced")
			}
			if !strings.Contains(parts["ppt/charts/_rels/chart1.xml.rels"], "Microsoft_Excel_Worksheet1.xlsx") {
				t.Fatal("embedded workbook relationship missing")
			}
			// This property requests automatic update; only a desktop A/B exercise
			// can establish whether Office refreshes while Edit Data is active.
		})
	}
}
