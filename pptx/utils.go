// utils.go ports src/gen-utils.ts: numeric/color/XML helper functions plus the
// shared JS-numeric helpers required for byte-identical output.
package pptx

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// ptr returns a pointer to v. Shared helper for optional (pointer) fields.
func ptr[T any](v T) *T { return &v }

// ftoa formats a float the way JS `String(number)` does: shortest round-trip
// decimal, with integers printed without a decimal point.
func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// jsRound mirrors JS `Math.round`: rounds half toward +Infinity (unlike Go's
// math.Round, which rounds half away from zero — they differ for negatives).
func jsRound(f float64) float64 {
	return math.Floor(f + 0.5)
}

// getSmartParseNumber translates an x/y/w/h Coord to EMU.
//   - percentage: fraction of the layout width (X/default) or height (Y)
//   - inches (< 100): converted via inch2Emu
//   - >= 100: assumed already EMU, returned as-is
//
// Mirrors gen-utils.ts getSmartParseNumber.
func getSmartParseNumber(size Coord, xyDir string, layout PresLayout) int {
	// CASE 3: Percentage
	if size.IsPct {
		if xyDir == "Y" {
			return int(jsRound((size.Val / 100) * float64(layout.Height)))
		}
		// Default (including "X"): assume width
		return int(jsRound((size.Val / 100) * float64(layout.Width)))
	}

	// CASE 1: Number in inches (assume any number < 100 is inches)
	if size.Val < 100 {
		return inch2Emu(size.Val)
	}

	// CASE 2: Number already converted to EMU
	return int(size.Val)
}

