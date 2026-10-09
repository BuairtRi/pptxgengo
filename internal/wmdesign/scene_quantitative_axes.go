package wmdesign

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"math"
)

// Optional explicit axes extend the source contract without changing the frozen
// auto scales. They cannot silently clip observations or target annotations.
func jsonNumberPair(p []json.RawMessage, x, y *float64) error {
	if e := json.Unmarshal(p[0], x); e != nil {
		return e
	}
	return json.Unmarshal(p[1], y)
}
func applyQuantitativeAxes(o *pptx.ChartOptions, n sceneChartSource, scatter bool) error {
	if n.XMin == nil && n.XMax == nil && n.YMin == nil && n.YMax == nil {
		return nil
	}
	for _, v := range []*float64{n.XMin, n.XMax, n.YMin, n.YMax} {
		if v != nil && (math.IsNaN(*v) || math.IsInf(*v, 0) || math.Abs(*v) > 1e9) {
			return fmt.Errorf("scene.chart_axis_requires_finite_bound_within_1e9")
		}
	}
	if !scatter && (n.XMin != nil || n.XMax != nil) {
		return fmt.Errorf("scene.chart_x_bounds_require_scatter")
	}
	if n.YMin != nil {
		o.ValAxisMinVal = n.YMin
	}
	if n.YMax != nil {
		o.ValAxisMaxVal = n.YMax
	}
	if o.ValAxisMinVal == nil || o.ValAxisMaxVal == nil || *o.ValAxisMinVal >= *o.ValAxisMaxVal {
		return fmt.Errorf("scene.chart_y_bounds_require_min_below_max")
	}
	if scatter {
		if n.XMin != nil {
			o.CatAxisMinVal = n.XMin
		}
		if n.XMax != nil {
			o.CatAxisMaxVal = n.XMax
		}
		if *o.CatAxisMinVal >= *o.CatAxisMaxVal {
			return fmt.Errorf("scene.chart_x_bounds_require_min_below_max")
		}
		for _, p := range n.Points {
			if len(p) < 2 {
				continue
			}
			var x, y float64
			if e := jsonNumberPair(p, &x, &y); e != nil {
				return e
			}
			if x < *o.CatAxisMinVal || x > *o.CatAxisMaxVal || y < *o.ValAxisMinVal || y > *o.ValAxisMaxVal {
				return fmt.Errorf("scene.chart_explicit_axes_would_clip_point")
			}
		}
		o.CatAxisMajorUnit = sceneChartFloat((*o.CatAxisMaxVal - *o.CatAxisMinVal) / 4)
	} else {
		for i := range n.Categories {
			total := 0.0
			for _, s := range n.Series {
				if i >= len(s.Values) || s.Values[i] == nil {
					continue
				}
				v := *s.Values[i]
				if n.Mode == "stacked" {
					total += v
				} else if n.Mode != "percent" && (v < *o.ValAxisMinVal || v > *o.ValAxisMaxVal) {
					return fmt.Errorf("scene.chart_explicit_axes_would_clip_value")
				}
			}
			if n.Mode == "stacked" && (total < *o.ValAxisMinVal || total > *o.ValAxisMaxVal) {
				return fmt.Errorf("scene.chart_explicit_axes_would_clip_stack")
			}
		}
	}
	if n.Target != nil && (n.Target.Value < *o.ValAxisMinVal || n.Target.Value > *o.ValAxisMaxVal) {
		return fmt.Errorf("scene.chart_explicit_axes_would_clip_target")
	}
	o.ValAxisMajorUnit = sceneChartFloat((*o.ValAxisMaxVal - *o.ValAxisMinVal) / 4)
	return nil
}
