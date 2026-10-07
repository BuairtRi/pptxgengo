package deckproject

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

const roundTripFixtureSchema = "pptxgengo.native-roundtrip-fixture.v1"

type desktopRoundTripFixture struct {
	Schema         string                     `json:"schema"`
	BuildID        string                     `json:"build_id"`
	ReceiptSHA256  string                     `json:"receipt_sha256"`
	BaselineSHA256 string                     `json:"baseline_sha256"`
	SlideOrder     []string                   `json:"slide_order"`
	NativeNames    []string                   `json:"native_names"`
	Plan           nativeexport.RoundTripPlan `json:"plan"`
}

type desktopRoundTripEvidence struct {
	Schema           string                           `json:"schema"`
	Status           string                           `json:"status"`
	Scope            string                           `json:"scope"`
	Execution        string                           `json:"execution"`
	HumanAcceptance  string                           `json:"human_acceptance"`
	Platform         string                           `json:"verification_platform"`
	GoVersion        string                           `json:"go_version"`
	VerifiedAt       string                           `json:"verified_at"`
	BuildID          string                           `json:"baseline_build_id"`
	ReceiptSHA256    string                           `json:"baseline_receipt_sha256"`
	BaselineSHA256   string                           `json:"baseline_sha256"`
	SavedAsSHA256    string                           `json:"saved_as_sha256"`
	EditedSHA256     string                           `json:"edited_sha256"`
	EditedOrder      []string                         `json:"edited_slide_order"`
	ManualReview     []TextReconciliationIssue        `json:"manual_review"`
	Failures         []string                         `json:"failures"`
	RebuiltBuildID   string                           `json:"rebuilt_build_id,omitempty"`
	DesktopExecution *nativeexport.RoundTripExecution `json:"desktop_execution,omitempty"`
	AutomaticNumbers []desktopAutomaticNumber         `json:"automatic_slide_numbers,omitempty"`
}

type desktopAutomaticNumber struct {
	ShapeToken string `json:"shape_token"`
	SlideToken string `json:"slide_token"`
	Before     string `json:"before"`
	After      string `json:"after"`
}

// Only the cached value of an existing, single slide-number field may follow
// the verified native slide order. Other dynamic fields and ordinary text still
// require exact equality. This does not adopt numbering, geometry or formatting.
func roundTripSlideNumber(shape *xmlNode) (string, string, bool) {
	if shape == nil || shape.Name != (xml.Name{Space: lineagePML, Local: "sp"}) {
		return "", "", false
	}
	body := lineageChild(shape, lineagePML, "txBody")
	if body == nil {
		return "", "", false
	}
	var field *xmlNode
	paragraphs := 0
	for _, paragraph := range body.Children {
		if paragraph.Name != (xml.Name{Space: drawingML, Local: "p"}) {
			continue
		}
		paragraphs++
		for _, child := range paragraph.Children {
			if child.Name.Space != drawingML {
				return "", "", false
			}
			switch child.Name.Local {
			case "pPr", "endParaRPr":
			case "fld":
				if field != nil || attr(child, "type") != "slidenum" || attr(child, "id") == "" {
					return "", "", false
				}
				field = child
			default:
				return "", "", false
			}
		}
	}
	if paragraphs != 1 || field == nil {
		return "", "", false
	}
	var text *xmlNode
	for _, child := range field.Children {
		if child.Name.Space != drawingML {
			return "", "", false
		}
		switch child.Name.Local {
		case "rPr":
		case "t":
			if text != nil || len(child.Children) != 0 {
				return "", "", false
			}
			text = child
		default:
			return "", "", false
		}
	}
	if text == nil {
		return "", "", false
	}
	return attr(field, "id"), text.Text, true
}

