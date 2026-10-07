package wmdesign

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// BrowsingCoverage is a complete inventory, not a visual qualification claim.
type BrowsingCoverage struct {
	FrameMode       string                   `json:"frame_mode"`
	Schema          string                   `json:"schema"`
	SourceRevision  string                   `json:"source_revision"`
	SourceCommit    string                   `json:"source_commit"`
	SourceFiles     []SourceFile             `json:"source_files"`
	Entries         []BrowsingEntry          `json:"entries"`
	FrameAliases    []BrowsingFrameAlias     `json:"frame_aliases"`
	FrameCandidates int                      `json:"frame_candidates"`
	FrameExclusions []BrowsingFrameExclusion `json:"frame_exclusions"`
	Policy          string                   `json:"policy"`
}
type BrowsingEntry struct {
	SlideID      string        `json:"slide_id"`
	Kind         string        `json:"kind"`
	Key          string        `json:"key"`
	Family       string        `json:"family,omitempty"`
	Revision     int           `json:"revision,omitempty"`
	Lifecycle    string        `json:"lifecycle,omitempty"`
	SourceSHA256 string        `json:"source_sha256,omitempty"`
	Frame        *FrameRequest `json:"frame,omitempty"`
}
type BrowsingFrameAlias struct {
	Frame   FrameRequest `json:"frame"`
	SlideID string       `json:"slide_id"`
	Basis   string       `json:"basis"`
}
type BrowsingFrameExclusion struct {
	Frame  FrameRequest `json:"frame"`
	Reason string       `json:"reason"`
}

// BrowsingInformationSlide emits real editable text; it is never reusable client content.
func BrowsingInformationSlide(id, title, copy string) SlideSpec {
	return SlideSpec{ID: id, Frame: FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1}, Title: title, Eyebrow: "Library guide", ContentKind: "synthetic_example", Nodes: []Node{{ID: "instructions", Kind: "text", Style: "body", Text: copy, Rect: Rect{X: 75, Y: 144, W: 810, H: 0}}}}
}

// TemplateBrowsingDocument includes every retained catalog entry and independent
// frame specimens. Template content uses declared synthetic examples unchanged.
func TemplateBrowsingDocument(bundle, override string, year int) (Document, BrowsingCoverage, error) {
	return TemplateBrowsingDocumentWithFrames(bundle, override, year, "exhaustive")
}

