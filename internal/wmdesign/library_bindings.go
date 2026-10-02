package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// LibraryValues is content-only. Its exact names, scalar types and array counts
// are closed by the selected source-pinned LibraryTemplate, not caller paths.
type LibraryValues struct {
	Slots map[string]json.RawMessage `json:"slots"`
	Keys  map[string][]string        `json:"keys"`
}

func setLibraryContent(root any, pointer string, value any) error {
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	parent := root
	for i, part := range parts {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		last := i == len(parts)-1
		switch object := parent.(type) {
		case map[string]any:
			old, exists := object[part]
			if !exists {
				return fmt.Errorf("library.binding_target_missing: %s", pointer)
			}
			if last {
				object[part] = value
				return nil
			}
			parent = old
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(object) {
				return fmt.Errorf("library.binding_index_missing: %s", pointer)
			}
			if last {
				object[index] = value
				return nil
			}
			parent = object[index]
		default:
			return fmt.Errorf("library.binding_target_not_container: %s", pointer)
		}
	}
	return fmt.Errorf("library.binding_invalid_target: %s", pointer)
}

func bindLibraryTemplate(def LibraryTemplate, bound BoundSlide) (SlideSpec, TemplateSlideRecord, error) {
	var values LibraryValues
	if err := bindingStrictDecode(bound.Values, &values); err != nil {
		return SlideSpec{}, TemplateSlideRecord{}, err
	}
	if len(values.Slots) != len(def.Slots) || len(values.Keys) != len(def.Arrays) {
		return SlideSpec{}, TemplateSlideRecord{}, fmt.Errorf("library.binding_required_slot_or_array_count: %s needs %d slots and %d arrays", def.Key, len(def.Slots), len(def.Arrays))
	}
	obj, err := libraryObject(def.RawSlide)
	if err != nil {
		return SlideSpec{}, TemplateSlideRecord{}, err
	}
	record := TemplateSlideRecord{SlideID: bound.ID, Template: def.Key, Contract: LibraryBindingsContract, SourceFile: def.SourceFile, SourceSHA256: def.SourceSHA256, ContentKind: bound.ContentKind, Identities: def.Identities}
	for _, slot := range def.Slots {
		raw, ok := values.Slots[slot.Name]
		if !ok {
			return SlideSpec{}, record, fmt.Errorf("library.binding_missing_slot: %s", slot.Name)
		}
		if string(raw) == "null" {
			return SlideSpec{}, record, fmt.Errorf("library.binding_null_slot: %s", slot.Name)
		}
		var value any
		switch slot.Kind {
		case "string":
			var text string
			if err = bindingStrictDecode(raw, &text); err != nil {
				return SlideSpec{}, record, fmt.Errorf("%s: %w", slot.Name, err)
			}
			if !slot.AllowEmpty && strings.TrimSpace(text) == "" {
				return SlideSpec{}, record, fmt.Errorf("library.binding_empty_slot: %s", slot.Name)
			}
			value = text
		case "number":
			var number float64
			if err = bindingStrictDecode(raw, &number); err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
				return SlideSpec{}, record, fmt.Errorf("library.binding_invalid_number: %s", slot.Name)
			}
			value = number
		case "boolean":
			var boolean bool
			if err = bindingStrictDecode(raw, &boolean); err != nil {
				return SlideSpec{}, record, fmt.Errorf("library.binding_invalid_boolean: %s: %w", slot.Name, err)
			}
			value = boolean
		default:
			return SlideSpec{}, record, fmt.Errorf("library.binding_unknown_slot_kind: %s", slot.Kind)
		}
		if err = setLibraryContent(obj, slot.SourcePointer, value); err != nil {
			return SlideSpec{}, record, err
		}
		record.Assignments = append(record.Assignments, TemplateSlotAssignment{Slot: slot.Name, TargetID: strings.Split(slot.Name, ".")[0], Property: slot.Name, ValueKind: slot.Kind, SourcePointer: slot.SourcePointer, Value: raw})
	}
	keyOverlay := map[string][]string{}
	for _, array := range def.Arrays {
		keys, ok := values.Keys[array.Name]
		if !ok || len(keys) != array.Count {
			return SlideSpec{}, record, fmt.Errorf("library.binding_array_keys: %s needs %d", array.Name, array.Count)
		}
		seen := map[string]bool{}
		for i, key := range keys {
			if !validPartKey(key) || seen[key] {
				return SlideSpec{}, record, fmt.Errorf("library.binding_invalid_or_duplicate_key: %s/%s", array.Name, key)
			}
			seen[key] = true
			record.CardKeys = append(record.CardKeys, TemplateCardKey{Key: key, Ordinal: i + 1, TargetID: array.Name + "." + key, SourcePointer: array.SourcePointer})
		}
		keyOverlay[array.SourcePointer] = keys
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return SlideSpec{}, record, err
	}
	slide, err := compileLibrarySlide(raw, keyOverlay)
	if err != nil {
		return SlideSpec{}, record, err
	}
	slide.ID, slide.ContentKind, slide.TemplateBinding = bound.ID, bound.ContentKind, &record
	if err := applyLibraryRefinements(def.Key, &slide); err != nil {
		return SlideSpec{}, record, err
	}
	return slide, record, nil
}

