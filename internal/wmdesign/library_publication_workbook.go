package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

// Embedded workbooks are ZIP containers. Compression, entry order and ZIP
// timestamps do not affect their contents; every inner part is compared.
func publicationWorkbookParts(raw []byte) (map[string][]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	parts := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		if _, exists := parts[file.Name]; exists {
			return nil, fmt.Errorf("workbook_duplicate_part: %s", file.Name)
		}
		stream, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(stream)
		closeErr := stream.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		parts[file.Name] = data
	}
	return parts, nil
}

// Preserve every byte except valid, direct-child created/modified timestamp
// text in the namespace-qualified OPC core properties part. XML attributes,
// other metadata, whitespace, namespaces and element topology stay strict.
func publicationWorkbookCoreWithoutTimestamps(raw []byte) ([]byte, error) {
	const coreNS = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
	const dateNS = "http://purl.org/dc/terms/"
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	var stack []xml.Name
	seen := map[string]bool{}
	var out bytes.Buffer
	var last int64
	for {
		start := decoder.InputOffset()
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case xml.StartElement:
			stack = append(stack, token.Name)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) != 2 || stack[0] != (xml.Name{Space: coreNS, Local: "coreProperties"}) || stack[1].Space != dateNS || (stack[1].Local != "created" && stack[1].Local != "modified") {
				continue
			}
			name := stack[1].Local
			if seen[name] {
				return nil, fmt.Errorf("workbook_duplicate_core_timestamp: %s", name)
			}
			seen[name] = true
			if _, err := time.Parse(time.RFC3339Nano, string(token)); err != nil {
				return nil, fmt.Errorf("workbook_invalid_core_timestamp: %s", name)
			}
			end := decoder.InputOffset()
			if !bytes.Equal(raw[start:end], token) {
				return nil, fmt.Errorf("workbook_encoded_core_timestamp: %s", name)
			}
			out.Write(raw[last:start])
			last = end
		}
	}
	out.Write(raw[last:])
	return out.Bytes(), nil
}

func comparePublicationWorkbooks(a, b []byte) error {
	left, err := publicationWorkbookParts(a)
	if err != nil {
		return err
	}
	right, err := publicationWorkbookParts(b)
	if err != nil {
		return err
	}
	if len(left) != len(right) {
		return fmt.Errorf("workbook_part_count_changed")
	}
	for name, l := range left {
		r, exists := right[name]
		if !exists {
			return fmt.Errorf("workbook_part_missing: %s", name)
		}
		if name == "docProps/core.xml" {
			l, err = publicationWorkbookCoreWithoutTimestamps(l)
			if err != nil {
				return err
			}
			r, err = publicationWorkbookCoreWithoutTimestamps(r)
			if err != nil {
				return err
			}
		}
		if !bytes.Equal(l, r) {
			return fmt.Errorf("workbook_part_changed: %s", name)
		}
	}
	return nil
}
