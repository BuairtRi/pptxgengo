package wmdesign

import (
	"math"
	"strings"
)

// The positive-minimum heat maps use narrow score columns. Keep their authored
// font families and column geometry, but trade excess header padding for native
// width to keep a single-word axis label intact. Earlier heat maps retain their
// accepted allocation. Multiword headings retain ordinary wrapping.
func (r *renderer) v6HeatHeaderFit(c sceneTableColumn, st Style, minimum *float64) (Style, float64, float64, error) {
	if c.Min != nil {
		minimum = c.Min
	}
	if !isV6OrLaterLibrary(r.source.Revision) || c.Type != "heat" || minimum == nil || *minimum <= 0 || len(strings.Fields(c.Label)) != 1 {
		return st, 12, 12, nil
	}
	layout, err := r.measureText(c.Label, st, 960)
	if err != nil {
		return st, 12, 12, err
	}
	if len(layout.Lines) != 1 {
		return st, 12, 12, nil
	}
	need := sequenceInlineWidth(layout.Lines[0].Advance, st)
	if need <= c.Width-24 {
		return st, 12, 12, nil
	}
	// The 58–61pt axes cannot hold ten 9pt monospaced letters plus native
	// allowance. Only these long heat headings may use the existing 8pt label
	// size; tracking retains its authored em value. Never shrink below 8pt.
	if need > c.Width-2 && st.Size > 8 {
		fitted := st
		fitted.TrackingPt *= 8 / fitted.Size
		fitted.Size = 8
		layout, err = r.measureText(c.Label, fitted, 960)
		if err != nil {
			return st, 12, 12, err
		}
		if len(layout.Lines) == 1 {
			fittedNeed := sequenceInlineWidth(layout.Lines[0].Advance, fitted)
			if fittedNeed <= c.Width-2 {
				st, need = fitted, fittedNeed
			}
		}
	}
	if need <= c.Width-2 {
		inset := math.Max(1, (c.Width-need)/2)
		return st, inset, inset, nil
	}
	return st, 12, 12, nil
}
