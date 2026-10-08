package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/deckproject"
	"os"
	"strings"
)

func runProjectDiagram(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch" && args[0] != "connect" && args[0] != "arrange") {
		return fmt.Errorf("usage: project diagram <inspect|patch|connect|arrange> --project PATH --slide ID; patch --patch FILE [--apply]")
	}
	f := flag.NewFlagSet("project diagram "+args[0], flag.ContinueOnError)
	project := f.String("project", ".", "project directory or deck.yaml")
	slide := f.String("slide", "", "stable slide ID; requires a detached/local template")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to lock")
	engine := f.String("engine", "", "engine; defaults to lock")
	patch := f.String("patch", "", "strict YAML/JSON diagram patch; add, remove, move, resize, transform")
	apply := f.Bool("apply", false, "commit the validated preview through a guarded source transaction")
	id := f.String("id", "", "new connection node ID")
	from := f.String("from", "", "source block ID")
	to := f.String("to", "", "destination block ID")
	fromSite := f.String("from-site", "right", "top, left, bottom or right")
	toSite := f.String("to-site", "left", "top, left, bottom or right")
	head := f.String("head", "end", "none, start, end or both")
	style := f.String("style", "solid", "solid, dashed or dotted")
	nodes := f.String("nodes", "", "comma-separated node IDs in desired distribution order; first node is alignment anchor")
	align := f.String("align", "", "left, right, top, bottom, center or middle")
	distribute := f.String("distribute", "", "horizontal or vertical")
	actor := f.String("actor", "", "named author for connect/arrange")
	reason := f.String("reason", "", "adaptation reason for connect/arrange")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	common := map[string]bool{"project": true, "slide": true, "bundle": true, "engine": true}
	allowed := map[string]map[string]bool{"inspect": {}, "patch": {"patch": true, "apply": true}, "connect": {"id": true, "from": true, "to": true, "from-site": true, "to-site": true, "head": true, "style": true, "actor": true, "reason": true, "apply": true}, "arrange": {"nodes": true, "align": true, "distribute": true, "actor": true, "reason": true, "apply": true}}
	var flagErr error
	f.Visit(func(v *flag.Flag) {
		if !common[v.Name] && !allowed[args[0]][v.Name] {
			flagErr = fmt.Errorf("--%s is not accepted by diagram %s", v.Name, args[0])
		}
	})
	if flagErr != nil {
		return flagErr
	}
	if f.NArg() != 0 || *slide == "" {
		return fmt.Errorf("diagram requires --slide ID and no positional arguments")
	}
	if args[0] == "inspect" && (*patch != "" || *apply) {
		return fmt.Errorf("inspect does not accept patch/apply")
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
		result, e = deckproject.InspectDiagram(p, *slide, b, en)
	} else if args[0] == "connect" {
		result, e = deckproject.ConnectDiagram(p, *slide, *id, *from, *fromSite, *to, *toSite, *head, *style, *actor, *reason, b, en, *apply)
	} else if args[0] == "arrange" {
		ids := strings.Split(*nodes, ",")
		for i := range ids {
			ids[i] = strings.TrimSpace(ids[i])
		}
		result, e = deckproject.ArrangeDiagram(p, *slide, ids, *align, *distribute, *actor, *reason, b, en, *apply)
	} else {
		if *patch == "" {
			return fmt.Errorf("diagram patch requires --patch FILE")
		}
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		ops, err := deckproject.DecodeDiagramPatch(raw, *patch)
		if err != nil {
			return err
		}
		result, e = deckproject.PatchDiagram(p, *slide, ops, b, en, *apply)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
