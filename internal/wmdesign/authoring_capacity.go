package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
)

type authoringFontLoader func() (*Typography, error)

// A catalog construction is serial and owns one mutable shaping engine. Reuse
// its pinned font/calibration load across fixed component plans; never retain
// this shaping engine in the shared catalog or share it between callers.
func authoringFonts(source *Source) authoringFontLoader {
	var typography *Typography
	var err error
	loaded := false
	return func() (*Typography, error) {
		if !loaded {
			loaded = true
			if source == nil {
				err = fmt.Errorf("pinned source unavailable")
			} else {
				fontRoot := source.loadedFontRoot
				if fontRoot == "" {
					fontRoot = filepath.Join(filepath.Dir(source.Root), "fonts")
				}
				typography, err = NewSourceTypographyEngine(source, fontRoot, CandidateEngine)
			}
		}
		return typography, err
	}
}

func preferredTemplateDensity(def LibraryTemplate, source *Source) SlideDensityRecord {
	comfortable := SlideDensityRecord{Requested: "comfortable", Resolved: "comfortable", Header: "comfortable"}
	if source == nil || len(def.RawSlide) == 0 {
		return comfortable
	}
	slide, err := compileLibrarySlide(def.RawSlide, nil)
	if err != nil {
		return comfortable
	}
	density, err := slideDensity(source, slide)
	if err != nil {
		return comfortable
	}
	return density
}

func preferredDensityLabel(source *Source, density SlideDensityRecord) string {
	if source != nil && source.Tokens.Density != nil {
		return density.Requested
	}
	return ""
}

// Typed cards share their actual renderer planner. Reserve the full supported
// two-line title before reporting body space; never reuse the shorter example
// title to inflate body capacity. Padding, numbering, band and gap geometry all
// come from the pinned card row and planComponent, rather than a second model.
func typedCardAuthoringCapacities(def LibraryTemplate, source *Source, fonts authoringFontLoader) map[string]LibrarySlotCapacity {
	out := map[string]LibrarySlotCapacity{}
	if source == nil || (def.Key != "cards/3" && def.Key != "cards/4") {
		return out
	}
	slide, _, err := compileTemplateSource(source, def.Key)
	if err != nil {
		return out
	}
	frame, err := source.ResolveFrame(slide.Frame)
	if err != nil {
		return out
	}
	density := preferredTemplateDensity(def, source)
	densityLabel := preferredDensityLabel(source, density)
	typography, err := fonts()
	if err != nil {
		return out
	}
	for _, node := range slide.Nodes {
		if node.Kind != "cardrow" || node.CardRow == nil {
			continue
		}
		for i := range node.CardRow.Items {
			card := &node.CardRow.Items[i].Card
			card.Title = "Capacity\nCapacity"
			card.Body = []BodyBlock{{Key: "copy", Paragraph: "Capacity"}}
		}
		r := renderer{source: source, typeEngine: typography, bodyDensity: density.Requested, headerDensity: density.Header, densityScope: "body"}
		plan, err := r.planCardRow(node, node.Rect, frame.Body, node.Surface)
		if err != nil {
			return out
		}
		for i, child := range plan.children {
			for _, text := range child.texts {
				field := ""
				height := text.Rect.H
				lines := 2
				if strings.HasSuffix(text.ID, ".title") {
					field = "title"
				} else if strings.HasSuffix(text.ID, ".body.copy") {
					field = "body"
					height = child.record.Rect.Y + child.record.Rect.H - child.record.Padding - text.Rect.Y
					lines = shapedComponentLineBudget(typography, text.Layout.Style, text.Rect.W, height)
				}
				if field == "" || lines < 1 || height <= 0 {
					continue
				}
				style := text.Layout.Style
				capacity := LibrarySlotCapacity{Status: "estimated_geometry", Basis: "pinned_card_row_renderer_plan_with_two_line_title", Density: densityLabel, WidthPt: text.Rect.W, HeightPt: height, LineBudget: lines, Style: &style, NativeFit: "not_evaluated", Assumptions: []string{
					"Actual card row renderer geometry, padding, numbered title and band placement; fixed pinned font size.",
					"Body space reserves a full two-line title in every card; shorter titles can leave additional space.",
					"Plain text only. Go shaping is advisory; wrap boundaries and final native PowerPoint fit still require build/native review.",
				}}
				if densityLabel != "" {
					capacity.Assumptions = append(capacity.Assumptions, "Preferred authored density only; automatic density adjustment is not included in this estimate.")
				}
				capacity, err = EstimateLibrarySlotCapacity(typography, capacity)
				if err != nil {
					continue
				}
				out[fmt.Sprintf("/cards/%d/%s", i, field)] = capacity
			}
		}
	}
	return out
}

// planText allocates the greater of calibrated allocation and occupied bounds.
// Probe complete lines with that same rule instead of approximating baseline
// offsets from nominal point size. Typography uses the bundle's calibration.
func shapedComponentLineBudget(t *Typography, style Style, width, height float64) int {
	budget := 0
	for lines := 1; lines <= 128; lines++ {
		copy := strings.TrimSuffix(strings.Repeat("M\n", lines), "\n")
		layout, err := t.Measure(copy, style, width)
		if err != nil || len(layout.Lines) != lines || math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight) > height+.02 {
			break
		}
		budget = lines
	}
	return budget
}

