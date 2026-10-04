package nativepkg

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const SourceInventorySchema = "pptxgengo.source-inventory.v1"

const maxInventoryPPTXBytes = 512 << 20
const maxInventoryPartBytes = 64 << 20

type InventoryBounds struct {
	X  int64 `json:"x_emu"`
	Y  int64 `json:"y_emu"`
	CX int64 `json:"cx_emu"`
	CY int64 `json:"cy_emu"`
}

type InventoryText struct {
	ObjectName  string           `json:"object_name,omitempty"`
	Placeholder string           `json:"placeholder,omitempty"`
	Bounds      *InventoryBounds `json:"bounds,omitempty"`
	Text        string           `json:"text"`
}

type InventoryTable struct {
	ObjectName string           `json:"object_name,omitempty"`
	Bounds     *InventoryBounds `json:"bounds,omitempty"`
	Rows       [][]string       `json:"rows"`
}

type InventoryChart struct {
	ObjectName string           `json:"object_name,omitempty"`
	Bounds     *InventoryBounds `json:"bounds,omitempty"`
	Part       string           `json:"part,omitempty"`
	Title      string           `json:"title,omitempty"`
	Types      []string         `json:"types"`
	Series     int              `json:"series_count"`
}

type InventoryImage struct {
	ObjectName  string           `json:"object_name,omitempty"`
	Description string           `json:"description,omitempty"`
	Title       string           `json:"title,omitempty"`
	Bounds      *InventoryBounds `json:"bounds,omitempty"`
	Target      string           `json:"target,omitempty"`
	MIME        string           `json:"mime,omitempty"`
	Bytes       int              `json:"bytes,omitempty"`
	External    bool             `json:"external,omitempty"`
}

type SourceSlideInventory struct {
	OriginalSlideIndex int              `json:"original_slide_index"`
	Part               string           `json:"part"`
	Title              string           `json:"title,omitempty"`
	Hidden             bool             `json:"hidden"`
	ShapeCount         int              `json:"shape_count"`
	TextShapeCount     int              `json:"text_shape_count"`
	PictureCount       int              `json:"picture_count"`
	TableCount         int              `json:"table_count"`
	ChartCount         int              `json:"chart_count"`
	OtherVisualCount   int              `json:"other_visual_count"`
	Texts              []InventoryText  `json:"texts"`
	Tables             []InventoryTable `json:"tables"`
	Charts             []InventoryChart `json:"charts"`
	Images             []InventoryImage `json:"images"`
	SpeakerNotes       []string         `json:"speaker_notes,omitempty"`
}

type SourceInventory struct {
	Schema       string                 `json:"schema"`
	Source       string                 `json:"source"`
	SourceSHA256 string                 `json:"source_sha256"`
	SlideCount   int                    `json:"slide_count"`
	Slides       []SourceSlideInventory `json:"slides"`
	Policy       []string               `json:"policy"`
}

type inventoryRelationship struct {
	Target string
	Type   string
	Mode   string
}

