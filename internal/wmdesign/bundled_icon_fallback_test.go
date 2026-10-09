package wmdesign

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func iconBundleFixture(t *testing.T) (string, string) {
	t.Helper()
	root, e := filepath.Abs("../../library/wm-design-system/v11")
	if e != nil {
		t.Fatal(e)
	}
	return root, "icon/browser-gear/white"
}
func TestBundledRegisteredIconExactOriginalAndMissing(t *testing.T) {
	root, key := iconBundleFixture(t)
	ref := primitiveAssetRegistry[key]
	data, found, e := bundledRegisteredIcon(root, key, ref.SHA256)
	if e != nil || !found || len(data) == 0 {
		t.Fatal(found, e)
	}
	if _, found, e = bundledRegisteredIcon(t.TempDir(), key, ref.SHA256); e != nil || found {
		t.Fatal(found, e)
	}
	if _, found, e = bundledRegisteredIcon(root, "photo-business-team-report-review", ref.SHA256); e != nil || found {
		t.Fatal("photo fallback allowed")
	}
}
func TestBundledRegisteredIconDriftDerivedAndTraversalRefused(t *testing.T) {
	root, key := iconBundleFixture(t)
	raw, e := os.ReadFile(filepath.Join(root, "catalog/assets/assets.json"))
	if e != nil {
		t.Fatal(e)
	}
	var catalog []AssetSelection
	if e = json.Unmarshal(raw, &catalog); e != nil {
		t.Fatal(e)
	}
	var selected AssetSelection
	for _, a := range catalog {
		for _, v := range a.Variants {
			if v.ID == key {
				selected = AssetSelection{ID: a.ID, Kind: a.Kind, Variants: []AssetVariant{v}}
			}
		}
	}
	if len(selected.Variants) == 0 {
		t.Fatal("fixture missing")
	}
	original := selected.Variants[0]
	data, e := os.ReadFile(filepath.Join(root, "catalog/assets", original.ThumbnailPath))
	if e != nil {
		t.Fatal(e)
	}
	for _, scenario := range []string{"drift", "derived", "path", "duplicate", "missing"} {
		t.Run(scenario, func(t *testing.T) {
			dst := t.TempDir()
			a := selected
			payload := append([]byte(nil), data...)
			a.Variants = append([]AssetVariant(nil), selected.Variants...)
			switch scenario {
			case "drift":
				payload = []byte("not original")
			case "derived":
				a.Variants[0].ThumbnailState = "derived_preview"
			case "path":
				a.Variants[0].ThumbnailPath = "../outside.svg"
			case "duplicate":
				a.Variants = append(a.Variants, a.Variants[0])
			}
			path := filepath.Join(dst, "catalog/assets")
			if e = os.MkdirAll(filepath.Join(path, "assets"), 0700); e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(path, "assets.json"), mustJSON(t, []AssetSelection{a}), 0600); e != nil {
				t.Fatal(e)
			}
			if scenario != "missing" {
				if e = os.WriteFile(filepath.Join(path, original.ThumbnailPath), payload, 0600); e != nil {
					t.Fatal(e)
				}
			}
			if _, _, e = bundledRegisteredIcon(dst, key, original.SHA256); e == nil {
				t.Fatal("invalid original accepted")
			}
		})
	}
}
func TestBundledRegisteredIconRendererOnlyMissingFallback(t *testing.T) {
	root, key := iconBundleFixture(t)
	source, e := Load(root, "")
	if e != nil {
		t.Fatal(e)
	}
	r := &renderer{source: source}
	branding := t.TempDir()
	t.Setenv("WMDS_BRANDING_ROOT", branding)
	data, a, e := r.primitiveAssetBytes(key)
	if e != nil || len(data) == 0 || a.SHA256 != primitiveAssetRegistry[key].SHA256 {
		t.Fatal(e)
	}
	path := filepath.Join(branding, primitiveAssetRegistry[key].Path)
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, []byte("corrupt local original"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, _, e = r.primitiveAssetBytes(key); e == nil || !strings.Contains(e.Error(), "hash_drift") {
		t.Fatal("local drift hidden by fallback", e)
	}
}
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func TestBundledRegisteredIconApprovedArrowsOnly(t *testing.T) {
	root, _ := iconBundleFixture(t)
	source, e := Load(root, "")
	if e != nil {
		t.Fatal(e)
	}
	r := &renderer{source: source}
	t.Setenv("WMDS_BRANDING_ROOT", t.TempDir())
	for _, key := range []string{"arrow-straight", "arrow-connecting", "arrow-dashed", "arrow-double", "arrow-right-angle"} {
		t.Run(key, func(t *testing.T) {
			ref := primitiveAssetRegistry[key]
			original, found, err := bundledRegisteredIcon(root, key, ref.SHA256)
			if err != nil || !found {
				t.Fatal(found, err)
			}
			data, registered, err := r.primitiveAssetBytes(key)
			if err != nil || !bytes.Equal(data, original) || registered != ref {
				t.Fatal("registered arrow changed", err)
			}
			if _, _, err = bundledRegisteredIcon(root, key, strings.Repeat("0", 64)); err == nil {
				t.Fatal("arrow hash mismatch accepted")
			}
		})
	}
	for _, key := range []string{"circle", "highlight-1", "photo-business-team-report-review", "logo-west-monroe", "arrow-unknown", "icon/unknown/white"} {
		if bundledOriginalEligible(key) {
			t.Fatalf("unapproved original eligible: %s", key)
		}
		if _, found, err := bundledRegisteredIcon(root, key, strings.Repeat("0", 64)); err != nil || found {
			t.Fatalf("unapproved fallback %s: %v", key, err)
		}
	}
}
