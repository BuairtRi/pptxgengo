package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestVennFeedbackCircleBoundsRegionTextAndPointMeaning(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV4
	for _, fixture := range frozenIntakeVennNodes {
		t.Run(fixture.key, func(t *testing.T) {
			var n intakeVennSource
			if err := json.Unmarshal([]byte(fixture.raw), &n); err != nil {
				t.Fatal(err)
			}
			if len(n.Sets) < 3 {
				return
			}
			p := intakeVennPlan(t, r, fixture.raw, SceneContext{Surface: "light", Path: "/body/0"})
			oldRadius, oldCenters := intakeVennGeometry(n)
			radius, centers, _ := intakeVennRoomierGeometry(n, SceneContext{}, oldRadius, oldCenters)
			if radius < oldRadius*1.2 || radius > oldRadius*1.3 {
				t.Fatalf("circle enlargement %.3f", radius/oldRadius)
			}
			for _, c := range centers {
				if c[0]-radius < n.X+.5 || c[0]+radius > n.X+n.W-.5 || c[1]-radius < n.Y+.5 || c[1]+radius > n.Y+n.H-.5 {
					t.Fatal("circle or outline outside body")
				}
			}
			for kind, count := range map[string]int{"sets": len(n.Sets), "regions": len(n.Regions)} {
				for i := 0; i < count; i++ {
					prefix := "venn." + kind + ".source-" + fmtVennOrdinal(i) + "."
					var b Rect
					first := true
					for _, it := range p.Items {
						if it.Text == nil || !strings.HasPrefix(it.Text.ID, prefix) {
							continue
						}
						if strings.HasSuffix(it.Text.ID, ".label") && len(strings.Fields(it.Text.Layout.Original)) == 1 && len(it.Text.Layout.Lines) != 1 {
							t.Fatalf("label word split: %s", it.Text.Layout.Original)
						}
						if first {
							b, first = it.Text.Rect, false
						} else {
							b = diagramUnion(b, it.Text.Rect)
						}
					}
					selected := []int{i}
					if kind == "regions" {
						selected = n.Regions[i].In
					}
					if !first && !intakeVennTextRectInRegion(b, intakeVennTextFit{centers: centers, radius: radius, selected: selected}) {
						t.Fatalf("%s text outside intended circle region: %+v", prefix, b)
					}
				}
			}
			for i, a := range p.Items {
				if a.Text == nil {
					continue
				}
				for _, bb := range p.Items[i+1:] {
					if bb.Text == nil {
						continue
					}
					ar, br := a.Text.Rect, bb.Text.Rect
					if math.Min(ar.X+ar.W, br.X+br.W) > math.Max(ar.X, br.X)+.02 && math.Min(ar.Y+ar.H, br.Y+br.H) > math.Max(ar.Y, br.Y)+.02 {
						t.Errorf("text allocations collide: %s %s", a.Text.ID, bb.Text.ID)
					}
				}
			}
			for i, pt := range n.Points {
				x, y := n.X+*pt.X*n.W, n.Y+*pt.Y*n.H
				originalMask := vennFeedbackMask(x, y, oldCenters, oldRadius)
				marker := enhancementShape(t, p, ".points.source-"+fmtVennOrdinal(i)+".marker")
				b := marker.Record.Rect
				if actual := vennFeedbackMask(b.X+b.W/2, b.Y+b.H/2, centers, radius); actual != originalMask {
					t.Fatalf("point %d membership %04b -> %04b", i, originalMask, actual)
				}
			}
		})
	}
}

func TestVennFeedbackLocalAllocationAndNaturalLensWrap(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV4
	raw := `{"type":"venn","x":100,"y":80,"w":420,"h":280,"sets":[{"label":"Desirable"},{"label":"Feasible"},{"label":"Viable"}],"regions":[{"in":[0,1,2],"label":"Sweet spot","w":65}],"labelStyle":"mono"}`
	p := intakeVennPlan(t, r, raw, SceneContext{Zone: Rect{100, 80, 420, 414}})
	if !inside(p.Bounds, Rect{100, 80, 420, 280}) {
		t.Fatal("local component used surrounding frame space")
	}
	var label *TextRecord
	for _, it := range p.Items {
		if it.Text != nil && strings.HasSuffix(it.Text.ID, ".regions.source-001.label") {
			label = it.Text
		}
	}
	if label == nil {
		t.Fatal("central label missing")
	}
	if len(label.Layout.Lines) != 2 || label.Layout.Original != "Sweet spot" {
		t.Fatal("narrow central lens did not wrap the complete original words")
	}
}

