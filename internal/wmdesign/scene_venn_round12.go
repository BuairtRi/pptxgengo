package wmdesign

import "math"

// Match the frozen renderer's 301-sample search for the exclusive overlap
// region. An all-sets centre remains the mean of its circle centres.
func intakeVennRound12Anchor(centers [][2]float64, indices []int, center [2]float64, radius float64) [2]float64 {
	p := [2]float64{}
	selected := make(map[int]bool, len(indices))
	for _, i := range indices {
		p[0] += centers[i][0]
		p[1] += centers[i][1]
		selected[i] = true
	}
	p[0] /= float64(len(indices))
	p[1] /= float64(len(indices))
	dx, dy := p[0]-center[0], p[1]-center[1]
	length := math.Hypot(dx, dy)
	if len(indices) == len(centers) || length <= 1 {
		return p
	}
	lo, hi, found := 0., 0., false
	for s := -100; s <= 200; s++ {
		t := float64(s) / 100 * radius
		x, y := p[0]+dx/length*t, p[1]+dy/length*t
		ok := true
		for i, c := range centers {
			if (math.Hypot(x-c[0], y-c[1]) <= radius) != selected[i] {
				ok = false
				break
			}
		}
		if ok {
			if !found {
				lo = t
			}
			hi, found = t, true
		}
	}
	if found {
		p[0] += dx / length * (lo + hi) / 2
		p[1] += dy / length * (lo + hi) / 2
	}
	return p
}
