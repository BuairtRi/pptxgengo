package component

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

type runValue struct {
	BindingID string `json:"binding_id"`
	Text      string `json:"text"`
}
type paragraphValue struct {
	Runs []runValue `json:"runs"`
}
type structuredSlot struct {
	Paragraphs []paragraphValue `json:"paragraphs"`
}

type inspectedRun struct {
	BindingID       string            `json:"binding_id"`
	SourceText      string            `json:"source_text"`
	RunPath         []int             `json:"run_path"`
	StyleAttributes map[string]string `json:"style_attributes,omitempty"`
}
type inspectedParagraph struct {
	ParagraphPath []int          `json:"paragraph_path"`
	Runs          []inspectedRun `json:"runs"`
}

func nodeAt(root *nativepkg.Node, path []int) (*nativepkg.Node, error) {
	n := root
	for _, i := range path {
		if n == nil || i < 0 || i >= len(n.Children) {
			return nil, fmt.Errorf("invalid scene path %v", path)
		}
		n = n.Children[i]
	}
	if n == nil {
		return nil, fmt.Errorf("nil scene node at %v", path)
	}
	return n, nil
}
func paragraphAndRun(root *nativepkg.Node, b nativepkg.Binding) ([]int, []int, error) {
	if b.Property != "text" || len(b.NodePath) < 3 {
		return nil, nil, fmt.Errorf("binding %s is not a text run", b.BindingID)
	}
	textPath := b.NodePath[:len(b.NodePath)-1]
	textNode, e := nodeAt(root, textPath)
	if e != nil || textNode.Name != "a:t" {
		return nil, nil, fmt.Errorf("binding %s text node invalid", b.BindingID)
	}
	runPath := textPath[:len(textPath)-1]
	run, e := nodeAt(root, runPath)
	if e != nil || run.Name != "a:r" {
		return nil, nil, fmt.Errorf("binding %s run invalid", b.BindingID)
	}
	pPath := runPath[:len(runPath)-1]
	p, e := nodeAt(root, pPath)
	if e != nil || p.Name != "a:p" {
		return nil, nil, fmt.Errorf("binding %s paragraph invalid", b.BindingID)
	}
	return pPath, runPath, nil
}
func pathKey(p []int) string {
	var x []string
	for _, i := range p {
		x = append(x, strconv.Itoa(i))
	}
	return strings.Join(x, ".")
}
func inspectParagraphs(slots map[string]Slot, root *nativepkg.Node, all []nativepkg.Binding, idx map[string]int) (map[string][]inspectedParagraph, error) {
	out := map[string][]inspectedParagraph{}
	for name, slot := range slots {
		var paras []inspectedParagraph
		for _, id := range slot.BindingIDs {
			b := all[idx[id]]
			pp, rp, e := paragraphAndRun(root, b)
			if e != nil {
				return nil, e
			}
			if len(paras) == 0 || pathKey(paras[len(paras)-1].ParagraphPath) != pathKey(pp) {
				paras = append(paras, inspectedParagraph{ParagraphPath: append([]int(nil), pp...)})
			}
			run, _ := nodeAt(root, rp)
			attrs := map[string]string{}
			if pr := run.Child("rPr"); pr != nil {
				for _, a := range pr.Attrs {
					attrs[a.Name] = a.Value
				}
			}
			j := len(paras) - 1
			paras[j].Runs = append(paras[j].Runs, inspectedRun{BindingID: id, SourceText: b.Value, RunPath: append([]int(nil), rp...), StyleAttributes: attrs})
		}
		out[name] = paras
	}
	return out, nil
}
func parseSlotValue(raw json.RawMessage, slot Slot, root *nativepkg.Node, all []nativepkg.Binding, idx map[string]int) ([]string, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty slot value")
	}
	if raw[0] == '[' {
		if slot.ValueFormat == "paragraphs" {
			return nil, fmt.Errorf("structured paragraph values required")
		}
		var parts []string
		if e := decode(raw, &parts); e != nil {
			return nil, e
		}
		return parts, nil
	}
	var sv structuredSlot
	if e := decode(raw, &sv); e != nil {
		return nil, e
	}
	expected, e := inspectParagraphs(map[string]Slot{"slot": slot}, root, all, idx)
	if e != nil {
		return nil, e
	}
	paras := expected["slot"]
	if len(sv.Paragraphs) != len(paras) {
		return nil, fmt.Errorf("needs %d paragraphs; received %d", len(paras), len(sv.Paragraphs))
	}
	var parts []string
	for i, p := range sv.Paragraphs {
		if len(p.Runs) != len(paras[i].Runs) {
			return nil, fmt.Errorf("paragraph %d needs %d runs; received %d", i+1, len(paras[i].Runs), len(p.Runs))
		}
		for j, r := range p.Runs {
			if r.BindingID != paras[i].Runs[j].BindingID {
				return nil, fmt.Errorf("paragraph %d run %d binding ID/order mismatch", i+1, j+1)
			}
			parts = append(parts, r.Text)
		}
	}
	return parts, nil
}
func cloneNode(n *nativepkg.Node) *nativepkg.Node {
	if n == nil {
		return nil
	}
	c := &nativepkg.Node{Name: n.Name, Text: n.Text, Attrs: append([]nativepkg.Attr(nil), n.Attrs...)}
	for _, x := range n.Children {
		c.Children = append(c.Children, cloneNode(x))
	}
	return c
}
func insertChild(n *nativepkg.Node, at int, child *nativepkg.Node) {
	n.Children = append(n.Children, nil)
	copy(n.Children[at+1:], n.Children[at:])
	n.Children[at] = child
}
func validateText(value, where string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("empty text in %s", where)
	}
	for _, r := range value {
		if r < 32 || r == 127 {
			return fmt.Errorf("control/newline character in %s", where)
		}
	}
	if strings.HasPrefix(value, "__BINDING:") {
		return fmt.Errorf("reserved sentinel in %s", where)
	}
	return nil
}
func objectIDAt(root *nativepkg.Node, path []int) string {
	for k := len(path); k >= 0; k-- {
		n, e := nodeAt(root, path[:k])
		if e != nil {
			continue
		}
		if n.Name == "p:sp" {
			if nv := n.Child("nvSpPr"); nv != nil {
				if cn := nv.Child("cNvPr"); cn != nil {
					return cn.Attr("id")
				}
			}
		}
	}
	return ""
}
func enclosingShape(root *nativepkg.Node, path []int) (*nativepkg.Node, error) {
	for k := len(path); k >= 0; k-- {
		n, e := nodeAt(root, path[:k])
		if e == nil && n.Name == "p:sp" {
			return n, nil
		}
	}
	return nil, fmt.Errorf("no enclosing shape at %v", path)
}
func resolvedAttr(n *nativepkg.Node, name string, values map[string]string) (string, error) {
	v := n.Attr(name)
	if strings.HasPrefix(v, "__BINDING:") {
		r, ok := values[strings.TrimPrefix(v, "__BINDING:")]
		if !ok {
			return "", fmt.Errorf("unresolved geometry binding")
		}
		return r, nil
	}
	return v, nil
}
func shapeFrame(n *nativepkg.Node, values map[string]string) ([4]int64, error) {
	var frame [4]int64
	sp := n.Child("spPr")
	if sp == nil {
		return frame, fmt.Errorf("shape has no spPr")
	}
	xf := sp.Child("xfrm")
	if xf == nil || xf.Child("off") == nil || xf.Child("ext") == nil {
		return frame, fmt.Errorf("shape has no explicit frame")
	}
	for i, item := range []struct {
		node *nativepkg.Node
		key  string
	}{{xf.Child("off"), "x"}, {xf.Child("off"), "y"}, {xf.Child("ext"), "cx"}, {xf.Child("ext"), "cy"}} {
		v, e := resolvedAttr(item.node, item.key, values)
		if e != nil {
			return frame, e
		}
		frame[i], e = strconv.ParseInt(v, 10, 64)
		if e != nil {
			return frame, e
		}
	}
	return frame, nil
}
func validateZones(zones map[string]Zone, root *nativepkg.Node, all []nativepkg.Binding, idx map[string]int, objects map[string]bool, hexColor *regexp.Regexp) error {
	seen := map[string]bool{}
	newIDs := map[string]bool{}
	values := map[string]string{}
	for _, b := range all {
		values[b.BindingID] = b.Value
	}
	existingIDs := map[string]bool{}
	root.Walk(func(n *nativepkg.Node) {
		if n.Name == "p:cNvPr" {
			existingIDs[n.Attr("id")] = true
		}
	})
	for name, z := range zones {
		if name == "" || len(z.ParagraphPath) == 0 {
			return fmt.Errorf("invalid zone name/path")
		}
		key := pathKey(z.ParagraphPath)
		if seen[key] {
			return fmt.Errorf("duplicate zone target %s", key)
		}
		seen[key] = true
		p, e := nodeAt(root, z.ParagraphPath)
		if e != nil || p.Name != "a:p" {
			return fmt.Errorf("zone %s must target existing a:p", name)
		}
		parent, e := nodeAt(root, z.ParagraphPath[:len(z.ParagraphPath)-1])
		if e != nil || parent.Name != "p:txBody" {
			return fmt.Errorf("zone %s must be in existing text body", name)
		}
		id := objectIDAt(root, z.ParagraphPath)
		if !objects[id] {
			return fmt.Errorf("zone %s object %s is not owned", name, id)
		}
		for _, c := range p.Children {
			if c.Name == "a:r" || c.Name == "a:fld" || c.Name == "a:br" || strings.TrimSpace(c.Text) != "" {
				return fmt.Errorf("zone %s is not empty", name)
			}
		}
		if z.StyleDonorBindingID == "" && z.Style == nil {
			return fmt.Errorf("zone %s needs donor or explicit style", name)
		}
		if z.StyleDonorBindingID != "" {
			i, ok := idx[z.StyleDonorBindingID]
			if !ok || all[i].Property != "text" {
				return fmt.Errorf("zone %s donor must be same-slide text binding", name)
			}
			if !objects[all[i].ObjectID] {
				return fmt.Errorf("zone %s donor is not owned", name)
			}
			if _, _, e := paragraphAndRun(root, all[i]); e != nil {
				return e
			}
		}
		if z.Mode != "" && z.Mode != "overlay_textbox" {
			return fmt.Errorf("zone %s has unsupported mode", name)
		}
		if z.Mode == "overlay_textbox" {
			if z.StyleDonorBindingID == "" || z.OverlayObjectID == "" || existingIDs[z.OverlayObjectID] || newIDs[z.OverlayObjectID] || len(z.FrameEMU) != 4 {
				return fmt.Errorf("zone %s needs unique overlay ID, donor, and four-value frame", name)
			}
			newIDs[z.OverlayObjectID] = true
			if _, e := strconv.ParseUint(z.OverlayObjectID, 10, 32); e != nil {
				return fmt.Errorf("zone %s overlay ID must be numeric", name)
			}
			shape, e := enclosingShape(root, z.ParagraphPath)
			if e != nil {
				return e
			}
			bounds, e := shapeFrame(shape, values)
			if e != nil {
				return e
			}
			f := z.FrameEMU
			if f[2] <= 0 || f[3] <= 0 || f[0] < bounds[0] || f[1] < bounds[1] || f[0]+f[2] > bounds[0]+bounds[2] || f[1]+f[3] > bounds[1]+bounds[3] {
				return fmt.Errorf("zone %s overlay frame is outside target bounds", name)
			}
			donor := all[idx[z.StyleDonorBindingID]]
			ds, e := enclosingShape(root, donor.NodePath)
			if e != nil || ds.Child("txBody") == nil {
				return fmt.Errorf("zone %s donor is not a native text shape", name)
			}
		}
		if z.Style != nil {
			if z.Style.FontSize < 0 || z.Style.FontSize > 9600 || z.Style.Typeface != "" && strings.TrimSpace(z.Style.Typeface) == "" || z.Style.ColorRGB != "" && !hexColor.MatchString(z.Style.ColorRGB) {
				return fmt.Errorf("zone %s invalid style", name)
			}
			if z.StyleDonorBindingID == "" && (z.Style.FontSize == 0 || z.Style.Typeface == "" || z.Style.ColorRGB == "") {
				return fmt.Errorf("zone %s explicit style requires size, typeface, RGB color", name)
			}
		}
	}
	return nil
}
func materialize(n *nativepkg.Node, byID map[string]string) error {
	if n == nil {
		return nil
	}
	if strings.HasPrefix(n.Text, "__BINDING:") {
		v, ok := byID[strings.TrimPrefix(n.Text, "__BINDING:")]
		if !ok {
			return fmt.Errorf("unknown style sentinel")
		}
		n.Text = v
	}
	for i := range n.Attrs {
		if strings.HasPrefix(n.Attrs[i].Value, "__BINDING:") {
			v, ok := byID[strings.TrimPrefix(n.Attrs[i].Value, "__BINDING:")]
			if !ok {
				return fmt.Errorf("unknown style sentinel")
			}
			n.Attrs[i].Value = v
		}
	}
	for _, c := range n.Children {
		if e := materialize(c, byID); e != nil {
			return e
		}
	}
	return nil
}
func applyZones(values map[string]string, zones map[string]Zone, root *nativepkg.Node, all []nativepkg.Binding, idx map[string]int) ([]map[string]any, error) {
	byID := map[string]string{}
	for _, b := range all {
		byID[b.BindingID] = b.Value
	}
	var changes []map[string]any
	for name, value := range values {
		z, ok := zones[name]
		if !ok {
			return nil, fmt.Errorf("unknown zone %q", name)
		}
		if e := validateText(value, "zone "+name); e != nil {
			return nil, e
		}
		p, _ := nodeAt(root, z.ParagraphPath)
		var pr *nativepkg.Node
		if z.StyleDonorBindingID != "" {
			_, rp, e := paragraphAndRun(root, all[idx[z.StyleDonorBindingID]])
			if e != nil {
				return nil, e
			}
			r, _ := nodeAt(root, rp)
			pr = cloneNode(r.Child("rPr"))
			if pr == nil {
				return nil, fmt.Errorf("zone %s donor has no run style", name)
			}
			if e := materialize(pr, byID); e != nil {
				return nil, e
			}
		}
		if pr == nil {
			pr = &nativepkg.Node{Name: "a:rPr", Attrs: []nativepkg.Attr{{Name: "lang", Value: "en-US"}}}
		}
		if z.Style != nil {
			st := z.Style
			if st.FontSize > 0 {
				pr.SetAttr("sz", strconv.Itoa(st.FontSize))
			}
			if st.Bold {
				pr.SetAttr("b", "1")
			}
			if st.Typeface != "" {
				latin := pr.Child("latin")
				if latin == nil {
					latin = &nativepkg.Node{Name: "a:latin"}
					at := len(pr.Children)
					for i, c := range pr.Children {
						if c.Name == "a:ea" || c.Name == "a:cs" || c.Name == "a:hlinkClick" {
							at = i
							break
						}
					}
					insertChild(pr, at, latin)
				}
				latin.SetAttr("typeface", st.Typeface)
			}
			if st.ColorRGB != "" {
				fill := &nativepkg.Node{Name: "a:solidFill", Children: []*nativepkg.Node{{Name: "a:srgbClr", Attrs: []nativepkg.Attr{{Name: "val", Value: strings.ToUpper(st.ColorRGB)}}}}}
				found := false
				for i, c := range pr.Children {
					if c.Name == "a:solidFill" {
						pr.Children[i] = fill
						found = true
						break
					}
				}
				if !found {
					at := len(pr.Children)
					for i, c := range pr.Children {
						if c.Name == "a:effectLst" || c.Name == "a:effectDag" || c.Name == "a:latin" || c.Name == "a:ea" || c.Name == "a:cs" {
							at = i
							break
						}
					}
					insertChild(pr, at, fill)
				}
			}
		}
		run := &nativepkg.Node{Name: "a:r", Children: []*nativepkg.Node{pr, {Name: "a:t", Children: []*nativepkg.Node{{Text: value}}}}}
		if z.Mode == "overlay_textbox" {
			donor, e := enclosingShape(root, all[idx[z.StyleDonorBindingID]].NodePath)
			if e != nil {
				return nil, e
			}
			overlay := cloneNode(donor)
			if e := materialize(overlay, byID); e != nil {
				return nil, e
			}
			nv := overlay.Child("nvSpPr")
			if nv == nil || nv.Child("cNvPr") == nil {
				return nil, fmt.Errorf("zone %s donor has no shape identity", name)
			}
			cn := nv.Child("cNvPr")
			cn.SetAttr("id", z.OverlayObjectID)
			cn.SetAttr("name", "Component zone "+name)
			var kept []*nativepkg.Node
			for _, c := range cn.Children {
				if c.Name != "a:extLst" {
					kept = append(kept, c)
				}
			}
			cn.Children = kept
			xf := overlay.Child("spPr").Child("xfrm")
			if xf == nil || xf.Child("off") == nil || xf.Child("ext") == nil {
				return nil, fmt.Errorf("zone %s donor frame missing", name)
			}
			var attrs []nativepkg.Attr
			for _, a := range xf.Attrs {
				if a.Name != "rot" && a.Name != "flipH" && a.Name != "flipV" {
					attrs = append(attrs, a)
				}
			}
			xf.Attrs = attrs
			xf.Child("off").SetAttr("x", strconv.FormatInt(z.FrameEMU[0], 10))
			xf.Child("off").SetAttr("y", strconv.FormatInt(z.FrameEMU[1], 10))
			xf.Child("ext").SetAttr("cx", strconv.FormatInt(z.FrameEMU[2], 10))
			xf.Child("ext").SetAttr("cy", strconv.FormatInt(z.FrameEMU[3], 10))
			body := overlay.Child("txBody")
			if body == nil {
				return nil, fmt.Errorf("zone %s donor text body missing", name)
			}
			var bodyChildren []*nativepkg.Node
			for _, c := range body.Children {
				if c.Name != "a:p" {
					bodyChildren = append(bodyChildren, c)
					continue
				}
				pnew := &nativepkg.Node{Name: "a:p"}
				for _, pc := range c.Children {
					if pc.Name == "a:pPr" {
						pnew.Children = append(pnew.Children, cloneNode(pc))
						break
					}
				}
				pnew.Children = append(pnew.Children, run)
				for _, pc := range c.Children {
					if pc.Name == "a:endParaRPr" {
						pnew.Children = append(pnew.Children, cloneNode(pc))
						break
					}
				}
				bodyChildren = append(bodyChildren, pnew)
			}
			body.Children = bodyChildren
			cs := root.Child("cSld")
			if cs == nil || cs.Child("spTree") == nil {
				return nil, fmt.Errorf("zone %s slide tree missing", name)
			}
			cs.Child("spTree").Children = append(cs.Child("spTree").Children, overlay)
			changes = append(changes, map[string]any{"zone": name, "mode": z.Mode, "target_paragraph_path": z.ParagraphPath, "overlay_object_id": z.OverlayObjectID, "frame_emu": z.FrameEMU, "text": value, "style_donor_binding_id": z.StyleDonorBindingID})
			continue
		}
		insert := len(p.Children)
		for i, c := range p.Children {
			if c.Name == "a:endParaRPr" {
				insert = i
				break
			}
		}
		// Inserting before endParaRPr keeps valid DrawingML order. Keep pinned
		// binding IDs, but readdress existing descendants shifted by this run.
		for i := range all {
			path := all[i].NodePath
			if len(path) > len(z.ParagraphPath) {
				match := true
				for j, v := range z.ParagraphPath {
					if path[j] != v {
						match = false
						break
					}
				}
				if match && path[len(z.ParagraphPath)] >= insert {
					path[len(z.ParagraphPath)]++
				}
			}
		}
		p.Children = append(p.Children, nil)
		copy(p.Children[insert+1:], p.Children[insert:])
		p.Children[insert] = run
		changes = append(changes, map[string]any{"zone": name, "paragraph_path": z.ParagraphPath, "text": value, "style_donor_binding_id": z.StyleDonorBindingID})
	}
	return changes, nil
}
