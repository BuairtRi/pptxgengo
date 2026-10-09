package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"math"
	"os"
	"testing"
)

func TestSchematicUnderscoreRegisteredEnvelope(t *testing.T) {
	data, err := os.ReadFile("../../library/wm-design-system/v11/catalog/assets/assets/760-01.svg")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != primitiveAssetRegistry["underscore"].SHA256 {
		t.Fatal("registered underscore source drift")
	}
	paths, err := primitiveSVGPaths(data, true)
	if err != nil {
		t.Fatal(err)
	}
	box := primitivePathBounds(paths)
	if math.Abs(box[2]-schematicUnderscoreWidth) > 1e-9 || math.Abs(box[3]-schematicUnderscoreHeight) > 1e-9 {
		t.Fatal("schematic underscore envelope differs from pinned registered mark dimensions", box)
	}
	assets, err := TemplatePlaceholderAssets()
	if err != nil {
		t.Fatal(err)
	}
	paths, err = primitiveSVGPaths(assets["underscore"].Data, true)
	if err != nil {
		t.Fatal(err)
	}
	synthetic := primitivePathBounds(paths)
	if synthetic[2] != schematicUnderscoreWidth || synthetic[3] != schematicUnderscoreHeight {
		t.Fatal("schematic mark inflated", synthetic)
	}
	if 138*(synthetic[3]+.8)/(synthetic[2]+.8) > 13 {
		t.Fatal("title underline became a large rectangle")
	}
	for _, id := range []string{"highlight-1", "highlight-2", "highlight-3", "highlight-4"} {
		cfg, _, err := image.DecodeConfig(bytes.NewReader(assets[id].Data))
		if err != nil || cfg.Width != schematicHighlightWidth || cfg.Height != schematicHighlightHeight {
			t.Fatal("schematic highlight aspect drift", id, cfg, err)
		}
	}
}

func TestBrowsingPlaceholderAssetsAreExplicitDeterministicOverrides(t *testing.T) {
	a, err := TemplatePlaceholderAssets()
	if err != nil {
		t.Fatal(err)
	}
	b, err := TemplatePlaceholderAssets()
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(primitiveAssetRegistry) {
		t.Fatal("incomplete media placeholder coverage")
	}
	for key, asset := range a {
		if asset.SHA256 != fmt.Sprintf("%x", sha256.Sum256(asset.Data)) || asset.SHA256 != b[key].SHA256 || asset.SHA256 == primitiveAssetRegistry[key].SHA256 {
			t.Fatal("placeholder has incorrect or original identity", key)
		}
		if asset.MIME == "image/svg+xml" {
			paths, err := primitiveSVGPaths(asset.Data, true)
			if err != nil || len(paths) == 0 {
				t.Fatal("placeholder artwork cannot render", key, err)
			}
		}
	}
}

func TestTemplateNativeCoverageRetainsEveryTemplate(t *testing.T) {
	c := BrowsingCoverage{Entries: []BrowsingEntry{{Kind: "template", Key: "first", SlideID: "a"}, {Kind: "frame", SlideID: "frame"}, {Kind: "template", Key: "second", SlideID: "b"}}}
	r := Report{EditingProfile: NativeEditingProfile, Slides: []SlideReport{{ID: "a", Texts: []TextRecord{{NativeParagraphContract: EditableListContract}, {NativeShape: &NativeTextShape{ParagraphContract: EditableCardContract}}}, Tables: []SceneTableRecord{{}}, Scenes: []SceneRecord{{Warnings: []string{"native-v1 converted source component: example", "other warning"}}}}, {ID: "b"}}}
	out := TemplateNativeCoverage(c, r)
	if out.Templates != 2 || out.NativeLists != 1 || out.NativeCards != 1 || out.NativeTables != 1 || len(out.Entries[0].Decisions) != 1 || out.Entries[1].Template != "second" {
		t.Fatal(out)
	}
}
