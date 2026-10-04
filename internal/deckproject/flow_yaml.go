package deckproject

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// flowMapCommaDiagnostic catches the common case where a comma inside an
// unquoted flow-map value turns the remaining prose into a null-valued key.
// Restricting the check to phrase-like keys with null values avoids treating
// ordinary flow-map fields and explicit nulls as syntax mistakes.
func flowMapCommaDiagnostic(raw []byte, source string) error {
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return nil // Let the normal YAML decoder return the syntax error.
	}
	var visit func(*yaml.Node) error
	visit = func(node *yaml.Node) error {
		if node == nil {
			return nil
		}
		if node.Kind == yaml.MappingNode && node.Style&yaml.FlowStyle != 0 {
			for i := 0; i+1 < len(node.Content); i += 2 {
				key, value := node.Content[i], node.Content[i+1]
				if key.Kind == yaml.ScalarNode && key.Tag == "!!str" && value.Kind == yaml.ScalarNode && value.Tag == "!!null" && value.Value == "" && phraseLikeFlowKey(key.Value) {
					return fmt.Errorf("%s:%d:%d: possible unquoted comma in a flow-style mapping: %q was parsed as a key; quote the complete comma-containing value, for example {text: \"owner, date, and next step\"}", source, key.Line, key.Column, key.Value)
				}
			}
		}
		for _, child := range node.Content {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	return visit(&document)
}

func phraseLikeFlowKey(key string) bool {
	key = strings.TrimSpace(key)
	if strings.ContainsAny(key, " \t") {
		return true
	}
	// A trailing conjunction commonly results from `{text: claim, and next}`.
	switch strings.ToLower(key) {
	case "and", "or", "but", "with", "including":
		return true
	default:
		return false
	}
}
