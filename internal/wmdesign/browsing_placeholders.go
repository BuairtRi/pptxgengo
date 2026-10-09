package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
)

// These dimensions describe only the schematic envelope. The underscore
// envelope is pinned to registered SVG artwork bounds (not its square viewBox);
// highlight dimensions follow the V11 registered derived preview's aspect.
const schematicUnderscoreWidth = 72.03073503971098
const schematicUnderscoreHeight = 5.770592057704931
const schematicHighlightWidth = 420
const schematicHighlightHeight = 57

// TemplatePlaceholderAssets supplies conspicuously schematic media for a
// browsing library. It never reads originals or represents these bytes as the
// registered artwork. The normal renderer still requires verified originals.
func TemplatePlaceholderAssets() (map[string]AssetData, error) {
	out := map[string]AssetData{}
	photo, err := browsingPlaceholderPNG(640, 360)
	if err != nil {
		return nil, err
	}
	highlight, err := browsingPlaceholderPNG(schematicHighlightWidth, schematicHighlightHeight)
	if err != nil {
		return nil, err
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 60"><g id="artwork"><path fill="#B8BFC6" d="M0 0H100V60H0Z"/><path fill="#77828C" d="M0 0L100 60L100 56L4 0Z M100 0L0 60L0 56L96 0Z"/></g></svg>`)
	underscore := []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.17g %.17g"><g id="artwork"><path fill="#B8BFC6" d="M0 0H%.17gV%.17gH0Z"/></g></svg>`, schematicUnderscoreWidth, schematicUnderscoreHeight, schematicUnderscoreWidth, schematicUnderscoreHeight))
	for key, asset := range primitiveAssetRegistry {
		data, mime := photo, "image/png"
		if strings.HasSuffix(strings.ToLower(asset.Path), ".svg") {
			data, mime = svg, "image/svg+xml"
		}
		if strings.HasPrefix(key, "highlight-") {
			data, mime = highlight, "image/png"
		}
		if key == "underscore" {
			data, mime = underscore, "image/svg+xml"
		}
		out[key] = AssetData{Data: data, SHA256: fmt.Sprintf("%x", sha256.Sum256(data)), MIME: mime}
	}
	return out, nil
}

func browsingPlaceholderPNG(w, h int) ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.NRGBA{R: 228, G: 232, B: 236, A: 255}
			if x < 3 || y < 3 || x >= w-3 || y >= h-3 || absPlaceholder(x*h-y*w) < w*3 || absPlaceholder((w-x)*h-y*w) < w*3 {
				c = color.NRGBA{R: 139, G: 150, B: 161, A: 255}
			}
			img.SetNRGBA(x, y, c)
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func absPlaceholder(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
