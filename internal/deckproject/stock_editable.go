package deckproject

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

// MarshalSlideSource orders authoring fields with visible copy above bindings.
func MarshalSlideSource(authored map[string]any) ([]byte, error) {
	node, err := editYAMLNode(authored)
	if err != nil {
		return nil, err
	}
	orderAuthoredSlide(node)
	return encodeSourceYAML(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}})
}

// StockScaffoldSlide creates editable example content while preserving a genuine
// shared template reference. Example content is explicitly marked synthetic.
func StockScaffoldSlide(bundle, key, id string, year int) (map[string]any, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("stock scaffold requires a slide ID and template key")
	}
	catalog, err := wmdesign.LibraryCatalog(bundle, "")
	if err != nil {
		return nil, err
	}
	for _, def := range catalog {
		if def.Key != key {
			continue
		}
		examples, err := wmdesign.LibraryReference(bundle, "", def.Family, year)
		if err != nil {
			return nil, err
		}
		for _, example := range examples.Slides {
			if example.Template != key {
				continue
			}
			values := map[string]any{}
			if err := json.Unmarshal(example.Values, &values); err != nil {
				return nil, err
			}
			return StockEditableSlide(Slide{ID: id, ContentKind: "synthetic_example", Template: Reference{Scope: "shared", ID: key, Revision: strconv.Itoa(def.Revision)}, Values: values}, def)
		}
		return nil, fmt.Errorf("stock example unavailable for %s", key)
	}
	return nil, fmt.Errorf("unknown shared template %s", key)
}

// StockEditableSlide exposes supplied copy separately from the stock contract.
// Binding pointers translate friendly fields back to the original closed values;
// this function never changes a layout, adds specimen copy, or creates a local
// template. The returned authoring map compiles to the same Slide.
func StockEditableSlide(slide Slide, def wmdesign.LibraryTemplate) (map[string]any, error) {
	if slide.Template.Scope != "shared" || slide.Template.ID != def.Key {
		return nil, fmt.Errorf("stock content requires the slide's exact shared template")
	}
	var authored map[string]any
	if err := json.Unmarshal(canonical(slide), &authored); err != nil {
		return nil, err
	}
	values, _ := authored["values"].(map[string]any)
	if values == nil {
		return nil, fmt.Errorf("stock slide %s requires supplied values", slide.ID)
	}
	content, bindings := map[string]any{}, map[string]any{}
	metadata, err := wmdesign.LibraryAuthoringMetadata(def)
	if err != nil {
		return nil, err
	}
	semantic := map[string]wmdesign.LibraryAuthoringSlot{}
	for _, slot := range metadata.Slots {
		semantic[slot.Name] = slot
	}
	if raw, exists := values["slots"]; exists && len(def.Slots) > 0 {
		slots, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("stock slide %s slots must be a mapping", slide.ID)
		}
		known := map[string]bool{}
		for _, slot := range def.Slots {
			known[slot.Name] = true
			value, present := slots[slot.Name]
			if !present {
				continue
			}
			info, exists := semantic[slot.Name]
			if !exists {
				return nil, fmt.Errorf("stock content metadata missing: %s", slot.Name)
			}
			if info.Classification == "decorative" && value == "" {
				continue
			}
			parts := strings.Split(strings.TrimPrefix(info.Alias, "/"), "/")
			pointer := "/" + strings.Join(parts, "/")
			if _, duplicate := bindings[pointer]; duplicate {
				parts = []string{"additional_content", stockFieldName(slot.Name)}
				pointer = "/" + strings.Join(parts, "/")
			}
			if _, duplicate := bindings[pointer]; duplicate {
				return nil, fmt.Errorf("stock content alias collision: %s", slot.Name)
			}
			if err := stockSetContent(content, parts, value); err != nil {
				return nil, fmt.Errorf("stock content %s: %w", slot.Name, err)
			}
			bindings[pointer] = "/slots/" + escape(slot.Name)
			delete(slots, slot.Name)
		}
		for name := range slots {
			if !known[name] {
				return nil, fmt.Errorf("stock slide %s has undeclared slot %s", slide.ID, name)
			}
		}
		if len(slots) == 0 {
			delete(values, "slots")
		}
	} else {
		// Typed templates already have human-readable arrays and field names.
		// Keep technical stable keys and navigation beside the bindings.
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if key == "keys" || key == "nav" || key == "emphasis" || key == "whiteboard" {
				continue
			}
			alias := key
			if key == "title" {
				alias = "headline"
			}
			if _, duplicate := content[alias]; duplicate {
				alias = "value_" + key
			}
			content[alias] = values[key]
			bindings[alias] = "/" + escape(key)
			delete(values, key)
		}
	}
	if len(content) > 0 {
		authored["content"], authored["bindings"] = content, bindings
	}
	if len(values) == 0 {
		delete(authored, "values")
	}
	return authored, nil
}