// Inventory extracts compact text and visual-object topology from an existing
// PPTX. It does not modify or retain raw source XML in its result.
func Inventory(source string) (SourceInventory, error) {
	out := SourceInventory{Schema: SourceInventorySchema, Source: filepath.Base(source), Slides: []SourceSlideInventory{}, Policy: []string{"Source slide indices follow presentation order, including hidden slides.", "Text is extracted from OOXML runs; appearance and semantic role are not inferred.", "Image paths identify package parts, not a recommendation to extract or reuse copyrighted content.", "This report is a source inventory, not editable PPTX reconciliation or a project import."}}
	info, err := os.Stat(source)
	if err != nil {
		return out, err
	}
	if info.Size() > maxInventoryPPTXBytes {
		return out, fmt.Errorf("source PPTX exceeds the %d MiB inventory limit", maxInventoryPPTXBytes>>20)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return out, err
	}
	hash := sha256.Sum256(data)
	out.SourceSHA256 = hex.EncodeToString(hash[:])
	parts, err := readInventoryParts(data)
	if err != nil {
		return out, err
	}
	presentation, err := Parse(parts["ppt/presentation.xml"])
	if err != nil {
		return out, fmt.Errorf("invalid source presentation: %w", err)
	}
	relations, err := parseInventoryRelationships(parts["ppt/_rels/presentation.xml.rels"])
	if err != nil {
		return out, fmt.Errorf("invalid source presentation relationships: %w", err)
	}
	list := presentation.Child("sldIdLst")
	if list == nil {
		return out, fmt.Errorf("source PPTX has no slide list")
	}
	for _, item := range list.Children {
		if local(item.Name) != "sldId" {
			continue
		}
		relation := relations[item.Attr("r:id")]
		if relation.Type == "" || !strings.HasSuffix(relation.Type, "/slide") || relation.Mode == "External" {
			return out, fmt.Errorf("source slide has missing or unsupported relationship %q", item.Attr("r:id"))
		}
		part := resolve("ppt/presentation.xml", relation.Target)
		slideXML, exists := parts[part]
		if !exists {
			return out, fmt.Errorf("source slide part missing: %s", part)
		}
		slide, parseErr := Parse(slideXML)
		if parseErr != nil {
			return out, fmt.Errorf("invalid source slide %s: %w", part, parseErr)
		}
		slideRelations, relErr := parseInventoryRelationships(parts[relpart(part)])
		if relErr != nil {
			return out, fmt.Errorf("invalid source slide relationships %s: %w", part, relErr)
		}
		result := inventorySlide(len(out.Slides)+1, part, slide, slideRelations, parts)
		result.SpeakerNotes = inventoryNotes(slideRelations, parts)
		out.Slides = append(out.Slides, result)
	}
	if len(out.Slides) == 0 {
		return out, fmt.Errorf("source PPTX contains no slides")
	}
	out.SlideCount = len(out.Slides)
	return out, nil
}

