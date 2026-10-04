package wmdesign

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SectionSpec anchors a contiguous native PowerPoint group. The next anchor
// closes the preceding group; slide IDs remain authoritative after reordering.
type SectionSpec struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	BeforeSlideID string `json:"before_slide_id"`
}

type SectionValidationError struct {
	Index  int
	Field  string
	Reason string
}

func (e *SectionValidationError) Error() string {
	return fmt.Sprintf("document.sections/%d/%s: %s", e.Index, e.Field, e.Reason)
}

var sectionStableID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,119}$`)

// ValidateSections applies the same closed anchor contract to maintained YAML
// and direct scene documents. No sections retains the historical slide model.
func ValidateSections(sections []SectionSpec, slideIDs []string) error {
	if len(sections) == 0 {
		return nil
	}
	positions := map[string]int{}
	for i, id := range slideIDs {
		if _, ok := positions[id]; ok || id == "" {
			return fmt.Errorf("slide.invalid_or_duplicate_id")
		}
		positions[id] = i
	}
	ids, titles := map[string]bool{}, map[string]bool{}
	last := -1
	for i, section := range sections {
		fail := func(field, reason string) error { return &SectionValidationError{i, field, reason} }
		if !sectionStableID.MatchString(section.ID) || ids[section.ID] {
			return fail("id", "invalid or duplicate stable section ID")
		}
		ids[section.ID] = true
		key := strings.ToLower(section.Title)
		if strings.TrimSpace(section.Title) != section.Title || section.Title == "" || utf8.RuneCountInString(section.Title) > 160 || titles[key] {
			return fail("title", "requires a unique trimmed title of 1 to 160 characters")
		}
		for _, char := range section.Title {
			if unicode.IsControl(char) || char == 0xfffe || char == 0xffff || !utf8.ValidString(section.Title) {
				return fail("title", "control characters are unsupported")
			}
		}
		titles[key] = true
		position, ok := positions[section.BeforeSlideID]
		if !ok {
			return fail("before_slide_id", "anchor must name an existing slide ID")
		}
		if i == 0 && position != 0 {
			return fail("before_slide_id", "first section must anchor the first slide")
		}
		if position <= last {
			return fail("before_slide_id", "anchors must be strictly ordered; empty sections are unsupported")
		}
		last = position
	}
	return nil
}

// ValidateSpeakerNotes checks optional authored text before it reaches native XML.
func ValidateSpeakerNotes(s string) error {
	if len(s) > 1048576 || !utf8.ValidString(s) {
		return fmt.Errorf("notes must be valid XML text, at most 1 MiB")
	}
	for _, r := range s {
		if r != '\t' && r != '\n' && r != '\r' && (r < 0x20 || r == 0xfffe || r == 0xffff) {
			return fmt.Errorf("notes contain an XML-invalid character")
		}
	}
	return nil
}
