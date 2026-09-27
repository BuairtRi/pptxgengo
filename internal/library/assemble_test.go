package library

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assemblyFixture(t *testing.T) (Store, string, string) {
	t.Helper()
	root := t.TempDir()
	s, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog := filepath.Join(root, "catalog.sqlite")
	if _, err := sqlite(catalog, "CREATE TABLE items(id TEXT PRIMARY KEY,kind TEXT,source_id TEXT,category TEXT,title TEXT,body TEXT,json TEXT); INSERT INTO items VALUES('item','component','synthetic','','Item','path','{}');", false); err != nil {
		t.Fatal(err)
	}
	cb, _ := os.ReadFile(catalog)
	mb, _ := json.Marshal(map[string]any{"artifacts": map[string]any{"sqlite": Artifact{Path: "catalog.sqlite", SHA256: hashBytes(cb)}}})
	put(t, filepath.Join(root, "library/catalog-manifest.json"), mb)
	short := []byte(`{"references":[]}`)
	put(t, filepath.Join(root, "library/reference-shortlist.json"), short)
	pb, _ := json.Marshal(map[string]any{"shortlist_sha256": hashBytes(short), "choices": []any{}})
	put(t, filepath.Join(root, "library/reference-preferences.json"), pb)
	source := []byte("Synthetic source text")
	put(t, filepath.Join(root, "source.txt"), source)
	template := []byte(`{"schema":"pptxgengo.compose-spec.v1","slides":[{"id":"source-slide","title":"Old title","role":"proposal","takeaway":"Synthetic takeaway","page":"old"}]}`)
	put(t, filepath.Join(root, "template.json"), template)
	c := Contract{Schema: ContractSchema, ID: "example", Version: "1.0.0", Kind: "recipe", Name: "Example", Purpose: "A page", Source: Source{SourceID: "synthetic", Path: "source.txt", SourceSHA256: hashBytes(source), Slide: 1}, Composition: Composition{SpecPath: "template.json", SpecSHA256: hashBytes(template), SlideID: "source-slide", PaginationBinding: "/page", Slots: []Slot{{Name: "title", Role: "assertion", Pointer: "/title", ValueType: "string", Required: true, MaxChars: 100}}}, FitEnvelope: FitEnvelope{FontPolicy: "no_silent_shrink"}, Transforms: Transforms{Translation: "tested", Resize: "unsupported", Rotation: "unsupported"}, Preference: Preference{Value: "preferred"}, Qualification: Qualification{State: "measured_fixture"}}
	contractBytes, _ := json.Marshal(c)
	put(t, filepath.Join(root, "library/contracts/example.json"), contractBytes)
	s.CatalogPath = catalog
	index := filepath.Join(root, "index.sqlite")
	if _, err := s.BuildIndex(index); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "first.json"), []byte(`{"slots":{"title":"First argument"}}`))
	put(t, filepath.Join(root, "second.json"), []byte(`{"slots":{"title":"Second argument"}}`))
	config := AssemblyConfig{Schema: AssemblySchema, Slides: []AssemblySlideInput{{ContractID: "example", ValuesPath: "first.json", SlideID: "first"}, {ContractID: "example", ValuesPath: "second.json", SlideID: "second"}}}
	b, _ := json.Marshal(config)
	configPath := filepath.Join(root, "assembly.json")
	put(t, configPath, b)
	return s, index, configPath
}

