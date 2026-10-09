package deckproject

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

const TeamPatchSchema = "pptxgengo.team-patch.v1"

// Team entities are source semantics, not meanings inferred from native position.
type TeamRole struct {
	Label string `json:"label"`
}
type TeamPodLayout struct {
	Rect    wmdesign.Rect `json:"rect"`
	Columns int           `json:"columns"`
	Gap     float64       `json:"gap_pt"`
}
type TeamMember struct {
	Label string `json:"label"`
	Org   string `json:"org"`
}
type TeamReport struct {
	Key      string       `json:"key"`
	Org      string       `json:"org"`
	Title    string       `json:"title"`
	Name     string       `json:"name"`
	Dotted   bool         `json:"dotted,omitempty"`
	Children []TeamReport `json:"children,omitempty"`
}
type TeamTier struct {
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	Cadence   string     `json:"cadence"`
	Members   [][]string `json:"members"`
	Decisions []any      `json:"decisions"`
}
type TeamOperation struct {
	Action        string         `json:"action"`
	Component     string         `json:"component,omitempty"`
	ID            string         `json:"id,omitempty"`
	Target        string         `json:"target,omitempty"`
	Parent        string         `json:"parent,omitempty"`
	Order         []string       `json:"order,omitempty"`
	Role          *TeamRole      `json:"role,omitempty"`
	Report        *TeamReport    `json:"report,omitempty"`
	Tier          *TeamTier      `json:"tier,omitempty"`
	Member        *TeamMember    `json:"member,omitempty"`
	Decision      *string        `json:"decision,omitempty"`
	Node          *Node          `json:"node,omitempty"`
	Rect          *wmdesign.Rect `json:"rect,omitempty"`
	IncidentEdges string         `json:"incident_edges,omitempty"`
	Layout        *TeamPodLayout `json:"layout,omitempty"`
}
type TeamPatch struct {
	Schema               string          `json:"schema"`
	Actor                string          `json:"actor"`
	Reason               string          `json:"reason"`
	ExpectedSourceSHA256 string          `json:"expected_source_sha256"`
	Operations           []TeamOperation `json:"operations"`
}
type TeamComponent struct {
	ID                string              `json:"id"`
	Type              string              `json:"type"`
	Placement         *Placement          `json:"placement"`
	Arguments         map[string]any      `json:"arguments"`
	ResolvedArguments map[string]any      `json:"resolved_arguments"`
	Keys              map[string][]string `json:"keys"`
	BoundArguments    bool                `json:"bound_arguments"`
}
type TeamInspection struct {
	Schema                 string            `json:"schema"`
	Components             []TeamComponent   `json:"components"`
	Diagram                DiagramInspection `json:"diagram"`
	SemanticReconciliation string            `json:"semantic_reconciliation"`
}
type TeamPatchResult struct {
	CompositionResult
	Patch TeamPatch `json:"patch"`
}

