package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestEditSlidesPreservesMetadataAndRejectsInvalidBindings(t *testing.T) {
	p := example(t)
	text := append([]byte("# exact project comment\n"), p.Raw...)
	text = bytes.Replace(text, []byte("  - id: local-composition"), []byte("  - id: local-composition\n    hidden: true\n    notes: \"\\nExact speaker notes\\n\""), 1)
	if err := os.WriteFile(p.SourcePath, text, 0644); err != nil {
		t.Fatal(err)
	}
	var err error
	p, err = Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = AddSection(p, SectionAddOptions{ID: "opening", Title: "Opening", BeforeSlideID: p.Document.Slides[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), p.Raw...)
	metadata := p.Document.Slides[1]
	sections := p.Document.Sections
	values := map[string]any{}
	for key, value := range metadata.Values {
		values[key] = value
	}
	values["summary"] = "\nLeading blank paragraph\nExact line two\n"
	r, err := EditSlides(p, map[string]SlideEdit{metadata.ID: {Values: values}}, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	got := updated.Document.Slides[1]
	if got.Notes != metadata.Notes || got.Hidden != metadata.Hidden || !reflect.DeepEqual(updated.Document.Sections, sections) || got.Values["summary"] != values["summary"] {
		t.Fatalf("edit changed metadata/strings: %+v", got)
	}
	if !bytes.Contains(updated.Raw, []byte("# exact project comment")) {
		t.Fatal("lost comments")
	}
	snapshot, err := os.ReadFile(filepath.Join(p.Root, "decisions/sources", r.BeforeSHA256+".yaml"))
	if err != nil || !bytes.Equal(snapshot, before) {
		t.Fatal("lost exact predecessor", err)
	}
	for _, bad := range []map[string]SlideEdit{
		{"absent": {Values: values}},
		{metadata.ID: {Values: map[string]any{"summary": 1}}},
		{metadata.ID: {Template: &Reference{Scope: "shared", ID: "missing"}}},
		{metadata.ID: {}},
	} {
		if _, err = EditSlides(updated, bad, bundle(t), wmdesign.CandidateEngine); err == nil {
			t.Fatal("invalid edit accepted")
		}
		after, _ := os.ReadFile(p.SourcePath)
		if !bytes.Equal(after, updated.Raw) {
			t.Fatal("invalid edit mutated source")
		}
	}
	if _, err = EditSlides(p, map[string]SlideEdit{metadata.ID: {Values: values}}, bundle(t), wmdesign.CandidateEngine); err == nil {
		t.Fatal("stale project edited source")
	}
}

func TestEditSharedTemplateAndValuesTogether(t *testing.T) {
	p := example(t)
	values := map[string]any{"title": "Use actual arguments", "eyebrow": "Recommendation", "cards": []any{map[string]any{"key": "a", "title": "Assess", "body": "Inspect the original purpose."}, map[string]any{"key": "b", "title": "Map", "body": "Select a matching template."}, map[string]any{"key": "c", "title": "Review", "body": "Review the native slide."}}}
	id := p.Document.Slides[0].ID
	_, err := EditSlides(p, map[string]SlideEdit{id: {Template: &Reference{Scope: "shared", ID: "cards/3", Revision: "1"}, Values: values}}, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Document.Slides[0].Template.ID != "cards/3" {
		t.Fatal("template not replaced")
	}
}

func TestSlidePatchStrictYAMLAndJSON(t *testing.T) {
	for _, raw := range []string{"one:\n  brief: context/page-briefs/one.md\n  values:\n    title: \"\\nLeading blank\\n\"\n", `{"one":{"brief":"context/page-briefs/one.md","values":{"title":"\nLeading blank\n"}}}`} {
		edits, err := DecodeSlideEdits([]byte(raw), "patch.yaml")
		if err != nil || edits["one"].Values["title"] != "\nLeading blank\n" {
			t.Fatalf("lost patch values: %v %+v", err, edits)
		}
	}
	for _, raw := range []string{"one: {brief: a, brief: b}", "one: {unknown: bad}", "one: {values: {title: a, title: b}}", "one: {brief: &b a}\ntwo: {brief: *b}", "one: {brief: a}\n---\ntwo: {brief: b}", "one: {hidden: false}", "one: {notes: overwritten}"} {
		if _, err := DecodeSlideEdits([]byte(raw), "patch.yaml"); err == nil {
			t.Fatalf("invalid patch accepted: %s", raw)
		}
	}
}

func TestSlideBriefIsRelativeFilePathAndInvalidEditsAreAtomic(t *testing.T) {
	p := example(t)
	brief := "context/page-briefs/one.md"
	if err := os.MkdirAll(filepath.Join(p.Root, "context/page-briefs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Root, brief), []byte("Page purpose: explain the decision.\nInclude source qualifications.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	id := p.Document.Slides[0].ID
	if _, err := EditSlides(p, map[string]SlideEdit{id: {Brief: &brief}}, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	p, err := Load(p.SourcePath)
	if err != nil || p.Document.Slides[0].Brief != brief {
		t.Fatalf("relative page brief path not retained: %v", err)
	}
	before := append([]byte(nil), p.Raw...)
	for _, bad := range []string{strings.Repeat("Editorial purpose ", 30), "Purpose: explain this decision", "Explain the decision\nInclude source qualifications", "Explain\tthe decision", "../brief.md"} {
		if _, err := EditSlides(p, map[string]SlideEdit{id: {Brief: &bad}}, bundle(t), wmdesign.CandidateEngine); err == nil || !strings.Contains(err.Error(), "/brief") || !strings.Contains(err.Error(), "project-relative path") {
			t.Fatalf("invalid brief should identify the path contract: %v", err)
		}
		after, err := os.ReadFile(p.SourcePath)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal("invalid brief edit changed source", err)
		}
	}
}
