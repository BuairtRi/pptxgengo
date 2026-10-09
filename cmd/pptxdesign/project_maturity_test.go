package main

import (
	"strings"
	"testing"
)

func TestProjectMaturityRejectsAmbiguousFlags(t *testing.T) {
	for _, args := range [][]string{{"inspect", "--slide", "loop", "--node", "maturity", "--apply=false"}, {"patch", "--slide", "loop", "--patch", "patch.yaml", "--node", "maturity"}, {"inspect", "--slide", "loop"}, {"patch", "--slide", "loop"}, {"unknown"}, {"inspect", "--slide", "loop", "--node", "maturity", "extra"}} {
		if e := runProjectMaturity(args); e == nil || strings.Contains(e.Error(), "no such file") {
			t.Fatalf("flags were not rejected before filesystem access: %v: %v", args, e)
		}
	}
}