func DecodeTeamPatch(raw []byte, file string) (TeamPatch, error) {
	var out TeamPatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("team patch exceeds 1 MiB")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if err := d.Decode(&doc); err != nil {
		return out, err
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty team patch")
	}
	if err := d.Decode(&extra); err != io.EOF {
		return out, fmt.Errorf("team patch requires one document")
	}
	v, err := p.yamlValue(doc.Content[0], "", 0)
	if err != nil {
		return out, err
	}
	if err = p.shapeType(v, reflect.TypeOf(out), ""); err != nil {
		return out, err
	}
	// Validate authored field presence before omitempty serialization can erase
	// an irrelevant known field whose value is false, empty or null.
	if object, ok := v.(map[string]any); ok {
		if operations, ok := object["operations"].([]any); ok {
			for _, value := range operations {
				operation, ok := value.(map[string]any)
				if !ok {
					continue
				}
				action, _ := operation["action"].(string)
				fields, ok := teamOperationFields(action)
				if !ok {
					return out, fmt.Errorf("unknown team action %s", action)
				}
				permit := map[string]bool{"action": true}
				for _, field := range strings.Fields(fields) {
					permit[field] = true
				}
				for field := range operation {
					if !permit[field] {
						return out, fmt.Errorf("%s does not accept %s", action, field)
					}
				}
			}
		}
	}
	if err = strictInto(v, &out); err != nil {
		return out, err
	}
	if out.Schema != TeamPatchSchema || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || len(out.Actor) > 256 || len(out.Reason) > 4096 || len(out.Operations) < 1 || len(out.Operations) > 500 || !validTeamDigest(out.ExpectedSourceSHA256) {
		return out, fmt.Errorf("team patch requires schema, actor, reason, expected_source_sha256 and 1..500 operations")
	}
	for _, op := range out.Operations {
		if err = validateTeamOperation(op); err != nil {
			return out, err
		}
	}
	return out, nil
}
func validTeamDigest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func teamKind(n Node) string {
	if n.Kind != "component" || n.Definition == nil || n.Definition.Scope != "shared" {
		return ""
	}
	if !strings.HasPrefix(n.Definition.ID, "wmds/component/") && !strings.HasPrefix(n.Definition.ID, "wmds/composite/") {
		return ""
	}
	k := strings.TrimPrefix(strings.TrimPrefix(n.Definition.ID, "wmds/component/"), "wmds/composite/")
	switch k {
	case "pod", "orgchart", "governance", "role":
		return k
	}
	return ""
}
func teamBound(args map[string]any) bool {
	bound := false
	collectBindings(args, func(string) { bound = true })
	return bound
}
func InspectTeam(p *Project, id, bundle, engine string) (TeamInspection, error) {
	out := TeamInspection{Schema: "pptxgengo.team-inspection.v1", Components: []TeamComponent{}, SemanticReconciliation: "Native transforms and supported copy edits do not imply reporting, membership, cadence or decision changes. Project team reconcile proposes authenticated role translations between unchanged pods for explicit membership review; other meaning requires source operations."}
	idx, t, err := diagramSlide(p, id)
	if err != nil {
		return out, err
	}
	var visit func([]Node) error
	visit = func(nodes []Node) error {
		for _, n := range nodes {
			if k := teamKind(n); k != "" {
				var c Node
				if err := strictInto(n, &c); err != nil {
					return err
				}
				bound := teamBound(c.Arguments)
				args := c.Arguments
				resolved, err := resolveArguments(c.Arguments, p.Document.Slides[idx].Values)
				if err != nil {
					return err
				}
				c.Arguments = resolved
				if err := normalizeTeamNode(&c); err != nil {
					return err
				}
				out.Components = append(out.Components, TeamComponent{ID: c.ID, Type: k, Placement: c.Placement, Arguments: args, ResolvedArguments: c.Arguments, Keys: c.Keys, BoundArguments: bound})
			}
			if err := visit(n.Nodes); err != nil {
				return err
			}
		}
		return nil
	}
	if err = visit(t.Nodes); err != nil {
		return out, err
	}
	if len(out.Components) == 0 {
		return out, fmt.Errorf("no typed pod/orgchart/governance/role components; detach a catalog example, then explicitly add supported components with project diagram patch; loose shapes are not inferred as a team")
	}
	out.Diagram, err = InspectDiagram(p, id, bundle, engine)
	return out, err
}
func PatchTeam(p *Project, slideID string, patch TeamPatch, bundle, engine string, apply bool) (TeamPatchResult, error) {
	out := TeamPatchResult{Patch: patch}
	validated, err := DecodeTeamPatch(canonical(patch), "team-patch")
	if err != nil {
		return out, err
	}
	patch = validated
	if p.SourceHash() != patch.ExpectedSourceSHA256 {
		return out, fmt.Errorf("team source changed since inspection; inspect again and review a fresh patch")
	}
	idx, t, err := diagramSlide(p, slideID)
	if err != nil {
		return out, err
	}
	var cloned LocalTemplate
	if err = strictInto(t, &cloned); err != nil {
		return out, err
	}
	t = cloned
	for i, op := range patch.Operations {
		if op.Action == "materialize-component" {
			list, j := findDiagramNode(&t.Nodes, op.Component)
			if list == nil || teamKind((*list)[j]) == "" {
				return out, fmt.Errorf("materialization requires a typed team component")
			}
			args, e := resolveArguments((*list)[j].Arguments, p.Document.Slides[idx].Values)
			if e != nil {
				return out, e
			}
			(*list)[j].Arguments = args
			if e = normalizeTeamNode(&(*list)[j]); e != nil {
				return out, e
			}
			continue
		}
		if err = applyTeamOperation(&t, op); err != nil {
			return out, fmt.Errorf("team operation %d (%s): %w", i+1, op.Action, err)
		}
	}
	out.CompositionResult, err = CompositionCandidate(p, slideID, "team", patch.Actor, patch.Reason, t, bundle, engine, apply)
	return out, err
}

