package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func addBrowsingFiles(files map[string]Input, root string) error {
	var inventory struct {
		Schema string            `json:"schema"`
		Files  map[string]string `json:"files_sha256"`
	}
	if e := readJSON(filepath.Join(root, "browsing-manifest.json"), &inventory); e != nil {
		return fmt.Errorf("installation archive requires generated private template-library.pptx and reusable-slides.pptx: %w", e)
	}
	required := []string{"browsing/template-library.pptx", "browsing/template-library.manifest.json", "browsing/reusable-slides.pptx", "browsing/reusable-slides.manifest.json"}
	if inventory.Schema != "pptxgengo.release-browsing-files.v1" || len(inventory.Files) != len(required) {
		return fmt.Errorf("browsing inventory incomplete or unsupported")
	}
	for _, rel := range required {
		expected, ok := inventory.Files[rel]
		if !ok || !browsingHash(expected) {
			return fmt.Errorf("browsing file hash missing: %s", rel)
		}
		path := filepath.Join(root, filepath.FromSlash(rel))
		actual, e := digest(path)
		if e != nil {
			return e
		}
		if actual != expected {
			return fmt.Errorf("browsing file drift: %s", rel)
		}
		if strings.HasSuffix(rel, ".manifest.json") {
			var m struct {
				Schema     string `json:"schema"`
				Kind       string `json:"kind"`
				DeckSHA256 string `json:"deck_sha256"`
				Slides     int    `json:"slides"`
			}
			raw, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			if e = json.Unmarshal(raw, &m); e != nil {
				return e
			}
			kind := "templates"
			deck := "browsing/template-library.pptx"
			if strings.Contains(rel, "reusable-slides") {
				kind = "reusable"
				deck = "browsing/reusable-slides.pptx"
			}
			if m.Schema != "pptxgengo.browsing-library.v1" || m.Kind != kind || m.Slides < 2 || m.DeckSHA256 != inventory.Files[deck] {
				return fmt.Errorf("browsing coverage manifest does not pin its generated deck: %s", rel)
			}
		}
		files[rel] = Input{Path: path}
	}
	files["browsing-manifest.json"] = Input{Path: filepath.Join(root, "browsing-manifest.json")}
	return nil
}

func browsingHash(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && strings.ToLower(s) == s
}
