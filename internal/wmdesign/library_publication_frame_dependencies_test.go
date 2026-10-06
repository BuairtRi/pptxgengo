package wmdesign

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPublicationFrameDocumentationPreservesRenderingDependencies(t *testing.T) {
	previous, err := Load(v7IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	current, err := Load(v9IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), current.Frames.Features[0].Geometry...)
	if reflect.DeepEqual(previous.Frames, current.Frames) {
		t.Fatal("fixture lost documented geometry additions")
	}
	if !compatiblePublicationStyle(previous, current) {
		t.Fatal("redundant frame documentation blocked unchanged rendering")
	}
	if !reflect.DeepEqual(before, []byte(current.Frames.Features[0].Geometry)) {
		t.Fatal("normalization mutated source geometry")
	}
	for _, tc := range []struct {
		name, id string
		change   func(map[string]any)
	}{
		{"real title geometry", "title-zone", func(g map[string]any) { g["oneLine"].(map[string]any)["rule"] = 109.0 }},
		{"real source geometry", "zone.source", func(g map[string]any) { g["lineHeight"] = 13.0 }},
		{"incorrect documented fallback", "title-zone", func(g map[string]any) { g["threeLine"].(map[string]any)["rule"] = 181.0 }},
		{"incorrect documented budget", "zone.source", func(g map[string]any) { g["bodyBottomTable"].(map[string]any)["slim"].([]any)[0] = 487.0 }},
		{"unknown geometry", "title-zone", func(g map[string]any) { g["newLayout"] = 1.0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := *current
			changed.Frames.Features = append(current.Frames.Features[:0:0], current.Frames.Features...)
			for i := range changed.Frames.Features {
				feature := &changed.Frames.Features[i]
				if feature.ID != tc.id {
					continue
				}
				var geometry map[string]any
				if err := json.Unmarshal(feature.Geometry, &geometry); err != nil {
					t.Fatal(err)
				}
				tc.change(geometry)
				feature.Geometry, err = json.Marshal(geometry)
				if err != nil {
					t.Fatal(err)
				}
			}
			if compatiblePublicationStyle(previous, &changed) {
				t.Fatal("changed rendering dependency accepted")
			}
		})
	}
}
