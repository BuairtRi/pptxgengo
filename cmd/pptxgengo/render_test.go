package main

import (
	"reflect"
	"testing"
)

func TestNativeRenderNeedsNoBundleOrEngine(t *testing.T) {
	input := []string{"render", "--pptx", "/tmp/a.pptx", "--out", "/tmp/native", "--pdf", "--png"}
	got, err := designArgs("/nonexistent-release", input)
	if err != nil || !reflect.DeepEqual(got, input) {
		t.Fatal(got, err)
	}
}
