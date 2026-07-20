package pptx

import (
	"strings"
	"testing"
)

// newTestSlide builds a minimal PresSlide with a 16x9 layout for object tests.
func newTestSlide() *PresSlide {
	s := &PresSlide{}
	s.PresLayout = PresLayout{Name: "screen16x9", Width: 9144000, Height: 5143500, SizeW: 9144000, SizeH: 5143500}
	s.SlideNum = 1
	return s
}

// ---------------------------------------------------------------------------
// addTextDefinition
// ---------------------------------------------------------------------------

func TestAddTextDefinition_Defaults(t *testing.T) {
	s := newTestSlide()
	addTextDefinition(s, []TextProps{{Text: "hello"}}, &ObjectOptions{}, false)

	if len(s.SlideObjects) != 1 {
		t.Fatalf("want 1 slide object, got %d", len(s.SlideObjects))
	}
	obj := s.SlideObjects[0]
	if obj.Type != SlideObjectTypeText {
		t.Errorf("Type = %q, want text", obj.Type)
	}
	if obj.Shape != ShapeTypeRect {
		t.Errorf("Shape = %q, want rect", obj.Shape)
	}
	if obj.Options.Color != DEF_FONT_COLOR {
		t.Errorf("Color = %q, want %q", obj.Options.Color, DEF_FONT_COLOR)
	}
	if obj.Options.ObjectName != "Text 0" {
		t.Errorf("ObjectName = %q, want %q", obj.Options.ObjectName, "Text 0")
	}
	if obj.Options.BodyProp == nil || obj.Options.BodyProp.Anchor != TextVAlignCTR {
		t.Errorf("BodyProp.Anchor = %+v, want ctr", obj.Options.BodyProp)
	}
	if obj.Options.BodyProp.Wrap == nil || !*obj.Options.BodyProp.Wrap {
		t.Errorf("BodyProp.Wrap should default true")
	}
	// text run options should also be cleaned + default color
	if len(obj.Text) != 1 || obj.Text[0].Options == nil {
		t.Fatalf("text run options missing")
	}
	if obj.Text[0].Options.Color != DEF_FONT_COLOR {
		t.Errorf("run Color = %q, want %q", obj.Text[0].Options.Color, DEF_FONT_COLOR)
	}
}

func TestAddTextDefinition_ColorInheritsFromSlide(t *testing.T) {
	s := newTestSlide()
	s.Color = "ABCDEF"
	addTextDefinition(s, []TextProps{{Text: "x"}}, &ObjectOptions{}, false)
	if got := s.SlideObjects[0].Options.Color; got != "ABCDEF" {
		t.Errorf("Color = %q, want ABCDEF (inherited from slide)", got)
	}
}

func TestAddTextDefinition_EmptyTextGetsPlaceholderRun(t *testing.T) {
	s := newTestSlide()
	addTextDefinition(s, nil, &ObjectOptions{}, false)
	obj := s.SlideObjects[0]
	if len(obj.Text) != 1 || obj.Text[0].Text != "" {
		t.Errorf("empty text should yield one empty run, got %+v", obj.Text)
	}
}

func TestAddTextDefinition_Placeholder(t *testing.T) {
	s := newTestSlide()
	addTextDefinition(s, []TextProps{{Text: ""}}, &ObjectOptions{Placeholder: "title"}, true)
	obj := s.SlideObjects[0]
	if obj.Type != SlideObjectTypePlaceholder {
		t.Errorf("Type = %q, want placeholder", obj.Type)
	}
	// placeholders must NOT get a defaulted color (they inherit)
	if obj.Options.Color != "" {
		t.Errorf("placeholder Color = %q, want empty (inherit)", obj.Options.Color)
	}
	// placeholder anchor is cleared (null), not ctr
	if obj.Options.BodyProp.Anchor != "" {
		t.Errorf("placeholder Anchor = %q, want empty", obj.Options.BodyProp.Anchor)
	}
}

