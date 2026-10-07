package deckproject

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

const finishedDependenciesSchema = "pptxgengo.finished-slide-dependencies.v1"

type FinishedSlideDependencies struct {
	Schema        string           `json:"schema"`
	SourceProject string           `json:"source_project"`
	SourceSHA256  string           `json:"source_sha256"`
	Year          int              `json:"year"`
	Assets        map[string]Asset `json:"assets"`
	Claims        []Claim          `json:"claims,omitempty"`
}

type LibraryLineage struct {
	ID             string            `json:"id"`
	Revision       int               `json:"revision"`
	RevisionSHA256 string            `json:"revision_sha256"`
	SourceProject  string            `json:"source_project"`
	SourceSlide    string            `json:"source_slide"`
	SourceSHA256   string            `json:"source_sha256"`
	AssetRemaps    map[string]string `json:"asset_remaps"`
	ItemRemaps     map[string]string `json:"item_remaps"`
	EvidenceRemaps map[string]string `json:"evidence_remaps,omitempty"`
	ContentPolicy  string            `json:"content_policy"`
}

type FinishedSlidePublishOptions struct {
	SlideID, Out, Bundle, Engine string
	Manifest                     finishedslide.Manifest
	// Explicit preview/review artifacts do not grant approval by themselves.
	Artifacts map[string][]byte
}

type FinishedSlideInsertOptions struct {
	Package, ID, Before, After, IntoSection, Rationale, Bundle, Engine string
	AllowDraft                                                         bool
}

type FinishedSlideInsertReceipt struct {
	SlideOperationReceipt
	Library LibraryLineage `json:"library"`
}

func finishedTemplate(bundle, key string) (wmdesign.LibraryTemplate, error) {
	catalog, err := wmdesign.LibraryCatalog(bundle, "")
	if err != nil {
		return wmdesign.LibraryTemplate{}, err
	}
	for _, def := range catalog {
		if def.Key == key {
			return def, nil
		}
	}
	return wmdesign.LibraryTemplate{}, fmt.Errorf("finished-slide.template_missing: %s", key)
}

func finishedPins(p *Project, bundle, engine string, def wmdesign.LibraryTemplate) (finishedslide.Pins, []byte, error) {
	lock, raw, err := ReadLock(p)
	if err != nil {
		return finishedslide.Pins{}, nil, err
	}
	actual, err := makeLock(bundle, engine)
	if err != nil {
		return finishedslide.Pins{}, nil, err
	}
	if !reflect.DeepEqual(actual, lock) {
		return finishedslide.Pins{}, nil, fmt.Errorf("finished-slide.toolchain_drift")
	}
	var tree any
	if err := json.Unmarshal(def.RawSlide, &tree); err != nil {
		return finishedslide.Pins{}, nil, err
	}
	return finishedslide.Pins{Bundle: lock.BundleSHA256, SourceRevision: def.SourceRevision, TemplateID: def.Key, TemplateRevision: def.Revision, TemplateSourceSHA256: def.SourceSHA256, TemplateDefinitionSHA256: digest(canonical(tree)), ToolchainLockSHA256: digest(raw), Compiler: lock.Runtime + ";" + lock.Engine + ";" + lock.ExecutableSHA256}, raw, nil
}

// Only template-declared image slots are asset references. Matching business
// copy is never interpreted as an asset or rewritten during insertion.
func finishedMediaSlots(slide Slide, def wmdesign.LibraryTemplate) (map[string]string, error) {
	result := map[string]string{}
	var source map[string]any
	if err := json.Unmarshal(def.RawSlide, &source); err != nil {
		return nil, err
	}
	slots, _ := slide.Values["slots"].(map[string]any)
	for _, slot := range def.Slots {
		parts, err := contentPointerParts(slot.SourcePointer)
		if err != nil || !sharedImagePointer(source, parts) {
			continue
		}
		value, ok := slots[slot.Name].(string)
		if !ok {
			return nil, fmt.Errorf("finished-slide.media_slot_invalid: %s", slot.Name)
		}
		result[slot.Name] = value
	}
	return result, nil
}

func finishedSupported(slide Slide, def wmdesign.LibraryTemplate) error {
	if !stableID.MatchString(slide.ID) || slide.Template.Scope != "shared" || slide.Template.ID != def.Key || slide.ContentKind != "supplied_content" {
		return fmt.Errorf("finished-slide.requires_content_complete_shared_slide")
	}
	// This initial insertion contract handles declared slot/array bindings and
	// typed cards. Other identity-bearing typed families require explicit mapping.
	if def.ContentContract != wmdesign.LibraryBindingsContract && def.Key != "cards/3" && def.Key != "cards/4" {
		return fmt.Errorf("finished-slide.identity_contract_unsupported: %s", def.Key)
	}
	if _, nav := slide.Values["nav"]; nav {
		return fmt.Errorf("finished-slide.navigation_dependency_unsupported")
	}
	if slide.Template.Revision != "" && slide.Template.Revision != strconv.Itoa(def.Revision) {
		return fmt.Errorf("finished-slide.template_revision_mismatch")
	}
	return nil
}

