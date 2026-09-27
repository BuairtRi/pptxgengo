package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"io"
	"math"
	"path"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

// validateImageStructure binds every final picture to its pinned media bytes
// and to the exact DrawingML crop/fill form derived from the image contract.
func validateImageStructure(deck []byte, slides []renderSlide) error {
	parts, err := imagePackageParts(deck)
	if err != nil {
		return err
	}
	for i, slide := range slides {
		var expected []element
		for _, e := range slide.Elements {
			if e.Kind == "image" {
				expected = append(expected, e)
			}
		}
		if len(expected) == 0 {
			continue
		}
		slidePart := "ppt/slides/slide" + strconv.Itoa(i+1) + ".xml"
		root, err := nativepkg.Parse(parts[slidePart])
		if err != nil {
			return fmt.Errorf("image structure %s: %w", slidePart, err)
		}
		objects := map[string]*nativepkg.Node{}
		counts := map[string]int{}
		root.Walk(func(n *nativepkg.Node) {
			if localName(n.Name) == "pic" {
				if name := shapeName(n); name != "" {
					counts[name]++
					if counts[name] == 1 {
						objects[name] = n
					}
				}
			}
		})
		for _, want := range expected {
			n := objects[want.Name]
			if n == nil || counts[want.Name] != 1 {
				return fmt.Errorf("image structure %s missing/duplicate %s", slidePart, want.Name)
			}
			if err := validateImageNode(parts, slidePart, n, want); err != nil {
				return fmt.Errorf("image structure %s %s: %w", slidePart, want.Name, err)
			}
		}
	}
	return nil
}

func imagePackageParts(deck []byte) (map[string][]byte, error) {
	z, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		return nil, fmt.Errorf("image structure: %w", err)
	}
	parts := make(map[string][]byte, len(z.File))
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if _, duplicate := parts[f.Name]; duplicate {
			return nil, fmt.Errorf("image structure: duplicate package part %s", f.Name)
		}
		r, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("image structure %s: %w", f.Name, err)
		}
		b, err := io.ReadAll(io.LimitReader(r, 64*1024*1024+1))
		r.Close()
		if err != nil {
			return nil, fmt.Errorf("image structure %s: %w", f.Name, err)
		}
		if len(b) > 64*1024*1024 {
			return nil, fmt.Errorf("image structure: package part exceeds 64MB")
		}
		parts[f.Name] = b
	}
	return parts, nil
}

