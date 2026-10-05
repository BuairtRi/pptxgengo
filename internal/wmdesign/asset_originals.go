package wmdesign

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func primitiveAssetOriginalPath(relative string) (string, error) {
	root := os.Getenv("WMDS_BRANDING_ROOT")
	if root == "" && os.Getenv("PPTXGENGO_RELEASE_ROOT") != "" {
		root = filepath.Join(os.Getenv("PPTXGENGO_RELEASE_ROOT"), "branding")
	}
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, "Documents/branding")
	}
	if filepath.IsAbs(relative) {
		return relative, nil
	}
	return filepath.Join(root, filepath.Clean(relative)), nil
}

// PrimitiveAssetOriginal verifies one selected original with a streaming hash.
// Discovery lists registered facts; preview, rendering and packaging validate
// the bytes when the caller actually chooses the asset.
func PrimitiveAssetOriginal(key string) (PrimitiveAssetReference, string, error) {
	asset, ok := primitiveAssetRegistry[key]
	if !ok {
		return PrimitiveAssetReference{}, "", fmt.Errorf("scene.unregistered_asset: %s", key)
	}
	path, err := primitiveAssetOriginalPath(asset.Path)
	if err != nil {
		return PrimitiveAssetReference{}, "", err
	}
	file, err := os.Open(path)
	if err != nil {
		return PrimitiveAssetReference{}, "", fmt.Errorf("scene.asset_read: %s: %w", key, err)
	}
	digest := sha256.New()
	_, readErr := io.Copy(digest, file)
	closeErr := file.Close()
	if readErr != nil {
		return PrimitiveAssetReference{}, "", readErr
	}
	if closeErr != nil {
		return PrimitiveAssetReference{}, "", closeErr
	}
	if fmt.Sprintf("%x", digest.Sum(nil)) != asset.SHA256 {
		return PrimitiveAssetReference{}, "", fmt.Errorf("scene.asset_hash_drift: %s", key)
	}
	return PrimitiveAssetReference{Key: key, Path: asset.Path, SHA256: asset.SHA256, Crop: asset.Crop}, path, nil
}
