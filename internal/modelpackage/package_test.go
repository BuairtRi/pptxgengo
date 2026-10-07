package modelpackage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsUnpinnedClosedPackage(t *testing.T) {
	root := t.TempDir()
	report := Report{Schema: Schema, Directory: ".", Identity: PinnedIdentity(), Source: Source(), License: "Apache-2.0"}
	raw, _ := json.Marshal(report)
	for _, a := range Artifacts() {
		if e := os.WriteFile(filepath.Join(root, a.File), []byte("untrusted payload"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	for name, raw := range map[string][]byte{"manifest.json": raw, "LICENSE": License(), "README.txt": Readme()} {
		if e := os.WriteFile(filepath.Join(root, name), raw, 0600); e != nil {
			t.Fatal(e)
		}
	}
	if _, _, e := Verify(root); e == nil || !strings.Contains(e.Error(), "pin mismatch") {
		t.Fatal("caller-provided hashes trusted", e)
	}
	if e := os.WriteFile(filepath.Join(root, "extra"), []byte("unlisted"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, _, e := Verify(root); e == nil {
		t.Fatal("unclosed package accepted")
	}
}
func TestModelIdentityReturnsIndependentValues(t *testing.T) {
	one := PinnedIdentity()
	one.Artifacts[0].SHA256 = "changed"
	if PinnedIdentity().Artifacts[0].SHA256 == "changed" {
		t.Fatal("pin global changed")
	}
	raw := License()
	raw[0] = 'X'
	if License()[0] == 'X' {
		t.Fatal("license global changed")
	}
	bounds := FileBounds()
	bounds["model.onnx"] = 1
	if FileBounds()["model.onnx"] == 1 {
		t.Fatal("bounds global changed")
	}
}
func TestPinnedModelPackageClosure(t *testing.T) {
	root := os.Getenv("PPTXGENGO_EMBED_MODEL_DIR")
	if root == "" {
		t.Skip("optional pinned package")
	}
	report, hashes, e := Verify(root)
	if e != nil {
		t.Fatal(e)
	}
	if report.Identity.Revision != ModelRevision || len(hashes) != 6 {
		t.Fatal("model package pins missing")
	}
}
