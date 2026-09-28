// pptxtemplate builds review candidates from the full template implementation queue.
// It deliberately keeps fixed-source editing separate from qualified dynamic layout.
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
	"sort"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/component"
	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

type entry struct {
	ID        string `json:"template_id"`
	Family    string `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"representative"`
	Category  string `json:"category"`
	Lane      string `json:"lane"`
	Project   string `json:"source_project"`
	Scene     string `json:"source_scene"`
	Directory string `json:"implementation_directory"`
}
type artifact struct {
	Path string `json:"path"`
	SHA  string `json:"sha256"`
}
type implementation struct {
	entry
	Contract     component.Contract `json:"contract"`
	Metadata     map[string]any     `json:"metadata"`
	ContractPath string             `json:"contract_path"`
	ValuesPath   string             `json:"values_path"`
	MetadataPath string             `json:"metadata_path"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "pptxtemplate:", err)
		os.Exit(1)
	}
}
func read(path string, dest any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = decodeJSON(b, dest); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
func decodeJSON(b []byte, dest any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	if err := d.Decode(dest); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}
func write(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
func hashFile(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
func pin(path string) (artifact, error) { h, e := hashFile(path); return artifact{path, h}, e }
func printJSON(v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e == nil {
		fmt.Println(string(b))
	}
	return e
}
func rooted(root, path string) (string, error) {
	if !filepath.IsLocal(path) {
		return "", fmt.Errorf("nonlocal repository path %s", path)
	}
	return filepath.Join(root, path), nil
}

// Resolve an absent output through any symlinks in its existing parents.
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
			return "", fmt.Errorf("cannot resolve output path")
		}
		missing = append(missing, filepath.Base(path))
		path = parent
	}
}

func validateMetadata(i implementation) error {
	m := i.Metadata
	if m["schema"] != "pptxgengo.template-implementation.v1" || m["template_id"] != i.ID ||
		m["family_id"] != i.Family || m["source"] != i.Source ||
		m["lane"] != i.Lane || m["category"] != i.Category {
		return fmt.Errorf("implementation identity mismatch for %s", i.ID)
	}
	if m["source_sha256"] != i.Contract.SourceSHA256 || m["scene_sha256"] != i.Contract.SceneSHA256 {
		return fmt.Errorf("metadata contract hash mismatch: %s", i.ID)
	}
	if m["adaptation_mode"] != "fixed_source_geometry" || m["native_fit"] != "not_measured" ||
		m["adaptation_qualified"] != false {
		return fmt.Errorf("unqualified implementation flags required: %s", i.ID)
	}
	if i.Contract.ID != "component-contract:"+i.ID || i.Contract.ComponentID == "" {
		return fmt.Errorf("contract identity mismatch for %s", i.ID)
	}
	return nil
}

func validatePreview(root string, i implementation) error {
	preview, ok := i.Metadata["preview"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing preview identity for %s", i.ID)
	}
	path, pathOK := preview["path"].(string)
	want, hashOK := preview["sha256"].(string)
	if !pathOK || !hashOK {
		return fmt.Errorf("invalid preview identity for %s", i.ID)
	}
	full, err := rooted(root, path)
	if err != nil {
		return err
	}
	got, err := hashFile(full)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("stale preview for %s", i.ID)
	}
	return nil
}

func load(root string, e entry) (implementation, error) {
	i := implementation{entry: e}
	d, err := rooted(root, e.Directory)
	if err != nil {
		return i, err
	}
	i.ContractPath = filepath.Join(d, "contract.json")
	i.ValuesPath = filepath.Join(d, "example-values.json")
	i.MetadataPath = filepath.Join(d, "implementation.json")
	if err = read(i.ContractPath, &i.Contract); err != nil {
		return i, err
	}
	if err = read(i.MetadataPath, &i.Metadata); err != nil {
		return i, err
	}
	if err = validateMetadata(i); err != nil {
		return i, err
	}
	if err = validatePreview(root, i); err != nil {
		return i, err
	}
	parts := strings.Split(e.Source, ":")
	if len(parts) != 2 {
		return i, fmt.Errorf("invalid source %s", e.Source)
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil || n != i.Contract.Slide {
		return i, fmt.Errorf("contract/source slide mismatch: %s", e.ID)
	}
	scene, err := rooted(root, e.Scene)
	if err != nil {
		return i, err
	}
	h, err := hashFile(scene)
	if err != nil {
		return i, err
	}
	if h != i.Contract.SceneSHA256 {
		return i, fmt.Errorf("stale scene: %s", e.ID)
	}
	project, err := rooted(root, e.Project)
	if err != nil {
		return i, err
	}
	if err = component.Run([]string{"inspect", "--project", project, "--contract", i.ContractPath}, io.Discard); err != nil {
		return i, fmt.Errorf("%s: %w", e.ID, err)
	}
	if err = component.Run([]string{"check", "--project", project, "--contract", i.ContractPath, "--values", i.ValuesPath}, io.Discard); err != nil {
		return i, fmt.Errorf("%s example: %w", e.ID, err)
	}
	return i, nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in source project: %s", path)
		}
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("nonregular source artifact")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0644)
	})
}

func run(args []string) error {
	if len(args) > 0 && args[0] == "apply-gauge" {
		return applyGauge(args[1:])
	}
	if len(args) > 0 && args[0] == "apply-accent" {
		return applyAccent(args[1:])
	}
	if len(args) > 0 && args[0] == "adapt-accents" {
		return adaptAccents(args[1:])
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: pptxtemplate list|inspect|values|components|build-review|adapt-accents|apply-accent|apply-gauge [--id ID] [--lane LANE] [--category CATEGORY] [--out NEW_DIR]")
	}
	cmd := args[0]
	f := flag.NewFlagSet(cmd, flag.ContinueOnError)
	rootFlag := f.String("root", ".", "repository root")
	id := f.String("id", "", "exact template ID")
	ids := f.String("ids", "", "comma-separated exact template IDs")
	sourceValues := f.Bool("source-values", false, "use original source content, preserving the layout as a reference")
	lane := f.String("lane", "", "workstream")
	category := f.String("category", "", "primary category")
	out := f.String("out", "", "new review bundle directory")
	values := f.String("values", "", "custom values file; requires --id")
	available := f.Bool("available", false, "build only complete implementations; still fail invalid implementations")
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if cmd != "list" && cmd != "inspect" && cmd != "build-review" && cmd != "values" && cmd != "components" {
		return fmt.Errorf("unknown command %s", cmd)
	}
	if *values != "" && cmd != "build-review" {
		return fmt.Errorf("--values is only supported by build-review")
	}
	if *sourceValues && cmd != "build-review" && cmd != "values" {
		return fmt.Errorf("--source-values is supported by values and build-review")
	}
	if *out != "" && cmd != "build-review" {
		return fmt.Errorf("--out is only supported by build-review")
	}
	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		return err
	}
	var assignments struct {
		Schema  string  `json:"schema"`
		Target  int     `json:"target"`
		Entries []entry `json:"entries"`
	}
	assignmentPath := filepath.Join(root, "library/templates/rollout/assignments.json")
	assignmentBytes, err := os.ReadFile(assignmentPath)
	if err != nil {
		return err
	}
	if err = decodeJSON(assignmentBytes, &assignments); err != nil {
		return err
	}
	if assignments.Schema != "pptxgengo.template-rollout.v1" || len(assignments.Entries) != assignments.Target {
		return fmt.Errorf("invalid assignment coverage")
	}
	selected := map[string]bool{}
	if *ids != "" {
		if *id != "" {
			return fmt.Errorf("choose --id or --ids")
		}
		for _, ident := range strings.Split(*ids, ",") {
			ident = strings.TrimSpace(ident)
			if ident == "" || selected[ident] {
				return fmt.Errorf("empty/duplicate --ids member")
			}
			selected[ident] = true
		}
	}
	if *sourceValues && *values != "" {
		return fmt.Errorf("choose --values or --source-values")
	}
	var chosen []entry
	seen := map[string]bool{}
	for _, e := range assignments.Entries {
		if e.ID == "" || seen[e.ID] {
			return fmt.Errorf("missing/duplicate template ID")
		}
		seen[e.ID] = true
		if (*id == "" || e.ID == *id) && (len(selected) == 0 || selected[e.ID]) && (*lane == "" || e.Lane == *lane) && (*category == "" || e.Category == *category) {
			chosen = append(chosen, e)
		}
	}
	for ident := range selected {
		if !seen[ident] {
			return fmt.Errorf("unknown template ID %s", ident)
		}
	}
	if len(chosen) == 0 {
		return fmt.Errorf("no matching templates")
	}
	if cmd == "list" {
		rows := []map[string]any{}
		for _, e := range chosen {
			status := "assigned"
			counts := map[string]int{}
			var problem string
			d, _ := rooted(root, e.Directory)
			_, err := os.Stat(filepath.Join(d, "contract.json"))
			if err == nil {
				status = "incomplete"
				i, loadErr := load(root, e)
				if loadErr == nil {
					status = "bindings_inspected"
					counts = map[string]int{"semantic_slots": len(i.Contract.Slots), "style_profiles": len(i.Contract.Profiles)}
				} else {
					problem = loadErr.Error()
				}
			}
			rows = append(rows, map[string]any{"id": e.ID, "family_id": e.Family, "name": e.Name, "source": e.Source, "category": e.Category, "lane": e.Lane, "status": status, "counts": counts, "issue": problem, "adaptation_qualified": false})
		}
		return printJSON(rows)
	}
	if cmd == "components" {
		if *id == "" || len(chosen) != 1 {
			return fmt.Errorf("components requires --id")
		}
		i, err := load(root, chosen[0])
		if err != nil {
			return err
		}
		var v any
		path := filepath.Join(root, i.Directory, "customization.json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return printJSON(map[string]any{"template_id": i.ID, "components": i.Metadata["components"], "roles": i.Contract.Roles, "profiles": i.Contract.Profiles, "scope": "Source groups; inspect the contract for exact editable bindings."})
		}
		if err := read(path, &v); err != nil {
			return err
		}
		return printJSON(v)
	}
	if cmd == "values" {
		if *id == "" || len(chosen) != 1 {
			return fmt.Errorf("values requires --id")
		}
		i, err := load(root, chosen[0])
		if err != nil {
			return err
		}
		var v any
		if *sourceValues {
			v, err = originalValues(root, i)
		} else {
			err = read(i.ValuesPath, &v)
		}
		if err != nil {
			return err
		}
		return printJSON(v)
	}
	if cmd == "inspect" {
		if *id == "" || len(chosen) != 1 {
			return fmt.Errorf("inspect requires --id")
		}
		i, err := load(root, chosen[0])
		if err != nil {
			return err
		}
		return printJSON(i)
	}
	if *out == "" {
		return fmt.Errorf("build-review requires --out NEW_DIR")
	}
	if *values != "" && (*id == "" || len(chosen) != 1) {
		return fmt.Errorf("custom --values requires one --id")
	}
	var ready []implementation
	var skipped []string
	for _, e := range chosen {
		d, err := rooted(root, e.Directory)
		if err != nil {
			return err
		}
		missing := false
		for _, name := range []string{"contract.json", "example-values.json", "implementation.json"} {
			if _, err := os.Stat(filepath.Join(d, name)); os.IsNotExist(err) {
				missing = true
			} else if err != nil {
				return err
			}
		}
		if missing && *available {
			skipped = append(skipped, e.ID)
			continue
		}
		if missing {
			return fmt.Errorf("implementation %s is not complete; other workers may still be preparing it", e.ID)
		}
		i, err := load(root, e)
		if err != nil {
			return err
		}
		if *values != "" {
			i.ValuesPath = *values
		}
		ready = append(ready, i)
	}
	if len(ready) == 0 {
		return fmt.Errorf("no complete implementations to build")
	}
	return build(root, *out, assignmentBytes, ready, skipped, *sourceValues)
}

func build(root, out string, assignmentBytes []byte, items []implementation, skipped []string, sourceValues bool) error {
	dest, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	dest, err = realDestination(dest)
	if err != nil {
		return err
	}
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("output already exists or cannot be checked: %s", dest)
	}
	for _, i := range items {
		src, err := rooted(root, i.Project)
		if err != nil {
			return err
		}
		src, err = filepath.EvalSymlinks(src)
		if err != nil {
			return err
		}
		if dest == src || strings.HasPrefix(dest, src+string(os.PathSeparator)) {
			return fmt.Errorf("review output must be outside source project")
		}
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".template-review-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	// Freeze all selected authoring files before any application. Every later
	// validation and application reads these copies, which are retained in the bundle.
	for n := range items {
		i := &items[n]
		if !filepath.IsLocal(i.ID) || strings.ContainsAny(i.ID, `/\\`) {
			return fmt.Errorf("invalid template ID path %q", i.ID)
		}
		snapshotDir := filepath.Join(stage, "inputs", i.ID)
		if err = os.MkdirAll(snapshotDir, 0755); err != nil {
			return err
		}
		for _, p := range []struct{ original, name string }{
			{i.ContractPath, "contract.json"}, {i.ValuesPath, "values.json"}, {i.MetadataPath, "implementation.json"},
		} {
			b, err := os.ReadFile(p.original)
			if sourceValues && p.name == "values.json" {
				var v any
				v, err = originalValues(root, *i)
				if err == nil {
					b, err = json.MarshalIndent(v, "", "  ")
				}
			}
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(snapshotDir, p.name), b, 0644); err != nil {
				return err
			}
		}
		i.ContractPath = filepath.Join(snapshotDir, "contract.json")
		i.ValuesPath = filepath.Join(snapshotDir, "values.json")
		i.MetadataPath = filepath.Join(snapshotDir, "implementation.json")
		if err = read(i.ContractPath, &i.Contract); err != nil {
			return err
		}
		if err = read(i.MetadataPath, &i.Metadata); err != nil {
			return err
		}
		if err = validateMetadata(*i); err != nil {
			return err
		}
		if err = validatePreview(root, *i); err != nil {
			return err
		}
	}
	assignmentSnapshot := filepath.Join(stage, "inputs", "assignments.json")
	if err = os.WriteFile(assignmentSnapshot, assignmentBytes, 0644); err != nil {
		return err
	}
	groups := map[string][]implementation{}
	for _, i := range items {
		source := strings.Split(i.Source, ":")[0]
		if !filepath.IsLocal(source) || strings.ContainsAny(source, `/\\`) {
			return fmt.Errorf("invalid source path %q", source)
		}
		groups[source] = append(groups[source], i)
	}
	names := []string{}
	for s := range groups {
		names = append(names, s)
	}
	sort.Strings(names)
	var decks []map[string]any
	var records []map[string]any
	for _, source := range names {
		members := groups[source]
		sort.Slice(members, func(a, b int) bool { return members[a].Contract.Slide < members[b].Contract.Slide })
		project := filepath.Join(stage, "projects", source)
		original, err := rooted(root, members[0].Project)
		if err != nil {
			return err
		}
		if err = copyTree(original, project); err != nil {
			return err
		}
		var numbers []int
		var pages []map[string]any
		seenSlides := map[int]bool{}
		for _, i := range members {
			if i.Project != members[0].Project || seenSlides[i.Contract.Slide] {
				return fmt.Errorf("inconsistent or duplicate source slide in %s", source)
			}
			seenSlides[i.Contract.Slide] = true
			fmt.Fprintln(os.Stderr, "applying", i.ID)
			inputs := []artifact{}
			for _, p := range []struct{ path, name string }{
				{i.ContractPath, "contract.json"}, {i.ValuesPath, "values.json"}, {i.MetadataPath, "implementation.json"},
			} {
				a, err := pin(p.path)
				if err != nil {
					return err
				}
				a.Path = filepath.Join("inputs", i.ID, p.name)
				inputs = append(inputs, a)
			}
			if err = component.Run([]string{"inspect", "--project", project, "--contract", i.ContractPath}, io.Discard); err != nil {
				return fmt.Errorf("%s snapshot: %w", i.ID, err)
			}
			if err = component.Run([]string{"check", "--project", project, "--contract", i.ContractPath, "--values", i.ValuesPath}, io.Discard); err != nil {
				return fmt.Errorf("%s snapshot values: %w", i.ID, err)
			}
			sceneRel, err := filepath.Rel(original, filepath.Join(root, i.Scene))
			if err != nil || !filepath.IsLocal(sceneRel) {
				return fmt.Errorf("scene outside source project")
			}
			sourceScene, err := os.ReadFile(filepath.Join(project, sceneRel))
			if err != nil {
				return err
			}
			sceneHash := sha256.Sum256(sourceScene)
			if hex.EncodeToString(sceneHash[:]) != i.Contract.SceneSHA256 {
				return fmt.Errorf("source scene changed before application: %s", i.ID)
			}
			sceneSnapshot := filepath.Join(stage, "inputs", i.ID, "source-scene.json")
			if err = os.WriteFile(sceneSnapshot, sourceScene, 0644); err != nil {
				return err
			}
			inputs = append(inputs, artifact{Path: filepath.Join("inputs", i.ID, "source-scene.json"), SHA: i.Contract.SceneSHA256})
			applied := filepath.Join(stage, "application")
			if err = component.Run([]string{"apply", "--project", project, "--contract", i.ContractPath, "--values", i.ValuesPath, "--out", applied}, io.Discard); err != nil {
				return fmt.Errorf("%s: %w", i.ID, err)
			}
			var changed nativepkg.Slide
			if err = read(filepath.Join(applied, sceneRel), &changed); err != nil {
				return err
			}
			wasHidden := changed.Scene.Attr("show") == "0" || changed.Scene.Attr("show") == "false"
			// Review copies expose selected hidden slides; dependencies remain hidden.
			if wasHidden {
				changed.Scene.SetAttr("show", "1")
			}
			if err = write(filepath.Join(project, sceneRel), changed); err != nil {
				return err
			}
			changeLog := map[string]any{}
			if err = read(filepath.Join(applied, "component-application.json"), &changeLog); err != nil {
				return err
			}
			// Component applications can edit retained chart/user-shape resources.
			// Copy only reported, hash-pinned resource outputs into this aggregate.
			resourceChanges, _ := changeLog["resource_color_changes"].([]any)
			copiedResources := map[string]bool{}
			for _, raw := range resourceChanges {
				r, ok := raw.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid resource change")
				}
				part, _ := r["part"].(string)
				sourceHash, _ := r["source_sha256"].(string)
				outputHash, _ := r["output_sha256"].(string)
				if !filepath.IsLocal(part) || !strings.HasPrefix(filepath.ToSlash(part), "resources/") {
					return fmt.Errorf("invalid resource path")
				}
				if copiedResources[part] {
					continue
				}
				copiedResources[part] = true
				before := filepath.Join(project, part)
				after := filepath.Join(applied, part)
				h, e := hashFile(before)
				if e != nil {
					return e
				}
				if h != sourceHash {
					return fmt.Errorf("aggregate resource changed: %s", part)
				}
				h, e = hashFile(after)
				if e != nil {
					return e
				}
				if h != outputHash {
					return fmt.Errorf("applied resource changed: %s", part)
				}
				data, e := os.ReadFile(before)
				if e != nil {
					return e
				}
				snapshotRel := filepath.Join("inputs", i.ID, part)
				snapshot := filepath.Join(stage, snapshotRel)
				if e = os.MkdirAll(filepath.Dir(snapshot), 0755); e != nil {
					return e
				}
				if e = os.WriteFile(snapshot, data, 0644); e != nil {
					return e
				}
				inputs = append(inputs, artifact{Path: snapshotRel, SHA: sourceHash})
				data, e = os.ReadFile(after)
				if e != nil {
					return e
				}
				if e = os.WriteFile(before, data, 0644); e != nil {
					return e
				}
			}
			logDir := filepath.Join(stage, "applications")
			if err = os.MkdirAll(logDir, 0755); err != nil {
				return err
			}
			if err = write(filepath.Join(logDir, i.ID+".json"), changeLog); err != nil {
				return err
			}
			changes, _ := changeLog["changes"].([]any)
			actual := 0
			for _, c := range changes {
				r, ok := c.(map[string]any)
				if ok && r["before"] != r["after"] {
					actual++
				}
			}
			if actual == 0 && !sourceValues && i.Metadata["example_kind"] != "source_reference" {
				return fmt.Errorf("%s example makes no actual text or style change", i.ID)
			}
			records = append(records, map[string]any{"template_id": i.ID, "family_id": i.Family, "source": i.Source, "inputs": inputs, "actual_binding_changes": actual, "selected_source_was_hidden": wasHidden, "native_fit": "not_measured", "adaptation_qualified": false, "limitations": i.Metadata["limitations"], "opaque_areas": i.Metadata["opaque_areas"], "retained_source_content": i.Metadata["retained_source_content"]})
			numbers = append(numbers, i.Contract.Slide)
			pages = append(pages, map[string]any{"pdf_page": len(pages) + 1, "template_id": i.ID, "source": i.Source, "source_slide_number": i.Contract.Slide})
			if err = os.RemoveAll(applied); err != nil {
				return err
			}
		}
		pptx := filepath.Join(stage, source+"-review.pptx")
		if err = nativepkg.Build(project, pptx, numbers, true); err != nil {
			return err
		}
		var nativeReport nativepkg.BuildReport
		if err = read(pptx+".build.json", &nativeReport); err != nil {
			return err
		}
		if len(nativeReport.Slides) != len(numbers) {
			return fmt.Errorf("native build page mapping differs for %s", source)
		}
		for n := range numbers {
			if nativeReport.Slides[n] != numbers[n] {
				return fmt.Errorf("native build page order differs for %s", source)
			}
		}
		h, err := hashFile(pptx)
		if err != nil {
			return err
		}
		buildReport, err := pin(pptx + ".build.json")
		if err != nil {
			return err
		}
		buildReport.Path = filepath.Base(pptx) + ".build.json"
		decks = append(decks, map[string]any{"path": filepath.Base(pptx), "sha256": h, "native_build_report": buildReport, "expected_pdf_pages": len(numbers), "pages": pages, "hidden_dependency_slides": nativeReport.HiddenDependencies, "source_slide_numbers_frozen": nativeReport.Frozen, "selected_hidden_slides_shown": true})
	}
	a, err := pin(assignmentSnapshot)
	if err != nil {
		return err
	}
	a.Path = filepath.Join("inputs", "assignments.json")
	report := map[string]any{"schema": "pptxgengo.template-review-bundle.v1", "assignments": a, "template_count": len(items), "source_values": sourceValues, "skipped_incomplete": skipped, "decks": decks, "templates": records, "status": "built_pending_native_render_and_review", "adaptation_qualified": false, "disclosure": "Source or edited review copies retain original assets and may retain source facts. source_values=true preserves original copy. Fixed source geometry/cardinality; not client-ready proposals or qualified dynamic templates."}
	if err = write(filepath.Join(stage, "review-bundle.json"), report); err != nil {
		return err
	}
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("destination appeared during build")
	}
	if err = os.Rename(stage, dest); err != nil {
		return err
	}
	return printJSON(map[string]any{"out": dest, "template_count": len(items), "source_decks": len(decks), "status": "built_pending_native_render_and_review"})
}

// originalValues copies native run contents, including their rich-run order.
// It never synthesizes copy or rewrites source facts as illustrative claims.
func originalValues(root string, i implementation) (any, error) {
	var scene nativepkg.Slide
	if err := read(filepath.Join(root, i.Scene), &scene); err != nil {
		return nil, err
	}
	bindings := map[string]nativepkg.Binding{}
	for _, b := range scene.Bindings {
		bindings[b.BindingID] = b
	}
	slots := map[string]any{}
	for name, slot := range i.Contract.Slots {
		parts := []string{}
		paras := []map[string]any{}
		last := ""
		for _, id := range slot.BindingIDs {
			b, ok := bindings[id]
			if !ok {
				return nil, fmt.Errorf("missing source binding %s", id)
			}
			parts = append(parts, b.Value)
			if slot.ValueFormat == "paragraphs" {
				if len(b.NodePath) < 3 {
					return nil, fmt.Errorf("invalid paragraph binding %s", id)
				}
				key := fmt.Sprint(b.NodePath[:len(b.NodePath)-3])
				if key != last || len(paras) == 0 {
					paras = append(paras, map[string]any{"runs": []map[string]string{}})
					last = key
				}
				p := paras[len(paras)-1]
				p["runs"] = append(p["runs"].([]map[string]string), map[string]string{"binding_id": id, "text": b.Value})
			}
		}
		if slot.ValueFormat == "paragraphs" {
			slots[name] = map[string]any{"paragraphs": paras}
		} else {
			slots[name] = parts
		}
	}
	return map[string]any{"slots": slots}, nil
}
