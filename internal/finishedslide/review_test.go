package finishedslide

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func reviewFixture(t *testing.T) (string, Manifest, ReviewDecision) {
	t.Helper()
	_, m := fixture(t)
	files := map[string][]byte{m.Source: []byte("id: maintained-message\ncontent_kind: supplied_content\ntemplate: {scope: shared, id: cards/3}\nvalues: {}\n"), "preview.png": []byte("Owned structural fixture preview; no native approval"), "review.json": []byte(`{"fixture":"explicit test stewardship; not a production approval"}`)}
	m.Files = []File{{Path: m.Source, Role: "source"}, {Path: "preview.png", Role: "preview"}, {Path: "review.json", Role: "review"}}
	root := filepath.Join(t.TempDir(), "draft with spaces")
	m, e := Create(root, m, files)
	if e != nil {
		t.Fatal(e)
	}
	d := ReviewDecision{Schema: ReviewDecisionSchema, RevisionSHA256: m.RevisionSHA256, NewRevision: 2, Action: "approve", Actor: "fixture steward", Date: "2026-10-07", Reason: "Explicitly reviewed test artifacts; does not grant real content or native approval", ReuseScope: "Owned synthetic test invocation only", ValidUntil: "2026-11-07", SourceSHA256: sha(files[m.Source]), Preview: &ReviewedArtifact{"preview.png", sha(files["preview.png"])}, Report: &ReviewedArtifact{"review.json", sha(files["review.json"])}}
	return root, m, d
}

func TestReviewRevisionPreservesSourceAndHistoryAcrossLifecycle(t *testing.T) {
	root, original, decision := reviewFixture(t)
	draftRaw, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	decisionRaw, _ := json.MarshalIndent(decision, "", "  ")
	destination := filepath.Join(t.TempDir(), "approved")
	result, e := reviewRevisionAt(root, destination, decisionRaw, time.Date(2026, 10, 7, 1, 0, 0, 0, time.UTC))
	if e != nil {
		t.Fatal(e)
	}
	m, e := Read(destination)
	if e != nil || m.RevisionSHA256 != result.Manifest.RevisionSHA256 || m.Lifecycle != "approved" || m.Approval.By != decision.Actor || m.Approval.ReuseScope != decision.ReuseScope || m.ReviewedAt != decision.Date || m.ValidUntil != decision.ValidUntil || m.Pins != original.Pins {
		t.Fatal("approval metadata/pins wrong", e, m)
	}
	for _, f := range original.Files {
		a, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		b, _ := os.ReadFile(filepath.Join(destination, filepath.FromSlash(f.Path)))
		if !bytes.Equal(a, b) {
			t.Fatal("review changed payload", f.Path)
		}
	}
	retained, _ := os.ReadFile(filepath.Join(destination, filepath.FromSlash(result.Receipt.DecisionPath)))
	if !bytes.Equal(retained, decisionRaw) {
		t.Fatal("exact decision not retained")
	}
	var predecessor Manifest
	retained, _ = os.ReadFile(filepath.Join(destination, filepath.FromSlash(result.Receipt.PreviousManifestPath)))
	if e := json.Unmarshal(retained, &predecessor); e != nil || predecessor.RevisionSHA256 != original.RevisionSHA256 || predecessor.Validate() != nil {
		t.Fatal("predecessor proof absent", e)
	}
	retained, _ = os.ReadFile(filepath.Join(root, "manifest.json"))
	if !bytes.Equal(retained, draftRaw) {
		t.Fatal("draft modified")
	}
	if _, e := reviewRevisionAt(root, destination, decisionRaw, time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)); e == nil {
		t.Fatal("existing revision overwritten")
	}
	for index, action := range []string{"deprecate", "draft"} {
		d := ReviewDecision{Schema: ReviewDecisionSchema, RevisionSHA256: m.RevisionSHA256, SourceSHA256: decision.SourceSHA256, NewRevision: index + 3, Action: action, Actor: "fixture steward", Date: "2026-10-08", Reason: "Explicit fixture lifecycle change"}
		raw, _ := json.Marshal(d)
		next := filepath.Join(t.TempDir(), action)
		result, e = reviewRevisionAt(destination, next, raw, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC))
		if e != nil {
			t.Fatal(e)
		}
		m, e = Read(next)
		if e != nil || m.Lifecycle == "approved" || m.Approval != nil || m.ValidUntil != "" {
			t.Fatal("stale approval carried forward", e, m)
		}
		destination = next
	}
}

