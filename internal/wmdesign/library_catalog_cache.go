package wmdesign

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Four entries bound retained catalogs independently of caller paths. Cold
// construction is serialized, so concurrent misses cannot duplicate large
// catalog builds or create an unbounded collection of in-flight entries.
const libraryCatalogCacheLimit = 4

var sharedLibraryCatalogCache = libraryCatalogCache{limit: libraryCatalogCacheLimit}

type libraryCatalogCacheEntry struct {
	definitions []LibraryTemplate
	used        uint64
}

type libraryCatalogCache struct {
	mu      sync.Mutex
	entries map[[32]byte]libraryCatalogCacheEntry
	clock   uint64
	limit   int
	builds  uint64
}

// Hash actual content, not path, mtime, or a manifest's asserted hashes. Include
// the whole manifest and all source metadata dependencies, including resolved
// styles used by authoring capacities. Root is deliberately excluded: identical
// verified snapshots in different directories have identical derived catalogs.
func libraryCatalogFingerprint(s *Source) ([32]byte, error) {
	if s == nil {
		return [32]byte{}, fmt.Errorf("source.catalog_snapshot_required")
	}
	dependencies := struct {
		Revision    string
		Commit      string
		Files       []SourceFile
		Tokens      Tokens
		Frames      Frames
		Styles      map[string]Style
		Fonts       []SourceFile
		Calibration string
	}{s.Revision, s.Commit, s.Files, s.Tokens, s.Frames, s.styles, s.loadedFontFiles, CandidateCalibrationSHA}
	raw, err := json.Marshal(dependencies)
	if err != nil {
		return [32]byte{}, fmt.Errorf("source.catalog_dependency_encoding: %w", err)
	}
	h := sha256.New()
	write := func(data []byte) {
		var length [8]byte
		binary.LittleEndian.PutUint64(length[:], uint64(len(data)))
		h.Write(length[:])
		h.Write(data)
	}
	write(raw)
	write(s.Components)
	paths := make([]string, 0, len(s.Templates))
	for path := range s.Templates {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		write([]byte(path))
		write(s.Templates[path])
	}
	var key [32]byte
	copy(key[:], h.Sum(nil))
	return key, nil
}

