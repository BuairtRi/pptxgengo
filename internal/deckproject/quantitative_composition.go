package deckproject

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

const QuantitativePatchSchema = "pptxgengo.quantitative-patch.v1"

// Categories and series have stable source keys. Values address both keys;
// they never address a label or a mutable array index.
type QuantitativeOperation struct {
	Action  string              `json:"action"`
	Entity  string              `json:"entity"`
	Key     string              `json:"key,omitempty"`
	Series  string              `json:"series,omitempty"`
	Label   string              `json:"label,omitempty"`
	Value   *float64            `json:"value,omitempty"`
	Missing bool                `json:"missing,omitempty"`
	Cascade bool                `json:"cascade,omitempty"`
	Order   []string            `json:"order,omitempty"`
	Data    map[string]any      `json:"data,omitempty"`
	Keys    map[string][]string `json:"keys,omitempty"`
	Point   []any               `json:"point,omitempty"`
	Field   string              `json:"field,omitempty"`
	Setting any                 `json:"setting,omitempty"`
}
type QuantitativePatch struct {
	Schema               string                  `json:"schema"`
	ExpectedSourceSHA256 string                  `json:"expected_source_sha256"`
	Actor                string                  `json:"actor"`
	Reason               string                  `json:"reason"`
	NodeID               string                  `json:"node_id"`
	Operations           []QuantitativeOperation `json:"operations"`
}
type QuantitativeInspection struct {
	Schema          string              `json:"schema"`
	SlideID         string              `json:"slide_id"`
	NodeID          string              `json:"node_id"`
	SourceSHA256    string              `json:"source_sha256"`
	Source          map[string]any      `json:"source"`
	Keys            map[string][]string `json:"keys"`
	Contract        string              `json:"contract"`
	MutationBlocked string              `json:"mutation_blocked,omitempty"`
	RenderError     string              `json:"render_error,omitempty"`
	Geometry        DiagramInspection   `json:"geometry"`
}

func quantitativeFields(a, e string) (map[string]bool, error) {
	fields := "action entity"
	switch a + "/" + e {
	case "materialize/source":
	case "replace/source":
		fields += " data keys cascade"
	case "set/category":
		fields += " key label"
	case "set/series", "set/item":
		fields += " key data"
	case "set/point":
		fields += " key point"
	case "set/value":
		fields += " key series value missing"
	case "remove/category", "remove/series", "remove/point", "remove/item":
		fields += " key cascade"
	case "reorder/category", "reorder/series", "reorder/point", "reorder/item":
		fields += " order"
	case "set/setting":
		fields += " field setting"
	default:
		return nil, fmt.Errorf("unsupported quantitative operation %s/%s", a, e)
	}
	out := map[string]bool{}
	for _, f := range strings.Fields(fields) {
		out[f] = true
	}
	return out, nil
}

