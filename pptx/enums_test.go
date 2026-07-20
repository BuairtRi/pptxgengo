package pptx

import (
	"strings"
	"testing"
)

func TestMeasurementConstants(t *testing.T) {
	if EMU != 914400 {
		t.Errorf("EMU = %d, want 914400", EMU)
	}
	if ONEPT != 12700 {
		t.Errorf("ONEPT = %d, want 12700", ONEPT)
	}
	if CRLF != "\r\n" {
		t.Errorf("CRLF = %q, want CRLF", CRLF)
	}
	if LINEH_MODIFIER != 1.67 {
		t.Errorf("LINEH_MODIFIER = %v, want 1.67", LINEH_MODIFIER)
	}
	if LAYOUT_IDX_SERIES_BASE != 2147483649 {
		t.Errorf("LAYOUT_IDX_SERIES_BASE = %d", LAYOUT_IDX_SERIES_BASE)
	}
}

func TestDefaultConstants(t *testing.T) {
	if DEF_FONT_COLOR != "000000" {
		t.Errorf("DEF_FONT_COLOR = %q", DEF_FONT_COLOR)
	}
	if DEF_FONT_SIZE != 12 {
		t.Errorf("DEF_FONT_SIZE = %d", DEF_FONT_SIZE)
	}
	if DEF_PRES_LAYOUT != "LAYOUT_16x9" {
		t.Errorf("DEF_PRES_LAYOUT = %q", DEF_PRES_LAYOUT)
	}
	if SLDNUMFLDID != "{F7021451-1387-4CA6-816F-3879F97B5CBC}" {
		t.Errorf("SLDNUMFLDID = %q", SLDNUMFLDID)
	}
}

func TestDefaultComposites(t *testing.T) {
	if DEF_CELL_BORDER.Color != "666666" || DEF_CELL_BORDER.Pt != 1 {
		t.Errorf("DEF_CELL_BORDER = %+v", DEF_CELL_BORDER)
	}
	if DEF_TEXT_GLOW.Size != 8 || DEF_TEXT_GLOW.Color != "FFFFFF" || DEF_TEXT_GLOW.Opacity != 0.75 {
		t.Errorf("DEF_TEXT_GLOW = %+v", DEF_TEXT_GLOW)
	}
	if DEF_SHAPE_SHADOW.RotateWithShape == nil || *DEF_SHAPE_SHADOW.RotateWithShape != true {
		t.Errorf("DEF_SHAPE_SHADOW.RotateWithShape not true")
	}
	// offset = 23000/12700
	if DEF_SHAPE_SHADOW.Offset != 23000.0/12700.0 {
		t.Errorf("DEF_SHAPE_SHADOW.Offset = %v", DEF_SHAPE_SHADOW.Offset)
	}
	if DEF_CELL_MARGIN_IN != [4]float64{0.05, 0.1, 0.05, 0.1} {
		t.Errorf("DEF_CELL_MARGIN_IN = %v", DEF_CELL_MARGIN_IN)
	}
}

func TestColorArrays(t *testing.T) {
	if len(BARCHART_COLORS) != 16 {
		t.Errorf("len(BARCHART_COLORS) = %d, want 16", len(BARCHART_COLORS))
	}
	if BARCHART_COLORS[0] != "C0504D" {
		t.Errorf("BARCHART_COLORS[0] = %q", BARCHART_COLORS[0])
	}
	if len(PIECHART_COLORS) != 18 {
		t.Errorf("len(PIECHART_COLORS) = %d, want 18", len(PIECHART_COLORS))
	}
	if PIECHART_COLORS[0] != "5DA5DA" {
		t.Errorf("PIECHART_COLORS[0] = %q", PIECHART_COLORS[0])
	}
	if len(LETTERS) != 26 || LETTERS[0] != "A" || LETTERS[25] != "Z" {
		t.Errorf("LETTERS malformed")
	}
}

