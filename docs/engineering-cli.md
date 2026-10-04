# Engineering command reference

These examples use the installed `pptxgengo` wrapper. Output paths marked `NEW-DIR` must not already exist. Commands that inspect PPTX or project sources do not edit those inputs.

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

Moving a section anchor requires `--reanchor`. Removed slide/notes files remain
in the project and are listed in the receipt. When a composition log is present,
update its authored entries after adding/removing a slide or changing a template;
the next check rejects missing, obsolete or mismatched entries.

Render a quick subset after an edit, or check native-render prerequisites before the full pass:

```bash
pptxgengo design render-doctor --json
pptxgengo design render --pptx ./my-project/builds/BUILD-ID/deck.pptx \
  --out "$work/render" --pdf --png --slides 3,5-7 --contact-sheet
```

Rendering uses local Microsoft PowerPoint and PDFKit. The render command leaves the source PPTX unchanged; inspect the resulting pages and contact sheet after export.

### Agent dispatch recovery on this Mac

The current agent's native events fail with −10827/−600. Read-only probes show
that its inherited task bootstrap port is a dead Mach port. Its background
daemon started September 30 and runs Codex 0.159.3; the desktop app bundles
0.160.0. A desktop restart previously left that background daemon running.
The version mismatch is an observation, not proof of the cause.

After all agent tasks are checkpointed, use a normal macOS Terminal to update
the daemon from the app's bundled CLI and restart it. These commands are
supported by the installed CLI's help and interrupt agent sessions:

```bash
"/Applications/ChatGPT.app/Contents/Resources/codex-cli/CodexCLI.app/Contents/MacOS/codex" app-server daemon update --from-cli
"/Applications/ChatGPT.app/Contents/Resources/codex-cli/CodexCLI.app/Contents/MacOS/codex" app-server daemon restart
```

Reopen ChatGPT/Codex, resume the thread and run `render-doctor` in the fresh
agent session. Require a successful operational presentation count and a usable
bootstrap port before a one-slide export. Recovery has **not yet been tested**.
An explicit Automation denial (−1743) is a separate permission check; the present
dispatch errors do not establish such a denial. The inspected official OpenAI
documentation does not provide a specific fix for this dead-port failure; this
procedure follows the local probes and installed CLI help.

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
build or artifacts invalidate the attachment's coverage. Hidden or unrendered
slides remain explicitly uncovered.

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

Markdown claims use explicit `## claim-id` headings or `{#claim-id}` anchors. Plain prose does not establish an evidence ID. `pptxgengo design project check --project ./my-project` verifies the project structure, composition entries, and linked claim references.

## Review a fresh project packet

Generate an outline, content, or deck reviewer packet in a new directory:

```bash
work=$(mktemp -d)
pptxgengo design project review \
  --project ./my-project --stage outline --out "$work/review"
```

The packet includes the project's composition rationale and any linked claims. Review stages are `outline`, `content`, and `deck`.

The `deck` stage requires a current build. Its HTML page includes attached native
thumbnails and review status when available; otherwise it reports the missing
native evidence. Copies of the deck, PDF, PNGs and evidence are retained in the
packet with their hashes. `project view --out NEW-DIR` is the deck-stage shortcut.
