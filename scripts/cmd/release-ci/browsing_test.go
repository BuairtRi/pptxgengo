package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/browsingartifact"
	"github.com/buairtri/pptxgengo/internal/browsingfixture"
)

func TestBrowsingArchiveRequiresBothDecksAndHashes(t *testing.T) {
	t.Setenv("CI_PIPELINE_CREATED_AT", "")
	root := t.TempDir()
	files := map[string]Input{}
	if e := addBrowsingFiles(files, root); e == nil {
		t.Fatal("missing decks accepted")
	}
	os.Mkdir(filepath.Join(root, "browsing"), 0755)
	hashes := map[string]string{}
	names := []string{"template-library.pptx", "template-library.manifest.json", "reusable-slides.pptx", "reusable-slides.manifest.json"}
	for _, name := range names {
		path := filepath.Join(root, "browsing", name)
		if e := os.WriteFile(path, browsingfixture.Deck(t, func() string {
			if strings.HasPrefix(name, "template-") {
				return "templates"
			}
			return "reusable"
		}()), 0644); e != nil {
			t.Fatal(e)
		}
		hash, e := digest(path)
		if e != nil {
			t.Fatal(e)
		}
		hashes["browsing/"+name] = hash
	}
	for _, pair := range []struct{ manifest, deck, kind string }{{"template-library.manifest.json", "template-library.pptx", "templates"}, {"reusable-slides.manifest.json", "reusable-slides.pptx", "reusable"}} {
		path := filepath.Join(root, "browsing", pair.manifest)
		data := browsingfixture.Manifest(t, pair.kind, hashes["browsing/"+pair.deck])
		if e := os.WriteFile(path, data, 0644); e != nil {
			t.Fatal(e)
		}
		hashes["browsing/"+pair.manifest], _ = digest(path)
	}
	raw := browsingfixture.Inventory(t, hashes)
	os.WriteFile(filepath.Join(root, "browsing-manifest.json"), raw, 0644)
	if e := addBrowsingFiles(files, root); e != nil || len(files) != 5 {
		t.Fatal(files, e)
	}
	os.WriteFile(filepath.Join(root, "browsing", names[0]), []byte("tampered"), 0644)
	if e := addBrowsingFiles(map[string]Input{}, root); e == nil {
		t.Fatal("tampered deck accepted")
	}
}

func TestBrowsingArchiveRejectsLinkedDirectory(t *testing.T) {
	root := t.TempDir()
	if e := os.Symlink(t.TempDir(), filepath.Join(root, "browsing")); e != nil {
		t.Fatal(e)
	}
	if e := addBrowsingFiles(map[string]Input{}, root); e == nil || !strings.Contains(e.Error(), "real directories") {
		t.Fatal("linked browsing directory accepted", e)
	}
}

// Set PPTXGENGO_TEST_TEMPLATE_RESOURCE_ROOT to exercise the same staging
// directory produced by prepare-resources.sh through release-ci's archive
// closure. This remains opt-in so ordinary unit tests stay hermetic.
func TestStagedTemplateOnlyResourcesArchiveClosure(t *testing.T) {
	root := os.Getenv("PPTXGENGO_TEST_TEMPLATE_RESOURCE_ROOT")
	version, commit, pipeline := os.Getenv("CI_COMMIT_TAG"), os.Getenv("CI_COMMIT_SHA"), os.Getenv("CI_PIPELINE_CREATED_AT")
	if root == "" || version == "" || commit == "" || pipeline == "" {
		t.Skip("staged template resource integration not requested")
	}
	t.Setenv("PPTXGENGO_PACKAGE_KIND", "cli-only")
	t.Setenv("PPTXGENGO_BROWSING_POLICY", "templates-only")
	t.Setenv("CI_PIPELINE_CREATED_AT", pipeline)
	files := map[string]Input{}
	if err := addScopedBrowsingFiles(files, root, version, commit); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"browsing/template-library.pptx", "browsing/template-library.manifest.json", "browsing/native-editing-coverage.json", "library/wm-design-system/v11/library.sqlite", "library/wm-design-system/v11/bundle.json", "library/wm-design-system/v11/fonts/IBMPlexSans-Regular.ttf", browsingartifact.TemplateCatalogInventoryName, "skills/west-monroe-presentations/SKILL.md", "scripts/install-skill.py", "SKILL-INSTALL.md", "library/wm-design-system/v11/catalog/assets/assets.json"} {
		if _, ok := files[required]; !ok {
			t.Fatalf("staged archive closure omitted %s", required)
		}
	}
	for name := range files {
		if strings.Contains(name, "reusable-slides") || strings.Contains(name, "/photos/") || strings.Contains(name, "/branding/") {
			t.Fatalf("deferred resource entered templates-only archive: %s", name)
		}
	}
}
