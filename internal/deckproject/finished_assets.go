package deckproject

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Both receipts remain inspectable: original bytes and the explicitly remapped
// source identity. No image bytes, operation or parameters are adapted.
type FinishedAssetDerivation struct {
	SourceReceiptPath     string `json:"source_receipt_path"`
	SourceReceiptSHA256   string `json:"source_receipt_sha256"`
	OriginalReceiptPath   string `json:"original_receipt_path"`
	RemappedReceiptPath   string `json:"remapped_receipt_path"`
	RemappedReceiptSHA256 string `json:"remapped_receipt_sha256"`
}

func finishedAssetRegistry() map[string]wmdesign.PrimitiveAssetReference {
	registry := map[string]wmdesign.PrimitiveAssetReference{}
	for _, a := range wmdesign.PrimitiveAssetCatalog() {
		registry[a.Key] = a
	}
	return registry
}

// Only actual declared media slots and their ancestry are dependencies. Return
// parents first; reject cycles, missing ancestors and excessively large graphs.
func finishedAssetOrder(media map[string]string, assets map[string]Asset) ([]string, error) {
	registry := finishedAssetRegistry()
	visiting := map[string]int{}
	order := []string{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] == 1 {
			return fmt.Errorf("finished-slide.asset_ancestry_cycle: %s", id)
		}
		if visiting[id] == 2 {
			return nil
		}
		a, ok := assets[id]
		if !ok || !stableID.MatchString(id) {
			return fmt.Errorf("finished-slide.asset_missing: %s", id)
		}
		if len(visiting) >= 256 {
			return fmt.Errorf("finished-slide.asset_ancestry_limit")
		}
		if (a.Path == "") == (a.RegistryID == "") || (a.DerivedFrom == "") != (a.DerivationReceipt == "") {
			return fmt.Errorf("finished-slide.asset_dependency_invalid: %s", id)
		}
		visiting[id] = 1
		if a.DerivedFrom != "" {
			if err := visit(a.DerivedFrom); err != nil {
				return err
			}
		}
		visiting[id] = 2
		order = append(order, id)
		return nil
	}
	names := []string{}
	for name := range media {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		value := media[name]
		if _, ok := registry[value]; ok {
			continue
		}
		if err := visit(strings.TrimPrefix(value, "project:")); err != nil {
			return nil, err
		}
	}
	return order, nil
}

func finishedDerivation(raw []byte, parent string, parentSHA, resultSHA string) (DerivationReceipt, error) {
	var r DerivationReceipt
	if len(raw) == 0 || len(raw) > 1<<20 {
		return r, fmt.Errorf("finished-slide.derivation_receipt_size_invalid")
	}
	if err := strictInto(json.RawMessage(raw), &r); err != nil {
		return r, err
	}
	if r.Schema != "pptxgengo.asset-derivation.v1" || r.SourceAsset != parent ||
		r.SourceSHA256 != parentSHA || r.ResultSHA256 != resultSHA || strings.TrimSpace(r.Operation) == "" {
		return r, fmt.Errorf("finished-slide.derivation_receipt_mismatch")
	}
	return r, nil
}

