package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

// LibraryPublicationOptions publishes source specimens with retained native
// evidence. It does not qualify user copy or perform a fresh native export.
type LibraryPublicationOptions struct {
	Bundle          string
	Source          string
	PreviousBundle  string
	PreviousGallery string
	NativeReviews   []string
	Version         string
	Year            int
}

type LibraryPublicationReport struct {
	Schema         string `json:"schema"`
	SourceRevision string `json:"source_revision"`
	Entries        int    `json:"entries"`
	Inherited      int    `json:"inherited_native_previews"`
	NewlyReviewed  int    `json:"newly_reviewed_native_previews"`
}

type publicationDesign struct {
	Purpose             string           `json:"purpose,omitempty"`
	Template            string           `json:"template"`
	Name                string           `json:"name"`
	Family              string           `json:"family"`
	Status              string           `json:"status"`
	SourceCommit        string           `json:"source_commit"`
	SourceRevision      string           `json:"source_revision"`
	SourcePage          int              `json:"source_page"`
	NativeReview        string           `json:"native_review"`
	Contract            string           `json:"contract"`
	SourceFoundation    string           `json:"source_foundation"`
	SourceValues        string           `json:"source_values,omitempty"`
	SourcePreview       string           `json:"source_preview"`
	SourcePreviewSHA256 string           `json:"source_preview_sha256"`
	SlotCount           int              `json:"slot_count"`
	ArrayCount          int              `json:"array_count"`
	Discovery           LibraryDiscovery `json:"discovery"`
	ReplacedBy          string           `json:"replaced_by,omitempty"`
	Evidence            string           `json:"native_evidence"`
	CompositionSHA256   string           `json:"composition_sha256"`
}

type publicationIndex struct {
	Schema         string              `json:"schema"`
	Version        string              `json:"version"`
	SourceRevision string              `json:"source_revision"`
	SourceCommit   string              `json:"source_commit"`
	Entries        int                 `json:"entries"`
	Active         int                 `json:"active"`
	Deprecated     int                 `json:"deprecated"`
	Qualification  map[string]any      `json:"qualification"`
	Designs        []publicationDesign `json:"designs"`
}

type publicationPreview struct {
	path, sha, evidence string
	page                int
	slide               SlideSpec
}

func publicationComposition(slide SlideSpec) ([]byte, error) {
	// Source provenance and enumeration change when snapshots grow; authored
	// frame, text, nodes and source notes must remain exactly identical.
	slide.ID = ""
	slide.TemplateBinding = nil
	raw, err := json.Marshal(slide)
	if err != nil {
		return nil, err
	}
	// Raw scene objects can arrive with a different property order after a
	// compiler repair or JSON round trip. Compare every authored value through
	// canonical object encoding while preserving numeric representations.
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var canonical any
	if err = decoder.Decode(&canonical); err != nil {
		return nil, err
	}
	return json.Marshal(canonical)
}

func samePublicationComposition(a, b SlideSpec) bool {
	ar, ae := publicationComposition(a)
	br, be := publicationComposition(b)
	return ae == nil && be == nil && bytes.Equal(ar, br)
}

func publicationSlides(bundle, source string, year int) (map[string]SlideSpec, error) {
	doc, err := LibrarySourceReference(bundle, source, "", year)
	if err != nil {
		return nil, err
	}
	result := map[string]SlideSpec{}
	for _, slide := range doc.Slides {
		result[slide.TemplateBinding.Template] = slide
	}
	return result, nil
}

func verifiedPublicationFile(path, declared string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if declared == "" || fmt.Sprintf("%x", sha256.Sum256(raw)) != declared {
		return nil, fmt.Errorf("publication.evidence_drift: %s", path)
	}
	return raw, nil
}

func compatiblePublicationStyle(a, b *Source) bool {
	// Components can grow independently; rendered scene nodes and resolved
	// styles/tokens/chrome are what affect retained source specimens.
	af, aok := publicationRenderingFrames(a.Frames)
	bf, bok := publicationRenderingFrames(b.Frames)
	return aok && bok && reflect.DeepEqual(a.Tokens, b.Tokens) && reflect.DeepEqual(af, bf) && reflect.DeepEqual(a.styles, b.styles) && reflect.DeepEqual(a.loadedFontFiles, b.loadedFontFiles)
}

