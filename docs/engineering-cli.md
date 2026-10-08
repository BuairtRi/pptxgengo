# Engineering command reference

These examples target the current `0.1.0-local.21/v11` package and use its `pptxgengo` wrapper. New projects default to native-v1 wherever the measured template records support safe conversion; this implementation default does not establish native visual qualification. V11 native review remains pending. Output paths marked `NEW-DIR` must not already exist. Commands that inspect PPTX or project sources do not edit those inputs.

## Wrapper defaults and project pins

The installed wrapper reads `release/default-bundle.txt` for the default
published library (`v11` in local.21). Bundle-backed `pptxgengo design` commands
receive that bundle path and the Go engine `wmds-go-foundation.v2` unless you
pass explicit flags. Project commands use the project lock: `project init`
creates a new lock against the published default and candidate engine; later
project operations use the locked bundle and engine. Native render, doctor,
source-inventory, asset-gallery, and project commands do not receive a wrapper
bundle or engine injection.

For `library-find`, `library-inspect`, and `library-preview`, the wrapper also
supplies the V11 SQLite index and catalog path when `--index` is omitted. If you
pass a custom index, pass its matching `--gallery` explicitly when verified
preview paths are needed.

`library-find` ranks all matching indexed kinds unless you pass `--kinds`.
Unrelated zero-score results are omitted when text is queried; `--include-weak`
shows them for broad inventory work. Icon color instances are represented as a
single concept hit with each registered ID retained in `variant_ids`. For a
focused template search, set `--kinds template`; asset summaries use
`--kinds asset --summary --asset-kind icon|photo|graphic|logo` and rank
curated terms and filename-derived tags with partial matches.

The photo registry covers all 521 originals in the local
`~/Documents/branding/West Monroe Photos` collection, within 523 photo catalog
records and 1,204 total asset variants. Their descriptive
sidecars supply the subject labels and the checked-in registry pins the local
relative path and current file hash. The bundled brand-assets inventory lists
opaque stock filenames but has no mapping to these renamed files, and the
sidecars do not retain original asset IDs, URLs, or licensing records. Registry
membership is not a license or approval assertion; confirm those records before
external publication.

To upgrade an existing YAML project to the current V11 source:

```bash
pptxgengo design project migrate --project ./my-project --dry-run
pptxgengo design project migrate --project ./my-project
pptxgengo design project build --project ./my-project
```

`project migrate` validates template compatibility and compiler fit against the
target before changing anything. Successful migration preserves the exact old
lock beside the configured lock as `.pre-migrate-<hash>`, then atomically replaces
the pin. Dry runs and failed validation leave the lock unchanged. It defaults to
V11; use `--bundle` or `--engine` for an explicit target. Repeating an already
current migration creates no extra backup. YAML copy, template definitions and
existing PowerPoint files are not rewritten. Custom ancestry or revision
incompatibilities are reported for repair. Build and native-review the updated
deck; old visual acceptance does not qualify its new rendering.

`project init` remains the command for a project without a lock.

## Inventory an existing PowerPoint deck

Create JSON and Markdown inventories with source slide order, titles, hidden state, text, table cells, chart types and series, picture alt text and package paths, and speaker notes:

```bash
work=$(mktemp -d)
pptxgengo design source-inventory \
  --in "samples/final/dentalxchange/deck.pptx" \
  --out "$work/source-inventory"
```

The output contains `source_inventory.json` and `source_inventory.md`. It includes a source SHA-256 and does not include raw XML. It does not infer a layout contract, create an editable project, or compare imported text with a later PowerPoint edit.

To add project source references, first make and initialize a separate working copy of a project. The mapping report lists source slides and project slides independently and marks them unmapped unless you map them yourself:

```bash
work=$(mktemp -d)
cp -R examples/deck-project "$work/project"
pptxgengo design project init --project "$work/project"
pptxgengo design source-inventory \
  --in "samples/final/dentalxchange/deck.pptx" \
  --out "$work/source-and-project" \
  --project "$work/project"
```

With `--project`, the command verifies the project's lock and compiles its current source. `mapping_report.json` includes template scope, template ID and revision, template source pins, project slide source paths, and explicit `unmapped` statuses. The command does not match pages by number, title, text, imagery, or visual layout.

## Find and review registered assets

Search a specific asset kind with a semantic query. Icon colors are grouped together while each original registry ID remains listed:

