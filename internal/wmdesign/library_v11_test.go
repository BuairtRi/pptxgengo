package wmdesign

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func v11IntakeBundle() string {
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v11", "intake-20261006-649-frozen", "bundle")
}

func v11Historical4fceBundle() string {
	return filepath.Join(filepath.Dir(v11IntakeBundle()), "historical-source", "4fce3cd", "bundle")
}

func v11ReadJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

func v11TemplateKeys(t *testing.T, name string, count int) []string {
	t.Helper()
	var keys []string
	v11ReadJSON(t, filepath.Join(filepath.Dir(v11IntakeBundle()), name+"-template-keys.json"), &keys)
	seen := map[string]bool{}
	for _, key := range keys {
		if key == "" || seen[key] {
			t.Fatalf("invalid audited key: %s", key)
		}
		seen[key] = true
	}
	if len(keys) != count {
		t.Fatalf("%s cardinality: %d; want %d", name, len(keys), count)
	}
	return keys
}

func v11KeySet(keys []string) map[string]bool {
	set := map[string]bool{}
	for _, key := range keys {
		set[key] = true
	}
	return set
}

func TestLibraryV11OrderedAuditKeys(t *testing.T) {
	var catalog struct {
		Templates []struct {
			Key string `json:"key"`
		} `json:"templates"`
	}
	v11ReadJSON(t, filepath.Join(v11IntakeBundle(), "source/templates/catalog.json"), &catalog)
	positions := map[string]int{}
	for i, entry := range catalog.Templates {
		positions[entry.Key] = i
	}
	for name, count := range map[string]int{"changed": 231, "density": 215, "pillar": 7, "unchanged": 418, "additional-render-review": 121, "native-review": 352, "inherit-eligible": 297} {
		previous := -1
		for _, key := range v11TemplateKeys(t, name, count) {
			position, exists := positions[key]
			if !exists || position <= previous {
				t.Fatalf("%s keys do not follow pinned catalog order: %s", name, key)
			}
			previous = position
		}
	}
}

func TestLibraryV11NativeReviewPartition(t *testing.T) {
	changed := v11KeySet(v11TemplateKeys(t, "changed", 231))
	unchanged := v11KeySet(v11TemplateKeys(t, "unchanged", 418))
	extra := v11KeySet(v11TemplateKeys(t, "additional-render-review", 121))
	review := v11KeySet(v11TemplateKeys(t, "native-review", 352))
	inherit := v11KeySet(v11TemplateKeys(t, "inherit-eligible", 297))
	for key := range extra {
		if !unchanged[key] || changed[key] {
			t.Fatalf("additional render specimen must have unchanged source: %s", key)
		}
	}
	for key := range changed {
		if !review[key] || inherit[key] {
			t.Fatalf("changed source requires new native review: %s", key)
		}
	}
	for key := range unchanged {
		if review[key] != extra[key] || inherit[key] == extra[key] {
			t.Fatalf("unchanged specimen partition invalid: %s", key)
		}
	}
}

