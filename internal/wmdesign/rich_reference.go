package wmdesign

// RichReference contains fictional copy and diagnostic splits, not engagement
// claims. Each case remains one editable text object with authored paragraphs.
func RichReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	run := func(key, text string, weight int, ink string) RichRunSpec {
		return RichRunSpec{Key: key, Text: text, Weight: weight, Ink: ink}
	}
	p := func(key string, runs ...RichRunSpec) RichParagraphSpec {
		return RichParagraphSpec{Key: key, Runs: runs}
	}
	node := func(id, style, ink string, start, span int, y float64, paragraphs ...RichParagraphSpec) Node {
		return Node{ID: id, Kind: "richtext", Style: style, Ink: ink, Grid: "columns", Start: start, Span: span, Rect: Rect{Y: y}, RichText: &RichTextSpec{Paragraphs: paragraphs}}
	}
	slide := func(id, title, surface string, nodes ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: "rich-" + id, Eyebrow: "WMDS mixed-run reference", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact", Surface: surface}, Nodes: nodes})
	}
	slide("weights", "Inline emphasis uses real font weights", "light",
		node("rich.body", "body", "primary", 1, 6, 144,
			p("copy", run("first", "The close calendar gives every exception ", 0, ""), run("emphasis", "a named owner", 600, ""), run("last", ". Controllers review open items throughout the month.", 0, ""))),
		node("rich.weights", "body", "primary", 7, 6, 144,
			p("weights", run("regular", "Regular 400. ", 400, ""), run("medium", "Medium 500. ", 500, ""), run("semibold", "Semibold 600. ", 600, ""), run("bold", "Bold 700.", 700, "")),
			p("continued", run("copy", "The paragraph preserves point size and leading when its weight changes.", 0, "secondary"))))
	slide("wrapping", "Runs wrap as one paragraph", "light",
		node("rich.wrap-narrow", "body", "primary", 1, 3, 144,
			p("copy", run("prefix", "Every exception has ", 0, ""), run("owner", "an accountable owner", 600, "emphasis"), run("suffix", " and a review date in the shared calendar.", 0, ""))),
		node("rich.wrap-wide", "lead", "primary", 4, 9, 144,
			p("copy", run("prefix", "The team reviews ", 0, ""), run("phrase", "open reconciliations and intercompany balances", 600, "emphasis"), run("suffix", " before period end. Every decision has an owner and a date.", 0, ""))))
	slide("ligatures", "Run boundaries and ligature words", "light",
		node("rich.unsplit", "body", "primary", 1, 6, 144,
			p("words", run("whole", "office affine efficient official", 0, ""))),
		node("rich.split", "body", "primary", 7, 6, 144,
			p("words", run("one", "of", 0, ""), run("two", "fice af", 0, ""), run("three", "fine ef", 0, ""), run("four", "ficient of", 0, ""), run("five", "ficial", 0, ""))),
		node("rich.color-split", "body", "primary", 1, 6, 234,
			p("words", run("one", "of", 0, ""), run("two", "fice", 0, "emphasis"), run("three", " affine efficient", 0, ""))),
		node("rich.weight-split", "body", "primary", 7, 6, 234,
			p("words", run("one", "of", 0, ""), run("two", "fice", 600, ""), run("three", " affine efficient", 0, ""))))
	slide("paragraphs", "Empty and trailing paragraphs stay editable", "light",
		node("rich.breaks", "body", "primary", 1, 6, 144,
			p("first", run("copy", "First paragraph with ", 0, ""), run("emphasis", "an owner", 600, ""), run("period", ".", 0, "")),
			p("blank"),
			p("second", run("copy", "The second paragraph follows a blank line.", 0, "")),
			p("trailing")),
		node("rich.leading", "body", "primary", 7, 6, 144,
			p("leading"),
			p("copy", run("text", "Leading space is explicit. ", 0, ""), run("strong", "Review stays visible.", 700, "")),
			p("last", run("text", "The final run can end in a different weight.", 600, ""))))
	slide("inverse", "Inline color follows the frame surface", "inverse",
		node("rich.inverse", "lead", "primary", 1, 8, 144,
			p("copy", run("prefix", "The team sees ", 0, ""), run("emphasis", "one shared close calendar", 600, "emphasis"), run("suffix", " with open decisions and the next review date.", 0, ""))),
		node("rich.mono", "number", "primary", 1, 8, 288,
			p("numbers", run("first", "12", 400, ""), run("label", " DAYS TODAY / ", 500, "secondary"), run("second", "5", 700, "emphasis"), run("last", " DAY TARGET", 600, ""))))
	slide("small", "Small copy and tracked uppercase tokens", "light",
		node("rich.small", "small", "secondary", 1, 6, 144,
			p("copy", run("prefix", "Reference copy preserves ", 0, ""), run("emphasis", "the same 12-point size", 600, "primary"), run("suffix", " while emphasizing a phrase.", 0, ""))),
		node("rich.label", "label", "primary", 7, 6, 144,
			p("label", run("prefix", "wave one / ", 0, ""), run("emphasis", "close review", 600, "emphasis"))),
		node("rich.source", "source", "secondary", 1, 8, 270,
			p("source", run("prefix", "Source: ", 600, "primary"), run("detail", "Fictional close-calendar workshop notes. All values and engagement details are illustrative.", 0, ""))))
	return d
}
