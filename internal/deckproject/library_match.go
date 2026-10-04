package deckproject

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// PageSpec is content before a template is selected. Each nonempty leaf needs
// a visible destination; IDs supply array identities and never become copy.
type PageSpec struct {
	Title        string     `json:"title" yaml:"title"`
	Eyebrow      string     `json:"eyebrow,omitempty" yaml:"eyebrow,omitempty"`
	Source       string     `json:"source,omitempty" yaml:"source,omitempty"`
	Relationship string     `json:"relationship,omitempty" yaml:"relationship,omitempty"`
	Items        []PageItem `json:"items,omitempty" yaml:"items,omitempty"`
	Callout      *PageItem  `json:"callout,omitempty" yaml:"callout,omitempty"`
}
type PageItem struct {
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Lead   string `json:"lead,omitempty" yaml:"lead,omitempty"`
	Text   string `json:"text,omitempty" yaml:"text,omitempty"`
	Value  string `json:"value,omitempty" yaml:"value,omitempty"`
	Label  string `json:"label,omitempty" yaml:"label,omitempty"`
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
}
type ContentDisposition struct {
	SourcePointer string `json:"source_pointer"`
	Value         string `json:"value"`
	Status        string `json:"status"`
	Destination   string `json:"destination,omitempty"`
}
type MatchCandidate struct {
	Template                 string                    `json:"template"`
	SlideID                  string                    `json:"slide_id"`
	Status                   string                    `json:"status"`
	Error                    string                    `json:"error,omitempty"`
	MappingComplete          bool                      `json:"mapping_complete"`
	AuthoredSlide            map[string]any            `json:"authored_slide,omitempty"`
	BoundSlide               *wmdesign.BoundSlide      `json:"bound_slide,omitempty"`
	MissingSlots             []string                  `json:"missing_slots"`
	UnresolvedFields         []string                  `json:"unresolved_fields"`
	Disposition              []ContentDisposition      `json:"content_disposition"`
	SlotFits                 []wmdesign.LibrarySlotFit `json:"slot_fits"`
	UnsupportedCapacitySlots []string                  `json:"unsupported_capacity_slots"`
	SlideFile                string                    `json:"slide_file,omitempty"`
	Deck                     string                    `json:"deck,omitempty"`
	LayoutReport             string                    `json:"layout_report,omitempty"`
	NativeStatus             string                    `json:"native_status"`
	VisualStatus             string                    `json:"visual_status"`
	MetadataStatus           string                    `json:"metadata_status"`
	deckBytes                []byte
	layout                   *wmdesign.Report
}
type LibraryMatchOptions struct {
	Bundle    string
	Engine    string
	Out       string
	Templates []string
	Limit     int
	Year      int
}
type LibraryMatchReport struct {
	Schema         string           `json:"schema"`
	Engine         string           `json:"engine"`
	SourceRevision string           `json:"source_revision"`
	Candidates     []MatchCandidate `json:"candidates"`
	Passed         int              `json:"passed"`
	Failed         int              `json:"failed"`
	Requested      int              `json:"requested_candidates"`
	LibraryGap     string           `json:"library_gap,omitempty"`
	CombinedDeck   string           `json:"combined_deck,omitempty"`
	Policy         []string         `json:"policy"`
}

