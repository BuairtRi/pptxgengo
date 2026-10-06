package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestV10AddedTemplateIndividualBuilds(t *testing.T) {
	source, err := LibrarySourceReference(v10IntakeBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	for _, key := range v10AddedTemplateKeys {
		keys[key] = true
	}
	for _, slide := range source.Slides {
		if !keys[slide.TemplateBinding.Template] {
			continue
		}
		t.Run(slide.TemplateBinding.Template, func(t *testing.T) {
			doc := source
			doc.Slides = []SlideSpec{slide}
			if _, _, err := BuildWithEngine(v10IntakeBundle(), "", doc, CandidateEngine); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRoadForkNativeCubicChosenAndKeyedContent(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV10
	raw := json.RawMessage(`{"type":"roadfork","x":57,"y":126,"w":846,"h":264,"mode":"decision","chosen":0,"small":true,"trunk":[{"label":"Foundation","n":7,"here":true,"hereLabel":"Current"}],"fork":{"label":"Choose","at":0.3,"tag":"Approved"},"branches":[{"title":"First","milestones":[{"label":"Build"}]},{"title":"Second","milestones":[{"label":"Buy"}]}]}`)
	ctx := SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/trunk": {"base"}, "/body/0/branches": {"first", "second"}, "/body/0/branches/0/milestones": {"build"}, "/body/0/branches/1/milestones": {"buy"}}}
	p, handled, err := r.planRoadForkScene("fork", raw, ctx)
	if !handled || err != nil {
		t.Fatalf("handled=%v: %v", handled, err)
	}
	chosen := enhancementShape(t, p, ".first.road")
	grey := enhancementShape(t, p, ".second.road")
	if chosen.Record.Color != "070154" || grey.Record.Color != "CED7E6" || chosen.Props.Points[1].Curve == nil || chosen.Props.Points[1].Curve.Type != "cubic" {
		t.Fatal("source cubic/color semantics missing")
	}
	if enhancementShape(t, p, ".base.pin").Record.Color != "F900D3" {
		t.Fatal("here flag did not mark authored pin")
	}
	choiceIndex, otherIndex := -1, -1
	for i, it := range p.Items {
		if it.Shape != nil {
			if it.Shape.Record.ID == "fork.branches.first.road" {
				choiceIndex = i
			}
			if it.Shape.Record.ID == "fork.branches.second.road" {
				otherIndex = i
			}
		}
	}
	if choiceIndex <= otherIndex {
		t.Fatal("chosen path not last")
	}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "fork", Frame: FrameRequest{Footer: "compact", TitleLines: 1}, Eyebrow: "Roadmap", Title: "Choose the road", Nodes: []Node{{ID: "fork", Kind: "scene", Scene: &SceneSpec{Node: raw, Path: ctx.Path, Keys: ctx.Keys}}}}}}
	deck, _, err := BuildWithEngine(v10IntakeBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		t.Fatal(err)
	}
	var xml string
	for _, f := range zr.File {
		if f.Name == "ppt/slides/slide1.xml" {
			s, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			b, e := io.ReadAll(s)
			s.Close()
			if e != nil {
				t.Fatal(e)
			}
			xml = string(b)
		}
	}
	for _, want := range []string{"<a:cubicBezTo>", "fork.branches.first.road", "fork.trunk.base.number", "APPROVED", "CURRENT"} {
		if !strings.Contains(xml, want) {
			t.Errorf("native XML missing %s", want)
		}
	}
	if strings.Contains(xml, "<p:pic>") {
		t.Fatal("roadfork rendered as image")
	}
}

func TestRoadForkSourceNumberAndNativeNoWrapClearance(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{{`0`, "A1"}, {`"0"`, "0"}, {`7`, "7"}, {`null`, "A1"}} {
		got, err := roadForkNumber(json.RawMessage(tc.raw), "A1")
		if err != nil || got != tc.want {
			t.Fatalf("number %s: got %q, want %q: %v", tc.raw, got, tc.want, err)
		}
	}
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV10
	p := enhancementPlan(t, r, `{"type":"roadfork","x":57,"y":126,"w":846,"h":264,"trunk":[{"label":"Foundation","here":true}],"branches":[{"title":"First","milestones":[{"label":"Build"}]},{"title":"Second","milestones":[{"label":"Buy"}]}]}`)
	labels := 0
	for _, item := range p.Items {
		if item.Text == nil || (!strings.HasSuffix(item.Text.ID, ".tag") && !strings.HasSuffix(item.Text.ID, ".here.label")) {
			continue
		}
		tr := item.Text
		if len(tr.Layout.Lines) != 1 || tr.Rect.W-tr.Layout.Lines[0].Advance < roadForkNoWrapClearance-.02 {
			t.Fatalf("source nowrap label lacks native clearance: %s", tr.ID)
		}
		labels++
	}
	if labels != 3 {
		t.Fatalf("expected two branch tags and current position label, got %d", labels)
	}
}

func TestV10WavesCriteriaTitleIsPresent(t *testing.T) {
	doc, err := LibrarySourceReference(v10IntakeBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range doc.Slides {
		if slide.TemplateBinding.Template != "roadmap-narrative/waves-criteria" {
			continue
		}
		doc.Slides = []SlideSpec{slide}
		_, report, err := BuildWithEngine(v10IntakeBundle(), "", doc, CandidateEngine)
		if err != nil {
			t.Fatal(err)
		}
		for _, tr := range report.Slides[0].Texts {
			if tr.ID == "title" {
				if len(tr.Layout.Lines) != 2 {
					t.Fatalf("title line count=%d", len(tr.Layout.Lines))
				}
				return
			}
		}
		t.Fatal("authored title is missing")
	}
	t.Fatal("waves-criteria template missing")
}
