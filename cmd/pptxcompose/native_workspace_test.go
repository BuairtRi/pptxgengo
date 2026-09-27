package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestStageNativeDeckCreatesAndReusesExactBytes(t *testing.T) {
	workspace := t.TempDir()
	source := filepath.Join(t.TempDir(), "source.pptx")
	deck := []byte("exact source deck bytes")
	if err := os.WriteFile(source, deck, 0644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(workspace, "source.pptx")
	for i := 0; i < 2; i++ {
		got, err := stageNativeDeck(workspace, source, deck)
		if err != nil || got != want {
			t.Fatalf("stage %d: path %q, error %v", i, got, err)
		}
		if err := checkNativeDeckUnchanged(got, hash(deck)); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(want)
	if err != nil || !bytes.Equal(got, deck) {
		t.Fatalf("staged bytes: %q, error %v", got, err)
	}
}

func TestStageNativeDeckRefusesDifferentExistingBytes(t *testing.T) {
	workspace := t.TempDir()
	dest := filepath.Join(workspace, "source.pptx")
	previous := []byte("keep existing deck")
	if err := os.WriteFile(dest, previous, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := stageNativeDeck(workspace, filepath.Join(t.TempDir(), "source.pptx"), []byte("new deck")); err == nil {
		t.Fatal("expected a collision error")
	}
	got, err := os.ReadFile(dest)
	if err != nil || !bytes.Equal(got, previous) {
		t.Fatalf("existing bytes changed: %q, error %v", got, err)
	}
}

func TestStageNativeDeckRefusesSymlinksAndMissingWorkspace(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	if err := os.Mkdir(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source.pptx")
	if _, err := stageNativeDeck(filepath.Join(root, "missing"), source, []byte("deck")); err == nil {
		t.Fatal("expected missing workspace error")
	}
	if err := os.Symlink(workspace, filepath.Join(root, "workspace-link")); err != nil {
		t.Fatal(err)
	}
	if _, err := stageNativeDeck(filepath.Join(root, "workspace-link"), source, []byte("deck")); err == nil {
		t.Fatal("expected symlink workspace error")
	}
	if err := os.Symlink(source, filepath.Join(workspace, "source.pptx")); err != nil {
		t.Fatal(err)
	}
	if _, err := stageNativeDeck(workspace, source, []byte("deck")); err == nil {
		t.Fatal("expected symlink destination error")
	}
}

func TestCheckNativeDeckUnchangedDetectsMutation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deck.pptx")
	if err := os.WriteFile(path, []byte("before"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkNativeDeckUnchanged(path, hash([]byte("before"))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("after"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkNativeDeckUnchanged(path, hash([]byte("before"))); err == nil {
		t.Fatal("expected a changed-deck error")
	}
}
