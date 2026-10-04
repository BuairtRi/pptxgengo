package textlayout

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
)

func fixtureEngine(t *testing.T) *EngineLayout {
	t.Helper()
	root := t.TempDir()
	for name, data := range map[string][]byte{"regular.ttf": goregular.TTF, "bold.ttf": gobold.TTF, "italic.ttf": goitalic.TTF, "bolditalic.ttf": gobolditalic.TTF} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	e, err := New([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func request(text string, width float64) compose.ProbeRequest {
	return compose.ProbeRequest{ID: "text", Text: text, TextWidthPt: width, FontFace: "Go", FontSizePt: 16, Align: "left"}
}
func measured(t *testing.T, e *EngineLayout, q compose.ProbeRequest) (compose.Measurement, RequestLayout) {
	t.Helper()
	r, err := e.Measure([]compose.ProbeRequest{q})
	if err != nil {
		t.Fatal(err)
	}
	if r.PowerPointVerified || r.Engine != Engine || len(r.Fonts) == 0 {
		t.Fatal("missing engine provenance")
	}
	return r.Measurements.ByRequestID[q.ID], r.Requests[q.ID]
}

func TestWrapHardBreakAndPointSizes(t *testing.T) {
	e := fixtureEngine(t)
	q := request("Modern slide templates use measured text.", 500)
	one, _ := measured(t, e, q)
	q.TextWidthPt = 120
	wrapped, detail := measured(t, e, q)
	if len(detail.Lines) < 2 || wrapped.RenderedHeightPt <= one.RenderedHeightPt || wrapped.RenderedWidthPt > 120.01 {
		t.Fatalf("bad wrap: %+v %+v", wrapped, detail)
	}
	q = request("First\r\n\r\nLast\n", 500)
	_, detail = measured(t, e, q)
	if len(detail.Lines) != 4 || detail.Lines[1].Text != "" || detail.Lines[3].Text != "" {
		t.Fatalf("hard breaks lost: %+v", detail)
	}
	q = request("Fractional font size", 500)
	q.FontSizePt = 16.5
	fraction, _ := measured(t, e, q)
	q.FontSizePt = 16
	regular, _ := measured(t, e, q)
	if math.Abs(fraction.RenderedWidthPt/regular.RenderedWidthPt-16.5/16) > .003 {
		t.Fatal("fractional size rounded to whole points")
	}
}

func TestWrapAtMeasuredBoundary(t *testing.T) {
	e := fixtureEngine(t)
	q := request("Measured boundary", 500)
	_, one := measured(t, e, q)
	width := one.Lines[0].AdvancePt
	q.TextWidthPt = width + .02
	_, at := measured(t, e, q)
	if len(at.Lines) != 1 {
		t.Fatal("text should fit at its measured advance")
	}
	q.TextWidthPt = width - .25
	_, below := measured(t, e, q)
	if len(below.Lines) != 2 {
		t.Fatal("text should wrap just below its measured advance")
	}
}

func TestStylesMixedRunsAndSpacing(t *testing.T) {
	e := fixtureEngine(t)
	q := request("", 300)
	for _, style := range []struct{ bold, italic bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		q.Paragraphs = append(q.Paragraphs, compose.ParagraphSpec{Align: "left", SpaceBeforePt: 2, SpaceAfterPt: 3, Runs: []compose.RunSpec{{Text: "Styled text", FontFace: "Go", FontSizePt: 16, Bold: style.bold, Italic: style.italic}, {Text: " with a smaller ending", FontFace: "Go", FontSizePt: 12}}})
	}
	m, detail := measured(t, e, q)
	if len(e.resolver.Records()) != 4 || len(detail.Lines) != 4 || m.RenderedHeightPt <= 20 {
		t.Fatalf("styles or paragraph boundaries lost: %+v", detail)
	}
	for _, record := range e.resolver.Records() {
		if len(record.SHA256) != 64 {
			t.Fatal("font bytes not fingerprinted")
		}
	}
	base := m.RenderedHeightPt
	q.Paragraphs[0].LineSpacingMultiple = 2
	twice, _ := measured(t, e, q)
	if twice.RenderedHeightPt <= base {
		t.Fatal("line spacing was ignored")
	}
}

func TestAlignmentAndBullets(t *testing.T) {
	e := fixtureEngine(t)
	q := request("Centered", 300)
	q.Align = "center"
	_, center := measured(t, e, q)
	if math.Abs(center.Lines[0].XPt-(300-center.Lines[0].AdvancePt)/2) > .01 {
		t.Fatal("center alignment ignored")
	}
	q.Align = "right"
	_, right := measured(t, e, q)
	if math.Abs(right.Lines[0].XPt+right.Lines[0].AdvancePt-300) > .01 {
		t.Fatal("right alignment ignored")
	}
	q.Paragraphs = []compose.ParagraphSpec{{Align: "left", Bullet: &compose.BulletSpec{Character: "•", MarginLeftPt: 24, HangingPt: 16}, Runs: []compose.RunSpec{{Text: strings.Repeat("Measured bullet text ", 5), FontFace: "Go", FontSizePt: 16}}}}
	m, bullets := measured(t, e, q)
	if len(bullets.Lines) < 2 || bullets.Lines[0].XPt != 24 || bullets.Lines[1].XPt != 24 || m.OffsetXPt != 8 || m.RenderedWidthPt > 300 {
		t.Fatalf("bad bullet layout: %+v %+v", m, bullets)
	}
}

func TestMissingFontsAndUnsupportedFeaturesFail(t *testing.T) {
	e := fixtureEngine(t)
	tests := []struct {
		q       compose.ProbeRequest
		message string
	}{
		{compose.ProbeRequest{ID: "missing", Text: "Text", TextWidthPt: 100, FontFace: "Imaginary Font", FontSizePt: 12}, "no font substitution"},
		{request("tab\tstop", 100), "tab stops"},
		{request("مرحبا", 100), "Latin scope"},
		{request("😊", 100), "no glyph fallback"},
	}
	q := request("Measured phrase", 100)
	q.PhraseRequests = []compose.PhraseRequest{{ID: "a", Phrase: "phrase"}}
	tests = append(tests, struct {
		q       compose.ProbeRequest
		message string
	}{q, "phrase bounds"})
	for _, tc := range tests {
		_, err := e.Measure([]compose.ProbeRequest{tc.q})
		if err == nil || !strings.Contains(err.Error(), tc.message) {
			t.Fatalf("expected %s, got %v", tc.message, err)
		}
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "regular.ttf"), goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	upright, err := New([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	q = request("Missing bold", 100)
	q.Bold = true
	if _, err = upright.Measure([]compose.ProbeRequest{q}); err == nil {
		t.Fatal("synthetic bold silently accepted")
	}
}

func TestVariableWeightSelection(t *testing.T) {
	c := candidate{description: font.Description{Family: "Example", Aspect: font.Aspect{Style: font.StyleNormal, Weight: 400, Stretch: 1}}, axes: []tables.VariationAxisRecord{{Tag: ot.MustNewTag("wght"), Minimum: 100, Default: 400, Maximum: 700}, {Tag: ot.MustNewTag("wdth"), Minimum: 75, Default: 100, Maximum: 100}}}
	axes, ok := selection(c, true, false)
	if !ok || axes["wght"] != 700 || axes["wdth"] != 100 {
		t.Fatalf("variable selection: %+v %v", axes, ok)
	}
	c.axes[0].Maximum = 600
	if _, ok = selection(c, true, false); ok {
		t.Fatal("out-of-range weight clamped silently")
	}
}

func TestLongWordAndAccents(t *testing.T) {
	e := fixtureEngine(t)
	_, d := measured(t, e, request(strings.Repeat("longword", 10), 60))
	if len(d.Lines) < 2 {
		t.Fatal("long word not wrapped")
	}
	m, d := measured(t, e, request("Café", 200))
	if m.RenderedWidthPt <= 0 || len(d.Lines) != 1 {
		t.Fatal("accented text layout failed")
	}
	if _, err := e.Measure([]compose.ProbeRequest{request("Cafe\u0301", 200)}); err == nil {
		t.Fatal("missing combining mark silently accepted")
	}
}

func TestDefaultResolverIncludesUserHomeFonts(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS HOME-based user font directory")
	}
	home := t.TempDir()
	dir := filepath.Join(home, "Library", "Fonts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "regular.ttf")
	if err := os.WriteFile(path, goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	r, err := NewResolver(nil)
	if err != nil {
		t.Fatal(err)
	}
	f, err := r.resolve("Go", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if f.record.File != path {
		t.Fatalf("HOME font fixture was not resolved exactly: %+v", f.record)
	}
}

func TestExplicitResolverRootsRemainIsolated(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, "Library", "Fonts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "regular.ttf"), goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	r, err := NewResolver([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.resolve("Go", false, false); err == nil {
		t.Fatal("explicit font directories leaked a HOME font")
	}
}

// This integration fixture uses the installed families used by the prototype
// review deck. Portable tests above use fonts bundled with golang.org/x/image.
func TestInstalledFontsShapingAndVariableStyles(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS installed-font integration fixture")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	plex := filepath.Join(home, "Library/Fonts/IBMPlexSans-VariableFont_wdth,wght.ttf")
	if _, err := os.Stat(plex); err != nil {
		t.Skip("IBM Plex Sans integration font is not installed")
	}
	e, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	q := request("AV", 500)
	q.FontFace = "Arial"
	_, pair := measured(t, e, q)
	q.Text = "A"
	_, a := measured(t, e, q)
	q.Text = "V"
	_, v := measured(t, e, q)
	if pair.Lines[0].AdvancePt >= a.Lines[0].AdvancePt+v.Lines[0].AdvancePt {
		t.Fatal("kerning was not applied")
	}
	q.FontFace = "IBM Plex Sans"
	q.Text = "office"
	_, ligatures := measured(t, e, q)
	if ligatures.GlyphCount >= len([]rune(q.Text)) {
		t.Fatal("ligatures were not applied")
	}
	q.Text = "Variable font styles"
	_, regular := measured(t, e, q)
	q.Bold = true
	_, bold := measured(t, e, q)
	if regular.Lines[0].AdvancePt == bold.Lines[0].AdvancePt {
		t.Fatal("variable weight did not affect advances")
	}
	q.Paragraphs = []compose.ParagraphSpec{{Align: "left", Runs: []compose.RunSpec{{Text: "Italic", FontFace: "IBM Plex Sans", FontSizePt: 16, Italic: true}, {Text: " and bold italic", FontFace: "IBM Plex Sans", FontSizePt: 16, Bold: true, Italic: true}}}}
	measured(t, e, q)
	for _, record := range e.resolver.Records() {
		if record.Family == "IBM Plex Sans" {
			wantPath := plex
			if strings.Contains(record.Style, "italic") {
				wantPath = filepath.Join(home, "Library/Fonts/IBMPlexSans-Italic-VariableFont_wdth,wght.ttf")
			}
			if record.File != wantPath {
				t.Fatalf("variable fixture resolved a different font file: %+v", record)
			}
			want := float32(400)
			if strings.Contains(record.Style, "bold") {
				want = 700
			}
			if record.Axes["wght"] != want {
				t.Fatalf("wrong variable instance: %+v", record)
			}
			if record.Axes["wdth"] != 100 {
				t.Fatalf("wrong variable width: %+v", record)
			}
		}
	}
	q.Paragraphs = nil
	q.Bold = false
	q.Text = "Cafe\u0301"
	measured(t, e, q)
	q.FontFace = "Helvetica Neue"
	q.Text = "Collection face"
	measured(t, e, q)
}
