package nativepkg

import (
	"strings"
	"testing"
)

const bindingFixture = `<p:sld xmlns:p="urn:p" xmlns:a="urn:a">
  <p:cSld><p:spTree>
    <p:sp><p:nvSpPr><p:cNvPr id="7" name="Metric"/></p:nvSpPr>
      <p:spPr><a:xfrm rot="0"><a:off x="100" y="200"/><a:ext cx="300" cy="400"/></a:xfrm><a:prstGeom prst="rect"/><a:solidFill><a:srgbClr val="001122"/></a:solidFill><a:ln w="12700"><a:prstDash val="solid"/></a:ln></p:spPr>
      <p:txBody><a:bodyPr wrap="square"/><a:p><a:pPr algn="l"/><a:r><a:rPr sz="1200" b="0"><a:solidFill><a:schemeClr val="accent1"/></a:solidFill></a:rPr><a:t>Original metric</a:t></a:r></a:p></p:txBody>
    </p:sp>
    <p:pic><p:nvPicPr><p:cNvPr id="8" name="Portrait"/></p:nvPicPr><p:blipFill><a:srcRect l="1" t="2" r="3" b="4"/></p:blipFill></p:pic>
    <p:graphicFrame><p:nvGraphicFramePr><p:cNvPr id="9" name="Table"/></p:nvGraphicFramePr><a:graphic><a:graphicData><a:tbl><a:tblGrid><a:gridCol w="500"/></a:tblGrid><a:tr h="600"><a:tc><a:tcPr marL="7"/></a:tc></a:tr></a:tbl></a:graphicData></a:graphic></p:graphicFrame>
  </p:spTree></p:cSld>
</p:sld>`

func TestBindingsDriveTextGeometryAndFont(t *testing.T) {
	scene, err := Parse([]byte(bindingFixture))
	if err != nil {
		t.Fatal(err)
	}
	bindings := ExtractBindings(scene)
	if len(bindings) == 0 {
		t.Fatal("expected bindings")
	}
	for i := range bindings {
		b := &bindings[i]
		if b.ObjectID == "7" && b.Property == "text" && b.Value == "Original metric" {
			b.Value = "Changed metric"
		}
		if b.ObjectID == "7" && b.Property == "transform.off.x" {
			b.Value = "91540"
		}
		if b.ObjectID == "7" && b.Property == "style.rPr.rPr.sz" {
			b.Value = "1800"
		}
	}
	if err := ApplyBindings(scene, bindings); err != nil {
		t.Fatal(err)
	}
	xml := string(scene.XML())
	for _, want := range []string{`>Changed metric<`, `x="91540"`, `sz="1800"`, `l="1"`, `w="500"`, `h="600"`, `marL="7"`} {
		if !strings.Contains(xml, want) {
			t.Errorf("rebuilt XML missing %s\n%s", want, xml)
		}
	}
	if strings.Contains(xml, bindingSentinelPrefix) {
		t.Errorf("rebuilt XML retains binding sentinel: %s", xml)
	}
}

func TestApplyBindingsRejectsMissingBinding(t *testing.T) {
	scene, err := Parse([]byte(bindingFixture))
	if err != nil {
		t.Fatal(err)
	}
	bindings := ExtractBindings(scene)
	var partial []Binding
	for _, b := range bindings {
		if b.ObjectID == "7" && b.Property == "text" {
			continue
		}
		partial = append(partial, b)
	}
	if err := ApplyBindings(scene, partial); err == nil || !strings.Contains(err.Error(), "missing binding") {
		t.Fatalf("missing binding error = %v, want missing binding", err)
	}
}