// M2/F: a placeholder's explicit ParaSpaceBefore:0 must win over a run's 12
// (TS spread `{...itemOpts, ...placeHold.options}` copies the explicit 0).
func TestPlaceholderMerge_ParaSpaceBeforeExplicitZeroWins(t *testing.T) {
	s := newTestSlide()
	// Register a body placeholder on the layout whose stored options set
	// ParaSpaceBefore to an explicit 0.
	s.SlideLayout.SlideObjects = []SlideObject{{
		Type: SlideObjectTypePlaceholder,
		Options: &ObjectOptions{
			Placeholder:     "body",
			ParaSpaceBefore: ptr(0.0),
		},
	}}
	// Add text targeting that placeholder with its own ParaSpaceBefore:12.
	addTextDefinition(s, []TextProps{{Text: "hi"}}, &ObjectOptions{
		Placeholder:     "body",
		ParaSpaceBefore: ptr(12.0),
	}, false)

	got := s.SlideObjects[0].Options.ParaSpaceBefore
	if got == nil || *got != 0 {
		t.Fatalf("placeholder ParaSpaceBefore:0 should win over run's 12; got %v", got)
	}
}

func TestAddTextDefinition_AlignValign(t *testing.T) {
	s := newTestSlide()
	addTextDefinition(s, []TextProps{{Text: "x"}}, &ObjectOptions{
		TextBaseProps: TextBaseProps{Align: "right", Valign: "bottom"},
	}, false)
	bp := s.SlideObjects[0].Options.BodyProp
	if bp.Align != TextHAlignRight {
		t.Errorf("Align = %q, want right", bp.Align)
	}
	if bp.Anchor != TextVAlignB {
		t.Errorf("Anchor = %q, want b", bp.Anchor)
	}
}

func TestAddTextDefinition_HyperlinkRel(t *testing.T) {
	s := newTestSlide()
	addTextDefinition(s, []TextProps{
		{Text: "click", Options: &TextPropsOptions{Hyperlink: &HyperlinkProps{URL: "https://example.com"}}},
	}, &ObjectOptions{}, false)
	if len(s.Rels) != 1 {
		t.Fatalf("want 1 hyperlink rel, got %d", len(s.Rels))
	}
	if s.Rels[0].Type != SlideObjectTypeHyperlink || s.Rels[0].Target != "https://example.com" {
		t.Errorf("rel = %+v", s.Rels[0])
	}
}

// ---------------------------------------------------------------------------
// addShapeDefinition
// ---------------------------------------------------------------------------

func TestAddShapeDefinition_RequiresShapeName(t *testing.T) {
	s := newTestSlide()
	if err := addShapeDefinition(s, "", &ShapeProps{}); err == nil {
		t.Fatal("expected error for missing shape name")
	}
	if len(s.SlideObjects) != 0 {
		t.Errorf("no object should be added on error")
	}
}

func TestAddShapeDefinition_Defaults(t *testing.T) {
	s := newTestSlide()
	if err := addShapeDefinition(s, ShapeTypeRect, &ShapeProps{}); err != nil {
		t.Fatal(err)
	}
	obj := s.SlideObjects[0]
	if obj.Type != SlideObjectTypeText { // shapes are stored as text type
		t.Errorf("Type = %q, want text", obj.Type)
	}
	if obj.Shape != ShapeTypeRect {
		t.Errorf("Shape = %q", obj.Shape)
	}
	for _, c := range []*Coord{obj.Options.X, obj.Options.Y, obj.Options.W, obj.Options.H} {
		if c == nil || c.Val != 1 {
			t.Errorf("default x/y/w/h should be Inches(1), got %+v", c)
		}
	}
	if obj.Options.ObjectName != "Shape 0" {
		t.Errorf("ObjectName = %q, want Shape 0", obj.Options.ObjectName)
	}
	// no explicit line -> line stays type "none"
	if obj.Options.Line == nil || obj.Options.Line.Type != "none" {
		t.Errorf("default line = %+v, want type none", obj.Options.Line)
	}
}

func TestAddShapeDefinition_LineColorDefault(t *testing.T) {
	s := newTestSlide()
	addShapeDefinition(s, ShapeTypeRect, &ShapeProps{
		Line: &ShapeLineProps{ShapeFillProps: ShapeFillProps{Type: "solid"}},
	})
	line := s.SlideObjects[0].Options.Line
	if line.Color != DEF_SHAPE_LINE_COLOR {
		t.Errorf("line Color = %q, want %q", line.Color, DEF_SHAPE_LINE_COLOR)
	}
	if line.Width != 1 || line.DashType != "solid" {
		t.Errorf("line defaults wrong: %+v", line)
	}
}

