package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"io"
	"sort"
	"strconv"
	"strings"
)

type ObjectRecord struct {
	NativeParentToken     string              `json:"native_parent_token,omitempty"`
	ShapeToken            string              `json:"shape_token,omitempty"`
	SourceTemplate        Reference           `json:"source_template"`
	SourceSlots           map[string]string   `json:"source_slots,omitempty"`
	NativeStructureSHA256 string              `json:"native_structure_sha256,omitempty"`
	NativeKind            string              `json:"native_kind"`
	Paragraphs            []NativeParagraph   `json:"native_paragraphs,omitempty"`
	Fields                []NativeSourceField `json:"source_fields,omitempty"`
	TextMapping           string              `json:"text_mapping,omitempty"`
	TextMappingReason     string              `json:"text_mapping_reason,omitempty"`
	LogicalID             string              `json:"logical_id"`
	SlideID               string              `json:"slide_id"`
	NodeID                string              `json:"node_id"`
	ItemKey               string              `json:"item_key,omitempty"`
	PartRole              string              `json:"part_role"`
	NativePart            string              `json:"native_part"`
	NativeID              string              `json:"native_id"`
	NativeName            string              `json:"native_name"`
	NativeText            string              `json:"baseline_native_text,omitempty"`
	SourcePointers        []string            `json:"source_pointers"`
	BaselineValues        map[string]any      `json:"baseline_values,omitempty"`
	Mapping               string              `json:"mapping"`
}
type Objects struct {
	Lineage         *NativeLineage `json:"native_lineage,omitempty"`
	TextModelSchema string         `json:"text_model_schema,omitempty"`
	Schema          string         `json:"schema"`
	DeckID          string         `json:"deck_id"`
	SourceSHA256    string         `json:"source_semantic_sha256"`
	Objects         []ObjectRecord `json:"objects"`
	Reconciliation  string         `json:"reconciliation"`
}
type xmlNode struct {
	Name     xml.Name
	Attrs    []xml.Attr
	Children []*xmlNode
	Text     string
}

