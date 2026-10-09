package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type ScaffoldPlaceholder struct {
	RegistryID  string `json:"registry_id"`
	AssetID     string `json:"asset_id"`
	Path        string `json:"path"`
	SHA256      string `json:"sha256"`
	MIME        string `json:"mime"`
	Description string `json:"description"`
}

func validateScaffoldPlaceholderAssets(assets map[string]Asset) error {
	registry := map[string]wmdesign.PrimitiveAssetReference{}
	for _, asset := range wmdesign.PrimitiveAssetCatalog() {
		registry[asset.Key] = asset
	}
	aliases := map[string]string{}
	for id, asset := range assets {
		if asset.PlaceholderFor == "" {
			continue
		}
		ref, exists := registry[asset.PlaceholderFor]
		if !exists || strings.HasPrefix(asset.PlaceholderFor, "icon/") || strings.HasPrefix(asset.PlaceholderFor, "arrow-") || asset.RegistryID != "" || !shaPattern.MatchString(asset.SHA256) || asset.SHA256 == ref.SHA256 || asset.Path != "assets/objects/sha256/"+asset.SHA256 || !strings.HasPrefix(asset.Description, "Schematic placeholder for ") || asset.DerivedFrom != "" || asset.DerivationReceipt != "" {
			return fmt.Errorf("asset %s: placeholder_for requires registered non-icon/non-arrow media, content-addressed private path, exact differing SHA and explicit schematic description; it never claims original artwork", id)
		}
		if other, exists := aliases[asset.PlaceholderFor]; exists {
			return fmt.Errorf("asset %s: competing placeholder aliases %s and %s", id, other, id)
		}
		aliases[asset.PlaceholderFor] = id
		for otherID, other := range assets {
			if otherID != id && other.RegistryID == asset.PlaceholderFor {
				return fmt.Errorf("asset %s: placeholder alias competes with declared original asset %s", id, otherID)
			}
		}
	}
	return nil
}

// Generated native picture provenance records the exact renderer-selected key
// and payload SHA, including implicit inline marks. Do not search/rewrite copy.
func addScaffoldPlaceholderPayloads(out *TemplateScaffold, deck []byte, placeholders map[string]wmdesign.AssetData, omitted map[string]bool) error {
	archive, e := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if e != nil {
		return e
	}
	used := map[string]bool{}
	for _, file := range archive.File {
		if !strings.HasPrefix(file.Name, "ppt/slides/slide") || !strings.HasSuffix(file.Name, ".xml") {
			continue
		}
		if file.UncompressedSize64 > 16<<20 {
			return fmt.Errorf("scaffold native slide XML exceeds limit")
		}
		input, e := file.Open()
		if e != nil {
			return e
		}
		decoder := xml.NewDecoder(io.LimitReader(input, 16<<20))
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				input.Close()
				return err
			}
			node, ok := token.(xml.StartElement)
			if !ok || node.Name.Local != "cNvPr" {
				continue
			}
			name, description := "", ""
			for _, attr := range node.Attr {
				if attr.Name.Local == "name" {
					name = attr.Value
				}
				if attr.Name.Local == "descr" {
					description = attr.Value
				}
			}
			exclude := false
			for id := range omitted {
				if name == id || strings.HasPrefix(name, id+".") {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
			for key, asset := range placeholders {
				if strings.Contains(description, key+"; canonical SHA256="+asset.SHA256) {
					used[key] = true
				}
			}
		}
		if e = input.Close(); e != nil {
			return e
		}
	}
	keys := []string{}
	for key := range used {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		data := placeholders[key]
		id := "placeholder-" + strings.ReplaceAll(key, "/", "-")
		if !stableID.MatchString(id) {
			return fmt.Errorf("placeholder ID is not portable: %s", id)
		}
		path := "assets/objects/sha256/" + data.SHA256
		description := "Schematic placeholder for " + key + "; not original artwork. Replace/review before delivery."
		if out.Assets == nil {
			out.Assets = map[string]Asset{}
			out.AssetPayloads = map[string][]byte{}
		}
		out.Assets[id] = Asset{Path: path, SHA256: data.SHA256, Description: description, PlaceholderFor: key}
		out.AssetPayloads[path] = append([]byte(nil), data.Data...)
		out.PlaceholderMedia = append(out.PlaceholderMedia, ScaffoldPlaceholder{RegistryID: key, AssetID: id, Path: path, SHA256: data.SHA256, MIME: data.MIME, Description: description})
	}
	return validateScaffoldPlaceholderAssets(out.Assets)
}
