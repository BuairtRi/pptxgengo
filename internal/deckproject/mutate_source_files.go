package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func applySourceTemplate(p *Project, id string, template LocalTemplate, values map[string]any, targets []string, operation, hash string) (Mutation, error) {
	m := Mutation{Operation: operation, TemplateID: id, SlideIDs: targets, BeforeSHA256: p.SourceHash(), DefinitionSHA256: hash}
	main, err := sourceYAML(p.Raw)
	if err != nil {
		return m, err
	}
	slides, documents, err := authoredSlides(p, main)
	if err != nil {
		return m, err
	}
	changed := map[string]bool{filepath.Base(p.SourcePath): true}
	selected := map[string]bool{}
	for _, target := range targets {
		if selected[target] {
			return m, fmt.Errorf("duplicate target slide")
		}
		selected[target] = true
		node := slides[target]
		if node == nil {
			return m, fmt.Errorf("unknown target slide ID %s", target)
		}
		reference, err := editYAMLNode(Reference{Scope: "local", ID: id})
		if err != nil {
			return m, err
		}
		replaceMappingField(node, "template", reference)
		if values != nil {
			if operation == "detach" {
				content, bindings, err := localAuthoredContent(template, values)
				if err != nil {
					return m, err
				}
				contentNode, err := editYAMLNode(content)
				if err != nil {
					return m, err
				}
				bindingsNode, err := editYAMLNode(bindings)
				if err != nil {
					return m, err
				}
				replaceMappingField(node, "content", contentNode)
				replaceMappingField(node, "bindings", bindingsNode)
				removeMappingField(node, "values")
				// Replace the original stock scaffold header while keeping other
				// human comments on the authored slide intact.
				clearStockHeader(node)
				orderAuthoredSlide(node)
			} else {
				replacement, err := editYAMLNode(values)
				if err != nil {
					return m, err
				}
				removeMappingField(node, "content")
				removeMappingField(node, "bindings")
				replaceMappingField(node, "values", replacement)
			}
		}
		relative := p.SlideFiles[target]
		if relative == "" {
			relative = filepath.Base(p.SourcePath)
		}
		changed[relative] = true
	}
	templates := mappingNode(main.Content[0], "local_templates")
	if templates == nil {
		templates = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		replaceMappingField(main.Content[0], "local_templates", templates)
	}
	relative := "templates/" + id + ".yaml"
	replaceMappingField(templates, id, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: relative})
	definition, err := editYAMLNode(template)
	if err != nil {
		return m, err
	}
	definitionRaw, err := encodeSourceYAML(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{definition}})
	if err != nil {
		return m, err
	}
	changes := map[string][]byte{relative: definitionRaw}
	for path := range changed {
		raw, err := encodeSourceYAML(documents[path])
		if err != nil {
			return m, err
		}
		changes[path] = raw
	}
	m.Decision = "decisions/" + operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, err := SafePath(p.Root, m.Decision)
	if err != nil {
		return m, err
	}
	_, err = commitSourceChanges(p, changes, func(candidate *Project) error {
		m.AfterSHA256 = candidate.SourceHash()
		return writeJSON(decision, m)
	})
	if err != nil {
		os.Remove(decision)
	}
	return m, err
}

func localAuthoredContent(template LocalTemplate, values map[string]any) (map[string]any, map[string]string, error) {
	content, bindings := map[string]any{}, map[string]string{}
	keys := make([]string, 0, len(values))
	for id := range values {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	for _, id := range keys {
		alias := template.Zones[id].AuthoringAlias
		if alias == "" {
			alias = "/" + id
		}
		if _, duplicate := bindings[alias]; duplicate {
			alias = "/additional_content/" + id
		}
		if err := stockSetContent(content, strings.Split(strings.TrimPrefix(alias, "/"), "/"), values[id]); err != nil {
			return nil, nil, err
		}
		bindings[alias] = "/" + escape(id)
	}
	return content, bindings, nil
}

func clearStockHeader(node *yaml.Node) {
	clean := func(comment string) string {
		lines := []string{}
		for _, line := range strings.Split(comment, "\n") {
			if strings.Contains(line, "Shared template:") || strings.Contains(line, "layout is unchanged") {
				continue
			}
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	}
	node.HeadComment, node.LineComment, node.FootComment = clean(node.HeadComment), clean(node.LineComment), clean(node.FootComment)
	for _, child := range node.Content {
		clearStockHeader(child)
	}
}
