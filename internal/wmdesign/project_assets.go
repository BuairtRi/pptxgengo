package wmdesign

import (
	"crypto/sha256"
	"fmt"
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
	return primitiveAssetBytes(key)
}
