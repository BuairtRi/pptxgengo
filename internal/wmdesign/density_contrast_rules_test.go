package wmdesign

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDensityContrastSourceRulesAndHistoricalIsolation(t *testing.T) {
	source := densityTestSource(t)
	raw, err := os.ReadFile(filepath.Join(source.Root, "explorations/components.src.html"))
	if err != nil {
		t.Fatal(err)
	}
	if source.densityContrastRules != recognizesDensityContrastRules(raw) {
		t.Fatal("source capability differs from verified executable source")
	}
	if !source.densityContrastRules {
		t.Fatal("current source lacks the approved contrast renderer contract")
	}
	left := `styled("label", roles.emphasis, n.inflectionLabel || "Inflection", bg)`
	right := `grey ? { ring: "#97A4BA", ink: "#070154" }`
	if recognizesDensityContrastRules([]byte(left)) || recognizesDensityContrastRules([]byte(right)) || !recognizesDensityContrastRules([]byte(left+"\n"+right)) {
		t.Fatal("contrast contract requires both executable signatures")
	}
	old, err := Load(v10IntakeBundle(), "")
	if err != nil || old.densityContrastRules {
		t.Fatalf("legacy source changed: %v", err)
	}
	doc, err := LibrarySourceReference(densityTestBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	wants := map[string]int{"maturity/ai-beyond": 1, "maturity/insights": 1, "road-fork/decision-chosen": 2}
	seen := 0
	for _, slide := range doc.Slides {
		count, ok := wants[slide.TemplateBinding.Template]
		if !ok {
			continue
		}
		seen++
		for _, level := range []string{"comfortable", "compact", "dense"} {
			r := tileTestRenderer(t, level)
			copySource := *r.source
			r.source = &copySource
			r.contrastProbe = &contrastProbe{}
			frame, err := r.source.ResolveFrame(slide.Frame)
			if err != nil {
				t.Fatal(err)
			}
			node := slide.Nodes[0]
			ctx := SceneContext{Surface: frame.Request.Surface, Zone: frame.Body, Path: node.Scene.Path, Keys: node.Scene.Keys}
			copySource.densityContrastRules = false
			before, err := r.planSceneNode(node.ID, node.Scene.Node, ctx)
			if err != nil {
				t.Fatal(err)
			}
			copySource.densityContrastRules = true
			after, err := r.planSceneNode(node.ID, node.Scene.Node, ctx)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			for _, item := range before.Items {
				if item.Text == nil {
					continue
				}
				text := item.Text
				if strings.HasSuffix(text.ID, ".inflection") {
					if text.Color != "F900D3" {
						t.Fatal("legacy inflection ink changed")
					}
					text.Color = "0047FF"
					changed++
				} else if text.Color == "97A4BA" && strings.HasSuffix(text.ID, ".number") {
					text.Color = "070154"
					changed++
				}
			}
			if changed != count || !reflect.DeepEqual(before, after) {
				t.Fatalf("%s/%s: color-only contract drift; changed%d want%d", slide.TemplateBinding.Template, level, changed, count)
			}
		}
	}
	if seen != len(wants) {
		t.Fatal("missing approved source specimens")
	}
}
