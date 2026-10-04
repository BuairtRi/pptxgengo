package deckproject

import (
	"fmt"
	"math"
	"reflect"
	"unicode/utf8"
)

// validateSchema deliberately rejects schema vocabulary the runtime cannot enforce.
func validateSchema(s map[string]any) error {
	allowed := map[string]bool{"type": true, "properties": true, "required": true, "additionalProperties": true, "items": true, "minItems": true, "maxItems": true, "minLength": true, "maxLength": true, "minimum": true, "maximum": true, "enum": true, "const": true, "description": true}
	for k := range s {
		if !allowed[k] {
			return fmt.Errorf("unsupported zone schema keyword %q", k)
		}
	}
	t, ok := s["type"].(string)
	if !ok {
		return fmt.Errorf("zone schema requires a single type")
	}
	switch t {
	case "string", "number", "integer", "boolean", "null", "object", "array":
	default:
		return fmt.Errorf("unsupported schema type %q", t)
	}
	specific := map[string]string{"properties": "object", "required": "object", "additionalProperties": "object", "items": "array", "minItems": "array", "maxItems": "array", "minLength": "string", "maxLength": "string"}
	for key, kind := range specific {
		if _, ok := s[key]; ok && t != kind {
			return fmt.Errorf("schema keyword %s requires type %s", key, kind)
		}
	}
	for _, key := range []string{"minimum", "maximum"} {
		if _, ok := s[key]; ok && t != "number" && t != "integer" {
			return fmt.Errorf("%s requires numeric schema", key)
		}
	}
	if v, ok := s["description"]; ok {
		if _, ok := v.(string); !ok {
			return fmt.Errorf("description must be a string")
		}
	}
	if v, ok := s["additionalProperties"]; ok {
		if v != false {
			return fmt.Errorf("only additionalProperties:false supported")
		}
	}
	for _, k := range []string{"minItems", "maxItems", "minLength", "maxLength", "minimum", "maximum"} {
		if v, ok := s[k]; ok {
			n, ok := v.(float64)
			if !ok {
				return fmt.Errorf("%s must be numeric", k)
			}
			if k != "minimum" && k != "maximum" && (n < 0 || n != math.Trunc(n)) {
				return fmt.Errorf("%s requires nonnegative integer", k)
			}
		}
	}
	if v, ok := s["enum"]; ok {
		a, ok := v.([]any)
		if !ok || len(a) == 0 {
			return fmt.Errorf("enum must be nonempty array")
		}
	}
	if v, ok := s["properties"]; ok {
		m, ok := v.(map[string]any)
		if !ok || t != "object" {
			return fmt.Errorf("properties requires object schema")
		}
		for _, v := range m {
			c, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("property schema must be an object")
			}
			if e := validateSchema(c); e != nil {
				return e
			}
		}
	}
	if v, ok := s["required"]; ok {
		a, ok := v.([]any)
		if !ok || t != "object" {
			return fmt.Errorf("required requires object array")
		}
		seen := map[string]bool{}
		for _, v := range a {
			k, ok := v.(string)
			if !ok || seen[k] {
				return fmt.Errorf("invalid required fields")
			}
			seen[k] = true
		}
	}
	if v, ok := s["items"]; ok {
		c, ok := v.(map[string]any)
		if !ok || t != "array" {
			return fmt.Errorf("items requires array schema")
		}
		return validateSchema(c)
	}
	return nil
}
func validateValue(s map[string]any, v any, path string) error {
	t := s["type"].(string)
	good := false
	switch t {
	case "null":
		good = v == nil
	case "string":
		_, good = v.(string)
	case "number":
		_, good = v.(float64)
	case "integer":
		f, ok := v.(float64)
		good = ok && f == math.Trunc(f)
	case "boolean":
		_, good = v.(bool)
	case "object":
		_, good = v.(map[string]any)
	case "array":
		_, good = v.([]any)
	}
	if !good {
		return fmt.Errorf("%s must be %s", path, t)
	}
	if c, ok := s["const"]; ok && !reflect.DeepEqual(c, v) {
		return fmt.Errorf("%s does not match const", path)
	}
	if es, ok := s["enum"].([]any); ok {
		match := false
		for _, e := range es {
			match = match || reflect.DeepEqual(e, v)
		}
		if !match {
			return fmt.Errorf("%s does not match enum", path)
		}
	}
	limit := func(key string, n float64, minimum bool) error {
		if bound, ok := s[key].(float64); ok && (minimum && n < bound || !minimum && n > bound) {
			return fmt.Errorf("%s violates %s", path, key)
		}
		return nil
	}
	var checks []error
	switch val := v.(type) {
	case string:
		n := float64(utf8.RuneCountInString(val))
		checks = []error{limit("minLength", n, true), limit("maxLength", n, false)}
	case float64:
		checks = []error{limit("minimum", val, true), limit("maximum", val, false)}
	case []any:
		checks = []error{limit("minItems", float64(len(val)), true), limit("maxItems", float64(len(val)), false)}
		if c, ok := s["items"].(map[string]any); ok {
			for i, item := range val {
				if e := validateValue(c, item, fmt.Sprintf("%s/%d", path, i)); e != nil {
					return e
				}
			}
		}
	case map[string]any:
		props, _ := s["properties"].(map[string]any)
		if req, ok := s["required"].([]any); ok {
			for _, k := range req {
				if _, ok := val[k.(string)]; !ok {
					return fmt.Errorf("%s missing %s", path, k)
				}
			}
		}
		for k, item := range val {
			if c, ok := props[k].(map[string]any); ok {
				if e := validateValue(c, item, path+"/"+escape(k)); e != nil {
					return e
				}
			} else if s["additionalProperties"] == false {
				return fmt.Errorf("%s unknown field %s", path, k)
			}
		}
	}
	for _, e := range checks {
		if e != nil {
			return e
		}
	}
	return nil
}
