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

func runLibrarySearch(args []string) error {
	f := flag.NewFlagSet("library-search", flag.ContinueOnError)
	bundle := f.String("bundle", "library/wm-design-system/v2", "pinned design library bundle path or v1/v2/v3/v4/v5")
	source := f.String("source", "", "optional source override; must match pinned snapshot")
	engine := f.String("engine", wmdesign.CandidateEngine, "wrapper compatibility passthrough; search does not evaluate engine/bundle compatibility")
	query := f.String("query", "", "scenario/purpose/label text; soft ranking signal")
	roles := f.String("roles", "", "comma-separated content-role hints, e.g. point,key-message")
	structures := f.String("structures", "", "comma-separated structural hints, e.g. repeated-items,sequence")
	visualForms := f.String("visual-forms", "", "comma-separated optional visual-form hints, e.g. image,icon")
	items := f.Int("items", 0, "desired repeated item count; ranking hint, not a fit guarantee")
	itemRole := f.String("item-role", "", "associate item count with a content role, e.g. point; requires --items")
	limit := f.Int("limit", 10, "maximum candidates, 1..100")
	includeDeprecated := f.Bool("include-deprecated", false, "include deprecated candidates and their replacement metadata")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *limit < 1 || *limit > 100 {
		return fmt.Errorf("library-search requires named flags and --limit 1..100")
	}
	if *itemRole != "" && *items == 0 {
		return fmt.Errorf("--item-role requires --items greater than zero")
	}
	if *engine != wmdesign.Engine && *engine != wmdesign.CandidateEngine {
		return fmt.Errorf("unsupported engine %q", *engine)
	}
	if *bundle == "v1" || *bundle == "v2" || *bundle == "v3" || *bundle == "v4" || *bundle == "v5" {
		*bundle = filepath.Join(os.Getenv("PPTXGENGO_RELEASE_ROOT"), "library", "wm-design-system", *bundle)
	}
	catalog, err := wmdesign.LibraryCatalog(*bundle, *source)
	if err != nil {
		return err
	}
	split := func(value string) []string {
		var out []string
		for _, part := range strings.Split(value, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
		return out
	}
	result, err := wmdesign.SearchLibrary(catalog, wmdesign.LibrarySearchOptions{EngineHint: *engine, Query: *query, ContentRoles: split(*roles), Structures: split(*structures), VisualForms: split(*visualForms), Items: *items, ItemRole: *itemRole, Limit: *limit, IncludeDeprecated: *includeDeprecated})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
