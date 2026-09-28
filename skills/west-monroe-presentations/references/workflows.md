# Workflow routes and current boundaries

Use the `pptxgengo` dispatcher shipped with the skill/package. Syntax is versioned; inspect `pptxgengo --help`, the selected route's help, and the current catalog before relying on a flag or assuming a resource path. `pptxgengo catalog --templates` and `--components` print the corresponding gallery path; add `--open` to display that gallery. `pptxgengo paths` returns JSON with the absolute installed `root`, `library`, `scripts`, catalog landing/templates/components pages, and `skill` locations. Do not hardcode the frozen release directory into task artifacts. The individual route binaries are implementation details, not required global commands.

## Template review (`template`)

The package includes a changing inventory of source-bound contracts. List and inspect the available records, then select one whose structure and source content support the requested argument. Supply values only for declared slots and styles. The workflow validates bindings and produces a new review deck/bundle; it preserves source geometry and fixed text/run segmentation. It does not resize, reflow, or semantically redesign the source slide. Existing retained images/charts/artwork and facts may remain and must be checked for relevance and accuracy.

Text replacement can overflow, collide with retained art, or disrupt emphasis even when application succeeds. Build readiness means bindings/values are technically applicable; it is not content acceptance, fit proof, or arbitrary-content qualification. Use native PowerPoint measurement and render review. Template accent adaptation is a separate measured step for supported single-line phrases and supported artwork; it does not choose emphasis for the author. Multiline/rotated/ambiguous targets can require manual handling.

## New composition (`compose`)

This route creates new editable PowerPoint content from supported components and explicit specs. It can support native editable cards, pods/team structures, measured grids/panels, canvas elements, diagrams, images, and accents depending on the selected packaged capability and its maturity. It is materially more flexible than fixed-source slot replacement, while remaining bounded by documented component schemas and qualification states. Do not claim that every library example can accept arbitrary content or that the package automatically invents a complete, structurally sound deck from prose.

For measured composition, make text, style, dimensions, assets, and relationships explicit. Probe text and component variants, measure with native PowerPoint, build from matching evidence, verify, and inspect rendered slides. Fit checks catch only declared/measured issues; they do not judge whether the visual hierarchy, argument, or density is good. Keep output paths new and preserve the spec, evidence, and rendered deck that identify the reviewed version.

## Source scene and component edits (`scene`, `component`)

Extract source slides by source slide index, which may differ from printed footer numbers. Scene projects retain source dependencies and supported explicit bindings; they are not semantic layout inference or a general-purpose OOXML merge. `scene build` preserves the selected source slides in a new deck; `template build-review` produces separate per-source review decks. Neither route merges arbitrary slides from multiple different source decks into an existing proposal. Component contracts can edit declared text segments, explicit color roles, and certain bounded empty text zones. Use only matching source/hash-bound contracts. A successful application does not move shapes or prove text fit. Build to a new deck and review all edited pages.

## Library (`lib`)

Use library search to discover then inspect exact records, states, slots, constraints, measured envelopes, and previews. Search modes may distinguish qualified records from reviewed, fixture, or raw inventory candidates; use an exploratory state only as exploration. A preference or strong visual resemblance does not establish technical approval. Do not bypass a qualification gate simply to use a candidate.

When assembling contracts into a deck, preserve source/evidence attribution and narrative IDs. Check that required content belongs in the selected semantic slot. Do not force the copy into an unsuitable pattern merely to keep a selected layout.

## Anchor and diff (`anchor`, `diff`)

Anchor calculations need measured phrase bounds and the exact intended asset/geometry; the tool calculates positions but does not edit or design the deck. `diff` takes `--reference reference.png --candidate candidate.png --out NEW_DIR`; it requires equal PNG dimensions and writes `report.json`, `difference.png`, and `overlay.png`. It compares rendered pixels only; it does not compare PPTX structure or prove acceptance. Review the rendered deck in context.

## Review evidence

Use the current packaged rollout checkpoint for the exact per-contract native/open/export and accent evidence. A visual review or geometry replay substantiates only its recorded artifacts and scope; it does not certify arbitrary content, every contract, or that a newly generated PPTX opens without repair. No source contract gains arbitrary-content capacity qualification from inventory size. Report the actual scope of evidence and unresolved visual issues.
