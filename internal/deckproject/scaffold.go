package deckproject

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type TemplateScaffold struct {
	Schema                string                       `json:"schema"`
	Template              LocalTemplate                `json:"template"`
	SyntheticSourceValues map[string]any               `json:"synthetic_source_values"`
	Policy                []string                     `json:"policy"`
	OmittedNodes          []string                     `json:"omitted_nodes,omitempty"`
	Assets                map[string]Asset             `json:"assets,omitempty"`
	AssetPayloads         map[string][]byte            `json:"asset_payloads,omitempty"`
	PlaceholderMedia      []ScaffoldPlaceholder        `json:"placeholder_media,omitempty"`
	GeometryAdjustments   []ScaffoldGeometryAdjustment `json:"geometry_adjustments,omitempty"`
}

type ScaffoldGeometryAdjustment struct {
	NodeID   string        `json:"node_id"`
	Original wmdesign.Rect `json:"original_measured_allocation"`
	Authored wmdesign.Rect `json:"authored_allocation"`
	Reason   string        `json:"reason"`
}

// ScaffoldTemplate derives editable scene components from the actual catalog
// definition before attempting to fit a particular slide's content. It never
// installs synthetic values into a project or changes the shared definition.
type ScaffoldOptions struct {
	PlaceholderMedia            bool
	OmitNodes                   []string
	SourceContainerClearanceFit bool
}

func ScaffoldTemplate(bundle, key, engine, reason string, year int, omitNodes ...string) (TemplateScaffold, error) {
	return ScaffoldTemplateWithOptions(bundle, key, engine, reason, year, ScaffoldOptions{OmitNodes: omitNodes})
}

