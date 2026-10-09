package main

import (
	"encoding/hex"
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
	raw, e := browsingartifact.ReadManifest(filepath.Join(root, "browsing-manifest.json"))
	if e != nil {
		return fmt.Errorf("installation archive requires generated private template-library.pptx and reusable-slides.pptx: %w", e)
	}
	inventory, e := browsingartifact.ReadInventory(raw)
	if e != nil {
		return e
	}
	required := []string{"browsing/template-library.pptx", "browsing/template-library.manifest.json", "browsing/reusable-slides.pptx", "browsing/reusable-slides.manifest.json"}
	manifests := map[string][]byte{}
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
			manifests[kind] = raw
		}
		files[rel] = Input{Path: path}
	}
	var expected *browsingartifact.InventoryExpectations
	if pipeline := os.Getenv("CI_PIPELINE_CREATED_AT"); pipeline != "" {
		asOf := pipeline
		if len(asOf) >= 10 {
			asOf = asOf[:10]
		}
		expected = &browsingartifact.InventoryExpectations{Inputs: browsingartifact.ReleaseInputs{BrandingArchiveSHA256: os.Getenv("WMDS_BRANDING_ARCHIVE_SHA256"), FinishedLibraryArchiveSHA256: os.Getenv("WMDS_FINISHED_LIBRARY_ARCHIVE_SHA256"), AsOf: asOf, PipelineCreatedAt: pipeline}, ReleaseIdentity: os.Getenv("CI_COMMIT_SHA")}
	}
	if e = browsingartifact.ValidateInventory(inventory, manifests, expected); e != nil {
		return e
	}
	files["browsing-manifest.json"] = Input{Path: filepath.Join(root, "browsing-manifest.json")}
	return nil
}

// addTemplateCatalogFiles stages a catalog-only resource closure. It carries
// the generated template deck plus the source bundle, font files, gallery and
// rebuilt SQLite index, without reusable content or private media originals.
func addTemplateCatalogFiles(files map[string]Input, root, version, commit string) error {
	if len(files) != 0 {
		return fmt.Errorf("template-only resources cannot be combined with other presentation payloads")
	}
	entries, e := os.ReadDir(root)
	if e != nil {
		return e
	}
	want := map[string]bool{"browsing": true, "library": true, browsingartifact.TemplateCatalogInventoryName: true, resourcePolicyName: true}
	for _, entry := range entries {
		if entry.Name() == "skills" || entry.Name() == "scripts" || entry.Name() == "SKILL-INSTALL.md" {
			want[entry.Name()] = true
		}
	}
	if len(entries) != len(want) {
		return fmt.Errorf("template-only resource directory has unexpected files")
	}
	for _, entry := range entries {
		if !want[entry.Name()] {
			return fmt.Errorf("template-only resource directory has unexpected file: %s", entry.Name())
		}
	}
	raw, e := os.ReadFile(filepath.Join(root, browsingartifact.TemplateCatalogInventoryName))
	if e != nil {
		return fmt.Errorf("template-only release requires a closed catalog inventory: %w", e)
	}
	in, e := browsingartifact.ReadTemplateCatalogInventory(raw)
	if e != nil {
		return e
	}
	if len(want) > 4 && !in.SkillIncluded {
		return fmt.Errorf("skill materials require explicit closed inventory")
	}
	if in.ReleaseIdentity != commit {
		return fmt.Errorf("template catalog release identity differs from release commit")
	}
	if pipeline := os.Getenv("CI_PIPELINE_CREATED_AT"); pipeline == "" || in.PipelineCreatedAt != pipeline {
		return fmt.Errorf("template catalog date differs from authoritative pipeline creation time")
	}
	for _, rel := range browsingartifact.TemplateCatalogFiles(in) {
		files[rel] = Input{Path: filepath.Join(root, filepath.FromSlash(rel))}
	}
	if e = browsingartifact.ValidateTemplateCatalogInventory(root, in, in.Files, commit); e != nil {
		return e
	}
	files[browsingartifact.TemplateCatalogInventoryName] = Input{Path: filepath.Join(root, browsingartifact.TemplateCatalogInventoryName)}
	// Bind the inventory to the exact source selected by the tagged release.
	selected, e := readSelectedReleaseBundle()
	if e != nil {
		return e
	}
	if strings.TrimSpace(string(selected)) != in.Bundle {
		return fmt.Errorf("template catalog bundle differs from tagged release selection")
	}
	return nil
}

func readSelectedReleaseBundle() ([]byte, error) {
	working, e := os.Getwd()
	if e != nil {
		return nil, e
	}
	for dir := working; ; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, "release", "default-bundle.txt")
		if raw, err := os.ReadFile(path); err == nil {
			return raw, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("release/default-bundle.txt not found from %s", working)
		}
	}
}

func browsingHash(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && strings.ToLower(s) == s
}
