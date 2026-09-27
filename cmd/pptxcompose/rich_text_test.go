package main

import (
	"bytes"
	"github.com/buairtri/pptxgengo/internal/compose"
	"testing"
)

func TestRichContractAndNativeStyleCheck(t *testing.T) {
	p := []compose.ParagraphSpec{{ID: "p", Align: "left", SpaceAfterPt: 4, Runs: []compose.RunSpec{{ID: "a", Text: "A", FontFace: "Arial", FontSizePt: 12, Bold: true, Foreground: "#070154"}, {ID: "b", Text: "b", FontFace: "Arial", FontSizePt: 11, Italic: true, Underline: true, Foreground: "#0047FF"}}}}
	row := nativeRow{Paragraphs: []nativeParagraph{{Alignment: "paragraph align left", SpaceAfterPt: 4, LineRuleWithin: true, SpaceWithin: 1, BulletVisible: pointer(false), BulletType: "ppBulletNone"}}, Characters: []nativeCharacter{{Text: "A", FontName: "Arial", FontSizePt: 12, Bold: true, Underline: "no underline", Color: []int{7, 1, 84}}, {Text: "b", FontName: "Arial", FontSizePt: 11, Italic: true, Underline: "underline single line", Color: []int{0, 71, 255}}}}
	if err := checkNativeRichText("shape", p, row); err != nil {
		t.Fatal(err)
	}
	row.Characters[1].Italic = false
	if err := checkNativeRichText("shape", p, row); err == nil {
		t.Fatal("native rich style mismatch accepted")
	}
	q := compose.ProbeRequest{ID: "q", Text: "Ab", Paragraphs: p, TextWidthPt: 100, Background: "#FFFFFF"}
	key := contractKey(q)
	q.Paragraphs[0].Runs[1].Italic = false
	if contractKey(q) == key {
		t.Fatal("rich run style omitted from cache contract")
	}
	q.Paragraphs[0].LineSpacingMultiple = .9
	if contractKey(q) == key {
		t.Fatal("paragraph line spacing omitted from cache contract")
	}
	row.Paragraphs[0].SpaceWithin = .9
	p[0].LineSpacingMultiple = 1
	if err := checkNativeRichText("shape", p, row); err == nil {
		t.Fatal("native paragraph line-spacing mismatch accepted")
	}
}

func TestRichBulletNativeContractAndCacheBinding(t *testing.T) {
	p := []compose.ParagraphSpec{{ID: "bullet", Align: "left", Bullet: &compose.BulletSpec{Character: "•", MarginLeftPt: 12, HangingPt: 9}, Runs: []compose.RunSpec{{ID: "r", Text: "Editable bullet", FontFace: "Arial", FontSizePt: 12, Foreground: "#070154"}}}}
	row := nativeRow{Paragraphs: []nativeParagraph{{Alignment: "paragraph align left", LineRuleWithin: true, SpaceWithin: 1, BulletVisible: pointer(true), BulletType: "ppBulletUnnumbered", BulletCharacter: "•", BulletRelativeSize: pointer(1.0), BulletUseTextColor: pointer(true), BulletUseTextFont: pointer(true)}}, Characters: []nativeCharacter{}}
	for _, ch := range "Editable bullet" {
		row.Characters = append(row.Characters, nativeCharacter{Text: string(ch), FontName: "Arial", FontSizePt: 12, Underline: "no underline", Color: []int{7, 1, 84}})
	}
	if err := checkNativeRichText("shape", p, row); err != nil {
		t.Fatal(err)
	}
	key := contractKey(compose.ProbeRequest{Text: "Editable bullet", Paragraphs: p, TextWidthPt: 100})
	p[0].Bullet.HangingPt = 10
	if contractKey(compose.ProbeRequest{Text: "Editable bullet", Paragraphs: p, TextWidthPt: 100}) == key {
		t.Fatal("bullet geometry omitted from cache contract")
	}
	p[0].Bullet.HangingPt = 9
	row.Paragraphs[0].BulletUseTextColor = pointer(false)
	if err := checkNativeRichText("shape", p, row); err == nil {
		t.Fatal("native bullet color inheritance mismatch accepted")
	}
	row.Paragraphs[0].BulletUseTextColor = pointer(true)
	row.Paragraphs[0].BulletCharacter = "–"
	if err := checkNativeRichText("shape", p, row); err == nil {
		t.Fatal("native bullet character mismatch accepted")
	}
}

