package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func intakeRepairEntries(t *testing.T, root string) map[string]libraryEntry {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]libraryEntry{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var family struct {
			Templates []libraryEntry `json:"templates"`
		}
		if err := json.Unmarshal(data, &family); err != nil {
			t.Fatal(err)
		}
		for _, entry := range family.Templates {
			out[entry.ID+"/"+entry.Variant] = entry
		}
	}
	return out
}

func intakeRepairSlide(t *testing.T, entry libraryEntry) SlideSpec {
	t.Helper()
	obj, err := libraryObject(entry.Slide)
	if err != nil {
		t.Fatal(err)
	}
	def := LibraryTemplate{SourceRevision: IntakeRepairRevision, RawSlide: entry.Slide}
	for i, n := range obj["body"].([]any) {
		libraryContentWalk(&def, n, fmt.Sprintf("/body/%d", i), fmt.Sprintf("node%02d", i+1), "", libraryProjectionContext{})
	}
	keys := map[string][]string{}
	for _, array := range def.Arrays {
		for i := 0; i < array.Count; i++ {
			keys[array.SourcePointer] = append(keys[array.SourcePointer], fmt.Sprintf("item-%03d", i+1))
		}
	}
	slide, err := compileLibrarySlide(entry.Slide, keys)
	if err != nil {
		t.Fatal(err)
	}
	slide.ID = "repair-" + strings.ReplaceAll(entry.ID+"/"+entry.Variant, "/", "-")
	return slide
}

func intakeRepairCopy(v any, field string, counts map[string]int) {
	switch value := v.(type) {
	case map[string]any:
		for key, child := range value {
			intakeRepairCopy(child, key, counts)
		}
	case []any:
		for _, child := range value {
			intakeRepairCopy(child, field, counts)
		}
	case string:
		switch field {
		case "text", "label", "title", "p", "sub", "lead", "value", "by", "axisLabel", "num", "items":
			if !strings.ContainsAny(value, "●○") {
				counts[value]++
			}
		}
	}
}

