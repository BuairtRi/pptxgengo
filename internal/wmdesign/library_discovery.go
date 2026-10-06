package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
)

const LibraryDiscoverySchema = "pptxgengo.wmds-discovery.v1"

// Discovery describes the pinned source, not a measured envelope for new copy.
// Its vocabulary is independent of the source author's scenario and slot labels.
type LibraryDiscovery struct {
	Schema         string                `json:"schema"`
	Basis          string                `json:"basis"`
	ContentRoles   []string              `json:"content_roles"`
	Structures     []string              `json:"structures"`
	VisualForms    []string              `json:"visual_forms"`
	ComponentTypes []string              `json:"component_types"`
	Groups         []LibraryContentGroup `json:"content_groups,omitempty"`
	Relationships  []LibraryRelationship `json:"relationships,omitempty"`
	Zones          []LibraryContentZone  `json:"zones,omitempty"`
	Frame          LibraryFrameMetadata  `json:"frame"`
	Capability     LibraryCapability     `json:"capability"`
}

type LibraryContentGroup struct {
	SourcePointer string `json:"source_pointer"`
	ComponentType string `json:"component_type"`
	Role          string `json:"role"`
	ExactCount    int    `json:"exact_count"`
	Basis         string `json:"basis"`
	Scope         string `json:"scope"`
	ScopeBasis    string `json:"scope_basis,omitempty"`
}

type LibraryRelationship struct {
	Kind           string   `json:"kind"`
	Basis          string   `json:"basis"`
	SourcePointers []string `json:"source_pointers"`
}

type LibraryContentZone struct {
	SourcePointer string         `json:"source_pointer"`
	ComponentType string         `json:"component_type"`
	Role          string         `json:"role"`
	Bounds        map[string]any `json:"source_bounds,omitempty"`
}

type LibraryFrameMetadata struct {
	Rail    string `json:"rail,omitempty"`
	Footer  string `json:"footer,omitempty"`
	Surface string `json:"surface,omitempty"`
	Split   string `json:"split,omitempty"`
	Nav     bool   `json:"navigation"`
	Basis   string `json:"basis"`
}

type LibraryCapability struct {
	ContentAdapter  string `json:"content_adapter"`
	BuildForQuery   string `json:"build_for_query"`
	SpecimenReview  string `json:"specimen_review"`
	ContentEnvelope string `json:"content_envelope"`
	CapacityBasis   string `json:"capacity_basis"`
}

type LibraryDiscoveryVocabulary struct {
	ContentRoles []string `json:"content_roles"`
	Structures   []string `json:"structures"`
	VisualForms  []string `json:"visual_forms"`
}

func DiscoveryVocabulary() LibraryDiscoveryVocabulary {
	return LibraryDiscoveryVocabulary{
		ContentRoles: []string{"column", "content", "headline", "icon", "image", "key-message", "metric", "navigation", "numeric-data", "person", "point", "quotation", "relationship", "repeated-content", "sequence-item", "tabular-data"},
		Structures:   []string{"comparison", "cycle", "funnel", "gantt", "heatmap", "hierarchy", "layers", "maturity", "matrix", "network", "process", "repeated-items", "sequence", "swimlanes", "table", "timeline", "venn"},
		VisualForms:  []string{"cards", "chart", "diagram", "icon", "image", "table", "text"},
	}
}

func discoveryTokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

