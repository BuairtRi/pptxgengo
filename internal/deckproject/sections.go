package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

type SectionView struct {
	wmdesign.SectionSpec
	SlideIDs        []string `json:"slide_ids"`
	DividerTemplate string   `json:"divider_template,omitempty"`
	DividerTitle    string   `json:"divider_title,omitempty"`
}

func ListSections(p *Project) []SectionView {
	result := make([]SectionView, 0, len(p.Document.Sections))
	positions := map[string]int{}
	for i, s := range p.Document.Slides {
		positions[s.ID] = i
	}
	for i, s := range p.Document.Sections {
		start, end := positions[s.BeforeSlideID], len(p.Document.Slides)
		if i+1 < len(p.Document.Sections) {
			end = positions[p.Document.Sections[i+1].BeforeSlideID]
		}
		view := SectionView{SectionSpec: s, SlideIDs: []string{}}
		for _, slide := range p.Document.Slides[start:end] {
			view.SlideIDs = append(view.SlideIDs, slide.ID)
		}
		anchor := p.Document.Slides[start]
		if anchor.Template.Scope == "shared" && strings.HasPrefix(anchor.Template.ID, "divider/") {
			view.DividerTemplate = anchor.Template.ID
			ptr := map[string]string{"divider/panel-edge": "node05.text", "divider/panel-photo": "node03.text", "divider/full-photo": "node04.text"}[anchor.Template.ID]
			if slots, ok := anchor.Values["slots"].(map[string]any); ok {
				view.DividerTitle, _ = slots[ptr].(string)
			}
		}
		result = append(result, view)
	}
	return result
}

type SectionChange struct {
	Operation      string        `json:"operation"`
	SectionID      string        `json:"section_id"`
	BeforeSHA256   string        `json:"before_sha256"`
	AfterSHA256    string        `json:"after_sha256"`
	Decision       string        `json:"decision"`
	AddedOpeningID string        `json:"added_opening_id,omitempty"`
	DividerSlideID string        `json:"divider_slide_id,omitempty"`
	Sections       []SectionView `json:"sections"`
}
type SectionAddOptions struct {
	ID, Title, BeforeSlideID              string
	Divider, DividerSlideID, DividerPhoto string
	DividerValues                         map[string]any
	Bundle, Engine                        string
}

