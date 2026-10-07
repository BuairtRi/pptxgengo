package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

type CreateOptions struct {
	Out, ID, Title, Bundle, Engine, Template string
	Year                                     int
}

func CreateProject(o CreateOptions) (*Project, error) {
	if !stableID.MatchString(o.ID) || strings.TrimSpace(o.Title) == "" || o.Year < 2000 || o.Year > 9999 || o.Out == "" || o.Bundle == "" || o.Template == "" {
		return nil, fmt.Errorf("create requires new directory, stable ID, title, year, bundle and shared template")
	}
	slide, e := StockScaffoldSlideSource(o.Bundle, o.Template, "first-slide", o.Year)
	if e != nil {
		return nil, e
	}
	root, e := filepath.Abs(o.Out)
	if e != nil {
		return nil, e
	}
	if e = os.Mkdir(root, 0755); e != nil {
		return nil, e
	}
	success := false
	defer func() {
		if !success {
			os.RemoveAll(root)
		}
	}()
	document := map[string]any{"schema": Schema, "id": o.ID, "title": o.Title, "year": o.Year, "toolchain": map[string]any{"lockfile": "toolchain.lock.json"}, "context": map[string]any{"project": "context/project.md"}, "slides": []any{"slides/first-slide.yaml"}}
	node, e := editYAMLNode(document)
	if e != nil {
		return nil, e
	}
	raw, e := encodeSourceYAML(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}})
	if e != nil {
		return nil, e
	}
	files := map[string][]byte{"deck.yaml": raw, "slides/first-slide.yaml": slide, "context/project.md": []byte("# " + o.Title + "\n\nStage: intake. Replace scaffold example content before delivery.\n")}
	for rel, data := range files {
		path, e := SafePath(root, rel)
		if e != nil {
			return nil, e
		}
		if e = writeExclusive(path, data, 0644); e != nil {
			return nil, e
		}
	}
	for _, rel := range []string{"slides/templates", "assets/objects/sha256", "versions"} {
		path, e := SafePath(root, rel)
		if e != nil {
			return nil, e
		}
		if e = os.MkdirAll(path, 0755); e != nil {
			return nil, e
		}
	}
	p, e := Load(root)
	if e != nil {
		return nil, e
	}
	engine := o.Engine
	if engine == "" {
		engine = wmdesign.CandidateEngine
	}
	if _, e = Pin(p, o.Bundle, engine); e != nil {
		return nil, e
	}
	success = true
	return p, nil
}
