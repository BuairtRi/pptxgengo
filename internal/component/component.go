// Package component applies source-bound component contracts to native scenes.
package component

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
	ValueFormat string   `json:"value_format,omitempty"`
}
type Role struct {
	BindingIDs      []string              `json:"binding_ids"`
	Description     string                `json:"description"`
	ColorKind       string                `json:"color_kind,omitempty"`
	ResourceTargets []ResourceColorTarget `json:"resource_targets,omitempty"`
}
type Zone struct {
	ParagraphPath       []int      `json:"paragraph_path"`
	StyleDonorBindingID string     `json:"style_donor_binding_id,omitempty"`
	Style               *ZoneStyle `json:"style,omitempty"`
	Mode                string     `json:"mode,omitempty"`
	OverlayObjectID     string     `json:"overlay_object_id,omitempty"`
	FrameEMU            []int64    `json:"frame_emu,omitempty"`
	Description         string     `json:"description,omitempty"`
}
type ZoneStyle struct {
	FontSize int    `json:"font_size"`
	Typeface string `json:"typeface"`
	ColorRGB string `json:"color_rgb"`
	Bold     bool   `json:"bold,omitempty"`
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
	Zones        map[string]Zone              `json:"zones,omitempty"`
	Profiles     map[string]map[string]string `json:"profiles"`
	Constraints  []string                     `json:"constraints"`
}
type Values struct {
	Slots   map[string]json.RawMessage `json:"slots"`
	Zones   map[string]string          `json:"zones,omitempty"`
	Profile string                     `json:"profile,omitempty"`
}

