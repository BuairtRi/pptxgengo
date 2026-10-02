package wmdesign

import "fmt"

// ComponentReference is synthetic fixture content with bounded native variants.
// Source example copy is fictional, not an assertion about a real engagement.
func ComponentReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	slide := func(id, title string, nodes ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: id, Eyebrow: "WMDS component reference", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact"}, Nodes: nodes})
	}
	tb := func(id string, start, span int, y, h float64, v TextBlockSpec) Node {
		return Node{ID: id, Kind: "textblock", Grid: "columns", Start: start, Span: span, Rect: Rect{Y: y, H: h}, TextBlock: &v}
	}
	card := func(id string, start, span int, y, h float64, surface string, v CardSpec) Node {
		return Node{ID: id, Kind: "card", Grid: "columns", Start: start, Span: span, Rect: Rect{Y: y, H: h}, Surface: surface, Card: &v}
	}
	para := func(k, t string) BodyBlock { return BodyBlock{Key: k, Paragraph: t} }
	bullets := func(items ...string) BodyBlock {
		b := BodyBlock{Key: "list"}
		for i, t := range items {
			b.Bullets = append(b.Bullets, BulletItem{fmt.Sprintf("item-%d", i+1), t})
		}
		return b
	}
	slide("textblock-variants", "Text blocks",
		tb("title-body", 1, 4, 144, 0, TextBlockSpec{Title: "Late reconciliations", Body: "Balance-sheet accounts are reconciled after close instead of during the month."}),
		tb("label-title-body", 5, 4, 144, 0, TextBlockSpec{Label: "Root cause 2", Title: "Manual intercompany", Body: "Eleven entities match balances by email and spreadsheet."}),
		tb("body-only", 9, 4, 144, 0, TextBlockSpec{Body: "Controllers see most entries for the first time on day 9, so errors are found when they are most expensive to fix."}))
	slide("textblock-widths", "Text block widths",
		tb("three-columns", 1, 3, 144, 0, TextBlockSpec{Label: "3 columns", Title: "Close calendar", Body: "A shared calendar shows each review and its owner."}),
		tb("six-columns", 4, 6, 144, 0, TextBlockSpec{Label: "6 columns", Title: "A shared view of the close", Body: "The team sees reconciliations and intercompany balances in one place. Each exception has a named owner."}),
		tb("eight-columns", 1, 8, 342, 0, TextBlockSpec{Label: "8 columns", Title: "Reading width", Body: "Running text stops at eight columns. A shorter line supports reading while preserving space for adjacent content."}))
	slide("card-basic", "Text, bullet and numbered cards",
		card("text-card", 1, 4, 144, 234, "subtle", CardSpec{Label: "Customer", Title: "Every signal in one place", TitleInk: "emphasis", Body: []BodyBlock{para("copy", "Tickets, calls, reviews and sales notes, tagged to the accounts they affect.")}}),
		card("bullet-card", 5, 4, 144, 234, "subtle", CardSpec{Title: "What changes", TitleInk: "emphasis", Body: []BodyBlock{bullets("Review moves to the start", "Rules check every entry", "Exceptions routed same day")}}),
		card("numbered-card", 9, 4, 144, 234, "subtle", CardSpec{InlineNumber: "01", Title: "Diagnose", TitleInk: "emphasis", Body: []BodyBlock{para("copy", "Map the close calendar and rank every delay by days lost.")}}))
	slide("card-paragraphs", "Cards with multiple paragraphs",
		card("paragraph-card", 1, 6, 144, 270, "subtle", CardSpec{Title: "Why the close takes 12 days", Body: []BodyBlock{para("waiting", "Most of the time is spent waiting. Reconciliations start after period end, and intercompany balances are matched by email."), para("review", "Review happens last, so errors surface on day 9, when they are most expensive to fix.")}}),
		card("heading-card", 7, 6, 144, 270, "outline", CardSpec{Label: "Wave one", Title: "Journal review", TitleStyle: "heading", Body: []BodyBlock{para("first", "Rules check entries at posting. Controllers review exceptions throughout the month."), para("second", "The close calendar tracks outstanding decisions and the next review date.")}}))
	for page, surfaces := range [][]string{{"light", "subtle", "strong", "outline"}, {"inverse", "deep", "callout"}} {
		ns := []Node{}
		span := 3
		if page == 1 {
			span = 4
		}
		for i, surf := range surfaces {
			ns = append(ns, card("surface-"+surf, 1+i*span, span, 144, 234, surf, CardSpec{Label: surf, Title: "Shared context", Body: []BodyBlock{para("copy", "Every exception has an owner and a review date.")}}))
		}
		slide(fmt.Sprintf("card-surfaces-%d", page+1), "Card surfaces", ns...)
	}
	ns := []Node{}
	for i, pair := range [][2]string{{"inverse", "subtle"}, {"subtle", "outline"}, {"callout", "outline"}, {"deep", "inverse"}} {
		ns = append(ns, card(fmt.Sprintf("band-%d", i+1), 1+i*3, 3, 144, 234, pair[1], CardSpec{Label: fmt.Sprintf("Wave %d", i+1), Title: []string{"Reconcile", "Match", "Review", "Report"}[i], Band: &CardBand{Surface: pair[0]}, Body: []BodyBlock{para("copy", []string{"Five entities, 12 weeks.", "Intercompany for all 11 entities.", "Rules-based review at posting.", "Consolidation and reporting."}[i])}}))
	}
	slide("card-title-bands", "Title bands", ns...)
	ns = []Node{}
	for i, side := range []string{"left", "top", "right", "bottom"} {
		weight := 3.0
		if i%2 == 1 {
			weight = 6
		}
		ns = append(ns, card("edge-"+side, 1+i*3, 3, 144, 234, "outline", CardSpec{Label: fmt.Sprintf("%s / %.0f pt", side, weight), Title: "Close review", Edge: &CardEdge{Side: side, Weight: weight, Ink: "emphasis"}, Body: []BodyBlock{para("copy", "The team reviews open exceptions before period end.")}}))
	}
	slide("card-edges", "Accent edges", ns...)
	slide("card-density", "Explicit body density",
		card("standard", 1, 6, 144, 252, "subtle", CardSpec{Label: "Standard 14 / 21", Title: "Review checklist", Body: []BodyBlock{bullets("Assign an owner to each exception", "Confirm the next review date", "Record the decision in the shared close calendar")}}),
		card("dense", 7, 6, 144, 252, "subtle", CardSpec{Label: "Dense 12 / 18", Title: "Reference checklist", BodySize: "small", Dense: true, Padding: 24, Body: []BodyBlock{bullets("Assign an owner to each exception", "Confirm the next review date", "Record the decision in the shared close calendar"), para("note", "Dense text is an explicit reference variant. It does not activate automatically when content grows.")}}))
	slide("card-autoheight", "Measured card heights",
		card("short", 1, 6, 144, 0, "outline", CardSpec{Title: "An assigned owner", Body: []BodyBlock{para("copy", "Each exception has a named owner.")}}),
		card("long", 7, 6, 144, 0, "outline", CardSpec{Label: "Review process", Title: "A shared close calendar", Body: []BodyBlock{para("copy", "Each exception has a named owner. The calendar records the next review date and the decision needed before period end."), para("next", "Controllers review outstanding items throughout the month, with additional detail available in the source records.")}}))
	slide("textblock-breaks", "Authored paragraph breaks",
		tb("hard-breaks", 1, 6, 144, 0, TextBlockSpec{Title: "One title\nTwo authored lines", Body: "First paragraph.\n\nA second paragraph follows a blank line.\n"}),
		tb("leading-blank", 7, 6, 144, 0, TextBlockSpec{Label: "Blank paragraphs", Title: "Preserved spacing", Body: "\nAn authored leading blank remains in the editable text.\n\nThe final paragraph keeps the same body style."}))
	sl := SlideSpec{ID: "card-rail", Eyebrow: "WMDS component reference", Title: "Cards in frame zones", Frame: FrameRequest{Rail: "left", Footer: "compact"}}
	n := card("rail-card", 1, 3, 144, 216, "outline", CardSpec{Label: "Rail", Title: "Review owner", Body: []BodyBlock{para("copy", "The owner maintains the close calendar.")}})
	n.Scope = "rail"
	sl.Nodes = append(sl.Nodes, n)
	sl.Nodes = append(sl.Nodes, card("main-card", 5, 6, 144, 234, "subtle", CardSpec{InlineNumber: "02", Title: "Review the exceptions", TitleStyle: "heading", Band: &CardBand{Surface: "inverse"}, Body: []BodyBlock{para("copy", "Controllers review exceptions throughout the month. Each decision has an owner and a date.")}}))
	d.Slides = append(d.Slides, sl)
	return d
}