```bash
pptxgengo design library-find \
  --kinds asset --summary --asset-kind photo \
  --query "people collaborating workshop" --limit 10
```

Queries require a positive match; there is no unrelated fallback result. The thumbnail paths point to hash-verified registered originals. SVGs remain SVG, and the MIME field identifies whether a result is vector or raster.

Create a browsable gallery from verified originals. Raster previews are derived
in Go at up to 420 × 280 pixels; SVG previews retain the original vector bytes.
Use the registry originals, not the browser thumbnails, in presentations:

```bash
work=$(mktemp -d)
pptxgengo design asset-gallery \
  --out "$work/people-photos" --kind photo \
  --query "people collaborating workshop"
open "$work/people-photos/index.html"
```

`--kind` accepts `icon`, `photo`, or `all`; it can also be set to `graphic` or `logo`. Use `--limit N` to keep a gallery small.

The release assets gallery is listed by the wrapper catalog command:

```bash
pptxgengo catalog --assets --print
# Or open the installed release gallery in the browser:
pptxgengo catalog --assets --open
```

The installed gallery is generated from registered asset metadata and verified originals when the release is assembled.

For a particular library template, inspect named slots, source descriptions, aliases, and measured text capacity. Run the coverage form without a template key to see alias coverage:

```bash
pptxgengo design library-authoring --template cards/3
pptxgengo design library-authoring
```

Alias coverage and capacity estimates are source-derived metadata. Some slots are still generic or ambiguous and have not had semantic review; native drawing support and visual qualification also differ by template. Treat the report as authoring guidance, inspect the exact candidate and test the populated slide before presenting it as a fit.

Measured estimates now cover the typed `cards/3` and `cards/4` titles and bodies,
plain fixed native table cells, and supported stepper fields. Card bodies reserve
a two-line title. These estimates reuse the renderer's geometry and pinned fonts;
they do not establish native fit. Rich table cells and many component internals
still report unavailable capacity. Engineering recipes clarify the main statement
in `key-message/statement`, status columns, and dense interview fields; these are
source-derived aliases, not a completed human semantic review of the library.

## Compare content-first template candidates

Write one semantic page description in YAML or JSON and ask the matcher for candidates:

```bash
work=$(mktemp -d)
cat > "$work/page.yaml" <<'YAML'
title: Make the decision clear
eyebrow: Recommendation
items:
  - {lead: Assess, text: Confirm the evidence and the decision owner.}
  - {lead: Compare, text: Show the tradeoffs and implications.}
  - {lead: Decide, text: Agree the next step and timing.}
relationship: parallel
YAML
pptxgengo design library-match \
  --page "$work/page.yaml" --limit 4 --out "$work/candidates"
```

`library-match` records slot mapping, fit outcomes, unused content and generated candidate decks. Add `--render` to request native PowerPoint output and a contact sheet. If native rendering fails, candidate files and the `render_failed` status remain available for review.

`--limit` is the number requested, not a guarantee. Matching currently uses
known structural roles and relationships; unsupported complex schemas are
reported as gaps. A report with zero complete candidates exits with an error
and retains the report for inspection. It does not create synthetic alternatives
to reach the requested count.

Incomplete mappings use `needs_copy`, list missing slots and unused source
fields, and expose a ranked `draft_slide` separately from a ready `authored_slide`.
Ready slide output requires a complete mapping and a successful Go layout build.
Missing copy is never silently filled from the specimen. Different item counts
can carry the first N items into a known group; additional items remain explicit.

Bounded adapters support `lifecycle/three-phases` with each item's `objective`
and nested `activities: [{lead: ..., text: ...}]`, and
`vendors/scorecard-generic` with this rectangular comparison shape:

```yaml
relationship: comparison
comparison:
  criterion_label: Criterion
  weight_label: Weight
  criteria:
    - {label: Delivery confidence, weight: "30%"}
  options:
    - {name: Option A, values: ["High"]}
    - {name: Option B, values: ["Medium"]}
  legend: High / Medium / Low
```

The example illustrates the schema; the chosen stock slide's fixed row and column
counts may require more copy. Sequence `owner`, `duration` and `state` fields map
only where the template declares corresponding editable roles. Unsupported
relationships or topology remain reported gaps; inspect candidates before use.

For a shorter deterministic example that exercises a known three-card template:

```bash
pptxgengo design library-match \
  --page "$work/page.yaml" --templates cards/3 --limit 1 \
  --out "$work/cards-3"
```

