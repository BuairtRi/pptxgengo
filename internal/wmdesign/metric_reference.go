package wmdesign

import "fmt"

// MetricReference exercises native metric-card anatomy with fictional data.
func MetricReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	card := func(id string, start, span int, y, h float64, surface string, v CardSpec) Node {
		return Node{ID: id, Kind: "card", Grid: "columns", Start: start, Span: span, Rect: Rect{Y: y, H: h}, Surface: surface, Card: &v}
	}
	slide := func(id, title string, ns ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: id, Eyebrow: "WMDS metric reference", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact"}, Nodes: ns})
	}
	single := func(value, label string) *MetricSpec { return &MetricSpec{Value: value, Label: label} }
	secondary := func(value, label string) *MetricValue { return &MetricValue{Value: value, Label: label} }
	group := func(value, label string, sm ...KeyedMetricValue) *MetricGroupSpec {
		return &MetricGroupSpec{Primary: MetricValue{Value: value, Label: label}, Secondary: sm}
	}
	slide("metric-presets", "Metric cards",
		card("journal", 1, 3, 144, 198, "subtle", CardSpec{Metric: &MetricSpec{Value: "58%", Label: "fewer manual journal entries", Secondary: secondary("1,240", "entries per month today")}}),
		card("close", 4, 3, 144, 198, "outline", CardSpec{Metric: &MetricSpec{Value: "5 days", Label: "to close", Secondary: secondary("12", "days today")}}),
		card("savings", 7, 3, 144, 198, "inverse", CardSpec{Metric: &MetricSpec{Value: "$1.9M", Label: "annual run-rate savings", Status: "risk"}}),
		card("payback", 10, 3, 144, 198, "callout", CardSpec{Metric: &MetricSpec{Value: "14 mo", Label: "to payback", Secondary: secondary("Q3 2027", "break-even")}}))
	ns := []Node{}
	for i, surf := range []string{"light", "subtle", "strong", "outline", "inverse", "deep", "callout"} {
		y := 144.0
		start := 1 + i*3
		if i >= 4 {
			y = 306
			start = 1 + (i-4)*3
		}
		ns = append(ns, card("surface-"+surf, start, 3, y, 144, surf, CardSpec{Label: surf, Metric: single("92%", "entries checked at posting")}))
	}
	slide("metric-surfaces", "Metric surfaces", ns...)
	ns = []Node{}
	for i, status := range []string{"on", "risk", "off"} {
		ns = append(ns, card("status-"+status, 1+i*4, 4, 144, 234, []string{"subtle", "inverse", "outline"}[i], CardSpec{Label: []string{"Wave one", "Wave two", "Wave three"}[i], Metric: &MetricSpec{Value: []string{"92%", "76%", "58%"}[i], Label: "entries checked at posting", Change: []string{"Up 8 points", "Up 3 points", "Down 2 points"}[i], Status: status}}))
	}
	slide("metric-status", "Status and change", ns...)
	slide("metric-widths", "Value sizes by card width",
		card("three-col", 1, 3, 144, 180, "outline", CardSpec{Label: "3 columns", Metric: single("58%", "fewer manual entries")}),
		card("four-col", 4, 4, 144, 180, "outline", CardSpec{Label: "4 columns", Metric: single("$4.2M", "run-rate savings")}),
		card("five-col", 8, 5, 144, 180, "outline", CardSpec{Label: "5 columns", Padding: 24, Metric: single("1,240", "entries per month")}))
	slide("metric-group-one", "Primary and comparison metrics",
		card("one-secondary", 1, 6, 144, 180, "subtle", CardSpec{Label: "Wave one outcomes", MetricGroup: group("$4.2M", "run-rate savings", KeyedMetricValue{"cost", "18%", "lower cost to serve"})}),
		card("two-secondaries", 7, 6, 144, 180, "outline", CardSpec{Label: "Adoption", Edge: &CardEdge{Side: "left", Weight: 3, Ink: "emphasis"}, MetricGroup: group("92%", "of entries checked at posting", KeyedMetricValue{"live", "11", "entities live"}, KeyedMetricValue{"pilot", "3", "in pilot"})}))
	slide("metric-group-two", "Groups across seven and twelve columns",
		card("wide-group", 1, 7, 144, 144, "subtle", CardSpec{MetricGroup: group("$4.2M", "run-rate savings", KeyedMetricValue{"cost", "18%", "lower cost to serve"}, KeyedMetricValue{"close", "5 days", "to close"})}),
		card("full-group", 1, 12, 306, 126, "inverse", CardSpec{MetricGroup: group("$4.2M", "annual run-rate savings", KeyedMetricValue{"close", "5 days", "to close"}, KeyedMetricValue{"cost", "18%", "lower cost to serve"})}))
	slide("metric-bands", "Metric cards with title bands",
		card("metric-band", 1, 6, 144, 288, "subtle", CardSpec{Label: "Wave one", Title: "Journal review", InlineNumber: "01", Band: &CardBand{Surface: "inverse"}, Metric: &MetricSpec{Value: "92%", Label: "entries checked at posting", Secondary: secondary("1,240", "entries per month"), Change: "Up 8 points", Status: "on"}}),
		card("group-band", 7, 6, 144, 270, "outline", CardSpec{Title: "Close performance", Band: &CardBand{Surface: "callout"}, Edge: &CardEdge{Side: "bottom", Weight: 6, Ink: "emphasis"}, MetricGroup: group("92%", "entries checked at posting", KeyedMetricValue{"live", "11", "entities live"}, KeyedMetricValue{"pilot", "3", "in pilot"})}))
	slide("metric-autoheight", "Measured heights and authored breaks",
		card("auto-short", 1, 3, 144, 0, "outline", CardSpec{Metric: single("58%", "fewer manual entries")}),
		card("auto-breaks", 4, 3, 144, 0, "subtle", CardSpec{Metric: &MetricSpec{Value: "92%", Label: "Checked entries\n\nMonthly report", Secondary: secondary("11", "entities live")}}),
		card("auto-footer", 7, 6, 144, 0, "inverse", CardSpec{Metric: &MetricSpec{Value: "$4.2M", Label: "annual run-rate savings", Secondary: secondary("14 mo", "to payback"), Change: "First two waves", Status: "risk"}}))
	sl := SlideSpec{ID: "metric-rail", Eyebrow: "WMDS metric reference", Title: "Metrics in frame zones", Frame: FrameRequest{Rail: "left", Footer: "compact"}}
	rail := card("rail-metric", 1, 3, 144, 180, "deep", CardSpec{Metric: &MetricSpec{Value: "92%", Label: "entries checked at posting", Status: "on"}})
	rail.Scope = "rail"
	sl.Nodes = []Node{rail, card("main-metrics", 5, 8, 144, 180, "subtle", CardSpec{Label: "Wave one outcomes", MetricGroup: group("$4.2M", "run-rate savings", KeyedMetricValue{"cost", "18%", "lower cost to serve"}, KeyedMetricValue{"close", "5 days", "to close"})})}
	d.Slides = append(d.Slides, sl)
	for i := range d.Slides {
		d.Slides[i].ID = fmt.Sprintf("%02d-%s", i+1, d.Slides[i].ID)
	}
	return d
}
