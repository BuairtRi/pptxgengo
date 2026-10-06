package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func draftReviewNativeMetadata(raw []byte, slides []SlideReport) ([]byte, error) {
	groups := make([]SlideReport, len(slides))
	notes := map[string]*DraftReviewRecord{}
	for i, slide := range slides {
		if slide.DraftReview == nil {
			continue
		}
		record := slide.DraftReview
		groups[i].Scenes = []SceneRecord{{ID: record.Group.ID, Parts: record.Group.Parts, Groups: []ComponentRecord{record.Group}}}
		notes[fmt.Sprintf("ppt/slides/slide%d.xml", i+1)] = record
	}
	if len(notes) == 0 {
		return raw, nil
	}
	grouped, err := sceneNativeGroups(raw, groups)
	if err != nil {
		return nil, err
	}
	return rewriteDraftReviewSlides(grouped, func(path string, data []byte) ([]byte, error) {
		note := notes[path]
		if note == nil {
			return data, nil
		}
		payload, err := json.Marshal(note.Note)
		if err != nil {
			return nil, err
		}
		var content, group bytes.Buffer
		xml.EscapeText(&content, payload)
		xml.EscapeText(&group, []byte(note.Group.ID))
		ext := []byte(fmt.Sprintf(`<p:ext uri="%s"><wm:review xmlns:wm="%s" group="%s">%s</wm:review></p:ext>`, DraftReviewContract, DraftReviewContract, group.String(), content.String()))
		start, end, err := slideExtensionList(data)
		if err != nil {
			return nil, err
		}
		if start >= 0 {
			close := bytes.LastIndexByte(data[start:end], '<') + start
			return joinDraftXML(data[:close], ext, data[close:]), nil
		}
		close := bytes.LastIndex(data, []byte("</p:sld>"))
		if close < 0 {
			return nil, fmt.Errorf("draft_review.invalid_slide_xml")
		}
		return joinDraftXML(data[:close], []byte("<p:extLst>"), ext, []byte("</p:extLst>"), data[close:]), nil
	})
}

func joinDraftXML(parts ...[]byte) []byte { return bytes.Join(parts, nil) }

func slideExtensionList(data []byte) (start, end int, err error) {
	start, end = -1, -1
	d := xml.NewDecoder(bytes.NewReader(data))
	depth := 0
	for {
		before := int(d.InputOffset())
		token, e := d.Token()
		if e == io.EOF {
			return start, end, nil
		}
		if e != nil {
			return -1, -1, e
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if depth == 2 && t.Name.Local == "extLst" {
				if start >= 0 {
					return -1, -1, fmt.Errorf("draft_review.multiple_slide_extension_lists")
				}
				start = before + bytes.LastIndexByte(data[before:int(d.InputOffset())], '<')
			}
		case xml.EndElement:
			if depth == 2 && t.Name.Local == "extLst" {
				end = int(d.InputOffset())
			}
			depth--
		}
	}
}

// RemoveDraftReviewNotes cleans the actual PPTX bytes for client delivery. A
// pasteboard object's absence from exported pixels does not make its copy private.
func RemoveDraftReviewNotes(raw []byte) ([]byte, error) {
	return rewriteDraftReviewSlides(raw, func(_ string, data []byte) ([]byte, error) {
		if !bytes.Contains(data, []byte(DraftReviewContract)) && !bytes.Contains(data, []byte("wm-review/")) {
			return data, nil
		}
		inventory, err := sceneNativeInventory(data)
		if err != nil {
			return nil, err
		}
		type span struct{ start, end int }
		var cuts []span
		for _, object := range inventory.objects {
			if object.kind == "grpSp" && strings.HasPrefix(object.name, "wm-review/") {
				cuts = append(cuts, span{object.start, object.end})
			}
		}
		// Remove only our extension, retaining other slide metadata verbatim.
		d := xml.NewDecoder(bytes.NewReader(data))
		depth, noteDepth, begin := 0, -1, -1
		for {
			before := int(d.InputOffset())
			token, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			switch t := token.(type) {
			case xml.StartElement:
				depth++
				if t.Name.Local == "ext" {
					for _, attr := range t.Attr {
						if attr.Name.Local == "uri" && attr.Value == DraftReviewContract {
							begin = before + bytes.LastIndexByte(data[before:int(d.InputOffset())], '<')
							noteDepth = depth
						}
					}
				}
			case xml.EndElement:
				if depth == noteDepth {
					cuts = append(cuts, span{begin, int(d.InputOffset())})
					noteDepth = -1
				}
				depth--
			}
		}
		// Native groups precede slide extensions; sort defensively for every case.
		for i := 1; i < len(cuts); i++ {
			for j := i; j > 0 && cuts[j].start < cuts[j-1].start; j-- {
				cuts[j], cuts[j-1] = cuts[j-1], cuts[j]
			}
		}
		var clean bytes.Buffer
		last := 0
		for _, cut := range cuts {
			if cut.start < last {
				return nil, fmt.Errorf("draft_review.overlapping_private_objects")
			}
			clean.Write(data[last:cut.start])
			last = cut.end
		}
		clean.Write(data[last:])
		result := clean.Bytes()
		// Our extension was the only child in newly created extension lists.
		result = bytes.ReplaceAll(result, []byte("<p:extLst></p:extLst>"), nil)
		remaining, err := sceneNativeInventory(result)
		if err != nil {
			return nil, err
		}
		for name := range remaining.names {
			if strings.HasPrefix(name, "wm-review/") {
				return nil, fmt.Errorf("draft_review.private_content_not_fully_removed: %s", name)
			}
		}
		return result, nil
	})
}

// Unchanged archive entries retain compressed bytes and headers; decks without
// draft objects are returned byte-for-byte, including their archive structure.
func rewriteDraftReviewSlides(raw []byte, patch func(string, []byte) ([]byte, error)) ([]byte, error) {
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	changed := map[string][]byte{}
	for _, file := range z.File {
		if !strings.HasPrefix(file.Name, "ppt/slides/slide") || !strings.HasSuffix(file.Name, ".xml") {
			continue
		}
		r, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			return nil, err
		}
		result, err := patch(file.Name, data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file.Name, err)
		}
		if !bytes.Equal(result, data) {
			changed[file.Name] = result
		}
	}
	if len(changed) == 0 {
		return raw, nil
	}
	var output bytes.Buffer
	w := zip.NewWriter(&output)
	for _, file := range z.File {
		if data, ok := changed[file.Name]; ok {
			entry, err := w.CreateHeader(&file.FileHeader)
			if err != nil {
				return nil, err
			}
			if _, err = entry.Write(data); err != nil {
				return nil, err
			}
		} else if err := w.Copy(file); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
