// Package browsingartifact validates the installed browsing deck inventories.
// Integrity and declared approval metadata are not publisher authenticity or
// native PowerPoint visual qualification.
package browsingartifact

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

type FilePin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type Frame struct {
	Rail          string `json:"rail"`
	Footer        string `json:"footer"`
	Surface       string `json:"surface"`
	RailSurface   string `json:"rail_surface"`
	TitleLines    int    `json:"title_lines"`
	Density       string `json:"density"`
	HeaderDensity string `json:"header_density"`
	SourceLines   int    `json:"source_lines"`
	NoHeader      bool   `json:"no_header"`
	NoPage        bool   `json:"no_page"`
	Split         string `json:"split"`
	Active        string `json:"active"`
	Nav           []struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	} `json:"nav"`
}
type Entry struct {
	SlideID      string `json:"slide_id"`
	Kind         string `json:"kind"`
	Key          string `json:"key"`
	Revision     int    `json:"revision"`
	Lifecycle    string `json:"lifecycle"`
	SourceSHA256 string `json:"source_sha256"`
	Frame        *Frame `json:"frame"`
}
type TemplateCoverage struct {
	Schema                string    `json:"schema"`
	FrameMode             string    `json:"frame_mode"`
	SourceRevision        string    `json:"source_revision"`
	SourceCommit          string    `json:"source_commit"`
	SourceFiles           []FilePin `json:"source_files"`
	ExpectedTemplates     int       `json:"expected_templates"`
	ExpectedFrameRequests int       `json:"expected_frame_requests"`
	FrameCandidates       int       `json:"frame_candidates"`
	Entries               []Entry   `json:"entries"`
	Aliases               []struct {
		Frame   Frame  `json:"frame"`
		SlideID string `json:"slide_id"`
		Basis   string `json:"basis"`
	} `json:"frame_aliases"`
	Exclusions []struct {
		Reason string `json:"reason"`
		Frame  Frame  `json:"frame"`
	} `json:"frame_exclusions"`
}
type Revision struct {
	Path     string                 `json:"path"`
	Manifest finishedslide.Manifest `json:"manifest"`
	Included bool                   `json:"included"`
	Reason   string                 `json:"reason"`
}
type Selection struct {
	Schema        string     `json:"schema"`
	AsOf          string     `json:"as_of"`
	LibrarySHA256 string     `json:"library_sha256"`
	Revisions     []Revision `json:"revisions"`
}
type Page struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}
type Manifest struct {
	Pages           []Page          `json:"pages"`
	BundleSHA256    string          `json:"bundle_sha256"`
	SourceRevision  string          `json:"source_revision"`
	SourceCommit    string          `json:"source_commit"`
	Schema          string          `json:"schema"`
	Kind            string          `json:"kind"`
	AsOf            string          `json:"as_of"`
	DeckSHA256      string          `json:"deck_sha256"`
	Compiler        string          `json:"compiler"`
	ReleaseIdentity string          `json:"release_identity"`
	Slides          int             `json:"slides"`
	Coverage        json.RawMessage `json:"coverage"`
	SourceFiles     []FilePin       `json:"source_files"`
	Assets          []FilePin       `json:"assets"`
	Fonts           []struct {
		File       string `json:"file"`
		SHA256     string `json:"sha256"`
		PostScript string `json:"postscript_name"`
	} `json:"fonts"`
	Qualification string `json:"qualification"`
}

