// objects.go ports src/gen-objects.ts: the slide-object generators that
// validate/default user input and build the internal presentation model
// (PresSlide.SlideObjects plus the Rels/RelsChart/RelsMedia relationship
// lists). These functions mutate the target slide in place. TS `throw`
// becomes a returned error; TS `console.warn/error` (non-fatal) becomes a
// silent default, matching the "libraries should not log" convention.
package pptx

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// _chartCounter is the global counter for included charts (used as the index
// in their filenames). Mirrors the module-level `let _chartCounter` in TS.
var _chartCounter = 0

// imageDataMimeRe extracts the subtype from an `image/<subtype>;` data URI.
var imageDataMimeRe = regexp.MustCompile(`image/(\w+);`)

// ---------------------------------------------------------------------------
// small helpers
// ---------------------------------------------------------------------------

func orStr(v, def string) string {
	if v != "" {
		return v
	}
	return def
}

func orF(v, def float64) float64 {
	if v != 0 {
		return v
	}
	return def
}

func strInSet(v string, set ...string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}

// countSlideObjectsOfType counts existing slide objects of the given type,
// used to build sequential default object names ("Text 0", "Shape 1", ...).
func countSlideObjectsOfType(target *PresSlide, t SlideObjectType) int {
	n := 0
	for i := range target.SlideObjects {
		if target.SlideObjects[i].Type == t {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// option-struct converters (TS relies on structural typing; Go needs explicit
// copies between the option surfaces that gen-objects treats interchangeably)
// ---------------------------------------------------------------------------

func objectOptionsFromShapeProps(s *ShapeProps) *ObjectOptions {
	o := &ObjectOptions{}
	if s == nil {
		return o
	}
	o.PositionProps = s.PositionProps
	o.ObjectNameProps = s.ObjectNameProps
	o.Align = s.Align
	o.AngleRange = s.AngleRange
	o.ArcThicknessRatio = s.ArcThicknessRatio
	o.Fill = s.Fill
	o.FlipH = s.FlipH
	o.FlipV = s.FlipV
	o.Hyperlink = s.Hyperlink
	o.Line = s.Line
	o.Points = s.Points
	o.RectRadius = s.RectRadius
	o.Rotate = s.Rotate
	o.Shadow = s.Shadow
	o.LineSize = s.LineSize
	o.LineDash = s.LineDash
	o.LineHead = s.LineHead
	o.LineTail = s.LineTail
	o.ShapeName = s.ShapeName
	return o
}

func objectOptionsFromPlaceholderProps(p *PlaceholderProps) *ObjectOptions {
	o := &ObjectOptions{}
	if p == nil {
		return o
	}
	o.PositionProps = p.PositionProps
	o.TextBaseProps = p.TextBaseProps
	o.Margin = p.Margin
	return o
}

// objectOptionsFromTextPropsOptions widens a per-run TextPropsOptions into the
// flattened ObjectOptions bag used by cleanOpts / SlideObject.Options.
func objectOptionsFromTextPropsOptions(t *TextPropsOptions) *ObjectOptions {
	o := &ObjectOptions{}
	if t == nil {
		return o
	}
	o.PositionProps = t.PositionProps
	o.DataOrPathProps = t.DataOrPathProps
	o.TextBaseProps = t.TextBaseProps
	o.ObjectNameProps = t.ObjectNameProps
	o.BodyProp = t.BodyProp
	o.LineIdx = t.LineIdx
	o.Baseline = t.Baseline
	o.CharSpacing = t.CharSpacing
	o.Fit = t.Fit
	o.Fill = t.Fill
	o.FlipH = t.FlipH
	o.FlipV = t.FlipV
	o.Glow = t.Glow
	o.Hyperlink = t.Hyperlink
	o.IndentLevel = t.IndentLevel
	o.IsTextBox = t.IsTextBox
	o.Line = t.Line
	o.LineSpacing = t.LineSpacing
	o.LineSpacingMultiple = t.LineSpacingMultiple
	o.Margin = t.Margin
	o.Outline = t.Outline
	o.ParaSpaceAfter = t.ParaSpaceAfter
	o.ParaSpaceBefore = t.ParaSpaceBefore
	o.Placeholder = t.Placeholder
	o.RectRadius = t.RectRadius
	o.Rotate = t.Rotate
	o.RtlMode = t.RtlMode
	o.Shadow = t.Shadow
	o.Shape = t.Shape
	o.Strike = t.Strike
	o.Subscript = t.Subscript
	o.Superscript = t.Superscript
	o.Vert = t.Vert
	o.Wrap = t.Wrap
	o.AutoFit = t.AutoFit
	o.ShrinkText = t.ShrinkText
	o.Inset = t.Inset
	o.LineSize = t.LineSize
	o.LineDash = t.LineDash
	o.LineHead = t.LineHead
	o.LineTail = t.LineTail
	// TextPropsOptions carries a direct Valign field that shadows the embedded
	// TextBaseProps.Valign; prefer the direct value when set.
	if t.Valign != "" {
		o.Valign = t.Valign
	}
	return o
}

// textPropsOptionsFromObjectOptions narrows a cleaned ObjectOptions back into a
// per-run TextPropsOptions. Fields unique to ObjectOptions (placeholder idx,
// image/table geometry) are irrelevant to a text run and dropped.
func textPropsOptionsFromObjectOptions(o *ObjectOptions) *TextPropsOptions {
	t := &TextPropsOptions{}
	if o == nil {
		return t
	}
	t.PositionProps = o.PositionProps
	t.DataOrPathProps = o.DataOrPathProps
	t.TextBaseProps = o.TextBaseProps
	t.ObjectNameProps = o.ObjectNameProps
	t.BodyProp = o.BodyProp
	t.LineIdx = o.LineIdx
	t.Baseline = o.Baseline
	t.CharSpacing = o.CharSpacing
	t.Fit = o.Fit
	t.Fill = o.Fill
	t.FlipH = o.FlipH
	t.FlipV = o.FlipV
	t.Glow = o.Glow
	t.Hyperlink = o.Hyperlink
	t.IndentLevel = o.IndentLevel
	t.IsTextBox = o.IsTextBox
	t.Line = o.Line
	t.LineSpacing = o.LineSpacing
	t.LineSpacingMultiple = o.LineSpacingMultiple
	t.Margin = o.Margin
	t.Outline = o.Outline
	t.ParaSpaceAfter = o.ParaSpaceAfter
	t.ParaSpaceBefore = o.ParaSpaceBefore
	t.Placeholder = o.Placeholder
	t.RectRadius = o.RectRadius
	t.Rotate = o.Rotate
	t.RtlMode = o.RtlMode
	t.Shadow = o.Shadow
	t.Shape = o.Shape
	t.Strike = o.Strike
	t.Subscript = o.Subscript
	t.Superscript = o.Superscript
	t.Vert = o.Vert
	t.Wrap = o.Wrap
	t.AutoFit = o.AutoFit
	t.ShrinkText = o.ShrinkText
	t.Inset = o.Inset
	t.LineSize = o.LineSize
	t.LineDash = o.LineDash
	t.LineHead = o.LineHead
	t.LineTail = o.LineTail
	t.Valign = o.Valign
	return t
}

// tablePropsToObjectOptions copies the table-level options needed by the stored
// SlideObject.Options for a (non-auto-paged) table.
func tablePropsToObjectOptions(t *TableProps) *ObjectOptions {
	o := &ObjectOptions{}
	if t == nil {
		return o
	}
	o.PositionProps = t.PositionProps
	o.TextBaseProps = t.TextBaseProps
	o.ObjectNameProps = t.ObjectNameProps
	o.Margin = t.Margin
	o.Border = t.Border
	o.ColW = t.ColW
	o.RowH = t.RowH
	o.Fill = t.Fill
	o.AutoPageCharWeight = t.AutoPageCharWeight
	o.AutoPageLineWeight = t.AutoPageLineWeight
	return o
}

// mergePlaceholderOptions copies the set (non-zero) placeholder options onto
// dst, approximating the TS `{ ...itemOpts, ...placeHold.options }` spread
// (which only copies the placeholder's explicitly-defined properties).
func mergePlaceholderOptions(dst, src *ObjectOptions) {
	if src == nil {
		return
	}
	if src.Align != "" {
		dst.Align = src.Align
	}
	if src.Bold != nil {
		dst.Bold = src.Bold
	}
	if src.Bullet != nil {
		dst.Bullet = src.Bullet
	}
	if src.Color != "" {
		dst.Color = src.Color
	}
	if src.FontFace != "" {
		dst.FontFace = src.FontFace
	}
	if src.FontSize != 0 {
		dst.FontSize = src.FontSize
	}
	if src.Italic != nil {
		dst.Italic = src.Italic
	}
	if src.Underline != nil {
		dst.Underline = src.Underline
	}
	if src.Valign != "" {
		dst.Valign = src.Valign
	}
	if src.Margin != nil {
		dst.Margin = src.Margin
	}
	if src.Fill != nil {
		dst.Fill = src.Fill
	}
	if src.Line != nil {
		dst.Line = src.Line
	}
	if src.Shadow != nil {
		dst.Shadow = src.Shadow
	}
	if src.BodyProp != nil {
		dst.BodyProp = src.BodyProp
	}
	if src.IndentLevel != 0 {
		dst.IndentLevel = src.IndentLevel
	}
	if src.ParaSpaceBefore != 0 {
		dst.ParaSpaceBefore = src.ParaSpaceBefore
	}
	if src.ParaSpaceAfter != 0 {
		dst.ParaSpaceAfter = src.ParaSpaceAfter
	}
	if src.LineSpacing != 0 {
		dst.LineSpacing = src.LineSpacing
	}
	if src.LineSpacingMultiple != 0 {
		dst.LineSpacingMultiple = src.LineSpacingMultiple
	}
	if src.RtlMode != nil {
		dst.RtlMode = src.RtlMode
	}
}

// ---------------------------------------------------------------------------
// createSlideMaster
// ---------------------------------------------------------------------------

// createSlideMaster transforms a slide-master definition into a SlideLayout by
// registering its background/objects/slide-number/placeholders. Ports the TS
// `createSlideMaster`. The layout's SlideBaseProps is bridged into a temporary
// PresSlide because the add* helpers operate on PresSlide (TS casts freely).
func createSlideMaster(props *SlideMasterProps, target *SlideLayout) error {
	// STEP 1: DEPRECATED bkgd passthrough (remove in v4.0.0)
	if props.Bkgd != nil {
		target.Bkgd = props.Bkgd
	}

	// Bridge: run the object generators against a PresSlide view of the layout.
	ps := &PresSlide{SlideBaseProps: target.SlideBaseProps}

	// STEP 2: Add all master objects in the order they were given.
	for idx := range props.Objects {
		object := props.Objects[idx]
		switch {
		case object.Chart != nil:
			if _, err := addChartDefinition(ps, object.Chart.Type, object.Chart.MultiTypes, nil, object.Chart); err != nil {
				return err
			}
		case object.Image != nil:
			if err := addImageDefinition(ps, object.Image); err != nil {
				return err
			}
		case object.Line != nil:
			if err := addShapeDefinition(ps, ShapeTypeLine, object.Line); err != nil {
				return err
			}
		case object.Rect != nil:
			if err := addShapeDefinition(ps, ShapeTypeRect, object.Rect); err != nil {
				return err
			}
		case object.Text != nil:
			opts := objectOptionsFromTextPropsOptions(object.Text.Options)
			addTextDefinition(ps, []TextProps{{Text: object.Text.Text}}, opts, false)
		case object.Placeholder != nil:
			ph := object.Placeholder
			opts := objectOptionsFromPlaceholderProps(&ph.Options)
			// remap name/type for internal handling (mirrors TS)
			opts.Placeholder = ph.Options.Name
			opts.PlaceholderType = ph.Options.Type
			opts.PlaceholderIdx = 100 + idx
			addTextDefinition(ps, []TextProps{{Text: ph.Text}}, opts, true)
		}
	}

	// Copy the mutated model back onto the layout.
	target.SlideBaseProps = ps.SlideBaseProps

	// STEP 3: Add slide numbers last so they are not covered by objects.
	if props.SlideNumber != nil {
		target.SlideNumberProps = props.SlideNumber
	}
	return nil
}

// ---------------------------------------------------------------------------
// addChartDefinition
// ---------------------------------------------------------------------------

// addChartDefinition validates chart input, applies the extensive option
// defaults, registers a chart relationship, and appends a chart slide object.
// Ports the TS `addChartDefinition`. TS uses console.warn (never throws) for
// invalid sub-options; here those are silently corrected, so the returned
// error is always nil (kept for signature symmetry / future validation).
func addChartDefinition(target *PresSlide, chartType ChartType, multiTypes []IChartMulti, data []ChartData, opt *ChartOptions) (*SlideObject, error) {
	correctGridLineOptions := func(gl *OptsChartGridLine) {
		if gl == nil || gl.Style == "none" {
			return
		}
		if gl.Size != 0 && gl.Size <= 0 {
			gl.Size = 0
		}
		if gl.Style != "" && !strInSet(gl.Style, "solid", "dash", "dot") {
			gl.Style = ""
		}
		if gl.Cap != "" && !strInSet(gl.Cap, "flat", "square", "round") {
			gl.Cap = ""
		}
	}

	_chartCounter++
	chartID := _chartCounter

	isMulti := len(multiTypes) > 0

	// DESIGN: multi-type charts concat each component's data.
	var tmpData []ChartData
	if isMulti {
		for i := range multiTypes {
			tmpData = append(tmpData, multiTypes[i].Data...)
		}
	} else {
		tmpData = data
	}
	for i := range tmpData {
		tmpData[i].DataIndex = i
		// NOTE: labels string[]->string[][] normalization is handled by the
		// ChartData type (Labels is already [][]string).
	}

	opts := opt
	if opts == nil {
		opts = &ChartOptions{}
	}

	// STEP 2: Core options.
	if isMulti {
		opts.MultiTypes = multiTypes
	} else {
		opts.Type = chartType
	}
	if opts.X == nil {
		opts.X = ptr(Inches(1))
	}
	if opts.Y == nil {
		opts.Y = ptr(Inches(1))
	}
	if opts.W == nil {
		opts.W = ptr(Percent(50))
	}
	if opts.H == nil {
		opts.H = ptr(Percent(50))
	}
	if opts.ObjectName != "" {
		opts.ObjectName = encodeXmlEntities(opts.ObjectName)
	} else {
		opts.ObjectName = fmt.Sprintf("Chart %d", countSlideObjectsOfType(target, SlideObjectTypeChart))
	}

	// B: misc
	if !strInSet(opts.BarDir, "bar", "col") {
		opts.BarDir = "col"
	}
	switch opts.Type {
	case ChartTypeArea:
		if !strInSet(opts.BarGrouping, "stacked", "standard", "percentStacked") {
			opts.BarGrouping = "standard"
		}
	case ChartTypeBar:
		if !strInSet(opts.BarGrouping, "clustered", "stacked", "percentStacked") {
			opts.BarGrouping = "clustered"
		}
	case ChartTypeBar3d:
		if !strInSet(opts.BarGrouping, "clustered", "stacked", "standard", "percentStacked") {
			opts.BarGrouping = "standard"
		}
	}
	if strings.Contains(opts.BarGrouping, "tacked") {
		if opts.BarGapWidthPct == 0 {
			opts.BarGapWidthPct = 50
		}
	}

	// data label position validation
	if opts.DataLabelPosition != "" {
		if opts.Type == ChartTypeArea || opts.Type == ChartTypeBar3d || opts.Type == ChartTypeDoughnut || opts.Type == ChartTypeRadar {
			opts.DataLabelPosition = ""
		}
		if opts.Type == ChartTypePie {
			if !strInSet(opts.DataLabelPosition, "bestFit", "ctr", "inEnd", "outEnd") {
				opts.DataLabelPosition = ""
			}
		}
		if opts.Type == ChartTypeBubble || opts.Type == ChartTypeBubble3d || opts.Type == ChartTypeLine || opts.Type == ChartTypeScatter {
			if !strInSet(opts.DataLabelPosition, "b", "ctr", "l", "r", "t") {
				opts.DataLabelPosition = ""
			}
		}
		if opts.Type == ChartTypeBar {
			if !strInSet(opts.BarGrouping, "stacked", "percentStacked") {
				if !strInSet(opts.DataLabelPosition, "ctr", "inBase", "inEnd") {
					opts.DataLabelPosition = ""
				}
			}
			if !strInSet(opts.BarGrouping, "clustered") {
				if !strInSet(opts.DataLabelPosition, "ctr", "inBase", "inEnd", "outEnd") {
					opts.DataLabelPosition = ""
				}
			}
		}
	}

	if !strInSet(opts.LegendPos, "b", "l", "r", "t", "tr") {
		opts.LegendPos = "r"
	}
	if !strInSet(opts.Bar3DShape, "cone", "coneToMax", "box", "cylinder", "pyramid", "pyramidToMax") {
		opts.Bar3DShape = "box"
	}
	if !strInSet(opts.LineDataSymbol, "circle", "dash", "diamond", "dot", "none", "square", "triangle") {
		opts.LineDataSymbol = "circle"
	}
	if !strInSet(opts.DisplayBlanksAs, "gap", "span") {
		opts.DisplayBlanksAs = "span"
	}
	if !strInSet(opts.RadarStyle, "standard", "marker", "filled") {
		opts.RadarStyle = "standard"
	}
	opts.LineDataSymbolSize = orF(opts.LineDataSymbolSize, 6)
	if opts.LineDataSymbolLineSize != 0 {
		opts.LineDataSymbolLineSize = float64(valToPts(opts.LineDataSymbolLineSize))
	} else {
		opts.LineDataSymbolLineSize = float64(valToPts(0.75))
	}

	// layout override (0..1 per key)
	if opts.Layout != nil {
		clampLayout := func(c **Coord) {
			if *c == nil {
				return
			}
			v := (*c).Val
			if v < 0 || v > 1 {
				*c = nil
			}
		}
		clampLayout(&opts.Layout.X)
		clampLayout(&opts.Layout.Y)
		clampLayout(&opts.Layout.W)
		clampLayout(&opts.Layout.H)
	}

	// gridline defaults
	if opts.CatGridLine == nil {
		if opts.Type == ChartTypeScatter {
			opts.CatGridLine = &OptsChartGridLine{Color: "D9D9D9", Size: 1}
		} else {
			opts.CatGridLine = &OptsChartGridLine{Style: "none"}
		}
	}
	if opts.ValGridLine == nil {
		if opts.Type == ChartTypeScatter {
			opts.ValGridLine = &OptsChartGridLine{Color: "D9D9D9", Size: 1}
		} else {
			opts.ValGridLine = &OptsChartGridLine{}
		}
	}
	if opts.SerGridLine == nil {
		if opts.Type == ChartTypeScatter {
			opts.SerGridLine = &OptsChartGridLine{Color: "D9D9D9", Size: 1}
		} else {
			opts.SerGridLine = &OptsChartGridLine{Style: "none"}
		}
	}
	correctGridLineOptions(opts.CatGridLine)
	correctGridLineOptions(opts.ValGridLine)
	correctGridLineOptions(opts.SerGridLine)
	correctShadowOptions(opts.Shadow)

	// C: plotArea axis line show (default true when unset)
	if opts.CatAxisLineShow == nil {
		opts.CatAxisLineShow = ptr(true)
	}
	if opts.ValAxisLineShow == nil {
		opts.ValAxisLineShow = ptr(true)
	}
	if opts.SerAxisLineShow == nil {
		opts.SerAxisLineShow = ptr(true)
	}

	// 3D view (default 30; note: a plain 0 is treated as unset per foundation
	// float64 convention)
	if !(opts.V3DRotX >= -90 && opts.V3DRotX <= 90) || opts.V3DRotX == 0 {
		opts.V3DRotX = 30
	}
	if !(opts.V3DRotY >= 0 && opts.V3DRotY <= 360) || opts.V3DRotY == 0 {
		opts.V3DRotY = 30
	}
	if !(opts.V3DPerspective >= 0 && opts.V3DPerspective <= 240) || opts.V3DPerspective == 0 {
		opts.V3DPerspective = 30
	}
	if opts.V3DRAngAx == nil {
		opts.V3DRAngAx = ptr(true)
	}

	// D: chart
	if !(opts.BarGapWidthPct >= 0 && opts.BarGapWidthPct <= 1000) || opts.BarGapWidthPct == 0 {
		opts.BarGapWidthPct = 150
	}
	if !(opts.BarGapDepthPct >= 0 && opts.BarGapDepthPct <= 1000) || opts.BarGapDepthPct == 0 {
		opts.BarGapDepthPct = 150
	}

	if opts.ChartColors == nil {
		if opts.Type == ChartTypePie || opts.Type == ChartTypeDoughnut {
			opts.ChartColors = PIECHART_COLORS
		} else {
			opts.ChartColors = BARCHART_COLORS
		}
	}

	// border (DEPRECATED v3.11.0) -> plotArea.border
	if opts.Border != nil {
		if opts.Border.Pt == 0 {
			opts.Border.Pt = DEF_CHART_BORDER.Pt
		}
		if opts.Border.Color == "" {
			opts.Border.Color = DEF_CHART_BORDER.Color
		}
	}
	if opts.PlotArea == nil {
		opts.PlotArea = &ChartFillLineProps{}
	}
	if opts.PlotArea.Border != nil {
		if opts.PlotArea.Border.Pt == 0 {
			opts.PlotArea.Border.Pt = DEF_CHART_BORDER.Pt
		}
		if opts.PlotArea.Border.Color == "" {
			opts.PlotArea.Border.Color = DEF_CHART_BORDER.Color
		}
	}
	if opts.Border != nil {
		opts.PlotArea.Border = opts.Border
	}
	if opts.PlotArea.Fill == nil {
		opts.PlotArea.Fill = &ShapeFillProps{}
	}
	if opts.Fill != "" {
		opts.PlotArea.Fill.Color = opts.Fill
	}

	if opts.ChartArea == nil {
		opts.ChartArea = &ChartAreaProps{}
	}
	if opts.ChartArea.Border != nil {
		opts.ChartArea.Border = &BorderProps{
			Color: orStr(opts.ChartArea.Border.Color, DEF_CHART_BORDER.Color),
			Pt:    orF(opts.ChartArea.Border.Pt, DEF_CHART_BORDER.Pt),
		}
	}
	if opts.ChartArea.RoundedCorners == nil {
		opts.ChartArea.RoundedCorners = ptr(true)
	}

	if opts.DataBorder != nil {
		if opts.DataBorder.Pt == 0 {
			opts.DataBorder.Pt = 0.75
		}
		if opts.DataBorder.Color != "" {
			isHex := len(opts.DataBorder.Color) == 6 && RegexHexColor.MatchString(opts.DataBorder.Color)
			if !isHex && !isSchemeColor(opts.DataBorder.Color) {
				opts.DataBorder.Color = "F9F9F9"
			}
		}
	}

	if opts.DataLabelFormatCode == "" && opts.Type == ChartTypeScatter {
		opts.DataLabelFormatCode = "General"
	}
	if opts.DataLabelFormatCode == "" && (opts.Type == ChartTypePie || opts.Type == ChartTypeDoughnut) {
		if opts.ShowPercent != nil && *opts.ShowPercent {
			opts.DataLabelFormatCode = "0%"
		} else {
			opts.DataLabelFormatCode = "General"
		}
	}
	if opts.DataLabelFormatCode == "" {
		opts.DataLabelFormatCode = "#,##0"
	}

	if opts.DataLabelFormatScatter == "" && opts.Type == ChartTypeScatter {
		opts.DataLabelFormatScatter = "custom"
	}

	opts.LineSize = orF(opts.LineSize, 2)

	if opts.Type == ChartTypeArea || opts.Type == ChartTypeBar || opts.Type == ChartTypeBar3d || opts.Type == ChartTypeLine {
		v := opts.CatAxisMultiLevelLabels != nil && *opts.CatAxisMultiLevelLabels
		opts.CatAxisMultiLevelLabels = ptr(v)
	} else {
		opts.CatAxisMultiLevelLabels = nil
	}

	// STEP 4/5: build slide object + chart relationship.
	rid := getNewRelId(target)
	resultObject := &SlideObject{
		Type:     SlideObjectTypeChart,
		ChartRID: rid,
		// Carry the geometry/name/altText onto the slide object so the chart's
		// <p:graphicFrame> (rendered by slideObjectToXml) gets its position,
		// cNvPr name, and descr. Mirrors TS storing `options` on the chart
		// slide object.
		Options: &ObjectOptions{
			PositionProps:   opts.PositionProps,
			ObjectNameProps: opts.ObjectNameProps,
			AltText:         opts.AltText,
		},
	}

	rel := SlideRelChart{
		Opts:       opts,
		Data:       tmpData,
		RID:        rid,
		GlobalID:   chartID,
		FileName:   fmt.Sprintf("chart%d.xml", chartID),
		Target:     fmt.Sprintf("/ppt/charts/chart%d.xml", chartID),
		MultiTypes: multiTypes,
	}
	if !isMulti {
		rel.Type = opts.Type
	}
	target.RelsChart = append(target.RelsChart, rel)

	target.SlideObjects = append(target.SlideObjects, *resultObject)
	return resultObject, nil
}

// ---------------------------------------------------------------------------
// addImageDefinition
// ---------------------------------------------------------------------------

// addImageDefinition registers an image (and its rels) on the slide. Ports the
// TS `addImageDefinition`. TS returns null on bad input (console.error); here
// those become returned errors.
func addImageDefinition(target *PresSlide, opt *ImageProps) error {
	if opt == nil {
		opt = &ImageProps{}
	}
	newObject := &SlideObject{}

	// Position/size (mirrors `opt.x || 0` then `|| 1` for w/h).
	x := Inches(0)
	if opt.X != nil {
		x = *opt.X
	}
	y := Inches(0)
	if opt.Y != nil {
		y = *opt.Y
	}
	w := Inches(1)
	if opt.W != nil && (opt.W.Val != 0 || opt.W.IsPct) {
		w = *opt.W
	}
	h := Inches(1)
	if opt.H != nil && (opt.H.Val != 0 || opt.H.IsPct) {
		h = *opt.H
	}

	sizing := opt.Sizing
	objHyperlink := opt.Hyperlink
	strImageData := opt.Data
	strImagePath := opt.Path
	imageRelID := getNewRelId(target)
	objectName := fmt.Sprintf("Image %d", countSlideObjectsOfType(target, SlideObjectTypeImage))
	if opt.ObjectName != "" {
		objectName = encodeXmlEntities(opt.ObjectName)
	}

	// REALITY-CHECK
	if strImagePath == "" && strImageData == "" {
		return errors.New("ERROR: addImage() requires either 'data' or 'path' parameter!")
	}
	if strImageData != "" && !strings.Contains(strings.ToLower(strImageData), "base64,") {
		return errors.New("ERROR: Image `data` value lacks a base64 header! Ex: 'image/png;base64,NMP[...]')")
	}

	// STEP 1: extension (split to address URLs with params)
	sExt := strImagePath
	if idx := strings.LastIndex(sExt, "/"); idx >= 0 {
		sExt = sExt[idx+1:]
	}
	sExt = strings.Split(sExt, "?")[0]
	dotParts := strings.Split(sExt, ".")
	sExt = dotParts[len(dotParts)-1]
	sExt = strings.Split(sExt, "#")[0]
	if sExt == "" {
		sExt = "png"
	}
	strImgExtn := strings.ToLower(sExt)

	if strImageData != "" {
		if m := imageDataMimeRe.FindStringSubmatch(strImageData); m != nil {
			strImgExtn = m[1]
		} else if strings.Contains(strings.ToLower(strImageData), "image/svg+xml") {
			strImgExtn = "svg"
		}
	}

	// STEP 2: type/path
	newObject.Type = SlideObjectTypeImage
	newObject.Image = orStr(strImagePath, "preencoded.png")

	// STEP 3: properties & options
	newObject.Options = &ObjectOptions{
		PositionProps:   PositionProps{X: &x, Y: &y, W: &w, H: &h},
		ObjectNameProps: ObjectNameProps{ObjectName: objectName},
		AltText:         opt.AltText,
		Rounding:        ptr(opt.Rounding != nil && *opt.Rounding),
		Sizing:          sizing,
		Placeholder:     opt.Placeholder,
		Rotate:          opt.Rotate,
		FlipV:           ptr(opt.FlipV != nil && *opt.FlipV),
		FlipH:           ptr(opt.FlipH != nil && *opt.FlipH),
		Shadow:          correctShadowOptions(opt.Shadow),
	}
	newObject.Options.Transparency = opt.Transparency

	// STEP 4: add to slide rels
	if strImgExtn == "svg" {
		// SVG consumes TWO rIds: a PNG fallback + the SVG itself.
		t1 := fmt.Sprintf("../media/image-%d-%d.png", target.SlideNum, len(target.RelsMedia)+1)
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:     orStr(strImagePath, strImageData+"png"),
			Type:     "image/png",
			Extn:     "png",
			Data:     strImageData,
			RID:      imageRelID,
			Target:   t1,
			IsSvgPng: ptr(true),
			SvgSize: &SlideRelMediaSize{
				W: float64(getSmartParseNumber(*newObject.Options.W, "X", target.PresLayout)),
				H: float64(getSmartParseNumber(*newObject.Options.H, "Y", target.PresLayout)),
			},
		})
		newObject.ImageRID = imageRelID
		t2 := fmt.Sprintf("../media/image-%d-%d.%s", target.SlideNum, len(target.RelsMedia)+1, strImgExtn)
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:   orStr(strImagePath, strImageData),
			Type:   "image/svg+xml",
			Extn:   strImgExtn,
			Data:   strImageData,
			RID:    imageRelID + 1,
			Target: t2,
		})
		newObject.ImageRID = imageRelID + 1
	} else {
		// PERF: duplicate media reuses the existing Target.
		dupeTarget := ""
		for i := range target.RelsMedia {
			item := target.RelsMedia[i]
			if item.Path != "" && item.Path == strImagePath && item.Type == "image/"+strImgExtn && !(item.IsDuplicate != nil && *item.IsDuplicate) {
				dupeTarget = item.Target
				break
			}
		}
		tgt := dupeTarget
		if tgt == "" {
			tgt = fmt.Sprintf("../media/image-%d-%d.%s", target.SlideNum, len(target.RelsMedia)+1, strImgExtn)
		}
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:        orStr(strImagePath, "preencoded."+strImgExtn),
			Type:        "image/" + strImgExtn,
			Extn:        strImgExtn,
			Data:        strImageData,
			RID:         imageRelID,
			IsDuplicate: ptr(dupeTarget != ""),
			Target:      tgt,
		})
		newObject.ImageRID = imageRelID
	}

	// STEP 5: hyperlink support
	if objHyperlink != nil {
		if objHyperlink.URL == "" && objHyperlink.Slide == 0 {
			return errors.New("ERROR: `hyperlink` option requires either: `url` or `slide`")
		}
		imageRelID++
		data := "dummy"
		if objHyperlink.Slide != 0 {
			data = "slide"
		}
		hTarget := objHyperlink.URL
		if hTarget == "" {
			hTarget = strconv.Itoa(objHyperlink.Slide)
		}
		target.Rels = append(target.Rels, SlideRel{
			Type:   SlideObjectTypeHyperlink,
			Data:   data,
			RID:    imageRelID,
			Target: hTarget,
		})
		objHyperlink.RID = imageRelID
		newObject.Hyperlink = objHyperlink
	}

	// STEP 6: add object to slide
	target.SlideObjects = append(target.SlideObjects, *newObject)
	return nil
}

