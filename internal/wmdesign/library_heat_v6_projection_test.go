package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func heatProjectionFixture(t *testing.T, revision string) LibraryTemplate {
	t.Helper()
	raw := json.RawMessage(`{"title":"Claim","body":[{"type":"table","heatMin":1,"groupW":24,"cols":[{"k":"c","label":"Capability","w":180},{"k":"p","label":"Priority","type":"priority","w":84},{"k":"nt","label":"Evidence","type":"bullets","size":"small","w":300}],"rowGroups":[{"label":"Front door","from":0,"to":0,"fill":"#50658E"}],"rows":[{"h":52,"c":{"text":"Referral","sub":"Front door","ref":"A1","refActive":true},"p":{"value":"p1","label":"Urgent"},"nt":["First observation","Second qualifier"]}]}]}`)
	def := LibraryTemplate{TemplateDefinition: TemplateDefinition{Key: "heat-notes/test", SourceRevision: revision}, Family: "heatmaps", RawSlide: raw}
	obj, err := libraryObject(raw)
	if err != nil {
		t.Fatal(err)
	}
	libraryContentWalk(&def, obj, "", "", "", libraryProjectionContext{})
	return def
}

func TestHeatV6ProjectionKeepsAllAuthoredFieldsAndStructuralRanges(t *testing.T) {
	def := heatProjectionFixture(t, LibraryRevisionV6)
	wanted := map[string]string{
		"/body/0/rows/0/c/ref": "string", "/body/0/rows/0/c/refActive": "boolean",
		"/body/0/rows/0/p/value": "string", "/body/0/rows/0/p/label": "string",
		"/body/0/rowGroups/0/label": "string", "/body/0/rows/0/nt/0": "string", "/body/0/rows/0/nt/1": "string",
	}
	for _, slot := range def.Slots {
		if kind, ok := wanted[slot.SourcePointer]; ok {
			if slot.Kind != kind {
				t.Fatal("wrong editable kind", slot)
			}
			delete(wanted, slot.SourcePointer)
		}
		if strings.HasSuffix(slot.SourcePointer, "/from") || strings.HasSuffix(slot.SourcePointer, "/to") || strings.HasSuffix(slot.SourcePointer, "/fill") || strings.HasSuffix(slot.SourcePointer, "/h") || strings.HasSuffix(slot.SourcePointer, "/heatMin") {
			t.Fatal("geometry entered content contract", slot)
		}
	}
	if len(wanted) != 0 {
		t.Fatal("source business fields left frozen", wanted)
	}
	metadata, err := LibraryAuthoringMetadata(def)
	if err != nil {
		t.Fatal(err)
	}
	roles := map[string]string{"/body/0/rows/0/c/ref": "reference_id", "/body/0/rows/0/c/refActive": "reference_active", "/body/0/rows/0/p/value": "priority_value", "/body/0/rows/0/p/label": "priority_label", "/body/0/rows/0/nt/0": "evidence_body", "/body/0/rowGroups/0/label": "group_label"}
	for _, slot := range metadata.Slots {
		if role, ok := roles[slot.SourcePointer]; ok {
			if slot.Role != role || slot.ReviewStatus != "explicit_source_recipe_unreviewed" {
				t.Fatal("wrong meaning/review status", slot)
			}
			delete(roles, slot.SourcePointer)
		}
		if strings.Contains(slot.Alias, "/evidence/") && strings.HasSuffix(slot.Alias, "/0") {
			t.Fatal("numeric source path leaked into friendly evidence alias", slot.Alias)
		}
	}
	if len(roles) != 0 {
		t.Fatal("semantic fields absent", roles)
	}
	if metadata.ReviewStatus != "inferred_from_pinned_source" {
		t.Fatal("human review invented")
	}
	// V5 has no new ref/priority contract; the revision gate cannot silently
	// make previously frozen fields editable in its source projection.
	old := heatProjectionFixture(t, LibraryRevisionV5)
	for _, slot := range old.Slots {
		if strings.Contains(slot.SourcePointer, "/p/") || strings.HasSuffix(slot.SourcePointer, "/ref") || strings.HasSuffix(slot.SourcePointer, "/refActive") {
			t.Fatal("v5 contract expanded", slot)
		}
	}
}
