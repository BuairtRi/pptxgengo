package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const missingLineSource = `{"type":"chart","kind":"line","x":100,"y":150,"w":650,"h":250,"categories":["A","B","C","D","E"],"series":[{"name":"Actual","values":[null,0,null,30,null]},{"name":"Plan","values":[10,20,30,40,50],"dashed":true}]}`

func TestSceneChartMissingValuesV4Native(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, rev := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, "unregistered"} {
		r.source.Revision = rev
		if _, _, e := r.planChartScene("chart", json.RawMessage(missingLineSource), SceneContext{Surface: "light"}); e == nil || !strings.Contains(e.Error(), "missing_value_not_supported") {
			t.Fatalf("legacy acceptance changed %s %v", rev, e)
		}
	}
	r.source.Revision = IntakeTeamCurveMonotoneRevision
	p, ok, e := r.planChartScene("chart", json.RawMessage(missingLineSource), SceneContext{Surface: "light"})
	if e != nil || !ok {
		t.Fatal(e)
	}
	found := false
	for _, it := range p.Items {
		if it.Chart != nil {
			found = true
			c := it.Chart
			if len(c.Data[0].Values) != 5 || len(c.Data[0].MissingValues) != 5 || !c.Data[0].MissingValues[0] || !c.Data[0].MissingValues[2] || !c.Data[0].MissingValues[4] || c.Data[0].MissingValues[1] || c.Options.DisplayBlanksAs != "span" || c.Options.MultiTypes[1].Options.LineDash != "dash" {
				t.Fatal("native mask/categories/dash lost")
			}
		}
	}
	if !found {
		t.Fatal("no native chart")
	}
}

func TestSceneChartMissingValuesStrict(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = IntakeTeamCurveMonotoneRevision
	for _, change := range []string{`"mode":"stacked"`, `"kind":"column"`, `"series":[{"name":"A","values":[null,null,null,null,null]}]`, `"series":[{"name":"A","values":[1,2]}]`, `"series":[{"name":"A","values":[1,2,3,4,1e308]}]`, `"yMin":-1e308`, `"series":[{"name":"A","values":[1,2,3,4,-1]}]`} {
		var n, patch map[string]any
		json.Unmarshal([]byte(missingLineSource), &n)
		json.Unmarshal([]byte("{"+change+"}"), &patch)
		for k, v := range patch {
			n[k] = v
		}
		raw, _ := json.Marshal(n)
		if _, _, e := r.planChartScene("chart", raw, SceneContext{Surface: "light"}); e == nil {
			t.Fatal("invalid data accepted", change)
		}
	}
	var n sceneChartSource
	json.Unmarshal([]byte(missingLineSource), &n)
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		n.Series[0].Values[1] = &v
		if _, e := r.sceneSourceChart("chart", n, SceneContext{Surface: "light"}); e == nil {
			t.Fatal("nonfinite value accepted")
		}
	}
	n.Series[0].Values[1] = sceneChartFloat(0)
	n.Series[0].Name = strings.Repeat("a", 4097)
	if _, e := r.sceneSourceChart("chart", n, SceneContext{Surface: "light"}); e == nil {
		t.Fatal("unbounded copy accepted")
	}
}

func TestSceneChartMissingValuesFrozenAdoptionFullSlide(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	for _, pin := range []struct{ path, hash string }{{"templates/library/change.json", "50c9b7c2bbd968fda1adb78ff0c4efdd2f823df0d17e6043514d442834c033ac"}, {"explorations/components.src.html", "72057c4f23f077f095e7b6f22705b35bbb538eee55d9d9da6cf752f6cc10d20c"}} {
		b, e := os.ReadFile(filepath.Join(root, pin.path))
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != pin.hash {
			t.Fatal("immutable chart source changed", pin.path)
		}
	}
	entries := intakeRepairEntries(t, filepath.Join(root, "templates", "library"))
	entry := entries["readiness/adoption-curve"]
	if len(entry.Slide) == 0 {
		t.Fatal("frozen adoption fixture missing")
	}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	source, e := Load(bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(root, "frames", "v0", "frames.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &source.Frames); e != nil {
		t.Fatal(e)
	}
	source.Revision = IntakeTeamCurveMonotoneRevision
	slide := intakeRepairSlide(t, entry)
	if e = ApplyIncomingIntakeRepairs("readiness/adoption-curve", IntakeRepairRevision, &slide); e != nil {
		t.Fatal(e)
	}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}
	pkg, _, e := buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(pkg), int64(len(pkg)))
	if e != nil {
		t.Fatal(e)
	}
	var chart string
	for _, f := range z.File {
		if f.Name == "ppt/charts/chart1.xml" {
			rd, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			b, e := io.ReadAll(rd)
			rd.Close()
			if e != nil {
				t.Fatal(e)
			}
			chart = string(b)
		}
	}
	if !strings.Contains(chart, `<c:ptCount val="7"/>`) || !strings.Contains(chart, `<c:dispBlanksAs val="span"/>`) || !strings.Contains(chart, "Actual, wave 1") {
		t.Fatal("editable frozen chart missing")
	}
}
