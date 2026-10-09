package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func teamTestPod(id string, roles ...string) Node {
	items := make([]any, len(roles))
	for i, r := range roles {
		items[i] = r
	}
	return Node{ID: id, Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/pod"}, Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 0, Y: 0, W: 270, H: 216}}, Arguments: map[string]any{"title": id, "band": "inverse", "roles": items}}
}
func teamTestTemplate() LocalTemplate {
	return LocalTemplate{Nodes: []Node{teamTestPod("close", "Lead", "Accountant"), teamTestPod("data", "Architect")}}
}
func TestTeamRolesIdentityAndAssignment(t *testing.T) {
	t.Parallel()
	template := teamTestTemplate()
	ops := []TeamOperation{{Action: "add-role", Component: "close", ID: "qa", Role: &TeamRole{Label: "Quality lead"}}, {Action: "reorder-roles", Component: "close", Order: []string{"qa", "slot-001", "slot-002"}}, {Action: "reassign-role", Component: "close", ID: "qa", Target: "data"}, {Action: "remove-role", Component: "close", ID: "slot-002"}}
	for _, op := range ops {
		if e := applyTeamOperation(&template, op); e != nil {
			t.Fatal(e)
		}
	}
	close, data := template.Nodes[0], template.Nodes[1]
	if got := strings.Join(close.Keys["roles"], ","); got != "slot-001" {
		t.Fatal(got)
	}
	if got := strings.Join(data.Keys["roles"], ","); got != "slot-001,qa" {
		t.Fatal(got)
	}
	if got := data.Arguments["roles"].([]any)[1]; got != "Quality lead" {
		t.Fatal(got)
	}
}
func TestTeamPodLayoutCountAndCapacity(t *testing.T) {
	t.Parallel()
	template := teamTestTemplate()
	layout := TeamPodLayout{Rect: wmdesign.Rect{X: 12, Y: 24, W: 600, H: 240}, Columns: 2, Gap: 24}
	if e := arrangeTeamPods(&template, []string{"data", "close"}, layout); e != nil {
		t.Fatal(e)
	}
	if r := template.Nodes[1].Placement.Rect; r.X != 12 || r.Y != 24 || r.W != 288 || r.H != 90 {
		t.Fatalf("%+v", r)
	}
	if r := template.Nodes[0].Placement.Rect; r.X != 324 || r.H != 132 {
		t.Fatalf("%+v", r)
	}
	layout.Rect.H = 100
	if e := arrangeTeamPods(&template, []string{"close", "data"}, layout); e == nil {
		t.Fatal("capacity silently exceeded")
	}
}
func TestTeamReportingReparentAndCycle(t *testing.T) {
	t.Parallel()
	n := Node{ID: "reports", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/orgchart"}, Arguments: map[string]any{"root": teamObject(TeamReport{Key: "sponsor", Org: "client", Title: "Sponsor", Children: []TeamReport{{Key: "lead", Org: "wm", Title: "Lead", Children: []TeamReport{{Key: "analyst", Org: "wm", Title: "Analyst"}}}, {Key: "liaison", Org: "client", Title: "Liaison", Dotted: true}}})}}
	template := LocalTemplate{Nodes: []Node{n}}
	if e := applyTeamOperation(&template, TeamOperation{Action: "reparent-report", Component: "reports", ID: "analyst", Parent: "liaison"}); e != nil {
		t.Fatal(e)
	}
	root := template.Nodes[0].Arguments["root"].(map[string]any)
	_, parent, _ := findTeamReport(root, "analyst")
	if parent["key"] != "liaison" {
		t.Fatal(parent)
	}
	if e := applyTeamOperation(&template, TeamOperation{Action: "reparent-report", Component: "reports", ID: "liaison", Parent: "analyst"}); e == nil {
		t.Fatal("cycle accepted")
	}
	if e := applyTeamOperation(&template, TeamOperation{Action: "remove-report", Component: "reports", ID: "liaison"}); e == nil {
		t.Fatal("implicit subtree deletion accepted")
	}
	if e := applyTeamOperation(&template, TeamOperation{Action: "update-report", Component: "reports", ID: "liaison", Report: &TeamReport{Key: "liaison", Title: "Client lead", Org: "client"}}); e != nil {
		t.Fatal(e)
	}
	liaison, _, _ := findTeamReport(root, "liaison")
	if liaison["dotted"] != false {
		t.Fatal("could not clear dotted reporting")
	}
}
func TestTeamStockScopedReportingKeys(t *testing.T) {
	t.Parallel()
	n := Node{ID: "org", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/orgchart"}, Arguments: map[string]any{"root": map[string]any{"org": "client", "title": "Sponsor", "children": []any{map[string]any{"org": "wm", "title": "Lead", "children": []any{map[string]any{"org": "wm", "title": "Analyst"}}}}}}}
	if e := normalizeTeamNode(&n); e != nil {
		t.Fatal(e)
	}
	root := n.Arguments["root"].(map[string]any)
	child := root["children"].([]any)[0].(map[string]any)
	grandchild := child["children"].([]any)[0].(map[string]any)
	if child["key"] == grandchild["key"] {
		t.Fatal("scoped stock IDs remained ambiguous")
	}
	before := canonical(n)
	if e := normalizeTeamNode(&n); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(before, canonical(n)) {
		t.Fatal("identities drifted on a second normalization")
	}
}
func TestTeamGovernanceReorderPreservesMembersAndDecisions(t *testing.T) {
	t.Parallel()
	tiers := []any{teamObject(TeamTier{Key: "steering", Name: "Steering", Cadence: "Monthly", Members: [][]string{{"Sponsor", "client"}}, Decisions: []any{"Funding"}}), teamObject(TeamTier{Key: "delivery", Name: "Delivery", Cadence: "Weekly", Members: [][]string{{"Lead", "wm"}}, Decisions: []any{"Scope"}})}
	n := Node{ID: "gov", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/governance"}, Arguments: map[string]any{"tiers": tiers}, Keys: map[string][]string{"tiers": {"steering", "delivery"}, "tiers/0/members": {"sponsor"}, "tiers/0/decisions": {"funding"}, "tiers/1/members": {"lead"}, "tiers/1/decisions": {"scope"}}}
	template := LocalTemplate{Nodes: []Node{n}}
	decision := "Release scope"
	ops := []TeamOperation{{Action: "reorder-tiers", Component: "gov", Order: []string{"delivery", "steering"}}, {Action: "add-member", Component: "gov", Parent: "delivery", ID: "qa", Member: &TeamMember{Label: "Quality lead", Org: "wm"}}, {Action: "update-decision", Component: "gov", Parent: "delivery", ID: "scope", Decision: &decision}}
	for _, op := range ops {
		if e := applyTeamOperation(&template, op); e != nil {
			t.Fatal(e)
		}
	}
	n = template.Nodes[0]
	if got := strings.Join(n.Keys["tiers/0/members"], ","); got != "lead,qa" {
		t.Fatal(got)
	}
	if got := n.Keys["tiers/1/decisions"][0]; got != "funding" {
		t.Fatal(got)
	}
	tier := n.Arguments["tiers"].([]any)[0].(map[string]any)
	if tier["decisions"].([]any)[0] != decision {
		t.Fatal(tier)
	}
}
func teamProjectFixture(t *testing.T) *Project {
	p := example(t)
	draft, e := ScaffoldTemplate(bundle(t), "team/pods", wmdesign.CandidateEngine, "Pod composition example", 2026)
	if e != nil {
		t.Fatal(e)
	}
	p.Document.LocalTemplates = map[string]LocalTemplate{"team": draft.Template}
	p.Document.Slides = []Slide{{ID: "team-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "team"}, Values: draft.SyntheticSourceValues}}
	raw, e := json.Marshal(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	return p
}
func TestTeamCatalogPreviewApplyAndStaleGuard(t *testing.T) {
	t.Parallel()
	p := teamProjectFixture(t)
	inspection, e := InspectTeam(p, "team-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	pods := []TeamComponent{}
	for _, c := range inspection.Components {
		if c.Type == "pod" {
			pods = append(pods, c)
		}
	}
	if len(pods) < 2 {
		t.Fatal("missing catalog pods")
	}
	patch := TeamPatch{Schema: TeamPatchSchema, Actor: "operator", Reason: "Move quality responsibility into the data pod", ExpectedSourceSHA256: p.SourceHash(), Operations: []TeamOperation{{Action: "materialize-component", Component: pods[0].ID}, {Action: "materialize-component", Component: pods[1].ID}}}
	// Removing one role resolves the receiving pod's measured capacity explicitly.
	firstKeys := pods[0].Keys["roles"]
	secondKeys := pods[1].Keys["roles"]
	if len(firstKeys) == 0 || len(secondKeys) == 0 {
		t.Fatal("missing stock role identities")
	}
	newID := "quality"
	patch.Operations = append(patch.Operations, TeamOperation{Action: "add-role", Component: pods[0].ID, ID: newID, Role: &TeamRole{Label: "Quality lead"}}, TeamOperation{Action: "remove-role", Component: pods[1].ID, ID: secondKeys[len(secondKeys)-1]}, TeamOperation{Action: "reassign-role", Component: pods[0].ID, ID: newID, Target: pods[1].ID})
	before := append([]byte(nil), p.Raw...)
	preview, e := PatchTeam(p, "team-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if preview.Applied || preview.BeforeSHA256 == preview.AfterSHA256 {
		t.Fatal("incorrect preview")
	}
	raw, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, raw) {
		t.Fatal("preview changed source")
	}
	applied, e := PatchTeam(p, "team-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if !applied.Applied || applied.Decision == "" || applied.AfterSHA256 != preview.AfterSHA256 {
		t.Fatal("incorrect commit")
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = PatchTeam(p, "team-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale source applied")
	}
	template := p.Document.LocalTemplates["team"]
	list, i := findDiagramNode(&template.Nodes, pods[1].ID)
	if list == nil || teamFind((*list)[i].Keys["roles"], newID) < 0 {
		t.Fatal("role identity lost")
	}
}
func TestTeamPatchStrictFields(t *testing.T) {
	t.Parallel()
	patch := TeamPatch{Schema: TeamPatchSchema, Actor: "operator", Reason: "Strict contract", ExpectedSourceSHA256: strings.Repeat("a", 64), Operations: []TeamOperation{{Action: "add-role", Component: "pod", ID: "qa", Role: &TeamRole{Label: "QA"}}}}
	if _, e := DecodeTeamPatch(canonical(patch), "patch"); e != nil {
		t.Fatal(e)
	}
	for _, raw := range [][]byte{bytes.Replace(canonical(patch), []byte(`"actor":"operator"`), []byte(`"actor":"operator","unsafe":true`), 1), append(canonical(patch), []byte("\n---\n{}")...)} {
		if _, e := DecodeTeamPatch(raw, "patch"); e == nil {
			t.Fatal("invalid patch accepted")
		}
	}
	patch.Operations[0].Parent = "unrelated"
	if _, e := DecodeTeamPatch(canonical(patch), "patch"); e == nil {
		t.Fatal("ignored irrelevant operation field")
	}
	patch.Operations = []TeamOperation{{Action: "update-decision", Component: "gov", Parent: "steering", ID: "funding", Decision: func() *string { s := "Approve funding for phase two"; return &s }()}}
	if _, e := DecodeTeamPatch(canonical(patch), "patch"); e != nil {
		t.Fatal("ordinary decision copy rejected", e)
	}
}

func TestTeamAuthoredIrrelevantEmptyFieldsRefused(t *testing.T) {
	t.Parallel()
	patch := TeamPatch{Schema: TeamPatchSchema, Actor: "operator", Reason: "Strict authored field relevance", ExpectedSourceSHA256: strings.Repeat("a", 64), Operations: []TeamOperation{{Action: "add-role", Component: "pod", ID: "qa", Role: &TeamRole{Label: "QA"}}}}
	for _, field := range []string{`"target":""`, `"rect":null`, `"order":[]`, `"incident_edges":""`} {
		raw := bytes.Replace(canonical(patch), []byte(`"action":"add-role"`), []byte(`"action":"add-role",`+field), 1)
		if _, e := DecodeTeamPatch(raw, "patch"); e == nil {
			t.Fatalf("irrelevant authored empty field accepted: %s", field)
		}
	}
}
func TestTeamComponentRoutingRequiresSharedQualifiedDefinition(t *testing.T) {
	t.Parallel()
	n := teamTestPod("pod", "Lead")
	if teamKind(n) != "pod" {
		t.Fatal("supported shared pod not recognized")
	}
	n.Definition.Scope = "local"
	if teamKind(n) != "" {
		t.Fatal("local definition treated as installed semantic component")
	}
	n.Definition.Scope = "shared"
	n.Definition.ID = "pod"
	if teamKind(n) != "" {
		t.Fatal("unqualified shared ID treated as semantic component")
	}
}

func TestTeamCatalogOverflowIsAtomic(t *testing.T) {
	t.Parallel()
	p := teamProjectFixture(t)
	inspection, e := InspectTeam(p, "team-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	var pod TeamComponent
	for _, c := range inspection.Components {
		if c.Type == "pod" {
			pod = c
			break
		}
	}
	patch := TeamPatch{Schema: TeamPatchSchema, Actor: "operator", Reason: "Reject a role count that does not fit", ExpectedSourceSHA256: p.SourceHash(), Operations: []TeamOperation{{Action: "materialize-component", Component: pod.ID}}}
	for _, id := range []string{"r1", "r2", "r3", "r4", "r5"} {
		patch.Operations = append(patch.Operations, TeamOperation{Action: "add-role", Component: pod.ID, ID: id, Role: &TeamRole{Label: "Additional role"}})
	}
	before := p.SourceHash()
	if _, e = PatchTeam(p, "team-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("overflow was applied")
	}
	reloaded, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if before != reloaded.SourceHash() {
		t.Fatal("overflow changed source")
	}
}

// Optional private specimen generator; no desktop acceptance is inferred here.
func TestTeamCompositionCatalogDemo(t *testing.T) {
	out := os.Getenv("PPTXGENGO_TEAM_COMPOSITION_DEMO_OUT")
	if out == "" {
		t.Skip("set a new private Documents output directory")
	}
	if !filepath.IsAbs(out) {
		t.Fatal("demo output must be absolute")
	}
	if _, e := os.Lstat(out); !os.IsNotExist(e) {
		t.Fatal("demo output must not exist")
	}
	original := teamProjectFixture(t)
	if e := filepath.WalkDir(original.Root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(original.Root, path)
		if e != nil {
			return e
		}
		target := filepath.Join(out, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(target, raw, 0600)
	}); e != nil {
		t.Fatal(e)
	}
	p, e := Load(out)
	if e != nil {
		t.Fatal(e)
	}
	baseline, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	inspection, e := InspectTeam(p, "team-slide", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	var pods []TeamComponent
	for _, c := range inspection.Components {
		if c.Type == "pod" {
			pods = append(pods, c)
		}
	}
	if len(pods) != 3 {
		t.Fatal("catalog count changed")
	}
	patch := TeamPatch{Schema: TeamPatchSchema, Actor: "qualification", Reason: "Adapt the three-pod catalog example to two unequal pods without adding a program manager", ExpectedSourceSHA256: p.SourceHash(), Operations: []TeamOperation{{Action: "materialize-component", Component: pods[0].ID}, {Action: "materialize-component", Component: pods[1].ID}, {Action: "materialize-component", Component: pods[2].ID}, {Action: "remove-component", Component: pods[2].ID}, {Action: "remove-role", Component: pods[1].ID, ID: pods[1].Keys["roles"][2]}, {Action: "arrange-pods", Order: []string{pods[0].ID, pods[1].ID}, Layout: &TeamPodLayout{Rect: wmdesign.Rect{X: 0, Y: 0, W: 846, H: 174}, Columns: 2, Gap: 18}}}}
	if e = os.WriteFile(filepath.Join(out, "team-patch.json"), canonical(patch), 0600); e != nil {
		t.Fatal(e)
	}
	preview, e := PatchTeam(p, "team-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "team-preview.json"), canonical(preview), 0600); e != nil {
		t.Fatal(e)
	}
	applied, e := PatchTeam(p, "team-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	p, e = Load(out)
	if e != nil {
		t.Fatal(e)
	}
	// Copy is authored explicitly after composition; no title meaning is inferred
	// from native geometry or pod counts by the production command.
	values := map[string]any{}
	for k, v := range p.Document.Slides[0].Values {
		values[k] = v
	}
	values["title"] = "Two unequal pods work beside four Northfield roles"
	editorial, e := EditSlidesWithOptions(p, map[string]SlideEdit{"team-slide": {Values: values}}, bundle(t), wmdesign.CandidateEngine, EditOptions{CheckFit: true})
	if e != nil {
		t.Fatal(e)
	}
	p, e = Load(out)
	if e != nil {
		t.Fatal(e)
	}
	built, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	evidence := map[string]any{"schema": "pptxgengo.team-composition-demo.v1", "status": "source_and_go_fit_only_native_review_pending", "baseline": baseline, "patch": applied, "editorial_edit": editorial, "built": built, "scope": "Catalog-derived two unequal pods with retained client role row. Title explicitly reauthored. Native geometry does not imply membership."}
	if e = os.WriteFile(filepath.Join(out, "team-demo.json"), canonical(evidence), 0600); e != nil {
		t.Fatal(e)
	}
}
