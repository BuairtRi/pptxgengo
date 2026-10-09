package deckproject

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Actual private six-slide sources supplement fast self-contained model tests.
// Read historical receipts/facts without changing executable pins or rebuilding
// their artifacts in a different caller.
func TestCommercialCoherentSixSourceClaims(t *testing.T) {
	root := os.Getenv("PPTXGENGO_COHERENT_COMMERCIAL_FIXTURES")
	if root == "" {
		t.Skip("private coherent CLI fixtures not selected")
	}
	for _, name := range []string{"chart-column-rail", "chart-column-rail-growth", "value-curve-scenarios", "pricing-options", "value-summary-big-number", "pricing-fixed-fee"} {
		t.Run(name, func(t *testing.T) {
			p, e := Load(filepath.Join(root, name, "project"))
			if e != nil {
				t.Fatal(e)
			}
			b, e := ReadTextBaseline(p, "", "")
			if e != nil {
				t.Fatal(e)
			}
			if b.Receipt.SourceSHA256 != p.SourceHash() {
				t.Fatal("source not same as genuine CLI build receipt")
			}
			slide := p.Document.Slides[0]
			if strings.Contains(slide.Values["title"].(string), "$") {
				t.Fatal("unmapped monetary headline retained")
			}
			local := p.Document.LocalTemplates[slide.Template.ID]
			nodes := map[string]Node{}
			for _, n := range local.Nodes {
				nodes[n.ID] = n
			}
			for _, n := range local.Nodes {
				if n.Definition.ID != "wmds/component/commercial" {
					continue
				}
				var s wmdesign.CommercialSpec
				if e = strictInto(n.Arguments, &s); e != nil {
					t.Fatal(e)
				}
				if _, _, e = wmdesign.MaterializeCommercialPresentation(s); e != nil {
					t.Fatal(e)
				}
				for _, id := range s.Group {
					other, ok := nodes[id]
					var mirror wmdesign.CommercialModel
					if !ok || strictInto(other.Arguments["model"], &mirror) != nil || !bytes.Equal(canonical(mirror), canonical(s.Model)) {
						t.Fatal("diverged or missing calculation participant")
					}
				}
				if bytes.Contains(canonical(s.Model), []byte(`"value":"123"`)) {
					t.Fatal("narrow sample123 survived full financial demo")
				}
			}
			if name == "value-curve-scenarios" {
				if _, ok := nodes["node02"]; ok {
					t.Fatal("obsolete month31 marker survived new model")
				}
				var s wmdesign.CommercialSpec
				if e = strictInto(nodes["node01"].Arguments, &s); e != nil {
					t.Fatal(e)
				}
				out, _, e := wmdesign.MaterializeCommercialPresentation(s)
				if e != nil {
					t.Fatal(e)
				}
				categories := out["categories"].([]any)
				if len(categories) != 6 || categories[0] != "Y0" || categories[5] != "Y5" {
					t.Fatal("reversed or incomplete timeline")
				}
				lineage, e := InspectNativeLineage(b.files["deck.pptx"], b.Objects)
				if e != nil {
					t.Fatal(e)
				}
				var chart *NativeLineageObject
				for i := range lineage.Objects {
					if lineage.Objects[i].NativeName == "node01.native" {
						chart = &lineage.Objects[i]
					}
				}
				if chart == nil {
					t.Fatal("owned native scenario chart absent")
				}
				facts, e := readNativeChartFacts(b.files["deck.pptx"], *chart, 6, 4, false)
				if e != nil {
					t.Fatal(e)
				}
				for _, value := range facts.data[3] {
					if value == nil || *value != 0 {
						t.Fatal("break-even target changed")
					}
				}
				for i, series := range out["series"].([]any) {
					values := series.(map[string]any)["values"].([]any)
					if len(values) != 6 {
						t.Fatal("cash-flow count lost")
					}
					for year, value := range values {
						want := -4.2 + []float64{2, 2.6, 3}[i]*float64(year)
						if math.Abs(value.(float64)-want) > 1e-12 || facts.data[i][year] == nil || math.Abs(*facts.data[i][year]-want) > 1e-12 {
							t.Fatalf("fact rounding/model/cache/workbook differ at series%d year%d:source%v cache%v expected%v", i, year, value, facts.data[i][year], want)
						}
					}
				}
			}
		})
	}
}

func TestCommercialNativeSixScenarioFacts(t *testing.T) {
	root := os.Getenv("PPTXGENGO_COHERENT_COMMERCIAL_FIXTURES")
	native := os.Getenv("PPTXGENGO_COHERENT_COMMERCIAL_NATIVE")
	if root == "" || native == "" {
		t.Skip("private actual native Save As not selected")
	}
	p, e := Load(filepath.Join(root, "combined-six", "project"))
	if e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(native)
	if e != nil {
		t.Fatal(e)
	}
	before, e := InspectNativeLineage(b.files["deck.pptx"], b.Objects)
	if e != nil {
		t.Fatal(e)
	}
	after, e := InspectNativeLineage(raw, b.Objects)
	if e != nil {
		t.Fatal(e)
	}
	var original, edited *NativeLineageObject
	for i := range before.Objects {
		o := &before.Objects[i]
		if o.NativePart == "ppt/slides/slide3.xml" && o.NativeName == "node01.native" {
			original = o
		}
	}
	if original == nil {
		t.Fatal("authored scenario chart missing")
	}
	for i := range after.Objects {
		if after.Objects[i].ShapeToken == original.ShapeToken {
			edited = &after.Objects[i]
		}
	}
	if edited == nil || edited.ParentToken != original.ParentToken || edited.SlideToken != original.SlideToken || edited.NativeName != original.NativeName {
		t.Fatal("native chart ownership changed")
	}
	a, e := readNativeChartFacts(b.files["deck.pptx"], *original, 6, 4, false)
	if e != nil {
		t.Fatal(e)
	}
	n, e := readNativeChartFacts(raw, *edited, 6, 4, false)
	if e != nil {
		t.Fatal(e)
	}
	if e = verifyChartSemanticClosure(a, n); e != nil {
		t.Fatal(e)
	}
	for i := range n.data {
		for year, v := range n.data[i] {
			want := 0.0
			if i < 3 {
				want = -4.2 + []float64{2, 2.6, 3}[i]*float64(year)
			}
			if v == nil || math.Abs(*v-want) > 1e-12 {
				t.Fatalf("Office changed series%d year%d fact: %v expected%v", i, year, v, want)
			}
		}
	}
}
