package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

func runProjectFinishedSlide(action string, args []string) error {
	f := flag.NewFlagSet("project slide "+action, flag.ContinueOnError)
	project := f.String("project", ".", "maintained deck directory or deck.yaml")
	id := f.String("id", "", "publish: source slide ID; insert: fresh destination slide ID")
	bundle := f.String("bundle", "", "bundle path/revision, defaults to project lock")
	engine := f.String("engine", "", "engine, defaults to project lock")
	if action == "publish" {
		out := f.String("out", "", "new closed revision directory; existing outputs are refused")
		libraryID := f.String("library-id", "", "curated/slide/<key> stable library identity")
		revision := f.Int("revision", 0, "positive immutable revision")
		name := f.String("name", "", "authored content name")
		purpose := f.String("purpose", "", "what this content communicates")
		owner := f.String("owner", "", "content steward")
		preview := f.String("preview", "", "optional reviewed preview PNG file")
		review := f.String("review", "", "optional review JSON file")
		if err := f.Parse(args); err != nil {
			return err
		}
		if f.NArg() != 0 {
			return fmt.Errorf("unexpected slide publish positional arguments")
		}
		p, err := deckproject.Load(*project)
		if err != nil {
			return err
		}
		b, e, err := projectRuntime(p, *bundle, *engine)
		if err != nil {
			return err
		}
		artifacts := map[string][]byte{}
		for name, path := range map[string]string{"preview.png": *preview, "review.json": *review} {
			if path == "" {
				continue
			}
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() || info.Size() > 16<<20 {
				return fmt.Errorf("review artifact must be regular and <=16MiB")
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			artifacts[name] = raw
		}
		// CLI publication always begins as draft. Approval is explicit stewardship
		// of an independently reviewed revision, not a side effect of compilation.
		m, err := deckproject.PublishFinishedSlide(p, deckproject.FinishedSlidePublishOptions{SlideID: *id, Out: *out, Bundle: b, Engine: e, Artifacts: artifacts, Manifest: finishedslide.Manifest{ID: *libraryID, Revision: *revision, Name: *name, Purpose: *purpose, Owner: *owner, Lifecycle: "draft"}})
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(m)
	}
	packagePath := f.String("package", "", "closed finished-slide revision directory")
	rationale := f.String("rationale", "", "authored reason this content belongs in this deck")
	before := f.String("before", "", "insert before stable slide ID")
	after := f.String("after", "", "insert after stable slide ID")
	section := f.String("into-section", "", "join this section")
	draft := f.Bool("allow-draft", false, "explicitly insert unapproved draft content for project review")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected slide insert positional arguments")
	}
	p, err := deckproject.Load(*project)
	if err != nil {
		return err
	}
	b, e, err := projectRuntime(p, *bundle, *engine)
	if err != nil {
		return err
	}
	r, err := deckproject.InsertFinishedSlide(p, deckproject.FinishedSlideInsertOptions{Package: *packagePath, ID: *id, Rationale: *rationale, Before: *before, After: *after, IntoSection: *section, AllowDraft: *draft, Bundle: b, Engine: e})
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
