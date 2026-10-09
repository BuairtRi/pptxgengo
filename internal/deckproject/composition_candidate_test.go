package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func compositionTransactionFixture(t *testing.T) (*Project, LocalTemplate) {
	t.Helper()
	p := example(t)
	if _, err := PortableLayout(p, true); err != nil {
		t.Fatal(err)
	}
	p, err := Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	rel := p.TemplateFiles["editorial-photo"]
	doc, err := sourceYAML(p.SourceFiles[rel])
	if err != nil {
		t.Fatal(err)
	}
	doc.Content[0].HeadComment = "Local template operator comment"
	group := mappingNode(doc.Content[0], "nodes").Content[0]
	summary := mappingNode(group, "nodes").Content[0]
	mappingNode(mappingNode(summary, "text"), "binding").LineComment = "Preserve authored binding comment"
	raw, err := encodeSourceYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(p.Root, rel), raw, 0600); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	pin(t, p)
	var candidate LocalTemplate
	if err = strictInto(p.Document.LocalTemplates["editorial-photo"], &candidate); err != nil {
		t.Fatal(err)
	}
	candidate.Nodes[0].Nodes[0].Placement.Span.Y += 2
	return p, candidate
}

func compositionFileSnapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[rel], err = os.ReadFile(path)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

func TestCompositionCandidatePreviewAndAuthoredPreservation(t *testing.T) {
	p, candidate := compositionTransactionFixture(t)
	before := compositionFileSnapshot(t, p.Root)
	sourceHash := p.SourceHash()
	originalDocument := append([]byte(nil), p.Canonical...)
	preview, err := CompositionCandidate(p, "local-composition", "composition-test", "operator", "Adjust source geometry", candidate, bundle(t), wmdesign.CandidateEngine, false)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Applied || preview.Decision != "" || preview.BeforeSHA256 != sourceHash || preview.AfterSHA256 == sourceHash || preview.Inspection.SourceSHA256 != preview.AfterSHA256 {
		t.Fatalf("invalid measured preview: %+v", preview)
	}
	if !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) || p.SourceHash() != sourceHash || !bytes.Equal(p.Canonical, originalDocument) {
		t.Fatal("preview wrote files or mutated loaded source")
	}
	applied, err := CompositionCandidate(p, "local-composition", "composition-test", "operator", "Adjust source geometry", candidate, bundle(t), wmdesign.CandidateEngine, true)
	if err != nil {
		t.Fatal(err)
	}
	next, err := Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if !applied.Applied || next.SourceHash() != preview.AfterSHA256 {
		t.Fatal("apply differs from measured preview", applied)
	}
	if _, err = os.Stat(filepath.Join(p.Root, applied.Decision)); err != nil {
		t.Fatal("missing decision receipt", err)
	}
	templateFile := p.TemplateFiles["editorial-photo"]
	for rel, data := range p.SourceFiles {
		if rel != templateFile && !bytes.Equal(next.SourceFiles[rel], data) {
			t.Fatalf("untouched authored file changed: %s", rel)
		}
	}
	updated := next.Document.LocalTemplates["editorial-photo"].Nodes[0].Nodes[0]
	original := p.Document.LocalTemplates["editorial-photo"].Nodes[0].Nodes[0]
	if !bytes.Equal(canonical(updated.Text), canonical(original.Text)) || updated.Placement.Span.Y != original.Placement.Span.Y+2 {
		t.Fatal("binding or requested geometry changed")
	}
	for _, comment := range []string{"Local template operator comment", "Preserve authored binding comment"} {
		if !bytes.Contains(next.SourceFiles[templateFile], []byte(comment)) {
			t.Errorf("composition discarded comment %q", comment)
		}
	}
}

