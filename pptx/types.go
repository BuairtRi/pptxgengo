// Package pptx is a Go port of PptxGenJS.
//
// types.go ports src/core-interfaces.ts: the presentation data model shared by
// all generator code. These are plain data structs (no behavior); user-facing
// classes (Slide, Presentation) live in slide.go / presentation.go.
package pptx

// ---------------------------------------------------------------------------
// Core primitive types
// ---------------------------------------------------------------------------

// Color is a hex RGB value ("FF0000") or a scheme/theme color name ("tx1").
type Color = string

// HexColor is a 6-digit hex RGB value, e.g. "FF3399".
type HexColor = string

// ThemeColorName is a scheme color name ("tx1", "accent1", …).
type ThemeColorName = string

// HAlign is a horizontal alignment: "left" | "center" | "right" | "justify".
type HAlign = string

// VAlign is a vertical alignment: "top" | "middle" | "bottom".
type VAlign = string

// MediaType is an audio/video media type: "audio" | "online" | "video".
type MediaType = string

// Coord is a position/size value: inches by default, or a percentage of the
// slide dimension when IsPct is true. Mirrors the TS `Coord = number | "n%"`.
type Coord struct {
	Val   float64
	IsPct bool
}

// Inches returns a Coord expressed in inches.
func Inches(v float64) Coord { return Coord{Val: v} }

// Percent returns a Coord expressed as a percentage (0-100) of the slide size.
func Percent(v float64) Coord { return Coord{Val: v, IsPct: true} }

// Margin is a points/inches margin. Semantics by length:
//   - nil     : unset
//   - len == 1: uniform margin on all four sides
//   - len == 4: [top, right, bottom, left]
//
// Mirrors the TS `Margin = number | [number, number, number, number]`.
type Margin []float64

// ---------------------------------------------------------------------------
// Position / data-source / name mixins
// ---------------------------------------------------------------------------

// PositionProps is the x/y/w/h location of an object. Nil = inherit/default.
type PositionProps struct {
	X *Coord
	Y *Coord
	H *Coord
	W *Coord
}

// DataOrPathProps supplies content by URL/path or by base64 data.
type DataOrPathProps struct {
	Path string
	Data string
}

// ObjectNameProps sets the object's name (PowerPoint Selection Pane).
type ObjectNameProps struct {
	ObjectName string
}

// ThemeProps sets heading/body theme fonts.
type ThemeProps struct {
	HeadFontFace string
	BodyFontFace string
}

// ---------------------------------------------------------------------------
// Borders / fills / lines / shadows / hyperlinks
// ---------------------------------------------------------------------------

// BorderProps describes a single border edge. Type: "none" | "dash" | "solid".
type BorderProps struct {
	Type  string
	Color HexColor
	Pt    float64
}

// HyperlinkProps is a hyperlink to a slide or URL.
type HyperlinkProps struct {
	RID     int // internal `_rId`
	Slide   int
	URL     string
	Tooltip string
}

// ShadowProps describes an outer/inner shadow. Type: "outer" | "inner" | "none".
type ShadowProps struct {
	Type            string
	Opacity         float64
	Blur            float64
	Angle           float64
	Offset          float64
	Color           HexColor
	RotateWithShape *bool
}

// ShapeFillProps describes a solid/none fill. Type: "none" | "solid".
type ShapeFillProps struct {
	Color        Color
	Transparency float64
	Type         string
	// Alpha is deprecated (v3.3.0) - use Transparency.
	Alpha float64
}

// ShapeLineProps describes a shape outline. Embeds ShapeFillProps.
type ShapeLineProps struct {
	ShapeFillProps
	Width          float64
	DashType       string
	BeginArrowType string
	EndArrowType   string
	// Deprecated (v3.3.0) aliases:
	LineDash string
	LineHead string
	LineTail string
	Pt       float64
	Size     float64
}

// OutlineProps is a text outline (color + size in points).
type OutlineProps struct {
	Color Color
	Size  float64
}

// ---------------------------------------------------------------------------
// Text base props
// ---------------------------------------------------------------------------

// TabStop is a paragraph tab stop. Alignment: "l" | "r" | "ctr" | "dec".
type TabStop struct {
	Position  float64
	Alignment string
}

