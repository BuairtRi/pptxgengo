package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func withSyntheticAssetRegistry(t *testing.T, entries map[string]primitiveAsset, contents map[string][]byte) {
	t.Helper()
	oldRegistry := primitiveAssetRegistry
	primitiveAssetRegistry = entries
	oldRoot := os.Getenv("WMDS_BRANDING_ROOT")
	root := t.TempDir()
	if err := os.Setenv("WMDS_BRANDING_ROOT", root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		primitiveAssetRegistry = oldRegistry
		_ = os.Setenv("WMDS_BRANDING_ROOT", oldRoot)
	})
	for relative, data := range contents {
		path := filepath.Join(root, filepath.Clean(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func syntheticSVG(path string) ([]byte, primitiveAsset) {
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"><circle cx="6" cy="6" r="4"/></svg>`)
	digest := sha256.Sum256(data)
	return data, primitiveAsset{Path: path, SHA256: fmt.Sprintf("%x", digest)}
}

func TestAssetSelectionsGroupIconVariantsAndPreserveRegisteredIDs(t *testing.T) {
	entries := map[string]primitiveAsset{}
	contents := map[string][]byte{}
	for _, color := range []string{"magenta", "navy", "white"} {
		key := "icon/risk-alert-arrow/" + color
		data, entry := syntheticSVG("icons/risk-alert-arrow-" + color + ".svg")
		entries[key] = entry
		contents[entry.Path] = data
	}
	withSyntheticAssetRegistry(t, entries, contents)
	results, err := AssetSelections("risk", "icon", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("grouped result count %d; want one", len(results))
	}
	if results[0].ID != "icon/risk-alert-arrow" || len(results[0].Variants) != 3 {
		t.Fatalf("group did not preserve the three color variants: %+v", results[0])
	}
	ids := []string{}
	for _, variant := range results[0].Variants {
		ids = append(ids, variant.ID)
		if variant.ThumbnailPath == "" || variant.ThumbnailState != "verified_registered_original" || variant.ThumbnailMIME != "image/svg+xml" {
			t.Errorf("missing verified browser-viewable source: %+v", variant)
		}
	}
	if want := []string{"icon/risk-alert-arrow/magenta", "icon/risk-alert-arrow/navy", "icon/risk-alert-arrow/white"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("variant ids %v; want %v", ids, want)
	}
	if got := surfaceGuidance("white"); !reflect.DeepEqual(got, []string{"dark surface"}) {
		t.Fatalf("white icon guidance %v", got)
	}
}

func TestAssetSelectionsPhotoQueryUsesCuratedRelevantTagsAndNoFallback(t *testing.T) {
	const key = "photo-working-session"
	imageBytes := image.NewRGBA(image.Rect(0, 0, 24, 16))
	imageBytes.Set(2, 2, color.RGBA{R: 255, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, imageBytes); err != nil {
		t.Fatal(err)
	}
	contents := encoded.Bytes()
	digest := sha256.Sum256(contents)
	entry := primitiveAsset{Path: "West Monroe Photos/working-session.png", SHA256: fmt.Sprintf("%x", digest)}
	withSyntheticAssetRegistry(t, map[string]primitiveAsset{key: entry}, map[string][]byte{entry.Path: contents})
	results, err := AssetSelections("people collaborating in a workshop", "photo", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ID != key || results[0].People != "yes" || results[0].Orientation != "landscape" {
		t.Fatalf("photo query returned unrelated or incorrect result: %+v", results)
	}
	missing, err := AssetSelections("people workshop", "photo", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 1 {
		t.Fatalf("expected one honestly tagged working session photo, got %+v", missing)
	}
	noMatch, err := AssetSelections("people underwater workshop", "photo", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(noMatch) != 1 || noMatch[0].ID != key {
		t.Fatalf("OR query should retain relevant partial match and rank it: %+v", noMatch)
	}
	noRelevant, err := AssetSelections("underwater", "photo", 0)
	if err != nil || len(noRelevant) != 0 {
		t.Fatalf("query with no matching terms returned fallback assets: %v %+v", err, noRelevant)
	}
}

func TestAssetSelectionsValidatesKindAndLimit(t *testing.T) {
	if _, err := AssetSelections("", "unknown", 1); err == nil {
		t.Fatal("accepted unknown kind")
	}
	if _, err := AssetSelections("", "icon", -1); err == nil {
		t.Fatal("accepted negative limit")
	}
}

func TestCuratedPhotoRegistryCoverageAndMetadata(t *testing.T) {
	iconConcepts, iconInstances, photos := map[string]bool{}, 0, 0
	for _, asset := range PrimitiveAssetCatalog() {
		if strings.HasPrefix(asset.Key, "icon/") {
			iconInstances++
			parts := strings.Split(asset.Key, "/")
			iconConcepts["icon/"+parts[1]] = true
		}
		if strings.HasPrefix(asset.Key, "photo-") {
			photos++
		}
	}
	if len(iconConcepts) != 222 || iconInstances != 666 {
		t.Fatalf("icon registry counts changed: %d concepts, %d color instances", len(iconConcepts), iconInstances)
	}
	if photos != 22 {
		t.Fatalf("photo registry has %d keys; want 22", photos)
	}
	for _, key := range []string{"photo-business-team-report-review", "photo-financial-adviser", "photo-financial-analyst", "photo-healthcare-leadership", "photo-software-developer-pair", "photo-software-engineering-team"} {
		meta := assetMetadataFor(key, "")
		if meta.kind != "photo" || meta.people != "yes" || meta.orientation != "landscape" || meta.description == "" || len(meta.tags) < 5 {
			t.Errorf("photo %s is missing curated searchable metadata: %+v", key, meta)
		}
	}
}

func TestAssetIndexProjectionIncludesCuratedDescriptionsAndTags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assets.sqlite")
	_, err := BuildLibraryIndex(path, LibraryIndexOptions{Bundle: filepath.Join("..", "..", "library", "wm-design-system", "v5")})
	if err != nil {
		t.Fatal(err)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	entity, err := index.Inspect("photo-working-session")
	if err != nil {
		t.Fatal(err)
	}
	if entity.Name != "Colleagues in a focused working session" || !strings.Contains(entity.Purpose, "collaborate around a table") {
		t.Fatalf("asset entity lacks curated human metadata: %+v", entity)
	}
	var definition map[string]any
	if err = json.Unmarshal(entity.Definition, &definition); err != nil {
		t.Fatal(err)
	}
	if definition["key"] != "photo-working-session" || definition["kind"] != "photo" || definition["people"] != "yes" {
		t.Fatalf("asset definition lost stable identity or tags: %+v", definition)
	}
}