// The documentation source can spell out existing renderer fallbacks. These
// exact redundant entries do not affect ResolveFrame; every other geometry
// field remains part of the dependency comparison. Native inheritance also
// requires unchanged compositions and paired visible-package-part comparison.
func publicationRenderingFrames(frames Frames) (Frames, bool) {
	frames.Features = append(frames.Features[:0:0], frames.Features...)
	for i := range frames.Features {
		feature := &frames.Features[i]
		if len(feature.Geometry) == 0 {
			continue
		}
		var geometry map[string]any
		if err := json.Unmarshal(feature.Geometry, &geometry); err != nil {
			return Frames{}, false
		}
		var redundant map[string]string
		switch feature.ID {
		case "title-zone":
			redundant = map[string]string{
				"threeLine": `{"rule":180,"bodyTop":198}`,
				"fourLine":  `{"rule":216,"bodyTop":234}`,
			}
		case "zone.source":
			redundant = map[string]string{
				"bodyBottomTable": `{"compact":[468,450,432],"tall":[450,450,432],"slim":[486,468,450]}`,
			}
		}
		for key, expectedJSON := range redundant {
			if value, exists := geometry[key]; exists {
				var expected any
				if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil || !reflect.DeepEqual(value, expected) {
					return Frames{}, false
				}
				delete(geometry, key)
			}
		}
		var err error
		feature.Geometry, err = json.Marshal(geometry)
		if err != nil {
			return Frames{}, false
		}
	}
	return frames, true
}

func previousPublicationPreviews(options LibraryPublicationOptions, current *Source) (map[string]publicationPreview, error) {
	result := map[string]publicationPreview{}
	if options.PreviousGallery == "" {
		return result, nil
	}
	if options.PreviousBundle == "" {
		return nil, fmt.Errorf("publication.previous_bundle_required")
	}
	prior, err := Load(options.PreviousBundle, "")
	if err != nil {
		return nil, err
	}
	if !compatiblePublicationStyle(prior, current) {
		return nil, fmt.Errorf("publication.inherited_render_dependencies_changed")
	}
	priorSlides, err := publicationSlides(options.PreviousBundle, "", options.Year)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(options.PreviousGallery, "design-system/index.json"))
	if err != nil {
		return nil, err
	}
	var index publicationIndex
	if err = json.Unmarshal(raw, &index); err != nil {
		return nil, err
	}
	if index.Schema != "pptxgengo.wmds-release-gallery.v1" || index.SourceRevision != prior.Revision || index.SourceCommit != prior.Commit || index.Entries != len(index.Designs) {
		return nil, fmt.Errorf("publication.previous_gallery_source_mismatch")
	}
	for _, entry := range index.Designs {
		if entry.NativeReview != "reviewed_source_specimen" {
			return nil, fmt.Errorf("publication.previous_specimen_unreviewed: %s", entry.Template)
		}
		if _, exists := result[entry.Template]; exists {
			return nil, fmt.Errorf("publication.duplicate_previous_template: %s", entry.Template)
		}
		foundation, err := indexRelative(options.PreviousGallery, entry.SourceFoundation)
		if err != nil {
			return nil, err
		}
		docRaw, err := os.ReadFile(foundation)
		if err != nil {
			return nil, err
		}
		var doc Document
		if err = json.Unmarshal(docRaw, &doc); err != nil {
			return nil, err
		}
		if doc.Schema != "pptxgengo.wmds-foundation.v1" || len(doc.Slides) != 1 || doc.Year != options.Year {
			return nil, fmt.Errorf("publication.previous_foundation_invalid: %s", entry.Template)
		}
		priorSlide, exists := priorSlides[entry.Template]
		if !exists || !samePublicationComposition(priorSlide, doc.Slides[0]) {
			return nil, fmt.Errorf("publication.previous_foundation_source_drift: %s", entry.Template)
		}
		preview, err := indexRelative(options.PreviousGallery, entry.SourcePreview)
		if err != nil {
			return nil, err
		}
		if _, err = verifiedPublicationFile(preview, entry.SourcePreviewSHA256); err != nil {
			return nil, err
		}
		result[entry.Template] = publicationPreview{preview, entry.SourcePreviewSHA256, "inherited_identical_authored_composition", entry.SourcePage, doc.Slides[0]}
	}
	return result, nil
}

