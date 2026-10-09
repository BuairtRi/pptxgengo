package deckproject

import (
	"bytes"
	"strings"
	"testing"
)

const saveAsRunProperties = `<a:solidFill><a:srgbClr val="FFFFFF"/></a:solidFill><a:latin typeface="IBM Plex Sans"/>`
const saveAsRunAttributes = `lang="en-US" sz="1200" b="0" i="0"`
const saveAsEmptySurfaceBody = `<p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:endParaRPr lang="en-US"/></a:p></p:txBody>`
const saveAsCreationExtension = `<a:extLst><a:ext uri="{FF2B5EF4-FFF2-40B4-BE49-F238E27FC236}"><a16:creationId id="{02A69C4C-74CE-4A1D-0865-E03437109085}"/></a:ext></a:extLst>`

func saveAsXMLFixture(office bool) string {
	id, body, end := "", "", `<a:endParaRPr `+saveAsRunAttributes+`>`+saveAsRunProperties+`</a:endParaRPr>`
	if office {
		id, body, end = saveAsCreationExtension, saveAsEmptySurfaceBody, ""
	}
	return `<p:grpSp xmlns:p="` + lineagePML + `" xmlns:a="` + drawingML + `" xmlns:a16="` + officeCreationIDNamespace + `"><p:nvGrpSpPr><p:cNvPr id="1" name="node">` + id + `</p:cNvPr></p:nvGrpSpPr><p:grpSpPr/><p:sp><p:nvSpPr><p:cNvPr id="2" name="node.surface">` + id + `</p:cNvPr></p:nvSpPr><p:spPr><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:solidFill><a:srgbClr val="070154"/></a:solidFill></p:spPr>` + body + `</p:sp><p:sp><p:nvSpPr><p:cNvPr id="3" name="node.text">` + id + `</p:cNvPr></p:nvSpPr><p:spPr/><p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:pPr/><a:r><a:rPr ` + saveAsRunAttributes + `>` + saveAsRunProperties + `</a:rPr><a:t>Service</a:t></a:r>` + end + `</a:p></p:txBody></p:sp></p:grpSp>`
}
func saveAsTree(t *testing.T, raw string) *xmlNode {
	t.Helper()
	n, e := readLineageXML([]byte(raw))
	if e != nil {
		t.Fatal(e)
	}
	return n
}
func TestSaveAsGeometryEquivalencePreservesExactInputs(t *testing.T) {
	a, b := saveAsTree(t, saveAsXMLFixture(false)), saveAsTree(t, saveAsXMLFixture(true))
	beforeA, beforeB := canonical(a), canonical(b)
	if !geometryOnlyChange(a, b) || !geometryOnlyChange(b, a) {
		t.Fatal("known serialization rewrite rejected", geometryCompatibilityMismatch(a, b))
	}
	if !bytes.Equal(beforeA, canonical(a)) || !bytes.Equal(beforeB, canonical(b)) {
		t.Fatal("analysis rewrote raw parsed input")
	}
	if reconcileStructureHash(a) == reconcileStructureHash(b) {
		t.Fatal("strict non-geometry text structure comparison was weakened")
	}
}
func TestSaveAsEquivalenceDoesNotHideMeaningfulNativeEdits(t *testing.T) {
	base := saveAsXMLFixture(false)
	office := saveAsXMLFixture(true)
	for name, modified := range map[string]string{
		"unknown extension":        strings.ReplaceAll(office, officeCreationIDExtension, "{UNKNOWN}"),
		"creation extra attribute": strings.ReplaceAll(office, `creationId id=`, `creationId custom="1" id=`),
		"creation invalid guid":    strings.ReplaceAll(office, "02A69C4C-74CE-4A1D-0865-E03437109085", "invalid"),
		"empty body style":         strings.Replace(office, saveAsEmptySurfaceBody, strings.Replace(saveAsEmptySurfaceBody, `<a:bodyPr/>`, `<a:bodyPr anchor="ctr"/>`, 1), 1),
		"empty body font":          strings.Replace(office, saveAsEmptySurfaceBody, strings.Replace(saveAsEmptySurfaceBody, `lang="en-US"`, `lang="en-US" sz="1800"`, 1), 1),
		"new surface text":         strings.Replace(office, saveAsEmptySurfaceBody, `<p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr/><a:t>New text</a:t></a:r></a:p></p:txBody>`, 1),
		"actual font":              strings.ReplaceAll(office, "IBM Plex Sans", "Arial"),
		"actual size":              strings.ReplaceAll(office, `sz="1200"`, `sz="1800"`),
		"actual text color":        strings.ReplaceAll(office, `val="FFFFFF"`, `val="FF0000"`),
		"actual shape paint":       strings.ReplaceAll(office, `val="070154"`, `val="FF0000"`),
		"actual preset":            strings.ReplaceAll(office, `prst="rect"`, `prst="ellipse"`),
		"nonredundant end style":   strings.Replace(office, `</a:r></a:p>`, `</a:r><a:endParaRPr lang="en-US" sz="1800"/></a:p>`, 1),
		"new rich run":             strings.Replace(office, `</a:r></a:p>`, `</a:r><a:r><a:rPr b="1"/><a:t>Extra</a:t></a:r></a:p>`, 1),
		"empty extension wrapper":  strings.Replace(office, saveAsCreationExtension, `<a:extLst/>`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			a, b := saveAsTree(t, base), saveAsTree(t, modified)
			if mismatch := geometryCompatibilityMismatch(a, b); mismatch == "" || !strings.Contains(mismatch, "/grpSp/") {
				t.Fatal("unsupported change accepted or mismatch path absent", mismatch)
			}
		})
	}
	// Missing end properties are not equivalent if the original end style
	// differs from its last run, even when the rendered current text matches.
	end := `<a:endParaRPr ` + saveAsRunAttributes + `>` + saveAsRunProperties + `</a:endParaRPr>`
	nonredundant := strings.Replace(base, end, strings.Replace(end, `sz="1200"`, `sz="1600"`, 1), 1)
	if geometryOnlyChange(saveAsTree(t, nonredundant), saveAsTree(t, office)) {
		t.Fatal("meaningful end properties silently removed")
	}
}
