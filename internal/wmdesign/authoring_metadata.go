package wmdesign

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const LibraryAuthoringSchema = "pptxgengo.library-authoring.v1"

// Authoring metadata is a separate projection: closed contract names, source
// pointers, examples and geometry remain untouched. Inference is never review.
type LibraryAuthoring struct {
	Schema         string                  `json:"schema"`
	SourceRevision string                  `json:"source_revision"`
	SourceSHA256   string                  `json:"source_sha256"`
	ReviewStatus   string                  `json:"review_status"`
	Recipe         string                  `json:"recipe,omitempty"`
	SemanticStatus string                  `json:"semantic_status"`
	Relationship   string                  `json:"relationship"`
	Groups         []LibraryAuthoringGroup `json:"groups"`
	Slots          []LibraryAuthoringSlot  `json:"slots"`
	Policy         []string                `json:"policy"`
}
type LibraryAuthoringGroup struct {
	Alias         string `json:"alias"`
	Role          string `json:"role"`
	Cardinality   int    `json:"cardinality"`
	SourcePointer string `json:"source_pointer"`
	Basis         string `json:"basis"`
}
type LibraryAuthoringSlot struct {
	Name                string              `json:"name"`
	Alias               string              `json:"alias"`
	SourcePointer       string              `json:"source_pointer"`
	Kind                string              `json:"kind"`
	AllowEmpty          bool                `json:"allow_empty"`
	Role                string              `json:"role"`
	Group               string              `json:"group,omitempty"`
	GroupIndex          int                 `json:"group_index"`
	Cardinality         int                 `json:"cardinality"`
	Description         string              `json:"description"`
	Classification      string              `json:"classification"`
	ClassificationBasis string              `json:"classification_basis"`
	ReviewStatus        string              `json:"review_status"`
	Capacity            LibrarySlotCapacity `json:"capacity"`
}

func authoringField(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		} else if b.Len() > 0 {
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "content"
	}
	return out
}
func authoringParts(pointer string) []string {
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	for i, s := range parts {
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(s, "~1", "/"), "~0", "~")
	}
	return parts
}
func authoringNode(obj map[string]any, pointer string) (map[string]any, int) {
	parts := authoringParts(pointer)
	if len(parts) < 3 || parts[0] != "body" {
		return nil, -1
	}
	n, e := strconv.Atoi(parts[1])
	body, _ := obj["body"].([]any)
	if e != nil || n < 0 || n >= len(body) {
		return nil, -1
	}
	node, _ := body[n].(map[string]any)
	return node, n
}
func authoringKind(node map[string]any) string { kind, _ := node["type"].(string); return kind }
func authoringGroupName(kind string) string {
	switch kind {
	case "card", "cardrow", "minicard":
		return "cards"
	case "textblock":
		return "sections"
	case "text":
		return "paragraphs"
	case "block":
		return "panels"
	case "metric", "stat", "kpi", "gauge":
		return "metrics"
	case "person", "pod", "role":
		return "people"
	case "phase", "phases", "phasehead":
		return "phases"
	case "step", "stepper", "vstepper", "chevron":
		return "steps"
	case "table", "matrix", "raci":
		return "tables"
	case "bullets", "list", "ol", "strongnum":
		return "lists"
	case "quote", "pullquote":
		return "quotations"
	default:
		return authoringField(kind) + "s"
	}
}
func authoringRole(field string, node map[string]any) string {
	if field == "text" && node != nil && authoringKind(node) == "text" {
		if node["style"] == "subhead" || node["style"] == "lead" {
			return "lead"
		}
		if node["style"] == "eyebrow" || node["style"] == "label" {
			return "label"
		}
	}
	switch field {
	case "title":
		return "lead"
	case "label", "name", "heading", "k":
		return "label"
	case "text", "body", "p", "description", "sub", "support":
		return "body"
	case "value", "number", "n", "amount":
		return "value"
	case "by", "author":
		return "attribution"
	case "source":
		return "source"
	case "icon", "image", "photo", "asset":
		return "asset"
	}
	if node != nil && authoringKind(node) == "text" {
		if node["style"] == "subhead" || node["style"] == "lead" {
			return "lead"
		}
		if node["style"] == "eyebrow" || node["style"] == "label" {
			return "label"
		}
	}
	return authoringField(field)
}