func validateImageNode(parts map[string][]byte, slidePart string, n *nativepkg.Node, want element) error {
	spPr := n.Child("spPr")
	if spPr == nil || spPr.Child("xfrm") == nil {
		return fmt.Errorf("missing explicit picture transform")
	}
	xfrm := spPr.Child("xfrm")
	if err := validatePictureTransformRotation(xfrm, want.RotationDeg); err != nil {
		return err
	}
	off, ext := xfrm.Child("off"), xfrm.Child("ext")
	if off == nil || ext == nil {
		return fmt.Errorf("missing explicit picture frame")
	}
	values := []struct {
		got  string
		want float64
	}{
		{off.Attr("x"), want.Frame.X}, {off.Attr("y"), want.Frame.Y},
		{ext.Attr("cx"), want.Frame.Width}, {ext.Attr("cy"), want.Frame.Height},
	}
	for _, v := range values {
		got, err := strconv.ParseInt(v.got, 10, 64)
		if err != nil || got != int64(math.Floor(v.want*12700+0.5)) {
			return fmt.Errorf("picture transform mismatch: %q expected %.6fpt", v.got, v.want)
		}
	}

	fill := n.Child("blipFill")
	if fill == nil || countNamed(fill, "blip") != 1 || countNamed(fill, "stretch") != 1 || countNamed(fill, "tile") != 0 {
		return fmt.Errorf("invalid picture fill structure")
	}
	blip := fill.Child("blip")
	if blip == nil || len(blip.Attrs) != 1 || blip.Attr("r:embed") == "" {
		return fmt.Errorf("picture blip requires only a pinned embedded relationship")
	}
	media, err := relatedImage(parts, slidePart, blip.Attr("r:embed"))
	if err != nil {
		return err
	}
	width, height := 0.0, 0.0
	isSVG := strings.HasSuffix(strings.ToLower(want.AssetPath), ".svg")
	if isSVG {
		ext := childNamed(blip, "extLst")
		if ext == nil || len(namedChildren(blip)) != 1 || len(namedChildren(ext)) != 1 {
			return fmt.Errorf("invalid SVG picture extension")
		}
		extension := childNamed(ext, "ext")
		if extension == nil || len(extension.Attrs) != 1 || extension.Attr("uri") != "{96DAC541-7B7A-43D3-8B79-37D633B846F1}" {
			return fmt.Errorf("invalid SVG extension URI")
		}
		svgBlip := childNamed(extension, "svgBlip")
		if svgBlip == nil || svgBlip.Name != "asvg:svgBlip" || len(namedChildren(extension)) != 1 || len(namedChildren(svgBlip)) != 0 || len(svgBlip.Attrs) != 2 || svgBlip.Attr("xmlns:asvg") != "http://schemas.microsoft.com/office/drawing/2016/SVG/main" || svgBlip.Attr("r:embed") == "" {
			return fmt.Errorf("missing SVG blip extension")
		}
		svg, err := relatedImage(parts, slidePart, svgBlip.Attr("r:embed"))
		if err != nil || hash(svg) != want.AssetSHA256 {
			return fmt.Errorf("embedded SVG hash mismatch")
		}
		if width, height, err = svgIntrinsicSize(svg); err != nil {
			return fmt.Errorf("embedded SVG dimensions: %w", err)
		}
		if hash(media) != want.FallbackAssetSHA256 {
			return fmt.Errorf("embedded SVG fallback hash mismatch")
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(media))
		if err != nil || format != "png" {
			return fmt.Errorf("embedded SVG fallback decode")
		}
		if float64(cfg.Width)/float64(cfg.Height) != 0 && math.Abs((float64(cfg.Width)/float64(cfg.Height))/(width/height)-1) > .005 {
			return fmt.Errorf("embedded SVG fallback aspect mismatch")
		}
	} else {
		if len(namedChildren(blip)) != 0 {
			return fmt.Errorf("unsupported picture blip effects")
		}
		if hash(media) != want.AssetSHA256 {
			return fmt.Errorf("embedded media hash mismatch")
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(media))
		if err != nil {
			return fmt.Errorf("embedded media decode: %w", err)
		}
		width, height = float64(cfg.Width), float64(cfg.Height)
	}
	if err := validatePictureOutline(spPr, want); err != nil {
		return err
	}
	crop, err := imageCropAspect(want, width, height)
	if err != nil {
		return err
	}
	stretch := fill.Child("stretch")
	if crop == nil {
		if countNamed(fill, "srcRect") != 0 || len(namedChildren(stretch)) != 1 || countNamed(stretch, "fillRect") != 1 {
			return fmt.Errorf("uncropped image requires one fillRect and no srcRect")
		}
		if err := validateDefaultFillRect(stretch.Child("fillRect")); err != nil {
			return err
		}
		return nil
	}
	if countNamed(fill, "srcRect") != 1 || len(namedChildren(stretch)) != 0 {
		return fmt.Errorf("cropped image requires one srcRect and empty stretch")
	}
	src := fill.Child("srcRect")
	if len(namedChildren(src)) != 0 {
		return fmt.Errorf("srcRect must not contain child elements")
	}
	expected := map[string]int{
		"l": imagePercent(crop.Left), "t": imagePercent(crop.Top),
		"r": imagePercent(crop.Right), "b": imagePercent(crop.Bottom),
	}
	for name, wantValue := range expected {
		got, err := strconv.Atoi(src.Attr(name))
		if err != nil || src.Attr(name) != strconv.Itoa(got) || got != wantValue {
			return fmt.Errorf("srcRect %s=%q expected %d", name, src.Attr(name), wantValue)
		}
	}
	return nil
}

func validatePictureOutline(spPr *nativepkg.Node, want element) error {
	ln := childNamed(spPr, "ln")
	if countNamed(spPr, "ln") > 1 {
		return fmt.Errorf("duplicate picture outline")
	}
	if want.OutlineColor == "" {
		if ln != nil {
			return fmt.Errorf("unexpected picture outline")
		}
		return nil
	}
	if ln == nil || len(ln.Attrs) != 1 || ln.Attr("w") != strconv.Itoa(int(math.Floor(want.OutlineWidthPt*12700+0.5))) {
		return fmt.Errorf("picture outline width mismatch")
	}
	solid := childNamed(ln, "solidFill")
	if solid == nil || len(namedChildren(ln)) != 1 {
		return fmt.Errorf("picture outline color mismatch")
	}
	clr := childNamed(solid, "srgbClr")
	if clr == nil || len(clr.Attrs) != 1 || len(namedChildren(clr)) != 0 || clr.Attr("val") != want.OutlineColor || len(namedChildren(solid)) != 1 {
		return fmt.Errorf("picture outline color mismatch")
	}
	return nil
}

