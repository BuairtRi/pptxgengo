package main

import (
	"reflect"
	"testing"
)

func TestPlatformToolAndCatalogCommands(t *testing.T) {
	path := `C:\Users\Colleague\West Monroe & slides\gallery.html`
	for _, c := range []struct {
		platform, command, binary string
		args                      []string
	}{
		{"windows", "rundll32.exe", "pptxdesign.exe", []string{"url.dll,FileProtocolHandler", "file:///C:/Users/Colleague/West%20Monroe%20&%20slides/gallery.html"}},
		{"darwin", "open", "pptxdesign", []string{path}},
		{"linux", "xdg-open", "pptxdesign", []string{path}},
	} {
		name, args := catalogOpenCommand(path, c.platform)
		if name != c.command || !reflect.DeepEqual(args, c.args) || toolFilename("pptxdesign", c.platform) != c.binary {
			t.Fatalf("%s: %s %q", c.platform, name, args)
		}
	}
}
