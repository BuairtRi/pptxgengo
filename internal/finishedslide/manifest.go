// Package finishedslide defines immutable, closed revisions of authored slides.
// It does not curate content, grant reuse approval, or mutate deck projects.
package finishedslide

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

const Schema = "pptxgengo.finished-slide.v1"
const Kind = "finished-slide"

type File struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type Pins struct {
	Bundle                   string `json:"bundle"`
	SourceRevision           string `json:"source_revision"`
	TemplateID               string `json:"template_id"`
	TemplateRevision         int    `json:"template_revision"`
	TemplateSourceSHA256     string `json:"template_source_sha256"`
	TemplateDefinitionSHA256 string `json:"template_definition_sha256"`
	ToolchainLockSHA256      string `json:"toolchain_lock_sha256"`
	Compiler                 string `json:"compiler"`
}

type Approval struct {
	By         string `json:"by"`
	Date       string `json:"date"`
	ReuseScope string `json:"reuse_scope"`
}

type Manifest struct {
	Schema         string    `json:"schema"`
	ID             string    `json:"id"`
	Revision       int       `json:"revision"`
	Name           string    `json:"name"`
	Purpose        string    `json:"purpose"`
	Owner          string    `json:"owner"`
	Lifecycle      string    `json:"lifecycle"`
	Approval       *Approval `json:"approval,omitempty"`
	ReviewedAt     string    `json:"reviewed_at,omitempty"`
	ValidUntil     string    `json:"valid_until,omitempty"`
	Keywords       []string  `json:"keywords,omitempty"`
	Source         string    `json:"source"`
	Pins           Pins      `json:"pins"`
	Files          []File    `json:"files"`
	RevisionSHA256 string    `json:"revision_sha256"`
}

var keyPattern = regexp.MustCompile(`^curated/slide/[a-z0-9][a-z0-9._-]*(/[a-z0-9][a-z0-9._-]*)*$`)

func sha(data []byte) string { sum := sha256.Sum256(data); return fmt.Sprintf("%x", sum) }
func validSHA(value string) bool {
	b, err := hex.DecodeString(value)
	return err == nil && len(b) == 32 && value == strings.ToLower(value)
}

func portablePath(value string) bool {
	if value == "" || len(value) > 240 || path.Clean(value) != value || strings.HasPrefix(value, "/") || strings.ContainsAny(value, "\\:\x00\r\n") {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." || strings.TrimRight(part, " .") != part {
			return false
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return false
		}
		for _, r := range part {
			if r < 32 || strings.ContainsRune(`<>"|?*`, r) {
				return false
			}
		}
	}
	return true
}

func (m Manifest) Digest() (string, error) {
	m.RevisionSHA256 = ""
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return sha(data), nil
}

func (m Manifest) Validate() error {
	if m.Schema != Schema || !keyPattern.MatchString(m.ID) || !portablePath(strings.TrimPrefix(m.ID, "curated/slide/")) || m.Revision < 1 {
		return fmt.Errorf("finished-slide.identity_invalid")
	}
	for _, field := range []string{m.Name, m.Purpose, m.Owner} {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("finished-slide.description_or_owner_missing")
		}
	}
	if m.Lifecycle != "draft" && m.Lifecycle != "approved" && m.Lifecycle != "deprecated" {
		return fmt.Errorf("finished-slide.lifecycle_invalid")
	}
	if m.Lifecycle == "approved" && m.Approval == nil {
		return fmt.Errorf("finished-slide.approval_missing")
	}
	if m.Approval != nil {
		if strings.TrimSpace(m.Approval.By) == "" || strings.TrimSpace(m.Approval.ReuseScope) == "" {
			return fmt.Errorf("finished-slide.approval_incomplete")
		}
		if _, err := time.Parse(time.DateOnly, m.Approval.Date); err != nil {
			return fmt.Errorf("finished-slide.approval_date_invalid")
		}
	}
	var reviewed time.Time
	if m.ReviewedAt != "" {
		var err error
		reviewed, err = time.Parse(time.DateOnly, m.ReviewedAt)
		if err != nil {
			return fmt.Errorf("finished-slide.review_date_invalid")
		}
	}
	if m.ValidUntil != "" {
		until, err := time.Parse(time.DateOnly, m.ValidUntil)
		if err != nil || reviewed.IsZero() || until.Before(reviewed) {
			return fmt.Errorf("finished-slide.freshness_dates_invalid")
		}
	}
	p := m.Pins
	if p.Bundle == "" || p.SourceRevision == "" || p.TemplateID == "" || p.TemplateRevision < 1 || p.Compiler == "" || !validSHA(p.TemplateSourceSHA256) || !validSHA(p.TemplateDefinitionSHA256) || !validSHA(p.ToolchainLockSHA256) {
		return fmt.Errorf("finished-slide.pins_incomplete")
	}
	if !portablePath(m.Source) || len(m.Files) == 0 || len(m.Files) > 256 {
		return fmt.Errorf("finished-slide.inventory_invalid")
	}
	names := map[string]bool{}
	directories := map[string]bool{"manifest.json": true}
	sourceCount := 0
	previewCount, reviewCount := 0, 0
	var total int64
	previous := ""
	for _, f := range m.Files {
		key := strings.ToLower(f.Path)
		if !portablePath(f.Path) || key == "manifest.json" || strings.HasPrefix(key, "manifest.json/") || names[key] || directories[key] || f.Path <= previous || f.Bytes < 0 || f.Bytes > 2<<30 || !validSHA(f.SHA256) {
			return fmt.Errorf("finished-slide.file_invalid: %s", f.Path)
		}
		for parent := path.Dir(key); parent != "."; parent = path.Dir(parent) {
			if names[parent] {
				return fmt.Errorf("finished-slide.file_directory_collision: %s", f.Path)
			}
			directories[parent] = true
		}
		total += f.Bytes
		if total > 2<<30 {
			return fmt.Errorf("finished-slide.package_too_large")
		}
		switch f.Role {
		case "source", "asset", "evidence", "preview", "review", "documentation":
		default:
			return fmt.Errorf("finished-slide.file_role_invalid: %s", f.Path)
		}
		if f.Role == "source" {
			sourceCount++
			if f.Path != m.Source {
				return fmt.Errorf("finished-slide.source_inventory_mismatch")
			}
		}
		if f.Role == "preview" {
			previewCount++
		}
		if f.Role == "review" {
			reviewCount++
		}
		names[key] = true
		previous = f.Path
	}
	if sourceCount != 1 {
		return fmt.Errorf("finished-slide.source_count_invalid")
	}
	if m.Lifecycle == "approved" && (previewCount == 0 || reviewCount == 0) {
		return fmt.Errorf("finished-slide.approved_artifacts_missing")
	}
	digest, err := m.Digest()
	if err != nil {
		return err
	}
	if m.RevisionSHA256 != digest {
		return fmt.Errorf("finished-slide.manifest_digest_mismatch")
	}
	return nil
}

func (m Manifest) Freshness(now time.Time) string {
	if m.ValidUntil == "" {
		return "no_expiry_declared"
	}
	until, err := time.Parse(time.DateOnly, m.ValidUntil)
	if err != nil {
		return "invalid"
	}
	now = now.UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if today.After(until) {
		return "stale"
	}
	return "within_declared_review_window"
}

// Seal sorts inventory and fills the revision digest before immutable creation.
// The caller supplies approval facts; this never upgrades lifecycle to approved.
func (m *Manifest) Seal() error {
	sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
	digest, err := m.Digest()
	if err != nil {
		return err
	}
	m.RevisionSHA256 = digest
	return m.Validate()
}
