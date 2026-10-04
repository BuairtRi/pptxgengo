package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

type inventoryMappingReport struct {
	Schema                string                           `json:"schema"`
	SourceInventorySHA256 string                           `json:"source_inventory_sha256"`
	ProjectID             string                           `json:"project_id"`
	ProjectSourceSHA256   string                           `json:"project_source_sha256"`
	BundleRevision        string                           `json:"bundle_revision"`
	Engine                string                           `json:"engine"`
	MappingStatus         string                           `json:"mapping_status"`
	MappingPolicy         string                           `json:"mapping_policy"`
	SourceSlides          []inventorySourceSlideMapping    `json:"source_slides"`
	ProjectSlides         []inventoryProjectSlideReference `json:"project_slides"`
}

type inventorySourceSlideMapping struct {
	SourceSlideIndex int    `json:"source_slide_index"`
	SourceTitle      string `json:"source_title,omitempty"`
	ProjectSlideID   string `json:"project_slide_id,omitempty"`
	Status           string `json:"status"`
}

type inventoryProjectSlideReference struct {
	Page                   int      `json:"page"`
	ID                     string   `json:"id"`
	Title                  string   `json:"compiled_title"`
	Hidden                 bool     `json:"hidden"`
	ContentKind            string   `json:"content_kind"`
	TemplateScope          string   `json:"template_scope"`
	TemplateID             string   `json:"template_id"`
	TemplateRevision       string   `json:"template_revision,omitempty"`
	TemplateSourceFile     string   `json:"template_source_file,omitempty"`
	TemplateSourceSHA256   string   `json:"template_source_sha256,omitempty"`
	SlideSourceFile        string   `json:"slide_source_file,omitempty"`
	NotesSourceFile        string   `json:"notes_source_file,omitempty"`
	MappedSourceSlideIndex int      `json:"mapped_source_slide_index,omitempty"`
	MappingStatus          string   `json:"mapping_status"`
	SourcePointers         []string `json:"template_source_pointers,omitempty"`
}

// runSourceInventory inventories an existing PPTX without changing it. When a
// project is supplied, its compiled source references are reported separately;
// no source pages are inferred to match project pages.
func runSourceInventory(args []string) error {
	flags := flag.NewFlagSet("source-inventory", flag.ContinueOnError)
	in := flags.String("in", "", "source PPTX to inventory")
	out := flags.String("out", "", "new output directory for JSON and Markdown reports")
	project := flags.String("project", "", "optional maintained deck project for explicit source-reference report")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *in == "" || *out == "" {
		return fmt.Errorf("source-inventory requires --in PPTX --out NEW-DIR")
	}
	report, err := nativepkg.Inventory(*in)
	if err != nil {
		return fmt.Errorf("source inventory: %w", err)
	}
	var mapping *inventoryMappingReport
	if *project != "" {
		mapping, err = buildInventoryMapping(*project, report)
		if err != nil {
			return err
		}
	}
	markdown := sourceInventoryMarkdown(report, mapping)
	output, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	if _, err = os.Lstat(output); err == nil {
		return fmt.Errorf("output directory must not already exist: %s", output)
	} else if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(output)
	if err = os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, "."+filepath.Base(output)+".tmp-")
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(stage)
		}
	}()
	if err = writeInventoryJSON(filepath.Join(stage, "source_inventory.json"), report); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(stage, "source_inventory.md"), []byte(markdown), 0644); err != nil {
		return err
	}
	if mapping != nil {
		if err = writeInventoryJSON(filepath.Join(stage, "mapping_report.json"), mapping); err != nil {
			return err
		}
	}
	if _, err = os.Lstat(output); err == nil {
		return fmt.Errorf("output directory appeared during inventory: %s", output)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.Rename(stage, output); err != nil {
		return err
	}
	committed = true
	fmt.Printf("Wrote source inventory for %d slides: %s\n", report.SlideCount, output)
	if mapping != nil {
		fmt.Println("Project mapping report marks all source pages unmapped; no source matching was inferred.")
	}
	return nil
}