// UnderlineProps describes underline style/color.
type UnderlineProps struct {
	Style string
	Color Color
}

// BulletProps holds custom bullet options. A non-nil *BulletProps means the
// bullet is enabled (mirrors TS `bullet: true | {…}`); nil means no bullet.
type BulletProps struct {
	Type          string // "bullet" | "number"
	CharacterCode string
	Indent        float64
	NumberType    string
	NumberStartAt int
	// Deprecated (v3.3.0) aliases:
	Code     string
	MarginPt float64
	StartAt  int
	Style    string
}

// TextBaseProps are shared text/paragraph run properties.
type TextBaseProps struct {
	Align           HAlign
	Bold            *bool
	BreakLine       *bool
	Bullet          *BulletProps
	Color           Color
	FontFace        string
	FontSize        float64
	Highlight       HexColor
	Italic          *bool
	Lang            string
	SoftBreakBefore *bool
	TabStops        []TabStop
	TextDirection   string // "horz" | "vert" | "vert270" | "wordArtVert"
	Transparency    float64
	Underline       *UnderlineProps
	Valign          VAlign
}

// PlaceholderProps defines a slide master placeholder.
type PlaceholderProps struct {
	PositionProps
	TextBaseProps
	Name   string
	Type   PlaceholderType
	Margin Margin
}

// ---------------------------------------------------------------------------
// Image / media
// ---------------------------------------------------------------------------

// ImageSizing controls image contain/cover/crop sizing.
type ImageSizing struct {
	Type string // "contain" | "cover" | "crop"
	W    Coord
	H    Coord
	X    *Coord
	Y    *Coord
}

// ImageProps are options for an added image.
type ImageProps struct {
	PositionProps
	DataOrPathProps
	ObjectNameProps
	AltText      string
	FlipH        *bool
	FlipV        *bool
	Hyperlink    *HyperlinkProps
	Placeholder  string
	Rotate       float64
	Rounding     *bool
	Shadow       *ShadowProps
	Sizing       *ImageSizing
	Transparency float64
}

// MediaProps are options for added audio/video/online media.
type MediaProps struct {
	PositionProps
	DataOrPathProps
	ObjectNameProps
	Type  MediaType
	Cover string
	Extn  string
	Link  string
}

// ---------------------------------------------------------------------------
// Shapes
// ---------------------------------------------------------------------------

// ShapeCurve is a custom-geometry curve segment. Type: "arc" | "cubic" | "quadratic".
type ShapeCurve struct {
	Type string
	// arc:
	HR    Coord
	WR    Coord
	StAng float64
	SwAng float64
	// cubic/quadratic control points:
	X1 Coord
	Y1 Coord
	X2 Coord
	Y2 Coord
}

// ShapePoint is a point in a custom-geometry path. Close==true is a path-close op.
type ShapePoint struct {
	X      Coord
	Y      Coord
	MoveTo *bool
	Curve  *ShapeCurve
	Close  bool
}

// ShapeProps are options for an added shape.
type ShapeProps struct {
	PositionProps
	ObjectNameProps
	Align             HAlign
	AngleRange        *[2]float64
	ArcThicknessRatio float64
	Fill              *ShapeFillProps
	FlipH             *bool
	FlipV             *bool
	Hyperlink         *HyperlinkProps
	Line              *ShapeLineProps
	Points            []ShapePoint
	RectRadius        float64
	Rotate            float64
	Shadow            *ShadowProps
	// Deprecated (v3.3.0 / v3.10.0):
	LineSize  float64
	LineDash  string
	LineHead  string
	LineTail  string
	ShapeName string
}

// ---------------------------------------------------------------------------
// Tables
// ---------------------------------------------------------------------------

// TableCellProps are per-cell table options.
type TableCellProps struct {
	TextBaseProps
	AutoPageCharWeight float64
	AutoPageLineWeight float64
	// Border: nil unset; len 1 all sides; len 4 [top,right,bottom,left].
	Border    []BorderProps
	Colspan   int
	Fill      *ShapeFillProps
	Hyperlink *HyperlinkProps
	Margin    Margin
	Rowspan   int
}

