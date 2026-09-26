package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/pptx"
)

// imageCrop calculates source fractions without modifying the pinned image bytes.
func imageCrop(e element, width, height int) (*compose.ImageCropSpec, error) {
	if width <= 0 || height <= 0 || e.Frame.Width <= 0 || e.Frame.Height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions")
	}
	mode := e.ImageFit
	if e.AssetMode == "stretch" && mode == "" {
		mode = "stretch"
	}
	ratio := float64(width) / float64(height)
	box := e.Frame.Width / e.Frame.Height
	switch mode {
	case "", "preserve":
		if math.Abs(box/ratio-1) > .005 {
			return nil, fmt.Errorf("image %s aspect ratio differs from source; use explicit contain/cover or a preserved-aspect frame", e.Name)
		}
		return nil, nil
	case "stretch":
		return nil, nil
	case "source_crop":
		if e.ImageCrop == nil {
			return nil, fmt.Errorf("image %s missing source crop", e.Name)
		}
		c := *e.ImageCrop
		visible := (1 - c.Left - c.Right) / (1 - c.Top - c.Bottom) * ratio
		if math.Abs(box/visible-1) > .005 {
			return nil, fmt.Errorf("image %s source crop would distort the image", e.Name)
		}
		return &c, nil
	case "contain":
		c := compose.ImageCropSpec{}
		if box > ratio {
			c.Left = (1 - box/ratio) / 2
			c.Right = c.Left
		} else {
			c.Top = (1 - ratio/box) / 2
			c.Bottom = c.Top
		}
		return &c, nil
	case "cover":
		x, y := .5, .5
		if e.FocalX != nil {
			x = *e.FocalX
		}
		if e.FocalY != nil {
			y = *e.FocalY
		}
		vw, vh := 1.0, 1.0
		if ratio > box {
			vw = box / ratio
		} else {
			vh = ratio / box
		}
		left := math.Max(0, math.Min(1-vw, x-vw/2))
		top := math.Max(0, math.Min(1-vh, y-vh/2))
		return &compose.ImageCropSpec{Left: left, Right: 1 - left - vw, Top: top, Bottom: 1 - top - vh}, nil
	default:
		return nil, fmt.Errorf("image %s unsupported fit mode %q", e.Name, mode)
	}
}

func renderImage(s *pptx.Slide, e element) error {
	b, err := os.ReadFile(e.AssetPath)
	if err != nil {
		return err
	}
	if len(b) > 50*1024*1024 {
		return fmt.Errorf("asset exceeds 50MB")
	}
	if hash(b) != e.AssetSHA256 {
		return fmt.Errorf("asset hash mismatch: %s", e.AssetPath)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return err
	}
	crop, err := imageCrop(e, cfg.Width, cfg.Height)
	if err != nil {
		return err
	}
	o := &pptx.ImageProps{PositionProps: pos(e.Frame), ObjectNameProps: pptx.ObjectNameProps{ObjectName: e.Name}, DataOrPathProps: pptx.DataOrPathProps{Data: "image/" + format + ";base64," + base64.StdEncoding.EncodeToString(b)}, AltText: e.AltText}
	if crop != nil {
		f := e.Frame
		f.Width /= 1 - crop.Left - crop.Right
		f.Height /= 1 - crop.Top - crop.Bottom
		o.PositionProps = pos(f)
		o.Sizing = &pptx.ImageSizing{Type: "crop", W: pptx.Inches(e.Frame.Width / 72), H: pptx.Inches(e.Frame.Height / 72), X: pointer(pptx.Inches(crop.Left * f.Width / 72)), Y: pointer(pptx.Inches(crop.Top * f.Height / 72))}
	}
	return s.AddImage(o)
}
