package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
)

// CardRowsContract implements a bounded keyed repetition of the existing cards.
const CardRowsContract = "pptxgengo.wmds-card-rows.v1"

// CardRowSpec owns the whole row rectangle. Each child has an equal width, with
// the frozen 18pt grid gutter. Numbering is absent, inline, or band.
type CardRowSpec struct {
	Items     []CardRowItem `json:"items"`
	Numbering string        `json:"numbering,omitempty"`
}

type CardRowItem struct {
	Key     string   `json:"key"`
	Surface string   `json:"surface,omitempty"`
	Card    CardSpec `json:"card"`
}

// Parts identifies native child groups rather than their constituent shapes.
type CardRowRecord struct {
	ID             string   `json:"id"`
	Definition     string   `json:"definition"`
	Contract       string   `json:"contract"`
	Rect           Rect     `json:"rect"`
	RequiredHeight float64  `json:"required_height_pt"`
	Gap            float64  `json:"gap_pt"`
	Numbering      string   `json:"numbering"`
	Keys           []string `json:"keys"`
	Parts          []string `json:"parts"`
}

type cardRowPlan struct {
	record   CardRowRecord
	children []componentPlan
}

func (r *renderer) planCardRow(n Node, b, zone Rect, surface string) (cardRowPlan, error) {
	p := cardRowPlan{record: CardRowRecord{ID: n.ID, Definition: "card.presets", Contract: CardRowsContract, Rect: b, Gap: 18, Numbering: "none"}}
	if r.typeEngine.engine != CandidateEngine {
		return p, fmt.Errorf("cardrow.requires_v2_engine")
	}
	if n.CardRow == nil || n.Card != nil || n.TextBlock != nil || n.Text != "" || n.Style != "" || n.Ink != "" || n.Align != "" {
		return p, fmt.Errorf("cardrow.invalid_or_conflicting_payload: %s", n.ID)
	}
	if n.Grid == "five-up" {
		return p, fmt.Errorf("cardrow.five_up_unsupported")
	}
	v := n.CardRow
	if len(v.Items) < 2 || len(v.Items) > 4 {
		return p, fmt.Errorf("cardrow.item_count: 2 to 4 items required")
	}
	switch v.Numbering {
	case "", "none":
	case "inline", "band":
		p.record.Numbering = v.Numbering
		p.record.Definition = "card.numbered"
	default:
		return p, fmt.Errorf("cardrow.unsupported_numbering: %s", v.Numbering)
	}
	w := (b.W - 18*float64(len(v.Items)-1)) / float64(len(v.Items))
	if w < 198-.01 {
		return p, fmt.Errorf("cardrow.minimum_child_width: each card needs at least 198pt")
	}
	keys := map[string]bool{}
	nodes := make([]Node, 0, len(v.Items))
	for i, it := range v.Items {
		if !validPartKey(it.Key) || keys[it.Key] {
			return p, fmt.Errorf("cardrow.invalid_or_duplicate_key: %s", it.Key)
		}
		keys[it.Key] = true
		c := it.Card
		if p.record.Numbering != "none" {
			if c.InlineNumber != "" {
				return p, fmt.Errorf("cardrow.numbering_conflicts_with_card_number: %s", it.Key)
			}
			if c.Title == "" {
				return p, fmt.Errorf("cardrow.numbering_requires_title: %s", it.Key)
			}
			if p.record.Numbering == "inline" && c.Band != nil {
				return p, fmt.Errorf("cardrow.inline_numbering_excludes_band: %s", it.Key)
			}
			if p.record.Numbering == "band" {
				if c.Band == nil {
					return p, fmt.Errorf("cardrow.band_numbering_requires_band: %s", it.Key)
				}
				if c.NumberInk != "" && c.NumberInk != "emphasis" {
					return p, fmt.Errorf("cardrow.band_number_ink_requires_emphasis: %s", it.Key)
				}
				c.NumberInk = "emphasis"
			}
			c.InlineNumber = fmt.Sprintf("%02d", i+1)
		}
		surf := it.Surface
		if surf == "" {
			surf = surface
		}
		if surf == "" {
			surf = "light"
		}
		child := Node{ID: n.ID + ".items." + it.Key, Kind: "card", Scope: n.Scope, Surface: surf, Card: &c}
		cb := Rect{b.X + float64(i)*(w+18), b.Y, w, 0}
		// First plan all cards at their minimum measured height. Replanning at
		// the shared height below also positions metric footers correctly.
		cp, e := r.planComponent(child, cb, zone, surf)
		if e != nil {
			return p, e
		}
		p.record.RequiredHeight = math.Max(p.record.RequiredHeight, cp.record.RequiredHeight)
		p.record.Keys = append(p.record.Keys, it.Key)
		p.record.Parts = append(p.record.Parts, child.ID)
		nodes = append(nodes, child)
	}
	if b.H == 0 {
		b.H = math.Ceil(p.record.RequiredHeight/18) * 18
	}
	if p.record.RequiredHeight > b.H+.02 {
		return p, fmt.Errorf("cardrow.vertical_overflow: %s needs %.3fpt, capacity %.3fpt", n.ID, p.record.RequiredHeight, b.H)
	}
	if e := r.source.Tokens.Grid.OuterBox(b); e != nil {
		return p, e
	}
	if !inside(b, zone) {
		return p, fmt.Errorf("node.outside_zone: %s", n.ID)
	}
	for i, child := range nodes {
		cb := Rect{b.X + float64(i)*(w+18), b.Y, w, b.H}
		cp, e := r.planComponent(child, cb, zone, child.Surface)
		if e != nil {
			return p, e
		}
		p.children = append(p.children, cp)
	}
	p.record.Rect = b
	return p, r.err
}