// The twenty rejection fixtures are immutable observations, not current live
// drafts. Source path/hash provenance remains in their original observation.
func TestIntakeRepairsFrozenTwentyRejections(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive frozen slide builds require registered private branding assets; run make test-integration")
	}
	snapshot := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003")
	entries := intakeRepairEntries(t, filepath.Join(snapshot, "source", "templates", "library"))
	data, err := os.ReadFile(filepath.Join(snapshot, "smoke-results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var results []struct{ Key, Status, Error string }
	if err = json.Unmarshal(data, &results); err != nil {
		t.Fatal(err)
	}
	var rejected []string
	for _, result := range results {
		if result.Status == "rejected" {
			rejected = append(rejected, result.Key)
		}
	}
	if len(rejected) != 20 {
		t.Fatalf("historical rejection fixture count changed: %d", len(rejected))
	}
	sort.Strings(rejected)
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	for _, key := range rejected {
		t.Run(key, func(t *testing.T) {
			entry, ok := entries[key]
			if !ok {
				t.Fatal("missing immutable rejection entry")
			}
			slide := intakeRepairSlide(t, entry)
			before, _ := json.Marshal(slide)
			if err := ApplyIncomingIntakeRepairs(key, LibraryRevisionV3, &slide); err != nil {
				t.Fatal(err)
			}
			unchanged, _ := json.Marshal(slide)
			if !bytes.Equal(before, unchanged) {
				t.Fatal("v3 composition changed")
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(slide)
			if bytes.Equal(before, after) {
				t.Fatal("rejected source received no explicit amendment")
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			repeated, _ := json.Marshal(slide)
			if !bytes.Equal(after, repeated) {
				t.Fatal("repair is not idempotent")
			}
			oldCopy, newCopy := map[string]int{}, map[string]int{}
			for i, node := range slide.Nodes {
				var old Node
				var original SlideSpec
				json.Unmarshal(before, &original)
				old = original.Nodes[i]
				if node.Scene != nil {
					x, _ := libraryObject(old.Scene.Node)
					y, _ := libraryObject(node.Scene.Node)
					intakeRepairCopy(x, "", oldCopy)
					intakeRepairCopy(y, "", newCopy)
				}
			}
			for text, count := range oldCopy {
				if newCopy[text] != count {
					t.Fatalf("source copy changed: %q old%d new%d", text, count, newCopy[text])
				}
			}
			if key == "vendors/rated-matrix" {
				for _, label := range []string{"Limited", "Partial", "Strong", "Leading"} {
					if newCopy[label] != 1 {
						t.Fatal("vendor rating label lost")
					}
				}
			}
			doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}
			_, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Slides) != 1 {
				t.Fatal("native repair slide missing")
			}
		})
	}
}

func TestIntakeRepairsAtomicAndFailClosed(t *testing.T) {
	slide := SlideSpec{Nodes: []Node{{ID: "badge", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(`{"type":"block","x":471,"y":306,"w":18,"h":18,"text":"1","style":"label","surface":"inverse"}`)}}, {ID: "table", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(`{"type":"table","x":57,"y":126,"w":558,"cols":[]}`)}}}}
	before, _ := json.Marshal(slide)
	if err := ApplyIncomingIntakeRepairs("capability-heat/callouts", IntakeRepairRevision, &slide); err == nil {
		t.Fatal("unexpected topology accepted")
	}
	after, _ := json.Marshal(slide)
	if !bytes.Equal(before, after) {
		t.Fatal("failed repair modified caller")
	}
	if err := ApplyIncomingIntakeRepairs("unknown/key", IntakeRepairRevision, nil); err == nil {
		t.Fatal("nil slide accepted")
	}
	callout := SlideSpec{Nodes: []Node{{ID: "callout", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(`{"type":"callout","x":579,"y":324,"w":324,"value":"-0.5","text":"off the weighted gap score.","secondary":{"value":"1","label":"Second"}}`)}}}}
	if err := ApplyIncomingIntakeRepairs("capability-heat/callouts", IntakeRepairRevision, &callout); err == nil {
		t.Fatal("new callout content silently discarded")
	}
	for _, text := range []string{"●○○○ Limited ●●○○ Partial", "●○○○ Limited ●●○○ Partial ●●●○ Strong ●●●●", "●○○○ Limited ●●○○ Partial ●●●○ Strong ●●●● Leading ○ Extra"} {
		if _, err := intakeRepairGlyphLegend(text); err == nil {
			t.Fatalf("malformed glyph legend accepted: %s", text)
		}
	}
}

func TestIntakeRepairsRound12IncomingTwentyTwo(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive frozen slide builds require registered private branding assets; run make test-integration")
	}
	testIntakeRepairsModernSnapshot(t, "intake-20261003-round12")
}

func TestIntakeRepairsFinalFrozenTwentyTwo(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive frozen slide builds require registered private branding assets; run make test-integration")
	}
	testIntakeRepairsModernSnapshot(t, "intake-20261003-frozen")
}