func discoverySorted(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Roles are structural affordances, not promises that every item is interchangeable.
func discoveryRole(kind string) string {
	switch kind {
	case "text", "textblock", "card", "cardrow", "block", "bullets", "list", "ol", "bullet", "bulletlist", "iconrow", "numbered", "num", "strongnum", "pillar", "checkrow", "check", "minicard":
		return "point"
	case "metric", "stat", "kpi", "progress", "gauge":
		return "metric"
	case "image", "imageframe", "photo", "picture", "logoslot":
		return "image"
	case "icon":
		return "icon"
	case "table", "tablerow", "tablegroup", "matrix", "raci":
		return "tabular-data"
	case "chart", "teamcurve":
		return "numeric-data"
	case "step", "phase", "phases", "phasehead", "schedule", "timeaxis", "stepper", "vstepper", "chevron", "gate", "milestone", "timeline", "gantt", "swimlane":
		return "sequence-item"
	case "layerrow", "layer", "diagram", "node", "flow", "orgchart", "pyramid", "funnel", "cycle", "road", "roadfork", "bracket", "cylinder", "beforeafter", "device", "plane", "dotmap", "venn", "maturity":
		return "relationship"
	case "quote", "pullquote":
		return "quotation"
	case "callout", "takeaway", "statement":
		return "key-message"
	default:
		return "content"
	}
}

func discoveryNodeRole(node map[string]any) string {
	kind, _ := node["type"].(string)
	if kind == "block" {
		if node["text"] == "" {
			return "content"
		}
		if node["style"] == "label" {
			return "relationship"
		}
	}
	if kind == "card" {
		for _, field := range []string{"metric", "metricGroup", "person", "bio", "quote"} {
			if node[field] != nil {
				switch field {
				case "metric", "metricGroup":
					return "metric"
				case "quote":
					return "quotation"
				default:
					return "person"
				}
			}
		}
	}
	if kind == "person" || kind == "pod" || kind == "role" {
		return "person"
	}
	return discoveryRole(kind)
}

func discoveryNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	case float64:
		return v, true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

func discoveryListCount(node map[string]any) int {
	count := 0
	if body, ok := node["body"].([]any); ok {
		for _, raw := range body {
			if block, ok := raw.(map[string]any); ok {
				if items, ok := block["bullets"].([]any); ok {
					count += len(items)
				}
			}
		}
	}
	return count
}

func discoveryBodyWidth(obj map[string]any) float64 {
	left, right := math.Inf(1), math.Inf(-1)
	if body, ok := obj["body"].([]any); ok {
		for _, raw := range body {
			if node, ok := raw.(map[string]any); ok {
				x, xOK := discoveryNumber(node["x"])
				w, wOK := discoveryNumber(node["w"])
				if xOK && wOK && w > 0 {
					left, right = math.Min(left, x), math.Max(right, x+w)
				}
			}
		}
	}
	if math.IsInf(left, 0) || math.IsInf(right, 0) {
		return 0
	}
	return right - left
}

// A paired comparison is corroborated by equal list topology, aligned cards,
// and an explicit connector joining those cards. Copy and scenario labels do
// not participate. Sequential cards without parallel lists are not sufficient.
func discoveryPairedComparison(body []any) []LibraryRelationship {
	var out []LibraryRelationship
	for i, raw := range body {
		left, ok := raw.(map[string]any)
		if !ok || left["type"] != "card" || discoveryListCount(left) < 2 {
			continue
		}
		for j := i + 1; j < len(body); j++ {
			right, ok := body[j].(map[string]any)
			if !ok || right["type"] != "card" || discoveryListCount(right) != discoveryListCount(left) {
				continue
			}
			lx, lxOK := discoveryNumber(left["x"])
			ly, lyOK := discoveryNumber(left["y"])
			lw, lwOK := discoveryNumber(left["w"])
			lh, lhOK := discoveryNumber(left["h"])
			rx, rxOK := discoveryNumber(right["x"])
			ry, ryOK := discoveryNumber(right["y"])
			rw, rwOK := discoveryNumber(right["w"])
			rh, rhOK := discoveryNumber(right["h"])
			if !lxOK || !lyOK || !lwOK || !lhOK || !rxOK || !ryOK || !rwOK || !rhOK || math.Abs(ly-ry) > 1 || math.Abs(lw-rw) > 1 || rx < lx+lw {
				continue
			}
			for k, rawLink := range body {
				link, ok := rawLink.(map[string]any)
				if !ok || link["type"] != "connector" {
					continue
				}
				points, ok := link["points"].([]any)
				if !ok || len(points) < 2 {
					continue
				}
				first, firstOK := points[0].([]any)
				last, lastOK := points[len(points)-1].([]any)
				if !firstOK || !lastOK || len(first) != 2 || len(last) != 2 {
					continue
				}
				x1, a := discoveryNumber(first[0])
				x2, b := discoveryNumber(last[0])
				y1, c := discoveryNumber(first[1])
				y2, e := discoveryNumber(last[1])
				if a && b && c && e && math.Abs(x1-(lx+lw)) <= 1 && math.Abs(x2-rx) <= 1 && y1 >= ly && y1 <= ly+lh && y2 >= ry && y2 <= ry+rh {
					out = append(out, LibraryRelationship{Kind: "comparison", Basis: "parallel_equal_lists_aligned_cards_explicit_connector", SourcePointers: []string{fmt.Sprintf("/body/%d", i), fmt.Sprintf("/body/%d", j), fmt.Sprintf("/body/%d", k)}})
				}
			}
		}
	}
	return out
}

func libraryDiscovery(def LibraryTemplate, obj map[string]any) LibraryDiscovery {
	d := LibraryDiscovery{Schema: LibraryDiscoverySchema, Basis: "derived_from_pinned_source_structure; labels_are_soft_signals", Frame: LibraryFrameMetadata{Basis: "pinned_source_slide_chrome"}, Capability: LibraryCapability{ContentAdapter: "defined_closed_binding", BuildForQuery: "not_executed", SpecimenReview: "not_loaded_by_discovery", ContentEnvelope: "not_established_by_discovery", CapacityBasis: "source_advisory_not_measured"}}
	if len(def.PendingCapabilities) > 0 {
		d.Capability.ContentAdapter = "capabilities_pending"
	}
	d.Frame.Rail, _ = obj["rail"].(string)
	d.Frame.Footer, _ = obj["footer"].(string)
	d.Frame.Surface, _ = obj["surface"].(string)
	d.Frame.Split, _ = obj["split"].(string)
	d.Frame.Nav = obj["nav"] != nil
	roles, structures, forms, types := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	if _, ok := obj["title"]; ok {
		roles["headline"], roles["key-message"] = true, true
	}
	if d.Frame.Nav {
		roles["navigation"] = true
	}
	var walk func(any, string)
	walk = func(value any, pointer string) {
		switch v := value.(type) {
		case map[string]any:
			kind, _ := v["type"].(string)
			parts := strings.Split(pointer, "/")
			// Source slide body entries are scene nodes. Nested objects are
			// component data, even when a table cell has a field named type.
			isNode := len(parts) == 3 && parts[1] == "body"
			if kind != "" && isNode {
				types[kind] = true
				role := discoveryNodeRole(v)
				roles[role] = true
				bounds := map[string]any{}
				for _, field := range []string{"x", "y", "w", "h"} {
					if x, ok := v[field]; ok {
						bounds[field] = x
					}
				}
				d.Zones = append(d.Zones, LibraryContentZone{SourcePointer: pointer, ComponentType: kind, Role: role, Bounds: bounds})
				switch role {
				case "image", "icon":
					forms[role] = true
				case "tabular-data":
					forms["table"], structures["table"] = true, true
				case "numeric-data":
					forms["chart"] = true
				case "relationship":
					forms["diagram"] = true
				case "sequence-item":
					structures["sequence"] = true
				case "point":
					forms["text"] = true
				}
				if kind == "beforeafter" {
					structures["comparison"] = true
					d.Relationships = append(d.Relationships, LibraryRelationship{Kind: "comparison", Basis: "explicit_beforeafter_component", SourcePointers: []string{pointer}})
				}
				switch kind {
				case "gantt":
					structures["gantt"], structures["timeline"] = true, true
				case "timeline", "timeaxis", "schedule", "milestone", "gate":
					structures["timeline"] = true
				case "swimlane":
					structures["swimlanes"] = true
				case "step", "phase", "phases", "stepper", "vstepper", "chevron":
					structures["process"] = true
				case "layerrow", "layer", "plane":
					structures["layers"] = true
				case "orgchart", "pyramid":
					structures["hierarchy"] = true
				case "flow", "connector":
					structures["network"], forms["diagram"] = true, true
				case "matrix", "raci":
					structures["matrix"] = true
				case "teamcurve":
					structures["timeline"] = true
				case "venn":
					structures["venn"] = true
					d.Relationships = append(d.Relationships, LibraryRelationship{Kind: "intersection", Basis: "explicit_venn_component", SourcePointers: []string{pointer}})
				case "maturity":
					structures["maturity"] = true
				case "funnel":
					structures["funnel"] = true
				case "cycle":
					structures["cycle"], structures["process"] = true, true
				case "road":
					structures["timeline"] = true
				case "roadfork":
					structures["timeline"], structures["sequence"], structures["network"] = true, true, true
					roles["sequence-item"] = true
					mode, _ := v["mode"].(string)
					if mode == "" {
						mode = "parallel"
					}
					if mode == "parallel" || mode == "decision" {
						d.Relationships = append(d.Relationships, LibraryRelationship{Kind: mode, Basis: "explicit_roadfork_mode", SourcePointers: []string{pointer + "/trunk", pointer + "/fork", pointer + "/branches"}})
					}
				case "table":
					if cols, ok := v["cols"].([]any); ok {
						for _, raw := range cols {
							if col, ok := raw.(map[string]any); ok && col["type"] == "heat" {
								structures["heatmap"] = true
							}
						}
					}
				case "block":
					if _, ok := v["heat"]; ok {
						structures["heatmap"] = true
					}
				}
				if strings.Contains(kind, "card") {
					forms["cards"] = true
				}
			}
			// Explicit artwork selectors also describe optional visual affordances.
			for _, field := range []string{"photo", "image", "icon"} {
				if raw, ok := v[field]; ok && raw != nil && raw != "" {
					form := field
					if form == "photo" {
						form = "image"
					}
					forms[form], roles[form] = true, true
				}
			}
			keys := make([]string, 0, len(v))
			for key := range v {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				walk(v[key], pointer+"/"+key)
			}
		case []any:
			counts := map[string]int{}
			spans := map[string][2]float64{}
			for _, item := range v {
				if node, ok := item.(map[string]any); ok {
					if kind, ok := node["type"].(string); ok && pointer == "/body" && kind != "text" && kind != "num" && kind != "strongnum" && kind != "check" {
						group := kind + ":" + discoveryNodeRole(node)
						counts[group]++
						if x, ok := discoveryNumber(node["x"]); ok {
							if w, ok := discoveryNumber(node["w"]); ok && w > 0 {
								span, exists := spans[group]
								if !exists {
									span = [2]float64{x, x + w}
								} else {
									span = [2]float64{math.Min(span[0], x), math.Max(span[1], x+w)}
								}
								spans[group] = span
							}
						}
					}
				}
			}
			kinds := make([]string, 0, len(counts))
			for kind := range counts {
				kinds = append(kinds, kind)
			}
			sort.Strings(kinds)
			for _, kind := range kinds {
				fields := strings.SplitN(kind, ":", 2)
				if counts[kind] > 1 && fields[1] != "content" {
					scope := "primary"
					if fields[1] == "point" {
						span, ok := spans[kind]
						mainWidth := 846.0
						if d.Frame.Rail == "left" || d.Frame.Rail == "right" {
							mainWidth = 558
						}
						if !ok || span[1]-span[0] < mainWidth*.45 {
							scope = "supporting"
						}
					}
					d.Groups = append(d.Groups, LibraryContentGroup{SourcePointer: pointer, ComponentType: fields[0], Role: fields[1], ExactCount: counts[kind], Basis: "same_type_and_role_scene_siblings", Scope: scope, ScopeBasis: "combined_horizontal_span_of_siblings"})
					structures["repeated-items"] = true
				}
			}
			for i, item := range v {
				walk(item, fmt.Sprintf("%s/%d", pointer, i))
			}
		}
	}
	walk(obj, "")
	if body, ok := obj["body"].([]any); ok {
		paired := discoveryPairedComparison(body)
		if len(paired) > 0 {
			structures["comparison"] = true
			d.Relationships = append(d.Relationships, paired...)
		}
	}
	// Binding arrays capture repeated items that have no individual node type.
	for _, array := range def.Arrays {
		field := array.SourcePointer[strings.LastIndex(array.SourcePointer, "/")+1:]
		// Binding arrays are implementation identities, not always semantic
		// units. Cell tuples, rich-text runs and body paragraph blocks do not
		// create another item group merely because they contain editable text.
		switch field {
		case "items", "cards", "rows", "steps", "phases", "lanes", "layers", "levels", "milestones", "trunk", "branches", "loops", "children", "bullets", "points", "people", "results", "secondary", "cols", "columns", "sets", "regions", "stages":
		default:
			continue
		}
		role, parentType, parentDepth, parentPointer := "repeated-content", "binding-array", 0, ""
		var parentBounds map[string]any
		for _, zone := range d.Zones {
			if strings.HasPrefix(array.SourcePointer, zone.SourcePointer+"/") && len(zone.SourcePointer) > parentDepth {
				role, parentType, parentDepth, parentPointer = zone.Role, zone.ComponentType, len(zone.SourcePointer), zone.SourcePointer
				parentBounds = zone.Bounds
			}
		}
		scope := "nested"
		if array.SourcePointer == parentPointer+"/"+field {
			scope = "primary"
		}
		if field == "bullets" || field == "points" {
			role = "point"
			if parentType == "connector" || parentType == "chart" {
				continue
			}
		}
		if field == "cols" || field == "columns" {
			role = "column"
		}
		if parentType == "roadfork" && (field == "trunk" || field == "milestones") {
			role = "sequence-item"
		}
		if parentType == "card" && scope == "primary" {
			// Lists inside a card are supporting details of that card.
			scope = "nested"
		}
		scopeBasis := "array_position_within_source_component"
		if scope == "primary" && role == "point" && parentType != "cardrow" {
			width, ok := discoveryNumber(parentBounds["w"])
			bodyWidth := discoveryBodyWidth(obj)
			if !ok || bodyWidth == 0 || width < 0.6*bodyWidth {
				scope, scopeBasis = "supporting", "point_list_occupies_narrow_source_column_or_has_unknown_width"
			}
		}
		d.Groups = append(d.Groups, LibraryContentGroup{SourcePointer: array.SourcePointer, ComponentType: parentType, Role: role, ExactCount: array.Count, Basis: "semantic_closed_binding_array", Scope: scope, ScopeBasis: scopeBasis})
		if array.Count > 1 {
			structures["repeated-items"] = true
		}
	}
	if def.ValueSchema != nil {
		keys := make([]string, 0, len(def.ValueSchema.ExactCounts))
		for key := range def.ValueSchema.ExactCounts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			count := def.ValueSchema.ExactCounts[key]
			role, kind := "repeated-content", "typed-binding-array"
			if def.ValueSchema.GoType == "BoundCardRowsContent" && key == "cards" {
				role, kind = "point", "cardrow"
			}
			d.Groups = append(d.Groups, LibraryContentGroup{SourcePointer: "values/" + key, ComponentType: kind, Role: role, ExactCount: count, Basis: "typed_values_exact_count", Scope: "primary"})
			if count > 1 {
				structures["repeated-items"] = true
			}
		}
	}
	d.ContentRoles, d.Structures = discoverySorted(roles), discoverySorted(structures)
	d.VisualForms, d.ComponentTypes = discoverySorted(forms), discoverySorted(types)
	return d
}