// PublishFinishedSlide writes a new closed revision without modifying its project.
// Synthetic examples and local templates cannot be relabeled as finished content.
func PublishFinishedSlide(p *Project, o FinishedSlidePublishOptions) (finishedslide.Manifest, error) {
	m := o.Manifest
	var slide Slide
	found := false
	for _, candidate := range p.Document.Slides {
		if candidate.ID == o.SlideID {
			slide = candidate
			found = true
		}
	}
	if !found {
		return m, fmt.Errorf("finished-slide.source_slide_missing")
	}
	def, err := finishedTemplate(o.Bundle, slide.Template.ID)
	if err != nil {
		return m, err
	}
	if err := finishedSupported(slide, def); err != nil {
		return m, err
	}
	if _, err := Check(p, o.Bundle, o.Engine); err != nil {
		return m, err
	}
	m.Pins, _, err = finishedPins(p, o.Bundle, o.Engine, def)
	if err != nil {
		return m, err
	}
	m.Schema, m.Source, m.Files, m.RevisionSHA256 = finishedslide.Schema, "slide.yaml", nil, ""
	files := map[string][]byte{}
	roles := map[string]string{}
	deps := FinishedSlideDependencies{Schema: finishedDependenciesSchema, SourceProject: p.Document.ID, SourceSHA256: p.SourceHash(), Year: p.Document.Year, Assets: map[string]Asset{}}
	claims, claimFiles, err := closeFinishedClaims(p, slide)
	if err != nil {
		return m, err
	}
	deps.Claims = claims
	for name, raw := range claimFiles {
		files[name], roles[name] = raw, "evidence"
	}
	media, err := finishedMediaSlots(slide, def)
	if err != nil {
		return m, err
	}
	registry := map[string]wmdesign.PrimitiveAssetReference{}
	for _, a := range wmdesign.PrimitiveAssetCatalog() {
		registry[a.Key] = a
	}
	names := []string{}
	for name := range media {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		value := media[name]
		if _, ok := registry[value]; ok {
			continue
		}
		id := strings.TrimPrefix(value, "project:")
		a, ok := p.Document.Assets[id]
		if !ok {
			return m, fmt.Errorf("finished-slide.asset_missing: %s", value)
		}
		if _, copied := deps.Assets[id]; copied {
			continue
		}
		if a.DerivedFrom != "" || a.DerivationReceipt != "" {
			return m, fmt.Errorf("finished-slide.derived_asset_dependency_unsupported: %s", id)
		}
		if a.RegistryID != "" {
			r, ok := registry[a.RegistryID]
			if !ok || (a.SHA256 != "" && a.SHA256 != r.SHA256) {
				return m, fmt.Errorf("finished-slide.registry_asset_pin_mismatch: %s", id)
			}
			a.SHA256 = r.SHA256
		} else {
			raw, err := projectDependency(p, a.Path, 64<<20)
			if err != nil {
				return m, err
			}
			if a.SHA256 != "" && a.SHA256 != digest(raw) {
				return m, fmt.Errorf("finished-slide.asset_pin_mismatch: %s", id)
			}
			mime, err := assetMIME(raw)
			if err != nil {
				return m, err
			}
			ext := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/svg+xml": ".svg"}[mime]
			a.Path = "assets/" + digest([]byte(id))[:16] + "-" + digest(raw)[:16] + ext
			a.SHA256 = digest(raw)
			files[a.Path] = raw
			roles[a.Path] = "asset"
		}
		deps.Assets[id] = a
	}
	slide.Template.Revision = strconv.Itoa(def.Revision)
	var source map[string]any
	if err := json.Unmarshal(canonical(slide), &source); err != nil {
		return m, err
	}
	files[m.Source], err = MarshalSlideSource(source)
	if err != nil {
		return m, err
	}
	roles[m.Source] = "source"
	files["dependencies.json"] = append(canonical(deps), '\n')
	roles["dependencies.json"] = "documentation"
	for name, raw := range o.Artifacts {
		role := ""
		switch name {
		case "preview.png":
			role = "preview"
		case "review.json":
			role = "review"
		default:
			return m, fmt.Errorf("finished-slide.artifact_unsupported: %s", name)
		}
		if len(raw) == 0 || len(raw) > 16<<20 {
			return m, fmt.Errorf("finished-slide.artifact_size_invalid: %s", name)
		}
		files[name] = append([]byte(nil), raw...)
		roles[name] = role
	}
	for name, raw := range files {
		m.Files = append(m.Files, finishedslide.File{Path: name, Role: roles[name], Bytes: int64(len(raw)), SHA256: digest(raw)})
	}
	return finishedslide.Create(o.Out, m, files)
}

