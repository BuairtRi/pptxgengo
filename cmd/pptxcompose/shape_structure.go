package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

// validateShapeStructure binds the semantics PowerPoint's scripting API does
// not expose reliably: preset geometry, adjustment guides, and pattern fills.
func validateShapeStructure(deck []byte, slides []renderSlide) error {
	z, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		return fmt.Errorf("shape structure: %w", err)
	}
	parts := map[string][]byte{}
	wantedParts := map[string]bool{}
	for i := range slides {
		wantedParts["ppt/slides/slide"+strconv.Itoa(i+1)+".xml"] = true
	}
	for _, f := range z.File {
		if !wantedParts[f.Name] {
			continue
		}
		r, e := f.Open()
		if e != nil {
			return e
		}
		b, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			return e
		}
		parts[f.Name] = b
	}
	for i, slide := range slides {
		var expected []element
		for _, e := range slide.Elements {
			if e.Kind == "shape" {
				expected = append(expected, e)
			}
		}
		if len(expected) == 0 {
			continue
		}
		part := "ppt/slides/slide" + strconv.Itoa(i+1) + ".xml"
		root, e := nativepkg.Parse(parts[part])
		if e != nil {
			return fmt.Errorf("shape structure %s: %w", part, e)
		}
		objects := map[string]*nativepkg.Node{}
		objectCounts := map[string]int{}
		root.Walk(func(n *nativepkg.Node) {
			if localName(n.Name) == "sp" {
				if name := shapeName(n); name != "" {
					objectCounts[name]++
					if objectCounts[name] == 1 {
						objects[name] = n
					}
				}
			}
		})
		for _, want := range expected {
			n := objects[want.Name]
			if n == nil || objectCounts[want.Name] != 1 {
				return fmt.Errorf("shape structure %s missing/duplicate %s", part, want.Name)
			}
			if e := validateShapeNode(n, want); e != nil {
				return fmt.Errorf("shape structure %s %s: %w", part, want.Name, e)
			}
		}
	}
	return nil
}

func validateShapeNode(n *nativepkg.Node, want element) error {
	spPr := n.Child("spPr")
	if spPr == nil || spPr.Child("xfrm") == nil || spPr.Child("prstGeom") == nil {
		return fmt.Errorf("missing native preset geometry")
	}
	xfrm := spPr.Child("xfrm")
	off, ext := xfrm.Child("off"), xfrm.Child("ext")
	if off == nil || ext == nil {
		return fmt.Errorf("missing explicit transform")
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
		if err != nil || got != int64(math.Round(v.want*12700)) {
			return fmt.Errorf("transform mismatch: %q expected %.6fpt", v.got, v.want)
		}
	}
	geom := spPr.Child("prstGeom")
	if geom.Attr("prst") != want.Preset {
		return fmt.Errorf("preset %q expected %q", geom.Attr("prst"), want.Preset)
	}
	guides := map[string]int{}
	if av := geom.Child("avLst"); av != nil {
		for _, gd := range av.Children {
			if gd.Name == "" {
				continue
			}
			if localName(gd.Name) != "gd" {
				return fmt.Errorf("unexpected adjustment-list child %s", localName(gd.Name))
			}
			name := gd.Attr("name")
			formula := gd.Attr("fmla")
			if name == "" || len(formula) < 5 || formula[:4] != "val " {
				return fmt.Errorf("invalid adjustment guide")
			}
			value, err := strconv.Atoi(formula[4:])
			if err != nil || formula != "val "+strconv.Itoa(value) {
				return fmt.Errorf("invalid adjustment guide")
			}
			if _, duplicate := guides[name]; duplicate {
				return fmt.Errorf("duplicate adjustment guide %s", name)
			}
			guides[name] = value
		}
	}
	if len(guides) != len(want.Adjustments) {
		return fmt.Errorf("adjustment count %d expected %d", len(guides), len(want.Adjustments))
	}
	for name, value := range want.Adjustments {
		if guides[name] != value {
			return fmt.Errorf("adjustment %s=%d expected %d", name, guides[name], value)
		}
	}
	if want.Pattern != nil {
		p := spPr.Child("pattFill")
		if shapeFillCount(spPr) != 1 || p == nil || len(namedChildren(p)) != 2 || p.Attr("prst") != want.Pattern.Preset || shapeColor(p.Child("fgClr")) != color(want.Pattern.Foreground) || shapeColor(p.Child("bgClr")) != color(want.Pattern.Background) {
			return fmt.Errorf("pattern fill mismatch")
		}
		if spPr.Child("solidFill") != nil {
			return fmt.Errorf("pattern shape also contains solid fill")
		}
	} else {
		solid := spPr.Child("solidFill")
		if shapeFillCount(spPr) != 1 || solid == nil || shapeColor(solid) != want.Background || spPr.Child("pattFill") != nil {
			return fmt.Errorf("solid fill mismatch")
		}
	}
	return nil
}

func shapeColor(n *nativepkg.Node) string {
	if n == nil {
		return ""
	}
	children := namedChildren(n)
	if len(children) != 1 || len(namedChildren(children[0])) != 0 {
		return ""
	}
	if c := n.Child("srgbClr"); c != nil {
		return c.Attr("val")
	}
	if c := n.Child("schemeClr"); c != nil {
		return c.Attr("val")
	}
	return ""
}

func namedChildren(n *nativepkg.Node) []*nativepkg.Node {
	var out []*nativepkg.Node
	if n == nil {
		return out
	}
	for _, child := range n.Children {
		if child.Name != "" {
			out = append(out, child)
		}
	}
	return out
}

func shapeFillCount(spPr *nativepkg.Node) int {
	n := 0
	for _, name := range []string{"solidFill", "pattFill", "gradFill", "blipFill", "noFill", "grpFill"} {
		if spPr.Child(name) != nil {
			n++
		}
	}
	return n
}
