package deckproject

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const ComponentPatchSchema = "pptxgengo.component-patch.v1"

type ComponentOperation struct {
	Action  string         `json:"action"`
	Entity  string         `json:"entity"`
	Path    string         `json:"path,omitempty"`
	Key     string         `json:"key,omitempty"`
	Value   any            `json:"value,omitempty"`
	Order   []string       `json:"order,omitempty"`
	Cascade bool           `json:"cascade,omitempty"`
	Rect    *wmdesign.Rect `json:"rect,omitempty"`
}
type ComponentPatch struct {
	Schema               string               `json:"schema"`
	ExpectedSourceSHA256 string               `json:"expected_source_sha256"`
	Actor                string               `json:"actor"`
	Reason               string               `json:"reason"`
	SemanticReview       string               `json:"semantic_review"`
	NodeID               string               `json:"node_id"`
	Operations           []ComponentOperation `json:"operations"`
}
type ComponentCollection struct {
	sourcePath string
	Path       string   `json:"path"`
	Keys       []string `json:"keys"`
	Count      int      `json:"count"`
}
type ComponentInspection struct {
	Schema          string                `json:"schema"`
	SlideID         string                `json:"slide_id"`
	NodeID          string                `json:"node_id"`
	SourceSHA256    string                `json:"source_sha256"`
	Component       string                `json:"component"`
	Arguments       map[string]any        `json:"arguments"`
	Relationships   map[string][]string   `json:"relationships,omitempty"`
	Markers         map[string]string     `json:"markers,omitempty"`
	Collections     []ComponentCollection `json:"collections"`
	MutationBlocked string                `json:"mutation_blocked,omitempty"`
	RenderError     string                `json:"render_error,omitempty"`
	Geometry        DiagramInspection     `json:"geometry"`
}