func nativePublicationPreviews(manifest string, current *Source, year int) (map[string]publicationPreview, error) {
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}
	var review struct {
		Schema         string          `json:"schema"`
		SourceRevision string          `json:"source_revision"`
		PPTX           string          `json:"pptx"`
		PPTXSHA        string          `json:"pptx_sha256"`
		PDF            json.RawMessage `json:"pdf"`
		PDFSHA         string          `json:"pdf_sha256"`
		Source         struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
		} `json:"source"`
		Review struct {
			Status string `json:"status"`
		} `json:"review"`
		Pages []struct {
			Page     int    `json:"page"`
			Template string `json:"template"`
			PNG      string `json:"png"`
			SHA      string `json:"sha256"`
			Status   string `json:"status"`
		} `json:"pages"`
		PNGs []struct {
			Page     int    `json:"page"`
			Template string `json:"template"`
			Path     string `json:"path"`
			SHA      string `json:"sha256"`
			Status   string `json:"visual_status"`
		} `json:"pngs"`
	}
	if err = json.Unmarshal(raw, &review); err != nil {
		return nil, err
	}
	base := filepath.Dir(manifest)
	// Both persisted intake manifest formats locate the frozen bundle beside
	// native-review. Load verifies its pins before its composition is reused.
	bundle := filepath.Join(base, "..", "bundle")
	source, err := Load(bundle, "")
	if err != nil {
		return nil, err
	}
	if !compatiblePublicationStyle(source, current) {
		return nil, fmt.Errorf("publication.review_render_dependencies_changed: %s", manifest)
	}
	slides, err := publicationSlides(bundle, "", year)
	if err != nil {
		return nil, err
	}
	var pdf, pdfSHA string
	switch review.Schema {
	case "pptxgengo.native-visual-review.v1":
		if review.SourceRevision != source.Revision {
			return nil, fmt.Errorf("publication.review_revision_mismatch")
		}
		if err = json.Unmarshal(review.PDF, &pdf); err != nil {
			return nil, err
		}
		pdfSHA = review.PDFSHA
	case "pptxgengo.manual-native-review.v1":
		if review.Review.Status != "accepted" {
			return nil, fmt.Errorf("publication.native_review_unaccepted")
		}
		review.PPTX, review.PPTXSHA = review.Source.Path, review.Source.SHA
		var p struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
		}
		if err = json.Unmarshal(review.PDF, &p); err != nil {
			return nil, err
		}
		pdf, pdfSHA = p.Path, p.SHA
		for _, p := range review.PNGs {
			review.Pages = append(review.Pages, struct {
				Page     int    `json:"page"`
				Template string `json:"template"`
				PNG      string `json:"png"`
				SHA      string `json:"sha256"`
				Status   string `json:"status"`
			}{p.Page, p.Template, p.Path, p.SHA, p.Status})
		}
	default:
		return nil, fmt.Errorf("publication.unsupported_native_manifest: %s", review.Schema)
	}
	// The deck is a sibling artifact, so its persisted ../ path is intentional.
	acceptedDeck, err := verifiedPublicationFile(filepath.Join(base, review.PPTX), review.PPTXSHA)
	if err != nil {
		return nil, err
	}
	acceptedSlides, err := publicationDeckSlides(acceptedDeck)
	if err != nil {
		return nil, err
	}
	if _, err = verifiedPublicationFile(filepath.Join(base, pdf), pdfSHA); err != nil {
		return nil, err
	}
	result := map[string]publicationPreview{}
	checkedDoc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	for _, page := range review.Pages {
		slide, ok := slides[page.Template]
		if !ok || page.Status != "accepted" || page.Page != len(checkedDoc.Slides)+1 {
			return nil, fmt.Errorf("publication.native_page_invalid: %s", page.Template)
		}
		if _, exists := result[page.Template]; exists {
			return nil, fmt.Errorf("publication.duplicate_native_template: %s", page.Template)
		}
		path, err := indexRelative(base, page.PNG)
		if err != nil {
			return nil, err
		}
		if _, err = verifiedPublicationFile(path, page.SHA); err != nil {
			return nil, err
		}
		slideXML, exists := acceptedSlides[page.Page]
		if !exists {
			return nil, fmt.Errorf("publication.native_deck_page_missing: %d", page.Page)
		}
		slide.ID, err = publicationSlideName(slideXML)
		if err != nil {
			return nil, err
		}
		checkedDoc.Slides = append(checkedDoc.Slides, slide)
		result[page.Template] = publicationPreview{path, page.SHA, "native_gui_review; manifest_sha256=" + fmt.Sprintf("%x", sha256.Sum256(raw)), page.Page, slide}
	}
	if len(checkedDoc.Slides) == 0 || len(checkedDoc.Slides) != len(acceptedSlides) {
		return nil, fmt.Errorf("publication.native_deck_page_count_mismatch")
	}
	rebuilt, _, err := BuildWithEngine(bundle, "", checkedDoc, CandidateEngine)
	if err != nil {
		return nil, err
	}
	rebuiltSlides, err := publicationDeckSlides(rebuilt)
	if err != nil {
		return nil, err
	}
	for page, accepted := range acceptedSlides {
		if !bytes.Equal(accepted, rebuiltSlides[page]) {
			return nil, fmt.Errorf("publication.native_deck_composition_drift: %s page %d", manifest, page)
		}
	}
	if err = comparePublicationVisibleParts(acceptedDeck, rebuilt); err != nil {
		return nil, fmt.Errorf("publication.native_deck_dependency_drift: %s: %w", manifest, err)
	}
	return result, nil
}

