package wmdesign

import (
	"math/big"
	"strings"
	"testing"
)

func cdecimal(v string) *string { return &v }
func commercialExample() CommercialModel {
	return CommercialModel{Precision: 2, Rounding: "half_even", Assumptions: []string{"Illustrative, no taxes; cash flows at period end."}, Rows: []CommercialRow{
		{Key: "investment", Label: "Investment", Unit: "USD", Value: cdecimal("0.10"), Source: "Example source"},
		{Key: "benefit", Label: "Benefit", Unit: "USD", Value: cdecimal("0.20"), Assumption: "Illustrative benefit"},
		{Key: "net", Label: "Net", Unit: "USD", Formula: &CommercialFormula{Operation: "subtract", Inputs: []string{"benefit", "investment"}}},
		{Key: "roi", Label: "ROI", Unit: "ratio", Formula: &CommercialFormula{Operation: "roi", Inputs: []string{"benefit", "investment"}}},
	}}
}
func TestCommercialRationalCalculationRoundingAndUnits(t *testing.T) {
	m := commercialExample()
	out, e := EvaluateCommercial(m)
	if e != nil {
		t.Fatal(e)
	}
	if out[2].Exact != "1/10" || out[2].Rounded != "0.10" || out[3].Rounded != "1.00" {
		t.Fatalf("decimal identity lost %+v", out)
	}
	for _, tc := range []struct{ in, mode, want string }{{"2.345", "half_even", "2.34"}, {"2.355", "half_even", "2.36"}, {"-2.345", "half_even", "-2.34"}, {"-2.345", "half_away", "-2.35"}, {"0.005", "half_even", "0.00"}} {
		r, _ := new(big.Rat).SetString(tc.in)
		v, e := RoundCommercial(r, 2, tc.mode)
		if e != nil || v != tc.want {
			t.Fatalf("round %+v %s %v", tc, v, e)
		}
	}
	m.Rows[1].Unit = "EUR"
	if _, e = EvaluateCommercial(m); e == nil {
		t.Fatal("mixed currencies accepted")
	}
}
func TestCommercialFormulaNPVPaybackAndCycles(t *testing.T) {
	m := CommercialModel{Precision: 4, Rounding: "half_even", Rows: []CommercialRow{
		{Key: "rate", Label: "Rate", Unit: "ratio", Value: cdecimal("0.10"), Assumption: "Discount rate"},
		{Key: "initial", Label: "Initial", Unit: "USD", Value: cdecimal("-100"), Source: "Example"},
		{Key: "cash", Label: "Cash", Unit: "USD", Value: cdecimal("110"), Source: "Example"},
		{Key: "npv", Label: "NPV", Unit: "USD", Formula: &CommercialFormula{Operation: "npv", Inputs: []string{"rate", "initial", "cash"}}},
		{Key: "cost", Label: "Cost", Unit: "USD", Value: cdecimal("100"), Source: "Example"},
		{Key: "payback", Label: "Payback", Unit: "years", Assumption: "Cash is annual net benefit in USD/year", Formula: &CommercialFormula{Operation: "payback", Inputs: []string{"cost", "cash"}}},
	}}
	out, e := EvaluateCommercial(m)
	if e != nil {
		t.Fatal(e)
	}
	if out[3].Rounded != "0.0000" || out[5].Rounded != "0.9091" {
		t.Fatalf("bad finance %+v", out)
	}
	m.Rows[2].Value = nil
	m.Rows[2].Formula = &CommercialFormula{Operation: "sum", Inputs: []string{"npv"}}
	if _, e = EvaluateCommercial(m); e == nil || !strings.Contains(e.Error(), "cycle") {
		t.Fatalf("cycle accepted %v", e)
	}
}
func TestCommercialInvalidFactsOverflowAndMissing(t *testing.T) {
	for _, v := range []string{"NaN", "1e3", "", "1/2", "1000000000000000000.000000000001"} {
		m := commercialExample()
		m.Rows[0].Value = &v
		if _, e := EvaluateCommercial(m); e == nil {
			t.Fatalf("invalid decimal/overflow %q", v)
		}
	}
	m := commercialExample()
	m.Rows[0].Value = cdecimal("0")
	if _, e := EvaluateCommercial(m); e == nil {
		t.Fatal("zero investment accepted for ROI")
	}
	m = commercialExample()
	m.Rows[0].Value = nil
	if _, e := EvaluateCommercial(m); e == nil {
		t.Fatal("missing silently zero")
	}
	m = commercialExample()
	m.Rows[0].Source = ""
	if _, e := EvaluateCommercial(m); e == nil {
		t.Fatal("input provenance missing")
	}
}
func TestCommercialPresentationMappingAndGeometrySafety(t *testing.T) {
	m := commercialExample()
	m.Targets = []CommercialTarget{{Row: "net", Path: "/value", Prefix: "$"}}
	s := CommercialSpec{Model: m, Presentation: map[string]any{"type": "metric", "value": "$old", "label": "Net value"}}
	p, _, e := MaterializeCommercialPresentation(s)
	if e != nil {
		t.Fatal(e)
	}
	if p["value"] != "$0.10" || p["label"] != "Net value" || s.Presentation["value"] != "$old" {
		t.Fatal("presentation changed unrelated copy/source")
	}
	for _, path := range []string{"/missing", "/x", "/type"} {
		s.Model.Targets[0].Path = path
		if _, _, e = MaterializeCommercialPresentation(s); e == nil {
			t.Fatalf("bad target accepted %s", path)
		}
	}
}

