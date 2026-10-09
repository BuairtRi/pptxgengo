package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectAssessmentReconcile(args []string) error {
	f := flag.NewFlagSet("project assessment reconcile", flag.ContinueOnError)
	project := f.String("project", ".", "project directory or deck.yaml")
	node := f.String("node", "", "stable assessment node ID")
	slide := f.String("slide", "", "stable slide ID")
	packetPath := f.String("packet", "", "closed receipt-backed geometry review packet")
	decisions := f.String("decisions", "", "explicit ordinal score decisions; preview unless --apply")
	apply := f.Bool("apply", false, "commit measured source score changes and retained evidence")
	bundle := f.String("bundle", "", "defaults to lock")
	engine := f.String("engine", "", "defaults to lock")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" || *node == "" || *packetPath == "" || *apply && *decisions == "" {
		return fmt.Errorf("usage: project assessment reconcile --project PATH --slide ID --node ID --packet DIRECTORY [--decisions FILE --apply]")
	}
	p, e := deckproject.Load(*project)
	if e != nil {
		return e
	}
	b, en, e := projectRuntime(p, *bundle, *engine)
	if e != nil {
		return e
	}
	packet, e := deckproject.ReadTextReviewPacket(*packetPath)
	if e != nil {
		return e
	}
	var out any
	if *decisions == "" {
		report, e := deckproject.ProposeAssessmentSemantics(p, packet, *slide, *node, b, en)
		if e != nil {
			return e
		}
		out = struct {
			Report       deckproject.AssessmentSemanticReport `json:"report"`
			ReportSHA256 string                               `json:"report_sha256"`
		}{report, deckproject.AssessmentSemanticReportHash(report)}
	} else {
		raw, e := readReconciliationInput(*decisions, 1<<20)
		if e != nil {
			return e
		}
		out, e = deckproject.AdoptAssessmentSemantics(p, packet, *slide, *node, raw, b, en, *apply)
		if e != nil {
			return e
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}
