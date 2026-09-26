# Dense preview v1 visual review (slides 1–4)

## Scope

Reviewed the existing rendered PNGs for dense-preview-v1 slides 1–4 individually at 1920×1080. Slide 5 was excluded as requested because its rendered calendar is historical and a corrected deck is being measured. This is a visual review only; it did not edit the deck, its spec, or the renderer.

**Overall read:** Slides 1–4 now exercise the major proposal relationships identified in the complexity benchmark: sequential phase work, phase scope and exit evidence, row-by-row transformation logic, and a reporting/ownership model. They read much closer to the UHG and EnableComp reference patterns than showcase-v1. The density is carried by structured rows and ownership links rather than by copy alone. Review the new slide-4 position change in a fresh native render before closing that item.

## Slide-by-slide

### Slide 1 — Five-phase delivery sequence

[Preview PNG](../samples/component-adaptation/dynamic-pods/dense-preview-v1-render/slide-001.png) · [Reference: EnableComp 3](../samples/catalog/render/enablecomp/png/slide-003.png)

- **Typography:** Two-line 23pt headline, with 10pt body as the dominant card size and 11–12pt section labels. At this 1920×1080 view the body remains readable. Phase headings and week ranges have clear separation from the bullets.
- **Padding and density:** Five equal-width cards are close to their practical text-density limit, but each of the three activity bullets and evidence/output paragraph has breathing room. Longest bullets wrap cleanly; no visible overflow or collision. The three-item activity lists and outcome blocks create a useful 5 × (work + evidence) pattern like EnableComp 3, though the individual panel widths are tighter.
- **Reading order:** Numbers 00–04 and the aligned phase/date labels make left-to-right progression obvious. “Pre-engagement” and “Subsequent waves” distinguish outside the 12-week first release. Repeated “Activities” and “Evidence / output” labels stabilize scanning across columns. No added arrow is needed to infer sequence.
- **Assessment:** Good density without flattening every step into an identical bare card. It is comparable in work/output coverage to EnableComp 3 while keeping the hypothetical planning window explicit.

### Slide 2 — Phase 02 scope and acceptance

[Preview PNG](../samples/component-adaptation/dynamic-pods/dense-preview-v1-render/slide-002.png) · [References: UHG 24](../samples/reconstruction/reference/png/slide-024.png) and [UHG 28](../samples/reconstruction/reference/png/slide-028.png)

- **Typography:** The large phase/title type in the dark left panel establishes the anchor; 10pt is the most common body size, with 10.5–16pt headings/labels. Center copy is compact but clear at presentation scale. The 8–9pt end of the size set appears limited to small labels/footers.
- **Padding and density:** The three-zone balance is effective: phase objective/exit evidence at left; four workstream-aligned Activities/Work Products rows in the middle; acceptance package, assumptions, and variable scope at right. Bullet indents and row rules preserve alignment. The final variable-scope paragraph sits low in its column but remains visibly above the slide edge/page number; retain a bottom-safe-area check on the final render.
- **Reading order:** Phase panel → workstream matrix → acceptance/assumption evidence is natural. Within each row, activities precede the corresponding work products. The right-side package’s three acceptance views read top-down before assumptions and variable scope.
- **Assessment:** This is the closest page to UHG 24/28’s three-zone scope treatment. It conveys a complete bounded work package, not only phase copy. Avoid adding more right-panel prose without a specific decision reason.

### Slide 3 — Workflow transformation matrix

[Preview PNG](../samples/component-adaptation/dynamic-pods/dense-preview-v1-render/slide-003.png) · [References: EnableComp 4](../samples/catalog/render/enablecomp/png/slide-004.png) and [EnableComp 33](../samples/component-expansion/proposals/enablecomp/slide-033.png)

- **Typography:** The 23pt two-line headline is distinct from the 10.5pt dominant cell text. Row labels and column heads are 11pt. Important copy is not forced to tiny type to fit the grid.
- **Padding and density:** Four columns, five rows, and explicit gutters give this page substantial density. Current-state paragraphs occupy roughly three to four lines; operating-model cells remain similarly controlled, while the measures column breaks into short, scannable metric lists. Row gutters are consistent and there is no visible crop or collision. The closing baseline caveat is separated from the matrix and reads as a summary, not another row.
- **Reading order:** Workflow → Current constraints → Proposed operating model → Measures to baseline is explicit in the headers. Numbered dark workflow cells establish a down-page sequence; alternating fills and the pink measure column maintain row correspondence across the wide grid.
- **Assessment:** Strong. This is a credible row-aligned transformation page and compares well with EnableComp 4/33’s three-column matrices. The last row and summary line remain inside the safe area.

