package main

import (
	"strings"
	"testing"
)

func TestProjectCycleRejectsAmbiguousFlags(t *testing.T) {
	for _, args := range [][]string{{"inspect", "--slide", "loop", "--node", "cycle", "--apply=false"}, {"patch", "--slide", "loop", "--patch", "patch.yaml", "--node", "cycle"}, {"inspect", "--slide", "loop"}, {"patch", "--slide", "loop"}, {"unknown"}, {"inspect", "--slide", "loop", "--node", "cycle", "extra"}} {
		if e := runProjectCycle(args); e == nil || strings.Contains(e.Error(), "no such file") {
			t.Fatalf("flags were not rejected before filesystem access: %v: %v", args, e)
		}
	}
}
