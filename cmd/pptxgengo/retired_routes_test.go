package main

import (
	"os"
	"strings"
	"testing"
)

func TestLegacyRoutesFailWithActionableDeprecation(t *testing.T) {
	for route, guidance := range retiredRoutes {
		if _, shipped := tools[route]; shipped {
			t.Errorf("retired route %q remains in shipped command map", route)
		}
		if strings.TrimSpace(guidance) == "" {
			t.Errorf("retired route %q has no actionable message", route)
		}
	}
	if _, ok := tools["design"]; !ok {
		t.Fatal("design route missing")
	}
}

func TestRetiredRouteReturnsClearError(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"pptxgengo", "compose"}
	err := run()
	if err == nil || !strings.Contains(err.Error(), "was removed") || !strings.Contains(err.Error(), "design build") {
		t.Fatalf("retired route message not actionable: %v", err)
	}
}
