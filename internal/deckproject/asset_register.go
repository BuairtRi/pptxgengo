package deckproject

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

type AssetRegistration struct {
	ID            string
	Data          []byte
	Description   string
	Focus         *AssetFocus
	ReplaceSHA256 string
}
type AssetRegistrationReceipt struct {
	Operation      string `json:"operation"`
	PreviousSHA256 string `json:"previous_sha256,omitempty"`
	ID             string `json:"id"`
	Path           string `json:"path"`
	SHA256         string `json:"sha256"`
	BeforeSHA256   string `json:"before_sha256"`
	AfterSHA256    string `json:"after_sha256"`
	Decision       string `json:"decision"`
	Policy         string `json:"policy"`
}

// RegisterAsset retains the exact approved/client original in the project. It
// does not authorize that image for publication or alter the global registry.
func RegisterAsset(p *Project, o AssetRegistration) (AssetRegistrationReceipt, error) {
	r := AssetRegistrationReceipt{Operation: "register-asset", ID: o.ID, BeforeSHA256: p.SourceHash(), Policy: "Original bytes retained. Delivery builds may derive JPEG media; layout-report.json records media_optimization hashes, dimensions and reasons. Focus is composition metadata; explicit slide fit/crop determines output."}
	if !stableID.MatchString(o.ID) || o.Description == "" {
		return r, fmt.Errorf("asset registration requires stable ID and description")
	}
	old, exists := p.Document.Assets[o.ID]
	guarded := map[string][]byte{}
	if o.ReplaceSHA256 != "" {
		if !exists || old.Path == "" || !shaPattern.MatchString(o.ReplaceSHA256) {
			return r, fmt.Errorf("asset revise requires an existing owned asset and exact predecessor SHA256")
		}
		previous, err := readProjectFile(p.Root, old.Path)
		if err != nil {
			return r, err
		}
		if digest(previous) != o.ReplaceSHA256 || (old.SHA256 != "" && old.SHA256 != o.ReplaceSHA256) {
			return r, fmt.Errorf("asset predecessor changed; preserve both revisions and resolve explicitly")
		}
		guarded[old.Path] = previous
		r.Operation = "revise-asset"
		r.PreviousSHA256 = o.ReplaceSHA256
	} else if exists {
		return r, fmt.Errorf("asset ID already exists: %s; use asset revise with predecessor hash", o.ID)
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
	_ = mime
	r.SHA256 = digest(o.Data)
	r.Path = "assets/objects/sha256/" + r.SHA256
	path, err := SafePath(p.Root, r.Path)
	if err != nil {
		return r, err
	}
	prior, err := readOptional(path)
	if err != nil {
		return r, err
	}
	created := prior == nil
	if created {
		if err = writeExclusive(path, o.Data, 0444); err != nil {
			return r, err
		}
	} else if digest(prior) != r.SHA256 {
		return r, fmt.Errorf("immutable asset object drift")
	}
	guarded[r.Path] = o.Data
	// A concurrent registration can adopt a deduplicated object before this
	// source transaction finishes. Retain unused immutable objects on failure;
	// deleting a newly created object could break the other author's source.
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
	candidate, err := commitSourceChangesChecked(p, changes, nil, guarded, func(next *Project) error { r.AfterSHA256 = next.SourceHash(); return writeJSON(decision, r) })
	if err != nil {
		os.Remove(decision)
		return r, err
	}
	_ = candidate
	return r, nil
}
