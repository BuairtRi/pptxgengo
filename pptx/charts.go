// charts.go ports src/gen-charts.ts: the chart XML generator (makeXmlCharts and
// its makeChartType/axis/title/shadow/gridline helpers) plus the embedded XLSX
// data-source builder (createExcelWorksheet).
//
// Fidelity note: XML is produced by string concatenation mirroring the TS source
// exactly (element/attribute order, whitespace, CRLF/LF, number formatting via
// ftoa/jsRound). Do not "clean up" the whitespace — byte-identical output to the
// JS library is the goal.
package pptx

import (
	"archive/zip"
	"bytes"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// small local helpers (unexported, distinctive names to avoid cross-module
// collisions with other in-flight generator files)
// ---------------------------------------------------------------------------

// chartBool reports the JS truthiness of an optional (pointer) bool: nil or
// *false is false, *true is true.
func chartBool(p *bool) bool { return p != nil && *p }

// chartB2S returns "1"/"0" for a *bool used as a numeric XML attribute.
func chartB2S(p *bool) string {
	if chartBool(p) {
		return "1"
	}
	return "0"
}

// chartColorSel wraps a bare color string in a solid-fill selection, mirroring
// the TS `genXmlColorSelection(colorString)` overload (foundation's Go version
// only accepts *ShapeFillProps).
func chartColorSel(color string) string {
	return genXmlColorSelection(&ShapeFillProps{Color: color})
}

// chartColorsEqual reports whether two color slices are element-wise equal. Used
// to detect the built-in BARCHART_COLORS/PIECHART_COLORS default (the TS code
// uses reference identity, e.g. `opts.chartColors !== BARCHART_COLORS`).
func chartColorsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// itoa is a tiny alias for readability in the many int->string interpolations.
func itoa(v int) string { return strconv.Itoa(v) }

// chartTitleOpts mirrors IChartPropsTitle (the arg to genXmlTitle).
type chartTitleOpts struct {
	Title       string
	Color       string
	FontFace    string
	FontSize    float64
	TitleAlign  string
	TitleBold   *bool
	TitlePos    *ChartTitlePos
	TitleRotate float64
}

// ---------------------------------------------------------------------------
// makeXmlCharts
// ---------------------------------------------------------------------------

// makeXmlCharts is the main entry point: it builds the full chart XML for a
// chart relationship. Ports gen-charts.ts makeXmlCharts.
// validateChartConfig replicates the two combo-chart validation throws in
// gen-charts.ts:625-632. Non-pie/doughnut charts with multiple value axes must
// have at least one series that targets the secondary value axis, and the count
// of category axes must match the count of value axes. Returns an error instead
// of panicking (Go idiom; TS throws at render time).
func validateChartConfig(opts *ChartOptions) error {
	if opts.Type == ChartTypePie || opts.Type == ChartTypeDoughnut {
		return nil
	}
	usesSecondaryValAxis := chartBool(opts.SecondaryValAxis)
	for i := range opts.MultiTypes {
		if opts.MultiTypes[i].Options != nil && chartBool(opts.MultiTypes[i].Options.SecondaryValAxis) {
			usesSecondaryValAxis = true
		}
	}
	if len(opts.ValAxes) > 1 && !usesSecondaryValAxis {
		return errors.New("secondary axis must be used by one of the multiple charts")
	}
	if len(opts.CatAxes) > 0 && len(opts.ValAxes) != len(opts.CatAxes) {
		return errors.New("there must be the same number of value and category axes")
	}
	return nil
}

func makeXmlCharts(rel *SlideRelChart) string {
	opts := rel.Opts
	var strXml strings.Builder
	strXml.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	usesSecondaryValAxis := false

	// STEP 1: Create chart
	strXml.WriteString(`<c:chartSpace xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	strXml.WriteString(`<c:date1904 val="0"/>`)
	roundedCorners := "0"
	if opts.ChartArea != nil && chartBool(opts.ChartArea.RoundedCorners) {
		roundedCorners = "1"
	}
	strXml.WriteString(`<c:roundedCorners val="` + roundedCorners + `"/>`)
	strXml.WriteString(`<c:chart>`)

	// OPTION: Title
	if chartBool(opts.ShowTitle) {
		title := opts.Title
		if title == "" {
			title = "Chart Title"
		}
		fontSize := opts.TitleFontSize
		if fontSize == 0 {
			fontSize = float64(DEF_FONT_TITLE_SIZE)
		}
		strXml.WriteString(genXmlTitle(chartTitleOpts{
			Title:       title,
			Color:       opts.TitleColor,
			FontFace:    opts.TitleFontFace,
			FontSize:    fontSize,
			TitleAlign:  opts.TitleAlign,
			TitleBold:   opts.TitleBold,
			TitlePos:    opts.TitlePos,
			TitleRotate: opts.TitleRotate,
		}, coordVal(opts.X), coordVal(opts.Y)))
		strXml.WriteString(`<c:autoTitleDeleted val="0"/>`)
	} else {
		strXml.WriteString(`<c:autoTitleDeleted val="1"/>`)
	}

	// Add 3D view tag
	if opts.Type == ChartTypeBar3d {
		rAngAx := "1"
		if !chartBool(opts.V3DRAngAx) {
			rAngAx = "0"
		}
		strXml.WriteString(`<c:view3D><c:rotX val="` + ftoa(fptrOr(opts.V3DRotX, 30)) + `"/><c:rotY val="` + ftoa(fptrOr(opts.V3DRotY, 30)) + `"/><c:rAngAx val="` + rAngAx + `"/><c:perspective val="` + ftoa(fptrOr(opts.V3DPerspective, 30)) + `"/></c:view3D>`)
	}

	strXml.WriteString(`<c:plotArea>`)
	if opts.Layout != nil {
		strXml.WriteString(`<c:layout>`)
		strXml.WriteString(` <c:manualLayout>`)
		strXml.WriteString(`  <c:layoutTarget val="inner" />`)
		strXml.WriteString(`  <c:xMode val="edge" />`)
		strXml.WriteString(`  <c:yMode val="edge" />`)
		strXml.WriteString(`  <c:x val="` + ftoa(coordVal(opts.Layout.X)) + `" />`)
		strXml.WriteString(`  <c:y val="` + ftoa(coordVal(opts.Layout.Y)) + `" />`)
		strXml.WriteString(`  <c:w val="` + ftoa(coordValOr(opts.Layout.W, 1)) + `" />`)
		strXml.WriteString(`  <c:h val="` + ftoa(coordValOr(opts.Layout.H, 1)) + `" />`)
		strXml.WriteString(` </c:manualLayout>`)
		strXml.WriteString(`</c:layout>`)
	} else {
		strXml.WriteString(`<c:layout/>`)
	}

	// A: Create Chart XML
	if len(opts.MultiTypes) > 0 {
		for _, typ := range opts.MultiTypes {
			options := overlayChartOptions(opts, typ.Options)
			valAxisID := AXIS_ID_VALUE_PRIMARY
			if chartBool(options.SecondaryValAxis) {
				valAxisID = AXIS_ID_VALUE_SECONDARY
			}
			catAxisID := AXIS_ID_CATEGORY_PRIMARY
			if chartBool(options.SecondaryCatAxis) {
				catAxisID = AXIS_ID_CATEGORY_SECONDARY
			}
			usesSecondaryValAxis = usesSecondaryValAxis || chartBool(options.SecondaryValAxis)
			strXml.WriteString(makeChartType(typ.Type, typ.Data, options, valAxisID, catAxisID, true))
		}
	} else {
		strXml.WriteString(makeChartType(opts.Type, rel.Data, opts, AXIS_ID_VALUE_PRIMARY, AXIS_ID_CATEGORY_PRIMARY, false))
	}

	// B: Axes
	if opts.Type != ChartTypePie && opts.Type != ChartTypeDoughnut {
		if len(opts.ValAxes) > 1 && !usesSecondaryValAxis {
			// TS throws; library code returns best-effort (no panic). Mirror by
			// still emitting — matches JS only when input is valid.
		}

		// M4: guard against an explicitly-empty (non-nil, len 0) slice, which
		// would panic at &CatAxes[0]. JS spreads `{...opts, ...catAxes[0]}` where
		// catAxes[0] is undefined -> harmless no-op; the len>0 check mirrors that.
		if len(opts.CatAxes) > 0 {
			strXml.WriteString(makeCatAxis(overlayChartOptions(opts, &opts.CatAxes[0]), AXIS_ID_CATEGORY_PRIMARY, AXIS_ID_VALUE_PRIMARY))
		} else {
			strXml.WriteString(makeCatAxis(opts, AXIS_ID_CATEGORY_PRIMARY, AXIS_ID_VALUE_PRIMARY))
		}

		if len(opts.ValAxes) > 0 {
			strXml.WriteString(makeValAxis(overlayChartOptions(opts, &opts.ValAxes[0]), AXIS_ID_VALUE_PRIMARY))
			if len(opts.ValAxes) > 1 {
				strXml.WriteString(makeValAxis(overlayChartOptions(opts, &opts.ValAxes[1]), AXIS_ID_VALUE_SECONDARY))
			}
		} else {
			strXml.WriteString(makeValAxis(opts, AXIS_ID_VALUE_PRIMARY))
			if opts.Type == ChartTypeBar3d {
				strXml.WriteString(makeSerAxis(opts, AXIS_ID_SERIES_PRIMARY, AXIS_ID_VALUE_PRIMARY))
			}
		}

		if opts.CatAxes != nil && len(opts.CatAxes) > 1 {
			strXml.WriteString(makeCatAxis(overlayChartOptions(opts, &opts.CatAxes[1]), AXIS_ID_CATEGORY_SECONDARY, AXIS_ID_VALUE_SECONDARY))
		}
	}

	// C: Chart Properties and plotArea Options: Border, Data Table, Fill, Legend
	if chartBool(opts.ShowDataTable) {
		strXml.WriteString(`<c:dTable>`)
		strXml.WriteString(`  <c:showHorzBorder val="` + boolNot01(opts.ShowDataTableHorzBorder) + `"/>`)
		strXml.WriteString(`  <c:showVertBorder val="` + boolNot01(opts.ShowDataTableVertBorder) + `"/>`)
		strXml.WriteString(`  <c:showOutline    val="` + boolNot01(opts.ShowDataTableOutline) + `"/>`)
		strXml.WriteString(`  <c:showKeys       val="` + boolNot01(opts.ShowDataTableKeys) + `"/>`)
		strXml.WriteString(`  <c:spPr>`)
		strXml.WriteString(`    <a:noFill/>`)
		strXml.WriteString(`    <a:ln w="9525" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="tx1"><a:lumMod val="15000"/><a:lumOff val="85000"/></a:schemeClr></a:solidFill><a:round/></a:ln>`)
		strXml.WriteString(`    <a:effectLst/>`)
		strXml.WriteString(`  </c:spPr>`)
		strXml.WriteString(`  <c:txPr>`)
		strXml.WriteString(`   <a:bodyPr rot="0" spcFirstLastPara="1" vertOverflow="ellipsis" vert="horz" wrap="square" anchor="ctr" anchorCtr="1"/>`)
		strXml.WriteString(`   <a:lstStyle/>`)
		strXml.WriteString(`   <a:p>`)
		strXml.WriteString(`     <a:pPr rtl="0">`)
		dtFont := opts.DataTableFontSize
		if dtFont == 0 {
			dtFont = float64(DEF_FONT_SIZE)
		}
		strXml.WriteString(`       <a:defRPr sz="` + itoa(int(jsRound(dtFont*100))) + `" b="0" i="0" u="none" strike="noStrike" kern="1200" baseline="0">`)
		strXml.WriteString(`         <a:solidFill><a:schemeClr val="tx1"><a:lumMod val="65000"/><a:lumOff val="35000"/></a:schemeClr></a:solidFill>`)
		strXml.WriteString(`         <a:latin typeface="+mn-lt"/>`)
		strXml.WriteString(`         <a:ea typeface="+mn-ea"/>`)
		strXml.WriteString(`         <a:cs typeface="+mn-cs"/>`)
		strXml.WriteString(`       </a:defRPr>`)
		strXml.WriteString(`     </a:pPr>`)
		strXml.WriteString(`    <a:endParaRPr lang="en-US"/>`)
		strXml.WriteString(`   </a:p>`)
		strXml.WriteString(` </c:txPr>`)
		strXml.WriteString(`</c:dTable>`)
	}

	strXml.WriteString(`  <c:spPr>`)

	// OPTION: Fill
	if opts.PlotArea != nil && opts.PlotArea.Fill != nil && opts.PlotArea.Fill.Color != "" {
		strXml.WriteString(genXmlColorSelection(opts.PlotArea.Fill))
	} else {
		strXml.WriteString(`<a:noFill/>`)
	}

	// OPTION: Border
	if opts.PlotArea != nil && opts.PlotArea.Border != nil {
		strXml.WriteString(`<a:ln w="` + itoa(valToPts(opts.PlotArea.Border.Pt)) + `" cap="flat">` + chartColorSel(opts.PlotArea.Border.Color) + `</a:ln>`)
	} else {
		strXml.WriteString(`<a:ln><a:noFill/></a:ln>`)
	}

	strXml.WriteString(`    <a:effectLst/>`)
	strXml.WriteString(`  </c:spPr>`)
	strXml.WriteString(`</c:plotArea>`)

	// OPTION: Legend
	if chartBool(opts.ShowLegend) {
		strXml.WriteString(`<c:legend>`)
		strXml.WriteString(`<c:legendPos val="` + opts.LegendPos + `"/>`)
		strXml.WriteString(`<c:overlay val="0"/>`)
		if opts.LegendFontFace != "" || opts.LegendFontSize != 0 || opts.LegendColor != "" {
			strXml.WriteString(`<c:txPr>`)
			strXml.WriteString(`  <a:bodyPr/>`)
			strXml.WriteString(`  <a:lstStyle/>`)
			strXml.WriteString(`  <a:p>`)
			strXml.WriteString(`    <a:pPr>`)
			if opts.LegendFontSize != 0 {
				strXml.WriteString(`<a:defRPr sz="` + itoa(int(jsRound(opts.LegendFontSize*100))) + `">`)
			} else {
				strXml.WriteString(`<a:defRPr>`)
			}
			if opts.LegendColor != "" {
				strXml.WriteString(chartColorSel(opts.LegendColor))
			}
			if opts.LegendFontFace != "" {
				strXml.WriteString(`<a:latin typeface="` + opts.LegendFontFace + `"/>`)
			}
			if opts.LegendFontFace != "" {
				strXml.WriteString(`<a:cs    typeface="` + opts.LegendFontFace + `"/>`)
			}
			strXml.WriteString(`      </a:defRPr>`)
			strXml.WriteString(`    </a:pPr>`)
			strXml.WriteString(`    <a:endParaRPr lang="en-US"/>`)
			strXml.WriteString(`  </a:p>`)
			strXml.WriteString(`</c:txPr>`)
		}
		strXml.WriteString(`</c:legend>`)
	}

	strXml.WriteString(`  <c:plotVisOnly val="1"/>`)
	strXml.WriteString(`  <c:dispBlanksAs val="` + opts.DisplayBlanksAs + `"/>`)
	if opts.Type == ChartTypeScatter {
		strXml.WriteString(`<c:showDLblsOverMax val="1"/>`)
	}

	strXml.WriteString(`</c:chart>`)

	// D: CHARTSPACE SHAPE PROPS
	strXml.WriteString(`<c:spPr>`)
	if opts.ChartArea != nil && opts.ChartArea.Fill != nil && opts.ChartArea.Fill.Color != "" {
		strXml.WriteString(genXmlColorSelection(opts.ChartArea.Fill))
	} else {
		strXml.WriteString(`<a:noFill/>`)
	}
	if opts.ChartArea != nil && opts.ChartArea.Border != nil {
		strXml.WriteString(`<a:ln w="` + itoa(valToPts(opts.ChartArea.Border.Pt)) + `" cap="flat">` + chartColorSel(opts.ChartArea.Border.Color) + `</a:ln>`)
	} else {
		strXml.WriteString(`<a:ln><a:noFill/></a:ln>`)
	}
	strXml.WriteString(`  <a:effectLst/>`)
	strXml.WriteString(`</c:spPr>`)

	// E: DATA (Add relID)
	strXml.WriteString(`<c:externalData r:id="rId1"><c:autoUpdate val="0"/></c:externalData>`)

	// LAST: chartSpace end
	strXml.WriteString(`</c:chartSpace>`)

	return strXml.String()
}

// coordVal returns a Coord pointer's value (0 for nil).
func coordVal(c *Coord) float64 {
	if c == nil {
		return 0
	}
	return c.Val
}

// coordValOr returns a Coord pointer's value, or def when nil/zero (mirrors the
// TS `layout.w || 1` idiom for manual chart layout).
func coordValOr(c *Coord, def float64) float64 {
	if c == nil || c.Val == 0 {
		return def
	}
	return c.Val
}

// boolNot01 mirrors `!opts.x ? 0 : 1` for a *bool (nil/false -> "0", true -> "1").
// The TS data-table show* flags use `!x ? 0 : 1`, i.e. identical to chartB2S,
// kept as a separate name to match the source expression at each call site.
func boolNot01(p *bool) string { return chartB2S(p) }

// ---------------------------------------------------------------------------
// makeChartType
// ---------------------------------------------------------------------------

// makeChartType builds the chart-type-specific XML (`<c:barChart>` etc.). Ports
// gen-charts.ts makeChartType.
func makeChartType(chartType ChartType, data []ChartData, opts *ChartOptions, valAxisID, catAxisID string, isMultiTypeChart bool) string {
	_ = isMultiTypeChart // unused (matches TS: param passed but not consumed)
	colorIndex := -1
	idxColLtr := 1
	var strXml strings.Builder

	switch chartType {
	case ChartTypeArea, ChartTypeBar, ChartTypeBar3d, ChartTypeLine, ChartTypeRadar:
		// 1: Start Chart
		strXml.WriteString(`<c:` + string(chartType) + `Chart>`)
		if chartType == ChartTypeArea && opts.BarGrouping == "stacked" {
			strXml.WriteString(`<c:grouping val="` + opts.BarGrouping + `"/>`)
		}
		if chartType == ChartTypeBar || chartType == ChartTypeBar3d {
			strXml.WriteString(`<c:barDir val="` + opts.BarDir + `"/>`)
			grouping := opts.BarGrouping
			if grouping == "" {
				grouping = "clustered"
			}
			strXml.WriteString(`<c:grouping val="` + grouping + `"/>`)
		}
		if chartType == ChartTypeRadar {
			strXml.WriteString(`<c:radarStyle val="` + opts.RadarStyle + `"/>`)
		}
		strXml.WriteString(`<c:varyColors val="0"/>`)

		// 2: "Series" block for every data row
		for i := range data {
			obj := &data[i]
			colorIndex++
			strXml.WriteString(`<c:ser>`)
			strXml.WriteString(`  <c:idx val="` + itoa(obj.DataIndex) + `"/><c:order val="` + itoa(obj.DataIndex) + `"/>`)
			strXml.WriteString(`  <c:tx>`)
			strXml.WriteString(`    <c:strRef>`)
			strXml.WriteString(`      <c:f>Sheet1!$` + getExcelColName(obj.DataIndex+len(obj.Labels)+1) + `$1</c:f>`)
			strXml.WriteString(`      <c:strCache><c:ptCount val="1"/><c:pt idx="0"><c:v>` + encodeXmlEntities(obj.Name) + `</c:v></c:pt></c:strCache>`)
			strXml.WriteString(`    </c:strRef>`)
			strXml.WriteString(`  </c:tx>`)

			// Fill and Border
			seriesColor := ""
			if opts.ChartColors != nil {
				seriesColor = opts.ChartColors[colorIndex%len(opts.ChartColors)]
			}

			strXml.WriteString(`  <c:spPr>`)
			if seriesColor == "transparent" {
				strXml.WriteString(`<a:noFill/>`)
			} else if opts.ChartColorsOpacity != 0 {
				strXml.WriteString(`<a:solidFill>` + createColorElement(seriesColor, `<a:alpha val="`+itoa(int(jsRound(opts.ChartColorsOpacity*1000)))+`"/>`) + `</a:solidFill>`)
			} else {
				strXml.WriteString(`<a:solidFill>` + createColorElement(seriesColor, "") + `</a:solidFill>`)
			}

			if chartType == ChartTypeLine || chartType == ChartTypeRadar {
				if fptrOr(opts.LineSize, 2) == 0 {
					strXml.WriteString(`<a:ln><a:noFill/></a:ln>`)
				} else {
					lineDash := opts.LineDash
					if lineDash == "" {
						lineDash = "solid"
					}
					strXml.WriteString(`<a:ln w="` + itoa(valToPts(fptrOr(opts.LineSize, 2))) + `" cap="` + createLineCap(opts.LineCap) + `"><a:solidFill>` + createColorElement(seriesColor, "") + `</a:solidFill>`)
					strXml.WriteString(`<a:prstDash val="` + lineDash + `"/><a:round/></a:ln>`)
				}
			} else if opts.DataBorder != nil {
				strXml.WriteString(`<a:ln w="` + itoa(valToPts(opts.DataBorder.Pt)) + `" cap="` + createLineCap(opts.LineCap) + `"><a:solidFill>` + createColorElement(opts.DataBorder.Color, "") + `</a:solidFill><a:prstDash val="solid"/><a:round/></a:ln>`)
			}

			strXml.WriteString(createShadowElement(opts.Shadow, DEF_SHAPE_SHADOW))

			strXml.WriteString(`  </c:spPr>`)
			strXml.WriteString(`  <c:invertIfNegative val="0"/>`)

			// Data Labels per series (not for RADAR)
			if chartType != ChartTypeRadar {
				strXml.WriteString(`<c:dLbls>`)
				strXml.WriteString(`<c:numFmt formatCode="` + numFmtOrGeneral(opts.DataLabelFormatCode) + `" sourceLinked="0"/>`)
				if chartBool(opts.DataLabelBkgrdColors) {
					strXml.WriteString(`<c:spPr><a:solidFill>` + createColorElement(seriesColor, "") + `</a:solidFill></c:spPr>`)
				}
				strXml.WriteString(`<c:txPr><a:bodyPr/><a:lstStyle/><a:p><a:pPr>`)
				strXml.WriteString(`<a:defRPr b="` + chartB2S(opts.DataLabelFontBold) + `" i="` + chartB2S(opts.DataLabelFontItalic) + `" strike="noStrike" sz="` + itoa(int(jsRound(fontSizeOr(opts.DataLabelFontSize, DEF_FONT_SIZE)*100))) + `" u="none">`)
				strXml.WriteString(`<a:solidFill>` + createColorElement(strOr(opts.DataLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
				strXml.WriteString(`<a:latin typeface="` + strOr(opts.DataLabelFontFace, "Arial") + `"/>`)
				strXml.WriteString(`</a:defRPr></a:pPr></a:p></c:txPr>`)
				if opts.DataLabelPosition != "" {
					strXml.WriteString(`<c:dLblPos val="` + opts.DataLabelPosition + `"/>`)
				}
				strXml.WriteString(`<c:showLegendKey val="0"/>`)
				strXml.WriteString(`<c:showVal val="` + chartB2S(opts.ShowValue) + `"/>`)
				strXml.WriteString(`<c:showCatName val="0"/><c:showSerName val="` + chartB2S(opts.ShowSerName) + `"/><c:showPercent val="0"/><c:showBubbleSize val="0"/>`)
				strXml.WriteString(`<c:showLeaderLines val="` + chartB2S(opts.ShowLeaderLines) + `"/>`)
				strXml.WriteString(`</c:dLbls>`)
			}

			// 'c:marker' tag: lineDataSymbol
			if chartType == ChartTypeLine || chartType == ChartTypeRadar {
				strXml.WriteString(`<c:marker>`)
				strXml.WriteString(`  <c:symbol val="` + opts.LineDataSymbol + `"/>`)
				if opts.LineDataSymbolSize != 0 {
					strXml.WriteString(`<c:size val="` + ftoa(opts.LineDataSymbolSize) + `"/>`)
				}
				strXml.WriteString(`  <c:spPr>`)
				strXml.WriteString(`    <a:solidFill>` + createColorElement(opts.ChartColors[markerColorIdx(obj.DataIndex, len(opts.ChartColors))], "") + `</a:solidFill>`)
				strXml.WriteString(`    <a:ln w="` + ftoa(opts.LineDataSymbolLineSize) + `" cap="flat"><a:solidFill>` + createColorElement(strOr(opts.LineDataSymbolLineColor, seriesColor), "") + `</a:solidFill><a:prstDash val="solid"/><a:round/></a:ln>`)
				strXml.WriteString(`    <a:effectLst/>`)
				strXml.WriteString(`  </c:spPr>`)
				strXml.WriteString(`</c:marker>`)
			}

			// Data Point colors (single-series bar with custom colors)
			if (chartType == ChartTypeBar || chartType == ChartTypeBar3d) &&
				len(data) == 1 &&
				((opts.ChartColors != nil && !chartColorsEqual(opts.ChartColors, BARCHART_COLORS) && len(opts.ChartColors) > 1) || len(opts.InvertedColors) > 0) {
				for index, value := range obj.Values {
					arrColors := opts.ChartColors
					if value < 0 {
						if len(opts.InvertedColors) > 0 {
							arrColors = opts.InvertedColors
						} else if opts.ChartColors != nil {
							arrColors = opts.ChartColors
						} else {
							arrColors = BARCHART_COLORS
						}
					} else if opts.ChartColors == nil {
						arrColors = []string{}
					}

					strXml.WriteString(`  <c:dPt>`)
					strXml.WriteString(`    <c:idx val="` + itoa(index) + `"/>`)
					strXml.WriteString(`      <c:invertIfNegative val="0"/>`)
					strXml.WriteString(`    <c:bubble3D val="0"/>`)
					strXml.WriteString(`    <c:spPr>`)
					if fptrOr(opts.LineSize, 2) == 0 {
						strXml.WriteString(`<a:ln><a:noFill/></a:ln>`)
					} else if chartType == ChartTypeBar {
						strXml.WriteString(`<a:solidFill>`)
						strXml.WriteString(`  <a:srgbClr val="` + arrColors[index%len(arrColors)] + `"/>`)
						strXml.WriteString(`</a:solidFill>`)
					} else {
						strXml.WriteString(`<a:ln>`)
						strXml.WriteString(`  <a:solidFill>`)
						strXml.WriteString(`   <a:srgbClr val="` + arrColors[index%len(arrColors)] + `"/>`)
						strXml.WriteString(`  </a:solidFill>`)
						strXml.WriteString(`</a:ln>`)
					}
					strXml.WriteString(createShadowElement(opts.Shadow, DEF_SHAPE_SHADOW))
					strXml.WriteString(`    </c:spPr>`)
					strXml.WriteString(`  </c:dPt>`)
				}
			}

			// 2: "Categories"
			strXml.WriteString(`<c:cat>`)
			if opts.CatLabelFormatCode != "" {
				strXml.WriteString(`  <c:numRef>`)
				strXml.WriteString(`    <c:f>Sheet1!$A$2:$A$` + itoa(len(obj.Labels[0])+1) + `</c:f>`)
				strXml.WriteString(`    <c:numCache>`)
				strXml.WriteString(`      <c:formatCode>` + strOr(opts.CatLabelFormatCode, "General") + `</c:formatCode>`)
				strXml.WriteString(`      <c:ptCount val="` + itoa(len(obj.Labels[0])) + `"/>`)
				for idx, label := range obj.Labels[0] {
					strXml.WriteString(`<c:pt idx="` + itoa(idx) + `"><c:v>` + encodeXmlEntities(label) + `</c:v></c:pt>`)
				}
				strXml.WriteString(`    </c:numCache>`)
				strXml.WriteString(`  </c:numRef>`)
			} else {
				strXml.WriteString(`  <c:multiLvlStrRef>`)
				strXml.WriteString(`    <c:f>Sheet1!$A$2:$` + getExcelColName(len(obj.Labels)) + `$` + itoa(len(obj.Labels[0])+1) + `</c:f>`)
				strXml.WriteString(`    <c:multiLvlStrCache>`)
				strXml.WriteString(`      <c:ptCount val="` + itoa(len(obj.Labels[0])) + `"/>`)
				for _, labelsGroup := range obj.Labels {
					strXml.WriteString(`<c:lvl>`)
					for idx, label := range labelsGroup {
						strXml.WriteString(`<c:pt idx="` + itoa(idx) + `"><c:v>` + encodeXmlEntities(label) + `</c:v></c:pt>`)
					}
					strXml.WriteString(`</c:lvl>`)
				}
				strXml.WriteString(`    </c:multiLvlStrCache>`)
				strXml.WriteString(`  </c:multiLvlStrRef>`)
			}
			strXml.WriteString(`</c:cat>`)

			// 3: "Values"
			valCol := getExcelColName(obj.DataIndex + len(obj.Labels) + 1)
			strXml.WriteString(`<c:val>`)
			strXml.WriteString(`  <c:numRef>`)
			strXml.WriteString(`<c:f>Sheet1!$` + valCol + `$2:$` + valCol + `$` + itoa(len(obj.Labels[0])+1) + `</c:f>`)
			strXml.WriteString(`    <c:numCache>`)
			strXml.WriteString(`      <c:formatCode>` + strOr(strOr(opts.ValLabelFormatCode, opts.DataTableFormatCode), "General") + `</c:formatCode>`)
			strXml.WriteString(`      <c:ptCount val="` + itoa(len(obj.Labels[0])) + `"/>`)
			// Minor (gen-charts.ts:537): TS guards each point with
			// `if (value || value === 0)` to skip JS array holes (undefined). Go's
			// Values is []float64, which cannot hold a hole (every index is a real
			// number), so the guard is inert and unrepresentable — no change needed.
			// (Do NOT switch Values to []*float64 to model JS holes; out of scope.)
			for idx, value := range obj.Values {
				strXml.WriteString(`<c:pt idx="` + itoa(idx) + `"><c:v>` + ftoa(value) + `</c:v></c:pt>`)
			}
			strXml.WriteString(`    </c:numCache>`)
			strXml.WriteString(`  </c:numRef>`)
			strXml.WriteString(`</c:val>`)

			if chartType == ChartTypeLine {
				strXml.WriteString(`<c:smooth val="` + chartB2S(opts.LineSmooth) + `"/>`)
			}

			strXml.WriteString(`</c:ser>`)
		}

		// 3: "Data Labels"
		strXml.WriteString(`  <c:dLbls>`)
		strXml.WriteString(`    <c:numFmt formatCode="` + numFmtOrGeneral(opts.DataLabelFormatCode) + `" sourceLinked="0"/>`)
		strXml.WriteString(`    <c:txPr>`)
		strXml.WriteString(`      <a:bodyPr/>`)
		strXml.WriteString(`      <a:lstStyle/>`)
		strXml.WriteString(`      <a:p><a:pPr>`)
		strXml.WriteString(`        <a:defRPr b="` + chartB2S(opts.DataLabelFontBold) + `" i="` + chartB2S(opts.DataLabelFontItalic) + `" strike="noStrike" sz="` + itoa(int(jsRound(fontSizeOr(opts.DataLabelFontSize, DEF_FONT_SIZE)*100))) + `" u="none">`)
		strXml.WriteString(`          <a:solidFill>` + createColorElement(strOr(opts.DataLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
		strXml.WriteString(`          <a:latin typeface="` + strOr(opts.DataLabelFontFace, "Arial") + `"/>`)
		strXml.WriteString(`        </a:defRPr>`)
		strXml.WriteString(`      </a:pPr></a:p>`)
		strXml.WriteString(`    </c:txPr>`)
		if opts.DataLabelPosition != "" {
			strXml.WriteString(` <c:dLblPos val="` + opts.DataLabelPosition + `"/>`)
		}
		strXml.WriteString(`    <c:showLegendKey val="0"/>`)
		strXml.WriteString(`    <c:showVal val="` + chartB2S(opts.ShowValue) + `"/>`)
		strXml.WriteString(`    <c:showCatName val="0"/>`)
		strXml.WriteString(`    <c:showSerName val="` + chartB2S(opts.ShowSerName) + `"/>`)
		strXml.WriteString(`    <c:showPercent val="0"/>`)
		strXml.WriteString(`    <c:showBubbleSize val="0"/>`)
		strXml.WriteString(`    <c:showLeaderLines val="` + chartB2S(opts.ShowLeaderLines) + `"/>`)
		strXml.WriteString(`  </c:dLbls>`)

		// 4: chart options (gapWidth, marker, etc.)
		if chartType == ChartTypeBar {
			strXml.WriteString(`  <c:gapWidth val="` + ftoa(fptrOr(opts.BarGapWidthPct, 150)) + `"/>`)
			overlap := "0"
			if strings.Contains(opts.BarGrouping, "tacked") {
				overlap = "100"
			} else if opts.BarOverlapPct != 0 {
				overlap = ftoa(opts.BarOverlapPct)
			}
			strXml.WriteString(`  <c:overlap val="` + overlap + `"/>`)
		} else if chartType == ChartTypeBar3d {
			strXml.WriteString(`  <c:gapWidth val="` + ftoa(fptrOr(opts.BarGapWidthPct, 150)) + `"/>`)
			strXml.WriteString(`  <c:gapDepth val="` + ftoa(fptrOr(opts.BarGapDepthPct, 150)) + `"/>`)
			strXml.WriteString(`  <c:shape val="` + opts.Bar3DShape + `"/>`)
		} else if chartType == ChartTypeLine {
			strXml.WriteString(`  <c:marker val="1"/>`)
		}

		// 5: axisId (category first)
		strXml.WriteString(`<c:axId val="` + catAxisID + `"/><c:axId val="` + valAxisID + `"/><c:axId val="` + AXIS_ID_SERIES_PRIMARY + `"/>`)

		// 6: Close Chart tag
		strXml.WriteString(`</c:` + string(chartType) + `Chart>`)

	case ChartTypeScatter:
		strXml.WriteString(`<c:` + string(chartType) + `Chart>`)
		strXml.WriteString(`<c:scatterStyle val="lineMarker"/>`)
		strXml.WriteString(`<c:varyColors val="0"/>`)

		colorIndex = -1
		for idx := 1; idx < len(data); idx++ {
			obj := &data[idx]
			serIdx := idx - 1
			colorIndex++
			strXml.WriteString(`<c:ser>`)
			strXml.WriteString(`  <c:idx val="` + itoa(serIdx) + `"/>`)
			strXml.WriteString(`  <c:order val="` + itoa(serIdx) + `"/>`)
			strXml.WriteString(`  <c:tx>`)
			strXml.WriteString(`    <c:strRef>`)
			strXml.WriteString(`      <c:f>Sheet1!$` + getExcelColName(serIdx+2) + `$1</c:f>`)
			strXml.WriteString(`      <c:strCache><c:ptCount val="1"/><c:pt idx="0"><c:v>` + encodeXmlEntities(obj.Name) + `</c:v></c:pt></c:strCache>`)
			strXml.WriteString(`    </c:strRef>`)
			strXml.WriteString(`  </c:tx>`)

			strXml.WriteString(`  <c:spPr>`)
			tmpSerColor := opts.ChartColors[colorIndex%len(opts.ChartColors)]
			if tmpSerColor == "transparent" {
				strXml.WriteString(`<a:noFill/>`)
			} else if opts.ChartColorsOpacity != 0 {
				strXml.WriteString(`<a:solidFill>` + createColorElement(tmpSerColor, `<a:alpha val="`+itoa(int(jsRound(opts.ChartColorsOpacity*1000)))+`"/>`) + `</a:solidFill>`)
			} else {
				strXml.WriteString(`<a:solidFill>` + createColorElement(tmpSerColor, "") + `</a:solidFill>`)
			}
			if fptrOr(opts.LineSize, 2) == 0 {
				strXml.WriteString(`<a:ln><a:noFill/></a:ln>`)
			} else {
				lineDash := strOr(opts.LineDash, "solid")
				strXml.WriteString(`<a:ln w="` + itoa(valToPts(fptrOr(opts.LineSize, 2))) + `" cap="` + createLineCap(opts.LineCap) + `"><a:solidFill>` + createColorElement(tmpSerColor, "") + `</a:solidFill>`)
				strXml.WriteString(`<a:prstDash val="` + lineDash + `"/><a:round/></a:ln>`)
			}
			strXml.WriteString(createShadowElement(opts.Shadow, DEF_SHAPE_SHADOW))
			strXml.WriteString(`  </c:spPr>`)

			// marker
			strXml.WriteString(`<c:marker>`)
			strXml.WriteString(`  <c:symbol val="` + opts.LineDataSymbol + `"/>`)
			if opts.LineDataSymbolSize != 0 {
				strXml.WriteString(`<c:size val="` + ftoa(opts.LineDataSymbolSize) + `"/>`)
			}
			strXml.WriteString(`<c:spPr>`)
			strXml.WriteString(`<a:solidFill>` + createColorElement(opts.ChartColors[markerColorIdx(serIdx, len(opts.ChartColors))], "") + `</a:solidFill>`)
			strXml.WriteString(`<a:ln w="` + ftoa(opts.LineDataSymbolLineSize) + `" cap="flat"><a:solidFill>` + createColorElement(strOr(opts.LineDataSymbolLineColor, opts.ChartColors[colorIndex%len(opts.ChartColors)]), "") + `</a:solidFill><a:prstDash val="solid"/><a:round/></a:ln>`)
			strXml.WriteString(`<a:effectLst/>`)
			strXml.WriteString(`</c:spPr>`)
			strXml.WriteString(`</c:marker>`)

			// scatter data point labels
			if chartBool(opts.ShowLabel) {
				chartUUID := getUuid("-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
				if len(obj.Labels) > 0 && (opts.DataLabelFormatScatter == "custom" || opts.DataLabelFormatScatter == "customXY") {
					strXml.WriteString(`<c:dLbls>`)
					for lidx, label := range obj.Labels[0] {
						strXml.WriteString(`  <c:dLbl>`)
						strXml.WriteString(`    <c:idx val="` + itoa(lidx) + `"/>`)
						strXml.WriteString(`    <c:tx>`)
						strXml.WriteString(`      <c:rich>`)
						strXml.WriteString(`            <a:bodyPr>`)
						strXml.WriteString(`                <a:spAutoFit/>`)
						strXml.WriteString(`            </a:bodyPr>`)
						strXml.WriteString(`            <a:lstStyle/>`)
						strXml.WriteString(`            <a:p>`)
						strXml.WriteString(`                <a:pPr>`)
						strXml.WriteString(`                    <a:defRPr/>`)
						strXml.WriteString(`                </a:pPr>`)
						strXml.WriteString(`              <a:r>`)
						strXml.WriteString(`                    <a:rPr lang="` + strOr(opts.Lang, "en-US") + `" dirty="0"/>`)
						strXml.WriteString(`                    <a:t>` + encodeXmlEntities(label) + `</a:t>`)
						strXml.WriteString(`              </a:r>`)
						if opts.DataLabelFormatScatter == "customXY" && !isBlankLabel(label) {
							strXml.WriteString(`              <a:r>`)
							strXml.WriteString(`                  <a:rPr lang="` + strOr(opts.Lang, "en-US") + `" baseline="0" dirty="0"/>`)
							strXml.WriteString(`                  <a:t> (</a:t>`)
							strXml.WriteString(`              </a:r>`)
							strXml.WriteString(`              <a:fld id="{` + getUuid("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx") + `}" type="XVALUE">`)
							strXml.WriteString(`                  <a:rPr lang="` + strOr(opts.Lang, "en-US") + `" baseline="0"/>`)
							strXml.WriteString(`                  <a:pPr>`)
							strXml.WriteString(`                      <a:defRPr/>`)
							strXml.WriteString(`                  </a:pPr>`)
							strXml.WriteString(`                  <a:t>[` + encodeXmlEntities(obj.Name) + `</a:t>`)
							strXml.WriteString(`              </a:fld>`)
							strXml.WriteString(`              <a:r>`)
							strXml.WriteString(`                  <a:rPr lang="` + strOr(opts.Lang, "en-US") + `" baseline="0" dirty="0"/>`)
							strXml.WriteString(`                  <a:t>, </a:t>`)
							strXml.WriteString(`              </a:r>`)
							strXml.WriteString(`              <a:fld id="{` + getUuid("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx") + `}" type="YVALUE">`)
							strXml.WriteString(`                  <a:rPr lang="` + strOr(opts.Lang, "en-US") + `" baseline="0"/>`)
							strXml.WriteString(`                  <a:pPr>`)
							strXml.WriteString(`                      <a:defRPr/>`)
							strXml.WriteString(`                  </a:pPr>`)
							strXml.WriteString(`                  <a:t>[` + encodeXmlEntities(obj.Name) + `]</a:t>`)
							strXml.WriteString(`              </a:fld>`)
							strXml.WriteString(`              <a:r>`)
							strXml.WriteString(`                  <a:rPr lang="` + strOr(opts.Lang, "en-US") + `" baseline="0" dirty="0"/>`)
							strXml.WriteString(`                  <a:t>)</a:t>`)
							strXml.WriteString(`              </a:r>`)
							strXml.WriteString(`              <a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `" dirty="0"/>`)
						}
						strXml.WriteString(`            </a:p>`)
						strXml.WriteString(`      </c:rich>`)
						strXml.WriteString(`    </c:tx>`)
						strXml.WriteString(`    <c:spPr>`)
						strXml.WriteString(`        <a:noFill/>`)
						strXml.WriteString(`        <a:ln>`)
						strXml.WriteString(`            <a:noFill/>`)
						strXml.WriteString(`        </a:ln>`)
						strXml.WriteString(`        <a:effectLst/>`)
						strXml.WriteString(`    </c:spPr>`)
						if opts.DataLabelPosition != "" {
							strXml.WriteString(` <c:dLblPos val="` + opts.DataLabelPosition + `"/>`)
						}
						strXml.WriteString(`    <c:showLegendKey val="0"/>`)
						strXml.WriteString(`    <c:showVal val="0"/>`)
						strXml.WriteString(`    <c:showCatName val="0"/>`)
						strXml.WriteString(`    <c:showSerName val="0"/>`)
						strXml.WriteString(`    <c:showPercent val="0"/>`)
						strXml.WriteString(`    <c:showBubbleSize val="0"/>`)
						strXml.WriteString(`       <c:showLeaderLines val="1"/>`)
						strXml.WriteString(`    <c:extLst>`)
						strXml.WriteString(`      <c:ext uri="{CE6537A1-D6FC-4f65-9D91-7224C49458BB}" xmlns:c15="http://schemas.microsoft.com/office/drawing/2012/chart"/>`)
						strXml.WriteString(`      <c:ext uri="{C3380CC4-5D6E-409C-BE32-E72D297353CC}" xmlns:c16="http://schemas.microsoft.com/office/drawing/2014/chart">`)
						strXml.WriteString(`            <c16:uniqueId val="{` + scatterUniqueID(lidx+1) + chartUUID + `}"/>`)
						strXml.WriteString(`      </c:ext>`)
						strXml.WriteString(`        </c:extLst>`)
						strXml.WriteString(`</c:dLbl>`)
					}
					strXml.WriteString(`</c:dLbls>`)
				}
				if opts.DataLabelFormatScatter == "XY" {
					strXml.WriteString(`<c:dLbls>`)
					strXml.WriteString(`    <c:spPr>`)
					strXml.WriteString(`        <a:noFill/>`)
					strXml.WriteString(`        <a:ln>`)
					strXml.WriteString(`            <a:noFill/>`)
					strXml.WriteString(`        </a:ln>`)
					strXml.WriteString(`          <a:effectLst/>`)
					strXml.WriteString(`    </c:spPr>`)
					strXml.WriteString(`    <c:txPr>`)
					strXml.WriteString(`        <a:bodyPr>`)
					strXml.WriteString(`            <a:spAutoFit/>`)
					strXml.WriteString(`        </a:bodyPr>`)
					strXml.WriteString(`        <a:lstStyle/>`)
					strXml.WriteString(`        <a:p>`)
					strXml.WriteString(`            <a:pPr>`)
					strXml.WriteString(`                <a:defRPr/>`)
					strXml.WriteString(`            </a:pPr>`)
					strXml.WriteString(`            <a:endParaRPr lang="en-US"/>`)
					strXml.WriteString(`        </a:p>`)
					strXml.WriteString(`    </c:txPr>`)
					if opts.DataLabelPosition != "" {
						strXml.WriteString(` <c:dLblPos val="` + opts.DataLabelPosition + `"/>`)
					}
					strXml.WriteString(`    <c:showLegendKey val="0"/>`)
					strXml.WriteString(` <c:showVal val="` + chartB2S(opts.ShowLabel) + `"/>`)
					strXml.WriteString(` <c:showCatName val="` + chartB2S(opts.ShowLabel) + `"/>`)
					strXml.WriteString(` <c:showSerName val="` + chartB2S(opts.ShowSerName) + `"/>`)
					strXml.WriteString(`    <c:showPercent val="0"/>`)
					strXml.WriteString(`    <c:showBubbleSize val="0"/>`)
					strXml.WriteString(`    <c:extLst>`)
					strXml.WriteString(`        <c:ext uri="{CE6537A1-D6FC-4f65-9D91-7224C49458BB}" xmlns:c15="http://schemas.microsoft.com/office/drawing/2012/chart">`)
					strXml.WriteString(`            <c15:showLeaderLines val="1"/>`)
					strXml.WriteString(`        </c:ext>`)
					strXml.WriteString(`    </c:extLst>`)
					strXml.WriteString(`</c:dLbls>`)
				}
			}

			// Data point colors (single-series)
			if len(data) == 1 && !chartColorsEqual(opts.ChartColors, BARCHART_COLORS) {
				for index, value := range obj.Values {
					arrColors := opts.ChartColors
					if value < 0 {
						if len(opts.InvertedColors) > 0 {
							arrColors = opts.InvertedColors
						} else if opts.ChartColors != nil {
							arrColors = opts.ChartColors
						} else {
							arrColors = BARCHART_COLORS
						}
					} else if opts.ChartColors == nil {
						arrColors = []string{}
					}
					strXml.WriteString(`  <c:dPt>`)
					strXml.WriteString(`    <c:idx val="` + itoa(index) + `"/>`)
					strXml.WriteString(`      <c:invertIfNegative val="0"/>`)
					strXml.WriteString(`    <c:bubble3D val="0"/>`)
					strXml.WriteString(`    <c:spPr>`)
					if fptrOr(opts.LineSize, 2) == 0 {
						strXml.WriteString(`<a:ln><a:noFill/></a:ln>`)
					} else {
						strXml.WriteString(`<a:solidFill>`)
						strXml.WriteString(` <a:srgbClr val="` + arrColors[index%len(arrColors)] + `"/>`)
						strXml.WriteString(`</a:solidFill>`)
					}
					strXml.WriteString(createShadowElement(opts.Shadow, DEF_SHAPE_SHADOW))
					strXml.WriteString(`    </c:spPr>`)
					strXml.WriteString(`  </c:dPt>`)
				}
			}

			// 3: xVal / yVal
			strXml.WriteString(`<c:xVal>`)
			strXml.WriteString(`  <c:numRef>`)
			strXml.WriteString(`    <c:f>Sheet1!$A$2:$A$` + itoa(len(data[0].Values)+1) + `</c:f>`)
			strXml.WriteString(`    <c:numCache>`)
			strXml.WriteString(`      <c:formatCode>General</c:formatCode>`)
			strXml.WriteString(`      <c:ptCount val="` + itoa(len(data[0].Values)) + `"/>`)
			for vidx, value := range data[0].Values {
				strXml.WriteString(`<c:pt idx="` + itoa(vidx) + `"><c:v>` + ftoa(value) + `</c:v></c:pt>`)
			}
			strXml.WriteString(`    </c:numCache>`)
			strXml.WriteString(`  </c:numRef>`)
			strXml.WriteString(`</c:xVal>`)

			yCol := getExcelColName(serIdx + 2)
			strXml.WriteString(`<c:yVal>`)
			strXml.WriteString(`  <c:numRef>`)
			strXml.WriteString(`    <c:f>Sheet1!$` + yCol + `$2:$` + yCol + `$` + itoa(len(data[0].Values)+1) + `</c:f>`)
			strXml.WriteString(`    <c:numCache>`)
			strXml.WriteString(`      <c:formatCode>General</c:formatCode>`)
			strXml.WriteString(`      <c:ptCount val="` + itoa(len(data[0].Values)) + `"/>`)
			for vidx := range data[0].Values {
				strXml.WriteString(`<c:pt idx="` + itoa(vidx) + `"><c:v>` + numAt(obj.Values, vidx) + `</c:v></c:pt>`)
			}
			strXml.WriteString(`    </c:numCache>`)
			strXml.WriteString(`  </c:numRef>`)
			strXml.WriteString(`</c:yVal>`)

			strXml.WriteString(`<c:smooth val="` + chartB2S(opts.LineSmooth) + `"/>`)
			strXml.WriteString(`</c:ser>`)
		}

		// 3: Data Labels
		strXml.WriteString(`  <c:dLbls>`)
		strXml.WriteString(`    <c:numFmt formatCode="` + numFmtOrGeneral(opts.DataLabelFormatCode) + `" sourceLinked="0"/>`)
		strXml.WriteString(`    <c:txPr>`)
		strXml.WriteString(`      <a:bodyPr/>`)
		strXml.WriteString(`      <a:lstStyle/>`)
		strXml.WriteString(`      <a:p><a:pPr>`)
		strXml.WriteString(`        <a:defRPr b="` + chartB2S(opts.DataLabelFontBold) + `" i="` + chartB2S(opts.DataLabelFontItalic) + `" strike="noStrike" sz="` + itoa(int(jsRound(fontSizeOr(opts.DataLabelFontSize, DEF_FONT_SIZE)*100))) + `" u="none">`)
		strXml.WriteString(`          <a:solidFill>` + createColorElement(strOr(opts.DataLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
		strXml.WriteString(`          <a:latin typeface="` + strOr(opts.DataLabelFontFace, "Arial") + `"/>`)
		strXml.WriteString(`        </a:defRPr>`)
		strXml.WriteString(`      </a:pPr></a:p>`)
		strXml.WriteString(`    </c:txPr>`)
		if opts.DataLabelPosition != "" {
			strXml.WriteString(` <c:dLblPos val="` + opts.DataLabelPosition + `"/>`)
		}
		strXml.WriteString(`    <c:showLegendKey val="0"/>`)
		strXml.WriteString(`    <c:showVal val="` + chartB2S(opts.ShowValue) + `"/>`)
		strXml.WriteString(`    <c:showCatName val="0"/>`)
		strXml.WriteString(`    <c:showSerName val="` + chartB2S(opts.ShowSerName) + `"/>`)
		strXml.WriteString(`    <c:showPercent val="0"/>`)
		strXml.WriteString(`    <c:showBubbleSize val="0"/>`)
		strXml.WriteString(`  </c:dLbls>`)

		strXml.WriteString(`<c:axId val="` + catAxisID + `"/><c:axId val="` + valAxisID + `"/>`)
		strXml.WriteString(`</c:` + string(chartType) + `Chart>`)

	case ChartTypeBubble, ChartTypeBubble3d:
		strXml.WriteString(`<c:bubbleChart>`)
		strXml.WriteString(`<c:varyColors val="0"/>`)

		colorIndex = -1
		for idx := 1; idx < len(data); idx++ {
			obj := &data[idx]
			serIdx := idx - 1
			colorIndex++
			strXml.WriteString(`<c:ser>`)
			strXml.WriteString(`  <c:idx val="` + itoa(serIdx) + `"/>`)
			strXml.WriteString(`  <c:order val="` + itoa(serIdx) + `"/>`)

			strXml.WriteString(`  <c:tx>`)
			strXml.WriteString(`    <c:strRef>`)
			strXml.WriteString(`      <c:f>Sheet1!$` + getExcelColName(idxColLtr+1) + `$1</c:f>`)
			strXml.WriteString(`      <c:strCache><c:ptCount val="1"/><c:pt idx="0"><c:v>` + encodeXmlEntities(obj.Name) + `</c:v></c:pt></c:strCache>`)
			strXml.WriteString(`    </c:strRef>`)
			strXml.WriteString(`  </c:tx>`)

			strXml.WriteString(`<c:spPr>`)
			tmpSerColor := opts.ChartColors[colorIndex%len(opts.ChartColors)]
			if tmpSerColor == "transparent" {
				strXml.WriteString(`<a:noFill/>`)
			} else if opts.ChartColorsOpacity != 0 {
				strXml.WriteString(`<a:solidFill>` + createColorElement(tmpSerColor, `<a:alpha val="`+itoa(int(jsRound(opts.ChartColorsOpacity*1000)))+`"/>`) + `</a:solidFill>`)
			} else {
				strXml.WriteString(`<a:solidFill>` + createColorElement(tmpSerColor, "") + `</a:solidFill>`)
			}
			if fptrOr(opts.LineSize, 2) == 0 {
				strXml.WriteString(`<a:ln><a:noFill/></a:ln>`)
			} else if opts.DataBorder != nil {
				strXml.WriteString(`<a:ln w="` + itoa(valToPts(opts.DataBorder.Pt)) + `" cap="flat"><a:solidFill>` + createColorElement(opts.DataBorder.Color, "") + `</a:solidFill><a:prstDash val="solid"/><a:round/></a:ln>`)
			} else {
				lineDash := strOr(opts.LineDash, "solid")
				strXml.WriteString(`<a:ln w="` + itoa(valToPts(fptrOr(opts.LineSize, 2))) + `" cap="flat"><a:solidFill>` + createColorElement(tmpSerColor, "") + `</a:solidFill>`)
				strXml.WriteString(`<a:prstDash val="` + lineDash + `"/><a:round/></a:ln>`)
			}
			strXml.WriteString(createShadowElement(opts.Shadow, DEF_SHAPE_SHADOW))
			strXml.WriteString(`</c:spPr>`)

			// xVal / yVal
			strXml.WriteString(`<c:xVal>`)
			strXml.WriteString(`  <c:numRef>`)
			strXml.WriteString(`    <c:f>Sheet1!$A$2:$A$` + itoa(len(data[0].Values)+1) + `</c:f>`)
			strXml.WriteString(`    <c:numCache>`)
			strXml.WriteString(`      <c:formatCode>General</c:formatCode>`)
			strXml.WriteString(`      <c:ptCount val="` + itoa(len(data[0].Values)) + `"/>`)
			for vidx, value := range data[0].Values {
				strXml.WriteString(`<c:pt idx="` + itoa(vidx) + `"><c:v>` + ftoa(value) + `</c:v></c:pt>`)
			}
			strXml.WriteString(`    </c:numCache>`)
			strXml.WriteString(`  </c:numRef>`)
			strXml.WriteString(`</c:xVal>`)

			yCol := getExcelColName(idxColLtr + 1)
			strXml.WriteString(`<c:yVal>`)
			strXml.WriteString(`  <c:numRef>`)
			strXml.WriteString(`<c:f>Sheet1!$` + yCol + `$2:$` + yCol + `$` + itoa(len(data[0].Values)+1) + `</c:f>`)
			idxColLtr++
			strXml.WriteString(`    <c:numCache>`)
			strXml.WriteString(`      <c:formatCode>General</c:formatCode>`)
			strXml.WriteString(`      <c:ptCount val="` + itoa(len(data[0].Values)) + `"/>`)
			for vidx := range data[0].Values {
				strXml.WriteString(`<c:pt idx="` + itoa(vidx) + `"><c:v>` + numAt(obj.Values, vidx) + `</c:v></c:pt>`)
			}
			strXml.WriteString(`    </c:numCache>`)
			strXml.WriteString(`  </c:numRef>`)
			strXml.WriteString(`</c:yVal>`)

			// bubbleSize
			sizeCol := getExcelColName(idxColLtr + 1)
			strXml.WriteString(`  <c:bubbleSize>`)
			strXml.WriteString(`    <c:numRef>`)
			strXml.WriteString(`<c:f>Sheet1!$` + sizeCol + `$2:$` + sizeCol + `$` + itoa(len(obj.Sizes)+1) + `</c:f>`)
			idxColLtr++
			strXml.WriteString(`      <c:numCache>`)
			strXml.WriteString(`        <c:formatCode>General</c:formatCode>`)
			strXml.WriteString(`           <c:ptCount val="` + itoa(len(obj.Sizes)) + `"/>`)
			for sidx, value := range obj.Sizes {
				strXml.WriteString(`<c:pt idx="` + itoa(sidx) + `"><c:v>` + numTruthy(value) + `</c:v></c:pt>`)
			}
			strXml.WriteString(`      </c:numCache>`)
			strXml.WriteString(`    </c:numRef>`)
			strXml.WriteString(`  </c:bubbleSize>`)
			bubble3D := "0"
			if chartType == ChartTypeBubble3d {
				bubble3D = "1"
			}
			strXml.WriteString(`  <c:bubble3D val="` + bubble3D + `"/>`)

			strXml.WriteString(`</c:ser>`)
		}

		// Data Labels
		strXml.WriteString(`<c:dLbls>`)
		strXml.WriteString(`<c:numFmt formatCode="` + numFmtOrGeneral(opts.DataLabelFormatCode) + `" sourceLinked="0"/>`)
		strXml.WriteString(`<c:txPr><a:bodyPr/><a:lstStyle/><a:p><a:pPr>`)
		strXml.WriteString(`<a:defRPr b="` + chartB2S(opts.DataLabelFontBold) + `" i="` + chartB2S(opts.DataLabelFontItalic) + `" strike="noStrike" sz="` + itoa(int(jsRound(jsRound(fontSizeOr(opts.DataLabelFontSize, DEF_FONT_SIZE))*100))) + `" u="none">`)
		strXml.WriteString(`<a:solidFill>` + createColorElement(strOr(opts.DataLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
		strXml.WriteString(`<a:latin typeface="` + strOr(opts.DataLabelFontFace, "Arial") + `"/>`)
		strXml.WriteString(`</a:defRPr></a:pPr></a:p></c:txPr>`)
		if opts.DataLabelPosition != "" {
			strXml.WriteString(`<c:dLblPos val="` + opts.DataLabelPosition + `"/>`)
		}
		strXml.WriteString(`<c:showLegendKey val="0"/>`)
		strXml.WriteString(`<c:showVal val="` + chartB2S(opts.ShowValue) + `"/>`)
		strXml.WriteString(`<c:showCatName val="0"/><c:showSerName val="` + chartB2S(opts.ShowSerName) + `"/><c:showPercent val="0"/><c:showBubbleSize val="0"/>`)
		strXml.WriteString(`<c:extLst>`)
		strXml.WriteString(`  <c:ext uri="{CE6537A1-D6FC-4f65-9D91-7224C49458BB}" xmlns:c15="http://schemas.microsoft.com/office/drawing/2012/chart">`)
		strXml.WriteString(`    <c15:showLeaderLines val="` + chartB2S(opts.ShowLeaderLines) + `"/>`)
		strXml.WriteString(`  </c:ext>`)
		strXml.WriteString(`</c:extLst>`)
		strXml.WriteString(`</c:dLbls>`)

		strXml.WriteString(`<c:axId val="` + catAxisID + `"/><c:axId val="` + valAxisID + `"/>`)
		strXml.WriteString(`</c:bubbleChart>`)

	case ChartTypeDoughnut, ChartTypePie:
		optsChartData := &data[0]

		strXml.WriteString(`<c:` + string(chartType) + `Chart>`)
		strXml.WriteString(`  <c:varyColors val="1"/>`)
		strXml.WriteString(`<c:ser>`)
		strXml.WriteString(`  <c:idx val="0"/>`)
		strXml.WriteString(`  <c:order val="0"/>`)
		strXml.WriteString(`  <c:tx>`)
		strXml.WriteString(`    <c:strRef>`)
		strXml.WriteString(`      <c:f>Sheet1!$B$1</c:f>`)
		strXml.WriteString(`      <c:strCache>`)
		strXml.WriteString(`        <c:ptCount val="1"/>`)
		strXml.WriteString(`        <c:pt idx="0"><c:v>` + encodeXmlEntities(optsChartData.Name) + `</c:v></c:pt>`)
		strXml.WriteString(`      </c:strCache>`)
		strXml.WriteString(`    </c:strRef>`)
		strXml.WriteString(`  </c:tx>`)
		strXml.WriteString(`  <c:spPr>`)
		strXml.WriteString(`    <a:solidFill><a:schemeClr val="accent1"/></a:solidFill>`)
		strXml.WriteString(`    <a:ln w="9525" cap="flat"><a:solidFill><a:srgbClr val="F9F9F9"/></a:solidFill><a:prstDash val="solid"/><a:round/></a:ln>`)
		if chartBool(opts.DataNoEffects) {
			strXml.WriteString(`<a:effectLst/>`)
		} else {
			strXml.WriteString(createShadowElement(opts.Shadow, DEF_SHAPE_SHADOW))
		}
		strXml.WriteString(`  </c:spPr>`)

		// 2: Data Point block
		for idx := range optsChartData.Labels[0] {
			strXml.WriteString(`<c:dPt>`)
			strXml.WriteString(` <c:idx val="` + itoa(idx) + `"/>`)
			strXml.WriteString(` <c:bubble3D val="0"/>`)
			strXml.WriteString(` <c:spPr>`)
			strXml.WriteString(`<a:solidFill>` + createColorElement(opts.ChartColors[pieColorIdx(idx, len(opts.ChartColors))], "") + `</a:solidFill>`)
			if opts.DataBorder != nil {
				strXml.WriteString(`<a:ln w="` + itoa(valToPts(opts.DataBorder.Pt)) + `" cap="flat"><a:solidFill>` + createColorElement(opts.DataBorder.Color, "") + `</a:solidFill><a:prstDash val="solid"/><a:round/></a:ln>`)
			}
			strXml.WriteString(createShadowElement(opts.Shadow, DEF_SHAPE_SHADOW))
			strXml.WriteString(`  </c:spPr>`)
			strXml.WriteString(`</c:dPt>`)
		}

		// 3: Data Label block
		strXml.WriteString(`<c:dLbls>`)
		for idx := range optsChartData.Labels[0] {
			strXml.WriteString(`<c:dLbl>`)
			strXml.WriteString(` <c:idx val="` + itoa(idx) + `"/>`)
			strXml.WriteString(`  <c:numFmt formatCode="` + numFmtOrGeneral(opts.DataLabelFormatCode) + `" sourceLinked="0"/>`)
			strXml.WriteString(`  <c:spPr/><c:txPr>`)
			strXml.WriteString(`   <a:bodyPr/><a:lstStyle/>`)
			strXml.WriteString(`   <a:p><a:pPr>`)
			strXml.WriteString(`   <a:defRPr sz="` + itoa(int(jsRound(fontSizeOr(opts.DataLabelFontSize, DEF_FONT_SIZE)*100))) + `" b="` + chartB2S(opts.DataLabelFontBold) + `" i="` + chartB2S(opts.DataLabelFontItalic) + `" u="none" strike="noStrike">`)
			strXml.WriteString(`    <a:solidFill>` + createColorElement(strOr(opts.DataLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
			strXml.WriteString(`    <a:latin typeface="` + strOr(opts.DataLabelFontFace, "Arial") + `"/>`)
			strXml.WriteString(`   </a:defRPr>`)
			strXml.WriteString(`      </a:pPr></a:p>`)
			strXml.WriteString(`    </c:txPr>`)
			if chartType == ChartTypePie && opts.DataLabelPosition != "" {
				strXml.WriteString(`<c:dLblPos val="` + opts.DataLabelPosition + `"/>`)
			}
			strXml.WriteString(`    <c:showLegendKey val="0"/>`)
			strXml.WriteString(`    <c:showVal val="` + chartB2S(opts.ShowValue) + `"/>`)
			strXml.WriteString(`    <c:showCatName val="` + chartB2S(opts.ShowLabel) + `"/>`)
			strXml.WriteString(`    <c:showSerName val="` + chartB2S(opts.ShowSerName) + `"/>`)
			strXml.WriteString(`    <c:showPercent val="` + chartB2S(opts.ShowPercent) + `"/>`)
			strXml.WriteString(`    <c:showBubbleSize val="0"/>`)
			strXml.WriteString(`  </c:dLbl>`)
		}
		strXml.WriteString(` <c:numFmt formatCode="` + numFmtOrGeneral(opts.DataLabelFormatCode) + `" sourceLinked="0"/>`)
		strXml.WriteString(`    <c:txPr>`)
		strXml.WriteString(`      <a:bodyPr/>`)
		strXml.WriteString(`      <a:lstStyle/>`)
		strXml.WriteString(`      <a:p>`)
		strXml.WriteString(`        <a:pPr>`)
		strXml.WriteString(`          <a:defRPr sz="1800" b="` + chartB2S(opts.DataLabelFontBold) + `" i="` + chartB2S(opts.DataLabelFontItalic) + `" u="none" strike="noStrike">`)
		strXml.WriteString(`            <a:solidFill><a:srgbClr val="000000"/></a:solidFill><a:latin typeface="Arial"/>`)
		strXml.WriteString(`          </a:defRPr>`)
		strXml.WriteString(`        </a:pPr>`)
		strXml.WriteString(`      </a:p>`)
		strXml.WriteString(`    </c:txPr>`)
		if chartType == ChartTypePie {
			strXml.WriteString(`<c:dLblPos val="ctr"/>`)
		}
		strXml.WriteString(`    <c:showLegendKey val="0"/>`)
		strXml.WriteString(`    <c:showVal val="0"/>`)
		strXml.WriteString(`    <c:showCatName val="1"/>`)
		strXml.WriteString(`    <c:showSerName val="0"/>`)
		strXml.WriteString(`    <c:showPercent val="1"/>`)
		strXml.WriteString(`    <c:showBubbleSize val="0"/>`)
		strXml.WriteString(` <c:showLeaderLines val="` + chartB2S(opts.ShowLeaderLines) + `"/>`)
		strXml.WriteString(`</c:dLbls>`)

		// 2: Categories
		strXml.WriteString(`<c:cat>`)
		strXml.WriteString(`  <c:strRef>`)
		strXml.WriteString(`    <c:f>Sheet1!$A$2:$A$` + itoa(len(optsChartData.Labels[0])+1) + `</c:f>`)
		strXml.WriteString(`    <c:strCache>`)
		strXml.WriteString(`         <c:ptCount val="` + itoa(len(optsChartData.Labels[0])) + `"/>`)
		for idx, label := range optsChartData.Labels[0] {
			strXml.WriteString(`<c:pt idx="` + itoa(idx) + `"><c:v>` + encodeXmlEntities(label) + `</c:v></c:pt>`)
		}
		strXml.WriteString(`    </c:strCache>`)
		strXml.WriteString(`  </c:strRef>`)
		strXml.WriteString(`</c:cat>`)

		// 3: vals
		strXml.WriteString(`  <c:val>`)
		strXml.WriteString(`    <c:numRef>`)
		strXml.WriteString(`      <c:f>Sheet1!$B$2:$B$` + itoa(len(optsChartData.Labels[0])+1) + `</c:f>`)
		strXml.WriteString(`      <c:numCache>`)
		strXml.WriteString(`           <c:ptCount val="` + itoa(len(optsChartData.Labels[0])) + `"/>`)
		for idx, value := range optsChartData.Values {
			strXml.WriteString(`<c:pt idx="` + itoa(idx) + `"><c:v>` + ftoa(value) + `</c:v></c:pt>`)
		}
		strXml.WriteString(`      </c:numCache>`)
		strXml.WriteString(`    </c:numRef>`)
		strXml.WriteString(`  </c:val>`)

		strXml.WriteString(`  </c:ser>`)
		firstSlice := "0"
		if opts.FirstSliceAng != 0 {
			firstSlice = itoa(int(jsRound(opts.FirstSliceAng)))
		}
		strXml.WriteString(`  <c:firstSliceAng val="` + firstSlice + `"/>`)
		if chartType == ChartTypeDoughnut {
			// TS: `typeof opts.holeSize === 'number' ? opts.holeSize : '50'`
			// (gen-charts.ts:1616) — explicit 0 renders <c:holeSize val="0"/>.
			holeSize := "50"
			if opts.HoleSize != nil {
				holeSize = ftoa(*opts.HoleSize)
			}
			strXml.WriteString(`<c:holeSize val="` + holeSize + `"/>`)
		}
		strXml.WriteString(`</c:` + string(chartType) + `Chart>`)

	default:
		strXml.WriteString(``)
	}

	return strXml.String()
}

// ---------------------------------------------------------------------------
// axis / title / shadow / gridline helpers
// ---------------------------------------------------------------------------

// makeCatAxis ports gen-charts.ts makeCatAxis.
func makeCatAxis(opts *ChartOptions, axisID, valAxisID string) string {
	var strXml strings.Builder
	isXY := opts.Type == ChartTypeScatter || opts.Type == ChartTypeBubble || opts.Type == ChartTypeBubble3d

	if isXY {
		strXml.WriteString(`<c:valAx>`)
	} else {
		if opts.CatLabelFormatCode != "" {
			strXml.WriteString(`<c:dateAx>`)
		} else {
			strXml.WriteString(`<c:catAx>`)
		}
	}
	strXml.WriteString(`  <c:axId val="` + axisID + `"/>`)
	strXml.WriteString(`  <c:scaling>`)
	strXml.WriteString(`<c:orientation val="` + strOr(opts.CatAxisOrientation, "minMax") + `"/>`)
	if opts.CatAxisMaxVal != nil {
		strXml.WriteString(`<c:max val="` + ftoa(*opts.CatAxisMaxVal) + `"/>`)
	}
	if opts.CatAxisMinVal != nil {
		strXml.WriteString(`<c:min val="` + ftoa(*opts.CatAxisMinVal) + `"/>`)
	}
	strXml.WriteString(`</c:scaling>`)
	strXml.WriteString(`  <c:delete val="` + chartB2S(opts.CatAxisHidden) + `"/>`)
	strXml.WriteString(`  <c:axPos val="` + axPosBL(opts.BarDir) + `"/>`)
	if opts.CatGridLine != nil && opts.CatGridLine.Style != "none" {
		strXml.WriteString(createGridLineElement(opts.CatGridLine))
	}
	if chartBool(opts.ShowCatAxisTitle) {
		strXml.WriteString(genXmlTitle(chartTitleOpts{
			Color:       opts.CatAxisTitleColor,
			FontFace:    opts.CatAxisTitleFontFace,
			FontSize:    opts.CatAxisTitleFontSize,
			TitleRotate: opts.CatAxisTitleRotate,
			Title:       strOr(opts.CatAxisTitle, "Axis Title"),
		}, 0, 0))
	}
	if isXY {
		strXml.WriteString(`  <c:numFmt formatCode="` + valAxisFmt(opts.ValAxisLabelFormatCode) + `" sourceLinked="1"/>`)
	} else {
		strXml.WriteString(`  <c:numFmt formatCode="` + numFmtOrGeneral(opts.CatLabelFormatCode) + `" sourceLinked="1"/>`)
	}
	if opts.Type == ChartTypeScatter {
		strXml.WriteString(`  <c:majorTickMark val="none"/>`)
		strXml.WriteString(`  <c:minorTickMark val="none"/>`)
		strXml.WriteString(`  <c:tickLblPos val="nextTo"/>`)
	} else {
		strXml.WriteString(`  <c:majorTickMark val="` + strOr(opts.CatAxisMajorTickMark, "out") + `"/>`)
		strXml.WriteString(`  <c:minorTickMark val="` + strOr(opts.CatAxisMinorTickMark, "none") + `"/>`)
		tickLblPos := opts.CatAxisLabelPos
		if tickLblPos == "" {
			if opts.BarDir == "col" {
				tickLblPos = "low"
			} else {
				tickLblPos = "nextTo"
			}
		}
		strXml.WriteString(`  <c:tickLblPos val="` + tickLblPos + `"/>`)
	}
	strXml.WriteString(`  <c:spPr>`)
	catLineW := ONEPT
	if opts.CatAxisLineSize != 0 {
		catLineW = valToPts(opts.CatAxisLineSize)
	}
	strXml.WriteString(`    <a:ln w="` + itoa(catLineW) + `" cap="flat">`)
	if !chartBool(opts.CatAxisLineShow) {
		strXml.WriteString(`<a:noFill/>`)
	} else {
		strXml.WriteString(`<a:solidFill>` + createColorElement(strOr(opts.CatAxisLineColor, DEF_CHART_GRIDLINE.Color), "") + `</a:solidFill>`)
	}
	strXml.WriteString(`      <a:prstDash val="` + strOr(opts.CatAxisLineStyle, "solid") + `"/>`)
	strXml.WriteString(`      <a:round/>`)
	strXml.WriteString(`    </a:ln>`)
	strXml.WriteString(`  </c:spPr>`)
	strXml.WriteString(`  <c:txPr>`)
	if opts.CatAxisLabelRotate != 0 {
		strXml.WriteString(`<a:bodyPr rot="` + itoa(convertRotationDegrees(opts.CatAxisLabelRotate)) + `"/>`)
	} else {
		strXml.WriteString(`<a:bodyPr/>`)
	}
	strXml.WriteString(`    <a:lstStyle/>`)
	strXml.WriteString(`    <a:p>`)
	strXml.WriteString(`    <a:pPr>`)
	strXml.WriteString(`      <a:defRPr sz="` + itoa(int(jsRound(fontSizeOr(opts.CatAxisLabelFontSize, DEF_FONT_SIZE)*100))) + `" b="` + chartB2S(opts.CatAxisLabelFontBold) + `" i="` + chartB2S(opts.CatAxisLabelFontItalic) + `" u="none" strike="noStrike">`)
	strXml.WriteString(`      <a:solidFill>` + createColorElement(strOr(opts.CatAxisLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
	strXml.WriteString(`      <a:latin typeface="` + strOr(opts.CatAxisLabelFontFace, "Arial") + `"/>`)
	strXml.WriteString(`   </a:defRPr>`)
	strXml.WriteString(`  </a:pPr>`)
	strXml.WriteString(`  <a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `"/>`)
	strXml.WriteString(`  </a:p>`)
	strXml.WriteString(` </c:txPr>`)
	strXml.WriteString(` <c:crossAx val="` + valAxisID + `"/>`)
	crossTag, crossVal := crossesTagVal(opts.ValAxisCrossesAt)
	strXml.WriteString(` <c:` + crossTag + ` val="` + crossVal + `"/>`)
	strXml.WriteString(` <c:auto val="1"/>`)
	strXml.WriteString(` <c:lblAlgn val="ctr"/>`)
	noMulti := "1"
	if chartBool(opts.CatAxisMultiLevelLabels) {
		noMulti = "0"
	}
	strXml.WriteString(` <c:noMultiLvlLbl val="` + noMulti + `"/>`)
	if opts.CatAxisLabelFrequency != "" {
		strXml.WriteString(` <c:tickLblSkip val="` + opts.CatAxisLabelFrequency + `"/>`)
	}

	if opts.CatLabelFormatCode != "" || isXY {
		if opts.CatLabelFormatCode != "" {
			if timeUnitValid(opts.CatAxisBaseTimeUnit) {
				strXml.WriteString(`<c:baseTimeUnit val="` + strings.ToLower(opts.CatAxisBaseTimeUnit) + `"/>`)
			}
			if timeUnitValid(opts.CatAxisMajorTimeUnit) {
				strXml.WriteString(`<c:majorTimeUnit val="` + strings.ToLower(opts.CatAxisMajorTimeUnit) + `"/>`)
			}
			if timeUnitValid(opts.CatAxisMinorTimeUnit) {
				strXml.WriteString(`<c:minorTimeUnit val="` + strings.ToLower(opts.CatAxisMinorTimeUnit) + `"/>`)
			}
		}
		if opts.CatAxisMajorUnit != nil && *opts.CatAxisMajorUnit != 0 {
			strXml.WriteString(`<c:majorUnit val="` + ftoa(*opts.CatAxisMajorUnit) + `"/>`)
		}
		if opts.CatAxisMinorUnit != nil && *opts.CatAxisMinorUnit != 0 {
			strXml.WriteString(`<c:minorUnit val="` + ftoa(*opts.CatAxisMinorUnit) + `"/>`)
		}
	}

	if isXY {
		strXml.WriteString(`</c:valAx>`)
	} else {
		if opts.CatLabelFormatCode != "" {
			strXml.WriteString(`</c:dateAx>`)
		} else {
			strXml.WriteString(`</c:catAx>`)
		}
	}
	return strXml.String()
}

// makeValAxis ports gen-charts.ts makeValAxis.
func makeValAxis(opts *ChartOptions, valAxisID string) string {
	var axisPos string
	if valAxisID == AXIS_ID_VALUE_PRIMARY {
		if opts.BarDir == "col" {
			axisPos = "l"
		} else {
			axisPos = "b"
		}
	} else {
		if opts.BarDir != "col" {
			axisPos = "r"
		} else {
			axisPos = "t"
		}
	}
	if valAxisID == AXIS_ID_VALUE_SECONDARY {
		axisPos = "r"
	}
	crossAxID := AXIS_ID_CATEGORY_SECONDARY
	if valAxisID == AXIS_ID_VALUE_PRIMARY {
		crossAxID = AXIS_ID_CATEGORY_PRIMARY
	}

	var strXml strings.Builder
	strXml.WriteString(`<c:valAx>`)
	strXml.WriteString(`  <c:axId val="` + valAxisID + `"/>`)
	strXml.WriteString(`  <c:scaling>`)
	if opts.ValAxisLogScaleBase != nil && *opts.ValAxisLogScaleBase != 0 {
		strXml.WriteString(`<c:logBase val="` + ftoa(*opts.ValAxisLogScaleBase) + `"/>`)
	}
	strXml.WriteString(`<c:orientation val="` + strOr(opts.ValAxisOrientation, "minMax") + `"/>`)
	if opts.ValAxisMaxVal != nil {
		strXml.WriteString(`<c:max val="` + ftoa(*opts.ValAxisMaxVal) + `"/>`)
	}
	if opts.ValAxisMinVal != nil {
		strXml.WriteString(`<c:min val="` + ftoa(*opts.ValAxisMinVal) + `"/>`)
	}
	strXml.WriteString(`  </c:scaling>`)
	strXml.WriteString(`  <c:delete val="` + chartB2S(opts.ValAxisHidden) + `"/>`)
	strXml.WriteString(`  <c:axPos val="` + axisPos + `"/>`)
	if opts.ValGridLine != nil && opts.ValGridLine.Style != "none" {
		strXml.WriteString(createGridLineElement(opts.ValGridLine))
	}
	if chartBool(opts.ShowValAxisTitle) {
		strXml.WriteString(genXmlTitle(chartTitleOpts{
			Color:       opts.ValAxisTitleColor,
			FontFace:    opts.ValAxisTitleFontFace,
			FontSize:    opts.ValAxisTitleFontSize,
			TitleRotate: opts.ValAxisTitleRotate,
			Title:       strOr(opts.ValAxisTitle, "Axis Title"),
		}, 0, 0))
	}
	strXml.WriteString(`<c:numFmt formatCode="` + valAxisFmt(opts.ValAxisLabelFormatCode) + `" sourceLinked="0"/>`)
	if opts.Type == ChartTypeScatter {
		strXml.WriteString(`  <c:majorTickMark val="none"/>`)
		strXml.WriteString(`  <c:minorTickMark val="none"/>`)
		strXml.WriteString(`  <c:tickLblPos val="nextTo"/>`)
	} else {
		strXml.WriteString(` <c:majorTickMark val="` + strOr(opts.ValAxisMajorTickMark, "out") + `"/>`)
		strXml.WriteString(` <c:minorTickMark val="` + strOr(opts.ValAxisMinorTickMark, "none") + `"/>`)
		tickLblPos := opts.ValAxisLabelPos
		if tickLblPos == "" {
			if opts.BarDir == "col" {
				tickLblPos = "nextTo"
			} else {
				tickLblPos = "low"
			}
		}
		strXml.WriteString(` <c:tickLblPos val="` + tickLblPos + `"/>`)
	}
	strXml.WriteString(` <c:spPr>`)
	valLineW := ONEPT
	if opts.ValAxisLineSize != 0 {
		valLineW = valToPts(opts.ValAxisLineSize)
	}
	strXml.WriteString(`   <a:ln w="` + itoa(valLineW) + `" cap="flat">`)
	if !chartBool(opts.ValAxisLineShow) {
		strXml.WriteString(`<a:noFill/>`)
	} else {
		strXml.WriteString(`<a:solidFill>` + createColorElement(strOr(opts.ValAxisLineColor, DEF_CHART_GRIDLINE.Color), "") + `</a:solidFill>`)
	}
	strXml.WriteString(`     <a:prstDash val="` + strOr(opts.ValAxisLineStyle, "solid") + `"/>`)
	strXml.WriteString(`     <a:round/>`)
	strXml.WriteString(`   </a:ln>`)
	strXml.WriteString(` </c:spPr>`)
	strXml.WriteString(` <c:txPr>`)
	if opts.ValAxisLabelRotate != 0 {
		strXml.WriteString(`  <a:bodyPr rot="` + itoa(convertRotationDegrees(opts.ValAxisLabelRotate)) + `"/>`)
	} else {
		strXml.WriteString(`  <a:bodyPr/>`)
	}
	strXml.WriteString(`  <a:lstStyle/>`)
	strXml.WriteString(`  <a:p>`)
	strXml.WriteString(`    <a:pPr>`)
	strXml.WriteString(`      <a:defRPr sz="` + itoa(int(jsRound(fontSizeOr(opts.ValAxisLabelFontSize, DEF_FONT_SIZE)*100))) + `" b="` + chartB2S(opts.ValAxisLabelFontBold) + `" i="` + chartB2S(opts.ValAxisLabelFontItalic) + `" u="none" strike="noStrike">`)
	strXml.WriteString(`        <a:solidFill>` + createColorElement(strOr(opts.ValAxisLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
	strXml.WriteString(`        <a:latin typeface="` + strOr(opts.ValAxisLabelFontFace, "Arial") + `"/>`)
	strXml.WriteString(`      </a:defRPr>`)
	strXml.WriteString(`    </a:pPr>`)
	strXml.WriteString(`  <a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `"/>`)
	strXml.WriteString(`  </a:p>`)
	strXml.WriteString(` </c:txPr>`)
	strXml.WriteString(` <c:crossAx val="` + crossAxID + `"/>`)
	switch v := opts.CatAxisCrossesAt.(type) {
	case float64:
		strXml.WriteString(` <c:crossesAt val="` + ftoa(v) + `"/>`)
	case int:
		strXml.WriteString(` <c:crossesAt val="` + itoa(v) + `"/>`)
	case string:
		strXml.WriteString(` <c:crosses val="` + v + `"/>`)
	default:
		crosses := "autoZero"
		if axisPos == "r" || axisPos == "t" {
			crosses = "max"
		}
		strXml.WriteString(` <c:crosses val="` + crosses + `"/>`)
	}
	crossBetween := "between"
	if opts.Type == ChartTypeScatter || multiHasArea(opts) {
		crossBetween = "midCat"
	}
	strXml.WriteString(` <c:crossBetween val="` + crossBetween + `"/>`)
	if opts.ValAxisMajorUnit != nil && *opts.ValAxisMajorUnit != 0 {
		strXml.WriteString(` <c:majorUnit val="` + ftoa(*opts.ValAxisMajorUnit) + `"/>`)
	}
	if opts.ValAxisDisplayUnit != "" {
		lbl := ""
		if chartBool(opts.ValAxisDisplayUnitLabel) {
			lbl = `<c:dispUnitsLbl/>`
		}
		strXml.WriteString(`<c:dispUnits><c:builtInUnit val="` + opts.ValAxisDisplayUnit + `"/>` + lbl + `</c:dispUnits>`)
	}
	strXml.WriteString(`</c:valAx>`)
	return strXml.String()
}

// makeSerAxis ports gen-charts.ts makeSerAxis (used by bar3D).
func makeSerAxis(opts *ChartOptions, axisID, valAxisID string) string {
	var strXml strings.Builder
	strXml.WriteString(`<c:serAx>`)
	strXml.WriteString(`  <c:axId val="` + axisID + `"/>`)
	strXml.WriteString(`  <c:scaling><c:orientation val="` + strOr(opts.SerAxisOrientation, "minMax") + `"/></c:scaling>`)
	strXml.WriteString(`  <c:delete val="` + chartB2S(opts.SerAxisHidden) + `"/>`)
	strXml.WriteString(`  <c:axPos val="` + axPosBL(opts.BarDir) + `"/>`)
	if opts.SerGridLine != nil && opts.SerGridLine.Style != "none" {
		strXml.WriteString(createGridLineElement(opts.SerGridLine))
	}
	if chartBool(opts.ShowSerAxisTitle) {
		strXml.WriteString(genXmlTitle(chartTitleOpts{
			Color:       opts.SerAxisTitleColor,
			FontFace:    opts.SerAxisTitleFontFace,
			FontSize:    opts.SerAxisTitleFontSize,
			TitleRotate: opts.SerAxisTitleRotate,
			Title:       strOr(opts.SerAxisTitle, "Axis Title"),
		}, 0, 0))
	}
	strXml.WriteString(`  <c:numFmt formatCode="` + numFmtOrGeneral(opts.SerLabelFormatCode) + `" sourceLinked="0"/>`)
	strXml.WriteString(`  <c:majorTickMark val="out"/>`)
	strXml.WriteString(`  <c:minorTickMark val="none"/>`)
	// NOTE: mirrors TS operator precedence bug: `opts.serAxisLabelPos || opts.barDir === 'col' ? 'low' : 'nextTo'`
	// which parses as `(serAxisLabelPos || barDir==='col') ? 'low' : 'nextTo'`.
	serTick := "nextTo"
	if opts.SerAxisLabelPos != "" || opts.BarDir == "col" {
		serTick = "low"
	}
	strXml.WriteString(`  <c:tickLblPos val="` + serTick + `"/>`)
	strXml.WriteString(`  <c:spPr>`)
	strXml.WriteString(`    <a:ln w="12700" cap="flat">`)
	if !chartBool(opts.SerAxisLineShow) {
		strXml.WriteString(`<a:noFill/>`)
	} else {
		strXml.WriteString(`<a:solidFill>` + createColorElement(strOr(opts.SerAxisLineColor, DEF_CHART_GRIDLINE.Color), "") + `</a:solidFill>`)
	}
	strXml.WriteString(`      <a:prstDash val="solid"/>`)
	strXml.WriteString(`      <a:round/>`)
	strXml.WriteString(`    </a:ln>`)
	strXml.WriteString(`  </c:spPr>`)
	strXml.WriteString(`  <c:txPr>`)
	strXml.WriteString(`    <a:bodyPr/>`)
	strXml.WriteString(`    <a:lstStyle/>`)
	strXml.WriteString(`    <a:p>`)
	strXml.WriteString(`    <a:pPr>`)
	strXml.WriteString(`    <a:defRPr sz="` + itoa(int(jsRound(fontSizeOr(opts.SerAxisLabelFontSize, DEF_FONT_SIZE)*100))) + `" b="` + chartB2S(opts.SerAxisLabelFontBold) + `" i="` + chartB2S(opts.SerAxisLabelFontItalic) + `" u="none" strike="noStrike">`)
	strXml.WriteString(`      <a:solidFill>` + createColorElement(strOr(opts.SerAxisLabelColor, DEF_FONT_COLOR), "") + `</a:solidFill>`)
	strXml.WriteString(`      <a:latin typeface="` + strOr(opts.SerAxisLabelFontFace, "Arial") + `"/>`)
	strXml.WriteString(`   </a:defRPr>`)
	strXml.WriteString(`  </a:pPr>`)
	strXml.WriteString(`  <a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `"/>`)
	strXml.WriteString(`  </a:p>`)
	strXml.WriteString(` </c:txPr>`)
	strXml.WriteString(` <c:crossAx val="` + valAxisID + `"/>`)
	strXml.WriteString(` <c:crosses val="autoZero"/>`)
	if opts.SerAxisLabelFrequency != "" {
		strXml.WriteString(` <c:tickLblSkip val="` + opts.SerAxisLabelFrequency + `"/>`)
	}
	if opts.SerLabelFormatCode != "" {
		// M6 (fidelity): upstream NEVER emits serAxis base/major/minorTimeUnit.
		// gen-charts.ts:1883-1893 validates via `opt.toLowerCase()` where `opt`
		// is the KEY name string (e.g. "serAxisBaseTimeUnit"), never one of
		// 'days'/'months'/'years', so the guard always sets `opts[opt] = null`
		// and the three `if (opts.serAxis*TimeUnit)` emissions below it can never
		// fire. We replicate that bug for byte-fidelity by not emitting them.
		if opts.SerAxisMajorUnit != nil && *opts.SerAxisMajorUnit != 0 {
			strXml.WriteString(` <c:majorUnit val="` + ftoa(*opts.SerAxisMajorUnit) + `"/>`)
		}
		if opts.SerAxisMinorUnit != nil && *opts.SerAxisMinorUnit != 0 {
			strXml.WriteString(` <c:minorUnit val="` + ftoa(*opts.SerAxisMinorUnit) + `"/>`)
		}
	}
	strXml.WriteString(`</c:serAx>`)
	return strXml.String()
}

// genXmlTitle ports gen-charts.ts genXmlTitle. Output whitespace/newlines are
// significant and mirror the TS template literal exactly (LF newlines).
func genXmlTitle(opts chartTitleOpts, chartX, chartY float64) string {
	align := "<a:pPr>"
	if opts.TitleAlign == "left" || opts.TitleAlign == "right" {
		align = `<a:pPr algn="` + opts.TitleAlign[0:1] + `">`
	}
	rotate := "<a:bodyPr/>"
	if opts.TitleRotate != 0 {
		rotate = `<a:bodyPr rot="` + itoa(convertRotationDegrees(opts.TitleRotate)) + `"/>`
	}
	sizeAttr := ""
	if opts.FontSize != 0 {
		sizeAttr = `sz="` + itoa(int(jsRound(opts.FontSize*100))) + `"`
	}
	titleBold := "0"
	if chartBool(opts.TitleBold) {
		titleBold = "1"
	}

	layout := "<c:layout/>"
	if opts.TitlePos != nil {
		totalX := opts.TitlePos.X + chartX
		totalY := opts.TitlePos.Y + chartY
		valX := 0.0
		if totalX != 0 {
			valX = (totalX * (totalX / 5)) / 10
		}
		if valX >= 1 {
			valX = valX / 10
		}
		if valX >= 0.1 {
			valX = valX / 10
		}
		valY := 0.0
		if totalY != 0 {
			valY = (totalY * (totalY / 5)) / 10
		}
		if valY >= 1 {
			valY = valY / 10
		}
		if valY >= 0.1 {
			valY = valY / 10
		}
		layout = `<c:layout><c:manualLayout><c:xMode val="edge"/><c:yMode val="edge"/><c:x val="` + ftoa(valX) + `"/><c:y val="` + ftoa(valY) + `"/></c:manualLayout></c:layout>`
	}

	color := strOr(opts.Color, DEF_FONT_COLOR)
	fontFace := strOr(opts.FontFace, "Arial")

	return "<c:title>\n" +
		"      <c:tx>\n" +
		"        <c:rich>\n" +
		"          " + rotate + "\n" +
		"          <a:lstStyle/>\n" +
		"          <a:p>\n" +
		"            " + align + "\n" +
		"            <a:defRPr " + sizeAttr + " b=\"" + titleBold + "\" i=\"0\" u=\"none\" strike=\"noStrike\">\n" +
		"              <a:solidFill>" + createColorElement(color, "") + "</a:solidFill>\n" +
		"              <a:latin typeface=\"" + fontFace + "\"/>\n" +
		"            </a:defRPr>\n" +
		"          </a:pPr>\n" +
		"          <a:r>\n" +
		"            <a:rPr " + sizeAttr + " b=\"" + titleBold + "\" i=\"0\" u=\"none\" strike=\"noStrike\">\n" +
		"              <a:solidFill>" + createColorElement(color, "") + "</a:solidFill>\n" +
		"              <a:latin typeface=\"" + fontFace + "\"/>\n" +
		"            </a:rPr>\n" +
		"            <a:t>" + encodeXmlEntities(opts.Title) + "</a:t>\n" +
		"          </a:r>\n" +
		"        </a:p>\n" +
		"        </c:rich>\n" +
		"      </c:tx>\n" +
		"      " + layout + "\n" +
		"      <c:overlay val=\"0\"/>\n" +
		"    </c:title>"
}

// getExcelColName ports gen-charts.ts getExcelColName (1 -> "A", 27 -> "AA").
func getExcelColName(colIndex int) string {
	colIdx := colIndex - 1
	if colIdx <= 25 {
		return LETTERS[colIdx]
	}
	return LETTERS[int(math.Floor(float64(colIdx)/float64(len(LETTERS))-1))] + LETTERS[colIdx%len(LETTERS)]
}

// createShadowElement ports gen-charts.ts createShadowElement.
//
// The TS `{...defaults, ...options}` spread overrides every key the user object
// carries (key-presence based). With *float64 numeric fields, a non-nil field
// (including an explicit 0) overrides the default; a nil field falls through.
// options == nil returns `<a:effectLst/>`.
func createShadowElement(options *ShadowProps, defaults ShadowProps) string {
	if options == nil {
		return `<a:effectLst/>`
	}
	opts := defaults
	if options.Type != "" {
		opts.Type = options.Type
	}
	if options.Blur != nil {
		opts.Blur = options.Blur
	}
	if options.Offset != nil {
		opts.Offset = options.Offset
	}
	if options.Angle != nil {
		opts.Angle = options.Angle
	}
	if options.Color != "" {
		opts.Color = options.Color
	}
	if options.Opacity != nil {
		opts.Opacity = options.Opacity
	}
	if options.RotateWithShape != nil {
		opts.RotateWithShape = options.RotateWithShape
	}

	typ := opts.Type
	if typ == "" {
		typ = "outer"
	}
	blur := valToPts(fptrOr(opts.Blur, 0))
	offset := valToPts(fptrOr(opts.Offset, 0))
	angle := int(jsRound(fptrOr(opts.Angle, 0) * 60000))
	opacity := int(jsRound(fptrOr(opts.Opacity, 0) * 100000))
	rotShape := "0"
	if chartBool(opts.RotateWithShape) {
		rotShape = "1"
	}

	var b strings.Builder
	b.WriteString(`<a:effectLst>`)
	b.WriteString(`<a:` + typ + `Shdw sx="100000" sy="100000" kx="0" ky="0"  algn="bl" blurRad="` + itoa(blur) + `" rotWithShape="` + rotShape + `" dist="` + itoa(offset) + `" dir="` + itoa(angle) + `">`)
	b.WriteString(`<a:srgbClr val="` + opts.Color + `">`)
	b.WriteString(`<a:alpha val="` + itoa(opacity) + `"/></a:srgbClr>`)
	b.WriteString(`</a:` + typ + `Shdw>`)
	b.WriteString(`</a:effectLst>`)
	return b.String()
}

// createGridLineElement ports gen-charts.ts createGridLineElement.
func createGridLineElement(glOpts *OptsChartGridLine) string {
	size := glOpts.Size
	if size == 0 {
		size = DEF_CHART_GRIDLINE.Size
	}
	cap := glOpts.Cap
	if cap == "" {
		cap = DEF_CHART_GRIDLINE.Cap
	}
	color := strOr(glOpts.Color, DEF_CHART_GRIDLINE.Color)
	style := strOr(glOpts.Style, DEF_CHART_GRIDLINE.Style)

	var b strings.Builder
	b.WriteString(`<c:majorGridlines>`)
	b.WriteString(` <c:spPr>`)
	b.WriteString(`  <a:ln w="` + itoa(valToPts(size)) + `" cap="` + createLineCap(cap) + `">`)
	b.WriteString(`  <a:solidFill><a:srgbClr val="` + color + `"/></a:solidFill>`)
	b.WriteString(`   <a:prstDash val="` + style + `"/><a:round/>`)
	b.WriteString(`  </a:ln>`)
	b.WriteString(` </c:spPr>`)
	b.WriteString(`</c:majorGridlines>`)
	return b.String()
}

// createLineCap ports gen-charts.ts createLineCap. Unknown values default to
// "flat" (the TS throws; library code avoids panics).
func createLineCap(lineCap string) string {
	switch lineCap {
	case "", "flat":
		return "flat"
	case "square":
		return "sq"
	case "round":
		return "rnd"
	default:
		return "flat"
	}
}

// ---------------------------------------------------------------------------
// tiny value helpers used by the ports above
// ---------------------------------------------------------------------------

// strOr returns s when non-empty, else def (mirrors TS `s || def`).
func strOr(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

// fontSizeOr returns fs when non-zero, else def (mirrors `fs || DEF`).
func fontSizeOr(fs float64, def int) float64 {
	if fs != 0 {
		return fs
	}
	return float64(def)
}

// numFmtOrGeneral mirrors `encodeXmlEntities(code) || 'General'`.
func numFmtOrGeneral(code string) string {
	if code == "" {
		return "General"
	}
	return encodeXmlEntities(code)
}

// valAxisFmt mirrors `code ? encodeXmlEntities(code) : 'General'`.
func valAxisFmt(code string) string {
	if code != "" {
		return encodeXmlEntities(code)
	}
	return "General"
}

// axPosBL returns "b" for a column bar dir, else "l".
func axPosBL(barDir string) string {
	if barDir == "col" {
		return "b"
	}
	return "l"
}

// crossesTagVal implements the catAxis `valAxisCrossesAt` cross tag selection.
func crossesTagVal(v any) (tag, val string) {
	switch n := v.(type) {
	case float64:
		if n != 0 {
			return "crossesAt", ftoa(n)
		}
		return "crossesAt", "autoZero"
	case int:
		if n != 0 {
			return "crossesAt", itoa(n)
		}
		return "crossesAt", "autoZero"
	case string:
		if n != "" {
			return "crosses", n
		}
		return "crosses", "autoZero"
	default:
		return "crosses", "autoZero"
	}
}

// multiHasArea reports whether a combo chart includes an AREA component.
func multiHasArea(opts *ChartOptions) bool {
	for _, m := range opts.MultiTypes {
		if m.Type == ChartTypeArea {
			return true
		}
	}
	return false
}

// timeUnitValid reports whether a time-unit string is one of days/months/years.
func timeUnitValid(s string) bool {
	switch strings.ToLower(s) {
	case "days", "months", "years":
		return true
	}
	return false
}

// markerColorIdx mirrors `dataIndex + 1 > len ? random : dataIndex`.
func markerColorIdx(dataIndex, colorsLen int) int {
	if dataIndex+1 > colorsLen {
		return int(math.Floor(rand.Float64() * float64(colorsLen)))
	}
	return dataIndex
}

// pieColorIdx mirrors `idx + 1 > len ? random : idx`.
func pieColorIdx(idx, colorsLen int) int {
	if idx+1 > colorsLen {
		return int(math.Floor(rand.Float64() * float64(colorsLen)))
	}
	return idx
}

// numAt mirrors `arr[idx] || arr[idx] === 0 ? arr[idx] : ”` (prints 0, blank
// when out of range).
func numAt(arr []float64, idx int) string {
	if idx < len(arr) {
		return ftoa(arr[idx])
	}
	return ""
}

// numTruthy mirrors `value || ”` (0 and NaN print blank).
func numTruthy(v float64) string {
	if v != 0 {
		return ftoa(v)
	}
	return ""
}

// isBlankLabel mirrors `/^ *$/.test(label)` (empty or spaces only).
func isBlankLabel(s string) bool {
	return strings.TrimLeft(s, " ") == ""
}

// scatterUniqueID mirrors `'00000000'.substring(0, 8 - String(n).length) + n`.
func scatterUniqueID(n int) string {
	ns := itoa(n)
	pad := 8 - len(ns)
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat("0", pad) + ns
}

// overlayChartOptions returns a copy of base with the "present" (non-zero)
// fields of over applied on top, approximating the TS object spread
// `{...base, ...over}`. Used for combo-chart per-type option merges and axis
// merges. Deviation: false booleans and zero numbers in over cannot override
// base (indistinguishable from unset). Combo/axis merges are not covered by
// golden files; single-type charts never call this.
func overlayChartOptions(base *ChartOptions, over *ChartOptions) *ChartOptions {
	merged := *base
	if over == nil {
		return &merged
	}
	applyOverlay(&merged, over)
	return &merged
}

// applyOverlay copies the non-zero fields of src onto dst (recursing into
// embedded structs), approximating JS `{...dst, ...src}`.
func applyOverlay(dst, src *ChartOptions) {
	overlayValue(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem())
}

func overlayValue(dst, src reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		sf := src.Field(i)
		df := dst.Field(i)
		if !df.CanSet() {
			continue
		}
		if sf.Kind() == reflect.Struct {
			overlayValue(df, sf)
			continue
		}
		if !sf.IsZero() {
			df.Set(sf)
		}
	}
}

// ---------------------------------------------------------------------------
// createExcelWorksheet
// ---------------------------------------------------------------------------

// excelNowFunc supplies the timestamp for docProps/core.xml. Overridable in
// tests. Mirrors the TS `new Date().toISOString()`.
var excelNowFunc = func() time.Time { return time.Now().UTC() }

// createExcelWorksheet builds the embedded .xlsx (data source for a chart) and
// returns its complete bytes. Ports gen-charts.ts createExcelWorksheet, redesigned
// for Go stdlib: builds an in-memory zip (archive/zip) with STORE compression
// (matching JSZip's default). The internal part CONTENTS are byte-identical to
// the TS templates; zip container metadata differs (that is acceptable).
func createExcelWorksheet(chartObject *SlideRelChart) ([]byte, error) {
	data := chartObject.Data
	opts := chartObject.Opts
	intBubbleCols := (len(data)-1)*2 + 1
	isMultiCatAxes := len(data) > 0 && len(data[0].Labels) > 1
	isBubble := opts.Type == ChartTypeBubble || opts.Type == ChartTypeBubble3d
	isScatter := opts.Type == ChartTypeScatter

	// sharedStrings.xml
	var ss strings.Builder
	ss.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	if isBubble {
		ss.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="` + itoa(intBubbleCols) + `" uniqueCount="` + itoa(intBubbleCols) + `">`)
	} else if isScatter {
		ss.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="` + itoa(len(data)) + `" uniqueCount="` + itoa(len(data)) + `">`)
	} else if isMultiCatAxes {
		totCount := len(data)
		for _, arrLabel := range data[0].Labels {
			totCount += countNonEmpty(arrLabel)
		}
		ss.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="` + itoa(totCount) + `" uniqueCount="` + itoa(totCount) + `">`)
		ss.WriteString(`<si><t/></si>`)
	} else {
		totCount := len(data) + len(data[0].Labels)*len(data[0].Labels[0]) + len(data[0].Labels)
		unqCount := len(data) + len(data[0].Labels)*len(data[0].Labels[0]) + 1
		ss.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="` + itoa(totCount) + `" uniqueCount="` + itoa(unqCount) + `">`)
		ss.WriteString(`<si><t xml:space="preserve"></t></si>`)
	}

	// C: names/series
	if isBubble {
		for idx := range data {
			if idx == 0 {
				ss.WriteString(`<si><t>X-Axis</t></si>`)
			} else {
				name := data[idx].Name
				if name == "" {
					name = "Y-Axis" + itoa(idx)
				}
				ss.WriteString(`<si><t>` + encodeXmlEntities(name) + `</t></si>`)
				ss.WriteString(`<si><t>` + encodeXmlEntities("Size"+itoa(idx)) + `</t></si>`)
			}
		}
	} else {
		for idx := range data {
			name := data[idx].Name
			if name == "" {
				name = " "
			}
			ss.WriteString(`<si><t>` + encodeXmlEntities(strings.Replace(name, "X-Axis", "X-Values", 1)) + `</t></si>`)
		}
	}

	// D: labels/categories
	if !isBubble && !isScatter {
		for gi := len(data[0].Labels) - 1; gi >= 0; gi-- {
			for _, label := range data[0].Labels[gi] {
				if label != "" {
					ss.WriteString(`<si><t>` + encodeXmlEntities(label) + `</t></si>`)
				}
			}
		}
	}
	ss.WriteString("</sst>\n")

	// tables/table1.xml
	var tbl strings.Builder
	tbl.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	if isBubble {
		tbl.WriteString(`<table xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" id="1" name="Table1" displayName="Table1" ref="A1:` + getExcelColName(intBubbleCols) + itoa(intBubbleCols) + `" totalsRowShown="0">`)
		tbl.WriteString(`<tableColumns count="` + itoa(intBubbleCols) + `">`)
		idxColLtr := 1
		for idx := range data {
			if idx == 0 {
				tbl.WriteString(`<tableColumn id="` + itoa(idx+1) + `" name="X-Values"/>`)
			} else {
				tbl.WriteString(`<tableColumn id="` + itoa(idx+idxColLtr) + `" name="` + data[idx].Name + `"/>`)
				idxColLtr++
				tbl.WriteString(`<tableColumn id="` + itoa(idx+idxColLtr) + `" name="Size` + itoa(idx) + `"/>`)
			}
		}
	} else if isScatter {
		tbl.WriteString(`<table xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" id="1" name="Table1" displayName="Table1" ref="A1:` + getExcelColName(len(data)) + itoa(len(data[0].Values)+1) + `" totalsRowShown="0">`)
		tbl.WriteString(`<tableColumns count="` + itoa(len(data)) + `">`)
		for idx := range data {
			nm := "Y-Value "
			if idx == 0 {
				nm = "X-Values"
			}
			tbl.WriteString(`<tableColumn id="` + itoa(idx+1) + `" name="` + nm + itoa(idx) + `"/>`)
		}
	} else {
		tbl.WriteString(`<table xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" id="1" name="Table1" displayName="Table1" ref="A1:` + getExcelColName(len(data)+len(data[0].Labels)) + itoa(len(data[0].Labels[0])+1) + `'" totalsRowShown="0">`)
		tbl.WriteString(`<tableColumns count="` + itoa(len(data)+len(data[0].Labels)) + `">`)
		for idx := range data[0].Labels {
			tbl.WriteString(`<tableColumn id="` + itoa(idx+1) + `" name="Column` + itoa(idx+1) + `"/>`)
		}
		for idx := range data {
			tbl.WriteString(`<tableColumn id="` + itoa(idx+len(data[0].Labels)+1) + `" name="` + encodeXmlEntities(data[idx].Name) + `"/>`)
		}
	}
	tbl.WriteString(`</tableColumns>`)
	tbl.WriteString(`<tableStyleInfo showFirstColumn="0" showLastColumn="0" showRowStripes="1" showColumnStripes="0"/>`)
	tbl.WriteString(`</table>`)

	// worksheets/sheet1.xml
	sheet := buildSheet1(data, opts, intBubbleCols, isBubble, isScatter, isMultiCatAxes)

	// docProps/core.xml (timestamped)
	ts := excelNowFunc().Format("2006-01-02T15:04:05.000Z07:00")
	core := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">` +
		`<dc:creator>PptxGenJS</dc:creator>` +
		`<cp:lastModifiedBy>PptxGenJS</cp:lastModifiedBy>` +
		`<dcterms:created xsi:type="dcterms:W3CDTF">` + ts + `</dcterms:created>` +
		`<dcterms:modified xsi:type="dcterms:W3CDTF">` + ts + `</dcterms:modified>` +
		`</cp:coreProperties>`

	// Assemble zip (STORE)
	parts := []struct{ name, content string }{
		{"[Content_Types].xml", excelContentTypes},
		{"_rels/.rels", excelRootRels},
		{"docProps/app.xml", excelAppXML},
		{"docProps/core.xml", core},
		{"xl/_rels/workbook.xml.rels", excelWorkbookRels},
		{"xl/styles.xml", excelStylesXML},
		{"xl/theme/theme1.xml", excelThemeXML},
		{"xl/workbook.xml", excelWorkbookXML},
		{"xl/worksheets/_rels/sheet1.xml.rels", excelSheet1Rels},
		{"xl/sharedStrings.xml", ss.String()},
		{"xl/tables/table1.xml", tbl.String()},
		{"xl/worksheets/sheet1.xml", sheet},
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, p := range parts {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: p.name, Method: zip.Store})
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(p.content)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildSheet1 ports the worksheets/sheet1.xml body of createExcelWorksheet.
func buildSheet1(data []ChartData, opts *ChartOptions, intBubbleCols int, isBubble, isScatter, isMultiCatAxes bool) string {
	var s strings.Builder
	s.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	s.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" mc:Ignorable="x14ac" xmlns:x14ac="http://schemas.microsoft.com/office/spreadsheetml/2009/9/ac">`)

	if isBubble {
		s.WriteString(`<dimension ref="A1:` + getExcelColName(intBubbleCols) + itoa(len(data[0].Values)+1) + `"/>`)
	} else if isScatter {
		s.WriteString(`<dimension ref="A1:` + getExcelColName(len(data)) + itoa(len(data[0].Values)+1) + `"/>`)
	} else {
		s.WriteString(`<dimension ref="A1:` + getExcelColName(len(data)+1) + itoa(len(data[0].Values)+1) + `"/>`)
	}

	s.WriteString(`<sheetViews><sheetView tabSelected="1" workbookViewId="0"><selection activeCell="B1" sqref="B1"/></sheetView></sheetViews>`)
	s.WriteString(`<sheetFormatPr baseColWidth="10" defaultRowHeight="16"/>`)

	if isBubble {
		s.WriteString(`<sheetData>`)
		s.WriteString(`<row r="1" spans="1:` + itoa(intBubbleCols) + `">`)
		s.WriteString(`<c r="A1" t="s"><v>0</v></c>`)
		for idx := 1; idx < intBubbleCols; idx++ {
			s.WriteString(`<c r="` + getExcelColName(idx+1) + `1" t="s"><v>` + itoa(idx) + `</v></c>`)
		}
		s.WriteString(`</row>`)
		for idx, val := range data[0].Values {
			s.WriteString(`<row r="` + itoa(idx+2) + `" spans="1:` + itoa(intBubbleCols) + `">`)
			s.WriteString(`<c r="A` + itoa(idx+2) + `"><v>` + ftoa(val) + `</v></c>`)
			idxColLtr := 2
			for idy := 1; idy < len(data); idy++ {
				s.WriteString(`<c r="` + getExcelColName(idxColLtr) + itoa(idx+2) + `"><v>` + numTruthy(numOrEmptyRaw(data[idy].Values, idx)) + `</v></c>`)
				idxColLtr++
				s.WriteString(`<c r="` + getExcelColName(idxColLtr) + itoa(idx+2) + `"><v>` + numTruthy(numOrEmptyRaw(data[idy].Sizes, idx)) + `</v></c>`)
				idxColLtr++
			}
			s.WriteString(`</row>`)
		}
	} else if isScatter {
		s.WriteString(`<sheetData>`)
		s.WriteString(`<row r="1" spans="1:` + itoa(len(data)) + `">`)
		for idx := range data {
			s.WriteString(`<c r="` + getExcelColName(idx+1) + `1" t="s"><v>` + itoa(idx) + `</v></c>`)
		}
		s.WriteString(`</row>`)
		for idx, val := range data[0].Values {
			s.WriteString(`<row r="` + itoa(idx+2) + `" spans="1:` + itoa(len(data)) + `">`)
			s.WriteString(`<c r="A` + itoa(idx+2) + `"><v>` + ftoa(val) + `</v></c>`)
			for idy := 1; idy < len(data); idy++ {
				s.WriteString(`<c r="` + getExcelColName(idy+1) + itoa(idx+2) + `"><v>` + numAt(data[idy].Values, idx) + `</v></c>`)
			}
			s.WriteString(`</row>`)
		}
	} else if !isMultiCatAxes {
		s.WriteString(`<sheetData>`)
		s.WriteString(`<row r="1" spans="1:` + itoa(len(data)+len(data[0].Labels)) + `">`)
		for idx := range data[0].Labels {
			s.WriteString(`<c r="` + getExcelColName(idx+1) + `1" t="s"><v>0</v></c>`)
		}
		for idx := range data {
			s.WriteString(`<c r="` + getExcelColName(idx+1+len(data[0].Labels)) + `1" t="s"><v>` + itoa(idx+1) + `</v></c>`)
		}
		s.WriteString(`</row>`)
		for idx := range data[0].Labels[0] {
			s.WriteString(`<row r="` + itoa(idx+2) + `" spans="1:` + itoa(len(data)+len(data[0].Labels)) + `">`)
			for idx2 := len(data[0].Labels) - 1; idx2 >= 0; idx2-- {
				s.WriteString(`<c r="` + getExcelColName(len(data[0].Labels)-idx2) + itoa(idx+2) + `" t="s">`)
				s.WriteString(`<v>` + itoa(len(data)+idx+1) + `</v>`)
				s.WriteString(`</c>`)
			}
			for idy := range data {
				s.WriteString(`<c r="` + getExcelColName(len(data[0].Labels)+idy+1) + itoa(idx+2) + `"><v>` + numTruthy(numOrEmptyRaw(data[idy].Values, idx)) + `</v></c>`)
			}
			s.WriteString(`</row>`)
		}
	} else {
		s.WriteString(`<sheetData>`)
		s.WriteString(`<row r="1" spans="1:` + itoa(len(data)+len(data[0].Labels)) + `">`)
		for idx := range data[0].Labels {
			s.WriteString(`<c r="` + getExcelColName(idx+1) + `1" t="s"><v>0</v></c>`)
		}
		for idx := len(data[0].Labels) - 1; idx < len(data)+len(data[0].Labels)-1; idx++ {
			s.WriteString(`<c r="` + getExcelColName(idx+len(data[0].Labels)) + `1" t="s"><v>` + itoa(idx) + `</v></c>`)
		}
		s.WriteString(`</row>`)

		totSer := len(data)
		totCat := len(data[0].Labels[0])
		totLvl := len(data[0].Labels)
		revLabelGroups := reversedLabels(data[0].Labels)
		for idx := 0; idx < totCat; idx++ {
			s.WriteString(`<row r="` + itoa(idx+2) + `" spans="1:` + itoa(totSer+totLvl) + `">`)
			totLabels := totSer
			for idy, labelsGroup := range revLabelGroups {
				var colLabel string
				if idx < len(labelsGroup) {
					colLabel = labelsGroup[idx]
				}
				if colLabel != "" {
					totGrpLbls := 1
					if idy != 0 {
						totGrpLbls = countNonEmpty(revLabelGroups[idy-1])
					}
					totLabels += totGrpLbls
					s.WriteString(`<c r="` + getExcelColName(idx+1+idy) + itoa(idx+2) + `" t="s"><v>` + itoa(totLabels) + `</v></c>`)
				}
			}
			for idy := 0; idy < totSer; idy++ {
				v := 0.0
				if idx < len(data[idy].Values) {
					v = data[idy].Values[idx]
				}
				s.WriteString(`<c r="` + getExcelColName(totLvl+idy+1) + itoa(idx+2) + `"><v>` + ftoa(v) + `</v></c>`)
			}
			s.WriteString(`</row>`)
		}
	}
	s.WriteString(`</sheetData>`)
	s.WriteString(`<pageMargins left="0.7" right="0.7" top="0.75" bottom="0.75" header="0.3" footer="0.3"/>`)
	s.WriteString("</worksheet>\n")
	return s.String()
}

// numOrEmptyRaw returns the value at idx or 0 (for the numTruthy `|| ”` idiom).
func numOrEmptyRaw(arr []float64, idx int) float64 {
	if idx < len(arr) {
		return arr[idx]
	}
	return 0
}

// countNonEmpty counts non-empty strings in a label group.
func countNonEmpty(labels []string) int {
	n := 0
	for _, l := range labels {
		if l != "" {
			n++
		}
	}
	return n
}

// reversedLabels returns a reversed copy of the label groups.
func reversedLabels(labels [][]string) [][]string {
	out := make([][]string, len(labels))
	for i := range labels {
		out[len(labels)-1-i] = labels[i]
	}
	return out
}
