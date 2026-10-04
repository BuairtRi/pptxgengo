package main

import "testing"

func TestRenderErrorOutputAfterBadFlag(t *testing.T) {
	for _, item := range []struct {
		args []string
		want string
	}{
		{[]string{"--timeout", "nope", "--out", "/tmp/task output"}, "/tmp/task output"},
		{[]string{"--unknown", "--out=/tmp/task"}, "/tmp/task"},
		{[]string{"--out", "/tmp/first", "-out=/tmp/last"}, "/tmp/last"},
		{[]string{"--unknown", "--", "--out", "/tmp/positional"}, ""},
		{[]string{"--out"}, ""},
		{[]string{"--pptx", "--out=/tmp/not-an-option", "--unknown"}, ""},
		{[]string{"--pptx", "--out", "--timeout", "nope", "--out", "/tmp/actual-output"}, "/tmp/actual-output"},
	} {
		if got := renderErrorOutput(item.args); got != item.want {
			t.Fatal(item.args, got, item.want)
		}
	}
}
