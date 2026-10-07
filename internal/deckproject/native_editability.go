package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

const NativeEditabilitySchema = "pptxgengo.native-editability.v1"

type NativeEditabilityUnit struct {
	ShapeToken     string                       `json:"shape_token"`
	SlideID        string                       `json:"slide_id"`
	LogicalID      string                       `json:"logical_id"`
	SelectionName  string                       `json:"selection_name"`
	NativeKind     string                       `json:"native_kind"`
	Family         string                       `json:"family"`
	Definition     string                       `json:"definition,omitempty"`
	ParentToken    string                       `json:"parent_token,omitempty"`
	TopLevelToken  string                       `json:"top_level_token"`
	GroupDepth     int                          `json:"group_depth"`
	ChildTokens    []string                     `json:"child_tokens"`
	ParagraphCount int                          `json:"paragraph_count"`
	TableCells     int                          `json:"table_cells"`
	TextAndPaint   bool                         `json:"text_and_paint_in_same_shape"`
	Fields         []NativeSourceField          `json:"source_fields"`
	Caveats        []string                     `json:"caveats"`
	Connection     *NativeEditabilityConnection `json:"connection,omitempty"`
}
type NativeEditabilityEndpoint struct {
	ShapeToken string `json:"shape_token,omitempty"`
	Site       int    `json:"site"`
	Status     string `json:"status"`
}
type NativeEditabilityConnection struct {
	Begin NativeEditabilityEndpoint `json:"begin"`
	End   NativeEditabilityEndpoint `json:"end"`
}
type NativeEditabilityReport struct {
	Schema               string                  `json:"schema"`
	ProjectID            string                  `json:"project_id"`
	BuildID              string                  `json:"build_id"`
	ReceiptSHA256        string                  `json:"receipt_sha256"`
	PPTXSHA256           string                  `json:"pptx_sha256"`
	GenerationToken      string                  `json:"generation_token"`
	Objects              []NativeEditabilityUnit `json:"objects"`
	Counts               map[string]int          `json:"counts"`
	DesktopQualification string                  `json:"desktop_qualification"`
	Scope                string                  `json:"scope"`
}

