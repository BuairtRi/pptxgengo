package wmdesign

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryV11PublishedGalleryInheritedRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("paired exhaustive rendering of every retained native source specimen")
	}
	root := filepath.Join("..", "..")
	previous := filepath.Join(root, "planning/wm-design-contracts/v10/intake-20261006-649-frozen/bundle")
	current := filepath.Join(root, "library/wm-design-system/v11")
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
	if index.Entries != 649 || len(index.Designs) != index.Entries {
		t.Fatalf("published gallery count=%d/%d; want649", index.Entries, len(index.Designs))
	}
	if len(keys) != 297 {
		t.Fatalf("retained composition count=%d; want297", len(keys))
	}
	if index.Entries-len(keys) != 352 {
		t.Fatalf("newly reviewed specimens=%d; want352", index.Entries-len(keys))
	}
	if index.Qualification["reviewed_source_specimens"] != float64(649) || index.Qualification["arbitrary_content_qualified"] != false {
		t.Fatalf("published specimen qualification=%v", index.Qualification)
	}
	count, err := VerifyLibraryPublicationRenderInheritance(LibraryPublicationOptions{Bundle: current, PreviousBundle: previous, Year: 2026}, keys)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(keys) {
		t.Fatalf("qualified%d/%d inherited renders", count, len(keys))
	}
}
