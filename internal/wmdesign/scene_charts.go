package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

type sceneChartSeries struct {
	Name   string     `json:"name"`
	Values []*float64 `json:"values"`
	Color  string     `json:"color,omitempty"`
	Dashed bool       `json:"dashed,omitempty"`
}
type sceneQuadrant struct {
	Name  string `json:"name"`
	State string `json:"state,omitempty"`
	Tag   string `json:"tag,omitempty"`
}
type sceneQuadrantItem struct {
	X     float64   `json:"x"`
	Y     float64   `json:"y"`
	Label string    `json:"label"`
	To    []float64 `json:"to,omitempty"`
}

// Quadrant keys are a closed false/true/"markers" union.
type sceneQuadrantKey string

func (k *sceneQuadrantKey) UnmarshalJSON(raw []byte) error {
	switch string(raw) {
	case "false":
		*k = ""
	case "true":
		*k = "legend"
	case `"markers"`:
		*k = "markers"
	default:
		return fmt.Errorf("scene.quadrant_key_requires_boolean_or_markers")
	}
	return nil
}
func (k sceneQuadrantKey) numbered() bool    { return k != "" }
func (k sceneQuadrantKey) ownLegend() bool   { return k == "legend" }
func (k sceneQuadrantKey) markersOnly() bool { return k == "markers" }

type sceneChartSource struct {
	Type        string             `json:"type"`
	Kind        string             `json:"kind"`
	X           float64            `json:"x"`
	Y           float64            `json:"y"`
	W           float64            `json:"w"`
	H           float64            `json:"h"`
	Title       string             `json:"title,omitempty"`
	Units       string             `json:"units,omitempty"`
	Source      string             `json:"source,omitempty"`
	Categories  []string           `json:"categories,omitempty"`
	Series      []sceneChartSeries `json:"series,omitempty"`
	Colors      []string           `json:"colors,omitempty"`
	Highlight   []int              `json:"highlight,omitempty"`
	Mode        string             `json:"mode,omitempty"`
	Format      *NumberFormatSpec  `json:"format,omitempty"`
	ValueSuffix string             `json:"valueSuffix,omitempty"`
	YMin        *float64           `json:"yMin,omitempty"`
	XTitle      string             `json:"xTitle,omitempty"`
	YTitle      string             `json:"yTitle,omitempty"`
	Target      *struct {
		Value float64 `json:"value"`
		Label string  `json:"label"`
	} `json:"target,omitempty"`
	Progress     *float64                 `json:"progress,omitempty"`
	Center       *sceneMetricValue        `json:"center,omitempty"`
	Points       [][]json.RawMessage      `json:"points,omitempty"`
	Labels       map[string]string        `json:"labels,omitempty"`
	Trend        bool                     `json:"trend,omitempty"`
	Quadrants    map[string]sceneQuadrant `json:"quadrants,omitempty"`
	Items        []sceneQuadrantItem      `json:"items,omitempty"`
	Key          sceneQuadrantKey         `json:"key,omitempty"`
	Style        string                   `json:"style,omitempty"`
	Fill         string                   `json:"fill,omitempty"`
	StrongState  string                   `json:"strongState,omitempty"`
	PositionMode string                   `json:"positionMode,omitempty"`
}

func (r *renderer) planChartScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "chart" {
		return nil, false, nil
	}
	var n sceneChartSource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	p, err := r.sceneSourceChart(id, n, ctx)
	return p, true, err
}

func sceneChartNiceMax(v float64) float64 {
	if v <= 0 {
		return 1
	}
	e := math.Pow(10, math.Floor(math.Log10(v)))
	for _, f := range []float64{1, 1.2, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10} {
		if v <= f*e {
			return f * e
		}
	}
	return e * 10
}
func sceneChartFloat(v float64) *float64 { return &v }
func sceneChartLayout(x, y, w, h float64) *pptx.PositionProps {
	xx, yy, ww, hh := pptx.Inches(x), pptx.Inches(y), pptx.Inches(w), pptx.Inches(h)
	return &pptx.PositionProps{X: &xx, Y: &yy, W: &ww, H: &hh}
}
func (r *renderer) sceneChartOptions(surface string) (pptx.ChartOptions, error) {
	st, err := r.sceneStyle("small")
	if err != nil {
		return pptx.ChartOptions{}, err
	}
	font, err := r.typeEngine.Resolve(st)
	if err != nil {
		return pptx.ChartOptions{}, err
	}
	label, err := r.sceneDataToken("label", 600)
	if err != nil {
		return pptx.ChartOptions{}, err
	}
	lf, err := r.typeEngine.Resolve(label)
	if err != nil {
		return pptx.ChartOptions{}, err
	}
	mono, err := r.sceneStyle("number")
	if err != nil {
		return pptx.ChartOptions{}, err
	}
	mf, err := r.typeEngine.Resolve(mono)
	if err != nil {
		return pptx.ChartOptions{}, err
	}
	navy, _ := r.sceneColor(surface, "strong")
	secondary, _ := r.sceneColor(surface, "secondary")
	grid, _ := r.sceneColor("subtle", "bg")
	o := pptx.ChartOptions{TextBaseProps: pptx.TextBaseProps{FontFace: font.Typeface, FontSize: 12, Color: navy}, ShowTitle: ptrSceneBool(false), ShowLegend: ptrSceneBool(false), ShowValue: ptrSceneBool(true), ShowSerName: ptrSceneBool(false), ShowLabel: ptrSceneBool(false), ShowPercent: ptrSceneBool(false), ShowLeaderLines: ptrSceneBool(false), DisplayBlanksAs: "gap", DataBorder: &pptx.BorderProps{Color: navy, Pt: 1}, DataLabelFontFace: mf.Typeface, DataLabelFontSize: 10, DataLabelFontBold: ptrSceneBool(true), DataLabelColor: navy, DataLabelFormatCode: "#,##0.#", DataLabelPosition: "outEnd", CatAxisLabelFontFace: lf.Typeface, CatAxisLabelFontSize: 9, CatAxisLabelColor: secondary, CatAxisLabelFontBold: ptrSceneBool(false), CatAxisLineColor: navy, CatAxisLineShow: ptrSceneBool(true), CatAxisLineSize: 1, CatAxisMajorTickMark: "none", CatAxisMinorTickMark: "none", ValAxisLabelFontFace: lf.Typeface, ValAxisLabelFontSize: 9, ValAxisLabelColor: secondary, ValAxisLineColor: navy, ValAxisLineShow: ptrSceneBool(false), ValAxisMajorTickMark: "none", ValAxisMinorTickMark: "none", ValGridLine: &pptx.OptsChartGridLine{Color: grid, Size: .75, Style: "none"}, CatGridLine: &pptx.OptsChartGridLine{Style: "none"}, SerGridLine: &pptx.OptsChartGridLine{Style: "none"}, LegendFontFace: font.Typeface, LegendFontSize: 12, LegendColor: navy, LegendPos: "t", TitleFontFace: font.Typeface, TitleFontSize: 18, TitleBold: ptrSceneBool(true), TitleColor: navy, CatAxisTitleFontFace: lf.Typeface, CatAxisTitleFontSize: 9, ValAxisTitleFontFace: lf.Typeface, ValAxisTitleFontSize: 9, SerAxisTitleFontFace: lf.Typeface, SerAxisTitleFontSize: 9, SerAxisLabelFontFace: lf.Typeface, SerAxisLabelFontSize: 9, SerAxisLabelColor: secondary, ChartArea: &pptx.ChartAreaProps{RoundedCorners: ptrSceneBool(false), Fill: &pptx.ShapeFillProps{Type: "none"}, Border: &pptx.BorderProps{Type: "none"}}, PlotArea: &pptx.ChartFillLineProps{Fill: &pptx.ShapeFillProps{Type: "none"}, Border: &pptx.BorderProps{Type: "none"}}}
	return o, nil
}

