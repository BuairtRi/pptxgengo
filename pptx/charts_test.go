package pptx

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Test scaffolding: replicate the option defaults that gen-objects.ts
// addChartDefinition() applies, since makeXmlCharts/createExcelWorksheet consume
// a fully-defaulted SlideRelChart. This is TEST-ONLY input construction (it does
// NOT implement the gen-objects module).
// ---------------------------------------------------------------------------

// applyChartDefaults mirrors the relevant default assignments of
// addChartDefinition for a single-type chart.
func applyChartDefaults(chartType ChartType, data []ChartData, opts *ChartOptions) {
	opts.Type = chartType

	// _dataIndex + labels[][] normalization is handled by callers (they build
	// ChartData directly with Labels as [][]string and DataIndex set).

	// B: barDir
	if opts.BarDir != "bar" && opts.BarDir != "col" {
		opts.BarDir = "col"
	}
	// barGrouping
	switch chartType {
	case ChartTypeArea:
		if opts.BarGrouping != "stacked" && opts.BarGrouping != "standard" && opts.BarGrouping != "percentStacked" {
			opts.BarGrouping = "standard"
		}
	case ChartTypeBar:
		if opts.BarGrouping != "clustered" && opts.BarGrouping != "stacked" && opts.BarGrouping != "percentStacked" {
			opts.BarGrouping = "clustered"
		}
	case ChartTypeBar3d:
		if opts.BarGrouping != "clustered" && opts.BarGrouping != "stacked" && opts.BarGrouping != "standard" && opts.BarGrouping != "percentStacked" {
			opts.BarGrouping = "standard"
		}
	}
	if strings.Contains(opts.BarGrouping, "tacked") && opts.BarGapWidthPct == 0 {
		opts.BarGapWidthPct = 50
	}
	// legendPos
	switch opts.LegendPos {
	case "b", "l", "r", "t", "tr":
	default:
		opts.LegendPos = "r"
	}
	// bar3DShape
	switch opts.Bar3DShape {
	case "cone", "coneToMax", "box", "cylinder", "pyramid", "pyramidToMax":
	default:
		opts.Bar3DShape = "box"
	}
	// lineDataSymbol
	switch opts.LineDataSymbol {
	case "circle", "dash", "diamond", "dot", "none", "square", "triangle":
	default:
		opts.LineDataSymbol = "circle"
	}
	// displayBlanksAs
	if opts.DisplayBlanksAs != "gap" && opts.DisplayBlanksAs != "span" {
		opts.DisplayBlanksAs = "span"
	}
	// radarStyle
	switch opts.RadarStyle {
	case "standard", "marker", "filled":
	default:
		opts.RadarStyle = "standard"
	}
	if opts.LineDataSymbolSize == 0 {
		opts.LineDataSymbolSize = 6
	}
	if opts.LineDataSymbolLineSize == 0 {
		opts.LineDataSymbolLineSize = float64(valToPts(0.75))
	} else {
		opts.LineDataSymbolLineSize = float64(valToPts(opts.LineDataSymbolLineSize))
	}

	// gridlines
	if opts.CatGridLine == nil {
		if chartType == ChartTypeScatter {
			opts.CatGridLine = &OptsChartGridLine{Color: "D9D9D9", Size: 1}
		} else {
			opts.CatGridLine = &OptsChartGridLine{Style: "none"}
		}
	}
	if opts.ValGridLine == nil {
		if chartType == ChartTypeScatter {
			opts.ValGridLine = &OptsChartGridLine{Color: "D9D9D9", Size: 1}
		} else {
			opts.ValGridLine = &OptsChartGridLine{}
		}
	}
	if opts.SerGridLine == nil {
		if chartType == ChartTypeScatter {
			opts.SerGridLine = &OptsChartGridLine{Color: "D9D9D9", Size: 1}
		} else {
			opts.SerGridLine = &OptsChartGridLine{Style: "none"}
		}
	}
	correctShadowOptions(opts.Shadow)

	// axis line show defaults (true)
	if opts.CatAxisLineShow == nil {
		opts.CatAxisLineShow = ptr(true)
	}
	if opts.ValAxisLineShow == nil {
		opts.ValAxisLineShow = ptr(true)
	}
	if opts.SerAxisLineShow == nil {
		opts.SerAxisLineShow = ptr(true)
	}

	// 3D
	opts.V3DRotX = 30
	opts.V3DRotY = 30
	opts.V3DPerspective = 30

	// gaps
	if !(opts.BarGapWidthPct >= 0 && opts.BarGapWidthPct <= 1000) || opts.BarGapWidthPct == 0 {
		opts.BarGapWidthPct = 150
	}
	if !(opts.BarGapDepthPct >= 0 && opts.BarGapDepthPct <= 1000) || opts.BarGapDepthPct == 0 {
		opts.BarGapDepthPct = 150
	}

	// chartColors
	if opts.ChartColors == nil {
		if chartType == ChartTypePie || chartType == ChartTypeDoughnut {
			opts.ChartColors = PIECHART_COLORS
		} else {
			opts.ChartColors = BARCHART_COLORS
		}
	}

	// plotArea / chartArea
	if opts.PlotArea == nil {
		opts.PlotArea = &ChartFillLineProps{}
	}
	if opts.PlotArea.Fill == nil {
		opts.PlotArea.Fill = &ShapeFillProps{}
	}
	if opts.ChartArea == nil {
		opts.ChartArea = &ChartAreaProps{}
	}
	if opts.ChartArea.RoundedCorners == nil {
		opts.ChartArea.RoundedCorners = ptr(true)
	}

	// dataLabelFormatCode
	if opts.DataLabelFormatCode == "" && chartType == ChartTypeScatter {
		opts.DataLabelFormatCode = "General"
	}
	if opts.DataLabelFormatCode == "" && (chartType == ChartTypePie || chartType == ChartTypeDoughnut) {
		if chartBool(opts.ShowPercent) {
			opts.DataLabelFormatCode = "0%"
		} else {
			opts.DataLabelFormatCode = "General"
		}
	}
	if opts.DataLabelFormatCode == "" {
		opts.DataLabelFormatCode = "#,##0"
	}

	if opts.DataLabelFormatScatter == "" && chartType == ChartTypeScatter {
		opts.DataLabelFormatScatter = "custom"
	}

	if opts.LineSize == 0 {
		opts.LineSize = 2
	}

	if chartType == ChartTypeArea || chartType == ChartTypeBar || chartType == ChartTypeBar3d || chartType == ChartTypeLine {
		if opts.CatAxisMultiLevelLabels == nil {
			opts.CatAxisMultiLevelLabels = ptr(false)
		}
	}
}

