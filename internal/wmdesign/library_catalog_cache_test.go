package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func catalogCacheBundle() string {
	return filepath.Join("..", "..", "library", "wm-design-system", "v5")
}

func catalogCacheSource(t testing.TB) *Source {
	t.Helper()
	s, err := Load(catalogCacheBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestLibraryCatalogCacheConcurrentColdAndWarm(t *testing.T) {
	cache := libraryCatalogCache{limit: 2}
	for _, phase := range []string{"cold", "warm"} {
		t.Run(phase, func(t *testing.T) {
			const callers = 6
			start := make(chan struct{})
			errors := make(chan error, callers)
			var waiting sync.WaitGroup
			for i := 0; i < callers; i++ {
				waiting.Add(1)
				go func(i int) {
					defer waiting.Done()
					<-start
					s, err := Load(catalogCacheBundle(), "")
					if err != nil {
						errors <- err
						return
					}
					definitions, err := cache.catalog(s)
					if err != nil {
						errors <- err
						return
					}
					if len(definitions) != 587 || definitions[0].RawSlide[0] != '{' {
						errors <- fmt.Errorf("caller %d observed another caller's mutations", i)
						return
					}
					definitions[0].RawSlide[0] = byte('a' + i)
					definitions[0].Authoring.Slots[0].Alias = fmt.Sprintf("caller-%d", i)
					definitions[0].Discovery.Zones[0].Bounds["x"] = i
				}(i)
			}
			close(start)
			waiting.Wait()
			close(errors)
			for err := range errors {
				t.Error(err)
			}
			cache.mu.Lock()
			builds, entries := cache.builds, len(cache.entries)
			cache.mu.Unlock()
			if builds != 1 || entries != 1 {
				t.Fatalf("concurrent %s callers constructed %d catalogs, retained %d", phase, builds, entries)
			}
		})
	}
}

func TestLibraryCatalogCacheCloneIsolation(t *testing.T) {
	cache := libraryCatalogCache{limit: 1}
	s := catalogCacheSource(t)
	original, err := buildLibraryCatalog(s)
	if err != nil {
		t.Fatal(err)
	}
	first, err := cache.catalog(s)
	if err != nil || !reflect.DeepEqual(first, original) {
		t.Fatal("cache changed catalog values", err)
	}
	mutateCatalogTestValue(reflect.ValueOf(&first).Elem())
	second, err := cache.catalog(s)
	if err != nil || !reflect.DeepEqual(second, original) {
		t.Fatal("caller mutation escaped into retained catalog", err)
	}
	// Populate fields absent from the real source so new nested containers are
	// covered too. RawSlide has json:"-", so JSON serialization is not a clone.
	synthetic := []LibraryTemplate{{TemplateDefinition: TemplateDefinition{Guidance: []TemplateGuidance{{Slot: "one"}}}, PendingCapabilities: []string{"pending"}, Discovery: LibraryDiscovery{Relationships: []LibraryRelationship{{SourcePointers: []string{"/body/0"}}}, Zones: []LibraryContentZone{{Bounds: map[string]any{"tree": []any{map[string]any{"raw": json.RawMessage(`{"x":1}`)}}}}}}}}
	untouched := cloneLibraryCatalog(synthetic)
	clone := cloneLibraryCatalog(synthetic)
	mutateCatalogTestValue(reflect.ValueOf(&clone).Elem())
	if !reflect.DeepEqual(synthetic, untouched) {
		t.Fatal("nested synthetic containers were shared")
	}
}

// Mutate every populated field, recursively including maps with non-addressable
// values. This catches clone omissions if definition structs gain new fields.
func mutateCatalogTestValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			mutateCatalogTestValue(v.Elem())
		}
	case reflect.Interface:
		if !v.IsNil() {
			child := reflect.New(v.Elem().Type()).Elem()
			child.Set(v.Elem())
			mutateCatalogTestValue(child)
			v.Set(child)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			mutateCatalogTestValue(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			mutateCatalogTestValue(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			child := reflect.New(v.Type().Elem()).Elem()
			child.Set(v.MapIndex(key))
			mutateCatalogTestValue(child)
			v.SetMapIndex(key, child)
		}
	case reflect.String:
		v.SetString(v.String() + "-mutated")
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(v.Int() + 1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(v.Uint() + 1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(v.Float() + 1)
	}
}

func TestLibraryCatalogCacheFingerprintDependencies(t *testing.T) {
	s := catalogCacheSource(t)
	baseline, err := libraryCatalogFingerprint(s)
	if err != nil {
		t.Fatal(err)
	}
	changes := map[string]func(*Source){
		"revision":       func(s *Source) { s.Revision = "changed" },
		"commit":         func(s *Source) { s.Commit = "changed" },
		"manifest":       func(s *Source) { s.Files[0].SHA256 = "changed" },
		"font_manifest":  func(s *Source) { s.loadedFontFiles[0].SHA256 = "changed" },
		"tokens":         func(s *Source) { s.Tokens.Type[0].Size++ },
		"frames":         func(s *Source) { s.Frames.Splits["new"] = json.RawMessage(`{}`) },
		"components":     func(s *Source) { s.Components[0]++ },
		"resolved_style": func(s *Source) { st := s.styles["body"]; st.Size++; s.styles["body"] = st },
		"template": func(s *Source) {
			for key := range s.Templates {
				s.Templates[key][0]++
				break
			}
		},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			modified := catalogCacheSource(t)
			change(modified)
			key, err := libraryCatalogFingerprint(modified)
			if err != nil || key == baseline {
				t.Fatal("dependency absent from fingerprint", err)
			}
			if _, err := LibraryCatalogFromSource(modified); err == nil {
				t.Fatal("modified loaded snapshot accepted")
			}
		})
	}
	otherRoot := *s
	otherRoot.Root = "/another/verified/snapshot"
	key, err := libraryCatalogFingerprint(&otherRoot)
	if err != nil || key != baseline {
		t.Fatal("identical content unnecessarily partitioned by root", err)
	}
	if _, err := LibraryCatalogFromSource(&otherRoot); err == nil {
		t.Fatal("changed loaded snapshot root accepted")
	}
	if _, err := LibraryCatalogFromSource(&Source{}); err == nil {
		t.Fatal("fabricated source accepted")
	}
	if _, err := LibraryCatalogFromSource(nil); err == nil {
		t.Fatal("nil source accepted")
	}
}

func TestLibraryCatalogCacheBoundedEvictionAndErrors(t *testing.T) {
	cache := libraryCatalogCache{limit: 2}
	s := catalogCacheSource(t)
	// Eviction policy needs distinct keys, not four expensive full-library
	// rebuilds. Real 587-template cold concurrency and clone isolation are above.
	path := "templates/library/core.json"
	var catalog templateSourceCatalog
	if err := json.Unmarshal(s.Templates[path], &catalog); err != nil {
		t.Fatal(err)
	}
	var selected json.RawMessage
	for _, raw := range catalog.Templates {
		var entry libraryEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			t.Fatal(err)
		}
		if entry.ID == "key-message" && entry.Variant == "statement" {
			selected = raw
			break
		}
	}
	if selected == nil {
		t.Fatal("small catalog fixture absent")
	}
	// Retain the earliest profile's count gate with small, unique text-only
	// entries. This is an internal cache policy fixture, never a pinned release.
	catalog.Templates = nil
	s.Revision = LibraryRevisionV1
	for i := 0; i < 97; i++ {
		var entry libraryEntry
		if err := json.Unmarshal(selected, &entry); err != nil {
			t.Fatal(err)
		}
		entry.ID = "cache-fixture"
		entry.Variant = fmt.Sprintf("case-%03d", i+1)
		item, err := json.Marshal(entry)
		if err != nil {
			t.Fatal(err)
		}
		catalog.Templates = append(catalog.Templates, item)
	}
	raw, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	s.Templates = map[string]json.RawMessage{path: raw}
	s.Files = []SourceFile{{Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw))}}
	for _, commit := range []string{"one", "two", "one", "three", "two"} {
		s.Commit = commit
		if _, err := cache.catalog(s); err != nil {
			t.Fatal(err)
		}
		if len(cache.entries) > 2 {
			t.Fatal("cache exceeded entry limit")
		}
	}
	if cache.builds != 4 {
		t.Fatalf("LRU did not retain recently used entries: builds=%d", cache.builds)
	}
	for path, raw := range s.Templates {
		s.Templates[path] = append([]byte(nil), raw...)
		s.Templates[path][0] = 'x'
		if _, err := cache.catalog(s); err == nil || !strings.Contains(err.Error(), "invalid character") {
			t.Fatal("modified source masked by cache", err)
		}
		s.Templates[path] = raw
		break
	}
	if cache.builds != 4 || len(cache.entries) != 2 {
		t.Fatal("failed construction retained")
	}
}