// TemplateBrowsingDocumentWithFrames offers catalog-complete browsing or the
// exhaustive independent frame request matrix; neither mode filters templates.
func TemplateBrowsingDocumentWithFrames(bundle, override string, year int, mode string) (Document, BrowsingCoverage, error) {
	if mode != "catalog" && mode != "exhaustive" {
		return Document{}, BrowsingCoverage{}, fmt.Errorf("browsing.frames_requires_catalog_or_exhaustive")
	}
	s, e := Load(bundle, override)
	if e != nil {
		return Document{}, BrowsingCoverage{}, e
	}
	catalog, e := LibraryCatalogFromSource(s)
	if e != nil {
		return Document{}, BrowsingCoverage{}, e
	}
	coverage := BrowsingCoverage{FrameMode: mode, Schema: "pptxgengo.browsing-coverage.v1", SourceRevision: s.Revision, SourceCommit: s.Commit, SourceFiles: s.Files, Policy: "All retained catalog entries, including visibly labeled deprecated entries. Declared template variants appear once each; independent frame specimens enumerate valid geometry combinations, not arbitrary unqualified template/frame recombinations. Native PowerPoint visual/copy-paste qualification pending."}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year, Title: "Complete template library", BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: s.Revision}}
	doc.Sections = []SectionSpec{{ID: "guide", Title: "How to use", BeforeSlideID: "how-to-use"}}
	doc.Slides = append(doc.Slides, BrowsingInformationSlide("how-to-use", "How to use this deck", "Browse family dividers or PowerPoint sections. Copy an entire slide into your presentation with Keep Source Formatting. Replace the illustrative placeholder copy; review layout, evidence, fonts and brand assets in your destination deck. Deprecated templates are retained for reference and labeled: choose active alternatives for new work. Frame examples show valid geometry combinations independently; they do not authorize every template in every frame. Library identifiers and input pins are in slide notes and the coverage manifest."))
	sort.Slice(catalog, func(i, j int) bool {
		if catalog[i].Family != catalog[j].Family {
			return catalog[i].Family < catalog[j].Family
		}
		return catalog[i].Key < catalog[j].Key
	})
	legacy := TemplateReference(year)
	family := ""
	for i, def := range catalog {
		if def.Family != family {
			family = def.Family
			id := fmt.Sprintf("family-%04d", i)
			doc.Slides = append(doc.Slides, BrowsingInformationSlide(id, family, "Editable placeholder templates. Read each slide's notes for its exact template identity, revision and lifecycle."))
			doc.Sections = append(doc.Sections, SectionSpec{ID: id, Title: "Templates: " + family, BeforeSlideID: id})
		}
		if len(def.PendingCapabilities) > 0 {
			return Document{}, coverage, fmt.Errorf("browsing.template_unsupported: %s: %s; complete library cannot omit this entry", def.Key, strings.Join(def.PendingCapabilities, ", "))
		}
		var input BoundSlide
		input.ID = fmt.Sprintf("template-%04d", i)
		input.Template = def.Key
		input.ContentKind = "synthetic_example"
		if def.ContentContract == TemplateBindingsContract {
			for _, old := range legacy.Slides {
				if old.Template == def.Key {
					input.Values = old.Values
					break
				}
			}
			if input.Values == nil {
				return Document{}, coverage, fmt.Errorf("browsing.placeholder_missing: %s", def.Key)
			}
		} else {
			input.Values, e = json.Marshal(libraryExampleValues(def))
			if e != nil {
				return Document{}, coverage, e
			}
		}
		// Bind all retained entries, including deprecated: compilation preserves the
		// lifecycle warning and does not publish these as approved reusable slides.
		slide, _, err := bindLibraryTemplate(def, input)
		if def.ContentContract == TemplateBindingsContract {
			bound, _, err2 := BindTemplates(bundle, override, BoundDocument{Schema: BoundDocumentSchema, Year: year, Slides: []BoundSlide{input}})
			err = err2
			if err == nil {
				slide = bound.Slides[0]
			}
		}
		if err != nil {
			return Document{}, coverage, fmt.Errorf("browsing.template_generation_failed: %s: %w", def.Key, err)
		}
		if def.Status == "deprecated" {
			doc.Slides = append(doc.Slides, BrowsingInformationSlide(input.ID+"-deprecated", "Deprecated template", "The next slide retains deprecated template "+def.Key+". Use an active alternative for new work. Replacement: "+def.ReplacedBy))
		}
		slide.Notes = fmt.Sprintf("Placeholder template: %s\nName: %s\nRevision: %d\nLifecycle: %s\nFamily: %s\nSource SHA-256: %s\nReplace synthetic illustrative copy; destination review required.", def.Key, def.Name, def.Revision, def.Status, def.Family, def.SourceSHA256)
		doc.Slides = append(doc.Slides, slide)
		coverage.Entries = append(coverage.Entries, BrowsingEntry{SlideID: slide.ID, Kind: "template", Key: def.Key, Family: def.Family, Revision: def.Revision, Lifecycle: def.Status, SourceSHA256: def.SourceSHA256})
	}
	var frames []FrameRequest
	var exclusions []BrowsingFrameExclusion
	var candidates int
	if mode == "exhaustive" {
		frames, exclusions, candidates, e = browsingFrameMatrix(s)
	} else {
		frames, e = browsingCatalogFrames(s)
		candidates = len(frames)
	}

	if e != nil {
		return Document{}, coverage, e
	}
	coverage.FrameCandidates = candidates
	coverage.FrameExclusions = exclusions
	divider := BrowsingInformationSlide("frame-divider", "Frames and rails", "Independent editable frame placeholders: rail, footer, split, title allocation, source allocation, density, surface and navigation combinations (coverage mode: "+mode+"). These pages are geometry building blocks, not a qualification of arbitrary template/frame combinations.")
	doc.Slides = append(doc.Slides, divider)
	doc.Sections = append(doc.Sections, SectionSpec{ID: divider.ID, Title: "Frames and rails", BeforeSlideID: divider.ID})
	unique := map[string]string{}
	for _, original := range frames {
		q := original
		if q.NoHeader {
			q.TitleLines = 1
			q.Density = "standard"
		}
		keyRaw, _ := json.Marshal(q)
		key := string(keyRaw)
		if existing, ok := unique[key]; ok {
			coverage.FrameAliases = append(coverage.FrameAliases, BrowsingFrameAlias{Frame: original, SlideID: existing, Basis: "No header: title allocation and header density have no visible or editable output; same normalized resolved geometry, typography and chrome."})
			continue
		}
		i := len(unique)
		unique[key] = fmt.Sprintf("frame-%05d", i)
		if original.NoHeader && (original.TitleLines != q.TitleLines || original.Density != q.Density) {
			coverage.FrameAliases = append(coverage.FrameAliases, BrowsingFrameAlias{Frame: original, SlideID: unique[key], Basis: "No header: title allocation and header density have no visible or editable output."})
		}

		id := fmt.Sprintf("frame-%05d", i)
		f, err := s.ResolveFrame(q)
		if err != nil {
			return Document{}, coverage, err
		}
		title := strings.TrimSuffix(strings.Repeat("Title\n", q.TitleLines), "\n")
		slide := SlideSpec{ID: id, Frame: q, Title: title, Eyebrow: "Frame placeholder", ContentKind: "synthetic_example"}
		if q.NoHeader {
			slide.Title = ""
			slide.Eyebrow = ""
		}
		if q.SourceLines > 0 {
			slide.Source = strings.TrimSuffix(strings.Repeat("Source: review.\n", q.SourceLines), "\n")
		}
		zones := []Rect{f.Body}
		if q.Split != "" {
			zones = []Rect{f.ShortBody, f.TallBody}
		}
		for j, r := range zones {
			slide.Nodes = append(slide.Nodes, Node{ID: fmt.Sprintf("placeholder-%d", j), Kind: "text", Style: "small", Text: "Your content", Rect: Rect{X: r.X + 18, Y: r.Y + 18, W: r.W - 36, H: 36}})
		}
		raw, _ := json.Marshal(q)
		slide.Notes = "Independent frame: " + string(raw)
		doc.Slides = append(doc.Slides, slide)
		copy := q
		keyID := id
		if mode == "catalog" {
			keyID = "wmds/frame/" + q.Rail + "-" + q.Footer
			if q.Split != "" {
				keyID += "/split/" + q.Split
			}
		}
		coverage.Entries = append(coverage.Entries, BrowsingEntry{SlideID: id, Kind: "frame", Key: keyID, Frame: &copy})
	}
	return doc, coverage, nil
}

