package deckproject

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"math"
	"reflect"
	"strings"
)

// decodeCurveDocument checks authored presence before typed decoding. Inapplicable
// zero, false and null fields must never disappear into a default-valued struct.
func decodeCurveDocument(raw []byte, file string, out any, schema string, allowed func(string, string) (string, error)) error {
	if len(raw) > 1<<20 {
		return fmt.Errorf("composition patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return e
	}
	if len(doc.Content) != 1 {
		return fmt.Errorf("empty composition patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("composition requires one document")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return e
	}
	typ := reflect.TypeOf(out).Elem()
	if e = p.shapeType(v, typ, ""); e != nil {
		return e
	}
	if e = strictInto(v, out); e != nil {
		return e
	}
	m, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("patch must be object")
	}
	hash, _ := m["expected_source_sha256"].(string)
	actor, _ := m["actor"].(string)
	reason, _ := m["reason"].(string)
	actual, _ := m["schema"].(string)
	_, e = hex.DecodeString(hash)
	ops, ok := m["operations"].([]any)
	if actual != schema || len(hash) != 64 || e != nil || strings.TrimSpace(actor) == "" || len(actor) > 256 || strings.TrimSpace(reason) == "" || len(reason) > 4096 || !ok || len(ops) < 1 || len(ops) > 500 {
		return fmt.Errorf("patch requires schema, source SHA256, actor, reason and 1..500 operations")
	}
	for _, rawop := range ops {
		op, ok := rawop.(map[string]any)
		if !ok {
			return fmt.Errorf("operation must be object")
		}
		a, _ := op["action"].(string)
		b, _ := op["entity"].(string)
		fields, e := allowed(a, b)
		if e != nil {
			return e
		}
		want := map[string]bool{}
		for _, f := range strings.Fields("action entity " + fields) {
			want[f] = true
		}
		for f := range op {
			if !want[f] {
				return fmt.Errorf("%s/%s does not accept authored field %s", a, b, f)
			}
		}
	}
	return nil
}
func curveFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func curveKeys(n *Node, path string, count int) ([]string, error) {
	// The shared cycle key reader generates source identities only when the node
	// has no explicit overlay. Every other overlay is retained by the caller.
	copy := map[string][]string{}
	for _, k := range []string{path, "/" + path} {
		if v, ok := n.Keys[k]; ok {
			copy[k] = v
		}
	}
	return cycleKeys(copy, path, count)
}
func curveNode(t *LocalTemplate, id, kind string) (*Node, error) {
	nodes, i := findDiagramNode(&t.Nodes, id)
	if nodes == nil {
		return nil, fmt.Errorf("unknown %s node %s", kind, id)
	}
	n := &(*nodes)[i]
	if n.Kind != "component" || n.Definition == nil || n.Definition.ID != "wmds/component/"+kind {
		return nil, fmt.Errorf("node %s must be a detached typed %s component", id, kind)
	}
	if n.Placement == nil || n.Placement.Rect == nil {
		return nil, fmt.Errorf("%s requires explicit rect placement", kind)
	}
	return n, nil
}
func curveSource(n *Node, values map[string]any) (map[string]any, bool, error) {
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	args, e := resolveArguments(n.Arguments, values)
	return args, bound, e
}
func curveSetKeys(n *Node, path string, keys []string) {
	if n.Keys == nil {
		n.Keys = map[string][]string{}
	}
	delete(n.Keys, "/"+path)
	if len(keys) == 0 {
		delete(n.Keys, path)
		return
	}
	n.Keys[path] = keys
}
func curveIndex(keys []string, key string) int {
	for i, k := range keys {
		if k == key {
			return i
		}
	}
	return -1
}
func curveReorder(keys, order []string) ([]int, error) {
	if len(keys) != len(order) {
		return nil, fmt.Errorf("reorder must enumerate every key exactly once")
	}
	seen := map[string]bool{}
	out := []int{}
	for _, k := range order {
		i := curveIndex(keys, k)
		if i < 0 || seen[k] {
			return nil, fmt.Errorf("unknown/duplicate reorder key %s", k)
		}
		seen[k] = true
		out = append(out, i)
	}
	return out, nil
}