func TestShapeTypeConstants(t *testing.T) {
	if ShapeTypeRect != "rect" {
		t.Errorf("ShapeTypeRect = %q, want rect", ShapeTypeRect)
	}
	if ShapeTypeRoundRect != "roundRect" {
		t.Errorf("ShapeTypeRoundRect = %q", ShapeTypeRoundRect)
	}
	if ShapeTypeWedgeRoundRectCallout != "wedgeRoundRectCallout" {
		t.Errorf("ShapeTypeWedgeRoundRectCallout = %q", ShapeTypeWedgeRoundRectCallout)
	}
	// MSO friendly alias map
	if MsoRectangle != "rect" {
		t.Errorf("MsoRectangle = %q, want rect", MsoRectangle)
	}
	if MsoRoundedRectangle != "roundRect" {
		t.Errorf("MsoRoundedRectangle = %q", MsoRoundedRectangle)
	}
	// buggy value preserved verbatim from source
	if MsoLineCallout4AccentBar != "accentCallout3=4" {
		t.Errorf("MsoLineCallout4AccentBar = %q", MsoLineCallout4AccentBar)
	}
}

func TestChartTypeConstants(t *testing.T) {
	if ChartTypeBar != "bar" {
		t.Errorf("ChartTypeBar = %q", ChartTypeBar)
	}
	if ChartTypeBar3d != "bar3D" {
		t.Errorf("ChartTypeBar3d = %q, want bar3D", ChartTypeBar3d)
	}
	if ChartTypeDoughnut != "doughnut" {
		t.Errorf("ChartTypeDoughnut = %q", ChartTypeDoughnut)
	}
}

func TestSchemeColorConstants(t *testing.T) {
	if SchemeColorText1 != "tx1" {
		t.Errorf("SchemeColorText1 = %q", SchemeColorText1)
	}
	if SchemeColorBackground1 != "bg1" {
		t.Errorf("SchemeColorBackground1 = %q", SchemeColorBackground1)
	}
	if SchemeColorAccent6 != "accent6" {
		t.Errorf("SchemeColorAccent6 = %q", SchemeColorAccent6)
	}
}

func TestAlignConstants(t *testing.T) {
	if AlignHLeft != "left" || AlignHJustify != "justify" {
		t.Errorf("AlignH constants wrong")
	}
	if AlignVMiddle != "middle" {
		t.Errorf("AlignVMiddle = %q", AlignVMiddle)
	}
	if TextVAlignCTR != "ctr" {
		t.Errorf("TextVAlignCTR = %q", TextVAlignCTR)
	}
}

func TestSlideObjectAndPlaceholderTypes(t *testing.T) {
	if SlideObjectTypeChart != "chart" || SlideObjectTypeTablecell != "tablecell" {
		t.Errorf("SlideObjectType constants wrong")
	}
	if PlaceholderTypeImage != "pic" {
		t.Errorf("PlaceholderTypeImage = %q, want pic", PlaceholderTypeImage)
	}
	if PlaceholderTypeTable != "tbl" {
		t.Errorf("PlaceholderTypeTable = %q, want tbl", PlaceholderTypeTable)
	}
}

func TestBulletTypes(t *testing.T) {
	if BulletTypeDefault != "&#x2022;" {
		t.Errorf("BulletTypeDefault = %q", BulletTypeDefault)
	}
	if BulletTypeStar != "&#x2605;" {
		t.Errorf("BulletTypeStar = %q", BulletTypeStar)
	}
}

func TestImageConstants(t *testing.T) {
	if !strings.HasPrefix(IMG_BROKEN, "data:image/png;base64,iVBORw0KG") {
		t.Errorf("IMG_BROKEN prefix wrong: %.40q", IMG_BROKEN)
	}
	if !strings.HasPrefix(IMG_PLAYBTN, "data:image/png;base64,iVBORw0KG") {
		t.Errorf("IMG_PLAYBTN prefix wrong: %.40q", IMG_PLAYBTN)
	}
	if len(IMG_BROKEN) < 1000 || len(IMG_PLAYBTN) < 1000 {
		t.Errorf("IMG constants too short: broken=%d playbtn=%d", len(IMG_BROKEN), len(IMG_PLAYBTN))
	}
}

func TestRegexHexColor(t *testing.T) {
	if !RegexHexColor.MatchString("FF00AA") {
		t.Error("RegexHexColor should match FF00AA")
	}
	if RegexHexColor.MatchString("FF00A") {
		t.Error("RegexHexColor should not match 5-digit")
	}
	if RegexHexColor.MatchString("GGGGGG") {
		t.Error("RegexHexColor should not match non-hex")
	}
}
