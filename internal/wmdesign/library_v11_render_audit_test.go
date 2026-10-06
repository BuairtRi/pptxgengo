package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Collect every affected source key instead of stopping at the first changed
// slide. Each visible relationship graph is compared independently, including
// images, charts, workbooks, layouts and themes. Reports retain fit decisions.
func TestLibraryV11UnchangedSourceRenderAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("paired audit of every unchanged source specimen")
	}
	keys := v11TemplateKeys(t, "unchanged", 418)
	previous, err := Load(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	current, err := Load(v11IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !compatiblePublicationStyle(previous, current) {
		t.Fatal("unaccepted source style dependency change")
	}
	oldSlides, err := publicationSlides(v10IntakeBundle(), "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	newSlides, err := publicationSlides(v11IntakeBundle(), "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	before := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "library-v11-source-render-audit"}}
	after := before
	for i, key := range keys {
		left, ok1 := oldSlides[key]
		right, ok2 := newSlides[key]
		if !ok1 || !ok2 || !samePublicationComposition(left, right) {
			t.Fatalf("unchanged composition invalid: %s", key)
		}
		left.ID = fmt.Sprintf("audit-%03d", i+1)
		right.ID = left.ID
		before.Slides = append(before.Slides, left)
		after.Slides = append(after.Slides, right)
	}
	leftDeck, leftReport, err := BuildWithEngine(v10IntakeBundle(), "", before, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	rightDeck, rightReport, err := BuildWithEngine(v11IntakeBundle(), "", after, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	left, err := publicationDeckParts(leftDeck)
	if err != nil {
		t.Fatal(err)
	}
	right, err := publicationDeckParts(rightDeck)
	if err != nil {
		t.Fatal(err)
	}
	type entry struct {
		Page       int    `json:"page"`
		Key        string `json:"key"`
		Equivalent bool   `json:"equivalent"`
		Difference string `json:"difference,omitempty"`
	}
	var entries []entry
	var changed []string
	for i, key := range keys {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		err := v11CompareVisibleDependency(left, right, part, part, map[string]bool{})
		item := entry{Page: i + 1, Key: key, Equivalent: err == nil}
		if err != nil {
			item.Difference = err.Error()
			changed = append(changed, key)
		}
		entries = append(entries, item)
	}
	dir, err := os.MkdirTemp("/private/tmp", "wmds-v11-unchanged-render-audit-")
	if err != nil {
		t.Fatal(err)
	}
	writeJSON := func(name string, value any) {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range map[string][]byte{fmt.Sprintf("before%d.pptx", len(keys)): leftDeck, fmt.Sprintf("after%d.pptx", len(keys)): rightDeck} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeJSON("before-report.json", leftReport)
	writeJSON("after-report.json", rightReport)
	writeJSON("audit.json", map[string]any{"schema": "pptxgengo.v11-unchanged-source-render-audit.v1", "source_commit": current.Commit, "previous_commit": previous.Commit, "specimens": len(keys), "equivalent": len(keys) - len(changed), "changed_keys": changed, "pages": entries, "qualification_scope": "paired visible dependencies only; changed specimens require native review"})
	t.Logf("complete render audit: %s", dir)
	// Compare a finalized partition only with the source that established it.
	// A newly approved candidate first collects its complete actual delta; it
	// cannot reuse the preceding source's qualification partition.
	var finalized struct {
		Commit  string   `json:"source_commit"`
		Changed []string `json:"changed_keys"`
	}
	v11ReadJSON(t, filepath.Join(filepath.Dir(v11IntakeBundle()), "final-unchanged-render-audit.json"), &finalized)
	if finalized.Commit == current.Commit {
		expected := v11TemplateKeys(t, "additional-render-review", len(finalized.Changed))
		if !reflect.DeepEqual(changed, expected) {
			t.Errorf("unexpected additional render delta: got %v; want %v", changed, expected)
		}
	} else {
		t.Log("new unqualified source delta collected; regenerate its review partition before qualification")
	}
}

func v11CompareVisibleDependency(left, right map[string][]byte, lp, rp string, checked map[string]bool) error {
	key := lp + "\x00" + rp
	if checked[key] {
		return nil
	}
	checked[key] = true
	l, lok := left[lp]
	r, rok := right[rp]
	if !lok || !rok {
		return fmt.Errorf("visible_part_missing: %s / %s", lp, rp)
	}
	if !bytes.Equal(l, r) {
		return fmt.Errorf("visible_part_changed: %s", lp)
	}
	lr, err := publicationVisibleRelationships(left, lp)
	if err != nil {
		return err
	}
	rr, err := publicationVisibleRelationships(right, rp)
	if err != nil {
		return err
	}
	if len(lr) != len(rr) {
		return fmt.Errorf("visible_relationship_count_changed: %s", lp)
	}
	for i, lrel := range lr {
		rrel := rr[i]
		if lrel.ID != rrel.ID || lrel.Type != rrel.Type || lrel.Mode != rrel.Mode {
			return fmt.Errorf("visible_relationship_changed: %s", lp)
		}
		if lrel.Mode == "External" {
			if lrel.Target != rrel.Target {
				return fmt.Errorf("external_relationship_changed: %s", lp)
			}
			continue
		}
		if err := v11CompareVisibleDependency(left, right, publicationRelationshipPart(lp, lrel.Target), publicationRelationshipPart(rp, rrel.Target), checked); err != nil {
			return err
		}
	}
	return nil
}
