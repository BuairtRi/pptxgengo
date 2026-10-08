package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestNativeProfileListConvertsMeasuredLeadBodyRuns(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"type":"bullets","x":60,"y":100,"w":340,"items":[{"lead":"Review first","text":"Confirm each exception has an owner."},{"lead":"Route quickly","text":"Escalate unresolved items."}]}`)
	before, err := r.planSceneNode("lead-list", raw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	native, reason := r.nativeProfileListResult(before, raw)
	if native == nil {
		t.Fatalf("lead/body rows not converted: %s", reason)
	}
	text := native.Items[0].Text
	if text.NativeParagraphContract != EditableListContract || len(text.Rich.Paragraphs) != 2 {
		t.Fatal("native list contract/paragraph count not set")
	}
	for _, para := range text.Rich.Paragraphs {
		if !para.Bullet || len(para.Runs) != 2 || para.Runs[0].Style.Weight != 600 {
			t.Fatalf("lead/body runs were not preserved: %+v", para)
		}
	}
}

func TestNativeProfileListConvertsSceneDataBulletBlocksInPlace(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	items := []json.RawMessage{json.RawMessage(`{"lead":"Review first","text":"Confirm the exception owner."}`), json.RawMessage(`"Route unresolved items."`)}
	for _, density := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = density
		style, err := r.sceneStyle("body")
		if err != nil {
			t.Fatal(err)
		}
		build := func(profile string) (*scenePlan, float64) {
			r.editingProfile = profile
			prior := ComponentRecord{ID: "preexisting", Definition: "sentinel", Parts: []string{"kept"}}
			p := &scenePlan{ID: "container", Bounds: Rect{X: 40, Y: 70, W: 10, H: 10}, Items: []sceneItem{{Shape: &sceneShape{Record: ShapeRecord{ID: "kept", Rect: Rect{X: 40, Y: 70, W: 10, H: 10}}}}}, Groups: []ComponentRecord{prior}}
			end, err := r.sceneDataBullets(p, "container.body.bullets", items, SceneContext{Surface: "light"}, "body/0/bullets", Rect{X: 60, Y: 100, W: 340}, "light", style)
			if err != nil {
				t.Fatal(err)
			}
			return p, end
		}
		stock, stockEnd := build("")
		native, nativeEnd := build(NativeEditingProfile)
		if nativeEnd != stockEnd || native.Bounds != stock.Bounds {
			t.Fatalf("%s list conversion changed flow geometry: end %g/%g bounds %+v/%+v", density, stockEnd, nativeEnd, stock.Bounds, native.Bounds)
		}
		if len(native.Items) != 2 || native.Items[1].Text == nil || len(native.Items[1].Text.Rich.Paragraphs) != len(items) {
			t.Fatalf("%s body bullet block not converted to one text object: %+v", density, native.Items)
		}
		if len(native.Groups) != 1 || !reflect.DeepEqual(native.Groups[0], stock.Groups[0]) {
			t.Fatalf("%s preexisting group reference changed: %+v", density, native.Groups)
		}
		sceneDataGroup(native, "outer", "body", 0, native.Bounds)
		parts := map[string]bool{}
		for _, part := range native.Groups[1].Parts {
			parts[part] = true
		}
		if len(native.Groups[1].Parts) != 2 || !parts["container.body.bullets"] || !parts["preexisting"] {
			t.Fatalf("%s enclosing group did not reference the combined bullet object: %+v", density, native.Groups[1].Parts)
		}
	}
}

func TestNativeProfileListRichParagraphXMLPreservesSourceFormatting(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"type":"bullets","x":60,"y":100,"w":340,"items":["A [[measured]] phrase without literal markers.",{"lead":"Route quickly","text":"Escalate unresolved items."}]}`)
	for _, density := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = density
		p, err := r.planSceneNode("xml-list", raw, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		native, reason := r.nativeProfileListResult(p, raw)
		if native == nil {
			t.Fatalf("%s list was retained: %s", density, reason)
		}
		text := native.Items[0].Text
		for i, para := range text.Rich.Paragraphs {
			marker := p.Items[i*2].Shape
			if para.BulletColor != marker.Props.Fill.Color {
				t.Fatalf("%s marker color changed: %q != %q", density, para.BulletColor, marker.Props.Fill.Color)
			}
			xml := string(editableListParagraphXML(*text))
			if !strings.Contains(xml, `<a:buClr><a:srgbClr val="`+para.BulletColor+`"/></a:buClr>`) ||
				!strings.Contains(xml, `typeface="Wingdings" charset="2"`) ||
				!strings.Contains(xml, `typeface="`+para.Runs[0].Font.Typeface+`"`) ||
				!strings.Contains(xml, fmt.Sprintf(`marL="%d"`, int(math.Round(para.BulletIndentPt*12700)))) ||
				!strings.Contains(xml, fmt.Sprintf(`<a:spcAft><a:spcPts val="%s"`, strconv.Itoa(int(math.Round(para.ParagraphGapAfter*100))))) {
				t.Fatalf("%s paragraph XML lost native style, marker color, indent or gap: %s", density, xml)
			}
		}
	}
}

func TestNativeProfileListReportsUnsupportedStructure(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"type":"bullets","x":60,"y":100,"w":340,"items":[{"lead":"Lead","text":"Body","sub":["Nested"]}]}`)
	p, err := r.planSceneNode("nested-list", raw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	_, reason := r.nativeProfileListResult(p, raw)
	if reason != "nested bullet markers use source-specific outline geometry and cannot be represented by the solid native paragraph marker" {
		t.Fatalf("unexpected fallback reason: %q", reason)
	}
}

func TestNativeProfileListConvertsPlannerResolvedInlineMarkup(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"type":"bullets","x":60,"y":100,"w":340,"items":["A [[measured]] phrase without literal markers."]}`)
	p, err := r.planSceneNode("marked-list", raw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	native, reason := r.nativeProfileListResult(p, raw)
	if native == nil {
		t.Fatalf("planner-resolved markup was retained: %s", reason)
	}
	paragraph := native.Items[0].Text.Rich.Paragraphs[0]
	if len(paragraph.Runs) < 2 || strings.Contains(paragraph.Displayed, "[[") || strings.Contains(paragraph.Displayed, "]]") {
		t.Fatalf("inline source syntax was not emitted as measured rich runs: %+v", paragraph)
	}
}

func TestNativeProfileListDisplayedMetadataMatchesParagraphText(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	r.editingProfile = NativeEditingProfile
	p, err := r.planSceneNode("styled-list", json.RawMessage(`{"type":"bullets","x":60,"y":100,"w":340,"items":["[[Formatted]] text","Next item"]}`), SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 1 || p.Items[0].Text == nil {
		t.Fatal("styled list did not convert", p.Warnings)
	}
	tr := p.Items[0].Text
	var display []string
	for _, para := range tr.Rich.Paragraphs {
		display = append(display, para.Displayed)
	}
	if tr.Layout.Displayed != strings.Join(display, "\n") || strings.Contains(tr.Layout.Displayed, "[[") {
		t.Fatal("displayed metadata includes authored syntax", tr.Layout.Displayed)
	}
}
