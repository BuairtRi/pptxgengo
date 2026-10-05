package wmdesign

import "testing"

func TestV6HeatHeaderNativeWidthAndLegacyIsolation(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	st, err := r.sceneDataToken("label", 600)
	if err != nil {
		t.Fatal(err)
	}
	minimum := 1.0
	for _, width := range []float64{58, 60, 61, 66} {
		for _, label := range []string{"Technology", "Governance"} {
			c := sceneTableColumn{Label: label, Type: "heat", Width: width, Min: &minimum}
			fitted, left, right, err := r.v6HeatHeaderFit(c, st, nil)
			if err != nil || left < 1 || left > 12 || right != left || fitted.Family != st.Family || fitted.Size < 8 || fitted.Size > st.Size {
				t.Fatalf("%s %.0f: insets %.2f %.2f, %v", label, width, left, right, err)
			}
			layout, err := r.typeEngine.Measure(label, fitted, width-left-right)
			if err != nil || len(layout.Lines) != 1 || sequenceInlineWidth(layout.Lines[0].Advance, fitted) > width-left-right+0.0001 {
				t.Fatalf("%s %.0f: native word does not fit: %+v %v", label, width, layout, err)
			}
		}
	}
	for _, tc := range []struct {
		revision string
		c        sceneTableColumn
	}{
		{LibraryRevisionV5, sceneTableColumn{Label: "Technology", Type: "heat", Width: 58, Min: &minimum}},
		{LibraryRevisionV6, sceneTableColumn{Label: "Technology", Type: "heat", Width: 58}},
		{LibraryRevisionV6, sceneTableColumn{Label: "Technology", Type: "", Width: 58, Min: &minimum}},
		{LibraryRevisionV6, sceneTableColumn{Label: "Technology readiness", Type: "heat", Width: 58, Min: &minimum}},
		{LibraryRevisionV6, sceneTableColumn{Label: "Technology", Type: "heat", Width: 28, Min: &minimum}},
	} {
		r.source.Revision = tc.revision
		fitted, left, right, err := r.v6HeatHeaderFit(tc.c, st, nil)
		if err != nil || left != 12 || right != 12 || fitted != st {
			t.Fatalf("unsupported allocation changed: %+v %.2f %.2f %v", tc, left, right, err)
		}
	}
}