func teamOperationFields(action string) (string, bool) {
	allowed := map[string]string{
		"arrange-pods":  "order layout",
		"add-component": "id node", "remove-component": "component incident_edges", "move-component": "component rect", "materialize-component": "component",
		"add-role": "component id role", "update-role": "component id role", "remove-role": "component id", "reorder-roles": "component order", "reassign-role": "component id target", "rename-role": "component id target",
		"add-report": "component id parent report", "update-report": "component id report", "remove-report": "component id", "reparent-report": "component id parent", "reorder-reports": "component parent order",
		"add-tier": "component id tier", "update-tier": "component id tier", "remove-tier": "component id", "reorder-tiers": "component order",
		"add-member": "component parent id member", "update-member": "component parent id member", "remove-member": "component parent id", "reorder-members": "component parent order",
		"add-decision": "component parent id decision", "update-decision": "component parent id decision", "remove-decision": "component parent id", "reorder-decisions": "component parent order",
	}
	fields, ok := allowed[action]
	return fields, ok
}

func validateTeamOperation(op TeamOperation) error {
	fields, ok := teamOperationFields(op.Action)
	if !ok {
		return fmt.Errorf("unknown team action %s", op.Action)
	}
	var value map[string]any
	if err := strictInto(op, &value); err != nil {
		return err
	}
	permit := map[string]bool{"action": true}
	for _, f := range strings.Fields(fields) {
		permit[f] = true
	}
	for field := range value {
		if !permit[field] {
			return fmt.Errorf("%s does not accept %s", op.Action, field)
		}
	}
	for _, field := range strings.Fields(fields) {
		if field == "incident_edges" {
			continue
		}
		v, ok := value[field]
		if !ok {
			return fmt.Errorf("%s requires %s", op.Action, field)
		}
		if s, yes := v.(string); yes && field != "decision" && (!stableID.MatchString(s) || len(s) > 128) {
			return fmt.Errorf("%s requires stable %s", op.Action, field)
		}
	}
	if len(op.Order) > 500 {
		return fmt.Errorf("team order exceeds 500 entities")
	}
	for _, s := range op.Order {
		if !stableID.MatchString(s) {
			return fmt.Errorf("invalid team order ID %s", s)
		}
	}
	if op.IncidentEdges != "" && op.IncidentEdges != "remove" {
		return fmt.Errorf("incident_edges must be remove")
	}
	return nil
}

