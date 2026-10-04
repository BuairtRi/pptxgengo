package deckproject

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
)

type VisualDecision struct {
	Status   string `json:"status"`
	Reviewer string `json:"reviewer"`
	Note     string `json:"note,omitempty"`
}
type NativeAttachment struct {
	Schema               string                    `json:"schema"`
	ID                   string                    `json:"id"`
	BuildID              string                    `json:"build_id"`
	SourceSemanticSHA256 string                    `json:"source_semantic_sha256"`
	PPTXSHA256           string                    `json:"pptx_sha256"`
	RenderManifestSHA256 string                    `json:"render_manifest_sha256"`
	Created              string                    `json:"created"`
	SlideIDs             []string                  `json:"slide_ids"`
	Files                map[string]string         `json:"files"`
	Decisions            map[string]VisualDecision `json:"visual_decisions"`
}
type NativeCoverage struct {
	SlideID    string `json:"slide_id"`
	Rendered   bool   `json:"rendered"`
	Status     string `json:"visual_status"`
	Image      string `json:"image,omitempty"`
	Attachment string `json:"attachment,omitempty"`
	Note       string `json:"note,omitempty"`
}

// AttachNativeRender verifies and copies immutable rendering evidence. Rendering
// success alone never implies that a slide was visually reviewed or accepted.
func AttachNativeRender(p *Project, root string, decisions map[string]VisualDecision) (NativeAttachment, error) {
	a := NativeAttachment{Schema: "pptxgengo.native-review-attachment.v1", SlideIDs: []string{}, Files: map[string]string{}, Decisions: decisions}
	s, err := Status(p)
	if err != nil {
		return a, err
	}
	if err = protectBaseline(p, s); err != nil {
		return a, err
	}
	deps, err := dependencies(p)
	if err != nil {
		return a, err
	}
	if s.CurrentBuild == "" || s.SemanticSHA256 != digest(p.Canonical) || !reflect.DeepEqual(deps, s.Dependencies) {
		return a, fmt.Errorf("attach-render requires a current generated build with unchanged inputs")
	}
	deckPath, err := SafePath(p.Root, "builds/"+s.CurrentBuild+"/deck.pptx")
	if err != nil {
		return a, err
	}
	deck, err := os.ReadFile(deckPath)
	if err != nil {
		return a, err
	}
	manifestPath, err := SafePath(root, "render-manifest.json")
	if err != nil {
		return a, err
	}
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return a, err
	}
	var receipt nativeexport.Receipt
	if err = json.Unmarshal(raw, &receipt); err != nil {
		return a, err
	}
	if !strings.HasPrefix(receipt.Renderer, "Microsoft PowerPoint (local native PDF)") {
		return a, fmt.Errorf("attachment requires an explicit local PowerPoint native receipt; previews cannot establish native coverage")
	}
	if receipt.Source.SHA256 != digest(deck) || receipt.Slides != len(p.Document.Slides) {
		return a, fmt.Errorf("render receipt does not match the current generated PowerPoint")
	}
	if receipt.Pages < 1 {
		return a, fmt.Errorf("render receipt has no exported pages")
	}
	if len(receipt.PageMappings) == 0 {
		for i, slide := range p.Document.Slides {
			if receipt.IncludeHidden || !slide.Hidden {
				receipt.PageMappings = append(receipt.PageMappings, nativeexport.PageMapping{Page: len(receipt.PageMappings) + 1, SourceSlide: i + 1, SourceHidden: slide.Hidden})
			}
		}
	}
	if len(receipt.PageMappings) != receipt.Pages {
		return a, fmt.Errorf("native receipt page mapping count is inconsistent")
	}
	if len(receipt.PNGs) > 0 && len(receipt.PNGs) != receipt.Pages {
		return a, fmt.Errorf("native PNG coverage count is inconsistent")
	}
	seen := map[string]bool{}
	for i, m := range receipt.PageMappings {
		if m.Page != i+1 || m.SourceSlide < 1 || m.SourceSlide > len(p.Document.Slides) {
			return a, fmt.Errorf("invalid native page mapping")
		}
		if i > 0 && m.SourceSlide <= receipt.PageMappings[i-1].SourceSlide {
			return a, fmt.Errorf("native source slide order is inconsistent")
		}
		if m.PNG != "" && (i >= len(receipt.PNGs) || receipt.PNGs[i].Path != m.PNG) {
			return a, fmt.Errorf("native mapped image is absent from verified page artifacts")
		}
		slide := p.Document.Slides[m.SourceSlide-1]
		if seen[slide.ID] || m.SourceHidden != slide.Hidden {
			return a, fmt.Errorf("duplicate or incorrect hidden-state mapping")
		}
		seen[slide.ID] = true
		a.SlideIDs = append(a.SlideIDs, slide.ID)
	}
	for id, d := range decisions {
		if !seen[id] {
			return a, fmt.Errorf("visual decision references unrendered slide %s", id)
		}
		if d.Status != "reviewed" && d.Status != "accepted" && d.Status != "issues_found" {
			return a, fmt.Errorf("visual decision for %s requires reviewed, accepted or issues_found", id)
		}
		if strings.TrimSpace(d.Reviewer) == "" {
			return a, fmt.Errorf("visual decision for %s requires reviewer", id)
		}
	}
	files := map[string][]byte{"render-manifest.json": raw}
	artifacts := append([]nativeexport.Artifact(nil), receipt.PNGs...)
	if receipt.PDF != nil {
		artifacts = append(artifacts, *receipt.PDF)
	}
	if receipt.ContactSheet != nil {
		artifacts = append(artifacts, *receipt.ContactSheet)
	}
	for _, artifact := range artifacts {
		if _, exists := files[artifact.Path]; exists {
			return a, fmt.Errorf("duplicate native artifact path %s", artifact.Path)
		}
		path, e := SafePath(root, artifact.Path)
		if e != nil {
			return a, e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return a, e
		}
		if digest(data) != artifact.SHA256 {
			return a, fmt.Errorf("native artifact hash mismatch: %s", artifact.Path)
		}
		files[artifact.Path] = data
		a.Files[artifact.Path] = artifact.SHA256
	}
	if receipt.PDF == nil && len(receipt.PNGs) == 0 {
		return a, fmt.Errorf("native receipt contains no page artifacts")
	}
	a.BuildID = s.CurrentBuild
	a.SourceSemanticSHA256 = digest(p.Canonical)
	a.PPTXSHA256 = digest(deck)
	a.RenderManifestSHA256 = digest(raw)
	a.Files["render-manifest.json"] = digest(raw)
	a.ID = nativeAttachmentID(a)
	base := "reviews/native/" + a.ID
	path, err := SafePath(p.Root, base)
	if err != nil {
		return a, err
	}
	if prior, e := os.ReadFile(filepath.Join(path, "attachment.json")); e == nil {
		var stored NativeAttachment
		if e = json.Unmarshal(prior, &stored); e != nil {
			return a, e
		}
		a.Created = stored.Created
		if !reflect.DeepEqual(prior, canonical(a)) {
			return a, fmt.Errorf("native attachment drift")
		}
		if _, e = validateNativeAttachment(path, stored); e != nil {
			return a, e
		}
		return stored, nil
	}
	a.Created = time.Now().UTC().Format(time.RFC3339Nano)
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return a, err
	}
	stage, err := os.MkdirTemp(filepath.Dir(path), ".native-attach-")
	if err != nil {
		return a, err
	}
	defer os.RemoveAll(stage)
	for rel, data := range files {
		destination, e := SafePath(stage, rel)
		if e != nil {
			return a, e
		}
		if e = writeExclusive(destination, data, 0444); e != nil {
			return a, e
		}
	}
	if err = writeExclusive(filepath.Join(stage, "attachment.json"), canonical(a), 0444); err != nil {
		return a, err
	}
	// Recheck both source dependencies and baseline immediately before publishing.
	now, err := Status(p)
	if err != nil {
		return a, err
	}
	if now.CurrentBuild != s.CurrentBuild {
		return a, fmt.Errorf("build changed during native attachment")
	}
	if err = protectBaseline(p, now); err != nil {
		return a, err
	}
	actual, err := Load(p.SourcePath)
	if err != nil {
		return a, err
	}
	currentDeps, err := dependencies(actual)
	if err != nil {
		return a, err
	}
	if actual.SourceHash() != p.SourceHash() || !reflect.DeepEqual(currentDeps, deps) {
		return a, fmt.Errorf("source changed during native attachment")
	}
	if err = os.Rename(stage, path); err != nil {
		return a, err
	}
	return a, nil
}

