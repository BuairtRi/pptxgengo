package deckproject

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func ganttSemanticFixture(t *testing.T) (*Project, *TextBaseline, string, string) {
	t.Helper()
	p, node := ganttCompositionFixture(t)
	inspection, e := InspectGantt(p, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	group := inspection.Schedule.Groups[0]
	lane := group.Lanes[0]
	var kind string
	for k := range inspection.Schedule.Kinds {
		kind = k
		break
	}
	lane.Items = []wmdesign.GanttItem{{Key: "delivery-task", Kind: kind, Label: "Deliver", From: 1, To: 3}}
	patch := ganttPatch(p, node, GanttOperation{Action: "remove", Entity: "gate", Key: ganttKey(inspection.Schedule.Gates[1].Key)}, GanttOperation{Action: "set", Entity: "lane", Group: group.Key, Key: lane.Key, LaneValue: &lane, Cascade: true}, GanttOperation{Action: "set", Entity: "gate", Key: "approval", Gate: &GanttGateValue{Key: "approval", Label: "Approve", At: 4}})
	if _, e = PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
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
	return p, b, node, node + ".groups." + group.Key + ".lanes." + lane.Key + ".items.delivery-task.segment-0"
}
func ganttSemanticPacket(t *testing.T, p *Project, b *TextBaseline, changes map[string]NativeGeometry) *TextReviewPacket {
	t.Helper()
	edited := geometryEdited(t, p, b, changes, nil)
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	return packet
}
func TestGanttSemanticReviewedTaskAndGateRoundTrip(t *testing.T) {
	p, b, node, bar := ganttSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	var period *geometryObject
	for name, o := range objects {
		if strings.HasPrefix(name, node+".periods.") && strings.HasSuffix(name, ".label") {
			period = o
			break
		}
	}
	if period == nil || objects[bar] == nil {
		t.Fatalf("fixture names missing: %s", bar)
	}
	width := period.geometry.W
	moved := objects[bar].geometry
	moved.X += width * .25
	moved.W += width * .5
	gate := objects[node+".gates.approval.line"].geometry
	gate.X += width * .75
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: moved, node + ".gates.approval.line": gate})
	report, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	decisions := GanttSemanticDecisions{Schema: GanttSemanticDecisionsSchema, ReportSHA256: GanttSemanticReportHash(report), Actor: "test operator", Reason: "Confirm work plan changes"}
	foundTask, foundGate := false, false
	for _, q := range report.Proposals {
		if q.SourceObject != bar && q.SourceObject != node+".gates.approval.line" {
			continue
		}
		if q.Status != "proposed" {
			t.Fatalf("not proposed: %+v", q)
		}
		if q.Entity == "task" {
			foundTask = true
			if math.Abs(*q.ProposedFrom-1.25) > .000002 || math.Abs(*q.ProposedTo-3.75) > .000002 {
				t.Fatalf("wrong interval: %+v", q)
			}
		} else {
			foundGate = true
			if math.Abs(*q.ProposedAt-4.75) > .000002 {
				t.Fatalf("wrong gate %.9f axis=%f width=%f edit=%+v", *q.ProposedAt, report.AxisLeftPT, report.PeriodWidthPT, gate)
			}
		}
		decisions.Decisions = append(decisions.Decisions, GanttSemanticDecision{q.ID, "retime", "Confirmed intended period change"})
	}
	if !foundTask || !foundGate {
		t.Fatalf("missing proposals: %+v", report)
	}
	before := append([]byte(nil), p.Raw...)
	preview, e := AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(decisions), bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if preview.Applied || preview.AfterSHA256 == preview.BeforeSHA256 {
		t.Fatal("invalid preview")
	}
	if disk, _ := os.ReadFile(p.SourcePath); !bytes.Equal(disk, before) {
		t.Fatal("preview wrote source")
	}
	result, e := AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(decisions), bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if !result.Applied {
		t.Fatal("not applied")
	}
	retained, e := os.ReadFile(filepath.Join(p.Root, "assets/objects/sha256", report.EditedPPTXSHA256))
	if e != nil || !bytes.Equal(retained, packet.Edited) {
		t.Fatal("exact native evidence not retained")
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectGantt(p, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Document.Slides[0].NativeGeometry) > 0 {
		t.Fatal("semantic adoption retained stale geometry override")
	}
	var taskFound, gateFound bool
	for _, g := range inspect.Schedule.Groups {
		for _, l := range g.Lanes {
			for _, it := range l.Items {
				if it.Key == "delivery-task" {
					taskFound = true
					if math.Abs(it.From-1.25) > .000002 || math.Abs(it.To-3.75) > .000002 {
						t.Fatal("semantic intervals lost")
					}
				}
			}
		}
	}
	for _, g := range inspect.Schedule.Gates {
		if ganttKey(g.Key) == "approval" {
			gateFound = true
			if math.Abs(g.At-4.75) > .000002 {
				t.Fatal("gate period lost")
			}
		}
	}
	if !taskFound || !gateFound {
		t.Fatal("stable keys lost")
	}
	if _, e = AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(decisions), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale packet accepted")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}
func TestGanttSemanticAmbiguousMovementRemainsManual(t *testing.T) {
	p, b, node, bar := ganttSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	moved := objects[bar].geometry
	moved.X += 12
	moved.Y += 6
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: moved})
	report, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	var q GanttSemanticProposal
	for _, v := range report.Proposals {
		if v.SourceObject == bar {
			q = v
		}
	}
	if q.ID == "" || q.Status != "manual_review" || q.ProposedFrom != nil {
		t.Fatal("vertical movement inferred task timing")
	}
	d := GanttSemanticDecisions{Schema: GanttSemanticDecisionsSchema, ReportSHA256: GanttSemanticReportHash(report), Actor: "operator", Reason: "review", Decisions: []GanttSemanticDecision{{q.ID, "retime", "accept"}}}
	if _, e = AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(d), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("manual finding retimed")
	}
	d.Decisions[0].Action = "keep_source"
	if _, e = AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(d), bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}
