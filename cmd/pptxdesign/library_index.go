package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runLibraryIndex(command string, args []string) error {
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	bundle := f.String("bundle", "", "pinned bundle path or v11; new indices default to v11")
	source := f.String("source", "", "matching source override")
	legacy := f.String("legacy-index", "", "optional legacy catalog-library.sqlite projection")
	legacyRoot := f.String("legacy-root", "", "original/relocated release root for legacy contract resource paths")
	slideLibrary := f.String("slide-library", "", "closed finished-slide revisions root; read overrides may relocate it")
	gallery := f.String("gallery", "", "optional matching release catalog root for hash-pinned preview links")
	indexPath := f.String("index", "", "unified SQLite index for read-only discovery")
	out := f.String("out", "", "new SQLite file for library-index, or new directory for library-fit")
	id := f.String("id", "", "exact entity ID or unique modern canonical key")
	query := f.String("query", "", "soft scenario query")
	retrieval := f.String("retrieval", "metadata", "ranking method: metadata or keyword (SQLite FTS5 BM25)")
	requireShape := f.Bool("require-shape", false, "require all supplied structural hints and an exact source item-count group; does not establish measured fit")
	kinds := f.String("kinds", "", "explicit comma-separated entity-kind filter")
	assetKind := f.String("asset-kind", "all", "asset summary filter: all, icon, photo, graphic, logo")
	namespace := f.String("namespace", "", "explicit wmds, legacy or curated namespace filter")
	lifecycles := f.String("lifecycles", "", "explicit comma-separated lifecycle filter; deprecated still requires --include-deprecated")
	contentAdapter := f.String("content-adapter", "", "exact declared content-adapter capability filter; does not establish measured fit")
	roles := f.String("roles", "", "comma-separated soft content-role hints")
	structures := f.String("structures", "", "comma-separated soft structural hints")
	forms := f.String("visual-forms", "", "comma-separated optional visual-form hints")
	items := f.Int("items", 0, "desired semantic item count; soft hint")
	itemRole := f.String("item-role", "", "associate item count with a content role")
	limit := f.Int("limit", 10, "maximum search results, 1..100")
	deprecated := f.Bool("include-deprecated", false, "include deprecated results")
	includeWeak := f.Bool("include-weak", false, "include zero-score results when a text query is present")
	summary := f.Bool("summary", false, "compact discovery output with verified screenshot paths; inspect includes zone-to-binding map")
	engine := f.String("engine", wmdesign.CandidateEngine, "build engine for fit; discovery records this passthrough without evaluating compatibility")
	spec := f.String("spec", "", "supplied-content candidate BoundDocument JSON for library-fit")
	if e := f.Parse(args); e != nil {
		return e
	}
	allowed := map[string]map[string]bool{
		"library-index":   {"bundle": true, "source": true, "legacy-index": true, "legacy-root": true, "gallery": true, "slide-library": true, "out": true},
		"library-find":    {"bundle": true, "source": true, "legacy-index": true, "legacy-root": true, "gallery": true, "slide-library": true, "index": true, "query": true, "retrieval": true, "require-shape": true, "lifecycles": true, "content-adapter": true, "kinds": true, "asset-kind": true, "namespace": true, "roles": true, "structures": true, "visual-forms": true, "items": true, "item-role": true, "limit": true, "include-deprecated": true, "include-weak": true, "engine": true, "summary": true},
		"library-inspect": {"bundle": true, "source": true, "legacy-index": true, "legacy-root": true, "gallery": true, "slide-library": true, "index": true, "id": true, "summary": true},
		"library-preview": {"bundle": true, "source": true, "legacy-index": true, "legacy-root": true, "gallery": true, "slide-library": true, "index": true, "id": true},
		"library-fit":     {"bundle": true, "source": true, "engine": true, "spec": true, "out": true},
	}
	var invalid string
	f.Visit(func(value *flag.Flag) {
		if !allowed[command][value.Name] && invalid == "" {
			invalid = value.Name
		}
	})
	if invalid != "" {
		return fmt.Errorf("%s does not accept --%s", command, invalid)
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if command == "library-find" && *retrieval != "metadata" && *retrieval != "keyword" {
		return fmt.Errorf("--retrieval must be metadata or keyword; model-backed semantic/hybrid retrieval is not available in this build")
	}
	if command == "library-find" && *summary && *kinds == "asset" {
		if *retrieval != "metadata" || *requireShape || *lifecycles != "" || *contentAdapter != "" {
			return fmt.Errorf("asset registry summaries do not support keyword ranking, require-shape, lifecycle or content-adapter filters; omit --summary for indexed asset discovery")
		}
		if *limit < 1 || *limit > 100 {
			return fmt.Errorf("--limit must be 1..100")
		}
		if *namespace != "" && *namespace != "wmds" {
			return fmt.Errorf("asset summaries require namespace wmds")
		}
		if *roles != "" || *structures != "" || *forms != "" || *items != 0 || *itemRole != "" {
			return fmt.Errorf("asset summaries use --query and --asset-kind; template structure filters are unsupported")
		}
		result, err := wmdesign.AssetSelections(*query, *assetKind, *limit)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"schema": "pptxgengo.asset-selection.v1", "matches": wmdesign.CompactAssetSelections(result), "policy": []string{"Uses the installed asset registry; photo metadata and dimensions come from its pinned sidecar snapshot. Photo search does not rehash originals; preview, rendering and packaging verify the selected original. Template SQLite, bundle and gallery filters do not select an asset snapshot.", "Icon color variants are grouped; preview MIME identifies original SVG or raster format."}})
	}
	var assetFilter bool
	f.Visit(func(value *flag.Flag) {
		if value.Name == "asset-kind" {
			assetFilter = true
		}
	})
	if assetFilter {
		return fmt.Errorf("--asset-kind requires library-find --kinds asset --summary")
	}
	if validLockedBundle(*bundle) {
		*bundle = designBundlePath(*bundle)
	}
	options := wmdesign.LibraryIndexOptions{Bundle: *bundle, Source: *source, LegacyIndex: *legacy, LegacyRoot: *legacyRoot, Gallery: *gallery, SlideLibrary: *slideLibrary}
	if command == "library-index" {
		if options.Bundle == "" {
			options.Bundle = designBundlePath(currentDesignBundle)
		}
		if *out == "" {
			return fmt.Errorf("library-index requires --out NEW-SQLITE-FILE")
		}
		report, e := wmdesign.BuildLibraryIndex(*out, options)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(report)
	}
	if command == "library-fit" {
		if options.Bundle == "" {
			options.Bundle = designBundlePath(currentDesignBundle)
		}
		if *spec == "" || *out == "" {
			return fmt.Errorf("library-fit requires --spec FILE --out NEW-DIR")
		}
		raw, e := os.ReadFile(*spec)
		if e != nil {
			return e
		}
		input, e := wmdesign.DecodeBoundDocument(raw)
		if e != nil {
			return e
		}
		report, e := wmdesign.FitLibraryCandidates(options.Bundle, options.Source, *engine, *out, input)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(report)
	}
	if *indexPath == "" {
		if options.Bundle == "" {
			options.Bundle = designBundlePath(currentDesignBundle)
		}
		*indexPath = filepath.Join(options.Bundle, "library.sqlite")
		if options.Gallery == "" {
			options.Gallery = filepath.Join(options.Bundle, "catalog")
		}
	}
	index, e := wmdesign.OpenLibraryIndex(*indexPath, options)
	if e != nil {
		return e
	}
	defer index.Close()
	split := func(value string) []string {
		var result []string
		for _, part := range strings.Split(value, ",") {
			if part = strings.TrimSpace(part); part != "" {
				result = append(result, part)
			}
		}
		return result
	}
	switch command {
	case "library-find":
		if *limit < 1 || *limit > 100 {
			return fmt.Errorf("--limit must be 1..100")
		}
		options := wmdesign.LibraryIndexFindOptions{Shape: wmdesign.LibrarySearchOptions{EngineHint: *engine, Query: *query, ContentRoles: split(*roles), Structures: split(*structures), VisualForms: split(*forms), Items: *items, ItemRole: *itemRole, Limit: *limit, IncludeDeprecated: *deprecated}, Kinds: split(*kinds), Namespace: *namespace, IncludeWeak: *includeWeak, Retrieval: *retrieval, RequireShape: *requireShape, Lifecycles: split(*lifecycles), ContentAdapter: *contentAdapter}
		if *summary {
			result, e := index.FindSummary(options)
			if e != nil {
				return e
			}
			return json.NewEncoder(os.Stdout).Encode(result)
		}
		result, e := index.Find(options)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	case "library-inspect":
		if *id == "" {
			return fmt.Errorf("--id required")
		}
		if *summary {
			result, e := index.SelectionCard(*id)
			if e != nil {
				return e
			}
			return json.NewEncoder(os.Stdout).Encode(result)
		}
		result, e := index.Inspect(*id)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	case "library-preview":
		if *id == "" {
			return fmt.Errorf("--id required")
		}
		result, e := index.Preview(*id)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	default:
		return fmt.Errorf("unknown unified library command %q", command)
	}
}