// NativeEditability inventories the actual receipt-pinned generated package.
// Object names describe Selection Pane entries; generation tags remain identity.
// Counts and groups are structural evidence, never desktop editing acceptance.
func NativeEditability(b *TextBaseline) (NativeEditabilityReport, error) {
	out := NativeEditabilityReport{Schema: NativeEditabilitySchema, Objects: []NativeEditabilityUnit{}, Counts: map[string]int{}, DesktopQualification: "not_recorded", Scope: "receipt_pinned_native_structure; selection, movement, resize and visual behavior require Mac and Windows desktop tasks"}
	if b == nil {
		return out, fmt.Errorf("editability.verified_baseline_required")
	}
	for _, kind := range []string{"sp", "grpSp", "pic", "graphicFrame", "cxnSp"} {
		out.Counts["native_"+kind] = 0
	}
	out.Counts["maximum_group_depth"] = 0
	out.Counts["text_and_paint_in_same_shape"] = 0
	var stored Receipt
	if e := strictInto(json.RawMessage(b.files["receipt.json"]), &stored); e != nil {
		return out, e
	}
	if !bytes.Equal(canonical(stored), canonical(b.Receipt)) || digest(b.files["receipt.json"]) != b.ReceiptSHA256 || !bytes.Equal(canonical(b.Objects), b.files["object-map.json"]) {
		return out, fmt.Errorf("editability.baseline_metadata_changed")
	}
	out.ProjectID, out.BuildID, out.ReceiptSHA256, out.PPTXSHA256, out.GenerationToken = b.Receipt.ProjectID, b.Receipt.BuildID, b.ReceiptSHA256, b.Receipt.Outputs["deck.pptx"], b.Objects.Lineage.BuildToken
	var layout wmdesign.Report
	if e := strictInto(json.RawMessage(b.files["layout-report.json"]), &layout); e != nil {
		return out, e
	}
	// Family labels come from declared compiled component definitions, not from
	// guessing a source match from visible text, current object names or geometry.
	definitions := map[string]string{}
	for _, slide := range layout.Slides {
		for _, component := range slide.Components {
			definitions[slide.ID+"\x00"+component.ID] = component.Definition
		}
		for _, row := range slide.CardRows {
			definitions[slide.ID+"\x00"+row.ID] = row.Definition
		}
		for _, scene := range slide.Scenes {
			if scene.Definition != "" {
				definitions[slide.ID+"\x00"+scene.ID] = scene.Definition
			}
			for _, group := range scene.Groups {
				definitions[slide.ID+"\x00"+group.ID] = group.Definition
			}
		}
	}
	records := map[string]ObjectRecord{}
	children := map[string][]string{}
	native := map[string]NativeLineageObject{}
	nativeIDs := map[string][]NativeLineageObject{}
	for _, object := range b.inspection.Objects {
		native[object.ShapeToken] = object
		key := object.NativePart + "\x00" + object.NativeID
		nativeIDs[key] = append(nativeIDs[key], object)
	}
	for _, object := range b.Objects.Objects {
		if object.ShapeToken == "" || records[object.ShapeToken].ShapeToken != "" {
			return out, fmt.Errorf("editability.invalid_shape_identity")
		}
		records[object.ShapeToken] = object
		if object.NativeParentToken != "" {
			children[object.NativeParentToken] = append(children[object.NativeParentToken], object.ShapeToken)
		}
	}
	for _, object := range b.Objects.Objects {
		unit := NativeEditabilityUnit{ShapeToken: object.ShapeToken, SlideID: object.SlideID, LogicalID: object.LogicalID, SelectionName: object.NativeName, NativeKind: object.NativeKind, Family: "shape", ParentToken: object.NativeParentToken, TopLevelToken: object.ShapeToken, ParagraphCount: len(object.Paragraphs), ChildTokens: append([]string{}, children[object.ShapeToken]...), Fields: append([]NativeSourceField{}, object.Fields...), Caveats: []string{}}
		unit.Definition = definitions[object.SlideID+"\x00"+object.NativeName]
		if unit.Definition == "scene.connector" && object.NativeKind == "grpSp" {
			unit.Caveats = append(unit.Caveats, "connector_is_grouped_segments_and_arrowheads; endpoint_attachment_is_not_established")
		}
		parent := object.NativeParentToken
		seen := map[string]bool{object.ShapeToken: true}
		for parent != "" {
			ancestor, ok := records[parent]
			if !ok || seen[parent] || ancestor.NativeKind != "grpSp" || ancestor.SlideID != object.SlideID || unit.GroupDepth >= 100 {
				return out, fmt.Errorf("editability.invalid_group_ownership")
			}
			seen[parent] = true
			unit.GroupDepth++
			unit.TopLevelToken, parent = parent, ancestor.NativeParentToken
		}
		if unit.GroupDepth > 1 {
			unit.Caveats = append(unit.Caveats, "nested_selection_requires_desktop_task")
		}
		switch object.NativeKind {
		case "grpSp":
			unit.Family = "group"
			unit.Caveats = append(unit.Caveats, "group_presence_does_not_qualify_movement_or_resize")
		case "pic":
			unit.Family = "image"
		case "cxnSp":
			unit.Family = "connector"
			unit.Caveats = append(unit.Caveats, "endpoint_attachment_and_adjustment_unqualified")
		}
		shape := native[object.ShapeToken].shape
		if shape == nil {
			return out, fmt.Errorf("editability.missing_native_object")
		}
		if object.NativeKind == "cxnSp" {
			props := directXML(directXML(shape, lineagePML, "nvCxnSpPr"), lineagePML, "cNvCxnSpPr")
			endpoint := func(name string) NativeEditabilityEndpoint {
				out := NativeEditabilityEndpoint{Site: -1, Status: "missing_or_ambiguous"}
				if props == nil {
					return out
				}
				var matches []*xmlNode
				for _, child := range props.Children {
					if child.Name.Space == drawingML && child.Name.Local == name {
						matches = append(matches, child)
					}
				}
				if len(matches) != 1 {
					return out
				}
				site, e := strconv.Atoi(attr(matches[0], "idx"))
				if e != nil || site < 0 {
					return out
				}
				out.Site = site
				targets := nativeIDs[object.NativePart+"\x00"+attr(matches[0], "id")]
				if len(targets) != 1 || targets[0].ShapeToken == "" {
					return out
				}
				out.ShapeToken, out.Status = targets[0].ShapeToken, "declared_native_reference"
				return out
			}
			unit.Connection = &NativeEditabilityConnection{Begin: endpoint("stCxn"), End: endpoint("endCxn")}
		}
		if object.NativeKind == "graphicFrame" {
			unit.Family = "graphic_frame"
			if len(descendants(shape, "tbl")) != 0 {
				unit.Family = "table"
				unit.Caveats = append(unit.Caveats, "native_cell_editing_unqualified; source_cell_adoption_unsupported")
			}
			if len(descendants(shape, "chart")) != 0 {
				unit.Family = "chart"
				unit.Caveats = append(unit.Caveats, "native_data_editing_unqualified; source_chart_adoption_unsupported")
			}
		}
		cells := map[string]bool{}
		for _, paragraph := range object.Paragraphs {
			if paragraph.Address.TableRow != nil && paragraph.Address.TableColumn != nil {
				cells[fmt.Sprintf("%d/%d", *paragraph.Address.TableRow, *paragraph.Address.TableColumn)] = true
			}
		}
		unit.TableCells = len(cells)
		if object.NativeKind == "sp" && object.NativeText != "" {
			props := directXML(shape, lineagePML, "spPr")
			if props != nil {
				unit.TextAndPaint = directXML(props, drawingML, "solidFill") != nil || directXML(props, drawingML, "gradFill") != nil || directXML(props, drawingML, "pattFill") != nil || directXML(props, drawingML, "blipFill") != nil
			}
		}
		if len(object.Fields) == 0 && object.NativeText != "" {
			unit.Caveats = append(unit.Caveats, "text_has_no_unique_source_field")
		}
		for _, field := range object.Fields {
			out.Counts["field_"+field.Status]++
			if field.Status != "plain_text_baseline" {
				unit.Caveats = append(unit.Caveats, "source_field_requires_manual_review: "+field.Reason)
			}
		}
		sort.Strings(unit.ChildTokens)
		sort.Strings(unit.Caveats)
		out.Objects = append(out.Objects, unit)
		out.Counts["native_"+unit.NativeKind]++
		out.Counts["family_"+unit.Family]++
		if unit.Definition != "" {
			out.Counts["definition_"+unit.Definition]++
		}
		if unit.GroupDepth > out.Counts["maximum_group_depth"] {
			out.Counts["maximum_group_depth"] = unit.GroupDepth
		}
		if unit.TextAndPaint {
			out.Counts["text_and_paint_in_same_shape"]++
		}
		if unit.ParentToken == "" {
			out.Counts["top_level_objects"]++
		}
	}
	sort.Slice(out.Objects, func(i, j int) bool {
		a, z := out.Objects[i], out.Objects[j]
		return a.SlideID+"\x00"+a.LogicalID+"\x00"+a.ShapeToken < z.SlideID+"\x00"+z.LogicalID+"\x00"+z.ShapeToken
	})
	out.Counts["objects"] = len(out.Objects)
	return out, nil
}
