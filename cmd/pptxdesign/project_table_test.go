package main

import (
	"strings"
	"testing"
)

func TestProjectTableArgumentGuards(t *testing.T) {
	for _, args := range [][]string{nil, {"inspect"}, {"patch", "--slide", "slide"}, {"inspect", "--slide", "slide", "--node", "table", "--apply=false"}, {"patch", "--slide", "slide", "--patch", "missing", "--node", "table"}, {"inspect", "--slide", "slide", "--node", "table", "extra"}} {
		if e := runProjectTable(args); e == nil {
			t.Fatalf("invalid arguments accepted %v", args)
		}
	}
	e := runProjectTable([]string{"inspect", "--slide", "slide", "--node", "table", "--patch", "missing"})
	if e == nil || !strings.Contains(e.Error(), "not accepted") {
		t.Fatalf("irrelevant flag %v", e)
	}
}