// Font-derived plans read external files rather than only decoded Source
// fields. Check those dependencies on warm snapshot reuse as well as cold
// construction; path/mtime and asserted hashes alone are insufficient.
func verifyCatalogFonts(s *Source) error {
	known := map[string]bool{}
	for _, file := range s.loadedFontFiles {
		known[file.Path] = true
		raw, err := os.ReadFile(filepath.Join(s.loadedFontRoot, file.Path))
		if err != nil {
			return fmt.Errorf("source.bundle_asset_drift: fonts/%s: %w", file.Path, err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != file.SHA256 {
			return fmt.Errorf("source.bundle_asset_drift: fonts/%s", file.Path)
		}
	}
	entries, err := os.ReadDir(s.loadedFontRoot)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if (ext == ".ttf" || ext == ".otf" || ext == ".ttc") && !known[entry.Name()] {
			return fmt.Errorf("source.unpinned_catalog_font: %s", entry.Name())
		}
	}
	path, err := CandidateCalibrationPath(s.loadedFontRoot)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != CandidateCalibrationSHA {
		return fmt.Errorf("text.calibration_drift")
	}
	return nil
}

func (c *libraryCatalogCache) catalog(s *Source) ([]LibraryTemplate, error) {
	key, err := libraryCatalogFingerprint(s)
	if err != nil {
		return nil, err
	}
	return c.catalogKey(s, key)
}

func (c *libraryCatalogCache) catalogKey(s *Source, key [32]byte) ([]LibraryTemplate, error) {
	definitions, err := c.retained(s, key)
	if err != nil {
		return nil, err
	}
	// Clone outside the cache lock. Retained definitions are immutable, including
	// after eviction; each caller owns every returned mutable container/pointer.
	return cloneLibraryCatalog(definitions), nil
}

func (c *libraryCatalogCache) retained(s *Source, key [32]byte) ([]LibraryTemplate, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clock++
	if entry, ok := c.entries[key]; ok {
		entry.used = c.clock
		c.entries[key] = entry
		return entry.definitions, nil
	}
	definitions, err := buildLibraryCatalog(s)
	if err != nil {
		// Errors are never retained: correcting a snapshot must be observable.
		return nil, err
	}
	c.builds++
	if c.limit <= 0 {
		return definitions, nil
	}
	if c.entries == nil {
		c.entries = make(map[[32]byte]libraryCatalogCacheEntry)
	}
	if len(c.entries) >= c.limit {
		var oldest [32]byte
		var age uint64
		first := true
		for candidate, entry := range c.entries {
			if first || entry.used < age {
				oldest, age, first = candidate, entry.used, false
			}
		}
		delete(c.entries, oldest)
	}
	c.entries[key] = libraryCatalogCacheEntry{definitions, c.clock}
	return definitions, nil
}

func catalogCopySlice[T any](value []T) []T {
	if value == nil {
		return nil
	}
	out := make([]T, len(value))
	copy(out, value)
	return out
}

func catalogCopyMap[T any](value map[string]T) map[string]T {
	if value == nil {
		return nil
	}
	out := make(map[string]T, len(value))
	for key, child := range value {
		out[key] = child
	}
	return out
}

func cloneLibraryCatalog(definitions []LibraryTemplate) []LibraryTemplate {
	out := catalogCopySlice(definitions)
	for i := range out {
		d := &out[i]
		d.Identities = catalogCopySlice(d.Identities)
		d.Guidance = catalogCopySlice(d.Guidance)
		d.Uses = catalogCopySlice(d.Uses)
		d.AdvisoryBudget = catalogCopySlice(d.AdvisoryBudget)
		d.AdvisorySlots = catalogCopySlice(d.AdvisorySlots)
		d.RawSlide = catalogCopySlice(d.RawSlide)
		d.Slots = catalogCopySlice(d.Slots)
		for j := range d.Slots {
			d.Slots[j].Example = catalogCopySlice(d.Slots[j].Example)
		}
		d.Arrays = catalogCopySlice(d.Arrays)
		d.Policy = catalogCopySlice(d.Policy)
		d.PendingCapabilities = catalogCopySlice(d.PendingCapabilities)
		if d.ValueSchema != nil {
			schema := *d.ValueSchema
			schema.Fields = catalogCopyMap(schema.Fields)
			schema.ExactCounts = catalogCopyMap(schema.ExactCounts)
			d.ValueSchema = &schema
		}
		if d.Nav != nil {
			nav := *d.Nav
			nav.Example.Items = catalogCopySlice(nav.Example.Items)
			d.Nav = &nav
		}
		d.Discovery = cloneCatalogDiscovery(d.Discovery)
		if d.Authoring != nil {
			authoring := *d.Authoring
			authoring.Groups = catalogCopySlice(authoring.Groups)
			authoring.Policy = catalogCopySlice(authoring.Policy)
			authoring.Slots = catalogCopySlice(authoring.Slots)
			for j := range authoring.Slots {
				capacity := &authoring.Slots[j].Capacity
				capacity.Assumptions = catalogCopySlice(capacity.Assumptions)
				if capacity.Style != nil {
					style := *capacity.Style
					capacity.Style = &style
				}
			}
			d.Authoring = &authoring
		}
	}
	return out
}

func cloneCatalogDiscovery(discovery LibraryDiscovery) LibraryDiscovery {
	discovery.ContentRoles = catalogCopySlice(discovery.ContentRoles)
	discovery.Structures = catalogCopySlice(discovery.Structures)
	discovery.VisualForms = catalogCopySlice(discovery.VisualForms)
	discovery.ComponentTypes = catalogCopySlice(discovery.ComponentTypes)
	discovery.Groups = catalogCopySlice(discovery.Groups)
	discovery.Relationships = catalogCopySlice(discovery.Relationships)
	for i := range discovery.Relationships {
		discovery.Relationships[i].SourcePointers = catalogCopySlice(discovery.Relationships[i].SourcePointers)
	}
	discovery.Zones = catalogCopySlice(discovery.Zones)
	for i := range discovery.Zones {
		bounds := discovery.Zones[i].Bounds
		if bounds == nil {
			continue
		}
		discovery.Zones[i].Bounds = make(map[string]any, len(bounds))
		for key, value := range bounds {
			discovery.Zones[i].Bounds[key] = cloneCatalogJSON(value)
		}
	}
	return discovery
}

func cloneCatalogJSON(value any) any {
	switch value := value.(type) {
	case map[string]any:
		out := catalogCopyMap(value)
		for key, child := range out {
			out[key] = cloneCatalogJSON(child)
		}
		return out
	case []any:
		out := catalogCopySlice(value)
		for i, child := range out {
			out[i] = cloneCatalogJSON(child)
		}
		return out
	case json.RawMessage:
		return json.RawMessage(catalogCopySlice(value))
	default:
		// Bounds are decoded JSON: remaining values are immutable scalars.
		return value
	}
}
