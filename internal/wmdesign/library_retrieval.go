package wmdesign

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/localembed"
)

// Text preparation is versioned separately from the entity projection. Changing
// this recipe requires rebuilding the FTS index and any future vector snapshot.
const LibraryRetrievalTextVersion = "pptxgengo.discovery-text.v1"
const libraryFTSSchema = "CREATE VIRTUAL TABLE entity_fts USING fts5(id UNINDEXED,name,purpose,body,tokenize='porter unicode61')"

type LibraryRetrievalReport struct {
	Requested       string               `json:"requested"`
	Actual          string               `json:"actual"`
	TextPreparation string               `json:"text_preparation,omitempty"`
	Ranking         string               `json:"ranking"`
	Notice          string               `json:"notice,omitempty"`
	Model           *localembed.Identity `json:"model,omitempty"`
	VectorCoverage  int                  `json:"vector_coverage,omitempty"`
	FusionK         int                  `json:"fusion_k,omitempty"`
}

type LibraryRetrievalHit struct {
	Method      string   `json:"method"`
	Rank        int      `json:"rank"`
	BM25        float64  `json:"bm25,omitempty"`
	Cosine      *float64 `json:"cosine,omitempty"`
	KeywordRank int      `json:"keyword_rank,omitempty"`
	VectorRank  int      `json:"vector_rank,omitempty"`
	RRF         float64  `json:"rrf,omitempty"`
}

func libraryRetrievalBody(entity LibraryEntity) string {
	parts := []string{entity.ID, entity.Key, entity.Family}
	parts = append(parts, entity.Discovery.ContentRoles...)
	parts = append(parts, entity.Discovery.Structures...)
	parts = append(parts, entity.Discovery.VisualForms...)
	parts = append(parts, entity.Discovery.ComponentTypes...)
	for _, group := range entity.Discovery.Groups {
		parts = append(parts, group.Role, group.ComponentType)
	}
	for _, relationship := range entity.Discovery.Relationships {
		parts = append(parts, relationship.Kind)
	}
	if entity.Template != nil {
		parts = append(parts, entity.Template.Uses...)
		if authoring := entity.Template.Authoring; authoring != nil {
			parts = append(parts, authoring.Recipe, authoring.Relationship)
			for _, group := range authoring.Groups {
				parts = append(parts, group.Alias, group.Role)
			}
			for _, slot := range authoring.Slots {
				parts = append(parts, slot.Alias, slot.Role, slot.Description)
			}
		}
	}
	// Sort and deduplicate metadata so repeated slot descriptions do not inflate
	// term frequency. Deliberately omit synthetic example copy and policy boilerplate.
	seen := map[string]bool{}
	clean := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && !seen[part] {
			seen[part] = true
			clean = append(clean, part)
		}
	}
	sort.Strings(clean)
	return strings.Join(clean, "\n")
}

func retrievalProjectionHash(entities []LibraryEntity) string {
	h := sha256.New()
	h.Write([]byte(LibraryRetrievalTextVersion))
	for _, entity := range entities {
		// Length boundaries come from JSON, not ambiguous concatenation.
		h.Write(indexJSON([]string{entity.ID, entity.Name, entity.Purpose, libraryRetrievalBody(entity)}))
		h.Write([]byte{0})
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (index *LibraryIndex) verifyRetrievalProjection(entities []LibraryEntity) error {
	var schema string
	if err := index.db.QueryRow("SELECT sql FROM sqlite_master WHERE name='entity_fts'").Scan(&schema); err != nil {
		return err
	}
	if schema != libraryFTSSchema {
		return fmt.Errorf("index.retrieval_schema_changed: rebuild the library index")
	}
	rows, err := index.db.Query("SELECT id,name,purpose,body FROM entity_fts ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	position := 0
	for rows.Next() {
		var id, name, purpose, body string
		if err := rows.Scan(&id, &name, &purpose, &body); err != nil {
			return err
		}
		if position >= len(entities) {
			return fmt.Errorf("index.retrieval_integrity_failed: unexpected FTS row")
		}
		entity := entities[position]
		if id != entity.ID || name != entity.Name || purpose != entity.Purpose || body != libraryRetrievalBody(entity) {
			return fmt.Errorf("index.retrieval_integrity_failed: %s", id)
		}
		position++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if position != len(entities) {
		return fmt.Errorf("index.retrieval_integrity_failed: missing FTS rows")
	}
	return nil
}

// Treat input as literal natural language, never as the FTS5 query language.
// OR permits partial lexical coverage; punctuation and reserved operators cannot
// change the SQL statement or introduce an FTS expression.
func libraryFTSQuery(query string) (string, error) {
	if len(query) > 4096 {
		return "", fmt.Errorf("library.query_too_long: maximum 4096 bytes")
	}
	seen := map[string]bool{}
	parts := []string{}
	for _, token := range discoveryTokens(query) {
		if seen[token] {
			continue
		}
		seen[token] = true
		if len(parts) == 128 {
			return "", fmt.Errorf("library.query_too_many_terms: maximum 128 distinct terms")
		}
		parts = append(parts, `"`+token+`"`)
	}
	return strings.Join(parts, " OR "), nil
}

func (index *LibraryIndex) keywordRanks(query string, eligible map[string]bool, exactID string) (map[string]LibraryRetrievalHit, error) {
	if index.Report.RetrievalText != LibraryRetrievalTextVersion {
		return nil, fmt.Errorf("index.keyword_rebuild_required: build a new library index with this toolkit; metadata discovery still works")
	}
	expression, err := libraryFTSQuery(query)
	if err != nil {
		return nil, err
	}
	ranks := map[string]LibraryRetrievalHit{}
	if expression == "" {
		return ranks, nil
	}
	// FTS5 BM25 is negative; smaller values rank first. Weights: id unindexed,
	// name 5, purpose 2, structured discovery body 1. The corpus is the pinned
	// complete library; eligibility filters are applied before rank numbering.
	rows, err := index.db.Query("SELECT id,bm25(entity_fts,0.0,5.0,2.0,1.0) AS relevance FROM entity_fts WHERE entity_fts MATCH ? ORDER BY CASE WHEN id=? THEN 0 ELSE 1 END,relevance,id", expression, exactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var score float64
		if err := rows.Scan(&id, &score); err != nil {
			return nil, err
		}
		if eligible[id] {
			method := "sqlite_fts5_bm25"
			if id == exactID {
				method = "exact_entity_identity"
			}
			ranks[id] = LibraryRetrievalHit{Method: method, Rank: len(ranks) + 1, BM25: score}
		}
	}
	return ranks, rows.Err()
}

func libraryStructureStatus(hit LibrarySearchHit, options LibrarySearchOptions) string {
	if len(options.ContentRoles)+len(options.Structures)+len(options.VisualForms) == 0 && options.Items == 0 {
		return "not_requested"
	}
	for _, hint := range hit.UnmatchedHints {
		if !strings.HasPrefix(hint, "text:") {
			return "source_hints_mismatch"
		}
	}
	return "source_hints_match_not_measured_fit"
}
