package wmdesign

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestScenarioCoverageBreaksSaturatedTies(t *testing.T) {
	defs := []LibraryTemplate{
		{TemplateDefinition: TemplateDefinition{Key: "aaa/general"}, Name: "Modernization", Purpose: "Economics"},
		{TemplateDefinition: TemplateDefinition{Key: "decision/buy-build-economics"}, Name: "Modernization economics buy build", Purpose: "Compare when each option fits"},
	}
	result, err := SearchLibrary(defs, LibrarySearchOptions{Query: "modernization economics buy build", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Matches[0].Template.Key != "decision/buy-build-economics" || result.Matches[0].ScenarioScore <= result.Matches[1].ScenarioScore {
		t.Fatalf("lost full query coverage: %+v", result.Matches)
	}
	// Scenario remains subordinate to explicit shape affordances.
	defs[0].Discovery.Structures = []string{"comparison"}
	result, err = SearchLibrary(defs, LibrarySearchOptions{Query: "modernization economics buy build", Structures: []string{"comparison"}, Limit: 2})
	if err != nil || result.Matches[0].Template.Key != "aaa/general" {
		t.Fatalf("scenario overrode shape: %v %+v", err, result.Matches)
	}
}

func TestSelectionCardsAndSQLiteZoneViews(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.sqlite")
	_, err := BuildLibraryIndex(path, LibraryIndexOptions{Bundle: filepath.Join("..", "..", "library", "wm-design-system", "v5")})
	if err != nil {
		t.Fatal(err)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	card, err := index.SelectionCard("decision/buy-build-economics")
	if err != nil {
		t.Fatal(err)
	}
	if len(card.Zones) == 0 || len(card.Slots) == 0 || card.Preparation == "" || len(card.ScreenshotPaths) != 0 {
		t.Fatalf("incomplete or fabricated card: %+v", card)
	}
	matched := false
	for _, zone := range card.Zones {
		for _, name := range zone.Bindings {
			for _, slot := range card.Slots {
				if slot.Name == name && strings.HasPrefix(slot.SourcePointer, zone.SourcePointer+"/") {
					matched = true
				}
			}
		}
	}
	if !matched {
		t.Fatal("missing zone-to-value bindings")
	}
	for view, want := range map[string]int{"content_zones": len(card.Zones), "content_slots": len(card.Slots), "artifacts": 0} {
		var count int
		if err = index.db.QueryRow("SELECT count(*) FROM "+view+" WHERE entity_id=?", card.ID).Scan(&count); err != nil || count != want {
			t.Fatalf("view %s: %d != %d: %v", view, count, want, err)
		}
	}
	result, err := index.FindSummary(LibraryIndexFindOptions{Kinds: []string{"template"}, Shape: LibrarySearchOptions{Query: "buy build economics", Limit: 3}})
	if err != nil || len(result.Matches) != 3 || result.Matches[0].Key != card.Key {
		t.Fatalf("summary search: %v %+v", err, result)
	}
	if strings.Contains(string(indexJSON(card)), "synthetic_source_example") {
		t.Fatal("summary leaked specimen copy")
	}
}
