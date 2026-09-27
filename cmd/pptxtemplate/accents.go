package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

type accentEvidence struct {
	Schema      string   `json:"schema"`
	Scene       artifact `json:"scene"`
	Measurement artifact `json:"measurement"`
	Placement   artifact `json:"placement"`
	Asset       artifact `json:"asset"`
	NativeDeck  artifact `json:"native_deck"`
	NativeSlide int      `json:"native_slide"`
	TextID      string   `json:"text_object_id"`
	PictureID   string   `json:"picture_object_id"`
	Phrase      string   `json:"phrase"`
	AssetPart   string   `json:"asset_part"`
}

type accentMeasurement struct {
	Presentation string                                     `json:"presentation"`
	Slide        int                                        `json:"slide"`
	ShapeIndex   int                                        `json:"shape_index"`
	ShapeName    string                                     `json:"shape_name"`
	Phrase       string                                     `json:"phrase"`
	Text         string                                     `json:"text"`
	Rotation     *float64                                   `json:"rotation_degrees"`
	ShapeBounds  struct{ Left, Top, Width, Height float64 } `json:"shape_bounds"`
	Bounds       struct{ Left, Top, Width, Height float64 } `json:"bounds"`
	Lines        []json.RawMessage                          `json:"lines"`
}

type accentPlacement struct {
	Status       string                                 `json:"status"`
	Asset        string                                 `json:"asset"`
	PhraseBounds *struct{ X, Y, Width, Height float64 } `json:"phrase_bounds_points"`
	Mode         string                                 `json:"mode"`
	Phrase       string                                 `json:"phrase"`
	Placement    *struct{ X, Y, CX, CY int64 }          `json:"placement_emu"`
	Rotation     *int64                                 `json:"rotation_ooxml_60000"`
}

func accentNodeID(n *nativepkg.Node) string {
	var id string
	for _, name := range []string{"nvSpPr", "nvPicPr", "nvGraphicFramePr", "nvCxnSpPr", "nvGrpSpPr"} {
		if c := n.Child(name); c != nil {
			if p := c.Child("cNvPr"); p != nil {
				id = p.Attr("id")
			}
		}
	}
	return id
}
func accentObject(scene *nativepkg.Node, id string) (*nativepkg.Node, *nativepkg.Node, int, error) {
	cs := scene.Child("cSld")
	if cs == nil {
		return nil, nil, 0, fmt.Errorf("missing cSld")
	}
	tree := cs.Child("spTree")
	if tree == nil {
		return nil, nil, 0, fmt.Errorf("missing spTree")
	}
	var found *nativepkg.Node
	index := 0
	nativeIndex := 0
	for _, n := range tree.Children {
		if accentNodeID(n) != "" {
			nativeIndex++
		}
		if accentNodeID(n) == id {
			if found != nil {
				return nil, nil, 0, fmt.Errorf("ambiguous object ID %s", id)
			}
			found = n
			index = nativeIndex
		}
	}
	if found == nil {
		return nil, nil, 0, fmt.Errorf("object %s is absent or nested; grouped accents require manual placement", id)
	}
	return found, tree, index, nil
}
func accentXfrm(n *nativepkg.Node) (*nativepkg.Node, error) {
	p := n.Child("spPr")
	if p == nil {
		return nil, fmt.Errorf("object has no direct spPr")
	}
	x := p.Child("xfrm")
	if x == nil || x.Child("off") == nil || x.Child("ext") == nil {
		return nil, fmt.Errorf("object has no explicit transform")
	}
	return x, nil
}
func accentGeometry(n *nativepkg.Node) ([]float64, error) {
	x, e := accentXfrm(n)
	if e != nil {
		return nil, e
	}
	out := []float64{}
	for _, a := range []struct {
		n *nativepkg.Node
		k string
	}{{x.Child("off"), "x"}, {x.Child("off"), "y"}, {x.Child("ext"), "cx"}, {x.Child("ext"), "cy"}} {
		v, e := strconv.ParseFloat(a.n.Attr(a.k), 64)
		if e != nil {
			return nil, e
		}
		out = append(out, v/12700)
	}
	return out, nil
}
func accentText(n *nativepkg.Node) string {
	var parts []string
	n.Walk(func(p *nativepkg.Node) {
		if p.Name == "a:p" {
			var b strings.Builder
			p.Walk(func(t *nativepkg.Node) {
				if t.Name == "a:t" {
					b.WriteString(t.TextContent())
				}
				if t.Name == "a:br" {
					b.WriteString(" ")
				}
			})
			parts = append(parts, b.String())
		}
	})
	return strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
}
func accentNativeSlide(deck string, num int, expectedPart string) (*nativepkg.Node, error) {
	r, e := zip.OpenReader(deck)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	fetch := func(name string) (*nativepkg.Node, error) {
		for _, f := range r.File {
			if f.Name == name {
				r, e := f.Open()
				if e != nil {
					return nil, e
				}
				defer r.Close()
				b, e := io.ReadAll(r)
				if e != nil {
					return nil, e
				}
				return nativepkg.Parse(b)
			}
		}
		return nil, fmt.Errorf("missing native part %s", name)
	}
	p, e := fetch("ppt/presentation.xml")
	if e != nil {
		return nil, e
	}
	list := p.Child("sldIdLst")
	if list == nil {
		return nil, fmt.Errorf("no native slide list")
	}
	var ids []*nativepkg.Node
	for _, n := range list.Children {
		if n.Name == "p:sldId" {
			ids = append(ids, n)
		}
	}
	if num < 1 || num > len(ids) {
		return nil, fmt.Errorf("native slide index out of range")
	}
	rels, e := fetch("ppt/_rels/presentation.xml.rels")
	if e != nil {
		return nil, e
	}
	for _, r := range rels.Children {
		if r.Attr("Id") == ids[num-1].Attr("r:id") {
			target := filepath.Clean(filepath.Join("ppt", r.Attr("Target")))
			if strings.HasPrefix(r.Attr("Target"), "/") {
				target = strings.TrimPrefix(r.Attr("Target"), "/")
			}
			if target != expectedPart {
				return nil, fmt.Errorf("native slide part mismatch")
			}
			return fetch(target)
		}
	}
	return nil, fmt.Errorf("native slide relationship missing")
}

