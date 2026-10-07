package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func reuseProject(t *testing.T) *Project {
	t.Helper()
	p := example(t)
	p = rewrite(t, p, "content_kind: synthetic_example", "content_kind: supplied_content")
	entries := map[string]CompositionEntry{}
	for _, s := range p.Document.Slides {
		entries[s.ID] = CompositionEntry{Purpose: "Fixture authored content", Rationale: "Explicit test composition decision", ChosenTemplate: s.Template.ID}
	}
	if err := os.WriteFile(filepath.Join(p.Root, "composition-log.yaml"), canonical(CompositionLog{Schema: "pptxgengo.composition-log.v1", Slides: entries}), 0644); err != nil {
		t.Fatal(err)
	}
	pin(t, p)
	return p
}

func publishReuse(t *testing.T, p *Project, revision int) (string, finishedslide.Manifest) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "closed revision")
	m, err := PublishFinishedSlide(p, FinishedSlidePublishOptions{SlideID: p.Document.Slides[0].ID, Out: out, Bundle: bundle(t), Engine: wmdesign.CandidateEngine, Manifest: finishedslide.Manifest{ID: "curated/slide/test-message", Revision: revision, Name: "Authored test message", Purpose: "Demonstrate independent maintained content", Owner: "Test fixture", Lifecycle: "draft"}})
	if err != nil {
		t.Fatal(err)
	}
	return out, m
}

