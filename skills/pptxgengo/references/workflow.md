# Repository workflow

Run commands from the repository root. Check the current parser/help before relying on flags; these are repository-local commands, not a promise of a stable external CLI.

## Find source patterns and assets

The catalog is a rebuildable SQLite/FTS5 index. Search a known local catalog and inspect promising records by exact ID:

```sh
python3 scripts/catalog-index.py search --db samples/component-expansion/catalog-v2.sqlite --query 'metric card' --kind component --limit 10
python3 scripts/catalog-index.py inspect --db samples/component-expansion/catalog-v2.sqlite --id component:slice-modern-metric-card:instance-001
```

`--kind` is optional. The search query uses SQLite FTS5 syntax. The database is disposable; its input JSON/JSONL, hashes, and reviewed decisions are the durable record. Rebuild with `scripts/catalog-index.py build` only when needed; consult `library/README.md` for the current inputs and options. Do not edit the source catalog to make a search result look more authoritative.

Review component source, status, and reference evidence before adapting it. `library/reference-preferences.json` records design preferences, while technical approval is a separate status; “preferred” does not mean technically approved. Existing component records may be candidates or slot candidates only. A family resemblance or catalog hit alone does not establish safe binding, layout transfer, or reuse.

For composition images, pin local `asset_path` and `asset_sha256` in JSON and keep the bytes available at that path. `library/showcase/assets.json` documents the showcase examples, including source path, hashes, dimensions, and use notes; its raster copies live under ignored `samples/showcase/assets/`. Use a local pinned equivalent where available. `pptxcompose` currently decodes PNG and JPEG raster images; use SVG through a workflow that explicitly supports it. Never trust a filename in place of a content hash.

## Source-bound scene workflow

Extract by **source slide index** (which can differ from a printed footer number):

```sh
go run ./cmd/pptxscene extract --source input.pptx --slides 12,13 --out samples/source-scenes
```

Inspect the generated `manifest.json` and scene JSON. The v2 scene project retains source resources and supported explicit bindings; it is not semantic layout inference. `pptxcomponent inspect --project samples/source-scenes --contract path/to/reviewed-contract.json` reports explicit bindings and unresolved typography inheritance. Apply only reviewed, source/hash-bound contracts:

```sh
go run ./cmd/pptxcomponent apply --project samples/source-scenes --contract path/to/contract.json --values path/to/values.json --out samples/edited-scenes
go run ./cmd/pptxscene build --project samples/edited-scenes --out samples/edited.pptx
```

`apply` changes only declared text segments and/or explicitly bound colors and returns a new project with a change log. It does not move objects, infer semantic relationships, set arbitrary fonts, or prove text fit. Preserve the source scene and contract identity; a changed scene requires explicit contract review/rebase. `pptxscene build` serializes structured scene JSON and retained resources into a new package. It is not a general XML/YAML semantic merge or a safe way to import arbitrary raw objects. Avoid hand-merging raw scene fragments or implying that unsupported OOXML properties round-trip semantically. Build defaults to visible source slides; `--slides` narrows the output, and `--freeze-slide-numbers` explicitly converts slide-number fields to text.

## Measured composition workflow

A composition spec uses `pptxgengo.compose-spec.v1` JSON with point coordinates. The current schema covers titles, pods/team roles, cards (`numbered` or `metric`), canvas (`text`, `surface`, `line`, `image`), connectors, phases/legend, and attached accents. Keep every text string, font, width, weight, and inset explicit. Probe and measure with the repository's PowerPoint adapter, then build and verify:

```sh
go run ./cmd/pptxcompose probe --spec library/showcase/deck.json --out samples/showcase-probes
go run ./cmd/pptxcompose measure --bundle samples/showcase-probes --out samples/showcase-evidence.json
go run ./cmd/pptxcompose build --spec library/showcase/deck.json --bundle samples/showcase-probes --evidence samples/showcase-evidence.json --out samples/showcase-built
go run ./cmd/pptxcompose verify --bundle samples/showcase-built --out samples/showcase-verification.json
```

Outputs must be new. Measurement and verification use Microsoft PowerPoint via `scripts/measure-compose-text.applescript`; close an existing open presentation with the same filename first. Measurement evidence pins the spec, bundle manifest, deck, adapter, and raw native observations. Use actual native measurements for fit and placement. Character counts, rough estimates, or fitting based on a different font/width are not substitutes. A passing verify checks specified text/font/frame/color/margins and bounds; it does not establish visual quality. Export and inspect final slides at presentation size.

By default, build requires the exact probed spec hash. `--reuse-measurements` permits a changed overall spec only when the complete ordered `ProbeRequests` are identical to the probed bundle and its evidence/deck/manifest and raw PowerPoint measurements remain valid. It does not allow changed text, font, weight, color, alignment, or usable width to reuse old measurements. Keep the default exact-hash behavior unless there is a concrete, reviewed reason to change only non-probed inputs such as placement or shapes.

