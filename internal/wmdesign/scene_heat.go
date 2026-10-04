package wmdesign

import (
	"fmt"
	"math"
)

// Frozen browser-source contract; later heat-map changes require a new review.
const SceneHeatContract = "pptxgengo.wmds-heat.v1"
const SceneHeatSourceSHA256 = "a7f1f046d38723cb86d6e0dbacd97886c183784df1e4949192322e74d6a33cf6"

func sceneHeat(value float64, maximum *float64, scale string) (fill, ink string, err error) {
	max := 4.0
	if maximum != nil {
		max = *maximum
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || math.IsNaN(max) || math.IsInf(max, 0) || max <= 0 {
		return "", "", fmt.Errorf("scene.heat_range: finite value and positive maximum required")
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
	} else if value > 0 {
		step = int(math.Floor(value/max*4 + .5))
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
