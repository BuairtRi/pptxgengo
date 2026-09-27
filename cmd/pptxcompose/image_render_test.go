package main

import (
	"archive/zip"
	"bytes"
	"image"
	stdcolor "image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
)

func TestImageCropModes(t *testing.T) {
	e := element{Name: "sample", Frame: frame{Width: 100, Height: 100}}
	if _, err := imageCrop(e, 200, 100); err == nil {
		t.Fatal("default allowed distortion")
	}
	e.ImageFit = "contain"
	c, err := imageCrop(e, 200, 100)
	if err != nil || c.Top != -.5 || c.Bottom != -.5 || c.Left != 0 {
		t.Fatalf("contain: %+v %v", c, err)
	}
	e.ImageFit = "cover"
	e.FocalX = pointer(1.0)
	c, err = imageCrop(e, 200, 100)
	if err != nil || c.Left != .5 || c.Right != 0 {
		t.Fatalf("right cover: %+v %v", c, err)
	}
	e.FocalX = nil
	e.ImageFit = "source_crop"
	e.ImageCrop = &compose.ImageCropSpec{Left: .5}
	if _, err = imageCrop(e, 200, 100); err != nil {
		t.Fatal(err)
	}
	e.ImageCrop = &compose.ImageCropSpec{Left: .2}
	if _, err = imageCrop(e, 200, 100); err == nil {
		t.Fatal("source crop allowed distortion")
	}
}

func TestImageCropNativePackage(t *testing.T) {
	var buf bytes.Buffer
	im := image.NewRGBA(image.Rect(0, 0, 200, 100))
	im.Set(0, 0, stdcolor.RGBA{R: 255, A: 255})
	if err := png.Encode(&buf, im); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "fixture.png")
	if err := os.WriteFile(p, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	e := element{Name: "picture", Kind: "image", Frame: frame{X: 10, Y: 20, Width: 100, Height: 100}, AssetPath: p, AssetSHA256: hash(buf.Bytes()), AltText: "test", ImageFit: "cover", FocalX: pointer(1.0)}
	b, err := render([]renderSlide{{ID: "test", Width: 960, Height: 540, Elements: []element{e}}})
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if f.Name != "ppt/slides/slide1.xml" {
			continue
		}
		r, _ := f.Open()
		raw, _ := io.ReadAll(r)
		r.Close()
		s := string(raw)
		if !strings.Contains(s, `<a:srcRect l="50000" r="0" t="0" b="0"/>`) {
			t.Fatalf("wrong crop: %s", s)
		}
		if !strings.Contains(s, `<a:ext cx="1270000" cy="1270000"/>`) {
			t.Fatalf("wrong picture frame: %s", s)
		}
		return
	}
	t.Fatal("missing slide")
}

func TestImageContainPreservesRatio(t *testing.T) {
	for _, wh := range [][2]int{{200, 100}, {100, 300}, {512, 512}} {
		e := element{Frame: frame{Width: 230, Height: 145}, ImageFit: "contain"}
		c, err := imageCrop(e, wh[0], wh[1])
		if err != nil {
			t.Fatal(err)
		}
		displayRatio := (e.Frame.Width / (1 - c.Left - c.Right)) / (e.Frame.Height / (1 - c.Top - c.Bottom))
		if math.Abs(displayRatio-float64(wh[0])/float64(wh[1])) > 1e-9 {
			t.Fatal("contain distorted source")
		}
	}
}

func TestRenderSVGPreservesSourceAndPngFallback(t *testing.T) {
	dir := t.TempDir()
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path d="M0 0h100v100z"/></svg>`)
	svgPath := filepath.Join(dir, "source.svg")
	if err := os.WriteFile(svgPath, svg, 0600); err != nil {
		t.Fatal(err)
	}
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 100, 100))); err != nil {
		t.Fatal(err)
	}
	pngPath := filepath.Join(dir, "fallback.png")
	if err := os.WriteFile(pngPath, pngData.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	e := element{Name: "svg", Kind: "image", Frame: frame{Width: 100, Height: 100}, AssetPath: svgPath, AssetSHA256: hash(svg), FallbackAssetPath: pngPath, FallbackAssetSHA256: hash(pngData.Bytes()), AltText: "svg", OutlineColor: "112233", OutlineWidthPt: 1}
	deck, err := render([]renderSlide{{ID: "svg", Width: 960, Height: 540, Elements: []element{e}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := validateImageStructure(deck, []renderSlide{{ID: "svg", Width: 960, Height: 540, Elements: []element{e}}}); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(string, []byte) []byte{
		func(name string, body []byte) []byte {
			if name == "ppt/slides/slide1.xml" {
				return bytes.Replace(body, []byte(`<a:blip r:embed=`), []byte(`<a:blip r:link="external" r:embed=`), 1)
			}
			return body
		},
		func(name string, body []byte) []byte {
			if name == "ppt/slides/slide1.xml" {
				return bytes.Replace(body, []byte(`<asvg:svgBlip `), []byte(`<asvg:svgBlip r:link="external" `), 1)
			}
			return body
		},

		func(name string, body []byte) []byte {
			if name == "ppt/slides/slide1.xml" {
				return bytes.Replace(body, []byte(`</a:ln>`), []byte(`</a:ln><a:ln w="12700"><a:solidFill><a:srgbClr val="112233"/></a:solidFill></a:ln>`), 1)
			}
			return body
		},
		func(name string, body []byte) []byte {
			if name == "ppt/slides/slide1.xml" {
				return bytes.Replace(body, []byte(`<a:srgbClr val="112233"/>`), []byte(`<a:srgbClr val="112233"><a:alpha val="50000"/></a:srgbClr>`), 1)
			}
			return body
		},
	} {
		if err := validateImageStructure(rewriteImageFixture(t, deck, mutate), []renderSlide{{ID: "svg", Width: 960, Height: 540, Elements: []element{e}}}); err == nil {
			t.Fatal("accepted invalid picture outline")
		}
	}
}

func TestSVGIntrinsicSizeFailsClosedOnExternalReference(t *testing.T) {
	if _, _, err := svgIntrinsicSize([]byte(`<svg viewBox="0 0 10 10"><image href="https://example.test/x.png"/></svg>`)); err == nil {
		t.Fatal("external SVG reference accepted")
	}
}

func TestSVGIntrinsicSizeRejectsUnsupportedStaticContract(t *testing.T) {
	for _, body := range []string{
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><svg viewBox="0 0 10 10"/></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path xmlns="urn:foreign"/></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="NaN 0 10 10"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg" width="" height="10" viewBox="0 0 10 10"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg" width="10" viewBox="0 0 10 10"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg" xmlns:x="urn:x" x:width="10" height="10" viewBox="0 0 10 10"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><style>.a{fill:u\72l(https://example.test/a)}</style></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"/>trailing`,

		`<svg xmlns="http://www.w3.org/2000/svg" width="10%" height="10" viewBox="0 0 10 10"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="NaN 0 10 10"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><animateMotion/></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" onclick="x()"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect/></svg>trailing`,
		`<?xml-stylesheet href="https://example.test/x.css"?><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"/>`,
	} {
		if _, _, err := svgIntrinsicSize([]byte(body)); err == nil {
			t.Fatalf("accepted unsupported SVG: %s", body)
		}
	}
	if w, h, err := svgIntrinsicSize([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10.1 1.1"><path d="M0 0"/></svg>`)); err != nil || w != 10.1 || h != 1.1 {
		t.Fatalf("fractional viewBox = %v,%v %v", w, h, err)
	}
}
