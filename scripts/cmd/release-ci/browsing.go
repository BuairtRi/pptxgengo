package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/browsingartifact"
)

func addBrowsingFiles(files map[string]Input, root string) error {
	for _, dir := range []string{root, filepath.Join(root, "browsing")} {
		info, e := os.Lstat(dir)
		if e != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("browsing resource directories must be real directories: %s", dir)
		}
	}
	var inventory struct {
		Schema string            `json:"schema"`
		Files  map[string]string `json:"files_sha256"`
	}
	raw, e := browsingartifact.ReadManifest(filepath.Join(root, "browsing-manifest.json"))
	if e != nil {
		return fmt.Errorf("installation archive requires generated private template-library.pptx and reusable-slides.pptx: %w", e)
	}
	if e = json.Unmarshal(raw, &inventory); e != nil {
		return e
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

			raw, e := browsingartifact.ReadManifest(path)
			if e != nil {
				return e
			}
			kind := "templates"
			deck := "browsing/template-library.pptx"
			if strings.Contains(rel, "reusable-slides") {
				kind = "reusable"
				deck = "browsing/reusable-slides.pptx"
			}

			if e = browsingartifact.ValidateFile(raw, kind, inventory.Files[deck], filepath.Join(root, filepath.FromSlash(deck))); e != nil {
				return e
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
