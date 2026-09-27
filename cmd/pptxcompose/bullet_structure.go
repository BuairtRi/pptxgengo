package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

// validateRichBulletStructure binds the editable OOXML paragraph and bullet
// geometry that PowerPoint's scripting model cannot report per paragraph.
func validateRichBulletStructure(deck []byte, slides []renderSlide) error {
	parts, err := imagePackageParts(deck)
	if err != nil {
		return err
	}
	for i, slide := range slides {
		var expected []element
		for _, e := range slide.Elements {
			if e.Kind == "text" && len(e.Paragraphs) > 0 {
				expected = append(expected, e)
			}
		}
		if len(expected) == 0 {
			continue
		}
		part := "ppt/slides/slide" + strconv.Itoa(i+1) + ".xml"
		root, err := nativepkg.Parse(parts[part])
		if err != nil {
			return fmt.Errorf("rich bullet structure %s: %w", part, err)
		}
		objects, counts := map[string]*nativepkg.Node{}, map[string]int{}
		root.Walk(func(n *nativepkg.Node) {
			if localName(n.Name) != "sp" {
				return
			}
			if name := shapeName(n); name != "" {
				counts[name]++
				if counts[name] == 1 {
					objects[name] = n
				}
			}
		})
		for _, want := range expected {
			n := objects[want.Name]
			if n == nil || counts[want.Name] != 1 {
				return fmt.Errorf("rich bullet structure %s missing/duplicate %s", part, want.Name)
			}
			if err := validateRichTextNode(n, want); err != nil {
				return fmt.Errorf("rich bullet structure %s %s: %w", part, want.Name, err)
			}
		}
	}
	return nil
}

func validateRichTextNode(shape *nativepkg.Node, want element) error {
	body := shape.Child("txBody")
	if body == nil {
		return fmt.Errorf("missing text body")
	}
	var paragraphs []*nativepkg.Node
	for _, child := range namedChildren(body) {
		if localName(child.Name) == "p" {
			paragraphs = append(paragraphs, child)
		}
	}
	if len(paragraphs) != len(want.Paragraphs) {
		return fmt.Errorf("paragraph count %d expected %d", len(paragraphs), len(want.Paragraphs))
	}
	for i, spec := range want.Paragraphs {
		p := paragraphs[i]
		if count := countNamed(p, "pPr"); count != 1 {
			return fmt.Errorf("paragraph %s requires exactly one pPr, got %d", spec.ID, count)
		}
		props := p.Child("pPr")
		align := map[string]string{"left": "l", "center": "ctr", "right": "r"}[spec.Align]
		if err := exactAttrs(props, map[string]string{"algn": align, "marL": paragraphMargin(spec), "indent": paragraphIndent(spec)}); err != nil {
			return fmt.Errorf("paragraph %s pPr: %w", spec.ID, err)
		}
		if spec.Bullet == nil {
			if countNamed(props, "buNone") != 1 || bulletChoiceCount(props) != 1 || bulletSizeCount(props) != 0 || bulletStyleDirectiveCount(props) != 0 || len(namedChildren(props.Child("buNone"))) != 0 {
				return fmt.Errorf("paragraph %s requires explicit unbulleted geometry", spec.ID)
			}
			continue
		}
		if countNamed(props, "buChar") != 1 || bulletChoiceCount(props) != 1 || bulletSizeCount(props) != 1 || countNamed(props, "buSzPct") != 1 || bulletStyleDirectiveCount(props) != 0 {
			return fmt.Errorf("paragraph %s requires one 100%% character bullet", spec.ID)
		}
		if err := exactAttrs(props.Child("buSzPct"), map[string]string{"val": "100000"}); err != nil || len(namedChildren(props.Child("buSzPct"))) != 0 {
			return fmt.Errorf("paragraph %s bullet size mismatch", spec.ID)
		}
		if err := exactAttrs(props.Child("buChar"), map[string]string{"char": spec.Bullet.Character}); err != nil || len(namedChildren(props.Child("buChar"))) != 0 {
			return fmt.Errorf("paragraph %s bullet character mismatch", spec.ID)
		}
	}
	return nil
}

func bulletChoiceCount(p *nativepkg.Node) int {
	n := 0
	for _, name := range []string{"buNone", "buChar", "buAutoNum", "buBlip"} {
		n += countNamed(p, name)
	}
	return n
}

func bulletSizeCount(p *nativepkg.Node) int {
	n := 0
	for _, name := range []string{"buSzPct", "buSzPts", "buSzTx"} {
		n += countNamed(p, name)
	}
	return n
}

func bulletStyleDirectiveCount(p *nativepkg.Node) int {
	n := 0
	for _, name := range []string{"buFont", "buFontTx", "buClr", "buClrTx"} {
		n += countNamed(p, name)
	}
	return n
}

func paragraphMargin(p compose.ParagraphSpec) string {
	if p.Bullet == nil {
		return "0"
	}
	return strconv.FormatInt(pointEMU(p.Bullet.MarginLeftPt), 10)
}

func paragraphIndent(p compose.ParagraphSpec) string {
	if p.Bullet == nil {
		return "0"
	}
	return strconv.FormatInt(-pointEMU(p.Bullet.HangingPt), 10)
}

func pointEMU(points float64) int64 {
	return int64(math.Floor(points*12700 + .5))
}

func exactAttrs(n *nativepkg.Node, expected map[string]string) error {
	if n == nil || len(n.Attrs) != len(expected) {
		return fmt.Errorf("attribute count %d expected %d", len(n.Attrs), len(expected))
	}
	seen := map[string]bool{}
	for _, attr := range n.Attrs {
		name := localName(attr.Name)
		want, ok := expected[name]
		if !ok || seen[name] || attr.Value != want {
			return fmt.Errorf("unexpected %s=%q", attr.Name, attr.Value)
		}
		seen[name] = true
	}
	return nil
}
