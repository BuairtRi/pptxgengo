package wmdesign

import (
	"fmt"
	"math"
)

// SplitFrameReference is a native review specimen, not a client slide library.
func SplitFrameReference(bundle, override string, year int) (Document, error) {
	s, err := Load(bundle, override)
	if err != nil {
		return Document{}, err
	}
	return splitFrameReference(s, year)
}

func splitFrameReference(s *Source, year int) (Document, error) {
	if !isModernLibrary(s.Revision) {
		return Document{}, fmt.Errorf("frame.reference_requires_library_v2")
	}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	add := func(q FrameRequest, title string) error {
		f, err := s.ResolveFrame(q)
		if err != nil {
			return err
		}
		slide := SlideSpec{ID: fmt.Sprintf("frame-%02d", len(doc.Slides)+1), Frame: q, Eyebrow: "Frame reference", Title: title, ContentKind: "synthetic_example"}
		if q.SourceLines == 1 {
			slide.Source = "Source: synthetic frame example."
		}
		if q.SourceLines == 2 {
			slide.Source = "1 Synthetic reference, not client data.\nSource: WMDS round 4."
		}
		zones := []struct {
			scope, label string
			b            Rect
		}{{"short", "Short content column", f.ShortBody}, {"tall", "Tall content column", f.TallBody}}
		if q.Split == "" {
			zones = []struct {
				scope, label string
				b            Rect
			}{{"body", "Full content zone", f.Body}}
		}
		for _, z := range zones {
			height := math.Floor(z.b.H/18) * 18
			slide.Nodes = append(slide.Nodes, Node{ID: z.scope + ".surface", Kind: "box", Scope: z.scope, Surface: "subtle", Rect: Rect{z.b.X, z.b.Y, z.b.W, height}},
				Node{ID: z.scope + ".label", Kind: "text", Scope: z.scope, Style: "subhead", Text: z.label, Rect: Rect{z.b.X + 18, z.b.Y + 18, z.b.W - 36, 0}},
				Node{ID: z.scope + ".copy", Kind: "text", Scope: z.scope, Style: "small", Text: fmt.Sprintf("x %.0f–%.0f pt\ny %.0f–%.0f pt\n\nNative shapes and text\nFooter: %s", z.b.X, z.b.X+z.b.W, z.b.Y, z.b.Y+z.b.H, q.Footer), Rect: Rect{z.b.X + 18, z.b.Y + 90, z.b.W - 36, 0}})
		}
		doc.Slides = append(doc.Slides, slide)
		return nil
	}
	splits := []string{"tall-right", "tall-left", "tall-right-narrow", "tall-left-narrow"}
	for i, split := range splits {
		for j, footer := range []string{"compact", "tall"} {
			lines := 1 + (i+j)%3
			title := []string{"A split frame", "A split title\non two lines", "Split headlines\ncan use\nthree lines"}[lines-1]
			q := FrameRequest{Split: split, Rail: "none", Footer: footer, TitleLines: lines, SourceLines: 1}
			if i == 3 && j == 1 {
				q.SourceLines = 2
			}
			if err := add(q, title); err != nil {
				return Document{}, err
			}
		}
	}
	for i, n := range []int{2, 6, 2, 6} {
		q := FrameRequest{Rail: "nav", Footer: "compact", SourceLines: 1, TitleLines: 1}
		if i >= 2 {
			q.Split = splits[i-2]
		}
		if i%2 == 1 {
			q.Footer = "tall"
		}
		labels := []string{"Context", "Solution", "Approach", "Team", "Evidence", "Decision"}
		for j := 0; j < n; j++ {
			q.Nav = append(q.Nav, NavTab{ID: fmt.Sprintf("section-%d", j+1), Label: labels[j]})
		}
		q.Active = q.Nav[n-1].ID
		if err := add(q, fmt.Sprintf("%d navigation tabs", n)); err != nil {
			return Document{}, err
		}
	}
	if _, updated := s.Frames.Footers["slim"]; updated {
		for _, rail := range []string{"none", "left", "right", "nav"} {
			for sourceLines := 0; sourceLines <= 2; sourceLines++ {
				q := FrameRequest{Rail: rail, Footer: "slim", TitleLines: 4, SourceLines: sourceLines}
				if rail == "nav" {
					q.Nav = []NavTab{{ID: "context", Label: "Context"}, {ID: "decision", Label: "Decision"}}
					q.Active = "decision"
				}
				if err := add(q, "Four authored lines\nkeep the headline\nabove the body\nwith a slim footer"); err != nil {
					return Document{}, err
				}
			}
		}
		for _, split := range splits {
			q := FrameRequest{Split: split, Rail: "none", Footer: "slim", TitleLines: 4, SourceLines: 1}
			if err := add(q, "A split frame\ncan preserve\nfour authored\ntitle lines"); err != nil {
				return Document{}, err
			}
		}
		// Full-width three-line allocation is new alongside the fourth line.
		if err := add(FrameRequest{Rail: "none", Footer: "compact", TitleLines: 3}, "Three title lines\nnow have an explicit\nheader allocation"); err != nil {
			return Document{}, err
		}
	}
	return doc, nil
}
