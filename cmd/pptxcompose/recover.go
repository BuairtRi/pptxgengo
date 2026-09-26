package main

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/internal/nativepkg"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type recoveredChange struct {
	SlideID string `json:"slide_id"`
	Shape   string `json:"shape"`
	Before  string `json:"before"`
	After   string `json:"after"`
}

func recoverZip(file string) (map[string][]byte, error) {
	z, e := zip.OpenReader(file)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	out := map[string][]byte{}
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if _, ok := out[f.Name]; ok {
			return nil, fmt.Errorf("duplicate package part")
		}
		r, e := f.Open()
		if e != nil {
			return nil, e
		}
		b, e := io.ReadAll(io.LimitReader(r, 64*1024*1024+1))
		r.Close()
		if e != nil {
			return nil, e
		}
		if len(b) > 64*1024*1024 {
			return nil, fmt.Errorf("package part exceeds 64MB")
		}
		out[f.Name] = b
	}
	return out, nil
}

type returnedSlide struct {
	Part  string
	Scene *nativepkg.Node
}

func recoveryObjects(scene *nativepkg.Node) (map[string]*nativepkg.Node, error) {
	if scene == nil || scene.Child("cSld") == nil || scene.Child("cSld").Child("spTree") == nil {
		return nil, fmt.Errorf("missing shape tree")
	}
	objects := map[string]*nativepkg.Node{}
	for _, n := range scene.Child("cSld").Child("spTree").Children {
		switch localName(n.Name) {
		case "sp", "pic", "cxnSp":
			name := shapeName(n)
			if name == "" || objects[name] != nil {
				return nil, fmt.Errorf("missing/duplicate stable object name")
			}
			objects[name] = n
		case "nvGrpSpPr", "grpSpPr", "extLst":
		default:
			return nil, fmt.Errorf("unsupported object in shape tree; use pptxscene")
		}
	}
	return objects, nil
}

func recoveryKindMatches(n *nativepkg.Node, kind string) bool {
	if n == nil {
		return false
	}
	tag := localName(n.Name)
	switch kind {
	case "text":
		return tag == "sp" && n.Child("txBody") != nil
	case "surface":
		return tag == "sp"
	case "shape":
		return tag == "sp"
	case "image":
		return tag == "pic"
	case "line":
		p := n.Child("spPr")
		return (tag == "sp" || tag == "cxnSp") && p != nil && p.Child("prstGeom") != nil && p.Child("prstGeom").Attr("prst") == "line"
	}
	return false
}

func recoveryPageSize(parts map[string][]byte, slides []renderSlide) error {
	p, e := nativepkg.Parse(parts["ppt/presentation.xml"])
	if e != nil {
		return e
	}
	if p == nil || p.Child("sldSz") == nil {
		return fmt.Errorf("missing presentation size")
	}
	s := p.Child("sldSz")
	w, we := strconv.ParseFloat(s.Attr("cx"), 64)
	h, he := strconv.ParseFloat(s.Attr("cy"), 64)
	if we != nil || he != nil {
		return fmt.Errorf("invalid presentation size")
	}
	for _, sl := range slides {
		if !near(w/12700, sl.Width) || !near(h/12700, sl.Height) {
			return fmt.Errorf("presentation size changed; use pptxscene")
		}
	}
	return nil
}

