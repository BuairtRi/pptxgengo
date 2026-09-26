// pptxcomponent applies reviewed, source-bound component contracts to scenes.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

type Slot struct {
	BindingIDs  []string `json:"binding_ids"`
	Description string   `json:"description"`
}
type Role struct {
	BindingIDs  []string `json:"binding_ids"`
	Description string   `json:"description"`
}
type Contract struct {
	Schema       string                       `json:"schema"`
	ID           string                       `json:"id"`
	ComponentID  string                       `json:"component_id"`
	SourceSHA256 string                       `json:"source_sha256"`
	Slide        int                          `json:"slide"`
	SceneSHA256  string                       `json:"scene_sha256"`
	ObjectIDs    []string                     `json:"object_ids"`
	Slots        map[string]Slot              `json:"slots"`
	Roles        map[string]Role              `json:"roles"`
	Profiles     map[string]map[string]string `json:"profiles"`
	Constraints  []string                     `json:"constraints"`
}
type Values struct {
	Slots   map[string][]string `json:"slots"`
	Profile string              `json:"profile,omitempty"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
func read(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return fmt.Errorf("%s: trailing JSON data", path)
	}
	return nil
}
func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func write(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}

// Resolve an as-yet nonexistent destination through existing parent aliases.
func realDestination(path string) (string, error) {
	var missing []string
	for {
		if _, err := os.Lstat(path); err == nil {
			real, err := filepath.EvalSymlinks(path)
			if err != nil {
				return "", err
			}
			for i := len(missing) - 1; i >= 0; i-- {
				real = filepath.Join(real, missing[i])
			}
			return real, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("cannot resolve destination")
		}
		missing = append(missing, filepath.Base(path))
		path = parent
	}
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: pptxcomponent inspect|apply --project DIR --contract FILE [--values FILE --out DIR]")
	}
	command := os.Args[1]
	if command != "inspect" && command != "apply" {
		return fmt.Errorf("unknown command %s", command)
	}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	project := f.String("project", "", "native scene project")
	contractPath := f.String("contract", "", "reviewed JSON component contract")
	valuesPath := f.String("values", "", "JSON slot values and optional profile")
	out := f.String("out", "", "new scene project directory")
	if err := f.Parse(os.Args[2:]); err != nil {
		return err
	}
	if f.NArg() != 0 || *project == "" || *contractPath == "" {
		return fmt.Errorf("--project and --contract required; no positional arguments")
	}
	var c Contract
	if err := read(*contractPath, &c); err != nil {
		return err
	}
	if c.Schema != "pptxgengo.component-contract.v1" || c.ID == "" || c.ComponentID == "" || c.Slide < 1 {
		return fmt.Errorf("invalid contract identity/schema")
	}
	var m nativepkg.Manifest
	if err := read(filepath.Join(*project, "manifest.json"), &m); err != nil {
		return err
	}
	if m.Schema != "pptxgengo.native-scene.v2" || m.SourceSHA256 != c.SourceSHA256 {
		return fmt.Errorf("contract source does not match v2 project")
	}
	var s nativepkg.Slide
	var sceneFile string
	for _, p := range m.Slides {
		if !filepath.IsLocal(p) {
			return fmt.Errorf("nonlocal manifest path %q", p)
		}
		var candidate nativepkg.Slide
		if err := read(filepath.Join(*project, p), &candidate); err != nil {
			return err
		}
		if candidate.Number == c.Slide {
			if sceneFile != "" {
				return fmt.Errorf("duplicate source slide")
			}
			s = candidate
			sceneFile = p
		}
	}
	if sceneFile == "" || s.ReferenceOnly {
		return fmt.Errorf("contract slide missing or reference-only")
	}
	data, err := os.ReadFile(filepath.Join(*project, sceneFile))
	if err != nil {
		return err
	}
	if hash(data) != c.SceneSHA256 {
		return fmt.Errorf("scene differs from pinned reference; rebase contract explicitly")
	}
	objects := map[string]bool{}
	for _, id := range c.ObjectIDs {
		if id == "" || objects[id] {
			return fmt.Errorf("empty/duplicate object ID")
		}
		objects[id] = true
	}
	bindings := map[string]int{}
	presentObjects := map[string]bool{}
	for i, b := range s.Bindings {
		if _, ok := bindings[b.BindingID]; ok {
			return fmt.Errorf("duplicate scene binding")
		}
		bindings[b.BindingID] = i
		presentObjects[b.ObjectID] = true
	}
	for id := range objects {
		if !presentObjects[id] {
			return fmt.Errorf("contract declares absent object %s", id)
		}
	}
	if len(c.Slots) == 0 {
		return fmt.Errorf("component contract must declare at least one text slot")
	}
	used := map[string]bool{}
	validate := func(name string, ids []string, color bool) error {
		if name == "" || len(ids) == 0 {
			return fmt.Errorf("empty slot/role")
		}
		for _, id := range ids {
			i, ok := bindings[id]
			if !ok {
				return fmt.Errorf("%s references absent binding %s", name, id)
			}
			b := s.Bindings[i]
			if !objects[b.ObjectID] || used[id] {
				return fmt.Errorf("unowned or reused binding %s", id)
			}
			want := "text"
			if color {
				want = "fill.srgbClr.val"
			}
			if b.Property != want {
				return fmt.Errorf("%s requires %s; got %s", name, want, b.Property)
			}
			used[id] = true
		}
		return nil
	}
	for name, slot := range c.Slots {
		if err := validate(name, slot.BindingIDs, false); err != nil {
			return err
		}
	}
	for name, role := range c.Roles {
		if err := validate(name, role.BindingIDs, true); err != nil {
			return err
		}
	}
	hexColor := regexp.MustCompile(`^[A-Fa-f0-9]{6}$`)
	for name, profile := range c.Profiles {
		if name == "" || len(profile) == 0 || len(profile) != len(c.Roles) {
			return fmt.Errorf("profile %q must bind every declared role", name)
		}
		for role, value := range profile {
			if _, ok := c.Roles[role]; !ok || !hexColor.MatchString(value) {
				return fmt.Errorf("invalid role/color in profile %q", name)
			}
		}
	}
	// Validate a separate decoded scene: ApplyBindings consumes its sentinels.
	var checked nativepkg.Slide
	if err = json.Unmarshal(data, &checked); err != nil {
		return err
	}
	if err = nativepkg.ApplyBindings(checked.Scene, checked.Bindings); err != nil {
		return err
	}
	if command == "inspect" {
		if *valuesPath != "" || *out != "" {
			return fmt.Errorf("inspect does not take --values or --out")
		}
		selected := []nativepkg.Binding{}
		for _, b := range s.Bindings {
			if objects[b.ObjectID] {
				selected = append(selected, b)
			}
		}
		report := map[string]any{"contract": c, "explicit_bindings": selected, "fit_status": "not_measured", "typography_status": "explicit values; inheritance unresolved"}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			return e
		}
		fmt.Println(string(b))
		return nil
	}
	if *valuesPath == "" || *out == "" {
		return fmt.Errorf("apply requires --values and --out")
	}
	var values Values
	if err = read(*valuesPath, &values); err != nil {
		return err
	}
	if len(values.Slots) == 0 && values.Profile == "" {
		return fmt.Errorf("no edits requested")
	}
	changes := []map[string]string{}
	set := func(id, value, role string) {
		i := bindings[id]
		b := &s.Bindings[i]
		changes = append(changes, map[string]string{"binding_id": id, "role": role, "before": b.Value, "after": value})
		b.Value = value
	}
	for name, parts := range values.Slots {
		slot, ok := c.Slots[name]
		if !ok {
			return fmt.Errorf("unknown slot %q", name)
		}
		if len(parts) != len(slot.BindingIDs) {
			return fmt.Errorf("slot %s needs exactly %d text segments; received %d", name, len(slot.BindingIDs), len(parts))
		}
		for i, value := range parts {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("empty text in slot %s", name)
			}
			for _, r := range value {
				if r < 32 || r == 127 {
					return fmt.Errorf("slot %s contains control/newline characters; use existing paragraph segments", name)
				}
			}
			if strings.HasPrefix(value, "__BINDING:") {
				return fmt.Errorf("reserved sentinel text")
			}
			set(slot.BindingIDs[i], value, "slot:"+name)
		}
	}
	if values.Profile != "" {
		profile, ok := c.Profiles[values.Profile]
		if !ok {
			return fmt.Errorf("unsupported profile %q", values.Profile)
		}
		for role, value := range profile {
			for _, id := range c.Roles[role].BindingIDs {
				set(id, strings.ToUpper(value), "style:"+role)
			}
		}
	}
	encoded, err := json.Marshal(s)
	if err != nil {
		return err
	}
	var validateOutput nativepkg.Slide
	if err = json.Unmarshal(encoded, &validateOutput); err != nil {
		return err
	}
	if err = nativepkg.ApplyBindings(validateOutput.Scene, validateOutput.Bindings); err != nil {
		return err
	}
	// Stage beside the destination, reject symlinks, and never mutate the input.
	src, err := filepath.Abs(*project)
	if err != nil {
		return err
	}
	src, err = filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}
	dest, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	dest, err = realDestination(dest)
	if err != nil {
		return err
	}
	if dest == src || strings.HasPrefix(dest, src+string(os.PathSeparator)) {
		return fmt.Errorf("output must be outside input project")
	}
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("output exists or cannot be checked: %s", dest)
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".component-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(src, path)
		if e != nil {
			return e
		}
		target := filepath.Join(stage, rel)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in project: %s", rel)
		}
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("nonregular project file: %s", rel)
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(target, b, 0644)
	})
	if err != nil {
		return err
	}
	if err = write(filepath.Join(stage, sceneFile), s); err != nil {
		return err
	}
	report := map[string]any{"schema": "pptxgengo.component-application.v1", "contract_id": c.ID, "component_id": c.ComponentID, "source_sha256": c.SourceSHA256, "reference_scene_sha256": c.SceneSHA256, "changes": changes, "profile": values.Profile, "fit_status": "requires_native_measurement_and_visual_review", "adaptation_approved": false}
	if err = write(filepath.Join(stage, "component-application.json"), report); err != nil {
		return err
	}
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("destination appeared during staging")
	}
	if err = os.Rename(stage, dest); err != nil {
		return err
	}
	fmt.Println(dest)
	return nil
}
