// Package wmdesign implements the explicitly versioned WMDS foundation profile.
// Existing compose and textlayout profiles are independent of this package.
package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const Profile = "wmds-native.v1"
const Engine = "wmds-go-foundation.v1"
const inventorySHA256 = "efe3801e16c282d1ea1fcffefc7f360a6a0afea924b0f9fe45798ccf92d8a012"
const bundleSHA256 = "02a975693995f9aca88602cd64e159e054da5a9771d8e3205d3beffd3d3ee45e"

const LibraryRevisionV1 = "wmds-library.v1"
const LibraryRevisionV2 = "wmds-library.v2"
const LibraryRevisionV3 = "wmds-library.v3"
const LibraryRevisionV4 = "wmds-library.v4"
const LibraryRevisionV5 = "wmds-library.v5"
const LibraryRevisionV6 = "wmds-library.v6"
const LibraryRevisionV7 = "wmds-library.v7"
const LibraryRevisionV8 = "wmds-library.v8"
const LibraryRevisionV9 = "wmds-library.v9"

func isV6OrLaterLibrary(revision string) bool {
	return revision == LibraryRevisionV6 || revision == LibraryRevisionV7 || revision == LibraryRevisionV8 || revision == LibraryRevisionV9
}

// V6 carries the accepted V5 rendering semantics for unchanged compositions.
// New heat-map fields are separately gated to the pinned V6 source.
func isV5OrLaterLibrary(revision string) bool {
	return revision == LibraryRevisionV5 || isV6OrLaterLibrary(revision)
}

// Expanded revisions share the incoming diagram, fit and data semantics. Named
// composition amendments remain gated by each revision's frozen identity map.
func isExpandedLibrary(revision string) bool {
	return revision == LibraryRevisionV4 || isV5OrLaterLibrary(revision)
}

func isModernLibrary(revision string) bool {
	return revision == LibraryRevisionV2 || revision == LibraryRevisionV3 || isExpandedLibrary(revision)
}

type sourcePin struct {
	Inventory, Revision string
}

var sourcePins = map[string]sourcePin{
	"0ad33b662d7e3a9b47a47237f0037c09cba959ee2693d39b4cc36ed230ab75c1": {"9ae0d5af692d4dfe3c9807d232202c2461f55168fe926689ee833f82f8001fb3", LibraryRevisionV9},
	"0e9846c1cf96187a16ca210279297239169707c90743c80011ef76869fbf26c9": {"9ae84d46370e91146395d860379dcfb8ef4bf72c6623482f3451c68d9eaf64d3", LibraryRevisionV8},
	"ce5bd8ad00261da22418f9ec9e93a79d6938f368aef3df3f3a6aa46d3833dc93": {"4f057cfbfd94feb4007f82ca6dfbbba1a0d410e89ec4f37f1b3e74343489abd2", LibraryRevisionV7},
	bundleSHA256: {inventorySHA256, LibraryRevisionV1},
	"c0926ec4e65d36b3a9fd53e74ae0a3204d03acd5d0ba9fe4919849700c8f3690": {"8e70c96c07b5346906c983f0893e686cd433fd73fae4c31642a71984f395dce3", LibraryRevisionV2},
	"38819ed1eb6f48288935e30afe5894a471dda068488d067364585a1aa79706fa": {"56968b1e859945f7cea178fce43e5f1bacc66280cbe515c532c99b37924774cf", LibraryRevisionV3},
	"1e70050967a80c2d232b6109ac8428f321062fc829ba0358a9a56b54726d14b3": {"f9bcc2b425fb04feb6be9d9d0b5e5838f3c70799ff873b2f01d78d9d0c90e804", LibraryRevisionV4},
	"0bdb3c6c7b327ed98a62b0527826068db371cb5d9cb43a1c6f44a1d4a74bf7eb": {"c4c64f6fc07eba612ccb2af6d05e70402fb26d583e42f44bbaa1325e49a57090", LibraryRevisionV5},
	"9b1c303958152e6adddbb84b1fcde1bdb73c1a6f43ac2906ecbbd9b446c28716": {"586a742ee2bdfb05ee955af613f4c36e2059ed3504abd69cf652966e9781132e", LibraryRevisionV6},
}