func TestAddShapeDefinition_DeprecatedLineSize(t *testing.T) {
	s := newTestSlide()
	addShapeDefinition(s, ShapeTypeRect, &ShapeProps{
		Line:     &ShapeLineProps{ShapeFillProps: ShapeFillProps{Type: "solid"}},
		LineSize: 4,
	})
	if w := s.SlideObjects[0].Options.Line.Width; w != 4 {
		t.Errorf("deprecated lineSize should set width=4, got %v", w)
	}
}

func TestAddShapeDefinition_HyperlinkRel(t *testing.T) {
	s := newTestSlide()
	addShapeDefinition(s, ShapeTypeRect, &ShapeProps{Hyperlink: &HyperlinkProps{Slide: 3}})
	if len(s.Rels) != 1 || s.Rels[0].Data != "slide" || s.Rels[0].Target != "3" {
		t.Errorf("shape slide-hyperlink rel wrong: %+v", s.Rels)
	}
}

// ---------------------------------------------------------------------------
// addImageDefinition
// ---------------------------------------------------------------------------

func TestAddImageDefinition_RequiresDataOrPath(t *testing.T) {
	s := newTestSlide()
	if err := addImageDefinition(s, &ImageProps{}); err == nil {
		t.Fatal("expected error when neither data nor path given")
	}
}

func TestAddImageDefinition_ExtnFromPath(t *testing.T) {
	s := newTestSlide()
	if err := addImageDefinition(s, &ImageProps{DataOrPathProps: DataOrPathProps{Path: "/img/photo.JPG?x=1"}}); err != nil {
		t.Fatal(err)
	}
	if len(s.RelsMedia) != 1 {
		t.Fatalf("want 1 media rel, got %d", len(s.RelsMedia))
	}
	if s.RelsMedia[0].Extn != "jpg" || s.RelsMedia[0].Type != "image/jpg" {
		t.Errorf("extn sniff wrong: %+v", s.RelsMedia[0])
	}
	obj := s.SlideObjects[0]
	if obj.Type != SlideObjectTypeImage || obj.Image != "/img/photo.JPG?x=1" {
		t.Errorf("image object wrong: %+v", obj)
	}
	if obj.Options.W == nil || obj.Options.W.Val != 1 {
		t.Errorf("default width should be 1, got %+v", obj.Options.W)
	}
	if obj.Options.ObjectName != "Image 0" {
		t.Errorf("ObjectName = %q", obj.Options.ObjectName)
	}
}

func TestAddImageDefinition_ExtnFromData(t *testing.T) {
	s := newTestSlide()
	addImageDefinition(s, &ImageProps{DataOrPathProps: DataOrPathProps{Data: "image/png;base64,AAAA"}})
	if s.RelsMedia[0].Extn != "png" {
		t.Errorf("data extn sniff = %q, want png", s.RelsMedia[0].Extn)
	}
	if s.SlideObjects[0].Image != "preencoded.png" {
		t.Errorf("image path fallback = %q", s.SlideObjects[0].Image)
	}
}

func TestAddImageDefinition_SvgTwoRels(t *testing.T) {
	s := newTestSlide()
	addImageDefinition(s, &ImageProps{DataOrPathProps: DataOrPathProps{Data: "data:image/svg+xml;base64,AAAA"}})
	if len(s.RelsMedia) != 2 {
		t.Fatalf("svg should create 2 media rels, got %d", len(s.RelsMedia))
	}
	if s.RelsMedia[0].Type != "image/png" || s.RelsMedia[0].IsSvgPng == nil || !*s.RelsMedia[0].IsSvgPng {
		t.Errorf("first svg rel should be png fallback: %+v", s.RelsMedia[0])
	}
	if s.RelsMedia[1].Type != "image/svg+xml" {
		t.Errorf("second svg rel should be svg: %+v", s.RelsMedia[1])
	}
	if s.SlideObjects[0].ImageRID != 2 {
		t.Errorf("ImageRID = %d, want 2 (svg image)", s.SlideObjects[0].ImageRID)
	}
}

func TestAddImageDefinition_Hyperlink(t *testing.T) {
	s := newTestSlide()
	addImageDefinition(s, &ImageProps{
		DataOrPathProps: DataOrPathProps{Path: "/a.png"},
		Hyperlink:       &HyperlinkProps{URL: "https://x.io"},
	})
	if len(s.Rels) != 1 || s.Rels[0].Type != SlideObjectTypeHyperlink {
		t.Fatalf("want 1 hyperlink rel, got %+v", s.Rels)
	}
	// image rel is rId1, hyperlink rel is rId2
	if s.Rels[0].RID != 2 {
		t.Errorf("hyperlink RID = %d, want 2", s.Rels[0].RID)
	}
	if s.SlideObjects[0].Hyperlink == nil || s.SlideObjects[0].Hyperlink.RID != 2 {
		t.Errorf("image object hyperlink not wired: %+v", s.SlideObjects[0].Hyperlink)
	}
}