func TestRichBulletRenderingAndStructure(t *testing.T) {
	p := compose.ParagraphSpec{ID: "bullet", Align: "left", Bullet: &compose.BulletSpec{Character: "•", MarginLeftPt: 12, HangingPt: 9}, Runs: []compose.RunSpec{{ID: "plain", Text: "Editable ", FontFace: "Arial", FontSizePt: 12, Foreground: "#070154"}, {ID: "bold", Text: "bullet", FontFace: "Arial", FontSizePt: 12, Bold: true, Foreground: "#070154"}}}
	slides := []renderSlide{{ID: "rich-bullet", Width: 960, Height: 540, Elements: []element{{Name: "rich", Kind: "text", Frame: frame{X: 20, Y: 20, Width: 200, Height: 60}, Paragraphs: []compose.ParagraphSpec{p}, Text: "Editable bullet", Align: "left", Valign: "top"}}}}
	deck, err := render(slides)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRichBulletStructure(deck, slides); err != nil {
		t.Fatal(err)
	}
	xml := shapeZipPart(t, deck, "ppt/slides/slide1.xml")
	if bytes.Count(xml, []byte(`<a:pPr`)) != 1 || !bytes.Contains(xml, []byte(`marL="152400" indent="-114300"`)) || !bytes.Contains(xml, []byte(`<a:buSzPct val="100000"/><a:buChar char="&#x2022;"/>`)) {
		t.Fatalf("exact editable bullet OOXML missing: %s", xml)
	}
	mutations := map[string]func([]byte) []byte{
		"wrong margin": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`marL="152400"`), []byte(`marL="152401"`), 1)
		},
		"missing indent": func(body []byte) []byte {
			return bytes.Replace(body, []byte(` indent="-114300"`), nil, 1)
		},
		"extra pPr attribute": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:pPr algn="l"`), []byte(`<a:pPr algn="l" lvl="0"`), 1)
		},
		"wrong character": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`char="&#x2022;"`), []byte(`char="&#x2013;"`), 1)
		},
		"missing character": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buChar char="&#x2022;"/>`), nil, 1)
		},
		"duplicate character": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buChar char="&#x2022;"/>`), []byte(`<a:buChar char="&#x2022;"/><a:buChar char="&#x2022;"/>`), 1)
		},
		"wrong relative size": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buSzPct val="100000"/>`), []byte(`<a:buSzPct val="99000"/>`), 1)
		},
		"point size": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buSzPct val="100000"/>`), []byte(`<a:buSzPts val="1200"/>`), 1)
		},
		"conflicting size": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buSzPct val="100000"/>`), []byte(`<a:buSzPct val="100000"/><a:buSzTx/>`), 1)
		},
		"explicit font": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buChar`), []byte(`<a:buFont typeface="Arial"/><a:buChar`), 1)
		},
		"auto numbering": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buChar char="&#x2022;"/>`), []byte(`<a:buAutoNum type="arabicPeriod"/>`), 1)
		},
		"picture bullet": func(body []byte) []byte {
			return bytes.Replace(body, []byte(`<a:buChar char="&#x2022;"/>`), []byte(`<a:buBlip/>`), 1)
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			corrupt := rewriteImageFixture(t, deck, func(part string, body []byte) []byte {
				if part == "ppt/slides/slide1.xml" {
					return mutate(body)
				}
				return body
			})
			if err := validateRichBulletStructure(corrupt, slides); err == nil {
				t.Fatal("corrupted bullet structure accepted")
			}
		})
	}
}

func TestRichUnbulletedStructureRejectsBulletDirectives(t *testing.T) {
	p := compose.ParagraphSpec{ID: "plain", Align: "left", Runs: []compose.RunSpec{{ID: "r", Text: "Plain", FontFace: "Arial", FontSizePt: 12, Foreground: "#070154"}}}
	slides := []renderSlide{{ID: "rich-plain", Width: 960, Height: 540, Elements: []element{{Name: "rich", Kind: "text", Frame: frame{X: 20, Y: 20, Width: 200, Height: 60}, Paragraphs: []compose.ParagraphSpec{p}, Text: "Plain", Align: "left", Valign: "top"}}}}
	deck, err := render(slides)
	if err != nil {
		t.Fatal(err)
	}
	for name, replacement := range map[string][]byte{
		"missing buNone": nil,
		"font directive": []byte(`<a:buNone/><a:buFontTx/>`),
		"size directive": []byte(`<a:buNone/><a:buSzTx/>`),
	} {
		t.Run(name, func(t *testing.T) {
			corrupt := rewriteImageFixture(t, deck, func(part string, body []byte) []byte {
				if part == "ppt/slides/slide1.xml" {
					return bytes.Replace(body, []byte(`<a:buNone/>`), replacement, 1)
				}
				return body
			})
			if err := validateRichBulletStructure(corrupt, slides); err == nil {
				t.Fatal("corrupted unbulleted structure accepted")
			}
		})
	}
}

func TestRichParagraphLineSpacingXML(t *testing.T) {
	p := compose.ParagraphSpec{ID: "p", Align: "left", LineSpacingMultiple: .9, Runs: []compose.RunSpec{{ID: "r", Text: "Two wrapped lines for spacing", FontFace: "Arial", FontSizePt: 9, Italic: true, Foreground: "#070154"}}}
	slide := renderSlide{ID: "rich-spacing", Width: 960, Height: 540, Elements: []element{{Name: "rich", Kind: "text", Frame: frame{X: 20, Y: 20, Width: 60, Height: 40}, Paragraphs: []compose.ParagraphSpec{p}, Text: "Two wrapped lines for spacing", Align: "left", Valign: "top"}}}
	deck, err := render([]renderSlide{slide})
	if err != nil {
		t.Fatal(err)
	}
	xml := shapeZipPart(t, deck, "ppt/slides/slide1.xml")
	if !bytes.Contains(xml, []byte(`<a:lnSpc><a:spcPct val="90000"/></a:lnSpc>`)) {
		t.Fatal("rich paragraph 0.9 line spacing missing from OOXML")
	}
}
