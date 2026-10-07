package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

func splitExample(t *testing.T, friendly bool) (*Project, []byte) {
	t.Helper()
	p := example(t)
	p.Document.Slides[1].Notes = "\n\n# Exact notes\nA qualifier: 0 is observed, not missing.\n"
	node, err := editYAMLNode(p.Document)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := encodeSourceYAML(node)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.SourcePath, append([]byte("# Main source comment\n"), raw...), 0600); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), p.Canonical...)
	options := SplitOptions{}
	if friendly {
		options.Bundle, options.StockEditor = bundle(t), StockEditableSlide
	}
	if _, err := Split(p, options); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	return p, before
}

func TestSplitPreservesCanonicalNotesAndIndependentFiles(t *testing.T) {
	p, before := splitExample(t, true)
	if !bytes.Equal(before, p.Canonical) {
		t.Fatal("split changed canonical content")
	}
	if len(p.SlideFiles) != len(p.Document.Slides) || len(p.TemplateFiles) == 0 || p.NotesFiles[p.Document.Slides[1].ID] == "" {
		t.Fatal("source references missing")
	}
	if !bytes.Contains(p.Raw, []byte("# Main source comment")) || bytes.Contains(p.Raw, []byte("nodes:")) {
		t.Fatal("main comments or compact structure lost")
	}
	if p.Document.Slides[1].Notes != "\n\n# Exact notes\nA qualifier: 0 is observed, not missing.\n" {
		t.Fatal("Markdown note bytes changed")
	}
	shared := p.Document.Slides[0]
	sharedRaw := p.SourceFiles[p.SlideFiles[shared.ID]]
	if !bytes.Contains(sharedRaw, []byte("content:")) || !bytes.Contains(sharedRaw, []byte("bindings:")) {
		t.Fatal("stock copy aliases missing")
	}
	firstHash := p.SourceHash()
	firstFiles := map[string][]byte{}
	for relative, raw := range p.SourceFiles {
		firstFiles[relative] = append([]byte(nil), raw...)
	}
	if _, err := Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); err != nil {
		t.Fatal(err)
	}
	next, err := Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != next.SourceHash() || !reflect.DeepEqual(firstFiles, next.SourceFiles) {
		t.Fatal("repeated split rewrote authored files")
	}
}

func TestExternalReferencesRejectUnsafeAndMalformedSources(t *testing.T) {
	for _, test := range []struct{ name, source, message string }{
		{"unknown-field", "unknown_property: true\n", "unknown"},
		{"duplicate", "id: repeated\nid: duplicate\n", "duplicate"},
		{"alias", "value: &x hello\ncopy: *x\n", "aliases"},
		{"multiple-documents", "id: one\n---\nid: two\n", "one YAML"},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, _ := splitExample(t, false)
			relative := p.SlideFiles[p.Document.Slides[0].ID]
			if err := os.WriteFile(filepath.Join(p.Root, relative), []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(p.Root)
			if err == nil || !strings.Contains(err.Error(), test.message) || !strings.Contains(filepath.ToSlash(err.Error()), filepath.ToSlash(relative)) {
				t.Fatalf("external diagnostic: %v", err)
			}
		})
	}
	p, _ := splitExample(t, false)
	raw := strings.Replace(string(p.Raw), p.SlideFiles[p.Document.Slides[0].ID], "../escape.yaml", 1)
	if err := os.WriteFile(p.SourcePath, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p.Root); err == nil || !strings.Contains(err.Error(), "unsafe relative") {
		t.Fatalf("unsafe reference: %v", err)
	}
	if err := os.WriteFile(p.SourcePath, p.Raw, 0600); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(p.Root, p.SlideFiles[p.Document.Slides[0].ID])
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(p.SourcePath, file); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p.Root); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlink reference: %v", err)
	}
}

