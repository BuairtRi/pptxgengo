package installstate

import (
	"fmt"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/browsingartifact"
)

// Browsing is optional for legacy CLI packages. When supplied its exact closure
// is required and checked; these hashes provide integrity, not an approval or
// signature claim. Release signing covers the enclosing installation archive.
func allowBrowsingFiles(root string, files map[string]string, allowed map[string]bool) error {
	if _, exists := files[browsingartifact.TemplateCatalogInventoryName]; exists {
		raw, e := browsingartifact.ReadManifest(filepath.Join(root, browsingartifact.TemplateCatalogInventoryName))
		if e != nil {
			return e
		}
		in, e := browsingartifact.ReadTemplateCatalogInventory(raw)
		if e != nil {
			return e
		}
		if e = browsingartifact.ValidateTemplateCatalogInventory(root, in, files, ""); e != nil {
			return e
		}
		for _, name := range browsingartifact.TemplateCatalogFiles(in) {
			allowed[name] = true
		}
		allowed[browsingartifact.TemplateCatalogInventoryName] = true
		return nil
	}
	if _, exists := files["browsing-manifest.json"]; !exists {
		return nil
	}
	raw, e := browsingartifact.ReadManifest(filepath.Join(root, "browsing-manifest.json"))
	if e != nil {
		return e
	}
	inventory, e := browsingartifact.ReadInventory(raw)
	if e != nil {
		return e
	}
	names := []string{"browsing/template-library.pptx", "browsing/template-library.manifest.json", "browsing/reusable-slides.pptx", "browsing/reusable-slides.manifest.json"}
	for _, name := range names {
		expected, ok := inventory.Files[name]
		if !ok || !digestPattern.MatchString(expected) || files[name] != expected {
			return fmt.Errorf("browsing installation hash mismatch: %s", name)
		}
	}
	manifests := map[string][]byte{}
	for _, pair := range []struct{ manifest, deck, kind string }{{names[1], names[0], "templates"}, {names[3], names[2], "reusable"}} {
		raw, e := browsingartifact.ReadManifest(filepath.Join(root, filepath.FromSlash(pair.manifest)))
		if e != nil {
			return e
		}

		if e = browsingartifact.ValidateFile(raw, pair.kind, files[pair.deck], filepath.Join(root, filepath.FromSlash(pair.deck))); e != nil {
			return e
		}
		manifests[pair.kind] = raw
	}
	if e = browsingartifact.ValidateInventory(inventory, manifests, nil); e != nil {
		return e
	}
	for _, name := range names {
		allowed[name] = true
	}
	allowed["browsing-manifest.json"] = true
	return nil
}
