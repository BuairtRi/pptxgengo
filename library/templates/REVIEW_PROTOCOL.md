# Template expansion review protocol

This catalog classifies source designs; it does not grant adaptation approval.
Source registry: `planning/source-registry.json`. Reuse existing decisions in
`library/layout-decisions.json` before proposing another family.

## Independent fields

- `category`: narrative, comparison_evidence, approach_delivery, roadmap,
  team_credentials, outcomes_commercial, visual_storytelling,
  business_architecture, technical_architecture, product_overview,
  product_detail, layers_components, framing_navigation, reference_instruction.
- `altitude`: framing, overview, explanation, detail, reference.
- `density`: sparse, medium, dense. Altitude is purpose, not word count.
- `pattern`: concrete arrangement, e.g. four-column phase overview.
- `disposition`: prioritize, candidate, reference_only, avoid.
- `review_status`: visual_reviewed or metadata_only (never imply visual evidence
  from source text alone).

## Worker output

Write ONE JSON file at your assigned path. Schema `pptxgengo.template-review.v1`.
Top-level: `reviewer`, `source_hashes`, `slides`, `families`, `sequences`, `gaps`.
Each source occurrence must appear once in slides:
`source_id`, `slide_number`, `title`, `category`, `altitude`, `density`, `pattern`,
`family_id`, `disposition`, `rationale`, `component_needs` (string list),
`review_status`, `preview_path`, `preview_sha256`.
Use `family_id` from existing decisions for existing group members; otherwise
choose `candidate:<slice>:<descriptive-name>`. Group confirmed arrangements;
color/content changes alone are variants. Similar topics do NOT establish a match.
Every proposed family: `id`, `name`, `representative` (source:NNN), `members`,
`dedup_rationale`, `adaptation_complexity` (low/medium/high), `suggested_slots`,
`variant_notes`, `priority_reason`. Keep different slot structures distinct.
Sequences are separate from template counts: `id`, `name`, `source_examples`,
`steps` (role/altitude/possible_source_refs), `repeat_rule`, `continuity_rules`.
Record uncertainty as a gap; do not invent review evidence or capacity limits.

## Visual evidence

Root prepares full-slide PNGs at
`samples/template-expansion/render/<source_id>/slide-NNN.png` from hash-verified
existing native PowerPoint PDFs. Inspect each slide individually (multiple
individual images per tool call is fine). Read structural text/metadata in
`samples/inspection/<source_id>/inventory.json` and sidecars. If previews are
still rendering, work on available sources first. Do not run PowerPoint.
Existing inspected geometry and source component manifests are useful inputs.

Select a useful breadth of source arrangements, not a quota of color variants.
Product and architecture views belong in the library even where advanced diagram
routing is not implemented. Capture overview-to-detail relationships and framing
slides, phase overview -> 1–3 pages per phase, product portfolio -> product overview
-> capabilities/components -> detail, architecture context -> layers -> layer detail.
Do not commit or edit shared index/schema/contracts. Root integrates all slices.
