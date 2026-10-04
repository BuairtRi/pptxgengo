package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type CompositionEntry struct {
	Purpose        string   `json:"purpose"`
	Relationship   string   `json:"relationship,omitempty"`
	Candidates     []string `json:"candidates,omitempty"`
	ChosenTemplate string   `json:"chosen_template"`
	Rationale      string   `json:"rationale"`
	Unresolved     []string `json:"unresolved,omitempty"`
}
type CompositionLog struct {
	Schema string                      `json:"schema"`
	Slides map[string]CompositionEntry `json:"slides"`
}
type Claim struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}
type ClaimsRegistry struct {
	Schema string  `json:"schema"`
	Claims []Claim `json:"claims"`
}

func compositionPath(p *Project) (string, error) {
	if path := p.Document.Context["composition_log"]; path != "" {
		return path, nil
	}
	path, err := SafePath(p.Root, "composition-log.yaml")
	if err != nil {
		return "", err
	}
	_, err = os.Lstat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return "composition-log.yaml", nil
}

func editorialValue(p *Project, relative string, target any) error {
	path, raw, err := editorialSource(p, relative)
	if err != nil {
		return err
	}
	parser := &Project{SourcePath: path, Positions: map[string]Position{}, positionFiles: map[string]string{}}
	value, err := parser.parseSource(raw, relative, "")
	if err != nil {
		return err
	}
	if err = parser.shapeType(value, reflect.TypeOf(target).Elem(), ""); err != nil {
		return err
	}
	return strictInto(value, target)
}

func editorialSource(p *Project, relative string) (string, []byte, error) {
	path, err := SafePath(p.Root, relative)
	if err != nil {
		return "", nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16<<20 {
		return "", nil, fmt.Errorf("%s: expected regular file <=16MiB", relative)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	return path, raw, nil
}

func ReadCompositionLog(p *Project) (map[string]CompositionEntry, error) {
	relative, err := compositionPath(p)
	if err != nil {
		return nil, err
	}
	if relative == "" {
		return nil, nil
	}
	var log CompositionLog
	if err = editorialValue(p, relative, &log); err != nil {
		return nil, err
	}
	if log.Schema != "pptxgengo.composition-log.v1" {
		return nil, fmt.Errorf("%s: unsupported composition schema", relative)
	}
	known := map[string]bool{}
	for _, slide := range p.Document.Slides {
		known[slide.ID] = true
		entry, ok := log.Slides[slide.ID]
		if !ok {
			return nil, fmt.Errorf("composition log missing slide %s", slide.ID)
		}
		if strings.TrimSpace(entry.Purpose) == "" || strings.TrimSpace(entry.Rationale) == "" || entry.ChosenTemplate == "" {
			return nil, fmt.Errorf("composition log %s requires purpose, rationale and chosen_template", slide.ID)
		}
		if entry.ChosenTemplate != slide.Template.ID {
			return nil, fmt.Errorf("composition log %s chosen_template differs from source", slide.ID)
		}
	}
	for id := range log.Slides {
		if !known[id] {
			return nil, fmt.Errorf("composition log references unknown slide %s", id)
		}
	}
	return log.Slides, nil
}

// Markdown registries use explicit headings (## claim-id) or {#claim-id}
// anchors. Plain prose is not interpreted as evidence identity.
var claimAnchor = regexp.MustCompile(`\{#([A-Za-z0-9][A-Za-z0-9._-]{0,119})\}`)

func ValidateEditorial(p *Project) error {
	if _, err := ReadCompositionLog(p); err != nil {
		return err
	}
	relative := p.Document.Context["claims"]
	if relative == "" {
		return nil
	}
	ids := map[string]bool{}
	add := func(id string) error {
		if !stableID.MatchString(id) || ids[id] {
			return fmt.Errorf("claims registry invalid or duplicate ID %s", id)
		}
		ids[id] = true
		return nil
	}
	if strings.ToLower(filepath.Ext(relative)) == ".md" {
		_, raw, err := editorialSource(p, relative)
		if err != nil {
			return err
		}
		for _, line := range markdownClaimLines(string(raw)) {
			matches := claimAnchor.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				if err := add(m[1]); err != nil {
					return err
				}
			}
			if len(matches) == 0 && strings.HasPrefix(line, "## ") {
				id := strings.TrimSpace(strings.TrimPrefix(line, "## "))
				if stableID.MatchString(id) {
					if err := add(id); err != nil {
						return err
					}
				}
			}
		}
	} else {
		var registry ClaimsRegistry
		if err := editorialValue(p, relative, &registry); err != nil {
			return err
		}
		if registry.Schema != "pptxgengo.claims.v1" {
			return fmt.Errorf("unsupported claims schema")
		}
		for _, claim := range registry.Claims {
			if err := add(claim.ID); err != nil {
				return err
			}
			if strings.TrimSpace(claim.Text) == "" {
				return fmt.Errorf("claim %s requires text", claim.ID)
			}
		}
	}
	for _, slide := range p.Document.Slides {
		for _, id := range slide.EvidenceRefs {
			if !ids[id] {
				return fmt.Errorf("slide %s evidence_refs ID %s not found in %s", slide.ID, id, relative)
			}
		}
	}
	return nil
}

// Examples in code are not claim identities. This intentionally recognizes a
// small registry format rather than interpreting arbitrary Markdown as proof.
func markdownClaimLines(raw string) []string {
	lines := []string{}
	fence := byte(0)
	fenceLength := 0
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		run := 0
		if indent <= 3 && len(trimmed) > 0 && (trimmed[0] == '`' || trimmed[0] == '~') {
			for run < len(trimmed) && trimmed[run] == trimmed[0] {
				run++
			}
		}
		if fence != 0 {
			if run >= fenceLength && trimmed[0] == fence && strings.TrimSpace(trimmed[run:]) == "" {
				fence = 0
			}
			continue
		}
		if run >= 3 {
			fence, fenceLength = trimmed[0], run
			continue
		}
		if indent >= 4 || strings.HasPrefix(line, "\t") {
			continue
		}
		var visible strings.Builder
		for i := 0; i < len(line); {
			if line[i] != '`' {
				visible.WriteByte(line[i])
				i++
				continue
			}
			start := i
			for i < len(line) && line[i] == '`' {
				i++
			}
			marker := line[start:i]
			end := strings.Index(line[i:], marker)
			if end < 0 {
				visible.WriteString(marker)
			} else {
				i += end + len(marker)
			}
		}
		lines = append(lines, visible.String())
	}
	return lines
}
