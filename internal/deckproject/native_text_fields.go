package deckproject

import (
	"sort"
	"strings"
)

const NativeTextModelSchema = "pptxgengo.native-text-model.v1"
const drawingML = "http://schemas.openxmlformats.org/drawingml/2006/main"

type NativeTextAddress struct {
	Body        int  `json:"body"`
	Paragraph   int  `json:"paragraph"`
	TableRow    *int `json:"table_row,omitempty"`
	TableColumn *int `json:"table_column,omitempty"`
}

type NativeTextRun struct {
	Ordinal          int    `json:"ordinal"`
	Kind             string `json:"kind"`
	Text             string `json:"text"`
	PropertiesSHA256 string `json:"properties_sha256"`
}

type NativeParagraph struct {
	Address             NativeTextAddress `json:"address"`
	Text                string            `json:"text"`
	Runs                []NativeTextRun   `json:"runs"`
	PropertiesSHA256    string            `json:"properties_sha256"`
	EndPropertiesSHA256 string            `json:"end_properties_sha256"`
	ReviewItems         []string          `json:"review_items,omitempty"`
}

type NativeSourceField struct {
	Identity             string              `json:"identity"`
	SourceSlot           string              `json:"source_slot,omitempty"`
	SourcePointer        string              `json:"source_pointer"`
	SourceValueSHA256    string              `json:"source_value_sha256"`
	BaselineNativeSHA256 string              `json:"baseline_native_sha256"`
	Addresses            []NativeTextAddress `json:"addresses"`
	Status               string              `json:"status"`
	Reason               string              `json:"reason,omitempty"`
}

func directXML(n *xmlNode, space, local string) *xmlNode {
	for _, c := range n.Children {
		if c.Name.Local == local && c.Name.Space == space {
			return c
		}
	}
	return nil
}

// Concatenate runs within paragraphs; preserve explicit breaks and paragraph
// boundaries. Tables retain distinct cell addresses instead of one guessed field.
func nativeParagraphs(shape *xmlNode) []NativeParagraph {
	result := []NativeParagraph{}
	bodyOrdinal := 0
	add := func(body *xmlNode, row, column *int) {
		paragraph := 0
		for _, p := range body.Children {
			if p.Name.Space != drawingML || p.Name.Local != "p" {
				continue
			}
			out := NativeParagraph{Address: NativeTextAddress{Body: bodyOrdinal, Paragraph: paragraph, TableRow: row, TableColumn: column}, Runs: []NativeTextRun{}, PropertiesSHA256: digest(canonical(directXML(p, drawingML, "pPr"))), EndPropertiesSHA256: digest(canonical(directXML(p, drawingML, "endParaRPr")))}
			for _, run := range p.Children {
				if run.Name.Space != drawingML {
					out.ReviewItems = append(out.ReviewItems, "foreign_paragraph_content")
					continue
				}
				switch run.Name.Local {
				case "pPr", "endParaRPr":
				case "r", "fld", "br":
					text := ""
					if run.Name.Local == "br" {
						text = "\n"
					} else {
						for _, child := range run.Children {
							if child.Name.Space == drawingML && child.Name.Local == "t" {
								text += child.Text
							}
						}
					}
					out.Runs = append(out.Runs, NativeTextRun{Ordinal: len(out.Runs), Kind: run.Name.Local, Text: text, PropertiesSHA256: digest(canonical(directXML(run, drawingML, "rPr")))})
					out.Text += text
					if run.Name.Local == "fld" {
						out.ReviewItems = append(out.ReviewItems, "dynamic_field")
					}
				default:
					out.ReviewItems = append(out.ReviewItems, "unsupported_paragraph_child:"+run.Name.Local)
				}
			}
			if props := directXML(p, drawingML, "pPr"); props != nil {
				for _, child := range props.Children {
					if child.Name.Space == drawingML && (child.Name.Local == "buChar" || child.Name.Local == "buAutoNum" || child.Name.Local == "buBlip") {
						out.ReviewItems = append(out.ReviewItems, "bullet_or_numbered_paragraph")
					}
				}
			}
			result = append(result, out)
			paragraph++
		}
		bodyOrdinal++
	}
	if shape.Name.Local == "sp" {
		// p:txBody is a presentation container; its paragraphs are DrawingML.
		for _, child := range shape.Children {
			if child.Name.Space == "http://schemas.openxmlformats.org/presentationml/2006/main" && child.Name.Local == "txBody" {
				add(child, nil, nil)
			}
		}
	} else if shape.Name.Local == "graphicFrame" {
		var walk func(*xmlNode)
		walk = func(n *xmlNode) {
			if n.Name.Space == drawingML && n.Name.Local == "tbl" {
				row := 0
				for _, tr := range n.Children {
					if tr.Name.Space != drawingML || tr.Name.Local != "tr" {
						continue
					}
					column := 0
					for _, tc := range tr.Children {
						if tc.Name.Space != drawingML || tc.Name.Local != "tc" {
							continue
						}
						r, c := row, column
						if body := directXML(tc, drawingML, "txBody"); body != nil {
							add(body, &r, &c)
						}
						column++
					}
					row++
				}
				return
			}
			for _, child := range n.Children {
				walk(child)
			}
		}
		walk(shape)
	}
	return result
}

