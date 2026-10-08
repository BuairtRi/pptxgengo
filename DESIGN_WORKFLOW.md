# Search, narrative, composition, and design passes

Architecture and workflow notes, first drafted 2026-09-25; complements
[PRODUCT_PLAN.md](PRODUCT_PLAN.md). Several catalog and authoring foundations
described here are implemented in the current CLI: source-pinned templates,
SQLite FTS5 discovery, project locks, measured layout, content matching and
native-v1 editing where measured records permit it. This document remains a
design reference for broader narrative planning, candidate evaluation and
unimplemented evaluation stages. The published default is v11; its native
rendering and visual qualification remain pending. See [current release
status](docs/release-status.md) and the [documentation index](docs/README.md).

## A unified SQLite catalog

The implemented library uses a source-pinned SQLite FTS5 index for discovery;
the catalog remains a derived index over versioned manifests and source files.
Keep original files and manifests as the durable source, rebuild indexes from
them, and store large media and PPTX files in the filesystem with hashes and
resolvable locations. The corpus counts in the sections below are historical
examples from the original architecture draft, not current inventory totals.

SQLite FTS5 supports ranked full-text retrieval and contextual snippets. That provides a practical foundation for searching descriptive metadata and document chunks; it does not infer the visual content of photographs. [SQLite FTS5 documentation](https://www.sqlite.org/fts5.html).

### Logical data model

| Record | Main fields / relationships |
|---|---|
| `sources` | Source ID, portable root alias/path, content hash, type, version, extraction version, origin |
| `items` | Stable ID/version, kind, title, description, brand/profile, review state, source ID, preview location |
| `variants` | Same concept in SVG/PNG, color variants, crop/size variants, resolution and aspect ratio |
| `chunks` | Searchable title, summary, text, tags; source locator such as slide/object ID or Markdown heading |
| `relations` | Item uses asset; layout contains component; slide exemplifies layout; rule applies to profile; revision supersedes item |
| `contracts` | Layout/component variant, item cardinality, content zones, geometry, font roles, supported transforms and fallback |
| `zones` | Slot ID, permitted content, line/item budgets, advisory character budget, minimum font size and measurement policy |
| `evaluations` | Test content, renderer/font versions, measured fit, structural/visual results, reviewer and acceptance state |
| `usage` | Deck/slide/build references to exact selected item versions; supports impact analysis when a library item changes |

Guidance is indexed by section, with metadata distinguishing corporate guidance, PowerPoint-specific rules, personal preferences, approved examples, and draft examples. Examples demonstrate usage; they do not automatically override a brand rule. Preserve the path to the original passage in each result.

Use a generated FTS table over title/description/tags/text; maintain it transactionally with the index. Do not expose raw SQL as the agent interface. The CLI should accept intent and filters and return bounded JSON results with previews, source excerpts, version, compatibility, capacity hints, review state and reason for selection.

Illustrative future query:

```text
pptxgengo library search "delivery pods with shared specialists" \
  --kind layout --brand west-monroe --profile document-proposal --json
```

Separate retrieval from selection. First constrain brand/profile, approval, editable capability and cardinality; then rank topical/design relevance. The agent examines previews of the best candidates, binds the actual content, and asks the fit solver to assess them. A high lexical score does not establish that a layout fits five pods or that a photo communicates collaboration.

Indexing pipeline:

1. Discover sources and hash them; deduplicate copied decks and asset variants
2. Extract slide/layout/object text, table text, notes and references; preserve extraction limitations for embedded objects
3. Extract Markdown by heading and assets by local/remote inventory
4. Generate previews with a recorded renderer; create image descriptions/tags with review provenance
5. Enrich use cases, layout families and content contracts, distinguishing inference from reviewed facts
6. Reindex changed sources and invalidate dependent previews/contracts when hashes change

Keep manifests authoritative for curation metadata too: catalog mutations must durably update versioned manifests or an exportable journal, not exist only inside a disposable index. Pin source and item versions in each deck's lockfile. Shared installations can combine a standard library, WM pack and project-specific content through one CLI search surface.

Use a local database with short write transactions; serialize indexing writes. SQLite WAL permits concurrent readers and a writer, but still has a single writer and should not become a shared network-drive database. [SQLite isolation](https://www.sqlite.org/isolation.html), [WAL documentation](https://www.sqlite.org/wal.html). Distribute library snapshots and rebuild local indexes. Select the Go SQLite driver after checking FTS5 support, binary size, CGO/cross-compilation and release requirements.

Start with full-text search, tags, filters and curated synonyms. Evaluate retrieval against concrete requests from the two proposals. Add embeddings only if the measured failures warrant them; retain source IDs and explanations regardless of retrieval method.

## Narrative as a traceable argument

Maintain three linked artifacts:

1. **Deck brief:** audience, client context, desired decision, reading mode and narrative progression
2. **Argument map:** section claims, evidence, assumptions, recommendations and dependencies between them
3. **Slide specification:** one main claim/theme, supporting evidence, explanatory content, visual intent and connection to the next slide

An evidence record includes source/locator, exact underlying data or passage, supported claim, date/context and validation state. Facts, hypotheses and recommendations have distinct roles. A proposal's planned work can be supported by scope/approach detail without pretending that it is an already-proven outcome.

Two evaluations are required. First, faithfully reconstruct the sample proposals, preserving their content and design choices; report any quality issues without silently rewriting them. Second, generate a new narrative from a source packet, applying WM voice and improving content choices. Perfect reconstruction alone does not establish narrative skill, and a new narrative should not be judged by identical wording to the samples.

The local brand voice guidance supplies four useful editorial checks: lead with the point, connect business and technology, use collaborative language, and create momentum through specific actions. Apply these to the deck and to each slide. Keep explanations substantial enough for the document profile.

## Choosing how to make each slide

Evaluate these construction routes explicitly:

| Route | Appropriate when | Required checks |
|---|---|---|
| Reuse an approved content slide | Existing facts and story are applicable; permitted edits are clear | Current content approval, exact dependencies, editable bindings, no client-specific leakage into a new context |
| Fill a stock layout | Argument structure matches a known composition | Number/type of slots, hierarchy, capacity, native dependencies and fit with actual text |
| Compose approved components | Familiar pieces need a new arrangement | Grid/anchors, spacing, ordering, connectors, component capacity and grouping |
| Author a new composition | Existing structures cannot express the relationship clearly | Explicit visual rationale, measurable geometry, brand constraints and full rendered review |

Do not force content into the first retrieved template. Generate a small candidate set (initially up to three), assess fit and semantic suitability, then select one. Store the rationale and rejected capacity constraints so another agent can revise the slide without repeating the search.

Avoid a global slide word limit. The sample corpus contains dense scope pages, resumes, timelines and team diagrams with different reading needs. Define capacity by layout zone and font role. The source's observed text count is a calibration input, not a guarantee that other strings of the same length will fit.

## Specialized slide families

The first catalog should include families grounded in the samples, not just generic cards:

- Team/pod composition and shared specialist pools: people/roles, pod membership, reporting versus support relationships, engagement level, client versus WM ownership and changes over time
- Individual/team resumes: identity/headshot, role, credentials and selected relevant experience, with explicit density limits
- Workplans and timelines: phases, workstreams, dependencies, milestones and temporal scale
- Scope/deliverable/responsibility tables: cell hierarchy, repeated headers, grouped rows, footnotes and pagination
- Architecture/process diagrams: nodes, containers, typed edges, legends and consistent connector routing
- Case studies: client context, problem, intervention, evidence and outcome; source and reuse approval
- Pricing/assumptions: units, periods, optional scope, totals and qualifiers; preserve numerical meaning
- Covers, section dividers, confidentiality/legal pages and closing slides: different title/evidence rules from substantive slides

For team slides, use a graph model rather than a fixed collection of boxes. A pod owns roles; a specialist can support several pods; reporting, delivery and shared-service edges have different visual encodings. Validate against fixtures with 1, 3 and 6 pods, shared people, fractional allocation and long role names. Define when a single slide becomes unreadable and needs a summary plus detail pages. Inventory the examples in EnableComp slides 15–16 and UHG slides 41–46 first; exact diagram classification still needs visual review.

## A sequence of inspectable design passes

Passes should produce artifacts and findings. They can be cached and invalidated by dependency, so a text edit does not require repeating all research. Their existence does not imply a separate model call for every slide at every stage.

| Pass | Output | Gate |
|---|---|---|
| 1. Brief and sources | Audience/profile, source index, desired decision | Scope and missing information explicit |
| 2. Narrative | Claim spine and argument/evidence map | Logical progression; no unsupported claims |
| 3. Slide content | Titles, body, evidence, notes/appendix allocation | Standalone clarity, appropriate detail, WM voice |
| 4. Composition planning | Contact-sheet plan and ranked layout/component candidates | Structure expresses the argument; varied deck rhythm |
| 5. Binding and fit | Slot bindings, resolved fonts, geometry and fit report | No truncation; readable sizes; conflicts surfaced |
| 6. Visual asset selection | Icons, illustrations, photos with crop/placement intent | Meaning matches the slide and deck; consistent style and asset variants |
| 7. Brand expression | Chosen emphasis points and anchored graphic accents | Accents clarify hierarchy and comply with brand rules |
| 8. Structural/rendered QA | Package checks, previews, object-level findings | No repair prompts, clipping, broken dependencies or unintended overlap |
| 9. Editorial/deck review | Full-size review plus contact sheet | Narrative and terminology consistent; design rhythm and emphasis coherent |
| 10. Handoff/revision | PPTX, source, assets lock, source map, baseline, QA status | Colleague edits can be reconciled and rebuilt |

Visual asset QA should ask what the image communicates, whether that matches the adjacent claim, and whether style/visual weight match the rest of the deck. A decorative stock image that matches a keyword may still be irrelevant. Record the intended meaning and focal crop to make that judgment reviewable.

### Brand accents are anchored composition objects

Model accents with explicit targets and constraints. Examples: highlighter behind a specific title phrase; hand-drawn underscore below that phrase; arrow from an evidence label to the outcome it explains. Avoid storing only arbitrary x/y coordinates.

Each accent needs asset/variant, target ID (and text range where relevant), anchor points, allowed scale, optical padding, bounds, z-order, clearance from text and other objects, collision policy and movement/grouping behavior. Raster images need transparent-padding or visible-content bounds; aligning the image rectangle alone may misalign the visible stroke.

Solve accent geometry after the target text is laid out. A line wrap invalidates its previous placement. If a highlighted phrase spans lines, either choose a different phrase or use an explicitly supported treatment; never stretch a rectangle across unrelated text. Native PowerPoint edits also invalidate the placement assumption and require a new render pass on import.

The supplied `Highlight Graphic` guidance specifically restricts highlights to **one to four words in the main headline**, using **Highlight Yellow**, on **white backgrounds**, with the mark centered behind the phrase and sized relative to cap height. Use them sparingly across the deck. Hand-drawn elements can supply emphasis on colored backgrounds. These rules belong in the brand pack and QA, not just in a prompt.

Recommended layering: frame/background → accent behind text → proof-object fills → content text → foreground annotations, with per-component overrides. Check that overlays neither obscure data nor collide with neighboring labels. An arrow's endpoints need to remain attached when either target moves. Grouping should preserve colleague editability; image assets may stay images while the slide's text/data remain native.

Use automatic bounds and occlusion checks plus a dedicated full-size rendered review for accents. Add fixtures for a changed title, one-to-four-word phrases, cap-height differences, multi-line titles, transparent image margins, rotated arrows and PowerPoint-resaved output.

## Prioritized proof

Start with ten slides selected from the 120-slide ledger to cover dense narrative, scope table, timeline, pod/team composition, resumes, photography, and layered accents. Establish all four construction routes, then expand to every source slide. In parallel, give every one of the 166 stock slides a review state and content contract; the initial three/four layouts are a development slice, not the final library scope.

Measure catalog retrieval with 20–30 tasks drawn from the proposals, including "shared specialists across pods", "phase dependencies", and "scope and assumptions". Measure annotation placement with deliberate title edits and resaves. These extend E2, E4, E5 and E6 in the product plan. A stock slide becomes agent-ready only after both successful content rebinding and rendered verification.
