package browsingartifact_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/browsingartifact"
	"github.com/buairtri/pptxgengo/internal/browsingfixture"
)

func artifact(t *testing.T, raw []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "native.pptx")
	if e := os.WriteFile(p, raw, 0644); e != nil {
		t.Fatal(e)
	}
	return p
}
func mutate(t *testing.T, base []byte, change func(map[string][]byte)) []byte {
	t.Helper()
	z, e := zip.NewReader(bytes.NewReader(base), int64(len(base)))
	if e != nil {
		t.Fatal(e)
	}
	parts := map[string][]byte{}
	for _, f := range z.File {
		r, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		var b bytes.Buffer
		b.ReadFrom(r)
		r.Close()
		parts[f.Name] = b.Bytes()
	}
	change(parts)
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for name, data := range parts {
		f, e := w.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = f.Write(data); e != nil {
			t.Fatal(e)
		}
	}
	if e = w.Close(); e != nil {
		t.Fatal(e)
	}
	return out.Bytes()
}
func TestBrowsingPresentationClosure(t *testing.T) {
	base := browsingfixture.Deck(t, "templates")
	file := artifact(t, base)
	raw := browsingfixture.Manifest(t, "templates", browsingfixture.Hash(base))
	if e := browsingartifact.ValidateFile(raw, "templates", browsingfixture.Hash(base), file); e != nil {
		t.Fatal(e)
	}
	cases := map[string]func(map[string][]byte){
		"orphan_part":  func(p map[string][]byte) { p["ppt/slides/orphan.xml"] = p["ppt/slides/slide1.xml"] },
		"missing_part": func(p map[string][]byte) { delete(p, "ppt/slides/slide1.xml") },
		"external_slide": func(p map[string][]byte) {
			p["ppt/_rels/presentation.xml.rels"] = bytes.Replace(p["ppt/_rels/presentation.xml.rels"], []byte(`Target="slides/slide1.xml"`), []byte(`Target="https://example.invalid/slide.xml" TargetMode="External"`), 1)
		},
		"duplicate_reference": func(p map[string][]byte) {
			pattern := regexp.MustCompile(`<p:sldId\b[^>]*/>`)
			ids := pattern.FindAll(p["ppt/presentation.xml"], -1)
			p["ppt/presentation.xml"] = bytes.Replace(p["ppt/presentation.xml"], ids[1], ids[0], 1)
		},
		"duplicate_slide_name": func(p map[string][]byte) {
			p["ppt/slides/slide2.xml"] = bytes.Replace(p["ppt/slides/slide2.xml"], []byte(`name="page-2"`), []byte(`name="page-1"`), 1)
		},
		"invalid_slide_root": func(p map[string][]byte) {
			p["ppt/slides/slide1.xml"] = bytes.ReplaceAll(p["ppt/slides/slide1.xml"], []byte("p:sld"), []byte("p:invalid"))
		},
		"path_traversal": func(p map[string][]byte) { p["../outside.xml"] = []byte("fixture") },
		"case_collision": func(p map[string][]byte) { p["PPT/slides/slide1.xml"] = p["ppt/slides/slide1.xml"] },
		"xml_limit":      func(p map[string][]byte) { p["ppt/presentation.xml"] = bytes.Repeat([]byte("x"), 9<<20) },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			file := artifact(t, mutate(t, base, change))
			if _, e := browsingartifact.PresentationPages(file); e == nil {
				t.Fatal("invalid actual presentation accepted")
			}
		})
	}
	reordered := mutate(t, base, func(p map[string][]byte) {
		pattern := regexp.MustCompile(`<p:sldId\b[^>]*/>`)
		ids := pattern.FindAll(p["ppt/presentation.xml"], -1)
		p["ppt/presentation.xml"] = bytes.Replace(p["ppt/presentation.xml"], append(append([]byte{}, ids[0]...), ids[1]...), append(append([]byte{}, ids[1]...), ids[0]...), 1)
	})
	file = artifact(t, reordered)
	pages, e := browsingartifact.PresentationPages(file)
	if e != nil || pages[0] != "page-2" {
		t.Fatal(pages, e)
	}
	raw = browsingfixture.Manifest(t, "templates", browsingfixture.Hash(reordered))
	if e = browsingartifact.ValidateFile(raw, "templates", browsingfixture.Hash(reordered), file); e == nil {
		t.Fatal("manifest order ignored")
	}
}
func TestBrowsingTypedCoverageAndApproval(t *testing.T) {
	for _, kind := range []string{"templates", "reusable"} {
		deck := browsingfixture.Deck(t, kind)
		file := artifact(t, deck)
		hash := browsingfixture.Hash(deck)
		original := browsingfixture.Manifest(t, kind, hash)
		if e := browsingartifact.ValidateFile(original, kind, hash, file); e != nil {
			t.Fatal(e)
		}
		var m map[string]any
		json.Unmarshal(original, &m)
		coverage := m["coverage"].(map[string]any)
		if kind == "templates" {
			coverage["expected_templates"] = 2
		} else {
			revision := coverage["revisions"].([]any)[0].(map[string]any)
			revision["included"] = false
		}
		raw, _ := json.Marshal(m)
		if e := browsingartifact.ValidateFile(raw, kind, hash, file); e == nil {
			t.Fatal("coverage/approval omission accepted", kind)
		}
	}
	p := filepath.Join(t.TempDir(), "too-large.json")
	if e := os.WriteFile(p, bytes.Repeat([]byte("x"), browsingartifact.MaxManifestBytes+1), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := browsingartifact.ReadManifest(p); e == nil || !strings.Contains(e.Error(), "16MiB") {
		t.Fatal("manifest size bound ignored", e)
	}
}