// newChartRel builds a fully-defaulted SlideRelChart for tests.
func newChartRel(chartType ChartType, data []ChartData, userOpts ChartOptions) *SlideRelChart {
	opts := userOpts
	applyChartDefaults(chartType, data, &opts)
	return &SlideRelChart{
		Type: chartType,
		Opts: &opts,
		Data: data,
	}
}

// ---------------------------------------------------------------------------
// Golden: makeXmlCharts
// ---------------------------------------------------------------------------

func goldenPath(parts ...string) string {
	return filepath.Join(append([]string{"testdata", "golden"}, parts...)...)
}

func readGolden(t *testing.T, parts ...string) []byte {
	t.Helper()
	b, err := os.ReadFile(goldenPath(parts...))
	if err != nil {
		t.Fatalf("read golden %v: %v", parts, err)
	}
	return b
}

func assertEqualXML(t *testing.T, got, want string) {
	t.Helper()
	if got == want {
		return
	}
	// Find first divergence for a helpful message.
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	div := n
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			div = i
			break
		}
	}
	lo := div - 80
	if lo < 0 {
		lo = 0
	}
	gHi := div + 80
	if gHi > len(got) {
		gHi = len(got)
	}
	wHi := div + 80
	if wHi > len(want) {
		wHi = len(want)
	}
	t.Fatalf("XML mismatch at byte %d (got %d bytes, want %d bytes)\n GOT ...%q...\nWANT ...%q...",
		div, len(got), len(want), got[lo:gHi], want[lo:wHi])
}