func TestVennFeedbackFrozenTemplateFramesAndNativeBadges(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	entries := intakeRepairEntries(t, filepath.Join(root, "templates", "library"))
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for key, entry := range entries {
		if !strings.HasPrefix(key, "venn/") {
			continue
		}
		count++
		t.Run(key, func(t *testing.T) {
			slide := intakeRepairSlide(t, entry)
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			data, _, err := buildWithLoadedSource(bundle, source, Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}, CandidateEngine, nil)
			if err != nil {
				t.Fatal(err)
			}
			if key != "venn/two-points" {
				return
			}
			z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range z.File {
				if f.Name != "ppt/slides/slide1.xml" {
					continue
				}
				rd, err := f.Open()
				if err != nil {
					t.Fatal(err)
				}
				b, err := io.ReadAll(rd)
				rd.Close()
				if err != nil {
					t.Fatal(err)
				}
				labels := 0
				for _, shape := range strings.Split(string(b), "<p:sp>")[1:] {
					shape = strings.Split(shape, "</p:sp>")[0]
					if !strings.Contains(shape, ".points.") || !strings.Contains(shape, ".label\"") {
						continue
					}
					labels++
					if !strings.Contains(shape, `anchor="ctr"`) || !strings.Contains(shape, `cy="196850"`) {
						t.Fatal("native badge full allocation/middle alignment missing")
					}
				}
				if labels != 5 {
					t.Fatalf("native badge labels %d", labels)
				}
			}
		})
	}
	if count != 12 {
		t.Fatalf("Venn templates %d", count)
	}
}

func fmtVennOrdinal(i int) string {
	s := strconv.Itoa(i + 1)
	return strings.Repeat("0", 3-len(s)) + s
}

func vennFeedbackMask(x, y float64, centers [][2]float64, radius float64) int {
	mask := 0
	for i, c := range centers {
		if math.Hypot(x-c[0], y-c[1]) <= radius {
			mask |= 1 << i
		}
	}
	return mask
}

func TestVennFeedbackNativeBadgeCentering(t *testing.T) {
	for _, rev := range []string{LibraryRevisionV3, LibraryRevisionV4} {
		r := intakeTestRenderer(t)
		r.source.Revision = rev
		p := intakeVennPlan(t, r, frozenIntakeVennNodes[1].raw, SceneContext{})
		for _, it := range p.Items {
			if it.Text == nil || !strings.Contains(it.Text.ID, ".points.") || !strings.HasSuffix(it.Text.ID, ".label") {
				continue
			}
			if rev == LibraryRevisionV4 {
				if it.Text.VerticalAlign != "middle" || it.Text.Rect.H != 15.5 {
					t.Fatal("native vertical centre or badge allocation missing")
				}
			} else if it.Text.VerticalAlign != "" || it.Text.Rect.H == 15.5 {
				t.Fatal("v3 centering changed")
			}
		}
	}
}

func TestFourCircleNativeFeedbackKeepsDistinctSemanticRegions(t *testing.T) {
	for _, version := range []string{"v5"} {
		bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", version, "intake-20261003-587-frozen", "bundle")
		source, err := Load(bundle, "")
		if err != nil {
			t.Fatal(err)
		}
		entries := intakeRepairEntries(t, filepath.Join(bundle, "source", "templates", "library"))
		r := intakeTestRenderer(t)
		r.source = source
		for _, key := range []string{"venn/four-text", "venn/four-points", "venn/four-callout"} {
			t.Run(version+"/"+key, func(t *testing.T) {
				slide := intakeRepairSlide(t, entries[key])
				if err := applyLibraryRefinements(key, source.Revision, &slide); err != nil {
					t.Fatal(err)
				}
				var n intakeVennSource
				if err := json.Unmarshal(slide.Nodes[0].Scene.Node, &n); err != nil {
					t.Fatal(err)
				}
				frame, err := source.ResolveFrame(slide.Frame)
				if err != nil {
					t.Fatal(err)
				}
				zone := frame.Body
				if frame.TallBody.W > 0 && n.X >= frame.TallBody.X && n.X+n.W <= frame.TallBody.X+frame.TallBody.W {
					zone = frame.TallBody
				}
				ctx := SceneContext{Surface: "light", Zone: zone, Path: "/body/0"}
				p, handled, err := r.planIntakeVennScene("four", slide.Nodes[0].Scene.Node, ctx)
				if err != nil || !handled {
					t.Fatal(handled, err)
				}
				oldRadius, oldCenters := intakeVennGeometry(n)
				radius, centers, _ := intakeVennRoomierGeometry(n, ctx, oldRadius, oldCenters)
				core, businessCase := false, false
				for i, region := range n.Regions {
					id := "four.regions.source-" + fmtVennOrdinal(i) + ".label"
					found := false
					for _, item := range p.Items {
						if item.Text == nil || item.Text.ID != id {
							continue
						}
						found = true
						if item.Text.Layout.Original != region.Label || !intakeVennTextRectInRegion(item.Text.Rect, intakeVennTextFit{centers: centers, radius: radius, selected: region.In}) {
							t.Fatalf("label %q changed or outside its exact region: %+v", region.Label, item.Text.Rect)
						}
						b := item.Text.Rect
						mask := intakeVennPointMask(b.X+b.W/2, b.Y+b.H/2, centers, radius)
						if strings.EqualFold(region.Label, "Core") {
							core = len(region.In) == 4 && mask == 15
						}
						if strings.EqualFold(region.Label, "Business case") {
							businessCase = len(region.In) == 2 && region.In[0] == 1 && region.In[1] == 3 && mask == 10
						}
					}
					if !found {
						t.Fatalf("region %q missing", region.Label)
					}
				}
				if key == "venn/four-callout" && (!core || !businessCase) {
					t.Fatal("CORE (all four) and BUSINESS CASE (Finance/Technology) were merged or moved")
				}
			})
		}
	}
}