func TestCompositionCandidateRefusesSharedPinnedAndNativeState(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		mutate     func(*Project)
	}{
		{"shared-template", "project detach", func(p *Project) { p.Document.Slides[1].Template.Scope = "shared" }},
		{"shared-local-template", "shared", func(p *Project) {
			s := p.Document.Slides[1]
			s.ID = "second-local"
			p.Document.Slides = append(p.Document.Slides, s)
		}},
		{"pinned-template", "pinned", func(p *Project) { p.Document.Slides[1].Template.Revision = "r1" }},
		{"native-geometry", "reset_native_layout", func(p *Project) {
			p.Document.Slides[1].NativeGeometry = map[string]NativeGeometry{"summary": {Kind: "sp"}}
		}},
		{"native-order", "reset_native_layout", func(p *Project) {
			p.Document.Slides[1].NativeOrder = map[string][]string{"editorial": {"summary", "photo"}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, candidate := compositionTransactionFixture(t)
			tc.mutate(p)
			before := compositionFileSnapshot(t, p.Root)
			result, err := CompositionCandidate(p, "local-composition", "composition-test", "operator", "Guarded edit", candidate, bundle(t), wmdesign.CandidateEngine, true)
			if err == nil || !strings.Contains(err.Error(), tc.want) || result.Applied {
				t.Fatalf("expected refusal %q: %+v %v", tc.want, result, err)
			}
			if !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
				t.Fatal("refused candidate changed files")
			}
		})
	}
}

func TestCompositionCandidateNestedCommentsFollowNodeIdentity(t *testing.T) {
	p, candidate := compositionTransactionFixture(t)
	children := candidate.Nodes[0].Nodes
	children[0], children[1] = children[1], children[0]
	_, err := CompositionCandidate(p, "local-composition", "composition-test", "operator", "Reorder nested composition pieces", candidate, bundle(t), wmdesign.CandidateEngine, true)
	if err != nil {
		t.Fatal(err)
	}
	next, err := Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := sourceYAML(next.SourceFiles[next.TemplateFiles["editorial-photo"]])
	if err != nil {
		t.Fatal(err)
	}
	group := mappingNode(doc.Content[0], "nodes").Content[0]
	for _, child := range mappingNode(group, "nodes").Content {
		if mappingNode(child, "id").Value != "summary" {
			continue
		}
		binding := mappingNode(mappingNode(child, "text"), "binding")
		if binding.Value != "summary" || !strings.Contains(binding.LineComment, "Preserve authored binding comment") {
			t.Fatal("nested comment did not follow its stable node identity", binding)
		}
		return
	}
	t.Fatal("reordered source node disappeared")
}

func TestCompositionCandidateGuardsStaleSourceAndFrame(t *testing.T) {
	t.Run("source-drift", func(t *testing.T) {
		p, candidate := compositionTransactionFixture(t)
		// A colleague changes an unrelated authored slide after we loaded the
		// project. The shared transaction must guard the complete source tree.
		rel := p.SlideFiles["maintain-the-source"]
		if err := os.WriteFile(filepath.Join(p.Root, rel), append([]byte("# Concurrent colleague edit\n"), p.SourceFiles[rel]...), 0600); err != nil {
			t.Fatal(err)
		}
		before := compositionFileSnapshot(t, p.Root)
		result, err := CompositionCandidate(p, "local-composition", "composition-test", "operator", "Apply stale preview", candidate, bundle(t), wmdesign.CandidateEngine, true)
		if err == nil || !strings.Contains(err.Error(), "authored source changed") || result.Applied {
			t.Fatalf("stale predecessor accepted: %+v %v", result, err)
		}
		if !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
			t.Fatal("stale apply changed files or left a receipt/lock")
		}
	})
	t.Run("frame-contract", func(t *testing.T) {
		p, candidate := compositionTransactionFixture(t)
		candidate.Name = "Unauthorized template metadata change"
		before := compositionFileSnapshot(t, p.Root)
		_, err := CompositionCandidate(p, "local-composition", "composition-test", "operator", "Edit provenance", candidate, bundle(t), wmdesign.CandidateEngine, true)
		if err == nil || !strings.Contains(err.Error(), "preserve template frame, grid and provenance") || !reflect.DeepEqual(before, compositionFileSnapshot(t, p.Root)) {
			t.Fatal("frame/metadata contract not preserved", err)
		}
	})
}
