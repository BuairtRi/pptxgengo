package wmdesign

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestIntakeGeographyFrozenDataAndProjection(t *testing.T) {
	if fmt.Sprintf("%x", sha256.Sum256([]byte(intakeGeoDataJSON))) != intakeGeoDataSHA256 {
		t.Fatal("frozen geo hash drift")
	}
	data, e := intakeGeography()
	if e != nil {
		t.Fatal(e)
	}
	if len(data.Geo["US"]) != 1 || len(data.Geo["CA"]) != 10 || len(data.Geo["GB"]) != 14 || len(data.Lakes) != 5 {
		t.Fatal("missing exact source geographic rings")
	}
	panels, e := intakeMapPanels(intakeMapSource{W: 558, Region: "americas-uk"})
	if e != nil {
		t.Fatal(e)
	}
	x, y, e := intakeMapLocation(panels, -.13, 51.51)
	if e != nil {
		t.Fatal(e)
	}
	expectedX := 558*.74 + 18 + (-.13+6.4)*(558*.26-18)/8.8
	expectedY := (59 - 51.51) * (558*.26 - 18) / 8.8 / .6
	if math.Abs(x-expectedX) > 1e-8 || math.Abs(y-expectedY) > 1e-8 {
		t.Fatalf("wrong GB inset projection %.4f %.4f", x, y)
	}
	if _, _, e = intakeMapLocation(panels, 140, 35); e == nil {
		t.Fatal("outside location silently uses US")
	}
	if intakeMapLand(panels[0], -87, 44) {
		t.Fatal("Great Lakes hole filled")
	}
}
func TestIntakeGeographyNativeMarkersAndSVG(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"dotmap","x":345,"y":54,"w":558,"region":"americas-uk","points":[{"id":"chicago","name":"Chicago","lon":-87.63,"lat":41.88,"kind":"office","labelPos":"t"},{"id":"london","name":"London","lon":-0.13,"lat":51.51,"kind":"hub","labelPos":"b"}]}`)
	p, ok, e := r.planIntakeGeographyScene("map", raw, SceneContext{Surface: "light", Path: "/body/0"})
	if e != nil || !ok {
		t.Fatal(e)
	}
	images, shapes, texts := 0, 0, 0
	for _, item := range p.Items {
		if item.Image != nil {
			images++
			encoded := strings.SplitN(item.Image.Data, ",", 2)
			if len(encoded) != 2 {
				t.Fatal("bad SVG data")
			}
			svg, e := base64.StdEncoding.DecodeString(encoded[1])
			if e != nil {
				t.Fatal(e)
			}
			var root struct{ XMLName xml.Name }
			if e = xml.Unmarshal(svg, &root); e != nil || root.XMLName.Local != "svg" {
				t.Fatal("invalid SVG")
			}
			if item.Image.SVGFallbackData == "" {
				t.Fatal("missing PNG fallback")
			}
		}
		if item.Shape != nil {
			shapes++
		}
		if item.Text != nil {
			texts++
		}
	}
	if images != 1 || shapes != 2 || texts != 2 || len(p.Groups) != 1 || p.Groups[0].Contract != IntakeGeographyContract {
		t.Fatalf("unexpected editable map composition %d/%d/%d", images, shapes, texts)
	}
}
func TestIntakeGeographyStrictValidation(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, extra := range []string{
		`"region":"world"`, `"pitch":0.01`, `"points":[{"name":"X","lon":140,"lat":35}]`, `"points":[{"name":"X","lat":51}]`, `"points":[{"name":"X","lon":-87,"lat":41,"labelPos":"diagonal"}]`, `"points":[{"name":"X","lon":-87,"lat":41,"unexpected":true}]`,
	} {
		raw := json.RawMessage(`{"type":"dotmap","x":0,"y":0,"w":558,` + extra + `}`)
		_, ok, e := r.planIntakeGeographyScene("map", raw, SceneContext{})
		if !ok || e == nil {
			t.Fatalf("accepted invalid %s", raw)
		}
	}
	raw := json.RawMessage(`{"type":"dotmap","x":0,"y":0,"w":558,"points":[{"name":"X","lon":-87,"lat":41}]}`)
	_, _, e := r.planIntakeGeographyScene("map", raw, SceneContext{Keys: map[string][]string{}})
	if e == nil {
		t.Fatal("bound map lacks stable keys")
	}
}

func TestIntakeGeographyMatchesFrozenRenderer(t *testing.T) {
	data, e := os.ReadFile(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle", "source", "explorations", "components.src.html"))
	if e != nil {
		t.Fatal(e)
	}
	source := string(data)
	extract := func(marker string, out any) {
		t.Helper()
		i := strings.Index(source, marker)
		if i < 0 {
			t.Fatalf("frozen renderer lacks %s", marker)
		}
		rest := source[i+len(marker):]
		end := strings.Index(rest, ";")
		if end < 0 {
			t.Fatal("unterminated geographic data")
		}
		if e := json.Unmarshal([]byte(rest[:end]), out); e != nil {
			t.Fatal(e)
		}
	}
	var geo map[string][][][2]float64
	var lakes [][][2]float64
	extract("var GEO = ", &geo)
	extract("var LAKES = ", &lakes)
	frozen, e := intakeGeography()
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(frozen.Geo, geo) || !reflect.DeepEqual(frozen.Lakes, lakes) {
		t.Fatal("compiled geography does not match the pinned v5 source renderer")
	}
}

func TestIntakeGeographyV3LocationRefinementsFitSplit(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewTypographyEngine(filepath.Join(bundle, "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	r := &renderer{source: source, typeEngine: engine, bundle: bundle}
	catalog, err := libraryCatalog(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"about/locations", "about/locations-international"} {
		entry, err := libraryTemplate(catalog, key)
		if err != nil {
			t.Fatal(err)
		}
		raw := entry.RawSlide
		original := append([]byte(nil), raw...)
		doc, err := compileLibrarySlide(raw, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err = applyLibraryRefinements(key, source.Revision, &doc); err != nil {
			t.Fatal(err)
		}
		f, err := source.ResolveFrame(doc.Frame)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, n := range doc.Nodes {
			var tag struct{ Type string }
			if err = json.Unmarshal(n.Scene.Node, &tag); err != nil {
				t.Fatal(err)
			}
			if tag.Type != "dotmap" {
				continue
			}
			found = true
			p, _, err := r.planIntakeGeographyScene(n.ID, n.Scene.Node, SceneContext{Surface: "light", Path: n.Scene.Path, Zone: f.TallBody})
			if err != nil {
				t.Fatal(err)
			}
			if !inside(p.Bounds, f.TallBody) {
				t.Fatalf("%s still exceeds split: %+v vs %+v", key, p.Bounds, f.TallBody)
			}
			if err = sceneTextEnvelope(p, SceneContext{Zone: f.TallBody}); err != nil {
				t.Fatal(err)
			}
			if key == "about/locations" && p.Bounds.Y+p.Bounds.H > 280 {
				t.Fatalf("US map overlaps legend starting at288: %+v", p.Bounds)
			}
			if len(n.Scene.Resolutions) == 0 {
				t.Fatal("unrecorded geometry amendment")
			}
		}
		if !found || string(original) != string(entry.RawSlide) {
			t.Fatal("map absent or frozen source modified")
		}
	}
}