func TestAddImageDefinition_HyperlinkRequiresTarget(t *testing.T) {
	s := newTestSlide()
	err := addImageDefinition(s, &ImageProps{
		DataOrPathProps: DataOrPathProps{Path: "/a.png"},
		Hyperlink:       &HyperlinkProps{},
	})
	if err == nil {
		t.Fatal("expected error for hyperlink without url/slide")
	}
}

func TestAddImageDefinition_DuplicateReusesTarget(t *testing.T) {
	s := newTestSlide()
	addImageDefinition(s, &ImageProps{DataOrPathProps: DataOrPathProps{Path: "/same.png"}})
	addImageDefinition(s, &ImageProps{DataOrPathProps: DataOrPathProps{Path: "/same.png"}})
	if len(s.RelsMedia) != 2 {
		t.Fatalf("want 2 media rels, got %d", len(s.RelsMedia))
	}
	if s.RelsMedia[1].IsDuplicate == nil || !*s.RelsMedia[1].IsDuplicate {
		t.Errorf("second image should be marked duplicate")
	}
	if s.RelsMedia[0].Target != s.RelsMedia[1].Target {
		t.Errorf("duplicate should reuse Target: %q vs %q", s.RelsMedia[0].Target, s.RelsMedia[1].Target)
	}
}

// ---------------------------------------------------------------------------
// addChartDefinition
// ---------------------------------------------------------------------------

func TestAddChartDefinition_Defaults(t *testing.T) {
	s := newTestSlide()
	obj, err := addChartDefinition(s, &chartCounter{}, ChartTypeBar, nil,
		[]ChartData{{Name: "S1", Values: []float64{1, 2, 3}}},
		&ChartOptions{Bar3DShape: "invalid"})
	if err != nil {
		t.Fatal(err)
	}
	if obj.Type != SlideObjectTypeChart || obj.ChartRID != 1 {
		t.Errorf("chart object wrong: %+v", obj)
	}
	if len(s.RelsChart) != 1 {
		t.Fatalf("want 1 chart rel, got %d", len(s.RelsChart))
	}
	rel := s.RelsChart[0]
	o := rel.Opts
	if o.Bar3DShape != "box" {
		t.Errorf("bar3DShape whitelist: got %q, want box", o.Bar3DShape)
	}
	if o.BarDir != "col" {
		t.Errorf("barDir default = %q, want col", o.BarDir)
	}
	if o.LegendPos != "r" {
		t.Errorf("legendPos default = %q, want r", o.LegendPos)
	}
	if fptrOr(o.BarGapWidthPct, 0) != 150 {
		t.Errorf("barGapWidthPct default = %v, want 150", o.BarGapWidthPct)
	}
	if len(o.ChartColors) == 0 || o.ChartColors[0] != BARCHART_COLORS[0] {
		t.Errorf("chartColors should default to BARCHART_COLORS")
	}
	if o.ObjectName != "Chart 0" {
		t.Errorf("ObjectName = %q", o.ObjectName)
	}
	if o.DataLabelFormatCode != "#,##0" {
		t.Errorf("dataLabelFormatCode default = %q", o.DataLabelFormatCode)
	}
	if rel.Type != ChartTypeBar {
		t.Errorf("rel.Type = %q", rel.Type)
	}
	if len(rel.Data) != 1 || rel.Data[0].DataIndex != 0 {
		t.Errorf("chart data not indexed: %+v", rel.Data)
	}
}

