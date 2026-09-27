package main

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/pptx"
)

// imageCrop calculates source fractions without modifying the pinned image bytes.
func imageCrop(e element, width, height int) (*compose.ImageCropSpec, error) {
	return imageCropAspect(e, float64(width), float64(height))
}

func imageCropAspect(e element, width, height float64) (*compose.ImageCropSpec, error) {
	if width <= 0 || height <= 0 || e.Frame.Width <= 0 || e.Frame.Height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions")
	}
	mode := e.ImageFit
	if e.AssetMode == "stretch" && mode == "" {
		mode = "stretch"
	}
	ratio := width / height
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
	width, height := 0.0, 0.0
	format := ""
	isSVG := strings.HasSuffix(strings.ToLower(e.AssetPath), ".svg")
	if isSVG {
		width, height, err = svgIntrinsicSize(b)
		if err != nil {
			return fmt.Errorf("unsupported SVG %s: %w", e.AssetPath, err)
		}
	} else {
		cfg, detected, decodeErr := image.DecodeConfig(bytes.NewReader(b))
		if decodeErr != nil {
			return decodeErr
		}
		width, height, format = float64(cfg.Width), float64(cfg.Height), detected
	}
	crop, err := imageCropAspect(e, width, height)
	if err != nil {
		return err
	}
	o := &pptx.ImageProps{PositionProps: pos(e.Frame), ObjectNameProps: pptx.ObjectNameProps{ObjectName: e.Name}, DataOrPathProps: pptx.DataOrPathProps{Data: "image/" + format + ";base64," + base64.StdEncoding.EncodeToString(b)}, AltText: e.AltText}
	if isSVG {
		fallback, err := os.ReadFile(e.FallbackAssetPath)
		if err != nil {
			return fmt.Errorf("read SVG fallback: %w", err)
		}
		if len(fallback) > 50*1024*1024 || hash(fallback) != e.FallbackAssetSHA256 {
			return fmt.Errorf("SVG fallback hash mismatch: %s", e.FallbackAssetPath)
		}
		fallbackConfig, fallbackFormat, err := image.DecodeConfig(bytes.NewReader(fallback))
		if err != nil || fallbackFormat != "png" {
			return fmt.Errorf("SVG fallback must be a decodable PNG")
		}
		if math.Abs((float64(fallbackConfig.Width)/float64(fallbackConfig.Height))/(float64(width)/float64(height))-1) > .005 {
			return fmt.Errorf("SVG fallback aspect ratio differs from source")
		}
		o.Data = "image/svg+xml;base64," + base64.StdEncoding.EncodeToString(b)
		o.SVGFallbackData = "image/png;base64," + base64.StdEncoding.EncodeToString(fallback)
	}
	if e.OutlineColor != "" {
		o.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: e.OutlineColor}, Width: e.OutlineWidthPt}
	}
	if crop != nil {
		f := e.Frame
		f.Width /= 1 - crop.Left - crop.Right
		f.Height /= 1 - crop.Top - crop.Bottom
		o.PositionProps = pos(f)
		o.Sizing = &pptx.ImageSizing{Type: "crop", W: pptx.Inches(e.Frame.Width / 72), H: pptx.Inches(e.Frame.Height / 72), X: pointer(pptx.Inches(crop.Left * f.Width / 72)), Y: pointer(pptx.Inches(crop.Top * f.Height / 72))}
	}
	return s.AddImage(o)
}

