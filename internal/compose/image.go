package compose

import "fmt"

// ImageCropSpec uses fractional source-image edges, matching OOXML srcRect.
// Negative edges add transparent extent (used by contain), not stretched pixels.
type ImageCropSpec struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
}

func validateImagePlacement(c CanvasSpec) error {
	switch c.ImageFit {
	case "", "preserve", "contain", "cover", "stretch", "source_crop":
	default:
		return fmt.Errorf("image %s has unsupported image_fit %q", c.ID, c.ImageFit)
	}
	if c.ImageFit == "source_crop" {
		if c.ImageCrop == nil {
			return fmt.Errorf("image %s source_crop requires image_crop", c.ID)
		}
		v := c.ImageCrop
		for _, x := range []float64{v.Left, v.Top, v.Right, v.Bottom} {
			if !finite(x) || x < -4 || x >= 1 {
				return fmt.Errorf("image %s crop edges must be finite in [-4,1)", c.ID)
			}
		}
		if v.Left+v.Right >= 1 || v.Top+v.Bottom >= 1 {
			return fmt.Errorf("image %s crop has no visible source area", c.ID)
		}
	} else if c.ImageCrop != nil {
		return fmt.Errorf("image %s image_crop requires source_crop mode", c.ID)
	}
	for _, v := range []*float64{c.FocalX, c.FocalY} {
		if v != nil && (!finite(*v) || *v < 0 || *v > 1 || c.ImageFit != "cover") {
			return fmt.Errorf("image %s focal coordinates require cover mode and values in [0,1]", c.ID)
		}
	}
	return nil
}