// TableProps are options for an added table.
type TableProps struct {
	PositionProps
	TextBaseProps
	ObjectNameProps
	ArrObjTabHeadRows    []TableRow // internal `_arrObjTabHeadRows`
	AutoPage             *bool
	AutoPageCharWeight   float64
	AutoPageLineWeight   float64
	AutoPageRepeatHeader *bool
	AutoPageHeaderRows   int
	AutoPageSlideStartY  float64
	// Border: nil unset; len 1 all sides; len 4 [top,right,bottom,left].
	Border  []BorderProps
	ColW    []float64
	Fill    *ShapeFillProps
	Margin  Margin
	RowH    []float64
	Verbose *bool
	// Deprecated (v3.3.0):
	NewSlideStartY float64
}

// TableCell is a single table cell.
type TableCell struct {
	Type SlideObjectType // internal `_type` (always "tablecell")
	// internal fields:
	Lines       [][]TableCell // `_lines`
	TableCells  []TableCell   // `_tableCells`
	LineHeight  float64       // `_lineHeight`, EMU
	Hmerge      *bool
	Vmerge      *bool
	RowContinue int
	OptImp      any // `_optImp`

	// Text holds the cell's plain text; TextCells holds nested cells when the
	// TS `text` value is a TableCell[] (union simplification).
	Text      string
	TextCells []TableCell
	Options   *TableCellProps
}

// TableRow is one row of table cells.
type TableRow = []TableCell

// TableRowSlide is a set of rows produced by table auto-paging.
type TableRowSlide struct {
	Rows []TableRow
}

// TableToSlidesAddImage adds an image to auto-paged slides.
type TableToSlidesAddImage struct {
	Image   DataOrPathProps
	Options PositionProps
}

// TableToSlidesAddShape adds a shape to auto-paged slides.
type TableToSlidesAddShape struct {
	ShapeName ShapeType
	Options   ShapeProps
}

// TableToSlidesAddTable adds a table to auto-paged slides.
type TableToSlidesAddTable struct {
	Rows    []TableRow
	Options TableProps
}

// TableToSlidesAddText adds text to auto-paged slides.
type TableToSlidesAddText struct {
	Text    []TextProps
	Options TextPropsOptions
}

// TableToSlidesProps are options for tableToSlides.
type TableToSlidesProps struct {
	TableProps
	AddImage             *TableToSlidesAddImage
	AddShape             *TableToSlidesAddShape
	AddTable             *TableToSlidesAddTable
	AddText              *TableToSlidesAddText
	AutoPageRepeatHeader *bool
	MasterSlideName      string
	SlideMargin          Margin
	// Deprecated (v3.3.0):
	AddHeaderToEach *bool
	NewSlideStartY  float64
}

// ---------------------------------------------------------------------------
// Text
// ---------------------------------------------------------------------------

// TextGlowProps describes a text glow effect. Size is required.
type TextGlowProps struct {
	Color   HexColor
	Opacity float64
	Size    float64
}

// BodyProps are internal text-body layout properties (`_bodyProp`).
type BodyProps struct {
	AutoFit *bool
	Align   TextHAlign
	Anchor  TextVAlign
	LIns    float64
	RIns    float64
	TIns    float64
	BIns    float64
	Vert    string
	Wrap    *bool
}

// TextPropsOptions are options for a text run/paragraph.
type TextPropsOptions struct {
	PositionProps
	DataOrPathProps
	TextBaseProps
	ObjectNameProps
	BodyProp *BodyProps // internal `_bodyProp`
	LineIdx  int        // internal `_lineIdx`

	Baseline            float64
	CharSpacing         float64
	Fit                 string // "none" | "shrink" | "resize"
	Fill                *ShapeFillProps
	FlipH               *bool
	FlipV               *bool
	Glow                *TextGlowProps
	Hyperlink           *HyperlinkProps
	IndentLevel         int
	IsTextBox           *bool
	Line                *ShapeLineProps
	LineSpacing         float64
	LineSpacingMultiple float64
	Margin              Margin
	Outline             *OutlineProps
	ParaSpaceAfter      float64
	ParaSpaceBefore     float64
	Placeholder         string
	RectRadius          float64
	Rotate              float64
	RtlMode             *bool
	Shadow              *ShadowProps
	Shape               ShapeType
	Strike              string // "" | "sngStrike" | "dblStrike"
	Subscript           *bool
	Superscript         *bool
	Valign              VAlign
	Vert                string
	Wrap                *bool
	// Deprecated (v3.3.0 / v3.10.0):
	AutoFit    *bool
	ShrinkText *bool
	Inset      float64
	LineDash   string
	LineHead   string
	LineSize   float64
	LineTail   string
}

