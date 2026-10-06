package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func densityTestBundle() string {
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v11", "intake-20261006-649-frozen", "bundle")
}

func TestDensitySplitFitClassification(t *testing.T) {
	zone := Rect{57, 162, 270, 288}
	text := sceneItem{Text: &TextRecord{Rect: Rect{57, 400, 270, 60}}}
	shape := sceneItem{Shape: &sceneShape{Record: ShapeRecord{Rect: Rect{57, 162, 270, 288}}}}
	if !densityTextOnlyBottomOverflow(&scenePlan{Items: []sceneItem{shape, text}}, zone) {
		t.Fatal("bounded text-only bottom overflow did not qualify")
	}
	flow := &scenePlan{ID: "flow", Bounds: Rect{57, 400, 270, 60}, TextFlowBounds: true, Items: []sceneItem{{Shape: &sceneShape{Record: ShapeRecord{ID: "flow.container", Rect: Rect{57, 400, 270, 60}}}}, {Text: &TextRecord{Rect: Rect{75, 418, 234, 21}}}}}
	if !densityTextOnlyBottomOverflow(flow, zone) {
		t.Fatal("auto-height callout bottom padding did not qualify")
	}
	flow.TextFlowBounds = false
	if densityTextOnlyBottomOverflow(flow, zone) {
		t.Fatal("fixed authored container qualified as text flow")
	}
	for _, tc := range []struct {
		name string
		item sceneItem
	}{
		{"shape", sceneItem{Shape: &sceneShape{Record: ShapeRecord{Rect: Rect{57, 400, 270, 60}}}}},
		{"table", sceneItem{Table: &sceneTable{Rect: Rect{57, 400, 270, 60}}}},
		{"chart", sceneItem{Chart: &sceneChart{Rect: Rect{57, 400, 270, 60}}}},
		{"position", sceneItem{Text: &TextRecord{Rect: Rect{56, 400, 270, 60}}}},
		{"width", sceneItem{Text: &TextRecord{Rect: Rect{57, 400, 271, 60}}}},
		{"start", sceneItem{Text: &TextRecord{Rect: Rect{57, 451, 270, 60}}}},
		{"rotation", sceneItem{Text: &TextRecord{Rect: Rect{57, 400, 270, 60}, Rotation: 1}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if densityTextOnlyBottomOverflow(&scenePlan{Items: []sceneItem{text, tc.item}}, zone) {
				t.Fatal("density accepted authored geometry or non-text overflow")
			}
		})
	}
}

