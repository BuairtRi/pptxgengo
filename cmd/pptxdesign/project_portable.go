package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func runProjectPortable(command string, args []string) error {
	action := ""
	if command == "version" {
		if len(args) == 0 {
			return fmt.Errorf("usage: project version <save|list|verify|materialize|recover> --project PATH")
		}
		action, args = args[0], args[1:]
	}
	f := flag.NewFlagSet("project "+command+" "+action, flag.ContinueOnError)
	archive := f.String("archive", "", "complete private project ZIP")
	path := f.String("project", ".", "project directory")
	apply := f.Bool("apply", false, "apply lossless layout migration; default dry run")
	dry := f.Bool("dry-run", false, "preview migration without changing files")
	actor := f.String("actor", "", "named snapshot author")
	message := f.String("message", "", "snapshot description")
	number := f.String("number", "", "six-digit numbered version")
	expect := f.String("expect-current-sha256", "", "recover: exact pointer preimage hash, or absent")
	out := f.String("out", "", "new materialization directory or private share ZIP outside project")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	used := map[string]bool{}
	f.Visit(func(x *flag.Flag) { used[x.Name] = true })
	for key := range used {
		allowed := key == "project" || (command == "layout" && (key == "apply" || key == "dry-run")) || (command == "share" && key == "out") || (command == "share-extract" && (key == "archive" || key == "out")) || (command == "version" && ((action == "save" && (key == "actor" || key == "message")) || ((action == "verify" || action == "materialize" || action == "recover") && key == "number") || (action == "materialize" && key == "out") || (action == "recover" && key == "expect-current-sha256")))
		if !allowed {
			return fmt.Errorf("--%s is not applicable to project %s %s", key, command, action)
		}
	}
	if command == "share-extract" {
		if used["project"] {
			return fmt.Errorf("share-extract uses --archive and --out")
		}
		result, e := deckproject.ExtractShare(*archive, *out)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	if command == "share-verify" || (command == "version" && action != "save") {
		absolute, e := filepath.Abs(*path)
		if e != nil {
			return e
		}
		info, e := os.Stat(absolute)
		if e != nil {
			return e
		}
		root := absolute
		if !info.IsDir() {
			root = filepath.Dir(absolute)
			if _, e = deckproject.SafePath(root, filepath.Base(absolute)); e != nil {
				return e
			}
		}
		var result any
		switch {
		case command == "share-verify":
			result, e = deckproject.VerifyShare(root)
		case action == "list":
			result, e = deckproject.ListVersions(root)
		case action == "verify":
			result, e = deckproject.VerifyVersion(root, *number)
		case action == "materialize":
			if *out == "" {
				return fmt.Errorf("materialize requires --out")
			}
			result, e = deckproject.MaterializeVersion(root, *number, *out)
		case action == "recover":
			result, e = deckproject.RecoverVersion(root, *number, *expect)
		default:
			return fmt.Errorf("unknown version action %s", action)
		}
		if e != nil {
			return e
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
	p, e := deckproject.Load(*path)
	if e != nil {
		return e
	}
	var result any
	switch command {
	case "layout":
		if *apply && *dry {
			return fmt.Errorf("choose --apply or --dry-run")
		}
		result, e = deckproject.PortableLayout(p, *apply)
	case "version":
		result, e = deckproject.SaveVersion(p, *actor, *message)
	case "share":
		result, e = deckproject.ShareProject(p, *out)
	default:
		return fmt.Errorf("unknown portable command %s", command)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func runProjectCreate(args []string) error {
	f := flag.NewFlagSet("project create", flag.ContinueOnError)
	out := f.String("out", "", "new project directory")
	id := f.String("id", "", "stable deck ID")
	title := f.String("title", "", "deck title")
	year := f.Int("year", 2026, "deck year")
	bundle := f.String("bundle", currentDesignBundle, "shared bundle revision or path")
	template := f.String("template", "", "initial shared template key")
	engine := f.String("engine", wmdesign.CandidateEngine, "pinned compiler engine")
	editing := f.String("editing-profile", "", "persisted editing profile: native-v1 (default for v2) or stock")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	b := deckproject.BundlePath(*bundle)
	if validLockedBundle(*bundle) {
		b = designBundlePath(*bundle)
	}
	p, e := deckproject.CreateProject(deckproject.CreateOptions{Out: *out, ID: *id, Title: *title, Year: *year, Bundle: b, Template: *template, Engine: *engine, EditingProfile: *editing})
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"schema": "pptxgengo.project-create.v1", "project_id": p.Document.ID, "path": p.Root, "source_sha256": p.SourceHash(), "next_action": "Replace scaffold examples, check, build, and save a complete numbered version; native review pending"})
}
