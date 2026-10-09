package main

import (
	"strings"
	"testing"
)

func TestProjectLayerRejectsAmbiguousFlags(t *testing.T) {
	for _, args := range [][]string{{"inspect", "--slide", "architecture", "--apply=false"}, {"patch", "--slide", "architecture", "--patch", "patch.yaml", "--selection", "layers.json"}, {"inspect"}, {"patch", "--slide", "architecture"}, {"unknown"}, {"inspect", "--slide", "architecture", "extra"}} {
		if e := runProjectLayer(args); e == nil || strings.Contains(e.Error(), "no such file") {
			t.Fatalf("flags not rejected before filesystem access %v: %v", args, e)
		}
	}
}