func TestMakeXmlChartsBarGolden(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "Revenue", Labels: [][]string{{"Q1", "Q2", "Q3", "Q4"}}, Values: []float64{100, 150, 130, 175}},
	}
	rel := newChartRel(ChartTypeBar, data, ChartOptions{
		PositionProps: PositionProps{X: ptr(Inches(0.5)), Y: ptr(Inches(0.5)), W: ptr(Inches(9)), H: ptr(Inches(5))},
		ShowTitle:     ptr(true), Title: "Quarterly Revenue",
		ShowCatAxisTitle: ptr(true), CatAxisTitle: "Quarter",
		ShowValAxisTitle: ptr(true), ValAxisTitle: "USD (thousands)",
		ChartColors: []string{"2E86AB", "A23B72", "F18F01", "4CAF50"},
	})
	got := makeXmlCharts(rel)
	want := string(readGolden(t, "05-chart-bar", "ppt", "charts", "chart1.xml"))
	assertEqualXML(t, got, want)
}

func TestMakeXmlChartsLineGolden(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "Series A", Labels: [][]string{{"Jan", "Feb", "Mar", "Apr"}}, Values: []float64{10, 20, 15, 25}},
		{DataIndex: 1, Name: "Series B", Labels: [][]string{{"Jan", "Feb", "Mar", "Apr"}}, Values: []float64{5, 12, 18, 9}},
	}
	rel := newChartRel(ChartTypeLine, data, ChartOptions{
		PositionProps: PositionProps{X: ptr(Inches(0.5)), Y: ptr(Inches(0.5)), W: ptr(Inches(9)), H: ptr(Inches(5))},
		ShowTitle:     ptr(true), Title: "Two Series Line Chart",
		ShowLegend:  ptr(true),
		ChartColors: []string{"2E86AB", "A23B72"},
	})
	got := makeXmlCharts(rel)
	want := string(readGolden(t, "06-chart-multi", "ppt", "charts", "chart2.xml"))
	assertEqualXML(t, got, want)
}

func TestMakeXmlChartsPieGolden(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "Share", Labels: [][]string{{"Alpha", "Beta", "Gamma"}}, Values: []float64{40, 35, 25}},
	}
	rel := newChartRel(ChartTypePie, data, ChartOptions{
		PositionProps: PositionProps{X: ptr(Inches(1)), Y: ptr(Inches(0.5)), W: ptr(Inches(7)), H: ptr(Inches(5))},
		ShowTitle:     ptr(true), Title: "Market Share",
		ShowLegend:  ptr(true),
		ChartColors: []string{"2E86AB", "A23B72", "F18F01"},
	})
	got := makeXmlCharts(rel)
	want := string(readGolden(t, "06-chart-multi", "ppt", "charts", "chart3.xml"))
	assertEqualXML(t, got, want)
}

// ---------------------------------------------------------------------------
// Golden: createExcelWorksheet (compare each internal part)
// ---------------------------------------------------------------------------

var coreTimestampRe = regexp.MustCompile(`(<dcterms:(?:created|modified)[^>]*>)[^<]*(</dcterms:(?:created|modified)>)`)

func normalizeCore(s string) string {
	return coreTimestampRe.ReplaceAllString(s, `${1}TS${2}`)
}

func unzipParts(t *testing.T, data []byte) map[string]string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	out := map[string]string{}
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "/") {
			continue // directory entry
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		out[f.Name] = string(b)
	}
	return out
}

func compareXlsx(t *testing.T, goldenXlsx string, rel *SlideRelChart) {
	t.Helper()
	// Fixed timestamp for determinism (core.xml is compared with TS normalized).
	prev := excelNowFunc
	excelNowFunc = func() time.Time { return time.Date(2026, 7, 20, 4, 28, 22, 542000000, time.UTC) }
	defer func() { excelNowFunc = prev }()

	gotBytes, err := createExcelWorksheet(rel)
	if err != nil {
		t.Fatalf("createExcelWorksheet: %v", err)
	}
	gotParts := unzipParts(t, gotBytes)
	wantParts := unzipParts(t, string2bytes(t, goldenXlsx))

	for name, want := range wantParts {
		got, ok := gotParts[name]
		if !ok {
			t.Errorf("missing part %q in generated xlsx", name)
			continue
		}
		if name == "docProps/core.xml" {
			if normalizeCore(got) != normalizeCore(want) {
				t.Errorf("part %q (timestamp-normalized) mismatch:\n got %q\nwant %q", name, normalizeCore(got), normalizeCore(want))
			}
			continue
		}
		if got != want {
			assertEqualNamed(t, name, got, want)
		}
	}
	// Ensure we didn't produce extra non-directory parts the golden lacks.
	for name := range gotParts {
		if _, ok := wantParts[name]; !ok {
			t.Errorf("unexpected extra part %q in generated xlsx", name)
		}
	}
}

