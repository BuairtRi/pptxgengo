package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"time"
)

type publicationReviews []string

func (p *publicationReviews) String() string         { return fmt.Sprint([]string(*p)) }
func (p *publicationReviews) Set(value string) error { *p = append(*p, value); return nil }

func runLibraryPublication(args []string) error {
	f := flag.NewFlagSet("library-publish", flag.ContinueOnError)
	var options wmdesign.LibraryPublicationOptions
	var reviews publicationReviews
	f.StringVar(&options.Bundle, "bundle", "", "explicit pinned publication bundle")
	f.StringVar(&options.Source, "source", "", "optional matching source override")
	f.StringVar(&options.PreviousBundle, "previous-bundle", "", "pinned snapshot for retained previews")
	f.StringVar(&options.PreviousGallery, "previous-gallery", "", "previous accepted source gallery")
	f.StringVar(&options.Version, "version", "", "release version")
	f.IntVar(&options.Year, "year", time.Now().Year(), "legal year of reviewed source specimens")
	f.Var(&reviews, "native-review", "accepted native review manifest; repeat for each intake")
	out := f.String("out", "", "new gallery directory")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || options.Bundle == "" || *out == "" {
		return fmt.Errorf("library-publish requires --bundle PATH --version VERSION --out NEWDIR")
	}
	options.NativeReviews = reviews
	report, err := wmdesign.PublishLibraryGallery(*out, options)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(report)
}
