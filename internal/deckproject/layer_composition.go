package deckproject

import (
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"math"
	"sort"
	"strconv"
	"strings"
)

const LayerPatchSchema = "pptxgengo.layer-patch.v1"

// A layer can comprise one native row/plane/node or several explicitly mapped
// catalog objects. Membership is authored, never guessed from native proximity.
type LayerSelection struct {
	Key        string   `json:"key"`
	Members    []string `json:"members"`
	LabelNode  string   `json:"label_node"`
	TextNode   string   `json:"text_node,omitempty"`
	NumberNode string   `json:"number_node,omitempty"`
}
type LayerEntry struct {
	LayerSelection
	Label   string        `json:"label"`
	Text    string        `json:"text,omitempty"`
	Surface string        `json:"surface,omitempty"`
	Rect    wmdesign.Rect `json:"rect"`
}
type LayerLayout struct {
	Rect           wmdesign.Rect `json:"rect"`
	Gap            float64       `json:"gap_pt"`
	Overlap        float64       `json:"overlap_pt"`
	Palette        string        `json:"palette"`
	FoundationNode string        `json:"foundation_node,omitempty"`
	FoundationGap  float64       `json:"foundation_gap_pt,omitempty"`
	ControlsNode   string        `json:"controls_node,omitempty"`
	ControlsMode   string        `json:"controls_mode,omitempty"`
}
type LayerOperation struct {
	Action    string         `json:"action"`
	Entity    string         `json:"entity"`
	Key       string         `json:"key,omitempty"`
	Prototype string         `json:"prototype,omitempty"`
	Label     *string        `json:"label,omitempty"`
	Text      *string        `json:"text,omitempty"`
	Surface   *string        `json:"surface,omitempty"`
	Order     []string       `json:"order,omitempty"`
	Layout    *LayerLayout   `json:"layout,omitempty"`
	Cascade   bool           `json:"cascade,omitempty"`
	NodeID    string         `json:"node_id,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}
type LayerPatch struct {
	Schema               string           `json:"schema"`
	ExpectedSourceSHA256 string           `json:"expected_source_sha256"`
	Actor                string           `json:"actor"`
	Reason               string           `json:"reason"`
	Layers               []LayerSelection `json:"layers"`
	Operations           []LayerOperation `json:"operations"`
}
type LayerInspection struct {
	Schema          string            `json:"schema"`
	SlideID         string            `json:"slide_id"`
	SourceSHA256    string            `json:"source_sha256"`
	Layers          []LayerEntry      `json:"layers"`
	Layout          LayerLayout       `json:"layout"`
	MutationBlocked string            `json:"mutation_blocked,omitempty"`
	RenderError     string            `json:"render_error,omitempty"`
	Geometry        DiagramInspection `json:"geometry"`
}

func layerFields(a, b string) (string, error) {
	switch a + "/" + b {
	case "materialize/source":
		return "", nil
	case "add/layer":
		return "key prototype label text surface", nil
	case "set/layer":
		return "key label text surface", nil
	case "remove/layer":
		return "key cascade", nil
	case "reorder/layer":
		return "order", nil
	case "set/layout":
		return "layout", nil
	case "set/foundation", "set/controls":
		return "node_id arguments", nil
	}
	return "", fmt.Errorf("unsupported layer operation %s/%s", a, b)
}
func DecodeLayerPatch(raw []byte, file string) (LayerPatch, error) {
	var out LayerPatch
	e := decodeCurveDocument(raw, file, &out, LayerPatchSchema, layerFields)
	if e == nil && (len(out.Layers) < 1 || len(out.Layers) > 20) {
		e = fmt.Errorf("layer patch requires explicit 1..20 layer selections from inspect")
	}
	return out, e
}
func layerLeaf(t *LocalTemplate, id string) (*Node, error) {
	nodes, i := findDiagramNode(&t.Nodes, id)
	if nodes == nil {
		return nil, fmt.Errorf("unknown layer member %s", id)
	}
	n := &(*nodes)[i]
	if len(n.Nodes) > 0 || n.Placement == nil || n.Placement.Rect == nil {
		return nil, fmt.Errorf("layer member %s requires explicit leaf rect", id)
	}
	if n.Kind != "component" && n.Kind != "text" {
		return nil, fmt.Errorf("unsupported layer member kind %s", n.Kind)
	}
	if n.Kind == "component" && (n.Definition == nil || !strings.HasPrefix(n.Definition.ID, "wmds/component/")) {
		return nil, fmt.Errorf("layer requires typed design-system member")
	}
	return n, nil
}
func layerKind(n *Node) string {
	if n.Kind == "text" {
		return "text"
	}
	return strings.TrimPrefix(n.Definition.ID, "wmds/component/")
}
func layerText(n *Node, args map[string]any, role string) (string, error) {
	if n.Kind == "text" {
		s, ok := n.Text.(string)
		if !ok {
			return "", fmt.Errorf("bound text node requires materialize/source")
		}
		return s, nil
	}
	field := role
	switch layerKind(n) {
	case "node":
		if role == "label" {
			field = "text"
		} else if role == "text" {
			field = "sub"
		}
	case "block":
		field = "text"
	case "text":
		field = "text"
	}
	s, _ := args[field].(string)
	return s, nil
}
func layerWrite(n *Node, role, value string) error {
	if n.Kind == "text" {
		n.Text = value
		return nil
	}
	field := role
	switch layerKind(n) {
	case "layerrow", "plane":
		if role == "number" {
			field = "n"
		}
	case "node":
		if role == "label" {
			field = "text"
		} else if role == "text" {
			field = "sub"
		} else {
			return fmt.Errorf("node has no numbering")
		}
	case "block", "text":
		field = "text"
	default:
		return fmt.Errorf("unsupported layer label member %s; explicitly map row/plane/node/text/block", layerKind(n))
	}
	if n.Arguments == nil {
		n.Arguments = map[string]any{}
	}
	n.Arguments[field] = value
	return nil
}
func layerRead(t *LocalTemplate, selections []LayerSelection, values map[string]any) ([]LayerEntry, bool, error) {
	if len(selections) < 1 || len(selections) > 20 {
		return nil, false, fmt.Errorf("explicit layer selections required")
	}
	out := []LayerEntry{}
	seen := map[string]bool{}
	bound := false
	for _, s := range selections {
		if !stableID.MatchString(s.Key) || seen[s.Key] || len(s.Members) < 1 || len(s.Members) > 20 {
			return nil, bound, fmt.Errorf("invalid layer key/member count")
		}
		seen[s.Key] = true
		members := map[string]bool{}
		rect := wmdesign.Rect{}
		for i, id := range s.Members {
			if members[id] || seen["node:"+id] {
				return nil, bound, fmt.Errorf("layer member belongs to more than one layer")
			}
			members[id] = true
			seen["node:"+id] = true
			n, e := layerLeaf(t, id)
			if e != nil {
				return nil, bound, e
			}
			if i == 0 {
				rect = *n.Placement.Rect
			} else {
				r := *n.Placement.Rect
				x, y := math.Min(rect.X, r.X), math.Min(rect.Y, r.Y)
				rect = wmdesign.Rect{X: x, Y: y, W: math.Max(rect.X+rect.W, r.X+r.W) - x, H: math.Max(rect.Y+rect.H, r.Y+r.H) - y}
			}
			collectBindings(n.Arguments, func(string) { bound = true })
			collectBindings(n.Text, func(string) { bound = true })
		}
		if !members[s.LabelNode] || s.TextNode != "" && !members[s.TextNode] || s.NumberNode != "" && !members[s.NumberNode] {
			return nil, bound, fmt.Errorf("layer label/text/number mappings must address its members")
		}
		entry := LayerEntry{LayerSelection: s, Rect: rect}
		for _, p := range []struct {
			id, role string
			dst      *string
		}{{s.LabelNode, "label", &entry.Label}, {s.TextNode, "text", &entry.Text}} {
			if p.id != "" {
				n, _ := layerLeaf(t, p.id)
				args, _, e := curveSource(n, values)
				if e != nil {
					return nil, bound, e
				}
				if n.Kind == "text" {
					v, e := resolveBinding(n.Text, values)
					if e != nil {
						return nil, bound, e
					}
					copy := *n
					copy.Text = v
					*p.dst, e = layerText(&copy, args, p.role)
					if e != nil {
						return nil, bound, e
					}
				} else {
					*p.dst, e = layerText(n, args, p.role)
					if e != nil {
						return nil, bound, e
					}
				}
			}
		}
		n, _ := layerLeaf(t, s.LabelNode)
		args, _, e := curveSource(n, values)
		if e != nil {
			return nil, bound, e
		}
		entry.Surface, _ = args["surface"].(string)
		if entry.Surface == "" {
			entry.Surface = n.Surface
		}
		out = append(out, entry)
	}
	return out, bound, nil
}

// Selection is deliberately explicit for ambiguous 3-D, system planes and
// multi-object layer-map rows. Numbered row order follows authored unique
// positive n values, never native movement or the incidental paint array.
// Ambiguity returns no automatic selection; InspectLayers supplies the error.
func DiscoverLayerSelections(t LocalTemplate, values ...map[string]any) []LayerSelection {
	var resolved map[string]any
	if len(values) > 0 {
		resolved = values[0]
	}
	out, _ := discoverNumberedLayerSelections(t, resolved)
	return out
}

func discoverNumberedLayerSelections(t LocalTemplate, values map[string]any) ([]LayerSelection, error) {
	type numbered struct {
		selection LayerSelection
		number    int
	}
	rows := []numbered{}
	seen := map[int]bool{}
	for _, n := range t.Nodes {
		if n.Kind != "component" || n.Definition == nil {
			continue
		}
		k := layerKind(&n)
		if k == "layerrow" && n.Arguments["n"] != nil {
			value, err := resolveBinding(n.Arguments["n"], values)
			if err != nil {
				return nil, fmt.Errorf("layer numbering for %s is unresolved: %w; supply reviewed explicit selections", n.ID, err)
			}
			text, ok := value.(string)
			number, err := strconv.Atoi(text)
			if !ok || err != nil || number <= 0 || seen[number] {
				return nil, fmt.Errorf("layer numbering is ambiguous at %s: automatic order requires unique positive integer n; supply reviewed explicit selections", n.ID)
			}
			seen[number] = true
			rows = append(rows, numbered{LayerSelection{n.ID, []string{n.ID}, n.ID, n.ID, n.ID}, number})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].number < rows[j].number })
	out := make([]LayerSelection, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.selection)
	}
	return out, nil
}
func InspectLayers(p *Project, slideID string, selections []LayerSelection, bundle, engine string) (LayerInspection, error) {
	out := LayerInspection{Schema: "pptxgengo.layer-inspection.v1", SlideID: slideID, SourceSHA256: p.SourceHash()}
	i, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	if len(selections) == 0 {
		selections, e = discoverNumberedLayerSelections(t, p.Document.Slides[i].Values)
		if e != nil {
			return out, e
		}
	}
	var bound bool
	out.Layers, bound, e = layerRead(&t, selections, p.Document.Slides[i].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Bound member facts require explicit first materialize/source."
	}
	r := out.Layers[0].Rect
	for _, v := range out.Layers[1:] {
		x, y := math.Min(r.X, v.Rect.X), math.Min(r.Y, v.Rect.Y)
		r = wmdesign.Rect{X: x, Y: y, W: math.Max(r.X+r.W, v.Rect.X+v.Rect.W) - x, H: math.Max(r.Y+r.H, v.Rect.Y+v.Rect.H) - y}
	}
	out.Layout = LayerLayout{Rect: r, Palette: "preserve"}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func layerEntryIndex(m []LayerEntry, key string) int {
	for i, v := range m {
		if v.Key == key {
			return i
		}
	}
	return -1
}
func layerMaterialize(t *LocalTemplate, m []LayerEntry, values map[string]any) error {
	for _, v := range m {
		for _, id := range v.Members {
			n, e := layerLeaf(t, id)
			if e != nil {
				return e
			}
			args, _, e := curveSource(n, values)
			if e != nil {
				return e
			}
			n.Arguments = args
			if n.Kind == "text" {
				n.Text, e = resolveBinding(n.Text, values)
				if e != nil {
					return e
				}
			}
		}
	}
	return nil
}
func layerArrange(t *LocalTemplate, m []LayerEntry, l LayerLayout) error {
	if len(m) < 1 || len(m) > 20 || !curveFinite(l.Gap) || !curveFinite(l.Overlap) || l.Gap < 0 || l.Overlap < 0 || l.Gap > 0 && l.Overlap > 0 {
		return fmt.Errorf("layer layout gap/overlap must be finite nonnegative and mutually exclusive")
	}
	r := l.Rect
	if !curveFinite(r.X) || !curveFinite(r.Y) || !curveFinite(r.W) || !curveFinite(r.H) || r.W <= 0 || r.H <= 0 {
		return fmt.Errorf("layer layout requires positive finite rect")
	}
	h := (r.H - float64(len(m)-1)*(l.Gap-l.Overlap)) / float64(len(m))
	if h <= l.Overlap || h <= 0 {
		return fmt.Errorf("layer allocation cannot fit positive row spacing")
	}
	if l.Palette != "preserve" && l.Palette != "sequence" {
		return fmt.Errorf("palette must be preserve or sequence")
	}
	for i, v := range m {
		if v.Rect.W <= 0 || v.Rect.H <= 0 {
			return fmt.Errorf("invalid original layer allocation")
		}
		dest := wmdesign.Rect{X: r.X, Y: r.Y + float64(i)*(h+l.Gap-l.Overlap), W: r.W, H: h}
		for _, id := range v.Members {
			n, e := layerLeaf(t, id)
			if e != nil {
				return e
			}
			old := *n.Placement.Rect
			n.Placement.Rect = &wmdesign.Rect{X: dest.X + (old.X-v.Rect.X)*dest.W/v.Rect.W, Y: dest.Y + (old.Y-v.Rect.Y)*dest.H/v.Rect.H, W: old.W * dest.W / v.Rect.W, H: old.H * dest.H / v.Rect.H}
			if l.Palette == "sequence" && n.Kind == "component" && (layerKind(n) == "layerrow" || layerKind(n) == "plane" || layerKind(n) == "node" || layerKind(n) == "block") {
				surface := []string{"inverse", "strong", "subtle", "outline"}[int(math.Round(float64(i)*3/math.Max(1, float64(len(m)-1))))]
				n.Arguments["surface"] = surface
			}
		}
		if v.NumberNode != "" {
			n, _ := layerLeaf(t, v.NumberNode)
			if e := layerWrite(n, "number", fmt.Sprint(i+1)); e != nil {
				return e
			}
		}
	}
	allPlanes := true
	var siblingList *[]Node
	selected := map[string]Node{}
	for _, v := range m {
		if len(v.Members) != 1 {
			allPlanes = false
			break
		}
		list, i := findDiagramNode(&t.Nodes, v.Members[0])
		n := (*list)[i]
		if layerKind(&n) != "plane" {
			allPlanes = false
			break
		}
		if siblingList == nil {
			siblingList = list
		} else if siblingList != list {
			return fmt.Errorf("plane stack requires a common authored parent")
		}
		selected[n.ID] = n
	}
	if allPlanes {
		// Paint the bottom plane first so foreground top edges stay visible.
		ordered := []Node{}
		for i := len(m) - 1; i >= 0; i-- {
			ordered = append(ordered, selected[m[i].Members[0]])
		}
		first := -1
		kept := []Node{}
		for _, n := range *siblingList {
			if _, ok := selected[n.ID]; ok {
				if first < 0 {
					first = len(kept)
				}
				continue
			}
			kept = append(kept, n)
		}
		if first < 0 {
			return fmt.Errorf("missing plane membership")
		}
		*siblingList = append(append(append([]Node{}, kept[:first]...), ordered...), kept[first:]...)
	}
	if l.FoundationNode != "" {
		n, e := layerLeaf(t, l.FoundationNode)
		if e != nil {
			return e
		}
		if !curveFinite(l.FoundationGap) || l.FoundationGap < 0 {
			return fmt.Errorf("foundation gap must be finite nonnegative")
		}
		copy := *n.Placement.Rect
		copy.Y = r.Y + r.H + l.FoundationGap
		n.Placement.Rect = &copy
	}
	if l.ControlsNode != "" {
		n, e := layerLeaf(t, l.ControlsNode)
		if e != nil {
			return e
		}
		if layerKind(n) != "card" {
			return fmt.Errorf("controls must address typed card")
		}
		if l.ControlsMode != "span" && l.ControlsMode != "label_only" {
			return fmt.Errorf("controls_mode must be span or label_only")
		}
		n.Arguments["label"] = fmt.Sprintf("Across layers 1–%d", len(m))
		if l.ControlsMode == "span" {
			copy := *n.Placement.Rect
			copy.Y = r.Y
			copy.H = r.H
			n.Placement.Rect = &copy
		}
	}
	return nil
}
func PatchLayers(p *Project, slideID string, patch LayerPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeLayerPatch(canonical(patch), "layer-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("layer source hash mismatch; inspect again")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	m, bound, e := layerRead(&clone, patch.Layers, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	var layout *LayerLayout
	materialized := false
	for i, op := range patch.Operations {
		if op.Action == "materialize" && op.Entity == "source" {
			if i != 0 {
				return empty, fmt.Errorf("materialize source must be first")
			}
			if e = layerMaterialize(&clone, m, p.Document.Slides[idx].Values); e != nil {
				return empty, e
			}
			bound = false
			materialized = true
			continue
		}
		if bound {
			return empty, fmt.Errorf("layer arguments require explicit materialize source")
		}
		j := layerEntryIndex(m, op.Key)
		switch op.Action + "/" + op.Entity {
		case "add/layer":
			if !stableID.MatchString(op.Key) || j >= 0 || op.Label == nil {
				return empty, fmt.Errorf("add requires new stable key, prototype and label")
			}
			pi := layerEntryIndex(m, op.Prototype)
			if pi < 0 {
				return empty, fmt.Errorf("unknown layer prototype")
			}
			v := m[pi]
			v.Key = op.Key
			oldIDs := v.Members
			v.Members = nil
			mapping := map[string]string{}
			for k, id := range oldIDs {
				n, _ := layerLeaf(&clone, id)
				var copy Node
				if e = strictInto(n, &copy); e != nil {
					return empty, e
				}
				copy.ID = fmt.Sprintf("%s-part-%02d", op.Key, k+1)
				if len(oldIDs) == 1 {
					copy.ID = op.Key
				}
				if a, _ := findDiagramNode(&clone.Nodes, copy.ID); a != nil {
					return empty, fmt.Errorf("new layer node ID collision")
				}
				mapping[id] = copy.ID
				v.Members = append(v.Members, copy.ID)
				siblings, _ := findDiagramNode(&clone.Nodes, id)
				*siblings = append(*siblings, copy)
			}
			v.LabelNode = mapping[v.LabelNode]
			v.TextNode = mapping[v.TextNode]
			v.NumberNode = mapping[v.NumberNode]
			m = append(m, v)
			j = len(m) - 1
			fallthrough
		case "set/layer":
			if j < 0 {
				return empty, fmt.Errorf("unknown layer")
			}
			v := &m[j]
			if op.Label != nil {
				if strings.TrimSpace(*op.Label) == "" {
					return empty, fmt.Errorf("layer label cannot be empty")
				}
				n, _ := layerLeaf(&clone, v.LabelNode)
				if e = layerWrite(n, "label", *op.Label); e != nil {
					return empty, e
				}
				v.Label = *op.Label
			}
			if op.Text != nil {
				if v.TextNode == "" {
					return empty, fmt.Errorf("layer has no text mapping")
				}
				n, _ := layerLeaf(&clone, v.TextNode)
				if e = layerWrite(n, "text", *op.Text); e != nil {
					return empty, e
				}
				v.Text = *op.Text
			}
			if op.Surface != nil {
				for _, id := range v.Members {
					n, _ := layerLeaf(&clone, id)
					if n.Kind == "component" && (layerKind(n) == "layerrow" || layerKind(n) == "plane" || layerKind(n) == "node" || layerKind(n) == "block") {
						n.Arguments["surface"] = *op.Surface
					}
				}
				v.Surface = *op.Surface
			}
		case "remove/layer":
			if j < 0 {
				return empty, fmt.Errorf("unknown layer")
			}
			for _, id := range m[j].Members {
				if e = removeIncident(&clone.Nodes, id, op.Cascade); e != nil {
					return empty, e
				}
				nodes, k := findDiagramNode(&clone.Nodes, id)
				*nodes = append((*nodes)[:k], (*nodes)[k+1:]...)
			}
			m = append(m[:j], m[j+1:]...)
		case "reorder/layer":
			keys := []string{}
			for _, v := range m {
				keys = append(keys, v.Key)
			}
			ii, e := curveReorder(keys, op.Order)
			if e != nil {
				return empty, e
			}
			old := m
			m = nil
			for _, k := range ii {
				m = append(m, old[k])
			}
		case "set/layout":
			if op.Layout == nil {
				return empty, fmt.Errorf("layout required")
			}
			layout = op.Layout
		case "set/foundation", "set/controls":
			n, e := layerLeaf(&clone, op.NodeID)
			if e != nil {
				return empty, e
			}
			if op.Entity == "controls" && layerKind(n) != "card" || op.Entity == "foundation" && layerKind(n) != "layerrow" && layerKind(n) != "node" && layerKind(n) != "plane" {
				return empty, fmt.Errorf("invalid foundation/control target")
			}
			args, has, e := curveSource(n, p.Document.Slides[idx].Values)
			if e != nil {
				return empty, e
			}
			if has && !materialized {
				return empty, fmt.Errorf("foundation/control bound facts require first materialize/source")
			}
			if len(op.Arguments) == 0 {
				return empty, fmt.Errorf("set foundation/controls requires complete arguments")
			}
			_ = args
			n.Arguments = op.Arguments
		default:
			return empty, fmt.Errorf("unsupported layer operation")
		}
	}
	if layout == nil {
		return empty, fmt.Errorf("layer patch requires explicit final set/layout so count and control span are reviewable")
	}
	for _, v := range m {
		for _, id := range v.Members {
			if id == layout.FoundationNode || id == layout.ControlsNode {
				return empty, fmt.Errorf("foundation/control target cannot also be a selected layer member")
			}
		}
	}
	if layout.ControlsNode != "" {
		n, e := layerLeaf(&clone, layout.ControlsNode)
		if e != nil {
			return empty, e
		}
		args, b, e := curveSource(n, p.Document.Slides[idx].Values)
		if e != nil {
			return empty, e
		}
		if b && !materialized {
			return empty, fmt.Errorf("controls label update requires explicit materialize/source")
		}
		if b {
			n.Arguments = args
		}
	}
	if e = layerArrange(&clone, m, *layout); e != nil {
		return empty, e
	}
	return CompositionCandidate(p, slideID, "layer", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
