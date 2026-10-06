package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogSelectorsUseSingleCurrentGallery(t *testing.T) {
	root := t.TempDir()
	gallery := filepath.Join(root, "library", "wm-design-system", "v11", "catalog", "design-system.html")
	if err := os.MkdirAll(filepath.Dir(gallery), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gallery, []byte("qualified gallery fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, selector := range []string{"", "--templates", "--design-system"} {
		t.Run(selector, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			previous := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = previous }()
			args := []string{"--print"}
			if selector != "" {
				args = append(args, selector)
			}
			if err := runCatalog(root, args); err != nil {
				t.Fatal(err)
			}
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(output)
			if err != nil || strings.TrimSpace(string(data)) != gallery {
				t.Fatalf("catalog path %q; want %q: %v", data, gallery, err)
			}
		})
	}
	if err := runCatalog(root, []string{"--templates", "--design-system"}); err == nil || !strings.Contains(err.Error(), "choose one") {
		t.Fatalf("conflicting selectors accepted: %v", err)
	}
}

func TestCatalogComponentsGiveActionableDiscovery(t *testing.T) {
	err := runCatalog(t.TempDir(), []string{"--components"})
	if err == nil || !strings.Contains(err.Error(), "pptxgengo design library-find --kinds component") {
		t.Fatalf("component guidance: %v", err)
	}
	err = runCatalog(t.TempDir(), []string{"--unknown"})
	if err == nil || strings.Contains(err.Error(), "--components") || !strings.Contains(err.Error(), "--templates|--design-system") {
		t.Fatalf("catalog usage advertises unavailable gallery: %v", err)
	}
}

func TestCatalogAssetsPrintUsesInstalledAssetGallery(t *testing.T) {
	root := t.TempDir()
	gallery := filepath.Join(root, "library", "wm-design-system", "v11", "catalog", "assets", "index.html")
	if err := os.MkdirAll(filepath.Dir(gallery), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gallery, []byte("asset gallery fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	previous := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = previous }()
	if err := runCatalog(root, []string{"--assets", "--print"}); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(output)
	if err != nil || strings.TrimSpace(string(data)) != gallery {
		t.Fatalf("asset catalog path %q; want %q: %v", data, gallery, err)
	}
	if err := runCatalog(root, []string{"--assets", "--templates"}); err == nil || !strings.Contains(err.Error(), "choose one catalog gallery") {
		t.Fatalf("conflicting asset selector accepted: %v", err)
	}
}
