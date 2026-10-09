package deckproject

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func nativeChartFixture(t *testing.T, kind string) (*Project, *TextBaseline) {
	t.Helper()
	p, _ := quantitativeFixture(t, kind)
	if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, QuantitativeOperation{Action: "set", Entity: "setting", Field: "preserveWorkbookZeros", Setting: true}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	return p, b
}
func chartEditValue(t *testing.T, data []byte, cache, workbook *string) []byte {
	t.Helper()
	pkg, e := openLineagePackage(data)
	if e != nil {
		t.Fatal(e)
	}
	chart := ""
	book := ""
	for name := range pkg.files {
		if strings.HasPrefix(name, "ppt/charts/") && strings.HasSuffix(name, ".xml") {
			chart = name
		}
		if strings.HasPrefix(name, "ppt/embeddings/") && strings.HasSuffix(name, ".xlsx") {
			book = name
		}
	}
	if chart == "" || book == "" {
		t.Fatal("fixture chart/workbook absent")
	}
	changes := map[string][]byte{}
	if cache != nil {
		raw, e := pkg.read(chart)
		if e != nil {
			t.Fatal(e)
		}
		re := regexp.MustCompile(`(<c:numCache>.*?<c:pt idx="0"><c:v>)[^<]*(</c:v></c:pt>)`)
		if !re.Match(raw) {
			t.Fatal("cache sample missing")
		}
		raw = re.ReplaceAll(raw, []byte("${1}"+*cache+"${2}"))
		changes[chart] = raw
	}
	if workbook != nil {
		raw, e := pkg.read(book)
		if e != nil {
			t.Fatal(e)
		}
		wb, e := openLineagePackage(raw)
		if e != nil {
			t.Fatal(e)
		}
		sheet, e := wb.read("xl/worksheets/sheet1.xml")
		if e != nil {
			t.Fatal(e)
		}
		re := regexp.MustCompile(`(<c r="B2"><v>)[^<]*(</v></c>)`)
		if !re.Match(sheet) {
			t.Fatal("workbook sample missing")
		}
		sheet = re.ReplaceAll(sheet, []byte("${1}"+*workbook+"${2}"))
		raw, e = lineageRewrite(wb, map[string][]byte{"xl/worksheets/sheet1.xml": sheet})
		if e != nil {
			t.Fatal(e)
		}
		changes[book] = raw
	}
	edited, e := lineageRewrite(pkg, changes)
	if e != nil {
		t.Fatal(e)
	}
	return edited
}
func chartSemanticPacket(t *testing.T, p *Project, b *TextBaseline, edited []byte) *TextReviewPacket {
	t.Helper()
	packet, e := WriteGeometryReviewPacket(p, b, edited, filepath.Join(t.TempDir(), "review"), bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	return packet
}
func TestQuantitativeNativeChartCacheWorkbookAdoption(t *testing.T) {
	for _, kind := range []string{"column", "bar", "line", "pie", "doughnut"} {
		t.Run(kind, func(t *testing.T) {
			p, b := nativeChartFixture(t, kind)
			v := "1.25"
			edited := chartEditValue(t, b.files["deck.pptx"], &v, &v)
			packet := chartSemanticPacket(t, p, b, edited)
			report, e := ProposeQuantitativeSemantics(p, packet, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if len(report.Proposals) != 1 || report.Proposals[0].Category != "current" || report.Proposals[0].Series != "capacity" || *report.Proposals[0].Before != 0 || *report.Proposals[0].Proposed != 1.25 {
				t.Fatalf("wrong source-keyed proposal %+v", report)
			}
			d := GanttSemanticDecisions{Schema: QuantitativeSemanticDecisionsSchema, ReportSHA256: QuantitativeSemanticReportHash(report), Actor: "Operator", Reason: "Confirm explicit native data edit", Decisions: []GanttSemanticDecision{{ProposalID: report.Proposals[0].ID, Action: "set_value", Reason: "Use edited workbook fact"}}}
			before := append([]byte(nil), p.Raw...)
			if out, e := AdoptQuantitativeSemantics(p, packet, "quantitative-slide", "chart", canonical(d), bundle(t), wmdesign.CandidateEngine, false); e != nil || out.Applied {
				t.Fatalf("preview %v", e)
			}
			disk, _ := os.ReadFile(p.SourcePath)
			if !bytes.Equal(before, disk) {
				t.Fatal("preview changed source")
			}
			if out, e := AdoptQuantitativeSemantics(p, packet, "quantitative-slide", "chart", canonical(d), bundle(t), wmdesign.CandidateEngine, true); e != nil || !out.Applied {
				t.Fatalf("apply %v", e)
			}
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			inspect, e := InspectQuantitative(p, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine)
			if e != nil || inspect.RenderError != "" {
				t.Fatalf("rebuild inspect %v %s", e, inspect.RenderError)
			}
			if inspect.Source["series"].([]any)[0].(map[string]any)["values"].([]any)[0] != 1.25 || inspect.Source["units"] != "FTE" {
				t.Fatal("facts/units lost")
			}
			if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestQuantitativeNativeChartContradictoryCacheWorkbookRefused(t *testing.T) {
	p, b := nativeChartFixture(t, "column")
	a, z := "1.25", "9"
	for _, tc := range []struct{ cache, book *string }{{&a, &z}, {&a, nil}, {nil, &z}} {
		edited := chartEditValue(t, b.files["deck.pptx"], tc.cache, tc.book)
		packet := chartSemanticPacket(t, p, b, edited)
		if _, e := ProposeQuantitativeSemantics(p, packet, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine); e == nil {
			t.Fatal("contradictory workbook/cache accepted")
		}
	}
}
func TestQuantitativeNativeChartMalformedSourcesAndDecisionHashes(t *testing.T) {
	p, b := nativeChartFixture(t, "column")
	v := "2"
	edited := chartEditValue(t, b.files["deck.pptx"], &v, &v)
	packet := chartSemanticPacket(t, p, b, edited)
	report, e := ProposeQuantitativeSemantics(p, packet, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	bad := GanttSemanticDecisions{Schema: QuantitativeSemanticDecisionsSchema, ReportSHA256: strings.Repeat("0", 64), Actor: "Operator", Reason: "Bad report", Decisions: []GanttSemanticDecision{{ProposalID: report.Proposals[0].ID, Action: "set_value", Reason: "Bad"}}}
	if _, e := AdoptQuantitativeSemantics(p, packet, "quantitative-slide", "chart", canonical(bad), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("wrong report accepted")
	}
	bad.ReportSHA256 = QuantitativeSemanticReportHash(report)
	bad.Decisions[0].ProposalID = strings.Repeat("1", 64)
	if _, e := AdoptQuantitativeSemantics(p, packet, "quantitative-slide", "chart", canonical(bad), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("unknown proposal accepted")
	}
	for _, formula := range []string{"External.xlsx!$B$2:$B$3", "Sheet1!$B$2:$C$3", "Sheet1!$B$2:$B$4"} {
		if _, e = chartReferences(formula, 2); e == nil {
			t.Fatalf("invalid range %s", formula)
		}
	}
	if _, e = chartReferences("'Sheet1'!$B$2:$B$3", 2); e != nil {
		t.Fatal(e)
	}
	_ = fmt.Sprint(report)
}

func TestQuantitativeNativeChartMissingAndScatterSourceFacts(t *testing.T) {
	for _, kind := range []string{"line", "column", "scatter"} {
		t.Run(kind, func(t *testing.T) {
			p, b := nativeChartFixture(t, kind)
			if kind == "column" {
				if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, QuantitativeOperation{Action: "set", Entity: "setting", Field: "allowMissing", Setting: true}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
					t.Fatal(e)
				}
				var e error
				p, e = Load(p.SourcePath)
				if e != nil {
					t.Fatal(e)
				}
				if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
					t.Fatal(e)
				}
				b, e = ReadTextBaseline(p, "", "")
				if e != nil {
					t.Fatal(e)
				}
			}
			pkg, e := openLineagePackage(b.files["deck.pptx"])
			if e != nil {
				t.Fatal(e)
			}
			chart, book := "", ""
			for part := range pkg.files {
				if strings.HasPrefix(part, "ppt/charts/") && strings.HasSuffix(part, ".xml") {
					chart = part
				}
				if strings.HasSuffix(part, ".xlsx") {
					book = part
				}
			}
			raw, e := pkg.read(chart)
			if e != nil {
				t.Fatal(e)
			}
			cached := regexp.MustCompile(`(?s)<c:numCache>.*?</c:numCache>`)
			matches := cached.FindAllIndex(raw, -1)
			if len(matches) == 0 {
				t.Fatal("cache absent")
			}
			section := append([]byte(nil), raw[matches[0][0]:matches[0][1]]...)
			cell := "B3"
			if kind == "scatter" {
				section = regexp.MustCompile(`(<c:pt idx="0"><c:v>)[^<]*(</c:v></c:pt>)`).ReplaceAll(section, []byte("${1}1.5${2}"))
				cell = "A2"
			} else {
				section = regexp.MustCompile(`<c:pt idx="1"><c:v>[^<]*</c:v></c:pt>`).ReplaceAll(section, nil)
			}
			raw = append(append(append([]byte(nil), raw[:matches[0][0]]...), section...), raw[matches[0][1]:]...)
			wbraw, e := pkg.read(book)
			if e != nil {
				t.Fatal(e)
			}
			wb, e := openLineagePackage(wbraw)
			if e != nil {
				t.Fatal(e)
			}
			sheet, e := wb.read("xl/worksheets/sheet1.xml")
			if e != nil {
				t.Fatal(e)
			}
			cellRE := regexp.MustCompile(`<c r="` + cell + `"><v>[^<]*</v></c>`)
			if !cellRE.Match(sheet) {
				t.Fatal("numeric cell absent")
			}
			var replacement []byte
			if kind == "scatter" {
				replacement = []byte(`<c r="A2"><v>1.5</v></c>`)
			}
			sheet = cellRE.ReplaceAll(sheet, replacement)
			wbraw, e = lineageRewrite(wb, map[string][]byte{"xl/worksheets/sheet1.xml": sheet})
			if e != nil {
				t.Fatal(e)
			}
			edited, e := lineageRewrite(pkg, map[string][]byte{chart: raw, book: wbraw})
			if e != nil {
				t.Fatal(e)
			}
			packet := chartSemanticPacket(t, p, b, edited)
			r, e := ProposeQuantitativeSemantics(p, packet, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if len(r.Proposals) != 1 {
				t.Fatalf("proposals %+v", r)
			}
			q := r.Proposals[0]
			if kind == "scatter" {
				if q.Axis != "x" || q.Point != "one" || *q.Proposed != 1.5 {
					t.Fatal("scatter wrong mapping")
				}
			} else if q.Proposed != nil || q.Category != "future" {
				t.Fatal("missing not distinct")
			}
			d := GanttSemanticDecisions{Schema: QuantitativeSemanticDecisionsSchema, ReportSHA256: QuantitativeSemanticReportHash(r), Actor: "Operator", Reason: "Confirm explicit missing/point update", Decisions: []GanttSemanticDecision{{ProposalID: q.ID, Action: "set_value", Reason: "Native data is intended"}}}
			if _, e = AdoptQuantitativeSemantics(p, packet, "quantitative-slide", "chart", canonical(d), bundle(t), wmdesign.CandidateEngine, true); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestQuantitativeNativeChartMetadataAndUnknownWorkbookFactsRefused(t *testing.T) {
	p, b := nativeChartFixture(t, "column")
	for _, tc := range []string{"overflow", "category", "formula", "foreign-sheet"} {
		t.Run(tc, func(t *testing.T) {
			v := "2"
			if tc == "overflow" {
				v = "1000000000000"
			}
			edited := chartEditValue(t, b.files["deck.pptx"], &v, &v)
			pkg, e := openLineagePackage(edited)
			if e != nil {
				t.Fatal(e)
			}
			changes := map[string][]byte{}
			for name := range pkg.files {
				if strings.HasPrefix(name, "ppt/charts/") && strings.HasSuffix(name, ".xml") && (tc == "category" || tc == "foreign-sheet") {
					raw, e := pkg.read(name)
					if e != nil {
						t.Fatal(e)
					}
					if tc == "category" {
						raw = bytes.ReplaceAll(raw, []byte("CURRENT"), []byte("OTHER"))
					} else {
						raw = bytes.ReplaceAll(raw, []byte("Sheet1!"), []byte("Sheet2!"))
					}
					changes[name] = raw
				}
				if strings.HasPrefix(name, "ppt/embeddings/") && strings.HasSuffix(name, ".xlsx") && tc == "formula" {
					raw, e := pkg.read(name)
					if e != nil {
						t.Fatal(e)
					}
					wb, e := openLineagePackage(raw)
					if e != nil {
						t.Fatal(e)
					}
					sheet, e := wb.read("xl/worksheets/sheet1.xml")
					if e != nil {
						t.Fatal(e)
					}
					sheet = bytes.Replace(sheet, []byte(`<c r="B2"><v>`), []byte(`<c r="B2"><f>1+1</f><v>`), 1)
					raw, e = lineageRewrite(wb, map[string][]byte{"xl/worksheets/sheet1.xml": sheet})
					if e != nil {
						t.Fatal(e)
					}
					changes[name] = raw
				}
			}
			edited, e = lineageRewrite(pkg, changes)
			if e != nil {
				t.Fatal(e)
			}
			packet := chartSemanticPacket(t, p, b, edited)
			before, _ := os.ReadFile(p.SourcePath)
			if _, e = ProposeQuantitativeSemantics(p, packet, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine); e == nil {
				t.Fatalf("unsupported edit %s accepted", tc)
			}
			after, _ := os.ReadFile(p.SourcePath)
			if !bytes.Equal(before, after) {
				t.Fatal("unsupported edit changed source")
			}
		})
	}
}

func TestQuantitativeNativeChartOfficeSerializationNumericOnly(t *testing.T) {
	p, baseline := nativeChartFixture(t, "column")
	v := "2"
	edited := chartEditValue(t, baseline.files["deck.pptx"], &v, &v)
	pkg, e := openLineagePackage(edited)
	if e != nil {
		t.Fatal(e)
	}
	changes := map[string][]byte{}
	for name := range pkg.files {
		raw, e := pkg.read(name)
		if e != nil {
			t.Fatal(e)
		}
		if strings.HasPrefix(name, "ppt/charts/") && strings.HasSuffix(name, ".xml") {
			// These changes are retained as manual formatting, not declared equal.
			raw = bytes.Replace(raw, []byte(`<c:date1904 val="0"/>`), []byte(`<c:date1904 val="0"/><c:style val="2"/>`), 1)
			raw = bytes.ReplaceAll(raw, []byte(`multiLvlStrRef`), []byte(`strRef`))
			raw = bytes.ReplaceAll(raw, []byte(`multiLvlStrCache`), []byte(`strCache`))
			raw = bytes.ReplaceAll(raw, []byte(`<c:lvl>`), nil)
			raw = bytes.ReplaceAll(raw, []byte(`</c:lvl>`), nil)
			changes[name] = raw
		}
		if strings.HasPrefix(name, "ppt/embeddings/") && strings.HasSuffix(name, ".xlsx") {
			book, e := openLineagePackage(raw)
			if e != nil {
				t.Fatal(e)
			}
			sheet, e := book.read("xl/worksheets/sheet1.xml")
			if e != nil {
				t.Fatal(e)
			}
			sheet = bytes.ReplaceAll(sheet, []byte(`activeCell="B1" sqref="B1"`), []byte(`activeCell="B3" sqref="B3"`))
			sheet = bytes.ReplaceAll(sheet, []byte(`defaultRowHeight="16"`), []byte(`defaultRowHeight="13"`))
			raw, e = lineageRewrite(book, map[string][]byte{"xl/worksheets/sheet1.xml": sheet})
			if e != nil {
				t.Fatal(e)
			}
			changes[name] = raw
		}
	}
	edited, e = lineageRewrite(pkg, changes)
	if e != nil {
		t.Fatal(e)
	}
	packet := chartSemanticPacket(t, p, baseline, edited)
	report, e := ProposeQuantitativeSemantics(p, packet, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(report.Proposals) != 1 || *report.Proposals[0].Proposed != 2 {
		t.Fatalf("wrong facts %+v", report)
	}
	manual := false
	for _, i := range report.ManualReview {
		if i.Kind == "chart_formatting_not_adopted" {
			manual = true
		}
	}
	if !manual {
		t.Fatal("formatting falsely declared equivalent")
	}
	d := GanttSemanticDecisions{Schema: QuantitativeSemanticDecisionsSchema, ReportSHA256: QuantitativeSemanticReportHash(report), Actor: "Operator", Reason: "Adopt numeric fact only; retain reviewed formatting findings", Decisions: []GanttSemanticDecision{{ProposalID: report.Proposals[0].ID, Action: "set_value", Reason: "Cache and workbook agree"}}}
	if _, e = AdoptQuantitativeSemantics(p, packet, "quantitative-slide", "chart", canonical(d), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	z, e := openLineagePackage(b.files["deck.pptx"])
	if e != nil {
		t.Fatal(e)
	}
	for n := range z.files {
		if strings.HasPrefix(n, "ppt/charts/") && strings.HasSuffix(n, ".xml") {
			raw, e := z.read(n)
			if e != nil {
				t.Fatal(e)
			}
			if bytes.Contains(raw, []byte(`<c:style val="2"/>`)) {
				t.Fatal("manual Office style adopted into source")
			}
		}
	}
}

func TestQuantitativeNativeChartExtraWorkbookCellRefused(t *testing.T) {
	p, b := nativeChartFixture(t, "column")
	v := "2"
	edited := chartEditValue(t, b.files["deck.pptx"], &v, &v)
	pkg, e := openLineagePackage(edited)
	if e != nil {
		t.Fatal(e)
	}
	changes := map[string][]byte{}
	for n := range pkg.files {
		raw, e := pkg.read(n)
		if e != nil {
			t.Fatal(e)
		}
		if strings.HasPrefix(n, "ppt/embeddings/") && strings.HasSuffix(n, ".xlsx") {
			book, e := openLineagePackage(raw)
			if e != nil {
				t.Fatal(e)
			}
			sheet, e := book.read("xl/worksheets/sheet1.xml")
			if e != nil {
				t.Fatal(e)
			}
			sheet = bytes.Replace(sheet, []byte(`</sheetData>`), []byte(`<row r="12"><c r="Z12"><v>42</v></c></row></sheetData>`), 1)
			raw, e = lineageRewrite(book, map[string][]byte{"xl/worksheets/sheet1.xml": sheet})
			if e != nil {
				t.Fatal(e)
			}
			changes[n] = raw
		}
	}
	edited, e = lineageRewrite(pkg, changes)
	if e != nil {
		t.Fatal(e)
	}
	packet := chartSemanticPacket(t, p, b, edited)
	if _, e = ProposeQuantitativeSemantics(p, packet, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "extra workbook cell") {
		t.Fatalf("extra cell accepted %v", e)
	}
}

func TestQuantitativeAutoUpdateWorkbookPatchBinding(t *testing.T) {
	for _, flag := range []bool{false, true} {
		p, _ := nativeChartFixture(t, "column")
		patch := quantitativePatch(p, QuantitativeOperation{Action: "set", Entity: "setting", Field: "autoUpdateWorkbook", Setting: flag})
		if _, e := PatchQuantitative(p, "quantitative-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
			t.Fatal(e)
		}
		p, e := Load(p.SourcePath)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
			t.Fatal(e)
		}
		b, e := ReadTextBaseline(p, "", "")
		if e != nil {
			t.Fatal(e)
		}
		lineage, e := InspectNativeLineage(b.files["deck.pptx"], b.Objects)
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, o := range lineage.Objects {
			if o.NativeName == "chart.native" {
				facts, e := readNativeChartFacts(b.files["deck.pptx"], o, 2, 1, false)
				if e != nil {
					t.Fatal(e)
				}
				found = true
				value := "0"
				if flag {
					value = "1"
				}
				external := chartNodes(facts.tree, "externalData")
				if len(external) != 1 || attr(chartChild(external[0], "autoUpdate"), "val") != value {
					t.Fatal("source setting lost in native package")
				}
				if *facts.data[0][0] != 0 || *facts.data[0][1] != 3 {
					t.Fatal("flag changed cache/workbook facts")
				}
			}
		}
		if !found {
			t.Fatal("native owned chart missing")
		}
	}
}
