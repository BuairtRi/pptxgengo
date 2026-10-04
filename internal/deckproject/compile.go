package deckproject

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Compilation struct {
	Document    wmdesign.Document
	Assets      map[string]wmdesign.AssetData
	AssetHashes map[string]string
}

func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func strictInto(v any, out any) error {
	d := json.NewDecoder(bytes.NewReader(canonical(v)))
	d.DisallowUnknownFields()
	return d.Decode(out)
}
func Compile(p *Project, bundle, engine string) (Compilation, error) {
	c := Compilation{Document: wmdesign.Document{BuildIdentity: &wmdesign.BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: digest(p.Canonical)}, Title: p.Document.Title, Schema: "pptxgengo.wmds-foundation.v1", Year: p.Document.Year, MediaOptimization: p.Document.MediaOptimization}, Assets: map[string]wmdesign.AssetData{}, AssetHashes: map[string]string{}}
	c.Document.Sections = append([]wmdesign.SectionSpec(nil), p.Document.Sections...)
	source, e := wmdesign.Load(bundle, "")
	if e != nil {
		return c, e
	}
	catalog, e := wmdesign.LibraryCatalog(bundle, "")
	if e != nil {
		return c, e
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	for _, d := range catalog {
		defs[d.Key] = d
	}
	for id, t := range p.Document.LocalTemplates {
		if t.Provenance == nil {
			continue
		}
		v := t.Provenance
		if v.DefinitionSnapshot != "" {
			path, e := SafePath(p.Root, v.DefinitionSnapshot)
			if e != nil {
				return c, e
			}
			raw, e := os.ReadFile(path)
			if e != nil {
				return c, e
			}
			if digest(raw) != v.DefinitionSHA256 {
				return c, fmt.Errorf("local template %s: frozen ancestor snapshot hash mismatch", id)
			}
		} else {
			var data []byte
			switch v.Parent.Scope {
			case "shared":
				d, ok := defs[v.Parent.ID]
				if !ok {
					return c, fmt.Errorf("unknown shared provenance parent")
				}
				var tree any
				if e := json.Unmarshal(d.RawSlide, &tree); e != nil {
					return c, e
				}
				data = canonical(tree)
				if v.SourceFile != "" && (v.SourceFile != d.SourceFile || v.SourceFileSHA256 != d.SourceSHA256) {
					return c, fmt.Errorf("provenance source file pin mismatch")
				}
			case "local":
				parent, ok := p.Document.LocalTemplates[v.Parent.ID]
				if !ok || v.Parent.ID == id {
					return c, fmt.Errorf("unknown/self local provenance parent")
				}
				data = canonical(parent)
			default:
				return c, fmt.Errorf("invalid provenance scope")
			}
			if digest(data) != v.DefinitionSHA256 {
				return c, fmt.Errorf("local template %s: ancestor hash mismatch; preserve a verified frozen snapshot when forking", id)
			}
		}
	}
	assetKeys := map[string]string{}
	registry := map[string]wmdesign.PrimitiveAssetReference{}
	for _, r := range wmdesign.PrimitiveAssetCatalog() {
		registry[r.Key] = r
	}
	for id, a := range p.Document.Assets {
		if a.RegistryID != "" {
			r, ok := registry[a.RegistryID]
			if !ok {
				return c, p.fail("/assets/"+escape(id), "unknown registry_id")
			}
			if a.SHA256 != "" && a.SHA256 != r.SHA256 {
				return c, p.fail("/assets/"+escape(id), "registry hash mismatch")
			}
			assetKeys[id] = r.Key
			c.AssetHashes[id] = r.SHA256
			continue
		}
		path, e := SafePath(p.Root, a.Path)
		if e != nil {
			return c, e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return c, e
		}
		hash := digest(data)
		if a.SHA256 != "" && a.SHA256 != hash {
			return c, p.fail("/assets/"+escape(id), "asset hash mismatch")
		}
		if a.DerivationReceipt != "" {
			rp, e := SafePath(p.Root, a.DerivationReceipt)
			if e != nil {
				return c, e
			}
			if _, e = os.ReadFile(rp); e != nil {
				return c, e
			}
		}
		key := "project:" + id
		assetKeys[id] = key
		c.AssetHashes[id] = hash
		mime, e := assetMIME(data)
		if e != nil {
			return c, p.fail("/assets/"+escape(id), "%v", e)
		}
		c.Assets[key] = wmdesign.AssetData{Data: data, SHA256: hash, MIME: mime}
	}
	for id, a := range p.Document.Assets {
		if a.DerivationReceipt == "" {
			continue
		}
		path, e := SafePath(p.Root, a.DerivationReceipt)
		if e != nil {
			return c, e
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return c, e
		}
		var receipt DerivationReceipt
		if e = strictInto(json.RawMessage(raw), &receipt); e != nil {
			return c, e
		}
		if receipt.Schema != "pptxgengo.asset-derivation.v1" || receipt.SourceAsset != a.DerivedFrom || receipt.SourceSHA256 != c.AssetHashes[a.DerivedFrom] || receipt.ResultSHA256 != c.AssetHashes[id] || strings.TrimSpace(receipt.Operation) == "" {
			return c, p.fail("/assets/"+escape(id)+"/derivation_receipt", "asset derivative receipt does not match original/result bytes")
		}
	}
	for i, slide := range p.Document.Slides {
		ptr := fmt.Sprintf("/slides/%d", i)
		if slide.Template.Scope == "shared" {
			d, ok := defs[slide.Template.ID]
			if !ok {
				return c, p.fail(ptr+"/template", "unknown shared template %s", slide.Template.ID)
			}
			if slide.Template.Revision != "" && slide.Template.Revision != strconv.Itoa(d.Revision) {
				return c, p.fail(ptr+"/template/revision", "revision mismatch: expected %d", d.Revision)
			}
			resolved, e := sharedAssetValues(slide.Values, d, assetKeys, registry)
			if e != nil {
				return c, p.fail(ptr+"/values", "%v", e)
			}
			values := canonical(resolved)
			doc, _, e := wmdesign.BindTemplates(bundle, "", wmdesign.BoundDocument{Schema: wmdesign.BoundDocumentSchema, Year: p.Document.Year, Slides: []wmdesign.BoundSlide{{ID: slide.ID, Template: slide.Template.ID, ContentKind: slide.ContentKind, Values: values}}})
			if e != nil {
				return c, p.fail(ptr+"/values", "%v", e)
			}
			c.Document.Slides = append(c.Document.Slides, doc.Slides...)
			c.Document.Slides[len(c.Document.Slides)-1].Hidden = slide.Hidden
			c.Document.Slides[len(c.Document.Slides)-1].Notes = slide.Notes
			continue
		}
		t := p.Document.LocalTemplates[slide.Template.ID]
		if slide.Template.Revision != "" && slide.Template.Revision != digest(canonical(t)) {
			return c, p.fail(ptr+"/template/revision", "local template revision must match canonical definition SHA256")
		}
		for k, v := range slide.Values {
			z, ok := t.Zones[k]
			if !ok {
				return c, p.fail(ptr+"/values/"+escape(k), "undeclared content zone")
			}
			if e := validateValue(z.Schema, v, k); e != nil {
				return c, p.fail(ptr+"/values/"+escape(k), "%v", e)
			}
		}
		for k, z := range t.Zones {
			if _, ok := slide.Values[k]; !ok && z.Required {
				return c, p.fail(ptr+"/values", "required content zone %s missing", k)
			}
		}
		frame, e := resolveFrameRef(t.Frame, source)
		if e != nil {
			return c, p.fail("/local_templates/"+escape(slide.Template.ID)+"/frame", "%v", e)
		}
		if t.Grid.ID != "wmds/grid/12-columns" || t.Grid.Revision != "" && t.Grid.Revision != source.Revision {
			return c, fmt.Errorf("local grid requires wmds/grid/12-columns and current source revision")
		}
		if t.FrameOptions != nil {
			q := *t.FrameOptions
			if q.Rail != "" && q.Rail != frame.Rail || q.Footer != "" && q.Footer != frame.Footer {
				return c, fmt.Errorf("frame_options rail/footer must match referenced frame")
			}
			q.Rail = frame.Rail
			q.Footer = frame.Footer
			if q.TitleLines == 0 {
				q.TitleLines = 2
			}
			frame = q
		}
		for k, z := range t.Zones {
			if z.Role == "nav" {
				value, ok := slide.Values[k]
				if !ok {
					continue
				}
				var nav wmdesign.LibraryNavContent
				if e := strictInto(value, &nav); e != nil {
					return c, e
				}
				frame.Nav = nil
				for _, item := range nav.Items {
					frame.Nav = append(frame.Nav, wmdesign.NavTab{ID: item.Key, Label: item.Label})
				}
				frame.Active = nav.Active
			}
		}
		f, e := source.ResolveFrame(frame)
		if e != nil {
			return c, e
		}
		out := wmdesign.SlideSpec{ID: slide.ID, Hidden: slide.Hidden, Notes: slide.Notes, ContentKind: slide.ContentKind, Frame: frame}
		if t.FrameChrome != nil {
			out.LibraryChrome = &wmdesign.LibraryChrome{Emphasis: t.FrameChrome.Emphasis, Whiteboard: t.FrameChrome.Whiteboard, CustomWhiteboard: t.FrameChrome.CustomWhiteboard}
		}
		for k, z := range t.Zones {
			v, ok := slide.Values[k]
			if !ok {
				continue
			}
			switch z.Role {
			case "slide-title":
				s, ok := v.(string)
				if !ok {
					return c, fmt.Errorf("slide-title zone must be string")
				}
				out.Title = s
			case "eyebrow":
				s, ok := v.(string)
				if !ok {
					return c, fmt.Errorf("eyebrow zone must be string")
				}
				out.Eyebrow = s
			case "source":
				s, ok := v.(string)
				if !ok {
					return c, fmt.Errorf("source zone must be string")
				}
				out.Source = s
			}
		}
		if frame.NoHeader && (out.Title != "" || out.Eyebrow != "") {
			return c, fmt.Errorf("no-header frame cannot bind visible title/eyebrow copy")
		}
		var walk func([]Node, string) error
		walk = func(nodes []Node, prefix string) error {
			for _, n := range nodes {
				id := prefix + n.ID
				if n.Kind == "group" {
					if e := walk(n.Nodes, id+"."); e != nil {
						return e
					}
					continue
				}
				b, scope, e := placementRect(*n.Placement, f, source.Tokens.Grid)
				if e != nil {
					return fmt.Errorf("%s: %w", id, e)
				}
				outNode := wmdesign.Node{ID: id, Rect: b, Scope: scope, Kind: n.Kind, Style: n.Style, Ink: n.Ink, Align: n.Align, Surface: n.Surface}
				switch n.Kind {
				case "text":
					v, e := resolveBinding(n.Text, slide.Values)
					if e != nil {
						return e
					}
					text, ok := v.(string)
					if !ok {
						return fmt.Errorf("%s text binding must be string", id)
					}
					outNode.Text = text
				case "box":
					if n.Border != "" {
						return fmt.Errorf("%s: box border token unsupported; use outline surface", id)
					}
				case "image":
					v, e := resolveBinding(n.Asset, slide.Values)
					if e != nil {
						return e
					}
					asset, ok := v.(string)
					if !ok {
						return fmt.Errorf("image asset binding must be string")
					}
					key, ok := assetKeys[asset]
					if !ok {
						return fmt.Errorf("%s: undeclared asset %s", id, asset)
					}
					raw := canonical(map[string]any{"type": "imageframe", "x": b.X, "y": b.Y, "w": b.W, "h": b.H, "photo": key, "focus": "50% 50%", "alt": p.Document.Assets[asset].Description, "fit": n.Fit, "rotate": n.Rotation})
					outNode.Kind = "scene"
					outNode.Scene = &wmdesign.SceneSpec{Node: raw, Path: "/local_templates/" + escape(slide.Template.ID) + "/nodes/" + escape(id)}
				case "rule":
					if n.Weight != .75 || n.Ink != "line" {
						return fmt.Errorf("%s: current rule requires line ink and .75pt weight", id)
					}
					outNode.Rect.H = .75
				case "component", "composite":
					if n.Definition.Revision != "" && n.Definition.Revision != source.Revision {
						return fmt.Errorf("component revision mismatch")
					}
					args, e := resolveArguments(n.Arguments, slide.Values)
					if e != nil {
						return e
					}
					resolveAssetFields(args, assetKeys)
					switch n.Definition.ID {
					case "text.block":
						outNode.Kind = "textblock"
						var s wmdesign.TextBlockSpec
						if e := strictInto(args, &s); e != nil {
							return e
						}
						outNode.TextBlock = &s
					case "card":
						outNode.Kind = "card"
						var s wmdesign.CardSpec
						if e := strictInto(args, &s); e != nil {
							return e
						}
						outNode.Card = &s
					case "data.metric":
						outNode.Kind = "metric"
						var s wmdesign.DataMetricSpec
						if e := strictInto(args, &s); e != nil {
							return e
						}
						outNode.DataMetric = &s
					case "richtext":
						outNode.Kind = "richtext"
						outNode.Style = "body"
						var s wmdesign.RichTextSpec
						if e := strictInto(args, &s); e != nil {
							return e
						}
						outNode.RichText = &s
					default:
						kind := strings.TrimPrefix(strings.TrimPrefix(n.Definition.ID, "wmds/component/"), "wmds/composite/")
						if kind == n.Definition.ID {
							return fmt.Errorf("unsupported local definition %s; use wmds/component/<source-type> or wmds/composite/<source-type>", n.Definition.ID)
						}
						for _, reserved := range []string{"type", "id", "x", "y", "w", "h"} {
							if _, exists := args[reserved]; exists {
								return fmt.Errorf("scene arguments cannot override reserved %s", reserved)
							}
						}
						raw, e := wmdesign.ComposeSceneNode(kind, args, b)
						if e != nil {
							return e
						}
						path := "/local_templates/" + escape(slide.Template.ID) + "/nodes/" + escape(id)
						keys := map[string][]string{}
						for field, items := range n.Keys {
							v, e := lookupPointer(args, "/"+strings.TrimPrefix(field, "/"))
							if e != nil {
								return e
							}
							array, ok := v.([]any)
							if !ok || len(array) != len(items) {
								return fmt.Errorf("%s: keys/%s must match array length", id, field)
							}
							seen := map[string]bool{}
							for _, key := range items {
								if !stableID.MatchString(key) || seen[key] {
									return fmt.Errorf("invalid/duplicate item key %s", key)
								}
								seen[key] = true
							}
							keys[path+"/"+strings.TrimPrefix(field, "/")] = items
						}
						outNode.Kind = "scene"
						outNode.Scene = &wmdesign.SceneSpec{Node: raw, Path: path, Keys: keys, Allocation: &b}
					}
				}
				out.Nodes = append(out.Nodes, outNode)
			}
			return nil
		}
		if e = walk(t.Nodes, ""); e != nil {
			return c, p.fail(ptr+"/values", "%v", e)
		}
		c.Document.Slides = append(c.Document.Slides, out)
	}
	return c, nil
}
func resolveBinding(v any, values map[string]any) (any, error) {
	if m, ok := v.(map[string]any); ok {
		if len(m) != 1 {
			return nil, fmt.Errorf("binding object must contain only binding")
		}
		key, ok := m["binding"].(string)
		if !ok {
			return nil, fmt.Errorf("binding requires string")
		}
		val, ok := values[key]
		if !ok {
			return nil, fmt.Errorf("binding %s has no value", key)
		}
		return val, nil
	}
	return v, nil
}
func resolveArguments(v map[string]any, values map[string]any) (map[string]any, error) {
	var walk func(any) (any, error)
	walk = func(x any) (any, error) {
		switch x := x.(type) {
		case map[string]any:
			if _, ok := x["binding"]; ok {
				return resolveBinding(x, values)
			}
			m := map[string]any{}
			for k, v := range x {
				w, e := walk(v)
				if e != nil {
					return nil, e
				}
				m[k] = w
			}
			return m, nil
		case []any:
			a := []any{}
			for _, v := range x {
				w, e := walk(v)
				if e != nil {
					return nil, e
				}
				a = append(a, w)
			}
			return a, nil
		default:
			return x, nil
		}
	}
	x, e := walk(v)
	if e != nil {
		return nil, e
	}
	return x.(map[string]any), nil
}
func resolveFrameRef(ref Reference, s *wmdesign.Source) (wmdesign.FrameRequest, error) {
	if ref.Revision != "" && ref.Revision != s.Revision {
		return wmdesign.FrameRequest{}, fmt.Errorf("frame revision mismatch")
	}
	id := strings.TrimPrefix(ref.ID, "wmds/frame/")
	parts := strings.Split(id, "-")
	if !strings.HasPrefix(ref.ID, "wmds/frame/") || len(parts) != 2 || (parts[0] != "none" && parts[0] != "left" && parts[0] != "right" && parts[0] != "nav") || (parts[1] != "compact" && parts[1] != "tall" && parts[1] != "slim") {
		return wmdesign.FrameRequest{}, fmt.Errorf("unsupported frame ID; use wmds/frame/{none,left,right,nav}-{compact,tall,slim}")
	}
	if _, exists := s.Frames.Footers[parts[1]]; !exists {
		return wmdesign.FrameRequest{}, fmt.Errorf("frame footer %s is not declared by source revision %s", parts[1], s.Revision)
	}
	return wmdesign.FrameRequest{Rail: parts[0], Footer: parts[1], Surface: "light", TitleLines: 2}, nil
}
func placementRect(p Placement, f wmdesign.ResolvedFrame, g wmdesign.Grid) (wmdesign.Rect, string, error) {
	var z wmdesign.Rect
	scope := "body"
	switch p.Zone {
	case "body":
		z = f.Body
	case "rail":
		z = f.Rail
		scope = "rail"
	case "short_body":
		z = f.ShortBody
		scope = "short"
	case "tall_body":
		z = f.TallBody
		scope = "tall"
	default:
		return z, "", fmt.Errorf("unsupported local placement zone %s (body/rail/short_body/tall_body implemented)", p.Zone)
	}
	if z.W <= 0 || z.H <= 0 {
		return z, "", fmt.Errorf("frame zone %s unavailable", p.Zone)
	}
	var b wmdesign.Rect
	if p.Rect != nil {
		b = *p.Rect
		b.X += z.X
		b.Y += z.Y
	} else {
		r, e := g.Span(p.Span.Start, p.Span.Count)
		if e != nil {
			return b, "", e
		}
		b = wmdesign.Rect{X: z.X + r.X - g.MarginX, Y: z.Y + p.Span.Y, W: r.W, H: p.Span.H}
	}
	if b.W <= 0 || b.H <= 0 || b.X < z.X-.001 || b.Y < z.Y-.001 || b.X+b.W > z.X+z.W+.001 || b.Y+b.H > z.Y+z.H+.001 {
		return b, "", fmt.Errorf("local placement exceeds frame zone")
	}
	return b, scope, nil
}

