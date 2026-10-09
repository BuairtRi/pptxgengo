package browsingartifact

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/assetregistry"
	"os"
	"path/filepath"
	"strings"
)

// Only the source bundle's registered, hash-identical SVG diagram pieces are
// eligible. Photo previews and visually similar derivatives never qualify.
func validateRegisteredGraphics(root string, in TemplateCatalogInventory) error {
	prefix := "library/wm-design-system/" + in.Bundle + "/catalog/assets/"
	raw, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(prefix+"assets.json")))
	if e != nil {
		return e
	}
	var rows []struct {
		ID       string `json:"id"`
		Kind     string `json:"kind"`
		Variants []struct {
			ID           string `json:"id"`
			SHA          string `json:"sha256"`
			Path         string `json:"thumbnail_path"`
			ThumbnailSHA string `json:"thumbnail_sha256"`
			State        string `json:"thumbnail_state"`
			MIME         string `json:"thumbnail_mime"`
		} `json:"variants"`
	}
	if e = json.Unmarshal(raw, &rows); e != nil {
		return e
	}
	if len(rows) == 0 || len(rows) > 1000 {
		return fmt.Errorf("browsing.registered_graphics_invalid_count")
	}
	expected := map[string]bool{prefix + "assets.json": true}
	ids := map[string]bool{}
	trusted := assetregistry.Catalog()
	for _, row := range rows {
		if len(row.Variants) == 0 || len(row.Variants) > 1000 {
			return fmt.Errorf("browsing.registered_graphics_empty_or_oversize_variant_set")
		}
		for _, v := range row.Variants {
			ref, known := trusted[v.ID]
			allowed := row.Kind == "icon" && strings.HasPrefix(v.ID, "icon/") || row.Kind == "graphic" && strings.HasPrefix(v.ID, "arrow-")
			if !allowed || !known || ref.SHA256 != v.SHA || filepath.Ext(ref.Path) != ".svg" || ids[v.ID] || !hash(v.SHA) || v.SHA != v.ThumbnailSHA || v.State != "verified_registered_original" || v.MIME != "image/svg+xml" || !portable(v.Path) || !strings.HasPrefix(v.Path, "assets/") || filepath.Ext(v.Path) != ".svg" || in.Files[prefix+v.Path] != v.SHA {
				return fmt.Errorf("browsing.registered_graphics_unapproved_original: %s", v.ID)
			}
			ids[v.ID] = true
			expected[prefix+v.Path] = true
		}
	}
	for name := range in.Files {
		if strings.HasPrefix(name, prefix) && !expected[name] {
			return fmt.Errorf("browsing.registered_graphics_unregistered_file: %s", name)
		}
	}
	return nil
}
