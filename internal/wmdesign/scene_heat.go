package wmdesign

import (
	"fmt"
	"math"
)

// Frozen browser-source contract; later heat-map changes require a new review.
const SceneHeatContract = "pptxgengo.wmds-heat.v1"
const SceneHeatSourceSHA256 = "a7f1f046d38723cb86d6e0dbacd97886c183784df1e4949192322e74d6a33cf6"
const SceneHeatMinimumContract = "pptxgengo.wmds-heat-minimum.v2"
const SceneHeatMinimumSourceSHA256 = "89278f99cf8f0574bfb3d64e63dae5738d660ab74385462340d44364075312b4"

func sceneHeatWarning(minimum *float64) string {
	contract, hash := SceneHeatContract, SceneHeatSourceSHA256
	if minimum != nil {
		contract, hash = SceneHeatMinimumContract, SceneHeatMinimumSourceSHA256
	}
	return contract + " source_sha256=" + hash + "; native specimen review pending"
}

func sceneHeat(value float64, maximum *float64, scale string) (fill, ink string, err error) {
	return sceneHeatDomain(value, nil, maximum, scale)
}

// Explicit minima belong to the new source contract. Nil preserves the exact
// zero-based ramp used by the frozen v5 table, block and legend renderers.
func sceneHeatDomain(value float64, minimum, maximum *float64, scale string) (fill, ink string, err error) {
	min := 0.0
	if minimum != nil {
		min = *minimum
	}
	max := 4.0
	if maximum != nil {
		max = *maximum
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || math.IsNaN(min) || math.IsInf(min, 0) || math.IsNaN(max) || math.IsInf(max, 0) || max <= min {
		return "", "", fmt.Errorf("scene.heat_range: finite value and finite minimum below maximum required")
	}
	if scale == "" {
		scale = "seq"
	}
	if scale != "seq" && scale != "risk" {
		return "", "", fmt.Errorf("scene.heat_scale: expected seq or risk")
	}
	// Saturate before dividing: finite extremes may overflow value/max.
	step := 0
	if value >= max {
		step = 4
	} else if value > min {
		span := max - min
		fraction := (value - min) / span
		if math.IsInf(span, 0) {
			fraction = (value/2 - min/2) / (max/2 - min/2)
		}
		step = int(math.Floor(fraction*4 + .5))
	}
	ramp := [5]string{"E8EEF8", "B9C9F0", "7C9BFF", "0047FF", "070154"}
	ink = "070154"
	if scale == "risk" {
		ramp = [5]string{"E8EEF8", "CED7E6", "F3D6EE", "FB7FE8", "F900D3"}
	} else if step >= 3 {
		ink = "FFFFFF"
	}
	return ramp[step], ink, nil
}
