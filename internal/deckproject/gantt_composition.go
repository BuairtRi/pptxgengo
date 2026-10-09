package deckproject

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

const GanttPatchSchema = "pptxgengo.gantt-patch.v1"

type GanttGateValue struct {
	Key     string  `json:"key"`
	At      float64 `json:"at"`
	Label   string  `json:"label"`
	Callout bool    `json:"callout,omitempty"`
}

// Records supplied to set replace the complete entity. No missing value is
// interpreted as an instruction to keep an old field.
type GanttOperation struct {
	Action          string               `json:"action"`
	Entity          string               `json:"entity"`
	Key             string               `json:"key,omitempty"`
	Group           string               `json:"group,omitempty"`
	Lane            string               `json:"lane,omitempty"`
	ToGroup         string               `json:"to_group,omitempty"`
	ToLane          string               `json:"to_lane,omitempty"`
	Cascade         bool                 `json:"cascade,omitempty"`
	Order           []string             `json:"order,omitempty"`
	GroupValue      *wmdesign.GanttGroup `json:"group_value,omitempty"`
	LaneValue       *wmdesign.GanttLane  `json:"lane_value,omitempty"`
	Task            *wmdesign.GanttItem  `json:"task,omitempty"`
	Phase           *wmdesign.GanttPhase `json:"phase,omitempty"`
	Gate            *GanttGateValue      `json:"gate,omitempty"`
	Labels          []string             `json:"labels,omitempty"`
	Sublabels       []string             `json:"sublabels,omitempty"`
	At              *float64             `json:"at,omitempty"`
	TrackPitch      *float64             `json:"track_pitch_pt,omitempty"`
	GroupWidth      *float64             `json:"group_width_pt,omitempty"`
	LaneWidth       *float64             `json:"lane_width_pt,omitempty"`
	LegendFullWidth *bool                `json:"legend_full_width,omitempty"`
	LegendSize      *float64             `json:"legend_size_pt,omitempty"`
}
type GanttPatch struct {
	Schema               string           `json:"schema"`
	ExpectedSourceSHA256 string           `json:"expected_source_sha256"`
	Actor                string           `json:"actor"`
	Reason               string           `json:"reason"`
	NodeID               string           `json:"node_id"`
	Timebase             string           `json:"timebase"`
	Operations           []GanttOperation `json:"operations"`
}
type GanttInspection struct {
	Schema             string              `json:"schema"`
	SlideID            string              `json:"slide_id"`
	NodeID             string              `json:"node_id"`
	SourceSHA256       string              `json:"source_sha256"`
	Timebase           string              `json:"timebase"`
	CoordinateContract string              `json:"coordinate_contract"`
	Schedule           wmdesign.GanttSpec  `json:"schedule"`
	Keys               map[string][]string `json:"keys"`
	MutationBlocked    string              `json:"mutation_blocked,omitempty"`
	RenderError        string              `json:"render_error,omitempty"`
	Geometry           DiagramInspection   `json:"geometry"`
}

