package deckproject

import (
	"encoding/json"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// sharedAssetValues resolves only image fields declared by the shared contract.
// Business strings and the authored document remain unchanged. Existing registry
// keys win over colliding project IDs; project:<id> explicitly selects the latter.
func sharedAssetValues(values map[string]any, def wmdesign.LibraryTemplate, keys map[string]string, registry map[string]wmdesign.PrimitiveAssetReference) (map[string]any, error) {
	var resolved map[string]any
	if err := json.Unmarshal(canonical(values), &resolved); err != nil {
		return nil, err
	}
	slots, _ := resolved["slots"].(map[string]any)
	var source map[string]any
	if err := json.Unmarshal(def.RawSlide, &source); err != nil {
		// Typed templates without projected source slots have no image fields.
		if len(def.Slots) == 0 {
			return resolved, nil
		}
		return nil, err
	}
	for _, slot := range def.Slots {
		if slot.Kind != "string" {
			continue
		}
		parts, err := contentPointerParts(slot.SourcePointer)
		if err != nil || len(parts) < 2 || !sharedImagePointer(source, parts) {
			continue
		}
		value, ok := slots[slot.Name].(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(value, "project:") {
			if key, declared := keys[strings.TrimPrefix(value, "project:")]; declared {
				slots[slot.Name] = key
			}
			continue
		}
		if _, registered := registry[value]; registered {
			continue
		}
		if key, declared := keys[value]; declared {
			slots[slot.Name] = key
		}
	}
	return resolved, nil
}

// Qualify the source schema as well as the field name: a table column named
// "photo" or "src" is business copy, not a media registry binding.
func sharedImagePointer(source map[string]any, parts []string) bool {
	if len(parts) < 2 {
		return false
	}
	field := parts[len(parts)-1]
	parentPointer := ""
	for _, part := range parts[:len(parts)-1] {
		parentPointer += "/" + escape(part)
	}
	raw, err := lookupPointer(source, parentPointer)
	if err != nil {
		return false
	}
	parent, _ := raw.(map[string]any)
	typ, _ := parent["type"].(string)
	switch typ {
	case "imageframe", "square", "person":
		return field == "photo"
	case "art", "logoslot":
		return field == "src"
	case "thumbnail":
		return field == "photo" || field == "src"
	}
	if len(parts) < 3 {
		return false
	}
	container := parts[len(parts)-2]
	grandparentPointer := strings.TrimSuffix(parentPointer, "/"+escape(container))
	raw, err = lookupPointer(source, grandparentPointer)
	if err != nil {
		return false
	}
	grandparent, _ := raw.(map[string]any)
	if grandparent["type"] != "card" {
		return false
	}
	return container == "media" && field == "src" || container == "bio" && field == "photo" || container == "badge" && field == "src"
}