func componentFields(a, e string) (map[string]bool, error) {
	s := "action entity"
	switch a + "/" + e {
	case "materialize/source":
	case "set/argument":
		s += " path value"
	case "set/item":
		s += " path key value"
	case "remove/item":
		s += " path key cascade"
	case "reorder/item":
		s += " path order"
	case "set/marker":
		s += " path key"
	case "remove/marker":
		s += " path"
	case "set/layout":
		s += " rect"
	default:
		return nil, fmt.Errorf("unsupported component %s/%s", a, e)
	}
	out := map[string]bool{}
	for _, k := range strings.Fields(s) {
		out[k] = true
	}
	return out, nil
}
func DecodeComponentPatch(raw []byte, file string) (ComponentPatch, error) {
	var out ComponentPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("component patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty component patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("one component patch document required")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = p.shapeType(v, reflect.TypeOf(out), ""); e != nil {
		return out, e
	}
	if e = strictInto(v, &out); e != nil {
		return out, e
	}
	ops, _ := v.(map[string]any)["operations"].([]any)
	for i, op := range out.Operations {
		fields, e := componentFields(op.Action, op.Entity)
		if e != nil {
			return out, e
		}
		authored := ops[i].(map[string]any)
		for k := range authored {
			if !fields[k] {
				return out, fmt.Errorf("component operation does not accept field %s", k)
			}
		}
		if op.Action == "set" && (op.Entity == "argument" || op.Entity == "item") {
			if _, ok := authored["value"]; !ok {
				return out, fmt.Errorf("set requires authored value, including explicit null")
			}
		}
	}
	_, e = hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != ComponentPatchSchema || len(out.ExpectedSourceSHA256) != 64 || e != nil || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || strings.TrimSpace(out.SemanticReview) == "" || len(out.Actor) > 256 || len(out.Reason) > 4096 || len(out.SemanticReview) > 4096 || !stableID.MatchString(out.NodeID) || len(out.Operations) < 1 || len(out.Operations) > 500 {
		return out, fmt.Errorf("component patch requires schema, SHA256, actor, reason, semantic_review, node_id and operations")
	}
	return out, nil
}
func componentSpecialized(kind string) string {
	switch kind {
	case "table", "editable-table":
		return "table"
	case "process", "swimlane":
		return "process"
	case "portfolio":
		return "portfolio"
	case "gantt":
		return "gantt"
	case "road", "roadfork":
		return "journey"
	case "cycle":
		return "cycle"
	case "maturity":
		return "maturity"
	case "curve", "teamcurve", "staffing":
		return "staffing"
	case "assessment":
		return "assessment"
	case "orgchart", "pod", "pods", "role", "org", "governance":
		return "team"
	case "commercial":
		return "commercial"
	case "chart":
		return "quantitative"
	}
	return ""
}
func componentArgs(n *Node, values map[string]any) (map[string]any, bool, error) {
	if n.Kind != "component" || n.Definition == nil || n.Definition.Scope != "shared" || !strings.HasPrefix(n.Definition.ID, "wmds/component/") || len(n.Nodes) > 0 {
		return nil, false, fmt.Errorf("component requires a typed local leaf")
	}
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	v, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return nil, bound, e
	}
	var args map[string]any
	e = strictInto(v, &args)
	return args, bound, e
}
func componentKeyPath(path string) string { return strings.TrimPrefix(path, "/") }
func componentArrayKeys(n *Node, path string, list []any) ([]string, error) {
	path = componentKeyPath(path)
	a, ok := n.Keys[path]
	b, other := n.Keys["/"+path]
	if ok && other && !bytes.Equal(canonical(a), canonical(b)) {
		return nil, fmt.Errorf("conflicting array keys %s", path)
	}
	if !ok && other {
		a, ok = b, true
	}
	if !ok {
		for i, v := range list {
			k := ""
			if obj, ok := v.(map[string]any); ok {
				if value, ok := obj["key"].(string); ok {
					k = value
				}
			}
			if k == "" {
				k = fmt.Sprintf("source-%03d", i+1)
			}
			a = append(a, k)
		}
	}
	if len(a) != len(list) {
		return nil, fmt.Errorf("array key count differs %s", path)
	}
	seen := map[string]bool{}
	for _, k := range a {
		if !stableID.MatchString(k) || seen[k] {
			return nil, fmt.Errorf("invalid/duplicate array key %s", path)
		}
		seen[k] = true
	}
	return append([]string{}, a...), nil
}
func componentCollections(n *Node, args map[string]any) ([]ComponentCollection, error) {
	out := []ComponentCollection{}
	var walk func(any, string, string) error
	walk = func(v any, sourcePath, publicPath string) error {
		switch obj := v.(type) {
		case map[string]any:
			names := []string{}
			for k := range obj {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, k := range names {
				escaped := strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1")
				if e := walk(obj[k], sourcePath+"/"+escaped, publicPath+"/"+escaped); e != nil {
					return e
				}
			}
		case []any:
			keys, e := componentArrayKeys(n, sourcePath, obj)
			if e != nil {
				return e
			}
			out = append(out, ComponentCollection{sourcePath: sourcePath, Path: publicPath, Keys: keys, Count: len(obj)})
			for i, value := range obj {
				if e := walk(value, sourcePath+"/"+strconv.Itoa(i), publicPath+"/@"+keys[i]); e != nil {
					return e
				}
			}
		}
		return nil
	}
	e := walk(args, "", "")
	return out, e
}

// Public selectors address array keys as @KEY, never raw numeric indices.
func componentResolvePath(n *Node, args map[string]any, path string) (string, error) {
	if path == "" || path[0] != '/' || len(path) > 2048 {
		return "", fmt.Errorf("component path must be a bounded JSON pointer")
	}
	parts := strings.Split(path[1:], "/")
	cur := any(args)
	resolved := ""
	for _, part := range parts {
		key := strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[key]
			if !ok {
				return "", fmt.Errorf("unknown argument path %s", path)
			}
			cur = next
			resolved += "/" + part
		case []any:
			if !strings.HasPrefix(key, "@") {
				return "", fmt.Errorf("array paths must use @stable-key selectors")
			}
			keys, e := componentArrayKeys(n, resolved, v)
			if e != nil {
				return "", e
			}
			idx := -1
			for i, k := range keys {
				if k == key[1:] {
					idx = i
				}
			}
			if idx < 0 {
				return "", fmt.Errorf("unknown stable item %s", key)
			}
			cur = v[idx]
			resolved += "/" + strconv.Itoa(idx)
		default:
			return "", fmt.Errorf("path descends through scalar")
		}
	}
	return resolved, nil
}
func componentRebaseKeys(n *Node, path string, old, new []string) {
	prefix := componentKeyPath(path)
	next := map[string][]string{}
	newIndex := map[string]int{}
	for i, k := range new {
		newIndex[k] = i
	}
	for key, keys := range n.Keys {
		clean := componentKeyPath(key)
		if clean == prefix {
			continue
		}
		if strings.HasPrefix(clean, prefix+"/") {
			tail := strings.TrimPrefix(clean, prefix+"/")
			parts := strings.SplitN(tail, "/", 2)
			idx, e := strconv.Atoi(parts[0])
			if e == nil && idx >= 0 && idx < len(old) {
				ni, exists := newIndex[old[idx]]
				if !exists {
					continue
				}
				clean = prefix + "/" + strconv.Itoa(ni)
				if len(parts) == 2 {
					clean += "/" + parts[1]
				}
			}
		}
		next[clean] = keys
	}
	next[prefix] = append([]string{}, new...)
	n.Keys = next
}
func componentPopulateKeys(n *Node, args map[string]any) error {
	for path := range n.Keys {
		value, e := lookupPointer(args, "/"+componentKeyPath(path))
		if e != nil {
			delete(n.Keys, path)
			continue
		}
		if _, ok := value.([]any); !ok {
			delete(n.Keys, path)
		}
	}
	collections, e := componentCollections(n, args)
	if e != nil {
		return e
	}
	if n.Keys == nil {
		n.Keys = map[string][]string{}
	}
	for _, c := range collections {
		delete(n.Keys, "/"+componentKeyPath(c.sourcePath))
		n.Keys[componentKeyPath(c.sourcePath)] = c.Keys
	}
	return nil
}
func InspectComponent(p *Project, slideID, nodeID, bundle, engine string) (ComponentInspection, error) {
	out := ComponentInspection{Schema: "pptxgengo.component-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	list, i := findDiagramNode(&t.Nodes, nodeID)
	if list == nil {
		return out, fmt.Errorf("unknown component node")
	}
	n := &(*list)[i]
	var bound bool
	out.Arguments, bound, e = componentArgs(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	out.Component = strings.TrimPrefix(n.Definition.ID, "wmds/component/")
	out.Relationships, out.Markers, e = componentSemanticReferences(n, out.Arguments)
	if e != nil {
		return out, e
	}
	out.Collections, e = componentCollections(n, out.Arguments)
	if e != nil {
		return out, e
	}
	if command := componentSpecialized(out.Component); command != "" {
		out.MutationBlocked = "Use project " + command + " for this semantic model"
	} else if bound {
		out.MutationBlocked = "Bound source: edit values or materialize/source first"
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func PatchComponent(p *Project, slideID string, patch ComponentPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeComponentPatch(canonical(patch), "component-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("component source hash mismatch")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	list, i := findDiagramNode(&clone.Nodes, patch.NodeID)
	if list == nil {
		return empty, fmt.Errorf("unknown component")
	}
	n := &(*list)[i]
	args, bound, e := componentArgs(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	if command := componentSpecialized(strings.TrimPrefix(n.Definition.ID, "wmds/component/")); command != "" {
		return empty, fmt.Errorf("use project %s; generic mutation would bypass semantic invariants", command)
	}
	if e = componentPopulateKeys(n, args); e != nil {
		return empty, e
	}
	for j, op := range patch.Operations {
		if op.Action == "materialize" && op.Entity == "source" && j == 0 {
			bound = false
			continue
		}
		if bound {
			return empty, fmt.Errorf("bound component requires materialize/source first")
		}
		if op.Action == "set" && op.Entity == "layout" {
			if op.Rect == nil || n.Placement == nil {
				return empty, fmt.Errorf("existing placement and rect required")
			}
			n.Placement.Rect = op.Rect
			continue
		}
		if op.Entity == "marker" {
			if n.Definition.ID != "wmds/component/phases" || op.Path != "/current" {
				return empty, fmt.Errorf("only phases current supports marker operations")
			}
			if op.Action == "remove" {
				delete(args, "current")
				continue
			}
			list, _ := args["phases"].([]any)
			keys, e := componentArrayKeys(n, "phases", list)
			if e != nil {
				return empty, e
			}
			found := false
			for i, k := range keys {
				if k == op.Key {
					args["current"] = float64(i)
					found = true
				}
			}
			if !found {
				return empty, fmt.Errorf("unknown current phase")
			}
			continue
		}
		path, e := componentResolvePath(n, args, op.Path)
		if e != nil {
			return empty, e
		}
		root := strings.Split(strings.TrimPrefix(path, "/"), "/")[0]
		if root == "type" || root == "id" || root == "x" || root == "y" || root == "w" || root == "h" || root == "_h" || root == "_source_geometry" {
			return empty, fmt.Errorf("reserved geometry field; use layout")
		}
		value, e := lookupPointer(args, path)
		if e != nil {
			return empty, e
		}
		if op.Entity == "argument" {
			if n.Definition.ID == "wmds/component/phases" && path == "/current" {
				return empty, fmt.Errorf("use set/marker with stable phase key")
			}
			if n.Definition.ID == "wmds/component/venn" && strings.HasPrefix(path, "/regions/") && strings.Contains(path, "/in") {
				return empty, fmt.Errorf("edit region with named members, not indices")
			}
			switch value.(type) {
			case map[string]any, []any:
				return empty, fmt.Errorf("argument operation accepts scalar fields only")
			}
			switch op.Value.(type) {
			case map[string]any, []any:
				return empty, fmt.Errorf("scalar replacement required")
			}
			if e = setPointer(args, path, op.Value); e != nil {
				return empty, e
			}
			continue
		}
		items, ok := value.([]any)
		if !ok {
			return empty, fmt.Errorf("item path requires array")
		}
		keys, e := componentArrayKeys(n, path, items)
		if e != nil {
			return empty, e
		}
		old := append([]string{}, keys...)
		currentKey := ""
		if n.Definition.ID == "wmds/component/phases" && path == "/phases" {
			if current, ok := args["current"].(float64); ok {
				if current != float64(int(current)) || int(current) < 0 || int(current) >= len(old) {
					return empty, fmt.Errorf("invalid phase current index")
				}
				currentKey = old[int(current)]
			}
		}
		keyIndex := -1
		for i, k := range keys {
			if k == op.Key {
				keyIndex = i
			}
		}
		originalOp := op
		if e = componentAdaptItem(n, args, path, &op); e != nil {
			return empty, e
		}
		switch op.Action {
		case "set":
			if !stableID.MatchString(op.Key) {
				return empty, fmt.Errorf("item requires stable key")
			}
			if obj, ok := op.Value.(map[string]any); ok {
				if k, exists := obj["key"]; exists && k != op.Key {
					return empty, fmt.Errorf("item key and authored key differ")
				}
			}
			if keyIndex < 0 {
				keys = append(keys, op.Key)
				items = append(items, op.Value)
			} else {
				items[keyIndex] = op.Value
			}
		case "remove":
			if keyIndex < 0 {
				return empty, fmt.Errorf("unknown item")
			}
			if !op.Cascade {
				return empty, fmt.Errorf("item removal requires cascade acknowledgement of nested content and dependent references")
			}
			keys = append(keys[:keyIndex], keys[keyIndex+1:]...)
			items = append(items[:keyIndex], items[keyIndex+1:]...)
		case "reorder":
			if len(op.Order) != len(keys) {
				return empty, fmt.Errorf("reorder requires all keys")
			}
			seen := map[string]bool{}
			next := []any{}
			for _, k := range op.Order {
				if seen[k] {
					return empty, fmt.Errorf("duplicate order key")
				}
				seen[k] = true
				found := false
				for i, key := range keys {
					if k == key {
						next = append(next, items[i])
						found = true
					}
				}
				if !found {
					return empty, fmt.Errorf("unknown order key")
				}
			}
			keys = append([]string{}, op.Order...)
			items = next
		default:
			return empty, fmt.Errorf("materialize/source must be first")
		}
		if e = setPointer(args, path, items); e != nil {
			return empty, e
		}
		if currentKey != "" {
			delete(args, "current")
			for i, k := range keys {
				if k == currentKey {
					args["current"] = i
				}
			}
		}
		if e = componentCoupleCollections(n, args, path, old, keys, originalOp); e != nil {
			return empty, e
		}
		componentRebaseKeys(n, path, old, keys)
		if e = componentPopulateKeys(n, args); e != nil {
			return empty, e
		}
	}
	n.Arguments = args
	return CompositionCandidate(p, slideID, "component", patch.Actor, patch.Reason+"; semantic review: "+patch.SemanticReview, clone, bundle, engine, apply)
}

func (op ComponentOperation) MarshalJSON() ([]byte, error) {
	type plain ComponentOperation
	b, e := json.Marshal(plain(op))
	if e != nil {
		return nil, e
	}
	var obj map[string]any
	if e = json.Unmarshal(b, &obj); e != nil {
		return nil, e
	}
	if op.Action == "set" && (op.Entity == "argument" || op.Entity == "item") {
		obj["value"] = op.Value
	}
	return json.Marshal(obj)
}
