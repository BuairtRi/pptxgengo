package nativeexport

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func createContactSheet(work, out string, mappings []PageMapping) (*Artifact, error) {
	columns := min(3, len(mappings))
	const cellWidth, cellHeight, padding = 420, 280, 12
	rows := (len(mappings) + columns - 1) / columns
	canvas := image.NewRGBA(image.Rect(0, 0, columns*cellWidth, rows*cellHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.RGBA{245, 245, 245, 255}}, image.Point{}, draw.Src)
	for index, mapping := range mappings {
		data, err := os.ReadFile(filepath.Join(work, "native-pages", fmt.Sprintf("slide-%03d.png", index+1)))
		if err != nil {
			return nil, err
		}
		source, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		x, y := (index%columns)*cellWidth+padding, (index/columns)*cellHeight+padding
		bounds := source.Bounds()
		width, height := cellWidth-2*padding, cellHeight-54
		if bounds.Dx()*height > bounds.Dy()*width {
			height = max(1, bounds.Dy()*width/bounds.Dx())
		} else {
			width = max(1, bounds.Dx()*height/bounds.Dy())
		}
		draw.CatmullRom.Scale(canvas, image.Rect(x, y, x+width, y+height), source, bounds, draw.Src, nil)
		label := fmt.Sprintf("Slide %d | PDF page %d", mapping.SourceSlide, mapping.Page)
		if mapping.SourceHidden {
			label += " | originally hidden"
		}
		drawer := font.Drawer{Dst: canvas, Src: image.Black, Face: basicfont.Face7x13, Dot: fixed.P(x, y+cellHeight-30)}
		drawer.DrawString(label)
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, canvas); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(out, "contact-sheet.png"), buffer.Bytes(), 0644); err != nil {
		return nil, err
	}
	return &Artifact{Path: "contact-sheet.png", SHA256: hash(buffer.Bytes()), Width: canvas.Bounds().Dx(), Height: canvas.Bounds().Dy()}, nil
}