func TestLibraryV11SupersededCandidateRoleDelta(t *testing.T) {
	var trail struct {
		Commit  string         `json:"superseded_source_commit"`
		Bundle  string         `json:"superseded_bundle_sha256"`
		Policy  map[string]any `json:"superseded_density_policy"`
		Delta   int            `json:"composition_delta_from_superseded"`
		Sources []struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
		} `json:"superseded_source_file_evidence"`
	}
	v11ReadJSON(t, filepath.Join(filepath.Dir(v11IntakeBundle()), "superseded-candidate-audit.json"), &trail)
	if trail.Commit != "d7027329522724b7c7225b0ca8309385cb7d0d42" || trail.Bundle != "b1de136b72ba09194d769c86993240cb1fbdce475068d426e31e4a92702fc92c" || trail.Delta != 0 || len(trail.Sources) != 40 {
		t.Fatal("superseded unqualified candidate evidence changed")
	}
	if _, exists := sourcePins[trail.Bundle]; exists {
		t.Fatal("superseded unqualified candidate remains an executable pin")
	}
	var tokens map[string]any
	v11ReadJSON(t, filepath.Join(v11Historical4fceBundle(), "source/tokens/v0/tokens.json"), &tokens)
	policy := tokens["density"].(map[string]any)
	body := policy["body"].(map[string]any)
	var expected any
	if err := json.Unmarshal([]byte(`[[13,17],[12,16],[11,14]]`), &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body["number-long"], expected) {
		t.Fatal("committed number-long metrics changed")
	}
	exceptions, ok := policy["exceptions"].([]any)
	if !ok || len(exceptions) != 2 {
		t.Fatal("fixed below-floor exceptions changed")
	}
	for i, want := range []struct {
		node string
		size float64
		font string
	}{{"gantt periods.sublabels", 7.5, "mono"}, {"reviewnote", 6.5, "mono caps"}} {
		exception := exceptions[i].(map[string]any)
		if exception["node"] != want.node || exception["font"] != want.font || !reflect.DeepEqual(exception["size"], []any{want.size, float64(9)}) || exception["reason"] == "" {
			t.Fatalf("fixed exception %d changed", i)
		}
	}
	delete(body, "number-long")
	delete(policy, "exceptions")
	if !reflect.DeepEqual(policy, trail.Policy) {
		t.Fatal("source refresh changed prior density role metrics")
	}
	changed := map[string]bool{}
	for _, entry := range trail.Sources {
		data, err := os.ReadFile(filepath.Join(v11Historical4fceBundle(), "source", entry.Path))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != entry.SHA {
			changed[entry.Path] = true
		}
	}
	wantFiles := v11KeySet([]string{"docs/authoring-reference.md", "explorations/components.src.html", "templates/change-notes.json", "templates/changelog.md", "templates/changes.json", "templates/library/argument.json", "templates/library/heatmaps.json", "tokens/v0/tokens.json"})
	if !reflect.DeepEqual(changed, wantFiles) {
		t.Fatalf("unexpected source refresh delta: %v", changed)
	}
}

func TestLibraryV11CommittedColorAndTileRefresh(t *testing.T) {
	var audit struct {
		Commit      string `json:"superseded_source_commit"`
		Bundle      string `json:"superseded_bundle_sha256"`
		Replacement string `json:"replacement_source_commit"`
		Browser     string `json:"browser_num_tile_implementation_commit"`
		Delta       map[string][]struct {
			Path   string `json:"path"`
			Before any    `json:"before"`
			After  any    `json:"after"`
		} `json:"composition_delta_from_superseded"`
		Sources []struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
		} `json:"superseded_source_file_evidence"`
		Intermediate []struct {
			Commit string `json:"source_commit"`
			Bundle string `json:"bundle_sha256"`
		} `json:"intermediate_unqualified_acquisitions"`
	}
	v11ReadJSON(t, filepath.Join(filepath.Dir(v11IntakeBundle()), "color-tile-source-refresh-audit.json"), &audit)
	if audit.Commit != "8c0dc2ff1176af73a14265835cccb277e4716447" || audit.Bundle != "493702876b53d7a0420283092ca0578fda509b39bcb1d9d2ed6c5d5125dbb4d1" || audit.Replacement != "4fce3cd7ecdfaf8daadfeb14bf6e4b596d731c9a" || audit.Browser != audit.Replacement || len(audit.Delta) != 4 || len(audit.Sources) != 40 {
		t.Fatal("final committed contrast refresh identity changed")
	}
	for _, pin := range append([]string{audit.Bundle}, func() []string {
		var pins []string
		for _, candidate := range audit.Intermediate {
			pins = append(pins, candidate.Bundle)
		}
		return pins
	}()...) {
		if _, exists := sourcePins[pin]; exists {
			t.Fatal("superseded unqualified contrast candidate remains executable")
		}
	}
	for key, changes := range audit.Delta {
		tile := key == "pillars/two-categories-six-magenta" || key == "pillars/two-categories-six-stacked-magenta"
		heat := key == "capability-heat/annotated" || key == "risk-heat/annotated"
		if !tile && !heat || tile && len(changes) != 6 || heat && len(changes) != 1 {
			t.Fatalf("unexpected contrast composition delta: %s", key)
		}
		for _, change := range changes {
			prop, want := "/titleInk", "primary"
			if tile {
				prop, want = "/numTile", "callout"
			}
			if change.Before != "<MISSING>" || change.After != want || !strings.HasSuffix(change.Path, prop) {
				t.Fatalf("contrast refresh changed non-approved field: %s %s", key, change.Path)
			}
		}
	}
	changed := map[string]bool{}
	for _, entry := range audit.Sources {
		data, err := os.ReadFile(filepath.Join(v11Historical4fceBundle(), "source", entry.Path))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != entry.SHA {
			changed[entry.Path] = true
		}
	}
	want := v11KeySet([]string{"docs/authoring-reference.md", "explorations/components.src.html", "templates/change-notes.json", "templates/changelog.md", "templates/changes.json", "templates/library/argument.json", "templates/library/heatmaps.json"})
	if !reflect.DeepEqual(changed, want) {
		t.Fatalf("contrast source refresh modified unexpected files: %v", changed)
	}
}