func LoadPageSpec(path string) (PageSpec, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return PageSpec{}, e
	}
	var page PageSpec
	if len(data) > 16<<20 {
		return page, fmt.Errorf("page spec exceeds 16MiB")
	}
	parser := &Project{SourcePath: path, Positions: map[string]Position{}, positionFiles: map[string]string{}}
	value, e := parser.parseSource(data, filepath.Base(path), "")
	if e != nil {
		return page, e
	}
	if e = parser.shapeType(value, reflect.TypeOf(page), ""); e != nil {
		return page, e
	}
	if e = strictInto(value, &page); e != nil {
		return page, e
	}
	return page, validatePageSpec(page)
}
func validatePageSpec(page PageSpec) error {
	if strings.TrimSpace(page.Title) == "" {
		return fmt.Errorf("library.match_title_required")
	}
	switch page.Relationship {
	case "", "parallel", "sequence", "comparison", "cycle", "quantities", "table", "ownership":
	default:
		return fmt.Errorf("library.match_unknown_relationship: %s", page.Relationship)
	}
	seen := map[string]bool{}
	for i, item := range page.Items {
		if item.ID != "" {
			if !stableID.MatchString(item.ID) || seen[item.ID] {
				return fmt.Errorf("library.match_invalid_or_duplicate_item_id: %s", item.ID)
			}
			seen[item.ID] = true
		}
		if item.Lead == "" && item.Text == "" && item.Value == "" && item.Label == "" {
			return fmt.Errorf("library.match_empty_item: %d", i)
		}
	}
	return nil
}
func pageLeaves(page PageSpec) map[string]string {
	out := map[string]string{}
	put := func(path, value string) {
		if value != "" {
			out[path] = value
		}
	}
	put("/title", page.Title)
	put("/eyebrow", page.Eyebrow)
	put("/source", page.Source)
	item := func(prefix string, value PageItem) {
		put(prefix+"/lead", value.Lead)
		put(prefix+"/text", value.Text)
		put(prefix+"/value", value.Value)
		put(prefix+"/label", value.Label)
		put(prefix+"/source", value.Source)
	}
	for i, v := range page.Items {
		item(fmt.Sprintf("/items/%d", i), v)
	}
	if page.Callout != nil {
		item("/callout", *page.Callout)
	}
	return out
}
func rolePageField(role string) string {
	switch role {
	case "lead", "title":
		return "lead"
	case "body", "text", "description":
		return "text"
	case "value":
		return "value"
	case "label", "name":
		return "label"
	case "source":
		return "source"
	}
	return ""
}
func matchRelationship(page PageSpec, metadata wmdesign.LibraryAuthoring) bool {
	wanted := page.Relationship
	if wanted == "" {
		wanted = "parallel"
	}
	actual := metadata.Relationship
	if wanted == "quantities" {
		for _, g := range metadata.Groups {
			if g.Role == "metric" {
				return true
			}
		}
		return false
	}
	if wanted == "ownership" {
		return actual == "hierarchy" || actual == "table"
	}
	return wanted == actual
}

