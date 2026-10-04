package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

type SlideOperation struct {
	Action, ID, Before, After string
	Source                    []byte
	Reanchor                  bool
	Bundle, Engine            string
}

type SlideOperationReceipt struct {
	Operation       string        `json:"operation"`
	SlideID         string        `json:"slide_id"`
	BeforeSHA256    string        `json:"before_sha256"`
	AfterSHA256     string        `json:"after_sha256"`
	Decision        string        `json:"decision"`
	Order           []string      `json:"slide_order"`
	Sections        []SectionView `json:"sections"`
	SectionChanges  []string      `json:"section_changes"`
	RetainedFiles   []string      `json:"retained_unreferenced_files,omitempty"`
	EditorialAction string        `json:"editorial_action,omitempty"`
}

// OperateSlide changes ordering or visibility without rewriting unselected slide
// files. Removed sources remain recoverable and are explicitly listed in receipts.
func OperateSlide(p *Project, o SlideOperation) (SlideOperationReceipt, error) {
	r := SlideOperationReceipt{Operation: "slide-" + o.Action, SlideID: o.ID, BeforeSHA256: p.SourceHash(), SectionChanges: []string{}}
	if o.Action == "add" || o.Action == "remove" {
		log, err := compositionPath(p)
		if err != nil {
			return r, err
		}
		if log != "" {
			r.EditorialAction = "Update " + log + " for the added/removed slide before project check/build; composition rationale is authored separately."
		}
	}
	if !stableID.MatchString(o.ID) {
		return r, fmt.Errorf("slide operation requires a valid stable --id")
	}
	if o.Before != "" && o.After != "" {
		return r, fmt.Errorf("choose --before or --after")
	}
	if (o.Action != "add" && o.Action != "move") && (o.Before != "" || o.After != "" || len(o.Source) > 0 || o.Reanchor) {
		return r, fmt.Errorf("position/source/reanchor options require add or move")
	}
	main, err := sourceYAML(p.Raw)
	if err != nil {
		return r, err
	}
	nodes, documents, err := authoredSlides(p, main)
	if err != nil {
		return r, err
	}
	sequence := mappingNode(main.Content[0], "slides")
	index := -1
	for i, s := range p.Document.Slides {
		if s.ID == o.ID {
			index = i
		}
	}
	sections := append([]wmdesign.SectionSpec(nil), p.Document.Sections...)
	changes := map[string][]byte{}
	order := make([]string, len(p.Document.Slides))
	for i, s := range p.Document.Slides {
		order[i] = s.ID
	}
	anchor := -1
	for i, s := range sections {
		if s.BeforeSlideID == o.ID {
			anchor = i
		}
	}
	reanchor := func() {
		if anchor < 0 {
			return
		}
		end := len(order)
		if anchor+1 < len(sections) {
			for i, id := range order {
				if id == sections[anchor+1].BeforeSlideID {
					end = i
				}
			}
		}
		if index+1 < end {
			sections[anchor].BeforeSlideID = order[index+1]
			r.SectionChanges = append(r.SectionChanges, "Reanchored "+sections[anchor].ID+" to "+order[index+1])
		} else {
			r.SectionChanges = append(r.SectionChanges, "Removed empty section "+sections[anchor].ID)
			sections = append(sections[:anchor], sections[anchor+1:]...)
		}
	}
	switch o.Action {
	case "hide", "show":
		if index < 0 {
			return r, fmt.Errorf("unknown slide ID %q", o.ID)
		}
		if o.Action == "hide" {
			replaceMappingField(nodes[o.ID], "hidden", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"})
		} else {
			removeMappingField(nodes[o.ID], "hidden")
		}
		relative := p.SlideFiles[o.ID]
		if relative == "" {
			relative = filepath.Base(p.SourcePath)
		}
		changes[relative], err = encodeSourceYAML(documents[relative])
		if err != nil {
			return r, err
		}
	case "remove", "move":
		if index < 0 {
			return r, fmt.Errorf("unknown slide ID %q", o.ID)
		}
		if len(order) == 1 && o.Action == "remove" {
			return r, fmt.Errorf("cannot remove the final slide")
		}
		if o.Action == "move" {
			if len(o.Source) > 0 {
				return r, fmt.Errorf("move does not accept slide source")
			}
			if o.Before == "" && o.After == "" {
				return r, fmt.Errorf("move requires --before or --after")
			}
			if o.Before == o.ID || o.After == o.ID {
				return r, fmt.Errorf("cannot position a slide relative to itself")
			}
			if anchor >= 0 && !o.Reanchor {
				return r, fmt.Errorf("slide %s anchors section %s; use --reanchor to leave its section in place", o.ID, sections[anchor].ID)
			}
		}
		reanchor()
		node := sequence.Content[index]
		sequence.Content = append(sequence.Content[:index], sequence.Content[index+1:]...)
		order = append(order[:index], order[index+1:]...)
		if o.Action == "move" {
			position, e := slidePosition(order, o.Before, o.After)
			if e != nil {
				return r, e
			}
			sequence.Content = insertSlideNode(sequence.Content, position, node)
			order = insertSlideID(order, position, o.ID)
		} else {
			for _, path := range []string{p.SlideFiles[o.ID], p.NotesFiles[o.ID]} {
				if path != "" {
					r.RetainedFiles = append(r.RetainedFiles, path)
				}
			}
		}
	case "add":
		if index >= 0 {
			return r, fmt.Errorf("slide ID %s already exists", o.ID)
		}
		if len(o.Source) == 0 {
			return r, fmt.Errorf("add requires a supplied slide file")
		}
		if o.Reanchor {
			return r, fmt.Errorf("add does not accept --reanchor")
		}
		// The candidate loader performs strict field, alias and content validation.
		parser := &Project{Positions: map[string]Position{}, SourceFiles: map[string][]byte{}, positionFiles: map[string]string{}}
		parsed, e := parser.parseSource(o.Source, "new-slide.yaml", "")
		if e != nil {
			return r, e
		}
		object, ok := parsed.(map[string]any)
		if !ok || object["id"] != o.ID {
			return r, fmt.Errorf("slide file must contain id: %s", o.ID)
		}
		doc, e := sourceYAML(o.Source)
		if e != nil {
			return r, e
		}
		node := doc.Content[0]
		position, e := slidePosition(order, o.Before, o.After)
		if e != nil {
			return r, e
		}
		if p.hasExternalSources() {
			relative := "slides/" + o.ID + ".yaml"
			changes[relative] = o.Source
			node = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: relative}
		}
		sequence.Content = insertSlideNode(sequence.Content, position, node)
		order = insertSlideID(order, position, o.ID)
	default:
		return r, fmt.Errorf("unknown slide operation %q", o.Action)
	}
	if o.Action == "add" || o.Action == "move" || o.Action == "remove" {
		positions := map[string]int{}
		for i, id := range order {
			positions[id] = i
		}
		if len(sections) > 0 {
			sort.SliceStable(sections, func(i, j int) bool {
				return positions[sections[i].BeforeSlideID] < positions[sections[j].BeforeSlideID]
			})
			if sections[0].BeforeSlideID != order[0] {
				sections[0].BeforeSlideID = order[0]
				r.SectionChanges = append(r.SectionChanges, "Extended first section to "+order[0])
			}
		}
		sectionNode, e := yamlNode(sections)
		if e != nil {
			return r, e
		}
		if len(sections) > 0 {
			if previous := mappingNode(main.Content[0], "sections"); previous != nil {
				old := map[string]*yaml.Node{}
				for _, node := range previous.Content {
					if id := mappingNode(node, "id"); id != nil {
						old[id.Value] = node
					}
				}
				for i, section := range sections {
					if node := old[section.ID]; node != nil {
						replaceMappingField(node, "before_slide_id", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: section.BeforeSlideID})
						sectionNode.Content[i] = node
					}
				}
			}
			replaceMappingField(main.Content[0], "sections", sectionNode)
		} else {
			removeMappingField(main.Content[0], "sections")
		}
		changes[filepath.Base(p.SourcePath)], err = encodeSourceYAML(main)
		if err != nil {
			return r, err
		}
	}
	r.Decision = "decisions/" + r.Operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, err := SafePath(p.Root, r.Decision)
	if err != nil {
		return r, err
	}
	_, err = commitSourceChanges(p, changes, func(candidate *Project) error {
		if o.Action == "add" {
			if o.Bundle == "" {
				return fmt.Errorf("adding a slide requires the pinned bundle")
			}
			if _, e := Compile(candidate, o.Bundle, o.Engine); e != nil {
				return e
			}
		}
		r.AfterSHA256 = candidate.SourceHash()
		r.Order = order
		r.Sections = ListSections(candidate)
		return writeJSON(decision, r)
	})
	if err != nil {
		os.Remove(decision)
	}
	return r, err
}

func slidePosition(order []string, before, after string) (int, error) {
	if before == "" && after == "" {
		return len(order), nil
	}
	for i, id := range order {
		if id == before {
			return i, nil
		}
		if id == after {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("unknown position slide ID %q", before+after)
}
func insertSlideNode(nodes []*yaml.Node, index int, node *yaml.Node) []*yaml.Node {
	nodes = append(nodes, nil)
	copy(nodes[index+1:], nodes[index:])
	nodes[index] = node
	return nodes
}
func insertSlideID(ids []string, index int, id string) []string {
	ids = append(ids, "")
	copy(ids[index+1:], ids[index:])
	ids[index] = id
	return ids
}
