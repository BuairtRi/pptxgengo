package installstate

import (
	"fmt"
	"github.com/buairtri/pptxgengo/internal/browsingartifact"
	"path/filepath"
)

// Browsing is optional for legacy CLI packages. When supplied its exact closure
// is required and checked; these hashes provide integrity, not an approval or
// signature claim. Release signing covers the enclosing installation archive.
func allowBrowsingFiles(root string, files map[string]string, allowed map[string]bool) error {
	if _, exists := files["browsing-manifest.json"]; !exists {
		return nil
	}
	var inventory struct {
		Schema string            `json:"schema"`
		Files  map[string]string `json:"files_sha256"`
	}
	if e := readJSON(filepath.Join(root, "browsing-manifest.json"), &inventory); e != nil {
		return e
	}
	names := []string{"browsing/template-library.pptx", "browsing/template-library.manifest.json", "browsing/reusable-slides.pptx", "browsing/reusable-slides.manifest.json"}
	if inventory.Schema != "pptxgengo.release-browsing-files.v1" || len(inventory.Files) != len(names) {
		return fmt.Errorf("browsing installation inventory incomplete")
	}
	for _, name := range names {
		expected, ok := inventory.Files[name]
		if !ok || !digestPattern.MatchString(expected) || files[name] != expected {
			return fmt.Errorf("browsing installation hash mismatch: %s", name)
		}
	}
	for _, pair := range []struct{ manifest, deck, kind string }{{names[1], names[0], "templates"}, {names[3], names[2], "reusable"}} {
		raw, e := browsingartifact.ReadManifest(filepath.Join(root, filepath.FromSlash(pair.manifest)))
		if e != nil {
			return e
		}

		if e = browsingartifact.ValidateFile(raw, pair.kind, files[pair.deck], filepath.Join(root, filepath.FromSlash(pair.deck))); e != nil {
			return e
		}
	}
	for _, name := range names {
		allowed[name] = true
	}
	allowed["browsing-manifest.json"] = true
	return nil
}
