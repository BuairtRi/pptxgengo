package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func teamSemanticFixture(t *testing.T) (*Project, *TextBaseline) {
	t.Helper()
	p := teamProjectFixture(t)
	_, local, e := diagramSlide(p, "team-slide")
	if e != nil {
		t.Fatal(e)
	}
	pod := func(id, title string, x float64, roles []string, keys []string) Node {
		return Node{ID: id, Kind: "component", Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: x, Y: 12, W: 240, H: 250}}, Definition: &Reference{Scope: "shared", ID: "wmds/component/pod"}, Arguments: map[string]any{"title": title, "band": "strong", "roles": roles}, Keys: map[string][]string{"roles": keys}}
	}
	local.Nodes = []Node{pod("delivery", "Delivery pod", 20, []string{"Architect", "Analyst"}, []string{"architect", "analyst"}), pod("operations", "Operations pod", 350, []string{"Engineer"}, []string{"engineer"})}
	if _, e = CompositionCandidate(p, "team-slide", "team-fixture", "test", "Explicit catalog pod derivative", local, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	values := map[string]any{}
	for key, value := range p.Document.Slides[0].Values {
		values[key] = value
	}
	values["title"] = "Two pods coordinate three delivery roles"
	if _, e = EditSlidesWithOptions(p, map[string]SlideEdit{"team-slide": {Values: values}}, bundle(t), wmdesign.CandidateEngine, EditOptions{CheckFit: true}); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	return p, b
}
func teamSemanticPacket(t *testing.T, p *Project, b *TextBaseline, dx, dy float64, both bool) *TextReviewPacket {
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	changes := map[string]NativeGeometry{}
	for _, suffix := range []string{"surface", "text"} {
		if suffix == "text" && !both {
			continue
		}
		name := "delivery.roles.architect." + suffix
		g := objects[name].geometry
		g.X += dx
		g.Y += dy
		changes[name] = g
	}
	return ganttSemanticPacket(t, p, b, changes)
}
func TestTeamSemanticReassignmentPreviewApplyAndRetainedEvidence(t *testing.T) {
	p, b := teamSemanticFixture(t)
	packet := teamSemanticPacket(t, p, b, 330, 42, true)
	r, e := ProposeTeamSemantics(p, packet, "team-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Proposals) != 1 || r.Proposals[0].Status != "proposed" || r.Proposals[0].ToPod != "operations" {
		t.Fatal("role not proposed", r.Proposals)
	}
	q := r.Proposals[0]
	d := GanttSemanticDecisions{Schema: TeamSemanticDecisionsSchema, ReportSHA256: TeamSemanticReportHash(r), Actor: "test operator", Reason: "Confirm membership rather than spacing", Decisions: []GanttSemanticDecision{{ProposalID: q.ID, Action: "reassign", Reason: "Architect joins operations"}}}
	raw := canonical(d)
	before := p.SourceHash()
	preview, e := AdoptTeamSemantics(p, packet, "team-slide", raw, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if preview.Applied || p.SourceHash() != before {
		t.Fatal("preview mutated source")
	}
	applied, e := AdoptTeamSemantics(p, packet, "team-slide", raw, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if !applied.Applied {
		t.Fatal("not applied")
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectTeam(p, "team-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	roles := map[string][]string{}
	for _, c := range inspect.Components {
		roles[c.ID] = c.Keys["roles"]
	}
	if !reflect.DeepEqual(roles["delivery"], []string{"analyst"}) || !reflect.DeepEqual(roles["operations"], []string{"engineer", "architect"}) {
		t.Fatal("identity/membership lost", roles)
	}
	retained, e := os.ReadFile(filepath.Join(p.Root, "assets/objects/sha256", digest(packet.Edited)))
	if e != nil || !bytes.Equal(retained, packet.Edited) {
		t.Fatal("native evidence lost", e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptTeamSemantics(p, packet, "team-slide", raw, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale baseline adopted")
	}
}
func TestTeamSemanticPartialMovementAndAmbiguityRemainManual(t *testing.T) {
	for _, test := range []struct {
		name   string
		dx, dy float64
		both   bool
	}{{"only surface", 330, 42, false}, {"outside target", 290, 42, true}, {"only spacing", 10, 0, true}} {
		t.Run(test.name, func(t *testing.T) {
			p, b := teamSemanticFixture(t)
			packet := teamSemanticPacket(t, p, b, test.dx, test.dy, test.both)
			r, e := ProposeTeamSemantics(p, packet, "team-slide", bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if len(r.Proposals) != 1 || r.Proposals[0].Status != "manual_review" {
				t.Fatal("ambiguous membership inferred", r.Proposals)
			}
		})
	}
}
func TestFamilySemanticStrictDecisions(t *testing.T) {
	valid := GanttSemanticDecisions{Schema: TeamSemanticDecisionsSchema, ReportSHA256: strings.Repeat("a", 64), Actor: "test", Reason: "review", Decisions: []GanttSemanticDecision{{ProposalID: strings.Repeat("b", 64), Action: "reassign", Reason: "confirmed"}}}
	if _, e := DecodeTeamSemanticDecisions(canonical(valid), "file"); e != nil {
		t.Fatal(e)
	}
	for _, action := range []string{"retime", "guess", ""} {
		bad := valid
		bad.Decisions = append([]GanttSemanticDecision(nil), valid.Decisions...)
		bad.Decisions[0].Action = action
		if _, e := DecodeTeamSemanticDecisions(canonical(bad), "file"); e == nil {
			t.Fatal("invalid action", action)
		}
	}
	raw := strings.Replace(string(canonical(valid)), `"schema":`, `"unknown":true,"schema":`, 1)
	if _, e := DecodeTeamSemanticDecisions([]byte(raw), "file"); e == nil {
		t.Fatal("unknown field accepted")
	}
}
func assessmentSemanticFixture(t *testing.T) (*Project, *TextBaseline, string) {
	t.Helper()
	p, id := assessmentFixture(t)
	m := assessmentModel()
	patch := assessmentPatch(p, id, AssessmentOperation{Action: "initialize", Entity: "source", Model: &m, LegendNode: "key", Cascade: true})
	if _, e := PatchAssessment(p, "assessment-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	return p, b, id
}
func assessmentSemanticEdited(t *testing.T, b *TextBaseline, cellChanges map[[2]int]string) []byte {
	t.Helper()
	pkg, e := openLineagePackage(b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	raw, e := pkg.read("ppt/slides/slide1.xml")
	if e != nil {
		t.Fatal(e)
	}
	spans, e := lineageSpans(raw)
	if e != nil {
		t.Fatal(e)
	}
	patches := []lineagePatch{}
	var walk func(*lineageSpan)
	walk = func(span *lineageSpan) {
		if span.node.Name.Space == drawingML && span.node.Name.Local == "tbl" {
			row := 0
			for _, tr := range span.children {
				if tr.node.Name.Space != drawingML || tr.node.Name.Local != "tr" {
					continue
				}
				col := 0
				for _, tc := range tr.children {
					if tc.node.Name.Space != drawingML || tc.node.Name.Local != "tc" {
						continue
					}
					if value, ok := cellChanges[[2]int{row, col}]; ok {
						for _, body := range tc.children {
							if body.node.Name.Space != drawingML || body.node.Name.Local != "txBody" {
								continue
							}
							part := string(raw[body.start:body.end])
							match := regexp.MustCompile(`<a:t(?:\s[^>]*)?>`).FindStringIndex(part)
							start, end := -1, strings.Index(part, "</a:t>")
							if match != nil {
								start = match[1]
							}
							if start < 0 || end < start {
								t.Fatal("cell has no plain text", row, col, part)
							}
							patches = append(patches, lineagePatch{body.start, body.end, part[:start] + value + part[end:]})
						}
					}
					col++
				}
				row++
			}
			return
		}
		for _, c := range span.children {
			walk(c)
		}
	}
	walk(spans)
	raw, e = lineageApply(raw, patches)
	if e != nil {
		t.Fatal(e)
	}
	out, e := lineageRewrite(pkg, map[string][]byte{"ppt/slides/slide1.xml": raw})
	if e != nil {
		t.Fatal(e)
	}

	return out
}
func TestAssessmentSemanticScoresZeroMissingAndInvalid(t *testing.T) {
	p, b, id := assessmentSemanticFixture(t)
	edited := assessmentSemanticEdited(t, b, map[[2]int]string{{1, 1}: "2", {2, 1}: "0", {2, 2}: ""})
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	report, e := ProposeAssessmentSemantics(p, packet, "assessment-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(report.Proposals) != 3 {
		t.Fatal(report.Proposals)
	}
	d := GanttSemanticDecisions{Schema: AssessmentSemanticDecisionsSchema, ReportSHA256: AssessmentSemanticReportHash(report), Actor: "test", Reason: "Confirm reviewed ordinal observations"}
	for _, q := range report.Proposals {
		if q.Status != "proposed" {
			t.Fatal(q)
		}
		d.Decisions = append(d.Decisions, GanttSemanticDecision{ProposalID: q.ID, Action: "set_score", Reason: "Confirmed score/blank"})
	}
	if _, e = AdoptAssessmentSemantics(p, packet, "assessment-slide", id, canonical(d), bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
	if _, e = AdoptAssessmentSemantics(p, packet, "assessment-slide", id, canonical(d), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectAssessment(p, "assessment-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if *inspect.Model.Rows[0].Scores["north"] != 2 || *inspect.Model.Rows[1].Scores["north"] != 0 || inspect.Model.Rows[1].Scores["south"] != nil {
		t.Fatal("scores lost", inspect.Model.Rows)
	}
	retained, e := os.ReadFile(filepath.Join(p.Root, "assets/objects/sha256", digest(edited)))
	if e != nil || !bytes.Equal(retained, edited) {
		t.Fatal("evidence missing")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}

func TestAssessmentSemanticRejectsNonOrdinalAndIdentityEdits(t *testing.T) {
	for _, value := range []string{"5", "-1", "1.5", " 1", "01", "unknown"} {
		t.Run(value, func(t *testing.T) {
			p, b, id := assessmentSemanticFixture(t)
			edited := assessmentSemanticEdited(t, b, map[[2]int]string{{1, 1}: value})
			packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			report, e := ProposeAssessmentSemantics(p, packet, "assessment-slide", id, bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if len(report.Proposals) != 1 || report.Proposals[0].Status != "manual_review" {
				t.Fatal(report.Proposals)
			}
			d := GanttSemanticDecisions{Schema: AssessmentSemanticDecisionsSchema, ReportSHA256: AssessmentSemanticReportHash(report), Actor: "test", Reason: "Check refusal", Decisions: []GanttSemanticDecision{{ProposalID: report.Proposals[0].ID, Action: "set_score", Reason: "Must not silently parse invalid ordinal"}}}
			if _, e = AdoptAssessmentSemantics(p, packet, "assessment-slide", id, canonical(d), bundle(t), wmdesign.CandidateEngine, true); e == nil {
				t.Fatal("invalid score adopted")
			}
		})
	}
	for _, address := range [][2]int{{0, 1}, {1, 0}} {
		t.Run("identity", func(t *testing.T) {
			p, b, id := assessmentSemanticFixture(t)
			edited := assessmentSemanticEdited(t, b, map[[2]int]string{address: "Changed identity"})
			packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = ProposeAssessmentSemantics(p, packet, "assessment-slide", id, bundle(t), wmdesign.CandidateEngine); e == nil {
				t.Fatal("changed axis label interpreted as stable source identity")
			}
		})
	}
}
func TestAssessmentSemanticRejectsNativeScoreFormattingChange(t *testing.T) {
	p, b, id := assessmentSemanticFixture(t)
	edited := assessmentSemanticEdited(t, b, map[[2]int]string{{1, 1}: "2"})
	edited = lineageEdit(t, edited, "ppt/slides/slide1.xml", func(raw []byte) []byte {
		a, z := bytes.Index(raw, []byte("<a:tbl>")), bytes.Index(raw, []byte("</a:tbl>"))
		if a < 0 || z < 0 {
			t.Fatal("missing native table")
		}
		table := string(raw[a:z])
		match := regexp.MustCompile(`sz="[0-9]+"`).FindStringIndex(table)
		if match == nil {
			t.Fatal("table typography absent")
		}
		table = table[:match[0]] + `sz="1701"` + table[match[1]:]
		return append(append(append([]byte{}, raw[:a]...), []byte(table)...), raw[z:]...)
	})
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ProposeAssessmentSemantics(p, packet, "assessment-slide", id, bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("score with native paragraph/run formatting change accepted")
	}
}

func TestTeamSemanticOfficeEquivalentGroupEnvelope(t *testing.T) {
	p, b := teamSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	changes := map[string]NativeGeometry{}
	for _, suffix := range []string{"surface", "text"} {
		name := "delivery.roles.architect." + suffix
		g := objects[name].geometry
		g.X += 330
		g.Y += 42
		changes[name] = g
	}
	g := objects["delivery"].geometry
	child := *g.Child
	g.Child = &child
	g.W = 561
	g.Child.W = 561
	changes["delivery"] = g
	packet := ganttSemanticPacket(t, p, b, changes)
	report, e := ProposeTeamSemantics(p, packet, "team-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil || len(report.Proposals) != 1 || report.Proposals[0].Status != "proposed" || report.Proposals[0].ToPod != "operations" {
		t.Fatal("equivalent Office envelope rejected", report.Proposals, e)
	}
	if len(report.UnresolvedGeometryIDs) != 3 {
		t.Fatal("group envelope must remain retained unresolved evidence", report.UnresolvedGeometryIDs)
	}
	g.Child.W = 560
	changes["delivery"] = g
	packet = ganttSemanticPacket(t, p, b, changes)
	report, e = ProposeTeamSemantics(p, packet, "team-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil || report.Proposals[0].Status != "manual_review" {
		t.Fatal("real scaling accepted as envelope normalization", report.Proposals, e)
	}
}

func TestTeamSemanticParentFrameNormalizationGuards(t *testing.T) {
	base := NativeGeometry{Kind: "grpSp", X: 77, Y: 138, W: 240, H: 132, Child: &wmdesign.Rect{X: 77, Y: 138, W: 240, H: 132}}
	normalized := base
	child := *base.Child
	normalized.Child = &child
	normalized.W = 561
	normalized.Child.W = 561
	field := GeometryReconciliationField{Status: "native_only", Baseline: &base, EditedNative: &normalized}
	if !semanticEquivalentParentFrame(field) {
		t.Fatal("identity envelope rejected")
	}
	for _, mutate := range []func(*NativeGeometry){func(g *NativeGeometry) { g.X++ }, func(g *NativeGeometry) { g.W++ }, func(g *NativeGeometry) { g.Rotation = 1 }, func(g *NativeGeometry) { g.FlipH = true }, func(g *NativeGeometry) { g.Parent = "different" }, func(g *NativeGeometry) { g.Child = nil }} {
		changed := normalized
		copyChild := *normalized.Child
		changed.Child = &copyChild
		mutate(&changed)
		field.EditedNative = &changed
		if semanticEquivalentParentFrame(field) {
			t.Fatal("changed group coordinate frame admitted", changed)
		}
	}
}
