package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

// Selection cards expose the visual and semantic contract without source
// specimen copy. Bindings identify the editable values belonging to each zone.
type LibrarySelectionSlot struct {
	Name           string              `json:"name"`
	Kind           string              `json:"kind"`
	SourcePointer  string              `json:"source_pointer"`
	AllowEmpty     bool                `json:"allow_empty"`
	Alias          string              `json:"alias,omitempty"`
	Role           string              `json:"role,omitempty"`
	Description    string              `json:"description,omitempty"`
	Classification string              `json:"classification,omitempty"`
	Capacity       LibrarySlotCapacity `json:"capacity"`
}

type LibrarySelectionZone struct {
	LibraryContentZone
	Bindings []string `json:"bindings"`
}

type LibrarySelectionCard struct {
	Kind            string                 `json:"kind"`
	Namespace       string                 `json:"namespace"`
	Lifecycle       string                 `json:"lifecycle"`
	Revision        int                    `json:"revision,omitempty"`
	Capacity        json.RawMessage        `json:"capacity,omitempty"`
	ReuseStatus     string                 `json:"reuse_status,omitempty"`
	Schema          string                 `json:"schema"`
	ID              string                 `json:"id"`
	Key             string                 `json:"key"`
	Name            string                 `json:"name"`
	Purpose         string                 `json:"purpose"`
	SourceRevision  string                 `json:"source_revision"`
	SourceSHA256    string                 `json:"source_sha256"`
	Discovery       LibraryDiscovery       `json:"discovery"`
	ScreenshotPaths []string               `json:"screenshot_paths"`
	Zones           []LibrarySelectionZone `json:"zones"`
	Slots           []LibrarySelectionSlot `json:"slots"`
	Arrays          []LibraryArray         `json:"arrays,omitempty"`
	ValueSchema     *LibraryValueSchema    `json:"typed_values,omitempty"`
	Preparation     string                 `json:"preparation,omitempty"`
	Policy          []string               `json:"policy"`
	Authoring       *LibraryAuthoring      `json:"authoring,omitempty"`
}

func (index *LibraryIndex) SelectionCard(id string) (LibrarySelectionCard, error) {
	entity, err := index.Inspect(id)
	if err != nil {
		return LibrarySelectionCard{}, err
	}
	preview, err := index.Preview(id)
	if err != nil {
		return LibrarySelectionCard{}, err
	}
	card := LibrarySelectionCard{Kind: entity.Kind, Namespace: entity.Namespace, Lifecycle: entity.Lifecycle, Revision: entity.Revision, Capacity: entity.Capacity, ReuseStatus: finishedReuseStatus(entity), Schema: "pptxgengo.library-selection.v1", ID: entity.ID, Key: entity.Key, Name: entity.Name, Purpose: entity.Purpose, SourceRevision: entity.SourceRevision, SourceSHA256: entity.SourceSHA256, Discovery: entity.Discovery, ScreenshotPaths: []string{}, Zones: []LibrarySelectionZone{}, Slots: []LibrarySelectionSlot{}, Preparation: preview.Preparation, Policy: []string{"Screenshot paths are verified against their pinned hashes.", "Zone bounds are source scene coordinates in points; they are not measured fit for supplied content.", "Preserve the slide argument, inspect the screenshot and map content to named bindings before building."}}
	card.Discovery.Zones = nil
	for i, artifact := range preview.Artifacts {
		if artifact.Role == "source_preview" || artifact.Role == "alternate_preview" {
			card.ScreenshotPaths = append(card.ScreenshotPaths, preview.Paths[i])
		}
	}
	if entity.Template != nil {
		metadata, err := index.entityAuthoringMetadata(entity)
		if err != nil {
			return card, err
		}
		card.Authoring = &metadata
		byName := map[string]LibraryAuthoringSlot{}
		for _, slot := range metadata.Slots {
			byName[slot.Name] = slot
		}
		card.Arrays, card.ValueSchema = entity.Template.Arrays, entity.Template.ValueSchema
		for _, slot := range entity.Template.Slots {
			info := byName[slot.Name]
			card.Slots = append(card.Slots, LibrarySelectionSlot{Name: slot.Name, Kind: slot.Kind, SourcePointer: slot.SourcePointer, AllowEmpty: slot.AllowEmpty, Alias: info.Alias, Role: info.Role, Description: info.Description, Classification: info.Classification, Capacity: info.Capacity})
		}
	}
	for _, zone := range entity.Discovery.Zones {
		mapped := LibrarySelectionZone{LibraryContentZone: zone, Bindings: []string{}}
		for _, slot := range card.Slots {
			if slot.SourcePointer == zone.SourcePointer || strings.HasPrefix(slot.SourcePointer, zone.SourcePointer+"/") {
				mapped.Bindings = append(mapped.Bindings, slot.Name)
			}
		}
		card.Zones = append(card.Zones, mapped)
	}
	return card, nil
}

