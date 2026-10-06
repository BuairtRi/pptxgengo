package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

func validateDraftReview(note *wmdesign.DraftReviewNote) error {
	return wmdesign.ValidateDraftReviewNote(note)
}

type DraftReviewOperation struct {
	Action, ID string
	// Fields contains only explicitly supplied fields; omitted fields survive set.
	Fields map[string]string
}

type DraftReviewReceipt struct {
	Operation    string                    `json:"operation"`
	SlideID      string                    `json:"slide_id"`
	DraftReview  *wmdesign.DraftReviewNote `json:"draft_review"`
	BeforeSHA256 string                    `json:"before_sha256,omitempty"`
	AfterSHA256  string                    `json:"after_sha256,omitempty"`
	Decision     string                    `json:"decision,omitempty"`
}

// OperateDraftReview attaches drafting metadata directly to a slide. It uses the
// source mutation guard and preserves unselected sources and existing YAML nodes.
func OperateDraftReview(p *Project, o DraftReviewOperation) (DraftReviewReceipt, error) {
	r := DraftReviewReceipt{Operation: "slide-draft-review-" + o.Action, SlideID: o.ID}
	if o.Action != "set" && o.Action != "clear" && o.Action != "show" {
		return r, fmt.Errorf("draft-review operation must be set, clear or show")
	}
	if !stableID.MatchString(o.ID) {
		return r, fmt.Errorf("draft-review requires a valid stable --id")
	}
	index := -1
	for i, slide := range p.Document.Slides {
		if slide.ID == o.ID {
			index = i
			r.DraftReview = slide.DraftReview
			break
		}
	}
	if index < 0 {
		return r, fmt.Errorf("unknown slide ID %q", o.ID)
	}
	if o.Action != "set" && len(o.Fields) > 0 {
		return r, fmt.Errorf("draft-review fields require set")
	}
	if o.Action == "show" {
		return r, nil
	}
	if o.Action == "set" && len(o.Fields) == 0 {
		return r, fmt.Errorf("draft-review set requires at least one field")
	}
	for key := range o.Fields {
		switch key {
		case "status", "status_text", "status_color", "owner", "due", "updated", "notes", "placement":
		default:
			return r, fmt.Errorf("unknown draft-review field %q", key)
		}
	}
	unchanged := o.Action == "clear" && r.DraftReview == nil
	if unchanged {
		// A null YAML field still needs removal even though it decodes to nil.
		if slides, ok := p.tree["slides"].([]any); ok {
			if slide, ok := slides[index].(map[string]any); ok {
				_, exists := slide["draft_review"]
				unchanged = !exists
			}
		}
	}
	if o.Action == "set" && r.DraftReview != nil {
		note := r.DraftReview
		current := map[string]string{"status": note.Status, "status_text": note.StatusText, "status_color": note.StatusColor, "owner": note.Owner, "due": note.Due, "updated": note.Updated, "notes": note.Notes, "placement": note.Placement}
		unchanged = true
		for key, value := range o.Fields {
			if current[key] != value {
				unchanged = false
			}
		}
	}
	if unchanged {
		r.BeforeSHA256, r.AfterSHA256 = p.SourceHash(), p.SourceHash()
		// Check predecessor bytes under the regular guard without reserializing.
		_, err := commitSourceChanges(p, map[string][]byte{}, nil)
		return r, err
	}
	main, err := sourceYAML(p.Raw)
	if err != nil {
		return r, err
	}
	nodes, documents, err := authoredSlides(p, main)
	if err != nil {
		return r, err
	}
	node := nodes[o.ID]
	if o.Action == "clear" {
		removeMappingField(node, "draft_review")
	} else {
		note := mappingNode(node, "draft_review")
		if note == nil || note.Kind != yaml.MappingNode {
			note = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			replaceMappingField(note, "status", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "notstarted"})
			replaceMappingField(node, "draft_review", note)
		}
		for _, key := range []string{"status", "status_text", "status_color", "owner", "due", "updated", "notes", "placement"} {
			value, supplied := o.Fields[key]
			if !supplied {
				continue
			}
			replacement := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
			if previous := mappingNode(note, key); previous != nil {
				replacement.Style = previous.Style
			}
			if key == "notes" {
				replacement.Style = yaml.LiteralStyle
			}
			replaceMappingField(note, key, replacement)
		}
	}
	relative := p.SlideFiles[o.ID]
	if relative == "" {
		relative = filepath.Base(p.SourcePath)
	}
	raw, err := encodeSourceYAML(documents[relative])
	if err != nil {
		return r, err
	}
	r.BeforeSHA256 = p.SourceHash()
	r.Decision = "decisions/" + r.Operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, err := SafePath(p.Root, r.Decision)
	if err != nil {
		return r, err
	}
	_, err = commitSourceChanges(p, map[string][]byte{relative: raw}, func(candidate *Project) error {
		r.AfterSHA256 = candidate.SourceHash()
		r.DraftReview = candidate.Document.Slides[index].DraftReview
		return writeJSON(decision, r)
	})
	if err != nil {
		os.Remove(decision)
	}
	return r, err
}
