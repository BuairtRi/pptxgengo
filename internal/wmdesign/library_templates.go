package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const LibraryBindingsContract = "pptxgengo.wmds-library-bindings.v1"

type LibraryWhiteboard struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
	Fade  string  `json:"fade,omitempty"`
	On    string  `json:"on,omitempty"`
	Above bool    `json:"above,omitempty"`
}
type LibraryChrome struct {
	Emphasis         string              `json:"emphasis,omitempty"`
	Stamp            string              `json:"stamp,omitempty"`
	Whiteboard       []LibraryWhiteboard `json:"whiteboard,omitempty"`
	CustomWhiteboard bool                `json:"custom_whiteboard,omitempty"`
	Notes            []string            `json:"notes,omitempty"`
	Tint             []LibraryTint       `json:"tint,omitempty"`
}
type LibraryTint struct {
	X       float64 `json:"x"`
	W       float64 `json:"w"`
	Surface string  `json:"surface,omitempty"`
}
type librarySlide struct {
	Type        string      `json:"type"`
	Rail        string      `json:"rail"`
	Footer      string      `json:"footer"`
	Surface     string      `json:"surface,omitempty"`
	RailSurface string      `json:"railSurface,omitempty"`
	Eyebrow     string      `json:"eyebrow,omitempty"`
	Title       string      `json:"title,omitempty"`
	TitleLines  int         `json:"titleLines,omitempty"`
	Density     string      `json:"density,omitempty"`
	Page        string      `json:"page,omitempty"`
	NoPage      bool        `json:"noPage,omitempty"`
	Emphasis    string      `json:"emphasis,omitempty"`
	Stamp       string      `json:"stamp,omitempty"`
	Split       string      `json:"split,omitempty"`
	Nav         *libraryNav `json:"nav,omitempty"`
	Source      *struct {
		Text  string   `json:"text,omitempty"`
		Notes []string `json:"notes,omitempty"`
	} `json:"source,omitempty"`
	Whiteboard json.RawMessage   `json:"whiteboard,omitempty"`
	Tint       json.RawMessage   `json:"tint,omitempty"`
	Body       []json.RawMessage `json:"body"`
}
type libraryNav struct {
	Items  []string `json:"items"`
	Active *int     `json:"active"`
}
type libraryEntry struct {
	ID         string          `json:"id"`
	Variant    string          `json:"variant"`
	Name       string          `json:"name"`
	Tier       string          `json:"tier"`
	Purpose    string          `json:"purpose"`
	Uses       []string        `json:"uses"`
	Legacy     string          `json:"legacy"`
	Budget     json.RawMessage `json:"budget"`
	Slots      json.RawMessage `json:"slots"`
	Slide      json.RawMessage `json:"slide"`
	Revision   int             `json:"rev,omitempty"`
	Added      string          `json:"added,omitempty"`
	Revised    string          `json:"revised,omitempty"`
	Status     string          `json:"status,omitempty"`
	ReplacedBy string          `json:"replacedBy,omitempty"`
}
type LibrarySlot struct {
	Name          string          `json:"name"`
	SourcePointer string          `json:"source_pointer"`
	Kind          string          `json:"kind"`
	AllowEmpty    bool            `json:"allow_empty,omitempty"`
	Example       json.RawMessage `json:"synthetic_source_example"`
}
type LibraryArray struct {
	Name          string `json:"name"`
	SourcePointer string `json:"source_pointer"`
	Count         int    `json:"count"`
}
type LibraryValueSchema struct {
	Kind        string            `json:"kind"`
	GoType      string            `json:"go_type"`
	Reference   string            `json:"reference"`
	Fields      map[string]string `json:"fields"`
	ExactCounts map[string]int    `json:"exact_counts,omitempty"`
}

type LibraryTemplate struct {
	TemplateDefinition
	Family              string              `json:"family"`
	Tier                string              `json:"tier"`
	Name                string              `json:"name"`
	Purpose             string              `json:"purpose,omitempty"`
	Uses                []string            `json:"uses,omitempty"`
	AdvisoryBudget      json.RawMessage     `json:"source_advisory_budget,omitempty"`
	AdvisorySlots       json.RawMessage     `json:"source_advisory_slots,omitempty"`
	Discovery           LibraryDiscovery    `json:"discovery"`
	Authoring           *LibraryAuthoring   `json:"authoring,omitempty"`
	ContentContract     string              `json:"content_contract"`
	ValueSchema         *LibraryValueSchema `json:"value_schema,omitempty"`
	Slots               []LibrarySlot       `json:"slots,omitempty"`
	Arrays              []LibraryArray      `json:"arrays,omitempty"`
	RenderStatus        string              `json:"render_status"`
	Policy              []string            `json:"policy"`
	RawSlide            json.RawMessage     `json:"-"`
	Revision            int                 `json:"revision"`
	Added               string              `json:"added,omitempty"`
	Revised             string              `json:"revised,omitempty"`
	Status              string              `json:"status"`
	ReplacedBy          string              `json:"replaced_by,omitempty"`
	Nav                 *LibraryNavContract `json:"nav,omitempty"`
	PendingCapabilities []string            `json:"pending_capabilities,omitempty"`
}
type LibraryNavContract struct {
	MinItems int               `json:"min_items"`
	MaxItems int               `json:"max_items"`
	Example  LibraryNavContent `json:"synthetic_source_example"`
}

