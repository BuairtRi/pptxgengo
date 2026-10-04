// Package textlayout implements an opt-in, pure Go text measurement prototype.
// Its results describe this engine's layout, not PowerPoint's rendered layout.
package textlayout

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	"github.com/go-text/typesetting/fontscan"
)

type FontRecord struct {
	Family    string             `json:"family"`
	Style     string             `json:"style"`
	File      string             `json:"file"`
	FaceIndex int                `json:"face_index"`
	SHA256    string             `json:"sha256"`
	Axes      map[string]float32 `json:"axes,omitempty"`
	Weight    float32            `json:"weight"`
	Subfamily string             `json:"subfamily"`
}

type candidate struct {
	path        string
	index       int
	description font.Description
	axes        []tables.VariationAxisRecord
	boldFlag    bool
	subfamily   string
}

type resolvedFont struct {
	face   *font.Face
	record FontRecord
}

// Resolver selects an exact family, regular/bold weight, and upright/italic
// style. It never synthesizes a style or substitutes another family.
type Resolver struct {
	candidates []candidate
	loaded     map[string]resolvedFont
	Warnings   []string
}

// NewResolver scans TTF, OTF and TTC files. Explicit roots replace system font
// directories, making it possible to use a controlled, portable font set.
// No font index, system service, subprocess, or persistent cache is used.
func NewResolver(roots []string) (*Resolver, error) {
	if len(roots) == 0 {
		var err error
		roots, err = fontscan.DefaultFontDirectories(log.New(io.Discard, "", 0))
		// fontscan expands ~ through os/user.Current. Sandboxed macOS
		// processes can have a valid HOME without a directory-service user
		// record, so also resolve the user font directory through UserHomeDir.
		if runtime.GOOS == "darwin" {
			if home, homeErr := os.UserHomeDir(); homeErr == nil {
				userFonts := filepath.Join(home, "Library", "Fonts")
				if info, statErr := os.Stat(userFonts); statErr == nil && info.IsDir() {
					roots = append(roots, userFonts)
				}
			}
		}
		if err != nil && len(roots) == 0 {
			return nil, err
		}
	}
	r := &Resolver{loaded: map[string]resolvedFont{}}
	seen := map[string]bool{}
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("font directory %q is not accessible", root)
		}
		err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".ttf", ".otf", ".ttc":
			default:
				return nil
			}
			path, err = filepath.Abs(path)
			if err != nil {
				return err
			}
			if seen[path] {
				return nil
			}
			seen[path] = true
			if err := r.index(path); err != nil {
				r.Warnings = append(r.Warnings, fmt.Sprintf("skipped font %s: %v", path, err))
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scan font directory %q: %w", root, err)
		}
	}
	sort.Slice(r.candidates, func(i, j int) bool {
		a, b := r.candidates[i], r.candidates[j]
		if a.path != b.path {
			return a.path < b.path
		}
		return a.index < b.index
	})
	return r, nil
}

func (r *Resolver) index(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	loaders, err := ot.NewLoaders(f)
	if err != nil {
		return err
	}
	for i, loader := range loaders {
		desc, _ := font.Describe(loader, nil)
		boldFlag := false
		if raw, err := loader.RawTable(ot.MustNewTag("OS/2")); err == nil {
			os2, _, err := tables.ParseOs2(raw)
			if err != nil {
				return err
			}
			boldFlag = os2.FsSelection&(1<<5) != 0
		}
		if raw, err := loader.RawTable(ot.MustNewTag("head")); err == nil {
			head, _, err := tables.ParseHead(raw)
			if err != nil {
				return err
			}
			boldFlag = boldFlag || head.MacStyle&1 != 0
		}
		subfamily := ""
		if raw, err := loader.RawTable(ot.MustNewTag("name")); err == nil {
			names, _, err := tables.ParseName(raw)
			if err != nil {
				return err
			}
			subfamily = names.Name(17)
			if subfamily == "" {
				subfamily = names.Name(2)
			}
		}
		var axes []tables.VariationAxisRecord
		if raw, err := loader.RawTable(ot.MustNewTag("fvar")); err == nil {
			fv, _, err := tables.ParseFvar(raw)
			if err != nil {
				return err
			}
			axes = fv.FvarRecords.Axis
		}
		r.candidates = append(r.candidates, candidate{path: path, index: i, description: desc, axes: axes, boldFlag: boldFlag, subfamily: subfamily})
	}
	return nil
}

func styleName(bold, italic bool) string {
	if bold && italic {
		return "bold_italic"
	}
	if bold {
		return "bold"
	}
	if italic {
		return "italic"
	}
	return "regular"
}

