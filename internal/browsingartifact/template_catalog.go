package browsingartifact

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// TemplateCatalogInventory closes the template-only release payload. The
// inventory itself is covered by the signed platform archive; Files closes the
// browsing deck and authoring source tree beneath it.
type TemplateCatalogInventory struct {
	SkillIncluded      bool              `json:"skill_included,omitempty"`
	RegisteredGraphics bool              `json:"registered_graphics,omitempty"`
	Schema             string            `json:"schema"`
	Files              map[string]string `json:"files_sha256"`
	Bundle             string            `json:"bundle"`
	SourceRevision     string            `json:"source_revision"`
	SourceCommit       string            `json:"source_commit"`
	BundleSHA256       string            `json:"bundle_sha256"`
	Compiler           string            `json:"compiler"`
	ReleaseIdentity    string            `json:"release_identity"`
	AsOf               string            `json:"as_of"`
	PipelineCreatedAt  string            `json:"pipeline_created_at"`
	EditingProfile     string            `json:"editing_profile"`
	MediaPolicy        string            `json:"media_policy"`
}

const TemplateCatalogInventoryName = "template-catalog-manifest.json"

func ReadTemplateCatalogInventory(raw []byte) (TemplateCatalogInventory, error) {
	var in TemplateCatalogInventory
	if len(raw) > MaxManifestBytes {
		return in, fmt.Errorf("browsing.template_catalog_inventory_too_large")
	}
	if e := json.Unmarshal(raw, &in); e != nil {
		return in, e
	}
	if in.Schema != "pptxgengo.template-catalog-release.v1" || in.Bundle == "" || in.SourceRevision == "" || in.SourceCommit == "" || !hash(in.BundleSHA256) || in.Compiler == "" || in.ReleaseIdentity == "" || in.EditingProfile != "native-v1" || in.MediaPolicy != "placeholders" || len(in.Files) < 4 || len(in.Files) > 200000 {
		return in, fmt.Errorf("browsing.template_catalog_inventory_invalid")
	}
	if len(in.Bundle) < 2 || in.Bundle[0] != 'v' || in.Bundle[1] < '1' || in.Bundle[1] > '9' {
		return in, fmt.Errorf("browsing.template_catalog_bundle_invalid")
	}
	for _, c := range in.Bundle[2:] {
		if c < '0' || c > '9' {
			return in, fmt.Errorf("browsing.template_catalog_bundle_invalid")
		}
	}
	if in.SourceRevision != "wmds-library."+in.Bundle {
		return in, fmt.Errorf("browsing.template_catalog_source_revision_mismatch")
	}
	pipeline, e := time.Parse(time.RFC3339Nano, in.PipelineCreatedAt)
	if e != nil || pipeline.UTC().Format(time.DateOnly) != in.AsOf {
		return in, fmt.Errorf("browsing.template_catalog_release_date_invalid")
	}
	if _, e = time.Parse(time.DateOnly, in.AsOf); e != nil {
		return in, fmt.Errorf("browsing.template_catalog_release_date_invalid")
	}
	for name, sum := range in.Files {
		bundleRoot := "library/wm-design-system/" + in.Bundle + "/"
		if !portable(name) || (!strings.HasPrefix(name, "browsing/") && !strings.HasPrefix(name, bundleRoot) && !(in.SkillIncluded && (strings.HasPrefix(name, "skills/west-monroe-presentations/") || name == "scripts/install-skill.py" || name == "SKILL-INSTALL.md"))) || !hash(sum) {
			return in, fmt.Errorf("browsing.template_catalog_path_or_hash_invalid: %s", name)
		}
		lower := strings.ToLower(name)
		for _, private := range []string{"/photos/", "/branding/", "/reusable"} {
			if strings.Contains(lower, private) {
				return in, fmt.Errorf("browsing.template_catalog_private_asset_path: %s", name)
			}
		}
		if strings.Contains(lower, "/catalog/assets/") {
			if !in.RegisteredGraphics || !strings.HasPrefix(name, bundleRoot+"catalog/assets/") || (filepath.Ext(name) != ".svg" && name != bundleRoot+"catalog/assets/assets.json") {
				return in, fmt.Errorf("browsing.template_catalog_unapproved_catalog_asset: %s", name)
			}
			continue
		}
		if strings.Contains(lower, "/assets/") {
			base := filepath.Base(name)
			if !strings.HasPrefix(name, bundleRoot+"assets/") || (base != "wm_h_pos_clr_rgb_august2024.png" && base != "wm_h_pos_clr_rgb_august2024.svg" && base != "wm_h_rev_wht_rgb_august2024.png" && base != "wm_h_rev_wht_rgb_august2024.svg") {
				return in, fmt.Errorf("browsing.template_catalog_unapproved_mark: %s", name)
			}
		}
	}
	return in, nil
}

