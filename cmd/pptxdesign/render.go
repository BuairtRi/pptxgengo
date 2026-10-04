package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
)

func runRender(args []string) error {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	source := flags.String("pptx", "", "existing PowerPoint deck (never modified)")
	out := flags.String("out", "", "new output directory")
	pdf := flags.Bool("pdf", false, "export PDF through local Microsoft PowerPoint")
	png := flags.Bool("png", false, "render each native PDF page to PNG with macOS PDFKit")
	hidden := flags.Bool("include-hidden", false, "make hidden slides visible in the temporary review copy")
	slides := flags.String("slides", "", "one-based source slide numbers or ranges, e.g. 3,5-7")
	contact := flags.Bool("contact-sheet", false, "create a PNG contact sheet labeled with source slide numbers")
	staging := flags.String("staging-dir", "", "PowerPoint-accessible staging folder (or PPTXGENGO_NATIVE_STAGING; default user cache)")
	timeout := flags.Duration("timeout", 5*time.Minute, "total native export and rasterization timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("render accepts no positional arguments")
	}
	receipt, err := nativeexport.Render(context.Background(), nativeexport.Options{PPTX: *source, Out: *out, PDF: *pdf, PNG: *png, IncludeHidden: *hidden, Timeout: *timeout, Slides: *slides, ContactSheet: *contact, StagingRoot: *staging})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(receipt)
}
