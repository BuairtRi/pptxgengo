package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/localembed"
)

func runLibraryModel(args []string) error {
	f := flag.NewFlagSet("library-model", flag.ContinueOnError)
	out := f.String("out", "", "new offline model package directory")
	from := f.String("from", "", "copy an already pinned offline model directory")
	download := f.Bool("download", false, "explicitly download the pinned public model over TLS")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	report, err := localembed.PreparePackage(context.Background(), *out, *from, *download)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(report)
}