func publicationDeckSlides(deck []byte) (map[int][]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		return nil, err
	}
	slides := map[int][]byte{}
	for _, file := range reader.File {
		var page int
		if n, err := fmt.Sscanf(file.Name, "ppt/slides/slide%d.xml", &page); err != nil || n != 1 || file.Name != fmt.Sprintf("ppt/slides/slide%d.xml", page) {
			continue
		}
		stream, err := file.Open()
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(stream)
		closeErr := stream.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		slides[page] = raw
	}
	return slides, nil
}

// VerifyLibraryPublicationRenderInheritance qualifies retained renderer output
// independently of native GUI acceptance. Caller keys must be exactly those
// inherited by a publication; both snapshots build the same source ordering.
func VerifyLibraryPublicationRenderInheritance(options LibraryPublicationOptions, keys []string) (int, error) {
	previous, err := Load(options.PreviousBundle, "")
	if err != nil {
		return 0, err
	}
	current, err := Load(options.Bundle, options.Source)
	if err != nil {
		return 0, err
	}
	if !compatiblePublicationStyle(previous, current) {
		return 0, fmt.Errorf("publication.inherited_render_dependencies_changed")
	}
	oldSlides, err := publicationSlides(options.PreviousBundle, "", options.Year)
	if err != nil {
		return 0, err
	}
	newSlides, err := publicationSlides(options.Bundle, options.Source, options.Year)
	if err != nil {
		return 0, err
	}
	oldDoc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: options.Year, BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "library-publication-inheritance"}}
	newDoc := oldDoc
	seen := map[string]bool{}
	for i, key := range keys {
		old, oldOK := oldSlides[key]
		current, newOK := newSlides[key]
		if !oldOK || !newOK || seen[key] || !samePublicationComposition(old, current) {
			return 0, fmt.Errorf("publication.inheritance_composition_invalid: %s", key)
		}
		seen[key] = true
		old.ID = fmt.Sprintf("inherit-%03d", i+1)
		current.ID = old.ID
		oldDoc.Slides = append(oldDoc.Slides, old)
		newDoc.Slides = append(newDoc.Slides, current)
	}
	if len(keys) == 0 {
		return 0, fmt.Errorf("publication.inheritance_keys_required")
	}
	oldDeck, _, err := BuildWithEngine(options.PreviousBundle, "", oldDoc, CandidateEngine)
	if err != nil {
		return 0, err
	}
	newDeck, _, err := BuildWithEngine(options.Bundle, options.Source, newDoc, CandidateEngine)
	if err != nil {
		return 0, err
	}
	if err = comparePublicationVisibleParts(oldDeck, newDeck); err != nil {
		return 0, fmt.Errorf("publication.inherited_render_drift: %w", err)
	}
	return len(keys), nil
}

func publicationDeckParts(deck []byte) (map[string][]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		return nil, err
	}
	parts := map[string][]byte{}
	for _, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(stream)
		closeErr := stream.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		parts[file.Name] = raw
	}
	return parts, nil
}

type publicationRelationship struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
	Mode   string `xml:"TargetMode,attr"`
}

