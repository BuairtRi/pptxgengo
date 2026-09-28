# Adaptive slide families

## Availability and route choice

This route is in development after 0.1.0-local.3. Check `pptxgengo --help` before
using it; the frozen .3 installation does not have `adapt`. Do not replace the
user's installed release merely to access a development command.

Use `adapt` when counts, relationships, or semantic states should drive new
editable geometry. Use `template` for the existing source-bound text/style
contracts, and `compose` for supported structures outside these five families.
A source reference is useful design context; it does not select a source slide
for automatic reshaping.

Use `pptxgengo paths` to locate the installed library. Read only the relevant
family fixture and schema guidance under `library/adaptive/`:

| Family | Content that drives geometry | Detail to retrieve |
| --- | --- | --- |
| Roadmap | Ordered periods, workstreams, interval endpoints, milestones | `roadmap-comparison.md`, `roadmap-normal.json` |
| Architecture | Layers, individual component boxes, relative widths, left rail, relationships | `architecture-team.md`, `architecture-standard.json` |
| Process | Two to six stages, summaries, activities, outputs, semantic state | `process.md`, `process-normal.json` |
| Team | Pods, roles, staffing tokens, reporting lines, matrix rows/columns | `architecture-team.md`, `team-standard.json` |
| Comparison | Options, criteria, supplied gauge values/targets, findings/statuses | `roadmap-comparison.md`, `comparison-illustrative.json` |

Inspect `library/adaptive/checkpoint.json` for exact reviewed examples. The
capability catalog keeps source editing contracts separate from semantic
family discovery hints. A category hint requires structural inspection and does not qualify that source template
for structural adaptation or arbitrary copy. Only the recorded input, geometry,
style, and render have been reviewed.

## Content contract

Input is strict JSON with schema `pptxgengo.adaptive-deck.v1` and `slides[]`.
Each slide requires a unique `id`, `family`, `title`, one-sentence `role`,
`takeaway`, and family-specific `content`. The role and takeaway are speaker-note
metadata; write the visible title and body so they communicate the argument.

Optional fields are `template_id`, `style`, `bounds`, and `footer`.
`template_id` is provenance metadata echoed in the report; it does not select or import a source slide. Omit the
footer or supply the intended visible footer. The library examples explicitly
label themselves illustrative; replace their synthetic content before client use.

Keep substantive evidence, ownership, qualifications, and acceptance conditions
in the content. Comparison scores and targets must be supplied or explicitly
hypothetical. The engine positions them; it does not infer a score from prose.
Array order controls ordering. Relationship endpoints use stable IDs. Unknown
fields and invalid references are errors, so inspect the relevant schema instead
of inventing controls.

## Style and composition

Shared styles support `neutral`, `subtle`, and `inverse` profiles; Arial; explicit
title/body/label point sizes; and named color roles. Consult the current catalog
for accepted role names and ranges. Foreground colors may resolve to contrasting
brand ink on individual surfaces. The shared style check compares `text.primary`
against `surface`; family and compose checks apply to actual component surfaces.
Process body ink resolves to navy/white. The fixed logo/title/footer area has its
own styling, including fixed footer ink. Staffing tokens retain their ownership meaning and accompanying legend.

`bounds` relocates/resizes the body within the declared slide region. Family
builders recompute boxes and routes from the content counts. They are not
free-form layout editors. When more control is needed, use a separate, clearly
named compose spec derived from the compilation, preserve its provenance, and
run the complete compose review on that new spec. Do not silently edit a measured
bundle and continue using its old evidence.

## Compile and review

On a release whose help lists `adapt`:

```sh
pptxgengo adapt compile --spec /path/to/content.json --out /path/to/new-compilation
```

For repository development before such a release, run from the repository root:

```sh
go run ./cmd/pptxadapt compile --spec /path/to/content.json --out /path/to/new-compilation
```

The new bundle contains the semantic input, generated `spec.json`, resolved
controls/limitations, and hashes. It has not passed native review merely because
compilation succeeded.

Follow [the compose workflow](compose-authoring.md) on that `spec.json`:
probe, measure with the matching binary/environment, inspect the fit report,
build, verify in PowerPoint, render, and visually review every final page.
Use one binary for a measurement run. Reuse the user's established native staging
folder. Do not run multiple PowerPoint automation jobs concurrently.

If text does not fit, change the layout, redistribute content, or split the slide
while preserving meaning. Do not quietly delete qualifications, lower font size,
or reuse measurements from a different contract to claim success. Native fit
checks do not catch every optical alignment or semantic-legend problem.

## Visual relationships to check

Check the relationships a reader will infer, as well as text fit. A workstream
label and its activity bars should share one vertical centerline; milestone
captions belong below that line. Architecture layer labels should stay attached
to their component rows. Group cross-cutting controls separately, state that they
apply to every layer, and use a visible divider and spacing so their positions
do not suggest a false one-to-one row mapping. Count mismatches need a clear
visual explanation. Inspect these relationships in every final render.
