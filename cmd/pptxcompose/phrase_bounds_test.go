package main

import (
	"github.com/buairtri/pptxgengo/internal/compose"
	"testing"
)

func TestNativePhraseFragments(t *testing.T) {
	e := element{Text: "alpha beta", Frame: frame{X: 100, Y: 50, Width: 100, Height: 60}, PhraseRequests: []compose.PhraseRequest{{ID: "p", Phrase: "alpha beta"}}}
	row := nativeRow{}
	for i, ch := range []rune(e.Text) {
		x, y := float64(i*8+100), 52.0
		if i >= 6 {
			x, y = float64((i-6)*8+100), 72
		}
		row.Characters = append(row.Characters, nativeCharacter{Text: string(ch), Bounds: &nativeRect{Left: x, Top: y, Width: 8, Height: 18}})
	}
	b, err := nativePhraseBounds(e, row)
	if err != nil {
		t.Fatal(err)
	}
	if len(b["p"]) != 2 || b["p"][0].X != 0 || b["p"][0].Width != 40 || b["p"][1].Y != 22 {
		t.Fatalf("wrong line fragments: %+v", b)
	}
	row.Characters[0].Bounds = nil
	if _, err = nativePhraseBounds(e, row); err == nil {
		t.Fatal("missing native geometry accepted")
	}
}
func TestPhraseSelectionChangesCacheContract(t *testing.T) {
	q := compose.ProbeRequest{Text: "key and key", PhraseRequests: []compose.PhraseRequest{{ID: "a", Phrase: "key", Occurrence: 1}}}
	one := contractKey(q)
	q.PhraseRequests[0].Occurrence = 2
	if one == contractKey(q) {
		t.Fatal("phrase selector not bound to cache")
	}
}
