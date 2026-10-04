package wmdesign

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func indexFixture(t *testing.T) (string, LibraryIndexReport) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "catalog.sqlite")
	report, err := BuildLibraryIndex(path, LibraryIndexOptions{Bundle: filepath.Join("..", "..", "library", "wm-design-system", "v3")})
	if err != nil {
		t.Fatal(err)
	}
	return path, report
}

func TestUnifiedLibraryNativeProjectionAndDiscovery(t *testing.T) {
	path, report := indexFixture(t)
	for _, kind := range []string{"template", "primitive", "component", "composite", "frame", "asset"} {
		if report.Counts[kind] == 0 {
			t.Fatalf("missing %s projection", kind)
		}
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	entity, err := index.Inspect("cards/3")
	if err != nil {
		t.Fatal(err)
	}
	if entity.ID != "wmds/template/cards/3" || entity.Template == nil || len(entity.Definition) == 0 || len(entity.Capacity) == 0 || entity.SourceSHA256 == "" {
		t.Fatalf("incomplete entity: %+v", entity)
	}
	if _, err = index.db.Exec("DELETE FROM entities"); err == nil {
		t.Fatal("read-only connection accepted mutation")
	}
	opts := LibraryIndexFindOptions{Kinds: []string{"template"}, Shape: LibrarySearchOptions{Items: 3, ItemRole: "point", ContentRoles: []string{"point"}, Limit: 100, Query: "unrelated scenario"}}
	a, err := index.Find(opts)
	if err != nil {
		t.Fatal(err)
	}
	b, err := index.Find(opts)
	if err != nil {
		t.Fatal(err)
	}
	if string(indexJSON(a)) != string(indexJSON(b)) {
		t.Fatal("search not deterministic")
	}
	found := false
	for _, hit := range a.Matches {
		if hit.Entity.Key == "cards/3" && hit.CountMatch != nil && hit.CountMatch.Group.Scope == "primary" {
			found = true
		}
	}
	if !found {
		t.Fatal("primary cards lost from independent structural search")
	}
	if a.Matches[0].Score == 0 {
		t.Fatal("unknown scenario suppressed independent content hints")
	}
	injection, err := index.Find(LibraryIndexFindOptions{Kinds: []string{"template') OR 1=1 --"}})
	if err != nil || len(injection.Matches) != 0 {
		t.Fatalf("kind query was not parameterized: %v %+v", err, injection)
	}
	preview, err := index.Preview("cards/3")
	if err != nil || preview.Preparation == "" || len(preview.Definition) == 0 {
		t.Fatalf("missing honest no-preview fallback: %v %+v", err, preview)
	}
}

func TestUnifiedLibraryProjectionTamperingRejected(t *testing.T) {
	for _, query := range []string{"UPDATE entities SET name='changed' WHERE id='wmds/template/cards/3'", "UPDATE entities SET json=replace(json,'Three cards','Changed cards') WHERE id='wmds/template/cards/3'"} {
		t.Run(query, func(t *testing.T) {
			path, _ := indexFixture(t)
			db, err := indexDB(path, false)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(query, "json=") {
				query = "UPDATE entities SET json=json_set(json,'$.purpose','Changed purpose') WHERE id='wmds/template/cards/3'"
			}
			if _, err = db.Exec(query); err != nil {
				t.Fatal(err)
			}
			db.Close()
			index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
			if err == nil {
				index.Close()
				t.Fatal("modified projection accepted")
			}
		})
	}
}

func TestUnifiedLibraryLegacyNamespaceAndInputPin(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, "legacy.sqlite")
	if err := os.WriteFile(legacy, nil, 0600); err != nil {
		t.Fatal(err)
	}
	db, err := indexDB(legacy, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE inventory(id TEXT,kind TEXT,title TEXT,body TEXT,json TEXT,preference TEXT); CREATE TABLE contracts(id TEXT,kind TEXT,name TEXT,purpose TEXT,state TEXT,path TEXT,sha256 TEXT); INSERT INTO inventory VALUES('cards/3','slide','Legacy cards','example','{"id":"cards/3","review":"unknown"}','prefer'); INSERT INTO contracts VALUES('old','template','Old template','bounded source','reviewed','contracts/old.json','declared-hash');`)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	path := filepath.Join(root, "unified.sqlite")
	report, err := BuildLibraryIndex(path, LibraryIndexOptions{Bundle: filepath.Join("..", "..", "library", "wm-design-system", "v3"), LegacyIndex: legacy})
	if err != nil {
		t.Fatal(err)
	}
	if report.Counts["slide"] != 1 {
		t.Fatal("legacy inventory missing")
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	stored, e := index.entities("id=?", []any{"legacy/inventory/cards/3"})
	if e != nil || len(stored) != 1 || !strings.Contains(string(stored[0].Definition), "record_sha256") {
		t.Fatalf("legacy raw metadata duplicated in projection: %+v %v", stored, e)
	}
	full, e := index.Inspect("legacy/inventory/cards/3")
	if e != nil || !strings.Contains(string(full.Definition), `"review":"unknown"`) {
		t.Fatalf("lazy original definition not returned: %+v %v", full, e)
	}
	entity, err := index.Inspect("legacy/contract/old")
	if err != nil || entity.Discovery.Capability.SpecimenReview != "reviewed" {
		t.Fatalf("qualification lost: %+v %v", entity, err)
	}
	index.Close()
	db, err = indexDB(legacy, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("UPDATE inventory SET title='Changed'")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	index, err = OpenLibraryIndex(path, LibraryIndexOptions{})
	if err == nil {
		index.Close()
		t.Fatal("changed legacy input accepted")
	}
}

func TestUnifiedLibraryActualContentAlternatives(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	values := BoundCardsContent{Eyebrow: "Delivery", Title: "Keep review evidence and next actions together", Cards: []BoundCardContent{{Key: "source", Title: "Trace sources", Body: "Link every claim to its source."}, {Key: "review", Title: "Review clearly", Body: "Give reviewers the audience and visible draft."}, {Key: "retain", Title: "Retain context and next actions", Body: "Keep decisions with the source deck; name the next decision and accountable owner."}}}
	input := BoundDocument{Schema: BoundDocumentSchema, Year: 2026, Slides: []BoundSlide{{ID: "option-a", Template: "cards/3", ContentKind: "supplied_content", Values: indexJSON(values)}, {ID: "option-b", Template: "cards/4", ContentKind: "supplied_content", Values: indexJSON(BoundCardsContent{Eyebrow: values.Eyebrow, Title: values.Title, Cards: []BoundCardContent{values.Cards[0], values.Cards[1], {Key: "retain", Title: "Retain context", Body: "Keep decisions with the source deck."}, {Key: "next", Title: "Define the next action", Body: "Name the next decision and accountable owner."}}})}, {ID: "invalid-option", Template: "cards/3", ContentKind: "supplied_content", Values: indexJSON(BoundCardsContent{Title: values.Title, Cards: values.Cards[:2]})}}}
	out := filepath.Join(t.TempDir(), "alternatives")
	report, err := FitLibraryCandidates(bundle, "", CandidateEngine, out, input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed != 2 || report.Failed != 1 {
		t.Fatalf("unexpected actual-content fit: %+v", report)
	}
	for _, result := range report.Candidates[:2] {
		if _, err := os.Stat(result.Deck); err != nil {
			t.Fatal(err)
		}
		if result.Status != "go_layout_succeeded_native_review_pending" {
			t.Fatal(result.Status)
		}
	}
	var saved BoundDocument
	raw, err := os.ReadFile(filepath.Join(out, "candidate-content.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if string(saved.Slides[0].Values) != string(indexJSON(values)) { // formatting may differ; compare decoded value
		var got BoundCardsContent
		if err = json.Unmarshal(saved.Slides[0].Values, &got); err != nil || got.Title != values.Title {
			t.Fatal("actual copy not retained")
		}
	}
	if _, err = FitLibraryCandidates(bundle, "", CandidateEngine, out, input); err == nil {
		t.Fatal("existing output accepted")
	}
	input.Slides[0].ContentKind = "synthetic_example"
	if _, err = FitLibraryCandidates(bundle, "", CandidateEngine, filepath.Join(t.TempDir(), "rejected"), input); err == nil {
		t.Fatal("specimen fallback admitted")
	}
}

func TestUnifiedLibraryGalleryPinsAndPreviewDrift(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	catalog, err := LibraryCatalog(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	def := discoveryPinnedTemplate(t, catalog, "cards/3")
	root := t.TempDir()
	gallery := filepath.Join(root, "gallery")
	if err = os.MkdirAll(filepath.Join(gallery, "design-system"), 0755); err != nil {
		t.Fatal(err)
	}
	contract := "design-system/contract.json"
	preview := "design-system/source.png"
	if err = os.WriteFile(filepath.Join(gallery, contract), indexJSON(def), 0644); err != nil {
		t.Fatal(err)
	}
	// Same family file hash does not establish template identity.
	other := discoveryPinnedTemplate(t, catalog, "cards/4")
	if other.SourceSHA256 != def.SourceSHA256 {
		t.Fatal("fixture no longer shares family source")
	}
	if err = os.WriteFile(filepath.Join(gallery, contract), indexJSON(other), 0644); err != nil {
		t.Fatal(err)
	}
	wrongManifest := map[string]any{"source_revision": def.SourceRevision, "designs": []any{map[string]any{"template": def.Key, "contract": contract}}}
	if err = os.WriteFile(filepath.Join(gallery, "design-system/index.json"), indexJSON(wrongManifest), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = BuildLibraryIndex(filepath.Join(root, "wrong-contract.sqlite"), LibraryIndexOptions{Bundle: bundle, Gallery: gallery}); err == nil {
		t.Fatal("same-family swapped contract accepted")
	}
	if err = os.WriteFile(filepath.Join(gallery, contract), indexJSON(def), 0644); err != nil {
		t.Fatal(err)
	}
	bytes := []byte("pinned preview fixture")
	if err = os.WriteFile(filepath.Join(gallery, preview), bytes, 0644); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{"source_revision": def.SourceRevision, "designs": []any{map[string]any{"template": def.Key, "contract": contract, "source_preview": preview, "source_preview_sha256": indexDigest(bytes), "native_review": "fixture_only"}}}
	if err = os.WriteFile(filepath.Join(gallery, "design-system/index.json"), indexJSON(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "catalog.sqlite")
	if _, err = BuildLibraryIndex(path, LibraryIndexOptions{Bundle: bundle, Gallery: gallery}); err != nil {
		t.Fatal(err)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	p, err := index.Preview("cards/3")
	if err != nil || len(p.Paths) != 1 || p.Capability.SpecimenReview != "fixture_only" {
		t.Fatalf("pinned preview not imported: %v %+v", err, p)
	}
	card, err := index.SelectionCard("cards/3")
	if err != nil || len(card.ScreenshotPaths) != 1 || card.ScreenshotPaths[0] != p.Paths[0] {
		t.Fatalf("selection card lost verified preview: %v %+v", err, card)
	}
	if err = os.WriteFile(filepath.Join(gallery, preview), []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = index.Preview("cards/3"); err == nil {
		t.Fatal("changed preview accepted")
	}
	if _, err = index.SelectionCard("cards/3"); err == nil {
		t.Fatal("selection card accepted changed screenshot")
	}
	manifest["source_revision"] = "different"
	if err = os.WriteFile(filepath.Join(gallery, "design-system/index.json"), indexJSON(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = BuildLibraryIndex(filepath.Join(root, "wrong.sqlite"), LibraryIndexOptions{Bundle: bundle, Gallery: gallery}); err == nil {
		t.Fatal("mismatched gallery revision accepted")
	}
}

func TestUnifiedLibraryLegacyResourceRelocation(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, "library", "catalog.sqlite")
	if err := os.MkdirAll(filepath.Dir(legacy), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "library", "contracts"), 0755); err != nil {
		t.Fatal(err)
	}
	contract := []byte(`{"id":"old","qualification":{"state":"reviewed"},"roles":["point"]}`)
	if err := os.WriteFile(filepath.Join(root, "library", "contracts", "old.json"), contract, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, nil, 0600); err != nil {
		t.Fatal(err)
	}
	db, err := indexDB(legacy, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE inventory(id TEXT,kind TEXT,title TEXT,body TEXT,json TEXT,preference TEXT);CREATE TABLE contracts(id TEXT,kind TEXT,name TEXT,purpose TEXT,state TEXT,path TEXT,sha256 TEXT)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO contracts VALUES(?,?,?,?,?,?,?)", "old", "template", "Old", "purpose", "reviewed", "library/contracts/old.json", indexDigest(contract))
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	path := filepath.Join(root, "unified.sqlite")
	if _, err = BuildLibraryIndex(path, LibraryIndexOptions{Bundle: filepath.Join("..", "..", "library", "wm-design-system", "v3"), LegacyIndex: legacy}); err != nil {
		t.Fatal(err)
	}
	relocated := t.TempDir()
	if err = os.MkdirAll(filepath.Join(relocated, "samples", "showcase"), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Join(relocated, "library", "contracts"), 0755); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(legacy)
	if err != nil {
		t.Fatal(err)
	}
	newDB := filepath.Join(relocated, "samples", "showcase", "catalog.sqlite")
	if err = os.WriteFile(newDB, raw, 0644); err != nil {
		t.Fatal(err)
	}
	newContract := filepath.Join(relocated, "library", "contracts", "old.json")
	if err = os.WriteFile(newContract, contract, 0644); err != nil {
		t.Fatal(err)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{LegacyIndex: newDB, LegacyRoot: relocated})
	if err != nil {
		t.Fatal(err)
	}
	entity, err := index.Inspect("legacy/contract/old")
	if err != nil || !strings.Contains(string(entity.Definition), `"contract":`) {
		t.Fatalf("full verified contract missing: %+v %v", entity, err)
	}
	index.Close()
	if err = os.WriteFile(newContract, []byte(`{"id":"changed"}`), 0644); err != nil {
		t.Fatal(err)
	}
	index, err = OpenLibraryIndex(path, LibraryIndexOptions{LegacyIndex: newDB, LegacyRoot: relocated})
	if err == nil {
		index.Close()
		t.Fatal("relocated contract resource drift accepted")
	}
}
