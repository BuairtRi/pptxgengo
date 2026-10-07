package deckproject

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// LayoutPlan only changes file references, never stable IDs or expanded content.
// Old paths and the transaction's source preimages remain available to history.
type LayoutPlan struct {
	Schema         string            `json:"schema"`
	BeforeSHA256   string            `json:"before_sha256"`
	AfterSHA256    string            `json:"after_sha256,omitempty"`
	SemanticSHA256 string            `json:"semantic_sha256"`
	Files          map[string]string `json:"files"`
	Applied        bool              `json:"applied"`
	ApprovalPolicy string            `json:"approval_policy"`
}

func PortableLayout(p *Project, apply bool) (LayoutPlan, error) {
	r := LayoutPlan{Schema: "pptxgengo.project-layout.v1", BeforeSHA256: p.SourceHash(), SemanticSHA256: digest(p.Canonical), Files: map[string]string{}, ApprovalPolicy: "History and receipts retained; source-file dependency approvals are recomputed by project status. Rebuild before saving a numbered version."}
	if e := refuseSyncConflicts(p.Root); e != nil {
		return r, e
	}
	main, e := sourceYAML(p.Raw)
	if e != nil {
		return r, e
	}
	objects, docs, e := authoredSlides(p, main)
	if e != nil {
		return r, e
	}
	changes := map[string][]byte{}
	slides := mappingNode(main.Content[0], "slides")
	for i, s := range p.Document.Slides {
		relative := "slides/" + s.ID + ".yaml"
		if e := portableName(relative); e != nil {
			return r, e
		}
		node := objects[s.ID]
		document := docs[p.SlideFiles[s.ID]]
		if document == nil {
			document = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}}
		}
		useBlockCollections(document)
		raw, e := encodeSourceYAML(document)
		if e != nil {
			return r, e
		}
		changes[relative] = raw
		r.Files[relative] = digest(raw)
		old := slides.Content[i]
		slides.Content[i] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: relative, HeadComment: old.HeadComment, LineComment: old.LineComment, FootComment: old.FootComment}
	}
	templates := mappingNode(main.Content[0], "local_templates")
	if templates != nil {
		for i := 0; i+1 < len(templates.Content); i += 2 {
			id, node := templates.Content[i].Value, templates.Content[i+1]
			var document *yaml.Node
			if old := p.TemplateFiles[id]; old != "" {
				document, e = sourceYAML(p.SourceFiles[old])
				if e != nil {
					return r, e
				}
				node = document.Content[0]
			} else {
				document = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}}
			}
			relative := "slides/templates/" + id + ".yaml"
			if e := portableName(relative); e != nil {
				return r, e
			}
			useBlockCollections(document)
			raw, e := encodeSourceYAML(document)
			if e != nil {
				return r, e
			}
			changes[relative] = raw
			r.Files[relative] = digest(raw)
			templates.Content[i+1] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: relative, HeadComment: node.HeadComment, LineComment: node.LineComment, FootComment: node.FootComment}
		}
	}
	useBlockCollections(main)
	raw, e := encodeSourceYAML(main)
	if e != nil {
		return r, e
	}
	changes[filepath.Base(p.SourcePath)] = raw
	candidate, e := loadProject(p.SourcePath, changes)
	if e != nil {
		return r, e
	}
	if !bytes.Equal(candidate.Canonical, p.Canonical) {
		return r, fmt.Errorf("portable layout would change expanded source")
	}
	r.AfterSHA256 = candidate.SourceHash()
	// Dry runs must detect occupied destinations too, without writing anything.
	for rel, data := range changes {
		if _, owned := p.SourceFiles[rel]; owned {
			continue
		}
		path, e := SafePath(p.Root, rel)
		if e != nil {
			return r, e
		}
		prior, e := readOptional(path)
		if e != nil {
			return r, e
		}
		if prior != nil && !bytes.Equal(prior, data) {
			return r, fmt.Errorf("layout destination already exists: %s", rel)
		}
		if prior != nil {
			delete(changes, rel)
		}
	}
	if !apply {
		return r, nil
	}
	_, e = commitSourceChanges(p, changes, func(next *Project) error {
		if !bytes.Equal(next.Canonical, p.Canonical) {
			return fmt.Errorf("layout changed semantic content")
		}
		return nil
	})
	r.Applied = e == nil
	return r, e
}

func sortedFileKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
