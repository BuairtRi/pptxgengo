# Modernization component slice

Reviewed by the primary agent against native PowerPoint PDF previews and resolved
OOXML geometry. Source identity is pinned in `planning/source-registry.json`.

## Scope and result

- 83-slide structural inventory available; 28 selected source pages inspected
  visually: 4, 6, 8, 11, 18, 19, 22, 25, 27, 29, 30, 33, 34, 37, 38, 46, 47,
  48, 49, 50, 52, 56, 57, 74, 78, 79, 81 and 82.
- 69 source compositions in 30 source patterns; 257 named slot candidates.
- Every selected root and slot resolves through the component ingester with
  zero unresolved composition bounds. Slot text was inspected after ingestion.
- Each selected instance carries its native preview path/hash, structural
  variant and source style description. Unselected pages are not marked reviewed.

## Main findings

Useful boundaries often cross ungrouped native shapes. Examples include the
buy/build comparison, catalyst cards, icon/implication panels, three-column
evidence rows, process steps, service panels, metric panels and takeaway bands.
Native groups alone would miss most of these boundaries.

Several source patterns share semantic roles but not geometry. A single metric
and a metric pair belong to the metric family, while keeping different slot
counts. A circle-numbered caption, a side-tab catalyst card and a dense workflow
card must retain their structural variants. Nested metric panels inside case
study panels are intentionally retained as separately useful reuse boundaries.

## QA and exclusions

- Slide 78 has dense commercial copy with an overlapping fee line and bullets;
  do not promote its whole pricing panel as polished content without review.
- Dot colors on slides 27/79 encode intensity or responsibility. They are not
  interchangeable decorative accents and need semantic data bindings.
- Process steps on slides 29/33 exclude shared connectors that span components.
  Recomposition must restore those relationships at the layout level.
- Source underlines, image crops and arrows are observations. These examples do
  not establish dynamically correct word anchoring or connector attachment.
- Rich text often combines a title and body in one shape. Slots preserve that
  compound structure; independent replacement and capacity remain unproven.
- Slide 57's before/after composition retains both pictures, labels and the
  curved dashed connector. The source crop/device treatment is a dependency.

The source remains unchanged. All candidates support inspection and preservation;
none is approved for arbitrary resizing, recoloring, reflow or content replacement.
