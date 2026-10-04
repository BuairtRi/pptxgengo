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

func TestLibraryMatchNeedsCopyDraftContainsOnlySuppliedContent(t *testing.T) {
	page := matchPage()
	page.Items = page.Items[:2]
	page.Source = "Source with no destination"
	out := filepath.Join(t.TempDir(), "drafts")
	report, err := MatchPage(page, LibraryMatchOptions{Bundle: bundle(t), Templates: []string{"cards/3", "cards/4"}, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed != 0 || report.NeedsCopy != 2 || report.Candidates[0].Template != "cards/3" {
		t.Fatalf("near misses not ranked: %+v", report)
	}
	for _, candidate := range report.Candidates {
		if candidate.Status != "needs_copy" || candidate.MappingComplete || candidate.AuthoredSlide != nil || candidate.BoundSlide != nil || candidate.DraftSlide == nil || candidate.DraftSlideFile == "" || candidate.Deck != "" {
			t.Fatal("draft claimed ready", candidate)
		}
		if len(candidate.UnresolvedFields) != 1 || candidate.UnresolvedFields[0] != "/source" {
			t.Fatal("unmapped source lost", candidate.Disposition)
		}
		data, err := os.ReadFile(candidate.DraftSlideFile)
		if err != nil || !strings.Contains(string(data), "Draft needs supplied copy") || strings.Contains(string(data), "Validate the decision") {
			t.Fatal("draft missing or borrowed source copy", err)
		}
		for _, d := range candidate.Disposition {
			if d.SourcePointer == "/items/0/text" && d.Status != "mapped_visible" {
				t.Fatal("partial supplied items were not carried")
			}
		}
	}
}

func TestLibraryMatchNestedPhasesAndExplicitComparison(t *testing.T) {
	page := PageSpec{Title: "A clear phased delivery plan", Eyebrow: "Approach", Relationship: "sequence"}
	for i := 0; i < 3; i++ {
		phase := PageItem{Lead: "Phase", Objective: "An objective."}
		for j := 0; j < 4; j++ {
			phase.Activities = append(phase.Activities, PageItem{Lead: "Activity", Text: "Perform the work."})
		}
		page.Items = append(page.Items, phase)
	}
	candidate, err := MapPageToTemplate(page, matchDef(t, "lifecycle/three-phases"), bundle(t), wmdesign.CandidateEngine, "phases", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if !candidate.MappingComplete || candidate.Status != "go_layout_succeeded_native_review_pending" {
		t.Fatalf("phase activities did not map: %+v", candidate)
	}
	page.Items[0].Owner = "Finance"
	candidate, err = MapPageToTemplate(page, matchDef(t, "lifecycle/three-phases"), bundle(t), wmdesign.CandidateEngine, "phases", 2026)
	if err != nil || candidate.Status != "needs_copy" || len(candidate.UnresolvedFields) != 1 || candidate.UnresolvedFields[0] != "/items/0/owner" {
		t.Fatal("unsupported owner vanished", err, candidate.UnresolvedFields)
	}
	page = PageSpec{Title: "Compare options against agreed criteria", Eyebrow: "Options", Relationship: "comparison", Comparison: &PageComparison{CriterionLabel: "Criterion", WeightLabel: "Weight", Legend: "Scores reflect the supplied assessment."}}
	for i := 0; i < 7; i++ {
		page.Comparison.Criteria = append(page.Comparison.Criteria, PageCriterion{Label: "Criterion", Weight: "10%"})
	}
	for i := 0; i < 4; i++ {
		page.Comparison.Options = append(page.Comparison.Options, PageOption{Name: "Option", Values: []string{"1", "2", "3", "4", "3", "2", "1"}})
	}
	candidate, err = MapPageToTemplate(page, matchDef(t, "vendors/scorecard-generic"), bundle(t), wmdesign.CandidateEngine, "comparison", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if !candidate.MappingComplete || candidate.Status != "go_layout_succeeded_native_review_pending" {
		t.Fatalf("explicit comparison did not map: %+v", candidate)
	}
}

func TestLibraryMatchFitFailureDoesNotExposeReadySlide(t *testing.T) {
	page := matchPage()
	page.Items[0].Text = strings.Repeat("This copy cannot fit inside the fixed card row. ", 150)
	candidate, err := MapPageToTemplate(page, matchDef(t, "cards/3"), bundle(t), wmdesign.CandidateEngine, "overflow", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if !candidate.MappingComplete || candidate.Status == "go_layout_succeeded_native_review_pending" || candidate.AuthoredSlide != nil || candidate.BoundSlide != nil || candidate.Deck != "" {
		t.Fatal("layout failure exposed ready-use slide", candidate.Status, candidate.Error)
	}
}
