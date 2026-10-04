package wmdesign

import (
	"math"
	"path/filepath"
	"strings"
	"testing"
)

// These assertions inspect primitives inside each composition. Outer frame
// bounds alone cannot detect a rule crossing a preceding wrapped paragraph.
func TestV5NativeOfferSectionsHavePrimitiveClearance(t *testing.T) {
	if testing.Short() {
		t.Skip("frozen composition render requires registered private branding assets; run make test-integration")
	}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	entries := intakeRepairEntries(t, filepath.Join(bundle, "source", "templates", "library"))
	for _, key := range []string{"offers-gantt/six-week-split", "offers-onepager/panel-right", "offers-sku/catalog-split"} {
		t.Run(key, func(t *testing.T) {
			slide := intakeRepairSlide(t, entries[key])
			if err := applyLibraryRefinements(key, LibraryRevisionV5, &slide); err != nil {
				t.Fatal(err)
			}
			_, report, err := buildWithLoadedSource(bundle, source, Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}, CandidateEngine, nil)
			if err != nil {
				t.Fatal(err)
			}
			s := report.Slides[0]
			if s.Frame.Request.Footer != "slim" {
				t.Fatal("source spacing has no supported footer reservation")
			}
			rules := map[string]bool{}
			if key == "offers-gantt/six-week-split" {
				rules["node08"], rules["node11"] = true, true
			}
			if key == "offers-onepager/panel-right" {
				rules["node12"], rules["node15"] = true, true
			}
			found := 0
			for _, shape := range s.Shapes {
				if !rules[shape.ID] {
					continue
				}
				found++
				for _, text := range s.Texts {
					if !v5RectsShareX(shape.Rect, text.Rect) {
						continue
					}
					gap := math.Max(text.Rect.Y-(shape.Rect.Y+shape.Rect.H), shape.Rect.Y-(text.Rect.Y+text.Rect.H))
					if gap < 3-.02 {
						t.Errorf("%s rule crosses or crowds %s (gap %.3fpt)", shape.ID, text.ID, gap)
					}
				}
			}
			if found != len(rules) {
				t.Fatal("native section rules missing")
			}
			if key == "offers-sku/catalog-split" {
				for _, pair := range [][2]string{{"node03", "node04"}, {"node04", "node05"}, {"node05", "node06"}, {"node06", "node07"}, {"node07", "node08"}} {
					a, b := v5TextStackBounds(t, s.Texts, pair[0]), v5TextStackBounds(t, s.Texts, pair[1])
					if b.Y-(a.Y+a.H) < 3-.02 {
						t.Errorf("section%s overlaps/crowds%s", pair[0], pair[1])
					}
				}
			}
		})
	}
}

func v5RectsShareX(a, b Rect) bool { return math.Min(a.X+a.W, b.X+b.W) > math.Max(a.X, b.X)+.02 }

func v5TextStackBounds(t *testing.T, texts []TextRecord, prefix string) Rect {
	t.Helper()
	var b Rect
	found := false
	for _, tr := range texts {
		if tr.ID != prefix && !strings.HasPrefix(tr.ID, prefix+".") {
			continue
		}
		if !found {
			b = tr.Rect
			found = true
		} else {
			b = diagramUnion(b, tr.Rect)
		}
	}
	if !found {
		t.Fatalf("text stack%s missing", prefix)
	}
	return b
}