func TestLibraryV11FinalContrastPolicyAndHistoricalPin(t *testing.T) {
	latest, err := Load(v11IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := Load(v11Historical4fceBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if previous.Commit != "4fce3cd7ecdfaf8daadfeb14bf6e4b596d731c9a" || previous.densityVisualRules || !latest.densityVisualRules {
		t.Fatal("source-owned modern renderer capability changed historical source")
	}
	limited := v11KeySet(v11TemplateKeys(t, "density-limited", 10))
	policy := v11KeySet(v11TemplateKeys(t, "contrast-policy", 13))
	var defaults, restrictions int
	catalog, err := LibraryCatalogFromSource(latest)
	if err != nil {
		t.Fatal(err)
	}
	for _, def := range catalog {
		var slide map[string]any
		if err := json.Unmarshal(def.RawSlide, &slide); err != nil {
			t.Fatal(err)
		}
		if limited[def.Key] {
			if slide["densityLimit"] != "comfortable" || slide["density"] != nil || !policy[def.Key] {
				t.Fatalf("source density restriction changed: %s", def.Key)
			}
			restrictions++
		} else {
			if slide["densityLimit"] != nil {
				t.Fatalf("unexpected source density restriction: %s", def.Key)
			}
			defaults++
		}
	}
	if restrictions != 10 || defaults != 639 || defaults*3+restrictions != 1927 {
		t.Fatal("supported density configuration cardinality changed")
	}
	for _, file := range []string{"tokens/v0/tokens.json", "frames/v0/frames.json", "components/v0/components.json"} {
		a, err := os.ReadFile(filepath.Join(previous.Root, file))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(latest.Root, file))
		if err != nil || string(a) != string(b) {
			t.Fatalf("latest contrast policy changed prior token/component/frame source: %s", file)
		}
	}
}

func TestLibraryV11FrozenSnapshotClosure(t *testing.T) {
	bundle := v11IntakeBundle()
	for file, want := range map[string]string{
		"bundle.json":    "eb7dbb02d78b0a32b8bba552ba54bfd96f60e462985d7829f6fe91b4d00ff633",
		"inventory.json": "5fd96a039055e8d281c95fbb55fc472ca816308dfde34c62a0b869f223cacba1",
	} {
		data, err := os.ReadFile(filepath.Join(bundle, file))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != want {
			t.Fatalf("pin changed: %s: %v", file, err)
		}
	}
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if source.Revision != LibraryRevisionV11 || source.Commit != "3c56d842ba3abb5f24eb33cf0082be7a7a67f116" || len(source.Files) != 40 {
		t.Fatal("frozen V11 source identity or closure changed")
	}
	var audit struct {
		SourceCommit string `json:"source_commit"`
		Physical     int    `json:"physical_files"`
		Links        int    `json:"historical_links"`
		Evidence     []struct {
			Path  string `json:"path"`
			Blob  string `json:"git_blob"`
			SHA   string `json:"sha256"`
			Bytes int    `json:"bytes"`
		} `json:"source_file_evidence"`
	}
	v11ReadJSON(t, filepath.Join(filepath.Dir(bundle), "intake-audit.json"), &audit)
	if audit.SourceCommit != source.Commit || audit.Physical != 38 || audit.Links != 2 || len(audit.Evidence) != 40 {
		t.Fatal("committed Git acquisition evidence changed")
	}
	seen := map[string]bool{}
	for _, entry := range audit.Evidence {
		data, err := os.ReadFile(filepath.Join(source.Root, entry.Path))
		if err != nil || len(data) != entry.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != entry.SHA || seen[entry.Path] {
			t.Fatalf("source evidence drift: %s: %v", entry.Path, err)
		}
		blob := sha1.New()
		fmt.Fprintf(blob, "blob %d%c", len(data), 0)
		blob.Write(data)
		if fmt.Sprintf("%x", blob.Sum(nil)) != entry.Blob {
			t.Fatalf("committed Git blob mismatch: %s", entry.Path)
		}
		seen[entry.Path] = true
	}
	for _, entry := range source.Files {
		if !seen[entry.Path] {
			t.Fatalf("source path absent from Git evidence: %s", entry.Path)
		}
	}
	var manifests [2]struct {
		Files []SourceFile `json:"files"`
	}
	for i, root := range []string{v10IntakeBundle(), bundle} {
		v11ReadJSON(t, filepath.Join(root, "bundle.json"), &manifests[i])
	}
	if len(manifests[1].Files) != 15 || !reflect.DeepEqual(manifests[0].Files, manifests[1].Files) {
		t.Fatal("font/license/asset identities changed")
	}
	var before, after map[string]any
	v11ReadJSON(t, filepath.Join(v10IntakeBundle(), "source/tokens/v0/tokens.json"), &before)
	v11ReadJSON(t, filepath.Join(bundle, "source/tokens/v0/tokens.json"), &after)
	density, ok := after["density"].(map[string]any)
	if !ok || !reflect.DeepEqual(density["levels"], []any{"comfortable", "compact", "dense"}) {
		t.Fatal("missing committed typography density policy")
	}
	delete(after, "density")
	if !reflect.DeepEqual(before, after) {
		t.Fatal("prior typography/color/grid token fields changed")
	}
	before, after = nil, nil
	v11ReadJSON(t, filepath.Join(v10IntakeBundle(), "source/frames/v0/frames.json"), &before)
	v11ReadJSON(t, filepath.Join(bundle, "source/frames/v0/frames.json"), &after)
	feature := func(frames map[string]any) (int, map[string]any) {
		for i, value := range frames["features"].([]any) {
			item := value.(map[string]any)
			if item["id"] == "density" {
				return i, item
			}
		}
		t.Fatal("missing frame density metadata")
		return -1, nil
	}
	_, oldFeature := feature(before)
	i, newFeature := feature(after)
	oldModes, newModes := oldFeature["modes"].(map[string]any), newFeature["modes"].(map[string]any)
	if len(newModes) != 4 || !reflect.DeepEqual(oldModes["standard"], newModes["comfortable"]) || !reflect.DeepEqual(oldModes["appendix"], newModes["appendix"]) {
		t.Fatal("frame baseline or legacy appendix geometry changed")
	}
	for _, name := range []string{"compact", "dense"} {
		mode, ok := newModes[name].(map[string]any)
		if !ok || len(mode) != 1 || mode["note"] == nil {
			t.Fatalf("unexpected density frame geometry: %s", name)
		}
	}
	after["features"].([]any)[i] = oldFeature
	if !reflect.DeepEqual(before, after) {
		t.Fatal("frame geometry or chrome changed outside density policy metadata")
	}
}

// Restore only the exact removed small aliases. Remaining differences must be
// the separately audited pillar revisions, including their intentional copy.
func v11RestoreDensityAliases(before, after any) int {
	count := 0
	switch a := before.(type) {
	case map[string]any:
		b, ok := after.(map[string]any)
		if !ok {
			return 0
		}
		for key, value := range a {
			if _, exists := b[key]; !exists && (key == "size" || key == "bodySize") && value == "small" {
				b[key] = value
				count++
				continue
			}
			count += v11RestoreDensityAliases(value, b[key])
		}
	case []any:
		b, ok := after.([]any)
		if !ok {
			return 0
		}
		for i := 0; i < len(a) && i < len(b); i++ {
			count += v11RestoreDensityAliases(a[i], b[i])
		}
	}
	return count
}

func TestLibraryV11PinnedDensityAndPillarDelta(t *testing.T) {
	current, err := LibraryCatalog(v11IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := LibraryCatalog(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	changed := v11KeySet(v11TemplateKeys(t, "changed", 231))
	density := v11KeySet(v11TemplateKeys(t, "density", 215))
	pillars := v11KeySet(v11TemplateKeys(t, "pillar", 7))
	unchanged := v11KeySet(v11TemplateKeys(t, "unchanged", 418))
	old := map[string]LibraryTemplate{}
	for _, def := range previous {
		old[def.Key] = def
	}
	aliases, compact, appendix, defaults, deprecated, workshops, overlap := 0, 0, 0, 0, 0, 0, 0
	for _, def := range current {
		before, exists := old[def.Key]
		if !exists || def.SourceRevision != LibraryRevisionV11 {
			t.Fatalf("unexpected key/revision: %s", def.Key)
		}
		delete(old, def.Key)
		var a, b map[string]any
		if err := json.Unmarshal(before.RawSlide, &a); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(def.RawSlide, &b); err != nil {
			t.Fatal(err)
		}
		rawChanged := !reflect.DeepEqual(a, b)
		if rawChanged != changed[def.Key] || unchanged[def.Key] == rawChanged {
			t.Fatalf("audited composition delta changed: %s", def.Key)
		}
		switch b["density"] {
		case "compact":
			compact++
			if !density[def.Key] || a["density"] != nil {
				t.Fatalf("unexpected density migration: %s", def.Key)
			}
			delete(b, "density")
		case "appendix":
			appendix++
		case nil:
			defaults++
		default:
			t.Fatalf("unexpected stock policy: %s", def.Key)
		}
		if b["headerDensity"] != nil {
			t.Fatalf("source header policy changed: %s", def.Key)
		}
		aliases += v11RestoreDensityAliases(a, b)
		v11RemoveApprovedColorAndTileFields(t, def.Key, b)
		v11RestoreApprovedContrastPolicy(t, def.Key, b)
		if (!reflect.DeepEqual(a, b)) != pillars[def.Key] {
			t.Fatalf("copy/geometry/topology changed outside pillar audit: %s", def.Key)
		}
		if density[def.Key] && pillars[def.Key] {
			overlap++
		}
		if !pillars[def.Key] && (def.ContentContract != before.ContentContract || !reflect.DeepEqual(def.Slots, before.Slots) || !reflect.DeepEqual(def.Arrays, before.Arrays) || !reflect.DeepEqual(def.ValueSchema, before.ValueSchema)) {
			t.Fatalf("density-only migration changed content contract: %s", def.Key)
		}
		compiled, err := compileLibrarySlide(def.RawSlide, nil)
		if err != nil {
			t.Fatalf("%s: %v", def.Key, err)
		}
		if err := applyLibraryRefinements(def.Key, def.SourceRevision, &compiled); err != nil {
			t.Fatalf("inherited amendment rejected %s: %v", def.Key, err)
		}
		if def.Status == "deprecated" {
			deprecated++
		}
		if def.Family == "workshops" {
			workshops++
		}
	}
	if len(current) != 649 || len(previous) != 649 || len(old) != 0 || aliases != 659 || compact != 215 || appendix != 5 || defaults != 429 || deprecated != 1 || workshops != 29 || overlap != 4 {
		t.Fatalf("V11 delta drift: current=%d removed=%d aliases=%d compact=%d appendix=%d defaults=%d deprecated=%d workshops=%d overlap=%d", len(current), len(old), aliases, compact, appendix, defaults, deprecated, workshops, overlap)
	}
}

func v11RestoreApprovedContrastPolicy(t *testing.T, key string, slide map[string]any) {
	t.Helper()
	limited := v11KeySet(v11TemplateKeys(t, "density-limited", 10))
	if limited[key] {
		if slide["densityLimit"] != "comfortable" || slide["density"] != nil {
			t.Fatalf("approved comfortable-only density policy changed: %s", key)
		}
		delete(slide, "densityLimit")
	}
	indices, prop := []int{}, "ink"
	switch key {
	case "guide/deck-overview":
		indices = []int{3, 6, 9, 12, 15}
	case "status/steering-update":
		indices, prop = []int{6}, "keyInk"
	case "venn/four-text":
		indices = []int{3}
	}
	for _, i := range indices {
		node := slide["body"].([]any)[i].(map[string]any)
		if node[prop] != "emphasis" {
			t.Fatalf("approved Blue emphasis changed: %s/body/%d/%s", key, i, prop)
		}
		node[prop] = "callout"
	}
}

// Normalize only the four exact designer contrast amendments. The equality
// check in the caller then continues to reject any copy, geometry or topology
// change outside the separately audited seven pillar compositions.
func v11RemoveApprovedColorAndTileFields(t *testing.T, key string, slide map[string]any) {
	t.Helper()
	want, prop := 0, ""
	switch key {
	case "capability-heat/annotated", "risk-heat/annotated":
		want, prop = 1, "titleInk"
	case "pillars/two-categories-six-magenta", "pillars/two-categories-six-stacked-magenta":
		want, prop = 6, "numTile"
	default:
		return
	}
	count := 0
	for _, raw := range slide["body"].([]any) {
		node := raw.(map[string]any)
		if node["type"] != "card" {
			continue
		}
		if prop == "titleInk" && node["surface"] == "callout" {
			if node[prop] != "primary" {
				t.Fatalf("approved Grounded callout title missing: %s", key)
			}
			delete(node, prop)
			count++
		}
		if prop == "numTile" {
			if node[prop] != "callout" || node["numInk"] != "callout" || node["inlineNumber"] == nil {
				t.Fatalf("approved fixed number tile missing: %s", key)
			}
			delete(node, prop)
			count++
		}
	}
	if count != want {
		t.Fatalf("approved %s amendments for %s = %d, want %d", prop, key, count, want)
	}
}

func TestLibraryV11NativeContrastSourceRefresh(t *testing.T) {
	dir := filepath.Join(filepath.Dir(v11IntakeBundle()), "native-contrast-source-refresh")
	var audit struct {
		Commit           string `json:"source_commit"`
		Previous         string `json:"previous_commit"`
		CompositionDelta int    `json:"raw_composition_changes"`
		Paths            []struct {
			Path string `json:"path"`
			SHA  string `json:"after_sha256"`
		} `json:"changed_snapshot_paths"`
		Keys          []string `json:"changed_render_keys"`
		RegressionSHA string   `json:"browser_regression_sha256"`
		ToolSHA       string   `json:"browser_regression_tool_sha256"`
	}
	v11ReadJSON(t, filepath.Join(dir, "audit.json"), &audit)
	source, err := Load(v11IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if source.Commit != audit.Commit || audit.Previous != "ab8b0658b8a22773b5a8173f03a886ac7a64f59c" || audit.CompositionDelta != 0 || !source.densityContrastRules {
		t.Fatal("final contrast source capability or identity changed")
	}
	if !reflect.DeepEqual(audit.Keys, []string{"maturity/ai-beyond", "maturity/insights", "road-fork/decision-chosen"}) || len(audit.Paths) != 2 {
		t.Fatal("contrast refresh scope changed")
	}
	for _, entry := range audit.Paths {
		if entry.Path != "explorations/components.src.html" && entry.Path != "templates/changelog.md" {
			t.Fatal("unexpected contrast source edit")
		}
		data, err := os.ReadFile(filepath.Join(source.Root, entry.Path))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != entry.SHA {
			t.Fatalf("contrast source hash changed: %s", entry.Path)
		}
	}
	for file, want := range map[string]string{"browser-regression.json": audit.RegressionSHA, "check_density_contrast.js": audit.ToolSHA} {
		data, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != want {
			t.Fatalf("browser contrast regression evidence changed: %s", file)
		}
	}
	var regression struct {
		Status    string `json:"status"`
		Specimens int    `json:"primitiveSpecimens"`
		Rows      []struct {
			Key     string `json:"key"`
			Density string `json:"density"`
			Checks  []struct {
				FG      string `json:"fg"`
				Ring    string `json:"ring"`
				BG      string `json:"bg"`
				Warning string `json:"warning"`
			} `json:"checks"`
		} `json:"primitiveRows"`
	}
	v11ReadJSON(t, filepath.Join(dir, "browser-regression.json"), &regression)
	if regression.Status != "passed" || regression.Specimens != 9 || len(regression.Rows) != 9 {
		t.Fatal("incomplete browser contrast regression")
	}
	seen := map[string]bool{}
	for _, row := range regression.Rows {
		id := row.Key + ":" + row.Density
		if seen[id] || !v11KeySet(audit.Keys)[row.Key] || (row.Density != "comfortable" && row.Density != "compact" && row.Density != "dense") {
			t.Fatal("invalid browser contrast specimen")
		}
		seen[id] = true
		want, count := "#0047FF", 1
		if row.Key == "road-fork/decision-chosen" {
			want, count = "#070154", 6
		}
		if len(row.Checks) != count {
			t.Fatal("browser contrast check cardinality changed")
		}
		inactive := 0
		for _, check := range row.Checks {
			if strings.HasPrefix(check.Ring, "rgb(151, 164, 186)") {
				inactive++
			}
			if check.FG != want || check.BG != "#FFFFFF" || check.Warning != "" {
				t.Fatal("browser contrast check failed")
			}
		}
		if row.Key == "road-fork/decision-chosen" && inactive != 2 {
			t.Fatal("inactive pin outlines changed")
		}
	}
}
