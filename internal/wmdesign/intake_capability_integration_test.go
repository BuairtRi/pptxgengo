package wmdesign

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestIntakeCapabilityBindingsAndDiscovery(t *testing.T) {
	raw := json.RawMessage(`{"type":"slide","rail":"none","footer":"compact","body":[
	{"type":"venn","x":57,"y":144,"w":500,"h":300,"labelStyle":"mono","sets":[{"label":"A","bullets":["First"]},{"label":"B"}],"regions":[{"in":[0,1],"label":"Shared"}],"points":[{"x":0.5,"y":0.5,"label":"Spot","n":2,"side":"left"}]},
	{"type":"maturity","x":57,"y":144,"w":500,"h":300,"at":[0.1,0.9],"active":1,"stages":[{"label":"First","n":1},{"label":"Second","n":2}],"branch":{"from":0,"n":3,"label":"Branch"}},
	{"type":"table","x":57,"y":144,"w":400,"cols":[{"k":"name","label":"Name","w":150},{"k":"fit","label":"Fit","w":150,"type":"dots","max":4,"ink":"emphasis"},{"k":"heat","label":"Heat","w":100,"type":"heat","scale":"risk"}],"rows":[{"name":{"text":"Primary","sub":"Secondary"},"fit":{"value":2,"max":4,"ink":"primary","text":["Detail",{"text":"Nested","children":["Child"]}]},"heat":{"value":3,"text":"High","scale":"risk"},"ink":"primary","scale":"seq"}]}
	]}`)
	obj, err := libraryObject(raw)
	if err != nil {
		t.Fatal(err)
	}
	def := LibraryTemplate{Key: "test/intake", SourceRevision: LibraryRevisionV3, RawSlide: raw}
	for i, n := range obj["body"].([]any) {
		libraryContentWalk(&def, n, "/body/"+string(rune('0'+i)), "node"+string(rune('0'+i)), "", libraryProjectionContext{})
	}
	slots := map[string]bool{}
	for _, s := range def.Slots {
		slots[s.SourcePointer] = true
	}
	for _, p := range []string{"/body/0/sets/0/label", "/body/0/sets/0/bullets/0", "/body/0/points/0/n", "/body/1/active", "/body/1/stages/0/n", "/body/1/branch/n", "/body/2/rows/0/name/text", "/body/2/rows/0/name/sub", "/body/2/rows/0/fit/value", "/body/2/rows/0/fit/text/0", "/body/2/rows/0/fit/text/1/children/0", "/body/2/rows/0/heat/value", "/body/2/rows/0/heat/text"} {
		if !slots[p] {
			t.Errorf("missing content slot %s", p)
		}
	}
	for p := range slots {
		if strings.HasPrefix(p, "/body/1/at/") {
			t.Errorf("maturity stage geometry exposed as content: %s", p)
		}
		for _, fixed := range []string{"/x", "/y", "/w", "/h", "/side", "/labelStyle", "/ink", "/scale", "/max", "/in/0", "/from"} {
			if strings.HasSuffix(p, fixed) {
				t.Errorf("source policy or geometry exposed as content: %s", p)
			}
		}
	}
	arrays := map[string]bool{}
	for _, a := range def.Arrays {
		if a.SourcePointer == "/body/1/at" {
			t.Error("maturity geometry received content identity keys")
		}
		arrays[a.SourcePointer] = true
	}
	for _, p := range []string{"/body/0/sets", "/body/0/sets/0/bullets", "/body/0/regions", "/body/0/points", "/body/1/stages", "/body/2/rows", "/body/2/rows/0/fit/text", "/body/2/rows/0/fit/text/1/children"} {
		if !arrays[p] {
			t.Errorf("missing stable array %s", p)
		}
	}
	d := libraryDiscovery(def, obj)
	for _, kind := range []string{"venn", "maturity", "heatmap", "table"} {
		if !discoveryHas(d.Structures, kind) {
			t.Errorf("missing structural affordance %s", kind)
		}
	}
	for _, group := range []string{"sets", "regions", "stages"} {
		found := false
		for _, g := range d.Groups {
			if strings.HasSuffix(g.SourcePointer, "/"+group) {
				found = true
			}
		}
		if !found {
			t.Errorf("missing semantic group %s", group)
		}
	}
}