// NativeReviewCoverage verifies attached files each time, exposing stale and
// missing coverage instead of upgrading build success to visual acceptance.
func NativeReviewCoverage(p *Project) ([]NativeCoverage, error) {
	rows := make([]NativeCoverage, len(p.Document.Slides))
	positions := map[string]int{}
	for i, s := range p.Document.Slides {
		rows[i] = NativeCoverage{SlideID: s.ID, Status: "not_reviewed"}
		positions[s.ID] = i
	}
	state, err := Status(p)
	if err != nil {
		return nil, err
	}
	deps, err := dependencies(p)
	if err != nil {
		return nil, err
	}
	if state.CurrentBuild == "" {
		return rows, nil
	}
	if err = protectBaseline(p, state); err != nil {
		return nil, err
	}
	deckPath, err := SafePath(p.Root, "builds/"+state.CurrentBuild+"/deck.pptx")
	if err != nil {
		return nil, err
	}
	deck, err := os.ReadFile(deckPath)
	if err != nil {
		return nil, err
	}
	deckHash := digest(deck)
	base := filepath.Join(p.Root, "reviews/native")
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return rows, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	type storedAttachment struct {
		path       string
		attachment NativeAttachment
	}
	stored := []storedAttachment{}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path, e := SafePath(p.Root, "reviews/native/"+entry.Name()+"/attachment.json")
		if e != nil {
			return nil, e
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		var a NativeAttachment
		if e = json.Unmarshal(raw, &a); e != nil {
			return nil, e
		}
		if a.Schema != "pptxgengo.native-review-attachment.v1" {
			return nil, fmt.Errorf("unknown native attachment schema")
		}
		stored = append(stored, storedAttachment{path, a})
	}
	sort.Slice(stored, func(i, j int) bool { return stored[i].attachment.Created < stored[j].attachment.Created })
	for _, record := range stored {
		path, a := record.path, record.attachment
		current := a.BuildID == state.CurrentBuild && a.PPTXSHA256 == deckHash && a.SourceSemanticSHA256 == digest(p.Canonical) && state.SemanticSHA256 == digest(p.Canonical) && reflect.DeepEqual(deps, state.Dependencies)
		manifest, e := validateNativeAttachment(filepath.Dir(path), a)
		if e != nil {
			return nil, e
		}
		for j, id := range a.SlideIDs {
			i, exists := positions[id]
			if !exists {
				continue
			}
			if !current {
				if !rows[i].Rendered {
					rows[i].Status = "stale"
				}
				continue
			}
			if len(manifest.PageMappings) > 0 {
				mapping := manifest.PageMappings[j]
				if mapping.SourceSlide != i+1 || mapping.SourceHidden != p.Document.Slides[i].Hidden {
					return nil, fmt.Errorf("native slide mapping disagrees with current source")
				}
			}
			status := "not_reviewed"
			note := ""
			if d, ok := a.Decisions[id]; ok {
				status = d.Status
				note = d.Note
			}
			if rows[i].Rendered && status == "not_reviewed" && rows[i].Status != "not_reviewed" {
				continue
			}
			image := ""
			if j < len(manifest.PageMappings) && manifest.PageMappings[j].PNG != "" {
				image = filepath.ToSlash(filepath.Join("reviews/native", a.ID, manifest.PageMappings[j].PNG))
			} else if j < len(manifest.PNGs) {
				image = filepath.ToSlash(filepath.Join("reviews/native", a.ID, manifest.PNGs[j].Path))
			}
			rows[i] = NativeCoverage{SlideID: id, Rendered: true, Status: status, Image: image, Attachment: a.ID, Note: note}
		}
	}
	return rows, nil
}
