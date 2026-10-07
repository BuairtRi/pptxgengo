package wmdesign

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/localembed"
)

const LibraryEmbeddingTextVersion = "pptxgengo.semantic-discovery-text.v1"
const LibraryEmbeddingSchema = "pptxgengo.library-embeddings.v1"
const LibraryRRFK = 60

type LibraryEmbeddingRow struct {
	ID             string    `json:"id"`
	Revision       int       `json:"revision,omitempty"`
	SourceRevision string    `json:"source_revision,omitempty"`
	SourceSHA256   string    `json:"source_sha256,omitempty"`
	TextSHA256     string    `json:"text_sha256"`
	Vector         []float32 `json:"vector"`
}
type LibraryEmbeddingSnapshot struct {
	Schema          string                `json:"schema"`
	Model           localembed.Identity   `json:"model"`
	TextPreparation string                `json:"text_preparation"`
	IndexSHA256     string                `json:"index_sha256"`
	Rows            []LibraryEmbeddingRow `json:"rows"`
	RowsSHA256      string                `json:"rows_sha256"`
}

func libraryEmbeddingIndexHash(r LibraryIndexReport) string {
	// Bind the full entity/source/asset projection without installation paths.
	return indexDigest(indexJSON(struct {
		Projection, Retrieval, Assets, Source string
		Pins                                  []LibraryIndexPin
	}{r.ProjectionSHA256, r.RetrievalSHA256, r.AssetRegistrySHA256, r.SourceRevision, r.Pins}))
}

func libraryEmbeddingText(e LibraryEntity) string {
	// Lead with human meaning before bounded token truncation. Discovery metadata
	// is deterministic; synthetic examples, source definitions and policies stay out.
	humanize := strings.NewReplacer("_", " ", "-", " ", "/", " ", "\n", ". ")
	parts := []string{e.Name, e.Purpose}
	if e.Template != nil {
		parts = append(parts, e.Template.Uses...)
	}
	parts = append(parts, humanize.Replace(libraryRetrievalBody(e)))
	return strings.Join(parts, ". ")
}

