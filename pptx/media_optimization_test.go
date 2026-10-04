package pptx

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"sort"
	"strings"
	"testing"

	"golang.org/x/image/draw"
)

func TestMediaLargeResamplingMatchesCubicScale(t *testing.T) {
	// This size exceeds the 16 MiB separable scratch threshold. Mix gradients
	// and narrow stripes to expose edge alignment and antialiasing differences.
	im := image.NewRGBA(image.Rect(0, 0, 3200, 1200))
	for y := 0; y < 1200; y++ {
		for x := 0; x < 3200; x++ {
			im.SetRGBA(x, y, color.RGBA{R: byte(x / 13), G: byte(y / 5), B: byte((x % 7) * 40), A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, im, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	source, err := jpeg.Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	got := image.NewRGBA(image.Rect(0, 0, 500, 188))
	want := image.NewRGBA(got.Bounds())
	scaleMediaJPEG(got, source)
	draw.CatmullRom.Scale(want, want.Bounds(), source, source.Bounds(), draw.Src, nil)
	for i, value := range got.Pix {
		delta := int(value) - int(want.Pix[i])
		if delta < -1 || delta > 1 {
			t.Fatalf("cubic resampling differs at channel %d: %d versus %d", i, value, want.Pix[i])
		}
	}
}

func TestMediaContentTypesIgnoreExtensionAttributes(t *testing.T) {
	input := []byte(`<Types xmlns="` + packageContentTypesNS + `" xmlns:x="urn:extension"><Default Extension="png" ContentType="image/png" x:ContentType="unrelated"/><Override PartName="/ppt/media/a.png" ContentType="image/png" x:PartName="/ppt/media/b.png"/><Override PartName="/ppt/media/b.png" ContentType="image/png"/><x:Override PartName="/ppt/media/b.png" ContentType="extension-value"/></Types>`)
	types, err := readMediaContentTypes(input)
	if err != nil || types["ext:png"] != "image/png" || types["ppt/media/a.png"] != "image/png" || types["ppt/media/b.png"] != "image/png" {
		t.Fatalf("extension attributes changed content types: %v, %v", types, err)
	}
	output, err := removeMediaOverrides(input, map[string]bool{"ppt/media/b.png": true})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output, []byte(`PartName="/ppt/media/a.png"`)) || !bytes.Contains(output, []byte(`<x:Override PartName="/ppt/media/b.png"`)) {
		t.Fatalf("removed a retained or extension override: %s", output)
	}
	types, err = readMediaContentTypes(output)
	if err != nil || types["ppt/media/b.png"] != "" {
		t.Fatalf("obsolete standard override remained: %v, %v", types, err)
	}
}

func TestMediaGroupAndRotationSizing(t *testing.T) {
	for _, tc := range []struct {
		name, groupRotation, pictureRotation string
		width, height                        float64
	}{
		{"scaled", "", "", 2, 3},
		{"rotated-group", `rot="5400000"`, "", 3, 3},
		{"rotated-picture", "", `rot="5400000"`, 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pic := optimizationPicture("rId1", 914400, 914400, "")
			if tc.pictureRotation != "" {
				pic = strings.Replace(pic, `<a:xfrm>`, `<a:xfrm `+tc.pictureRotation+`>`, 1)
			}
			group := `<p:grpSp><p:grpSpPr><a:xfrm ` + tc.groupRotation + `><a:ext cx="2" cy="3"/><a:chExt cx="1" cy="1"/></a:xfrm></p:grpSpPr>` + pic + `</p:grpSp>`
			root, err := readMediaXML(optimizationSlide(group))
			if err != nil {
				t.Fatal(err)
			}
			seen := map[string]mediaUsage{}
			collectMediaUsage(root, 1, 1, seen)
			got := seen["rId1"]
			if got.unsafe || !got.known || got.width != tc.width || got.height != tc.height {
				t.Fatalf("unsafe or undersized transform bound: %+v", got)
			}
		})
	}
}

func optimizationFixture(t *testing.T, parts map[string][]byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	names := make([]string, 0, len(parts))
	for name := range parts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		f, err := w.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write(parts[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func optimizationParts(t *testing.T, raw []byte) map[string][]byte {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	parts := map[string][]byte{}
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		parts[f.Name] = data
	}
	return parts
}

func optimizationJPEG(t *testing.T) []byte {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, 1200, 800))
	state := uint32(7)
	for y := 0; y < 800; y++ {
		for x := 0; x < 1200; x++ {
			state = state*1664525 + 1013904223
			im.SetRGBA(x, y, color.RGBA{R: byte(state >> 24), G: byte(state >> 16), B: byte(state >> 8), A: 255})
		}
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, im, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func optimizationPicture(id string, width, height int, crop string) string {
	return fmt.Sprintf(`<p:pic><p:nvPicPr><p:cNvPr id="1" name="stable-object"/></p:nvPicPr><p:blipFill><a:blip r:embed="%s"/>%s<a:stretch><a:fillRect/></a:stretch></p:blipFill><p:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm></p:spPr></p:pic>`, id, crop, width, height)
}

func optimizationSlide(body string) []byte {
	return []byte(`<p:sld xmlns:p="` + presentationNS + `" xmlns:a="` + drawingNS + `" xmlns:r="` + relationshipNS + `"><p:cSld><p:spTree>` + body + `</p:spTree></p:cSld></p:sld>`)
}

func optimizationRelations(body string) []byte {
	return []byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships" xmlns:x="urn:example:extension">` + body + `</Relationships>`)
}

func optimizationRelationship(id, target string) string {
	return `<Relationship Id="` + id + `" Type="` + relationshipNS + `/image" Target="` + target + `"/>`
}

func TestOptimizeMediaNamespaceTargetRegression(t *testing.T) {
	for _, attrs := range []string{
		`x:Target="extension-value" Target="../media/b.png"`,
		`Target="../media/b.png" x:Target="extension-value"`,
		`x:note=" Target='extension-value' " x:Target="extension-value" Target="../media/b.png"`,
		`x:note=' Target="extension-value" ' x:Target="extension-value" Target='../media/b.png'`,
	} {
		t.Run(attrs, func(t *testing.T) {
			parts := map[string][]byte{
				"[Content_Types].xml": []byte(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="png" ContentType="image/png"/></Types>`),
				"ppt/media/a.png":     tinyPNGBytes(t), "ppt/media/b.png": tinyPNGBytes(t),
				"ppt/slides/slide1.xml":            optimizationSlide(""),
				"ppt/slides/_rels/slide1.xml.rels": optimizationRelations(optimizationRelationship("rId1", "../media/a.png") + `<Relationship Id="rId2" Type="` + relationshipNS + `/image" ` + attrs + `/>`),
			}
			output, report, err := OptimizeMedia(optimizationFixture(t, parts), DeliveryMediaOptions())
			if err != nil {
				t.Fatal(err)
			}
			if report.OutputMediaParts != 1 {
				t.Fatalf("media parts=%d, want 1", report.OutputMediaParts)
			}
			out := optimizationParts(t, output)
			d := xml.NewDecoder(bytes.NewReader(out["ppt/slides/_rels/slide1.xml.rels"]))
			found := false
			for {
				token, err := d.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				start, ok := token.(xml.StartElement)
				if !ok || start.Name.Local != "Relationship" {
					continue
				}
				id, target, extension := "", "", ""
				for _, attr := range start.Attr {
					if attr.Name.Space == "" && attr.Name.Local == "Id" {
						id = attr.Value
					}
					if attr.Name.Space == "" && attr.Name.Local == "Target" {
						target = attr.Value
					}
					if attr.Name.Space == "urn:example:extension" && attr.Name.Local == "Target" {
						extension = attr.Value
					}
				}
				if id == "rId2" {
					found = true
					if target != "../media/a.png" || extension != "extension-value" {
						t.Fatalf("target=%q extension=%q", target, extension)
					}
				}
			}
			if !found {
				t.Fatal("lost relationship identity")
			}
			if !bytes.Equal(out["ppt/slides/slide1.xml"], parts["ppt/slides/slide1.xml"]) {
				t.Fatal("slide content changed")
			}
		})
	}
}

func TestOptimizeMediaLargestUsageAndCrop(t *testing.T) {
	photo := optimizationJPEG(t)
	parts := map[string][]byte{
		"[Content_Types].xml": []byte(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="jpg" ContentType="image/jpg"/><Default Extension="jpeg" ContentType="image/jpeg"/></Types>`),
		"ppt/media/a.jpg":     photo, "ppt/media/b.jpeg": photo,
		"ppt/slides/slide1.xml":            optimizationSlide(optimizationPicture("rId1", 914400, 609600, `<a:srcRect l="25000" r="25000"/>`)),
		"ppt/slides/slide2.xml":            optimizationSlide(optimizationPicture("rId1", 2743200, 1828800, "")),
		"ppt/slides/_rels/slide1.xml.rels": optimizationRelations(optimizationRelationship("rId1", "../media/a.jpg")),
		"ppt/slides/_rels/slide2.xml.rels": optimizationRelations(optimizationRelationship("rId1", "../media/b.jpeg")),
	}
	output, report, err := OptimizeMedia(optimizationFixture(t, parts), DeliveryMediaOptions())
	if err != nil {
		t.Fatal(err)
	}
	if report.OutputMediaParts != 1 || len(report.Parts) != 2 {
		t.Fatalf("unexpected receipt %+v", report)
	}
	for _, part := range report.Parts {
		if !part.Resized || part.RequiredWidth != 660 || part.RequiredHeight != 440 || part.OutputWidth < 660 || part.OutputHeight < 440 {
			t.Fatalf("undersized or unexpected derivative %+v", part)
		}
		if part.SourceSHA256 != mediaHash(photo) || part.OutputSHA256 == part.SourceSHA256 {
			t.Fatal("invalid derivative provenance")
		}
	}
	out := optimizationParts(t, output)
	if !bytes.Equal(out["ppt/slides/slide1.xml"], parts["ppt/slides/slide1.xml"]) {
		t.Fatal("crop or slide XML changed")
	}
	z, _ := zip.NewReader(bytes.NewReader(output), int64(len(output)))
	for _, f := range z.File {
		if f.Method != zip.Deflate {
			t.Fatalf("part %s is not compressed", f.Name)
		}
	}
	// Optimizing an already compressed package must preserve payloads and
	// validate DEFLATE streams rather than assuming every input is STORE.
	again, _, err := OptimizeMedia(output, DeliveryMediaOptions())
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range optimizationParts(t, again) {
		if !bytes.Equal(data, out[name]) {
			t.Fatalf("second pass changed %s", name)
		}
	}
}

func TestOptimizeMediaPreservesUnsafeAndSensitiveJPEG(t *testing.T) {
	photo := optimizationJPEG(t)
	payload := []byte("ICC_PROFILE\x00placeholder")
	segment := []byte{0xff, 0xe2, 0, 0}
	binary.BigEndian.PutUint16(segment[2:], uint16(len(payload)+2))
	profiled := append(append(append([]byte{}, photo[:2]...), segment...), payload...)
	profiled = append(profiled, photo[2:]...)
	for _, tc := range []struct {
		name  string
		photo []byte
		body  string
	}{
		{"unknown", photo, ""},
		{"negative-crop", photo, optimizationPicture("rId1", 914400, 609600, `<a:srcRect l="-100"/>`)},
		{"color-profile", profiled, optimizationPicture("rId1", 914400, 609600, "")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parts := map[string][]byte{
				"[Content_Types].xml":              []byte(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="jpg" ContentType="image/jpeg"/></Types>`),
				"ppt/media/a.jpg":                  tc.photo,
				"ppt/slides/slide1.xml":            optimizationSlide(tc.body),
				"ppt/slides/_rels/slide1.xml.rels": optimizationRelations(optimizationRelationship("rId1", "../media/a.jpg")),
			}
			output, report, err := OptimizeMedia(optimizationFixture(t, parts), DeliveryMediaOptions())
			if err != nil {
				t.Fatal(err)
			}
			if report.Parts[0].Resized || !bytes.Equal(optimizationParts(t, output)["ppt/media/a.jpg"], tc.photo) {
				t.Fatal("unsafe image changed")
			}
		})
	}
}

func TestOptimizeMediaRejectsInvalidPolicy(t *testing.T) {
	options := DeliveryMediaOptions()
	options.PixelsPerInch = -1
	if _, _, err := OptimizeMedia(nil, options); err == nil || !strings.Contains(err.Error(), "policy") {
		t.Fatalf("expected policy error, got %v", err)
	}
}

func TestOptimizeMediaPreservesCompressedEmptyPart(t *testing.T) {
	var input bytes.Buffer
	w := zip.NewWriter(&input)
	if _, err := w.CreateHeader(&zip.FileHeader{Name: "empty/", Method: zip.Deflate}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out, _, err := OptimizeMedia(input.Bytes(), DeliveryMediaOptions())
	if err != nil {
		t.Fatal(err)
	}
	if data, ok := optimizationParts(t, out)["empty/"]; !ok || len(data) != 0 {
		t.Fatal("empty part changed")
	}
}

func TestOptimizeMediaRejectsCorruptStoredPart(t *testing.T) {
	raw := optimizationFixture(t, map[string][]byte{"ppt/media/a.png": tinyPNGBytes(t)})
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	offset, err := z.File[0].DataOffset()
	if err != nil {
		t.Fatal(err)
	}
	raw[offset] ^= 1
	if _, _, err := OptimizeMedia(raw, DeliveryMediaOptions()); err == nil {
		t.Fatal("corrupted payload accepted")
	}
}