func sceneChartFormatCode(n sceneChartSource) (string, error) {
	if n.Format == nil {
		suffix := n.ValueSuffix
		if suffix == "" {
			return "#,##0.#", nil
		}
		if strings.ContainsAny(suffix, "\";[]\\") {
			return "", fmt.Errorf("scene.chart_unsupported_suffix")
		}
		if strings.HasSuffix(suffix, "s") {
			return `[=1]#,##0.#"` + strings.TrimSuffix(suffix, "s") + `";#,##0.#"` + suffix + `"`, nil
		}
		return `#,##0.#"` + suffix + `"`, nil
	}
	if n.ValueSuffix != "" {
		return "", fmt.Errorf("scene.chart_format_suffix_union")
	}
	f := *n.Format
	if len(f.Value) > 0 && string(f.Value) != "null" {
		return "", fmt.Errorf("scene.chart_format_value_is_bound_from_series")
	}
	if f.Kind == "range" {
		return "", fmt.Errorf("scene.chart_unsupported_number_format")
	}
	validation := f
	validation.Value = json.RawMessage("0")
	if _, err := FormatNumber(validation); err != nil {
		return "", err
	}
	if f.Unit != "" && strings.ContainsAny(f.Unit, "\";[]\\") {
		return "", fmt.Errorf("scene.chart_invalid_format_unit")
	}
	switch f.Kind {
	case "number":
		code := "#,##0.##"
		if f.Unit != "" {
			code += `" ` + f.Unit + `"`
		}
		return code, nil
	case "percent":
		return `[<0.1]0.#%;0%`, nil
	case "currency":
		symbol := map[string]string{"": "$", "USD": "$", "EUR": "€", "GBP": "£"}[f.Currency]
		if symbol == "" {
			return "", fmt.Errorf("scene.chart_currency_enum")
		}
		switch f.Scale {
		case "K":
			return `"` + symbol + `"#,##0,"K"`, nil
		case "M":
			return `"` + symbol + `"0.#,,"M"`, nil
		case "B":
			return `"` + symbol + `"0.#,,,"B"`, nil
		case "":
			return `[>=1000000]"` + symbol + `"0.#,,"M";[>=1000]"` + symbol + `"#,##0,"K";"` + symbol + `"#,##0`, nil
		}
	case "delta":
		if f.Unit == "pts" {
			return `+0.#%" pts";"−"0.#%" pts"`, nil
		}
		return `+0.##" ` + f.Unit + `";"−"0.##" ` + f.Unit + `"`, nil
	}
	return "", fmt.Errorf("scene.chart_unsupported_number_format")
}

