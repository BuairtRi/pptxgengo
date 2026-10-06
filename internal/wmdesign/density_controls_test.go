package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/pptx"
)

type densityControl struct {
	ID    string       `json:"id"`
	Roles []string     `json:"roles"`
	Style Style        `json:"style"`
	Font  FontIdentity `json:"font"`
	Rect  Rect         `json:"rect"`
	Page  int          `json:"page"`
	Text  string       `json:"text"`
}

// Quote scales are authored by the designer. Capture only missing exact pairs;
// existing native evidence and the historical base remain immutable.
func TestWriteDensityQuoteNativeControls(t *testing.T) {
	out := os.Getenv("WMDS_DENSITY_QUOTE_OUT")
	if out == "" {
		t.Skip("set WMDS_DENSITY_QUOTE_OUT for quote-role native fixtures")
	}
	source := densityTestSource(t)
	typography, err := NewSourceTypographyEngine(source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	var controls []densityControl
	seen := map[string]bool{}
	page, y := 1, 36.0
	for _, level := range source.Tokens.Density.Levels {
		r := renderer{source: source, bodyDensity: level}
		for _, role := range []string{"card-quote", "pullquote"} {
			st, e := r.bodyStyle("heading")
			if e != nil {
				t.Fatal(e)
			}
			ref, e := r.bodyStyle("subhead")
			if e != nil {
				t.Fatal(e)
			}
			size, ratio := math.Max(40, ref.Size*48/18), .75
			if role == "pullquote" {
				size, ratio = math.Max(40, st.Size*2.5), .5
			}
			st, e = primitiveStyleSize(st, size)
			if e != nil {
				t.Fatal(e)
			}
			st.Leading, st.Case = size*ratio, ""
			font, e := typography.Resolve(st)
			if e != nil {
				t.Fatal(e)
			}
			key := anchorKey(font.SHA256, st.Size, st.Leading)
			if _, exists := typography.anchors[key]; exists || seen[key] {
				continue
			}
			seen[key] = true
			h := st.Leading + math.Max(st.Leading, 1.5*st.Size) + 12
			if y+h > 504 {
				page++
				y = 36
			}
			id := fmt.Sprintf("density-quote-control-%03d", len(controls)+1)
			st.ID = id
			controls = append(controls, densityControl{ID: id, Roles: []string{role + "/" + level}, Style: st, Font: font, Rect: Rect{57, y, 846, h}, Page: page, Text: "Agyp Q012\nAgyp Q012"})
			y += h + 12
		}
	}
	if len(controls) == 0 {
		t.Fatal("no new exact quote pairs")
	}
	deck, report := densityBareNativeControls(t, source, typography, controls)
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "bare-controls.pptx"), deck, 0644); err != nil {
		t.Fatal(err)
	}
	manifest := struct {
		Schema       string           `json:"schema"`
		SourceCommit string           `json:"source_commit"`
		BareSHA      string           `json:"bare_pptx_sha256"`
		Controls     []densityControl `json:"controls"`
	}{"pptxgengo.wmds-density-quote-native-controls.v1", source.Commit, report.PPTXSHA256, controls}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "control-manifest.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %d exact quote pairs on %d bare pages to %s", len(controls), page, out)
}

