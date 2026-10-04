package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func example(t *testing.T) *Project {
	t.Helper()
	dst := t.TempDir()
	src := "../../examples/deck-project"
	e := filepath.WalkDir(src, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(src, path)
		if e != nil {
			return e
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if rel == "toolchain.lock.json" || rel == "state.json" || strings.HasPrefix(rel, "builds/") {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(target, b, 0644)
	})
	if e != nil {
		t.Fatal(e)
	}
	p, e := Load(dst)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func bundle(t *testing.T) string {
	t.Helper()
	p, e := filepath.Abs("../../library/wm-design-system/v2")
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func pin(t *testing.T, p *Project) {
	t.Helper()
	if _, e := Pin(p, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
}
func rewrite(t *testing.T, p *Project, old, new string) *Project {
	t.Helper()
	b := strings.Replace(string(p.Raw), old, new, 1)
	if b == string(p.Raw) {
		t.Fatal("fixture replacement did not match")
	}
	if e := os.WriteFile(p.SourcePath, []byte(b), 0644); e != nil {
		t.Fatal(e)
	}
	out, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func TestStrictYAML(t *testing.T) {
	p := example(t)
	for _, tc := range []struct{ name, from, to, err string }{{"duplicate", "id: editable-project", "id: editable-project\nid: duplicate", "duplicate field"}, {"unknown", "title: Maintained presentation source", "title: Maintained presentation source\nunknown: value", "unknown field"}, {"aliases", "title: Maintained presentation source", "title: &name Maintained presentation source\ncontext: {project: *name}", "aliases"}, {"custom-tag", "title: Maintained presentation source", "title: !tag text", "unsupported YAML tag"}, {"multiple-docs", "year: 2026", "year: 2026\n---\nother: document", "exactly one YAML"}, {"wrong-variant", "            style: body", "            style: body\n            fit: cover", "not valid for text"}, {"unsupported-schema", "schema: {type: string, minLength: 1, maxLength: 70}", "schema: {type: string, pattern: text}", "unsupported zone schema"}, {"missing-required", "required: true", "description: missing explicit required", "required boolean"}} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(p.Root, "bad.yaml")
			raw := strings.Replace(string(p.Raw), tc.from, tc.to, 1)
			if e := os.WriteFile(path, []byte(raw), 0644); e != nil {
				t.Fatal(e)
			}
			_, e := Load(path)
			if e == nil || !strings.Contains(e.Error(), tc.err) {
				t.Fatalf("wanted %s got %v", tc.err, e)
			}
		})
	}
	summary := p.Document.Slides[1].Values["summary"].(string)
	if !strings.HasSuffix(summary, "\n") || strings.Count(summary, "\n") != 3 {
		t.Fatalf("block text lost hard breaks: %q", summary)
	}
}
func TestSafePaths(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"../outside", "/absolute", "file://remote", "x\\y", "a/../b", "C:/path"} {
		if _, e := SafePath(root, path); e == nil {
			t.Errorf("accepted unsafe %s", path)
		}
	}
	target := t.TempDir()
	if e := os.Symlink(target, filepath.Join(root, "linked")); e != nil {
		t.Fatal(e)
	}
	if _, e := SafePath(root, "linked/file"); e == nil {
		t.Fatal("accepted symlink")
	}
	if _, e := SafePath(root, "assets/new/file.png"); e != nil {
		t.Fatal(e)
	}
}
func TestMixedBuildDeterminismAndProtection(t *testing.T) {
	p := example(t)
	pin(t, p)
	opts := BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}
	first, e := Build(p, opts)
	if e != nil {
		t.Fatal(e)
	}
	second, e := Build(p, opts)
	if e != nil {
		t.Fatal(e)
	}
	if first.BuildID == second.BuildID {
		t.Fatal("build overwritten")
	}
	if first.Outputs["deck.pptx"] != second.Outputs["deck.pptx"] {
		t.Fatal("native PPTX bytes not reproducible")
	}
	for _, name := range []string{"scene.json", "object-map.json"} {
		if first.Outputs[name] != second.Outputs[name] {
			t.Errorf("%s nondeterministic", name)
		}
	}
	data, e := os.ReadFile(filepath.Join(p.Root, "builds", second.BuildID, "object-map.json"))
	if e != nil {
		t.Fatal(e)
	}
	var objects Objects
	if e = json.Unmarshal(data, &objects); e != nil {
		t.Fatal(e)
	}
	found := false
	for _, o := range objects.Objects {
		if o.SlideID == "local-composition" && o.NodeID == "editorial.summary" && len(o.SourcePointers) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("local logical source map missing")
	}
	baseline := filepath.Join(p.Root, "builds", second.BuildID, "deck.pptx")
	if e = os.Chmod(baseline, 0644); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(baseline, []byte("manual edit"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, opts); e == nil || !strings.Contains(e.Error(), "baseline_divergence") {
		t.Fatalf("manual edit overwritten: %v", e)
	}
	st, e := Status(p)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(st.NextAction, "divergence") {
		t.Fatalf("missing divergence status %+v", st)
	}
}
func TestApprovalsScopeAndWhitespace(t *testing.T) {
	p := example(t)
	pin(t, p)
	a, e := Approve(p, "content", "reviewer", []string{"maintain-the-source"})
	if e != nil {
		t.Fatal(e)
	}
	p = rewrite(t, p, "This paragraph is maintained in deck.yaml.", "This paragraph changed on only the local slide.")
	s, e := Status(p)
	if e != nil {
		t.Fatal(e)
	}
	if len(s.Approvals) != 1 || !s.Approvals[0].Valid || s.Approvals[0].ID != a.ID {
		t.Fatalf("unaffected approval invalidated: %+v", s)
	}
	p = rewrite(t, p, "Edit the human-readable YAML source.", "Edit the maintained source.")
	s, e = Status(p)
	if e != nil {
		t.Fatal(e)
	}
	if s.Approvals[0].Valid {
		t.Fatal("changed approved copy retained approval")
	}
	if _, e = Approve(p, "review", "reviewer", nil); e == nil {
		t.Fatal("review approved without generated build")
	}
	p = rewrite(t, p, "year: 2026", "year: 2026 # purely formatting")
	if _, e = Resume(p); e != nil {
		t.Fatal(e)
	}
}
func TestExportPrivacyAndNewDestination(t *testing.T) {
	p := example(t)
	pin(t, p)
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"client", "reviewer", "maintainer", "offline"} {
		t.Run(mode, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "export.zip")
			r, e := Export(p, ExportOptions{Mode: mode, Out: path, Bundle: bundle(t)})
			if e != nil {
				t.Fatal(e)
			}
			raw, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
			if e != nil {
				t.Fatal(e)
			}
			names := map[string]bool{}
			for _, f := range z.File {
				names[f.Name] = true
				if strings.Contains(f.Name, "..") || strings.HasPrefix(f.Name, "/") {
					t.Fatal("unsafe ZIP member")
				}
			}
			if r.SHA256 == "" {
				t.Fatal("missing export hash")
			}
			if mode == "client" || mode == "reviewer" {
				if names["deck.yaml"] || names["context/project.md"] {
					t.Fatal("private source leaked")
				}
			} else if !names["deck.yaml"] || !names["context/project.md"] || !names["assets/originals/sample.png"] {
				t.Fatal("maintainer sources missing")
			}
			if mode == "offline" && !names["runtime/pptxdesign"] {
				t.Fatal("offline runtime missing")
			}
			if _, e = Export(p, ExportOptions{Mode: mode, Out: path, Bundle: bundle(t)}); e == nil {
				t.Fatal("existing ZIP overwritten")
			}
		})
	}
}
func TestPinAndAssetDrift(t *testing.T) {
	p := example(t)
	pin(t, p)
	if _, e := Pin(p, bundle(t), wmdesign.CandidateEngine); e == nil {
		t.Fatal("lock silently replaced")
	}
	if _, e := Check(p, bundle(t), wmdesign.Engine); e == nil {
		t.Fatal("engine drift accepted")
	}
	p = rewrite(t, p, "path: assets/originals/sample.png", "path: assets/originals/sample.png\n    sha256: \""+strings.Repeat("0", 64)+"\"")
	if _, e := Check(p, bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "asset hash mismatch") {
		t.Fatalf("expected asset drift error: %v", e)
	}
}

