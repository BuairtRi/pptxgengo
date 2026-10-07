package deckproject

import (
	"bytes"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/noreplacedir"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeEditabilityInventoriesActualCardOwnership(t *testing.T) {
	p, b := reconcileFixture(t)
	before := append([]byte(nil), p.Raw...)
	report, e := NativeEditability(b)
	if e != nil {
		t.Fatal(e)
	}
	if report.Schema != NativeEditabilitySchema || report.GenerationToken != b.Objects.Lineage.BuildToken || report.PPTXSHA256 != b.Receipt.Outputs["deck.pptx"] || report.DesktopQualification != "not_recorded" {
		t.Fatal("missing evidence pins", report)
	}
	if report.Counts["native_grpSp"] < 4 || report.Counts["field_plain_text_baseline"] < 3 || report.Counts["maximum_group_depth"] < 2 {
		t.Fatal("fixture grouping inventory missing", report.Counts)
	}
	units := map[string]NativeEditabilityUnit{}
	for _, u := range report.Objects {
		units[u.ShapeToken] = u
	}
	mapped := 0
	for _, u := range report.Objects {
		if strings.HasPrefix(u.SelectionName, "cards.items.source.") && len(u.Fields) == 1 {
			mapped++
			parent := units[u.ParentToken]
			if parent.SelectionName != "cards.items.source" || len(parent.ChildTokens) < 3 || u.GroupDepth != 2 || parent.Definition == "" {
				t.Fatalf("selection unit did not match actual group %+v %+v", u, parent)
			}
		}
	}
	if mapped != 2 {
		t.Fatalf("source title/body missing: %d", mapped)
	}
	repeat, e := NativeEditability(b)
	if e != nil || !bytes.Equal(canonical(report), canonical(repeat)) {
		t.Fatal("inventory nondeterministic", e)
	}
	if !bytes.Equal(p.Raw, before) {
		t.Fatal("inventory changed source")
	}
}

func TestNativeEditabilityRejectsMutableMetadata(t *testing.T) {
	if _, e := NativeEditability(nil); e == nil {
		t.Fatal("nil baseline accepted")
	}
	_, b := reconcileFixture(t)
	b.Objects.Objects[0].NativeName = "spoofed selection identity"
	if _, e := NativeEditability(b); e == nil {
		t.Fatal("metadata override accepted")
	}
}

