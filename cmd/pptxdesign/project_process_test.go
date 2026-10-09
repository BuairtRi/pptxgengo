package main

import (
	"strings"
	"testing"
)

func TestProjectProcessArgumentShape(t *testing.T) {
	for _, args := range [][]string{{}, {"invalid"}, {"inspect", "--slide", "s"}, {"inspect", "--slide", "s", "--node", "n", "--apply"}, {"patch", "--slide", "s", "--patch", "p", "--node", "n"}, {"patch", "--slide", "s"}, {"inspect", "--slide", "s", "--node", "n", "extra"}} {
		if e := runProjectProcess(args); e == nil {
			t.Fatal("invalid process arguments accepted", args)
		}
	}
	if e := runProjectProcess([]string{"inspect", "--slide", "s", "--node", "n", "--project", "/missing-process-project"}); e == nil || strings.Contains(e.Error(), "usage:") {
		t.Fatalf("valid argument routing did not reach loading: %v", e)
	}
}
