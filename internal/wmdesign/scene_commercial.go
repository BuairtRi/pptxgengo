package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// CommercialModel stores decimal source facts and explicit dimensional formulas.
// Calculations use rational arithmetic; rounding is only applied to outputs.
type CommercialFormula struct {
	Operation string   `json:"operation"`
	Inputs    []string `json:"inputs"`
}
type CommercialRow struct {
	Key        string             `json:"key"`
	Label      string             `json:"label"`
	Unit       string             `json:"unit"`
	Value      *string            `json:"value,omitempty"`
	Formula    *CommercialFormula `json:"formula,omitempty"`
	Source     string             `json:"source,omitempty"`
	Assumption string             `json:"assumption,omitempty"`
}
type CommercialTarget struct {
	NodeID    string `json:"node_id,omitempty"`
	Row       string `json:"row"`
	Path      string `json:"path"`
	Precision *int   `json:"precision,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Suffix    string `json:"suffix,omitempty"`
	Numeric   bool   `json:"numeric,omitempty"`
	Scale     string `json:"scale,omitempty"`
}
type CommercialModel struct {
	Rows        []CommercialRow    `json:"rows"`
	Precision   int                `json:"precision"`
	Rounding    string             `json:"rounding"`
	Assumptions []string           `json:"assumptions"`
	Targets     []CommercialTarget `json:"targets,omitempty"`
}
type CommercialResult struct {
	Key     string `json:"key"`
	Unit    string `json:"unit"`
	Exact   string `json:"exact"`
	Rounded string `json:"rounded"`
}
type CommercialSpec struct {
	Type         string          `json:"type,omitempty"`
	X            float64         `json:"x,omitempty"`
	Y            float64         `json:"y,omitempty"`
	W            float64         `json:"w,omitempty"`
	H            float64         `json:"h,omitempty"`
	Model        CommercialModel `json:"model"`
	Presentation map[string]any  `json:"presentation"`
	NodeID       string          `json:"node_id,omitempty"`
	Group        []string        `json:"group,omitempty"`
}

var commercialDecimal = regexp.MustCompile(`^[+-]?[0-9]{1,18}(\.[0-9]{1,12})?$`)

func ParseCommercialDecimal(v string) (*big.Rat, error) {
	if !commercialDecimal.MatchString(v) {
		return nil, fmt.Errorf("decimal requires a string with <=18 integer and <=12 fractional digits (no exponent)")
	}
	r, ok := new(big.Rat).SetString(v)
	if !ok {
		return nil, fmt.Errorf("invalid decimal")
	}
	return r, nil
}
func RoundCommercial(r *big.Rat, places int, mode string) (string, error) {
	if places < 0 || places > 8 || (mode != "half_even" && mode != "half_away") {
		return "", fmt.Errorf("rounding requires precision 0..8 and half_even/half_away")
	}
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	num := new(big.Int).Mul(r.Num(), multiplier)
	negative := num.Sign() < 0
	num.Abs(num)
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(num, r.Denom(), rem)
	cmp := new(big.Int).Lsh(rem, 1).Cmp(r.Denom())
	if cmp > 0 || cmp == 0 && (mode == "half_away" || q.Bit(0) == 1) {
		q.Add(q, big.NewInt(1))
	}
	digits := q.String()
	if places > 0 {
		for len(digits) <= places {
			digits = "0" + digits
		}
		digits = digits[:len(digits)-places] + "." + digits[len(digits)-places:]
	}
	if negative && q.Sign() != 0 {
		digits = "-" + digits
	}
	return digits, nil
}
func EvaluateCommercial(m CommercialModel) ([]CommercialResult, error) {
	if len(m.Rows) < 1 || len(m.Rows) > 100 || len(m.Assumptions) > 100 || len(m.Targets) > 500 {
		return nil, fmt.Errorf("commercial model requires 1..100 rows and <=100 assumptions")
	}
	if _, e := RoundCommercial(new(big.Rat), m.Precision, m.Rounding); e != nil {
		return nil, e
	}
	for _, a := range m.Assumptions {
		if strings.TrimSpace(a) == "" || len(a) > 4096 {
			return nil, fmt.Errorf("commercial assumptions require nonblank text within 4096 bytes")
		}
	}
	rows := map[string]CommercialRow{}
	for _, r := range m.Rows {
		if !validPartKey(r.Key) || rows[r.Key].Key != "" || strings.TrimSpace(r.Label) == "" || strings.TrimSpace(r.Unit) == "" || len(r.Unit) > 128 || len(r.Label) > 4096 || len(r.Source) > 4096 || len(r.Assumption) > 4096 {
			return nil, fmt.Errorf("commercial rows require unique stable keys, labels and units")
		}
		if (r.Value == nil) == (r.Formula == nil) {
			return nil, fmt.Errorf("commercial row %s requires exactly one decimal value or formula", r.Key)
		}
		if r.Value != nil && strings.TrimSpace(r.Source) == "" && strings.TrimSpace(r.Assumption) == "" {
			return nil, fmt.Errorf("commercial input %s requires source or explicit assumption", r.Key)
		}
		rows[r.Key] = r
	}
	cache := map[string]*big.Rat{}
	active := map[string]bool{}
	var eval func(string) (*big.Rat, error)
	eval = func(key string) (*big.Rat, error) {
		if v := cache[key]; v != nil {
			return new(big.Rat).Set(v), nil
		}
		r, ok := rows[key]
		if !ok {
			return nil, fmt.Errorf("unknown formula input %s", key)
		}
		if active[key] {
			return nil, fmt.Errorf("commercial formula cycle at %s", key)
		}
		active[key] = true
		defer delete(active, key)
		var value *big.Rat
		var e error
		if r.Value != nil {
			value, e = ParseCommercialDecimal(*r.Value)
		} else {
			f := r.Formula
			if len(f.Inputs) < 1 || len(f.Inputs) > 100 {
				return nil, fmt.Errorf("formula requires 1..100 keyed inputs")
			}
			vs := []*big.Rat{}
			for _, k := range f.Inputs {
				v, err := eval(k)
				if err != nil {
					return nil, err
				}
				vs = append(vs, v)
			}
			value = new(big.Rat)
			same := func() bool {
				for _, k := range f.Inputs {
					if rows[k].Unit != r.Unit {
						return false
					}
				}
				return true
			}
			switch f.Operation {
			case "sum":
				if !same() {
					return nil, fmt.Errorf("sum requires identical input/output units")
				}
				for _, v := range vs {
					value.Add(value, v)
				}
			case "subtract":
				if len(vs) != 2 || !same() {
					return nil, fmt.Errorf("subtract requires two inputs with identical units")
				}
				value.Sub(vs[0], vs[1])
			case "multiply", "convert":
				if len(vs) != 2 {
					return nil, fmt.Errorf("multiply/convert requires two inputs")
				}
				a, b := rows[f.Inputs[0]].Unit, rows[f.Inputs[1]].Unit
				if f.Operation == "multiply" && !(a == r.Unit && b == "ratio" || b == r.Unit && a == "ratio") {
					return nil, fmt.Errorf("multiply requires one ratio and one output-unit input; use explicit convert assumption for dimensional conversion")
				}
				if f.Operation == "convert" && strings.TrimSpace(r.Assumption) == "" {
					return nil, fmt.Errorf("conversion requires an explicit assumption")
				}
				value.Mul(vs[0], vs[1])
			case "divide":
				if len(vs) != 2 || vs[1].Sign() == 0 {
					return nil, fmt.Errorf("divide requires two inputs and nonzero denominator")
				}
				a, b := rows[f.Inputs[0]].Unit, rows[f.Inputs[1]].Unit
				if !(a == b && r.Unit == "ratio" || a == r.Unit && b == "ratio") && strings.TrimSpace(r.Assumption) == "" {
					return nil, fmt.Errorf("derived-unit division requires an explicit dimensional assumption")
				}
				value.Quo(vs[0], vs[1])
			case "roi":
				if len(vs) != 2 || vs[1].Sign() <= 0 || rows[f.Inputs[0]].Unit != rows[f.Inputs[1]].Unit || r.Unit != "ratio" {
					return nil, fmt.Errorf("ROI requires benefit, positive investment in same units and ratio output")
				}
				value.Quo(new(big.Rat).Sub(vs[0], vs[1]), vs[1])
			case "payback":
				if len(vs) != 2 || vs[0].Sign() < 0 || vs[1].Sign() <= 0 || strings.TrimSpace(r.Assumption) == "" {
					return nil, fmt.Errorf("payback requires nonnegative investment, positive net annual benefit and explicit period/unit assumption")
				}
				period := strings.TrimSuffix(r.Unit, "s")
				if period != "year" && period != "month" && period != "quarter" && period != "week" && period != "day" && period != "period" {
					return nil, fmt.Errorf("payback output unit must declare years, months, quarters, weeks, days or periods")
				}
				investmentUnit, cashUnit := rows[f.Inputs[0]].Unit, rows[f.Inputs[1]].Unit
				if cashUnit != investmentUnit && cashUnit != investmentUnit+"/"+period && cashUnit != investmentUnit+"/"+r.Unit {
					return nil, fmt.Errorf("payback investment and periodic net benefit require compatible base units")
				}
				value.Quo(vs[0], vs[1])
			case "npv":
				if len(vs) < 2 || rows[f.Inputs[0]].Unit != "ratio" || vs[0].Cmp(big.NewRat(-1, 1)) <= 0 {
					return nil, fmt.Errorf("NPV requires discount ratio >-1 then cash flows in periods 0..N")
				}
				discount := new(big.Rat).Add(big.NewRat(1, 1), vs[0])
				power := big.NewRat(1, 1)
				for i, v := range vs[1:] {
					if rows[f.Inputs[i+1]].Unit != r.Unit {
						return nil, fmt.Errorf("NPV cash flows must share output unit")
					}
					value.Add(value, new(big.Rat).Quo(v, power))
					power.Mul(power, discount)
				}
			default:
				return nil, fmt.Errorf("unknown commercial formula operation %s", f.Operation)
			}
		}
		if e != nil {
			return nil, e
		}
		if value.Num().BitLen() > 4096 || value.Denom().BitLen() > 4096 || new(big.Rat).Abs(value).Cmp(new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))) > 0 {
			return nil, fmt.Errorf("commercial calculation overflow at %s", key)
		}
		cache[key] = new(big.Rat).Set(value)
		return value, nil
	}
	out := []CommercialResult{}
	for _, r := range m.Rows {
		v, e := eval(r.Key)
		if e != nil {
			return nil, e
		}
		rounded, e := RoundCommercial(v, m.Precision, m.Rounding)
		if e != nil {
			return nil, e
		}
		out = append(out, CommercialResult{r.Key, r.Unit, v.RatString(), rounded})
	}
	return out, nil
}
func commercialSet(root any, path string, value any) error {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "//") {
		return fmt.Errorf("commercial target requires exact JSON pointer")
	}
	parts := strings.Split(path[1:], "/")
	parent := root
	for i, raw := range parts {
		part := strings.ReplaceAll(strings.ReplaceAll(raw, "~1", "/"), "~0", "~")
		last := i == len(parts)-1
		switch v := parent.(type) {
		case map[string]any:
			next, ok := v[part]
			if !ok {
				return fmt.Errorf("commercial target path does not exist: %s", path)
			}
			if last {
				v[part] = value
				return nil
			}
			parent = next
		case []any:
			j, e := strconv.Atoi(part)
			if e != nil || j < 0 || j >= len(v) {
				return fmt.Errorf("commercial target index invalid")
			}
			if last {
				v[j] = value
				return nil
			}
			parent = v[j]
		default:
			return fmt.Errorf("commercial target path crosses scalar")
		}
	}
	return fmt.Errorf("invalid target")
}

// MaterializeCommercialPresentation retains the selected source shape and only
// updates declared targets. Its measured planner remains authoritative.
func MaterializeCommercialPresentation(s CommercialSpec) (map[string]any, []CommercialResult, error) {
	results, e := EvaluateCommercial(s.Model)
	if e != nil {
		return nil, nil, e
	}
	var p map[string]any
	b, e := json.Marshal(s.Presentation)
	if e != nil {
		return nil, nil, e
	}
	if e = json.Unmarshal(b, &p); e != nil {
		return nil, nil, e
	}
	kind, _ := p["type"].(string)
	if kind == "" || kind == "commercial" {
		return nil, nil, fmt.Errorf("commercial presentation requires non-commercial source type")
	}
	for _, field := range []string{"id", "x", "y", "w", "h"} {
		if _, ok := p[field]; ok {
			return nil, nil, fmt.Errorf("presentation geometry is owned by outer component placement")
		}
	}
	seen := map[string]bool{}
	byKey := map[string]CommercialResult{}
	for _, r := range results {
		byKey[r.Key] = r
	}
	for _, t := range s.Model.Targets {
		if t.NodeID != "" && t.NodeID != s.NodeID {
			member := false
			for _, id := range s.Group {
				if id == t.NodeID {
					member = true
				}
			}
			if !member {
				return nil, nil, fmt.Errorf("target refers to undeclared calculation group node %s", t.NodeID)
			}
			continue
		}
		r, ok := byKey[t.Row]
		if !ok {
			return nil, nil, fmt.Errorf("commercial target references unknown row")
		}
		if seen[t.Path] {
			return nil, nil, fmt.Errorf("duplicate commercial target path")
		}
		seen[t.Path] = true
		for _, part := range strings.Split(t.Path, "/") {
			if part == "type" || part == "x" || part == "y" || part == "w" || part == "h" || part == "points" || part == "_source_geometry" {
				return nil, nil, fmt.Errorf("commercial target cannot overwrite topology/geometry")
			}
		}
		rat, _ := new(big.Rat).SetString(r.Exact)
		if t.Scale != "" {
			scale, e := ParseCommercialDecimal(t.Scale)
			if e != nil || scale.Sign() == 0 {
				return nil, nil, fmt.Errorf("target scale requires nonzero decimal divisor")
			}
			rat.Quo(rat, scale)
		}
		precision := s.Model.Precision
		if t.Precision != nil {
			precision = *t.Precision
		}
		value, e := RoundCommercial(rat, precision, s.Model.Rounding)
		if e != nil {
			return nil, nil, e
		}
		var replacement any = t.Prefix + value + t.Suffix
		if t.Numeric {
			if t.Prefix != "" || t.Suffix != "" {
				return nil, nil, fmt.Errorf("numeric target refuses affixes")
			}
			// Numeric targets are renderer facts, not formatted displays. Keep
			// the scaled exact rational until the float64 chart boundary; display
			// precision must never turn 10.8 into an authored workbook value 11.
			if rat.Cmp(big.NewRat(-1e9, 1)) < 0 || rat.Cmp(big.NewRat(1e9, 1)) > 0 {
				return nil, nil, fmt.Errorf("numeric presentation target outside renderer bound")
			}
			f, _ := rat.Float64()
			if math.IsNaN(f) || math.IsInf(f, 0) || math.Abs(f) > 1e9 || f == 0 && rat.Sign() != 0 {
				return nil, nil, fmt.Errorf("numeric presentation target outside renderer bound")
			}
			replacement = f
		}
		if e = commercialSet(p, t.Path, replacement); e != nil {
			return nil, nil, e
		}
	}
	return p, results, nil
}
func (r *renderer) planCommercialScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type != "commercial" {
		return nil, false, nil
	}
	var s CommercialSpec
	if e := sceneDecode(raw, &s); e != nil {
		return nil, true, e
	}
	p, _, e := MaterializeCommercialPresentation(s)
	if e != nil {
		return nil, true, e
	}
	kind := p["type"].(string)
	delete(p, "type")
	node, e := ComposeSceneNode(kind, p, Rect{s.X, s.Y, s.W, s.H})
	if e != nil {
		return nil, true, e
	}
	ctx.Path = strings.TrimRight(ctx.Path, "/") + "/presentation"
	plan, e := r.planSceneNode(id, node, ctx)
	return plan, true, e
}