func roundTripNumberChange(before, after NativeLineageObject, oldOrder, newOrder []string) bool {
	oldID, oldText, oldOK := roundTripSlideNumber(before.shape)
	newID, newText, newOK := roundTripSlideNumber(after.shape)
	ordinal := func(order []string, token string) string {
		for i, t := range order {
			if t == token {
				return strconv.Itoa(i + 1)
			}
		}
		return ""
	}
	return oldOK && newOK && oldID == newID && before.ShapeToken == after.ShapeToken &&
		before.SlideToken == after.SlideToken && oldText == ordinal(oldOrder, before.SlideToken) &&
		newText == ordinal(newOrder, after.SlideToken)
}

func roundTripWrite(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func roundTripOrder(data []byte) ([]string, error) {
	pkg, err := openLineagePackage(data)
	if err != nil {
		return nil, err
	}
	rels, err := pkg.relationships("")
	if err != nil {
		return nil, err
	}
	presentation := ""
	for _, rel := range rels {
		if rel.Type == lineageRML+"/officeDocument" {
			if presentation != "" || rel.Mode == "External" {
				return nil, fmt.Errorf("ambiguous presentation")
			}
			presentation, err = lineageTarget("", rel.Target)
			if err != nil {
				return nil, err
			}
		}
	}
	root, err := pkg.tree(presentation)
	if err != nil {
		return nil, err
	}
	slides := lineageChild(root, lineagePML, "sldIdLst")
	if slides == nil {
		return nil, fmt.Errorf("missing slide order")
	}
	rels, err = pkg.relationships(presentation)
	if err != nil {
		return nil, err
	}
	order := []string{}
	for _, entry := range slides.Children {
		rel := rels[lineageAttr(entry, lineageRML, "id")]
		if rel.Type != lineageRML+"/slide" || rel.Mode == "External" {
			return nil, fmt.Errorf("invalid slide relationship")
		}
		part, err := lineageTarget(presentation, rel.Target)
		if err != nil {
			return nil, err
		}
		root, err := pkg.tree(part)
		if err != nil {
			return nil, err
		}
		tags, err := pkg.tags(part, lineageChild(root, lineagePML, "cSld"))
		if err != nil {
			return nil, err
		}
		if !validLineageToken(tags["PPTXGENGO_SLIDE"]) {
			return nil, fmt.Errorf("missing slide token")
		}
		order = append(order, tags["PPTXGENGO_SLIDE"])
	}
	return order, nil
}

func roundTripFixturePlan(b *TextBaseline) (desktopRoundTripFixture, error) {
	f := desktopRoundTripFixture{Schema: roundTripFixtureSchema, BuildID: b.Receipt.BuildID, ReceiptSHA256: b.ReceiptSHA256, BaselineSHA256: digest(b.files["deck.pptx"]), Plan: nativeexport.RoundTripPlan{DeckToken: b.Objects.Lineage.DeckToken, BuildToken: b.Objects.Lineage.BuildToken}}
	var err error
	f.SlideOrder, err = roundTripOrder(b.files["deck.pptx"])
	if err != nil {
		return f, err
	}
	if len(f.SlideOrder) != 2 || f.SlideOrder[0] == f.SlideOrder[1] {
		return f, fmt.Errorf("fixture requires two distinct slides")
	}
	f.Plan.MoveSlideToken = f.SlideOrder[1]
	for i, name := range []string{"cards.items.source.title", "cards.items.source.body.copy", "cards.items.build.title"} {
		found := false
		for _, o := range b.Objects.Objects {
			if o.NativeName != name {
				continue
			}
			if found || len(o.Fields) != 1 || o.Fields[0].Status != "plain_text_baseline" {
				return f, fmt.Errorf("fixture field ambiguous or unsupported: %s", name)
			}
			found = true
			f.NativeNames = append(f.NativeNames, name)
			var slideToken string
			for _, native := range b.inspection.Objects {
				if native.ShapeToken == o.ShapeToken {
					slideToken = native.SlideToken
				}
			}
			if !validLineageToken(slideToken) {
				return f, fmt.Errorf("missing fixture slide identity")
			}
			f.Plan.Edits = append(f.Plan.Edits, nativeexport.RoundTripEdit{ShapeToken: o.ShapeToken, SlideToken: slideToken, Before: o.NativeText, After: []string{"Native author", "Native editable source.", "Native build"}[i]})
		}
		if !found {
			return f, fmt.Errorf("missing fixture field: %s", name)
		}
	}
	return f, nil
}

func prepareRoundTripFixture(t *testing.T, destination string) (*Project, *TextBaseline, desktopRoundTripFixture) {
	t.Helper()
	if err := os.Mkdir(destination, 0700); err != nil {
		t.Fatal(err)
	}
	source := example(t)
	root := filepath.Join(destination, "project")
	err := filepath.WalkDir(source.Root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source.Root, path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, rel)
		if entry.IsDir() {
			return os.Mkdir(target, 0700)
		}
		data, err := readReconcileFile(path, 64<<20)
		if err != nil {
			return err
		}
		return roundTripWrite(target, data)
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	pin(t, p)
	if _, err = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	b, err := ReadTextBaseline(p, "", "")
	if err != nil {
		t.Fatal(err)
	}
	f, err := roundTripFixturePlan(b)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"baseline.pptx": b.files["deck.pptx"], "plan.json": canonical(f)} {
		if err = roundTripWrite(filepath.Join(destination, name), data); err != nil {
			t.Fatal(err)
		}
	}
	instructions := "# Owned native round-trip fixture\n\nSynthetic example only. Preserve baseline.pptx and project/builds unchanged.\n\n1. Open baseline.pptx in PowerPoint. Save As a NEW saved-as.pptx before editing.\n2. Use the Selection Pane to find and edit these named fixture objects:\n"
	for i, edit := range f.Plan.Edits {
		instructions += fmt.Sprintf("   - %s: %q -> %q\n", f.NativeNames[i], edit.Before, edit.After)
	}
	instructions += "3. Move the second slide to the first position. Save As a NEW edited.pptx. Close only your fixture.\n4. Run the verifier described in docs/native-roundtrip.md. It checks actual tag identity, text and order, then adopts only the prescribed synthetic changes and rebuilds the disposable project.\n\nMove/align/resize/table/diagram ergonomics and native visual acceptance are separate human tasks; this fixture does not record their approval.\n"
	if err = roundTripWrite(filepath.Join(destination, "instructions.md"), []byte(instructions)); err != nil {
		t.Fatal(err)
	}
	return p, b, f
}