func ScaffoldTemplateWithOptions(bundle, key, engine, reason string, year int, options ScaffoldOptions) (TemplateScaffold, error) {
	out := TemplateScaffold{Schema: "pptxgengo.local-template-scaffold.v1", SyntheticSourceValues: map[string]any{}, Policy: []string{"Synthetic source values illustrate the original contract; replace them with supplied content before use.", "The shared parent is pinned; topology or geometry edits require a reason and supplied-content native review.", "This scaffold is not applied to a project automatically."}}
	if strings.TrimSpace(reason) == "" {
		return out, fmt.Errorf("template scaffold requires an explicit adaptation reason")
	}
	catalog, err := wmdesign.LibraryCatalog(bundle, "")
	if err != nil {
		return out, err
	}
	var def *wmdesign.LibraryTemplate
	for i := range catalog {
		if catalog[i].Key == key {
			def = &catalog[i]
			break
		}
	}
	if def == nil {
		return out, fmt.Errorf("unknown shared template %s", key)
	}
	if def.ContentContract == wmdesign.TemplateBindingsContract {
		return out, fmt.Errorf("template scaffold requires a named source-scene contract; typed template %s should remain shared or be deliberately composed", key)
	}
	examples, err := wmdesign.LibraryReference(bundle, "", def.Family, year)
	if err != nil {
		return out, err
	}
	for _, example := range examples.Slides {
		if example.Template == key {
			examples.Slides = []wmdesign.BoundSlide{example}
			break
		}
	}
	doc, _, err := wmdesign.BindTemplates(bundle, "", examples)
	if err != nil {
		return out, err
	}
	var compiled *wmdesign.SlideSpec
	for i := range doc.Slides {
		if doc.Slides[i].TemplateBinding.Template == key {
			compiled = &doc.Slides[i]
			break
		}
	}
	if compiled == nil {
		return out, fmt.Errorf("source specimen unavailable for %s", key)
	}
	omitted := map[string]bool{}
	for _, id := range options.OmitNodes {
		found := false
		for _, node := range compiled.Nodes {
			if node.ID == id {
				found = true
				break
			}
		}
		if !found || omitted[id] {
			return out, fmt.Errorf("unknown or duplicate omitted source node %s", id)
		}
		omitted[id] = true
	}
	if ch := compiled.LibraryChrome; ch != nil && (ch.Stamp != "" || len(ch.Notes) > 0 || len(ch.Tint) > 0) {
		return out, fmt.Errorf("template scaffold does not yet support source-note, stamp or tint chrome; shared definition remains unchanged")
	}
	doc.Slides = []wmdesign.SlideSpec{*compiled}
	var placeholders map[string]wmdesign.AssetData
	if options.PlaceholderMedia {
		placeholders, err = wmdesign.TemplatePlaceholderAssets()
		if err != nil {
			return out, err
		}
		for key := range placeholders {
			if strings.HasPrefix(key, "icon/") || strings.HasPrefix(key, "arrow-") {
				delete(placeholders, key)
			}
		}
		out.Policy = append(out.Policy, "Explicit schematic placeholders replace only non-icon/non-arrow registered media; they are not original artwork and must be reviewed/replaced before delivery.")
	}
	payload, report, err := wmdesign.BuildWithEngineAndAssets(bundle, "", doc, engine, placeholders)
	if err != nil {
		return out, err
	}
	if options.PlaceholderMedia {
		if err = addScaffoldPlaceholderPayloads(&out, payload, placeholders, omitted); err != nil {
			return out, err
		}
	}
	frame := report.Slides[0].Frame
	frameOptions := compiled.Frame
	frameOptions.Nav = nil
	frameOptions.Active = ""
	var raw any
	if err = json.Unmarshal(def.RawSlide, &raw); err != nil {
		return out, err
	}
	out.Template = LocalTemplate{Name: "Adapted " + def.Name, Description: "Editable composition derived from the actual shared source scene", Frame: Reference{Scope: "shared", ID: "wmds/frame/" + frame.Request.Rail + "-" + frame.Request.Footer}, FrameOptions: &frameOptions, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{}, Nodes: []Node{}, Provenance: &Provenance{Operation: "scaffold", Parent: Reference{Scope: "shared", ID: key, Revision: strconv.Itoa(def.Revision)}, DefinitionSHA256: digest(canonical(raw)), SourceFile: def.SourceFile, SourceFileSHA256: def.SourceSHA256, Reason: reason}}
	if ch := compiled.LibraryChrome; ch != nil {
		out.Template.FrameChrome = &FrameChrome{Emphasis: ch.Emphasis, Whiteboard: ch.Whiteboard, CustomWhiteboard: ch.CustomWhiteboard}
	}
	add := func(name, role string, value any) {
		kind := "string"
		switch value.(type) {
		case float64:
			kind = "number"
		case bool:
			kind = "boolean"
		case []any:
			kind = "array"
		case map[string]any:
			kind = "object"
		case nil:
			kind = "null"
		}
		out.Template.Zones[name] = Zone{Role: role, Required: true, Schema: map[string]any{"type": kind}, Description: name}
		out.SyntheticSourceValues[name] = value
	}
	if !compiled.Frame.NoHeader {
		add("title", "slide-title", compiled.Title)
		add("eyebrow", "eyebrow", compiled.Eyebrow)
	}
	if compiled.Source != "" {
		add("source", "source", compiled.Source)
	}
	if len(compiled.Frame.Nav) > 0 {
		items := []any{}
		for _, tab := range compiled.Frame.Nav {
			items = append(items, map[string]any{"key": tab.ID, "label": tab.Label})
		}
		add("navigation", "nav", map[string]any{"items": items, "active": compiled.Frame.Active})
	}
	bounds := map[string]wmdesign.Rect{}
	for _, scene := range report.Slides[0].Scenes {
		bounds[scene.ID] = scene.Bounds
	}
	for _, node := range compiled.Nodes {
		if omitted[node.ID] {
			out.OmittedNodes = append(out.OmittedNodes, node.ID)
			continue
		}
		if node.Kind != "scene" || node.Scene == nil {
			return out, fmt.Errorf("template scaffold requires generic source-scene nodes: %s", node.ID)
		}
		var args map[string]any
		if err = json.Unmarshal(node.Scene.Node, &args); err != nil {
			return out, err
		}
		kind, _ := args["type"].(string)
		b, ok := bounds[node.ID]
		if !ok {
			return out, fmt.Errorf("measured scene bounds missing: %s", node.ID)
		}
		b, errBounds := editableSceneAllocation(kind, args, b, frame)
		if errBounds != nil {
			return out, fmt.Errorf("%s: %w", node.ID, errBounds)
		}
		if options.SourceContainerClearanceFit && (kind == "container" || kind == "frame") && b.Y+b.H > frame.Body.Y+frame.Body.H {
			limit := sourceContainerZone(frame)
			if b.X >= limit.X && b.X+b.W <= limit.X+limit.W && b.Y < limit.H && b.Y+b.H <= frame.Body.Y+frame.Body.H+6 {
				adjusted := b
				adjusted.H = limit.H - b.Y
				if adjusted.H < b.H {
					out.GeometryAdjustments = append(out.GeometryAdjustments, ScaffoldGeometryAdjustment{NodeID: node.ID, Original: b, Authored: adjusted, Reason: "Explicit source-container clearance fit: shrink original height to retain six-point caption/footer clearance; not unchanged-source visual fidelity"})
					b = adjusted
				}
			}
		}
		zone, origin, err := chooseSourceSceneZone(kind, args, b, frame)
		if err != nil {
			return out, fmt.Errorf("template scaffold %s bounds %+v original geometry %s: %w", node.ID, b, node.Scene.Node, err)
		}
		for _, slot := range def.Slots {
			if !strings.HasPrefix(slot.SourcePointer, node.Scene.Path+"/") {
				continue
			}
			pointer := strings.TrimPrefix(slot.SourcePointer, node.Scene.Path)
			value, err := lookupPointer(args, pointer)
			if err != nil {
				return out, err
			}
			add(slot.Name, "content", value)
			if err = setPointer(args, pointer, map[string]any{"binding": slot.Name}); err != nil {
				return out, err
			}
		}
		for _, field := range []string{"type", "id", "x", "y", "w", "h"} {
			delete(args, field)
		}
		keys := map[string][]string{}
		for path, ids := range node.Scene.Keys {
			if strings.HasPrefix(path, node.Scene.Path+"/") {
				keys[strings.TrimPrefix(path, node.Scene.Path+"/")] = ids
			}
		}
		r := wmdesign.Rect{X: b.X - origin.X, Y: b.Y - origin.Y, W: b.W, H: b.H}
		out.Template.Nodes = append(out.Template.Nodes, Node{ID: node.ID, Kind: "component", Placement: &Placement{Zone: zone, Rect: &r}, Definition: &Reference{Scope: "shared", ID: "wmds/component/" + kind}, Arguments: args, Keys: keys})
	}
	return out, nil
}