// MapPageToTemplate rejects gaps instead of borrowing source examples, dropping
// copy, splitting ideas, shrinking fonts or rewriting supplied business meaning.
func MapPageToTemplate(page PageSpec, def wmdesign.LibraryTemplate, bundle, engine, id string, year int) (MatchCandidate, error) {
	c := MatchCandidate{Template: def.Key, SlideID: id, Status: "mapping_failed", MissingSlots: []string{}, UnresolvedFields: []string{}, Disposition: []ContentDisposition{}, SlotFits: []wmdesign.LibrarySlotFit{}, UnsupportedCapacitySlots: []string{}, NativeStatus: "not_rendered", VisualStatus: "not_reviewed"}
	if e := validatePageSpec(page); e != nil {
		return c, e
	}
	if engine == "" {
		engine = wmdesign.CandidateEngine
	}
	if year == 0 {
		year = 2026
	}
	if id == "" {
		id = "candidate"
	}
	c.SlideID = id
	metadata, e := wmdesign.LibraryAuthoringMetadata(def)
	if e != nil {
		return c, e
	}
	c.MetadataStatus = metadata.ReviewStatus
	leaves := pageLeaves(page)
	consumed := map[string]string{}
	slots := map[string]any{}
	values := map[string]any{}
	typed := def.ContentContract == wmdesign.TemplateBindingsContract
	var selectedGroup string
	if matchRelationship(page, metadata) {
		best := -1
		for _, g := range metadata.Groups {
			if g.Cardinality != len(page.Items) || len(page.Items) == 0 {
				continue
			}
			score := 0
			for _, s := range metadata.Slots {
				if s.Group == g.Alias && rolePageField(s.Role) != "" {
					score++
				}
			}
			if score > best {
				best = score
				selectedGroup = g.Alias
			}
		}
	} else {
		c.Error = "Requested relationship is not established by this source template."
	}
	// More than one source group needs an explicit page field. Choose one item
	// group and one independent metric callout; other groups remain missing.
	calloutGroup := ""
	if page.Callout != nil {
		for _, g := range metadata.Groups {
			if g.Role == "metric" && g.Cardinality == 1 && g.Alias != selectedGroup {
				calloutGroup = g.Alias
				break
			}
		}
	}
	usedSlots := map[string]bool{}
	for _, s := range metadata.Slots {
		path := ""
		switch s.Role {
		case "headline":
			path = "/title"
		case "eyebrow":
			path = "/eyebrow"
		case "source":
			if s.Group == "" {
				path = "/source"
			}
		}
		if path == "" && s.Group == selectedGroup && selectedGroup != "" && s.GroupIndex >= 0 && s.GroupIndex < len(page.Items) {
			if field := rolePageField(s.Role); field != "" {
				path = fmt.Sprintf("/items/%d/%s", s.GroupIndex, field)
			}
		}
		if path == "" && s.Group == calloutGroup && calloutGroup != "" {
			if field := rolePageField(s.Role); field != "" {
				path = "/callout/" + field
			}
		}
		value, present := leaves[path]
		if present && consumed[path] != "" {
			present = false
		}
		if present && s.Kind == "string" {
			consumed[path] = s.Alias
			usedSlots[s.Name] = true
			if !typed {
				slots[s.Name] = value
			}
			continue
		}
		if !typed && s.Classification == "decorative" {
			slots[s.Name] = ""
			continue
		}
		if !typed && s.Kind == "string" && s.AllowEmpty {
			slots[s.Name] = ""
			continue
		}
		c.MissingSlots = append(c.MissingSlots, s.Alias)
	}
	if typed {
		values["title"] = page.Title
		values["eyebrow"] = page.Eyebrow
		if def.ValueSchema != nil && def.ValueSchema.GoType == "BoundCardRowsContent" && selectedGroup == "/cards" {
			cards := []any{}
			for i, item := range page.Items {
				key := item.ID
				if key == "" {
					key = fmt.Sprintf("item-%03d", i+1)
				}
				cards = append(cards, map[string]any{"key": key, "title": item.Lead, "body": item.Text})
			}
			values["cards"] = cards
		} else {
			c.MissingSlots = append(c.MissingSlots, "typed_values_adapter_not_supported")
		}
	} else {
		values["slots"] = slots
		keys := map[string]any{}
		for _, array := range def.Arrays {
			items := make([]any, array.Count)
			for i := range items {
				items[i] = fmt.Sprintf("item-%03d", i+1)
			}
			keys[array.Name] = items
		}
		values["keys"] = keys
		if def.Nav != nil {
			c.MissingSlots = append(c.MissingSlots, "navigation: caller supplied navigation is required")
		}
	}
	pointers := make([]string, 0, len(leaves))
	for path := range leaves {
		pointers = append(pointers, path)
	}
	sort.Strings(pointers)
	for _, path := range pointers {
		disposition := ContentDisposition{SourcePointer: path, Value: leaves[path], Status: "unresolved"}
		if destination := consumed[path]; destination != "" {
			disposition.Status = "mapped_visible"
			disposition.Destination = destination
		} else {
			c.UnresolvedFields = append(c.UnresolvedFields, path)
		}
		c.Disposition = append(c.Disposition, disposition)
	}
	if len(c.MissingSlots) > 0 || len(c.UnresolvedFields) > 0 || c.Error != "" {
		return c, nil
	}
	c.MappingComplete = true
	raw, e := json.Marshal(values)
	if e != nil {
		return c, e
	}
	bound := wmdesign.BoundSlide{ID: id, Template: def.Key, ContentKind: "supplied_content", Values: raw}
	c.BoundSlide = &bound
	slide := Slide{ID: id, Template: Reference{Scope: "shared", ID: def.Key, Revision: fmt.Sprint(def.Revision)}, ContentKind: "supplied_content", Values: values}
	authored, e := StockEditableSlide(slide, def)
	if e != nil {
		return c, e
	}
	c.AuthoredSlide = authored
	typography, e := wmdesign.NewTypographyEngine(filepath.Join(bundle, "fonts"), engine)
	if e != nil {
		return c, e
	}
	overflow := false
	for _, s := range metadata.Slots {
		if !usedSlots[s.Name] {
			continue
		}
		s.Capacity, e = wmdesign.EstimateLibrarySlotCapacity(typography, s.Capacity)
		if e != nil {
			return c, e
		}
		value := ""
		for path, destination := range consumed {
			if destination == s.Alias {
				value = leaves[path]
				break
			}
		}
		fit, e := wmdesign.MeasureLibrarySlot(typography, s, value)
		if e != nil {
			c.Error = e.Error()
			c.Status = "go_measurement_failed"
			return c, nil
		}
		c.SlotFits = append(c.SlotFits, fit)
		if fit.Status == "unsupported" {
			c.UnsupportedCapacitySlots = append(c.UnsupportedCapacitySlots, s.Alias)
		}
		if fit.Status == "overflow_estimate" {
			overflow = true
		}
	}
	if overflow {
		c.Status = "slot_overflow"
		return c, nil
	}
	input := wmdesign.BoundDocument{Schema: wmdesign.BoundDocumentSchema, Year: year, Slides: []wmdesign.BoundSlide{bound}}
	doc, _, e := wmdesign.BindTemplates(bundle, "", input)
	if e != nil {
		c.Status = "go_build_failed"
		c.Error = e.Error()
		return c, nil
	}
	deck, layout, e := wmdesign.BuildWithEngine(bundle, "", doc, engine)
	if e != nil {
		c.Status = "go_build_failed"
		c.Error = e.Error()
		return c, nil
	}
	c.Status = "go_layout_succeeded_native_review_pending"
	c.deckBytes = deck
	c.layout = &layout
	return c, nil
}

