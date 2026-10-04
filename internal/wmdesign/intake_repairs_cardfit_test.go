package wmdesign

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func cardFitCopy(v any, counts map[string]int) {
	switch n := v.(type) {
	case map[string]any:
		for _, x := range n {
			cardFitCopy(x, counts)
		}
	case []any:
		for _, x := range n {
			cardFitCopy(x, counts)
		}
	case string:
		counts[n]++
	}
}

func TestIntakeCardFitAllFrozenFailures(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen")
	entries := intakeRepairEntries(t, filepath.Join(root, "source", "templates", "library"))
	b, err := os.ReadFile(filepath.Join(root, "node-results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var failures []struct{ Template, Node, Type, Error string }
	if err = json.Unmarshal(b, &failures); err != nil {
		t.Fatal(err)
	}
	nodes := map[string][]string{}
	count := 0
	for _, n := range failures {
		if strings.Contains(n.Error, "scene.card_vertical_overflow") {
			nodes[n.Template] = append(nodes[n.Template], n.Node)
			count++
		}
	}
	if count != 43 || len(nodes) != 14 {
		t.Fatalf("frozen diagnostic scope changed: %d failures / %d templates", count, len(nodes))
	}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(filepath.Join(root, "source", "frames", "v0", "frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &source.Frames); err != nil {
		t.Fatal(err)
	}
	source.Revision = IntakeRepairRevision
	base := intakeTestRenderer(t)
	base.source = source
	keys := make([]string, 0, len(nodes))
	for key := range nodes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			slide := intakeRepairSlide(t, entries[key])
			before, _ := json.Marshal(slide)
			if err := ApplyIncomingIntakeRepairs(key, LibraryRevisionV3, &slide); err != nil {
				t.Fatal(err)
			}
			v3, _ := json.Marshal(slide)
			if !bytes.Equal(before, v3) {
				t.Fatal("v3 source changed")
			}
			originalCopy, newCopy := map[string]int{}, map[string]int{}
			for _, n := range slide.Nodes {
				if n.Scene != nil {
					obj, _ := libraryObject(n.Scene.Node)
					cardFitCopy(obj, originalCopy)
				}
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(slide)
			if bytes.Equal(before, after) {
				t.Fatal("card fit allocation unchanged")
			}
			if err := applyIncomingCardFitRepairs(key, &slide); err != nil {
				t.Fatal(err)
			}
			repeated, _ := json.Marshal(slide)
			if !bytes.Equal(after, repeated) {
				t.Fatal("card amendments not idempotent")
			}
			for _, n := range slide.Nodes {
				if n.Scene != nil {
					obj, _ := libraryObject(n.Scene.Node)
					cardFitCopy(obj, newCopy)
				}
			}
			if !reflect.DeepEqual(originalCopy, newCopy) {
				t.Fatal("copy/style/asset string content changed")
			}
			frame, err := source.ResolveFrame(slide.Frame)
			if err != nil {
				t.Fatal(err)
			}
			r := *base
			for _, id := range nodes[key] {
				t.Run(id, func(t *testing.T) {
					for _, n := range slide.Nodes {
						if n.ID == id {
							ctx := SceneContext{Surface: frame.Request.Surface, Zone: frame.Body, Path: n.Scene.Path, Keys: n.Scene.Keys, Notes: n.Scene.Notes}
							p, err := r.planSceneNode(n.ID, n.Scene.Node, ctx)
							if err != nil {
								t.Fatal(err)
							}
							if p == nil {
								t.Fatal("no card plan")
							}
							return
						}
					}
					t.Fatal("rejected source node missing")
				})
			}
			doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}
			if _, _, err := buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIntakeCardFitUnexpectedTopology(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source", "templates", "library")
	entries := intakeRepairEntries(t, root)
	slide := intakeRepairSlide(t, entries["cards/narrative-2x3-rows"])
	slide.Nodes = slide.Nodes[:len(slide.Nodes)-1]
	if err := applyIncomingCardFitRepairs("cards/narrative-2x3-rows", &slide); err == nil {
		t.Fatal("changed topology accepted")
	}
	slide = intakeRepairSlide(t, entries["road/cards"])
	obj, _ := libraryObject(slide.Nodes[1].Scene.Node)
	obj["card"].(map[string]any)["pad"] = 11
	slide.Nodes[1].Scene.Node, _ = json.Marshal(obj)
	if err := applyIncomingCardFitRepairs("road/cards", &slide); err == nil {
		t.Fatal("unreviewed padding accepted")
	}
	slide = intakeRepairSlide(t, entries["cards/narrative-2x3-rows"])
	obj, _ = libraryObject(slide.Nodes[len(slide.Nodes)-1].Scene.Node)
	obj["x"] = 704
	slide.Nodes[len(slide.Nodes)-1].Scene.Node, _ = json.Marshal(obj)
	before, _ := json.Marshal(slide)
	if err := ApplyIncomingIntakeRepairs("cards/narrative-2x3-rows", IntakeRepairRevision, &slide); err == nil {
		t.Fatal("changed position accepted")
	}
	after, _ := json.Marshal(slide)
	if !bytes.Equal(before, after) {
		t.Fatal("failed card amendment partially mutated caller")
	}
	slide = intakeRepairSlide(t, entries["sequence/seven-r"])
	obj, _ = libraryObject(slide.Nodes[4].Scene.Node)
	obj["w"] = 126
	slide.Nodes[4].Scene.Node, _ = json.Marshal(obj)
	before, _ = json.Marshal(slide)
	if err := ApplyIncomingIntakeRepairs("sequence/seven-r", IntakeRepairRevision, &slide); err == nil {
		t.Fatal("overlapping wider seven-R card accepted")
	}
	after, _ = json.Marshal(slide)
	if !bytes.Equal(before, after) {
		t.Fatal("failed seven-R width repair modifiedcaller")
	}
}
