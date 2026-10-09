package deckproject

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CompositionResult is a measured source proposal, not desktop qualification.
type CompositionResult struct {
	Schema       string            `json:"schema"`
	Operation    string            `json:"operation"`
	Actor        string            `json:"actor"`
	Reason       string            `json:"reason"`
	Applied      bool              `json:"applied"`
	BeforeSHA256 string            `json:"before_sha256"`
	AfterSHA256  string            `json:"after_sha256"`
	Decision     string            `json:"decision,omitempty"`
	Inspection   DiagramInspection `json:"inspection"`
	Evidence     any               `json:"evidence,omitempty"`
}

// CompositionCandidate shares one source transaction for domain composers.
// Family validation belongs to the caller; this boundary preserves the frame,
// refuses stale native layout, measures the result, and guards all source bytes.
func CompositionCandidate(p *Project, slideID, operation, actor, reason string, template LocalTemplate, bundle, engine string, apply bool) (CompositionResult, error) {
	return compositionCandidate(p, slideID, operation, actor, reason, template, bundle, engine, apply, nil, nil)
}

// Additional private evidence is committed atomically with authored source.
func compositionCandidate(p *Project, slideID, operation, actor, reason string, template LocalTemplate, bundle, engine string, apply bool, evidence any, artifacts map[string][]byte) (CompositionResult, error) {
	out := CompositionResult{Schema: "pptxgengo.composition-result.v1", Operation: operation, Actor: actor, Reason: reason, BeforeSHA256: p.SourceHash(), Evidence: evidence}
	if !stableID.MatchString(operation) || strings.TrimSpace(actor) == "" || len(actor) > 256 || strings.TrimSpace(reason) == "" || len(reason) > 4096 {
		return out, fmt.Errorf("composition requires a stable operation, actor and reason")
	}
	idx, original, err := diagramSlide(p, slideID)
	if err != nil {
		return out, err
	}
	s := p.Document.Slides[idx]
	if s.Template.Revision != "" {
		return out, fmt.Errorf("local template revision is pinned; explicitly revise the reference first")
	}
	for _, other := range p.Document.Slides {
		if other.ID != slideID && other.Template.Scope == "local" && other.Template.ID == s.Template.ID {
			return out, fmt.Errorf("local template is shared; project fork it for this slide first")
		}
	}
	if len(s.NativeGeometry) > 0 || len(s.NativeOrder) > 0 {
		return out, fmt.Errorf("composition changes require explicit reset_native_layout first; preserve and review native edits")
	}
	var candidateTemplate LocalTemplate
	if err = strictInto(template, &candidateTemplate); err != nil {
		return out, err
	}
	// Domain edits cannot silently alter chrome, grid, provenance or style contract.
	x, y := original, candidateTemplate
	x.Nodes, x.Zones = nil, nil
	y.Nodes, y.Zones = nil, nil
	if !bytes.Equal(canonical(x), canonical(y)) {
		return out, fmt.Errorf("composition edits must preserve template frame, grid and provenance")
	}
	main, err := sourceYAML(p.Raw)
	if err != nil {
		return out, err
	}
	slides, documents, err := authoredSlides(p, main)
	if err != nil {
		return out, err
	}
	templateNode := mappingNode(mappingNode(main.Content[0], "local_templates"), s.Template.ID)
	file := p.TemplateFiles[s.Template.ID]
	if file != "" {
		doc, e := sourceYAML(p.SourceFiles[file])
		if e != nil {
			return out, e
		}
		documents[file] = doc
		templateNode = doc.Content[0]
	} else {
		file = filepath.Base(p.SourcePath)
	}
	if templateNode == nil || slides[slideID] == nil {
		return out, fmt.Errorf("composition authored source missing")
	}
	slideChanged := pruneRemovedNodeContainment(original.Nodes, candidateTemplate.Nodes, slides[slideID])
	pruned, err := updateDiagramNodes(candidateTemplate, s, templateNode, slides[slideID], false)
	if err != nil {
		return out, err
	}
	slideChanged = slideChanged || pruned
	changes := map[string][]byte{}
	changes[file], err = encodeSourceYAML(documents[file])
	if err != nil {
		return out, err
	}
	if slideChanged {
		sf := p.SlideFiles[slideID]
		if sf == "" {
			sf = filepath.Base(p.SourcePath)
		}
		changes[sf], err = encodeSourceYAML(documents[sf])
		if err != nil {
			return out, err
		}
	}
	candidate, err := loadProject(p.SourcePath, mergeTextOverrides(p.SourceFiles, changes))
	if err != nil {
		return out, err
	}
	out.AfterSHA256 = candidate.SourceHash()
	out.Inspection, err = InspectDiagram(candidate, slideID, bundle, engine)
	if err != nil {
		return out, err
	}
	if !apply {
		return out, nil
	}
	out.Decision = "decisions/" + operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	guarded := map[string][]byte{}
	for path, data := range artifacts {
		if path != "assets/objects/sha256/"+digest(data) {
			return out, fmt.Errorf("evidence must use a content-addressed asset path")
		}
		if _, exists := changes[path]; exists {
			return out, fmt.Errorf("evidence path conflicts with source transaction")
		}
		absolute, e := SafePath(p.Root, path)
		if e != nil {
			return out, e
		}
		old, e := os.ReadFile(absolute)
		if e == nil {
			if !bytes.Equal(old, data) {
				return out, fmt.Errorf("retained evidence hash collision")
			}
			guarded[path] = old
			continue // Immutable global assets need no rewrite or predecessor copy.
		} else if !os.IsNotExist(e) {
			return out, e
		}
		changes[path] = data
	}
	out.Applied = true
	changes[out.Decision] = canonical(out)
	_, err = commitSourceChangesChecked(p, changes, nil, guarded, func(c *Project) error { _, e := InspectDiagram(c, slideID, bundle, engine); return e })
	if err != nil {
		out.Applied = false
	}
	return out, err
}
