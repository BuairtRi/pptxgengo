package pptx

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// tinyRedPNGBase64 is the same 5x5 solid-red PNG used by scripts/gen-golden.mjs
// (case 07-image). The Go image part is produced by base64-decoding this string,
// so it must byte-match the golden image-1-1.png.
const tinyRedPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEUlEQVR42mP4z8CAjBgo5AMA/XwY6DI3DH0AAAAASUVORK5CYII="

// goldenTS is the fixed timestamp injected into every generated deck so that
// docProps/core.xml (and the embedded workbook's core.xml) are deterministic
// and match the golden fixtures byte-for-byte.
func goldenTS() time.Time { return time.Date(2026, 7, 20, 4, 28, 22, 0, time.UTC) }

// setGoldenMeta applies the doc-prop overrides every golden case shares.
func setGoldenMeta(p *Presentation) {
	p.Author = "PptxGenGo Test Author"
	p.Company = "PptxGenGo Test Co"
	p.Subject = "PptxGenGo Golden Test Subject"
	p.Title = "PptxGenGo Golden Test Title"
	p.nowFunc = goldenTS
}

// ---------------------------------------------------------------------------
// Case builders — each mirrors the matching defineCase() in gen-golden.mjs.
// ---------------------------------------------------------------------------

func buildCase01(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)
	slide := p.AddSlide()
	must(t, slide.AddText(
		[]TextProps{{Text: "Hello World"}},
		&TextPropsOptions{
			PositionProps: PositionProps{X: cp(Inches(1)), Y: cp(Inches(1)), W: cp(Inches(8)), H: cp(Inches(1))},
			TextBaseProps: TextBaseProps{FontSize: 24, Color: "363636"},
		},
	))
	return p
}

func buildCase02(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)
	slide := p.AddSlide()

	must(t, slide.AddText(
		[]TextProps{
			{Text: "Bold ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Bold: ptr(true)}}},
			{Text: "Italic ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Italic: ptr(true)}}},
			{Text: "Underline ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Underline: &UnderlineProps{Style: "sng"}}}},
			{Text: "CourierFace ", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{FontFace: "Courier New"}}},
			{Text: "ExampleLink", Options: &TextPropsOptions{Hyperlink: &HyperlinkProps{URL: "https://example.com", Tooltip: "Visit Example"}}},
		},
		&TextPropsOptions{
			PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(0.5)), W: cp(Inches(9)), H: cp(Inches(1))},
			TextBaseProps: TextBaseProps{FontSize: 18, Align: "center"},
		},
	))

	must(t, slide.AddText(
		[]TextProps{
			{Text: "First bullet point", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Bullet: &BulletProps{}, BreakLine: ptr(true)}}},
			{Text: "Second bullet point", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Bullet: &BulletProps{}, BreakLine: ptr(true)}}},
			{Text: "Numbered item one", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Bullet: &BulletProps{Type: "number"}, BreakLine: ptr(true)}}},
			{Text: "Numbered item two", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Bullet: &BulletProps{Type: "number"}}}},
		},
		&TextPropsOptions{
			PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(2)), W: cp(Inches(5)), H: cp(Inches(2.5))},
			TextBaseProps: TextBaseProps{FontSize: 14},
		},
	))

	must(t, slide.AddText(
		[]TextProps{
			{Text: "Left aligned line", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Align: "left", BreakLine: ptr(true)}}},
			{Text: "Center aligned line", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Align: "center", BreakLine: ptr(true)}}},
			{Text: "Right aligned line", Options: &TextPropsOptions{TextBaseProps: TextBaseProps{Align: "right", BreakLine: ptr(true)}}},
			{Text: "H", Options: &TextPropsOptions{}},
			{Text: "2", Options: &TextPropsOptions{Subscript: ptr(true)}},
			{Text: "O and E=mc", Options: &TextPropsOptions{}},
			{Text: "2", Options: &TextPropsOptions{Superscript: ptr(true)}},
		},
		&TextPropsOptions{
			PositionProps: PositionProps{X: cp(Inches(6)), Y: cp(Inches(2)), W: cp(Inches(3.5)), H: cp(Inches(2.5))},
			TextBaseProps: TextBaseProps{FontSize: 12},
		},
	))
	return p
}

