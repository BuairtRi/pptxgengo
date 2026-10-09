package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runProjectDiagram(args []string) error {
	if len(args) == 0 || (args[0] != "inspect" && args[0] != "patch" && args[0] != "connect" && args[0] != "arrange" && args[0] != "contain" && args[0] != "uncontain" && args[0] != "route") {
		return fmt.Errorf("usage: project diagram <inspect|patch|connect|arrange|contain|uncontain|route> --project PATH --slide ID; patch --patch FILE [--apply]")
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
	route := f.String("route", "straight", "straight, horizontal or vertical elbow")
	bend := f.Float64("bend", .5, "elbow bend fraction in [0,1]")
	nodes := f.String("nodes", "", "comma-separated source IDs for arrange, or native member names for containment")
	align := f.String("align", "", "left, right, top, bottom, center or middle")
	distribute := f.String("distribute", "", "horizontal or vertical")
	container := f.String("container", "", "native object name of the logical container")
	padding := f.Float64("padding", 12, "uniform container-axis padding in points")
	topPadding := f.Float64("padding-top", -1, "override top padding, for example to reserve a container header")
	rightPadding := f.Float64("padding-right", -1, "override right padding")
	bottomPadding := f.Float64("padding-bottom", -1, "override bottom padding")
	leftPadding := f.Float64("padding-left", -1, "override left padding")
	actor := f.String("actor", "", "named author for diagram edits")
	reason := f.String("reason", "", "adaptation reason for diagram edits")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	common := map[string]bool{"project": true, "slide": true, "bundle": true, "engine": true}
	allowed := map[string]map[string]bool{"inspect": {}, "route": {"patch": true, "apply": true}, "patch": {"patch": true, "apply": true}, "connect": {"id": true, "from": true, "to": true, "from-site": true, "to-site": true, "head": true, "style": true, "route": true, "bend": true, "actor": true, "reason": true, "apply": true}, "contain": {"container": true, "nodes": true, "padding": true, "padding-top": true, "padding-right": true, "padding-bottom": true, "padding-left": true, "actor": true, "reason": true, "apply": true}, "uncontain": {"nodes": true, "actor": true, "reason": true, "apply": true}, "arrange": {"nodes": true, "align": true, "distribute": true, "actor": true, "reason": true, "apply": true}}
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
	} else if args[0] == "route" {
		if *patch == "" {
			return fmt.Errorf("diagram route requires --patch FILE")
		}
		raw, err := readReconciliationInput(*patch, 1<<20)
		if err != nil {
			return err
		}
		request, err := deckproject.DecodeDiagramRoutePatch(raw, *patch)
		if err != nil {
			return err
		}
		result, e = deckproject.PlanDiagramRoute(p, *slide, request, b, en, *apply)
	} else if args[0] == "connect" {
		var bendValue *float64
		f.Visit(func(v *flag.Flag) {
			if v.Name == "bend" {
				bendValue = bend
			}
		})
		result, e = deckproject.ConnectRoutedDiagram(p, *slide, *id, *from, *fromSite, *to, *toSite, *head, *style, *route, bendValue, *actor, *reason, b, en, *apply)
	} else if args[0] == "contain" || args[0] == "uncontain" {
		ids := strings.Split(*nodes, ",")
		for i := range ids {
			ids[i] = strings.TrimSpace(ids[i])
		}
		pad := wmdesign.DiagramPadding{Top: *padding, Right: *padding, Bottom: *padding, Left: *padding}
		f.Visit(func(v *flag.Flag) {
			switch v.Name {
			case "padding-top":
				pad.Top = *topPadding
			case "padding-right":
				pad.Right = *rightPadding
			case "padding-bottom":
				pad.Bottom = *bottomPadding
			case "padding-left":
				pad.Left = *leftPadding
			}
		})
		if args[0] == "contain" && *container == "" {
			return fmt.Errorf("contain requires --container")
		}
		if args[0] == "uncontain" {
			pad = wmdesign.DiagramPadding{}
		}
		result, e = deckproject.ContainDiagram(p, *slide, ids, *container, pad, *actor, *reason, b, en, *apply)
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
