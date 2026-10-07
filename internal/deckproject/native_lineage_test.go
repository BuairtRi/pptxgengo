package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func lineageFixture(t *testing.T) ([]byte, Objects) {
	t.Helper()
	p := example(t)
	pin(t, p)
	r, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(p.Root, "builds", r.BuildID)
	data, e := os.ReadFile(filepath.Join(dir, "deck.pptx"))
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(dir, "object-map.json"))
	if e != nil {
		t.Fatal(e)
	}
	var o Objects
	if e = json.Unmarshal(raw, &o); e != nil {
		t.Fatal(e)
	}
	return data, o
}
func lineageEdit(t *testing.T, data []byte, part string, fn func([]byte) []byte) []byte {
	t.Helper()
	p, e := openLineagePackage(data)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := p.read(part)
	if e != nil {
		t.Fatal(e)
	}
	b, e := lineageRewrite(p, map[string][]byte{part: fn(raw)})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func lineageIssue(out NativeLineageInspection, kind string) bool {
	for _, i := range out.Issues {
		if i.Kind == kind {
			return true
		}
	}
	return false
}
func firstLineageShape(t *testing.T, b []byte) *lineageSpan {
	t.Helper()
	s, e := lineageSpans(b)
	if e != nil {
		t.Fatal(e)
	}
	var found *lineageSpan
	var walk func(*lineageSpan)
	walk = func(s *lineageSpan) {
		if found == nil && s.node.Name.Space == lineagePML && s.node.Name.Local == "sp" {
			found = s
		}
		for _, c := range s.children {
			walk(c)
		}
	}
	walk(s)
	if found == nil {
		t.Fatal("no native text shape")
	}
	return found
}

func TestNativeLineageBuildAndInspect(t *testing.T) {
	data, baseline := lineageFixture(t)
	out, e := InspectNativeLineage(data, baseline)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Issues) != 0 || len(out.Objects) != len(baseline.Objects) {
		t.Fatalf("baseline mismatch: objects=%d expected=%d issues=%+v", len(out.Objects), len(baseline.Objects), out.Issues)
	}
	for _, o := range out.Objects {
		if !validLineageToken(o.ShapeToken) || o.shape == nil {
			t.Fatalf("missing identity: %+v", o)
		}
	}
	// Renaming native names and physical IDs must not change tag-based matching.
	changed := lineageEdit(t, data, baseline.Objects[0].NativePart, func(b []byte) []byte {
		return []byte(strings.ReplaceAll(strings.ReplaceAll(string(b), `name="`, `name="renamed-`), `<p:cNvPr id="`, `<p:cNvPr id="8`))
	})
	out, e = InspectNativeLineage(changed, baseline)
	if e != nil || len(out.Issues) != 0 {
		t.Fatalf("physical identity used as match: %v %+v", e, out.Issues)
	}
	// Native visible text is irrelevant to identity matching.
	changed = lineageEdit(t, data, baseline.Objects[0].NativePart, func(b []byte) []byte { return []byte(strings.ReplaceAll(string(b), `<a:t>`, `<a:t>edited `)) })
	out, e = InspectNativeLineage(changed, baseline)
	if e != nil || len(out.Issues) != 0 {
		t.Fatalf("text used as match: %v %+v", e, out.Issues)
	}
}
func TestNativeLineageReorderedSlides(t *testing.T) {
	data, o := lineageFixture(t)
	changed := lineageEdit(t, data, "ppt/presentation.xml", func(b []byte) []byte {
		s, e := lineageSpans(b)
		if e != nil {
			t.Fatal(e)
		}
		var list *lineageSpan
		for _, c := range s.children {
			if c.node.Name.Local == "sldIdLst" {
				list = c
			}
		}
		if list == nil || len(list.children) < 2 {
			t.Fatal("fixture needs two slides")
		}
		a, z := list.children[0], list.children[1]
		replacement := string(b[z.start:z.end]) + string(b[a.end:z.start]) + string(b[a.start:a.end])
		v, e := lineageApply(b, []lineagePatch{{a.start, z.end, replacement}})
		if e != nil {
			t.Fatal(e)
		}
		return v
	})
	out, e := InspectNativeLineage(changed, o)
	if e != nil || len(out.Issues) != 0 {
		t.Fatalf("slide order used as identity: %v %+v", e, out.Issues)
	}
}
func TestNativeLineageMissingDuplicateAndUntagged(t *testing.T) {
	data, o := lineageFixture(t)
	part := o.Objects[0].NativePart
	for _, kind := range []string{"shape_missing", "shape_duplicated", "shape_untagged"} {
		t.Run(kind, func(t *testing.T) {
			changed := lineageEdit(t, data, part, func(b []byte) []byte {
				s := firstLineageShape(t, b)
				patch := lineagePatch{s.start, s.end, ""}
				switch kind {
				case "shape_duplicated":
					patch = lineagePatch{s.end, s.end, string(b[s.start:s.end])}
				case "shape_untagged":
					r := string(b[s.start:s.end])
					start := strings.Index(r, "<p:custDataLst")
					end := strings.Index(r, "</p:custDataLst>")
					if start < 0 || end < 0 {
						t.Fatal("missing shape tags")
					}
					r = r[:start] + r[end+len("</p:custDataLst>"):]
					patch = lineagePatch{s.start, s.end, r}
				}
				v, e := lineageApply(b, []lineagePatch{patch})
				if e != nil {
					t.Fatal(e)
				}
				return v
			})
			out, e := InspectNativeLineage(changed, o)
			if e != nil {
				t.Fatal(e)
			}
			if !lineageIssue(out, kind) {
				t.Fatalf("missing %s: %+v", kind, out.Issues)
			}
		})
	}
}
func TestNativeLineageWrongBaselineAndMalformedTags(t *testing.T) {
	data, o := lineageFixture(t)
	t.Run("old baseline", func(t *testing.T) {
		old := o
		old.Lineage = nil
		if _, e := InspectNativeLineage(data, old); e == nil {
			t.Fatal("old baseline accepted")
		}
	})
	for _, mutation := range []string{"wrong-build", "external-tag", "duplicate-tag", "duplicate-rel"} {
		t.Run(mutation, func(t *testing.T) {
			part := "ppt/tags/pptxgengo1.xml"
			if mutation == "external-tag" || mutation == "duplicate-rel" {
				part = "ppt/_rels/presentation.xml.rels"
			}
			changed := lineageEdit(t, data, part, func(b []byte) []byte {
				s := string(b)
				switch mutation {
				case "wrong-build":
					s = strings.ReplaceAll(s, o.Lineage.BuildToken, strings.Repeat("0", 64))
				case "external-tag":
					s = strings.Replace(s, `Id="pptxgengoTag1"`, `TargetMode="External" Id="pptxgengoTag1"`, 1)
				case "duplicate-tag":
					s = strings.Replace(s, "</p:tagLst>", `<p:tag name="pptxgengo_build" val="`+o.Lineage.BuildToken+`"/></p:tagLst>`, 1)
				case "duplicate-rel":
					tree, e := readLineageXML(b)
					if e != nil {
						t.Fatal(e)
					}
					for _, n := range tree.Children {
						if lineageAttr(n, "", "Id") == "pptxgengoTag1" {
							s = strings.Replace(s, "</Relationships>", `<Relationship xmlns="`+lineageOPC+`" Id="pptxgengoTag1" Type="`+lineageTagType+`" Target="tags/pptxgengo1.xml"/></Relationships>`, 1)
						}
					}
				}
				return []byte(s)
			})
			if _, e := InspectNativeLineage(changed, o); e == nil {
				t.Fatal("malformed lineage accepted")
			}
		})
	}
}
func TestNativeLineageStampPreservesVisibleXML(t *testing.T) {
	p := example(t)
	pin(t, p)
	c, e := Check(p, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	data, _, e := wmdesign.BuildWithEngineAndAssets(bundle(t), "", c.Document, wmdesign.CandidateEngine, c.Assets)
	if e != nil {
		t.Fatal(e)
	}
	o, e := ObjectMap(p, c.Document, data)
	if e != nil {
		t.Fatal(e)
	}
	_, lock, e := ReadLock(p)
	if e != nil {
		t.Fatal(e)
	}
	stamped, o, e := StampNativeLineage(data, o, digest(lock))
	if e != nil {
		t.Fatal(e)
	}
	out, e := InspectNativeLineage(stamped, o)
	if e != nil || len(out.Issues) != 0 {
		t.Fatalf("stamp/inspect: %v %+v", e, out.Issues)
	}
	before, _ := openLineagePackage(data)
	after, _ := openLineagePackage(stamped)
	for name := range before.files {
		if !strings.HasPrefix(name, "ppt/slides/slide") || !strings.HasSuffix(name, ".xml") {
			continue
		}
		original, _ := before.read(name)
		b, _ := after.read(name)
		root, e := lineageSpans(b)
		if e != nil {
			t.Fatal(e)
		}
		patches := []lineagePatch{}
		var walk func(*lineageSpan)
		walk = func(s *lineageSpan) {
			if s.node.Name.Space == lineagePML && s.node.Name.Local == "custDataLst" {
				patches = append(patches, lineagePatch{s.start, s.end, ""})
				return
			}
			for _, c := range s.children {
				walk(c)
			}
		}
		walk(root)
		restored, e := lineageApply(b, patches)
		if e != nil {
			t.Fatal(e)
		}
		restored = bytes.ReplaceAll(restored, []byte("<p:nvPr></p:nvPr>"), []byte("<p:nvPr/>"))
		original = bytes.ReplaceAll(original, []byte("<p:nvPr></p:nvPr>"), []byte("<p:nvPr/>"))
		if !bytes.Equal(original, restored) {
			t.Fatalf("visible native XML changed: %s", name)
		}
	}
	if _, _, e = StampNativeLineage(stamped, o, digest(lock)); e == nil {
		t.Fatal("double stamp accepted")
	}
}
func TestNativeLineagePackageBounds(t *testing.T) {
	for _, name := range []string{"../escape.xml", "/absolute.xml", "a\\b.xml", "a/../b.xml"} {
		t.Run(name, func(t *testing.T) {
			var b bytes.Buffer
			z := zip.NewWriter(&b)
			w, e := z.Create(name)
			if e != nil {
				t.Fatal(e)
			}
			fmt.Fprint(w, "x")
			if e = z.Close(); e != nil {
				t.Fatal(e)
			}
			if _, e = openLineagePackage(b.Bytes()); e == nil {
				t.Fatal("unsafe part accepted")
			}
		})
	}
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	for i := 0; i < 2; i++ {
		w, _ := z.Create("duplicate.xml")
		fmt.Fprint(w, "x")
	}
	z.Close()
	if _, e := openLineagePackage(b.Bytes()); e == nil {
		t.Fatal("duplicate part accepted")
	}
	for _, xml := range []string{"<!DOCTYPE root><root/>", "<one/><two/>", strings.Repeat("<x>", 101) + strings.Repeat("</x>", 101)} {
		if _, e := readLineageXML([]byte(xml)); e == nil {
			t.Fatal("unsafe XML accepted")
		}
	}
}

// Explicit desktop evidence hooks. Normal short/race lanes never open Office.
func TestNativeLineageDesktopFixture(t *testing.T) {
	dir := os.Getenv("PPTXGENGO_NATIVE_LINEAGE_FIXTURE_OUT")
	if dir == "" {
		t.Skip("desktop fixture creation is opt-in")
	}
	if e := os.Mkdir(dir, 0700); e != nil {
		t.Fatal(e)
	}
	data, o := lineageFixture(t)
	raw := canonical(o)
	for name, b := range map[string][]byte{"baseline.pptx": data, "object-map.json": raw, "fixture-integrity.json": canonical(map[string]string{"schema": "pptxgengo.native-lineage-fixture.v1", "baseline_sha256": digest(data), "object_map_sha256": digest(raw)})} {
		if e := writeExclusive(filepath.Join(dir, name), b, 0444); e != nil {
			t.Fatal(e)
		}
	}
}
func TestNativeLineageDesktopEdited(t *testing.T) {
	dir := os.Getenv("PPTXGENGO_NATIVE_LINEAGE_FIXTURE")
	edited := os.Getenv("PPTXGENGO_NATIVE_LINEAGE_EDITED")
	if dir == "" || edited == "" {
		t.Skip("desktop edited-deck verification is opt-in")
	}
	read := func(p string) []byte {
		t.Helper()
		st, e := os.Lstat(p)
		if e != nil || !st.Mode().IsRegular() || st.Size() > lineageMaxPackage {
			t.Fatalf("invalid desktop fixture file: %s %v", p, e)
		}
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	baseline := read(filepath.Join(dir, "baseline.pptx"))
	raw := read(filepath.Join(dir, "object-map.json"))
	var integrity map[string]string
	if e := json.Unmarshal(read(filepath.Join(dir, "fixture-integrity.json")), &integrity); e != nil {
		t.Fatal(e)
	}
	if integrity["schema"] != "pptxgengo.native-lineage-fixture.v1" || integrity["baseline_sha256"] != digest(baseline) || integrity["object_map_sha256"] != digest(raw) {
		t.Fatal("desktop baseline drift")
	}
	var o Objects
	if e := json.Unmarshal(raw, &o); e != nil {
		t.Fatal(e)
	}
	data := read(edited)
	out, e := InspectNativeLineage(data, o)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Issues) != 0 || len(out.Objects) != len(o.Objects) {
		t.Fatalf("desktop lineage did not survive: %+v", out.Issues)
	}
	evidence := canonical(map[string]any{"schema": "pptxgengo.native-lineage-desktop-inspection.v1", "baseline_sha256": digest(baseline), "edited_sha256": digest(data), "object_map_sha256": digest(raw), "scope": "identity survival only; not visual, editing ergonomics, text adoption, or platform-wide qualification", "inspection": out})
	evidencePath := edited + ".lineage.json"
	if e = writeExclusive(evidencePath, evidence, 0444); e != nil {
		t.Fatal(e)
	}
	t.Logf("desktop identity evidence: %s", evidencePath)
}
