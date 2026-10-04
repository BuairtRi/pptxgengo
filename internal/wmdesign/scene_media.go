package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"golang.org/x/image/vector"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const MediaSceneContract = "pptxgengo.wmds-source-media.v1"

// AgendaSchedulePhotoRefinement is a reviewed composition amendment to the
// pinned agenda/schedule example. It introduces no copy or inferred facts.
const AgendaSchedulePhotoRefinement = "wmds.agenda-schedule-photo.v1"

func applyPrimitiveLibraryRefinement(key string, doc *SlideSpec) error {
	if key != "agenda/schedule" {
		return nil
	}
	const id = "agenda-workshop-photo"
	raw, err := json.Marshal(map[string]any{
		"type": "imageframe", "x": 705, "y": 252, "w": 198, "h": 162,
		"photo": "photo-clinical-team", "focus": "65% 45%",
		"alt": "Healthcare colleagues reviewing information on a tablet",
	})
	if err != nil {
		return err
	}
	for _, node := range doc.Nodes {
		if node.ID == id {
			if node.Scene != nil && bytes.Equal(node.Scene.Node, raw) {
				return nil
			}
			return fmt.Errorf("scene.refinement_id_collision: %s", id)
		}
	}
	doc.Nodes = append(doc.Nodes, Node{ID: id, Kind: "scene", Scene: &SceneSpec{
		Node: raw, Path: "/adapters/" + AgendaSchedulePhotoRefinement,
	}})
	return nil
}

type primitiveAsset struct {
	Path, SHA256 string
	Crop         [4]int
}

func primitiveAssetBytes(key string) ([]byte, primitiveAsset, error) {
	a, ok := primitiveAssetRegistry[key]
	if !ok {
		return nil, a, fmt.Errorf("scene.unregistered_asset: %s", key)
	}
	root := os.Getenv("WMDS_BRANDING_ROOT")
	if root == "" && os.Getenv("PPTXGENGO_RELEASE_ROOT") != "" {
		root = filepath.Join(os.Getenv("PPTXGENGO_RELEASE_ROOT"), "branding")
	}
	if root == "" {
		home, e := os.UserHomeDir()
		if e != nil {
			return nil, a, e
		}
		root = filepath.Join(home, "Documents/branding")
	}
	path := a.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, a, fmt.Errorf("scene.asset_read: %s: %w", key, e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != a.SHA256 {
		return nil, a, fmt.Errorf("scene.asset_hash_drift: %s", key)
	}
	return data, a, nil
}
func primitiveDataURI(mime string, data []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}
func primitivePNG(img image.Image) ([]byte, error) {
	var b bytes.Buffer
	e := png.Encode(&b, img)
	return b.Bytes(), e
}
func primitiveFocus(focus string) (float64, float64, error) {
	if focus == "" {
		return .5, .5, nil
	}
	v := strings.Fields(focus)
	if len(v) != 2 {
		return 0, 0, fmt.Errorf("scene.invalid_focus: %s", focus)
	}
	out := [2]float64{}
	for i, s := range v {
		if !strings.HasSuffix(s, "%") {
			return 0, 0, fmt.Errorf("scene.focus_requires_percent: %s", focus)
		}
		n, e := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if e != nil || n < 0 || n > 100 {
			return 0, 0, fmt.Errorf("scene.invalid_focus: %s", focus)
		}
		out[i] = n / 100
	}
	return out[0], out[1], nil
}
func (r *renderer) primitiveMediaImage(id, asset string, b Rect, focus string, gray bool) (*pptx.ImageProps, error) {
	return r.primitiveMediaImageFit(id, asset, b, focus, gray, "cover")
}

func (r *renderer) primitiveMediaImageFit(id, asset string, b Rect, focus string, gray bool, fit string) (*pptx.ImageProps, error) {
	if fit != "" && fit != "cover" && fit != "contain" {
		return nil, fmt.Errorf("scene.invalid_image_fit: %s", fit)
	}
	data, a, e := r.primitiveAssetBytes(asset)
	if e != nil {
		return nil, e
	}
	if b.W <= 0 || b.H <= 0 {
		return nil, fmt.Errorf("scene.invalid_media_geometry: %s", id)
	}
	fx, fy, e := primitiveFocus(focus)
	if e != nil {
		return nil, e
	}
	props := &pptx.ImageProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, AltText: asset + "; canonical SHA256=" + a.SHA256}
	if strings.HasSuffix(strings.ToLower(a.Path), ".svg") {
		if gray {
			return nil, fmt.Errorf("scene.svg_grayscale_unsupported: %s", asset)
		}
		vb, e := primitiveSVGViewBox(data)
		if e != nil {
			return nil, e
		}
		if fit == "contain" {
			scale := math.Min(b.W/vb[2], b.H/vb[3])
			w, h := vb[2]*scale, vb[3]*scale
			b = Rect{b.X + (b.W-w)/2, b.Y + (b.H-h)/2, w, h}
			props.PositionProps = pos(b)
		}
		fallback, e := primitiveRasterSVG(data, int(math.Ceil(b.W*3)), int(math.Ceil(b.H*3)), vb)
		if e != nil {
			return nil, e
		}
		props.Data = primitiveDataURI("image/svg+xml", data)
		props.SVGFallbackData = primitiveDataURI("image/png", fallback)
		return props, nil
	}
	img, format, e := image.Decode(bytes.NewReader(data))
	if e != nil {
		return nil, e
	}
	bounds := img.Bounds()
	if a.Crop[0] > 0 {
		h, w, top, left := a.Crop[0], a.Crop[1], a.Crop[2], a.Crop[3]
		cb := image.Rect(bounds.Min.X+left, bounds.Min.Y+top, bounds.Min.X+left+w, bounds.Min.Y+top+h)
		if !cb.In(bounds) {
			return nil, fmt.Errorf("scene.asset_crop_out_of_bounds: %s", asset)
		}
		dst := image.NewNRGBA(image.Rect(0, 0, w, h))
		draw.Draw(dst, dst.Bounds(), img, cb.Min, draw.Src)
		img = dst
		bounds = dst.Bounds()
		data, e = primitivePNG(img)
		if e != nil {
			return nil, e
		}
		format = "png"
	}
	if gray {
		dst := image.NewNRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				v := uint8(math.Round(.2126*float64(c.R) + .7152*float64(c.G) + .0722*float64(c.B)))
				dst.SetNRGBA(x, y, color.NRGBA{R: v, G: v, B: v, A: c.A})
			}
		}
		data, e = primitivePNG(dst)
		if e != nil {
			return nil, e
		}
		format = "png"
	}
	mime := "image/" + format
	if format == "jpeg" {
		mime = "image/jpeg"
	}
	props.Data = primitiveDataURI(mime, data)
	if fit == "contain" {
		scale := math.Min(b.W/float64(bounds.Dx()), b.H/float64(bounds.Dy()))
		w, h := float64(bounds.Dx())*scale, float64(bounds.Dy())*scale
		props.PositionProps = pos(Rect{b.X + (b.W-w)/2, b.Y + (b.H-h)/2, w, h})
		return props, nil
	}
	// Explicit crop fractions preserve object-position on the native picture.
	iw, ih := float64(bounds.Dx()), float64(bounds.Dy())
	scale := math.Max(b.W/iw, b.H/ih)
	cw, ch := b.W/scale, b.H/scale
	cropX, cropY := (iw-cw)*fx, (ih-ch)*fy
	fullW, fullH := pptx.Inches(iw*scale/72), pptx.Inches(ih*scale/72)
	props.W = &fullW
	props.H = &fullH
	props.Sizing = &pptx.ImageSizing{Type: "crop", W: pptx.Inches(b.W / 72), H: pptx.Inches(b.H / 72)}
	x, y := pptx.Inches(cropX*scale/72), pptx.Inches(cropY*scale/72)
	props.Sizing.X = &x
	props.Sizing.Y = &y
	return props, nil
}

