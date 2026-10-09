package main

import (
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/deckproject"
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldSourceContainerClearanceFitCLI(t *testing.T) {
	bundle, err := filepath.Abs("../../library/wm-design-system/v11")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "scaffold.json")
	args := []string{"--bundle", bundle, "--template", "architecture/layer-map", "--reason", "Explicit height adaptation for source caption clearance", "--placeholder-media", "--out", out}
	if err := runProjectScaffold(args); err == nil {
		t.Fatal("original narrow caption clearance silently fitted")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("refused scaffold wrote output", err)
	}
	args = append(args, "--source-container-clearance-fit")
	if err := runProjectScaffold(args); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var receipt deckproject.TemplateScaffold
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.GeometryAdjustments) != 1 || receipt.GeometryAdjustments[0].NodeID != "node16" || receipt.GeometryAdjustments[0].Original.H != 294 || receipt.GeometryAdjustments[0].Authored.H != 291 {
		t.Fatal(receipt.GeometryAdjustments)
	}
	if receipt.Template.Provenance.Parent.ID != "architecture/layer-map" || receipt.Template.Provenance.SourceFileSHA256 == "" {
		t.Fatal("source ancestry lost")
	}
	if err := runProjectScaffold(args); err == nil {
		t.Fatal("existing preview overwritten")
	}
	after, err := os.ReadFile(out)
	if err != nil || string(after) != string(raw) {
		t.Fatal("failed output changed receipt", err)
	}
}