func TestIntakeRepairsFinalFrozenLifecycleCycles(t *testing.T) {
	snapshot := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	entries := intakeRepairEntries(t, filepath.Join(snapshot, "templates", "library"))
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	frames, err := os.ReadFile(filepath.Join(snapshot, "frames", "v0", "frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(frames, &source.Frames); err != nil {
		t.Fatal(err)
	}
	source.Revision = IntakeRepairRevision
	for _, key := range []string{"pdlc/cycle-overview", "pdlc/cycle-overview-split", "sdlc/cycle", "sdlc/cycle-split", "stakeholders/quadrant"} {
		t.Run(key, func(t *testing.T) {
			entry, ok := entries[key]
			if !ok {
				t.Fatal("incoming fixture missing")
			}
			slide := intakeRepairSlide(t, entry)
			before, _ := json.Marshal(slide)
			if err := ApplyIncomingIntakeRepairs(key, LibraryRevisionV3, &slide); err != nil {
				t.Fatal(err)
			}
			v3, _ := json.Marshal(slide)
			if !bytes.Equal(before, v3) {
				t.Fatal("v3 composition changed")
			}
			oldCopy, newCopy := map[string]int{}, map[string]int{}
			for _, node := range slide.Nodes {
				if node.Scene != nil {
					obj, _ := libraryObject(node.Scene.Node)
					intakeRepairCopy(obj, "", oldCopy)
				}
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(slide)
			if bytes.Equal(before, after) {
				t.Fatal("incoming composition not amended")
			}
			for _, node := range slide.Nodes {
				if node.Scene != nil {
					obj, _ := libraryObject(node.Scene.Node)
					intakeRepairCopy(obj, "", newCopy)
				}
			}
			for text, count := range oldCopy {
				if newCopy[text] != count {
					t.Fatalf("copy changed: %q", text)
				}
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			repeated, _ := json.Marshal(slide)
			if !bytes.Equal(after, repeated) {
				t.Fatal("incoming amendment not idempotent")
			}
			doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}
			if _, _, err := buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testIntakeRepairsModernSnapshot(t *testing.T, observation string) {
	t.Helper()
	snapshot := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", observation, "source")
	entries := intakeRepairEntries(t, filepath.Join(snapshot, "templates", "library"))
	history, err := os.ReadFile(filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003", "smoke-results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var results []struct{ Key, Status string }
	if err = json.Unmarshal(history, &results); err != nil {
		t.Fatal(err)
	}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
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
	results = append(results, struct{ Key, Status string }{"maturity/table-tall", "rejected"}, struct{ Key, Status string }{"maturity/insights", "rejected"})
	for _, result := range results {
		if result.Status != "rejected" {
			continue
		}
		t.Run(result.Key, func(t *testing.T) {
			entry, ok := entries[result.Key]
			if !ok {
				t.Fatal("historical template missing from newer intake")
			}
			slide := intakeRepairSlide(t, entry)
			if err := ApplyIncomingIntakeRepairs(result.Key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}
			if _, _, err = buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIntakeScoreLegendNativeAndStrict(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"scorelegend","x":57,"y":126,"w":432,"max":4,"items":[{"value":1,"label":"Limited"},{"value":2,"label":"Partial"},{"value":3,"label":"Strong"},{"value":4,"label":"Leading"}]}`)
	p, handled, err := r.planIntakeScoreLegendScene("legend", raw, SceneContext{Surface: "light"})
	if err != nil || !handled {
		t.Fatalf("handled%v error%v", handled, err)
	}
	shapes, texts := 0, 0
	for _, item := range p.Items {
		if item.Shape != nil {
			shapes++
		}
		if item.Text != nil {
			texts++
		}
		if item.Image != nil || item.Table != nil {
			t.Fatal("native score legend flattened")
		}
	}
	if shapes != 16 || texts != 4 || len(p.Groups) != 1 || p.Groups[0].Contract != IntakeScoreLegendContract {
		t.Fatal("native marker/label/group identities missing")
	}
	if err := sceneTextEnvelope(p, SceneContext{Zone: Rect{57, 126, 432, 18}}); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`{"type":"scorelegend","x":null,"y":0,"w":432,"max":4,"items":[{"value":1,"label":"X"}]}`, `{"type":"scorelegend","x":0,"y":0,"w":432,"max":13,"items":[{"value":1,"label":"X"}]}`, `{"type":"scorelegend","x":0,"y":0,"w":432,"max":4,"items":[{"value":null,"label":"X"}]}`, `{"type":"scorelegend","x":0,"y":0,"w":432,"max":4,"items":[{"value":5,"label":"X"}]}`, `{"type":"scorelegend","x":0,"y":0,"w":20,"max":4,"items":[{"value":1,"label":"X"}]}`, `{"type":"scorelegend","x":0,"y":0,"w":432,"max":4,"items":[{"value":1,"label":"X","other":true}]}`} {
		if _, _, err := r.planIntakeScoreLegendScene("bad", json.RawMessage(bad), SceneContext{}); err == nil {
			t.Fatalf("invalid score legend accepted: %s", bad)
		}
	}
}