func buildCase03(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)
	slide := p.AddSlide()

	must(t, slide.AddShape(ShapeTypeRect, &ShapeProps{
		PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(0.5)), W: cp(Inches(2)), H: cp(Inches(1))},
		Fill:          &ShapeFillProps{Color: "2E86AB"},
		Line:          &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "1B4965"}, Width: 1},
	}))
	must(t, slide.AddShape(ShapeTypeRoundRect, &ShapeProps{
		PositionProps: PositionProps{X: cp(Inches(3)), Y: cp(Inches(0.5)), W: cp(Inches(2)), H: cp(Inches(1))},
		RectRadius:    0.15,
		Fill:          &ShapeFillProps{Color: "A23B72"},
		Line:          &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "6A2049"}, Width: 1},
	}))
	must(t, slide.AddShape(ShapeTypeEllipse, &ShapeProps{
		PositionProps: PositionProps{X: cp(Inches(5.5)), Y: cp(Inches(0.5)), W: cp(Inches(2)), H: cp(Inches(1))},
		Fill:          &ShapeFillProps{Color: "F18F01"},
		Line:          &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "C46F00"}, Width: 2, DashType: "dash"},
	}))
	must(t, slide.AddShape(ShapeTypeLine, &ShapeProps{
		PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(2)), W: cp(Inches(3)), H: cp(Inches(0))},
		Line:          &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "363636"}, Width: 2, BeginArrowType: "triangle", EndArrowType: "arrow"},
	}))
	must(t, slide.AddShape(ShapeTypeTriangle, &ShapeProps{
		PositionProps: PositionProps{X: cp(Inches(4.5)), Y: cp(Inches(2)), W: cp(Inches(1.5)), H: cp(Inches(1.5))},
		Rotate:        45,
		Fill:          &ShapeFillProps{Color: "4CAF50"},
		Line:          &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "2E7D32"}, Width: 1},
	}))
	must(t, slide.AddShape(ShapeTypeRect, &ShapeProps{
		PositionProps: PositionProps{X: cp(Inches(7)), Y: cp(Inches(2)), W: cp(Inches(2)), H: cp(Inches(1.5))},
		Fill:          &ShapeFillProps{Color: "FFFFFF"},
		Line:          &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "333333"}, Width: 1},
		Shadow:        &ShadowProps{Type: "outer", Color: "000000", Opacity: 0.5, Blur: 3, Angle: 45, Offset: 3},
	}))
	return p
}

func buildCase04(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)
	slide := p.AddSlide()

	border := []BorderProps{{Type: "solid", Color: "999999", Pt: 1}}
	rows := []TableRow{
		{
			{Text: "Merged Header", Options: &TableCellProps{
				TextBaseProps: TextBaseProps{Bold: ptr(true), Color: "FFFFFF", Align: "center"},
				Colspan:       3, Fill: &ShapeFillProps{Color: "2E86AB"}, Border: border,
			}},
		},
		{
			{Text: "Col A", Options: &TableCellProps{TextBaseProps: TextBaseProps{Bold: ptr(true)}, Fill: &ShapeFillProps{Color: "DDDDDD"}, Border: border}},
			{Text: "Col B", Options: &TableCellProps{TextBaseProps: TextBaseProps{Bold: ptr(true)}, Fill: &ShapeFillProps{Color: "DDDDDD"}, Border: border}},
			{Text: "Col C", Options: &TableCellProps{TextBaseProps: TextBaseProps{Bold: ptr(true)}, Fill: &ShapeFillProps{Color: "DDDDDD"}, Border: border}},
		},
		{
			{Text: "r1c1", Options: &TableCellProps{Border: border}},
			{Text: "r1c2", Options: &TableCellProps{Border: border, Fill: &ShapeFillProps{Color: "FFF3CD"}}},
			{Text: "r1c3", Options: &TableCellProps{Border: border}},
		},
		{
			{Text: "r2c1", Options: &TableCellProps{Border: border}},
			{Text: "r2c2", Options: &TableCellProps{Border: border}},
			{Text: "r2c3", Options: &TableCellProps{Border: border}},
		},
	}
	must(t, slide.AddTable(rows, &TableProps{
		PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(0.5)), W: cp(Inches(9))},
		ColW:          []float64{3, 3, 3},
		Border:        border,
	}))
	return p
}