func TestContentAliasesNestedPointersAndClosedConsumption(t *testing.T) {
	p := &Project{SourcePath: "fixture.yaml", Positions: map[string]Position{}, positionFiles: map[string]string{}}
	base := func() map[string]any {
		return map[string]any{
			"content":  map[string]any{"headline": "Interview findings", "interviews": []any{map[string]any{"name": "Sam", "role": "Director"}}},
			"bindings": map[string]any{"headline": "/slots/title", "/interviews/0/name": "/slots/node01.rows.item01.n", "/interviews/0/role": "/slots/node01.rows.item01.r"},
			"values":   map[string]any{"keys": map[string]any{"rows": []any{"sam"}}},
		}
	}
	slide := base()
	if err := p.expandContentAliases(slide, "/slides/0"); err != nil {
		t.Fatal(err)
	}
	if _, exists := slide["content"]; exists {
		t.Fatal("authoring content remained in canonical source")
	}
	if slide["values"].(map[string]any)["slots"].(map[string]any)["node01.rows.item01.r"] != "Director" {
		t.Fatal("nested content did not reach closed binding")
	}
	for _, name := range []string{"unused", "overlap", "collision", "escape", "missing"} {
		t.Run(name, func(t *testing.T) {
			slide := base()
			bindings := slide["bindings"].(map[string]any)
			switch name {
			case "unused":
				slide["content"].(map[string]any)["unbound"] = "silent copy loss"
			case "overlap":
				bindings["/interviews"] = "/slots/extra"
			case "collision":
				slide["values"].(map[string]any)["slots"] = map[string]any{"title": "conflict"}
			case "escape":
				bindings["headline"] = "/slots/bad~2pointer"
			case "missing":
				bindings["/interviews/2/name"] = "/slots/extra"
			}
			if err := p.expandContentAliases(slide, "/slides/0"); err == nil {
				t.Fatal("invalid aliases accepted")
			}
		})
	}
}

func TestExternalSourceReferencesRejectNormalizedDuplicate(t *testing.T) {
	p, _ := splitExample(t, false)
	first := p.SlideFiles[p.Document.Slides[0].ID]
	second := p.SlideFiles[p.Document.Slides[1].ID]
	raw := strings.Replace(string(p.Raw), second, "./"+first, 1)
	if err := os.WriteFile(p.SourcePath, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p.Root); err == nil || !strings.Contains(err.Error(), "duplicate authored source") {
		t.Fatalf("normalized duplicate accepted: %v", err)
	}
}

func TestSplitSlideEditsLeaveUnselectedFilesByteExact(t *testing.T) {
	p, _ := splitExample(t, false)
	pin(t, p)
	selected := p.Document.Slides[1]
	before := map[string][]byte{}
	for relative, raw := range p.SourceFiles {
		before[relative] = append([]byte(nil), raw...)
	}
	values := map[string]any{}
	if err := json.Unmarshal(canonical(selected.Values), &values); err != nil {
		t.Fatal(err)
	}
	values["summary"] = "Edit one slide without touching another."
	if _, err := EditSlides(p, map[string]SlideEdit{selected.ID: {Values: values}}, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	next, err := Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	for relative, raw := range before {
		if relative != p.SlideFiles[selected.ID] && !bytes.Equal(raw, next.SourceFiles[relative]) {
			t.Fatalf("unselected file changed: %s", relative)
		}
	}
	if next.Document.Slides[1].Notes != selected.Notes {
		t.Fatal("slide edit changed Markdown notes")
	}
	if _, err := EditSlides(p, map[string]SlideEdit{selected.ID: {Values: values}}, bundle(t), wmdesign.CandidateEngine); err == nil || !strings.Contains(err.Error(), "source changed") {
		t.Fatalf("stale source mutation accepted: %v", err)
	}
}

func TestSplitForkAndSectionChangesKeepExternalFormat(t *testing.T) {
	p, _ := splitExample(t, false)
	pin(t, p)
	local := p.Document.Slides[1]
	firstFile, err := os.ReadFile(filepath.Join(p.Root, p.SlideFiles[p.Document.Slides[0].ID]))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Fork(p, local.Template.ID, "editable-fork", "Authored alternative", []string{local.ID}); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if p.TemplateFiles["editable-fork"] != "templates/editable-fork.yaml" {
		t.Fatal("fork collapsed external templates")
	}
	if _, err := AddSection(p, SectionAddOptions{ID: "copy", Title: "Editable source", BeforeSlideID: local.ID}); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.SlideFiles) != len(p.Document.Slides) {
		t.Fatal("section mutation collapsed external slides")
	}
	if _, err := RenameSection(p, "copy", "Final source"); err != nil {
		t.Fatal(err)
	}
	unchanged, err := os.ReadFile(filepath.Join(p.Root, p.SlideFiles[p.Document.Slides[0].ID]))
	if err != nil || !bytes.Equal(firstFile, unchanged) {
		t.Fatal("unselected slide was rewritten by fork/section")
	}
}

