package deckproject

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

const DiagramPatchSchema = "pptxgengo.diagram-patch.v1"

type DiagramOperation struct {
	Action        string          `json:"action"`
	ID            string          `json:"id"`
	Node          *Node           `json:"node,omitempty"`
	Rect          *wmdesign.Rect  `json:"rect,omitempty"`
	DX            *float64        `json:"dx_pt,omitempty"`
	DY            *float64        `json:"dy_pt,omitempty"`
	Geometry      *NativeGeometry `json:"geometry,omitempty"`
	IncidentEdges string          `json:"incident_edges,omitempty"`
}
type DiagramPatch struct {
	Schema     string             `json:"schema"`
	Actor      string             `json:"actor"`
	Reason     string             `json:"reason"`
	Operations []DiagramOperation `json:"operations"`
}
type DiagramInspection struct {
	Containment         []wmdesign.DiagramContainmentObservation `json:"containment,omitempty"`
	Overlaps            []DiagramOverlap                         `json:"overlaps,omitempty"`
	SourceGeometryBasis map[string]string                        `json:"source_geometry_basis"`
	Ports               []DiagramPort                            `json:"ports,omitempty"`
	Connections         []DiagramConnection                      `json:"connections,omitempty"`
	Schema              string                                   `json:"schema"`
	SlideID             string                                   `json:"slide_id"`
	TemplateID          string                                   `json:"template_id"`
	SourceSHA256        string                                   `json:"source_sha256"`
	Frame               wmdesign.ResolvedFrame                   `json:"frame"`
	Nodes               []Node                                   `json:"nodes"`
	NativeGeometry      map[string]NativeGeometry                `json:"native_geometry,omitempty"`
	NativeOrder         map[string][]string                      `json:"native_order,omitempty"`
	MeasuredScenes      []wmdesign.SceneRecord                   `json:"measured_scenes"`
	MeasuredText        []wmdesign.TextRecord                    `json:"measured_text"`
	FinalNative         []NativeGeometryObservation              `json:"final_native_geometry"`
	Warnings            []string                                 `json:"warnings"`
	Validation          string                                   `json:"validation"`
}
type DiagramPatchResult struct {
	Applied      bool              `json:"applied"`
	BeforeSHA256 string            `json:"before_sha256"`
	AfterSHA256  string            `json:"after_sha256"`
	Decision     string            `json:"decision,omitempty"`
	Patch        DiagramPatch      `json:"patch"`
	Inspection   DiagramInspection `json:"inspection"`
}

