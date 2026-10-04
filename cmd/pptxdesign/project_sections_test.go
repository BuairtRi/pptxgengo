package main

import (
	"testing"
)

func TestSectionValuesStrictJSON(t *testing.T) {
	for _, raw := range []string{`{}broken`, `{} {}`, `{"slots":{"a":1,"a":2}}`, `{"a":[{"b":1,"b":2}]}`, `null`, `[]`, `{"a":1} {broken`} {
		if _, e := decodeSectionValues([]byte(raw)); e == nil {
			t.Fatalf("invalid JSON accepted %s", raw)
		}
	}
	if _, e := decodeSectionValues([]byte(`{"slots":{"a":"caller"},"keys":{"x":["one","two"]}}  `)); e != nil {
		t.Fatal(e)
	}
}
func TestSectionCLIRejectsUnsupportedFlags(t *testing.T) {
	for _, args := range [][]string{{"list", "--id", "x"}, {"rename", "--before", "x"}, {"remove", "--title", "x"}, {"what"}, {"add", "extra"}} {
		if e := runProjectSection(args); e == nil {
			t.Fatal("unsupported CLI accepted", args)
		}
	}
}
