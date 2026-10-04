package deckproject

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

func sourceYAML(raw []byte) (*yaml.Node, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if len(document.Content) != 1 {
		return nil, fmt.Errorf("source requires one YAML mapping")
	}
	return &document, nil
}

func encodeSourceYAML(document *yaml.Node) ([]byte, error) {
	quoteLeadingNewlines(document)
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func replaceMappingField(node *yaml.Node, key string, value *yaml.Node) {
	if previous := mappingNode(node, key); previous != nil {
		value.HeadComment, value.LineComment, value.FootComment = previous.HeadComment, previous.LineComment, previous.FootComment
		*previous = *value
	} else {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
	}
}

func removeMappingField(node *yaml.Node, key string) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content = append(node.Content[:i], node.Content[i+2:]...)
			return
		}
	}
}

func orderMappingFields(node *yaml.Node, preferred ...string) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	ordered := []*yaml.Node{}
	used := map[string]bool{}
	for _, key := range preferred {
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				ordered = append(ordered, node.Content[i], node.Content[i+1])
				used[key] = true
				break
			}
		}
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if !used[node.Content[i].Value] {
			ordered = append(ordered, node.Content[i], node.Content[i+1])
		}
	}
	node.Content = ordered
}

func orderAuthoredSlide(node *yaml.Node) {
	orderMappingFields(node, "id", "template", "content_kind", "hidden", "brief", "notes_file", "notes", "content", "values", "bindings", "evidence_refs")
	orderMappingFields(mappingNode(node, "content"), "headline", "section_label", "source_note")
}

// authoredSlides preserves each file's YAML nodes and comments. Inline slides
// point directly into the main document; referenced slides use their own AST.
func authoredSlides(p *Project, main *yaml.Node) (map[string]*yaml.Node, map[string]*yaml.Node, error) {
	slides := mappingNode(main.Content[0], "slides")
	objects := map[string]*yaml.Node{}
	documents := map[string]*yaml.Node{filepath.Base(p.SourcePath): main}
	for i, slide := range p.Document.Slides {
		node := slides.Content[i]
		if relative := p.SlideFiles[slide.ID]; relative != "" {
			document, err := sourceYAML(p.SourceFiles[filepath.ToSlash(filepath.Clean(relative))])
			if err != nil {
				return nil, nil, err
			}
			documents[relative], node = document, document.Content[0]
		}
		objects[slide.ID] = node
	}
	return objects, documents, nil
}

// commitSourceChanges validates the expanded candidate before replacing any
// file, checks every predecessor under one mutation guard, and restores bytes
// on an I/O failure. Unselected files are never serialized or rewritten.
func commitSourceChanges(p *Project, changes map[string][]byte, validate func(*Project) error) (*Project, error) {
	candidate, err := loadProject(p.SourcePath, changes)
	if err != nil {
		return nil, err
	}
	if validate != nil {
		if err := validate(candidate); err != nil {
			return nil, err
		}
	}
	guard, err := SafePath(p.Root, ".deck-source-mutation.lock")
	if err != nil {
		return nil, err
	}
	if err := writeExclusive(guard, []byte(p.SourceHash()), 0600); err != nil {
		return nil, fmt.Errorf("another source mutation is active: %w", err)
	}
	defer os.Remove(guard)
	// A retained source can become referenced again without being overwritten.
	// Verify its validated bytes under the guard too, closing the same drift
	// window that is checked for the previous authored tree below.
	for relative, expected := range candidate.SourceFiles {
		if _, previous := p.SourceFiles[relative]; previous {
			continue
		}
		if _, written := changes[relative]; written {
			continue
		}
		path, err := SafePath(p.Root, relative)
		if err != nil {
			return nil, err
		}
		actual, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(actual, expected) {
			return nil, fmt.Errorf("retained source changed during mutation: %s", relative)
		}
	}
	for relative, expected := range p.SourceFiles {
		path, err := SafePath(p.Root, relative)
		if err != nil {
			return nil, err
		}
		actual, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(actual, expected) {
			return nil, fmt.Errorf("authored source changed during mutation: %s", relative)
		}
	}
	type preparedFile struct {
		relative, path, temporary string
		previous                  []byte
		existed                   bool
		mode                      os.FileMode
	}
	prepared := []preparedFile{}
	defer func() {
		for _, file := range prepared {
			os.Remove(file.temporary)
		}
	}()
	keys := []string{}
	for relative, data := range changes {
		if previous, exists := p.SourceFiles[relative]; exists && bytes.Equal(previous, data) {
			continue
		}
		keys = append(keys, relative)
	}
	sort.Strings(keys)
	for _, relative := range keys {
		path, err := SafePath(p.Root, relative)
		if err != nil {
			return nil, err
		}
		file := preparedFile{relative: relative, path: path, mode: 0644}
		if info, err := os.Stat(path); err == nil {
			previous, authored := p.SourceFiles[relative]
			if !authored || !info.Mode().IsRegular() {
				return nil, fmt.Errorf("source destination already exists: %s", relative)
			}
			file.previous, file.existed, file.mode = previous, true, info.Mode().Perm()
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		file.temporary = filepath.Join(filepath.Dir(path), ".source-write-"+nonce()+".tmp")
		if err := writeExclusive(file.temporary, changes[relative], file.mode); err != nil {
			os.Remove(file.temporary)
			return nil, err
		}
		prepared = append(prepared, file)
	}
	// Preserve every preimage as a recoverable source tree.
	for relative, raw := range p.SourceFiles {
		path, err := SafePath(p.Root, "decisions/sources/"+p.SourceHash()+"/"+relative)
		if err != nil {
			return nil, err
		}
		if prior, err := os.ReadFile(path); err == nil {
			if !bytes.Equal(prior, raw) {
				return nil, fmt.Errorf("source predecessor snapshot drift: %s", relative)
			}
		} else if os.IsNotExist(err) {
			if err := writeExclusive(path, raw, 0444); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	for i, file := range prepared {
		if err := os.Rename(file.temporary, file.path); err != nil {
			for j := i - 1; j >= 0; j-- {
				prior := prepared[j]
				if !prior.existed {
					os.Remove(prior.path)
					continue
				}
				recovery := prior.path + ".recovery-" + nonce()
				if restoreErr := writeExclusive(recovery, prior.previous, prior.mode); restoreErr == nil {
					if restoreErr = os.Rename(recovery, prior.path); restoreErr != nil {
						return nil, fmt.Errorf("source mutation failed: %v; rollback failed: %v; predecessor tree retained", err, restoreErr)
					}
				} else {
					return nil, fmt.Errorf("source mutation failed: %v; rollback failed: %v; predecessor tree retained", err, restoreErr)
				}
			}
			return nil, err
		}
	}
	return candidate, nil
}