type mediaSource struct {
	Fit          string  `json:"fit,omitempty"`
	Type         string  `json:"type"`
	ID           string  `json:"id,omitempty"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	W            float64 `json:"w,omitempty"`
	H            float64 `json:"h,omitempty"`
	Size         float64 `json:"size,omitempty"`
	Photo        string  `json:"photo,omitempty"`
	Focus        string  `json:"focus,omitempty"`
	Grayscale    bool    `json:"grayscale,omitempty"`
	Alt          string  `json:"alt,omitempty"`
	Variant      string  `json:"variant,omitempty"`
	Src          string  `json:"src,omitempty"`
	Surface      string  `json:"surface,omitempty"`
	Stat         string  `json:"stat,omitempty"`
	Label        string  `json:"label,omitempty"`
	Mark         string  `json:"mark,omitempty"`
	Ink          string  `json:"ink,omitempty"`
	Rotate       float64 `json:"rotate,omitempty"`
	FlipX        bool    `json:"flipX,omitempty"`
	FlipY        bool    `json:"flipY,omitempty"`
	Kind         string  `json:"kind,omitempty"`
	Stack        bool    `json:"stack,omitempty"`
	Caption      string  `json:"caption,omitempty"`
	CaptionStyle string  `json:"captionStyle,omitempty"`
}

var mediaFields = map[string]string{"imageframe": "h photo focus grayscale alt fit rotate", "logo": "variant", "art": "src alt", "square": "size photo focus grayscale surface stat label", "mark": "h mark ink rotate flipX flipY", "thumbnail": "h kind stack caption captionStyle photo src"}

func sceneMediaRotation(degrees float64) (float64, error) {
	if math.IsNaN(degrees) || math.IsInf(degrees, 0) || math.Abs(degrees) > 360000 {
		return 0, fmt.Errorf("scene.invalid_media_rotation")
	}
	// Normalize complete turns before DrawingML's integer angle conversion.
	return math.Mod(degrees, 360), nil
}

func (r *renderer) planMediaScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var head struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &head); e != nil {
		return nil, false, e
	}
	allowed, ok := mediaFields[head.Type]
	if !ok {
		return nil, false, nil
	}
	var n mediaSource
	if e := primitiveDecode(raw, &n, allowed); e != nil {
		return nil, true, e
	}
	b := Rect{X: n.X, Y: n.Y, W: n.W, H: n.H}
	if n.Type == "square" {
		b.W = n.Size
		b.H = n.Size
	}
	if b.W <= 0 || b.H < 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
		return nil, true, fmt.Errorf("scene.invalid_media_geometry: %s", id)
	}
	p := &scenePlan{ID: id, Bounds: b}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	var err error
	picture := func(asset, focus string, gray bool) error {
		im, e := r.primitiveMediaImageFit(id+".image", asset, b, focus, gray, n.Fit)
		if e != nil {
			return e
		}
		if n.Alt != "" {
			im.AltText = n.Alt + "; " + im.AltText
		}
		rotation, err := sceneMediaRotation(n.Rotate)
		if err != nil {
			return err
		}
		im.Rotate = rotation
		pictureBounds := b
		if n.Fit == "contain" {
			pictureBounds = Rect{im.X.Val * 72, im.Y.Val * 72, im.W.Val * 72, im.H.Val * 72}
		}
		p.Bounds = diagramRotatedRect(pictureBounds, rotation)
		p.Items = append(p.Items, sceneItem{Image: im})
		return nil
	}
	switch n.Type {
	case "imageframe":
		if n.Photo == "" || b.H <= 0 {
			err = fmt.Errorf("scene.missing_photo_or_height: %s", id)
		} else {
			err = picture(n.Photo, n.Focus, n.Grayscale)
		}
	case "logo", "art":
		asset := n.Src
		if n.Type == "logo" {
			if n.Variant != "pos" && n.Variant != "rev" {
				err = fmt.Errorf("scene.invalid_logo_variant: %s", n.Variant)
				break
			}
			asset = "logo-" + n.Variant
		}
		data, _, e := r.primitiveAssetBytes(asset)
		if e != nil {
			err = e
			break
		}
		vb, e := primitiveSVGViewBox(data)
		if e != nil {
			err = e
			break
		}
		b.H = b.W * vb[3] / vb[2]
		p.Bounds = b
		err = picture(asset, "", false)
	case "square":
		if n.Photo != "" {
			if n.Stat != "" || n.Label != "" || n.Surface != "" {
				err = fmt.Errorf("scene.invalid_square_content_union: %s", id)
				break
			}
			err = picture(n.Photo, n.Focus, n.Grayscale)
			break
		}
		if n.Focus != "" || n.Grayscale {
			err = fmt.Errorf("scene.square_photo_options_require_photo: %s", id)
			break
		}
		if n.Surface == "" {
			err = fmt.Errorf("scene.missing_square_surface: %s", id)
			break
		}
		err = r.sceneRect(p, id+".surface", b, n.Surface)
		if err != nil {
			break
		}
		if n.Stat != "" {
			st, e := r.sceneStyle("stat-sm")
			if e != nil {
				err = e
				break
			}
			sl, e := r.typeEngine.Measure(n.Stat, st, b.W-24)
			if e != nil {
				err = e
				break
			}
			labelH := 0.
			var ls Style
			if n.Label != "" {
				ls, e = r.sceneStyle("small")
				if e != nil {
					err = e
					break
				}
				ll, e := r.typeEngine.Measure(n.Label, ls, b.W-24)
				if e != nil {
					err = e
					break
				}
				labelH = math.Max(ll.AllocationHeight, ll.OccupiedTop+ll.EstimatedOccupiedHeight) + 3
			}
			statH := math.Max(sl.AllocationHeight, sl.OccupiedTop+sl.EstimatedOccupiedHeight)
			y := b.Y + b.H - 12 - labelH - statH
			err = r.primitiveRichText(p, id+".stat", n.Stat, st, Rect{X: b.X + 12, Y: y, W: b.W - 24, H: statH}, n.Surface, "display", "left", "", "", ctx)
			if err == nil && n.Label != "" {
				err = r.primitiveRichText(p, id+".label", n.Label, ls, Rect{X: b.X + 12, Y: b.Y + b.H - 12 - labelH + 3, W: b.W - 24, H: labelH - 3}, n.Surface, "primary", "left", "", "", ctx)
			}
		} else if n.Label != "" {
			err = fmt.Errorf("scene.square_label_requires_stat: %s", id)
		}
	case "mark":
		rotation, e := sceneMediaRotation(n.Rotate)
		if e != nil {
			err = e
			break
		}
		im, e := r.primitiveArtworkImage(id+".mark", n.Mark, b, surface, n.Ink)
		if e != nil {
			err = e
			break
		}
		im.Rotate = rotation
		im.FlipH = &n.FlipX
		im.FlipV = &n.FlipY
		actual := Rect{im.X.Val * 72, im.Y.Val * 72, im.W.Val * 72, im.H.Val * 72}
		p.Bounds = diagramRotatedRect(actual, rotation)
		p.Items = append(p.Items, sceneItem{Image: im})
	case "thumbnail":
		err = r.primitiveThumbnail(p, n, ctx)
	}
	if err != nil {
		return nil, true, err
	}
	primitiveFinish(p, n.Type)
	for i := range p.Groups {
		if p.Groups[i].ID == id {
			p.Groups[i].Contract = MediaSceneContract
		}
	}
	return p, true, nil
}
func (r *renderer) planIconScene(id, name string, size float64, b Rect, surface, ink string) (*scenePlan, error) {
	aliases := map[string]string{"people": "people-group", "layers": "layered-documents", "chart": "bar-chart", "monitor": "computer-workstation", "server": "server-rack"}
	if _, ok := primitiveAssetRegistry["icon/"+name+"/navy"]; !ok {
		if alias, yes := aliases[name]; yes {
			name = alias
		}
	}
	c, e := r.sceneColor(surface, ink)
	if e != nil {
		return nil, e
	}
	variant := "navy"
	switch c {
	case "FFFFFF":
		variant = "white"
	case "F900D3":
		variant = "magenta"
	case "070154":
	default:
		if primitiveLuminance(c) > .5 {
			variant = "white"
		}
	}
	if size <= 0 {
		return nil, fmt.Errorf("scene.invalid_icon_size: %s", id)
	}
	b.W = size
	b.H = size
	im, e := r.primitiveMediaImage(id+".image", "icon/"+name+"/"+variant, b, "", false)
	if e != nil {
		return nil, e
	}
	p := &scenePlan{ID: id, Bounds: b, Items: []sceneItem{{Image: im}}}
	primitiveFinish(p, "icon")
	return p, nil
}
func (r *renderer) primitiveArtworkImage(id, asset string, b Rect, surface, ink string) (*pptx.ImageProps, error) {
	data, a, e := r.primitiveAssetBytes(asset)
	if e != nil {
		return nil, e
	}
	if strings.HasPrefix(asset, "highlight-") {
		if b.H == 0 {
			cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
			if e != nil {
				return nil, e
			}
			b.H = b.W * float64(cfg.Height) / float64(cfg.Width)
		}
		im, e := r.primitiveMediaImage(id, asset, b, "", false)
		if e == nil {
			// The source background stretches the complete mark into its inline
			// box. Cover sizing expands W/H to the asset aspect ratio, so removing
			// only the crop would leave a long, displaced mark behind short text.
			im.Sizing = nil
			im.PositionProps = pos(b)
		}
		return im, e
	}
	c, e := r.sceneColor(surface, ink)
	if e != nil {
		return nil, e
	}
	paths, e := primitiveSVGPaths(data, true)
	if e != nil {
		return nil, e
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("scene.empty_mark_artwork: %s", asset)
	}
	vb := primitivePathBounds(paths)
	vb[0] -= .4
	vb[1] -= .4
	vb[2] += .8
	vb[3] += .8
	if vb[2] <= 0 || vb[3] <= 0 {
		return nil, fmt.Errorf("scene.invalid_mark_bounds: %s", asset)
	}
	if b.H == 0 {
		b.H = b.W * vb[3] / vb[2]
	}
	var buf strings.Builder
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%g %g %g %g">`, vb[0], vb[1], vb[2], vb[3])
	for _, path := range paths {
		fmt.Fprintf(&buf, `<path fill="#%s" d="%s"/>`, c, path.D)
	}
	buf.WriteString("</svg>")
	svg := []byte(buf.String())
	fallback, e := primitiveRasterSVG(svg, int(math.Ceil(b.W*3)), int(math.Ceil(b.H*3)), vb)
	if e != nil {
		return nil, e
	}
	return &pptx.ImageProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, DataOrPathProps: pptx.DataOrPathProps{Data: primitiveDataURI("image/svg+xml", svg)}, SVGFallbackData: primitiveDataURI("image/png", fallback), AltText: asset + "; canonical SHA256=" + a.SHA256 + "; resolved ink=" + c}, nil
}

