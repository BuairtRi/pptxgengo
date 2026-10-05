package wmdesign

import (
	"fmt"
	"strconv"
	"strings"
)

// These recipes interpret frozen field/geometry topology, never supplied copy.
// They are engineering mappings awaiting human semantic review.
func authoringFamilyRecipes(out *LibraryAuthoring, def LibraryTemplate, obj map[string]any) {
	if !isV5OrLaterLibrary(def.SourceRevision) {
		return
	}
	recipe := ""
	switch {
	case def.Key == "key-message/statement":
		recipe = "key-message-statement-source-topology.v1"
	case def.Key == "vendors/scorecard-generic":
		recipe = "vendors-scorecard-generic-source-topology.v1"
	case def.Key == "status/classic":
		recipe = "status-classic-source-topology.v1"
	case def.Key == "interviews-detail/dense":
		recipe = "interview-detail-dense-source-topology.v1"
	case strings.HasPrefix(def.Key, "interviews-readout/"):
		recipe = "interview-readout-columns.v1"
	case def.Family == "heatmaps":
		recipe = "heatmap-component-fields.v1"
	case strings.HasPrefix(def.Key, "roadmap/"):
		recipe = "roadmap-component-fields.v1"
	case def.Key == "offers-onepager/classic":
		recipe = "offer-classic-source-topology.v1"
	case def.Key == "decision/buy-build-economics":
		recipe = "buy-build-source-topology.v1"
	default:
		return
	}
	for i := range out.Slots {
		s := &out.Slots[i]
		node, index := authoringNode(obj, s.SourcePointer)
		if node == nil || s.Classification == "decorative" {
			continue
		}
		parts := authoringParts(s.SourcePointer)
		kind := authoringKind(node)
		tail := parts[2:]
		switch recipe {
		case "status-classic-source-topology.v1":
			if kind == "table" {
				group := "/status/overall"
				if index == 1 {
					group = "/status/areas"
				}
				if index == 6 {
					group = "/status/risks"
				}
				if len(tail) >= 3 && tail[0] == "rows" {
					field := authoringTableField(tail[2], node)
					if field == "status" || field == "overall_status" {
						field = "state"
					}
					s.Role = field
					s.Group, s.GroupIndex = group, mustOrdinal(tail[1])
					rows, _ := node["rows"].([]any)
					s.Cardinality = len(rows)
					s.Alias = fmt.Sprintf("%s/item_%02d/%s", group, s.GroupIndex+1, field)
				} else {
					s.Alias = group + "/" + authoringAliasTail(tail)
				}
			} else if kind == "bullets" || kind == "grouplabel" {
				section := map[int]string{2: "completed", 3: "completed", 4: "planned", 5: "planned", 7: "decisions", 8: "decisions"}[index]
				if section == "" {
					continue
				}
				s.Alias = "/status/" + section + "/" + authoringAliasTail(tail)
			} else {
				continue
			}
		case "interview-detail-dense-source-topology.v1":
			x, _ := discoveryNumber(node["x"])
			y, _ := discoveryNumber(node["y"])
			if kind == "text" && y < 198 {
				field := ""
				switch {
				case x == 57 && y == 126:
					field = "stakeholder_label"
				case x == 57 && y == 144:
					field = "stakeholder_name"
				case x == 57 && y == 168:
					field = "stakeholder_role"
				case x == 273 && y == 126:
					field = "department_date_label"
				case x == 273 && y == 144:
					field = "department"
				case x == 273 && y == 168:
					field = "date_and_duration"
				case x == 489 && y == 126:
					field = "format_label"
				case x == 489 && y == 144:
					field = "format_and_location"
				case x == 705 && y == 126:
					field = "interviewers_label"
				case x == 705 && y == 144:
					field = "interviewers"
				}
				if field == "" {
					continue
				}
				s.Alias, s.Role = "/interview/"+field, field
				s.Group, s.GroupIndex, s.Cardinality = "", -1, 0
			} else if (kind == "textblock" && y == 210) || (kind == "bullets" && y == 264) {
				column := int((x - 57) / 288)
				if column < 0 || column > 2 {
					continue
				}
				s.Alias = fmt.Sprintf("/topics/item_%02d/%s", column+1, authoringAliasTail(tail))
				s.Group, s.GroupIndex, s.Cardinality = "/topics", column, 3
			} else if kind == "pullquote" {
				s.Alias = "/interview_quote/" + authoringAliasTail(tail)
			} else if (kind == "textblock" || kind == "bullets") && x == 417 && y >= 336 {
				s.Alias = "/pain_and_opportunities/" + authoringAliasTail(tail)
			} else if (kind == "textblock" || kind == "bullets") && x == 633 && y >= 336 {
				s.Alias = "/follow_ups_and_questions/" + authoringAliasTail(tail)
			} else {
				continue
			}
		case "vendors-scorecard-generic-source-topology.v1":
			if kind == "text" && index == 1 {
				s.Alias, s.Role = "/comparison/legend", "source"
				s.Group, s.GroupIndex, s.Cardinality = "", -1, 0
			} else if kind == "table" && index == 0 && len(tail) == 3 {
				columns, _ := node["cols"].([]any)
				if tail[0] == "cols" && tail[2] == "label" {
					col := mustOrdinal(tail[1])
					switch col {
					case 0:
						s.Alias = "/comparison/criterion_label"
					case 1:
						s.Alias = "/comparison/weight_label"
					default:
						s.Alias = fmt.Sprintf("/comparison/options/item_%02d/name", col-1)
					}
				} else if tail[0] == "rows" {
					row := mustOrdinal(tail[1])
					switch tail[2] {
					case "c":
						s.Alias = fmt.Sprintf("/comparison/criteria/item_%02d/label", row+1)
					case "w":
						s.Alias = fmt.Sprintf("/comparison/criteria/item_%02d/weight", row+1)
					default:
						for col, raw := range columns {
							column, _ := raw.(map[string]any)
							if col >= 2 && column["k"] == tail[2] {
								s.Alias = fmt.Sprintf("/comparison/options/item_%02d/values/item_%02d", col-1, row+1)
							}
						}
					}
				} else {
					continue
				}
				s.Role = "comparison-field"
				s.Group, s.GroupIndex, s.Cardinality = "", -1, 0
			} else {
				continue
			}
		case "key-message-statement-source-topology.v1":
			if kind != "text" {
				continue
			}
			switch node["style"] {
			case "display":
				s.Alias, s.Role = "/headline", "headline"
			case "eyebrow":
				s.Alias, s.Role = "/eyebrow", "eyebrow"
			case "lead":
				s.Alias, s.Role = "/supporting_statement", "body"
			default:
				continue
			}
			s.Group, s.GroupIndex, s.Cardinality = "", -1, 0
		case "interview-readout-columns.v1":
			if kind != "table" {
				continue
			}
			if len(tail) >= 3 && tail[0] == "rows" {
				row, e := strconv.Atoi(tail[1])
				if e != nil {
					continue
				}
				field := authoringTableField(tail[2], node)
				s.Alias = fmt.Sprintf("/interviewees/item_%02d/%s", row+1, field)
				if len(tail) > 3 {
					s.Alias += "/" + strings.Join(tail[3:], "/")
				}
				rows, _ := node["rows"].([]any)
				s.Group = "/interviewees"
				s.GroupIndex = row
				s.Cardinality = len(rows)
				s.Role = field
			} else if len(tail) >= 3 && tail[0] == "cols" {
				s.Alias = "/interview_columns/" + fmt.Sprintf("item_%02d", mustOrdinal(tail[1])+1) + "/" + strings.Join(tail[2:], "/")
				s.Group = "/interview_columns"
				s.Role = "label"
			}
		case "heatmap-component-fields.v1":
			prefix := fmt.Sprintf("/heatmap/item_%02d", index+1)
			if kind == "table" && len(tail) >= 3 && tail[0] == "rows" {
				s.Alias = prefix + "/rows/" + fmt.Sprintf("item_%02d", mustOrdinal(tail[1])+1) + "/" + authoringTableField(tail[2], node)
				if len(tail) > 3 {
					suffix := strings.Join(tail[3:], "/")
					if def.SourceRevision == LibraryRevisionV6 {
						suffix = authoringAliasTail(tail[3:])
					}
					s.Alias += "/" + suffix
				}
				s.Role = "heatmap-cell"
				if def.SourceRevision == LibraryRevisionV6 {
					if len(tail) > 3 && tail[3] == "ref" {
						s.Role = "reference_id"
					}
					if len(tail) > 3 && tail[3] == "refActive" {
						s.Role = "reference_active"
					}
					if columns, ok := node["cols"].([]any); ok {
						for _, raw := range columns {
							column, _ := raw.(map[string]any)
							if column["k"] != tail[2] {
								continue
							}
							if column["type"] == "priority" {
								s.Role = "priority_value"
								if len(tail) > 3 && tail[3] == "label" {
									s.Role = "priority_label"
								}
							}
							if column["type"] == "bullets" {
								s.Role = "evidence_body"
							}
						}
					}
				}
				s.Group = prefix + "/rows"
				s.GroupIndex = mustOrdinal(tail[1])
				rows, _ := node["rows"].([]any)
				s.Cardinality = len(rows)
			} else if kind == "table" {
				s.Alias = prefix + "/" + authoringAliasTail(tail)
				if def.SourceRevision == LibraryRevisionV6 && len(tail) >= 3 && tail[0] == "rowGroups" {
					s.Alias = prefix + "/row_groups/" + authoringAliasTail(tail[1:])
					s.Group = prefix + "/row_groups"
					s.GroupIndex = mustOrdinal(tail[1])
					groups, _ := node["rowGroups"].([]any)
					s.Cardinality = len(groups)
					s.Role = "group_label"
				}
				if len(tail) >= 2 && tail[0] == "cols" {
					s.Group = prefix + "/columns"
					s.GroupIndex = mustOrdinal(tail[1])
					cols, _ := node["cols"].([]any)
					s.Cardinality = len(cols)
				}
			} else if kind == "legend" {
				s.Alias = fmt.Sprintf("/heatmap_legends/item_%02d/", index+1) + authoringAliasTail(tail)
			} else {
				continue
			}
		case "roadmap-component-fields.v1":
			if kind != "gantt" && kind != "timeaxis" {
				continue
			}
			s.Alias = "/roadmap/" + authoringAliasTail(tail)
			if s.Group != "" {
				groupParts := authoringParts(s.Group)
				if len(groupParts) > 2 {
					s.Group = "/roadmap/" + strings.Join(groupParts[2:], "/")
				}
			}
		case "offer-classic-source-topology.v1":
			x, _ := discoveryNumber(node["x"])
			y, _ := discoveryNumber(node["y"])
			field := ""
			if kind == "text" {
				switch {
				case y == 126 && x == 57:
					field = "/offer/narrative_label"
				case y == 144 && x == 57:
					field = "/offer/narrative"
				case y == 138 && x == 645:
					field = "/fee/label"
				case y == 156 && x == 645:
					field = "/fee/value"
				case y == 159 && x == 765:
					field = "/fee/terms"
				case y == 177 && x == 765:
					field = "/fee/payment_schedule"
				case y == 228:
					field = "/approach/label"
				case y == 405 && x == 69:
					field = "/deliverables/label"
				case y == 402 && x == 165:
					field = "/deliverables/text"
				case y == 405 && x == 501:
					field = "/assumptions/label"
				case y == 402 && x == 597:
					field = "/assumptions/text"
				}
			}
			if (kind == "text" && (y == 258 || y == 276)) || kind == "bullets" && y == 303 {
				phase := int((x - 57) / 216)
				s.Group = "/approach/phases"
				s.GroupIndex = phase
				s.Cardinality = 4
				field = fmt.Sprintf("/approach/phases/item_%02d/", phase+1)
				if y == 258 {
					field += "timing"
				} else if y == 276 {
					field += "title"
				} else {
					field += "activities/" + authoringAliasTail(tail[1:])
				}
			}
			if field == "" {
				continue
			}
			s.Alias = field
			if field == "/fee/value" {
				s.Role = "value"
			}
		case "buy-build-source-topology.v1":
			x, _ := discoveryNumber(node["x"])
			y, _ := discoveryNumber(node["y"])
			if kind == "card" && index < 2 {
				s.Alias = fmt.Sprintf("/options/item_%02d/%s", index+1, authoringAliasTail(tail))
				s.Group = "/options"
				s.GroupIndex = index
				s.Cardinality = 2
			} else if (kind == "text" || kind == "bullets") && (x == 75 || x == 327) {
				option := 0
				if x == 327 {
					option = 1
				}
				name := "best_when"
				if y >= 324 {
					name = "tradeoffs"
				}
				s.Alias = fmt.Sprintf("/options/item_%02d/%s", option+1, name)
				if kind == "text" {
					s.Alias += "_label"
				} else {
					s.Alias += "/" + authoringAliasTail(tail[1:])
				}
				s.Group = "/options"
				s.GroupIndex = option
				s.Cardinality = 2
			} else if x >= 561 {
				s.Alias = fmt.Sprintf("/economics/item_%02d/%s", index+1, authoringAliasTail(tail))
			} else {
				continue
			}
		}
		s.ReviewStatus = "explicit_source_recipe_unreviewed"
		s.Description = strings.ReplaceAll(strings.TrimPrefix(s.Alias, "/"), "/", " ") + "; pinned source field " + s.SourcePointer
	}
	out.Recipe = recipe
	out.SemanticStatus = "engineering_recipe_requires_review"
	if recipe == "buy-build-source-topology.v1" || recipe == "vendors-scorecard-generic-source-topology.v1" {
		out.Relationship = "comparison"
	}
	if recipe == "offer-classic-source-topology.v1" || recipe == "roadmap-component-fields.v1" {
		out.Relationship = "sequence"
	}
}
func mustOrdinal(s string) int { n, _ := strconv.Atoi(s); return n }
func authoringAliasTail(parts []string) string {
	out := make([]string, len(parts))
	for i, s := range parts {
		if n, e := strconv.Atoi(s); e == nil {
			out[i] = fmt.Sprintf("item_%02d", n+1)
		} else {
			if s == "cols" {
				s = "columns"
			}
			out[i] = authoringField(s)
		}
	}
	return strings.Join(out, "/")
}