// TextProps is a single text fragment plus options.
type TextProps struct {
	Text    string
	Options *TextPropsOptions
}

// ---------------------------------------------------------------------------
// Charts
// ---------------------------------------------------------------------------

// OptsChartGridLine describes chart gridline style. Style: "solid"|"dash"|"dot"|"none".
type OptsChartGridLine struct {
	Cap   string // "flat" | "round" | "square"
	Color HexColor
	Size  float64
	Style string
}

// ChartData is one chart data series. Labels covers both string[] and
// string[][] source forms (single-level labels use one inner slice).
type ChartData struct {
	DataIndex int // internal `_dataIndex`
	Labels    [][]string
	Name      string
	Sizes     []float64
	Values    []float64
}

// IChartMulti is one component of a multi-type (combo) chart.
type IChartMulti struct {
	Type    ChartType
	Data    []ChartData
	Options *ChartOptions
}

// ChartAreaProps is chart-area fill/border with rounded corners.
type ChartAreaProps struct {
	Border         *BorderProps
	Fill           *ShapeFillProps
	RoundedCorners *bool
}

// ChartFillLineProps is plot-area (or chart-area) fill/border.
type ChartFillLineProps struct {
	Border *BorderProps
	Fill   *ShapeFillProps
}

// ChartTitlePos is an explicit chart title position (x, y in EMU).
type ChartTitlePos struct {
	X float64
	Y float64
}

