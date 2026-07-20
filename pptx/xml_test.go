package pptx

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func readGoldenStr(t *testing.T, parts ...string) string {
	t.Helper()
	b, err := os.ReadFile(goldenPath(parts...))
	if err != nil {
		t.Fatalf("read golden %v: %v", parts, err)
	}
	return string(b)
}

// diffAt reports the first byte offset where a and b differ, with context.
func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got == want {
		return
	}
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	i := 0
	for i < n && got[i] == want[i] {
		i++
	}
	lo := i - 60
	if lo < 0 {
		lo = 0
	}
	hiG := i + 60
	if hiG > len(got) {
		hiG = len(got)
	}
	hiW := i + 60
	if hiW > len(want) {
		hiW = len(want)
	}
	t.Errorf("mismatch at byte %d (got len %d, want len %d)\n got: ...%q...\nwant: ...%q...", i, len(got), len(want), got[lo:hiG], want[lo:hiW])
}

func coordPtr(c Coord) *Coord { return &c }

var defLayout = PresLayout{Name: "LAYOUT_16x9", Width: 9144000, Height: 5143500, SizeW: 9144000, SizeH: 5143500}

// ---------------------------------------------------------------------------
// Small static parts
// ---------------------------------------------------------------------------

func TestMakeXmlRootRels(t *testing.T) {
	assertEqual(t, makeXmlRootRels(), readGoldenStr(t, "01-basic", "_rels", ".rels"))
}

func TestMakeXmlPresProps(t *testing.T) {
	assertEqual(t, makeXmlPresProps(), readGoldenStr(t, "01-basic", "ppt", "presProps.xml"))
}

func TestMakeXmlTableStyles(t *testing.T) {
	assertEqual(t, makeXmlTableStyles(), readGoldenStr(t, "01-basic", "ppt", "tableStyles.xml"))
}

func TestMakeXmlViewProps(t *testing.T) {
	assertEqual(t, makeXmlViewProps(), readGoldenStr(t, "01-basic", "ppt", "viewProps.xml"))
}

func TestMakeXmlTheme(t *testing.T) {
	pres := &IPresentationProps{}
	assertEqual(t, makeXmlTheme(pres), readGoldenStr(t, "01-basic", "ppt", "theme", "theme1.xml"))
}

func TestMakeXmlNotesMaster(t *testing.T) {
	assertEqual(t, makeXmlNotesMaster(), readGoldenStr(t, "01-basic", "ppt", "notesMasters", "notesMaster1.xml"))
}

func TestMakeXmlNotesMasterRel(t *testing.T) {
	assertEqual(t, makeXmlNotesMasterRel(), readGoldenStr(t, "01-basic", "ppt", "notesMasters", "_rels", "notesMaster1.xml.rels"))
}

func TestMakeXmlNotesSlideRel(t *testing.T) {
	assertEqual(t, makeXmlNotesSlideRel(1), readGoldenStr(t, "01-basic", "ppt", "notesSlides", "_rels", "notesSlide1.xml.rels"))
}

func TestMakeXmlApp(t *testing.T) {
	slides := []PresSlide{{}}
	assertEqual(t, makeXmlApp(slides, "PptxGenGo Test Co"), readGoldenStr(t, "01-basic", "docProps", "app.xml"))
}

func TestMakeXmlContTypesBasic(t *testing.T) {
	slides := []PresSlide{{}}
	layouts := []SlideLayout{{}}
	master := &PresSlide{}
	assertEqual(t, makeXmlContTypes(slides, layouts, master, nil), readGoldenStr(t, "01-basic", "[Content_Types].xml"))
}

// ---------------------------------------------------------------------------
// core.xml — timestamps are injected via xmlNowFunc.
// ---------------------------------------------------------------------------

var coreTsRe = regexp.MustCompile(`<dcterms:(created|modified) xsi:type="dcterms:W3CDTF">[^<]*</dcterms:(created|modified)>`)

func normalizeCoreTs(s string) string {
	return coreTsRe.ReplaceAllString(s, `<dcterms:$1 xsi:type="dcterms:W3CDTF">TS</dcterms:$2>`)
}

func TestMakeXmlCore(t *testing.T) {
	old := xmlNowFunc
	xmlNowFunc = func() time.Time { return time.Date(2026, 7, 20, 4, 28, 22, 0, time.UTC) }
	defer func() { xmlNowFunc = old }()

	got := makeXmlCore("PptxGenGo Golden Test Title", "PptxGenGo Golden Test Subject", "PptxGenGo Test Author", "1")
	want := readGoldenStr(t, "01-basic", "docProps", "core.xml")

	// Exact (timestamp chosen to match golden), plus normalized safety net.
	assertEqual(t, got, want)
	if normalizeCoreTs(got) != normalizeCoreTs(want) {
		t.Errorf("core.xml differs beyond timestamps")
	}
}

