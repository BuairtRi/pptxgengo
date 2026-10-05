package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestHeatMinimumSharedTableBlockLegendAndV5Default(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	minimum, maximum := 1.0, 5.0
	ramp := []string{"E8EEF8", "B9C9F0", "7C9BFF", "0047FF", "070154"}
	for i, want := range ramp {
		value := float64(i + 1)
		fill, ink, err := sceneHeatDomain(value, &minimum, &maximum, "seq")
		if err != nil || fill != want || (i >= 3 && ink != "FFFFFF") {
			t.Fatal("incorrect 1–5 ramp", value, fill, ink, err)
		}
		for _, spec := range []struct{ raw, suffix string }{
			{fmt.Sprintf(`{"type":"table","x":117,"y":126,"w":180,"rowH":48,"heatMin":1,"heatMax":5,"cols":[{"k":"score","label":"Score","w":180,"type":"heat","showValue":true}],"rows":[{"score":%g}]}`, value), ".score.heat"},
			{fmt.Sprintf(`{"type":"table","x":117,"y":126,"w":180,"rowH":48,"cols":[{"k":"score","label":"Score","w":180,"type":"heat","min":1,"max":5,"showValue":true}],"rows":[{"score":%g}]}`, value), ".score.heat"},
			{fmt.Sprintf(`{"type":"block","x":117,"y":126,"w":180,"h":54,"heat":%g,"heatMin":1,"heatMax":5,"text":"Score","style":"small"}`, value), ".surface"},
			{fmt.Sprintf(`{"type":"legend","x":117,"y":126,"w":180,"items":[{"text":"Score","heat":%g,"heatMin":1,"heatMax":5}]}`, value), "slot-001.swatch"},
		} {
			plan := enhancementPlan(t, r, spec.raw)
			if got := enhancementShape(t, plan, spec.suffix).Record.Color; got != want {
				t.Fatalf("%s got%s want%s", spec.suffix, got, want)
			}
			found := false
			for _, warning := range plan.Warnings {
				if strings.Contains(warning, SceneHeatMinimumContract) && strings.Contains(warning, SceneHeatMinimumSourceSHA256) {
					found = true
				}
			}
			if !found {
				t.Fatal("minimum extension provenance absent", plan.Warnings)
			}
		}
	}
	for _, value := range []float64{-math.MaxFloat64, -1, 0, 1, 2, 3, 4, 5, math.MaxFloat64} {
		oldFill, oldInk, err := sceneHeat(value, nil, "seq")
		zero := 0.0
		newFill, newInk, newErr := sceneHeatDomain(value, &zero, nil, "seq")
		if err != nil || newErr != nil || oldFill != newFill || oldInk != newInk {
			t.Fatal("zero-min contract changed")
		}
	}
	r.source.Revision = LibraryRevisionV5
	for _, raw := range []string{
		`{"type":"block","x":117,"y":126,"w":180,"h":54,"heat":1,"heatMin":1,"heatMax":5,"text":"Score"}`,
		`{"type":"legend","x":117,"y":126,"w":180,"items":[{"text":"Score","heat":1,"heatMin":1,"heatMax":5}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"heatMin":1,"cols":[{"k":"h","label":"Score","w":180,"type":"heat"}],"rows":[{"h":1}]}`,
	} {
		if _, err := r.planSceneNode("old", json.RawMessage(raw), SceneContext{Surface: "light"}); err == nil {
			t.Fatal("v5 accepted new domain", raw)
		}
	}
}

func TestHeatReferenceBooleanEmphasisMatchesPinnedBrowser(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	plain := enhancementPlan(t, r, `{"type":"text","style":"small","x":57,"y":126,"w":180,"text":"[[A7]]"}`)
	for _, value := range []string{"true", "false"} {
		raw := json.RawMessage(fmt.Sprintf(`{"type":"text","style":"small","x":57,"y":126,"w":180,"text":"[[A7]]","emphasis":%s}`, value))
		before := string(raw)
		plan, err := r.planSceneNode("sample", raw, SceneContext{Surface: "light", Path: "/body/0"})
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != before {
			t.Fatal("caller source mutated")
		}
		if !reflect.DeepEqual(plain.Items, plan.Items) {
			t.Fatal("Boolean emphasis invented a mark or bold styling")
		}
		r.source.Revision = LibraryRevisionV5
		if _, err := r.planSceneNode("enhancement", raw, SceneContext{Surface: "light"}); err == nil {
			t.Fatal("v5 Boolean emphasis contract expanded")
		}
		r.source.Revision = LibraryRevisionV6
	}
}

func TestHeatMinimumInvalidDomainsAndFiniteExtremes(t *testing.T) {
	for _, test := range []struct{ value, min, max float64 }{{1, 1, 1}, {1, 2, 1}, {1, math.NaN(), 5}, {1, 1, math.Inf(1)}, {math.NaN(), 1, 5}} {
		if _, _, err := sceneHeatDomain(test.value, &test.min, &test.max, "seq"); err == nil {
			t.Fatal("invalid domain accepted", test)
		}
	}
	min, max := -math.MaxFloat64, math.MaxFloat64
	if fill, _, err := sceneHeatDomain(0, &min, &max, "seq"); err != nil || fill != "7C9BFF" {
		t.Fatal("finite extreme domain overflowed", fill, err)
	}
	min, max = 1, 5
	if fill, ink, err := sceneHeatDomain(5, &min, &max, "risk"); err != nil || fill != "F900D3" || ink != "070154" {
		t.Fatal("risk scale changed", fill, ink, err)
	}
}

func TestHeatMinimumPresenceRequiresV6IncludingNull(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4, LibraryRevisionV5} {
		r.source.Revision = revision
		for _, value := range []string{"null", "0", "1"} {
			for _, raw := range []string{
				fmt.Sprintf(`{"type":"block","x":117,"y":126,"w":180,"h":54,"heat":1,"heatMin":%s,"text":"Score"}`, value),
				fmt.Sprintf(`{"type":"legend","x":117,"y":126,"w":180,"items":[{"text":"Score","heat":1,"heatMin":%s}]}`, value),
			} {
				if _, err := r.planSceneNode("old", json.RawMessage(raw), SceneContext{Surface: "light"}); err == nil || !strings.Contains(err.Error(), "heat_min_requires_v6") {
					t.Fatalf("%s accepted new field %s: %v", revision, raw, err)
				}
			}
		}
	}
	r.source.Revision = LibraryRevisionV6
	for _, raw := range []string{
		`{"type":"block","x":117,"y":126,"w":180,"h":54,"heat":1,"heatMin":null,"text":"Score"}`,
		`{"type":"legend","x":117,"y":126,"w":180,"items":[{"text":"Score","heat":1,"heatMin":null}]}`,
	} {
		if _, err := r.planSceneNode("new", json.RawMessage(raw), SceneContext{Surface: "light"}); err != nil {
			t.Fatal(err)
		}
	}
}
