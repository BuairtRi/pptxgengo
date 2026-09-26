# Wave 2 source fixture contracts

These are measured reconstruction references, not approved reusable components. Geometry is expressed in slide points on a 960 × 540 pt canvas. Source assets and slide XML are pinned by SHA-256 in [`wave2-fixture-contracts.json`](../library/component-contracts/wave2-fixture-contracts.json); extracted media paths and hashes are in the ignored [`asset manifest`](../samples/visual-wave2/assets/manifest.json). The source decks, XML, and native renders remain the evidence of record.

## UHG slide 28: deliverable thumbnail panel

The source is four independent pictures and two captions, not a grouped object. The panel’s visible source bounds span approximately x=680.27–932.56 pt and y=52.57–281.52 pt, including the “Deliverables” label at `[680.27,52.57,214.75,16.48]`. The label uses Arial bold 12 pt, blue `#0047FF`, with 7.2 pt left/right insets and centered vertical anchor. The source thumbnail/caption sequence is enumerated in the JSON. Captions use centered, italic Arial 9 pt in `#070154`, with zero text insets.

Upper caption: “Operating Model & Governance Framework”; lower caption: “Support Model & Service Levels”. Both are source-faithful reconstruction text only. Their source paragraphs use Arial 9 pt italic, centered, with XML `a:lnSpc/a:spcPct=90000` (within-line multiple 0.9). The native first-character read for the upper caption reports x=723.946655273438 pt, y=161.641265869141 pt, height=9.720000267029 pt. The source upper caption has a trailing empty paragraph that is not visible; the reconstruction control omits it and records that omission explicitly in the JSON. The upper pictures overlap: shape 11 paints over shape 8. The lower pair similarly places shape 19 over shape 15. Captions do not have explicit relationship/group metadata tying them to those pairs. Preserve tree/z order and all individual frames. A proposed rigid virtual-panel fit into the target region x=716–924, y=65–280 is scale ≈0.8245 plus translation `[155.0,21.66]`; it is a fixture transform for separately placed children, not a native group transform.

An additional native source read reports `pic7` line weight as 0.75 pt. That width is not explicit in the inspected XML; keep it identified as native-read evidence rather than an XML-authored value.

Every screenshot’s `a:srcRect` is zero crop. PowerPoint stretches each source across its frame. Image149 is conspicuously mismatched (intrinsic ratio ≈2.417 against frame ratio ≈1.744); fidelity reproduction therefore preserves the observed stretch, while any improved crop/contain behavior must be identified as a changed design. Media has no alt description, so descriptions must not be inferred from screenshot content. The changed-content fixture may use neutral synthetic diagrams, or reuse the pinned UHG28 screenshots explicitly as source examples with attribution and wording that they are not completed work or evidence for the illustrative engagement. It must never present the source screenshots as depicting that engagement.

## UHG slide 44: roster card

The audited Cam Cross card is a 172.8 × 36 pt accent4 (`#E8EEF8`) rectangle at `[36.25,148.0109]`, with no outline. The portrait occupies the card’s leftmost 36 × 36 pt; the shape frame and image frame coincide at the left. Native text-frame margins are left 43.200000762939 pt, right 3.599999904633 pt, top 3.599999904633 pt, and bottom 3.599999904633 pt. Text has one paragraph with a vertical-tab soft break between Arial 9 pt bold “Cam Cross” and Arial 9 pt regular “Executive Sponsor”; a native first-character read confirms navy RGB `{7,1,84}` (`#070154`). Paragraph alignment is left with zero before/after spacing, middle vertical anchor, `line_rule_within=true`, and `space_within=1.0`. Native glyph starts are `(79.449996948242,155.2109375)` pt for “C” and `(79.449996948242,166.010940551758)` pt for “E”; both glyph heights are about 10.8 pt. The review page approximates the one source paragraph with separate boxes at those starts, so does not claim identical source run/paragraph structure. The later picture tree item overlays the earlier card shape.

The square 600 × 600 px source portrait `image164.png` uses `srcRect` l=19,194, t=6,977, r=19,402, b=31,619 in 100,000ths; no independent focal coordinate is encoded. Slide 44 contains 10 core-team tiles, 9 SME tiles, and 19 portraits; the selected Cam example is one representative instance, not the whole roster. Changed-content card copy is “Illustrative Person A” / “Platform delivery lead” and uses a neutral initial or placeholder, with no real portrait.

## UHG slide 67: profile / bio

Slide placeholders inherit their frames and much of their style from layout 80. Coordinates below are the resolved layout positions; they are not direct transforms in slide 67 XML. The portrait is `[36.25,112.1317,126,125.1703]` pt. Identity is `[36.25,242.31,126,75.242]`; body copy `[178.588,111.25,443.412,362.75]`; industries `[635.951,106.09,272.424,178.91]`; skills `[635.951,294.759,272.424,178.91]`. The two right panels have 14.4 pt insets, accent4 fill `#E8EEF8`, 14 pt bold underlined accent2 `#50658E` headings, and Arial 12 pt `#070154` bullets. Bullet glyph `•` and several paragraph properties come from the layout/master. Paragraph spacing and default body insets not explicitly set remain inherited/unresolved; don’t substitute character-count guesses for native measurement.

Portrait `image236.png` is 600 × 600 px. Its crop is l=17,779, t=6,677, r=17,779, b=29,327 in 100,000ths; the retained region is about 64.442% × 64.0%. There is no separate focal-point value. Identity styling, body paragraph styles, and some color settings are inherited. The rendered short line under “BIO” was not located in the inspected slide/layout text-shape XML; its source is unresolved and must not be recreated as an invented line.

The JSON contains the exact extracted source text for source reconstruction controls and attributes it to the named source slide. It also marks a changed-content profile as synthetic. Any changed profile must say “Illustrative synthetic profile — no real person depicted,” use neutral initials/geometric placeholder art, and avoid invented credentials, results, or named-person facts.

## Known fidelity limits

- These contracts describe selected examples, not general template approval or automated quality/complexity scoring.
- Unspecified/inherited runs, insets, and paragraph spacing must be resolved from full layout/master/theme evidence or remain explicitly unresolved.
- The source files do not encode focal points separately from crop rectangles.
- UHG28 has no group object; panel movement is a virtual transform applied consistently to each child.
- Source images lack semantic descriptions in the available package metadata.