func TestSplitBuildAndOfflineExportKeepByteIdentity(t *testing.T) {
	p := example(t)
	// Legacy clients also author JSON in a .yaml file. Exercise that actual
	// flow-style source, not only the already-block-form example fixture.
	if err := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); err != nil {
		t.Fatal(err)
	}
	var err error
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	pin(t, p)
	first, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(p.Root, "builds", first.BuildID, "deck.pptx"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(p.Root, "builds", second.BuildID, "deck.pptx"))
	if err != nil || !bytes.Equal(want, got) {
		t.Fatal("source split changed PowerPoint bytes")
	}
	if second.SourceSHA256 != p.SourceHash() {
		t.Fatal("build receipt omitted external source hashes")
	}
	archive := filepath.Join(t.TempDir(), "offline.zip")
	if _, err := Export(p, ExportOptions{Mode: "offline", Bundle: bundle(t), Out: archive}); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	root := t.TempDir()
	for _, file := range z.File {
		path := filepath.Join(root, filepath.FromSlash(file.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		in, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(in)
		in.Close()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	q, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(q.Canonical, p.Canonical) || q.SourceHash() != p.SourceHash() {
		t.Fatal("export lost authored source closure")
	}
	if _, err := Check(q, filepath.Join(root, "runtime/library/wm-design-system/pinned"), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
}

func TestSplitLegacyJSONUsesReadableBlockYAMLAndIsIdempotent(t *testing.T) {
	p := example(t)
	p.Document.Slides[1].Notes = "\n\n# Exact notes\nColon: yes.\n"
	raw := append([]byte("# Preserve main source comment\n"), canonical(p.Document)...)
	if err := os.WriteFile(p.SourcePath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var err error
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte(nil), p.Canonical...)
	options := SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}
	if _, err := Split(p, options); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p.Canonical, want) || !bytes.Contains(p.Raw, []byte("# Preserve main source comment")) || p.Document.Slides[1].Notes != "\n\n# Exact notes\nColon: yes.\n" {
		t.Fatal("formatting changed canonical source, comments or notes")
	}
	for relative, data := range p.SourceFiles {
		if strings.HasSuffix(relative, ".md") {
			continue
		}
		if bytes.Count(data, []byte("\n")) < 4 {
			t.Fatalf("source remains a dense flow line: %s", relative)
		}
		document, err := sourceYAML(data)
		if err != nil {
			t.Fatal(err)
		}
		var check func(*yaml.Node)
		check = func(node *yaml.Node) {
			if (node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode) && len(node.Content) > 0 && node.Style&yaml.FlowStyle != 0 {
				t.Errorf("nonempty authored collection remains flow-style: %s:%d", relative, node.Line)
			}
			for _, child := range node.Content {
				check(child)
			}
		}
		check(document)
	}
	beforeHash := p.SourceHash()
	if _, err := Split(p, options); err != nil {
		t.Fatal(err)
	}
	q, err := Load(p.Root)
	if err != nil || q.SourceHash() != beforeHash {
		t.Fatalf("repeated block split is not byte-idempotent: %v", err)
	}
}

func TestSplitNormalizesReferencedJSONTemplateWithoutChangingCanonical(t *testing.T) {
	p, _ := splitExample(t, false)
	var relative string
	for id, path := range p.TemplateFiles {
		relative = path
		raw := append([]byte("# Preserve template comment\n"), canonical(p.Document.LocalTemplates[id])...)
		if err := os.WriteFile(filepath.Join(p.Root, path), raw, 0600); err != nil {
			t.Fatal(err)
		}
		break
	}
	if relative == "" {
		t.Fatal("expected referenced template fixture")
	}
	var err error
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte(nil), p.Canonical...)
	if _, err := Split(p, SplitOptions{}); err != nil {
		t.Fatal(err)
	}
	q, err := Load(p.Root)
	if err != nil || !bytes.Equal(q.Canonical, want) {
		t.Fatalf("referenced template normalization changed canonical: %v", err)
	}
	if raw := q.SourceFiles[relative]; bytes.Count(raw, []byte("\n")) < 4 || !bytes.Contains(raw, []byte("# Preserve template comment")) {
		t.Fatal("referenced template remains dense or lost comment")
	}
}

