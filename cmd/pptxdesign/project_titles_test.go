package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectTitlesReportsCompiledOrderAndMetadata(t *testing.T) {
	project, err := filepath.Abs("../../examples/deck-project")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.CreateTemp(t.TempDir(), "titles-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	previous := os.Stdout
	os.Stdout = file
	err = runProjectTitles([]string{"--project", project, "--bundle", bundle, "--format", "json"})
	os.Stdout = previous
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	var rows []projectTitleRow
	if err = json.NewDecoder(file).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("no slide titles")
	}
	for i, row := range rows {
		if row.Page != i+1 || row.ID == "" || row.Title == "" || row.Template == "" {
			t.Errorf("row %d lacks compiled source metadata: %+v", i, row)
		}
	}
}

func TestProjectTitlesRejectsUnsupportedFormat(t *testing.T) {
	if err := runProjectTitles([]string{"--format", "csv"}); err == nil {
		t.Fatal("accepted unsupported output format")
	}
}
