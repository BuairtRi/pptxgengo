package main

import (
	"github.com/buairtri/pptxgengo/internal/browsingfixture"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
