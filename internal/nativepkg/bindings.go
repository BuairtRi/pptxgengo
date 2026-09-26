package nativepkg

import (
	"fmt"
	"strings"
)

// Binding identifies one editable native OOXML value. NodePath addresses the
// element (or character-data node when Attribute is empty) from the scene root.
// BindingID is deterministic for a given scene tree; ObjectID and ObjectName
// are source trace data, not persistent authoring identities.
type Binding struct {
	BindingID  string `json:"binding_id"`
	ObjectID   string `json:"object_id,omitempty"`
	ObjectName string `json:"object_name,omitempty"`
	Property   string `json:"property"`
	Units      string `json:"units"`
	NodePath   []int  `json:"node_path"`
	Attribute  string `json:"attribute,omitempty"`
	Value      string `json:"value"`
}

const bindingSentinelPrefix = "__BINDING:"

type bindingObject struct {
	id   string
	name string
}

// ExtractBindings replaces supported editable values in scene with sentinels
// and returns their typed binding records. It intentionally leaves extension
// payloads, unsupported metadata, and asset bytes untouched.
func ExtractBindings(scene *Node) []Binding {
	if scene == nil {
		return nil
	}
	var out []Binding
	seen := map[string]bool{}
	var walk func(*Node, []int, []string, bindingObject)
	walk = func(n *Node, path []int, ancestors []string, object bindingObject) {
		if n == nil {
			return
		}
		name := local(n.Name)
		if isVisualObject(name) {
			object = objectFor(n)
		}

		if n.Name == "" {
			if len(ancestors) > 0 && ancestors[len(ancestors)-1] == "t" {
				addTextBinding(&out, seen, n, path, object)
			}
			return
		}

		for i := range n.Attrs {
			attr := n.Attrs[i].Name
			property, ok := bindingProperty(name, attr, ancestors)
			if !ok {
				continue
			}
			addAttributeBinding(&out, seen, n, path, object, property, attr, i)
		}

		nextAncestors := append(append([]string(nil), ancestors...), name)
		for i, child := range n.Children {
			walk(child, appendPath(path, i), nextAncestors, object)
		}
	}
	walk(scene, nil, nil, bindingObject{})
	return out
}

// ApplyBindings restores binding values into a scene that ExtractBindings has
// prepared. It rejects duplicate, unknown, missing, or already-applied slots
// so a raw-XML mutation cannot silently masquerade as a binding-driven build.
func ApplyBindings(scene *Node, bindings []Binding) error {
	if scene == nil {
		return fmt.Errorf("nil scene")
	}
	placeholders := sentinelSlots(scene)
	byID := make(map[string]Binding, len(bindings))
	for _, b := range bindings {
		if b.BindingID == "" {
			return fmt.Errorf("binding without binding_id")
		}
		if _, exists := byID[b.BindingID]; exists {
			return fmt.Errorf("duplicate binding %q", b.BindingID)
		}
		if strings.HasPrefix(b.Value, bindingSentinelPrefix) {
			return fmt.Errorf("binding %q value is reserved sentinel text", b.BindingID)
		}
		byID[b.BindingID] = b
	}
	for id := range placeholders {
		if _, ok := byID[id]; !ok {
			return fmt.Errorf("missing binding %q", id)
		}
	}
	for id, b := range byID {
		slot, ok := placeholders[id]
		if !ok {
			return fmt.Errorf("unknown binding %q", id)
		}
		n, err := nodeAt(scene, b.NodePath)
		if err != nil {
			return fmt.Errorf("binding %q: %w", id, err)
		}
		if b.Attribute == "" {
			if n.Name != "" || n.Text != slot.sentinel {
				return fmt.Errorf("binding %q does not address its text sentinel", id)
			}
		} else if n.Attr(b.Attribute) != slot.sentinel {
			return fmt.Errorf("binding %q does not address its attribute sentinel", id)
		}
	}
	for id, slot := range placeholders {
		b := byID[id]
		if slot.attribute == "" {
			slot.node.Text = b.Value
		} else {
			slot.node.SetAttr(slot.attribute, b.Value)
		}
	}
	if leftovers := sentinelSlots(scene); len(leftovers) != 0 {
		var ids []string
		for id := range leftovers {
			ids = append(ids, id)
		}
		return fmt.Errorf("binding sentinels remain: %s", strings.Join(ids, ", "))
	}
	return nil
}

type sentinelSlot struct {
	node      *Node
	attribute string
	sentinel  string
}

