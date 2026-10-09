package deckproject

import (
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"strings"
)

func applyAssessmentOperation(s *wmdesign.AssessmentSpec, op AssessmentOperation) error {
	rowIndex := func(key string) int {
		for i, v := range s.Rows {
			if v.Key == key {
				return i
			}
		}
		return -1
	}
	colIndex := func(key string) int {
		for i, v := range s.Columns {
			if v.Key == key {
				return i
			}
		}
		return -1
	}
	switch {
	case op.Action == "set" && (op.Entity == "row" || op.Entity == "column"):
		if !stableID.MatchString(op.Key) || strings.TrimSpace(op.Label) == "" {
			return fmt.Errorf("assessment axis requires stable key and nonblank label")
		}
		if op.Entity == "row" {
			i := rowIndex(op.Key)
			if i < 0 {
				s.Rows = append(s.Rows, wmdesign.AssessmentRow{Key: op.Key, Label: op.Label, Scores: map[string]*int{}})
			} else {
				s.Rows[i].Label = op.Label
			}
		} else {
			i := colIndex(op.Key)
			if i < 0 {
				s.Columns = append(s.Columns, wmdesign.AssessmentAxis{Key: op.Key, Label: op.Label})
			} else {
				s.Columns[i].Label = op.Label
			}
		}
	case op.Action == "remove" && op.Entity == "row":
		i := rowIndex(op.Key)
		if i < 0 {
			return fmt.Errorf("unknown assessment row %s", op.Key)
		}
		for _, v := range s.Rows[i].Scores {
			if v != nil && !op.Cascade {
				return fmt.Errorf("removing scored row requires cascade acknowledgement")
			}
		}
		s.Rows = append(s.Rows[:i], s.Rows[i+1:]...)
	case op.Action == "remove" && op.Entity == "column":
		i := colIndex(op.Key)
		if i < 0 {
			return fmt.Errorf("unknown assessment column %s", op.Key)
		}
		for _, row := range s.Rows {
			if row.Scores[op.Key] != nil && !op.Cascade {
				return fmt.Errorf("removing scored column requires cascade acknowledgement")
			}
		}
		s.Columns = append(s.Columns[:i], s.Columns[i+1:]...)
		for j := range s.Rows {
			delete(s.Rows[j].Scores, op.Key)
		}
	case op.Action == "reorder" && (op.Entity == "row" || op.Entity == "column"):
		count := len(s.Rows)
		if op.Entity == "column" {
			count = len(s.Columns)
		}
		if len(op.Order) != count {
			return fmt.Errorf("assessment reorder requires every existing key exactly once")
		}
		seen := map[string]bool{}
		rows := []wmdesign.AssessmentRow{}
		cols := []wmdesign.AssessmentAxis{}
		for _, key := range op.Order {
			if seen[key] {
				return fmt.Errorf("duplicate assessment reorder key %s", key)
			}
			seen[key] = true
			if op.Entity == "row" {
				i := rowIndex(key)
				if i < 0 {
					return fmt.Errorf("unknown assessment row %s", key)
				}
				rows = append(rows, s.Rows[i])
			} else {
				i := colIndex(key)
				if i < 0 {
					return fmt.Errorf("unknown assessment column %s", key)
				}
				cols = append(cols, s.Columns[i])
			}
		}
		if op.Entity == "row" {
			s.Rows = rows
		} else {
			s.Columns = cols
		}
	case op.Action == "set" && op.Entity == "score":
		i := rowIndex(op.Key)
		if i < 0 || colIndex(op.Column) < 0 {
			return fmt.Errorf("score requires existing row and column")
		}
		if (op.Score == nil) == !op.Missing {
			return fmt.Errorf("score requires exactly one numeric score (including zero) or missing: true")
		}
		if op.Score != nil && (*op.Score < 0 || *op.Score > s.Domain.Max) {
			return fmt.Errorf("score outside 0..%d", s.Domain.Max)
		}
		if s.Rows[i].Scores == nil {
			s.Rows[i].Scores = map[string]*int{}
		}
		s.Rows[i].Scores[op.Column] = op.Score
	case op.Action == "set" && op.Entity == "domain":
		if op.Domain == nil {
			return fmt.Errorf("domain requires complete value")
		}
		s.Domain = *op.Domain
	case op.Action == "set" && op.Entity == "layout":
		if op.LabelWidth == nil && op.RowHeight == nil && op.ShowScores == nil {
			return fmt.Errorf("assessment layout requires label width or row height")
		}
		if op.LabelWidth != nil {
			s.LabelWidth = *op.LabelWidth
		}
		if op.ShowScores != nil {
			s.ShowScores = *op.ShowScores
		}
		if op.RowHeight != nil {
			s.RowHeight = *op.RowHeight
		}
	default:
		return fmt.Errorf("unsupported assessment operation %s %s", op.Action, op.Entity)
	}
	return nil
}