func DecodeDiagramPatch(raw []byte, file string) (DiagramPatch, error) {
	var out DiagramPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("diagram patch exceeds 1 MiB")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty diagram patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("diagram patch requires one document")
	}
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
	if out.Schema != DiagramPatchSchema || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || len(out.Actor) > 256 || len(out.Reason) > 4096 || len(out.Operations) == 0 || len(out.Operations) > 500 {
		return out, fmt.Errorf("diagram patch requires schema, actor, reason and 1..500 operations")
	}
	return out, nil
}
func diagramSlide(p *Project, id string) (int, LocalTemplate, error) {
	for i, s := range p.Document.Slides {
		if s.ID == id {
			if s.Template.Scope != "local" {
				return 0, LocalTemplate{}, fmt.Errorf("diagram editing requires project detach first; shared catalog is immutable")
			}
			return i, p.Document.LocalTemplates[s.Template.ID], nil
		}
	}
	return 0, LocalTemplate{}, fmt.Errorf("unknown slide ID %s", id)
}
func InspectDiagram(p *Project, id, bundle, engine string) (DiagramInspection, error) {
	out := DiagramInspection{Schema: "pptxgengo.diagram-inspection.v1", SlideID: id, SourceSHA256: p.SourceHash(), Warnings: []string{}, Validation: "renderer_measured_frame_checked_desktop_review_pending"}
	idx, t, e := diagramSlide(p, id)
	if e != nil {
		return out, e
	}
	s := p.Document.Slides[idx]
	out.TemplateID = s.Template.ID
	out.Nodes = t.Nodes
	out.NativeGeometry = s.NativeGeometry
	out.NativeOrder = s.NativeOrder
	c, e := Check(p, bundle, engine)
	if e != nil {
		return out, e
	}
	c.Document.Slides = []wmdesign.SlideSpec{c.Document.Slides[idx]}
	c.Document.Sections = nil
	raw, report, e := wmdesign.BuildWithEngineAndAssets(bundle, "", c.Document, engine, c.Assets)
	if e != nil {
		return out, e
	}
	sourcePkg, e := openLineagePackage(raw)
	if e != nil {
		return out, e
	}
	sourcePart, e := sourcePkg.read("ppt/slides/slide1.xml")
	if e != nil {
		return out, e
	}
	sourceObjects, _, e := geometryInventory(sourcePart)
	if e != nil {
		return out, e
	}
	out.SourceGeometryBasis = map[string]string{}
	for name := range sourceObjects {
		out.SourceGeometryBasis[name] = nativeGeometryBasis(sourceObjects, name)
	}
	raw, e = applyNativeGeometry(raw, c.Document, &report)
	if e != nil {
		return out, e
	}
	measured := report.Slides[0]
	out.Frame = measured.Frame
	out.MeasuredScenes = measured.Scenes
	out.MeasuredText = measured.Texts
	pkg, e := openLineagePackage(raw)
	if e != nil {
		return out, e
	}
	data, e := pkg.read("ppt/slides/slide1.xml")
	if e != nil {
		return out, e
	}
	objects, _, e := geometryInventory(data)
	if e != nil {
		return out, e
	}
	world, e := geometryWorld(objects)
	if e != nil {
		return out, e
	}
	names := []string{}
	for n := range objects {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		out.FinalNative = append(out.FinalNative, NativeGeometryObservation{n, objects[n].geometry, world[n]})
	}
	out.Containment, out.Overlaps, e = checkDiagramContainment(objects, s.DiagramContainment)
	if e != nil {
		return out, e
	}
	for _, overlap := range out.Overlaps {
		out.Warnings = append(out.Warnings, fmt.Sprintf("Allocation envelopes of declared siblings %s and %s overlap inside %s; review the diagram.", overlap.A, overlap.B, overlap.Container))
	}
	// Text remains measured in authored coordinates. Native group scaling is
	// reported separately instead of claiming that PowerPoint reflow was tested.
	if len(s.NativeGeometry) > 0 {
		out.Warnings = append(out.Warnings, "Native transforms preserve object geometry; measured text is in authored coordinates. Review scaled typography in PowerPoint.")
	}
	if e = addDiagramPorts(&out, c.Document); e != nil {
		return out, e
	}
	out.Warnings = append(out.Warnings, report.Warnings...)
	return out, nil
}
func findDiagramNode(nodes *[]Node, id string) (*[]Node, int) {
	for i := range *nodes {
		n := &(*nodes)[i]
		if n.ID == id {
			return nodes, i
		}
		if p, j := findDiagramNode(&n.Nodes, id); p != nil {
			return p, j
		}
	}
	return nil, -1
}
func diagramEndpoint(n Node, target string) bool {
	if n.Definition == nil || n.Definition.ID != "wmds/component/attached-connector" {
		return false
	}
	for _, k := range []string{"from", "to"} {
		if m, ok := n.Arguments[k].(map[string]any); ok && m["node"] == target {
			return true
		}
	}
	return false
}
func removeIncident(nodes *[]Node, id string, remove bool) error {
	for i := 0; i < len(*nodes); {
		n := &(*nodes)[i]
		if diagramEndpoint(*n, id) {
			if !remove {
				return fmt.Errorf("node %s has incident edge %s; specify incident_edges: remove", id, n.ID)
			}
			*nodes = append((*nodes)[:i], (*nodes)[i+1:]...)
			continue
		}
		if e := removeIncident(&n.Nodes, id, remove); e != nil {
			return e
		}
		i++
	}
	return nil
}

