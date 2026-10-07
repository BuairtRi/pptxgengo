package deckproject

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

// CompileBrowsingRevision compiles immutable authored content without insertion
// or project mutation. Original compiler identity is retained in its manifest;
// regeneration verifies current bundle/template pins and records current CI
// compiler separately, rather than claiming the new binary is the old compiler.
func CompileBrowsingRevision(root, bundle, engine, id string, m finishedslide.Manifest) (Compilation, error) {
	c := Compilation{}
	if m.Lifecycle != "approved" || m.Approval == nil {
		return c, fmt.Errorf("browsing.approved_revision_required")
	}
	actual, e := finishedslide.Read(root)
	if e != nil {
		return c, e
	}
	if !reflect.DeepEqual(actual, m) {
		return c, fmt.Errorf("browsing.revision_changed")
	}
	slide, deps, payload, e := readFinishedSource(root, m)
	if e != nil {
		return c, e
	}
	def, e := finishedTemplate(bundle, m.Pins.TemplateID)
	if e != nil {
		return c, e
	}
	if e = finishedSupported(slide, def); e != nil {
		return c, e
	}
	var tree any
	if e = json.Unmarshal(def.RawSlide, &tree); e != nil {
		return c, e
	}
	lock, e := makeLock(bundle, engine)
	if e != nil {
		return c, e
	}
	if m.Pins.Bundle != lock.BundleSHA256 || m.Pins.SourceRevision != def.SourceRevision || m.Pins.TemplateRevision != def.Revision || m.Pins.TemplateSourceSHA256 != def.SourceSHA256 || m.Pins.TemplateDefinitionSHA256 != digest(canonical(tree)) {
		return c, fmt.Errorf("browsing.revision_pins_mismatch: %s@%d", m.ID, m.Revision)
	}
	if slide.Template.Scope != "shared" || slide.Template.ID != def.Key || slide.Template.Revision != strconv.Itoa(def.Revision) {
		return c, fmt.Errorf("browsing.revision_template_mismatch")
	}
	media, e := finishedMediaSlots(slide, def)
	if e != nil {
		return c, e
	}
	remapped := map[string]Asset{}
	overrides := map[string][]byte{}
	for key, asset := range deps.Assets {
		if asset.DerivedFrom != "" {
			var receipt DerivationReceipt
			if e := strictInto(json.RawMessage(payload[asset.DerivationReceipt]), &receipt); e != nil {
				return c, e
			}
			if receipt.SourceAsset != asset.DerivedFrom {
				return c, fmt.Errorf("browsing.derivation_source_mismatch")
			}
			receipt.SourceAsset = id + "-" + receipt.SourceAsset
			overrides[asset.DerivationReceipt] = canonical(receipt)
			asset.DerivedFrom = id + "-" + asset.DerivedFrom
		}
		remapped[id+"-"+key] = asset
	}
	for name, value := range media {
		key := strings.TrimPrefix(value, "project:")
		if _, ok := deps.Assets[key]; ok {
			slide.Values["slots"].(map[string]any)[name] = "project:" + id + "-" + key
		}
	}
	deps.Assets = remapped
	slide.ID = id
	slide.Hidden = false
	p := &Project{Root: root, Document: Document{Schema: Schema, ID: "browsing-library", Title: "Reusable slides", Year: deps.Year, Assets: deps.Assets, Slides: []Slide{slide}}, Positions: map[string]Position{}, positionFiles: map[string]string{}, sourceOverrides: overrides}
	p.Canonical = canonical(p.Document)
	c, e = Compile(p, bundle, engine)
	if e != nil {
		return c, e
	}
	s := &c.Document.Slides[0]
	s.Notes = strings.TrimSpace(s.Notes) + fmt.Sprintf("\nReusable identity: %s\nRevision: %d\nRevision SHA-256: %s\nApproved by: %s on %s\nReuse scope: %s\nReviewed: %s\nValid until: %s\nOwner: %s\nDestination content/evidence/layout review remains required.", m.ID, m.Revision, m.RevisionSHA256, m.Approval.By, m.Approval.Date, m.Approval.ReuseScope, m.ReviewedAt, m.ValidUntil, m.Owner)
	return c, nil
}

// BrowsingBundleSHA256 identifies the exact authoring bundle independently of
// the original approved revision's compiler/OS. It does not grant approval.
func BrowsingBundleSHA256(bundle, engine string) (string, error) {
	lock, err := makeLock(bundle, engine)
	if err != nil {
		return "", err
	}
	return lock.BundleSHA256, nil
}