### Slide 4 — Delivery organization and ownership

[Preview PNG](../samples/component-adaptation/dynamic-pods/dense-preview-v1-render/slide-004.png) · [References: UHG 43](../samples/reconstruction/reference/png/slide-043.png) and [UHG 44](../samples/reconstruction/reference/png/slide-044.png)

- **Typography:** Headline is 23pt. The chart uses mostly 8.5–10.5pt node labels; right-panel text is 10–11pt. This is the smallest text of slides 1–4, but role labels remain readable in the high-resolution render and chart density is comparable to the 9–15pt direct-format span on UHG 43. Keep node labels at their current size or larger; do not reduce them to accommodate extra roles.
- **Padding and density:** The chart occupies the left two-thirds with clear node boxes and good text insets. The right decision-rights panel gives short responsibilities and staffing assumptions enough vertical separation. The staffing legend has a narrow but usable gap above the footer. Compared with UHG 43, this retains 21 role nodes and gives more space to role descriptions; compared with UHG 44, it omits the named/photographic roster, appropriately keeping the scenario synthetic.
- **Reading order:** The chart currently reads from paired leadership roles, to the shared specialist row, then down to the individual specialist/pod roles. The right panel is an independent read of client decisions and transition assumptions. Distinct colors and the three-item legend clarify staffing category.
- **Connection geometry:** The current preview places Program Manager to the right of Change & Training Lead while Program Analyst & Release Control sits below the other x-position. The routed analyst path creates two horizontal segments only **3.5pt apart vertically** (about 7 pixels in this render) and their ends are separated by only 9pt horizontally; at normal scale they visually approach a single crowded junction beside the pod trunk. Moving Program Manager to x=152pt, directly above the analyst, and Change & Training Lead to x=268pt should make the analyst relationship a simple vertical drop and separate it from the pod trunk. That is a sound geometry change, and it preserves role order. The current PNG appears to show the pre-swap arrangement; confirm the corrected coordinates in the next native render rather than treating this preview as proof of the fix.
- **Assessment:** The page now carries more than pod membership: client counterparts, specialists, team roles, decision rights, and transition assumptions are all represented. After the connector change, check that no specialist-to-pod links cross role labels and that the legend remains clear.

## Manifest context (not a visual score)

The dense-preview-v1 manifest describes slides 1–4 as 78/87/31/72 rendered elements, 66/79/29/42 text blocks, and 2,231/2,493/2,390/1,365 characters respectively. The corresponding font sizes range from 8–25pt on slide 1, 8–25pt on slide 2, 8–23pt on slide 3, and 8–23pt on slide 4. These are implementation diagnostics: the renderer separates native text blocks and surfaces differently than source PPTX XML groups tables and chart parts. They cannot establish design quality, native editability of every visual, or equivalence to a reference pattern. Visual quality still depends on the measured PowerPoint output and human review at presentation size.

## Final candidate review — proposal-complexity-v2

Root inspected all five final native PowerPoint renders individually. The JSON review is `samples/showcase/dense-visual-review.json`. All five are accepted as the next complexity benchmark, not as exact reference reconstructions.

- Slide 1: Five aligned phase columns, activities and evidence lanes. Longest activity list is tight but contained, with no visible collision. Timing distinguishes pre-engagement, weeks 1–12 and subsequent waves.
- Slide 2: Objective and exit sidebar, four paired workstream rows and acceptance/scope panels are readable. Variable-scope paragraph has a narrow bottom gap; no additional copy should be added without reflow.
- Slide 3: All five workflow rows preserve column alignment and explicit cell padding. Four-line constraints fit without reducing 10.5pt body. Summary remains above footer.
- Slide 4: Program Manager sits directly above the analyst; the crowded branch is removed. Role labels, pod membership, counterpart rules and staffing legend remain clear. Reporting routes are editable line segments.
- Slide 5: Twelve-month roadmap aligns seven workstreams with owners and evidence. First release is within Month 3; subsequent expansion follows. Month labels and bar labels are legible and contained.
