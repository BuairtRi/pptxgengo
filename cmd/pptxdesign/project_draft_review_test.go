package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func TestProjectDraftReviewCLI(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "deck.yaml"), []byte("schema: pptxgengo.deck-document.v1\nid: draft-cli\ntitle: Draft CLI\nyear: 2026\ntoolchain: {lockfile: toolchain.lock.json}\nslides:\n  - id: one\n    content_kind: supplied_content\n    template: {scope: shared, id: cards/3}\n    values: {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"set", "--status", "wip", "--status-text", "Custom status", "--status-color", "brand.blue", "--owner", "Ri", "--notes", "Exact\nnotes"}, {"set", "--status", "qa"}, {"show"}, {"set", "--owner", ""}} {
		command := append([]string{"slide", "draft-review"}, args...)
		command = append(command, "--project", root, "--id", "one")
		if err := runProject(command); err != nil {
			t.Fatal(command, err)
		}
	}
	p, err := deckproject.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if note := p.Document.Slides[0].DraftReview; note.Status != "qa" || note.Owner != "" || note.Notes != "Exact\nnotes" || note.StatusText != "Custom status" || note.StatusColor != "brand.blue" {
		t.Fatal("partial CLI update changed other fields", note)
	}
	for _, step := range []struct {
		args        []string
		text, color string
	}{
		{[]string{"--status-color", "kpi.off"}, "Custom status", "kpi.off"},
		{[]string{"--status-text", "Ready for signoff"}, "Ready for signoff", "kpi.off"},
		{[]string{"--status-text", "", "--status-color", ""}, "", ""},
	} {
		args := append([]string{"slide", "draft-review", "set", "--project", root, "--id", "one"}, step.args...)
		if err = runProject(args); err != nil {
			t.Fatal(err)
		}
		p, err = deckproject.Load(root)
		if err != nil {
			t.Fatal(err)
		}
		if note := p.Document.Slides[0].DraftReview; note.Status != "qa" || note.StatusText != step.text || note.StatusColor != step.color {
			t.Fatal("overrides are not independent", note)
		}
	}
	if err = runProject([]string{"slide", "draft-review", "clear", "--project", root, "--id", "one"}); err != nil {
		t.Fatal(err)
	}
	p, _ = deckproject.Load(root)
	if p.Document.Slides[0].DraftReview != nil {
		t.Fatal("clear failed")
	}
	for _, args := range [][]string{{}, {"bad"}, {"show", "--status", "qa"}, {"clear", "--notes", "x"}, {"set", "unexpected"}, {"set", "--unknown", "x"}} {
		if err := runProjectSlideDraftReview(args); err == nil {
			t.Fatal("invalid CLI accepted", args)
		}
	}
	for _, args := range [][]string{{"set", "--status", "invalid"}, {"set"}, {"set", "--placement", "middle"}, {"set", "--status-color", "orange"}, {"set", "--status-text", "invalid\nlabel"}} {
		args = append(args, "--project", root, "--id", "one")
		if err := runProjectSlideDraftReview(args); err == nil {
			t.Fatal("invalid operation accepted", args)
		}
	}
}
