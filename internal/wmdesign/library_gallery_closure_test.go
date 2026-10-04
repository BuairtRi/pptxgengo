package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFrozenGalleryNativePreviewClosure(t *testing.T) {
	root := filepath.Join("..", "..", "library", "wm-design-system", "v5", "catalog")
	raw, err := os.ReadFile(filepath.Join(root, "design-system", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Entries int `json:"entries"`
		Designs []struct {
			Template   string `json:"template"`
			Preview    string `json:"source_preview"`
			SHA        string `json:"source_preview_sha256"`
			Contract   string `json:"contract"`
			Foundation string `json:"source_foundation"`
			Values     string `json:"source_values"`
		} `json:"designs"`
	}
	if err = json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	if index.Entries != 587 || len(index.Designs) != index.Entries {
		t.Fatalf("incomplete frozen gallery: %d/%d", index.Entries, len(index.Designs))
	}
	for _, design := range index.Designs {
		dependencies := []string{design.Preview, design.Contract, design.Foundation}
		// Deprecated specimens can omit a source-values link. Every declared
		// link must resolve; inventing a link would misstate that contract.
		if design.Values != "" {
			dependencies = append(dependencies, design.Values)
		}
		for _, relative := range dependencies {
			if relative == "" || !filepath.IsLocal(relative) {
				t.Fatalf("invalid gallery dependency for %s: %q", design.Template, relative)
			}
			data, err := os.ReadFile(filepath.Join(root, relative))
			if err != nil {
				t.Fatalf("missing frozen dependency for %s: %v", design.Template, err)
			}
			if relative == design.Preview && fmt.Sprintf("%x", sha256.Sum256(data)) != design.SHA {
				t.Fatalf("native preview drift: %s", design.Template)
			}
		}
	}
}
