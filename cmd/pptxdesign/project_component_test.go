package main

import (
	"flag"
	"strings"
	"testing"
)

func TestProjectComponentArgumentGuards(t *testing.T) {
	for _, args := range [][]string{nil, {"inspect"}, {"patch", "--slide", "slide"}, {"inspect", "--slide", "slide", "--node", "card", "--apply=false"}, {"patch", "--slide", "slide", "--patch", "missing", "--node", "card"}, {"inspect", "--slide", "slide", "--node", "card", "extra"}} {
		if e := runProjectComponent(args); e == nil {
			t.Fatalf("invalid arguments accepted %v", args)
		}
	}
	e := runProjectComponent([]string{"inspect", "--slide", "slide", "--node", "card", "--patch", "missing"})
	if e == nil || !strings.Contains(e.Error(), "not accepted") {
		t.Fatalf("irrelevant flag %v", e)
	}
}

func TestProjectOwnedCompositionHelpAndRouting(t *testing.T) {
	// Exercises the actual public dispatch paths used by the six operator
	// references. Help must work without trying to load a nonexistent project.
	for _, name := range []string{"process", "journey", "portfolio", "component", "table"} {
		for _, sub := range []string{"inspect", "patch"} {
			if e := runProject([]string{name, sub, "--help"}); e != flag.ErrHelp {
				t.Fatalf("%s %s --help: %v", name, sub, e)
			}
		}
	}
}
