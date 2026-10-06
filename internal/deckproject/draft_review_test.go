package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestDraftReviewMutationPreservesSourcesAndMetadata(t *testing.T) {
	for _, split := range []bool{false, true} {
		t.Run(map[bool]string{false: "inline", true: "split"}[split], func(t *testing.T) {
			p := example(t)
			if split {
				p, _ = splitExample(t, true)
			}
			id := p.Document.Slides[1].ID
			rel := p.SlideFiles[id]
			if rel == "" {
				rel = filepath.Base(p.SourcePath)
			}
			path := filepath.Join(p.Root, rel)
			raw := append([]byte("# Keep selected source comment\n"), p.SourceFiles[rel]...)
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			p, _ = Load(p.Root)
			before := p.Document.Slides[1]
			files := p.SourceFiles
			r, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status": "wip", "status_text": "Work in progress", "status_color": "brand.blue", "owner": "Ri", "due": "10/20", "updated": "10/05", "notes": "Check exact draft details\nSecond line\n"}})
			if err != nil {
				t.Fatal(err)
			}
			if r.BeforeSHA256 == r.AfterSHA256 || r.Decision == "" {
				t.Fatal("missing mutation receipt")
			}
			p, err = Load(p.Root)
			if err != nil {
				t.Fatal(err)
			}
			for name, raw := range files {
				if name != rel && !bytes.Equal(raw, p.SourceFiles[name]) {
					t.Fatalf("unselected source changed: %s", name)
				}
			}
			if !bytes.Contains(p.SourceFiles[rel], []byte("# Keep selected source comment")) {
				t.Fatal("source comment lost")
			}
			actual := p.Document.Slides[1]
			actual.DraftReview = nil
			if !reflect.DeepEqual(actual, before) {
				t.Fatal("unrelated slide metadata changed")
			}
			note := p.Document.Slides[1].DraftReview
			if note.Notes != "Check exact draft details\nSecond line\n" || note.Status != "wip" || note.StatusText != "Work in progress" || note.StatusColor != "brand.blue" {
				t.Fatal("note fields changed", note)
			}
			if _, err = OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status": "qa"}}); err != nil {
				t.Fatal(err)
			}
			p, _ = Load(p.Root)
			if p.Document.Slides[1].DraftReview.Owner != "Ri" || p.Document.Slides[1].DraftReview.Notes != note.Notes || p.Document.Slides[1].DraftReview.StatusText != note.StatusText || p.Document.Slides[1].DraftReview.StatusColor != note.StatusColor {
				t.Fatal("partial set erased fields")
			}
			r, err = OperateDraftReview(p, DraftReviewOperation{Action: "show", ID: id})
			if err != nil || r.DraftReview.Status != "qa" {
				t.Fatal(r, err)
			}
			if _, err = OperateDraftReview(p, DraftReviewOperation{Action: "clear", ID: id}); err != nil {
				t.Fatal(err)
			}
			p, _ = Load(p.Root)
			if p.Document.Slides[1].DraftReview != nil {
				t.Fatal("clear retained notes")
			}
		})
	}
}

func TestDraftReviewStrictValidationAndAtomicFailure(t *testing.T) {
	p := example(t)
	id := p.Document.Slides[0].ID
	for _, fields := range []map[string]string{{"status": "bad"}, {"status": "WIP"}, {"placement": "center"}, {"notes": "bad\x01"}, {"unknown": "x"}, {"status_color": "brand.orange"}, {"status_text": "bad\nlabel"}, {"status_text": strings.Repeat("x", 129)}} {
		if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: fields}); err == nil {
			t.Fatal("accepted invalid fields", fields)
		}
		raw, _ := os.ReadFile(p.SourcePath)
		if !bytes.Equal(raw, p.Raw) {
			t.Fatal("invalid mutation changed source")
		}
	}
	for _, raw := range []string{"{status: wip, unknown: x}", "{status: 1}", "{owner: Ri}", "{status: WIP}", "{status: wip, placement: middle}", "{status: wip, status_color: orange}", "{status: wip, status_text: 42}", "{status: wip, status_text: null}", "{status: wip, status_color: 42}"} {
		path := filepath.Join(p.Root, "invalid.yaml")
		data := strings.Replace(string(p.Raw), "  - id: maintain-the-source", "  - id: maintain-the-source\n    draft_review: "+raw, 1)
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatal("invalid YAML accepted", raw)
		}
	}
}