func TestIntakeLocalDiagramsAndStraightArrow(t *testing.T) {
	if testing.Short() {
		t.Skip("renderer integration requires registered private branding assets; run make test-integration")
	}
	r := intakeTestRenderer(t)
	for kind, args := range map[string]map[string]any{
		"venn":     {"sets": []any{map[string]any{"label": "A"}, map[string]any{"label": "B"}}},
		"maturity": {"stages": []any{map[string]any{"label": "First"}, map[string]any{"label": "Second"}}},
	} {
		raw, err := ComposeSceneNode(kind, args, Rect{100, 120, 550, 320})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = r.planSceneNode(kind, raw, SceneContext{Surface: "light"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, flip := range []bool{false, true} {
		raw, _ := json.Marshal(map[string]any{"type": "mark", "x": 100, "y": 150, "w": 140, "h": 40, "mark": "arrow-straight", "rotate": 335, "flipX": flip})
		p, err := r.planSceneNode("arrow", raw, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		im := p.Items[0].Image
		if im == nil || im.Rotate != 335 || im.FlipH == nil || *im.FlipH != flip {
			t.Fatal("straight arrow transforms not preserved")
		}
		want := diagramRotatedRect(Rect{100, 150, 140, 40}, 335)
		if math.Abs(p.Bounds.W-want.W) > 1e-8 || math.Abs(p.Bounds.H-want.H) > 1e-8 {
			t.Fatalf("rotated bounds missing: %+v", p.Bounds)
		}
	}
	if _, err := r.planSceneNode("arrow", json.RawMessage(`{"type":"mark","x":100,"y":150,"w":140,"h":40,"mark":"arrow-straight","rotate":1e200}`), SceneContext{Surface: "light"}); err == nil {
		t.Fatal("unbounded arrow rotation accepted")
	}
	p, err := r.planSceneNode("arrow", json.RawMessage(`{"type":"mark","x":100,"y":150,"w":140,"mark":"arrow-straight","rotate":335}`), SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	im := p.Items[0].Image
	want := diagramRotatedRect(Rect{im.X.Val * 72, im.Y.Val * 72, im.W.Val * 72, im.H.Val * 72}, 335)
	if im.H.Val <= 0 || math.Abs(p.Bounds.H-want.H) > 1e-8 {
		t.Fatal("intrinsic arrow height omitted from rotated bounds")
	}
	for _, test := range []struct{ angle, want float64 }{{1415, 335}, {-1105, -25}, {360000, 0}} {
		got, err := sceneMediaRotation(test.angle)
		if err != nil || got != test.want {
			t.Fatalf("multi-turn rotation%v =>%v: %v", test.angle, got, err)
		}
	}
}

func TestIntakeBareDiagramArrayIdentities(t *testing.T) {
	raw := json.RawMessage(`{"type":"venn","x":100,"y":150,"w":450,"h":300,"sets":[{"label":"A"},{"label":"B"}],"regions":[{"in":[0,1],"label":"Shared"}],"points":[{"x":0.5,"y":0.5}]}`)
	obj, err := libraryObject(raw)
	if err != nil {
		t.Fatal(err)
	}
	def := LibraryTemplate{SourceRevision: LibraryRevisionV3}
	libraryContentWalk(&def, obj, "/body/0", "node01", "", libraryProjectionContext{})
	keys := map[string][]string{}
	for _, a := range def.Arrays {
		for i := 0; i < a.Count; i++ {
			keys[a.SourcePointer] = append(keys[a.SourcePointer], "item"+string(rune('a'+i)))
		}
	}
	for _, path := range []string{"/body/0/sets", "/body/0/regions", "/body/0/points"} {
		if len(keys[path]) == 0 {
			t.Fatalf("unlabeled diagram identity missing: %s", path)
		}
	}
	r := intakeTestRenderer(t)
	if _, err = r.planSceneNode("venn", raw, SceneContext{Surface: "light", Path: "/body/0", Keys: keys}); err != nil {
		t.Fatal(err)
	}
}