// getUuid returns a UUID by replacing 'x'/'y' in uuidFormat with hex digits,
// mirroring the JS Math.random()-based implementation.
func getUuid(uuidFormat string) string {
	var b strings.Builder
	for _, c := range uuidFormat {
		switch c {
		case 'x':
			r := int(rand.Float64()*16) | 0
			b.WriteString(strconv.FormatInt(int64(r), 16))
		case 'y':
			r := int(rand.Float64()*16) | 0
			v := (r & 0x3) | 0x8
			b.WriteString(strconv.FormatInt(int64(v), 16))
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// encodeXmlEntities escapes XML special characters, in the exact replacement
// order of the JS implementation (& < > " ').
func encodeXmlEntities(xml string) string {
	xml = strings.ReplaceAll(xml, "&", "&amp;")
	xml = strings.ReplaceAll(xml, "<", "&lt;")
	xml = strings.ReplaceAll(xml, ">", "&gt;")
	xml = strings.ReplaceAll(xml, "\"", "&quot;")
	xml = strings.ReplaceAll(xml, "'", "&apos;")
	return xml
}

// inch2Emu converts inches to EMU. Values > 100 are assumed already EMU and
// returned unchanged (caller-safety, matching the JS behavior).
func inch2Emu(inches float64) int {
	if inches > 100 {
		return int(inches)
	}
	return int(jsRound(EMU * inches))
}

// valToPts converts a point value to EMU-points (ONEPT units).
func valToPts(pt float64) int {
	return int(jsRound(pt * ONEPT))
}

// convertRotationDegrees converts degrees (0..360) to a PowerPoint `rot` value.
func convertRotationDegrees(d float64) int {
	deg := d
	if deg > 360 {
		deg = deg - 360
	}
	return int(jsRound(deg * 60000))
}

// componentToHex converts an 8-bit component value to a 2-char lowercase hex.
func componentToHex(c int) string {
	hex := strconv.FormatInt(int64(c), 16)
	if len(hex) == 1 {
		return "0" + hex
	}
	return hex
}

// rgbToHex converts r/g/b components to an uppercase hex color string.
func rgbToHex(r, g, b int) string {
	return strings.ToUpper(componentToHex(r) + componentToHex(g) + componentToHex(b))
}

// isSchemeColor reports whether colorVal is one of the valid scheme colors.
func isSchemeColor(colorVal string) bool {
	switch colorVal {
	case string(SchemeColorText1), string(SchemeColorText2),
		string(SchemeColorBackground1), string(SchemeColorBackground2),
		string(SchemeColorAccent1), string(SchemeColorAccent2),
		string(SchemeColorAccent3), string(SchemeColorAccent4),
		string(SchemeColorAccent5), string(SchemeColorAccent6):
		return true
	}
	return false
}

// createColorElement builds an `a:srgbClr` (hex) or `a:schemeClr` (theme) XML
// element. Invalid input falls back to DEF_FONT_COLOR. innerElements, when
// non-empty, are wrapped inside the color element.
func createColorElement(colorStr, innerElements string) string {
	colorVal := strings.Replace(colorStr, "#", "", 1)

	if !RegexHexColor.MatchString(colorVal) && !isSchemeColor(colorVal) {
		// NOTE: JS logs a console.warn here; libraries should not log.
		colorVal = DEF_FONT_COLOR
	}

	isHex := RegexHexColor.MatchString(colorVal)
	var tagName, val string
	if isHex {
		tagName = "srgbClr"
		val = strings.ToUpper(colorVal)
	} else {
		tagName = "schemeClr"
		val = colorVal
	}
	colorAttr := "val=\"" + val + "\""

	if innerElements != "" {
		return "<a:" + tagName + " " + colorAttr + ">" + innerElements + "</a:" + tagName + ">"
	}
	return "<a:" + tagName + " " + colorAttr + "/>"
}

// createGlowElement builds an `a:glow` element, merging options over defaults
// (unset/zero option fields fall back to the corresponding default).
func createGlowElement(options, defaults TextGlowProps) string {
	opts := defaults
	if options.Size != 0 {
		opts.Size = options.Size
	}
	if options.Color != "" {
		opts.Color = options.Color
	}
	if options.Opacity != 0 {
		opts.Opacity = options.Opacity
	}

	size := int(jsRound(opts.Size * ONEPT))
	opacity := int(jsRound(opts.Opacity * 100000))

	var b strings.Builder
	b.WriteString("<a:glow rad=\"" + strconv.Itoa(size) + "\">")
	b.WriteString(createColorElement(opts.Color, "<a:alpha val=\""+strconv.Itoa(opacity)+"\"/>"))
	b.WriteString("</a:glow>")
	return b.String()
}

// genXmlColorSelection builds a fill color selection (`a:solidFill`). Pass a
// ShapeFillProps (for a plain color string, use &ShapeFillProps{Color: c}).
// Returns "" for a nil argument or a non-"solid" fill type.
func genXmlColorSelection(props *ShapeFillProps) string {
	if props == nil {
		return ""
	}

	fillType := "solid"
	colorVal := ""
	internalElements := ""

	if props.Type != "" {
		fillType = props.Type
	}
	if props.Color != "" {
		colorVal = props.Color
	}
	if props.Alpha != 0 { // DEPRECATED v3.3.0
		internalElements += "<a:alpha val=\"" + strconv.Itoa(int(jsRound((100-props.Alpha)*1000))) + "\"/>"
	}
	if props.Transparency != 0 {
		internalElements += "<a:alpha val=\"" + strconv.Itoa(int(jsRound((100-props.Transparency)*1000))) + "\"/>"
	}

	if fillType == "solid" {
		return "<a:solidFill>" + createColorElement(colorVal, internalElements) + "</a:solidFill>"
	}
	return ""
}

// getNewRelId returns the next relationship id (rId) for a slide.
func getNewRelId(target *PresSlide) int {
	return len(target.Rels) + len(target.RelsChart) + len(target.RelsMedia) + 1
}

// correctShadowOptions validates/normalizes shadow options in place and returns
// it (nil for a nil input). Mirrors gen-utils.ts correctShadowOptions.
func correctShadowOptions(shadow *ShadowProps) *ShadowProps {
	if shadow == nil {
		return nil
	}

	// OPT: type
	if shadow.Type != "outer" && shadow.Type != "inner" && shadow.Type != "none" {
		shadow.Type = "outer"
	}

	// OPT: angle
	if shadow.Angle != 0 {
		if shadow.Angle < 0 || shadow.Angle > 359 {
			shadow.Angle = 270
		}
		shadow.Angle = jsRound(shadow.Angle)
	}

	// OPT: opacity
	if shadow.Opacity != 0 {
		if shadow.Opacity < 0 || shadow.Opacity > 1 {
			shadow.Opacity = 0.75
		}
	}

	// OPT: color
	if shadow.Color != "" {
		if strings.HasPrefix(shadow.Color, "#") {
			shadow.Color = strings.Replace(shadow.Color, "#", "", 1)
		}
	}

	return shadow
}