func TestGanttSemanticChangedAxisAndForgedReportRefused(t *testing.T) {
	p, b, node, bar := ganttSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	var name string
	for n := range objects {
		if strings.HasPrefix(n, node+".periods.") && strings.HasSuffix(n, ".label") {
			if name == "" || n < name {
				name = n
			}
		}
	}
	moved := objects[name].geometry
	moved.X += 12
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{name: moved})
	if _, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("moved timeline inferred")
	}
	moved = objects[bar].geometry
	moved.X += 12
	packet = ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: moved})
	packet.Report.Geometry = nil
	if _, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatalf("mutated in-memory observation should be replaced by closed verified packet: %v", e)
	}
}
func TestGanttSemanticDecisionsStrict(t *testing.T) {
	valid := GanttSemanticDecisions{Schema: GanttSemanticDecisionsSchema, ReportSHA256: strings.Repeat("a", 64), Actor: "operator", Reason: "review", Decisions: []GanttSemanticDecision{{strings.Repeat("b", 64), "keep_source", "retained"}}}
	for _, raw := range [][]byte{append(canonical(valid), []byte("\n---\n{}\n")...), []byte(strings.Replace(string(canonical(valid)), `"actor":"operator"`, `"actor":"operator","unexpected":false`, 1)), []byte(strings.Replace(string(canonical(valid)), "keep_source", "use_native", 1))} {
		if _, e := DecodeGanttSemanticDecisions(raw, "decisions"); e == nil {
			t.Fatal("ambiguous decisions accepted")
		}
	}
}

func TestGanttSemanticRejectsWrongReportAndStaleDecisionWithoutWrites(t *testing.T) {
	p, b, node, bar := ganttSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	moved := objects[bar].geometry
	moved.X += 12
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: moved})
	report, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	var proposal GanttSemanticProposal
	for _, q := range report.Proposals {
		if q.SourceObject == bar {
			proposal = q
		}
	}
	d := GanttSemanticDecisions{Schema: GanttSemanticDecisionsSchema, ReportSHA256: strings.Repeat("0", 64), Actor: "operator", Reason: "review", Decisions: []GanttSemanticDecision{{proposal.ID, "retime", "confirm"}}}
	before := append([]byte(nil), p.Raw...)
	if _, e = AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(d), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("wrong report hash accepted")
	}
	d.ReportSHA256 = GanttSemanticReportHash(report)
	d.Decisions[0].ProposalID = strings.Repeat("1", 64)
	if _, e = AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(d), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("unknown proposal accepted")
	}
	if disk, _ := os.ReadFile(p.SourcePath); !bytes.Equal(disk, before) {
		t.Fatal("refused decisions wrote source")
	}
	_, template, e := diagramSlide(p, "plan-slide")
	if e != nil {
		t.Fatal(e)
	}
	_, e = compositionCandidate(p, "plan-slide", "guard-test", "operator", "Verify evidence guard", template, bundle(t), wmdesign.CandidateEngine, true, nil, nil, map[string][]byte{p.Document.Toolchain.Lockfile: []byte("forged lock")})
	if e == nil || !strings.Contains(e.Error(), "immutable build input changed") {
		t.Fatalf("evidence guard ignored: %v", e)
	}
	if disk, _ := os.ReadFile(p.SourcePath); !bytes.Equal(disk, before) {
		t.Fatal("guard refusal wrote source")
	}
}