func browsingFrameMatrix(s *Source) ([]FrameRequest, []BrowsingFrameExclusion, int, error) {
	keys := func(m map[string]Rail) []string {
		a := []string{}
		for k := range m {
			a = append(a, k)
		}
		sort.Strings(a)
		return a
	}
	footers := []string{}
	for k := range s.Frames.Footers {
		footers = append(footers, k)
	}
	sort.Strings(footers)
	splits := []string{""}
	for k := range s.Frames.Splits {
		if strings.HasPrefix(k, "tall-") {
			splits = append(splits, k)
		}
	}
	sort.Strings(splits)
	surfaces := []string{}
	for k := range s.Tokens.Colors.Surfaces {
		surfaces = append(surfaces, k)
	}
	sort.Strings(surfaces)
	max := 2
	if _, ok := s.Frames.Footers["slim"]; ok {
		max = 4
	}
	frames := []FrameRequest{}
	excluded := []BrowsingFrameExclusion{}
	candidates := 0
	for _, rail := range keys(s.Frames.Rails) {
		rs := []string{"inverse"}
		if rail == "left" || rail == "right" {
			rs = append([]string{}, s.Frames.Rails[rail].Surfaces...)
			sort.Strings(rs)
		}
		for _, footer := range footers {
			for _, split := range splits {
				for _, surface := range surfaces {
					for _, railSurface := range rs {
						for lines := 1; lines <= max; lines++ {
							for source := 0; source <= 2; source++ {
								for _, density := range []string{"standard", "appendix"} {
									for _, header := range []bool{false, true} {
										navCounts := []int{0}
										if rail == "nav" {
											navCounts = []int{2, 3, 4, 5, 6}
										}
										for _, tabs := range navCounts {
											q := FrameRequest{Rail: rail, Footer: footer, Split: split, Surface: surface, RailSurface: railSurface, TitleLines: lines, SourceLines: source, Density: density, NoHeader: header}
											for j := 0; j < tabs; j++ {
												q.Nav = append(q.Nav, NavTab{ID: fmt.Sprintf("tab-%d", j+1), Label: fmt.Sprintf("Section %d", j+1)})
											}
											if tabs > 0 {
												q.Active = q.Nav[0].ID
											}
											candidates++
											_, e := s.ResolveFrame(q)
											if e != nil {
												reason := e.Error()
												known := []string{"frame.split_requires_no_panel_rail", "frame.split_requires_standard_header", "frame.appendix_requires_one_title_line", "frame.invalid_line_allocation", "frame.empty_body"}
												ok := false
												for _, prefix := range known {
													ok = ok || strings.HasPrefix(reason, prefix)
												}
												if !ok {
													return nil, nil, candidates, e
												}
												excluded = append(excluded, BrowsingFrameExclusion{Frame: q, Reason: reason})
												continue
											}
											frames = append(frames, q)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if len(frames) == 0 {
		return nil, nil, candidates, fmt.Errorf("browsing.frame_coverage_empty")
	}
	return frames, excluded, candidates, nil
}

func browsingCatalogFrames(source *Source) ([]FrameRequest, error) {
	entities, e := collectFrameEntities(source, map[string]string{})
	if e != nil {
		return nil, e
	}
	frames := make([]FrameRequest, 0, len(entities))
	for _, entity := range entities {
		var def struct {
			Request FrameRequest `json:"default_request"`
		}
		if e = json.Unmarshal(entity.Definition, &def); e != nil {
			return nil, e
		}
		if _, e = source.ResolveFrame(def.Request); e != nil {
			return nil, e
		}
		frames = append(frames, def.Request)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("browsing.catalog_frame_coverage_empty")
	}
	return frames, nil
}