func writeInventoryJSON(path string, value any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(value)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func buildInventoryMapping(projectPath string, source nativepkg.SourceInventory) (*inventoryMappingReport, error) {
	p, err := deckproject.Load(projectPath)
	if err != nil {
		return nil, err
	}
	lock, _, err := deckproject.ReadLock(p)
	if err != nil {
		return nil, fmt.Errorf("source-inventory --project requires a pinned project lock: %w", err)
	}
	bundlePath, engine, err := projectRuntime(p, "", "")
	if err != nil {
		return nil, err
	}
	compiled, err := deckproject.Check(p, bundlePath, engine)
	if err != nil {
		return nil, fmt.Errorf("compile project for source mapping: %w", err)
	}
	report := &inventoryMappingReport{Schema: "pptxgengo.source-project-mapping.v1", SourceInventorySHA256: source.SourceSHA256, ProjectID: p.Document.ID, ProjectSourceSHA256: p.SourceHash(), BundleRevision: lock.BundleRevision, Engine: engine, MappingStatus: "source_slides_unmapped", MappingPolicy: "Source pages and project slides are listed independently. No page-order, title, text, image, or layout heuristic was used to claim a match.", SourceSlides: []inventorySourceSlideMapping{}, ProjectSlides: []inventoryProjectSlideReference{}}
	for _, slide := range source.Slides {
		report.SourceSlides = append(report.SourceSlides, inventorySourceSlideMapping{SourceSlideIndex: slide.OriginalSlideIndex, SourceTitle: slide.Title, Status: "unmapped"})
	}
	for i, slide := range compiled.Document.Slides {
		if i >= len(p.Document.Slides) {
			return nil, fmt.Errorf("compiled project slide count does not match source project")
		}
		authored := p.Document.Slides[i]
		row := inventoryProjectSlideReference{Page: i + 1, ID: slide.ID, Title: slide.Title, Hidden: slide.Hidden, ContentKind: slide.ContentKind, TemplateScope: authored.Template.Scope, TemplateID: authored.Template.ID, TemplateRevision: authored.Template.Revision, SlideSourceFile: p.SlideFiles[slide.ID], NotesSourceFile: p.NotesFiles[slide.ID], MappingStatus: "unmapped"}
		if slide.TemplateBinding != nil {
			row.TemplateRevision = slide.TemplateBinding.SourceRevision
			row.TemplateSourceFile = slide.TemplateBinding.SourceFile
			row.TemplateSourceSHA256 = slide.TemplateBinding.SourceSHA256
			for _, assignment := range slide.TemplateBinding.Assignments {
				if assignment.SourcePointer != "" {
					row.SourcePointers = append(row.SourcePointers, assignment.SourcePointer)
				}
			}
		} else if authored.Template.Scope == "local" {
			row.TemplateSourceFile = p.TemplateFiles[authored.Template.ID]
			if raw := p.SourceFiles[row.TemplateSourceFile]; len(raw) > 0 {
				row.TemplateSourceSHA256 = fmt.Sprintf("%x", sha256.Sum256(raw))
			}
		}
		report.ProjectSlides = append(report.ProjectSlides, row)
	}
	return report, nil
}

func sourceInventoryMarkdown(report nativepkg.SourceInventory, mapping *inventoryMappingReport) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# Source PPTX inventory\n\n- File: `%s`\n- SHA-256: `%s`\n- Slides: %d\n\n", report.Source, report.SourceSHA256, report.SlideCount)
	for _, slide := range report.Slides {
		title := slide.Title
		if title == "" {
			title = "(untitled)"
		}
		fmt.Fprintf(&output, "## Slide %d: %s\n\n", slide.OriginalSlideIndex, markdownInline(title))
		visibility := "visible"
		if slide.Hidden {
			visibility = "hidden"
		}
		fmt.Fprintf(&output, "- Package part: `%s`\n- Visibility: %s\n- Shapes: %d; text shapes: %d; pictures: %d; tables: %d; charts: %d; other visuals: %d\n\n", slide.Part, visibility, slide.ShapeCount, slide.TextShapeCount, slide.PictureCount, slide.TableCount, slide.ChartCount, slide.OtherVisualCount)
		for _, text := range slide.Texts {
			label := text.ObjectName
			if text.Placeholder != "" {
				label += " (" + text.Placeholder + ")"
			}
			if label == "" {
				label = "text"
			}
			fmt.Fprintf(&output, "- **%s**\n", markdownInline(label))
			for _, line := range strings.Split(text.Text, "\n") {
				fmt.Fprintf(&output, "  - %s\n", markdownInline(line))
			}
		}
		if len(slide.Texts) > 0 {
			output.WriteByte('\n')
		}
		for i, table := range slide.Tables {
			fmt.Fprintf(&output, "### Table %d: %s\n\n", i+1, markdownInline(table.ObjectName))
			output.WriteString("```text\n")
			for _, row := range table.Rows {
				for j := range row {
					row[j] = strings.ReplaceAll(strings.ReplaceAll(row[j], "\n", " "), "\t", " ")
				}
				fmt.Fprintf(&output, "%s\n", strings.Join(row, " | "))
			}
			output.WriteString("```\n\n")
		}
		for i, chart := range slide.Charts {
			fmt.Fprintf(&output, "- **Chart %d: %s** — type: %s; series: %d; part: `%s`\n", i+1, markdownInline(chart.Title), strings.Join(chart.Types, ", "), chart.Series, chart.Part)
		}
		if len(slide.Charts) > 0 {
			output.WriteByte('\n')
		}
		for i, picture := range slide.Images {
			fmt.Fprintf(&output, "- **Picture %d: %s** — alt: %s; package part: `%s`; MIME: %s; bytes: %d\n", i+1, markdownInline(picture.ObjectName), markdownInline(picture.Description), picture.Target, picture.MIME, picture.Bytes)
		}
		if len(slide.Images) > 0 {
			output.WriteByte('\n')
		}
		for _, note := range slide.SpeakerNotes {
			fmt.Fprintf(&output, "- **Speaker note:** %s\n", markdownInline(note))
		}
		if len(slide.SpeakerNotes) > 0 {
			output.WriteByte('\n')
		}
	}
	if mapping != nil {
		output.WriteString("# Project references (unmapped)\n\n")
		fmt.Fprintf(&output, "Project `%s`; source hash `%s`; bundle `%s`; engine `%s`. Source-to-project page matching remains **unmapped**.\n\n", mapping.ProjectID, mapping.ProjectSourceSHA256, mapping.BundleRevision, mapping.Engine)
		output.WriteString(mapping.MappingPolicy + "\n\n")
		output.WriteString("| Project page | Slide ID | Compiled title | Template scope | Template ID | Template source | Mapping |\n|---:|---|---|---|---|---|---|\n")
		for _, slide := range mapping.ProjectSlides {
			fmt.Fprintf(&output, "| %d | `%s` | %s | %s | `%s` | `%s` | %s |\n", slide.Page, markdownTable(slide.ID), markdownTable(slide.Title), markdownTable(slide.TemplateScope), markdownTable(slide.TemplateID), markdownTable(slide.TemplateSourceFile), slide.MappingStatus)
		}
		output.WriteByte('\n')
	}
	return output.String()
}

func markdownInline(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(value, "`", "'"), "\r", " "), "\n", " ")
}

func markdownTable(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(markdownInline(value), "|", "\\|"), "\r", " ")
}
