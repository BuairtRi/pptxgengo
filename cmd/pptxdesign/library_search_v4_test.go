package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestLibrarySearchV5Shorthand(t *testing.T) {
	testLibrarySearchPinnedShorthand(t, "v5", wmdesign.LibraryRevisionV5)
}

func TestLibrarySearchLatestDefault(t *testing.T) {
	testLibrarySearchPinnedShorthand(t, "", wmdesign.LibraryRevisionV5)
}

func testLibrarySearchPinnedShorthand(t *testing.T, shorthand, revision string) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
	output, err := os.CreateTemp(t.TempDir(), "search-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	previous := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = previous }()
	args := []string{"--query", "maturity", "--limit", "1"}
	if shorthand != "" {
		args = append(args, "--bundle", shorthand)
	}
	if err := runLibrarySearch(args); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	var result wmdesign.LibrarySearchResult
	if err := json.NewDecoder(output).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Matches) != 1 || result.Matches[0].Template.SourceRevision != revision {
		t.Fatalf("%s shorthand did not select %s catalog: %+v", shorthand, revision, result.Matches)
	}
}
