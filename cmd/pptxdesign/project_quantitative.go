package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectQuantitative(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch" && args[0] != "reconcile") {
		return fmt.Errorf("usage: project quantitative inspect --project PATH --slide ID --node ID; project quantitative patch --project PATH --slide ID --patch FILE [--apply]; project quantitative reconcile --project PATH --slide ID --node ID --packet DIRECTORY [--decisions FILE [--apply]]")
	}
	f := flag.NewFlagSet("project quantitative "+args[0], flag.ContinueOnError)
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID; requires detached local semantic Quantitative component")
	node := f.String("node", "", "Quantitative node ID for inspect/reconcile; patch carries node_id")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to lock")
	engine := f.String("engine", "", "engine; defaults to lock")
	packet := f.String("packet", "", "closed receipt-backed geometry review packet for semantic reconcile")
	decisions := f.String("decisions", "", "explicit Quantitative semantic decisions; omission proposes only")
	patch := f.String("patch", "", "strict YAML/JSON Quantitative composition patch")
	apply := f.Bool("apply", false, "apply measured preview through guarded source transaction")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("Quantitative requires --slide and no positional arguments")
	}
	var bad error
	f.Visit(func(v *flag.Flag) {
		if args[0] == "inspect" && (v.Name == "patch" || v.Name == "apply" || v.Name == "packet" || v.Name == "decisions") || args[0] == "patch" && (v.Name == "node" || v.Name == "packet" || v.Name == "decisions") || args[0] == "reconcile" && v.Name == "patch" {
			bad = fmt.Errorf("--%s is not accepted by Quantitative %s", v.Name, args[0])
		}
	})
	if bad != nil {
		return bad
	}
	if args[0] == "inspect" && *node == "" {
		return fmt.Errorf("Quantitative inspect requires --node")
	}
	if args[0] == "patch" && *patch == "" {
		return fmt.Errorf("Quantitative patch requires --patch")
	}
	if args[0] == "reconcile" && (*node == "" || *packet == "" || *apply && *decisions == "") {
		return fmt.Errorf("Quantitative reconcile requires --node and --packet; --apply requires --decisions")
	}
	p, e := deckproject.Load(*project)
	if e != nil {
		return e
	}
	b, en, e := projectRuntime(p, *bundle, *engine)
	if e != nil {
		return e
	}
	var result any
	if args[0] == "inspect" {
		result, e = deckproject.InspectQuantitative(p, *slide, *node, b, en)
	} else if args[0] == "reconcile" {
		review, err := deckproject.ReadTextReviewPacket(*packet)
		if err != nil {
			return err
		}
		if *decisions == "" {
			report, err := deckproject.ProposeQuantitativeSemantics(p, review, *slide, *node, b, en)
			if err != nil {
				return err
			}
			result = struct {
				ReportSHA256 string                                 `json:"report_sha256"`
				Report       deckproject.QuantitativeSemanticReport `json:"report"`
			}{deckproject.QuantitativeSemanticReportHash(report), report}
		} else {
			raw, err := readReconciliationInput(*decisions, 1<<20)
			if err != nil {
				return err
			}
			result, e = deckproject.AdoptQuantitativeSemantics(p, review, *slide, *node, raw, b, en, *apply)
		}
	} else {
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeQuantitativePatch(raw, *patch)
		if err != nil {
			return err
		}
		result, e = deckproject.PatchQuantitative(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