func sentinelSlots(scene *Node) map[string]sentinelSlot {
	out := map[string]sentinelSlot{}
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		if n.Name == "" && strings.HasPrefix(n.Text, bindingSentinelPrefix) {
			out[strings.TrimPrefix(n.Text, bindingSentinelPrefix)] = sentinelSlot{node: n, sentinel: n.Text}
		}
		for _, a := range n.Attrs {
			if strings.HasPrefix(a.Value, bindingSentinelPrefix) {
				out[strings.TrimPrefix(a.Value, bindingSentinelPrefix)] = sentinelSlot{node: n, attribute: a.Name, sentinel: a.Value}
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(scene)
	return out
}

func addTextBinding(out *[]Binding, seen map[string]bool, n *Node, path []int, object bindingObject) {
	property := "text"
	id := bindingID(object, property, path, "text")
	if seen[id] {
		return
	}
	seen[id] = true
	value := n.Text
	n.Text = bindingSentinelPrefix + id
	*out = append(*out, Binding{
		BindingID: id, ObjectID: object.id, ObjectName: object.name,
		Property: property, Units: "UTF-8", NodePath: copyPath(path), Value: value,
	})
}

func addAttributeBinding(out *[]Binding, seen map[string]bool, n *Node, path []int, object bindingObject, property, attr string, index int) {
	id := bindingID(object, property, path, attr)
	if seen[id] {
		return
	}
	seen[id] = true
	value := n.Attrs[index].Value
	n.Attrs[index].Value = bindingSentinelPrefix + id
	*out = append(*out, Binding{
		BindingID: id, ObjectID: object.id, ObjectName: object.name,
		Property: property, Units: bindingUnits(property, attr), NodePath: copyPath(path), Attribute: attr, Value: value,
	})
}

func bindingProperty(name, attr string, ancestors []string) (string, bool) {
	if name == "xfrm" {
		switch attr {
		case "rot", "flipH", "flipV":
			return "transform." + attr, true
		}
	}
	if isTransformChild(name) && hasAncestor(ancestors, "xfrm") {
		switch attr {
		case "x", "y", "cx", "cy":
			return "transform." + name + "." + attr, true
		}
	}
	if (name == "srgbClr" || name == "schemeClr") && attr == "val" {
		return "fill." + name + ".val", true
	}
	if (name == "alpha" || name == "tint") && attr == "val" {
		return "fill." + name + ".val", true
	}
	if name == "ln" && attr == "w" {
		return "line.width", true
	}
	if name == "prstDash" && attr == "val" {
		return "line.dash", true
	}
	if name == "prstGeom" && attr == "prst" {
		return "geometry.preset", true
	}
	if name == "srcRect" {
		switch attr {
		case "l", "t", "r", "b":
			return "media_crop." + attr, true
		}
	}
	if name == "gridCol" && attr == "w" {
		return "table.grid.width", true
	}
	if name == "tr" && attr == "h" {
		return "table.row.height", true
	}
	if name == "tcPr" {
		switch attr {
		case "marL", "marR", "marT", "marB":
			return "table.cell_margin." + attr, true
		}
	}
	if root := styleRoot(name, ancestors); root != "" && !strings.HasPrefix(attr, "xmlns") {
		return "style." + root + "." + name + "." + attr, true
	}
	return "", false
}

func styleRoot(name string, ancestors []string) string {
	if isStyleRoot(name) {
		return name
	}
	for i := len(ancestors) - 1; i >= 0; i-- {
		if isStyleRoot(ancestors[i]) {
			return ancestors[i]
		}
	}
	return ""
}

func isStyleRoot(name string) bool {
	switch name {
	case "rPr", "defRPr", "endParaRPr", "pPr", "bodyPr":
		return true
	}
	return false
}

func isVisualObject(name string) bool {
	switch name {
	case "sp", "pic", "graphicFrame", "cxnSp", "grpSp":
		return true
	}
	return false
}

func isTransformChild(name string) bool {
	switch name {
	case "off", "ext", "chOff", "chExt":
		return true
	}
	return false
}

func hasAncestor(ancestors []string, name string) bool {
	for i := len(ancestors) - 1; i >= 0; i-- {
		if ancestors[i] == name {
			return true
		}
	}
	return false
}

func objectFor(n *Node) bindingObject {
	var find func(*Node) *Node
	find = func(current *Node) *Node {
		for _, child := range current.Children {
			if child == nil {
				continue
			}
			if local(child.Name) == "cNvPr" {
				return child
			}
			if isVisualObject(local(child.Name)) {
				continue
			}
			if found := find(child); found != nil {
				return found
			}
		}
		return nil
	}
	if c := find(n); c != nil {
		return bindingObject{id: c.Attr("id"), name: c.Attr("name")}
	}
	return bindingObject{}
}

func bindingID(object bindingObject, property string, path []int, attribute string) string {
	objectID := object.id
	if objectID == "" {
		objectID = "root"
	}
	return "b/" + objectID + "/" + property + "/" + pathString(path) + "/" + attribute
}

func bindingUnits(property, attr string) string {
	if property == "text" {
		return "UTF-8"
	}
	switch attr {
	case "x", "y", "cx", "cy", "w", "h", "marL", "marR", "marT", "marB", "indent":
		return "EMU"
	case "rot":
		return "1/60000 degree"
	case "flipH", "flipV":
		return "boolean"
	case "sz":
		return "centipoint"
	case "baseline":
		return "1/1000 em"
	case "alpha", "tint":
		return "1/1000 percent"
	default:
		return "string/enum"
	}
}

func appendPath(path []int, index int) []int {
	next := make([]int, len(path)+1)
	copy(next, path)
	next[len(path)] = index
	return next
}

func copyPath(path []int) []int {
	next := make([]int, len(path))
	copy(next, path)
	return next
}

func pathString(path []int) string {
	if len(path) == 0 {
		return "root"
	}
	parts := make([]string, len(path))
	for i, n := range path {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(parts, ".")
}

func nodeAt(scene *Node, path []int) (*Node, error) {
	n := scene
	for _, index := range path {
		if index < 0 || index >= len(n.Children) {
			return nil, fmt.Errorf("node path %s is outside scene", pathString(path))
		}
		n = n.Children[index]
		if n == nil {
			return nil, fmt.Errorf("node path %s contains nil node", pathString(path))
		}
	}
	return n, nil
}
