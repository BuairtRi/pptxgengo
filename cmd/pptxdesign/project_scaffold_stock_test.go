package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProjectScaffoldStockIsSharedReadableAndExclusive(t *testing.T) {
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "slides", "slide.yaml")
	args := []string{"--stock", "--bundle", bundle, "--template", "cards/3", "--id", "three-actions", "--out", out}
	if err := runProjectScaffold(args); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(raw), "\n") < 10 || !strings.Contains(string(raw), "\ncontent:\n") {
		t.Fatal("stock source must use readable block YAML")
	}
	var slide map[string]any
	if err := yaml.Unmarshal(raw, &slide); err != nil {
		t.Fatal(err)
	}
	ref := slide["template"].(map[string]any)
	if ref["scope"] != "shared" || ref["id"] != "cards/3" || slide["id"] != "three-actions" || slide["content_kind"] != "synthetic_example" {
		t.Fatalf("invalid stock slide: %#v", slide)
	}
	content := slide["content"].(map[string]any)
	if content["headline"] == nil || content["cards"] == nil || slide["local_templates"] != nil {
		t.Fatalf("missing editable copy: %#v", content)
	}
	if strings.Index(string(raw), "content:") > strings.Index(string(raw), "bindings:") {
		t.Fatal("copy must precede technical bindings")
	}
	if err := runProjectScaffold(args); err == nil {
		t.Fatal("overwrote existing slide source")
	}
	if err := runProjectScaffold([]string{"--stock", "--template", "cards/3", "--id", "three-actions", "--reason", "adapt"}); err == nil {
		t.Fatal("stock accepted adaptation flags")
	}
}
