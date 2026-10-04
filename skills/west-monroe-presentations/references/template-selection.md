# Template discovery and actual-content alternatives

## Search two ways

Describe the scenario and the content shape separately. For example, “weekly status” is a scenario; “four peer points, one key message, optional image/icon zones” is a shape. Original labels such as problem, goal or recommendation are soft clues. They must not prevent a structurally suitable template from appearing.

```sh
pptxgengo design library-index --bundle v5 --out ./library.sqlite
pptxgengo design library-find --index ./library.sqlite --query 'weekly status' --kinds template --limit 10
pptxgengo design library-find --index ./library.sqlite --roles point,key-message --items 4 --item-role point --kinds template --limit 10
pptxgengo design library-find --index ./library.sqlite --structures comparison --visual-forms table --kinds template --limit 10
pptxgengo design library-inspect --index ./library.sqlite --id cards/3
pptxgengo design library-preview --index ./library.sqlite --id cards/3
```

Index creation uses native Go SQLite. The output must be a new file. `--legacy-index` optionally projects the legacy catalog in its own namespace; `--legacy-root` identifies its release root when the original database lives in a different directory. Retain the original database for full legacy inspection. `--gallery` imports hash-pinned preview links only when that gallery matches the source revision. Raw legacy definitions are retrieved lazily from their hash-pinned original SQLite;
searchable summaries remain in the unified index. Read-only open validates source
pins and projection integrity. Relocation overrides are explicit; stale inputs fail instead of silently remapping.

Find exposes its available role/structure/form vocabulary. Unsupported labels remain unmatched hints. Inspect returns full definition, source identity, dependencies, capacity basis and supported adaptations. Preview returns verified artifact paths or a definition-only explanation if no matching preview exists; it does not fabricate a rendered preview.

Kinds and namespace are explicit filters. Scenario, roles, relationships and item count affect deterministic ranking. Counts associated with primary groups are stronger evidence than nested bio bullets or illustrated component examples. An exact count is a fixed source topology, not evidence that arbitrary text fits. Source max-character notes and budgets are advisory; engine compatibility is not evaluated by search.

## Pick and adapt honestly

Inspect likely candidates: exact content slots/types, item counts, stable keys, frame/content zones, usable visual roles, build support and separate specimen/qualification states. Prefer reuse without forcing unequal ideas into equal boxes or removing necessary explanation. Reuse percentage is an aspiration, never a hard gate.

Choose another template or a local composition when the source topology does not express the intended relationship. Source examples help discover primitives/components/composites; they do not imply all described source options have runtime adapters. Local definitions support the bounded node/component set documented in [source format](source-format.md).

## Build two to four alternatives using real content

The implemented `library-fit` command accepts `pptxgengo.wmds-template-document.v1` JSON with explicit `supplied_content` candidate slides. Each candidate has a unique ID, template key and its own complete closed values. Map the same supported argument into each candidate; preserve all material evidence. Candidate-specific headings/grouping may change only when the meaning remains intact or an operator approves the tradeoff.

```sh
pptxgengo design library-fit --bundle v5 --spec alternatives.json --out ./candidate-review
```

The command writes the candidate input, independent decks/foundation/layout reports and `fit-report.json`. Inspect per-candidate status and `passed`/`failed`: completing the experiment is not a promise every candidate passed. No source specimen copy is substituted for missing values. Go layout success still leaves native and visual review pending.

Present rendered alternatives with: the preserved argument, fit caveats, reading/density tradeoffs and a recommendation. Record the selected shared reference and source revision in `deck.yaml`. A new alternative can also be a separate authored YAML project using the same supported copy; the project build uses its single authoritative deck source.
