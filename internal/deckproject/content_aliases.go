package deckproject

import (
	"fmt"
	"sort"
	"strings"
)

func contentPointerParts(pointer string) ([]string, error) {
	if !strings.HasPrefix(pointer, "/") || len(pointer) > 2048 {
		return nil, fmt.Errorf("binding target must be a nonempty JSON pointer into values")
	}
	parts := strings.Split(pointer[1:], "/")
	if len(parts) > 30 {
		return nil, fmt.Errorf("binding target exceeds 30 levels")
	}
	for i, part := range parts {
		for j := 0; j < len(part); j++ {
			if part[j] == '~' {
				if j+1 >= len(part) || part[j+1] != '0' && part[j+1] != '1' {
					return nil, fmt.Errorf("invalid JSON pointer escape in %q", pointer)
				}
				j++
			}
		}
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
	}
	return parts, nil
}

func (p *Project) expandContentAliases(slide map[string]any, path string) error {
	bindings, hasBindings := slide["bindings"]
	content, hasContent := slide["content"]
	if !hasBindings && !hasContent {
		return nil
	}
	aliases, ok := bindings.(map[string]any)
	if !hasBindings || !ok || len(aliases) == 0 {
		return p.fail(path+"/bindings", "content requires a nonempty bindings mapping")
	}
	copy, ok := content.(map[string]any)
	if !hasContent || !ok {
		return p.fail(path+"/content", "content must be a mapping consumed by the declared bindings")
	}
	values, exists := slide["values"]
	if !exists {
		values = map[string]any{}
		slide["values"] = values
	}
	object, ok := values.(map[string]any)
	if !ok {
		return p.fail(path+"/values", "values must be a mapping")
	}
	names := make([]string, 0, len(aliases))
	for alias := range aliases {
		names = append(names, alias)
	}
	sort.Strings(names)
	targets := map[string]bool{}
	consumed := map[string]bool{}
	for _, alias := range names {
		rawPointer := aliases[alias]
		pointer, ok := rawPointer.(string)
		if !ok || !stableID.MatchString(alias) && !strings.HasPrefix(alias, "/") {
			return p.fail(path+"/bindings/"+escape(alias), "binding requires a stable alias or content JSON pointer and a values JSON pointer string")
		}
		contentPointer := alias
		if !strings.HasPrefix(alias, "/") {
			contentPointer = "/" + escape(alias)
		}
		if _, err := contentPointerParts(contentPointer); err != nil {
			return p.fail(path+"/bindings/"+escape(alias), "%v", err)
		}
		value, err := lookupPointer(copy, contentPointer)
		if err != nil {
			return p.fail(path+"/bindings/"+escape(alias), "missing content value at %s", contentPointer)
		}
		for previous := range consumed {
			if contentPointer == previous || strings.HasPrefix(contentPointer, previous+"/") || strings.HasPrefix(previous, contentPointer+"/") {
				return p.fail(path+"/bindings/"+escape(alias), "content binding sources must be distinct and cannot overlap")
			}
		}
		consumed[contentPointer] = true
		parts, err := contentPointerParts(pointer)
		if err != nil {
			return p.fail(path+"/bindings/"+escape(alias), "%v", err)
		}
		normalized := ""
		for _, part := range parts {
			normalized += "/" + escape(part)
		}
		for target := range targets {
			if normalized == target || strings.HasPrefix(normalized, target+"/") || strings.HasPrefix(target, normalized+"/") {
				return p.fail(path+"/bindings/"+escape(alias), "binding targets must be distinct and cannot overlap")
			}
		}
		targets[normalized] = true
		current := object
		for i, part := range parts {
			if i == len(parts)-1 {
				if _, exists := current[part]; exists {
					return p.fail(path+"/bindings/"+escape(alias), "binding target %s collides with values or another binding", pointer)
				}
				current[part] = value
				break
			}
			child, exists := current[part]
			if !exists {
				child = map[string]any{}
				current[part] = child
			}
			next, ok := child.(map[string]any)
			if !ok {
				return p.fail(path+"/bindings/"+escape(alias), "binding target %s crosses a non-mapping value", pointer)
			}
			current = next
		}
		p.Positions[path+"/values"+pointer] = p.Positions[path+"/content"+contentPointer]
		p.positionFiles[path+"/values"+pointer] = p.positionFiles[path+"/content"+contentPointer]
	}
	var unused func(any, string) error
	unused = func(value any, pointer string) error {
		for bound := range consumed {
			if pointer == bound || strings.HasPrefix(pointer, bound+"/") {
				return nil
			}
		}
		switch v := value.(type) {
		case map[string]any:
			if len(v) > 0 {
				for key, child := range v {
					if err := unused(child, pointer+"/"+escape(key)); err != nil {
						return err
					}
				}
				return nil
			}
		case []any:
			if len(v) > 0 {
				for i, child := range v {
					if err := unused(child, fmt.Sprintf("%s/%d", pointer, i)); err != nil {
						return err
					}
				}
				return nil
			}
		}
		return p.fail(path+"/content"+pointer, "unused content value; every authored content leaf needs an explicit binding")
	}
	if err := unused(copy, ""); err != nil {
		return err
	}
	delete(slide, "content")
	delete(slide, "bindings")
	return nil
}
