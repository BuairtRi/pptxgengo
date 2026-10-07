package deckproject

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func markdownClaimIDs(raw []byte) (map[string]bool, error) {
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("claims markdown requires UTF-8")
	}
	ids := map[string]bool{}
	add := func(id string) error {
		if !stableID.MatchString(id) || ids[id] {
			return fmt.Errorf("claims registry invalid or duplicate ID %s", id)
		}
		ids[id] = true
		if len(ids) > 10000 {
			return fmt.Errorf("claims registry count exceeds limit")
		}
		return nil
	}
	for _, line := range markdownClaimLines(string(raw)) {
		matches := claimAnchor.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if e := add(match[1]); e != nil {
				return nil, e
			}
		}
		if len(matches) == 0 && strings.HasPrefix(line, "## ") {
			id := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if stableID.MatchString(id) {
				if e := add(id); e != nil {
					return nil, e
				}
			}
		}
	}
	return ids, nil
}

// A Markdown claim remains a reference to its exact anchored registry bytes;
// prose is never converted into invented structured claim text.
func readClaimDependencies(p *Project) (ClaimsRegistry, map[string][]byte, error) {
	registry := ClaimsRegistry{Schema: "pptxgengo.claims.v1", Claims: []Claim{}}
	files := map[string][]byte{}
	relative := p.Document.Context["claims"]
	if relative == "" {
		return registry, files, nil
	}
	raw, e := projectDependency(p, relative, 16<<20)
	if e != nil {
		return registry, nil, e
	}
	files[relative] = raw
	if strings.ToLower(filepath.Ext(relative)) == ".md" {
		ids, e := markdownClaimIDs(raw)
		if e != nil {
			return registry, nil, e
		}
		for id := range ids {
			registry.Claims = append(registry.Claims, Claim{ID: id, Registry: &ClaimRegistryReference{Path: relative, SHA256: digest(raw), ClaimID: id, Format: "markdown"}})
		}
		sort.Slice(registry.Claims, func(i, j int) bool { return registry.Claims[i].ID < registry.Claims[j].ID })
		return registry, files, nil
	}
	parser := &Project{SourcePath: relative, Positions: map[string]Position{}, positionFiles: map[string]string{}}
	value, e := parser.parseSource(raw, relative, "")
	if e != nil {
		return registry, nil, e
	}
	if e := parser.shapeType(value, reflect.TypeOf(registry), ""); e != nil {
		return registry, nil, e
	}
	if e := strictInto(value, &registry); e != nil {
		return registry, nil, e
	}
	if registry.Schema != "pptxgengo.claims.v1" || len(registry.Claims) > 10000 {
		return registry, nil, fmt.Errorf("unsupported claims schema or count")
	}
	seen := map[string]bool{}
	parsedMarkdown := map[string]map[string]bool{}
	var total int64 = int64(len(raw))
	read := func(path, expected string) ([]byte, error) {
		if !shaPattern.MatchString(expected) {
			return nil, fmt.Errorf("claims artifact hash required: %s", path)
		}
		data, ok := files[path]
		if !ok {
			if len(files) >= 512 {
				return nil, fmt.Errorf("claims dependency file count exceeds limit")
			}
			var e error
			data, e = projectDependency(p, path, 64<<20)
			if e != nil {
				return nil, e
			}
			total += int64(len(data))
			if total > 512<<20 {
				return nil, fmt.Errorf("claims dependency bytes exceed limit")
			}
			files[path] = data
		}
		if digest(data) != expected {
			return nil, fmt.Errorf("claims artifact hash changed: %s", path)
		}
		return data, nil
	}
	for _, claim := range registry.Claims {
		if !stableID.MatchString(claim.ID) || seen[claim.ID] {
			return registry, nil, fmt.Errorf("claims registry invalid or duplicate ID %s", claim.ID)
		}
		seen[claim.ID] = true
		if claim.Registry == nil && strings.TrimSpace(claim.Text) == "" {
			return registry, nil, fmt.Errorf("claim %s requires text or an exact registry reference", claim.ID)
		}
		if ref := claim.Registry; ref != nil {
			if claim.Text != "" || ref.Format != "markdown" || !stableID.MatchString(ref.ClaimID) {
				return registry, nil, fmt.Errorf("claim %s registry reference invalid", claim.ID)
			}
			data, e := read(ref.Path, ref.SHA256)
			if e != nil {
				return registry, nil, e
			}
			ids := parsedMarkdown[ref.Path]
			if ids == nil {
				var e error
				ids, e = markdownClaimIDs(data)
				if e != nil {
					return registry, nil, e
				}
				parsedMarkdown[ref.Path] = ids
			}
			if !ids[ref.ClaimID] {
				return registry, nil, fmt.Errorf("claim %s registry selector missing", claim.ID)
			}
		}
		if len(claim.Artifacts) > 256 {
			return registry, nil, fmt.Errorf("claim artifact count exceeds limit")
		}
		paths := map[string]bool{}
		for _, artifact := range claim.Artifacts {
			if paths[artifact.Path] {
				return registry, nil, fmt.Errorf("claim duplicate artifact")
			}
			paths[artifact.Path] = true
			if _, e := read(artifact.Path, artifact.SHA256); e != nil {
				return registry, nil, e
			}
		}
	}
	return registry, files, nil
}