func TestCommercialAllArithmeticAndPaybackDimensions(t *testing.T) {
	m := CommercialModel{Precision: 2, Rounding: "half_even", Rows: []CommercialRow{
		{Key: "base", Label: "Base", Unit: "USD", Value: cdecimal("100"), Source: "Illustrative source"},
		{Key: "ratio", Label: "Ratio", Unit: "ratio", Value: cdecimal("0.2"), Assumption: "Illustrative factor"},
		{Key: "scale", Label: "Scaled", Unit: "USD", Formula: &CommercialFormula{Operation: "multiply", Inputs: []string{"base", "ratio"}}},
		{Key: "sum", Label: "Combined", Unit: "USD", Formula: &CommercialFormula{Operation: "sum", Inputs: []string{"base", "scale"}}},
		{Key: "divide", Label: "Ratio", Unit: "ratio", Formula: &CommercialFormula{Operation: "divide", Inputs: []string{"scale", "base"}}},
		{Key: "annual", Label: "Annual net benefit", Unit: "USD/year", Value: cdecimal("20"), Source: "Illustrative annual flow"},
		{Key: "payback", Label: "Payback", Unit: "years", Assumption: "Annual net benefit; undiscounted", Formula: &CommercialFormula{Operation: "payback", Inputs: []string{"base", "annual"}}},
	}}
	out, e := EvaluateCommercial(m)
	if e != nil {
		t.Fatal(e)
	}
	if out[2].Rounded != "20.00" || out[3].Rounded != "120.00" || out[4].Rounded != "0.20" || out[6].Rounded != "5.00" {
		t.Fatalf("arithmetic %+v", out)
	}
	m.Rows[5].Unit = "EUR/year"
	if _, e = EvaluateCommercial(m); e == nil {
		t.Fatal("mixed-currency payback accepted")
	}
	m.Rows[5].Unit = "USD/year"
	m.Rows[6].Unit = "USD"
	if _, e = EvaluateCommercial(m); e == nil {
		t.Fatal("monetary payback duration accepted")
	}
	m.Rows[6].Unit = "years"
	m.Rows[0].Value = cdecimal("0")
	if _, e = EvaluateCommercial(m); e == nil {
		t.Fatal("division by zero accepted")
	}
}
