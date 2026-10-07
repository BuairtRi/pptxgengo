package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNativeEditingEditableTableRemovesOnlyPlainWrapper(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = density
		raw := `{"type":"table","x":57,"y":198,"w":414,"cols":[{"k":"step","label":"Step","w":207},{"k":"owner","label":"Owner","w":207}],"rows":[{"step":"Review","owner":"Reviewer"},{"step":"Build","owner":"Author"}]}`
		before, handled, err := r.planTableScene("table", json.RawMessage(raw), SceneContext{Surface: "light"})
		if err != nil || !handled {
			t.Fatal(density, handled, err)
		}
		after, handled, err := r.planTableScene("table", json.RawMessage(strings.Replace(raw, `"type":"table"`, `"type":"editable-table"`, 1)), SceneContext{Surface: "light"})
		if err != nil || !handled {
			t.Fatal(density, handled, err)
		}
		if len(before.Groups) != 1 || len(after.Groups) != 0 || len(after.Items) != 1 || after.Items[0].Table == nil || after.Definition != "scene.editable-table" {
			t.Fatal("wrapper/selection contract", density)
		}
		if !reflect.DeepEqual(before.Items, after.Items) || before.Bounds != after.Bounds {
			t.Fatal("table payload or geometry changed", density)
		}
	}
}

func TestNativeEditingEditableTableRefusesExternalAdornments(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, raw := range []string{
		`{"type":"editable-table","x":57,"y":156,"w":240,"rowH":48,"cols":[{"k":"score","label":"Score","w":240,"type":"heat"}],"rows":[{"score":3}]}`,
		`{"type":"editable-table","x":57,"y":156,"w":240,"rowH":48,"continued":"Continued","cols":[{"k":"step","label":"Step","w":240}],"rows":[{"step":"Review"}]}`,
	} {
		_, handled, err := r.planTableScene("table", json.RawMessage(raw), SceneContext{Surface: "light"})
		if !handled || err == nil || !strings.Contains(err.Error(), "editable_table_requires_one_native_table_without_external_adornments") {
			t.Fatal("decorated table accepted", raw)
		}
	}
}
