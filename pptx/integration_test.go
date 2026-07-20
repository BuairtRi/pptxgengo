package pptx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
		Shadow:        &ShadowProps{Type: "outer", Color: "000000", Opacity: ptr(0.5), Blur: ptr(3.0), Angle: ptr(45.0), Offset: ptr(3.0)},
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

	// Golden 06 was generated by the JS lib in the SAME process right after case
	// 05 built one chart, and JS's module-level _chartCounter never resets — so
	// 06's parts genuinely reference chart2.xml/chart3.xml (and
	// Microsoft_Excel_Worksheet2/3.xlsx). The Go port numbers charts
	// per-Presentation from 1 (REVIEW C2/M3 fix), so a fresh Presentation would
	// emit chart1/chart2. To reproduce the JS-era golden byte-for-byte we prime
	// this presentation's counter to 1 (unexported test hook; library consumers
	// always get deterministic 1-based numbering) so its first chart is chart2.
	p.chartCtr.n = 1

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

	// Pin the section GUID via the uuid hook (PORTING.md's promised uuidFunc,
	// now wired through the build context) to the exact value baked into the
	// golden's ppt/presentation.xml, so this case compares byte-for-byte with no
	// GUID normalization. getUuid's format arg is ignored by this fixed hook.
	p.uuidFunc = func(string) string { return "174ea01b-917e-75b4-41ad-6ea1d74d32aa" }

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
// Go API and compares every emitted part against the golden fixtures. Chart part
// numbering is now per-Presentation (REVIEW C2/M3), so no shared global counter
// is reset here; buildCase06 primes its own counter to reproduce the JS-era
// cross-presentation numbering baked into golden 06 (see that builder). Only two
// normalizations are permitted for the .pptx XML parts: (a) docProps/core.xml
// dcterms timestamps, (b) documented-random UUIDs (kept as a safety net —
// buildCase08 pins its section GUID so it already matches exactly). The embedded
// .xlsx workbook part is a nested ZIP whose byte encoding is produced by JSZip
// (not reproducible by archive/zip); it is therefore compared by recursively
// unzipping and matching its inner parts (its own core.xml timestamp
// normalized) — see the report note.
func TestGoldenIntegration(t *testing.T) {
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

// ---------------------------------------------------------------------------
// Concurrency + per-presentation-state regressions (REVIEW C1, C2, M3).
// ---------------------------------------------------------------------------

// chartDeck builds a single-slide deck carrying one bar chart, using the given
// clock. Shared by the concurrency tests.
func chartDeck(t *testing.T, now func() time.Time) *Presentation {
	t.Helper()
	p := New()
	p.nowFunc = now
	s := p.AddSlide()
	must(t, s.AddChart(ChartTypeBar,
		[]ChartData{{Name: "S", Labels: [][]string{{"A", "B"}}, Values: []float64{1, 2}}},
		&ChartOptions{}))
	return p
}

// coreTimestamp extracts docProps/core.xml's created timestamp from a package.
func coreTimestamp(t *testing.T, pkg []byte) string {
	t.Helper()
	core := unzipParts(t, pkg)["docProps/core.xml"]
	m := regexp.MustCompile(`<dcterms:created[^>]*>([^<]*)</dcterms:created>`).FindStringSubmatch(core)
	if m == nil {
		t.Fatalf("no created timestamp in core.xml:\n%s", core)
	}
	return m[1]
}

// TestConcurrentWriteRace writes two presentations with distinct pinned clocks
// concurrently and asserts each output carries ITS OWN timestamp. Under the old
// package-global clock-swap design (xmlNowFunc/excelNowFunc) the two writes
// raced and could cross-contaminate timestamps; -race catches the data race and
// this assertion catches the value bleed.
func TestConcurrentWriteRace(t *testing.T) {
	tsA := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	tsB := time.Date(2022, 2, 2, 3, 4, 5, 0, time.UTC)
	pA := chartDeck(t, func() time.Time { return tsA })
	pB := chartDeck(t, func() time.Time { return tsB })

	var wg sync.WaitGroup
	var outA, outB []byte
	var errA, errB error
	wg.Add(2)
	go func() { defer wg.Done(); outA, errA = pA.Write() }()
	go func() { defer wg.Done(); outB, errB = pB.Write() }()
	wg.Wait()
	must(t, errA)
	must(t, errB)

	if got := coreTimestamp(t, outA); got != "2001-01-01T00:00:00Z" {
		t.Errorf("deck A core.xml timestamp = %q, want its own 2001-01-01T00:00:00Z", got)
	}
	if got := coreTimestamp(t, outB); got != "2022-02-02T03:04:05Z" {
		t.Errorf("deck B core.xml timestamp = %q, want its own 2022-02-02T03:04:05Z", got)
	}
}

// TestConcurrentAddChartRace adds N charts across two presentations concurrently
// (each chart on its own pre-created slide, so the only shared state is each
// presentation's chart counter) and asserts numbering is exactly 1..N with no
// duplicates. Run under -race, this exercises the per-Presentation mutex on the
// counter.
func TestConcurrentAddChartRace(t *testing.T) {
	const n = 8
	build := func() *Presentation {
		p := New()
		p.nowFunc = goldenTS
		slides := make([]*Slide, n)
		for i := range slides {
			slides[i] = p.AddSlide() // pre-create; AddSlide itself isn't concurrent-safe
		}
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func(sl *Slide) {
				defer wg.Done()
				_ = sl.AddChart(ChartTypeBar,
					[]ChartData{{Name: "S", Labels: [][]string{{"A", "B"}}, Values: []float64{1, 2}}},
					&ChartOptions{})
			}(slides[i])
		}
		wg.Wait()
		return p
	}

	var wg sync.WaitGroup
	var pA, pB *Presentation
	wg.Add(2)
	go func() { defer wg.Done(); pA = build() }()
	go func() { defer wg.Done(); pB = build() }()
	wg.Wait()

	for _, tc := range []struct {
		name string
		p    *Presentation
	}{{"A", pA}, {"B", pB}} {
		seen := map[string]bool{}
		count := 0
		for _, ps := range tc.p.Slides() {
			for j := range ps.RelsChart {
				fn := ps.RelsChart[j].FileName
				if seen[fn] {
					t.Errorf("[%s] duplicate chart part name %q", tc.name, fn)
				}
				seen[fn] = true
				count++
			}
		}
		if count != n {
			t.Errorf("[%s] want %d charts, got %d", tc.name, n, count)
		}
		for i := 1; i <= n; i++ {
			if fn := fmt.Sprintf("chart%d.xml", i); !seen[fn] {
				t.Errorf("[%s] missing %s — numbering must be 1..N with no gaps", tc.name, fn)
			}
		}
		// The full package must also assemble with N distinct chart parts.
		data, err := tc.p.Write()
		must(t, err)
		parts := unzipParts(t, data)
		pkgCharts := 0
		for pn := range parts {
			if strings.HasPrefix(pn, "ppt/charts/chart") && strings.HasSuffix(pn, ".xml") {
				pkgCharts++
			}
		}
		if pkgCharts != n {
			t.Errorf("[%s] want %d chart parts in package, got %d", tc.name, n, pkgCharts)
		}
	}
}

