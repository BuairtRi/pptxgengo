package main

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
)

func runLibrary(command, bundle, source, engine, out, family, templateKeys, templateKeysFile string, includeDeprecated bool, year int) error {
	if engine != wmdesign.CandidateEngine {
		return fmt.Errorf("%s requires --engine %s", command, wmdesign.CandidateEngine)
	}
	catalog, err := wmdesign.LibraryCatalog(bundle, source)
	if err != nil {
		return err
	}
	if includeDeprecated && command != "library-catalog" {
		return fmt.Errorf("--include-deprecated is supported only by library-catalog")
	}
	selected, seen, err := libraryTemplateSelection(catalog, family, templateKeys, templateKeysFile)
	if err != nil {
		return err
	}
	if command == "library-catalog" {
		if family != "" {
			return fmt.Errorf("library-catalog does not accept --family")
		}
		if out != "" {
			return fmt.Errorf("library-catalog does not accept --out")
		}
		visible := []wmdesign.LibraryTemplate{}
		for _, t := range catalog {
			if (includeDeprecated || t.Status != "deprecated") && (len(selected) == 0 || seen[t.Key]) {
				visible = append(visible, t)
			}
		}
		return json.NewEncoder(os.Stdout).Encode(visible)
	}
	if out == "" {
		return fmt.Errorf("--out NEW-DIR required")
	}
	if _, err = os.Stat(out); !os.IsNotExist(err) {
		return fmt.Errorf("output directory must not already exist: %s", out)
	}
	var doc wmdesign.Document
	var content any
	var bindings any
	if command == "library-reference" || command == "library-bound-sweep" {
		input, e := wmdesign.LibraryReference(bundle, source, family, year)
		if e != nil {
			return e
		}
		if len(selected) > 0 {
			all := input.Slides
			input.Slides = nil
			for _, key := range selected {
				found := false
				for _, slide := range all {
					if slide.Template == key {
						input.Slides = append(input.Slides, slide)
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("template unavailable for reference: %s", key)
				}
			}
		}
		compiled, report, e := wmdesign.BindTemplates(bundle, source, input)
		if e != nil {
			return e
		}
		doc = compiled
		content = input
		bindings = report
	} else {
		doc, err = wmdesign.LibrarySourceReference(bundle, source, family, year)
		if err != nil {
			return err
		}
		if len(selected) > 0 {
			all := doc.Slides
			doc.Slides = nil
			for _, key := range selected {
				for _, slide := range all {
					if slide.TemplateBinding.Template == key {
						doc.Slides = append(doc.Slides, slide)
						break
					}
				}
			}
		}
	}
	if command == "library-sweep" || command == "library-bound-sweep" {
		return librarySweep(bundle, source, engine, out, doc)
	}
	deck, report, err := wmdesign.BuildWithEngine(bundle, source, doc, engine)
	if err != nil {
		return err
	}
	emitDensityWarnings(report)
	if err = os.MkdirAll(out, 0755); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(out, "library-reference.pptx"), deck, 0644); err != nil {
		return err
	}
	assets, err := wmdesign.UsedPrimitiveAssets(deck)
	if err != nil {
		return err
	}
	for _, a := range []struct {
		name  string
		value any
	}{{"compiled-document.json", doc}, {"layout-report.json", report}, {"catalog.json", catalog}, {"template-content.json", content}, {"binding-report.json", bindings}, {"media-assets.json", assets}} {
		if a.value != nil {
			if err = wmdesign.WriteJSON(filepath.Join(out, a.name), a.value); err != nil {
				return err
			}
		}
	}
	fmt.Printf("Generated %d library slides: %s\nNative visual review pending.\n", len(doc.Slides), out)
	return nil
}

// This developer artifact generator records each source fixture independently
// so one fit error never hides the status of the remaining library.
func librarySweep(bundle, source, engine, out string, doc wmdesign.Document) error {
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	type receipt struct {
		Template string `json:"template"`
		Status   string `json:"status"`
		Error    string `json:"error,omitempty"`
		Artifact string `json:"artifact,omitempty"`
	}
	var receipts []receipt
	passed := doc
	passed.Slides = nil
	for _, slide := range doc.Slides {
		one := doc
		one.Slides = []wmdesign.SlideSpec{slide}
		r := receipt{Template: slide.TemplateBinding.Template, Status: "generation_failed"}
		deck, report, err := wmdesign.BuildWithEngine(bundle, source, one, engine)
		if err != nil {
			r.Error = err.Error()
			fmt.Printf("FAIL %s: %s\n", r.Template, r.Error)
		} else {
			emitDensityWarnings(report)
			dir := filepath.Join(out, slide.ID)
			if err = os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			r.Artifact = filepath.Join(slide.ID, "source-reference.pptx")
			r.Status = "generated_native_review_pending"
			if err = os.WriteFile(filepath.Join(out, r.Artifact), deck, 0644); err != nil {
				return err
			}
			if err = wmdesign.WriteJSON(filepath.Join(dir, "layout-report.json"), report); err != nil {
				return err
			}
			passed.Slides = append(passed.Slides, slide)
		}
		receipts = append(receipts, r)
	}
	if err := wmdesign.WriteJSON(filepath.Join(out, "generation-receipts.json"), receipts); err != nil {
		return err
	}
	if err := wmdesign.WriteJSON(filepath.Join(out, "compiled-document.json"), doc); err != nil {
		return err
	}
	if len(passed.Slides) > 0 {
		deck, report, err := wmdesign.BuildWithEngine(bundle, source, passed, engine)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(out, "generated-library.pptx"), deck, 0644); err != nil {
			return err
		}
		assets, err := wmdesign.UsedPrimitiveAssets(deck)
		if err != nil {
			return err
		}
		if err = wmdesign.WriteJSON(filepath.Join(out, "media-assets.json"), assets); err != nil {
			return err
		}
		if err = wmdesign.WriteJSON(filepath.Join(out, "layout-report.json"), report); err != nil {
			return err
		}
	}
	fmt.Printf("Generated %d/%d source fixtures. Receipts: %s\n", len(passed.Slides), len(doc.Slides), out)
	return nil
}