func buildCase05(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)
	slide := p.AddSlide()

	data := []ChartData{{Name: "Revenue", Labels: [][]string{{"Q1", "Q2", "Q3", "Q4"}}, Values: []float64{100, 150, 130, 175}}}
	must(t, slide.AddChart(ChartTypeBar, data, &ChartOptions{
		PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(0.5)), W: cp(Inches(9)), H: cp(Inches(5))},
		ShowTitle:     ptr(true), Title: "Quarterly Revenue",
		ShowCatAxisTitle: ptr(true), CatAxisTitle: "Quarter",
		ShowValAxisTitle: ptr(true), ValAxisTitle: "USD (thousands)",
		ChartColors: []string{"2E86AB", "A23B72", "F18F01", "4CAF50"},
	}))
	return p
}

func buildCase06(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)

	slide1 := p.AddSlide()
	lineData := []ChartData{
		{Name: "Series A", Labels: [][]string{{"Jan", "Feb", "Mar", "Apr"}}, Values: []float64{10, 20, 15, 25}},
		{Name: "Series B", Labels: [][]string{{"Jan", "Feb", "Mar", "Apr"}}, Values: []float64{5, 12, 18, 9}},
	}
	must(t, slide1.AddChart(ChartTypeLine, lineData, &ChartOptions{
		PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(0.5)), W: cp(Inches(9)), H: cp(Inches(5))},
		ShowTitle:     ptr(true), Title: "Two Series Line Chart",
		ShowLegend:  ptr(true),
		ChartColors: []string{"2E86AB", "A23B72"},
	}))

	slide2 := p.AddSlide()
	pieData := []ChartData{{Name: "Share", Labels: [][]string{{"Alpha", "Beta", "Gamma"}}, Values: []float64{40, 35, 25}}}
	must(t, slide2.AddChart(ChartTypePie, pieData, &ChartOptions{
		PositionProps: PositionProps{X: cp(Inches(1)), Y: cp(Inches(0.5)), W: cp(Inches(7)), H: cp(Inches(5))},
		ShowTitle:     ptr(true), Title: "Market Share",
		ShowLegend:  ptr(true),
		ChartColors: []string{"2E86AB", "A23B72", "F18F01"},
	}))
	return p
}

func buildCase07(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)
	slide := p.AddSlide()
	must(t, slide.AddImage(&ImageProps{
		DataOrPathProps: DataOrPathProps{Data: "image/png;base64," + tinyRedPNGBase64},
		PositionProps:   PositionProps{X: cp(Inches(1)), Y: cp(Inches(1)), W: cp(Inches(2)), H: cp(Inches(2))},
	}))
	return p
}

