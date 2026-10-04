package wmdesign

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIntakeTextFitFrozenLayouts(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	entries := intakeRepairEntries(t, filepath.Join(root, "templates", "library"))
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "frames", "v0", "frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &source.Frames); err != nil {
		t.Fatal(err)
	}
	source.Revision = IntakeRepairRevision
	for _, key := range []string{"adoption/dashboard-split", "funnel/compare", "timeline/swimlanes", "timeline/vertical", "cards/narrative-2x2-badge", "cards/narrative-2x3-badge", "activities/subtrack-timeline"} {
		t.Run(key, func(t *testing.T) {
			slide := intakeRepairSlide(t, entries[key])
			before, _ := json.Marshal(slide)
			if err := ApplyIncomingIntakeRepairs(key, LibraryRevisionV3, &slide); err != nil {
				t.Fatal(err)
			}
			old, _ := json.Marshal(slide)
			if !bytes.Equal(old, before) {
				t.Fatal("v3 changed")
			}
			copyBefore := map[string]int{}
			for _, n := range slide.Nodes {
				if n.Scene != nil {
					o, _ := libraryObject(n.Scene.Node)
					cardFitCopy(o, copyBefore)
				}
			}
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(slide)
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			repeated, _ := json.Marshal(slide)
			if !bytes.Equal(after, repeated) {
				t.Fatal("non-idempotent")
			}
			copyAfter := map[string]int{}
			for _, n := range slide.Nodes {
				if n.Scene != nil {
					o, _ := libraryObject(n.Scene.Node)
					cardFitCopy(o, copyAfter)
				}
			}
			if !reflect.DeepEqual(copyBefore, copyAfter) {
				t.Fatal("copy/style strings changed")
			}
			doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}
			_, report, err := buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Slides) != 1 || report.PowerPointVerified || report.VisuallyReviewed {
				t.Fatal("native build report incorrect")
			}
		})
	}
}

func TestIntakeNumberBadgeAndHypercarePadding(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = IntakeRepairRevision
	for _, raw := range []string{`{"type":"block","x":100,"y":180,"w":36,"h":36,"surface":"inverse","style":"number","text":"01"}`, `{"type":"block","x":100,"y":432,"w":72,"h":30,"surface":"outline","style":"small","text":"Hypercare"}`} {
		p, _, err := r.planDiagramScene("badge", json.RawMessage(raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, it := range p.Items {
			if it.Text != nil {
				found = true
				if len(it.Text.Layout.Lines) != 1 {
					t.Fatal("compact label wrapped")
				}
				if it.Text.Rect.X != 106 {
					t.Fatal("unexpected inset")
				}
			}
		}
		if !found {
			t.Fatal("no native text")
		}
	}
	r.source.Revision = LibraryRevisionV3
	for _, style := range []string{"small", "label"} {
		raw := `{"type":"block","x":100,"y":180,"w":54,"h":36,"surface":"inverse","style":"` + style + `","text":"Hi"}`
		p, _, err := r.planDiagramScene("legacy", json.RawMessage(raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range p.Items {
			if it.Text != nil && (it.Text.Rect.X != 106 || it.Text.Rect.W != 42) {
				t.Fatal("legacy compact padding changed")
			}
		}
	}
	_, _, err := r.planDiagramScene("badge", json.RawMessage(`{"type":"block","x":100,"y":180,"w":36,"h":36,"surface":"inverse","style":"number","text":"01"}`), SceneContext{Surface: "light"})
	if err == nil {
		t.Fatal("frozen v3 padding silently changed")
	}
}

func TestIntakeTextFitAtomicTopologyFailure(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	entries := intakeRepairEntries(t, filepath.Join(root, "templates", "library"))
	slide := intakeRepairSlide(t, entries["timeline/vertical"])
	// Fail after the title and preceding five rows have been amended on the clone.
	for _, n := range slide.Nodes {
		if n.Scene == nil {
			continue
		}
		obj, _ := libraryObject(n.Scene.Node)
		if obj["type"] == "block" && intakeRepairNumber(obj, "x") == 165 && intakeRepairNumber(obj, "y") == 396 {
			obj["y"] = 397
			n.Scene.Node, _ = json.Marshal(obj)
		}
	}
	before, _ := json.Marshal(slide)
	if err := ApplyIncomingIntakeRepairs("timeline/vertical", IntakeRepairRevision, &slide); err == nil {
		t.Fatal("unexpected timeline topology accepted")
	}
	after, _ := json.Marshal(slide)
	if !bytes.Equal(before, after) {
		t.Fatal("failed repair committed mutation")
	}
}

func TestIntakeTimelineSuppliedCopyRemainsBounded(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	entries := intakeRepairEntries(t, filepath.Join(root, "templates", "library"))
	r := intakeTestRenderer(t)
	r.source.Revision = IntakeRepairRevision
	for _, style := range []string{"subhead", "small"} {
		slide := intakeRepairSlide(t, entries["timeline/vertical"])
		if err := ApplyIncomingIntakeRepairs("timeline/vertical", IntakeRepairRevision, &slide); err != nil {
			t.Fatal(err)
		}
		checked := false
		for _, n := range slide.Nodes {
			if n.Scene == nil {
				continue
			}
			o, _ := libraryObject(n.Scene.Node)
			if intakeRepairNumber(o, "x") != 219 || o["style"] != style {
				continue
			}
			o["text"] = "This intentionally long supplied timeline description occupies multiple lines and must never overlap the following row even when all copy would fit the slide footer."
			raw, _ := json.Marshal(o)
			if _, err := r.planSceneNode(n.ID, raw, SceneContext{Surface: "light"}); err == nil {
				t.Fatal("long row text bypassed allocation")
			}
			checked = true
			break
		}
		if !checked {
			t.Fatal("missing bounded row fixture")
		}
	}
	// A resolution receipt never permits geometry drift to bypass revalidation.
	slide := intakeRepairSlide(t, entries["timeline/vertical"])
	if err := ApplyIncomingIntakeRepairs("timeline/vertical", IntakeRepairRevision, &slide); err != nil {
		t.Fatal(err)
	}
	for _, n := range slide.Nodes {
		if n.Scene == nil {
			continue
		}
		o, _ := libraryObject(n.Scene.Node)
		if intakeRepairNumber(o, "x") == 219 && o["style"] == "subhead" {
			o["h"] = 54
			n.Scene.Node, _ = json.Marshal(o)
			break
		}
	}
	before, _ := json.Marshal(slide)
	if err := ApplyIncomingIntakeRepairs("timeline/vertical", IntakeRepairRevision, &slide); err == nil {
		t.Fatal("receipt bypassed geometry guard")
	}
	after, _ := json.Marshal(slide)
	if !bytes.Equal(before, after) {
		t.Fatal("invalid repeated amendment partially mutated source")
	}
}
