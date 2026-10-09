package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AssetData is the immutable payload of a validated project-relative original.
// SHA256 is required; native packaging records later embedded derivatives.
type AssetData struct {
	Data   []byte
	SHA256 string
	MIME   string
}

func (r *renderer) primitiveAssetBytes(key string) ([]byte, primitiveAsset, error) {
	if a, ok := r.projectAssets[key]; ok {
		hash := fmt.Sprintf("%x", sha256.Sum256(a.Data))
		if len(a.Data) == 0 || a.SHA256 != hash {
			return nil, primitiveAsset{}, fmt.Errorf("scene.project_asset_hash_drift: %s", key)
		}
		path := key
		if a.MIME == "image/svg+xml" && !strings.HasSuffix(strings.ToLower(path), ".svg") {
			path += ".svg"
		}
		return a.Data, primitiveAsset{Path: path, SHA256: hash}, nil
	}
	data, a, e := primitiveAssetBytes(key)
	if e == nil || !errors.Is(e, os.ErrNotExist) || !bundledOriginalEligible(key) {
		return data, a, e
	}
	// V11 distributes hash-identical registered icon/arrow originals in its catalog.
	// Missing machine branding may fall back to that installed bundle. Denials,
	// drift, logos, photos and unknown/derived previews retain their original failure.
	if r.source == nil {
		return nil, a, e
	}
	bundleRoot := filepath.Dir(r.source.loadedFontRoot)
	if r.source.loadedFontRoot == "" {
		bundleRoot = filepath.Dir(r.source.Root)
	}
	bundled, found, err := bundledRegisteredIcon(bundleRoot, key, a.SHA256)
	if err != nil {
		return nil, a, err
	}
	if !found {
		return nil, a, e
	}
	return bundled, a, nil
}

// bundledOriginalEligible limits portable originals to registered icons and arrows.
func bundledOriginalEligible(key string) bool {
	_, registered := primitiveAssetRegistry[key]
	return registered && (strings.HasPrefix(key, "icon/") || strings.HasPrefix(key, "arrow-"))
}

// bundledRegisteredIcon accepts canonical registered IDs and exact original bytes,
// not a visually similar thumbnail or a path outside the catalog bundle.
func bundledRegisteredIcon(bundleRoot, key, hash string) ([]byte, bool, error) {
	if !bundledOriginalEligible(key) {
		return nil, false, nil
	}
	root := filepath.Join(bundleRoot, "catalog", "assets")
	raw, e := os.ReadFile(filepath.Join(root, "assets.json"))
	if os.IsNotExist(e) {
		return nil, false, nil
	}
	if e != nil {
		return nil, false, fmt.Errorf("scene.bundled_icon_catalog_read: %w", e)
	}
	var catalog []AssetSelection
	if e = json.Unmarshal(raw, &catalog); e != nil {
		return nil, false, fmt.Errorf("scene.bundled_icon_catalog_invalid: %w", e)
	}
	var match *AssetVariant
	for _, asset := range catalog {
		for _, v := range asset.Variants {
			if v.ID != key {
				continue
			}
			if match != nil {
				return nil, false, fmt.Errorf("scene.bundled_icon_ambiguous: %s", key)
			}
			copy := v
			match = &copy
		}
	}
	if match == nil {
		return nil, false, nil
	}
	if match.SHA256 != hash || match.ThumbnailState != "verified_registered_original" || match.ThumbnailSHA256 != hash || match.ThumbnailMIME != "image/svg+xml" {
		return nil, false, fmt.Errorf("scene.bundled_icon_provenance_mismatch: %s", key)
	}
	rel := filepath.Clean(match.ThumbnailPath)
	if rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || strings.Contains(match.ThumbnailPath, "\\") || filepath.VolumeName(rel) != "" {
		return nil, false, fmt.Errorf("scene.bundled_icon_unsafe_path: %s", key)
	}
	path := filepath.Join(root, rel)
	// Reject symlink escape even when it points to hash-identical content; a
	// portable release original must actually reside in its packaged catalog.
	resolved, e := filepath.EvalSymlinks(path)
	if e != nil {
		return nil, false, fmt.Errorf("scene.bundled_icon_read: %s: %w", key, e)
	}
	absRoot, e := filepath.Abs(root)
	if e != nil {
		return nil, false, e
	}
	resolvedRoot, e := filepath.EvalSymlinks(absRoot)
	if e != nil {
		return nil, false, e
	}
	within, e := filepath.Rel(resolvedRoot, resolved)
	if e != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return nil, false, fmt.Errorf("scene.bundled_icon_unsafe_path: %s", key)
	}
	data, e := os.ReadFile(resolved)
	if e != nil {
		return nil, false, fmt.Errorf("scene.bundled_icon_read: %s: %w", key, e)
	}
	if len(data) == 0 || len(data) > 8<<20 || fmt.Sprintf("%x", sha256.Sum256(data)) != hash {
		return nil, false, fmt.Errorf("scene.bundled_icon_hash_drift: %s", key)
	}
	return data, true, nil
}
