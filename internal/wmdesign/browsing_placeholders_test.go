package wmdesign

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

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
