package wmdesign

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/localembed"
)

type LibraryIndexFindOptions struct {
	Shape          LibrarySearchOptions `json:"shape"`
	Kinds          []string             `json:"kinds,omitempty"`
	Namespace      string               `json:"namespace,omitempty"`
	IncludeWeak    bool                 `json:"include_weak,omitempty"`
	Retrieval      string               `json:"retrieval,omitempty"`
	Embeddings     string               `json:"embeddings,omitempty"`
	ModelDir       string               `json:"model_dir,omitempty"`
	RequireShape   bool                 `json:"require_shape,omitempty"`
	Lifecycles     []string             `json:"lifecycles,omitempty"`
	ContentAdapter string               `json:"content_adapter,omitempty"`
}

type LibraryEntitySummary struct {
	ID                   string           `json:"id"`
	Namespace            string           `json:"namespace"`
	Kind                 string           `json:"kind"`
	Key                  string           `json:"key"`
	Name                 string           `json:"name"`
	Purpose              string           `json:"purpose,omitempty"`
	Lifecycle            string           `json:"lifecycle"`
	Revision             int              `json:"revision,omitempty"`
	SourceRevision       string           `json:"source_revision,omitempty"`
	SourceSHA256         string           `json:"source_sha256,omitempty"`
	Discovery            LibraryDiscovery `json:"discovery"`
	Capacity             json.RawMessage  `json:"capacity,omitempty"`
	SupportedAdaptations []string         `json:"supported_adaptations"`
	PreviewCount         int              `json:"preview_count"`
}

type LibraryIndexHit struct {
	Entity           LibraryEntitySummary `json:"entity"`
	GroupID          string               `json:"group_id,omitempty"`
	VariantIDs       []string             `json:"variant_ids,omitempty"`
	Score            int                  `json:"score"`
	ScenarioScore    int                  `json:"scenario_score"`
	Reasons          []string             `json:"reasons"`
	UnmatchedHints   []string             `json:"unmatched_hints,omitempty"`
	CountMatch       *LibraryCountMatch   `json:"count_match,omitempty"`
	FitStatus        string               `json:"fit_status"`
	StructuralStatus string               `json:"structural_status"`
	Retrieval        *LibraryRetrievalHit `json:"retrieval,omitempty"`
}

type LibraryIndexFindResult struct {
	Engine     LibrarySearchEngine        `json:"engine"`
	Schema     string                     `json:"schema"`
	Query      LibraryIndexFindOptions    `json:"query"`
	Vocabulary LibraryDiscoveryVocabulary `json:"vocabulary"`
	Matches    []LibraryIndexHit          `json:"matches"`
	Policy     []string                   `json:"policy"`
	Retrieval  LibraryRetrievalReport     `json:"retrieval"`
}