func hash(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && strings.ToLower(s) == s
}
func portable(s string) bool {
	if s == "" || path.Clean(s) != s || strings.HasPrefix(s, "/") || strings.ContainsAny(s, "\\:\x00\r\n") {
		return false
	}
	for _, p := range strings.Split(s, "/") {
		if strings.ContainsAny(p, `<>"|?*`) {
			return false
		}
		base := strings.ToUpper(strings.SplitN(p, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return false
		}
		if p == "" || p == "." || p == ".." || strings.TrimRight(p, " .") != p {
			return false
		}
	}
	return true
}
func pins(files []FilePin) bool {
	if len(files) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, f := range files {
		if !portable(f.Path) || !hash(f.SHA256) || seen[strings.ToLower(f.Path)] {
			return false
		}
		seen[strings.ToLower(f.Path)] = true
	}
	return true
}

func Validate(raw []byte, kind, deckHash string, actualSlides int) error {
	if len(raw) > MaxManifestBytes {
		return fmt.Errorf("browsing.manifest_exceeds_16MiB")
	}
	var m Manifest
	if e := json.Unmarshal(raw, &m); e != nil {
		return e
	}
	date, e := time.Parse(time.DateOnly, m.AsOf)
	if e != nil {
		return fmt.Errorf("browsing.as_of_invalid")
	}
	if m.Schema != "pptxgengo.browsing-library.v1" || m.Kind != kind || !hash(m.DeckSHA256) || m.DeckSHA256 != deckHash || m.Slides != actualSlides || len(m.Pages) != m.Slides || !hash(m.BundleSHA256) || m.SourceRevision == "" || m.SourceCommit == "" || m.Compiler == "" || m.ReleaseIdentity == "" || m.Qualification != "native_visual_copy_paste_qualification_pending" || !pins(m.SourceFiles) || len(m.Fonts) == 0 {
		return fmt.Errorf("browsing.manifest_or_input_pins_invalid")
	}
	for _, f := range m.Fonts {
		if !portable(f.File) || strings.Contains(f.File, "/") || !hash(f.SHA256) || f.PostScript == "" {
			return fmt.Errorf("browsing.font_pin_invalid")
		}
	}
	for _, a := range m.Assets {
		if !hash(a.SHA256) || a.Path == "" {
			return fmt.Errorf("browsing.asset_pin_invalid")
		}
	}
	if len(m.Pages) < 2 || m.Pages[0].Kind != "guide" {
		return fmt.Errorf("browsing.guide_missing")
	}
	ids := map[string]bool{}
	for _, page := range m.Pages {
		if page.ID == "" || ids[page.ID] {
			return fmt.Errorf("browsing.page_inventory_invalid")
		}
		ids[page.ID] = true
		switch page.Kind {
		case "guide", "family_divider", "deprecation_notice", "frame_divider", "template", "frame", "reuse_metadata", "reusable_slide":
		default:
			return fmt.Errorf("browsing.page_kind_invalid")
		}
	}
	switch kind {
	case "templates":
		return templates(m.Coverage, m)
	case "reusable":
		return reusable(m.Coverage, date, m.Slides, m)
	default:
		return fmt.Errorf("browsing.kind_invalid")
	}
}
func templates(raw []byte, m Manifest) error {
	var c TemplateCoverage
	if e := json.Unmarshal(raw, &c); e != nil {
		return e
	}
	if c.Schema != "pptxgengo.browsing-coverage.v1" || c.SourceRevision != m.SourceRevision || c.SourceCommit != m.SourceCommit || !pins(c.SourceFiles) || !reflect.DeepEqual(c.SourceFiles, m.SourceFiles) || c.ExpectedTemplates < 1 || c.ExpectedFrameRequests < 1 || (c.FrameMode != "catalog" && c.FrameMode != "exhaustive") {
		return fmt.Errorf("browsing.coverage_source_or_counts_invalid")
	}
	pageRoles := map[string]string{}
	guides, frameDividers := 0, 0
	for _, page := range m.Pages {
		pageRoles[page.ID] = page.Kind
		switch page.Kind {
		case "guide":
			guides++
		case "frame_divider":
			frameDividers++
		case "reuse_metadata", "reusable_slide":
			return fmt.Errorf("browsing.template_deck_reusable_role_invalid")
		}
	}
	if guides != 1 || frameDividers != 1 {
		return fmt.Errorf("browsing.template_guide_or_frame_divider_missing")
	}
	ids := map[string]Entry{}
	keys := map[string]bool{}
	requests := map[string]bool{}
	recordRequest := func(frame Frame) error {
		raw, e := json.Marshal(frame)
		if e != nil {
			return e
		}
		if requests[string(raw)] {
			return fmt.Errorf("browsing.frame_request_duplicate")
		}
		requests[string(raw)] = true
		return nil
	}
	templates, frames := 0, 0
	sourceHashes := map[string]bool{}
	for _, pin := range c.SourceFiles {
		sourceHashes[pin.SHA256] = true
	}
	for _, entry := range c.Entries {
		if entry.SlideID == "" || entry.Key == "" || pageRoles[entry.SlideID] != entry.Kind {
			return fmt.Errorf("browsing.coverage_identity_missing")
		}
		if _, ok := ids[entry.SlideID]; ok || keys[entry.Kind+"/"+entry.Key] {
			return fmt.Errorf("browsing.coverage_duplicate")
		}
		ids[entry.SlideID] = entry
		keys[entry.Kind+"/"+entry.Key] = true
		switch entry.Kind {
		case "template":
			templates++
			if entry.Revision < 1 || !hash(entry.SourceSHA256) || !sourceHashes[entry.SourceSHA256] || (entry.Lifecycle != "active" && entry.Lifecycle != "deprecated") {
				return fmt.Errorf("browsing.template_pin_invalid")
			}
		case "frame":
			frames++
			if entry.Frame == nil || entry.Frame.Rail == "" || entry.Frame.Footer == "" || !hash(entry.SourceSHA256) || !sourceHashes[entry.SourceSHA256] {
				return fmt.Errorf("browsing.frame_pin_invalid")
			}
			if e := recordRequest(*entry.Frame); e != nil {
				return e
			}
		default:
			return fmt.Errorf("browsing.coverage_kind_invalid")
		}
	}
	for _, page := range m.Pages {
		if page.Kind == "template" || page.Kind == "frame" {
			if _, ok := ids[page.ID]; !ok {
				return fmt.Errorf("browsing.page_coverage_omission")
			}
		}
	}
	if templates != c.ExpectedTemplates || frames+len(c.Aliases) != c.ExpectedFrameRequests || c.FrameCandidates != c.ExpectedFrameRequests+len(c.Exclusions) {
		return fmt.Errorf("browsing.coverage_omission")
	}
	for _, alias := range c.Aliases {
		if e := recordRequest(alias.Frame); e != nil {
			return e
		}
		entry, ok := ids[alias.SlideID]
		if !ok || entry.Kind != "frame" || entry.Frame == nil || !entry.Frame.NoHeader || !alias.Frame.NoHeader || alias.Basis == "" {
			return fmt.Errorf("browsing.alias_invalid")
		}
		normalized := alias.Frame
		normalized.TitleLines = 1
		normalized.Density = "standard"
		if !reflect.DeepEqual(normalized, *entry.Frame) {
			return fmt.Errorf("browsing.alias_changed_active_request")
		}
	}
	for _, x := range c.Exclusions {
		if e := recordRequest(x.Frame); e != nil {
			return e
		}
		switch x.Reason {
		case "frame.split_requires_no_panel_rail", "frame.split_requires_standard_header", "frame.appendix_requires_one_title_line", "frame.invalid_line_allocation", "frame.empty_body":
		default:
			return fmt.Errorf("browsing.nonstructural_exclusion")
		}
	}
	if c.FrameMode == "catalog" && (len(c.Aliases) != 0 || len(c.Exclusions) != 0) {
		return fmt.Errorf("browsing.catalog_frame_omission")
	}
	return nil
}
func reusable(raw []byte, date time.Time, slides int, current Manifest) error {
	var c Selection
	if e := json.Unmarshal(raw, &c); e != nil {
		return e
	}
	if c.Schema != "pptxgengo.reusable-browsing-selection.v1" || c.AsOf != date.Format(time.DateOnly) || !hash(c.LibrarySHA256) || len(c.Revisions) == 0 {
		return fmt.Errorf("browsing.selection_pins_invalid")
	}
	latest := map[string]int{}
	withdrawn := map[string]int{}
	seen := map[string]bool{}
	for _, r := range c.Revisions {
		m := r.Manifest
		if !portable(r.Path) {
			return fmt.Errorf("browsing.revision_path_invalid")
		}
		if e := m.Validate(); e != nil {
			return e
		}
		key := fmt.Sprintf("%s@%d", m.ID, m.Revision)
		if seen[key] {
			return fmt.Errorf("browsing.revision_duplicate")
		}
		seen[key] = true
		if m.Lifecycle == "approved" && m.Revision > latest[m.ID] {
			latest[m.ID] = m.Revision
		}
		if m.Lifecycle == "deprecated" && m.Revision > withdrawn[m.ID] {
			withdrawn[m.ID] = m.Revision
		}
	}
	pageCounts := map[string]int{}
	for _, p := range current.Pages {
		pageCounts[p.Kind]++
	}
	if pageCounts["guide"] != 1 || len(pageCounts) != 3 {
		return fmt.Errorf("browsing.reusable_page_roles_invalid")
	}
	selected := 0
	for _, r := range c.Revisions {
		m := r.Manifest
		eligible := m.Lifecycle == "approved" && m.Revision == latest[m.ID] && m.Revision > withdrawn[m.ID] && m.Freshness(date) != "stale"
		if eligible {
			d, e := time.Parse(time.DateOnly, m.Approval.Date)
			eligible = e == nil && !d.After(date)
		}
		if eligible != r.Included {
			return fmt.Errorf("browsing.approval_selection_mismatch")
		}
		if r.Included {
			if r.Reason != "latest_approved" {
				return fmt.Errorf("browsing.approval_reason_invalid")
			}
			if m.Pins.Bundle != current.BundleSHA256 || m.Pins.SourceRevision != current.SourceRevision {
				return fmt.Errorf("browsing.approved_bundle_pin_mismatch")
			}
			pinned := false
			for _, file := range current.SourceFiles {
				pinned = pinned || file.SHA256 == m.Pins.TemplateSourceSHA256
			}
			if !pinned {
				return fmt.Errorf("browsing.approved_template_source_pin_mismatch")
			}
			selected++
		}
	}
	if selected < 1 || slides != 1+2*selected || pageCounts["reuse_metadata"] != selected || pageCounts["reusable_slide"] != selected {
		return fmt.Errorf("browsing.approved_slides_or_metadata_missing")
	}
	return nil
}
