package deckproject

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func matchDef(t *testing.T, key string) wmdesign.LibraryTemplate {
	t.Helper()
	catalog, e := wmdesign.LibraryCatalog(bundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	for _, def := range catalog {
		if def.Key == key {
			return def
		}
	}
	t.Fatal("missing test template")
	return wmdesign.LibraryTemplate{}
}
func matchPage() PageSpec {
	return PageSpec{Title: "Three controls guide product delivery", Eyebrow: "Delivery", Relationship: "parallel", Items: []PageItem{{ID: "trace", Lead: "Trace", Text: "Keep evidence."}, {ID: "review", Lead: "Review", Text: "Agree priorities."}, {ID: "ship", Lead: "Ship", Text: "Release useful work."}}}
}
func TestLibraryMatchTypedCardsAndReadySlide(t *testing.T) {
	out := filepath.Join(t.TempDir(), "candidates")
	report, e := MatchPage(matchPage(), LibraryMatchOptions{Bundle: bundle(t), Templates: []string{"cards/3"}, Out: out, Limit: 4, Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	if report.Passed != 1 || report.Failed != 0 || report.CombinedDeck == "" {
		t.Fatalf("failed candidate: %+v", report)
	}
	c := report.Candidates[0]
	if !c.MappingComplete || c.AuthoredSlide == nil || c.BoundSlide == nil {
		t.Fatal("missing reusable mapped slide")
	}
	for _, d := range c.Disposition {
		if d.Status != "mapped_visible" || d.Destination == "" {
			t.Fatalf("content discarded: %+v", d)
		}
	}
	for _, path := range []string{c.SlideFile, c.Deck, c.LayoutReport, report.CombinedDeck} {
		if st, e := os.Stat(path); e != nil || st.Size() == 0 {
			t.Fatalf("missing artifact %s", path)
		}
	}
	data, e := os.ReadFile(c.SlideFile)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(data), "supplied_content") || strings.Contains(string(data), "example-001") {
		t.Fatal("source specimen identity leaked")
	}
	authored := c.AuthoredSlide
	p := &Project{SourcePath: "candidate.yaml", Positions: map[string]Position{}, positionFiles: map[string]string{}}
	if e = p.expandContentAliases(authored, "/slide"); e != nil {
		t.Fatal(e)
	}
	var slide Slide
	if e = json.Unmarshal(canonical(authored), &slide); e != nil {
		t.Fatal(e)
	}
	if slide.ContentKind != "supplied_content" || slide.Template.Scope != "shared" || slide.Template.ID != "cards/3" {
		t.Fatal("candidate is not genuine stock supplied content")
	}
}
func TestLibraryMatchNoTruncationOrMissingSource(t *testing.T) {
	page := matchPage()
	page.Source = "Discovery interviews, September 2026"
	c, e := MapPageToTemplate(page, matchDef(t, "cards/3"), bundle(t), wmdesign.CandidateEngine, "one", 2026)
	if e != nil {
		t.Fatal(e)
	}
	if c.MappingComplete || len(c.UnresolvedFields) != 1 || c.UnresolvedFields[0] != "/source" || c.AuthoredSlide != nil {
		t.Fatalf("unrepresented source accepted: %+v", c)
	}
	page = matchPage()
	page.Items = page.Items[:2]
	c, e = MapPageToTemplate(page, matchDef(t, "cards/3"), bundle(t), wmdesign.CandidateEngine, "one", 2026)
	if e != nil {
		t.Fatal(e)
	}
	if c.MappingComplete || len(c.MissingSlots) == 0 {
		t.Fatal("cardinality gap accepted")
	}
	page = matchPage()
	page.Relationship = "sequence"
	c, e = MapPageToTemplate(page, matchDef(t, "cards/3"), bundle(t), wmdesign.CandidateEngine, "one", 2026)
	if e != nil {
		t.Fatal(e)
	}
	if c.MappingComplete || c.Error == "" {
		t.Fatal("relationship gap accepted")
	}
}
func TestLibraryMatchStrictSemanticInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.yaml")
	if e := os.WriteFile(path, []byte("title: A clear claim\nunknown: extra\n"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := LoadPageSpec(path); e == nil {
		t.Fatal("unknown semantic field accepted")
	}
}

func TestLibraryMatchAutoFindsContentCandidates(t *testing.T) {
	report, e := MatchPage(matchPage(), LibraryMatchOptions{Bundle: bundle(t), Limit: 4, Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	if report.Passed != 1 || len(report.Candidates) <= 100 || report.LibraryGap == "" {
		for _, c := range report.Candidates {
			t.Logf("%s: %s %s missing %v", c.Template, c.Status, c.Error, c.MissingSlots)
		}
		t.Fatal("automatic search did not evaluate full catalog and report the genuine candidate gap")
	}
	t.Logf("automatic candidates passed=%d failed=%d", report.Passed, report.Failed)
}

func TestLibraryMatchTitleOnlyDoesNotBorrowBusinessCopy(t *testing.T) {
	report, err := MatchPage(PageSpec{Title: "A supplied claim"}, LibraryMatchOptions{Bundle: bundle(t), Limit: 4, Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed != 0 || report.LibraryGap == "" {
		t.Fatal("title alone filled source business copy")
	}
	for _, candidate := range report.Candidates {
		if candidate.AuthoredSlide != nil || candidate.BoundSlide != nil || candidate.MappingComplete {
			t.Fatal("unfilled source specimen published", candidate.Template)
		}
	}
}

func TestLibraryMatchReportsComplexRelationshipGaps(t *testing.T) {
	for _, example := range []struct{ key, relationship string }{{"lifecycle/three-phases", "sequence"}, {"capability-heat/dense", "table"}, {"decision/buy-build-economics", "comparison"}} {
		page := matchPage()
		page.Relationship = example.relationship
		page.Source = "Supplied evidence."
		candidate, err := MapPageToTemplate(page, matchDef(t, example.key), bundle(t), wmdesign.CandidateEngine, "gap", 2026)
		if err != nil {
			t.Fatal(err)
		}
		if candidate.MappingComplete || candidate.AuthoredSlide != nil || len(candidate.MissingSlots)+len(candidate.UnresolvedFields) == 0 {
			t.Fatal("complex schema gap silently filled", example.key)
		}
		for _, d := range candidate.Disposition {
			if d.Status != "mapped_visible" && d.Status != "unresolved" {
				t.Fatal("unreported content disposition", d)
			}
		}
	}
}

func TestMatcherRejectsUncertainAllTextParallelInference(t *testing.T) {
	def := wmdesign.LibraryTemplate{TemplateDefinition: wmdesign.TemplateDefinition{Key: "unknown/text"}, RawSlide: json.RawMessage(`{"body":[{"type":"text","style":"subhead","text":"Specimen lead"},{"type":"text","style":"body","text":"Specimen body"}]}`), Slots: []wmdesign.LibrarySlot{{Name: "title", SourcePointer: "/title", Kind: "string"}, {Name: "lead", SourcePointer: "/body/0/text", Kind: "string"}, {Name: "body", SourcePointer: "/body/1/text", Kind: "string"}}}
	page := PageSpec{Title: "Claim", Relationship: "parallel", Items: []PageItem{{Lead: "Lead", Text: "Text"}}}
	c, e := MapPageToTemplate(page, def, bundle(t), wmdesign.CandidateEngine, "one", 2026)
	if e != nil {
		t.Fatal(e)
	}
	if c.MappingComplete || c.Error == "" {
		t.Fatal("uncertain all-text topology treated as parallel")
	}
}
