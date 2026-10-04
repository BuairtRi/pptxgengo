package wmdesign

import (
	"math"
	"strings"
)

// Enlarge each circle 25% while keeping the complete diagram inside its body
// allocation. Horizontal separation is retained; vertical separation uses the
// available height. Split diagrams can spend the unused body space below them.
func intakeVennRoomierGeometry(n intakeVennSource, ctx SceneContext, radius float64, centers [][2]float64) (float64, [][2]float64, [2]float64) {
	area := Rect{n.X, n.Y, n.W, n.H}
	if strings.HasPrefix(ctx.Path, "/body/") && ctx.Zone.Y == n.Y && ctx.Zone.W >= n.W && ctx.Zone.W <= n.W+36 {
		area.H = math.Max(area.H, ctx.Zone.Y+ctx.Zone.H-n.Y)
	}
	x, y := area.X+area.W/2, area.Y+area.H/2
	rr := math.Min(radius*1.25, (area.H-2)/2)
	dx := math.Abs(centers[0][0] - (n.X + n.W/2))
	dx = math.Min(dx, (area.W-2)/2-rr)
	if len(centers) == 3 {
		dy := math.Min(radius*.66, (area.H-2-2*rr)/1.4)
		// The asymmetric triangular arrangement has a .15dy offset.
		y -= .15 * dy
		return rr, [][2]float64{{x - dx, y - dy*.55}, {x + dx, y - dy*.55}, {x, y + dy*.85}}, [2]float64{x, y}
	}
	dy := math.Min(radius*.62, (area.H-2)/2-rr)
	return rr, [][2]float64{{x - dx, y - dy}, {x + dx, y - dy}, {x - dx, y + dy}, {x + dx, y + dy}}, [2]float64{x, y}
}

type intakeVennTextFit struct {
	centers  [][2]float64
	radius   float64
	selected []int
	avoid    []Rect
}

// A text allocation must be inside every selected circle and outside the
// remaining circles. Testing complete rectangles protects wrapped lines,
// rather than fitting only a label's centre point.
func intakeVennTextRectInRegion(b Rect, fit intakeVennTextFit) bool {
	for _, obstacle := range fit.avoid {
		if math.Min(b.X+b.W, obstacle.X+obstacle.W) > math.Max(b.X, obstacle.X)-3 && math.Min(b.Y+b.H, obstacle.Y+obstacle.H) > math.Max(b.Y, obstacle.Y)-3 {
			return false
		}
	}
	selected := make(map[int]bool, len(fit.selected))
	for _, i := range fit.selected {
		selected[i] = true
	}
	for i, c := range fit.centers {
		if selected[i] {
			for _, x := range []float64{b.X, b.X + b.W} {
				for _, y := range []float64{b.Y, b.Y + b.H} {
					if math.Hypot(x-c[0], y-c[1]) > fit.radius-3 {
						return false
					}
				}
			}
		} else {
			x := math.Max(b.X, math.Min(c[0], b.X+b.W))
			y := math.Max(b.Y, math.Min(c[1], b.Y+b.H))
			if math.Hypot(x-c[0], y-c[1]) < fit.radius+3 {
				return false
			}
		}
	}
	return true
}

func intakeVennFitTextRect(anchor [2]float64, w, h float64, fit intakeVennTextFit) ([2]float64, bool) {
	valid := func(x, y float64) bool {
		return intakeVennTextRectInRegion(Rect{x - w/2, y - h/2, w, h}, fit)
	}
	if valid(anchor[0], anchor[1]) {
		return anchor, true
	}
	loX, hiX, loY, hiY := fit.centers[0][0], fit.centers[0][0], fit.centers[0][1], fit.centers[0][1]
	for _, c := range fit.centers {
		loX, hiX = math.Min(loX, c[0]), math.Max(hiX, c[0])
		loY, hiY = math.Min(loY, c[1]), math.Max(hiY, c[1])
	}
	loX, hiX, loY, hiY = loX-fit.radius, hiX+fit.radius, loY-fit.radius, hiY+fit.radius
	best, position := math.Inf(1), anchor
	// A 2pt search grid is finer than the native glyph padding reserve.
	for x := loX + w/2; x <= hiX-w/2; x += 2 {
		for y := loY + h/2; y <= hiY-h/2; y += 2 {
			if d := math.Hypot(x-anchor[0], y-anchor[1]); d < best && valid(x, y) {
				best, position = d, [2]float64{x, y}
			}
		}
	}
	return position, !math.IsInf(best, 1)
}

// Point membership carries meaning. Circle enlargement must not silently turn
// a pairwise initiative into an all-sets initiative. Move only points whose
// original membership changes, to the nearest matching region with marker room.
func intakeVennPreservePointRegion(x, y float64, original [][2]float64, oldRadius float64, centers [][2]float64, radius float64) (float64, float64) {
	mask := intakeVennPointMask(x, y, original, oldRadius)
	match := func(px, py, margin float64) bool {
		for i, c := range centers {
			d := math.Hypot(px-c[0], py-c[1])
			if mask&(1<<i) != 0 {
				if d > radius-margin {
					return false
				}
			} else if d < radius+margin {
				return false
			}
		}
		return true
	}
	if match(x, y, 10.5) {
		return x, y
	}
	loX, hiX, loY, hiY := centers[0][0], centers[0][0], centers[0][1], centers[0][1]
	for _, c := range centers {
		loX, hiX = math.Min(loX, c[0]), math.Max(hiX, c[0])
		loY, hiY = math.Min(loY, c[1]), math.Max(hiY, c[1])
	}
	loX, hiX, loY, hiY = loX-radius, hiX+radius, loY-radius, hiY+radius
	for _, margin := range []float64{10.5, 2, 0} {
		best, bx, by := math.Inf(1), x, y
		for i := 0; i <= 60; i++ {
			px := loX + (hiX-loX)*float64(i)/60
			for j := 0; j <= 60; j++ {
				py := loY + (hiY-loY)*float64(j)/60
				if d := math.Hypot(px-x, py-y); d < best && match(px, py, margin) {
					best, bx, by = d, px, py
				}
			}
		}
		if !math.IsInf(best, 1) {
			return bx, by
		}
	}
	return x, y
}

func intakeVennPointMask(x, y float64, centers [][2]float64, radius float64) int {
	mask := 0
	for i, c := range centers {
		if math.Hypot(x-c[0], y-c[1]) <= radius {
			mask |= 1 << i
		}
	}
	return mask
}
