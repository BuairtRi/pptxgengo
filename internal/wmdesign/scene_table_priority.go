package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// scenePriorityChip describes the source priority value and the editable chip
// treatment used by native table renderers. Geometry is in points.
type scenePriorityChip struct {
	Value      string
	Label      string
	Background string
	Foreground string
	Border     string
	PaddingX   float64
	PaddingY   float64
	MinWidth   float64
	BorderPt   float64
	Known      bool
}

const (
	scenePriorityNavy    = "#070154"
	scenePriorityWhite   = "#FFFFFF"
	scenePriorityMidNavy = "#32477B"
	scenePriorityPale    = "#CED7E6"
)

// sceneTablePriority accepts a string or {value,label?}. It preserves the
// exact source value separately from the display label and resolves the
// upstream palette case-insensitively. Unrecognized values keep their authored
// label and receive the upstream outline fallback.
func sceneTablePriority(raw json.RawMessage) (scenePriorityChip, error) {
	var value, label string
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return scenePriorityChip{}, fmt.Errorf("scene.table_priority_requires_value")
	}
	if trimmed[0] == '"' {
		if err := json.Unmarshal(trimmed, &value); err != nil {
			return scenePriorityChip{}, fmt.Errorf("scene.table_priority_requires_string: %w", err)
		}
	} else if trimmed[0] == '{' {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &fields); err != nil {
			return scenePriorityChip{}, fmt.Errorf("scene.table_priority_invalid_object: %w", err)
		}
		for field := range fields {
			if field != "value" && field != "label" {
				return scenePriorityChip{}, fmt.Errorf("scene.table_priority_unknown_field: %s", field)
			}
		}
		if rawValue, ok := fields["value"]; !ok || json.Unmarshal(rawValue, &value) != nil {
			return scenePriorityChip{}, fmt.Errorf("scene.table_priority_requires_string_value")
		}
		if rawLabel, ok := fields["label"]; ok && !bytes.Equal(bytes.TrimSpace(rawLabel), []byte("null")) {
			if err := json.Unmarshal(rawLabel, &label); err != nil {
				return scenePriorityChip{}, fmt.Errorf("scene.table_priority_requires_string_label: %w", err)
			}
		}
	} else {
		return scenePriorityChip{}, fmt.Errorf("scene.table_priority_requires_string_or_object")
	}
	if strings.TrimSpace(value) == "" {
		return scenePriorityChip{}, fmt.Errorf("scene.table_priority_requires_value")
	}
	chip := scenePriorityChip{Value: value, Label: label, Background: "transparent", Foreground: scenePriorityNavy, Border: scenePriorityNavy, PaddingX: 8, PaddingY: 2, MinWidth: 40, BorderPt: .75}
	key := strings.ToLower(strings.TrimSpace(value))
	styles := map[string]struct {
		label, background, foreground string
	}{
		"p1":    {"P1", scenePriorityNavy, scenePriorityWhite},
		"p2":    {"P2", scenePriorityMidNavy, scenePriorityWhite},
		"p3":    {"P3", scenePriorityPale, scenePriorityNavy},
		"p4":    {"P4", "transparent", scenePriorityNavy},
		"now":   {"Now", scenePriorityNavy, scenePriorityWhite},
		"next":  {"Next", scenePriorityPale, scenePriorityNavy},
		"later": {"Later", "transparent", scenePriorityNavy},
	}
	if style, ok := styles[key]; ok {
		chip.Known = true
		chip.Background = style.background
		chip.Foreground = style.foreground
		if chip.Label == "" {
			chip.Label = style.label
		}
	} else if chip.Label == "" {
		chip.Label = value
	}
	return chip, nil
}

// WidthForLabel combines a measured text width with the upstream horizontal
// padding and minimum chip width. The caller supplies a typeEngine measurement.
func (chip scenePriorityChip) WidthForLabel(measuredTextWidth float64) (float64, error) {
	if measuredTextWidth < 0 || math.IsNaN(measuredTextWidth) || math.IsInf(measuredTextWidth, 0) {
		return 0, fmt.Errorf("scene.table_priority_invalid_text_width")
	}
	width := measuredTextWidth + 2*chip.PaddingX
	if width < chip.MinWidth {
		width = chip.MinWidth
	}
	return width, nil
}