// LibraryAuthoringMetadata also supports manually constructed definitions. For
// catalog definitions, source styles and frames have already been pinned.
func LibraryAuthoringMetadata(def LibraryTemplate) (LibraryAuthoring, error) {
	if def.Authoring != nil {
		return *def.Authoring, nil
	}
	obj := map[string]any{}
	if len(def.RawSlide) > 0 {
		if err := json.Unmarshal(def.RawSlide, &obj); err != nil {
			return LibraryAuthoring{}, err
		}
	}
	return libraryAuthoringMetadata(def, obj, nil)
}
func libraryAuthoringMetadata(def LibraryTemplate, obj map[string]any, source *Source, sharedFonts ...authoringFontLoader) (LibraryAuthoring, error) {
	fonts := authoringFonts(source)
	if len(sharedFonts) > 0 {
		fonts = sharedFonts[0]
	}
	var capacityErr error
	checkedFonts := func() (*Typography, error) {
		t, err := fonts()
		if err != nil {
			capacityErr = err
		}
		return t, err
	}
	componentCapacities := fixedSceneAuthoringCapacities(def, obj, source, checkedFonts)
	out := LibraryAuthoring{Schema: LibraryAuthoringSchema, SourceRevision: def.SourceRevision, SourceSHA256: def.SourceSHA256, ReviewStatus: "inferred_from_pinned_source", SemanticStatus: "structural_inference_requires_review", Relationship: "unknown", Groups: []LibraryAuthoringGroup{}, Slots: []LibraryAuthoringSlot{}, Policy: []string{"Aliases are stable source structure projections, independent of supplied copy.", "Inferred metadata has not received semantic human review; recipes are explicit engineering interpretations.", "Decorative fields stay in the original values contract; capacity is advisory and native fit is not evaluated."}}
	for _, r := range def.Discovery.Relationships {
		out.Relationship = r.Kind
		break
	}
	for _, structure := range def.Discovery.Structures {
		if structure == "sequence" || structure == "cycle" || structure == "comparison" || structure == "table" {
			out.Relationship = structure
			break
		}
	}
	if out.Relationship == "unknown" {
		for _, group := range def.Discovery.Groups {
			switch group.ComponentType {
			case "card", "cardrow", "textblock", "bullets", "list", "ol", "metric":
				if group.ExactCount > 1 {
					out.Relationship = "parallel"
				}
			}
		}
		if def.ValueSchema != nil && def.ValueSchema.GoType == "BoundCardRowsContent" {
			out.Relationship = "parallel"
		}
	}
	body, _ := obj["body"].([]any)
	counts := map[string]int{}
	ordinal := map[int]int{}
	for i, raw := range body {
		node, _ := raw.(map[string]any)
		name := authoringGroupName(authoringKind(node))
		ordinal[i] = counts[name]
		counts[name]++
	}
	groups := map[string]LibraryAuthoringGroup{}
	aliases := map[string]bool{}
	for _, slot := range def.Slots {
		node, index := authoringNode(obj, slot.SourcePointer)
		parts := authoringParts(slot.SourcePointer)
		field := parts[len(parts)-1]
		s := LibraryAuthoringSlot{Name: slot.Name, SourcePointer: slot.SourcePointer, Kind: slot.Kind, AllowEmpty: slot.AllowEmpty, Role: authoringRole(field, node), GroupIndex: -1, Cardinality: 1, Classification: "content", ClassificationBasis: "declared_editable_scalar", ReviewStatus: out.ReviewStatus}
		alias := []string{}
		switch slot.SourcePointer {
		case "/title":
			alias = []string{"headline"}
			s.Role = "headline"
		case "/eyebrow":
			alias = []string{"section_label"}
			s.Role = "eyebrow"
		case "/source/text":
			alias = []string{"source_note"}
			s.Role = "source"
		default:
			if node != nil {
				kind := authoringKind(node)
				name := authoringGroupName(kind)
				s.Group = "/" + name
				s.GroupIndex = ordinal[index]
				s.Cardinality = counts[name]
				alias = []string{name, fmt.Sprintf("item_%02d", s.GroupIndex+1)}
				for partIndex, p := range parts[2:] {
					if n, e := strconv.Atoi(p); e == nil {
						alias = append(alias, fmt.Sprintf("item_%02d", n+1))
					} else {
						if p == "cols" {
							p = "columns"
						}
						if partIndex == 2 && parts[2] == "rows" {
							p = authoringTableField(p, node)
						}
						alias = append(alias, authoringField(p))
					}
				}
				// Rich typed arrays describe the primary repeated units directly.
				for _, g := range def.Discovery.Groups {
					if g.Scope == "primary" && strings.HasPrefix(slot.SourcePointer, g.SourcePointer+"/") && g.SourcePointer != "/body" {
						rel := authoringParts(strings.TrimPrefix(slot.SourcePointer, g.SourcePointer))
						if len(rel) > 0 {
							if n, e := strconv.Atoi(rel[0]); e == nil {
								s.Group = "/" + name + fmt.Sprintf("/item_%02d", ordinal[index]+1) + "/" + authoringField(g.SourcePointer[strings.LastIndex(g.SourcePointer, "/")+1:])
								s.GroupIndex = n
								s.Cardinality = g.ExactCount
								groups[s.Group] = LibraryAuthoringGroup{s.Group, g.Role, g.ExactCount, g.SourcePointer, g.Basis}
							}
						}
					}
				}
				if kind == "block" && field == "text" && slot.AllowEmpty && string(slot.Example) == `""` {
					s.Classification = "decorative"
					s.ClassificationBasis = "pinned_empty_block_surface_text"
				}
				if kind == "text" && strings.HasPrefix(def.Key, "cover/") {
					st, _ := node["style"].(string)
					names := map[string]string{"display": "headline", "eyebrow": "section_label", "lead": "subtitle"}
					if names[st] != "" {
						candidate := "/" + names[st]
						if !aliases[candidate] {
							alias = []string{names[st]}
							s.Group = ""
							s.GroupIndex = -1
							s.Cardinality = 1
							if st == "display" {
								s.Role = "headline"
							}
						}
					}
				}
				if _, present := groups[s.Group]; !present && s.Group != "" {
					groups[s.Group] = LibraryAuthoringGroup{s.Group, discoveryRole(kind), s.Cardinality, "/body", "same_component_type_source_siblings"}
				}
			} else {
				for _, p := range parts {
					if n, e := strconv.Atoi(p); e == nil {
						alias = append(alias, fmt.Sprintf("item_%02d", n+1))
					} else {
						alias = append(alias, authoringField(p))
					}
				}
			}
		}
		s.Alias = "/" + strings.Join(alias, "/")
		if aliases[s.Alias] {
			return out, fmt.Errorf("authoring.alias_collision: %s %s", def.Key, s.Alias)
		}
		aliases[s.Alias] = true
		s.Description = fmt.Sprintf("%s (%s) at %s", strings.ReplaceAll(s.Role, "_", " "), slot.Kind, s.Alias)
		s.Capacity = authoringCapacity(def, obj, node, s, source)
		if capacity, ok := componentCapacities[s.SourcePointer]; ok {
			s.Capacity = capacity
		}
		out.Slots = append(out.Slots, s)
	}
	if def.ValueSchema != nil {
		authoringTypedSlots(&out, def, source, obj, checkedFonts)
	}
	if capacityErr != nil {
		return out, fmt.Errorf("authoring capacity pinned font/calibration load: %w", capacityErr)
	}
	for _, group := range out.Groups {
		groups[group.Alias] = group
	}
	out.Groups = nil
	authoringLifecycleRecipe(&out, def, obj, source)
	authoringFamilyRecipes(&out, def, obj)
	// Rebuild recipe groups after generic slot construction.
	used := map[string]bool{}
	for _, s := range out.Slots {
		if s.Group != "" {
			used[s.Group] = true
			if _, ok := groups[s.Group]; !ok {
				role := "repeated-content"
				switch {
				case s.Group == "/phases" || strings.HasPrefix(s.Group, "/approach/phases") || strings.HasPrefix(s.Group, "/roadmap"):
					role = "sequence-item"
				case s.Group == "/interviewees":
					role = "person"
				case s.Group == "/interview_columns" || strings.HasSuffix(s.Group, "/columns"):
					role = "column"
				case s.Group == "/options":
					role = "point"
				case strings.HasPrefix(s.Group, "/heatmap"):
					role = "tabular-data"
				}
				groups[s.Group] = LibraryAuthoringGroup{s.Group, role, s.Cardinality, "/body", "explicit_source_recipe"}
			}
		}
	}
	for key, g := range groups {
		if used[key] {
			out.Groups = append(out.Groups, g)
		}
	}
	sort.Slice(out.Groups, func(i, j int) bool { return out.Groups[i].Alias < out.Groups[j].Alias })
	return out, nil
}

