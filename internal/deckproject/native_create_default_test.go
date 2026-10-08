package deckproject

import (
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestPortableCreateNativeDefaultAndExplicitStock(t *testing.T) {
	for _, profile := range []string{"", "stock"} {
		p, err := CreateProject(CreateOptions{Out: filepath.Join(t.TempDir(), "deck"), ID: "default-profile", Title: "Default profile", Year: 2026, Bundle: bundle(t), Template: "cards/3", EditingProfile: profile})
		if err != nil {
			t.Fatal(err)
		}
		want := profile
		if want == "" {
			want = wmdesign.NativeEditingProfile
		}
		if p.Document.EditingProfile != want {
			t.Fatal("new project did not persist rendering intent", p.Document.EditingProfile, want)
		}
	}
}
