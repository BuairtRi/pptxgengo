package main

import (
	"strings"
	"testing"
)

func TestLibraryRetrievalFlagsRejectUnsupportedPaths(t *testing.T) {
	for _, args := range [][]string{
		{"--summary", "--kinds", "asset", "--retrieval", "keyword"},
		{"--summary", "--kinds", "asset", "--retrieval", "semantic"},
		{"--summary", "--kinds", "asset", "--retrieval", "hybrid"},
		{"--summary", "--kinds", "asset", "--require-shape"},
		{"--summary", "--kinds", "asset", "--lifecycles", "active"},
		{"--summary", "--kinds", "asset", "--content-adapter", "registered_asset"},
	} {
		if err := runLibraryIndex("library-find", args); err == nil || !strings.Contains(err.Error(), "omit --summary") {
			t.Fatalf("unsupported registry route accepted: %v %v", args, err)
		}
	}
	for _, mode := range []string{"unknown", "semantics"} {
		if err := runLibraryIndex("library-find", []string{"--retrieval", mode}); err == nil || !strings.Contains(err.Error(), "--retrieval") {
			t.Fatal("unsupported mode did not fail before index access", mode, err)
		}
	}
	if err := runLibraryIndex("library-index", []string{"--retrieval", "keyword"}); err == nil || !strings.Contains(err.Error(), "does not accept") {
		t.Fatal("discovery flag accepted by producer", err)
	}
}

func TestEmbeddingCommandsRejectAmbiguousArgumentsBeforeIO(t *testing.T) {
	for _, args := range [][]string{{"--out", "new"}, {"--out", "new", "--from", "source", "--download"}, {"--query", "cards"}} {
		if err := runLibraryModel(args); err == nil {
			t.Fatal("invalid model maintenance accepted", args)
		}
	}
	for _, command := range []string{"library-index", "library-inspect", "library-embed"} {
		if err := runLibraryIndex(command, []string{"--retrieval", "hybrid"}); err == nil || !strings.Contains(err.Error(), "does not accept") {
			t.Fatal(command, err)
		}
	}
	if err := runLibraryIndex("library-find", []string{"--retrieval", "keyword", "--model-dir", "model"}); err == nil || !strings.Contains(err.Error(), "require semantic or hybrid") {
		t.Fatal(err)
	}
}