func packageSlides(parts map[string][]byte) ([]returnedSlide, error) {
	p, e := nativepkg.Parse(parts["ppt/presentation.xml"])
	if e != nil {
		return nil, e
	}
	rels, e := nativepkg.Parse(parts["ppt/_rels/presentation.xml.rels"])
	if e != nil {
		return nil, e
	}
	targets := map[string]string{}
	for _, r := range rels.Children {
		if strings.HasSuffix(r.Attr("Type"), "/slide") && r.Attr("TargetMode") != "External" {
			t := r.Attr("Target")
			if strings.HasPrefix(t, "/") {
				t = strings.TrimPrefix(t, "/")
			} else {
				t = path.Clean("ppt/" + t)
			}
			targets[r.Attr("Id")] = t
		}
	}
	if p.Child("sldIdLst") == nil {
		return nil, fmt.Errorf("missing slide list")
	}
	var slides []returnedSlide
	for _, id := range p.Child("sldIdLst").Children {
		if id.Name == "" {
			continue
		}
		t := targets[id.Attr("r:id")]
		if t == "" {
			return nil, fmt.Errorf("unresolved slide relationship")
		}
		s, e := nativepkg.Parse(parts[t])
		if e != nil {
			return nil, e
		}
		slides = append(slides, returnedSlide{t, s})
	}
	return slides, nil
}
func localName(s string) string {
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		return s[i+1:]
	}
	return s
}
func shapeName(n *nativepkg.Node) string {
	name := ""
	n.Walk(func(q *nativepkg.Node) {
		if name == "" && localName(q.Name) == "cNvPr" {
			name = q.Attr("name")
		}
	})
	return name
}
func sceneText(n *nativepkg.Node) string {
	body := n.Child("txBody")
	if body == nil {
		return ""
	}
	var paragraphs []string
	for _, p := range body.Children {
		if localName(p.Name) != "p" {
			continue
		}
		var b strings.Builder
		for _, r := range p.Children {
			switch localName(r.Name) {
			case "r", "fld":
				if t := r.Child("t"); t != nil {
					b.WriteString(t.TextContent())
				}
			case "br":
				b.WriteByte('\n')
			}
		}
		paragraphs = append(paragraphs, b.String())
	}
	return strings.Join(paragraphs, "\n")
}
func sceneFrame(n *nativepkg.Node) (frame, error) {
	props := n.Child("spPr")
	if props == nil {
		return frame{}, fmt.Errorf("unsupported object frame")
	}
	x := props.Child("xfrm")
	if x == nil || x.Child("off") == nil || x.Child("ext") == nil {
		return frame{}, fmt.Errorf("missing explicit frame")
	}
	v := func(s string) (float64, error) {
		n, e := strconv.ParseFloat(s, 64)
		if e != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return 0, fmt.Errorf("invalid frame coordinate %q", s)
		}
		return n / 12700, nil
	}
	var f frame
	var e error
	if f.X, e = v(x.Child("off").Attr("x")); e != nil {
		return f, e
	}
	if f.Y, e = v(x.Child("off").Attr("y")); e != nil {
		return f, e
	}
	if f.Width, e = v(x.Child("ext").Attr("cx")); e != nil {
		return f, e
	}
	if f.Height, e = v(x.Child("ext").Attr("cy")); e != nil {
		return f, e
	}
	if r := x.Attr("rot"); r != "" && r != "0" {
		return f, fmt.Errorf("rotated object")
	}
	for _, attr := range []string{"flipH", "flipV"} {
		if v := x.Attr(attr); v != "" && v != "0" && v != "false" {
			return f, fmt.Errorf("reflected object")
		}
	}
	return f, nil
}
func setRecoveredText(s *compose.SlideSpec, name, value string) error {
	if name == "slide-title" {
		s.Title = value
		return nil
	}
	decode := func(v string) string { b, _ := base64.RawURLEncoding.DecodeString(v); return string(b) }
	if strings.HasPrefix(name, "canvas:") {
		id := decode(strings.TrimPrefix(name, "canvas:"))
		for i := range s.Canvas {
			if s.Canvas[i].ID == id && s.Canvas[i].Kind == "text" {
				s.Canvas[i].Text = value
				return nil
			}
		}
		for i := range s.Layouts {
			c := &s.Layouts[i]
			for j := range c.Cells {
				cell := &c.Cells[j]
				for k := range cell.Blocks {
					b := &cell.Blocks[k]
					key := c.ID + "/" + cell.ID + "/" + b.ID
					if b.Marker != "" {
						key += "/text"
					}
					if key == id {
						b.Text = value
						return nil
					}
				}
			}
		}
	}
	if strings.HasPrefix(name, "role:") {
		id := decode(strings.TrimPrefix(name, "role:"))
		for i := range s.Roles {
			if s.Roles[i].ID == id {
				s.Roles[i].Label = value
				return nil
			}
		}
		for i := range s.Pods {
			for j := range s.Pods[i].Roles {
				if s.Pods[i].Roles[j].ID == id {
					s.Pods[i].Roles[j].Label = value
					return nil
				}
			}
		}
	}
	for i := range s.Pods {
		if name == "pod:"+base64.RawURLEncoding.EncodeToString([]byte(s.Pods[i].ID))+"-title" {
			s.Pods[i].Title = value
			return nil
		}
	}
	for i := range s.Phases {
		if name == "phase:"+base64.RawURLEncoding.EncodeToString([]byte(s.Phases[i].ID))+"-label" {
			s.Phases[i].Title = value
			return nil
		}
	}
	for i := range s.Cards {
		c := &s.Cards[i]
		prefix := "card:" + base64.RawURLEncoding.EncodeToString([]byte(c.ID)) + "-"
		if strings.HasPrefix(name, prefix) {
			switch strings.TrimPrefix(name, prefix) {
			case "number":
				c.Number = value
			case "title":
				c.Title = value
			case "body":
				c.Body = value
			case "value":
				c.Value = value
			case "label":
				c.Label = value
			default:
				return fmt.Errorf("unsupported card text slot")
			}
			return nil
		}
	}
	return fmt.Errorf("text in derived or unsupported slot %s cannot be recovered; use pptxscene", name)
}

