package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// These source structs list consumed rendering fields explicitly. Frozen catalog
// metadata is retained separately; it does not silently add rendering semantics.
type templateSourceCatalog struct {
	Schema    string            `json:"schema"`
	Family    string            `json:"family"`
	Name      string            `json:"name"`
	Summary   string            `json:"summary"`
	Templates []json.RawMessage `json:"templates"`
}
type templateSourceEntry struct {
	ID      string          `json:"id"`
	Variant string          `json:"variant"`
	Name    string          `json:"name"`
	Tier    string          `json:"tier"`
	Purpose string          `json:"purpose"`
	Uses    []string        `json:"uses"`
	Legacy  string          `json:"legacy"`
	Budget  json.RawMessage `json:"budget"`
	Slots   []struct {
		Slot     string `json:"slot"`
		MaxChars int    `json:"maxChars"`
		Count    int    `json:"count"`
		TitleMax int    `json:"titleMax"`
		BodyMax  int    `json:"bodyMax"`
	} `json:"slots"`
	Slide templateSourceSlide `json:"slide"`
}
type templateSourceSlide struct {
	Type        string `json:"type"`
	Rail        string `json:"rail"`
	Footer      string `json:"footer"`
	Surface     string `json:"surface,omitempty"`
	RailSurface string `json:"railSurface,omitempty"`
	Eyebrow     string `json:"eyebrow"`
	Title       string `json:"title"`
	Page        string `json:"page,omitempty"`
	Emphasis    string `json:"emphasis,omitempty"`
	Source      *struct {
		Text string `json:"text"`
	} `json:"source,omitempty"`
	Body []json.RawMessage `json:"body"`
}
type templateSourceText struct {
	Type  string  `json:"type"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	Style string  `json:"style"`
	Ink   string  `json:"ink"`
	Text  string  `json:"text"`
	On    string  `json:"on,omitempty"`
}
type templateSourceMetric struct {
	Type  string  `json:"type"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	Value string  `json:"value"`
	Label string  `json:"label"`
}
type templateSourceRule struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	W    float64 `json:"w"`
}
type templateSourceCardRow struct {
	Type      string  `json:"type"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	W         float64 `json:"w"`
	H         float64 `json:"h"`
	Gap       float64 `json:"gap"`
	Numbering string  `json:"numbering"`
	Card      struct {
		Surface string    `json:"surface"`
		Pad     float64   `json:"pad"`
		Band    *CardBand `json:"band,omitempty"`
		NumInk  string    `json:"numInk,omitempty"`
	} `json:"card"`
	Items []struct {
		Title string `json:"title"`
		Body  []struct {
			Paragraph string `json:"p"`
		} `json:"body"`
	} `json:"items"`
}

func decodeTemplateSource(raw []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return fmt.Errorf("template.unsupported_source_field: %w", e)
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return fmt.Errorf("template.invalid_source_json")
	}
	return nil
}

// compileTemplateSource reads exact frozen variants and clears their example
// copy. Binding supplies all presentation content; geometry stays source bound.
func compileTemplateSource(s *Source, key string) (SlideSpec, TemplateDefinition, error) {
	var slide SlideSpec
	def := TemplateDefinition{Key: key}
	var family, id, variant string
	var names, expected []string
	switch key {
	case "cards/3", "cards/4":
		family, id, variant = "openers", "cards", key[len("cards/"):]
		names, expected = []string{"cards"}, []string{"cardrow"}
	case "stats/four-metrics":
		family, id, variant = "evidence", "stats", "four-metrics"
		names = []string{"divider", "metric1", "metric2", "metric3", "metric4", "support"}
		expected = []string{"rule", "metric", "metric", "metric", "metric", "text"}
	case "takeaway-rail/metrics-rail":
		family, id, variant = "argument", "takeaway-rail", "metrics-rail"
		names = []string{"metric1", "metric2", "divider", "support", "rail-eyebrow", "rail-heading", "rail-body"}
		expected = []string{"metric", "metric", "rule", "text", "text", "text", "text"}
	default:
		return slide, def, fmt.Errorf("template.unsupported_key: %s", key)
	}
	def.SourceFile = "templates/library/" + family + ".json"
	raw, ok := s.Templates[def.SourceFile]
	if !ok {
		return slide, def, fmt.Errorf("template.source_not_found: %s", def.SourceFile)
	}
	for _, f := range s.Files {
		if f.Path == def.SourceFile {
			if def.SourceSHA256 != "" {
				return slide, def, fmt.Errorf("template.duplicate_source_identity: %s", def.SourceFile)
			}
			def.SourceSHA256 = f.SHA256
		}
	}
	if def.SourceSHA256 == "" || fmt.Sprintf("%x", sha256.Sum256(raw)) != def.SourceSHA256 {
		return slide, def, fmt.Errorf("template.source_hash_mismatch: %s", def.SourceFile)
	}
	var catalog templateSourceCatalog
	if e := decodeTemplateSource(raw, &catalog); e != nil {
		return slide, def, e
	}
	if catalog.Schema != "wmds.templates.v2" || catalog.Family != family {
		return slide, def, fmt.Errorf("template.unsupported_catalog_schema")
	}
	var entry templateSourceEntry
	found := false
	for _, candidate := range catalog.Templates {
		var identity struct {
			ID      string `json:"id"`
			Variant string `json:"variant"`
		}
		if e := json.Unmarshal(candidate, &identity); e != nil {
			return slide, def, e
		}
		if identity.ID != id || identity.Variant != variant {
			continue
		}
		if found {
			return slide, def, fmt.Errorf("template.duplicate_variant: %s", key)
		}
		if e := decodeTemplateSource(candidate, &entry); e != nil {
			return slide, def, e
		}
		found = true
	}
	if !found {
		return slide, def, fmt.Errorf("template.variant_not_found: %s", key)
	}
	if entry.Slide.Type != "slide" || len(entry.Slide.Body) != len(names) {
		return slide, def, fmt.Errorf("template.source_structure_mismatch: %s", key)
	}
	if entry.Slide.Emphasis != "" && entry.Slide.Emphasis != "underscore" && entry.Slide.Emphasis != "highlight" {
		return slide, def, fmt.Errorf("template.unsupported_header_emphasis")
	}
	// Source marks belong to source example copy. Supplied plain titles contain
	// no marks, so this adapter leaves the source emphasis configuration inactive.
	q := FrameRequest{Rail: entry.Slide.Rail, Footer: entry.Slide.Footer, Surface: entry.Slide.Surface, RailSurface: entry.Slide.RailSurface, TitleLines: 1}
	if entry.Slide.Source != nil {
		q.SourceLines = 2
	}
	f, e := s.ResolveFrame(q)
	if e != nil {
		return slide, def, e
	}
	slide.Frame = f.Request
	def.Identities = append(def.Identities, TemplateIdentity{ID: "eyebrow", SourcePointer: "/eyebrow", ExpectedType: "string"}, TemplateIdentity{ID: "title", SourcePointer: "/title", ExpectedType: "string"})
	if entry.Slide.Source != nil {
		def.Identities = append(def.Identities, TemplateIdentity{ID: "source", SourcePointer: "/source/text", ExpectedType: "string"})
	}
	for _, hint := range entry.Slots {
		if hint.MaxChars > 0 {
			def.Guidance = append(def.Guidance, TemplateGuidance{Slot: hint.Slot, MaxCharacters: hint.MaxChars, Advisory: true})
		}
		if hint.TitleMax > 0 {
			def.Guidance = append(def.Guidance, TemplateGuidance{Slot: hint.Slot + "[].title", MaxCharacters: hint.TitleMax, Advisory: true})
		}
		if hint.BodyMax > 0 {
			def.Guidance = append(def.Guidance, TemplateGuidance{Slot: hint.Slot + "[].body", MaxCharacters: hint.BodyMax, Advisory: true})
		}
	}
	for i, body := range entry.Slide.Body {
		var tag struct {
			Type string `json:"type"`
		}
		if e := json.Unmarshal(body, &tag); e != nil {
			return slide, def, e
		}
		if tag.Type != expected[i] {
			return slide, def, fmt.Errorf("template.changed_target_type: %s /body/%d", key, i)
		}
		n := Node{ID: names[i], Kind: tag.Type}
		switch tag.Type {
		case "cardrow":
			var src templateSourceCardRow
			if e := decodeTemplateSource(body, &src); e != nil {
				return slide, def, e
			}
			count := 3
			if variant == "4" {
				count = 4
			}
			if len(src.Items) != count || src.Gap != 18 || src.Card.Pad != 18 || src.H <= 0 || src.W <= 0 || src.Numbering != "inline" && src.Numbering != "band" {
				return slide, def, fmt.Errorf("template.unsupported_card_row_source")
			}
			n.Rect = Rect{src.X, src.Y, float64(count)*src.W + float64(count-1)*src.Gap, src.H}
			n.Surface = src.Card.Surface
			n.CardRow = &CardRowSpec{Numbering: src.Numbering}
			for j, item := range src.Items {
				if len(item.Body) != 1 {
					return slide, def, fmt.Errorf("template.unsupported_card_body_source")
				}
				n.CardRow.Items = append(n.CardRow.Items, CardRowItem{Key: fmt.Sprintf("slot-%d", j+1), Card: CardSpec{Padding: src.Card.Pad, Band: src.Card.Band, NumberInk: src.Card.NumInk, Body: []BodyBlock{{Key: "copy"}}}})
			}
		case "metric":
			var src templateSourceMetric
			if e := decodeTemplateSource(body, &src); e != nil {
				return slide, def, e
			}
			n.Rect = Rect{X: src.X, Y: src.Y, W: src.W}
			n.DataMetric = &DataMetricSpec{}
		case "text":
			var src templateSourceText
			if e := decodeTemplateSource(body, &src); e != nil {
				return slide, def, e
			}
			n.Rect = Rect{X: src.X, Y: src.Y, W: src.W}
			n.Style, n.Ink, n.Surface = src.Style, src.Ink, src.On
			if _, e := s.Style(n.Style); e != nil {
				return slide, def, e
			}
		case "rule":
			var src templateSourceRule
			if e := decodeTemplateSource(body, &src); e != nil {
				return slide, def, e
			}
			n.Rect = Rect{src.X, src.Y, src.W, .75}
			n.Ink = "line"
		}
		b := n.Rect
		if math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) || b.W <= 0 {
			return slide, def, fmt.Errorf("template.invalid_source_geometry: %s", n.ID)
		}
		if f.Rail.W > 0 && b.X >= f.Rail.X-.01 && b.X+b.W <= f.Rail.X+f.Rail.W+.01 {
			n.Scope = "rail"
		}
		surface := f.Request.Surface
		if n.Scope == "rail" {
			surface = f.Request.RailSurface
		}
		if n.Surface == "" {
			n.Surface = surface
		} else if n.Kind == "text" && n.Surface != surface {
			return slide, def, fmt.Errorf("template.text_surface_zone_mismatch: %s", n.ID)
		}
		if _, e := s.Ink(n.Surface, "bg"); e != nil {
			return slide, def, e
		}
		if n.Ink != "" {
			if _, e := s.Ink(n.Surface, n.Ink); e != nil {
				return slide, def, e
			}
		}
		slide.Nodes = append(slide.Nodes, n)
		def.Identities = append(def.Identities, TemplateIdentity{ID: n.ID, SourcePointer: fmt.Sprintf("/body/%d", i), ExpectedType: tag.Type})
	}
	// Metric rows share their next source flow boundary, including the portion
	// above a narrower support paragraph. Plain text ends before the next sibling
	// with overlapping x extent. Neither calculation rewrites source x/y/width.
	for i := range slide.Nodes {
		n := &slide.Nodes[i]
		zone := f.Body
		if n.Scope == "rail" {
			zone = f.Rail
		}
		if n.Rect.H == 0 {
			bottom := zone.Y + zone.H
			for j, next := range slide.Nodes {
				if j == i || next.Scope != n.Scope || next.Rect.Y <= n.Rect.Y {
					continue
				}
				overlap := n.Rect.X < next.Rect.X+next.Rect.W && next.Rect.X < n.Rect.X+n.Rect.W
				if n.Kind == "metric" || overlap {
					bottom = math.Min(bottom, next.Rect.Y)
				}
			}
			n.Rect.H = bottom - n.Rect.Y
		}
		if n.Rect.H <= 0 || !inside(n.Rect, zone) {
			return slide, def, fmt.Errorf("template.source_outside_zone: %s", n.ID)
		}
		if n.Kind != "rule" {
			if e := s.Tokens.Grid.OuterBox(n.Rect); e != nil {
				return slide, def, e
			}
		}
	}
	return slide, def, nil
}