// PatchDiagram previews by default, validates the complete result, and commits
// through the existing guarded transaction and exact predecessor retention.
func PatchDiagram(p *Project, slideID string, patch DiagramPatch, bundle, engine string, apply bool) (DiagramPatchResult, error) {
	out := DiagramPatchResult{BeforeSHA256: p.SourceHash(), Patch: patch}
	checked, e := DecodeDiagramPatch(canonical(patch), "diagram-patch")
	if e != nil {
		return out, e
	}
	patch = checked
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	s := p.Document.Slides[idx]
	for _, other := range p.Document.Slides {
		if other.ID != slideID && other.Template.Scope == "local" && other.Template.ID == s.Template.ID {
			return out, fmt.Errorf("local template is used by multiple slides; project fork it for this slide before editing")
		}
	}
	if s.Template.Revision != "" {
		return out, fmt.Errorf("local template revision is pinned; explicitly revise the local reference before changing its definition")
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return out, e
	}
	t = clone
	main, e := sourceYAML(p.Raw)
	if e != nil {
		return out, e
	}
	slides, documents, e := authoredSlides(p, main)
	if e != nil {
		return out, e
	}
	templateNode := mappingNode(mappingNode(main.Content[0], "local_templates"), s.Template.ID)
	file := p.TemplateFiles[s.Template.ID]
	if file != "" {
		doc, e := sourceYAML(p.SourceFiles[file])
		if e != nil {
			return out, e
		}
		documents[file] = doc
		templateNode = doc.Content[0]
	} else {
		file = filepath.Base(p.SourcePath)
	}
	nativeLayoutActive := len(s.NativeGeometry) > 0 || len(s.NativeOrder) > 0
	nodesChanged := false
	slideChanged := false
	for _, op := range patch.Operations {
		if (op.Action != "transform" && !stableID.MatchString(op.ID)) || op.ID == "" || len(op.ID) > 512 {
			return out, fmt.Errorf("diagram operation requires a stable ID")
		}
		list, j := findDiagramNode(&t.Nodes, op.ID)
		if nativeLayoutActive && op.Action != "transform" && op.Action != "reset_native_layout" {
			return out, fmt.Errorf("source placement/topology edits require an explicit reset_native_layout operation first; exact predecessors are retained")
		}
		switch op.Action {
		case "reset_native_layout":
			if op.ID != slideID || op.Node != nil || op.Geometry != nil || op.Rect != nil || op.DX != nil || op.DY != nil || op.IncidentEdges != "" {
				return out, fmt.Errorf("reset_native_layout requires the slide ID only")
			}
			removeMappingField(slides[slideID], "native_geometry")
			removeMappingField(slides[slideID], "native_order")
			if len(s.DiagramContainment) == 0 {
				removeMappingField(slides[slideID], "native_geometry_template")
			}
			nativeLayoutActive = false
			slideChanged = true
		case "add":
			if list != nil || op.Node == nil || op.Node.ID != op.ID || op.Rect != nil || op.Geometry != nil || op.DX != nil || op.DY != nil || op.IncidentEdges != "" {
				return out, fmt.Errorf("add requires a new node with matching ID only")
			}
			t.Nodes = append(t.Nodes, *op.Node)
			nodesChanged = true
		case "remove":
			if list == nil || op.Node != nil || op.Rect != nil || op.Geometry != nil || op.DX != nil || op.DY != nil || (op.IncidentEdges != "" && op.IncidentEdges != "remove") {
				return out, fmt.Errorf("invalid remove operation")
			}
			if e = removeIncident(&t.Nodes, op.ID, op.IncidentEdges == "remove"); e != nil {
				return out, e
			}
			list, j = findDiagramNode(&t.Nodes, op.ID)
			*list = append((*list)[:j], (*list)[j+1:]...)
			nodesChanged = true
		case "move", "resize":
			if list == nil || op.Node != nil || op.Geometry != nil || op.IncidentEdges != "" {
				return out, fmt.Errorf("move/resize requires existing node placement")
			}
			n := &(*list)[j]
			if n.Placement == nil || n.Placement.Rect == nil {
				return out, fmt.Errorf("move/resize requires rect placement")
			}
			if op.Rect != nil {
				if op.DX != nil || op.DY != nil {
					return out, fmt.Errorf("rect and deltas are mutually exclusive")
				}
				r := *op.Rect
				n.Placement.Rect = &r
			} else {
				if op.Action != "move" || op.DX == nil && op.DY == nil {
					return out, fmt.Errorf("resize requires rect; move requires rect or deltas")
				}
				if op.DX != nil {
					n.Placement.Rect.X += *op.DX
				}
				if op.DY != nil {
					n.Placement.Rect.Y += *op.DY
				}
			}
			nodesChanged = true
		case "transform":
			if op.Geometry == nil || op.Node != nil || op.Rect != nil || op.DX != nil || op.DY != nil || op.IncidentEdges != "" {
				return out, fmt.Errorf("transform requires native geometry only")
			}
			if e = validateNativeGeometry(*op.Geometry); e != nil {
				return out, e
			}
			m := mappingNode(slides[slideID], "native_geometry")
			if m == nil {
				m = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				replaceMappingField(slides[slideID], "native_geometry", m)
			}
			original, e := InspectDiagram(p, slideID, bundle, engine)
			if e != nil {
				return out, e
			}
			g := *op.Geometry
			basis := original.SourceGeometryBasis[op.ID]
			if basis == "" {
				return out, fmt.Errorf("transform source object missing")
			}
			if g.SourceGeometrySHA256 != "" && g.SourceGeometrySHA256 != basis {
				return out, fmt.Errorf("transform source geometry basis mismatch")
			}
			g.SourceGeometrySHA256 = basis
			v, e := editYAMLNode(g)
			if e != nil {
				return out, e
			}
			replaceMappingField(m, op.ID, v)
			pin, e := editYAMLNode(s.Template)
			if e != nil {
				return out, e
			}
			replaceMappingField(slides[slideID], "native_geometry_template", pin)
			nativeLayoutActive = true
			slideChanged = true
		default:
			return out, fmt.Errorf("unknown diagram action %s", op.Action)
		}
	}
	changes := map[string][]byte{}
	if nodesChanged {
		if pruneRemovedNodeContainment(p.Document.LocalTemplates[s.Template.ID].Nodes, t.Nodes, slides[slideID]) {
			slideChanged = true
		}
		pruned, err := updateDiagramNodes(t, s, templateNode, slides[slideID], false)
		if err != nil {
			return out, err
		}
		slideChanged = slideChanged || pruned
		raw, e := encodeSourceYAML(documents[file])
		if e != nil {
			return out, e
		}
		changes[file] = raw
	}
	if slideChanged {
		sf := p.SlideFiles[slideID]
		if sf == "" {
			sf = filepath.Base(p.SourcePath)
		}
		raw, e := encodeSourceYAML(documents[sf])
		if e != nil {
			return out, e
		}
		changes[sf] = raw
	}
	candidate, e := loadProject(p.SourcePath, mergeTextOverrides(p.SourceFiles, changes))
	if e != nil {
		return out, e
	}
	out.AfterSHA256 = candidate.SourceHash()
	out.Inspection, e = InspectDiagram(candidate, slideID, bundle, engine)
	if e != nil {
		return out, e
	}
	if !apply {
		return out, nil
	}
	out.Applied = true
	out.Decision = "decisions/diagram-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	changes[out.Decision] = canonical(out)
	_, e = commitSourceChanges(p, changes, func(c *Project) error { _, err := InspectDiagram(c, slideID, bundle, engine); return err })
	return out, e
}
func yamlNodeMatchesNode(old *yaml.Node, n Node) (bool, error) {
	p := &Project{Positions: map[string]Position{}}
	v, e := p.yamlValue(old, "", 0)
	if e != nil {
		return false, e
	}
	return bytes.Equal(canonical(v), canonical(n)), nil
}
func preserveDiagramComments(old, current *yaml.Node) {
	if old == nil || current == nil {
		return
	}
	current.HeadComment, current.LineComment, current.FootComment = old.HeadComment, old.LineComment, old.FootComment
	if old.Kind == yaml.MappingNode && current.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(current.Content); i += 2 {
			if v := mappingNode(old, current.Content[i].Value); v != nil {
				preserveDiagramComments(v, current.Content[i+1])
			}
		}
	}
}

