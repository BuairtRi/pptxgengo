package main

import (
	"strings"
	"testing"
)

func TestDiagramContainmentFlagsAreCommandSpecific(t *testing.T) {
	for _, args := range [][]string{
		{"uncontain", "--slide", "test", "--padding", "12"},
		{"uncontain", "--slide", "test", "--container", "box"},
		{"inspect", "--slide", "test", "--padding-top", "28"},
		{"connect", "--slide", "test", "--container", "box"},
		{"contain", "--slide", "test", "--align", "left"},
	} {
		if e := runProjectDiagram(args); e == nil || !strings.Contains(e.Error(), "not accepted") {
			t.Fatalf("irrelevant flags not refused: %v: %v", args, e)
		}
	}
}
