package wmdesign

// CardRowReference demonstrates bounded row composition using fictional copy.
func CardRowReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	paragraph := func(key, title, body string) CardRowItem {
		return CardRowItem{Key: key, Card: CardSpec{Title: title, Body: []BodyBlock{{Key: "summary", Paragraph: body}}}}
	}
	row := func(id string, start, span int, y, h float64, numbering string, items ...CardRowItem) Node {
		return Node{ID: id, Kind: "cardrow", Grid: "columns", Start: start, Span: span, Rect: Rect{Y: y, H: h}, Surface: "subtle", CardRow: &CardRowSpec{Items: items, Numbering: numbering}}
	}
	slide := func(id, title string, ns ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: "cardrow-" + id, Eyebrow: "WMDS keyed card rows", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact"}, Nodes: ns})
	}
	a := paragraph("diagnose", "Diagnose", "Map the close calendar and name the owner of each handoff.")
	b := paragraph("design", "Design", "Define the rules that check entries as they post.")
	a.Card.Body = append(a.Card.Body, BodyBlock{Key: "detail", Paragraph: "Use the same definitions across all entities."})
	b.Surface = "outline"
	slide("two", "Two equal cards, measured height", row("paired", 1, 12, 144, 0, "", a, b))
	three := []CardRowItem{
		paragraph("reconcile", "Reconcile", "Five entities in the first wave."),
		paragraph("match", "Match", "Intercompany across 11 entities."),
		paragraph("review", "Review", "Rules check entries at posting."),
	}
	three[1].Surface = "outline"
	three[2].Surface = "inverse"
	four := []CardRowItem{
		paragraph("diagnose", "Diagnose", "Rank delays by days lost."),
		paragraph("design", "Design", "Move review to the start."),
		paragraph("deploy", "Deploy", "Five entities in wave one."),
		paragraph("sustain", "Sustain", "Review close health monthly."),
	}
	slide("counts", "Three and four cards on the grid",
		row("three", 1, 12, 144, 126, "", three...),
		row("four", 1, 12, 288, 162, "", four...))
	slide("inline", "Inline numbering follows visual order", row("steps", 1, 12, 144, 180, "inline", four...))
	banded := []CardRowItem{
		paragraph("diagnose", "Diagnose", "Map the close calendar for all 14 business units."),
		paragraph("design", "Design", "Define rules that check entries as they post."),
		paragraph("deploy", "Deploy", "Roll out in two waves over 22 weeks."),
	}
	for i := range banded {
		banded[i].Surface = "outline"
		banded[i].Card.Band = &CardBand{Surface: "inverse"}
	}
	slide("bands", "Numbering in title bands", row("banded-steps", 1, 12, 144, 198, "band", banded...))
	metrics := []CardRowItem{
		{Key: "journal", Card: CardSpec{Metric: &MetricSpec{Value: "58%", Label: "fewer manual journal entries", Secondary: &MetricValue{Value: "1,240", Label: "entries per month today"}}}},
		{Key: "close", Surface: "outline", Card: CardSpec{Metric: &MetricSpec{Value: "5 days", Label: "to close", Secondary: &MetricValue{Value: "12", Label: "days today"}}}},
		{Key: "savings", Surface: "inverse", Card: CardSpec{Metric: &MetricSpec{Value: "$1.9M", Label: "annual run-rate savings", Status: "risk"}}},
		{Key: "payback", Surface: "callout", Card: CardSpec{Metric: &MetricSpec{Value: "14 mo", Label: "to payback", Secondary: &MetricValue{Value: "Q3 2027", Label: "break-even"}}}},
	}
	slide("metrics", "Metric footers share a measured row height", row("outcomes", 1, 12, 144, 0, "", metrics...))
	groups := []CardRowItem{
		{Key: "outcomes", Card: CardSpec{Label: "Wave one", MetricGroup: &MetricGroupSpec{Primary: MetricValue{Value: "$4.2M", Label: "run-rate savings"}, Secondary: []KeyedMetricValue{{Key: "cost", Value: "18%", Label: "lower cost to serve"}}}}},
		{Key: "adoption", Surface: "outline", Card: CardSpec{Label: "Adoption", Edge: &CardEdge{Side: "left", Weight: 3, Ink: "emphasis"}, MetricGroup: &MetricGroupSpec{Primary: MetricValue{Value: "92%", Label: "entries checked at posting"}, Secondary: []KeyedMetricValue{{Key: "live", Value: "11", Label: "entities live"}, {Key: "pilot", Value: "3", Label: "in pilot"}}}}},
	}
	slide("groups", "Metric groups remain editable within a row", row("grouped-outcomes", 1, 12, 144, 0, "", groups...))
	ordered := []CardRowItem{
		paragraph("diagnose", "Diagnose", "Rank delays by days lost."),
		paragraph("design", "Design", "Move review to the start."),
		paragraph("deploy", "Deploy", "Roll out in two waves."),
	}
	slide("reorder", "Keys retain identity; order sets numbering",
		row("original-order", 1, 12, 144, 126, "inline", ordered...),
		row("reordered", 1, 12, 288, 126, "inline", ordered[2], ordered[0], ordered[1]))
	return d
}
