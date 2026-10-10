package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func TestProjectMigrateCommand(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", repo)
	root := projectDefaultFixture(t)
	projectCommandJSON(t, "init", "--project", root, "--bundle", currentDesignBundle)
	raw := projectCommandJSON(t, "migrate", "--project", root, "--dry-run")
	var r deckproject.Migration
	if err := json.Unmarshal(raw, &r); err != nil || r.Status != "already_current" || r.To != "wmds-library.v12" {
		t.Fatalf("migrate command: %s %v", raw, err)
	}
	if err := runProject([]string{"migrate", "--project", root, "--bundle", ""}); err == nil {
		t.Fatal("empty target accepted")
	}
	unlocked := projectDefaultFixture(t)
	if err := runProject([]string{"migrate", "--project", unlocked}); err == nil || !strings.Contains(err.Error(), "project init") {
		t.Fatalf("unlocked migration: %v", err)
	}
}