func buildCase08(t *testing.T) *Presentation {
	p := New()
	setGoldenMeta(p)

	must(t, p.DefineSlideMaster(&SlideMasterProps{
		Title:       "GOLDEN_MASTER",
		Background:  &BackgroundProps{ShapeFillProps: ShapeFillProps{Color: "F1F1F1"}},
		SlideNumber: &SlideNumberProps{PositionProps: PositionProps{X: cp(Inches(9)), Y: cp(Inches(5.5))}, TextBaseProps: TextBaseProps{Color: "363636"}},
		Objects: []SlideMasterObject{
			{Placeholder: &SlideMasterPlaceholder{
				Options: PlaceholderProps{
					PositionProps: PositionProps{X: cp(Inches(0.5)), Y: cp(Inches(0.3)), W: cp(Inches(9)), H: cp(Inches(1))},
					Name:          "title", Type: PlaceholderTypeTitle,
				},
				Text: "Click to add title",
			}},
			{Rect: &ShapeProps{
				PositionProps: PositionProps{X: cp(Inches(0)), Y: cp(Inches(5.3)), W: cp(Percent(100)), H: cp(Inches(0.3))},
				Fill:          &ShapeFillProps{Color: "2E86AB"},
			}},
		},
	}))

	p.AddSection(SectionProps{Title: "Golden Section"})

	slide := p.AddSlide(&AddSlideProps{MasterName: "GOLDEN_MASTER", SectionTitle: "Golden Section"})
	// Mirror the JS string-form addText('...', {placeholder:'title'}): the single
	// run shares the box options, so the run also carries the placeholder (which
	// suppresses default run color, as placeholders inherit from the master).
	must(t, slide.AddText(
		[]TextProps{{Text: "Slide using GOLDEN_MASTER", Options: &TextPropsOptions{Placeholder: "title"}}},
		&TextPropsOptions{Placeholder: "title"},
	))
	return p
}

// ---------------------------------------------------------------------------
// The golden gate: build every case through the public API and byte-compare.
// ---------------------------------------------------------------------------

// TestGoldenIntegration rebuilds each golden deck (01..08) through the public
// Go API, in the same order gen-golden.mjs generated them (so the shared global
// chart counter lines up), and compares every emitted part against the golden
// fixtures. Only two normalizations are permitted for the .pptx XML parts:
// (a) docProps/core.xml dcterms timestamps, (b) documented-random UUIDs. The
// embedded .xlsx workbook part is a nested ZIP whose byte encoding is produced
// by JSZip (not reproducible by archive/zip); it is therefore compared by
// recursively unzipping and matching its inner parts (its own core.xml
// timestamp normalized) — see the report note.
func TestGoldenIntegration(t *testing.T) {
	// Deterministic global state.
	_chartCounter = 0

	cases := []struct {
		name  string
		build func(*testing.T) *Presentation
	}{
		{"01-basic", buildCase01},
		{"02-text-rich", buildCase02},
		{"03-shapes", buildCase03},
		{"04-table", buildCase04},
		{"05-chart-bar", buildCase05},
		{"06-chart-multi", buildCase06},
		{"07-image", buildCase07},
		{"08-master", buildCase08},
	}

	for _, c := range cases {
		p := c.build(t)
		data, err := p.Write()
		if err != nil {
			t.Fatalf("[%s] Write: %v", c.name, err)
		}
		compareAgainstGolden(t, c.name, data)
	}
}

var guidRe = regexp.MustCompile(`\{[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}\}`)

func normGUID(s string) string { return guidRe.ReplaceAllString(s, "{GUID}") }

func compareAgainstGolden(t *testing.T, caseName string, gotZip []byte) {
	t.Helper()
	got := unzipParts(t, gotZip)
	want := readGoldenDir(t, caseName)

	// Part-set equality (no missing / extra files).
	for name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("[%s] MISSING part %q in generated deck", caseName, name)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("[%s] EXTRA part %q in generated deck", caseName, name)
		}
	}

	for name, w := range want {
		g, ok := got[name]
		if !ok {
			continue
		}
		comparePart(t, caseName, name, g, w)
	}
}

