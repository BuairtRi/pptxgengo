package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
	"strings"
)

// runProject is isolated from the legacy scene/semantic JSON build commands.
func runProject(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pptxdesign project <init|check|build|status|resume|approve|export|review|section> --project PATH [--bundle v1|v2|v3|v4|v5|PATH]")
	}
	if args[0] == "section" {
		return runProjectSection(args[1:])
	}
	if args[0] == "edit" {
		return runProjectEdit(args[1:])
	}
	if args[0] == "scaffold" {
		return runProjectScaffold(args[1:])
	}
	if args[0] == "measure" {
		return runProjectMeasure(args[1:])
	}
	cmd := args[0]
	f := flag.NewFlagSet("project "+cmd, flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	bundle := f.String("bundle", "", "bundle path or v1/v2/v3/v4/v5 (defaults to project lock; new projects use published bundle)")
	engine := f.String("engine", "", "engine (defaults to existing lock; init uses candidate v2)")
	out := f.String("out", "", "new export ZIP path")
	stage := f.String("stage", "", "approval stage")
	actor := f.String("actor", "", "named approval actor")
	slideIDs := f.String("slides", "", "comma-separated approval stable slide IDs; empty means whole deck")
	templateID := f.String("template", "", "local template to fork")
	newID := f.String("as", "", "new stable local template ID")
	reason := f.String("reason", "", "authoring decision explaining fork/detach")
	slideID := f.String("slide", "", "stable shared slide ID to detach")
	mode := f.String("mode", "client", "export: maintainer/client/offline/reviewer")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected project positional arguments")
	}
	allowed := map[string]bool{"init": true, "check": true, "build": true, "status": true, "resume": true, "approve": true, "export": true, "review": true, "fork": true, "detach": true}
	if !allowed[cmd] {
		return fmt.Errorf("unknown project command %s", cmd)
	}
	used := map[string]bool{}
	f.Visit(func(flag *flag.Flag) { used[flag.Name] = true })
	for key := range used {
		switch key {
		case "stage", "actor", "slides":
			if cmd != "approve" && !(cmd == "fork" && key == "slides") {
				return fmt.Errorf("--%s requires project approve", key)
			}
		case "template":
			if cmd != "fork" {
				return fmt.Errorf("--template requires project fork")
			}
		case "as", "reason":
			if cmd != "fork" && cmd != "detach" {
				return fmt.Errorf("--%s requires project fork/detach", key)
			}
		case "slide":
			if cmd != "detach" {
				return fmt.Errorf("--slide requires project detach")
			}
		case "mode":
			if cmd != "export" {
				return fmt.Errorf("--mode requires project export")
			}
		case "out":
			if cmd != "export" && cmd != "review" {
				return fmt.Errorf("--out requires project export/review")
			}
		}
	}
	p, e := deckproject.Load(*path)
	if e != nil {
		return e
	}
	lock, _, lockErr := deckproject.ReadLock(p)
	if lockErr != nil && !(cmd == "init" && os.IsNotExist(lockErr)) {
		return lockErr
	}
	if !used["bundle"] {
		if lockErr == nil {
			*bundle = strings.TrimPrefix(lock.BundleRevision, "wmds-library.")
			if !validPublishedBundle(*bundle) || lock.BundleRevision != "wmds-library."+*bundle {
				return fmt.Errorf("unsupported locked bundle revision %q; supply --bundle explicitly", lock.BundleRevision)
			}
			// Offline exports carry the exact locked bundle under this stable
			// project-relative path; normal Check still verifies every pin.
			pinned := filepath.Join(p.Root, "runtime", "library", "wm-design-system", "pinned")
			if info, err := os.Stat(pinned); err == nil && info.IsDir() {
				*bundle = pinned
			}
		} else {
			*bundle, e = publishedProjectBundle(os.Getenv("PPTXGENGO_RELEASE_ROOT"))
			if e != nil {
				return e
			}
		}
	} else if *bundle == "" {
		return fmt.Errorf("--bundle must not be empty")
	}
	b := deckproject.BundlePath(*bundle)
	if *engine == "" {
		if lockErr != nil {
			*engine = wmdesign.CandidateEngine
		} else {
			*engine = lock.Engine
		}
	}
	var result any
	switch cmd {
	case "init":
		result, e = deckproject.Pin(p, b, *engine)
	case "check":
		var c deckproject.Compilation
		c, e = deckproject.Check(p, b, *engine)
		result = map[string]any{"schema": p.Document.Schema, "deck_id": p.Document.ID, "slides": len(c.Document.Slides), "status": "source_and_pins_checked", "native_review": "not_performed"}
	case "build":
		result, e = deckproject.Build(p, deckproject.BuildOptions{Bundle: b, Engine: *engine})
	case "status":
		result, e = deckproject.Status(p)
	case "resume":
		result, e = deckproject.Resume(p)
	case "approve":
		ids := []string{}
		if *slideIDs != "" {
			ids = strings.Split(*slideIDs, ",")
		}
		result, e = deckproject.Approve(p, *stage, *actor, ids)
	case "export":
		result, e = deckproject.Export(p, deckproject.ExportOptions{Mode: *mode, Out: *out, Bundle: b})
	case "review":
		result, e = deckproject.ReviewPacket(p, *out)
	case "fork":
		ids := []string{}
		if *slideIDs != "" {
			ids = strings.Split(*slideIDs, ",")
		}
		result, e = deckproject.Fork(p, *templateID, *newID, *reason, ids)
	case "detach":
		result, e = deckproject.Detach(p, *slideID, *newID, b, *engine, *reason)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func validPublishedBundle(value string) bool {
	return value == "v1" || value == "v2" || value == "v3" || value == "v4" || value == "v5"
}

func publishedProjectBundle(root string) (string, error) {
	path := filepath.Join(root, "release", "default-bundle.txt")
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "v3", nil // Older installed releases predate explicit metadata.
	}
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(b))
	if !validPublishedBundle(value) {
		return "", fmt.Errorf("invalid published bundle in %s: %q", path, value)
	}
	return value, nil
}
