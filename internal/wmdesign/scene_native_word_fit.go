package wmdesign

import "strings"

// PowerPoint emits the decimal separator for optional fractional placeholders
// even when all fractional digits are zero. Keep the requested precision, but
// make its fractional digits explicit. Quoted source units remain literal.
func v5NativeChartFormat(code string) string {
	var out strings.Builder
	quoted := false
	for i := 0; i < len(code); i++ {
		if code[i] == '"' {
			quoted = !quoted
		}
		if !quoted && strings.HasPrefix(code[i:], "0.#") {
			out.WriteString("0.")
			i += 2
			for i < len(code) && code[i] == '#' {
				out.WriteByte('0')
				i++
			}
			i--
			continue
		}
		out.WriteByte(code[i])
	}
	return out.String()
}

// Keep a source font and outer box while using a bounded six-point inset when
// twelve-point padding would force a word to break inside itself. The trailing
// reserve matches the existing native single-line textbox allocation policy.
func (r *renderer) v5WordInsets(text string, st Style, width, left, right float64) (float64, float64, error) {
	if !isV5OrLaterLibrary(r.source.Revision) || left != 12 || right != 12 || width > 126 || strings.Contains(text, "[^") {
		return left, right, nil
	}
	need := 0.0
	for _, word := range strings.Fields(text) {
		layout, err := r.measureText(word, st, 960)
		if err != nil {
			return left, right, err
		}
		if len(layout.Lines) > 0 {
			w := sequenceInlineWidth(layout.Lines[0].Advance, st)
			if w > need {
				need = w
			}
		}
	}
	if need > width-left-right && need <= width-12 {
		return 6, 6, nil
	}
	return left, right, nil
}