func TestGenericLocalComposition(t *testing.T) {
	p := example(t)
	pin(t, p)
	raw := strings.Replace(string(p.Raw), "kind: text\n            placement:", "kind: component\n            definition: {scope: shared, id: wmds/component/text}\n            arguments: {style: body, ink: primary, text: {binding: summary}}\n            placement:", 1)
	raw = strings.Replace(raw, "            style: body\n            ink: primary\n            text: {binding: summary}\n", "", 1)
	if e := os.WriteFile(p.SourcePath, []byte(raw), 0644); e != nil {
		t.Fatal(e)
	}
	var e error
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	p = rewrite(t, p, "arguments: {style: body, ink: primary, text: {binding: summary}}", "arguments: {style: body, ink: primary, text: {binding: summary}, unsupported: value}")
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e == nil || !strings.Contains(e.Error(), "unsupported_source_field") {
		t.Fatalf("generic source fields not checked: %v", e)
	}
}
func TestStaleBuildCannotBeApproved(t *testing.T) {
	p := example(t)
	pin(t, p)
	p = rewrite(t, p, "project: context/project.md", "project: context/project.md\n  state: state.json")
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e := Approve(p, "review", "reviewer", nil); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 2; i++ {
		if _, e := Resume(p); e != nil {
			t.Fatal(e)
		}
	}
	if e := os.WriteFile(filepath.Join(p.Root, "context/project.md"), []byte("new evidence"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := Approve(p, "review", "reviewer", nil); e == nil || !strings.Contains(e.Error(), "inputs unchanged") {
		t.Fatalf("stale evidence approved: %v", e)
	}
	if _, e := Approve(p, "content", "reviewer", nil); e != nil {
		t.Fatal(e)
	}
	if _, e := Export(p, ExportOptions{Mode: "client", Out: filepath.Join(t.TempDir(), "client.zip")}); e == nil {
		t.Fatal("approval hid stale build dependencies")
	}
}
func TestForkPreservesAncestry(t *testing.T) {
	p := example(t)
	pin(t, p)
	m, e := Fork(p, "editorial-photo", "editorial-custom", "Allow deck-specific adjustments", []string{"local-composition"})
	if e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if p.Document.Slides[1].Template.ID != m.TemplateID {
		t.Fatal("slide not rebound")
	}
	fork := p.Document.LocalTemplates[m.TemplateID]
	if fork.Provenance.Parent.ID != "editorial-photo" || fork.Provenance.DefinitionSnapshot == "" {
		t.Fatal("ancestry not frozen")
	}
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	snapshot := filepath.Join(p.Root, fork.Provenance.DefinitionSnapshot)
	if e = os.Chmod(snapshot, 0644); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(snapshot, []byte("changed"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := Check(p, bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "ancestor snapshot hash") {
		t.Fatalf("bad ancestry accepted: %v", e)
	}
}
func TestDetachSourceScenePreservesAuthoredCopy(t *testing.T) {
	p := example(t)
	catalog, e := wmdesign.LibraryCatalog(bundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	var selected wmdesign.LibraryTemplate
	for _, d := range catalog {
		if d.Key == "quote/light" {
			selected = d
		}
	}
	if selected.Key == "" {
		t.Fatal("fixture not found")
	}
	values := map[string]any{}
	slots := map[string]any{}
	for _, s := range selected.Slots {
		switch s.Kind {
		case "string":
			if strings.HasSuffix(s.SourcePointer, "/photo") {
				slots[s.Name] = "photo-working-session"
			} else {
				slots[s.Name] = "Authored example"
			}
		case "number":
			slots[s.Name] = float64(1)
		case "boolean":
			slots[s.Name] = true
		}
	}
	keys := map[string]any{}
	for _, a := range selected.Arrays {
		ids := []any{}
		for i := 0; i < a.Count; i++ {
			ids = append(ids, "key-"+string(rune('a'+i)))
		}
		keys[a.Name] = ids
	}
	values["slots"] = slots
	values["keys"] = keys
	p.Document.Slides = []Slide{{ID: "authored-quote", ContentKind: "synthetic_example", Template: Reference{Scope: "shared", ID: selected.Key}, Values: values}}
	var tree any
	if e = json.Unmarshal(canonical(p.Document), &tree); e != nil {
		t.Fatal(e)
	}
	yamlBytes, e := yaml.Marshal(tree)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, yamlBytes, 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	if _, e = Detach(p, "authored-quote", "local-quote", bundle(t), wmdesign.CandidateEngine, "Tune geometry for this deck"); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if p.Document.Slides[0].Template.Scope != "local" {
		t.Fatal("not detached")
	}
	for _, v := range p.Document.Slides[0].Values {
		if s, ok := v.(string); ok && s != "Authored example" && s != "photo-working-session" {
			t.Fatalf("sample copy leaked into authored content: %q", s)
		}
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}
func TestProjectSVGAndUnsafeResource(t *testing.T) {
	p := example(t)
	pin(t, p)
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 48"><rect width="96" height="48" fill="#203F77"/></svg>`
	if e := os.WriteFile(filepath.Join(p.Root, "assets/originals/sample.svg"), []byte(svg), 0644); e != nil {
		t.Fatal(e)
	}
	p = rewrite(t, p, "assets/originals/sample.png", "assets/originals/sample.svg")
	p = rewrite(t, p, "fit: contain", "fit: contain\n            rotation_deg: 335")
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	raw := strings.Replace(string(p.Raw), "kind: image", "kind: component\n            definition: {scope: shared, id: wmds/component/logoslot}\n            arguments: {src: {binding: photo}, name: Client}", 1)
	raw = strings.Replace(raw, "            asset: {binding: photo}\n            fit: contain\n            rotation_deg: 335\n", "", 1)
	if e := os.WriteFile(p.SourcePath, []byte(raw), 0644); e != nil {
		t.Fatal(e)
	}
	var e error
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	bad := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 48"><image href="https://example.com/image.png"/></svg>`
	if e := os.WriteFile(filepath.Join(p.Root, "assets/originals/sample.svg"), []byte(bad), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := Compile(p, bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "unsupported SVG") {
		t.Fatalf("external SVG accepted: %v", e)
	}
}
func TestForkRejectsUnrelatedSlide(t *testing.T) {
	p := example(t)
	before := append([]byte{}, p.Raw...)
	if _, e := Fork(p, "editorial-photo", "unrelated", "bad target", []string{"maintain-the-source"}); e == nil {
		t.Fatal("fork overwrote unrelated shared slide")
	}
	after, e := os.ReadFile(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("rejected fork changed source")
	}
}
func TestOfflineRelocatedCalibration(t *testing.T) {
	p := example(t)
	path, e := filepath.Abs("../../library/wm-design-system/v3")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Pin(p, path, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: path, Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	zipPath := filepath.Join(t.TempDir(), "offline.zip")
	if _, e = Export(p, ExportOptions{Mode: "offline", Out: zipPath, Bundle: path}); e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(zipPath)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		t.Fatal(e)
	}
	relocated := t.TempDir()
	for _, f := range z.File {
		out, e := SafePath(relocated, f.Name)
		if e != nil {
			t.Fatal(e)
		}
		r, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		data, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			t.Fatal(e)
		}
		if e = writeExclusive(out, data, 0644); e != nil {
			t.Fatal(e)
		}
	}
	copy, e := Load(relocated)
	if e != nil {
		t.Fatal(e)
	}
	offlineBundle := filepath.Join(relocated, "runtime/library/wm-design-system/pinned")
	if _, e = Check(copy, offlineBundle, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e = Build(copy, BuildOptions{Bundle: offlineBundle, Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}
func TestSVGUnsupportedVisibleFeatures(t *testing.T) {
	for _, attr := range []string{`display="none"`, `visibility="hidden"`} {
		if _, err := assetMIME([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="10" height="10" ` + attr + `/></svg>`)); err == nil {
			t.Fatalf("unsupported SVG attribute accepted: %s", attr)
		}
	}
	if _, err := assetMIME([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><style>path {fill:#FFFFFF}</style><path d="M0 0H10V10Z"/></svg>`)); err == nil {
		t.Fatal("SVG selector stylesheet accepted")
	}
	if _, err := assetMIME([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><path xmlns:x="urn:foreign" d="M0 0H10V10H0Z" fill="#070154" x:fill="#FFFFFF"/></svg>`)); err == nil {
		t.Fatal("namespaced SVG fill accepted")
	}
	for _, body := range []string{`<foreign:rect xmlns:foreign="http://example.com" width="10" height="10" fill="#000000"/>`, `<rect width="10" height="10" fill="#000000" opacity="0.2"/>`, `<path d="M0 0H10V10Z" fill="#000000" stroke="#FF0000"/>`, `<path d="M0 0H10V10Z" fill="#000000" style="opacity:0.5"/>`, `<style>.hidden { opacity : 0.5; }</style><rect width="10" height="10" class="hidden"/>`} {
		raw := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">` + body + `</svg>`)
		if _, e := assetMIME(raw); e == nil {
			t.Fatalf("SVG feature silently accepted: %s", body)
		}
	}
}
func TestDottedStableIDs(t *testing.T) {
	values := map[string]any{"cards": []any{map[string]any{"key": "source.original", "title": "Source"}}}
	if got := slotPointer(values, "/slides/0/values", "cards.source.original.title"); got != "/slides/0/values/cards/0/title" {
		t.Fatalf("dotted card pointer: %s", got)
	}
	p := example(t)
	pin(t, p)
	p = rewrite(t, p, "id: maintain-the-source", "id: maintain.the.source")
	p = rewrite(t, p, "key: source", "key: source.original")
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}
