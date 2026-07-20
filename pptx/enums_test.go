package pptx

import (
	"crypto/sha256"
	"encoding/hex"
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
	if fptrOr(DEF_TEXT_GLOW.Size, 0) != 8 || DEF_TEXT_GLOW.Color != "FFFFFF" || fptrOr(DEF_TEXT_GLOW.Opacity, 0) != 0.75 {
		t.Errorf("DEF_TEXT_GLOW = %+v", DEF_TEXT_GLOW)
	}
	if DEF_SHAPE_SHADOW.RotateWithShape == nil || *DEF_SHAPE_SHADOW.RotateWithShape != true {
		t.Errorf("DEF_SHAPE_SHADOW.RotateWithShape not true")
	}
	// offset = 23000/12700
	if fptrOr(DEF_SHAPE_SHADOW.Offset, 0) != 23000.0/12700.0 {
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

// TestImageConstants freezes IMG_BROKEN/IMG_PLAYBTN as full SHA-256 hashes
// of the complete base64 data-URI strings (M11: a 40-char prefix check
// would not catch mid-blob corruption).
//
// The expected hashes were computed by independently re-extracting the
// literal string constants from src/core-enums.ts (the `export const
// IMG_BROKEN = '...'` / `IMG_PLAYBTN` string literals) and hashing them
// with Python's hashlib.sha256 — NOT by hashing whatever ended up in this
// Go source, so a match here certifies the Go constant against the TS
// source of truth, not just against itself:
//
//	IMG_BROKEN  len=2150  sha256=3c23bca3209cf22899e8a09dc3e82b6707c8ed54e28dca77338e7799d45f1ae8
//	IMG_PLAYBTN len=74402 sha256=60781f4b64d3afa871fc59d2ff49999d5bc8ab0f47c77efbff489e0cb404847a
func TestImageConstants(t *testing.T) {
	const wantBrokenSHA256 = "3c23bca3209cf22899e8a09dc3e82b6707c8ed54e28dca77338e7799d45f1ae8"
	const wantPlaybtnSHA256 = "60781f4b64d3afa871fc59d2ff49999d5bc8ab0f47c77efbff489e0cb404847a"

	if got := sha256Hex(IMG_BROKEN); got != wantBrokenSHA256 {
		t.Errorf("IMG_BROKEN sha256 = %s, want %s (len=%d)", got, wantBrokenSHA256, len(IMG_BROKEN))
	}
	if got := sha256Hex(IMG_PLAYBTN); got != wantPlaybtnSHA256 {
		t.Errorf("IMG_PLAYBTN sha256 = %s, want %s (len=%d)", got, wantPlaybtnSHA256, len(IMG_PLAYBTN))
	}
	if len(IMG_BROKEN) != 2150 {
		t.Errorf("IMG_BROKEN len = %d, want 2150", len(IMG_BROKEN))
	}
	if len(IMG_PLAYBTN) != 74402 {
		t.Errorf("IMG_PLAYBTN len = %d, want 74402", len(IMG_PLAYBTN))
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
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
