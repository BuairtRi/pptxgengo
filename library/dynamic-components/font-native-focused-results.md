# Focused capture and native example verification

The focused run completed successfully on October 1, 2026. Its frozen v4 CLI
captured **45 controls**, then separately verified the **two-slide Go-generated
example** in PowerPoint. The new observations and example are archived under
[font-native-calibration/focused](font-native-calibration/focused/analysis-v4.json).

## What passed

- All 45 controls passed the production native content/style/frame validator.
  Font file hashes, real styles, variable axes, source, decks and manifests match.
- All **14 example text shapes** match the predicted line breaks and fit their
  native safe zones. The deck includes Arial, IBM Plex Sans and IBM Plex Mono.
- The example's 16 shapes and two slides passed native final verification,
  including structure, text, fonts, paragraph/bullet properties and fit checks.
- Native line tops agree in every focused case whose wrapping agrees.
- All six focused wrapping differences were already flagged by the Go engine.

This verifies the exact archived v4 example. It does not qualify every future
deck or provide a visual design review. Its Go report still correctly says
`powerpoint_verified=false`; the separate native evidence records final fit.

## What the focused data established

**33 spacing controls:** line breaks agree in all 33. The three default-spacing
heights agree within 0.01 pt. The remaining 30 use expanded spacing, where the
full terminal allocation overestimates native height by 1.545–5.340 pt; none
underestimates height by more than 0.01 pt. Isolating before/after spacing
confirms that final space after does not contribute to the occupied union,
while spacing before the second paragraph contributes to its first box.
At 16 pt, regular/bold/italic terminal boxes agree within each family in this
packet; the 24 pt observations show the difference grows with size.

**12 boundary controls:** six line breaks agree and six differ. Plex Sans
16.5 pt fits on one native line at widths 174.625 pt and above, but wraps at
174.600 pt. Arial 26 pt wraps at 285.850 pt and fits at 285.900 pt. These are
threshold brackets for these two labels, fonts, styles and native environment,
not a general safety allowance. Applying a font-specific patch would conceal
the shaping/kerning differences without establishing a general rule.

The combined archived sets contain 438 authored cases: 393 broad references
and 45 deliberately focused controls. The focused controls concentrate on
known failure regions; their match rate is not an estimate for ordinary slides.

## Paragraph defaults need separate engineering work

Expanded terminal extents differ between the broad and focused packets, even
at the same authored family and point size. For example, the Plex Sans 16 pt,
1.5 terminal box is 25.6600 pt in the broad packet and 25.5750 pt in the focused
packet, before space before. The source content differs between those packets.

Package inspection also found that `<a:endParaRPr>` declares size but omits
an explicit font family. The theme body font is Plex Sans in the broad packet
and Arial in the focused packet. Paragraph-end font inheritance is therefore
a **candidate explanation**, not established causality. An isolated experiment
with identical content and explicit paragraph-end styling would be needed to
separate theme/default effects from other native layout behavior.

The engine remains v4. It retains conservative expanded terminal allocations
and review warnings, rather than introducing an unsupported font-specific
height formula. Its wrapping rules also remain unchanged. No source engine or
writer change has been made after the successful example verification.

## Current use and next step

The prototype is ready for a modern template pilot with authored font choices.
Default-spacing heights have broad reference coverage; expanded-spacing and
near-boundary requests retain review warnings. Go generation needs no
AppleScript, Swift or PowerPoint. Native final verification remains a separate
qualification of the exact generated deck.

No further calibration capture is required before the pilot. Exact expanded
terminal sizing and PowerPoint kerning/rounding parity remain future engine
work; they are not claimed as resolved. The installed frozen release has not
been updated.

## Evidence

- [Reproducible focused analysis](font-native-calibration/focused/analysis-v4.json).
- [Current qualification receipt](font-native-calibration/qualification-v4.json).
- [Focused native capture](font-native-calibration/focused/native.json).
- [Successful final example evidence](font-native-calibration/focused/example-v4/native-verification.json).
- [Verified example PPTX](font-native-calibration/focused/example-v4/go-example-v4.pptx).

`scripts/analyze-focused-font-capture.py` checks source/deck/manifest/receipt
hashes, current matching font files and variable defaults, source text, shape
order and raw coordinate scale. It independently audits the saved example's
native safe-zone bounds. Outputs are new-only:

```sh
python3 scripts/analyze-focused-font-capture.py --out /tmp/focused-font-analysis.json
```

The packet's captured executable, adapter and environment inspector hashes
also agree with its preparation receipt. Earlier frozen reference/calibration
artifacts remain unchanged. No unit tests were added or run.
