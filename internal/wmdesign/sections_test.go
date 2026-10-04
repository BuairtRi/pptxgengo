package wmdesign

import (
	"strings"
	"testing"
)

func TestSectionsNativeContract(t *testing.T) {
	good := []SectionSpec{{"one", "Opening", "a"}, {"two", "Résultats 東京", "c"}}
	if e := ValidateSections(good, []string{"a", "b", "c"}); e != nil {
		t.Fatal(e)
	}
	for _, bad := range [][]SectionSpec{{{"one", "Opening", "b"}}, {{"one", "Opening", "a"}, {"two", "Later", "a"}}, {{"one", "Opening", "a"}, {"two", "OPENING", "c"}}, {{"one", "bad\x01", "a"}}} {
		if e := ValidateSections(bad, []string{"a", "b", "c"}); e == nil {
			t.Fatal("invalid native sections accepted", bad)
		}
	}
	for _, bad := range []string{"bad\x00", "bad\uffff", strings.Repeat("x", 1048577), string([]byte{0xff})} {
		if e := ValidateSpeakerNotes(bad); e == nil {
			t.Fatal("invalid notes accepted")
		}
	}
	if e := ValidateSpeakerNotes("First & original.\n\nSecond 東京.\n"); e != nil {
		t.Fatal(e)
	}
}