// ---------------------------------------------------------------------------
// addMediaDefinition
// ---------------------------------------------------------------------------

// addMediaDefinition registers audio/video/online media (and its rels). Ports
// the TS `addMediaDefinition`; TS `throw` becomes a returned error.
func addMediaDefinition(target *PresSlide, opt *MediaProps) error {
	if opt == nil {
		opt = &MediaProps{}
	}
	intPosX := Inches(0)
	if opt.X != nil {
		intPosX = *opt.X
	}
	intPosY := Inches(0)
	if opt.Y != nil {
		intPosY = *opt.Y
	}
	intSizeX := Inches(2)
	if opt.W != nil && (opt.W.Val != 0 || opt.W.IsPct) {
		intSizeX = *opt.W
	}
	intSizeY := Inches(2)
	if opt.H != nil && (opt.H.Val != 0 || opt.H.IsPct) {
		intSizeY = *opt.H
	}
	strData := opt.Data
	strLink := opt.Link
	strPath := opt.Path
	strType := orStr(opt.Type, "audio")
	strCover := orStr(opt.Cover, IMG_PLAYBTN)
	objectName := fmt.Sprintf("Media %d", countSlideObjectsOfType(target, SlideObjectTypeMedia))
	if opt.ObjectName != "" {
		objectName = encodeXmlEntities(opt.ObjectName)
	}

	slideData := &SlideObject{Type: SlideObjectTypeMedia}

	// STEP 1: REALITY-CHECK
	if strPath == "" && strData == "" && strType != "online" {
		return errors.New("addMedia() error: either `data` or `path` are required!")
	}
	if strData != "" && !strings.Contains(strings.ToLower(strData), "base64,") {
		return errors.New("addMedia() error: `data` value lacks a base64 header! Ex: 'video/mpeg;base64,NMP[...]')")
	}
	if strCover != "" && !strings.Contains(strings.ToLower(strCover), "base64,") {
		return errors.New("addMedia() error: `cover` value lacks a base64 header! Ex: 'data:image/png;base64,iV[...]')")
	}
	if strType == "online" && strLink == "" {
		return errors.New("addMedia() error: online videos require `link` value")
	}

	strExtn := opt.Extn
	if strExtn == "" {
		if strData != "" {
			parts := strings.Split(strData, ";")
			mparts := strings.Split(parts[0], "/")
			if len(mparts) > 1 {
				strExtn = mparts[1]
			}
		} else if strPath != "" {
			pparts := strings.Split(strPath, ".")
			strExtn = pparts[len(pparts)-1]
		}
	}
	if strExtn == "" {
		strExtn = "mp3"
	}

	// STEP 2: type, media
	slideData.Mtype = strType
	slideData.Media = orStr(strPath, "preencoded.mov")
	slideData.Options = &ObjectOptions{}

	// STEP 3: properties
	slideData.Options.X = &intPosX
	slideData.Options.Y = &intPosY
	slideData.Options.W = &intSizeX
	slideData.Options.H = &intSizeY
	slideData.Options.ObjectName = objectName

	// STEP 4: add to slide rels
	if strType == "online" {
		relID1 := getNewRelId(target)
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:   orStr(strPath, "preencoded"+strExtn),
			Data:   "dummy",
			Type:   "online",
			Extn:   strExtn,
			RID:    relID1,
			Target: strLink,
		})
		slideData.MediaRID = relID1

		coverTarget := fmt.Sprintf("../media/image-%d-%d.png", target.SlideNum, len(target.RelsMedia)+1)
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:   "preencoded.png",
			Data:   strCover,
			Type:   "image/png",
			Extn:   "png",
			RID:    getNewRelId(target),
			Target: coverTarget,
		})
	} else {
		dupeTarget := ""
		for i := range target.RelsMedia {
			item := target.RelsMedia[i]
			if item.Path != "" && item.Path == strPath && item.Type == strType+"/"+strExtn && !(item.IsDuplicate != nil && *item.IsDuplicate) {
				dupeTarget = item.Target
				break
			}
		}

		// A: relationships/video
		relID1 := getNewRelId(target)
		vidTarget := dupeTarget
		if vidTarget == "" {
			vidTarget = fmt.Sprintf("../media/media-%d-%d.%s", target.SlideNum, len(target.RelsMedia)+1, strExtn)
		}
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:        orStr(strPath, "preencoded"+strExtn),
			Type:        strType + "/" + strExtn,
			Extn:        strExtn,
			Data:        strData,
			RID:         relID1,
			IsDuplicate: ptr(dupeTarget != ""),
			Target:      vidTarget,
		})
		slideData.MediaRID = relID1

		// B: relationships/media
		medTarget := dupeTarget
		if medTarget == "" {
			medTarget = fmt.Sprintf("../media/media-%d-%d.%s", target.SlideNum, len(target.RelsMedia)+0, strExtn)
		}
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:        orStr(strPath, "preencoded"+strExtn),
			Type:        strType + "/" + strExtn,
			Extn:        strExtn,
			Data:        strData,
			RID:         getNewRelId(target),
			IsDuplicate: ptr(dupeTarget != ""),
			Target:      medTarget,
		})

		// C: cover image
		coverTarget := fmt.Sprintf("../media/image-%d-%d.png", target.SlideNum, len(target.RelsMedia)+1)
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:   "preencoded.png",
			Type:   "image/png",
			Extn:   "png",
			Data:   strCover,
			RID:    getNewRelId(target),
			Target: coverTarget,
		})
	}

	// LAST
	target.SlideObjects = append(target.SlideObjects, *slideData)
	return nil
}

