package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPublishedBundleDefaultsAndClosedMetadata(t *testing.T) {
	root := t.TempDir()
	if got, err := publishedBundle(root); err != nil || got != "v5" {
		t.Fatalf("latest default: %q %v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, "release"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"v1", "v2", "v3", "v4", "v5", "v6", "../v5", "v5 v3", ""} {
		if err := os.WriteFile(filepath.Join(root, "release/default-bundle.txt"), []byte(value+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := publishedBundle(root)
		if value == "v5" {
			if err != nil || got != value {
				t.Errorf("metadata %q: %q %v", value, got, err)
			}
		} else if err == nil {
			t.Errorf("invalid metadata %q accepted", value)
		}
	}
}

func TestDesignArgsUseStagedV5ButRespectProjectPins(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "release"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "release/default-bundle.txt"), []byte("v5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"library-catalog", "library-search", "library-index", "library-find", "library-inspect", "library-preview", "library-fit", "build"} {
		args, err := designArgs(root, []string{command})
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(root, "library", "wm-design-system", "v5")
		found := false
		for i := range args {
			if args[i] == "--bundle" && i+1 < len(args) && args[i+1] == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s omitted staged published bundle: %v", command, args)
		}
	}
	for _, operation := range []string{"init", "check", "build", "export", "resume"} {
		input := []string{"project", operation, "--project", "maintained-project"}
		got, err := designArgs(root, input)
		if err != nil || !reflect.DeepEqual(got, input) {
			t.Fatalf("wrapper overrode project pin: %v %v", got, err)
		}
	}
	for _, flag := range [][]string{{"--bundle", "v3"}, {"--bundle=v4"}} {
		input := append([]string{"library-catalog"}, flag...)
		got, err := designArgs(root, input)
		if err != nil || !reflect.DeepEqual(got, append(input, "--engine", "wmds-go-foundation.v2")) {
			t.Fatalf("explicit bundle changed: %v %v", got, err)
		}
	}
}

func TestDiscoveryDefaultsUseOnlyLatestLibraryResources(t *testing.T) {
	root := t.TempDir()
	for _, command := range []string{"library-find", "library-inspect", "library-preview"} {
		got, err := designArgs(root, []string{command})
		if err != nil {
			t.Fatal(err)
		}
		values := map[string]string{}
		for i := 1; i+1 < len(got); i += 2 {
			values[got[i]] = got[i+1]
		}
		bundle := filepath.Join(root, "library", "wm-design-system", "v5")
		if values["--bundle"] != bundle || values["--index"] != filepath.Join(bundle, "library.sqlite") || values["--gallery"] != filepath.Join(bundle, "catalog") {
			t.Fatalf("latest discovery paths: %v", got)
		}
		if hasFlag(got, "--legacy-index") || hasFlag(got, "--legacy-root") {
			t.Fatalf("legacy resources injected: %v", got)
		}
	}
	input := []string{"library-inspect", "--index", "custom.sqlite", "--gallery", "custom-gallery", "--bundle", "custom-bundle"}
	got, err := designArgs(root, input)
	if err != nil || !reflect.DeepEqual(got, input) {
		t.Fatalf("explicit discovery paths changed: %v %v", got, err)
	}
}

func TestNativeAndEngineeringCommandsPassThroughWithoutDefaults(t *testing.T) {
	commands := [][]string{
		{"render-doctor", "--input", "source.pptx", "--out", "doctor"},
		{"asset-gallery", "--out", "gallery", "--kind", "photo"},
		{"source-inventory", "--in", "source.pptx", "--out", "inventory"},
		{"render-native-worker", "--worker", "--request", "request.json"},
	}
	for _, input := range commands {
		got, err := designArgs("/missing-release", input)
		if err != nil || !reflect.DeepEqual(got, input) {
			t.Errorf("%s received wrapper defaults: %v (%v)", input[0], got, err)
		}
	}
}