func (index *LibraryIndex) BuildEmbeddings(ctx context.Context, modelDir, out string) (LibraryEmbeddingSnapshot, error) {
	snapshot := LibraryEmbeddingSnapshot{Schema: LibraryEmbeddingSchema, Model: localembed.PinnedIdentity(), TextPreparation: LibraryEmbeddingTextVersion, IndexSHA256: libraryEmbeddingIndexHash(index.Report), Rows: []LibraryEmbeddingRow{}}
	if out == "" {
		return snapshot, fmt.Errorf("embedding.output_required")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		return snapshot, fmt.Errorf("embedding.output_exists: %s", out)
	}
	entities, err := index.discoveryEntities("1=1", nil)
	if err != nil {
		return snapshot, err
	}
	m, err := localembed.Load(modelDir)
	if err != nil {
		return snapshot, err
	}
	defer m.Close()
	for _, e := range entities {
		text := libraryEmbeddingText(e)
		v, err := m.Embed(ctx, text)
		if err != nil {
			return snapshot, fmt.Errorf("embedding.entity %s: %w", e.ID, err)
		}
		snapshot.Rows = append(snapshot.Rows, LibraryEmbeddingRow{e.ID, e.Revision, e.SourceRevision, e.SourceSHA256, indexDigest([]byte(text)), v})
	}
	snapshot.RowsSHA256 = indexDigest(indexJSON(snapshot.Rows))
	data, err := json.Marshal(snapshot)
	if err != nil {
		return snapshot, err
	}
	if err = localembed.WriteSnapshot(out, append(data, '\n')); err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func (index *LibraryIndex) readEmbeddings(path string) (*LibraryEmbeddingSnapshot, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 128<<20 {
		return nil, fmt.Errorf("embedding.snapshot_invalid_file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 128<<20+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 128<<20 {
		return nil, fmt.Errorf("embedding.snapshot_too_large")
	}
	var snapshot LibraryEmbeddingSnapshot
	if err = json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("embedding.snapshot_invalid: %w", err)
	}
	entities, err := index.discoveryEntities("1=1", nil)
	if err != nil {
		return nil, err
	}
	if err = validateLibraryEmbeddings(snapshot, index.Report, entities); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func validateLibraryEmbeddings(s LibraryEmbeddingSnapshot, r LibraryIndexReport, entities []LibraryEntity) error {
	if s.Schema != LibraryEmbeddingSchema || s.TextPreparation != LibraryEmbeddingTextVersion || string(indexJSON(s.Model)) != string(indexJSON(localembed.PinnedIdentity())) {
		return fmt.Errorf("embedding.snapshot_incompatible: rebuild with the pinned model/runtime/text recipe")
	}
	if s.IndexSHA256 != libraryEmbeddingIndexHash(r) || len(s.Rows) != len(entities) {
		return fmt.Errorf("embedding.snapshot_stale: rebuild for this complete source/index projection")
	}
	if s.RowsSHA256 != indexDigest(indexJSON(s.Rows)) {
		return fmt.Errorf("embedding.snapshot_integrity_failed")
	}
	for i, row := range s.Rows {
		e := entities[i]
		if row.ID != e.ID || row.Revision != e.Revision || row.SourceRevision != e.SourceRevision || row.SourceSHA256 != e.SourceSHA256 || row.TextSHA256 != indexDigest([]byte(libraryEmbeddingText(e))) {
			return fmt.Errorf("embedding.snapshot_stale: source/text/identity %s", row.ID)
		}
		if len(row.Vector) != localembed.Dimensions {
			return fmt.Errorf("embedding.vector_dimensions: %s", row.ID)
		}
		norm := 0.0
		for _, x := range row.Vector {
			v := float64(x)
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("embedding.vector_nonfinite: %s", row.ID)
			}
			norm += v * v
		}
		if math.Abs(norm-1) > 1e-4 {
			return fmt.Errorf("embedding.vector_not_normalized: %s", row.ID)
		}
	}
	return nil
}

func libraryVectorRanks(s *LibraryEmbeddingSnapshot, query []float32, eligible map[string]bool) map[string]LibraryRetrievalHit {
	type scored struct {
		id  string
		cos float64
	}
	ordered := []scored{}
	for _, row := range s.Rows {
		if !eligible[row.ID] {
			continue
		}
		cos := 0.0
		for j, x := range row.Vector {
			cos += float64(x) * float64(query[j])
		}
		ordered = append(ordered, scored{row.ID, cos})
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].cos != ordered[j].cos {
			return ordered[i].cos > ordered[j].cos
		}
		return ordered[i].id < ordered[j].id
	})
	ranks := map[string]LibraryRetrievalHit{}
	for i, e := range ordered {
		cos := e.cos
		ranks[e.id] = LibraryRetrievalHit{Method: "local_minilm_cosine", Rank: i + 1, VectorRank: i + 1, Cosine: &cos}
	}
	return ranks
}

func libraryHybridRanks(lexical, vectors map[string]LibraryRetrievalHit, exact string) map[string]LibraryRetrievalHit {
	ranks := map[string]LibraryRetrievalHit{}
	for id, vector := range vectors {
		h := vector
		h.Method = "reciprocal_rank_fusion"
		h.RRF = 1 / float64(LibraryRRFK+vector.Rank)
		if lex, ok := lexical[id]; ok {
			h.KeywordRank = lex.Rank
			h.BM25 = lex.BM25
			h.RRF += 1 / float64(LibraryRRFK+lex.Rank)
		}
		ranks[id] = h
	}
	for id, lex := range lexical {
		if _, ok := ranks[id]; !ok {
			lex.Method = "reciprocal_rank_fusion"
			lex.KeywordRank = lex.Rank
			lex.RRF = 1 / float64(LibraryRRFK+lex.Rank)
			ranks[id] = lex
		}
	}
	ids := make([]string, 0, len(ranks))
	for id := range ranks {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		if (ids[i] == exact) != (ids[j] == exact) {
			return ids[i] == exact
		}
		if ranks[ids[i]].RRF != ranks[ids[j]].RRF {
			return ranks[ids[i]].RRF > ranks[ids[j]].RRF
		}
		return ids[i] < ids[j]
	})
	for i, id := range ids {
		h := ranks[id]
		h.Rank = i + 1
		if id == exact {
			h.Method = "exact_entity_identity"
		}
		ranks[id] = h
	}
	return ranks
}