func libraryExampleValues(def LibraryTemplate) LibraryValues {
	values := LibraryValues{Slots: map[string]json.RawMessage{}, Keys: map[string][]string{}}
	for _, slot := range def.Slots {
		values.Slots[slot.Name] = slot.Example
	}
	for _, array := range def.Arrays {
		keys := make([]string, array.Count)
		for i := range keys {
			keys[i] = fmt.Sprintf("example-%03d", i+1)
		}
		values.Keys[array.Name] = keys
	}
	return values
}

// LibraryReference uses explicit synthetic source examples through the same
// closed content projection as callers. It never serves as a missing-value default.
func LibraryReference(bundle, override, family string, year int) (BoundDocument, error) {
	catalog, err := LibraryCatalog(bundle, override)
	if err != nil {
		return BoundDocument{}, err
	}
	doc := BoundDocument{Schema: BoundDocumentSchema, Year: year}
	legacy := TemplateReference(year)
	for _, def := range catalog {
		if family != "" && def.Family != family {
			continue
		}
		if legacyTemplate(def.Key) {
			for _, slide := range legacy.Slides {
				if slide.Template == def.Key {
					slide.ID = fmt.Sprintf("library-%03d", len(doc.Slides)+1)
					doc.Slides = append(doc.Slides, slide)
					break
				}
			}
			continue
		}
		raw, e := json.Marshal(libraryExampleValues(def))
		if e != nil {
			return BoundDocument{}, e
		}
		doc.Slides = append(doc.Slides, BoundSlide{ID: fmt.Sprintf("library-%03d", len(doc.Slides)+1), Template: def.Key, ContentKind: "synthetic_example", Values: raw})
	}
	if len(doc.Slides) == 0 {
		return BoundDocument{}, fmt.Errorf("library.unknown_or_empty_family: %s", family)
	}
	return doc, nil
}

func LibrarySourceReference(bundle, override, family string, year int) (Document, error) {
	catalog, err := LibraryCatalog(bundle, override)
	if err != nil {
		return Document{}, err
	}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	for _, def := range catalog {
		if family != "" && def.Family != family {
			continue
		}
		slide, e := compileLibrarySlide(def.RawSlide, nil)
		if e != nil {
			return Document{}, e
		}
		slide.ID = fmt.Sprintf("source-%03d", len(doc.Slides)+1)
		slide.ContentKind = "synthetic_example"
		slide.TemplateBinding = &TemplateSlideRecord{SlideID: slide.ID, Template: def.Key, Contract: LibraryBindingsContract, SourceFile: def.SourceFile, SourceSHA256: def.SourceSHA256, ContentKind: "synthetic_example", Identities: def.Identities}
		if e := applyLibraryRefinements(def.Key, &slide); e != nil {
			return Document{}, e
		}
		doc.Slides = append(doc.Slides, slide)
	}
	if len(doc.Slides) == 0 {
		return Document{}, fmt.Errorf("library.unknown_or_empty_family: %s", family)
	}
	return doc, nil
}
