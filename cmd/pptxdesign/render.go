package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
)

func runRender(args []string) (err error) {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	source := flags.String("pptx", "", "existing PowerPoint deck (never modified)")
	out := flags.String("out", "", "new output directory")
	pdf := flags.Bool("pdf", false, "export PDF through local Microsoft PowerPoint")
	png := flags.Bool("png", false, "export native PNGs: macOS PDFKit or Windows PowerPoint COM")
	hidden := flags.Bool("include-hidden", false, "make hidden slides visible in the temporary review copy")
	slides := flags.String("slides", "", "one-based source slide numbers or ranges, e.g. 3,5-7")
	contact := flags.Bool("contact-sheet", false, "create a PNG contact sheet labeled with source slide numbers")
	staging := flags.String("staging-dir", "", "stable PowerPoint PPTX/PDF folder (same as render-doctor; or PPTXGENGO_NATIVE_STAGING; default user cache)")
	timeout := flags.Duration("timeout", 5*time.Minute, "total native export and rasterization timeout")
	renderStarted := false
	defer func() {
		failureOut := *out
		if failureOut == "" {
			failureOut = renderErrorOutput(args)
		}
		if err != nil && !errors.Is(err, flag.ErrHelp) && !renderStarted && failureOut != "" {
			if recordErr := nativeexport.RecordRenderError(context.Background(), failureOut, err); recordErr != nil {
				err = fmt.Errorf("%w; render-error.txt could not be recorded: %v", err, recordErr)
			}
		}
	}()
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("render accepts no positional arguments")
	}
	renderStarted = true
	if runtime.GOOS == "windows" {
		fmt.Fprintln(os.Stderr, "warning: Windows native PowerPoint rendering is experimental; review exported pages and report any COM or policy failures.")
	}
	receipt, err := nativeexport.Render(context.Background(), nativeexport.Options{PPTX: *source, Out: *out, PDF: *pdf, PNG: *png, IncludeHidden: *hidden, Timeout: *timeout, Slides: *slides, ContactSheet: *contact, StagingRoot: *staging})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(receipt)
}

// flag.Parse stops at the first invalid flag. Locate an explicitly supplied
// output even when it occurs later, so that preflight errors can be recorded.
func renderErrorOutput(args []string) string {
	out := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		// A string/duration flag's value is not another option, even if its
		// spelling begins with --out. Match the flag parser's consumption.
		switch arg {
		case "--pptx", "-pptx", "--slides", "-slides", "--staging-dir", "-staging-dir", "--timeout", "-timeout":
			if i+1 < len(args) {
				i++
			}
			continue
		}
		if arg == "--out" || arg == "-out" {
			if i+1 < len(args) {
				i++
				out = args[i]
			}
		} else if strings.HasPrefix(arg, "--out=") {
			out = strings.TrimPrefix(arg, "--out=")
		} else if strings.HasPrefix(arg, "-out=") {
			out = strings.TrimPrefix(arg, "-out=")
		}
	}
	return out
}
