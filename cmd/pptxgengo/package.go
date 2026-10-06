package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/releasepackage"
	"os"
	"path/filepath"
)

func runPackage(root string, args []string) error {
	f := flag.NewFlagSet("package", flag.ContinueOnError)
	release := f.String("release", root, "installed release to package (never modified)")
	bin := f.String("binaries", "", "directory with three cross-compiled Windows .exe files")
	skill := f.String("skill", "", "latest skill directory; defaults to release snapshot")
	out := f.String("out", "", "new portable ZIP path")
	arch := f.String("arch", "amd64", "Windows architecture: amd64 or arm64")
	version := f.String("version", "", "Windows preview version matching the supplied executables")
	withoutPhotos := f.Bool("without-photos", false, "omit original photo bytes; keep searchable metadata and gallery thumbnails")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("package accepts no positional arguments")
	}
	if *skill == "" {
		*skill = filepath.Join(*release, "skills", "west-monroe-presentations")
	}
	report, err := releasepackage.Create(releasepackage.Options{Release: *release, Binaries: *bin, Skill: *skill, Out: *out, Architecture: *arch, Version: *version, WithoutPhotos: *withoutPhotos})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(report)
}