func readInventoryParts(data []byte) (map[string][]byte, error) {
	if int64(len(data)) > maxInventoryPPTXBytes {
		return nil, fmt.Errorf("source PPTX exceeds the %d MiB inventory limit", maxInventoryPPTXBytes>>20)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	parts := make(map[string][]byte, len(archive.File))
	var total int64
	for _, file := range archive.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if _, duplicate := parts[file.Name]; duplicate {
			return nil, fmt.Errorf("duplicate source PPTX part: %s", file.Name)
		}
		if file.UncompressedSize64 > maxInventoryPartBytes {
			return nil, fmt.Errorf("PPTX part %s exceeds the %d MiB inventory limit", file.Name, maxInventoryPartBytes>>20)
		}
		if total+int64(file.UncompressedSize64) > maxInventoryPPTXBytes {
			return nil, fmt.Errorf("expanded PPTX exceeds the %d MiB inventory limit", maxInventoryPPTXBytes>>20)
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		content, readErr := io.ReadAll(io.LimitReader(reader, maxInventoryPartBytes+1))
		closeErr := reader.Close()
		if readErr != nil {
			return nil, readErr
		}
		if int64(len(content)) > maxInventoryPartBytes {
			return nil, fmt.Errorf("PPTX part %s expands beyond the %d MiB inventory limit", file.Name, maxInventoryPartBytes>>20)
		}
		if total+int64(len(content)) > maxInventoryPPTXBytes {
			return nil, fmt.Errorf("expanded PPTX exceeds the %d MiB inventory limit", maxInventoryPPTXBytes>>20)
		}
		total += int64(len(content))
		if closeErr != nil {
			return nil, closeErr
		}
		parts[file.Name] = content
	}
	return parts, nil
}

func parseInventoryRelationships(raw []byte) (map[string]inventoryRelationship, error) {
	if len(raw) == 0 {
		return map[string]inventoryRelationship{}, nil
	}
	root, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	result := map[string]inventoryRelationship{}
	for _, child := range root.Children {
		if local(child.Name) != "Relationship" {
			continue
		}
		id := child.Attr("Id")
		if id == "" {
			continue
		}
		if _, exists := result[id]; exists {
			return nil, fmt.Errorf("duplicate relationship id: %s", id)
		}
		result[id] = inventoryRelationship{Target: child.Attr("Target"), Type: child.Attr("Type"), Mode: child.Attr("TargetMode")}
	}
	return result, nil
}

func inventorySlide(number int, part string, slide *Node, rels map[string]inventoryRelationship, parts map[string][]byte) SourceSlideInventory {
	result := SourceSlideInventory{OriginalSlideIndex: number, Part: part, Hidden: slide.Attr("show") == "0" || strings.EqualFold(slide.Attr("show"), "false"), Texts: []InventoryText{}, Tables: []InventoryTable{}, Charts: []InventoryChart{}, Images: []InventoryImage{}}
	if common := slide.Child("cSld"); common != nil {
		if tree := common.Child("spTree"); tree != nil {
			for _, child := range tree.Children {
				switch local(child.Name) {
				case "sp":
					result.ShapeCount++
					text := inventoryText(child)
					if text.Text != "" {
						result.TextShapeCount++
						result.Texts = append(result.Texts, text)
						if result.Title == "" && (text.Placeholder == "title" || text.Placeholder == "ctrTitle") {
							result.Title = text.Text
						}
					}
				case "cxnSp":
					result.ShapeCount++
					if text := inventoryText(child); text.Text != "" {
						result.TextShapeCount++
						result.Texts = append(result.Texts, text)
					}
				case "pic":
					result.PictureCount++
					result.Images = append(result.Images, inventoryImage(child, rels, parts))
				case "graphicFrame":
					result.ShapeCount++
					if table := findNamed(child, "tbl"); table != nil {
						result.Tables = append(result.Tables, inventoryTable(child, table))
					} else if chartID := inventoryChartRelationship(child); chartID != "" {
						result.Charts = append(result.Charts, inventoryChart(child, chartID, rels, parts))
					} else {
						result.OtherVisualCount++
					}
				case "grpSp":
					result.ShapeCount++
					result.OtherVisualCount++
				}
			}
		}
	}
	result.TableCount, result.ChartCount = len(result.Tables), len(result.Charts)
	if result.Title == "" && len(result.Texts) > 0 {
		result.Title = result.Texts[0].Text
	}
	return result
}

func inventoryText(object *Node) InventoryText {
	result := InventoryText{ObjectName: visualObjectName(object), Bounds: inventoryBounds(object)}
	if placeholder := findNamed(object, "ph"); placeholder != nil {
		result.Placeholder = placeholder.Attr("type")
	}
	body := findNamed(object, "txBody")
	if body != nil {
		paragraphs := []*Node{}
		body.Walk(func(n *Node) {
			if local(n.Name) == "p" {
				paragraphs = append(paragraphs, n)
			}
		})
		var lines []string
		for _, paragraph := range paragraphs {
			var builder strings.Builder
			paragraph.Walk(func(n *Node) {
				if local(n.Name) == "t" {
					builder.WriteString(n.TextContent())
				}
			})
			line := strings.TrimSpace(builder.String())
			if line != "" {
				lines = append(lines, line)
			}
		}
		result.Text = strings.Join(lines, "\n")
	}
	return result
}

func inventoryTable(frame, table *Node) InventoryTable {
	result := InventoryTable{ObjectName: visualObjectName(frame), Bounds: inventoryBounds(frame), Rows: [][]string{}}
	for _, row := range table.Children {
		if local(row.Name) != "tr" {
			continue
		}
		cells := []string{}
		for _, cell := range row.Children {
			if local(cell.Name) == "tc" {
				cells = append(cells, inventoryText(cell).Text)
			}
		}
		result.Rows = append(result.Rows, cells)
	}
	return result
}

func inventoryChartRelationship(frame *Node) string {
	graphicData := findNamed(frame, "graphicData")
	if graphicData == nil || !strings.Contains(graphicData.Attr("uri"), "chart") {
		return ""
	}
	chart := findNamed(graphicData, "chart")
	if chart == nil {
		return ""
	}
	for _, attr := range chart.Attrs {
		if strings.HasSuffix(attr.Name, ":id") {
			return attr.Value
		}
	}
	return ""
}

func inventoryChart(frame *Node, id string, rels map[string]inventoryRelationship, parts map[string][]byte) InventoryChart {
	result := InventoryChart{ObjectName: visualObjectName(frame), Bounds: inventoryBounds(frame), Types: []string{}}
	rel := rels[id]
	if rel.Target == "" || rel.Mode == "External" {
		return result
	}
	result.Part = resolve("ppt/slides/slide.xml", rel.Target)
	raw := parts[result.Part]
	chart, err := Parse(raw)
	if err != nil {
		return result
	}
	if title := findNamed(chart, "title"); title != nil {
		result.Title = strings.TrimSpace(allText(title))
	}
	seenTypes := map[string]bool{}
	chart.Walk(func(node *Node) {
		name := local(node.Name)
		if strings.HasSuffix(name, "Chart") && name != "chart" && !seenTypes[name] {
			seenTypes[name] = true
			result.Types = append(result.Types, strings.TrimSuffix(name, "Chart"))
		}
		if name == "ser" {
			result.Series++
		}
	})
	sort.Strings(result.Types)
	return result
}

func inventoryImage(pic *Node, rels map[string]inventoryRelationship, parts map[string][]byte) InventoryImage {
	result := InventoryImage{ObjectName: visualObjectName(pic), Bounds: inventoryBounds(pic)}
	if props := findNamed(pic, "cNvPr"); props != nil {
		result.Description = props.Attr("descr")
		result.Title = props.Attr("title")
	}
	var embed, link string
	if blip := findNamed(pic, "blip"); blip != nil {
		for _, attr := range blip.Attrs {
			if strings.HasSuffix(attr.Name, ":embed") {
				embed = attr.Value
			}
			if strings.HasSuffix(attr.Name, ":link") {
				link = attr.Value
			}
		}
	}
	relationID := embed
	if relationID == "" {
		relationID = link
	}
	rel := rels[relationID]
	if rel.Mode == "External" || link != "" && embed == "" {
		result.Target, result.External = rel.Target, true
		return result
	}
	if rel.Target != "" {
		result.Target = resolve("ppt/slides/slide.xml", rel.Target)
	}
	if raw, ok := parts[result.Target]; ok {
		result.Bytes = len(raw)
		result.MIME = mime.TypeByExtension(filepath.Ext(result.Target))
	}
	return result
}

func inventoryNotes(rels map[string]inventoryRelationship, parts map[string][]byte) []string {
	var target string
	for _, rel := range rels {
		if strings.HasSuffix(rel.Type, "/notesSlide") && rel.Mode != "External" {
			target = resolve("ppt/slides/slide.xml", rel.Target)
			break
		}
	}
	if target == "" {
		return nil
	}
	note, err := Parse(parts[target])
	if err != nil {
		return nil
	}
	var result []string
	if tree := findNamed(note, "spTree"); tree != nil {
		for _, shape := range tree.Children {
			if local(shape.Name) != "sp" {
				continue
			}
			placeholder := findNamed(shape, "ph")
			kind := ""
			if placeholder != nil {
				kind = placeholder.Attr("type")
			}
			if kind != "" && kind != "body" {
				continue
			}
			text := strings.TrimSpace(inventoryText(shape).Text)
			if text != "" {
				result = append(result, text)
			}
		}
	}
	return result
}

func inventoryBounds(object *Node) *InventoryBounds {
	xfrm := findNamed(object, "xfrm")
	if xfrm == nil {
		return nil
	}
	off, ext := findNamed(xfrm, "off"), findNamed(xfrm, "ext")
	if off == nil || ext == nil {
		return nil
	}
	parse := func(value string) int64 {
		n, _ := strconv.ParseInt(value, 10, 64)
		return n
	}
	return &InventoryBounds{X: parse(off.Attr("x")), Y: parse(off.Attr("y")), CX: parse(ext.Attr("cx")), CY: parse(ext.Attr("cy"))}
}

func visualObjectName(object *Node) string {
	if props := findNamed(object, "cNvPr"); props != nil {
		return props.Attr("name")
	}
	return ""
}

func findNamed(node *Node, name string) *Node {
	if node == nil {
		return nil
	}
	if local(node.Name) == name {
		return node
	}
	for _, child := range node.Children {
		if found := findNamed(child, name); found != nil {
			return found
		}
	}
	return nil
}

func allText(node *Node) string {
	var parts []string
	if node == nil {
		return ""
	}
	node.Walk(func(child *Node) {
		if local(child.Name) == "t" {
			text := strings.TrimSpace(child.TextContent())
			if text != "" {
				parts = append(parts, text)
			}
		}
	})
	return strings.Join(parts, " ")
}