// svgIntrinsicSize accepts self-contained SVG with an explicit positive width
// and height, or a positive viewBox. External references are rejected so the
// package contains every byte required to render the picture.
func svgIntrinsicSize(data []byte) (float64, float64, error) {
	d := xml.NewDecoder(bytes.NewReader(data))
	var stack []string
	roots := 0
	width, height := 0.0, 0.0
	for {
		token, err := d.Token()
		if err == io.EOF {
			if roots != 1 || len(stack) != 0 {
				return 0, 0, fmt.Errorf("SVG requires one complete root")
			}
			return width, height, nil
		}
		if err != nil {
			return 0, 0, err
		}
		switch v := token.(type) {
		case xml.StartElement:
			if v.Name.Space != "http://www.w3.org/2000/svg" || !staticSVGElement(v.Name.Local) {
				return 0, 0, fmt.Errorf("unsupported SVG element or namespace %s", v.Name.Local)
			}
			root := len(stack) == 0
			if root {
				roots++
				if roots != 1 || v.Name.Local != "svg" {
					return 0, 0, fmt.Errorf("SVG requires one svg root")
				}
			} else if v.Name.Local == "svg" {
				return 0, 0, fmt.Errorf("nested SVG unsupported")
			}
			attrs := map[string]string{}
			seen := map[xml.Name]bool{}
			for _, a := range v.Attr {
				if seen[a.Name] {
					return 0, 0, fmt.Errorf("duplicate SVG attribute %s", a.Name.Local)
				}
				seen[a.Name] = true
				if strings.HasPrefix(strings.ToLower(a.Name.Local), "on") || a.Name.Space == "http://www.w3.org/XML/1998/namespace" && a.Name.Local == "base" {
					return 0, 0, fmt.Errorf("active/external SVG attribute")
				}
				if a.Name.Local == "width" || a.Name.Local == "height" || a.Name.Local == "viewBox" {
					if a.Name.Space != "" {
						return 0, 0, fmt.Errorf("namespaced SVG sizing unsupported")
					}
					attrs[a.Name.Local] = a.Value
				}
				if err := validateSVGAttribute(a.Name.Local, a.Value); err != nil {
					return 0, 0, err
				}
			}
			if root {
				var err error
				width, height, err = svgRootDimensions(attrs)
				if err != nil {
					return 0, 0, err
				}
			}
			stack = append(stack, v.Name.Local)
		case xml.EndElement:
			if len(stack) == 0 {
				return 0, 0, fmt.Errorf("invalid SVG nesting")
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			value := strings.TrimSpace(string(v))
			if value == "" {
				continue
			}
			if len(stack) == 0 || stack[len(stack)-1] != "style" {
				return 0, 0, fmt.Errorf("unsupported SVG text")
			}
			if err := validateSVGStaticValue(value); err != nil {
				return 0, 0, err
			}
		case xml.ProcInst:
			if v.Target != "xml" || roots != 0 {
				return 0, 0, fmt.Errorf("SVG processing instruction unsupported")
			}
		case xml.Directive:
			return 0, 0, fmt.Errorf("SVG directive unsupported")
		}
	}
}

func svgRootDimensions(attrs map[string]string) (float64, float64, error) {
	var view [4]float64
	rawView, hasView := attrs["viewBox"]
	if hasView {
		fields := strings.Fields(strings.ReplaceAll(rawView, ",", " "))
		if len(fields) != 4 {
			return 0, 0, fmt.Errorf("invalid SVG viewBox")
		}
		for i, field := range fields {
			n, err := strconv.ParseFloat(field, 64)
			if err != nil || !finite(n) {
				return 0, 0, fmt.Errorf("invalid SVG viewBox")
			}
			view[i] = n
		}
		if view[2] <= 0 || view[3] <= 0 {
			return 0, 0, fmt.Errorf("invalid SVG viewBox size")
		}
	}
	rawW, hasW := attrs["width"]
	rawH, hasH := attrs["height"]
	if hasW || hasH {
		w, wok := svgDimension(rawW)
		h, hok := svgDimension(rawH)
		if !hasW || !hasH || !wok || !hok {
			return 0, 0, fmt.Errorf("SVG requires both positive unitless/px width and height")
		}
		return w, h, nil
	}
	if hasView {
		return view[2], view[3], nil
	}
	return 0, 0, fmt.Errorf("missing SVG intrinsic dimensions")
}

func staticSVGElement(name string) bool {
	return name == "svg" || name == "defs" || name == "g" || name == "path" || name == "style"
}

func validateSVGAttribute(name, value string) error {
	v := strings.ToLower(strings.TrimSpace(value))
	if (name == "href" || name == "src") && !strings.HasPrefix(v, "#") {
		return fmt.Errorf("external SVG reference")
	}
	return validateSVGStaticValue(v)
}

func validateSVGStaticValue(value string) error {
	value = strings.ToLower(value)
	// Restrict this lane to simple static source styles; escaped identifiers and
	// CSS at-rules require a real CSS parser and are deliberately unsupported.
	if strings.ContainsAny(value, "@\\") || strings.Contains(value, "/*") || !svgLocalURLsOnly(value) {
		return fmt.Errorf("unsupported SVG CSS/dependency syntax")
	}
	return nil
}

func svgLocalURLsOnly(v string) bool {
	for {
		i := strings.Index(v, "url(")
		if i < 0 {
			return true
		}
		v = v[i+4:]
		j := strings.IndexByte(v, ')')
		if j < 0 {
			return false
		}
		if !strings.HasPrefix(strings.TrimSpace(v[:j]), "#") {
			return false
		}
		v = v[j+1:]
	}
}

func svgDimension(v string) (float64, bool) {
	v = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), "px"))
	if v == "" || strings.Contains(v, "%") {
		return 0, false
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n <= 0 || math.IsInf(n, 0) || math.IsNaN(n) {
		return 0, false
	}
	return n, true
}
