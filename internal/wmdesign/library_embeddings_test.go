package wmdesign

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/localembed"
)

func embeddingFixture() (LibraryEmbeddingSnapshot, LibraryIndexReport, []LibraryEntity) {
	entities := []LibraryEntity{{ID: "a", Name: "Interview findings", Revision: 2, SourceSHA256: "source-a"}, {ID: "b", Name: "Future operating model", SourceSHA256: "source-b"}}
	report := LibraryIndexReport{ProjectionSHA256: "projection", RetrievalSHA256: "retrieval", SourceRevision: "revision"}
	s := LibraryEmbeddingSnapshot{Schema: LibraryEmbeddingSchema, Model: localembed.PinnedIdentity(), TextPreparation: LibraryEmbeddingTextVersion, IndexSHA256: libraryEmbeddingIndexHash(report)}
	for i, e := range entities {
		v := make([]float32, localembed.Dimensions)
		v[i] = 1
		s.Rows = append(s.Rows, LibraryEmbeddingRow{ID: e.ID, Revision: e.Revision, SourceSHA256: e.SourceSHA256, TextSHA256: indexDigest([]byte(libraryEmbeddingText(e))), Vector: v})
	}
	s.RowsSHA256 = indexDigest(indexJSON(s.Rows))
	return s, report, entities
}

func TestEmbeddingPinsCoverageAndNormalization(t *testing.T) {
	s, r, e := embeddingFixture()
	if err := validateLibraryEmbeddings(s, r, e); err != nil {
		t.Fatal(err)
	}
	mutations := []func(*LibraryEmbeddingSnapshot){func(s *LibraryEmbeddingSnapshot) { s.Model.Revision = "changed" }, func(s *LibraryEmbeddingSnapshot) { s.TextPreparation = "changed" }, func(s *LibraryEmbeddingSnapshot) { s.IndexSHA256 = "changed" }, func(s *LibraryEmbeddingSnapshot) { s.Rows = s.Rows[:1] }, func(s *LibraryEmbeddingSnapshot) { s.Rows[0].ID = "b" }, func(s *LibraryEmbeddingSnapshot) { s.Rows[0].TextSHA256 = "changed" }, func(s *LibraryEmbeddingSnapshot) { s.Rows[0].Revision++ }, func(s *LibraryEmbeddingSnapshot) { s.Rows[0].Vector = s.Rows[0].Vector[:2] }, func(s *LibraryEmbeddingSnapshot) { s.Rows[0].Vector[0] = 0 }, func(s *LibraryEmbeddingSnapshot) { s.Rows[0].Vector[0] = 2 }}
	for i, mutate := range mutations {
		raw := indexJSON(s)
		var changed LibraryEmbeddingSnapshot
		if err := json.Unmarshal(raw, &changed); err != nil {
			t.Fatal(err)
		}
		mutate(&changed)
		changed.RowsSHA256 = indexDigest(indexJSON(changed.Rows))
		if err := validateLibraryEmbeddings(changed, r, e); err == nil {
			t.Fatal("mutation accepted", i)
		}
	}
	s.RowsSHA256 = "changed"
	if err := validateLibraryEmbeddings(s, r, e); err == nil {
		t.Fatal("checksum changed")
	}
}

func TestEmbeddingProjectionRetainsAuthoringWithoutSceneDefinitions(t *testing.T) {
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
	prepared, err := index.embeddingEntities("kind='template'", nil)
	if err != nil || len(full) != len(prepared) {
		t.Fatal(err)
	}
	for i, e := range full {
		p := prepared[i]
		if len(p.Definition) != 0 || len(p.Discovery.Zones) != 0 || string(indexJSON(e.Template.Authoring)) != string(indexJSON(p.Template.Authoring)) || libraryEmbeddingText(e) != libraryEmbeddingText(p) {
			t.Fatal("embedding projection changed discovery text or loaded scene", e.ID)
		}
	}
}