func legacyTemplate(key string) bool {
	for _, k := range executableTemplateKeys {
		if key == k {
			return true
		}
	}
	return false
}

func libraryCatalog(s *Source) ([]LibraryTemplate, error) {
	if s == nil {
		return nil, fmt.Errorf("source.catalog_snapshot_required")
	}
	if s.loadedCatalogKey == ([32]byte{}) || s.Root != s.loadedRoot {
		// Arbitrary Source callers still execute the complete derivation. A warm
		// cache must not conceal dependency errors at an unattested source root.
		return buildLibraryCatalog(s)
	}
	if err := verifyCatalogFonts(s); err != nil {
		return nil, err
	}
	return sharedLibraryCatalogCache.catalog(s)
}

func buildLibraryCatalog(s *Source) ([]LibraryTemplate, error) {
	fonts := authoringFonts(s)
	paths := make([]string, 0, len(s.Templates))
	for path := range s.Templates {
		if strings.HasPrefix(path, "templates/library/") && !strings.HasPrefix(path[strings.LastIndex(path, "/")+1:], "_") {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	var out []LibraryTemplate
	seen := map[string]bool{}
	for _, path := range paths {
		raw := s.Templates[path]
		var catalog templateSourceCatalog
		if err := sceneDecode(raw, &catalog); err != nil {
			return nil, err
		}
		if catalog.Schema != "wmds.templates.v2" && !v4UnversionedFamily(s.Revision, path, catalog) {
			return nil, fmt.Errorf("library.unsupported_schema: %s", path)
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(raw))
		pinned := false
		for _, f := range s.Files {
			if f.Path == path && f.SHA256 == digest {
				pinned = true
			}
		}
		if !pinned {
			return nil, fmt.Errorf("library.source_hash_mismatch: %s", path)
		}
		for _, entryRaw := range catalog.Templates {
			var entry libraryEntry
			if err := sceneDecode(entryRaw, &entry); err != nil {
				return nil, err
			}
			key := entry.ID + "/" + entry.Variant
			if seen[key] {
				return nil, fmt.Errorf("library.duplicate_template: %s", key)
			}
			seen[key] = true
			var slide librarySlide
			if err := sceneDecode(entry.Slide, &slide); err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			if slide.Type != "slide" || len(slide.Body) == 0 {
				return nil, fmt.Errorf("library.invalid_source_slide: %s", key)
			}
			def := LibraryTemplate{TemplateDefinition: TemplateDefinition{Key: key, SourceFile: path, SourceSHA256: digest}, Family: catalog.Family, Tier: entry.Tier, Name: entry.Name, ContentContract: LibraryBindingsContract, RawSlide: entry.Slide, RenderStatus: "binding_defined_render_review_pending", Policy: []string{"Named content slots and exact-count key overlays; geometry/base styles remain frozen.", "Source slot metadata is advisory; this definition explicitly names the drawn content fields.", "All scalar content is required; no source-example fallback. Array order remains the declared fixed topology.", "Binding availability does not establish successful rendering, native review or a qualified envelope."}}
			def.SourceRevision = s.Revision
			def.Purpose, def.Uses = entry.Purpose, append([]string(nil), entry.Uses...)
			def.AdvisoryBudget = append(json.RawMessage(nil), entry.Budget...)
			def.AdvisorySlots = append(json.RawMessage(nil), entry.Slots...)
			def.Revision, def.Added, def.Revised, def.Status, def.ReplacedBy = entry.Revision, entry.Added, entry.Revised, entry.Status, entry.ReplacedBy
			if def.Revision == 0 {
				def.Revision = 1
			}
			if def.Status == "" {
				def.Status = "active"
			}
			if def.Status != "active" && def.Status != "deprecated" {
				return nil, fmt.Errorf("library.unknown_lifecycle_status: %s", key)
			}
			if def.Status == "deprecated" && def.ReplacedBy == "" {
				return nil, fmt.Errorf("library.deprecated_requires_replacement: %s", key)
			}
			if slide.Nav != nil {
				if slide.Rail != "nav" {
					return nil, fmt.Errorf("library.nav_without_nav_rail: %s", key)
				}
				if len(slide.Nav.Items) < 2 || len(slide.Nav.Items) > 6 || slide.Nav.Active == nil || *slide.Nav.Active < 0 || *slide.Nav.Active >= len(slide.Nav.Items) {
					return nil, fmt.Errorf("library.invalid_nav: %s", key)
				}
				nav := LibraryNavContent{}
				for i, label := range slide.Nav.Items {
					nav.Items = append(nav.Items, LibraryNavItem{Key: fmt.Sprintf("section-%02d", i+1), Label: label})
				}
				nav.Active = nav.Items[*slide.Nav.Active].Key
				def.Nav = &LibraryNavContract{MinItems: 2, MaxItems: 6, Example: nav}
			} else if slide.Rail == "nav" {
				return nil, fmt.Errorf("library.nav_content_required: %s", key)
			}
			obj, err := libraryObject(entry.Slide)
			if err != nil {
				return nil, err
			}
			for _, field := range []string{"eyebrow", "title"} {
				if v, ok := obj[field]; ok {
					libraryContentWalk(&def, v, "/"+field, field, field, libraryProjectionContext{})
				}
			}
			if v, ok := obj["stamp"]; ok {
				libraryContentWalk(&def, v, "/stamp", "stamp", "stamp", libraryProjectionContext{})
			}
			if v, ok := obj["source"]; ok {
				libraryContentWalk(&def, v, "/source", "source", "source", libraryProjectionContext{})
			}
			body, ok := obj["body"].([]any)
			if !ok {
				return nil, fmt.Errorf("library.invalid_body: %s", key)
			}
			for i, n := range body {
				name := fmt.Sprintf("node%02d", i+1)
				if node, ok := n.(map[string]any); ok {
					if id, ok := node["id"].(string); ok {
						name = id
					}
				}
				pointer := "/body/" + strconv.Itoa(i)
				libraryContentWalk(&def, n, pointer, name, "", libraryProjectionContext{})
				var tag struct {
					Type string `json:"type"`
				}
				json.Unmarshal(slide.Body[i], &tag)
				def.Identities = append(def.Identities, TemplateIdentity{ID: name, SourcePointer: pointer, ExpectedType: tag.Type})
			}
			if usesLegacyTemplate(s.Revision, key) {
				def.ContentContract = TemplateBindingsContract
				def.ValueSchema = libraryLegacyValueSchema(key)
				def.Slots, def.Arrays = nil, nil
				def.Policy = []string{"This template retains its typed v2 values API; the library slots/keys projection is not accepted.", "All declared typed content is required; geometry/base styles remain frozen.", "Binding availability does not establish successful rendering, native review or a qualified envelope."}
			}
			def.Discovery = libraryDiscovery(def, obj)
			authoring, err := libraryAuthoringMetadata(def, obj, s, fonts)
			if err != nil {
				return nil, err
			}
			def.Authoring = &authoring
			out = append(out, def)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	expected := 97
	if isModernLibrary(s.Revision) {
		expected = 167
	}
	if s.Revision == LibraryRevisionV3 {
		expected = 248
	}
	if s.Revision == LibraryRevisionV4 {
		expected = 522
	}
	if s.Revision == LibraryRevisionV5 {
		expected = 587
	}
	if s.Revision == LibraryRevisionV6 {
		expected = 602
	}
	if s.Revision == LibraryRevisionV7 {
		expected = 616
	}
	if s.Revision == LibraryRevisionV8 {
		expected = 623
	}
	if s.Revision == LibraryRevisionV9 {
		expected = 631
	}
	if s.Revision == LibraryRevisionV10 {
		expected = 649
	}
	if len(out) != expected {
		return nil, fmt.Errorf("library.inventory_migration_required: %d templates", len(out))
	}
	return out, nil
}

// Nine family files in the immutable v4/v5 intakes omit the schema key. Accept
// only their exact pinned family identities; earlier revisions keep the v2 gate.
func v4UnversionedFamily(revision, path string, catalog templateSourceCatalog) bool {
	if !isExpandedLibrary(revision) || catalog.Schema != "" || path != "templates/library/"+catalog.Family+".json" {
		return false
	}
	if revision == LibraryRevisionV10 && catalog.Family == "roadmaps" {
		return true
	}
	switch catalog.Family {
	case "change", "diagrams", "heatmaps", "lifecycle", "maturity", "software", "status", "team-curves", "venn":
		return true
	}
	return false
}

// Only the unchanged card APIs carry forward into the refreshed source revision.
func usesLegacyTemplate(revision, key string) bool {
	return legacyTemplate(key) && (!isModernLibrary(revision) || key == "cards/3" || key == "cards/4")
}

func LibraryCatalog(bundle, override string) ([]LibraryTemplate, error) {
	s, err := Load(bundle, override)
	if err != nil {
		return nil, err
	}
	return LibraryCatalogFromSource(s)
}

// LibraryCatalogFromSource reuses the decoded Load-validated snapshot. A
// fabricated or modified Source is rejected; external font/calibration bytes
// are checked again because authoring plans depend on them. Returned definitions
// are independent mutable copies, never the cache's retained definitions.
func LibraryCatalogFromSource(s *Source) ([]LibraryTemplate, error) {
	if s == nil || s.loadedCatalogKey == ([32]byte{}) || s.Root != s.loadedRoot {
		return nil, fmt.Errorf("source.loaded_snapshot_required")
	}
	key, err := libraryCatalogFingerprint(s)
	if err != nil {
		return nil, err
	}
	if key != s.loadedCatalogKey {
		return nil, fmt.Errorf("source.loaded_snapshot_changed")
	}
	if err := verifyCatalogFonts(s); err != nil {
		return nil, err
	}
	return sharedLibraryCatalogCache.catalogKey(s, key)
}

func libraryObject(raw []byte) (map[string]any, error) {
	var obj map[string]any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// The content projection is closed by a source-pinned definition. A token,
// geometry or structural discriminator never becomes an editable content slot.
var libraryFixedString = map[string]bool{"target": true, "placement": true, "arrow": true, "type": true, "key": true, "from": true, "to": true, "id": true, "k": true, "style": true, "surface": true, "on": true, "ink": true, "titleInk": true, "titleStyle": true, "numInk": true, "numStyle": true, "keyInk": true, "markInk": true, "size": true, "bodySize": true, "band": true, "bands": true, "edge": true, "rule": true, "layout": true, "variant": true, "mode": true, "kind": true, "org": true, "state": true, "status": true, "color": true, "colors": true, "swatch": true, "fill": true, "head": true, "elbow": true, "dir": true, "labelPos": true, "labels": true, "numbering": true, "emphasis": true, "mark": true, "icon": true, "focus": true, "align": true, "valign": true, "header": true, "preset": true, "side": true, "rail": true, "footer": true, "railSurface": true, "density": true, "corner": true, "event": true}
var libraryNumbers = map[string]bool{"values": true, "value": true, "alloc": true, "from": true, "to": true, "at": true, "softStart": true, "softEnd": true}
var libraryIntakeFixedString = map[string]bool{"curve": true, "labelStyle": true, "scale": true, "heatScale": true, "orient": true, "ramp": true, "direction": true, "rowHeader": true}

type libraryProjectionContext struct {
	NodeType  string
	Parent    string
	ChartKind string
}

func libraryLegacyValueSchema(key string) *LibraryValueSchema {
	fields := map[string]string{"eyebrow": "string", "title": "string"}
	schema := &LibraryValueSchema{Kind: "typed_values", Reference: "planning/wm-design-contracts/v1/template-execution.md", Fields: fields}
	switch key {
	case "cards/3", "cards/4":
		schema.GoType = "BoundCardRowsContent"
		fields["cards"] = "array of {key:string,title:string,body:string}; unique authored keys; array order controls card order"
		count := 3
		if key == "cards/4" {
			count = 4
		}
		schema.ExactCounts = map[string]int{"cards": count}
	case "stats/four-metrics":
		schema.GoType = "BoundStatsContent"
		for _, f := range []string{"metric1", "metric2", "metric3", "metric4"} {
			fields[f] = "BoundMetric: {value:string XOR format:NumberFormatSpec,label:string}"
		}
		fields["support"] = "string OR RichTextSpec"
		fields["source"] = "string"
	case "takeaway-rail/metrics-rail":
		schema.GoType = "BoundRailContent"
		for _, f := range []string{"metric1", "metric2"} {
			fields[f] = "BoundMetric: {value:string XOR format:NumberFormatSpec,label:string}"
		}
		for _, f := range []string{"support", "railHeading", "railBody"} {
			fields[f] = "string OR RichTextSpec"
		}
		fields["source"] = "string"
		fields["railEyebrow"] = "string"
	}
	return schema
}

func libraryPointerChild(pointer, field string) string {
	return pointer + "/" + strings.ReplaceAll(strings.ReplaceAll(field, "~", "~0"), "/", "~1")
}

func libraryScalarSlot(def *LibraryTemplate, v any, pointer, name, kind string) bool {
	switch kind {
	case "string":
		if _, ok := v.(string); !ok {
			return false
		}
	case "number":
		if _, ok := v.(json.Number); !ok {
			return false
		}
	case "nullable_number":
		if _, ok := v.(json.Number); !ok && v != nil {
			return false
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return false
		}
	default:
		return false
	}
	raw, _ := json.Marshal(v)
	allowEmpty := false
	if text, ok := v.(string); ok {
		allowEmpty = text == ""
	}
	def.Slots = append(def.Slots, LibrarySlot{Name: name, SourcePointer: pointer, Kind: kind, AllowEmpty: allowEmpty, Example: raw})
	return true
}

// A table cell's schema, not the spelling of its column key, decides content.
// Nested numeric-format/icon objects cannot expose styles, geometry or arbitrary
// fields through the previous recursive row exemption.
func libraryTableCellWalk(def *LibraryTemplate, v any, pointer, name, kind string) bool {
	if v == nil {
		return false
	}
	switch kind {
	case "":
		if obj, ok := v.(map[string]any); ok {
			found := false
			for _, k := range []string{"text", "sub"} {
				found = libraryScalarSlot(def, obj[k], libraryPointerChild(pointer, k), name+"."+k, "string") || found
			}
			if isV6OrLaterLibrary(def.SourceRevision) {
				found = libraryScalarSlot(def, obj["ref"], libraryPointerChild(pointer, "ref"), name+".ref", "string") || found
				found = libraryScalarSlot(def, obj["refActive"], libraryPointerChild(pointer, "refActive"), name+".refActive", "boolean") || found
			}
			return found
		}
		return libraryScalarSlot(def, v, pointer, name, "string")
	case "status", "tag", "raci":
		return libraryScalarSlot(def, v, pointer, name, "string")
	case "priority":
		if !isV6OrLaterLibrary(def.SourceRevision) {
			return false
		}
		if obj, ok := v.(map[string]any); ok {
			found := libraryScalarSlot(def, obj["value"], libraryPointerChild(pointer, "value"), name+".value", "string")
			return libraryScalarSlot(def, obj["label"], libraryPointerChild(pointer, "label"), name+".label", "string") || found
		}
		return libraryScalarSlot(def, v, pointer, name, "string")
	case "rating", "dots", "harvey", "heat":
		if obj, ok := v.(map[string]any); ok {
			found := libraryScalarSlot(def, obj["value"], pointer+"/value", name+".value", "number")
			if kind == "dots" || kind == "heat" {
				found = libraryCellTextWalk(def, obj["text"], pointer+"/text", name+".text", kind == "dots") || found
			}
			return found
		}
		return libraryScalarSlot(def, v, pointer, name, "number")
	case "delta", "allocation", "gauge":
		return libraryScalarSlot(def, v, pointer, name, "number")
	case "check", "checkbox":
		return libraryScalarSlot(def, v, pointer, name, "boolean")
	case "icon":
		if _, ok := v.(string); ok {
			return libraryScalarSlot(def, v, pointer, name, "string")
		}
		if obj, ok := v.(map[string]any); ok {
			found := false
			for _, k := range []string{"icon", "text"} {
				if val, ok := obj[k]; ok {
					found = libraryScalarSlot(def, val, libraryPointerChild(pointer, k), name+"."+k, "string") || found
				}
			}
			return found
		}
	case "num":
		if _, ok := v.(string); ok {
			return libraryScalarSlot(def, v, pointer, name, "string")
		}
		if obj, ok := v.(map[string]any); ok {
			value, ok := obj["value"]
			if !ok {
				return false
			}
			if _, ok := value.(json.Number); ok {
				return libraryScalarSlot(def, value, pointer+"/value", name+".value", "number")
			}
			if endpoints, ok := value.([]any); ok {
				found := false
				for i, value := range endpoints {
					found = libraryScalarSlot(def, value, pointer+"/value/"+strconv.Itoa(i), name+".value"+fmt.Sprintf(".item%02d", i+1), "number") || found
				}
				return found
			}
		}
	case "bullets", "maturity":
		if values, ok := v.([]any); ok {
			scalar := "string"
			if kind == "maturity" {
				scalar = "number"
			}
			found := false
			for i, value := range values {
				found = libraryScalarSlot(def, value, pointer+"/"+strconv.Itoa(i), name+fmt.Sprintf(".item%02d", i+1), scalar) || found
			}
			if found {
				def.Arrays = append(def.Arrays, LibraryArray{Name: name, SourcePointer: pointer, Count: len(values)})
			}
			return found
		}
	}
	return false
}

// Score-cell supporting bullets have their own stable keys. Cell policy fields
// (max, ink and scale) remain frozen template definitions rather than content.
func libraryCellTextWalk(def *LibraryTemplate, v any, pointer, name string, bullets bool) bool {
	if libraryScalarSlot(def, v, pointer, name, "string") {
		return true
	}
	if !bullets {
		return false
	}
	values, ok := v.([]any)
	if !ok || len(values) == 0 {
		return false
	}
	start := len(def.Slots)
	for i, value := range values {
		libraryContentWalk(def, value, pointer+"/"+strconv.Itoa(i), name+fmt.Sprintf(".item%02d", i+1), "array_content", libraryProjectionContext{})
	}
	if len(def.Slots) > start {
		def.Arrays = append(def.Arrays, LibraryArray{Name: name, SourcePointer: pointer, Count: len(values)})
		return true
	}
	return false
}

func libraryTableRowsWalk(def *LibraryTemplate, obj map[string]any, pointer, name string) bool {
	columns := map[string]string{}
	if cols, ok := obj["cols"].([]any); ok {
		for _, raw := range cols {
			if col, ok := raw.(map[string]any); ok {
				if key, ok := col["k"].(string); ok {
					kind, _ := col["type"].(string)
					columns[key] = kind
				}
			}
		}
	}
	rows, ok := obj["rows"].([]any)
	if !ok {
		return false
	}
	found := false
	for i, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rowPointer := pointer + "/" + strconv.Itoa(i)
		rowName := name + fmt.Sprintf(".item%02d", i+1)
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if kind, ok := columns[key]; ok {
				found = libraryTableCellWalk(def, row[key], libraryPointerChild(rowPointer, key), rowName+"."+key, kind) || found
			} else if key == "group" && len(row) == 1 {
				found = libraryScalarSlot(def, row[key], rowPointer+"/group", rowName+".group", "string") || found
			}
		}
	}
	if found {
		def.Arrays = append(def.Arrays, LibraryArray{Name: name, SourcePointer: pointer, Count: len(rows)})
	}
	return found
}

func librarySwimlaneLinksWalk(def *LibraryTemplate, v any, pointer, name string) bool {
	links, ok := v.([]any)
	if !ok {
		return false
	}
	found := false
	for i, raw := range links {
		link, ok := raw.([]any)
		if !ok || len(link) < 3 {
			continue
		}
		found = libraryScalarSlot(def, link[2], pointer+"/"+strconv.Itoa(i)+"/2", name+fmt.Sprintf(".item%02d.item03", i+1), "string") || found
	}
	// Link identity is the frozen endpoint pair; labels do not create key arrays.
	return found
}

func libraryContentWalk(def *LibraryTemplate, v any, pointer, name, field string, ctx libraryProjectionContext) bool {
	switch x := v.(type) {
	case map[string]any:
		if typ, ok := x["type"].(string); ok {
			ctx.NodeType = typ
			ctx.ChartKind, _ = x["kind"].(string)
		}
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		found := false
		for _, k := range keys {
			if ctx.NodeType == "gauge" && k == "segments" {
				continue
			}
			childPointer := libraryPointerChild(pointer, k)
			childName := name + "." + k
			if ctx.NodeType == "table" && k == "rows" {
				found = libraryTableRowsWalk(def, x, childPointer, childName) || found
				continue
			}
			if ctx.NodeType == "table" && k == "rowGroups" && isV6OrLaterLibrary(def.SourceRevision) {
				// Inclusive source row ranges and fill define the authored topology.
				// Only labels are copy; group identity has its own stable key array.
				if groups, ok := x[k].([]any); ok && len(groups) > 0 {
					for i, value := range groups {
						if group, ok := value.(map[string]any); ok {
							p := childPointer + "/" + strconv.Itoa(i) + "/label"
							n := childName + fmt.Sprintf(".item%02d.label", i+1)
							found = libraryScalarSlot(def, group["label"], p, n, "string") || found
						}
					}
					def.Arrays = append(def.Arrays, LibraryArray{Name: childName, SourcePointer: childPointer, Count: len(groups)})
				}
				continue
			}
			if ctx.NodeType == "swimlane" && k == "links" {
				found = librarySwimlaneLinksWalk(def, x[k], childPointer, childName) || found
				continue
			}
			if x["type"] == "connector" && (k == "from" || k == "to" || k == "points") {
				continue
			}
			if ctx.NodeType == "maturity" && k == "at" {
				// Stage positions are the template's authored curve geometry.
				continue
			}
			if ctx.NodeType == "maturity" && ctx.Parent == "branch" && k == "from" {
				// The branch anchor is source geometry, not a visible stage label.
				continue
			}
			if ctx.NodeType == "road" && ctx.Parent == "milestones" && k == "at" {
				continue
			}
			if ctx.NodeType == "roadfork" && ctx.Parent == "fork" && k == "at" {
				continue
			}
			if ctx.NodeType == "roadfork" && (k == "chosen" || k == "n") {
				if libraryScalarSlot(def, x[k], childPointer, childName, "number") {
					found = true
					continue
				}
			}
			if ctx.NodeType == "roadfork" && k == "here" {
				found = libraryScalarSlot(def, x[k], childPointer, childName, "boolean") || found
				continue
			}
			if ctx.NodeType == "cycle" && ctx.Parent == "loops" && (k == "from" || k == "to") {
				continue
			}
			if isModernLibrary(def.SourceRevision) && ctx.NodeType == "table" && ctx.Parent == "groups" && (k == "from" || k == "to") {
				// Group row boundaries are topology, not numeric table data.
				// The visible group label remains a required content slot.
				continue
			}
			if k == "text" && x["fill"] != nil && strings.HasPrefix(fmt.Sprint(x[k]), "#") {
				continue
			}
			childCtx := libraryProjectionContext{NodeType: ctx.NodeType, Parent: k, ChartKind: ctx.ChartKind}
			if k == "metric" {
				childCtx.NodeType = "metric"
			}
			projectedField := k
			if isModernLibrary(def.SourceRevision) && ctx.NodeType == "chart" && ctx.Parent == "items" && (k == "x" || k == "y") {
				projectedField = "quadrant_coordinate_content"
			}
			if k == "icon" {
				if _, ok := x[k].(string); ok {
					projectedField = "registry_asset_content"
				}
			}
			if k == "k" && ctx.NodeType == "schedule" && ctx.Parent == "items" {
				projectedField = "schedule_time_content"
			}
			if k == "status" && ctx.NodeType == "metric" {
				projectedField = "metric_status_content"
			}
			if isV5OrLaterLibrary(def.SourceRevision) && k == "status" && ctx.NodeType == "legend" {
				projectedField = "legend_status_content"
			}
			if k == "state" && (ctx.NodeType == "stepper" || ctx.NodeType == "vstepper") && ctx.Parent == "steps" {
				projectedField = "step_state_content"
			}
			if k == "on" && ctx.Parent == "checklist" {
				found = libraryScalarSlot(def, x[k], childPointer, childName, "boolean") || found
				continue
			}
			if k == "n" && (ctx.NodeType == "maturity" || ctx.NodeType == "venn" || ctx.NodeType == "road" || ctx.NodeType == "cycle") || k == "active" && (ctx.NodeType == "maturity" || ctx.NodeType == "cycle") || k == "heat" && (ctx.NodeType == "block" || ctx.NodeType == "legend") {
				if libraryScalarSlot(def, x[k], childPointer, childName, "number") {
					found = true
					continue
				}
			}
			found = libraryContentWalk(def, x[k], childPointer, childName, projectedField, childCtx) || found
		}
		return found
	case []any:
		start := len(def.Slots)
		if isExpandedLibrary(def.SourceRevision) && ctx.NodeType == "chart" && ctx.ChartKind == "line" && field == "values" {
			for i, item := range x {
				if !libraryScalarSlot(def, item, pointer+"/"+strconv.Itoa(i), name+fmt.Sprintf(".item%02d", i+1), "nullable_number") {
					// Malformed source values stay visible to source validation.
					continue
				}
			}
			if len(def.Slots) > start {
				def.Arrays = append(def.Arrays, LibraryArray{Name: name, SourcePointer: pointer, Count: len(x)})
				return true
			}
			return false
		}
		arrayField := field
		if field == "labels" || field == "bands" || field == "points" {
			arrayField = "array_content"
		}
		for i, item := range x {
			libraryContentWalk(def, item, pointer+"/"+strconv.Itoa(i), name+fmt.Sprintf(".item%02d", i+1), arrayField, ctx)
		}
		identityArray := field == "items" && ctx.NodeType == "chart" || ctx.NodeType == "venn" && (field == "sets" || field == "regions" || field == "points") || ctx.NodeType == "maturity" && field == "stages" || ctx.NodeType == "funnel" && field == "stages" || ctx.NodeType == "pyramid" && (field == "levels" || field == "stages") || ctx.NodeType == "road" && field == "milestones" || ctx.NodeType == "cycle" && (field == "items" || field == "loops")
		if (len(def.Slots) > start || identityArray) && len(x) > 0 {
			def.Arrays = append(def.Arrays, LibraryArray{Name: name, SourcePointer: pointer, Count: len(x)})
			return true
		}
	case string:
		if libraryFixedString[field] || libraryIntakeFixedString[field] || field == "h" || field == "w" || field == "points" || field == "links" || field == "page" && pointer == "/page" {
			return false
		}
		return libraryScalarSlot(def, x, pointer, name, "string")
	case json.Number:
		if libraryNumbers[field] || field == "quadrant_coordinate_content" {
			return libraryScalarSlot(def, x, pointer, name, "number")
		}
	}
	return false
}

func compileLibrarySlide(raw json.RawMessage, keys map[string][]string) (SlideSpec, error) {
	var src librarySlide
	if err := sceneDecode(raw, &src); err != nil {
		return SlideSpec{}, err
	}
	q := FrameRequest{Rail: src.Rail, Footer: src.Footer, Surface: src.Surface, RailSurface: src.RailSurface, TitleLines: src.TitleLines, Density: src.Density, NoHeader: src.Title == "" && src.Eyebrow == "", NoPage: src.NoPage, Split: src.Split}
	if src.Nav != nil {
		if src.Rail != "nav" || src.Nav.Active == nil || len(src.Nav.Items) < 2 || len(src.Nav.Items) > 6 || *src.Nav.Active < 0 || *src.Nav.Active >= len(src.Nav.Items) {
			return SlideSpec{}, fmt.Errorf("library.invalid_nav")
		}
		navKeys := keys["/nav/items"]
		if len(navKeys) > 0 && len(navKeys) != len(src.Nav.Items) {
			return SlideSpec{}, fmt.Errorf("library.nav_key_count")
		}
		for i, label := range src.Nav.Items {
			key := fmt.Sprintf("section-%02d", i+1)
			if len(navKeys) > 0 {
				key = navKeys[i]
			}
			q.Nav = append(q.Nav, NavTab{ID: key, Label: label})
		}
		q.Active = q.Nav[*src.Nav.Active].ID
	}
	if q.TitleLines == 0 {
		q.TitleLines = 1
	}
	chrome := &LibraryChrome{Emphasis: src.Emphasis, Stamp: src.Stamp}
	if len(src.Tint) > 0 {
		raw := bytes.TrimSpace(src.Tint)
		if len(raw) > 0 && raw[0] == '{' {
			raw = append(append([]byte{'['}, raw...), ']')
		}
		if err := bindingStrictDecode(raw, &chrome.Tint); err != nil {
			return SlideSpec{}, fmt.Errorf("library.invalid_tint: %w", err)
		}
		if len(chrome.Tint) < 1 || len(chrome.Tint) > 8 {
			return SlideSpec{}, fmt.Errorf("library.invalid_tint_count")
		}
		for i := range chrome.Tint {
			p := &chrome.Tint[i]
			if !intakeFinite(p.X, p.W) || p.X < 0 || p.W <= 0 || p.X+p.W > 960 {
				return SlideSpec{}, fmt.Errorf("library.invalid_tint_geometry")
			}
			if p.Surface == "" {
				p.Surface = "subtle"
			}
		}
	}
	doc := SlideSpec{Frame: q, Eyebrow: src.Eyebrow, Title: src.Title, LibraryChrome: chrome}
	if src.Source != nil {
		chrome.Notes = src.Source.Notes
		doc.Source = ""
		for i, note := range src.Source.Notes {
			if doc.Source != "" {
				doc.Source += "\n"
			}
			doc.Source += strconv.Itoa(i+1) + " " + note
		}
		if src.Source.Text != "" {
			if doc.Source != "" {
				doc.Source += "\n"
			}
			doc.Source += "Source: " + src.Source.Text
		}
		if doc.Source != "" {
			doc.Frame.SourceLines = 1
			if len(src.Source.Notes) > 1 || (len(src.Source.Notes) > 0 && src.Source.Text != "") {
				doc.Frame.SourceLines = 2
			}
		}
	}
	if len(src.Whiteboard) > 0 && !bytes.Equal(bytes.TrimSpace(src.Whiteboard), []byte(`"header"`)) {
		chrome.CustomWhiteboard = true
		if err := sceneDecode(src.Whiteboard, &chrome.Whiteboard); err != nil {
			return SlideSpec{}, err
		}
	}
	seenIDs := map[string]bool{}
	for i, node := range src.Body {
		path := "/body/" + strconv.Itoa(i)
		id := fmt.Sprintf("node%02d", i+1)
		var tag struct{ ID string }
		if err := json.Unmarshal(node, &tag); err != nil {
			return SlideSpec{}, err
		}
		if tag.ID != "" {
			id = tag.ID
		}
		if !validPartKey(id) || seenIDs[id] {
			return SlideSpec{}, fmt.Errorf("library.invalid_or_duplicate_node_id: %s", id)
		}
		seenIDs[id] = true
		doc.Nodes = append(doc.Nodes, Node{ID: id, Kind: "scene", Scene: &SceneSpec{Node: node, Path: path, Keys: keys, Notes: chrome.Notes}})
	}
	return doc, nil
}

func libraryTemplate(catalog []LibraryTemplate, key string) (LibraryTemplate, error) {
	for _, t := range catalog {
		if t.Key == key {
			return t, nil
		}
	}
	return LibraryTemplate{}, fmt.Errorf("library.unknown_template: %s", key)
}
