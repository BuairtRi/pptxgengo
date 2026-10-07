package finishedslide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const ReviewDecisionSchema = "pptxgengo.finished-slide-review.v1"
const ReviewReceiptSchema = "pptxgengo.finished-slide-review-receipt.v1"

type ReviewedArtifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type ReviewDecision struct {
	Schema         string            `json:"schema"`
	RevisionSHA256 string            `json:"revision_sha256"`
	NewRevision    int               `json:"new_revision"`
	Action         string            `json:"action"` // approve, deprecate, draft
	Actor          string            `json:"actor"`
	Date           string            `json:"date"`
	Reason         string            `json:"reason"`
	ReuseScope     string            `json:"reuse_scope,omitempty"`
	ValidUntil     string            `json:"valid_until,omitempty"` // approve requires a date or explicit "none"
	SourceSHA256   string            `json:"source_sha256"`
	Preview        *ReviewedArtifact `json:"preview,omitempty"`
	Report         *ReviewedArtifact `json:"report,omitempty"`
}
type ReviewReceipt struct {
	Schema                 string `json:"schema"`
	LibraryID              string `json:"library_id"`
	PreviousRevision       int    `json:"previous_revision"`
	PreviousRevisionSHA256 string `json:"previous_revision_sha256"`
	NewRevision            int    `json:"new_revision"`
	Action                 string `json:"action"`
	Actor                  string `json:"actor"`
	Date                   string `json:"date"`
	Reason                 string `json:"reason"`
	DecisionSHA256         string `json:"decision_sha256"`
	DecisionPath           string `json:"decision_path"`
	PreviousManifestPath   string `json:"previous_manifest_path"`
	Scope                  string `json:"scope"`
}
type ReviewResult struct {
	Manifest Manifest      `json:"manifest"`
	Receipt  ReviewReceipt `json:"receipt"`
}

func reviewedContentSource(raw []byte, pins Pins) error {
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var document, extra yaml.Node
	if e := d.Decode(&document); e != nil {
		return e
	}
	if e := d.Decode(&extra); e != io.EOF || len(document.Content) != 1 {
		return fmt.Errorf("finished-slide.review_source_document_invalid")
	}
	count := 0
	var walk func(*yaml.Node, int) error
	walk = func(node *yaml.Node, depth int) error {
		count++
		if depth > 100 || count > 200000 || node.Kind == yaml.AliasNode || node.Anchor != "" {
			return fmt.Errorf("finished-slide.review_source_structure_invalid")
		}
		if node.Kind == yaml.MappingNode {
			seen := map[string]bool{}
			for i := 0; i < len(node.Content); i += 2 {
				key := node.Content[i]
				if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || seen[key.Value] {
					return fmt.Errorf("finished-slide.review_source_duplicate_or_invalid_key")
				}
				seen[key.Value] = true
			}
		}
		for _, child := range node.Content {
			if e := walk(child, depth+1); e != nil {
				return e
			}
		}
		return nil
	}
	root := document.Content[0]
	if e := walk(root, 0); e != nil {
		return e
	}
	field := func(node *yaml.Node, name string) *yaml.Node {
		if node == nil || node.Kind != yaml.MappingNode {
			return nil
		}
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == name {
				return node.Content[i+1]
			}
		}
		return nil
	}
	text := func(node *yaml.Node) string {
		if node != nil && node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
			return node.Value
		}
		return ""
	}
	if text(field(root, "content_kind")) != "supplied_content" || text(field(root, "id")) == "" {
		return fmt.Errorf("finished-slide.review_requires_supplied_content")
	}
	template := field(root, "template")
	if text(field(template, "scope")) != "shared" || text(field(template, "id")) != pins.TemplateID {
		return fmt.Errorf("finished-slide.review_template_mismatch")
	}
	return nil
}

// Check duplicate JSON keys and structural bounds before typed decoding.
func reviewJSON(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	nodes := 0
	var value func(int) error
	value = func(depth int) error {
		nodes++
		if depth > 32 || nodes > 20000 {
			return fmt.Errorf("finished-slide.review_json_bounds")
		}
		token, e := d.Token()
		if e != nil {
			return e
		}
		if delimiter, ok := token.(json.Delim); ok {
			switch delimiter {
			case '{':
				seen := map[string]bool{}
				for d.More() {
					key, e := d.Token()
					if e != nil {
						return e
					}
					name, ok := key.(string)
					if !ok || seen[name] {
						return fmt.Errorf("finished-slide.review_duplicate_key")
					}
					seen[name] = true
					if e := value(depth + 1); e != nil {
						return e
					}
				}
				end, e := d.Token()
				if e != nil || end != json.Delim('}') {
					return fmt.Errorf("finished-slide.review_invalid_object")
				}
			case '[':
				for d.More() {
					if e := value(depth + 1); e != nil {
						return e
					}
				}
				end, e := d.Token()
				if e != nil || end != json.Delim(']') {
					return fmt.Errorf("finished-slide.review_invalid_array")
				}
			default:
				return fmt.Errorf("finished-slide.review_invalid_delimiter")
			}
		}
		return nil
	}
	if e := value(0); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return fmt.Errorf("finished-slide.review_trailing_data")
	}
	return nil
}

