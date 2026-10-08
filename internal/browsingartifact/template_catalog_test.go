package browsingartifact_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/browsingartifact"
	"github.com/buairtri/pptxgengo/internal/browsingfixture"
)

func TestTemplateCatalogInventoryClosesSourceAndTemplateDeck(t *testing.T) {
	root := t.TempDir()
	put := func(name string, data []byte) string {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		return hex.EncodeToString(sum[:])
	}
	deck := "browsing/template-library.pptx"
	deckHash := put(deck, browsingfixture.Deck(t, "templates"))
	coverageReport, _ := json.Marshal(map[string]any{"schema": "pptxgengo.native-template-coverage.v1", "editing_profile": "native-v1", "templates": 1, "native_list_boxes": 0, "native_card_shapes": 0, "native_tables": 0, "entries": []any{map[string]any{"template": "cards/3", "slide_id": "page-3", "native_list_boxes": 0, "native_card_shapes": 0, "native_tables": 0}}})
	coverageHash := put("browsing/native-editing-coverage.json", coverageReport)
	manifestRaw := browsingfixture.Manifest(t, "templates", deckHash)
	index := put("library/wm-design-system/v11/library.sqlite", []byte("sqlite fixture"))
	font := put("library/wm-design-system/v11/fonts/fixture.ttf", []byte("font fixture"))
	source := put("library/wm-design-system/v11/source/templates/catalog.json", []byte("{}"))
	firstSource := put("library/wm-design-system/v11/source/templates/fixture.json", []byte("template"))
	secondSource := put("library/wm-design-system/v11/source/frames/v0/frames.json", []byte("frames"))
	var m map[string]any
	if err := json.Unmarshal(manifestRaw, &m); err != nil {
		t.Fatal(err)
	}
	pins := []any{map[string]any{"path": "templates/fixture.json", "sha256": firstSource}, map[string]any{"path": "frames/v0/frames.json", "sha256": secondSource}}
	m["source_files"] = pins
	m["source_revision"] = "wmds-library.v11"
	m["source_commit"] = strings.Repeat("a", 40)
	m["editing_profile"] = "native-v1"
	m["media_policy"] = "placeholders"
	fonts := m["fonts"].([]any)
	fonts[0].(map[string]any)["sha256"] = font
	coverage := m["coverage"].(map[string]any)
	coverage["source_files"] = pins
	coverage["source_revision"] = "wmds-library.v11"
	coverage["source_commit"] = strings.Repeat("a", 40)
	entries := coverage["entries"].([]any)
	entries[0].(map[string]any)["source_sha256"] = firstSource
	entries[1].(map[string]any)["source_sha256"] = secondSource
	manifestRaw, _ = json.Marshal(m)
	manifest := put("browsing/template-library.manifest.json", manifestRaw)
	bundleJSON, _ := json.Marshal(map[string]any{"schema": "pptxgengo.wmds-bundle.v1", "files": []any{map[string]string{"path": "fonts/fixture.ttf", "sha256": font}}})
	bundle := put("library/wm-design-system/v11/bundle.json", bundleJSON)
	landing := put("library/wm-design-system/v11/catalog/index.json", []byte("{}"))
	designIndex := put("library/wm-design-system/v11/catalog/design-system/index.json", []byte("{}"))
	html := put("library/wm-design-system/v11/catalog/design-system.html", []byte("<html/>"))
	files := map[string]string{deck: deckHash, "browsing/native-editing-coverage.json": coverageHash, "browsing/template-library.manifest.json": manifest, "library/wm-design-system/v11/library.sqlite": index, "library/wm-design-system/v11/fonts/fixture.ttf": font, "library/wm-design-system/v11/source/templates/catalog.json": source, "library/wm-design-system/v11/source/templates/fixture.json": firstSource, "library/wm-design-system/v11/source/frames/v0/frames.json": secondSource, "library/wm-design-system/v11/bundle.json": bundle, "library/wm-design-system/v11/catalog/index.json": landing, "library/wm-design-system/v11/catalog/design-system/index.json": designIndex, "library/wm-design-system/v11/catalog/design-system.html": html}
	in := browsingartifact.TemplateCatalogInventory{Schema: "pptxgengo.template-catalog-release.v1", Files: files, Bundle: "v11", SourceRevision: "wmds-library.v11", SourceCommit: strings.Repeat("a", 40), BundleSHA256: strings.Repeat("a", 64), Compiler: "unit-test-only", ReleaseIdentity: "unit-test-only", AsOf: "2026-10-07", PipelineCreatedAt: "2026-10-07T12:00:00Z", EditingProfile: "native-v1", MediaPolicy: "placeholders"}
	raw, _ := json.Marshal(in)
	if err := os.WriteFile(filepath.Join(root, browsingartifact.TemplateCatalogInventoryName), raw, 0644); err != nil {
		t.Fatal(err)
	}
	parsed, err := browsingartifact.ReadTemplateCatalogInventory(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err = browsingartifact.ValidateTemplateCatalogInventory(root, parsed, files, "unit-test-only"); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "library/wm-design-system/v11/library.sqlite"), []byte("tampered SQLite"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = browsingartifact.ValidateTemplateCatalogInventory(root, parsed, files, "unit-test-only"); err == nil {
		t.Fatal("tampered SQLite accepted")
	}
	if err = os.WriteFile(filepath.Join(root, "library/wm-design-system/v11/library.sqlite"), []byte("sqlite fixture"), 0644); err != nil {
		t.Fatal(err)
	}
	missingFont := parsed
	missingFont.Files = map[string]string{}
	for name, hash := range parsed.Files {
		if name != "library/wm-design-system/v11/fonts/fixture.ttf" {
			missingFont.Files[name] = hash
		}
	}
	if err = browsingartifact.ValidateTemplateCatalogInventory(root, missingFont, missingFont.Files, "unit-test-only"); err == nil {
		t.Fatal("missing pinned font accepted")
	}
	withReusable := parsed
	withReusable.Files = map[string]string{}
	for name, hash := range parsed.Files {
		withReusable.Files[name] = hash
	}
	withReusable.Files["browsing/reusable-slides.pptx"] = deckHash
	reusableRaw, _ := json.Marshal(withReusable)
	if _, err = browsingartifact.ReadTemplateCatalogInventory(reusableRaw); err == nil {
		t.Fatal("reusable deck path accepted by templates-only inventory")
	}
	withTraversal := parsed
	withTraversal.Files = map[string]string{}
	for name, hash := range parsed.Files {
		withTraversal.Files[name] = hash
	}
	withTraversal.Files["library/wm-design-system/v11/source/../photos/private.jpg"] = deckHash
	traversalRaw, _ := json.Marshal(withTraversal)
	if _, err = browsingartifact.ReadTemplateCatalogInventory(traversalRaw); err == nil {
		t.Fatal("traversal/photo path accepted")
	}
	if err = os.WriteFile(filepath.Join(root, "library/wm-design-system/v11/library.sqlite-wal"), []byte("stale database sidecar"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = browsingartifact.ValidateTemplateCatalogInventory(root, parsed, files, "unit-test-only"); err == nil {
		t.Fatal("unlisted SQLite sidecar accepted")
	}
	if err = os.Remove(filepath.Join(root, "library/wm-design-system/v11/library.sqlite-wal")); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "browsing/reusable-slides.pptx"), browsingfixture.Deck(t, "reusable"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = browsingartifact.ValidateTemplateCatalogInventory(root, parsed, files, "unit-test-only"); err == nil {
		t.Fatal("stale reusable-slide deck accepted")
	}
}
