package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var stableID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,119}$`)
var shaPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func escape(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1") }
func (p *Project) fail(path, format string, args ...any) error {
	location := path
	pos, found := p.Positions[location]
	for !found && location != "" {
		if slash := strings.LastIndex(location, "/"); slash >= 0 {
			location = location[:slash]
		} else {
			location = ""
		}
		pos, found = p.Positions[location]
	}
	file := p.SourcePath
	if source := p.positionFiles[location]; source != "" {
		file = filepath.Join(p.Root, filepath.FromSlash(source))
	}
	return fmt.Errorf("%s:%d:%d (%s): %s", file, pos.Line, pos.Column, path, fmt.Sprintf(format, args...))
}
func Load(path string) (*Project, error) {
	return loadProject(path, nil)
}

func loadProject(path string, overrides map[string][]byte) (*Project, error) {
	a, e := filepath.Abs(path)
	if e != nil {
		return nil, e
	}
	st, e := os.Stat(a)
	if e != nil {
		return nil, e
	}
	if st.IsDir() {
		a = filepath.Join(a, "deck.yaml")
	}
	safe, e := SafePath(filepath.Dir(a), filepath.Base(a))
	if e != nil {
		return nil, e
	}
	a = safe
	p := &Project{Root: filepath.Dir(a), SourcePath: a, Positions: map[string]Position{}, SourceFiles: map[string][]byte{}, SlideFiles: map[string]string{}, TemplateFiles: map[string]string{}, NotesFiles: map[string]string{}, positionFiles: map[string]string{}, sourceOverrides: overrides}
	raw, e := p.readSource(filepath.Base(a), 16<<20)
	if e != nil {
		return nil, e
	}
	p.Raw = raw
	v, e := p.parseSource(raw, filepath.Base(a), "")
	if e != nil {
		return nil, e
	}
	tree, ok := v.(map[string]any)
	if !ok {
		return nil, p.fail("", "root must be a mapping")
	}
	if e = p.expandSourceReferences(tree); e != nil {
		return nil, e
	}
	p.tree = tree
	p.Canonical = canonical(tree)
	if e = p.shapeType(tree, reflect.TypeOf(Document{}), ""); e != nil {
		return nil, e
	}
	dec := json.NewDecoder(bytes.NewReader(p.Canonical))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&p.Document); e != nil {
		path := ""
		if strings.HasPrefix(e.Error(), "json: unknown field ") {
			key := strings.TrimPrefix(e.Error(), "json: unknown field ")
			key = strings.Trim(key, "\"")
			for k := range p.Positions {
				if strings.HasSuffix(k, "/"+escape(key)) {
					path = k
					break
				}
			}
		}
		return nil, p.fail(path, "%v", e)
	}
	if e = p.validate(); e != nil {
		return nil, e
	}
	if e = p.strictShape(); e != nil {
		return nil, e
	}
	return p, nil
}
func (p *Project) yamlValue(n *yaml.Node, path string, depth int) (any, error) {
	p.Positions[path] = Position{n.Line, n.Column}
	if p.activeSource != "" {
		p.positionFiles[path] = p.activeSource
	}
	if depth > 100 {
		return nil, p.fail(path, "nesting exceeds 100")
	}
	switch n.Kind {
	case yaml.AliasNode:
		return nil, p.fail(path, "YAML aliases are unsupported; use explicit values")
	case yaml.MappingNode:
		m := map[string]any{}
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Kind != yaml.ScalarNode || k.Tag != "!!str" || k.Value == "<<" {
				return nil, p.fail(path, "mapping keys must be literal strings; merges unsupported")
			}
			child := path + "/" + escape(k.Value)
			if _, ok := m[k.Value]; ok {
				return nil, p.fail(path, "duplicate field %q at line %d", k.Value, k.Line)
			}
			v, e := p.yamlValue(n.Content[i+1], child, depth+1)
			if e != nil {
				return nil, e
			}
			m[k.Value] = v
		}
		return m, nil
	case yaml.SequenceNode:
		a := []any{}
		for i, c := range n.Content {
			v, e := p.yamlValue(c, path+"/"+strconv.Itoa(i), depth+1)
			if e != nil {
				return nil, e
			}
			a = append(a, v)
		}
		return a, nil
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!str":
			return n.Value, nil
		case "!!null":
			return nil, nil
		case "!!bool":
			var b bool
			if e := n.Decode(&b); e != nil {
				return nil, e
			}
			return b, nil
		case "!!int", "!!float":
			var f float64
			if e := n.Decode(&f); e != nil {
				return nil, e
			}
			if math.IsInf(f, 0) || math.IsNaN(f) {
				return nil, p.fail(path, "nonfinite numbers unsupported")
			}
			if n.Tag == "!!int" && math.Abs(f) > 9007199254740991 {
				return nil, p.fail(path, "integer exceeds exact JSON range")
			}
			return f, nil
		default:
			return nil, p.fail(path, "unsupported YAML tag %s; quote date-like values", n.Tag)
		}
	}
	return nil, p.fail(path, "unsupported YAML node")
}

// SafePath rejects lexical escapes and existing symlinks anywhere in the path.
// File creation callers must also ensure their private destination is unchanged.
func SafePath(root, rel string) (string, error) {
	if rel == "" || filepath.IsAbs(rel) || strings.Contains(rel, "\\") || strings.Contains(rel, ":") || strings.ContainsRune(rel, 0) {
		return "", fmt.Errorf("unsafe relative path %q", rel)
	}
	for _, part := range strings.Split(rel, "/") {
		if part == ".." || part == "" {
			return "", fmt.Errorf("unsafe relative path %q", rel)
		}
	}
	base, e := filepath.Abs(root)
	if e != nil {
		return "", e
	}
	resolved, e := filepath.EvalSymlinks(base)
	if e != nil {
		return "", e
	}
	cur := resolved
	for _, part := range strings.Split(filepath.Clean(rel), string(filepath.Separator)) {
		cur = filepath.Join(cur, part)
		st, e := os.Lstat(cur)
		if e == nil && st.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink path forbidden: %s", rel)
		}
		if e != nil && !os.IsNotExist(e) {
			return "", e
		}
	}
	return cur, nil
}
func (p *Project) validate() error {
	d := p.Document
	if err := wmdesign.ValidateEditingProfile(d.EditingProfile); err != nil {
		return p.fail("/editing_profile", "%v", err)
	}
	if d.Schema != Schema || !stableID.MatchString(d.ID) || strings.TrimSpace(d.Title) == "" || d.Year < 2000 || d.Year > 9999 || len(d.Slides) == 0 {
		return p.fail("", "schema/id/title/year/slides invalid or missing")
	}
	if _, e := SafePath(p.Root, d.Toolchain.Lockfile); e != nil {
		return p.fail("/toolchain/lockfile", "%v", e)
	}
	allowed := map[string]bool{"project": true, "audience": true, "outline": true, "sources": true, "claims": true, "composition_log": true, "decisions": true, "win_strategy": true, "state": true}
	for k, path := range d.Context {
		if filepath.Clean(path) == "state.json" && k != "state" {
			return p.fail("/context/"+escape(k), "generated state.json is not an authored dependency; use context.state")
		}
		if k == "state" && filepath.Clean(path) != "state.json" {
			return p.fail("/context/state", "generated project state path must be state.json")
		}
		if !allowed[k] {
			return p.fail("/context/"+escape(k), "unknown context field")
		}
		if _, e := SafePath(p.Root, path); e != nil {
			return p.fail("/context/"+escape(k), "%v", e)
		}
	}
	if err := validateScaffoldPlaceholderAssets(d.Assets); err != nil {
		return p.fail("/assets", "%v", err)
	}
	for id, a := range d.Assets {
		path := "/assets/" + escape(id)
		if a.Focus != nil {
			assets, _ := p.tree["assets"].(map[string]any)
			asset, _ := assets[id].(map[string]any)
			focus, _ := asset["focus"].(map[string]any)
			if _, ok := focus["x"]; !ok {
				return p.fail(path+"/focus", "focus requires explicit x and y")
			}
			if _, ok := focus["y"]; !ok {
				return p.fail(path+"/focus", "focus requires explicit x and y")
			}
		}
		if a.Focus != nil && (a.Focus.X < 0 || a.Focus.X > 1 || a.Focus.Y < 0 || a.Focus.Y > 1) {
			return p.fail(path+"/focus", "focus x/y must be in [0,1]")
		}
		if !stableID.MatchString(id) || (a.Path == "") == (a.RegistryID == "") {
			return p.fail(path, "asset requires valid ID and exactly one path or registry_id")
		}
		if a.SHA256 != "" && !shaPattern.MatchString(a.SHA256) {
			return p.fail(path, "invalid sha256")
		}
		if a.Path != "" {
			if _, e := SafePath(p.Root, a.Path); e != nil {
				return p.fail(path, "%v", e)
			}
		}
		if (a.DerivedFrom == "") != (a.DerivationReceipt == "") {
			return p.fail(path, "derivatives require derived_from and derivation_receipt")
		}
		if a.DerivedFrom != "" {
			if _, ok := d.Assets[a.DerivedFrom]; !ok || a.DerivedFrom == id {
				return p.fail(path, "unknown/self derived_from")
			}
			if _, e := SafePath(p.Root, a.DerivationReceipt); e != nil {
				return p.fail(path, "%v", e)
			}
		}
	}
	visiting := map[string]int{}
	var visitAsset func(string) error
	visitAsset = func(id string) error {
		if visiting[id] == 1 {
			return p.fail("/assets/"+escape(id), "asset derivative ancestry cycle")
		}
		if visiting[id] == 2 {
			return nil
		}
		visiting[id] = 1
		if parent := d.Assets[id].DerivedFrom; parent != "" {
			if e := visitAsset(parent); e != nil {
				return e
			}
		}
		visiting[id] = 2
		return nil
	}
	for id := range d.Assets {
		if e := visitAsset(id); e != nil {
			return e
		}
	}
	ids := map[string]bool{}
	for i, s := range d.Slides {
		path := fmt.Sprintf("/slides/%d", i)
		if !stableID.MatchString(s.ID) || ids[s.ID] {
			return p.fail(path+"/id", "invalid or duplicate stable slide ID")
		}
		ids[s.ID] = true
		if s.ContentKind != "supplied_content" && s.ContentKind != "synthetic_example" {
			return p.fail(path+"/content_kind", "explicit content kind required")
		}
		if s.Density != "" && !validTypographyDensity(s.Density) {
			return p.fail(path+"/density", "must be comfortable, compact, or dense")
		}
		if s.HeaderDensity != "" && !validTypographyDensity(s.HeaderDensity) {
			return p.fail(path+"/header_density", "must be comfortable, compact, or dense")
		}
		if err := wmdesign.ValidateSpeakerNotes(s.Notes); err != nil {
			return p.fail(path+"/notes", "notes must be valid XML text, at most 1 MiB")
		}
		if err := validateDraftReview(s.DraftReview); err != nil {
			return p.fail(path+"/draft_review", "%v", err)
		}
		if s.Values == nil {
			return p.fail(path+"/values", "values mapping required")
		}
		if s.Template.Scope != "shared" && s.Template.Scope != "local" || s.Template.ID == "" {
			return p.fail(path+"/template", "invalid reference")
		}
		if s.Template.Scope == "local" {
			if _, ok := d.LocalTemplates[s.Template.ID]; !ok {
				return p.fail(path+"/template", "unknown local template")
			}
		}
		if s.Brief != "" {
			if strings.ContainsAny(s.Brief, "\r\n\t") {
				return p.fail(path+"/brief", "brief must be a single-line project-relative path to a page brief file, not inline editorial text")
			}
			if _, e := SafePath(p.Root, s.Brief); e != nil {
				return p.fail(path+"/brief", "brief must be a project-relative path to a page brief file, not inline editorial text: %v", e)
			}
		}
		if len(s.NativeGeometry) > 0 || len(s.NativeOrder) > 0 || len(s.DiagramContainment) > 0 {
			if s.NativeGeometryTemplate == nil || *s.NativeGeometryTemplate != s.Template {
				return p.fail(path, "native geometry/containment is pinned to another template; reset/review native layout and containment before a template upgrade")
			}
		}
		if len(s.NativeGeometry) > 10000 || len(s.NativeOrder) > 10000 {
			return p.fail(path, "native geometry/order exceeds 10000 entries")
		}
		if e := validateContainmentDeclarations(s.DiagramContainment); e != nil {
			return p.fail(path+"/diagram_containment", "%v", e)
		}
		for name, g := range s.NativeGeometry {
			if !shaPattern.MatchString(g.SourceGeometrySHA256) {
				return p.fail(path+"/native_geometry/"+escape(name), "native transform requires an authored source geometry pin; propose/adopt from a compatible baseline")
			}
			if name == "" || len(name) > 512 {
				return p.fail(path, "invalid native geometry key")
			}
			if e := validateNativeGeometry(g); e != nil {
				return p.fail(path+"/native_geometry/"+escape(name), "%v", e)
			}
		}
		for _, names := range s.NativeOrder {
			if len(names) > 10000 {
				return p.fail(path, "native order exceeds 10000 objects")
			}
			seen := map[string]bool{}
			for _, name := range names {
				if name == "" || len(name) > 512 || seen[name] {
					return p.fail(path, "invalid or duplicate native order name")
				}
				seen[name] = true
			}
		}
		seen := map[string]bool{}
		for _, r := range s.EvidenceRefs {
			if !stableID.MatchString(r) || seen[r] {
				return p.fail(path+"/evidence_refs", "invalid/duplicate evidence ID")
			}
			seen[r] = true
		}
	}
	slideIDs := make([]string, len(d.Slides))
	for i, slide := range d.Slides {
		slideIDs[i] = slide.ID
	}
	if err := wmdesign.ValidateSections(d.Sections, slideIDs); err != nil {
		if failure, ok := err.(*wmdesign.SectionValidationError); ok {
			return p.fail(fmt.Sprintf("/sections/%d/%s", failure.Index, failure.Field), "%s", failure.Reason)
		}
		return p.fail("/sections", "%v", err)
	}
	for id, t := range d.LocalTemplates {
		path := "/local_templates/" + escape(id)
		if !stableID.MatchString(id) || strings.TrimSpace(t.Name) == "" || len(t.Nodes) == 0 || t.Zones == nil {
			return p.fail(path, "local template missing name/zones/nodes")
		}
		if t.Frame.Scope != "shared" || t.Grid.Scope != "shared" {
			return p.fail(path, "frame and grid must reference shared definitions")
		}
		if t.Provenance != nil {
			v := t.Provenance
			if (v.Parent.Scope != "shared" && v.Parent.Scope != "local") || v.Parent.ID == "" {
				return p.fail(path+"/provenance/parent", "invalid parent reference")
			}
			if v.DefinitionSnapshot != "" {
				if _, e := SafePath(p.Root, v.DefinitionSnapshot); e != nil {
					return p.fail(path+"/provenance", "%v", e)
				}
			}
			if (v.SourceFile == "") != (v.SourceFileSHA256 == "") {
				return p.fail(path+"/provenance", "source_file and source_file_sha256 must occur together")
			}
			if v.SourceFileSHA256 != "" && !shaPattern.MatchString(v.SourceFileSHA256) {
				return p.fail(path+"/provenance", "invalid source file hash")
			}
			if (v.Operation != "fork" && v.Operation != "detach" && v.Operation != "scaffold") || !shaPattern.MatchString(v.DefinitionSHA256) {
				return p.fail(path+"/provenance", "invalid provenance")
			}
		}
		reservedRoles := map[string]bool{}
		for k, z := range t.Zones {
			if z.Role == "slide-title" || z.Role == "eyebrow" || z.Role == "source" || z.Role == "nav" {
				if reservedRoles[z.Role] {
					return p.fail(path+"/zones/"+escape(k), "duplicate reserved frame role %s", z.Role)
				}
				reservedRoles[z.Role] = true
			}
			if !stableID.MatchString(k) || z.Role == "" || len(z.Schema) == 0 {
				return p.fail(path+"/zones/"+escape(k), "zone requires ID, role and schema")
			}
			if e := validateSchema(z.Schema); e != nil {
				return p.fail(path+"/zones/"+escape(k), "%v", e)
			}
		}
		seen := map[string]bool{}
		var walk func([]Node, string) error
		walk = func(ns []Node, ptr string) error {
			for i, n := range ns {
				np := fmt.Sprintf("%s/%d", ptr, i)
				if !stableID.MatchString(n.ID) || seen[n.ID] {
					return p.fail(np+"/id", "invalid/duplicate node ID")
				}
				seen[n.ID] = true
				if n.Kind == "group" {
					if len(n.Nodes) == 0 || n.Placement != nil {
						return p.fail(np, "group requires nodes and has no placement")
					}
					if e := walk(n.Nodes, np+"/nodes"); e != nil {
						return e
					}
					continue
				}
				if n.Placement == nil || (n.Placement.Rect == nil) == (n.Placement.Span == nil) {
					return p.fail(np, "node requires exactly one rect or span placement")
				}
				if n.Placement.Zone == "source_container" {
					if e := validateSourceContainerLocal(t, n); e != nil {
						return p.fail(np, "%v", e)
					}
				}
				if n.Placement.Zone == "source_canvas" && ((n.Kind != "component" && n.Kind != "composite") || t.Provenance == nil || t.Provenance.SourceFile == "" || t.Provenance.Parent.Scope != "shared") {
					return p.fail(np, "source_canvas requires a pinned source-derived scene component; review intentional margin/header interaction")
				}
				if n.Placement.Zone == "source_body" && (n.Kind != "component" && n.Kind != "composite") {
					return p.fail(np, "source_body requires an explicit scene component; review intentional header interaction")
				}
				if n.Placement.Zone == "tall_plot" && (n.Kind != "component" && n.Kind != "composite" || n.Definition == nil || n.Definition.ID != "wmds/component/chart") {
					return p.fail(np, "tall_plot is reserved for typed quadrant charts")
				}
				switch n.Kind {
				case "text":
					if n.Style == "" || n.Text == nil {
						return p.fail(np, "text requires style and text")
					}
				case "box":
					if n.Surface == "" {
						return p.fail(np, "box requires surface")
					}
				case "image":
					if n.Asset == nil || (n.Fit != "contain" && n.Fit != "cover") || n.Rotation < 0 || n.Rotation >= 360 {
						return p.fail(np, "invalid image fields")
					}
				case "rule":
					if n.Ink == "" || n.Weight <= 0 {
						return p.fail(np, "rule requires ink/positive weight")
					}
				case "component", "composite":
					if n.Definition == nil || n.Definition.Scope != "shared" || (n.Arguments == nil && n.Definition.ID != "wmds/component/rule") {
						return p.fail(np, "typed definition and arguments required")
					}
				default:
					return p.fail(np, "unsupported node kind %q", n.Kind)
				}
			}
			return nil
		}
		if e := walk(t.Nodes, path+"/nodes"); e != nil {
			return e
		}
		// Semantic roles remain extensible, but every non-frame zone must be
		// consumed by a node. Otherwise valid-looking authored copy is dropped.
		boundZones := map[string]bool{}
		var collectNodeBindings func([]Node)
		collectNodeBindings = func(nodes []Node) {
			for _, n := range nodes {
				if n.Kind == "group" {
					collectNodeBindings(n.Nodes)
					continue
				}
				for _, value := range []any{n.Text, n.Asset, n.Arguments} {
					collectBindings(value, func(key string) { boundZones[key] = true })
				}
			}
		}
		collectNodeBindings(t.Nodes)
		zoneIDs := make([]string, 0, len(t.Zones))
		for k := range t.Zones {
			zoneIDs = append(zoneIDs, k)
		}
		sort.Strings(zoneIDs)
		for _, k := range zoneIDs {
			z := t.Zones[k]
			if reservedRoles[z.Role] || boundZones[k] {
				continue
			}
			if z.Role == "title" {
				return p.fail(path+"/zones/"+escape(k)+"/role", "unsupported implicit frame role %q: use slide-title for the frame headline, or bind this zone explicitly to a node", z.Role)
			}
			return p.fail(path+"/zones/"+escape(k)+"/role", "zone role %q has no renderer binding: use slide-title, eyebrow, source or nav for a frame field, or bind this zone explicitly to a node", z.Role)
		}
	}
	return nil
}

func validTypographyDensity(value string) bool {
	switch value {
	case "comfortable", "compact", "dense":
		return true
	default:
		return false
	}
}
