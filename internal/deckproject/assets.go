package deckproject

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// assetMIME validates project-owned payloads before their bytes reach the writer.
// SVG stays vector in PowerPoint and gets the existing strict raster fallback.
func assetMIME(data []byte) (string, error) {
	trimmed := bytes.TrimSpace(data)
	if bytes.HasPrefix(trimmed, []byte("<")) {
		d := xml.NewDecoder(bytes.NewReader(data))
		root := false
		depth := 0
		allowed := map[string]bool{"svg": true, "g": true, "defs": true, "title": true, "desc": true, "path": true, "rect": true, "circle": true}
		attributes := map[string]bool{"id": true, "class": true, "viewBox": true, "width": true, "height": true, "x": true, "y": true, "rx": true, "ry": true, "cx": true, "cy": true, "r": true, "d": true, "fill": true, "stroke": true, "style": true, "transform": true, "version": true}
		for {
			tok, e := d.Token()
			if e == io.EOF {
				if !root || depth != 0 {
					return "", fmt.Errorf("invalid SVG document")
				}
				return "image/svg+xml", nil
			}
			if e != nil {
				return "", fmt.Errorf("invalid SVG XML: %w", e)
			}
			switch t := tok.(type) {
			case xml.Directive:
				return "", fmt.Errorf("SVG DTD/directives are unsupported")
			case xml.ProcInst:
				if t.Target != "xml" {
					return "", fmt.Errorf("SVG processing instructions unsupported")
				}
			case xml.StartElement:
				if !root {
					if t.Name.Local != "svg" || (t.Name.Space != "" && t.Name.Space != "http://www.w3.org/2000/svg") {
						return "", fmt.Errorf("asset XML root must be SVG")
					}
					root = true
				} else if depth == 0 {
					return "", fmt.Errorf("multiple SVG roots unsupported")
				}
				if t.Name.Space != "" && t.Name.Space != "http://www.w3.org/2000/svg" {
					return "", fmt.Errorf("unsupported foreign SVG namespace")
				}
				depth++
				if depth > 100 || !allowed[t.Name.Local] {
					return "", fmt.Errorf("unsupported SVG element %s", t.Name.Local)
				}
				for _, a := range t.Attr {
					if a.Name.Local == "xmlns" || a.Name.Space == "xmlns" {
						continue
					}
					if a.Name.Space != "" {
						return "", fmt.Errorf("unsupported namespaced SVG attribute %s", a.Name.Local)
					}
					if !attributes[a.Name.Local] {
						return "", fmt.Errorf("unsupported SVG attribute %s", a.Name.Local)
					}
					if a.Name.Local == "id" && a.Value == "stroke" {
						return "", fmt.Errorf("reserved SVG stroke id unsupported for custom assets")
					}
					if strings.Contains("|x|y|width|height|rx|ry|cx|cy|r|", "|"+a.Name.Local+"|") {
						n, err := strconv.ParseFloat(a.Value, 64)
						if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
							return "", fmt.Errorf("SVG geometry %s requires finite unitless numbers", a.Name.Local)
						}
					}
					name := strings.ToLower(a.Name.Local)
					value := strings.ToLower(a.Value)
					if name == "stroke" && value != "none" && value != "" {
						return "", fmt.Errorf("unsupported SVG visible stroke")
					}
					if (name == "opacity" || name == "fill-opacity" || name == "stroke-opacity") && value != "1" {
						return "", fmt.Errorf("unsupported SVG opacity")
					}
					if name == "fill-rule" && value != "nonzero" {
						return "", fmt.Errorf("unsupported SVG fill rule")
					}
					if name == "clip-path" || name == "mask" || name == "filter" || name == "stroke-dasharray" {
						return "", fmt.Errorf("unsupported SVG visual attribute %s", name)
					}
					if name == "style" {
						if e := validateSVGStyle(a.Value); e != nil {
							return "", e
						}
					}
					if name == "href" || strings.HasPrefix(name, "on") || strings.Contains(value, "url(") || strings.Contains(value, "@import") || strings.Contains(value, "://") || strings.Contains(value, "data:") {
						return "", fmt.Errorf("SVG external resources/events unsupported")
					}
				}
			case xml.EndElement:
				depth--
			case xml.CharData:
				value := strings.Map(func(r rune) rune {
					if unicode.IsSpace(r) {
						return -1
					}
					return unicode.ToLower(r)
				}, string(t))
				if strings.Contains(value, "stroke:") || strings.Contains(value, "opacity:") || strings.Contains(value, "fill-rule:") || strings.Contains(value, "clip-path:") || strings.Contains(value, "mask:") || strings.Contains(value, "filter:") {
					return "", fmt.Errorf("unsupported SVG visual stylesheet")
				}
				if strings.Contains(value, "url(") || strings.Contains(value, "@import") {
					return "", fmt.Errorf("SVG external style resources unsupported")
				}
			}
		}
	}
	config, format, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil {
		return "", fmt.Errorf("project asset requires PNG, JPEG or self-contained SVG: %w", e)
	}
	if config.Width <= 0 || config.Height <= 0 || uint64(config.Width)*uint64(config.Height) > 100_000_000 {
		return "", fmt.Errorf("project raster asset exceeds 100 megapixels")
	}
	switch format {
	case "png":
		return "image/png", nil
	case "jpeg":
		return "image/jpeg", nil
	default:
		return "", fmt.Errorf("unsupported project raster format %s", format)
	}
}

type DerivationReceipt struct {
	Schema       string         `json:"schema"`
	SourceAsset  string         `json:"source_asset"`
	SourceSHA256 string         `json:"source_sha256"`
	ResultSHA256 string         `json:"result_sha256"`
	Operation    string         `json:"operation"`
	Parameters   map[string]any `json:"parameters,omitempty"`
}

func validateSVGStyle(style string) error {
	for _, part := range strings.Split(style, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pieces := strings.SplitN(part, ":", 2)
		if len(pieces) != 2 {
			return fmt.Errorf("invalid SVG inline style")
		}
		key := strings.TrimSpace(strings.ToLower(pieces[0]))
		value := strings.TrimSpace(strings.ToLower(pieces[1]))
		switch key {
		case "fill":
		case "stroke":
			if value != "none" {
				return fmt.Errorf("unsupported SVG visible stroke")
			}
		default:
			return fmt.Errorf("unsupported SVG inline style %s", key)
		}
	}
	return nil
}
