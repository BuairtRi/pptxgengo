# Venn feedback handoff

## Checkpoint

User reported native text overruns in Venn diagrams, particularly boxed labels. The exact deck/version being reviewed was not confirmed. The last complete 522-slide deck and 6,307-node zero-rejection audit **precede this feedback fix**. No new deck was generated for this fix; regeneration and PowerPoint visual review are pending. No commit, release, installation, or font changes were made.

Immutable source basis: `intake-20261003-frozen/source`, upstream HEAD `8f9f16ab8a7e2a6fff45a97627308668f4a12753`, renderer SHA256 `72057c4f23f077f095e7b6f22705b35bbb538eee55d9d9da6cf752f6cc10d20c`. No upstream edits.

## Finding and small fix

In `internal/wmdesign/scene_intake_venn.go`, point badge width was `shapedAdvance + 10pt`. Its native editable text box was then exactly `shapedAdvance` wide; the source 5pt padding on each side lies outside that textbox and provides no trailing native allocation reserve. Existing `sequenceInlineWidth` already documents and applies a 2pt reserve for Plex Mono textboxes.

The v4 Venn badge now uses `sequenceInlineWidth(advance, layout.Style) + 10pt`. Its native text box is 2pt wider and its outline grows by the same amount. Left badges remain anchored to the left of their marker; right badges remain on the right. Font size, weight, tracking, text case, displayed content, vertical geometry, inset stroke, 5pt source padding, marker positions, and the 140pt maximum badge allocation remain intact. v3 badge geometry remains unchanged. The maximum includes the reserve; bounds are not relaxed.

This is a concrete allocation inconsistency and uses an existing native textbox policy. It is **not yet proof that it explains the user's complete visual complaint**. Native confirmation is required.

Files changed for this feedback:

- `internal/wmdesign/scene_intake_venn.go`: v4-only point badge allocation reserve.
- `internal/wmdesign/scene_venn_badge_feedback_test.go`: v3/v4 reserve, font/tracking, left/right badge padding, literal bracket copy, and 140pt bound regressions.

Focused check passed using cached Go1.27.1 environment:

```sh
go test ./internal/wmdesign -run '^(TestVennBadgeNative|TestIntakeVenn)' -count=1
```

Result: `ok github.com/buairtri/pptxgengo/internal/wmdesign 0.619s`. No new full-library audit or race run was started during graceful stop.

## What was reviewed without further changes

The frozen renderer uses own-set/all-set-center labels at 13pt Mono, pair-overlap labels at 11pt Mono, letter spacing0.06em, and line height1.25. Point badges use9.5pt Mono600 and point numbers10pt. Go mirrors those settings for v4. Badge measurement and emission use the same resolved static face, normalized hundredth-point tracking, and displayed uppercase copy. Native textbox margin is zero and autofit is disabled.

Set/region label widths use the authored `w` override or the source radius-derived width. For example, some pair labels deliberately have narrow24–30pt widths and wrap into multiple lines. No exact source/emission mismatch was established for these wider label stacks before the stop. Their text remains editable. Normal dispatcher routing supplies footnote context to rich text; point labels retain the frozen renderer's literal `textContent` semantics, including brackets.

Candidate typography only uses vertical anchors for exact face hash/size/leading combinations. Unmatched custom Venn sizes report an uncalibrated conservative vertical estimate. Do not claim those sizes or installed native font file selection are qualified because Go plans fit. Do not shrink fonts or truncate copy to hide disagreement.

## Resume next steps

1. Confirm which deck/slide the user reviewed. Rebuild the current frozen522 diagnostic deck so the2pt badge fix is included; keep the prior deck identified as prefeedback evidence.
2. Recover native automation in a fresh session using the root capture watchdog/handoff. Native access was blocked during this feedback task, so there is no new screenshot or bounds capture.
3. Capture representative two-, three-, and four-set point diagrams, including left badges, `Master index`, and long near-bound labels. Also capture narrow overlap stacks and13pt set labels.
4. Compare actual native glyph/line bounds and selected font identity against the Go records and emitted XML. Check character spacing, first baseline, final line allocation, and whether wrapping differs. Add targeted calibration/probes or allocation corrections based on that evidence.
5. Preserve normal IBM font names in final PDFs/decks, source copy/footnotes, and native editability. Only then run the complete frozen-library audit and qualify visual results.