func TestNativeEditabilityPilotListTableAndDiagram(t *testing.T) {
	p := example(t)
	titleZone := map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string", "maxLength": 70}}}
	local := func(id, kind string, args map[string]any, keys map[string][]string) {
		p.Document.LocalTemplates[id] = LocalTemplate{Name: "Synthetic native editing " + id, Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: titleZone, Nodes: []Node{{ID: id, Kind: "component", Placement: &Placement{Zone: "body", Span: &Span{Start: 1, Count: 12, Y: 36, H: 216}}, Definition: &Reference{Scope: "shared", ID: "wmds/component/" + kind}, Arguments: args, Keys: keys}}}
		p.Document.Slides = append(p.Document.Slides, Slide{ID: "pilot-" + id, ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: id}, Values: map[string]any{"title": "Synthetic " + id + " editing baseline"}})
	}
	local("list", "bullets", map[string]any{"items": []any{"Select the first item", "Edit the second item", "Review the third item"}}, map[string][]string{"items": {"first", "second", "third"}})
	local("table", "table", map[string]any{"cols": []any{map[string]any{"k": "step", "label": "Step", "w": 423}, map[string]any{"k": "owner", "label": "Owner", "w": 423}}, "rows": []any{map[string]any{"step": "Review", "owner": "Synthetic reviewer"}, map[string]any{"step": "Build", "owner": "Synthetic author"}}}, map[string][]string{"rows": {"review", "build"}})
	local("diagram", "block", map[string]any{"surface": "subtle", "text": "Synthetic editable input", "style": "body"}, nil)
	diagram := p.Document.LocalTemplates["diagram"]
	diagram.Nodes[0].Placement.Span.Count = 6
	diagram.Nodes[0].Placement.Span.H = 144
	diagram.Nodes = append(diagram.Nodes,
		Node{ID: "output", Kind: "component", Placement: &Placement{Zone: "body", Span: &Span{Start: 7, Count: 6, Y: 36, H: 144}}, Definition: &Reference{Scope: "shared", ID: "wmds/component/block"}, Arguments: map[string]any{"surface": "subtle", "text": "Synthetic editable output", "style": "body"}},
		Node{ID: "input-output", Kind: "component", Placement: &Placement{Zone: "body", Span: &Span{Start: 1, Count: 12, Y: 36, H: 144}}, Definition: &Reference{Scope: "shared", ID: "wmds/component/connector"}, Arguments: map[string]any{"points": []any{[]any{471, 252}, []any{489, 252}}, "head": "end"}},
	)
	p.Document.LocalTemplates["diagram"] = diagram
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
	pin(t, p)
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
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
	if report.Counts["family_table"] != 1 {
		t.Fatal("native table not inventoried", report.Counts)
	}
	foundTable := false
	families := map[string]bool{}
	for _, unit := range report.Objects {
		families[unit.SlideID] = true
		if unit.Family == "table" {
			foundTable = true
			if unit.TableCells != 6 {
				t.Fatal("cell count drift", unit)
			}
		}
	}
	if !foundTable || !families["pilot-list"] || !families["pilot-diagram"] {
		t.Fatal("pilot family inventory missing")
	}
	if output := os.Getenv("PPTXGENGO_EDITABILITY_PILOT_OUT"); output != "" {
		writeNativeEditabilityPilotFixture(t, p, report, output)
	}
}

// Opt-in desktop fixture preparation retains synthetic source, exact immutable
// build evidence and inventory outside testing's temporary directory. It opens
// no desktop app and makes no qualification claim.
func writeNativeEditabilityPilotFixture(t *testing.T, p *Project, report NativeEditabilityReport, destination string) {
	t.Helper()
	absolute, e := filepath.Abs(destination)
	if e != nil {
		t.Fatal(e)
	}
	parent, e := filepath.EvalSymlinks(filepath.Dir(absolute))
	if e != nil {
		t.Fatal(e)
	}
	absolute = filepath.Join(parent, filepath.Base(absolute))
	if _, e = os.Lstat(absolute); !os.IsNotExist(e) {
		t.Fatalf("pilot fixture destination must be new: %v", e)
	}
	stage, e := os.MkdirTemp(parent, ".editing-pilot-stage-")
	if e != nil {
		t.Fatal(e)
	}
	defer os.RemoveAll(stage)
	manifest := map[string]string{}
	e = filepath.WalkDir(p.Root, func(path string, entry fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		relative, e := filepath.Rel(p.Root, path)
		if e != nil {
			return e
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(stage, relative), 0700)
		}
		info, e := entry.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("pilot fixture requires regular files")
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		name := filepath.ToSlash(relative)
		manifest[name] = digest(raw)
		mode := os.FileMode(0600)
		if strings.HasPrefix(name, "builds/") {
			mode = 0444
		}
		return writeExclusive(filepath.Join(stage, relative), raw, mode)
	})
	if e != nil {
		t.Fatal(e)
	}
	raw := canonical(report)
	if e = writeExclusive(filepath.Join(stage, "native-editability.json"), raw, 0444); e != nil {
		t.Fatal(e)
	}
	manifest["native-editability.json"] = digest(raw)
	if e = writeExclusive(filepath.Join(stage, "pilot-fixture-hashes.json"), canonical(manifest), 0444); e != nil {
		t.Fatal(e)
	}
	if e = noreplacedir.Publish(stage, absolute); e != nil {
		t.Fatal(e)
	}
	t.Logf("synthetic editing pilot fixture: %s", absolute)
}
