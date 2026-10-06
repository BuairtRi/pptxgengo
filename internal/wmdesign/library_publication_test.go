package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPublicationCompositionIgnoresOnlyProvenance(t *testing.T) {
	original := SlideSpec{ID: "source-001", Title: "Accepted title", Source: "Accepted source", Frame: FrameRequest{Rail: "none"}, TemplateBinding: &TemplateSlideRecord{Template: "example/one", SourceRevision: "wmds-library.v5"}}
	newer := original
	newer.ID = "source-615"
	newer.TemplateBinding = &TemplateSlideRecord{Template: "example/one", SourceRevision: "wmds-library.v7"}
	if !samePublicationComposition(original, newer) {
		t.Fatal("identical authored specimen rejected after provenance changed")
	}
	cases := []struct {
		name   string
		change func(*SlideSpec)
	}{
		{"headline", func(s *SlideSpec) { s.Title = "Changed title" }},
		{"source", func(s *SlideSpec) { s.Source = "Changed source" }},
		{"frame", func(s *SlideSpec) { s.Frame.Rail = "left" }},
		{"nodes", func(s *SlideSpec) { s.Nodes = []Node{{ID: "added", Kind: "text"}} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := newer
			tc.change(&changed)
			if samePublicationComposition(original, changed) {
				t.Fatal("changed authored composition accepted")
			}
		})
	}
}

func TestPublicationEvidenceHashRequired(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accepted.png")
	raw := []byte("native accepted bytes")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))
	if _, err := verifiedPublicationFile(path, hash); err != nil {
		t.Fatal(err)
	}
	if _, err := verifiedPublicationFile(path, ""); err == nil {
		t.Fatal("missing hash accepted")
	}
	if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifiedPublicationFile(path, hash); err == nil {
		t.Fatal("changed evidence accepted")
	}
}

func TestPublicationOutputRejectsDanglingSymlink(t *testing.T) {
	out := filepath.Join(t.TempDir(), "existing-link")
	if err := os.Symlink("missing-target", out); err != nil {
		t.Fatal(err)
	}
	_, err := PublishLibraryGallery(out, LibraryPublicationOptions{Year: 2026, Version: "candidate"})
	if err == nil {
		t.Fatal("existing dangling output symlink accepted")
	}
}

func TestPublicationCompositionCanonicalizesSceneObjectKeys(t *testing.T) {
	first := SlideSpec{Nodes: []Node{{Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(`{"type":"text","x":12,"text":"Accepted"}`)}}}}
	second := SlideSpec{Nodes: []Node{{Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(`{"text":"Accepted","x":12,"type":"text"}`)}}}}
	if !samePublicationComposition(first, second) {
		t.Fatal("JSON property order changed authored composition")
	}
	second.Nodes[0].Scene.Node = json.RawMessage(`{"text":"Accepted","x":13,"type":"text"}`)
	if samePublicationComposition(first, second) {
		t.Fatal("geometry amendment hidden by canonical encoding")
	}
}

func publicationFixtureZip(t *testing.T, parts map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for name, value := range parts {
		stream, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = stream.Write([]byte(value)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestPublicationVisibleDependencyClosure(t *testing.T) {
	parts := map[string]string{
		"ppt/slides/slide1.xml":             `<sld><image embed="rId3"/></sld>`,
		"ppt/slides/_rels/slide1.xml.rels":  `<Relationships><Relationship Id="rId1" Type="example/slideLayout" Target="../slideLayouts/slideLayout1.xml"/><Relationship Id="rId2" Type="example/notesSlide" Target="../notesSlides/notesSlide1.xml"/><Relationship Id="rId3" Type="example/image" Target="/ppt/media/image1.png"/></Relationships>`,
		"ppt/slideLayouts/slideLayout1.xml": `<layout/>`,
		"ppt/notesSlides/notesSlide1.xml":   `binding-v5`,
		"ppt/media/image1.png":              `accepted-image`,
	}
	original := publicationFixtureZip(t, parts)
	parts["ppt/notesSlides/notesSlide1.xml"] = "binding-v7"
	if err := comparePublicationVisibleParts(original, publicationFixtureZip(t, parts)); err != nil {
		t.Fatalf("binding-only notes changed visible composition: %v", err)
	}
	parts["ppt/media/image1.png"] = "changed-image"
	if err := comparePublicationVisibleParts(original, publicationFixtureZip(t, parts)); err == nil {
		t.Fatal("changed referenced image accepted")
	}
	parts["ppt/media/image1.png"] = "accepted-image"
	parts["ppt/slideLayouts/slideLayout1.xml"] = "changed-layout"
	if err := comparePublicationVisibleParts(original, publicationFixtureZip(t, parts)); err == nil {
		t.Fatal("changed referenced layout accepted")
	}
}

func TestPublicationRelationshipPart(t *testing.T) {
	for _, target := range []string{"/ppt/charts/chart23.xml", "../charts/chart23.xml"} {
		if got := publicationRelationshipPart("ppt/slides/slide1.xml", target); got != "ppt/charts/chart23.xml" {
			t.Fatalf("target%s resolved%s", target, got)
		}
	}
}

func TestPublicationPairedChartRendering(t *testing.T) {
	root := filepath.Join("..", "..")
	_, err := VerifyLibraryPublicationRenderInheritance(LibraryPublicationOptions{Bundle: filepath.Join(root, "library/wm-design-system/v9"), PreviousBundle: filepath.Join(root, "planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle"), Year: 2026}, []string{"chart/column-full"})
	if err != nil {
		t.Fatal(err)
	}
}
