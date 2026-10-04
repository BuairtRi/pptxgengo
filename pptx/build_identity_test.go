package pptx

import (
	"bytes"
	"testing"
	"time"
)

func TestBuildIdentityRepeatedWrites(t *testing.T) {
	p := New()
	if err := p.SetBuildIdentity(BuildIdentity{Timestamp: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Seed: "source-a"}); err != nil {
		t.Fatal(err)
	}
	p.AddSection(SectionProps{Title: "First section"})
	s := p.AddSlide(&AddSlideProps{SectionTitle: "First section"})
	if err := s.AddText([]TextProps{{Text: "Editable copy"}}, nil); err != nil {
		t.Fatal(err)
	}
	a, err := p.Write()
	if err != nil {
		t.Fatal(err)
	}
	b, err := p.Write()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("same presentation and build identity changed package bytes")
	}
	if err := p.SetBuildIdentity(BuildIdentity{}); err == nil {
		t.Fatal("empty identity accepted")
	}
}
