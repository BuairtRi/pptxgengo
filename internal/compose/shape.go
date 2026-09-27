package compose

import (
	"fmt"
	"math"
)

// PatternSpec is the deliberately small editable DrawingML pattern vocabulary
// qualified for roadmap extension tails.
type PatternSpec struct {
	Preset     string `json:"preset"`
	Foreground string `json:"foreground"`
	Background string `json:"background"`
}

func validateShape(c CanvasSpec) error {
	switch c.Preset {
	case "rect", "homePlate", "star5", "rightArrow", "blockArc", "triangle":
	default:
		return fmt.Errorf("canvas shape %s has unsupported preset %q", c.ID, c.Preset)
	}
	if c.Text != "" || len(c.Paragraphs) != 0 || c.FontFace != "" || c.FontSizePt != 0 || c.Bold || c.Foreground != "" || c.ContrastBackground != "" || c.InsetX != 0 || c.InsetY != 0 || c.Align != "" || c.Valign != "" {
		return fmt.Errorf("canvas shape %s is shape-only; use a separate named text canvas", c.ID)
	}
	if c.LineWidthPt != 0 || c.AssetPath != "" || c.AssetSHA256 != "" || c.AltText != "" || c.ImageFit != "" || c.ImageCrop != nil || c.FocalX != nil || c.FocalY != nil {
		return fmt.Errorf("canvas shape %s contains fields for another canvas kind", c.ID)
	}
	if c.Pattern != nil {
		if c.Background != "" {
			return fmt.Errorf("canvas shape %s cannot combine solid background and pattern", c.ID)
		}
		if c.Pattern.Preset != "wdUpDiag" {
			return fmt.Errorf("canvas shape %s has unsupported pattern %q", c.ID, c.Pattern.Preset)
		}
		if err := validateConcreteShapeColor(c.Pattern.Foreground); err != nil {
			return fmt.Errorf("canvas shape %s pattern foreground: %w", c.ID, err)
		}
		if err := validateConcreteShapeColor(c.Pattern.Background); err != nil {
			return fmt.Errorf("canvas shape %s pattern background: %w", c.ID, err)
		}
	} else if err := validateConcreteShapeColor(c.Background); err != nil {
		return fmt.Errorf("canvas shape %s background: %w", c.ID, err)
	}
	if len(c.Adjustments) != 0 && c.Preset != "homePlate" && c.Preset != "rightArrow" {
		return fmt.Errorf("canvas shape %s preset %s does not support adjustments", c.ID, c.Preset)
	}
	if math.IsNaN(c.RotationDeg) || math.IsInf(c.RotationDeg, 0) || c.RotationDeg < -180 || c.RotationDeg > 180 {
		return fmt.Errorf("canvas shape %s rotation must be finite in [-180,180]", c.ID)
	}
	if c.Preset == "blockArc" {
		if c.RotationDeg != 0 || math.IsNaN(c.ArcStartDeg) || math.IsNaN(c.ArcEndDeg) || math.IsNaN(c.ArcThicknessRatio) || math.IsInf(c.ArcStartDeg, 0) || math.IsInf(c.ArcEndDeg, 0) || math.IsInf(c.ArcThicknessRatio, 0) || c.ArcStartDeg < 0 || c.ArcEndDeg > 360 || c.ArcStartDeg >= c.ArcEndDeg || c.ArcThicknessRatio < .15 || c.ArcThicknessRatio > .7 {
			return fmt.Errorf("canvas shape %s blockArc requires 0<=start<end<=360, thickness .15..0.70 and no rotation", c.ID)
		}
	} else if c.ArcStartDeg != 0 || c.ArcEndDeg != 0 || c.ArcThicknessRatio != 0 {
		return fmt.Errorf("canvas shape %s arc fields require blockArc", c.ID)
	}
	if c.RotationDeg != 0 && c.Preset != "triangle" {
		return fmt.Errorf("canvas shape %s rotation requires triangle", c.ID)
	}
	for name, value := range c.Adjustments {
		if name != "adj" || value < 0 || value > 100000 {
			return fmt.Errorf("canvas shape %s adjustments require adj in [0,100000]", c.ID)
		}
	}
	return nil
}

func validateConcreteShapeColor(color string) error {
	if len(color) != 7 || color[0] != '#' {
		return fmt.Errorf("expected concrete #RRGGBB color")
	}
	_, err := resolveColor(color)
	return err
}

func resolvePattern(p *PatternSpec) *PatternSpec {
	if p == nil {
		return nil
	}
	out := *p
	out.Foreground, _ = resolveColor(p.Foreground)
	out.Background, _ = resolveColor(p.Background)
	return &out
}