func authoringTableField(field string, node map[string]any) string {
	columns, _ := node["cols"].([]any)
	for _, raw := range columns {
		col, _ := raw.(map[string]any)
		key := col["key"]
		if key == nil {
			key = col["k"]
		}
		if key != field {
			continue
		}
		label, _ := col["label"].(string)
		name := authoringField(label)
		if label == "" {
			return field
		}
		for _, otherRaw := range columns {
			other, _ := otherRaw.(map[string]any)
			otherLabel, _ := other["label"].(string)
			otherKey := other["key"]
			if otherKey == nil {
				otherKey = other["k"]
			}
			if otherKey != field && authoringField(otherLabel) == name {
				return name + "_" + authoringField(field)
			}
		}
		return name
	}
	return field
}

func authoringTypedSlots(out *LibraryAuthoring, def LibraryTemplate, source *Source, obj map[string]any, fonts authoringFontLoader) {
	cardCapacities := typedCardAuthoringCapacities(def, source, fonts)
	keys := make([]string, 0, len(def.ValueSchema.Fields))
	for key := range def.ValueSchema.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if count, ok := def.ValueSchema.ExactCounts[key]; ok && key == "cards" {
			out.Groups = append(out.Groups, LibraryAuthoringGroup{"/cards", "point", count, "values/cards", "typed_values_exact_count"})
			for i := 0; i < count; i++ {
				for _, field := range []string{"title", "body"} {
					s := LibraryAuthoringSlot{Name: fmt.Sprintf("cards.%d.%s", i, field), Alias: fmt.Sprintf("/cards/%d/%s", i, field), SourcePointer: fmt.Sprintf("/cards/%d/%s", i, field), Kind: "string", Role: authoringRole(field, nil), Group: "/cards", GroupIndex: i, Cardinality: count, Classification: "content", ClassificationBasis: "typed_values_contract", ReviewStatus: out.ReviewStatus, Description: fmt.Sprintf("Card %d %s", i+1, field)}
					s.Capacity = unknownSlotCapacity("typed_component_capacity_not_available")
					if capacity, ok := cardCapacities[s.Alias]; ok {
						s.Capacity = capacity
					}
					out.Slots = append(out.Slots, s)
				}
			}
		} else {
			alias := key
			role := authoringRole(key, nil)
			if key == "title" {
				alias = "headline"
				role = "headline"
			}
			if key == "eyebrow" {
				alias = "section_label"
				role = "eyebrow"
			}
			s := LibraryAuthoringSlot{Name: key, Alias: "/" + alias, SourcePointer: "/" + key, Kind: "string", Role: role, GroupIndex: -1, Cardinality: 1, Description: strings.ReplaceAll(key, "_", " "), Classification: "content", ClassificationBasis: "typed_values_contract", ReviewStatus: out.ReviewStatus}
			s.Capacity = authoringCapacity(def, obj, nil, s, source)
			out.Slots = append(out.Slots, s)
		}
	}
}

