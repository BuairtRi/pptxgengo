package nativeexport

import "testing"

func TestDialogMonitorOnlyUsesPrivateTaskScript(t *testing.T) {
	for _, item := range []struct {
		args []string
		want bool
	}{
		{[]string{"-e", "with timeout", "-e", "count presentations", "-e", "end timeout"}, false},
		{[]string{"/private/task/export.applescript", "/private/task/deck.pptx", "/private/task/deck.pdf", "task.pptx", "12"}, true},
		{[]string{"/private/task/export.applescript", "/private/task/deck.pptx", "/private/task/unused.pdf", "task.pptx", "12", "probe"}, true},
		{[]string{"/private/task/export.applescript", "/private/task/deck.pptx", "/private/task/unused.pdf", "task.pptx", "3", "close"}, false},
		{[]string{"/private/task/export.applescript", "/stable/task.pptx", "/stable/task.pdf", "task.pptx", "12", "export", "/private/task/presentation-identity.pptx"}, true},
		{[]string{"/private/task/export.applescript", "/stable/task.pptx", "/stable/task.pdf", "task.pptx", "3", "close", "/private/task/presentation-identity.pptx"}, false},
		{[]string{"relative.applescript", "deck.pptx", "deck.pdf", "task.pptx", "12"}, false},
	} {
		if got := monitorAppleScript(item.args); got != item.want {
			t.Fatalf("%v: monitor=%v, want %v", item.args, got, item.want)
		}
	}
}
