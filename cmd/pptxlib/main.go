package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/library"
)

func printJSON(v any) { b, _ := json.MarshalIndent(v, "", "  "); fmt.Println(string(b)) }
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pptxlib index|find|inspect|preview|instantiate|assemble [flags]")
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	root := fs.String("root", ".", "repository root")
	index := fs.String("index", "", "library SQLite index")
	id := fs.String("id", "", "contract ID")
	query := fs.String("query", "", "search text")
	kind := fs.String("kind", "", "item kind")
	state := fs.String("state", "", "qualification state")
	pref := fs.String("preference", "", "preference")
	source := fs.String("source", "", "source ID")
	category := fs.String("category", "", "template category (inventory only)")
	altitude := fs.String("altitude", "", "framing, overview, explanation, detail or reference (inventory only)")
	density := fs.String("density", "", "sparse, medium or dense (inventory only)")
	limit := fs.Int("limit", 20, "maximum results (1..100)")
	inventory := fs.Bool("inventory", false, "search exploratory inventory")
	includeAvoid := fs.Bool("include-avoid", false, "include avoided references")
	variant := fs.String("variant", "", "preview/style variant")
	values := fs.String("values", "", "semantic values JSON")
	config := fs.String("config", "", "assembly config JSON")
	out := fs.String("out", "", "new output index or instantiation directory")
	allow := fs.Bool("allow-unqualified", false, "explicit experimental instantiation")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("positional arguments unsupported")
	}
	s, err := library.NewStore(*root)
	if err != nil {
		return err
	}
	idx := *index
	if idx == "" {
		idx = s.IndexPath
	}
	switch args[0] {
	case "index":
		path := *out
		if path == "" {
			path = s.IndexPath
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		r, err := s.BuildIndex(path)
		if err != nil {
			return err
		}
		printJSON(r)
	case "find":
		r, err := s.Find(idx, library.FindOptions{Query: *query, Kind: *kind, State: *state, Preference: *pref, Source: *source, Category: *category, Altitude: *altitude, Density: *density, Limit: *limit, Inventory: *inventory, IncludeAvoid: *includeAvoid})
		if err != nil {
			return err
		}
		printJSON(r)
	case "inspect":
		if *id == "" {
			return fmt.Errorf("--id required")
		}
		if *inventory {
			r, err := s.InspectInventory(idx, *id)
			if err != nil {
				return err
			}
			printJSON(r)
			return nil
		}
		r, err := s.Inspect(idx, *id)
		if err != nil {
			return err
		}
		printJSON(r)
	case "preview":
		if *id == "" {
			return fmt.Errorf("--id required")
		}
		if *inventory {
			if *variant != "" {
				return fmt.Errorf("inventory preview has no qualified style variants")
			}
			r, err := s.PreviewInventory(idx, *id)
			if err != nil {
				return err
			}
			printJSON(r)
			return nil
		}
		r, err := s.Preview(idx, *id, *variant)
		if err != nil {
			return err
		}
		printJSON(r)
	case "instantiate":
		if *id == "" || *values == "" || *out == "" {
			return fmt.Errorf("--id, --values and --out required")
		}
		r, err := s.Instantiate(idx, *id, *values, *out, *allow)
		if err != nil {
			return err
		}
		printJSON(r)
	case "assemble":
		if *config == "" || *out == "" {
			return fmt.Errorf("--config and --out required")
		}
		r, err := s.Assemble(idx, *config, *out, *allow)
		if err != nil {
			return err
		}
		printJSON(r)
	default:
		return fmt.Errorf("unknown command %s", args[0])
	}
	return nil
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "pptxlib:", err)
		os.Exit(1)
	}
}
