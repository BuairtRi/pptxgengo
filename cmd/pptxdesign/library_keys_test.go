package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestParseLibraryTemplateKeyFile(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
		want []string
	}{
		{name: "newline", data: "cards/3\n\n  frames/title\n", want: []string{"cards/3", "frames/title"}},
		{name: "csv", data: "cards/3, frames/title\ncomponents/card", want: []string{"cards/3", " frames/title", "components/card"}},
		{name: "json", data: `["cards/3", "frames/title"]`, want: []string{"cards/3", "frames/title"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseLibraryTemplateKeyFile([]byte(tc.data))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("keys = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestParseLibraryTemplateKeyFileEmptyAndMalformed(t *testing.T) {
	for _, data := range []string{"", " \n\t", "[]", "[\"cards/3\""} {
		if _, err := parseLibraryTemplateKeyFile([]byte(data)); err == nil {
			t.Errorf("parseLibraryTemplateKeyFile(%q) unexpectedly succeeded", data)
		}
	}
}

func TestLibraryTemplateSelectionValidatesFileAndKeys(t *testing.T) {
	catalog := []wmdesign.LibraryTemplate{{Key: "cards/3"}, {Key: "frames/title"}}
	file := filepath.Join(t.TempDir(), "template-keys.txt")
	if err := os.WriteFile(file, []byte("cards/3\nframes/title\n"), 0600); err != nil {
		t.Fatal(err)
	}
	selected, seen, err := libraryTemplateSelection(catalog, "", "", file)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(selected, []string{"cards/3", "frames/title"}) || !seen["cards/3"] || !seen["frames/title"] {
		t.Fatalf("selection = %#v, seen = %#v", selected, seen)
	}

	tests := []struct {
		name, family, inline, file, contains string
	}{
		{name: "inline and file", inline: "cards/3", file: file, contains: "choose either"},
		{name: "family and file", family: "cards", file: file, contains: "choose --family"},
		{name: "duplicate", file: writeKeyFile(t, "cards/3,cards/3"), contains: "duplicate"},
		{name: "unknown", file: writeKeyFile(t, "cards/nope"), contains: "unknown template key"},
		{name: "missing file", file: filepath.Join(t.TempDir(), "missing.txt"), contains: "read --template-keys-file"},
		{name: "empty file", file: writeKeyFile(t, " \n"), contains: "file is empty"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := libraryTemplateSelection(catalog, tc.family, tc.inline, tc.file)
			if err == nil || !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("error = %v; want text %q", err, tc.contains)
			}
		})
	}
}

func TestLibraryCatalogTemplateKeysFileKeepsJSONOutput(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
	keysFile := writeKeyFile(t, "cards/3\nquote/light\n")
	output, err := os.CreateTemp(t.TempDir(), "catalog-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	previousArgs, previousStdout := os.Args, os.Stdout
	os.Args = []string{"pptxdesign", "library-catalog", "--bundle", "v10", "--template-keys-file", keysFile}
	os.Stdout = output
	defer func() { os.Args, os.Stdout = previousArgs, previousStdout }()
	if err := run(); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	var rows []wmdesign.LibraryTemplate
	if err := json.NewDecoder(output).Decode(&rows); err != nil {
		t.Fatalf("catalog stdout was not JSON: %v", err)
	}
	if len(rows) != 2 || rows[0].Key != "cards/3" || rows[1].Key != "quote/light" {
		t.Fatalf("catalog keys = %#v", rows)
	}
}

func writeKeyFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "template-keys.txt")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
