package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/browsingartifact"
	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runBrowsingLibrary(args []string) (err error) {
	f := flag.NewFlagSet("browsing-library", flag.ContinueOnError)
	engine := f.String("engine", wmdesign.CandidateEngine, "required source-pinned candidate engine")
	frames := f.String("frames", "catalog", "templates: catalog frame variants or exhaustive valid request combinations")
	kind := f.String("kind", "", "templates or reusable")
	bundle := f.String("bundle", currentDesignBundle, "pinned bundle")
	source := f.String("source", "", "exact pinned source override (templates only)")
	library := f.String("finished-library", "", "private closed immutable revisions (reusable only)")
	out := f.String("out", "", "new output directory")
	asOf := f.String("as-of", "", "explicit UTC date YYYY-MM-DD for freshness and legal year")
	if err = f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *out == "" || (*kind != "templates" && *kind != "reusable") {
		return fmt.Errorf("browsing-library requires --kind templates|reusable --as-of YYYY-MM-DD --out NEW-DIR")
	}
	if *engine != wmdesign.CandidateEngine {
		return fmt.Errorf("browsing-library requires --engine %s", wmdesign.CandidateEngine)
	}
	date, e := time.Parse(time.DateOnly, *asOf)
	if e != nil {
		return fmt.Errorf("browsing-library requires explicit --as-of YYYY-MM-DD")
	}
	if validLockedBundle(*bundle) {
		*bundle = designBundlePath(*bundle)
	}
	if *kind == "reusable" && *frames != "catalog" {
		return fmt.Errorf("--frames is only supported for template libraries")
	}
	if *kind == "templates" && *library != "" || *kind == "reusable" && (*library == "" || *source != "") {
		return fmt.Errorf("--finished-library required only for reusable; --source only for templates")
	}
	if _, e = os.Lstat(*out); !os.IsNotExist(e) {
		return fmt.Errorf("output directory must be new: %s", *out)
	}
	var doc wmdesign.Document
	var manifest any
	assets := map[string]wmdesign.AssetData{}
	name := "template-library.pptx"
	if *kind == "templates" {
		doc, manifest, err = wmdesign.TemplateBrowsingDocumentWithFrames(*bundle, *source, date.Year(), *frames)
		if err != nil {
			return err
		}
	} else {
		name = "reusable-slides.pptx"
		selection, e := wmdesign.SelectBrowsingRevisions(*library, date)
		if e != nil {
			return e
		}
		manifest = selection
		doc = wmdesign.Document{Schema: "pptxgengo.wmds-foundation.v1", Year: date.Year(), Title: "Latest approved reusable slides", BuildIdentity: &wmdesign.BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: selection.LibrarySHA256}}
		doc.Sections = []wmdesign.SectionSpec{{ID: "guide", Title: "How to use", BeforeSlideID: "how-to-use"}}
		doc.Slides = append(doc.Slides, wmdesign.BrowsingInformationSlide("how-to-use", "How to use this deck", "Copy an approved content slide into your presentation using Keep Source Formatting. The metadata page immediately before it records identity, revision, approval scope and expiry; retain those facts with your project. Approval applies to the recorded reuse scope, not to your destination deck. Review evidence, dates, layout, fonts and any adapted copy. Only latest approved nonexpired revisions are included; newer withdrawals prevent older content from returning."))
		for i, r := range selection.Revisions {
			if !r.Included {
				continue
			}
			id := fmt.Sprintf("reuse-%04d", i)
			m := r.Manifest
			c, e := deckproject.CompileBrowsingRevision(filepath.Join(*library, filepath.FromSlash(r.Path)), *bundle, wmdesign.CandidateEngine, id, m)
			if e != nil {
				return fmt.Errorf("%s@%d: %w", m.ID, m.Revision, e)
			}
			if c.Document.Year != date.Year() {
				return fmt.Errorf("browsing.year_mismatch: %s: republish and review date-bound content for %d", m.ID, date.Year())
			}
			metadata := wmdesign.BrowsingInformationSlide(id+"-metadata", m.Name, fmt.Sprintf("%s · revision %d\nApproved: %s on %s\nReuse scope: %s\nOwner: %s\nFreshness: %s · valid until: %s\nThe next slide is editable authored content; destination review is required.", m.ID, m.Revision, m.Approval.By, m.Approval.Date, m.Approval.ReuseScope, m.Owner, m.Freshness(date), m.ValidUntil))
			doc.Slides = append(doc.Slides, metadata)
			doc.Sections = append(doc.Sections, wmdesign.SectionSpec{ID: id, Title: fmt.Sprintf("%s · %d", strings.TrimPrefix(m.ID, "curated/slide/"), m.Revision), BeforeSlideID: metadata.ID})
			doc.Slides = append(doc.Slides, c.Document.Slides...)
			for key, a := range c.Assets {
				if _, exists := assets[key]; exists {
					return fmt.Errorf("browsing.asset_collision: %s", key)
				}
				assets[key] = a
			}
		}
	}
	deck, layout, e := wmdesign.BuildWithEngineAndAssets(*bundle, *source, doc, wmdesign.CandidateEngine, assets)
	if e != nil {
		return e
	}
	// Publish atomically into a newly owned directory after all slides succeeded.
	parent := filepath.Dir(*out)
	if e = os.MkdirAll(parent, 0755); e != nil {
		return e
	}
	stage, e := os.MkdirTemp(parent, ".browsing-stage-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	bundleHash, e := deckproject.BrowsingBundleSHA256(*bundle, wmdesign.CandidateEngine)
	if e != nil {
		return e
	}
	sum := sha256.Sum256(deck)
	for i := range layout.Fonts {
		layout.Fonts[i].File = filepath.Base(layout.Fonts[i].File)
	}
	pages := make([]browsingartifact.Page, 0, len(doc.Slides))
	for _, slide := range doc.Slides {
		role := ""
		switch {
		case slide.ID == "how-to-use":
			role = "guide"
		case *kind == "reusable" && strings.HasSuffix(slide.ID, "-metadata"):
			role = "reuse_metadata"
		case *kind == "reusable":
			role = "reusable_slide"
		case strings.HasPrefix(slide.ID, "family-"):
			role = "family_divider"
		case strings.HasSuffix(slide.ID, "-deprecated"):
			role = "deprecation_notice"
		case slide.ID == "frame-divider":
			role = "frame_divider"
		case strings.HasPrefix(slide.ID, "template-"):
			role = "template"
		case strings.HasPrefix(slide.ID, "frame-"):
			role = "frame"
		default:
			return fmt.Errorf("browsing.page_role_missing: %s", slide.ID)
		}
		pages = append(pages, browsingartifact.Page{ID: slide.ID, Kind: role})
	}
	output := struct {
		Pages           []browsingartifact.Page `json:"pages"`
		BundleSHA256    string                  `json:"bundle_sha256"`
		SourceRevision  string                  `json:"source_revision"`
		SourceCommit    string                  `json:"source_commit"`
		Schema          string                  `json:"schema"`
		Kind            string                  `json:"kind"`
		AsOf            string                  `json:"as_of"`
		DeckSHA256      string                  `json:"deck_sha256"`
		Compiler        string                  `json:"compiler"`
		ReleaseIdentity string                  `json:"release_identity"`
		Slides          int                     `json:"slides"`
		Coverage        any                     `json:"coverage"`
		SourceFiles     []wmdesign.SourceFile   `json:"source_files"`
		Fonts           []wmdesign.FontIdentity `json:"fonts"`
		Assets          []wmdesign.SourceFile   `json:"assets"`
		Qualification   string                  `json:"qualification"`
	}{Pages: pages, BundleSHA256: bundleHash, SourceRevision: layout.SourceRevision, SourceCommit: layout.SourceCommit, Schema: "pptxgengo.browsing-library.v1", Kind: *kind, AsOf: *asOf, DeckSHA256: fmt.Sprintf("%x", sum), Compiler: wmdesign.CandidateEngine, ReleaseIdentity: releaseIdentity, Slides: len(doc.Slides), Coverage: manifest, SourceFiles: layout.SourceFiles, Fonts: layout.Fonts, Assets: layout.Assets, Qualification: "native_visual_copy_paste_qualification_pending"}
	for _, a := range []struct {
		name  string
		value any
	}{{"compiled-document.json", doc}, {"layout-report.json", layout}} {
		if e = wmdesign.WriteJSON(filepath.Join(stage, a.name), a.value); e != nil {
			return e
		}
	}
	if e = os.WriteFile(filepath.Join(stage, name), deck, 0644); e != nil {
		return e
	}
	manifestBytes, e := json.Marshal(output)
	if e != nil {
		return e
	}
	if len(manifestBytes) > browsingartifact.MaxManifestBytes {
		return fmt.Errorf("browsing manifest exceeds bounded16MiB contract")
	}
	if e = browsingartifact.ValidateFile(manifestBytes, *kind, fmt.Sprintf("%x", sum), filepath.Join(stage, name)); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(stage, "browsing-manifest.json"), append(manifestBytes, '\n'), 0644); e != nil {
		return e
	}
	if _, e = os.Lstat(*out); !os.IsNotExist(e) {
		return fmt.Errorf("output appeared during generation")
	}
	if e = os.Rename(stage, *out); e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"deck": filepath.Join(*out, name), "slides": len(doc.Slides), "sha256": fmt.Sprintf("%x", sum), "native_visual_qualification": "pending"})
}