// ValidateTemplateCatalogInventory requires the file tree under browsing/ and
// library/ to match the signed inventory exactly and validates the generated
// template deck against its own provenance manifest.
func ValidateTemplateCatalogInventory(root string, in TemplateCatalogInventory, archiveFiles map[string]string, expectedRelease string) error {
	if expectedRelease != "" && in.ReleaseIdentity != expectedRelease {
		return fmt.Errorf("browsing.template_catalog_release_identity_mismatch")
	}
	if _, ok := in.Files["browsing/template-library.pptx"]; !ok {
		return fmt.Errorf("browsing.template_catalog_template_deck_missing")
	}
	if _, ok := in.Files["browsing/template-library.manifest.json"]; !ok {
		return fmt.Errorf("browsing.template_catalog_template_manifest_missing")
	}
	bundleRoot := "library/wm-design-system/" + in.Bundle + "/"
	for _, name := range []string{bundleRoot + "bundle.json", bundleRoot + "library.sqlite", bundleRoot + "catalog/index.json", bundleRoot + "catalog/design-system/index.json", bundleRoot + "catalog/design-system.html"} {
		if _, ok := in.Files[name]; !ok {
			return fmt.Errorf("browsing.template_catalog_required_file_missing: %s", name)
		}
	}
	if _, ok := in.Files["browsing/native-editing-coverage.json"]; !ok {
		return fmt.Errorf("browsing.template_catalog_native_coverage_missing")
	}
	if in.SkillIncluded {
		for _, name := range []string{"skills/west-monroe-presentations/SKILL.md", "skills/west-monroe-presentations/references/installation.md", "scripts/install-skill.py", "SKILL-INSTALL.md"} {
			if _, ok := in.Files[name]; !ok {
				return fmt.Errorf("browsing.template_catalog_skill_material_missing: %s", name)
			}
		}
	}
	if in.RegisteredGraphics {
		if e := validateRegisteredGraphics(root, in); e != nil {
			return e
		}
	}
	seen := map[string]bool{}
	for rel, want := range in.Files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		st, e := os.Lstat(path)
		if e != nil || !st.Mode().IsRegular() {
			return fmt.Errorf("browsing.template_catalog_file_missing_or_unsafe: %s", rel)
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		sum := sha256.Sum256(raw)
		got := hex.EncodeToString(sum[:])
		if got != want || archiveFiles[rel] != want {
			return fmt.Errorf("browsing.template_catalog_hash_mismatch: %s", rel)
		}
		seen[rel] = true
	}
	subtrees := []string{"browsing", "library"}
	if in.SkillIncluded {
		subtrees = append(subtrees, "skills", "scripts")
	}
	for _, subtree := range subtrees {
		base := filepath.Join(root, subtree)
		st, e := os.Lstat(base)
		if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("browsing.template_catalog_tree_not_directory: %s", subtree)
		}
		e = filepath.WalkDir(base, func(p string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if p == base {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("browsing.template_catalog_symlink: %s", p)
			}
			if d.IsDir() {
				return nil
			}
			rel, e := filepath.Rel(root, p)
			if e != nil {
				return e
			}
			rel = filepath.ToSlash(rel)
			if !seen[rel] {
				return fmt.Errorf("browsing.template_catalog_unlisted_file: %s", rel)
			}
			return nil
		})
		if e != nil {
			return e
		}
	}
	deck := "browsing/template-library.pptx"
	manifestPath := filepath.Join(root, "browsing/template-library.manifest.json")
	raw, e := ReadManifest(manifestPath)
	if e != nil {
		return e
	}
	if e = ValidateFile(raw, "templates", in.Files[deck], filepath.Join(root, filepath.FromSlash(deck))); e != nil {
		return e
	}
	var m Manifest
	if e = json.Unmarshal(raw, &m); e != nil {
		return e
	}
	coverageRaw, e := os.ReadFile(filepath.Join(root, "browsing/native-editing-coverage.json"))
	if e != nil {
		return e
	}
	if m.BundleSHA256 != in.BundleSHA256 || m.SourceRevision != in.SourceRevision || m.SourceCommit != in.SourceCommit || m.Compiler != in.Compiler || m.ReleaseIdentity != in.ReleaseIdentity || m.AsOf != in.AsOf || m.EditingProfile != in.EditingProfile || m.MediaPolicy != in.MediaPolicy {
		return fmt.Errorf("browsing.template_catalog_manifest_provenance_mismatch")
	}
	var deckCoverage TemplateCoverage
	if e = json.Unmarshal(m.Coverage, &deckCoverage); e != nil {
		return e
	}
	if e = validateNativeTemplateCoverage(coverageRaw, in.EditingProfile, deckCoverage); e != nil {
		return e
	}
	var bundle struct {
		Schema string    `json:"schema"`
		Files  []FilePin `json:"files"`
	}
	bundleRaw, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(bundleRoot+"bundle.json")))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(bundleRaw, &bundle); e != nil {
		return e
	}
	if bundle.Schema != "pptxgengo.wmds-bundle.v1" || len(bundle.Files) == 0 {
		return fmt.Errorf("browsing.template_catalog_bundle_manifest_invalid")
	}
	for _, pin := range bundle.Files {
		name := bundleRoot + pin.Path
		if in.Files[name] != pin.SHA256 {
			return fmt.Errorf("browsing.template_catalog_bundle_pin_missing_or_changed: %s", pin.Path)
		}
	}
	for _, pin := range m.SourceFiles {
		name := bundleRoot + "source/" + pin.Path
		if in.Files[name] != pin.SHA256 {
			return fmt.Errorf("browsing.template_catalog_source_pin_missing_or_changed: %s", pin.Path)
		}
	}
	for _, pin := range m.Assets {
		name := bundleRoot + pin.Path
		if in.Files[name] != pin.SHA256 {
			return fmt.Errorf("browsing.template_catalog_asset_pin_missing_or_changed: %s", pin.Path)
		}
	}
	for _, font := range m.Fonts {
		name := bundleRoot + "fonts/" + filepath.Base(font.File)
		if in.Files[name] != font.SHA256 {
			return fmt.Errorf("browsing.template_catalog_font_pin_missing_or_changed: %s", font.File)
		}
	}
	return nil
}

