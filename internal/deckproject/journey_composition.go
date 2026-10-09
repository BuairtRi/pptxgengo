package deckproject

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
	"strings"
)

const JourneyPatchSchema = "pptxgengo.journey-patch.v1"

type JourneyMilestone struct {
	Key       string          `json:"key"`
	Label     string          `json:"label,omitempty"`
	Date      string          `json:"date,omitempty"`
	Text      string          `json:"text,omitempty"`
	Number    json.RawMessage `json:"n,omitempty"`
	At        *float64        `json:"at,omitempty"`
	Side      string          `json:"side,omitempty"`
	HereLabel string          `json:"hereLabel,omitempty"`
}
type JourneyBranch struct {
	Key        string             `json:"key"`
	Title      string             `json:"title"`
	Text       string             `json:"text,omitempty"`
	Milestones []JourneyMilestone `json:"milestones"`
}
type JourneyFork struct {
	Label string   `json:"label,omitempty"`
	Text  string   `json:"text,omitempty"`
	Tag   string   `json:"tag,omitempty"`
	At    *float64 `json:"at,omitempty"`
}
type JourneyLayout struct {
	Rect      *wmdesign.Rect `json:"rect,omitempty"`
	Direction string         `json:"direction,omitempty"`
	RoadW     *float64       `json:"roadW,omitempty"`
	Amp       *float64       `json:"amp,omitempty"`
	Waves     *float64       `json:"waves,omitempty"`
	LabelW    *float64       `json:"labelW,omitempty"`
	ForkW     *float64       `json:"forkW,omitempty"`
	Small     *bool          `json:"small,omitempty"`
}
type JourneyModel struct {
	Kind     string             `json:"kind"`
	Mode     string             `json:"mode,omitempty"`
	Trunk    []JourneyMilestone `json:"trunk"`
	Branches []JourneyBranch    `json:"branches,omitempty"`
	Fork     *JourneyFork       `json:"fork,omitempty"`
	Chosen   string             `json:"chosen,omitempty"`
	Current  string             `json:"current,omitempty"`
	Layout   JourneyLayout      `json:"layout"`
}
type JourneyOperation struct {
	Action      string            `json:"action"`
	Entity      string            `json:"entity"`
	Key         string            `json:"key,omitempty"`
	Branch      string            `json:"branch,omitempty"`
	Order       []string          `json:"order,omitempty"`
	Cascade     bool              `json:"cascade,omitempty"`
	Milestone   *JourneyMilestone `json:"milestone,omitempty"`
	BranchValue *JourneyBranch    `json:"branch_value,omitempty"`
	Fork        *JourneyFork      `json:"fork,omitempty"`
	Layout      *JourneyLayout    `json:"layout,omitempty"`
}
type JourneyPatch struct {
	Schema               string             `json:"schema"`
	ExpectedSourceSHA256 string             `json:"expected_source_sha256"`
	Actor                string             `json:"actor"`
	Reason               string             `json:"reason"`
	NodeID               string             `json:"node_id"`
	Operations           []JourneyOperation `json:"operations"`
}
type JourneyInspection struct {
	Schema          string            `json:"schema"`
	SlideID         string            `json:"slide_id"`
	NodeID          string            `json:"node_id"`
	SourceSHA256    string            `json:"source_sha256"`
	Model           JourneyModel      `json:"model"`
	MutationBlocked string            `json:"mutation_blocked,omitempty"`
	RenderError     string            `json:"render_error,omitempty"`
	Geometry        DiagramInspection `json:"geometry"`
}

