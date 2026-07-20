package pptx

// Regression tests for the "explicit numeric zero vs unset" family of review
// findings (REVIEW.md C4, M2, M3, M4, M6, M8 and related minors). Each test
// exercises a field where JS distinguishes an explicitly-supplied 0/false from
// `undefined`, which the Go port now models with pointers.

import (
	"strings"
	"testing"
)

func chartData3() []ChartData {
	return []ChartData{{DataIndex: 0, Name: "S", Labels: [][]string{{"A", "B", "C"}}, Values: []float64{40, 35, 25}}}
}

// --- C4: chart 3D + bar-gap + line-size explicit zero ------------------------

func TestExplicitZero_V3DRotXSurvives(t *testing.T) {
	rel := newChartRel(ChartTypeBar3d, chartData3(), ChartOptions{V3DRotX: ptr(0.0)})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:rotX val="0"/>`) {
		t.Errorf("explicit V3DRotX:0 not honored; want <c:rotX val=\"0\"/> in output")
	}
	if strings.Contains(got, `<c:rotX val="30"/>`) {
		t.Errorf("explicit V3DRotX:0 was replaced by default 30")
	}
}

func TestExplicitZero_V3DRotYAndPerspectiveSurvive(t *testing.T) {
	rel := newChartRel(ChartTypeBar3d, chartData3(), ChartOptions{V3DRotY: ptr(0.0), V3DPerspective: ptr(0.0)})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:rotY val="0"/>`) {
		t.Errorf("explicit V3DRotY:0 not honored")
	}
	if !strings.Contains(got, `<c:perspective val="0"/>`) {
		t.Errorf("explicit V3DPerspective:0 not honored")
	}
}

func TestExplicitZero_BarGapWidthPctSurvives(t *testing.T) {
	rel := newChartRel(ChartTypeBar, chartData3(), ChartOptions{BarGapWidthPct: ptr(0.0)})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:gapWidth val="0"/>`) {
		t.Errorf("explicit BarGapWidthPct:0 not honored; want <c:gapWidth val=\"0\"/>")
	}
	if strings.Contains(got, `<c:gapWidth val="150"/>`) {
		t.Errorf("explicit BarGapWidthPct:0 was replaced by default 150")
	}
}

func TestExplicitZero_BarGapDepthPctSurvives(t *testing.T) {
	rel := newChartRel(ChartTypeBar3d, chartData3(), ChartOptions{BarGapDepthPct: ptr(0.0)})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:gapDepth val="0"/>`) {
		t.Errorf("explicit BarGapDepthPct:0 not honored; want <c:gapDepth val=\"0\"/>")
	}
}

func TestExplicitZero_LineSizeRendersNoFill(t *testing.T) {
	rel := newChartRel(ChartTypeLine, chartData3(), ChartOptions{LineSize: ptr(0.0)})
	got := makeXmlCharts(rel)
	// gen-charts.ts:860: lineSize===0 -> a noFill line.
	if !strings.Contains(got, `<a:ln><a:noFill/></a:ln>`) {
		t.Errorf("explicit LineSize:0 should render a noFill line (<a:ln><a:noFill/></a:ln>)")
	}
}

func TestExplicitZero_LineSizeDefaultsToTwoWhenUnset(t *testing.T) {
	rel := newChartRel(ChartTypeLine, chartData3(), ChartOptions{})
	got := makeXmlCharts(rel)
	// Unset LineSize must default to 2pt => a series line at w=valToPts(2)=25400.
	if !strings.Contains(got, `<a:ln w="25400" cap=`) {
		t.Errorf("unset LineSize should render a 2pt series line (w=25400)")
	}
}

// --- M2: doughnut holeSize explicit zero -------------------------------------