func checkBundle(root string) error {
	data, err := os.ReadFile(filepath.Join(root, "bundle.json"))
	if err != nil {
		return err
	}
	if _, ok := sourcePins[fmt.Sprintf("%x", sha256.Sum256(data))]; !ok {
		return fmt.Errorf("source.bundle_drift: adapter migration required")
	}
	var manifest struct {
		Files []SourceFile `json:"files"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	for _, f := range manifest.Files {
		if filepath.IsAbs(f.Path) || strings.Contains(f.Path, "..") {
			return fmt.Errorf("source.invalid_bundle_path")
		}
		b, e := os.ReadFile(filepath.Join(root, f.Path))
		if e != nil {
			return e
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != f.SHA256 {
			return fmt.Errorf("source.bundle_asset_drift: %s", f.Path)
		}
	}
	return nil
}

type SourceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type Inventory struct {
	Sources  []SourceFile `json:"sources"`
	Revision string       `json:"source_revision,omitempty"`
	Commit   string       `json:"source_commit,omitempty"`
}
type Style struct {
	ID         string  `json:"id"`
	Family     string  `json:"font"`
	Weight     int     `json:"weight"`
	Italic     bool    `json:"italic"`
	Size       float64 `json:"size"`
	Leading    float64 `json:"leading"`
	Tracking   string  `json:"tracking"`
	Case       string  `json:"case,omitempty"`
	Use        string  `json:"use"`
	TrackingPt float64 `json:"tracking_pt"`
}
type Grid struct {
	Slide    [2]float64 `json:"slide"`
	Module   float64    `json:"module"`
	Baseline float64    `json:"baseline"`
	MarginX  float64    `json:"marginX"`
	MarginY  float64    `json:"marginY"`
	Content  float64    `json:"content"`
	Columns  int        `json:"columns"`
	Column   float64    `json:"column"`
	Gutter   float64    `json:"gutter"`
	FiveUp   struct {
		Columns int       `json:"columns"`
		Column  float64   `json:"column"`
		Gutter  float64   `json:"gutter"`
		Starts  []float64 `json:"starts"`
	} `json:"fiveUp"`
}
type Tokens struct {
	Schema string  `json:"schema"`
	Units  string  `json:"units"`
	Type   []Style `json:"type"`
	Grid   Grid    `json:"grid"`
	Colors struct {
		Surfaces map[string]map[string]string `json:"surfaces"`
		Dataviz  struct {
			KPI map[string]string `json:"kpi"`
		} `json:"dataviz"`
	} `json:"colors"`
}
type Rail struct {
	Main     [2]float64 `json:"main"`
	Panel    [2]float64 `json:"panel"`
	Content  [2]float64 `json:"railContent"`
	Surfaces []string   `json:"surfaces"`
}
type Footer struct {
	Rule    float64    `json:"rule"`
	Row     [2]float64 `json:"row"`
	Bottom  float64    `json:"bodyBottom"`
	Band    [2]float64 `json:"band"`
	Surface string     `json:"surface"`
}
type Frames struct {
	Schema  string                     `json:"schema"`
	Units   string                     `json:"units"`
	Rails   map[string]Rail            `json:"rails"`
	Footers map[string]Footer          `json:"footers"`
	Splits  map[string]json.RawMessage `json:"splits,omitempty"`
	Chrome  []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"chrome"`
	Features []struct {
		ID       string          `json:"id"`
		Geometry json.RawMessage `json:"geometry"`
	} `json:"features"`
}
type Source struct {
	Root     string       `json:"root"`
	Revision string       `json:"source_revision"`
	Commit   string       `json:"source_commit,omitempty"`
	Files    []SourceFile `json:"files"`
	Tokens   Tokens       `json:"tokens"`
	Frames   Frames       `json:"frames"`
	// Frozen definitions retain full metadata. Only explicitly contracted text, card,
	// card-row and standalone-metric subsets have executable v2 adapter semantics.
	Components json.RawMessage            `json:"components"`
	Templates  map[string]json.RawMessage `json:"templates"`
	styles     map[string]Style
	// These attest only to this in-memory snapshot. Load still validates every
	// file on every call; they never allow a path/mtime cache to bypass drift.
	loadedCatalogKey [32]byte
	loadedRoot       string
	loadedFontRoot   string
	loadedFontFiles  []SourceFile
}

// Load accepts only explicitly pinned source revisions. Hash mismatches are source
// drift errors: unknown fields cannot silently acquire rendering semantics.
// A source override must match this inventory until a new adapter is qualified.
func Load(bundle, override string) (*Source, error) {
	if err := checkBundle(bundle); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(bundle, "inventory.json"))
	if err != nil {
		return nil, err
	}
	bundleData, err := os.ReadFile(filepath.Join(bundle, "bundle.json"))
	if err != nil {
		return nil, err
	}
	pin := sourcePins[fmt.Sprintf("%x", sha256.Sum256(bundleData))]
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != pin.Inventory {
		return nil, fmt.Errorf("source.inventory_drift: adapter migration required")
	}
	var inv Inventory
	if err = json.Unmarshal(raw, &inv); err != nil {
		return nil, err
	}
	if inv.Revision != "" && inv.Revision != pin.Revision {
		return nil, fmt.Errorf("source.revision_mismatch")
	}
	root := override
	if root == "" {
		root = filepath.Join(bundle, "source")
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	s := &Source{Root: root, Revision: pin.Revision, Commit: inv.Commit, Files: inv.Sources, Templates: map[string]json.RawMessage{}, styles: map[string]Style{}}
	// Font-derived authoring plans use the verified bundle fonts even when an
	// independently verified source override lives outside that bundle.
	s.loadedFontRoot, err = filepath.Abs(filepath.Join(bundle, "fonts"))
	if err != nil {
		return nil, err
	}
	var bundleManifest struct {
		Files []SourceFile `json:"files"`
	}
	if err = json.Unmarshal(bundleData, &bundleManifest); err != nil {
		return nil, err
	}
	for _, file := range bundleManifest.Files {
		if strings.HasPrefix(file.Path, "fonts/") {
			file.Path = strings.TrimPrefix(file.Path, "fonts/")
			s.loadedFontFiles = append(s.loadedFontFiles, file)
		}
	}
	if len(inv.Sources) == 0 {
		return nil, fmt.Errorf("source.invalid_inventory: empty snapshot")
	}
	for _, f := range inv.Sources {
		if filepath.IsAbs(f.Path) || strings.Contains(f.Path, "..") {
			return nil, fmt.Errorf("source.invalid_path: %s", f.Path)
		}
		b, e := os.ReadFile(filepath.Join(root, f.Path))
		if e != nil {
			return nil, e
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != f.SHA256 {
			return nil, fmt.Errorf("source.snapshot_drift: %s; explicit adapter migration required", f.Path)
		}
		switch f.Path {
		case "tokens/v0/tokens.json":
			err = json.Unmarshal(b, &s.Tokens)
		case "frames/v0/frames.json":
			err = json.Unmarshal(b, &s.Frames)
		case "components/v0/components.json":
			s.Components = append([]byte(nil), b...)
		default:
			if strings.HasPrefix(f.Path, "templates/library/") && !strings.HasPrefix(filepath.Base(f.Path), "_") {
				s.Templates[f.Path] = append([]byte(nil), b...)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("source.invalid_json: %s: %w", f.Path, err)
		}
	}
	if s.Tokens.Schema != "wmds.tokens.v0" || s.Frames.Schema != "wmds.frames.v0" || s.Tokens.Units != "pt" || s.Frames.Units != "pt" {
		return nil, fmt.Errorf("source.unsupported_schema")
	}
	for i, st := range s.Tokens.Type {
		if _, ok := s.styles[st.ID]; ok {
			return nil, fmt.Errorf("source.duplicate_style: %s", st.ID)
		}
		tr := strings.TrimSuffix(st.Tracking, "em")
		v, e := strconv.ParseFloat(tr, 64)
		if e != nil {
			return nil, fmt.Errorf("style %s tracking: %w", st.ID, e)
		}
		if strings.HasSuffix(st.Tracking, "em") {
			v *= st.Size
		}
		st.TrackingPt = math.Round(v*100) / 100
		if st.Weight != 400 && st.Weight != 500 && st.Weight != 600 && st.Weight != 700 {
			return nil, fmt.Errorf("font.unsupported_weight: %d", st.Weight)
		}
		if st.Size <= 0 || st.Leading <= 0 {
			return nil, fmt.Errorf("style.invalid_metrics: %s", st.ID)
		}
		s.styles[st.ID] = st
		s.Tokens.Type[i] = st
	}
	if len(s.styles) != 14 {
		return nil, fmt.Errorf("source.conflict: expected 14 styles")
	}
	s.loadedCatalogKey, err = libraryCatalogFingerprint(s)
	if err != nil {
		return nil, err
	}
	s.loadedRoot = s.Root
	return s, nil
}
func (s *Source) Style(id string) (Style, error) {
	v, ok := s.styles[id]
	if !ok {
		return v, fmt.Errorf("style.unknown: %s", id)
	}
	return v, nil
}
func (s *Source) Ink(surface, role string) (string, error) {
	if surface == "outline" {
		surface = "light"
	}
	r, ok := s.Tokens.Colors.Surfaces[surface]
	if !ok {
		return "", fmt.Errorf("surface.unknown: %s", surface)
	}
	v, ok := r[role]
	if !ok {
		return "", fmt.Errorf("ink.unknown: %s/%s", surface, role)
	}
	return strings.TrimPrefix(v, "#"), nil
}
