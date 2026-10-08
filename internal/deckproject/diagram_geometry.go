package deckproject

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// A line's zero axis needs an allocation, not a visible change. One EMU
// matches the renderer's identity group extent. Explicit relative points keep
// its actual endpoints unchanged and make future placement edits effective.
func editableSceneAllocation(kind string, args map[string]any, b wmdesign.Rect) (wmdesign.Rect, error) {
	if math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) || b.W < 0 || b.H < 0 {
		return b, fmt.Errorf("invalid measured scene geometry")
	}
	if kind == "connector" {
		if b.W == 0 {
			b.W = 1.0 / 12700
		}
		if b.H == 0 {
			b.H = 1.0 / 12700
		}
		var points [][2]float64
		raw, err := json.Marshal(args["points"])
		if err != nil {
			return b, err
		}
		if err = json.Unmarshal(raw, &points); err != nil || len(points) < 2 {
			return b, fmt.Errorf("connector requires explicit coordinate pairs")
		}
		for i := range points {
			points[i][0] = (points[i][0] - b.X) / b.W
			points[i][1] = (points[i][1] - b.Y) / b.H
		}
		args["points"], args["point_space"] = points, "allocation"
	}
	if b.W <= 0 || b.H <= 0 {
		return b, fmt.Errorf("positive measured allocation required")
	}
	return b, nil
}
