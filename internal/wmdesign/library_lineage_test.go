package wmdesign

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLibraryV5RetainedQualifiedCompositionLineage(t *testing.T) {
	catalog, err := LibraryCatalog(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 587 {
		t.Fatalf("catalog count %d", len(catalog))
	}
	// The historical intake observation records which 248 qualified source
	// compositions survived the first expansion. It is evidence, not an installed
	// legacy bundle. Check that lineage against the sole current library.
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen")
	old := intakeRepairEntries(t, filepath.Join(root, "source", "templates", "library"))
	data, err := os.ReadFile(filepath.Join(root, "observation.json"))
	if err != nil {
		t.Fatal(err)
	}
	var observation struct {
		Delta struct {
			Added, Revised []string
		} `json:"committed_delta_from_v3"`
	}
	if err := json.Unmarshal(data, &observation); err != nil {
		t.Fatal(err)
	}
	excluded := map[string]bool{}
	for _, keys := range [][]string{observation.Delta.Added, observation.Delta.Revised} {
		for _, key := range keys {
			excluded[key] = true
		}
	}
	for key := range old {
		if v4RetainedV3Compositions[key] != !excluded[key] {
			t.Fatalf("historical qualified lineage mismatch: %s", key)
		}
	}
	count := 0
	carried := 0
	for _, def := range catalog {
		if def.SourceRevision != LibraryRevisionV5 {
			t.Fatal("catalog identity lost")
		}
		if !v4RetainedV3Compositions[def.Key] {
			continue
		}
		previous, ok := old[def.Key]
		count++
		if !v5RetainedV4Compositions[def.Key] {
			// A revised current source receives its own guarded repair rather
			// than inheriting a historical composition amendment blindly.
			continue
		}
		a, _ := libraryObject(previous.Slide)
		b, _ := libraryObject(def.RawSlide)
		if !ok || !v5RetainedV4Compositions[def.Key] || !reflect.DeepEqual(a, b) {
			t.Fatalf("current library lost qualified composition: %s", def.Key)
		}
		carried++
	}
	if count != len(v4RetainedV3Compositions) || count != 232 {
		t.Fatal("retained catalog drift")
	}
	if carried != 231 {
		t.Fatalf("current qualified lineage drift: %d", carried)
	}
}

func TestLibraryV5RequiredNullableObservations(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	catalog, err := LibraryCatalog(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	var def LibraryTemplate
	for _, d := range catalog {
		if d.Key == "readiness/adoption-curve" {
			def = d
		}
	}
	values := libraryExampleValues(def)
	nulls, numeric := 0, ""
	for _, slot := range def.Slots {
		if slot.Kind != "nullable_number" {
			continue
		}
		if bytes.Equal(bytes.TrimSpace(slot.Example), []byte("null")) {
			nulls++
		} else {
			numeric = slot.Name
		}
	}
	if nulls != 3 || numeric == "" {
		t.Fatal("missing source observations omitted from required slots")
	}
	bind := func(v LibraryValues) (SlideSpec, error) {
		raw, _ := json.Marshal(v)
		s, _, err := bindLibraryTemplate(def, BoundSlide{ID: "nullable", Template: def.Key, ContentKind: "synthetic_example", Values: raw})
		return s, err
	}
	if _, err := bind(values); err != nil {
		t.Fatal(err)
	}
	values.Slots[numeric] = json.RawMessage(" null ")
	if _, err := bind(values); err != nil {
		t.Fatal("observed number cannot become missing", err)
	}
	values.Slots[numeric] = json.RawMessage("0")
	if _, err := bind(values); err != nil {
		t.Fatal("observed zero rejected", err)
	}
	values.Slots[numeric] = json.RawMessage(`"0"`)
	if _, err := bind(values); err == nil {
		t.Fatal("numeric string accepted")
	}
	delete(values.Slots, numeric)
	if _, err := bind(values); err == nil {
		t.Fatal("missing required slot accepted")
	}
	values = libraryExampleValues(def)
	values.Slots["title"] = json.RawMessage(" null ")
	if _, err := bind(values); err == nil {
		t.Fatal("ordinary text slot accepts null")
	}
}