// Keep renderer-equivalent slot keys for old unkeyed arrays; authored identities
// thereafter travel with entities when ordering changes, rather than with indexes.
func teamArray(n *Node, path string, items []any, explicit []string) ([]string, error) {
	if n.Keys == nil {
		n.Keys = map[string][]string{}
	}
	keys := n.Keys[path]
	if keys == nil {
		keys = make([]string, len(items))
		for i := range items {
			keys[i] = fmt.Sprintf("slot-%03d", i+1)
			if len(explicit) > i && explicit[i] != "" {
				keys[i] = explicit[i]
			}
		}
	}
	if len(keys) != len(items) {
		return nil, fmt.Errorf("%s keys must match array", path)
	}
	seen := map[string]bool{}
	for i, k := range keys {
		if !stableID.MatchString(k) || seen[k] {
			return nil, fmt.Errorf("invalid or duplicate %s key %s", path, k)
		}
		seen[k] = true
		if len(explicit) > i && explicit[i] != "" && explicit[i] != k {
			return nil, fmt.Errorf("%s explicit key disagrees", path)
		}
	}
	n.Keys[path] = keys
	return keys, nil
}
func teamSlice(v any) ([]any, error) {
	if v == nil {
		return []any{}, nil
	}
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected a literal array; materialize binding arguments explicitly")
	}
	return items, nil
}
func teamMap(v any) (map[string]any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected a literal entity object")
	}
	return m, nil
}
func normalizeTeamNode(n *Node) error {
	if teamBound(n.Arguments) {
		return fmt.Errorf("bound team arguments require explicit local materialization; source content remains authoritative")
	}
	switch teamKind(*n) {
	case "pod":
		items, e := teamSlice(n.Arguments["roles"])
		if e != nil {
			return e
		}
		for _, v := range items {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("pod roles require strings")
			}
		}
		n.Arguments["roles"] = items
		_, e = teamArray(n, "roles", items, nil)
		return e
	case "orgchart":
		root, e := teamMap(n.Arguments["root"])
		if e != nil {
			return e
		}
		return normalizeTeamReports(n, root, "root", map[string]bool{})
	case "governance":
		tiers, e := teamSlice(n.Arguments["tiers"])
		if e != nil {
			return e
		}
		n.Arguments["tiers"] = tiers
		explicit := make([]string, len(tiers))
		for i, v := range tiers {
			m, e := teamMap(v)
			if e != nil {
				return e
			}
			explicit[i], _ = m["key"].(string)
		}
		keys, e := teamArray(n, "tiers", tiers, explicit)
		if e != nil {
			return e
		}
		for i, v := range tiers {
			m, _ := teamMap(v)
			m["key"] = keys[i]
			for _, f := range []string{"members", "decisions"} {
				items, e := teamSlice(m[f])
				if e != nil {
					return e
				}
				m[f] = items
				if _, e = teamArray(n, fmt.Sprintf("tiers/%d/%s", i, f), items, nil); e != nil {
					return e
				}
			}
		}
		return nil
	case "role":
		return nil
	}
	return fmt.Errorf("unsupported team component")
}
func normalizeTeamReports(n *Node, m map[string]any, path string, seen map[string]bool) error {
	key, _ := m["key"].(string)
	if path == "root" && key == "" {
		key = "root"
		m["key"] = key
	}
	if key != "" {
		if !stableID.MatchString(key) || seen[key] {
			return fmt.Errorf("duplicate/invalid reporting ID %s", key)
		}
		seen[key] = true
	}
	children, e := teamSlice(m["children"])
	if e != nil {
		return e
	}
	m["children"] = children
	explicit := make([]string, len(children))
	for i, v := range children {
		c, e := teamMap(v)
		if e != nil {
			return e
		}
		explicit[i], _ = c["key"].(string)
	}
	keys, e := teamArray(n, path+"/children", children, explicit)
	if e != nil {
		return e
	}
	for i, v := range children {
		c, _ := teamMap(v)
		if seen[keys[i]] {
			if explicit[i] != "" {
				return fmt.Errorf("duplicate authored reporting ID %s", keys[i])
			}
			keys[i] = "report-" + digest([]byte(fmt.Sprintf("%s/%d", path, i)))[:12]
		}
		c["key"] = keys[i]
		if e = normalizeTeamReports(n, c, fmt.Sprintf("%s/children/%d", path, i), seen); e != nil {
			return e
		}
	}
	n.Keys[path+"/children"] = keys
	return nil
}
func teamFind(keys []string, id string) int {
	for i, k := range keys {
		if k == id {
			return i
		}
	}
	return -1
}
func teamOrder(items []any, keys, order []string) ([]any, []string, error) {
	if len(keys) != len(order) {
		return nil, nil, fmt.Errorf("order must list every ID exactly once")
	}
	out := make([]any, len(items))
	seen := map[string]bool{}
	for i, k := range order {
		j := teamFind(keys, k)
		if j < 0 || seen[k] {
			return nil, nil, fmt.Errorf("order must list every ID exactly once")
		}
		seen[k] = true
		out[i] = items[j]
	}
	return out, append([]string(nil), order...), nil
}
func teamRemove(items []any, keys []string, i int) ([]any, []string) {
	remainingItems := append([]any(nil), items[:i]...)
	remainingKeys := append([]string(nil), keys[:i]...)
	return append(remainingItems, items[i+1:]...), append(remainingKeys, keys[i+1:]...)
}
func teamObject(v any) map[string]any { var out map[string]any; _ = strictInto(v, &out); return out }