func (r *renderer) sceneSourceChart(id string, n sceneChartSource, ctx SceneContext) (*scenePlan, error) {
	b := Rect{n.X, n.Y, n.W, n.H}
	if b.W <= 0 || b.H <= 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
		return nil, fmt.Errorf("scene.chart_invalid_geometry")
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	p := &scenePlan{ID: id, Bounds: b}
	top, bottom := b.Y, b.Y+b.H
	if n.Title != "" {
		st, _ := r.sceneDataToken("subhead", 600)
		if _, err := r.sceneDataText(p, id+".title", n.Title, st, Rect{b.X, top, b.W, 24}, surface, "display", "left"); err != nil {
			return nil, err
		}
		top += 24
	}
	if n.Units != "" {
		st, _ := r.sceneStyle("label")
		st.TrackingPt = .6
		if _, err := r.sceneDataText(p, id+".units", n.Units, st, Rect{b.X, top, b.W, 15}, surface, "secondary", "left"); err != nil {
			return nil, err
		}
		top += 15
	}
	if n.Source != "" {
		st, _ := r.sceneStyle("source")
		if _, err := r.sceneDataText(p, id+".source", "Source: "+n.Source, st, Rect{b.X, bottom - 15, b.W, 15}, surface, "secondary", "left"); err != nil {
			return nil, err
		}
		bottom -= 18
	}
	if n.Kind != "quadrant" || r.source.Revision != LibraryRevisionV2 {
		top += 9
	}
	plot := Rect{b.X, top, b.W, bottom - top}
	if plot.H <= 36 {
		return nil, fmt.Errorf("scene.chart_no_plot_capacity")
	}
	if n.Kind == "quadrant" {
		if err := r.sceneQuadrantChart(p, id, n, plot, surface, ctx); err != nil {
			return nil, err
		}
		if r.source.Revision == LibraryRevisionV2 {
			p.Bounds = Rect{}
			for _, item := range p.Items {
				if item.Shape != nil {
					p.Bounds = diagramUnion(p.Bounds, item.Shape.Record.Rect)
				}
				if item.Text != nil {
					p.Bounds = diagramUnion(p.Bounds, diagramRotatedRect(item.Text.Rect, item.Text.Rotation))
				}
				if item.Chart != nil {
					p.Bounds = diagramUnion(p.Bounds, item.Chart.Rect)
				}
			}
		}
		sceneDataGroup(p, id, "chart.quadrant.source", 0, p.Bounds)
		return p, nil
	}
	o, err := r.sceneChartOptions(surface)
	if err != nil {
		return nil, err
	}
	o.DataLabelFormatCode, err = sceneChartFormatCode(n)
	if err != nil {
		return nil, err
	}
	typ := pptx.ChartTypeBar
	switch n.Kind {
	case "column", "bar", "line", "pie", "doughnut", "scatter":
	default:
		return nil, fmt.Errorf("scene.chart_kind_unsupported: %s", n.Kind)
	}
	if n.Kind == "scatter" {
		return r.sceneScatterChart(p, id, n, plot, surface, ctx, o)
	}
	if n.Kind == "pie" || n.Kind == "doughnut" {
		if n.Progress != nil {
			if *n.Progress < 0 || *n.Progress > 1 || len(n.Series) > 0 {
				return nil, fmt.Errorf("scene.chart_progress_union")
			}
			a, z := *n.Progress, 1-*n.Progress
			n.Categories = []string{"Complete", "Remaining"}
			n.Series = []sceneChartSeries{{Name: "Progress", Values: []*float64{&a, &z}}}
			n.Colors = []string{"series.1", "deemph.1"}
		}
		if len(n.Series) != 1 {
			return nil, fmt.Errorf("scene.chart_pie_requires_single_series")
		}
		if n.Kind == "pie" {
			typ = pptx.ChartTypePie
		} else {
			typ = pptx.ChartTypeDoughnut
			o.HoleSize = sceneChartFloat(60)
		}
		o.DataLabelPosition = "outEnd"
		o.ShowLabel = ptrSceneBool(true)
		o.ShowLeaderLines = ptrSceneBool(true)
	}
	if len(n.Categories) == 0 || len(n.Categories) > 12 || len(n.Series) == 0 || len(n.Series) > 4 {
		return nil, fmt.Errorf("scene.chart_category_or_series_count")
	}
	if n.Kind == "column" && len(n.Categories) > 8 {
		return nil, fmt.Errorf("scene.chart_column_category_count")
	}
	categorySet := map[string]bool{}
	for i, label := range n.Categories {
		if strings.TrimSpace(label) == "" || categorySet[label] {
			return nil, fmt.Errorf("scene.chart_duplicate_or_empty_category")
		}
		categorySet[label] = true
		if _, err := sceneDataKey(ctx, "categories", i); err != nil {
			return nil, err
		}
	}
	data := make([]pptx.ChartData, len(n.Series))
	colors := make([]string, len(n.Series))
	max := 0.
	nameSet := map[string]bool{}
	for i, s := range n.Series {
		if s.Name == "" || nameSet[s.Name] || len(s.Values) != len(n.Categories) {
			return nil, fmt.Errorf("scene.chart_series_shape: %s", s.Name)
		}
		nameSet[s.Name] = true
		if _, err := sceneDataKey(ctx, "series", i); err != nil {
			return nil, err
		}
		ref := s.Color
		if i < len(n.Colors) {
			ref = n.Colors[i]
		}
		if ref == "" {
			ref = fmt.Sprintf("series.%d", i+1)
		}
		colors[i], err = r.sceneColor(surface, ref)
		if err != nil {
			return nil, err
		}
		vals := make([]float64, len(s.Values))
		for j, v := range s.Values {
			if v == nil {
				return nil, fmt.Errorf("scene.chart_native_missing_value_not_supported: %s/%d", s.Name, j)
			}
			if math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 {
				return nil, fmt.Errorf("scene.chart_requires_finite_nonnegative_values")
			}
			vals[j] = *v
			max = math.Max(max, *v)
		}
		labels := append([]string(nil), n.Categories...)
		if n.Kind == "column" || n.Kind == "line" {
			for j := range labels {
				labels[j] = strings.ToUpper(labels[j])
			}
		}
		data[i] = pptx.ChartData{DataIndex: i, Name: s.Name, Labels: [][]string{labels}, Values: vals}
	}
	if n.Format == nil && n.ValueSuffix == "" {
		integerData := true
		for _, series := range data {
			for _, value := range series.Values {
				integerData = integerData && value == math.Trunc(value)
			}
		}
		if integerData {
			o.DataLabelFormatCode = "#,##0"
			if n.Mode == "stacked" {
				o.DataLabelFormatCode = "#,##0;-#,##0;"
				p.Warnings = append(p.Warnings, "user-review.v1/stacked-zero-labels-hidden: "+id+" omits native zero-segment labels; external total values retained.")
			}
		}
	}
	if len(n.Colors) > len(n.Series) && (n.Kind != "pie" && n.Kind != "doughnut") {
		return nil, fmt.Errorf("scene.chart_extra_series_colors")
	}
	o.ChartColors = colors
	o.ValAxisMinVal = sceneChartFloat(0)
	o.ValAxisMaxVal = sceneChartFloat(sceneChartNiceMax(max))
	o.ValAxisHidden = ptrSceneBool(true)
	if n.Kind == "pie" || n.Kind == "doughnut" {
		sum := 0.
		for _, v := range data[0].Values {
			sum += v
		}
		if sum <= 0 {
			return nil, fmt.Errorf("scene.chart_pie_zero_total")
		}
		o.ChartColors = nil
		for i := range data[0].Values {
			ref := fmt.Sprintf("series.%d", i%9+1)
			if i < len(n.Colors) {
				ref = n.Colors[i]
			}
			color, err := r.sceneColor(surface, ref)
			if err != nil {
				return nil, err
			}
			o.ChartColors = append(o.ChartColors, color)
		}
	}
	if n.Kind == "column" || n.Kind == "bar" {
		o.BarDir = "col"
		if n.Kind == "bar" {
			o.BarDir = "bar"
			if len(data) != 1 {
				return nil, fmt.Errorf("scene.chart_bar_requires_single_series")
			}
		}
		o.BarGrouping = "clustered"
		o.BarGapWidthPct = sceneChartFloat(56.25 * float64(len(data)))
		switch n.Mode {
		case "", "clustered":
		case "stacked", "percent":
			o.BarGrouping = "stacked"
			if n.Mode == "percent" {
				o.BarGrouping = "percentStacked"
				o.ValAxisMaxVal = sceneChartFloat(1)
			} else {
				totalMax := 0.
				for j := range n.Categories {
					total := 0.
					for i := range data {
						total += data[i].Values[j]
					}
					totalMax = math.Max(totalMax, total)
				}
				o.ValAxisMaxVal = sceneChartFloat(sceneChartNiceMax(totalMax))
			}
			o.BarOverlapPct = 100
			o.BarGapWidthPct = sceneChartFloat(56.25)
			o.DataLabelFontSize = 9
			o.DataLabelPosition = "ctr"
			o.Layout = sceneChartLayout(0, 16/plot.H, (plot.W-96)/plot.W, (plot.H-34)/plot.H)
		default:
			return nil, fmt.Errorf("scene.chart_mode_unsupported")
		}
		if o.Layout == nil {
			o.Layout = sceneChartLayout(0, 16/plot.H, 1, (plot.H-34)/plot.H)
		}
		if n.Mode == "clustered" && len(data) > 1 {
			o.ShowLegend = ptrSceneBool(true)
			o.Layout = sceneChartLayout(0, 37/plot.H, 1, (plot.H-55)/plot.H)
		}
		if len(n.Highlight) > 0 {
			if len(data) != 1 || n.Mode == "stacked" || n.Mode == "percent" {
				return nil, fmt.Errorf("scene.chart_highlight_requires_single_series")
			}
			marked := map[int]bool{}
			for _, i := range n.Highlight {
				if i < 0 || i >= len(n.Categories) || marked[i] {
					return nil, fmt.Errorf("scene.chart_invalid_highlight")
				}
				marked[i] = true
			}
			o.ChartColors = nil
			for j := range n.Categories {
				ref := "deemph.2"
				if marked[j] {
					ref = "series.1"
				}
				color, _ := r.sceneColor(surface, ref)
				o.ChartColors = append(o.ChartColors, color)
			}
		}
	}
	if n.Kind == "line" {
		typ = pptx.ChartTypeLine
		o.Type = typ
		o.DataLabelPosition = "r"
		o.ShowValue = ptrSceneBool(false)
		o.ValAxisHidden = ptrSceneBool(false)
		o.ValGridLine.Style = "solid"
		o.ValAxisMajorUnit = sceneChartFloat((*o.ValAxisMaxVal) / 4)
		if n.YMin != nil {
			o.ValAxisMinVal = n.YMin
			if *n.YMin >= *o.ValAxisMaxVal {
				return nil, fmt.Errorf("scene.chart_invalid_y_min")
			}
			o.ValAxisMajorUnit = sceneChartFloat((*o.ValAxisMaxVal - *n.YMin) / 4)
		}
		o.LineSize = sceneChartFloat(2.25)
		o.LineDataSymbol = "square"
		o.LineDataSymbolSize = 7
		o.LineDataSymbolLineColor = "070154"
		o.LineDataSymbolLineSize = 1
		o.LineSmooth = ptrSceneBool(false)
		o.Layout = sceneChartLayout(30/plot.W, 0, (plot.W-138)/plot.W, (plot.H-18)/plot.H)
		for i, s := range n.Series {
			so := o
			so.MultiTypes = nil
			so.ChartColors = []string{colors[i]}
			if s.Dashed {
				so.LineDash = "dash"
			}
			o.MultiTypes = append(o.MultiTypes, pptx.IChartMulti{Type: typ, Data: []pptx.ChartData{data[i]}, Options: &so})
		}
		if n.Target != nil {
			if n.Target.Value < *o.ValAxisMinVal || n.Target.Value > *o.ValAxisMaxVal || n.Target.Label == "" {
				return nil, fmt.Errorf("scene.chart_invalid_target")
			}
			td := pptx.ChartData{DataIndex: len(data), Name: n.Target.Label, Labels: data[0].Labels, Values: make([]float64, len(n.Categories))}
			for i := range td.Values {
				td.Values[i] = n.Target.Value
			}
			to := o
			to.MultiTypes = nil
			to.ChartColors = []string{"070154"}
			to.LineDash = "dot"
			to.LineSize = sceneChartFloat(1)
			to.LineDataSymbol = "none"
			o.MultiTypes = append(o.MultiTypes, pptx.IChartMulti{Type: typ, Data: []pptx.ChartData{td}, Options: &to})
			data = append(data, td)
		}
	}
	chart := &sceneChart{ID: id + ".native", Rect: plot, Type: typ, Data: data, Options: o}
	p.Items = append(p.Items, sceneItem{Chart: chart})
	if n.Kind == "line" {
		if err := r.sceneLineEndLabels(p, id, n, plot, o, surface); err != nil {
			return nil, err
		}
	}
	if n.Mode == "stacked" || n.Mode == "percent" {
		if err := r.sceneStackedAnnotations(p, id, n, plot, o, surface); err != nil {
			return nil, err
		}
	}
	if n.Center != nil {
		if typ != pptx.ChartTypeDoughnut {
			return nil, fmt.Errorf("scene.chart_center_requires_doughnut")
		}
		value, err := sceneDataValue(*n.Center)
		if err != nil {
			return nil, err
		}
		st, _ := r.sceneStyle("number")
		if _, err := r.sceneDataText(p, id+".center.value", value, st, Rect{plot.X + plot.W*.3, plot.Y + plot.H/2 - 12, plot.W * .4, 0}, surface, "display", "center"); err != nil {
			return nil, err
		}
		st, _ = r.sceneStyle("small")
		if _, err := r.sceneDataText(p, id+".center.label", n.Center.Label, st, Rect{plot.X + plot.W*.3, plot.Y + plot.H/2 + 15, plot.W * .4, 0}, surface, "secondary", "center"); err != nil {
			return nil, err
		}
	}
	sceneDataGroup(p, id, "chart.native.source", 0, p.Bounds)
	p.Warnings = append(p.Warnings, "Native chart/workbook data is editable. Exact plot, native labels and source annotations require PowerPoint review; edit authored data through binding and rerender to synchronize external direct labels/totals.")
	return p, nil
}

