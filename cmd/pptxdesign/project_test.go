package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectFlagContracts(t *testing.T) {
	for _, args := range [][]string{{}, {"bad"}, {"status", "--out", "unused.zip"}, {"build", "--actor", "reviewer"}, {"approve", "--mode", "client"}, {"check", "--as", "variant"}, {"check", "unexpected"}} {
		if e := runProject(args); e == nil {
			t.Errorf("accepted invalid command %v", args)
		}
	}
}
func TestProjectCLIWorkflow(t *testing.T) {
	src := "../../examples/deck-project"
	root := t.TempDir()
	if e := filepath.WalkDir(src, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(src, path)
		if e != nil {
			return e
		}
		out := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0755)
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(out, b, 0644)
	}); e != nil {
		t.Fatal(e)
	}
	bundle, e := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if e != nil {
		t.Fatal(e)
	}
	commands := [][]string{{"init"}, {"check"}, {"build"}, {"status"}, {"resume"}, {"approve", "--stage", "content", "--actor", "maintainer"}, {"review", "--out", filepath.Join(t.TempDir(), "review.zip")}, {"export", "--mode", "client", "--out", filepath.Join(t.TempDir(), "client.zip")}}
	for _, cmd := range commands {
		args := append(cmd, "--project", root, "--bundle", bundle)
		r, w, e := os.Pipe()
		if e != nil {
			t.Fatal(e)
		}
		old := os.Stdout
		os.Stdout = w
		var output bytes.Buffer
		done := make(chan error, 1)
		go func() { _, e := io.Copy(&output, r); done <- e }()
		err := runProject(args)
		w.Close()
		os.Stdout = old
		if e := <-done; e != nil {
			t.Fatal(e)
		}
		r.Close()
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !json.Valid(output.Bytes()) {
			t.Fatalf("%s returned invalid JSON", cmd[0])
		}
	}
}
