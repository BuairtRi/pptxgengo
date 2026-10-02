package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
)

func validateOwnedParts(sr SlideReport) error {
	names := map[string]bool{}
	add := func(id string) error {
		if names[id] {
			return fmt.Errorf("node.owned_part_collision: %s", id)
		}
		names[id] = true
		return nil
	}
	for _, t := range sr.Texts {
		if e := add(t.ID); e != nil {
			return e
		}
	}
	for _, n := range sr.Nodes {
		if n.Kind == "box" || n.Kind == "rule" {
			if e := add(n.ID); e != nil {
				return e
			}
		}
	}
	for _, row := range sr.CardRows {
		if e := add(row.ID); e != nil {
			return e
		}
	}
	for _, c := range sr.Components {
		if e := add(c.ID); e != nil {
			return e
		}
		for _, sh := range c.Shapes {
			if e := add(sh.ID); e != nil {
				return e
			}
		}
	}

	// Scene text is already recorded above. Add remaining native objects and
	// group identities once; bottom-up hierarchy is checked by sceneNativeGroups.
	sceneTexts := map[string]bool{}
	for _, t := range sr.Texts {
		sceneTexts[t.ID] = true
	}
	for _, scene := range sr.Scenes {
		for _, part := range scene.Parts {
			if !sceneTexts[part] {
				if err := add(part); err != nil {
					return err
				}
			}
		}
		for _, group := range scene.Groups {
			if err := add(group.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// componentGroups adds actual DrawingML groups after typography serialization.
// Children stay in slide coordinates; chOff/chExt equal off/ext, giving an
// identity transform initially and ordinary move/resize behavior in PowerPoint.
func componentGroups(raw []byte, slides []SlideReport) ([]byte, error) {
	any := false
	byPart := map[string][]ComponentRecord{}
	for i, s := range slides {
		if len(s.Components) > 0 {
			any = true
			byPart[fmt.Sprintf("ppt/slides/slide%d.xml", i+1)] = s.Components
		}
	}
	if !any {
		return raw, nil
	}
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		return nil, e
	}
	shapes := regexp.MustCompile(`(?s)<p:sp>.*?</p:sp>`)
	names := regexp.MustCompile(`<p:cNvPr\b[^>]*\bname="([^"]*)"`)
	ids := regexp.MustCompile(`<p:cNvPr\b[^>]*\bid="(\d+)"`)
	esc := func(s string) string { var b bytes.Buffer; xml.EscapeText(&b, []byte(s)); return b.String() }
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range z.File {
		rc, e := f.Open()
		if e != nil {
			return nil, e
		}
		b, e := io.ReadAll(rc)
		rc.Close()
		if e != nil {
			return nil, e
		}
		if cs := byPart[f.Name]; len(cs) > 0 {
			next := 1
			for _, m := range ids.FindAllSubmatch(b, -1) {
				id, _ := strconv.Atoi(string(m[1]))
				if id >= next {
					next = id + 1
				}
			}
			for _, c := range cs {
				owned := map[string]bool{}
				for _, id := range c.Parts {
					owned[id] = true
				}
				count := 0
				first, last := -1, -1
				var children bytes.Buffer
				for _, loc := range shapes.FindAllIndex(b, -1) {
					sp := b[loc[0]:loc[1]]
					m := names.FindSubmatch(sp)
					if m == nil {
						continue
					}
					var attr struct {
						Name string `xml:"name,attr"`
					}
					if e := xml.Unmarshal([]byte(`<n name="`+string(m[1])+`"/>`), &attr); e != nil {
						return nil, e
					}
					if !owned[attr.Name] {
						continue
					}
					if first < 0 {
						first = loc[0]
					} else if len(bytes.TrimSpace(b[last:loc[0]])) > 0 {
						return nil, fmt.Errorf("component.noncontiguous_parts: %s", c.ID)
					}
					last = loc[1]
					count++
					children.Write(sp)
				}
				if count != len(c.Parts) || count == 0 {
					return nil, fmt.Errorf("component.missing_native_parts: %s", c.ID)
				}
				emu := func(v float64) int64 { return int64(math.Round(v * 12700)) }
				r := c.Rect
				group := fmt.Sprintf(`<p:grpSp><p:nvGrpSpPr><p:cNvPr id="%d" name="%s"/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/><a:chOff x="%d" y="%d"/><a:chExt cx="%d" cy="%d"/></a:xfrm></p:grpSpPr>`, next, esc(c.ID), emu(r.X), emu(r.Y), emu(r.W), emu(r.H), emu(r.X), emu(r.Y), emu(r.W), emu(r.H))
				next++
				var updated bytes.Buffer
				updated.Write(b[:first])
				updated.WriteString(group)
				updated.Write(children.Bytes())
				updated.WriteString("</p:grpSp>")
				updated.Write(b[last:])
				b = updated.Bytes()
			}
		}
		dst, e := w.CreateHeader(&f.FileHeader)
		if e != nil {
			return nil, e
		}
		if _, e = dst.Write(b); e != nil {
			return nil, e
		}
	}
	if e = w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
