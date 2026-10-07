package installstate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/browsingfixture"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowsingInstallationClosure(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{}
	allowed := map[string]bool{}
	if e := allowBrowsingFiles(root, files, allowed); e != nil || len(allowed) != 0 {
		t.Fatal("legacy no-deck package changed", e)
	}
	if e := os.Mkdir(filepath.Join(root, "browsing"), 0755); e != nil {
		t.Fatal(e)
	}
	put := func(name string, value any) {
		t.Helper()
		var raw []byte
		if b, ok := value.([]byte); ok {
			raw = b
		} else {
			raw, _ = json.Marshal(value)
		}
		if e := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), raw, 0644); e != nil {
			t.Fatal(e)
		}
		sum := sha256.Sum256(raw)
		files[name] = fmt.Sprintf("%x", sum)
	}
	put("browsing/template-library.pptx", browsingfixture.Deck(t, "templates"))
	put("browsing/reusable-slides.pptx", browsingfixture.Deck(t, "reusable"))
	for _, pair := range []struct{ manifest, deck, kind string }{{"browsing/template-library.manifest.json", "browsing/template-library.pptx", "templates"}, {"browsing/reusable-slides.manifest.json", "browsing/reusable-slides.pptx", "reusable"}} {
		put(pair.manifest, browsingfixture.Manifest(t, pair.kind, files[pair.deck]))
	}
	hashes := map[string]string{}
	for name, hash := range files {
		hashes[name] = hash
	}
	put("browsing-manifest.json", browsingfixture.Inventory(t, hashes))
	if e := allowBrowsingFiles(root, files, allowed); e != nil || len(allowed) != 5 {
		t.Fatal(e, allowed)
	}
	files["browsing/template-library.pptx"] = "tampered"
	if e := allowBrowsingFiles(root, files, map[string]bool{}); e == nil {
		t.Fatal("tampered deck accepted")
	}
	files["browsing/template-library.pptx"] = hashes["browsing/template-library.pptx"]
	delete(files, "browsing/reusable-slides.pptx")
	if e := allowBrowsingFiles(root, files, map[string]bool{}); e == nil {
		t.Fatal("missing second deck accepted")
	}
}

func TestBrowsingFullInstallationChecksTypedClosure(t *testing.T) {
	// Only the closed browsing inputs and full manifest are needed to exercise
	// rejection before the later runtime/authoring resource checks.
	root := t.TempDir()
	if e := os.Mkdir(filepath.Join(root, "browsing"), 0755); e != nil {
		t.Fatal(e)
	}
	files := map[string]string{}
	put := func(name string, raw []byte) {
		t.Helper()
		if e := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), raw, 0644); e != nil {
			t.Fatal(e)
		}
		files[name] = browsingfixture.Hash(raw)
	}
	for _, pair := range []struct{ name, kind string }{{"template-library", "templates"}, {"reusable-slides", "reusable"}} {
		deck := "browsing/" + pair.name + ".pptx"
		put(deck, browsingfixture.Deck(t, pair.kind))
		raw := browsingfixture.Manifest(t, pair.kind, files[deck])
		if pair.kind == "templates" {
			var m map[string]any
			if e := json.Unmarshal(raw, &m); e != nil {
				t.Fatal(e)
			}
			m["coverage"].(map[string]any)["expected_templates"] = 2
			raw, _ = json.Marshal(m)
		}
		put("browsing/"+pair.name+".manifest.json", raw)
	}
	browsingHashes := map[string]string{}
	for name, hash := range files {
		browsingHashes[name] = hash
	}
	raw := browsingfixture.Inventory(t, browsingHashes)
	put("browsing-manifest.json", raw)
	raw, _ = json.Marshal(map[string]any{"schema": "pptxgengo.local-release-manifest.v1", "version": "4.2.0", "selected_bundle": "v11", "file_count": len(files), "files_sha256": files})
	put("release-manifest.json", raw)
	if _, e := Verify(root); e == nil || !strings.Contains(e.Error(), "browsing.coverage_omission") {
		t.Fatal("full installation skipped typed browsing closure", e)
	}
}
