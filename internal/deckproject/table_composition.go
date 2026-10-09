package deckproject

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
	"io"
	"math"
	"reflect"
	"strings"
)

const TablePatchSchema = "pptxgengo.table-patch.v1"

type TableRow struct {
	Key    string         `json:"key"`
	Values map[string]any `json:"values"`
}
type TableGroup struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Members []string `json:"members"`
	Fill    string   `json:"fill,omitempty"`
}

// Inline sections are full-width group-header rows followed by their members.
// They differ from rowGroups, which label explicit contiguous ranges at the side.
type TableInlineSection struct {
	Header  string   `json:"header"`
	Label   string   `json:"label"`
	Members []string `json:"members"`
}
type TableModel struct {
	Arguments           map[string]any       `json:"arguments"`
	Columns             []map[string]any     `json:"columns"`
	Rows                []TableRow           `json:"rows"`
	RowGroups           []TableGroup         `json:"row_groups,omitempty"`
	ColumnGroups        []TableGroup         `json:"column_groups,omitempty"`
	InlineSections      []TableInlineSection `json:"inline_sections,omitempty"`
	SparseLeadingLabels []string             `json:"sparse_leading_labels,omitempty"`
}
type TableOperation struct {
	Action      string         `json:"action"`
	Entity      string         `json:"entity"`
	Key         string         `json:"key,omitempty"`
	Column      string         `json:"column,omitempty"`
	Section     string         `json:"section,omitempty"`
	Cascade     bool           `json:"cascade,omitempty"`
	Order       []string       `json:"order,omitempty"`
	Row         map[string]any `json:"row,omitempty"`
	ColumnValue map[string]any `json:"column_value,omitempty"`
	Cells       map[string]any `json:"cells,omitempty"`
	Value       any            `json:"value,omitempty"`
	Group       *TableGroup    `json:"group,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
	Rect        *wmdesign.Rect `json:"rect,omitempty"`
}
type TablePatch struct {
	Schema               string           `json:"schema"`
	ExpectedSourceSHA256 string           `json:"expected_source_sha256"`
	Actor                string           `json:"actor"`
	Reason               string           `json:"reason"`
	SemanticReview       string           `json:"semantic_review"`
	NodeID               string           `json:"node_id"`
	Operations           []TableOperation `json:"operations"`
}
type TableInspection struct {
	Schema          string            `json:"schema"`
	SlideID         string            `json:"slide_id"`
	NodeID          string            `json:"node_id"`
	SourceSHA256    string            `json:"source_sha256"`
	Model           TableModel        `json:"model"`
	MutationBlocked string            `json:"mutation_blocked,omitempty"`
	RenderError     string            `json:"render_error,omitempty"`
	Geometry        DiagramInspection `json:"geometry"`
}

func tableFields(a, e string) (map[string]bool, error) {
	s := "action entity"
	switch a + "/" + e {
	case "materialize/source":
	case "set/row":
		s += " key row section"
	case "move/row":
		s += " key section"
	case "set/column":
		s += " key column_value cells"
	case "set/cell":
		s += " key column value"
	case "set/row_group", "set/column_group":
		s += " key group"
	case "remove/row", "remove/column":
		s += " key cascade"
	case "remove/row_group", "remove/column_group":
		s += " key"
	case "reorder/row", "reorder/column", "reorder/row_group", "reorder/column_group":
		s += " order"
	case "set/layout":
		s += " options rect"
	default:
		return nil, fmt.Errorf("unsupported table %s/%s", a, e)
	}
	out := map[string]bool{}
	for _, k := range strings.Fields(s) {
		out[k] = true
	}
	return out, nil
}
func DecodeTablePatch(raw []byte, file string) (TablePatch, error) {
	var out TablePatch
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("table patch exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty table patch")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("one table patch document required")
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
		fields, e := tableFields(op.Action, op.Entity)
		if e != nil {
			return out, e
		}
		authored := ops[i].(map[string]any)
		for k := range authored {
			if !fields[k] {
				return out, fmt.Errorf("table operation does not accept field %s", k)
			}
		}
		if op.Action == "set" && op.Entity == "cell" {
			if _, ok := authored["value"]; !ok {
				return out, fmt.Errorf("set cell requires explicit value including null")
			}
		}
	}
	_, e = hex.DecodeString(out.ExpectedSourceSHA256)
	if out.Schema != TablePatchSchema || len(out.ExpectedSourceSHA256) != 64 || e != nil || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.Reason) == "" || strings.TrimSpace(out.SemanticReview) == "" || len(out.Actor) > 256 || len(out.Reason) > 4096 || len(out.SemanticReview) > 4096 || !stableID.MatchString(out.NodeID) || len(out.Operations) < 1 || len(out.Operations) > 500 {
		return out, fmt.Errorf("table patch requires schema, SHA256, actor, reason, semantic_review, node_id and operations")
	}
	return out, nil
}
func (op TableOperation) MarshalJSON() ([]byte, error) {
	type plain TableOperation
	b, e := json.Marshal(plain(op))
	if e != nil {
		return nil, e
	}
	var obj map[string]any
	if e = json.Unmarshal(b, &obj); e != nil {
		return nil, e
	}
	if op.Action == "set" && op.Entity == "cell" {
		obj["value"] = op.Value
	}
	return json.Marshal(obj)
}

var tableMetadata = map[string]bool{"group": true, "total": true, "ink": true, "scale": true, "h": true}

func tableColumnKeys(m TableModel) ([]string, error) {
	keys := []string{}
	seen := map[string]bool{}
	for _, v := range m.Columns {
		k, ok := v["k"].(string)
		if !ok || !stableID.MatchString(k) || seen[k] {
			return nil, fmt.Errorf("invalid/reserved/duplicate table column %s", k)
		}
		seen[k] = true
		keys = append(keys, k)
	}
	return keys, nil
}
func tableSource(n *Node, values map[string]any) (TableModel, bool, error) {
	var m TableModel
	args, bound, e := componentArgs(n, values)
	if e != nil {
		return m, bound, e
	}
	if n.Definition.ID != "wmds/component/table" && n.Definition.ID != "wmds/component/editable-table" {
		return m, bound, fmt.Errorf("table command requires typed table component")
	}
	m.Arguments = args
	cols, e := lookupPointer(args, "/cols")
	if e != nil {
		return m, bound, e
	}
	if e = strictInto(cols, &m.Columns); e != nil {
		return m, bound, e
	}
	columnKeys, e := tableColumnKeys(m)
	if e != nil {
		return m, bound, e
	}
	rows, ok := args["rows"].([]any)
	if !ok {
		return m, bound, fmt.Errorf("table rows must be array")
	}
	keys, e := componentArrayKeys(n, "rows", rows)
	if e != nil {
		return m, bound, e
	}
	for i, row := range rows {
		values, ok := row.(map[string]any)
		if !ok {
			return m, bound, fmt.Errorf("table row must be object")
		}
		m.Rows = append(m.Rows, TableRow{Key: keys[i], Values: values})
	}
	m.InlineSections = tableInlineSections(m.Rows)
	m.SparseLeadingLabels = tableSparseLeadingLabels(m)
	decodeGroups := func(field string, members []string) ([]TableGroup, error) {
		if args[field] == nil {
			return nil, nil
		}
		list, ok := args[field].([]any)
		if !ok {
			return nil, fmt.Errorf("table groups require array")
		}
		keys, e := componentArrayKeys(n, field, list)
		if e != nil {
			return nil, e
		}
		out := []TableGroup{}
		for i, v := range list {
			obj, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid table group")
			}
			from, ok1 := obj["from"].(float64)
			to, ok2 := obj["to"].(float64)
			if !ok1 || !ok2 || from != float64(int(from)) || to != float64(int(to)) || from < 0 || to < from || int(to) >= len(members) {
				return nil, fmt.Errorf("invalid table group range")
			}
			g := TableGroup{Key: keys[i], Members: append([]string{}, members[int(from):int(to)+1]...)}
			g.Label, _ = obj["label"].(string)
			g.Fill, _ = obj["fill"].(string)
			for k := range obj {
				if k != "label" && k != "from" && k != "to" && k != "fill" {
					return nil, fmt.Errorf("unsupported group field %s", k)
				}
			}
			out = append(out, g)
		}
		return out, nil
	}
	m.RowGroups, e = decodeGroups("rowGroups", keys)
	if e != nil {
		return m, bound, e
	}
	m.ColumnGroups, e = decodeGroups("groups", columnKeys)
	if e != nil {
		return m, bound, e
	}
	for _, k := range []string{"cols", "rows", "rowGroups", "groups"} {
		delete(m.Arguments, k)
	}
	return m, bound, nil
}
func tableInlineSections(rows []TableRow) []TableInlineSection {
	out := []TableInlineSection{}
	for _, row := range rows {
		if label, ok := row.Values["group"].(string); ok && label != "" {
			out = append(out, TableInlineSection{Header: row.Key, Label: label, Members: []string{}})
		} else if len(out) > 0 {
			i := len(out) - 1
			out[i].Members = append(out[i].Members, row.Key)
		}
	}
	return out
}
func tableInlineMembership(rows []TableRow) map[string]string {
	out := map[string]string{}
	header := ""
	for _, row := range rows {
		if label, ok := row.Values["group"].(string); ok && label != "" {
			header = row.Key
			continue
		}
		out[row.Key] = header
	}
	return out
}
func tableSparseLeadingLabels(m TableModel) []string {
	if len(m.Columns) == 0 || len(m.Rows) == 0 {
		return nil
	}
	k, _ := m.Columns[0]["k"].(string)
	typ, _ := m.Columns[0]["type"].(string)
	if typ != "" && typ != "text" {
		return nil
	}
	missing := []string{}
	visible := false
	for _, row := range m.Rows {
		if _, header := row.Values["group"]; header {
			continue
		}
		value := row.Values[k]
		if value == nil || value == "" {
			missing = append(missing, row.Key)
		} else if _, ok := value.(string); ok {
			visible = true
		}
	}
	if visible && len(missing) > 0 {
		return missing
	}
	return nil
}
func tableInsertIntoSection(m *TableModel, row TableRow, section string) error {
	last := -1
	for _, s := range tableInlineSections(m.Rows) {
		if s.Header != section {
			continue
		}
		last = tableRowIndex(*m, s.Header)
		for _, key := range s.Members {
			last = tableRowIndex(*m, key)
		}
	}
	if last < 0 {
		return fmt.Errorf("target section must name an existing declared inline header key")
	}
	m.Rows = append(m.Rows, TableRow{})
	copy(m.Rows[last+2:], m.Rows[last+1:])
	m.Rows[last+1] = row
	return nil
}
func tableRowIndex(m TableModel, k string) int {
	for i, v := range m.Rows {
		if v.Key == k {
			return i
		}
	}
	return -1
}
func tableColumnIndex(m TableModel, k string) int {
	for i, v := range m.Columns {
		if v["k"] == k {
			return i
		}
	}
	return -1
}
func tableRemoveMember(groups []TableGroup, key string, cascade bool) ([]TableGroup, error) {
	out := []TableGroup{}
	for _, g := range groups {
		next := []string{}
		for _, k := range g.Members {
			if k != key {
				next = append(next, k)
			}
		}
		if len(next) == 0 {
			if !cascade {
				return nil, fmt.Errorf("deletion removes group; cascade required")
			}
			continue
		}
		g.Members = next
		out = append(out, g)
	}
	return out, nil
}
func tableRowHeaderEnabled(m TableModel) bool {
	if m.Arguments["rowHeader"] == true {
		return true
	}
	text, _ := m.Arguments["rowHeader"].(string)
	return text != ""
}
func applyTableOperation(m *TableModel, n *Node, op TableOperation) error {
	switch op.Action + "/" + op.Entity {
	case "set/row":
		if !stableID.MatchString(op.Key) || op.Row == nil {
			return fmt.Errorf("row requires stable key and values")
		}
		i := tableRowIndex(*m, op.Key)
		if i < 0 {
			_, isHeader := op.Row["group"]
			for _, c := range m.Columns {
				if isHeader {
					break
				}
				k := c["k"].(string)
				if _, ok := op.Row[k]; !ok {
					return fmt.Errorf("new row requires explicit cell %s including null", k)
				}
			}
			row := TableRow{Key: op.Key, Values: op.Row}
			if len(tableInlineSections(m.Rows)) > 0 {
				if _, header := op.Row["group"]; !header {
					if op.Section == "" {
						return fmt.Errorf("new row in an inline section table requires explicit section header key")
					}
					return tableInsertIntoSection(m, row, op.Section)
				}
			}
			if op.Section != "" {
				return fmt.Errorf("section only applies to a new ordinary row in an inline section table")
			}
			m.Rows = append(m.Rows, row)
		} else {
			if op.Section != "" {
				return fmt.Errorf("use explicit move/row to change existing section membership")
			}
			for k, v := range op.Row {
				m.Rows[i].Values[k] = v
			}
		}
	case "move/row":
		i := tableRowIndex(*m, op.Key)
		if i < 0 || op.Section == "" {
			return fmt.Errorf("move row requires existing row and explicit target section header key")
		}
		if len(tableSparseLeadingLabels(*m)) > 0 {
			return fmt.Errorf("sparse leading labels require explicit per-row source labels before moving rows; blank cells do not declare membership")
		}
		row := m.Rows[i]
		if _, header := row.Values["group"]; header {
			return fmt.Errorf("move row cannot move section headers; reorder the complete section block")
		}
		found := false
		for _, s := range tableInlineSections(m.Rows) {
			if s.Header == op.Section {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("unknown target inline section")
		}
		m.Rows = append(m.Rows[:i], m.Rows[i+1:]...)
		return tableInsertIntoSection(m, row, op.Section)
	case "set/cell":
		i, j := tableRowIndex(*m, op.Key), tableColumnIndex(*m, op.Column)
		if i < 0 || j < 0 {
			return fmt.Errorf("unknown row/column")
		}
		m.Rows[i].Values[op.Column] = op.Value
	case "set/column":
		if op.ColumnValue == nil || op.ColumnValue["k"] != op.Key || !stableID.MatchString(op.Key) {
			return fmt.Errorf("matching complete nonreserved column required")
		}
		i := tableColumnIndex(*m, op.Key)
		if i < 0 {
			if tableMetadata[op.Key] {
				return fmt.Errorf("new reserved column would collide with row metadata")
			}
			if len(op.Cells) != len(m.Rows) {
				return fmt.Errorf("new column requires explicit cell for every stable row")
			}
			for _, r := range m.Rows {
				v, ok := op.Cells[r.Key]
				if !ok {
					return fmt.Errorf("missing explicit column cell for %s", r.Key)
				}
				r.Values[op.Key] = v
			}
			m.Columns = append(m.Columns, op.ColumnValue)
		} else {
			if len(op.Cells) > 0 {
				return fmt.Errorf("existing column domain changes preserve cells; patch cells separately")
			}
			m.Columns[i] = op.ColumnValue
		}
	case "remove/row":
		i := tableRowIndex(*m, op.Key)
		if i < 0 {
			return fmt.Errorf("unknown row")
		}
		if !op.Cascade {
			return fmt.Errorf("row removal requires cascade acknowledgement of cells/metadata/group relationships")
		}
		for _, section := range tableInlineSections(m.Rows) {
			if section.Header == op.Key && len(section.Members) > 0 {
				return fmt.Errorf("inline section header has members; remove its members first or explicitly replace the reviewed section source")
			}
		}
		groups, e := tableRemoveMember(m.RowGroups, op.Key, true)
		if e != nil {
			return e
		}
		m.RowGroups = groups
		m.Rows = append(m.Rows[:i], m.Rows[i+1:]...)
	case "remove/column":
		i := tableColumnIndex(*m, op.Key)
		if i == 0 && tableRowHeaderEnabled(*m) {
			return fmt.Errorf("styled row header column removal requires explicit rowHeader disable/reassignment")
		}
		if i < 0 {
			return fmt.Errorf("unknown column")
		}
		if !op.Cascade {
			return fmt.Errorf("column removal requires cascade acknowledgement including zero-valued cells")
		}
		for _, r := range m.Rows {
			delete(r.Values, op.Key)
		}
		groups, e := tableRemoveMember(m.ColumnGroups, op.Key, true)
		if e != nil {
			return e
		}
		m.ColumnGroups = groups
		if m.Arguments["highlight"] == op.Key {
			delete(m.Arguments, "highlight")
		}
		m.Columns = append(m.Columns[:i], m.Columns[i+1:]...)
	case "set/row_group", "set/column_group":
		if op.Group == nil || op.Group.Key != op.Key || !stableID.MatchString(op.Key) {
			return fmt.Errorf("matching complete group required")
		}
		groups := &m.RowGroups
		if op.Entity == "column_group" {
			groups = &m.ColumnGroups
			if op.Group.Fill != "" {
				return fmt.Errorf("column group does not support fill")
			}
		}
		for i, g := range *groups {
			if g.Key == op.Key {
				(*groups)[i] = *op.Group
				return nil
			}
		}
		*groups = append(*groups, *op.Group)
	case "remove/row_group", "remove/column_group":
		groups := &m.RowGroups
		if op.Entity == "column_group" {
			groups = &m.ColumnGroups
		}
		for i, g := range *groups {
			if g.Key == op.Key {
				*groups = append((*groups)[:i], (*groups)[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("unknown group")
	case "set/layout":
		allowed := map[string]bool{"header": true, "dense": true, "rowH": true, "rowHeader": true, "preset": true, "highlight": true, "continued": true, "deltaUnit": true, "heatMax": true, "heatMin": true, "groupW": true, "runRate": true}
		for k, v := range op.Options {
			if !allowed[k] {
				return fmt.Errorf("unknown table option %s", k)
			}
			m.Arguments[k] = v
		}
		if op.Rect != nil {
			if n.Placement == nil {
				return fmt.Errorf("missing table placement")
			}
			n.Placement.Rect = op.Rect
		}
	case "reorder/row", "reorder/column", "reorder/row_group", "reorder/column_group":
		if op.Entity == "column" && len(op.Order) > 0 && len(m.Columns) > 0 {
			if tableRowHeaderEnabled(*m) && op.Order[0] != m.Columns[0]["k"] {
				return fmt.Errorf("styled row header must retain first column; explicitly disable rowHeader before reassigning it")
			}
		}
		count := 0
		switch op.Entity {
		case "row":
			count = len(m.Rows)
		case "column":
			count = len(m.Columns)
		case "row_group":
			count = len(m.RowGroups)
		case "column_group":
			count = len(m.ColumnGroups)
		}
		if len(op.Order) != count {
			return fmt.Errorf("reorder requires every key exactly once")
		}
		seen := map[string]bool{}
		rows := []TableRow{}
		cols := []map[string]any{}
		groups := []TableGroup{}
		for _, k := range op.Order {
			if seen[k] {
				return fmt.Errorf("duplicate reorder key")
			}
			seen[k] = true
			found := false
			switch op.Entity {
			case "row":
				for _, v := range m.Rows {
					if v.Key == k {
						rows = append(rows, v)
						found = true
					}
				}
			case "column":
				for _, v := range m.Columns {
					if v["k"] == k {
						cols = append(cols, v)
						found = true
					}
				}
			default:
				g := m.RowGroups
				if op.Entity == "column_group" {
					g = m.ColumnGroups
				}
				for _, v := range g {
					if v.Key == k {
						groups = append(groups, v)
						found = true
					}
				}
			}
			if !found {
				return fmt.Errorf("unknown reorder key")
			}
		}
		switch op.Entity {
		case "row":
			if len(tableSparseLeadingLabels(*m)) > 0 {
				for i, row := range m.Rows {
					if rows[i].Key != row.Key {
						return fmt.Errorf("sparse leading labels require explicit per-row source labels before reordering; blank cells do not declare membership")
					}
				}
			}
			before, after := tableInlineMembership(m.Rows), tableInlineMembership(rows)
			for key, group := range before {
				if after[key] != group {
					return fmt.Errorf("row reorder changes inline section membership for %s; move complete sections or reorder members within their original section", key)
				}
			}
			m.Rows = rows
		case "column":
			m.Columns = cols
		case "row_group":
			m.RowGroups = groups
		case "column_group":
			m.ColumnGroups = groups
		}
	default:
		return fmt.Errorf("materialize/source must be first")
	}
	return nil
}

// Table source edits preserve declared heat meaning rather than relying on the
// legacy renderer's visual clamping. Null/empty cells remain explicitly missing.
func tableValidateHeatDomains(m TableModel) error {
	number := func(v any) (float64, bool) {
		f, ok := v.(float64)
		if !ok {
			switch x := v.(type) {
			case int:
				f, ok = float64(x), true
			case json.Number:
				var e error
				f, e = x.Float64()
				ok = e == nil
			}
		}
		return f, ok && !math.IsNaN(f) && !math.IsInf(f, 0)
	}
	for _, col := range m.Columns {
		if col["type"] != "heat" {
			continue
		}
		key, _ := col["k"].(string)
		domain := func(field, global string, fallback float64) (float64, error) {
			v := col[field]
			if v == nil {
				v = m.Arguments[global]
			}
			if v == nil {
				return fallback, nil
			}
			f, ok := number(v)
			if !ok {
				return 0, fmt.Errorf("heat column %s has invalid %s", key, field)
			}
			return f, nil
		}
		min, e := domain("min", "heatMin", 0)
		if e != nil {
			return e
		}
		max, e := domain("max", "heatMax", 4)
		if e != nil {
			return e
		}
		if max <= min {
			return fmt.Errorf("heat column %s requires max greater than min", key)
		}
		for _, row := range m.Rows {
			v := row.Values[key]
			if v == nil || v == "" {
				continue
			}
			if obj, ok := v.(map[string]any); ok {
				v = obj["value"]
			}
			f, ok := number(v)
			if !ok || f < min || f > max {
				return fmt.Errorf("heat cell %s/%s must be an explicit value in [%g,%g] or missing", row.Key, key, min, max)
			}
		}
	}
	return nil
}
func lowerTable(n *Node, m TableModel) error {
	if e := tableValidateHeatDomains(m); e != nil {
		return e
	}

	columnKeys, e := tableColumnKeys(m)
	if e != nil {
		return e
	}
	if len(m.Columns) < 1 || len(m.Columns) > 12 || len(m.Rows) > 60 {
		return fmt.Errorf("table count outside renderer bounds")
	}
	validCols := map[string]bool{}
	for _, k := range columnKeys {
		validCols[k] = true
	}
	rowKeys := []string{}
	rows := []any{}
	seen := map[string]bool{}
	for _, r := range m.Rows {
		if !stableID.MatchString(r.Key) || seen[r.Key] {
			return fmt.Errorf("invalid/duplicate row key")
		}
		seen[r.Key] = true
		for k := range r.Values {
			if !validCols[k] && !tableMetadata[k] {
				return fmt.Errorf("row has unknown cell/metadata field %s", k)
			}
		}
		rowKeys = append(rowKeys, r.Key)
		rows = append(rows, r.Values)
	}
	args := m.Arguments
	args["cols"] = m.Columns
	args["rows"] = rows
	oldRows, _ := n.Arguments["rows"].([]any)
	oldRowKeys, e := componentArrayKeys(n, "rows", oldRows)
	if e != nil {
		return e
	}
	componentRebaseKeys(n, "rows", oldRowKeys, rowKeys)
	oldCols, _ := n.Arguments["cols"].([]any)
	oldColOverlay, e := componentArrayKeys(n, "cols", oldCols)
	if e != nil {
		return e
	}
	colIdentity := map[string]string{}
	for i, v := range oldCols {
		if obj, ok := v.(map[string]any); ok {
			if k, ok := obj["k"].(string); ok {
				colIdentity[k] = oldColOverlay[i]
			}
		}
	}
	newColOverlay := []string{}
	for _, k := range columnKeys {
		identity := colIdentity[k]
		if identity == "" {
			identity = k
		}
		newColOverlay = append(newColOverlay, identity)
	}
	componentRebaseKeys(n, "cols", oldColOverlay, newColOverlay)
	keys := n.Keys
	keys["rows"] = rowKeys
	lowerGroups := func(groups []TableGroup, field string, members []string) error {
		delete(args, field)
		delete(keys, field)
		delete(keys, "/"+field)
		if len(groups) == 0 {
			if field == "rowGroups" {
				delete(args, "groupW")
			}
			return nil
		}
		pos := map[string]int{}
		for i, k := range members {
			pos[k] = i
		}
		seen, used := map[string]bool{}, map[string]bool{}
		out := []any{}
		for _, g := range groups {
			if !stableID.MatchString(g.Key) || seen[g.Key] || strings.TrimSpace(g.Label) == "" || len(g.Members) == 0 {
				return fmt.Errorf("invalid group")
			}
			seen[g.Key] = true
			min, max := len(members), -1
			unique := map[string]bool{}
			for _, k := range g.Members {
				i, ok := pos[k]
				if !ok || used[k] || unique[k] {
					return fmt.Errorf("group unknown/overlapping/duplicate member %s", k)
				}
				unique[k] = true
				used[k] = true
				if i < min {
					min = i
				}
				if i > max {
					max = i
				}
			}
			if max-min+1 != len(unique) {
				return fmt.Errorf("group members must remain contiguous; explicitly revise group mapping")
			}
			v := map[string]any{"label": g.Label, "from": min, "to": max}
			if g.Fill != "" {
				v["fill"] = g.Fill
			}
			out = append(out, v)
			keys[field] = append(keys[field], g.Key)
		}
		args[field] = out
		return nil
	}
	if e = lowerGroups(m.RowGroups, "rowGroups", rowKeys); e != nil {
		return e
	}
	if e = lowerGroups(m.ColumnGroups, "groups", columnKeys); e != nil {
		return e
	}
	n.Arguments = args
	n.Keys = keys
	return componentPopulateKeys(n, args)
}
func InspectTable(p *Project, slideID, nodeID, bundle, engine string) (TableInspection, error) {
	out := TableInspection{Schema: "pptxgengo.table-inspection.v1", SlideID: slideID, NodeID: nodeID, SourceSHA256: p.SourceHash()}
	idx, t, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	list, i := findDiagramNode(&t.Nodes, nodeID)
	if list == nil {
		return out, fmt.Errorf("unknown table node")
	}
	var bound bool
	out.Model, bound, e = tableSource(&(*list)[i], p.Document.Slides[idx].Values)
	if e != nil {
		return out, e
	}
	if bound {
		out.MutationBlocked = "Bound source: edit values or materialize/source first"
	}
	out.Geometry, e = InspectDiagram(p, slideID, bundle, engine)
	if e != nil {
		out.RenderError = e.Error()
	}
	return out, nil
}
func PatchTable(p *Project, slideID string, patch TablePatch, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	checked, e := DecodeTablePatch(canonical(patch), "table-patch")
	if e != nil {
		return empty, e
	}
	patch = checked
	if patch.ExpectedSourceSHA256 != p.SourceHash() {
		return empty, fmt.Errorf("table source hash mismatch")
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
		return empty, fmt.Errorf("unknown table node")
	}
	n := &(*list)[i]
	m, bound, e := tableSource(n, p.Document.Slides[idx].Values)
	if e != nil {
		return empty, e
	}
	for j, op := range patch.Operations {
		if op.Action == "materialize" && op.Entity == "source" && j == 0 {
			bound = false
			continue
		}
		if bound {
			return empty, fmt.Errorf("bound table requires materialize/source first")
		}
		if e = applyTableOperation(&m, n, op); e != nil {
			return empty, e
		}
	}
	if e = lowerTable(n, m); e != nil {
		return empty, e
	}
	return CompositionCandidate(p, slideID, "table", patch.Actor, patch.Reason+"; semantic review: "+patch.SemanticReview, clone, bundle, engine, apply)
}
