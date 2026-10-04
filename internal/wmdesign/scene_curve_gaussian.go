package wmdesign

import (
	"github.com/buairtri/pptxgengo/pptx"
	"math"
)

func teamCurvePolylinePath(points []curvePoint, move bool) []pptx.ShapePoint {
	out := make([]pptx.ShapePoint, len(points))
	for i, p := range points {
		out[i] = pptx.ShapePoint{X: pptx.Inches(p.x / 72), Y: pptx.Inches(p.y / 72)}
		if i == 0 {
			out[i].MoveTo = ptrSceneBool(move)
		}
	}
	return out
}

// Sample and blur cumulative edges, as the frozen Round 12 renderer does.
// A line series is drawn above the current base without changing that base.
func teamCurveGaussianEdges(at, values, base []float64, smooth float64, line bool) (hi, lo, next []float64) {
	next = append([]float64(nil), base...)
	top := make([]float64, 241)
	interval := 0
	for i := range top {
		u := at[0] + (at[len(at)-1]-at[0])*float64(i)/240
		for interval < len(at)-2 && u > at[interval+1] {
			interval++
		}
		f := math.Max(0, math.Min(1, (u-at[interval])/(at[interval+1]-at[interval])))
		top[i] = base[i] + values[interval] + (values[interval+1]-values[interval])*f
		if !line {
			next[i] = top[i]
		}
	}
	return teamCurveGaussianBlur(top, smooth), teamCurveGaussianBlur(base, smooth), next
}

func teamCurveGaussianBlur(values []float64, smooth float64) []float64 {
	sigma := smooth * 240
	// For subnormal sigma only the centre tap survives. Avoid division by an
	// underflowed sigma squared (0/0) and return the exact limiting result.
	if sigma < 1e-150 {
		return append([]float64(nil), values...)
	}
	radius := int(math.Ceil(sigma * 3))
	weights := make([]float64, 2*radius+1)
	sum := 0.
	for k := -radius; k <= radius; k++ {
		x := float64(k) / sigma
		weights[k+radius] = math.Exp(-x * x / 2)
		sum += weights[k+radius]
	}
	out := make([]float64, len(values))
	for i := range out {
		for k := -radius; k <= radius; k++ {
			out[i] += values[max(0, min(len(values)-1, i+k))] * weights[k+radius]
		}
		out[i] /= sum
	}
	return out
}