func readXML(r io.Reader) (*xmlNode, error) {
	d := xml.NewDecoder(r)
	root := &xmlNode{}
	stack := []*xmlNode{root}
	for {
		t, e := d.Token()
		if e == io.EOF {
			return root, nil
		}
		if e != nil {
			return nil, e
		}
		switch t := t.(type) {
		case xml.StartElement:
			n := &xmlNode{Name: t.Name, Attrs: t.Attr}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, n)
			stack = append(stack, n)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			stack[len(stack)-1].Text += string(t)
		}
	}
}
func attr(n *xmlNode, key string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == key {
			return a.Value
		}
	}
	return ""
}
func descendants(n *xmlNode, key string) []*xmlNode {
	out := []*xmlNode{}
	for _, c := range n.Children {
		if c.Name.Local == key {
			out = append(out, c)
		}
		out = append(out, descendants(c, key)...)
	}
	return out
}
func ObjectMap(p *Project, doc wmdesign.Document, data []byte) (Objects, error) {
	out := Objects{TextModelSchema: NativeTextModelSchema, Schema: "pptxgengo.deck-object-map.v1", DeckID: p.Document.ID, SourceSHA256: digest(p.Canonical), Objects: []ObjectRecord{}, Reconciliation: "baseline plus stable IDs available; edited-PPTX three-way reconciliation is not implemented"}
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		return out, e
	}
	files := map[string]*zip.File{}
	for _, f := range z.File {
		files[f.Name] = f
	}
	for i, s := range doc.Slides {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		f := files[part]
		if f == nil {
			return out, fmt.Errorf("missing native slide %s", part)
		}
		r, e := f.Open()
		if e != nil {
			return out, e
		}
		tree, e := readXML(r)
		r.Close()
		if e != nil {
			return out, e
		}
		var identityError error
		var walk func(*xmlNode)
		walk = func(n *xmlNode) {
			switch n.Name.Local {
			case "sp", "pic", "graphicFrame", "grpSp":
				id, identityErr := nativeObjectIdentity(n)
				if identityErr != nil {
					identityError = identityErr
					return
				}
				if id != nil {
					name := attr(id, "name")
					rec := ObjectRecord{NativeKind: n.Name.Local, LogicalID: s.ID + "/" + name, SlideID: s.ID, NodeID: name, PartRole: "native-object", NativePart: part, NativeID: attr(id, "id"), NativeName: name, SourcePointers: []string{}, BaselineValues: map[string]any{}, Mapping: "generated-unbound"}
					rec.NativeStructureSHA256 = nativeStructureHash(n)
					rec.Paragraphs = nativeParagraphs(n)
					rec.NativeText = nativeParagraphText(rec.Paragraphs)
					rec = attachNativeSourceFields(p, bindObject(p, i, s, rec))
					out.Objects = append(out.Objects, rec)
				}
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(tree)
		if identityError != nil {
			return out, identityError
		}
	}
	sort.Slice(out.Objects, func(i, j int) bool {
		a, b := out.Objects[i], out.Objects[j]
		if a.SlideID == b.SlideID {
			return a.NativeID < b.NativeID
		}
		return a.SlideID < b.SlideID
	})
	return out, nil
}
func bindObject(p *Project, index int, s wmdesign.SlideSpec, r ObjectRecord) ObjectRecord {
	ptr := "/slides/" + strconv.Itoa(index) + "/values"
	r.SourceTemplate = p.Document.Slides[index].Template
	r.SourceSlots = map[string]string{}
	longest := ""
	for _, n := range s.Nodes {
		if (r.NativeName == n.ID || strings.HasPrefix(r.NativeName, n.ID+".")) && len(n.ID) > len(longest) {
			longest = n.ID
		}
	}
	if longest != "" {
		r.NodeID = longest
		r.PartRole = strings.TrimPrefix(strings.TrimPrefix(r.NativeName, longest), ".")
		if r.PartRole == "" {
			r.PartRole = "body"
		}
	}
	if s.TemplateBinding != nil {
		for _, a := range s.TemplateBinding.Assignments {
			if typedCardAssignmentMatches(r.NativeName, a.TargetID, a.Property) {
				pointer := slotPointer(p.Document.Slides[index].Values, ptr, a.Slot)
				r.SourcePointers = append(r.SourcePointers, pointer)
				r.SourceSlots[pointer] = a.Slot
				var val any
				_ = json.Unmarshal(a.Value, &val)
				r.BaselineValues[a.Slot] = val
				r.Mapping = "template-field-baseline"
			}
		}
		for _, k := range s.TemplateBinding.CardKeys {
			if r.NativeName == k.TargetID || strings.HasPrefix(r.NativeName, k.TargetID+".") {
				r.ItemKey = k.Key
			}
		}
	} else {
		t := p.Document.LocalTemplates[p.Document.Slides[index].Template.ID]
		var walk func([]Node, string, string)
		walk = func(ns []Node, prefix string, nodePath string) {
			for ni, n := range ns {
				currentPath := nodePath + "/" + strconv.Itoa(ni)
				id := prefix + n.ID
				if n.Kind == "group" {
					walk(n.Nodes, id+".", currentPath+"/nodes")
					continue
				}
				if id != longest {
					continue
				}
				r.SourcePointers = append(r.SourcePointers, currentPath)
				for _, v := range []any{n.Text, n.Asset, n.Arguments} {
					collectBindings(v, func(key string) {
						r.SourcePointers = append(r.SourcePointers, ptr+"/"+escape(key))
						r.SourceSlots[ptr+"/"+escape(key)] = key
						r.BaselineValues[key] = p.Document.Slides[index].Values[key]
					})
				}
				r.Mapping = "local-node-baseline"
			}
		}
		walk(t.Nodes, "", "/local_templates/"+escape(p.Document.Slides[index].Template.ID)+"/nodes")
	}
	if r.NativeName == "title" || r.NativeName == "eyebrow" || r.NativeName == "source" {
		r.Mapping = "frame-field-baseline"
		field := r.NativeName
		if p.Document.Slides[index].Template.Scope == "local" {
			t := p.Document.LocalTemplates[p.Document.Slides[index].Template.ID]
			role := field
			if field == "title" {
				role = "slide-title"
			}
			for name, zone := range t.Zones {
				if zone.Role == role {
					field = name
					break
				}
			}
		}
		pointer := slotPointer(p.Document.Slides[index].Values, ptr, field)
		if value, err := lookupPointer(p.tree, pointer); err == nil {
			r.SourcePointers = append(r.SourcePointers, pointer)
			r.SourceSlots[pointer] = field
			r.BaselineValues[field] = value
		}
	}
	if r.NodeID == "" {
		r.NodeID = "native-" + r.NativeID
	}
	r.LogicalID = r.SlideID + "/" + r.NodeID + "/" + r.PartRole
	if r.ItemKey != "" {
		r.LogicalID += "/" + r.ItemKey
	}
	sort.Strings(r.SourcePointers)
	unique := []string{}
	for _, pointer := range r.SourcePointers {
		if len(unique) == 0 || unique[len(unique)-1] != pointer {
			unique = append(unique, pointer)
		}
	}
	r.SourcePointers = unique
	return r
}
func collectBindings(v any, found func(string)) {
	switch x := v.(type) {
	case map[string]any:
		if k, ok := x["binding"].(string); ok {
			found(k)
		} else {
			for _, v := range x {
				collectBindings(v, found)
			}
		}
	case []any:
		for _, v := range x {
			collectBindings(v, found)
		}
	}
}

func slotPointer(values map[string]any, base, slot string) string {
	if slots, ok := values["slots"].(map[string]any); ok {
		if _, exists := slots[slot]; exists {
			return base + "/slots/" + escape(slot)
		}
	}
	var v any = values
	path := base
	parts := strings.Split(slot, ".")
	for pos := 0; pos < len(parts); pos++ {
		key := parts[pos]
		switch cur := v.(type) {
		case map[string]any:
			next, ok := cur[key]
			if !ok {
				return base
			}
			path += "/" + escape(key)
			v = next
		case []any:
			found := false
			for end := len(parts); end > pos && !found; end-- {
				key = strings.Join(parts[pos:end], ".")
				for i, item := range cur {
					m, ok := item.(map[string]any)
					if ok && m["key"] == key {
						path += "/" + strconv.Itoa(i)
						v = m
						found = true
						pos = end - 1
						break
					}
				}
			}
			if !found {
				return base
			}
		default:
			return base
		}
	}
	return path
}
