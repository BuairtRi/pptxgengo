# Process input

A `process` slide requires `content.steps`: 2–6 ordered stage objects. Each has
`id`, `label`, `summary`, `activities` (1–8 nonempty strings), and `outputs`
(1–5 nonempty strings). IDs must be unique and contain no slash. Optional
`state` is `planned` (default), `active`, or `complete`.

Optional `content.gap_pt` accepts 12–36 pt and defaults to 18 pt. Counts and the
available body width determine equal stage widths; widths below 115 pt are
rejected. The stage heading band is 42 pt. Summary, activity, and output bands
share measured row heights across all stages, so corresponding sections align
when copy wraps. The native planner rejects content that exceeds the available
height.

Use `process-normal.json` as the complete input example. Other process fixtures
exercise two, three, four, and six stages, longer summary text, and inverse
styling. `checkpoint.json` records the exact five examples with reviewed native
renders and evidence. Those records apply only to those inputs; they do not
prove every allowed count, font size, or content combination fits.

The source reference T027 contributes a sequencing motif. Its lower cutover
matrix is not implemented by this builder. Add a separate measured composition
for that matrix or choose the original source-bound contract when that source
layout should be retained.
