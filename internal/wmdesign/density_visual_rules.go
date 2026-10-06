package wmdesign

import (
	"math"
	"strings"
)

// These are executable source signatures, not prose notes or commit dates.
// The snapshot is hash-validated by Load before recognizing its contract.
func recognizesDensityVisualRules(raw []byte) bool {
	code := strings.Join(strings.Fields(string(raw)), " ")
	for _, signature := range []string{
		`var CALLOUT_WHITE = { display: 1, title: 1, heading: 1, stat: 1, "stat-sm": 1 };`,
		`if (surf === "callout") return CALLOUT_WHITE[style] ? r.display : r.primary; return LARGE[style] ? r.display : r.primary;`,
		`qm.style.fontSize = "max(" + P(40) + ", calc(var(--d-subhead-fs, " + P(18) + ") * 48 / 18))"`,
		`qm.style.lineHeight = "calc(max(" + P(40) + ", calc(var(--d-subhead-fs, " + P(18) + ") * 48 / 18)) * 0.75)"`,
		`mk.style.fontSize = "max(" + P(40) + ", calc(var(--d-heading-fs, " + P(24) + ") * 2.5))"`,
		`mk.style.lineHeight = "calc(max(" + P(40) + ", calc(var(--d-heading-fs, " + P(24) + ") * 2.5)) * 0.5)"`,
		`n.surface === "callout" ? rr2.primary : (n.surface === "inverse" || n.surface === "deep" ? "#F900D3" : rr2.emphasis)`,
	} {
		if !strings.Contains(code, signature) {
			return false
		}
	}
	return true
}

func (r *renderer) hasDensityVisualRules() bool {
	return r.source != nil && r.source.Tokens.Density != nil && r.source.densityVisualRules
}

func recognizesDensityContrastRules(raw []byte) bool {
	code := strings.Join(strings.Fields(string(raw)), " ")
	return strings.Contains(code, `styled("label", roles.emphasis, n.inflectionLabel || "Inflection", bg)`) &&
		strings.Contains(code, `grey ? { ring: "#97A4BA", ink: "#070154" }`)
}

func (r *renderer) hasDensityContrastRules() bool {
	return r.source != nil && r.source.Tokens.Density != nil && r.source.densityContrastRules
}

func (r *renderer) shapeTextInk(surface, token string) string {
	if !r.hasDensityVisualRules() {
		return "display"
	}
	large := token == "display" || token == "title" || token == "heading" || token == "stat" || token == "stat-sm"
	if surface != "callout" {
		large = large || token == "subhead" || token == "number"
	}
	if large {
		return "display"
	}
	return "primary"
}

func (r *renderer) chevronNumberInk(surface string) string {
	if !r.hasDensityVisualRules() {
		return "emphasis"
	}
	if surface == "callout" {
		return "primary"
	}
	if surface == "inverse" || surface == "deep" {
		return "#F900D3"
	}
	return "emphasis"
}

func (r *renderer) quoteMarkStyle(card bool) (Style, error) {
	st, err := r.sceneStyle("heading")
	if err != nil {
		return st, err
	}
	size, lead := 60., 30.
	if card {
		size, lead = 48, 36
	}
	if r.hasDensityVisualRules() {
		role, multiplier, leadingRatio := "heading", 2.5, .5
		if card {
			role, multiplier, leadingRatio = "subhead", 48./18., .75
		}
		text, err := r.sceneStyle(role)
		if err != nil {
			return st, err
		}
		size = math.Max(40, text.Size*multiplier)
		lead = size * leadingRatio
	}
	st, err = primitiveStyleSize(st, size)
	st.Leading = lead
	return st, err
}
