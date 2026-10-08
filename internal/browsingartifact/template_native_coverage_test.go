package browsingartifact

import (
	"encoding/json"
	"testing"
)

func TestValidateNativeTemplateCoverageRequiresExactDeckMappingAndTotals(t *testing.T) {
	deck := TemplateCoverage{ExpectedTemplates: 2, Entries: []Entry{{Kind: "template", Key: "cards/3", SlideID: "template-0001"}, {Kind: "template", Key: "lists/4", SlideID: "template-0002"}, {Kind: "frame", Key: "frame/one", SlideID: "frame-0001"}}}
	valid := map[string]any{"schema": "pptxgengo.native-template-coverage.v1", "editing_profile": "native-v1", "templates": 2, "native_list_boxes": 3, "native_card_shapes": 2, "native_tables": 1, "entries": []any{map[string]any{"template": "cards/3", "slide_id": "template-0001", "native_list_boxes": 2, "native_card_shapes": 0, "native_tables": 0}, map[string]any{"template": "lists/4", "slide_id": "template-0002", "native_list_boxes": 1, "native_card_shapes": 2, "native_tables": 1}}}
	encode := func() []byte { raw, _ := json.Marshal(valid); return raw }
	if err := validateNativeTemplateCoverage(encode(), "native-v1", deck); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		change func(map[string]any)
	}{
		{"missing template", func(m map[string]any) {
			m["entries"] = []any{map[string]any{"template": "cards/3", "slide_id": "template-0001", "native_list_boxes": 2, "native_card_shapes": 0, "native_tables": 0}}
		}},
		{"stale profile", func(m map[string]any) { m["editing_profile"] = "source-v0" }},
		{"invalid sums", func(m map[string]any) { m["native_tables"] = 0 }},
		{"duplicate template identity", func(m map[string]any) { rows := m["entries"].([]any); rows[1].(map[string]any)["template"] = "cards/3" }},
		{"negative count", func(m map[string]any) {
			rows := m["entries"].([]any)
			rows[0].(map[string]any)["native_list_boxes"] = -1
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			copy := map[string]any{}
			for k, v := range valid {
				copy[k] = v
			}
			rows := make([]any, len(valid["entries"].([]any)))
			for i, row := range valid["entries"].([]any) {
				fields := map[string]any{}
				for k, v := range row.(map[string]any) {
					fields[k] = v
				}
				rows[i] = fields
			}
			copy["entries"] = rows
			tc.change(copy)
			raw, _ := json.Marshal(copy)
			if err := validateNativeTemplateCoverage(raw, "native-v1", deck); err == nil {
				t.Fatal("invalid native coverage accepted")
			}
		})
	}
}
