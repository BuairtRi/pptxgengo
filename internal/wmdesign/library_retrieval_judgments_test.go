package wmdesign

import (
	"os"
	"path/filepath"
	"testing"
)

// Engineering relevance judgments from the pinned source purposes. These are
// lexical baselines, not an operator-approved semantic/native qualification set.
var keywordJudgments = []struct {
	query      string
	acceptable []string
}{
	{"cards/3", []string{"cards/3"}},
	{"interview lists", []string{"interviews-summary/list-split", "interviews-detail/portrait-column"}},
	{"practices heat maps", []string{"heat-tile-map/plain", "heat-ref/cards-split", "heat-ref/locator-split"}},
	{"modernization economics", []string{"decision/buy-build-economics"}},
	{"roadmap", []string{"roadmap/staggered-phases", "roadmap/now-next-later", "roadmap/dependencies"}},
	{"pillars", []string{"pillars/three-why-matters", "pillars/two-categories-six"}},
}

func keywordV11Index(t testing.TB) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "discovery.sqlite")
	_, err := BuildLibraryIndex(path, LibraryIndexOptions{Bundle: filepath.Join("..", "..", "library", "wm-design-system", "v11")})
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestKeywordPinnedV11EngineeringJudgments(t *testing.T) {
	path := keywordV11Index(t)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	for _, judgment := range keywordJudgments {
		t.Run(judgment.query, func(t *testing.T) {
			result, err := index.Find(LibraryIndexFindOptions{Retrieval: "keyword", Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: judgment.query, Limit: 10}})
			if err != nil {
				t.Fatal(err)
			}
			for _, h := range result.Matches {
				for _, key := range judgment.acceptable {
					if h.Entity.Key == key {
						return
					}
				}
			}
			t.Fatalf("no engineering-judged acceptable candidate in first ten: %+v", result.Matches)
		})
	}
	for _, opts := range []LibraryIndexFindOptions{
		{Retrieval: "keyword", Lifecycles: []string{"unknown-lifecycle"}, Shape: LibrarySearchOptions{Query: "cards"}},
		{Retrieval: "keyword", ContentAdapter: "unknown-adapter", Shape: LibrarySearchOptions{Query: "cards"}},
		{Retrieval: "keyword", Kinds: []string{"template') OR 1=1 --"}, Shape: LibrarySearchOptions{Query: "cards"}},
	} {
		r, err := index.Find(opts)
		if err != nil || len(r.Matches) != 0 {
			t.Fatalf("filter not applied: %v %+v", err, r)
		}
	}
}

func BenchmarkKeywordSearchWarm(b *testing.B) {
	path := keywordV11Index(b)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		b.Fatal(err)
	}
	defer index.Close()
	info, _ := os.Stat(path)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = index.Find(LibraryIndexFindOptions{Retrieval: "keyword", Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: "modernization economics", Limit: 10}})
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(info.Size()), "index_bytes")
}

func BenchmarkKeywordVerifiedOpen(b *testing.B) {
	path := keywordV11Index(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
		if err != nil {
			b.Fatal(err)
		}
		if err = index.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
