package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func projectRuntime(p *deckproject.Project, bundle, engine string) (string, string, error) {
	lock, _, err := deckproject.ReadLock(p)
	if err != nil {
		return "", "", err
	}
	if bundle == "" {
		bundle = strings.TrimPrefix(lock.BundleRevision, "wmds-library.")
		if !validLockedBundle(bundle) {
			return "", "", fmt.Errorf("unsupported locked bundle revision %q", lock.BundleRevision)
		}
		pinned := filepath.Join(p.Root, "runtime/library/wm-design-system/pinned")
		if stat, e := os.Stat(pinned); e == nil && stat.IsDir() {
			bundle = pinned
		}
	}
	if engine == "" {
		engine = lock.Engine
	}
	b := deckproject.BundlePath(bundle)
	if validLockedBundle(bundle) {
		b = designBundlePath(bundle)
	}
	return b, engine, nil
}

func runProjectSlide(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: project slide <add|insert|publish|review-reuse|move|remove|hide|show|draft-review> --project PATH --id ID")
	}
	if args[0] == "review-reuse" {
		return runFinishedSlideReview(args[1:])
	}
	if args[0] == "publish" || args[0] == "insert" {
		return runProjectFinishedSlide(args[0], args[1:])
	}
	if args[0] == "draft-review" {
		return runProjectSlideDraftReview(args[1:])
	}
	action := args[0]
	if action != "add" && action != "move" && action != "remove" && action != "hide" && action != "show" {
		return fmt.Errorf("unknown slide operation %q", action)
	}
	f := flag.NewFlagSet("project slide "+action, flag.ContinueOnError)
	path := f.String("project", ".", "project directory or deck.yaml")
	id := f.String("id", "", "stable slide ID")
	file := f.String("file", "", "add: supplied slide YAML file (original unchanged)")
	before := f.String("before", "", "position before this stable slide ID")
	after := f.String("after", "", "position after this stable slide ID")
	reanchor := f.Bool("reanchor", false, "move: leave an anchored section in place and report its replacement anchor")
	as := f.String("as", "", "add: assign this new stable ID without editing the supplied file")
	intoSection := f.String("into-section", "", "add/move: join this section; defaults to its first position and reanchors it")
	checkFit := f.Bool("check-fit", false, "add: check Go layout fit before committing")
	bundle := f.String("bundle", "", "add: bundle path/revision, defaults to lock")
	engine := f.String("engine", "", "add: engine, defaults to lock")
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected slide operation positional arguments")
	}
	var invalid string
	f.Visit(func(value *flag.Flag) {
		name := value.Name
		if name == "project" || name == "id" {
			return
		}
		if action == "add" && name != "reanchor" {
			return
		}
		if action == "move" && (name == "before" || name == "after" || name == "reanchor" || name == "into-section") {
			return
		}
		invalid = name
	})
	if invalid != "" {
		return fmt.Errorf("--%s is unsupported by slide %s", invalid, action)
	}
	p, err := deckproject.Load(*path)
	if err != nil {
		return err
	}
	o := deckproject.SlideOperation{Action: action, ID: *id, Before: *before, After: *after, Reanchor: *reanchor, As: *as, IntoSection: *intoSection, CheckFit: *checkFit}
	if action == "add" {
		if *file == "" {
			return fmt.Errorf("slide add requires --file")
		}
		o.Source, err = os.ReadFile(*file)
		if err != nil {
			return err
		}
		o.Bundle, o.Engine, err = projectRuntime(p, *bundle, *engine)
		if err != nil {
			return err
		}
	}
	receipt, err := deckproject.OperateSlide(p, o)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(receipt)
}
