package deckproject

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

func (p *Project) readSource(relative string, limit int) ([]byte, error) {
	path, err := SafePath(p.Root, relative)
	if err != nil {
		return nil, err
	}
	relative = filepath.ToSlash(filepath.Clean(relative))
	data, exists := p.sourceOverrides[relative]
	if !exists {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > int64(limit) {
			return nil, fmt.Errorf("%s: source must be a regular file at most %d bytes", path, limit)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	if len(data) > limit {
		return nil, fmt.Errorf("%s: source exceeds %d bytes", path, limit)
	}
	p.SourceFiles[relative] = data
	total := 0
	for _, raw := range p.SourceFiles {
		total += len(raw)
	}
	if total > 64<<20 {
		return nil, fmt.Errorf("authored source files exceed 64 MiB")
	}
	return data, nil
}

func (p *Project) parseSource(raw []byte, relative, pointer string) (any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var node, extra yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return nil, fmt.Errorf("%s: %w", relative, err)
	}
	if len(node.Content) != 1 {
		return nil, fmt.Errorf("%s: empty YAML document", relative)
	}
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("%s: exactly one YAML document required", relative)
	}
	previous := p.activeSource
	p.activeSource = relative
	defer func() { p.activeSource = previous }()
	return p.yamlValue(node.Content[0], pointer, 0)
}

func (p *Project) referencedMapping(relative, pointer string) (map[string]any, error) {
	raw, err := p.readSource(relative, 16<<20)
	if err != nil {
		return nil, p.fail(pointer, "source reference: %v", err)
	}
	value, err := p.parseSource(raw, relative, pointer)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, p.fail(pointer, "referenced YAML must contain one mapping")
	}
	return object, nil
}

func (p *Project) expandSourceReferences(tree map[string]any) error {
	seen := map[string]bool{filepath.Base(p.SourcePath): true}
	if templates, ok := tree["local_templates"].(map[string]any); ok {
		for id, value := range templates {
			if relative, ok := value.(string); ok {
				pointer := "/local_templates/" + escape(id)
				key := filepath.ToSlash(filepath.Clean(relative))
				if seen[key] {
					return p.fail(pointer, "duplicate authored source reference %s", relative)
				}
				seen[key] = true
				object, err := p.referencedMapping(relative, pointer)
				if err != nil {
					return err
				}
				p.TemplateFiles[id], templates[id] = filepath.ToSlash(filepath.Clean(relative)), object
			}
		}
	}
	if slides, ok := tree["slides"].([]any); ok {
		for i, value := range slides {
			pointer := "/slides/" + strconv.Itoa(i)
			var file string
			if relative, ok := value.(string); ok {
				key := filepath.ToSlash(filepath.Clean(relative))
				if seen[key] {
					return p.fail(pointer, "duplicate authored source reference %s", relative)
				}
				seen[key] = true
				object, err := p.referencedMapping(relative, pointer)
				if err != nil {
					return err
				}
				file, value, slides[i] = relative, object, object
			}
			slide, ok := value.(map[string]any)
			if !ok {
				continue // The existing strict shape validator explains this.
			}
			id, _ := slide["id"].(string)
			if file != "" {
				p.SlideFiles[id] = filepath.ToSlash(filepath.Clean(file))
			}
			if rawRelative, exists := slide["notes_file"]; exists {
				relative, ok := rawRelative.(string)
				if !ok || relative == "" {
					return p.fail(pointer+"/notes_file", "notes_file requires a project-relative Markdown path")
				}
				if _, exists := slide["notes"]; exists {
					return p.fail(pointer+"/notes_file", "use notes or notes_file, not both")
				}
				raw, err := p.readSource(relative, 1<<20)
				if err != nil {
					return p.fail(pointer+"/notes_file", "%v", err)
				}
				p.NotesFiles[id], slide["notes"] = filepath.ToSlash(filepath.Clean(relative)), string(raw)
				p.Positions[pointer+"/notes"] = Position{Line: 1, Column: 1}
				p.positionFiles[pointer+"/notes"] = relative
				delete(slide, "notes_file")
			}
			if err := p.expandContentAliases(slide, pointer); err != nil {
				return err
			}
		}
	}
	return nil
}

// SourceHash keeps the single-file contract while hashing every external
// authored file in a multi-file project, including Markdown notes.
func (p *Project) SourceHash() string {
	if len(p.SourceFiles) <= 1 {
		return digest(p.Raw)
	}
	files := map[string]string{}
	for relative, raw := range p.SourceFiles {
		files[relative] = digest(raw)
	}
	return digest(canonical(files))
}