func publicationVisibleRelationships(parts map[string][]byte, part string) ([]publicationRelationship, error) {
	relPath := path.Join(path.Dir(part), "_rels", path.Base(part)+".rels")
	raw := parts[relPath]
	if len(raw) == 0 {
		return nil, nil
	}
	var document struct {
		Relationships []publicationRelationship `xml:"Relationship"`
	}
	if err := xml.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	var result []publicationRelationship
	for _, rel := range document.Relationships {
		if !strings.HasSuffix(rel.Type, "/notesSlide") {
			result = append(result, rel)
		}
	}
	return result, nil
}

// OPC relationship targets beginning with / address the package root;
// relative targets address the owning part's directory. path.Join alone
// does not discard its first argument when the second begins with /.
func publicationRelationshipPart(owner, target string) string {
	if strings.HasPrefix(target, "/") {
		return path.Clean(strings.TrimPrefix(target, "/"))
	}
	return path.Clean(path.Join(path.Dir(owner), target))
}

// Follow visible slide dependency graphs, including layouts, themes, charts,
// workbooks and registered images. Binding audit notes and core timestamps do
// not change source specimen appearance and are excluded from this comparison.
func comparePublicationVisibleParts(a, b []byte) error {
	left, err := publicationDeckParts(a)
	if err != nil {
		return err
	}
	right, err := publicationDeckParts(b)
	if err != nil {
		return err
	}
	checked := map[string]bool{}
	var compare func(string, string) error
	compare = func(lp, rp string) error {
		key := lp + "\x00" + rp
		if checked[key] {
			return nil
		}
		checked[key] = true
		l, lOK := left[lp]
		r, rOK := right[rp]
		if !lOK || !rOK {
			return fmt.Errorf("visible_part_missing: %s / %s", lp, rp)
		}
		if !bytes.Equal(l, r) {
			return fmt.Errorf("visible_part_changed: %s", lp)
		}
		lr, err := publicationVisibleRelationships(left, lp)
		if err != nil {
			return err
		}
		rr, err := publicationVisibleRelationships(right, rp)
		if err != nil {
			return err
		}
		if len(lr) != len(rr) {
			return fmt.Errorf("visible_relationship_count_changed: %s", lp)
		}
		for i, lrel := range lr {
			rrel := rr[i]
			if lrel.ID != rrel.ID || lrel.Type != rrel.Type || lrel.Mode != rrel.Mode {
				return fmt.Errorf("visible_relationship_changed: %s", lp)
			}
			if lrel.Mode == "External" {
				if lrel.Target != rrel.Target {
					return fmt.Errorf("external_relationship_changed: %s", lp)
				}
				continue
			}
			lt := publicationRelationshipPart(lp, lrel.Target)
			rt := publicationRelationshipPart(rp, rrel.Target)
			if err := compare(lt, rt); err != nil {
				return err
			}
		}
		return nil
	}
	slides, err := publicationDeckSlides(a)
	if err != nil {
		return err
	}
	other, err := publicationDeckSlides(b)
	if err != nil {
		return err
	}
	if len(slides) != len(other) {
		return fmt.Errorf("visible_slide_count_changed")
	}
	for page := range slides {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", page)
		if err := compare(part, part); err != nil {
			return err
		}
	}
	return nil
}

func publicationSlideName(raw []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", fmt.Errorf("publication.native_slide_name_missing: %w", err)
		}
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "cSld" {
			for _, attr := range start.Attr {
				if attr.Name.Local == "name" && attr.Value != "" {
					return attr.Value, nil
				}
			}
		}
	}
}

