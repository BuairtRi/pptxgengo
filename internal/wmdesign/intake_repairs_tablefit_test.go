package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var incomingTableFitKeys = []string{"stakeholders/quadrant-split", "readiness/scorecard", "activities/by-phase-table", "activities/ownership-split", "value-tracking/planned-vs-realized", "maturity/table-left"}

func incomingTableFitFixtures(t *testing.T) map[string]libraryEntry {
	t.Helper()
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source", "templates", "library")
	for _, pin := range []struct{ name, hash string }{
		{"change", "50c9b7c2bbd968fda1adb78ff0c4efdd2f823df0d17e6043514d442834c033ac"},
		{"modernization", "a4c4e5664d03f4d002d7967057962fb4b1ea93182d326cc3a4ab89300c2af2bb"},
		{"value", "26d1df3a65503b37e7f18e548c907ca17c1743f54491fd8f2750fccc04042298"}} {
		b, err := os.ReadFile(filepath.Join(root, pin.name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != pin.hash {
			t.Fatal("frozen table fixture changed", pin.name)
		}
	}
	return intakeRepairEntries(t, root)
}
func tableFitSemanticSlide(t *testing.T, slide SlideSpec) any {
	t.Helper()
	b, _ := json.Marshal(slide)
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	// Remove only the permitted allocation widths/row height and generated receipts.
	for _, raw := range v["nodes"].([]any) {
		node := raw.(map[string]any)
		scene, ok := node["scene"].(map[string]any)
		if !ok {
			continue
		}
		delete(scene, "resolutions")
		obj := scene["node"].(map[string]any)
		if obj["type"] == "schedule" {
			delete(obj, "rowHeight")
		}
		if obj["type"] == "table" {
			for _, raw := range obj["cols"].([]any) {
				delete(raw.(map[string]any), "w")
			}
		}
	}
	return v
}

func TestIncomingTableFitFullSlides(t *testing.T) {
	snapshot := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	entries := incomingTableFitFixtures(t)
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	frames, err := os.ReadFile(filepath.Join(snapshot, "frames", "v0", "frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(frames, &source.Frames); err != nil {
		t.Fatal(err)
	}
	source.Revision = IntakeRepairRevision
	for _, key := range incomingTableFitKeys {
		t.Run(key, func(t *testing.T) {
			slide := intakeRepairSlide(t, entries[key])
			before, _ := json.Marshal(slide)
			semantic := tableFitSemanticSlide(t, slide)
			if err := ApplyIncomingIntakeRepairs(key, LibraryRevisionV3, &slide); err != nil {
				t.Fatal(err)
			}
			old, _ := json.Marshal(slide)
			if !bytes.Equal(before, old) {
				t.Fatal("v3 changed")
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(slide)
			if bytes.Equal(before, after) {
				t.Fatal("no amendment")
			}
			if !reflect.DeepEqual(semantic, tableFitSemanticSlide(t, slide)) {
				t.Fatal("copy/data/identity or non-allocation field changed")
			}
			receipts := 0
			for _, node := range slide.Nodes {
				if node.Scene != nil {
					receipts += len(node.Scene.Resolutions)
				}
			}
			if receipts != 1 {
				t.Fatalf("resolution receipts%d", receipts)
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			repeated, _ := json.Marshal(slide)
			if !bytes.Equal(after, repeated) {
				t.Fatal("non-idempotent")
			}
			doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}
			pres, report, err := buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Slides) != 1 {
				t.Fatal("full slide missing")
			}
			path := filepath.Join(t.TempDir(), "native.pptx")
			if err := os.WriteFile(path, pres, 0600); err != nil {
				t.Fatal(err)
			}
			z, err := zip.OpenReader(path)
			if err != nil {
				t.Fatal(err)
			}
			defer z.Close()
			var xml []byte
			for _, file := range z.File {
				if file.Name == "ppt/slides/slide1.xml" {
					f, err := file.Open()
					if err != nil {
						t.Fatal(err)
					}
					xml, err = io.ReadAll(f)
					f.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			if !bytes.Contains(xml, []byte("<a:tbl>")) && key != "stakeholders/quadrant-split" {
				t.Fatal("table not native")
			}
			if key == "readiness/scorecard" {
				if bytes.Count(xml, []byte(`prst="blockArc"`)) != 6 || !bytes.Contains(xml, []byte("3 of 7")) || !bytes.Contains(xml, []byte("Overall readiness (of 4)")) {
					t.Fatal("later gauge/metric nodes missing")
				}
			}
			if key == "stakeholders/quadrant-split" {
				r := intakeTestRenderer(t)
				p, _, err := r.planPrimitiveScene("schedule", slide.Nodes[0].Scene.Node, SceneContext{Surface: "light"})
				if err != nil {
					t.Fatal(err)
				}
				for _, item := range p.Items {
					if item.Text != nil && item.Text.Rect.Y+item.Text.Rect.H > 394 {
						t.Fatal("schedule collides with note at396")
					}
				}
			}
		})
	}
}

func TestIncomingTableFitStrictAtomicTopology(t *testing.T) {
	entries := incomingTableFitFixtures(t)
	for _, key := range incomingTableFitKeys {
		for _, mutation := range []string{"row-count", "row-fields", "geometry", "duplicate", "missing", "column-width", "column-key", "column-type", "column-type-number"} {
			if key == "stakeholders/quadrant-split" && strings.HasPrefix(mutation, "column-") {
				continue
			}
			t.Run(key+"/"+mutation, func(t *testing.T) {
				slide := intakeRepairSlide(t, entries[key])
				index := 0
				for i, node := range slide.Nodes {
					if node.Scene != nil {
						o, _ := libraryObject(node.Scene.Node)
						if o["type"] == "table" || o["type"] == "schedule" {
							index = i
							break
						}
					}
				}
				scene := slide.Nodes[index].Scene
				obj, _ := libraryObject(scene.Node)
				rows := "rows"
				if obj["type"] == "schedule" {
					rows = "items"
				}
				switch mutation {
				case "row-count":
					obj[rows] = obj[rows].([]any)[:1]
				case "row-fields":
					obj[rows].([]any)[0].(map[string]any)["new"] = 1
				case "geometry":
					obj["x"] = 58
				case "duplicate":
					slide.Nodes = append(slide.Nodes, slide.Nodes[index])
				case "missing":
					scene.Path = "/body/missing"
				case "column-width":
					obj["cols"].([]any)[0].(map[string]any)["w"] = 215
				case "column-key":
					obj["cols"].([]any)[0].(map[string]any)["k"] = "new"
				case "column-type-number":
					obj["cols"].([]any)[0].(map[string]any)["type"] = 9
				case "column-type":
					obj["cols"].([]any)[0].(map[string]any)["type"] = "num"
				}
				scene.Node, _ = json.Marshal(obj)
				before, _ := json.Marshal(slide)
				if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err == nil {
					t.Fatal("unexpected topology accepted")
				}
				after, _ := json.Marshal(slide)
				if !bytes.Equal(before, after) {
					t.Fatal("failed atomic repair changed slide")
				}
			})
		}
	}
}