// ---------------------------------------------------------------------------
// addNotesDefinition
// ---------------------------------------------------------------------------

// addNotesDefinition adds a notes slide object. Ports the TS `addNotesDefinition`.
func addNotesDefinition(target *PresSlide, notes string) {
	target.SlideObjects = append(target.SlideObjects, SlideObject{
		Type: SlideObjectTypeNotes,
		Text: []TextProps{{Text: notes}},
	})
}

// ---------------------------------------------------------------------------
// addShapeDefinition
// ---------------------------------------------------------------------------

// addShapeDefinition adds a shape object. Ports the TS `addShapeDefinition`;
// a missing shape name returns an error (TS throws).
func addShapeDefinition(target *PresSlide, shapeName ShapeType, opts *ShapeProps) error {
	if opts == nil {
		opts = &ShapeProps{}
	}
	if opts.Line == nil {
		opts.Line = &ShapeLineProps{ShapeFillProps: ShapeFillProps{Type: "none"}}
	}
	options := objectOptionsFromShapeProps(opts)

	shp := shapeName
	if shp == "" {
		shp = ShapeTypeRect
	}
	newObject := &SlideObject{
		Type:    SlideObjectTypeText,
		Shape:   shp,
		Options: options,
	}

	// Reality check
	if shapeName == "" {
		return errors.New("Missing/Invalid shape parameter! Example: `addShape(pptxgen.shapes.LINE, {x:1, y:1, w:1, h:1});`")
	}

	// 1: ShapeLineProps defaults
	line := options.Line
	newLineOpts := &ShapeLineProps{}
	newLineOpts.Type = orStr(line.Type, "solid")
	newLineOpts.Color = orStr(line.Color, DEF_SHAPE_LINE_COLOR)
	newLineOpts.Transparency = line.Transparency
	newLineOpts.Width = orF(line.Width, 1)
	newLineOpts.DashType = orStr(line.DashType, "solid")
	newLineOpts.BeginArrowType = line.BeginArrowType
	newLineOpts.EndArrowType = line.EndArrowType
	if options.Line != nil && options.Line.Type != "none" {
		options.Line = newLineOpts
	}

	// 2: option defaults
	if options.X == nil {
		options.X = ptr(Inches(1))
	}
	if options.Y == nil {
		options.Y = ptr(Inches(1))
	}
	if options.W == nil {
		options.W = ptr(Inches(1))
	}
	if options.H == nil {
		options.H = ptr(Inches(1))
	}
	if options.ObjectName != "" {
		options.ObjectName = encodeXmlEntities(options.ObjectName)
	} else {
		options.ObjectName = fmt.Sprintf("Shape %d", countSlideObjectsOfType(target, SlideObjectTypeText))
	}

	// 3: deprecated line opts
	if options.LineSize != 0 {
		options.Line.Width = options.LineSize
	}
	if options.LineDash != "" {
		options.Line.DashType = options.LineDash
	}
	if options.LineHead != "" {
		options.Line.BeginArrowType = options.LineHead
	}
	if options.LineTail != "" {
		options.Line.EndArrowType = options.LineTail
	}

	// 4: hyperlink rels
	createHyperlinkRels(target, newObject)

	// LAST
	target.SlideObjects = append(target.SlideObjects, *newObject)
	return nil
}

