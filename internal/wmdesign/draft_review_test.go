package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDraftReviewValidationAndBounds(t *testing.T) {
	for _, status := range []string{"notstarted", "wip", "complete", "qa"} {
		note := &DraftReviewNote{Status: status, Placement: "pasteboard", Owner: "Morgan", Due: "Friday", Updated: "2026-10-05", Notes: "Ready for review\nSecond line"}
		if err := ValidateDraftReviewNote(note); err != nil {
			t.Errorf("valid %q note rejected: %v", status, err)
		}
	}
	for _, note := range []*DraftReviewNote{
		{}, {Status: "Notstarted"}, {Status: "blocked"}, {Status: "wip", Placement: "offslide"},
		{Status: "qa", Owner: strings.Repeat("x", 257)},
		{Status: "wip", StatusColor: "#FF0000"},
		{Status: "wip", StatusText: "Needs\nreview"},
		{Status: "wip", StatusText: "Contains\x7fcontrol"},
		{Status: "wip", StatusText: strings.Repeat("x", 129)},
		{Status: "complete", Notes: "line\x00break"},
	} {
		if err := ValidateDraftReviewNote(note); err == nil {
			t.Errorf("invalid note accepted: %+v", note)
		}
	}
	if err := ValidateDraftReviewNote(nil); err != nil {
		t.Fatalf("nil note should be absent: %v", err)
	}
}

func TestDraftReviewStatusPaletteAndEditableGeometry(t *testing.T) {
	r := intakeTestRenderer(t)
	want := map[string]string{"notstarted": "F52C00", "wip": "FFC700", "complete": "1DD566", "qa": "0047FF"}
	wantText := map[string]string{"notstarted": "Not started", "wip": "In progress", "complete": "Complete", "qa": "QA'd"}
	for status, color := range want {
		plan, err := r.planDraftReview("palette", DraftReviewNote{Status: status})
		if err != nil {
			t.Fatalf("%s plan: %v", status, err)
		}
		var tab *ShapeRecord
		for i := range plan.Items {
			if plan.Items[i].Shape != nil && plan.Items[i].Shape.Record.ID == "wm-review/palette.tab" {
				tab = &plan.Items[i].Shape.Record
			}
		}
		if tab == nil || tab.Color != color || tab.Rect != (Rect{942, 0, 18, 144}) {
			t.Fatalf("%s tab palette/geometry mismatch: %+v", status, tab)
		}
		if got := draftReviewText(plan, ".status"); got != wantText[status] {
			t.Errorf("%s visible default status label = %q, want %q", status, got, wantText[status])
		}
		if plan.Bounds != (Rect{942, 0, 234, 144}) || len(plan.Groups) != 1 || plan.Groups[0].Contract != DraftReviewContract {
			t.Fatalf("%s card lost native editable group bounds/contract: %+v", status, plan)
		}
	}
}

func TestDraftReviewStatusTextAndColorOverridesAreIndependent(t *testing.T) {
	r := intakeTestRenderer(t)
	note := DraftReviewNote{Status: "wip", StatusText: "Ready <&> reviewed", StatusColor: "kpi.on"}
	plan, err := r.planDraftReview("override", note)
	if err != nil {
		t.Fatal(err)
	}
	if got := draftReviewText(plan, ".status"); got != note.StatusText {
		t.Fatalf("status_text was not rendered literally: %q", got)
	}
	var tab *ShapeRecord
	for i := range plan.Items {
		if plan.Items[i].Shape != nil && plan.Items[i].Shape.Record.ID == "wm-review/override.tab" {
			tab = &plan.Items[i].Shape.Record
		}
	}
	if tab == nil || tab.Color != "1DD566" {
		t.Fatalf("status_color did not override visual fill independent of wip status: %+v", tab)
	}
	grounded, err := r.sceneColor("light", "brand.grounded")
	if err != nil {
		t.Fatal(err)
	}
	statusTextColor := draftReviewTextColor(plan, ".status")
	if statusTextColor != grounded {
		t.Fatalf("status label did not choose contrast ink for the overridden fill: got %s want %s", statusTextColor, grounded)
	}
	tooWide := DraftReviewNote{Status: "qa", StatusText: strings.Repeat("W", 128)}
	if _, err := r.planDraftReview("overflow", tooWide); err == nil || !strings.Contains(err.Error(), "draft_review.text_overflow") {
		t.Fatalf("expected visual status_text overflow error, got %v", err)
	}
}