func stockFieldName(s string) string {
	var out strings.Builder
	underscore := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out.WriteRune(r)
			underscore = false
		} else if out.Len() > 0 && !underscore {
			out.WriteByte('_')
			underscore = true
		}
	}
	name := strings.Trim(out.String(), "_")
	if name == "" {
		return "value"
	}
	return name
}

func stockContentPath(slot wmdesign.LibrarySlot, source map[string]any, def wmdesign.LibraryTemplate, values map[string]any) []string {
	if slot.SourcePointer == "/title" {
		return []string{"headline"}
	}
	if slot.SourcePointer == "/eyebrow" {
		return []string{"section_label"}
	}
	if slot.SourcePointer == "/source/text" {
		return []string{"source_note"}
	}
	parts := strings.Split(strings.TrimPrefix(slot.SourcePointer, "/"), "/")
	if len(parts) < 3 || parts[0] != "body" {
		return []string{stockFieldName(slot.Name)}
	}
	index, err := strconv.Atoi(parts[1])
	body, ok := source["body"].([]any)
	if err != nil || !ok || index < 0 || index >= len(body) {
		return []string{stockFieldName(slot.Name)}
	}
	node, ok := body[index].(map[string]any)
	if !ok {
		return []string{stockFieldName(slot.Name)}
	}
	kind, _ := node["type"].(string)
	if strings.HasPrefix(def.Key, "cover/") && kind == "text" && len(parts) == 3 && parts[2] == "text" {
		style, _ := node["style"].(string)
		aliases := map[string]string{"display": "headline", "eyebrow": "section_label", "lead": "subtitle"}
		if alias := aliases[style]; alias != "" {
			matches, conflicting := 0, false
			for _, raw := range body {
				if n, ok := raw.(map[string]any); ok && n["type"] == "text" && n["style"] == style {
					matches++
				}
			}
			for _, candidate := range def.Slots {
				if (alias == "headline" && candidate.SourcePointer == "/title") || (alias == "section_label" && candidate.SourcePointer == "/eyebrow") {
					conflicting = true
				}
			}
			if matches == 1 && !conflicting {
				return []string{alias}
			}
		}
	}
	if kind == "" {
		kind = "content"
	}
	out := []string{stockFieldName(kind) + "_" + strconv.Itoa(index+1)}
	for i, part := range parts[2:] {
		if _, err := strconv.Atoi(part); err == nil {
			// Named items avoid sparse arrays when fixed decoration has no
			// editable text, and remain stable while the copy changes.
			n, _ := strconv.Atoi(part)
			out = append(out, fmt.Sprintf("item_%02d", n+1))
			continue
		}
		if part == "cols" {
			part = "columns"
		}
		if i == 2 && parts[2] == "rows" {
			part = stockRowField(part, node, index, def, values)
		}
		out = append(out, stockFieldName(part))
	}
	return out
}

func stockRowField(field string, node map[string]any, index int, def wmdesign.LibraryTemplate, values map[string]any) string {
	cols, _ := node["cols"].([]any)
	for colIndex, raw := range cols {
		col, ok := raw.(map[string]any)
		if !ok || col["key"] != field {
			continue
		}
		label := stockColumnLabel(col, index, colIndex, def, values)
		if strings.TrimSpace(label) != "" {
			// Retain the source field as a suffix when two columns share a
			// heading, keeping each row's binding deterministic.
			name := stockFieldName(label)
			for otherIndex, other := range cols {
				m, _ := other.(map[string]any)
				if m["key"] != field && stockFieldName(stockColumnLabel(m, index, otherIndex, def, values)) == name {
					return name + "_" + stockFieldName(field)
				}
			}
			return name
		}
	}
	return field
}

func stockColumnLabel(col map[string]any, index, colIndex int, def wmdesign.LibraryTemplate, values map[string]any) string {
	label, _ := col["label"].(string)
	pointer := fmt.Sprintf("/body/%d/cols/%d/label", index, colIndex)
	for _, slot := range def.Slots {
		if slot.SourcePointer == pointer {
			if actual, ok := values[slot.Name].(string); ok && strings.TrimSpace(actual) != "" {
				return actual
			}
		}
	}
	return label
}

func stockSetContent(root map[string]any, path []string, value any) error {
	current := root
	for _, key := range path[:len(path)-1] {
		if next, present := current[key]; present {
			var ok bool
			current, ok = next.(map[string]any)
			if !ok {
				return fmt.Errorf("overlapping content path")
			}
		} else {
			next := map[string]any{}
			current[key] = next
			current = next
		}
	}
	key := path[len(path)-1]
	if _, present := current[key]; present {
		return fmt.Errorf("duplicate content path")
	}
	current[key] = value
	return nil
}
