package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func nativeImportCLIInput(t *testing.T) (string, []byte) {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, err := z.Create("ppt/slides/slide1.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write([]byte(`<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:cSld><p:spTree><p:sp><p:nvSpPr><p:cNvPr id="2" name="Selected textbox"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr><p:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="2540000" cy="762000"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr><p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:r><a:t>Imported CLI source</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`))
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "native.pptx")
	if err = os.WriteFile(path, b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path, b.Bytes()
}

func nativeImportCLIFileSnapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[rel], err = os.ReadFile(path)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

func TestNativeImportCLIFlags(t *testing.T) {
	in, _ := nativeImportCLIInput(t)
	for _, args := range [][]string{
		{}, {"unknown"}, {"inspect"}, {"inspect", "--in", in, "extra"},
		{"inspect", "--in", in, "--apply=false"}, {"inspect", "--in", in, "--project", "."},
		{"inspect", "--in", in, "--slide", "unwanted"}, {"inspect", "--in", in, "--map", "unwanted"},
		{"inspect", "--in", in, "--bundle", "unwanted"}, {"inspect", "--in", in, "--engine", "unwanted"},
		{"patch", "--in", in}, {"patch", "--in", in, "--slide", "stable-page"},
	} {
		if err := runProjectNativeImport(args); err == nil {
			t.Fatalf("invalid/ambiguous native-import flags accepted: %v", args)
		}
	}
}

func TestNativeImportCLIInspectPreviewApplyAndStaleGuard(t *testing.T) {
	in, data := nativeImportCLIInput(t)
	var inv deckproject.NativeImportInventory
	if err := json.Unmarshal(portableCommandJSON(t, "native-import", "inspect", "--in", in), &inv); err != nil || len(inv.Objects) != 1 || !inv.Objects[0].Supported || inv.FormattingPolicy != "reauthor_with_design_style" {
		t.Fatal(inv, err)
	}
	root := projectDefaultFixture(t)
	p, err := deckproject.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = deckproject.Pin(p, bundle, wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	packageHash := hex.EncodeToString(hash[:])
	mapping := deckproject.NativeImportMap{Schema: deckproject.NativeImportSchema, ExpectedSourceSHA256: p.SourceHash(), ExpectedPPTXSHA256: packageHash, Actor: "CLI operator", Reason: "Import explicit design-styled text", FormattingPolicy: "reauthor_with_design_style", Selections: []deckproject.NativeImportSelection{{Part: "ppt/slides/slide1.xml", Name: "Selected textbox", NodeID: "imported-cli", Kind: "text", Style: "body", Align: "left", Rect: &wmdesign.Rect{X: 0, Y: 0, W: 200, H: 60}}}}
	raw, err := json.Marshal(mapping)
	if err != nil {
		t.Fatal(err)
	}
	mapPath := filepath.Join(t.TempDir(), "mapping.json")
	if err = os.WriteFile(mapPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"native-import", "patch", "--in", in, "--project", root, "--slide", "stable-page", "--map", mapPath, "--bundle", bundle}
	before := nativeImportCLIFileSnapshot(t, root)
	var preview deckproject.CompositionResult
	if err = json.Unmarshal(portableCommandJSON(t, args...), &preview); err != nil || preview.Applied || preview.Evidence == nil || !reflect.DeepEqual(before, nativeImportCLIFileSnapshot(t, root)) {
		t.Fatal("preview changed files or failed to disclose evidence", preview, err)
	}
	var applied deckproject.CompositionResult
	if err = json.Unmarshal(portableCommandJSON(t, append(args, "--apply")...), &applied); err != nil || !applied.Applied || applied.AfterSHA256 != preview.AfterSHA256 {
		t.Fatal(applied, err)
	}
	retained, err := os.ReadFile(filepath.Join(root, "assets/objects/sha256", packageHash))
	if err != nil || !bytes.Equal(retained, data) {
		t.Fatal("CLI did not retain exact raw package", err)
	}
	after := nativeImportCLIFileSnapshot(t, root)
	if err = runProjectNativeImport(args[1:]); err == nil || !strings.Contains(err.Error(), "hash mismatch") || !reflect.DeepEqual(after, nativeImportCLIFileSnapshot(t, root)) {
		t.Fatal("stale import map was accepted or changed project", err)
	}
}