func readFinishedSource(root string, m finishedslide.Manifest) (Slide, FinishedSlideDependencies, map[string][]byte, error) {
	var slide Slide
	var deps FinishedSlideDependencies
	files, err := finishedslide.ReadPayload(root, m)
	if err != nil {
		return slide, deps, nil, err
	}
	parser := &Project{Positions: map[string]Position{}, positionFiles: map[string]string{}}
	value, err := parser.parseSource(files[m.Source], m.Source, "")
	if err != nil {
		return slide, deps, nil, err
	}
	if err := strictInto(value, &slide); err != nil {
		return slide, deps, nil, err
	}
	if err := strictInto(json.RawMessage(files["dependencies.json"]), &deps); err != nil {
		return slide, deps, nil, err
	}
	if deps.Schema != finishedDependenciesSchema || !stableID.MatchString(deps.SourceProject) || !shaPattern.MatchString(deps.SourceSHA256) {
		return slide, deps, nil, fmt.Errorf("finished-slide.dependencies_invalid")
	}
	return slide, deps, files, nil
}

func freshFinishedItems(slide *Slide, def wmdesign.LibraryTemplate, remaps map[string]string) error {
	if def.ContentContract == wmdesign.LibraryBindingsContract {
		keys, _ := slide.Values["keys"].(map[string]any)
		for _, array := range def.Arrays {
			values, ok := keys[array.Name].([]any)
			if !ok || len(values) != array.Count {
				return fmt.Errorf("finished-slide.item_keys_invalid")
			}
			for i, old := range values {
				key, ok := old.(string)
				if !ok {
					return fmt.Errorf("finished-slide.item_key_invalid")
				}
				newKey := "reuse-" + nonce()
				remaps["keys/"+array.Name+"/"+key] = newKey
				values[i] = newKey
			}
		}
		return nil
	}
	cards, ok := slide.Values["cards"].([]any)
	if !ok {
		return fmt.Errorf("finished-slide.cards_invalid")
	}
	for _, item := range cards {
		card, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("finished-slide.card_invalid")
		}
		key, ok := card["key"].(string)
		if !ok {
			return fmt.Errorf("finished-slide.card_key_invalid")
		}
		newKey := "reuse-" + nonce()
		remaps["cards/"+key] = newKey
		card["key"] = newKey
	}
	return nil
}