func draftReviewText(plan *scenePlan, suffix string) string {
	for _, item := range plan.Items {
		if item.Text != nil && strings.HasSuffix(item.Text.ID, suffix) {
			return item.Text.Layout.Original
		}
	}
	return ""
}

func draftReviewTextColor(plan *scenePlan, suffix string) string {
	for _, item := range plan.Items {
		if item.Text != nil && strings.HasSuffix(item.Text.ID, suffix) {
			return item.Text.Color
		}
	}
	return ""
}

func draftReviewDocument(note *DraftReviewNote) Document {
	return Document{
		Schema: "pptxgengo.wmds-foundation.v1", Year: 2026,
		BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "draft-review-test"},
		Slides:        []SlideSpec{{ID: "review-target", Title: "A reviewable slide", Frame: FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1}, DraftReview: note}},
	}
}

func draftReviewBuild(t *testing.T, doc Document) ([]byte, Report) {
	t.Helper()
	deck, report, err := BuildWithEngine(v7IntakeBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	return deck, report
}

func TestDraftReviewRenderingIsPrivateAndDeterministic(t *testing.T) {
	baseDoc := draftReviewDocument(nil)
	baseDeck, baseReport := draftReviewBuild(t, baseDoc)
	noOverlay, noOverlayReport := draftReviewBuild(t, draftReviewDocument(nil))
	if !bytes.Equal(baseDeck, noOverlay) || !reflect.DeepEqual(baseReport, noOverlayReport) {
		t.Fatal("absent draft review changed deterministic baseline output")
	}

	note := &DraftReviewNote{Status: "wip", StatusText: "Ready <&> reviewed", StatusColor: "kpi.on", Owner: "R&D <team> & partners", Due: "Oct 12", Updated: "2026-10-05", Notes: "Needs <native> review & sign-off"}
	doc := draftReviewDocument(note)
	deck, report := draftReviewBuild(t, doc)
	repeat, repeatReport := draftReviewBuild(t, doc)
	if !bytes.Equal(deck, repeat) || !reflect.DeepEqual(report, repeatReport) {
		t.Fatal("draft review build is not deterministic")
	}
	if len(report.Slides) != 1 || report.Slides[0].DraftReview == nil {
		t.Fatal("draft review missing from the private report record")
	}
	review := report.Slides[0].DraftReview
	if review.Group.ID != "wm-review/review-target" || review.Group.Contract != DraftReviewContract || review.Group.Rect != (Rect{942, 0, 234, 144}) {
		t.Fatalf("unexpected native review group: %+v", review.Group)
	}
	if review.VisibleTab != (Rect{942, 0, 18, 144}) {
		t.Fatalf("unexpected on-canvas tab geometry: %+v", review.VisibleTab)
	}
	wantNote := *note
	wantNote.Placement = "edge"
	if review.Note != wantNote {
		t.Fatalf("report metadata did not retain authored status overrides: %+v", review.Note)
	}
	if len(report.Slides[0].Texts) != len(baseReport.Slides[0].Texts) || !reflect.DeepEqual(report.Slides[0].Nodes, baseReport.Slides[0].Nodes) {
		t.Fatal("draft review copy polluted the slide's main text or node records")
	}
	if !reflect.DeepEqual(report.MeasurementPolicy, baseReport.MeasurementPolicy) {
		t.Fatal("draft review changed report measurement policy")
	}
	for _, text := range report.Slides[0].Texts {
		if strings.Contains(text.ID, "wm-review/") {
			t.Fatalf("private text entered main text report: %s", text.ID)
		}
	}

	slideXML := draftReviewZipPart(t, deck, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "wm-review/review-target") || !strings.Contains(slideXML, DraftReviewContract) {
		t.Fatal("native editable group or private metadata missing from slide XML")
	}
	if !strings.Contains(slideXML, "<p:grpSp>") || !strings.Contains(slideXML, "<p:sp>") {
		t.Fatal("review card was not serialized as an editable native shape group")
	}
	if !strings.Contains(slideXML, "R&amp;D &lt;team&gt; &amp; partners") || !strings.Contains(slideXML, "Needs &lt;native&gt; review &amp; sign-off") || !strings.Contains(slideXML, `&#34;status_text&#34;:&#34;Ready \u003c\u0026\u003e reviewed&#34;`) || !strings.Contains(slideXML, `&#34;status_color&#34;:&#34;kpi.on&#34;`) {
		t.Fatal("draft review metadata was not XML escaped or omitted authored override values")
	}
}

func TestDraftReviewPasteboardAndOverflow(t *testing.T) {
	note := &DraftReviewNote{Status: "complete", Placement: "pasteboard", Owner: "Morgan"}
	deck, report := draftReviewBuild(t, draftReviewDocument(note))
	review := report.Slides[0].DraftReview
	if review == nil || review.Group.Rect != (Rect{960, 0, 234, 144}) || review.VisibleTab != (Rect{}) {
		t.Fatalf("pasteboard placement did not move the card and hide its tab: %+v", review)
	}
	if !strings.Contains(draftReviewZipPart(t, deck, "ppt/slides/slide1.xml"), "wm-review/review-target") {
		t.Fatal("pasteboard card missing from editable slide XML")
	}
	tooLong := &DraftReviewNote{Status: "wip", Notes: strings.Repeat("This note has substantial content that cannot fit in the fixed review card. ", 80)}
	if _, _, err := BuildWithEngine(v7IntakeBundle(), "", draftReviewDocument(tooLong), CandidateEngine); err == nil || !strings.Contains(err.Error(), "draft_review.text_overflow") {
		t.Fatalf("expected actionable off-slide note overflow error, got %v", err)
	}
}

func TestRemoveDraftReviewNotesPreservesBaselineAndNoopBytes(t *testing.T) {
	baseDeck, _ := draftReviewBuild(t, draftReviewDocument(nil))
	unchanged, err := RemoveDraftReviewNotes(baseDeck)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unchanged, baseDeck) {
		t.Fatal("removing draft notes changed a deck with no draft note")
	}

	note := &DraftReviewNote{Status: "wip", Owner: "R&D <team>", Due: "Oct 12", Updated: "2026-10-05", Notes: "Review <copy> & confirm"}
	withReview, _ := draftReviewBuild(t, draftReviewDocument(note))
	clean, err := RemoveDraftReviewNotes(withReview)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(clean, []byte("wm-review/")) || bytes.Contains(clean, []byte(DraftReviewContract)) || bytes.Contains(clean, []byte("Review &lt;copy&gt;")) {
		t.Fatal("private draft review objects or metadata survived cleanup")
	}
	if !bytes.Equal(clean, baseDeck) {
		baseParts, cleanParts := draftReviewZipParts(t, baseDeck), draftReviewZipParts(t, clean)
		if !reflect.DeepEqual(baseParts, cleanParts) {
			for path, base := range baseParts {
				if got, ok := cleanParts[path]; !ok || !bytes.Equal(got, base) {
					t.Errorf("cleanup changed baseline package part %s", path)
				}
			}
			for path := range cleanParts {
				if _, ok := baseParts[path]; !ok {
					t.Errorf("cleanup left or added package part %s", path)
				}
			}
		}
	}
	if again, err := RemoveDraftReviewNotes(clean); err != nil || !bytes.Equal(again, clean) {
		t.Fatalf("cleanup should be idempotent: %v", err)
	}
}

func draftReviewZipPart(t *testing.T, deck []byte, name string) string {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if f.Name != name {
			continue
		}
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		_ = r.Close()
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	t.Fatalf("package part not found: %s", name)
	return ""
}

func draftReviewZipParts(t *testing.T, deck []byte) map[string][]byte {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		t.Fatal(err)
	}
	parts := make(map[string][]byte, len(z.File))
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		_ = r.Close()
		if err != nil {
			t.Fatal(err)
		}
		parts[f.Name] = b
	}
	return parts
}

func TestDraftReviewNoteJSONRoundTrip(t *testing.T) {
	note := DraftReviewNote{Status: "qa", StatusText: "Accepted", StatusColor: "brand.blue", Owner: "Morgan", Due: "Oct 12", Updated: "2026-10-05", Notes: "Verified", Placement: "edge"}
	b, err := json.Marshal(note)
	if err != nil {
		t.Fatal(err)
	}
	var got DraftReviewNote
	if err := json.Unmarshal(b, &got); err != nil || got != note {
		t.Fatalf("draft review JSON round trip mismatch: %+v, %v", got, err)
	}
}