type LibrarySearchOptions struct {
	EngineHint        string   `json:"requested_engine,omitempty"`
	Query             string   `json:"query,omitempty"`
	ContentRoles      []string `json:"content_roles,omitempty"`
	Structures        []string `json:"structures,omitempty"`
	VisualForms       []string `json:"visual_forms,omitempty"`
	Items             int      `json:"items,omitempty"`
	ItemRole          string   `json:"item_role,omitempty"`
	Limit             int      `json:"limit,omitempty"`
	IncludeDeprecated bool     `json:"include_deprecated,omitempty"`
}

type LibrarySearchHit struct {
	Template       LibraryTemplate    `json:"template"`
	Score          int                `json:"score"`
	ScenarioScore  int                `json:"scenario_score"`
	Reasons        []string           `json:"reasons"`
	UnmatchedHints []string           `json:"unmatched_hints,omitempty"`
	FitStatus      string             `json:"fit_status"`
	CountMatch     *LibraryCountMatch `json:"count_match,omitempty"`
}

type LibraryCountMatch struct {
	Group  LibraryContentGroup `json:"group"`
	Weight int                 `json:"weight"`
}

type LibrarySearchResult struct {
	Engine     LibrarySearchEngine        `json:"engine"`
	Schema     string                     `json:"schema"`
	Query      LibrarySearchOptions       `json:"query"`
	Vocabulary LibraryDiscoveryVocabulary `json:"vocabulary"`
	Policy     []string                   `json:"policy"`
	Matches    []LibrarySearchHit         `json:"matches"`
}

