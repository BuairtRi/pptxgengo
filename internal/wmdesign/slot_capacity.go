package wmdesign

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// ApproxCharacters is a drafting guide for a stated Latin prose sample, never
// a maximum. Line and height limits are geometry-derived; actual copy is shaped.
type LibrarySlotCapacity struct {
	Status           string   `json:"status"`
	Basis            string   `json:"basis"`
	Density          string   `json:"density,omitempty"`
	WidthPt          float64  `json:"width_pt,omitempty"`
	HeightPt         float64  `json:"height_pt,omitempty"`
	LineBudget       int      `json:"line_budget,omitempty"`
	ApproxCharacters int      `json:"approx_characters,omitempty"`
	Style            *Style   `json:"style,omitempty"`
	Assumptions      []string `json:"assumptions"`
	NativeFit        string   `json:"native_fit"`
}

func unknownSlotCapacity(basis string) LibrarySlotCapacity {
	return LibrarySlotCapacity{Status: "unsupported", Basis: basis, Assumptions: []string{"Component internal layout requires a built slide measurement."}, NativeFit: "not_evaluated"}
}
func authoringCapacity(def LibraryTemplate, obj, node map[string]any, slot LibraryAuthoringSlot, source *Source) LibrarySlotCapacity {
	if slot.Kind != "string" {
		return unknownSlotCapacity("non_text_scalar")
	}
	if slot.Classification == "decorative" {
		return unknownSlotCapacity("decorative_no_authoring_budget")
	}
	if source == nil {
		return unknownSlotCapacity("pinned_style_not_available")
	}
	c := LibrarySlotCapacity{Status: "estimated_geometry", Basis: "source_geometry_and_pinned_type_style", NativeFit: "not_evaluated", Assumptions: []string{"Fixed type size; no automatic shrinking; Go shaping is advisory, not native PowerPoint fit."}}
	token := ""
	explicitLines := 0
	if slot.SourcePointer == "/title" || slot.SourcePointer == "/eyebrow" || slot.SourcePointer == "/source/text" {
		slide, err := compileLibrarySlide(def.RawSlide, nil)
		if err != nil {
			return unknownSlotCapacity("frame_not_resolved")
		}
		f, err := source.ResolveFrame(slide.Frame)
		if err != nil {
			return unknownSlotCapacity("frame_not_resolved")
		}
		switch slot.SourcePointer {
		case "/title":
			c.WidthPt = f.Header.W
			c.HeightPt = f.TitleRule - 54
			token = f.TitleStyle
			explicitLines = f.Request.TitleLines
		case "/eyebrow":
			c.WidthPt = f.Header.W
			c.HeightPt = 12
			token = "eyebrow"
			explicitLines = 1
			if obj["stamp"] != "" && obj["stamp"] != nil {
				c.WidthPt -= 180
			}
		case "/source/text":
			c.WidthPt = f.Source.W
			c.HeightPt = f.Source.H
			token = "source"
			explicitLines = f.Request.SourceLines
		}
	} else if node != nil && (node["type"] == "text" || node["type"] == "block") && strings.HasSuffix(slot.SourcePointer, "/text") {
		token, _ = node["style"].(string)
		if token == "" && node["type"] == "block" {
			token = "body"
		}
		c.WidthPt, _ = discoveryNumber(node["w"])
		c.HeightPt, _ = discoveryNumber(node["h"])
		if node["type"] == "block" {
			pad := 12.0
			if (token == "number" && c.WidthPt <= 54) || (token == "small" && c.WidthPt <= 72) || ((token == "label" || token == "small") && (c.HeightPt <= 18 || c.WidthPt <= 54)) {
				pad = math.Min(6, c.WidthPt/6)
			}
			c.WidthPt -= 2 * pad
			c.Assumptions = append(c.Assumptions, "Conservative fixed block inset; content dependent native padding is not modeled.")
		}
		if c.HeightPt <= 0 {
			x, _ := discoveryNumber(node["x"])
			y, _ := discoveryNumber(node["y"])
			width, _ := discoveryNumber(node["w"])
			bottom := 486.0
			if slide, e := compileLibrarySlide(def.RawSlide, nil); e == nil {
				if f, e := source.ResolveFrame(slide.Frame); e == nil {
					bottom = f.Body.Y + f.Body.H
				}
			}
			body, _ := obj["body"].([]any)
			for _, raw := range body {
				next, _ := raw.(map[string]any)
				ny, yok := discoveryNumber(next["y"])
				nx, xok := discoveryNumber(next["x"])
				nw, wok := discoveryNumber(next["w"])
				if !yok || !xok || !wok || ny <= y+.1 || nx >= x+width || nx+nw <= x {
					continue
				}
				if next["type"] == "block" && next["text"] == "" {
					continue
				}
				bottom = math.Min(bottom, ny)
			}
			c.HeightPt = math.Max(0, bottom-y)
			c.Basis = "pinned_type_and_inferred_sibling_clearance"
			c.Assumptions = append(c.Assumptions, "Height extends to the next overlapping source content or body bottom; this is a drafting clearance, not a renderer constraint.")
		}
	} else {
		return unknownSlotCapacity("component_internal_layout_not_modeled")
	}
	density := preferredTemplateDensity(def, source)
	scope := "body"
	if slot.SourcePointer == "/title" || slot.SourcePointer == "/eyebrow" {
		scope = "header"
	}
	st, e := source.StyleForDensity(token, density.Requested, scope)
	if e != nil || c.WidthPt <= 0 || c.HeightPt <= 0 {
		return unknownSlotCapacity("no_fixed_text_geometry")
	}
	if weight, ok := discoveryNumber(node["weight"]); ok && weight > 0 {
		st.Weight = int(weight)
	}
	if source.Tokens.Density != nil {
		c.Density = density.Requested
		if scope == "header" {
			c.Density = density.Header
			c.Assumptions = append(c.Assumptions, "Capacity uses the preferred header density; automatic fit adjustment is not included.")
		} else {
			c.Assumptions = append(c.Assumptions, "Capacity uses the template's preferred authored body density; automatic fit adjustment is not included.")
		}
	}
	c.Style = &st
	c.LineBudget = int(math.Floor((c.HeightPt-1.2*st.Size)/st.Leading)) + 1
	if explicitLines > 0 {
		c.LineBudget = explicitLines
	}
	if c.LineBudget < 1 {
		c.LineBudget = 1
	}
	return c
}