func journeyFields(a, e string) (map[string]bool, error) {
	s := "action entity"
	switch a + "/" + e {
	case "materialize/source", "remove/current", "remove/chosen":
	case "set/milestone":
		s += " key branch milestone"
	case "remove/milestone":
		s += " key cascade"
	case "set/branch":
		s += " key branch_value"
	case "remove/branch":
		s += " key cascade"
	case "reorder/milestone":
		s += " branch order"
	case "reorder/branch":
		s += " order"
	case "set/current", "set/chosen":
		s += " key"
	case "set/fork":
		s += " fork"
	case "set/layout":
		s += " layout"
	default:
		return nil, fmt.Errorf("unsupported journey %s/%s", a, e)
	}
	out := map[string]bool{}
	for _, k := range strings.Fields(s) {
		out[k] = true
	}
	return out, nil
}
func DecodeJourneyPatch(raw []byte, file string) (JourneyPatch, error) {
	var out JourneyPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("journey patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty journey patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("one journey patch document required")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = p.shapeType(v, reflect.TypeOf(out), ""); e != nil {
		return out, e
	}
	if e = strictInto(v, &out); e != nil {
		return out, e
	}
	ops, _ := v.(map[string]any)["operations"].([]any)
	for i, op := range out.Operations {
		fields, e := journeyFields(op.Action, op.Entity)
		if e != nil {
			return out, e
		}
		for k := range ops[i].(map[string]any) {
			if !fields[k] {
				return out, fmt.Errorf("journey operation does not accept field %s", k)
			}
		}
	}
	_, e = hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != JourneyPatchSchema || len(out.ExpectedSourceSHA256) != 64 || e != nil || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || len(out.Actor) > 256 || len(out.Reason) > 4096 || !stableID.MatchString(out.NodeID) || len(out.Operations) < 1 || len(out.Operations) > 500 {
		return out, fmt.Errorf("journey patch requires schema, source SHA256, actor, reason, node_id and 1..500 operations")
	}
	return out, nil
}
func journeySource(n *Node, values map[string]any) (JourneyModel, map[string]any, bool, error) {
	var m JourneyModel
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	if n.Definition == nil || (n.Definition.ID != "wmds/component/road" && n.Definition.ID != "wmds/component/roadfork") {
		return m, nil, bound, fmt.Errorf("journey requires a typed road or roadfork component")
	}
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return m, nil, bound, e
	}
	var source map[string]any
	if e = strictInto(args, &source); e != nil {
		return m, nil, bound, e
	}
	m.Kind = strings.TrimPrefix(n.Definition.ID, "wmds/component/")
	var layout JourneyLayout
	layoutFields := map[string]any{}
	for _, k := range []string{"direction", "roadW", "amp", "waves", "labelW", "forkW", "small"} {
		if v, ok := source[k]; ok {
			layoutFields[k] = v
		}
	}
	if e = strictInto(layoutFields, &layout); e != nil {
		return m, nil, bound, e
	}
	if n.Placement != nil {
		layout.Rect = n.Placement.Rect
	}
	m.Layout = layout
	decodeList := func(raw any, path string) ([]JourneyMilestone, error) {
		var list []map[string]any
		if e := strictInto(raw, &list); e != nil {
			return nil, e
		}
		keys, e := cycleKeys(n.Keys, path, len(list))
		if e != nil {
			return nil, e
		}
		out := []JourneyMilestone{}
		for i, v := range list {
			active := v["here"] == true || v["active"] == true
			delete(v, "here")
			delete(v, "active")
			var it JourneyMilestone
			if e = strictInto(v, &it); e != nil {
				return nil, e
			}
			it.Key = keys[i]
			if len(n.Keys) == 0 && strings.HasPrefix(it.Key, "source-") {
				it.Key = strings.ReplaceAll(path, "/", "-") + "-" + it.Key
			}
			if active {
				if m.Current != "" {
					return nil, fmt.Errorf("journey has multiple current markers")
				}
				m.Current = it.Key
			}
			out = append(out, it)
		}
		return out, nil
	}
	if m.Kind == "road" {
		m.Trunk, e = decodeList(source["milestones"], "milestones")
	} else {
		m.Mode, _ = source["mode"].(string)
		if source["trunk"] != nil {
			m.Trunk, e = decodeList(source["trunk"], "trunk")
		}
		if e != nil {
			return m, nil, bound, e
		}
		if source["fork"] != nil {
			var f JourneyFork
			if e = strictInto(source["fork"], &f); e != nil {
				return m, nil, bound, e
			}
			m.Fork = &f
		}
		var branches []map[string]any
		if e = strictInto(source["branches"], &branches); e != nil {
			return m, nil, bound, e
		}
		keys, e := cycleKeys(n.Keys, "branches", len(branches))
		if e != nil {
			return m, nil, bound, e
		}
		for i, b := range branches {
			branch := JourneyBranch{Key: keys[i]}
			branch.Title, _ = b["title"].(string)
			branch.Text, _ = b["text"].(string)
			branch.Milestones, e = decodeList(b["milestones"], fmt.Sprintf("branches/%d/milestones", i))
			if e != nil {
				return m, nil, bound, e
			}
			m.Branches = append(m.Branches, branch)
		}
		if chosen, ok := source["chosen"].(float64); ok {
			if chosen != float64(int(chosen)) || int(chosen) < 0 || int(chosen) >= len(keys) {
				return m, nil, bound, fmt.Errorf("invalid chosen branch index")
			}
			m.Chosen = keys[int(chosen)]
		}
	}
	if e != nil {
		return m, nil, bound, e
	}
	return m, source, bound, validateJourney(m)
}
func validateJourney(m JourneyModel) error {
	seen := map[string]bool{}
	check := func(list []JourneyMilestone) error {
		for _, v := range list {
			if !stableID.MatchString(v.Key) || seen[v.Key] || strings.TrimSpace(v.Label) == "" {
				return fmt.Errorf("invalid/duplicate journey milestone %s", v.Key)
			}
			seen[v.Key] = true
			if v.At != nil && (*v.At < 0 || *v.At > 1) {
				return fmt.Errorf("milestone position outside 0..1")
			}
		}
		return nil
	}
	if m.Kind != "road" && m.Kind != "roadfork" {
		return fmt.Errorf("unknown journey kind")
	}
	if e := check(m.Trunk); e != nil {
		return e
	}
	branches := map[string]bool{}
	if m.Kind == "road" {
		if len(m.Trunk) < 1 || len(m.Trunk) > 24 || len(m.Branches) > 0 || m.Chosen != "" || m.Fork != nil {
			return fmt.Errorf("road requires 1..24 milestones without branches/fork")
		}
	} else {
		if len(m.Trunk) > 12 || len(m.Branches) < 2 || len(m.Branches) > 3 || m.Mode != "" && m.Mode != "parallel" && m.Mode != "decision" {
			return fmt.Errorf("roadfork requires 0..12 trunk steps and 2..3 options")
		}
		for _, b := range m.Branches {
			if !stableID.MatchString(b.Key) || branches[b.Key] || strings.TrimSpace(b.Title) == "" || len(b.Milestones) < 1 || len(b.Milestones) > 6 {
				return fmt.Errorf("invalid journey branch %s", b.Key)
			}
			branches[b.Key] = true
			if e := check(b.Milestones); e != nil {
				return e
			}
		}
		if m.Chosen != "" && (!branches[m.Chosen] || m.Mode != "decision") {
			return fmt.Errorf("chosen requires known branch in decision mode")
		}
	}
	if m.Current != "" && !seen[m.Current] {
		return fmt.Errorf("unknown current milestone")
	}
	return nil
}
func journeyList(m *JourneyModel, branch string) (*[]JourneyMilestone, error) {
	if branch == "" {
		return &m.Trunk, nil
	}
	for i, b := range m.Branches {
		if b.Key == branch {
			return &m.Branches[i].Milestones, nil
		}
	}
	return nil, fmt.Errorf("unknown branch %s", branch)
}
func applyJourneyOperation(m *JourneyModel, op JourneyOperation) error {
	switch op.Action + "/" + op.Entity {
	case "set/milestone":
		if op.Milestone == nil || op.Milestone.Key != op.Key {
			return fmt.Errorf("matching complete milestone required")
		}
		list, e := journeyList(m, op.Branch)
		if e != nil {
			return e
		}
		for i, v := range *list {
			if v.Key == op.Key {
				(*list)[i] = *op.Milestone
				return nil
			}
		}
		*list = append(*list, *op.Milestone)
	case "remove/milestone":
		if m.Current == op.Key && !op.Cascade {
			return fmt.Errorf("current milestone removal requires cascade")
		}
		lists := []*[]JourneyMilestone{&m.Trunk}
		for i := range m.Branches {
			lists = append(lists, &m.Branches[i].Milestones)
		}
		for _, list := range lists {
			for i, v := range *list {
				if v.Key == op.Key {
					*list = append((*list)[:i], (*list)[i+1:]...)
					if m.Current == op.Key {
						m.Current = ""
					}
					return nil
				}
			}
		}
		return fmt.Errorf("unknown milestone")
	case "set/branch":
		if op.BranchValue == nil || op.BranchValue.Key != op.Key {
			return fmt.Errorf("matching complete branch required")
		}
		for i, b := range m.Branches {
			if b.Key == op.Key {
				if !op.Cascade {
					old := map[string]bool{}
					for _, v := range op.BranchValue.Milestones {
						old[v.Key] = true
					}
					for _, v := range b.Milestones {
						if v.Key == m.Current && !old[v.Key] {
							return fmt.Errorf("branch replacement drops current milestone; reassign marker first")
						}
					}
				}
				m.Branches[i] = *op.BranchValue
				return nil
			}
		}
		m.Branches = append(m.Branches, *op.BranchValue)
	case "remove/branch":
		for i, b := range m.Branches {
			if b.Key == op.Key {
				dependent := m.Chosen == op.Key
				for _, v := range b.Milestones {
					dependent = dependent || m.Current == v.Key
				}
				if !op.Cascade {
					return fmt.Errorf("populated branch removal requires cascade")
				}
				if dependent {
					if m.Chosen == op.Key {
						m.Chosen = ""
					}
					for _, v := range b.Milestones {
						if m.Current == v.Key {
							m.Current = ""
						}
					}
				}
				m.Branches = append(m.Branches[:i], m.Branches[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("unknown branch")
	case "reorder/milestone":
		list, e := journeyList(m, op.Branch)
		if e != nil {
			return e
		}
		if len(op.Order) != len(*list) {
			return fmt.Errorf("reorder requires all milestones")
		}
		seen := map[string]bool{}
		next := []JourneyMilestone{}
		for _, k := range op.Order {
			if seen[k] {
				return fmt.Errorf("duplicate reorder key")
			}
			seen[k] = true
			found := false
			for _, v := range *list {
				if v.Key == k {
					next = append(next, v)
					found = true
				}
			}
			if !found {
				return fmt.Errorf("unknown reorder key")
			}
		}
		*list = next
	case "reorder/branch":
		if len(op.Order) != len(m.Branches) {
			return fmt.Errorf("reorder requires all branches")
		}
		seen := map[string]bool{}
		next := []JourneyBranch{}
		for _, k := range op.Order {
			if seen[k] {
				return fmt.Errorf("duplicate reorder branch")
			}
			seen[k] = true
			found := false
			for _, b := range m.Branches {
				if b.Key == k {
					next = append(next, b)
					found = true
				}
			}
			if !found {
				return fmt.Errorf("unknown reorder branch")
			}
		}
		m.Branches = next
	case "set/current":
		m.Current = op.Key
	case "remove/current":
		m.Current = ""
	case "set/chosen":
		m.Chosen = op.Key
	case "remove/chosen":
		m.Chosen = ""
	case "set/fork":
		if m.Kind != "roadfork" || op.Fork == nil {
			return fmt.Errorf("fork required on roadfork")
		}
		m.Fork = op.Fork
	case "set/layout":
		if op.Layout == nil {
			return fmt.Errorf("complete layout required")
		}
		rect := m.Layout.Rect
		m.Layout = *op.Layout
		if m.Layout.Rect == nil {
			m.Layout.Rect = rect
		}
	default:
		return fmt.Errorf("materialize/source must be first")
	}
	return nil
}
func lowerJourney(n *Node, m JourneyModel, source map[string]any) error {
	if e := validateJourney(m); e != nil {
		return e
	}
	keys := map[string][]string{}
	lower := func(list []JourneyMilestone, path string) ([]any, error) {
		out := []any{}
		keys[path] = []string{}
		for _, it := range list {
			var v map[string]any
			_ = json.Unmarshal(canonical(it), &v)
			delete(v, "key")
			if m.Kind == "roadfork" {
				if it.At != nil || it.Side != "" {
					return nil, fmt.Errorf("roadfork positions use branch order; at/side applies only to road")
				}
				if it.Key == m.Current {
					v["here"] = true
				}
			} else {
				delete(v, "hereLabel")
				if it.HereLabel != "" {
					return nil, fmt.Errorf("road current marker has stock label; hereLabel only supports roadfork")
				}
				if it.Key == m.Current {
					v["active"] = true
				}
			}
			keys[path] = append(keys[path], it.Key)
			out = append(out, v)
		}
		return out, nil
	}
	trunk, e := lower(m.Trunk, "trunk")
	if e != nil {
		return e
	}
	for _, k := range []string{"direction", "roadW", "amp", "waves", "labelW", "forkW", "small"} {
		delete(source, k)
	}
	var layout map[string]any
	_ = json.Unmarshal(canonical(m.Layout), &layout)
	delete(layout, "rect")
	for k, v := range layout {
		source[k] = v
	}
	delete(source, "chosen")
	if m.Kind == "road" {
		delete(keys, "trunk")
		keys["milestones"] = []string{}
		for _, v := range m.Trunk {
			keys["milestones"] = append(keys["milestones"], v.Key)
		}
		source["milestones"] = trunk
		for _, k := range []string{"forkW", "small"} {
			if _, ok := source[k]; ok {
				return fmt.Errorf("road does not accept %s", k)
			}
		}
	} else {
		for _, k := range []string{"direction", "amp", "waves", "labelW"} {
			if _, ok := source[k]; ok {
				return fmt.Errorf("roadfork does not accept %s", k)
			}
		}
		source["trunk"] = trunk
		source["mode"] = m.Mode
		if m.Fork != nil {
			source["fork"] = m.Fork
		}
		branches := []any{}
		keys["branches"] = []string{}
		for i, b := range m.Branches {
			list, e := lower(b.Milestones, fmt.Sprintf("branches/%d/milestones", i))
			if e != nil {
				return e
			}
			branches = append(branches, map[string]any{"title": b.Title, "text": b.Text, "milestones": list})
			keys["branches"] = append(keys["branches"], b.Key)
			if b.Key == m.Chosen {
				source["chosen"] = i
			}
		}
		source["branches"] = branches
	}
	n.Arguments = source
	n.Keys = keys
	if m.Layout.Rect != nil {
		if n.Placement == nil {
			return fmt.Errorf("missing journey placement")
		}
		n.Placement.Rect = m.Layout.Rect
	}
	return nil
}
func InspectJourney(p *Project, slideID, nodeID, bundle, engine string) (JourneyInspection, error) {
	out := JourneyInspection{Schema: "pptxgengo.journey-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	list, i := findDiagramNode(&t.Nodes, nodeID)
	if list == nil {
		return out, fmt.Errorf("unknown journey node")
	}
	var bound bool
	out.Model, _, bound, e = journeySource(&(*list)[i], p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Bound source: edit values or explicitly materialize/source first"
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func PatchJourney(p *Project, slideID string, patch JourneyPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeJourneyPatch(canonical(patch), "journey-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("journey source hash mismatch")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	list, i := findDiagramNode(&clone.Nodes, patch.NodeID)
	if list == nil {
		return empty, fmt.Errorf("unknown journey node")
	}
	n := &(*list)[i]
	m, s, bound, e := journeySource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	for i, op := range patch.Operations {
		if op.Action == "materialize" && op.Entity == "source" && i == 0 {
			bound = false
			continue
		}
		if bound {
			return empty, fmt.Errorf("bound journey requires materialize/source first")
		}
		if e = applyJourneyOperation(&m, op); e != nil {
			return empty, e
		}
	}
	if e = lowerJourney(n, m, s); e != nil {
		return empty, e
	}
	return CompositionCandidate(p, slideID, "journey", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