func TestAddChartDefinition_BarGapWidthBounds(t *testing.T) {
	s := newTestSlide()
	_, _ = addChartDefinition(s, &chartCounter{}, ChartTypeBar, nil,
		[]ChartData{{Values: []float64{1}}}, &ChartOptions{BarGapWidthPct: ptr(5000.0)})
	if got := fptrOr(s.RelsChart[0].Opts.BarGapWidthPct, 0); got != 150 {
		t.Errorf("out-of-range barGapWidthPct should reset to 150, got %v", got)
	}

	s2 := newTestSlide()
	_, _ = addChartDefinition(s2, &chartCounter{}, ChartTypeBar, nil,
		[]ChartData{{Values: []float64{1}}}, &ChartOptions{BarGapWidthPct: ptr(300.0)})
	if got := fptrOr(s2.RelsChart[0].Opts.BarGapWidthPct, 0); got != 300 {
		t.Errorf("in-range barGapWidthPct should be kept, got %v", got)
	}
}

func TestAddChartDefinition_DataLabelPosWhitelist(t *testing.T) {
	s := newTestSlide()
	_, _ = addChartDefinition(s, &chartCounter{}, ChartTypePie, nil,
		[]ChartData{{Values: []float64{1}}}, &ChartOptions{DataLabelPosition: "bogus"})
	if got := s.RelsChart[0].Opts.DataLabelPosition; got != "" {
		t.Errorf("invalid pie dataLabelPosition should be cleared, got %q", got)
	}

	s2 := newTestSlide()
	_, _ = addChartDefinition(s2, &chartCounter{}, ChartTypePie, nil,
		[]ChartData{{Values: []float64{1}}}, &ChartOptions{DataLabelPosition: "ctr"})
	if got := s2.RelsChart[0].Opts.DataLabelPosition; got != "ctr" {
		t.Errorf("valid pie dataLabelPosition should be kept, got %q", got)
	}
}

func TestAddChartDefinition_PieColorsAndFormat(t *testing.T) {
	s := newTestSlide()
	_, _ = addChartDefinition(s, &chartCounter{}, ChartTypePie, nil,
		[]ChartData{{Values: []float64{1}}}, &ChartOptions{ShowPercent: ptr(true)})
	o := s.RelsChart[0].Opts
	if o.ChartColors[0] != PIECHART_COLORS[0] {
		t.Errorf("pie should default to PIECHART_COLORS")
	}
	if o.DataLabelFormatCode != "0%" {
		t.Errorf("pie+showPercent dataLabelFormatCode = %q, want 0%%", o.DataLabelFormatCode)
	}
}