// EstimateLibrarySlotCapacity shapes a declared representative Latin prose
// sample using actual fonts. Long words, capitalization, markup and explicit
// breaks can change capacity substantially; supplied text must be measured.
func EstimateLibrarySlotCapacity(t *Typography, c LibrarySlotCapacity) (LibrarySlotCapacity, error) {
	if c.Style == nil || c.WidthPt <= 0 || c.LineBudget <= 0 {
		return c, nil
	}
	const sample = "The team reviews product priorities and builds useful services. "
	layout, e := t.Measure(sample, *c.Style, 960)
	if e != nil {
		return c, e
	}
	advance := 0.0
	for _, line := range layout.Lines {
		advance += line.Advance
	}
	if advance <= 0 {
		return c, fmt.Errorf("capacity.empty_probe")
	}
	c.ApproxCharacters = int(math.Floor(c.WidthPt/(advance/float64(utf8.RuneCountInString(sample))))) * c.LineBudget
	c.Status = "estimated_go_shaping"
	c.Assumptions = append(c.Assumptions, "Approximate characters use a mixed case Latin prose probe including spaces; this is not a maximum character count.")
	return c, nil
}

// RestyleLibrarySlotCapacityAtDensity recalculates a static preferred-tier
// estimate for an explicit per-slide override. It never tries alternate tiers
// or claims that supplied copy fits natively.
func RestyleLibrarySlotCapacityAtDensity(source *Source, typography *Typography, capacity LibrarySlotCapacity, level, scope string, maxLines int) (LibrarySlotCapacity, error) {
	if capacity.Style == nil || source == nil || typography == nil || capacity.WidthPt <= 0 || capacity.HeightPt <= 0 {
		return unsupportedDensityCapacity(capacity, "density_capacity_missing_geometry_or_style"), nil
	}
	style, err := source.StyleForDensity(capacity.Style.ID, level, scope)
	if err != nil {
		return unsupportedDensityCapacity(capacity, "density_capacity_style_unavailable"), nil
	}
	// Density changes size and leading. Preserve source-literal weight and
	// decorative settings that aren't represented by the role token.
	style.Weight = capacity.Style.Weight
	style.Italic = capacity.Style.Italic
	style.Tracking = capacity.Style.Tracking
	style.TrackingPt = capacity.Style.TrackingPt
	style.Case = capacity.Style.Case
	lines := shapedComponentLineBudget(typography, style, capacity.WidthPt, capacity.HeightPt)
	if maxLines > 0 && lines > maxLines {
		lines = maxLines
	}
	if lines < 1 {
		return unsupportedDensityCapacity(capacity, "density_capacity_no_preferred_tier_line"), nil
	}
	capacity.Style = &style
	capacity.LineBudget = lines
	capacity.Density = level
	capacity.ApproxCharacters = 0
	capacity.Status = "estimated_geometry"
	capacity.Assumptions = append(capacity.Assumptions, "Uses the explicit per-slide preferred density; automatic density adjustment and native fit are not included.")
	return EstimateLibrarySlotCapacity(typography, capacity)
}