// Older indexes retain the pinned source slide even before authoring metadata
// was added. Hydrate from that source rather than declaring its slots unknown.
func (index *LibraryIndex) entityAuthoringMetadata(entity LibraryEntity) (LibraryAuthoring, error) {
	def := *entity.Template
	if def.Authoring == nil {
		var stored struct {
			SourceSlide json.RawMessage `json:"source_slide"`
		}
		if err := json.Unmarshal(entity.Definition, &stored); err != nil {
			return LibraryAuthoring{}, err
		}
		def.RawSlide = stored.SourceSlide
		obj, err := libraryObject(def.RawSlide)
		if err != nil {
			return LibraryAuthoring{}, err
		}
		source, err := Load(index.Options.Bundle, index.Options.Source)
		if err != nil {
			return LibraryAuthoring{}, err
		}
		metadata, err := libraryAuthoringMetadata(def, obj, source)
		if err != nil {
			return LibraryAuthoring{}, err
		}
		def.Authoring = &metadata
	}
	metadata := *def.Authoring
	metadata.Slots = append([]LibraryAuthoringSlot(nil), metadata.Slots...)
	typography, err := NewTypography(filepath.Join(index.Options.Bundle, "fonts"))
	if err != nil {
		return metadata, err
	}
	for i := range metadata.Slots {
		metadata.Slots[i].Capacity, err = EstimateLibrarySlotCapacity(typography, metadata.Slots[i].Capacity)
		if err != nil {
			return metadata, err
		}
	}
	return metadata, nil
}

type LibraryFindSummaryHit struct {
	Kind             string                `json:"kind"`
	Namespace        string                `json:"namespace"`
	Lifecycle        string                `json:"lifecycle"`
	Revision         int                   `json:"revision,omitempty"`
	ReuseStatus      string                `json:"reuse_status,omitempty"`
	Capacity         json.RawMessage       `json:"capacity,omitempty"`
	ID               string                `json:"id"`
	GroupID          string                `json:"group_id,omitempty"`
	VariantIDs       []string              `json:"variant_ids,omitempty"`
	Key              string                `json:"key"`
	Name             string                `json:"name"`
	Purpose          string                `json:"purpose"`
	Score            int                   `json:"score"`
	ScenarioScore    int                   `json:"scenario_score"`
	Reasons          []string              `json:"reasons"`
	UnmatchedHints   []string              `json:"unmatched_hints,omitempty"`
	Structures       []string              `json:"structures"`
	Groups           []LibraryContentGroup `json:"content_groups,omitempty"`
	ScreenshotPaths  []string              `json:"screenshot_paths"`
	FitStatus        string                `json:"fit_status"`
	StructuralStatus string                `json:"structural_status"`
	Retrieval        *LibraryRetrievalHit  `json:"retrieval,omitempty"`
}

type LibraryFindSummary struct {
	Schema    string                  `json:"schema"`
	Query     LibraryIndexFindOptions `json:"query"`
	Matches   []LibraryFindSummaryHit `json:"matches"`
	Policy    []string                `json:"policy"`
	Retrieval LibraryRetrievalReport  `json:"retrieval"`
}

func (index *LibraryIndex) FindSummary(options LibraryIndexFindOptions) (LibraryFindSummary, error) {
	result, err := index.Find(options)
	if err != nil {
		return LibraryFindSummary{}, err
	}
	out := LibraryFindSummary{Schema: "pptxgengo.library-search-summary.v1", Query: options, Matches: []LibraryFindSummaryHit{}, Policy: result.Policy, Retrieval: result.Retrieval}
	for _, hit := range result.Matches {
		card, err := index.SelectionCard(hit.Entity.ID)
		if err != nil {
			return out, err
		}
		out.Matches = append(out.Matches, LibraryFindSummaryHit{Kind: card.Kind, Namespace: card.Namespace, Lifecycle: card.Lifecycle, Revision: card.Revision, ReuseStatus: card.ReuseStatus, Capacity: card.Capacity, ID: card.ID, GroupID: hit.GroupID, VariantIDs: hit.VariantIDs, Key: hit.Entity.Key, Name: card.Name, Purpose: card.Purpose, Score: hit.Score, ScenarioScore: hit.ScenarioScore, Reasons: hit.Reasons, UnmatchedHints: hit.UnmatchedHints, Structures: card.Discovery.Structures, Groups: card.Discovery.Groups, ScreenshotPaths: card.ScreenshotPaths, FitStatus: hit.FitStatus, StructuralStatus: hit.StructuralStatus, Retrieval: hit.Retrieval})
	}
	return out, nil
}
