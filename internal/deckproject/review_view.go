package deckproject

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type ProjectReviewOptions struct {
	Stage, Out, Bundle, Engine string
	Audience                   bool
}
type ReviewSlide struct {
	Page           int               `json:"page"`
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	Template       string            `json:"template"`
	Scope          string            `json:"scope"`
	Hidden         bool              `json:"hidden"`
	Copy           map[string]any    `json:"copy,omitempty"`
	VisibleCopy    []string          `json:"visible_copy,omitempty"`
	CopyReviewGaps []string          `json:"copy_review_gaps,omitempty"`
	Notes          string            `json:"notes,omitempty"`
	Brief          string            `json:"brief,omitempty"`
	EvidenceRefs   []string          `json:"evidence_refs,omitempty"`
	Composition    *CompositionEntry `json:"composition,omitempty"`
	Review         NativeCoverage    `json:"native_review"`
	Image          string            `json:"image,omitempty"`
	NativeEvidence string            `json:"native_evidence,omitempty"`
	NativePDF      string            `json:"native_pdf,omitempty"`
}
type ProjectReviewManifest struct {
	Schema       string            `json:"schema"`
	Stage        string            `json:"stage"`
	Audience     bool              `json:"audience"`
	DeckID       string            `json:"deck_id"`
	SourceSHA256 string            `json:"source_sha256"`
	Slides       []ReviewSlide     `json:"slides"`
	StockSlides  int               `json:"stock_slides"`
	Files        map[string]string `json:"files"`
	Policy       []string          `json:"policy"`
}

