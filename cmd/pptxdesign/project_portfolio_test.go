package main

import (
	"strings"
	"testing"
)

func TestProjectPortfolioArgumentShape(t *testing.T) {
	for _, args := range [][]string{{}, {"invalid"}, {"inspect", "--slide", "s"}, {"inspect", "--slide", "s", "--node", "n", "--apply"}, {"patch", "--slide", "s", "--patch", "p", "--node", "n"}, {"patch", "--slide", "s"}, {"inspect", "--slide", "s", "--node", "n", "extra"}} {
		if e := runProjectPortfolio(args); e == nil {
			t.Fatal("invalid portfolio arguments accepted", args)
		}
	}
	if e := runProjectPortfolio([]string{"inspect", "--slide", "s", "--node", "n", "--project", "/missing-portfolio-project"}); e == nil || strings.Contains(e.Error(), "usage:") {
		t.Fatalf("valid argument routing did not reach loading: %v", e)
	}
}