## Register client-supplied images

Register an original image in a project without automatically cropping or applying it to a slide:

```bash
pptxgengo design project asset add \
  --project ./my-project --id client-workshop-photo \
  --file ./approved/workshop.jpg \
  --description "Client workshop, Phoenix, June 2026" \
  --focus 0.5,0.5
```

The focus point is advisory metadata. The project retains the supplied original and its content hash. A slide can refer to it through the project's declared asset ID.

## Titles, measurements, and template swaps

List the compiled titles in source slide order. JSON includes page, stable slide ID, title, hidden state, and template ID:

```bash
pptxgengo design project titles --project ./my-project
pptxgengo design project titles --project ./my-project --format json
```

Measure only selected pages or stable IDs from an existing build report. Numeric pages and ranges are one-based; mixed selections return in original report order. Unknown IDs and out-of-range pages fail explicitly.

```bash
pptxgengo design project measure \
  --report ./my-project/builds/BUILD-ID/layout-report.json \
  --slides recommendation,4-6
```

Propose a template swap first, then apply the displayed proposal after reviewing its mapped and removed fields:

```bash
pptxgengo design project swap --project ./my-project \
  --slide recommendation --template key-message/statement
pptxgengo design project swap --project ./my-project \
  --slide recommendation --template key-message/statement --apply
```

If the proposal includes visible fields that cannot map, the apply step fails unless `--allow-unmapped` explicitly authorizes removing them. Source history is retained.

Add, position, hide or remove a slide by stable identity:

```bash
pptxgengo design project slide add --project ./my-project --id recommendation --file candidate.yaml --after opening
pptxgengo design project slide move --project ./my-project --id recommendation --before summary
pptxgengo design project slide hide --project ./my-project --id recommendation
pptxgengo design project slide show --project ./my-project --id recommendation
pptxgengo design project slide remove --project ./my-project --id recommendation
```

`--id` must match the supplied slide file's `id`. Matcher files use IDs such as
`candidate-001`; use that ID when adding one, or edit its `id` before adding it
under a meaningful project identity. The original supplied file remains intact.

Attach internal Draft Review Notes to any slide without changing its template:

```bash
pptxgengo design project slide draft-review set --project ./my-project --id recommendation \
  --status wip --status-text 'Work in progress' --status-color kpi.risk --owner Ri --due 10/20 --updated 10/05 --notes 'Confirm the source.'
pptxgengo design project slide draft-review set --project ./my-project --id recommendation --status qa
pptxgengo design project slide draft-review show --project ./my-project --id recommendation
pptxgengo design project slide draft-review clear --project ./my-project --id recommendation
```

Statuses are `notstarted`, `wip`, `complete`, and `qa`. The tab displays the
selected status label by default. `--status-text` sets a custom label independently
of `--status-color`, which accepts `kpi.off`, `kpi.risk`, `kpi.on`, or `brand.blue`.
Custom text must be a single line of at most 128 bytes and fit the tab. Empty
status text/color resets the override to the selected status default. Other
optional fields are `--owner`, `--due`, `--updated`, and `--notes`; dates are display
strings.
`--placement edge` (the default) shows the status tab at the top-right slide edge
with its note body outside the canvas; `--placement pasteboard` places the entire
234 × 144 pt component outside the canvas. `set` preserves omitted fields,
initializes a new note with `notstarted`, and accepts explicit empty strings to
clear optional text. The commands return JSON receipts and preserve comments,
unrelated metadata, and unselected files in inline and split projects. Identical
`set` calls and clearing an absent note preserve source bytes without a new decision.

Normal builds and reviewer/maintainer/offline exports retain the editable native
component. Client exports remove its status, body and metadata while preserving
the immutable draft build. Audience copy review excludes Draft Review Notes.
YAML authors can add the optional slide-level `draft_review` mapping with
`status`, `status_text`, `status_color`, `owner`, `due`, `updated`, `notes`, and `placement` directly; quote date
values that YAML could interpret as timestamps.

Draft Review Notes were visually checked in Microsoft PowerPoint on 2026-10-05:
the four default status labels, whole-pasteboard placement, and a custom
`Work in progress` label with a blue tab all fit the fixed component. The standard
short test suite and focused race checks passed. Automated native export was not
qualified in that session because Apple event dispatch failed; the visual check
used PowerPoint directly.