// Capture only the newly introduced role-ratio pairs. The original 250-control
// deck and manifest are read as immutable evidence and never overwritten.
func TestWriteDensityExtraNativeControls(t *testing.T) {
	out := os.Getenv("WMDS_DENSITY_EXTRA_OUT")
	if out == "" {
		t.Skip("set WMDS_DENSITY_EXTRA_OUT for new-role native fixtures")
	}
	basePath := os.Getenv("WMDS_DENSITY_BASE_MANIFEST")
	if basePath == "" {
		basePath = "/private/tmp/pptxgengo-density-controls/font-controls.manifest.json"
	}
	data, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatal(err)
	}
	var base struct {
		SourceCommit string           `json:"source_commit"`
		BareSHA      string           `json:"bare_pptx_sha256"`
		Controls     []densityControl `json:"controls"`
	}
	if err = json.Unmarshal(data, &base); err != nil {
		t.Fatal(err)
	}
	if len(base.Controls) != 250 || len(base.BareSHA) != 64 {
		t.Fatal("original control identity/count missing")
	}
	source := densityTestSource(t)
	if !densityRoleCorrections(source) {
		t.Fatal("new role-token source required")
	}
	typography, err := NewTypographyEngine(filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, control := range base.Controls {
		seen[anchorKey(control.Font.SHA256, control.Style.Size, control.Style.Leading)] = true
	}
	pairs := map[string]*densityControl{}
	add := func(st Style, role string) {
		st.Case = ""
		for _, weight := range []int{400, 500, 600, 700} {
			variant := st
			variant.Weight = weight
			font, err := typography.Resolve(variant)
			if err != nil {
				t.Fatal(err)
			}
			key := anchorKey(font.SHA256, variant.Size, variant.Leading)
			if _, known := typography.anchors[key]; known || seen[key] {
				continue
			}
			if old := pairs[key]; old != nil {
				old.Roles = append(old.Roles, fmt.Sprintf("%s/%d", role, weight))
				continue
			}
			pairs[key] = &densityControl{Style: variant, Font: font, Roles: []string{fmt.Sprintf("%s/%d", role, weight)}, Text: "Agyp Q012\nAgyp Q012"}
		}
	}
	for _, level := range source.Tokens.Density.Levels {
		r := renderer{source: source, bodyDensity: level}
		for _, role := range []string{"body", "small"} {
			number, numberErr := r.bodyStyle("number")
			metrics, metricsErr := r.bodyStyle(role)
			if numberErr != nil || metricsErr != nil {
				t.Fatalf("strongnum metrics: %v %v", numberErr, metricsErr)
			}
			number, numberErr = primitiveStyleSize(number, metrics.Size)
			if numberErr != nil {
				t.Fatal(numberErr)
			}
			number.Leading = metrics.Leading
			add(number, "strongnum-"+role+"/"+level)
		}
		st, err := r.bodyStyle("number-long")
		if err != nil {
			t.Fatal(err)
		}
		add(st, "number-long/"+level)
		st, err = r.bodyStyle("small")
		if err != nil {
			t.Fatal(err)
		}
		st, err = primitiveStyleSize(st, st.Size*12.5/12)
		if err != nil {
			t.Fatal(err)
		}
		add(st, "gantt-lane-title/"+level)
		st, err = r.bodyStyle("label")
		if err != nil {
			t.Fatal(err)
		}
		st, err = primitiveStyleSize(st, st.Size*11/9)
		if err != nil {
			t.Fatal(err)
		}
		add(st, "index-number/"+level)
		st, err = r.intakeVennDensityBadgeStyle(Style{Family: "IBM Plex Mono"})
		if err != nil {
			t.Fatal(err)
		}
		add(st, "venn-point-badge/"+level)
		st, err = r.bodyStyle("label")
		if err != nil {
			t.Fatal(err)
		}
		st.Leading = st.Size * 13 / 9
		st.Family = "IBM Plex Mono"
		add(st, "table-reference-badge/"+level)
	}
	keys := make([]string, 0, len(pairs))
	for key := range pairs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	controls := make([]densityControl, 0, len(keys))
	page, y := 1, 36.
	for i, key := range keys {
		control := *pairs[key]
		control.ID = fmt.Sprintf("density-extra-control-%03d", i+1)
		control.Style.ID = control.ID
		h := control.Style.Leading + math.Max(control.Style.Leading, 1.5*control.Style.Size) + 12
		if y+h > 504 {
			page++
			y = 36
		}
		control.Page, control.Rect = page, Rect{57, y, 846, h}
		y += h + 12
		controls = append(controls, control)
	}
	if len(controls) == 0 {
		t.Fatal("no new unique role-ratio pairs")
	}
	deck, report := densityBareNativeControls(t, source, typography, controls)
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "font-controls-extra-bare.pptx"), deck, 0644); err != nil {
		t.Fatal(err)
	}
	for file, value := range map[string]any{"font-controls-extra-bare.layout-report.json": report, "font-controls-extra.manifest.json": struct {
		Schema           string           `json:"schema"`
		SourceCommit     string           `json:"source_commit"`
		BaseSourceCommit string           `json:"base_source_commit"`
		BaseBareSHA      string           `json:"base_bare_pptx_sha256"`
		BareSHA          string           `json:"bare_pptx_sha256"`
		Controls         []densityControl `json:"controls"`
	}{"pptxgengo.wmds-density-extra-native-controls.v1", source.Commit, base.SourceCommit, base.BareSHA, report.PPTXSHA256, controls}} {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("wrote %d new unique pairs on %d bare pages to %s", len(controls), page, out)
}

