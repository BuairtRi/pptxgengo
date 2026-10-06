package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// This maintained opt-in probe reproduces publication's rebuild without
// modifying the accepted native artifacts.
func TestWritePublicationWorkbookProbe(t *testing.T) {
	out := os.Getenv("WMDS_PUBLICATION_WORKBOOK_PROBE_OUT")
	if out == "" {
		t.Skip("set WMDS_PUBLICATION_WORKBOOK_PROBE_OUT, _DOC and _ACCEPTED")
	}
	raw, err := os.ReadFile(os.Getenv("WMDS_PUBLICATION_WORKBOOK_PROBE_DOC"))
	if err != nil {
		t.Fatal(err)
	}
	var doc Document
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	rebuilt, _, err := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := os.ReadFile(os.Getenv("WMDS_PUBLICATION_WORKBOOK_PROBE_ACCEPTED"))
	if err != nil {
		t.Fatal(err)
	}
	left, err := publicationDeckParts(accepted)
	if err != nil {
		t.Fatal(err)
	}
	right, err := publicationDeckParts(rebuilt)
	if err != nil {
		t.Fatal(err)
	}
	type diff struct {
		Part    string   `json:"part"`
		Changed []string `json:"changed_inner_parts"`
	}
	var diffs []diff
	for name, data := range left {
		if !strings.HasSuffix(name, ".xlsx") {
			continue
		}
		lp, err := publicationDeckParts(data)
		if err != nil {
			t.Fatal(err)
		}
		rp, err := publicationDeckParts(right[name])
		if err != nil {
			t.Fatal(err)
		}
		d := diff{Part: name}
		for part, value := range lp {
			if !bytes.Equal(value, rp[part]) {
				d.Changed = append(d.Changed, part)
			}
		}
		for part := range rp {
			if _, ok := lp[part]; !ok {
				d.Changed = append(d.Changed, "added:"+part)
			}
		}
		diffs = append(diffs, d)
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "rebuilt.pptx"), rebuilt, 0644); err != nil {
		t.Fatal(err)
	}
	raw, err = json.MarshalIndent(diffs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "inner-part-diffs.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
	if err := comparePublicationVisibleParts(accepted, rebuilt); err != nil {
		t.Fatal(err)
	}
}

