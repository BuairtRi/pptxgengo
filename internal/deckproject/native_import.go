package deckproject

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

const NativeImportSchema = "pptxgengo.native-import-map.v1"
const NativeImportMaxBytes = 32 << 20

type NativeImportObject struct {
	Part      string         `json:"part"`
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Geometry  NativeGeometry `json:"geometry"`
	Text      string         `json:"text,omitempty"`
	Supported bool           `json:"supported"`
	Reasons   []string       `json:"reasons,omitempty"`
}
type NativeImportInventory struct {
	Schema           string               `json:"schema"`
	PPTXSHA256       string               `json:"pptx_sha256"`
	FormattingPolicy string               `json:"formatting_policy"`
	Objects          []NativeImportObject `json:"objects"`
}
type NativeImportSelection struct {
	Part    string         `json:"part"`
	Name    string         `json:"name"`
	NodeID  string         `json:"node_id"`
	Kind    string         `json:"kind"` // text or editable-block
	Rect    *wmdesign.Rect `json:"rect,omitempty"`
	Style   string         `json:"style"`
	Surface string         `json:"surface,omitempty"`
	Align   string         `json:"align"`
}
type NativeImportMap struct {
	Schema               string                  `json:"schema"`
	ExpectedSourceSHA256 string                  `json:"expected_source_sha256"`
	ExpectedPPTXSHA256   string                  `json:"expected_pptx_sha256"`
	Actor                string                  `json:"actor"`
	Reason               string                  `json:"reason"`
	FormattingPolicy     string                  `json:"formatting_policy"`
	Selections           []NativeImportSelection `json:"selections"`
}
type nativeImportEvidence struct {
	SourceAsset      string               `json:"source_asset"`
	SourcePPTXSHA256 string               `json:"source_pptx_sha256"`
	Mapping          NativeImportMap      `json:"mapping"`
	SelectedObjects  []NativeImportObject `json:"selected_objects"`
}