func TestAddChartDefinition_MultiType(t *testing.T) {
	s := newTestSlide()
	multi := []IChartMulti{
		{Type: ChartTypeBar, Data: []ChartData{{Name: "A", Values: []float64{1}}}},
		{Type: ChartTypeLine, Data: []ChartData{{Name: "B", Values: []float64{2}}}},
	}
	_, err := addChartDefinition(s, &chartCounter{}, "", multi, nil, &ChartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rel := s.RelsChart[0]
	if len(rel.MultiTypes) != 2 {
		t.Errorf("multi-chart rel should carry MultiTypes, got %+v", rel.MultiTypes)
	}
	if len(rel.Data) != 2 {
		t.Errorf("multi-chart data should be concatenated, got %d", len(rel.Data))
	}
}

// ---------------------------------------------------------------------------
// addTableDefinition
// ---------------------------------------------------------------------------

func TestAddTableDefinition_EmptyErrors(t *testing.T) {
	s := newTestSlide()
	if _, err := addTableDefinition(s, nil, nil, nil, s.PresLayout, nil, nil); err == nil {
		t.Fatal("expected error for empty rows")
	}
}

func TestAddTableDefinition_Basic(t *testing.T) {
	s := newTestSlide()
	rows := []TableRow{
		{{Text: "A"}, {Text: "B"}},
		{{Text: "C"}, {Text: "D"}},
	}
	_, err := addTableDefinition(s, rows, nil, nil, s.PresLayout, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.SlideObjects) != 1 || s.SlideObjects[0].Type != SlideObjectTypeTable {
		t.Fatalf("want 1 table object, got %+v", s.SlideObjects)
	}
	obj := s.SlideObjects[0]
	if len(obj.ArrTabRows) != 2 || len(obj.ArrTabRows[0]) != 2 {
		t.Errorf("table rows shape wrong: %d rows", len(obj.ArrTabRows))
	}
	// default color applied (no hyperlinks)
	if obj.Options.Color != DEF_FONT_COLOR {
		t.Errorf("table color = %q, want %q", obj.Options.Color, DEF_FONT_COLOR)
	}
	// default width: floor(10 - 0.5 - 0.5)=9in -> EMU
	if obj.Options.W == nil || obj.Options.W.Val != float64(inch2Emu(9)) {
		t.Errorf("default table width = %+v, want %d EMU", obj.Options.W, inch2Emu(9))
	}
	// x default EMU/2 = 457200
	if obj.Options.X == nil || obj.Options.X.Val != EMU/2 {
		t.Errorf("default x = %+v, want %d", obj.Options.X, EMU/2)
	}
	// cell border defaults
	b := obj.ArrTabRows[0][0].Options.Border
	if len(b) != 4 || b[0].Type != "none" || b[0].Color != DEF_CELL_BORDER.Color || b[0].Pt != DEF_CELL_BORDER.Pt {
		t.Errorf("cell border defaults wrong: %+v", b)
	}
	if obj.Options.ObjectName != "Table 0" {
		t.Errorf("ObjectName = %q", obj.Options.ObjectName)
	}
}

func TestAddTableDefinition_ColWSingleValue(t *testing.T) {
	s := newTestSlide()
	rows := []TableRow{{{Text: "A"}, {Text: "B"}}}
	opt := &TableProps{ColW: []float64{3}}
	_, err := addTableDefinition(s, rows, opt, nil, s.PresLayout, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	obj := s.SlideObjects[0]
	// colW=[3] with 2 cols -> w = floor(3*2)=6 in -> EMU, colW cleared
	if obj.Options.W == nil || obj.Options.W.Val != float64(inch2Emu(6)) {
		t.Errorf("colW single -> width = %+v, want %d EMU", obj.Options.W, inch2Emu(6))
	}
	if obj.Options.ColW != nil {
		t.Errorf("colW should be cleared, got %+v", obj.Options.ColW)
	}
}

func TestAddTableDefinition_ColWMatching(t *testing.T) {
	s := newTestSlide()
	rows := []TableRow{{{Text: "A"}, {Text: "B"}}}
	opt := &TableProps{ColW: []float64{2, 3}}
	_, _ = addTableDefinition(s, rows, opt, nil, s.PresLayout, nil, nil)
	obj := s.SlideObjects[0]
	if len(obj.Options.ColW) != 2 {
		t.Errorf("matching colW should be preserved, got %+v", obj.Options.ColW)
	}
}

// M8: a mis-wired getSlide callback (always returns nil) must produce an error
// naming the missing slide(s), not silently drop rows.
func TestAddTableDefinition_AutoPageMissingSlideErrors(t *testing.T) {
	s := newTestSlide()
	rows := manyShortRows(100) // enough to overflow into multiple slides
	opt := &TableProps{AutoPage: ptr(true)}
	addSlide := func(_ *AddSlideProps) *PresSlide { return &PresSlide{} }
	getSlide := func(int) *PresSlide { return nil } // mis-wired: never delivers

	_, err := addTableDefinition(s, rows, opt, &s.SlideLayout, s.PresLayout, addSlide, getSlide)
	if err == nil {
		t.Fatal("mis-wired getSlide should return an error, not silent success")
	}
	if !strings.Contains(err.Error(), "getSlide returned nil") {
		t.Errorf("error should name the missing-slide condition, got: %v", err)
	}
}

// Minor (K): combo-chart multi-axis validation throws are surfaced as errors.
func TestValidateChartConfig_SecondaryAxisRequired(t *testing.T) {
	s := newTestSlide()
	data := []ChartData{{Values: []float64{1}, Labels: [][]string{{"A"}}}}
	_, err := addChartDefinition(s, &chartCounter{}, ChartTypeBar, nil, data, &ChartOptions{
		ValAxes: []ChartOptions{{}, {}}, // 2 value axes, none secondary
	})
	if err == nil || !strings.Contains(err.Error(), "secondary axis must be used") {
		t.Fatalf("expected secondary-axis error, got: %v", err)
	}
}

func TestValidateChartConfig_AxesCountMismatch(t *testing.T) {
	s := newTestSlide()
	data := []ChartData{{Values: []float64{1}, Labels: [][]string{{"A"}}}}
	_, err := addChartDefinition(s, &chartCounter{}, ChartTypeBar, nil, data, &ChartOptions{
		CatAxes: []ChartOptions{{}}, // 1 category axis, 0 value axes
	})
	if err == nil || !strings.Contains(err.Error(), "same number of value and category axes") {
		t.Fatalf("expected axes-count error, got: %v", err)
	}
}

func TestAddTableDefinition_HyperlinkSkipsDefaultColor(t *testing.T) {
	s := newTestSlide()
	rows := []TableRow{{{Text: "A", Options: &TableCellProps{Hyperlink: &HyperlinkProps{URL: "https://x"}}}}}
	_, _ = addTableDefinition(s, rows, nil, nil, s.PresLayout, nil, nil)
	if s.SlideObjects[0].Options.Color != "" {
		t.Errorf("table with hyperlink should not get default color, got %q", s.SlideObjects[0].Options.Color)
	}
	if len(s.Rels) != 1 {
		t.Errorf("cell hyperlink rel should be registered, got %d rels", len(s.Rels))
	}
}

// ---------------------------------------------------------------------------
// addMediaDefinition
// ---------------------------------------------------------------------------

func TestAddMediaDefinition_OnlineRequiresLink(t *testing.T) {
	s := newTestSlide()
	if err := addMediaDefinition(s, &MediaProps{Type: "online"}); err == nil {
		t.Fatal("expected error: online requires link")
	}
}

func TestAddMediaDefinition_Online(t *testing.T) {
	s := newTestSlide()
	err := addMediaDefinition(s, &MediaProps{Type: "online", Link: "https://youtu.be/x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.RelsMedia) != 2 {
		t.Fatalf("online media -> 2 rels (video+cover), got %d", len(s.RelsMedia))
	}
	if s.RelsMedia[0].Type != "online" || s.RelsMedia[0].Target != "https://youtu.be/x" {
		t.Errorf("online video rel wrong: %+v", s.RelsMedia[0])
	}
	obj := s.SlideObjects[0]
	if obj.Type != SlideObjectTypeMedia || obj.Mtype != "online" || obj.MediaRID != 1 {
		t.Errorf("media object wrong: %+v", obj)
	}
}

func TestAddMediaDefinition_AudioPath(t *testing.T) {
	s := newTestSlide()
	err := addMediaDefinition(s, &MediaProps{Type: "audio", DataOrPathProps: DataOrPathProps{Path: "/sound.mp3"}})
	if err != nil {
		t.Fatal(err)
	}
	// audio -> video rel + media rel + cover image rel
	if len(s.RelsMedia) != 3 {
		t.Fatalf("audio -> 3 rels, got %d", len(s.RelsMedia))
	}
	if s.RelsMedia[0].Type != "audio/mp3" {
		t.Errorf("audio type = %q, want audio/mp3", s.RelsMedia[0].Type)
	}
	if s.SlideObjects[0].Media != "/sound.mp3" {
		t.Errorf("media path = %q", s.SlideObjects[0].Media)
	}
}

func TestAddMediaDefinition_Base64Check(t *testing.T) {
	s := newTestSlide()
	err := addMediaDefinition(s, &MediaProps{Type: "video", DataOrPathProps: DataOrPathProps{Data: "notbase64"}})
	if err == nil {
		t.Fatal("expected error for data lacking base64 header")
	}
}

// ---------------------------------------------------------------------------
// addNotesDefinition
// ---------------------------------------------------------------------------

func TestAddNotesDefinition(t *testing.T) {
	s := newTestSlide()
	addNotesDefinition(s, "speaker note")
	if len(s.SlideObjects) != 1 || s.SlideObjects[0].Type != SlideObjectTypeNotes {
		t.Fatalf("notes object missing: %+v", s.SlideObjects)
	}
	if s.SlideObjects[0].Text[0].Text != "speaker note" {
		t.Errorf("notes text = %q", s.SlideObjects[0].Text[0].Text)
	}
}

// ---------------------------------------------------------------------------
// addBackgroundDefinition
// ---------------------------------------------------------------------------

func TestAddBackgroundDefinition_Color(t *testing.T) {
	target := &SlideBaseProps{Bkgd: "FF0000"}
	addBackgroundDefinition(nil, target)
	if target.Background == nil || target.Background.Color != "FF0000" {
		t.Errorf("bkgd color not applied: %+v", target.Background)
	}
}

func TestAddBackgroundDefinition_Image(t *testing.T) {
	target := &SlideBaseProps{Name: "Slide 1"}
	props := &BackgroundProps{DataOrPathProps: DataOrPathProps{Path: "bg.png"}}
	addBackgroundDefinition(props, target)
	if len(target.RelsMedia) != 1 {
		t.Fatalf("want 1 bkgd media rel, got %d", len(target.RelsMedia))
	}
	rel := target.RelsMedia[0]
	if rel.Extn != "png" || rel.Type != "image" {
		t.Errorf("bkgd rel wrong: %+v", rel)
	}
	if rel.Target != "../media/Slide-1-image-1.png" {
		t.Errorf("bkgd Target = %q", rel.Target)
	}
	if target.BkgdImgRid != 1 {
		t.Errorf("BkgdImgRid = %d, want 1", target.BkgdImgRid)
	}
}

func TestAddBackgroundDefinition_JpgToJpeg(t *testing.T) {
	target := &SlideBaseProps{Name: "S"}
	addBackgroundDefinition(&BackgroundProps{DataOrPathProps: DataOrPathProps{Path: "pic.jpg"}}, target)
	if target.RelsMedia[0].Extn != "jpeg" {
		t.Errorf("jpg extn should be normalized to jpeg, got %q", target.RelsMedia[0].Extn)
	}
}

// ---------------------------------------------------------------------------
// createSlideMaster
// ---------------------------------------------------------------------------

func TestCreateSlideMaster(t *testing.T) {
	target := &SlideLayout{}
	target.PresLayout = PresLayout{Width: 9144000, Height: 5143500, SizeW: 9144000, SizeH: 5143500}
	props := &SlideMasterProps{
		Objects: []SlideMasterObject{
			{Text: &TextProps{Text: "Title", Options: &TextPropsOptions{}}},
			{Placeholder: &SlideMasterPlaceholder{
				Options: PlaceholderProps{Name: "body", Type: PlaceholderTypeBody},
				Text:    "",
			}},
		},
		SlideNumber: &SlideNumberProps{},
	}
	if err := createSlideMaster(props, &chartCounter{}, target); err != nil {
		t.Fatal(err)
	}
	if len(target.SlideObjects) != 2 {
		t.Fatalf("want 2 master objects, got %d", len(target.SlideObjects))
	}
	if target.SlideObjects[0].Type != SlideObjectTypeText {
		t.Errorf("first master object should be text, got %q", target.SlideObjects[0].Type)
	}
	ph := target.SlideObjects[1]
	if ph.Type != SlideObjectTypePlaceholder {
		t.Errorf("second should be placeholder, got %q", ph.Type)
	}
	if ph.Options.Placeholder != "body" {
		t.Errorf("placeholder name = %q, want body", ph.Options.Placeholder)
	}
	if ph.Options.PlaceholderType != PlaceholderTypeBody {
		t.Errorf("placeholder type = %q, want body", ph.Options.PlaceholderType)
	}
	if ph.Options.PlaceholderIdx != 101 {
		t.Errorf("placeholder idx = %d, want 101 (100+idx)", ph.Options.PlaceholderIdx)
	}
	if target.SlideNumberProps == nil {
		t.Errorf("slide number props should be registered")
	}
}

// ---------------------------------------------------------------------------
// addPlaceholdersToSlideLayouts
// ---------------------------------------------------------------------------

func TestAddPlaceholdersToSlideLayouts(t *testing.T) {
	s := newTestSlide()
	s.SlideLayout.SlideObjects = []SlideObject{
		{Type: SlideObjectTypePlaceholder, Options: &ObjectOptions{Placeholder: "body"}},
	}
	addPlaceholdersToSlideLayouts(s)
	if len(s.SlideObjects) != 1 {
		t.Fatalf("placeholder should be added to slide, got %d objects", len(s.SlideObjects))
	}
	if s.SlideObjects[0].Options.Placeholder != "body" {
		t.Errorf("added placeholder name = %q", s.SlideObjects[0].Options.Placeholder)
	}
}

func TestAddPlaceholdersToSlideLayouts_SkipsExisting(t *testing.T) {
	s := newTestSlide()
	s.SlideLayout.SlideObjects = []SlideObject{
		{Type: SlideObjectTypePlaceholder, Options: &ObjectOptions{Placeholder: "body"}},
	}
	// already present on the slide
	s.SlideObjects = []SlideObject{
		{Type: SlideObjectTypeText, Options: &ObjectOptions{Placeholder: "body"}},
	}
	addPlaceholdersToSlideLayouts(s)
	if len(s.SlideObjects) != 1 {
		t.Errorf("existing placeholder should not be duplicated, got %d", len(s.SlideObjects))
	}
}