func sceneChartDisplay(n sceneChartSource, value float64) (string, error) {
	if n.Format != nil {
		f := *n.Format
		f.Value = json.RawMessage(strconv.FormatFloat(value, 'g', -1, 64))
		return FormatNumber(f)
	}
	text := strconv.FormatFloat(value, 'f', 1, 64)
	text = strings.TrimSuffix(text, ".0")
	suffix := n.ValueSuffix
	if value == 1 && strings.HasSuffix(suffix, "s") {
		suffix = strings.TrimSuffix(suffix, "s")
	}
	return text + suffix, nil
}

func (r *renderer) sceneLineEndLabels(p *scenePlan, id string, n sceneChartSource, b Rect, o pptx.ChartOptions, surface string) error {
	type annotation struct {
		y           float64
		name, value string
	}
	var list []annotation
	min, max := *o.ValAxisMinVal, *o.ValAxisMaxVal
	for _, s := range n.Series {
		last := len(s.Values) - 1
		for last >= 0 && s.Values[last] == nil {
			last--
		}
		if last < 0 {
			continue
		}
		text, err := sceneChartDisplay(n, *s.Values[last])
		if err != nil {
			return err
		}
		list = append(list, annotation{b.Y + (b.H-18)*(1-(*s.Values[last]-min)/(max-min)) - 8, s.Name + " · ", text})
	}
	if n.Target != nil {
		list = append(list, annotation{b.Y + (b.H-18)*(1-(n.Target.Value-min)/(max-min)) - 8, n.Target.Label, ""})
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].y < list[j].y })
	for i := 1; i < len(list); i++ {
		list[i].y = math.Max(list[i].y, list[i-1].y+14)
	}
	st, _ := r.sceneDataToken("small", 600)
	st.ID = "source.chart-direct.11.14"
	st.Size = 11
	st.Leading = 14
	for i, it := range list {
		if _, err := r.sceneDataText(p, fmt.Sprintf("%s.direct-%d", id, i+1), it.name+it.value, st, Rect{b.X + b.W - 96, it.y, 96, 0}, surface, "primary", "left"); err != nil {
			return err
		}
	}
	return nil
}
func (r *renderer) sceneStackedAnnotations(p *scenePlan, id string, n sceneChartSource, b Rect, o pptx.ChartOptions, surface string) error {
	pw := b.W - 96
	slot := pw / float64(len(n.Categories))
	max := *o.ValAxisMaxVal
	st, _ := r.sceneStyle("number")
	st.ID = "source.chart-total.10.12"
	st.Size = 10
	st.Leading = 12
	for j := range n.Categories {
		total := 0.
		for _, s := range n.Series {
			total += *s.Values[j]
		}
		height := total / max * (b.H - 34)
		if n.Mode == "percent" {
			height = b.H - 34
		}
		text, err := sceneChartDisplay(n, total)
		if err != nil {
			return err
		}
		if _, err := r.sceneDataText(p, fmt.Sprintf("%s.total-%d", id, j+1), text, st, Rect{b.X + float64(j)*slot, b.Y + b.H - 18 - height - 16, slot, 0}, surface, "display", "center"); err != nil {
			return err
		}
	}
	last := len(n.Categories) - 1
	total := 0.
	for _, s := range n.Series {
		total += *s.Values[last]
	}
	if total == 0 {
		return nil
	}
	type anno struct {
		y    float64
		name string
	}
	var list []anno
	acc := 0.
	for _, s := range n.Series {
		v := *s.Values[last]
		if n.Mode == "percent" {
			v /= total
		}
		height := v / max * (b.H - 34)
		list = append(list, anno{b.Y + b.H - 18 - acc/max*(b.H-34) - height/2 - 9, s.Name})
		acc += v
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].y > list[j].y })
	for i := 1; i < len(list); i++ {
		list[i].y = math.Min(list[i].y, list[i-1].y-14)
	}
	label, _ := r.sceneStyle("small")
	for i, it := range list {
		if _, err := r.sceneDataText(p, fmt.Sprintf("%s.series-label-%d", id, i+1), it.name, label, Rect{b.X + pw - 3, it.y, 99, 0}, surface, "primary", "left"); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) sceneScatterChart(p *scenePlan, id string, n sceneChartSource, b Rect, surface string, ctx SceneContext, o pptx.ChartOptions) (*scenePlan, error) {
	if len(n.Points) < 1 || len(n.Points) > 60 || len(n.Series) > 0 || len(n.Categories) > 0 {
		return nil, fmt.Errorf("scene.chart_scatter_content_union")
	}
	xs, ys := []float64{}, []float64{}
	labels := []string{}
	maxX, maxY := 0., 0.
	for i, point := range n.Points {
		if len(point) < 2 || len(point) > 3 {
			return nil, fmt.Errorf("scene.chart_scatter_point_requires_pair_or_label")
		}
		var x, y float64
		if err := json.Unmarshal(point[0], &x); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(point[1], &y); err != nil {
			return nil, err
		}
		if x < 0 || y < 0 {
			return nil, fmt.Errorf("scene.chart_scatter_positive_domain")
		}
		xs = append(xs, x)
		ys = append(ys, y)
		maxX, maxY = math.Max(maxX, x), math.Max(maxY, y)
		label := n.Labels[strconv.Itoa(i)]
		if len(point) == 3 {
			if err := json.Unmarshal(point[2], &label); err != nil {
				return nil, err
			}
		}
		labels = append(labels, label)
		if _, err := sceneDataKey(ctx, "points", i); err != nil {
			return nil, err
		}
	}
	o.ChartColors = []string{"0047FF"}
	o.LineSize = sceneChartFloat(0)
	o.LineDataSymbol = "circle"
	o.LineDataSymbolSize = 8
	o.CatAxisMinVal = sceneChartFloat(0)
	o.CatAxisMaxVal = sceneChartFloat(sceneChartNiceMax(maxX))
	o.ValAxisMinVal = sceneChartFloat(0)
	o.ValAxisMaxVal = sceneChartFloat(sceneChartNiceMax(maxY))
	o.CatAxisMajorUnit = sceneChartFloat(*o.CatAxisMaxVal / 4)
	o.ValAxisMajorUnit = sceneChartFloat(*o.ValAxisMaxVal / 4)
	o.ValAxisHidden = ptrSceneBool(false)
	o.CatGridLine.Style = "solid"
	o.ValGridLine.Style = "solid"
	o.DataLabelFormatScatter = "custom"
	o.ShowLabel = ptrSceneBool(true)
	o.ShowValue = ptrSceneBool(false)
	o.DataLabelFontFace = o.FontFace
	o.DataLabelFontSize = 11
	o.DataLabelPosition = "r"
	o.Layout = sceneChartLayout(36/b.W, 0, (b.W-48)/b.W, (b.H-30)/b.H)
	data := []pptx.ChartData{{Name: n.XTitle, Values: xs}, {Name: n.YTitle, Values: ys, Labels: [][]string{labels}}}
	if n.Trend {
		p.Warnings = append(p.Warnings, "Native scatter trend line support requires an explicit writer trendline contract.")
		return nil, fmt.Errorf("scene.chart_native_trendline_not_supported")
	}
	p.Items = append(p.Items, sceneItem{Chart: &sceneChart{ID: id + ".native", Rect: b, Type: pptx.ChartTypeScatter, Data: data, Options: o}})
	sceneDataGroup(p, id, "chart.scatter.native", 0, p.Bounds)
	return p, nil
}