// Plain native table cells have source-fixed column widths and row heights.
// Resolve them with the table planner so dense tokens, row-header weights and
// native insets match rendering. Rich cells/adornments remain unsupported.
func fixedSceneAuthoringCapacities(def LibraryTemplate, obj map[string]any, source *Source, fonts authoringFontLoader) map[string]LibrarySlotCapacity {
	out := map[string]LibrarySlotCapacity{}
	if source == nil {
		return out
	}
	body, _ := obj["body"].([]any)
	needed := false
	for _, raw := range body {
		node, _ := raw.(map[string]any)
		if node["type"] == "table" || node["type"] == "vstepper" {
			needed = true
			break
		}
	}
	if !needed {
		return out
	}
	density := preferredTemplateDensity(def, source)
	densityLabel := preferredDensityLabel(source, density)
	typography, err := fonts()
	if err != nil {
		return out
	}
	r := renderer{source: source, typeEngine: typography, bodyDensity: density.Requested, headerDensity: density.Header, densityScope: "body"}
	for index, raw := range body {
		node, _ := raw.(map[string]any)
		data, err := json.Marshal(node)
		if err != nil {
			continue
		}
		path := fmt.Sprintf("/body/%d", index)
		if node["type"] == "table" {
			var table sceneTableSource
			if err = sceneDecode(data, &table); err != nil || len(table.Groups) != 0 || table.RunRate != nil {
				continue
			}
			plain := true
			for i := range table.Columns {
				if table.Columns[i].Type != "" {
					plain = false
				}
				table.Columns[i].Label = "M"
			}
			for _, row := range table.Rows {
				for key, value := range row {
					if key == "total" {
						continue
					}
					var text string
					if json.Unmarshal(value, &text) != nil {
						plain = false
					}
					row[key] = json.RawMessage(`"M"`)
				}
			}
			if !plain {
				continue
			}
			plan, err := r.sceneTable("capacity", table, SceneContext{Surface: "light", Path: path})
			if err != nil {
				continue
			}
			for _, item := range plan.Items {
				if item.Table == nil {
					continue
				}
				for _, cell := range item.Table.CellRecords {
					if cell.Column < 0 || cell.Column >= len(table.Columns) {
						continue
					}
					pointer := fmt.Sprintf("%s/cols/%d/label", path, cell.Column)
					if cell.Row > 0 {
						pointer = fmt.Sprintf("%s/rows/%d/%s", path, cell.Row-1, table.Columns[cell.Column].Key)
					}
					out[pointer] = plannedAuthoringCapacity(typography, cell.Text, cell.Text.Rect.H, "pinned_plain_native_table_cell_plan", []string{"Source-fixed column widths and row heights; actual native table planner tokens, weights and insets.", "Plain string cells only. Rich cells, marks, merged groups and content-dependent padding reductions require build measurement."}, densityLabel)
				}
			}
		} else if node["type"] == "vstepper" {
			var clone map[string]any
			if json.Unmarshal(data, &clone) != nil {
				continue
			}
			steps, _ := clone["steps"].([]any)
			for _, rawStep := range steps {
				step, _ := rawStep.(map[string]any)
				step["label"] = "M\nM"
				step["title"] = "M\nM"
				step["text"] = "M"
			}
			data, err = json.Marshal(clone)
			if err != nil {
				continue
			}
			plan, _, err := r.planSequenceScene("capacity", data, SceneContext{Surface: "light", Path: path})
			if err != nil {
				continue
			}
			var sequence sequenceSpec
			if json.Unmarshal(data, &sequence) != nil {
				continue
			}
			for i, step := range sequence.Steps {
				key := step.Key
				if key == "" {
					key = fmt.Sprintf("slot-%03d", i+1)
				}
				for _, item := range plan.Items {
					if item.Text == nil || item.Text.ID != "capacity.steps."+key+".description.part-0" {
						continue
					}
					titleStyle, _ := source.StyleForDensity("body", density.Requested, "body")
					titleStyle.Weight = 600
					labelStyle, _ := source.StyleForDensity("label", density.Requested, "body")
					titleHeight, _, err := r.sequenceNeed(step.Title, titleStyle, sequence.W-36)
					if err != nil {
						continue
					}
					labelHeight, _, err := r.sequenceNeed(step.Label, labelStyle, sequence.W-36)
					if err != nil {
						continue
					}
					height := verticalSequenceStepPitch - math.Max(titleHeight, labelHeight) - 3
					out[fmt.Sprintf("%s/steps/%d/text", path, i)] = plannedAuthoringCapacity(typography, *item.Text, height, "pinned_vertical_stepper_description_plan", []string{"Actual vertical stepper renderer pitch and description width; reserves a two-line heading and label.", "Title and label share a content-dependent inline width and remain unsupported; headings longer than two lines require build measurement."}, densityLabel)
				}
			}
		}
	}
	return out
}

func plannedAuthoringCapacity(t *Typography, text TextRecord, height float64, basis string, assumptions []string, density ...string) LibrarySlotCapacity {
	style := text.Layout.Style
	lines := shapedComponentLineBudget(t, style, text.Rect.W, height)
	if lines < 1 {
		return unknownSlotCapacity("planned_text_has_no_fixed_line_budget")
	}
	level := ""
	if len(density) > 0 {
		level = density[0]
	}
	assumptions = append(assumptions, "Go shaping is advisory, not native PowerPoint fit; fixed type size and no automatic shrinking.")
	if level != "" {
		assumptions = append(assumptions, "Preferred authored density only; automatic density adjustment is not included in this estimate.")
	}
	c := LibrarySlotCapacity{Status: "estimated_geometry", Basis: basis, Density: level, WidthPt: text.Rect.W, HeightPt: height, LineBudget: lines, Style: &style, Assumptions: assumptions, NativeFit: "not_evaluated"}
	c, err := EstimateLibrarySlotCapacity(t, c)
	if err != nil {
		return unknownSlotCapacity("pinned_fonts_unavailable")
	}
	return c
}