func TestGanttSemanticSegmentedBarsAndBoundSourceStayExplicit(t *testing.T) {
	t.Run("segmented", func(t *testing.T) {
		p, _, node, bar := ganttSemanticFixture(t)
		idx, template, e := diagramSlide(p, "plan-slide")
		if e != nil {
			t.Fatal(e)
		}
		n, e := ganttNode(&template, node)
		if e != nil {
			t.Fatal(e)
		}
		schedule, _, _, e := ganttSource(n, p.Document.Slides[idx].Values)
		if e != nil {
			t.Fatal(e)
		}
		item := schedule.Groups[0].Lanes[0].Items[0]
		progress := .5
		item.Progress = &progress
		patch := ganttPatch(p, node, GanttOperation{Action: "set", Entity: "task", Group: schedule.Groups[0].Key, Lane: schedule.Groups[0].Lanes[0].Key, Key: item.Key, Task: &item})
		if _, e = PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
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
		objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
		moved := objects[bar].geometry
		moved.X += 12
		packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: moved})
		report, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
		if e != nil {
			t.Fatal(e)
		}
		for _, q := range report.Proposals {
			if q.SourceObject == bar {
				t.Fatal("partial segment inferred entire task")
			}
		}
		if len(report.UnresolvedGeometryIDs) == 0 {
			t.Fatal("segmented movement vanished")
		}
	})
	t.Run("bound", func(t *testing.T) {
		p, node := ganttCompositionFixture(t)
		if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
			t.Fatal(e)
		}
		b, e := ReadTextBaseline(p, "", "")
		if e != nil {
			t.Fatal(e)
		}
		packet, e := WriteGeometryReviewPacket(p, b, b.files["deck.pptx"], filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "materialized") {
			t.Fatalf("bound source silently moved: %v", e)
		}
	})
}

func TestGanttSemanticForgedClosedPacketCannotAuthorizeProposal(t *testing.T) {
	p, b, node, bar := ganttSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	moved := objects[bar].geometry
	moved.X += 12
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: moved})
	report := packet.Report
	report.Counts["forged_count"] = 1
	reportBytes := canonical(report)
	reportPath := filepath.Join(packet.Root, "report.json")
	if e := os.Chmod(reportPath, 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(reportPath, reportBytes, 0600); e != nil {
		t.Fatal(e)
	}
	manifestPath := filepath.Join(packet.Root, "manifest.json")
	manifestRaw, e := os.ReadFile(manifestPath)
	if e != nil {
		t.Fatal(e)
	}
	var manifest TextReviewPacketManifest
	if e = strictInto(json.RawMessage(manifestRaw), &manifest); e != nil {
		t.Fatal(e)
	}
	manifest.Files["report.json"] = TextReviewPacketFile{SHA256: digest(reportBytes), Size: int64(len(reportBytes))}
	if e = os.Chmod(manifestPath, 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(manifestPath, canonical(manifest), 0600); e != nil {
		t.Fatal(e)
	}
	packet, e = ReadTextReviewPacket(packet.Root)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "verified_inputs") {
		t.Fatalf("rehashed forged report authorized semantics: %v", e)
	}
}

