package wmdesign

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

func TestBrowsingTemplateCoverage(t *testing.T) {
	_, report := indexFixture(t)
	doc, coverage, e := TemplateBrowsingDocument(report.Options.Bundle, "", 2026)
	if e != nil {
		t.Fatal(e)
	}
	catalog, e := LibraryCatalog(report.Options.Bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	seen := map[string]bool{}
	frames := 0
	for _, entry := range coverage.Entries {
		if entry.Kind == "template" {
			if seen[entry.Key] {
				t.Fatal("duplicate", entry.Key)
			}
			seen[entry.Key] = true
		} else {
			frames++
			if entry.Frame == nil {
				t.Fatal(entry)
			}
		}
	}
	if len(seen) != len(catalog) || frames == 0 || len(doc.Sections) < 2 || doc.Slides[0].Title != "How to use this deck" {
		t.Fatal(len(seen), len(catalog), frames, doc.Sections)
	}
	t.Logf("templates=%d frames=%d candidates=%d structural_exclusions=%d aliases=%d total_slides=%d", len(seen), frames, coverage.FrameCandidates, len(coverage.FrameExclusions), len(coverage.FrameAliases), len(doc.Slides))
	if coverage.FrameCandidates != frames+len(coverage.FrameAliases)+len(coverage.FrameExclusions) {
		t.Fatal("matrix omission")
	}
	for _, def := range catalog {
		if !seen[def.Key] {
			t.Fatal("omitted", def.Key)
		}
	}
}

func changeBrowsingManifest(t *testing.T, root string, revision int, lifecycle, until string) {
	t.Helper()
	path := filepath.Join(root, "authored-message", string(rune('0'+revision)))
	m, e := finishedslide.Read(path)
	if e != nil {
		t.Fatal(e)
	}
	files, e := finishedslide.ReadPayload(path, m)
	if e != nil {
		t.Fatal(e)
	}
	m.Lifecycle = lifecycle
	if lifecycle == "approved" {
		files["review.json"] = []byte(`{"fixture_only":true}`)
		exists := false
		for _, file := range m.Files {
			exists = exists || file.Path == "review.json"
		}
		if !exists {
			m.Files = append(m.Files, finishedslide.File{Path: "review.json", Role: "review"})
		}
	}
	m.ValidUntil = until
	m.ReviewedAt = "2026-01-01"
	if lifecycle == "approved" {
		m.Approval = &finishedslide.Approval{By: "Explicit test fixture", Date: "2026-01-01", ReuseScope: "Generic unit tests only"}
	}
	if e = os.RemoveAll(path); e != nil {
		t.Fatal(e)
	}
	if _, e = finishedslide.Create(path, m, files); e != nil {
		t.Fatal(e)
	}
}
func TestBrowsingLatestApprovedAndWithdrawal(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "2026-10-07")
	root := t.TempDir()
	finishedIndexRevision(t, root, 1)
	changeBrowsingManifest(t, root, 1, "approved", "2026-12-31")
	finishedIndexRevision(t, root, 2)
	selection, e := SelectBrowsingRevisions(root, date)
	if e != nil {
		t.Fatal(e)
	}
	if !selection.Revisions[0].Included || selection.Revisions[1].Included {
		t.Fatal(selection)
	}
	changeBrowsingManifest(t, root, 2, "approved", "2026-12-31")
	selection, e = SelectBrowsingRevisions(root, date)
	if e != nil {
		t.Fatal(e)
	}
	if selection.Revisions[0].Included || !selection.Revisions[1].Included {
		t.Fatal(selection)
	}
	changeBrowsingManifest(t, root, 2, "approved", "2026-01-02")
	if _, e = SelectBrowsingRevisions(root, date); e == nil || !strings.Contains(e.Error(), "no_current_approved") {
		t.Fatal("expired newest resurrected old approval", e)
	}
	changeBrowsingManifest(t, root, 2, "deprecated", "2026-12-31")
	if _, e = SelectBrowsingRevisions(root, date); e == nil || !strings.Contains(e.Error(), "no_current_approved") {
		t.Fatal("withdrawal ignored", e)
	}
}

