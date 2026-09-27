package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func pointerParts(ptr string) ([]string, error) {
	if !strings.HasPrefix(ptr, "/") {
		return nil, fmt.Errorf("JSON pointer must begin /: %s", ptr)
	}
	parts := strings.Split(ptr[1:], "/")
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(p, "~1", "/"), "~0", "~")
	}
	return parts, nil
}
func pointerSet(root any, ptr string, value any) error {
	parts, err := pointerParts(ptr)
	if err != nil {
		return err
	}
	cur := root
	for i, p := range parts {
		last := i == len(parts)-1
		switch node := cur.(type) {
		case map[string]any:
			old, ok := node[p]
			if !ok {
				return fmt.Errorf("slot pointer %s missing key %s", ptr, p)
			}
			if last {
				node[p] = value
				return nil
			}
			cur = old
		case []any:
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 || n >= len(node) {
				return fmt.Errorf("slot pointer %s invalid index %s", ptr, p)
			}
			if last {
				node[n] = value
				return nil
			}
			cur = node[n]
		default:
			return fmt.Errorf("slot pointer %s enters scalar", ptr)
		}
	}
	return fmt.Errorf("empty slot pointer")
}
func pointerGet(root any, ptr string) (any, error) {
	parts, err := pointerParts(ptr)
	if err != nil {
		return nil, err
	}
	cur := root
	for _, p := range parts {
		switch node := cur.(type) {
		case map[string]any:
			next, ok := node[p]
			if !ok {
				return nil, fmt.Errorf("pointer %s missing key %s", ptr, p)
			}
			cur = next
		case []any:
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 || n >= len(node) {
				return nil, fmt.Errorf("pointer %s invalid index %s", ptr, p)
			}
			cur = node[n]
		default:
			return nil, fmt.Errorf("pointer %s enters scalar", ptr)
		}
	}
	return cur, nil
}
func validContentPointer(ptr string) bool {
	parts, err := pointerParts(ptr)
	if err != nil || len(parts) == 0 {
		return false
	}
	switch parts[len(parts)-1] {
	case "text", "label", "title", "role", "takeaway", "notes", "labels", "roles", "paragraphs", "phrase", "page", "page_label":
		return true
	}
	return false
}

type ValuesFile struct {
	Slots            map[string]json.RawMessage `json:"slots"`
	NarrativePath    string                     `json:"narrative_path,omitempty"`
	NarrativeSlideID string                     `json:"narrative_slide_id,omitempty"`
	StyleVariant     string                     `json:"style_variant,omitempty"`
}
type SelectionTrace struct {
	Schema           string          `json:"schema"`
	ContractID       string          `json:"contract_id"`
	ContractVersion  string          `json:"contract_version"`
	ContractSHA256   string          `json:"contract_sha256"`
	TemplateSHA256   string          `json:"template_sha256"`
	SourceSHA256     string          `json:"source_sha256"`
	Qualification    string          `json:"qualification"`
	Preference       string          `json:"preference"`
	CanonicalID      string          `json:"canonical_id"`
	Purpose          string          `json:"purpose"`
	ContentRoles     []string        `json:"content_roles"`
	SelectionReason  string          `json:"selection_reason"`
	Proof            []Artifact      `json:"proof,omitempty"`
	FitEnvelope      FitEnvelope     `json:"fit_envelope"`
	SlotNames        []string        `json:"slot_names"`
	StyleVariant     string          `json:"style_variant,omitempty"`
	NarrativeSlideID string          `json:"narrative_slide_id,omitempty"`
	NarrativeSHA256  string          `json:"narrative_sha256,omitempty"`
	NarrativeBrief   *NarrativeBrief `json:"narrative_brief,omitempty"`
	NarrativeSources []Artifact      `json:"narrative_sources,omitempty"`
	EvidenceRefs     []string        `json:"evidence_refs,omitempty"`
	FitStatus        string          `json:"fit_status"`
	RequiredNext     []string        `json:"required_next"`
}