func applyTeamOperation(t *LocalTemplate, op TeamOperation) error {
	if op.Action == "arrange-pods" {
		return arrangeTeamPods(t, op.Order, *op.Layout)
	}
	if op.Action == "add-component" {
		if list, _ := findDiagramNode(&t.Nodes, op.ID); list != nil {
			return fmt.Errorf("component ID already exists")
		}
		if op.Node.ID != op.ID || teamKind(*op.Node) == "" {
			return fmt.Errorf("add-component requires matching typed pod/orgchart/governance/role node")
		}
		var n Node
		if e := strictInto(*op.Node, &n); e != nil {
			return e
		}
		if e := normalizeTeamNode(&n); e != nil {
			return e
		}
		t.Nodes = append(t.Nodes, n)
		return nil
	}
	list, i := findDiagramNode(&t.Nodes, op.Component)
	if list == nil {
		return fmt.Errorf("unknown component %s", op.Component)
	}
	n := &(*list)[i]
	kind := teamKind(*n)
	if kind == "" {
		return fmt.Errorf("component %s is not a typed team component; explicitly convert stock shapes", op.Component)
	}
	if e := normalizeTeamNode(n); e != nil {
		return e
	}
	if op.Action == "remove-component" {
		if e := removeIncident(&t.Nodes, n.ID, op.IncidentEdges == "remove"); e != nil {
			return e
		}
		list, i = findDiagramNode(&t.Nodes, op.Component)
		*list = append((*list)[:i], (*list)[i+1:]...)
		return nil
	}
	if op.Action == "move-component" {
		if n.Placement == nil || n.Placement.Rect == nil {
			return fmt.Errorf("move-component requires rectangle placement; use diagram arrangement for span placement")
		}
		r := *op.Rect
		n.Placement.Rect = &r
		return nil
	}
	if strings.Contains(op.Action, "role") {
		if kind != "pod" {
			return fmt.Errorf("role operations require a pod")
		}
		return patchTeamRoles(t, n, op)
	}
	if strings.Contains(op.Action, "report") {
		if kind != "orgchart" {
			return fmt.Errorf("report operations require orgchart")
		}
		return patchTeamReports(n, op)
	}
	if kind != "governance" {
		return fmt.Errorf("tier/member operations require governance")
	}
	return patchTeamGovernance(n, op)
}

