package wmdesign

import (
	"fmt"
	"math"
)

// RuleContract implements the quiet, decorative horizontal separator used by
// the frozen source renderer. It is a thin native rectangle, not an outer box.
const RuleContract = "pptxgengo.wmds-rule.v1"

func (r *renderer) planRule(n Node, b, zone Rect, surface string) (ShapeRecord, error) {
	if r.typeEngine.engine != CandidateEngine {
		return ShapeRecord{}, fmt.Errorf("rule.requires_v2: %s", n.ID)
	}
	if n.Text != "" || n.Style != "" || n.Align != "" || n.Ink != "" && n.Ink != "line" {
		return ShapeRecord{}, fmt.Errorf("rule.unsupported_fields: %s", n.ID)
	}
	if n.Grid == "five-up" || b.H != 0 && math.Abs(b.H-.75) > 1e-8 {
		return ShapeRecord{}, fmt.Errorf("rule.invalid_geometry: %s", n.ID)
	}
	if e := r.source.Tokens.Grid.OuterBox(Rect{X: b.X, Y: b.Y, W: b.W, H: 18}); e != nil {
		return ShapeRecord{}, fmt.Errorf("rule.off_lattice: %s: %w", n.ID, e)
	}
	b.H = .75
	if !inside(b, zone) {
		return ShapeRecord{}, fmt.Errorf("rule.outside_zone: %s", n.ID)
	}
	ink, e := r.source.Ink(surface, "line")
	if e != nil {
		return ShapeRecord{}, e
	}
	return ShapeRecord{ID: n.ID, Rect: b, Color: ink}, nil
}
