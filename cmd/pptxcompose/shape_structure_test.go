package main

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

func shapeZipPart(t *testing.T, deck []byte, name string) []byte {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if f.Name == name {
			r, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			b, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
	}
	t.Fatalf("missing zip part %s", name)
	return nil
}

func roadmapShapeSlide(width float64) renderSlide {
	return renderSlide{ID: "roadmap", Width: 960, Height: 540, Elements: []element{
		// Phase 01 and its first milestone use the audited slide-38 source
		// coordinates, flattened from the source group into editable siblings.
		{Name: "canvas:c29saWQ", Kind: "shape", Frame: frame{X: 460375.0 / 12700, Y: 2298215.0 / 12700, Width: width, Height: 502920.0 / 12700}, Preset: "homePlate", Adjustments: map[string]int{"adj": 39542}, Background: "CED7E6"},
		{Name: "canvas:dGFpbA", Kind: "shape", Frame: frame{X: 2570185.0 / 12700, Y: 2298215.0 / 12700, Width: 930305.0 / 12700, Height: 502920.0 / 12700}, Preset: "homePlate", Adjustments: map[string]int{"adj": 39542}, Pattern: &compose.PatternSpec{Preset: "wdUpDiag", Foreground: "CED7E6", Background: "FFFFFF"}},
		{Name: "canvas:c3Rhcg", Kind: "shape", Frame: frame{X: 3467057.0 / 12700, Y: 2457358.0 / 12700, Width: 207336.0 / 12700, Height: 212513.0 / 12700}, Preset: "star5", Background: "F900D3"},
	}}
}

func TestRoadmapShapesEmitDeterministicEditableGeometry(t *testing.T) {
	for _, width := range []float64{2853602.0 / 12700, 280.125} {
		slide := roadmapShapeSlide(width)
		deck, err := render([]renderSlide{slide})
		if err != nil {
			t.Fatal(err)
		}
		if err := validateShapeStructure(deck, []renderSlide{slide}); err != nil {
			t.Fatal(err)
		}
		xml := shapeZipPart(t, deck, "ppt/slides/slide1.xml")
		for _, want := range []string{`<a:prstGeom prst="homePlate"><a:avLst><a:gd name="adj" fmla="val 39542"/>`, `<a:pattFill prst="wdUpDiag"><a:fgClr><a:srgbClr val="CED7E6"/></a:fgClr><a:bgClr><a:srgbClr val="FFFFFF"/></a:bgClr></a:pattFill>`, `<a:prstGeom prst="star5"><a:avLst></a:avLst></a:prstGeom>`} {
			if !bytes.Contains(xml, []byte(want)) {
				t.Fatalf("missing editable shape XML %q", want)
			}
		}
		if width == 2853602.0/12700 {
			for _, want := range []string{`<a:off x="460375" y="2298215"/>`, `<a:ext cx="2853602" cy="502920"/>`, `<a:off x="2570185" y="2298215"/>`, `<a:ext cx="930305" cy="502920"/>`, `<a:off x="3467057" y="2457358"/>`, `<a:ext cx="207336" cy="212513"/>`} {
				if !bytes.Contains(xml, []byte(want)) {
					t.Fatalf("source geometry missing %q", want)
				}
			}
		}
	}
}

func TestRoadmapShapeStructuralValidatorRejectsSemanticChanges(t *testing.T) {
	slide := roadmapShapeSlide(2853602.0 / 12700)
	deck, err := render([]renderSlide{slide})
	if err != nil {
		t.Fatal(err)
	}
	patterns := roadmapShapeSlide(2853602.0 / 12700)
	p := *patterns.Elements[1].Pattern
	p.Preset = "smCheck"
	patterns.Elements[1].Pattern = &p
	adjustment := roadmapShapeSlide(2853602.0 / 12700)
	adjustment.Elements[0].Adjustments = map[string]int{"adj": 39543}
	preset := roadmapShapeSlide(2853602.0 / 12700)
	preset.Elements[2].Preset = "star6"
	for _, changed := range []renderSlide{patterns, adjustment, preset} {
		if err := validateShapeStructure(deck, []renderSlide{changed}); err == nil {
			t.Fatal("changed shape contract accepted against original deck")
		}
	}
}

func TestRoadmapShapeStructuralValidatorRejectsTripleDuplicateNames(t *testing.T) {
	slide := roadmapShapeSlide(2853602.0 / 12700)
	duplicate := slide.Elements[0]
	slide.Elements = append(slide.Elements, duplicate, duplicate)
	deck, err := render([]renderSlide{slide})
	if err != nil {
		t.Fatal(err)
	}
	if err := validateShapeStructure(deck, []renderSlide{slide}); err == nil {
		t.Fatal("three native shapes with one stable name were accepted")
	}
}

func TestRoadmapShapeStructuralValidatorRejectsMalformedGuides(t *testing.T) {
	want := element{Kind: "shape", Frame: frame{Width: 1, Height: 1}, Preset: "homePlate", Adjustments: map[string]int{"adj": 39542}, Background: "CED7E6"}
	for _, guides := range []string{
		`<a:gd name="adj" fmla="val 39542 junk"/>`,
		`<a:gd name="adj" fmla="val 1"/><a:gd name="adj" fmla="val 39542"/>`,
		`<a:gd name="adj" fmla="val 39542"/><a:other/>`,
	} {
		raw := `<p:sp xmlns:p="p" xmlns:a="a"><p:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="12700" cy="12700"/></a:xfrm><a:prstGeom prst="homePlate"><a:avLst>` + guides + `</a:avLst></a:prstGeom><a:solidFill><a:srgbClr val="CED7E6"/></a:solidFill></p:spPr></p:sp>`
		n, err := nativepkg.Parse([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateShapeNode(n, want); err == nil {
			t.Fatalf("malformed guides accepted: %s", guides)
		}
	}
}
