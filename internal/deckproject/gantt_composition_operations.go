package deckproject

import (
	"encoding/json"
	"fmt"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func ganttGroupIndex(s *wmdesign.GanttSpec, key string) int {
	for i, g := range s.Groups {
		if g.Key == key {
			return i
		}
	}
	return -1
}
func ganttLaneIndex(g *wmdesign.GanttGroup, key string) int {
	for i, l := range g.Lanes {
		if l.Key == key {
			return i
		}
	}
	return -1
}
func ganttTaskIndex(l *wmdesign.GanttLane, key string) int {
	for i, it := range l.Items {
		if it.Key == key {
			return i
		}
	}
	return -1
}
func ganttKey(raw json.RawMessage) string { var k string; _ = json.Unmarshal(raw, &k); return k }

func ganttTaskMeaning(it wmdesign.GanttItem) error {
	if it.Kind != "" {
		if it.At != 0 || it.Event != "" || it.Tag || it.Milestone || it.LabelSide != "" {
			return fmt.Errorf("interval task must not carry event, marker or label-side fields")
		}
	} else {
		if it.From != 0 || it.To != 0 || it.Progress != nil || it.SoftStart != 0 || it.SoftEnd != 0 || it.Hatch {
			return fmt.Errorf("event task must not carry interval/progress fields")
		}
		if it.Tag && (it.Milestone || it.Event != "" || it.LabelSide != "") {
			return fmt.Errorf("tag task must not carry another marker or label-side")
		}
		if it.Milestone && it.Event != "" {
			return fmt.Errorf("milestone task must not also carry an event kind")
		}
	}
	return nil
}
func ganttLaneMeaning(l wmdesign.GanttLane) error {
	for _, it := range l.Items {
		if e := ganttTaskMeaning(it); e != nil {
			return e
		}
	}
	return nil
}
func ganttGroupMeaning(g wmdesign.GanttGroup) error {
	for _, l := range g.Lanes {
		if e := ganttLaneMeaning(l); e != nil {
			return e
		}
	}
	return nil
}

func ganttLaneDrops(old, new wmdesign.GanttLane) bool {
	for _, it := range old.Items {
		if ganttTaskIndex(&new, it.Key) < 0 {
			return true
		}
	}
	return false
}
func ganttGroupDrops(old, new wmdesign.GanttGroup) bool {
	for _, lane := range old.Lanes {
		i := ganttLaneIndex(&new, lane.Key)
		if i < 0 || ganttLaneDrops(lane, new.Lanes[i]) {
			return true
		}
	}
	return false
}

func ganttReorder[T any](in []T, order []string, key func(T) string) ([]T, error) {
	if len(order) != len(in) {
		return nil, fmt.Errorf("reorder must list every existing key exactly once")
	}
	by := map[string]T{}
	for _, v := range in {
		by[key(v)] = v
	}
	out := []T{}
	for _, k := range order {
		v, ok := by[k]
		if !ok {
			return nil, fmt.Errorf("unknown or duplicate reorder key %s", k)
		}
		out = append(out, v)
		delete(by, k)
	}
	return out, nil
}

func ganttOperationFields(op GanttOperation, allowed ...string) error {
	fields := map[string]any{}
	_ = json.Unmarshal(canonical(op), &fields)
	ok := map[string]bool{"action": true, "entity": true}
	for _, k := range allowed {
		ok[k] = true
	}
	for k := range fields {
		if !ok[k] {
			return fmt.Errorf("Gantt %s %s does not accept %s", op.Action, op.Entity, k)
		}
	}
	return nil
}

// Validate authored field presence before omitempty can erase a zero, false or
// empty irrelevant field. API callers also receive semantic checks below.
func ganttAuthoredOperationFields(action, entity string) (map[string]bool, error) {
	allowed := map[string]bool{"action": true, "entity": true}
	var fields []string
	switch entity + "/" + action {
	case "source/materialize":
	case "periods/set":
		fields = []string{"labels", "sublabels"}
	case "today/set":
		fields = []string{"at"}
	case "today/remove":
	case "layout/set":
		fields = []string{"track_pitch_pt", "group_width_pt", "lane_width_pt", "legend_full_width", "legend_size_pt"}
	case "group/reorder", "phase/reorder", "gate/reorder":
		fields = []string{"order"}
	case "group/set":
		fields = []string{"key", "group_value", "cascade"}
	case "group/remove":
		fields = []string{"key", "cascade"}
	case "phase/set", "gate/set":
		fields = []string{"key", entity}
	case "phase/remove", "gate/remove":
		fields = []string{"key"}
	case "lane/set":
		fields = []string{"group", "key", "lane_value", "cascade"}
	case "lane/remove":
		fields = []string{"group", "key", "cascade"}
	case "lane/reorder":
		fields = []string{"group", "order"}
	case "lane/move":
		fields = []string{"group", "key", "to_group"}
	case "task/set":
		fields = []string{"group", "lane", "key", "task"}
	case "task/remove":
		fields = []string{"group", "lane", "key"}
	case "task/reorder":
		fields = []string{"group", "lane", "order"}
	case "task/move":
		fields = []string{"group", "lane", "key", "to_group", "to_lane"}
	default:
		return nil, fmt.Errorf("unsupported Gantt action/entity %s/%s", action, entity)
	}
	for _, field := range fields {
		allowed[field] = true
	}
	return allowed, nil
}

func applyGanttOperation(s *wmdesign.GanttSpec, op GanttOperation) error {
	switch op.Entity {
	case "periods":
		if op.Action != "set" {
			return fmt.Errorf("periods requires set")
		}
		if e := ganttOperationFields(op, "labels", "sublabels"); e != nil {
			return e
		}
		if len(op.Labels) == 0 || len(op.Labels) > 100 || len(op.Sublabels) > 0 && len(op.Sublabels) != len(op.Labels) {
			return fmt.Errorf("periods needs 1..100 labels and matching optional sublabels")
		}
		s.Periods.Labels, s.Periods.Sublabels = op.Labels, op.Sublabels
		return nil
	case "today":
		if e := ganttOperationFields(op, "at"); e != nil {
			return e
		}
		if op.Action == "set" && op.At != nil {
			s.Today = op.At
			return nil
		}
		if op.Action == "remove" && op.At == nil {
			s.Today = nil
			return nil
		}
		return fmt.Errorf("today requires set with at or remove")
	case "layout":
		if op.Action != "set" {
			return fmt.Errorf("layout requires set")
		}
		if e := ganttOperationFields(op, "track_pitch_pt", "group_width_pt", "lane_width_pt", "legend_full_width", "legend_size_pt"); e != nil {
			return e
		}
		if op.TrackPitch == nil && op.GroupWidth == nil && op.LaneWidth == nil && op.LegendFullWidth == nil && op.LegendSize == nil {
			return fmt.Errorf("layout requires a width or pitch")
		}
		if op.TrackPitch != nil {
			if *op.TrackPitch < 24 || *op.TrackPitch > 60 {
				return fmt.Errorf("track_pitch_pt requires 24..60")
			}
			s.TrackPitch = *op.TrackPitch
		}
		if op.GroupWidth != nil {
			s.Cols.Group = *op.GroupWidth
		}
		if op.LaneWidth != nil {
			s.Cols.Lane = *op.LaneWidth
		}
		if op.LegendFullWidth != nil {
			s.LegendFullWidth = *op.LegendFullWidth
		}
		if op.LegendSize != nil {
			s.LegendSize = *op.LegendSize
		}
		return nil
	case "group":
		if op.Action == "reorder" {
			if e := ganttOperationFields(op, "order"); e != nil {
				return e
			}
			var e error
			s.Groups, e = ganttReorder(s.Groups, op.Order, func(g wmdesign.GanttGroup) string { return g.Key })
			return e
		}
		if !stableID.MatchString(op.Key) {
			return fmt.Errorf("group needs stable key")
		}
		i := ganttGroupIndex(s, op.Key)
		if op.Action == "set" {
			if e := ganttOperationFields(op, "key", "group_value", "cascade"); e != nil {
				return e
			}
			if op.GroupValue == nil || op.GroupValue.Key != op.Key {
				return fmt.Errorf("group set requires matching group_value key")
			}
			if e := ganttGroupMeaning(*op.GroupValue); e != nil {
				return e
			}
			if i < 0 {
				s.Groups = append(s.Groups, *op.GroupValue)
			} else {
				if ganttGroupDrops(s.Groups[i], *op.GroupValue) && !op.Cascade {
					return fmt.Errorf("group replacement drops descendants; requires cascade true")
				}
				s.Groups[i] = *op.GroupValue
			}
			return nil
		}
		if op.Action == "remove" {
			if e := ganttOperationFields(op, "key", "cascade"); e != nil {
				return e
			}
			if i < 0 {
				return fmt.Errorf("unknown group %s", op.Key)
			}
			if len(s.Groups[i].Lanes) > 0 && !op.Cascade {
				return fmt.Errorf("nonempty group removal requires cascade true")
			}
			s.Groups = append(s.Groups[:i], s.Groups[i+1:]...)
			return nil
		}
	case "phase", "gate":
		if op.Action == "reorder" {
			if e := ganttOperationFields(op, "order"); e != nil {
				return e
			}
			var e error
			if op.Entity == "phase" {
				s.Phases, e = ganttReorder(s.Phases, op.Order, func(p wmdesign.GanttPhase) string { return p.Key })
			} else {
				s.Gates, e = ganttReorder(s.Gates, op.Order, func(g wmdesign.GanttGate) string { return ganttKey(g.Key) })
			}
			return e
		}
		if !stableID.MatchString(op.Key) {
			return fmt.Errorf("%s needs stable key", op.Entity)
		}
		i := -1
		if op.Entity == "phase" {
			for j, p := range s.Phases {
				if p.Key == op.Key {
					i = j
				}
			}
		} else {
			for j, g := range s.Gates {
				if ganttKey(g.Key) == op.Key {
					i = j
				}
			}
		}
		if op.Action == "set" {
			if e := ganttOperationFields(op, "key", op.Entity); e != nil {
				return e
			}
			if op.Entity == "phase" {
				if op.Phase == nil || op.Phase.Key != op.Key {
					return fmt.Errorf("phase set requires matching phase key")
				}
				if i < 0 {
					s.Phases = append(s.Phases, *op.Phase)
				} else {
					s.Phases[i] = *op.Phase
				}
			} else {
				if op.Gate == nil || op.Gate.Key != op.Key {
					return fmt.Errorf("gate set requires matching string gate key")
				}
				gate := wmdesign.GanttGate{Key: canonical(op.Gate.Key), At: op.Gate.At, Label: op.Gate.Label, Callout: op.Gate.Callout}
				if i < 0 {
					s.Gates = append(s.Gates, gate)
				} else {
					s.Gates[i] = gate
				}
			}
			return nil
		}
		if op.Action == "remove" {
			if e := ganttOperationFields(op, "key"); e != nil {
				return e
			}
			if i < 0 {
				return fmt.Errorf("unknown %s %s", op.Entity, op.Key)
			}
			if op.Entity == "phase" {
				s.Phases = append(s.Phases[:i], s.Phases[i+1:]...)
			} else {
				s.Gates = append(s.Gates[:i], s.Gates[i+1:]...)
			}
			return nil
		}
	case "lane", "task":
		gi := ganttGroupIndex(s, op.Group)
		if gi < 0 {
			return fmt.Errorf("unknown parent group %s", op.Group)
		}
		g := &s.Groups[gi]
		if op.Entity == "lane" {
			if op.Action == "reorder" {
				if e := ganttOperationFields(op, "group", "order"); e != nil {
					return e
				}
				var e error
				g.Lanes, e = ganttReorder(g.Lanes, op.Order, func(l wmdesign.GanttLane) string { return l.Key })
				return e
			}
			if !stableID.MatchString(op.Key) {
				return fmt.Errorf("lane needs stable key")
			}
			li := ganttLaneIndex(g, op.Key)
			if op.Action == "set" {
				if e := ganttOperationFields(op, "group", "key", "lane_value", "cascade"); e != nil {
					return e
				}
				if op.LaneValue == nil || op.LaneValue.Key != op.Key {
					return fmt.Errorf("lane set requires matching lane_value key")
				}
				if e := ganttLaneMeaning(*op.LaneValue); e != nil {
					return e
				}
				if li < 0 {
					g.Lanes = append(g.Lanes, *op.LaneValue)
				} else {
					if ganttLaneDrops(g.Lanes[li], *op.LaneValue) && !op.Cascade {
						return fmt.Errorf("lane replacement drops tasks; requires cascade true")
					}
					g.Lanes[li] = *op.LaneValue
				}
				return nil
			}
			if op.Action == "remove" {
				if e := ganttOperationFields(op, "group", "key", "cascade"); e != nil {
					return e
				}
				if li < 0 {
					return fmt.Errorf("unknown lane %s", op.Key)
				}
				if len(g.Lanes[li].Items) > 0 && !op.Cascade {
					return fmt.Errorf("nonempty lane removal requires cascade true")
				}
				g.Lanes = append(g.Lanes[:li], g.Lanes[li+1:]...)
				return nil
			}
			if op.Action == "move" {
				if e := ganttOperationFields(op, "group", "key", "to_group"); e != nil {
					return e
				}
				di := ganttGroupIndex(s, op.ToGroup)
				if li < 0 || di < 0 || di == gi {
					return fmt.Errorf("lane move requires existing different destination group")
				}
				if ganttLaneIndex(&s.Groups[di], op.Key) >= 0 {
					return fmt.Errorf("destination lane key already exists")
				}
				l := g.Lanes[li]
				g.Lanes = append(g.Lanes[:li], g.Lanes[li+1:]...)
				s.Groups[di].Lanes = append(s.Groups[di].Lanes, l)
				return nil
			}
		} else {
			li := ganttLaneIndex(g, op.Lane)
			if li < 0 {
				return fmt.Errorf("unknown parent lane %s", op.Lane)
			}
			l := &g.Lanes[li]
			if op.Action == "reorder" {
				if e := ganttOperationFields(op, "group", "lane", "order"); e != nil {
					return e
				}
				var e error
				l.Items, e = ganttReorder(l.Items, op.Order, func(it wmdesign.GanttItem) string { return it.Key })
				return e
			}
			if !stableID.MatchString(op.Key) {
				return fmt.Errorf("task needs stable key")
			}
			ii := ganttTaskIndex(l, op.Key)
			if op.Action == "set" {
				if e := ganttOperationFields(op, "group", "lane", "key", "task"); e != nil {
					return e
				}
				if op.Task == nil || op.Task.Key != op.Key {
					return fmt.Errorf("task set requires matching task key")
				}
				if e := ganttTaskMeaning(*op.Task); e != nil {
					return e
				}
				if ii < 0 {
					l.Items = append(l.Items, *op.Task)
				} else {
					l.Items[ii] = *op.Task
				}
				return nil
			}
			if op.Action == "remove" {
				if e := ganttOperationFields(op, "group", "lane", "key"); e != nil {
					return e
				}
				if ii < 0 {
					return fmt.Errorf("unknown task %s", op.Key)
				}
				l.Items = append(l.Items[:ii], l.Items[ii+1:]...)
				return nil
			}
			if op.Action == "move" {
				if e := ganttOperationFields(op, "group", "lane", "key", "to_group", "to_lane"); e != nil {
					return e
				}
				di := ganttGroupIndex(s, op.ToGroup)
				if ii < 0 || di < 0 {
					return fmt.Errorf("task move requires existing task and destination group")
				}
				dl := ganttLaneIndex(&s.Groups[di], op.ToLane)
				if dl < 0 || di == gi && dl == li {
					return fmt.Errorf("task move requires existing different destination lane")
				}
				dest := &s.Groups[di].Lanes[dl]
				if ganttTaskIndex(dest, op.Key) >= 0 {
					return fmt.Errorf("destination task key already exists")
				}
				it := l.Items[ii]
				l.Items = append(l.Items[:ii], l.Items[ii+1:]...)
				dest.Items = append(dest.Items, it)
				return nil
			}
		}
	}
	return fmt.Errorf("unsupported Gantt action/entity %s/%s", op.Action, op.Entity)
}