// ---------------------------------------------------------------------------
// addTableDefinition
// ---------------------------------------------------------------------------

// addTableDefinition validates and normalizes table rows/options, computes the
// table width, and appends a table slide object (or, when auto-paging, defers
// to getSlidesForTableRows and the addSlide/getSlide callbacks). Ports the TS
// `addTableDefinition`. Returns any auto-paged slides beyond the first.
func addTableDefinition(
	target *PresSlide,
	tableRows []TableRow,
	options *TableProps,
	slideLayout *SlideLayout,
	presLayout PresLayout,
	addSlide func(*AddSlideProps) *PresSlide,
	getSlide func(int) *PresSlide,
) ([]*PresSlide, error) {
	opt := options
	if opt == nil {
		opt = &TableProps{}
	}
	if opt.ObjectName != "" {
		opt.ObjectName = encodeXmlEntities(opt.ObjectName)
	} else {
		opt.ObjectName = fmt.Sprintf("Table %d", countSlideObjectsOfType(target, SlideObjectTypeTable))
	}

	// STEP 1: REALITY-CHECK
	if tableRows == nil || len(tableRows) == 0 {
		return nil, errors.New("addTable: Array expected! EX: 'slide.addTable( [rows], {options} );' (https://gitbrent.github.io/PptxGenJS/docs/api-tables.html)")
	}
	if tableRows[0] == nil {
		return nil, errors.New("addTable: 'rows' should be an array of cells! EX: 'slide.addTable( [ ['A'], ['B'], {text:'C',options:{align:'center'}} ] );' (https://gitbrent.github.io/PptxGenJS/docs/api-tables.html)")
	}

	// STEP 2: transform rows into well-formatted TableCell's
	arrRows := make([][]TableCell, 0, len(tableRows))
	for _, row := range tableRows {
		newRow := make([]TableCell, 0, len(row))
		for _, cell := range row {
			newCell := TableCell{
				Type:    SlideObjectTypeTablecell,
				Text:    cell.Text,
				Options: &TableCellProps{},
			}
			if cell.Options != nil {
				newCell.Options = cell.Options
			}
			if len(cell.TextCells) > 0 {
				newCell.TextCells = cell.TextCells
			}

			// C: cell borders
			border := newCell.Options.Border
			if border == nil {
				border = opt.Border
			}
			if border == nil {
				border = []BorderProps{{Type: "none"}, {Type: "none"}, {Type: "none"}, {Type: "none"}}
			}
			// single border -> replicate to 4 sides
			if len(border) == 1 {
				b := border[0]
				border = []BorderProps{b, b, b, b}
			}
			// ensure 4 sides
			for len(border) < 4 {
				border = append(border, BorderProps{Type: "none"})
			}
			// null/empty side -> none
			for i := 0; i < 4; i++ {
				if border[i].Type == "" && border[i].Color == "" && border[i].Pt == 0 {
					border[i] = BorderProps{Type: "none"}
				}
			}
			// complete each side
			completed := make([]BorderProps, 4)
			for i := 0; i < 4; i++ {
				completed[i] = BorderProps{
					Type:  orStr(border[i].Type, DEF_CELL_BORDER.Type),
					Color: orStr(border[i].Color, DEF_CELL_BORDER.Color),
					Pt:    orF(border[i].Pt, DEF_CELL_BORDER.Pt),
				}
			}
			newCell.Options.Border = completed

			newRow = append(newRow, newCell)
		}
		arrRows = append(arrRows, newRow)
	}

	// STEP 3: options
	xc := Coord{Val: EMU / 2}
	if opt.X != nil {
		xc = *opt.X
	}
	opt.X = ptr(Coord{Val: float64(getSmartParseNumber(xc, "X", presLayout))})
	yc := Coord{Val: EMU / 2}
	if opt.Y != nil {
		yc = *opt.Y
	}
	opt.Y = ptr(Coord{Val: float64(getSmartParseNumber(yc, "Y", presLayout))})
	if opt.H != nil {
		opt.H = ptr(Coord{Val: float64(getSmartParseNumber(*opt.H, "Y", presLayout))})
	}
	opt.FontSize = orF(opt.FontSize, DEF_FONT_SIZE)
	if opt.Margin == nil {
		opt.Margin = []float64{DEF_CELL_MARGIN_IN[0], DEF_CELL_MARGIN_IN[1], DEF_CELL_MARGIN_IN[2], DEF_CELL_MARGIN_IN[3]}
	}
	if len(opt.Margin) == 1 {
		m := opt.Margin[0]
		opt.Margin = []float64{m, m, m, m}
	}

	// hyperlink detection (don't force default color on tables w/ hyperlinks)
	hasHyperlink := false
	for r := range arrRows {
		for c := range arrRows[r] {
			if arrRows[r][c].Options != nil && arrRows[r][c].Options.Hyperlink != nil {
				hasHyperlink = true
			}
		}
	}
	if !hasHyperlink {
		if opt.Color == "" {
			opt.Color = DEF_FONT_COLOR
		}
	}

	// opt-level border completion
	if opt.Border != nil {
		completed := make([]BorderProps, 4)
		for i := 0; i < 4; i++ {
			if i < len(opt.Border) && !(opt.Border[i].Type == "" && opt.Border[i].Color == "" && opt.Border[i].Pt == 0) {
				completed[i] = BorderProps{
					Type:  orStr(opt.Border[i].Type, DEF_CELL_BORDER.Type),
					Color: orStr(opt.Border[i].Color, DEF_CELL_BORDER.Color),
					Pt:    orF(opt.Border[i].Pt, DEF_CELL_BORDER.Pt),
				}
			} else {
				completed[i] = BorderProps{Type: "none"}
			}
		}
		opt.Border = completed
	}

	autoPage := opt.AutoPage != nil && *opt.AutoPage
	opt.AutoPage = ptr(autoPage)
	opt.AutoPageRepeatHeader = ptr(opt.AutoPageRepeatHeader != nil && *opt.AutoPageRepeatHeader)
	if opt.AutoPageHeaderRows == 0 {
		opt.AutoPageHeaderRows = 1
	}
	if opt.AutoPageLineWeight != 0 {
		if opt.AutoPageLineWeight > 1 {
			opt.AutoPageLineWeight = 1
		} else if opt.AutoPageLineWeight < -1 {
			opt.AutoPageLineWeight = -1
		}
	}

	// Slide margins for width calc.
	arrTableMargin := []float64{DEF_SLIDE_MARGIN_IN[0], DEF_SLIDE_MARGIN_IN[1], DEF_SLIDE_MARGIN_IN[2], DEF_SLIDE_MARGIN_IN[3]}
	if slideLayout != nil && slideLayout.Margin != nil {
		if len(slideLayout.Margin) >= 4 {
			arrTableMargin = slideLayout.Margin
		} else if len(slideLayout.Margin) == 1 {
			m := slideLayout.Margin[0]
			arrTableMargin = []float64{m, m, m, m}
		}
	}

	// Calc table width.
	if opt.ColW != nil && len(opt.ColW) > 0 {
		firstRowColCnt := 0
		for c := range arrRows[0] {
			if arrRows[0][c].Options != nil && arrRows[0][c].Options.Colspan > 0 {
				firstRowColCnt += arrRows[0][c].Options.Colspan
			} else {
				firstRowColCnt++
			}
		}
		if len(opt.ColW) == 1 && firstRowColCnt > 1 {
			opt.W = ptr(Coord{Val: math.Floor(opt.ColW[0] * float64(firstRowColCnt))})
			opt.ColW = nil
		} else if len(opt.ColW) != firstRowColCnt {
			opt.ColW = nil
		}
	} else if opt.W != nil {
		opt.W = ptr(Coord{Val: float64(getSmartParseNumber(*opt.W, "X", presLayout))})
	} else {
		opt.W = ptr(Coord{Val: math.Floor(float64(presLayout.SizeW)/EMU - arrTableMargin[1] - arrTableMargin[3])})
	}

	// STEP 4: convert small (inch) units to EMU now.
	if opt.X != nil && opt.X.Val != 0 && opt.X.Val < 20 {
		opt.X.Val = float64(inch2Emu(opt.X.Val))
	}
	if opt.Y != nil && opt.Y.Val != 0 && opt.Y.Val < 20 {
		opt.Y.Val = float64(inch2Emu(opt.Y.Val))
	}
	if opt.W != nil && opt.W.Val != 0 && opt.W.Val < 20 {
		opt.W.Val = float64(inch2Emu(opt.W.Val))
	}
	if opt.H != nil && opt.H.Val != 0 && opt.H.Val < 20 {
		opt.H.Val = float64(inch2Emu(opt.H.Val))
	}

	// STEP 5: cells already normalized to TableCell in STEP 2 (Go input is typed).

	newAutoPagedSlides := []*PresSlide{}

	// STEP 6: auto-paging
	if !autoPage {
		createHyperlinkRels(target, arrRows)
		optCopy := *opt
		target.SlideObjects = append(target.SlideObjects, SlideObject{
			Type:       SlideObjectTypeTable,
			ArrTabRows: arrRows,
			Options:    tablePropsToObjectOptions(&optCopy),
		})
	} else {
		if opt.AutoPageRepeatHeader != nil && *opt.AutoPageRepeatHeader {
			var head []TableRow
			for idx := range arrRows {
				if idx < opt.AutoPageHeaderRows {
					head = append(head, arrRows[idx])
				}
			}
			opt.ArrObjTabHeadRows = head
		}

		// Reconciled with tables.go: wrap TableProps in TableToSlidesProps,
		// mirroring the fields that shadow between the outer and embedded structs.
		ttsProps := &TableToSlidesProps{
			TableProps:           *opt,
			AutoPageRepeatHeader: opt.AutoPageRepeatHeader,
			NewSlideStartY:       opt.NewSlideStartY,
		}
		for idx, slide := range GetSlidesForTableRows(arrRows, ttsProps, presLayout, slideLayout) {
			// A: create new slide when needed
			if getSlide == nil || getSlide(target.SlideNum+idx) == nil {
				if addSlide != nil {
					masterName := ""
					if slideLayout != nil {
						masterName = slideLayout.Name
					}
					addSlide(&AddSlideProps{MasterName: masterName})
				}
			}

			// B: reset y after first slide
			if idx > 0 {
				startY := opt.AutoPageSlideStartY
				if startY == 0 {
					startY = opt.NewSlideStartY
				}
				if startY == 0 {
					startY = arrTableMargin[0]
				}
				opt.Y = ptr(Coord{Val: float64(inch2Emu(startY))})
			}

			// C: add table to slide
			var newSlide *PresSlide
			if getSlide != nil {
				newSlide = getSlide(target.SlideNum + idx)
			}
			if newSlide == nil {
				continue
			}
			opt.AutoPage = ptr(false)
			createHyperlinkRels(newSlide, slide.Rows)
			optCopy := *opt
			_, _ = addTableDefinition(newSlide, slide.Rows, &optCopy, &newSlide.SlideLayout, presLayout, addSlide, getSlide)
			if idx > 0 {
				newAutoPagedSlides = append(newAutoPagedSlides, newSlide)
			}
		}
	}

	return newAutoPagedSlides, nil
}

