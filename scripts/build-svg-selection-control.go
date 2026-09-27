//go:build ignore

// Build a QA-only control whose SVG primary is green and PNG fallback magenta.
// Usage: go run scripts/build-svg-selection-control.go new-output.pptx
package main

import (
	"bytes"
	"encoding/base64"
	"github.com/buairtri/pptxgengo/pptx"
	"image"
	"image/color"
	"image/png"
	"os"
)

func ptr[T any](v T) *T { return &v }
func main() {
	if len(os.Args) != 2 {
		panic("Usage: build-svg-selection-control new-output.pptx")
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><path fill="#00AA44" d="M0 0H100V100H0Z"/></svg>`)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 255, 255})
		}
	}
	var pngBytes bytes.Buffer
	if err := png.Encode(&pngBytes, img); err != nil {
		panic(err)
	}
	p := pptx.New()
	p.DefineLayout("PROOF", 13.333333, 7.5)
	if err := p.SetLayout("PROOF"); err != nil {
		panic(err)
	}
	s := p.AddSlide()
	s.Background(&pptx.BackgroundProps{Color: "FFFFFF"})
	err := s.AddImage(&pptx.ImageProps{PositionProps: pptx.PositionProps{X: ptr(pptx.Inches(1)), Y: ptr(pptx.Inches(1)), W: ptr(pptx.Inches(2)), H: ptr(pptx.Inches(2))}, DataOrPathProps: pptx.DataOrPathProps{Data: "image/svg+xml;base64," + base64.StdEncoding.EncodeToString(svg)}, SVGFallbackData: "image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes.Bytes()), AltText: "QA only: green SVG primary, magenta PNG fallback"})
	if err != nil {
		panic(err)
	}
	b, err := p.Write()
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(os.Args[1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(b); err != nil {
		f.Close()
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
}