func (index *LibraryIndex) Find(options LibraryIndexFindOptions) (LibraryIndexFindResult, error) {
	result := LibraryIndexFindResult{Engine: LibrarySearchEngine{Requested: options.Shape.EngineHint, Compatibility: "not_evaluated_search_only"}, Schema: "pptxgengo.unified-library-search.v1", Query: options, Vocabulary: DiscoveryVocabulary(), Matches: []LibraryIndexHit{}, Policy: []string{"Kinds and namespace are explicit filters; scenario and content-shape fields are ranking hints.", "Deprecated entries are hidden unless requested; discovery does not impose a qualification gate.", "Primary/source example counts and source-advisory budgets are not measured fit for supplied content.", "Metadata ties prefer modern entities, then canonical identity. Text rankings use their declared retrieval ranks and canonical identity ties."}}
	if _, e := SearchLibrary(nil, options.Shape); e != nil {
		return result, e
	}
	mode := options.Retrieval
	if mode == "" {
		mode = "metadata"
	}
	if mode != "metadata" && mode != "keyword" && mode != "semantic" && mode != "hybrid" {
		return result, fmt.Errorf("library.retrieval_invalid: choose metadata, keyword, semantic or hybrid")
	}
	requested := mode
	notice := ""
	var snapshot *LibraryEmbeddingSnapshot
	var queryVector []float32
	textQuery := strings.TrimSpace(options.Shape.Query) != ""
	if (mode == "metadata" || mode == "keyword") && (options.Embeddings != "" || options.ModelDir != "") {
		return result, fmt.Errorf("library.model_options: --embeddings and --model-dir require semantic or hybrid retrieval")
	}
	if (mode == "semantic" || mode == "hybrid") && textQuery {
		missing := options.Embeddings == "" || options.ModelDir == ""
		if !missing {
			var err error
			snapshot, err = index.readEmbeddings(options.Embeddings)
			if err != nil {
				if os.IsNotExist(err) && mode == "hybrid" {
					missing = true
				} else {
					return result, err
				}
			}
			if !missing {
				m, err := localembed.Load(options.ModelDir)
				if err != nil {
					if os.IsNotExist(err) && mode == "hybrid" {
						missing = true
					} else {
						return result, err
					}
				} else {
					queryVector, err = m.Embed(context.Background(), options.Shape.Query)
					m.Close()
					if err != nil {
						return result, err
					}
				}
			}
		}
		if missing {
			if mode == "semantic" {
				return result, fmt.Errorf("embedding.resources_missing: semantic retrieval requires compatible --embeddings FILE and offline --model-dir DIR")
			}
			mode = "keyword"
			snapshot = nil
			notice = "Offline model or embedding snapshot unavailable; hybrid retrieval explicitly fell back to keyword. No network request was made."
		}
	}
	ranked := mode != "metadata" && textQuery
	result.Retrieval = LibraryRetrievalReport{Requested: requested, Actual: mode, Ranking: "metadata_and_structural_hints", Notice: notice}
	if mode != "metadata" {
		result.Retrieval.TextPreparation = index.Report.RetrievalText
		result.Retrieval.Ranking = "exact_identity_then_sqlite_fts5_bm25_ascending_then_canonical_id"
		if snapshot != nil {
			result.Retrieval.Model = &snapshot.Model
			result.Retrieval.VectorCoverage = len(snapshot.Rows)
			result.Retrieval.TextPreparation = snapshot.TextPreparation
			result.Retrieval.Ranking = "normalized_cosine_descending_then_canonical_id"
			if mode == "hybrid" {
				result.Retrieval.Ranking = "exact_identity_then_reciprocal_rank_fusion_descending_then_canonical_id"
				result.Retrieval.FusionK = LibraryRRFK
			}
		}
		result.Policy = append(result.Policy, "Lexical relevance, normalized cosine and reciprocal rank fusion are separate from structural scores and measured content fit. Hybrid uses the sum of 1/(60+eligible rank) for each available retriever, never an addition of raw BM25 and cosine values.")
		if !ranked {
			result.Retrieval.Actual = "metadata"
			result.Retrieval.Ranking = "metadata_and_structural_hints"
			result.Retrieval.Notice = "An empty text query browses structural metadata; no text retrieval was performed."
		}
	}
	result.Policy = append(result.Policy, "--require-shape promotes every supplied role, structure, visual-form and exact source item-count hint to an eligibility constraint. Source agreement is not measured fit for new content.")
	filters, args := []string{"1=1"}, []any{}
	if options.Namespace != "" {
		filters = append(filters, "namespace=?")
		args = append(args, options.Namespace)
	}
	if len(options.Kinds) > 0 {
		placeholders := make([]string, len(options.Kinds))
		for i, kind := range options.Kinds {
			placeholders[i] = "?"
			args = append(args, kind)
		}
		filters = append(filters, "kind IN ("+strings.Join(placeholders, ",")+")")
	}
	if !options.Shape.IncludeDeprecated {
		filters = append(filters, "lifecycle!='deprecated'")
	}
	if len(options.Lifecycles) > 0 {
		placeholders := make([]string, len(options.Lifecycles))
		for i, lifecycle := range options.Lifecycles {
			placeholders[i] = "?"
			args = append(args, lifecycle)
		}
		filters = append(filters, "lifecycle IN ("+strings.Join(placeholders, ",")+")")
	}
	if options.ContentAdapter != "" {
		filters = append(filters, "json_extract(json,'$.discovery.capability.content_adapter')=?")
		args = append(args, options.ContentAdapter)
	}
	entities, e := index.discoveryEntities(strings.Join(filters, " AND "), args)
	if e != nil {
		return result, e
	}
	queryCoverage := map[string]int{}
	eligible := map[string]bool{}
	for _, entity := range entities {
		def := LibraryTemplate{TemplateDefinition: TemplateDefinition{Key: entity.Key}, Name: entity.Name, Purpose: entity.Purpose, Family: entity.Family, Status: entity.Lifecycle, Discovery: entity.Discovery}
		if entity.Template != nil {
			def.Uses = entity.Template.Uses
		}
		shape := options.Shape
		if ranked {
			shape.Query = ""
		}
		part, e := SearchLibrary([]LibraryTemplate{def}, shape)
		if e != nil {
			return result, e
		}
		if len(part.Matches) == 0 {
			continue
		}
		hit := part.Matches[0]
		structuralStatus := libraryStructureStatus(hit, options.Shape)
		if options.RequireShape && structuralStatus == "source_hints_mismatch" {
			continue
		}
		if !ranked && strings.TrimSpace(options.Shape.Query) != "" && hit.Score == 0 && !options.IncludeWeak {
			continue
		}
		eligible[entity.ID] = true
		queryCoverage[entity.ID] = libraryScenarioQueryCoverage(def, options.Shape.Query)
		discovery := entity.Discovery
		discovery.Zones = nil
		summary := LibraryEntitySummary{ID: entity.ID, Namespace: entity.Namespace, Kind: entity.Kind, Key: entity.Key, Name: entity.Name, Purpose: entity.Purpose, Lifecycle: entity.Lifecycle, Revision: entity.Revision, SourceRevision: entity.SourceRevision, SourceSHA256: entity.SourceSHA256, Discovery: discovery, Capacity: entity.Capacity, SupportedAdaptations: entity.SupportedAdaptations, PreviewCount: len(entity.Artifacts)}
		result.Matches = append(result.Matches, LibraryIndexHit{Entity: summary, Score: hit.Score, ScenarioScore: hit.ScenarioScore, Reasons: hit.Reasons, UnmatchedHints: hit.UnmatchedHints, CountMatch: hit.CountMatch, FitStatus: hit.FitStatus, StructuralStatus: structuralStatus})
	}
	if ranked {
		exactID, uniqueKeyID := "", ""
		keyMatches := 0
		for _, entity := range entities {
			if !eligible[entity.ID] {
				continue
			}
			if entity.ID == strings.TrimSpace(options.Shape.Query) {
				exactID = entity.ID
			}
			if entity.Key == strings.TrimSpace(options.Shape.Query) {
				uniqueKeyID = entity.ID
				keyMatches++
			}
		}
		if exactID == "" && keyMatches == 1 {
			exactID = uniqueKeyID
		}
		var ranks map[string]LibraryRetrievalHit
		if mode == "keyword" || mode == "hybrid" {
			var err error
			ranks, err = index.keywordRanks(options.Shape.Query, eligible, exactID)
			if err != nil {
				return result, err
			}
		}
		if mode == "semantic" {
			ranks = libraryVectorRanks(snapshot, queryVector, eligible)
		}
		if mode == "hybrid" {
			ranks = libraryHybridRanks(ranks, libraryVectorRanks(snapshot, queryVector, eligible), exactID)
		}
		matches := []LibraryIndexHit{}
		for _, hit := range result.Matches {
			if rank, ok := ranks[hit.Entity.ID]; ok {
				hit.Retrieval = &rank
				if rank.Method == "exact_entity_identity" {
					hit.Reasons = append(hit.Reasons, "Exact eligible entity ID or unique canonical key matches the query.")
				} else if mode == "keyword" {
					hit.Reasons = append(hit.Reasons, fmt.Sprintf("FTS5 BM25 lexical rank %d; smaller raw BM25 values rank first", rank.Rank))
				} else {
					hit.Reasons = append(hit.Reasons, fmt.Sprintf("%s retrieval rank %d; source constraints remain separate from relevance", mode, rank.Rank))
				}
			} else if !options.IncludeWeak {
				continue
			} else {
				hit.Reasons = append(hit.Reasons, "No literal keyword match; included by --include-weak.")
			}
			matches = append(matches, hit)
		}
		result.Matches = matches
	}
	sort.Slice(result.Matches, func(i, j int) bool {
		a, b := result.Matches[i], result.Matches[j]
		if ranked {
			if (a.Retrieval == nil) != (b.Retrieval == nil) {
				return a.Retrieval != nil
			}
			if a.Retrieval != nil && b.Retrieval != nil && a.Retrieval.Rank != b.Retrieval.Rank {
				return a.Retrieval.Rank < b.Retrieval.Rank
			}
		}
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if queryCoverage[a.Entity.ID] != queryCoverage[b.Entity.ID] {
			return queryCoverage[a.Entity.ID] > queryCoverage[b.Entity.ID]
		}
		if a.ScenarioScore != b.ScenarioScore {
			return a.ScenarioScore > b.ScenarioScore
		}
		if a.Entity.Namespace != b.Entity.Namespace {
			return a.Entity.Namespace == "wmds"
		}
		return a.Entity.ID < b.Entity.ID
	})
	result.Matches = groupIndexedIconVariants(result.Matches)
	limit := options.Shape.Limit
	if limit == 0 {
		limit = 10
	}
	if len(result.Matches) > limit {
		result.Matches = result.Matches[:limit]
	}
	return result, nil
}