// ---------------------------------------------------------------------------
// addTextDefinition
// ---------------------------------------------------------------------------

// addTextDefinition adds a text (or placeholder) object, applying the cleanOpts
// defaulting to both the object options and each text run. Ports the TS
// `addTextDefinition`. opts is *ObjectOptions (rather than TS TextPropsOptions)
// so placeholder idx/type set by createSlideMaster survive onto the object.
func addTextDefinition(target *PresSlide, text []TextProps, opts *ObjectOptions, isPlaceholder bool) {
	typ := SlideObjectTypeText
	if isPlaceholder {
		typ = SlideObjectTypePlaceholder
	}
	if opts == nil {
		opts = &ObjectOptions{}
	}
	shape := opts.Shape
	if shape == "" {
		shape = ShapeTypeRect
	}
	txt := text
	if len(txt) == 0 {
		txt = []TextProps{{Text: "", Options: nil}}
	}
	newObject := &SlideObject{
		Type:    typ,
		Shape:   shape,
		Text:    txt,
		Options: opts,
	}

	cleanOpts := func(io *ObjectOptions) *ObjectOptions {
		// STEP 1
		// A.1: color (placeholders inherit/override; don't default them)
		if io.Placeholder == "" {
			if io.Color == "" {
				if newObject.Options != nil && newObject.Options.Color != "" {
					io.Color = newObject.Options.Color
				} else if target.Color != "" {
					io.Color = target.Color
				} else {
					io.Color = DEF_FONT_COLOR
				}
			}
		}

		// A.2: placeholders inherit their bullets (bullet nil == false already)

		// A.3: text targeting a placeholder inherits the placeholder's options
		if io.Placeholder != "" && len(target.SlideLayout.SlideObjects) > 0 {
			for i := range target.SlideLayout.SlideObjects {
				plh := target.SlideLayout.SlideObjects[i]
				if plh.Type == SlideObjectTypePlaceholder && plh.Options != nil && plh.Options.Placeholder != "" && plh.Options.Placeholder == io.Placeholder {
					mergePlaceholderOptions(io, plh.Options)
					break
				}
			}
		}

		// A.4: object name
		if io.ObjectName != "" {
			io.ObjectName = encodeXmlEntities(io.ObjectName)
		} else {
			io.ObjectName = fmt.Sprintf("Text %d", countSlideObjectsOfType(target, SlideObjectTypeText))
		}

		// B: line shape handling (deprecated opts)
		if io.Shape == ShapeTypeLine {
			line := io.Line
			if line == nil {
				line = &ShapeLineProps{}
			}
			nlo := &ShapeLineProps{}
			nlo.Type = orStr(line.Type, "solid")
			nlo.Color = orStr(line.Color, DEF_SHAPE_LINE_COLOR)
			nlo.Transparency = line.Transparency
			nlo.Width = orF(line.Width, 1)
			nlo.DashType = orStr(line.DashType, "solid")
			nlo.BeginArrowType = line.BeginArrowType
			nlo.EndArrowType = line.EndArrowType
			io.Line = nlo
			if io.LineSize != 0 {
				io.Line.Width = io.LineSize
			}
			if io.LineDash != "" {
				io.Line.DashType = io.LineDash
			}
			if io.LineHead != "" {
				io.Line.BeginArrowType = io.LineHead
			}
			if io.LineTail != "" {
				io.Line.EndArrowType = io.LineTail
			}
		}

		// C: line opts
		if io.Line == nil {
			io.Line = &ShapeLineProps{}
		}

		// D: transform text options to body properties
		if io.BodyProp == nil {
			io.BodyProp = &BodyProps{}
		}
		io.BodyProp.AutoFit = ptr(io.AutoFit != nil && *io.AutoFit)
		if io.Placeholder == "" {
			io.BodyProp.Anchor = TextVAlignCTR
		} else {
			io.BodyProp.Anchor = ""
		}
		io.BodyProp.Vert = io.Vert
		if io.Wrap != nil {
			io.BodyProp.Wrap = ptr(*io.Wrap)
		} else {
			io.BodyProp.Wrap = ptr(true)
		}

		// E: inset (deprecated - use margin)
		if io.Inset != 0 {
			io.BodyProp.LIns = float64(inch2Emu(io.Inset))
			io.BodyProp.RIns = float64(inch2Emu(io.Inset))
			io.BodyProp.TIns = float64(inch2Emu(io.Inset))
			io.BodyProp.BIns = float64(inch2Emu(io.Inset))
		}

		// STEP 2: align/valign -> body props
		al := strings.ToLower(io.Align)
		switch {
		case strings.HasPrefix(al, "c"):
			io.BodyProp.Align = TextHAlignCenter
		case strings.HasPrefix(al, "l"):
			io.BodyProp.Align = TextHAlignLeft
		case strings.HasPrefix(al, "r"):
			io.BodyProp.Align = TextHAlignRight
		case strings.HasPrefix(al, "j"):
			io.BodyProp.Align = TextHAlignJustify
		}
		va := strings.ToLower(io.Valign)
		switch {
		case strings.HasPrefix(va, "b"):
			io.BodyProp.Anchor = TextVAlignB
		case strings.HasPrefix(va, "m"):
			io.BodyProp.Anchor = TextVAlignCTR
		case strings.HasPrefix(va, "t"):
			io.BodyProp.Anchor = TextVAlignT
		}

		// STEP 3: shadow
		correctShadowOptions(io.Shadow)

		return io
	}

	// STEP 1: clean object options
	newObject.Options = cleanOpts(newObject.Options)

	// STEP 2: clean each text run's options
	for i := range newObject.Text {
		to := newObject.Text[i].Options
		oo := objectOptionsFromTextPropsOptions(to)
		oo = cleanOpts(oo)
		newObject.Text[i].Options = textPropsOptionsFromObjectOptions(oo)
	}

	// STEP 3: hyperlinks
	createHyperlinkRels(target, newObject.Text)

	// LAST
	target.SlideObjects = append(target.SlideObjects, *newObject)
}

