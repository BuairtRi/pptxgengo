package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

const nativeImportPlainShape = `<p:sp><p:nvSpPr><p:cNvPr id="2" name="Source textbox"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr><p:spPr><a:xfrm><a:off x="127000" y="254000"/><a:ext cx="2540000" cy="762000"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:solidFill><a:srgbClr val="FF0000"/></a:solidFill></p:spPr><p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr b="1" sz="2800"><a:solidFill><a:srgbClr val="FF0000"/></a:solidFill></a:rPr><a:t>Imported source text</a:t></a:r></a:p></p:txBody></p:sp>`

func nativeImportPackage(t *testing.T, shapes string) []byte {
	t.Helper()
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	w, err := z.Create("ppt/slides/slide1.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprintf(w, `<p:sld xmlns:p="%s" xmlns:a="%s" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><p:cSld><p:spTree>%s</p:spTree></p:cSld></p:sld>`, lineagePML, drawingML, shapes)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func nativeImportMapping(p *Project, data []byte, id string) NativeImportMap {
	return NativeImportMap{Schema: NativeImportSchema, ExpectedSourceSHA256: p.SourceHash(), ExpectedPPTXSHA256: digest(data), Actor: "import operator", Reason: "Reauthor selected native text with design styling", FormattingPolicy: "reauthor_with_design_style", Selections: []NativeImportSelection{{Part: "ppt/slides/slide1.xml", Name: "Source textbox", NodeID: id, Kind: "text", Style: "body", Align: "left", Rect: &wmdesign.Rect{X: 0, Y: 0, W: 300, H: 48}}}}
}

func TestNativeImportPreviewApplyAndEvidenceDeduplication(t *testing.T) {
	p, _ := compositionTransactionFixture(t)
	data := nativeImportPackage(t, nativeImportPlainShape)
	inv, err := InspectNativeImport(data)
	if err != nil || len(inv.Objects) != 1 || !inv.Objects[0].Supported || inv.FormattingPolicy != "reauthor_with_design_style" || inv.Objects[0].Text != "Imported source text" {
		t.Fatal(inv, err)
	}
	mapping := nativeImportMapping(p, data, "imported-one")
	before := compositionFileSnapshot(t, p.Root)
	preview, err := ImportNative(p, "local-composition", data, mapping, bundle(t), wmdesign.CandidateEngine, false)
	if err != nil || preview.Applied || preview.AfterSHA256 == p.SourceHash() || preview.Evidence == nil || !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
		t.Fatal("invalid preview or preview wrote evidence", preview, err)
	}
	applied, err := ImportNative(p, "local-composition", data, mapping, bundle(t), wmdesign.CandidateEngine, true)
	if err != nil || !applied.Applied {
		t.Fatal(applied, err)
	}
	next, err := Load(p.Root)
	if err != nil || next.SourceHash() != preview.AfterSHA256 {
		t.Fatal(next, err)
	}
	template := next.Document.LocalTemplates["editorial-photo"]
	list, index := findDiagramNode(&template.Nodes, "imported-one")
	if list == nil {
		t.Fatal("missing imported source node")
	}
	node := (*list)[index]
	if node.Kind != "text" || node.Style != "body" || node.Text != "Imported source text" || node.Placement.Rect.X != 0 || node.Placement.Rect.Y != 0 {
		t.Fatalf("native text not explicitly reauthored: %+v", node)
	}
	asset := "assets/objects/sha256/" + digest(data)
	retained, err := os.ReadFile(filepath.Join(p.Root, asset))
	if err != nil || !bytes.Equal(retained, data) {
		t.Fatal("exact source package not retained", err)
	}
	receiptRaw, err := os.ReadFile(filepath.Join(p.Root, applied.Decision))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Evidence nativeImportEvidence `json:"evidence"`
	}
	if err = json.Unmarshal(receiptRaw, &receipt); err != nil || receipt.Evidence.SourceAsset != asset || receipt.Evidence.Mapping.FormattingPolicy != "reauthor_with_design_style" || receipt.Evidence.SourcePPTXSHA256 != digest(data) {
		t.Fatal("receipt lacks explicit formatting policy/raw source address", string(receiptRaw), err)
	}
	second := nativeImportMapping(next, data, "imported-two")
	second.Selections[0].Rect.X = 340
	if _, err = ImportNative(next, "local-composition", data, second, bundle(t), wmdesign.CandidateEngine, true); err != nil {
		t.Fatal(err)
	}
	copies := []string{}
	for rel, b := range compositionFileSnapshot(t, p.Root) {
		if bytes.Equal(b, data) {
			copies = append(copies, rel)
		}
	}
	if len(copies) != 1 || copies[0] != asset {
		t.Fatalf("raw package was duplicated rather than content-addressed once: %v", copies)
	}
}

