package wmdesign

import (
	"encoding/json"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeEditingEditableCardSingleNativePlanAndRefusals(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	typography, e := NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	r.typeEngine = typography
	for _, level := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = level
		p, e := r.planEditableCard("card", json.RawMessage(`{"type":"editable-card","x":100,"y":120,"w":300,"h":200,"surface":"subtle","title":"A plain title","body":"Measured plain body copy that wraps within the card allocation."}`), SceneContext{Surface: "light"})
		if e != nil {
			t.Fatal(level, e)
		}
		if len(p.Items) != 1 || len(p.Groups) != 0 || p.Items[0].Text == nil {
			t.Fatal("card is not one native unit", p)
		}
		tr := p.Items[0].Text
		if tr.NativeShape == nil || tr.NativeShape.ParagraphContract != EditableCardContract || tr.NativeShape.Rect != (Rect{100, 120, 300, 200}) || len(tr.Rich.Paragraphs) != 2 {
			t.Fatal(tr)
		}
		if tr.Rich.Paragraphs[0].Key != "title" || tr.Rich.Paragraphs[1].Key != "body" || tr.Rich.Paragraphs[0].Runs[0].Style.Size == tr.Rich.Paragraphs[1].Runs[0].Style.Size {
			t.Fatal("roles collapsed", tr.Rich)
		}
		xml := string(editableCardParagraphXML(*tr))
		if strings.Count(xml, "<a:p>") != 2 || strings.Count(xml, "<a:r>") != 2 || strings.Contains(xml, "<a:br") {
			t.Fatal("paragraph source boundaries changed", xml)
		}
	}
	for _, raw := range []string{
		`{"type":"editable-card","w":300,"h":200,"title":"[[Rich]]","body":"Copy"}`,
		`{"type":"editable-card","w":300,"h":200,"title":"Title","body":"Two\nparagraphs"}`,
		`{"type":"editable-card","w":300,"h":200,"title":"Title","body":"Copy","style":"body"}`,
		`{"type":"editable-card","w":300,"h":200,"title":"Title","body":"Copy","surface":"outline"}`,
		`{"type":"editable-card","w":20,"h":200,"title":"Title","body":"Copy"}`,
		`{"type":"editable-card","w":300,"h":25,"title":"Title","body":"Copy"}`,
		`{"type":"editable-card","w":300,"h":200,"title":"","body":"Copy"}`,
	} {
		if _, e := r.planEditableCard("bad", json.RawMessage(raw), SceneContext{Surface: "light"}); e == nil {
			t.Fatal("accepted unsupported card", raw)
		}
	}
}

func TestNativeEditingEditableCardMatchesPlainSourceCardParagraphPositions(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = density
		before, handled, err := r.planCardScene("card", json.RawMessage(`{"type":"card","x":57,"y":198,"w":414,"h":216,"surface":"subtle","title":"Keep the deck editable","titleInk":"primary","pad":12,"gap":8,"body":[{"p":"Update the source, review the changes with the team, and share the latest approved deck."}]}`), SceneContext{Surface: "light"})
		if err != nil || !handled {
			t.Fatal(density, handled, err)
		}
		after, err := r.planEditableCard("card", json.RawMessage(`{"type":"editable-card","x":57,"y":198,"w":414,"h":216,"surface":"subtle","title":"Keep the deck editable","body":"Update the source, review the changes with the team, and share the latest approved deck."}`), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(density, err)
		}
		native := after.Items[0].Text
		line := 0
		for _, item := range before.Items {
			if item.Text == nil {
				continue
			}
			for _, expected := range item.Text.Layout.Lines {
				if line >= len(native.Layout.Lines) {
					t.Fatal("missing native line")
				}
				actual := native.Layout.Lines[line]
				want := item.Text.Rect.Y - native.Rect.Y + expected.Baseline
				if actual.Text != expected.Text || math.Abs(actual.Baseline-want) > .02 {
					t.Fatal("source card paragraph position changed", density, actual, expected, want)
				}
				line++
			}
		}
		if line != len(native.Layout.Lines) || native.Rich.Paragraphs[0].ParagraphGapAfter != 14 {
			t.Fatal("source body margin lost", density)
		}
	}
}