// ---------------------------------------------------------------------------
// addPlaceholdersToSlideLayouts
// ---------------------------------------------------------------------------

// addPlaceholdersToSlideLayouts adds any layout placeholders not already
// present on the slide. Ports the TS `addPlaceholdersToSlideLayouts`.
func addPlaceholdersToSlideLayouts(slide *PresSlide) {
	for i := range slide.SlideLayout.SlideObjects {
		layoutObj := slide.SlideLayout.SlideObjects[i]
		if layoutObj.Type != SlideObjectTypePlaceholder || layoutObj.Options == nil {
			continue
		}
		exists := false
		for j := range slide.SlideObjects {
			so := slide.SlideObjects[j]
			if so.Options != nil && so.Options.Placeholder == layoutObj.Options.Placeholder {
				exists = true
				break
			}
		}
		if !exists {
			addTextDefinition(slide, []TextProps{{Text: ""}}, layoutObj.Options, false)
		}
	}
}

// ---------------------------------------------------------------------------
// addBackgroundDefinition
// ---------------------------------------------------------------------------

// addBackgroundDefinition sets a slide/master background (color or image).
// Ports the TS `addBackgroundDefinition`. Operates on SlideBaseProps so it
// serves both PresSlide and SlideLayout targets.
func addBackgroundDefinition(props *BackgroundProps, target *SlideBaseProps) {
	// A: DEPRECATED bkgd
	if target.Bkgd != nil {
		if target.Background == nil {
			target.Background = &BackgroundProps{}
		}
		switch b := target.Bkgd.(type) {
		case string:
			target.Background.Color = b
		case *BackgroundProps:
			if b.Data != "" {
				target.Background.Data = b.Data
			}
			if b.Path != "" {
				target.Background.Path = b.Path
			}
			if b.Src != "" {
				target.Background.Path = b.Src
			}
		case BackgroundProps:
			if b.Data != "" {
				target.Background.Data = b.Data
			}
			if b.Path != "" {
				target.Background.Path = b.Path
			}
			if b.Src != "" {
				target.Background.Path = b.Src
			}
		}
	}
	if target.Background != nil && target.Background.Fill != "" {
		target.Background.Color = target.Background.Fill
	}

	// B: media
	if props != nil && (props.Path != "" || props.Data != "") {
		path := orStr(props.Path, "preencoded.png")
		dotParts := strings.Split(path, ".")
		strImgExtn := orStr(dotParts[len(dotParts)-1], "png")
		strImgExtn = strings.Split(strImgExtn, "?")[0]
		if strImgExtn == "jpg" {
			strImgExtn = "jpeg"
		}
		intRels := len(target.RelsMedia) + 1
		var data any
		if props.Data != "" {
			data = props.Data
		}
		target.RelsMedia = append(target.RelsMedia, SlideRelMedia{
			Path:   path,
			Type:   string(SlideObjectTypeImage),
			Extn:   strImgExtn,
			Data:   data,
			RID:    intRels,
			Target: fmt.Sprintf("../media/%s-image-%d.%s", strings.ReplaceAll(target.Name, " ", "-"), len(target.RelsMedia)+1, strImgExtn),
		})
		target.BkgdImgRid = intRels
	}
}

