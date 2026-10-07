package installstate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/browsingfixture"
	"os"
	"path/filepath"
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
	put("browsing-manifest.json", map[string]any{"schema": "pptxgengo.release-browsing-files.v1", "files_sha256": hashes})
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