func DecodeGanttPatch(raw []byte, file string) (GanttPatch, error) {
	var out GanttPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("Gantt patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty Gantt patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("Gantt patch requires one document")
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
	object, _ := v.(map[string]any)
	operations, _ := object["operations"].([]any)
	for i, rawOperation := range operations {
		fields, ok := rawOperation.(map[string]any)
		if !ok {
			return out, fmt.Errorf("Gantt operation %d requires an object", i+1)
		}
		if i >= len(out.Operations) {
			return out, fmt.Errorf("invalid Gantt operations")
		}
		allowed, e := ganttAuthoredOperationFields(out.Operations[i].Action, out.Operations[i].Entity)
		if e != nil {
			return out, e
		}
		for key := range fields {
			if !allowed[key] {
				return out, fmt.Errorf("Gantt %s %s does not accept authored field %s", out.Operations[i].Action, out.Operations[i].Entity, key)
			}
		}
	}
	_, digestErr := hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != GanttPatchSchema || len(out.ExpectedSourceSHA256) != 64 || digestErr != nil || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || strings.TrimSpace(out.Reason) == "" || len(out.Reason) > 4096 || !stableID.MatchString(out.NodeID) || out.Timebase != "periods" || len(out.Operations) == 0 || len(out.Operations) > 500 {
		return out, fmt.Errorf("Gantt patch requires schema, source SHA256, actor, reason, node_id, timebase periods and 1..500 operations")
	}
	return out, nil
}

func ganttNode(t *LocalTemplate, id string) (*Node, error) {
	list, i := findDiagramNode(&t.Nodes, id)
	if list == nil {
		return nil, fmt.Errorf("unknown Gantt node %s", id)
	}
	n := &(*list)[i]
	if n.Definition == nil || n.Definition.ID != "wmds/component/gantt" {
		return nil, fmt.Errorf("node %s is not a semantic Gantt component; detach a Gantt source template first", id)
	}
	return n, nil
}

func ganttSource(n *Node, values map[string]any) (wmdesign.GanttSpec, map[string][]string, bool, error) {
	var s wmdesign.GanttSpec
	bound := false
	collectBindings(n.Arguments, func(string) { bound = true })
	args, e := resolveArguments(n.Arguments, values)
	if e != nil {
		return s, nil, bound, e
	}
	if e = strictInto(args, &s); e != nil {
		return s, nil, bound, e
	}
	keys, e := normalizeGanttKeys(&s, n.Keys)
	return s, keys, bound, e
}

// Freeze current renderer identities into the records before reordering. Legacy
// boolean gate keys carry emphasis; they are converted to callout + stable key.
func normalizeGanttKeys(s *wmdesign.GanttSpec, old map[string][]string) (map[string][]string, error) {
	keys := map[string][]string{}
	assign := func(path string, index int, key string) (string, error) {
		external, has := old[path]
		if !has {
			external, has = old["/"+path]
		}
		if has {
			if index >= len(external) {
				return "", fmt.Errorf("missing Gantt array key %s", path)
			}
			if key != "" && key != external[index] {
				return "", fmt.Errorf("conflicting Gantt key %s", path)
			}
			key = external[index]
		}
		if key == "" {
			key = fmt.Sprintf("slot-%03d", index+1)
		}
		if !stableID.MatchString(key) {
			return "", fmt.Errorf("invalid Gantt key %s", key)
		}
		for _, k := range keys[path] {
			if key == k {
				return "", fmt.Errorf("duplicate Gantt key %s in %s", key, path)
			}
		}
		keys[path] = append(keys[path], key)
		return key, nil
	}
	for i := range s.Periods.Labels {
		if _, e := assign("periods/labels", i, ""); e != nil {
			return nil, e
		}
	}
	for i := range s.Phases {
		k, e := assign("phases", i, s.Phases[i].Key)
		if e != nil {
			return nil, e
		}
		s.Phases[i].Key = k
	}
	for i := range s.Gates {
		var key string
		var mark bool
		raw := s.Gates[i].Key
		if len(raw) > 0 && string(raw) != "null" {
			if e := json.Unmarshal(raw, &key); e != nil {
				if e = json.Unmarshal(raw, &mark); e != nil {
					return nil, e
				}
				s.Gates[i].Callout = s.Gates[i].Callout || mark
			}
		}
		k, e := assign("gates", i, key)
		if e != nil {
			return nil, e
		}
		s.Gates[i].Key = canonical(k)
	}
	for gi := range s.Groups {
		g := &s.Groups[gi]
		k, e := assign("groups", gi, g.Key)
		if e != nil {
			return nil, e
		}
		g.Key = k
		for li := range g.Lanes {
			l := &g.Lanes[li]
			path := fmt.Sprintf("groups/%d/lanes", gi)
			k, e := assign(path, li, l.Key)
			if e != nil {
				return nil, e
			}
			l.Key = k
			for ii := range l.Items {
				it := &l.Items[ii]
				path := fmt.Sprintf("groups/%d/lanes/%d/items", gi, li)
				k, e := assign(path, ii, it.Key)
				if e != nil {
					return nil, e
				}
				it.Key = k
			}
		}
	}
	return keys, nil
}

func InspectGantt(p *Project, slideID, nodeID, bundle, engine string) (GanttInspection, error) {
	out := GanttInspection{Schema: "pptxgengo.gantt-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash(), Timebase: "periods", CoordinateContract: "0 is the first period's left boundary; N is the last period's right boundary. Fractional positions are allowed. Labels do not define calendar dates, timezone or business-day arithmetic."}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	n, e := ganttNode(&t, nodeID)
	if e != nil {
		return out, e
	}
	var bound bool
	out.Schedule, out.Keys, bound, e = ganttSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Source arguments contain bindings: edit authored bound values, or explicitly begin the patch with action materialize, entity source. Materialization moves this component's resolved copy into its local template; other components' bindings are retained."
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
		out.Geometry.Validation = "renderer_failed_no_measurement_qualification"
	}
	return out, nil
}

func PatchGantt(p *Project, slideID string, patch GanttPatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeGanttPatch(canonical(patch), "gantt-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("Gantt source hash mismatch; inspect again before applying")
	}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return empty, e
	}
	var clone LocalTemplate
	if e = strictInto(t, &clone); e != nil {
		return empty, e
	}
	n, e := ganttNode(&clone, patch.NodeID)
	if e != nil {
		return empty, e
	}
	s, keys, bound, e := ganttSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	for i, op := range patch.Operations {
		if op.Action == "materialize" && op.Entity == "source" {
			if i != 0 || !bytes.Equal(canonical(op), canonical(GanttOperation{Action: "materialize", Entity: "source"})) {
				return empty, fmt.Errorf("materialize must be the first operation with entity source only")
			}
			bound = false
			continue
		}
		if bound {
			return empty, fmt.Errorf("Gantt patch refuses bound arguments without an explicit first materialize source operation")
		}
		if e = applyGanttOperation(&s, op); e != nil {
			return empty, e
		}
		if op.Entity == "periods" {
			keys["periods/labels"] = nil
			for i := range s.Periods.Labels {
				keys["periods/labels"] = append(keys["periods/labels"], fmt.Sprintf("slot-%03d", i+1))
			}
		}
	}
	// Recreate positional key maps from the stable keys carried by the records.
	keys, e = normalizeGanttKeys(&s, map[string][]string{"periods/labels": keys["periods/labels"]})
	if e != nil {
		return empty, e
	}
	var args map[string]any
	if e = json.Unmarshal(canonical(s), &args); e != nil {
		return empty, e
	}
	for _, reserved := range []string{"type", "x", "y", "w"} {
		delete(args, reserved)
	}
	n.Arguments, n.Keys = args, keys
	return CompositionCandidate(p, slideID, "gantt", patch.Actor, patch.Reason, clone, bundle, engine, apply)
}