func read(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := decode(b, out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
func decode(b []byte, out any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
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
func Run(args []string, stdout io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pptxcomponent inspect|check|apply --project DIR --contract FILE [--values FILE --out DIR]")
	}
	command := args[0]
	if command != "inspect" && command != "check" && command != "apply" {
		return fmt.Errorf("unknown command %s", command)
	}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	project := f.String("project", "", "native scene project")
	contractPath := f.String("contract", "", "reviewed JSON component contract")
	valuesPath := f.String("values", "", "JSON slot values and optional profile")
	out := f.String("out", "", "new scene project directory")
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 || *project == "" || *contractPath == "" {
		return fmt.Errorf("--project and --contract required; no positional arguments")
	}
	var c Contract
	contractData, err := os.ReadFile(*contractPath)
	if err != nil {
		return err
	}
	if err := decode(contractData, &c); err != nil {
		return err
	}
	if c.Schema != "pptxgengo.component-contract.v1" || c.ID == "" || c.ComponentID == "" || c.Slide < 1 {
		return fmt.Errorf("invalid contract identity/schema")
	}
	var m nativepkg.Manifest
	manifestData, err := os.ReadFile(filepath.Join(*project, "manifest.json"))
	if err != nil {
		return err
	}
	if err := decode(manifestData, &m); err != nil {
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
	validate := func(name string, ids []string, property string) error {
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
			if b.Property != property {
				return fmt.Errorf("%s requires %s; got %s", name, property, b.Property)
			}
			used[id] = true
		}
		return nil
	}
	for name, slot := range c.Slots {
		if slot.ValueFormat != "" && slot.ValueFormat != "paragraphs" {
			return fmt.Errorf("slot %s has unsupported value_format", name)
		}
		if err := validate(name, slot.BindingIDs, "text"); err != nil {
			return err
		}
	}
	for name, role := range c.Roles {
		if len(role.BindingIDs) == 0 && len(role.ResourceTargets) == 0 {
			return fmt.Errorf("role %s has no targets", name)
		}
		property := "fill.srgbClr.val"
		if role.ColorKind == "scheme" {
			property = "fill.schemeClr.val"
		} else if role.ColorKind != "" && role.ColorKind != "rgb" {
			return fmt.Errorf("role %s has unsupported color_kind", name)
		}
		if len(role.BindingIDs) > 0 {
			if err := validate(name, role.BindingIDs, property); err != nil {
				return err
			}
		}
	}
	hexColor := regexp.MustCompile(`^[A-Fa-f0-9]{6}$`)
	schemeColor := regexp.MustCompile(`^(accent[1-6]|dk[12]|lt[12]|tx[12]|bg[12]|hlink|folHlink)$`)
	for name, profile := range c.Profiles {
		if name == "" || len(profile) == 0 || len(profile) != len(c.Roles) {
			return fmt.Errorf("profile %q must bind every declared role", name)
		}
		for role, value := range profile {
			r, ok := c.Roles[role]
			if !ok || (r.ColorKind == "scheme" && !schemeColor.MatchString(value)) || (r.ColorKind != "scheme" && !hexColor.MatchString(value)) {
				return fmt.Errorf("invalid role/color in profile %q", name)
			}
		}
	}
	hasScheme := false
	for _, role := range c.Roles {
		if role.ColorKind == "scheme" {
			hasScheme = true
			break
		}
	}
	if hasScheme {
		source, ok := c.Profiles["source"]
		if !ok {
			return fmt.Errorf("scheme-color roles require a source profile")
		}
		for name, role := range c.Roles {
			if role.ColorKind != "scheme" {
				continue
			}
			for _, id := range role.BindingIDs {
				if source[name] != s.Bindings[bindings[id]].Value {
					return fmt.Errorf("source scheme profile role %s differs from pinned source", name)
				}
			}
		}
	}
	resourceTargets, err := validateResourceTargets(*project, c.Roles, c.Profiles, hexColor, schemeColor)
	if err != nil {
		return err
	}
	if err := validateZones(c.Zones, s.Scene, s.Bindings, bindings, objects, hexColor); err != nil {
		return err
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
		paragraphs, err := inspectParagraphs(c.Slots, s.Scene, s.Bindings, bindings)
		if err != nil {
			return err
		}
		report := map[string]any{"contract": c, "explicit_bindings": selected, "slot_paragraphs": paragraphs, "fit_status": "not_measured", "typography_status": "explicit values; inheritance unresolved"}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			return e
		}
		fmt.Fprintln(stdout, string(b))
		return nil
	}
	if *valuesPath == "" || (command == "apply" && *out == "") {
		return fmt.Errorf("check requires --values; apply requires --values and --out")
	}
	if command == "check" && *out != "" {
		return fmt.Errorf("check does not create output; omit --out")
	}
	var values Values
	valuesData, err := os.ReadFile(*valuesPath)
	if err != nil {
		return err
	}
	if err = decode(valuesData, &values); err != nil {
		return err
	}
	if len(values.Slots) == 0 && len(values.Zones) == 0 && values.Profile == "" {
		return fmt.Errorf("no edits requested")
	}
	changes := []map[string]string{}
	set := func(id, value, role string) {
		i := bindings[id]
		b := &s.Bindings[i]
		changes = append(changes, map[string]string{"binding_id": id, "role": role, "before": b.Value, "after": value})
		b.Value = value
	}
	for name, raw := range values.Slots {
		slot, ok := c.Slots[name]
		if !ok {
			return fmt.Errorf("unknown slot %q", name)
		}
		parts, err := parseSlotValue(raw, slot, s.Scene, s.Bindings, bindings)
		if err != nil {
			return fmt.Errorf("slot %s: %w", name, err)
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
	zoneChanges, err := applyZones(values.Zones, c.Zones, s.Scene, s.Bindings, bindings)
	if err != nil {
		return err
	}
	if values.Profile != "" {
		profile, ok := c.Profiles[values.Profile]
		if !ok {
			return fmt.Errorf("unsupported profile %q", values.Profile)
		}
		for role, value := range profile {
			for _, id := range c.Roles[role].BindingIDs {
				if c.Roles[role].ColorKind == "scheme" {
					set(id, value, "style:"+role)
				} else {
					set(id, strings.ToUpper(value), "style:"+role)
				}
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
	if command == "check" {
		actual := 0
		for _, change := range changes {
			if change["before"] != change["after"] {
				actual++
			}
		}
		resourceChanges := resourceProfileChanges(values.Profile, c.Roles, c.Profiles, resourceTargets)
		report := map[string]any{"contract_id": c.ID, "values_valid": true, "requested_bindings": len(changes), "changed_bindings": actual, "created_zones": len(zoneChanges), "resource_color_changes": resourceChanges, "fit_status": "not_measured", "adaptation_approved": false}
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, string(b))
		return nil
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
	stagedManifest, err := os.ReadFile(filepath.Join(stage, "manifest.json"))
	if err != nil {
		return err
	}
	stagedScene, err := os.ReadFile(filepath.Join(stage, sceneFile))
	if err != nil {
		return err
	}
	if hash(stagedManifest) != hash(manifestData) || hash(stagedScene) != c.SceneSHA256 {
		return fmt.Errorf("project changed during application copy")
	}
	if err = write(filepath.Join(stage, sceneFile), s); err != nil {
		return err
	}
	resourceChanges, err := applyResourceProfile(stage, values.Profile, c.Roles, c.Profiles, resourceTargets)
	if err != nil {
		return err
	}
	report := map[string]any{"schema": "pptxgengo.component-application.v1", "contract_id": c.ID, "component_id": c.ComponentID, "source_sha256": c.SourceSHA256, "reference_scene_sha256": c.SceneSHA256, "contract_sha256": hash(contractData), "values_sha256": hash(valuesData), "project_manifest_sha256": hash(manifestData), "changes": changes, "created_zones": zoneChanges, "resource_color_changes": resourceChanges, "profile": values.Profile, "fit_status": "requires_native_measurement_and_visual_review", "adaptation_approved": false}
	if err = write(filepath.Join(stage, "component-application.json"), report); err != nil {
		return err
	}
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("destination appeared during staging")
	}
	if err = os.Rename(stage, dest); err != nil {
		return err
	}
	fmt.Fprintln(stdout, dest)
	return nil
}
