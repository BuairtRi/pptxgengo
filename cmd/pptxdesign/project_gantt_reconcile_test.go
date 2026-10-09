package main

import (
	"strings"
	"testing"
)

func TestProjectGanttSemanticReconcileFlags(t *testing.T) {
	for _, args := range [][]string{{"reconcile", "--slide", "s", "--node", "n"}, {"reconcile", "--slide", "s", "--packet", "p"}, {"reconcile", "--slide", "s", "--node", "n", "--packet", "p", "--apply"}, {"reconcile", "--slide", "s", "--node", "n", "--packet", "p", "--patch", "patch"}, {"patch", "--slide", "s", "--patch", "patch", "--packet", "p"}} {
		if e := runProjectGantt(args); e == nil || strings.Contains(e.Error(), "no such file") {
			t.Fatalf("invalid flags reached files: %v: %v", args, e)
		}
	}
}
