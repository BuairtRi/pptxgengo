package deckproject

import (
	"bytes"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Legacy fields retain whole-object comparison. Only an explicit paragraph
// contract can select a subset of native paragraphs for a maintained field.
func nativeObjectFieldText(object ObjectRecord, field NativeSourceField, paragraphs []NativeParagraph) (string, bool) {
	if object.TextMappingContract == "" {
		return nativeParagraphText(paragraphs), true
	}
	return nativeFieldText(field, paragraphs)
}

func editableCardBaselineParagraphsValid(object ObjectRecord, actual []NativeParagraph) bool {
	return bytes.Equal(canonical(object.Paragraphs), canonical(actual)) && editableCardFieldContractValid(object)
}

func nativeFieldText(field NativeSourceField, paragraphs []NativeParagraph) (string, bool) {
	if len(field.Addresses) == 0 {
		return "", false
	}
	texts := []string{}
	seen := map[NativeTextAddress]bool{}
	for _, address := range field.Addresses {
		if seen[address] {
			return "", false
		}
		seen[address] = true
		found := false
		for _, paragraph := range paragraphs {
			if paragraph.Address == address {
				if found {
					return "", false
				}
				found = true
				texts = append(texts, paragraph.Text)
			}
		}
		if !found {
			return "", false
		}
	}
	return strings.Join(texts, "\n"), true
}

func cardPlainParagraph(p NativeParagraph) bool {
	return p.Address.Body == 0 && p.Address.TableRow == nil && p.Address.TableColumn == nil && len(p.ReviewItems) == 0 && len(p.Runs) == 1 && p.Runs[0].Kind == "r" && strings.TrimSpace(p.Text) != "" && !strings.ContainsAny(p.Text, "\r\n\t") && !strings.Contains(p.Text, "[[") && !strings.Contains(p.Text, "]]") && !strings.Contains(p.Text, "[^")
}

func attachEditableCardFields(p *Project, r ObjectRecord) ObjectRecord {
	r.TextMapping = "manual_review"
	r.TextMappingReason = "Editable card requires two distinct explicit title/body bindings and exact plain paragraph baselines."
	if r.NativeKind != "sp" || len(r.Paragraphs) != 2 || len(r.ParagraphSourceSlots) != 2 || r.ParagraphSourceSlots["title"] == r.ParagraphSourceSlots["body"] {
		return r
	}
	fields := []NativeSourceField{}
	for i, role := range []string{"title", "body"} {
		para := r.Paragraphs[i]
		if !cardPlainParagraph(para) || para.Address.Paragraph != i {
			return r
		}
		slot := r.ParagraphSourceSlots[role]
		pointer := ""
		for path, bound := range r.SourceSlots {
			if bound == slot {
				if pointer != "" {
					return r
				}
				pointer = path
			}
		}
		value, err := lookupPointer(p.tree, pointer)
		copy, ok := value.(string)
		if err != nil || !ok || copy != para.Text || slot == "" {
			return r
		}
		fields = append(fields, NativeSourceField{Identity: r.LogicalID + "/paragraph/" + role, SourceSlot: slot, SourcePointer: pointer, SourceValueSHA256: digest([]byte(copy)), BaselineNativeSHA256: digest([]byte(para.Text)), Addresses: []NativeTextAddress{para.Address}, Status: "plain_text_baseline"})
	}
	r.Fields = fields
	r.TextMapping = "plain_text_baseline"
	r.TextMappingReason = ""
	return r
}

func nativeFieldIdentityValid(object ObjectRecord, field NativeSourceField) bool {
	if object.TextMappingContract == "" {
		return field.Identity == object.LogicalID+"/text"
	}
	if object.TextMappingContract != wmdesign.EditableCardContract {
		return false
	}
	for i, role := range []string{"title", "body"} {
		if field.Identity == object.LogicalID+"/paragraph/"+role {
			return field.SourceSlot == object.ParagraphSourceSlots[role] && len(field.Addresses) == 1 && field.Addresses[0] == (NativeTextAddress{Body: 0, Paragraph: i})
		}
	}
	return false
}

func editableCardFieldContractValid(object ObjectRecord) bool {
	if object.TextMappingContract != wmdesign.EditableCardContract || object.NativeKind != "sp" || len(object.Paragraphs) != 2 {
		return false
	}
	if len(object.Fields) == 0 {
		return object.TextMapping == "manual_review"
	}
	if len(object.ParagraphSourceSlots) != 2 || object.ParagraphSourceSlots["title"] == "" || object.ParagraphSourceSlots["body"] == "" || object.ParagraphSourceSlots["title"] == object.ParagraphSourceSlots["body"] {
		return false
	}
	if len(object.Fields) != 2 {
		return false
	}
	seen := map[string]bool{}
	for _, field := range object.Fields {
		if !nativeFieldIdentityValid(object, field) || seen[field.Identity] || field.Status != "plain_text_baseline" {
			return false
		}
		seen[field.Identity] = true
	}
	return editableCardEditedParagraphsSupported(object, object.Paragraphs)
}

func editableCardEditedParagraphsSupported(object ObjectRecord, paragraphs []NativeParagraph) bool {
	if len(paragraphs) != 2 || len(object.Paragraphs) != 2 {
		return false
	}
	for i, p := range paragraphs {
		b := object.Paragraphs[i]
		if !cardPlainParagraph(p) || p.Address != b.Address || p.PropertiesSHA256 != b.PropertiesSHA256 || p.EndPropertiesSHA256 != b.EndPropertiesSHA256 || len(b.Runs) != 1 || p.Runs[0].PropertiesSHA256 != b.Runs[0].PropertiesSHA256 {
			return false
		}
	}
	return true
}