func closeFinishedClaims(p *Project, slide Slide) ([]Claim, map[string][]byte, error) {
	if len(slide.EvidenceRefs) == 0 {
		return nil, nil, nil
	}
	if p.Document.Context["claims"] == "" {
		return nil, nil, fmt.Errorf("finished-slide.claims_registry_required")
	}
	registry, dependencies, e := readClaimDependencies(p)
	if e != nil {
		return nil, nil, e
	}
	byID := map[string]Claim{}
	for _, claim := range registry.Claims {
		byID[claim.ID] = claim
	}
	files := map[string][]byte{}
	copyDependency := func(path, hash string, markdown bool) (string, error) {
		raw, ok := dependencies[path]
		if !ok || digest(raw) != hash {
			return "", fmt.Errorf("finished-slide.claim_dependency_missing")
		}
		name := "evidence/" + hash
		if markdown {
			name += ".md"
		}
		if _, exists := files[name]; !exists {
			files[name] = append([]byte(nil), raw...)
		}
		return name, nil
	}
	claims := []Claim{}
	for _, id := range slide.EvidenceRefs {
		claim, ok := byID[id]
		if !ok {
			return nil, nil, fmt.Errorf("finished-slide.claim_missing: %s", id)
		}
		if claim.Registry != nil {
			ref := *claim.Registry
			ref.Path, e = copyDependency(ref.Path, ref.SHA256, true)
			if e != nil {
				return nil, nil, e
			}
			claim.Registry = &ref
		}
		claim.Artifacts = append([]ClaimArtifact(nil), claim.Artifacts...)
		for i := range claim.Artifacts {
			a := &claim.Artifacts[i]
			a.Path, e = copyDependency(a.Path, a.SHA256, false)
			if e != nil {
				return nil, nil, e
			}
		}
		claims = append(claims, claim)
	}
	return claims, files, nil
}

func insertFinishedClaims(p *Project, slide *Slide, claims []Claim, packageFiles, files, observed map[string][]byte, remaps, roles map[string]string) (string, error) {
	if len(slide.EvidenceRefs) == 0 {
		if len(claims) != 0 {
			return "", fmt.Errorf("finished-slide.unreferenced_claims")
		}
		return "", nil
	}
	byID := map[string]Claim{}
	for _, claim := range claims {
		if !stableID.MatchString(claim.ID) || byID[claim.ID].ID != "" {
			return "", fmt.Errorf("finished-slide.claim_dependencies_invalid")
		}
		byID[claim.ID] = claim
	}
	if len(byID) != len(slide.EvidenceRefs) {
		return "", fmt.Errorf("finished-slide.claim_dependencies_incomplete")
	}
	existing, oldFiles, e := readClaimDependencies(p)
	if e != nil {
		return "", e
	}
	for name, raw := range oldFiles {
		observed[name] = raw
	}
	known := map[string]bool{}
	for _, claim := range existing.Claims {
		known[claim.ID] = true
	}
	copied := map[string]string{}
	copyArtifact := func(path, hash string, markdown bool) (string, error) {
		key := fmt.Sprintf("%s/%s/%t", path, hash, markdown)
		if name := copied[key]; name != "" {
			return name, nil
		}
		raw, ok := packageFiles[path]
		if !ok || roles[path] != "evidence" || !shaPattern.MatchString(hash) || digest(raw) != hash {
			return "", fmt.Errorf("finished-slide.claim_artifact_missing_or_changed")
		}
		name := "evidence/reuse-" + nonce() + "-" + hash[:16]
		if markdown {
			name += ".md"
		}
		files[name] = append([]byte(nil), raw...)
		copied[key] = name
		return name, nil
	}
	for index, id := range slide.EvidenceRefs {
		claim, ok := byID[id]
		if !ok || remaps[id] != "" {
			return "", fmt.Errorf("finished-slide.claim_ref_missing_or_duplicate")
		}
		newID := "reuse-" + nonce()
		if known[newID] {
			return "", fmt.Errorf("finished-slide.claim_id_collision")
		}
		known[newID] = true
		remaps[id], slide.EvidenceRefs[index], claim.ID = newID, newID, newID
		if claim.Registry != nil {
			ref := *claim.Registry
			ref.Path, e = copyArtifact(ref.Path, ref.SHA256, true)
			if e != nil {
				return "", e
			}
			claim.Registry = &ref
		}
		claim.Artifacts = append([]ClaimArtifact(nil), claim.Artifacts...)
		for i := range claim.Artifacts {
			a := &claim.Artifacts[i]
			a.Path, e = copyArtifact(a.Path, a.SHA256, false)
			if e != nil {
				return "", e
			}
		}
		existing.Claims = append(existing.Claims, claim)
	}
	relative := p.Document.Context["claims"]
	var tree *yaml.Node
	if relative != "" && strings.ToLower(filepath.Ext(relative)) != ".md" {
		tree, e = sourceYAML(oldFiles[relative])
		if e != nil {
			return "", e
		}
		sequence := mappingNode(tree.Content[0], "claims")
		if sequence == nil || sequence.Kind != yaml.SequenceNode {
			return "", fmt.Errorf("finished-slide.claim_registry_sequence_invalid")
		}
		for _, claim := range existing.Claims[len(existing.Claims)-len(claims):] {
			node, e := editYAMLNode(claim)
			if e != nil {
				return "", e
			}
			sequence.Content = append(sequence.Content, node)
		}
	} else {
		relative = "claims/reuse-" + nonce() + ".yaml"
		tree, e = editYAMLNode(existing)
		if e != nil {
			return "", e
		}
		tree = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{tree}}
	}
	raw, e := encodeSourceYAML(tree)
	if e != nil {
		return "", e
	}
	files[relative] = raw
	return relative, nil
}