// PublishLibraryGallery writes a new directory atomically. Every preview needs
// accepted native evidence for an identical source composition; contracts and
// authoring fixtures are freshly derived from the requested pinned snapshot.
func PublishLibraryGallery(out string, options LibraryPublicationOptions) (report LibraryPublicationReport, err error) {
	if options.Year < 2000 || options.Year > 9999 || options.Version == "" || out == "" {
		return report, fmt.Errorf("publication.output_version_year_required")
	}
	if _, err = os.Lstat(out); !os.IsNotExist(err) {
		return report, fmt.Errorf("publication.output_must_not_exist: %s", out)
	}
	source, err := Load(options.Bundle, options.Source)
	if err != nil {
		return report, err
	}
	catalog, err := LibraryCatalogFromSource(source)
	if err != nil {
		return report, err
	}
	slides, err := publicationSlides(options.Bundle, options.Source, options.Year)
	if err != nil {
		return report, err
	}
	previews, err := previousPublicationPreviews(options, source)
	if err != nil {
		return report, err
	}
	nativeSeen := map[string]bool{}
	for _, manifest := range options.NativeReviews {
		reviewed, e := nativePublicationPreviews(manifest, source, options.Year)
		if e != nil {
			return report, e
		}
		for key, preview := range reviewed {
			if nativeSeen[key] {
				return report, fmt.Errorf("publication.duplicate_reviewed_template: %s", key)
			}
			nativeSeen[key] = true
			previews[key] = preview
		}
	}
	parent := filepath.Dir(out)
	if err = os.MkdirAll(parent, 0755); err != nil {
		return report, err
	}
	stage, err := os.MkdirTemp(parent, ".library-publish-")
	if err != nil {
		return report, err
	}
	defer os.RemoveAll(stage)
	index := publicationIndex{Schema: "pptxgengo.wmds-release-gallery.v1", Version: options.Version, SourceRevision: source.Revision, SourceCommit: source.Commit, Entries: len(catalog), Qualification: map[string]any{"reviewed_source_specimens": len(catalog), "arbitrary_content_qualified": false, "scope": "Individual source specimens only. Supplied copy requires a separate fit/build and native review."}}
	report = LibraryPublicationReport{Schema: "pptxgengo.library-publication.v1", SourceRevision: source.Revision, Entries: len(catalog)}
	references, err := LibraryReference(options.Bundle, options.Source, "", options.Year)
	if err != nil {
		return report, err
	}
	values := map[string]BoundSlide{}
	for _, slide := range references.Slides {
		values[slide.Template] = slide
	}
	for i, def := range catalog {
		if !filepath.IsLocal(def.Key) || strings.Contains(def.Key, "\\") {
			return report, fmt.Errorf("publication.invalid_template_key: %s", def.Key)
		}
		slide := slides[def.Key]
		preview, ok := previews[def.Key]
		if !ok || !samePublicationComposition(slide, preview.slide) {
			return report, fmt.Errorf("publication.identical_review_required: %s", def.Key)
		}
		rel := filepath.ToSlash(filepath.Join("design-system", def.Key))
		dir := filepath.Join(stage, rel)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return report, err
		}
		if err = WriteJSON(filepath.Join(dir, "contract.json"), def); err != nil {
			return report, err
		}
		if err = WriteJSON(filepath.Join(dir, "source.foundation.json"), Document{Schema: "pptxgengo.wmds-foundation.v1", Year: options.Year, Slides: []SlideSpec{slide}}); err != nil {
			return report, err
		}
		raw, e := verifiedPublicationFile(preview.path, preview.sha)
		if e != nil {
			return report, e
		}
		if err = os.WriteFile(filepath.Join(dir, "source.png"), raw, 0644); err != nil {
			return report, err
		}
		composition, e := publicationComposition(slide)
		if e != nil {
			return report, e
		}
		entry := publicationDesign{Template: def.Key, Name: def.Name, Purpose: def.Purpose, Family: def.Family, Status: def.Status, SourceCommit: source.Commit, SourceRevision: source.Revision, SourcePage: i + 1, NativeReview: "reviewed_source_specimen", Contract: rel + "/contract.json", SourceFoundation: rel + "/source.foundation.json", SourcePreview: rel + "/source.png", SourcePreviewSHA256: preview.sha, SlotCount: len(def.Slots), ArrayCount: len(def.Arrays), Discovery: def.Discovery, ReplacedBy: def.ReplacedBy, Evidence: preview.evidence, CompositionSHA256: fmt.Sprintf("%x", sha256.Sum256(composition))}
		if reference, exists := values[def.Key]; exists {
			entry.SourceValues = rel + "/source-values.json"
			if err = WriteJSON(filepath.Join(dir, "source-values.json"), BoundDocument{Schema: BoundDocumentSchema, Year: options.Year, Slides: []BoundSlide{reference}}); err != nil {
				return report, err
			}
		}
		if def.Status == "deprecated" {
			index.Deprecated++
		} else {
			index.Active++
		}
		if preview.evidence == "inherited_identical_authored_composition" {
			report.Inherited++
		} else {
			report.NewlyReviewed++
		}
		index.Designs = append(index.Designs, entry)
	}
	if err = WriteJSON(filepath.Join(stage, "design-system/index.json"), index); err != nil {
		return report, err
	}
	if err = WriteJSON(filepath.Join(stage, "index.json"), map[string]any{"schema": "pptxgengo.release-catalog.v2", "version": options.Version, "design_system": map[string]any{"entries": index.Entries, "active": index.Active, "deprecated": index.Deprecated, "source_revision": source.Revision, "source_commit": source.Commit, "qualification": index.Qualification, "gallery": "design-system.html", "index": "design-system/index.json"}}); err != nil {
		return report, err
	}
	if err = WriteJSON(filepath.Join(stage, "publication-report.json"), report); err != nil {
		return report, err
	}
	html, err := os.Create(filepath.Join(stage, "design-system.html"))
	if err != nil {
		return report, err
	}
	tmpl, err := template.New("gallery").Parse(publicationGalleryHTML)
	if err == nil {
		err = tmpl.Execute(html, index)
	}
	closeErr := html.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return report, err
	}
	if err = os.Rename(stage, out); err != nil {
		return report, err
	}
	return report, nil
}