// updateDiagramNodes writes a typed local definition and prunes only unused
// non-chrome bindings. Callers share one authored AST and one guarded commit.
func updateDiagramNodes(t LocalTemplate, s Slide, templateNode, slideNode *yaml.Node, forceBindings bool) (bool, error) {
	slideChanged := false
	used := map[string]bool{}
	var collect func([]Node)
	collect = func(nodes []Node) {
		for _, n := range nodes {
			for _, v := range []any{n.Text, n.Asset, n.Arguments} {
				collectBindings(v, func(key string) { used[key] = true })
			}
			collect(n.Nodes)
		}
	}
	collect(t.Nodes)
	removed := false
	values := map[string]any{}
	for key, v := range s.Values {
		values[key] = v
	}
	for key, z := range t.Zones {
		if z.Role != "slide-title" && z.Role != "eyebrow" && z.Role != "source" && z.Role != "nav" && !used[key] {
			delete(t.Zones, key)
			delete(values, key)
			removed = true
		}
	}
	if removed || forceBindings {
		zones, e := editYAMLNode(t.Zones)
		if e != nil {
			return false, e
		}
		preserveDiagramComments(mappingNode(templateNode, "zones"), zones)
		replaceMappingField(templateNode, "zones", zones)
		node := slideNode
		if mappingNode(node, "content") != nil {
			content, bindings, e := localAuthoredContent(t, values)
			if e != nil {
				return false, e
			}
			cn, e := editYAMLNode(content)
			if e != nil {
				return false, e
			}
			bn, e := editYAMLNode(bindings)
			if e != nil {
				return false, e
			}
			preserveDiagramComments(mappingNode(node, "content"), cn)
			preserveDiagramComments(mappingNode(node, "bindings"), bn)
			replaceMappingField(node, "content", cn)
			replaceMappingField(node, "bindings", bn)
			removeMappingField(node, "values")
		} else {
			vn, e := editYAMLNode(values)
			if e != nil {
				return false, e
			}
			preserveDiagramComments(mappingNode(node, "values"), vn)
			replaceMappingField(node, "values", vn)
		}
		slideChanged = true
	}
	// Preserve comments on untouched source nodes, including their bindings.
	originalNodes := mappingNode(templateNode, "nodes")
	byID := map[string]*yaml.Node{}
	if originalNodes != nil {
		for _, n := range originalNodes.Content {
			if id := mappingNode(n, "id"); id != nil {
				byID[id.Value] = n
			}
		}
	}
	replacement := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, n := range t.Nodes {
		v, e := editYAMLNode(n)
		if e != nil {
			return false, e
		}
		if old := byID[n.ID]; old != nil {
			if same, err := yamlNodeMatchesNode(old, n); err == nil && same {
				v = old
			} else {
				preserveDiagramComments(old, v)
			}
		}
		replacement.Content = append(replacement.Content, v)
	}
	replaceMappingField(templateNode, "nodes", replacement)
	return slideChanged, nil
}

