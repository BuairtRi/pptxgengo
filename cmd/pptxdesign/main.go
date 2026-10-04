// pptxdesign is the opt-in WMDS foundation command, independent of the installed
// release and the legacy compose/native font calibration profiles.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"io"
	"os"
	"path/filepath"
	"time"
)

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: pptxdesign <project|render|render-doctor|source-inventory|asset-gallery|asset-catalog|library-find|library-inspect|library-preview|library-match|library-authoring|library-fit|library-index|library-catalog|library-search|templates|template|build|inspect|reference|template-reference|library-reference|library-source-reference|library-sweep|library-bound-sweep|frame-reference|component-reference|metric-reference|card-row-reference|data-metric-reference|rich-reference|parallel-reference|typography-probes> [flags]")
	}
	command := os.Args[1]
	if command == "render-native-worker" {
		return runRenderWorker(os.Args[2:])
	}
	if command == "asset-gallery" {
		return runAssetGallery(os.Args[2:])
	}
	if command == "source-inventory" {
		return runSourceInventory(os.Args[2:])
	}
	if command == "library-authoring" {
		return runLibraryAuthoring(os.Args[2:])
	}
	if command == "library-match" {
		return runLibraryMatch(os.Args[2:])
	}
	if command == "render-doctor" {
		return runRenderDoctor(os.Args[2:])
	}
	if command == "render" {
		return runRender(os.Args[2:])
	}
	if command == "project" {
		return runProject(os.Args[2:])
	}
	if command == "library-index" || command == "library-find" || command == "library-inspect" || command == "library-preview" || command == "library-fit" {
		return runLibraryIndex(command, os.Args[2:])
	}
	if command == "library-search" {
		return runLibrarySearch(os.Args[2:])
	}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	bundle := f.String("bundle", "v5", "pinned foundation bundle path or v5 (default)")
	engine := f.String("engine", wmdesign.CandidateEngine, "typography engine: wmds-go-foundation.v1 or wmds-go-foundation.v2 (candidate)")
	source := f.String("source", "", "optional WMDS source override; must match pinned snapshot")
	out := f.String("out", "", "new output directory")
	spec := f.String("spec", "", "foundation document JSON")
	family := f.String("family", "", "source library family for library references")
	templateKeys := f.String("template-keys", "", "comma-separated library keys for a focused reference")
	includeDeprecated := f.Bool("include-deprecated", false, "include deprecated entries in library-catalog")
	year := f.Int("year", time.Now().Year(), "explicit legal year for reference generation")
	if e := f.Parse(os.Args[2:]); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *bundle == "v1" || *bundle == "v2" || *bundle == "v3" || *bundle == "v4" || *bundle == "v5" {
		root := designReleaseRoot()
		*bundle = filepath.Join(root, "library", "wm-design-system", *bundle)
	}
	if command == "asset-catalog" {
		if *out != "" || *spec != "" || *family != "" || *source != "" || *templateKeys != "" || *includeDeprecated {
			return fmt.Errorf("asset-catalog does not accept output, content or source filters")
		}
		return json.NewEncoder(os.Stdout).Encode(wmdesign.PrimitiveAssetCatalog())
	}
	if command == "library-catalog" || command == "library-reference" || command == "library-source-reference" || command == "library-sweep" || command == "library-bound-sweep" {
		if *spec != "" {
			return fmt.Errorf("%s does not accept --spec", command)
		}
		return runLibrary(command, *bundle, *source, *engine, *out, *family, *templateKeys, *includeDeprecated, *year)
	}
	if *templateKeys != "" || *includeDeprecated {
		return fmt.Errorf("--template-keys/--include-deprecated require a library command")
	}
	if *family != "" {
		return fmt.Errorf("--family is supported only by library reference and sweep commands")
	}
	if command == "templates" || command == "template" || command == "template-reference" {
		return runTemplates(command, *bundle, *source, *engine, *out, *spec, *year)
	}
	if command == "inspect" {
		s, e := wmdesign.Load(*bundle, *source)
		if e != nil {
			return e
		}
		t, e := wmdesign.NewTypographyEngine(filepath.Join(*bundle, "fonts"), *engine)
		if e != nil {
			return e
		}
		for _, st := range s.Tokens.Type {
			if _, e = t.Resolve(st); e != nil {
				return e
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"profile": wmdesign.ProfileForEngine(*engine), "engine": *engine, "status": "implemented_unqualified", "styles": s.Tokens.Type, "grid": s.Tokens.Grid, "frames": s.Frames, "sources": s.Files, "source_revision": s.Revision, "source_commit": s.Commit, "fonts": t.Fonts(), "executable_node_kinds": []string{"text", "box", "textblock (v2)", "card (v2)", "cardrow (v2)", "metric (v2)", "richtext (v2)", "rule (v2)", "scene (v2): primitives, media, cards, tables, charts, diagrams, sequences, people"}, "components_and_templates": "Source-pinned closed library bindings in candidate v2; library-catalog reports revision and pending capabilities. Native review and reusable content qualification are separate from implementation."})
	}
	if command != "reference" && command != "frame-reference" && command != "component-reference" && command != "metric-reference" && command != "card-row-reference" && command != "data-metric-reference" && command != "rich-reference" && command != "parallel-reference" && command != "build" && command != "typography-probes" {
		return fmt.Errorf("unknown command %q", command)
	}
	if *out == "" {
		return fmt.Errorf("--out NEW-DIR is required")
	}
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		return fmt.Errorf("output directory must not already exist: %s", *out)
	}
	if command == "typography-probes" {
		if *spec != "" {
			return fmt.Errorf("typography-probes does not accept --spec")
		}
		deck, manifest, e := wmdesign.TypographyProbesWithEngine(*bundle, *source, *engine)
		if e != nil {
			return e
		}
		if e = os.MkdirAll(*out, 0755); e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(*out, "typography-probes.pptx"), deck, 0644); e != nil {
			return e
		}
		if e = wmdesign.WriteJSON(filepath.Join(*out, "probes.json"), manifest); e != nil {
			return e
		}
		fmt.Printf("Generated %d controls and %d rejection records: %s\nNative capture pending.\n", len(manifest.Probes), len(manifest.Rejections), *out)
		return nil
	}
	doc := wmdesign.Reference(*year)
	specName := "foundation.json"
	name := "reference.pptx"
	if command == "frame-reference" {
		if *spec != "" || *engine != wmdesign.CandidateEngine {
			return fmt.Errorf("frame-reference requires candidate v2 and does not accept --spec")
		}
		var err error
		doc, err = wmdesign.SplitFrameReference(*bundle, *source, *year)
		if err != nil {
			return err
		}
		name, specName = "frame-reference.pptx", "frames.json"
	}
	if command == "component-reference" {
		if *engine != wmdesign.CandidateEngine {
			return fmt.Errorf("component-reference requires --engine %s", wmdesign.CandidateEngine)
		}
		doc = wmdesign.ComponentReference(*year)
		name = "component-reference.pptx"
		specName = "components.json"
	}
	if command == "metric-reference" {
		if *engine != wmdesign.CandidateEngine {
			return fmt.Errorf("metric-reference requires --engine %s", wmdesign.CandidateEngine)
		}
		doc = wmdesign.MetricReference(*year)
		name = "metric-reference.pptx"
		specName = "metrics.json"
	}
	if command == "card-row-reference" || command == "data-metric-reference" || command == "rich-reference" || command == "parallel-reference" {
		if *engine != wmdesign.CandidateEngine {
			return fmt.Errorf("%s requires --engine %s", command, wmdesign.CandidateEngine)
		}
		switch command {
		case "card-row-reference":
			doc = wmdesign.CardRowReference(*year)
			name = "card-row-reference.pptx"
			specName = "card-rows.json"
		case "data-metric-reference":
			doc = wmdesign.DataMetricReference(*year)
			name = "data-metric-reference.pptx"
			specName = "data-metrics.json"
		case "rich-reference":
			doc = wmdesign.RichReference(*year)
			name = "rich-reference.pptx"
			specName = "rich-text.json"
		case "parallel-reference":
			doc = wmdesign.ParallelReference(*year)
			name = "parallel-reference.pptx"
			specName = "slices.json"
		}
	}
	if command == "build" {
		if *spec == "" {
			return fmt.Errorf("build requires --spec")
		}
		b, e := os.ReadFile(*spec)
		if e != nil {
			return e
		}
		doc = wmdesign.Document{}
		d := json.NewDecoder(bytes.NewReader(b))
		d.DisallowUnknownFields()
		if e = d.Decode(&doc); e != nil {
			return fmt.Errorf("source.unsupported_field_or_invalid_spec: %w", e)
		}
		var extra any
		if e = d.Decode(&extra); e != nil && e != io.EOF {
			return e
		} else if e == nil {
			return fmt.Errorf("multiple JSON documents")
		}
	} else if *spec != "" {
		return fmt.Errorf("reference does not accept --spec")
	}
	deck, report, e := wmdesign.BuildWithEngine(*bundle, *source, doc, *engine)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(*out, 0755); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, name), deck, 0644); e != nil {
		return e
	}
	if e = wmdesign.WriteJSON(filepath.Join(*out, "layout-report.json"), report); e != nil {
		return e
	}
	if e = wmdesign.WriteJSON(filepath.Join(*out, specName), doc); e != nil {
		return e
	}
	fmt.Printf("Generated %d slides: %s\nProfile: %s; native qualification pending.\n", len(doc.Slides), filepath.Join(*out, name), wmdesign.ProfileForEngine(*engine))
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "pptxdesign:", e)
		os.Exit(1)
	}
}