func TestSplitRefusesOccupiedDestinationsWithoutChangingSource(t *testing.T) {
	p := example(t)
	before := append([]byte(nil), p.Raw...)
	path := filepath.Join(p.Root, "slides", "001-"+p.Document.Slides[0].ID+".yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("existing author work"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Split(p, SplitOptions{}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("destination overwrite allowed: %v", err)
	}
	after, err := os.ReadFile(p.SourcePath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed split changed main source")
	}
	existing, err := os.ReadFile(path)
	if err != nil || string(existing) != "existing author work" {
		t.Fatal("failed split overwrote author work")
	}
}

func TestExternalNotesEmptyConflictAndSourceInvalidation(t *testing.T) {
	p, _ := splitExample(t, false)
	pin(t, p)
	last := p.Document.Slides[1]
	relative := p.NotesFiles[last.ID]
	path := filepath.Join(p.Root, relative)
	before := p.SourceHash()
	depsBefore, err := dependencies(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}
	next, err := Load(p.Root)
	if err != nil || next.Document.Slides[1].Notes != "" {
		t.Fatalf("empty Markdown notes: %v", err)
	}
	depsAfter, err := dependencies(next)
	if err != nil || reflect.DeepEqual(depsBefore, depsAfter) || next.SourceHash() == before {
		t.Fatal("external note change escaped source dependencies")
	}
	if _, err := EditSlides(p, map[string]SlideEdit{last.ID: {Values: last.Values}}, bundle(t), wmdesign.CandidateEngine); err == nil || !strings.Contains(err.Error(), "source changed") {
		t.Fatalf("concurrent notes change accepted: %v", err)
	}
	slideFile := filepath.Join(p.Root, next.SlideFiles[last.ID])
	raw, err := os.ReadFile(slideFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(slideFile, append(raw, []byte("\nnotes: Conflicting inline notes\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p.Root); err == nil || !strings.Contains(err.Error(), "not both") {
		t.Fatalf("ambiguous notes accepted: %v", err)
	}
}

func TestSplitSectionDividerAndDetachPreserveSourceFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("stock detachment and divider rendering require registered private photography; run make test-integration")
	}
	p, _ := splitExample(t, false)
	pin(t, p)
	shared := p.Document.Slides[0]
	// A supported generic source-scene template provides an exact detach test.
	catalog, err := wmdesign.LibraryCatalog(bundle(t), "")
	if err != nil {
		t.Fatal(err)
	}
	var def wmdesign.LibraryTemplate
	for _, item := range catalog {
		if item.Key == "quote/light" {
			def = item
		}
	}
	specimen, err := wmdesign.LibraryReference(bundle(t), "", def.Family, p.Document.Year)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]any
	for _, slide := range specimen.Slides {
		if slide.Template == def.Key {
			if err := json.Unmarshal(slide.Values, &values); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if values == nil {
		t.Fatal("source fixture missing")
	}
	if _, err := EditSlides(p, map[string]SlideEdit{shared.ID: {Template: &Reference{Scope: "shared", ID: def.Key}, Values: values}}, bundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	otherPath := p.SlideFiles[p.Document.Slides[1].ID]
	otherBefore := append([]byte(nil), p.SourceFiles[otherPath]...)
	if _, err := Detach(p, shared.ID, "detached-quote", bundle(t), wmdesign.CandidateEngine, "Deliberate authored alternative"); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if p.TemplateFiles["detached-quote"] == "" {
		t.Fatal("detach collapsed source format")
	}
	if !bytes.Equal(otherBefore, p.SourceFiles[otherPath]) {
		t.Fatal("detach rewrote another slide")
	}
	if _, err := AddSection(p, SectionAddOptions{ID: "new-section", Title: "Next discussion", BeforeSlideID: p.Document.Slides[1].ID, Divider: "divider/panel-edge", DividerPhoto: "photo-working-session", Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if p.SlideFiles["new-section-divider"] == "" || !bytes.Equal(otherBefore, p.SourceFiles[otherPath]) {
		t.Fatal("divider failed external format or changed another slide")
	}
}