func TestReviewRevisionRefusesUnpinnedOrInvalidDecisionsBeforeWrites(t *testing.T) {
	for name, mutate := range map[string]func(*ReviewDecision){
		"revision hash":      func(d *ReviewDecision) { d.RevisionSHA256 = sha([]byte("wrong")) },
		"source hash":        func(d *ReviewDecision) { d.SourceSHA256 = sha([]byte("wrong")) },
		"preview hash":       func(d *ReviewDecision) { d.Preview.SHA256 = sha([]byte("wrong")) },
		"wrong role":         func(d *ReviewDecision) { d.Report = d.Preview },
		"actor absent":       func(d *ReviewDecision) { d.Actor = " " },
		"same revision":      func(d *ReviewDecision) { d.NewRevision = 1 },
		"reuse scope absent": func(d *ReviewDecision) { d.ReuseScope = "" },
		"expiry implicit":    func(d *ReviewDecision) { d.ValidUntil = "" },
		"expired":            func(d *ReviewDecision) { d.Date = "2026-09-01"; d.ValidUntil = "2026-09-02" },
		"future":             func(d *ReviewDecision) { d.Date = "2027-01-01"; d.ValidUntil = "2027-02-01" },
		"unknown action":     func(d *ReviewDecision) { d.Action = "accept" },
		"report missing":     func(d *ReviewDecision) { d.Report = nil },
		"traversal":          func(d *ReviewDecision) { d.Preview.Path = "../preview.png" },
	} {
		t.Run(name, func(t *testing.T) {
			root, _, d := reviewFixture(t)
			mutate(&d)
			raw, _ := json.Marshal(d)
			dest := filepath.Join(t.TempDir(), "new")
			if _, e := reviewRevisionAt(root, dest, raw, time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)); e == nil {
				t.Fatal("invalid review accepted")
			}
			if _, e := os.Lstat(dest); !os.IsNotExist(e) {
				t.Fatal("rejection created output")
			}
		})
	}
}

func TestReviewDecisionRejectsDuplicateUnknownAndMalformedJSON(t *testing.T) {
	_, _, d := reviewFixture(t)
	valid, _ := json.Marshal(d)
	for _, raw := range [][]byte{
		[]byte(strings.TrimSuffix(string(valid), "}") + `,"actor":"override"}`),
		[]byte(strings.TrimSuffix(string(valid), "}") + `,"unknown":true}`),
		append(append([]byte(nil), valid...), []byte(" {}")...),
		bytes.Replace(valid, []byte(`"new_revision":2`), []byte(`"new_revision":"2"`), 1),
		bytes.Replace(valid, []byte(`"path":"preview.png"`), []byte(`"path":"preview.png","path":"other"`), 1),
		[]byte("null"), []byte("[]"), bytes.Repeat([]byte(" "), 2<<20+1),
	} {
		if _, e := DecodeReviewDecision(raw); e == nil {
			t.Fatal("ambiguous/malformed decision accepted")
		}
	}
}

func TestReviewRevisionDetectsChangedReviewedPayload(t *testing.T) {
	root, _, d := reviewFixture(t)
	if e := os.WriteFile(filepath.Join(root, "preview.png"), []byte("Changed pixels"), 0600); e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(d)
	if _, e := ReviewRevision(root, filepath.Join(t.TempDir(), "new"), raw); e == nil {
		t.Fatal("changed preview approved")
	}
}

func TestReviewSourceRefusesSyntheticLocalAndAmbiguousPayloads(t *testing.T) {
	_, m, _ := reviewFixture(t)
	for _, raw := range []string{
		"id: sample\ncontent_kind: synthetic_example\ntemplate: {scope: shared, id: cards/3}\n",
		"id: sample\ncontent_kind: supplied_content\ntemplate: {scope: local, id: cards/3}\n",
		"id: sample\ncontent_kind: supplied_content\ntemplate: {scope: shared, id: cards/4}\n",
		"id: sample\ncontent_kind: synthetic_example\ncontent_kind: supplied_content\ntemplate: {scope: shared, id: cards/3}\n",
		"id: sample\ncontent_kind: &kind supplied_content\ntemplate: {scope: shared, id: cards/3}\n",
		"id: sample\ncontent_kind: supplied_content\ntemplate: {scope: shared, id: cards/3}\n---\nextra: true\n",
	} {
		if e := reviewedContentSource([]byte(raw), m.Pins); e == nil {
			t.Fatal("unsafe approval source accepted", raw)
		}
	}
}
