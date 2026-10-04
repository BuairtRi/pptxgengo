package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func TestProjectSplitCLIUsesFileReferencesAndKeepsLock(t *testing.T) {
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", repo)
	root := projectDefaultFixture(t)
	projectCommandJSON(t, "init", "--project", root)
	p, err := deckproject.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), p.Canonical...)
	lockBefore, err := os.ReadFile(filepath.Join(root, "toolchain.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	projectCommandJSON(t, "split", "--project", root, "--bundle", "v5")
	q, err := deckproject.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(q.Canonical, before) || q.SlideFiles["stable-page"] == "" || len(q.TemplateFiles) != 1 {
		t.Fatal("CLI split changed content or omitted referenced files")
	}
	lockAfter, err := os.ReadFile(filepath.Join(root, "toolchain.lock.json"))
	if err != nil || !bytes.Equal(lockBefore, lockAfter) {
		t.Fatal("CLI split replaced lock")
	}
	projectCommandJSON(t, "check", "--project", root)
	projectCommandJSON(t, "build", "--project", root)
}