func repeatValue(slot Slot, values []string) (any, error) {
	if len(slot.ItemTemplate) == 0 {
		return values, nil
	}
	var out []any
	for i, v := range values {
		var item any
		if err := json.Unmarshal(slot.ItemTemplate, &item); err != nil {
			return nil, err
		}
		if err := pointerSet(item, slot.ItemValuePointer, v); err != nil {
			return nil, err
		}
		if slot.ItemIDPointer != "" {
			if err := pointerSet(item, slot.ItemIDPointer, fmt.Sprintf("%s%d", slot.ItemIDPrefix, i+1)); err != nil {
				return nil, err
			}
		}
		out = append(out, item)
	}
	return out, nil
}
func (s Store) Instantiate(indexPath, id, valuesPath, out string, allowUnqualified bool) (SelectionTrace, error) {
	ins, err := s.Inspect(indexPath, id)
	if err != nil {
		return SelectionTrace{}, err
	}
	c := ins.Contract
	if c.Qualification.State != "adaptation_qualified" && !allowUnqualified {
		return SelectionTrace{}, fmt.Errorf("contract %s is %s; --allow-unqualified required for experiment", id, c.Qualification.State)
	}
	if c.Preference.Value == "avoid" && !allowUnqualified {
		return SelectionTrace{}, fmt.Errorf("avoided contract requires explicit experimental override")
	}
	if _, err := os.Stat(out); err == nil {
		return SelectionTrace{}, fmt.Errorf("output exists: %s", out)
	} else if !os.IsNotExist(err) {
		return SelectionTrace{}, err
	}
	vb, err := os.ReadFile(valuesPath)
	if err != nil {
		return SelectionTrace{}, err
	}
	var vf ValuesFile
	if err := json.Unmarshal(vb, &vf); err != nil {
		return SelectionTrace{}, err
	}
	if vf.Slots == nil {
		return SelectionTrace{}, fmt.Errorf("values.slots required")
	}
	p, err := s.SafePath(c.Composition.SpecPath)
	if err != nil {
		return SelectionTrace{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return SelectionTrace{}, err
	}
	if hashBytes(b) != c.Composition.SpecSHA256 {
		return SelectionTrace{}, fmt.Errorf("stale template hash")
	}
	var spec map[string]any
	if err := json.Unmarshal(b, &spec); err != nil {
		return SelectionTrace{}, err
	}
	slides, ok := spec["slides"].([]any)
	if !ok {
		return SelectionTrace{}, fmt.Errorf("template slides missing")
	}
	var slide map[string]any
	for _, x := range slides {
		m, _ := x.(map[string]any)
		if m["id"] == c.Composition.SlideID {
			slide = m
			break
		}
	}
	if slide == nil {
		return SelectionTrace{}, fmt.Errorf("template slide %s missing", c.Composition.SlideID)
	}
	known := map[string]bool{}
	var applied []string
	var allCopy strings.Builder
	for _, slot := range c.Composition.Slots {
		known[slot.Name] = true
		raw, has := vf.Slots[slot.Name]
		if !has {
			if slot.Required {
				return SelectionTrace{}, fmt.Errorf("required slot %s missing", slot.Name)
			}
			continue
		}
		if !validContentPointer(slot.Pointer) {
			return SelectionTrace{}, fmt.Errorf("slot %s targets non-content pointer", slot.Name)
		}
		var value any
		switch slot.ValueType {
		case "string":
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				return SelectionTrace{}, fmt.Errorf("slot %s requires string", slot.Name)
			}
			if strings.TrimSpace(v) == "" || (slot.MaxChars > 0 && len([]rune(v)) > slot.MaxChars) {
				return SelectionTrace{}, fmt.Errorf("slot %s exceeds bounded copy or is empty", slot.Name)
			}
			value = v
			allCopy.WriteString(v + "\n")
		case "string_array":
			var v []string
			if err := json.Unmarshal(raw, &v); err != nil {
				return SelectionTrace{}, fmt.Errorf("slot %s requires string array", slot.Name)
			}
			if len(v) < slot.MinItems || len(v) > slot.MaxItems {
				return SelectionTrace{}, fmt.Errorf("slot %s count outside [%d,%d]", slot.Name, slot.MinItems, slot.MaxItems)
			}
			for _, x := range v {
				if strings.TrimSpace(x) == "" || (slot.MaxChars > 0 && len([]rune(x)) > slot.MaxChars) {
					return SelectionTrace{}, fmt.Errorf("slot %s item exceeds bounded copy or is empty", slot.Name)
				}
				allCopy.WriteString(x + "\n")
			}
			value, err = repeatValue(slot, v)
			if err != nil {
				return SelectionTrace{}, fmt.Errorf("slot %s: %w", slot.Name, err)
			}
		default:
			return SelectionTrace{}, fmt.Errorf("unsupported slot type")
		}
		if err := pointerSet(slide, slot.Pointer, value); err != nil {
			return SelectionTrace{}, err
		}
		applied = append(applied, slot.Name)
	}
	for name := range vf.Slots {
		if !known[name] {
			return SelectionTrace{}, fmt.Errorf("unknown slot %s", name)
		}
	}
	if vf.StyleVariant != "" {
		found := false
		for _, variant := range c.StyleVariants {
			if variant.ID == vf.StyleVariant {
				found = true
				if !variant.Qualified && !allowUnqualified {
					return SelectionTrace{}, fmt.Errorf("unqualified style variant %s", variant.ID)
				}
				replaceTokens(slide, variant.Tokens)
			}
		}
		if !found {
			return SelectionTrace{}, fmt.Errorf("unknown style variant %s", vf.StyleVariant)
		}
	}
	trace := SelectionTrace{Schema: "pptxgengo.library-instantiation.v1", ContractID: id, ContractVersion: c.Version, ContractSHA256: ins.SHA256, TemplateSHA256: c.Composition.SpecSHA256, SourceSHA256: c.Source.SourceSHA256, Qualification: c.Qualification.State, Preference: c.Preference.Value, CanonicalID: ins.CanonicalID, Purpose: c.Purpose, ContentRoles: c.ContentRoles, SelectionReason: "Explicit contract ID selected; semantic slots and cardinality checked against the portable contract.", Proof: c.Qualification.Evidence, FitEnvelope: c.FitEnvelope, SlotNames: applied, StyleVariant: vf.StyleVariant, FitStatus: "unmeasured_changed_copy", RequiredNext: []string{"pptxcompose probe", "native measurement", "fit-report", "build", "verify", "visual review"}}
	if vf.NarrativePath != "" {
		narrBytes, err := os.ReadFile(vf.NarrativePath)
		if err != nil {
			return SelectionTrace{}, err
		}
		narr, err := s.LoadNarrative(vf.NarrativePath)
		if err != nil {
			return SelectionTrace{}, err
		}
		ns, err := narr.Slide(vf.NarrativeSlideID)
		if err != nil {
			return SelectionTrace{}, err
		}
		for _, detail := range append(append([]string{}, ns.RequiredDetail...), ns.Qualifications...) {
			if !strings.Contains(allCopy.String(), detail) {
				return SelectionTrace{}, fmt.Errorf("narrative detail/qualification omitted from slots: %q", detail)
			}
		}
		bindings := map[string]string{"assertion_title": "/title", "role": "/role", "takeaway": "/takeaway"}
		for k, pointer := range c.Composition.NarrativeBindings {
			bindings[k] = pointer
		}
		for field, want := range map[string]string{"assertion_title": ns.AssertionTitle, "role": ns.Role, "takeaway": ns.Takeaway} {
			gotValue, e := pointerGet(slide, bindings[field])
			if e != nil {
				return SelectionTrace{}, e
			}
			got, _ := gotValue.(string)
			if strings.TrimSpace(got) != strings.TrimSpace(want) {
				return SelectionTrace{}, fmt.Errorf("narrative %s does not match instantiated slide binding %s", field, bindings[field])
			}
		}
		trace.NarrativeSlideID = ns.ID
		trace.NarrativeSHA256 = hashBytes(narrBytes)
		trace.NarrativeBrief = &narr.Brief
		trace.NarrativeSources = append(trace.NarrativeSources, narr.Brief.SourcePacket...)
		for _, e := range narr.Evidence {
			for _, id := range ns.EvidenceRefs {
				if e.ID == id {
					trace.NarrativeSources = append(trace.NarrativeSources, e.Source)
				}
			}
		}
		trace.EvidenceRefs = ns.EvidenceRefs
	}
	spec["slides"] = []any{slide}
	outSpec, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return SelectionTrace{}, err
	}
	if err := os.Mkdir(out, 0755); err != nil {
		return SelectionTrace{}, err
	}
	if err := os.WriteFile(filepath.Join(out, "spec.json"), append(outSpec, '\n'), 0644); err != nil {
		return SelectionTrace{}, err
	}
	tb, _ := json.MarshalIndent(trace, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "selection-trace.json"), append(tb, '\n'), 0644); err != nil {
		return SelectionTrace{}, err
	}
	return trace, nil
}
func replaceTokens(v any, tokens map[string]string) {
	colorField := map[string]bool{"background": true, "foreground": true, "contrast_background": true, "surface": true, "accent": true, "outline_color": true, "title_foreground": true, "color": true}
	switch x := v.(type) {
	case map[string]any:
		for k, q := range x {
			if str, ok := q.(string); ok {
				if replacement, found := tokens[str]; found && colorField[k] {
					x[k] = replacement
				}
			} else {
				replaceTokens(q, tokens)
			}
		}
	case []any:
		for _, q := range x {
			replaceTokens(q, tokens)
		}
	}
}
