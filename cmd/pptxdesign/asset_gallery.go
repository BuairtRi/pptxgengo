package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"golang.org/x/image/draw"
)

func runAssetGallery(args []string) error {
	f := flag.NewFlagSet("asset-gallery", flag.ContinueOnError)
	out := f.String("out", "", "new gallery directory")
	kind := f.String("kind", "all", "all, icon, photo, graphic, logo")
	query := f.String("query", "", "asset search")
	limit := f.Int("limit", 0, "maximum asset concepts; zero means all")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *out == "" {
		return fmt.Errorf("asset-gallery requires --out NEW-DIR")
	}
	if _, err := os.Lstat(*out); !os.IsNotExist(err) {
		return fmt.Errorf("output must be new: %s", *out)
	}
	items, err := wmdesign.AssetSelections(*query, *kind, *limit)
	if err != nil {
		return err
	}
	parent := filepath.Dir(*out)
	if err = os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, ".asset-gallery-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err = os.Mkdir(filepath.Join(stage, "assets"), 0755); err != nil {
		return err
	}
	for i := range items {
		for j := range items[i].Variants {
			v := &items[i].Variants[j]
			data, ref, _, e := wmdesign.PrimitiveAssetData(v.ID)
			if e != nil {
				return e
			}
			if ref.SHA256 != v.SHA256 {
				return fmt.Errorf("asset changed: %s", v.ID)
			}
			preview, extension, derived, e := assetGalleryPreview(data, ref.Path)
			if e != nil {
				return e
			}
			name := fmt.Sprintf("assets/%03d-%02d%s", i+1, j+1, extension)
			if e = os.WriteFile(filepath.Join(stage, filepath.FromSlash(name)), preview, 0644); e != nil {
				return e
			}
			v.ThumbnailPath = name
			hash := sha256.Sum256(preview)
			v.ThumbnailSHA256 = hex.EncodeToString(hash[:])
			if derived {
				v.ThumbnailMIME = "image/png"
				v.ThumbnailState = "derived_go_preview_from_verified_original"
			}
		}
	}
	raw, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(stage, "assets.json"), append(raw, '\n'), 0644); err != nil {
		return err
	}
	t := template.Must(template.New("gallery").Parse(assetGalleryHTML))
	file, err := os.Create(filepath.Join(stage, "index.html"))
	if err != nil {
		return err
	}
	err = t.Execute(file, items)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if _, err = os.Lstat(*out); !os.IsNotExist(err) {
		return fmt.Errorf("output appeared during generation")
	}
	if err = os.Rename(stage, *out); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"gallery": filepath.Join(*out, "index.html"), "concepts": len(items), "status": "original_bytes_verified", "policy": "Raster thumbnails are derived in Go for browsing; SVGs retain original bytes. Registered originals and SHA256 remain authoritative for slide use."})
}

func assetGalleryPreview(data []byte, path string) ([]byte, string, bool, error) {
	extension := strings.ToLower(filepath.Ext(path))
	if extension == ".svg" {
		return data, extension, false, nil
	}
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", false, err
	}
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, "", false, fmt.Errorf("empty asset image")
	}
	if width > 420 || height > 280 {
		if width*280 > height*420 {
			height = max(1, height*420/width)
			width = 420
		} else {
			width = max(1, width*280/height)
			height = 280
		}
	}
	target := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(target, target.Bounds(), source, bounds, draw.Src, nil)
	var buffer bytes.Buffer
	if err = png.Encode(&buffer, target); err != nil {
		return nil, "", false, err
	}
	return buffer.Bytes(), ".png", true, nil
}

const assetGalleryHTML = `<!doctype html><html><head><meta charset="utf-8"><title>West Monroe assets</title><style>body{font:16px system-ui;background:#f2f4f8;color:#061c44;margin:2rem}main{display:grid;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));gap:1rem}article{background:white;padding:1rem;border-radius:8px}.variants{display:flex;flex-wrap:wrap;gap:1rem}figure{margin:0;max-width:100%}img{width:150px;height:110px;object-fit:contain}.white{background:#061c44}figcaption{font-size:12px;max-width:180px;overflow-wrap:anywhere}input{padding:.7rem;width:24rem;margin-bottom:1rem}code{font-size:12px}</style></head><body><h1>West Monroe assets</h1><p>Browser previews from verified originals; color variants retain their library keys. Use registry originals for slides.</p><input id="filter" placeholder="Filter assets"><main>{{range .}}<article><h2>{{.Name}}</h2><code>{{.ID}}</code><p>{{.Description}}</p><p>{{range .Tags}}{{.}} · {{end}}</p><div class="variants">{{range .Variants}}<figure><img loading="lazy" src="{{.ThumbnailPath}}" class="{{.Color}}" alt="{{.ID}}"><figcaption>{{.ID}}<br>{{range .RecommendedSurfaces}}{{.}} {{end}}</figcaption></figure>{{end}}</div></article>{{end}}</main><script>document.getElementById('filter').oninput=e=>{const q=e.target.value.toLowerCase();document.querySelectorAll('article').forEach(a=>a.hidden=!a.textContent.toLowerCase().includes(q));};</script></body></html>`