type LibrarySearchEngine struct {
	Requested     string `json:"requested,omitempty"`
	Compatibility string `json:"compatibility"`
}

func libraryScenarioQueryCoverage(def LibraryTemplate, query string) int {
	available := map[string]bool{}
	for _, token := range discoveryTokens(strings.Join([]string{def.Key, def.Name, def.Family, def.Purpose, strings.Join(def.Uses, " ")}, " ")) {
		available[token] = true
	}
	seen := map[string]bool{}
	matched := 0
	for _, token := range discoveryTokens(query) {
		if !seen[token] && available[token] {
			matched++
		}
		seen[token] = true
	}
	return matched
}

// SearchLibrary ranks scenario and structure independently. Hints do not exclude
// a usable template just because its original labels describe another scenario.
func SearchLibrary(catalog []LibraryTemplate, options LibrarySearchOptions) (LibrarySearchResult, error) {
	result := LibrarySearchResult{Schema: "pptxgengo.wmds-library-search.v1", Query: options, Vocabulary: DiscoveryVocabulary(), Policy: []string{"All query fields are ranking hints; only deprecated lifecycle is filtered by default.", "Structure is derived from the pinned source; purpose and slot budgets remain advisory.", "A score is not a measured fit or native review result. Inspect the contract and build actual content."}, Matches: []LibrarySearchHit{}}
	result.Engine = LibrarySearchEngine{Requested: options.EngineHint, Compatibility: "not_evaluated_search_only"}
	if options.Items < 0 || options.Limit < 0 || options.Limit > 100 {
		return result, fmt.Errorf("library.search_invalid_options: items must be nonnegative and limit 0..100")
	}
	if strings.TrimSpace(options.ItemRole) != "" && options.Items == 0 {
		return result, fmt.Errorf("library.search_invalid_options: item_role requires items greater than zero")
	}
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	queryCoverage := map[string]int{}
	for _, def := range catalog {
		if def.Status == "deprecated" && !options.IncludeDeprecated {
			continue
		}
		hit := LibrarySearchHit{Template: def, Reasons: []string{}, FitStatus: "not_measured_for_query"}
		for _, dimension := range []struct {
			name      string
			requested []string
			available []string
			weight    int
		}{{"role", options.ContentRoles, def.Discovery.ContentRoles, 12}, {"structure", options.Structures, def.Discovery.Structures, 16}, {"visual-form", options.VisualForms, def.Discovery.VisualForms, 6}} {
			seen := map[string]bool{}
			for _, hint := range dimension.requested {
				hint = strings.ToLower(strings.TrimSpace(hint))
				if hint == "" || seen[hint] {
					continue
				}
				seen[hint] = true
				found := false
				for _, available := range dimension.available {
					if available == hint {
						found = true
						break
					}
				}
				if found {
					hit.Score += dimension.weight
					hit.Reasons = append(hit.Reasons, dimension.name+" matches "+hint)
				} else {
					hit.UnmatchedHints = append(hit.UnmatchedHints, dimension.name+":"+hint)
				}
			}
		}
		if options.Items > 0 {
			itemRole := strings.ToLower(strings.TrimSpace(options.ItemRole))
			for _, group := range def.Discovery.Groups {
				if group.ExactCount == options.Items && (itemRole == "" || itemRole == group.Role) {
					weight := 10
					if group.Scope == "primary" {
						weight = 35
					} else if group.Scope == "supporting" {
						weight = 15
					}
					if hit.CountMatch == nil || weight > hit.CountMatch.Weight {
						hit.CountMatch = &LibraryCountMatch{Group: group, Weight: weight}
					}
					hit.Reasons = append(hit.Reasons, fmt.Sprintf("%d source items at %s (%s; role=%s; scope=%s; %s)", options.Items, group.SourcePointer, group.ComponentType, group.Role, group.Scope, group.Basis))
				}
			}
			if hit.CountMatch != nil {
				hit.Score += hit.CountMatch.Weight
			} else {
				hit.UnmatchedHints = append(hit.UnmatchedHints, fmt.Sprintf("items:%d; role:%s; no matching exact-count group found", options.Items, itemRole))
			}
		}
		identityWords, annotationWords := map[string]bool{}, map[string]bool{}
		for _, token := range discoveryTokens(strings.Join([]string{def.Key, def.Name, def.Family}, " ")) {
			identityWords[token] = true
		}
		for _, token := range discoveryTokens(strings.Join([]string{def.Purpose, strings.Join(def.Uses, " ")}, " ")) {
			annotationWords[token] = true
		}
		seen := map[string]bool{}
		textScore := 0
		matchedTokens := 0
		for _, token := range discoveryTokens(options.Query) {
			if seen[token] {
				continue
			}
			seen[token] = true
			weight, field := 0, ""
			if identityWords[token] {
				weight, field = 8, "key/name/family"
			} else if annotationWords[token] {
				weight, field = 3, "purpose/uses"
			}
			if weight > 0 {
				matchedTokens++
				hit.ScenarioScore += weight
				bonus := min(weight, 12-textScore)
				hit.Score += bonus
				textScore += bonus
				hit.Reasons = append(hit.Reasons, "scenario/label token matches "+token+" in "+field)
			} else {
				hit.UnmatchedHints = append(hit.UnmatchedHints, "text:"+token)
			}
		}
		queryCoverage[def.Key] = matchedTokens
		if len(seen) > 1 && matchedTokens == len(seen) {
			// Purpose annotations can express the complete scenario even when
			// one incidental key/name token has a higher individual weight.
			// Keep the existing 12-point text budget so structural hints retain
			// their independent influence on the ranking.
			bonus := min(6, 12-textScore)
			hit.Score += bonus
			hit.ScenarioScore += bonus
			hit.Reasons = append(hit.Reasons, "scenario matches all query tokens in key/name/family/purpose/uses")
		}
		result.Matches = append(result.Matches, hit)
	}
	sort.Slice(result.Matches, func(i, j int) bool {
		if result.Matches[i].Score != result.Matches[j].Score {
			return result.Matches[i].Score > result.Matches[j].Score
		}
		if queryCoverage[result.Matches[i].Template.Key] != queryCoverage[result.Matches[j].Template.Key] {
			return queryCoverage[result.Matches[i].Template.Key] > queryCoverage[result.Matches[j].Template.Key]
		}
		if result.Matches[i].ScenarioScore != result.Matches[j].ScenarioScore {
			return result.Matches[i].ScenarioScore > result.Matches[j].ScenarioScore
		}
		return result.Matches[i].Template.Key < result.Matches[j].Template.Key
	})
	if len(result.Matches) > limit {
		result.Matches = result.Matches[:limit]
	}
	return result, nil
}
