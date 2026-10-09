package wmdesign

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
)

func TestCommercialCompleteFiveYearValueDisplays(t *testing.T) {
	raw, e := os.ReadFile("testdata/commercial-five-year-model.json")
	if e != nil {
		t.Fatal(e)
	}
	var m CommercialModel
	if e = json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	results, e := EvaluateCommercial(m)
	if e != nil {
		t.Fatal(e)
	}
	by := map[string]CommercialResult{}
	for _, r := range results {
		by[r.Key] = r
	}
	if by["benefit"].Exact != "13000000" || by["net"].Exact != "8800000" || by["roi"].Exact != "44/21" || by["value-dollar"].Exact != "65/21" || by["payback-months"].Exact != "252/13" {
		t.Fatalf("inconsistent financial case %+v", by)
	}
	expected := map[string]string{"node02": "$4.2M", "node06": "$2.6M", "node09": "3.1x", "node12": "210%", "node13": "19", "node14": "$6.2M", "node15": "$13.0M", "node16": "Five annual benefits of $2.6M; all investment at inception."}
	group := []string{"node02", "node06", "node09", "node12", "node13", "node14", "node15", "node16"}
	for id, want := range expected {
		kind, field := "text", "text"
		if id == "node12" || id == "node13" || id == "node14" || id == "node15" {
			kind, field = "metric", "value"
		}
		if id == "node16" {
			kind, field = "textblock", "body"
		}
		p := map[string]any{"type": kind, field: "UNREVIEWED", "retained": "Unrelated authored content"}
		out, _, e := MaterializeCommercialPresentation(CommercialSpec{NodeID: id, Group: group, Model: m, Presentation: p})
		if e != nil {
			t.Fatal(e)
		}
		if out[field] != want || out["retained"] != p["retained"] {
			t.Fatalf("%s got %v wanted %s", id, out[field], want)
		}
	}
	// A changed investment propagates through every declared return/payback
	// display, rather than changing one hero number and retaining an old story.
	for i := range m.Rows {
		if m.Rows[i].Key == "investment" {
			value := "5200000"
			m.Rows[i].Value = &value
		}
	}
	for id, field := range map[string]string{"node02": "text", "node09": "text", "node12": "value", "node13": "value", "node14": "value"} {
		out, _, e := MaterializeCommercialPresentation(CommercialSpec{NodeID: id, Group: group, Model: m, Presentation: map[string]any{"type": map[string]string{"node02": "text", "node09": "text", "node12": "metric", "node13": "metric", "node14": "metric"}[id], field: "UNREVIEWED"}})
		if e != nil {
			t.Fatal(e)
		}
		if out[field] == expected[id] || strings.Contains(out[field].(string), "UNREVIEWED") {
			t.Fatalf("dependent %s claim stayed stale", id)
		}
	}
}

// Display precision is not a rounding transform on chart/workbook facts.
func TestCommercialNumericTargetsPreserveExactFacts(t *testing.T) {
	zero := 0
	m := CommercialModel{Precision: 0, Rounding: "half_even", Rows: []CommercialRow{
		{Key: "net", Label: "Net", Unit: "USD", Value: cdecimal("10800000"), Source: "Reviewed illustrative input"},
		{Key: "one", Label: "One", Unit: "ratio", Value: cdecimal("1"), Assumption: "Illustrative ratio"},
		{Key: "three", Label: "Three", Unit: "ratio", Value: cdecimal("3"), Assumption: "Illustrative ratio"},
		{Key: "third", Label: "Third", Unit: "ratio", Formula: &CommercialFormula{Operation: "divide", Inputs: []string{"one", "three"}}},
	}, Targets: []CommercialTarget{
		{Row: "net", Path: "/series/0/values/0", Scale: "1000000", Numeric: true, Precision: &zero},
		{Row: "net", Path: "/label", Scale: "1000000", Precision: &zero},
		{Row: "third", Path: "/series/0/values/1", Numeric: true},
	}}
	s := CommercialSpec{Model: m, Presentation: map[string]any{"type": "chart", "label": "old", "series": []any{map[string]any{"values": []any{0.0, 0.0}}}}}
	out, _, e := MaterializeCommercialPresentation(s)
	if e != nil {
		t.Fatal(e)
	}
	values := out["series"].([]any)[0].(map[string]any)["values"].([]any)
	if values[0] != 10.8 || math.Abs(values[1].(float64)-1.0/3) > 1e-16 || out["label"] != "11" {
		t.Fatalf("facts/display rounding mixed %+v", out)
	}
	label, e := sceneChartDisplay(sceneChartSource{}, values[0].(float64))
	if e != nil || label != "10.8" {
		t.Fatalf("chart endlabel %s %v", label, e)
	}
	s.Model.Targets[0].Scale = "0.001"
	if _, _, e = MaterializeCommercialPresentation(s); e == nil {
		t.Fatal("overflow numeric fact accepted")
	}
	s.Model.Targets[0].Scale = "1000000"
	bad := 9
	s.Model.Targets[0].Precision = &bad
	if _, _, e = MaterializeCommercialPresentation(s); e == nil {
		t.Fatal("invalid display precision accepted")
	}
}

func TestCommercialNumericTargetExactBoundary(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{{"1000000000", true}, {"-1000000000", true}, {"1000000000.000000001", false}, {"-1000000000.000000001", false}, {"0", true}, {"0.000000000001", true}} {
		m := CommercialModel{Precision: 0, Rounding: "half_even", Rows: []CommercialRow{{Key: "fact", Label: "Reviewed fact", Unit: "USD", Value: cdecimal(tc.value), Source: "Illustrative bound control"}}, Targets: []CommercialTarget{{Row: "fact", Path: "/value", Numeric: true}}}
		_, _, e := MaterializeCommercialPresentation(CommercialSpec{Model: m, Presentation: map[string]any{"type": "num", "value": 0.0}})
		if (e == nil) != tc.valid {
			t.Fatalf("exact bound %s: %v", tc.value, e)
		}
	}
}
