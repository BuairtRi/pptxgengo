package browsingartifact

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

const MaxManifestBytes = 16 << 20
const maxXMLBytes = 8 << 20
const presentationNS = "http://schemas.openxmlformats.org/presentationml/2006/main"
const officeNS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const packageNS = "http://schemas.openxmlformats.org/package/2006/relationships"

func ReadManifest(file string) ([]byte, error) {
	f, e := os.Open(file)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	raw, e := io.ReadAll(io.LimitReader(f, MaxManifestBytes+1))
	if e != nil {
		return nil, e
	}
	if len(raw) > MaxManifestBytes {
		return nil, fmt.Errorf("browsing.manifest_exceeds_16MiB")
	}
	return raw, nil
}
func Slides(file string) (int, error) { pages, e := PresentationPages(file); return len(pages), e }

// PresentationPages follows sldIdLst internal relationships and rejects duplicate,
// dangling, external and orphan slide parts. It returns native cSld names in the
// exact presentation order. Limits cover the full developer matrix, not only the
// convenient catalog browsing package.
func PresentationPages(file string) ([]string, error) {
	info, e := os.Lstat(file)
	if e != nil {
		return nil, e
	}
	if !info.Mode().IsRegular() || info.Size() > 512<<20 {
		return nil, fmt.Errorf("browsing.pptx_size_or_type_invalid")
	}
	z, e := zip.OpenReader(file)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	if len(z.File) > 100000 {
		return nil, fmt.Errorf("browsing.pptx_member_limit")
	}
	files := map[string]*zip.File{}
	folded := map[string]bool{}
	var total uint64
	for _, f := range z.File {
		name := strings.TrimSuffix(f.Name, "/")
		if !portable(name) || folded[strings.ToLower(f.Name)] || f.Mode()&os.ModeSymlink != 0 || f.UncompressedSize64 > 64<<20 {
			return nil, fmt.Errorf("browsing.pptx_member_invalid")
		}
		folded[strings.ToLower(f.Name)] = true
		total += f.UncompressedSize64
		if total > 4<<30 {
			return nil, fmt.Errorf("browsing.pptx_uncompressed_limit")
		}
		files[f.Name] = f
	}
	decode := func(name string, value any) error {
		f := files[name]
		if f == nil || f.UncompressedSize64 > maxXMLBytes {
			return fmt.Errorf("browsing.xml_missing_or_too_large: %s", name)
		}
		r, e := f.Open()
		if e != nil {
			return e
		}
		defer r.Close()
		raw, e := io.ReadAll(io.LimitReader(r, maxXMLBytes+1))
		if e != nil {
			return e
		}
		if len(raw) > maxXMLBytes {
			return fmt.Errorf("browsing.xml_limit")
		}
		d := xml.NewDecoder(strings.NewReader(string(raw)))
		if e = d.Decode(value); e != nil {
			return e
		}
		for {
			token, e := d.Token()
			if e == io.EOF {
				return nil
			}
			if e != nil {
				return e
			}
			if data, ok := token.(xml.CharData); ok && strings.TrimSpace(string(data)) == "" {
				continue
			}
			return fmt.Errorf("browsing.xml_trailing_content")
		}
	}
	var p struct {
		XMLName xml.Name `xml:"presentation"`
		Slides  []struct {
			XMLName xml.Name   `xml:"sldId"`
			Attrs   []xml.Attr `xml:",any,attr"`
		} `xml:"sldIdLst>sldId"`
	}
	if e = decode("ppt/presentation.xml", &p); e != nil {
		return nil, e
	}
	if p.XMLName.Space != presentationNS || len(p.Slides) < 2 || len(p.Slides) > 20000 {
		return nil, fmt.Errorf("browsing.presentation_slide_limit_or_namespace")
	}
	type relation struct {
		XMLName xml.Name `xml:"Relationship"`
		ID      string   `xml:"Id,attr"`
		Type    string   `xml:"Type,attr"`
		Target  string   `xml:"Target,attr"`
		Mode    string   `xml:"TargetMode,attr"`
	}
	var rels struct {
		XMLName xml.Name   `xml:"Relationships"`
		Items   []relation `xml:"Relationship"`
	}
	if e = decode("ppt/_rels/presentation.xml.rels", &rels); e != nil {
		return nil, e
	}
	if rels.XMLName.Space != packageNS {
		return nil, fmt.Errorf("browsing.relationship_namespace_invalid")
	}
	targets := map[string]string{}
	slideRelations := map[string]bool{}
	for _, r := range rels.Items {
		if r.XMLName.Space != packageNS || r.ID == "" || targets[r.ID] != "" || r.Mode != "" && r.Mode != "Internal" {
			return nil, fmt.Errorf("browsing.relationship_duplicate_or_external")
		}
		u, e := url.Parse(r.Target)
		if e != nil || u.Scheme != "" || u.Host != "" || u.RawQuery != "" || u.Fragment != "" || u.Path == "" || strings.Contains(u.Path, "\\") {
			return nil, fmt.Errorf("browsing.relationship_target_invalid")
		}
		name := path.Clean(path.Join("ppt", u.Path))
		if strings.HasPrefix(u.Path, "/") {
			name = path.Clean(strings.TrimPrefix(u.Path, "/"))
		}
		if !portable(name) || files[name] == nil {
			return nil, fmt.Errorf("browsing.relationship_dangling: %s", r.ID)
		}
		targets[r.ID] = name
		if r.Type == officeNS+"/slide" {
			slideRelations[r.ID] = true
		}
	}
	usedIDs := map[string]bool{}
	usedParts := map[string]bool{}
	names := map[string]bool{}
	numericIDs := map[string]bool{}
	pages := []string{}
	for _, item := range p.Slides {
		if item.XMLName.Space != presentationNS {
			return nil, fmt.Errorf("browsing.slide_reference_namespace_invalid")
		}
		rid, id := "", ""
		for _, a := range item.Attrs {
			if a.Name.Local != "id" {
				continue
			}
			if a.Name.Space == officeNS {
				if rid != "" {
					return nil, fmt.Errorf("browsing.slide_reference_duplicate")
				}
				rid = a.Value
			} else if a.Name.Space == "" {
				id = a.Value
			}
		}
		part := targets[rid]
		if rid == "" || id == "" || numericIDs[id] || usedIDs[rid] || part == "" || !slideRelations[rid] || usedParts[part] || path.Dir(part) != "ppt/slides" {
			return nil, fmt.Errorf("browsing.slide_reference_duplicate_or_dangling")
		}
		usedIDs[rid] = true
		numericIDs[id] = true
		usedParts[part] = true
		var slide struct {
			XMLName xml.Name `xml:"sld"`
			Common  struct {
				XMLName xml.Name `xml:"cSld"`
				Name    string   `xml:"name,attr"`
			} `xml:"cSld"`
		}
		if e = decode(part, &slide); e != nil {
			return nil, e
		}
		if slide.XMLName.Space != presentationNS || slide.Common.XMLName.Space != presentationNS || slide.Common.Name == "" || names[slide.Common.Name] {
			return nil, fmt.Errorf("browsing.slide_root_or_identity_invalid")
		}
		names[slide.Common.Name] = true
		pages = append(pages, slide.Common.Name)
	}
	for rid := range slideRelations {
		if !usedIDs[rid] {
			return nil, fmt.Errorf("browsing.orphan_slide_relationship")
		}
	}
	for name := range files {
		if path.Dir(name) == "ppt/slides" && strings.HasSuffix(name, ".xml") && !usedParts[name] {
			return nil, fmt.Errorf("browsing.orphan_slide_part")
		}
	}
	return pages, nil
}
func ValidateFile(raw []byte, kind, deckHash, file string) error {
	if len(raw) > MaxManifestBytes {
		return fmt.Errorf("browsing.manifest_exceeds_16MiB")
	}
	pages, e := PresentationPages(file)
	if e != nil {
		return e
	}
	if e = Validate(raw, kind, deckHash, len(pages)); e != nil {
		return e
	}
	var m Manifest
	if e = json.Unmarshal(raw, &m); e != nil {
		return e
	}
	for i, page := range m.Pages {
		if page.ID != pages[i] {
			return fmt.Errorf("browsing.page_order_or_identity_mismatch")
		}
	}
	return nil
}
