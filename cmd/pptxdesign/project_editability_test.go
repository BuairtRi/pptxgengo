package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func TestProjectEditabilityCLI(t *testing.T) {
	root := projectDefaultFixture(t)
	bundle, e := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if e != nil {
		t.Fatal(e)
	}
	projectCommandJSON(t, "init", "--project", root, "--bundle", bundle)
	if e := runProject([]string{"editability", "--project", root}); e == nil {
		t.Fatal("unbuilt project accepted")
	}
	projectCommandJSON(t, "build", "--project", root, "--bundle", bundle)
	var report deckproject.NativeEditabilityReport
	if e := json.Unmarshal(projectCommandJSON(t, "editability", "--project", root), &report); e != nil {
		t.Fatal(e)
	}
	if report.Schema != deckproject.NativeEditabilitySchema || report.Counts["objects"] == 0 || report.DesktopQualification != "not_recorded" {
		t.Fatal("invalid inventory", report)
	}
	if e := runProject([]string{"editability", "--project", root, "extra"}); e == nil {
		t.Fatal("positionals accepted")
	}
}
