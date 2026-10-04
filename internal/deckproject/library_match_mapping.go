package deckproject

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// A recipe may bridge only named fields from the pinned source topology. A
// caller's phrasing never supplies evidence about what an arbitrary box means.
func recipePagePath(page PageSpec, metadata wmdesign.LibraryAuthoring, slot wmdesign.LibraryAuthoringSlot) string {
	switch metadata.Recipe {
	case "lifecycle-three-phases-source-topology.v1":
		if slot.Group != "/phases" || slot.GroupIndex < 0 {
			return ""
		}
		prefix := fmt.Sprintf("/items/%d", slot.GroupIndex)
		parts := strings.Split(slot.Alias, "/")
		if len(parts) == 4 {
			switch parts[3] {
			case "title":
				return prefix + "/lead"
			case "objective":
				if slot.GroupIndex < len(page.Items) && page.Items[slot.GroupIndex].Objective == "" {
					return prefix + "/text"
				}
				return prefix + "/objective"
			}
		}
		if len(parts) == 6 && parts[3] == "activities" {
			index, e := strconv.Atoi(strings.TrimPrefix(parts[4], "item_"))
			if e == nil && index > 0 {
				field := "text"
				if parts[5] == "title" {
					field = "lead"
				}
				return fmt.Sprintf("%s/activities/%d/%s", prefix, index-1, field)
			}
		}
	case "vendors-scorecard-generic-source-topology.v1":
		if page.Comparison != nil && strings.HasPrefix(slot.Alias, "/comparison/") {
			parts := strings.Split(slot.Alias, "/")
			for i, part := range parts {
				if strings.HasPrefix(part, "item_") {
					index, e := strconv.Atoi(strings.TrimPrefix(part, "item_"))
					if e != nil || index < 1 {
						return ""
					}
					parts[i] = strconv.Itoa(index - 1)
				}
			}
			return strings.Join(parts, "/")
		}
	}
	return ""
}

func rankMatchCandidates(candidates []MatchCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		level := func(c MatchCandidate) int {
			if c.Status == "go_layout_succeeded_native_review_pending" {
				return 3
			}
			if c.MappingComplete {
				return 2
			}
			if c.Status == "needs_copy" {
				return 1
			}
			return 0
		}
		if level(a) != level(b) {
			return level(a) > level(b)
		}
		mapped := func(c MatchCandidate) int { return len(c.Disposition) - len(c.UnresolvedFields) }
		if mapped(a) != mapped(b) {
			return mapped(a) > mapped(b)
		}
		if len(a.UnresolvedFields) != len(b.UnresolvedFields) {
			return len(a.UnresolvedFields) < len(b.UnresolvedFields)
		}
		return len(a.MissingSlots) < len(b.MissingSlots)
	})
	rank := 0
	for i := range candidates {
		if candidates[i].Status == "needs_copy" {
			rank++
			candidates[i].NearMissRank = rank
		}
	}
}
