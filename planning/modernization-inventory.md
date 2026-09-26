# Software modernization deck inventory

**Source:** `samples/software modernization campaign pick deck v1 - Repaired.pptx`

**SHA-256:** `c732e693a6370a0abdf3b54f34ff456304a4a25025303c508a48559f1bee73e6`

**Evidence:** full OOXML inventory at `samples/inspection/software-modernization/inventory.json` (local, ignored), generated with `scripts/inventory-pptx.py`. The [compact metadata](modernization-inventory.json) is tracked; the full inventory can be regenerated from the committed source.

The deck contains **83 slides**, **72 native layout parts**, and **2 masters**. Two slides are marked hidden: ordered slides 7 and 20, mapped to `ppt/slides/slide7.xml` and `ppt/slides/slide20.xml`. Hidden status is read from each slide part’s `p:sld` root `show` attribute after mapping ordered `p:sldIdLst` entries through `ppt/_rels/presentation.xml.rels`; absent `show` attributes are treated as shown.

| Package parts surveyed | Object nodes | Direct/top-level objects | Native group containers |
|---|---:|---:|---:|
| Slides | 2,504 | 2,438 | 25 |
| Layouts | 625 | 592 | 6 |
| Masters | 8 | 8 | 0 |

**Counting definitions:** The script recursively counts OOXML elements `sp`, `pic`, `graphicFrame`, `cxnSp`, `grpSp`, and `contentPart` in each part's `p:spTree`. Object nodes include group containers and nested descendants; top-level objects count only direct `p:spTree` children. Native groups count `grpSp` elements, including nested groups, and do not mean editable leaf shapes. Per-kind counts are in the JSON. Totals span each slide, layout, or master part once; they are not deduplicated across inherited layout/master composition.

## Likely reusable families

These are text- and structure-based hypotheses; slides have **not** been visually reviewed.

| Slides | Likely family | Evidence basis |
|---|---|---|
| 1–2, 13, 21, 28, 32, 36, 39–43, 51, 76, 83 | Pick-deck navigation, section guidance, suggested flow | Repeated `INTERNAL PICK-DECK GUIDE` labels and guidance copy |
| 3–12 | Modernization context, economics, catalysts, value, use cases, process examples | Extracted headings and text patterns |
| 14–38 | Modernization point of view, lifecycle, delivery approach, AI-enabled methods | Repeated Rationalize / Build / Sustain themes |
| 44–50 | Intellio Evolve capability, workflow, industry examples | Extracted headings; slides 44–45 also contain image, group, and connector objects |
| 52–74 | Client proof, case studies, testimonials, quantified outcomes | `CLIENT STORIES` labels, testimonials, quantified outcome language |
| 77–82 | Offer, engagement model, rationalization scope, activity plan | Offer/scope/plan content; slides 81–82 have especially high object counts |

## Limitations

This evidence is an OOXML structural inventory, not a rendered review or a lossless-import assessment. It does not resolve inherited styles from layouts, masters, or themes; flatten group transforms; establish actual text capacity; or establish semantic object bindings. Slide text is retained in the full inventory evidence, not repeated in the compact metadata. Family labels above are provisional and should be visually checked before use as design templates.
