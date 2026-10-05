package wmdesign

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryV7PublicationInheritedRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("paired exhaustive rendering of every retained native source specimen")
	}
	root := filepath.Join("..", "..")
	previous := filepath.Join(root, "planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	current := filepath.Join(root, "library/wm-design-system/v7")
	gallery := filepath.Join(current, "catalog/design-system/index.json")
	raw, err := os.ReadFile(gallery)
	if err != nil {
		t.Fatal(err)
	}
	var index publicationIndex
	if err = json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, entry := range index.Designs {
		if entry.Evidence == "inherited_identical_authored_composition" {
			keys = append(keys, entry.Template)
		}
	}
	if len(keys) != 586 {
		t.Fatalf("retained composition count=%d; want586", len(keys))
	}
	count, err := VerifyLibraryPublicationRenderInheritance(LibraryPublicationOptions{Bundle: current, PreviousBundle: previous, Year: 2026}, keys)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(keys) {
		t.Fatalf("qualified%d/%d inherited renders", count, len(keys))
	}
}