func (r *renderer) drawCardRow(p cardRowPlan) {
	for _, c := range p.children {
		r.drawComponent(c)
	}
}

type nativeCardGroup struct {
	start, end int
	name       string
}

// nativeCardGroups identifies only direct spTree groups, using balanced XML
// tokens. Nested child groups cannot be mistaken for sibling card groups.
func nativeCardGroups(b []byte) ([]nativeCardGroup, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	depth, treeDepth, groupDepth := 0, -1, -1
	current := nativeCardGroup{}
	var out []nativeCardGroup
	for {
		before := int(d.InputOffset())
		t, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		switch t := t.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "spTree" {
				treeDepth = depth
			}
			if t.Name.Local == "grpSp" && depth == treeDepth+1 {
				at := bytes.LastIndex(b[before:int(d.InputOffset())], []byte("<"))
				if at < 0 {
					return nil, fmt.Errorf("cardrow.invalid_native_group_start")
				}
				current = nativeCardGroup{start: before + at}
				groupDepth = depth
			}
			if t.Name.Local == "cNvPr" && groupDepth >= 0 && depth == groupDepth+2 {
				for _, a := range t.Attr {
					if a.Name.Local == "name" {
						current.name = a.Value
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local == "grpSp" && depth == groupDepth {
				current.end = int(d.InputOffset())
				out = append(out, current)
				groupDepth = -1
			}
			if t.Name.Local == "spTree" {
				treeDepth = -1
			}
			depth--
		}
	}
	return out, nil
}

// cardRowGroups runs after componentGroups. It wraps the actual card groups in
// a composite group while retaining their transforms and native editable parts.
func cardRowGroups(raw []byte, slides []SlideReport) ([]byte, error) {
	byPart := map[string][]CardRowRecord{}
	for i, s := range slides {
		if len(s.CardRows) > 0 {
			byPart[fmt.Sprintf("ppt/slides/slide%d.xml", i+1)] = s.CardRows
		}
	}
	if len(byPart) == 0 {
		return raw, nil
	}
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		return nil, e
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	ids := regexp.MustCompile(`<p:cNvPr\b[^>]*\bid="(\d+)"`)
	for _, f := range z.File {
		rc, e := f.Open()
		if e != nil {
			return nil, e
		}
		b, e := io.ReadAll(rc)
		rc.Close()
		if e != nil {
			return nil, e
		}
		for _, row := range byPart[f.Name] {
			groups, e := nativeCardGroups(b)
			if e != nil {
				return nil, e
			}
			owned := map[string]bool{}
			for _, part := range row.Parts {
				if owned[part] {
					return nil, fmt.Errorf("cardrow.duplicate_native_child: %s", part)
				}
				owned[part] = true
			}
			first, last, count := -1, -1, 0
			var children bytes.Buffer
			for _, g := range groups {
				if !owned[g.name] {
					continue
				}
				if count >= len(row.Parts) || g.name != row.Parts[count] {
					return nil, fmt.Errorf("cardrow.native_child_order_mismatch: %s", row.ID)
				}
				if first < 0 {
					first = g.start
				} else if len(bytes.TrimSpace(b[last:g.start])) != 0 {
					return nil, fmt.Errorf("cardrow.noncontiguous_native_children: %s", row.ID)
				}
				last = g.end
				count++
				children.Write(b[g.start:g.end])
			}
			if count != len(row.Parts) || count == 0 {
				return nil, fmt.Errorf("cardrow.missing_native_children: %s", row.ID)
			}
			next := 1
			for _, m := range ids.FindAllSubmatch(b, -1) {
				id, _ := strconv.Atoi(string(m[1]))
				if id >= next {
					next = id + 1
				}
			}
			var name bytes.Buffer
			xml.EscapeText(&name, []byte(row.ID))
			emu := func(v float64) int64 { return int64(math.Round(v * 12700)) }
			r := row.Rect
			header := fmt.Sprintf(`<p:grpSp><p:nvGrpSpPr><p:cNvPr id="%d" name="%s"/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/><a:chOff x="%d" y="%d"/><a:chExt cx="%d" cy="%d"/></a:xfrm></p:grpSpPr>`, next, name.String(), emu(r.X), emu(r.Y), emu(r.W), emu(r.H), emu(r.X), emu(r.Y), emu(r.W), emu(r.H))
			b = bytes.Join([][]byte{b[:first], []byte(header), children.Bytes(), []byte("</p:grpSp>"), b[last:]}, nil)
		}
		dst, e := w.CreateHeader(&f.FileHeader)
		if e != nil {
			return nil, e
		}
		if _, e = dst.Write(b); e != nil {
			return nil, e
		}
	}
	if e = w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
