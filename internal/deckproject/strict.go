package deckproject

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

func (p *Project) strictShape() error {
	if raw, ok := p.tree["media_optimization"].(map[string]any); ok {
		for _, key := range []string{"deduplicate", "resize_jpeg", "compression", "pixels_per_inch", "jpeg_quality", "min_savings_percent"} {
			if _, ok := raw[key]; !ok {
				return p.fail("/media_optimization", "explicit override requires %s", key)
			}
		}
	}
	// Variant-specific field lists prevent accepted JSON fields from being silently ignored.
	var walk func([]any, string) error
	walk = func(nodes []any, path string) error {
		for i, v := range nodes {
			m, ok := v.(map[string]any)
			if !ok {
				return p.fail(path, "node must be mapping")
			}
			ptr := fmt.Sprintf("%s/%d", path, i)
			kind, _ := m["kind"].(string)
			fields := "id kind placement"
			switch kind {
			case "text":
				fields += " style ink align text"
			case "box":
				fields += " surface border"
			case "image":
				fields += " asset fit rotation_deg"
			case "rule":
				fields += " ink weight_pt"
			case "group":
				fields = "id kind nodes"
			case "component", "composite":
				fields += " definition arguments keys"
			}
			allowed := map[string]bool{}
			for _, k := range strings.Fields(fields) {
				allowed[k] = true
			}
			for k := range m {
				if !allowed[k] {
					return p.fail(ptr+"/"+escape(k), "field %s is not valid for %s", k, kind)
				}
			}
			if kind == "group" {
				a, ok := m["nodes"].([]any)
				if !ok {
					return p.fail(ptr, "group requires nodes")
				}
				if e := walk(a, ptr+"/nodes"); e != nil {
					return e
				}
			}
		}
		return nil
	}
	ts, _ := p.tree["local_templates"].(map[string]any)
	for id, v := range ts {
		m := v.(map[string]any)
		ptr := "/local_templates/" + escape(id)
		zs, _ := m["zones"].(map[string]any)
		for name, v := range zs {
			z := v.(map[string]any)
			if _, ok := z["required"]; !ok {
				return p.fail(ptr+"/zones/"+escape(name), "zone required boolean must be explicit")
			}
		}
		ns, _ := m["nodes"].([]any)
		if e := walk(ns, ptr+"/nodes"); e != nil {
			return e
		}
	}
	return nil
}

// shapeType checks the declared model before JSON decoding, retaining exact YAML
// positions instead of guessing from a decoder's flattened field names.
func (p *Project) shapeType(v any, t reflect.Type, path string) error {
	if t.Kind() == reflect.Pointer {
		if v == nil {
			return nil
		}
		return p.shapeType(v, t.Elem(), path)
	}
	switch t.Kind() {
	case reflect.Interface:
		return nil
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return p.fail(path, "expected mapping")
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name != "" && name != "-" {
				fields[name] = f.Type
			}
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			ft, ok := fields[k]
			if !ok {
				return p.fail(path+"/"+escape(k), "unknown field %q", k)
			}
			if e := p.shapeType(m[k], ft, path+"/"+escape(k)); e != nil {
				return e
			}
		}
	case reflect.Map:
		m, ok := v.(map[string]any)
		if !ok {
			return p.fail(path, "expected mapping")
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if e := p.shapeType(m[k], t.Elem(), path+"/"+escape(k)); e != nil {
				return e
			}
		}
	case reflect.Slice:
		a, ok := v.([]any)
		if !ok {
			return p.fail(path, "expected array")
		}
		for i, item := range a {
			if e := p.shapeType(item, t.Elem(), fmt.Sprintf("%s/%d", path, i)); e != nil {
				return e
			}
		}
	case reflect.String:
		if _, ok := v.(string); !ok {
			return p.fail(path, "expected string; quote numeric/date-like text")
		}
	case reflect.Bool:
		if _, ok := v.(bool); !ok {
			return p.fail(path, "expected boolean")
		}
	case reflect.Int, reflect.Int64, reflect.Int32:
		f, ok := v.(float64)
		if !ok || math.Trunc(f) != f {
			return p.fail(path, "expected integer")
		}
	case reflect.Float32, reflect.Float64:
		if _, ok := v.(float64); !ok {
			return p.fail(path, "expected number")
		}
	default:
		return p.fail(path, "unsupported model type %s", t)
	}
	return nil
}