func loadRoundTripFixture(root string) (*Project, *TextBaseline, desktopRoundTripFixture, error) {
	var f desktopRoundTripFixture
	p, err := Load(filepath.Join(root, "project"))
	if err != nil {
		return nil, nil, f, err
	}
	raw, err := readReconcileFile(filepath.Join(root, "plan.json"), 2<<20)
	if err != nil {
		return nil, nil, f, err
	}
	if err = strictInto(json.RawMessage(raw), &f); err != nil {
		return nil, nil, f, err
	}
	b, err := ReadTextBaseline(p, f.BuildID, f.ReceiptSHA256)
	if err != nil {
		return nil, nil, f, err
	}
	want, err := roundTripFixturePlan(b)
	if err != nil {
		return nil, nil, f, err
	}
	if !bytes.Equal(canonical(f), canonical(want)) {
		return nil, nil, f, fmt.Errorf("fixture plan does not match receipt-pinned baseline")
	}
	baseline, err := readReconcileFile(filepath.Join(root, "baseline.pptx"), lineageMaxPackage)
	if err != nil {
		return nil, nil, f, err
	}
	if !bytes.Equal(baseline, b.files["deck.pptx"]) {
		return nil, nil, f, fmt.Errorf("fixture baseline changed")
	}
	return p, b, f, nil
}