func unsupportedDensityCapacity(capacity LibrarySlotCapacity, basis string) LibrarySlotCapacity {
	capacity.Status = "unsupported"
	capacity.Basis = basis
	capacity.LineBudget = 0
	capacity.ApproxCharacters = 0
	capacity.Style = nil
	capacity.Density = ""
	capacity.Assumptions = append(capacity.Assumptions, "Capacity was not recalculated for the per-slide density override.")
	return capacity
}

type LibrarySlotFit struct {
	Alias           string              `json:"alias"`
	Capacity        LibrarySlotCapacity `json:"capacity"`
	Status          string              `json:"status"`
	Lines           int                 `json:"line_count"`
	LineBudget      int                 `json:"line_budget"`
	LineOverrun     int                 `json:"line_overrun"`
	HeightOverrunPt float64             `json:"height_overrun_pt"`
	WidthOverrunPt  float64             `json:"width_overrun_pt"`
	NativeFit       string              `json:"native_fit"`
	Assumptions     []string            `json:"assumptions"`
}

func MeasureLibrarySlot(t *Typography, slot LibraryAuthoringSlot, text string) (LibrarySlotFit, error) {
	c := slot.Capacity
	fit := LibrarySlotFit{Alias: slot.Alias, Capacity: c, Status: "unsupported", LineBudget: c.LineBudget, NativeFit: "not_evaluated", Assumptions: c.Assumptions}
	if c.Style == nil || c.WidthPt <= 0 || c.HeightPt <= 0 {
		return fit, nil
	}
	if strings.Contains(text, "[[") || strings.Contains(text, "[^") {
		fit.Assumptions = append(fit.Assumptions, "Rich markup requires component layout measurement.")
		return fit, nil
	}
	layout, e := t.Measure(text, *c.Style, c.WidthPt)
	if e != nil {
		return fit, e
	}
	fit.Lines = len(layout.Lines)
	fit.LineOverrun = max(0, fit.Lines-c.LineBudget)
	fit.HeightOverrunPt = math.Max(0, layout.EstimatedOccupiedHeight-c.HeightPt)
	for _, line := range layout.Lines {
		fit.WidthOverrunPt = math.Max(fit.WidthOverrunPt, line.Advance-c.WidthPt)
	}
	fit.Status = "fits_estimate"
	if fit.LineOverrun > 0 || fit.HeightOverrunPt > .02 || fit.WidthOverrunPt > .02 {
		fit.Status = "overflow_estimate"
	}
	return fit, nil
}
