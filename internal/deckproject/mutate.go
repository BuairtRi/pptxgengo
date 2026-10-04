package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
	"time"
)

type Mutation struct {
	Operation        string   `json:"operation"`
	TemplateID       string   `json:"template_id"`
	SlideIDs         []string `json:"slide_ids"`
	BeforeSHA256     string   `json:"before_sha256"`
	AfterSHA256      string   `json:"after_sha256"`
	DefinitionSHA256 string   `json:"definition_sha256"`
	Decision         string   `json:"decision"`
}

func Fork(p *Project, parentID, newID, reason string, slides []string) (Mutation, error) {
	t, ok := p.Document.LocalTemplates[parentID]
	if !ok {
		return Mutation{}, fmt.Errorf("fork requires an existing local template")
	}
	if !stableID.MatchString(newID) {
		return Mutation{}, fmt.Errorf("invalid new template ID")
	}
	if _, ok := p.Document.LocalTemplates[newID]; ok {
		return Mutation{}, fmt.Errorf("new template already exists")
	}
	for _, id := range slides {
		found := false
		for _, slide := range p.Document.Slides {
			if slide.ID == id {
				found = true
				if slide.Template.Scope != "local" || slide.Template.ID != parentID {
					return Mutation{}, fmt.Errorf("fork target must reference parent local template: %s", id)
				}
				for key, value := range slide.Values {
					zone, ok := t.Zones[key]
					if !ok {
						return Mutation{}, fmt.Errorf("undeclared parent content zone %s", key)
					}
					if e := validateValue(zone.Schema, value, key); e != nil {
						return Mutation{}, e
					}
				}
				for key, zone := range t.Zones {
					if _, ok := slide.Values[key]; !ok && zone.Required {
						return Mutation{}, fmt.Errorf("missing parent content zone %s", key)
					}
				}
			}
		}
		if !found {
			return Mutation{}, fmt.Errorf("unknown fork target slide %s", id)
		}
	}
	hash := digest(canonical(t))
	snapshot, e := definitionSnapshot(p, canonical(t))
	if e != nil {
		return Mutation{}, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return Mutation{}, e
	}
	clone.Provenance = &Provenance{Operation: "fork", Parent: Reference{Scope: "local", ID: parentID, Revision: hash}, DefinitionSHA256: hash, DefinitionSnapshot: snapshot, Reason: reason}
	return applyTemplate(p, newID, clone, nil, slides, "fork", hash)
}
func Detach(p *Project, slideID, newID, bundle, engine, reason string) (Mutation, error) {
	if !stableID.MatchString(newID) {
		return Mutation{}, fmt.Errorf("invalid new template ID")
	}
	if _, ok := p.Document.LocalTemplates[newID]; ok {
		return Mutation{}, fmt.Errorf("new template already exists")
	}
	idx := -1
	for i, s := range p.Document.Slides {
		if s.ID == slideID {
			idx = i
		}
	}
	if idx < 0 {
		return Mutation{}, fmt.Errorf("unknown stable slide ID")
	}
	authored := p.Document.Slides[idx]
	if authored.Template.Scope != "shared" {
		return Mutation{}, fmt.Errorf("detach requires shared template; use fork for local templates")
	}
	c, e := Check(p, bundle, engine)
	if e != nil {
		return Mutation{}, e
	}
	native, report, e := wmdesign.BuildWithEngineAndAssets(bundle, "", c.Document, engine, c.Assets)
	_ = native
	if e != nil {
		return Mutation{}, e
	}
	compiled := c.Document.Slides[idx]
	if compiled.LibraryChrome != nil && (compiled.LibraryChrome.Stamp != "" || len(compiled.LibraryChrome.Notes) > 0) {
		return Mutation{}, fmt.Errorf("detach does not yet support stamped/source-note chrome; original shared slide remains unchanged")
	}
	for _, n := range compiled.Nodes {
		if n.Kind != "scene" || n.Scene == nil {
			return Mutation{}, fmt.Errorf("detach requires generic source-scene IR; legacy typed card-row/metric templates remain shared")
		}
	}
	catalog, e := wmdesign.LibraryCatalog(bundle, "")
	if e != nil {
		return Mutation{}, e
	}
	var def *wmdesign.LibraryTemplate
	for i := range catalog {
		if catalog[i].Key == authored.Template.ID {
			def = &catalog[i]
			break
		}
	}
	if def == nil {
		return Mutation{}, fmt.Errorf("unknown parent definition")
	}
	var rawDefinition any
	if e = json.Unmarshal(def.RawSlide, &rawDefinition); e != nil {
		return Mutation{}, e
	}
	definitionBytes := canonical(rawDefinition)
	hash := digest(definitionBytes)
	snapshot, e := definitionSnapshot(p, definitionBytes)
	if e != nil {
		return Mutation{}, e
	}
	source, e := wmdesign.Load(bundle, "")
	if e != nil {
		return Mutation{}, e
	}
	frame, e := source.ResolveFrame(compiled.Frame)
	if e != nil {
		return Mutation{}, e
	}
	frameOpts := compiled.Frame
	frameOpts.Nav = nil
	frameOpts.Active = ""
	t := LocalTemplate{Name: "Detached " + def.Name, Description: "Deck-local source-scene composition retaining authored content and frozen ancestry", Frame: Reference{Scope: "shared", ID: "wmds/frame/" + frame.Request.Rail + "-" + frame.Request.Footer}, FrameOptions: &frameOpts, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{}, Nodes: []Node{}, Provenance: &Provenance{Operation: "detach", Parent: Reference{Scope: "shared", ID: authored.Template.ID, Revision: strconv.Itoa(def.Revision)}, DefinitionSHA256: hash, DefinitionSnapshot: snapshot, SourceFile: def.SourceFile, SourceFileSHA256: def.SourceSHA256, Reason: reason}}
	if compiled.LibraryChrome != nil {
		ch := compiled.LibraryChrome
		t.FrameChrome = &FrameChrome{Emphasis: ch.Emphasis, Whiteboard: ch.Whiteboard, CustomWhiteboard: ch.CustomWhiteboard}
	}
	values := map[string]any{}
	addZone := func(id, role string, v any) {
		schema := map[string]any{"type": "string"}
		switch v.(type) {
		case float64:
			schema["type"] = "number"
		case bool:
			schema["type"] = "boolean"
		case map[string]any:
			schema["type"] = "object"
		case []any:
			schema["type"] = "array"
		}
		t.Zones[id] = Zone{Role: role, Required: true, Schema: schema}
		values[id] = v
	}
	if !compiled.Frame.NoHeader {
		addZone("title", "slide-title", compiled.Title)
		addZone("eyebrow", "eyebrow", compiled.Eyebrow)
	}
	if compiled.Source != "" {
		addZone("source", "source", compiled.Source)
	}
	if len(compiled.Frame.Nav) > 0 {
		items := []any{}
		for _, n := range compiled.Frame.Nav {
			items = append(items, map[string]any{"key": n.ID, "label": n.Label})
		}
		addZone("navigation", "nav", map[string]any{"items": items, "active": compiled.Frame.Active})
	}
	bounds := map[string]wmdesign.Rect{}
	for _, r := range report.Slides[idx].Scenes {
		bounds[r.ID] = r.Bounds
	}
	for _, n := range compiled.Nodes {
		var args map[string]any
		if e = json.Unmarshal(n.Scene.Node, &args); e != nil {
			return Mutation{}, e
		}
		kind, _ := args["type"].(string)
		b, ok := bounds[n.ID]
		if !ok || b.W <= 0 || b.H <= 0 {
			return Mutation{}, fmt.Errorf("detach requires positive measured scene bounds: %s", n.ID)
		}
		zone, origin, e := chooseZone(b, frame)
		if e != nil {
			return Mutation{}, fmt.Errorf("detach %s: %w", n.ID, e)
		}
		rect := wmdesign.Rect{X: b.X - origin.X, Y: b.Y - origin.Y, W: b.W, H: b.H}
		for _, slot := range def.Slots {
			if strings.HasPrefix(slot.SourcePointer, n.Scene.Path+"/") {
				relative := strings.TrimPrefix(slot.SourcePointer, n.Scene.Path)
				v, e := lookupPointer(args, relative)
				if e != nil {
					return Mutation{}, e
				}
				id := "value-" + digest([]byte(slot.Name))[:12]
				addZone(id, "content", v)
				z := t.Zones[id]
				z.Description = slot.Name
				t.Zones[id] = z
				if e = setPointer(args, relative, map[string]any{"binding": id}); e != nil {
					return Mutation{}, e
				}
			}
		}
		for _, field := range []string{"type", "id", "x", "y", "w", "h"} {
			delete(args, field)
		}
		keys := map[string][]string{}
		for path, ids := range n.Scene.Keys {
			if strings.HasPrefix(path, n.Scene.Path+"/") {
				keys[strings.TrimPrefix(path, n.Scene.Path+"/")] = ids
			}
		}
		t.Nodes = append(t.Nodes, Node{ID: n.ID, Kind: "component", Placement: &Placement{Zone: zone, Rect: &rect}, Definition: &Reference{Scope: "shared", ID: "wmds/component/" + kind}, Arguments: args, Keys: keys})
	}
	return applyTemplate(p, newID, t, values, []string{slideID}, "detach", hash)
}
func chooseZone(b wmdesign.Rect, f wmdesign.ResolvedFrame) (string, wmdesign.Rect, error) {
	zones := []struct {
		name string
		r    wmdesign.Rect
	}{{"short_body", f.ShortBody}, {"tall_body", f.TallBody}, {"body", f.Body}, {"rail", f.Rail}}
	for _, z := range zones {
		r := z.r
		if r.W > 0 && r.H > 0 && b.X >= r.X-.02 && b.Y >= r.Y-.02 && b.X+b.W <= r.X+r.W+.02 && b.Y+b.H <= r.Y+r.H+.02 {
			return z.name, r, nil
		}
	}
	return "", wmdesign.Rect{}, fmt.Errorf("source bounds extend outside supported local frame zones; no source mutation applied")
}
func setPointer(root any, path string, value any) error {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	parent := root
	for i, part := range parts {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		last := i == len(parts)-1
		switch cur := parent.(type) {
		case map[string]any:
			next, ok := cur[part]
			if !ok {
				return fmt.Errorf("missing pointer %s", path)
			}
			if last {
				cur[part] = value
				return nil
			}
			parent = next
		case []any:
			n, e := strconv.Atoi(part)
			if e != nil || n < 0 || n >= len(cur) {
				return fmt.Errorf("invalid pointer %s", path)
			}
			if last {
				cur[n] = value
				return nil
			}
			parent = cur[n]
		default:
			return fmt.Errorf("invalid pointer %s", path)
		}
	}
	return fmt.Errorf("invalid pointer %s", path)
}
func definitionSnapshot(p *Project, data []byte) (string, error) {
	path := "decisions/definitions/" + digest(data) + ".json"
	abs, e := SafePath(p.Root, path)
	if e != nil {
		return "", e
	}
	if prior, e := os.ReadFile(abs); e == nil {
		if !bytes.Equal(prior, data) {
			return "", fmt.Errorf("definition snapshot drift")
		}
		return path, nil
	}
	if e = writeExclusive(abs, data, 0444); e != nil {
		return "", e
	}
	return path, nil
}
func applyTemplate(p *Project, id string, t LocalTemplate, values map[string]any, slides []string, operation, hash string) (Mutation, error) {
	m := Mutation{Operation: operation, TemplateID: id, SlideIDs: slides, BeforeSHA256: digest(p.Raw), DefinitionSHA256: hash}
	selected := map[string]bool{}
	for _, s := range slides {
		if selected[s] {
			return m, fmt.Errorf("duplicate target slide")
		}
		selected[s] = true
	}
	var doc Document
	if e := strictInto(p.Document, &doc); e != nil {
		return m, e
	}
	if doc.LocalTemplates == nil {
		doc.LocalTemplates = map[string]LocalTemplate{}
	}
	doc.LocalTemplates[id] = t
	for i := range doc.Slides {
		if selected[doc.Slides[i].ID] {
			doc.Slides[i].Template = Reference{Scope: "local", ID: id}
			if values != nil {
				doc.Slides[i].Values = values
			}
			delete(selected, doc.Slides[i].ID)
		}
	}
	if len(selected) > 0 {
		return m, fmt.Errorf("unknown target slide IDs")
	}
	var tree any
	if e := json.Unmarshal(canonical(doc), &tree); e != nil {
		return m, e
	}
	yamlBytes, e := yaml.Marshal(tree)
	if e != nil {
		return m, e
	}
	temp, e := SafePath(p.Root, ".deck-mutation-"+nonce()+".yaml")
	if e != nil {
		return m, e
	}
	if e = writeExclusive(temp, yamlBytes, 0644); e != nil {
		return m, e
	}
	defer os.Remove(temp)
	candidate, e := Load(temp)
	if e != nil {
		return m, e
	}
	_ = candidate
	original, e := os.ReadFile(p.SourcePath)
	if e != nil {
		return m, e
	}
	if !bytes.Equal(original, p.Raw) {
		return m, fmt.Errorf("source changed during mutation")
	}
	beforePath, e := SafePath(p.Root, "decisions/sources/"+m.BeforeSHA256+".yaml")
	if e != nil {
		return m, e
	}
	if data, e := os.ReadFile(beforePath); e == nil {
		if !bytes.Equal(data, p.Raw) {
			return m, fmt.Errorf("source predecessor snapshot drift")
		}
	} else if os.IsNotExist(e) {
		if e = writeExclusive(beforePath, p.Raw, 0444); e != nil {
			return m, e
		}
	} else {
		return m, e
	}
	m.AfterSHA256 = digest(yamlBytes)
	m.Decision = "decisions/" + operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, e := SafePath(p.Root, m.Decision)
	if e != nil {
		return m, e
	}
	if e = writeJSON(decision, m); e != nil {
		return m, e
	}
	if e = os.Rename(temp, p.SourcePath); e != nil {
		return m, e
	}
	return m, nil
}
