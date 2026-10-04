# Project and resumption

The authoritative visible deck source is `deck.yaml`, schema `pptxgengo.deck-document.v1`. Supporting Markdown preserves evidence, audience context, discussion decisions and review findings. Generated PowerPoint, canonical JSON, foundation JSON and layout reports are derived artifacts.

## Durable artifacts

Create artifacts when needed, without filler:

| Artifact | Purpose |
| --- | --- |
| `project.md` | Current stage, next useful move, confirmed/working/open frame, decision log, meaningful changes |
| `sources/index.md` | Stable source IDs, origin/path, useful pages/timestamps, authority and limitations |
| `audience-context.md` | Reader situation, expertise, prior knowledge and uncertainty; no author answer key |
| `outline.md` | Narrative and ordered page IDs with points, support and transitions |
| `claims.md` | Claim, type, source IDs, units/baseline, qualification and validation owner |
| `briefs/<slide-id>.md` | Internal purpose, required content, evidence and unresolved choices |
| `reviews/` | Reviewer findings and recorded operator decisions |
| `deck.yaml` | Final planned visible copy, order, templates, assets and local compositions |

Use the document's `context` mapping to link the supporting files under the project root. These authored references participate in dependency tracking. The reserved
`context.state` may name generated `state.json`; it is excluded from authored
dependency hashing. Content drafts for review should be generated from current `deck.yaml` once it exists; earlier working drafts are superseded and clearly marked.

## Implemented commands

`init` pins an **existing authored project**, not a blank-file wizard:

```sh
pptxgengo design project init --project ./client-deck --bundle v3
pptxgengo design project check --project ./client-deck --bundle v3
pptxgengo design project build --project ./client-deck --bundle v3
pptxgengo design project status --project ./client-deck
pptxgengo design project resume --project ./client-deck
pptxgengo design project approve --project ./client-deck --stage outline --actor 'operator-name'
```

Approval stages are `intake`, `framing`, `outline`, `content`, `selection`, `build`, `review`, `delivered`. `--slides` optionally narrows scope to stable slide IDs. Record a real decision and its actor; a command invocation does not manufacture the operator's approval. Changed relevant hashes invalidate prior approvals. `status` reports that change; `resume` persists the recomputed state.

Build directories under `builds/` contain source snapshot, source-to-object map, receipt, deck and layout report. IDs do not change when slides are reordered. Build receipts distinguish Go fit, native rendering and visual review. A successful build leaves native/visual review pending until separately performed.

On resume: read project memory, run status, inspect invalidations and baseline divergence, and proceed from the first unresolved dependency. Do not redo approved stages whose inputs remain current. If a collaborator has edited the generated deck, preserve that file and record the edit source before reconciling into YAML; automatic round trip is not implemented.

## Export scopes

```sh
pptxgengo design project export --project ./client-deck --mode client --out ./client-delivery.zip
pptxgengo design project export --project ./client-deck --mode maintainer --out ./maintenance.zip
pptxgengo design project export --project ./client-deck --mode offline --bundle v3 --out ./offline.zip
```

- `client`: current generated PowerPoint and export receipt; excludes project memory and internal sources. PDF is not yet a recorded build artifact and cannot silently be added.
- `maintainer`: current source/context/assets and latest build; contains private material and requires access appropriate to that audience.
- `offline`: maintainer material plus pinned bundle/required assets and toolchain material; relocation requires matching toolchain constraints.
- `reviewer`: bounded technical build packet. Use [review packets](review-packets.md) for independent comprehension reviews; it is not a substitute for a stage-specific fresh-context packet.

Exports require a current source/dependency build and an unchanged generated baseline. Output paths must be new. Export creates a local archive; it does not send or publish it.

## Retain current work

Keep final reference decks in the current sample destination; build experiments in temporary directories. Project `builds/` are immutable baselines for drift and later reconciliation. Preserve the current lock/state, original assets, ancestry/decisions, reviewer findings and the matching baseline when handing off an edited deck. Do not automatically prune project history. Use a maintainer archive to preserve this context; a client archive has only audience deliverables.
