package deckproject

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type SwapField struct {
	Source      string `json:"source"`
	Destination string `json:"destination,omitempty"`
	Value       any    `json:"value"`
	Basis       string `json:"basis"`
}
type SwapProposal struct {
	Schema       string               `json:"schema"`
	SlideID      string               `json:"slide_id"`
	SourceSHA256 string               `json:"source_sha256"`
	From         Reference            `json:"from"`
	To           Reference            `json:"to"`
	Mapped       []SwapField          `json:"mapped"`
	Unmapped     []SwapField          `json:"unmapped"`
	Missing      []string             `json:"missing"`
	Patch        map[string]SlideEdit `json:"patch"`
	Policy       []string             `json:"policy"`
}

// ProposeSwap carries only compatible structural roles. It never silently
// clears old copy or borrows business copy from the target's source specimen.
func ProposeSwap(p *Project, id, key, bundle string) (SwapProposal, error) {
	r := SwapProposal{Schema: "pptxgengo.template-swap.v1", SlideID: id, SourceSHA256: p.SourceHash(), Mapped: []SwapField{}, Unmapped: []SwapField{}, Missing: []string{}, Policy: []string{"Proposal only; no source changed.", "Matching roles and indices propose placement, not semantic or visual acceptance.", "Unmapped copy blocks application unless explicitly allowed; source notes and visibility are retained."}}
	var slide *Slide
	for i := range p.Document.Slides {
		if p.Document.Slides[i].ID == id {
			slide = &p.Document.Slides[i]
			break
		}
	}
	if slide == nil {
		return r, fmt.Errorf("unknown slide ID %s", id)
	}
	if slide.Template.Scope != "shared" {
		return r, fmt.Errorf("automatic swap requires a shared source template; supply a complete replacement patch for a local composition")
	}
	catalog, err := wmdesign.LibraryCatalog(bundle, "")
	if err != nil {
		return r, err
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	for _, d := range catalog {
		defs[d.Key] = d
	}
	oldDef, ok := defs[slide.Template.ID]
	if !ok {
		return r, fmt.Errorf("unknown source template")
	}
	newDef, ok := defs[key]
	if !ok {
		return r, fmt.Errorf("unknown target template %s", key)
	}
	old, err := StockEditableSlide(*slide, oldDef)
	if err != nil {
		return r, err
	}
	next, err := StockScaffoldSlide(bundle, key, id, p.Document.Year)
	if err != nil {
		return r, err
	}
	oldContent, _ := old["content"].(map[string]any)
	nextContent, _ := next["content"].(map[string]any)
	oldMeta, err := wmdesign.LibraryAuthoringMetadata(oldDef)
	if err != nil {
		return r, err
	}
	nextMeta, err := wmdesign.LibraryAuthoringMetadata(newDef)
	if err != nil {
		return r, err
	}
	oldSlots := swapAuthoredSlots(old, oldDef, oldMeta)
	nextSlots := swapAuthoredSlots(next, newDef, nextMeta)
	oldLeaves := map[string]any{}
	flattenAuthored(oldContent, "", oldLeaves)
	nextLeaves := map[string]any{}
	flattenAuthored(nextContent, "", nextLeaves)
	keys := []string{}
	for path := range nextLeaves {
		keys = append(keys, path)
	}
	sort.Strings(keys)
	used := map[string]bool{}
	sameTemplate := oldDef.Key == newDef.Key
	for _, path := range keys {
		destination, known := nextSlots[path]
		value, present := oldLeaves[path]
		source, sourceKnown := oldSlots[path]
		sourcePath := path
		basis := "same_role_group_index"
		compatible := present && ((sameTemplate && reflect.TypeOf(value) == reflect.TypeOf(nextLeaves[path])) || (known && sourceKnown && source.Role == destination.Role && source.Kind == destination.Kind))
		if compatible && !sameTemplate && destination.Group != "" {
			compatible = source.Group == destination.Group && source.GroupIndex == destination.GroupIndex && !strings.HasPrefix(destination.Group, "/paragraphs") && !strings.HasPrefix(destination.Group, "/panels")
		}
		if sameTemplate {
			basis = "same_template_exact_field"
		}
		// Headline/statement recipes name a unique title-bearing field. A role
		// bridge is safe only when there is one available source and destination;
		// generic paragraphs are never promoted to headlines by their copy.
		if !compatible && known && destination.Role == "headline" {
			var matches []string
			for candidate, slot := range oldSlots {
				if slot.Role == "headline" && slot.Kind == destination.Kind && !used[candidate] {
					if _, exists := oldLeaves[candidate]; exists {
						matches = append(matches, candidate)
					}
				}
			}
			destinations := 0
			for candidate, slot := range nextSlots {
				if slot.Role == "headline" {
					if _, exists := nextLeaves[candidate]; exists {
						destinations++
					}
				}
			}
			if len(matches) == 1 && destinations == 1 {
				sourcePath = matches[0]
				value = oldLeaves[sourcePath]
				compatible = true
				basis = "unique_headline_role"
			}
		}
		if compatible {
			if err := replaceContentLeaf(nextContent, path, value); err != nil {
				return r, err
			}
			used[sourcePath] = true
			r.Mapped = append(r.Mapped, SwapField{sourcePath, path, value, basis})
		} else {
			if known && destination.Role == "structural_key" {
				// New items retain the scaffold's generated identity. It is not
				// visible business copy and does not satisfy a missing text slot.
				continue
			}
			blank := ""
			if _, ok := nextLeaves[path].(string); !ok {
				return r, fmt.Errorf("automatic swap cannot initialize non-text field %s; use a supplied replacement", path)
			}
			if err := replaceContentLeaf(nextContent, path, blank); err != nil {
				return r, err
			}
			if !known || !destination.AllowEmpty {
				r.Missing = append(r.Missing, path)
			}
		}
	}
	oldPaths := []string{}
	for path := range oldLeaves {
		oldPaths = append(oldPaths, path)
	}
	sort.Strings(oldPaths)
	for _, path := range oldPaths {
		value := oldLeaves[path]
		if !used[path] && !reflect.DeepEqual(value, "") {
			r.Unmapped = append(r.Unmapped, SwapField{Source: path, Value: value, Basis: "no_compatible_target_field"})
		}
	}
	// Navigation labels are authored material even though they are stored beside
	// structural values rather than in the human content projection.
	if nav, exists := slide.Values["nav"]; exists {
		r.Unmapped = append(r.Unmapped, SwapField{Source: "/values/nav", Value: nav, Basis: "navigation_requires_explicit_mapping"})
	}
	if oldDef.Key != newDef.Key {
		for _, field := range []string{"emphasis", "whiteboard"} {
			if value, exists := slide.Values[field]; exists {
				r.Unmapped = append(r.Unmapped, SwapField{Source: "/values/" + field, Value: value, Basis: "technical_setting_requires_explicit_mapping"})
			}
		}
	}
	if values, ok := next["values"].(map[string]any); ok {
		if _, exists := values["nav"]; exists {
			return r, fmt.Errorf("target navigation requires supplied values; use a complete replacement patch")
		}
		// Array keys identify authored items. Preserve them only for the same
		// closed array contract; they are never replacement business copy.
		if oldKeys, ok := slide.Values["keys"].(map[string]any); ok {
			if nextKeys, ok := values["keys"].(map[string]any); ok {
				for _, destination := range newDef.Arrays {
					for _, source := range oldDef.Arrays {
						if source.Name == destination.Name && source.SourcePointer == destination.SourcePointer {
							oldItems, oldOK := oldKeys[source.Name].([]any)
							nextItems, nextOK := nextKeys[destination.Name].([]any)
							if oldOK && nextOK {
								for i := 0; i < len(oldItems) && i < len(nextItems); i++ {
									nextItems[i] = oldItems[i]
								}
							}
						}
					}
				}
			}
		}
	}
	if oldDef.Key == newDef.Key {
		// A same-template proposal must preserve the complete technical contract,
		// including caller emphasis/whiteboard settings and hidden empty slots.
		if values, ok := old["values"].(map[string]any); ok {
			next["values"] = values
		} else {
			delete(next, "values")
		}
	}
	var edit SlideEdit
	if err := strictInto(map[string]any{"template": next["template"], "values": next["values"], "content": nextContent, "bindings": next["bindings"]}, &edit); err != nil {
		return r, err
	}
	if edit.Values == nil {
		edit.Values = map[string]any{}
	}
	r.From = slide.Template
	r.To = *edit.Template
	r.Patch = map[string]SlideEdit{id: edit}
	return r, nil
}

func ApplySwap(p *Project, proposal SwapProposal, allowUnmapped bool, bundle, engine string) (EditReceipt, error) {
	if proposal.Schema != "pptxgengo.template-swap.v1" || proposal.SourceSHA256 != p.SourceHash() {
		return EditReceipt{}, fmt.Errorf("swap proposal is stale or unsupported; regenerate against current source")
	}
	verified, err := ProposeSwap(p, proposal.SlideID, proposal.To.ID, bundle)
	if err != nil {
		return EditReceipt{}, err
	}
	if !reflect.DeepEqual(proposal.From, verified.From) || !reflect.DeepEqual(proposal.To, verified.To) || !reflect.DeepEqual(proposal.Patch, verified.Patch) || !reflect.DeepEqual(proposal.Mapped, verified.Mapped) || !reflect.DeepEqual(proposal.Unmapped, verified.Unmapped) || !reflect.DeepEqual(proposal.Missing, verified.Missing) {
		return EditReceipt{}, fmt.Errorf("swap proposal differs from verified source mapping; regenerate or use a supplied edit patch")
	}
	if len(proposal.Unmapped) > 0 && !allowUnmapped {
		return EditReceipt{}, fmt.Errorf("swap has %d unmapped content fields; retain them in supplied copy or explicitly pass --allow-unmapped", len(proposal.Unmapped))
	}
	if len(proposal.Missing) > 0 {
		return EditReceipt{}, fmt.Errorf("swap requires %d target fields; use a supplied edit patch to fill them", len(proposal.Missing))
	}
	if len(proposal.Patch) != 1 {
		return EditReceipt{}, fmt.Errorf("swap patch requires exactly one slide")
	}
	edit, exists := proposal.Patch[proposal.SlideID]
	if !exists || edit.Template == nil || *edit.Template != proposal.To {
		return EditReceipt{}, fmt.Errorf("swap patch identity mismatch")
	}
	return EditSlidesWithOptions(p, proposal.Patch, bundle, engine, EditOptions{Operation: "swap"})
}

func flattenAuthored(value any, path string, out map[string]any) {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			flattenAuthored(child, path+"/"+escape(key), out)
		}
	case []any:
		for i, child := range v {
			flattenAuthored(child, fmt.Sprintf("%s/%d", path, i), out)
		}
	default:
		out[path] = value
	}
}

