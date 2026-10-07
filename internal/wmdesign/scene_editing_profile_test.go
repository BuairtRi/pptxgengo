package wmdesign

import (
	"encoding/json"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNativeEditingProfilePreservesResolvedSourceStylesAndPositions(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = density
		for _, raw := range []string{
			`{"type":"bullets","x":60,"y":100,"w":340,"items":["Short bullet","Longer plain copy that should wrap into a second line without moving a separate shape.","Final bullet"]}`,
			`{"type":"bullets","x":60,"y":100,"w":340,"size":"small","items":["Short bullet","Longer plain copy that should wrap into a second line without moving a separate shape.","Final bullet"]}`,
			`{"type":"card","x":60,"y":100,"w":360,"h":230,"surface":"subtle","title":"Preserve original styles","body":[{"p":"The original padding, typography, colors and paragraph positions remain measured from the source plan."}]}`,
			`{"type":"card","x":60,"y":100,"w":360,"h":230,"surface":"subtle","pad":12,"gap":8,"titleInk":"primary","title":"Preserve original styles","body":[{"p":"The original padding, typography, colors and paragraph positions remain measured from the source plan."}]}`,
		} {
			r.editingProfile = ""
			before, err := r.planSceneNode("example", json.RawMessage(raw), SceneContext{Surface: "light"})
			if err != nil {
				t.Fatal(err)
			}
			r.editingProfile = NativeEditingProfile
			after, err := r.planSceneNode("example", json.RawMessage(raw), SceneContext{Surface: "light"})
			if err != nil {
				t.Fatal(err)
			}
			if len(after.Items) != 1 || after.Items[0].Text == nil || len(after.Groups) != 0 {
				for _, item := range before.Items {
					t.Logf("original: shape=%+v text=%+v", item.Shape, item.Text)
				}
				t.Fatal("not converted", density, raw, after.Warnings)
			}
			native := after.Items[0].Text
			next := 0
			para := 0
			for _, item := range before.Items {
				if item.Text == nil {
					continue
				}
				original := item.Text
				run := native.Rich.Paragraphs[para].Runs[0]
				if run.Style != original.Layout.Style || !reflect.DeepEqual(run.Font, original.Layout.Font) || run.Color != original.Color || run.Displayed != original.Layout.Displayed {
					t.Fatal("source style changed", density, raw, run, original)
				}
				for _, line := range original.Layout.Lines {
					actual := native.Layout.Lines[next]
					want := original.Rect.Y - native.Rect.Y + line.Baseline
					if actual.Text != line.Text || math.Abs(actual.Baseline-want) > .02 {
						t.Fatal("line position changed", density, raw)
					}
					next++
				}
				para++
			}
			if next != len(native.Layout.Lines) {
				t.Fatal("extra lines")
			}
			if strings.Contains(raw, `"card"`) && (native.NativeShape == nil || native.NativeShape.Rect != before.Items[0].Shape.Record.Rect || native.NativeShape.Fill != before.Items[0].Shape.Props.Fill.Color) {
				t.Fatal("card geometry/fill changed")
			}
		}
	}
}

func TestNativeEditingProfileLeavesComplexScenesAndDefaultsUnchanged(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		`{"type":"bullets","x":60,"y":100,"w":340,"items":[{"lead":"Lead","text":"Plain body"}]}`,
		`{"type":"bullets","x":60,"y":100,"w":340,"items":["[[Rich]] body"]}`,
		`{"type":"card","x":60,"y":100,"w":360,"h":230,"title":"Title","body":[{"p":"First body"},{"p":"Second body"}]}`,
	} {
		r.editingProfile = ""
		before, err := r.planSceneNode("example", json.RawMessage(raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		r.editingProfile = "stock"
		stock, err := r.planSceneNode("example", json.RawMessage(raw), SceneContext{Surface: "light"})
		if err != nil || !reflect.DeepEqual(before, stock) {
			t.Fatal("default changed", err)
		}
		r.editingProfile = NativeEditingProfile
		after, err := r.planSceneNode("example", json.RawMessage(raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before.Items, after.Items) || !reflect.DeepEqual(before.Groups, after.Groups) || len(after.Warnings) <= len(before.Warnings) {
			t.Fatal("complex scene changed", raw)
		}
	}
	if ValidateEditingProfile("unknown") == nil {
		t.Fatal("unknown profile accepted")
	}
}

func TestNativeEditingProfileRetainsBorderedCardsAndReservedFooter(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		raw  string
		zone Rect
	}{
		{`{"type":"card","x":60,"y":100,"w":360,"h":230,"state":"deemph","title":"Title","body":[{"p":"Plain copy"}]}`, Rect{}},
		{`{"type":"card","x":60,"y":100,"w":360,"h":230,"state":"placeholder","title":"Title","body":[{"p":"Plain copy"}]}`, Rect{}},
		{`{"type":"card","x":60,"y":100,"w":360,"h":300,"title":"Title","body":[{"p":"Plain copy"}]}`, Rect{X: 60, Y: 100, W: 360, H: 150}},
	} {
		ctx := SceneContext{Surface: "light", Zone: tc.zone}
		r.editingProfile = ""
		before, err := r.planSceneNode("example", json.RawMessage(tc.raw), ctx)
		if err != nil {
			t.Fatal(err)
		}
		r.editingProfile = NativeEditingProfile
		after, err := r.planSceneNode("example", json.RawMessage(tc.raw), ctx)
		if err != nil {
			t.Fatal("valid source now fails", err)
		}
		if !reflect.DeepEqual(before.Items, after.Items) || !reflect.DeepEqual(before.Groups, after.Groups) || len(after.Warnings) <= len(before.Warnings) {
			t.Fatal("border or footer compatibility changed", tc.raw)
		}
	}
}

func TestNativeEditingProfileTableAndCardRow(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = density
		raw := json.RawMessage(`{"type":"table","x":57,"y":198,"w":414,"cols":[{"k":"step","label":"Step","w":207},{"k":"owner","label":"Owner","w":207}],"rows":[{"step":"Review","owner":"Reviewer"},{"step":"Build","owner":"Author"}]}`)
		r.editingProfile = ""
		before, err := r.planSceneNode("table", raw, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		r.editingProfile = NativeEditingProfile
		after, err := r.planSceneNode("table", raw, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before.Items, after.Items) || len(after.Groups) != 0 {
			t.Fatal("table changed beyond wrapper", density)
		}
		row := json.RawMessage(`{"type":"cardrow","x":60,"y":100,"w":300,"h":200,"gap":18,"card":{"surface":"subtle"},"items":[{"title":"First","body":[{"p":"Plain body one"}]},{"title":"Second","body":[{"p":"Plain body two"}]}]}`)
		p, err := r.planSceneNode("row", row, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Items) != 2 || len(p.Groups) != 1 {
			t.Fatal("row ownership changed", density)
		}
		for _, item := range p.Items {
			if item.Text == nil || item.Text.NativeShape == nil {
				t.Fatal("eligible child not converted", density)
			}
		}
	}
}
