# Native verifier batching spike

2026-09-26. Read-only inspection against the saved `wave2-review.pptx`.

PowerPoint 16.113.2 does not return one record per character for
`properties of every character of tr`. It returns a single character-class
record covering the entire title: text length 25, content `Deliverable preview
panel`, and the full range bounds. Likewise, `left bounds of every character`
returns one left edge; `properties of font of every character` returns an
aggregate font record. Treating these as individual character observations would
silently weaken the fit/style proof. No such optimization was adopted.

The v8 adapter reads paragraph formatting as one property snapshot, replacing
five individual spacing/alignment reads per paragraph. Character bounds and
styles remain individually inspected. No large speedup is claimed. Final native
verification remains expensive and must not be replaced by cached probe results.

The adapter additionally snapshots native bullet properties for every paragraph.
Indentation is verified from exact OOXML because PowerPoint's AppleScript API
exposes outline-level rulers, not reliable per-paragraph hanging-indent values.
The bullet itself is absent from text-range content; visual checks must include
its glyph and the first/wrapped-line alignment.

Smoke regression: the v8 adapter inspected the prior six-object smoke deck in
14.28 seconds. Projecting the new output onto every field present in its saved
v6 native evidence produced exact equality for all six objects, including each
character style and glyph union. New paragraph fields were additionally captured.
Raw output: `samples/visual-wave2-followup/adapter-v8-smoke.json`.
