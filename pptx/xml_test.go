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

func TestCustomizeThemeColorsAndEscapeFonts(t *testing.T) {
	pres := &IPresentationProps{PresentationProps: PresentationProps{Theme: ThemeProps{
		HeadFontFace: `Heading & Display`,
		BodyFontFace: `Body "Sans"`,
	}}}
	xml := customizeThemeColors(makeXmlTheme(pres), ThemeColorScheme{
		Dark1: "112233", Light1: "FDFCFB", Accent1: "a1b2c3", FollowedLink: "654321",
	})
	for _, want := range []string{
		`<a:dk1><a:srgbClr val="112233"/></a:dk1>`,
		`<a:lt1><a:srgbClr val="FDFCFB"/></a:lt1>`,
		`<a:accent1><a:srgbClr val="A1B2C3"/></a:accent1>`,
		`<a:folHlink><a:srgbClr val="654321"/></a:folHlink>`,
		`typeface="Heading &amp; Display"`,
		`typeface="Body &quot;Sans&quot;"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("custom theme XML does not contain %q", want)
		}
	}
	if strings.Contains(xml, `accent1><a:srgbClr val="4472C4"`) {
		t.Error("custom accent1 retained the Office default")
	}
}

func TestImageSizingUsesIntrinsicSourceRatio(t *testing.T) {
	obj := &SlideObject{
		Type: SlideObjectTypeImage, ImageRID: 1, Image: "wide.png",
		Options: &ObjectOptions{
			ObjectNameProps: ObjectNameProps{ObjectName: "wide"},
			Sizing:          &ImageSizing{Type: "cover", W: Inches(1), H: Inches(1), SourceW: 200, SourceH: 100},
		},
	}
	slide := &SlideBaseProps{SlideObjects: []SlideObject{*obj}, RelsMedia: []SlideRelMedia{{RID: 1, Extn: "png"}}}
	xml := slideObjectImageToXml(&slide.SlideObjects[0], slide, nil, 0, 0, 100, 100, 100, 100, slide.SlideObjects[0].Options.Sizing, nil, "")
	if !strings.Contains(xml, `<a:srcRect l="25000" r="25000" t="0" b="0"/>`) {
		t.Fatalf("cover sizing did not use 2:1 source ratio: %s", xml)
	}
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
// core.xml — timestamps come from the build context's clock (bc.now).
// ---------------------------------------------------------------------------

var coreTsRe = regexp.MustCompile(`<dcterms:(created|modified) xsi:type="dcterms:W3CDTF">[^<]*</dcterms:(created|modified)>`)

func normalizeCoreTs(s string) string {
	return coreTsRe.ReplaceAllString(s, `<dcterms:$1 xsi:type="dcterms:W3CDTF">TS</dcterms:$2>`)
}

// testBC returns a build context with wall-clock time and the real UUID
// generator, for generators that don't assert on timestamps/GUIDs.
func testBC() *buildContext { return &buildContext{now: time.Now, uuid: getUuid} }

func TestMakeXmlCore(t *testing.T) {
	bc := &buildContext{
		now:  func() time.Time { return time.Date(2026, 7, 20, 4, 28, 22, 0, time.UTC) },
		uuid: getUuid,
	}

	got := makeXmlCore(bc, "PptxGenGo Golden Test Title", "PptxGenGo Golden Test Subject", "PptxGenGo Test Author", "1")
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
	assertEqual(t, makeXmlPresentation(pres, testBC()), readGoldenStr(t, "01-basic", "ppt", "presentation.xml"))
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

// TestBkgdEmptyStringTreatedAsAbsent covers REVIEW.md Nit (xml.go:111): TS
// checks `!slide.bkgd` (falsy), so `bkgd: ""` is equivalent to `bkgd:
// undefined`. Bkgd is `any` in Go, so a non-nil interface holding "" must be
// treated the same as nil — otherwise the DEFAULT-layout bgRef fallback never
// fires for a slide/layout that explicitly (if pointlessly) set `bkgd: ""`.
func TestBkgdEmptyStringTreatedAsAbsent(t *testing.T) {
	slide := &SlideBaseProps{Name: DEF_PRES_LAYOUT_NAME, Bkgd: ""}
	got := slideObjectToXml(slide, nil)
	want := `<p:bg><p:bgRef idx="1001"><a:schemeClr val="bg1"/></p:bgRef></p:bg>`
	if !strings.Contains(got, want) {
		t.Errorf("Bkgd:\"\" should be treated as absent (JS falsy), triggering the default bgRef; got: %s", got)
	}
}

// TestMarginOutOfContractLenZeroFillsInsets covers REVIEW.md Minor
// (xml.go:213-224): TS's margin-to-bodyPr conversion only special-cases
// `typeof margin === 'number'` (Go's len==1 uniform slice); any other array
// length goes through `margin[i] || 0` and zero-fills missing indices. The Go
// port used to require len==4 exactly, silently no-op'ing (leaving all four
// insets at their previous/zero value) for len 2 or 3. It must now populate
// the present indices and zero-fill the rest, for any length other than 1.
func TestMarginOutOfContractLenZeroFillsInsets(t *testing.T) {
	for _, tc := range []struct {
		name   string
		margin Margin
	}{
		{"len2", Margin{10, 20}},
		{"len3", Margin{10, 20, 30}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := &ObjectOptions{
				PositionProps:   PositionProps{X: coordPtr(Inches(1)), Y: coordPtr(Inches(1)), W: coordPtr(Inches(8)), H: coordPtr(Inches(1))},
				TextBaseProps:   TextBaseProps{FontSize: 12, Color: "000000"},
				ObjectNameProps: ObjectNameProps{ObjectName: "Text 0"},
				Line:            &ShapeLineProps{},
				Margin:          tc.margin,
			}
			slide := &SlideBaseProps{SlideNum: 1, SlideObjects: []SlideObject{{
				Type: SlideObjectTypeText, Shape: ShapeTypeRect, Options: opts,
				Text: []TextProps{{Text: "Hi", Options: &TextPropsOptions{}}},
			}}}
			_ = slideObjectToXml(slide, nil)

			wantL := float64(valToPts(marginAt(tc.margin, 0)))
			wantR := float64(valToPts(marginAt(tc.margin, 1)))
			wantB := float64(valToPts(marginAt(tc.margin, 2))) // 0 for len2, real for len3
			wantT := float64(valToPts(marginAt(tc.margin, 3))) // always 0 (out of contract)
			if opts.BodyProp == nil {
				t.Fatalf("BodyProp not set")
			}
			if opts.BodyProp.LIns != wantL {
				t.Errorf("LIns = %v, want %v", opts.BodyProp.LIns, wantL)
			}
			if opts.BodyProp.RIns != wantR {
				t.Errorf("RIns = %v, want %v", opts.BodyProp.RIns, wantR)
			}
			if opts.BodyProp.BIns != wantB {
				t.Errorf("BIns = %v, want %v", opts.BodyProp.BIns, wantB)
			}
			if opts.BodyProp.TIns != wantT {
				t.Errorf("TIns = %v, want %v (out-of-contract index zero-filled)", opts.BodyProp.TIns, wantT)
			}
		})
	}
}

// TestSlideNumberMarginOutOfContractLenZeroFillsInsets is the sibling of
// TestMarginOutOfContractLenZeroFillsInsets for the slide-number bodyPr
// insets (REVIEW.md Minor, xml.go:415-420). Unlike the shape-margin path,
// this one writes the lIns/tIns/rIns/bIns attributes unconditionally, so the
// zero-fill is directly observable in the rendered XML.
func TestSlideNumberMarginOutOfContractLenZeroFillsInsets(t *testing.T) {
	margin := Margin{10, 20} // len2: TS zero-fills margin[2]/margin[3]
	slide := &SlideBaseProps{
		SlideNum:         1,
		SlideNumberProps: &SlideNumberProps{Margin: margin},
	}
	got := slideObjectToXml(slide, nil)

	want := ` lIns="` + itoa(valToPts(marginAt(margin, 3))) + `"` + // 0, out of contract
		` tIns="` + itoa(valToPts(marginAt(margin, 0))) + `"` +
		` rIns="` + itoa(valToPts(marginAt(margin, 1))) + `"` +
		` bIns="` + itoa(valToPts(marginAt(margin, 2))) + `"` // 0, out of contract
	if !strings.Contains(got, want) {
		t.Errorf("missing zero-filled slide-number bodyPr insets.\nwant substring: %s\ngot: %s", want, got)
	}
}

// TestInheritRunOptionsFalsyOverwrite reproduces the JS falsy-inheritance
// semantics verified against live pptxgenjs (REVIEW.md M7): a shape-level
// `bold: true` OVERWRITES a run's explicit `bold: false`, because the TS
// inheritance loop (`gen-xml.ts:1292-1296`) does `if (!textObj.options[key])
// textObj.options[key] = val`, and `false` is falsy in JS. This must hold for
// Bold/Italic/Subscript/Superscript (boolean-valued options); non-boolean
// (object/array/pointer) fields like Underline/Outline/Glow/Hyperlink/TabStops
// stay nil-check gated since JS objects are always truthy when non-null.
func TestInheritRunOptionsFalsyOverwrite(t *testing.T) {
	shape := &ObjectOptions{
		PositionProps:   PositionProps{X: coordPtr(Inches(0.5)), Y: coordPtr(Inches(0.5)), W: coordPtr(Inches(9)), H: coordPtr(Inches(1))},
		TextBaseProps:   TextBaseProps{Color: "000000", Bold: ptr(true), Italic: ptr(true)},
		ObjectNameProps: ObjectNameProps{ObjectName: "Text 0"},
		Line:            &ShapeLineProps{},
		BodyProp:        &BodyProps{},
		Subscript:       ptr(true),
		Superscript:     ptr(true),
	}
	obj := SlideObject{
		Type:    SlideObjectTypeText,
		Shape:   ShapeTypeRect,
		Options: shape,
		Text: []TextProps{
			{Text: "Run", Options: &TextPropsOptions{
				TextBaseProps: TextBaseProps{Color: "000000", Bold: ptr(false), Italic: ptr(false)},
				Subscript:     ptr(false),
				Superscript:   ptr(false),
			}},
		},
	}
	got := genXmlTextBody(&obj)

	run := extractBetween(t, got, "<a:r>", "</a:r>")
	if !strings.Contains(run, `b="1"`) {
		t.Errorf("shape bold:true must overwrite run bold:false (JS falsy semantics); run xml: %s", run)
	}
	if !strings.Contains(run, `i="1"`) {
		t.Errorf("shape italic:true must overwrite run italic:false; run xml: %s", run)
	}
	// Subscript/Superscript both true is a contradiction in practice, but the
	// inheritance rule only cares that an explicit `false` on the run gets
	// overwritten; assert at least one baseline attribute reflecting the
	// overwritten (truthy) shape value made it through rather than being
	// suppressed by the run's stale falsy pointer.
	if !strings.Contains(run, `baseline=`) {
		t.Errorf("expected baseline attr from overwritten subscript/superscript; run xml: %s", run)
	}
}

// TestVmergeDummyCellClampAppends reproduces the TS splice-clamp behavior for
// irregular colspan+rowspan grids (REVIEW.md Minor, xml.go:832-839). When a
// rowspan>1 cell's column index (cIdx) lands beyond the next row's current
// length, TS's `nextRow.splice(cIdx, 0, hMergeCell)` still inserts the cell —
// JS Array.prototype.splice clamps an out-of-range start index to the array's
// length instead of no-op'ing. The Go port must append the merge cell at the
// end in that case rather than silently dropping it.
//
// Row0 (post hmerge-expansion) has 4 cells at indices 0..3, only the last
// (index 3) has rowspan=2; row1 starts with 0 real cells, so by the time cIdx
// reaches 3, len(nextRow) is still 0 and the naive `cIdx <= len(nextRow)`
// guard used to drop the merge cell entirely.
func TestVmergeDummyCellClampAppends(t *testing.T) {
	row0 := []TableCell{
		{Type: SlideObjectTypeTablecell, Text: "A", Options: &TableCellProps{}},
		{Type: SlideObjectTypeTablecell, Text: "B", Options: &TableCellProps{}},
		{Type: SlideObjectTypeTablecell, Text: "C", Options: &TableCellProps{}},
		{Type: SlideObjectTypeTablecell, Text: "D", Options: &TableCellProps{Rowspan: 2}},
	}
	row1 := []TableCell{} // deliberately shorter than row0 at the merge point
	obj := &SlideObject{
		Type:       SlideObjectTypeTable,
		Options:    &ObjectOptions{ObjectNameProps: ObjectNameProps{ObjectName: "Table 1"}},
		ArrTabRows: [][]TableCell{row0, row1},
	}
	slide := &SlideBaseProps{SlideNum: 1}

	got := slideObjectTableToXml(obj, slide, 1, 0, 0, 9144000, 1000000)

	trs := strings.Split(got, `<a:tr `)
	if len(trs) < 3 {
		t.Fatalf("expected 2 <a:tr> rows, got %d\nxml: %s", len(trs)-1, got)
	}
	row1Xml := trs[2]
	if !strings.Contains(row1Xml, `vMerge="1"`) {
		t.Errorf("expected the rowspan=2 dummy cell to be appended (clamped) into row1, not dropped; row1 xml: %s", row1Xml)
	}
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

	got := makeXmlPresentation(pres, testBC())

	if !strings.Contains(got, `embedTrueTypeFonts="1"`) {
		t.Errorf("presentation missing embedTrueTypeFonts attribute")
	}
	wantLst := `<p:embeddedFontLst><p:embeddedFont><p:font typeface="Lato"/><p:regular r:id="rId8"/><p:bold r:id="rId9"/></p:embeddedFont><p:embeddedFont><p:font typeface="Merri"/><p:italic r:id="rId10"/></p:embeddedFont></p:embeddedFontLst>`
	if !strings.Contains(got, wantLst) {
		t.Errorf("presentation missing/incorrect embeddedFontLst.\ngot: %s", got)
	}
	// placement: ECMA-376 CT_Presentation sequence puts embeddedFontLst AFTER
	// sldIdLst/sldSz/notesSz (not immediately after sldMasterIdLst). Verify it
	// lands immediately after the (self-closing) <p:notesSz.../>, and after
	// sldIdLst/sldSz too.
	notesSzRe := regexp.MustCompile(`<p:notesSz[^>]*/>`)
	notesSzLoc := notesSzRe.FindStringIndex(got)
	iSld := strings.Index(got, "<p:sldIdLst>")
	iSldSz := strings.Index(got, "<p:sldSz")
	iLst := strings.Index(got, "<p:embeddedFontLst>")
	if iSld < 0 || iSldSz < 0 || notesSzLoc == nil || iLst < 0 {
		t.Fatalf("missing expected elements: sldIdLst=%d sldSz=%d notesSz=%v embeddedFontLst=%d", iSld, iSldSz, notesSzLoc, iLst)
	}
	iNotesSzEnd := notesSzLoc[1]
	if !(iSld < iSldSz && iSldSz < iNotesSzEnd && iNotesSzEnd <= iLst) {
		t.Fatalf("embeddedFontLst misplaced: sldIdLst=%d sldSz=%d notesSzEnd=%d lst=%d (want sldIdLst < sldSz < notesSzEnd <= embeddedFontLst)", iSld, iSldSz, iNotesSzEnd, iLst)
	}
	// must be immediately after the closing notesSz tag (no other elements between).
	if got[iNotesSzEnd:iLst] != "" {
		t.Errorf("embeddedFontLst must immediately follow <p:notesSz.../>; found %q in between", got[iNotesSzEnd:iLst])
	}
}