// InspectNativeImport reports addresses, never guesses source identities or style.
// Parts are sorted by address, not presented as slide order.
func InspectNativeImport(data []byte) (NativeImportInventory, error) {
	out := NativeImportInventory{Schema: "pptxgengo.native-import-inventory.v1", PPTXSHA256: digest(data), FormattingPolicy: "reauthor_with_design_style", Objects: []NativeImportObject{}}
	if len(data) > NativeImportMaxBytes {
		return out, fmt.Errorf("native import exceeds 32 MiB")
	}
	pkg, e := openLineagePackage(data)
	if e != nil {
		return out, e
	}
	parts := []string{}
	for name := range pkg.files {
		if strings.HasPrefix(name, "ppt/slides/slide") && strings.HasSuffix(name, ".xml") && !strings.Contains(strings.TrimPrefix(name, "ppt/slides/"), "/") {
			parts = append(parts, name)
		}
	}
	sort.Strings(parts)
	if len(parts) == 0 {
		return out, fmt.Errorf("native import has no slide parts")
	}
	for _, part := range parts {
		tree, e := pkg.tree(part)
		if e != nil {
			return out, e
		}
		var roots []*xmlNode
		var walk func(*xmlNode)
		walk = func(n *xmlNode) {
			if nativeObjectKind(n) {
				roots = append(roots, n)
				return
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(tree)
		names := map[string]int{}
		for _, n := range roots {
			names[geometryName(n)]++
		}
		for _, n := range roots {
			o := NativeImportObject{Part: part, Name: geometryName(n), Kind: n.Name.Local}
			fail := func(reason string) { o.Reasons = append(o.Reasons, reason) }
			if o.Name == "" || names[o.Name] != 1 {
				fail("missing_or_ambiguous_top_level_name")
			}
			if n.Name.Local != "sp" {
				fail("requires_top_level_rectangular_text_shape")
			} else {
				g, e := readNativeGeometry(n, "")
				o.Geometry = g
				if e != nil {
					fail("explicit_transform_required")
				} else if g.Rotation != 0 || g.FlipH || g.FlipV || g.W <= 0 || g.H <= 0 {
					fail("rotation_flip_or_empty_extent")
				}
				props := lineageChild(n, lineagePML, "spPr")
				var geom *xmlNode
				if props != nil {
					geom = lineageChild(props, drawingML, "prstGeom")
				}
				if geom == nil || lineageAttr(geom, "", "prst") != "rect" {
					fail("rectangle_geometry_required")
				}
				if geom != nil {
					for _, c := range geom.Children {
						if len(c.Children) > 0 {
							fail("adjusted_geometry_unsupported")
						}
					}
				}
				paragraphs := nativeParagraphs(n)
				o.Text = nativeParagraphText(paragraphs)
				if strings.TrimSpace(o.Text) == "" || len(o.Text) > 32768 {
					fail("nonempty_plain_text_required")
				}
				for _, p := range paragraphs {
					if len(p.ReviewItems) > 0 || len(p.Runs) > 1 {
						fail("rich_text_fields_or_bullets_unsupported")
						break
					}
					for _, run := range p.Runs {
						if run.Kind != "r" {
							fail("plain_runs_required")
						}
					}
				}
				// Do not discard charts, hyperlinks, tags or relationship-bound payloads.
				var check func(*xmlNode)
				check = func(x *xmlNode) {
					for _, a := range x.Attrs {
						if a.Name.Space == "http://schemas.openxmlformats.org/officeDocument/2006/relationships" {
							fail("relationship_bearing_shape_unsupported")
						}
					}
					if x.Name.Space == drawingML && (x.Name.Local == "hlinkClick" || x.Name.Local == "hlinkMouseOver" || x.Name.Local == "blipFill" || x.Name.Local == "custGeom") {
						fail("interactive_or_complex_payload_unsupported")
					}
					for _, c := range x.Children {
						check(c)
					}
				}
				check(n)
			}
			o.Supported = len(o.Reasons) == 0
			out.Objects = append(out.Objects, o)
		}
	}
	return out, nil
}
func DecodeNativeImportMap(raw []byte, file string) (NativeImportMap, error) {
	var out NativeImportMap
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("native import map exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty native import map")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("native import map requires one document")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = p.shapeType(v, reflect.TypeOf(out), ""); e != nil {
		return out, e
	}
	if e = strictInto(v, &out); e != nil {
		return out, e
	}
	validHash := func(s string) bool { _, e := hex.DecodeString(s); return len(s) == 64 && e == nil }
	if out.Schema != NativeImportSchema || !validHash(out.ExpectedSourceSHA256) || !validHash(out.ExpectedPPTXSHA256) || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || out.FormattingPolicy != "reauthor_with_design_style" || len(out.Selections) == 0 || len(out.Selections) > 100 {
		return out, fmt.Errorf("native import requires schema, both SHA256 guards, actor, reason, explicit reauthor_with_design_style policy and 1..100 selections")
	}
	return out, nil
}

// ImportNative creates independent local source nodes. This explicitly replaces
// native formatting; the exact source package and map remain private evidence.
func ImportNative(p *Project, slide string, data []byte, mapping NativeImportMap, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeNativeImportMap(canonical(mapping), "native-import-map")
	if e != nil {
		return empty, e
	}
	mapping = checked
	if mapping.ExpectedSourceSHA256 != p.SourceHash() || mapping.ExpectedPPTXSHA256 != digest(data) {
		return empty, fmt.Errorf("native import source/PPTX hash mismatch")
	}
	inventory, e := InspectNativeImport(data)
	if e != nil {
		return empty, e
	}
	_, original, e := diagramSlide(p, slide)
	if e != nil {
		return empty, e
	}
	var candidate LocalTemplate
	if e = strictInto(original, &candidate); e != nil {
		return empty, e
	}
	destination, e := InspectDiagram(p, slide, bundle, engine)
	if e != nil {
		return empty, e
	}
	selected := []NativeImportObject{}
	seen := map[string]bool{}
	for _, s := range mapping.Selections {
		if !stableID.MatchString(s.NodeID) || s.Style == "" || (s.Align != "left" && s.Align != "center" && s.Align != "right") {
			return empty, fmt.Errorf("native import requires stable node_id, design style and explicit alignment")
		}
		if list, _ := findDiagramNode(&candidate.Nodes, s.NodeID); list != nil {
			return empty, fmt.Errorf("native import node ID already exists: %s", s.NodeID)
		}
		key := s.Part + "\x00" + s.Name
		if seen[key] {
			return empty, fmt.Errorf("duplicate native import selection")
		}
		seen[key] = true
		matches := []NativeImportObject{}
		for _, o := range inventory.Objects {
			if o.Part == s.Part && o.Name == s.Name {
				matches = append(matches, o)
			}
		}
		if len(matches) != 1 || !matches[0].Supported {
			return empty, fmt.Errorf("native import selection unsupported or ambiguous: %s / %s", s.Part, s.Name)
		}
		o := matches[0]
		selected = append(selected, o)
		// Inventory coordinates are slide-absolute; authored placement is body-relative.
		rect := wmdesign.Rect{X: o.Geometry.X - destination.Frame.Body.X, Y: o.Geometry.Y - destination.Frame.Body.Y, W: o.Geometry.W, H: o.Geometry.H}
		if s.Rect != nil {
			rect = *s.Rect
		}
		if rect.W <= 0 || rect.H <= 0 {
			return empty, fmt.Errorf("native import requires positive extent")
		}
		n := Node{ID: s.NodeID, Placement: &Placement{Zone: "body", Rect: &rect}, Style: s.Style, Align: s.Align}
		switch s.Kind {
		case "text":
			if s.Surface != "" {
				return empty, fmt.Errorf("text import cannot declare surface")
			}
			n.Kind = "text"
			n.Text = o.Text
		case "editable-block":
			if s.Surface == "" {
				return empty, fmt.Errorf("editable-block import requires explicit design surface")
			}
			n.Kind = "component"
			n.Style = ""
			n.Align = ""
			n.Definition = &Reference{Scope: "shared", ID: "wmds/component/editable-block"}
			n.Arguments = map[string]any{"text": o.Text, "surface": s.Surface, "style": s.Style, "align": s.Align}
		default:
			return empty, fmt.Errorf("native import kind must be text or editable-block")
		}
		candidate.Nodes = append(candidate.Nodes, n)
	}
	asset := "assets/objects/sha256/" + digest(data)
	evidence := nativeImportEvidence{SourceAsset: asset, SourcePPTXSHA256: digest(data), Mapping: mapping, SelectedObjects: selected}
	return compositionCandidate(p, slide, "native-import", mapping.Actor, mapping.Reason, candidate, bundle, engine, apply, evidence, map[string][]byte{asset: data})
}
