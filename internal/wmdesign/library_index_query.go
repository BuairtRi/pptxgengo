package wmdesign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LibraryIndexFindOptions struct {
	Shape     LibrarySearchOptions `json:"shape"`
	Kinds     []string             `json:"kinds,omitempty"`
	Namespace string               `json:"namespace,omitempty"`
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
	Entity         LibraryEntitySummary `json:"entity"`
	Score          int                  `json:"score"`
	ScenarioScore  int                  `json:"scenario_score"`
	Reasons        []string             `json:"reasons"`
	UnmatchedHints []string             `json:"unmatched_hints,omitempty"`
	CountMatch     *LibraryCountMatch   `json:"count_match,omitempty"`
	FitStatus      string               `json:"fit_status"`
}

type LibraryIndexFindResult struct {
	Engine     LibrarySearchEngine        `json:"engine"`
	Schema     string                     `json:"schema"`
	Query      LibraryIndexFindOptions    `json:"query"`
	Vocabulary LibraryDiscoveryVocabulary `json:"vocabulary"`
	Matches    []LibraryIndexHit          `json:"matches"`
	Policy     []string                   `json:"policy"`
}

func (index *LibraryIndex) Find(options LibraryIndexFindOptions) (LibraryIndexFindResult, error) {
	result := LibraryIndexFindResult{Engine: LibrarySearchEngine{Requested: options.Shape.EngineHint, Compatibility: "not_evaluated_search_only"}, Schema: "pptxgengo.unified-library-search.v1", Query: options, Vocabulary: DiscoveryVocabulary(), Matches: []LibraryIndexHit{}, Policy: []string{"Kinds and namespace are explicit filters; scenario and content-shape fields are ranking hints.", "Deprecated entries are hidden unless requested; discovery does not impose a qualification gate.", "Primary/source example counts and source-advisory budgets are not measured fit for supplied content.", "Modern entities precede legacy entities only on equal score; canonical identity breaks remaining ties."}}
	if _, e := SearchLibrary(nil, options.Shape); e != nil {
		return result, e
	}
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
	entities, e := index.entities(strings.Join(filters, " AND "), args)
	if e != nil {
		return result, e
	}
	for _, entity := range entities {
		def := LibraryTemplate{TemplateDefinition: TemplateDefinition{Key: entity.Key}, Name: entity.Name, Purpose: entity.Purpose, Family: entity.Family, Status: entity.Lifecycle, Discovery: entity.Discovery}
		if entity.Template != nil {
			def = *entity.Template
			def.Discovery = entity.Discovery
		}
		part, e := SearchLibrary([]LibraryTemplate{def}, options.Shape)
		if e != nil {
			return result, e
		}
		if len(part.Matches) == 0 {
			continue
		}
		hit := part.Matches[0]
		discovery := entity.Discovery
		discovery.Zones = nil
		summary := LibraryEntitySummary{ID: entity.ID, Namespace: entity.Namespace, Kind: entity.Kind, Key: entity.Key, Name: entity.Name, Purpose: entity.Purpose, Lifecycle: entity.Lifecycle, Revision: entity.Revision, SourceRevision: entity.SourceRevision, SourceSHA256: entity.SourceSHA256, Discovery: discovery, Capacity: entity.Capacity, SupportedAdaptations: entity.SupportedAdaptations, PreviewCount: len(entity.Artifacts)}
		result.Matches = append(result.Matches, LibraryIndexHit{Entity: summary, Score: hit.Score, ScenarioScore: hit.ScenarioScore, Reasons: hit.Reasons, UnmatchedHints: hit.UnmatchedHints, CountMatch: hit.CountMatch, FitStatus: hit.FitStatus})
	}
	sort.Slice(result.Matches, func(i, j int) bool {
		a, b := result.Matches[i], result.Matches[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.ScenarioScore != b.ScenarioScore {
			return a.ScenarioScore > b.ScenarioScore
		}
		if a.Entity.Namespace != b.Entity.Namespace {
			return a.Entity.Namespace == "wmds"
		}
		return a.Entity.ID < b.Entity.ID
	})
	limit := options.Shape.Limit
	if limit == 0 {
		limit = 10
	}
	if len(result.Matches) > limit {
		result.Matches = result.Matches[:limit]
	}
	return result, nil
}

type LibraryPreviewResult struct {
	Schema      string                `json:"schema"`
	EntityID    string                `json:"entity_id"`
	Artifacts   []LibraryArtifactLink `json:"artifacts"`
	Paths       []string              `json:"paths"`
	Preparation string                `json:"preparation,omitempty"`
	Definition  json.RawMessage       `json:"definition,omitempty"`
	Capability  LibraryCapability     `json:"capability"`
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
		root := os.Getenv("WMDS_BRANDING_ROOT")
		if root == "" && os.Getenv("PPTXGENGO_RELEASE_ROOT") != "" {
			root = filepath.Join(os.Getenv("PPTXGENGO_RELEASE_ROOT"), "branding")
		}
		if root == "" {
			home, e := os.UserHomeDir()
			if e != nil {
				return result, e
			}
			root = filepath.Join(home, "Documents", "branding")
		}
		// Paths come from the pinned native registry, including approved sibling career assets.
		path := filepath.Join(root, filepath.Clean(asset.Path))
		h, e := indexFileDigest(path)
		if e != nil {
			return result, e
		}
		if h != asset.SHA256 {
			return result, fmt.Errorf("index.asset_drift: %s", asset.Key)
		}
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
