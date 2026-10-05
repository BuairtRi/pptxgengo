package wmdesign

import (
	"encoding/json"
	"math"
	"testing"
)

func TestSceneTablePriorityVocabularyAndPalette(t *testing.T) {
	tests := []struct {
		value, label, background, foreground string
	}{
		{"P1", "P1", scenePriorityNavy, scenePriorityWhite},
		{"p2", "P2", scenePriorityMidNavy, scenePriorityWhite},
		{"P3", "P3", scenePriorityPale, scenePriorityNavy},
		{"p4", "P4", "transparent", scenePriorityNavy},
		{"NOW", "Now", scenePriorityNavy, scenePriorityWhite},
		{"Next", "Next", scenePriorityPale, scenePriorityNavy},
		{"later", "Later", "transparent", scenePriorityNavy},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			chip, err := sceneTablePriority(json.RawMessage(`"` + test.value + `"`))
			if err != nil {
				t.Fatal(err)
			}
			if !chip.Known || chip.Value != test.value || chip.Label != test.label || chip.Background != test.background || chip.Foreground != test.foreground || chip.Border != scenePriorityNavy {
				t.Fatalf("unexpected chip: %+v", chip)
			}
		})
	}
}

func TestSceneTablePriorityObjectEditableValueAndCustomLabel(t *testing.T) {
	chip, err := sceneTablePriority(json.RawMessage(`{"value":"p1","label":"Critical"}`))
	if err != nil {
		t.Fatal(err)
	}
	if chip.Value != "p1" || chip.Label != "Critical" || !chip.Known {
		t.Fatalf("semantic value or display label changed: %+v", chip)
	}
	unknown, err := sceneTablePriority(json.RawMessage(`{"value":"Escalate","label":"Needs escalation"}`))
	if err != nil {
		t.Fatal(err)
	}
	if unknown.Known || unknown.Value != "Escalate" || unknown.Label != "Needs escalation" || unknown.Background != "transparent" || unknown.Foreground != scenePriorityNavy {
		t.Fatalf("unknown priority should retain editable copy with outline fallback: %+v", unknown)
	}
	withoutLabel, err := sceneTablePriority(json.RawMessage(`"Review"`))
	if err != nil || withoutLabel.Label != "Review" || withoutLabel.Value != "Review" {
		t.Fatalf("unknown string fallback lost authored value: %v %+v", err, withoutLabel)
	}
}

func TestSceneTablePriorityMeasuredWidthAndInvalidValues(t *testing.T) {
	chip, err := sceneTablePriority(json.RawMessage(`"P1"`))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ measured, want float64 }{{0, 40}, {20, 40}, {30, 46}} {
		got, err := chip.WidthForLabel(test.measured)
		if err != nil || got != test.want {
			t.Fatalf("WidthForLabel(%v) = %v, %v; want %v", test.measured, got, err, test.want)
		}
	}
	for _, raw := range []string{`null`, `4`, `true`, `" "`, `{}`, `{"value":2}`, `{"value":"P1","other":"x"}`, `{"value":"P1","label":false}`} {
		if _, err := sceneTablePriority(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted malformed priority %s", raw)
		}
	}
	for _, width := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, err := chip.WidthForLabel(width); err == nil {
			t.Errorf("accepted invalid measured text width %v", width)
		}
	}
}