// Arrange uses renderer row budgets, not font scaling or role omission.
func arrangeTeamPods(t *LocalTemplate, ids []string, layout TeamPodLayout) error {
	r := layout.Rect
	if len(ids) == 0 || layout.Columns < 1 || layout.Columns > len(ids) || layout.Gap < 0 || math.IsNaN(r.X+r.Y+r.W+r.H+layout.Gap) || math.IsInf(r.X+r.Y+r.W+r.H+layout.Gap, 0) || r.W <= 0 || r.H <= 0 {
		return fmt.Errorf("arrange-pods requires positive finite allocation, columns within pod count and nonnegative gap")
	}
	width := (r.W - float64(layout.Columns-1)*layout.Gap) / float64(layout.Columns)
	if width <= 24 {
		return fmt.Errorf("pod allocation too narrow")
	}
	seen := map[string]bool{}
	pods := make([]*Node, len(ids))
	zone := ""
	for i, id := range ids {
		list, j := findDiagramNode(&t.Nodes, id)
		if list == nil || teamKind((*list)[j]) != "pod" || seen[id] {
			return fmt.Errorf("arrange-pods requires unique existing pod IDs")
		}
		seen[id] = true
		pods[i] = &(*list)[j]
		if pods[i].Placement == nil || pods[i].Placement.Rect == nil {
			return fmt.Errorf("arrange-pods requires rectangle placements")
		}
		if i == 0 {
			zone = pods[i].Placement.Zone
		} else if pods[i].Placement.Zone != zone {
			return fmt.Errorf("pods must share a frame zone")
		}
		if e := normalizeTeamNode(pods[i]); e != nil {
			return e
		}
	}
	y := r.Y
	for start := 0; start < len(pods); start += layout.Columns {
		end := start + layout.Columns
		if end > len(pods) {
			end = len(pods)
		}
		rowHeight := 0.0
		for _, pod := range pods[start:end] {
			roles, _ := teamSlice(pod.Arguments["roles"])
			rowHeight = math.Max(rowHeight, 48+42*float64(len(roles)))
		}
		if y+rowHeight > r.Y+r.H+.02 {
			return fmt.Errorf("pod rows require more than %.3fpt allocation height; increase capacity or split the slide", r.H)
		}
		for col, pod := range pods[start:end] {
			roles, _ := teamSlice(pod.Arguments["roles"])
			pod.Placement.Rect = &wmdesign.Rect{X: r.X + float64(col)*(width+layout.Gap), Y: y, W: width, H: 48 + 42*float64(len(roles))}
		}
		y += rowHeight + layout.Gap
	}
	return nil
}

func patchTeamRoles(t *LocalTemplate, n *Node, op TeamOperation) error {
	items, _ := teamSlice(n.Arguments["roles"])
	keys := n.Keys["roles"]
	i := teamFind(keys, op.ID)
	switch op.Action {
	case "add-role":
		if i >= 0 {
			return fmt.Errorf("role ID exists")
		}
		if strings.TrimSpace(op.Role.Label) == "" {
			return fmt.Errorf("role label is required")
		}
		items = append(items, op.Role.Label)
		keys = append(keys, op.ID)
	case "update-role":
		if i < 0 {
			return fmt.Errorf("unknown role")
		}
		if strings.TrimSpace(op.Role.Label) == "" {
			return fmt.Errorf("role label is required")
		}
		items[i] = op.Role.Label
	case "remove-role":
		if i < 0 {
			return fmt.Errorf("unknown role")
		}
		items, keys = teamRemove(items, keys, i)
	case "reorder-roles":
		var e error
		items, keys, e = teamOrder(items, keys, op.Order)
		if e != nil {
			return e
		}
	case "reassign-role":
		if i < 0 || op.Target == n.ID {
			return fmt.Errorf("reassign-role requires existing role and different target pod")
		}
		list, j := findDiagramNode(&t.Nodes, op.Target)
		if list == nil || teamKind((*list)[j]) != "pod" {
			return fmt.Errorf("target must be an existing pod")
		}
		target := &(*list)[j]
		if e := normalizeTeamNode(target); e != nil {
			return e
		}
		if teamFind(target.Keys["roles"], op.ID) >= 0 {
			return fmt.Errorf("role ID already exists in target")
		}
		to, _ := teamSlice(target.Arguments["roles"])
		target.Arguments["roles"] = append(to, items[i])
		target.Keys["roles"] = append(target.Keys["roles"], op.ID)
		items, keys = teamRemove(items, keys, i)
	case "rename-role":
		if i < 0 || teamFind(keys, op.Target) >= 0 {
			return fmt.Errorf("rename-role requires existing ID and unused target ID")
		}
		keys[i] = op.Target
	default:
		return fmt.Errorf("unsupported role action")
	}
	n.Arguments["roles"] = items
	n.Keys["roles"] = keys
	return nil
}