func TestLibraryCatalogCacheDriftAfterWarm(t *testing.T) {
	root := filepath.Join(t.TempDir(), "bundle")
	copyCatalogCacheVerifiedFiles(t, root)
	for _, file := range []string{"bundle.json", "inventory.json", "source/templates/library/heatmaps.json", "source/tokens/v0/tokens.json", "source/frames/v0/frames.json", "source/components/v0/components.json", "fonts/IBMPlexSans-Regular.ttf"} {
		t.Run(file, func(t *testing.T) {
			before, err := LibraryCatalog(root, "")
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, file)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, append(bytes.Clone(raw), '\n'), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := LibraryCatalog(root, ""); err == nil || !strings.Contains(err.Error(), "drift") {
				t.Fatal("warm cache hid on-disk drift", err)
			}
			if err := os.WriteFile(path, raw, 0644); err != nil {
				t.Fatal(err)
			}
			after, err := LibraryCatalog(root, "")
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("restored snapshot did not recover", err)
			}
		})
	}
	loaded, err := Load(root, "")
	if err != nil {
		t.Fatal(err)
	}
	fontPath := filepath.Join(root, "fonts", "IBMPlexSans-Regular.ttf")
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = LibraryCatalogFromSource(loaded); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(fontPath, append(bytes.Clone(fontBytes), '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = LibraryCatalogFromSource(loaded); err == nil || !strings.Contains(err.Error(), "bundle_asset_drift") {
		t.Fatal("loaded snapshot warm cache hid font drift", err)
	}
	if _, err = libraryCatalog(loaded); err == nil || !strings.Contains(err.Error(), "bundle_asset_drift") {
		t.Fatal("internal warm cache hid font drift", err)
	}
	if err = os.WriteFile(fontPath, fontBytes, 0644); err != nil {
		t.Fatal(err)
	}
	calibration, err := CandidateCalibrationPath(loaded.loadedFontRoot)
	if err != nil {
		t.Fatal(err)
	}
	calibrationBytes, err := os.ReadFile(calibration)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(calibration, append(bytes.Clone(calibrationBytes), '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = LibraryCatalogFromSource(loaded); err == nil || !strings.Contains(err.Error(), "calibration_drift") {
		t.Fatal("loaded warm cache hid calibration drift", err)
	}
	if _, err = LibraryCatalog(root, ""); err == nil || !strings.Contains(err.Error(), "calibration_drift") {
		t.Fatal("validated warm cache hid calibration drift", err)
	}
	if err = os.WriteFile(calibration, calibrationBytes, 0644); err != nil {
		t.Fatal(err)
	}
	extraFont := filepath.Join(loaded.loadedFontRoot, "extra.ttf")
	if err = os.WriteFile(extraFont, fontBytes, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = LibraryCatalogFromSource(loaded); err == nil || !strings.Contains(err.Error(), "unpinned_catalog_font") {
		t.Fatal("warm cache hid additional unverified font", err)
	}
	if err = os.Remove(extraFont); err != nil {
		t.Fatal(err)
	}
	override := filepath.Join(t.TempDir(), "source")
	if err := os.CopyFS(override, os.DirFS(filepath.Join(root, "source"))); err != nil {
		t.Fatal(err)
	}
	if _, err := LibraryCatalog(root, override); err != nil {
		t.Fatal(err)
	}
	loaded, err = Load(root, override)
	if err != nil {
		t.Fatal(err)
	}
	// Force a cold derivation: an override has no adjacent fonts directory and
	// must still use the verified bundle fonts, not a cached result by accident.
	isolated := libraryCatalogCache{limit: 1}
	if _, err = isolated.catalog(loaded); err != nil {
		t.Fatal("cold source override lost verified bundle fonts", err)
	}
	path := filepath.Join(override, "templates/library/heatmaps.json")
	if err := os.WriteFile(path, []byte("changed override"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LibraryCatalog(root, override); err == nil || !strings.Contains(err.Error(), "snapshot_drift") {
		t.Fatal("warm cache hid override drift", err)
	}
}

func copyCatalogCacheVerifiedFiles(t *testing.T, root string) {
	t.Helper()
	// Copy only files Load validates, not 170MiB of unrelated review artifacts.
	paths := map[string]bool{"bundle.json": true, "inventory.json": true}
	manifest, err := os.ReadFile(filepath.Join(catalogCacheBundle(), "bundle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var bundle struct{ Files []SourceFile }
	if err := json.Unmarshal(manifest, &bundle); err != nil {
		t.Fatal(err)
	}
	for _, file := range bundle.Files {
		paths[file.Path] = true
	}
	raw, err := os.ReadFile(filepath.Join(catalogCacheBundle(), "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inventory Inventory
	if err := json.Unmarshal(raw, &inventory); err != nil {
		t.Fatal(err)
	}
	for _, file := range inventory.Sources {
		paths["source/"+file.Path] = true
	}
	for relative := range paths {
		raw, err := os.ReadFile(filepath.Join(catalogCacheBundle(), relative))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	calibration, err := CandidateCalibrationPath(filepath.Join(catalogCacheBundle(), "fonts"))
	if err != nil {
		t.Fatal(err)
	}
	calibrationBytes, err := os.ReadFile(calibration)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "typography", "calibration.json")
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, calibrationBytes, 0644); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkLibraryCatalogCacheUncached(b *testing.B) {
	s := catalogCacheSource(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := buildLibraryCatalog(s); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLibraryCatalogCacheWarm(b *testing.B) {
	s := catalogCacheSource(b)
	cache := libraryCatalogCache{limit: 1}
	if _, err := cache.catalog(s); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := cache.catalog(s); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLibraryCatalogCacheValidatedLoad(b *testing.B) {
	if _, err := LibraryCatalog(catalogCacheBundle(), ""); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := LibraryCatalog(catalogCacheBundle(), ""); err != nil {
			b.Fatal(err)
		}
	}
}