// Native names use the source node namespace plus component part suffixes.
// Prune removed members only. Surviving children of a removed container need an
// explicit membership decision and will fail final validation until revised.
func pruneRemovedNodeContainment(before, after []Node, slide *yaml.Node) bool {
	var names func([]Node, string, map[string]bool)
	names = func(nodes []Node, prefix string, out map[string]bool) {
		for _, n := range nodes {
			id := prefix + n.ID
			out[id] = true
			names(n.Nodes, id+".", out)
		}
	}
	old, new := map[string]bool{}, map[string]bool{}
	names(before, "", old)
	names(after, "", new)
	m := mappingNode(slide, "diagram_containment")
	if m == nil {
		return false
	}
	changed := false
	for i := 0; i+1 < len(m.Content); {
		name := m.Content[i].Value
		removed := false
		for id := range old {
			if !new[id] && (name == id || strings.HasPrefix(name, id+".")) {
				removed = true
				break
			}
		}
		if removed {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			changed = true
		} else {
			i += 2
		}
	}
	if len(m.Content) == 0 {
		removeMappingField(slide, "diagram_containment")
		if mappingNode(slide, "native_geometry") == nil && mappingNode(slide, "native_order") == nil {
			removeMappingField(slide, "native_geometry_template")
		}
	}
	return changed
}
