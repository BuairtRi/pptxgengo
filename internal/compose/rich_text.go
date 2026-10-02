package compose

import (
	"fmt"
	"strings"
)

// ParagraphSpec is one native PowerPoint paragraph. Rich paragraphs are
// intentionally bounded to explicit font families in runs and point paragraph spacing.
type ParagraphSpec struct {
	ID                  string      `json:"id"`
	Align               string      `json:"align"` // left, center, right
	SpaceBeforePt       float64     `json:"space_before_pt,omitempty"`
	SpaceAfterPt        float64     `json:"space_after_pt,omitempty"`
	LineSpacingMultiple float64     `json:"line_spacing_multiple,omitempty"`
	Bullet              *BulletSpec `json:"bullet,omitempty"`
	Runs                []RunSpec   `json:"runs"`
}

// BulletSpec is a single editable, unnumbered PowerPoint bullet. MarginLeftPt
// positions the text; HangingPt moves the bullet left from that text position.
type BulletSpec struct {
	Character    string  `json:"character"`
	MarginLeftPt float64 `json:"margin_left_pt"`
	HangingPt    float64 `json:"hanging_pt"`
}

type RunSpec struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	FontFace   string  `json:"font_face"`
	FontSizePt float64 `json:"font_size_pt"`
	Bold       bool    `json:"bold,omitempty"`
	Italic     bool    `json:"italic,omitempty"`
	Underline  bool    `json:"underline,omitempty"`
	Foreground string  `json:"foreground"`
}

func richText(paragraphs []ParagraphSpec) string {
	var ps []string
	for _, p := range paragraphs {
		var b strings.Builder
		for _, r := range p.Runs {
			b.WriteString(r.Text)
		}
		ps = append(ps, b.String())
	}
	return strings.Join(ps, "\n")
}

func validateRichText(paragraphs []ParagraphSpec, background string) error {
	if err := validateRichTextStructure(paragraphs); err != nil {
		return err
	}
	for _, p := range paragraphs {
		for _, r := range p.Runs {
			if _, err := foreground(r.Foreground, background); err != nil {
				return fmt.Errorf("paragraph %s run %s: %w", p.ID, r.ID, err)
			}
		}
	}
	return nil
}

func validateRichTextStructure(paragraphs []ParagraphSpec) error {
	if len(paragraphs) == 0 {
		return fmt.Errorf("rich text requires at least one paragraph")
	}
	paragraphIDs := map[string]bool{}
	for _, p := range paragraphs {
		if !validID(p.ID) || paragraphIDs[p.ID] {
			return fmt.Errorf("paragraph IDs must be nonempty and unique: %q", p.ID)
		}
		paragraphIDs[p.ID] = true
		if p.Align != "left" && p.Align != "center" && p.Align != "right" {
			return fmt.Errorf("paragraph %s requires left/center/right alignment", p.ID)
		}
		if !nonnegative(p.SpaceBeforePt) || !nonnegative(p.SpaceAfterPt) {
			return fmt.Errorf("paragraph %s spacing must be finite and nonnegative", p.ID)
		}
		if p.LineSpacingMultiple != 0 && (!finite(p.LineSpacingMultiple) || p.LineSpacingMultiple < .5 || p.LineSpacingMultiple > 4) {
			return fmt.Errorf("paragraph %s line spacing multiple must be zero/default or in [0.5,4]", p.ID)
		}
		if len(p.Runs) == 0 {
			return fmt.Errorf("paragraph %s requires at least one run", p.ID)
		}
		if p.Bullet != nil {
			if p.Align != "left" {
				return fmt.Errorf("paragraph %s bullets require left alignment", p.ID)
			}
			if p.Bullet.Character != "•" && p.Bullet.Character != "–" {
				return fmt.Errorf("paragraph %s bullet character must be • or – (square bullets are not visually qualified)", p.ID)
			}
			if !finite(p.Bullet.MarginLeftPt) || p.Bullet.MarginLeftPt < 1 || p.Bullet.MarginLeftPt > 144 || !finite(p.Bullet.HangingPt) || p.Bullet.HangingPt < .75*p.Runs[0].FontSizePt || p.Bullet.HangingPt > p.Bullet.MarginLeftPt {
				return fmt.Errorf("paragraph %s bullet margin must be in [1,144]pt and hanging indent at least 75%% of the first run font size and no greater than the margin", p.ID)
			}
		}
		runIDs := map[string]bool{}
		for _, r := range p.Runs {
			if !validID(r.ID) || runIDs[r.ID] || !validID(r.Text) {
				return fmt.Errorf("paragraph %s run IDs and text must be nonempty; IDs must be unique: %q", p.ID, r.ID)
			}
			runIDs[r.ID] = true
			if !ValidFontFace(r.FontFace) || !positive(r.FontSizePt) {
				return fmt.Errorf("paragraph %s run %s requires an explicit font family and a positive font size", p.ID, r.ID)
			}
			for _, ch := range r.Text {
				if ch == '\r' || ch == '\n' {
					return fmt.Errorf("paragraph %s run %s contains a paragraph break; use another ParagraphSpec", p.ID, r.ID)
				}
				if ch > 0xFFFF {
					return fmt.Errorf("paragraph %s run %s contains a non-BMP character unsupported by native style verification", p.ID, r.ID)
				}
				if (ch < 0x20 && ch != '\t') || ch == 0xFFFE || ch == 0xFFFF {
					return fmt.Errorf("paragraph %s run %s contains an unsupported XML text character", p.ID, r.ID)
				}
			}
			if _, err := resolveColor(r.Foreground); err != nil {
				return fmt.Errorf("paragraph %s run %s: %w", p.ID, r.ID, err)
			}
		}
	}
	if !validID(richText(paragraphs)) {
		return fmt.Errorf("rich text must contain visible content")
	}
	return nil
}

func paragraphLineSpacing(p ParagraphSpec) float64 {
	if p.LineSpacingMultiple == 0 {
		return 1
	}
	return p.LineSpacingMultiple
}

func validateRichBulletWidth(paragraphs []ParagraphSpec, width float64) error {
	for _, p := range paragraphs {
		if p.Bullet != nil && p.Bullet.MarginLeftPt >= width {
			return fmt.Errorf("paragraph %s bullet margin %.2fpt leaves no usable width in %.2fpt frame", p.ID, p.Bullet.MarginLeftPt, width)
		}
	}
	return nil
}

func resolveRichText(paragraphs []ParagraphSpec, background string) []ParagraphSpec {
	out := make([]ParagraphSpec, len(paragraphs))
	for i, p := range paragraphs {
		out[i] = p
		if p.Bullet != nil {
			bullet := *p.Bullet
			out[i].Bullet = &bullet
		}
		out[i].Runs = make([]RunSpec, len(p.Runs))
		for j, r := range p.Runs {
			out[i].Runs[j] = r
			out[i].Runs[j].Foreground, _ = foreground(r.Foreground, background)
		}
	}
	return out
}
