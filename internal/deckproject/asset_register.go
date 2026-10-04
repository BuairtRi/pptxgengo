package deckproject

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

type AssetRegistration struct {
	ID          string
	Data        []byte
	Description string
	Focus       *AssetFocus
}
type AssetRegistrationReceipt struct {
	Operation    string `json:"operation"`
	ID           string `json:"id"`
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	BeforeSHA256 string `json:"before_sha256"`
	AfterSHA256  string `json:"after_sha256"`
	Decision     string `json:"decision"`
	Policy       string `json:"policy"`
}

// RegisterAsset retains the exact approved/client original in the project. It
// does not authorize that image for publication or alter the global registry.
func RegisterAsset(p *Project, o AssetRegistration) (AssetRegistrationReceipt, error) {
	r := AssetRegistrationReceipt{Operation: "register-asset", ID: o.ID, BeforeSHA256: p.SourceHash(), Policy: "Original bytes retained. Focus is composition metadata; explicit slide fit/crop determines output."}
	if !stableID.MatchString(o.ID) || o.Description == "" {
		return r, fmt.Errorf("asset registration requires stable ID and description")
	}
	if _, exists := p.Document.Assets[o.ID]; exists {
		return r, fmt.Errorf("asset ID already exists: %s", o.ID)
	}
	if len(o.Data) > 64<<20 {
		return r, fmt.Errorf("asset exceeds 64MiB")
	}
	if o.Focus != nil && (math.IsNaN(o.Focus.X) || math.IsNaN(o.Focus.Y) || math.IsInf(o.Focus.X, 0) || math.IsInf(o.Focus.Y, 0) || o.Focus.X < 0 || o.Focus.X > 1 || o.Focus.Y < 0 || o.Focus.Y > 1) {
		return r, fmt.Errorf("asset focus x/y must be in [0,1]")
	}
	mime, err := assetMIME(o.Data)
	if err != nil {
		return r, err
	}
	ext := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/svg+xml": ".svg"}[mime]
	r.SHA256 = digest(o.Data)
	r.Path = "assets/originals/" + o.ID + "-" + r.SHA256[:16] + ext
	path, err := SafePath(p.Root, r.Path)
	if err != nil {
		return r, err
	}
	if err = writeExclusive(path, o.Data, 0444); err != nil {
		return r, err
	}
	committed := false
	defer func() {
		if !committed {
			os.Remove(path)
		}
	}()
	main, err := sourceYAML(p.Raw)
	if err != nil {
		return r, err
	}
	assets := map[string]Asset{}
	for key, value := range p.Document.Assets {
		assets[key] = value
	}
	assets[o.ID] = Asset{Path: r.Path, SHA256: r.SHA256, Description: o.Description, Focus: o.Focus}
	node, err := editYAMLNode(assets)
	if err != nil {
		return r, err
	}
	replaceMappingField(main.Content[0], "assets", node)
	raw, err := encodeSourceYAML(main)
	if err != nil {
		return r, err
	}
	changes := map[string][]byte{filepath.Base(p.SourcePath): raw}
	r.Decision = "decisions/asset-" + nonce() + ".json"
	decision, err := SafePath(p.Root, r.Decision)
	if err != nil {
		return r, err
	}
	candidate, err := commitSourceChanges(p, changes, func(next *Project) error { r.AfterSHA256 = next.SourceHash(); return writeJSON(decision, r) })
	if err != nil {
		os.Remove(decision)
		return r, err
	}
	_ = candidate
	committed = true
	return r, nil
}
