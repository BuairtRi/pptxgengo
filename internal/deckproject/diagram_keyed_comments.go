package deckproject

import (
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// preserveDiagramKeyedArgumentComments attaches changed array records to their
// explicit authored identity. Unkeyed and ambiguous arrays retain the ordinary
// value-based policy; neither positions nor matching labels establish identity.
func preserveDiagramKeyedArgumentComments(old, current *yaml.Node) {
	if old == nil || current == nil || old.Kind != yaml.MappingNode || current.Kind != yaml.MappingNode {
		return
	}
	oldKeys, newKeys := mappingNode(old, "keys"), mappingNode(current, "keys")
	oldArguments, newArguments := mappingNode(old, "arguments"), mappingNode(current, "arguments")
	if oldKeys == nil || newKeys == nil || oldArguments == nil || newArguments == nil || oldKeys.Kind != yaml.MappingNode || newKeys.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(newKeys.Content); i += 2 {
		path := newKeys.Content[i].Value
		oldPath, ok := diagramCommentOldPath(oldKeys, newKeys, path)
		if !ok {
			continue
		}
		previous, now := diagramCommentOverlay(oldKeys, oldPath), newKeys.Content[i+1]
		before, after := diagramCommentPointer(oldArguments, oldPath), diagramCommentPointer(newArguments, path)
		if previous == nil || now.Kind != yaml.SequenceNode || previous.Kind != yaml.SequenceNode || before == nil || after == nil || before.Kind != yaml.SequenceNode || after.Kind != yaml.SequenceNode || len(previous.Content) != len(before.Content) || len(now.Content) != len(after.Content) {
			continue
		}
		oldIndex := map[string]int{}
		ambiguous := map[string]bool{}
		for j, key := range previous.Content {
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || key.Value == "" {
				ambiguous[key.Value] = true
				continue
			}
			if _, exists := oldIndex[key.Value]; exists {
				ambiguous[key.Value] = true
			}
			oldIndex[key.Value] = j
		}
		newCount := map[string]int{}
		for _, key := range now.Content {
			if key.Kind == yaml.ScalarNode && key.Tag == "!!str" {
				newCount[key.Value]++
			}
		}
		for j, key := range now.Content {
			pos, exists := oldIndex[key.Value]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || !exists || ambiguous[key.Value] || newCount[key.Value] != 1 {
				continue
			}
			preserveDiagramComments(before.Content[pos], after.Content[j])
		}
	}
	// A staffing series is a matrix over the point identities carried by at.
	// Each series uses the same point keys, while series order has its own keys.
	// Preserve scalar observation comments by both identities, never value/index.
	oldDef, newDef := mappingNode(old, "definition"), mappingNode(current, "definition")
	if oldDef == nil || newDef == nil {
		return
	}
	oldID, newID := mappingNode(oldDef, "id"), mappingNode(newDef, "id")
	if oldID == nil || newID == nil || oldID.Value != "wmds/component/teamcurve" || newID.Value != oldID.Value {
		return
	}
	oldSeries, newSeries := mappingNode(oldArguments, "series"), mappingNode(newArguments, "series")
	if oldSeries == nil || newSeries == nil || oldSeries.Kind != yaml.SequenceNode || newSeries.Kind != yaml.SequenceNode {
		return
	}
	beforeSeries := diagramCommentKeyIndices(mappingNode(oldKeys, "series"), len(oldSeries.Content))
	afterSeries := diagramCommentKeyIndices(mappingNode(newKeys, "series"), len(newSeries.Content))
	if beforeSeries == nil || afterSeries == nil {
		return
	}
	for key, nowIndex := range afterSeries {
		oldIndex, exists := beforeSeries[key]
		if !exists {
			continue
		}
		beforeValues := mappingNode(oldSeries.Content[oldIndex], "values")
		afterValues := mappingNode(newSeries.Content[nowIndex], "values")
		if beforeValues == nil || afterValues == nil || beforeValues.Kind != yaml.SequenceNode || afterValues.Kind != yaml.SequenceNode {
			continue
		}
		oldPoints := diagramCommentKeyIndices(mappingNode(oldKeys, "at"), len(beforeValues.Content))
		newPoints := diagramCommentKeyIndices(mappingNode(newKeys, "at"), len(afterValues.Content))
		if oldPoints == nil || newPoints == nil {
			continue
		}
		for point, pointIndex := range newPoints {
			previous, exists := oldPoints[point]
			if exists {
				preserveDiagramComments(beforeValues.Content[previous], afterValues.Content[pointIndex])
			}
		}
	}
}

func diagramCommentKeyIndices(keys *yaml.Node, count int) map[string]int {
	if keys == nil || keys.Kind != yaml.SequenceNode || len(keys.Content) != count {
		return nil
	}
	out := map[string]int{}
	for i, key := range keys.Content {
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || key.Value == "" {
			return nil
		}
		if _, exists := out[key.Value]; exists {
			return nil
		}
		out[key.Value] = i
	}
	return out
}

func diagramCommentPointer(node *yaml.Node, path string) *yaml.Node {
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		if node == nil {
			return nil
		}
		switch node.Kind {
		case yaml.MappingNode:
			node = mappingNode(node, part)
		case yaml.SequenceNode:
			i, e := strconv.Atoi(part)
			if e != nil || i < 0 || i >= len(node.Content) {
				return nil
			}
			node = node.Content[i]
		default:
			return nil
		}
	}
	return node
}

// Descendant overlays follow a keyed ancestor when its record index changes.
// An explicit ancestor overlay that is ambiguous or removes the ancestor refuses
// the descendant match instead of borrowing the old numeric array position.
func diagramCommentOldPath(oldKeys, newKeys *yaml.Node, path string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	before := append([]string(nil), parts...)
	for i, part := range parts {
		if i == 0 {
			continue
		}
		index, e := strconv.Atoi(part)
		if e != nil {
			continue
		}
		now := diagramCommentOverlay(newKeys, strings.Join(parts[:i], "/"))
		if now == nil {
			continue
		}
		previous := diagramCommentOverlay(oldKeys, strings.Join(before[:i], "/"))
		if now.Kind != yaml.SequenceNode || index < 0 || index >= len(now.Content) || previous == nil || previous.Kind != yaml.SequenceNode {
			return "", false
		}
		nowIndex := diagramCommentKeyIndices(now, len(now.Content))
		oldIndex := diagramCommentKeyIndices(previous, len(previous.Content))
		if nowIndex == nil || oldIndex == nil {
			return "", false
		}
		prior, exists := oldIndex[now.Content[index].Value]
		if !exists {
			return "", false
		}
		before[i] = strconv.Itoa(prior)
	}
	return strings.Join(before, "/"), true
}

func diagramCommentOverlay(keys *yaml.Node, path string) *yaml.Node {
	if keys == nil {
		return nil
	}
	path = strings.TrimPrefix(path, "/")
	if n := mappingNode(keys, path); n != nil {
		return n
	}
	return mappingNode(keys, "/"+path)
}