// InsertFinishedSlide validates a closed revision and exact pins, then commits
// the independent slide, owned assets, composition entry and receipt together.
func InsertFinishedSlide(p *Project, o FinishedSlideInsertOptions) (FinishedSlideInsertReceipt, error) {
	r := FinishedSlideInsertReceipt{}
	if !stableID.MatchString(o.ID) || strings.TrimSpace(o.Rationale) == "" {
		return r, fmt.Errorf("finished-slide.insert_requires_fresh_id_and_rationale")
	}
	for _, s := range p.Document.Slides {
		if s.ID == o.ID {
			return r, fmt.Errorf("finished-slide.slide_id_collision: %s", o.ID)
		}
	}
	m, err := finishedslide.Read(o.Package)
	if err != nil {
		return r, err
	}
	if m.Lifecycle == "deprecated" || m.Freshness(time.Now()) == "stale" {
		return r, fmt.Errorf("finished-slide.revision_not_reusable: %s", m.Lifecycle)
	}
	if m.Lifecycle != "approved" && !o.AllowDraft {
		return r, fmt.Errorf("finished-slide.draft_requires_explicit_allow_draft")
	}
	slide, deps, packageFiles, err := readFinishedSource(o.Package, m)
	if err != nil {
		return r, err
	}
	if deps.Year != p.Document.Year {
		return r, fmt.Errorf("finished-slide.year_mismatch: republish after reviewing date-bound content")
	}
	def, err := finishedTemplate(o.Bundle, slide.Template.ID)
	if err != nil {
		return r, err
	}
	if err := finishedSupported(slide, def); err != nil {
		return r, err
	}
	pins, lockRaw, err := finishedPins(p, o.Bundle, o.Engine, def)
	if err != nil {
		return r, err
	}
	if pins != m.Pins {
		return r, fmt.Errorf("finished-slide.incompatible_pins")
	}
	if _, err := Check(p, o.Bundle, o.Engine); err != nil {
		return r, err
	}
	lineage := LibraryLineage{ID: m.ID, Revision: m.Revision, RevisionSHA256: m.RevisionSHA256, SourceProject: deps.SourceProject, SourceSlide: slide.ID, SourceSHA256: deps.SourceSHA256, AssetRemaps: map[string]string{}, ItemRemaps: map[string]string{}, EvidenceRemaps: map[string]string{}, ContentPolicy: "Copy and selected claim/evidence payloads preserved; item, asset and claim identities remapped. New context and any later copy adaptations require project review; library approval does not approve this deck."}
	observed := map[string][]byte{p.Document.Toolchain.Lockfile: lockRaw}
	assets := map[string]Asset{}
	for k, v := range p.Document.Assets {
		assets[k] = v
	}
	files := map[string][]byte{}
	claimRoles := map[string]string{}
	for _, file := range m.Files {
		claimRoles[file.Path] = file.Role
	}
	claimsPath, err := insertFinishedClaims(p, &slide, deps.Claims, packageFiles, files, observed, lineage.EvidenceRemaps, claimRoles)
	if err != nil {
		return r, err
	}
	media, err := finishedMediaSlots(slide, def)
	if err != nil {
		return r, err
	}
	ids := []string{}
	for id := range deps.Assets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		a := deps.Assets[id]
		if !stableID.MatchString(id) || !shaPattern.MatchString(a.SHA256) || a.DerivedFrom != "" || a.DerivationReceipt != "" {
			return r, fmt.Errorf("finished-slide.asset_dependency_invalid: %s", id)
		}
		newID := "reuse-" + nonce()
		if _, ok := assets[newID]; ok {
			return r, fmt.Errorf("finished-slide.asset_id_collision")
		}
		lineage.AssetRemaps[id] = newID
		if a.RegistryID != "" {
			if a.Path != "" {
				return r, fmt.Errorf("finished-slide.asset_dependency_invalid: %s", id)
			}
		} else {
			raw, ok := packageFiles[a.Path]
			if !ok || digest(raw) != a.SHA256 {
				return r, fmt.Errorf("finished-slide.asset_bytes_missing: %s", id)
			}
			mime, err := assetMIME(raw)
			if err != nil {
				return r, err
			}
			ext := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/svg+xml": ".svg"}[mime]
			a.Path = "assets/originals/" + newID + "-" + a.SHA256[:16] + ext
			files[a.Path] = raw
		}
		assets[newID] = a
	}
	for name, value := range media {
		id := strings.TrimPrefix(value, "project:")
		if newID, ok := lineage.AssetRemaps[id]; ok {
			slide.Values["slots"].(map[string]any)[name] = "project:" + newID
		}
	}
	if err := freshFinishedItems(&slide, def, lineage.ItemRemaps); err != nil {
		return r, err
	}
	slide.ID = o.ID
	var source map[string]any
	if err := json.Unmarshal(canonical(slide), &source); err != nil {
		return r, err
	}
	raw, err := MarshalSlideSource(source)
	if err != nil {
		return r, err
	}
	_, err = ReadCompositionLog(p)
	if err != nil {
		return r, err
	}
	logPath, err := compositionPath(p)
	if err != nil {
		return r, err
	}
	if logPath == "" {
		return r, fmt.Errorf("finished-slide.composition_required: author entries for the existing deck before insertion")
	}
	oldLog, err := projectDependency(p, logPath, 16<<20)
	if err != nil {
		return r, err
	}
	observed[logPath] = oldLog
	logDocument, err := sourceYAML(oldLog)
	if err != nil {
		return r, err
	}
	logSlides := mappingNode(logDocument.Content[0], "slides")
	if logSlides == nil || logSlides.Kind != yaml.MappingNode {
		return r, fmt.Errorf("finished-slide.composition_slides_invalid")
	}
	entry, err := editYAMLNode(CompositionEntry{Purpose: m.Purpose, ChosenTemplate: slide.Template.ID, Rationale: o.Rationale, Library: &lineage})
	if err != nil {
		return r, err
	}
	replaceMappingField(logSlides, o.ID, entry)
	logRaw, err := encodeSourceYAML(logDocument)
	if err != nil {
		return r, err
	}
	addition := &reuseAddition{Assets: assets, Files: files, Observed: observed, ClaimsPath: claimsPath, CompositionPath: logPath, Composition: logRaw, Validate: func(candidate *Project) error {
		if err := ValidateEditorial(candidate); err != nil {
			return err
		}
		_, err := Compile(candidate, o.Bundle, o.Engine)
		return err
	}}
	receipt, err := operateSlide(p, SlideOperation{Action: "add", ID: o.ID, Source: raw, Before: o.Before, After: o.After, IntoSection: o.IntoSection, Bundle: o.Bundle, Engine: o.Engine}, addition)
	r.SlideOperationReceipt, r.Library = receipt, lineage
	return r, err
}
