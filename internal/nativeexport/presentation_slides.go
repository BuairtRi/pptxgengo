package nativeexport

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
)

const officeRelationships = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const strictOfficeRelationships = "http://purl.oclc.org/ooxml/officeDocument/relationships"

// presentationSlides follows the actual presentation order. Package parts that
// are not in sldIdLst are not exported pages, even if named slide123.xml.
func presentationSlides(archive *zip.Reader) ([]string, error) {
	files := map[string]*zip.File{}
	for _, file := range archive.File {
		if _, duplicate := files[file.Name]; duplicate {
			return nil, fmt.Errorf("duplicate PPTX part: %s", file.Name)
		}
		files[file.Name] = file
	}
	decode := func(name string, target any) error {
		file := files[name]
		if file == nil {
			return fmt.Errorf("missing PPTX part: %s", name)
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		if err := xml.Unmarshal(data, target); err != nil {
			return fmt.Errorf("invalid PPTX XML %s: %w", name, err)
		}
		return nil
	}
	var presentation struct {
		XMLName xml.Name `xml:"presentation"`
		Slides  []struct {
			Attributes []xml.Attr `xml:",any,attr"`
		} `xml:"sldIdLst>sldId"`
	}
	if err := decode("ppt/presentation.xml", &presentation); err != nil {
		return nil, err
	}
	if len(presentation.Slides) == 0 {
		return nil, fmt.Errorf("PPTX contains no presentation slides")
	}
	var relationships struct {
		XMLName xml.Name `xml:"Relationships"`
		Items   []struct {
			ID         string `xml:"Id,attr"`
			Type       string `xml:"Type,attr"`
			Target     string `xml:"Target,attr"`
			TargetMode string `xml:"TargetMode,attr"`
		} `xml:"Relationship"`
	}
	if err := decode("ppt/_rels/presentation.xml.rels", &relationships); err != nil {
		return nil, err
	}
	indices := map[string]int{}
	for i, relation := range relationships.Items {
		if _, duplicate := indices[relation.ID]; duplicate {
			return nil, fmt.Errorf("duplicate presentation relationship: %s", relation.ID)
		}
		indices[relation.ID] = i
	}
	ordered := []string{}
	for i, slide := range presentation.Slides {
		id := ""
		for _, attribute := range slide.Attributes {
			if attribute.Name.Local == "id" && (attribute.Name.Space == officeRelationships || attribute.Name.Space == strictOfficeRelationships) {
				id = attribute.Value
			}
		}
		index, found := indices[id]
		if id == "" || !found {
			return nil, fmt.Errorf("presentation slide %d has a missing relationship: %s", i+1, id)
		}
		relation := relationships.Items[index]
		if relation.Type != officeRelationships+"/slide" && relation.Type != strictOfficeRelationships+"/slide" || relation.TargetMode != "" && relation.TargetMode != "Internal" {
			return nil, fmt.Errorf("presentation slide %d relationship is not an internal slide", i+1)
		}
		target, err := url.Parse(relation.Target)
		if err != nil || target.Scheme != "" || target.Host != "" || target.RawQuery != "" || target.Fragment != "" || target.Path == "" {
			return nil, fmt.Errorf("invalid presentation slide target: %s", relation.Target)
		}
		name := path.Clean(path.Join("ppt", target.Path))
		if strings.HasPrefix(target.Path, "/") {
			name = path.Clean(strings.TrimPrefix(target.Path, "/"))
		}
		if files[name] == nil {
			// OPC part names can retain URI-escaped characters in the ZIP entry.
			escaped := path.Clean(path.Join("ppt", target.EscapedPath()))
			if strings.HasPrefix(target.Path, "/") {
				escaped = path.Clean(strings.TrimPrefix(target.EscapedPath(), "/"))
			}
			if files[escaped] == nil {
				return nil, fmt.Errorf("missing presentation slide part: %s", name)
			}
			name = escaped
		}
		ordered = append(ordered, name)
	}
	return ordered, nil
}
