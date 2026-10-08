# Density design follow-ups

This records V11 source-side density and contrast decisions. They do not mean
the V11 package has completed native rendering or visual qualification. That
release gate remains pending; see [current release status](release-status.md)
and the [V11 intake qualification record](../planning/wm-design-contracts/v11/intake-20261006-649-frozen/README.md).

## Accent contrast: use the designer's color treatment

Operator direction on 2026-10-06 supersedes the earlier temporary Bold repair.
The production treatment must use the brand's Semibold 600 and the revised
colors, with no density-dependent weight exception.

| Template | Required treatment | Previous Compact contrast |
| --- | --- | --- |
| `capability-heat/annotated` | Grounded card titles on Magenta | White on Magenta: 3.5184:1 |
| `risk-heat/annotated` | Grounded card titles on Magenta | White on Magenta: 3.5184:1 |
| `pillars/two-categories-six-magenta` | Grounded numerals on small Magenta tiles | Magenta on Light Blue: 3.0185:1 |
| `pillars/two-categories-six-stacked-magenta` | Grounded numerals on small Magenta tiles | Magenta on Light Blue: 3.0185:1 |

The heat-map source revision is committed in `7e030cfd579a3d690c58e502659b9c70ea0e183d`.
The pillar tiles are defined in designer commit `440f1b7`: `numTile: "callout"`
adds a 27 × 27 pt Magenta square with no outline, positioned beside the title's
first line with a 9 pt gap. The centered IBM Plex Mono Semibold numeral is
14 pt Grounded, and both tile and numeral stay fixed at every density.
`numInk` is ignored when `numTile` is set. Grounded on Magenta has 5.23:1
contrast. Both
changes require source and bound builds, contrast checks at Comfortable,
Compact and Dense, and native PowerPoint review before release. The interim
700 implementation was never published and must be removed before qualification.

### Contrast rule

Use each text element's actual point size and weight at the selected slide
body or header density. Magenta or White accent text may only use the large-text
contrast threshold when it truly meets the size and weight criteria at that
level and still passes the required contrast ratio. On a Magenta surface,
normal text uses Grounded. Do not preserve a
Comfortable-size exemption after Compact or Dense reduces the font.

Inventory every template at all three body densities. Check every supported
configuration; explicitly record configurations prohibited by the source's
`densityLimit` as unsupported, never as contrast-passed. Contrast inspection must
continue beyond copy-fit failures so an overflowing Comfortable example does
not hide a later contrast failure. Unsupported source handlers or invalid
geometry remain coverage failures, rather than silently passing the audit.

## Fixed typography exceptions

The designer keeps tile numerals, numbers inside road/fork pins, Venn point dots and the
fixed decision marker at their original sizes because the circles do not resize.
Designer commit `03fad66` supersedes the earlier fixed-quote direction. Card
quote marks scale as `max(40, subhead size × 48/18)` with leading `size × .75`.
Pull quote marks scale as `max(40, heading size × 2.5)` with leading `size × .5`.
The six exact quote font/size/leading pairs now have preserved local native
measurements and independently derived anchors. Earlier source versions keep
their fixed quote sizes and original calibration provenance.

The source documents two utility exceptions below the normal 8 pt reading
floor: Gantt period sub-labels (7.5 pt / 9 pt leading) and Draft Review Note
status chips (6.5 pt / 9 pt leading). Their dimensions and placement stay fixed.
Other density roles follow the committed role scales.

## Latest designer revisions and qualification

Commit `2fc6c29` adds `densityLimit: comfortable` to five `agenda/schedule*`
and five `key-message/lead-questions*` templates. Their authored density remains
Comfortable. Automatic fitting must stop there; explicit Compact or Dense is
rejected with guidance to shorten content, split, or choose another template.
The limit belongs to the template and cannot be relaxed by removing compiled
slide metadata. Headers still remain Comfortable unless explicitly changed.

The same commit uses Highlight Blue for small emphasis text in
`guide/deck-overview`, `status/steering-update`, and `venn/four-text`. It also
retains the earlier Grounded-on-Magenta rendering rules. The browser's metadata
contrast sweep reports no failures in its supported scope; native CLI traversal
and native visual qualification remain separate release gates.

Commit `ab8b065` adds browser validation matching the CLI's `numTile` guards.
The final source is `3c56d842ba3abb5f24eb33cf0082be7a7a67f116`.
It contains 649 templates:
639 support all three body tiers and ten support Comfortable only, for 1,927
supported configurations and 20 explicitly unsupported configurations.

The operator approved two additional color repairs uncovered by the native-output
collector: Highlight Blue inflection labels in `maturity/ai-beyond` and
`maturity/insights`, and Grounded inactive-pin numerals in
`road-fork/decision-chosen`. The inactive gray outlines remain unchanged. The
browser sweep now includes those labels and numerals, with regression assertions
at every supported density. The final CLI audit checks 62,578 text/font/color
objects with zero contrast failures and zero traversal gaps. All three repaired
specimens also pass independent full-size PowerPoint review.

## Pending native qualification

The source-side qualification plan identifies 352 specimens for new native
review and 297 for inheritance checks against prior accepted previews. These
counts describe required coverage, not completed v11 publication evidence. The
all-source/bound build and exact native packet/inheritance receipts remain
release gates. Even after stock qualification, populated user content still
requires fit checks and native review. Publication and installation status are
tracked separately in the V11 intake checkpoint.