func findTeamReport(m map[string]any, id string) (map[string]any, map[string]any, int) {
	if m["key"] == id {
		return m, nil, -1
	}
	children, _ := teamSlice(m["children"])
	for i, v := range children {
		c, _ := teamMap(v)
		if c["key"] == id {
			return c, m, i
		}
		if found, parent, j := findTeamReport(c, id); found != nil {
			return found, parent, j
		}
	}
	return nil, nil, -1
}
func reportContains(m map[string]any, id string) bool {
	found, _, _ := findTeamReport(m, id)
	return found != nil
}
func resetTeamReportKeys(n *Node) error {
	for k := range n.Keys {
		if strings.HasPrefix(k, "root/") {
			delete(n.Keys, k)
		}
	}
	root, _ := teamMap(n.Arguments["root"])
	return normalizeTeamReports(n, root, "root", map[string]bool{})
}
func patchTeamReports(n *Node, op TeamOperation) error {
	root, _ := teamMap(n.Arguments["root"])
	current, parent, index := findTeamReport(root, op.ID)
	switch op.Action {
	case "add-report":
		if current != nil || op.Report.Key != op.ID {
			return fmt.Errorf("add-report requires new matching ID")
		}
		dst, _, _ := findTeamReport(root, op.Parent)
		if dst == nil {
			return fmt.Errorf("unknown reporting parent")
		}
		children, _ := teamSlice(dst["children"])
		dst["children"] = append(children, teamObject(op.Report))
	case "update-report":
		if current == nil || op.Report.Key != op.ID || len(op.Report.Children) > 0 {
			return fmt.Errorf("update-report requires matching ID and omits children; use topology operations")
		}
		current["org"] = op.Report.Org
		current["title"] = op.Report.Title
		current["name"] = op.Report.Name
		current["dotted"] = op.Report.Dotted
	case "remove-report":
		if current == nil || parent == nil {
			return fmt.Errorf("cannot remove root or unknown report")
		}
		children, _ := teamSlice(current["children"])
		if len(children) > 0 {
			return fmt.Errorf("report has children; reparent or remove them explicitly first")
		}
		siblings, _ := teamSlice(parent["children"])
		parent["children"] = append(siblings[:index], siblings[index+1:]...)
	case "reparent-report":
		if current == nil || parent == nil {
			return fmt.Errorf("cannot reparent root or unknown report")
		}
		dst, _, _ := findTeamReport(root, op.Parent)
		if dst == nil || reportContains(current, op.Parent) {
			return fmt.Errorf("unknown parent or reporting cycle")
		}
		siblings, _ := teamSlice(parent["children"])
		parent["children"] = append(siblings[:index], siblings[index+1:]...)
		children, _ := teamSlice(dst["children"])
		dst["children"] = append(children, current)
	case "reorder-reports":
		dst, _, _ := findTeamReport(root, op.Parent)
		if dst == nil {
			return fmt.Errorf("unknown reporting parent")
		}
		children, _ := teamSlice(dst["children"])
		keys := make([]string, len(children))
		for i, v := range children {
			c, _ := teamMap(v)
			keys[i], _ = c["key"].(string)
		}
		ordered, _, e := teamOrder(children, keys, op.Order)
		if e != nil {
			return e
		}
		dst["children"] = ordered
	default:
		return fmt.Errorf("unsupported reporting action")
	}
	return resetTeamReportKeys(n)
}

