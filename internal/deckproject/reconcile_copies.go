package deckproject

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const StructureMapSchema = "pptxgengo.native-structure-map.v1"

type NativeCopyMapping struct {
	SlideID      string `json:"slide_id"`
	NativeObject string `json:"native_object"`
	SourceNode   string `json:"source_node"`
	NewNode      string `json:"new_node"`
}
type NativeStructureMap struct {
	Schema string              `json:"schema"`
	Copies []NativeCopyMapping `json:"copies"`
}

func DecodeNativeStructureMap(raw []byte) (NativeStructureMap, error) {
	var out NativeStructureMap
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("structure map exceeds 1 MiB")
	}
	// Reuse the strict YAML reader: one document, no aliases or unknown fields.
	var v map[string]any
	p := &Project{Positions: map[string]Position{}}
	n, e := sourceYAML(raw)
	if e != nil {
		return out, e
	}
	x, e := p.yamlValue(n.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = strictInto(x, &v); e != nil {
		return out, e
	}
	if e = strictInto(v, &out); e != nil {
		return out, e
	}
	if out.Schema != StructureMapSchema || len(out.Copies) == 0 || len(out.Copies) > 100 {
		return out, fmt.Errorf("structure map requires schema and 1..100 copies")
	}
	seen := map[string]bool{}
	for _, m := range out.Copies {
		if !stableID.MatchString(m.SlideID) || !stableID.MatchString(m.SourceNode) || !stableID.MatchString(m.NewNode) || m.NativeObject == "" || len(m.NativeObject) > 512 {
			return out, fmt.Errorf("invalid native copy mapping")
		}
		for _, k := range []string{m.SlideID + "/source/" + m.NativeObject, m.SlideID + "/new/" + m.NewNode} {
			if seen[k] {
				return out, fmt.Errorf("duplicate native copy mapping")
			}
			seen[k] = true
		}
	}
	return out, nil
}
func copyObjectTree(n *xmlNode) []*xmlNode {
	out := []*xmlNode{}
	var walk func(*xmlNode)
	walk = func(n *xmlNode) {
		if nativeObjectKind(n) {
			out = append(out, n)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	return out
}

// Explicit selectors are diagnostic addresses in the pinned edited input, not
// persistent identity. Only known block payloads are reconstructed as typed
// source. New IDs and independent content bindings are mandatory.
func addCopyReconciliation(p *Project, b *TextBaseline, native NativeLineageInspection, report *TextReconciliationReport, mappings []NativeCopyMapping) (NativeLineageInspection, error) {
	if len(mappings) == 0 {
		return native, nil
	}
	var base Document
	if e := strictInto(b.source, &base); e != nil {
		return native, e
	}
	bindings := map[string]ObjectRecord{}
	for _, o := range b.Objects.Objects {
		bindings[o.ShapeToken] = o
	}
	originalByName := map[string]NativeLineageObject{}
	for _, o := range b.inspection.Objects {
		r := bindings[o.ShapeToken]
		originalByName[r.SlideID+"\x00"+r.NativeName] = o
	}
	selected := map[*xmlNode]bool{}
	for _, m := range mappings {
		var bs, cs *Slide
		for i := range base.Slides {
			if base.Slides[i].ID == m.SlideID {
				bs = &base.Slides[i]
			}
		}
		for i := range p.Document.Slides {
			if p.Document.Slides[i].ID == m.SlideID {
				cs = &p.Document.Slides[i]
			}
		}
		if bs == nil || cs == nil || bs.Template.Scope != "local" || !bytes.Equal(canonical(bs.Template), canonical(cs.Template)) || cs.Template.Revision != "" {
			return native, fmt.Errorf("copy mapping needs the same unpinned local template and slide")
		}
		t := base.LocalTemplates[bs.Template.ID]
		ct := p.Document.LocalTemplates[cs.Template.ID]
		n := topDiagramNode(t, m.SourceNode)
		if n == nil || n.Kind != "component" || n.Definition == nil || n.Definition.Scope != "shared" || (n.Definition.ID != "wmds/component/block" && n.Definition.ID != "wmds/component/editable-block") || len(n.Nodes) > 0 {
			return native, fmt.Errorf("copy mapping supports project-owned block/editable-block components only")
		}
		if list, _ := findDiagramNode(&ct.Nodes, m.NewNode); list != nil {
			return native, fmt.Errorf("copy mapping new node ID already exists")
		}
		for _, other := range p.Document.Slides {
			if other.ID != m.SlideID && other.Template.Scope == "local" && other.Template.ID == cs.Template.ID {
				return native, fmt.Errorf("fork the shared local definition before copying")
			}
		}
		original, ok := originalByName[m.SlideID+"\x00"+m.SourceNode]
		if !ok || original.ParentToken != "" {
			return native, fmt.Errorf("copy source needs one complete top-level native component")
		}
		var edited *NativeLineageObject
		for i := range native.Objects {
			o := &native.Objects[i]
			if b.Objects.Lineage.Slides[o.SlideToken] == m.SlideID && o.NativeName == m.NativeObject {
				if edited != nil {
					return native, fmt.Errorf("native copy selector is ambiguous")
				}
				edited = o
			}
		}
		if edited == nil || edited.ShapeToken != "" || edited.ParentToken != "" {
			return native, fmt.Errorf("native copy selector needs one untagged top-level object")
		}
		// Pointer containment, not ParentToken, distinguishes children of untagged
		// groups. Untagged children themselves have an empty ParentToken.
		for _, parent := range native.Objects {
			if parent.shape == edited.shape {
				continue
			}
			for _, child := range copyObjectTree(parent.shape)[1:] {
				if child == edited.shape {
					return native, fmt.Errorf("native copy selector is nested under another object")
				}
			}
		}
		a, z := copyObjectTree(original.shape), copyObjectTree(edited.shape)
		if len(a) != len(z) {
			return native, fmt.Errorf("native copy differs in style, structure, route or unsupported content: native object count %d != %d", len(a), len(z))
		}
		if mismatch := geometryCompatibilityMismatch(original.shape, edited.shape); mismatch != "" {
			return native, fmt.Errorf("native copy differs in style, structure, route or unsupported content: %s", mismatch)
		}
		owned := map[*xmlNode]ObjectRecord{}
		for _, o := range b.inspection.Objects {
			owned[o.shape] = bindings[o.ShapeToken]
		}
		names := []string{}
		for _, node := range a {
			r := owned[node]
			if r.NodeID != m.SourceNode {
				return native, fmt.Errorf("copy source subtree has mixed ownership")
			}
			names = append(names, r.NativeName)
		}
		sort.Strings(names)
		current := topDiagramNode(ct, m.SourceNode)
		baseHash := structureNodeHash(base, *bs, t, n, names)
		currentHash := structureNodeHash(p.Document, *cs, ct, current, names)
		if current == nil || currentHash != baseHash {
			return native, fmt.Errorf("copy source changed since baseline; rebuild before mapping native copies")
		}
		var clone Node
		if e := strictInto(n, &clone); e != nil {
			return native, e
		}
		clone.ID = m.NewNode
		zones := map[string]Zone{}
		values := map[string]any{}
		renames := map[string]string{}
		// A known block has a single maintained plain string text binding. Literal
		// copy is retained too; arbitrary rich payload is deliberately not imported.
		for _, node := range a {
			r := owned[node]
			if len(r.Fields) > 1 {
				return native, fmt.Errorf("copy has ambiguous text ownership")
			}
			for _, field := range r.Fields {
				if field.Status != "plain_text_baseline" {
					return native, fmt.Errorf("copy requires a maintained plain-text source field")
				}
				old := field.SourceSlot
				key := m.NewNode + ".text"
				if _, exists := ct.Zones[key]; exists {
					return native, fmt.Errorf("new copy binding already exists")
				}
				renames[old] = key
				zone, exists := t.Zones[old]
				if !exists {
					return native, fmt.Errorf("copy source zone missing")
				}
				zone.AuthoringAlias = m.NewNode + "_text"
				zones[key] = zone
			}
		}
		geometry := map[string]NativeGeometry{}
		for i, node := range z {
			if selected[node] {
				return native, fmt.Errorf("native copy mappings overlap")
			}
			var view *NativeLineageObject
			for j := range native.Objects {
				if native.Objects[j].shape == node {
					view = &native.Objects[j]
					break
				}
			}
			if view == nil || view.ShapeToken != "" {
				return native, fmt.Errorf("copy subtree must be entirely untagged; duplicated identities require separate review")
			}
			record := owned[a[i]]
			parent := ""
			if i > 0 {
				parent = m.NewNode
			}
			g, e := readNativeGeometry(node, parent)
			if e != nil {
				return native, e
			}
			if e = validateNativeGeometry(g); e != nil {
				return native, e
			}
			name := m.NewNode + strings.TrimPrefix(record.NativeName, m.SourceNode)
			geometry[name] = g
			paragraphs := nativeParagraphs(node)
			if len(record.Fields) == 1 {
				field := record.Fields[0]
				text, ok := nativeObjectFieldText(record, field, paragraphs)
				if !ok || !supportedEditedPlainText(view.Kind, paragraphs) || !utf8.ValidString(text) || len(text) > 64<<10 {
					return native, fmt.Errorf("native copy text is not supported plain copy")
				}
				values[renames[field.SourceSlot]] = text
			} else if nativeParagraphText(paragraphs) != record.NativeText {
				return native, fmt.Errorf("literal native copy text changed without a maintained field")
			}
			selected[node] = true
		}
		var replace func(any) any
		replace = func(v any) any {
			switch x := v.(type) {
			case map[string]any:
				y := map[string]any{}
				for k, v := range x {
					if k == "binding" {
						if old, ok := v.(string); ok && renames[old] != "" {
							v = renames[old]
						}
					}
					y[k] = replace(v)
				}
				return y
			case []any:
				y := make([]any, len(x))
				for i, v := range x {
					y[i] = replace(v)
				}
				return y
			default:
				return v
			}
		}
		if clone.Arguments != nil {
			clone.Arguments = replace(clone.Arguments).(map[string]any)
		}
		for old, key := range renames {
			if _, ok := values[key]; !ok {
				return native, fmt.Errorf("copy binding %s has no supported native text", old)
			}
		}
		f := StructureReconciliationField{SlideID: m.SlideID, NodeID: m.NewNode, SourceNode: m.SourceNode, Action: "copy_node", BaselineSHA256: baseHash, CurrentSHA256: currentHash, NewNode: &clone, Zones: zones, Values: values, NativeGeometry: geometry, Status: "native_only"}
		f.ID = digest(canonical(f))
		report.Structure = append(report.Structure, f)
	}
	// Suppress untagged-object issues only when every untagged object in a part
	// has an explicit validated mapping. Partial maps keep outstanding review.
	parts := map[string]bool{}
	for _, o := range native.Objects {
		if selected[o.shape] {
			parts[o.NativePart] = true
		}
	}
	for _, o := range native.Objects {
		if o.ShapeToken == "" && !selected[o.shape] {
			parts[o.NativePart] = false
		}
	}
	filtered := native
	filtered.Issues = nil
	for _, issue := range native.Issues {
		if issue.Kind == "shape_untagged" && parts[issue.NativePart] {
			continue
		}
		filtered.Issues = append(filtered.Issues, issue)
	}
	issues := report.ManualReview[:0]
	for _, issue := range report.ManualReview {
		if issue.Kind == "shape_untagged" && parts[issue.NativePart] {
			continue
		}
		issues = append(issues, issue)
	}
	report.ManualReview = issues
	return filtered, nil
}

// Read copies from the report only through a closed manifest input. Replaying
// arbitrary proposed source fields is forbidden; recompute them from this map.
func packetStructureMappings(packet *TextReviewPacket) ([]NativeCopyMapping, error) {
	raw, ok := packet.files["structure-map.json"]
	if !ok {
		return nil, nil
	}
	m, e := DecodeNativeStructureMap(raw)
	return m.Copies, e
}

func addStructureOrder(b *TextBaseline, native NativeLineageInspection, report *TextReconciliationReport, mappings []NativeCopyMapping) error {
	if len(mappings) == 0 {
		return nil
	}
	names := map[string]ObjectRecord{}
	for _, o := range b.Objects.Objects {
		names[o.ShapeToken] = o
	}
	roots := map[*xmlNode]string{}
	for _, m := range mappings {
		for _, o := range native.Objects {
			if b.Objects.Lineage.Slides[o.SlideToken] == m.SlideID && o.NativeName == m.NativeObject {
				roots[o.shape] = m.NewNode
			}
		}
	}
	orders := map[string][]string{}
	for _, o := range native.Objects {
		slide := b.Objects.Lineage.Slides[o.SlideToken]
		if name, ok := roots[o.shape]; ok {
			orders[slide] = append(orders[slide], name)
			continue
		}
		r, ok := names[o.ShapeToken]
		if ok && r.NativeParentToken == "" && o.ParentToken == "" {
			orders[slide] = append(orders[slide], r.NativeName)
		}
	}
	for i := range report.Structure {
		f := &report.Structure[i]
		if f.Action != "copy_node" {
			continue
		}
		// A complete root order is a coupled structural decision. All supported
		// native additions/deletions on that slide must be reviewed together.
		f.EditedRootOrder = orders[f.SlideID]
		for _, other := range report.Structure {
			if other.SlideID == f.SlideID && other.NodeID != f.NodeID && (other.Status == "native_only" || other.Status == "conflict") {
				f.Requires = append(f.Requires, other.NodeID)
			}
		}
		sort.Strings(f.Requires)
		f.ID = digest(canonical(*f))
	}
	sort.Slice(report.Structure, func(i, j int) bool { return report.Structure[i].ID < report.Structure[j].ID })
	return nil
}

// PowerPoint may copy our tags with the shape. The operator explicitly names
// each new copy. Ignore tags on those selected subtrees for analysis only;
// retain the exact original edited bytes in the packet and adoption receipt.
// Any duplicate identities left outside the selection remain unresolved.
func prepareMappedNativeCopies(edited []byte, b *TextBaseline, mappings []NativeCopyMapping) ([]byte, error) {
	if len(mappings) == 0 {
		return edited, nil
	}
	view, e := InspectNativeLineage(edited, b.Objects)
	if e != nil {
		return nil, e
	}
	pkg, e := openLineagePackage(edited)
	if e != nil {
		return nil, e
	}

	byAddress := map[string]*NativeLineageObject{}
	ambiguous := map[string]bool{}
	nested := map[*xmlNode]bool{}
	for i := range view.Objects {
		o := &view.Objects[i]
		key := b.Objects.Lineage.Slides[o.SlideToken] + "\x00" + o.NativeName
		if byAddress[key] != nil {
			ambiguous[key] = true
		}
		byAddress[key] = o
		for _, child := range copyObjectTree(o.shape)[1:] {
			nested[child] = true
		}
	}
	roots := []*NativeLineageObject{}
	selectedMembers := map[*xmlNode]bool{}
	for _, m := range mappings {
		key := m.SlideID + "\x00" + m.NativeObject
		selected := byAddress[key]
		if ambiguous[key] {
			return nil, fmt.Errorf("native copy selector is ambiguous")
		}
		if selected == nil {
			return nil, fmt.Errorf("native copy selector not found")
		}
		if selected.ParentToken != "" || nested[selected.shape] {
			return nil, fmt.Errorf("native copy selector must be top-level")
		}
		for _, node := range copyObjectTree(selected.shape) {
			if selectedMembers[node] {
				return nil, fmt.Errorf("native copy mappings overlap")
			}
			selectedMembers[node] = true
		}
		roots = append(roots, selected)
	}
	remainingTokens := map[string]int{}
	for _, o := range view.Objects {
		if !selectedMembers[o.shape] && o.ShapeToken != "" {
			remainingTokens[o.ShapeToken]++
		}
	}
	for _, o := range view.Objects {
		if selectedMembers[o.shape] && o.ShapeToken != "" && remainingTokens[o.ShapeToken] != 1 {
			return nil, fmt.Errorf("mapped copy must leave exactly one original instance of each inherited identity")
		}
	}
	trees := map[string]map[string]*lineageSpan{}
	patches := map[string][]lineagePatch{}
	for _, selected := range roots {
		raw, e := pkg.read(selected.NativePart)
		if e != nil {
			return nil, e
		}
		if trees[selected.NativePart] == nil {
			tree, e := lineageSpans(raw)
			if e != nil {
				return nil, e
			}
			byName := map[string]*lineageSpan{}
			var find func(*lineageSpan)
			find = func(s *lineageSpan) {
				if nativeObjectKind(s.node) {
					byName[geometrySpanName(s)] = s
				}
				for _, child := range s.children {
					find(child)
				}
			}
			find(tree)
			trees[selected.NativePart] = byName
		}
		target := trees[selected.NativePart][selected.NativeName]
		if target == nil {
			return nil, fmt.Errorf("native copy span missing")
		}
		var strip func(*lineageSpan)
		strip = func(s *lineageSpan) {
			if s.node.Name.Space == lineagePML && s.node.Name.Local == "custDataLst" {
				patches[selected.NativePart] = append(patches[selected.NativePart], lineagePatch{s.start, s.end, ""})
				return
			}
			for _, child := range s.children {
				strip(child)
			}
		}
		strip(target)
	}

	changes := map[string][]byte{}
	for part, edits := range patches {
		if len(edits) == 0 {
			continue
		}
		raw, e := pkg.read(part)
		if e != nil {
			return nil, e
		}
		raw, e = lineageApply(raw, edits)
		if e != nil {
			return nil, e
		}
		changes[part] = raw
	}
	if len(changes) == 0 {
		return edited, nil
	}
	return lineageRewrite(pkg, changes)
}

func geometrySpanName(s *lineageSpan) string {
	for _, nv := range s.children {
		if nv.node.Name.Space != lineagePML || !strings.HasPrefix(nv.node.Name.Local, "nv") {
			continue
		}
		for _, id := range nv.children {
			if id.node.Name.Space == lineagePML && id.node.Name.Local == "cNvPr" {
				return lineageAttr(id.node, "", "name")
			}
		}
	}
	return ""
}