func TestNativeImportRejectsUnsupportedNativeObjects(t *testing.T) {
	for _, tc := range []struct{ name, shape, reason string }{
		{"group", `<p:grpSp><p:nvGrpSpPr><p:cNvPr id="8" name="Source textbox"/></p:nvGrpSpPr>` + nativeImportPlainShape + `</p:grpSp>`, "requires_top_level_rectangular_text_shape"},
		{"duplicate-name", nativeImportPlainShape + nativeImportPlainShape, "missing_or_ambiguous_top_level_name"},
		{"ellipse", strings.Replace(nativeImportPlainShape, `prst="rect"`, `prst="ellipse"`, 1), "rectangle_geometry_required"},
		{"rotation", strings.Replace(nativeImportPlainShape, `<a:xfrm>`, `<a:xfrm rot="60000">`, 1), "rotation_flip_or_empty_extent"},
		{"relationship", strings.Replace(nativeImportPlainShape, `<p:cNvPr id="2" name="Source textbox"/>`, `<p:cNvPr id="2" name="Source textbox"><a:hlinkClick r:id="rId1"/></p:cNvPr>`, 1), "relationship_bearing_shape_unsupported"},
		{"rich-runs", strings.Replace(nativeImportPlainShape, `</a:r></a:p>`, `</a:r><a:r><a:t>More styled text</a:t></a:r></a:p>`, 1), "rich_text_fields_or_bullets_unsupported"},
		{"bullet", strings.Replace(nativeImportPlainShape, `<a:p>`, `<a:p><a:pPr><a:buChar char="•"/></a:pPr>`, 1), "rich_text_fields_or_bullets_unsupported"},
		{"empty-text", strings.Replace(nativeImportPlainShape, "Imported source text", " ", 1), "nonempty_plain_text_required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := nativeImportPackage(t, tc.shape)
			inv, err := InspectNativeImport(data)
			if err != nil || len(inv.Objects) == 0 {
				t.Fatal(inv, err)
			}
			for _, o := range inv.Objects {
				if o.Supported || !strings.Contains(strings.Join(o.Reasons, ","), tc.reason) {
					t.Fatal("unsupported native object classified as importable", o)
				}
			}
			p, _ := compositionTransactionFixture(t)
			before := compositionFileSnapshot(t, p.Root)
			result, err := ImportNative(p, "local-composition", data, nativeImportMapping(p, data, "rejected"), bundle(t), wmdesign.CandidateEngine, true)
			if err == nil || result.Applied || !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
				t.Fatal("unsupported import changed project", result, err)
			}
		})
	}
}

func TestNativeImportHashCollisionAndOverflowAreAtomic(t *testing.T) {
	for _, kind := range []string{"source-hash", "package-hash", "policy", "overflow", "collision", "stale-disk"} {
		t.Run(kind, func(t *testing.T) {
			p, _ := compositionTransactionFixture(t)
			data := nativeImportPackage(t, nativeImportPlainShape)
			mapping := nativeImportMapping(p, data, "new-import")
			switch kind {
			case "source-hash":
				mapping.ExpectedSourceSHA256 = strings.Repeat("0", 64)
			case "package-hash":
				mapping.ExpectedPPTXSHA256 = strings.Repeat("0", 64)
			case "policy":
				mapping.FormattingPolicy = "preserve_native_formatting"
			case "overflow":
				mapping.Selections[0].Rect.W = 10000
			case "collision":
				path := filepath.Join(p.Root, "assets/objects/sha256", digest(data))
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("different evidence"), 0600); err != nil {
					t.Fatal(err)
				}
			case "stale-disk":
				if err := os.WriteFile(p.SourcePath, append([]byte("# Concurrent source edit\n"), p.Raw...), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := compositionFileSnapshot(t, p.Root)
			result, err := ImportNative(p, "local-composition", data, mapping, bundle(t), wmdesign.CandidateEngine, true)
			if err == nil || result.Applied || !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
				t.Fatal("refused import must preserve bytes and report not applied", result, err)
			}
		})
	}
}

