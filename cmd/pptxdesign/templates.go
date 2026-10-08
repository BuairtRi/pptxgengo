package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runTemplates(command, bundle, source, engine, out, spec string, year int, editing ...string) error {
	if engine != wmdesign.CandidateEngine {
		return fmt.Errorf("%s requires --engine %s", command, wmdesign.CandidateEngine)
	}
	if command == "templates" {
		if out != "" || spec != "" {
			return fmt.Errorf("templates does not accept --out or --spec")
		}
		catalog, e := wmdesign.TemplateCatalog(bundle, source)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(catalog)
	}
	if out == "" {
		return fmt.Errorf("--out NEW-DIR is required")
	}
	if _, e := os.Stat(out); !os.IsNotExist(e) {
		return fmt.Errorf("output directory must not already exist: %s", out)
	}
	var input wmdesign.BoundDocument
	name := "bound-templates.pptx"
	if command == "template-reference" {
		if spec != "" {
			return fmt.Errorf("template-reference does not accept --spec")
		}
		input = wmdesign.TemplateReference(year)
		name = "template-reference.pptx"
	} else {
		if spec == "" {
			return fmt.Errorf("template requires --spec bound-content.json")
		}
		raw, e := os.ReadFile(spec)
		if e != nil {
			return e
		}
		input, e = wmdesign.DecodeBoundDocument(raw)
		if e != nil {
			return e
		}
	}
	compiled, bindings, e := wmdesign.BindTemplates(bundle, source, input)
	if e != nil {
		return e
	}
	compiled.EditingProfile = templateEditingProfile(editing)
	deck, layout, e := wmdesign.BuildWithEngine(bundle, source, compiled, engine)
	if e != nil {
		return e
	}
	emitDensityWarnings(layout)
	if e = os.MkdirAll(out, 0755); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(out, name), deck, 0644); e != nil {
		return e
	}
	for _, artifact := range []struct {
		name  string
		value any
	}{
		{"template-content.json", input}, {"compiled-document.json", compiled},
		{"binding-report.json", bindings}, {"layout-report.json", layout},
	} {
		if e = wmdesign.WriteJSON(filepath.Join(out, artifact.name), artifact.value); e != nil {
			return e
		}
	}
	fmt.Printf("Generated %d bound slides: %s\nProfile: %s; native qualification pending.\n", len(compiled.Slides), filepath.Join(out, name), wmdesign.ProfileForEngine(engine))
	return nil
}

func templateEditingProfile(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return wmdesign.NativeEditingProfile
}
