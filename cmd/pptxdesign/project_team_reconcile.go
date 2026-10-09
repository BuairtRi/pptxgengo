package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectTeamReconcile(args []string) error {
	f := flag.NewFlagSet("project team reconcile", flag.ContinueOnError)
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID")
	packetPath := f.String("packet", "", "closed receipt-backed geometry review packet")
	decisions := f.String("decisions", "", "explicit membership decisions; preview unless --apply")
	apply := f.Bool("apply", false, "commit measured source reassignment and retained evidence")
	bundle := f.String("bundle", "", "defaults to lock")
	engine := f.String("engine", "", "defaults to lock")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" || *packetPath == "" || *apply && *decisions == "" {
		return fmt.Errorf("usage: project team reconcile --project PATH --slide ID --packet DIRECTORY [--decisions FILE --apply]")
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
		report, e := deckproject.ProposeTeamSemantics(p, packet, *slide, b, en)
		if e != nil {
			return e
		}
		out = struct {
			Report       deckproject.TeamSemanticReport `json:"report"`
			ReportSHA256 string                         `json:"report_sha256"`
		}{report, deckproject.TeamSemanticReportHash(report)}
	} else {
		raw, e := readReconciliationInput(*decisions, 1<<20)
		if e != nil {
			return e
		}
		out, e = deckproject.AdoptTeamSemantics(p, packet, *slide, raw, b, en, *apply)
		if e != nil {
			return e
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}