func (r *renderer) sceneChartPolyline(p *scenePlan, id string, points [][2]float64, color string, width float64, dash string, arrow bool) {
	minX, minY, maxX, maxY := points[0][0], points[0][1], points[0][0], points[0][1]
	for _, v := range points {
		minX, minY, maxX, maxY = math.Min(minX, v[0]), math.Min(minY, v[1]), math.Max(maxX, v[0]), math.Max(maxY, v[1])
	}
	b := Rect{minX, minY, math.Max(maxX-minX, .01), math.Max(maxY-minY, .01)}
	line := &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color}, Width: width, DashType: dash}
	if arrow {
		line.EndArrowType = "arrow"
	}
	sh := sceneShape{Type: pptx.ShapeTypeCustGeom, Props: pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: line}, Record: ShapeRecord{ID: id, Rect: b, Color: color, Geometry: "custGeom"}}
	for i, v := range points {
		sh.Props.Points = append(sh.Props.Points, pptx.ShapePoint{X: pptx.Inches((v[0] - minX) / 72), Y: pptx.Inches((v[1] - minY) / 72), MoveTo: ptrSceneBool(i == 0)})
	}
	p.Items = append(p.Items, sceneItem{Shape: &sh})
	sceneDataBounds(p, b)
}

func (r *renderer) sceneQuadrantChart(p *scenePlan, id string, n sceneChartSource, b Rect, surface string, ctx SceneContext) error {
	if len(n.Items) < 1 || len(n.Items) > 12 || len(n.Quadrants) != 4 {
		return fmt.Errorf("scene.quadrant_requires_four_quadrants_and_1_to_12_items")
	}
	if n.Style != "" && n.Style != "strong" && n.Style != "subtle" {
		return fmt.Errorf("scene.quadrant_style_enum")
	}
	if n.PositionMode != "" && n.PositionMode != "qualitative" && n.PositionMode != "data" {
		return fmt.Errorf("scene.quadrant_position_mode_enum")
	}
	if len(n.Series) > 0 || len(n.Points) > 0 || n.Format != nil {
		return fmt.Errorf("scene.quadrant_content_union")
	}
	plotW, plotH := b.W-36, b.H-24
	if n.Key.ownLegend() {
		plotW = b.W - 330
	}
	if r.source.Revision == LibraryRevisionV2 {
		availableW := b.W - 36
		if n.Key.ownLegend() {
			availableW = b.W - 300
		}
		plotW = math.Min(b.H-36, availableW)
		plotH = plotW
	}
	if plotW < 72 || plotH < 72 {
		return fmt.Errorf("scene.quadrant_minimum_plot")
	}
	x, y := b.X+24, b.Y
	midX, midY := plotW/2, plotH/2
	if r.source.Revision == LibraryRevisionV1 {
		p.Warnings = append(p.Warnings, "user-review.v1/quadrant-rectangular-plot: "+id+" uses allocated rectangular plot,24pt axis reserve and fixed306pt key reserve when enabled.")
	}
	fill := n.Fill
	if fill == "" {
		fill = "inverse"
	}
	strongState := n.StrongState
	if strongState == "" {
		strongState = "desirable"
		for _, q := range n.Quadrants {
			if q.State == "undesirable" {
				strongState = "undesirable"
			}
		}
	}
	if strongState != "desirable" && strongState != "undesirable" {
		return fmt.Errorf("scene.quadrant_strong_state_enum")
	}
	posq := map[string][2]float64{"tl": {0, 0}, "tr": {1, 0}, "bl": {0, 1}, "br": {1, 1}}
	surfaces := map[string]string{}
	for _, key := range []string{"tl", "tr", "bl", "br"} {
		q, ok := n.Quadrants[key]
		if !ok || q.Name == "" {
			return fmt.Errorf("scene.quadrant_missing_quadrant: %s", key)
		}
		if q.State != "" && q.State != "desirable" && q.State != "undesirable" {
			return fmt.Errorf("scene.quadrant_state_enum")
		}
		if q.State != "" && q.Tag == "" {
			return fmt.Errorf("scene.quadrant_marked_state_requires_tag")
		}
		on := "light"
		if n.Style == "strong" && q.State == strongState {
			on = fill
		} else if q.State == "desirable" {
			on = "subtle"
		}
		surfaces[key] = on
		qp := posq[key]
		box := Rect{x + qp[0]*midX, y + qp[1]*midY, midX, midY}
		if err := r.sceneRect(p, id+".quadrant."+key, box, on); err != nil {
			return err
		}
		if q.State == "undesirable" && on == "light" {
			hatch, _ := r.sceneColor("light", "line")
			for d := -midY; d < midX; d += 6 {
				ax, ay := box.X+math.Max(0, d), box.Y+math.Max(0, -d)
				bx, by := box.X+math.Min(midX, d+midY), box.Y+math.Min(midY, midX-d)
				r.sceneChartPolyline(p, fmt.Sprintf("%s.hatch.%s.%03d", id, key, int((d+midY)/6)), [][2]float64{{ax, ay}, {bx, by}}, hatch, 2, "solid", false)
			}
		}
		st, _ := r.sceneDataToken("label", 600)
		st.TrackingPt = .6
		ny := box.Y + 6
		if qp[1] == 1 {
			ny = box.Y + midY - 24
		}
		align := "left"
		if qp[0] == 1 {
			align = "right"
		}
		if _, err := r.sceneDataText(p, id+".name."+key, q.Name, st, Rect{box.X + 9, ny, midX - 18, 0}, on, "primary", align); err != nil {
			return err
		}
		if q.Tag != "" {
			ts := st
			ts.ID = "source.quadrant-tag.8.5.12"
			ts.Size = 8.5
			tl, err := r.typeEngine.Measure(q.Tag, ts, midX-18)
			if err != nil {
				return err
			}
			if len(tl.Lines) != 1 {
				return fmt.Errorf("scene.quadrant_tag_wrap")
			}
			tw := tl.Lines[0].Advance + 14
			tx := box.X + 9
			if qp[0] == 1 {
				tx = box.X + midX - 9 - tw
			}
			ty := ny + 18
			if qp[1] == 1 {
				ty = ny - 18
			}
			tagSurface := "inverse"
			if on == "inverse" {
				tagSurface = "light"
			} else if on == "callout" || q.State == "undesirable" {
				tagSurface = "inverse"
			} else {
				tagSurface = "deep"
			}
			if err := r.sceneRect(p, id+".tag-bg."+key, Rect{tx, ty, tw, 16}, tagSurface); err != nil {
				return err
			}
			if _, err := r.sceneDataText(p, id+".tag."+key, q.Tag, ts, Rect{tx + 6, ty + 1, tw - 12, 14}, tagSurface, "primary", "left"); err != nil {
				return err
			}
		}
	}
	navy, _ := r.sceneColor("light", "strong")
	for i, line := range [][][2]float64{{{x, y}, {x + plotW, y}, {x + plotW, y + plotH}, {x, y + plotH}, {x, y}}, {{x + midX, y}, {x + midX, y + plotH}}, {{x, y + midY}, {x + plotW, y + midY}}} {
		width := .75
		if i == 0 {
			width = 1
		}
		r.sceneChartPolyline(p, fmt.Sprintf("%s.axis-%d", id, i+1), line, navy, width, "solid", false)
	}
	axis, _ := r.sceneDataToken("label", 600)
	axis.TrackingPt = .6
	for _, a := range []struct {
		id, text, align string
		b               Rect
	}{{"x.low", "Low", "left", Rect{x, y + plotH + 2, 36, 14}}, {"x.high", "High", "right", Rect{x + plotW - 36, y + plotH + 2, 36, 14}}, {"x.title", n.XTitle + " →", "center", Rect{x + 36, y + plotH + 2, plotW - 72, 14}}} {
		if _, err := r.sceneDataText(p, id+"."+a.id, a.text, axis, a.b, surface, "primary", a.align); err != nil {
			return err
		}
	}
	ys, _ := r.sceneStyle("label")
	for i, a := range []struct {
		text string
		cy   float64
	}{{"High", y + 18}, {"Low", y + plotH - 18}, {n.YTitle + " →", y + midY}} {
		l, err := r.typeEngine.Measure(a.text, ys, plotH)
		if err != nil {
			return err
		}
		w := l.Lines[0].Advance + 1
		tr, err := r.sceneDataText(p, fmt.Sprintf("%s.y-%d", id, i), a.text, ys, Rect{x - 10 - w/2, a.cy - 7, w, 14}, surface, "secondary", "center")
		if err != nil {
			return err
		}
		p.Items[len(p.Items)-1].Text.Rotation = -90
		_ = tr
	}
	valuesX, valuesY, labels := []float64{}, []float64{}, []string{}
	for i, it := range n.Items {
		key, err := sceneDataKey(ctx, "items", i)
		if err != nil {
			return err
		}
		if it.X < 0 || it.X > 1 || it.Y < 0 || it.Y > 1 || !n.Key.markersOnly() && it.Label == "" || math.IsNaN(it.X+it.Y) || math.IsInf(it.X+it.Y, 0) {
			return fmt.Errorf("scene.quadrant_item_domain")
		}
		iid := id + ".item." + key
		px, py := x+it.X*plotW, y+(1-it.Y)*plotH
		qkey := "b"
		if it.Y >= .5 {
			qkey = "t"
		}
		if it.X >= .5 {
			qkey += "r"
		} else {
			qkey += "l"
		}
		on := surfaces[qkey]
		if len(it.To) > 0 {
			if len(it.To) != 2 || it.To[0] < 0 || it.To[0] > 1 || it.To[1] < 0 || it.To[1] > 1 {
				return fmt.Errorf("scene.quadrant_destination_domain")
			}
			tx, ty := x+it.To[0]*plotW, y+(1-it.To[1])*plotH
			dx, dy := tx-px, ty-py
			length := math.Hypot(dx, dy)
			if length <= 20 {
				return fmt.Errorf("scene.quadrant_arrow_too_short")
			}
			ux, uy := dx/length, dy/length
			points := [][2]float64{{px + ux*10, py + uy*10}, {tx - ux*10, ty - uy*10}}
			r.sceneChartPolyline(p, iid+".arrow-underlay", points, "FFFFFF", 3.5, "solid", true)
			r.sceneChartPolyline(p, iid+".arrow", points, navy, 1.25, "dash", true)
		}
		valuesX = append(valuesX, it.X)
		valuesY = append(valuesY, it.Y)
		labels = append(labels, it.Label)
		if n.PositionMode == "data" {
			continue
		}
		fg, ink := navy, "primary"
		markSurface := "inverse"
		if on == "inverse" {
			fg = "FFFFFF"
			markSurface = "light"
		}
		radius := 5.
		if n.Key.numbered() {
			radius = 8
		}
		r.sceneDataShape(p, iid+".point", Rect{px - radius, py - radius, 2 * radius, 2 * radius}, pptx.ShapeTypeEllipse, fg, nil)
		if n.Key.numbered() {
			st, _ := r.sceneDataToken("label", 600)
			if _, err := r.sceneDataText(p, iid+".number", strconv.Itoa(i+1), st, Rect{px - 7, py - 7, 14, 14}, markSurface, "primary", "center"); err != nil {
				return err
			}
			if n.Key.markersOnly() {
				continue
			}
			keyGap := 24.
			if r.source.Revision == LibraryRevisionV2 {
				keyGap = 36
			}
			kx, ky := x+plotW+keyGap, y+6+float64(i)*24
			r.sceneDataShape(p, iid+".key-point", Rect{kx, ky, 16, 16}, pptx.ShapeTypeEllipse, navy, nil)
			if _, err := r.sceneDataText(p, iid+".key-number", strconv.Itoa(i+1), st, Rect{kx + 1, ky + 1, 14, 14}, "inverse", "primary", "center"); err != nil {
				return err
			}
			st, _ = r.sceneStyle("small")
			if _, err := r.sceneDataText(p, iid+".key-label", it.Label, st, Rect{kx + 24, ky - 1, b.X + b.W - kx - 24, 0}, surface, ink, "left"); err != nil {
				return err
			}
		} else {
			st, _ := r.sceneDataToken("small", 600)
			st.ID = "source.quadrant-item.11.14"
			st.Size = 11
			st.Leading = 14
			if _, err := r.sceneDataText(p, iid+".label", it.Label, st, Rect{px + 9, py - 8, math.Max(36, b.X+b.W-px-9), 0}, on, "primary", "left"); err != nil {
				return err
			}
		}
	}
	if n.PositionMode == "data" {
		if n.Key.numbered() {
			return fmt.Errorf("scene.quadrant_native_data_key_not_supported")
		}
		o, err := r.sceneChartOptions(surface)
		if err != nil {
			return err
		}
		o.ChartColors = []string{navy}
		o.LineSize = sceneChartFloat(0)
		o.LineDataSymbol = "circle"
		o.LineDataSymbolSize = 10
		o.ShowLabel = ptrSceneBool(true)
		o.ShowValue = ptrSceneBool(false)
		o.DataLabelPosition = "r"
		o.DataLabelFormatScatter = "custom"
		o.CatAxisHidden = ptrSceneBool(true)
		o.ValAxisHidden = ptrSceneBool(true)
		o.CatAxisMinVal = sceneChartFloat(0)
		o.CatAxisMaxVal = sceneChartFloat(1)
		o.ValAxisMinVal = sceneChartFloat(0)
		o.ValAxisMaxVal = sceneChartFloat(1)
		o.Layout = sceneChartLayout(0, 0, 1, 1)
		p.Items = append(p.Items, sceneItem{Chart: &sceneChart{ID: id + ".native", Rect: Rect{x, y, plotW, plotH}, Type: pptx.ChartTypeScatter, Data: []pptx.ChartData{{Name: n.XTitle, Values: valuesX}, {Name: n.YTitle, Values: valuesY, Labels: [][]string{labels}}}, Options: o}})
	} else {
		p.Warnings = append(p.Warnings, "Quadrant positions are authored qualitative normalized coordinates, retained as editable native shapes; no numerical scatter/workbook claim.")
	}
	return nil
}