// RecoverText is intentionally bounded to stable names, slide order and frames.
// Formatting is regenerated from the original spec; --text-only is mandatory.
func recoverText(specPath, bundleDir, returned, out string, textOnly bool) error {
	if !textOnly {
		return fmt.Errorf("recover-text requires --text-only: it restores original spec styling and imports text only")
	}
	if _, e := os.Lstat(out); !os.IsNotExist(e) {
		return fmt.Errorf("output must be a new directory")
	}
	var spec compose.Spec
	raw, e := readJSON(specPath, &spec)
	if e != nil {
		return e
	}
	var m manifest
	_, e = readJSON(filepath.Join(bundleDir, "manifest.json"), &m)
	if e != nil {
		return e
	}
	if m.Plan == nil || m.Schema != "pptxgengo.compose-bundle.v1" || m.SpecSHA != hash(raw) {
		return fmt.Errorf("original spec/bundle mismatch")
	}
	original, e := deckPath(bundleDir, m)
	if e != nil {
		return e
	}
	originalBytes, e := os.ReadFile(original)
	if e != nil {
		return e
	}
	if hash(originalBytes) != m.DeckSHA {
		return fmt.Errorf("original bundle hash mismatch")
	}
	parts, e := recoverZip(returned)
	if e != nil {
		return e
	}
	slides, e := packageSlides(parts)
	if e != nil {
		return e
	}
	if len(spec.Slides) != len(m.Slides) {
		return fmt.Errorf("spec/bundle slide count mismatch")
	}
	if len(slides) != len(m.Slides) {
		return fmt.Errorf("slide count changed; use pptxscene")
	}
	originalParts, e := recoverZip(original)
	if e != nil {
		return e
	}
	originalSlides, e := packageSlides(originalParts)
	if e != nil {
		return e
	}
	if len(originalSlides) != len(m.Slides) {
		return fmt.Errorf("original deck/bundle slide count mismatch")
	}
	if e = recoveryPageSize(originalParts, m.Slides); e != nil {
		return e
	}
	if e = recoveryPageSize(parts, m.Slides); e != nil {
		return e
	}
	media := func(p map[string][]byte) map[string]bool {
		r := map[string]bool{}
		for n, b := range p {
			if strings.HasPrefix(n, "ppt/media/") {
				r[hash(b)] = true
			}
		}
		return r
	}
	a, b := media(originalParts), media(parts)
	if len(a) != len(b) {
		return fmt.Errorf("media changed; use pptxscene")
	}
	for h := range a {
		if !b[h] {
			return fmt.Errorf("media changed; use pptxscene")
		}
	}
	var changes []recoveredChange
	for i, sl := range slides {
		originalObjects, err := recoveryObjects(originalSlides[i].Scene)
		if err != nil {
			return err
		}
		objects, err := recoveryObjects(sl.Scene)
		if err != nil {
			return err
		}
		originalCS := originalSlides[i].Scene.Child("cSld")
		if originalCS.Attr("name") != "compose:"+base64.RawURLEncoding.EncodeToString([]byte(m.Slides[i].ID)) || len(originalObjects) != len(m.Slides[i].Elements) {
			return fmt.Errorf("original deck/manifest object identity mismatch")
		}
		// Bind editable manifest entries back to the actual original package.
		for _, el := range m.Slides[i].Elements {
			n := originalObjects[el.Name]
			if !recoveryKindMatches(n, el.Kind) {
				return fmt.Errorf("original deck/manifest object type mismatch: %s", el.Name)
			}
			f, err := sceneFrame(n)
			if err != nil {
				return err
			}
			if !near(f.X, el.Frame.X) || !near(f.Y, el.Frame.Y) || !near(f.Width, el.Frame.Width) || !near(f.Height, el.Frame.Height) {
				return fmt.Errorf("original deck/manifest frame mismatch: %s", el.Name)
			}
			if (el.Kind == "text" && sceneText(n) != el.Text) || (el.Kind != "text" && sceneText(n) != "") {
				return fmt.Errorf("original deck/manifest text mismatch: %s", el.Name)
			}
			if el.Kind == "image" {
				h, err := pictureHash(originalParts, originalSlides[i].Part, n)
				if err != nil {
					return err
				}
				if h != el.AssetSHA256 {
					return fmt.Errorf("original deck/manifest image mismatch: %s", el.Name)
				}
			}
		}
		cs := sl.Scene.Child("cSld")
		if cs == nil || cs.Child("spTree") == nil {
			return fmt.Errorf("missing shape tree")
		}
		if !sameRootTransform(cs.Child("spTree").Child("grpSpPr"), originalSlides[i].Scene.Child("cSld").Child("spTree").Child("grpSpPr")) {
			return fmt.Errorf("root coordinate frame moved/resized; use pptxscene")
		}
		if cs.Attr("name") != "compose:"+base64.RawURLEncoding.EncodeToString([]byte(m.Slides[i].ID)) {
			return fmt.Errorf("slide identity/order is missing or changed; use pptxscene")
		}
		if len(objects) != len(m.Slides[i].Elements) {
			return fmt.Errorf("slide %d object count changed; use pptxscene", i+1)
		}
		for _, el := range m.Slides[i].Elements {
			n := objects[el.Name]
			if n == nil {
				return fmt.Errorf("slide %d missing object %s; slide order/names changed", i+1, el.Name)
			}
			if !recoveryKindMatches(n, el.Kind) {
				return fmt.Errorf("object type changed for %s; use pptxscene", el.Name)
			}
			f, e := sceneFrame(n)
			if e != nil {
				return e
			}
			for _, d := range []float64{f.X - el.Frame.X, f.Y - el.Frame.Y, f.Width - el.Frame.Width, f.Height - el.Frame.Height} {
				if math.Abs(d) > 0.12 {
					return fmt.Errorf("object %s moved/resized; use pptxscene", el.Name)
				}
			}
			if el.Kind != "text" && sceneText(n) != "" {
				return fmt.Errorf("new text in non-text object %s; use pptxscene", el.Name)
			}
			if el.Kind == "image" {
				h, e := pictureHash(parts, sl.Part, n)
				if e != nil {
					return e
				}
				if h != el.AssetSHA256 {
					return fmt.Errorf("image assignment changed for %s; use pptxscene", el.Name)
				}
			}
			if el.Kind == "text" {
				txt := sceneText(n)
				if txt != el.Text {
					if len(el.Paragraphs) > 0 {
						return fmt.Errorf("rich text recovery is unsupported for %s; use pptxscene or edit the source rich-text spec", el.Name)
					}
					if e := setRecoveredText(&spec.Slides[i], el.Name, txt); e != nil {
						return e
					}
					changes = append(changes, recoveredChange{m.Slides[i].ID, el.Name, el.Text, txt})
				}
			}
		}
	}
	if _, e := compose.ProbeRequests(spec); e != nil {
		return fmt.Errorf("recovered content invalid: %w", e)
	}
	returnedBytes, e := os.ReadFile(returned)
	if e != nil {
		return e
	}
	if e = validateShapeStructure(returnedBytes, m.Slides); e != nil {
		return fmt.Errorf("returned roadmap shape semantics changed; use pptxscene: %w", e)
	}
	recoveredBytes := jsonBytes(spec)
	report := map[string]any{"recovered_spec_sha256": hash(recoveredBytes), "schema": "pptxgengo.text-recovery.v1", "original_spec_sha256": m.SpecSHA, "original_deck_sha256": m.DeckSHA, "returned_deck_sha256": hash(returnedBytes), "changes": changes, "requires_new_native_measurement": true, "formatting_policy": "Original spec styling is restored. Returned formatting and notes are not imported.", "limits": "Stable slide order/object names/counts/frames and unchanged media required. This is text recovery, not general semantic reconciliation."}
	parent := filepath.Dir(out)
	if e = os.MkdirAll(parent, 0755); e != nil {
		return e
	}
	tmp, e := os.MkdirTemp(parent, ".recover-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(tmp)
	if e = writeNew(filepath.Join(tmp, "spec.json"), recoveredBytes); e != nil {
		return e
	}
	if e = writeNew(filepath.Join(tmp, "recovery.json"), jsonBytes(report)); e != nil {
		return e
	}
	return os.Rename(tmp, out)
}

func pictureHash(parts map[string][]byte, part string, n *nativepkg.Node) (string, error) {
	fill := n.Child("blipFill")
	if fill == nil || fill.Child("blip") == nil {
		return "", fmt.Errorf("picture binding missing")
	}
	id := fill.Child("blip").Attr("r:embed")
	rp := path.Join(path.Dir(part), "_rels", path.Base(part)+".rels")
	rels, e := nativepkg.Parse(parts[rp])
	if e != nil {
		return "", e
	}
	for _, r := range rels.Children {
		if r.Attr("Id") == id && r.Attr("TargetMode") != "External" {
			t := r.Attr("Target")
			if strings.HasPrefix(t, "/") {
				t = strings.TrimPrefix(t, "/")
			} else {
				t = path.Clean(path.Join(path.Dir(part), t))
			}
			b, ok := parts[t]
			if !ok {
				return "", fmt.Errorf("missing picture resource")
			}
			return hash(b), nil
		}
	}
	return "", fmt.Errorf("unresolved picture binding")
}

func sameRootTransform(a, b *nativepkg.Node) bool {
	for _, n := range []*nativepkg.Node{a, b} {
		if n != nil && n.Child("xfrm") != nil {
			x := n.Child("xfrm")
			for _, attr := range []string{"rot", "flipH", "flipV"} {
				v := x.Attr(attr)
				if v != "" && v != "0" && v != "false" {
					return false
				}
			}
		}
	}
	value := func(n *nativepkg.Node, child, attr string) float64 {
		if n == nil || n.Child("xfrm") == nil || n.Child("xfrm").Child(child) == nil {
			return 0
		}
		v := n.Child("xfrm").Child(child).Attr(attr)
		if v == "" {
			return 0
		}
		x, e := strconv.ParseFloat(v, 64)
		if e != nil {
			return math.NaN()
		}
		return x
	}
	for _, p := range [][2]string{{"off", "x"}, {"off", "y"}, {"ext", "cx"}, {"ext", "cy"}, {"chOff", "x"}, {"chOff", "y"}, {"chExt", "cx"}, {"chExt", "cy"}} {
		x, y := value(a, p[0], p[1]), value(b, p[0], p[1])
		if !finite(x) || !finite(y) || math.Abs(x-y) > 1 {
			return false
		}
	}
	return true
}