func TestDraftReviewCompilesAndInvalidatesBuild(t *testing.T) {
	p := example(t)
	pin(t, p)
	deps, err := dependencies(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range p.Document.Slides {
		if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: slide.ID, Fields: map[string]string{"status": "complete", "status_text": "Custom complete", "status_color": "kpi.off", "owner": "PRIVATE-OWNER"}}); err != nil {
			t.Fatal(err)
		}
		p, err = Load(p.Root)
		if err != nil {
			t.Fatal(err)
		}
	}
	compiled, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range compiled.Document.Slides {
		if slide.DraftReview == nil || slide.DraftReview.Owner != "PRIVATE-OWNER" || slide.DraftReview.StatusText != "Custom complete" || slide.DraftReview.StatusColor != "kpi.off" {
			t.Fatal("shared/local note lost", slide.ID)
		}
	}
	next, err := dependencies(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range p.Document.Slides {
		if deps["slide:"+slide.ID+":draft-review"] == next["slide:"+slide.ID+":draft-review"] {
			t.Fatal("note absent from invalidation")
		}
	}
	before := p.Canonical
	if _, err := Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	if !bytes.Equal(before, p.Canonical) {
		t.Fatal("split lost draft metadata")
	}
}

func TestDraftReviewClientExportCleanAndPrivateExportsRetain(t *testing.T) {
	p := example(t)
	pin(t, p)
	id := p.Document.Slides[0].ID
	if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status": "wip", "status_text": "PRIVATE-DRAFT-LABEL", "status_color": "brand.blue", "owner": "PRIVATE-DRAFT-OWNER", "notes": "PRIVATE-DRAFT-NOTES"}}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	build, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(p.Root, "builds", build.BuildID, "deck.pptx"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(nativeZIP(t, original)["ppt/slides/slide1.xml"]), "PRIVATE-DRAFT-NOTES") {
		t.Fatal("draft build missing note")
	}
	for _, mode := range []string{"client", "reviewer", "maintainer", "offline"} {
		out := filepath.Join(t.TempDir(), mode+".zip")
		if _, err := Export(p, ExportOptions{Mode: mode, Out: out, Bundle: bundle(t)}); err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(out)
		files := nativeZIP(t, data)
		pptxPath := "deck.pptx"
		if mode == "maintainer" || mode == "offline" {
			pptxPath = "builds/" + build.BuildID + "/deck.pptx"
		}
		deck := nativeZIP(t, []byte(files[pptxPath]))
		joined := ""
		for _, member := range deck {
			joined += member
		}
		contains := strings.Contains(joined, "PRIVATE-DRAFT") || strings.Contains(joined, "wm-review/") || strings.Contains(joined, "urn:pptxgengo:draft-review:v1")
		if contains != (mode != "client") {
			t.Fatalf("incorrect privacy in %s export", mode)
		}
	}
	current, _ := os.ReadFile(filepath.Join(p.Root, "builds", build.BuildID, "deck.pptx"))
	if !bytes.Equal(current, original) {
		t.Fatal("export mutated draft build")
	}
	packet, err := ProjectReview(p, ProjectReviewOptions{Stage: "content", Out: filepath.Join(t.TempDir(), "audience"), Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Audience: true})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(packet)
	if bytes.Contains(raw, []byte("PRIVATE-DRAFT")) {
		t.Fatal("draft note leaked into audience copy review")
	}
}