Moving a section anchor requires `--reanchor`. Removed slide/notes files remain
in the project and are listed in the receipt. When a composition log is present,
update its authored entries after adding/removing a slide or changing a template;
the next check rejects missing, obsolete or mismatched entries.

Removing a slide that starts a section moves that section anchor to the next
slide in its current range, or removes the section if it becomes empty. The
receipt records the section change. Section declarations can also be managed
directly:

```bash
pptxgengo design project section list --project ./my-project
pptxgengo design project section add --project ./my-project \
  --id findings --title Findings --before recommendation
pptxgengo design project section rename --project ./my-project \
  --id findings --title Evidence and implications
pptxgengo design project section remove --project ./my-project --id findings
```

Removing a section drops its boundary. Its slides join the preceding section;
when removing the first section, the next section is reanchored to the deck's
first slide. The section receipt records the new anchors.

Render a quick subset after an edit, or check native-render prerequisites before the full pass:

```bash
pptxgengo design render-doctor --json
pptxgengo design render --pptx ./my-project/builds/BUILD-ID/deck.pptx \
  --out "$work/render" --pdf --png --slides 3,5-7 --contact-sheet
```

The production Mac release uses local Microsoft PowerPoint and PDFKit. The
experimental Windows preview uses desktop PowerPoint COM for native PDF and
slide PNG export. Both leave the source PPTX unchanged; inspect full-size pages
and the contact sheet after export. The Windows tester ZIP includes a user-level
installer and smoke-test script; see [Windows tester guide](../internal/releasepackage/WINDOWS.md).

Render detects PowerPoint's Grant File Access dialog and reports it as a file
access failure. Failed runs attempt to leave `render-error.txt`, including
preflight failures and timeouts; an inaccessible output directory is reported
explicitly. `render-doctor` probes file access by opening a generated deck,
exporting a one-slide PDF, and closing that exact deck after the operational
automation check succeeds. Doctor and render place their uniquely named PPTX/PDF
files directly in the same stable staging folder. Scripts and rasterization
files remain in private task subfolders. Use the same `--staging-dir` (or
`PPTXGENGO_NATIVE_STAGING`) for both commands; the default is the user cache.
The probe may show a grant dialog for the operator to clear for that folder.
A pass records the observed read/write operation, not a guarantee of future
access. Unknown visibility or a broken
GUI caller is reported as unknown; it does not establish permission denial.

### Local agent-session recovery

Machine-specific Codex daemon and dead-port observations are kept in
[local-agent-macos-recovery.md](local-agent-macos-recovery.md). Those notes are
for the inspected local installation and do not describe PowerPoint or
`pptxgengo` recovery steps generally.

Attach the native output to the build whose exact PPTX was rendered:

```bash
pptxgengo design project attach-render \
  --project ./my-project --render "$work/render"
pptxgengo design project status --project ./my-project
```

The attachment validates the render receipt, page identities, source and output
hashes. Rendering alone records `rendered`; it does not mark a slide reviewed or
accepted. After inspecting the pages, pass a JSON decisions file to
`attach-render --decisions decisions.json`:

Receipts must carry the local Ed25519 issuance signature written by a successful
`render`. Unsigned older or hand-written receipts are rejected; rerender with the
current CLI. Verification uses the locally provisioned issuer key, stored under
the user's configuration directory in `pptxgengo/native-trust`. This establishes
local CLI issuance and integrity, not OS attestation: code running as the same
user can access that key. Cross-machine verification requires the original
trusted key; this release has no trust-transfer command. Source/build/artifact
hash checks still apply. Unknown fields in the decisions JSON are rejected.

```json
{
  "recommendation": {
    "status": "accepted",
    "reviewer": "operator",
    "note": "Reviewed the native page; labels and copy fit."
  }
}
```

Decision statuses are `reviewed`, `accepted`, or `issues_found`. Changed source,
build or artifacts invalidate the attachment's coverage. Editing a context file
or composition log changes project inputs and marks all prior native review
coverage stale. Rebuild, render the current deck and attach its fresh receipt
before recording new review decisions. Hidden or unrendered slides remain
explicitly uncovered.

## Safe slide and section edits

Slide operations use stable IDs. `slide add --file` leaves the supplied YAML
untouched; `--as NEW-ID` assigns the project identity, and `--into-section ID`
places the slide at that section's first position and reanchors it. `--check-fit`
builds the candidate with the project's locked Go engine before committing it.
It reports engine fit checks, not native PowerPoint visual review. A retained
slide file must be re-added byte-for-byte if removed; editing an existing source
file requires the project edit workflow.

