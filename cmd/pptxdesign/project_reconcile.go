package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectReconcile(args []string) error {
	if len(args) == 0 || (args[0] != "propose" && args[0] != "adopt") {
		return fmt.Errorf("usage: project reconcile <propose|adopt> --project PATH; propose requires --edited FILE --out NEW_DIRECTORY; adopt requires --packet DIRECTORY --decisions FILE")
	}
	action := args[0]
	f := flag.NewFlagSet("project reconcile "+action, flag.ContinueOnError)
	project := f.String("project", ".", "maintained deck directory or deck.yaml")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if action == "propose" {
		structureMap := f.String("structure-map", "", "explicit YAML/JSON mappings for untagged copies of local block components; requires --geometry")
		geometry := f.Bool("geometry", false, "also propose tagged transforms, paint order and reviewed whole local component deletions")
		bundle := f.String("bundle", "", "bundle path/revision, defaults to project lock")
		engine := f.String("engine", "", "engine, defaults to project lock")
		build := f.String("build", "", "baseline build ID; defaults to the state-pinned current build")
		receipt := f.String("receipt-sha256", "", "trusted receipt SHA-256, required for a historical baseline")
		edited := f.String("edited", "", "edited PPTX copy; immutable build originals must be preserved")
		out := f.String("out", "", "new closed review packet directory; existing destinations are refused")
		if err := f.Parse(args[1:]); err != nil {
			return err
		}
		if f.NArg() != 0 || *edited == "" || *out == "" {
			return fmt.Errorf("reconcile propose requires --edited FILE and --out NEW_DIRECTORY, with no positional arguments")
		}
		if *structureMap != "" && !*geometry {
			return fmt.Errorf("--structure-map requires --geometry")
		}
		p, err := deckproject.Load(*project)
		if err != nil {
			return err
		}
		b, err := deckproject.ReadTextBaseline(p, *build, *receipt)
		if err != nil {
			return err
		}
		raw, err := readReconciliationInput(*edited, 512<<20)
		if err != nil {
			return err
		}
		var packet *deckproject.TextReviewPacket
		if *geometry {
			runtimeBundle, runtimeEngine, e := projectRuntime(p, *bundle, *engine)
			if e != nil {
				return e
			}
			if *structureMap != "" {
				mappingRaw, e := readReconciliationInput(*structureMap, 1<<20)
				if e != nil {
					return e
				}
				packet, err = deckproject.WriteMappedGeometryReviewPacket(p, b, raw, *out, runtimeBundle, runtimeEngine, mappingRaw)
			} else {
				packet, err = deckproject.WriteGeometryReviewPacket(p, b, raw, *out, runtimeBundle, runtimeEngine)
			}
		} else {
			packet, err = deckproject.WriteTextReviewPacket(p, b, raw, *out)
		}
		if err != nil {
			return err
		}
		return enc.Encode(struct {
			Packet       string                               `json:"packet"`
			ReportSHA256 string                               `json:"report_sha256"`
			Report       deckproject.TextReconciliationReport `json:"report"`
		}{packet.Root, packet.ReportSHA256, packet.Report})
	}
	packetPath := f.String("packet", "", "closed packet produced by reconcile propose")
	decisions := f.String("decisions", "", "explicit YAML/JSON review decisions naming actor, report hash and field IDs")
	bundle := f.String("bundle", "", "bundle path/revision, defaults to project lock")
	engine := f.String("engine", "", "engine, defaults to project lock")
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 || *packetPath == "" || *decisions == "" {
		return fmt.Errorf("reconcile adopt requires --packet DIRECTORY and --decisions FILE, with no positional arguments")
	}
	p, err := deckproject.Load(*project)
	if err != nil {
		return err
	}
	packet, err := deckproject.ReadTextReviewPacket(*packetPath)
	if err != nil {
		return err
	}
	raw, err := readReconciliationInput(*decisions, 16<<20)
	if err != nil {
		return err
	}
	b, e, err := projectRuntime(p, *bundle, *engine)
	if err != nil {
		return err
	}
	result, err := deckproject.AdoptTextReviewPacket(p, packet, raw, b, e)
	if err != nil {
		return err
	}
	return enc.Encode(result)
}

// Check identity both before and after a bounded read. Inputs must be ordinary
// files, not symlinks, devices or streams. No archive extraction is needed.
func readReconciliationInput(path string, max int64) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > max {
		return nil, fmt.Errorf("reconciliation input must be a regular file of at most %d bytes", max)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(before, opened) || !opened.Mode().IsRegular() {
		return nil, fmt.Errorf("reconciliation input changed while opening")
	}
	raw, err := io.ReadAll(io.LimitReader(file, max+1))
	if err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > max || int64(len(raw)) != before.Size() || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return nil, fmt.Errorf("reconciliation input changed or exceeded its size limit")
	}
	return raw, nil
}