// Optional targeted diagnostics retain exact stock input and the fail-closed
// contrast policy, useful while qualifying new native font-pair calibration.
func TestDensityStockFitDiagnostics(t *testing.T) {
	if os.Getenv("WMDS_DENSITY_FIT_DIAGNOSTICS") == "" {
		t.Skip("set WMDS_DENSITY_FIT_DIAGNOSTICS to print stock fit diagnostics")
	}
	doc, err := LibrarySourceReference(densityTestBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{"understanding-situation/split": true, "pillars/two-categories-six": true, "pillars/two-categories-six-magenta": true, "pillars/two-categories-six-stacked-magenta": true, "capability-heat/annotated": true, "risk-heat/annotated": true, "heat-tile-map/insight-cards": true}
	for _, slide := range doc.Slides {
		if !keys[slide.TemplateBinding.Template] {
			continue
		}
		if slide.TemplateBinding.Template == "understanding-situation/split" {
			source := densityTestSource(t)
			frame, e := source.ResolveFrame(slide.Frame)
			if e != nil {
				t.Fatal(e)
			}
			t.Logf("split frame=%+v node=%s", frame, slide.Nodes[1].Scene.Node)
		}
		probe := doc
		probe.Slides = []SlideSpec{slide}
		_, report, err := BuildWithEngine(densityTestBundle(), "", probe, CandidateEngine)
		if err != nil {
			t.Logf("%s preferred=%s error=%v", slide.TemplateBinding.Template, slide.Density, err)
		} else {
			t.Logf("%s %+v", slide.TemplateBinding.Template, report.Slides[0].Density)
		}
	}
}
func densityTestSource(t *testing.T) *Source {
	t.Helper()
	s, e := Load(densityTestBundle(), "")
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func densityText(t *testing.T, sr SlideReport, id string) TextRecord {
	t.Helper()
	for _, tr := range sr.Texts {
		if tr.ID == id {
			return tr
		}
	}
	t.Fatalf("missing text %s", id)
	return TextRecord{}
}
func TestTypographyDensityRolesAndImmutableSource(t *testing.T) {
	s := densityTestSource(t)
	before, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		token, scope string
		sizes, leads [3]float64
	}{
		{"body", "body", [3]float64{14, 12, 11}, [3]float64{21, 17, 15}},
		{"small", "body", [3]float64{12, 11, 10}, [3]float64{18, 15, 13}},
		{"small", "cell", [3]float64{12, 10, 8}, [3]float64{18, 13, 11}},
		{"body", "cell", [3]float64{14, 12, 10}, [3]float64{21, 17, 13}},
		{"subhead", "body", [3]float64{18, 16, 14}, [3]float64{24, 21, 18}},
		{"title", "header", [3]float64{32, 28, 24}, [3]float64{36, 32, 28}},
	} {
		for i, level := range s.Tokens.Density.Levels {
			st, e := s.StyleForDensity(tc.token, level, tc.scope)
			if e != nil || st.Size != tc.sizes[i] || st.Leading != tc.leads[i] {
				t.Fatalf("%s/%s/%s: %v %v", level, tc.scope, tc.token, st, e)
			}
			base, _ := s.Style(tc.token)
			if st.Family != base.Family || st.Weight != base.Weight || st.Case != base.Case {
				t.Fatal("density changed non-typographic style identity")
			}
		}
	}
	for _, token := range []string{"source", "footer", "display"} {
		base, _ := s.Style(token)
		st, e := s.StyleForDensity(token, "dense", "body")
		if e != nil || !reflect.DeepEqual(base, st) {
			t.Fatalf("fixed role %s changed", token)
		}
	}
	after, _ := json.Marshal(s)
	if !bytes.Equal(before, after) {
		t.Fatal("density mutated immutable source")
	}
}

func TestDensityDefaultsAndLegacyV10OptInRejection(t *testing.T) {
	current, err := Load(densityTestBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	currentDefault, err := slideDensity(current, SlideSpec{Frame: FrameRequest{}})
	if err != nil {
		t.Fatal(err)
	}
	if currentDefault.Requested != "comfortable" || currentDefault.Header != "comfortable" || !currentDefault.Auto {
		t.Fatalf("V11 default density policy changed: %+v", currentDefault)
	}
	noAuto := false
	currentOptOut, err := slideDensity(current, SlideSpec{Frame: FrameRequest{}, AutoDensity: &noAuto})
	if err != nil || currentOptOut.Auto {
		t.Fatalf("V11 explicit auto_density:false was ignored: %+v, %v", currentOptOut, err)
	}

	legacy, err := Load(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	legacyDefault, err := slideDensity(legacy, SlideSpec{Frame: FrameRequest{}})
	if err != nil {
		t.Fatal(err)
	}
	if legacyDefault.Requested != "comfortable" || legacyDefault.Header != "comfortable" || legacyDefault.Auto || legacyDefault.Adjusted {
		t.Fatalf("V10 default rendering must retain the comfortable legacy path: %+v", legacyDefault)
	}
	for _, slide := range []SlideSpec{
		{Frame: FrameRequest{}, Density: "compact"},
		{Frame: FrameRequest{}, AutoDensity: ptrDensityBool(true)},
	} {
		if _, err := slideDensity(legacy, slide); err == nil || !strings.Contains(err.Error(), "density.tokens_required") {
			t.Fatalf("legacy source accepted a density feature: %+v, %v", slide, err)
		}
	}
}

func ptrDensityBool(value bool) *bool { return &value }

func TestDensityNativeTableAndNestedRichRoles(t *testing.T) {
	doc := densityNativeDemoDocument(t, false)
	deck, report, err := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for i, sr := range report.Slides {
		bodySize := []float64{14, 12, 11}[i]
		cellSmall := []float64{12, 10, 8}[i]
		if densityText(t, sr, "title").Layout.Style.Size != 32 {
			t.Fatal("body level changed comfortable header")
		}
		if densityText(t, sr, "node01.title").Layout.Style.Size != []float64{18, 16, 14}[i] {
			t.Fatal("nested card title did not use slide tier")
		}
		richFound, smallFound, bulletsFound := false, false, false
		for _, tr := range sr.Texts {
			if strings.HasPrefix(tr.ID, "node03.") && tr.Rich != nil {
				richFound = true
				for _, para := range tr.Rich.Paragraphs {
					for _, run := range para.Runs {
						if run.Style.Size != bodySize {
							t.Fatalf("nested body run escaped tier: %+v", run.Style)
						}
					}
				}
			}
		}
		for _, table := range sr.Tables {
			for _, cell := range table.Cells {
				if cell.Row == 0 {
					continue
				}
				tr := cell.Text
				if tr.Layout.Style.Size != cellSmall {
					t.Fatalf("cell size %+v; expected %v", tr.Layout.Style, cellSmall)
				}
				if tr.Rich == nil {
					continue
				}
				for _, para := range tr.Rich.Paragraphs {
					if para.Bullet {
						bulletsFound = true
						if para.BulletIndentPt != []float64{12, 11, 10}[i] || para.BulletMarkerPt != []float64{3, 3, 2.5}[i] {
							t.Fatal("native cell marker/indent escaped tier")
						}
					}
					for _, run := range para.Runs {
						if run.Style.Size != cellSmall {
							t.Fatalf("nested table run escaped tier: %+v", run.Style)
						}
						smallFound = true
					}
				}
			}
		}
		if !richFound || !smallFound || !bulletsFound {
			t.Fatal("nested density control missing")
		}
	}
	xml := draftReviewZipPart(t, deck, "ppt/slides/slide3.xml")
	for _, want := range []string{`<a:tbl>`, `sz="800"`, `<a:buSzPts val="250"/>`, `marL="127000" indent="-127000"`, `<a:spcPts val="1100"/>`} {
		if !strings.Contains(xml, want) {
			t.Fatalf("native Dense cell XML lacks %s", want)
		}
	}
	if !strings.Contains(xml, `h="1219200"`) {
		t.Fatal("density changed fixed 96pt table row height")
	}
}

func TestDensityReviewedRoleCorrectionsAndExceptionBounds(t *testing.T) {
	s := densityTestSource(t)
	if !densityRoleCorrections(s) {
		t.Skip("latest role-token source has not been frozen yet")
	}
	if len(s.Tokens.Density.Exceptions) != 2 {
		t.Fatal("explicit literal exception metadata missing")
	}
	for i, level := range s.Tokens.Density.Levels {
		st, err := s.StyleForDensity("number-long", level, "body")
		if err != nil || st.Size != []float64{13, 12, 11}[i] || st.Leading != []float64{17, 16, 14}[i] {
			t.Fatalf("long numeral role %+v %v", st, err)
		}
		r := renderer{source: s, bodyDensity: level}
		badge, err := r.intakeVennDensityBadgeStyle(Style{Family: "IBM Plex Mono"})
		if err != nil || math.Abs(badge.Size-[]float64{9.5, 8 * 9.5 / 9, 8 * 9.5 / 9}[i]) > .0001 || badge.Leading != badge.Size {
			t.Fatalf("point badge role %+v %v", badge, err)
		}
	}
	copyDensity := *s.Tokens.Density
	copyDensity.Exceptions = []DensityException{{Element: "Everything", Node: "body", Size: [2]float64{6.5, 9}, Font: "mono", Reason: "Not a declared exception"}}
	if err := validateTypographyDensity(&copyDensity); err == nil {
		t.Fatal("general below-floor exception accepted")
	}
	copyDensity.Exceptions = append([]DensityException(nil), s.Tokens.Density.Exceptions...)
	copyDensity.Exceptions[0].Size = [2]float64{7, 9}
	if err := validateTypographyDensity(&copyDensity); err == nil {
		t.Fatal("declared exception floor drift accepted")
	}
}
func densityProbeDoc() Document {
	return Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "density", Frame: FrameRequest{Footer: "compact", TitleLines: 2}, Eyebrow: "Density", Title: "A full-size action title\nstays independent of the body", Nodes: []Node{{ID: "copy", Kind: "text", Style: "body", Text: strings.Repeat("Teams review priorities and deliver useful services. ", 10), Rect: Rect{57, 162, 270, 230}}}}}}
}
func TestAutoDensityWholeSlideRetriesWithoutInputMutation(t *testing.T) {
	s := densityTestSource(t)
	typography, e := NewTypographyEngine(filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	doc := densityProbeDoc()
	copyText := doc.Slides[0].Nodes[0].Text
	needs := make([]float64, 3)
	for i, level := range s.Tokens.Density.Levels {
		st, _ := s.StyleForDensity("body", level, "body")
		l, e := typography.Measure(copyText, st, 270)
		if e != nil {
			t.Fatal(e)
		}
		needs[i] = math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
	}
	if !(needs[0] > needs[1] && needs[1] > needs[2]) {
		t.Fatalf("invalid test heights %v", needs)
	}
	doc.Slides[0].Nodes[0].Rect.H = (needs[1] + needs[2]) / 2
	before, _ := json.Marshal(doc)
	_, report, e := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	dr := report.Slides[0].Density
	if dr == nil || dr.Requested != "comfortable" || dr.Resolved != "dense" || dr.Header != "comfortable" || !dr.Auto || len(dr.Steps) != 2 || len(report.DensityAdjustments) != 1 {
		t.Fatalf("bad automatic report %+v", dr)
	}
	if densityText(t, report.Slides[0], "copy").Layout.Style.Size != 11 || densityText(t, report.Slides[0], "title").Layout.Style.Size != 32 {
		t.Fatal("body/header density leaked")
	}
	after, _ := json.Marshal(doc)
	if !bytes.Equal(before, after) {
		t.Fatal("automatic density persisted in caller input")
	}
	doc.Slides[0].Nodes[0].Text = "Less content returns to the preferred density."
	_, report, e = BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if report.Slides[0].Density.Resolved != "comfortable" || len(report.DensityAdjustments) != 0 {
		t.Fatal("automatic tier was persisted")
	}
	disabled := false
	doc.Slides[0].AutoDensity = &disabled
	doc.Slides[0].Nodes[0].Text = copyText
	if _, _, e = BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine); e == nil || !strings.Contains(e.Error(), "overflow") {
		t.Fatalf("opt-out did not retain preferred fit error: %v", e)
	}
}
func TestTypographyDensityTwoLineHeaderFits(t *testing.T) {
	for _, level := range []string{"comfortable", "compact", "dense"} {
		t.Run(level, func(t *testing.T) {
			doc := densityProbeDoc()
			doc.Slides[0].Nodes = nil
			doc.Slides[0].Frame.HeaderDensity = level
			_, report, e := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			tr := densityText(t, report.Slides[0], "title")
			if len(tr.Layout.Lines) != 2 {
				t.Fatalf("two-line title lost: %d", len(tr.Layout.Lines))
			}
			if report.Slides[0].Frame.TitleRule != 144 || report.Slides[0].Frame.Body.Y != 162 {
				t.Fatal("header density changed geometry")
			}
		})
	}
}
func TestAutoDensityDoesNotRetryHeadersValidationOrGeometry(t *testing.T) {
	for name, mutate := range map[string]func(*SlideSpec){
		"header":       func(s *SlideSpec) { s.Title = strings.Repeat("A title that cannot fit in two lines ", 20) },
		"invalid-body": func(s *SlideSpec) { s.Nodes[0].Style = "unknown" },
		"geometry":     func(s *SlideSpec) { s.Nodes[0].Rect.X = -100 },
	} {
		t.Run(name, func(t *testing.T) {
			doc := densityProbeDoc()
			mutate(&doc.Slides[0])
			_, _, e := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
			if e == nil || isDensityFitError(e) || strings.Contains(e.Error(), "dense_body_overflow") {
				t.Fatalf("non-body fit error retried: %v", e)
			}
		})
	}
}
func TestOldBundleDensityGateAndUnchangedDefault(t *testing.T) {
	s, e := Load(v10IntakeBundle(), "")
	if e != nil {
		t.Fatal(e)
	}
	doc := densityProbeDoc()
	doc.Slides[0].Nodes = nil
	_, report, e := BuildWithEngine(v10IntakeBundle(), "", doc, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if report.Slides[0].Density != nil || len(report.DensityAdjustments) != 0 {
		t.Fatal("old default opted into density")
	}
	if _, e = s.StyleForDensity("body", "compact", "body"); e == nil {
		t.Fatal("old bundle accepted unavailable density")
	}
	doc.Slides[0].Density = "compact"
	if _, _, e = BuildWithEngine(v10IntakeBundle(), "", doc, CandidateEngine); e == nil {
		t.Fatal("old bundle silently changed type")
	}
}

func TestAutoDensityBodyWrapUncertaintyStepsWholeSlide(t *testing.T) {
	s := densityTestSource(t)
	typography, e := NewTypographyEngine(filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	st, _ := s.StyleForDensity("body", "comfortable", "body")
	layout, e := typography.Measure("AV office", st, 8191)
	if e != nil {
		t.Fatal(e)
	}
	doc := densityProbeDoc()
	doc.Slides[0].Nodes = []Node{{ID: "boundary", Kind: "richtext", Style: "body", Rect: Rect{57, 162, layout.Lines[0].Advance, 100}, RichText: &RichTextSpec{Paragraphs: []RichParagraphSpec{{Key: "copy", Runs: []RichRunSpec{{Key: "text", Text: "AV office"}}}}}}}
	_, report, e := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	dr := report.Slides[0].Density
	if dr.Resolved != "compact" || len(dr.Steps) != 1 || !strings.Contains(dr.Steps[0].Reason, "rich.uncertain_wrap_boundary") {
		t.Fatalf("body uncertainty not resolved atomically: %+v", dr)
	}
	if densityText(t, report.Slides[0], "title").Layout.Style.Size != 32 {
		t.Fatal("body uncertainty changed header")
	}
	// The same quality boundary in a rich header stays a header failure.
	headerErr := fmt.Errorf("rich.uncertain_wrap_boundary: title")
	if isDensityFitError(bodyDensityFitFailure(headerErr, false)) || !isDensityFitError(bodyDensityFitFailure(headerErr, true)) {
		t.Fatal("fit quality boundary leaked across header/body phase")
	}
}
