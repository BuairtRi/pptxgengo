package wmdesign

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func discoveryPinnedCatalog(t *testing.T) []LibraryTemplate {
	t.Helper()
	catalog, err := LibraryCatalog(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "")
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func discoveryPinnedTemplate(t *testing.T, catalog []LibraryTemplate, key string) LibraryTemplate {
	t.Helper()
	for _, def := range catalog {
		if def.Key == key {
			return def
		}
	}
	t.Fatalf("missing pinned specimen %s", key)
	return LibraryTemplate{}
}

func discoveryHas(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func TestLibraryDiscoverySceneNodesExcludeCellData(t *testing.T) {
	catalog := discoveryPinnedCatalog(t)
	raid := discoveryPinnedTemplate(t, catalog, "status/raid-log")
	for _, dataType := range []string{"Risk", "Issue", "Assumption", "Dependency"} {
		if discoveryHas(raid.Discovery.ComponentTypes, dataType) {
			t.Fatalf("table data discriminator %s became a component", dataType)
		}
	}
	if !discoveryHas(raid.Discovery.ComponentTypes, "table") {
		t.Fatal("actual RAID table node was lost")
	}
	for _, def := range catalog {
		for _, zone := range def.Discovery.Zones {
			parts := strings.Split(zone.SourcePointer, "/")
			if len(parts) != 3 || parts[1] != "body" {
				t.Fatalf("%s advertises component data as a scene zone: %+v", def.Key, zone)
			}
		}
	}
}

func TestLibraryDiscoveryBeforeAfterGroupsAreCompleteRows(t *testing.T) {
	def := discoveryPinnedTemplate(t, discoveryPinnedCatalog(t), "transformation/before-after")
	// The current template uses separate scene blocks. Keep the beforeafter
	// component's row-versus-cell regression as an explicit synthetic fixture.
	def.RawSlide = json.RawMessage(`{"type":"slide","body":[{"type":"beforeafter","x":0,"y":0,"w":500,"h":200,"rows":[["A","B","C"],["D","E","F"],["G","H","I"],["J","K","L"]]}]}`)
	obj, err := libraryObject(def.RawSlide)
	if err != nil {
		t.Fatal(err)
	}
	def.Slots, def.Arrays = nil, nil
	libraryContentWalk(&def, obj["body"].([]any)[0], "/body/0", "node01", "", libraryProjectionContext{})
	def.Discovery = libraryDiscovery(def, obj)
	rows := 0
	for _, group := range def.Discovery.Groups {
		if strings.HasPrefix(group.SourcePointer, "/body/0/rows/") {
			t.Fatalf("row cell tuple became a semantic group: %+v", group)
		}
		if group.SourcePointer == "/body/0/rows" {
			if group.ExactCount != 4 || group.Role != "relationship" || group.Scope != "primary" {
				t.Fatalf("unexpected complete-row group: %+v", group)
			}
			rows++
		}
	}
	if rows != 1 {
		t.Fatalf("want exactly one complete-row group, got %d", rows)
	}
	three, err := SearchLibrary([]LibraryTemplate{def}, LibrarySearchOptions{Items: 3, ItemRole: "relationship"})
	if err != nil || len(three.Matches) == 0 || three.Matches[0].CountMatch != nil {
		t.Fatalf("three fields per row matched three relationships: %+v, %v", three, err)
	}
	four, err := SearchLibrary([]LibraryTemplate{def}, LibrarySearchOptions{Items: 4, ItemRole: "relationship"})
	if err != nil || len(four.Matches) == 0 || four.Matches[0].CountMatch == nil || four.Matches[0].CountMatch.Group.ExactCount != 4 {
		t.Fatalf("four complete relationships did not match: %+v, %v", four, err)
	}
}

func TestLibraryDiscoveryPrimaryCardsOutrankNestedSupport(t *testing.T) {
	catalog := discoveryPinnedCatalog(t)
	keys := []string{"cards/3", "about/glance", "architecture/reference", "bio-full/portrait-list", "bio-full/portrait-nav", "bio-full/portrait-quote", "bios/leadership-specialists"}
	var specimens []LibraryTemplate
	for _, key := range keys {
		specimens = append(specimens, discoveryPinnedTemplate(t, catalog, key))
	}
	result, err := SearchLibrary(specimens, LibrarySearchOptions{Items: 3, ItemRole: "point", ContentRoles: []string{"point"}, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Matches) == 0 {
		t.Fatal("no candidates for primary three-card arrangement")
	}
	if result.Matches[0].Template.Key != "cards/3" || result.Matches[0].CountMatch == nil || result.Matches[0].CountMatch.Group.Scope != "primary" {
		t.Fatalf("primary three-card arrangement was not first: %+v", result.Matches[0])
	}
	for _, hit := range result.Matches[1:] {
		if hit.Score >= result.Matches[0].Score {
			t.Fatalf("supporting text in %s competes with primary cards: %d >= %d", hit.Template.Key, hit.Score, result.Matches[0].Score)
		}
		if hit.CountMatch != nil && hit.CountMatch.Group.Scope != "nested" && hit.CountMatch.Group.Scope != "supporting" {
			t.Fatalf("supporting detail in %s presented as primary points: %+v", hit.Template.Key, hit.CountMatch)
		}
	}
}

func TestLibraryDiscoveryPairedComparisonsDoNotDependOnLabels(t *testing.T) {
	catalog := discoveryPinnedCatalog(t)
	for _, key := range []string{"comparison/us-versus-others", "comparison/us-versus-others-four"} {
		t.Run(key, func(t *testing.T) {
			def := discoveryPinnedTemplate(t, catalog, key)
			if !discoveryHas(def.Discovery.Structures, "comparison") || len(def.Discovery.Relationships) == 0 {
				t.Fatalf("paired comparison not discovered: %+v", def.Discovery)
			}
			obj, err := libraryObject(def.RawSlide)
			if err != nil {
				t.Fatal(err)
			}
			def.Key, def.Name, def.Purpose, def.Uses = "opaque/arrangement", "", "", nil
			obj["title"], obj["eyebrow"] = "New message", "New topic"
			for _, raw := range obj["body"].([]any) {
				if node, ok := raw.(map[string]any); ok {
					if _, exists := node["label"]; exists {
						node["label"] = "Renamed category"
					}
				}
			}
			def.Discovery = libraryDiscovery(def, obj)
			result, err := SearchLibrary([]LibraryTemplate{def}, LibrarySearchOptions{Structures: []string{"comparison"}})
			if err != nil || len(result.Matches) == 0 || result.Matches[0].Score != 16 {
				t.Fatalf("renamed paired arrangement lost structural match: %+v, %v", result, err)
			}
			// A connector alone cannot establish the paired relationship.
			for _, raw := range obj["body"].([]any) {
				if node, ok := raw.(map[string]any); ok && node["type"] == "card" {
					node["body"] = nil
				}
			}
			if discoveryHas(libraryDiscovery(def, obj).Structures, "comparison") {
				t.Fatal("connector without parallel lists established comparison")
			}
		})
	}
}

func TestLibraryDiscoverySearchDeterministicAndPermissive(t *testing.T) {
	catalog := discoveryPinnedCatalog(t)
	options := LibrarySearchOptions{EngineHint: CandidateEngine, Items: 4, ItemRole: "point", ContentRoles: []string{"point", "key-message"}, VisualForms: []string{"image", "icon"}, Query: "unrelated client terminology", Limit: 100}
	a, err := SearchLibrary(catalog, options)
	if err != nil {
		t.Fatal(err)
	}
	reversed := append([]LibraryTemplate(nil), catalog...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	b, err := SearchLibrary(reversed, options)
	if err != nil {
		t.Fatal(err)
	}
	aj, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	bj, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(aj, bj) {
		t.Fatal("search result depends on catalog input order")
	}
	if len(a.Matches) != 100 {
		t.Fatalf("unmatched scenario terms filtered candidates: %d", len(a.Matches))
	}
	if a.Engine.Requested != CandidateEngine || a.Engine.Compatibility != "not_evaluated_search_only" {
		t.Fatalf("search implied engine compatibility it did not evaluate: %+v", a.Engine)
	}
	for _, hit := range a.Matches {
		if hit.Template.Status == "deprecated" || hit.FitStatus != "not_measured_for_query" || hit.Template.Discovery.Capability.BuildForQuery != "not_executed" {
			t.Fatalf("incorrect lifecycle or qualification claim: %+v", hit)
		}
	}
}

func TestLibraryDiscoveryWeeklyStatusRanksDirectScenarioAboveIncidentalCopy(t *testing.T) {
	catalog := discoveryPinnedCatalog(t)
	result, err := SearchLibrary(catalog, LibrarySearchOptions{Query: "weekly status", Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Matches) == 0 {
		t.Fatal("weekly status returned no candidates")
	}
	// v5 authors a weekly trend and site status table in adoption/dashboard.
	// Generic status labels on a static capability map are a partial match.
	first := result.Matches[0].Template
	if first.Key != "adoption/dashboard" || !strings.Contains(first.Purpose, "weekly trend") || !strings.Contains(first.Purpose, "status table") {
		t.Fatalf("weekly status did not prefer a direct status arrangement: %+v", result.Matches[0])
	}
	firstScore := result.Matches[0].Score
	for _, hit := range result.Matches {
		if hit.Template.Key == "confidential/notice" || hit.Template.Key == "runbook/go-no-go" || hit.Template.Key == "capability-map/domain-columns" {
			if hit.Score >= firstScore {
				t.Fatalf("incidental status mention in %s outranks direct scenario: %d >= %d", hit.Template.Key, hit.Score, firstScore)
			}
		}
	}
	// Scenario terms remain soft signals: a structural request still discovers
	// source arrangements whose original client/domain wording is unrelated.
	structural, err := SearchLibrary(catalog, LibrarySearchOptions{Query: "unrelated weekly context", Items: 3, ItemRole: "point", Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, hit := range structural.Matches {
		if hit.Template.Key == "cards/3" {
			found = hit.CountMatch != nil && hit.CountMatch.Group.Scope == "primary"
		}
	}
	if !found {
		t.Fatal("scenario wording gated a valid primary three-point arrangement")
	}
}

func TestLibraryDiscoveryCompleteScenarioCoverageKeepsTextBudget(t *testing.T) {
	complete := LibraryTemplate{TemplateDefinition: TemplateDefinition{Key: "opaque/complete"}, Purpose: "Weekly status dashboard"}
	partial := LibraryTemplate{TemplateDefinition: TemplateDefinition{Key: "weekly/status"}}
	for _, query := range []string{"weekly status dashboard", "weekly weekly status dashboard"} {
		result, err := SearchLibrary([]LibraryTemplate{partial, complete}, LibrarySearchOptions{Query: query})
		if err != nil || len(result.Matches) != 2 {
			t.Fatalf("search failed: %v, %+v", err, result)
		}
		if result.Matches[0].Template.Key != complete.Key || result.Matches[0].Score != 12 || result.Matches[1].Score != 12 {
			t.Fatalf("full coverage did not break capped-score tie: %+v", result.Matches)
		}
		if result.Matches[0].ScenarioScore >= result.Matches[1].ScenarioScore {
			t.Fatal("fixture no longer checks coverage before uncapped identity weights")
		}
	}
	result, err := SearchLibrary([]LibraryTemplate{complete}, LibrarySearchOptions{Query: "weekly"})
	if err != nil || len(result.Matches) != 1 || result.Matches[0].Score != 3 {
		t.Fatalf("single-token purpose match received a completeness bonus: %v %+v", err, result)
	}
	partial.Discovery.Structures = []string{"matrix"}
	result, err = SearchLibrary([]LibraryTemplate{complete, partial}, LibrarySearchOptions{Query: "weekly status dashboard", Structures: []string{"matrix"}})
	if err != nil || len(result.Matches) != 2 || result.Matches[0].Template.Key != partial.Key || result.Matches[0].Score != 28 {
		t.Fatalf("scenario bonus overwhelmed independent structural hints: %v %+v", err, result)
	}
}