// ChartOptions is the flattened union of all chart option interfaces
// (IChartOptsLib). Embeds PositionProps/TextBaseProps/ObjectNameProps.
type ChartOptions struct {
	PositionProps
	TextBaseProps
	ObjectNameProps

	// _type (single type) plus multi-chart component list.
	Type       ChartType
	MultiTypes []IChartMulti

	AltText string

	// --- OptsChartGridLine (top-level; Color/Style shared w/ TextBaseProps.Color) ---
	Cap   string
	Size  float64
	Style string

	// --- base ---
	AxisPos            string
	ChartColors        []HexColor
	ChartColorsOpacity float64
	DataBorder         *BorderProps
	DisplayBlanksAs    string
	InvertedColors     []HexColor
	Layout             *PositionProps
	Shadow             *ShadowProps
	ShowLabel          *bool
	ShowLeaderLines    *bool
	ShowLegend         *bool
	ShowPercent        *bool
	ShowSerName        *bool
	ShowTitle          *bool
	ShowValue          *bool
	V3DPerspective     float64
	V3DRAngAx          *bool
	V3DRotX            float64
	V3DRotY            float64
	ChartArea          *ChartAreaProps
	PlotArea           *ChartFillLineProps
	// Deprecated (v3.11.0):
	Border *BorderProps
	Fill   HexColor

	// --- category axis ---
	CatAxes                 []ChartOptions
	CatAxisBaseTimeUnit     string
	CatAxisCrossesAt        any // number | "autoZero"
	CatAxisHidden           *bool
	CatAxisLabelColor       string
	CatAxisLabelFontBold    *bool
	CatAxisLabelFontFace    string
	CatAxisLabelFontItalic  *bool
	CatAxisLabelFontSize    float64
	CatAxisLabelFrequency   string
	CatAxisLabelPos         string
	CatAxisLabelRotate      float64
	CatAxisLineColor        string
	CatAxisLineShow         *bool
	CatAxisLineSize         float64
	CatAxisLineStyle        string
	CatAxisMajorTickMark    string
	CatAxisMajorTimeUnit    string
	CatAxisMajorUnit        *float64
	CatAxisMaxVal           *float64
	CatAxisMinorTickMark    string
	CatAxisMinorTimeUnit    string
	CatAxisMinorUnit        *float64
	CatAxisMinVal           *float64
	CatAxisMultiLevelLabels *bool
	CatAxisOrientation      string
	CatAxisTitle            string
	CatAxisTitleColor       string
	CatAxisTitleFontFace    string
	CatAxisTitleFontSize    float64
	CatAxisTitleRotate      float64
	CatGridLine             *OptsChartGridLine
	CatLabelFormatCode      string
	SecondaryCatAxis        *bool
	ShowCatAxisTitle        *bool

	// --- series axis ---
	SerAxisBaseTimeUnit    string
	SerAxisHidden          *bool
	SerAxisLabelColor      string
	SerAxisLabelFontBold   *bool
	SerAxisLabelFontFace   string
	SerAxisLabelFontItalic *bool
	SerAxisLabelFontSize   float64
	SerAxisLabelFrequency  string
	SerAxisLabelPos        string
	SerAxisLineColor       string
	SerAxisLineShow        *bool
	SerAxisMajorTimeUnit   string
	SerAxisMajorUnit       *float64
	SerAxisMinorTimeUnit   string
	SerAxisMinorUnit       *float64
	SerAxisOrientation     string
	SerAxisTitle           string
	SerAxisTitleColor      string
	SerAxisTitleFontFace   string
	SerAxisTitleFontSize   float64
	SerAxisTitleRotate     float64
	SerGridLine            *OptsChartGridLine
	SerLabelFormatCode     string
	ShowSerAxisTitle       *bool

	// --- value axis ---
	SecondaryValAxis        *bool
	ShowValAxisTitle        *bool
	ValAxes                 []ChartOptions
	ValAxisCrossesAt        any // number | "autoZero"
	ValAxisDisplayUnit      string
	ValAxisDisplayUnitLabel *bool
	ValAxisHidden           *bool
	ValAxisLabelColor       string
	ValAxisLabelFontBold    *bool
	ValAxisLabelFontFace    string
	ValAxisLabelFontItalic  *bool
	ValAxisLabelFontSize    float64
	ValAxisLabelFormatCode  string
	ValAxisLabelPos         string
	ValAxisLabelRotate      float64
	ValAxisLineColor        string
	ValAxisLineShow         *bool
	ValAxisLineSize         float64
	ValAxisLineStyle        string
	ValAxisLogScaleBase     *float64
	ValAxisMajorTickMark    string
	ValAxisMajorUnit        *float64
	ValAxisMaxVal           *float64
	ValAxisMinorTickMark    string
	ValAxisMinVal           *float64
	ValAxisOrientation      string
	ValAxisTitle            string
	ValAxisTitleColor       string
	ValAxisTitleFontFace    string
	ValAxisTitleFontSize    float64
	ValAxisTitleRotate      float64
	ValGridLine             *OptsChartGridLine
	ValLabelFormatCode      string

	// --- bar ---
	Bar3DShape     string
	BarDir         string
	BarGapDepthPct float64
	BarGapWidthPct float64
	BarGrouping    string
	BarOverlapPct  float64

	// --- doughnut / pie ---
	DataNoEffects *bool
	HoleSize      float64
	FirstSliceAng float64

	// --- line ---
	LineCap                 string
	LineDash                string
	LineDataSymbol          string
	LineDataSymbolLineColor string
	LineDataSymbolLineSize  float64
	LineDataSymbolSize      float64
	LineSize                float64
	LineSmooth              *bool

	// --- radar ---
	RadarStyle string

	// --- data labels ---
	DataLabelBkgrdColors   *bool
	DataLabelColor         string
	DataLabelFontBold      *bool
	DataLabelFontFace      string
	DataLabelFontItalic    *bool
	DataLabelFontSize      float64
	DataLabelFormatCode    string
	DataLabelFormatScatter string
	DataLabelPosition      string

	// --- data table ---
	DataTableFontSize       float64
	DataTableFormatCode     string
	ShowDataTable           *bool
	ShowDataTableHorzBorder *bool
	ShowDataTableKeys       *bool
	ShowDataTableOutline    *bool
	ShowDataTableVertBorder *bool

	// --- legend ---
	LegendColor    string
	LegendFontFace string
	LegendFontSize float64
	LegendPos      string

	// --- title ---
	Title         string
	TitleAlign    string
	TitleBold     *bool
	TitleColor    string
	TitleFontFace string
	TitleFontSize float64
	TitlePos      *ChartTitlePos
	TitleRotate   float64
}

