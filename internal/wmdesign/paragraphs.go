package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// candidateParagraphs writes the uniform text contract explicitly, including
// paragraphs with no runs. It is scoped to v2 and leaves the shared writer alone.
func candidateParagraphs(raw []byte, records map[int][]TextRecord) ([]byte, error) {
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	parts := map[string][]byte{}
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
		parts[f.Name] = b
	}
	catalog := map[string]map[string]TextRecord{}
	for slide, rs := range records {
		name := fmt.Sprintf("ppt/slides/slide%d.xml", slide)
		catalog[name] = map[string]TextRecord{}
		var rels struct {
			Relationships []struct {
				Type   string `xml:"Type,attr"`
				Target string `xml:"Target,attr"`
			} `xml:"Relationship"`
		}
		if e := xml.Unmarshal(parts[fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slide)], &rels); e != nil {
			return nil, e
		}
		layout := ""
		for _, r := range rels.Relationships {
			if strings.HasSuffix(r.Type, "/slideLayout") {
				layout = path.Clean(path.Join("ppt/slides", r.Target))
			}
		}
		if layout != "" {
			catalog[layout] = map[string]TextRecord{}
		}
		for _, r := range rs {
			catalog[name][r.ID] = r
			if layout != "" {
				catalog[layout][r.ID] = r
			}
		}
	}
	shape := regexp.MustCompile(`(?s)<p:sp>.*?</p:sp>`)
	objectName := regexp.MustCompile(`<p:cNvPr\b[^>]*\bname="([^"]*)"`)
	paragraphs := regexp.MustCompile(`(?s)<a:p>.*?</a:p>`)
	ppr := regexp.MustCompile(`(?s)<a:pPr\b[^>]*>.*?</a:pPr>|<a:pPr\b[^>]*/>`)
	escape := func(s string) string { var b bytes.Buffer; xml.EscapeText(&b, []byte(s)); return b.String() }
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range z.File {
		b := parts[f.Name]
		var rewriteErr error
		if byName, ok := catalog[f.Name]; ok {
			b = shape.ReplaceAllFunc(b, func(sp []byte) []byte {
				// The shared shape writer emits an empty line for Type:none. Make
				// no outline explicit in v2 so PowerPoint cannot inherit a theme line.
				if !bytes.Contains(sp, []byte("<p:txBody>")) {
					return bytes.ReplaceAll(sp, []byte("<a:ln></a:ln>"), []byte("<a:ln><a:noFill/></a:ln>"))
				}

				match := objectName.FindSubmatch(sp)
				if match == nil {
					return sp
				}
				// Attribute values use the same XML escaping as text. Decode through XML.
				var named struct {
					Name string `xml:"name,attr"`
				}
				if e := xml.Unmarshal([]byte(`<n name="`+string(match[1])+`"/>`), &named); e != nil {
					rewriteErr = e
					return sp
				}
				r, ok := byName[named.Name]
				if !ok {
					return sp
				}
				old := paragraphs.FindAll(sp, -1)
				if len(old) == 0 {
					rewriteErr = fmt.Errorf("text.missing_paragraph: %s", r.ID)
					return sp
				}
				pp := ppr.Find(old[0])
				if pp == nil {
					rewriteErr = fmt.Errorf("text.missing_paragraph_properties: %s", r.ID)
					return sp
				}
				if r.Rich != nil {
					text := richParagraphXML(r, pp)
					first, last := bytes.Index(sp, old[0]), bytes.LastIndex(sp, old[len(old)-1])+len(old[len(old)-1])
					updated := append([]byte{}, sp[:first]...)
					updated = append(updated, text...)
					updated = append(updated, sp[last:]...)
					return updated
				}
				s, id := r.Layout.Style, r.Layout.Font
				boolean := func(v bool) string {
					if v {
						return "1"
					}
					return "0"
				}
				attrs := ` lang="en-US" sz="` + strconv.Itoa(int(math.Round(s.Size*100))) + `" spc="` + strconv.Itoa(int(math.Round(s.TrackingPt*100))) + `" kern="0" b="` + boolean(id.Bold) + `" i="` + boolean(id.NativeItalic) + `" dirty="0"`
				children := `<a:solidFill><a:srgbClr val="` + escape(r.Color) + `"/></a:solidFill><a:latin typeface="` + escape(id.Typeface) + `"/><a:ea typeface="` + escape(id.Typeface) + `"/><a:cs typeface="` + escape(id.Typeface) + `"/>`
				var text strings.Builder
				for _, line := range strings.Split(r.Layout.Displayed, "\n") {
					text.WriteString("<a:p>")
					text.Write(pp)
					if line != "" {
						text.WriteString("<a:r><a:rPr" + attrs + ">" + children + "</a:rPr><a:t>" + escape(line) + "</a:t></a:r>")
					}
					text.WriteString("<a:endParaRPr" + attrs + ">" + children + "</a:endParaRPr></a:p>")
				}
				first, last := bytes.Index(sp, old[0]), bytes.LastIndex(sp, old[len(old)-1])+len(old[len(old)-1])
				updated := append([]byte{}, sp[:first]...)
				updated = append(updated, []byte(text.String())...)
				updated = append(updated, sp[last:]...)
				return updated
			})
		}
		if rewriteErr != nil {
			return nil, rewriteErr
		}
		h := f.FileHeader
		dst, e := w.CreateHeader(&h)
		if e != nil {
			return nil, e
		}
		if _, e = dst.Write(b); e != nil {
			return nil, e
		}
	}
	if e := w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
