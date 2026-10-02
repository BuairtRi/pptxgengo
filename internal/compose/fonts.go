package compose

import (
	"strings"
	"unicode"
)

// ValidFontFace accepts an explicit font family, rather than an inherited theme
// alias. Availability and requested styles are checked in the native environment.
func ValidFontFace(face string) bool {
	if face == "" || strings.TrimSpace(face) != face || strings.HasPrefix(face, "+") {
		return false
	}
	for _, ch := range face {
		if unicode.IsControl(ch) || ch == unicode.ReplacementChar {
			return false
		}
	}
	return true
}