// ---------------------------------------------------------------------------
// Slide relationships (internal)
// ---------------------------------------------------------------------------

// SlideRel is a slide relationship (image/chart/hyperlink/media). `ISlideRel`.
type SlideRel struct {
	Type     SlideObjectType
	Target   string
	FileName string
	Data     any // []byte | string
	Opts     *ChartOptions
	Path     string
	Extn     string
	GlobalID int
	RID      int
}

// SlideRelChart is a chart relationship. `ISlideRelChart`.
type SlideRelChart struct {
	ChartData  // embeds labels/name/sizes/values/_dataIndex
	Type       ChartType
	MultiTypes []IChartMulti
	Opts       *ChartOptions
	Data       []ChartData
	RID        int
	Target     string
	GlobalID   int
	FileName   string
}

// SlideRelMediaSize is the natural size of an SVG/PNG media.
type SlideRelMediaSize struct {
	W float64
	H float64
}

// SlideRelMedia is a media relationship. `ISlideRelMedia`.
type SlideRelMedia struct {
	Type        string
	Opts        *MediaProps
	Path        string
	Extn        string
	Data        any // string | []byte
	IsDuplicate *bool
	IsSvgPng    *bool
	SvgSize     *SlideRelMediaSize
	RID         int
	Target      string
}

// ---------------------------------------------------------------------------
// Slide objects (internal)
// ---------------------------------------------------------------------------

// SlideObject is a single placed slide object (`ISlideObject`). Type selects
// which of the optional fields are meaningful.
type SlideObject struct {
	Type       SlideObjectType
	Options    *ObjectOptions
	Text       []TextProps
	ArrTabRows [][]TableCell
	ChartRID   int
	Image      string
	ImageRID   int
	Hyperlink  *HyperlinkProps
	Media      string
	Mtype      MediaType
	MediaRID   int
	Shape      ShapeType
}

// ObjectOptions is the flattened per-object option bag (`ObjectOptions`),
// combining image/shape/table-cell/text option surfaces.
type ObjectOptions struct {
	PositionProps
	TextBaseProps
	DataOrPathProps
	ObjectNameProps

	// internal:
	PlaceholderIdx  int
	PlaceholderType PlaceholderType

	// image:
	AltText     string
	FlipH       *bool
	FlipV       *bool
	Hyperlink   *HyperlinkProps
	Placeholder string
	Rotate      float64
	Rounding    *bool
	Shadow      *ShadowProps
	Sizing      *ImageSizing

	// shape:
	AngleRange        *[2]float64
	ArcThicknessRatio float64
	Fill              *ShapeFillProps
	Line              *ShapeLineProps
	Points            []ShapePoint
	RectRadius        float64
	LineSize          float64
	LineDash          string
	LineHead          string
	LineTail          string
	ShapeName         string

	// table cell:
	AutoPageCharWeight float64
	AutoPageLineWeight float64
	Border             []BorderProps
	Colspan            int
	Rowspan            int
	Margin             Margin
	ColW               []float64
	RowH               []float64

	// text:
	BodyProp            *BodyProps
	LineIdx             int
	Baseline            float64
	CharSpacing         float64
	Fit                 string
	Glow                *TextGlowProps
	IndentLevel         int
	IsTextBox           *bool
	LineSpacing         float64
	LineSpacingMultiple float64
	Outline             *OutlineProps
	ParaSpaceAfter      float64
	ParaSpaceBefore     float64
	Shape               ShapeType
	Strike              string
	Subscript           *bool
	Superscript         *bool
	Vert                string
	Wrap                *bool
	AutoFit             *bool
	ShrinkText          *bool
	Inset               float64
	RtlMode             *bool

	// position aliases used by tables:
	Cx *Coord
	Cy *Coord
}

// ---------------------------------------------------------------------------
// Write / section / layout / master / presentation
// ---------------------------------------------------------------------------

