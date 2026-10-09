package main

import (
	"strings"
	"testing"
)

func TestProjectGanttRejectsAmbiguousFlags(t *testing.T) {
	for _, args := range [][]string{{"inspect", "--slide", "plan", "--node", "schedule", "--apply=false"}, {"patch", "--slide", "plan", "--patch", "patch.yaml", "--node", "schedule"}, {"inspect", "--slide", "plan"}, {"patch", "--slide", "plan"}, {"unknown"}} {
		if e := runProjectGantt(args); e == nil || strings.Contains(e.Error(), "no such file") {
			t.Fatalf("command flags were not rejected before filesystem access: %v: %v", args, e)
		}
	}
}