func publicationWorkbookFixture(t *testing.T, parts map[string]string, varied bool, duplicate string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	names := make([]string, 0, len(parts))
	for name := range parts {
		names = append(names, name)
	}
	sort.Strings(names)
	if varied {
		sort.Sort(sort.Reverse(sort.StringSlice(names)))
		if err := archive.SetComment("different ZIP container metadata"); err != nil {
			t.Fatal(err)
		}
	}
	if duplicate != "" {
		names = append(names, duplicate)
	}
	for _, name := range names {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		if varied {
			header.Method = zip.Store
			header.SetModTime(time.Date(2026, 10, 6, 12, 34, 56, 0, time.UTC))
			header.Comment = "entry metadata"
		}
		stream, err := archive.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = stream.Write([]byte(parts[name])); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestPublicationWorkbookTimestampAndContainerNormalization(t *testing.T) {
	core := `<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:creator>Accepted</dc:creator><dcterms:created xsi:type="dcterms:W3CDTF">2026-10-06T16:54:56.352Z</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">2026-10-06T16:54:56.352Z</dcterms:modified></cp:coreProperties>`
	parts := map[string]string{"[Content_Types].xml": "<Types/>", "docProps/core.xml": core, "xl/worksheets/sheet1.xml": "<worksheet><v>42</v></worksheet>", "xl/styles.xml": "<styles/>", "xl/sharedStrings.xml": "<strings>Accepted</strings>", "xl/_rels/workbook.xml.rels": "<Relationships/>"}
	clone := func() map[string]string {
		result := map[string]string{}
		for k, v := range parts {
			result[k] = v
		}
		return result
	}
	deck := func(workbook []byte) []byte {
		return publicationFixtureZip(t, map[string]string{
			"ppt/slides/slide1.xml":            "<sld/>",
			"ppt/slides/_rels/slide1.xml.rels": `<Relationships><Relationship Id="rId1" Type="example/chart" Target="../charts/chart1.xml"/></Relationships>`,
			"ppt/charts/chart1.xml":            "<chart/>",
			"ppt/charts/_rels/chart1.xml.rels": `<Relationships><Relationship Id="rId1" Type="example/package" Target="../embeddings/workbook.xlsx"/></Relationships>`,
			"ppt/embeddings/workbook.xlsx":     string(workbook),
		})
	}
	original := deck(publicationWorkbookFixture(t, parts, false, ""))
	for _, timestamps := range []bool{false, true} {
		changed := clone()
		if timestamps {
			changed["docProps/core.xml"] = strings.ReplaceAll(core, "2026-10-06T16:54:56.352Z", "2026-10-06T18:20:10.001Z")
		}
		if err := comparePublicationVisibleParts(original, deck(publicationWorkbookFixture(t, changed, true, ""))); err != nil {
			t.Fatalf("nonvisual metadata rejected: %v", err)
		}
	}
	cases := []struct {
		name   string
		change func(map[string]string)
	}{
		{"data", func(p map[string]string) { p["xl/worksheets/sheet1.xml"] = "<worksheet><v>43</v></worksheet>" }},
		{"styles", func(p map[string]string) { p["xl/styles.xml"] = "<styles changed='true'/>" }},
		{"strings", func(p map[string]string) { p["xl/sharedStrings.xml"] = "<strings>Changed</strings>" }},
		{"relationships", func(p map[string]string) { p["xl/_rels/workbook.xml.rels"] = "<Relationships changed='true'/>" }},
		{"content types", func(p map[string]string) { p["[Content_Types].xml"] = "<Types changed='true'/>" }},
		{"missing part", func(p map[string]string) { delete(p, "xl/styles.xml") }},
		{"added part", func(p map[string]string) { p["xl/extra.xml"] = "<extra/>" }},
		{"metadata creator", func(p map[string]string) {
			p["docProps/core.xml"] = strings.ReplaceAll(core, ">Accepted<", ">Changed<")
		}},
		{"timestamp attributes", func(p map[string]string) {
			p["docProps/core.xml"] = strings.ReplaceAll(core, "dcterms:W3CDTF", "dcterms:Changed")
		}},
		{"timestamp whitespace", func(p map[string]string) {
			p["docProps/core.xml"] = strings.ReplaceAll(core, "2026-10-06T16:54:56.352Z", " 2026-10-06T16:54:56.352Z ")
		}},
		{"timestamp namespace", func(p map[string]string) {
			p["docProps/core.xml"] = strings.ReplaceAll(strings.ReplaceAll(core, "http://purl.org/dc/terms/", "urn:other"), "2026-10-06T16:54:56.352Z", "2026-10-06T18:20:10.001Z")
		}},
		{"timestamp syntax", func(p map[string]string) {
			p["docProps/core.xml"] = strings.ReplaceAll(core, "2026-10-06T16:54:56.352Z", "<![CDATA[2026-10-06T18:20:10.001Z]]>")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := clone()
			tc.change(p)
			if err := comparePublicationVisibleParts(original, deck(publicationWorkbookFixture(t, p, true, ""))); err == nil {
				t.Fatal("meaningful or unverified workbook change accepted")
			}
		})
	}
	for _, name := range []string{"xl/styles.xml", "docProps/core.xml"} {
		duplicate := deck(publicationWorkbookFixture(t, parts, false, name))
		if err := comparePublicationVisibleParts(original, duplicate); err == nil {
			t.Fatalf("duplicate part %s accepted", name)
		}
		if err := comparePublicationVisibleParts(duplicate, duplicate); err == nil {
			t.Fatalf("identical duplicate archives accepted for %s", name)
		}
	}
	if err := comparePublicationVisibleParts(original, deck([]byte("not a workbook"))); err == nil {
		t.Fatal("malformed workbook accepted")
	}
	chartParts, err := publicationDeckParts(original)
	if err != nil {
		t.Fatal(err)
	}
	changedChart := map[string]string{}
	for name, raw := range chartParts {
		changedChart[name] = string(raw)
	}
	changedChart["ppt/charts/chart1.xml"] = "<chart changed='true'/>"
	if err := comparePublicationVisibleParts(original, publicationFixtureZip(t, changedChart)); err == nil {
		t.Fatal("changed chart XML accepted")
	}
}
