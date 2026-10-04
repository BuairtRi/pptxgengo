package pptx

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func zeroWorkbookCells(t *testing.T, raw string) map[string]string {
	t.Helper()
	var sheet struct {
		Rows []struct {
			Cells []struct {
				Ref   string `xml:"r,attr"`
				Value string `xml:"v"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal([]byte(raw), &sheet); err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, r := range sheet.Rows {
		for _, c := range r.Cells {
			out[c.Ref] = c.Value
		}
	}
	return out
}

func TestWorkbookPreserveZerosOptIn(t *testing.T) {
	data := []ChartData{{Name: "S", Labels: [][]string{{"A", "B", "C", "D"}}, Values: []float64{0, -2.5, 7}}}
	legacy := buildSheet1(data, &ChartOptions{}, 0, false, false, false)
	if legacy != buildSheet1(data, &ChartOptions{PreserveWorkbookZeros: false}, 0, false, false, false) {
		t.Fatal("false changed legacy bytes")
	}
	got := zeroWorkbookCells(t, buildSheet1(data, &ChartOptions{PreserveWorkbookZeros: true}, 0, false, false, false))
	for cell, want := range map[string]string{"B2": "0", "B3": "-2.5", "B4": "7", "B5": ""} {
		if got[cell] != want {
			t.Fatalf("%s=%q want %q", cell, got[cell], want)
		}
	}
	old := zeroWorkbookCells(t, legacy)
	if old["B2"] != "" || old["B5"] != "" {
		t.Fatal("historical zero/missing blanks changed")
	}
	if !reflect.DeepEqual(data[0].Values, []float64{0, -2.5, 7}) {
		t.Fatal("source mutated")
	}
}

func TestWorkbookPreserveZerosSparseAndShortValues(t *testing.T) {
	data := []ChartData{
		{Name: "Sparse", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{0, 5, 0}, MissingValues: []bool{false, true, false}},
		{Name: "Short", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{0}},
	}
	cells := zeroWorkbookCells(t, buildSheet1(data, &ChartOptions{PreserveWorkbookZeros: true}, 0, false, false, false))
	if cells["B2"] != "0" || cells["B4"] != "0" || cells["C2"] != "0" || cells["C3"] != "" || cells["C4"] != "" {
		t.Fatalf("observed/missing confused: %v", cells)
	}
	if _, ok := cells["B3"]; ok {
		t.Fatal("explicit missing observation emitted")
	}
}

func TestWorkbookPreserveZerosMultiCategoryShortValues(t *testing.T) {
	data := []ChartData{{Name: "S", Labels: [][]string{{"A", "B"}, {"Group", ""}}, Values: []float64{0}}}
	cells := zeroWorkbookCells(t, buildSheet1(data, &ChartOptions{PreserveWorkbookZeros: true}, 0, false, false, true))
	if cells["C2"] != "0" || cells["C3"] != "" {
		t.Fatalf("multi-category zero/missing confused: %v", cells)
	}
}
