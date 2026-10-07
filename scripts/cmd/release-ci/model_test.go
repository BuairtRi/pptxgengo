package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/modelpackage"
)

func TestModelReleaseIdentityAndAbsentEvidence(t *testing.T) {
	root := t.TempDir()
	version := "v4.1.1"
	commit := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if e := verifyModelRelease(root, version, commit); e == nil {
		t.Fatal("missing proof accepted")
	}
	ev := ModelEvidence{Schema: "pptxgengo.offline-model-release/v1", Version: version, Commit: commit, Archive: modelArchiveName(version), Identity: modelpackage.PinnedIdentity(), License: "Apache-2.0", Source: modelpackage.Source()}
	ev.Identity.Revision = "untrusted"
	if e := writeJSON(filepath.Join(root, ev.Archive+".evidence.json"), ev); e != nil {
		t.Fatal(e)
	}
	if e := verifyModelRelease(root, version, commit); e == nil {
		t.Fatal("unpinned identity accepted")
	}
}
func TestPinnedOfflineModelReleaseReproducibleAndRelocatable(t *testing.T) {
	from := os.Getenv("PPTXGENGO_EMBED_MODEL_DIR")
	if from == "" {
		t.Skip("optional pinned package")
	}
	version := "v4.1.1"
	commit := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	first, second := filepath.Join(t.TempDir(), "first"), filepath.Join(t.TempDir(), "second")
	if e := assembleModel(from, first, version, commit); e != nil {
		t.Fatal(e)
	}
	if e := assembleModel(from, second, version, commit); e != nil {
		t.Fatal(e)
	}
	name := modelArchiveName(version)
	a, _ := digest(filepath.Join(first, name))
	b, _ := digest(filepath.Join(second, name))
	if a == "" || a != b {
		t.Fatal("model archive not deterministic")
	}
	for _, suffix := range []string{".sbom.json", ".vulnerabilities.json"} {
		if e := os.WriteFile(filepath.Join(first, name+suffix), []byte("owned fixture only; not a scan verdict"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	dest := t.TempDir()
	if e := copyModelRelease(first, dest, version, commit); e != nil {
		t.Fatal(e)
	}
	if e := verifyModelRelease(dest, version, commit); e != nil {
		t.Fatal(e)
	}
	root := filepath.Join(t.TempDir(), "extracted with spaces")
	if e := extract(filepath.Join(dest, name), root); e != nil {
		t.Fatal(e)
	}
	report, _, e := modelpackage.Verify(root)
	if e != nil || report.Directory != "." {
		t.Fatal("runner path retained", e)
	}
	if e := assembleModel(from, first, version, commit); e == nil {
		t.Fatal("existing output replaced")
	}
	var ev ModelEvidence
	if e := readJSON(filepath.Join(dest, name+".evidence.json"), &ev); e != nil {
		t.Fatal(e)
	}
	ev.Files["model.onnx"] = "changed"
	if e := writeJSON(filepath.Join(dest, name+".evidence.json"), ev); e != nil {
		t.Fatal(e)
	}
	if e := verifyModelRelease(dest, version, commit); e == nil {
		t.Fatal("changed payload proof accepted")
	}
	if e := verifyModelRelease(first, "v4.1.2", commit); e == nil {
		t.Fatal("different tag accepted")
	}
}