func TestBrowsingCurrentPublishedCoverage(t *testing.T) {
	bundle := "../../library/wm-design-system/v11"
	doc, coverage, e := TemplateBrowsingDocument(bundle, "", 2026)
	if e != nil {
		t.Fatal(e)
	}
	templates, frames := 0, 0
	for _, entry := range coverage.Entries {
		if entry.Kind == "template" {
			templates++
		} else {
			frames++
		}
	}
	catalog, e := LibraryCatalog(bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	if templates != len(catalog) || coverage.FrameCandidates != frames+len(coverage.FrameAliases)+len(coverage.FrameExclusions) {
		t.Fatal("coverage omissions")
	}
	t.Logf("current published: templates=%d frames=%d candidates=%d structural_exclusions=%d aliases=%d total_slides=%d", templates, frames, coverage.FrameCandidates, len(coverage.FrameExclusions), len(coverage.FrameAliases), len(doc.Slides))
}

func TestBrowsingFramesGenerateOptIn(t *testing.T) {
	out := os.Getenv("PPTXGENGO_BROWSING_FRAME_OUT")
	if out == "" {
		t.Skip("explicit owned artifact directory required")
	}
	if _, e := os.Lstat(out); !os.IsNotExist(e) {
		t.Fatal("output must be absent")
	}
	doc, coverage, e := TemplateBrowsingDocument("../../library/wm-design-system/v11", "", 2026)
	if e != nil {
		t.Fatal(e)
	}
	frames := []SlideSpec{BrowsingInformationSlide("how-to-use", "How to use this deck", "Generic frame qualification specimen only; no authored content or visual acceptance.")}
	for _, slide := range doc.Slides {
		if strings.HasPrefix(slide.ID, "frame-") && slide.ID != "frame-divider" {
			frames = append(frames, slide)
		}
	}
	doc.Slides = frames
	doc.Sections = nil
	native, report, e := BuildWithEngine("../../library/wm-design-system/v11", "", doc, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(out, 0755); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "all-frames.pptx"), native, 0644); e != nil {
		t.Fatal(e)
	}
	if e = WriteJSON(filepath.Join(out, "layout-report.json"), report); e != nil {
		t.Fatal(e)
	}
	if e = WriteJSON(filepath.Join(out, "coverage.json"), coverage); e != nil {
		t.Fatal(e)
	}
	t.Logf("generated %d frame pages + guide, %d actual PPTX bytes; native visual qualification pending", len(frames)-1, len(native))
}

func TestBrowsingFrameAliasesResolvedEquivalence(t *testing.T) {
	doc, coverage, e := TemplateBrowsingDocument("../../library/wm-design-system/v11", "", 2026)
	if e != nil {
		t.Fatal(e)
	}
	source, e := Load("../../library/wm-design-system/v11", "")
	if e != nil {
		t.Fatal(e)
	}
	slides := map[string]SlideSpec{}
	for _, slide := range doc.Slides {
		slides[slide.ID] = slide
	}
	for _, alias := range coverage.FrameAliases {
		canonical, ok := slides[alias.SlideID]
		if !ok || !canonical.Frame.NoHeader || !alias.Frame.NoHeader {
			t.Fatal("invalid alias", alias)
		}
		before, e := source.ResolveFrame(alias.Frame)
		if e != nil {
			t.Fatal(e)
		}
		after, e := source.ResolveFrame(canonical.Frame)
		if e != nil {
			t.Fatal(e)
		}
		before.Request = after.Request
		before.TitleStyle = ""
		after.TitleStyle = ""
		if !reflect.DeepEqual(before, after) || before.TitleRule != 0 || before.Header != (Rect{}) {
			t.Fatal("alias changed visible geometry", alias)
		}
	}
	if len(coverage.FrameAliases) == 0 {
		t.Fatal("no equivalent requests identified")
	}
	// Check the real renderer, including measured typography and chrome text,
	// for standard and appendix requests with header content suppressed.
	original := slides[coverage.FrameAliases[0].SlideID]
	alternative := original
	alternative.Frame = coverage.FrameAliases[0].Frame
	one := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{original}}
	two := one
	two.Slides = []SlideSpec{alternative}
	_, a, e := BuildWithEngine("../../library/wm-design-system/v11", "", one, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	_, b, e := BuildWithEngine("../../library/wm-design-system/v11", "", two, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(a.Slides[0].Texts, b.Slides[0].Texts) || !reflect.DeepEqual(a.Slides[0].Shapes, b.Slides[0].Shapes) {
		t.Fatal("alias changed editable rendering")
	}
}

func TestBrowsingCatalogModeComplete(t *testing.T) {
	bundle := "../../library/wm-design-system/v11"
	doc, coverage, e := TemplateBrowsingDocumentWithFrames(bundle, "", 2026, "catalog")
	if e != nil {
		t.Fatal(e)
	}
	source, e := Load(bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	entities, e := collectFrameEntities(source, map[string]string{})
	if e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, entry := range coverage.Entries {
		if entry.Kind == "frame" {
			count++
		}
	}
	if count != len(entities) || count != coverage.FrameCandidates || coverage.FrameMode != "catalog" || len(coverage.FrameAliases) != 0 || len(coverage.FrameExclusions) != 0 {
		t.Fatal("catalog variants omitted", coverage)
	}
	t.Logf("catalog browsing: %d retained templates, %d frame entities, %d total pages", len(coverage.Entries)-count, count, len(doc.Slides))
	if _, _, e = TemplateBrowsingDocumentWithFrames(bundle, "", 2026, "partial"); e == nil {
		t.Fatal("unknown mode accepted")
	}
}
