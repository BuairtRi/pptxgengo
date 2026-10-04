package pptx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"strings"
	"testing"
)

func TestChartMissingValuesNativeWorkbookAndCache(t *testing.T) {
	p := New()
	d := ChartData{Name: "Actual", Labels: [][]string{{"A", "B", "C", "D", "E"}}, Values: []float64{0, 999, 2, 999, 999}, MissingValues: []bool{false, true, false, true, true}}
	err := p.AddSlide().AddChart(ChartTypeLine, []ChartData{d, {Name: "Plan", Labels: d.Labels, Values: []float64{0, 1, 2, 3, 4}}}, &ChartOptions{DisplayBlanksAs: "span"})
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := p.Write()
	if err != nil {
		t.Fatal(err)
	}
	parts := unzipParts(t, pkg)
	chart := parts["ppt/charts/chart1.xml"]
	var got struct {
		Series []struct {
			Cat struct {
				Ref struct {
					Cache struct {
						Count struct {
							V int `xml:"val,attr"`
						} `xml:"ptCount"`
						Points []struct {
							Index int    `xml:"idx,attr"`
							Value string `xml:"v"`
						} `xml:"lvl>pt"`
					} `xml:"multiLvlStrCache"`
				} `xml:"multiLvlStrRef"`
			} `xml:"cat"`
			Val struct {
				Ref struct {
					Formula string `xml:"f"`
					Cache   struct {
						Count struct {
							V int `xml:"val,attr"`
						} `xml:"ptCount"`
						Points []struct {
							Index int    `xml:"idx,attr"`
							Value string `xml:"v"`
						} `xml:"pt"`
					} `xml:"numCache"`
				} `xml:"numRef"`
			} `xml:"val"`
		} `xml:"chart>plotArea>lineChart>ser"`
	}
	if err = xml.Unmarshal([]byte(chart), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Series) != 2 {
		t.Fatal("missing editable series")
	}
	s := got.Series[0]
	if s.Cat.Ref.Cache.Count.V != 5 || len(s.Cat.Ref.Cache.Points) != 5 || s.Val.Ref.Cache.Count.V != 5 || s.Val.Ref.Formula != "Sheet1!$B$2:$B$6" || len(s.Val.Ref.Cache.Points) != 2 || s.Val.Ref.Cache.Points[0].Index != 0 || s.Val.Ref.Cache.Points[0].Value != "0" || s.Val.Ref.Cache.Points[1].Index != 2 || s.Val.Ref.Cache.Points[1].Value != "2" {
		t.Fatalf("misaligned cache %+v", s)
	}
	if !strings.Contains(chart, `<c:dispBlanksAs val="span"/>`) {
		t.Fatal("source connection policy lost")
	}
	var wb []byte
	z, err := zip.NewReader(bytes.NewReader(pkg), int64(len(pkg)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if strings.HasSuffix(f.Name, ".xlsx") {
			r, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			wb, e = io.ReadAll(r)
			r.Close()
			if e != nil {
				t.Fatal(e)
			}
		}
	}
	if len(wb) == 0 {
		t.Fatal("no editable workbook")
	}
	sheet := unzipParts(t, wb)["xl/worksheets/sheet1.xml"]
	if !strings.Contains(sheet, `<c r="B2"><v>0</v></c>`) || !strings.Contains(sheet, `<c r="B4"><v>2</v></c>`) {
		t.Fatalf("observed values lost: %s", sheet)
	}
	if !strings.Contains(sheet, `<c r="C2"><v>0</v></c>`) {
		t.Fatal("dense companion zero lost")
	}
	for _, cell := range []string{"B3", "B5", "B6"} {
		if strings.Contains(sheet, `<c r="`+cell+`"`) {
			t.Fatal("missing observation emitted as cell", cell)
		}
	}
	for _, cell := range []string{"A2", "A3", "A4", "A5", "A6"} {
		if !strings.Contains(sheet, `<c r="`+cell+`"`) {
			t.Fatal("category row omitted", cell)
		}
	}
}

func TestChartMissingValuesStrictAndCombo(t *testing.T) {
	base := ChartData{Name: "A", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{1, 0, 3}, MissingValues: []bool{false, true, false}}
	for _, tc := range []struct {
		name   string
		kind   ChartType
		mutate func(*ChartData)
		opt    ChartOptions
	}{
		{"mask length", ChartTypeLine, func(d *ChartData) { d.MissingValues = []bool{true} }, ChartOptions{}},
		{"label length", ChartTypeLine, func(d *ChartData) { d.Labels = [][]string{{"A"}} }, ChartOptions{}},
		{"second category level", ChartTypeLine, func(d *ChartData) { d.Labels = append(d.Labels, []string{"A"}) }, ChartOptions{}},
		{"all absent", ChartTypeLine, func(d *ChartData) { d.MissingValues = []bool{true, true, true} }, ChartOptions{}},
		{"NaN", ChartTypeLine, func(d *ChartData) { d.Values = []float64{1, math.NaN(), 3} }, ChartOptions{}},
		{"Inf", ChartTypeLine, func(d *ChartData) { d.Values = []float64{1, math.Inf(1), 3} }, ChartOptions{}},
		{"column", ChartTypeBar, func(d *ChartData) {}, ChartOptions{}},
		{"stacked", ChartTypeLine, func(d *ChartData) {}, ChartOptions{BarGrouping: "stacked"}},
		{"oversized", ChartTypeLine, func(d *ChartData) {
			d.Values = make([]float64, 10001)
			d.MissingValues = make([]bool, 10001)
			d.Labels = [][]string{make([]string, 10001)}
		}, ChartOptions{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := base
			tc.mutate(&d)
			if e := New().AddSlide().AddChart(tc.kind, []ChartData{d}, &tc.opt); e == nil {
				t.Fatal("invalid sparse series accepted")
			}
		})
	}
	p := New()
	s := p.AddSlide()
	e := s.AddMultiChart([]IChartMulti{{Type: ChartTypeLine, Data: []ChartData{base}, Options: &ChartOptions{DisplayBlanksAs: "span"}}}, &ChartOptions{DisplayBlanksAs: "span"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Write(); e != nil {
		t.Fatal(e)
	}
	if e = New().AddSlide().AddMultiChart([]IChartMulti{{Type: ChartTypeBar, Data: []ChartData{base}}}, nil); e == nil {
		t.Fatal("combo bypassed sparse-kind validation")
	}
}

func TestChartMissingValuesCommonCategoryGrid(t *testing.T) {
	sparse := ChartData{Name: "Sparse", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{0, 0, 2}, MissingValues: []bool{false, true, false}}
	for _, tc := range []struct {
		name  string
		other ChartData
	}{
		{"no labels", ChartData{Name: "Dense", Values: []float64{1, 2, 3}}},
		{"length", ChartData{Name: "Dense", Labels: [][]string{{"A", "B", "C", "D"}}, Values: []float64{1, 2, 3, 4}}},
		{"names", ChartData{Name: "Dense", Labels: [][]string{{"A", "changed", "C"}}, Values: []float64{1, 2, 3}}},
		{"levels", ChartData{Name: "Dense", Labels: [][]string{{"A", "B", "C"}, {"extra", "extra", "extra"}}, Values: []float64{1, 2, 3}}},
		{"NaN dense", ChartData{Name: "Dense", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{1, math.NaN(), 3}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, order := range [][]ChartData{{sparse, tc.other}, {tc.other, sparse}} {
				if e := New().AddSlide().AddChart(ChartTypeLine, order, nil); e == nil {
					t.Fatal("unaligned companion accepted")
				}
				if e := New().AddSlide().AddMultiChart([]IChartMulti{{Type: ChartTypeLine, Data: order[:1]}, {Type: ChartTypeLine, Data: order[1:]}}, nil); e == nil {
					t.Fatal("unaligned combo companion accepted")
				}
			}
		})
	}
	// Dense-only input keeps its historical API contract.
	dense := sparse
	dense.MissingValues = nil
	if e := New().AddSlide().AddChart(ChartTypeLine, []ChartData{dense, {Name: "Long", Labels: [][]string{{"A", "B", "C", "D"}}, Values: []float64{1, 2, 3, 4}}}, nil); e != nil {
		t.Fatal("dense legacy API changed", e)
	}
}

func TestChartMissingValuesComboGlobalIndices(t *testing.T) {
	a := ChartData{Name: "One", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{0, 0, 2}, MissingValues: []bool{false, true, false}}
	b := ChartData{Name: "Two", Labels: a.Labels, Values: []float64{3, 4, 0}, MissingValues: []bool{false, false, true}}
	multi := []IChartMulti{{Type: ChartTypeLine, Data: []ChartData{a}}, {Type: ChartTypeLine, Data: []ChartData{b}}}
	p := New()
	if e := p.AddSlide().AddMultiChart(multi, &ChartOptions{DisplayBlanksAs: "span"}); e != nil {
		t.Fatal(e)
	}
	if multi[1].Data[0].DataIndex != 0 {
		t.Fatal("rewrote caller component")
	}
	pkg, e := p.Write()
	if e != nil {
		t.Fatal(e)
	}
	chart := unzipParts(t, pkg)["ppt/charts/chart1.xml"]
	for _, want := range []string{`<c:idx val="0"/>`, `<c:idx val="1"/>`, `<c:f>Sheet1!$B$1</c:f>`, `<c:f>Sheet1!$C$1</c:f>`, `<c:f>Sheet1!$B$2:$B$4</c:f>`, `<c:f>Sheet1!$C$2:$C$4</c:f>`} {
		if strings.Count(chart, want) != 1 {
			t.Fatal("combo cache doesn't match workbook column", want)
		}
	}
	z, e := zip.NewReader(bytes.NewReader(pkg), int64(len(pkg)))
	if e != nil {
		t.Fatal(e)
	}
	var wb []byte
	for _, f := range z.File {
		if strings.HasSuffix(f.Name, ".xlsx") {
			r, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			wb, e = io.ReadAll(r)
			r.Close()
			if e != nil {
				t.Fatal(e)
			}
		}
	}
	sheet := unzipParts(t, wb)["xl/worksheets/sheet1.xml"]
	for _, want := range []string{`<c r="B2"><v>0</v></c>`, `<c r="B4"><v>2</v></c>`, `<c r="C2"><v>3</v></c>`, `<c r="C3"><v>4</v></c>`} {
		if !strings.Contains(sheet, want) {
			t.Fatal("combo observation lost", want)
		}
	}
	for _, want := range []string{`<c r="B3"`, `<c r="C4"`} {
		if strings.Contains(sheet, want) {
			t.Fatal("combo missing observation emitted", want)
		}
	}
}

func TestChartMissingValuesBoundedComboAndCopy(t *testing.T) {
	a := ChartData{Name: "Sparse", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{0, 0, 2}, MissingValues: []bool{false, true, false}}
	dense := a
	dense.MissingValues = nil
	for _, kind := range []ChartType{ChartTypeBar, ChartTypeScatter, ChartTypeBubble, ChartTypeArea} {
		if e := New().AddSlide().AddMultiChart([]IChartMulti{{Type: ChartTypeLine, Data: []ChartData{a}}, {Type: kind, Data: []ChartData{dense}}}, nil); e == nil {
			t.Fatal("nonline sparse combo accepted", kind)
		}
	}
	many := make([]ChartData, 65)
	for i := range many {
		many[i] = a
	}
	if e := New().AddSlide().AddChart(ChartTypeLine, many, nil); e == nil {
		t.Fatal("unbounded sparse series accepted")
	}
	a.Name = strings.Repeat("a", 4097)
	if e := New().AddSlide().AddChart(ChartTypeLine, []ChartData{a}, nil); e == nil {
		t.Fatal("unbounded sparse name accepted")
	}
	a.Name = "A"
	a.Labels = [][]string{{strings.Repeat("a", 4097), "B", "C"}}
	if e := New().AddSlide().AddChart(ChartTypeLine, []ChartData{a}, nil); e == nil {
		t.Fatal("unbounded sparse label accepted")
	}
}
