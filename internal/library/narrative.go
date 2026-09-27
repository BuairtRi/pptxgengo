package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (n Narrative) Slide(id string) (NarrativeSlide, error) {
	for _, s := range n.Slides {
		if s.ID == id {
			return s, nil
		}
	}
	return NarrativeSlide{}, fmt.Errorf("narrative slide %s not found", id)
}
func (s Store) LoadNarrative(path string) (Narrative, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Narrative{}, err
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return Narrative{}, err
	}
	var n Narrative
	if err := json.Unmarshal(b, &n); err != nil {
		return n, err
	}
	if n.Schema != NarrativeSchema || n.Brief.Name == "" || n.Brief.Audience == "" || n.Brief.Decision == "" || len(n.Slides) == 0 {
		return n, fmt.Errorf("invalid narrative brief")
	}
	for _, a := range n.Brief.SourcePacket {
		if err := s.VerifyArtifact(a); err != nil {
			return n, err
		}
	}
	evidence := map[string]Evidence{}
	for _, e := range n.Evidence {
		if e.ID == "" || evidence[e.ID].ID != "" || e.Locator == "" {
			return n, fmt.Errorf("invalid/duplicate evidence locator %s", e.ID)
		}
		if err := s.VerifyArtifact(e.Source); err != nil {
			return n, err
		}
		if e.Quote != "" {
			p, err := s.SafePath(e.Source.Path)
			if err != nil {
				return n, err
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return n, err
			}
			if !strings.Contains(string(b), e.Quote) {
				return n, fmt.Errorf("evidence %s exact quote not found in source text", e.ID)
			}
		}
		evidence[e.ID] = e
	}
	claims := map[string]bool{}
	for _, c := range n.Claims {
		if c.ID == "" || claims[c.ID] || c.Assertion == "" {
			return n, fmt.Errorf("invalid/duplicate claim %s", c.ID)
		}
		claims[c.ID] = true
		switch c.Status {
		case "supported", "hypothesis", "synthetic", "unverified":
		default:
			return n, fmt.Errorf("invalid claim status %s", c.Status)
		}
		if c.Status == "synthetic" && !n.Brief.Synthetic {
			return n, fmt.Errorf("synthetic claim %s requires explicit synthetic brief", c.ID)
		}
		if c.Status == "supported" && len(c.EvidenceRefs) == 0 {
			return n, fmt.Errorf("supported claim %s lacks evidence", c.ID)
		}
		for _, id := range c.EvidenceRefs {
			if evidence[id].ID == "" {
				return n, fmt.Errorf("claim %s unknown evidence %s", c.ID, id)
			}
		}
	}
	seen := map[string]bool{}
	for _, sl := range n.Slides {
		if sl.ID == "" || seen[sl.ID] || sl.Role == "" || sl.Takeaway == "" || sl.AssertionTitle == "" || sl.Audience == "" || sl.VisualRelationship == "" {
			return n, fmt.Errorf("invalid narrative slide %s", sl.ID)
		}
		seen[sl.ID] = true
		for _, id := range sl.EvidenceRefs {
			if evidence[id].ID == "" {
				return n, fmt.Errorf("slide %s unknown evidence %s", sl.ID, id)
			}
		}
	}
	return n, nil
}
