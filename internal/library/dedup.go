package library

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Fingerprint describes interchangeable composition behavior. Source evidence,
// preference and qualification intentionally remain on each contract record.
func (s Store) Fingerprint(c Contract) (string, error) {
	p, err := s.SafePath(c.Composition.SpecPath)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	var spec map[string]any
	if err := json.Unmarshal(b, &spec); err != nil {
		return "", err
	}
	var shape any
	if slides, ok := spec["slides"].([]any); ok {
		for _, slide := range slides {
			m, _ := slide.(map[string]any)
			if m["id"] == c.Composition.SlideID {
				shape = m
				break
			}
		}
	}
	if shape == nil {
		return "", fmt.Errorf("contract %s template slide %s missing", c.ID, c.Composition.SlideID)
	}
	for _, slot := range c.Composition.Slots {
		if !validContentPointer(slot.Pointer) {
			return "", fmt.Errorf("slot %s targets non-content pointer", slot.Name)
		}
		if err := pointerSet(shape, slot.Pointer, "<slot>"); err != nil {
			return "", err
		}
	}
	for field, ptr := range c.Composition.NarrativeBindings {
		if field != "assertion_title" && field != "role" && field != "takeaway" {
			return "", fmt.Errorf("unsupported narrative binding %s", field)
		}
		if !validContentPointer(ptr) {
			return "", fmt.Errorf("narrative binding %s targets non-content pointer", field)
		}
		v, err := pointerGet(shape, ptr)
		if err != nil {
			return "", err
		}
		if _, ok := v.(string); !ok {
			return "", fmt.Errorf("narrative binding %s must target string", field)
		}
	}
	if ptr := c.Composition.PaginationBinding; ptr != "" {
		if !validContentPointer(ptr) {
			return "", fmt.Errorf("pagination binding targets non-content pointer")
		}
		for _, slot := range c.Composition.Slots {
			if slot.Pointer == ptr {
				return "", fmt.Errorf("pagination binding overlaps slot %s", slot.Name)
			}
		}
		v, err := pointerGet(shape, ptr)
		if err != nil {
			return "", err
		}
		if _, ok := v.(string); !ok {
			return "", fmt.Errorf("pagination binding must target string")
		}
		if err := pointerSet(shape, ptr, "<page>"); err != nil {
			return "", err
		}
	}
	ids := map[string]string{}
	collectIDs(shape, ids)
	collectCompoundIDs(shape, ids)
	colors := declaredColors(c.StyleVariants)
	if err := normalizeShape(s, shape, ids, colors); err != nil {
		return "", err
	}
	slots := make([]Slot, len(c.Composition.Slots))
	copy(slots, c.Composition.Slots)
	cardinality := make(map[string]Range, len(c.Cardinality))
	for i := range slots {
		if r, ok := c.Cardinality[slots[i].Name]; ok {
			cardinality[slots[i].Pointer] = r
		}
		slots[i].Name = "" // names often contain a source slide or object ID
		if len(slots[i].ItemTemplate) > 0 {
			var item any
			if err := json.Unmarshal(slots[i].ItemTemplate, &item); err != nil {
				return "", err
			}
			if slots[i].ItemValuePointer != "" {
				if err := pointerSet(item, slots[i].ItemValuePointer, "<slot>"); err != nil {
					return "", err
				}
			}
			if slots[i].ItemIDPointer != "" {
				if err := pointerSet(item, slots[i].ItemIDPointer, "<generated-id>"); err != nil {
					return "", err
				}
				slots[i].ItemIDPrefix = ""
			}
			itemIDs := map[string]string{}
			collectIDs(item, itemIDs)
			if err := normalizeShape(s, item, itemIDs, colors); err != nil {
				return "", err
			}
			slots[i].ItemTemplate, err = json.Marshal(item)
			if err != nil {
				return "", err
			}
		}
	}
	// Unknown cardinality keys are retained, so they cannot silently disappear.
	for name, r := range c.Cardinality {
		found := false
		for _, slot := range c.Composition.Slots {
			if slot.Name == name {
				found = true
				break
			}
		}
		if !found {
			cardinality["unknown:"+name] = r
		}
	}
	key, err := json.Marshal(struct {
		Shape       any
		Slots       []Slot
		Cardinality map[string]Range
		Transforms  Transforms
		Bindings    map[string]string
		Pagination  string
	}{shape, slots, cardinality, c.Transforms, c.Composition.NarrativeBindings, c.Composition.PaginationBinding})
	if err != nil {
		return "", err
	}
	return hashBytes(key), nil
}

func collectIDs(v any, ids map[string]string) {
	counts := map[string]int{}
	collectIDsWalk(v, ids, counts)
	for id, count := range counts {
		if count > 1 {
			delete(ids, id)
		}
	}
}