// The lifecycle scaffold that motivated this feature is built from separate
// empty panels and text nodes. Source x/style/y topology, never copy, identifies
// three columns, each with a title, objective and four activity title/body pairs.
func authoringLifecycleRecipe(out *LibraryAuthoring, def LibraryTemplate, obj map[string]any, source *Source) {
	if def.Key != "lifecycle/three-phases" {
		return
	}
	body, _ := obj["body"].([]any)
	var heads []map[string]any
	for _, raw := range body {
		n, _ := raw.(map[string]any)
		if n["type"] == "text" && n["style"] == "subhead" {
			heads = append(heads, n)
		}
	}
	sort.Slice(heads, func(i, j int) bool {
		a, _ := discoveryNumber(heads[i]["x"])
		b, _ := discoveryNumber(heads[j]["x"])
		return a < b
	})
	if len(heads) != 3 {
		return
	}
	for i := range out.Slots {
		s := &out.Slots[i]
		n, _ := authoringNode(obj, s.SourcePointer)
		if n == nil || n["type"] != "text" {
			continue
		}
		x, _ := discoveryNumber(n["x"])
		y, _ := discoveryNumber(n["y"])
		col := -1
		for j, h := range heads {
			hx, _ := discoveryNumber(h["x"])
			if x >= hx-12 && x < hx+246 {
				col = j
				break
			}
		}
		if col < 0 {
			continue
		}
		s.Group = "/phases"
		s.GroupIndex = col
		s.Cardinality = 3
		s.ReviewStatus = "explicit_source_recipe_unreviewed"
		s.Alias = fmt.Sprintf("/phases/item_%02d/", col+1)
		if n["style"] == "subhead" {
			s.Alias += "title"
			s.Role = "lead"
		} else if y < 252 {
			s.Alias += "objective"
			s.Role = "body"
		} else {
			activity := int((y - 252) / 54)
			if activity < 0 || activity > 3 {
				continue
			}
			s.Alias += fmt.Sprintf("activities/item_%02d/", activity+1)
			if n["style"] == "small" {
				s.Alias += "description"
				s.Role = "body"
			} else {
				s.Alias += "title"
				s.Role = "lead"
			}
		}
		s.Description = "Phase " + strconv.Itoa(col+1) + " " + strings.ReplaceAll(strings.TrimPrefix(s.Alias, fmt.Sprintf("/phases/item_%02d/", col+1)), "/", " ")
	}
	out.Relationship = "sequence"
	out.Recipe = "lifecycle-three-phases-source-topology.v1"
	out.SemanticStatus = "engineering_recipe_requires_review"
}

