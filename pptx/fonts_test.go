package pptx

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeFont builds a minimal byte blob with the given sfnt magic — enough to
// satisfy validateFontData's 12-byte header check.
func fakeFont(magic uint32) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint32(b[:4], magic)
	return b
}

func TestValidateFontData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr string // substring; "" = no error
	}{
		{"ttf outline", fakeFont(0x00010000), ""},
		{"apple true", fakeFont(0x74727565), ""},
		{"otf cff", fakeFont(0x4F54544F), ""},
		{"ttc collection", fakeFont(0x74746366), "collection"},
		{"garbage", []byte("this is definitely not a font!!"), "sfnt"},
		{"too short", []byte{0x00, 0x01}, "sfnt"},
		{"empty", nil, "sfnt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFontData(tt.data)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateFontData() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateFontData() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewEmbeddedFont(t *testing.T) {
	ttf := fakeFont(0x00010000)

	t.Run("requires typeface", func(t *testing.T) {
		_, err := newEmbeddedFont(FontEmbedProps{Regular: ttf})
		if err == nil || !strings.Contains(err.Error(), "Typeface") {
			t.Fatalf("want Typeface-required error, got %v", err)
		}
	})

	t.Run("requires at least one variant", func(t *testing.T) {
		_, err := newEmbeddedFont(FontEmbedProps{Typeface: "Lato"})
		if err == nil || !strings.Contains(err.Error(), "at least one style") {
			t.Fatalf("want no-variants error, got %v", err)
		}
	})

	t.Run("single regular variant", func(t *testing.T) {
		ef, err := newEmbeddedFont(FontEmbedProps{Typeface: "Lato", Regular: ttf})
		if err != nil {
			t.Fatal(err)
		}
		if ef.Typeface != "Lato" || len(ef.Variants) != 1 || ef.Variants[FontRegular] == nil {
			t.Fatalf("unexpected result: %+v", ef)
		}
	})

	t.Run("all four variants", func(t *testing.T) {
		ef, err := newEmbeddedFont(FontEmbedProps{
			Typeface: "Lato", Regular: ttf, Bold: ttf, Italic: ttf, BoldItalic: ttf,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(ef.Variants) != 4 {
			t.Fatalf("want 4 variants, got %d", len(ef.Variants))
		}
	})

	t.Run("invalid variant names its style", func(t *testing.T) {
		_, err := newEmbeddedFont(FontEmbedProps{Typeface: "Lato", Regular: ttf, Bold: []byte("not a font at all!")})
		if err == nil || !strings.Contains(err.Error(), "bold") {
			t.Fatalf("want error naming style bold, got %v", err)
		}
	})

	t.Run("reads from path when bytes absent", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "r.ttf")
		if err := os.WriteFile(p, ttf, 0o644); err != nil {
			t.Fatal(err)
		}
		ef, err := newEmbeddedFont(FontEmbedProps{Typeface: "Lato", RegularPath: p})
		if err != nil {
			t.Fatal(err)
		}
		if len(ef.Variants[FontRegular]) != len(ttf) {
			t.Fatalf("path variant not loaded: %+v", ef)
		}
	})

	t.Run("bytes win over path", func(t *testing.T) {
		otf := fakeFont(0x4F54544F)
		ef, err := newEmbeddedFont(FontEmbedProps{Typeface: "Lato", Regular: otf, RegularPath: "/nonexistent/nope.ttf"})
		if err != nil {
			t.Fatal(err)
		}
		if binary.BigEndian.Uint32(ef.Variants[FontRegular][:4]) != 0x4F54544F {
			t.Fatal("inline bytes should take precedence over path")
		}
	})

	t.Run("missing path errors", func(t *testing.T) {
		_, err := newEmbeddedFont(FontEmbedProps{Typeface: "Lato", RegularPath: "/nonexistent/nope.ttf"})
		if err == nil || !strings.Contains(err.Error(), "reading font file") {
			t.Fatalf("want read error, got %v", err)
		}
	})
}

func TestFontStyleOrderMatchesECMA376(t *testing.T) {
	want := []FontStyle{FontRegular, FontBold, FontItalic, FontBoldItalic}
	if len(fontStyleOrder) != len(want) {
		t.Fatalf("fontStyleOrder length %d, want %d", len(fontStyleOrder), len(want))
	}
	for i, s := range want {
		if fontStyleOrder[i] != s {
			t.Fatalf("fontStyleOrder[%d] = %s, want %s", i, fontStyleOrder[i], s)
		}
	}
	for _, s := range fontStyleOrder {
		if embeddedFontXMLTags[s] == "" {
			t.Fatalf("no XML tag mapped for style %s", s)
		}
	}
}