func collectIDsWalk(v any, ids map[string]string, counts map[string]int) {
	switch x := v.(type) {
	case map[string]any:
		if id, ok := x["id"].(string); ok {
			counts[id]++
			if _, exists := ids[id]; !exists {
				ids[id] = fmt.Sprintf("<id:%d>", len(ids))
			}
		}
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			collectIDsWalk(x[k], ids, counts)
		}
	case []any:
		for _, item := range x {
			collectIDsWalk(item, ids, counts)
		}
	}
}

func collectCompoundIDs(v any, ids map[string]string) {
	slide, ok := v.(map[string]any)
	if !ok {
		return
	}
	layouts, _ := slide["layouts"].([]any)
	compound := map[string]string{}
	counts := map[string]int{}
	for _, raw := range layouts {
		layout, _ := raw.(map[string]any)
		container, _ := layout["id"].(string)
		if ids[container] == "" {
			continue
		}
		cells, _ := layout["cells"].([]any)
		for ci, rawCell := range cells {
			cell, _ := rawCell.(map[string]any)
			cellID, _ := cell["id"].(string)
			if cellID == "" {
				continue
			}
			path := container + "/" + cellID
			counts[path]++
			compound[path] = fmt.Sprintf("%s/<cell:%d>", ids[container], ci)
			blocks, _ := cell["blocks"].([]any)
			for bi, rawBlock := range blocks {
				block, _ := rawBlock.(map[string]any)
				blockID, _ := block["id"].(string)
				if blockID == "" {
					continue
				}
				blockPath := path + "/" + blockID
				counts[blockPath]++
				compound[blockPath] = fmt.Sprintf("%s/<block:%d>", compound[path], bi)
			}
		}
	}
	for path, canonical := range compound {
		if counts[path] == 1 {
			if _, exists := ids[path]; !exists {
				ids[path] = canonical
			}
		}
	}
}

var fingerprintColorFields = map[string]bool{
	"background": true, "foreground": true, "contrast_background": true,
	"surface": true, "accent": true, "outline_color": true,
	"title_foreground": true, "color": true, "node_background": true,
	"arrow_color": true,
}

// Only explicit token assignments identify equivalent colors. Ambiguous colors
// (two semantic meanings assigned the same value) stay concrete.
func declaredColors(variants []StyleVariant) map[string]string {
	colors := map[string]string{}
	ambiguous := map[string]bool{}
	for _, variant := range variants {
		for token, color := range variant.Tokens {
			color = strings.ToUpper(color)
			if old, exists := colors[color]; exists && old != token {
				ambiguous[color] = true
			}
			colors[color] = token
		}
	}
	for color := range ambiguous {
		delete(colors, color)
	}
	return colors
}

func normalizeShape(s Store, v any, ids, colors map[string]string) error {
	switch x := v.(type) {
	case map[string]any:
		delete(x, "notes")
		delete(x, "metadata")
		delete(x, "provenance")
		delete(x, "source_id")
		delete(x, "source_path")
		delete(x, "source_part")
		delete(x, "source_slide")
		delete(x, "source_slide_id")
		delete(x, "source_object_ids")
		delete(x, "endpoint_provenance")
		delete(x, "ink_regions_provenance")
		// An asset path is interchangeable only after its declared digest has
		// been checked against the actual bytes.
		for _, pair := range [][2]string{{"asset_path", "asset_sha256"}, {"fallback_asset_path", "fallback_asset_sha256"}} {
			if path, ok := x[pair[0]].(string); ok {
				if digest, ok := x[pair[1]].(string); ok && digest != "" {
					if err := s.VerifyArtifact(Artifact{Path: path, SHA256: digest}); err != nil {
						return err
					}
					x[pair[0]] = "<asset:" + digest + ">"
				}
			}
		}
		for k, item := range x {
			if str, ok := item.(string); ok {
				if (k == "id" || k == "parent_id" || k == "from" || k == "to" || k == "target" || strings.HasSuffix(k, "_id")) && ids[str] != "" {
					x[k] = ids[str]
				} else if fingerprintColorFields[k] {
					if token := colors[strings.ToUpper(str)]; token != "" {
						x[k] = "<color:" + token + ">"
					}
				}
			} else if k == "allow_overlap" {
				if peers, ok := item.([]any); ok {
					for i, peer := range peers {
						if str, ok := peer.(string); ok && ids[str] != "" {
							peers[i] = ids[str]
						}
					}
				}
			} else if err := normalizeShape(s, item, ids, colors); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range x {
			if err := normalizeShape(s, item, ids, colors); err != nil {
				return err
			}
		}
	}
	return nil
}