// ---------------------------------------------------------------------------
// createHyperlinkRels
// ---------------------------------------------------------------------------

// createHyperlinkRels parses text/shape/table structures and registers a
// 'hyperlink'-type slide rel for each hyperlink found. Ports the TS
// `createHyperlinkRels` (which relies on structural recursion). `text` accepts
// the heterogeneous shapes gen-objects passes: *SlideObject, []TextProps, and
// [][]TableCell.
func createHyperlinkRels(target *PresSlide, text any) {
	switch v := text.(type) {
	case *SlideObject:
		if v == nil {
			return
		}
		if v.Options != nil {
			registerHyperlink(target, v.Options.Hyperlink)
		}
		if len(v.Text) > 0 {
			createHyperlinkRels(target, v.Text)
		}
	case []TextProps:
		for i := range v {
			registerHyperlink(target, textPropsHyperlink(&v[i]))
		}
	case [][]TableCell:
		for r := range v {
			for c := range v[r] {
				cell := v[r][c]
				if cell.Options != nil {
					registerHyperlink(target, cell.Options.Hyperlink)
				}
				if len(cell.TextCells) > 0 {
					for k := range cell.TextCells {
						if cell.TextCells[k].Options != nil {
							registerHyperlink(target, cell.TextCells[k].Options.Hyperlink)
						}
					}
				}
			}
		}
	}
}

