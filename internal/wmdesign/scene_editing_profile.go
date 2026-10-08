package wmdesign

import (
	"encoding/json"
	"fmt"
)

// NativeEditingProfile changes native object structure only for eligible source
// scenes. It is persisted in authored source; frozen bundles remain unchanged.
const NativeEditingProfile = "native-v1"

func ValidateEditingProfile(profile string) error {
	if profile != "" && profile != "stock" && profile != NativeEditingProfile {
		return fmt.Errorf("unknown editing profile: %s", profile)
	}
	return nil
}

func (r *renderer) applyNativeEditingProfile(p *scenePlan, raw json.RawMessage, ctx SceneContext) *scenePlan {
	if r.editingProfile != NativeEditingProfile || p == nil {
		return p
	}
	var tag struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(raw, &tag) != nil {
		return p
	}
	var candidate *scenePlan
	reason := "requires compatible measured geometry and supported native paragraph formatting"
	switch tag.Type {
	case "bullets":
		candidate, reason = r.nativeProfileListResult(p, raw)
	case "card":
		candidate, reason = nativeProfileCardResult(p, raw)
	case "table":
		if len(p.Items) == 1 && p.Items[0].Table != nil {
			copy := *p
			copy.Groups = nil
			candidate = &copy
		}
	default:
		return p
	}
	if candidate != nil {
		if err := sceneTextEnvelope(candidate, ctx); err != nil {
			p.Warnings = append(p.Warnings, "native-v1 retained source object structure: "+p.ID+" (converted text allocation crosses the declared content envelope).")
			return p
		}
	}
	if candidate == nil {
		p.Warnings = append(p.Warnings, "native-v1 retained source object structure: "+p.ID+" ("+reason+").")
		return p
	}
	candidate.Warnings = append(candidate.Warnings, "native-v1 converted source component: "+p.ID+"; source-resolved geometry, fonts and colors retained. Native edits require review; catalog previews do not qualify this editing profile.")
	return candidate
}

func plainProfileParagraph(tr TextRecord, key string) (RichParagraphLayout, bool) {
	if !editableCardPlain(tr.Layout.Displayed) || tr.Layout.Original != tr.Layout.Displayed || tr.Align != "left" || tr.Rotation != 0 || tr.NativeShape != nil {
		return RichParagraphLayout{}, false
	}
	if tr.Rich != nil {
		if len(tr.Rich.Paragraphs) != 1 || len(tr.Rich.Paragraphs[0].Runs) != 1 || len(tr.Rich.Paragraphs[0].LineBreaks) != 0 {
			return RichParagraphLayout{}, false
		}
		para := tr.Rich.Paragraphs[0]
		para.Key = key
		return para, true
	}
	return RichParagraphLayout{Key: key, Displayed: tr.Layout.Displayed, LineCount: len(tr.Layout.Lines), Runs: []RichRunLayout{{Key: "text", Original: tr.Layout.Original, Displayed: tr.Layout.Displayed, Style: tr.Layout.Style, Font: tr.Layout.Font, Color: tr.Color, Start: 0, End: len([]rune(tr.Layout.Displayed))}}}, true
}