func (r *renderer) sceneIcon(p *scenePlan, id, name string, b Rect, surface, ink string) error {
	child, e := r.planIconScene(id, name, b.W, b, surface, ink)
	if e == nil {
		p.Append(child)
	}
	return e
}
func primitiveLuminance(c string) float64 {
	v, _ := strconv.ParseUint(c, 16, 24)
	linear := func(x float64) float64 {
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	return .2126*linear(float64(v>>16&255)/255) + .7152*linear(float64(v>>8&255)/255) + .0722*linear(float64(v&255)/255)
}
func (r *renderer) primitiveThumbnail(p *scenePlan, n mediaSource, ctx SceneContext) error {
	b := p.Bounds
	if b.H <= 0 {
		return fmt.Errorf("scene.invalid_thumbnail_height: %s", p.ID)
	}
	if n.CaptionStyle != "" && n.CaptionStyle != "italic" && n.CaptionStyle != "normal" {
		return fmt.Errorf("scene.unsupported_caption_style: %s", n.CaptionStyle)
	}
	stack := 0
	if n.Stack {
		stack = 2
	}
	page := func(id string, box Rect) error {
		if e := r.sceneRect(p, id+".fill", box, "light"); e != nil {
			return e
		}
		return r.primitiveShape(p, id+".border", box, "light", "line", true)
	}
	for i := stack; i > 0; i-- {
		if e := page(fmt.Sprintf("%s.page-%d", p.ID, i), Rect{X: b.X + float64(i)*6, Y: b.Y - float64(i)*6, W: b.W, H: b.H}); e != nil {
			return e
		}
	}
	if n.Photo != "" || n.Src != "" {
		if n.Photo != "" && n.Src != "" {
			return fmt.Errorf("scene.thumbnail_media_union: %s", p.ID)
		}
		asset := n.Photo
		if asset == "" {
			asset = n.Src
		}
		im, e := r.primitiveMediaImage(p.ID+".image", asset, b, "50% 0%", false)
		if e != nil {
			return e
		}
		p.Items = append(p.Items, sceneItem{Image: im})
		if e = r.primitiveShape(p, p.ID+".border", b, "light", "line", true); e != nil {
			return e
		}
	} else {
		if e := page(p.ID+".page", b); e != nil {
			return e
		}
		// A blank page is a composition surface for native preview content.
		// Other kinds retain the source's intentionally schematic placeholders.
		if n.Kind != "blank" {
			pad, gap := math.Max(6, b.W*.07), math.Max(3, b.H*.05)
			x, y, w := b.X+pad, b.Y+pad, b.W-2*pad
			head := math.Max(3, b.H*.07)
			if e := r.primitiveShape(p, p.ID+".head", Rect{X: x, Y: y, W: w * .55, H: head}, "light", "display", false); e != nil {
				return e
			}
			y += head + gap
			h := b.Y + b.H - pad - y
			if h <= 0 {
				return fmt.Errorf("scene.thumbnail_placeholder_overflow: %s", p.ID)
			}
			shape := func(part string, box Rect, ink string) error {
				return r.primitiveShape(p, p.ID+"."+part, box, "light", ink, false)
			}
			kind := n.Kind
			if kind == "" {
				kind = "text"
			}
			switch kind {
			case "text":
				for i, v := range []float64{.9, .75, .85, .6} {
					if e := shape(fmt.Sprintf("line-%d", i+1), Rect{X: x, Y: y + float64(i)*(math.Max(2, b.H*.04)+gap), W: w * v, H: math.Max(2, b.H*.04)}, "line"); e != nil {
						return e
					}
				}
			case "chart":
				cw := (w - 12) / 5
				for i, v := range []float64{.4, .65, .5, .85, .7} {
					ink := "line"
					if i == 3 {
						ink = "emphasis"
					}
					if e := shape(fmt.Sprintf("bar-%d", i+1), Rect{X: x + float64(i)*(cw+3), Y: y + h*(1-v), W: cw, H: h * v}, ink); e != nil {
						return e
					}
				}
			case "table":
				rh := (h - 8) / 5
				for i := 0; i < 5; i++ {
					ink := "bg"
					surf := "subtle"
					if i == 0 {
						surf = "light"
						ink = "secondary"
					}
					if e := r.primitiveShape(p, fmt.Sprintf("%s.row-%d", p.ID, i+1), Rect{X: x, Y: y + float64(i)*(rh+2), W: w, H: rh}, surf, ink, false); e != nil {
						return e
					}
				}
			case "diagram":
				cw := (w - 8) / 3
				bh := math.Max(4, b.H*.12)
				for i, count := range []int{2, 3, 2} {
					ink := "line"
					if i == 1 {
						ink = "display"
					}
					top := y + (h-float64(count)*bh-float64(count-1)*3)/2
					for j := 0; j < count; j++ {
						if e := shape(fmt.Sprintf("box-%d-%d", i+1, j+1), Rect{X: x + float64(i)*(cw+4), Y: top + float64(j)*(bh+3), W: cw, H: bh}, ink); e != nil {
							return e
						}
					}
				}
			default:
				return fmt.Errorf("scene.unsupported_thumbnail_kind: %s", kind)
			}
		}
	}
	if n.Caption != "" {
		st, e := r.sceneStyle("small")
		if e != nil {
			return e
		}
		st.Italic = n.CaptionStyle == "italic"
		if e = r.primitiveRichText(p, p.ID+".caption", n.Caption, st, Rect{X: b.X, Y: b.Y + b.H + 6, W: b.W + float64(stack)*6}, ctx.Surface, "secondary", "left", "", "", ctx); e != nil {
			return e
		}
	}
	return nil
}

// The fallback rasterizer accepts the closed filled SVG geometry used by the
// pinned official assets. Unsupported visible elements fail explicitly.
type primitiveSVGPath struct {
	D      string
	Fill   string
	Matrix [6]float64
}

var primitiveNumberRE = regexp.MustCompile(`[-+]?(?:\d*\.\d+|\d+\.?\d*)(?:[eE][-+]?\d+)?`)

func primitiveSVGViewBox(data []byte) ([4]float64, error) {
	var out [4]float64
	d := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, e := d.Token()
		if e != nil {
			return out, e
		}
		if start, ok := tok.(xml.StartElement); ok && start.Name.Local == "svg" {
			for _, a := range start.Attr {
				if a.Name.Local == "viewBox" {
					nums := primitiveNumberRE.FindAllString(a.Value, -1)
					if len(nums) != 4 {
						return out, fmt.Errorf("scene.invalid_svg_viewbox")
					}
					for i, n := range nums {
						out[i], e = strconv.ParseFloat(n, 64)
						if e != nil {
							return out, e
						}
					}
					if out[2] <= 0 || out[3] <= 0 {
						return out, fmt.Errorf("scene.invalid_svg_viewbox")
					}
					return out, nil
				}
			}
			return out, fmt.Errorf("scene.svg_missing_viewbox")
		}
	}
}
func primitiveMatrixMul(a, b [6]float64) [6]float64 {
	return [6]float64{a[0]*b[0] + a[2]*b[1], a[1]*b[0] + a[3]*b[1], a[0]*b[2] + a[2]*b[3], a[1]*b[2] + a[3]*b[3], a[0]*b[4] + a[2]*b[5] + a[4], a[1]*b[4] + a[3]*b[5] + a[5]}
}
func primitiveSVGPaths(data []byte, artOnly bool) ([]primitiveSVGPath, error) {
	css := map[string]string{}
	re := regexp.MustCompile(`(?s)\.([\w-]+)\s*\{([^}]+)\}`)
	for _, m := range re.FindAllStringSubmatch(string(data), -1) {
		css[m[1]] = m[2]
	}
	type state struct {
		matrix      [6]float64
		fill        string
		hidden, art bool
	}
	states := []state{{matrix: [6]float64{1, 0, 0, 1, 0, 0}, fill: "#000000"}}
	var paths []primitiveSVGPath
	d := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		switch t := tok.(type) {
		case xml.StartElement:
			s := states[len(states)-1]
			attrs := map[string]string{}
			for _, a := range t.Attr {
				attrs[a.Name.Local] = a.Value
			}
			sty := attrs["style"] + ";" + css[attrs["class"]]
			if attrs["id"] == "artwork" {
				s.art = true
			}
			if attrs["id"] == "stroke" || t.Name.Local == "defs" || strings.Contains(strings.ReplaceAll(sty, " ", ""), "display:none") {
				s.hidden = true
			}
			if fill := attrs["fill"]; fill != "" {
				s.fill = fill
			}
			for _, part := range strings.Split(sty, ";") {
				kv := strings.SplitN(part, ":", 2)
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "fill" {
					s.fill = strings.TrimSpace(kv[1])
				}
			}
			if transform := attrs["transform"]; transform != "" {
				if !strings.HasPrefix(transform, "matrix(") {
					return nil, fmt.Errorf("scene.unsupported_svg_transform: %s", transform)
				}
				v := primitiveNumberRE.FindAllString(transform, -1)
				if len(v) != 6 {
					return nil, fmt.Errorf("scene.invalid_svg_transform")
				}
				var m [6]float64
				for i, n := range v {
					m[i], e = strconv.ParseFloat(n, 64)
					if e != nil {
						return nil, e
					}
				}
				s.matrix = primitiveMatrixMul(s.matrix, m)
			}
			states = append(states, s)
			if s.hidden || (artOnly && !s.art) {
				continue
			}
			path := attrs["d"]
			val := func(k string) float64 { v, _ := strconv.ParseFloat(attrs[k], 64); return v }
			switch t.Name.Local {
			case "path":
			case "rect":
				x, y, w, h := val("x"), val("y"), val("width"), val("height")
				rx, ry := val("rx"), val("ry")
				if rx == 0 {
					rx = ry
				}
				if ry == 0 {
					ry = rx
				}
				rx = math.Min(rx, w/2)
				ry = math.Min(ry, h/2)
				if rx > 0 && ry > 0 {
					path = fmt.Sprintf("M%g %gH%gA%g %g 0 0 1 %g %gV%gA%g %g 0 0 1 %g %gH%gA%g %g 0 0 1 %g %gV%gA%g %g 0 0 1 %g %gZ", x+rx, y, x+w-rx, rx, ry, x+w, y+ry, y+h-ry, rx, ry, x+w-rx, y+h, x+rx, rx, ry, x, y+h-ry, y+ry, rx, ry, x+rx, y)
				} else {
					path = fmt.Sprintf("M%g %gH%gV%gH%gZ", x, y, x+w, y+h, x)
				}
			case "circle":
				cx, cy, rad := val("cx"), val("cy"), val("r")
				path = fmt.Sprintf("M%g %gA%g %g 0 1 0 %g %gA%g %g 0 1 0 %g %gZ", cx+rad, cy, rad, rad, cx-rad, cy, rad, rad, cx+rad, cy)
			case "svg", "g", "style", "defs", "title", "desc":
				continue
			default:
				return nil, fmt.Errorf("scene.unsupported_svg_element: %s", t.Name.Local)
			}
			if s.fill == "none" {
				if attrs["stroke"] != "" && attrs["stroke"] != "none" {
					return nil, fmt.Errorf("scene.unsupported_svg_visible_stroke")
				}
				continue
			}
			if path != "" {
				paths = append(paths, primitiveSVGPath{D: path, Fill: s.fill, Matrix: s.matrix})
			}
		case xml.EndElement:
			if len(states) > 1 {
				states = states[:len(states)-1]
			}
		}
	}
	return paths, nil
}