func TestVectorAndHybridRanksSeparateSignals(t *testing.T) {
	s, _, _ := embeddingFixture()
	query := make([]float32, localembed.Dimensions)
	query[1] = 1
	vectors := libraryVectorRanks(&s, query, map[string]bool{"a": true, "b": true})
	if vectors["b"].Rank != 1 || *vectors["b"].Cosine != 1 {
		t.Fatal(vectors)
	}
	lex := map[string]LibraryRetrievalHit{"a": {Rank: 1, BM25: -1000000}}
	fused := libraryHybridRanks(lex, vectors, "")
	if fused["a"].Rank != 1 || math.Abs(fused["a"].RRF-(1.0/61+1.0/62)) > 1e-15 || fused["a"].BM25 != -1000000 || fused["a"].KeywordRank != 1 || fused["a"].VectorRank != 2 {
		t.Fatal(fused)
	}
	if got := libraryHybridRanks(lex, vectors, "b"); got["b"].Rank != 1 || got["b"].Method != "exact_entity_identity" {
		t.Fatal(got)
	}
	filtered := libraryVectorRanks(&s, query, map[string]bool{"a": true})
	if len(filtered) != 1 || filtered["a"].Rank != 1 {
		t.Fatal(filtered)
	}
	query[0] = 1
	vectors = libraryVectorRanks(&s, query, map[string]bool{"a": true, "b": true})
	if vectors["a"].Rank != 1 {
		t.Fatal("canonical tie changed")
	}
}

func TestHybridMissingResourcesFallbackAndCorruptionReject(t *testing.T) {
	path, _ := indexFixture(t)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	opts := LibraryIndexFindOptions{Retrieval: "hybrid", Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: "cards/3", Items: 4, ItemRole: "point"}}
	r, err := index.Find(opts)
	if err != nil || r.Retrieval.Requested != "hybrid" || r.Retrieval.Actual != "keyword" || r.Retrieval.Notice == "" || len(r.Matches) == 0 {
		t.Fatal(r, err)
	}
	if r.Matches[0].StructuralStatus != "source_hints_mismatch" {
		t.Fatal("fallback hid source mismatch")
	}
	opts.Retrieval = "semantic"
	if _, err = index.Find(opts); err == nil {
		t.Fatal("semantic missing resources accepted")
	}
	opts.Shape.Query = ""
	r, err = index.Find(opts)
	if err != nil || r.Retrieval.Actual != "metadata" {
		t.Fatal(r, err)
	}
	opts.Retrieval = "hybrid"
	opts.Shape.Query = "cards"
	opts.Embeddings = filepath.Join(t.TempDir(), "corrupt.json")
	opts.ModelDir = "missing"
	if err = os.WriteFile(opts.Embeddings, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = index.Find(opts); err == nil || !strings.Contains(err.Error(), "incompatible") {
		t.Fatal("corruption silently fell back", err)
	}
}

func TestPinnedLibraryEmbeddingsEndToEnd(t *testing.T) {
	modelDir := os.Getenv("PPTXGENGO_EMBED_MODEL_DIR")
	if modelDir == "" {
		t.Skip("optional pinned weights")
	}
	path, _ := indexFixture(t)
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	out := filepath.Join(t.TempDir(), "embeddings.json")
	snapshot, err := index.BuildEmbeddings(context.Background(), modelDir, out)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, n := range index.Report.Counts {
		count += n
	}
	if len(snapshot.Rows) != count {
		t.Fatal("partial embedding coverage", len(snapshot.Rows), index.Report.Counts)
	}
	for _, mode := range []string{"semantic", "hybrid"} {
		opts := LibraryIndexFindOptions{Retrieval: mode, Embeddings: out, ModelDir: modelDir, Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: "three parallel findings", Limit: 10}}
		r, err := index.Find(opts)
		if err != nil || r.Retrieval.Actual != mode || r.Retrieval.VectorCoverage != len(snapshot.Rows) || len(r.Matches) == 0 {
			t.Fatal(r, err)
		}
		for _, h := range r.Matches {
			if h.FitStatus != "not_measured_for_query" || h.Retrieval.Cosine == nil {
				t.Fatal("retrieval misreported", h)
			}
		}
		again, err := index.Find(opts)
		if err != nil || string(indexJSON(r)) != string(indexJSON(again)) {
			t.Fatal("nondeterministic", err)
		}
	}
}