```bash
pptxgengo design project slide add --project ./my-project \
  --file ./new-slide.yaml --as recommendation --into-section findings --check-fit
pptxgengo design project slide move --project ./my-project \
  --id recommendation --into-section findings
```

Section removal reanchors slides at the removed section's former position to a
neighboring section when possible. If the removed section becomes empty it is
removed. Inspect the operation receipt and run `project check` after structural
changes.

## Media delivery policy

Project builds use the existing delivery policy when `document.media_optimization`
is absent: deduplicate media, compress the package, and safely resize JPEGs to
220 pixels per inch at quality 90 when savings reach 10 percent. Vectors and
lossless graphics remain byte-identical. Unsafe JPEG placements, EXIF orientation
or color profiles, low savings, and already adequate resolution can preserve the
original bytes. `layout-report.json` records source/output hashes, byte counts,
dimensions, and skip or resize reasons per package part. Original registered
assets remain in `assets/originals`.

An explicit `media_optimization` policy must provide all six settings:
`deduplicate`, `resize_jpeg`, `compression`, `pixels_per_inch`, `jpeg_quality`,
and `min_savings_percent`. Set `resize_jpeg: false` to preserve JPEG payloads.

## Keep composition and evidence references checkable

An optional `composition-log.yaml` at the project root (or the path named by `context.composition_log`) records the rationale for each slide. Every project slide needs one entry, and `chosen_template` must match the slide source:

```yaml
schema: pptxgengo.composition-log.v1
slides:
  recommendation:
    purpose: State the recommended decision and next step
    relationship: sequence
    candidates: [key-message/statement, cards/3]
    chosen_template: cards/3
    rationale: Three actions need parallel emphasis under one conclusion.
    unresolved: []
```

When `context.claims` points to a claims registry, each `evidence_refs` ID on a slide must exist there. A YAML registry looks like:

```yaml
schema: pptxgengo.claims.v1
claims:
  - id: claim-001
    text: The client takes about two days to prepare each product requirements document.
    source: Discovery interview, 2026-09-18
```

Markdown claims use explicit `## claim-id` headings or `{#claim-id}` anchors. Plain prose does not establish an evidence ID. If `context.claims` is absent, evidence references are unchecked; add a registry path to enable claim validation. `pptxgengo design project check --project ./my-project` verifies the project structure, composition entries, and linked claim references.

## Review a fresh project packet

Generate an outline, content, or deck reviewer packet in a new directory:

```bash
work=$(mktemp -d)
pptxgengo design project review \
  --project ./my-project --stage outline --out "$work/review"
```

The staged HTML packet includes the project's composition rationale and any linked claims. Review stages are `outline`, `content`, and `deck`. These packets are requested with `--stage`; `project review --out PATH` without it keeps the earlier ZIP export behavior. `project view --out NEW-DIR` is the deck-stage shortcut.

To create a presentation for external audience review, add `--audience`:

```bash
pptxgengo design project review --project ./my-project \
  --stage content --audience --out "$work/audience-review"
```

Audience packets include visible-slide content and omit hidden slides, internal
notes, composition rationale, briefs, decisions and prior review verdicts. The
deck stage requires current native PNGs or a PDF containing visible pages. The
outline stage needs no rendered deck. Content stage measures the current authored
copy with the Go renderer; it does not require a native attachment. Deck stage
requires a current build and verified native output for every visible slide.
Visible text includes shape text and native table cells. Native chart text is
explicitly marked as requiring rendered-page review; workbook data is not exposed
as visible copy because chart options can hide series names and values.

The `deck` stage requires a current build. Its HTML page includes attached native
thumbnails and review status when available; otherwise it reports the missing
native evidence. Copies of the deck, PDF, PNGs and evidence are retained in the
packet with their hashes. Audience mode uses only visible-page native output and
does not include internal notes or verdicts.

## Design-system documentation

`pptxgengo docs` serves the installed foundations, primitives, components, composites, frames and template reference board. `pptxgengo paths` lists `design_docs` and its pinned `design_docs_source`. The installer verifies the docs commit, embedded canonical JSON and assets against the selected native library. The docs board explains the system; `pptxgengo catalog --design-system --open` opens the separate accepted native specimen gallery.