func TestExplicitZero_HoleSizeSurvives(t *testing.T) {
	rel := newChartRel(ChartTypeDoughnut, chartData3(), ChartOptions{HoleSize: ptr(0.0)})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:holeSize val="0"/>`) {
		t.Errorf("explicit HoleSize:0 not honored; want <c:holeSize val=\"0\"/>")
	}
	if strings.Contains(got, `<c:holeSize val="50"/>`) {
		t.Errorf("explicit HoleSize:0 was replaced by default 50")
	}
}

func TestExplicitZero_HoleSizeDefaultsWhenUnset(t *testing.T) {
	rel := newChartRel(ChartTypeDoughnut, chartData3(), ChartOptions{})
	got := makeXmlCharts(rel)
	if !strings.Contains(got, `<c:holeSize val="50"/>`) {
		t.Errorf("unset HoleSize should default to 50")
	}
}

// --- M2: shadow + glow explicit zero -----------------------------------------

func TestExplicitZero_ShadowFieldsHonored(t *testing.T) {
	sh := &ShadowProps{Angle: ptr(0.0), Blur: ptr(0.0), Offset: ptr(0.0), Opacity: ptr(0.0)}
	got := createShadowElement(sh, DEF_SHAPE_SHADOW)
	for _, want := range []string{`dir="0"`, `blurRad="0"`, `dist="0"`, `<a:alpha val="0"/>`} {
		if !strings.Contains(got, want) {
			t.Errorf("explicit shadow zero not honored; missing %q in %q", want, got)
		}
	}
	// Fields NOT overridden keep the defaults (color, rotateWithShape).
	if !strings.Contains(got, `val="000000"`) || !strings.Contains(got, `rotWithShape="1"`) {
		t.Errorf("unset shadow fields should keep defaults; got %q", got)
	}
}

func TestExplicitZero_GlowFieldsHonored(t *testing.T) {
	got := createGlowElement(TextGlowProps{Size: ptr(0.0), Opacity: ptr(0.0)}, DEF_TEXT_GLOW)
	want := `<a:glow rad="0"><a:srgbClr val="FFFFFF"><a:alpha val="0"/></a:srgbClr></a:glow>`
	if got != want {
		t.Errorf("explicit glow zero not honored:\n got %q\nwant %q", got, want)
	}
}

// --- M2: combo-chart per-type overlay of explicit zero -----------------------

func TestExplicitZero_OverlayLineSizeOverride(t *testing.T) {
	base := &ChartOptions{LineSize: ptr(2.0), BarGapWidthPct: ptr(150.0)}
	over := &ChartOptions{LineSize: ptr(0.0)} // per-type "no line" override
	merged := overlayChartOptions(base, over)
	if merged.LineSize == nil || *merged.LineSize != 0 {
		t.Fatalf("per-type LineSize:0 override lost; merged=%v", merged.LineSize)
	}
	// The base's other fields survive.
	if merged.BarGapWidthPct == nil || *merged.BarGapWidthPct != 150 {
		t.Errorf("base BarGapWidthPct should survive overlay; got %v", merged.BarGapWidthPct)
	}
}

func TestExplicitZero_OverlayBoolPointerOverride(t *testing.T) {
	base := &ChartOptions{ShowValue: ptr(true)}
	over := &ChartOptions{ShowValue: ptr(false)}
	merged := overlayChartOptions(base, over)
	if merged.ShowValue == nil || *merged.ShowValue != false {
		t.Fatalf("explicit ShowValue:false override lost; got %v", merged.ShowValue)
	}
}

// --- M3: ColW scalar shorthand -----------------------------------------------

func TestColW_ScalarShorthandExpandsUniform(t *testing.T) {
	rows := []TableRow{{
		{Type: SlideObjectTypeTablecell, Text: "A"},
		{Type: SlideObjectTypeTablecell, Text: "B"},
		{Type: SlideObjectTypeTablecell, Text: "C"},
	}}
	props := &TableToSlidesProps{TableProps: TableProps{ColW: []float64{2.0}}}
	_ = GetSlidesForTableRows(rows, props, testLayout, nil)
	// After STEP 5 the single scalar is expanded to one entry per column.
	if len(props.ColW) != 3 {
		t.Fatalf("scalar ColW should expand to 3 columns, got %d (%v)", len(props.ColW), props.ColW)
	}
	total := 0.0
	for _, w := range props.ColW {
		if w != 2.0 {
			t.Errorf("each column should be 2in, got %v", w)
		}
		total += w
	}
	if total != 6.0 {
		t.Errorf("total width should be 6in, got %v", total)
	}
}

// --- M2: table paging AutoPageSlideStartY explicit zero ----------------------

func TestExplicitZero_AutoPageSlideStartY(t *testing.T) {
	// 100 short rows. With AutoPageSlideStartY:0 (explicit), the page-3+ budget
	// grows to (height - inch2Emu(0 + margin[2])) => 34 rows/page, giving the
	// TS-verified distribution [32,32,34,2]. The old float64 code treated 0 as
	// unset and produced [32,32,32,4].
	rows := manyShortRows(100)
	slides := GetSlidesForTableRows(rows, &TableToSlidesProps{
		TableProps: TableProps{AutoPageSlideStartY: ptr(0.0)},
	}, testLayout, nil)
	got := make([]int, len(slides))
	for i, s := range slides {
		got[i] = len(s.Rows)
	}
	want := []int{32, 32, 34, 2}
	if len(got) != len(want) {
		t.Fatalf("row distribution = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("row distribution = %v, want %v (TS honors explicit AutoPageSlideStartY:0)", got, want)
		}
	}
}

// --- N (minor): empty row is skipped, not crashed ----------------------------

func TestGetSlidesForTableRows_EmptyRowSkipped(t *testing.T) {
	rows := []TableRow{
		shortRow("1"),
		{}, // empty row: upstream TS crashes; Go skips it (documented deviation)
		shortRow("2"),
	}
	var slides []TableRowSlide
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("empty row should be skipped, not panic: %v", r)
			}
		}()
		slides = GetSlidesForTableRows(rows, &TableToSlidesProps{}, testLayout, nil)
	}()
	if totalRows(slides) != 2 {
		t.Errorf("empty row should be skipped, expected 2 rows total, got %d", totalRows(slides))
	}
}

// --- M4: empty (non-nil) ChartColors / CatAxes must not panic ----------------

func TestEmptyChartColors_NoPanic_SameAsNil(t *testing.T) {
	data := chartData3()
	nilColors := makeXmlCharts(newChartRel(ChartTypeBar, data, ChartOptions{}))
	emptyColors := makeXmlCharts(newChartRel(ChartTypeBar, chartData3(), ChartOptions{ChartColors: []string{}}))
	if emptyColors != nilColors {
		t.Errorf("empty ChartColors should render identically to nil (default palette)")
	}
}

func TestEmptyCatAxes_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("empty CatAxes should not panic: %v", r)
		}
	}()
	rel := newChartRel(ChartTypeBar, chartData3(), ChartOptions{CatAxes: []ChartOptions{}, ValAxes: []ChartOptions{}})
	_ = makeXmlCharts(rel)
}

// --- M6: serAxis time-unit is never emitted (upstream bug replicated) --------

func TestSerAxisTimeUnit_NeverEmitted(t *testing.T) {
	rel := newChartRel(ChartTypeBar3d, chartData3(), ChartOptions{
		SerLabelFormatCode:   "yyyy",
		SerAxisBaseTimeUnit:  "years",
		SerAxisMajorTimeUnit: "years",
		SerAxisMinorTimeUnit: "months",
	})
	got := makeXmlCharts(rel)
	for _, banned := range []string{"baseTimeUnit", "majorTimeUnit", "minorTimeUnit"} {
		if strings.Contains(got, banned) {
			t.Errorf("serAxis %q must never be emitted (gen-charts.ts:1883-1893 upstream bug)", banned)
		}
	}
}
