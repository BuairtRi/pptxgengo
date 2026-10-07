package wmdesign

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// A deliberately bounded corpus copied from real pinned entities. Open verifies
// the full source pins first; the subset keeps complete rows, FTS text and its
// own counts/projection hashes. This is test-only, not a production index mode.
func indexFixtureSubset(t *testing.T, perKind int) (string, LibraryIndexReport) {
	t.Helper()
	path, report := indexFixture(t)
	full, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	entities, err := full.entities("1=1", nil)
	closeErr := full.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	subset := []LibraryEntity{}
	counts := map[string]int{}
	ids := []any{}
	placeholders := []string{}
	for _, e := range entities {
		if counts[e.Kind] >= perKind && e.ID != "wmds/template/cards/3" && e.ID != "wmds/template/cards/4" {
			continue
		}
		subset = append(subset, e)
		counts[e.Kind]++
		ids = append(ids, e.ID)
		placeholders = append(placeholders, "?")
	}
	report.Counts = counts
	report.ProjectionSHA256 = projectionHash(subset)
	report.RetrievalSHA256 = retrievalProjectionHash(subset)
	report.Warnings = append(report.Warnings, fmt.Sprintf("Test-only bounded corpus: %d original pinned entities; not a complete catalog qualification.", len(subset)))
	db, err := indexDB(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	keep := strings.Join(placeholders, ",")
	for _, table := range []string{"entities", "entity_fts", "dimensions", "content_groups", "dependencies"} {
		column := "entity_id"
		if table == "entities" || table == "entity_fts" {
			column = "id"
		}
		if _, err = tx.Exec("DELETE FROM "+table+" WHERE "+column+" NOT IN ("+keep+")", ids...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = tx.Exec("UPDATE meta SET value=? WHERE key='report'", string(indexJSON(report))); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return path, report
}