func TestAssembleDeterministicRepeatedTemplateAndPage(t *testing.T) {
	s, index, config := assemblyFixture(t)
	root := filepath.Dir(config)
	first, err := s.Assemble(index, config, filepath.Join(root, "out-one"), true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Assemble(index, config, filepath.Join(root, "out-two"), true)
	if err != nil {
		t.Fatal(err)
	}
	b1, _ := os.ReadFile(filepath.Join(root, "out-one/spec.json"))
	b2, _ := os.ReadFile(filepath.Join(root, "out-two/spec.json"))
	if !bytes.Equal(b1, b2) || first.SpecSHA256 != second.SpecSHA256 {
		t.Fatal("assembly output is not deterministic")
	}
	var spec struct {
		Slides []struct{ ID, Title, Page string } `json:"slides"`
	}
	if err := json.Unmarshal(b1, &spec); err != nil {
		t.Fatal(err)
	}
	if len(spec.Slides) != 2 || spec.Slides[0].ID != "first" || spec.Slides[1].ID != "second" || spec.Slides[0].Page != "1" || spec.Slides[1].Page != "2" || spec.Slides[0].Title != "First argument" {
		t.Fatalf("wrong assembly: %+v", spec)
	}
	if len(first.Slides) != 2 || first.Slides[0].Selection.ContractSHA256 == "" || first.Slides[0].ValuesSHA256 == "" {
		t.Fatalf("lineage missing: %+v", first)
	}
}

func TestAssembleRejectsDuplicateExistingAndNarrativeMismatchAtomically(t *testing.T) {
	s, index, configPath := assemblyFixture(t)
	root := filepath.Dir(configPath)
	configBytes, _ := os.ReadFile(configPath)
	var config AssemblyConfig
	json.Unmarshal(configBytes, &config)
	config.Slides[1].SlideID = "first"
	b, _ := json.Marshal(config)
	put(t, configPath, b)
	if _, err := s.Assemble(index, configPath, filepath.Join(root, "duplicate"), true); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate accepted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "duplicate")); !os.IsNotExist(err) {
		t.Fatal("partial duplicate output")
	}
	config.Slides[1].SlideID = "second"
	narr := Narrative{Schema: NarrativeSchema, Brief: NarrativeBrief{Name: "Synthetic", Synthetic: true, Audience: "Leaders", Decision: "Review"}, Slides: []NarrativeSlide{{ID: "first", Audience: "Leaders", Role: "proposal", Takeaway: "Synthetic takeaway", AssertionTitle: "Wrong title", VisualRelationship: "sequence"}, {ID: "second", Audience: "Leaders", Role: "proposal", Takeaway: "Synthetic takeaway", AssertionTitle: "Second argument", VisualRelationship: "sequence"}}}
	nb, _ := json.Marshal(narr)
	put(t, filepath.Join(root, "narrative.json"), nb)
	config.NarrativePath = "narrative.json"
	b, _ = json.Marshal(config)
	put(t, configPath, b)
	if _, err := s.Assemble(index, configPath, filepath.Join(root, "bad-narrative"), true); err == nil || !strings.Contains(err.Error(), "narrative assertion_title") {
		t.Fatalf("narrative mismatch accepted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "bad-narrative")); !os.IsNotExist(err) {
		t.Fatal("partial narrative output")
	}
	narr.Slides[0].AssertionTitle = "First argument"
	nb, _ = json.Marshal(narr)
	put(t, filepath.Join(root, "narrative.json"), nb)
	withNarrative, err := s.Assemble(index, configPath, filepath.Join(root, "good-narrative"), true)
	if err != nil || withNarrative.NarrativeSHA256 != hashBytes(nb) || withNarrative.Slides[0].Selection.NarrativeSlideID != "first" {
		t.Fatalf("valid narrative not traced: %+v %v", withNarrative, err)
	}
	config.NarrativePath = ""
	b, _ = json.Marshal(config)
	put(t, configPath, b)
	put(t, filepath.Join(root, "existing/sentinel"), []byte("preserve"))
	if _, err := s.Assemble(index, configPath, filepath.Join(root, "existing"), true); err == nil || !strings.Contains(err.Error(), "output exists") {
		t.Fatalf("existing output accepted: %v", err)
	}
	sentinel, _ := os.ReadFile(filepath.Join(root, "existing/sentinel"))
	if string(sentinel) != "preserve" {
		t.Fatal("existing output modified")
	}
	if _, err := s.Assemble(index, configPath, filepath.Join(root, "unqualified"), false); err == nil || !strings.Contains(err.Error(), "allow-unqualified") {
		t.Fatalf("unqualified contract accepted: %v", err)
	}
	config.Slides[1].ContractID = "unknown"
	b, _ = json.Marshal(config)
	put(t, configPath, b)
	if _, err := s.Assemble(index, configPath, filepath.Join(root, "unknown"), true); err == nil || !strings.Contains(err.Error(), "not indexed") {
		t.Fatalf("unknown contract accepted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "unknown")); !os.IsNotExist(err) {
		t.Fatal("partial unknown-contract output")
	}
	config.Slides[1].ContractID = "example"
	b, _ = json.Marshal(config)
	put(t, configPath, b)
	contractPath := filepath.Join(root, "library/contracts/example.json")
	contractBytes, _ := os.ReadFile(contractPath)
	put(t, contractPath, append(contractBytes, '\n'))
	if _, err := s.Assemble(index, configPath, filepath.Join(root, "stale"), true); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale contract accepted: %v", err)
	}
}
