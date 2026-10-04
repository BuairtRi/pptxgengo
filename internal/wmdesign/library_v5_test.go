package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLibraryV5PinnedCatalogAndInheritance(t *testing.T) {
	root := filepath.Join("..", "..", "library", "wm-design-system")
	bundle := filepath.Join(root, "v5")
	s, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if s.Revision != LibraryRevisionV5 || s.Commit != "d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf" {
		t.Fatal("v5 source identity changed")
	}
	for file, want := range map[string]string{
		"inventory.json": "c4c64f6fc07eba612ccb2af6d05e70402fb26d583e42f44bbaa1325e49a57090",
		"bundle.json":    "0bdb3c6c7b327ed98a62b0527826068db371cb5d9cb43a1c6f44a1d4a74bf7eb",
	} {
		data, err := os.ReadFile(filepath.Join(bundle, file))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != want {
			t.Fatalf("pin changed: %s, %v", file, err)
		}
	}
	catalog, err := LibraryCatalog(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 587 {
		t.Fatalf("catalog count %d", len(catalog))
	}
	old := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source", "templates", "library"))
	count, revised, added := 0, 0, 0
	for _, def := range catalog {
		if def.SourceRevision != LibraryRevisionV5 {
			t.Fatal("catalog revision lost")
		}
		previous, ok := old[def.Key]
		unchanged := false
		if ok {
			a, _ := libraryObject(previous.Slide)
			b, _ := libraryObject(def.RawSlide)
			unchanged = reflect.DeepEqual(a, b)
			if !unchanged {
				revised++
			}
		} else {
			added++
		}
		if v5RetainedV4Compositions[def.Key] != unchanged {
			t.Fatalf("inheritance mismatch: %s", def.Key)
		}
		if unchanged {
			count++
		}
	}
	if count != 519 || len(v5RetainedV4Compositions) != count || revised != 3 || added != 65 {
		t.Fatalf("inventory drift: retained %d, revised %d, added %d", count, revised, added)
	}
	for _, rev := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, "unrecognized"} {
		if isExpandedLibrary(rev) {
			t.Fatalf("expanded semantics leaked into %s", rev)
		}
	}
}

func TestLibraryV5UnusedRowMetadataIsNotEditableContent(t *testing.T) {
	catalog, err := LibraryCatalog(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "")
	if err != nil {
		t.Fatal(err)
	}
	var def LibraryTemplate
	for _, candidate := range catalog {
		if candidate.Key == "interviews-readout/full" {
			def = candidate
		}
	}
	values := libraryExampleValues(def)
	changed := 0
	for _, slot := range def.Slots {
		if strings.HasPrefix(slot.SourcePointer, "/body/0/rows/") {
			parts := strings.Split(slot.SourcePointer, "/")
			if len(parts) >= 6 && parts[5] == "p" {
				t.Fatal("invisible duplicate row metadata exposed", slot.SourcePointer)
			}
		}
		if slot.SourcePointer == "/body/0/rows/0/n" {
			values.Slots[slot.Name] = json.RawMessage("\"New person\"")
			changed++
		}
		if slot.SourcePointer == "/body/0/rows/0/r" {
			values.Slots[slot.Name] = json.RawMessage("\"New role\"")
			changed++
		}
	}
	if changed != 2 {
		t.Fatal("visible person/role slots missing")
	}
	raw, _ := json.Marshal(values)
	if _, _, err := bindLibraryTemplate(def, BoundSlide{ID: "changed-visible", Template: def.Key, ContentKind: "synthetic_example", Values: raw}); err != nil {
		t.Fatal("unused metadata constrained caller copy", err)
	}
}

func TestLibraryV5InheritedBindingsPreserveV4Amendments(t *testing.T) {
	catalog, err := LibraryCatalog(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "")
	if err != nil {
		t.Fatal(err)
	}
	// Reconstruct the historical content projection from immutable intake
	// evidence in memory. Only v5 is loaded as an installed executable bundle.
	baseline, err := Load(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "")
	if err != nil {
		t.Fatal(err)
	}
	baseline.Revision = LibraryRevisionV4
	baseline.Templates = map[string]json.RawMessage{}
	baseline.Files = nil
	paths, err := filepath.Glob(filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source", "templates", "library", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		key := "templates/library/" + filepath.Base(path)
		baseline.Templates[key] = raw
		baseline.Files = append(baseline.Files, SourceFile{Path: key, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw))})
	}
	previous, err := libraryCatalog(baseline)
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]LibraryTemplate{}
	for _, def := range previous {
		old[def.Key] = def
	}
	for _, def := range catalog {
		if !v5RetainedV4Compositions[def.Key] {
			continue
		}
		prior := old[def.Key]
		if def.ContentContract != prior.ContentContract || !reflect.DeepEqual(def.Slots, prior.Slots) || !reflect.DeepEqual(def.Arrays, prior.Arrays) || !reflect.DeepEqual(def.ValueSchema, prior.ValueSchema) {
			t.Fatalf("inherited content API changed: %s", def.Key)
		}
		if def.ContentContract != LibraryBindingsContract {
			continue // Retained cards keep their independently checked typed API.
		}
		values, _ := json.Marshal(libraryExampleValues(def))
		input := BoundSlide{ID: "inherited", Template: def.Key, ContentKind: "synthetic_example", Values: values}
		before, _, err := bindLibraryTemplate(prior, input)
		if err != nil {
			t.Fatalf("v4 %s: %v", def.Key, err)
		}
		after, _, err := bindLibraryTemplate(def, input)
		if err != nil {
			t.Fatalf("v5 %s: %v", def.Key, err)
		}
		// Binding provenance must name its own source revision/file hash. The
		// composed content, geometry and amendment records must remain identical.
		before.TemplateBinding, after.TemplateBinding = nil, nil
		if def.Key == "value-types/hard-soft-right" {
			// This unchanged source inherits v4 content, then receives the
			// separately tested v5 ring-band clearance amendment.
			if err := applyV5NativeChartRefinement(def.Key, &before); err != nil {
				t.Fatal(err)
			}
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("inherited amendment changed: %s", def.Key)
		}
	}
}

func TestLibraryV5NullableLineBindings(t *testing.T) {
	catalog, err := LibraryCatalog(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "")
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
		if slot.Kind == "nullable_number" {
			if bytes.Equal(bytes.TrimSpace(slot.Example), []byte("null")) {
				nulls++
			} else {
				numeric = slot.Name
			}
		}
	}
	if nulls != 3 || numeric == "" {
		t.Fatal("missing observations are not required nullable slots")
	}
	bind := func(v LibraryValues) error {
		raw, _ := json.Marshal(v)
		_, _, err := bindLibraryTemplate(def, BoundSlide{ID: "nullable", Template: def.Key, ContentKind: "synthetic_example", Values: raw})
		return err
	}
	for _, value := range []json.RawMessage{json.RawMessage(" null "), json.RawMessage("0"), json.RawMessage("-2")} {
		values.Slots[numeric] = value
		if err := bind(values); err != nil {
			t.Fatal(err)
		}
	}
	values.Slots[numeric] = json.RawMessage(`"0"`)
	if bind(values) == nil {
		t.Fatal("numeric string accepted")
	}
	delete(values.Slots, numeric)
	if bind(values) == nil {
		t.Fatal("omitted required observation accepted")
	}
	values = libraryExampleValues(def)
	values.Slots["title"] = json.RawMessage("null")
	if bind(values) == nil {
		t.Fatal("ordinary text accepts null")
	}
}