// ProjectReview creates a fresh-context packet without changing the deck. Source
// and notes are private inputs; native images only come from verified attachments.
func ProjectReview(p *Project, o ProjectReviewOptions) (ProjectReviewManifest, error) {
	m := ProjectReviewManifest{Schema: "pptxgengo.project-review.v1", Stage: o.Stage, DeckID: p.Document.ID, SourceSHA256: p.SourceHash(), Slides: []ReviewSlide{}, Files: map[string]string{}, Policy: []string{"Contains private authored copy, notes and briefs.", "Rendering, visual review and acceptance are distinct; missing or stale native coverage remains visible.", "Thumbnails help navigation; review every native page individually before acceptance."}}
	m.Audience = o.Audience
	if o.Audience {
		m.Policy = []string{"Independent audience packet: visible slides and audience context only.", "Internal rationale, notes, briefs, decisions and prior verdicts are excluded.", "Rendering alone does not imply visual acceptance."}
	}
	if o.Stage != "outline" && o.Stage != "content" && o.Stage != "deck" {
		return m, fmt.Errorf("review stage must be outline, content or deck")
	}
	if o.Out == "" {
		return m, fmt.Errorf("review requires a new output directory")
	}
	compiled, err := Compile(p, o.Bundle, o.Engine)
	if err != nil {
		return m, err
	}
	if err = ValidateEditorial(p); err != nil {
		return m, err
	}
	composition, err := ReadCompositionLog(p)
	if err != nil {
		return m, err
	}
	coverage := make([]NativeCoverage, len(p.Document.Slides))
	if !o.Audience || o.Stage == "deck" {
		coverage, err = NativeReviewCoverage(p)
		if err != nil {
			return m, err
		}
	}
	packetDeps, err := dependencies(p)
	if err != nil {
		return m, err
	}
	packetBuild := ""
	if o.Stage == "deck" {
		state, e := Status(p)
		if e != nil {
			return m, e
		}
		if e = protectBaseline(p, state); e != nil {
			return m, e
		}
		deps, e := dependencies(p)
		if e != nil {
			return m, e
		}
		if state.CurrentBuild == "" || state.SemanticSHA256 != digest(p.Canonical) || !reflect.DeepEqual(deps, state.Dependencies) {
			return m, fmt.Errorf("deck review requires a current build with unchanged inputs")
		}
		packetBuild = state.CurrentBuild
	}
	catalog, err := wmdesign.LibraryCatalog(o.Bundle, "")
	if err != nil {
		return m, err
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	for _, d := range catalog {
		defs[d.Key] = d
	}
	var visibleReport wmdesign.Report
	if o.Audience && o.Stage != "outline" {
		// Take visible text from the renderer's measured records, rather than
		// exposing arbitrary authored values that may contain internal material.
		_, visibleReport, err = wmdesign.BuildWithEngineAndAssets(o.Bundle, "", compiled.Document, o.Engine, compiled.Assets)
		if err != nil {
			return m, err
		}
	}
	files := map[string][]byte{}
	attachmentsCopied := map[string]bool{}
	for i, slide := range p.Document.Slides {
		if o.Audience && slide.Hidden {
			continue
		}
		if o.Audience && o.Stage == "deck" && !coverage[i].Rendered {
			return m, fmt.Errorf("audience deck review requires current native output for visible slide %s; render and attach it first", slide.ID)
		}
		r := ReviewSlide{Page: i + 1, ID: slide.ID, Title: compiled.Document.Slides[i].Title, Template: slide.Template.ID, Scope: slide.Template.Scope, Hidden: slide.Hidden, Review: coverage[i]}
		if entry, exists := composition[slide.ID]; exists && !o.Audience {
			r.Composition = &entry
		}
		if slide.Template.Scope == "shared" {
			m.StockSlides++
		}
		if o.Audience {
			r.Review = NativeCoverage{SlideID: slide.ID, Rendered: coverage[i].Rendered, Status: "not_rendered"}
			if r.Review.Rendered {
				r.Review.Status = "rendered_review_pending"
			}
			if o.Stage != "outline" {
				r.VisibleCopy = reviewVisibleCopy(visibleReport.Slides[i])
				if len(visibleReport.Slides[i].Charts) != 0 {
					r.CopyReviewGaps = []string{"Review chart labels and values on the rendered slide; this text view does not include native chart text."}
				}
			}
		} else if o.Stage != "outline" {
			r.Copy = slide.Values
			r.Notes = slide.Notes
			r.EvidenceRefs = slide.EvidenceRefs
			if d, exists := defs[slide.Template.ID]; exists && slide.Template.Scope == "shared" {
				authored, e := StockEditableSlide(slide, d)
				if e != nil {
					return m, e
				}
				if content, ok := authored["content"].(map[string]any); ok {
					r.Copy = content
				}
			}
			if slide.Brief != "" {
				path, e := SafePath(p.Root, slide.Brief)
				if e != nil {
					return m, e
				}
				raw, e := os.ReadFile(path)
				if e != nil {
					return m, e
				}
				r.Brief = string(raw)
			}
		}
		if o.Stage == "deck" && r.Review.Rendered {
			base := "reviews/native/" + coverage[i].Attachment
			path, e := SafePath(p.Root, base+"/attachment.json")
			if e != nil {
				return m, e
			}
			raw, e := os.ReadFile(path)
			if e != nil {
				return m, e
			}
			var attachment NativeAttachment
			if e = json.Unmarshal(raw, &attachment); e != nil {
				return m, e
			}
			manifest, e := validateNativeAttachment(filepath.Dir(path), attachment)
			if e != nil {
				return m, e
			}
			packetBase := "native/" + attachment.ID
			if o.Audience {
				// Copy just this visible page's image. A PDF is safe only if every
				// exported page is visible; a deck with hidden pages needs PNGs.
				if coverage[i].Image != "" {
					relative := strings.TrimPrefix(coverage[i].Image, base+"/")
					if relative == coverage[i].Image {
						return m, fmt.Errorf("native image path is outside its verified attachment")
					}
					r.Image = packetBase + "/" + relative
					if err = copyReviewArtifact(files, filepath.Dir(path), relative, r.Image, attachment.Files); err != nil {
						return m, err
					}
				}
				pdfSafe := len(manifest.PageMappings) == manifest.Pages
				for _, page := range manifest.PageMappings {
					if page.SourceSlide < 1 || page.SourceSlide > len(p.Document.Slides) || p.Document.Slides[page.SourceSlide-1].Hidden {
						pdfSafe = false
					}
				}
				if manifest.PDF != nil && pdfSafe {
					r.NativePDF = packetBase + "/" + manifest.PDF.Path
					if err = copyReviewArtifact(files, filepath.Dir(path), manifest.PDF.Path, r.NativePDF, attachment.Files); err != nil {
						return m, err
					}
				}
				if r.Image == "" && r.NativePDF == "" {
					return m, fmt.Errorf("audience deck review requires visible-page PNGs or a PDF containing only visible pages; rerun render --png")
				}
			} else {
				r.NativeEvidence = packetBase + "/attachment.json"
				if manifest.PDF != nil {
					r.NativePDF = packetBase + "/" + manifest.PDF.Path
				}
				if !attachmentsCopied[attachment.ID] {
					files[r.NativeEvidence] = raw
					for relative := range attachment.Files {
						source, e := SafePath(filepath.Dir(path), relative)
						if e != nil {
							return m, e
						}
						data, e := os.ReadFile(source)
						if e != nil {
							return m, e
						}
						if digest(data) != attachment.Files[relative] {
							return m, fmt.Errorf("native artifact changed while copying reviewer evidence: %s", relative)
						}
						files[packetBase+"/"+relative] = data
					}
					attachmentsCopied[attachment.ID] = true
				}
				if r.Review.Image != "" {
					relative := strings.TrimPrefix(r.Review.Image, base+"/")
					if relative == r.Review.Image {
						return m, fmt.Errorf("native image path is outside its verified attachment")
					}
					r.Image = packetBase + "/" + relative
					r.Review.Image = r.Image
				}
			}
		}
		m.Slides = append(m.Slides, r)
	}
	compositionSource, e := compositionPath(p)
	if e != nil {
		return m, e
	}
	if compositionSource != "" && !o.Audience {
		path, e := SafePath(p.Root, compositionSource)
		if e != nil {
			return m, e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return m, e
		}
		files["context/composition-log.yaml"] = data
	}
	for _, key := range []string{"project", "audience", "outline", "claims", "decisions"} {
		if o.Audience && key != "audience" && !(key == "outline" && o.Stage == "outline") {
			continue
		}
		if relative := p.Document.Context[key]; relative != "" {
			path, e := SafePath(p.Root, relative)
			if e != nil {
				return m, e
			}
			info, e := os.Stat(path)
			if e != nil {
				return m, e
			}
			if info.IsDir() {
				continue
			}
			if key == "claims" && o.Stage == "outline" {
				continue
			}
			data, e := os.ReadFile(path)
			if e != nil {
				return m, e
			}
			files["context/"+key+filepath.Ext(relative)] = data
		}
	}
	if o.Stage == "deck" && !o.Audience {
		state, e := Status(p)
		if e != nil {
			return m, e
		}
		data, e := os.ReadFile(filepath.Join(p.Root, "builds", state.CurrentBuild, "deck.pptx"))
		if e != nil {
			return m, e
		}
		files["deck.pptx"] = data
	}
	var titles strings.Builder
	for _, s := range m.Slides {
		fmt.Fprintf(&titles, "%02d. %s [%s]\n", s.Page, s.Title, s.ID)
	}
	files["titles.txt"] = []byte(titles.String())
	files["REVIEW.md"] = []byte(reviewInstructions(o.Stage))
	if o.Audience {
		files["REVIEW.md"] = []byte("# Independent audience review\n\nReview the audience context and visible slide content. Report findings by stable slide ID with evidence and a concrete repair. Evaluate the argument, clarity, visual fit and suitability for the audience. No author rationale or prior verdicts are included.\n")
	}
	if err = os.MkdirAll(filepath.Dir(o.Out), 0755); err != nil {
		return m, err
	}
	if _, err = os.Lstat(o.Out); err == nil {
		return m, fmt.Errorf("review output directory must be new")
	} else if !os.IsNotExist(err) {
		return m, err
	}
	stage, err := os.MkdirTemp(filepath.Dir(o.Out), ".project-review-")
	if err != nil {
		return m, err
	}
	defer os.RemoveAll(stage)
	for rel, data := range files {
		path, e := SafePath(stage, rel)
		if e != nil {
			return m, e
		}
		if e = writeExclusive(path, data, 0644); e != nil {
			return m, e
		}
		m.Files[rel] = digest(data)
	}
	file, err := os.Create(filepath.Join(stage, "index.html"))
	if err != nil {
		return m, err
	}
	tmpl, err := template.New("review").Parse(projectReviewHTML)
	if err == nil {
		err = tmpl.Execute(file, m)
	}
	closeErr := file.Close()
	if err != nil {
		return m, err
	}
	if closeErr != nil {
		return m, closeErr
	}
	html, err := os.ReadFile(filepath.Join(stage, "index.html"))
	if err != nil {
		return m, err
	}
	m.Files["index.html"] = digest(html)
	if err = writeExclusive(filepath.Join(stage, "manifest.json"), canonical(m), 0644); err != nil {
		return m, err
	}
	current, err := Load(p.SourcePath)
	if err != nil {
		return m, err
	}
	if current.SourceHash() != p.SourceHash() {
		return m, fmt.Errorf("source changed while making reviewer packet")
	}
	currentDeps, err := dependencies(current)
	if err != nil {
		return m, err
	}
	if !reflect.DeepEqual(packetDeps, currentDeps) {
		return m, fmt.Errorf("dependencies changed while making reviewer packet")
	}
	if o.Stage == "deck" {
		state, e := Status(current)
		if e != nil {
			return m, e
		}
		if state.CurrentBuild != packetBuild {
			return m, fmt.Errorf("build changed while making reviewer packet")
		}
		if e = protectBaseline(current, state); e != nil {
			return m, e
		}
	}
	if err = os.Rename(stage, o.Out); err != nil {
		return m, err
	}
	return m, nil
}

func copyReviewArtifact(files map[string][]byte, root, relative, destination string, hashes map[string]string) error {
	path, err := SafePath(root, relative)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if expected, ok := hashes[relative]; !ok || digest(data) != expected {
		return fmt.Errorf("native artifact changed while copying reviewer evidence: %s", relative)
	}
	files[destination] = data
	return nil
}

func reviewVisibleCopy(slide wmdesign.SlideReport) []string {
	copy := []string{}
	appendText := func(text wmdesign.TextRecord) {
		if strings.TrimSpace(text.Layout.Displayed) != "" {
			copy = append(copy, text.Layout.Displayed)
		}
	}
	for _, text := range slide.Texts {
		appendText(text)
	}
	// Native table cells are stored separately from ordinary shape text.
	for _, table := range slide.Tables {
		for _, cell := range table.Cells {
			appendText(cell.Text)
		}
	}
	return copy
}

func reviewInstructions(stage string) string {
	common := "# Independent presentation review\n\nReport findings with stable slide ID, severity, evidence and a concrete repair. Do not infer acceptance from a successful build or thumbnail. Preserve material facts, qualifiers, units and relationships.\n\n"
	switch stage {
	case "outline":
		return common + "Review titles.txt and context: does the argument support the audience's decision? Identify missing transitions, repetition, unsupported assertions and scope gaps.\n"
	case "content":
		return common + "Review manifest.json copy, notes, briefs and linked evidence. Check source accuracy, claim strength, qualifiers, relationships and copy length. Identify omissions and invented content.\n"
	default:
		return common + "Inspect every native PDF page or image individually and compare copy and relationships with manifest.json. Native PDF files, render receipts and attachment metadata are included under native/. Check wrapping, alignment, spacing, clipping, contrast, data labels and template suitability. Missing or stale native coverage means review is incomplete. Record visual decisions separately with project attach-render --decisions.\n"
	}
}

const projectReviewHTML = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>{{.DeckID}} review</title><style>body{font:16px system-ui;margin:32px;background:#f3f4f8;color:#16263f}header{max-width:1200px;margin:auto}input{padding:12px;width:min(90%,600px);border:1px solid #8893a5;border-radius:6px}main{max-width:1200px;margin:24px auto;display:grid;grid-template-columns:repeat(auto-fit,minmax(320px,1fr));gap:20px}article{background:white;padding:20px;border-radius:8px;box-shadow:0 2px 8px #16263f15}img{width:100%;height:auto}h2{font-size:19px;margin:12px 0}p{line-height:1.5}small{color:#536076}pre{white-space:pre-wrap;overflow-wrap:anywhere;font-size:12px}.status{font-weight:650}.accepted{color:#12613a}.issues_found,.stale{color:#a32942}a{color:#12396a}</style><header><h1>{{.DeckID}}</h1><p>{{.Stage}} review · {{len .Slides}} slides · {{.StockSlides}} shared stock layouts</p><p>{{if .Audience}}Independent audience packet.{{else}}Private operator packet. Rendering, review and acceptance are recorded separately.{{end}}</p><input id="filter" aria-label="Filter slides" placeholder="Find a title, slide ID or template"><p><a href="titles.txt">Titles in order</a> · <a href="manifest.json">Slide content</a> · <a href="REVIEW.md">Review instructions</a></p></header><main>{{range .Slides}}<article data-search="{{.ID}} {{.Title}} {{.Template}}"><small>{{.Page}} · {{.ID}}{{if .Hidden}} · hidden{{end}}</small>{{if .Image}}<a href="{{.Image}}"><img loading="lazy" src="{{.Image}}" alt="Native PowerPoint slide {{.Page}}"></a>{{else}}<p>Native image unavailable for this review stage.</p>{{end}}<h2>{{.Title}}</h2><p class="status {{.Review.Status}}">{{if .Review.Rendered}}Rendered · {{end}}{{.Review.Status}}</p><small>{{.Scope}} · {{.Template}}</small>{{if .Review.Note}}<p>{{.Review.Note}}</p>{{end}}{{if .VisibleCopy}}<section aria-label="Visible copy"><h3>Visible copy</h3>{{range .VisibleCopy}}<p>{{.}}</p>{{end}}</section>{{end}}{{range .CopyReviewGaps}}<p class="status">{{.}}</p>{{end}}{{if .Copy}}<details><summary>Authored copy</summary><pre>{{.Copy}}</pre></details>{{end}}{{if .NativePDF}}<p><a href="{{.NativePDF}}">Native PDF</a></p>{{end}}{{if .Notes}}<details><summary>Speaker notes</summary><pre>{{.Notes}}</pre></details>{{end}}{{if .Brief}}<details><summary>Slide brief</summary><pre>{{.Brief}}</pre></details>{{end}}{{if .NativeEvidence}}<p><a href="{{.NativeEvidence}}">Native render receipts</a>{{if .NativePDF}} · <a href="{{.NativePDF}}">Native PDF</a>{{end}}</p>{{end}}{{if .Composition}}<details><summary>Composition rationale</summary><p><strong>Purpose:</strong> {{.Composition.Purpose}}</p>{{if .Composition.Relationship}}<p><strong>Relationship:</strong> {{.Composition.Relationship}}</p>{{end}}<p><strong>Chosen template:</strong> {{.Composition.ChosenTemplate}}</p><p>{{.Composition.Rationale}}</p>{{if .Composition.Candidates}}<p>Considered templates:</p><ul>{{range .Composition.Candidates}}<li>{{.}}</li>{{end}}</ul>{{end}}{{if .Composition.Unresolved}}<p>Unresolved:</p><ul>{{range .Composition.Unresolved}}<li>{{.}}</li>{{end}}</ul>{{end}}</details>{{end}}</article>{{end}}</main><script>document.getElementById('filter').addEventListener('input',function(){const q=this.value.toLowerCase();for(const a of document.querySelectorAll('article'))a.hidden=!a.dataset.search.toLowerCase().includes(q)})</script></html>`

// MarshalVisualDecisions is useful to tools creating a separate review decision
// file; decisions do not mutate the source slide's content.
func MarshalVisualDecisions(decisions map[string]VisualDecision) ([]byte, error) {
	return json.MarshalIndent(decisions, "", "  ")
}