// Reproducible roomy native controls. They add styles to an in-memory copy only;
// frozen bundle files and base calibration stay untouched.
// WMDS_DENSITY_CONTROL_OUT=/private/tmp/pptxgengo-density-controls go test
// ./internal/wmdesign -run '^TestWriteDensityNativeControls$' -count=1
func TestWriteDensityNativeControls(t *testing.T) {
	out := os.Getenv("WMDS_DENSITY_CONTROL_OUT")
	if out == "" {
		t.Skip("set WMDS_DENSITY_CONTROL_OUT for native fixtures")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	source := densityTestSource(t)
	typography, err := NewTypographyEngine(filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	pairs := map[string]*densityControl{}
	add := func(st Style, role string) {
		st.Case = ""
		font, err := typography.Resolve(st)
		if err != nil {
			t.Fatal(err)
		}
		key := anchorKey(font.SHA256, st.Size, st.Leading)
		if _, known := typography.anchors[key]; known {
			return
		}
		if old := pairs[key]; old != nil {
			old.Roles = append(old.Roles, role)
			return
		}
		pairs[key] = &densityControl{Roles: []string{role}, Style: st, Font: font, Text: "Agyp Q012\nAgyp Q012"}
	}
	scopes := []string{"body", "header", "cell"}
	for _, scope := range scopes {
		roles := source.Tokens.Density.Body
		if scope == "header" {
			roles = source.Tokens.Density.Header
		}
		names := make([]string, 0, len(roles))
		for role := range roles {
			names = append(names, role)
		}
		sort.Strings(names)
		for _, role := range names {
			token := role
			if strings.HasPrefix(role, "list-") || strings.HasPrefix(role, "ol-col") {
				continue
			}
			if strings.HasPrefix(role, "cell-") {
				if scope != "cell" {
					continue
				}
				token = strings.TrimPrefix(role, "cell-")
			}
			if scope == "cell" && role != "cell-body" && role != "cell-small" {
				continue
			}
			for _, level := range source.Tokens.Density.Levels {
				base, err := source.StyleForDensity(token, level, scope)
				if err != nil {
					t.Fatal(err)
				}
				families := []string{base.Family}
				if scope == "cell" {
					families = append(families, "IBM Plex Mono")
				}
				for _, family := range families {
					for _, weight := range []int{400, 500, 600, 700} {
						st := base
						st.Weight = weight
						st.Family = family
						tag := fmt.Sprintf("%s/%s/%s/%s/%d", scope, role, level, family, weight)
						add(st, tag)
						// Frozen branch/cell reference stacks inherit CSS ratio leading.
						if token == "body" || token == "small" || token == "subhead" {
							for _, ratio := range []float64{1.15, 1.2, 1.25} {
								variant := st
								variant.Leading = st.Size * ratio
								add(variant, tag+fmt.Sprintf("/ratio%.2f", ratio))
							}
						}
					}
				}
			}
		}
	}
	keys := make([]string, 0, len(pairs))
	for key := range pairs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	copySource := *source
	copySource.styles = map[string]Style{}
	for id, st := range source.styles {
		copySource.styles[id] = st
	}
	disabled := false
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Title: "WMDS density native font controls"}
	var manifest []densityControl
	y := 126.
	slide := SlideSpec{}
	newSlide := func() {
		slide = SlideSpec{ID: fmt.Sprintf("density-calibration-%03d", len(doc.Slides)+1), AutoDensity: &disabled, Frame: FrameRequest{Footer: "compact", TitleLines: 1}, Eyebrow: "Density font calibration", Title: "Roomy native font controls"}
		y = 126
	}
	newSlide()
	for i, key := range keys {
		control := *pairs[key]
		control.ID = fmt.Sprintf("density-control-%03d", i+1)
		h := control.Style.Leading + math.Max(control.Style.Leading, 1.5*control.Style.Size) + 12
		if y+h > 468 {
			doc.Slides = append(doc.Slides, slide)
			newSlide()
		}
		control.Rect = Rect{57, y, 846, h}
		control.Page = len(doc.Slides) + 1
		control.Style.ID = control.ID
		copySource.styles[control.ID] = control.Style
		slide.Nodes = append(slide.Nodes, Node{ID: control.ID, Kind: "text", Style: control.ID, Text: control.Text, Rect: control.Rect})
		manifest = append(manifest, control)
		y += h + 12
	}
	if len(slide.Nodes) > 0 {
		doc.Slides = append(doc.Slides, slide)
	}
	deck, report, err := buildWithLoadedSource(densityTestBundle(), &copySource, doc, CandidateEngine, nil)
	if err != nil {
		t.Fatal(err)
	}
	writeDensityFixture(t, out, "font-controls", doc, report, deck)
	bareDeck, bareReport := densityBareNativeControls(t, source, typography, manifest)
	if err = os.WriteFile(filepath.Join(out, "font-controls-bare.pptx"), bareDeck, 0644); err != nil {
		t.Fatal(err)
	}
	bareJSON, err := json.MarshalIndent(bareReport, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "font-controls-bare.layout-report.json"), bareJSON, 0644); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		Schema             string           `json:"schema"`
		SourceCommit       string           `json:"source_commit"`
		OriginalPPTXSHA256 string           `json:"original_pptx_sha256"`
		BarePPTXSHA256     string           `json:"bare_pptx_sha256"`
		Controls           []densityControl `json:"controls"`
	}{"pptxgengo.wmds-density-native-controls.v1", source.Commit, report.PPTXSHA256, bareReport.PPTXSHA256, manifest}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "font-controls.manifest.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	// Body demos retain an ordinary Comfortable header at every body level.
	demo := densityNativeDemoDocument(t, false)
	deck, report, err = BuildWithEngine(densityTestBundle(), "", demo, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	writeDensityFixture(t, out, "body-demo", demo, report, deck)
	// Explicit header demos retain the ordinary fixed header geometry.
	demo = densityNativeDemoDocument(t, true)
	deck, report, err = BuildWithEngine(densityTestBundle(), "", demo, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	writeDensityFixture(t, out, "header-demo", demo, report, deck)
	t.Logf("wrote %d unique new font pairs on %d pages to %s", len(manifest), len(doc.Slides), out)
}

// Bare controls intentionally have no slide master, chrome, pictures or title.
// Their layout report and SHA identify a separate calibration artifact; the
// compiled normal document above continues to describe only the normal deck.
func densityBareNativeControls(t *testing.T, source *Source, typography *Typography, controls []densityControl) ([]byte, Report) {
	t.Helper()
	p := pptx.New()
	p.DefineLayout("WMDS", 960./72, 540./72)
	if err := p.SetLayout("WMDS"); err != nil {
		t.Fatal(err)
	}
	p.Title = "WMDS bare density native font controls"
	p.Theme = pptx.ThemeProps{HeadFontFace: "IBM Plex Sans SemiBold", BodyFontFace: "IBM Plex Sans"}
	report := Report{Schema: "pptxgengo.wmds-density-bare-controls.v1", Engine: CandidateEngine, SourceRevision: source.Revision, SourceCommit: source.Commit, SourceFiles: source.Files, Fonts: typography.Fonts(), Qualification: "implemented_unqualified"}
	records := map[int][]TextRecord{}
	r := renderer{source: source, typeEngine: typography, pres: p}
	page := 0
	var sr SlideReport
	for _, control := range controls {
		if control.Page != page {
			if page > 0 {
				records[page] = sr.Texts
				report.Slides = append(report.Slides, sr)
			}
			page = control.Page
			sr = SlideReport{ID: fmt.Sprintf("density-bare-%03d", page), Page: page}
			r.slide = p.AddSlide(nil)
			r.slide.PresSlide().Name = sr.ID
			r.records = &sr.Texts
		}
		r.text(control.ID, control.Text, control.Style, control.Rect, "070154", "left", 0)
		if r.err != nil {
			t.Fatal(r.err)
		}
	}
	if page > 0 {
		records[page] = sr.Texts
		report.Slides = append(report.Slides, sr)
	}
	deck, err := p.Write()
	if err != nil {
		t.Fatal(err)
	}
	deck, err = candidateParagraphs(deck, records)
	if err != nil {
		t.Fatal(err)
	}
	deck, err = explicitTypography(deck)
	if err != nil {
		t.Fatal(err)
	}
	report.PPTXSHA256 = fmt.Sprintf("%x", sha256.Sum256(deck))
	return deck, report
}
func writeDensityFixture(t *testing.T, out, name string, doc Document, report Report, deck []byte) {
	t.Helper()
	for file, value := range map[string]any{name + ".compiled.json": doc, name + ".layout-report.json": report} {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(out, name+".pptx"), deck, 0644); err != nil {
		t.Fatal(err)
	}
}
func densityNativeDemoDocument(t *testing.T, headers bool) Document {
	t.Helper()
	disabled := false
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Title: "WMDS density role controls"}
	for _, level := range []string{"comfortable", "compact", "dense"} {
		slide := SlideSpec{ID: "density-" + level, AutoDensity: &disabled, Density: level, Frame: FrameRequest{Footer: "compact", TitleLines: 2}, Eyebrow: "Body density · " + level, Title: "One slide uses one body tier\nThe action title stays Comfortable"}
		nodes := []string{
			`{"type":"card","x":57,"y":162,"w":270,"h":150,"surface":"subtle","title":"Card title role","body":[{"p":"Card copy uses the slide body tier."}]}`,
			`{"type":"table","x":345,"y":162,"w":558,"dense":true,"rowHeader":true,"cols":[{"k":"plain","label":"Nested small text","w":270},{"k":"items","label":"Cell bullets","w":288,"type":"bullets"}],"rows":[{"h":96,"plain":{"text":"Main cell paragraph","sub":"Nested small paragraph"},"items":["Editable native bullet one","Editable native bullet two"]},{"h":96,"plain":"Agyp Q012: table readability","items":["Same cell density in every row","Whole slide density control"]}]}`,
			`{"type":"bullets","x":57,"y":330,"w":270,"items":[{"lead":"Bold lead","text":"shares one body tier"},"Plain body bullet"]}`,
			`{"type":"text","style":"small","x":345,"y":410,"w":558,"h":36,"text":"Running small text keeps its semantic role. Agyp Q012"}`,
		}
		if headers {
			slide.ID = "header-" + level
			slide.Frame.HeaderDensity = level
			slide.Eyebrow = "Header density · " + level
			slide.Title = "An explicit header density\nkeeps the title zones fixed"
		}
		for i, raw := range nodes {
			slide.Nodes = append(slide.Nodes, Node{ID: fmt.Sprintf("node%02d", i+1), Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(raw), Path: fmt.Sprintf("/body/%d", i)}})
		}
		doc.Slides = append(doc.Slides, slide)
	}
	return doc
}

func TestWriteDensityPublicNativeDemo(t *testing.T) {
	out := os.Getenv("WMDS_DENSITY_DEMO_OUT")
	if out == "" {
		t.Skip("set WMDS_DENSITY_DEMO_OUT for the public density/header fixture")
	}
	doc := densityNativeDemoDocument(t, false)
	headers := densityNativeDemoDocument(t, true)
	doc.Slides = append(doc.Slides, headers.Slides...)
	deck, report, err := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	writeDensityFixture(t, out, "density-demo", doc, report, deck)
}
