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

type sourcePin struct {
	Inventory, Revision string
}

var sourcePins = map[string]sourcePin{
	bundleSHA256: {inventorySHA256, LibraryRevisionV1},
	"c0926ec4e65d36b3a9fd53e74ae0a3204d03acd5d0ba9fe4919849700c8f3690": {"8e70c96c07b5346906c983f0893e686cd433fd73fae4c31642a71984f395dce3", LibraryRevisionV2},
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
