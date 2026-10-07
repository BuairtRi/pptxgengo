package wmdesign

import (
	"fmt"
	"sort"
	"time"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

type BrowsingRevision struct {
	Path     string                 `json:"path"`
	Manifest finishedslide.Manifest `json:"manifest"`
	Included bool                   `json:"included"`
	Reason   string                 `json:"reason"`
}
type BrowsingSelection struct {
	Schema        string             `json:"schema"`
	AsOf          string             `json:"as_of"`
	LibrarySHA256 string             `json:"library_sha256"`
	Revisions     []BrowsingRevision `json:"revisions"`
	Policy        string             `json:"policy"`
}

// SelectBrowsingRevisions verifies every immutable revision, then selects latest
// approvals. A newer draft does not erase an approval; a newer deprecation does.
// An expired latest approval never falls back to an older approval.
func SelectBrowsingRevisions(root string, asOf time.Time) (BrowsingSelection, error) {
	result := BrowsingSelection{Schema: "pptxgengo.reusable-browsing-selection.v1", AsOf: asOf.UTC().Format(time.DateOnly), Policy: "Latest approved revision per identity; newer deprecation withdraws earlier approvals; newer drafts remain omitted; expired latest approval does not resurrect older revisions; approval scope belongs to the recorded steward, not CI."}
	revisions, hash, e := readFinishedLibrary(root)
	if e != nil {
		return result, e
	}
	result.LibrarySHA256 = hash
	approved := map[string]int{}
	withdrawn := map[string]int{}
	seen := map[string]bool{}
	for _, r := range revisions {
		m := r.manifest
		key := fmt.Sprintf("%s@%d", m.ID, m.Revision)
		if seen[key] {
			return result, fmt.Errorf("browsing.duplicate_revision: %s", key)
		}
		seen[key] = true
		if e := validateFinishedIndexSource(root, r); e != nil {
			return result, e
		}
		if m.Lifecycle == "approved" && m.Revision > approved[m.ID] {
			approved[m.ID] = m.Revision
		}
		if m.Lifecycle == "deprecated" && m.Revision > withdrawn[m.ID] {
			withdrawn[m.ID] = m.Revision
		}
	}
	count := 0
	for _, r := range revisions {
		m := r.manifest
		entry := BrowsingRevision{Path: r.path, Manifest: m, Reason: "not_latest_approved"}
		switch {
		case m.Lifecycle == "draft":
			entry.Reason = "draft"
		case m.Lifecycle == "deprecated":
			entry.Reason = "deprecated"
		case withdrawn[m.ID] > m.Revision:
			entry.Reason = "withdrawn_by_newer_revision"
		case m.Revision != approved[m.ID]:
		case m.Freshness(asOf) == "stale":
			entry.Reason = "expired"
		default:
			if m.Approval == nil {
				return result, fmt.Errorf("browsing.approval_missing")
			}
			date, _ := time.Parse(time.DateOnly, m.Approval.Date)
			if date.After(asOf) {
				entry.Reason = "approval_after_as_of"
			} else {
				entry.Included = true
				entry.Reason = "latest_approved"
				count++
			}
		}
		result.Revisions = append(result.Revisions, entry)
	}
	sort.Slice(result.Revisions, func(i, j int) bool {
		a, b := result.Revisions[i].Manifest, result.Revisions[j].Manifest
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		return a.Revision < b.Revision
	})
	if count == 0 {
		return result, fmt.Errorf("browsing.no_current_approved_slides: provide operator-approved nonexpired immutable revisions; CI cannot invent reusable content")
	}
	return result, nil
}
