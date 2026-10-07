package deckproject

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// The demo uses genuine before/after component planners with identical copy,
// allocations and density. It records structure, not desktop acceptance.
func TestNativeEditabilityComponentComparisonDemo(t *testing.T) {
	out := os.Getenv("PPTXGENGO_NATIVE_COMPONENT_DEMO_OUT")
	if out == "" {
		t.Skip("set a new private Documents output directory")
	}
	p := example(t)
	p.Document.ID, p.Document.Title = "native-component-demo", "Native component editing comparison"
	p.Document.LocalTemplates = map[string]LocalTemplate{}
	p.Document.Slides = nil
	p.Document.Assets = nil
	compositions := map[string]CompositionEntry{}
	items := []any{"Review the source", "Confirm the owners and decisions with the team before sharing the updated presentation.", "Share the deck"}
	keys := map[string][]string{"items": {"review", "confirm", "share"}}
	cardCopy := "Update the source, review the changes with the team, and share the latest approved deck."
	place := func(start int, y, h float64) *Placement {
		return &Placement{Zone: "body", Span: &Span{Start: start, Count: 6, Y: y, H: h}}
	}
	component := func(id, kind string, start int, args map[string]any, keys map[string][]string) Node {
		return Node{ID: id, Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/" + kind}, Placement: place(start, 36, 216), Arguments: args, Keys: keys}
	}
	for _, density := range []string{"comfortable", "compact", "dense"} {
		for _, family := range []string{"lists", "cards", "tables"} {
			id := family + "-" + density
			nodes := []Node{
				{ID: "before-label", Kind: "text", Placement: place(1, 0, 18), Style: "label", Ink: "secondary", Text: "Current"},
				{ID: "after-label", Kind: "text", Placement: place(7, 0, 18), Style: "label", Ink: "secondary", Text: "Native editing"},
			}
			before, after := "", ""
			switch family {
			case "lists":
				before, after = "bullets", "editable-list"
				nodes = append(nodes, component("before", before, 1, map[string]any{"items": items}, keys), component("after", after, 7, map[string]any{"items": items}, keys))
			case "cards":
				before, after = "card", "editable-card"
				nodes = append(nodes, component("before", before, 1, map[string]any{"surface": "subtle", "title": "Keep the deck editable", "titleInk": "primary", "pad": 12, "gap": 8, "body": []any{map[string]any{"p": cardCopy}}}, map[string][]string{"body": {"copy"}}), component("after", after, 7, map[string]any{"surface": "subtle", "title": "Keep the deck editable", "body": cardCopy}, nil))
			case "tables":
				before, after = "table", "editable-table"
				args := map[string]any{"cols": []any{map[string]any{"k": "step", "label": "Step", "w": 207}, map[string]any{"k": "owner", "label": "Owner", "w": 207}}, "rows": []any{map[string]any{"step": "Review", "owner": "Reviewer"}, map[string]any{"step": "Build", "owner": "Author"}}}
				rowKeys := map[string][]string{"rows": {"review", "build"}}
				nodes = append(nodes, component("before", before, 1, args, rowKeys), component("after", after, 7, args, rowKeys))
			}
			p.Document.LocalTemplates[id] = LocalTemplate{Name: "Controlled " + family + " comparison", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string", "maxLength": 70}}}, Nodes: nodes}
			p.Document.Slides = append(p.Document.Slides, Slide{ID: id, ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: id}, Density: density, HeaderDensity: "comfortable", Values: map[string]any{"title": fmt.Sprintf("%s editing / %s", strings.Title(family), strings.Title(density))}, Notes: "Synthetic comparison. Same copy and dimensions on both sides. Left: current renderer. Right: native editing candidate. Try extending the second bullet and pressing Enter; editing/moving/resizing the card; selecting the table and editing a cell. Save edits to a separate copy. Desktop acceptance requires actual observation."})
			compositions[id] = CompositionEntry{Purpose: "Compare editing ergonomics and visual fidelity", ChosenTemplate: id, Candidates: []string{before, after}, Rationale: "The operator requested a controlled before/after editing demonstration. A stock single-renderer template cannot compare both object structures at identical density and allocation."}
		}
	}
	log, e := editYAMLNode(CompositionLog{Schema: "pptxgengo.composition-log.v1", Slides: compositions})
	if e != nil {
		t.Fatal(e)
	}
	logRaw, e := encodeSourceYAML(log)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(p.Root, "composition-log.yaml"), logRaw, 0600); e != nil {
		t.Fatal(e)
	}
	node, e := editYAMLNode(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := encodeSourceYAML(node)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	bundle, e := filepath.Abs("../../library/wm-design-system/v11")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Pin(p, bundle, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle, Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	report, e := NativeEditability(b)
	if e != nil {
		t.Fatal(e)
	}
	tables := 0
	for _, u := range report.Objects {
		if strings.HasPrefix(u.SlideID, "tables-") && u.SelectionName == "after.native" && (u.NativeKind != "graphicFrame" || u.GroupDepth != 0 || u.TableCells != 6) {
			t.Fatal("table wrapper retained", u)
		}
		if strings.HasPrefix(u.SlideID, "tables-") && u.SelectionName == "after.native" {
			tables++
		}
	}
	if tables != 3 {
		t.Fatal("missing table comparison coverage", tables)
	}
	var layout wmdesign.Report
	if e = json.Unmarshal(b.files["layout-report.json"], &layout); e != nil {
		t.Fatal(e)
	}
	if len(layout.Slides) != 9 {
		t.Fatal("demo coverage changed")
	}
	writeNativeEditabilityPilotFixture(t, p, report, out)
}

// Opt-in verification of a real desktop-edited copy, with no Office control.
func TestNativeEditabilityComponentComparisonSuppliedEdits(t *testing.T) {
	root := os.Getenv("PPTXGENGO_NATIVE_COMPONENT_DEMO_VERIFY_ROOT")
	if root == "" {
		t.Skip("supply the private comparison fixture root after desktop editing")
	}
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ReadTextBaseline(p, "", "")
	if err != nil {
		t.Fatal(err)
	}
	edited, err := readReconcileFile(filepath.Join(root, "native-editing-demo-edited.pptx"), lineageMaxPackage)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := InspectNativeLineage(edited, b.Objects)
	if err != nil || len(actual.Issues) != 0 || len(actual.Objects) != len(b.inspection.Objects) {
		t.Fatal("desktop lineage changed", err, actual.Issues)
	}
	objects := map[string]NativeLineageObject{}
	for _, object := range actual.Objects {
		objects[object.ShapeToken] = object
	}
	var cardBefore, cardAfter pilotNativeRect
	checked := 0
	for _, object := range b.Objects.Objects {
		current, ok := objects[object.ShapeToken]
		if !ok {
			t.Fatal("desktop dropped generation identity", object.ShapeToken)
		}
		if object.NativeName != "after" && object.NativeName != "after.native" {
			continue
		}
		if current.ParentToken != "" {
			t.Fatal("candidate gained a wrapper", current)
		}
		paragraphs := nativeParagraphs(current.shape)
		switch object.SlideID {
		case "lists-comfortable":
			if len(paragraphs) != 4 || paragraphs[0].Text != "Review the source" || !strings.HasPrefix(paragraphs[1].Text, "Confirm all of the reviewers,") || paragraphs[2].Text != "Share the deck" || paragraphs[3].Text != "Ask colleagues for feedback" {
				t.Fatal("list desktop task changed", paragraphs)
			}
			for _, paragraph := range paragraphs {
				if !strings.Contains(strings.Join(paragraph.ReviewItems, ","), "bullet_or_numbered_paragraph") {
					t.Fatal("new item lost native bullet", paragraph)
				}
			}
			checked++
		case "cards-compact":
			if len(paragraphs) != 2 || paragraphs[0].Text != "Keep the deck editable" || paragraphs[1].Text != "Refresh the source, review the changes with the team, and share the latest approved deck." {
				t.Fatal("card role text changed", paragraphs)
			}
			for _, prior := range b.inspection.Objects {
				if prior.ShapeToken == object.ShapeToken {
					cardBefore, _, err = pilotDiagramGeometry(prior.shape)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			cardAfter, _, err = pilotDiagramGeometry(current.shape)
			if err != nil || cardAfter.X != cardBefore.X || cardAfter.W != cardBefore.W || cardAfter.H != cardBefore.H || cardAfter.Y <= cardBefore.Y {
				t.Fatal("whole card move/geometry changed", err, cardBefore, cardAfter)
			}
			checked++
		case "tables-compact":
			if current.Kind != "graphicFrame" || len(paragraphs) != 6 {
				t.Fatal("table structure changed", current, paragraphs)
			}
			for i, want := range []string{"STEP", "OWNER", "Reviewed", "Reviewer", "Build", "Author"} {
				if paragraphs[i].Text != want {
					t.Fatal("unexpected table cell change", i, paragraphs[i])
				}
			}
			checked++
		}
	}
	if checked != 3 {
		t.Fatal("missing desktop task coverage", checked)
	}
	review, err := ReconcileText(p, b, edited)
	if err != nil || len(review.ManualReview) == 0 {
		t.Fatal("native edits must remain visible for review", err)
	}
	again, err := ReadTextBaseline(p, "", "")
	if err != nil || again.ReceiptSHA256 != b.ReceiptSHA256 {
		t.Fatal("baseline changed", err)
	}
	record := map[string]any{"schema": "pptxgengo.native-component-desktop-demo.v1", "platform": runtime.GOOS, "baseline_pptx_sha256": b.Receipt.Outputs["deck.pptx"], "edited_pptx_sha256": digest(edited), "receipt_sha256": b.ReceiptSHA256, "retained_object_count": len(actual.Objects), "card_before": cardBefore, "card_after": cardAfter, "card_vertical_move_pt": float64(cardAfter.Y-cardBefore.Y) / 12700, "tasks": []string{"extend native list paragraph; subsequent items reflow", "Enter inserts a native fourth bullet", "move combined card and independently edit body", "directly edit table cell"}, "limits": []string{"Windows not qualified", "resize/alignment not qualified", "stock template migration awaits demo review", "visual observations recorded in private screenshots; not pixel parity proof"}}
	if err = writeExclusive(filepath.Join(root, "desktop-verification.json"), canonical(record), 0444); err != nil {
		t.Fatal(err)
	}
	t.Logf("verified real desktop demo: retained %d object identities; card moved %.2fpt", len(actual.Objects), float64(cardAfter.Y-cardBefore.Y)/12700)
}