type LibraryAuthoringCoverage struct {
	Schema                   string   `json:"schema"`
	Templates                int      `json:"templates"`
	Slots                    int      `json:"slots"`
	DecorativeSlots          int      `json:"decorative_slots"`
	InferredTemplates        int      `json:"inferred_templates"`
	ReviewedTemplates        int      `json:"reviewed_templates"`
	GenericTemplates         []string `json:"generic_templates"`
	AmbiguousTemplates       []string `json:"ambiguous_templates"`
	RecipeTemplates          int      `json:"recipe_templates"`
	UnsupportedCapacitySlots int      `json:"unsupported_capacity_slots"`
}

func AuthoringCoverage(catalog []LibraryTemplate) (LibraryAuthoringCoverage, error) {
	out := LibraryAuthoringCoverage{Schema: "pptxgengo.authoring-coverage.v1", Templates: len(catalog), GenericTemplates: []string{}, AmbiguousTemplates: []string{}}
	for _, def := range catalog {
		a, e := LibraryAuthoringMetadata(def)
		if e != nil {
			return out, e
		}
		if a.ReviewStatus == "reviewed" {
			out.ReviewedTemplates++
		} else {
			out.InferredTemplates++
		}
		generic := false
		ambiguous := false
		if a.Recipe != "" {
			out.RecipeTemplates++
		}
		for _, s := range a.Slots {
			out.Slots++
			if s.Classification == "decorative" {
				out.DecorativeSlots++
			}
			if s.Capacity.Status == "unsupported" {
				out.UnsupportedCapacitySlots++
			}
			for _, p := range authoringParts(s.Alias) {
				if authoringGenericPart(p) {
					generic = true
				}
				if s.Classification != "decorative" && (p == "paragraphs" || p == "panels" || p == "nodes") {
					ambiguous = true
				}
			}
		}
		if generic {
			out.GenericTemplates = append(out.GenericTemplates, def.Key)
		}
		if ambiguous || a.Relationship == "unknown" {
			out.AmbiguousTemplates = append(out.AmbiguousTemplates, def.Key)
		}
	}
	return out, nil
}

func authoringGenericPart(part string) bool {
	for _, prefix := range []string{"node", "block_"} {
		if strings.HasPrefix(part, prefix) {
			suffix := strings.TrimPrefix(part, prefix)
			if _, err := strconv.Atoi(suffix); err == nil {
				return true
			}
		}
	}
	return false
}
