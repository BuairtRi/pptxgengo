package deckproject

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

type SplitOptions struct {
	Bundle string
	// StockEditor supplies explicit friendly aliases without changing the
	// selected shared contract. Nil keeps existing values readable and exact.
	StockEditor func(Slide, wmdesign.LibraryTemplate) (map[string]any, error)
}

type SplitReceipt struct {
	Operation      string            `json:"operation"`
	BeforeSHA256   string            `json:"before_sha256"`
	AfterSHA256    string            `json:"after_sha256"`
	SemanticSHA256 string            `json:"semantic_sha256"`
	Files          map[string]string `json:"files"`
	Preservation   string            `json:"preservation"`
}

// useBlockCollections makes JSON-compatible flow YAML readable as authored
// files. Preserve scalar quoting, literal copy and every comment; only mapping
// and sequence presentation changes. Empty collections still encode as {} / [].
func useBlockCollections(node *yaml.Node) {
	if node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode {
		node.Style &^= yaml.FlowStyle
	}
	for _, child := range node.Content {
		useBlockCollections(child)
	}
}

func Split(p *Project, options SplitOptions) (SplitReceipt, error) {
	r := SplitReceipt{Operation: "split-source", BeforeSHA256: p.SourceHash(), SemanticSHA256: digest(p.Canonical), Files: map[string]string{}, Preservation: "expanded_canonical_source_unchanged; existing_toolchain_lock_preserved"}
	main, err := sourceYAML(p.Raw)
	if err != nil {
		return r, err
	}
	objects, documents, err := authoredSlides(p, main)
	if err != nil {
		return r, err
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	if options.StockEditor != nil {
		if options.Bundle == "" {
			return r, fmt.Errorf("friendly stock splitting requires --bundle")
		}
		catalog, err := wmdesign.LibraryCatalog(options.Bundle, "")
		if err != nil {
			return r, err
		}
		for _, def := range catalog {
			defs[def.Key] = def
		}
	}
	changes := map[string][]byte{}
	slides := mappingNode(main.Content[0], "slides")
	for i, slide := range p.Document.Slides {
		node := objects[slide.ID]
		relative := p.SlideFiles[slide.ID]
		if relative == "" {
			relative = "slides/" + slide.ID + ".yaml"
		}
		if options.StockEditor != nil && slide.Template.Scope == "shared" && mappingNode(node, "content") == nil {
			def, exists := defs[slide.Template.ID]
			if !exists {
				return r, fmt.Errorf("unknown shared template %s", slide.Template.ID)
			}
			editable, err := options.StockEditor(slide, def)
			if err != nil {
				return r, err
			}
			for _, field := range []string{"values", "content", "bindings"} {
				if value, exists := editable[field]; exists {
					replacement, err := editYAMLNode(value)
					if err != nil {
						return r, err
					}
					replaceMappingField(node, field, replacement)
				} else {
					removeMappingField(node, field)
				}
			}
		}
		if notes := mappingNode(node, "notes"); notes != nil {
			notesRelative := "notes/" + slide.ID + ".md"
			changes[notesRelative] = []byte(slide.Notes)
			removeMappingField(node, "notes")
			replaceMappingField(node, "notes_file", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: notesRelative})
		}
		orderAuthoredSlide(node)
		document := documents[relative]
		if document == nil {
			document = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}}
		}
		useBlockCollections(document)
		raw, err := encodeSourceYAML(document)
		if err != nil {
			return r, err
		}
		changes[relative] = raw
		slides.Content[i] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: relative, HeadComment: slides.Content[i].HeadComment, LineComment: slides.Content[i].LineComment, FootComment: slides.Content[i].FootComment}
	}
	if templates := mappingNode(main.Content[0], "local_templates"); templates != nil {
		for i := 0; i+1 < len(templates.Content); i += 2 {
			if templates.Content[i+1].Kind != yaml.MappingNode {
				continue
			}
			id, node := templates.Content[i].Value, templates.Content[i+1]
			relative := "slides/templates/" + id + ".yaml"
			useBlockCollections(node)
			raw, err := encodeSourceYAML(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}})
			if err != nil {
				return r, err
			}
			changes[relative] = raw
			templates.Content[i+1] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: relative, HeadComment: node.HeadComment, LineComment: node.LineComment, FootComment: node.FootComment}
		}
	}
	// A repeated split also normalizes previously referenced definitions, which
	// may themselves have been emitted from legacy JSON-form source.
	for _, relative := range p.TemplateFiles {
		document, err := sourceYAML(p.SourceFiles[relative])
		if err != nil {
			return r, err
		}
		useBlockCollections(document)
		raw, err := encodeSourceYAML(document)
		if err != nil {
			return r, err
		}
		changes[relative] = raw
	}
	useBlockCollections(main)
	raw, err := encodeSourceYAML(main)
	if err != nil {
		return r, err
	}
	changes[filepath.Base(p.SourcePath)] = raw
	next, err := commitSourceChanges(p, changes, func(candidate *Project) error {
		if !bytes.Equal(candidate.Canonical, p.Canonical) {
			return fmt.Errorf("split would change the expanded canonical source")
		}
		return nil
	})
	if err != nil {
		return r, err
	}
	r.AfterSHA256 = next.SourceHash()
	keys := []string{}
	for relative := range next.SourceFiles {
		keys = append(keys, relative)
	}
	sort.Strings(keys)
	for _, relative := range keys {
		r.Files[relative] = digest(next.SourceFiles[relative])
	}
	return r, nil
}