func TestGanttSemanticNestedNamespaceAndCoupledLabelRegeneration(t *testing.T) {
	p, _, node, leafBar := ganttSemanticFixture(t)
	local := p.Document.LocalTemplates["plan"]
	local.Nodes = []Node{{ID: "program", Kind: "group", Nodes: []Node{{ID: "inner", Kind: "group", Nodes: local.Nodes}}}}
	p.Document.LocalTemplates["plan"] = local
	raw, e := json.Marshal(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
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
	bar := "program.inner." + leafBar
	label := strings.TrimSuffix(bar, ".segment-0") + ".label"
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	if objects[bar] == nil || objects[label] == nil {
		t.Fatal("qualified native objects missing")
	}
	movedBar, movedLabel := objects[bar].geometry, objects[label].geometry
	movedBar.X += 12
	movedLabel.X += 12
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: movedBar, label: movedLabel})
	report, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if report.NativeNodeID != "program.inner.schedule" || report.AxisParent != "program.inner.schedule" {
		t.Fatalf("namespace not retained: %+v", report)
	}
	var q GanttSemanticProposal
	for _, v := range report.Proposals {
		if v.SourceObject == bar {
			q = v
		}
	}
	if q.Status != "proposed" {
		t.Fatal("nested bar not proposed")
	}
	// Labels follow the accepted model on regeneration; independent native
	// geometry stays explicitly visible as unadopted evidence in this packet.
	labelFound := false
	for _, f := range packet.Report.Geometry {
		if f.Name == label {
			for _, id := range report.UnresolvedGeometryIDs {
				if f.ID == id {
					labelFound = true
				}
			}
		}
	}
	if !labelFound {
		t.Fatal("coupled label edit vanished from evidence")
	}
	d := GanttSemanticDecisions{Schema: GanttSemanticDecisionsSchema, ReportSHA256: GanttSemanticReportHash(report), Actor: "operator", Reason: "Confirm translated task", Decisions: []GanttSemanticDecision{{q.ID, "retime", "Task shifted horizontally"}}}
	if _, e = AdoptGanttSemantics(p, packet, "plan-slide", node, canonical(d), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	rebuilt, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	rebuiltObjects, _ := geometryFromPackage(t, rebuilt.files["deck.pptx"])
	if math.Abs(rebuiltObjects[bar].geometry.X-movedBar.X) > .001 || math.Abs(rebuiltObjects[label].geometry.X-movedLabel.X) > .001 {
		t.Fatal("regenerated bar/label did not carry reviewed period shift")
	}
	resized := rebuiltObjects[report.AxisParent].geometry
	resized.W -= 12
	packet = ganttSemanticPacket(t, p, rebuilt, map[string]NativeGeometry{report.AxisParent: resized})
	if _, e = ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "parent_missing_or_changed") {
		t.Fatalf("rescaled group inferred periods: %v", e)
	}
}

func TestGanttSemanticIntervalEventMixtureCannotInferMeaning(t *testing.T) {
	p, _, node, bar := ganttSemanticFixture(t)
	idx, local, e := diagramSlide(p, "plan-slide")
	if e != nil {
		t.Fatal(e)
	}
	n, e := ganttNode(&local, node)
	if e != nil {
		t.Fatal(e)
	}
	schedule, _, _, e := ganttSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		t.Fatal(e)
	}
	schedule.Groups[0].Lanes[0].Items[0].At = 2
	// The underlying renderer draws the Kind interval branch, but a contradictory
	// authored event coordinate is not qualified semantics for retiming.
	var args map[string]any
	if e = json.Unmarshal(canonical(schedule), &args); e != nil {
		t.Fatal(e)
	}
	for _, k := range []string{"type", "x", "y", "w"} {
		delete(args, k)
	}
	n.Arguments = args
	p.Document.LocalTemplates["plan"] = local
	raw, e := json.Marshal(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
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
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	moved := objects[bar].geometry
	moved.X += 12
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{bar: moved})
	report, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	for _, q := range report.Proposals {
		if q.SourceObject == bar {
			t.Fatal("interval/event mixture acquired inferred meaning")
		}
	}
	if len(report.UnresolvedGeometryIDs) == 0 {
		t.Fatal("invalid semantic source disappeared")
	}
}

func TestGanttSemanticOriginalItemGroupHorizontalMovement(t *testing.T) {
	p, b, node, bar := ganttSemanticFixture(t)
	objects, _ := geometryFromPackage(t, b.files["deck.pptx"])
	parent := strings.TrimSuffix(bar, ".segment-0")
	original, exists := objects[parent]
	if !exists {
		t.Fatal("original stable item group missing", parent)
	}
	width := 0.0
	for name, o := range objects {
		if strings.HasPrefix(name, node+".periods.") && strings.HasSuffix(name, ".label") {
			width = o.geometry.W
			break
		}
	}
	moved := original.geometry
	moved.X += width * .25
	packet := ganttSemanticPacket(t, p, b, map[string]NativeGeometry{parent: moved})
	report, e := ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, q := range report.Proposals {
		if q.SourceObject == bar {
			found = true
			if q.Status != "proposed" || math.Abs(*q.ProposedFrom-1.25) > .000002 || math.Abs(*q.ProposedTo-3.25) > .000002 {
				t.Fatal("group translation did not preserve authored duration", q)
			}
		}
	}
	if !found {
		t.Fatal("group movement omitted task proposal", report)
	}
	moved.Y++
	packet = ganttSemanticPacket(t, p, b, map[string]NativeGeometry{parent: moved})
	report, e = ProposeGanttSemantics(p, packet, "plan-slide", node, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	for _, q := range report.Proposals {
		if q.SourceObject == bar && q.Status != "manual_review" {
			t.Fatal("vertical group change became time", q)
		}
	}
}
