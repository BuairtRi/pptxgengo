package deckproject

import (
	"fmt"
	"strings"
)

// Specific adapters protect the positional source contracts behind generic
// collection editing. Operators address members/markers with stable keys.
func componentAdaptItem(n *Node, args map[string]any, path string, op *ComponentOperation) error {
	kind := strings.TrimPrefix(n.Definition.ID, "wmds/component/")
	if kind == "venn" && path == "/regions" && op.Action == "set" {
		value, ok := op.Value.(map[string]any)
		if !ok {
			return fmt.Errorf("venn region requires object with named members")
		}
		if _, ok := value["in"]; ok {
			return fmt.Errorf("venn region edits require members keys, not positional in indices")
		}
		raw, ok := value["members"].([]any)
		if !ok || len(raw) == 0 {
			return fmt.Errorf("venn region requires explicit members")
		}
		sets, ok := args["sets"].([]any)
		if !ok {
			return fmt.Errorf("missing venn sets")
		}
		keys, e := componentArrayKeys(n, "sets", sets)
		if e != nil {
			return e
		}
		indices := []any{}
		seen := map[string]bool{}
		for _, v := range raw {
			k, ok := v.(string)
			if !ok || seen[k] {
				return fmt.Errorf("invalid duplicate venn member")
			}
			seen[k] = true
			found := false
			for i, key := range keys {
				if k == key {
					indices = append(indices, float64(i))
					found = true
				}
			}
			if !found {
				return fmt.Errorf("unknown venn member %s", k)
			}
		}
		delete(value, "members")
		value["in"] = indices
	}
	if kind == "matrix" && path == "/cols" && op.Action == "set" {
		obj, ok := op.Value.(map[string]any)
		if ok {
			label, ok := obj["label"].(string)
			if !ok {
				return fmt.Errorf("matrix new column needs label")
			}
			for k := range obj {
				if k != "label" && k != "cells" {
					return fmt.Errorf("unexpected matrix column field")
				}
			}
			op.Value = label
		}
	}
	return nil
}
func componentCoupleCollections(n *Node, args map[string]any, path string, old, new []string, op ComponentOperation) error {
	kind := strings.TrimPrefix(n.Definition.ID, "wmds/component/")
	newIndex := map[string]int{}
	for i, k := range new {
		newIndex[k] = i
	}
	if kind == "venn" && path == "/sets" {
		regions, _ := args["regions"].([]any)
		regionKeys, e := componentArrayKeys(n, "regions", regions)
		if e != nil {
			return e
		}
		kept := []any{}
		keys := []string{}
		for i, raw := range regions {
			v, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid venn region")
			}
			indices, ok := v["in"].([]any)
			if !ok {
				return fmt.Errorf("invalid venn source membership")
			}
			mapped := []any{}
			deleted := false
			for _, rawIndex := range indices {
				num, ok := rawIndex.(float64)
				if !ok || num != float64(int(num)) || num < 0 || int(num) >= len(old) {
					return fmt.Errorf("invalid venn source index")
				}
				idx, exists := newIndex[old[int(num)]]
				if !exists {
					deleted = true
					break
				}
				mapped = append(mapped, float64(idx))
			}
			if deleted {
				if !op.Cascade {
					return fmt.Errorf("set removal affects regions; cascade required")
				}
				continue
			}
			v["in"] = mapped
			kept = append(kept, v)
			keys = append(keys, regionKeys[i])
		}
		args["regions"] = kept
		componentRebaseKeys(n, "regions", regionKeys, keys)
	}
	if kind == "matrix" && path == "/cols" {
		rows, ok := args["rows"].([]any)
		if !ok {
			return fmt.Errorf("matrix rows missing")
		}
		rowKeys, e := componentArrayKeys(n, "rows", rows)
		if e != nil {
			return e
		}
		oldIndex := map[string]int{}
		for i, k := range old {
			oldIndex[k] = i
		}
		var cellValues map[string]any
		if v, ok := op.Value.(map[string]any); ok {
			cellValues, _ = v["cells"].(map[string]any)
		}
		for i, raw := range rows {
			row, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid matrix row")
			}
			cells, ok := row["cells"].([]any)
			if !ok || len(cells) != len(old) {
				return fmt.Errorf("matrix columns and cells must have equal counts before editing")
			}
			cellPath := fmt.Sprintf("rows/%d/cells", i)
			oldCellKeys, e := componentArrayKeys(n, cellPath, cells)
			if e != nil {
				return e
			}
			next := []any{}
			nextKeys := []string{}
			for _, k := range new {
				if index, exists := oldIndex[k]; exists {
					next = append(next, cells[index])
					nextKeys = append(nextKeys, oldCellKeys[index])
				} else {
					v, exists := cellValues[rowKeys[i]]
					if !exists {
						return fmt.Errorf("new matrix column needs explicit cell for stable row %s", rowKeys[i])
					}
					next = append(next, v)
					nextKeys = append(nextKeys, k)
				}
			}
			row["cells"] = next
			componentRebaseKeys(n, cellPath, oldCellKeys, nextKeys)
		}
	}
	return nil
}
func componentSemanticReferences(n *Node, args map[string]any) (map[string][]string, map[string]string, error) {
	relationships, markers := map[string][]string{}, map[string]string{}
	kind := strings.TrimPrefix(n.Definition.ID, "wmds/component/")
	if kind == "phases" {
		list, _ := args["phases"].([]any)
		keys, e := componentArrayKeys(n, "phases", list)
		if e != nil {
			return nil, nil, e
		}
		if raw, ok := args["current"].(float64); ok {
			if raw != float64(int(raw)) || raw < 0 || int(raw) >= len(keys) {
				return nil, nil, fmt.Errorf("invalid current phase")
			}
			markers["current"] = keys[int(raw)]
		}
	}
	if kind == "venn" {
		sets, _ := args["sets"].([]any)
		keys, e := componentArrayKeys(n, "sets", sets)
		if e != nil {
			return nil, nil, e
		}
		regions, _ := args["regions"].([]any)
		rkeys, e := componentArrayKeys(n, "regions", regions)
		if e != nil {
			return nil, nil, e
		}
		for i, raw := range regions {
			v, ok := raw.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("invalid venn region")
			}
			indices, _ := v["in"].([]any)
			members := []string{}
			for _, raw := range indices {
				idx, ok := raw.(float64)
				if !ok || idx != float64(int(idx)) || idx < 0 || int(idx) >= len(keys) {
					return nil, nil, fmt.Errorf("invalid venn membership")
				}
				members = append(members, keys[int(idx)])
			}
			relationships[rkeys[i]] = members
		}
	}
	return relationships, markers, nil
}