func validatePictureTransformAttrs(xfrm *nativepkg.Node) error {
	return validatePictureTransformRotation(xfrm, 0)
}

func validatePictureTransformRotation(xfrm *nativepkg.Node, expected float64) error {
	rotation := int64(0)
	for _, attr := range xfrm.Attrs {
		switch localName(attr.Name) {
		case "rot":
			v, err := strconv.ParseInt(attr.Value, 10, 64)
			if err != nil || strconv.FormatInt(v, 10) != attr.Value || math.Abs(float64(v)/60000-expected) > 0.00002 {
				return fmt.Errorf("picture rotation mismatch: %q expected %.6f", attr.Value, expected)
			}
			rotation = v
		case "flipH", "flipV":
			if attr.Value != "0" && attr.Value != "false" {
				return fmt.Errorf("picture flip is unsupported: %s=%q", localName(attr.Name), attr.Value)
			}
		}
	}
	if math.Abs(float64(rotation)/60000-expected) > 0.00002 {
		return fmt.Errorf("picture rotation missing/mismatched")
	}
	return nil
}

func validateDefaultFillRect(rect *nativepkg.Node) error {
	if rect == nil || len(namedChildren(rect)) != 0 {
		return fmt.Errorf("uncropped image requires an empty default fillRect")
	}
	for _, attr := range rect.Attrs {
		switch localName(attr.Name) {
		case "l", "t", "r", "b":
			if attr.Value != "0" {
				return fmt.Errorf("uncropped image fillRect %s must be zero", localName(attr.Name))
			}
		default:
			return fmt.Errorf("uncropped image fillRect has unsupported attribute %s", attr.Name)
		}
	}
	return nil
}

func relatedImage(parts map[string][]byte, slidePart, id string) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("picture relationship ID missing")
	}
	relsPart := path.Join(path.Dir(slidePart), "_rels", path.Base(slidePart)+".rels")
	rels, err := nativepkg.Parse(parts[relsPart])
	if err != nil {
		return nil, fmt.Errorf("picture relationships: %w", err)
	}
	var target string
	for _, rel := range rels.Children {
		if rel.Name == "" || rel.Attr("Id") != id {
			continue
		}
		if target != "" || rel.Attr("TargetMode") == "External" || !strings.HasSuffix(rel.Attr("Type"), "/image") {
			return nil, fmt.Errorf("invalid/duplicate picture relationship %s", id)
		}
		target = rel.Attr("Target")
	}
	if target == "" {
		return nil, fmt.Errorf("unresolved picture relationship %s", id)
	}
	if strings.HasPrefix(target, "/") {
		target = strings.TrimPrefix(target, "/")
	} else {
		target = path.Clean(path.Join(path.Dir(slidePart), target))
	}
	if target == "." || target == ".." || strings.HasPrefix(target, "../") {
		return nil, fmt.Errorf("picture relationship escapes package")
	}
	b, ok := parts[target]
	if !ok {
		return nil, fmt.Errorf("missing picture resource %s", target)
	}
	return b, nil
}

func imagePercent(v float64) int { return int(math.Floor(v*100000 + 0.5)) }

func countNamed(n *nativepkg.Node, name string) int {
	count := 0
	if n == nil {
		return count
	}
	for _, child := range n.Children {
		if localName(child.Name) == name {
			count++
		}
	}
	return count
}

func childNamed(n *nativepkg.Node, name string) *nativepkg.Node {
	if n == nil {
		return nil
	}
	for _, child := range n.Children {
		if localName(child.Name) == name {
			return child
		}
	}
	return nil
}

func nodeNames(n *nativepkg.Node) []string {
	var out []string
	if n != nil {
		for _, c := range n.Children {
			if c.Name != "" {
				out = append(out, c.Name)
			}
		}
	}
	return out
}
