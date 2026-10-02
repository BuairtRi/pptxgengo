package wmdesign

import "encoding/json"

// TemplateReference binds fictional copy to four fixed source templates. Each
// pair retains geometry and styling while changing only declared content.
func TemplateReference(year int) BoundDocument {
	d := BoundDocument{Schema: "pptxgengo.wmds-template-document.v1", Year: year}
	values := func(v any) json.RawMessage {
		b, err := json.Marshal(v)
		if err != nil {
			panic("invalid built-in template reference: " + err.Error())
		}
		return b
	}
	add := func(id, template string, v any) {
		d.Slides = append(d.Slides, BoundSlide{ID: id, Template: template, ContentKind: "synthetic_example", Values: values(v)})
	}
	type card struct {
		Key   string `json:"key"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	add("template-cards-three-a", "cards/3", map[string]any{
		"eyebrow": "Illustrative approach", "title": "Three moves for a sample close",
		"cards": []card{
			{"diagnose", "Diagnose", "Map the sample close calendar and identify each source of delay."},
			{"design", "Design", "Move review earlier and define the checks for each journal."},
			{"deploy", "Deploy", "Pilot the process with one team before a broader rollout."},
		},
	})
	add("template-cards-three-b", "cards/3", map[string]any{
		"eyebrow": "Illustrative approach", "title": "The same cards in a new order",
		"cards": []card{
			{"deploy", "Pilot first", "Start with a bounded pilot and record the lessons for rollout."},
			{"diagnose", "Map delays", "Review the sample calendar and assign an owner to each delay."},
			{"design", "Set checks", "Agree the entry checks and exception review before deployment."},
		},
	})
	add("template-cards-four-a", "cards/4", map[string]any{
		"eyebrow": "Illustrative approach", "title": "Four steps in a sample rollout",
		"cards": []card{
			{"diagnose", "Diagnose", "Rank the delays in a sample close."},
			{"design", "Design", "Agree the checks and exception owners."},
			{"deploy", "Deploy", "Pilot with one illustrative team."},
			{"sustain", "Sustain", "Review the process each month."},
		},
	})
	add("template-cards-four-b", "cards/4", map[string]any{
		"eyebrow": "Illustrative approach", "title": "Reorder the steps without resizing",
		"cards": []card{
			{"sustain", "Review", "Use a monthly check to refine the process."},
			{"design", "Set rules", "Define each check and its named owner."},
			{"diagnose", "Map gaps", "Map the sample process and find delays."},
			{"deploy", "Run pilot", "Apply the agreed checks in a small pilot."},
		},
	})
	metric := func(value, label string) map[string]any {
		return map[string]any{"value": value, "label": label}
	}
	formatted := func(raw, kind, label string) map[string]any {
		return map[string]any{"format": NumberFormatSpec{Value: json.RawMessage(raw), Kind: kind}, "label": label}
	}
	rich := func(key string, runs ...RichRunSpec) RichTextSpec {
		return RichTextSpec{Paragraphs: []RichParagraphSpec{{Key: key, Runs: runs}}}
	}
	add("template-stats-a", "stats/four-metrics", map[string]any{
		"eyebrow": "Illustrative scenario", "title": "A sample close scenario",
		"metric1": metric("5 days", "illustrative time to close"),
		"metric2": formatted("4250000", "currency", "illustrative run-rate savings"),
		"metric3": formatted("14", "number", "business units in the scenario"),
		"metric4": metric("3", "rollout waves"),
		"support": "Sample savings reflect automated matching and earlier review.",
		"source":  "Illustrative scenario only; no client results.",
	})
	add("template-stats-b", "stats/four-metrics", map[string]any{
		"eyebrow": "Illustrative scenario", "title": "A second sample scenario",
		"metric1": metric("6 days", "illustrative time to close"),
		"metric2": formatted("6500000", "currency", "illustrative run-rate savings"),
		"metric3": formatted("0.18", "percent", "illustrative cost reduction"),
		"metric4": formatted("4", "number", "rollout waves"),
		"support": rich("support", RichRunSpec{Key: "prefix", Text: "A "}, RichRunSpec{Key: "owner", Text: "named owner", Weight: 600, Ink: "emphasis"}, RichRunSpec{Key: "suffix", Text: " reviews each exception before the close."}),
		"source":  "Illustrative scenario only; no client results.",
	})
	add("template-rail-a", "takeaway-rail/metrics-rail", map[string]any{
		"eyebrow": "Illustrative scenario", "title": "Illustrative evidence",
		"metric1":     formatted("4200000", "currency", "sample run-rate savings"),
		"metric2":     formatted("0.18", "percent", "sample cost reduction"),
		"support":     "The sample model combines earlier review and automated matching.",
		"railEyebrow": "Sample takeaway", "railHeading": "Pilot starts in Q1",
		"railBody": "A bounded pilot gives the team a way to review the sample assumptions.",
		"source":   "Illustrative scenario only; no client results.",
	})
	add("template-rail-b", "takeaway-rail/metrics-rail", map[string]any{
		"eyebrow": "Illustrative scenario", "title": "Review the sample assumptions",
		"metric1":     formatted("1900000", "currency", "sample run-rate savings"),
		"metric2":     metric("14 mo", "sample payback period"),
		"support":     rich("support", RichRunSpec{Key: "prefix", Text: "The "}, RichRunSpec{Key: "model", Text: "sample model", Weight: 600, Ink: "emphasis"}, RichRunSpec{Key: "suffix", Text: " needs review before any delivery decision."}),
		"railEyebrow": "Sample next step",
		"railHeading": rich("heading", RichRunSpec{Key: "lead", Text: "Review "}, RichRunSpec{Key: "focus", Text: "the pilot", Weight: 600, Ink: "emphasis"}),
		"railBody":    rich("body", RichRunSpec{Key: "lead", Text: "Confirm "}, RichRunSpec{Key: "owners", Text: "named owners", Weight: 600}, RichRunSpec{Key: "tail", Text: " and check each assumption in the illustrative model."}),
		"source":      "Illustrative scenario only; no client results.",
	})
	return d
}
