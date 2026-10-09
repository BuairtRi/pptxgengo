package main

import (
	"strings"
	"testing"
)

func TestScaffoldPlaceholderCLIRequiresExplicitAdaptation(t *testing.T) {
	for _, args := range [][]string{
		{"--stock", "--template", "cards/3", "--id", "example", "--placeholder-media"},
		{"--stock", "--template", "architecture/layer-map", "--id", "example", "--source-container-clearance-fit"},
		{"--source-container-clearance-fit", "--bundle", "v11", "--template", "architecture/layer-map"},
		{"--placeholder-media", "--bundle", "v11", "--template", "team-curve/build-together"},
	} {
		err := runProjectScaffold(args)
		if err == nil || !(strings.Contains(err.Error(), "adaptation flags") || strings.Contains(err.Error(), "requires")) {
			t.Fatalf("invalid placeholder invocation accepted: %v %v", args, err)
		}
	}
}