// Metadata aliases describe the stable library projection. Bindings describe
// the actual authored projection (including the typed optional eyebrow field).
// Resolve through those bindings rather than assuming the two paths coincide.
func swapAuthoredSlots(authored map[string]any, def wmdesign.LibraryTemplate, metadata wmdesign.LibraryAuthoring) map[string]wmdesign.LibraryAuthoringSlot {
	bindings, _ := authored["bindings"].(map[string]any)
	out := map[string]wmdesign.LibraryAuthoringSlot{}
	for _, slot := range metadata.Slots {
		for alias, raw := range bindings {
			target, _ := raw.(string)
			path := alias
			if !strings.HasPrefix(path, "/") {
				path = "/" + escape(path)
			}
			if target == "/slots/"+escape(slot.Name) {
				slot.Alias = path
				out[path] = slot
				break
			}
			if def.ContentContract == wmdesign.TemplateBindingsContract && (slot.SourcePointer == target || strings.HasPrefix(slot.SourcePointer, target+"/")) {
				slot.Alias = path + strings.TrimPrefix(slot.SourcePointer, target)
				out[slot.Alias] = slot
				break
			}
		}
	}
	// Card keys are structural item identity, not business copy. An exact-count
	// typed card contract permits their preservation at the same group index.
	if def.ValueSchema != nil && def.ValueSchema.GoType == "BoundCardRowsContent" {
		count := def.ValueSchema.ExactCounts["cards"]
		for alias, raw := range bindings {
			if raw != "/cards" {
				continue
			}
			if !strings.HasPrefix(alias, "/") {
				alias = "/" + escape(alias)
			}
			for i := 0; i < count; i++ {
				path := fmt.Sprintf("%s/%d/key", alias, i)
				out[path] = wmdesign.LibraryAuthoringSlot{Alias: path, Role: "structural_key", Kind: "string", Group: "/cards", GroupIndex: i, Cardinality: count, Classification: "structural_identity"}
			}
		}
	}
	return out
}

func replaceContentLeaf(content map[string]any, pointer string, value any) error {
	parts, err := contentPointerParts(pointer)
	if err != nil {
		return err
	}
	var node any = content
	for i, part := range parts {
		switch v := node.(type) {
		case map[string]any:
			if i == len(parts)-1 {
				v[part] = value
				return nil
			}
			node = v[part]
		case []any:
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 || n >= len(v) {
				return fmt.Errorf("invalid content array pointer %s", pointer)
			}
			if i == len(parts)-1 {
				v[n] = value
				return nil
			}
			node = v[n]
		default:
			return fmt.Errorf("invalid content pointer %s", pointer)
		}
	}
	return fmt.Errorf("empty content pointer")
}