func comparePart(t *testing.T, caseName, name, got, want string) {
	t.Helper()

	// Nested workbook: recurse (see method doc comment).
	if strings.HasSuffix(name, ".xlsx") {
		compareXlsxPart(t, caseName, name, []byte(got), []byte(want))
		return
	}

	// Exact byte match — the required outcome for every XML/binary part.
	if got == want {
		return
	}

	// Permitted normalization #1: core.xml timestamps.
	if normalizeCore(got) == normalizeCore(want) {
		t.Logf("[%s] %s: identical after timestamp normalization (allowed)", caseName, name)
		return
	}
	// Permitted normalization #2: documented-random UUIDs (+ timestamps).
	if normGUID(normalizeCore(got)) == normGUID(normalizeCore(want)) {
		t.Logf("[%s] %s: identical after UUID normalization (allowed)", caseName, name)
		return
	}

	// Otherwise: a real difference — fail with first-divergence context.
	i := firstDiff(got, want)
	t.Errorf("[%s] %s: byte mismatch at offset %d\n got: ...%s...\nwant: ...%s...",
		caseName, name, i, snippet(got, i), snippet(want, i))
}

// compareXlsxPart unzips both workbook parts and compares their inner parts,
// normalizing only the workbook's own docProps/core.xml timestamp.
func compareXlsxPart(t *testing.T, caseName, name string, got, want []byte) {
	t.Helper()
	gotParts := unzipParts(t, got)
	wantParts := unzipParts(t, want)

	for inner := range wantParts {
		if _, ok := gotParts[inner]; !ok {
			t.Errorf("[%s] %s: MISSING inner part %q", caseName, name, inner)
		}
	}
	for inner := range gotParts {
		if _, ok := wantParts[inner]; !ok {
			t.Errorf("[%s] %s: EXTRA inner part %q", caseName, name, inner)
		}
	}
	for inner, w := range wantParts {
		g, ok := gotParts[inner]
		if !ok {
			continue
		}
		if inner == "docProps/core.xml" {
			if normalizeCore(g) != normalizeCore(w) {
				t.Errorf("[%s] %s: inner %q mismatch (timestamp-normalized)", caseName, name, inner)
			}
			continue
		}
		if g != w {
			i := firstDiff(g, w)
			t.Errorf("[%s] %s: inner %q byte mismatch at offset %d\n got: ...%s...\nwant: ...%s...",
				caseName, name, inner, i, snippet(g, i), snippet(w, i))
		}
	}
}

// ---------------------------------------------------------------------------
// Font-embedding end-to-end.
// ---------------------------------------------------------------------------

// fakeSFNT returns minimal bytes that pass validateFontData (TrueType magic).
func fakeSFNT(tag byte) []byte {
	b := make([]byte, 16)
	b[0], b[1], b[2], b[3] = 0x00, 0x01, 0x00, 0x00 // 0x00010000 sfnt magic
	b[15] = tag                                     // make each variant distinct
	return b
}

