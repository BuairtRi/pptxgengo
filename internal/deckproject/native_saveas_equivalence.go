package deckproject

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const officeCreationIDNamespace = "http://schemas.microsoft.com/office/drawing/2014/main"
const officeCreationIDExtension = "{FF2B5EF4-FFF2-40B4-BE49-F238E27FC236}"

var officeCreationGUID = regexp.MustCompile(`^\{[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}\}$`)
var officeEmptyTextLanguage = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

// Normalize an analysis view only. Raw PPTX, ownership tags, report inputs,
// paragraph baselines and retained evidence are never rewritten. This is not a
// general Office XML importer or permission to ignore meaningful formatting.
func normalizePowerPointSerialization(n *xmlNode) *xmlNode {
	c := *n
	c.Children = nil
	for _, child := range n.Children {
		if n.Name.Space == lineagePML && n.Name.Local == "sp" && isOfficeEmptySurfaceText(child) {
			continue
		}
		if n.Name.Space == drawingML && n.Name.Local == "p" && child.Name.Space == drawingML && child.Name.Local == "endParaRPr" && redundantPlainParagraphEnd(n, child) {
			continue
		}
		if n.Name.Space == lineagePML && n.Name.Local == "cNvPr" && child.Name.Space == drawingML && child.Name.Local == "extLst" && noNativeAttributes(child) && strings.TrimSpace(child.Text) == "" {
			x := *child
			x.Children = nil
			removed := false
			for _, ext := range child.Children {
				if isOfficeCreationExtension(ext) {
					removed = true
				} else {
					x.Children = append(x.Children, normalizePowerPointSerialization(ext))
				}
			}
			if !removed || len(x.Children) != 0 {
				c.Children = append(c.Children, &x)
			}
			continue
		}
		c.Children = append(c.Children, normalizePowerPointSerialization(child))
	}
	return &c
}

func noNativeAttributes(n *xmlNode) bool {
	for _, a := range n.Attrs {
		if a.Name.Space != "xmlns" && a.Name.Local != "xmlns" {
			return false
		}
	}
	return true
}
func oneNativeAttribute(n *xmlNode, key string) (string, bool) {
	v, count := "", 0
	for _, a := range n.Attrs {
		if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
			continue
		}
		if a.Name.Space != "" || a.Name.Local != key {
			return "", false
		}
		v, count = a.Value, count+1
	}
	return v, count == 1
}
func isOfficeCreationExtension(n *xmlNode) bool {
	uri, ok := oneNativeAttribute(n, "uri")
	if n.Name.Space != drawingML || n.Name.Local != "ext" || !ok || uri != officeCreationIDExtension || strings.TrimSpace(n.Text) != "" || len(n.Children) != 1 {
		return false
	}
	id := n.Children[0]
	value, ok := oneNativeAttribute(id, "id")
	return id.Name.Space == officeCreationIDNamespace && id.Name.Local == "creationId" && ok && officeCreationGUID.MatchString(value) && len(id.Children) == 0 && strings.TrimSpace(id.Text) == ""
}
func isOfficeEmptySurfaceText(n *xmlNode) bool {
	if n.Name.Space != lineagePML || n.Name.Local != "txBody" || !noNativeAttributes(n) || strings.TrimSpace(n.Text) != "" || len(n.Children) != 3 {
		return false
	}
	for i, name := range []string{"bodyPr", "lstStyle", "p"} {
		x := n.Children[i]
		if x.Name.Space != drawingML || x.Name.Local != name || !noNativeAttributes(x) || strings.TrimSpace(x.Text) != "" || (i < 2 && len(x.Children) != 0) {
			return false
		}
	}
	p := n.Children[2]
	if len(p.Children) != 1 {
		return false
	}
	end := p.Children[0]
	language, ok := oneNativeAttribute(end, "lang")
	return end.Name.Space == drawingML && end.Name.Local == "endParaRPr" && ok && officeEmptyTextLanguage.MatchString(language) && len(end.Children) == 0 && strings.TrimSpace(end.Text) == ""
}
func redundantPlainParagraphEnd(p, end *xmlNode) bool {
	var run, properties *xmlNode
	ends := 0
	for _, c := range p.Children {
		if c.Name.Space != drawingML {
			return false
		}
		switch c.Name.Local {
		case "r":
			if run != nil {
				return false
			}
			run = c
		case "pPr":
		case "endParaRPr":
			ends++
		default:
			return false
		}
	}
	if run == nil || ends != 1 || len(run.Children) != 2 || run.Children[0].Name.Space != drawingML || run.Children[0].Name.Local != "rPr" || run.Children[1].Name.Space != drawingML || run.Children[1].Name.Local != "t" || len(run.Children[1].Children) != 0 || strings.TrimSpace(run.Children[1].Text) == "" {
		return false
	}
	properties = run.Children[0]
	copyEnd := *end
	copyEnd.Name = properties.Name
	return bytes.Equal(canonical(reconcileStructureView(properties)), canonical(reconcileStructureView(&copyEnd)))
}

// Describe location/category without leaking slide text or arbitrary XML values.
func nativeViewMismatch(a, b *xmlNode, path string) string {
	if a.Name != b.Name {
		return path + ": element differs"
	}
	if !bytes.Equal(canonical(a.Attrs), canonical(b.Attrs)) {
		return path + ": attributes differ"
	}
	if a.Text != b.Text {
		return path + ": nontext payload differs"
	}
	if len(a.Children) != len(b.Children) {
		return fmt.Sprintf("%s: child count %d != %d", path, len(a.Children), len(b.Children))
	}
	for i := range a.Children {
		if mismatch := nativeViewMismatch(a.Children[i], b.Children[i], fmt.Sprintf("%s/%s[%d]", path, a.Children[i].Name.Local, i)); mismatch != "" {
			return mismatch
		}
	}
	return ""
}