// BundlePath maps shorthand using release root when set, otherwise the working tree.
func BundlePath(value string) string {
	if value == "v1" || value == "v2" || value == "v3" || value == "v4" || value == "v5" {
		return filepath.Join(os.Getenv("PPTXGENGO_RELEASE_ROOT"), "library", "wm-design-system", value)
	}
	return value
}

func lookupPointer(v any, path string) (any, error) {
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		key := strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch node := v.(type) {
		case map[string]any:
			next, ok := node[key]
			if !ok {
				return nil, fmt.Errorf("unknown key array %s", path)
			}
			v = next
		case []any:
			i, e := strconv.Atoi(key)
			if e != nil || i < 0 || i >= len(node) {
				return nil, fmt.Errorf("invalid key array pointer %s", path)
			}
			v = node[i]
		default:
			return nil, fmt.Errorf("invalid key array pointer %s", path)
		}
	}
	return v, nil
}

func resolveAssetFields(v any, keys map[string]string) {
	switch m := v.(type) {
	case map[string]any:
		for k, v := range m {
			if k == "photo" || k == "src" {
				if id, ok := v.(string); ok {
					if key, ok := keys[id]; ok {
						m[k] = key
					}
				}
			}
			resolveAssetFields(v, keys)
		}
	case []any:
		for _, v := range m {
			resolveAssetFields(v, keys)
		}
	}
}
