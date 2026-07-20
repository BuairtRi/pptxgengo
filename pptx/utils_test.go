package pptx

import (
	"math"
	"regexp"
	"testing"
)

// testLayout matches a 16:9 slide (13.333in x 7.5in) in EMU.
var testLayout = PresLayout{Name: "LAYOUT_16x9", Width: 12192000, Height: 6858000}

func TestFtoa(t *testing.T) {
	// Expectations below are literal `String(number)` output captured from
	// Node.js (v22) for each input — see the agent report for the exact
	// commands. JS switches to exponential notation at |x| >= 1e21 and at
	// 0 < |x| < 1e-6; Go's strconv never does, and its 'e' verb zero-pads
	// single-digit exponents ("1e-07") where JS does not ("1e-7").
	cases := []struct {
		in   float64
		want string
	}{
		// Normal range: unchanged shortest round-trip decimal.
		{1, "1"},
		{1.5, "1.5"},
		{0.1, "0.1"},
		{100, "100"},
		{12700, "12700"},
		{2.25, "2.25"},
		{0, "0"},
		{-1.5, "-1.5"},
		{-1, "-1"},

		// -0 prints as "0" in JS (String(-0) === "0").
		{math.Copysign(0, -1), "0"},

		// Large integers just under the 1e21 exponential threshold still
		// print as plain decimal (shortest round-trip).
		{9.999e20, "999900000000000000000"},
		{1e20, "100000000000000000000"},
		{9.99999e20, "999999000000000000000"},

		// Number.MAX_SAFE_INTEGER (2^53-1) and one past it: still decimal.
		{9007199254740991, "9007199254740991"},
		{9007199254740992, "9007199254740992"},
		{123456789012345680000, "123456789012345680000"},

		// >= 1e21: JS exponential form, lowercase 'e', explicit '+', no
		// leading zeros in the exponent.
		{1e21, "1e+21"},
		{1.23e21, "1.23e+21"},
		{2e21, "2e+21"},
		{1.1e21, "1.1e+21"},
		{-1e21, "-1e+21"},
		{1e22, "1e+22"},
		{1e100, "1e+100"},
		{1.23456e23, "1.23456e+23"},
		{1.7976931348623157e308, "1.7976931348623157e+308"}, // math.MaxFloat64

		// Decimal boundary at 1e-6: still plain decimal (not < 1e-6).
		{1e-6, "0.000001"},
		{-1e-6, "-0.000001"},
		{1e-5, "0.00001"},

		// < 1e-6: JS exponential form with negative, unpadded exponent.
		{1e-7, "1e-7"},
		{1.5e-7, "1.5e-7"},
		{-1e-7, "-1e-7"},
		{9.999999e-7, "9.999999e-7"},
		{9.9999999e-7, "9.9999999e-7"},
		{1.234e-10, "1.234e-10"},
		{-0.0000001234, "-1.234e-7"},
		{5e-324, "5e-324"}, // math.SmallestNonzeroFloat64
	}
	for _, c := range cases {
		if got := ftoa(c.in); got != c.want {
			t.Errorf("ftoa(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestJsRound(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0.5, 1},
		{-0.5, 0}, // Math.floor(-0.5 + 0.5) = 0 (differs from math.Round = -1)
		{2.5, 3},
		{-2.5, -2}, // Math.floor(-2.5 + 0.5) = -2
		{1.4, 1},
		{1.6, 2},
		{-1.6, -2},
	}
	for _, c := range cases {
		if got := jsRound(c.in); got != c.want {
			t.Errorf("jsRound(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestInch2Emu(t *testing.T) {
	cases := []struct {
		in   float64
		want int
	}{
		{1, 914400},
		{0.5, 457200},
		{2, 1828800},
		{50, 45720000},
		{200, 200}, // > 100 assumed already EMU
		{914400, 914400},
	}
	for _, c := range cases {
		if got := inch2Emu(c.in); got != c.want {
			t.Errorf("inch2Emu(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestGetSmartParseNumber(t *testing.T) {
	cases := []struct {
		size Coord
		dir  string
		want int
	}{
		{Inches(1), "X", 914400},
		{Inches(0.5), "Y", 457200},
		{Inches(50), "X", 45720000},
		{Coord{Val: 5000000}, "X", 5000000}, // >= 100, already EMU
		{Percent(50), "X", 6096000},         // 0.5 * 12192000
		{Percent(50), "Y", 3429000},         // 0.5 * 6858000
		{Percent(25), "", 3048000},          // default -> width: 0.25 * 12192000

		// Negative values (M5/Minor boundary coverage).
		{Inches(-1), "X", -914400},
		{Percent(-50), "X", -6096000}, // jsRound(-0.5 * 12192000)
		{Percent(-25), "Y", -1714500}, // jsRound(-0.25 * 6858000)

		// Val==100 is the CASE-1/CASE-2 boundary: `size.Val < 100` is false
		// at exactly 100, so it falls into CASE 2 (already-EMU), NOT
		// inch2Emu. (inch2Emu's own >100 special-case is a different,
		// unrelated boundary at a different threshold.)
		{Coord{Val: 100}, "X", 100},
		{Coord{Val: 99.999}, "X", 91439086}, // just under 100 -> CASE 1 (inches), not CASE 2

		// Percent > 100 is valid (oversized/offset shapes) and not clamped.
		{Percent(150), "X", 18288000}, // 1.5 * 12192000
		{Percent(200), "Y", 13716000}, // 2.0 * 6858000
	}
	for _, c := range cases {
		if got := getSmartParseNumber(c.size, c.dir, testLayout); got != c.want {
			t.Errorf("getSmartParseNumber(%+v, %q) = %d, want %d", c.size, c.dir, got, c.want)
		}
	}
}

// TestGetSmartParseNumberCase2Rounding locks the CASE-2 ("already EMU")
// rounding behavior: TS returns the raw float unchanged (gen-utils.ts:29),
// but Go's getSmartParseNumber returns int, so it rounds (jsRound) rather
// than truncates to minimize divergence from the TS fractional value. See
// the comment on getSmartParseNumber's CASE 2 in utils.go for the documented
// residual deviation.
func TestGetSmartParseNumberCase2Rounding(t *testing.T) {
	// NOTE: CASE 2 only triggers for size.Val >= 100 (the same raw
	// comparison JS uses: `typeof size === 'number' && size < 100` routes
	// to CASE 1/inches). A negative "already-EMU" value like -5000000.6 is
	// still < 100, so both JS and Go treat it as CASE 1 (inches), not
	// CASE 2 — there is no negative-EMU rounding case to test here.
	cases := []struct {
		val  float64
		dir  string
		want int
	}{
		{5000000.6, "X", 5000001}, // rounds up, would truncate to 5000000
		{5000000.4, "X", 5000000}, // rounds down
		{5000000.5, "X", 5000001}, // .5 rounds toward +Inf (jsRound)
		{100.5, "X", 101},         // boundary value with a fraction
	}
	for _, c := range cases {
		if got := getSmartParseNumber(Coord{Val: c.val}, c.dir, testLayout); got != c.want {
			t.Errorf("getSmartParseNumber(Val:%v) = %d, want %d", c.val, got, c.want)
		}
	}
}

func TestValToPts(t *testing.T) {
	cases := []struct {
		in   float64
		want int
	}{
		{1, 12700},
		{2.5, 31750},
		{0, 0},
		{0.5, 6350},
	}
	for _, c := range cases {
		if got := valToPts(c.in); got != c.want {
			t.Errorf("valToPts(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestConvertRotationDegrees(t *testing.T) {
	cases := []struct {
		in   float64
		want int
	}{
		{0, 0},
		{90, 5400000},
		{45, 2700000},
		{400, 2400000},  // (400-360)*60000
		{-90, -5400000}, // negative preserved
		{360, 21600000},
	}
	for _, c := range cases {
		if got := convertRotationDegrees(c.in); got != c.want {
			t.Errorf("convertRotationDegrees(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestComponentToHex(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "00"},
		{15, "0f"},
		{16, "10"},
		{255, "ff"},
		{128, "80"},
	}
	for _, c := range cases {
		if got := componentToHex(c.in); got != c.want {
			t.Errorf("componentToHex(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRgbToHex(t *testing.T) {
	cases := []struct {
		r, g, b int
		want    string
	}{
		{255, 0, 0, "FF0000"},
		{0, 128, 255, "0080FF"},
		{0, 0, 0, "000000"},
		{255, 255, 255, "FFFFFF"},
	}
	for _, c := range cases {
		if got := rgbToHex(c.r, c.g, c.b); got != c.want {
			t.Errorf("rgbToHex(%d,%d,%d) = %q, want %q", c.r, c.g, c.b, got, c.want)
		}
	}
}

func TestEncodeXmlEntities(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`a&b<c>d"e'f`, `a&amp;b&lt;c&gt;d&quot;e&apos;f`},
		{"", ""},
		{"plain", "plain"},
		{"<>&", "&lt;&gt;&amp;"},
		{"a & b", "a &amp; b"},
	}
	for _, c := range cases {
		if got := encodeXmlEntities(c.in); got != c.want {
			t.Errorf("encodeXmlEntities(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCreateColorElement(t *testing.T) {
	cases := []struct {
		color string
		inner string
		want  string
	}{
		{"FF0000", "", `<a:srgbClr val="FF0000"/>`},
		{"#00ff00", "", `<a:srgbClr val="00FF00"/>`}, // strip #, uppercase
		{"ff0000", "", `<a:srgbClr val="FF0000"/>`},
		{"tx1", "", `<a:schemeClr val="tx1"/>`},
		{"bg2", "", `<a:schemeClr val="bg2"/>`},
		{"accent6", "", `<a:schemeClr val="accent6"/>`},
		{"notacolor", "", `<a:srgbClr val="000000"/>`}, // invalid -> default
		{"FF0000", `<a:alpha val="50000"/>`, `<a:srgbClr val="FF0000"><a:alpha val="50000"/></a:srgbClr>`},
		{"tx1", `<a:alpha val="50000"/>`, `<a:schemeClr val="tx1"><a:alpha val="50000"/></a:schemeClr>`},
	}
	for _, c := range cases {
		if got := createColorElement(c.color, c.inner); got != c.want {
			t.Errorf("createColorElement(%q,%q) = %q, want %q", c.color, c.inner, got, c.want)
		}
	}
}

func TestGenXmlColorSelection(t *testing.T) {
	cases := []struct {
		props *ShapeFillProps
		want  string
	}{
		{nil, ""},
		{&ShapeFillProps{Color: "FF0000"}, `<a:solidFill><a:srgbClr val="FF0000"/></a:solidFill>`},
		{&ShapeFillProps{Color: "FF0000", Transparency: 50}, `<a:solidFill><a:srgbClr val="FF0000"><a:alpha val="50000"/></a:srgbClr></a:solidFill>`},
		{&ShapeFillProps{Color: "FF0000", Alpha: 25}, `<a:solidFill><a:srgbClr val="FF0000"><a:alpha val="75000"/></a:srgbClr></a:solidFill>`},
		{&ShapeFillProps{Type: "none", Color: "FF0000"}, ""},
		{&ShapeFillProps{Color: "tx1"}, `<a:solidFill><a:schemeClr val="tx1"/></a:solidFill>`},
	}
	for i, c := range cases {
		if got := genXmlColorSelection(c.props); got != c.want {
			t.Errorf("case %d: genXmlColorSelection() = %q, want %q", i, got, c.want)
		}
	}
}

func TestCreateGlowElement(t *testing.T) {
	// options {Size:5}, default DEF_TEXT_GLOW{8, FFFFFF, 0.75}
	got := createGlowElement(TextGlowProps{Size: 5}, DEF_TEXT_GLOW)
	want := `<a:glow rad="63500"><a:srgbClr val="FFFFFF"><a:alpha val="75000"/></a:srgbClr></a:glow>`
	if got != want {
		t.Errorf("createGlowElement(size5) = %q, want %q", got, want)
	}

	// full override
	got = createGlowElement(TextGlowProps{Size: 8, Color: "FF0000", Opacity: 0.5}, DEF_TEXT_GLOW)
	want = `<a:glow rad="101600"><a:srgbClr val="FF0000"><a:alpha val="50000"/></a:srgbClr></a:glow>`
	if got != want {
		t.Errorf("createGlowElement(override) = %q, want %q", got, want)
	}
}

func TestGetUuid(t *testing.T) {
	format := "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for i := 0; i < 50; i++ {
		got := getUuid(format)
		if len(got) != len(format) {
			t.Fatalf("getUuid length = %d, want %d (%q)", len(got), len(format), got)
		}
		if !re.MatchString(got) {
			t.Fatalf("getUuid = %q does not match UUID pattern", got)
		}
	}
}

func TestCorrectShadowOptions(t *testing.T) {
	if correctShadowOptions(nil) != nil {
		t.Error("correctShadowOptions(nil) should return nil")
	}

	// bogus type -> outer
	s := correctShadowOptions(&ShadowProps{Type: "bogus"})
	if s.Type != "outer" {
		t.Errorf("bogus type -> %q, want outer", s.Type)
	}

	// angle out of range -> 270
	s = correctShadowOptions(&ShadowProps{Type: "outer", Angle: 400})
	if s.Angle != 270 {
		t.Errorf("angle 400 -> %v, want 270", s.Angle)
	}

	// angle rounded
	s = correctShadowOptions(&ShadowProps{Type: "outer", Angle: 12.7})
	if s.Angle != 13 {
		t.Errorf("angle 12.7 -> %v, want 13", s.Angle)
	}

	// angle 0 untouched
	s = correctShadowOptions(&ShadowProps{Type: "outer", Angle: 0})
	if s.Angle != 0 {
		t.Errorf("angle 0 -> %v, want 0", s.Angle)
	}

	// opacity out of range -> 0.75
	s = correctShadowOptions(&ShadowProps{Type: "outer", Opacity: 1.5})
	if s.Opacity != 0.75 {
		t.Errorf("opacity 1.5 -> %v, want 0.75", s.Opacity)
	}

	// color hash stripped
	s = correctShadowOptions(&ShadowProps{Type: "outer", Color: "#FF0000"})
	if s.Color != "FF0000" {
		t.Errorf("color #FF0000 -> %q, want FF0000", s.Color)
	}
}

func TestGetNewRelId(t *testing.T) {
	sl := &PresSlide{}
	sl.Rels = []SlideRel{{}, {}}
	sl.RelsChart = []SlideRelChart{{}}
	sl.RelsMedia = []SlideRelMedia{{}}
	if got := getNewRelId(sl); got != 5 { // 2 + 1 + 1 + 1
		t.Errorf("getNewRelId = %d, want 5", got)
	}
}

func TestPtr(t *testing.T) {
	p := ptr(true)
	if p == nil || *p != true {
		t.Error("ptr(true) failed")
	}
	pi := ptr(42)
	if *pi != 42 {
		t.Error("ptr(42) failed")
	}
}