// selection translates the authored Boolean styles to real font faces or
// standard variable axes. Unsupported weights/styles are rejected.
func selection(c candidate, bold, italic bool) (map[string]float32, bool) {
	weight := float32(400)
	if bold {
		weight = 700
	}
	aspect := c.description.Aspect
	axes := map[string]float32{}
	hasWeight, hasStyle, hasWidth := false, false, false
	for _, axis := range c.axes {
		if axis.Tag.String() == "ital" {
			hasStyle = true
		}
	}
	for _, axis := range c.axes {
		min, max, def := axis.Minimum, axis.Maximum, axis.Default
		value := def
		switch axis.Tag.String() {
		case "wght":
			value = weight
			hasWeight = true
		case "wdth":
			value = 100
			hasWidth = true
		case "ital":
			value = 0
			if italic {
				value = 1
			}
			hasStyle = true
		case "slnt":
			// An arbitrary slant angle is not equivalent to an italic face.
			if !hasStyle && aspect.Style == font.StyleNormal && italic {
				return nil, false
			}
			if !italic {
				value = 0
			}
		}
		if value < min || value > max {
			return nil, false
		}
		axes[axis.Tag.String()] = value
	}
	if !hasWeight {
		if bold {
			if !c.boldFlag && math.Abs(float64(aspect.Weight)-700) > 1 {
				return nil, false
			}
		} else if c.boldFlag || math.Abs(float64(aspect.Weight)-400) > 1 {
			return nil, false
		}
	}
	if !hasWidth && math.Abs(float64(aspect.Stretch)-1) > 0.001 {
		return nil, false
	}
	if !hasStyle && ((!italic && aspect.Style != font.StyleNormal) || (italic && aspect.Style != font.StyleItalic)) {
		return nil, false
	}
	return axes, true
}

func (r *Resolver) resolve(family string, bold, italic bool) (resolvedFont, error) {
	key := strings.ToLower(family) + "/" + styleName(bold, italic)
	if f, ok := r.loaded[key]; ok {
		return f, nil
	}
	var eligible []candidate
	for _, c := range r.candidates {
		if !strings.EqualFold(c.description.Family, family) {
			continue
		}
		if _, ok := selection(c, bold, italic); ok {
			eligible = append(eligible, c)
		}
	}
	if len(eligible) == 0 {
		return resolvedFont{}, fmt.Errorf("font family %q style %s not found; install the face or supply --font-dir (no font substitution)", family, styleName(bold, italic))
	}
	// Prefer an exact static style when both static and variable files exist;
	// path/index ordering makes otherwise equal candidates deterministic.
	sort.SliceStable(eligible, func(i, j int) bool { return len(eligible[i].axes) < len(eligible[j].axes) })
	c := eligible[0]
	raw, err := os.ReadFile(c.path)
	if err != nil {
		return resolvedFont{}, err
	}
	loaders, err := ot.NewLoaders(bytes.NewReader(raw))
	if err != nil {
		return resolvedFont{}, err
	}
	if c.index >= len(loaders) {
		return resolvedFont{}, fmt.Errorf("font collection changed during scan: %s", c.path)
	}
	actual, _ := font.Describe(loaders[c.index], nil)
	if actual != c.description {
		return resolvedFont{}, fmt.Errorf("font metadata changed during scan: %s", c.path)
	}
	ft, err := font.NewFont(loaders[c.index])
	if err != nil {
		return resolvedFont{}, err
	}
	face := font.NewFace(ft)
	axes, _ := selection(c, bold, italic)
	var variations []font.Variation
	for tag, value := range axes {
		variations = append(variations, font.Variation{Tag: ot.MustNewTag(tag), Value: value})
	}
	face.SetVariations(variations)
	selectedWeight := float32(actual.Aspect.Weight)
	if v, ok := axes["wght"]; ok {
		selectedWeight = v
	}
	f := resolvedFont{face, FontRecord{Family: actual.Family, Style: styleName(bold, italic), File: c.path, FaceIndex: c.index, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), Axes: axes, Weight: selectedWeight, Subfamily: c.subfamily}}
	r.loaded[key] = f
	return f, nil
}

func (r *Resolver) Records() []FontRecord {
	out := make([]FontRecord, 0, len(r.loaded))
	for _, f := range r.loaded {
		out = append(out, f.record)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Family != out[j].Family {
			return out[i].Family < out[j].Family
		}
		return out[i].Style < out[j].Style
	})
	return out
}
