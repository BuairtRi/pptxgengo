package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

const IntakeArchitectureContract = "pptxgengo.wmds-source-intake-architecture.v1"

type intakeArchitectureSource struct {
	Type    string  `json:"type"`
	ID      string  `json:"id,omitempty"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	W       float64 `json:"w,omitempty"`
	H       float64 `json:"h,omitempty"`
	Src     string  `json:"src,omitempty"`
	Name    string  `json:"name,omitempty"`
	Caption string  `json:"caption,omitempty"`
	Icon    string  `json:"icon,omitempty"`
	Label   string  `json:"label,omitempty"`
	Sub     string  `json:"sub,omitempty"`
	Surface string  `json:"surface,omitempty"`
	Text    string  `json:"text,omitempty"`
}

func intakeFinite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func (r *renderer) planIntakeArchitectureScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, false, err
	}
	fields := map[string]string{"logoslot": "h src name caption", "device": "icon label sub", "plane": "h surface label text"}
	allowed, ok := fields[head.Type]
	if !ok {
		return nil, false, nil
	}
	var n intakeArchitectureSource
	if err := primitiveDecode(raw, &n, allowed); err != nil {
		return nil, true, err
	}
	if n.Type == "device" && n.W == 0 {
		n.W = 108
	}
	if !intakeFinite(n.X, n.Y, n.W, n.H) || n.W <= 0 || n.W > 1920 || n.H < 0 || n.H > 1080 || math.Abs(n.X) > 3840 || math.Abs(n.Y) > 2160 {
		return nil, true, fmt.Errorf("scene.invalid_intake_geometry: %s", id)
	}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	p := &scenePlan{ID: id, Bounds: Rect{n.X, n.Y, n.W, n.H}}
	var err error
	switch n.Type {
	case "logoslot":
		if n.H <= 0 {
			return nil, true, fmt.Errorf("scene.logoslot_requires_height: %s", id)
		}
		err = r.intakeLogoSlot(p, n, surface)
	case "device":
		if strings.TrimSpace(n.Label) == "" {
			return nil, true, fmt.Errorf("scene.device_requires_label: %s", id)
		}
		icon := n.Icon
		if icon == "" {
			icon = "monitor"
		}
		if n.W < 36 {
			return nil, true, fmt.Errorf("scene.device_icon_does_not_fit: %s", id)
		}
		var child *scenePlan
		child, err = r.planIconScene(id+".icon", icon, 36, Rect{n.X + (n.W-36)/2, n.Y, 36, 36}, surface, "display")
		if err != nil {
			break
		}
		diagramMerge(p, child)
		y := n.Y + 39
		err = r.diagramText(p, id+".label", n.Label, "small", Rect{n.X, y, n.W, 0}, surface, "display", "center", 600, false)
		if err != nil {
			break
		}
		last := p.Items[len(p.Items)-1].Text
		y = last.Rect.Y + last.Rect.H + 3
		if n.Sub != "" {
			err = r.diagramText(p, id+".sub", n.Sub, "label", Rect{n.X, y, n.W, 0}, surface, "secondary", "center", 0, false)
		}
	case "plane":
		if n.H <= 0 || strings.TrimSpace(n.Label) == "" || strings.TrimSpace(n.Text) == "" {
			return nil, true, fmt.Errorf("scene.plane_requires_height_label_text: %s", id)
		}
		if n.Surface == "" {
			n.Surface = "strong"
		}
		fill, e := r.sceneColor(n.Surface, "bg")
		if e != nil {
			return nil, true, e
		}
		err = r.diagramShape(p, id+".plane", p.Bounds, pptx.ShapeTypeCustGeom, fill, "070154", .75, "solid", [][2]float64{{0, n.H / 2}, {n.W / 2, 0}, {n.W, n.H / 2}, {n.W / 2, n.H}})
		if err != nil {
			break
		}
		token := "body"
		if n.W > 300 {
			token = "subhead"
		}
		labelStyle, e := r.sceneStyle("label")
		if e != nil {
			return nil, true, e
		}
		textStyle, e := r.sceneStyle(token)
		if e != nil {
			return nil, true, e
		}
		textStyle.Weight = 600
		width := n.W * .6
		lh, e := r.intakeTextHeight(n.Label, labelStyle, width)
		if e != nil {
			return nil, true, e
		}
		th, e := r.intakeTextHeight(n.Text, textStyle, width)
		if e != nil {
			return nil, true, e
		}
		if lh+2+th > n.H+.02 {
			return nil, true, fmt.Errorf("scene.plane_text_overflow: %s", id)
		}
		y := n.Y + (n.H-lh-2-th)/2
		err = r.diagramStyledText(p, id+".label", n.Label, labelStyle, Rect{n.X + n.W*.2, y, width, lh}, n.Surface, "primary", "center", false)
		if err == nil {
			err = r.diagramStyledText(p, id+".text", n.Text, textStyle, Rect{n.X + n.W*.2, y + lh + 2, width, th}, n.Surface, "display", "center", false)
		}
		if err == nil {
			err = intakePlaneTextEnvelope(p, Rect{n.X, n.Y, n.W, n.H})
		}
	}
	if err != nil {
		return nil, true, err
	}
	if len(p.Items) > 512 {
		return nil, true, fmt.Errorf("scene.intake_shape_budget: %s", id)
	}
	diagramFinish(p, n.Type)
	for i := range p.Groups {
		if p.Groups[i].ID == id {
			p.Groups[i].Contract = IntakeArchitectureContract
		}
	}
	return p, true, nil
}

func (r *renderer) intakeTextHeight(text string, style Style, width float64) (float64, error) {
	l, e := r.measureText(text, style, width)
	if e != nil {
		return 0, e
	}
	return math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight), nil
}

func (r *renderer) intakeLogoSlot(p *scenePlan, n intakeArchitectureSource, surface string) error {
	b := p.Bounds
	if n.Src != "" {
		if b.W <= 12 || b.H <= 12 {
			return fmt.Errorf("scene.logoslot_padding_does_not_fit: %s", p.ID)
		}
		data, a, e := r.primitiveAssetBytes(n.Src)
		if e != nil {
			return e
		}
		inner := Rect{b.X + 6, b.Y + 6, b.W - 12, b.H - 12}
		iw, ih := 0., 0.
		isSVG := strings.HasSuffix(strings.ToLower(a.Path), ".svg") || bytes.Contains(data[:min(len(data), 512)], []byte("<svg"))
		if isSVG {
			vb, e := primitiveSVGViewBox(data)
			if e != nil {
				return e
			}
			iw, ih = vb[2], vb[3]
		} else {
			cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
			if e != nil {
				return e
			}
			if int64(cfg.Width)*int64(cfg.Height) > 64_000_000 {
				return fmt.Errorf("scene.logo_decode_pixel_budget: %s", p.ID)
			}
			iw, ih = float64(cfg.Width), float64(cfg.Height)
		}
		if !intakeFinite(iw, ih) || iw <= 0 || ih <= 0 {
			return fmt.Errorf("scene.invalid_logo_asset_dimensions: %s", p.ID)
		}
		scale := math.Min(inner.W/iw, inner.H/ih)
		fit := Rect{inner.X + (inner.W-iw*scale)/2, inner.Y + (inner.H-ih*scale)/2, iw * scale, ih * scale}
		im, e := r.primitiveMediaImage(p.ID+".image", n.Src, fit, "", false)
		if e != nil {
			return e
		}
		im.Sizing = nil
		im.PositionProps = pos(fit)
		im.AltText = n.Name + "; canonical SHA256=" + a.SHA256
		p.Items = append(p.Items, sceneItem{Image: im})
	} else {
		if e := r.diagramShape(p, p.ID+".background", b, pptx.ShapeTypeRect, "FFFFFF", "", 0, "", nil); e != nil {
			return e
		}
		// CSS repeating-linear-gradient(135deg): clipped native polygons retain
		// the six-point perpendicular stripe width without bleeding past the box.
		stripe := 6 * math.Sqrt2
		for lo, index := 0., 0; lo < b.W+b.H; lo, index = lo+2*stripe, index+1 {
			points := intakeClipBand(b.W, b.H, lo+stripe, lo+2*stripe)
			if len(points) < 3 {
				continue
			}
			if e := r.diagramShape(p, fmt.Sprintf("%s.hatch.%03d", p.ID, index), b, pptx.ShapeTypeCustGeom, "F3F6FC", "", 0, "", points); e != nil {
				return e
			}
		}
		if e := r.diagramShape(p, p.ID+".border", b, pptx.ShapeTypeRect, "", "97A4BA", 1, "solid", nil); e != nil {
			return e
		}
		label := n.Name
		if label == "" {
			label = "Logo"
		}
		if e := r.diagramText(p, p.ID+".placeholder", label, "label", Rect{b.X + 2, b.Y, b.W - 4, b.H}, surface, "secondary", "center", 600, true); e != nil {
			return e
		}
	}
	if n.Caption != "" {
		return r.diagramText(p, p.ID+".caption", n.Caption, "small", Rect{b.X, b.Y + b.H + 6, b.W, 0}, surface, "secondary", "left", 0, false)
	}
	return nil
}

// Clip the local rectangle to lo <= x+y <= hi.
func intakeClipBand(w, h, lo, hi float64) [][2]float64 {
	points := [][2]float64{{0, 0}, {w, 0}, {w, h}, {0, h}}
	for _, bound := range []struct{ value, sign float64 }{{lo, 1}, {hi, -1}} {
		var out [][2]float64
		for i, current := range points {
			previous := points[(i+len(points)-1)%len(points)]
			dc := (current[0] + current[1] - bound.value) * bound.sign
			dp := (previous[0] + previous[1] - bound.value) * bound.sign
			if (dc >= 0) != (dp >= 0) {
				t := dp / (dp - dc)
				out = append(out, [2]float64{previous[0] + t*(current[0]-previous[0]), previous[1] + t*(current[1]-previous[1])})
			}
			if dc >= 0 {
				out = append(out, current)
			}
		}
		points = out
		if len(points) == 0 {
			break
		}
	}
	return points
}

// Check each line against the actual diamond, rather than only its rectangular
// text container. Measured occupied bounds remain an estimate until native QA.
func intakePlaneTextEnvelope(p *scenePlan, diamond Rect) error {
	center := diamond.Y + diamond.H/2
	for _, item := range p.Items {
		if item.Text == nil {
			continue
		}
		tr := item.Text
		l := tr.Layout
		terminal := l.EstimatedOccupiedHeight - float64(len(l.Lines)-1)*l.Style.Leading
		for i, line := range l.Lines {
			top := tr.Rect.Y + l.OccupiedTop + float64(i)*l.Style.Leading
			bottom := top + terminal
			distance := math.Max(math.Abs(top-center), math.Abs(bottom-center))
			available := diamond.W*math.Max(0, 1-2*distance/diamond.H) - 1.5
			if line.Advance > available+.02 {
				return fmt.Errorf("scene.plane_text_outside_diamond: %s line %d width %.3fpt capacity %.3fpt", tr.ID, i+1, line.Advance, available)
			}
		}
	}
	return nil
}
