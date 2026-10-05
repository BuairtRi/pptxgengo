package wmdesign

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedFrameLayoutsKeepSlideTypography(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	doc, err := LibrarySourceReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	var chosen SlideSpec
	for _, s := range doc.Slides {
		if s.TemplateBinding.Template == "cards/3" {
			chosen = s
			break
		}
	}
	a, b := chosen, chosen
	a.ID, b.ID = "first", "second"
	a.Title, b.Title = "First title", "Second title"
	doc.Slides = []SlideSpec{a, b}
	doc.BuildIdentity = &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "sharing"}
	raw, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	layouts := 0
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "ppt/slideLayouts/slideLayout") && strings.HasSuffix(f.Name, ".xml") {
			layouts++
		}
		if f.Name == "ppt/slides/slide1.xml" || f.Name == "ppt/slides/slide2.xml" {
			r, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			b, err := io.ReadAll(r)
			r.Close()
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(b), "wm.legal") {
				t.Fatal("shared legal text leaked onto slide")
			}
		}
	}
	if layouts != 2 {
		t.Fatalf("layouts=%d want default plus one shared frame", layouts)
	}
	for i, s := range report.Slides {
		legal, page := 0, 0
		for _, tr := range s.Texts {
			if tr.ID == "wm.legal" {
				legal++
			}
			if tr.ID == "wm.page" {
				page++
				if tr.Layout.Original != []string{"1", "2"}[i] {
					t.Fatalf("slide page text:%q", tr.Layout.Original)
				}
			}
		}
		if legal != 1 || page != 1 {
			t.Fatalf("slide %d legal=%d page=%d", i, legal, page)
		}
	}
}
