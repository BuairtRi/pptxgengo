package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

const slimFrameSourceSHA = "a0b35d7ce3eb97352b779872d46fc238476dbaddd3ade07327477df69236be2a"
const slimFrameRendererSHA = "3284f866ee67e37d8960a252fbc579ac10f332db1212f7806c98d70d45eacc64"

func slimFrameSource(t *testing.T) *Source {
	t.Helper()
	s, e := Load(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "")
	if e != nil {
		t.Fatal(e)
	}
	// This immutable Round12 snapshot is independent of the evolving WM checkout.
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-round12", "source")
	for _, f := range []struct{ path, sha string }{{"frames/v0/frames.json", slimFrameSourceSHA}, {"explorations/components.src.html", slimFrameRendererSHA}} {
		b, e := os.ReadFile(filepath.Join(root, f.path))
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != f.sha {
			t.Fatalf("immutable frame source changed: %s", f.path)
		}
		if f.path == "frames/v0/frames.json" {
			if e = json.Unmarshal(b, &s.Frames); e != nil {
				t.Fatal(e)
			}
		}
	}
	return s
}
func TestSlimFrameGeometryAndFourLineTitles(t *testing.T) {
	s := slimFrameSource(t)
	for _, rail := range []string{"none", "left", "right", "nav"} {
		for _, split := range []string{"", "tall-right", "tall-left", "tall-right-narrow", "tall-left-narrow"} {
			if split != "" && rail != "none" && rail != "nav" {
				continue
			}
			for lines := 1; lines <= 4; lines++ {
				for sourceLines := 0; sourceLines <= 2; sourceLines++ {
					q := FrameRequest{Rail: rail, Footer: "slim", Split: split, TitleLines: lines, SourceLines: sourceLines}
					if rail == "nav" {
						q.Nav = []NavTab{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}
						q.Active = "b"
					}
					f, e := s.ResolveFrame(q)
					if e != nil {
						t.Fatalf("%+v: %v", q, e)
					}
					bottom := 486 - float64(sourceLines)*18
					if f.Body.Y+f.Body.H != bottom || f.FooterRule != 495 || f.FooterRow.Y != 504 || f.FooterRow.H != 12 || f.FooterBand.H != 0 {
						t.Fatalf("slim geometry %+v", f)
					}
					if sourceLines > 0 && (f.Source.Y != 489-float64(sourceLines)*12 || f.Source.H != float64(sourceLines)*12) {
						t.Fatalf("source geometry %+v", f.Source)
					}
					if rail == "nav" && f.NavBottom != 486 {
						t.Fatal("source lines shortened nav tabs")
					}
					wantRule := 108 + float64(lines-1)*36
					if split != "" && f.Header.W <= 270 {
						wantRule = []float64{108, 126, 162, 198}[lines-1]
					}
					if f.TitleRule != wantRule {
						t.Fatalf("title %d split %s rule%g want%g", lines, split, f.TitleRule, wantRule)
					}
					if split == "" {
						if f.Body.Y != wantRule+18 {
							t.Fatal("full body top")
						}
					} else {
						if f.ShortBody.Y != wantRule+18 || f.TallBody.Y != 36 {
							t.Fatal("split body geometry")
						}
					}
				}
			}
		}
	}
	for _, footer := range []string{"compact", "tall"} {
		f, e := s.ResolveFrame(FrameRequest{Footer: footer, TitleLines: 4})
		if e != nil || f.TitleRule != 216 || f.Body.Y != 234 {
			t.Fatalf("extended title footer%s %+v %v", footer, f, e)
		}
	}
}
func TestSlimFrameCurrentBundleAndCompactGeometry(t *testing.T) {
	s, e := Load(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "")
	if e != nil {
		t.Fatal(e)
	}
	for _, q := range []FrameRequest{{Footer: "slim"}, {Footer: "compact", TitleLines: 3}, {Footer: "compact", TitleLines: 4}, {Footer: "compact", Split: "tall-right", TitleLines: 4}} {
		if _, e = s.ResolveFrame(q); e != nil {
			t.Fatalf("published current frame rejected %+v: %v", q, e)
		}
	}
	f, e := s.ResolveFrame(FrameRequest{Footer: "compact", SourceLines: 1, TitleLines: 2})
	if e != nil || f.TitleRule != 144 || f.Body.Y != 162 || f.Body.Y+f.Body.H != 450 || f.Source.Y != 459 || f.FooterRow.Y != 486 || f.FooterRow.H != 18 {
		t.Fatalf("inherited compact frame changed %+v %v", f, e)
	}
	doc, e := splitFrameReference(s, 2026)
	if e != nil || len(doc.Slides) != 29 {
		t.Fatalf("current frame reference changed %d %v", len(doc.Slides), e)
	}
}
func TestSlimFrameSourceGateAndReference(t *testing.T) {
	s := slimFrameSource(t)
	doc, e := splitFrameReference(s, 2026)
	if e != nil || len(doc.Slides) != 29 {
		t.Fatalf("slim reference %d %v", len(doc.Slides), e)
	}
	if _, e = s.ResolveFrame(FrameRequest{Footer: "slim", TitleLines: 5}); e == nil {
		t.Fatal("five title lines accepted")
	}
	if _, e = s.ResolveFrame(FrameRequest{Footer: "slim", Density: "appendix", TitleLines: 4}); e == nil {
		t.Fatal("appendix four lines accepted")
	}
	for _, value := range []float64{0, math.NaN(), math.Inf(1)} {
		s := slimFrameSource(t)
		f := s.Frames.Footers["slim"]
		f.Bottom = value
		s.Frames.Footers["slim"] = f
		if _, e = s.ResolveFrame(FrameRequest{Footer: "slim"}); e == nil {
			t.Fatal("invalid slim footer accepted")
		}
	}
	s = slimFrameSource(t)
	for i, f := range s.Frames.Features {
		if f.ID == "zone.source" {
			s.Frames.Features[i].Geometry = json.RawMessage(`{"compact":{"bottom":471},"tall":{"bottom":477},"lineHeight":12,"bodyBottom":{"1":450,"2":432}}`)
		}
	}
	if _, e = s.ResolveFrame(FrameRequest{Footer: "slim", SourceLines: 1}); e == nil {
		t.Fatal("missing source slim allocation fabricated")
	}
}
