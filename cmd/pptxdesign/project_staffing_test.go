package main

import (
	"strings"
	"testing"
)

func TestProjectStaffingRejectsAmbiguousFlags(t *testing.T) {
	for _, args := range [][]string{{"inspect", "--slide", "loop", "--node", "staffing", "--apply=false"}, {"patch", "--slide", "loop", "--patch", "patch.yaml", "--node", "staffing"}, {"inspect", "--slide", "loop"}, {"patch", "--slide", "loop"}, {"unknown"}, {"inspect", "--slide", "loop", "--node", "staffing", "extra"}} {
		if e := runProjectStaffing(args); e == nil || strings.Contains(e.Error(), "no such file") {
			t.Fatalf("flags were not rejected before filesystem access: %v: %v", args, e)
		}
	}
}
