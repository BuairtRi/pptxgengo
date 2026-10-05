package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestSectionMutationCommentsAndAtomicity(t *testing.T) {
	p := example(t)
	raw := append([]byte("# preserved source comment\n"), p.Raw...)
	raw = bytes.Replace(raw, []byte("summary: |"), []byte("summary: | # preserved scalar comment"), 1)
	if e := os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	first, e := AddSection(p, SectionAddOptions{ID: "results", Title: "Results", BeforeSlideID: "local-composition"})
	if e != nil {
		t.Fatal(e)
	}
	if first.AddedOpeningID != "opening" || len(first.Sections) != 2 {
		t.Fatalf("Opening not explicit: %+v", first)
	}
	predecessor, e := os.ReadFile(filepath.Join(p.Root, "decisions/sources/"+first.BeforeSHA256+".yaml"))
	if e != nil || !bytes.Equal(predecessor, raw) {
		t.Fatal("predecessor changed", e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(p.Raw, []byte("# preserved source comment")) || !bytes.Contains(p.Raw, []byte("| # preserved scalar comment")) {
		t.Fatal("comments/scalar style lost")
	}
	before := append([]byte(nil), p.Raw...)
	for _, bad := range []SectionAddOptions{{ID: "bad", Title: "Bad", BeforeSlideID: "missing"}, {ID: "results", Title: "Another", BeforeSlideID: "local-composition"}, {ID: "new", Title: "RESULTS", BeforeSlideID: "local-composition"}, {ID: "bad", Title: "Bad", BeforeSlideID: "local-composition", DividerPhoto: "x"}} {
		if _, e = AddSection(p, bad); e == nil {
			t.Fatal("invalid add accepted", bad)
		}
		after, _ := os.ReadFile(p.SourcePath)
		if !bytes.Equal(before, after) {
			t.Fatal("failed mutation changed source")
		}
	}
	renamed, e := RenameSection(p, "results", "Recommendations")
	if e != nil {
		t.Fatal(e)
	}
	if renamed.Sections[1].Title != "Recommendations" {
		t.Fatal(renamed)
	}
	if _, e = RenameSection(p, "results", "Stale"); e == nil || !strings.Contains(e.Error(), "source changed") {
		t.Fatal("stale source accepted", e)
	}
	p, _ = Load(p.SourcePath)
	removed, e := RemoveSection(p, "opening")
	if e != nil {
		t.Fatal(e)
	}
	if len(removed.Sections) != 1 || removed.Sections[0].BeforeSlideID != "maintain-the-source" {
		t.Fatal("first removal left orphan prefix", removed)
	}
	p, _ = Load(p.SourcePath)
	if _, e = RemoveSection(p, "results"); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if len(p.Document.Sections) != 0 || len(p.Document.Slides) != 2 {
		t.Fatal("remove modified slides")
	}
}
func nativeZIP(t *testing.T, data []byte) map[string]string {
	t.Helper()
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		t.Fatal(e)
	}
	out := map[string]string{}
	for _, f := range z.File {
		r, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		b, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			t.Fatal(e)
		}
		out[f.Name] = string(b)
	}
	return out
}
func TestSectionNativeDeterminismHiddenAndNotes(t *testing.T) {
	p := example(t)
	source := string(p.Raw) + "\nsections:\n  - {id: author, title: 'Author & plan', before_slide_id: maintain-the-source}\n  - {id: local, title: Local content, before_slide_id: local-composition}\n"
	source = strings.Replace(source, "  - id: local-composition", "  - id: local-composition\n    hidden: true\n    notes: |\n      First exact speaker paragraph & details.\n\n      Second exact speaker paragraph.", 1)
	if e := os.WriteFile(p.SourcePath, []byte(source), 0644); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	c, e := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	first, report, e := wmdesign.BuildWithEngineAndAssets(bundle(t), "", c.Document, wmdesign.CandidateEngine, c.Assets)
	if e != nil {
		t.Fatal(e)
	}
	second, _, e := wmdesign.BuildWithEngineAndAssets(bundle(t), "", c.Document, wmdesign.CandidateEngine, c.Assets)
	if e != nil || !bytes.Equal(first, second) {
		t.Fatal("native sections nondeterministic", e)
	}
	if !report.Slides[1].Hidden || len(report.Sections) != 2 {
		t.Fatal("report lost metadata")
	}
	files := nativeZIP(t, first)
	presentation := files["ppt/presentation.xml"]
	assertNativeSectionMembers(t, presentation, []string{"Author & plan", "Local content"}, [][]string{{"256"}, {"257"}})
	for _, want := range []string{"sectionLst", "Author &amp; plan", "Local content"} {
		if !strings.Contains(presentation, want) {
			t.Errorf("missing native grouping %s", want)
		}
	}
	if strings.Contains(files["ppt/slides/slide1.xml"], `show="0"`) || !strings.Contains(files["ppt/slides/slide2.xml"], `show="0"`) {
		t.Fatal("hidden states wrong")
	}
	notes := files["ppt/notesSlides/notesSlide2.xml"]
	for _, want := range []string{"First exact speaker paragraph &amp; details.", "Second exact speaker paragraph.", "WMDS foundation reference; synthetic labels."} {
		if !strings.Contains(notes, want) {
			t.Errorf("notes lost %q", want)
		}
	}
	pin(t, p)
	deps, e := dependencies(p)
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := affectedKeys(p, "outline", nil, deps)["deck:sections"]; !ok {
		t.Fatal("sections absent approval dependencies")
	}
}
func TestSectionsStrictSource(t *testing.T) {
	cases := []string{
		"sections: [{id: one, title: One, before_slide_id: local-composition}]",
		"sections: [{id: one, title: One, before_slide_id: missing}]",
		"sections: [{id: one, title: One, before_slide_id: maintain-the-source}, {id: one, title: Two, before_slide_id: local-composition}]",
		"sections: [{id: one, title: One, before_slide_id: maintain-the-source}, {id: two, title: ONE, before_slide_id: local-composition}]",
		"sections: [{id: one, title: One, before_slide_id: maintain-the-source}, {id: two, title: Two, before_slide_id: maintain-the-source}]",
		"sections: [{id: one, title: ' padded ', before_slide_id: maintain-the-source}]",
	}
	for _, extra := range cases {
		t.Run(extra, func(t *testing.T) {
			p := example(t)
			if e := os.WriteFile(p.SourcePath, append(p.Raw, []byte("\n"+extra+"\n")...), 0644); e != nil {
				t.Fatal(e)
			}
			if _, e := Load(p.SourcePath); e == nil {
				t.Fatal("invalid sections accepted")
			}
		})
	}
	for _, field := range []string{"hidden: null", "hidden: 'true'", "notes: []", "notes: \"bad\\u0001\""} {
		t.Run(field, func(t *testing.T) {
			p := example(t)
			raw := bytes.Replace(p.Raw, []byte("  - id: local-composition"), []byte("  - id: local-composition\n    "+field), 1)
			os.WriteFile(p.SourcePath, raw, 0644)
			if _, e := Load(p.SourcePath); e == nil {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
}
func TestDividerClosedContractAndRemoval(t *testing.T) {
	if testing.Short() {
		t.Skip("divider rendering requires registered private photography; run make test-integration")
	}
	p := example(t)
	v5, e := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if e != nil {
		t.Fatal(e)
	}
	// Existing V2 public cards bindings also remain supported by V5.
	if _, e = Pin(p, v5, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	add, e := AddSection(p, SectionAddOptions{ID: "results", Title: "Results", BeforeSlideID: "local-composition", Divider: "divider/panel-edge", DividerPhoto: "photo-abstract-cubes", Bundle: v5, Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	if add.DividerSlideID != "results-divider" || len(add.Sections) != 2 || add.Sections[1].DividerTitle != "Results" {
		t.Fatal(add)
	}
	p, _ = Load(p.SourcePath)
	dividerBefore := canonical(p.Document.Slides[1])
	rename, e := RenameSection(p, "results", "Renamed native group")
	if e != nil {
		t.Fatal(e)
	}
	if rename.Sections[1].DividerTitle != "Results" {
		t.Fatal("rename silently changed visible divider")
	}
	p, _ = Load(p.SourcePath)
	if _, e = RemoveSection(p, "results"); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if len(p.Document.Slides) != 3 || !bytes.Equal(dividerBefore, canonical(p.Document.Slides[1])) {
		t.Fatal("removing group changed divider")
	}
	var slots map[string]any
	raw, _ := json.Marshal(p.Document.Slides[1].Values["slots"])
	json.Unmarshal(raw, &slots)
	if slots["node05.text"] != "Results" {
		t.Fatal("wrong actual slots", slots)
	}
	before := append([]byte(nil), p.Raw...)
	_, e = AddSection(p, SectionAddOptions{ID: "bad", Title: "Bad", BeforeSlideID: "local-composition", Divider: "divider/progress-list", DividerPhoto: "photo-abstract-cubes", Bundle: v5})
	if e == nil {
		t.Fatal("unknown automatic recipe accepted")
	}
	after, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, after) {
		t.Fatal("failed binding mutated source")
	}
}

func assertNativeSectionMembers(t *testing.T, source string, titles []string, members [][]string) {
	t.Helper()
	d := xml.NewDecoder(strings.NewReader(source))
	index := -1
	inSection := false
	got := [][]string{}
	for {
		token, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "section" {
				index++
				inSection = true
				got = append(got, []string{})
				title := ""
				for _, attr := range element.Attr {
					if attr.Name.Local == "name" {
						title = attr.Value
					}
				}
				if index >= len(titles) || title != titles[index] {
					t.Fatalf("native section title %q", title)
				}
			}
			if inSection && element.Name.Local == "sldId" {
				for _, attr := range element.Attr {
					if attr.Name.Local == "id" {
						got[index] = append(got[index], attr.Value)
					}
				}
			}
		case xml.EndElement:
			if element.Name.Local == "section" {
				inSection = false
			}
		}
	}
	if !bytes.Equal(canonical(got), canonical(members)) {
		t.Fatalf("section members got %v want %v", got, members)
	}
}
func TestAbsentSectionMetadataCompatibility(t *testing.T) {
	p := example(t)
	pin(t, p)
	deps, e := dependencies(p)
	if e != nil {
		t.Fatal(e)
	}
	c, e := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	first, _, e := wmdesign.BuildWithEngineAndAssets(bundle(t), "", c.Document, wmdesign.CandidateEngine, c.Assets)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(canonical(c.Document), []byte(`"sections"`)) || bytes.Contains(canonical(c.Document), []byte(`"notes"`)) || bytes.Contains(canonical(c.Document), []byte(`"hidden"`)) {
		t.Fatal("absent metadata changed historical scene shape")
	}
	for _, s := range p.Document.Slides {
		if deps["slide:"+s.ID+":content"] != digest(canonical(s.Values)) || deps["slide:"+s.ID+":selection"] != digest(canonical(s.Template)) {
			t.Fatal("legacy dependency changed")
		}
	}
	c.Document.Sections = []wmdesign.SectionSpec{}
	for i := range c.Document.Slides {
		c.Document.Slides[i].Hidden = false
		c.Document.Slides[i].Notes = ""
	}
	second, _, e := wmdesign.BuildWithEngineAndAssets(bundle(t), "", c.Document, wmdesign.CandidateEngine, c.Assets)
	if e != nil || !bytes.Equal(first, second) {
		t.Fatal("empty metadata changed native output", e)
	}
	files := nativeZIP(t, first)
	if strings.Contains(files["ppt/presentation.xml"], "sectionLst") || strings.Contains(files["ppt/slides/slide1.xml"], `show="0"`) {
		t.Fatal("legacy native metadata changed")
	}
}
