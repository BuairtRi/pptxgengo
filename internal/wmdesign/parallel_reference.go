package wmdesign

// ParallelReference assembles the three independently authored slice fixtures
// into one deck for shared native review. It does not introduce new template bindings.
func ParallelReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	for _, family := range []struct {
		prefix   string
		document Document
	}{
		{"rows", CardRowReference(year)},
		{"data", DataMetricReference(year)},
		{"rich", RichReference(year)},
	} {
		for _, slide := range family.document.Slides {
			slide.ID = family.prefix + "." + slide.ID
			d.Slides = append(d.Slides, slide)
		}
	}
	return d
}