func AddSection(p *Project, o SectionAddOptions) (SectionChange, error) {
	sections := append([]wmdesign.SectionSpec(nil), p.Document.Sections...)
	change := SectionChange{Operation: "section-add", SectionID: o.ID}
	position := -1
	for i, s := range p.Document.Slides {
		if s.ID == o.BeforeSlideID {
			position = i
		}
	}
	if position < 0 {
		return change, fmt.Errorf("unknown before slide ID %q", o.BeforeSlideID)
	}
	if len(sections) == 0 && position > 0 {
		used := map[string]bool{o.ID: true}
		for _, s := range sections {
			used[s.ID] = true
		}
		id := "opening"
		for i := 2; used[id]; i++ {
			id = fmt.Sprintf("opening-%d", i)
		}
		title := "Opening"
		if strings.EqualFold(o.Title, title) {
			title = "Opening slides"
		}
		sections = append(sections, wmdesign.SectionSpec{ID: id, Title: title, BeforeSlideID: p.Document.Slides[0].ID})
		change.AddedOpeningID = id
	}
	var divider *Slide
	if o.Divider != "" {
		if o.DividerSlideID == "" {
			o.DividerSlideID = o.ID + "-divider"
		}
		if !stableID.MatchString(o.DividerSlideID) {
			return change, fmt.Errorf("invalid divider slide ID")
		}
		for _, s := range p.Document.Slides {
			if s.ID == o.DividerSlideID {
				return change, fmt.Errorf("divider slide ID already exists")
			}
		}
		ordinal := 1
		for _, s := range sections {
			for j, slide := range p.Document.Slides {
				if slide.ID == s.BeforeSlideID && j < position {
					ordinal++
				}
			}
		}
		values, e := dividerValues(o, ordinal)
		if e != nil {
			return change, e
		}
		divider = &Slide{ID: o.DividerSlideID, ContentKind: "supplied_content", Template: Reference{Scope: "shared", ID: o.Divider}, Values: values}
		change.DividerSlideID = divider.ID
	} else if o.DividerSlideID != "" || o.DividerPhoto != "" || o.DividerValues != nil {
		return change, fmt.Errorf("divider options require --divider")
	}
	anchor := o.BeforeSlideID
	if divider != nil {
		anchor = divider.ID
	}
	sections = append(sections, wmdesign.SectionSpec{ID: o.ID, Title: o.Title, BeforeSlideID: anchor})
	positions := map[string]int{}
	for i, s := range p.Document.Slides {
		positions[s.ID] = i*2 + 1
	}
	if divider != nil {
		positions[divider.ID] = position * 2
	}
	sort.SliceStable(sections, func(i, j int) bool {
		return positions[sections[i].BeforeSlideID] < positions[sections[j].BeforeSlideID]
	})
	return mutateSections(p, change, sections, divider, position, o.Bundle, o.Engine)
}
func RenameSection(p *Project, id, title string) (SectionChange, error) {
	sections := append([]wmdesign.SectionSpec(nil), p.Document.Sections...)
	found := false
	for i, s := range sections {
		if s.ID == id {
			sections[i].Title = title
			found = true
		}
	}
	if !found {
		return SectionChange{}, fmt.Errorf("unknown section ID %q", id)
	}
	return mutateSections(p, SectionChange{Operation: "section-rename", SectionID: id}, sections, nil, 0, "", "")
}
func RemoveSection(p *Project, id string) (SectionChange, error) {
	sections := []wmdesign.SectionSpec{}
	found := -1
	for i, s := range p.Document.Sections {
		if s.ID == id {
			found = i
		} else {
			sections = append(sections, s)
		}
	}
	if found < 0 {
		return SectionChange{}, fmt.Errorf("unknown section ID %q", id)
	}
	if found == 0 && len(sections) > 0 {
		sections[0].BeforeSlideID = p.Document.Slides[0].ID
	}
	return mutateSections(p, SectionChange{Operation: "section-remove", SectionID: id}, sections, nil, 0, "", "")
}
func dividerValues(o SectionAddOptions, ordinal int) (map[string]any, error) {
	if !strings.HasPrefix(o.Divider, "divider/") || o.Bundle == "" {
		return nil, fmt.Errorf("divider requires an actual divider/... template and a pinned bundle")
	}
	catalog, e := wmdesign.LibraryCatalog(o.Bundle, "")
	if e != nil {
		return nil, e
	}
	var definition *wmdesign.LibraryTemplate
	for i := range catalog {
		if catalog[i].Key == o.Divider {
			definition = &catalog[i]
			break
		}
	}
	if definition == nil {
		return nil, fmt.Errorf("unknown divider template %q", o.Divider)
	}
	if o.DividerValues != nil {
		if o.DividerPhoto != "" {
			return nil, fmt.Errorf("--divider-photo cannot accompany --divider-values")
		}
		return o.DividerValues, nil
	}
	if o.DividerPhoto == "" {
		return nil, fmt.Errorf("automatic divider binding requires caller-supplied --divider-photo; other variants require --divider-values")
	}
	number := fmt.Sprintf("%02d", ordinal)
	recipes := map[string]map[string]string{
		"divider/panel-edge":  {"/body/0/text": "", "/body/2/text": "Section", "/body/3/text": number, "/body/4/text": o.Title, "/body/5/photo": o.DividerPhoto},
		"divider/panel-photo": {"/body/0/text": "", "/body/1/text": "[[" + number + "]]", "/body/2/text": o.Title, "/body/3/photo": o.DividerPhoto},
		"divider/full-photo":  {"/body/0/photo": o.DividerPhoto, "/body/1/text": "", "/body/2/text": number, "/body/3/text": o.Title},
	}
	recipe, ok := recipes[o.Divider]
	if !ok {
		return nil, fmt.Errorf("divider %s requires explicit --divider-values for its closed contract", o.Divider)
	}
	if len(recipe) != len(definition.Slots) || len(definition.Arrays) != 0 {
		return nil, fmt.Errorf("divider contract changed; explicit values required")
	}
	slots := map[string]any{}
	for _, slot := range definition.Slots {
		value, ok := recipe[slot.SourcePointer]
		if !ok {
			return nil, fmt.Errorf("divider source topology changed")
		}
		slots[slot.Name] = value
	}
	return map[string]any{"slots": slots}, nil
}
func yamlNode(v any) (*yaml.Node, error) {
	var tree any
	if e := json.Unmarshal(canonical(v), &tree); e != nil {
		return nil, e
	}
	var node yaml.Node
	if e := node.Encode(tree); e != nil {
		return nil, e
	}
	return &node, nil
}
func mappingNode(n *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}
func mutateSections(p *Project, change SectionChange, sections []wmdesign.SectionSpec, divider *Slide, position int, bundle, engine string) (SectionChange, error) {
	additional := map[string][]byte{}
	var doc yaml.Node
	if e := yaml.Unmarshal(p.Raw, &doc); e != nil {
		return change, e
	}
	root := doc.Content[0]
	replacement, e := yamlNode(sections)
	if e != nil {
		return change, e
	}
	existing := mappingNode(root, "sections")
	if existing != nil {
		old := map[string]*yaml.Node{}
		for _, node := range existing.Content {
			if id := mappingNode(node, "id"); id != nil {
				old[id.Value] = node
			}
		}
		for i, section := range sections {
			if node := old[section.ID]; node != nil {
				for _, field := range []string{"title", "before_slide_id"} {
					value := section.Title
					if field == "before_slide_id" {
						value = section.BeforeSlideID
					}
					if scalar := mappingNode(node, field); scalar != nil {
						scalar.Value = value
						scalar.Tag = "!!str"
					}
				}
				replacement.Content[i] = node
			}
		}
		replacement.HeadComment = existing.HeadComment
		replacement.LineComment = existing.LineComment
		replacement.FootComment = existing.FootComment
		*existing = *replacement
	} else {
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "sections"}, replacement)
	}
	if divider != nil {
		slides := mappingNode(root, "slides")
		node, e := yamlNode(*divider)
		if e != nil {
			return change, e
		}
		if p.hasExternalSources() {
			relative := "slides/" + divider.ID + ".yaml"
			raw, err := encodeSourceYAML(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}})
			if err != nil {
				return change, err
			}
			additional[relative] = raw
			node = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: relative}
		}
		slides.Content = append(slides.Content, nil)
		copy(slides.Content[position+1:], slides.Content[position:])
		slides.Content[position] = node
	}
	var output bytes.Buffer
	enc := yaml.NewEncoder(&output)
	enc.SetIndent(2)
	if e := enc.Encode(&doc); e != nil {
		return change, e
	}
	if e := enc.Close(); e != nil {
		return change, e
	}
	raw := output.Bytes()
	if p.hasExternalSources() || p.hasContentAliases() {
		additional[filepath.Base(p.SourcePath)] = raw
		return commitSourceSections(p, change, additional, divider != nil, bundle, engine)
	}
	change.BeforeSHA256 = digest(p.Raw)
	change.AfterSHA256 = digest(raw)
	temp, e := SafePath(p.Root, ".deck-section-"+nonce()+".yaml")
	if e != nil {
		return change, e
	}
	stat, e := os.Stat(p.SourcePath)
	if e != nil {
		return change, e
	}
	if e = writeExclusive(temp, raw, stat.Mode().Perm()); e != nil {
		return change, e
	}
	defer os.Remove(temp)
	candidate, e := Load(temp)
	if e != nil {
		return change, e
	}
	if divider != nil {
		if engine == "" {
			engine = wmdesign.CandidateEngine
		}
		var compiled Compilation
		if _, _, lockErr := ReadLock(p); lockErr == nil {
			compiled, e = Check(candidate, bundle, engine)
		} else if os.IsNotExist(lockErr) {
			compiled, e = Compile(candidate, bundle, engine)
		} else {
			e = lockErr
		}
		if e != nil {
			return change, e
		}
		if _, _, e = wmdesign.BuildWithEngineAndAssets(bundle, "", compiled.Document, engine, compiled.Assets); e != nil {
			return change, e
		}
	}
	change.Sections = ListSections(candidate)
	lockPath, e := SafePath(p.Root, ".deck-section-mutation.lock")
	if e != nil {
		return change, e
	}
	if e = writeExclusive(lockPath, []byte(change.BeforeSHA256), 0600); e != nil {
		return change, fmt.Errorf("another section mutation is active: %w", e)
	}
	defer os.Remove(lockPath)
	current, e := os.ReadFile(p.SourcePath)
	if e != nil {
		return change, e
	}
	if !bytes.Equal(current, p.Raw) {
		return change, fmt.Errorf("source changed during mutation")
	}
	before, e := SafePath(p.Root, "decisions/sources/"+change.BeforeSHA256+".yaml")
	if e != nil {
		return change, e
	}
	if stored, e := os.ReadFile(before); e == nil {
		if !bytes.Equal(stored, p.Raw) {
			return change, fmt.Errorf("source predecessor snapshot drift")
		}
	} else if os.IsNotExist(e) {
		if e = writeExclusive(before, p.Raw, 0444); e != nil {
			return change, e
		}
	} else {
		return change, e
	}
	change.Decision = "decisions/" + change.Operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, e := SafePath(p.Root, change.Decision)
	if e != nil {
		return change, e
	}
	if e = writeJSON(decision, change); e != nil {
		return change, e
	}
	if e = os.Rename(temp, p.SourcePath); e != nil {
		os.Remove(decision)
		return change, e
	}
	return change, nil
}
