package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/nativeexport"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runLibraryMatch(args []string) error {
	flags := flag.NewFlagSet("library-match", flag.ContinueOnError)
	pagePath := flags.String("page", "", "one semantic page YAML or JSON file")
	bundle := flags.String("bundle", "v9", "shared bundle path or v9")
	engine := flags.String("engine", wmdesign.CandidateEngine, "Go build engine")
	templates := flags.String("templates", "", "comma-separated exact candidate template keys")
	limit := flags.Int("limit", 4, "maximum passing automatic candidates, 1..100")
	out := flags.String("out", "", "new candidate output directory")
	year := flags.Int("year", 2026, "footer year")
	render := flags.Bool("render", false, "export combined candidate deck through native PowerPoint, with PNG contact sheet")
	timeout := flags.Duration("timeout", 5*time.Minute, "native render timeout")
	if e := flags.Parse(args); e != nil {
		return e
	}
	if flags.NArg() != 0 || *pagePath == "" || *out == "" {
		return fmt.Errorf("library-match requires --page FILE --out NEW-DIR")
	}
	if validLockedBundle(*bundle) {
		*bundle = designBundlePath(*bundle)
	}
	page, e := deckproject.LoadPageSpec(*pagePath)
	if e != nil {
		return e
	}
	keys := []string{}
	if *templates != "" {
		for _, key := range strings.Split(*templates, ",") {
			keys = append(keys, strings.TrimSpace(key))
		}
	}
	report, e := deckproject.MatchPage(page, deckproject.LibraryMatchOptions{Bundle: *bundle, Engine: *engine, Templates: keys, Limit: *limit, Out: *out, Year: *year})
	if e != nil {
		return e
	}
	if *render && report.CombinedDeck != "" {
		nativeOut := filepath.Join(*out, "native")
		_, renderErr := nativeexport.Render(context.Background(), nativeexport.Options{PPTX: report.CombinedDeck, Out: nativeOut, PDF: true, PNG: true, ContactSheet: true, Timeout: *timeout})
		for i := range report.Candidates {
			if report.Candidates[i].MappingComplete && report.Candidates[i].Deck != "" {
				if renderErr == nil {
					report.Candidates[i].NativeStatus = "rendered"
				} else {
					report.Candidates[i].NativeStatus = "render_failed"
				}
			}
		}
		if e = wmdesign.WriteJSON(filepath.Join(*out, "match-report.json"), report); e != nil {
			return e
		}
		if renderErr != nil {
			_ = json.NewEncoder(os.Stdout).Encode(report)
			return fmt.Errorf("candidates remain in %s; native rendering failed: %w", *out, renderErr)
		}
	}
	if e = json.NewEncoder(os.Stdout).Encode(report); e != nil {
		return e
	}
	if report.Passed == 0 {
		return fmt.Errorf("library.match_no_complete_candidate: inspect %s for explicit content/layout gaps", filepath.Join(*out, "match-report.json"))
	}
	return nil
}