type primitivePathSink struct {
	Move  func(float64, float64)
	Line  func(float64, float64)
	Cube  func(float64, float64, float64, float64, float64, float64)
	Quad  func(float64, float64, float64, float64)
	Close func()
}

var primitivePathTokenRE = regexp.MustCompile(`[A-Za-z]|[-+]?(?:\d*\.\d+|\d+\.?\d*)(?:[eE][-+]?\d+)?`)

func primitiveWalkPath(path primitiveSVGPath, sink primitivePathSink) error {
	tokens := primitivePathTokenRE.FindAllString(path.D, -1)
	i := 0
	var cmd byte
	var x, y, sx, sy, cx, cy, qx, qy float64
	prev := byte(0)
	point := func(px, py float64) (float64, float64) {
		m := path.Matrix
		return m[0]*px + m[2]*py + m[4], m[1]*px + m[3]*py + m[5]
	}
	move := func(px, py float64) { a, b := point(px, py); sink.Move(a, b) }
	line := func(px, py float64) { a, b := point(px, py); sink.Line(a, b) }
	cube := func(a, b, c, d, e, f float64) {
		a, b = point(a, b)
		c, d = point(c, d)
		e, f = point(e, f)
		sink.Cube(a, b, c, d, e, f)
	}
	quad := func(a, b, c, d float64) { a, b = point(a, b); c, d = point(c, d); sink.Quad(a, b, c, d) }
	for i < len(tokens) {
		if len(tokens[i]) == 1 && ((tokens[i][0] >= 'A' && tokens[i][0] <= 'Z') || (tokens[i][0] >= 'a' && tokens[i][0] <= 'z')) {
			cmd = tokens[i][0]
			i++
		}
		if cmd == 0 {
			return fmt.Errorf("scene.invalid_svg_path")
		}
		upper := cmd
		if upper >= 'a' && upper <= 'z' {
			upper -= 32
		}
		if upper == 'Z' {
			sink.Close()
			x, y = sx, sy
			prev = upper
			cmd = 0
			continue
		}
		arity := map[byte]int{'M': 2, 'L': 2, 'H': 1, 'V': 1, 'C': 6, 'S': 4, 'Q': 4, 'T': 2, 'A': 7}[upper]
		if arity == 0 {
			return fmt.Errorf("scene.unsupported_svg_path_command: %c", cmd)
		}
		if i+arity > len(tokens) {
			return fmt.Errorf("scene.incomplete_svg_path")
		}
		v := make([]float64, arity)
		for j := range v {
			n, e := strconv.ParseFloat(tokens[i+j], 64)
			if e != nil {
				return fmt.Errorf("scene.invalid_svg_path_operand: %s", tokens[i+j])
			}
			v[j] = n
		}
		i += arity
		relative := cmd >= 'a' && cmd <= 'z'
		pair := func(j int) (float64, float64) {
			a, b := v[j], v[j+1]
			if relative {
				a += x
				b += y
			}
			return a, b
		}
		switch upper {
		case 'M':
			x, y = pair(0)
			sx, sy = x, y
			move(x, y)
			if relative {
				cmd = 'l'
			} else {
				cmd = 'L'
			}
		case 'L':
			x, y = pair(0)
			line(x, y)
		case 'H':
			nx := v[0]
			if relative {
				nx += x
			}
			x = nx
			line(x, y)
		case 'V':
			ny := v[0]
			if relative {
				ny += y
			}
			y = ny
			line(x, y)
		case 'C':
			a, b := pair(0)
			c, d := pair(2)
			e, f := pair(4)
			cube(a, b, c, d, e, f)
			cx, cy = c, d
			x, y = e, f
		case 'S':
			a, b := x, y
			if prev == 'C' || prev == 'S' {
				a, b = 2*x-cx, 2*y-cy
			}
			c, d := pair(0)
			e, f := pair(2)
			cube(a, b, c, d, e, f)
			cx, cy = c, d
			x, y = e, f
		case 'Q':
			a, b := pair(0)
			c, d := pair(2)
			quad(a, b, c, d)
			qx, qy = a, b
			x, y = c, d
		case 'T':
			a, b := x, y
			if prev == 'Q' || prev == 'T' {
				a, b = 2*x-qx, 2*y-qy
			}
			c, d := pair(0)
			quad(a, b, c, d)
			qx, qy = a, b
			x, y = c, d
		case 'A':
			nx, ny := pair(5)
			rx, ry := math.Abs(v[0]), math.Abs(v[1])
			if rx == 0 || ry == 0 {
				line(nx, ny)
				x, y = nx, ny
				break
			}
			if (v[3] != 0 && v[3] != 1) || (v[4] != 0 && v[4] != 1) {
				return fmt.Errorf("scene.invalid_svg_arc_flags")
			}
			if nx == x && ny == y {
				break
			}
			phi := v[2] * math.Pi / 180
			co, si := math.Cos(phi), math.Sin(phi)
			dx, dy := (x-nx)/2, (y-ny)/2
			xp, yp := co*dx+si*dy, -si*dx+co*dy
			lam := xp*xp/(rx*rx) + yp*yp/(ry*ry)
			if lam > 1 {
				rx *= math.Sqrt(lam)
				ry *= math.Sqrt(lam)
			}
			den := rx*rx*yp*yp + ry*ry*xp*xp
			factor := math.Sqrt(math.Max(0, (rx*rx*ry*ry-den)/den))
			if v[3] == v[4] {
				factor = -factor
			}
			cpX, cpY := factor*rx*yp/ry, -factor*ry*xp/rx
			centerX, centerY := co*cpX-si*cpY+(x+nx)/2, si*cpX+co*cpY+(y+ny)/2
			start := math.Atan2((yp-cpY)/ry, (xp-cpX)/rx)
			end := math.Atan2((-yp-cpY)/ry, (-xp-cpX)/rx)
			sweep := end - start
			if v[4] == 0 && sweep > 0 {
				sweep -= 2 * math.Pi
			} else if v[4] == 1 && sweep < 0 {
				sweep += 2 * math.Pi
			}
			steps := int(math.Ceil(math.Abs(sweep) * 180 / math.Pi))
			for step := 1; step <= steps; step++ {
				a := start + sweep*float64(step)/float64(steps)
				ax, ay := rx*math.Cos(a), ry*math.Sin(a)
				line(centerX+co*ax-si*ay, centerY+si*ax+co*ay)
			}
			x, y = nx, ny
		}
		prev = upper
	}
	return nil
}
func primitivePathBounds(paths []primitiveSVGPath) [4]float64 {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	var x, y float64
	save := func(a, b float64) {
		minX = math.Min(minX, a)
		minY = math.Min(minY, b)
		maxX = math.Max(maxX, a)
		maxY = math.Max(maxY, b)
		x, y = a, b
	}
	sink := primitivePathSink{Move: save, Line: save, Close: func() {}}
	sink.Cube = func(a, b, c, d, e, f float64) {
		ox, oy := x, y
		for i := 1; i <= 256; i++ {
			t := float64(i) / 256
			u := 1 - t
			save(u*u*u*ox+3*u*u*t*a+3*u*t*t*c+t*t*t*e, u*u*u*oy+3*u*u*t*b+3*u*t*t*d+t*t*t*f)
		}
	}
	sink.Quad = func(a, b, c, d float64) {
		ox, oy := x, y
		for i := 1; i <= 256; i++ {
			t := float64(i) / 256
			u := 1 - t
			save(u*u*ox+2*u*t*a+t*t*c, u*u*oy+2*u*t*b+t*t*d)
		}
	}
	for _, p := range paths {
		if e := primitiveWalkPath(p, sink); e != nil {
			return [4]float64{}
		}
	}
	return [4]float64{minX, minY, maxX - minX, maxY - minY}
}
func primitiveRasterSVG(data []byte, w, h int, vb [4]float64) ([]byte, error) {
	if w <= 0 || h <= 0 || w > 8192 || h > 8192 {
		return nil, fmt.Errorf("scene.svg_fallback_dimensions")
	}
	paths, e := primitiveSVGPaths(data, false)
	if e != nil {
		return nil, e
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("scene.svg_no_visible_paths")
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	transform := func(x, y float64) (float32, float32) {
		return float32((x - vb[0]) * float64(w) / vb[2]), float32((y - vb[1]) * float64(h) / vb[3])
	}
	for _, p := range paths {
		fill := strings.TrimPrefix(p.Fill, "#")
		if len(fill) == 3 {
			fill = string([]byte{fill[0], fill[0], fill[1], fill[1], fill[2], fill[2]})
		}
		if len(fill) != 6 {
			return nil, fmt.Errorf("scene.unsupported_svg_fill: %s", p.Fill)
		}
		col, e := strconv.ParseUint(fill, 16, 24)
		if e != nil {
			return nil, e
		}
		rast := vector.NewRasterizer(w, h)
		sink := primitivePathSink{Move: func(x, y float64) { a, b := transform(x, y); rast.MoveTo(a, b) }, Line: func(x, y float64) { a, b := transform(x, y); rast.LineTo(a, b) }, Cube: func(a, b, c, d, e, f float64) {
			a1, b1 := transform(a, b)
			c1, d1 := transform(c, d)
			e1, f1 := transform(e, f)
			rast.CubeTo(a1, b1, c1, d1, e1, f1)
		}, Quad: func(a, b, c, d float64) {
			a1, b1 := transform(a, b)
			c1, d1 := transform(c, d)
			rast.QuadTo(a1, b1, c1, d1)
		}, Close: rast.ClosePath}
		if e = primitiveWalkPath(p, sink); e != nil {
			return nil, e
		}
		rast.Draw(dst, dst.Bounds(), image.NewUniform(color.RGBA{R: uint8(col >> 16), G: uint8(col >> 8), B: uint8(col), A: 255}), image.Point{})
	}
	return primitivePNG(dst)
}