// apply-accent accepts pinned solver output only after checking the measured
// native object's text, name, order and geometry against the exact input scene.
func applyAccent(args []string) error {
	f := flag.NewFlagSet("apply-accent", flag.ContinueOnError)
	proof := f.String("evidence", "", "pinned accent evidence JSON")
	out := f.String("out", "", "new output scene JSON")
	if e := f.Parse(args); e != nil {
		return e
	}
	if *proof == "" || *out == "" || f.NArg() != 0 {
		return fmt.Errorf("apply-accent requires --evidence and --out")
	}
	var ev accentEvidence
	if e := read(*proof, &ev); e != nil {
		return e
	}
	if ev.Schema != "pptxgengo.accent-evidence.v1" || ev.Phrase == "" {
		return fmt.Errorf("invalid accent evidence")
	}
	for _, a := range []artifact{ev.Scene, ev.Measurement, ev.Placement, ev.Asset, ev.NativeDeck} {
		h, e := hashFile(a.Path)
		if e != nil {
			return e
		}
		if h != a.SHA {
			return fmt.Errorf("changed accent evidence %s", a.Path)
		}
	}
	var s nativepkg.Slide
	if e := read(ev.Scene.Path, &s); e != nil {
		return e
	}
	if e := nativepkg.ApplyBindings(s.Scene, s.Bindings); e != nil {
		return e
	}
	text, tree, index, e := accentObject(s.Scene, ev.TextID)
	if e != nil {
		return e
	}
	pic, _, _, e := accentObject(s.Scene, ev.PictureID)
	if e != nil {
		return e
	}
	if pic.Name != "p:pic" || text.Name != "p:sp" {
		return fmt.Errorf("requires a top-level picture and text shape")
	}
	tx, e := accentXfrm(text)
	if e != nil {
		return e
	}
	rotation := tx.Attr("rot")
	if rotation != "" && rotation != "0" {
		return fmt.Errorf("rotated text requires manual placement")
	}
	var m accentMeasurement
	if e = read(ev.Measurement.Path, &m); e != nil {
		return e
	}
	if m.Phrase != ev.Phrase || m.Slide != ev.NativeSlide || m.ShapeIndex != index || m.Presentation != filepath.Base(ev.NativeDeck.Path) || len(m.Lines) != 1 || m.Rotation == nil || *m.Rotation != 0 {
		return fmt.Errorf("measurement identity, line count or rotation mismatch")
	}
	if accentText(text) != strings.Join(strings.Fields(m.Text), " ") || strings.Count(m.Text, ev.Phrase) != 1 {
		return fmt.Errorf("measured text differs from scene or phrase is ambiguous")
	}
	native, e := accentNativeSlide(ev.NativeDeck.Path, ev.NativeSlide, s.Part)
	if e != nil {
		return e
	}
	nt, _, ni, e := accentObject(native, ev.TextID)
	if e != nil {
		return e
	}
	np, _, _, e := accentObject(native, ev.PictureID)
	if e != nil {
		return e
	}
	if ni != index || accentText(nt) != accentText(text) {
		return fmt.Errorf("native deck text/order differs from scene")
	}
	ntx, e := accentXfrm(nt)
	if e != nil {
		return e
	}
	if ntx.Attr("rot") != "" && ntx.Attr("rot") != "0" {
		return fmt.Errorf("native text is rotated")
	}
	nv := text.Child("nvSpPr")
	if nv == nil || nv.Child("cNvPr") == nil {
		return fmt.Errorf("missing text identity")
	}
	name := nv.Child("cNvPr").Attr("name")
	if name != m.ShapeName {
		return fmt.Errorf("native shape name mismatch")
	}
	tg, e := accentGeometry(text)
	if e != nil {
		return e
	}
	ng, e := accentGeometry(nt)
	if e != nil {
		return e
	}
	mg := []float64{m.ShapeBounds.Left, m.ShapeBounds.Top, m.ShapeBounds.Width, m.ShapeBounds.Height}
	pg, e := accentGeometry(pic)
	if e != nil {
		return e
	}
	npg, e := accentGeometry(np)
	if e != nil {
		return e
	}
	for i := range tg {
		if math.Abs(tg[i]-ng[i]) > 0.01 || math.Abs(tg[i]-mg[i]) > 0.05 || math.Abs(pg[i]-npg[i]) > 0.01 {
			return fmt.Errorf("native/scene shape geometry mismatch")
		}
	}
	var p accentPlacement
	if e = read(ev.Placement.Path, &p); e != nil {
		return e
	}
	if p.Status != "placed" || p.Placement == nil || p.Phrase != ev.Phrase || (p.Mode != "highlight" && p.Mode != "underline") {
		return fmt.Errorf("solver did not produce a valid placement")
	}
	if p.Asset != ev.Asset.Path || p.PhraseBounds == nil {
		return fmt.Errorf("solver asset or measured bounds missing/mismatched")
	}
	pb := p.PhraseBounds
	for _, d := range []float64{pb.X - m.Bounds.Left, pb.Y - m.Bounds.Top, pb.Width - m.Bounds.Width, pb.Height - m.Bounds.Height} {
		if math.IsNaN(d) || math.Abs(d) > 0.0001 {
			return fmt.Errorf("solver phrase bounds differ from measurement")
		}
	}
	// The measured alpha asset must be one of this native picture's embedded images.
	linked := false
	embeds := map[string]bool{}
	pic.Walk(func(n *nativepkg.Node) {
		if id := n.Attr("r:embed"); id != "" {
			embeds[id] = true
		}
	})
	for _, rel := range s.Relationships.Children {
		if embeds[rel.Attr("Id")] {
			part := filepath.Clean(filepath.Join(filepath.Dir(s.Part), rel.Attr("Target")))
			if strings.HasPrefix(rel.Attr("Target"), "/") {
				part = strings.TrimPrefix(rel.Attr("Target"), "/")
			}
			if part == ev.AssetPart {
				linked = true
			}
		}
	}
	if !linked {
		return fmt.Errorf("asset is not embedded by the target picture")
	}
	zr, e := zip.OpenReader(ev.NativeDeck.Path)
	if e != nil {
		return e
	}
	defer zr.Close()
	nativeEmbeds := map[string]bool{}
	np.Walk(func(n *nativepkg.Node) {
		if id := n.Attr("r:embed"); id != "" {
			nativeEmbeds[id] = true
		}
	})
	nativeAssetLinked := false
	relpart := filepath.Join(filepath.Dir(s.Part), "_rels", filepath.Base(s.Part)+".rels")
	for _, zf := range zr.File {
		if zf.Name == relpart {
			r, e := zf.Open()
			if e != nil {
				return e
			}
			b, e := io.ReadAll(r)
			r.Close()
			if e != nil {
				return e
			}
			rels, e := nativepkg.Parse(b)
			if e != nil {
				return e
			}
			for _, rel := range rels.Children {
				if nativeEmbeds[rel.Attr("Id")] {
					target := filepath.Clean(filepath.Join(filepath.Dir(s.Part), rel.Attr("Target")))
					if strings.HasPrefix(rel.Attr("Target"), "/") {
						target = strings.TrimPrefix(rel.Attr("Target"), "/")
					}
					if target == ev.AssetPart {
						nativeAssetLinked = true
					}
				}
			}
		}
	}
	if !nativeAssetLinked {
		return fmt.Errorf("native picture embeds a different asset")
	}
	assetMatched := false
	for _, zf := range zr.File {
		if zf.Name == ev.AssetPart {
			r, e := zf.Open()
			if e != nil {
				return e
			}
			b, e := io.ReadAll(r)
			r.Close()
			if e != nil {
				return e
			}
			h := sha256.Sum256(b)
			assetMatched = hex.EncodeToString(h[:]) == ev.Asset.SHA
		}
	}
	if !assetMatched {
		return fmt.Errorf("native picture asset hash mismatch")
	}
	v := p.Placement
	if v.CX <= 0 || v.CY <= 0 {
		return fmt.Errorf("invalid placement extent")
	}
	x, e := accentXfrm(pic)
	if e != nil {
		return e
	}
	nx, e := accentXfrm(np)
	if e != nil {
		return e
	}
	for _, k := range []string{"rot", "flipH", "flipV"} {
		if x.Attr(k) != nx.Attr(k) {
			return fmt.Errorf("native picture transform differs from scene")
		}
	}
	fill := pic.Child("blipFill")
	if fill == nil {
		return fmt.Errorf("picture has no blipFill")
	}
	crop := fill.Child("srcRect")
	if crop != nil {
		for _, a := range crop.Attrs {
			if a.Value != "0" {
				return fmt.Errorf("cropped pictures require manual placement")
			}
		}
	}
	if x.Attr("flipH") == "1" || x.Attr("flipV") == "1" {
		return fmt.Errorf("flipped pictures require manual placement")
	}
	if p.Mode == "underline" {
		sourceRotation, _ := strconv.ParseInt(x.Attr("rot"), 10, 64)
		if p.Rotation == nil || *p.Rotation != sourceRotation {
			return fmt.Errorf("underline solver must preserve source rotation")
		}
	}
	x.Child("off").SetAttr("x", strconv.FormatInt(v.X, 10))
	x.Child("off").SetAttr("y", strconv.FormatInt(v.Y, 10))
	x.Child("ext").SetAttr("cx", strconv.FormatInt(v.CX, 10))
	x.Child("ext").SetAttr("cy", strconv.FormatInt(v.CY, 10))
	if p.Rotation != nil {
		x.SetAttr("rot", strconv.FormatInt(*p.Rotation, 10))
	}
	// Keep a highlight behind its text without disturbing any other object order.
	if p.Mode == "highlight" {
		pi, ti := -1, -1
		for i, n := range tree.Children {
			if n == pic {
				pi = i
			}
			if n == text {
				ti = i
			}
		}
		if pi > ti {
			tree.Children = append(tree.Children[:pi], tree.Children[pi+1:]...)
			tree.Children = append(tree.Children, nil)
			copy(tree.Children[ti+1:], tree.Children[ti:])
			tree.Children[ti] = pic
		}
	}
	s.Bindings = nativepkg.ExtractBindings(s.Scene)
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	file, e := os.OpenFile(*out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	_, e = file.Write(append(b, '\n'))
	ce := file.Close()
	if e != nil {
		return e
	}
	return ce
}

func adaptAccents(args []string) error {
	// Native measurement is a macOS adapter. Geometry validation/application stays
	// in this Go CLI; the script orchestrates PowerPoint and the existing solver.
	root := "."
	for i, a := range args {
		if strings.HasPrefix(a, "--root=") {
			root = strings.TrimPrefix(a, "--root=")
		}
		if a == "--root" && i+1 < len(args) {
			root = args[i+1]
		}
	}
	exe, e := os.Executable()
	if e != nil {
		return e
	}
	cmd := exec.Command("python3", append([]string{filepath.Join(root, "scripts/adapt-template-accents.py"), "--template-bin", exe}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