func DecodeReviewDecision(raw []byte) (ReviewDecision, error) {
	var out ReviewDecision
	if len(raw) == 0 || len(raw) > 2<<20 || !utf8.Valid(raw) {
		return out, fmt.Errorf("finished-slide.review_input_invalid")
	}
	if e := reviewJSON(raw); e != nil {
		return out, e
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&out); e != nil {
		return out, e
	}
	if out.Schema != ReviewDecisionSchema || !validSHA(out.RevisionSHA256) || !validSHA(out.SourceSHA256) || out.NewRevision < 1 || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || strings.TrimSpace(out.Reason) == "" || len(out.Reason) > 4096 {
		return out, fmt.Errorf("finished-slide.review_identity_incomplete")
	}
	if out.Action != "approve" && out.Action != "deprecate" && out.Action != "draft" {
		return out, fmt.Errorf("finished-slide.review_action_invalid")
	}
	if _, e := time.Parse(time.DateOnly, out.Date); e != nil {
		return out, fmt.Errorf("finished-slide.review_date_invalid")
	}
	if out.Action == "approve" {
		if strings.TrimSpace(out.ReuseScope) == "" || len(out.ReuseScope) > 4096 || out.ValidUntil == "" || out.Preview == nil || out.Report == nil {
			return out, fmt.Errorf("finished-slide.review_approval_incomplete")
		}
		if out.ValidUntil != "none" {
			until, e := time.Parse(time.DateOnly, out.ValidUntil)
			date, _ := time.Parse(time.DateOnly, out.Date)
			if e != nil || until.Before(date) {
				return out, fmt.Errorf("finished-slide.review_freshness_invalid")
			}
		}
	} else if out.ReuseScope != "" || out.ValidUntil != "" || out.Preview != nil || out.Report != nil {
		return out, fmt.Errorf("finished-slide.review_nonapproval_options")
	}
	for _, artifact := range []*ReviewedArtifact{out.Preview, out.Report} {
		if artifact != nil && (!portablePath(artifact.Path) || !validSHA(artifact.SHA256)) {
			return out, fmt.Errorf("finished-slide.review_artifact_invalid")
		}
	}
	return out, nil
}

// ReviewRevision creates a new closed revision; it never edits the reviewed
// package. Approval facts are explicit operator input, not generated verdicts.
func ReviewRevision(root, destination string, decisionRaw []byte) (ReviewResult, error) {
	return reviewRevisionAt(root, destination, decisionRaw, time.Now())
}
func reviewRevisionAt(root, destination string, decisionRaw []byte, now time.Time) (ReviewResult, error) {
	var result ReviewResult
	d, e := DecodeReviewDecision(decisionRaw)
	if e != nil {
		return result, e
	}
	m, e := Read(root)
	if e != nil {
		return result, e
	}
	if m.RevisionSHA256 != d.RevisionSHA256 || d.NewRevision <= m.Revision {
		return result, fmt.Errorf("finished-slide.review_revision_mismatch")
	}
	if d.Date > now.UTC().Format(time.DateOnly) || d.Action == "approve" && d.ValidUntil != "none" && d.ValidUntil < now.UTC().Format(time.DateOnly) {
		return result, fmt.Errorf("finished-slide.review_date_future_or_expired")
	}
	inventory := map[string]File{}
	for _, f := range m.Files {
		inventory[f.Path] = f
	}
	if inventory[m.Source].SHA256 != d.SourceSHA256 {
		return result, fmt.Errorf("finished-slide.review_source_mismatch")
	}
	for role, artifact := range map[string]*ReviewedArtifact{"preview": d.Preview, "review": d.Report} {
		if artifact != nil {
			f, ok := inventory[artifact.Path]
			if !ok || f.Role != role || f.SHA256 != artifact.SHA256 || f.Bytes == 0 {
				return result, fmt.Errorf("finished-slide.review_artifact_mismatch: %s", role)
			}
		}
	}
	files, e := ReadPayload(root, m)
	if e != nil {
		return result, e
	}
	if d.Action == "approve" {
		if e := reviewedContentSource(files[m.Source], m.Pins); e != nil {
			return result, e
		}
	}
	previous, e := json.Marshal(m)
	if e != nil {
		return result, e
	}
	base := "review-history/" + m.RevisionSHA256 + "/"
	receipt := ReviewReceipt{Schema: ReviewReceiptSchema, LibraryID: m.ID, PreviousRevision: m.Revision, PreviousRevisionSHA256: m.RevisionSHA256, NewRevision: d.NewRevision, Action: d.Action, Actor: d.Actor, Date: d.Date, Reason: d.Reason, DecisionSHA256: sha(decisionRaw), DecisionPath: base + "decision.json", PreviousManifestPath: base + "previous-manifest.json", Scope: "explicit reuse stewardship; destination copy/evidence/fit/native approval remain separate"}
	receiptRaw, e := json.Marshal(receipt)
	if e != nil {
		return result, e
	}
	for name, raw := range map[string][]byte{receipt.DecisionPath: decisionRaw, receipt.PreviousManifestPath: previous, base + "receipt.json": receiptRaw} {
		if _, exists := inventory[name]; exists {
			return result, fmt.Errorf("finished-slide.review_history_collision")
		}
		files[name] = append([]byte(nil), raw...)
		m.Files = append(m.Files, File{Path: name, Role: "documentation"})
	}
	m.Revision, m.ReviewedAt, m.Approval, m.ValidUntil = d.NewRevision, d.Date, nil, ""
	switch d.Action {
	case "approve":
		m.Lifecycle = "approved"
		m.Approval = &Approval{By: d.Actor, Date: d.Date, ReuseScope: d.ReuseScope}
		if d.ValidUntil != "none" {
			m.ValidUntil = d.ValidUntil
		}
	case "deprecate":
		m.Lifecycle = "deprecated"
	case "draft":
		m.Lifecycle = "draft"
	}
	m, e = Create(destination, m, files)
	if e != nil {
		return result, e
	}
	result.Manifest, result.Receipt = m, receipt
	return result, nil
}