// Shared strict decoder checks authored field presence, including false/null.
func decodeDomainPatch(raw []byte, file string, out any, fields func(string, string) (map[string]bool, error)) error {
	if len(raw) > 1<<20 {
		return fmt.Errorf("domain patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return e
	}
	if len(doc.Content) != 1 {
		return fmt.Errorf("empty patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("patch requires one document")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return e
	}
	if e = p.shapeType(v, reflect.TypeOf(out).Elem(), ""); e != nil {
		return e
	}
	if e = strictInto(v, out); e != nil {
		return e
	}
	obj := v.(map[string]any)
	ops, _ := obj["operations"].([]any)
	for _, v := range ops {
		op := v.(map[string]any)
		a, _ := op["action"].(string)
		entity, _ := op["entity"].(string)
		allowed, e := fields(a, entity)
		if e != nil {
			return e
		}
		for k := range op {
			if !allowed[k] {
				return fmt.Errorf("%s/%s rejects authored field %s", a, entity, k)
			}
		}
	}
	return nil
}
func validDomainEnvelope(schema, want, hash, actor, reason, node string, count int) error {
	_, e := hex.DecodeString(hash)
	if schema != want || len(hash) != 64 || e != nil || strings.TrimSpace(actor) == "" || len(actor) > 256 || strings.TrimSpace(reason) == "" || len(reason) > 4096 || !stableID.MatchString(node) || count < 1 || count > 500 {
		return fmt.Errorf("patch requires schema, source SHA256, actor, reason, node_id and 1..500 operations")
	}
	return nil
}
func DecodeQuantitativePatch(raw []byte, file string) (QuantitativePatch, error) {
	var out QuantitativePatch
	e := decodeDomainPatch(raw, file, &out, quantitativeFields)
	if e == nil {
		e = validDomainEnvelope(out.Schema, QuantitativePatchSchema, out.ExpectedSourceSHA256, out.Actor, out.Reason, out.NodeID, len(out.Operations))
	}
	return out, e
}
func quantitativeNode(t *LocalTemplate, id string) (*Node, error) {
	list, i := findDiagramNode(&t.Nodes, id)
	if list == nil {
		return nil, fmt.Errorf("unknown quantitative node %s", id)
	}
	n := &(*list)[i]
	if n.Definition == nil || n.Definition.ID != "wmds/component/chart" || len(n.Nodes) > 0 {
		return nil, fmt.Errorf("quantitative requires a selected local chart component; fork/materialize the catalog example first")
	}
	return n, nil
}
func quantitativeSource(n *Node, values map[string]any) (map[string]any, map[string][]string, bool, error) {
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return nil, nil, bound, e
	}
	var source map[string]any
	if e = json.Unmarshal(canonical(args), &source); e != nil {
		return nil, nil, bound, e
	}
	// Normalize legacy positional scatter labels into their keyed point tuples.
	if source["kind"] == "scatter" {
		if labels, ok := source["labels"].(map[string]any); ok {
			points, _ := source["points"].([]any)
			for i, v := range points {
				point, ok := v.([]any)
				if ok && len(point) == 2 {
					if label, ok := labels[fmt.Sprint(i)].(string); ok {
						points[i] = append(point, label)
					}
				}
			}
			delete(source, "labels")
		}
	}
	keys := map[string][]string{}
	for path, v := range n.Keys {
		keys[strings.TrimPrefix(path, "/")] = append([]string(nil), v...)
	}
	for _, path := range []string{"categories", "series", "points", "items"} {
		if array, ok := source[path].([]any); ok {
			ids, e := cycleKeys(keys, path, len(array))
			if e != nil {
				return nil, nil, bound, e
			}
			keys[path] = ids
		}
	}
	return source, keys, bound, validateQuantitative(source, keys)
}
func validateQuantitative(source map[string]any, keys map[string][]string) error {
	switch source["kind"] {
	case "column", "bar", "line", "pie", "doughnut", "scatter", "quadrant":
	default:
		return fmt.Errorf("unsupported quantitative chart kind")
	}
	for _, path := range []string{"categories", "series", "points", "items"} {
		array, ok := source[path].([]any)
		if !ok {
			continue
		}
		if len(array) > 60 {
			return fmt.Errorf("quantitative %s exceeds 60", path)
		}
		ids, e := cycleKeys(keys, path, len(array))
		if e != nil || len(ids) != len(array) {
			return fmt.Errorf("quantitative keys mismatch %s", path)
		}
	}
	for _, path := range []string{"points", "items"} {
		array, _ := source[path].([]any)
		for _, v := range array {
			var coordinates []any
			if path == "points" {
				coordinates, _ = v.([]any)
				if len(coordinates) < 2 || len(coordinates) > 3 {
					return fmt.Errorf("scatter point requires x,y[,label]")
				}
				coordinates = coordinates[:2]
			} else {
				obj, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("quadrant item requires object")
				}
				coordinates = []any{obj["x"], obj["y"]}
			}
			for _, v := range coordinates {
				n, ok := v.(float64)
				if !ok || math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(n) > 1e9 {
					return fmt.Errorf("points require explicit observed finite coordinates; null is not zero")
				}
			}
		}
	}
	cats, catsOK := source["categories"].([]any)
	series, seriesOK := source["series"].([]any)
	if source["kind"] != "scatter" && source["kind"] != "quadrant" && source["progress"] == nil && (!catsOK || !seriesOK || len(cats) == 0 || len(series) == 0) {
		return fmt.Errorf("numerical chart requires nonempty categories and series")
	}
	for _, v := range series {
		s, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("series must be objects")
		}
		values, ok := s["values"].([]any)
		if !ok || len(values) != len(cats) {
			return fmt.Errorf("series values must match categories")
		}
		for _, v := range values {
			if v == nil {
				continue
			}
			n, ok := v.(float64)
			if !ok || math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(n) > 1e9 {
				return fmt.Errorf("quantitative values require finite numbers within 1e9 or null")
			}
		}
	}
	return nil
}
func InspectQuantitative(p *Project, slide, node, bundle, engine string) (QuantitativeInspection, error) {
	out := QuantitativeInspection{Schema: "pptxgengo.quantitative-inspection.v1", SlideID: slide, NodeID: node, SourceSHA256: p.SourceHash(), Contract: "Keyed source facts; source values and unit scale are authored, never inferred from geometry. Null differs from observed zero. Renderer validates kind-specific missing/scale/fit constraints."}
	idx, t, e := diagramSlide(p, slide)
	if e != nil {
		return out, e
	}
	n, e := quantitativeNode(&t, node)
	if e != nil {
		return out, e
	}
	var bound bool
	out.Source, out.Keys, bound, e = quantitativeSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Begin with explicit materialize/source or edit original bindings."
	}
	out.Geometry, e = InspectDiagram(p, slide, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func quantitativeArray(entity string) string {
	switch entity {
	case "category":
		return "categories"
	case "point":
		return "points"
	case "item":
		return "items"
	default:
		return "series"
	}
}
func applyQuantitative(source map[string]any, keys map[string][]string, op QuantitativeOperation) error {
	if op.Entity == "setting" {
		allowed := strings.Fields("title units source kind mode format valueSuffix yMin yMax xMin xMax allowMissing xTitle yTitle target progress center trend quadrants key style fill strongState positionMode holeSize preserveCategories preserveWorkbookZeros autoUpdateWorkbook colors highlight")
		if teamFind(allowed, op.Field) < 0 {
			return fmt.Errorf("unsupported quantitative setting %s", op.Field)
		}
		for path := range keys {
			if path == op.Field || strings.HasPrefix(path, op.Field+"/") {
				delete(keys, path)
			}
		}
		if op.Setting == nil {
			delete(source, op.Field)
		} else {
			source[op.Field] = op.Setting
		}
		return nil
	}
	if op.Entity == "value" {
		si := teamFind(keys["series"], op.Series)
		ci := teamFind(keys["categories"], op.Key)
		if si < 0 || ci < 0 {
			return fmt.Errorf("value requires existing series/category keys")
		}
		if (op.Value == nil) == !op.Missing {
			return fmt.Errorf("value requires numeric value or missing:true exclusively")
		}
		series, ok := source["series"].([]any)
		if !ok || si >= len(series) {
			return fmt.Errorf("value requires a numerical series source")
		}
		ss, ok := series[si].(map[string]any)
		if !ok {
			return fmt.Errorf("series must be an object")
		}
		vs, ok := ss["values"].([]any)
		if !ok || ci >= len(vs) {
			return fmt.Errorf("value requires aligned numerical source")
		}
		if op.Missing {
			vs[ci] = nil
		} else {
			vs[ci] = *op.Value
		}
		return nil
	}
	path := quantitativeArray(op.Entity)
	if (path == "points" && source["kind"] != "scatter") || (path == "items" && source["kind"] != "quadrant") || ((path == "categories" || path == "series") && (source["kind"] == "scatter" || source["kind"] == "quadrant" || source["progress"] != nil)) {
		return fmt.Errorf("entity %s is incompatible with selected chart source union", op.Entity)
	}
	items, _ := source[path].([]any)
	ids := keys[path]
	i := teamFind(ids, op.Key)
	if op.Action == "reorder" {
		ordered, orderedIDs, e := teamOrder(items, ids, op.Order)
		if e != nil {
			return e
		}
		if path == "categories" {
			for _, v := range source["series"].([]any) {
				s := v.(map[string]any)
				vals := s["values"].([]any)
				newVals := []any{}
				for _, key := range op.Order {
					newVals = append(newVals, vals[teamFind(ids, key)])
				}
				s["values"] = newVals
			}
			remapQuantitativeCategories(source, keys, ids, op.Order)
			remapQuantitativeValuesKeys(keys, ids, op.Order)
		}
		if path == "series" {
			remapQuantitativeColors(source, keys, ids, op.Order)
		}
		if path != "categories" {
			remapQuantitativeNestedKeys(keys, path, ids, orderedIDs)
		}
		source[path], keys[path] = ordered, orderedIDs
		return nil
	}
	if !stableID.MatchString(op.Key) {
		return fmt.Errorf("invalid quantitative key")
	}
	if op.Action == "remove" {
		if i < 0 {
			return fmt.Errorf("unknown %s key", op.Entity)
		}
		if !op.Cascade {
			return fmt.Errorf("removal requires cascade:true acknowledging lost observations/presentation")
		}
		newItems, newIDs := teamRemove(items, append([]string(nil), ids...), i)
		if path == "categories" {
			for _, v := range source["series"].([]any) {
				s := v.(map[string]any)
				vals := s["values"].([]any)
				s["values"] = append(vals[:i], vals[i+1:]...)
			}
			remapQuantitativeCategories(source, keys, ids, newIDs)
			remapQuantitativeValuesKeys(keys, ids, newIDs)
		}
		if path == "series" {
			remapQuantitativeColors(source, keys, ids, newIDs)
		}
		if path != "categories" {
			remapQuantitativeNestedKeys(keys, path, ids, newIDs)
		}
		source[path], keys[path] = newItems, newIDs
		return nil
	}
	var value any
	switch op.Entity {
	case "category":
		if strings.TrimSpace(op.Label) == "" {
			return fmt.Errorf("category requires label")
		}
		value = op.Label
	case "series":
		if op.Data == nil {
			return fmt.Errorf("series requires complete source data")
		}
		value = op.Data
	case "item":
		if op.Data == nil {
			return fmt.Errorf("quadrant item requires source data")
		}
		value = op.Data
	case "point":
		if len(op.Point) < 2 || len(op.Point) > 3 {
			return fmt.Errorf("scatter point requires x,y[,label]")
		}
		value = op.Point
	default:
		return fmt.Errorf("unsupported quantitative entity")
	}
	if i >= 0 {
		items[i] = value
	} else {
		items = append(items, value)
		oldIDs := append([]string(nil), ids...)
		ids = append(ids, op.Key)
		if path == "categories" {
			remapQuantitativeCategories(source, keys, oldIDs, ids)
			remapQuantitativeValuesKeys(keys, oldIDs, ids)
			for _, v := range source["series"].([]any) {
				s := v.(map[string]any)
				s["values"] = append(s["values"].([]any), nil)
			}
		}
	}
	source[path], keys[path] = items, ids
	return nil
}
func remapQuantitativeValuesKeys(keys map[string][]string, old, new []string) {
	for path, ids := range keys {
		if !strings.HasPrefix(path, "series/") || !strings.HasSuffix(path, "/values") {
			continue
		}
		out := []string{}
		for _, category := range new {
			j := teamFind(old, category)
			if j >= 0 && j < len(ids) {
				out = append(out, ids[j])
			} else {
				out = append(out, "value-"+category)
			}
		}
		keys[path] = out
	}
}
func remapQuantitativeNestedKeys(keys map[string][]string, path string, old, new []string) {
	children := map[string][]string{}
	for k, v := range keys {
		if strings.HasPrefix(k, path+"/") {
			children[k] = append([]string(nil), v...)
			delete(keys, k)
		}
	}
	for i, id := range new {
		j := teamFind(old, id)
		if j < 0 {
			continue
		}
		prefix := fmt.Sprintf("%s/%d", path, j)
		for k, v := range children {
			if k == prefix || strings.HasPrefix(k, prefix+"/") {
				keys[fmt.Sprintf("%s/%d", path, i)+strings.TrimPrefix(k, prefix)] = v
			}
		}
	}
}
func remapQuantitativeCategories(source map[string]any, keys map[string][]string, old, new []string) {
	if source["kind"] == "pie" || source["kind"] == "doughnut" {
		prior := keys["colors"]
		out := []string{}
		for _, key := range new {
			j := teamFind(old, key)
			id := "color-" + key
			if j >= 0 && j < len(prior) {
				id = prior[j]
			}
			out = append(out, id)
		}
		keys["colors"] = out
	}
	if highlights, ok := source["highlight"].([]any); ok {
		prior := keys["highlight"]
		out := []string{}
		for i, v := range highlights {
			f, ok := v.(float64)
			if !ok || f != math.Trunc(f) {
				continue
			}
			j := int(f)
			if j < 0 || j >= len(old) || teamFind(new, old[j]) < 0 {
				continue
			}
			id := "highlight-" + old[j]
			if i < len(prior) {
				id = prior[i]
			}
			out = append(out, id)
		}
		keys["highlight"] = out
	}

	if source["kind"] == "pie" || source["kind"] == "doughnut" {
		colors, _ := source["colors"].([]any)
		out := []any{}
		for next, key := range new {
			i := teamFind(old, key)
			colorIndex := i
			if colorIndex < 0 {
				colorIndex = next
			}
			var color any = fmt.Sprintf("series.%d", colorIndex%9+1)
			if i >= 0 && i < len(colors) {
				color = colors[i]
			}
			out = append(out, color)
		}
		source["colors"] = out
	}
	if hs, ok := source["highlight"].([]any); ok {
		out := []any{}
		for _, v := range hs {
			f, ok := v.(float64)
			if !ok {
				continue
			}
			i := int(f)
			if i >= 0 && i < len(old) {
				j := teamFind(new, old[i])
				if j >= 0 {
					out = append(out, float64(j))
				}
			}
		}
		source["highlight"] = out
	}
}
func remapQuantitativeColors(source map[string]any, keys map[string][]string, old, new []string) {
	if source["kind"] == "pie" || source["kind"] == "doughnut" {
		return
	}
	colors, _ := source["colors"].([]any)
	for i := range old {
		series := source["series"].([]any)
		if i >= len(series) {
			continue
		}
		item := series[i].(map[string]any)
		if i < len(colors) {
			item["color"] = colors[i]
		} else if item["color"] == nil || item["color"] == "" {
			item["color"] = fmt.Sprintf("series.%d", i+1)
		}
	}
	delete(source, "colors")
	delete(keys, "colors")
}
func PatchQuantitative(p *Project, slide string, patch QuantitativePatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeQuantitativePatch(canonical(patch), "quantitative-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if p.SourceHash() != patch.ExpectedSourceSHA256 {
		return empty, fmt.Errorf("quantitative source hash mismatch; inspect again")
	}
	idx, t, e := diagramSlide(p, slide)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := quantitativeNode(&clone, patch.NodeID)
	if e != nil {
		return empty, e
	}
	source, keys, bound, e := quantitativeSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	for i, op := range patch.Operations {
		if op.Action == "materialize" {
			if i != 0 {
				return empty, fmt.Errorf("materialize/source must be first")
			}
			bound = false
			continue
		}
		if op.Action == "replace" {
			if !op.Cascade || op.Data == nil || op.Keys == nil {
				return empty, fmt.Errorf("replace/source requires complete literal data, explicit collection keys and cascade:true")
			}
			for _, field := range []string{"type", "id", "x", "y", "w", "h", "_source_geometry"} {
				if _, ok := op.Data[field]; ok {
					return empty, fmt.Errorf("replacement source cannot alter placement or identity")
				}
			}
			bound = false
			geometry := source["_source_geometry"]
			source, keys = op.Data, op.Keys
			if geometry != nil {
				source["_source_geometry"] = geometry
			}
			if e = json.Unmarshal(canonical(source), &source); e != nil {
				return empty, e
			}
			for _, path := range []string{"categories", "series", "points", "items"} {
				if a, ok := source[path].([]any); ok && len(keys[path]) != len(a) {
					return empty, fmt.Errorf("replacement requires exact keys for %s", path)
				}
			}
			if e = validateQuantitative(source, keys); e != nil {
				return empty, e
			}
			continue
		}
		if bound {
			return empty, fmt.Errorf("quantitative refuses bound source without explicit materialize/source")
		}
		if e = applyQuantitative(source, keys, op); e != nil {
			return empty, e
		}
		if e = json.Unmarshal(canonical(source), &source); e != nil {
			return empty, e
		}
		if e = validateQuantitative(source, keys); e != nil {
			return empty, e
		}
	}
	// Convert all numeric authored fields to the same JSON representation before validation.
	if e = json.Unmarshal(canonical(source), &source); e != nil {
		return empty, e
	}
	if e = validateQuantitative(source, keys); e != nil {
		return empty, e
	}
	n.Arguments, n.Keys = source, keys
	if e = componentPopulateKeys(n, source); e != nil {
		return empty, e
	}
	return CompositionCandidate(p, slide, "quantitative", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