// TestChartNumberingPerPresentation is the regression test for the leaking
// process-global counter (REVIEW C2/M3): building a chart deck, then a SECOND
// chart deck in the same process, must number the second deck's chart chart1.xml
// — not chart2.xml as the old global counter produced.
func TestChartNumberingPerPresentation(t *testing.T) {
	first := chartDeck(t, goldenTS)
	if _, err := first.Write(); err != nil { // advance the (per-pres) counter fully
		t.Fatal(err)
	}

	second := chartDeck(t, goldenTS)
	parts := unzipParts(t, mustWrite(t, second))

	if _, ok := parts["ppt/charts/chart1.xml"]; !ok {
		t.Errorf("second deck must number its chart chart1.xml (per-presentation counter)")
	}
	if _, ok := parts["ppt/charts/chart2.xml"]; ok {
		t.Errorf("second deck emitted chart2.xml — the chart counter leaked across presentations")
	}
	if _, ok := parts["ppt/embeddings/Microsoft_Excel_Worksheet1.xlsx"]; !ok {
		t.Errorf("second deck must number its workbook Microsoft_Excel_Worksheet1.xlsx")
	}
	if _, ok := parts["ppt/embeddings/Microsoft_Excel_Worksheet2.xlsx"]; ok {
		t.Errorf("second deck emitted Microsoft_Excel_Worksheet2.xlsx — counter leaked")
	}
}

func mustWrite(t *testing.T, p *Presentation) []byte {
	t.Helper()
	data, err := p.Write()
	must(t, err)
	return data
}