func TestFontEmbeddingEndToEnd(t *testing.T) {
	p := New()
	p.nowFunc = goldenTS
	p.Title = "Fonts"
	slide := p.AddSlide()
	if err := slide.AddText([]TextProps{{Text: "Hi"}}, &TextPropsOptions{TextBaseProps: TextBaseProps{FontFace: "Lato"}}); err != nil {
		t.Fatal(err)
	}

	if err := p.EmbedFont(FontEmbedProps{Typeface: "Lato", Regular: fakeSFNT(1), Bold: fakeSFNT(2)}); err != nil {
		t.Fatalf("EmbedFont Lato: %v", err)
	}
	if err := p.EmbedFont(FontEmbedProps{Typeface: "Merri", Italic: fakeSFNT(3)}); err != nil {
		t.Fatalf("EmbedFont Merri: %v", err)
	}
	// Duplicate typeface must be rejected.
	if err := p.EmbedFont(FontEmbedProps{Typeface: "Lato", Regular: fakeSFNT(9)}); err == nil {
		t.Errorf("expected duplicate typeface rejection")
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	parts := unzipParts(t, data)

	// Cross-check numbering against embeddedFontVariants.
	vs := embeddedFontVariants(len(p.slides), p.embeddedFonts)
	if len(vs) != 3 {
		t.Fatalf("want 3 variants, got %d", len(vs))
	}
	wantBytes := map[int][]byte{1: fakeSFNT(1), 2: fakeSFNT(2), 3: fakeSFNT(3)}
	for _, v := range vs {
		partName := filepathJoinSlash("ppt", "fonts", "font"+itoa(v.fileNum)+".fntdata")
		got, ok := parts[partName]
		if !ok {
			t.Errorf("missing font part %q", partName)
			continue
		}
		if !bytes.Equal([]byte(got), wantBytes[v.fileNum]) {
			t.Errorf("font part %q bytes mismatch", partName)
		}
	}

	// [Content_Types].xml Default for fntdata.
	if ct := parts["[Content_Types].xml"]; !strings.Contains(ct, `<Default Extension="fntdata" ContentType="application/x-fontdata"/>`) {
		t.Errorf("content types missing fntdata Default")
	}

	// presentation.xml: embedTrueTypeFonts + embeddedFontLst with matching rIds.
	pres := parts["ppt/presentation.xml"]
	if !strings.Contains(pres, `embedTrueTypeFonts="1"`) {
		t.Errorf("presentation.xml missing embedTrueTypeFonts")
	}
	wantLst := `<p:embeddedFontLst><p:embeddedFont><p:font typeface="Lato"/><p:regular r:id="rId8"/><p:bold r:id="rId9"/></p:embeddedFont><p:embeddedFont><p:font typeface="Merri"/><p:italic r:id="rId10"/></p:embeddedFont></p:embeddedFontLst>`
	if !strings.Contains(pres, wantLst) {
		t.Errorf("presentation.xml embeddedFontLst mismatch")
	}

	// presentation.xml.rels: one font relationship per variant with matching rIds/targets.
	rels := parts["ppt/_rels/presentation.xml.rels"]
	for _, want := range []string{
		`<Relationship Id="rId8" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/font" Target="fonts/font1.fntdata"/>`,
		`<Relationship Id="rId9" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/font" Target="fonts/font2.fntdata"/>`,
		`<Relationship Id="rId10" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/font" Target="fonts/font3.fntdata"/>`,
	} {
		if !strings.Contains(rels, want) {
			t.Errorf("presentation.xml.rels missing %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// WriteFile round-trip + Compression=true.
// ---------------------------------------------------------------------------

func TestWriteFileRoundTrip(t *testing.T) {
	_chartCounter = 0
	p := buildCase01(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "deck") // extension appended by WriteFile
	if err := p.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	data, err := os.ReadFile(path + ".pptx")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	compareAgainstGolden(t, "01-basic", data)
}

func TestWriteCompression(t *testing.T) {
	_chartCounter = 0
	p := buildCase01(t)
	data, err := p.Write(&WriteProps{WriteBaseProps: WriteBaseProps{Compression: ptr(true)}})
	if err != nil {
		t.Fatalf("Write(compress): %v", err)
	}
	// Parts must still be byte-identical to golden after decompression.
	compareAgainstGolden(t, "01-basic", data)

	// And the archive must actually use DEFLATE for its parts.
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip open: %v", err)
	}
	sawDeflate := false
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		if f.Method == zip.Deflate {
			sawDeflate = true
		} else if f.Method == zip.Store {
			t.Errorf("part %q stored, expected deflate under Compression=true", f.Name)
		}
	}
	if !sawDeflate {
		t.Errorf("no deflate-compressed parts found")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func cp(c Coord) *Coord { return &c }

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func readGoldenDir(t *testing.T, caseName string) map[string]string {
	t.Helper()
	root := goldenPath(caseName)
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatalf("walk golden %s: %v", caseName, err)
	}
	return out
}

func firstDiff(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func snippet(s string, at int) string {
	lo := at - 60
	if lo < 0 {
		lo = 0
	}
	hi := at + 60
	if hi > len(s) {
		hi = len(s)
	}
	return s[lo:hi]
}

func filepathJoinSlash(parts ...string) string { return strings.Join(parts, "/") }
