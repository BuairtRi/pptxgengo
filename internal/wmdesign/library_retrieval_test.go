package wmdesign

import (
	"strings"
	"sync"
	"testing"
)

func TestKeywordConcurrentReadOnlyQueries(t *testing.T) {
	path, _ := indexFixture(t)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	opts := LibraryIndexFindOptions{Retrieval: "keyword", Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: "cards/3", Limit: 5}}
	want, err := index.Find(opts)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Go(func() {
			for j := 0; j < 3; j++ {
				got, err := index.Find(opts)
				if err != nil || string(indexJSON(got)) != string(indexJSON(want)) {
					t.Errorf("concurrent query changed results: %v", err)
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestKeywordRetrievalKeepsRankingAndShapeSeparate(t *testing.T) {
	path, _ := indexFixture(t)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	options := LibraryIndexFindOptions{Retrieval: "keyword", Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: "cards/3", Items: 4, ItemRole: "point", Limit: 100}}
	r, err := index.Find(options)
	if err != nil || len(r.Matches) == 0 {
		t.Fatalf("%v %+v", err, r)
	}
	if r.Retrieval.Actual != "keyword" || r.Matches[0].Entity.Key != "cards/3" || r.Matches[0].Retrieval.Rank != 1 {
		t.Fatalf("unexpected exact-key result: %+v", r)
	}
	if r.Matches[0].StructuralStatus != "source_hints_mismatch" || len(r.Matches[0].UnmatchedHints) == 0 || r.Matches[0].FitStatus != "not_measured_for_query" {
		t.Fatalf("lexical relevance hid count mismatch: %+v", r.Matches[0])
	}
	last := -1.0e100
	for _, h := range r.Matches {
		if h.ScenarioScore != 0 || (h.Retrieval.Method != "exact_entity_identity" && h.Retrieval.BM25 < last) {
			t.Fatalf("mixed scoring or wrong BM25 ordering: %+v", h)
		}
		if h.Retrieval.Method != "exact_entity_identity" {
			last = h.Retrieval.BM25
		}
	}
	options.RequireShape = true
	r, err = index.Find(options)
	if err != nil || len(r.Matches) == 0 {
		t.Fatalf("%v %+v", err, r)
	}
	for _, h := range r.Matches {
		if h.Entity.Key == "cards/3" || h.StructuralStatus != "source_hints_match_not_measured_fit" || h.CountMatch == nil {
			t.Fatalf("hard shape requirement ignored: %+v", h)
		}
	}
}

func TestKeywordLiteralQuerySafetyAndDeterminism(t *testing.T) {
	for query, expected := range map[string]string{"cards cards/3": `"cards" OR "3"`, `"*" OR name:cards NOT ():3`: `"or" OR "name" OR "cards" OR "not" OR "3"`, "() * - :": "", "café 東京": `"café" OR "東京"`} {
		actual, err := libraryFTSQuery(query)
		if err != nil || actual != expected {
			t.Fatalf("%q -> %q: %v; want %q", query, actual, err, expected)
		}
	}
	if _, err := libraryFTSQuery(strings.Repeat("x", 4097)); err == nil {
		t.Fatal("unbounded query accepted")
	}
	path, _ := indexFixture(t)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	for _, query := range []string{"cards cards/3", `cards' OR 1=1 --`, "() * - :", "qzxunknown"} {
		opts := LibraryIndexFindOptions{Retrieval: "keyword", Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: query}}
		a, err := index.Find(opts)
		if err != nil {
			t.Fatal(query, err)
		}
		b, err := index.Find(opts)
		if err != nil || string(indexJSON(a)) != string(indexJSON(b)) {
			t.Fatal("nondeterministic result", err)
		}
		if (query == "qzxunknown" || query == "() * - :") && len(a.Matches) != 0 {
			t.Fatal("unmatched query returned strong results", a.Matches)
		}
	}
	if _, err := index.Find(LibraryIndexFindOptions{Retrieval: "semantic", Shape: LibrarySearchOptions{Query: "cards"}}); err == nil {
		t.Fatal("semantic ranking silently accepted missing offline resources")
	}
	r, err := index.Find(LibraryIndexFindOptions{Retrieval: "keyword", Shape: LibrarySearchOptions{Items: 3}})
	if err != nil || r.Retrieval.Actual != "metadata" || r.Retrieval.Notice == "" {
		t.Fatal("empty-query browsing misreported", err, r.Retrieval)
	}
}

func TestKeywordCorpusIncludesAliasesButNotSyntheticCopy(t *testing.T) {
	entity := LibraryEntity{ID: "wmds/template/test", Key: "test", Template: &LibraryTemplate{Uses: []string{"stakeholder interviews"}, Authoring: &LibraryAuthoring{Relationship: "parallel ideas", Groups: []LibraryAuthoringGroup{{Alias: "practice-map", Role: "tabular-data"}}, Slots: []LibraryAuthoringSlot{{Alias: "findings", Description: "interview conclusions"}}}}, Definition: []byte(`{"source_slide":{"title":"SYNTHETIC_COPY_SHOULD_NOT_BE_INDEXED"}}`)}
	body := libraryRetrievalBody(entity)
	for _, term := range []string{"practice-map", "parallel ideas", "stakeholder interviews", "interview conclusions"} {
		if !strings.Contains(body, term) {
			t.Fatal("missing deliberate discovery text", term, body)
		}
	}
	if strings.Contains(body, "SYNTHETIC_COPY") {
		t.Fatal("synthetic specimen prose indexed")
	}
}

func TestDiscoveryReadKeepsVerifiedMetadataWithoutSourceDefinitions(t *testing.T) {
	path, _ := indexFixture(t)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	full, err := index.entities("kind='template'", nil)
	if err != nil {
		t.Fatal(err)
	}
	compact, err := index.discoveryEntities("kind='template'", nil)
	if err != nil || len(full) != len(compact) {
		t.Fatal(err)
	}
	for i, a := range full {
		b := compact[i]
		if len(b.Definition) != 0 || len(b.Discovery.Zones) != 0 {
			t.Fatal("source definition loaded for ranking", b.ID)
		}
		if a.ID != b.ID || a.Name != b.Name || a.Purpose != b.Purpose || a.Family != b.Family || a.Lifecycle != b.Lifecycle || a.SourceSHA256 != b.SourceSHA256 || string(indexJSON(a.Template.Uses)) != string(indexJSON(b.Template.Uses)) {
			t.Fatal("compact discovery changed identity/scenario", a.ID)
		}
		a.Discovery.Zones = nil
		if string(indexJSON(a.Discovery)) != string(indexJSON(b.Discovery)) || string(a.Capacity) != string(b.Capacity) || string(indexJSON(a.Artifacts)) != string(indexJSON(b.Artifacts)) {
			t.Fatal("compact discovery changed source affordances", a.ID)
		}
	}
}

func TestKeywordIndexDriftRejectedAndLegacyMetadataStillWorks(t *testing.T) {
	for _, mutation := range []string{"UPDATE entity_fts SET body='tampered' WHERE id='wmds/template/cards/3'", "DELETE FROM entity_fts WHERE id='wmds/template/cards/3'", "INSERT INTO entity_fts VALUES('unexpected','','','')"} {
		t.Run(mutation, func(t *testing.T) {
			path, _ := indexFixture(t)
			db, err := indexDB(path, false)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec(mutation); err != nil {
				db.Close()
				t.Fatal(err)
			}
			db.Close()
			index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
			if err == nil {
				index.Close()
				t.Fatal("altered keyword corpus accepted")
			}
		})
	}
	path, _ := indexFixture(t)
	db, err := indexDB(path, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("UPDATE meta SET value=json_remove(value,'$.retrieval_text','$.retrieval_sha256') WHERE key='report'")
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	if _, err = index.Find(LibraryIndexFindOptions{Shape: LibrarySearchOptions{Query: "cards"}}); err != nil {
		t.Fatal("legacy metadata discovery broke", err)
	}
	if _, err = index.Find(LibraryIndexFindOptions{Retrieval: "keyword", Shape: LibrarySearchOptions{Query: "cards"}}); err == nil || !strings.Contains(err.Error(), "keyword_rebuild_required") {
		t.Fatal("stale text preparation accepted", err)
	}
}
