// pptxscene is the native-object reconstruction spike, not the proposed
// high-level presentation authoring CLI.
package main

import (
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/nativepkg"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fail(fmt.Errorf("usage: pptxscene extract|build [flags]"))
	}
	f := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	source := f.String("source", "", "source PPTX for extract")
	project := f.String("project", "", "scene project for build")
	out := f.String("out", "", "new output path")
	list := f.String("slides", "", "comma-separated source slide numbers")
	freeze := f.Bool("freeze-slide-numbers", false, "convert source slide-number fields to editable text")
	_ = f.Parse(os.Args[2:])
	if *out == "" {
		fail(fmt.Errorf("--out required"))
	}
	numbers, e := nativepkg.Numbers(*list)
	if e != nil {
		fail(e)
	}
	switch os.Args[1] {
	case "extract":
		if *source == "" || len(numbers) == 0 {
			fail(fmt.Errorf("--source and --slides required"))
		}
		e = nativepkg.Extract(*source, *out, numbers)
	case "build":
		if *project == "" {
			fail(fmt.Errorf("--project required"))
		}
		e = nativepkg.Build(*project, *out, numbers, *freeze)
	default:
		e = fmt.Errorf("unknown command %s", os.Args[1])
	}
	if e != nil {
		fail(e)
	}
	fmt.Println(*out)
}
func fail(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