const publicationGalleryHTML = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>West Monroe design library</title><style>body{margin:0;background:#f5f5f3;color:#172c35;font:16px system-ui}header{padding:32px;background:#fff;position:sticky;top:0;z-index:1;border-bottom:1px solid #ccd3d4}h1{margin:0 0 12px;font-size:28px}p{margin:8px 0}input,select{font:inherit;padding:9px;border:1px solid #a7b4b8;border-radius:6px}input{width:min(65%,600px)}main{padding:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(310px,1fr));gap:20px}article{background:white;border:1px solid #ccd3d4;border-radius:8px;overflow:hidden}article[hidden]{display:none}img{display:block;width:100%;aspect-ratio:16/9;object-fit:contain;background:#fff}section{padding:16px}h2{font-size:18px;margin:0 0 8px}small{color:#486069}a{color:#086676}nav{display:flex;gap:12px;flex-wrap:wrap;margin-top:12px}</style><header><h1>West Monroe design library</h1><p>{{.Entries}} source specimens · {{.SourceRevision}}</p><p>Native review covers these source specimens. Supplied copy requires fit, build and native review.</p><input id="query" placeholder="Search names, families or template keys" aria-label="Search templates"> <select id="status" aria-label="Template status"><option value="">All statuses</option><option value="active">Active</option><option value="deprecated">Deprecated</option></select> <select id="family" aria-label="Template family"><option value="">All families</option></select> <small id="count"></small></header><main>{{range .Designs}}<article data-search="{{.Template}} {{.Name}} {{.Family}} {{.Purpose}} {{range .Discovery.ContentRoles}}{{.}} {{end}}{{range .Discovery.VisualForms}}{{.}} {{end}}" data-family="{{.Family}}" data-status="{{.Status}}"><a href="{{.SourcePreview}}" target="_blank"><img src="{{.SourcePreview}}" loading="lazy" alt="{{.Name}}"></a><section><h2>{{.Name}}</h2><small>{{.Template}} · {{.Status}}</small>{{if .Purpose}}<p>{{.Purpose}}</p>{{end}}{{if .ReplacedBy}}<p>Replaced by {{.ReplacedBy}}</p>{{end}}<nav><a href="{{.Contract}}">Contract</a><a href="{{.SourceFoundation}}">Foundation</a>{{if .SourceValues}}<a href="{{.SourceValues}}">Example values</a>{{end}}</nav></section></article>{{end}}</main><script>const q=document.querySelector('#query'),s=document.querySelector('#status'),f=document.querySelector('#family'),cards=[...document.querySelectorAll('article')];for(const family of [...new Set(cards.map(c=>c.dataset.family))].sort()){const o=document.createElement('option');o.value=family;o.textContent=family;f.append(o)}function filter(){let n=0;for(const c of cards){c.hidden=!(c.dataset.search.toLowerCase().includes(q.value.toLowerCase())&&(!s.value||c.dataset.status===s.value)&&(!f.value||c.dataset.family===f.value));if(!c.hidden)n++}document.querySelector('#count').textContent=n+' shown'}q.addEventListener('input',filter);s.addEventListener('change',filter);f.addEventListener('change',filter);filter();</script></html>`
