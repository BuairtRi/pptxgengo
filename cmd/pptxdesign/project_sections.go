package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func runProjectSection(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: project section <list|add|rename|remove> --project PATH")
	}
	command := args[0]
	if command != "list" && command != "add" && command != "rename" && command != "remove" {
		return fmt.Errorf("unknown section command %q", command)
	}
	f := flag.NewFlagSet("project section "+command, flag.ContinueOnError)
	path := f.String("project", ".", "maintained project directory or deck.yaml")
	id := f.String("id", "", "stable section ID")
	title := f.String("title", "", "native section title; rename leaves visible divider copy unchanged")
	before := f.String("before", "", "stable slide ID at start of section")
	divider := f.String("divider", "", "optional actual divider/... library template")
	dividerID := f.String("divider-slide-id", "", "stable ID of inserted visible divider (defaults to SECTION-divider)")
	photo := f.String("divider-photo", "", "caller-selected photo asset for panel-edge/panel-photo/full-photo")
	valuesPath := f.String("divider-values", "", "JSON file containing complete closed divider binding values")
	bundle := f.String("bundle", "", "bundle path/revision; defaults to project lock")
	engine := f.String("engine", "", "engine; defaults to project lock")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected section positional arguments")
	}
	used := map[string]bool{}
	f.Visit(func(flag *flag.Flag) { used[flag.Name] = true })
	for flag := range used {
		if flag == "project" {
			continue
		}
		if command == "list" || (command == "remove" && flag != "id") || (command == "rename" && flag != "id" && flag != "title") {
			return fmt.Errorf("--%s is not supported by section %s", flag, command)
		}
	}
	p, e := deckproject.Load(*path)
	if e != nil {
		return e
	}
	var result any
	switch command {
	case "list":
		result = deckproject.ListSections(p)
	case "rename":
		result, e = deckproject.RenameSection(p, *id, *title)
	case "remove":
		result, e = deckproject.RemoveSection(p, *id)
	case "add":
		if *divider != "" {
			lock, _, err := deckproject.ReadLock(p)
			if err != nil {
				return err
			}
			if !used["bundle"] {
				*bundle = strings.TrimPrefix(lock.BundleRevision, "wmds-library.")
				if !validPublishedBundle(*bundle) || lock.BundleRevision != "wmds-library."+*bundle {
					return fmt.Errorf("unsupported locked bundle revision %q", lock.BundleRevision)
				}
				pinned := filepath.Join(p.Root, "runtime", "library", "wm-design-system", "pinned")
				if info, err := os.Stat(pinned); err == nil && info.IsDir() {
					*bundle = pinned
				}
			}
			if !used["engine"] {
				*engine = lock.Engine
			}
		} else if used["bundle"] || used["engine"] {
			return fmt.Errorf("--bundle/--engine require --divider")
		}
		var values map[string]any
		if *valuesPath != "" {
			raw, err := os.ReadFile(*valuesPath)
			if err != nil {
				return err
			}

			values, err = decodeSectionValues(raw)
			if err != nil {
				return err
			}
		}
		result, e = deckproject.AddSection(p, deckproject.SectionAddOptions{ID: *id, Title: *title, BeforeSlideID: *before, Divider: *divider, DividerSlideID: *dividerID, DividerPhoto: *photo, DividerValues: values, Bundle: deckproject.BundlePath(*bundle), Engine: *engine})
	}
	if e != nil {
		return e
	}
	out, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return e
	}
	fmt.Println(string(out))
	return nil
}

// Decode complete JSON while rejecting duplicate keys at every nested object.
func decodeSectionValues(raw []byte) (map[string]any, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var read func() (any, error)
	read = func() (any, error) {
		token, e := d.Token()
		if e != nil {
			return nil, e
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return token, nil
		}
		switch delim {
		case '{':
			object := map[string]any{}
			for d.More() {
				key, e := d.Token()
				if e != nil {
					return nil, e
				}
				name, ok := key.(string)
				if !ok {
					return nil, fmt.Errorf("invalid JSON object key")
				}
				if _, exists := object[name]; exists {
					return nil, fmt.Errorf("duplicate divider JSON key %q", name)
				}
				value, e := read()
				if e != nil {
					return nil, e
				}
				object[name] = value
			}
			close, e := d.Token()
			if e != nil || close != json.Delim('}') {
				return nil, fmt.Errorf("invalid JSON object termination")
			}
			return object, nil
		case '[':
			array := []any{}
			for d.More() {
				value, e := read()
				if e != nil {
					return nil, e
				}
				array = append(array, value)
			}
			close, e := d.Token()
			if e != nil || close != json.Delim(']') {
				return nil, fmt.Errorf("invalid JSON array termination")
			}
			return array, nil
		default:
			return nil, fmt.Errorf("unexpected JSON delimiter")
		}
	}
	value, e := read()
	if e != nil {
		return nil, e
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, fmt.Errorf("divider values must contain exactly one complete JSON object")
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("divider values require a JSON object")
	}
	return object, nil
}