// groupIndexedIconVariants collapses indexed color instances into one ranked
// concept row while keeping each registered instance ID explicit.
func groupIndexedIconVariants(matches []LibraryIndexHit) []LibraryIndexHit {
	grouped := make([]LibraryIndexHit, 0, len(matches))
	positions := map[string]int{}
	for _, hit := range matches {
		parts := strings.Split(hit.Entity.Key, "/")
		if hit.Entity.Kind != "asset" || len(parts) != 3 || parts[0] != "icon" {
			grouped = append(grouped, hit)
			continue
		}
		groupID := "wmds/asset/icon/" + parts[1]
		position, ok := positions[groupID]
		if !ok {
			hit.VariantIDs = []string{hit.Entity.ID}
			hit.GroupID = groupID
			hit.Entity.Key = "icon/" + parts[1]
			positions[groupID] = len(grouped)
			grouped = append(grouped, hit)
			continue
		}
		grouped[position].VariantIDs = append(grouped[position].VariantIDs, hit.Entity.ID)
	}
	for i := range grouped {
		sort.Strings(grouped[i].VariantIDs)
	}
	return grouped
}

type LibraryPreviewResult struct {
	Schema             string                `json:"schema"`
	EntityID           string                `json:"entity_id"`
	Artifacts          []LibraryArtifactLink `json:"artifacts"`
	Paths              []string              `json:"paths"`
	Preparation        string                `json:"preparation,omitempty"`
	Definition         json.RawMessage       `json:"definition,omitempty"`
	Capability         LibraryCapability     `json:"capability"`
	OriginalValidation string                `json:"original_validation,omitempty"`
}

