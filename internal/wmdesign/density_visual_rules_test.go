package wmdesign

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDensityVisualSourceCapabilityAndHistoricalIsolation(t *testing.T) {
	s := densityTestSource(t)
	raw, err := os.ReadFile(filepath.Join(s.Root, "explorations/components.src.html"))
	if err != nil {
		t.Fatal(err)
	}
	if s.densityVisualRules != recognizesDensityVisualRules(raw) {
		t.Fatal("capability not derived from frozen source")
	}
	old, err := Load(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if old.densityVisualRules {
		t.Fatal("old source acquired new renderer semantics")
	}
	// A partial renderer edit never acquires the full visual contract.
	if recognizesDensityVisualRules([]byte(`var CALLOUT_WHITE = { display: 1, title: 1, heading: 1, stat: 1, "stat-sm": 1 };`)) {
		t.Fatal("partial source signature accepted")
	}
	before, _ := json.Marshal(s)
	copySource := *s
	copySource.densityVisualRules = true
	for _, level := range []string{"comfortable", "compact", "dense"} {
		r := renderer{source: &copySource, bodyDensity: level}
		for _, tc := range []struct{ surface, token, ink string }{{"callout", "body", "primary"}, {"callout", "small", "primary"}, {"callout", "subhead", "primary"}, {"callout", "number", "primary"}, {"callout", "heading", "display"}, {"callout", "stat-sm", "display"}, {"inverse", "small", "primary"}, {"inverse", "subhead", "display"}, {"light", "label", "primary"}} {
			if got := r.shapeTextInk(tc.surface, tc.token); got != tc.ink {
				t.Fatalf("%s/%s/%s=%s", level, tc.surface, tc.token, got)
			}
		}
		for _, tc := range []struct{ surface, ink string }{{"callout", "primary"}, {"strong", "emphasis"}, {"light", "emphasis"}, {"inverse", "#F900D3"}, {"deep", "#F900D3"}} {
			if r.chevronNumberInk(tc.surface) != tc.ink {
				t.Fatal("chevron source emphasis bypass")
			}
		}
	}
	copySource.densityVisualRules = false
	r := renderer{source: &copySource, bodyDensity: "dense"}
	if r.shapeTextInk("callout", "body") != "display" || r.chevronNumberInk("callout") != "emphasis" {
		t.Fatal("historical default changed")
	}
	after, _ := json.Marshal(s)
	if string(before) != string(after) {
		t.Fatal("renderer mutated cached source")
	}
}
func TestDensityQuoteScaleAndFixedLegacy(t *testing.T) {
	s := densityTestSource(t)
	copySource := *s
	copySource.densityVisualRules = true
	for i, level := range []string{"comfortable", "compact", "dense"} {
		r := renderer{source: &copySource, bodyDensity: level}
		card, err := r.quoteMarkStyle(true)
		if err != nil {
			t.Fatal(err)
		}
		pull, err := r.quoteMarkStyle(false)
		if err != nil {
			t.Fatal(err)
		}
		if card.Size != []float64{48, 16 * (48. / 18.), 40}[i] || card.Leading != []float64{36, 32, 30}[i] || pull.Size != []float64{60, 52.5, 45}[i] || pull.Leading != []float64{30, 26.25, 22.5}[i] {
			t.Fatalf("%s card%+v pull%+v", level, card, pull)
		}
		copySource.densityVisualRules = false
		legacyCard, _ := r.quoteMarkStyle(true)
		legacyPull, _ := r.quoteMarkStyle(false)
		if legacyCard.Size != 48 || legacyCard.Leading != 36 || legacyPull.Size != 60 || legacyPull.Leading != 30 {
			t.Fatal("historical quote changed")
		}
		copySource.densityVisualRules = true
	}
}
func TestDensitySourceLimitValidationAndBoundProtection(t *testing.T) {
	s := densityTestSource(t)
	for _, limit := range []string{"standard", "appendix", "unknown"} {
		if _, _, err := densityLimit(limit); err == nil {
			t.Fatalf("invalid limit accepted %s", limit)
		}
	}
	for _, limit := range []string{"comfortable", "compact", "dense"} {
		slide := densityProbeDoc().Slides[0]
		slide.DensityLimit = limit
		dr, err := slideDensity(s, slide)
		if err != nil || dr.Limit != limit {
			t.Fatalf("limit %s %v %+v", limit, err, dr)
		}
		if limit != "dense" {
			slide.Density = "dense"
			if _, err := slideDensity(s, slide); err == nil || !strings.Contains(err.Error(), "prohibited_tier") {
				t.Fatal("prohibited requested tier clamped")
			}
		}
	}
	doc, err := LibrarySourceReference(densityTestBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, slide := range doc.Slides {
		if slide.DensityLimit == "" {
			continue
		}
		found++
		if slide.DensityLimit != "comfortable" {
			t.Fatalf("unexpected stock limit %s", slide.DensityLimit)
		}
		// Neither removing nor relaxing compiled metadata bypasses the source.
		for _, edited := range []string{"", "dense"} {
			slide.DensityLimit = edited
			slide.Density = "compact"
			if _, err := slideDensity(s, slide); err == nil || !strings.Contains(err.Error(), "prohibited_tier") {
				t.Fatal("bound source limit bypassed")
			}
		}
	}
	if found != 0 && found != 10 {
		t.Fatalf("limited source templates %d", found)
	}
	if !libraryFixedString["densityLimit"] {
		t.Fatal("source limit exposed as content slot")
	}
}
func TestAutoDensityStopsAtSourceLimitWithoutMutation(t *testing.T) {
	doc := densityProbeDoc()
	doc.Slides[0].DensityLimit = "comfortable"
	before, _ := json.Marshal(doc)
	_, _, err := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if err == nil || !strings.Contains(err.Error(), "density.limit_exhausted") || !strings.Contains(err.Error(), "split the slide") {
		t.Fatalf("source limit exhaustion %v", err)
	}
	after, _ := json.Marshal(doc)
	if string(before) != string(after) {
		t.Fatal("limit fit retry mutated authored slide")
	}
	doc.Slides[0].Nodes = nil
	_, report, err := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if report.Slides[0].Density.Limit != "comfortable" || report.Slides[0].Density.Resolved != "comfortable" || len(report.DensityAdjustments) != 0 {
		t.Fatal("limit silently altered preferred tier")
	}
}

func TestDensityGanttKeyGateInkMatchesSourceAndPreservesLegacy(t *testing.T) {
	if testing.Short() {
		t.Skip("source specimen rendering requires registered private branding; run make test-integration")
	}
	doc, err := LibrarySourceReference(densityTestBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range doc.Slides {
		if slide.TemplateBinding.Template != "plan/gantt" && slide.TemplateBinding.Template != "architecture/site-rollout" {
			continue
		}
		for _, level := range []string{"comfortable", "compact", "dense"} {
			r := tileTestRenderer(t, level)
			sourceCopy := *r.source
			r.source = &sourceCopy
			frame, err := r.source.ResolveFrame(slide.Frame)
			if err != nil {
				t.Fatal(err)
			}
			node := slide.Nodes[0]
			ctx := SceneContext{Surface: frame.Request.Surface, Zone: frame.Body, Path: node.Scene.Path, Keys: node.Scene.Keys}
			for _, modern := range []bool{false, true} {
				sourceCopy.densityVisualRules = modern
				p, err := r.planSceneNode(node.ID, node.Scene.Node, ctx)
				if err != nil {
					t.Fatal(err)
				}
				count := 0
				for _, it := range p.Items {
					if it.Shape == nil || it.Shape.Record.Color != "F900D3" || !strings.HasSuffix(it.Shape.Record.ID, ".chip") {
						continue
					}
					labelID := strings.TrimSuffix(it.Shape.Record.ID, ".chip") + ".label"
					for _, text := range p.Items {
						if text.Text != nil && text.Text.ID == labelID {
							count++
							want := "FFFFFF"
							if modern {
								want = "070154"
							}
							if text.Text.Color != want {
								t.Fatalf("%s %s modern%v gateink%s", slide.TemplateBinding.Template, level, modern, text.Text.Color)
							}
						}
					}
				}
				if count != 1 {
					t.Fatalf("expected one source keygate, got%d", count)
				}
			}
		}
	}
}