Canvas image paths must be local and SHA-256 pinned. Canvas overlaps require declared `allow_overlap` peers; use those declarations only for intended intersections. Cards and team layouts are bounded native-shape components. The experimental canvas/cards/accent additions remain alpha until the complete native showcase has been measured, verified, and visually reviewed. Do not generalize a successful fixture into approval for every catalog pattern or a broad template library.

Wave 2 adds rich `paragraphs`/`runs`, explicit image fit/crop/focal fields, and
`canvas.kind: "shape"` for four bounded presets. Read
`library/visual-components/README.md` for those schemas and qualification limits.
Rich text requires v7 native evidence, including line-spacing multiples; a compact
source caption can depend on 0.9 spacing even when its font and frame match.
Specify image layers above their intended opaque backgrounds: correct image
bounds and hashes do not prove that an image remains visible. The review fixture
uses PNG previews traced to the original SVG hashes for EnableComp icons.

An `AccentSpec` can attach to a canvas text block or select an exact `phrase` with an occurrence or Unicode code-point range. Native character bounds resolve wrapped fragments. `multiline: per_line` explicitly requests separate marks; ambiguous or unsupported selections require a reserved manual asset/note area or fail. Underlines use calibrated baseline offsets; highlights solve the rotated visible envelope behind text. Final native verification checks phrase positions again. `ArtworkArrowSpec` uses vetted visible endpoints/tangents and uniform scale/rotation, with optional conservative ink regions for curve-aware collision checks. Read `library/diagram-components/README.md` for candidate status and exact limits. Native probes and structural tests do not establish visual approval. The older `pptxanchor` command remains a standalone placement calculator; it does not alter a deck.

## Review and handoff

Keep evidence distinct from preference and approval:

- A source/layout resemblance or catalog classification is discovery evidence, not endorsement.
- A preferred/alternate/avoid value is a design preference, not technical proof or adaptation approval.
- A component contract bounds which native properties can change; it does not establish effective inherited typography or text fit.
- Native measurement is technical evidence. Visual review determines whether hierarchy, density, asset use, and collisions work at presentation size.
- The compose alpha features must remain labeled experimental until the showcase review gate is complete. Report unverified areas rather than implying broad template support.

## Recover text from a returned generated deck

`pptxcompose recover-text --spec original.json --bundle original-built-bundle --returned colleague.pptx --text-only --out samples/recovered` produces a new `spec.json` and change report. The original spec/bundle hashes, slide order, object names/counts/frames, and image assignments/bytes must match. It restores original spec styling and does not import returned formatting or notes. It stops on incompatible changes. New text needs fresh native probes and final QA; arbitrary returned decks still use raw-scene extraction.

## Dense proposal authoring

For a detailed proposal, calibrate against the actual reference content relationships
before choosing a layout. The five patterns in `library/showcase/dense-deck.json`
exercise a five-phase activities/output matrix, a phase scope page, current/future/
measure rows, a detailed delivery team, and a release roadmap. The original
`showcase-v1` is a sparse mechanics baseline and is not the proposal-quality bar.
See `planning/COMPLEXITY_BENCHMARK.md` and `planning/DENSE_LAYOUT_GAPS.md`.

Use `pptxcompose fit-report` with the same `--spec`, `--bundle`, `--evidence` and
optional explicit `--reuse-measurements` as build. It writes all fixed text-zone
dimensions/overflows and the full planner result to a new `--out` JSON file.
Inspect `overflow_count` and `planner_passed`; a successful command exit only means
that the diagnostic report was written. Final native verification and visual
inspection are still required. Solve overflow through deliberate geometry or copy
revision, preserving reference-level hierarchy and essential detail. Never treat
word/object counts or reduced font sizes as proof of visual complexity.

Apply padding exactly once: either a text frame covering the cell with native
insets, or a frame already inset into the panel with zero native margins. In
multiline lists, separate bullet and text frames can preserve hanging alignment;
these remain plain native text shapes, not semantic PowerPoint bullet paragraphs.

## Measured grids and panels

Use `layouts` for named container/cell/block composition. Read
`library/layout-components/README.md` before authoring: child coordinates are
relative to the parent's padded content origin, row capacity uses native
measurements, and explicit limits reject overflow. `fit-report` separates
`layout_failures` from fixed text-zone failures. Inspect both plus `planner_passed`.
Do not change padding by applying it both to a cell and its already inset text.

For copy iteration, `probe --cache DIR`, `measure --cache DIR` and
`build --cache DIR` validate and reuse observations by complete rendering contract
and environment fingerprint. When probe returns an all-cached report, skip measure.
Recompiled CLI binaries invalidate cache compatibility. Final native verify and
render review are mandatory even when every probe was cached. Cache entries must
retain the original evidence/bundle artifacts; never manually upgrade old evidence
with a new environment stamp. Named layout text slots support the existing bounded
`recover-text --text-only` path; layout/formatting changes remain unsupported.