func string2bytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func assertEqualNamed(t *testing.T, name, got, want string) {
	t.Helper()
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	div := n
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			div = i
			break
		}
	}
	lo := div - 60
	if lo < 0 {
		lo = 0
	}
	gHi := div + 60
	if gHi > len(got) {
		gHi = len(got)
	}
	wHi := div + 60
	if wHi > len(want) {
		wHi = len(want)
	}
	t.Errorf("part %q mismatch at byte %d (got %d, want %d)\n GOT ...%q...\nWANT ...%q...",
		name, div, len(got), len(want), got[lo:gHi], want[lo:wHi])
}

func TestCreateExcelWorksheetBarGolden(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "Revenue", Labels: [][]string{{"Q1", "Q2", "Q3", "Q4"}}, Values: []float64{100, 150, 130, 175}},
	}
	rel := newChartRel(ChartTypeBar, data, ChartOptions{
		ChartColors: []string{"2E86AB", "A23B72", "F18F01", "4CAF50"},
	})
	compareXlsx(t, goldenPath("05-chart-bar", "ppt", "embeddings", "Microsoft_Excel_Worksheet1.xlsx"), rel)
}

func TestCreateExcelWorksheetPieGolden(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "Share", Labels: [][]string{{"Alpha", "Beta", "Gamma"}}, Values: []float64{40, 35, 25}},
	}
	rel := newChartRel(ChartTypePie, data, ChartOptions{
		ChartColors: []string{"2E86AB", "A23B72", "F18F01"},
	})
	compareXlsx(t, goldenPath("06-chart-multi", "ppt", "embeddings", "Microsoft_Excel_Worksheet3.xlsx"), rel)
}

func TestCreateExcelWorksheetLineGolden(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "Series A", Labels: [][]string{{"Jan", "Feb", "Mar", "Apr"}}, Values: []float64{10, 20, 15, 25}},
		{DataIndex: 1, Name: "Series B", Labels: [][]string{{"Jan", "Feb", "Mar", "Apr"}}, Values: []float64{5, 12, 18, 9}},
	}
	rel := newChartRel(ChartTypeLine, data, ChartOptions{
		ChartColors: []string{"2E86AB", "A23B72"},
	})
	compareXlsx(t, goldenPath("06-chart-multi", "ppt", "embeddings", "Microsoft_Excel_Worksheet2.xlsx"), rel)
}

// ---------------------------------------------------------------------------
// Focused unit tests (features not covered by golden)
// ---------------------------------------------------------------------------