func textPropsHyperlink(t *TextProps) *HyperlinkProps {
	if t == nil || t.Options == nil {
		return nil
	}
	return t.Options.Hyperlink
}

// registerHyperlink adds a hyperlink rel for h if it is valid and not already
// registered on target. Mirrors the inner logic of TS createHyperlinkRels.
func registerHyperlink(target *PresSlide, h *HyperlinkProps) {
	if h == nil {
		return
	}
	// requires url or slide
	if h.URL == "" && h.Slide == 0 {
		return
	}

	data := "dummy"
	if h.Slide != 0 {
		data = "slide"
	}
	tgt := encodeXmlEntities(h.URL)
	if tgt == "" {
		tgt = strconv.Itoa(h.Slide)
	}

	if h.RID == 0 {
		relID := getNewRelId(target)
		target.Rels = append(target.Rels, SlideRel{
			Type:   SlideObjectTypeHyperlink,
			Data:   data,
			RID:    relID,
			Target: tgt,
		})
		h.RID = relID
	} else {
		// auto-paged new slide: ensure the rel exists here too
		found := false
		for i := range target.Rels {
			if target.Rels[i].RID == h.RID {
				found = true
				break
			}
		}
		if !found {
			target.Rels = append(target.Rels, SlideRel{
				Type:   SlideObjectTypeHyperlink,
				Data:   data,
				RID:    h.RID,
				Target: tgt,
			})
		}
	}
}
