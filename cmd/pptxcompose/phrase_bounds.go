package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/internal/compose"
)

// Retain the native character geometry already read by the adapter. Group by
// actual line position, allowing mixed font sizes on a common baseline. Bounds
// are relative to the measured text frame's inner origin so parent moves work.
func nativePhraseBounds(e element, row nativeRow) (map[string][]compose.Rect, error) {
	if len(e.PhraseRequests) == 0 {
		return nil, nil
	}
	text := strings.ReplaceAll(strings.ReplaceAll(e.Text, "\r\n", "\n"), "\r", "\n")
	chars := map[int]nativeCharacter{}
	n := 0
	for i, ch := range []rune(text) {
		if ch == '\n' {
			continue
		}
		if n >= len(row.Characters) || row.Characters[n].Text != string(ch) {
			return nil, fmt.Errorf("character sequence differs at rune %d", i)
		}
		c := row.Characters[n]
		n++
		if c.Bounds == nil || !finite(c.Bounds.Left) || !finite(c.Bounds.Top) || !finite(c.Bounds.Width) || !finite(c.Bounds.Height) || c.Bounds.Width < 0 || c.Bounds.Height < 0 {
			return nil, fmt.Errorf("character bounds unavailable at rune %d; remeasure with current adapter", i)
		}
		chars[i] = c
	}
	if n != len(row.Characters) {
		return nil, fmt.Errorf("extra native characters")
	}
	out := map[string][]compose.Rect{}
	for _, q := range e.PhraseRequests {
		a, b, err := compose.ResolvePhrase(text, q)
		if err != nil {
			continue
		} // planner rejects or explicitly stages unresolved selectors
		var lines []compose.Rect
		newLine := true
		for i := a; i < b; i++ {
			c, ok := chars[i]
			if !ok {
				newLine = true
				continue
			}
			if strings.TrimSpace(c.Text) == "" || c.Bounds.Width == 0 || c.Bounds.Height == 0 {
				continue
			}
			r := compose.Rect{X: c.Bounds.Left - e.Frame.X - e.InsetX, Y: c.Bounds.Top - e.Frame.Y - e.InsetY, Width: c.Bounds.Width, Height: c.Bounds.Height}
			if len(lines) > 0 && !newLine {
				prev := &lines[len(lines)-1]
				overlap := math.Min(prev.Y+prev.Height, r.Y+r.Height) - math.Max(prev.Y, r.Y)
				if overlap > .5*math.Min(prev.Height, r.Height) && r.X >= prev.X-.1 {
					right, bottom := math.Max(prev.X+prev.Width, r.X+r.Width), math.Max(prev.Y+prev.Height, r.Y+r.Height)
					prev.X = math.Min(prev.X, r.X)
					prev.Y = math.Min(prev.Y, r.Y)
					prev.Width = right - prev.X
					prev.Height = bottom - prev.Y
					continue
				}
			}
			lines = append(lines, r)
			newLine = false
		}
		if len(lines) == 0 {
			return nil, fmt.Errorf("phrase %s has no visible character geometry", q.ID)
		}
		out[q.ID] = lines
	}
	return out, nil
}