// WriteBaseProps is shared export options.
type WriteBaseProps struct {
	Compression *bool
}

// WriteProps are options for writing to an in-memory buffer.
type WriteProps struct {
	WriteBaseProps
	OutputType OutputType
}

// WriteFileProps are options for writing to a file.
type WriteFileProps struct {
	WriteBaseProps
	FileName string
}

// SectionProps groups slides into a named section.
type SectionProps struct {
	Type   string // `_type`: "user" | "default"
	Slides []*PresSlide
	Title  string
	Order  int
}

// PresLayout is a presentation layout (name + size in EMU).
type PresLayout struct {
	SizeW  int // `_sizeW`
	SizeH  int // `_sizeH`
	Name   string
	Width  int
	Height int
}

// SlideNumberProps positions/styles the slide number.
type SlideNumberProps struct {
	PositionProps
	TextBaseProps
	Margin Margin
}

// BackgroundProps sets a slide/master background (color, fill, or image).
type BackgroundProps struct {
	DataOrPathProps
	ShapeFillProps
	// Deprecated (v3.6.0):
	Fill HexColor
	Src  string
}

// SlideMasterObject is one object placed on a slide master (union of kinds).
type SlideMasterObject struct {
	Chart       *ChartOptions
	Image       *ImageProps
	Line        *ShapeProps
	Rect        *ShapeProps
	Text        *TextProps
	Placeholder *SlideMasterPlaceholder
}

// SlideMasterPlaceholder is a placeholder object on a slide master.
type SlideMasterPlaceholder struct {
	Options PlaceholderProps
	Text    string
}

// SlideMasterProps defines a slide master.
type SlideMasterProps struct {
	Title       string
	Background  *BackgroundProps
	Margin      Margin
	SlideNumber *SlideNumberProps
	Objects     []SlideMasterObject
	// Deprecated (v3.3.0): string | BackgroundProps
	Bkgd any
}

// SlideBaseProps are fields common to layouts and slides.
type SlideBaseProps struct {
	BkgdImgRid       int               // `_bkgdImgRid`
	Margin           Margin            // `_margin`
	Name             string            // `_name`
	PresLayout       PresLayout        // `_presLayout`
	Rels             []SlideRel        // `_rels`
	RelsChart        []SlideRelChart   // `_relsChart`
	RelsMedia        []SlideRelMedia   // `_relsMedia`
	SlideNum         int               // `_slideNum`
	SlideNumberProps *SlideNumberProps // `_slideNumberProps`
	SlideObjects     []SlideObject     // `_slideObjects`

	Background *BackgroundProps
	// Deprecated (v3.3.0): string | BackgroundProps
	Bkgd any
}

// SlideLayoutInner holds a SlideLayout's private `_slide` metadata.
type SlideLayoutInner struct {
	BkgdImgRid int // `_bkgdImgRid`
	Back       string
	Color      string
	Hidden     *bool
}

// SlideLayout is a slide layout (`SlideLayout`).
type SlideLayout struct {
	SlideBaseProps
	Slide *SlideLayoutInner // `_slide`
}

// PresSlide is the pure-data model of a slide (`PresSlide`). User-facing
// methods (AddText, AddChart, …) live on the Slide type in slide.go.
type PresSlide struct {
	SlideBaseProps
	RID         int         // `_rId`
	SlideLayout SlideLayout // `_slideLayout`
	SlideID     int         // `_slideId`

	Background  *BackgroundProps
	Color       HexColor
	Hidden      *bool
	SlideNumber *SlideNumberProps
}

// AddSlideProps are options for adding a slide.
type AddSlideProps struct {
	MasterName   string
	SectionTitle string
}

// PresentationProps is the presentation-level configuration (`PresentationProps`).
type PresentationProps struct {
	Author      string
	Company     string
	Layout      string
	MasterSlide *PresSlide
	PresLayout  PresLayout
	Revision    string
	RtlMode     bool
	Subject     string
	Theme       ThemeProps
	Title       string
}

// IPresentationProps extends PresentationProps with the full slide model.
type IPresentationProps struct {
	PresentationProps
	Sections     []SectionProps
	SlideLayouts []SlideLayout
	Slides       []PresSlide
}
