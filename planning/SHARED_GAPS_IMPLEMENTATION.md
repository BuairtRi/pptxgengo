# Shared-gap implementation checkpoint — 2026-09-27

## Result

The seven previously open visual examples have reviewed corrections. The current
65-item gallery is `samples/template-expansion/rollout/gallery-v3/index.html`.
This is **65 specific reviewed editing examples**, not 65 designs qualified for
arbitrary content or automatic layout adaptation.

- 60 current examples have direct native PowerPoint open/export evidence.
- Five corrected accent examples have native picture-geometry replay evidence.
  PowerPoint started returning `-9074` when opening any new copy, including an
  unchanged copy of a previously verified deck, in the established working folder.
  The open review sessions still supported measurement and PDF export. The replay
  path checks the planned picture-only changes, applies them in the open session,
  exports, and restores the source frames. The source PPTX files are never saved.
  **Reopening the five newly generated accent PPTX examples is still pending.**
- Capacity envelopes, arbitrary-content qualification and variable item counts
  remain unproved. The existing qualification gate is unchanged.

## Shared capabilities delivered

| Capability | Implementation | Evidence / boundary |
|---|---|---|
| Measured accents in template workflow | `pptxtemplate adapt-accents` and `apply-accent` | Five phrase anchors; no rewritten text or moved text frames |
| Rotated underline placement | Alpha-bound placement preserves source aspect and rotation | t002's 2-degree underline and t048's phrase underline |
| Highlight placement | Fit rotated alpha bounds behind measured phrase | t004, t014, t054; neighboring words remain clear |
| Older image assets | Extract a single full-frame EMF+ bitmap for alpha measurement | t004/t054 retain original EMF bytes in the deck |
| Theme-aware semantic roles | Explicit `scheme` versus `rgb` roles; same-kind profile values | t008 both legend lines match the retained magenta series |
| Retained resource color targets | Exact XML part, source hash, node path and token | t008 chart drawing + slide connector; aggregate review builder carries both |
| Structured rich text | Explicit paragraph and run IDs, optional required format | t030 quote preserves original paragraph/run membership |
| Empty shape content zones | Styled insertion or bounded upright textbox overlay | t030 fills two bubbles without rotating their text |
| Native fit checker | Bound inner-frame comparison with identity/text/geometry validation | Added t030 zones fit; unresolved inherited/grouped coordinates remain inconclusive |

The 65 contracts still expose 895 named slots covering 1,570 original text
bindings, with two newly authored overlay zones demonstrated separately.
All 65 pass the current CLI's contract/value inspection. Go CLIs built, changed
examples were applied and packaged, native measurements/exports were exercised,
and the resulting PNGs were individually inspected. No test suite was run.

## QA findings and corrections retained

1. **Resource propagation:** component application changed a chart drawing but the
   review aggregator initially copied only the scene. It now copies reported
   resource outputs with before/after hashes and snapshots original resources.
2. **Rotated empty freeforms:** direct text insertion followed the speech bubbles'
   rotation and invalid text area. Explicit upright overlays now occupy their
   interiors. The initial failed PNG remains in the QA history.
3. **Wrong artwork identification:** an initial t004 intent selected the ProcessIQ
   logo instead of the yellow highlight. Visual QA caught it; the reviewed intent
   selects picture 5 / `image53.emf`. Source name and relationship validation alone
   cannot establish an asset's semantic purpose; its preview must be checked.
4. **Native replay restoration:** an early script used index references through
   z-order changes, which can resolve to a different shape in PowerPoint. The
   final helper uses unique names, checks order, and restores size before position.
   The affected task-only source slide was restored from its pinned source geometry
   before the final t004 review. The source file was never modified.
5. **Native fit evidence:** missing rotation originally produced invalid JSON.
   The adapter now emits `null`. The checker follows presentation relationship
   order, verifies full text and distinguishes PowerPoint soft line breaks.
6. **Fit is separate from appearance:** t030's added zones fit. Its existing quote
   reports a 3.95 pt right-bound overrun and the existing footer 2.47 pt. Both look
   contained in their visible design regions, but no blanket frame-fit pass is
   claimed. The inherited title and number frames remain inconclusive.

## Current limitations and next work

1. Reopen and render the newly built accent PPTX files directly once PowerPoint's
   new-open failure is resolved; compare against the recorded native replay PNGs.
2. Establish short/typical/dense capacity envelopes across the shortlist. The new
   fit checker reports evidence; `build-review` does not yet enforce it automatically.
3. Add bounded repeated-item adaptation and validated alternative counts. Empty
   text overlays are additive zones, not a general repeated-layout generator.
4. Broaden semantic styling beyond explicitly bound RGB/theme roles. Scheme colors
   are preserved as tokens, not flattened into one global RGB palette.
5. Extend measured accents to the Arrow family, grouped/cropped/flipped artwork,
   rotated text and multiline phrases. Current ambiguous cases produce a precise
   operator instruction and retain their original artwork. Visible parked assets
   with on-slide placement notes remain to be implemented.
6. Continue inventory qualification and import additional useful families. This
   turn fixes shared machinery; it does not add new items to the 65-template list.

The legacy West Monroe slide skill remains a historical reference asset. These
contracts, shared engine capabilities and native evidence define the new workflow.
