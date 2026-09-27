// pptxadapt compiles semantic slide-family inputs into native measured composition specs.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/adapt"
)

func fail(e error) { fmt.Fprintln(os.Stderr, "pptxadapt:", e); os.Exit(1) }
func encode(v any) []byte {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		panic(e)
	}
	return append(b, '\n')
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pptxadapt compile --spec INPUT.json --out NEW_DIR | capabilities [--id TEMPLATE_ID] [--root DIR]")
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	specPath := fs.String("spec", "", "semantic adaptive-deck JSON")
	out := fs.String("out", "", "new compilation bundle directory")
	root := fs.String("root", ".", "repository root")
	id := fs.String("id", "", "template ID for capability lookup")
	if e := fs.Parse(args[1:]); e != nil {
		return e
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	switch args[0] {
	case "capabilities":
		b, e := os.ReadFile(filepath.Join(*root, "library/adaptive/capabilities.json"))
		if e != nil {
			return e
		}
		if *id == "" {
			fmt.Print(string(b))
			return nil
		}
		var data struct {
			Items []map[string]any `json:"items"`
		}
		if e := json.Unmarshal(b, &data); e != nil {
			return e
		}
		for _, row := range data.Items {
			if row["template_id"] == *id {
				fmt.Print(string(encode(row)))
				return nil
			}
		}
		return fmt.Errorf("unknown template %s", *id)
	case "compile":
		if *specPath == "" || *out == "" {
			return fmt.Errorf("compile requires --spec and --out")
		}
		b, e := os.ReadFile(*specPath)
		if e != nil {
			return e
		}
		var spec adapt.Spec
		if e = adapt.Decode(b, &spec); e != nil {
			return e
		}
		compiled, report, e := adapt.Compile(spec)
		if e != nil {
			return e
		}
		cb := encode(compiled)
		if _, e = os.Lstat(*out); !os.IsNotExist(e) {
			return fmt.Errorf("output must be a new directory")
		}
		parent := filepath.Dir(*out)
		if e = os.MkdirAll(parent, 0755); e != nil {
			return e
		}
		tmp, e := os.MkdirTemp(parent, ".adaptive-")
		if e != nil {
			return e
		}
		defer os.RemoveAll(tmp)
		records := map[string][]byte{"input.json": b, "spec.json": cb, "report.json": encode(report), "manifest.json": encode(map[string]any{"schema": "pptxgengo.adaptive-bundle.v1", "input_sha256": digest(b), "spec_sha256": digest(cb), "status": "compiled_pending_native_measurement", "render_route": "pptxcompose probe/measure/build/verify"})}
		for n, b := range records {
			if e = os.WriteFile(filepath.Join(tmp, n), b, 0644); e != nil {
				return e
			}
		}
		if e = os.Rename(tmp, *out); e != nil {
			return e
		}
		fmt.Println(filepath.Join(*out, "spec.json"))
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
func main() {
	if e := run(os.Args[1:]); e != nil {
		fail(e)
	}
}
