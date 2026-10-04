package wmdesign

import "strings"

// Selection cards expose the visual and semantic contract without source
// specimen copy. Bindings identify the editable values belonging to each zone.
type LibrarySelectionSlot struct {
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	SourcePointer string `json:"source_pointer"`
	AllowEmpty    bool   `json:"allow_empty"`
}

type LibrarySelectionZone struct {
	LibraryContentZone
	Bindings []string `json:"bindings"`
}

type LibrarySelectionCard struct {
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
	card := LibrarySelectionCard{Schema: "pptxgengo.library-selection.v1", ID: entity.ID, Key: entity.Key, Name: entity.Name, Purpose: entity.Purpose, SourceRevision: entity.SourceRevision, SourceSHA256: entity.SourceSHA256, Discovery: entity.Discovery, ScreenshotPaths: []string{}, Zones: []LibrarySelectionZone{}, Slots: []LibrarySelectionSlot{}, Preparation: preview.Preparation, Policy: []string{"Screenshot paths are verified against their pinned hashes.", "Zone bounds are source scene coordinates in points; they are not measured fit for supplied content.", "Preserve the slide argument, inspect the screenshot and map content to named bindings before building."}}
	card.Discovery.Zones = nil
	for i, artifact := range preview.Artifacts {
		if artifact.Role == "source_preview" || artifact.Role == "alternate_preview" {
			card.ScreenshotPaths = append(card.ScreenshotPaths, preview.Paths[i])
		}
	}
	if entity.Template != nil {
		card.Arrays, card.ValueSchema = entity.Template.Arrays, entity.Template.ValueSchema
		for _, slot := range entity.Template.Slots {
			card.Slots = append(card.Slots, LibrarySelectionSlot{slot.Name, slot.Kind, slot.SourcePointer, slot.AllowEmpty})
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

type LibraryFindSummaryHit struct {
	ID              string                `json:"id"`
	Key             string                `json:"key"`
	Name            string                `json:"name"`
	Purpose         string                `json:"purpose"`
	Score           int                   `json:"score"`
	ScenarioScore   int                   `json:"scenario_score"`
	Reasons         []string              `json:"reasons"`
	UnmatchedHints  []string              `json:"unmatched_hints,omitempty"`
	Structures      []string              `json:"structures"`
	Groups          []LibraryContentGroup `json:"content_groups,omitempty"`
	ScreenshotPaths []string              `json:"screenshot_paths"`
	FitStatus       string                `json:"fit_status"`
}

type LibraryFindSummary struct {
	Schema  string                  `json:"schema"`
	Query   LibraryIndexFindOptions `json:"query"`
	Matches []LibraryFindSummaryHit `json:"matches"`
	Policy  []string                `json:"policy"`
}

func (index *LibraryIndex) FindSummary(options LibraryIndexFindOptions) (LibraryFindSummary, error) {
	result, err := index.Find(options)
	if err != nil {
		return LibraryFindSummary{}, err
	}
	out := LibraryFindSummary{Schema: "pptxgengo.library-search-summary.v1", Query: options, Matches: []LibraryFindSummaryHit{}, Policy: result.Policy}
	for _, hit := range result.Matches {
		card, err := index.SelectionCard(hit.Entity.ID)
		if err != nil {
			return out, err
		}
		out.Matches = append(out.Matches, LibraryFindSummaryHit{ID: card.ID, Key: card.Key, Name: card.Name, Purpose: card.Purpose, Score: hit.Score, ScenarioScore: hit.ScenarioScore, Reasons: hit.Reasons, UnmatchedHints: hit.UnmatchedHints, Structures: card.Discovery.Structures, Groups: card.Discovery.Groups, ScreenshotPaths: card.ScreenshotPaths, FitStatus: hit.FitStatus})
	}
	return out, nil
}
