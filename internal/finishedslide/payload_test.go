package finishedslide

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPayloadExactClosureAndRevision(t *testing.T) {
	root, m, _ := reviewFixture(t)
	files, e := ReadPayload(root, m)
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range m.Files {
		raw, e := os.ReadFile(filepath.Join(root, f.Path))
		if e != nil || !bytes.Equal(files[f.Path], raw) {
			t.Fatal("payload bytes differ", e)
		}
	}
	newer := m
	newer.Revision++
	if e := newer.Seal(); e != nil {
		t.Fatal(e)
	}
	if _, e := ReadPayload(root, newer); e == nil {
		t.Fatal("wrong revision materialized")
	}
	if e := os.WriteFile(filepath.Join(root, "unexpected.txt"), []byte("Unlisted file"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := ReadPayload(root, m); e == nil {
		t.Fatal("unclosed package materialized")
	}
}

func TestReadPayloadBoundsBeforeReading(t *testing.T) {
	for _, aggregate := range []bool{false, true} {
		_, m := fixture(t)
		if aggregate {
			m.Files = nil
			for i := 0; i < 9; i++ {
				name := string(rune('a'+i)) + ".bin"
				role := "documentation"
				if i == 0 {
					m.Source = name
					role = "source"
				}
				m.Files = append(m.Files, File{Path: name, Role: role, Bytes: 64 << 20, SHA256: sha([]byte(name))})
			}
		} else {
			m.Files[0].Bytes = (64 << 20) + 1
		}
		if e := m.Seal(); e != nil {
			t.Fatal(e)
		}
		if _, e := ReadPayload(filepath.Join(t.TempDir(), "does-not-exist"), m); e == nil || !strings.Contains(e.Error(), "bounds") {
			t.Fatal("bounds not checked before file I/O", e)
		}
	}
}