func MatchPage(page PageSpec, options LibraryMatchOptions) (LibraryMatchReport, error) {
	report := LibraryMatchReport{Schema: "pptxgengo.library-match.v1", Candidates: []MatchCandidate{}, Policy: []string{"Every supplied content leaf must map visibly; source business copy is never a fallback.", "Metadata is inferred from pinned source structure and may require semantic review.", "Passing means complete content mapping and successful Go build; native and visual review remain pending.", "Unsupported component capacities are listed explicitly; no universal maximum character count is claimed."}}
	if e := validatePageSpec(page); e != nil {
		return report, e
	}
	if options.Engine == "" {
		options.Engine = wmdesign.CandidateEngine
	}
	if options.Limit == 0 {
		options.Limit = 4
	}
	if options.Limit < 1 || options.Limit > 100 {
		return report, fmt.Errorf("library.match_limit_1_to_100")
	}
	report.Requested = options.Limit
	if options.Year == 0 {
		options.Year = 2026
	}
	report.Engine = options.Engine
	catalog, e := wmdesign.LibraryCatalog(options.Bundle, "")
	if e != nil {
		return report, e
	}
	if len(catalog) > 0 {
		report.SourceRevision = catalog[0].SourceRevision
	}
	defs := []wmdesign.LibraryTemplate{}
	if len(options.Templates) > 0 {
		lookup := map[string]wmdesign.LibraryTemplate{}
		for _, def := range catalog {
			lookup[def.Key] = def
		}
		for _, key := range options.Templates {
			def, ok := lookup[key]
			if !ok {
				return report, fmt.Errorf("library.match_unknown_template: %s", key)
			}
			defs = append(defs, def)
		}
	} else {
		structures := []string{}
		if page.Relationship != "" && page.Relationship != "parallel" {
			structures = append(structures, page.Relationship)
		}
		// Search's public result cap is a display limit, not a mapping limit.
		// Drain ranked batches so a usable layout outside the first 100 remains
		// eligible. Ranking uses only per-template features and is unchanged by
		// removing the previous batch.
		remaining := append([]wmdesign.LibraryTemplate(nil), catalog...)
		for len(remaining) > 0 {
			search, e := wmdesign.SearchLibrary(remaining, wmdesign.LibrarySearchOptions{Query: page.Title, Structures: structures, Items: len(page.Items), Limit: 100})
			if e != nil {
				return report, e
			}
			if len(search.Matches) == 0 {
				break
			}
			selected := map[string]bool{}
			for _, hit := range search.Matches {
				defs = append(defs, hit.Template)
				selected[hit.Template.Key] = true
			}
			next := remaining[:0]
			for _, def := range remaining {
				if !selected[def.Key] {
					next = append(next, def)
				}
			}
			remaining = next
		}
	}
	out := options.Out
	if out != "" {
		if e = os.MkdirAll(filepath.Dir(out), 0755); e != nil {
			return report, e
		}
		if e = os.Mkdir(out, 0755); e != nil {
			return report, fmt.Errorf("library.match_new_output_required: %w", e)
		}
		if e = wmdesign.WriteJSON(filepath.Join(out, "page.json"), page); e != nil {
			return report, e
		}
	}
	passing := []wmdesign.BoundSlide{}
	for i, def := range defs {
		if len(options.Templates) == 0 && report.Passed >= options.Limit {
			break
		}
		id := fmt.Sprintf("candidate-%03d", i+1)
		candidate, e := MapPageToTemplate(page, def, options.Bundle, options.Engine, id, options.Year)
		if e != nil {
			return report, e
		}
		if candidate.Status == "go_layout_succeeded_native_review_pending" {
			report.Passed++
			passing = append(passing, *candidate.BoundSlide)
			if out != "" {
				prefix := filepath.Join(out, id)
				candidate.Deck = prefix + ".pptx"
				candidate.LayoutReport = prefix + ".layout.json"
				candidate.SlideFile = prefix + ".yaml"
				metadata, e := wmdesign.LibraryAuthoringMetadata(def)
				if e != nil {
					return report, e
				}
				for i := range metadata.Slots {
					for _, fit := range candidate.SlotFits {
						if fit.Alias == metadata.Slots[i].Alias {
							metadata.Slots[i].Capacity = fit.Capacity
							break
						}
					}
				}
				data, e := MarshalStockSlideSource(candidate.AuthoredSlide, metadata)
				if e != nil {
					return report, e
				}
				if e = os.WriteFile(candidate.SlideFile, data, 0644); e != nil {
					return report, e
				}
				if e = os.WriteFile(candidate.Deck, candidate.deckBytes, 0644); e != nil {
					return report, e
				}
				if e = wmdesign.WriteJSON(candidate.LayoutReport, candidate.layout); e != nil {
					return report, e
				}
			}
		} else {
			report.Failed++
		}
		report.Candidates = append(report.Candidates, candidate)
	}
	if len(options.Templates) == 0 && report.Passed < options.Limit {
		report.LibraryGap = fmt.Sprintf("Only %d complete candidates for %d requested after evaluating %d eligible templates. Remaining layouts require different fields, cardinality, relationship support or Go layout fit; inspect their explicit gaps.", report.Passed, options.Limit, len(report.Candidates))
	}
	if out != "" && len(passing) > 0 {
		input := wmdesign.BoundDocument{Schema: wmdesign.BoundDocumentSchema, Year: options.Year, Slides: passing}
		doc, _, e := wmdesign.BindTemplates(options.Bundle, "", input)
		if e != nil {
			return report, e
		}
		deck, layout, e := wmdesign.BuildWithEngine(options.Bundle, "", doc, options.Engine)
		if e != nil {
			return report, e
		}
		report.CombinedDeck = filepath.Join(out, "candidates.pptx")
		if e = os.WriteFile(report.CombinedDeck, deck, 0644); e != nil {
			return report, e
		}
		if e = wmdesign.WriteJSON(filepath.Join(out, "candidates.layout.json"), layout); e != nil {
			return report, e
		}
	}
	if out != "" {
		if e = wmdesign.WriteJSON(filepath.Join(out, "match-report.json"), report); e != nil {
			return report, e
		}
	}
	return report, nil
}
