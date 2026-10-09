package deckproject

import (
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Only newly declared scale copy requires independent clear space. Native
// annotations and deliberate overlays in the original example remain valid;
// this is not a general ban on diagram/text overlap.
func staffingScaleContext(in DiagramInspection, nodeID string) error {
	for _, caption := range in.MeasuredText {
		if !strings.HasSuffix(caption.ID, ".scale") || !(strings.Contains(caption.ID, "."+nodeID+".") || strings.HasPrefix(caption.ID, nodeID+".")) {
			continue
		}
		prefix := strings.TrimSuffix(caption.ID, ".scale") + "."
		a := staffingTextInk(caption)
		for _, other := range in.MeasuredText {
			if other.ID == caption.ID || strings.HasPrefix(other.ID, prefix) {
				continue
			}
			b := staffingTextInk(other)
			if math.Min(a.X+a.W, b.X+b.W)-math.Max(a.X, b.X) > .02 && math.Min(a.Y+a.H, b.Y+b.H)-math.Max(a.Y, b.Y) > .02 {
				return fmt.Errorf("staffing scale caption %s intersects contextual text %s; explicitly relocate retained companion text before applying", caption.ID, other.ID)
			}
		}
	}
	return nil
}

func staffingTextInk(t wmdesign.TextRecord) wmdesign.Rect {
	w := 0.
	for _, line := range t.Layout.Lines {
		w = math.Max(w, line.Advance)
	}
	x := t.Rect.X
	if t.Align == "center" {
		x += (t.Rect.W - w) / 2
	} else if t.Align == "right" {
		x += t.Rect.W - w
	}
	return wmdesign.Rect{X: x, Y: t.Rect.Y + t.Layout.OccupiedTop, W: w, H: t.Layout.EstimatedOccupiedHeight}
}