func (index *LibraryIndex) Preview(id string) (LibraryPreviewResult, error) {
	entity, e := index.Inspect(id)
	if e != nil {
		return LibraryPreviewResult{}, e
	}
	result := LibraryPreviewResult{Schema: "pptxgengo.unified-library-preview.v1", EntityID: entity.ID, Artifacts: []LibraryArtifactLink{}, Paths: []string{}, Capability: entity.Discovery.Capability}
	for _, artifact := range entity.Artifacts {
		path, e := indexRelative(index.Options.Gallery, artifact.Path)
		if e != nil {
			return result, e
		}
		h, e := indexFileDigest(path)
		if e != nil {
			return result, e
		}
		if h != artifact.SHA256 {
			return result, fmt.Errorf("index.preview_drift: %s", artifact.Path)
		}
		result.Artifacts = append(result.Artifacts, artifact)
		result.Paths = append(result.Paths, path)
	}
	if entity.Namespace == "wmds" && entity.Kind == "asset" {
		var asset PrimitiveAssetReference
		if e = json.Unmarshal(entity.Definition, &asset); e != nil {
			return result, e
		}
		verified, path, e := PrimitiveAssetOriginal(asset.Key)
		if e != nil {
			return result, e
		}
		if verified.Path != asset.Path || verified.SHA256 != asset.SHA256 {
			return result, fmt.Errorf("index.asset_drift: %s", asset.Key)
		}
		result.OriginalValidation = "sha256_verified_selected_original"
		result.Paths = append(result.Paths, path)
		result.Artifacts = append(result.Artifacts, LibraryArtifactLink{Role: "registered_original", Path: asset.Path, SHA256: asset.SHA256})
		result.Definition = entity.Definition
	}
	if len(result.Paths) == 0 {
		result.Preparation = "No pinned raster preview was imported. Inspect the definition/source examples or build a supplied-content candidate; do not infer native review from catalog availability."
		result.Definition = entity.Definition
	}
	return result, nil
}