func TestFinishedSlideIndependentInsertion(t *testing.T) {
	source := reuseProject(t)
	library, m := publishReuse(t, source, 1)
	original, err := os.ReadFile(filepath.Join(library, "slide.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	targets := []*Project{reuseProject(t), reuseProject(t)}
	keys := []string{}
	for i, p := range targets {
		if _, err := Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); err != nil {
			t.Fatal(err)
		}
		p, err = Load(p.Root)
		if err != nil {
			t.Fatal(err)
		}
		logPath := filepath.Join(p.Root, "composition-log.yaml")
		oldLog := mustRead(t, logPath)
		if err := os.WriteFile(logPath, append([]byte("# Authored composition comment remains\n"), oldLog...), 0644); err != nil {
			t.Fatal(err)
		}
		r, err := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "reused-message", Rationale: "Fixture content selected for this deck", AllowDraft: true, Before: p.Document.Slides[0].ID, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
		if err != nil {
			t.Fatal(err)
		}
		if r.Library.ID != m.ID || r.Library.RevisionSHA256 != m.RevisionSHA256 || len(r.Library.ItemRemaps) != 3 {
			t.Fatal(r)
		}
		p, err = Load(p.Root)
		if err != nil {
			t.Fatal(err)
		}
		targets[i] = p
		if !bytes.Contains(mustRead(t, logPath), []byte("Authored composition comment remains")) {
			t.Fatal("composition comment lost")
		}
		entries, err := ReadCompositionLog(p)
		if err != nil {
			t.Fatal(err)
		}
		if entries["reused-message"].Library == nil || entries["reused-message"].Library.Revision != 1 {
			t.Fatal(entries)
		}
		s := p.Document.Slides[0]
		if s.ID != "reused-message" {
			t.Fatal("position changed", s.ID)
		}
		keys = append(keys, s.Values["cards"].([]any)[0].(map[string]any)["key"].(string))
		compiled, err := Check(p, bundle(t), wmdesign.CandidateEngine)
		if err != nil {
			t.Fatal(err)
		}
		native, _, err := wmdesign.BuildWithEngineAndAssets(bundle(t), "", compiled.Document, wmdesign.CandidateEngine, compiled.Assets)
		if err != nil || len(native) == 0 {
			t.Fatal("independent deck cannot build", err)
		}
	}
	if keys[0] == keys[1] || keys[0] == "source" || keys[1] == "source" {
		t.Fatal("item identities were shared", keys)
	}
	_, newer := publishReuse(t, source, 2)
	if newer.RevisionSHA256 == m.RevisionSHA256 {
		t.Fatal("revision digest unchanged")
	}
	beforeOther := append([]byte(nil), targets[1].SourceFiles[targets[1].SlideFiles["reused-message"]]...)
	values := targets[0].Document.Slides[0].Values
	values["title"] = "Adapted only in first deck"
	if _, err := EditSlides(targets[0], map[string]SlideEdit{"reused-message": {Values: values}}, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(filepath.Join(library, "slide.yaml")); !bytes.Equal(raw, original) {
		t.Fatal("library changed")
	}
	other, _ := Load(targets[1].Root)
	if !bytes.Equal(other.SourceFiles[other.SlideFiles["reused-message"]], beforeOther) {
		t.Fatal("other deck changed")
	}
	log, err := ReadCompositionLog(other)
	if err != nil || log["reused-message"].Library.Revision != 1 {
		t.Fatal("revision upgrade changed prior lineage", err)
	}
}

func TestFinishedSlideRejectsBeforeWrites(t *testing.T) {
	p := reuseProject(t)
	library, _ := publishReuse(t, p, 1)
	before := append([]byte(nil), p.Raw...)
	cases := []FinishedSlideInsertOptions{
		{ID: "fresh", Rationale: "Intentional reuse"},
		{ID: p.Document.Slides[0].ID, Rationale: "Intentional reuse", AllowDraft: true},
		{ID: "fresh", Rationale: "", AllowDraft: true},
		{ID: "fresh", Rationale: "Intentional reuse", AllowDraft: true, Before: "missing"},
	}
	for _, o := range cases {
		o.Package = library
		o.Bundle = bundle(t)
		o.Engine = wmdesign.CandidateEngine
		if _, err := InsertFinishedSlide(p, o); err == nil {
			t.Fatal("invalid insertion accepted", o)
		}
		raw, _ := os.ReadFile(p.SourcePath)
		if !bytes.Equal(raw, before) {
			t.Fatal("rejected insertion mutated source")
		}
	}
	if err := os.WriteFile(filepath.Join(p.Root, ".deck-source-mutation.lock"), []byte("other writer"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "fresh", Rationale: "Intentional reuse", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err == nil || !strings.Contains(err.Error(), "active") {
		t.Fatal("guard was bypassed", err)
	}
	if _, err := os.Stat(filepath.Join(p.Root, "slides/fresh.yaml")); !os.IsNotExist(err) {
		t.Fatal("guarded insertion wrote slide")
	}
	if err := os.Remove(filepath.Join(p.Root, ".deck-source-mutation.lock")); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(library, "slide.yaml"))
	raw = append(raw, ' ')
	if err := os.WriteFile(filepath.Join(library, "slide.yaml"), raw, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "fresh", Rationale: "Intentional reuse", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err == nil {
		t.Fatal("tampered package accepted")
	}
}

func TestFinishedSlideDoesNotRelabelExamples(t *testing.T) {
	p := example(t)
	pin(t, p)
	_, err := PublishFinishedSlide(p, FinishedSlidePublishOptions{SlideID: p.Document.Slides[0].ID, Out: filepath.Join(t.TempDir(), "new"), Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err == nil || !strings.Contains(err.Error(), "content_complete") {
		t.Fatal("synthetic specimen relabeled", err)
	}
	p = reuseProject(t)
	p.Document.Slides[0].EvidenceRefs = []string{"claim"}
	_, err = PublishFinishedSlide(p, FinishedSlidePublishOptions{SlideID: p.Document.Slides[0].ID, Out: filepath.Join(t.TempDir(), "new"), Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err == nil || !strings.Contains(err.Error(), "evidence_dependency") {
		t.Fatal("claims omitted", err)
	}
}

func TestObservedDependencyDriftDoesNotMutateSource(t *testing.T) {
	p := reuseProject(t)
	raw, err := os.ReadFile(filepath.Join(p.Root, "composition-log.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	observed := map[string][]byte{"composition-log.yaml": raw}
	_, err = commitSourceChangesObserved(p, map[string][]byte{filepath.Base(p.SourcePath): p.Raw}, observed, func(*Project) error {
		return os.WriteFile(filepath.Join(p.Root, "composition-log.yaml"), []byte("changed externally"), 0644)
	})
	if err == nil || !strings.Contains(err.Error(), "dependency changed") {
		t.Fatal("editorial drift lost", err)
	}
	now, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(now, p.Raw) {
		t.Fatal("source changed after dependency drift")
	}
}

func TestFinishedSlideLineageRoundTrip(t *testing.T) {
	entry := CompositionEntry{Library: &LibraryLineage{ID: "curated/slide/example", Revision: 1}}
	var decoded CompositionEntry
	if err := json.Unmarshal(canonical(entry), &decoded); err != nil || decoded.Library == nil {
		t.Fatal(err)
	}
}

func TestFinishedSlideAssetsCopyAndSlotIdentities(t *testing.T) {
	source := reuseProject(t)
	catalog, err := wmdesign.LibraryCatalog(bundle(t), "")
	if err != nil {
		t.Fatal(err)
	}
	examples, err := wmdesign.LibraryReference(bundle(t), "", "covers", 2026)
	if err != nil {
		t.Fatal(err)
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	for _, def := range catalog {
		defs[def.Key] = def
	}
	found := false
	for _, s := range examples.Slides {
		def := defs[s.Template]
		var values map[string]any
		if err := json.Unmarshal(s.Values, &values); err != nil {
			t.Fatal(err)
		}
		slide := Slide{ID: "authored-cover", ContentKind: "supplied_content", Template: Reference{Scope: "shared", ID: s.Template}, Values: values}
		media, err := finishedMediaSlots(slide, def)
		if err != nil {
			t.Fatal(err)
		}
		for slot := range media {
			values["slots"].(map[string]any)[slot] = "project:sample-image"
			// The business title deliberately equals the asset name and must retain it.
			for _, field := range def.Slots {
				if field.Name != slot && field.Kind == "string" {
					parts, e := contentPointerParts(field.SourcePointer)
					if e == nil && !sharedImagePointer(mustSourceMap(t, def), parts) {
						values["slots"].(map[string]any)[field.Name] = "sample-image"
						break
					}
				}
			}
			source.Document.Slides = []Slide{slide}
			source.Document.LocalTemplates = nil
			if err := os.WriteFile(source.SourcePath, canonical(source.Document), 0644); err != nil {
				t.Fatal(err)
			}
			log := CompositionLog{Schema: "pptxgengo.composition-log.v1", Slides: map[string]CompositionEntry{slide.ID: {Purpose: "Owned image fixture", Rationale: "Exercise actual declared image slot", ChosenTemplate: slide.Template.ID}}}
			if err := os.WriteFile(filepath.Join(source.Root, "composition-log.yaml"), canonical(log), 0644); err != nil {
				t.Fatal(err)
			}
			source, err = Load(source.Root)
			if err != nil {
				t.Fatal(err)
			}
			found = true
			break
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no cover with real image slot")
	}
	library, m := publishReuse(t, source, 1)
	_, deps, _, err := readFinishedSource(library, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(deps.Assets) != 1 {
		t.Fatal("asset closure wrong", deps.Assets)
	}
	original := append([]byte(nil), mustRead(t, filepath.Join(library, deps.Assets["sample-image"].Path))...)
	targets := []*Project{reuseProject(t), reuseProject(t)}
	paths := []string{}
	for i, p := range targets {
		r, err := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "inserted-cover", Rationale: "Exact authored content copied for testing", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Library.AssetRemaps) != 1 {
			t.Fatal(r)
		}
		p, err = Load(p.Root)
		if err != nil {
			t.Fatal(err)
		}
		targets[i] = p
		asset := p.Document.Assets[r.Library.AssetRemaps["sample-image"]]
		paths = append(paths, filepath.Join(p.Root, asset.Path))
		if !bytes.Equal(mustRead(t, paths[i]), original) {
			t.Fatal("asset bytes adapted")
		}
		compiled, err := Check(p, bundle(t), wmdesign.CandidateEngine)
		if err != nil {
			t.Fatal(err)
		}
		native, _, err := wmdesign.BuildWithEngineAndAssets(bundle(t), "", compiled.Document, wmdesign.CandidateEngine, compiled.Assets)
		if err != nil || len(native) == 0 {
			t.Fatal("asset insertion cannot build", err)
		}
		inserted := p.Document.Slides[len(p.Document.Slides)-1]
		media, err := finishedMediaSlots(inserted, defs[inserted.Template.ID])
		if err != nil {
			t.Fatal(err)
		}
		hasRemap := false
		for _, value := range media {
			if value == "project:"+r.Library.AssetRemaps["sample-image"] {
				hasRemap = true
			}
		}
		if !hasRemap {
			t.Fatal("image slot not remapped")
		}
		businessCopy := false
		for _, value := range inserted.Values["slots"].(map[string]any) {
			if value == "sample-image" {
				businessCopy = true
			}
		}
		if !businessCopy {
			t.Fatal("business string interpreted as asset")
		}
	}
	if filepath.Base(paths[0]) == filepath.Base(paths[1]) {
		t.Fatal("asset IDs reused across projects")
	}
	if err := os.WriteFile(paths[0], []byte("edited asset in one deck"), 0644); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mustRead(t, paths[1]), original) || !bytes.Equal(mustRead(t, filepath.Join(library, deps.Assets["sample-image"].Path)), original) {
		t.Fatal("asset mutation leaked")
	}
}

func mustSourceMap(t *testing.T, def wmdesign.LibraryTemplate) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal(def.RawSlide, &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestObservedDependencyWritePreparationFailure(t *testing.T) {
	p := reuseProject(t)
	original := mustRead(t, p.SourcePath)
	logPath := filepath.Join(p.Root, "composition-log.yaml")
	observed := map[string][]byte{"composition-log.yaml": mustRead(t, logPath)}
	if err := os.WriteFile(filepath.Join(p.Root, "blocked-parent"), []byte("owned file remains"), 0644); err != nil {
		t.Fatal(err)
	}
	main, err := sourceYAML(p.Raw)
	if err != nil {
		t.Fatal(err)
	}
	title, err := editYAMLNode("Candidate title that must not activate")
	if err != nil {
		t.Fatal(err)
	}
	replaceMappingField(main.Content[0], "title", title)
	raw, err := encodeSourceYAML(main)
	if err != nil {
		t.Fatal(err)
	}
	_, err = commitSourceChangesObserved(p, map[string][]byte{filepath.Base(p.SourcePath): raw, "blocked-parent/new.json": []byte("must not activate")}, observed, nil)
	if err == nil {
		t.Fatal("invalid destination activated")
	}
	if !bytes.Equal(mustRead(t, p.SourcePath), original) || !bytes.Equal(mustRead(t, logPath), observed["composition-log.yaml"]) {
		t.Fatal("write preparation failure changed predecessors")
	}
	if matches, _ := filepath.Glob(filepath.Join(p.Root, ".source-write-*.tmp")); len(matches) != 0 {
		t.Fatal("temporary source file leaked", matches)
	}
	if _, err := os.Stat(filepath.Join(p.Root, ".deck-source-mutation.lock")); !os.IsNotExist(err) {
		t.Fatal("mutation guard leaked")
	}
}

func TestFinishedSlideRejectsResealedPinMismatch(t *testing.T) {
	p := reuseProject(t)
	library, m := publishReuse(t, p, 1)
	m.Pins.ToolchainLockSHA256 = strings.Repeat("c", 64)
	if err := m.Seal(); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(library, "manifest.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "fresh-message", Rationale: "Fixture insertion", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err == nil || !strings.Contains(err.Error(), "incompatible_pins") {
		t.Fatal("mismatched lock pin accepted", err)
	}
	if !bytes.Equal(mustRead(t, p.SourcePath), p.Raw) {
		t.Fatal("pin mismatch mutated source")
	}
}