// ---------------------------------------------------------------------------
// presentation.xml.rels + presentation.xml
// ---------------------------------------------------------------------------

func TestMakeXmlPresentationRels(t *testing.T) {
	slides := []PresSlide{{}}
	assertEqual(t, makeXmlPresentationRels(slides, nil), readGoldenStr(t, "01-basic", "ppt", "_rels", "presentation.xml.rels"))
}

func TestMakeXmlPresentation(t *testing.T) {
	pres := &IPresentationProps{}
	pres.PresLayout = defLayout
	pres.Slides = []PresSlide{{RID: 2, SlideID: 256}}
	assertEqual(t, makeXmlPresentation(pres), readGoldenStr(t, "01-basic", "ppt", "presentation.xml"))
}

// ---------------------------------------------------------------------------
// 01-basic slide model — shared by genXmlTextBody + makeXmlSlide tests.
// ---------------------------------------------------------------------------

func basicTextObject() SlideObject {
	bodyProp := &BodyProps{Anchor: TextVAlignCTR, Wrap: ptr(true), AutoFit: ptr(false)}
	return SlideObject{
		Type:  SlideObjectTypeText,
		Shape: ShapeTypeRect,
		Options: &ObjectOptions{
			PositionProps: PositionProps{
				X: coordPtr(Inches(1)),
				Y: coordPtr(Inches(1)),
				W: coordPtr(Inches(8)),
				H: coordPtr(Inches(1)),
			},
			TextBaseProps:   TextBaseProps{FontSize: 24, Color: "363636"},
			ObjectNameProps: ObjectNameProps{ObjectName: "Text 0"},
			Line:            &ShapeLineProps{},
			BodyProp:        bodyProp,
		},
		Text: []TextProps{
			{Text: "Hello World", Options: &TextPropsOptions{}},
		},
	}
}

func TestGenXmlTextBodyBasic(t *testing.T) {
	obj := basicTextObject()
	got := genXmlTextBody(&obj)

	slide := readGoldenStr(t, "01-basic", "ppt", "slides", "slide1.xml")
	want := extractBetween(t, slide, "<p:txBody>", "</p:txBody>")
	assertEqual(t, got, want)
}

func TestMakeXmlSlideBasic(t *testing.T) {
	obj := basicTextObject()
	slide := &PresSlide{}
	slide.Name = "Slide 1"
	slide.PresLayout = defLayout
	slide.SlideObjects = []SlideObject{obj}

	got := makeXmlSlide(slide)
	want := readGoldenStr(t, "01-basic", "ppt", "slides", "slide1.xml")
	assertEqual(t, got, want)
}

// extractBetween returns s[open..close] inclusive of the tags (first match).
func extractBetween(t *testing.T, s, open, close string) string {
	t.Helper()
	i := strings.Index(s, open)
	if i < 0 {
		t.Fatalf("open tag %q not found", open)
	}
	j := strings.Index(s[i:], close)
	if j < 0 {
		t.Fatalf("close tag %q not found", close)
	}
	return s[i : i+j+len(close)]
}

// ---------------------------------------------------------------------------
// 02-text-rich: first text box (multi-run) txBody fragment.
// ---------------------------------------------------------------------------