func validateNativeTemplateCoverage(raw []byte, editingProfile string, deck TemplateCoverage) error {
	var native struct {
		Schema           string `json:"schema"`
		EditingProfile   string `json:"editing_profile"`
		Templates        int    `json:"templates"`
		NativeListBoxes  int    `json:"native_list_boxes"`
		NativeCardShapes int    `json:"native_card_shapes"`
		NativeTables     int    `json:"native_tables"`
		Entries          []struct {
			Template         string `json:"template"`
			SlideID          string `json:"slide_id"`
			NativeListBoxes  int    `json:"native_list_boxes"`
			NativeCardShapes int    `json:"native_card_shapes"`
			NativeTables     int    `json:"native_tables"`
		} `json:"entries"`
	}
	if e := json.Unmarshal(raw, &native); e != nil {
		return e
	}
	if native.Schema != "pptxgengo.native-template-coverage.v1" || native.EditingProfile != editingProfile || native.Templates < 1 || native.Templates != deck.ExpectedTemplates || len(native.Entries) != native.Templates {
		return fmt.Errorf("browsing.template_catalog_native_coverage_invalid")
	}
	expected := map[string]string{}
	for _, entry := range deck.Entries {
		if entry.Kind != "template" {
			continue
		}
		if entry.SlideID == "" || entry.Key == "" || expected[entry.SlideID] != "" {
			return fmt.Errorf("browsing.template_catalog_deck_coverage_invalid")
		}
		expected[entry.SlideID] = entry.Key
	}
	if len(expected) != native.Templates {
		return fmt.Errorf("browsing.template_catalog_deck_coverage_count_mismatch")
	}
	seen := map[string]bool{}
	listBoxes, cardShapes, tables := 0, 0, 0
	for _, entry := range native.Entries {
		if entry.Template == "" || entry.SlideID == "" || seen[entry.SlideID] || expected[entry.SlideID] != entry.Template || entry.NativeListBoxes < 0 || entry.NativeCardShapes < 0 || entry.NativeTables < 0 {
			return fmt.Errorf("browsing.template_catalog_native_coverage_entry_invalid")
		}
		seen[entry.SlideID] = true
		listBoxes += entry.NativeListBoxes
		cardShapes += entry.NativeCardShapes
		tables += entry.NativeTables
	}
	if native.NativeListBoxes < 0 || native.NativeCardShapes < 0 || native.NativeTables < 0 || listBoxes != native.NativeListBoxes || cardShapes != native.NativeCardShapes || tables != native.NativeTables {
		return fmt.Errorf("browsing.template_catalog_native_coverage_totals_invalid")
	}
	return nil
}

// TemplateCatalogFiles returns the ordered closed file list for callers that
// need deterministic reporting.
func TemplateCatalogFiles(in TemplateCatalogInventory) []string {
	keys := make([]string, 0, len(in.Files))
	for key := range in.Files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
