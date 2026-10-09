package main

import (
	"strings"
	"testing"
)

func TestAssessmentCLIRejectsAmbiguousFlags(t *testing.T) {
	for _, args := range [][]string{{"inspect", "--slide", "a", "--node", "b", "--apply=false"}, {"patch", "--slide", "a", "--patch", "patch.yaml", "--node", "b"}, {"inspect", "--slide", "a"}, {"patch", "--slide", "a"}, {"unknown"}} {
		if e := runProjectAssessment(args); e == nil || strings.Contains(e.Error(), "no such file") {
			t.Fatalf("flags not rejected before filesystem: %v %v", args, e)
		}
	}
}