func TestNativeImportMapIsStrict(t *testing.T) {
	p, _ := compositionTransactionFixture(t)
	data := nativeImportPackage(t, nativeImportPlainShape)
	raw := canonical(nativeImportMapping(p, data, "imported"))
	for _, bad := range [][]byte{bytes.Replace(raw, []byte(`"schema":`), []byte(`"unknown":true,"schema":`), 1), append(append([]byte(nil), raw...), []byte("\n---\nother: document\n")...), []byte("schema: one\nschema: two\n")} {
		if _, err := DecodeNativeImportMap(bad, "map.yaml"); err == nil {
			t.Fatal("ambiguous map accepted", string(bad))
		}
	}
}

func TestNativeImportDefaultPlacementPreservesWorldOrigin(t *testing.T) {
	p, _ := compositionTransactionFixture(t)
	destination, err := InspectDiagram(p, "local-composition", bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	x, y := destination.Frame.Body.X+12, destination.Frame.Body.Y+12
	shape := strings.Replace(nativeImportPlainShape, `<a:off x="127000" y="254000"/>`, fmt.Sprintf(`<a:off x="%.0f" y="%.0f"/>`, x*12700, y*12700), 1)
	data := nativeImportPackage(t, shape)
	mapping := nativeImportMapping(p, data, "default-import")
	mapping.Selections[0].Rect = nil
	before := compositionFileSnapshot(t, p.Root)
	preview, err := ImportNative(p, "local-composition", data, mapping, bundle(t), wmdesign.CandidateEngine, false)
	if err != nil || !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
		t.Fatal(preview, err)
	}
	for _, native := range preview.Inspection.FinalNative {
		if native.Name == "default-import" {
			if math.Abs(native.WorldBounds.X-x) > 0.0001 || math.Abs(native.WorldBounds.Y-y) > 0.0001 {
				t.Fatal("default import translated slide-absolute coordinates incorrectly", native, destination.Frame.Body)
			}
			return
		}
	}
	t.Fatal("missing imported native transform")
}

func TestNativeImportDefaultOutsideBodyIsRefusedAtomically(t *testing.T) {
	p, _ := compositionTransactionFixture(t)
	data := nativeImportPackage(t, strings.Replace(nativeImportPlainShape, `<a:off x="127000" y="254000"/>`, `<a:off x="-1270000" y="-1270000"/>`, 1))
	mapping := nativeImportMapping(p, data, "outside-import")
	mapping.Selections[0].Rect = nil
	before := compositionFileSnapshot(t, p.Root)
	result, err := ImportNative(p, "local-composition", data, mapping, bundle(t), wmdesign.CandidateEngine, true)
	if err == nil || result.Applied || !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
		t.Fatal("outside-body default import was accepted or changed files", result, err)
	}
}

func TestNativeImportEditableBlockReauthorsExplicitSurface(t *testing.T) {
	p, _ := compositionTransactionFixture(t)
	data := nativeImportPackage(t, nativeImportPlainShape)
	mapping := nativeImportMapping(p, data, "imported-block")
	mapping.Selections[0].Kind = "editable-block"
	mapping.Selections[0].Surface = "subtle"
	preview, err := ImportNative(p, "local-composition", data, mapping, bundle(t), wmdesign.CandidateEngine, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range preview.Inspection.Nodes {
		if node.ID != "imported-block" {
			continue
		}
		if node.Kind != "component" || node.Definition == nil || node.Definition.ID != "wmds/component/editable-block" || node.Arguments["surface"] != "subtle" || node.Arguments["style"] != "body" || node.Arguments["text"] != "Imported source text" {
			t.Fatal("imported block did not use explicit design component/surface", node)
		}
		return
	}
	t.Fatal("missing imported editable block")
}
