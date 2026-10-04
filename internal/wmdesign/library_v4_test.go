package wmdesign

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLibraryV4PinnedCatalogAndRetainedCompositions(t *testing.T) {
	root := filepath.Join("..", "..", "library", "wm-design-system")
	catalog, err := LibraryCatalog(filepath.Join(root, "v4"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 522 {
		t.Fatalf("catalog count %d", len(catalog))
	}
	old := intakeRepairEntries(t, filepath.Join(root, "v3", "source", "templates", "library"))
	count := 0
	for _, def := range catalog {
		if def.SourceRevision != LibraryRevisionV4 {
			t.Fatal("catalog identity lost")
		}
		previous, ok := old[def.Key]
		unchanged := false
		if ok {
			a, _ := libraryObject(previous.Slide)
			b, _ := libraryObject(def.RawSlide)
			unchanged = reflect.DeepEqual(a, b)
		}
		if v4RetainedV3Compositions[def.Key] != unchanged {
			t.Fatalf("retained composition mismatch: %s", def.Key)
		}
		if unchanged {
			count++
		}
	}
	if count != len(v4RetainedV3Compositions) || count != 232 {
		t.Fatal("retained catalog drift")
	}
}

func TestLibraryV4NullableLineBindings(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v4")
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