func nativeParagraphText(paragraphs []NativeParagraph) string {
	values := []string{}
	for _, p := range paragraphs {
		values = append(values, p.Text)
	}
	return strings.Join(values, "\n")
}

func attachNativeSourceFields(p *Project, r ObjectRecord) ObjectRecord {
	if len(r.Paragraphs) == 0 {
		return r
	}
	candidates := []string{}
	for _, pointer := range r.SourcePointers {
		value, err := lookupPointer(p.tree, pointer)
		if _, ok := value.(string); err == nil && ok {
			candidates = append(candidates, pointer)
		}
	}
	sort.Strings(candidates)
	unique := []string{}
	for _, pointer := range candidates {
		if len(unique) == 0 || unique[len(unique)-1] != pointer {
			unique = append(unique, pointer)
		}
	}
	if len(unique) != 1 {
		r.TextMapping = "manual_review"
		r.TextMappingReason = "Native text has zero or multiple explicit string source fields."
		return r
	}
	pointer := unique[0]
	value, _ := lookupPointer(p.tree, pointer)
	text := value.(string)
	field := NativeSourceField{Identity: r.LogicalID + "/text", SourceSlot: r.SourceSlots[pointer], SourcePointer: pointer, SourceValueSHA256: digest([]byte(text)), BaselineNativeSHA256: digest([]byte(r.NativeText)), Addresses: []NativeTextAddress{}, Status: "plain_text_baseline"}
	if field.SourceSlot == "" {
		field.Status = "manual_review"
		field.Reason = "Stable source field identity is unavailable."
	}
	if r.NativeKind != "sp" {
		field.Status = "manual_review"
		field.Reason = "Table/chart text requires cell/data source mapping."
	}
	for _, paragraph := range r.Paragraphs {
		field.Addresses = append(field.Addresses, paragraph.Address)
		if len(paragraph.ReviewItems) != 0 {
			field.Status = "manual_review"
			field.Reason = "Bullets, dynamic fields or unsupported paragraph content require explicit mapping."
		}
		if len(paragraph.Runs) > 1 {
			field.Status = "manual_review"
			field.Reason = "Multiple runs require a reviewed rich-text field contract."
		}
	}
	if text != r.NativeText {
		field.Status = "manual_review"
		field.Reason = "Native text differs from the exact source string; no whitespace or layout-break inference is permitted."
	}
	r.Fields = []NativeSourceField{field}
	r.TextMapping = field.Status
	r.TextMappingReason = field.Reason
	return r
}

func typedCardAssignmentMatches(name string, assignmentTarget, property string) bool {
	if strings.HasPrefix(property, "items.") {
		container := strings.LastIndex(property, ".card.")
		if container < 0 {
			return false
		}
		target := assignmentTarget + "." + property[:container] + property[container+len(".card"):]
		return name == target
	}
	return name == assignmentTarget || strings.HasPrefix(name, assignmentTarget+".")
}

// Remove only DrawingML text leaves. Changes to run structure, formatting,
// geometry, names, group ownership or other native payload remain fingerprinted.
func nativeStructureHash(n *xmlNode) string {
	var copyNode func(*xmlNode) *xmlNode
	copyNode = func(node *xmlNode) *xmlNode {
		clone := *node
		clone.Children = nil
		if clone.Name.Space == drawingML && clone.Name.Local == "t" {
			clone.Text = ""
		}
		for _, child := range node.Children {
			clone.Children = append(clone.Children, copyNode(child))
		}
		return &clone
	}
	return digest(canonical(copyNode(n)))
}
