package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func portableCommandJSON(t *testing.T, args ...string) []byte {
	t.Helper()
	r, w, e := os.Pipe()
	if e != nil {
		t.Fatal(e)
	}
	previous := os.Stdout
	os.Stdout = w
	var output bytes.Buffer
	done := make(chan error, 1)
	go func() { _, e := io.Copy(&output, r); done <- e }()
	err := runProject(args)
	w.Close()
	os.Stdout = previous
	readErr := <-done
	r.Close()
	if err != nil || readErr != nil {
		t.Fatalf("%v: %v %v", args, err, readErr)
	}
	if !json.Valid(output.Bytes()) {
		t.Fatalf("invalid JSON: %s", output.Bytes())
	}
	return output.Bytes()
}
func TestPortableCLIEndToEnd(t *testing.T) {
	bundle, e := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if e != nil {
		t.Fatal(e)
	}
	root := filepath.Join(t.TempDir(), "owned project")
	portableCommandJSON(t, "create", "--out", root, "--id", "portable-deck", "--title", "Portable generic deck", "--bundle", bundle, "--template", "cards/3")
	portableCommandJSON(t, "layout", "--project", root, "--dry-run")
	portableCommandJSON(t, "check", "--project", root, "--bundle", bundle)
	portableCommandJSON(t, "build", "--project", root, "--bundle", bundle)
	portableCommandJSON(t, "version", "save", "--project", root, "--actor", "Test operator", "--message", "Exact source and generated deck")
	portableCommandJSON(t, "version", "list", "--project", root)
	portableCommandJSON(t, "version", "verify", "--project", root, "--number", "000001")
	archive := filepath.Join(t.TempDir(), "complete.zip")
	portableCommandJSON(t, "share", "--project", root, "--out", archive)
	extracted := filepath.Join(t.TempDir(), "colleague")
	portableCommandJSON(t, "share-extract", "--archive", archive, "--out", extracted)
	portableCommandJSON(t, "share-verify", "--project", extracted)
	restored := filepath.Join(t.TempDir(), "restored")
	portableCommandJSON(t, "version", "materialize", "--project", extracted, "--number", "000001", "--out", restored)
	portableCommandJSON(t, "check", "--project", restored, "--bundle", bundle)
}
func TestPortableCLIFlagContracts(t *testing.T) {
	for _, args := range [][]string{{"layout", "--apply", "--dry-run"}, {"layout", "--actor", "unsupported"}, {"version", "save", "--number", "000001"}, {"version", "list", "--out", "unused"}, {"version", "recover", "--out", "unused"}, {"share-extract", "--project", "."}, {"share-verify", "--archive", "unused"}, {"create", "--out", "unused"}, {"asset", "revise", "--file", "missing"}, {"asset", "add", "--file", "missing", "--expect-sha256", "wrong"}} {
		if e := runProject(args); e == nil {
			t.Errorf("accepted invalid flags: %v", args)
		}
	}
}