func TestGetExcelColName(t *testing.T) {
	cases := map[int]string{1: "A", 2: "B", 26: "Z", 27: "AA", 28: "AB", 52: "AZ", 53: "BA"}
	for in, want := range cases {
		if got := getExcelColName(in); got != want {
			t.Errorf("getExcelColName(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestCreateLineCap(t *testing.T) {
	cases := map[string]string{"": "flat", "flat": "flat", "square": "sq", "round": "rnd", "bogus": "flat"}
	for in, want := range cases {
		if got := createLineCap(in); got != want {
			t.Errorf("createLineCap(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestScatterUniqueID(t *testing.T) {
	cases := map[int]string{1: "00000001", 12: "00000012", 123456789: "123456789"}
	for in, want := range cases {
		if got := scatterUniqueID(in); got != want {
			t.Errorf("scatterUniqueID(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestCreateShadowElementNil(t *testing.T) {
	if got := createShadowElement(nil, DEF_SHAPE_SHADOW); got != "<a:effectLst/>" {
		t.Errorf("nil shadow = %q", got)
	}
}

func TestCreateShadowElementOuter(t *testing.T) {
	sh := &ShadowProps{Type: "outer", Blur: 3, Offset: 23000.0 / 12700.0, Angle: 90, Color: "000000", Opacity: 0.35, RotateWithShape: ptr(true)}
	got := createShadowElement(sh, DEF_SHAPE_SHADOW)
	want := `<a:effectLst><a:outerShdw sx="100000" sy="100000" kx="0" ky="0"  algn="bl" blurRad="38100" rotWithShape="1" dist="23000" dir="5400000"><a:srgbClr val="000000"><a:alpha val="35000"/></a:srgbClr></a:outerShdw></a:effectLst>`
	if got != want {
		t.Errorf("shadow mismatch:\n got %q\nwant %q", got, want)
	}
}

func TestGridLineDefaults(t *testing.T) {
	got := createGridLineElement(&OptsChartGridLine{})
	want := `<c:majorGridlines> <c:spPr>  <a:ln w="12700" cap="flat">  <a:solidFill><a:srgbClr val="888888"/></a:solidFill>   <a:prstDash val="solid"/><a:round/>  </a:ln> </c:spPr></c:majorGridlines>`
	if got != want {
		t.Errorf("gridline mismatch:\n got %q\nwant %q", got, want)
	}
}

func TestGenXmlTitleAxisNoSize(t *testing.T) {
	// FontSize 0 -> sizeAttr empty -> double space before b= (matches golden axis titles).
	got := genXmlTitle(chartTitleOpts{Title: "Quarter", FontSize: 0}, 0, 0)
	if !strings.Contains(got, `<a:defRPr  b="0" i="0" u="none" strike="noStrike">`) {
		t.Errorf("expected empty size-attr double-space form, got:\n%s", got)
	}
	if !strings.Contains(got, `<a:t>Quarter</a:t>`) {
		t.Errorf("missing title text, got:\n%s", got)
	}
}

func TestMakeXmlChartsBar3DAxes(t *testing.T) {
	data := []ChartData{{DataIndex: 0, Name: "S", Labels: [][]string{{"A", "B"}}, Values: []float64{1, 2}}}
	rel := newChartRel(ChartTypeBar3d, data, ChartOptions{ShowTitle: ptr(false)})
	got := makeXmlCharts(rel)
	for _, frag := range []string{
		// v3DRAngAx default is undefined (TS "default true" is a no-op), so
		// `!undefined ? 0 : 1` => rAngAx val="0".
		`<c:view3D><c:rotX val="30"/><c:rotY val="30"/><c:rAngAx val="0"/><c:perspective val="30"/></c:view3D>`,
		`<c:bar3DChart>`,
		`<c:gapDepth val="150"/>`,
		`<c:shape val="box"/>`,
		`<c:serAx>`,
		`<c:axId val="` + AXIS_ID_SERIES_PRIMARY + `"/>`,
	} {
		if !strings.Contains(got, frag) {
			t.Errorf("bar3D missing fragment %q", frag)
		}
	}
}

func TestMakeXmlChartsDoughnutHoleSize(t *testing.T) {
	data := []ChartData{{DataIndex: 0, Name: "S", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{1, 2, 3}}}
	rel := newChartRel(ChartTypeDoughnut, data, ChartOptions{})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:doughnutChart>`) {
		t.Errorf("missing doughnutChart")
	}
	if !strings.Contains(got, `<c:holeSize val="50"/>`) {
		t.Errorf("missing default holeSize=50")
	}
	// Doughnut has no dLblPos ctr (that is pie-only).
	if strings.Contains(got, `<c:dLblPos val="ctr"/>`) {
		t.Errorf("doughnut should not emit dLblPos ctr")
	}
}

func TestMakeXmlChartsDoughnutHoleSizeCustom(t *testing.T) {
	data := []ChartData{{DataIndex: 0, Name: "S", Labels: [][]string{{"A", "B"}}, Values: []float64{1, 2}}}
	rel := newChartRel(ChartTypeDoughnut, data, ChartOptions{HoleSize: 75})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:holeSize val="75"/>`) {
		t.Errorf("missing custom holeSize=75, got holeSize region")
	}
}

func TestMakeXmlChartsRadar(t *testing.T) {
	data := []ChartData{{DataIndex: 0, Name: "S", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{1, 2, 3}}}
	rel := newChartRel(ChartTypeRadar, data, ChartOptions{})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:radarChart>`) {
		t.Errorf("missing radarChart")
	}
	if !strings.Contains(got, `<c:radarStyle val="standard"/>`) {
		t.Errorf("missing radarStyle")
	}
	// Radar must NOT include the per-series dLbls block (would corrupt the chart),
	// but must include the marker block.
	if !strings.Contains(got, `<c:marker>`) {
		t.Errorf("radar missing marker")
	}
}

func TestMakeXmlChartsScatter(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "X-Axis", Labels: [][]string{{"", "", ""}}, Values: []float64{1, 2, 3}},
		{DataIndex: 1, Name: "Y1", Labels: [][]string{{"", "", ""}}, Values: []float64{2, 4, 6}},
	}
	rel := newChartRel(ChartTypeScatter, data, ChartOptions{})
	got := makeXmlCharts(rel)
	for _, frag := range []string{
		`<c:scatterChart>`,
		`<c:scatterStyle val="lineMarker"/>`,
		`<c:xVal>`,
		`<c:yVal>`,
		`<c:showDLblsOverMax val="1"/>`,
		// Scatter cat axis is emitted as a valAx (numeric X).
		`<c:valAx>`,
	} {
		if !strings.Contains(got, frag) {
			t.Errorf("scatter missing fragment %q", frag)
		}
	}
	// Scatter cat/val gridlines default to D9D9D9.
	if !strings.Contains(got, `<a:srgbClr val="D9D9D9"/>`) {
		t.Errorf("scatter missing D9D9D9 gridlines")
	}
}

func TestMakeXmlChartsBubble(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "X-Axis", Labels: [][]string{{"", "", ""}}, Values: []float64{1, 2, 3}},
		{DataIndex: 1, Name: "Y1", Labels: [][]string{{"", "", ""}}, Values: []float64{2, 4, 6}, Sizes: []float64{5, 6, 7}},
	}
	rel := newChartRel(ChartTypeBubble, data, ChartOptions{})
	got := makeXmlCharts(rel)
	for _, frag := range []string{
		`<c:bubbleChart>`,
		`<c:bubbleSize>`,
		`<c:bubble3D val="0"/>`,
	} {
		if !strings.Contains(got, frag) {
			t.Errorf("bubble missing fragment %q", frag)
		}
	}
}

func TestMakeXmlChartsBubble3D(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "X-Axis", Labels: [][]string{{"", ""}}, Values: []float64{1, 2}},
		{DataIndex: 1, Name: "Y1", Labels: [][]string{{"", ""}}, Values: []float64{2, 4}, Sizes: []float64{5, 6}},
	}
	rel := newChartRel(ChartTypeBubble3d, data, ChartOptions{})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:bubble3D val="1"/>`) {
		t.Errorf("bubble3D must set bubble3D=1")
	}
}

// createExcelWorksheet for scatter uses the scatter sheet layout.
func TestCreateExcelWorksheetScatterShape(t *testing.T) {
	data := []ChartData{
		{DataIndex: 0, Name: "X-Axis", Labels: [][]string{{"", "", ""}}, Values: []float64{1, 2, 3}},
		{DataIndex: 1, Name: "Y1", Labels: [][]string{{"", "", ""}}, Values: []float64{2, 4, 6}},
	}
	rel := newChartRel(ChartTypeScatter, data, ChartOptions{})
	b, err := createExcelWorksheet(rel)
	if err != nil {
		t.Fatalf("createExcelWorksheet: %v", err)
	}
	parts := unzipParts(t, b)
	sheet := parts["xl/worksheets/sheet1.xml"]
	// scatter header: one t="s" cell per series, X-Values then Y col.
	if !strings.Contains(sheet, `<c r="A1" t="s"><v>0</v></c>`) {
		t.Errorf("scatter sheet missing header A1, got:\n%s", sheet)
	}
	ss := parts["xl/sharedStrings.xml"]
	if !strings.Contains(ss, `count="2" uniqueCount="2"`) {
		t.Errorf("scatter sharedStrings count wrong, got:\n%s", ss)
	}
	tbl := parts["xl/tables/table1.xml"]
	if !strings.Contains(tbl, `name="X-Values0"`) || !strings.Contains(tbl, `name="Y-Value 1"`) {
		t.Errorf("scatter table columns wrong, got:\n%s", tbl)
	}
}
