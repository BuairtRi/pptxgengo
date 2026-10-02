package wmdesign

import "encoding/json"

// DataMetricReference contains fictional data and demonstrates only bounded,
// source-derived standalone metric anatomy. It does not assert qualification.
func DataMetricReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	metric := func(id string, start, span int, y, h float64, m DataMetricSpec) Node {
		return Node{ID: id, Kind: "metric", Grid: "columns", Start: start, Span: span, Rect: Rect{Y: y, H: h}, DataMetric: &m}
	}
	slide := func(id, title string, nodes ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: "data-metric-" + id, Eyebrow: "WMDS standalone metric reference", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact"}, Nodes: nodes})
	}
	formatted := func(raw, kind, label string) DataMetricSpec {
		return DataMetricSpec{Format: &NumberFormatSpec{Value: json.RawMessage(raw), Kind: kind}, Label: label}
	}
	slide("primary", "Standalone values and comparisons",
		metric("data-primary", 1, 3, 144, 0, DataMetricSpec{Value: "58%", Label: "fewer manual journal entries"}),
		metric("data-change", 4, 3, 144, 0, DataMetricSpec{Value: "5 days", Label: "to close", Change: &DataMetricChange{Direction: "down", Text: "7 days vs FY25"}}),
		metric("data-comparisons", 7, 6, 144, 0, DataMetricSpec{Value: "$4.2M", Label: "run-rate savings from the first two waves", Secondary: []KeyedMetricValue{{Key: "cost", Value: "18%", Label: "lower cost to serve"}, {Key: "payback", Value: "14 mo", Label: "to payback"}}}))
	fte := formatted("0.5", "number", "full-time equivalent capacity")
	fte.Format.Unit = "FTE"
	slide("formats", "Formats computed from decimal inputs",
		metric("data-currency", 1, 3, 144, 0, formatted("4250000", "currency", "run-rate savings")),
		metric("data-percent", 4, 3, 144, 0, formatted("0.065", "percent", "lower cost to serve")),
		metric("data-number", 7, 3, 144, 0, formatted("1240", "number", "manual entries per month")),
		metric("data-unit", 10, 3, 144, 0, fte))
	rangePct := formatted("[0.15,0.20]", "range", "lower cost to serve")
	rangePct.Format.Of = "percent"
	rangeCurrency := formatted("[1900000,4200000]", "range", "potential run-rate savings")
	rangeCurrency.Format.Of = "currency"
	negative := formatted("-1900000", "currency", "illustrative adverse variance")
	slide("ranges", "Ranges and negative values",
		metric("data-percent-range", 1, 4, 144, 0, rangePct),
		metric("data-currency-range", 5, 4, 144, 0, rangeCurrency),
		metric("data-negative", 9, 4, 144, 0, negative))
	nodes := []Node{}
	for i, status := range []string{"on", "risk", "off"} {
		nodes = append(nodes, metric("data-status-"+status, 1+i*4, 4, 144, 0, DataMetricSpec{Value: []string{"4 days", "6 days", "8 days"}[i], Label: "to close today", Target: "Target 5 days", Status: status, Source: "Illustrative planning scenario"}))
	}
	slide("targets", "Authored targets and status", nodes...)
	delta := formatted("-7", "delta", "fewer days to close")
	delta.Format.Unit = "days"
	points := formatted("0.04", "delta", "posting validation change")
	points.Format.Unit = "pts"
	up := DataMetricSpec{Value: "92%", Label: "entries checked at posting", Change: &DataMetricChange{Direction: "up", Text: "Up 8 points"}, Secondary: []KeyedMetricValue{{Key: "live", Value: "11", Label: "entities live"}}, Source: "Illustrative operating model"}
	slide("changes", "Signed deltas and native direction marks",
		metric("data-negative-delta", 1, 3, 144, 0, delta),
		metric("data-points-delta", 4, 3, 144, 0, points),
		metric("data-up", 7, 6, 144, 0, up))
	rail := metric("data-rail", 1, 3, 144, 0, DataMetricSpec{Value: "92%", Label: "entries checked at posting", Status: "on"})
	rail.Scope = "rail"
	main := metric("data-main", 5, 8, 144, 0, DataMetricSpec{Value: "$4.2M", Label: "run-rate savings", Secondary: []KeyedMetricValue{{Key: "payback", Value: "14 mo", Label: "to payback"}, {Key: "cost", Value: "18%", Label: "lower cost to serve"}}, Source: "Fictional example. Secondary keys retained after reorder."})
	d.Slides = append(d.Slides, SlideSpec{ID: "data-metric-rail", Eyebrow: "WMDS standalone metric reference", Title: "Metrics in frame zones", Frame: FrameRequest{Rail: "left", Footer: "compact"}, Nodes: []Node{rail, main}})
	return d
}
