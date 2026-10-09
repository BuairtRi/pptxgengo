package deckproject

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Opt-in private evidence exercise. The first packet is an actual macOS Office
// Save As with B2=15 but cached14. The second is deliberately labeled XML-only:
// it changes only that cache to15 on a copy, proving closure after transport and
// formatting rewrites without claiming that Office refreshed its own cache.
func TestQuantitativeOfficeSavedWorkbookClosure(t *testing.T) {
	root := os.Getenv("PPTXGENGO_OFFICE_CHART_FIXTURE")
	if root == "" {
		t.Skip("private Office Save As fixture not selected")
	}
	p, e := Load(filepath.Join(root, "project"))
	if e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "working", "chart-column-rail-gui.pptx"))
	if e != nil {
		t.Fatal(e)
	}
	original, e := InspectNativeLineage(b.files["deck.pptx"], b.Objects)
	if e != nil {
		t.Fatal(e)
	}
	edited, e := InspectNativeLineage(raw, b.Objects)
	if e != nil {
		t.Fatal(e)
	}
	var before, after *NativeLineageObject
	for i := range original.Objects {
		if original.Objects[i].NativeName == "node01.native" {
			before = &original.Objects[i]
		}
	}
	if before == nil {
		t.Fatal("authenticated baseline chart missing")
	}
	for i := range edited.Objects {
		if edited.Objects[i].ShapeToken == before.ShapeToken {
			after = &edited.Objects[i]
		}
	}
	if after == nil || after.ParentToken != before.ParentToken || after.NativeName != before.NativeName {
		t.Fatal("actual chart ownership drift")
	}
	a, e := readNativeChartFacts(b.files["deck.pptx"], *before, 3, 2, false)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = readNativeChartFacts(raw, *after, 3, 2, false); e == nil || !strings.Contains(e.Error(), "cache/workbook disagree at B2") {
		t.Fatalf("actual contradiction not refused: %v", e)
	}
	pkg, e := openLineagePackage(raw)
	if e != nil {
		t.Fatal(e)
	}
	chart, e := pkg.read("ppt/charts/chart1.xml")
	if e != nil {
		t.Fatal(e)
	}
	re := regexp.MustCompile(`(<c:numCache>.*?<c:pt idx="0"><c:v>)14(</c:v></c:pt>)`)
	if !re.Match(chart) {
		t.Fatal("expected actual stale14 cache absent; preserve exact original evidence")
	}
	// Replace just the first numerical cache, not all series or the GUI artifact.
	loc := re.FindSubmatchIndex(chart)
	coherent := append([]byte(nil), chart[:loc[0]]...)
	coherent = append(coherent, re.ReplaceAll(chart[loc[0]:loc[1]], []byte("${1}15${2}"))...)
	coherent = append(coherent, chart[loc[1]:]...)
	raw, e = lineageRewrite(pkg, map[string][]byte{"ppt/charts/chart1.xml": coherent})
	if e != nil {
		t.Fatal(e)
	}
	next, e := readNativeChartFacts(raw, *after, 3, 2, false)
	if e != nil {
		t.Fatal(e)
	}
	if e = verifyChartSemanticClosure(a, next); e != nil {
		t.Fatal(e)
	}
	if *a.data[0][0] != 14 || *next.data[0][0] != 15 {
		t.Fatal("wrong coherent source facts")
	}
	if e = verifyChartShell(a, next); e == nil {
		t.Fatal("Office formatting falsely claimed identical")
	}
	// This test reads authentic historical Office facts without replacing pins.
	// The ordinary regression proves transactional apply/rebuild; a newly built
	// CLI must genuinely migrate/build a fresh project before another GUI edit.
}
