package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func (p *Project) hasExternalSources() bool { return len(p.SourceFiles) > 1 }

func (p *Project) hasContentAliases() bool {
	document, err := sourceYAML(p.Raw)
	if err != nil {
		return false
	}
	slides := mappingNode(document.Content[0], "slides")
	for _, slide := range slides.Content {
		if slide.Kind == yaml.MappingNode && mappingNode(slide, "content") != nil {
			return true
		}
	}
	return false
}

func extendedSlideEdits(edits map[string]SlideEdit) bool {
	for _, edit := range edits {
		if edit.Content != nil || edit.Bindings != nil {
			return true
		}
	}
	return false
}

func editSourceSlides(p *Project, edits map[string]SlideEdit, bundle, engine string) (EditReceipt, error) {
	r := EditReceipt{Operation: "edit-slides", SlideIDs: []string{}, BeforeSHA256: p.SourceHash(), Validation: "source_and_binding_checked_native_review_pending"}
	if len(edits) == 0 {
		return r, fmt.Errorf("slide patch must contain at least one stable slide ID")
	}
	main, err := sourceYAML(p.Raw)
	if err != nil {
		return r, err
	}
	slides, documents, err := authoredSlides(p, main)
	if err != nil {
		return r, err
	}
	changed := map[string]bool{}
	for _, slide := range p.Document.Slides {
		edit, selected := edits[slide.ID]
		if !selected {
			continue
		}
		if edit.Template == nil && edit.Values == nil && edit.Brief == nil && edit.Content == nil && edit.Bindings == nil {
			return r, fmt.Errorf("slide %s has no edit fields", slide.ID)
		}
		node := slides[slide.ID]
		if edit.Values != nil && edit.Content == nil && edit.Bindings == nil {
			removeMappingField(node, "content")
			removeMappingField(node, "bindings")
		}
		for _, field := range []struct {
			name  string
			value any
			used  bool
		}{
			{"template", edit.Template, edit.Template != nil}, {"values", edit.Values, edit.Values != nil},
			{"content", edit.Content, edit.Content != nil}, {"bindings", edit.Bindings, edit.Bindings != nil},
			{"brief", edit.Brief, edit.Brief != nil},
		} {
			if !field.used {
				continue
			}
			replacement, err := editYAMLNode(field.value)
			if err != nil {
				return r, err
			}
			replaceMappingField(node, field.name, replacement)
		}
		orderAuthoredSlide(node)
		relative := p.SlideFiles[slide.ID]
		if relative == "" {
			relative = filepath.Base(p.SourcePath)
		}
		changed[relative] = true
		r.SlideIDs = append(r.SlideIDs, slide.ID)
	}
	for id := range edits {
		if slides[id] == nil {
			return r, fmt.Errorf("unknown stable slide ID %s", id)
		}
	}
	changes := map[string][]byte{}
	for relative := range changed {
		raw, err := encodeSourceYAML(documents[relative])
		if err != nil {
			return r, err
		}
		changes[relative] = raw
	}
	r.Decision = "decisions/edit-slides-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, err := SafePath(p.Root, r.Decision)
	if err != nil {
		return r, err
	}
	_, err = commitSourceChanges(p, changes, func(candidate *Project) error {
		if _, err := Compile(candidate, bundle, engine); err != nil {
			return err
		}
		r.AfterSHA256 = candidate.SourceHash()
		return writeJSON(decision, r)
	})
	if err != nil {
		os.Remove(decision)
	}
	return r, err
}
