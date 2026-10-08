package deckproject

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

// Deletion is a proposal, never an inference applied without review. This first
// structural contract covers whole, exclusively owned top-level local leaves.
type StructureReconciliationField struct {
	SourceNode      string                    `json:"source_node,omitempty"`
	NewNode         *Node                     `json:"new_node,omitempty"`
	Zones           map[string]Zone           `json:"zones,omitempty"`
	Values          map[string]any            `json:"values,omitempty"`
	NativeGeometry  map[string]NativeGeometry `json:"native_geometry,omitempty"`
	EditedRootOrder []string                  `json:"edited_root_order,omitempty"`

	ID             string   `json:"id"`
	SlideID        string   `json:"slide_id"`
	NodeID         string   `json:"node_id"`
	Action         string   `json:"action"`
	NativeObjects  []string `json:"native_objects"`
	BaselineSHA256 string   `json:"baseline_sha256"`
	CurrentSHA256  string   `json:"current_sha256,omitempty"`
	Requires       []string `json:"requires,omitempty"`
	Status         string   `json:"status"`
	Reason         string   `json:"reason,omitempty"`
}

func topDiagramNode(t LocalTemplate, id string) *Node {
	for i := range t.Nodes {
		if t.Nodes[i].ID == id {
			return &t.Nodes[i]
		}
	}
	return nil
}
func structureNodeHash(d Document, s Slide, t LocalTemplate, n *Node, names []string) string {
	if n == nil {
		return ""
	}
	values := map[string]any{}
	zones := map[string]Zone{}
	for _, v := range []any{n.Text, n.Asset, n.Arguments} {
		collectBindings(v, func(key string) { values[key] = s.Values[key]; zones[key] = t.Zones[key] })
	}
	geometry := map[string]NativeGeometry{}
	for _, name := range names {
		if g, ok := s.NativeGeometry[name]; ok {
			geometry[name] = g
		}
	}
	return digest(canonical([]any{n, values, zones, s.Template, t.Frame, t.FrameOptions, t.FrameChrome, t.Grid, d.EditingProfile, geometry}))
}
func addStructureReconciliation(p *Project, b *TextBaseline, native NativeLineageInspection, report *TextReconciliationReport) error {
	var base Document
	if e := strictInto(b.source, &base); e != nil {
		return e
	}
	bindings := map[string]ObjectRecord{}
	missing := map[string]bool{}
	blockedSlides := map[string]bool{}
	for _, o := range b.Objects.Objects {
		bindings[o.ShapeToken] = o
	}
	for _, issue := range native.Issues {
		if issue.Kind == "shape_missing" {
			missing[issue.ShapeToken] = true
			continue
		}
		if issue.SlideToken != "" {
			blockedSlides[b.Objects.Lineage.Slides[issue.SlideToken]] = true
		}
		if o, ok := bindings[issue.ShapeToken]; ok {
			blockedSlides[o.SlideID] = true
		}
	}
	current := map[string]Slide{}
	for _, s := range p.Document.Slides {
		current[s.ID] = s
	}
	resolved := map[string]bool{}
	for _, s := range base.Slides {
		if s.Template.Scope != "local" {
			continue
		}
		t := base.LocalTemplates[s.Template.ID]
		now, ok := current[s.ID]
		if !ok {
			continue
		}
		nt := p.Document.LocalTemplates[now.Template.ID]
		// Ownership is from the immutable object map, not edited names/shape IDs.
		owned := map[string][]ObjectRecord{}
		for _, o := range b.Objects.Objects {
			if o.SlideID == s.ID {
				owned[o.NodeID] = append(owned[o.NodeID], o)
			}
		}
		candidates := map[string]StructureReconciliationField{}
		for _, n := range t.Nodes {
			if len(n.Nodes) > 0 || len(owned[n.ID]) == 0 {
				continue
			}
			allMissing := true
			names := []string{}
			tokens := map[string]bool{}
			for _, o := range owned[n.ID] {
				allMissing = allMissing && missing[o.ShapeToken]
				names = append(names, o.NativeName)
				tokens[o.ShapeToken] = true
			}
			if !allMissing {
				continue
			}
			sort.Strings(names)
			f := StructureReconciliationField{SlideID: s.ID, NodeID: n.ID, Action: "remove_node", NativeObjects: names, BaselineSHA256: structureNodeHash(base, s, t, &n, names), Status: "native_only"}
			cur := topDiagramNode(nt, n.ID)
			f.CurrentSHA256 = structureNodeHash(p.Document, now, nt, cur, names)
			switch {
			case blockedSlides[s.ID]:
				f.Status = "manual_review"
				f.Reason = "The slide has untracked or ambiguous objects/ownership; resolve those identities before proposing deletion."
			case !bytes.Equal(canonical(s.Template), canonical(now.Template)) || now.Template.Revision != "":
				f.Status = "manual_review"
				f.Reason = "The local template reference changed or is revision-pinned; review ownership first."
			case cur != nil && len(cur.Nodes) > 0:
				f.Status = "manual_review"
				f.Reason = "The source node now owns a subtree; deletion needs a new baseline."
			case cur == nil:
				f.Status = "matching_changes"
			case f.BaselineSHA256 != f.CurrentSHA256:
				f.Status = "conflict"
				f.Reason = "Current YAML changed the removed component; choose whether to retain or delete it."
			}
			for _, other := range p.Document.Slides {
				if other.ID != s.ID && other.Template.Scope == "local" && other.Template.ID == now.Template.ID {
					f.Status = "manual_review"
					f.Reason = "The local definition is shared by other slides; fork it and build a new baseline."
				}
			}
			// Every mapped object must be contained only in this component's tree.
			for _, o := range owned[n.ID] {
				if o.NativeParentToken != "" && !tokens[o.NativeParentToken] {
					f.Status = "manual_review"
					f.Reason = "The component is nested under another native owner; reparenting needs explicit mapping."
				}
			}
			candidates[n.ID] = f
		}
		// A surviving incident edge must never be silently removed. Whole missing
		// edges are separate proposals; both must be selected in the same transaction.
		for id, f := range candidates {
			for _, n := range nt.Nodes {
				if !diagramEndpoint(n, id) {
					continue
				}
				edge, ok := candidates[n.ID]
				if !ok || edge.Status == "manual_review" {
					f.Status = "manual_review"
					f.Reason = "An attached incident edge survives or is ambiguous; remove it in the native copy, then propose again."
				} else {
					f.Requires = append(f.Requires, n.ID)
				}
			}
			sort.Strings(f.Requires)
			f.ID = digest(canonical(f))
			report.Structure = append(report.Structure, f)
			if f.Status != "manual_review" {
				for _, o := range owned[id] {
					resolved[o.ShapeToken] = true
				}
			}
		}
	}
	issues := report.ManualReview[:0]
	for _, issue := range report.ManualReview {
		if issue.Kind == "shape_missing" && resolved[issue.ShapeToken] {
			continue
		}
		issues = append(issues, issue)
	}
	report.ManualReview = issues
	sort.Slice(report.Structure, func(i, j int) bool { return report.Structure[i].ID < report.Structure[j].ID })
	for _, f := range report.Structure {
		report.Counts["structure_"+f.Status]++
	}
	return nil
}

