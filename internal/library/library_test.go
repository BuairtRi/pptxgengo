package library

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0644); err != nil {
		t.Fatal(err)
	}
}
func TestPointerAndBoundedRepeat(t *testing.T) {
	var slide any
	json.Unmarshal([]byte(`{"paths":[{"labels":["old"]}],"pods":[{"roles":[{"id":"old","label":"old"}]}]}`), &slide)
	if err := pointerSet(slide, "/paths/0/labels", []string{"one", "two"}); err != nil {
		t.Fatal(err)
	}
	slot := Slot{ItemTemplate: json.RawMessage(`{"id":"x","label":"x"}`), ItemValuePointer: "/label", ItemIDPointer: "/id", ItemIDPrefix: "role-"}
	v, err := repeatValue(slot, []string{"Lead", "Engineer"})
	if err != nil {
		t.Fatal(err)
	}
	if err := pointerSet(slide, "/pods/0/roles", v); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(slide)
	if !strings.Contains(string(b), `"role-2"`) || !strings.Contains(string(b), `"two"`) {
		t.Fatalf("repeat binding failed: %s", b)
	}
	if validContentPointer("/canvas/0/bounds/x") {
		t.Fatal("geometry pointer accepted")
	}
}

func TestStyleTokensOnlyTouchColorFields(t *testing.T) {
	var slide any
	json.Unmarshal([]byte(`{"id":"brand.ink","title":"brand.ink","canvas":[{"text":"brand.ink","foreground":"brand.ink","background":"brand.surface","asset_path":"brand.ink"}]}`), &slide)
	replaceTokens(slide, map[string]string{"brand.ink": "#070154", "brand.surface": "#E8EEF8"})
	b, _ := json.Marshal(slide)
	if !strings.Contains(string(b), `"text":"brand.ink"`) || !strings.Contains(string(b), `"id":"brand.ink"`) || !strings.Contains(string(b), `"asset_path":"brand.ink"`) || !strings.Contains(string(b), `"foreground":"#070154"`) {
		t.Fatalf("style changed content/asset identity: %s", b)
	}
}
func TestContractRejectsStaleProofAndUnboundedArray(t *testing.T) {
	root := t.TempDir()
	s, _ := NewStore(root)
	template := []byte(`{"schema":"pptxgengo.compose-spec.v1","slides":[{"id":"s","paths":[{"labels":["a"]}]}]}`)
	put(t, filepath.Join(root, "template.json"), template)
	c := Contract{Schema: ContractSchema, ID: "test", Version: "1.0.0", Kind: "recipe", Name: "Test", Purpose: "A test", Source: Source{SourceID: "synthetic", SourceSHA256: strings.Repeat("a", 64), Slide: 1}, Composition: Composition{SpecPath: "template.json", SpecSHA256: hashBytes(template), SlideID: "s", Slots: []Slot{{Name: "nodes", Role: "process_steps", Pointer: "/paths/0/labels", ValueType: "string_array", Required: true, MaxItems: 0, MaxChars: 32}}}, FitEnvelope: FitEnvelope{FontPolicy: "no_silent_shrink"}, Transforms: Transforms{Translation: "tested", Resize: "unsupported", Rotation: "unsupported"}, Preference: Preference{Value: "unreviewed"}, Qualification: Qualification{State: "measured_fixture"}}
	if err := s.ValidateContract(c); err == nil || !strings.Contains(err.Error(), "max_items") {
		t.Fatalf("unbounded array accepted: %v", err)
	}
	c.Composition.Slots[0].MaxItems = 6
	source := []byte("Synthetic source")
	put(t, filepath.Join(root, "source.txt"), source)
	c.Source.Path = "source.txt"
	c.Source.SourceSHA256 = hashBytes(source)
	c.Qualification.State = "adaptation_qualified"
	if err := s.ValidateContract(c); err == nil || !strings.Contains(err.Error(), "proof") {
		t.Fatalf("unproven contract qualified: %v", err)
	}
	c.Qualification.State = "measured_fixture"
	c.Composition.SpecSHA256 = strings.Repeat("b", 64)
	if err := s.ValidateContract(c); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale template accepted: %v", err)
	}
}
func TestNarrativeQuoteMustMatchHashedSource(t *testing.T) {
	root := t.TempDir()
	s, _ := NewStore(root)
	src := []byte("Exact attributed line in an explicitly synthetic source packet.")
	put(t, filepath.Join(root, "source.txt"), src)
	n := Narrative{Schema: NarrativeSchema, Brief: NarrativeBrief{Name: "Synthetic", Synthetic: true, Audience: "Leaders", Decision: "Review"}, Slides: []NarrativeSlide{{ID: "s", Audience: "Leaders", Role: "evidence", Takeaway: "A claim", AssertionTitle: "A claim", VisualRelationship: "quote"}}, Evidence: []Evidence{{ID: "e", Source: Artifact{Path: "source.txt", SHA256: hashBytes(src)}, Locator: "line 1", Quote: "Exact attributed line"}}}
	b, _ := json.Marshal(n)
	put(t, filepath.Join(root, "narrative.json"), b)
	if _, err := s.LoadNarrative(filepath.Join(root, "narrative.json")); err != nil {
		t.Fatal(err)
	}
	n.Evidence[0].Quote = "Invented words"
	b, _ = json.Marshal(n)
	put(t, filepath.Join(root, "narrative.json"), b)
	if _, err := s.LoadNarrative(filepath.Join(root, "narrative.json")); err == nil || !strings.Contains(err.Error(), "exact quote") {
		t.Fatalf("false quote accepted: %v", err)
	}
}
func TestInventoryFindUsesPinnedCatalogAndPreferenceOverlay(t *testing.T) {
	// The retired production inventory is not a dependency of the latest-only
	// checkout. This small, explicitly synthetic inventory exercises the same
	// pin, projection, preference and retrieval behavior without a prior release.
	root := t.TempDir()
	catalog := filepath.Join(root, "catalog.sqlite")
	if _, err := sqlite(catalog, "CREATE TABLE items(id TEXT PRIMARY KEY,kind TEXT,source_id TEXT,category TEXT,title TEXT,body TEXT,json TEXT); INSERT INTO items VALUES('synthetic-governance','component','synthetic-fixture','controls','Governance controls','Record decision ownership and evidence','{}'),('synthetic-planning','component','synthetic-fixture','planning','Planning sequence','Prepare a synthetic next step','{}');", false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(catalog)
	if err != nil {
		t.Fatal(err)
	}
	pin := Artifact{Path: "catalog.sqlite", SHA256: hashBytes(data)}
	manifest, err := json.Marshal(map[string]any{"artifacts": map[string]Artifact{"sqlite": pin}})
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "library/catalog-manifest.json"), manifest)
	shortlist := []byte(`{"references":[]}`)
	put(t, filepath.Join(root, "library/reference-shortlist.json"), shortlist)
	preferences, err := json.Marshal(map[string]any{"shortlist_sha256": hashBytes(shortlist), "choices": []prefChoice{{ComponentID: "synthetic-governance", Preference: "avoid"}}})
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "library/reference-preferences.json"), preferences)
	s, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	s.CatalogPath = catalog
	index := filepath.Join(t.TempDir(), "index.sqlite")
	report, err := s.BuildIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	if report.InventoryCount != 2 || report.ContractCount != 0 {
		t.Fatalf("synthetic inventory projection changed: %+v", report)
	}
	if report.CatalogSHA256 != pin.SHA256 {
		t.Fatalf("inventory source pin changed: %+v", report)
	}
	hits, err := s.Find(index, FindOptions{Query: "governance", Inventory: true, IncludeAvoid: true, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ID != "synthetic-governance" || hits[0].State != "inventory" {
		t.Fatalf("inventory smoke failed: %+v", hits)
	}
	if hits, err := s.Find(index, FindOptions{Query: "governance", Inventory: true}); err != nil || len(hits) != 0 {
		t.Fatalf("avoid overlay did not guide default retrieval: %+v %v", hits, err)
	}
	if hits, err := s.Find(index, FindOptions{Query: "governance", IncludeAvoid: true}); err != nil || len(hits) != 0 {
		t.Fatalf("inventory occurrence promoted into qualified contract search: %+v %v", hits, err)
	}
	// Preferences guide retrieval but do not promote source occurrences.
	b, err := sqlite(index, "SELECT preference FROM inventory WHERE id='synthetic-governance';", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "avoid") {
		t.Fatalf("preference overlay missing: %s", b)
	}
	if err := os.WriteFile(catalog, append(data, []byte("drift")...), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BuildIndex(filepath.Join(t.TempDir(), "drift.sqlite")); err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("catalog pin drift accepted: %v", err)
	}
}

func TestIndexedContractInstantiateBoundedArray(t *testing.T) {
	root := t.TempDir()
	s, _ := NewStore(root)
	catalog := filepath.Join(root, "catalog.sqlite")
	if _, err := sqlite(catalog, "CREATE TABLE items(id TEXT PRIMARY KEY,kind TEXT,source_id TEXT,category TEXT,title TEXT,body TEXT,json TEXT); INSERT INTO items VALUES('item','component','synthetic','','Item','path','{}');", false); err != nil {
		t.Fatal(err)
	}
	cb, _ := os.ReadFile(catalog)
	manifest := map[string]any{"artifacts": map[string]any{"sqlite": Artifact{Path: "catalog.sqlite", SHA256: hashBytes(cb)}}}
	mb, _ := json.Marshal(manifest)
	put(t, filepath.Join(root, "library/catalog-manifest.json"), mb)
	short := []byte(`{"references":[]}`)
	put(t, filepath.Join(root, "library/reference-shortlist.json"), short)
	prefs := map[string]any{"shortlist_sha256": hashBytes(short), "choices": []any{}}
	pb, _ := json.Marshal(prefs)
	put(t, filepath.Join(root, "library/reference-preferences.json"), pb)
	source := []byte("Synthetic process fixture")
	put(t, filepath.Join(root, "source.txt"), source)
	template := []byte(`{"schema":"pptxgengo.compose-spec.v1","slides":[{"id":"s","title":"A governed process","role":"Process","takeaway":"Three steps","paths":[{"labels":["old","old"]}]}]}`)
	put(t, filepath.Join(root, "template.json"), template)
	c := Contract{Schema: ContractSchema, ID: "process", Version: "1.0.0", Kind: "recipe", Name: "Process path", Purpose: "Explain a path", ContentRoles: []string{"process_steps"}, Source: Source{SourceID: "synthetic", Path: "source.txt", SourceSHA256: hashBytes(source), Slide: 1}, Composition: Composition{SpecPath: "template.json", SpecSHA256: hashBytes(template), SlideID: "s", Slots: []Slot{{Name: "steps", Role: "process_steps", Pointer: "/paths/0/labels", ValueType: "string_array", Required: true, MinItems: 2, MaxItems: 6, MaxChars: 32}}}, FitEnvelope: FitEnvelope{FontPolicy: "no_silent_shrink"}, Transforms: Transforms{Translation: "tested", Resize: "unsupported", Rotation: "unsupported"}, Preference: Preference{Value: "preferred"}, Qualification: Qualification{State: "measured_fixture"}}
	contractBytes, _ := json.Marshal(c)
	put(t, filepath.Join(root, "library/contracts/process.json"), contractBytes)
	s.CatalogPath = catalog
	idx := filepath.Join(root, "new-index.sqlite")
	if _, err := s.BuildIndex(idx); err != nil {
		t.Fatal(err)
	}
	hits, err := s.Find(idx, FindOptions{Query: "path", State: "measured_fixture"})
	if err != nil || len(hits) != 1 || hits[0].ID != "process" {
		t.Fatalf("find: %+v %v", hits, err)
	}
	values := []byte(`{"slots":{"steps":["Capture","Review","Publish"]}}`)
	put(t, filepath.Join(root, "values.json"), values)
	if _, err := s.Instantiate(idx, "process", filepath.Join(root, "values.json"), filepath.Join(root, "blocked"), false); err == nil || !strings.Contains(err.Error(), "allow-unqualified") {
		t.Fatalf("unqualified instantiation accepted: %v", err)
	}
	trace, err := s.Instantiate(idx, "process", filepath.Join(root, "values.json"), filepath.Join(root, "output"), true)
	if err != nil {
		t.Fatal(err)
	}
	if trace.FitStatus != "unmeasured_changed_copy" {
		t.Fatalf("fit status overstated: %+v", trace)
	}
	b, err := os.ReadFile(filepath.Join(root, "output/spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Publish") || strings.Contains(string(b), `"old"`) {
		t.Fatalf("semantic slot not applied: %s", b)
	}
	narr := Narrative{Schema: NarrativeSchema, Brief: NarrativeBrief{Name: "Synthetic process", Synthetic: true, Audience: "Leaders", Decision: "Review", SourcePacket: []Artifact{{Path: "source.txt", SHA256: hashBytes(source)}}}, Slides: []NarrativeSlide{{ID: "story", Audience: "Leaders", Role: "Process", Takeaway: "Three steps", AssertionTitle: "A governed process", RequiredDetail: []string{"Capture"}, VisualRelationship: "flow"}}}
	narrBytes, _ := json.Marshal(narr)
	put(t, filepath.Join(root, "narrative.json"), narrBytes)
	values, err = json.Marshal(map[string]any{"slots": map[string]any{"steps": []string{"Capture", "Review", "Publish"}}, "narrative_path": filepath.Join(root, "narrative.json"), "narrative_slide_id": "story"})
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "narrative-values.json"), values)
	trace, err = s.Instantiate(idx, "process", filepath.Join(root, "narrative-values.json"), filepath.Join(root, "narrative-output"), true)
	if err != nil {
		t.Fatal(err)
	}
	if trace.NarrativeSHA256 != hashBytes(narrBytes) || len(trace.NarrativeSources) == 0 {
		t.Fatalf("narrative trace missing proof: %+v", trace)
	}
	narr.Slides[0].AssertionTitle = "Unrelated title"
	narrBytes, _ = json.Marshal(narr)
	put(t, filepath.Join(root, "narrative.json"), narrBytes)
	if _, err = s.Instantiate(idx, "process", filepath.Join(root, "narrative-values.json"), filepath.Join(root, "mismatch-output"), true); err == nil || !strings.Contains(err.Error(), "narrative assertion_title") {
		t.Fatalf("unrelated narrative accepted: %v", err)
	}
}