func verifyRoundTripFixture(t *testing.T, root, savedPath, editedPath, destination, execution string, desktop *nativeexport.RoundTripExecution) (desktopRoundTripEvidence, error) {
	t.Helper()
	out := desktopRoundTripEvidence{Schema: "pptxgengo.native-roundtrip-evidence.v1", Status: "fail", Scope: "synthetic fixture: Save As tag survival, three named plain-text proposals, slide reordering, bounded adoption and headless rebuild", Execution: execution, HumanAcceptance: "not_recorded", Platform: runtime.GOOS + "/" + runtime.GOARCH, GoVersion: runtime.Version(), VerifiedAt: time.Now().UTC().Format(time.RFC3339Nano), Failures: []string{}, ManualReview: []TextReconciliationIssue{}, DesktopExecution: desktop}
	p, b, f, err := loadRoundTripFixture(root)
	if err != nil {
		return out, err
	}
	saved, err := readReconcileFile(savedPath, lineageMaxPackage)
	if err != nil {
		return out, err
	}
	edited, err := readReconcileFile(editedPath, lineageMaxPackage)
	if err != nil {
		return out, err
	}
	if err = os.Mkdir(destination, 0700); err != nil {
		return out, err
	}
	out.BuildID = f.BuildID
	out.ReceiptSHA256 = f.ReceiptSHA256
	out.BaselineSHA256 = f.BaselineSHA256
	out.SavedAsSHA256 = digest(saved)
	out.EditedSHA256 = digest(edited)
	if desktop != nil && (desktop.InputSHA256 != out.BaselineSHA256 || desktop.SavedAsSHA256 != out.SavedAsSHA256 || desktop.EditedSHA256 != out.EditedSHA256 || desktop.OSVersion == "" || !desktop.Closed) {
		out.Failures = append(out.Failures, "desktop execution hashes/metadata differ from actual retained fixture bytes")
	}
	if err = roundTripWrite(filepath.Join(destination, "saved-as.pptx"), saved); err != nil {
		return out, err
	}
	if err = roundTripWrite(filepath.Join(destination, "edited.pptx"), edited); err != nil {
		return out, err
	}
	saveResult, err := InspectNativeLineage(saved, b.Objects)
	if err != nil {
		out.Failures = append(out.Failures, "Save As identity: "+err.Error())
	} else if len(saveResult.Issues) != 0 || len(saveResult.Objects) != len(b.Objects.Objects) {
		out.Failures = append(out.Failures, "Save As failed complete tag identity survival")
	}
	baselineText := map[string]string{}
	for _, object := range b.Objects.Objects {
		baselineText[object.ShapeToken] = object.NativeText
	}
	for _, object := range saveResult.Objects {
		if object.shape == nil || nativeParagraphText(nativeParagraphs(object.shape)) != baselineText[object.ShapeToken] {
			out.Failures = append(out.Failures, "Save As changed native text before the declared edits")
		}
	}
	if err = roundTripWrite(filepath.Join(destination, "save-as-lineage.json"), canonical(saveResult)); err != nil {
		return out, err
	}
	saveReport, err := ReconcileText(p, b, saved)
	if err != nil {
		out.Failures = append(out.Failures, "Save As comparison: "+err.Error())
	} else {
		if saveReport.Counts["native_only"] != 0 || saveReport.Counts["conflict"] != 0 {
			out.Failures = append(out.Failures, "Save As changed supported text before editing")
		}
		out.ManualReview = append(out.ManualReview, saveReport.ManualReview...)
		if err = roundTripWrite(filepath.Join(destination, "save-as-reconciliation.json"), canonical(saveReport)); err != nil {
			return out, err
		}
	}
	saveOrder, err := roundTripOrder(saved)
	if err != nil || !reflect.DeepEqual(saveOrder, f.SlideOrder) {
		out.Failures = append(out.Failures, "initial Save As changed or lost slide order")
	}
	identity, err := InspectNativeLineage(edited, b.Objects)
	if err != nil {
		out.Failures = append(out.Failures, "edited identity: "+err.Error())
	} else if len(identity.Issues) != 0 || len(identity.Objects) != len(b.Objects.Objects) {
		out.Failures = append(out.Failures, "edited deck failed complete tag identity survival")
	}
	for _, edit := range f.Plan.Edits {
		baselineText[edit.ShapeToken] = edit.After
	}
	out.EditedOrder, err = roundTripOrder(edited)
	if err != nil || !reflect.DeepEqual(out.EditedOrder, []string{f.SlideOrder[1], f.SlideOrder[0]}) {
		out.Failures = append(out.Failures, "actual edited slide order is not the requested reorder")
	}
	baselineObjects := map[string]NativeLineageObject{}
	for _, object := range b.inspection.Objects {
		baselineObjects[object.ShapeToken] = object
	}
	for _, object := range identity.Objects {
		if object.shape == nil || nativeParagraphText(nativeParagraphs(object.shape)) != baselineText[object.ShapeToken] {
			before := baselineObjects[object.ShapeToken]
			if roundTripNumberChange(before, object, f.SlideOrder, out.EditedOrder) {
				out.AutomaticNumbers = append(out.AutomaticNumbers, desktopAutomaticNumber{object.ShapeToken, object.SlideToken, baselineText[object.ShapeToken], nativeParagraphText(nativeParagraphs(object.shape))})
				continue
			}
			out.Failures = append(out.Failures, "edited native text differs from the three prescribed changes")
		}
	}
	if err = roundTripWrite(filepath.Join(destination, "edited-lineage.json"), canonical(identity)); err != nil {
		return out, err
	}
	packet, err := WriteTextReviewPacket(p, b, edited, filepath.Join(destination, "review"))
	if err != nil {
		out.Failures = append(out.Failures, "reconciliation: "+err.Error())
	} else {
		out.ManualReview = append(out.ManualReview, packet.Report.ManualReview...)
		if packet.Report.Counts["native_only"] != 3 || packet.Report.Counts["conflict"] != 0 || packet.Report.Counts["yaml_only"] != 0 {
			out.Failures = append(out.Failures, "expected exactly three native-only text proposals and unchanged YAML")
		}
		decisions := TextReviewDecisions{Schema: TextReviewDecisionsSchema, ReportSHA256: packet.ReportSHA256, Actor: "automated synthetic round-trip fixture; no human acceptance"}
		for _, edit := range f.Plan.Edits {
			found := false
			for _, field := range packet.Report.Fields {
				if field.ShapeToken != edit.ShapeToken {
					continue
				}
				if found || field.Status != "native_only" || field.EditedNative != edit.After || field.Baseline != edit.Before || field.CurrentYAML != edit.Before {
					out.Failures = append(out.Failures, "unsupported, ambiguous or incorrect proposal for prescribed fixture field")
				}
				found = true
				decisions.Decisions = append(decisions.Decisions, TextReviewDecision{FieldID: field.ID, Action: "use_native", Reason: "Exact prescribed edit in a disposable synthetic qualification fixture"})
			}
			if !found {
				out.Failures = append(out.Failures, "missing prescribed fixture proposal")
			}
		}
		if len(out.Failures) == 0 {
			raw := canonical(decisions)
			if err = roundTripWrite(filepath.Join(destination, "fixture-decisions.json"), raw); err != nil {
				return out, err
			}
			receipt, err := AdoptTextReviewPacket(p, packet, raw, bundle(t), wmdesign.CandidateEngine)
			if err != nil {
				return out, err
			}
			if len(receipt.Changed) != 3 {
				return out, fmt.Errorf("fixture adoption did not change three fields")
			}
			after, err := Load(p.SourcePath)
			if err != nil {
				return out, err
			}
			repeated, err := AdoptTextReviewPacket(after, packet, raw, bundle(t), wmdesign.CandidateEngine)
			if err != nil || !bytes.Equal(canonical(receipt), canonical(repeated)) {
				return out, fmt.Errorf("repeated adoption is not idempotent: %v", err)
			}
			report, err := ReconcileText(after, b, edited)
			if err != nil || report.Counts["matching_changes"] != 3 || report.Counts["native_only"] != 0 {
				return out, fmt.Errorf("repeated reconciliation did not retain matching changes: %v", err)
			}
			built, err := Build(after, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
			if err != nil {
				return out, err
			}
			rebuilt, err := ReadTextBaseline(after, built.BuildID, "")
			if err != nil {
				return out, err
			}
			for i, name := range f.NativeNames {
				found := false
				for _, o := range rebuilt.Objects.Objects {
					if o.NativeName == name {
						if found || o.NativeText != f.Plan.Edits[i].After {
							return out, fmt.Errorf("rebuild did not reproduce fixture field")
						}
						found = true
					}
				}
				if !found {
					return out, fmt.Errorf("rebuild lost fixture field")
				}
			}
			if _, err = ReadTextBaseline(after, b.Receipt.BuildID, b.ReceiptSHA256); err != nil {
				return out, err
			}
			retained, err := readReconcileFile(filepath.Join(after.Root, filepath.FromSlash(receipt.RetainedEdited)), lineageMaxPackage)
			if err != nil || !bytes.Equal(retained, edited) {
				return out, fmt.Errorf("adoption did not preserve exact edited fixture")
			}
			out.RebuiltBuildID = built.BuildID
			if err = roundTripWrite(filepath.Join(destination, "adoption.json"), canonical(receipt)); err != nil {
				return out, err
			}
		}
	}
	if len(out.Failures) == 0 {
		out.Status = "pass"
	}
	if err = roundTripWrite(filepath.Join(destination, "evidence.json"), canonical(out)); err != nil {
		return out, err
	}
	if out.Status != "pass" {
		return out, fmt.Errorf("native round-trip fixture failed: %s", strings.Join(out.Failures, "; "))
	}
	return out, nil
}

func roundTripOptIn(t *testing.T, key string) string {
	t.Helper()
	if testing.Short() {
		t.Skip("desktop qualification is separate from short tests")
	}
	value := os.Getenv(key)
	if value == "" {
		t.Skip("set " + key + " for this explicit fixture operation")
	}
	return value
}

func TestNativeRoundTripPrepare(t *testing.T) {
	out := roundTripOptIn(t, "PPTXGENGO_ROUNDTRIP_PREPARE_OUT")
	prepareRoundTripFixture(t, out)
}

func TestNativeRoundTripVerify(t *testing.T) {
	root := roundTripOptIn(t, "PPTXGENGO_ROUNDTRIP_FIXTURE")
	saved := roundTripOptIn(t, "PPTXGENGO_ROUNDTRIP_SAVED_AS")
	edited := roundTripOptIn(t, "PPTXGENGO_ROUNDTRIP_EDITED")
	out := roundTripOptIn(t, "PPTXGENGO_ROUNDTRIP_VERIFY_OUT")
	if _, err := verifyRoundTripFixture(t, root, saved, edited, out, "operator-supplied files; native application provenance not independently verified", nil); err != nil {
		t.Fatal(err)
	}
}

func TestNativeRoundTripLiveWindows(t *testing.T) {
	root := roundTripOptIn(t, "PPTXGENGO_ROUNDTRIP_WINDOWS_OUT")
	if runtime.GOOS != "windows" {
		t.Fatal("live COM round-trip requires Windows")
	}
	_, b, f := prepareRoundTripFixture(t, root)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	execution, err := nativeexport.WindowsRoundTrip(ctx, filepath.Join(root, "desktop"), b.files["deck.pptx"], f.Plan)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = verifyRoundTripFixture(t, root, execution.SavedAs, execution.Edited, filepath.Join(root, "verification"), "Windows PowerPoint COM execution observed in this run", &execution); err != nil {
		t.Fatal(err)
	}
}

func roundTripReorder(t *testing.T, data []byte) []byte {
	t.Helper()
	return lineageEdit(t, data, "ppt/presentation.xml", func(raw []byte) []byte {
		root, err := lineageSpans(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, node := range root.children {
			if node.node.Name == (xml.Name{Space: lineagePML, Local: "sldIdLst"}) && len(node.children) == 2 {
				a, z := node.children[0], node.children[1]
				result, err := lineageApply(raw, []lineagePatch{{a.start, z.end, string(raw[z.start:z.end]) + string(raw[a.end:z.start]) + string(raw[a.start:a.end])}})
				if err != nil {
					t.Fatal(err)
				}
				return result
			}
		}
		t.Fatal("missing two-slide fixture")
		return nil
	})
}

func TestNativeRoundTripCore(t *testing.T) {
	root := filepath.Join(t.TempDir(), "fixture with spaces")
	_, b, f := prepareRoundTripFixture(t, root)
	values := map[string]string{}
	for i, name := range f.NativeNames {
		values[name] = f.Plan.Edits[i].After
	}
	edited := roundTripReorder(t, reconcileEditFields(t, b, values))
	path := filepath.Join(root, "headless-edited.pptx")
	if err := roundTripWrite(path, edited); err != nil {
		t.Fatal(err)
	}
	out, err := verifyRoundTripFixture(t, root, filepath.Join(root, "baseline.pptx"), path, filepath.Join(root, "verification"), "hermetic XML fixture only; no native application executed", nil)
	if err != nil || out.Status != "pass" || out.RebuiltBuildID == "" || out.HumanAcceptance != "not_recorded" || out.DesktopExecution != nil {
		t.Fatalf("%v %+v", err, out)
	}
	if _, err := verifyRoundTripFixture(t, root, filepath.Join(root, "baseline.pptx"), path, filepath.Join(root, "verification"), "repeat", nil); err == nil {
		t.Fatal("verification output overwritten")
	}
}

func TestNativeRoundTripAutomaticNumbers(t *testing.T) {
	root := filepath.Join(t.TempDir(), "fixture")
	_, b, f := prepareRoundTripFixture(t, root)
	values := map[string]string{}
	for i, name := range f.NativeNames {
		values[name] = f.Plan.Edits[i].After
	}
	edited := roundTripReorder(t, reconcileEditFields(t, b, values))
	// Mimic PowerPoint refreshing the cached values of existing slide-number
	// fields after reorder. Source content and field identities remain exact.
	for i, part := range []string{"ppt/slides/slide1.xml", "ppt/slides/slide2.xml"} {
		edited = lineageEdit(t, edited, part, func(raw []byte) []byte {
			from := []byte(fmt.Sprintf("<a:t>%d</a:t></a:fld>", i+1))
			to := []byte(fmt.Sprintf("<a:t>%d</a:t></a:fld>", 2-i))
			if bytes.Count(raw, from) != 1 {
				t.Fatal("expected one generated slide-number cache")
			}
			return bytes.Replace(raw, from, to, 1)
		})
	}
	path := filepath.Join(root, "edited.pptx")
	if err := roundTripWrite(path, edited); err != nil {
		t.Fatal(err)
	}
	out, err := verifyRoundTripFixture(t, root, filepath.Join(root, "baseline.pptx"), path, filepath.Join(root, "verification"), "hermetic refreshed-number fixture", nil)
	if err != nil || out.Status != "pass" || len(out.AutomaticNumbers) != 2 || out.RebuiltBuildID == "" {
		t.Fatalf("%v %+v", err, out)
	}
}

func TestNativeRoundTripNumberChangeBounds(t *testing.T) {
	shape := func(field string) *xmlNode {
		n, err := readXML(strings.NewReader(`<p:sp xmlns:p="` + lineagePML + `" xmlns:a="` + drawingML + `"><p:txBody><a:p>` + field + `</a:p></p:txBody></p:sp>`))
		if err != nil {
			t.Fatal(err)
		}
		return n.Children[0]
	}
	field := `<a:fld id="same" type="slidenum"><a:t>1</a:t></a:fld>`
	before := NativeLineageObject{ShapeToken: "shape", SlideToken: "a", shape: shape(field)}
	good := strings.Replace(field, ">1<", ">2<", 1)
	for _, tc := range []struct {
		name, field string
		want        bool
	}{
		{"refresh", good, true},
		{"wrong-number", strings.Replace(good, ">2<", ">9<", 1), false},
		{"changed-id", strings.Replace(good, `id="same"`, `id="other"`, 1), false},
		{"other-field", strings.Replace(good, `type="slidenum"`, `type="datetime"`, 1), false},
		{"static-text", `<a:r><a:t>2</a:t></a:r>`, false},
		{"added-text", good + `<a:r><a:t>surprise</a:t></a:r>`, false},
		{"duplicate-field", good + good, false},
		{"extra-paragraph", good + `</a:p><a:p>`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			after := before
			after.shape = shape(tc.field)
			if got := roundTripNumberChange(before, after, []string{"a", "b"}, []string{"b", "a"}); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
	if roundTripNumberChange(before, before, []string{"b", "a"}, []string{"a", "b"}) {
		t.Fatal("incorrect original ordinal accepted")
	}
}

func TestNativeRoundTripRejectsWrongEditsOrderAndPlan(t *testing.T) {
	for _, mutation := range []string{"no-reorder", "wrong-text", "plan-drift", "baseline-drift", "missing-tags", "metadata-drift"} {
		t.Run(mutation, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "fixture")
			_, b, f := prepareRoundTripFixture(t, root)
			values := map[string]string{}
			for i, name := range f.NativeNames {
				values[name] = f.Plan.Edits[i].After
			}
			if mutation == "wrong-text" {
				values[f.NativeNames[0]] = "Unexpected copy"
			}
			edited := reconcileEditFields(t, b, values)
			if mutation != "no-reorder" {
				edited = roundTripReorder(t, edited)
			}
			if mutation == "plan-drift" {
				f.Plan.Edits[0].After = "Tampered plan"
				if err := os.WriteFile(filepath.Join(root, "plan.json"), canonical(f), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mutation == "baseline-drift" {
				if err := os.WriteFile(filepath.Join(root, "baseline.pptx"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mutation == "missing-tags" {
				edited = lineageEdit(t, edited, "ppt/tags/pptxgengo1.xml", func(raw []byte) []byte {
					return bytes.ReplaceAll(raw, []byte(f.Plan.BuildToken), []byte(strings.Repeat("0", 64)))
				})
			}
			path := filepath.Join(root, "edited.pptx")
			if err := roundTripWrite(path, edited); err != nil {
				t.Fatal(err)
			}
			var desktop *nativeexport.RoundTripExecution
			if mutation == "metadata-drift" {
				desktop = &nativeexport.RoundTripExecution{Closed: true, OSVersion: "simulated OS; mismatched hashes"}
			}
			out, err := verifyRoundTripFixture(t, root, filepath.Join(root, "baseline.pptx"), path, filepath.Join(root, "verification"), "hermetic rejection fixture", desktop)
			if err == nil || out.Status == "pass" {
				t.Fatalf("%s accepted: %+v", mutation, out)
			}
			p, err := Load(filepath.Join(root, "project"))
			if err != nil {
				t.Fatal(err)
			}
			if p.SourceHash() != b.Receipt.SourceSHA256 {
				t.Fatal("failed verification changed project source")
			}
		})
	}
}
