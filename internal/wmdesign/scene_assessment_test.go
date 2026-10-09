package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func sceneAssessmentFixture() AssessmentSpec {
	zero := 0
	return AssessmentSpec{Type: "assessment", X: 57, Y: 126, W: 846, H: 260, RowLabel: "Capability", LabelWidth: 180, RowHeight: 36, ShowScores: true, Domain: AssessmentDomain{Max: 4, Scale: "seq", Labels: []string{"Absent", "Initial", "Developing", "Established", "Leading"}, MissingLabel: "Not assessed"}, Columns: []AssessmentAxis{{Key: "north", Label: "North"}, {Key: "south", Label: "South"}}, Rows: []AssessmentRow{{Key: "access", Label: "Access", Scores: map[string]*int{"north": &zero, "south": nil}}}}
}
func TestAssessmentSceneZeroAndMissingRemainDistinct(t *testing.T) {
	r := intakeTestRenderer(t)
	s := sceneAssessmentFixture()
	raw, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	plan, handled, e := r.planAssessmentScene("assessment", raw, SceneContext{Surface: "light"})
	if e != nil || !handled {
		t.Fatal(handled, e)
	}
	zero, missing, zeroHeat, missingHeat, missingLegend, missingSwatch := false, false, false, false, false, false
	for _, it := range plan.Items {
		if it.Table != nil {
			for _, cell := range it.Table.CellRecords {
				if strings.HasSuffix(cell.Text.ID, ".row.access.north") {
					zero = cell.Text.Layout.Displayed == "0"
				}
				if strings.HasSuffix(cell.Text.ID, ".row.access.south") {
					missing = cell.Text.Layout.Displayed == ""
				}
			}
		}
		if it.Shape != nil {
			if strings.HasSuffix(it.Shape.Record.ID, ".legend.items.missing.swatch") {
				missingSwatch = it.Shape.Record.Color == "FFFFFF"
			}
			if strings.HasSuffix(it.Shape.Record.ID, ".row.access.north.heat") {
				zeroHeat = it.Shape.Record.Color == "E8EEF8"
			}
			if strings.HasSuffix(it.Shape.Record.ID, ".row.access.south.heat") {
				missingHeat = true
			}
		}
		if it.Text != nil && strings.Contains(it.Text.Layout.Displayed, "Blank: Not assessed") {
			missingLegend = true
		}
	}
	if !zero || !missing || !zeroHeat || missingHeat || !missingLegend || !missingSwatch {
		t.Fatalf("zero/missing contract lost: %v %v %v %v %v", zero, missing, zeroHeat, missingHeat, missingLegend)
	}
}
func TestAssessmentSceneReservedColumnKeysAndDomain(t *testing.T) {
	for _, key := range []string{"group", "total", "ink", "scale", "h", "assessment-label"} {
		s := sceneAssessmentFixture()
		s.Columns[0].Key = key
		s.Rows[0].Scores = map[string]*int{}
		if e := ValidateAssessment(s); e == nil {
			t.Fatal("reserved key accepted", key)
		}
	}
	s := sceneAssessmentFixture()
	s.Domain.Max = 3
	s.Domain.Labels = s.Domain.Labels[:4]
	high := 4
	s.Rows[0].Scores["north"] = &high
	if e := ValidateAssessment(s); e == nil {
		t.Fatal("domain reduction retained incompatible score")
	}
}
func TestAssessmentSceneLongLabelsAndBoundedFit(t *testing.T) {
	r := intakeTestRenderer(t)
	s := sceneAssessmentFixture()
	s.RowHeight = 72
	s.Rows[0].Label = "Clinical documentation and care coordination"
	raw, _ := json.Marshal(s)
	if _, _, e := r.planAssessmentScene("assessment", raw, SceneContext{Surface: "light"}); e != nil {
		t.Fatal("two-line label not supported", e)
	}
	s.H = 60
	raw, _ = json.Marshal(s)
	if _, _, e := r.planAssessmentScene("assessment", raw, SceneContext{Surface: "light"}); e == nil {
		t.Fatal("undersized allocation accepted")
	}
}