func TestGenXmlTextBodyRich(t *testing.T) {
	bodyProp := &BodyProps{Anchor: TextVAlignCTR, Wrap: ptr(true), AutoFit: ptr(false)}
	// addText(..., { x:0.5, y:0.5, w:9, h:1, fontSize:18, align:'center' })
	// cleanOpts defaults run color to DEF_FONT_COLOR ("000000").
	shape := &ObjectOptions{
		PositionProps:   PositionProps{X: coordPtr(Inches(0.5)), Y: coordPtr(Inches(0.5)), W: coordPtr(Inches(9)), H: coordPtr(Inches(1))},
		TextBaseProps:   TextBaseProps{FontSize: 18, Align: "center", Color: "000000"},
		ObjectNameProps: ObjectNameProps{ObjectName: "Text 0"},
		Line:            &ShapeLineProps{},
		BodyProp:        bodyProp,
	}
	runColor := func() TextBaseProps { return TextBaseProps{Color: "000000"} }
	obj := SlideObject{
		Type:    SlideObjectTypeText,
		Shape:   ShapeTypeRect,
		Options: shape,
		Text: []TextProps{
			{Text: "Bold ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Color: "000000", Bold: ptr(true)}}},
			{Text: "Italic ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Color: "000000", Italic: ptr(true)}}},
			{Text: "Underline ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Color: "000000", Underline: &UnderlineProps{Style: "sng"}}}},
			{Text: "CourierFace ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Color: "000000", FontFace: "Courier New"}}},
			{Text: "ExampleLink", Options: &TextPropsOptions{TextBaseProps: runColor(), Hyperlink: &HyperlinkProps{URL: "https://example.com", Tooltip: "Visit Example", RID: 1}}},
		},
	}
	got := genXmlTextBody(&obj)

	slide := readGoldenStr(t, "02-text-rich", "ppt", "slides", "slide1.xml")
	// first txBody in the file
	want := extractBetween(t, slide, "<p:txBody>", "</p:txBody>")
	assertEqual(t, got, want)
}

// ---------------------------------------------------------------------------
// Font-embedding hooks (no golden — asserted against ECMA-376-derived strings).
// ---------------------------------------------------------------------------

func fontFixture() []*EmbeddedFont {
	return []*EmbeddedFont{
		{Typeface: "Lato", Variants: map[FontStyle][]byte{FontRegular: {1}, FontBold: {1}}},
		{Typeface: "Merri", Variants: map[FontStyle][]byte{FontItalic: {1}}},
	}
}

func TestEmbeddedFontVariantsNumbering(t *testing.T) {
	vs := embeddedFontVariants(1, fontFixture())
	if len(vs) != 3 {
		t.Fatalf("want 3 variants, got %d", len(vs))
	}
	// base = slides(1) + 6 = 7; first font rId = 8
	wantRIDs := []int{8, 9, 10}
	wantFiles := []int{1, 2, 3}
	wantStyles := []FontStyle{FontRegular, FontBold, FontItalic}
	for i, v := range vs {
		if v.rID != wantRIDs[i] || v.fileNum != wantFiles[i] || v.style != wantStyles[i] {
			t.Errorf("variant %d = {rID:%d file:%d style:%s}, want {rID:%d file:%d style:%s}", i, v.rID, v.fileNum, v.style, wantRIDs[i], wantFiles[i], wantStyles[i])
		}
	}
}

func TestMakeXmlContTypesFontDefault(t *testing.T) {
	slides := []PresSlide{{}}
	layouts := []SlideLayout{{}}
	master := &PresSlide{}

	without := makeXmlContTypes(slides, layouts, master, nil)
	if strings.Contains(without, "fntdata") {
		t.Errorf("no-font content types should not mention fntdata")
	}

	with := makeXmlContTypes(slides, layouts, master, fontFixture())
	if !strings.Contains(with, `<Default Extension="fntdata" ContentType="application/x-fontdata"/>`) {
		t.Errorf("font content types missing fntdata Default")
	}
	// fntdata must sit among the Defaults (before the presentation Override).
	if strings.Index(with, "fntdata") > strings.Index(with, "/ppt/presentation.xml") {
		t.Errorf("fntdata Default must precede presentation.xml Override")
	}
}

func TestMakeXmlPresentationRelsFonts(t *testing.T) {
	slides := []PresSlide{{}}
	got := makeXmlPresentationRels(slides, fontFixture())
	wantRels := []string{
		`<Relationship Id="rId8" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/font" Target="fonts/font1.fntdata"/>`,
		`<Relationship Id="rId9" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/font" Target="fonts/font2.fntdata"/>`,
		`<Relationship Id="rId10" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/font" Target="fonts/font3.fntdata"/>`,
	}
	for _, r := range wantRels {
		if !strings.Contains(got, r) {
			t.Errorf("missing font relationship:\n%s", r)
		}
	}
	// fonts appended after tableStyles rel (rId7)
	if strings.Index(got, "tableStyles.xml") > strings.Index(got, "font1.fntdata") {
		t.Errorf("font rels must come after tableStyles rel")
	}
}

func TestMakeXmlPresentationFontHooks(t *testing.T) {
	pres := &IPresentationProps{}
	pres.PresLayout = defLayout
	pres.Slides = []PresSlide{{RID: 2, SlideID: 256}}
	pres.EmbeddedFonts = fontFixture()

	got := makeXmlPresentation(pres)

	if !strings.Contains(got, `embedTrueTypeFonts="1"`) {
		t.Errorf("presentation missing embedTrueTypeFonts attribute")
	}
	wantLst := `<p:embeddedFontLst><p:embeddedFont><p:font typeface="Lato"/><p:regular r:id="rId8"/><p:bold r:id="rId9"/></p:embeddedFont><p:embeddedFont><p:font typeface="Merri"/><p:italic r:id="rId10"/></p:embeddedFont></p:embeddedFontLst>`
	if !strings.Contains(got, wantLst) {
		t.Errorf("presentation missing/incorrect embeddedFontLst.\ngot: %s", got)
	}
	// placement: after sldMasterIdLst, before sldIdLst
	iMaster := strings.Index(got, "</p:sldMasterIdLst>")
	iLst := strings.Index(got, "<p:embeddedFontLst>")
	iSld := strings.Index(got, "<p:sldIdLst>")
	if !(iMaster < iLst && iLst < iSld) {
		t.Errorf("embeddedFontLst misplaced: master=%d lst=%d sld=%d", iMaster, iLst, iSld)
	}
}