func closeFinishedAssets(p *Project, media map[string]string, files map[string][]byte, roles map[string]string) (map[string]Asset, error) {
	order, err := finishedAssetOrder(media, p.Document.Assets)
	if err != nil {
		return nil, err
	}
	registry := finishedAssetRegistry()
	closed := map[string]Asset{}
	for _, id := range order {
		a := p.Document.Assets[id]
		if a.RegistryID != "" {
			r, ok := registry[a.RegistryID]
			if !ok || (a.SHA256 != "" && a.SHA256 != r.SHA256) {
				return nil, fmt.Errorf("finished-slide.registry_asset_pin_mismatch: %s", id)
			}
			a.SHA256 = r.SHA256
		} else {
			raw, err := projectDependency(p, a.Path, 64<<20)
			if err != nil {
				return nil, err
			}
			if a.SHA256 != "" && a.SHA256 != digest(raw) {
				return nil, fmt.Errorf("finished-slide.asset_pin_mismatch: %s", id)
			}
			mime, err := assetMIME(raw)
			if err != nil {
				return nil, err
			}
			ext := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/svg+xml": ".svg"}[mime]
			a.Path = "assets/" + digest([]byte(id))[:16] + "-" + digest(raw)[:16] + ext
			a.SHA256 = digest(raw)
			files[a.Path], roles[a.Path] = raw, "asset"
		}
		if a.DerivedFrom != "" {
			raw, err := projectDependency(p, a.DerivationReceipt, 1<<20)
			if err != nil {
				return nil, err
			}
			if _, err := finishedDerivation(raw, a.DerivedFrom, closed[a.DerivedFrom].SHA256, a.SHA256); err != nil {
				return nil, err
			}
			a.DerivationReceipt = "derivations/" + digest([]byte(id))[:16] + "-" + digest(raw)[:16] + ".json"
			files[a.DerivationReceipt], roles[a.DerivationReceipt] = raw, "documentation"
		}
		closed[id] = a
	}
	return closed, nil
}

func insertFinishedAssets(media map[string]string, dependencies map[string]Asset, payload map[string][]byte, roles map[string]string, assets map[string]Asset, files map[string][]byte, lineage *LibraryLineage) error {
	order, err := finishedAssetOrder(media, dependencies)
	if err != nil {
		return err
	}
	if len(order) != len(dependencies) {
		return fmt.Errorf("finished-slide.unreferenced_asset_dependency")
	}
	registry := finishedAssetRegistry()
	for _, id := range order {
		a := dependencies[id]
		if !shaPattern.MatchString(a.SHA256) {
			return fmt.Errorf("finished-slide.asset_dependency_invalid: %s", id)
		}
		newID := "reuse-" + nonce()
		if _, ok := assets[newID]; ok {
			return fmt.Errorf("finished-slide.asset_id_collision")
		}
		lineage.AssetRemaps[id] = newID
		if a.RegistryID != "" {
			r, ok := registry[a.RegistryID]
			if !ok || a.SHA256 != r.SHA256 {
				return fmt.Errorf("finished-slide.registry_asset_pin_mismatch: %s", id)
			}
		} else {
			raw, ok := payload[a.Path]
			if !ok || digest(raw) != a.SHA256 || roles[a.Path] != "asset" {
				return fmt.Errorf("finished-slide.asset_bytes_missing: %s", id)
			}
			mime, err := assetMIME(raw)
			if err != nil {
				return err
			}
			ext := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/svg+xml": ".svg"}[mime]
			a.Path = "assets/originals/" + newID + "-" + a.SHA256[:16] + ext
			files[a.Path] = raw
		}
		if a.DerivedFrom != "" {
			original, ok := payload[a.DerivationReceipt]
			if !ok || roles[a.DerivationReceipt] != "documentation" {
				return fmt.Errorf("finished-slide.derivation_receipt_missing: %s", id)
			}
			receipt, err := finishedDerivation(original, a.DerivedFrom, dependencies[a.DerivedFrom].SHA256, a.SHA256)
			if err != nil {
				return err
			}
			// Preserve the source receipt verbatim and change only the parent's identity
			// in the receipt that the destination compiler uses.
			sourcePath := a.DerivationReceipt
			originalPath := "assets/derivations/" + newID + "-source.json"
			remappedPath := "assets/derivations/" + newID + ".json"
			a.DerivedFrom = lineage.AssetRemaps[a.DerivedFrom]
			receipt.SourceAsset = a.DerivedFrom
			remapped := append(canonical(receipt), '\n')
			files[originalPath], files[remappedPath] = original, remapped
			a.DerivationReceipt = remappedPath
			if lineage.Derivations == nil {
				lineage.Derivations = map[string]FinishedAssetDerivation{}
			}
			lineage.Derivations[id] = FinishedAssetDerivation{SourceReceiptPath: sourcePath, SourceReceiptSHA256: digest(original), OriginalReceiptPath: originalPath, RemappedReceiptPath: remappedPath, RemappedReceiptSHA256: digest(remapped)}
		}
		assets[newID] = a
	}
	return nil
}
