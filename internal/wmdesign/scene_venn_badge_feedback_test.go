package wmdesign

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestVennBadgeNativeInlineReserveV4(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"venn","x":100,"y":80,"w":558,"h":432,"sets":[{"label":"A"},{"label":"B"}],"points":[{"x":0.3,"y":0.3,"label":"Master index","side":"left"},{"x":0.6,"y":0.6,"label":"[[Raw]]"}]}`)
	for _, rev := range []string{LibraryRevisionV3, IntakeTeamCurveMonotoneRevision} {
		r.source.Revision = rev
		p, handled, err := r.planIntakeVennScene("venn", raw, SceneContext{Surface: "light"})
		if err != nil || !handled {
			t.Fatal(err)
		}
		labels := 0
		for _, item := range p.Items {
			if item.Text == nil || !strings.Contains(item.Text.ID, ".points.") || !strings.HasSuffix(item.Text.ID, ".label") {
				continue
			}
			labels++
			tr := item.Text
			advance := tr.Layout.Lines[0].Advance
			reserve := 0.
			wantSize := 8.
			if rev == IntakeTeamCurveMonotoneRevision {
				reserve = 2
				wantSize = 9.5
			}
			if math.Abs(tr.Rect.W-advance-reserve) > .001 {
				t.Fatalf("%s width%.3f advance%.3f reserve%.3f", rev, tr.Rect.W, advance, reserve)
			}
			if tr.Layout.Style.Size != wantSize || tr.Layout.Style.Weight != 600 || tr.Layout.Style.TrackingPt != math.Round(wantSize*.06*100)/100 {
				t.Fatal("font or emitted tracking changed")
			}
			badge := enhancementShape(t, p, strings.TrimPrefix(strings.TrimSuffix(tr.ID, ".label"), "venn")+".badge")
			if math.Abs(badge.Record.Rect.W+.75-tr.Rect.W-10) > .001 {
				t.Fatal("source5pt padding or insetstroke changed")
			}
			if strings.Contains(tr.ID, "source-002") && tr.Layout.Displayed != "[[RAW]]" {
				t.Fatal("literal copy changed")
			}
		}
		if labels != 2 {
			t.Fatal("labels missing")
		}
	}
}

func TestVennBadgeNativeReserveKeepsWidthBound(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = IntakeTeamCurveMonotoneRevision
	for _, count := range []int{20, 21} {
		n := intakeVennSource{Type: "venn", X: 100, Y: 80, W: 558, H: 432, Sets: []intakeVennSet{{Label: "A"}, {Label: "B"}}, Points: []intakeVennPoint{{X: sceneChartFloat(.5), Y: sceneChartFloat(.5), Label: strings.Repeat("A", count)}}}
		raw, _ := json.Marshal(n)
		_, _, err := r.planIntakeVennScene("venn", raw, SceneContext{Surface: "light"})
		if count == 20 && err != nil {
			t.Fatal(err)
		}
		if count == 21 && (err == nil || !strings.Contains(err.Error(), "badge_width")) {
			t.Fatal("140pt allocation bound bypassed", err)
		}
	}
}