func TestDraftReviewMutationRetainsNoteCommentsAndRejectsSourceDrift(t *testing.T) {
	p := example(t)
	id := p.Document.Slides[0].ID
	p = rewrite(t, p, "  - id: maintain-the-source", "  - id: maintain-the-source\n    draft_review: # Keep component comment\n      status: wip # Keep status comment\n      owner: 'Ri' # Keep owner comment\n      notes: |\n        Original note")
	if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status": "qa"}}); err != nil {
		t.Fatal(err)
	}
	p, _ = Load(p.Root)
	for _, text := range []string{"# Keep component comment", "# Keep status comment", "owner: 'Ri' # Keep owner comment"} {
		if !bytes.Contains(p.Raw, []byte(text)) {
			t.Fatal("note comment/style lost", text)
		}
	}
	if err := os.WriteFile(p.SourcePath, append(append([]byte(nil), p.Raw...), []byte("\n# Concurrent change\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "clear", ID: id}); err == nil || !strings.Contains(err.Error(), "changed during mutation") {
		t.Fatal("stale source mutation accepted", err)
	}
	p, _ = Load(p.Root)
	if p.Document.Slides[0].DraftReview == nil {
		t.Fatal("source drift failure removed note")
	}
	guard := filepath.Join(p.Root, ".deck-source-mutation.lock")
	if err := os.WriteFile(guard, []byte("active"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "clear", ID: id}); err == nil || !strings.Contains(err.Error(), "another source mutation") {
		t.Fatal("active guard ignored", err)
	}
}

func TestDraftReviewNoOpIsByteIdempotent(t *testing.T) {
	p := example(t)
	id := p.Document.Slides[0].ID
	if r, err := OperateDraftReview(p, DraftReviewOperation{Action: "clear", ID: id}); err != nil || r.BeforeSHA256 != r.AfterSHA256 || r.Decision != "" {
		t.Fatal("empty clear was not a no-op", r, err)
	}
	raw, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(raw, p.Raw) {
		t.Fatal("empty clear reformatted source")
	}
	p = rewrite(t, p, "  - id: maintain-the-source", "  - id: maintain-the-source\n    draft_review: {status: wip, status_text: 'Custom status', status_color: brand.blue, notes: 'Original note'} # Keep flow style")
	before := append([]byte(nil), p.Raw...)
	if r, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status": "wip", "status_text": "Custom status", "status_color": "brand.blue", "notes": "Original note"}}); err != nil || r.BeforeSHA256 != r.AfterSHA256 || r.Decision != "" {
		t.Fatal("identical set was not a no-op", r, err)
	}
	raw, _ = os.ReadFile(p.SourcePath)
	if !bytes.Equal(raw, before) {
		t.Fatal("identical set reformatted source")
	}
	if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status_text": "", "status_color": ""}}); err != nil {
		t.Fatal(err)
	}
	p, err := Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	before = append([]byte(nil), p.Raw...)
	if r, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status_text": "", "status_color": ""}}); err != nil || r.BeforeSHA256 != r.AfterSHA256 || r.Decision != "" {
		t.Fatal("empty override repeat was not a no-op", r, err)
	}
	raw, _ = os.ReadFile(p.SourcePath)
	if !bytes.Equal(raw, before) {
		t.Fatal("empty override repeat reformatted source")
	}
}

func TestDraftReviewSurvivesTemplateSwap(t *testing.T) {
	for _, split := range []bool{false, true} {
		p := example(t)
		id := p.Document.Slides[0].ID
		if _, err := OperateDraftReview(p, DraftReviewOperation{Action: "set", ID: id, Fields: map[string]string{"status": "qa", "status_text": "Ready to review", "status_color": "kpi.risk", "owner": "Ri", "notes": "Keep internal note"}}); err != nil {
			t.Fatal(err)
		}
		p, _ = Load(p.Root)
		if split {
			if _, err := Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); err != nil {
				t.Fatal(err)
			}
			p, _ = Load(p.Root)
		}
		before := canonical(p.Document.Slides[0].DraftReview)
		proposal, err := ProposeSwap(p, id, "cards/3", bundle(t))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ApplySwap(p, proposal, false, bundle(t), wmdesign.CandidateEngine); err != nil {
			t.Fatal(err)
		}
		p, _ = Load(p.Root)
		if !bytes.Equal(before, canonical(p.Document.Slides[0].DraftReview)) {
			t.Fatal("template swap erased review note")
		}
	}
}
