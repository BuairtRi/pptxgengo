package wmdesign

import "fmt"

// Reference is engineering fixture content, not a reusable component/template
// implementation or a client narrative. All visible labels remain editable.
func Reference(year int) Document {
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	text := func(id, style, value string, x, y, w, h float64) Node {
		return Node{ID: id, Kind: "text", Style: style, Text: value, Rect: Rect{x, y, w, h}}
	}
	type sample struct {
		style, value string
		y, h         float64
	}
	sets := [][]sample{
		{{"display", "Shape what comes next", 144, 72}, {"title", "A clearer way forward", 270, 54}},
		{{"heading", "A consistent foundation", 144, 36}, {"subhead", "Built around a shared grid", 216, 30}, {"lead", "Each style has an explicit role.", 270, 30}, {"body", "The body style supports readable, editable slide content.", 324, 30}, {"small", "Small text belongs in designated dense zones.", 378, 24}, {"source", "Source: synthetic typography reference", 432, 12}},
		{{"stat", "42.8%", 144, 72}, {"stat-sm", "$1.2M", 270, 54}, {"number", "01 / 02 / 03", 378, 30}},
		{{"eyebrow", "Foundation reference", 144, 18}, {"label", "Grid and type", 252, 18}, {"footer", "Footer style / legal line", 360, 18}},
	}
	for i, set := range sets {
		sl := SlideSpec{ID: fmt.Sprintf("type-%d", i+1), Eyebrow: "Typography reference", Title: []string{"Display and title", "Reading styles", "Numbers and metrics", "Mono labels and footer"}[i], Frame: FrameRequest{Rail: "none", Footer: "compact"}}
		for _, sm := range set {
			sl.Nodes = append(sl.Nodes, text("sample."+sm.style, sm.style, sm.value, 57, sm.y, 846, sm.h))
			sl.Nodes = append(sl.Nodes, text("meta."+sm.style, "label", sm.style, 57, sm.y-18, 846, 12))
		}
		doc.Slides = append(doc.Slides, sl)
	}
	sl := SlideSpec{ID: "grid-columns", Eyebrow: "Grid reference", Title: "Twelve columns, one shared rhythm", Frame: FrameRequest{Rail: "none", Footer: "compact"}}
	for i := 1; i <= 12; i++ {
		sl.Nodes = append(sl.Nodes, Node{ID: fmt.Sprintf("column-%d", i), Kind: "box", Grid: "columns", Start: i, Span: 1, Surface: "subtle", Rect: Rect{Y: 144, H: 54}}, Node{ID: fmt.Sprintf("column-label-%d", i), Kind: "text", Style: "number", Text: fmt.Sprint(i), Grid: "columns", Start: i, Span: 1, Align: "center", Rect: Rect{Y: 162, H: 30}})
	}
	for _, b := range []struct {
		start, span int
		y           float64
	}{{1, 6, 234}, {7, 6, 234}, {1, 4, 342}, {5, 4, 342}, {9, 4, 342}} {
		id := fmt.Sprintf("span-%d-%d", b.start, b.span)
		sl.Nodes = append(sl.Nodes, Node{ID: id, Kind: "box", Grid: "columns", Start: b.start, Span: b.span, Surface: "subtle", Rect: Rect{Y: b.y, H: 72}}, Node{ID: id + ".label", Kind: "text", Style: "body", Text: fmt.Sprintf("%d-column span", b.span), Grid: "columns", Start: b.start, Span: b.span, Align: "center", Rect: Rect{Y: b.y + 18, H: 30}})
	}
	doc.Slides = append(doc.Slides, sl)
	sl = SlideSpec{ID: "grid-five-up", Eyebrow: "Grid reference", Title: "Five equal columns across the canvas", Frame: FrameRequest{Rail: "none", Footer: "compact"}}
	for i := 1; i <= 5; i++ {
		x := 57 + float64(i-1)*172.8
		sl.Nodes = append(sl.Nodes, Node{ID: fmt.Sprintf("five-%d", i), Kind: "box", Grid: "five-up", Start: i, Span: 1, Surface: "subtle", Rect: Rect{Y: 144, H: 216}}, text(fmt.Sprintf("five-title-%d", i), "subhead", fmt.Sprintf("Column %d", i), x+12, 162, 130.8, 30), text(fmt.Sprintf("five-copy-%d", i), "small", "154.8 pt wide\n18 pt gutter\n18 pt rhythm", x+12, 216, 130.8, 72))
	}
	sl.Nodes = append(sl.Nodes, text("five.rule", "body", "Use five-up only on a full-width frame without a rail.", 57, 396, 846, 30))
	doc.Slides = append(doc.Slides, sl)
	for _, rail := range []string{"none", "nav", "left", "right"} {
		for _, footer := range []string{"compact", "tall"} {
			q := FrameRequest{Rail: rail, Footer: footer}
			if rail == "nav" {
				q.Nav = []NavTab{{"identify", "Identify"}, {"define", "Define"}, {"build", "Build"}, {"optimize", "Optimize"}}
				q.Active = "define"
			}
			title := fmt.Sprintf("%s rail / %s footer", rail, footer)
			sl := SlideSpec{ID: "frame-" + rail + "-" + footer, Eyebrow: "Frame reference", Title: title, Frame: q}
			x, w := 57.0, 846.0
			if rail == "left" {
				x, w = 345, 558
			}
			if rail == "right" {
				w = 558
			}
			sl.Nodes = append(sl.Nodes, text("main-heading", "heading", "Main content zone", x, 144, w, 36), text("main-copy", "body", "The frame sets the available space. Content stays inside its assigned zone.", x, 216, w, 72), text("main-note", "small", "Logo, legal line, whiteboard field and live page number are required chrome.", x, 324, w, 54))
			if rail == "left" || rail == "right" {
				rx := 57.0
				if rail == "right" {
					rx = 705
				}
				n := text("rail-label", "eyebrow", "Rail content", rx, 144, 198, 18)
				n.Scope = "rail"
				sl.Nodes = append(sl.Nodes, n)
				n = text("rail-copy", "small", "A distinct zone with an explicit surface and its own content capacity.", rx, 198, 198, 90)
				n.Scope = "rail"
				sl.Nodes = append(sl.Nodes, n)
			}
			doc.Slides = append(doc.Slides, sl)
		}
	}
	doc.Slides = append(doc.Slides,
		SlideSpec{ID: "two-line-source", Eyebrow: "Frame features", Title: "A two-line title keeps its body\nwithin a fixed content zone", Frame: FrameRequest{Rail: "none", Footer: "tall", TitleLines: 2, SourceLines: 1}, Source: "Source: synthetic reference content", Nodes: []Node{text("body", "body", "Two title lines allocate the header through 144 pt. Body content starts at 162 pt; the source line reduces the body bottom to 450 pt.", 57, 198, 846, 72)}},
		SlideSpec{ID: "appendix-source", Eyebrow: "Explicit appendix density", Title: "Reference details use a compact header", Frame: FrameRequest{Rail: "none", Footer: "compact", Density: "appendix", SourceLines: 2}, Source: "1. Synthetic reference note.\nSource: WMDS frame specification.", Nodes: []Node{text("body", "small", "Appendix is selected explicitly. It uses the heading title style and begins the body at 108 pt. Two source lines reduce the body bottom to 432 pt.", 57, 144, 846, 72)}},
		SlideSpec{ID: "dark-frame", Eyebrow: "Surface reference", Title: "An inverse surface preserves the roles", Frame: FrameRequest{Rail: "none", Footer: "compact", Surface: "inverse"}, Nodes: []Node{text("dark-heading", "heading", "Primary ink stays readable", 57, 144, 846, 36), text("dark-body", "body", "The inverse surface uses white primary text, Medium Gray secondary text, and Magenta emphasis.", 57, 216, 846, 72)}},
		SlideSpec{ID: "no-header", Frame: FrameRequest{Rail: "none", Footer: "compact", NoHeader: true, NoPage: true}, Nodes: []Node{text("cover-eyebrow", "eyebrow", "WMDS foundation", 57, 144, 846, 18), text("cover-display", "display", "A foundation for\nmodern slide design", 57, 198, 846, 144), text("cover-copy", "lead", "Typography, grid and frames generated as native PowerPoint objects.", 57, 378, 846, 30)}},
	)
	return doc
}