// Carry tier member/decision keys by tier identity across index changes.
func teamTierKeys(n *Node, tiers []any) map[string]map[string][]string {
	out := map[string]map[string][]string{}
	for i, v := range tiers {
		m, _ := teamMap(v)
		id, _ := m["key"].(string)
		out[id] = map[string][]string{}
		for _, f := range []string{"members", "decisions"} {
			out[id][f] = append([]string(nil), n.Keys[fmt.Sprintf("tiers/%d/%s", i, f)]...)
		}
	}
	return out
}
func restoreTeamTierKeys(n *Node, tiers []any, retained map[string]map[string][]string) error {
	for k := range n.Keys {
		if k == "tiers" || strings.HasPrefix(k, "tiers/") {
			delete(n.Keys, k)
		}
	}
	for i, v := range tiers {
		m, _ := teamMap(v)
		id, _ := m["key"].(string)
		for f, keys := range retained[id] {
			n.Keys[fmt.Sprintf("tiers/%d/%s", i, f)] = keys
		}
	}
	n.Arguments["tiers"] = tiers
	return normalizeTeamNode(n)
}
func patchTeamGovernance(n *Node, op TeamOperation) error {
	tiers, _ := teamSlice(n.Arguments["tiers"])
	keys := n.Keys["tiers"]
	retained := teamTierKeys(n, tiers)
	if strings.Contains(op.Action, "decision") {
		i := teamFind(keys, op.Parent)
		if i < 0 {
			return fmt.Errorf("unknown governance tier")
		}
		tier, _ := teamMap(tiers[i])
		items, _ := teamSlice(tier["decisions"])
		dk := retained[op.Parent]["decisions"]
		j := teamFind(dk, op.ID)
		switch op.Action {
		case "add-decision":
			if j >= 0 {
				return fmt.Errorf("decision ID exists")
			}
			items = append(items, *op.Decision)
			dk = append(dk, op.ID)
		case "update-decision":
			if j < 0 {
				return fmt.Errorf("unknown decision")
			}
			items[j] = *op.Decision
		case "remove-decision":
			if j < 0 {
				return fmt.Errorf("unknown decision")
			}
			items, dk = teamRemove(items, dk, j)
		case "reorder-decisions":
			var e error
			items, dk, e = teamOrder(items, dk, op.Order)
			if e != nil {
				return e
			}
		}
		if op.Decision != nil && strings.TrimSpace(*op.Decision) == "" {
			return fmt.Errorf("decision copy is required")
		}
		tier["decisions"] = items
		retained[op.Parent]["decisions"] = dk
	} else if strings.Contains(op.Action, "member") {
		i := teamFind(keys, op.Parent)
		if i < 0 {
			return fmt.Errorf("unknown governance tier")
		}
		tier, _ := teamMap(tiers[i])
		members, _ := teamSlice(tier["members"])
		mk := retained[op.Parent]["members"]
		j := teamFind(mk, op.ID)
		switch op.Action {
		case "add-member":
			if j >= 0 {
				return fmt.Errorf("member ID exists")
			}
			members = append(members, []any{op.Member.Label, op.Member.Org})
			mk = append(mk, op.ID)
		case "update-member":
			if j < 0 {
				return fmt.Errorf("unknown member")
			}
			members[j] = []any{op.Member.Label, op.Member.Org}
		case "remove-member":
			if j < 0 {
				return fmt.Errorf("unknown member")
			}
			members, mk = teamRemove(members, mk, j)
		case "reorder-members":
			var e error
			members, mk, e = teamOrder(members, mk, op.Order)
			if e != nil {
				return e
			}
		}
		if op.Member != nil && (strings.TrimSpace(op.Member.Label) == "" || op.Member.Org != "wm" && op.Member.Org != "client" && op.Member.Org != "tbd") {
			return fmt.Errorf("member requires label and org wm/client/tbd")
		}
		tier["members"] = members
		retained[op.Parent]["members"] = mk
	} else {
		i := teamFind(keys, op.ID)
		switch op.Action {
		case "add-tier":
			if i >= 0 || op.Tier.Key != op.ID {
				return fmt.Errorf("add-tier requires new matching ID")
			}
			tiers = append(tiers, teamObject(op.Tier))
		case "update-tier":
			if i < 0 || op.Tier.Key != op.ID {
				return fmt.Errorf("update-tier requires matching existing ID")
			}
			tier, _ := teamMap(tiers[i])
			if len(op.Tier.Members) > 0 || len(op.Tier.Decisions) > 0 {
				return fmt.Errorf("update-tier omits members and decisions; use their operations to preserve identities")
			}
			tier["name"] = op.Tier.Name
			tier["cadence"] = op.Tier.Cadence
		case "remove-tier":
			if i < 0 {
				return fmt.Errorf("unknown tier")
			}
			tiers, keys = teamRemove(tiers, keys, i)
			delete(retained, op.ID)
		case "reorder-tiers":
			var e error
			tiers, keys, e = teamOrder(tiers, keys, op.Order)
			if e != nil {
				return e
			}
		default:
			return fmt.Errorf("unsupported tier action")
		}
	}
	return restoreTeamTierKeys(n, tiers, retained)
}