// applyReviewedStructure uses the same AST/transaction as text and transforms.
// It deliberately does not call PatchDiagram's independently committing API.
func applyReviewedStructure(p *Project, report TextReconciliationReport, decisions TextReviewDecisions, slides map[string]*yaml.Node, documents map[string]*yaml.Node, touched map[string]bool, bundle, engine string) ([]AdoptedTextField, error) {
	fields := map[string]StructureReconciliationField{}
	accepted := map[string]bool{}
	selected := map[string]bool{}
	changed := []AdoptedTextField{}
	for _, f := range report.Structure {
		if _, ok := fields[f.ID]; ok {
			return nil, fmt.Errorf("reconcile.duplicate_structure_field")
		}
		fields[f.ID] = f
	}
	for _, d := range decisions.Decisions {
		f, ok := fields[d.FieldID]
		if !ok {
			continue
		}
		if selected[f.ID] || (f.Status != "native_only" && f.Status != "conflict") {
			return nil, fmt.Errorf("reconcile.invalid_structure_decision")
		}
		selected[f.ID] = true
		if d.Action == "use_native" {
			accepted[f.SlideID+"\x00"+f.NodeID] = true
		}
	}
	templates := map[string]LocalTemplate{}
	for _, d := range decisions.Decisions {
		f, ok := fields[d.FieldID]
		if !ok || d.Action == "keep_yaml" {
			continue
		}
		if f.Action != "remove_node" && f.Action != "copy_node" {
			return nil, fmt.Errorf("reconcile.unsupported_structure_action")
		}
		for _, id := range f.Requires {
			if !accepted[f.SlideID+"\x00"+id] {
				return nil, fmt.Errorf("reconcile.related_structure_decision_required: %s", id)
			}
		}
		idx, t, e := diagramSlide(p, f.SlideID)
		if e != nil {
			return nil, e
		}
		s := p.Document.Slides[idx]
		if _, ok := templates[f.SlideID]; !ok {
			var copy LocalTemplate
			if e = strictInto(t, &copy); e != nil {
				return nil, e
			}
			templates[f.SlideID] = copy
		}
		t = templates[f.SlideID]
		if f.Action == "copy_node" {
			sourceTemplate := p.Document.LocalTemplates[s.Template.ID]
			source := topDiagramNode(sourceTemplate, f.SourceNode)
			sourceNames := []string{}
			for name := range f.NativeGeometry {
				sourceNames = append(sourceNames, f.SourceNode+strings.TrimPrefix(name, f.NodeID))
			}
			sort.Strings(sourceNames)
			if source == nil || structureNodeHash(p.Document, s, sourceTemplate, source, sourceNames) != f.CurrentSHA256 || topDiagramNode(t, f.NodeID) != nil || f.NewNode == nil {
				return nil, fmt.Errorf("reconcile.copy_source_changed")
			}
			t.Nodes = append(t.Nodes, *f.NewNode)
			for key, z := range f.Zones {
				if _, ok := t.Zones[key]; ok {
					return nil, fmt.Errorf("reconcile.copy_zone_collision")
				}
				t.Zones[key] = z
			}
			templates[f.SlideID] = t
			slide := slides[f.SlideID]
			order, _ := editYAMLNode(f.EditedRootOrder)
			orders := mappingNode(slide, "native_order")
			if orders == nil {
				orders = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				replaceMappingField(slide, "native_order", orders)
			}
			replaceMappingField(orders, "", order)
			pin, _ := editYAMLNode(s.Template)
			replaceMappingField(slide, "native_geometry_template", pin)
			changed = append(changed, AdoptedTextField{f.ID, f.SlideID, "nodes/" + f.NodeID, "absent", "copied from " + f.SourceNode})
			continue
		}
		n := topDiagramNode(t, f.NodeID)
		if n == nil || structureNodeHash(p.Document, s, t, n, f.NativeObjects) != f.CurrentSHA256 {
			return nil, fmt.Errorf("reconcile.structure_source_changed")
		}
		for i := range t.Nodes {
			if t.Nodes[i].ID == f.NodeID {
				t.Nodes = append(t.Nodes[:i], t.Nodes[i+1:]...)
				break
			}
		}
		templates[f.SlideID] = t
		slide := slides[f.SlideID]
		removed := map[string]bool{}
		for _, name := range f.NativeObjects {
			removed[name] = true
			if geometry := mappingNode(slide, "native_geometry"); geometry != nil {
				removeMappingField(geometry, name)
			}
		}
		orders := mappingNode(slide, "native_order")
		if orders != nil {
			for i := 0; i+1 < len(orders.Content); {
				key, value := orders.Content[i], orders.Content[i+1]
				if removed[key.Value] {
					orders.Content = append(orders.Content[:i], orders.Content[i+2:]...)
					continue
				}
				kept := value.Content[:0]
				for _, name := range value.Content {
					if !removed[name.Value] {
						kept = append(kept, name)
					}
				}
				value.Content = kept
				i += 2
			}
		}
		changed = append(changed, AdoptedTextField{f.ID, f.SlideID, "nodes/" + f.NodeID, f.CurrentSHA256, "removed"})
	}
	for slideID, t := range templates {
		idx, _, e := diagramSlide(p, slideID)
		if e != nil {
			return nil, e
		}
		s := p.Document.Slides[idx]
		main := documents[filepath.Base(p.SourcePath)]
		node := mappingNode(mappingNode(main.Content[0], "local_templates"), s.Template.ID)
		file := p.TemplateFiles[s.Template.ID]
		if file != "" {
			doc, e := sourceYAML(p.SourceFiles[file])
			if e != nil {
				return nil, e
			}
			documents[file] = doc
			node = doc.Content[0]
		} else {
			file = filepath.Base(p.SourcePath)
		}
		if node == nil {
			return nil, fmt.Errorf("reconcile.local_definition_missing")
		}
		// Preserve any reviewed text edits already made in the slide AST when pruning.
		values := s.Values
		// content/bindings encoding is rebuilt from the current AST by loading the
		// intermediate authored files before pruning, avoiding lost text decisions.
		sf := p.SlideFiles[slideID]
		if sf == "" {
			sf = filepath.Base(p.SourcePath)
		}
		intermediate := map[string][]byte{}
		for name, doc := range documents {
			raw, e := encodeSourceYAML(doc)
			if e != nil {
				return nil, e
			}
			intermediate[name] = raw
		}
		cp, e := loadProject(p.SourcePath, mergeTextOverrides(p.SourceFiles, intermediate))
		if e != nil {
			return nil, e
		}
		for _, cs := range cp.Document.Slides {
			if cs.ID == slideID {
				values = cs.Values
			}
		}
		for _, f := range report.Structure {
			if f.SlideID == slideID && f.Action == "copy_node" && accepted[slideID+"\x00"+f.NodeID] {
				for key, v := range f.Values {
					values[key] = v
				}
			}
		}
		s.Values = values
		if _, e = updateDiagramNodes(t, s, node, slides[slideID], true); e != nil {
			return nil, e
		}
		// Empty overrides no longer require a template pin.
		slide := slides[slideID]
		for _, key := range []string{"native_geometry", "native_order"} {
			m := mappingNode(slide, key)
			if m != nil && len(m.Content) == 0 {
				removeMappingField(slide, key)
			}
		}
		if mappingNode(slide, "native_geometry") == nil && mappingNode(slide, "native_order") == nil {
			removeMappingField(slide, "native_geometry_template")
		}
		touched[file] = true
		touched[sf] = true
	}
	if len(templates) == 0 {
		return changed, nil
	}
	// Pin new transforms to the actual generated source basis, before applying
	// any existing native overrides. No opaque native XML enters authored YAML.
	overrides := map[string][]byte{}
	for name, doc := range documents {
		raw, e := encodeSourceYAML(doc)
		if e != nil {
			return nil, e
		}
		overrides[name] = raw
	}
	candidate, e := loadProject(p.SourcePath, mergeTextOverrides(p.SourceFiles, overrides))
	if e != nil {
		return nil, e
	}
	c, e := Check(candidate, bundle, engine)
	if e != nil {
		return nil, e
	}
	generated, _, e := wmdesign.BuildWithEngineAndAssets(bundle, "", c.Document, engine, c.Assets)
	if e != nil {
		return nil, e
	}
	pkg, e := openLineagePackage(generated)
	if e != nil {
		return nil, e
	}
	for _, f := range report.Structure {
		if f.Action != "copy_node" || !accepted[f.SlideID+"\x00"+f.NodeID] {
			continue
		}
		index := -1
		for i, s := range candidate.Document.Slides {
			if s.ID == f.SlideID {
				index = i
			}
		}
		raw, e := pkg.read(fmt.Sprintf("ppt/slides/slide%d.xml", index+1))
		if e != nil {
			return nil, e
		}
		inv, _, e := geometryInventory(raw)
		if e != nil {
			return nil, e
		}
		slide := slides[f.SlideID]
		geometry := mappingNode(slide, "native_geometry")
		if geometry == nil {
			geometry = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			replaceMappingField(slide, "native_geometry", geometry)
		}
		names := []string{}
		for name := range f.NativeGeometry {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			g := f.NativeGeometry[name]
			if inv[name] == nil {
				return nil, fmt.Errorf("reconcile.copy_generated_object_missing")
			}
			g.SourceGeometrySHA256 = nativeGeometryBasis(inv, name)
			v, e := editYAMLNode(g)
			if e != nil {
				return nil, e
			}
			replaceMappingField(geometry, name, v)
		}
	}
	return changed, nil
}
